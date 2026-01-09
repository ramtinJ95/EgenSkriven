# Phase 6: Auto-Resume

> **Parent Document**: [ai-workflow-plan.md](./ai-workflow-plan.md)
> **Phase**: 6
> **Status**: Not Started
> **Prerequisites**: [Phase 5](./ai-workflow-phase-5.md) completed

---

## Phase 6 Todo List

### Backend Implementation

- [x] **Task 6.1: Create Auto-Resume Service**
  - [x] 6.1.1: Create `internal/autoresume/service.go`
  - [x] 6.1.2: Implement `NewService()` constructor
  - [x] 6.1.3: Implement `CheckAndResume()` method with all condition checks
  - [x] 6.1.4: Implement `triggerResume()` method
  - [x] 6.1.5: Implement `updateTaskForResume()` method
  - [x] 6.1.6: Implement `executeResume()` method for background execution
  - [x] 6.1.7: Implement `fetchComments()` helper
  - [x] 6.1.8: Implement `hasAgentMention()` helper
  - [x] 6.1.9: Implement `logAutoResume()` for debugging/logging
  - [x] 6.1.10: Implement `ensureHistorySlice()` helper

- [ ] **Task 6.2: Register Comment Hook**
  - [ ] 6.2.1: Create `internal/hooks/comments.go`
  - [ ] 6.2.2: Implement `RegisterCommentHooks()` function
  - [ ] 6.2.3: Add `OnRecordAfterCreateSuccess` hook for comments collection
  - [ ] 6.2.4: Ensure hook runs auto-resume check in goroutine (non-blocking)
  - [ ] 6.2.5: Register hook in main/app initialization

- [ ] **Task 6.3: Add Board Resume Mode Configuration**
  - [ ] 6.3.1: Update `internal/commands/board.go` with `--resume-mode` flag
  - [ ] 6.3.2: Add resume mode validation (manual, command, auto)
  - [ ] 6.3.3: Update `board show` command to display resume mode
  - [ ] 6.3.4: Ensure default resume mode is "command"

### Frontend Implementation

- [ ] **Task 6.4: Add Resume Mode UI Component**
  - [ ] 6.4.1: Create `ui/src/components/BoardSettings/ResumeModeSelector.tsx`
  - [ ] 6.4.2: Implement radio button selection for manual/command/auto modes
  - [ ] 6.4.3: Add mode descriptions for each option
  - [ ] 6.4.4: Add warning indicator when auto mode is selected
  - [ ] 6.4.5: Integrate component into board settings view

- [ ] **Task 6.5: Update Comments Panel with Auto-Resume Indicator**
  - [ ] 6.5.1: Update `ui/src/components/TaskDetail/CommentsPanel.tsx`
  - [ ] 6.5.2: Add `boardResumeMode` and `taskColumn` props
  - [ ] 6.5.3: Implement auto-resume enabled indicator (green pulsing dot)
  - [ ] 6.5.4: Show "@agent to trigger" hint when auto-resume is enabled
  - [ ] 6.5.5: Add visual feedback for auto-resume capable state

### Testing

- [ ] **Task 6.6: Write Unit Tests**
  - [ ] 6.6.1: Create `internal/autoresume/service_test.go`
  - [ ] 6.6.2: Test: All conditions met triggers resume
  - [ ] 6.6.3: Test: Agent comments do NOT trigger resume
  - [ ] 6.6.4: Test: Comments without @agent do NOT trigger
  - [ ] 6.6.5: Test: Manual mode does NOT auto-resume
  - [ ] 6.6.6: Test: Command mode does NOT auto-resume
  - [ ] 6.6.7: Test: No session does NOT trigger
  - [ ] 6.6.8: Test: Wrong column (not need_input) does NOT trigger
  - [ ] 6.6.9: Test: `hasAgentMention()` helper function
  - [ ] 6.6.10: Create test helpers (createTestComment, etc.)

- [ ] **Task 6.7: Add E2E Tests**
  - [ ] 6.7.1: Create `tests/e2e/autoresume_test.go`
  - [ ] 6.7.2: Test full auto-resume workflow end-to-end
  - [ ] 6.7.3: Verify task moves to in_progress after trigger
  - [ ] 6.7.4: Verify history contains auto_resumed action

### Verification Checklist

#### Auto-Resume Logic
- [ ] 6.V.1: Triggers when ALL conditions are met
- [ ] 6.V.2: Does NOT trigger for agent comments
- [ ] 6.V.3: Does NOT trigger without @agent mention
- [ ] 6.V.4: Does NOT trigger when resume_mode is manual
- [ ] 6.V.5: Does NOT trigger when resume_mode is command
- [ ] 6.V.6: Does NOT trigger without linked session
- [ ] 6.V.7: Does NOT trigger when task not in need_input
- [ ] 6.V.8: Updates task column to in_progress
- [ ] 6.V.9: Adds history entry with auto_resumed action
- [ ] 6.V.10: Executes resume command in background

#### Board Configuration
- [ ] 6.V.11: `egenskriven board update --resume-mode` works
- [ ] 6.V.12: `egenskriven board show` displays resume mode
- [ ] 6.V.13: Invalid resume mode is rejected
- [ ] 6.V.14: Default resume mode is "command"

#### UI
- [ ] 6.V.15: Resume mode selector works in board settings
- [ ] 6.V.16: Auto-resume indicator shows in comments panel
- [ ] 6.V.17: @agent mention shows appropriate visual feedback

#### Error Handling
- [ ] 6.V.18: Failures logged but don't break comment creation
- [ ] 6.V.19: Invalid session data handled gracefully
- [ ] 6.V.20: Missing board handled gracefully

---

## Overview

This phase implements the auto-resume functionality - automatically resuming an agent session when a human adds a comment with `@agent` mention. This enables a hands-free workflow for responsive collaboration.

**What we're building:**
- `@agent` mention detection in comments
- Auto-resume trigger when conditions are met
- Board-level resume mode configuration
- UI for configuring resume mode

---

## Prerequisites

Before starting this phase:

1. Phase 5 is complete (UI components work)
2. Comments have `metadata.mentions` populated
3. Resume command works (`egenskriven resume --exec`)
4. Boards have `resume_mode` field

---

## Auto-Resume Conditions

For auto-resume to trigger, ALL conditions must be met:

1. Task is in `need_input` column
2. Task has a linked `agent_session`
3. Board's `resume_mode` is set to `auto`
4. Comment contains `@agent` mention
5. Comment is from a human (not from agent)

---

## Tasks

### Task 6.1: Create Auto-Resume Service

**File**: `internal/autoresume/service.go`

**Description**: Service that checks conditions and triggers resume.

```go
package autoresume

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	
	"egenskriven/internal/resume"
)

// Service handles auto-resume logic
type Service struct {
	app *pocketbase.PocketBase
}

// NewService creates a new auto-resume service
func NewService(app *pocketbase.PocketBase) *Service {
	return &Service{app: app}
}

// CheckAndResume evaluates whether a comment should trigger auto-resume
func (s *Service) CheckAndResume(comment *core.Record) error {
	// 1. Check if comment is from human
	if comment.GetString("author_type") != "human" {
		return nil // Agent comments don't trigger
	}

	// 2. Check for @agent mention
	if !hasAgentMention(comment) {
		return nil // No @agent mention
	}

	// 3. Get the task
	taskId := comment.GetString("task")
	task, err := s.app.FindRecordById("tasks", taskId)
	if err != nil {
		return fmt.Errorf("failed to find task: %w", err)
	}

	// 4. Check task is in need_input
	if task.GetString("column") != "need_input" {
		return nil // Task not blocked
	}

	// 5. Check task has session
	sessionData := task.Get("agent_session")
	if sessionData == nil {
		return nil // No session linked
	}

	// 6. Check board resume mode
	board, err := s.app.FindRecordById("boards", task.GetString("board"))
	if err != nil {
		return fmt.Errorf("failed to find board: %w", err)
	}

	resumeMode := board.GetString("resume_mode")
	if resumeMode != "auto" {
		return nil // Auto-resume not enabled
	}

	// All conditions met - trigger resume
	return s.triggerResume(task, comment)
}

// triggerResume executes the resume flow
func (s *Service) triggerResume(task *core.Record, triggerComment *core.Record) error {
	session := task.Get("agent_session").(map[string]any)
	
	// Fetch all comments for context
	comments, err := s.fetchComments(task.Id)
	if err != nil {
		return fmt.Errorf("failed to fetch comments: %w", err)
	}

	// Build context prompt
	prompt := resume.BuildContextPrompt(task, comments)

	// Build resume command
	resumeCmd, err := resume.BuildResumeCommand(
		session["tool"].(string),
		session["ref"].(string),
		session["working_dir"].(string),
		prompt,
	)
	if err != nil {
		return fmt.Errorf("failed to build resume command: %w", err)
	}

	// Update task state
	if err := s.updateTaskForResume(task, triggerComment); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// Execute resume in background
	go s.executeResume(resumeCmd, task.Id)

	return nil
}

// updateTaskForResume updates task state before resume
func (s *Service) updateTaskForResume(task *core.Record, triggerComment *core.Record) error {
	task.Set("column", "in_progress")

	// Add history entry
	history := ensureHistorySlice(task.Get("history"))
	history = append(history, map[string]any{
		"timestamp":    time.Now().Format(time.RFC3339),
		"action":       "auto_resumed",
		"actor":        "system",
		"actor_detail": "auto-resume",
		"changes": map[string]any{
			"column": map[string]any{
				"from": "need_input",
				"to":   "in_progress",
			},
		},
		"metadata": map[string]any{
			"trigger_comment": triggerComment.Id,
		},
	})
	task.Set("history", history)

	return s.app.Save(task)
}

// executeResume runs the resume command in background
func (s *Service) executeResume(resumeCmd *resume.ResumeCommand, taskId string) {
	// Log start
	s.logAutoResume(taskId, "started", "")

	// Execute command
	cmd := exec.Command(resumeCmd.Args[0], resumeCmd.Args[1:]...)
	cmd.Dir = resumeCmd.WorkingDir

	err := cmd.Run()
	
	if err != nil {
		s.logAutoResume(taskId, "failed", err.Error())
	} else {
		s.logAutoResume(taskId, "completed", "")
	}
}

// fetchComments gets all comments for a task
func (s *Service) fetchComments(taskId string) ([]resume.Comment, error) {
	records, err := s.app.FindRecordsByFilter(
		"comments",
		fmt.Sprintf("task = '%s'", taskId),
		"+created",
		100,
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

// logAutoResume logs auto-resume events (could be to a log file or collection)
func (s *Service) logAutoResume(taskId, status, errorMsg string) {
	// Simple logging - could be expanded to a dedicated collection
	fmt.Printf("[auto-resume] task=%s status=%s error=%s\n", taskId, status, errorMsg)
}

// hasAgentMention checks if comment mentions @agent
func hasAgentMention(comment *core.Record) bool {
	metadata := comment.Get("metadata")
	if metadata == nil {
		return false
	}

	metaMap, ok := metadata.(map[string]any)
	if !ok {
		return false
	}

	mentions, ok := metaMap["mentions"].([]any)
	if !ok {
		return false
	}

	for _, m := range mentions {
		if mention, ok := m.(string); ok && mention == "@agent" {
			return true
		}
	}
	return false
}

// ensureHistorySlice converts history to proper type
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
```

---

### Task 6.2: Register Comment Hook

**File**: `internal/hooks/comments.go`

**Description**: Hook that fires when comments are created.

```go
package hooks

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	
	"egenskriven/internal/autoresume"
)

// RegisterCommentHooks registers hooks for the comments collection
func RegisterCommentHooks(app *pocketbase.PocketBase) {
	autoResumeService := autoresume.NewService(app)

	// After comment is created
	app.OnRecordAfterCreateSuccess("comments").Add(func(e *core.RecordEvent) error {
		// Check if this comment should trigger auto-resume
		// Run in goroutine to not block the create response
		go func() {
			if err := autoResumeService.CheckAndResume(e.Record); err != nil {
				// Log error but don't fail the request
				app.Logger().Error("auto-resume check failed",
					"comment", e.Record.Id,
					"error", err,
				)
			}
		}()

		return nil
	})
}
```

**Register in main/app setup**:

```go
// In main.go or app initialization
hooks.RegisterCommentHooks(app)
```

---

### Task 6.3: Add Board Resume Mode Configuration

**File**: `internal/commands/board.go` (update existing)

**Description**: Add ability to configure resume mode on boards.

```go
// Add to existing board update command

func newBoardUpdateCmd(app *pocketbase.PocketBase) *cobra.Command {
	var resumeMode string
	// ... other flags ...

	cmd := &cobra.Command{
		Use:   "update <board-ref>",
		Short: "Update board settings",
		// ...
		RunE: func(cmd *cobra.Command, args []string) error {
			boardRef := args[0]
			
			board, err := resolver.ResolveBoardRef(app, boardRef)
			if err != nil {
				return err
			}

			// Update resume mode if specified
			if resumeMode != "" {
				validModes := []string{"manual", "command", "auto"}
				if !contains(validModes, resumeMode) {
					return fmt.Errorf("invalid resume mode %q: must be one of %v", 
						resumeMode, validModes)
				}
				board.Set("resume_mode", resumeMode)
			}

			// ... other updates ...

			if err := app.Save(board); err != nil {
				return err
			}

			fmt.Printf("Board %s updated\n", board.GetString("name"))
			return nil
		},
	}

	cmd.Flags().StringVar(&resumeMode, "resume-mode", "", 
		"Resume mode (manual, command, auto)")
	// ... other flags ...

	return cmd
}
```

**Also add to board show command**:

```go
// In board show output
fmt.Printf("  Resume Mode: %s\n", board.GetString("resume_mode"))
```

---

### Task 6.4: Add Resume Mode to UI Board Settings

**File**: `ui/src/components/BoardSettings/ResumeModeSelector.tsx`

```typescript
import React from 'react';

interface ResumeModeSelectorProps {
  value: 'manual' | 'command' | 'auto';
  onChange: (mode: 'manual' | 'command' | 'auto') => void;
}

const modes = [
  {
    value: 'manual' as const,
    label: 'Manual',
    description: 'Print resume command for user to copy and run',
  },
  {
    value: 'command' as const,
    label: 'Command',
    description: 'User runs "egenskriven resume --exec" to resume',
  },
  {
    value: 'auto' as const,
    label: 'Auto',
    description: 'Automatically resume when human adds @agent comment',
  },
];

export function ResumeModeSelector({ value, onChange }: ResumeModeSelectorProps) {
  return (
    <div className="space-y-3">
      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
        Resume Mode
      </label>
      <p className="text-xs text-gray-500 dark:text-gray-400">
        How blocked tasks should be resumed after human input
      </p>
      
      <div className="space-y-2">
        {modes.map((mode) => (
          <label
            key={mode.value}
            className={`flex items-start gap-3 p-3 rounded-lg border cursor-pointer
                       transition-colors
                       ${value === mode.value
                         ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                         : 'border-gray-200 dark:border-gray-700 hover:border-gray-300'
                       }`}
          >
            <input
              type="radio"
              name="resumeMode"
              value={mode.value}
              checked={value === mode.value}
              onChange={() => onChange(mode.value)}
              className="mt-1"
            />
            <div>
              <div className="font-medium text-gray-900 dark:text-gray-100">
                {mode.label}
              </div>
              <div className="text-sm text-gray-500 dark:text-gray-400">
                {mode.description}
              </div>
            </div>
          </label>
        ))}
      </div>

      {value === 'auto' && (
        <div className="mt-3 p-3 bg-orange-50 dark:bg-orange-900/20 rounded-lg">
          <p className="text-sm text-orange-700 dark:text-orange-400">
            <strong>Note:</strong> Auto-resume will trigger when a human adds a 
            comment containing <code className="bg-orange-100 dark:bg-orange-800 
            px-1 rounded">@agent</code> to a blocked task with a linked session.
          </p>
        </div>
      )}
    </div>
  );
}
```

---

### Task 6.5: Update Comments Panel with Auto-Resume Indicator

**File**: `ui/src/components/TaskDetail/CommentsPanel.tsx` (update)

Add indicator when auto-resume is enabled:

```typescript
// Add to CommentsPanel props
interface CommentsPanelProps {
  taskId: string;
  boardResumeMode?: string;
  taskColumn?: string;
}

// In the component
export function CommentsPanel({ taskId, boardResumeMode, taskColumn }: CommentsPanelProps) {
  const isAutoResumeEnabled = boardResumeMode === 'auto' && taskColumn === 'need_input';
  
  // ... existing code ...

  return (
    <div className="border-t border-gray-200 dark:border-gray-700">
      {/* Auto-resume indicator */}
      {isAutoResumeEnabled && (
        <div className="px-4 py-2 bg-green-50 dark:bg-green-900/20 border-b 
                        border-green-100 dark:border-green-800">
          <p className="text-xs text-green-700 dark:text-green-400 flex items-center gap-2">
            <span className="relative flex h-2 w-2">
              <span className="animate-ping absolute inline-flex h-full w-full 
                               rounded-full bg-green-400 opacity-75"></span>
              <span className="relative inline-flex rounded-full h-2 w-2 
                               bg-green-500"></span>
            </span>
            Auto-resume enabled. Add a comment with @agent to trigger.
          </p>
        </div>
      )}

      {/* Rest of component */}
      {/* ... */}
    </div>
  );
}
```

---

### Task 6.6: Write Unit Tests

**File**: `internal/autoresume/service_test.go`

```go
package autoresume

import (
	"testing"
	"time"
)

func TestCheckAndResume_AllConditionsMet(t *testing.T) {
	app := setupTestApp(t)
	defer cleanupTestApp(t, app)

	// Setup: Create board with auto mode
	board := createTestBoard(t, app, "test-board")
	board.Set("resume_mode", "auto")
	app.Save(board)

	// Create task in need_input with session
	task := createTestTask(t, app, "Test task")
	task.Set("board", board.Id)
	task.Set("column", "need_input")
	task.Set("agent_session", map[string]any{
		"tool":        "opencode",
		"ref":         "test-session-123",
		"ref_type":    "uuid",
		"working_dir": "/tmp",
		"linked_at":   time.Now().Format(time.RFC3339),
	})
	app.Save(task)

	// Create comment with @agent mention from human
	comment := createTestComment(t, app, task.Id, "@agent please continue", "human")

	// Check
	service := NewService(app)
	err := service.CheckAndResume(comment)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify task moved to in_progress
	refreshedTask, _ := app.FindRecordById("tasks", task.Id)
	if refreshedTask.GetString("column") != "in_progress" {
		t.Errorf("task should be in_progress, got %s", refreshedTask.GetString("column"))
	}
}

func TestCheckAndResume_AgentComment(t *testing.T) {
	app := setupTestApp(t)
	defer cleanupTestApp(t, app)

	board := createTestBoard(t, app, "test-board")
	board.Set("resume_mode", "auto")
	app.Save(board)

	task := createTestTask(t, app, "Test task")
	task.Set("board", board.Id)
	task.Set("column", "need_input")
	task.Set("agent_session", map[string]any{
		"tool": "opencode",
		"ref":  "test-123",
	})
	app.Save(task)

	// Comment from agent - should NOT trigger
	comment := createTestComment(t, app, task.Id, "@agent status", "agent")

	service := NewService(app)
	service.CheckAndResume(comment)

	// Task should still be in need_input
	refreshedTask, _ := app.FindRecordById("tasks", task.Id)
	if refreshedTask.GetString("column") != "need_input" {
		t.Error("task should not move for agent comment")
	}
}

func TestCheckAndResume_NoMention(t *testing.T) {
	app := setupTestApp(t)
	defer cleanupTestApp(t, app)

	board := createTestBoard(t, app, "test-board")
	board.Set("resume_mode", "auto")
	app.Save(board)

	task := createTestTask(t, app, "Test task")
	task.Set("board", board.Id)
	task.Set("column", "need_input")
	task.Set("agent_session", map[string]any{
		"tool": "opencode",
		"ref":  "test-123",
	})
	app.Save(task)

	// Comment without @agent - should NOT trigger
	comment := createTestComment(t, app, task.Id, "Just a regular comment", "human")

	service := NewService(app)
	service.CheckAndResume(comment)

	// Task should still be in need_input
	refreshedTask, _ := app.FindRecordById("tasks", task.Id)
	if refreshedTask.GetString("column") != "need_input" {
		t.Error("task should not move without @agent mention")
	}
}

func TestCheckAndResume_ManualMode(t *testing.T) {
	app := setupTestApp(t)
	defer cleanupTestApp(t, app)

	// Board with manual mode - should not auto-resume
	board := createTestBoard(t, app, "test-board")
	board.Set("resume_mode", "manual")
	app.Save(board)

	task := createTestTask(t, app, "Test task")
	task.Set("board", board.Id)
	task.Set("column", "need_input")
	task.Set("agent_session", map[string]any{
		"tool": "opencode",
		"ref":  "test-123",
	})
	app.Save(task)

	comment := createTestComment(t, app, task.Id, "@agent continue", "human")

	service := NewService(app)
	service.CheckAndResume(comment)

	// Task should still be in need_input
	refreshedTask, _ := app.FindRecordById("tasks", task.Id)
	if refreshedTask.GetString("column") != "need_input" {
		t.Error("task should not move when resume_mode is manual")
	}
}

func TestCheckAndResume_NoSession(t *testing.T) {
	app := setupTestApp(t)
	defer cleanupTestApp(t, app)

	board := createTestBoard(t, app, "test-board")
	board.Set("resume_mode", "auto")
	app.Save(board)

	// Task without session
	task := createTestTask(t, app, "Test task")
	task.Set("board", board.Id)
	task.Set("column", "need_input")
	// No agent_session set
	app.Save(task)

	comment := createTestComment(t, app, task.Id, "@agent continue", "human")

	service := NewService(app)
	service.CheckAndResume(comment)

	// Task should still be in need_input
	refreshedTask, _ := app.FindRecordById("tasks", task.Id)
	if refreshedTask.GetString("column") != "need_input" {
		t.Error("task should not move without session")
	}
}

func TestCheckAndResume_WrongColumn(t *testing.T) {
	app := setupTestApp(t)
	defer cleanupTestApp(t, app)

	board := createTestBoard(t, app, "test-board")
	board.Set("resume_mode", "auto")
	app.Save(board)

	// Task not in need_input
	task := createTestTask(t, app, "Test task")
	task.Set("board", board.Id)
	task.Set("column", "in_progress") // Not need_input
	task.Set("agent_session", map[string]any{
		"tool": "opencode",
		"ref":  "test-123",
	})
	app.Save(task)

	comment := createTestComment(t, app, task.Id, "@agent continue", "human")

	service := NewService(app)
	service.CheckAndResume(comment)

	// Task should remain in_progress
	refreshedTask, _ := app.FindRecordById("tasks", task.Id)
	if refreshedTask.GetString("column") != "in_progress" {
		t.Error("task column should not change if not in need_input")
	}
}

func TestHasAgentMention(t *testing.T) {
	tests := []struct {
		metadata map[string]any
		expected bool
	}{
		{nil, false},
		{map[string]any{}, false},
		{map[string]any{"mentions": []any{}}, false},
		{map[string]any{"mentions": []any{"@user"}}, false},
		{map[string]any{"mentions": []any{"@agent"}}, true},
		{map[string]any{"mentions": []any{"@user", "@agent"}}, true},
	}

	for i, tt := range tests {
		comment := &mockRecord{data: map[string]any{"metadata": tt.metadata}}
		got := hasAgentMention(comment)
		if got != tt.expected {
			t.Errorf("test %d: hasAgentMention() = %v, want %v", i, got, tt.expected)
		}
	}
}

// Helper to create test comment with mentions
func createTestComment(t *testing.T, app *pocketbase.PocketBase, 
	taskId, content, authorType string) *core.Record {
	
	collection, _ := app.FindCollectionByNameOrId("comments")
	record := core.NewRecord(collection)
	record.Set("task", taskId)
	record.Set("content", content)
	record.Set("author_type", authorType)
	
	// Extract mentions
	mentions := extractMentionsFromContent(content)
	record.Set("metadata", map[string]any{"mentions": mentions})
	
	if err := app.Save(record); err != nil {
		t.Fatalf("failed to create test comment: %v", err)
	}
	return record
}

func extractMentionsFromContent(content string) []string {
	re := regexp.MustCompile(`@\w+`)
	matches := re.FindAllString(content, -1)
	return matches
}
```

---

### Task 6.7: Add E2E Test

**File**: `tests/e2e/autoresume_test.go`

```go
package e2e

import (
	"testing"
	"time"
)

func TestAutoResumeE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test")
	}

	app := setupE2EApp(t)
	defer cleanupE2EApp(t, app)

	// 1. Create board with auto mode
	runCommand(t, "egenskriven", "board", "create", "test-board")
	runCommand(t, "egenskriven", "board", "update", "test-board", "--resume-mode", "auto")

	// 2. Create task
	runCommand(t, "egenskriven", "add", "Test auto-resume", "--board", "test-board")

	// 3. Link session
	runCommand(t, "egenskriven", "session", "link", "WRK-1", 
		"--tool", "opencode", "--ref", "test-session-123")

	// 4. Block task
	runCommand(t, "egenskriven", "block", "WRK-1", "What should I do?")

	// 5. Add comment with @agent (via API or CLI)
	runCommand(t, "egenskriven", "comment", "WRK-1", "@agent Use JWT auth")

	// 6. Wait for auto-resume (happens async)
	time.Sleep(2 * time.Second)

	// 7. Verify task is now in_progress
	output := runCommand(t, "egenskriven", "show", "WRK-1", "--json")
	
	var task map[string]any
	json.Unmarshal([]byte(output), &task)
	
	if task["column"] != "in_progress" {
		t.Errorf("task should be in_progress after auto-resume, got %s", task["column"])
	}

	// 8. Verify history has auto_resumed entry
	history := task["history"].([]any)
	lastEntry := history[len(history)-1].(map[string]any)
	if lastEntry["action"] != "auto_resumed" {
		t.Errorf("last history entry should be auto_resumed, got %s", lastEntry["action"])
	}
}
```

---

## Files Changed/Created

| File | Change Type | Description |
|------|-------------|-------------|
| `internal/autoresume/service.go` | New | Auto-resume service |
| `internal/autoresume/service_test.go` | New | Service tests |
| `internal/hooks/comments.go` | New | Comment creation hook |
| `internal/commands/board.go` | Modified | Add resume-mode flag |
| `ui/src/components/BoardSettings/ResumeModeSelector.tsx` | New | Resume mode UI |
| `ui/src/components/TaskDetail/CommentsPanel.tsx` | Modified | Auto-resume indicator |
| `tests/e2e/autoresume_test.go` | New | E2E tests |

---

## Notes for Implementer

1. **Async execution**: The resume command runs in a goroutine to not block comment creation.

2. **Logging**: Auto-resume events should be logged for debugging.

3. **Race conditions**: Multiple rapid @agent comments should be handled gracefully (only first triggers).

4. **Error recovery**: If auto-resume fails, the task remains in need_input and user can manually resume.

5. **Security**: Only human comments can trigger auto-resume to prevent agents from triggering themselves.
