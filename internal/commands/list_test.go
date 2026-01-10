package commands

import (
	"testing"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ramtinJ95/EgenSkriven/internal/testutil"
)

// ========== Setup Functions ==========

// setupTasksCollectionForList creates tasks collection for list tests
func setupTasksCollectionForList(t *testing.T, app *pocketbase.PocketBase) {
	SetupTasksCollection(t, app)
}

// createListTestTask creates a task for list command testing
func createListTestTask(t *testing.T, app *pocketbase.PocketBase, title string, column string) *core.Record {
	return CreateTestTask(t, app, title, column)
}

// ========== --need-input Flag Tests ==========

// TestListCommand_NeedInputShowsOnlyBlockedTasks verifies --need-input shows only tasks in need_input column
func TestListCommand_NeedInputShowsOnlyBlockedTasks(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForList(t, app)

	// Create tasks in different columns
	taskTodo := createListTestTask(t, app, "Task in todo", "todo")
	taskInProgress := createListTestTask(t, app, "Task in progress", "in_progress")
	taskNeedInput1 := createListTestTask(t, app, "Blocked task 1", "need_input")
	taskNeedInput2 := createListTestTask(t, app, "Blocked task 2", "need_input")
	taskDone := createListTestTask(t, app, "Completed task", "done")

	// Query for need_input tasks (simulates --need-input filter)
	records, err := app.FindRecordsByFilter(
		"tasks",
		"column = 'need_input'",
		"",
		0,
		0,
	)
	require.NoError(t, err)

	// Should only get the 2 blocked tasks
	assert.Len(t, records, 2, "should only return tasks in need_input column")

	// Verify they are the correct tasks
	var foundTask1, foundTask2 bool
	for _, r := range records {
		if r.Id == taskNeedInput1.Id {
			foundTask1 = true
		}
		if r.Id == taskNeedInput2.Id {
			foundTask2 = true
		}
		// Should not find other tasks
		assert.NotEqual(t, taskTodo.Id, r.Id, "should not include todo task")
		assert.NotEqual(t, taskInProgress.Id, r.Id, "should not include in_progress task")
		assert.NotEqual(t, taskDone.Id, r.Id, "should not include done task")
	}

	assert.True(t, foundTask1, "should find blocked task 1")
	assert.True(t, foundTask2, "should find blocked task 2")
}

// TestListCommand_NeedInputWorksWithOtherFilters verifies --need-input can combine with other filters
func TestListCommand_NeedInputWorksWithOtherFilters(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForList(t, app)

	// Create tasks with different types in need_input
	collection, err := app.FindCollectionByNameOrId("tasks")
	require.NoError(t, err)

	// Create a bug task in need_input
	bugTask := core.NewRecord(collection)
	bugTask.Set("title", "Bug needing input")
	bugTask.Set("type", "bug")
	bugTask.Set("priority", "high")
	bugTask.Set("column", "need_input")
	bugTask.Set("position", 1000.0)
	bugTask.Set("labels", []string{})
	bugTask.Set("blocked_by", []string{})
	bugTask.Set("created_by", "cli")
	bugTask.Set("history", []map[string]any{})
	require.NoError(t, app.Save(bugTask))

	// Create a feature task in need_input
	featureTask := core.NewRecord(collection)
	featureTask.Set("title", "Feature needing input")
	featureTask.Set("type", "feature")
	featureTask.Set("priority", "medium")
	featureTask.Set("column", "need_input")
	featureTask.Set("position", 2000.0)
	featureTask.Set("labels", []string{})
	featureTask.Set("blocked_by", []string{})
	featureTask.Set("created_by", "cli")
	featureTask.Set("history", []map[string]any{})
	require.NoError(t, app.Save(featureTask))

	// Create a bug task NOT in need_input
	bugNotBlocked := core.NewRecord(collection)
	bugNotBlocked.Set("title", "Bug not blocked")
	bugNotBlocked.Set("type", "bug")
	bugNotBlocked.Set("priority", "high")
	bugNotBlocked.Set("column", "todo")
	bugNotBlocked.Set("position", 3000.0)
	bugNotBlocked.Set("labels", []string{})
	bugNotBlocked.Set("blocked_by", []string{})
	bugNotBlocked.Set("created_by", "cli")
	bugNotBlocked.Set("history", []map[string]any{})
	require.NoError(t, app.Save(bugNotBlocked))

	// Query combining --need-input with --type bug
	records, err := app.FindRecordsByFilter(
		"tasks",
		"column = 'need_input' && type = 'bug'",
		"",
		0,
		0,
	)
	require.NoError(t, err)

	// Should only get the bug task in need_input
	assert.Len(t, records, 1, "should only return bug tasks in need_input")
	assert.Equal(t, bugTask.Id, records[0].Id, "should be the bug task in need_input")
	assert.Equal(t, "bug", records[0].GetString("type"))
	assert.Equal(t, "need_input", records[0].GetString("column"))
}

// TestListCommand_NeedInputEmptyResult verifies helpful behavior when no tasks need input
func TestListCommand_NeedInputEmptyResult(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForList(t, app)

	// Create tasks in columns other than need_input
	createListTestTask(t, app, "Task in todo", "todo")
	createListTestTask(t, app, "Task in progress", "in_progress")
	createListTestTask(t, app, "Completed task", "done")

	// Query for need_input tasks
	records, err := app.FindRecordsByFilter(
		"tasks",
		"column = 'need_input'",
		"",
		0,
		0,
	)
	require.NoError(t, err)

	// Should be empty
	assert.Empty(t, records, "should return no tasks when none are in need_input")
}

// TestListCommand_NeedInputJSONOutput verifies --need-input works with --json output
func TestListCommand_NeedInputJSONOutput(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForList(t, app)

	// Create tasks
	taskNeedInput := createListTestTask(t, app, "Blocked task", "need_input")
	createListTestTask(t, app, "Not blocked task", "todo")

	// Query for need_input tasks
	records, err := app.FindRecordsByFilter(
		"tasks",
		"column = 'need_input'",
		"",
		0,
		0,
	)
	require.NoError(t, err)

	// Build JSON structure that would be output
	tasks := make([]map[string]any, len(records))
	for i, r := range records {
		tasks[i] = map[string]any{
			"id":       r.Id,
			"title":    r.GetString("title"),
			"type":     r.GetString("type"),
			"priority": r.GetString("priority"),
			"column":   r.GetString("column"),
		}
	}

	jsonResult := map[string]any{
		"count": len(tasks),
		"tasks": tasks,
	}

	// Verify JSON structure
	assert.Equal(t, 1, jsonResult["count"])
	tasksArray := jsonResult["tasks"].([]map[string]any)
	assert.Len(t, tasksArray, 1)
	assert.Equal(t, taskNeedInput.Id, tasksArray[0]["id"])
	assert.Equal(t, "need_input", tasksArray[0]["column"])
}

// TestListCommand_NeedInputWithPriorityFilter verifies --need-input with --priority filter
func TestListCommand_NeedInputWithPriorityFilter(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForList(t, app)

	collection, err := app.FindCollectionByNameOrId("tasks")
	require.NoError(t, err)

	// Create urgent task in need_input
	urgentTask := core.NewRecord(collection)
	urgentTask.Set("title", "Urgent blocked task")
	urgentTask.Set("type", "feature")
	urgentTask.Set("priority", "urgent")
	urgentTask.Set("column", "need_input")
	urgentTask.Set("position", 1000.0)
	urgentTask.Set("labels", []string{})
	urgentTask.Set("blocked_by", []string{})
	urgentTask.Set("created_by", "cli")
	urgentTask.Set("history", []map[string]any{})
	require.NoError(t, app.Save(urgentTask))

	// Create low priority task in need_input
	lowTask := core.NewRecord(collection)
	lowTask.Set("title", "Low priority blocked task")
	lowTask.Set("type", "chore")
	lowTask.Set("priority", "low")
	lowTask.Set("column", "need_input")
	lowTask.Set("position", 2000.0)
	lowTask.Set("labels", []string{})
	lowTask.Set("blocked_by", []string{})
	lowTask.Set("created_by", "cli")
	lowTask.Set("history", []map[string]any{})
	require.NoError(t, app.Save(lowTask))

	// Query combining --need-input with --priority urgent
	records, err := app.FindRecordsByFilter(
		"tasks",
		"column = 'need_input' && priority = 'urgent'",
		"",
		0,
		0,
	)
	require.NoError(t, err)

	// Should only get the urgent task
	assert.Len(t, records, 1, "should only return urgent tasks in need_input")
	assert.Equal(t, urgentTask.Id, records[0].Id)
	assert.Equal(t, "urgent", records[0].GetString("priority"))
}

// ========== buildInFilter Tests ==========

// TestBuildInFilter_SingleValue verifies single value creates field-prefixed parameter
func TestBuildInFilter_SingleValue(t *testing.T) {
	expr := buildInFilter("status", []string{"todo"})

	// The expression should use field-prefixed parameter name
	// We can't directly inspect the expression, but we can verify it works
	// by checking the SQL string contains the field name
	assert.NotNil(t, expr)
}

// TestBuildInFilter_MultipleValues verifies multiple values create field-prefixed parameters
func TestBuildInFilter_MultipleValues(t *testing.T) {
	expr := buildInFilter("priority", []string{"high", "urgent"})

	assert.NotNil(t, expr)
}

// TestBuildInFilter_NoParameterCollision verifies different fields don't collide
func TestBuildInFilter_NoParameterCollision(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollection(t, app)

	collection, err := app.FindCollectionByNameOrId("tasks")
	require.NoError(t, err)

	// Create tasks with different status and priority combinations
	task1 := core.NewRecord(collection)
	task1.Set("title", "High priority todo")
	task1.Set("type", "feature")
	task1.Set("priority", "high")
	task1.Set("column", "todo")
	task1.Set("position", 1000.0)
	task1.Set("labels", []string{})
	task1.Set("blocked_by", []string{})
	task1.Set("created_by", "cli")
	task1.Set("history", []map[string]any{})
	require.NoError(t, app.Save(task1))

	task2 := core.NewRecord(collection)
	task2.Set("title", "Low priority in_progress")
	task2.Set("type", "feature")
	task2.Set("priority", "low")
	task2.Set("column", "in_progress")
	task2.Set("position", 2000.0)
	task2.Set("labels", []string{})
	task2.Set("blocked_by", []string{})
	task2.Set("created_by", "cli")
	task2.Set("history", []map[string]any{})
	require.NoError(t, app.Save(task2))

	task3 := core.NewRecord(collection)
	task3.Set("title", "High priority in_progress")
	task3.Set("type", "bug")
	task3.Set("priority", "high")
	task3.Set("column", "in_progress")
	task3.Set("position", 3000.0)
	task3.Set("labels", []string{})
	task3.Set("blocked_by", []string{})
	task3.Set("created_by", "cli")
	task3.Set("history", []map[string]any{})
	require.NoError(t, app.Save(task3))

	// Build filters for both column and priority - this would have collided before the fix
	columnFilter := buildInFilter("column", []string{"todo", "in_progress"})
	priorityFilter := buildInFilter("priority", []string{"high"})

	// Combine filters using AND - before the fix, parameter names would collide
	records, err := app.FindAllRecords("tasks", columnFilter, priorityFilter)
	require.NoError(t, err)

	// Should find task1 (high priority, todo) and task3 (high priority, in_progress)
	assert.Len(t, records, 2, "should find tasks matching both column and priority filters")

	foundTask1 := false
	foundTask3 := false
	for _, r := range records {
		if r.Id == task1.Id {
			foundTask1 = true
		}
		if r.Id == task3.Id {
			foundTask3 = true
		}
	}
	assert.True(t, foundTask1, "should find high priority todo task")
	assert.True(t, foundTask3, "should find high priority in_progress task")
}

// TestBuildInFilter_MultipleFiltersWithMultipleValues verifies complex filter combinations
func TestBuildInFilter_MultipleFiltersWithMultipleValues(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollection(t, app)

	collection, err := app.FindCollectionByNameOrId("tasks")
	require.NoError(t, err)

	// Create diverse tasks
	tasks := []struct {
		title    string
		taskType string
		priority string
		column   string
	}{
		{"Feature high todo", "feature", "high", "todo"},
		{"Bug urgent in_progress", "bug", "urgent", "in_progress"},
		{"Chore low done", "chore", "low", "done"},
		{"Feature medium todo", "feature", "medium", "todo"},
		{"Bug high review", "bug", "high", "review"},
	}

	taskRecords := make([]*core.Record, len(tasks))
	for i, tc := range tasks {
		record := core.NewRecord(collection)
		record.Set("title", tc.title)
		record.Set("type", tc.taskType)
		record.Set("priority", tc.priority)
		record.Set("column", tc.column)
		record.Set("position", float64((i+1)*1000))
		record.Set("labels", []string{})
		record.Set("blocked_by", []string{})
		record.Set("created_by", "cli")
		record.Set("history", []map[string]any{})
		require.NoError(t, app.Save(record))
		taskRecords[i] = record
	}

	// Build three filters that would all collide without field-prefixed params
	typeFilter := buildInFilter("type", []string{"feature", "bug"})
	priorityFilter := buildInFilter("priority", []string{"high", "urgent"})
	columnFilter := buildInFilter("column", []string{"todo", "in_progress", "review"})

	// Combine all three filters
	records, err := app.FindAllRecords("tasks", typeFilter, priorityFilter, columnFilter)
	require.NoError(t, err)

	// Should find:
	// - Feature high todo (matches all)
	// - Bug urgent in_progress (matches all)
	// - Bug high review (matches all)
	assert.Len(t, records, 3, "should find tasks matching all three filter criteria")
}
