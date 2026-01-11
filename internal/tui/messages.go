package tui

import (
	"github.com/pocketbase/pocketbase/core"
)

// errMsg signals an error occurred during an async operation.
// The TUI will display this to the user.
type errMsg struct {
	err error
}

// Error implements the error interface for errMsg.
func (e errMsg) Error() string {
	return e.err.Error()
}

// boardsLoadedMsg signals that boards have been loaded from the database.
// Sent after the initial load or when boards are refreshed.
type boardsLoadedMsg struct {
	boards []*core.Record
}

// tasksLoadedMsg signals that tasks have been loaded for the current board.
// Contains all tasks grouped by column in the View.
type tasksLoadedMsg struct {
	tasks []*core.Record
}

// taskMovedMsg signals that a task was successfully moved to a new column.
// Used to update the UI after a move operation.
type taskMovedMsg struct {
	task      *core.Record
	oldColumn string
	newColumn string
}

// windowSizeMsg is sent when the terminal is resized.
// Components use this to recalculate their dimensions.
// Note: Bubble Tea provides tea.WindowSizeMsg, but we wrap it for consistency.
type windowSizeMsg struct {
	width  int
	height int
}

// statusMsg displays a temporary status message in the status bar.
// Used for success messages, warnings, and informational updates.
type statusMsg struct {
	message string
	isError bool
}
