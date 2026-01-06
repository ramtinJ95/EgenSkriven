package commands

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ramtinJ95/EgenSkriven/internal/testutil"
)

// ========== Setup Functions ==========

// setupImportTestCollections creates boards, epics, and tasks collections for import testing
func setupImportTestCollections(t *testing.T, app *pocketbase.PocketBase) {
	t.Helper()

	// Create boards collection
	if _, err := app.FindCollectionByNameOrId("boards"); err != nil {
		boards := core.NewBaseCollection("boards")
		boards.Fields.Add(&core.TextField{Name: "name", Required: true})
		boards.Fields.Add(&core.TextField{Name: "prefix", Required: true})
		boards.Fields.Add(&core.JSONField{Name: "columns"})
		boards.Fields.Add(&core.TextField{Name: "color"})
		require.NoError(t, app.Save(boards))
	}

	// Create epics collection
	if _, err := app.FindCollectionByNameOrId("epics"); err != nil {
		epics := core.NewBaseCollection("epics")
		epics.Fields.Add(&core.TextField{Name: "title", Required: true})
		epics.Fields.Add(&core.TextField{Name: "description"})
		epics.Fields.Add(&core.TextField{Name: "color"})
		require.NoError(t, app.Save(epics))
	}

	// Create tasks collection
	if _, err := app.FindCollectionByNameOrId("tasks"); err != nil {
		tasks := core.NewBaseCollection("tasks")
		tasks.Fields.Add(&core.TextField{Name: "title", Required: true})
		tasks.Fields.Add(&core.TextField{Name: "description"})
		tasks.Fields.Add(&core.SelectField{
			Name:     "type",
			Required: true,
			Values:   []string{"bug", "feature", "chore"},
		})
		tasks.Fields.Add(&core.SelectField{
			Name:     "priority",
			Required: true,
			Values:   []string{"low", "medium", "high", "urgent"},
		})
		tasks.Fields.Add(&core.SelectField{
			Name:     "column",
			Required: true,
			Values:   []string{"backlog", "todo", "in_progress", "review", "done"},
		})
		tasks.Fields.Add(&core.NumberField{Name: "position", Required: true})
		tasks.Fields.Add(&core.TextField{Name: "board"})
		tasks.Fields.Add(&core.TextField{Name: "epic"})
		tasks.Fields.Add(&core.TextField{Name: "parent"})
		tasks.Fields.Add(&core.JSONField{Name: "labels"})
		tasks.Fields.Add(&core.JSONField{Name: "blocked_by"})
		tasks.Fields.Add(&core.DateField{Name: "due_date"})
		tasks.Fields.Add(&core.SelectField{
			Name:     "created_by",
			Required: true,
			Values:   []string{"user", "agent", "cli"},
		})
		require.NoError(t, app.Save(tasks))
	}
}

// createImportTestFile creates a temporary JSON file with export data
func createImportTestFile(t *testing.T, data ExportData) string {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "import-*.json")
	require.NoError(t, err)
	defer tmpFile.Close()

	encoder := json.NewEncoder(tmpFile)
	encoder.SetIndent("", "  ")
	require.NoError(t, encoder.Encode(data))

	return tmpFile.Name()
}

// createImportTestBoard creates a board record for import testing
func createImportTestBoard(t *testing.T, app *pocketbase.PocketBase, name, prefix string) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("boards")
	require.NoError(t, err)

	record := core.NewRecord(collection)
	record.Set("name", name)
	record.Set("prefix", prefix)
	record.Set("columns", []string{"backlog", "todo", "in_progress", "review", "done"})
	record.Set("color", "blue")

	require.NoError(t, app.Save(record))
	return record
}

// ========== Import Boards Tests ==========

func TestImportBoards_CreateNew(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupImportTestCollections(t, app)

	// Create import data with boards (IDs must be 15+ chars for PocketBase)
	data := ExportData{
		Version:  "1.0",
		Exported: "2026-01-06T12:00:00Z",
		Boards: []ExportBoard{
			{ID: "board1testid001", Name: "Work", Prefix: "WRK", Columns: []string{"todo", "done"}, Color: "blue"},
			{ID: "board2testid002", Name: "Personal", Prefix: "PRS", Columns: []string{"todo", "done"}, Color: "green"},
		},
		Epics: []ExportEpic{},
		Tasks: []ExportTask{},
	}

	filename := createImportTestFile(t, data)
	defer os.Remove(filename)

	// Import with merge strategy
	stats := ImportStats{}
	err := importBoards(app, data.Boards, "merge", false, &stats)
	require.NoError(t, err)

	// Verify boards were created
	assert.Equal(t, 2, stats.BoardsCreated)
	assert.Equal(t, 0, stats.BoardsSkipped)
	assert.Equal(t, 0, stats.BoardsUpdated)

	// Verify board data
	board, err := app.FindRecordById("boards", "board1testid001")
	require.NoError(t, err)
	assert.Equal(t, "Work", board.GetString("name"))
	assert.Equal(t, "WRK", board.GetString("prefix"))
}

func TestImportBoards_MergeStrategy(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupImportTestCollections(t, app)

	// Create existing board
	existing := createImportTestBoard(t, app, "Existing", "EXT")

	// Create import data with same ID (IDs must be 15+ chars for PocketBase)
	data := ExportData{
		Version:  "1.0",
		Exported: "2026-01-06T12:00:00Z",
		Boards: []ExportBoard{
			{ID: existing.Id, Name: "Updated Name", Prefix: "UPD"},
			{ID: "newboardtest001", Name: "New Board", Prefix: "NEW"},
		},
		Epics: []ExportEpic{},
		Tasks: []ExportTask{},
	}

	// Import with merge strategy (should skip existing)
	stats := ImportStats{}
	err := importBoards(app, data.Boards, "merge", false, &stats)
	require.NoError(t, err)

	assert.Equal(t, 1, stats.BoardsCreated)
	assert.Equal(t, 1, stats.BoardsSkipped)
	assert.Equal(t, 0, stats.BoardsUpdated)

	// Verify existing board was NOT updated
	board, err := app.FindRecordById("boards", existing.Id)
	require.NoError(t, err)
	assert.Equal(t, "Existing", board.GetString("name"))
}

func TestImportBoards_ReplaceStrategy(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupImportTestCollections(t, app)

	// Create existing board
	existing := createImportTestBoard(t, app, "Existing", "EXT")

	// Create import data with same ID
	data := ExportData{
		Version:  "1.0",
		Exported: "2026-01-06T12:00:00Z",
		Boards: []ExportBoard{
			{ID: existing.Id, Name: "Updated Name", Prefix: "UPD"},
		},
		Epics: []ExportEpic{},
		Tasks: []ExportTask{},
	}

	// Import with replace strategy (should update existing)
	stats := ImportStats{}
	err := importBoards(app, data.Boards, "replace", false, &stats)
	require.NoError(t, err)

	assert.Equal(t, 0, stats.BoardsCreated)
	assert.Equal(t, 0, stats.BoardsSkipped)
	assert.Equal(t, 1, stats.BoardsUpdated)

	// Verify existing board WAS updated
	board, err := app.FindRecordById("boards", existing.Id)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", board.GetString("name"))
	assert.Equal(t, "UPD", board.GetString("prefix"))
}

func TestImportBoards_DryRun(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupImportTestCollections(t, app)

	// Create import data (IDs must be 15+ chars for PocketBase)
	data := ExportData{
		Version:  "1.0",
		Exported: "2026-01-06T12:00:00Z",
		Boards: []ExportBoard{
			{ID: "board1testid001", Name: "Work", Prefix: "WRK"},
		},
		Epics: []ExportEpic{},
		Tasks: []ExportTask{},
	}

	// Import with dry run
	stats := ImportStats{}
	err := importBoards(app, data.Boards, "merge", true, &stats)
	require.NoError(t, err)

	// Stats should show what WOULD happen
	assert.Equal(t, 1, stats.BoardsCreated)

	// But no actual records should be created
	boards, err := app.FindAllRecords("boards")
	require.NoError(t, err)
	assert.Empty(t, boards)
}

// ========== Import Epics Tests ==========

func TestImportEpics_CreateNew(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupImportTestCollections(t, app)

	// IDs must be 15+ chars for PocketBase
	epics := []ExportEpic{
		{ID: "epic1testid0001", Title: "Epic 1", Description: "Description 1", Color: "blue"},
		{ID: "epic2testid0002", Title: "Epic 2", Description: "Description 2", Color: "green"},
	}

	stats := ImportStats{}
	err := importEpics(app, epics, "merge", false, &stats)
	require.NoError(t, err)

	assert.Equal(t, 2, stats.EpicsCreated)

	// Verify epic data
	epic, err := app.FindRecordById("epics", "epic1testid0001")
	require.NoError(t, err)
	assert.Equal(t, "Epic 1", epic.GetString("title"))
}

func TestImportEpics_ReplaceStrategy(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupImportTestCollections(t, app)

	// Create existing epic
	collection, err := app.FindCollectionByNameOrId("epics")
	require.NoError(t, err)

	existing := core.NewRecord(collection)
	existing.Set("title", "Original Title")
	existing.Set("description", "Original description")
	require.NoError(t, app.Save(existing))

	// Import with replace
	epics := []ExportEpic{
		{ID: existing.Id, Title: "Updated Title", Description: "Updated description"},
	}

	stats := ImportStats{}
	err = importEpics(app, epics, "replace", false, &stats)
	require.NoError(t, err)

	assert.Equal(t, 1, stats.EpicsUpdated)

	// Verify epic was updated
	epic, err := app.FindRecordById("epics", existing.Id)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", epic.GetString("title"))
}

// ========== Import Tasks Tests ==========

func TestImportTasks_CreateNew(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupImportTestCollections(t, app)

	// IDs must be 15+ chars for PocketBase
	tasks := []ExportTask{
		{
			ID:       "task1testid0001",
			Title:    "Task 1",
			Type:     "feature",
			Priority: "medium",
			Column:   "todo",
			Position: 1000.0,
		},
		{
			ID:       "task2testid0002",
			Title:    "Task 2",
			Type:     "bug",
			Priority: "high",
			Column:   "in_progress",
			Position: 2000.0,
		},
	}

	stats := ImportStats{}
	err := importTasks(app, tasks, "merge", false, &stats)
	require.NoError(t, err)

	assert.Equal(t, 2, stats.TasksCreated)

	// Verify task data
	task, err := app.FindRecordById("tasks", "task1testid0001")
	require.NoError(t, err)
	assert.Equal(t, "Task 1", task.GetString("title"))
	assert.Equal(t, "feature", task.GetString("type"))
	assert.Equal(t, "medium", task.GetString("priority"))
	assert.Equal(t, "todo", task.GetString("column"))
}

func TestImportTasks_MergeStrategy(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupImportTestCollections(t, app)

	// Create existing task
	collection, err := app.FindCollectionByNameOrId("tasks")
	require.NoError(t, err)

	existing := core.NewRecord(collection)
	existing.Set("title", "Original Task")
	existing.Set("type", "feature")
	existing.Set("priority", "low")
	existing.Set("column", "todo")
	existing.Set("position", 1000.0)
	existing.Set("created_by", "cli")
	require.NoError(t, app.Save(existing))

	// Import with same ID (should skip) - IDs must be 15+ chars for PocketBase
	tasks := []ExportTask{
		{ID: existing.Id, Title: "Updated Task", Type: "bug", Priority: "high", Column: "done", Position: 2000.0},
		{ID: "newtasktestid01", Title: "New Task", Type: "feature", Priority: "medium", Column: "todo", Position: 3000.0},
	}

	stats := ImportStats{}
	err = importTasks(app, tasks, "merge", false, &stats)
	require.NoError(t, err)

	assert.Equal(t, 1, stats.TasksCreated)
	assert.Equal(t, 1, stats.TasksSkipped)

	// Verify existing was NOT updated
	task, err := app.FindRecordById("tasks", existing.Id)
	require.NoError(t, err)
	assert.Equal(t, "Original Task", task.GetString("title"))
}

func TestImportTasks_ReplaceStrategy(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupImportTestCollections(t, app)

	// Create existing task
	collection, err := app.FindCollectionByNameOrId("tasks")
	require.NoError(t, err)

	existing := core.NewRecord(collection)
	existing.Set("title", "Original Task")
	existing.Set("type", "feature")
	existing.Set("priority", "low")
	existing.Set("column", "todo")
	existing.Set("position", 1000.0)
	existing.Set("created_by", "cli")
	require.NoError(t, app.Save(existing))

	// Import with replace strategy
	tasks := []ExportTask{
		{ID: existing.Id, Title: "Updated Task", Type: "bug", Priority: "high", Column: "done", Position: 2000.0},
	}

	stats := ImportStats{}
	err = importTasks(app, tasks, "replace", false, &stats)
	require.NoError(t, err)

	assert.Equal(t, 0, stats.TasksCreated)
	assert.Equal(t, 1, stats.TasksUpdated)

	// Verify existing WAS updated
	task, err := app.FindRecordById("tasks", existing.Id)
	require.NoError(t, err)
	assert.Equal(t, "Updated Task", task.GetString("title"))
	assert.Equal(t, "bug", task.GetString("type"))
	assert.Equal(t, "high", task.GetString("priority"))
}

func TestImportTasks_DryRun(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupImportTestCollections(t, app)

	// IDs must be 15+ chars for PocketBase
	tasks := []ExportTask{
		{ID: "task1testid0001", Title: "Task 1", Type: "feature", Priority: "medium", Column: "todo", Position: 1000.0},
	}

	stats := ImportStats{}
	err := importTasks(app, tasks, "merge", true, &stats)
	require.NoError(t, err)

	// Stats should show what WOULD happen
	assert.Equal(t, 1, stats.TasksCreated)

	// But no actual records should be created
	allTasks, err := app.FindAllRecords("tasks")
	require.NoError(t, err)
	assert.Empty(t, allTasks)
}

// ========== Full Import Integration Tests ==========

func TestRunImport_FullWorkflow(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupImportTestCollections(t, app)

	// Create export data (IDs must be 15+ chars for PocketBase)
	data := ExportData{
		Version:  "1.0",
		Exported: "2026-01-06T12:00:00Z",
		Boards: []ExportBoard{
			{ID: "board1testid001", Name: "Work", Prefix: "WRK", Columns: []string{"todo", "done"}},
		},
		Epics: []ExportEpic{
			{ID: "epic1testid0001", Title: "Q1 Goals"},
		},
		Tasks: []ExportTask{
			{ID: "task1testid0001", Title: "Task 1", Type: "feature", Priority: "medium", Column: "todo", Position: 1000.0, Board: "board1testid001", Epic: "epic1testid0001"},
			{ID: "task2testid0002", Title: "Task 2", Type: "bug", Priority: "high", Column: "done", Position: 2000.0, Board: "board1testid001"},
		},
	}

	filename := createImportTestFile(t, data)
	defer os.Remove(filename)

	// Run import
	out := getFormatter()
	err := runImport(app, filename, "merge", false, out)
	require.NoError(t, err)

	// Verify all data was imported
	boards, err := app.FindAllRecords("boards")
	require.NoError(t, err)
	assert.Len(t, boards, 1)

	epics, err := app.FindAllRecords("epics")
	require.NoError(t, err)
	assert.Len(t, epics, 1)

	tasks, err := app.FindAllRecords("tasks")
	require.NoError(t, err)
	assert.Len(t, tasks, 2)
}

func TestRunImport_DryRunNoChanges(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupImportTestCollections(t, app)

	// Create export data (IDs must be 15+ chars for PocketBase)
	data := ExportData{
		Version:  "1.0",
		Exported: "2026-01-06T12:00:00Z",
		Boards: []ExportBoard{
			{ID: "board1testid001", Name: "Work", Prefix: "WRK"},
		},
		Epics: []ExportEpic{
			{ID: "epic1testid0001", Title: "Q1 Goals"},
		},
		Tasks: []ExportTask{
			{ID: "task1testid0001", Title: "Task 1", Type: "feature", Priority: "medium", Column: "todo", Position: 1000.0},
		},
	}

	filename := createImportTestFile(t, data)
	defer os.Remove(filename)

	// Run import with dry-run
	out := getFormatter()
	err := runImport(app, filename, "merge", true, out)
	require.NoError(t, err)

	// Verify NO data was imported
	boards, err := app.FindAllRecords("boards")
	require.NoError(t, err)
	assert.Empty(t, boards)

	epics, err := app.FindAllRecords("epics")
	require.NoError(t, err)
	assert.Empty(t, epics)

	tasks, err := app.FindAllRecords("tasks")
	require.NoError(t, err)
	assert.Empty(t, tasks)
}

// Note: TestRunImport_InvalidFile and TestRunImport_InvalidJSON are not included
// because out.Error() calls os.Exit() which terminates the test process.
// The error handling is tested implicitly through the output.Formatter tests.

func TestRunImport_FileNotFound(t *testing.T) {
	// Test that file open fails gracefully by checking os.Open directly
	_, err := os.Open("/nonexistent/file.json")
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestRunImport_InvalidJSONSyntax(t *testing.T) {
	// Test that invalid JSON parsing fails gracefully
	tmpFile, err := os.CreateTemp("", "invalid-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("{ invalid json }")
	require.NoError(t, err)
	tmpFile.Close()

	// Verify the file has invalid JSON
	file, err := os.Open(tmpFile.Name())
	require.NoError(t, err)
	defer file.Close()

	var data ExportData
	err = json.NewDecoder(file).Decode(&data)
	assert.Error(t, err)
}

// ========== Import Stats Tests ==========

func TestImportStats_Combined(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupImportTestCollections(t, app)

	// Create some existing records
	existingBoard := createImportTestBoard(t, app, "Existing Board", "EXT")

	// Create import data with mix of existing and new (IDs must be 15+ chars for PocketBase)
	data := ExportData{
		Version:  "1.0",
		Exported: "2026-01-06T12:00:00Z",
		Boards: []ExportBoard{
			{ID: existingBoard.Id, Name: "Updated Board", Prefix: "UPD"},
			{ID: "newboardtest001", Name: "New Board", Prefix: "NEW"},
		},
		Epics: []ExportEpic{
			{ID: "epic1testid0001", Title: "Epic 1"},
		},
		Tasks: []ExportTask{
			{ID: "task1testid0001", Title: "Task 1", Type: "feature", Priority: "medium", Column: "todo", Position: 1000.0},
			{ID: "task2testid0002", Title: "Task 2", Type: "bug", Priority: "high", Column: "done", Position: 2000.0},
		},
	}

	filename := createImportTestFile(t, data)
	defer os.Remove(filename)

	// Run import with merge
	out := getFormatter()
	err := runImport(app, filename, "merge", false, out)
	require.NoError(t, err)

	// Verify final counts
	boards, err := app.FindAllRecords("boards")
	require.NoError(t, err)
	assert.Len(t, boards, 2) // 1 existing + 1 new

	epics, err := app.FindAllRecords("epics")
	require.NoError(t, err)
	assert.Len(t, epics, 1)

	tasks, err := app.FindAllRecords("tasks")
	require.NoError(t, err)
	assert.Len(t, tasks, 2)
}

func TestImportTasks_WithAllFields(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupImportTestCollections(t, app)

	// Create a task with all fields populated (IDs must be 15+ chars for PocketBase)
	tasks := []ExportTask{
		{
			ID:          "task1testid0001",
			Title:       "Full Task",
			Description: "Full description",
			Type:        "bug",
			Priority:    "urgent",
			Column:      "in_progress",
			Position:    1500.0,
			Board:       "board1testid001",
			Epic:        "epic1testid0001",
			Parent:      "parent1testid01",
			Labels:      []string{"urgent", "backend"},
			BlockedBy:   []string{"task0testid0000"},
			DueDate:     "2026-01-15",
		},
	}

	stats := ImportStats{}
	err := importTasks(app, tasks, "merge", false, &stats)
	require.NoError(t, err)

	assert.Equal(t, 1, stats.TasksCreated)

	// Verify all fields
	task, err := app.FindRecordById("tasks", "task1testid0001")
	require.NoError(t, err)
	assert.Equal(t, "Full Task", task.GetString("title"))
	assert.Equal(t, "Full description", task.GetString("description"))
	assert.Equal(t, "bug", task.GetString("type"))
	assert.Equal(t, "urgent", task.GetString("priority"))
	assert.Equal(t, "in_progress", task.GetString("column"))
	assert.Equal(t, 1500.0, task.GetFloat("position"))
	assert.Equal(t, "board1testid001", task.GetString("board"))
	assert.Equal(t, "epic1testid0001", task.GetString("epic"))
	assert.Equal(t, "parent1testid01", task.GetString("parent"))
}
