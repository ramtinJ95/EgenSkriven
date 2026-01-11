package tui

import (
	"fmt"
	"sort"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ramtinJ95/EgenSkriven/internal/board"
)

// loadBoards creates a command that loads all boards from the database.
// Returns boardsLoadedMsg on success, errMsg on failure.
func loadBoards(app *pocketbase.PocketBase) tea.Cmd {
	return func() tea.Msg {
		records, err := board.GetAll(app)
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to load boards: %w", err)}
		}
		return boardsLoadedMsg{boards: records}
	}
}

// loadTasks creates a command that loads all tasks for a specific board.
// Tasks are loaded and will be grouped by column in the Update handler.
func loadTasks(app *pocketbase.PocketBase, boardID string) tea.Cmd {
	return func() tea.Msg {
		// Build query for tasks in this board
		records, err := app.FindAllRecords("tasks",
			dbx.NewExp("board = {:board}", dbx.Params{"board": boardID}),
		)
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to load tasks: %w", err)}
		}

		// Sort by column and position
		sort.Slice(records, func(i, j int) bool {
			colI := records[i].GetString("column")
			colJ := records[j].GetString("column")
			if colI != colJ {
				return getColumnOrder(colI) < getColumnOrder(colJ)
			}
			return records[i].GetFloat("position") < records[j].GetFloat("position")
		})

		return tasksLoadedMsg{tasks: records}
	}
}

// getColumnOrder returns the display order for a column.
// Used for sorting tasks by column.
func getColumnOrder(column string) int {
	order := map[string]int{
		"backlog":     0,
		"todo":        1,
		"in_progress": 2,
		"need_input":  3,
		"review":      4,
		"done":        5,
	}
	if o, ok := order[column]; ok {
		return o
	}
	return 99 // Unknown columns go to end
}

// loadBoardAndTasks creates a command that loads a specific board and its tasks.
// This is a convenience function for initial load.
func loadBoardAndTasks(app *pocketbase.PocketBase, boardRef string) tea.Cmd {
	return func() tea.Msg {
		// First, find the board
		var boardRecord *core.Record
		var err error

		if boardRef != "" {
			// Use board reference
			boardRecord, err = board.GetByNameOrPrefix(app, boardRef)
			if err != nil {
				return errMsg{err: fmt.Errorf("board not found: %s", boardRef)}
			}
		} else {
			// Get all boards and use the first one
			boards, err := board.GetAll(app)
			if err != nil {
				return errMsg{err: fmt.Errorf("failed to load boards: %w", err)}
			}
			if len(boards) == 0 {
				return errMsg{err: fmt.Errorf("no boards found - create one with 'egenskriven board create'")}
			}
			boardRecord = boards[0]
		}

		// Load tasks for this board
		records, err := app.FindAllRecords("tasks",
			dbx.NewExp("board = {:board}", dbx.Params{"board": boardRecord.Id}),
		)
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to load tasks: %w", err)}
		}

		// Sort by position within each column
		sort.Slice(records, func(i, j int) bool {
			colI := records[i].GetString("column")
			colJ := records[j].GetString("column")
			if colI != colJ {
				return getColumnOrder(colI) < getColumnOrder(colJ)
			}
			return records[i].GetFloat("position") < records[j].GetFloat("position")
		})

		// Return combined message
		return boardAndTasksLoadedMsg{
			board: boardRecord,
			tasks: records,
		}
	}
}

// boardAndTasksLoadedMsg is defined in messages.go

// setStatus creates a command that sends a status message.
// Used for displaying temporary feedback messages.
func setStatus(message string, isError bool) tea.Cmd {
	return func() tea.Msg {
		return statusMsg{
			message: message,
			isError: isError,
		}
	}
}

// clearStatus creates a delayed command that clears the status message.
// Called after displaying a status to auto-clear it.
func clearStatus() tea.Cmd {
	return tea.Tick(
		3*time.Second,
		func(_ time.Time) tea.Msg {
			return statusMsg{message: "", isError: false}
		},
	)
}
