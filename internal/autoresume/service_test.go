package autoresume

import (
	"regexp"
	"testing"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"github.com/ramtinJ95/EgenSkriven/internal/testutil"
)

// TestCheckAndResume_AgentCommentDoesNotTrigger verifies that comments
// from agents (author_type = "agent") do not trigger auto-resume.
func TestCheckAndResume_AgentCommentDoesNotTrigger(t *testing.T) {
	app := setupTestAppWithCollections(t)

	board := createTestBoard(t, app, "TEST", "auto")
	task := createTestTask(t, app, board.Id, "need_input", true)

	// Comment from agent - should NOT trigger
	comment := createTestComment(t, app, task.Id, "@agent status", "agent")

	service := NewService(app)
	err := service.CheckAndResume(comment)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Task should still be in need_input
	refreshedTask, _ := app.FindRecordById("tasks", task.Id)
	if refreshedTask.GetString("column") != "need_input" {
		t.Errorf("task should not move for agent comment, got column=%s", refreshedTask.GetString("column"))
	}
}

// TestCheckAndResume_NoMentionDoesNotTrigger verifies that comments
// without @agent mention do not trigger auto-resume.
func TestCheckAndResume_NoMentionDoesNotTrigger(t *testing.T) {
	app := setupTestAppWithCollections(t)

	board := createTestBoard(t, app, "TEST", "auto")
	task := createTestTask(t, app, board.Id, "need_input", true)

	// Comment without @agent - should NOT trigger
	comment := createTestComment(t, app, task.Id, "Just a regular comment", "human")

	service := NewService(app)
	err := service.CheckAndResume(comment)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Task should still be in need_input
	refreshedTask, _ := app.FindRecordById("tasks", task.Id)
	if refreshedTask.GetString("column") != "need_input" {
		t.Errorf("task should not move without @agent mention, got column=%s", refreshedTask.GetString("column"))
	}
}

// TestCheckAndResume_ManualModeDoesNotTrigger verifies that comments
// on boards with resume_mode="manual" do not trigger auto-resume.
func TestCheckAndResume_ManualModeDoesNotTrigger(t *testing.T) {
	app := setupTestAppWithCollections(t)

	board := createTestBoard(t, app, "TEST", "manual") // Manual mode
	task := createTestTask(t, app, board.Id, "need_input", true)

	comment := createTestComment(t, app, task.Id, "@agent continue", "human")

	service := NewService(app)
	err := service.CheckAndResume(comment)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Task should still be in need_input
	refreshedTask, _ := app.FindRecordById("tasks", task.Id)
	if refreshedTask.GetString("column") != "need_input" {
		t.Errorf("task should not move when resume_mode is manual, got column=%s", refreshedTask.GetString("column"))
	}
}

// TestCheckAndResume_CommandModeDoesNotTrigger verifies that comments
// on boards with resume_mode="command" do not trigger auto-resume.
func TestCheckAndResume_CommandModeDoesNotTrigger(t *testing.T) {
	app := setupTestAppWithCollections(t)

	board := createTestBoard(t, app, "TEST", "command") // Command mode
	task := createTestTask(t, app, board.Id, "need_input", true)

	comment := createTestComment(t, app, task.Id, "@agent continue", "human")

	service := NewService(app)
	err := service.CheckAndResume(comment)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Task should still be in need_input
	refreshedTask, _ := app.FindRecordById("tasks", task.Id)
	if refreshedTask.GetString("column") != "need_input" {
		t.Errorf("task should not move when resume_mode is command, got column=%s", refreshedTask.GetString("column"))
	}
}

// TestCheckAndResume_NoSessionDoesNotTrigger verifies that comments
// on tasks without agent_session do not trigger auto-resume.
func TestCheckAndResume_NoSessionDoesNotTrigger(t *testing.T) {
	app := setupTestAppWithCollections(t)

	board := createTestBoard(t, app, "TEST", "auto")
	task := createTestTask(t, app, board.Id, "need_input", false) // No session

	comment := createTestComment(t, app, task.Id, "@agent continue", "human")

	service := NewService(app)
	err := service.CheckAndResume(comment)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Task should still be in need_input
	refreshedTask, _ := app.FindRecordById("tasks", task.Id)
	if refreshedTask.GetString("column") != "need_input" {
		t.Errorf("task should not move without session, got column=%s", refreshedTask.GetString("column"))
	}
}

// TestCheckAndResume_WrongColumnDoesNotTrigger verifies that comments
// on tasks not in need_input column do not trigger auto-resume.
func TestCheckAndResume_WrongColumnDoesNotTrigger(t *testing.T) {
	app := setupTestAppWithCollections(t)

	board := createTestBoard(t, app, "TEST", "auto")
	task := createTestTask(t, app, board.Id, "in_progress", true) // Not need_input

	comment := createTestComment(t, app, task.Id, "@agent continue", "human")

	service := NewService(app)
	err := service.CheckAndResume(comment)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Task should remain in_progress
	refreshedTask, _ := app.FindRecordById("tasks", task.Id)
	if refreshedTask.GetString("column") != "in_progress" {
		t.Errorf("task column should not change if not in need_input, got column=%s", refreshedTask.GetString("column"))
	}
}

// TestHasAgentMention tests the hasAgentMention helper function with in-memory records.
// Note: This tests the function's behavior with various metadata formats.
func TestHasAgentMention(t *testing.T) {
	app := setupTestAppWithCollections(t)
	collection, _ := app.FindCollectionByNameOrId("comments")
	taskCollection, _ := app.FindCollectionByNameOrId("tasks")

	// Create a task for the comments
	task := core.NewRecord(taskCollection)
	task.Set("title", "Test")
	task.Set("column", "todo")
	app.Save(task)

	tests := []struct {
		name     string
		metadata map[string]any
		expected bool
	}{
		{"nil metadata", nil, false},
		{"empty metadata", map[string]any{}, false},
		{"empty mentions", map[string]any{"mentions": []any{}}, false},
		{"other mention", map[string]any{"mentions": []any{"@user"}}, false},
		{"agent mention", map[string]any{"mentions": []any{"@agent"}}, true},
		{"agent among others", map[string]any{"mentions": []any{"@user", "@agent"}}, true},
		{"multiple agents", map[string]any{"mentions": []any{"@agent", "@agent"}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create in-memory record without saving to avoid JSON serialization issues
			record := core.NewRecord(collection)
			record.Set("task", task.Id)
			record.Set("content", "test")
			record.Set("author_type", "human")
			record.Set("metadata", tt.metadata)

			got := hasAgentMention(record)
			if got != tt.expected {
				t.Errorf("hasAgentMention() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestEnsureHistorySlice tests the ensureHistorySlice helper function.
func TestEnsureHistorySlice(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected int // length of result
	}{
		{"nil input", nil, 0},
		{"empty slice", []any{}, 0},
		{"valid slice", []any{map[string]any{"action": "test"}}, 1},
		{"invalid type", "string", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ensureHistorySlice(tt.input)
			if len(result) != tt.expected {
				t.Errorf("ensureHistorySlice() returned %d items, want %d", len(result), tt.expected)
			}
		})
	}
}

// --- Test Helpers ---

// setupTestAppWithCollections creates a test app and ensures all required
// collections exist for auto-resume testing.
func setupTestAppWithCollections(t *testing.T) *pocketbase.PocketBase {
	t.Helper()

	// Use testutil to create isolated test app
	app := testutil.NewTestApp(t)

	// Create boards collection if not exists
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

	// Create tasks collection if not exists
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

	// Create comments collection if not exists
	if _, err := app.FindCollectionByNameOrId("comments"); err != nil {
		comments := core.NewBaseCollection("comments")
		comments.Fields.Add(&core.TextField{Name: "task", Required: true})
		comments.Fields.Add(&core.TextField{Name: "content", Required: true})
		comments.Fields.Add(&core.TextField{Name: "author_type", Required: true})
		comments.Fields.Add(&core.TextField{Name: "author_id"})
		comments.Fields.Add(&core.JSONField{Name: "metadata"})
		if err := app.Save(comments); err != nil {
			t.Fatalf("failed to create comments collection: %v", err)
		}
	}

	return app
}

// createTestBoard creates a board with specified resume mode.
func createTestBoard(t *testing.T, app *pocketbase.PocketBase, prefix, resumeMode string) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("boards")
	if err != nil {
		t.Fatalf("boards collection not found: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("name", "Test Board")
	record.Set("prefix", prefix)
	record.Set("resume_mode", resumeMode)
	record.Set("columns", []string{"backlog", "todo", "in_progress", "need_input", "done"})
	record.Set("next_seq", 1)

	if err := app.Save(record); err != nil {
		t.Fatalf("failed to create test board: %v", err)
	}

	return record
}

// createTestTask creates a task with specified column and optional session.
func createTestTask(t *testing.T, app *pocketbase.PocketBase, boardId, column string, withSession bool) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("tasks")
	if err != nil {
		t.Fatalf("tasks collection not found: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("title", "Test Task")
	record.Set("board", boardId)
	record.Set("column", column)
	record.Set("seq", 1)

	if withSession {
		record.Set("agent_session", map[string]any{
			"tool":        "claude",
			"ref":         "test-session-123",
			"ref_type":    "uuid",
			"working_dir": "/tmp",
		})
	}

	if err := app.Save(record); err != nil {
		t.Fatalf("failed to create test task: %v", err)
	}

	return record
}

// createTestComment creates a comment with auto-extracted mentions.
func createTestComment(t *testing.T, app *pocketbase.PocketBase, taskId, content, authorType string) *core.Record {
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
	mentions := extractMentionsFromContent(content)
	record.Set("metadata", map[string]any{"mentions": mentions})

	if err := app.Save(record); err != nil {
		t.Fatalf("failed to create test comment: %v", err)
	}

	return record
}

// createCommentWithMetadata creates a comment with explicit metadata for testing.
func createCommentWithMetadata(t *testing.T, app *pocketbase.PocketBase, metadata map[string]any) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("comments")
	if err != nil {
		t.Fatalf("comments collection not found: %v", err)
	}

	// Create a minimal task first (comment needs a task reference)
	taskCollection, _ := app.FindCollectionByNameOrId("tasks")
	task := core.NewRecord(taskCollection)
	task.Set("title", "Test")
	task.Set("column", "todo")
	app.Save(task)

	record := core.NewRecord(collection)
	record.Set("task", task.Id)
	record.Set("content", "test")
	record.Set("author_type", "human")
	record.Set("metadata", metadata)

	if err := app.Save(record); err != nil {
		t.Fatalf("failed to create test comment: %v", err)
	}

	return record
}

// extractMentionsFromContent extracts @mentions from content string.
func extractMentionsFromContent(content string) []string {
	re := regexp.MustCompile(`@\w+`)
	matches := re.FindAllString(content, -1)
	if matches == nil {
		return []string{}
	}
	return matches
}
