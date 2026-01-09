# Phase 3: Resume Flow

> **Parent Document**: [ai-workflow-plan.md](./ai-workflow-plan.md)  
> **Phase**: 3 of 7  
> **Status**: In Progress  
> **Estimated Effort**: 3-4 days  
> **Prerequisites**: [Phase 2](./ai-workflow-phase-2.md) completed

---

## Phase 3 Todo List

This is the comprehensive task list for completing Phase 3. All items must be completed before the phase can be marked as done.

### 1. Context Prompt Builder (`internal/resume/context.go`)

- [x] **1.1** Create `internal/resume/context.go` file
- [x] **1.2** Implement `Comment` struct with fields:
  - `Content` (string)
  - `AuthorType` (string) 
  - `AuthorId` (string)
  - `Created` (time.Time)
- [x] **1.3** Implement `BuildContextPrompt(task, comments)` function
  - Generate markdown with Task Context header
  - Include task display_id, title, status, priority
  - Include description (truncated if >500 chars)
  - Include Conversation Thread section with all comments
  - Include Instructions section for agent
- [x] **1.4** Implement `BuildMinimalPrompt(task, comments)` function
  - Only include last 3 comments
  - Truncate comments over 200 chars
  - Shorter format for token-constrained scenarios
- [x] **1.5** Implement `formatAuthorLabel(authorType, authorId)` helper
  - Return authorId if present, otherwise authorType
- [x] **1.6** Implement `getDisplayId(task)` helper
  - Check for display_id field first
  - Fall back to "WRK-{seq}" format
  - Fall back to first 8 chars of ID
- [x] **1.7** Handle empty comments case with "_No comments yet_" message

### 2. Resume Command Builder (`internal/resume/command.go`)

- [x] **2.1** Create `internal/resume/command.go` file
- [x] **2.2** Define tool constants:
  - `ToolOpenCode = "opencode"`
  - `ToolClaudeCode = "claude-code"`
  - `ToolCodex = "codex"`
- [x] **2.3** Implement `ResumeCommand` struct with fields:
  - `Tool` (string)
  - `SessionRef` (string)
  - `WorkingDir` (string)
  - `Prompt` (string)
  - `Command` (string) - full shell command
  - `Args` ([]string) - parsed arguments for exec
- [x] **2.4** Implement `BuildResumeCommand(tool, sessionRef, workingDir, prompt)` function
  - Return `*ResumeCommand` and error
  - Handle all three supported tools
- [x] **2.5** Implement `buildOpenCodeCommand(sessionRef, prompt)` helper
  - Format: `opencode run "<escaped-prompt>" --session <id>`
- [x] **2.6** Implement `buildClaudeCodeCommand(sessionRef, prompt)` helper
  - Format: `claude --resume <id> "<escaped-prompt>"`
- [x] **2.7** Implement `buildCodexCommand(sessionRef, prompt)` helper
  - Format: `codex exec resume <id> "<escaped-prompt>"`
- [x] **2.8** Implement `ValidateSessionRef(tool, ref)` function
  - Return error if ref is empty
  - Return error if ref is too short (<8 chars)
- [x] **2.9** Return appropriate error for unsupported tools

### 3. Shell Escape Handling

- [x] **3.1** Choose approach: external dependency OR custom implementation
- [x] **3.2** If external: Add `github.com/alessio/shellescape` to go.mod
- [x] **3.3** If custom: Create `internal/resume/escape.go` with `ShellQuote(s string)` function
  - Escape single quotes by replacing `'` with `'\''`
  - Wrap result in single quotes
- [x] **3.4** Verify special characters in prompts are safely escaped (quotes, newlines, etc.)

### 4. Resume Command Implementation (`internal/commands/resume.go`)

- [x] **4.1** Create `internal/commands/resume.go` file
- [x] **4.2** Implement `newResumeCmd(app *pocketbase.PocketBase)` function
- [x] **4.3** Add command flags:
  - `--exec` / `-e` (bool) - Execute the resume command
  - `--json` / `-j` (bool) - Output as JSON
  - `--minimal` / `-m` (bool) - Use minimal prompt
  - `--prompt` / `-p` (string) - Custom prompt override
  - `--dry-run` (bool) - Show command without running
- [x] **4.4** Implement task state validation
  - Error if task is not in `need_input` state
  - Include current state in error message
- [x] **4.5** Implement session validation
  - Error if no `agent_session` linked
  - Include helpful hint about `session link` command
- [x] **4.6** Implement `fetchCommentsForResume(app, taskId)` function
  - Query comments collection filtered by task
  - Sort by created ascending (chronological order)
  - Return `[]resume.Comment`
- [x] **4.7** Implement `updateTaskForResume(app, task)` function
  - Set column to "in_progress"
  - Add history entry with action "resumed"
  - Include actor and timestamp
- [x] **4.8** Implement `executeResumeCommand(rc *resume.ResumeCommand)` function
  - Change to working directory
  - Restore original directory on completion
  - Spawn process with stdin/stdout/stderr connected
- [x] **4.9** Implement `updateSessionStatusInHistory(app, taskId, externalRef, status)` function
  - Find session record by task and external_ref
  - Update status field
- [x] **4.10** Implement `indent(text, prefix)` helper for output formatting
- [x] **4.11** Implement print mode output (default)
  - Show resume command
  - Show working directory
  - Show prompt length
  - Show hint about --exec flag
- [x] **4.12** Implement JSON output mode
  - Include: task_id, display_id, tool, session_ref, working_dir, command, prompt, prompt_length
- [x] **4.13** Implement dry-run mode
  - Show what would be executed
  - Include prompt content (indented)
- [x] **4.14** Register command in root.go: `rootCmd.AddCommand(newResumeCmd(app))`

### 5. Unit Tests - Context Builder (`internal/resume/context_test.go`)

- [x] **5.1** Create `internal/resume/context_test.go` file
- [x] **5.2** Create mock record helper for testing (implements GetString, GetInt, Get methods)
- [x] **5.3** Test: `BuildContextPrompt` includes task title
- [x] **5.4** Test: `BuildContextPrompt` includes task priority
- [x] **5.5** Test: `BuildContextPrompt` includes all comments in chronological order
- [x] **5.6** Test: `BuildContextPrompt` formats authors correctly
  - Use authorId when present
  - Fall back to authorType when authorId is empty
- [x] **5.7** Test: `BuildContextPrompt` handles empty comments with placeholder text
- [x] **5.8** Test: `BuildContextPrompt` truncates descriptions over 500 chars
- [x] **5.9** Test: `BuildMinimalPrompt` only includes last 3 comments when >3 exist
- [x] **5.10** Test: `BuildMinimalPrompt` truncates comments over 200 chars

### 6. Unit Tests - Command Builder (`internal/resume/command_test.go`)

- [x] **6.1** Create `internal/resume/command_test.go` file
- [x] **6.2** Test: `BuildResumeCommand` generates correct command for opencode
  - Verify command contains "opencode run"
  - Verify session ref is included
- [x] **6.3** Test: `BuildResumeCommand` generates correct command for claude-code
  - Verify command contains "claude --resume"
  - Verify session ref is included
- [x] **6.4** Test: `BuildResumeCommand` generates correct command for codex
  - Verify command contains "codex exec resume"
  - Verify session ref is included
- [x] **6.5** Test: `BuildResumeCommand` returns error for unknown tools
- [x] **6.6** Test: Command properly escapes special characters
  - Test single quotes in prompt
  - Test double quotes in prompt
  - Verify no unbalanced quotes in output
- [x] **6.7** Test: `ValidateSessionRef` rejects empty refs
- [x] **6.8** Test: `ValidateSessionRef` rejects refs shorter than 8 chars

### 7. Unit Tests - Resume Command (`internal/commands/resume_test.go`)

- [x] **7.1** Create `internal/commands/resume_test.go` file
- [x] **7.2** Create `createTestTaskWithSession(t, app, title, tool, sessionRef)` helper
- [x] **7.3** Create `addTestComment(t, app, taskId, content, authorType, authorId)` helper
- [x] **7.4** Test: `resume` fails for task not in need_input state
  - Verify error message mentions current state
- [x] **7.5** Test: `resume` fails for task without agent_session linked
  - Verify error message includes hint about session link command
- [x] **7.6** Test: `resume` prints command by default (no --exec)
  - Verify output contains the resume command
  - Verify output mentions --exec flag
- [x] **7.7** Test: `resume --json` outputs valid JSON
  - Verify JSON parses successfully
  - Verify required fields present (tool, session_ref, command, prompt)
- [x] **7.8** Test: `resume --minimal` uses shorter prompt
  - Compare prompt_length between full and minimal modes
- [x] **7.9** Test: `resume --prompt "custom"` uses custom prompt override
- [x] **7.10** Test: `resume --exec --dry-run` shows command without executing

### 8. Integration Tests

- [x] **8.1** Test: Full workflow block → comment → resume --exec
  - Create task in in_progress
  - Block task with question
  - Add human comment response
  - Resume with --exec
  - Verify each step completes successfully
- [x] **8.2** Verify task moves from need_input to in_progress after resume --exec
- [x] **8.3** Verify history is updated with "resumed" action
  - Check timestamp is present
  - Check actor is correct
- [x] **8.4** Verify session status is updated to "active" in sessions collection

### 9. Final Validation

- [ ] **9.1** Run `go build` successfully with no errors
- [ ] **9.2** Run `go test ./internal/resume/...` - all tests pass
- [ ] **9.3** Run `go test ./internal/commands/...` - resume tests pass
- [ ] **9.4** Manual test: Print mode workflow
  - Create task
  - Link session with `session link`
  - Block task with `block` command
  - Add comment
  - Run `resume <task>` (print mode)
  - Verify command output is correct
- [ ] **9.5** Manual test: Execute mode with opencode
  - Run `resume <task> --exec` with opencode session
  - Verify tool spawns correctly
- [ ] **9.6** Manual test: Execute mode with claude-code
  - Run `resume <task> --exec` with claude-code session
  - Verify tool spawns correctly
- [ ] **9.7** Manual test: Execute mode with codex
  - Run `resume <task> --exec` with codex session
  - Verify tool spawns correctly
- [ ] **9.8** Verify JSON output format matches documented schema in this file
- [ ] **9.9** Run `go vet ./...` with no warnings
- [ ] **9.10** Verify no linting errors

---

## Phase Completion Criteria

This phase is **complete** when:

1. All 58 tasks above are checked off
2. All unit tests pass (`go test ./internal/resume/... ./internal/commands/...`)
3. Manual testing confirms resume works with all three supported tools
4. Code builds without errors or warnings
5. Documentation examples in this file match actual implementation

---

## Overview

This phase implements the resume functionality - the ability to resume an agent session on a blocked task with full context from the comments thread. This is the core feature that completes the human-AI collaboration loop.

**What we're building:**
- `egenskriven resume <task>` - Print the resume command
- `egenskriven resume <task> --exec` - Execute the resume command
- Context prompt builder (formats comments for injection)
- Tool-specific resume command generation

**What we're NOT building yet:**
- Tool integrations for session discovery (future phase)
- UI resume button (future phase)
- Auto-resume on @agent mention (future phase)

---

## Prerequisites

Before starting this phase:

1. Phase 2 is complete (session link/show/history work)
2. Tasks can have `agent_session` linked
3. Comments system works (from Phase 1)
4. Understand the resume command patterns for each tool

---

## Resume Command Patterns

Each tool has a specific pattern for resuming with a prompt:

| Tool | Resume Command Pattern |
|------|----------------------|
| OpenCode | `opencode run "<prompt>" --session <session-id>` |
| Claude Code | `claude --resume <session-id> "<prompt>"` |
| Codex | `codex exec resume <session-id> "<prompt>"` |

---

## Tasks

### Task 3.1: Create Context Prompt Builder

**File**: `internal/resume/context.go`

**Description**: Build a rich context prompt that includes task info and comment thread.

**Context Format**:
```markdown
## Task Context (from EgenSkriven)

**Task**: WRK-123 - Implement user authentication
**Status**: need_input -> in_progress
**Priority**: high
**Description**: 
Add JWT-based authentication to the API...

## Conversation Thread

[agent @ 10:30]: I'm implementing authentication and need guidance. What approach should I use - JWT tokens with refresh tokens, or session-based authentication with cookies?

[human @ 11:45]: Use JWT with refresh tokens. The refresh token should have a 7-day expiry, and the access token should be 15 minutes. Store refresh tokens in HttpOnly cookies.

[human @ 11:47]: Also make sure to implement token rotation on refresh.

## Instructions

Continue working on the task based on the human's response above. The conversation context should help you understand what was discussed.
```

**Implementation**:

```go
package resume

import (
	"fmt"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// Comment represents a comment record for building context
type Comment struct {
	Content    string
	AuthorType string
	AuthorId   string
	Created    time.Time
}

// BuildContextPrompt creates the full context prompt for resume
func BuildContextPrompt(task *core.Record, comments []Comment) string {
	var sb strings.Builder

	displayId := getDisplayId(task)
	title := task.GetString("title")
	priority := task.GetString("priority")
	description := task.GetString("description")

	// Header
	sb.WriteString("## Task Context (from EgenSkriven)\n\n")

	// Task info
	sb.WriteString(fmt.Sprintf("**Task**: %s - %s\n", displayId, title))
	sb.WriteString("**Status**: need_input -> in_progress\n")
	sb.WriteString(fmt.Sprintf("**Priority**: %s\n", priority))

	// Include description if present (truncated)
	if description != "" {
		sb.WriteString("**Description**:\n")
		if len(description) > 500 {
			sb.WriteString(description[:500] + "...\n")
		} else {
			sb.WriteString(description + "\n")
		}
	}
	sb.WriteString("\n")

	// Comments thread
	sb.WriteString("## Conversation Thread\n\n")

	if len(comments) == 0 {
		sb.WriteString("_No comments yet_\n\n")
	} else {
		for _, c := range comments {
			authorLabel := formatAuthorLabel(c.AuthorType, c.AuthorId)
			timeLabel := c.Created.Format("15:04")

			sb.WriteString(fmt.Sprintf("[%s @ %s]: %s\n\n", 
				authorLabel, timeLabel, c.Content))
		}
	}

	// Instructions
	sb.WriteString("## Instructions\n\n")
	sb.WriteString("Continue working on the task based on the human's response above. ")
	sb.WriteString("The conversation context should help you understand what was discussed. ")
	sb.WriteString("If you need more clarification, you can block the task again with a new question.\n")

	return sb.String()
}

// BuildMinimalPrompt creates a shorter prompt for token-constrained scenarios
func BuildMinimalPrompt(task *core.Record, comments []Comment) string {
	var sb strings.Builder

	displayId := getDisplayId(task)
	title := task.GetString("title")

	sb.WriteString(fmt.Sprintf("Task %s: %s\n\n", displayId, title))
	sb.WriteString("Recent comments:\n")

	// Only include last 3 comments for minimal version
	start := 0
	if len(comments) > 3 {
		start = len(comments) - 3
	}

	for _, c := range comments[start:] {
		authorLabel := formatAuthorLabel(c.AuthorType, c.AuthorId)
		// Truncate long comments
		content := c.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		sb.WriteString(fmt.Sprintf("- %s: %s\n", authorLabel, content))
	}

	sb.WriteString("\nContinue based on the above context.\n")

	return sb.String()
}

func formatAuthorLabel(authorType, authorId string) string {
	if authorId != "" {
		return authorId
	}
	return authorType
}

func getDisplayId(task *core.Record) string {
	// Check for display_id or construct from board prefix + seq
	if displayId := task.GetString("display_id"); displayId != "" {
		return displayId
	}
	seq := task.GetInt("seq")
	if seq > 0 {
		return fmt.Sprintf("WRK-%d", seq)
	}
	return task.Id[:8]
}
```

---

### Task 3.2: Create Resume Command Builder

**File**: `internal/resume/command.go`

**Description**: Generate the tool-specific resume command.

**Implementation**:

```go
package resume

import (
	"fmt"
	"strings"

	"github.com/alessio/shellescape"
)

// Tool constants
const (
	ToolOpenCode   = "opencode"
	ToolClaudeCode = "claude-code"
	ToolCodex      = "codex"
)

// ResumeCommand holds the details needed to resume a session
type ResumeCommand struct {
	Tool       string
	SessionRef string
	WorkingDir string
	Prompt     string
	Command    string   // The full command to execute
	Args       []string // Parsed arguments for exec
}

// BuildResumeCommand generates the resume command for a specific tool
func BuildResumeCommand(tool, sessionRef, workingDir, prompt string) (*ResumeCommand, error) {
	rc := &ResumeCommand{
		Tool:       tool,
		SessionRef: sessionRef,
		WorkingDir: workingDir,
		Prompt:     prompt,
	}

	switch tool {
	case ToolOpenCode:
		rc.Command = buildOpenCodeCommand(sessionRef, prompt)
		rc.Args = []string{"opencode", "run", prompt, "--session", sessionRef}

	case ToolClaudeCode:
		rc.Command = buildClaudeCodeCommand(sessionRef, prompt)
		rc.Args = []string{"claude", "--resume", sessionRef, prompt}

	case ToolCodex:
		rc.Command = buildCodexCommand(sessionRef, prompt)
		rc.Args = []string{"codex", "exec", "resume", sessionRef, prompt}

	default:
		return nil, fmt.Errorf("unsupported tool: %s", tool)
	}

	return rc, nil
}

func buildOpenCodeCommand(sessionRef, prompt string) string {
	escapedPrompt := shellescape.Quote(prompt)
	return fmt.Sprintf("opencode run %s --session %s", escapedPrompt, sessionRef)
}

func buildClaudeCodeCommand(sessionRef, prompt string) string {
	escapedPrompt := shellescape.Quote(prompt)
	return fmt.Sprintf("claude --resume %s %s", sessionRef, escapedPrompt)
}

func buildCodexCommand(sessionRef, prompt string) string {
	escapedPrompt := shellescape.Quote(prompt)
	return fmt.Sprintf("codex exec resume %s %s", sessionRef, escapedPrompt)
}

// ValidateSessionRef checks if the session ref format is valid for the tool
func ValidateSessionRef(tool, ref string) error {
	if ref == "" {
		return fmt.Errorf("session reference is empty")
	}

	// Basic validation - could be more strict per tool
	switch tool {
	case ToolOpenCode, ToolClaudeCode, ToolCodex:
		// These typically use UUIDs or similar
		if len(ref) < 8 {
			return fmt.Errorf("session reference seems too short: %s", ref)
		}
	}

	return nil
}
```

---

### Task 3.3: Implement `egenskriven resume` Command

**File**: `internal/commands/resume.go`

**Description**: The main resume command that fetches task context, builds the prompt, and either prints or executes the resume command.

**Usage**:
```bash
egenskriven resume <task-ref>           # Print command
egenskriven resume <task-ref> --exec    # Execute command
egenskriven resume <task-ref> --json    # Output as JSON
egenskriven resume <task-ref> --minimal # Use minimal prompt
```

**Flags**:
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--exec` | `-e` | bool | false | Execute the resume command |
| `--json` | `-j` | bool | false | Output as JSON |
| `--minimal` | `-m` | bool | false | Use minimal prompt (fewer tokens) |
| `--prompt` | `-p` | string | "" | Override the context prompt |
| `--dry-run` | | bool | false | With --exec, show command without running |

**Implementation**:

```go
package commands

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
	
	"egenskriven/internal/resume"
)

func newResumeCmd(app *pocketbase.PocketBase) *cobra.Command {
	var execFlag bool
	var jsonOutput bool
	var minimal bool
	var customPrompt string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "resume <task-ref>",
		Short: "Resume work on a blocked task",
		Long: `Generate or execute the command to resume an AI agent session
for a task that is in the need_input state.

By default, this prints the resume command that you can copy and run.
Use --exec to execute the command directly.

The resume command includes context from the task and comment thread,
which is injected into the agent's session.`,
		Example: `  # Print the resume command
  egenskriven resume WRK-123
  
  # Execute the resume directly
  egenskriven resume WRK-123 --exec
  
  # Dry run - show what would be executed
  egenskriven resume WRK-123 --exec --dry-run
  
  # Use minimal prompt (fewer tokens)
  egenskriven resume WRK-123 --minimal
  
  # Output as JSON (for scripting)
  egenskriven resume WRK-123 --json
  
  # Custom prompt override
  egenskriven resume WRK-123 --prompt "Continue with the JWT implementation"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskRef := args[0]

			// Resolve task
			task, err := resolver.MustResolve(app, taskRef)
			if err != nil {
				return err
			}

			displayId := getDisplayId(task)

			// Validate task state
			column := task.GetString("column")
			if column != "need_input" {
				return fmt.Errorf("task %s is not in need_input state (current: %s)", 
					displayId, column)
			}

			// Get session info
			sessionData := task.Get("agent_session")
			if sessionData == nil {
				return fmt.Errorf("no agent session linked to task %s\n\nTo resume, first link a session:\n  egenskriven session link %s --tool <tool> --ref <session-id>", 
					displayId, displayId)
			}

			session := sessionData.(map[string]any)
			tool := session["tool"].(string)
			sessionRef := session["ref"].(string)
			workingDir := session["working_dir"].(string)

			// Validate session ref
			if err := resume.ValidateSessionRef(tool, sessionRef); err != nil {
				return fmt.Errorf("invalid session: %w", err)
			}

			// Fetch comments
			comments, err := fetchCommentsForResume(app, task.Id)
			if err != nil {
				return fmt.Errorf("failed to fetch comments: %w", err)
			}

			// Build context prompt
			var prompt string
			if customPrompt != "" {
				prompt = customPrompt
			} else if minimal {
				prompt = resume.BuildMinimalPrompt(task, comments)
			} else {
				prompt = resume.BuildContextPrompt(task, comments)
			}

			// Build resume command
			resumeCmd, err := resume.BuildResumeCommand(tool, sessionRef, workingDir, prompt)
			if err != nil {
				return err
			}

			// JSON output
			if jsonOutput {
				return outputJSON(cmd.OutOrStdout(), map[string]any{
					"task_id":     task.Id,
					"display_id":  displayId,
					"tool":        tool,
					"session_ref": sessionRef,
					"working_dir": workingDir,
					"command":     resumeCmd.Command,
					"prompt":      prompt,
					"prompt_length": len(prompt),
				})
			}

			// Execute mode
			if execFlag {
				if dryRun {
					fmt.Fprintf(cmd.OutOrStdout(), "Would execute in %s:\n\n", workingDir)
					fmt.Fprintf(cmd.OutOrStdout(), "  %s\n\n", resumeCmd.Command)
					fmt.Fprintf(cmd.OutOrStdout(), "Prompt (%d chars):\n%s\n", 
						len(prompt), indent(prompt, "  "))
					return nil
				}

				// Update task status BEFORE executing
				if err := updateTaskForResume(app, task); err != nil {
					return fmt.Errorf("failed to update task: %w", err)
				}

				fmt.Fprintf(cmd.OutOrStdout(), "Resuming session for %s...\n", displayId)
				fmt.Fprintf(cmd.OutOrStdout(), "Tool: %s\n", tool)
				fmt.Fprintf(cmd.OutOrStdout(), "Working directory: %s\n\n", workingDir)

				return executeResumeCommand(resumeCmd)
			}

			// Print mode (default)
			fmt.Fprintf(cmd.OutOrStdout(), "Resume command for %s:\n\n", displayId)
			fmt.Fprintf(cmd.OutOrStdout(), "  %s\n\n", resumeCmd.Command)
			fmt.Fprintf(cmd.OutOrStdout(), "Working directory: %s\n", workingDir)
			fmt.Fprintf(cmd.OutOrStdout(), "Prompt length: %d characters\n\n", len(prompt))
			fmt.Fprintf(cmd.OutOrStdout(), "To execute directly, run:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  egenskriven resume %s --exec\n", displayId)

			return nil
		},
	}

	cmd.Flags().BoolVarP(&execFlag, "exec", "e", false, "Execute the resume command")
	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
	cmd.Flags().BoolVarP(&minimal, "minimal", "m", false, "Use minimal prompt")
	cmd.Flags().StringVarP(&customPrompt, "prompt", "p", "", "Custom prompt override")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show command without executing")

	return cmd
}

// fetchCommentsForResume gets comments formatted for the resume context
func fetchCommentsForResume(app *pocketbase.PocketBase, taskId string) ([]resume.Comment, error) {
	records, err := app.FindRecordsByFilter(
		"comments",
		fmt.Sprintf("task = '%s'", taskId),
		"+created", // Oldest first for chronological order
		100,        // Reasonable limit
		0,
	)
	if err != nil {
		return nil, err
	}

	comments := make([]resume.Comment, len(records))
	for i, r := range records {
		comments[i] = resume.Comment{
			Content:    r.GetString("content"),
			AuthorType: r.GetString("author_type"),
			AuthorId:   r.GetString("author_id"),
			Created:    r.GetDateTime("created").Time(),
		}
	}

	return comments, nil
}

// updateTaskForResume moves task to in_progress and adds history
func updateTaskForResume(app *pocketbase.PocketBase, task *core.Record) error {
	task.Set("column", "in_progress")

	// Add history entry
	history := task.Get("history")
	historyEntries := ensureHistorySlice(history)
	historyEntries = append(historyEntries, map[string]any{
		"timestamp":    time.Now().Format(time.RFC3339),
		"action":       "resumed",
		"actor":        "user",
		"actor_detail": os.Getenv("USER"),
		"changes": map[string]any{
			"column": map[string]any{
				"from": "need_input",
				"to":   "in_progress",
			},
		},
	})
	task.Set("history", historyEntries)

	// Update session status to active
	sessionData := task.Get("agent_session")
	if sessionData != nil {
		session := sessionData.(map[string]any)
		// Mark session in history as active/resumed
		updateSessionStatusInHistory(app, task.Id, session["ref"].(string), "active")
	}

	return app.Save(task)
}

// executeResumeCommand runs the resume command
func executeResumeCommand(rc *resume.ResumeCommand) error {
	// Change to working directory
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if err := os.Chdir(rc.WorkingDir); err != nil {
		return fmt.Errorf("failed to change to working directory %s: %w", rc.WorkingDir, err)
	}
	defer os.Chdir(originalDir)

	// Execute command
	// We use the first arg as the command and rest as arguments
	execCmd := exec.Command(rc.Args[0], rc.Args[1:]...)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}

// updateSessionStatusInHistory updates the session record status
func updateSessionStatusInHistory(app *pocketbase.PocketBase, taskId, externalRef, status string) {
	records, err := app.FindRecordsByFilter(
		"sessions",
		fmt.Sprintf("task = '%s' && external_ref = '%s'", taskId, externalRef),
		"-created",
		1,
		0,
	)
	if err != nil || len(records) == 0 {
		return
	}

	record := records[0]
	record.Set("status", status)
	app.Save(record)
}

// indent adds prefix to each line of text
func indent(text, prefix string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}
```

**Register the command**:
```go
rootCmd.AddCommand(newResumeCmd(app))
```

---

### Task 3.4: Add Shell Escape Dependency

The resume command needs to properly escape prompts for shell execution.

**Option 1**: Use an existing package
```bash
go get github.com/alessio/shellescape
```

**Option 2**: Implement simple escaping
```go
// internal/resume/escape.go
package resume

import "strings"

// ShellQuote wraps a string in single quotes for shell safety
func ShellQuote(s string) string {
    // Replace single quotes with '\''
    escaped := strings.ReplaceAll(s, "'", "'\\''")
    return "'" + escaped + "'"
}
```

---

### Task 3.5: Write Unit Tests

**File**: `internal/resume/context_test.go`

```go
package resume

import (
	"strings"
	"testing"
	"time"
)

func TestBuildContextPrompt(t *testing.T) {
	task := &mockRecord{
		data: map[string]any{
			"title":       "Implement authentication",
			"priority":    "high",
			"description": "Add JWT auth to API",
			"seq":         42,
		},
	}

	comments := []Comment{
		{
			Content:    "What approach should I use?",
			AuthorType: "agent",
			AuthorId:   "opencode",
			Created:    time.Now().Add(-1 * time.Hour),
		},
		{
			Content:    "Use JWT with refresh tokens",
			AuthorType: "human",
			AuthorId:   "john",
			Created:    time.Now().Add(-30 * time.Minute),
		},
	}

	prompt := BuildContextPrompt(task, comments)

	// Check that key elements are present
	assertions := []string{
		"## Task Context",
		"Implement authentication",
		"high",
		"## Conversation Thread",
		"What approach should I use?",
		"Use JWT with refresh tokens",
		"## Instructions",
	}

	for _, expected := range assertions {
		if !strings.Contains(prompt, expected) {
			t.Errorf("prompt should contain %q", expected)
		}
	}
}

func TestBuildMinimalPrompt(t *testing.T) {
	task := &mockRecord{
		data: map[string]any{
			"title": "Test task",
			"seq":   1,
		},
	}

	// Create 5 comments
	comments := make([]Comment, 5)
	for i := 0; i < 5; i++ {
		comments[i] = Comment{
			Content:    fmt.Sprintf("Comment %d", i),
			AuthorType: "human",
			Created:    time.Now(),
		}
	}

	prompt := BuildMinimalPrompt(task, comments)

	// Minimal should only include last 3 comments
	if strings.Contains(prompt, "Comment 0") {
		t.Error("minimal prompt should not contain old comments")
	}
	if !strings.Contains(prompt, "Comment 4") {
		t.Error("minimal prompt should contain recent comments")
	}
}

func TestBuildContextPromptNoComments(t *testing.T) {
	task := &mockRecord{
		data: map[string]any{
			"title":    "Test task",
			"priority": "medium",
			"seq":      1,
		},
	}

	prompt := BuildContextPrompt(task, []Comment{})

	if !strings.Contains(prompt, "No comments yet") {
		t.Error("should indicate no comments")
	}
}

// Mock record for testing
type mockRecord struct {
	data map[string]any
}

func (m *mockRecord) GetString(key string) string {
	if v, ok := m.data[key].(string); ok {
		return v
	}
	return ""
}

func (m *mockRecord) GetInt(key string) int {
	if v, ok := m.data[key].(int); ok {
		return v
	}
	return 0
}

func (m *mockRecord) Get(key string) any {
	return m.data[key]
}
```

**File**: `internal/resume/command_test.go`

```go
package resume

import (
	"strings"
	"testing"
)

func TestBuildResumeCommand(t *testing.T) {
	tests := []struct {
		tool       string
		sessionRef string
		prompt     string
		wantCmd    string
		wantErr    bool
	}{
		{
			tool:       "opencode",
			sessionRef: "abc-123",
			prompt:     "Continue",
			wantCmd:    "opencode run",
			wantErr:    false,
		},
		{
			tool:       "claude-code",
			sessionRef: "def-456",
			prompt:     "Continue",
			wantCmd:    "claude --resume",
			wantErr:    false,
		},
		{
			tool:       "codex",
			sessionRef: "ghi-789",
			prompt:     "Continue",
			wantCmd:    "codex exec resume",
			wantErr:    false,
		},
		{
			tool:       "unknown",
			sessionRef: "xxx",
			prompt:     "Continue",
			wantCmd:    "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			rc, err := BuildResumeCommand(tt.tool, tt.sessionRef, "/tmp", tt.prompt)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !strings.Contains(rc.Command, tt.wantCmd) {
				t.Errorf("command should contain %q, got %q", tt.wantCmd, rc.Command)
			}

			if !strings.Contains(rc.Command, tt.sessionRef) {
				t.Errorf("command should contain session ref %q", tt.sessionRef)
			}
		})
	}
}

func TestBuildResumeCommandEscaping(t *testing.T) {
	// Test that special characters in prompts are escaped
	prompt := "What's the status? Use 'single quotes' and \"double quotes\""
	
	rc, err := BuildResumeCommand("opencode", "abc-123", "/tmp", prompt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Command should be safe to execute
	// At minimum, shouldn't contain unescaped quotes that break parsing
	if strings.Count(rc.Command, "'") % 2 != 0 {
		t.Error("command has unbalanced single quotes")
	}
}

func TestValidateSessionRef(t *testing.T) {
	tests := []struct {
		tool    string
		ref     string
		wantErr bool
	}{
		{"opencode", "550e8400-e29b-41d4-a716-446655440000", false},
		{"opencode", "abc123", false},
		{"opencode", "", true},
		{"opencode", "short", true}, // Too short
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			err := ValidateSessionRef(tt.tool, tt.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSessionRef(%q, %q) error = %v, wantErr %v",
					tt.tool, tt.ref, err, tt.wantErr)
			}
		})
	}
}
```

**File**: `internal/commands/resume_test.go`

```go
package commands

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestResumeCommandValidation(t *testing.T) {
	app := setupTestApp(t)
	defer cleanupTestApp(t, app)

	// Task not in need_input
	task := createTestTask(t, app, "Test task")
	task.Set("column", "in_progress")
	app.Save(task)

	cmd := newResumeCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{task.Id})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "not in need_input") {
		t.Errorf("should fail for non-blocked task, got: %v", err)
	}
}

func TestResumeCommandNoSession(t *testing.T) {
	app := setupTestApp(t)
	defer cleanupTestApp(t, app)

	task := createTestTask(t, app, "Test task")
	task.Set("column", "need_input")
	app.Save(task)

	cmd := newResumeCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{task.Id})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "no agent session") {
		t.Errorf("should fail without session, got: %v", err)
	}
}

func TestResumeCommandPrintMode(t *testing.T) {
	app := setupTestApp(t)
	defer cleanupTestApp(t, app)

	task := createTestTaskWithSession(t, app, "Test task", "opencode", "test-session-123")
	addTestComment(t, app, task.Id, "What should I do?", "agent", "opencode")
	addTestComment(t, app, task.Id, "Use JWT", "human", "john")

	cmd := newResumeCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{task.Id})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()

	// Should print the command
	if !strings.Contains(output, "opencode run") {
		t.Error("should show opencode run command")
	}
	if !strings.Contains(output, "test-session-123") {
		t.Error("should include session ref")
	}
	if !strings.Contains(output, "--exec") {
		t.Error("should mention --exec flag")
	}
}

func TestResumeCommandJSONOutput(t *testing.T) {
	app := setupTestApp(t)
	defer cleanupTestApp(t, app)

	task := createTestTaskWithSession(t, app, "Test task", "opencode", "test-session-123")

	cmd := newResumeCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{task.Id, "--json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if result["tool"] != "opencode" {
		t.Errorf("expected tool 'opencode', got %v", result["tool"])
	}
	if result["session_ref"] != "test-session-123" {
		t.Errorf("expected session_ref 'test-session-123', got %v", result["session_ref"])
	}
	if result["command"] == nil || result["command"] == "" {
		t.Error("should include command")
	}
}

func TestResumeCommandMinimalPrompt(t *testing.T) {
	app := setupTestApp(t)
	defer cleanupTestApp(t, app)

	task := createTestTaskWithSession(t, app, "Test task", "opencode", "test-123")
	
	// Add many comments
	for i := 0; i < 10; i++ {
		addTestComment(t, app, task.Id, fmt.Sprintf("Comment %d", i), "human", "user")
	}

	// Get full prompt
	cmd := newResumeCmd(app)
	var fullOut bytes.Buffer
	cmd.SetOut(&fullOut)
	cmd.SetArgs([]string{task.Id, "--json"})
	cmd.Execute()

	var fullResult map[string]any
	json.Unmarshal(fullOut.Bytes(), &fullResult)
	fullPromptLen := int(fullResult["prompt_length"].(float64))

	// Get minimal prompt
	cmd = newResumeCmd(app)
	var minOut bytes.Buffer
	cmd.SetOut(&minOut)
	cmd.SetArgs([]string{task.Id, "--json", "--minimal"})
	cmd.Execute()

	var minResult map[string]any
	json.Unmarshal(minOut.Bytes(), &minResult)
	minPromptLen := int(minResult["prompt_length"].(float64))

	// Minimal should be shorter
	if minPromptLen >= fullPromptLen {
		t.Errorf("minimal prompt (%d) should be shorter than full (%d)", 
			minPromptLen, fullPromptLen)
	}
}

// Helper to create task with session
func createTestTaskWithSession(t *testing.T, app *pocketbase.PocketBase, 
	title, tool, sessionRef string) *core.Record {
	
	task := createTestTask(t, app, title)
	task.Set("column", "need_input")
	task.Set("agent_session", map[string]any{
		"tool":        tool,
		"ref":         sessionRef,
		"ref_type":    "uuid",
		"working_dir": "/tmp",
		"linked_at":   time.Now().Format(time.RFC3339),
	})
	
	if err := app.Save(task); err != nil {
		t.Fatalf("failed to save task: %v", err)
	}
	return task
}
```

---

## Testing Checklist

Before considering this phase complete:

### Context Building Tests

- [ ] `BuildContextPrompt` includes task title and priority
- [ ] `BuildContextPrompt` includes all comments in order
- [ ] `BuildContextPrompt` formats authors correctly
- [ ] `BuildContextPrompt` handles empty comments
- [ ] `BuildContextPrompt` truncates long descriptions
- [ ] `BuildMinimalPrompt` only includes last 3 comments
- [ ] `BuildMinimalPrompt` truncates long comments

### Command Building Tests

- [ ] `BuildResumeCommand` works for opencode
- [ ] `BuildResumeCommand` works for claude-code
- [ ] `BuildResumeCommand` works for codex
- [ ] `BuildResumeCommand` fails for unknown tools
- [ ] Commands properly escape special characters in prompts
- [ ] Session refs are included correctly

### Resume Command Tests

- [ ] `resume` fails for task not in need_input
- [ ] `resume` fails for task without session
- [ ] `resume` prints command by default
- [ ] `resume --json` outputs valid JSON
- [ ] `resume --minimal` uses shorter prompt
- [ ] `resume --prompt` uses custom prompt
- [ ] `resume --exec --dry-run` shows command without running
- [ ] `resume --exec` changes directory and runs command

### Integration Tests

- [ ] Full workflow: block → comment → resume --exec
- [ ] Task moves to in_progress after resume --exec
- [ ] History updated after resume
- [ ] Session status updated after resume

---

## Files Changed/Created

| File | Change Type | Description |
|------|-------------|-------------|
| `internal/resume/context.go` | New | Context prompt builder |
| `internal/resume/command.go` | New | Resume command builder |
| `internal/resume/context_test.go` | New | Context tests |
| `internal/resume/command_test.go` | New | Command tests |
| `internal/commands/resume.go` | New | Resume CLI command |
| `internal/commands/resume_test.go` | New | Resume command tests |
| `internal/commands/root.go` | Modified | Register resume command |
| `go.mod` | Modified | Add shellescape dependency |

---

## Notes for Implementer

1. **Shell escaping is critical**: The prompt may contain any characters. Use proper escaping to prevent shell injection.

2. **Working directory**: Always execute the resume command in the session's working directory.

3. **Task state update**: Update the task to `in_progress` BEFORE executing the command, so if the agent crashes, the task is still in the right state.

4. **Prompt length**: Consider adding a warning if the prompt is very long (> 10k chars) as it may consume many tokens.

5. **Error handling**: If the tool command fails to start, show a helpful error message.

6. **Cross-platform**: The `exec.Command` approach should work on Linux/macOS. Windows may need adjustments.
