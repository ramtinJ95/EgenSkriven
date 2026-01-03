package resolver

import (
	"testing"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ramtinj/egenskriven/internal/testutil"
)

func TestResolveTask_ExactID(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	// Create a test task
	task := createTestTask(t, app, "Test Task", "feature", "medium", "backlog")

	// Resolve by exact ID
	resolution, err := ResolveTask(app, task.Id)

	require.NoError(t, err)
	assert.False(t, resolution.IsAmbiguous())
	assert.False(t, resolution.IsNotFound())
	assert.Equal(t, task.Id, resolution.Task.Id)
}

func TestResolveTask_IDPrefix(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	task := createTestTask(t, app, "Test Task", "feature", "medium", "backlog")

	// Resolve by ID prefix (first 6 characters)
	prefix := task.Id[:6]
	resolution, err := ResolveTask(app, prefix)

	require.NoError(t, err)
	assert.False(t, resolution.IsAmbiguous())
	assert.Equal(t, task.Id, resolution.Task.Id)
}

func TestResolveTask_TitleMatch(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	task := createTestTask(t, app, "Fix login authentication bug", "bug", "high", "todo")

	// Resolve by title substring (case-insensitive)
	resolution, err := ResolveTask(app, "login auth")

	require.NoError(t, err)
	assert.False(t, resolution.IsAmbiguous())
	assert.Equal(t, task.Id, resolution.Task.Id)
}

func TestResolveTask_Ambiguous(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	// Create multiple tasks with similar titles
	createTestTask(t, app, "Fix login bug", "bug", "high", "todo")
	createTestTask(t, app, "Fix login crash", "bug", "urgent", "todo")

	// Resolve with ambiguous reference
	resolution, err := ResolveTask(app, "login")

	require.NoError(t, err)
	assert.True(t, resolution.IsAmbiguous())
	assert.Len(t, resolution.Matches, 2)
}

func TestResolveTask_NotFound(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	resolution, err := ResolveTask(app, "nonexistent")

	require.NoError(t, err)
	assert.True(t, resolution.IsNotFound())
	assert.Nil(t, resolution.Task)
}

func TestMustResolve_Success(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	task := createTestTask(t, app, "Test Task", "feature", "medium", "backlog")

	resolved, err := MustResolve(app, task.Id)

	require.NoError(t, err)
	assert.Equal(t, task.Id, resolved.Id)
}

func TestMustResolve_NotFound(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	_, err := MustResolve(app, "nonexistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no task found")
}

func TestMustResolve_Ambiguous(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	createTestTask(t, app, "Task A", "feature", "medium", "backlog")
	createTestTask(t, app, "Task B", "feature", "medium", "backlog")

	_, err := MustResolve(app, "Task")

	require.Error(t, err)
	_, ok := err.(*AmbiguousError)
	assert.True(t, ok, "expected AmbiguousError")
}

// Helper functions

func setupTasksCollection(t *testing.T, app *pocketbase.PocketBase) {
	t.Helper()

	// Check if collection exists
	_, err := app.FindCollectionByNameOrId("tasks")
	if err == nil {
		return // Collection already exists
	}

	// Create tasks collection
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
		Values:   []string{"backlog", "todo", "in_progress", "review", "done"},
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

func createTestTask(t *testing.T, app *pocketbase.PocketBase, title, taskType, priority, column string) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("tasks")
	if err != nil {
		t.Fatalf("tasks collection not found: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("title", title)
	record.Set("type", taskType)
	record.Set("priority", priority)
	record.Set("column", column)
	record.Set("position", 1000.0)
	record.Set("labels", []string{})
	record.Set("blocked_by", []string{})
	record.Set("created_by", "cli")
	record.Set("history", []map[string]any{})

	if err := app.Save(record); err != nil {
		t.Fatalf("failed to create test task: %v", err)
	}

	return record
}
