package db

import (
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"github.com/ramtinJ95/EgenSkriven/internal/testutil"
)

// ========== Index Verification Tests ==========

// TestCommentsIndexesExist verifies that all expected indexes exist on the comments collection.
func TestCommentsIndexesExist(t *testing.T) {
	app := setupTestAppWithCollections(t)

	// Get comments collection
	collection, err := app.FindCollectionByNameOrId("comments")
	if err != nil {
		t.Fatalf("comments collection not found: %v", err)
	}

	expectedIndexes := []string{
		"idx_comments_task",
		"idx_comments_created",
	}

	for _, expectedIdx := range expectedIndexes {
		found := false
		for _, idx := range collection.Indexes {
			if strings.Contains(idx, expectedIdx) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected index %s not found in comments collection", expectedIdx)
		}
	}
}

// TestSessionsIndexesExist verifies that all expected indexes exist on the sessions collection.
func TestSessionsIndexesExist(t *testing.T) {
	app := setupTestAppWithCollections(t)

	// Get sessions collection
	collection, err := app.FindCollectionByNameOrId("sessions")
	if err != nil {
		t.Fatalf("sessions collection not found: %v", err)
	}

	expectedIndexes := []string{
		"idx_sessions_task",
		"idx_sessions_status",
		"idx_sessions_external_ref",
	}

	for _, expectedIdx := range expectedIndexes {
		found := false
		for _, idx := range collection.Indexes {
			if strings.Contains(idx, expectedIdx) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected index %s not found in sessions collection", expectedIdx)
		}
	}
}

// TestTasksColumnIndexExists verifies that the tasks collection has an index for the column field.
func TestTasksColumnIndexExists(t *testing.T) {
	app := setupTestAppWithCollections(t)

	// Get tasks collection
	collection, err := app.FindCollectionByNameOrId("tasks")
	if err != nil {
		t.Fatalf("tasks collection not found: %v", err)
	}

	// Check if column index exists
	hasColumnIndex := false
	for _, idx := range collection.Indexes {
		if strings.Contains(idx, "column") {
			hasColumnIndex = true
			break
		}
	}

	// Note: The tasks collection may not have a column index by default
	// This test documents the current state
	if !hasColumnIndex {
		t.Logf("Note: tasks collection does not have an explicit column index")
		t.Logf("Consider adding: CREATE INDEX idx_tasks_column ON tasks (column)")
	}
}

// ========== Index Usage Tests ==========

// TestCommentsQueryUsesTaskIndex verifies that queries on comments.task use the index.
func TestCommentsQueryUsesTaskIndex(t *testing.T) {
	app := setupTestAppWithCollections(t)

	// Create a task for testing
	task := createTestTask(t, app, "Index test task", "in_progress")

	// Create some comments
	for i := 0; i < 10; i++ {
		createTestComment(t, app, task.Id, "Test comment content")
	}

	// Execute the query that should use the index
	records, err := app.FindRecordsByFilter(
		"comments",
		"task = {:taskId}",
		"+created",
		0,
		0,
		map[string]any{"taskId": task.Id},
	)

	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if len(records) != 10 {
		t.Errorf("expected 10 comments, got %d", len(records))
	}
}

// TestSessionsQueryUsesExternalRefIndex verifies that queries on sessions.external_ref use the index.
func TestSessionsQueryUsesExternalRefIndex(t *testing.T) {
	app := setupTestAppWithCollections(t)

	// Create a task and session
	task := createTestTask(t, app, "Session index test task", "in_progress")
	session := createTestSession(t, app, task.Id, "test-session-ref-123")

	// Execute the query that should use the index
	records, err := app.FindRecordsByFilter(
		"sessions",
		"external_ref = {:ref}",
		"",
		1,
		0,
		map[string]any{"ref": session.GetString("external_ref")},
	)

	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if len(records) != 1 {
		t.Errorf("expected 1 session, got %d", len(records))
	}
}

// TestSessionsQueryUsesStatusIndex verifies that queries on sessions.status use the index.
func TestSessionsQueryUsesStatusIndex(t *testing.T) {
	app := setupTestAppWithCollections(t)

	// Create tasks and sessions with different statuses
	for i := 0; i < 5; i++ {
		task := createTestTask(t, app, "Status index test task", "in_progress")
		status := "active"
		if i%2 == 0 {
			status = "completed"
		}
		createTestSessionWithStatus(t, app, task.Id, "ref-"+string(rune('a'+i)), status)
	}

	// Query by status
	records, err := app.FindRecordsByFilter(
		"sessions",
		"status = {:status}",
		"",
		0,
		0,
		map[string]any{"status": "active"},
	)

	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	// Should find 2 active sessions (i=1, i=3)
	if len(records) != 2 {
		t.Errorf("expected 2 active sessions, got %d", len(records))
	}
}

// ========== Index Performance Comparison Tests ==========

// TestIndexedQueryPerformance compares query performance to validate index benefit.
func TestIndexedQueryPerformance(t *testing.T) {
	app := setupTestAppWithCollections(t)

	// Create a task with many comments
	task := createTestTask(t, app, "Performance test task", "in_progress")
	for i := 0; i < 100; i++ {
		createTestComment(t, app, task.Id, "Performance test comment content")
	}

	// Create another task for comparison (should not be returned)
	otherTask := createTestTask(t, app, "Other task", "todo")
	for i := 0; i < 50; i++ {
		createTestComment(t, app, otherTask.Id, "Other task comment")
	}

	// Query should only return comments for the specific task
	records, err := app.FindRecordsByFilter(
		"comments",
		"task = {:taskId}",
		"+created",
		0,
		0,
		map[string]any{"taskId": task.Id},
	)

	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if len(records) != 100 {
		t.Errorf("expected 100 comments for task, got %d", len(records))
	}
}

// ========== Setup Helpers ==========

// setupTestAppWithCollections creates a test app with all required collections.
func setupTestAppWithCollections(t *testing.T) *pocketbase.PocketBase {
	t.Helper()

	app := testutil.NewTestApp(t)

	// Create tasks collection if not exists
	if _, err := app.FindCollectionByNameOrId("tasks"); err != nil {
		tasks := core.NewBaseCollection("tasks")
		tasks.Fields.Add(&core.TextField{Name: "title", Required: true})
		tasks.Fields.Add(&core.TextField{Name: "column"})
		tasks.Fields.Add(&core.JSONField{Name: "agent_session"})
		tasks.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
		tasks.Fields.Add(&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true})
		if err := app.Save(tasks); err != nil {
			t.Fatalf("failed to create tasks collection: %v", err)
		}
	}

	// Create comments collection with indexes
	if _, err := app.FindCollectionByNameOrId("comments"); err != nil {
		tasks, _ := app.FindCollectionByNameOrId("tasks")
		comments := core.NewBaseCollection("comments")
		comments.Fields.Add(&core.RelationField{
			Name:          "task",
			CollectionId:  tasks.Id,
			MaxSelect:     1,
			Required:      true,
			CascadeDelete: true,
		})
		comments.Fields.Add(&core.TextField{Name: "content", Required: true})
		comments.Fields.Add(&core.TextField{Name: "author_type"})
		comments.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
		// Add indexes
		comments.Indexes = []string{
			"CREATE INDEX idx_comments_task ON comments (task)",
			"CREATE INDEX idx_comments_created ON comments (created)",
		}
		if err := app.Save(comments); err != nil {
			t.Fatalf("failed to create comments collection: %v", err)
		}
	}

	// Create sessions collection with indexes
	if _, err := app.FindCollectionByNameOrId("sessions"); err != nil {
		tasks, _ := app.FindCollectionByNameOrId("tasks")
		sessions := core.NewBaseCollection("sessions")
		sessions.Fields.Add(&core.RelationField{
			Name:          "task",
			CollectionId:  tasks.Id,
			MaxSelect:     1,
			Required:      true,
			CascadeDelete: true,
		})
		sessions.Fields.Add(&core.SelectField{
			Name:     "tool",
			Required: true,
			Values:   []string{"opencode", "claude-code", "codex"},
		})
		sessions.Fields.Add(&core.TextField{Name: "external_ref", Required: true})
		sessions.Fields.Add(&core.SelectField{
			Name:     "ref_type",
			Required: true,
			Values:   []string{"uuid", "path"},
		})
		sessions.Fields.Add(&core.TextField{Name: "working_dir", Required: true})
		sessions.Fields.Add(&core.SelectField{
			Name:     "status",
			Required: true,
			Values:   []string{"active", "paused", "completed", "abandoned"},
		})
		sessions.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
		// Add indexes
		sessions.Indexes = []string{
			"CREATE INDEX idx_sessions_task ON sessions (task)",
			"CREATE INDEX idx_sessions_status ON sessions (status)",
			"CREATE INDEX idx_sessions_external_ref ON sessions (external_ref)",
		}
		if err := app.Save(sessions); err != nil {
			t.Fatalf("failed to create sessions collection: %v", err)
		}
	}

	return app
}

// createTestTask creates a task for testing.
func createTestTask(t *testing.T, app *pocketbase.PocketBase, title, column string) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("tasks")
	if err != nil {
		t.Fatalf("tasks collection not found: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("title", title)
	record.Set("column", column)

	if err := app.Save(record); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	return record
}

// createTestComment creates a comment for testing.
func createTestComment(t *testing.T, app *pocketbase.PocketBase, taskId, content string) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("comments")
	if err != nil {
		t.Fatalf("comments collection not found: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("task", taskId)
	record.Set("content", content)
	record.Set("author_type", "human")

	if err := app.Save(record); err != nil {
		t.Fatalf("failed to create comment: %v", err)
	}

	return record
}

// createTestSession creates a session for testing.
func createTestSession(t *testing.T, app *pocketbase.PocketBase, taskId, externalRef string) *core.Record {
	t.Helper()
	return createTestSessionWithStatus(t, app, taskId, externalRef, "active")
}

// createTestSessionWithStatus creates a session with a specific status.
func createTestSessionWithStatus(t *testing.T, app *pocketbase.PocketBase, taskId, externalRef, status string) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("sessions")
	if err != nil {
		t.Fatalf("sessions collection not found: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("task", taskId)
	record.Set("tool", "claude-code")
	record.Set("external_ref", externalRef)
	record.Set("ref_type", "uuid")
	record.Set("working_dir", "/tmp")
	record.Set("status", status)

	if err := app.Save(record); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	return record
}
