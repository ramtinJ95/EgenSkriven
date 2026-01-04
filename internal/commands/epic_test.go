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

func TestIsValidHexColor(t *testing.T) {
	tests := []struct {
		color string
		valid bool
	}{
		{"#3B82F6", true},
		{"#aabbcc", true},
		{"#AABBCC", true},
		{"#000000", true},
		{"#ffffff", true},
		{"#123456", true},
		{"red", false},
		{"#FFF", false},
		{"3B82F6", false},
		{"#GGGGGG", false},
		{"", false},
		{"#12345", false},
		{"#1234567", false},
		{"##123456", false},
		{"#12345G", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("color=%s", tt.color), func(t *testing.T) {
			result := isValidHexColor(tt.color)
			assert.Equal(t, tt.valid, result, "isValidHexColor(%q)", tt.color)
		})
	}
}

func TestResolveEpic_ExactID(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupEpicsCollection(t, app)

	epic := createTestEpic(t, app, "Q1 Launch", "#3B82F6")

	resolved, err := resolveEpic(app, epic.Id)

	require.NoError(t, err)
	assert.Equal(t, epic.Id, resolved.Id)
	assert.Equal(t, "Q1 Launch", resolved.GetString("title"))
}

func TestResolveEpic_IDPrefix(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupEpicsCollection(t, app)

	epic := createTestEpic(t, app, "Q1 Launch", "#3B82F6")

	// Use first 8 characters as prefix
	prefix := epic.Id[:8]
	resolved, err := resolveEpic(app, prefix)

	require.NoError(t, err)
	assert.Equal(t, epic.Id, resolved.Id)
}

func TestResolveEpic_TitleMatch(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupEpicsCollection(t, app)

	epic := createTestEpic(t, app, "Q1 Launch", "#3B82F6")

	resolved, err := resolveEpic(app, "Q1 Launch")

	require.NoError(t, err)
	assert.Equal(t, epic.Id, resolved.Id)
}

func TestResolveEpic_PartialTitleCaseInsensitive(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupEpicsCollection(t, app)

	epic := createTestEpic(t, app, "Q1 Launch", "#3B82F6")

	resolved, err := resolveEpic(app, "q1")

	require.NoError(t, err)
	assert.Equal(t, epic.Id, resolved.Id)
}

func TestResolveEpic_NotFound(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupEpicsCollection(t, app)

	_, err := resolveEpic(app, "nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no epic found")
}

func TestResolveEpic_Ambiguous(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupEpicsCollection(t, app)

	// Create two epics with similar titles
	createTestEpic(t, app, "Q1 Launch", "#3B82F6")
	createTestEpic(t, app, "Q1 Launch Prep", "#22C55E")

	_, err := resolveEpic(app, "Q1 Launch")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ambiguous")
}

func TestGetEpicTaskCount_NoTasks(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupEpicsCollection(t, app)
	setupTasksCollectionWithEpic(t, app)

	epic := createTestEpic(t, app, "Empty Epic", "")

	count := getEpicTaskCount(app, epic.Id)

	assert.Equal(t, 0, count)
}

func TestGetEpicTaskCount_WithTasks(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupEpicsCollection(t, app)
	setupTasksCollectionWithEpic(t, app)

	epic := createTestEpic(t, app, "Test Epic", "#3B82F6")

	// Create tasks linked to the epic
	for i := 0; i < 3; i++ {
		createTestTaskWithEpic(t, app, fmt.Sprintf("Task %d", i), epic.Id)
	}

	count := getEpicTaskCount(app, epic.Id)

	assert.Equal(t, 3, count)
}

func TestGetEpicTaskCount_NonexistentEpic(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupEpicsCollection(t, app)
	setupTasksCollectionWithEpic(t, app)

	count := getEpicTaskCount(app, "nonexistent-id")

	assert.Equal(t, 0, count)
}

// ========== Helper Functions ==========

func setupEpicsCollection(t *testing.T, app *pocketbase.PocketBase) {
	t.Helper()

	_, err := app.FindCollectionByNameOrId("epics")
	if err == nil {
		return
	}

	collection := core.NewBaseCollection("epics")
	collection.Fields.Add(&core.TextField{
		Name:     "title",
		Required: true,
		Min:      1,
		Max:      200,
	})
	collection.Fields.Add(&core.TextField{
		Name: "description",
		Max:  5000,
	})
	collection.Fields.Add(&core.TextField{
		Name:    "color",
		Pattern: `^#[0-9A-Fa-f]{6}$`,
		Max:     7,
	})

	if err := app.Save(collection); err != nil {
		t.Fatalf("failed to create epics collection: %v", err)
	}
}

func setupTasksCollectionWithEpic(t *testing.T, app *pocketbase.PocketBase) {
	t.Helper()

	_, err := app.FindCollectionByNameOrId("tasks")
	if err == nil {
		return
	}

	epicsCollection, err := app.FindCollectionByNameOrId("epics")
	if err != nil {
		t.Fatalf("epics collection must exist before tasks: %v", err)
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
	collection.Fields.Add(&core.RelationField{
		Name:          "epic",
		CollectionId:  epicsCollection.Id,
		MaxSelect:     1,
		CascadeDelete: false,
	})

	if err := app.Save(collection); err != nil {
		t.Fatalf("failed to create tasks collection: %v", err)
	}
}

func createTestEpic(t *testing.T, app *pocketbase.PocketBase, title, color string) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("epics")
	require.NoError(t, err)

	record := core.NewRecord(collection)
	record.Set("title", title)
	if color != "" {
		record.Set("color", color)
	}

	require.NoError(t, app.Save(record))

	return record
}

func createTestTaskWithEpic(t *testing.T, app *pocketbase.PocketBase, title, epicID string) *core.Record {
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
	record.Set("blocked_by", []string{})
	record.Set("created_by", "cli")
	record.Set("history", []map[string]any{})
	if epicID != "" {
		record.Set("epic", epicID)
	}

	require.NoError(t, app.Save(record))

	return record
}
