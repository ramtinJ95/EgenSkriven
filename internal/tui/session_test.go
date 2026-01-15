package tui

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionState_SaveAndLoad(t *testing.T) {
	// Clean up any existing session
	_ = ClearSession()
	defer func() { _ = ClearSession() }()

	// Create filter state with some data
	filterState := NewFilterState()
	filterState.SetSearchQuery("test query")
	filterState.AddFilter(Filter{Field: "priority", Operator: "is", Value: "high"})

	// Save session
	state := SessionState{
		FilterState:    filterState,
		CurrentBoardID: "board-123",
		FocusedColumn:  2,
	}

	err := SaveSession(state)
	require.NoError(t, err)

	// Load session
	loaded, err := LoadSession()
	require.NoError(t, err)
	require.NotNil(t, loaded)

	// Verify loaded state
	assert.Equal(t, "board-123", loaded.CurrentBoardID)
	assert.Equal(t, 2, loaded.FocusedColumn)
	assert.NotNil(t, loaded.FilterState)
	assert.Equal(t, "test query", loaded.FilterState.GetSearchQuery())
	assert.Len(t, loaded.FilterState.GetFilters(), 1)
}

func TestSessionState_LoadNoSession(t *testing.T) {
	// Clean up any existing session
	_ = ClearSession()

	// Load should return nil with no error
	loaded, err := LoadSession()
	assert.NoError(t, err)
	assert.Nil(t, loaded)
}

func TestSessionState_ClearSession(t *testing.T) {
	// Create and save a session
	state := SessionState{
		CurrentBoardID: "board-123",
		FocusedColumn:  1,
	}
	err := SaveSession(state)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(sessionFilePath())
	assert.NoError(t, err)

	// Clear session
	err = ClearSession()
	require.NoError(t, err)

	// Verify file is gone
	_, err = os.Stat(sessionFilePath())
	assert.True(t, os.IsNotExist(err))
}

func TestSessionState_ClearNonExistent(t *testing.T) {
	// Clear any existing session
	_ = ClearSession()

	// Clearing non-existent session should not error
	err := ClearSession()
	assert.NoError(t, err)
}

func TestSessionLoadedMsg(t *testing.T) {
	// Test that SessionLoadedMsg can carry session state
	filterState := NewFilterState()
	filterState.SetSearchQuery("test")

	msg := SessionLoadedMsg{
		Session: &SessionState{
			FilterState:    filterState,
			CurrentBoardID: "board-abc",
			FocusedColumn:  3,
		},
	}

	assert.NotNil(t, msg.Session)
	assert.Equal(t, "board-abc", msg.Session.CurrentBoardID)
	assert.Equal(t, "test", msg.Session.FilterState.GetSearchQuery())
}

func TestSessionState_EmptyFilterState(t *testing.T) {
	// Clean up any existing session
	_ = ClearSession()
	defer func() { _ = ClearSession() }()

	// Save session with nil filter state
	state := SessionState{
		FilterState:    nil,
		CurrentBoardID: "board-456",
		FocusedColumn:  0,
	}

	err := SaveSession(state)
	require.NoError(t, err)

	// Load and verify
	loaded, err := LoadSession()
	require.NoError(t, err)
	require.NotNil(t, loaded)

	assert.Equal(t, "board-456", loaded.CurrentBoardID)
	assert.Nil(t, loaded.FilterState)
}
