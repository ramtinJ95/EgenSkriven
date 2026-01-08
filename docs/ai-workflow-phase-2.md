# Phase 2: Session Management

> **Parent Document**: [ai-workflow-plan.md](./ai-workflow-plan.md)  
> **Phase**: 2 of 7  
> **Status**: Not Started  
> **Estimated Effort**: 2-3 days  
> **Prerequisites**: [Phase 1](./ai-workflow-phase-1.md) completed

## Overview

This phase implements session management - the ability to link AI tool sessions to EgenSkriven tasks and track session history. This is essential for the resume functionality in Phase 3.

**What we're building:**
- `egenskriven session link <task> --tool X --ref Y` - Link a session to a task
- `egenskriven session show <task>` - Show current session for a task
- `egenskriven session history <task>` - Show session history for a task
- Session lifecycle management (active, paused, completed, abandoned)

**What we're NOT building yet:**
- Resume command (Phase 3)
- Tool integrations for auto-discovery (Phase 4)
- UI for session info (Phase 5)

---

## Prerequisites

Before starting this phase:

1. Phase 1 is complete (block, comment, comments commands work)
2. `sessions` collection exists (from Phase 0)
3. `agent_session` field exists on tasks (from Phase 0)
4. Understand the session data model from the plan document

---

## Implementation Checklist

Use this checklist to track progress. Complete tasks in order from top to bottom.

### Phase Prerequisites (Verify First)

- [x] **PREREQ-1**: Verify Phase 1 is complete - run `egenskriven block`, `egenskriven comment`, `egenskriven comments` commands
- [x] **PREREQ-2**: Verify `sessions` collection exists in PocketBase schema
- [x] **PREREQ-3**: Verify `agent_session` JSON field exists on tasks collection

### Step 1: Create Session Command Group

| ID | Task | File | Priority | Status |
|----|------|------|----------|--------|
| 2.1 | Create session command group | `internal/commands/session.go` | High | [x] |
| 2.1a | Register session command in root.go | `internal/commands/root.go` | High | [x] |

**Verification**: `egenskriven session --help` should show subcommands

### Step 2: Implement Session Link Command

| ID | Task | File | Priority | Status |
|----|------|------|----------|--------|
| 2.2 | Create session link command | `internal/commands/session.go` | High | [x] |
| 2.2a | Implement tool validation (opencode, claude-code, codex) | `session.go` | High | [x] |
| 2.2b | Implement ref type detection (uuid vs path) | `session.go` | Medium | [x] |
| 2.2c | Implement transaction for session linking | `session.go` | High | [x] |
| 2.2d | Handle existing session replacement (mark old as abandoned) | `session.go` | High | [x] |

**Verification**:
```bash
egenskriven add "Test session" --type feature
egenskriven session link WRK-1 --tool opencode --ref test-123
egenskriven session link WRK-1 --tool opencode --ref test-123 --json
```

### Step 3: Implement Session Show Command

| ID | Task | File | Priority | Status |
|----|------|------|----------|--------|
| 2.3 | Create session show command | `internal/commands/session.go` | High | [x] |
| 2.3a | Handle case when no session is linked | `session.go` | Medium | [x] |
| 2.3b | Show resume hint for need_input tasks | `session.go` | Low | [x] |

**Verification**:
```bash
egenskriven session show WRK-1           # Should show session details
egenskriven session show WRK-1 --json    # Should output valid JSON
```

### Step 4: Implement Session History Command

| ID | Task | File | Priority | Status |
|----|------|------|----------|--------|
| 2.4 | Create session history command | `internal/commands/session.go` | High | [x] |
| 2.4a | Query sessions collection by task ID | `session.go` | Medium | [x] |
| 2.4b | Format session status with icons/labels | `session.go` | Low | [x] |

**Verification**:
```bash
# Link multiple sessions to create history
egenskriven session link WRK-1 --tool claude-code --ref session-2
egenskriven session link WRK-1 --tool codex --ref session-3
egenskriven session history WRK-1        # Should show 3 sessions
egenskriven session history WRK-1 --json
```

### Step 5: Implement Session Unlink Command (Optional)

| ID | Task | File | Priority | Status |
|----|------|------|----------|--------|
| 2.5 | Create session unlink command | `internal/commands/session.go` | Low | [x] |

**Verification**:
```bash
egenskriven session unlink WRK-1
egenskriven session unlink WRK-1 --status completed
```

### Step 6: Update Show Command

| ID | Task | File | Priority | Status |
|----|------|------|----------|--------|
| 2.6 | Display session info in task show command | `internal/commands/show.go` | Medium | [ ] |

**Verification**:
```bash
egenskriven show WRK-1    # Should include "Agent Session" section
```

### Step 7: Write Unit Tests

| ID | Task | File | Priority | Status |
|----|------|------|----------|--------|
| 2.7 | Create test file | `internal/commands/session_test.go` | High | [ ] |
| 2.7a | Test: session link creates session on task | `session_test.go` | Medium | [ ] |
| 2.7b | Test: session link creates record in sessions table | `session_test.go` | Medium | [ ] |
| 2.7c | Test: session link replaces existing (marks old as abandoned) | `session_test.go` | Medium | [ ] |
| 2.7d | Test: session link fails with invalid tool name | `session_test.go` | Medium | [ ] |
| 2.7e | Test: session link --json outputs valid JSON | `session_test.go` | Medium | [ ] |
| 2.7f | Test: session show displays "no session" for task without session | `session_test.go` | Medium | [ ] |
| 2.7g | Test: session show displays session details correctly | `session_test.go` | Medium | [ ] |
| 2.7h | Test: session history shows all sessions for a task | `session_test.go` | Medium | [ ] |
| 2.7i | Test: determineRefType function | `session_test.go` | Low | [ ] |

**Verification**:
```bash
go test ./internal/commands/... -run TestSession -v
```

### Step 8: Integration Testing

| ID | Task | Priority | Status |
|----|------|----------|--------|
| 2.8 | Full workflow test: link -> show -> link again -> history shows both | High | [ ] |

**Verification Script**:
```bash
#!/bin/bash
# Run this script to verify the full phase 2 implementation

echo "=== Phase 2 Integration Test ==="

# Create test task
egenskriven add "Integration test task" --type feature
TASK_ID=$(egenskriven list --json | jq -r '.tasks[0].display_id')

echo "Created task: $TASK_ID"

# Test 1: Link first session
echo "Test 1: Linking first session..."
egenskriven session link $TASK_ID --tool opencode --ref session-001
egenskriven session show $TASK_ID

# Test 2: Link second session (should mark first as abandoned)
echo "Test 2: Linking second session..."
egenskriven session link $TASK_ID --tool claude-code --ref session-002

# Test 3: Check history
echo "Test 3: Checking session history..."
egenskriven session history $TASK_ID

# Test 4: Verify JSON outputs
echo "Test 4: Verifying JSON outputs..."
egenskriven session show $TASK_ID --json | jq .
egenskriven session history $TASK_ID --json | jq .

# Test 5: Task show includes session
echo "Test 5: Task show includes session info..."
egenskriven show $TASK_ID

echo "=== All tests complete ==="
```

---

## Summary

| Category | Count |
|----------|-------|
| Prerequisites | 3 |
| High Priority Tasks | 14 |
| Medium Priority Tasks | 10 |
| Low Priority Tasks | 5 |
| **Total** | **32** |

**Files to Create**:
- `internal/commands/session.go`
- `internal/commands/session_link.go`
- `internal/commands/session_show.go`
- `internal/commands/session_history.go`
- `internal/commands/session_unlink.go` (optional)
- `internal/commands/session_test.go`

**Files to Modify**:
- `internal/commands/root.go`
- `internal/commands/show.go`

---

## Data Model Recap

### `agent_session` Field on Tasks (Current Session)

```typescript
interface AgentSession {
    tool: "opencode" | "claude-code" | "codex";
    ref: string;           // Session/thread ID
    ref_type: "uuid" | "path";
    working_dir: string;   // Project directory
    linked_at: string;     // ISO timestamp
}
```

### `sessions` Collection (Session History)

| Field | Type | Description |
|-------|------|-------------|
| `task` | relation | Link to task |
| `tool` | select | opencode, claude-code, codex |
| `external_ref` | text | Session ID from the tool |
| `ref_type` | select | uuid or path |
| `working_dir` | text | Project directory |
| `status` | select | active, paused, completed, abandoned |
| `created` | autodate | When linked |
| `ended_at` | date | When session ended |

---

## Tasks

### Task 2.1: Create Session Subcommand Group

**File**: `internal/commands/session.go`

**Description**: Create the `session` command group with subcommands.

**Usage**:
```bash
egenskriven session link <task> --tool opencode --ref <id>
egenskriven session show <task>
egenskriven session history <task>
```

**Implementation**:

```go
package commands

import (
	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"
)

func newSessionCmd(app *pocketbase.PocketBase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage agent sessions linked to tasks",
		Long: `Commands for linking and viewing AI agent sessions on tasks.

Sessions allow tracking which AI tool (OpenCode, Claude Code, Codex) is 
working on a task, enabling context-preserving resume when a task is blocked.`,
	}

	// Add subcommands
	cmd.AddCommand(newSessionLinkCmd(app))
	cmd.AddCommand(newSessionShowCmd(app))
	cmd.AddCommand(newSessionHistoryCmd(app))
	cmd.AddCommand(newSessionUnlinkCmd(app))

	return cmd
}
```

**Register in root.go**:
```go
rootCmd.AddCommand(newSessionCmd(app))
```

---

### Task 2.2: Implement `session link` Command

**File**: `internal/commands/session_link.go`

**Description**: Link an agent session to a task. If a session is already linked, mark the old one as abandoned and link the new one.

**Usage**:
```bash
egenskriven session link <task-ref> --tool opencode --ref abc-123-def
egenskriven session link WRK-123 --tool claude-code --ref 550e8400-e29b-41d4-a716-446655440000
egenskriven session link WRK-123 --tool codex --ref xyz-789 --working-dir /path/to/project
```

**Flags**:
| Flag | Short | Type | Required | Default | Description |
|------|-------|------|----------|---------|-------------|
| `--tool` | `-t` | string | yes | | Tool name (opencode, claude-code, codex) |
| `--ref` | `-r` | string | yes | | Session/thread ID |
| `--working-dir` | `-w` | string | no | cwd | Working directory |
| `--json` | `-j` | bool | no | false | Output as JSON |

**Implementation**:

```go
package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
)

var validTools = []string{"opencode", "claude-code", "codex"}

func newSessionLinkCmd(app *pocketbase.PocketBase) *cobra.Command {
	var tool, ref, workingDir string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "link <task-ref>",
		Short: "Link an agent session to a task",
		Long: `Link an AI agent session to a task for tracking and resume support.

If a session is already linked to the task, it will be marked as "abandoned"
and moved to history before linking the new session.

The session reference is typically a UUID (for OpenCode/Claude Code/Codex) 
or a file path (for some tools).`,
		Example: `  # Link an OpenCode session
  egenskriven session link WRK-123 --tool opencode --ref abc-123-def
  
  # Link a Claude Code session with explicit working directory
  egenskriven session link WRK-123 --tool claude-code --ref 550e8400-... --working-dir /home/user/project
  
  # Link and output JSON
  egenskriven session link WRK-123 -t codex -r xyz-789 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskRef := args[0]

			// Validate tool
			if !contains(validTools, tool) {
				return fmt.Errorf("invalid tool %q: must be one of %v", tool, validTools)
			}

			// Validate ref
			if ref == "" {
				return fmt.Errorf("--ref is required")
			}

			// Resolve working directory
			if workingDir == "" {
				var err error
				workingDir, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get working directory: %w", err)
				}
			}
			// Make absolute
			workingDir, _ = filepath.Abs(workingDir)

			// Determine ref type
			refType := determineRefType(ref)

			// Resolve task
			task, err := resolver.MustResolve(app, taskRef)
			if err != nil {
				return err
			}

			displayId := getDisplayId(task)
			now := time.Now()

			// Get sessions collection
			sessionsCollection, err := app.FindCollectionByNameOrId("sessions")
			if err != nil {
				return fmt.Errorf("sessions collection not found: %w", err)
			}

			// Execute in transaction
			var sessionId string
			err = app.RunInTransaction(func(txApp core.App) error {
				// Check for existing session
				existingSession := task.Get("agent_session")
				if existingSession != nil {
					oldSession := existingSession.(map[string]any)
					
					// Mark old session as abandoned in history
					err := markSessionStatus(txApp, sessionsCollection, task.Id, 
						oldSession["ref"].(string), "abandoned", now)
					if err != nil {
						// Log but don't fail - old session might not be in history
						fmt.Fprintf(os.Stderr, "Warning: could not update old session status: %v\n", err)
					}
				}

				// Create new agent_session on task
				newSession := map[string]any{
					"tool":        tool,
					"ref":         ref,
					"ref_type":    refType,
					"working_dir": workingDir,
					"linked_at":   now.Format(time.RFC3339),
				}
				task.Set("agent_session", newSession)

				// Add to history
				addHistoryEntry(task, "session_linked", "agent", map[string]any{
					"tool":        tool,
					"session_ref": ref,
				})

				if err := txApp.Save(task); err != nil {
					return fmt.Errorf("failed to update task: %w", err)
				}

				// Create session record in sessions table
				sessionRecord := core.NewRecord(sessionsCollection)
				sessionRecord.Set("task", task.Id)
				sessionRecord.Set("tool", tool)
				sessionRecord.Set("external_ref", ref)
				sessionRecord.Set("ref_type", refType)
				sessionRecord.Set("working_dir", workingDir)
				sessionRecord.Set("status", "active")

				if err := txApp.Save(sessionRecord); err != nil {
					return fmt.Errorf("failed to create session record: %w", err)
				}

				sessionId = sessionRecord.Id
				return nil
			})

			if err != nil {
				return err
			}

			// Output
			if jsonOutput {
				return outputJSON(cmd.OutOrStdout(), map[string]any{
					"success":     true,
					"task_id":     task.Id,
					"display_id":  displayId,
					"session_id":  sessionId,
					"tool":        tool,
					"ref":         ref,
					"ref_type":    refType,
					"working_dir": workingDir,
				})
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Session linked to %s\n", displayId)
			fmt.Fprintf(cmd.OutOrStdout(), "  Tool: %s\n", tool)
			fmt.Fprintf(cmd.OutOrStdout(), "  Ref:  %s\n", truncateMiddle(ref, 40))
			return nil
		},
	}

	cmd.Flags().StringVarP(&tool, "tool", "t", "", "Tool name (opencode, claude-code, codex)")
	cmd.Flags().StringVarP(&ref, "ref", "r", "", "Session/thread reference")
	cmd.Flags().StringVarP(&workingDir, "working-dir", "w", "", "Working directory (defaults to current)")
	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
	
	cmd.MarkFlagRequired("tool")
	cmd.MarkFlagRequired("ref")

	return cmd
}

// determineRefType guesses if ref is a UUID or path
func determineRefType(ref string) string {
	// If it starts with / or . or contains path separators, treat as path
	if strings.HasPrefix(ref, "/") || strings.HasPrefix(ref, ".") || 
	   strings.Contains(ref, "/") || strings.Contains(ref, "\\") {
		return "path"
	}
	return "uuid"
}

// markSessionStatus updates a session's status in the history table
func markSessionStatus(app core.App, collection *core.Collection, taskId, externalRef, status string, endTime time.Time) error {
	// Find the session record
	records, err := app.FindRecordsByFilter(
		"sessions",
		fmt.Sprintf("task = '%s' && external_ref = '%s' && status = 'active'", taskId, externalRef),
		"-created",
		1,
		0,
	)
	if err != nil || len(records) == 0 {
		return fmt.Errorf("session not found")
	}

	record := records[0]
	record.Set("status", status)
	record.Set("ended_at", endTime)
	
	return app.Save(record)
}

// truncateMiddle truncates a string in the middle if too long
func truncateMiddle(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	half := (maxLen - 3) / 2
	return s[:half] + "..." + s[len(s)-half:]
}

// contains checks if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
```

**Verification**:
```bash
# Create a task
egenskriven add "Test session linking" --type feature

# Link a session
egenskriven session link WRK-1 --tool opencode --ref test-session-123

# Verify task has session
egenskriven session show WRK-1
# Should show linked session

# Link another session (should mark first as abandoned)
egenskriven session link WRK-1 --tool claude-code --ref new-session-456

# Verify history shows both
egenskriven session history WRK-1
# Should show abandoned session and active session
```

---

### Task 2.3: Implement `session show` Command

**File**: `internal/commands/session_show.go`

**Description**: Display the currently linked session for a task.

**Usage**:
```bash
egenskriven session show <task-ref>
egenskriven session show WRK-123 --json
```

**Implementation**:

```go
package commands

import (
	"fmt"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"
)

func newSessionShowCmd(app *pocketbase.PocketBase) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "show <task-ref>",
		Short: "Show the current session linked to a task",
		Long: `Display information about the AI agent session currently linked to a task.

This shows which tool is working on the task, the session reference,
and when it was linked.`,
		Example: `  egenskriven session show WRK-123
  egenskriven session show abc --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskRef := args[0]

			// Resolve task
			task, err := resolver.MustResolve(app, taskRef)
			if err != nil {
				return err
			}

			displayId := getDisplayId(task)
			sessionData := task.Get("agent_session")

			if sessionData == nil {
				if jsonOutput {
					return outputJSON(cmd.OutOrStdout(), map[string]any{
						"task_id":     task.Id,
						"display_id":  displayId,
						"has_session": false,
						"session":     nil,
					})
				}
				fmt.Fprintf(cmd.OutOrStdout(), "No session linked to %s\n", displayId)
				return nil
			}

			session := sessionData.(map[string]any)

			if jsonOutput {
				return outputJSON(cmd.OutOrStdout(), map[string]any{
					"task_id":     task.Id,
					"display_id":  displayId,
					"has_session": true,
					"session":     session,
				})
			}

			// Human-readable output
			fmt.Fprintf(cmd.OutOrStdout(), "Session for %s:\n\n", displayId)
			fmt.Fprintf(cmd.OutOrStdout(), "  Tool:        %s\n", session["tool"])
			fmt.Fprintf(cmd.OutOrStdout(), "  Reference:   %s\n", session["ref"])
			fmt.Fprintf(cmd.OutOrStdout(), "  Type:        %s\n", session["ref_type"])
			fmt.Fprintf(cmd.OutOrStdout(), "  Working Dir: %s\n", session["working_dir"])
			
			if linkedAt, ok := session["linked_at"].(string); ok {
				t, _ := time.Parse(time.RFC3339, linkedAt)
				fmt.Fprintf(cmd.OutOrStdout(), "  Linked:      %s\n", formatRelativeTime(t))
			}

			// Show resume hint if task is in need_input
			if task.GetString("column") == "need_input" {
				fmt.Fprintf(cmd.OutOrStdout(), "\n  Tip: Run 'egenskriven resume %s' to resume this session\n", displayId)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")

	return cmd
}
```

**Verification**:
```bash
# Task without session
egenskriven add "No session" --type feature
egenskriven session show WRK-2
# Should say "No session linked"

# Task with session
egenskriven session show WRK-1
# Should show session details

# JSON output
egenskriven session show WRK-1 --json
```

---

### Task 2.4: Implement `session history` Command

**File**: `internal/commands/session_history.go`

**Description**: Show all sessions that have been linked to a task over time.

**Usage**:
```bash
egenskriven session history <task-ref>
egenskriven session history WRK-123 --json
```

**Implementation**:

```go
package commands

import (
	"fmt"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"
)

func newSessionHistoryCmd(app *pocketbase.PocketBase) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "history <task-ref>",
		Short: "Show session history for a task",
		Long: `Display all AI agent sessions that have been linked to a task,
including their status (active, abandoned, completed).

This helps track which tools have worked on a task over time.`,
		Example: `  egenskriven session history WRK-123
  egenskriven session history abc --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskRef := args[0]

			// Resolve task
			task, err := resolver.MustResolve(app, taskRef)
			if err != nil {
				return err
			}

			displayId := getDisplayId(task)

			// Fetch all sessions for this task
			records, err := app.FindRecordsByFilter(
				"sessions",
				fmt.Sprintf("task = '%s'", task.Id),
				"-created", // Most recent first
				100,        // Reasonable limit
				0,
			)
			if err != nil {
				return fmt.Errorf("failed to fetch sessions: %w", err)
			}

			if jsonOutput {
				sessions := make([]map[string]any, len(records))
				for i, r := range records {
					sessions[i] = map[string]any{
						"id":           r.Id,
						"tool":         r.GetString("tool"),
						"external_ref": r.GetString("external_ref"),
						"ref_type":     r.GetString("ref_type"),
						"working_dir":  r.GetString("working_dir"),
						"status":       r.GetString("status"),
						"created":      r.GetDateTime("created").Time().Format(time.RFC3339),
						"ended_at":     formatEndedAt(r),
					}
				}
				return outputJSON(cmd.OutOrStdout(), map[string]any{
					"task_id":    task.Id,
					"display_id": displayId,
					"count":      len(sessions),
					"sessions":   sessions,
				})
			}

			// Human-readable output
			if len(records) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No session history for %s\n", displayId)
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Session history for %s (%d sessions):\n\n", 
				displayId, len(records))

			for i, r := range records {
				status := r.GetString("status")
				tool := r.GetString("tool")
				ref := r.GetString("external_ref")
				created := r.GetDateTime("created").Time()

				// Status indicator
				statusIcon := getStatusIcon(status)

				fmt.Fprintf(cmd.OutOrStdout(), "%d. %s %s (%s)\n", 
					i+1, statusIcon, tool, status)
				fmt.Fprintf(cmd.OutOrStdout(), "   Ref: %s\n", truncateMiddle(ref, 50))
				fmt.Fprintf(cmd.OutOrStdout(), "   Started: %s\n", formatRelativeTime(created))
				
				if status != "active" {
					endedAt := r.GetDateTime("ended_at")
					if !endedAt.IsZero() {
						fmt.Fprintf(cmd.OutOrStdout(), "   Ended: %s\n", 
							formatRelativeTime(endedAt.Time()))
					}
				}
				fmt.Fprintln(cmd.OutOrStdout())
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")

	return cmd
}

func getStatusIcon(status string) string {
	switch status {
	case "active":
		return "[ACTIVE]"
	case "paused":
		return "[PAUSED]"
	case "completed":
		return "[DONE]  "
	case "abandoned":
		return "[OLD]   "
	default:
		return "[?]     "
	}
}

func formatEndedAt(r *core.Record) *string {
	endedAt := r.GetDateTime("ended_at")
	if endedAt.IsZero() {
		return nil
	}
	s := endedAt.Time().Format(time.RFC3339)
	return &s
}
```

**Verification**:
```bash
# Link multiple sessions to simulate history
egenskriven add "Test history" --type feature
egenskriven session link WRK-3 --tool opencode --ref session-1
egenskriven session link WRK-3 --tool claude-code --ref session-2
egenskriven session link WRK-3 --tool codex --ref session-3

# View history
egenskriven session history WRK-3
# Should show 3 sessions: 2 abandoned, 1 active

# JSON output
egenskriven session history WRK-3 --json
```

---

### Task 2.5: Implement `session unlink` Command (Optional)

**File**: `internal/commands/session_unlink.go`

**Description**: Manually unlink a session from a task without linking a new one.

**Usage**:
```bash
egenskriven session unlink <task-ref>
egenskriven session unlink WRK-123 --status completed
```

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--status` | string | "abandoned" | Final status (abandoned, completed) |
| `--json` | bool | false | Output as JSON |

**Implementation**:

```go
package commands

import (
	"fmt"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
)

func newSessionUnlinkCmd(app *pocketbase.PocketBase) *cobra.Command {
	var status string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "unlink <task-ref>",
		Short: "Unlink the current session from a task",
		Long: `Remove the current session link from a task, optionally marking it
with a final status (abandoned or completed).

Use this when you want to disconnect a session without linking a new one.`,
		Example: `  # Mark session as abandoned (default)
  egenskriven session unlink WRK-123
  
  # Mark session as completed
  egenskriven session unlink WRK-123 --status completed`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskRef := args[0]

			// Validate status
			validStatuses := []string{"abandoned", "completed"}
			if !contains(validStatuses, status) {
				return fmt.Errorf("invalid status %q: must be 'abandoned' or 'completed'", status)
			}

			// Resolve task
			task, err := resolver.MustResolve(app, taskRef)
			if err != nil {
				return err
			}

			displayId := getDisplayId(task)
			sessionData := task.Get("agent_session")

			if sessionData == nil {
				return fmt.Errorf("no session linked to %s", displayId)
			}

			session := sessionData.(map[string]any)
			now := time.Now()

			// Get sessions collection for history update
			sessionsCollection, err := app.FindCollectionByNameOrId("sessions")
			if err != nil {
				return fmt.Errorf("sessions collection not found: %w", err)
			}

			err = app.RunInTransaction(func(txApp core.App) error {
				// Update session in history table
				markSessionStatus(txApp, sessionsCollection, task.Id, 
					session["ref"].(string), status, now)

				// Clear agent_session on task
				task.Set("agent_session", nil)

				// Add history entry
				addHistoryEntry(task, "session_unlinked", "user", map[string]any{
					"tool":         session["tool"],
					"session_ref":  session["ref"],
					"final_status": status,
				})

				return txApp.Save(task)
			})

			if err != nil {
				return err
			}

			if jsonOutput {
				return outputJSON(cmd.OutOrStdout(), map[string]any{
					"success":      true,
					"task_id":      task.Id,
					"display_id":   displayId,
					"final_status": status,
				})
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Session unlinked from %s (marked as %s)\n", 
				displayId, status)
			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "abandoned", "Final status (abandoned, completed)")
	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")

	return cmd
}
```

---

### Task 2.6: Update `show` Command to Display Session Info

**File**: `internal/commands/show.go` (modify existing)

**Description**: When showing a task, include session information if present.

**Add to existing show command output**:

```go
// In the show command's RunE function, after displaying other task info:

// Display session info if present
sessionData := task.Get("agent_session")
if sessionData != nil {
    session := sessionData.(map[string]any)
    fmt.Fprintf(cmd.OutOrStdout(), "\nAgent Session:\n")
    fmt.Fprintf(cmd.OutOrStdout(), "  Tool: %s\n", session["tool"])
    fmt.Fprintf(cmd.OutOrStdout(), "  Ref:  %s\n", truncateMiddle(session["ref"].(string), 40))
    
    if task.GetString("column") == "need_input" {
        fmt.Fprintf(cmd.OutOrStdout(), "  (Use 'egenskriven resume %s' to continue)\n", displayId)
    }
}
```

For JSON output, add to the output map:
```go
output["agent_session"] = task.Get("agent_session")
```

---

### Task 2.7: Write Unit Tests

**File**: `internal/commands/session_test.go`

```go
package commands

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestSessionLinkCommand(t *testing.T) {
	app := setupTestApp(t)
	defer cleanupTestApp(t, app)

	task := createTestTask(t, app, "Test task")

	tests := []struct {
		name      string
		args      []string
		wantErr   bool
		errContains string
	}{
		{
			name:    "link with all flags",
			args:    []string{"link", task.Id, "--tool", "opencode", "--ref", "test-123"},
			wantErr: false,
		},
		{
			name:      "link without tool",
			args:      []string{"link", task.Id, "--ref", "test-123"},
			wantErr:   true,
			errContains: "required",
		},
		{
			name:      "link without ref",
			args:      []string{"link", task.Id, "--tool", "opencode"},
			wantErr:   true,
			errContains: "required",
		},
		{
			name:      "link with invalid tool",
			args:      []string{"link", task.Id, "--tool", "invalid", "--ref", "test-123"},
			wantErr:   true,
			errContains: "invalid tool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear session before each test
			task.Set("agent_session", nil)
			app.Save(task)

			cmd := newSessionCmd(app)
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				// Verify session was linked
				refreshed, _ := app.FindRecordById("tasks", task.Id)
				session := refreshed.Get("agent_session")
				if session == nil {
					t.Error("session should be linked")
				}
			}
		})
	}
}

func TestSessionLinkReplacesExisting(t *testing.T) {
	app := setupTestApp(t)
	defer cleanupTestApp(t, app)

	task := createTestTask(t, app, "Test task")

	// Link first session
	cmd := newSessionCmd(app)
	cmd.SetArgs([]string{"link", task.Id, "--tool", "opencode", "--ref", "session-1"})
	cmd.Execute()

	// Link second session
	cmd = newSessionCmd(app)
	cmd.SetArgs([]string{"link", task.Id, "--tool", "claude-code", "--ref", "session-2"})
	cmd.Execute()

	// Verify current session is session-2
	refreshed, _ := app.FindRecordById("tasks", task.Id)
	session := refreshed.Get("agent_session").(map[string]any)
	
	if session["ref"] != "session-2" {
		t.Errorf("expected session-2, got %s", session["ref"])
	}
	if session["tool"] != "claude-code" {
		t.Errorf("expected claude-code, got %s", session["tool"])
	}

	// Verify session-1 is in history as abandoned
	records, _ := app.FindRecordsByFilter(
		"sessions",
		fmt.Sprintf("task = '%s' && external_ref = 'session-1'", task.Id),
		"",
		1,
		0,
	)
	if len(records) == 0 {
		t.Error("session-1 should be in history")
	} else if records[0].GetString("status") != "abandoned" {
		t.Errorf("session-1 should be abandoned, got %s", records[0].GetString("status"))
	}
}

func TestSessionShowCommand(t *testing.T) {
	app := setupTestApp(t)
	defer cleanupTestApp(t, app)

	task := createTestTask(t, app, "Test task")

	// Test without session
	cmd := newSessionCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"show", task.Id})
	cmd.Execute()

	if !strings.Contains(out.String(), "No session") {
		t.Error("should indicate no session")
	}

	// Link a session
	cmd = newSessionCmd(app)
	cmd.SetArgs([]string{"link", task.Id, "--tool", "opencode", "--ref", "test-123"})
	cmd.Execute()

	// Test with session
	cmd = newSessionCmd(app)
	out.Reset()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"show", task.Id})
	cmd.Execute()

	if !strings.Contains(out.String(), "opencode") {
		t.Error("should show tool name")
	}
	if !strings.Contains(out.String(), "test-123") {
		t.Error("should show session ref")
	}
}

func TestSessionHistoryCommand(t *testing.T) {
	app := setupTestApp(t)
	defer cleanupTestApp(t, app)

	task := createTestTask(t, app, "Test task")

	// Link multiple sessions
	for i := 1; i <= 3; i++ {
		cmd := newSessionCmd(app)
		cmd.SetArgs([]string{"link", task.Id, "--tool", "opencode", 
			"--ref", fmt.Sprintf("session-%d", i)})
		cmd.Execute()
	}

	// Check history
	cmd := newSessionCmd(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"history", task.Id, "--json"})
	cmd.Execute()

	var result map[string]any
	json.Unmarshal(out.Bytes(), &result)

	count := int(result["count"].(float64))
	if count != 3 {
		t.Errorf("expected 3 sessions in history, got %d", count)
	}
}

func TestDetermineRefType(t *testing.T) {
	tests := []struct {
		ref      string
		expected string
	}{
		{"550e8400-e29b-41d4-a716-446655440000", "uuid"},
		{"abc123def456", "uuid"},
		{"/home/user/.aider/history.md", "path"},
		{"./relative/path", "path"},
		{"C:\\Users\\path", "path"},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			got := determineRefType(tt.ref)
			if got != tt.expected {
				t.Errorf("determineRefType(%q) = %q, want %q", tt.ref, got, tt.expected)
			}
		})
	}
}
```

---

## Final Verification Checklist

Before marking this phase complete, verify all items pass:

### Session Link Tests

- [ ] `session link` creates session on task
- [ ] `session link` creates record in sessions table
- [ ] `session link` replaces existing session (marks old as abandoned)
- [ ] `session link` fails with invalid tool name
- [ ] `session link` fails without required flags
- [ ] `session link --json` outputs valid JSON
- [ ] `session link` uses cwd when no --working-dir specified

### Session Show Tests

- [ ] `session show` displays "no session" for task without session
- [ ] `session show` displays session details correctly
- [ ] `session show --json` outputs valid JSON
- [ ] `session show` shows resume hint for need_input tasks

### Session History Tests

- [ ] `session history` shows all sessions for a task
- [ ] `session history` shows correct status for each session
- [ ] `session history` orders by most recent first
- [ ] `session history --json` outputs valid JSON
- [ ] `session history` shows helpful message when no history

### Integration Tests

- [ ] Full workflow: link → show → link again → history shows both
- [ ] Session status transitions work correctly
- [ ] History entries are created for session changes

### Unit Tests Pass

- [ ] `go test ./internal/commands/... -run TestSession -v` passes

---

## Files Changed/Created Summary

| File | Change Type | Description |
|------|-------------|-------------|
| `internal/commands/session.go` | New | Session command group |
| `internal/commands/session_link.go` | New | Link subcommand |
| `internal/commands/session_show.go` | New | Show subcommand |
| `internal/commands/session_history.go` | New | History subcommand |
| `internal/commands/session_unlink.go` | New | Unlink subcommand (optional) |
| `internal/commands/show.go` | Modified | Display session info |
| `internal/commands/root.go` | Modified | Register session command |
| `internal/commands/session_test.go` | New | Session command tests |

---

## Phase Completion

Once all checklist items are verified:

1. Update this document's status from "Not Started" to "Complete"
2. Commit all changes with message: `feat(cli): implement session management (Phase 2)`
3. Proceed to [Phase 3: Resume Flow](./ai-workflow-phase-3.md)

---

## Next Phase

Once all tests pass, proceed to [Phase 3: Resume Flow](./ai-workflow-phase-3.md).

Phase 3 will implement:
- `egenskriven resume` command (print mode)
- `egenskriven resume --exec` (execute mode)
- Context prompt builder
- Tool-specific resume command generation

---

## Notes for Implementer

1. **Session atomicity**: When linking a new session, both the task update AND the session record creation must succeed together.

2. **Abandoned sessions**: When replacing a session, the old one should be marked as "abandoned" in the history table, not deleted.

3. **Ref type detection**: The `determineRefType` function is a heuristic. It may need refinement based on real-world session IDs.

4. **Working directory**: Always store as absolute path to avoid issues when resuming from a different location.

5. **Display IDs**: Use the same display ID format as other commands (WRK-123 style).
