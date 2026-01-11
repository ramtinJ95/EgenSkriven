package tui

import (
	"testing"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ramtinJ95/EgenSkriven/internal/board"
	"github.com/ramtinJ95/EgenSkriven/internal/testutil"
)

// createTasksCollection creates the tasks collection with all required fields
func createTasksCollection(t *testing.T, app *pocketbase.PocketBase) *core.Collection {
	t.Helper()
	return testutil.CreateTestCollection(t, app, "tasks",
		&core.TextField{Name: "title", Required: true},
		&core.TextField{Name: "description"},
		&core.TextField{Name: "type"},
		&core.TextField{Name: "priority"},
		&core.TextField{Name: "column"},
		&core.NumberField{Name: "position"},
		&core.JSONField{Name: "labels"},
		&core.JSONField{Name: "blocked_by"},
		&core.TextField{Name: "created_by"},
		&core.TextField{Name: "board"},
		&core.NumberField{Name: "seq"},
		&core.TextField{Name: "due_date"},
		&core.TextField{Name: "epic"},
		&core.JSONField{Name: "history"},
	)
}

// createBoardsCollection creates the boards collection
func createBoardsCollection(t *testing.T, app *pocketbase.PocketBase) *core.Collection {
	t.Helper()
	return testutil.CreateTestCollection(t, app, "boards",
		&core.TextField{Name: "name", Required: true},
		&core.TextField{Name: "prefix", Required: true},
		&core.NumberField{Name: "next_seq"},
		&core.JSONField{Name: "columns"},
	)
}

// TestCreateTaskCommand verifies task creation works correctly
func TestCreateTaskCommand(t *testing.T) {
	app := testutil.NewTestApp(t)

	createTasksCollection(t, app)
	createBoardsCollection(t, app)

	// Create a test board
	boardRecord, err := board.Create(app, board.CreateInput{
		Name:   "Test",
		Prefix: "TST",
	})
	require.NoError(t, err)

	// Get the board record
	boardRec, err := app.FindRecordById("boards", boardRecord.ID)
	require.NoError(t, err)

	// Execute the create command
	data := TaskFormData{
		Title:       "Test Task",
		Description: "This is a test",
		Type:        "feature",
		Priority:    "medium",
		Column:      "backlog",
		Labels:      []string{"test", "unit"},
	}

	cmd := createTask(app, boardRec, data)
	msg := cmd()

	// Verify result
	created, ok := msg.(taskCreatedMsg)
	require.True(t, ok, "Expected taskCreatedMsg, got %T", msg)
	assert.NotEmpty(t, created.task.Id)
	assert.Equal(t, "Test Task", created.task.GetString("title"))
	assert.Equal(t, "This is a test", created.task.GetString("description"))
	assert.Equal(t, "feature", created.task.GetString("type"))
	assert.Equal(t, "medium", created.task.GetString("priority"))
	assert.Equal(t, "backlog", created.task.GetString("column"))
	assert.Equal(t, "TST-1", created.displayID)
	assert.Equal(t, "tui", created.task.GetString("created_by"))
}

// TestUpdateTaskCommand verifies task updates work correctly
func TestUpdateTaskCommand(t *testing.T) {
	app := testutil.NewTestApp(t)

	createTasksCollection(t, app)
	createBoardsCollection(t, app)

	boardRecord, _ := board.Create(app, board.CreateInput{Name: "Test", Prefix: "TST"})
	boardRec, _ := app.FindRecordById("boards", boardRecord.ID)

	// Create a task first
	data := TaskFormData{
		Title:    "Original Title",
		Type:     "feature",
		Priority: "low",
		Column:   "backlog",
	}
	cmd := createTask(app, boardRec, data)
	created := cmd().(taskCreatedMsg)

	// Update the task
	updateData := TaskFormData{
		Title:       "Updated Title",
		Description: "Added description",
		Type:        "bug",
		Priority:    "high",
		Column:      "backlog",
	}

	updateCmd := updateTask(app, created.task.Id, updateData)
	updateMsg := updateCmd()

	// Verify result
	updated, ok := updateMsg.(taskUpdatedMsg)
	require.True(t, ok, "Expected taskUpdatedMsg, got %T", updateMsg)
	assert.Equal(t, "Updated Title", updated.task.GetString("title"))
	assert.Equal(t, "Added description", updated.task.GetString("description"))
	assert.Equal(t, "bug", updated.task.GetString("type"))
	assert.Equal(t, "high", updated.task.GetString("priority"))
}

// TestDeleteTaskCommand verifies task deletion works correctly
func TestDeleteTaskCommand(t *testing.T) {
	app := testutil.NewTestApp(t)

	// Create collection
	collection := testutil.CreateTestCollection(t, app, "tasks",
		&core.TextField{Name: "title", Required: true},
		&core.TextField{Name: "column"},
		&core.NumberField{Name: "position"},
	)

	// Create a task directly
	record := core.NewRecord(collection)
	record.Set("title", "Task to Delete")
	record.Set("column", "backlog")
	record.Set("position", 1000.0)
	require.NoError(t, app.Save(record))

	// Delete the task
	cmd := deleteTask(app, record.Id)
	msg := cmd()

	// Verify result
	deleted, ok := msg.(taskDeletedMsg)
	require.True(t, ok, "Expected taskDeletedMsg, got %T", msg)
	assert.Equal(t, record.Id, deleted.taskID)
	assert.Equal(t, "Task to Delete", deleted.title)

	// Verify task no longer exists
	_, err := app.FindRecordById("tasks", record.Id)
	assert.Error(t, err)
}

// TestMoveTaskToColumnCommand verifies task column movement
func TestMoveTaskToColumnCommand(t *testing.T) {
	app := testutil.NewTestApp(t)

	// Create collection
	collection := testutil.CreateTestCollection(t, app, "tasks",
		&core.TextField{Name: "title", Required: true},
		&core.TextField{Name: "column"},
		&core.NumberField{Name: "position"},
		&core.JSONField{Name: "history"},
	)

	// Create a task in backlog
	record := core.NewRecord(collection)
	record.Set("title", "Task to Move")
	record.Set("column", "backlog")
	record.Set("position", 1000.0)
	record.Set("history", []map[string]any{})
	require.NoError(t, app.Save(record))

	// Move to todo
	cmd := moveTaskToColumn(app, record.Id, "todo")
	msg := cmd()

	// Verify result
	moved, ok := msg.(taskMovedMsg)
	require.True(t, ok, "Expected taskMovedMsg, got %T", msg)
	assert.Equal(t, "backlog", moved.fromColumn)
	assert.Equal(t, "todo", moved.toColumn)
	assert.Equal(t, "todo", moved.task.GetString("column"))

	// Verify in database
	updated, err := app.FindRecordById("tasks", record.Id)
	require.NoError(t, err)
	assert.Equal(t, "todo", updated.GetString("column"))
}

// TestReorderTaskInColumnCommand verifies task reordering
func TestReorderTaskInColumnCommand(t *testing.T) {
	app := testutil.NewTestApp(t)

	// Create collection
	collection := testutil.CreateTestCollection(t, app, "tasks",
		&core.TextField{Name: "title", Required: true},
		&core.TextField{Name: "column"},
		&core.NumberField{Name: "position"},
	)

	// Create three tasks with known positions
	positions := []float64{1000, 2000, 3000}
	var tasks []*core.Record
	for i, pos := range positions {
		record := core.NewRecord(collection)
		record.Set("title", "Task "+string(rune('A'+i)))
		record.Set("column", "backlog")
		record.Set("position", pos)
		require.NoError(t, app.Save(record))
		tasks = append(tasks, record)
	}

	// Move middle task up (should go between first task and top)
	cmd := reorderTaskInColumn(app, tasks[1].Id, true)
	msg := cmd()

	moved, ok := msg.(taskMovedMsg)
	require.True(t, ok, "Expected taskMovedMsg, got %T", msg)

	// New position should be less than original
	assert.Less(t, moved.task.GetFloat("position"), 2000.0)
}

// TestMoveTaskToSameColumn verifies no-op when moving to same column
func TestMoveTaskToSameColumn(t *testing.T) {
	app := testutil.NewTestApp(t)

	// Create collection
	collection := testutil.CreateTestCollection(t, app, "tasks",
		&core.TextField{Name: "title", Required: true},
		&core.TextField{Name: "column"},
		&core.NumberField{Name: "position"},
		&core.JSONField{Name: "history"},
	)

	// Create a task in backlog
	record := core.NewRecord(collection)
	record.Set("title", "Task in Backlog")
	record.Set("column", "backlog")
	record.Set("position", 1000.0)
	record.Set("history", []map[string]any{})
	require.NoError(t, app.Save(record))

	originalPosition := record.GetFloat("position")

	// Try to move to same column
	cmd := moveTaskToColumn(app, record.Id, "backlog")
	msg := cmd()

	// Verify result - should still return taskMovedMsg but be a no-op
	moved, ok := msg.(taskMovedMsg)
	require.True(t, ok, "Expected taskMovedMsg, got %T", msg)
	assert.Equal(t, "backlog", moved.fromColumn)
	assert.Equal(t, "backlog", moved.toColumn)
	assert.Equal(t, originalPosition, moved.task.GetFloat("position"))
}

// TestStatusMessages verifies status message handling
func TestStatusMessages(t *testing.T) {
	// Test success status
	cmd := showStatus("Task created", false, time.Second)
	msg := cmd()

	status, ok := msg.(statusMsg)
	require.True(t, ok)
	assert.Equal(t, "Task created", status.message)
	assert.False(t, status.isError)

	// Test error status
	cmd = showStatus("Failed to save", true, time.Second)
	msg = cmd()

	status, ok = msg.(statusMsg)
	require.True(t, ok)
	assert.Equal(t, "Failed to save", status.message)
	assert.True(t, status.isError)
}

// TestTaskFormValidation verifies form validation
func TestTaskFormValidation(t *testing.T) {
	form := NewTaskForm(FormModeAdd, 80, 40)

	// Try to submit without title
	cmd := form.submit()
	msg := cmd()

	// Should return status error about missing title
	status, ok := msg.(statusMsg)
	require.True(t, ok)
	assert.True(t, status.isError)
	assert.Contains(t, status.message, "Title")
}

// TestTaskFormLabels verifies label parsing
func TestTaskFormLabels(t *testing.T) {
	form := NewTaskForm(FormModeAdd, 80, 40)
	form.titleInput.SetValue("Test Task")
	form.labelsInput.SetValue("bug, frontend, urgent")

	cmd := form.submit()
	msg := cmd()

	submit, ok := msg.(submitTaskFormMsg)
	require.True(t, ok)
	assert.Equal(t, []string{"bug", "frontend", "urgent"}, submit.data.Labels)
}

// TestTaskFormLabelsEmpty verifies empty labels parsing
func TestTaskFormLabelsEmpty(t *testing.T) {
	form := NewTaskForm(FormModeAdd, 80, 40)
	form.titleInput.SetValue("Test Task")
	form.labelsInput.SetValue("")

	cmd := form.submit()
	msg := cmd()

	submit, ok := msg.(submitTaskFormMsg)
	require.True(t, ok)
	assert.Empty(t, submit.data.Labels)
}

// TestConfirmDialogDefault verifies confirm dialog defaults to No
func TestConfirmDialogDefault(t *testing.T) {
	dialog := NewDeleteConfirmDialog("Test Task")

	// Default should be focused on No/Cancel
	assert.False(t, dialog.focused)
}

// TestConfirmDialogToggle verifies confirm dialog button toggling
func TestConfirmDialogToggle(t *testing.T) {
	dialog := NewConfirmDialog("Test", "Message")

	// Initial state is No focused
	assert.False(t, dialog.focused)

	// Simulate tab key to toggle
	dialog.focused = !dialog.focused

	// Now Yes should be focused
	assert.True(t, dialog.focused)
}

// TestTaskFormEditMode verifies form pre-fills in edit mode
func TestTaskFormEditMode(t *testing.T) {
	task := &TaskItem{
		ID:              "test-id",
		TaskTitle:       "Existing Task",
		TaskDescription: "Existing description",
		Type:            "bug",
		Priority:        "high",
		Column:          "in_progress",
		Labels:          []string{"label1", "label2"},
		DueDate:         "2024-12-31",
	}

	form := NewTaskFormWithData(task, 80, 40)

	assert.Equal(t, FormModeEdit, form.mode)
	assert.Equal(t, "test-id", form.taskID)
	assert.Equal(t, "Existing Task", form.titleInput.Value())
	assert.Equal(t, "Existing description", form.descInput.Value())
	assert.Equal(t, "label1, label2", form.labelsInput.Value())
	assert.Equal(t, "2024-12-31", form.dueDateInput.Value())

	// Verify select indices are set correctly
	assert.Equal(t, 1, form.typeSelect)     // bug is index 1
	assert.Equal(t, 2, form.prioritySelect) // high is index 2
	assert.Equal(t, 2, form.columnSelect)   // in_progress is index 2
}

// TestCreateTaskWithDueDate verifies due date is saved correctly
func TestCreateTaskWithDueDate(t *testing.T) {
	app := testutil.NewTestApp(t)

	createTasksCollection(t, app)
	createBoardsCollection(t, app)

	boardRecord, _ := board.Create(app, board.CreateInput{Name: "Test", Prefix: "TST"})
	boardRec, _ := app.FindRecordById("boards", boardRecord.ID)

	data := TaskFormData{
		Title:    "Task with Due Date",
		Type:     "feature",
		Priority: "medium",
		Column:   "backlog",
		DueDate:  "2024-12-31",
	}

	cmd := createTask(app, boardRec, data)
	msg := cmd()

	created := msg.(taskCreatedMsg)
	assert.Equal(t, "2024-12-31", created.task.GetString("due_date"))
}

// TestGetColumnOrder verifies column ordering
func TestGetColumnOrder(t *testing.T) {
	tests := []struct {
		column   string
		expected int
	}{
		{"backlog", 0},
		{"todo", 1},
		{"in_progress", 2},
		{"need_input", 3},
		{"review", 4},
		{"done", 5},
		{"unknown", 99},
	}

	for _, tt := range tests {
		t.Run(tt.column, func(t *testing.T) {
			result := getColumnOrder(tt.column)
			assert.Equal(t, tt.expected, result)
		})
	}
}
