package board

import (
	"testing"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ramtinJ95/EgenSkriven/internal/testutil"
)

// Test Create function

func TestCreate_Success(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)

	b, err := Create(app, CreateInput{
		Name:   "Work",
		Prefix: "WRK",
		Color:  "#3B82F6",
	})

	require.NoError(t, err)
	assert.Equal(t, "Work", b.Name)
	assert.Equal(t, "WRK", b.Prefix)
	assert.Equal(t, "#3B82F6", b.Color)
	assert.Equal(t, DefaultColumns, b.Columns)
	assert.NotEmpty(t, b.ID)
}

func TestCreate_PrefixUppercased(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)

	b, err := Create(app, CreateInput{
		Name:   "Work",
		Prefix: "wrk", // lowercase input
	})

	require.NoError(t, err)
	assert.Equal(t, "WRK", b.Prefix) // should be uppercased
}

func TestCreate_CustomColumns(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)

	customColumns := []string{"idea", "doing", "done"}
	b, err := Create(app, CreateInput{
		Name:    "Simple Board",
		Prefix:  "SIM",
		Columns: customColumns,
	})

	require.NoError(t, err)
	assert.Equal(t, customColumns, b.Columns)
}

func TestCreate_DuplicatePrefixFails(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)

	// Create first board
	_, err := Create(app, CreateInput{Name: "Work", Prefix: "WRK"})
	require.NoError(t, err)

	// Try to create second board with same prefix
	_, err = Create(app, CreateInput{Name: "Work 2", Prefix: "WRK"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already in use")
}

func TestCreate_InvalidPrefixFails(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)

	tests := []struct {
		name        string
		prefix      string
		expectError string
	}{
		{"empty", "", "prefix is required"},
		{"too long", "VERYLONGPREFIX", "prefix must be 10 characters or less"},
		{"special chars", "WRK@#", "prefix must be alphanumeric"},
		{"spaces", "WR K", "prefix must be alphanumeric"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Create(app, CreateInput{
				Name:   "Test",
				Prefix: tc.prefix,
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectError)
		})
	}
}

func TestCreate_EmptyNameFails(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)

	_, err := Create(app, CreateInput{
		Name:   "",
		Prefix: "WRK",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

// Test GetByNameOrPrefix function

func TestGetByNameOrPrefix_ByExactPrefix(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)

	created, _ := Create(app, CreateInput{Name: "Work", Prefix: "WRK"})

	found, err := GetByNameOrPrefix(app, "WRK")
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.Id)
}

func TestGetByNameOrPrefix_ByLowercasePrefix(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)

	created, _ := Create(app, CreateInput{Name: "Work", Prefix: "WRK"})

	found, err := GetByNameOrPrefix(app, "wrk")
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.Id)
}

func TestGetByNameOrPrefix_ByExactName(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)

	created, _ := Create(app, CreateInput{Name: "My Work Board", Prefix: "WRK"})

	found, err := GetByNameOrPrefix(app, "My Work Board")
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.Id)
}

func TestGetByNameOrPrefix_ByLowercaseName(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)

	created, _ := Create(app, CreateInput{Name: "Work", Prefix: "WRK"})

	found, err := GetByNameOrPrefix(app, "work")
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.Id)
}

func TestGetByNameOrPrefix_ByPartialName(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)

	created, _ := Create(app, CreateInput{Name: "My Work Board", Prefix: "WRK"})

	found, err := GetByNameOrPrefix(app, "work")
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.Id)
}

func TestGetByNameOrPrefix_ByID(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)

	created, _ := Create(app, CreateInput{Name: "Work", Prefix: "WRK"})

	found, err := GetByNameOrPrefix(app, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.Id)
}

func TestGetByNameOrPrefix_AmbiguousFails(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)

	Create(app, CreateInput{Name: "Work Tasks", Prefix: "WRK"})
	Create(app, CreateInput{Name: "Work Projects", Prefix: "PRJ"})

	_, err := GetByNameOrPrefix(app, "work")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambiguous")
}

func TestGetByNameOrPrefix_NotFound(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)

	_, err := GetByNameOrPrefix(app, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "board not found")
}

func TestGetByNameOrPrefix_EmptyRefFails(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)

	_, err := GetByNameOrPrefix(app, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "board reference is required")
}

// Test GetAll function

func TestGetAll_Empty(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)

	boards, err := GetAll(app)
	require.NoError(t, err)
	assert.Len(t, boards, 0)
}

func TestGetAll_MultipleBoards(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)

	Create(app, CreateInput{Name: "Work", Prefix: "WRK"})
	Create(app, CreateInput{Name: "Personal", Prefix: "PER"})
	Create(app, CreateInput{Name: "Side Project", Prefix: "SIDE"})

	boards, err := GetAll(app)
	require.NoError(t, err)
	assert.Len(t, boards, 3)
}

// Test GetNextSequence function

func TestGetNextSequence_FirstTask(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)
	setupTasksCollection(t, app)

	b, _ := Create(app, CreateInput{Name: "Work", Prefix: "WRK"})

	seq, err := GetNextSequence(app, b.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, seq)
}

func TestGetNextSequence_WithExistingTasks(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)
	setupTasksCollection(t, app)

	b, _ := Create(app, CreateInput{Name: "Work", Prefix: "WRK"})

	// Call GetNextSequence 3 times to simulate creating 3 tasks
	// This increments the counter each time
	for i := 1; i <= 3; i++ {
		seq, err := GetNextSequence(app, b.ID)
		require.NoError(t, err)
		assert.Equal(t, i, seq) // Should get 1, 2, 3
	}

	// Next call should return 4
	seq, err := GetNextSequence(app, b.ID)
	require.NoError(t, err)
	assert.Equal(t, 4, seq)
}

func TestGetNextSequence_IsolatedPerBoard(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)
	setupTasksCollection(t, app)

	b1, _ := Create(app, CreateInput{Name: "Work", Prefix: "WRK"})
	b2, _ := Create(app, CreateInput{Name: "Personal", Prefix: "PER"})

	// Call GetNextSequence 5 times on board 1
	for i := 1; i <= 5; i++ {
		_, err := GetNextSequence(app, b1.ID)
		require.NoError(t, err)
	}

	// Board 2 should still start at 1 (counters are isolated per board)
	seq, err := GetNextSequence(app, b2.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, seq)
}

// Test GetAndIncrementSequence function (atomic sequence generation)

func TestGetAndIncrementSequence_IncrementsNextSeq(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)
	setupTasksCollection(t, app)

	b, _ := Create(app, CreateInput{Name: "Work", Prefix: "WRK"})

	// First call should return 1 and increment to 2
	seq1, err := GetAndIncrementSequence(app, b.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, seq1)

	// Second call should return 2 and increment to 3
	seq2, err := GetAndIncrementSequence(app, b.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, seq2)

	// Third call should return 3
	seq3, err := GetAndIncrementSequence(app, b.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, seq3)
}

func TestGetAndIncrementSequence_BoardNotFound(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)

	_, err := GetAndIncrementSequence(app, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "board not found")
}

func TestCreate_InitializesNextSeq(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)

	b, err := Create(app, CreateInput{Name: "Work", Prefix: "WRK"})
	require.NoError(t, err)

	// Check that next_seq is initialized to 1
	record, err := app.FindRecordById("boards", b.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, record.GetInt("next_seq"))
}

// Test FormatDisplayID function

func TestFormatDisplayID(t *testing.T) {
	tests := []struct {
		prefix   string
		seq      int
		expected string
	}{
		{"WRK", 1, "WRK-1"},
		{"WRK", 123, "WRK-123"},
		{"PER", 9999, "PER-9999"},
		{"A", 1, "A-1"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			result := FormatDisplayID(tc.prefix, tc.seq)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// Test ParseDisplayID function

func TestParseDisplayID_Valid(t *testing.T) {
	tests := []struct {
		input      string
		wantPrefix string
		wantSeq    int
	}{
		{"WRK-123", "WRK", 123},
		{"PER-1", "PER", 1},
		{"wrk-456", "WRK", 456}, // lowercase prefix should be uppercased
		{"A-9999", "A", 9999},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			prefix, seq, err := ParseDisplayID(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.wantPrefix, prefix)
			assert.Equal(t, tc.wantSeq, seq)
		})
	}
}

func TestParseDisplayID_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"no dash", "WRK123"},
		{"no sequence", "WRK-"},
		{"non-numeric sequence", "WRK-abc"},
		{"empty", ""},
		{"just dash", "-"},
		{"only prefix", "WRK"},
		{"negative sequence", "WRK--5"},
		{"zero sequence", "WRK-0"},
		{"empty prefix", "-123"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := ParseDisplayID(tc.input)
			require.Error(t, err)
		})
	}
}

// Test RecordToBoard function

func TestRecordToBoard(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)

	created, _ := Create(app, CreateInput{
		Name:   "Work",
		Prefix: "WRK",
		Color:  "#3B82F6",
	})

	record, _ := app.FindRecordById("boards", created.ID)
	b := RecordToBoard(record)

	assert.Equal(t, created.ID, b.ID)
	assert.Equal(t, "Work", b.Name)
	assert.Equal(t, "WRK", b.Prefix)
	assert.Equal(t, "#3B82F6", b.Color)
	assert.Equal(t, DefaultColumns, b.Columns)
}

// Test Delete function

func TestDelete_BoardOnly(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)

	b, _ := Create(app, CreateInput{Name: "Work", Prefix: "WRK"})

	err := Delete(app, b.ID, true)
	require.NoError(t, err)

	// Board should be gone
	_, err = app.FindRecordById("boards", b.ID)
	assert.Error(t, err)
}

func TestDelete_WithTasks_DeleteTasks(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)
	setupTasksCollection(t, app)

	b, _ := Create(app, CreateInput{Name: "Work", Prefix: "WRK"})

	// Create tasks
	task1 := createTestTask(t, app, "Task 1", b.ID, 1)
	task2 := createTestTask(t, app, "Task 2", b.ID, 2)

	err := Delete(app, b.ID, true)
	require.NoError(t, err)

	// Board should be gone
	_, err = app.FindRecordById("boards", b.ID)
	assert.Error(t, err)

	// Tasks should be gone
	_, err = app.FindRecordById("tasks", task1.Id)
	assert.Error(t, err)
	_, err = app.FindRecordById("tasks", task2.Id)
	assert.Error(t, err)
}

func TestDelete_WithTasks_OrphanTasks(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)
	setupTasksCollection(t, app)

	b, _ := Create(app, CreateInput{Name: "Work", Prefix: "WRK"})

	// Create task
	task := createTestTask(t, app, "Task 1", b.ID, 1)

	err := Delete(app, b.ID, false)
	require.NoError(t, err)

	// Board should be gone
	_, err = app.FindRecordById("boards", b.ID)
	assert.Error(t, err)

	// Task should still exist but with empty board
	orphanedTask, err := app.FindRecordById("tasks", task.Id)
	require.NoError(t, err)
	assert.Empty(t, orphanedTask.GetString("board"))
}

func TestDelete_NotFound(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupBoardsCollection(t, app)

	err := Delete(app, "nonexistent", true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "board not found")
}

// Test isAlphanumeric (internal function via Create)

func TestIsAlphanumeric(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"ABC", true},
		{"abc", true},
		{"ABC123", true},
		{"123", true},
		{"", true}, // empty is technically alphanumeric
		{"ABC DEF", false},
		{"ABC-DEF", false},
		{"ABC_DEF", false},
		{"ABC@DEF", false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.valid, isAlphanumeric(tc.input))
		})
	}
}

// Helper functions

func setupBoardsCollection(t *testing.T, app *pocketbase.PocketBase) {
	t.Helper()

	// Check if collection exists
	_, err := app.FindCollectionByNameOrId("boards")
	if err == nil {
		return // Collection already exists
	}

	// Create boards collection
	collection := core.NewBaseCollection("boards")

	collection.Fields.Add(&core.TextField{
		Name:     "name",
		Required: true,
		Min:      1,
		Max:      100,
	})
	collection.Fields.Add(&core.TextField{
		Name:     "prefix",
		Required: true,
		Min:      1,
		Max:      10,
	})
	collection.Fields.Add(&core.JSONField{
		Name:    "columns",
		MaxSize: 10000,
	})
	collection.Fields.Add(&core.TextField{
		Name: "color",
		Max:  7,
	})
	collection.Fields.Add(&core.NumberField{
		Name: "next_seq",
	})

	collection.Indexes = []string{
		"CREATE UNIQUE INDEX idx_boards_prefix ON boards(prefix)",
	}

	if err := app.Save(collection); err != nil {
		t.Fatalf("failed to create boards collection: %v", err)
	}
}

func setupTasksCollection(t *testing.T, app *pocketbase.PocketBase) {
	t.Helper()

	// Check if collection exists
	_, err := app.FindCollectionByNameOrId("tasks")
	if err == nil {
		return // Collection already exists
	}

	// Create tasks collection with board relation
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
	collection.Fields.Add(&core.TextField{Name: "board"}) // Board relation as text for simplicity
	collection.Fields.Add(&core.NumberField{Name: "seq"})

	if err := app.Save(collection); err != nil {
		t.Fatalf("failed to create tasks collection: %v", err)
	}
}

func createTestTask(t *testing.T, app *pocketbase.PocketBase, title, boardID string, seq int) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("tasks")
	if err != nil {
		t.Fatalf("tasks collection not found: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("title", title)
	record.Set("type", "feature")
	record.Set("priority", "medium")
	record.Set("column", "backlog")
	record.Set("position", 1000.0)
	record.Set("labels", []string{})
	record.Set("blocked_by", []string{})
	record.Set("created_by", "cli")
	record.Set("history", []map[string]any{})
	record.Set("board", boardID)
	record.Set("seq", seq)

	if err := app.Save(record); err != nil {
		t.Fatalf("failed to create test task: %v", err)
	}

	return record
}
