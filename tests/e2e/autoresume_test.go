// Package e2e contains end-to-end tests for EgenSkriven.
// These tests verify complete workflows through the application.
package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"github.com/ramtinJ95/EgenSkriven/internal/autoresume"
)

// TestAutoResumeE2E_FullWorkflow tests the complete auto-resume workflow.
// This is an integration test that verifies:
// 1. Task moves to in_progress after auto-resume trigger
// 2. History contains auto_resumed action entry
// 3. All conditions must be met for trigger
func TestAutoResumeE2E_FullWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	app := setupE2ETestApp(t)

	// 1. Create board with auto resume mode
	board := createE2EBoard(t, app, "TEST", "auto")

	// 2. Create task in need_input with linked session
	task := createE2ETask(t, app, board.Id, "Test auto-resume task", "need_input")
	task.Set("agent_session", map[string]any{
		"tool":        "claude-code",
		"ref":         "test-session-123",
		"ref_type":    "uuid",
		"working_dir": "/tmp",
		"linked_at":   time.Now().UTC().Format(time.RFC3339),
	})
	if err := app.Save(task); err != nil {
		t.Fatalf("failed to update task with session: %v", err)
	}

	// 3. Add comment with @agent mention from human
	comment := createE2EComment(t, app, task.Id, "@agent Please continue with the implementation", "human")

	// 4. Run auto-resume check (simulating what hook would do)
	service := autoresume.NewService(app)
	err := service.CheckAndResume(comment)
	if err != nil {
		t.Fatalf("auto-resume failed: %v", err)
	}

	// 5. Verify task moved to in_progress
	refreshedTask, err := app.FindRecordById("tasks", task.Id)
	if err != nil {
		t.Fatalf("failed to find task: %v", err)
	}

	if refreshedTask.GetString("column") != "in_progress" {
		t.Errorf("task should be in_progress after auto-resume, got %s", refreshedTask.GetString("column"))
	}

	// 6. Verify history contains auto_resumed entry
	history := refreshedTask.Get("history")
	historySlice, err := getHistorySlice(history)
	if err != nil {
		t.Fatalf("failed to parse history: %v", err)
	}

	if len(historySlice) == 0 {
		t.Fatal("history should contain at least one entry")
	}

	lastEntry := historySlice[len(historySlice)-1]

	if lastEntry["action"] != "auto_resumed" {
		t.Errorf("last history entry should have action 'auto_resumed', got %v", lastEntry["action"])
	}

	if lastEntry["actor"] != "system" {
		t.Errorf("last history entry should have actor 'system', got %v", lastEntry["actor"])
	}

	// Verify changes in history
	changes, ok := lastEntry["changes"].(map[string]any)
	if !ok {
		t.Fatal("history entry should have changes map")
	}

	columnChange, ok := changes["column"].(map[string]any)
	if !ok {
		t.Fatal("changes should have column change")
	}

	if columnChange["from"] != "need_input" {
		t.Errorf("column change 'from' should be 'need_input', got %v", columnChange["from"])
	}

	if columnChange["to"] != "in_progress" {
		t.Errorf("column change 'to' should be 'in_progress', got %v", columnChange["to"])
	}

	// Verify trigger comment is recorded in metadata
	metadata, ok := lastEntry["metadata"].(map[string]any)
	if !ok {
		t.Fatal("history entry should have metadata")
	}

	if metadata["trigger_comment"] != comment.Id {
		t.Errorf("metadata should reference trigger comment, got %v", metadata["trigger_comment"])
	}
}

// TestAutoResumeE2E_TaskInProgressAfterTrigger verifies task state after trigger.
func TestAutoResumeE2E_TaskInProgressAfterTrigger(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	app := setupE2ETestApp(t)

	board := createE2EBoard(t, app, "TEST", "auto")
	task := createE2ETask(t, app, board.Id, "Test task", "need_input")
	task.Set("agent_session", map[string]any{
		"tool":        "claude-code",
		"ref":         "session-456",
		"working_dir": "/tmp",
	})
	app.Save(task)

	comment := createE2EComment(t, app, task.Id, "@agent continue", "human")

	service := autoresume.NewService(app)
	service.CheckAndResume(comment)

	// Verify task is now in_progress
	refreshedTask, _ := app.FindRecordById("tasks", task.Id)
	if refreshedTask.GetString("column") != "in_progress" {
		t.Errorf("expected task column to be 'in_progress', got '%s'", refreshedTask.GetString("column"))
	}
}

// TestAutoResumeE2E_HistoryContainsAutoResumedAction verifies history entry.
func TestAutoResumeE2E_HistoryContainsAutoResumedAction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	app := setupE2ETestApp(t)

	board := createE2EBoard(t, app, "TEST", "auto")
	task := createE2ETask(t, app, board.Id, "Test task", "need_input")
	task.Set("agent_session", map[string]any{
		"tool":        "claude-code",
		"ref":         "session-789",
		"working_dir": "/tmp",
	})
	app.Save(task)

	comment := createE2EComment(t, app, task.Id, "@agent proceed", "human")

	service := autoresume.NewService(app)
	service.CheckAndResume(comment)

	// Verify history has auto_resumed action
	refreshedTask, _ := app.FindRecordById("tasks", task.Id)
	history := refreshedTask.Get("history")

	historySlice, err := getHistorySlice(history)
	if err != nil || len(historySlice) == 0 {
		t.Fatalf("expected non-empty history, err: %v", err)
	}

	lastEntry := historySlice[len(historySlice)-1]
	if lastEntry["action"] != "auto_resumed" {
		t.Errorf("expected 'auto_resumed' action, got '%v'", lastEntry["action"])
	}
}

// TestAutoResumeE2E_AllConditionsRequired verifies all conditions must be met.
func TestAutoResumeE2E_AllConditionsRequired(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	app := setupE2ETestApp(t)

	tests := []struct {
		name        string
		resumeMode  string
		column      string
		withSession bool
		authorType  string
		mention     string
		shouldMove  bool
	}{
		{"all conditions met", "auto", "need_input", true, "human", "@agent continue", true},
		{"manual mode", "manual", "need_input", true, "human", "@agent continue", false},
		{"command mode", "command", "need_input", true, "human", "@agent continue", false},
		{"wrong column", "auto", "in_progress", true, "human", "@agent continue", false},
		{"no session", "auto", "need_input", false, "human", "@agent continue", false},
		{"agent comment", "auto", "need_input", true, "agent", "@agent continue", false},
		{"no mention", "auto", "need_input", true, "human", "just a comment", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board := createE2EBoard(t, app, "T"+tt.name[:2], tt.resumeMode)
			task := createE2ETask(t, app, board.Id, "Task for "+tt.name, tt.column)

			if tt.withSession {
				task.Set("agent_session", map[string]any{
					"tool":        "claude-code",
					"ref":         "session-" + tt.name,
					"working_dir": "/tmp",
				})
				app.Save(task)
			}

			originalColumn := task.GetString("column")
			comment := createE2EComment(t, app, task.Id, tt.mention, tt.authorType)

			service := autoresume.NewService(app)
			service.CheckAndResume(comment)

			refreshedTask, _ := app.FindRecordById("tasks", task.Id)
			finalColumn := refreshedTask.GetString("column")

			// Check if auto-resume triggered by comparing original vs final column
			// Auto-resume triggers = task moved from need_input to in_progress
			autoResumeTriggered := originalColumn == "need_input" && finalColumn == "in_progress"

			if autoResumeTriggered != tt.shouldMove {
				t.Errorf("expected shouldMove=%v, but task went from '%s' to '%s'",
					tt.shouldMove, originalColumn, finalColumn)
			}
		})
	}
}

// --- E2E Test Helpers ---

// setupE2ETestApp creates an isolated PocketBase instance for E2E tests.
func setupE2ETestApp(t *testing.T) *pocketbase.PocketBase {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "egenskriven-e2e-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: tmpDir,
	})

	if err := app.Bootstrap(); err != nil {
		t.Fatalf("failed to bootstrap app: %v", err)
	}

	// Create required collections
	setupE2ECollections(t, app)

	return app
}

// setupE2ECollections creates the collections needed for E2E tests.
func setupE2ECollections(t *testing.T, app *pocketbase.PocketBase) {
	t.Helper()

	// Boards collection
	if _, err := app.FindCollectionByNameOrId("boards"); err != nil {
		boards := core.NewBaseCollection("boards")
		boards.Fields.Add(&core.TextField{Name: "name", Required: true})
		boards.Fields.Add(&core.TextField{Name: "prefix", Required: true})
		boards.Fields.Add(&core.TextField{Name: "resume_mode"})
		boards.Fields.Add(&core.JSONField{Name: "columns"})
		boards.Fields.Add(&core.NumberField{Name: "next_seq"})
		if err := app.Save(boards); err != nil {
			t.Fatalf("failed to create boards collection: %v", err)
		}
	}

	// Tasks collection
	if _, err := app.FindCollectionByNameOrId("tasks"); err != nil {
		tasks := core.NewBaseCollection("tasks")
		tasks.Fields.Add(&core.TextField{Name: "title", Required: true})
		tasks.Fields.Add(&core.TextField{Name: "column"})
		tasks.Fields.Add(&core.TextField{Name: "board"})
		tasks.Fields.Add(&core.JSONField{Name: "agent_session"})
		tasks.Fields.Add(&core.JSONField{Name: "history"})
		tasks.Fields.Add(&core.NumberField{Name: "seq"})
		if err := app.Save(tasks); err != nil {
			t.Fatalf("failed to create tasks collection: %v", err)
		}
	}

	// Comments collection
	if _, err := app.FindCollectionByNameOrId("comments"); err != nil {
		comments := core.NewBaseCollection("comments")
		comments.Fields.Add(&core.TextField{Name: "task", Required: true})
		comments.Fields.Add(&core.TextField{Name: "content", Required: true})
		comments.Fields.Add(&core.TextField{Name: "author_type", Required: true})
		comments.Fields.Add(&core.TextField{Name: "author_id"})
		comments.Fields.Add(&core.JSONField{Name: "metadata"})
		comments.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
		if err := app.Save(comments); err != nil {
			t.Fatalf("failed to create comments collection: %v", err)
		}
	}
}

// createE2EBoard creates a board for E2E testing.
func createE2EBoard(t *testing.T, app *pocketbase.PocketBase, prefix, resumeMode string) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("boards")
	if err != nil {
		t.Fatalf("boards collection not found: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("name", "E2E Test Board "+prefix)
	record.Set("prefix", prefix)
	record.Set("resume_mode", resumeMode)
	record.Set("columns", []string{"backlog", "todo", "in_progress", "need_input", "done"})
	record.Set("next_seq", 1)

	if err := app.Save(record); err != nil {
		t.Fatalf("failed to create board: %v", err)
	}

	return record
}

// createE2ETask creates a task for E2E testing.
func createE2ETask(t *testing.T, app *pocketbase.PocketBase, boardId, title, column string) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("tasks")
	if err != nil {
		t.Fatalf("tasks collection not found: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("title", title)
	record.Set("board", boardId)
	record.Set("column", column)
	record.Set("seq", 1)
	record.Set("history", []any{})

	if err := app.Save(record); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	return record
}

// createE2EComment creates a comment for E2E testing.
func createE2EComment(t *testing.T, app *pocketbase.PocketBase, taskId, content, authorType string) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("comments")
	if err != nil {
		t.Fatalf("comments collection not found: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("task", taskId)
	record.Set("content", content)
	record.Set("author_type", authorType)

	// Extract mentions from content
	mentions := extractE2EMentions(content)
	record.Set("metadata", map[string]any{"mentions": mentions})

	if err := app.Save(record); err != nil {
		t.Fatalf("failed to create comment: %v", err)
	}

	return record
}

// getHistorySlice parses history field handling types.JSONRaw.
func getHistorySlice(history any) ([]map[string]any, error) {
	if history == nil {
		return nil, fmt.Errorf("history is nil")
	}

	// Handle []any (in-memory)
	if slice, ok := history.([]any); ok {
		result := make([]map[string]any, 0, len(slice))
		for _, item := range slice {
			if m, ok := item.(map[string]any); ok {
				result = append(result, m)
			}
		}
		return result, nil
	}

	// Handle types.JSONRaw (from DB) - parse as JSON
	raw := []byte(fmt.Sprintf("%s", history))
	var result []map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal history: %w", err)
	}
	return result, nil
}

// extractE2EMentions extracts @mentions from content.
func extractE2EMentions(content string) []string {
	var mentions []string
	for i := 0; i < len(content); i++ {
		if content[i] == '@' {
			// Find end of mention
			j := i + 1
			for j < len(content) && (content[j] >= 'a' && content[j] <= 'z' ||
				content[j] >= 'A' && content[j] <= 'Z' ||
				content[j] >= '0' && content[j] <= '9' ||
				content[j] == '_') {
				j++
			}
			if j > i+1 {
				mentions = append(mentions, content[i:j])
			}
		}
	}
	return mentions
}
