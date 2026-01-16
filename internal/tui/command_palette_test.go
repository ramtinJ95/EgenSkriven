package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestNewCommandPalette(t *testing.T) {
	commands := []Command{
		{ID: "test", Name: "Test Command", Shortcut: "t"},
	}

	palette := NewCommandPalette(commands)

	assert.NotNil(t, palette)
	assert.False(t, palette.IsVisible())
	assert.Equal(t, 1, len(palette.commands))
	assert.Equal(t, 10, palette.maxResults)
}

func TestCommandPalette_ShowHide(t *testing.T) {
	commands := []Command{
		{ID: "test", Name: "Test Command"},
	}
	palette := NewCommandPalette(commands)

	// Initially hidden
	assert.False(t, palette.IsVisible())

	// Show
	palette.Show()
	assert.True(t, palette.IsVisible())

	// Hide
	palette.Hide()
	assert.False(t, palette.IsVisible())
}

func TestCommandPalette_ShowResetsState(t *testing.T) {
	commands := []Command{
		{ID: "test1", Name: "Test One"},
		{ID: "test2", Name: "Test Two"},
	}
	palette := NewCommandPalette(commands)
	palette.selected = 1
	palette.input.SetValue("something")

	palette.Show()

	assert.Equal(t, 0, palette.selected)
	assert.Equal(t, "", palette.input.Value())
	assert.Equal(t, len(commands), len(palette.filtered))
}

func TestCommandPalette_SetSize(t *testing.T) {
	palette := NewCommandPalette(nil)
	palette.SetSize(100, 50)

	assert.Equal(t, 100, palette.width)
	assert.Equal(t, 50, palette.height)
}

func TestCommandPalette_FilterCommands(t *testing.T) {
	commands := []Command{
		{ID: "new-task", Name: "New Task", Description: "Create new task"},
		{ID: "edit-task", Name: "Edit Task", Description: "Edit existing task"},
		{ID: "delete-task", Name: "Delete Task", Description: "Delete a task"},
		{ID: "move-left", Name: "Move Left", Description: "Move item left"},
	}
	palette := NewCommandPalette(commands)

	// Filter for "task" - matches 3 commands with "Task" in name
	palette.filterCommands("task")
	assert.Equal(t, 3, len(palette.filtered))

	// Filter for "move"
	palette.filterCommands("move")
	assert.Equal(t, 1, len(palette.filtered))
	assert.Equal(t, "move-left", palette.filtered[0].ID)

	// Empty filter shows all
	palette.filterCommands("")
	assert.Equal(t, len(commands), len(palette.filtered))
}

func TestCommandPalette_FilterCommandsNoMatch(t *testing.T) {
	commands := []Command{
		{ID: "new-task", Name: "New Task"},
	}
	palette := NewCommandPalette(commands)

	palette.filterCommands("xyz")
	assert.Equal(t, 0, len(palette.filtered))
	assert.Equal(t, 0, palette.selected)
}

func TestCommandPalette_NavigateUpDown(t *testing.T) {
	commands := []Command{
		{ID: "cmd1", Name: "Command One"},
		{ID: "cmd2", Name: "Command Two"},
		{ID: "cmd3", Name: "Command Three"},
	}
	palette := NewCommandPalette(commands)
	palette.Show()

	// Initial selection
	assert.Equal(t, 0, palette.selected)

	// Navigate down
	palette.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, palette.selected)

	palette.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 2, palette.selected)

	// Can't go past end
	palette.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 2, palette.selected)

	// Navigate up
	palette.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 1, palette.selected)

	// Can't go past beginning
	palette.Update(tea.KeyMsg{Type: tea.KeyUp})
	palette.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 0, palette.selected)
}

func TestCommandPalette_EscapeCloses(t *testing.T) {
	commands := []Command{
		{ID: "test", Name: "Test"},
	}
	palette := NewCommandPalette(commands)
	palette.Show()

	assert.True(t, palette.IsVisible())

	palette.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.False(t, palette.IsVisible())
}

func TestCommandPalette_ViewWhenHidden(t *testing.T) {
	palette := NewCommandPalette(nil)
	palette.Hide()

	view := palette.View()
	assert.Equal(t, "", view)
}

func TestCommandPalette_ViewShowsCommands(t *testing.T) {
	commands := []Command{
		{ID: "new", Name: "New Task", Shortcut: "n", Category: "Tasks"},
		{ID: "edit", Name: "Edit Task", Shortcut: "e", Category: "Tasks"},
	}
	palette := NewCommandPalette(commands)
	palette.Show()
	palette.SetSize(80, 40)

	view := palette.View()

	assert.Contains(t, view, "Command Palette")
	assert.Contains(t, view, "New Task")
	assert.Contains(t, view, "[n]")
	assert.Contains(t, view, "Edit Task")
}

func TestCommandPalette_ViewShowsNoResults(t *testing.T) {
	commands := []Command{
		{ID: "test", Name: "Test"},
	}
	palette := NewCommandPalette(commands)
	palette.Show()
	palette.SetSize(80, 40)
	palette.filterCommands("xyz")

	view := palette.View()

	assert.Contains(t, view, "No matching commands")
}

func TestFuzzyScore_ExactMatch(t *testing.T) {
	score := fuzzyScore("task", "task")
	assert.Greater(t, score, 0)
}

func TestFuzzyScore_SubstringMatch(t *testing.T) {
	score := fuzzyScore("task", "new task")
	assert.Greater(t, score, 0)
}

func TestFuzzyScore_FuzzyMatch(t *testing.T) {
	// "nt" matches "new task" (n...t)
	score := fuzzyScore("nt", "new task")
	assert.Greater(t, score, 0)
}

func TestFuzzyScore_NoMatch(t *testing.T) {
	score := fuzzyScore("xyz", "new task")
	assert.Equal(t, 0, score)
}

func TestFuzzyScore_PartialSequence(t *testing.T) {
	// "nwt" should match "new task" as n-e-w t-a-s-k
	score := fuzzyScore("nwt", "new task")
	assert.Greater(t, score, 0)
}

func TestFuzzyScore_OrderMatters(t *testing.T) {
	// "tn" should not match "new task" because t comes before n
	score := fuzzyScore("tn", "new task")
	assert.Equal(t, 0, score)
}

func TestDefaultCommands_ReturnsCommands(t *testing.T) {
	actions := &CommandActions{
		NewTask:          func() tea.Cmd { return nil },
		EditTask:         func() tea.Cmd { return nil },
		DeleteTask:       func() tea.Cmd { return nil },
		ViewTask:         func() tea.Cmd { return nil },
		MoveTaskLeft:     func() tea.Cmd { return nil },
		MoveTaskRight:    func() tea.Cmd { return nil },
		MoveToColumn:     func(col string) func() tea.Cmd { return func() tea.Cmd { return nil } },
		Search:           func() tea.Cmd { return nil },
		FilterByPriority: func() tea.Cmd { return nil },
		FilterByType:     func() tea.Cmd { return nil },
		FilterByEpic:     func() tea.Cmd { return nil },
		FilterByLabel:    func() tea.Cmd { return nil },
		ClearFilters:     func() tea.Cmd { return nil },
		SwitchBoard:      func() tea.Cmd { return nil },
		Refresh:          func() tea.Cmd { return nil },
		ToggleHelp:       func() tea.Cmd { return nil },
		SelectAll:        func() tea.Cmd { return nil },
		BulkMove:         func() tea.Cmd { return nil },
		BulkDelete:       func() tea.Cmd { return nil },
	}

	commands := DefaultCommands(actions)

	assert.Greater(t, len(commands), 0)

	// Check some expected commands exist
	found := make(map[string]bool)
	for _, cmd := range commands {
		found[cmd.ID] = true
	}

	assert.True(t, found["new-task"])
	assert.True(t, found["edit-task"])
	assert.True(t, found["move-left"])
	assert.True(t, found["search"])
	assert.True(t, found["switch-board"])
}

func TestCommandPalette_SetCommands(t *testing.T) {
	palette := NewCommandPalette(nil)
	assert.Equal(t, 0, len(palette.commands))

	commands := []Command{
		{ID: "test", Name: "Test"},
	}
	palette.SetCommands(commands)

	assert.Equal(t, 1, len(palette.commands))
	assert.Equal(t, 1, len(palette.filtered))
}

func TestCommandPalette_SelectionClampsOnFilter(t *testing.T) {
	commands := []Command{
		{ID: "cmd1", Name: "Alpha"},
		{ID: "cmd2", Name: "Beta"},
		{ID: "cmd3", Name: "Gamma"},
	}
	palette := NewCommandPalette(commands)
	palette.Show()

	// Select last item
	palette.selected = 2

	// Filter to single result
	palette.filterCommands("alpha")

	// Selection should be clamped
	assert.Equal(t, 0, palette.selected)
}

func TestCommandPalette_CtrlPCtrlNNavigation(t *testing.T) {
	commands := []Command{
		{ID: "cmd1", Name: "Command One"},
		{ID: "cmd2", Name: "Command Two"},
	}
	palette := NewCommandPalette(commands)
	palette.Show()

	// Ctrl+N goes down
	palette.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	assert.Equal(t, 1, palette.selected)

	// Ctrl+P goes up
	palette.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	assert.Equal(t, 0, palette.selected)
}
