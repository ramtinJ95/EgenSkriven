package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ramtinJ95/EgenSkriven/internal/config"
)

func TestBoardOptionFilterValue(t *testing.T) {
	opt := BoardOption{
		Name:   "Work Tasks",
		Prefix: "WRK",
	}

	filterVal := opt.FilterValue()
	assert.Contains(t, filterVal, "Work Tasks")
	assert.Contains(t, filterVal, "WRK")
}

func TestBoardOptionTitle(t *testing.T) {
	opt := BoardOption{
		Name:   "Work",
		Prefix: "WRK",
	}

	title := opt.Title()
	assert.Contains(t, title, "WRK")
	assert.Contains(t, title, "Work")

	// Test with IsSelected
	opt.IsSelected = true
	title = opt.Title()
	assert.Contains(t, title, "current")
}

func TestBoardOptionDescription(t *testing.T) {
	tests := []struct {
		name      string
		taskCount int
		expected  string
	}{
		{"zero tasks", 0, "No tasks"},
		{"one task", 1, "1 task"},
		{"multiple tasks", 5, "5 tasks"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := BoardOption{TaskCount: tt.taskCount}
			desc := opt.Description()
			assert.Equal(t, tt.expected, desc)
		})
	}
}

func TestBoardSelector(t *testing.T) {
	options := []BoardOption{
		{ID: "id1", Name: "Work", Prefix: "WRK", TaskCount: 10},
		{ID: "id2", Name: "Personal", Prefix: "PER", TaskCount: 5},
		{ID: "id3", Name: "Learning", Prefix: "LRN", TaskCount: 0},
	}

	selector := NewBoardSelector(options, "id1")
	selector.SetSize(50, 20)

	// Verify initial selection is current board
	selected, ok := selector.SelectedBoard()
	require.True(t, ok)
	assert.Equal(t, "id1", selected.ID)
	assert.True(t, selected.IsSelected)

	// Navigate down
	selector, _ = selector.Update(tea.KeyMsg{Type: tea.KeyDown})
	selected, ok = selector.SelectedBoard()
	require.True(t, ok)
	assert.Equal(t, "id2", selected.ID)

	// Navigate down again
	selector, _ = selector.Update(tea.KeyMsg{Type: tea.KeyDown})
	selected, ok = selector.SelectedBoard()
	require.True(t, ok)
	assert.Equal(t, "id3", selected.ID)
}

func TestBoardSelectorSelection(t *testing.T) {
	options := []BoardOption{
		{ID: "id1", Name: "Work", Prefix: "WRK"},
		{ID: "id2", Name: "Personal", Prefix: "PER"},
	}

	selector := NewBoardSelector(options, "id1")
	selector.SetSize(50, 20)

	// Navigate to second board
	selector, _ = selector.Update(tea.KeyMsg{Type: tea.KeyDown})

	// Select with Enter
	_, cmd := selector.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)

	// Execute command to get message
	msg := cmd()
	switchMsg, ok := msg.(boardSwitchedMsg)
	require.True(t, ok)
	assert.Equal(t, "id2", switchMsg.boardID)
}

func TestSaveLastBoard(t *testing.T) {
	// Create temp directory for config
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	err := os.Chdir(tmpDir)
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	// Reset config cache for fresh load
	config.ResetGlobalConfigCache()

	// Save a board ID
	cmd := saveLastBoard("test-board-id")
	msg := cmd()

	// Verify success message
	savedMsg, ok := msg.(lastBoardSavedMsg)
	require.True(t, ok)
	assert.Equal(t, "test-board-id", savedMsg.boardID)

	// Verify file was created
	configPath := filepath.Join(tmpDir, ".egenskriven", "config.json")
	_, err = os.Stat(configPath)
	require.NoError(t, err)

	// Verify content
	cfg, err := config.LoadProjectConfigFrom(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, "test-board-id", cfg.DefaultBoard)
}

func TestLoadDefaultBoard(t *testing.T) {
	// Create temp directory for config
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	err := os.Chdir(tmpDir)
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	// Reset config cache
	config.ResetGlobalConfigCache()

	// Create config with default board
	cfg := config.DefaultConfig()
	cfg.DefaultBoard = "my-default-board"
	err = config.SaveConfig(tmpDir, cfg)
	require.NoError(t, err)

	// Load default board
	cmd := loadDefaultBoard()
	msg := cmd()

	// Verify message
	switchMsg, ok := msg.(boardSwitchedMsg)
	require.True(t, ok)
	assert.Equal(t, "my-default-board", switchMsg.boardID)
}

func TestLoadDefaultBoardNoConfig(t *testing.T) {
	// Create temp directory without config
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	err := os.Chdir(tmpDir)
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	// Reset config cache
	config.ResetGlobalConfigCache()

	// Load default board (should return nil when no config)
	cmd := loadDefaultBoard()
	msg := cmd()

	// Should be nil when no default configured
	assert.Nil(t, msg)
}

func TestHeaderSetBoard(t *testing.T) {
	header := NewHeader()
	header.SetWidth(80)

	header.SetBoard("Work Tasks", "WRK", "#FF5733")
	header.SetTaskCount(42)

	view := header.View()

	assert.Contains(t, view, "WRK")
	assert.Contains(t, view, "Work Tasks")
	assert.Contains(t, view, "42 tasks")
}

func TestHeaderEmptyBoard(t *testing.T) {
	header := NewHeader()
	header.SetWidth(80)

	view := header.View()
	assert.Empty(t, view)
}

func TestHeaderTaskCount(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		expected string
	}{
		{"zero", 0, "No tasks"},
		{"one", 1, "1 task"},
		{"many", 10, "10 tasks"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := NewHeader()
			header.SetWidth(80)
			header.SetBoard("Test", "TST", "")
			header.SetTaskCount(tt.count)

			view := header.View()
			assert.Contains(t, view, tt.expected)
		})
	}
}

func TestHeaderHeight(t *testing.T) {
	header := NewHeader()
	assert.Equal(t, 2, header.Height())
}

func TestBoardSelectorCancellation(t *testing.T) {
	options := []BoardOption{
		{ID: "id1", Name: "Work", Prefix: "WRK"},
		{ID: "id2", Name: "Personal", Prefix: "PER"},
	}

	selector := NewBoardSelector(options, "id1")
	selector.SetSize(50, 20)

	// Cancel with Escape - should return nil command
	_, cmd := selector.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.Nil(t, cmd)

	// Cancel with 'q' - should also return nil command
	selector2 := NewBoardSelector(options, "id1")
	selector2.SetSize(50, 20)
	_, cmd = selector2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	assert.Nil(t, cmd)
}

func TestBoardOptionsFromRecordsEmpty(t *testing.T) {
	// Test with nil records
	options := BoardOptionsFromRecords(nil, nil)
	assert.Empty(t, options)

	// Test with empty slice
	options = BoardOptionsFromRecords([]*core.Record{}, nil)
	assert.Empty(t, options)
}

func TestBoardSelectorView(t *testing.T) {
	options := []BoardOption{
		{ID: "id1", Name: "Work", Prefix: "WRK", TaskCount: 10},
	}

	selector := NewBoardSelector(options, "id1")
	selector.SetSize(50, 20)

	view := selector.View()

	// View should contain the list content wrapped in a modal border
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Switch Board")
}

func TestBoardSwitchingFlow(t *testing.T) {
	// This test simulates the full board switching flow:
	// 1. User has multiple boards
	// 2. User opens board selector (b key)
	// 3. User navigates to a different board
	// 4. User selects the board (enter)
	// 5. boardSwitchedMsg is produced

	options := []BoardOption{
		{ID: "board-1", Name: "Work", Prefix: "WRK", TaskCount: 5},
		{ID: "board-2", Name: "Personal", Prefix: "PER", TaskCount: 3},
		{ID: "board-3", Name: "Learning", Prefix: "LRN", TaskCount: 0},
	}

	// Create selector with "Work" as current board
	selector := NewBoardSelector(options, "board-1")
	selector.SetSize(60, 20)

	// Verify initial state
	selected, ok := selector.SelectedBoard()
	require.True(t, ok)
	assert.Equal(t, "board-1", selected.ID)
	assert.True(t, selected.IsSelected, "Current board should be marked as selected")

	// Navigate down twice to "Learning"
	selector, _ = selector.Update(tea.KeyMsg{Type: tea.KeyDown})
	selector, _ = selector.Update(tea.KeyMsg{Type: tea.KeyDown})

	selected, ok = selector.SelectedBoard()
	require.True(t, ok)
	assert.Equal(t, "board-3", selected.ID)

	// Select the board
	_, cmd := selector.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd, "Selection should produce a command")

	// Execute command and verify message
	msg := cmd()
	switchMsg, ok := msg.(boardSwitchedMsg)
	require.True(t, ok, "Command should produce boardSwitchedMsg")
	assert.Equal(t, "board-3", switchMsg.boardID)
}

func TestBoardSelectorNavigationWrapping(t *testing.T) {
	options := []BoardOption{
		{ID: "id1", Name: "First", Prefix: "FST"},
		{ID: "id2", Name: "Second", Prefix: "SND"},
	}

	selector := NewBoardSelector(options, "id1")
	selector.SetSize(50, 20)

	// Start at first item
	selected, _ := selector.SelectedBoard()
	assert.Equal(t, "id1", selected.ID)

	// Navigate down to second
	selector, _ = selector.Update(tea.KeyMsg{Type: tea.KeyDown})
	selected, _ = selector.SelectedBoard()
	assert.Equal(t, "id2", selected.ID)

	// Navigate up back to first
	selector, _ = selector.Update(tea.KeyMsg{Type: tea.KeyUp})
	selected, _ = selector.SelectedBoard()
	assert.Equal(t, "id1", selected.ID)
}
