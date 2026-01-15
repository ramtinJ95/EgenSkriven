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

// boardSwitchedMsg is sent when the user selects a different board
type boardSwitchedMsg struct {
	boardID string
}

// boardTasksLoadedMsg is sent when tasks for a specific board are loaded
type boardTasksLoadedMsg struct {
	tasks []*core.Record
}

// boardColumnsMsg contains the columns for the current board
type boardColumnsMsg struct {
	columns []string
}

// lastBoardSavedMsg confirms the last board was persisted to config
type lastBoardSavedMsg struct {
	boardID string
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
	taskID string // Empty for create, set for update
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

// =============================================================================
// Realtime Messages
// =============================================================================

// RealtimeEvent represents a parsed realtime event from PocketBase.
type RealtimeEvent struct {
	Action     string                 // "create", "update", "delete"
	Collection string                 // "tasks", "boards", "epics"
	Record     map[string]interface{} // The record data
}

// realtimeConnectedMsg is sent when SSE connection is established.
type realtimeConnectedMsg struct {
	clientID string
}

// realtimeDisconnectedMsg is sent when SSE connection is lost.
type realtimeDisconnectedMsg struct {
	err error
}

// realtimeEventMsg wraps a realtime event for the Update loop.
type realtimeEventMsg struct {
	event RealtimeEvent
}

// realtimeErrorMsg indicates an error in the realtime subsystem.
type realtimeErrorMsg struct {
	err error
}

// realtimeReconnectMsg triggers a reconnection attempt.
type realtimeReconnectMsg struct {
	attempt int
}

// =============================================================================
// Connection Status Messages
// =============================================================================

// ConnectionStatus represents the current realtime connection state.
type ConnectionStatus int

const (
	ConnectionDisconnected ConnectionStatus = iota
	ConnectionConnecting
	ConnectionConnected
	ConnectionReconnecting
)

func (s ConnectionStatus) String() string {
	switch s {
	case ConnectionDisconnected:
		return "disconnected"
	case ConnectionConnecting:
		return "connecting"
	case ConnectionConnected:
		return "connected"
	case ConnectionReconnecting:
		return "reconnecting"
	default:
		return "unknown"
	}
}

// connectionStatusMsg updates the connection status indicator.
type connectionStatusMsg struct {
	status ConnectionStatus
}

// =============================================================================
// Polling Fallback Messages
// =============================================================================

// pollStartMsg initiates polling mode (fallback when SSE fails).
type pollStartMsg struct{}

// pollStopMsg stops polling mode.
type pollStopMsg struct{}

// pollTickMsg triggers a poll cycle.
type pollTickMsg struct {
	time time.Time
}

// pollResultMsg contains the results of a poll cycle.
type pollResultMsg struct {
	tasks     []*core.Record
	checkTime time.Time
	err       error
}

// =============================================================================
// Task Realtime Update Messages
// =============================================================================

// tasksReloadedMsg is sent when tasks are bulk reloaded (after reconnect).
type tasksReloadedMsg struct {
	tasks []*core.Record
}

// =============================================================================
// Server Status Messages
// =============================================================================

// serverOnlineMsg is sent when server becomes reachable.
type serverOnlineMsg struct{}

// serverOfflineMsg is sent when server becomes unreachable.
type serverOfflineMsg struct{}

// =============================================================================
// Filter Messages
// =============================================================================

// FilterChangedMsg indicates filters have been updated
type FilterChangedMsg struct {
	FilterState *FilterState
}

// QuickFilterMsg triggers a quick filter key sequence
type QuickFilterMsg struct {
	Prefix rune // 'f' for filter commands
	Key    rune // 'p', 't', 'l', 'e', 'c'
}

// ShowSearchMsg triggers the search overlay
type ShowSearchMsg struct{}

// HideSearchMsg closes the search overlay
type HideSearchMsg struct{}

// ShowFilterSelectorMsg opens a filter selector
type ShowFilterSelectorMsg struct {
	Type FilterSelectorType
}

// ClearFiltersMsg clears all active filters
type ClearFiltersMsg struct{}

// ToggleFilterBarFocusMsg toggles focus on the filter bar
type ToggleFilterBarFocusMsg struct{}

// RefreshFilteredTasksMsg triggers re-filtering of tasks
type RefreshFilteredTasksMsg struct{}

// LoadLabelsMsg requests loading available labels
type LoadLabelsMsg struct{}

// LabelsLoadedMsg contains available labels
type LabelsLoadedMsg struct {
	Labels []string
}

// LoadEpicsMsg requests loading available epics
type LoadEpicsMsg struct{}

// EpicsLoadedMsg contains available epics
type EpicsLoadedMsg struct {
	Epics []EpicOption
}
