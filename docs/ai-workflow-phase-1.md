# Phase 1: Core CLI Commands

> **Parent Document**: [ai-workflow-plan.md](./ai-workflow-plan.md)  
> **Phase**: 1 of 7  
> **Status**: Not Started  
> **Estimated Effort**: 3-4 days  
> **Prerequisites**: [Phase 0](./ai-workflow-phase-0.md) completed

## Overview

This phase implements the core CLI commands that enable the basic blocked workflow. After this phase, an agent can block a task with a question, and humans can add comments - all without session management.

**What we're building:**
- `egenskriven block <task> "question"` - Block task and add comment atomically
- `egenskriven comment <task> "text"` - Add a comment to a task
- `egenskriven comments <task>` - List comments for a task
- `egenskriven list --need-input` - Filter tasks needing human input

**What we're NOT building yet:**
- Session linking (Phase 2)
- Resume command (Phase 3)
- Tool integrations (Phase 4)

---

## Prerequisites

Before starting this phase:

1. Phase 0 is complete (all migrations applied)
2. `need_input` column works (`egenskriven move <task> need_input`)
3. `comments` collection exists and accepts records
4. Familiar with existing command structure in `internal/commands/`

---

## Tasks

### Task 1.1: Implement `egenskriven block` Command

**File**: `internal/commands/block.go`

**Description**: Create atomic command that moves a task to `need_input` and adds a comment with the blocking question.

**Usage**:
```bash
egenskriven block <task-ref> "What authentication approach should I use?"
egenskriven block WRK-123 "Should we use REST or GraphQL?" --json
echo "Long question..." | egenskriven block WRK-123 --stdin
```

**Flags**:
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--stdin` | | bool | false | Read question from stdin |
| `--json` | `-j` | bool | false | Output result as JSON |
| `--agent` | `-a` | string | "" | Agent identifier (auto-detected if possible) |

**Implementation**:

```go
package commands

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
)

func newBlockCmd(app *pocketbase.PocketBase) *cobra.Command {
	var useStdin bool
	var jsonOutput bool
	var agentName string

	cmd := &cobra.Command{
		Use:   "block <task-ref> [question]",
		Short: "Block a task and request human input",
		Long: `Move a task to the need_input column and add a comment with your question.

This is an atomic operation that ensures both the task state change and 
the comment are created together. If either fails, neither is applied.

Use this when you're blocked and need human guidance to proceed.`,
		Example: `  # Block with inline question
  egenskriven block WRK-123 "What authentication approach should I use?"
  
  # Block with question from stdin (for longer questions)
  echo "I need to decide between several options..." | egenskriven block WRK-123 --stdin
  
  # Block with JSON output
  egenskriven block abc "Should I use REST or GraphQL?" --json
  
  # Block with explicit agent name
  egenskriven block WRK-123 "Question?" --agent "opencode-build"`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			taskRef := args[0]

			// Get question from args or stdin
			var question string
			if useStdin {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("failed to read from stdin: %w", err)
				}
				question = strings.TrimSpace(string(data))
			} else if len(args) > 1 {
				question = args[1]
			} else {
				return fmt.Errorf("question is required: provide as argument or use --stdin")
			}

			if question == "" {
				return fmt.Errorf("question cannot be empty")
			}

			// Resolve task reference
			task, err := resolver.MustResolve(app, taskRef)
			if err != nil {
				return err
			}

			// Validate current state - can't block if already in need_input or done
			currentColumn := task.GetString("column")
			if currentColumn == "need_input" {
				return fmt.Errorf("task %s is already blocked (in need_input)", 
					getDisplayId(task))
			}
			if currentColumn == "done" {
				return fmt.Errorf("cannot block a completed task")
			}

			// Determine agent name
			if agentName == "" {
				agentName = getAgentName() // From environment or config
			}

			// Get comments collection
			commentsCollection, err := app.FindCollectionByNameOrId("comments")
			if err != nil {
				return fmt.Errorf("comments collection not found: %w", err)
			}

			// Execute in transaction for atomicity
			var commentId string
			err = app.RunInTransaction(func(txApp core.App) error {
				// 1. Update task column to need_input
				task.Set("column", "need_input")

				// 2. Add history entry
				history := task.Get("history")
				historyEntries := ensureHistorySlice(history)
				historyEntries = append(historyEntries, map[string]any{
					"timestamp":    time.Now().Format(time.RFC3339),
					"action":       "blocked",
					"actor":        "agent",
					"actor_detail": agentName,
					"changes": map[string]any{
						"column": map[string]any{
							"from": currentColumn,
							"to":   "need_input",
						},
					},
					"metadata": map[string]any{
						"reason": question,
					},
				})
				task.Set("history", historyEntries)

				// Save task
				if err := txApp.Save(task); err != nil {
					return fmt.Errorf("failed to update task: %w", err)
				}

				// 3. Create comment
				comment := core.NewRecord(commentsCollection)
				comment.Set("task", task.Id)
				comment.Set("content", question)
				comment.Set("author_type", "agent")
				comment.Set("author_id", agentName)
				comment.Set("metadata", map[string]any{
					"action": "block_question",
				})

				if err := txApp.Save(comment); err != nil {
					return fmt.Errorf("failed to create comment: %w", err)
				}

				commentId = comment.Id
				return nil
			})

			if err != nil {
				return err
			}

			// Output result
			displayId := getDisplayId(task)
			
			if jsonOutput {
				return outputJSON(cmd.OutOrStdout(), map[string]any{
					"success":    true,
					"task_id":    task.Id,
					"display_id": displayId,
					"column":     "need_input",
					"comment_id": commentId,
					"message":    fmt.Sprintf("Task %s blocked, awaiting human input", displayId),
				})
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Task %s blocked. Awaiting human input.\n", displayId)
			fmt.Fprintf(cmd.OutOrStdout(), "Question: %s\n", truncate(question, 100))
			return nil
		},
	}

	cmd.Flags().BoolVar(&useStdin, "stdin", false, "Read question from stdin")
	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output result as JSON")
	cmd.Flags().StringVarP(&agentName, "agent", "a", "", "Agent identifier")

	return cmd
}

// Helper functions (implement or import from existing code)

func getAgentName() string {
	// Priority: --agent flag > EGENSKRIVEN_AGENT env > config > default
	if name := os.Getenv("EGENSKRIVEN_AGENT"); name != "" {
		return name
	}
	// Could also check config file
	return "agent"
}

func getDisplayId(task *core.Record) string {
	// Return display_id if available, otherwise fallback to id prefix
	if displayId := task.GetString("display_id"); displayId != "" {
		return displayId
	}
	// Construct from board prefix + seq
	seq := task.GetInt("seq")
	if seq > 0 {
		// Would need to look up board prefix
		return fmt.Sprintf("WRK-%d", seq)
	}
	return task.Id[:8]
}

func ensureHistorySlice(history any) []map[string]any {
	if history == nil {
		return []map[string]any{}
	}
	if slice, ok := history.([]any); ok {
		result := make([]map[string]any, len(slice))
		for i, item := range slice {
			if m, ok := item.(map[string]any); ok {
				result[i] = m
			}
		}
		return result
	}
	return []map[string]any{}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func outputJSON(w io.Writer, data any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
```

**Register the command** in `internal/commands/root.go`:
```go
// In the init or setup function
rootCmd.AddCommand(newBlockCmd(app))
```

**Verification**:
```bash
# Create a test task first
egenskriven add "Test blocking" --type feature

# Block it
egenskriven block WRK-1 "What approach should I use?"

# Verify task is in need_input
egenskriven show WRK-1
# Should show column: need_input

# Verify comment was created
egenskriven comments WRK-1
# Should show the question

# Test JSON output
egenskriven block WRK-2 "Another question" --json
# Should output JSON with task_id, comment_id, etc.

# Test stdin
echo "A very long question that spans multiple lines
and contains detailed context about the problem
I'm facing..." | egenskriven block WRK-3 --stdin
```

---

### Task 1.2: Implement `egenskriven comment` Command

**File**: `internal/commands/comment.go`

**Description**: Add a comment to any task. Used by humans to respond to agent questions.

**Usage**:
```bash
egenskriven comment <task-ref> "response text"
egenskriven comment WRK-123 "Use JWT with refresh tokens" --author "john"
echo "Detailed response..." | egenskriven comment WRK-123 --stdin
```

**Flags**:
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--stdin` | | bool | false | Read comment from stdin |
| `--json` | `-j` | bool | false | Output result as JSON |
| `--author` | `-a` | string | "" | Author identifier |

**Implementation**:

```go
package commands

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
)

func newCommentCmd(app *pocketbase.PocketBase) *cobra.Command {
	var useStdin bool
	var jsonOutput bool
	var author string

	cmd := &cobra.Command{
		Use:   "comment <task-ref> [text]",
		Short: "Add a comment to a task",
		Long: `Add a comment to a task. Use this to respond to agent questions
or to add notes to any task.

If the task is in the need_input state and you include @agent in your
comment, it may trigger an auto-resume (depending on board configuration).`,
		Example: `  # Add a simple comment
  egenskriven comment WRK-123 "Use JWT with refresh tokens"
  
  # Add comment that mentions the agent (may trigger auto-resume)
  egenskriven comment WRK-123 "@agent I've decided to use OAuth2"
  
  # Add comment from stdin for longer responses
  cat response.txt | egenskriven comment WRK-123 --stdin
  
  # Specify author
  egenskriven comment WRK-123 "Approved" --author "jane.doe"`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskRef := args[0]

			// Get comment text
			var text string
			if useStdin {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("failed to read from stdin: %w", err)
				}
				text = strings.TrimSpace(string(data))
			} else if len(args) > 1 {
				text = args[1]
			} else {
				return fmt.Errorf("comment text is required: provide as argument or use --stdin")
			}

			if text == "" {
				return fmt.Errorf("comment text cannot be empty")
			}

			// Resolve task
			task, err := resolver.MustResolve(app, taskRef)
			if err != nil {
				return err
			}

			// Determine author
			authorId := resolveAuthor(author)
			authorType := "human"
			
			// Check if running in agent context
			if isAgentContext() {
				authorType = "agent"
				if authorId == "" {
					authorId = getAgentName()
				}
			}

			// Extract mentions from text
			mentions := extractMentions(text)

			// Get comments collection
			commentsCollection, err := app.FindCollectionByNameOrId("comments")
			if err != nil {
				return fmt.Errorf("comments collection not found: %w", err)
			}

			// Create comment
			comment := core.NewRecord(commentsCollection)
			comment.Set("task", task.Id)
			comment.Set("content", text)
			comment.Set("author_type", authorType)
			comment.Set("author_id", authorId)
			comment.Set("metadata", map[string]any{
				"mentions": mentions,
			})

			if err := app.Save(comment); err != nil {
				return fmt.Errorf("failed to save comment: %w", err)
			}

			// Output
			displayId := getDisplayId(task)

			if jsonOutput {
				return outputJSON(cmd.OutOrStdout(), map[string]any{
					"success":     true,
					"comment_id":  comment.Id,
					"task_id":     task.Id,
					"display_id":  displayId,
					"author_type": authorType,
					"author_id":   authorId,
					"mentions":    mentions,
				})
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Comment added to %s\n", displayId)
			if len(mentions) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Mentions: %s\n", strings.Join(mentions, ", "))
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&useStdin, "stdin", false, "Read comment from stdin")
	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output result as JSON")
	cmd.Flags().StringVarP(&author, "author", "a", "", "Author identifier")

	return cmd
}

// resolveAuthor returns the author identifier from various sources
func resolveAuthor(flagValue string) string {
	// Priority: --author flag > EGENSKRIVEN_AUTHOR env > USER env > empty
	if flagValue != "" {
		return flagValue
	}
	if author := os.Getenv("EGENSKRIVEN_AUTHOR"); author != "" {
		return author
	}
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	return ""
}

// isAgentContext checks if we're running in an AI agent context
func isAgentContext() bool {
	// Check for common agent environment indicators
	indicators := []string{
		"OPENCODE_SESSION_ID",
		"CLAUDE_SESSION_ID", 
		"CODEX_THREAD_ID",
		"EGENSKRIVEN_AGENT",
	}
	for _, env := range indicators {
		if os.Getenv(env) != "" {
			return true
		}
	}
	return false
}

// extractMentions finds @mentions in text
func extractMentions(text string) []string {
	// Match @word patterns
	re := regexp.MustCompile(`@(\w+)`)
	matches := re.FindAllStringSubmatch(text, -1)
	
	seen := make(map[string]bool)
	var mentions []string
	
	for _, match := range matches {
		mention := "@" + match[1]
		if !seen[mention] {
			seen[mention] = true
			mentions = append(mentions, mention)
		}
	}
	
	return mentions
}
```

**Register the command**:
```go
rootCmd.AddCommand(newCommentCmd(app))
```

**Verification**:
```bash
# Add a comment
egenskriven comment WRK-1 "Use JWT authentication"

# Add with author
egenskriven comment WRK-1 "Approved!" --author "jane"

# Add with mention
egenskriven comment WRK-1 "@agent I've decided to use REST"

# Verify with JSON output
egenskriven comment WRK-1 "Test" --json
# Should show mentions array if any

# Test stdin
echo "Here's my detailed response:
1. Use JWT for authentication
2. Refresh tokens should expire in 7 days
3. Access tokens should expire in 15 minutes" | egenskriven comment WRK-1 --stdin
```

---

### Task 1.3: Implement `egenskriven comments` Command

**File**: `internal/commands/comments.go`

**Description**: List comments for a task, with optional filtering.

**Usage**:
```bash
egenskriven comments <task-ref>
egenskriven comments WRK-123 --json
egenskriven comments WRK-123 --since "2026-01-07T10:00:00Z"
egenskriven comments WRK-123 --limit 5
```

**Flags**:
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--json` | `-j` | bool | false | Output as JSON |
| `--since` | | string | "" | Only show comments after timestamp |
| `--limit` | `-n` | int | 0 | Limit number of comments (0 = all) |

**Implementation**:

```go
package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
)

func newCommentsCmd(app *pocketbase.PocketBase) *cobra.Command {
	var jsonOutput bool
	var since string
	var limit int

	cmd := &cobra.Command{
		Use:   "comments <task-ref>",
		Short: "List comments for a task",
		Long: `Display all comments for a task, sorted by creation time.

This is useful for checking if there are new responses to your questions,
or for reviewing the discussion history on a task.`,
		Example: `  # List all comments
  egenskriven comments WRK-123
  
  # List as JSON
  egenskriven comments WRK-123 --json
  
  # Only show comments from the last hour
  egenskriven comments WRK-123 --since "2026-01-07T10:00:00Z"
  
  # Show only the last 5 comments
  egenskriven comments WRK-123 --limit 5`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskRef := args[0]

			// Resolve task
			task, err := resolver.MustResolve(app, taskRef)
			if err != nil {
				return err
			}

			// Build filter
			filter := fmt.Sprintf("task = '%s'", task.Id)
			
			if since != "" {
				// Parse and validate timestamp
				t, err := time.Parse(time.RFC3339, since)
				if err != nil {
					return fmt.Errorf("invalid --since timestamp (use RFC3339 format): %w", err)
				}
				filter += fmt.Sprintf(" && created > '%s'", t.Format(time.RFC3339))
			}

			// Query comments
			records, err := app.FindRecordsByFilter(
				"comments",
				filter,
				"+created", // Sort ascending (oldest first)
				limit,      // 0 means no limit
				0,          // No offset
			)
			if err != nil {
				return fmt.Errorf("failed to fetch comments: %w", err)
			}

			displayId := getDisplayId(task)

			// JSON output
			if jsonOutput {
				comments := make([]map[string]any, len(records))
				for i, r := range records {
					comments[i] = map[string]any{
						"id":          r.Id,
						"content":     r.GetString("content"),
						"author_type": r.GetString("author_type"),
						"author_id":   r.GetString("author_id"),
						"metadata":    r.Get("metadata"),
						"created":     r.GetDateTime("created").Time().Format(time.RFC3339),
					}
				}
				return outputJSON(cmd.OutOrStdout(), map[string]any{
					"task_id":    task.Id,
					"display_id": displayId,
					"count":      len(comments),
					"comments":   comments,
				})
			}

			// Human-readable output
			if len(records) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No comments on %s\n", displayId)
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Comments on %s (%d):\n\n", displayId, len(records))

			for _, r := range records {
				authorType := r.GetString("author_type")
				authorId := r.GetString("author_id")
				content := r.GetString("content")
				created := r.GetDateTime("created").Time()

				// Format author display
				var authorDisplay string
				if authorId != "" {
					authorDisplay = fmt.Sprintf("%s (%s)", authorId, authorType)
				} else {
					authorDisplay = authorType
				}

				// Format timestamp
				timeDisplay := formatRelativeTime(created)

				// Print comment
				fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s:\n", timeDisplay, authorDisplay)
				
				// Indent content
				lines := strings.Split(content, "\n")
				for _, line := range lines {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", line)
				}
				fmt.Fprintln(cmd.OutOrStdout())
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
	cmd.Flags().StringVar(&since, "since", "", "Only show comments after timestamp (RFC3339)")
	cmd.Flags().IntVarP(&limit, "limit", "n", 0, "Limit number of comments")

	return cmd
}

// formatRelativeTime formats a time as relative (e.g., "2h ago") or absolute if > 24h
func formatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		mins := int(diff.Minutes())
		return fmt.Sprintf("%dm ago", mins)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("%dh ago", hours)
	} else {
		return t.Format("Jan 2, 15:04")
	}
}
```

**Register the command**:
```go
rootCmd.AddCommand(newCommentsCmd(app))
```

**Verification**:
```bash
# Add some comments first
egenskriven comment WRK-1 "First comment"
sleep 1
egenskriven comment WRK-1 "Second comment"
sleep 1
egenskriven comment WRK-1 "Third comment"

# List all
egenskriven comments WRK-1
# Should show 3 comments with timestamps

# List as JSON
egenskriven comments WRK-1 --json
# Should output structured JSON

# Limit
egenskriven comments WRK-1 --limit 2
# Should show only 2 comments

# Since (use a recent timestamp)
egenskriven comments WRK-1 --since "$(date -u +%Y-%m-%dT%H:%M:%SZ -d '1 minute ago')"
# Should show only recent comments
```

---

### Task 1.4: Update `egenskriven list` with `--need-input` Flag

**File**: `internal/commands/list.go` (modify existing)

**Description**: Add a `--need-input` filter to quickly see tasks awaiting human input.

**New Flag**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--need-input` | bool | false | Show only tasks in need_input column |

**Implementation** (add to existing list command):

```go
// Add to the list command's flags
var needInput bool

// In the command definition
cmd.Flags().BoolVar(&needInput, "need-input", false, "Show only tasks needing human input")

// In the RunE function, add to filter building:
if needInput {
    // Override or add to column filter
    if filter != "" {
        filter += " && "
    }
    filter += "column = 'need_input'"
}
```

**Full context - where to add in existing code**:

Find the existing `list.go` file and locate:
1. The flag definitions section - add `--need-input` flag
2. The filter building section - add the column filter logic

Example of modified section:

```go
func newListCmd(app *pocketbase.PocketBase) *cobra.Command {
    // ... existing flags ...
    var needInput bool  // ADD THIS

    cmd := &cobra.Command{
        // ... existing definition ...
        RunE: func(cmd *cobra.Command, args []string) error {
            // ... existing code ...

            // Build filter
            var filters []string
            
            // ADD THIS BLOCK
            if needInput {
                filters = append(filters, "column = 'need_input'")
            }
            
            // ... rest of filter building ...
            
            filter := strings.Join(filters, " && ")
            
            // ... rest of function ...
        },
    }

    // ... existing flag definitions ...
    cmd.Flags().BoolVar(&needInput, "need-input", false, "Show only tasks needing human input")  // ADD THIS

    return cmd
}
```

**Verification**:
```bash
# Create tasks in different states
egenskriven add "Task A" --type feature
egenskriven add "Task B" --type feature  
egenskriven add "Task C" --type feature

# Block one
egenskriven block WRK-2 "Need input on this"

# List all
egenskriven list
# Should show all tasks

# List only need_input
egenskriven list --need-input
# Should show only WRK-2

# Combine with other filters
egenskriven list --need-input --json
# Should work with JSON output

# Verify empty case
egenskriven move WRK-2 in_progress  # Unblock it
egenskriven list --need-input
# Should show no tasks (or "No tasks match..." message)
```

---

### Task 1.5: Add Commands to Root Command

**File**: `internal/commands/root.go` (modify existing)

**Description**: Register all new commands with the root command.

**Implementation**:

Find the section where commands are added to the root command and add:

```go
// Add new workflow commands
rootCmd.AddCommand(newBlockCmd(app))
rootCmd.AddCommand(newCommentCmd(app))
rootCmd.AddCommand(newCommentsCmd(app))
```

**Verification**:
```bash
# Check help shows new commands
egenskriven --help
# Should list: block, comment, comments

# Check individual help
egenskriven block --help
egenskriven comment --help
egenskriven comments --help
```

---

### Task 1.6: Write Unit Tests

**File**: `internal/commands/block_test.go`

```go
package commands

import (
	"bytes"
	"strings"
	"testing"
)

func TestBlockCommand(t *testing.T) {
	app := setupTestApp(t)
	defer cleanupTestApp(t, app)

	// Create a test task
	task := createTestTask(t, app, "Test task")

	tests := []struct {
		name      string
		args      []string
		stdin     string
		wantErr   bool
		errContains string
	}{
		{
			name:    "block with question",
			args:    []string{"block", task.Id, "What should I do?"},
			wantErr: false,
		},
		{
			name:      "block without question",
			args:      []string{"block", task.Id},
			wantErr:   true,
			errContains: "question is required",
		},
		{
			name:      "block invalid task",
			args:      []string{"block", "nonexistent", "Question?"},
			wantErr:   true,
			errContains: "not found",
		},
		{
			name:    "block with stdin",
			args:    []string{"block", task.Id, "--stdin"},
			stdin:   "Long question from stdin",
			wantErr: false,
		},
		{
			name:      "block empty stdin",
			args:      []string{"block", task.Id, "--stdin"},
			stdin:    "",
			wantErr:   true,
			errContains: "cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset task state
			task.Set("column", "todo")
			app.Save(task)

			// Setup command
			cmd := newBlockCmd(app)
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetArgs(tt.args[1:]) // Skip "block"

			if tt.stdin != "" {
				cmd.SetIn(strings.NewReader(tt.stdin))
			}

			// Execute
			err := cmd.Execute()

			// Check result
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				
				// Verify task is in need_input
				refreshedTask, _ := app.FindRecordById("tasks", task.Id)
				if refreshedTask.GetString("column") != "need_input" {
					t.Errorf("task should be in need_input, got %s", refreshedTask.GetString("column"))
				}
			}
		})
	}
}

func TestBlockAlreadyBlocked(t *testing.T) {
	app := setupTestApp(t)
	defer cleanupTestApp(t, app)

	task := createTestTask(t, app, "Test task")
	task.Set("column", "need_input")
	app.Save(task)

	cmd := newBlockCmd(app)
	cmd.SetArgs([]string{task.Id, "Another question?"})
	
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "already blocked") {
		t.Errorf("expected 'already blocked' error, got: %v", err)
	}
}
```

**File**: `internal/commands/comment_test.go`

```go
package commands

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestCommentCommand(t *testing.T) {
	app := setupTestApp(t)
	defer cleanupTestApp(t, app)

	task := createTestTask(t, app, "Test task")

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "add comment",
			args:    []string{task.Id, "This is a comment"},
			wantErr: false,
		},
		{
			name:    "add comment with author",
			args:    []string{task.Id, "Comment with author", "--author", "john"},
			wantErr: false,
		},
		{
			name:    "add comment with mention",
			args:    []string{task.Id, "@agent I've decided"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newCommentCmd(app)
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCommentMentionExtraction(t *testing.T) {
	app := setupTestApp(t)
	defer cleanupTestApp(t, app)

	task := createTestTask(t, app, "Test task")

	cmd := newCommentCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{task.Id, "@agent @user please review", "--json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	json.Unmarshal(out.Bytes(), &result)

	mentions, ok := result["mentions"].([]any)
	if !ok {
		t.Fatal("mentions should be an array")
	}

	if len(mentions) != 2 {
		t.Errorf("expected 2 mentions, got %d", len(mentions))
	}
}

func TestExtractMentions(t *testing.T) {
	tests := []struct {
		text     string
		expected []string
	}{
		{"Hello @agent", []string{"@agent"}},
		{"@agent @user please help", []string{"@agent", "@user"}},
		{"No mentions here", []string{}},
		{"@agent @agent duplicate", []string{"@agent"}}, // Deduped
		{"Email test@example.com", []string{}},          // Not a mention
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got := extractMentions(tt.text)
			if len(got) != len(tt.expected) {
				t.Errorf("extractMentions(%q) = %v, want %v", tt.text, got, tt.expected)
			}
		})
	}
}
```

**File**: `internal/commands/comments_test.go`

```go
package commands

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestCommentsCommand(t *testing.T) {
	app := setupTestApp(t)
	defer cleanupTestApp(t, app)

	task := createTestTask(t, app, "Test task")
	
	// Add some comments
	addTestComment(t, app, task.Id, "First comment", "human", "user1")
	addTestComment(t, app, task.Id, "Second comment", "agent", "opencode")

	// Test listing
	cmd := newCommentsCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{task.Id})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "First comment") {
		t.Error("output should contain 'First comment'")
	}
	if !strings.Contains(output, "Second comment") {
		t.Error("output should contain 'Second comment'")
	}
}

func TestCommentsCommandJSON(t *testing.T) {
	app := setupTestApp(t)
	defer cleanupTestApp(t, app)

	task := createTestTask(t, app, "Test task")
	addTestComment(t, app, task.Id, "Test comment", "human", "user1")

	cmd := newCommentsCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{task.Id, "--json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	comments, ok := result["comments"].([]any)
	if !ok {
		t.Fatal("comments should be an array")
	}
	if len(comments) != 1 {
		t.Errorf("expected 1 comment, got %d", len(comments))
	}
}

func TestCommentsCommandLimit(t *testing.T) {
	app := setupTestApp(t)
	defer cleanupTestApp(t, app)

	task := createTestTask(t, app, "Test task")
	
	// Add 5 comments
	for i := 0; i < 5; i++ {
		addTestComment(t, app, task.Id, fmt.Sprintf("Comment %d", i), "human", "user")
	}

	cmd := newCommentsCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{task.Id, "--limit", "2", "--json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	json.Unmarshal(out.Bytes(), &result)
	
	count := int(result["count"].(float64))
	if count != 2 {
		t.Errorf("expected 2 comments with limit, got %d", count)
	}
}

// Helper functions for tests
func addTestComment(t *testing.T, app *pocketbase.PocketBase, taskId, content, authorType, authorId string) {
	collection, _ := app.FindCollectionByNameOrId("comments")
	record := core.NewRecord(collection)
	record.Set("task", taskId)
	record.Set("content", content)
	record.Set("author_type", authorType)
	record.Set("author_id", authorId)
	if err := app.Save(record); err != nil {
		t.Fatalf("failed to create test comment: %v", err)
	}
}
```

---

## Testing Checklist

Before considering this phase complete:

### Command Tests

- [ ] `egenskriven block <task> "question"` moves task to need_input
- [ ] `egenskriven block <task> "question"` creates comment with question
- [ ] `egenskriven block` fails gracefully if task already blocked
- [ ] `egenskriven block` fails gracefully if task is done
- [ ] `egenskriven block --stdin` reads from stdin
- [ ] `egenskriven block --json` outputs valid JSON

- [ ] `egenskriven comment <task> "text"` creates comment
- [ ] `egenskriven comment --author` sets author correctly
- [ ] `egenskriven comment` extracts @mentions
- [ ] `egenskriven comment --stdin` reads from stdin
- [ ] `egenskriven comment --json` outputs valid JSON

- [ ] `egenskriven comments <task>` lists all comments
- [ ] `egenskriven comments --limit N` limits output
- [ ] `egenskriven comments --since` filters by time
- [ ] `egenskriven comments --json` outputs valid JSON
- [ ] `egenskriven comments` shows helpful message when no comments

- [ ] `egenskriven list --need-input` shows only blocked tasks
- [ ] `egenskriven list --need-input` works with other filters
- [ ] `egenskriven list --need-input --json` outputs valid JSON

### Integration Tests

- [ ] Full workflow: create task → block → add comment → list comments
- [ ] Atomic block: if comment creation fails, task stays in original column
- [ ] History is updated correctly when task is blocked

### Error Handling

- [ ] Invalid task reference shows helpful error
- [ ] Empty question/comment shows helpful error
- [ ] All commands show usage when given --help

---

## Files Changed/Created

| File | Change Type | Description |
|------|-------------|-------------|
| `internal/commands/block.go` | New | Block command implementation |
| `internal/commands/comment.go` | New | Comment command implementation |
| `internal/commands/comments.go` | New | Comments list command |
| `internal/commands/list.go` | Modified | Add --need-input flag |
| `internal/commands/root.go` | Modified | Register new commands |
| `internal/commands/block_test.go` | New | Block command tests |
| `internal/commands/comment_test.go` | New | Comment command tests |
| `internal/commands/comments_test.go` | New | Comments command tests |

---

## Next Phase

Once all tests pass, proceed to [Phase 2: Session Management](./ai-workflow-phase-2.md).

Phase 2 will implement:
- `egenskriven session link` command
- `egenskriven session show` command
- `egenskriven session history` command
- Session tracking in the sessions table

---

## Notes for Implementer

1. **Look at existing commands**: Study how other commands like `move`, `update`, `show` are implemented for patterns and helpers.

2. **Resolver package**: There should be an existing resolver package for parsing task references (WRK-123, abc123, etc.). Use it.

3. **Output helpers**: Check if there are existing JSON output helpers. Use them for consistency.

4. **Error messages**: Follow the existing error message style in the codebase.

5. **Transaction handling**: The block command MUST be atomic. If the comment fails to save, the task column change must be rolled back.

6. **Testing**: Run the full test suite after changes to ensure no regressions.
