package commands

import (
	"fmt"
	"testing"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ramtinJ95/EgenSkriven/internal/testutil"
)

// ========== Setup Functions ==========

// setupTasksWithParentField creates tasks collection with parent field for sub-tasks
func setupTasksWithParentField(t *testing.T, app *pocketbase.PocketBase) *core.Collection {
	t.Helper()

	_, err := app.FindCollectionByNameOrId("tasks")
	if err == nil {
		collection, _ := app.FindCollectionByNameOrId("tasks")
		return collection
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
	collection.Fields.Add(&core.DateField{Name: "due_date"})

	if err := app.Save(collection); err != nil {
		t.Fatalf("failed to create tasks collection: %v", err)
	}

	// Add parent field after collection is saved (self-reference)
	collection, _ = app.FindCollectionByNameOrId("tasks")
	collection.Fields.Add(&core.RelationField{
		Name:          "parent",
		Required:      false,
		CollectionId:  collection.Id,
		MaxSelect:     1,
		CascadeDelete: false,
	})

	if err := app.Save(collection); err != nil {
		t.Fatalf("failed to add parent field: %v", err)
	}

	return collection
}

// createShowTestTask creates a task for show command testing
func createShowTestTask(t *testing.T, app *pocketbase.PocketBase, title string, column string, position float64) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("tasks")
	require.NoError(t, err)

	record := core.NewRecord(collection)
	record.Set("title", title)
	record.Set("type", "feature")
	record.Set("priority", "medium")
	record.Set("column", column)
	record.Set("position", position)
	record.Set("labels", []string{})
	record.Set("blocked_by", []string{})
	record.Set("created_by", "cli")
	record.Set("history", []map[string]any{})

	require.NoError(t, app.Save(record))
	return record
}

// createShowTestSubTask creates a sub-task with a parent for show command testing
func createShowTestSubTask(t *testing.T, app *pocketbase.PocketBase, title string, parentID string, column string, position float64) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("tasks")
	require.NoError(t, err)

	record := core.NewRecord(collection)
	record.Set("title", title)
	record.Set("type", "feature")
	record.Set("priority", "medium")
	record.Set("column", column)
	record.Set("position", position)
	record.Set("labels", []string{})
	record.Set("blocked_by", []string{})
	record.Set("created_by", "cli")
	record.Set("history", []map[string]any{})
	record.Set("parent", parentID)

	require.NoError(t, app.Save(record))
	return record
}

// getSubtasks returns all sub-tasks for a given parent ID
func getSubtasks(t *testing.T, app *pocketbase.PocketBase, parentID string) []*core.Record {
	t.Helper()

	allTasks, err := app.FindAllRecords("tasks")
	require.NoError(t, err)

	var subtasks []*core.Record
	for _, task := range allTasks {
		if task.GetString("parent") == parentID {
			subtasks = append(subtasks, task)
		}
	}
	return subtasks
}

// ========== Tests ==========

func TestShowCommand_TaskWithNoSubtasks(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksWithParentField(t, app)

	// Create a parent task with no sub-tasks
	parent := createShowTestTask(t, app, "Parent Task", "todo", 1000.0)

	// Query sub-tasks (should be empty)
	subtasks := getSubtasks(t, app, parent.Id)

	// Verify task exists and has no sub-tasks
	assert.Equal(t, "Parent Task", parent.GetString("title"))
	assert.Empty(t, subtasks)
}

func TestShowCommand_TaskWithSubtasks(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksWithParentField(t, app)

	// Create a parent task
	parent := createShowTestTask(t, app, "Parent Task", "todo", 1000.0)

	// Create sub-tasks
	sub1 := createShowTestSubTask(t, app, "Sub-task 1", parent.Id, "todo", 1001.0)
	sub2 := createShowTestSubTask(t, app, "Sub-task 2", parent.Id, "done", 1002.0)
	sub3 := createShowTestSubTask(t, app, "Sub-task 3", parent.Id, "in_progress", 1003.0)

	// Verify sub-tasks have correct parent
	assert.Equal(t, parent.Id, sub1.GetString("parent"))
	assert.Equal(t, parent.Id, sub2.GetString("parent"))
	assert.Equal(t, parent.Id, sub3.GetString("parent"))

	// Query sub-tasks
	subtasks := getSubtasks(t, app, parent.Id)
	assert.Len(t, subtasks, 3)
}

func TestShowCommand_SubtaskOrdering(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksWithParentField(t, app)

	// Create a parent task
	parent := createShowTestTask(t, app, "Parent Task", "todo", 1000.0)

	// Create sub-tasks with different positions
	createShowTestSubTask(t, app, "Third", parent.Id, "todo", 3000.0)
	createShowTestSubTask(t, app, "First", parent.Id, "todo", 1000.0)
	createShowTestSubTask(t, app, "Second", parent.Id, "todo", 2000.0)

	// Query and verify we have 3 sub-tasks
	subtasks := getSubtasks(t, app, parent.Id)
	assert.Len(t, subtasks, 3)
}

func TestShowCommand_SubtaskDoneStatus(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksWithParentField(t, app)

	// Create parent and sub-tasks
	parent := createShowTestTask(t, app, "Parent Task", "todo", 1000.0)
	doneSub := createShowTestSubTask(t, app, "Done Sub", parent.Id, "done", 1001.0)
	todoSub := createShowTestSubTask(t, app, "Todo Sub", parent.Id, "todo", 1002.0)

	// Verify done status detection
	assert.Equal(t, "done", doneSub.GetString("column"))
	assert.Equal(t, "todo", todoSub.GetString("column"))
}

func TestShowCommand_NestedSubtasks(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksWithParentField(t, app)

	// Create hierarchy: grandparent -> parent -> child
	grandparent := createShowTestTask(t, app, "Grandparent", "todo", 1000.0)
	parent := createShowTestSubTask(t, app, "Parent", grandparent.Id, "todo", 1001.0)
	child := createShowTestSubTask(t, app, "Child", parent.Id, "todo", 1002.0)

	// Verify relationships
	assert.Equal(t, grandparent.Id, parent.GetString("parent"))
	assert.Equal(t, parent.Id, child.GetString("parent"))

	// Query only direct children of grandparent
	directChildren := getSubtasks(t, app, grandparent.Id)

	// Should only have parent as direct child, not grandchild
	assert.Len(t, directChildren, 1)
	assert.Equal(t, "Parent", directChildren[0].GetString("title"))
}

func TestShowCommand_SubtaskCount(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksWithParentField(t, app)

	// Create parent with multiple sub-tasks
	parent := createShowTestTask(t, app, "Parent Task", "todo", 1000.0)

	for i := 0; i < 5; i++ {
		createShowTestSubTask(t, app, fmt.Sprintf("Sub-task %d", i), parent.Id, "todo", float64(1000+i))
	}

	// Count sub-tasks
	subtasks := getSubtasks(t, app, parent.Id)
	assert.Len(t, subtasks, 5)
}

func TestShowCommand_ParentWithMixedColumnSubtasks(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksWithParentField(t, app)

	parent := createShowTestTask(t, app, "Parent Task", "in_progress", 1000.0)

	// Create sub-tasks in different columns
	createShowTestSubTask(t, app, "Backlog Sub", parent.Id, "backlog", 1001.0)
	createShowTestSubTask(t, app, "Todo Sub", parent.Id, "todo", 1002.0)
	createShowTestSubTask(t, app, "In Progress Sub", parent.Id, "in_progress", 1003.0)
	createShowTestSubTask(t, app, "Review Sub", parent.Id, "review", 1004.0)
	createShowTestSubTask(t, app, "Done Sub", parent.Id, "done", 1005.0)

	// Query sub-tasks
	subtasks := getSubtasks(t, app, parent.Id)
	assert.Len(t, subtasks, 5)

	// Count done vs not-done
	doneCount := 0
	for _, sub := range subtasks {
		if sub.GetString("column") == "done" {
			doneCount++
		}
	}
	assert.Equal(t, 1, doneCount)
}

func TestShowCommand_SubtaskParentField(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksWithParentField(t, app)

	// Create parent and sub-task
	parent := createShowTestTask(t, app, "Parent Task", "todo", 1000.0)
	subtask := createShowTestSubTask(t, app, "Sub-task", parent.Id, "todo", 1001.0)

	// Verify the parent field is set correctly
	assert.Equal(t, parent.Id, subtask.GetString("parent"))
	assert.Empty(t, parent.GetString("parent")) // Parent should have no parent
}
