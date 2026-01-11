package tui

import (
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// =============================================================================
// Form Mode
// =============================================================================

// FormMode indicates whether a form is in add or edit mode
type FormMode int

const (
	FormModeAdd FormMode = iota
	FormModeEdit
)

// =============================================================================
// Task Loading Messages
// =============================================================================

// tasksLoadedMsg is sent when tasks are loaded from the database
type tasksLoadedMsg struct {
	tasks []*core.Record
}

// boardsLoadedMsg is sent when boards are loaded from the database
type boardsLoadedMsg struct {
	boards []*core.Record
}

// boardAndTasksLoadedMsg is sent when both board and tasks are loaded.
// Used for the initial load sequence.
type boardAndTasksLoadedMsg struct {
	board *core.Record
	tasks []*core.Record
}

// =============================================================================
// Task CRUD Messages
// =============================================================================

// taskCreatedMsg is sent when a task is successfully created
type taskCreatedMsg struct {
	task      *core.Record
	displayID string
}

// taskUpdatedMsg is sent when a task is successfully updated
type taskUpdatedMsg struct {
	task *core.Record
}

// taskDeletedMsg is sent when a task is successfully deleted
type taskDeletedMsg struct {
	taskID string
	title  string
}

// taskMovedMsg is sent when a task is moved between columns or reordered
type taskMovedMsg struct {
	task       *core.Record
	fromColumn string
	toColumn   string
}

// =============================================================================
// View State Messages
// =============================================================================

// openTaskDetailMsg requests opening the task detail panel
type openTaskDetailMsg struct {
	task TaskItem
}

// closeTaskDetailMsg requests closing the task detail panel
type closeTaskDetailMsg struct{}

// openTaskFormMsg requests opening the task form for add/edit
type openTaskFormMsg struct {
	mode   FormMode
	taskID string    // Empty for add mode
	task   *TaskItem // Existing task data for edit mode
}

// closeTaskFormMsg requests closing the task form
type closeTaskFormMsg struct {
	cancelled bool
}

// openConfirmDialogMsg requests opening a confirmation dialog
type openConfirmDialogMsg struct {
	title   string
	message string
	onYes   func() // Callback when user confirms
}

// closeConfirmDialogMsg requests closing the confirmation dialog
type closeConfirmDialogMsg struct {
	confirmed bool
}

// =============================================================================
// Error and Status Messages
// =============================================================================

// errMsg represents an error from an async operation
type errMsg struct {
	err     error
	context string // What operation failed
}

// Error implements the error interface for errMsg.
func (e errMsg) Error() string {
	if e.context != "" {
		return e.context + ": " + e.err.Error()
	}
	return e.err.Error()
}

// statusMsg shows a temporary status message to the user
type statusMsg struct {
	message  string
	isError  bool
	duration time.Duration
}

// clearStatusMsg clears the status message
type clearStatusMsg struct{}

// =============================================================================
// Form Submission Messages
// =============================================================================

// TaskFormData contains the data from a submitted task form
type TaskFormData struct {
	Title       string
	Description string
	Type        string
	Priority    string
	Column      string
	Labels      []string
	DueDate     string
	EpicID      string
}

// submitTaskFormMsg is sent when the task form is submitted
type submitTaskFormMsg struct {
	mode   FormMode
	taskID string       // Empty for create, set for update
	data   TaskFormData
}

// =============================================================================
// Window Messages
// =============================================================================

// windowSizeMsg is sent when the terminal is resized.
// Components use this to recalculate their dimensions.
// Note: Bubble Tea provides tea.WindowSizeMsg, but we wrap it for consistency.
type windowSizeMsg struct {
	width  int
	height int
}
