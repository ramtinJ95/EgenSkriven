package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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
