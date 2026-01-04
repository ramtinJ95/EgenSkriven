package commands

import (
	"testing"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ramtinJ95/EgenSkriven/internal/testutil"
)

func TestUpdateBlockedBy_AddsBlockingTask(t *testing.T) {
	current := []string{}
	add := []string{"task-1", "task-2"}
	remove := []string{}

	result := updateBlockedBy("my-task", current, add, remove)

	assert.Len(t, result, 2)
	assert.Contains(t, result, "task-1")
	assert.Contains(t, result, "task-2")
}

func TestUpdateBlockedBy_RemovesBlockingTask(t *testing.T) {
	current := []string{"task-1", "task-2", "task-3"}
	add := []string{}
	remove := []string{"task-2"}

	result := updateBlockedBy("my-task", current, add, remove)

	assert.Len(t, result, 2)
	assert.Contains(t, result, "task-1")
	assert.Contains(t, result, "task-3")
	assert.NotContains(t, result, "task-2")
}

func TestUpdateBlockedBy_PreventsSelfBlocking(t *testing.T) {
	current := []string{}
	add := []string{"my-task"} // Trying to block self
	remove := []string{}

	result := updateBlockedBy("my-task", current, add, remove)

	assert.Len(t, result, 0)
	assert.NotContains(t, result, "my-task")
}

func TestUpdateBlockedBy_CombinedAddAndRemove(t *testing.T) {
	current := []string{"task-1", "task-2"}
	add := []string{"task-3"}
	remove := []string{"task-1"}

	result := updateBlockedBy("my-task", current, add, remove)

	assert.Len(t, result, 2)
	assert.Contains(t, result, "task-2")
	assert.Contains(t, result, "task-3")
	assert.NotContains(t, result, "task-1")
}

func TestUpdateBlockedBy_NoDuplicates(t *testing.T) {
	current := []string{"task-1"}
	add := []string{"task-1", "task-2"} // task-1 already exists
	remove := []string{}

	result := updateBlockedBy("my-task", current, add, remove)

	// Should have task-1 once and task-2
	assert.Len(t, result, 2)
	count := 0
	for _, id := range result {
		if id == "task-1" {
			count++
		}
	}
	assert.Equal(t, 1, count, "task-1 should appear exactly once")
}

func TestGetTaskBlockedBy_Empty(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	task := createTestTaskWithBlockedBy(t, app, "Test", []string{})

	result := getTaskBlockedBy(task)

	assert.Len(t, result, 0)
}

func TestGetTaskBlockedBy_WithValues(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	task := createTestTaskWithBlockedBy(t, app, "Test", []string{"blocker-1", "blocker-2"})

	// Re-fetch from database to get the stored format
	task, err := app.FindRecordById("tasks", task.Id)
	require.NoError(t, err)

	result := getTaskBlockedBy(task)

	assert.Len(t, result, 2)
	assert.Contains(t, result, "blocker-1")
	assert.Contains(t, result, "blocker-2")
}

func TestGetTaskBlockedBy_Nil(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	// Create task without setting blocked_by
	collection, err := app.FindCollectionByNameOrId("tasks")
	require.NoError(t, err)

	record := core.NewRecord(collection)
	record.Set("title", "Test")
	record.Set("type", "feature")
	record.Set("priority", "medium")
	record.Set("column", "backlog")
	record.Set("position", 1000.0)
	record.Set("labels", []string{})
	record.Set("created_by", "cli")
	record.Set("history", []map[string]any{})
	// Note: blocked_by is intentionally not set

	require.NoError(t, app.Save(record))

	result := getTaskBlockedBy(record)

	assert.Len(t, result, 0)
}

func TestHasCircularDependency_NoCircle(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	// Create tasks: A -> B (A is blocked by B)
	taskB := createTestTaskWithBlockedBy(t, app, "Task B", []string{})
	taskA := createTestTaskWithBlockedBy(t, app, "Task A", []string{taskB.Id})

	// Re-fetch to get stored format
	taskA, _ = app.FindRecordById("tasks", taskA.Id)

	// Check if adding B blocked by A would create a cycle (it would)
	result := hasCircularDependency(app, taskA.Id, taskA)

	// taskA is blocked by taskB. If we check taskA's chain, we won't find taskA itself.
	// This tests the base case - no cycle in taskA's existing chain
	assert.False(t, result)
}

func TestHasCircularDependency_DirectCircle(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	// Create tasks: A -> B (A is blocked by B)
	taskB := createTestTaskWithBlockedBy(t, app, "Task B", []string{})
	taskA := createTestTaskWithBlockedBy(t, app, "Task A", []string{taskB.Id})

	// Re-fetch to get stored format
	taskB, _ = app.FindRecordById("tasks", taskB.Id)

	// Check if adding B blocked by A would create a cycle
	// B's blocked_by is empty, so we check if A is in B's chain (it's not)
	// But if we were to add A to B's blocked_by, then B->A->B would be a cycle
	// We need to check: does taskA's blocked_by chain contain B?
	taskA, _ = app.FindRecordById("tasks", taskA.Id)
	result := hasCircularDependency(app, taskB.Id, taskA)

	// taskA is blocked by taskB, so taskA's chain contains taskB
	assert.True(t, result)
}

func TestHasCircularDependency_IndirectCircle(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	// Create tasks: A -> B -> C (A blocked by B, B blocked by C)
	taskC := createTestTaskWithBlockedBy(t, app, "Task C", []string{})
	taskB := createTestTaskWithBlockedBy(t, app, "Task B", []string{taskC.Id})
	taskA := createTestTaskWithBlockedBy(t, app, "Task A", []string{taskB.Id})

	// Re-fetch to get stored format
	taskA, _ = app.FindRecordById("tasks", taskA.Id)

	// Check if adding C blocked by A would create a cycle (A -> B -> C -> A)
	result := hasCircularDependency(app, taskC.Id, taskA)

	// taskA's chain: taskB -> taskC, so taskC is in the chain
	assert.True(t, result)
}

// Helper function
func createTestTaskWithBlockedBy(t *testing.T, app *pocketbase.PocketBase, title string, blockedBy []string) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("tasks")
	require.NoError(t, err)

	record := core.NewRecord(collection)
	record.Set("title", title)
	record.Set("type", "feature")
	record.Set("priority", "medium")
	record.Set("column", "backlog")
	record.Set("position", 1000.0)
	record.Set("labels", []string{})
	record.Set("blocked_by", blockedBy)
	record.Set("created_by", "cli")
	record.Set("history", []map[string]any{})

	require.NoError(t, app.Save(record))

	return record
}
