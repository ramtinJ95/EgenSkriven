package commands

import (
	"testing"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"

	"github.com/ramtinJ95/EgenSkriven/internal/testutil"
)

func TestGetNextPosition_EmptyColumn(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	pos := GetNextPosition(app, "backlog")

	assert.Equal(t, DefaultPositionGap, pos)
}

func TestGetNextPosition_WithExistingTasks(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	// Create a task at position 1000
	createTestTask(t, app, "Task 1", "feature", "medium", "backlog", 1000)

	pos := GetNextPosition(app, "backlog")

	assert.Equal(t, 2000.0, pos)
}

func TestGetPositionBetween(t *testing.T) {
	tests := []struct {
		before   float64
		after    float64
		expected float64
	}{
		{1000, 2000, 1500},
		{0, 1000, 500},
		{500, 600, 550},
	}

	for _, tt := range tests {
		result := GetPositionBetween(tt.before, tt.after)
		assert.Equal(t, tt.expected, result)
	}
}

func TestGetPositionAtIndex_Top(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	createTestTask(t, app, "Task 1", "feature", "medium", "backlog", 1000)
	createTestTask(t, app, "Task 2", "feature", "medium", "backlog", 2000)

	pos := GetPositionAtIndex(app, "backlog", 0)

	// Should be half of the first task's position
	assert.Equal(t, 500.0, pos)
}

func TestGetPositionAtIndex_Bottom(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	createTestTask(t, app, "Task 1", "feature", "medium", "backlog", 1000)

	pos := GetPositionAtIndex(app, "backlog", -1)

	assert.Equal(t, 2000.0, pos)
}

func TestGetPositionAtIndex_Middle(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	createTestTask(t, app, "Task 1", "feature", "medium", "backlog", 1000)
	createTestTask(t, app, "Task 2", "feature", "medium", "backlog", 2000)
	createTestTask(t, app, "Task 3", "feature", "medium", "backlog", 3000)

	pos := GetPositionAtIndex(app, "backlog", 1)

	// Should be between first and second task
	assert.Equal(t, 1500.0, pos)
}

func TestGetPositionAtIndex_EmptyColumn(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	pos := GetPositionAtIndex(app, "backlog", 0)

	assert.Equal(t, DefaultPositionGap, pos)
}

func TestSortTasksByPosition(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	// Create tasks in non-sorted order
	task3 := createTestTask(t, app, "Task 3", "feature", "medium", "backlog", 3000)
	task1 := createTestTask(t, app, "Task 1", "feature", "medium", "backlog", 1000)
	task2 := createTestTask(t, app, "Task 2", "feature", "medium", "backlog", 2000)

	tasks := []*core.Record{task3, task1, task2}
	sortTasksByPosition(tasks)

	assert.Equal(t, task1.Id, tasks[0].Id)
	assert.Equal(t, task2.Id, tasks[1].Id)
	assert.Equal(t, task3.Id, tasks[2].Id)
}

// Helper functions for position tests

func setupTasksCollection(t *testing.T, app *pocketbase.PocketBase) {
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

func createTestTask(t *testing.T, app *pocketbase.PocketBase, title, taskType, priority, column string, position float64) *core.Record {
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
	record.Set("position", position)
	record.Set("labels", []string{})
	record.Set("blocked_by", []string{})
	record.Set("created_by", "cli")
	record.Set("history", []map[string]any{})

	if err := app.Save(record); err != nil {
		t.Fatalf("failed to create test task: %v", err)
	}

	return record
}
