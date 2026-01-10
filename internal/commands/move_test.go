package commands

import (
	"testing"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ramtinJ95/EgenSkriven/internal/board"
	"github.com/ramtinJ95/EgenSkriven/internal/testutil"
)

// ========== Setup Functions ==========

// setupMoveTestCollections creates both boards and tasks collections for move tests
func setupMoveTestCollections(t *testing.T, app *pocketbase.PocketBase) {
	t.Helper()

	// Create boards collection
	_, err := app.FindCollectionByNameOrId("boards")
	if err != nil {
		boardsCollection := core.NewBaseCollection("boards")
		boardsCollection.Fields.Add(&core.TextField{Name: "name", Required: true})
		boardsCollection.Fields.Add(&core.TextField{Name: "prefix", Required: true})
		boardsCollection.Fields.Add(&core.JSONField{Name: "columns"})
		boardsCollection.Fields.Add(&core.NumberField{Name: "next_seq"})
		boardsCollection.Fields.Add(&core.TextField{Name: "color"})
		require.NoError(t, app.Save(boardsCollection))
	}

	// Create tasks collection with board and seq fields
	_, err = app.FindCollectionByNameOrId("tasks")
	if err != nil {
		tasksCollection := core.NewBaseCollection("tasks")
		tasksCollection.Fields.Add(&core.TextField{Name: "title", Required: true})
		tasksCollection.Fields.Add(&core.TextField{Name: "description"})
		tasksCollection.Fields.Add(&core.SelectField{
			Name:     "type",
			Required: true,
			Values:   []string{"bug", "feature", "chore"},
		})
		tasksCollection.Fields.Add(&core.SelectField{
			Name:     "priority",
			Required: true,
			Values:   []string{"low", "medium", "high", "urgent"},
		})
		tasksCollection.Fields.Add(&core.SelectField{
			Name:     "column",
			Required: true,
			Values:   []string{"backlog", "todo", "in_progress", "need_input", "review", "done"},
		})
		tasksCollection.Fields.Add(&core.NumberField{Name: "position", Required: true})
		tasksCollection.Fields.Add(&core.JSONField{Name: "labels"})
		tasksCollection.Fields.Add(&core.JSONField{Name: "blocked_by"})
		tasksCollection.Fields.Add(&core.SelectField{
			Name:     "created_by",
			Required: true,
			Values:   []string{"user", "agent", "cli"},
		})
		tasksCollection.Fields.Add(&core.TextField{Name: "created_by_agent"})
		tasksCollection.Fields.Add(&core.JSONField{Name: "history"})
		tasksCollection.Fields.Add(&core.TextField{Name: "board"})    // Board reference
		tasksCollection.Fields.Add(&core.NumberField{Name: "seq"})    // Sequence number
		require.NoError(t, app.Save(tasksCollection))
	}
}

// createMoveTestBoard creates a board for testing
func createMoveTestBoard(t *testing.T, app *pocketbase.PocketBase, name, prefix string) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("boards")
	require.NoError(t, err)

	record := core.NewRecord(collection)
	record.Set("name", name)
	record.Set("prefix", prefix)
	record.Set("columns", board.DefaultColumns)
	record.Set("next_seq", 1)
	record.Set("color", "#007bff")

	require.NoError(t, app.Save(record))
	return record
}

// createMoveTestTask creates a task with board and sequence for display ID support
func createMoveTestTask(t *testing.T, app *pocketbase.PocketBase, title, column string, position float64, boardID string, seq int) *core.Record {
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
	record.Set("board", boardID)
	record.Set("seq", seq)

	require.NoError(t, app.Save(record))
	return record
}

// ========== Display ID Resolution Tests ==========

// TestMoveCommand_DisplayIDResolutionInAfterFlag verifies --after supports display IDs
func TestMoveCommand_DisplayIDResolutionInAfterFlag(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupMoveTestCollections(t, app)

	// Create a board with prefix "TST"
	testBoard := createMoveTestBoard(t, app, "Test Board", "TST")

	// Create tasks with sequence numbers
	// task1: TST-1, position 1000
	// task2: TST-2, position 2000
	// task3: TST-3, position 3000 (this will be moved)
	task1 := createMoveTestTask(t, app, "Task One", "todo", 1000.0, testBoard.Id, 1)
	task2 := createMoveTestTask(t, app, "Task Two", "todo", 2000.0, testBoard.Id, 2)
	task3 := createMoveTestTask(t, app, "Task Three", "todo", 3000.0, testBoard.Id, 3)

	// Verify display ID format is correct
	displayID1 := board.FormatDisplayID("TST", 1)
	assert.Equal(t, "TST-1", displayID1)

	// Verify task1 can be found by display ID
	prefix, seq, err := board.ParseDisplayID("TST-1")
	require.NoError(t, err)
	assert.Equal(t, "TST", prefix)
	assert.Equal(t, 1, seq)

	// Verify task lookup by board + seq works
	tasks, err := app.FindAllRecords("tasks")
	require.NoError(t, err)

	var foundTask1 *core.Record
	for _, task := range tasks {
		if task.GetString("board") == testBoard.Id && task.GetInt("seq") == 1 {
			foundTask1 = task
			break
		}
	}
	require.NotNil(t, foundTask1, "should find task by board and sequence")
	assert.Equal(t, task1.Id, foundTask1.Id)

	// Verify GetPositionAfter works with resolved ID
	pos, err := GetPositionAfter(app, task1.Id)
	require.NoError(t, err)
	// Position should be between task1 (1000) and task2 (2000)
	assert.Greater(t, pos, 1000.0)
	assert.Less(t, pos, 2000.0)

	// Verify GetPositionAfter for last task
	pos, err = GetPositionAfter(app, task3.Id)
	require.NoError(t, err)
	// Position should be after task3 (3000)
	assert.Greater(t, pos, 3000.0)

	// Keep task2 reference to avoid unused variable error
	_ = task2
}

// TestMoveCommand_DisplayIDResolutionInBeforeFlag verifies --before supports display IDs
func TestMoveCommand_DisplayIDResolutionInBeforeFlag(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupMoveTestCollections(t, app)

	// Create a board with prefix "WRK"
	workBoard := createMoveTestBoard(t, app, "Work Board", "WRK")

	// Create tasks with sequence numbers
	task1 := createMoveTestTask(t, app, "Work Task One", "in_progress", 1000.0, workBoard.Id, 1)
	task2 := createMoveTestTask(t, app, "Work Task Two", "in_progress", 2000.0, workBoard.Id, 2)
	task3 := createMoveTestTask(t, app, "Work Task Three", "in_progress", 3000.0, workBoard.Id, 3)

	// Verify display ID parsing
	prefix, seq, err := board.ParseDisplayID("WRK-2")
	require.NoError(t, err)
	assert.Equal(t, "WRK", prefix)
	assert.Equal(t, 2, seq)

	// Verify GetPositionBefore works with resolved ID
	pos, err := GetPositionBefore(app, task2.Id)
	require.NoError(t, err)
	// Position should be between task1 (1000) and task2 (2000)
	assert.Greater(t, pos, 1000.0)
	assert.Less(t, pos, 2000.0)

	// Verify GetPositionBefore for first task
	pos, err = GetPositionBefore(app, task1.Id)
	require.NoError(t, err)
	// Position should be before task1 (1000), so less than 1000
	assert.Less(t, pos, 1000.0)
	assert.Greater(t, pos, 0.0)

	// Keep task3 reference to avoid unused variable error
	_ = task3
}

// TestMoveCommand_DisplayIDWithDifferentBoards verifies display IDs are board-scoped
func TestMoveCommand_DisplayIDWithDifferentBoards(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupMoveTestCollections(t, app)

	// Create two boards
	board1 := createMoveTestBoard(t, app, "Board One", "ONE")
	board2 := createMoveTestBoard(t, app, "Board Two", "TWO")

	// Create tasks with same sequence number but different boards
	taskOne1 := createMoveTestTask(t, app, "One Task 1", "todo", 1000.0, board1.Id, 1)
	taskTwo1 := createMoveTestTask(t, app, "Two Task 1", "todo", 1000.0, board2.Id, 1)

	// Both have seq=1 but different boards
	assert.Equal(t, 1, taskOne1.GetInt("seq"))
	assert.Equal(t, 1, taskTwo1.GetInt("seq"))
	assert.NotEqual(t, taskOne1.GetString("board"), taskTwo1.GetString("board"))

	// Display IDs should be different
	displayID1 := board.FormatDisplayID("ONE", 1)
	displayID2 := board.FormatDisplayID("TWO", 1)
	assert.Equal(t, "ONE-1", displayID1)
	assert.Equal(t, "TWO-1", displayID2)
	assert.NotEqual(t, displayID1, displayID2)
}

// TestMoveCommand_PositionCalculation verifies position calculations are correct
func TestMoveCommand_PositionCalculation(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupMoveTestCollections(t, app)

	board := createMoveTestBoard(t, app, "Position Test", "POS")

	// Create tasks at specific positions
	task1 := createMoveTestTask(t, app, "Task 1", "todo", 1000.0, board.Id, 1)
	task2 := createMoveTestTask(t, app, "Task 2", "todo", 2000.0, board.Id, 2)
	task3 := createMoveTestTask(t, app, "Task 3", "todo", 3000.0, board.Id, 3)

	// Test GetPositionAfter for middle task
	posAfter1, err := GetPositionAfter(app, task1.Id)
	require.NoError(t, err)
	assert.Equal(t, 1500.0, posAfter1, "position after task1 should be midpoint to task2")

	// Test GetPositionBefore for middle task
	posBefore3, err := GetPositionBefore(app, task3.Id)
	require.NoError(t, err)
	assert.Equal(t, 2500.0, posBefore3, "position before task3 should be midpoint from task2")

	// Test GetPositionAfter for last task
	posAfterLast, err := GetPositionAfter(app, task3.Id)
	require.NoError(t, err)
	assert.Equal(t, 4000.0, posAfterLast, "position after last should add default gap")

	// Test GetPositionBefore for first task
	posBeforeFirst, err := GetPositionBefore(app, task1.Id)
	require.NoError(t, err)
	assert.Equal(t, 500.0, posBeforeFirst, "position before first should halve the position")

	// Keep task2 reference
	_ = task2
}
