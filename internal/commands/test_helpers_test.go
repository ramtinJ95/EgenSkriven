package commands

import (
	"testing"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/require"
)

// ========== Shared Test Setup Functions ==========
// These functions are used across multiple test files (block_test.go, comment_test.go, comments_test.go)
// to reduce code duplication and ensure consistent test setup.

// SetupTasksCollection creates the tasks collection with all required fields including need_input column.
// This is the canonical setup function - use this instead of file-specific variants.
func SetupTasksCollection(t *testing.T, app *pocketbase.PocketBase) {
	t.Helper()

	_, err := app.FindCollectionByNameOrId("tasks")
	if err == nil {
		return
	}

	collection := core.NewBaseCollection("tasks")
	collection.Fields.Add(&core.TextField{Name: "title", Required: true})
	collection.Fields.Add(&core.TextField{Name: "description"})
	collection.Fields.Add(&core.SelectField{
		Name:     "type",
		Required: true,
		Values:   []string{"bug", "feature", "chore"},
	})
	collection.Fields.Add(&core.SelectField{
		Name:     "priority",
		Required: true,
		Values:   []string{"low", "medium", "high", "urgent"},
	})
	collection.Fields.Add(&core.SelectField{
		Name:     "column",
		Required: true,
		Values:   []string{"backlog", "todo", "in_progress", "need_input", "review", "done"},
	})
	collection.Fields.Add(&core.NumberField{Name: "position", Required: true})
	collection.Fields.Add(&core.JSONField{Name: "labels"})
	collection.Fields.Add(&core.JSONField{Name: "blocked_by"})
	collection.Fields.Add(&core.SelectField{
		Name:     "created_by",
		Required: true,
		Values:   []string{"user", "agent", "cli"},
	})
	collection.Fields.Add(&core.TextField{Name: "created_by_agent"})
	collection.Fields.Add(&core.JSONField{Name: "history"})

	if err := app.Save(collection); err != nil {
		t.Fatalf("failed to create tasks collection: %v", err)
	}
}

// SetupCommentsCollection creates the comments collection without autodate fields.
// Use SetupCommentsCollectionWithAutodate if you need sorting by creation time.
func SetupCommentsCollection(t *testing.T, app *pocketbase.PocketBase) {
	t.Helper()

	_, err := app.FindCollectionByNameOrId("comments")
	if err == nil {
		return
	}

	collection := core.NewBaseCollection("comments")
	collection.Fields.Add(&core.TextField{Name: "task", Required: true})
	collection.Fields.Add(&core.TextField{Name: "content", Required: true})
	collection.Fields.Add(&core.SelectField{
		Name:     "author_type",
		Required: true,
		Values:   []string{"human", "agent"},
	})
	collection.Fields.Add(&core.TextField{Name: "author_id"})
	collection.Fields.Add(&core.JSONField{Name: "metadata"})

	if err := app.Save(collection); err != nil {
		t.Fatalf("failed to create comments collection: %v", err)
	}
}

// SetupCommentsCollectionWithAutodate creates the comments collection with autodate fields
// for proper sorting by creation time.
func SetupCommentsCollectionWithAutodate(t *testing.T, app *pocketbase.PocketBase) {
	t.Helper()

	_, err := app.FindCollectionByNameOrId("comments")
	if err == nil {
		return
	}

	collection := core.NewBaseCollection("comments")
	collection.Fields.Add(&core.TextField{Name: "task", Required: true})
	collection.Fields.Add(&core.TextField{Name: "content", Required: true})
	collection.Fields.Add(&core.SelectField{
		Name:     "author_type",
		Required: true,
		Values:   []string{"human", "agent"},
	})
	collection.Fields.Add(&core.TextField{Name: "author_id"})
	collection.Fields.Add(&core.JSONField{Name: "metadata"})
	collection.Fields.Add(&core.AutodateField{
		Name:     "created",
		OnCreate: true,
	})

	if err := app.Save(collection); err != nil {
		t.Fatalf("failed to create comments collection: %v", err)
	}
}

// CreateTestTask creates a task for testing with standard defaults.
func CreateTestTask(t *testing.T, app *pocketbase.PocketBase, title string, column string) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("tasks")
	require.NoError(t, err)

	record := core.NewRecord(collection)
	record.Set("title", title)
	record.Set("type", "feature")
	record.Set("priority", "medium")
	record.Set("column", column)
	record.Set("position", 1000.0)
	record.Set("labels", []string{})
	record.Set("blocked_by", []string{})
	record.Set("created_by", "cli")
	record.Set("history", []map[string]any{})

	require.NoError(t, app.Save(record))
	return record
}

// CreateTestComment creates a comment for testing.
func CreateTestComment(t *testing.T, app *pocketbase.PocketBase, taskId, content, authorType, authorId string) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("comments")
	require.NoError(t, err)

	record := core.NewRecord(collection)
	record.Set("task", taskId)
	record.Set("content", content)
	record.Set("author_type", authorType)
	record.Set("author_id", authorId)
	record.Set("metadata", map[string]any{})

	require.NoError(t, app.Save(record))
	return record
}

// GetCommentsForTask returns all comments for a given task ID.
func GetCommentsForTask(t *testing.T, app *pocketbase.PocketBase, taskId string) []*core.Record {
	t.Helper()

	records, err := app.FindRecordsByFilter(
		"comments",
		"task = {:taskId}",
		"", // No sorting - simplifies test setup
		0,
		0,
		dbx.Params{"taskId": taskId},
	)
	require.NoError(t, err)
	return records
}

// GetCommentsForTaskSorted returns all comments for a given task ID, sorted by creation time.
// Requires comments collection to be set up with SetupCommentsCollectionWithAutodate.
func GetCommentsForTaskSorted(t *testing.T, app *pocketbase.PocketBase, taskId string) []*core.Record {
	t.Helper()

	records, err := app.FindRecordsByFilter(
		"comments",
		"task = {:taskId}",
		"+created", // Sort by creation time ascending
		0,
		0,
		dbx.Params{"taskId": taskId},
	)
	require.NoError(t, err)
	return records
}
