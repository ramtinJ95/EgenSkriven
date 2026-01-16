package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"github.com/ramtinJ95/EgenSkriven/internal/board"
)

// ViewState represents which view is currently active
type ViewState int

const (
	ViewBoard ViewState = iota
	ViewTaskDetail
	ViewTaskForm
	ViewConfirm
	ViewBoardSelector
)

// App is the main TUI application model.
// It implements tea.Model and manages the overall application state.
type App struct {
	// Core dependencies
	pb        *pocketbase.PocketBase
	serverURL string

	// Board state
	currentBoard *core.Record   // Currently displayed board
	boards       []*core.Record // All available boards

	// Column state - 6 columns for default workflow
	columns     []Column
	focusedCol  int
	columnOrder []string // Column status order

	// Realtime state
	realtimeClient *RealtimeClient
	statusBar      *StatusBar
	usePolling     bool
	lastPollTime   time.Time

	// UI state
	width  int
	height int
	ready  bool // True once initial data is loaded
	view   ViewState

	// Overlays
	taskDetail    *TaskDetail
	taskForm      *TaskForm
	confirmDialog *ConfirmDialog
	boardSelector *BoardSelector

	// Header component
	header *Header

	// Pending operations
	pendingDeleteTaskID string
	pendingBulkDelete   []string // Task IDs for bulk delete
	pendingBulkMove     []string // Task IDs for bulk move
	pendingBulkMoveKey  bool     // True when waiting for column number after 'm'

	// Components
	help help.Model
	keys keyMap

	// Status message
	statusMessage string
	statusIsError bool

	// Initial board reference (from CLI flag)
	initialBoardRef string

	// Error state
	err error

	// Filter state
	filterState    *FilterState
	filterBar      FilterBar
	searchOverlay  SearchOverlay
	filterSelector FilterSelector

	// Quick filter key state
	pendingFilterKey bool // True after 'f' is pressed

	// Cached filter data
	availableLabels []string
	availableEpics  []EpicOption

	// Help overlay
	helpOverlay *HelpOverlay

	// Subtask state
	subtaskCounts      map[string]int        // taskID -> subtask count
	expandedSubtaskView *ExpandedSubtaskView // Tracks which tasks have expanded subtasks

	// Multi-select state
	selectionState *SelectionState

	// Command palette
	commandPalette *CommandPalette
}

// NewApp creates a new TUI application.
// boardRef is optional - if empty, uses the first available board.
func NewApp(pb *pocketbase.PocketBase, boardRef string) *App {
	h := help.New()
	h.ShowAll = false

	serverURL := DefaultServerURL
	filterState := NewFilterState()

	app := &App{
		pb:                  pb,
		serverURL:           serverURL,
		realtimeClient:      NewRealtimeClient(serverURL),
		statusBar:           NewStatusBar(),
		keys:                defaultKeyMap(),
		help:                h,
		header:              NewHeader(),
		focusedCol:          0,
		initialBoardRef:     boardRef,
		columnOrder:         []string{"backlog", "todo", "in_progress", "need_input", "review", "done"},
		view:                ViewBoard,
		lastPollTime:        time.Now(),
		filterState:         filterState,
		filterBar:           NewFilterBar(filterState),
		searchOverlay:       NewSearchOverlay(filterState),
		filterSelector:      NewFilterSelector(),
		pendingFilterKey:    false,
		helpOverlay:         NewHelpOverlay(),
		subtaskCounts:       make(map[string]int),
		expandedSubtaskView: NewExpandedSubtaskView(),
		selectionState:      NewSelectionState(),
	}

	// Initialize command palette with actions
	app.commandPalette = NewCommandPalette(DefaultCommands(app.createCommandActions()))

	return app
}

// Init implements tea.Model.
// Called once when the program starts. Returns initial commands.
func (a *App) Init() tea.Cmd {
	// Set initial connection status
	a.statusBar.SetConnectionStatus(ConnectionConnecting)

	// Load boards list and current board's tasks in parallel.
	// loadBoards populates a.boards for the board selector.
	// loadBoardAndTasks loads the initial board and its tasks.
	// Also check server status to determine if we can use realtime.
	// Load session state to restore filters from previous run.
	return tea.Batch(
		loadBoards(a.pb),
		loadBoardAndTasks(a.pb, a.initialBoardRef),
		CheckServerStatus(a.serverURL),
		loadSession,
	)
}

// loadSession loads the previous session state if available.
func loadSession() tea.Msg {
	session, err := LoadSession()
	if err != nil || session == nil {
		return nil
	}
	return SessionLoadedMsg{Session: session}
}

// Update implements tea.Model.
// Handles all messages and updates the model state.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	// Window resize
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.updateColumnSizes()
		// Update header width
		if a.header != nil {
			a.header.SetWidth(msg.Width)
		}
		// Update overlay sizes
		if a.taskDetail != nil {
			a.taskDetail.SetSize(a.width/2, a.height-4)
		}
		if a.taskForm != nil {
			a.taskForm.SetSize(a.width/2, a.height-10)
		}
		if a.boardSelector != nil {
			a.boardSelector.SetSize(min(60, a.width-4), min(20, a.height-4))
		}
		// Update filter component sizes
		a.filterBar.SetWidth(msg.Width)
		a.searchOverlay.SetSize(msg.Width, msg.Height)
		a.filterSelector.SetSize(msg.Width, msg.Height)
		return a, nil

	// =================================================================
	// Initial Load Messages
	// =================================================================

	case boardAndTasksLoadedMsg:
		a.currentBoard = msg.board
		a.initializeColumns(msg.tasks)
		a.updateHeaderInfo()
		a.ready = true
		// Load epics, labels, and subtask counts for filtering and badge display
		return a, tea.Batch(
			CmdLoadEpics(a.pb, msg.board.Id),
			CmdLoadLabels(a.pb, msg.board.Id),
			CmdLoadSubtaskCounts(a.pb, msg.board.Id),
		)

	// =================================================================
	// Board Switching Messages
	// =================================================================

	case boardsLoadedMsg:
		a.boards = msg.boards
		return a, nil

	case boardSwitchedMsg:
		// Find and set the board record
		for _, b := range a.boards {
			if b.Id == msg.boardID {
				a.currentBoard = b
				break
			}
		}

		// Close board selector if open
		if a.view == ViewBoardSelector {
			a.view = ViewBoard
			a.boardSelector = nil
		}

		// Clear selection state and expanded subtasks when switching boards
		// to avoid stale state from the previous board
		if a.selectionState != nil {
			a.selectionState.Clear()
		}
		if a.expandedSubtaskView != nil {
			a.expandedSubtaskView.Clear()
		}

		// Load the new board's data
		return a, switchBoard(a.pb, msg.boardID)

	case boardTasksLoadedMsg:
		if a.currentBoard != nil {
			a.updateColumnsWithTasks(msg.tasks)
			a.updateHeaderInfo()
		}
		return a, nil

	case boardColumnsMsg:
		// Update column order if board has custom columns
		if len(msg.columns) > 0 {
			a.columnOrder = msg.columns
		}
		return a, nil

	case lastBoardSavedMsg:
		// Silently acknowledge save
		return a, nil

	// =================================================================
	// Task CRUD Messages
	// =================================================================

	case taskCreatedMsg:
		// Guard against nil board (edge case during rapid actions)
		if a.currentBoard == nil {
			return a, nil
		}
		// Reload tasks and show success message
		cmds = append(cmds,
			loadTasks(a.pb, a.currentBoard.Id),
			showStatus("Created: "+msg.task.GetString("title")+" ["+msg.displayID+"]", false, 3*time.Second),
		)
		// Close form
		a.taskForm = nil
		a.view = ViewBoard

	case taskUpdatedMsg:
		// Guard against nil board (edge case during rapid actions)
		if a.currentBoard == nil {
			return a, nil
		}
		// Reload tasks and show success message
		cmds = append(cmds,
			loadTasks(a.pb, a.currentBoard.Id),
			showStatus("Updated: "+msg.task.GetString("title"), false, 3*time.Second),
		)
		// Close form and update detail if open
		a.taskForm = nil
		if a.taskDetail != nil {
			// Update the detail view with new data
			a.taskDetail.UpdateTask(a.recordToTaskItem(msg.task))
		}
		a.view = ViewBoard

	case taskDeletedMsg:
		// Guard against nil board (edge case during rapid actions)
		if a.currentBoard == nil {
			return a, nil
		}
		// Reload tasks and show success message
		cmds = append(cmds,
			loadTasks(a.pb, a.currentBoard.Id),
			showStatus("Deleted: "+msg.title, false, 3*time.Second),
		)
		// Close any overlays
		a.taskDetail = nil
		a.confirmDialog = nil
		a.view = ViewBoard

	case BulkResultMsg:
		// Handle bulk operation result
		if a.currentBoard == nil {
			return a, nil
		}
		resultMsg := FormatBulkResult(msg.Result)
		isError := len(msg.Result.FailedIDs) > 0
		cmds = append(cmds,
			loadTasks(a.pb, a.currentBoard.Id),
			showStatus(resultMsg, isError, 3*time.Second),
		)
		// Close any overlays
		a.taskDetail = nil
		a.confirmDialog = nil
		a.view = ViewBoard

	case taskMovedMsg:
		// Guard against nil board (edge case during rapid actions)
		if a.currentBoard == nil {
			return a, nil
		}
		// Reload tasks and show status
		cmds = append(cmds, loadTasks(a.pb, a.currentBoard.Id))
		if msg.fromColumn != msg.toColumn {
			cmds = append(cmds, showStatus("Moved to "+msg.toColumn, false, 2*time.Second))
		}

	case tasksLoadedMsg:
		a.updateColumnsWithTasks(msg.tasks)
		return a, tea.Batch(cmds...)

	// =================================================================
	// View State Messages
	// =================================================================

	case openTaskDetailMsg:
		a.taskDetail = NewTaskDetail(msg.task, a.width/2, a.height-4)
		a.view = ViewTaskDetail

	case closeTaskDetailMsg:
		a.taskDetail = nil
		a.view = ViewBoard

	case openTaskFormMsg:
		if msg.mode == FormModeEdit && msg.task != nil {
			a.taskForm = NewTaskFormWithData(msg.task, a.width/2, a.height-10)
		} else {
			a.taskForm = NewTaskForm(FormModeAdd, a.width/2, a.height-10)
		}
		a.view = ViewTaskForm

	case closeTaskFormMsg:
		a.taskForm = nil
		a.view = ViewBoard

	case openConfirmDialogMsg:
		a.confirmDialog = NewConfirmDialog(msg.title, msg.message)
		a.view = ViewConfirm

	case closeConfirmDialogMsg:
		if msg.confirmed {
			// Execute the pending action (delete)
			if len(a.pendingBulkDelete) > 0 {
				// Bulk delete
				cmds = append(cmds, BulkDelete(a.pb, a.pendingBulkDelete))
				a.pendingBulkDelete = nil
				// Clear selection
				if a.selectionState != nil {
					a.selectionState.Clear()
					a.refreshColumnSelections()
				}
			} else if a.pendingDeleteTaskID != "" {
				// Single delete
				cmds = append(cmds, deleteTask(a.pb, a.pendingDeleteTaskID))
				a.pendingDeleteTaskID = ""
			}
		} else {
			// Cancelled - clear pending operations
			a.pendingBulkDelete = nil
			a.pendingDeleteTaskID = ""
		}
		a.confirmDialog = nil
		a.view = ViewBoard

	case submitTaskFormMsg:
		if msg.mode == FormModeAdd {
			cmds = append(cmds, createTask(a.pb, a.currentBoard, msg.data))
		} else {
			cmds = append(cmds, updateTask(a.pb, msg.taskID, msg.data))
		}

	// =================================================================
	// Status Messages
	// =================================================================

	case statusMsg:
		a.statusMessage = msg.message
		a.statusIsError = msg.isError
		if msg.duration > 0 {
			cmds = append(cmds, clearStatusAfter(msg.duration))
		}

	case clearStatusMsg:
		a.statusMessage = ""
		a.statusIsError = false

	// =================================================================
	// Error Messages
	// =================================================================

	case errMsg:
		a.err = msg.err
		a.statusMessage = msg.Error()
		a.statusIsError = true
		cmds = append(cmds, clearStatusAfter(5*time.Second))

	// =================================================================
	// Server Status Messages
	// =================================================================

	case serverOnlineMsg:
		// Server is online, try to connect via SSE
		a.statusBar.SetConnectionStatus(ConnectionConnecting)
		cmds = append(cmds, a.realtimeClient.Connect())

	case serverOfflineMsg:
		// Server is offline, use polling fallback
		a.statusBar.SetConnectionStatus(ConnectionDisconnected)
		a.usePolling = true
		// Schedule a retry check
		cmds = append(cmds, ScheduleServerCheck(a.serverURL, 10*time.Second))

	// =================================================================
	// Realtime Connection Messages
	// =================================================================

	case realtimeConnectedMsg:
		// SSE connection established
		a.statusBar.SetConnectionStatus(ConnectionConnected)
		a.usePolling = false
		// Start listening for events
		cmds = append(cmds, WaitForEvent(a.realtimeClient))

	case realtimeDisconnectedMsg:
		// SSE connection lost
		a.statusBar.SetConnectionStatusWithMessage(ConnectionReconnecting, "reconnecting...")
		// Attempt to reconnect with backoff
		cmds = append(cmds, ReconnectWithBackoff(a.realtimeClient, 0))

	case realtimeReconnectMsg:
		// Attempt to reconnect
		a.statusBar.SetConnectionStatusWithMessage(ConnectionReconnecting,
			fmt.Sprintf("attempt %d/%d", msg.attempt, maxReconnectAttempts))
		cmds = append(cmds, a.realtimeClient.Connect())

	case realtimeErrorMsg:
		// Realtime error, fall back to polling
		a.statusBar.SetConnectionStatusWithMessage(ConnectionDisconnected, "using polling")
		a.usePolling = true
		if a.currentBoard != nil {
			cmds = append(cmds, StartPolling(PollConfig{
				Interval: pollInterval,
				BoardID:  a.currentBoard.Id,
			}))
		}

	case realtimeEventMsg:
		// Handle the realtime event
		cmds = append(cmds, a.handleRealtimeEvent(msg.event))
		// Continue listening for more events
		cmds = append(cmds, WaitForEvent(a.realtimeClient))

	// =================================================================
	// Polling Messages
	// =================================================================

	case pollStartMsg:
		// Switch to polling mode
		a.usePolling = true
		a.statusBar.SetConnectionStatusWithMessage(ConnectionDisconnected, "polling")
		if a.currentBoard != nil {
			cmds = append(cmds, StartPolling(PollConfig{
				Interval: pollInterval,
				BoardID:  a.currentBoard.Id,
			}))
		}

	case pollTickMsg:
		// Time to poll for changes
		if a.usePolling && a.currentBoard != nil {
			cmds = append(cmds, PollForChanges(a.pb, a.currentBoard.Id, a.lastPollTime))
		}

	case pollResultMsg:
		// Process poll results
		if msg.err != nil {
			// Poll failed, continue polling anyway
			cmds = append(cmds, ContinuePolling(pollInterval))
		} else {
			a.lastPollTime = msg.checkTime
			if len(msg.tasks) > 0 {
				// Changes detected, update the display
				a.updateColumnsWithTasks(msg.tasks)
			}
			// Schedule next poll
			if a.usePolling {
				cmds = append(cmds, ContinuePolling(pollInterval))
			}
		}

	// =================================================================
	// Filter Messages
	// =================================================================

	case FilterSelectedMsg:
		a.filterState.AddFilter(msg.Filter)
		a.refreshFilteredColumns()
		return a, nil

	case FilterCancelledMsg:
		return a, nil

	case SearchAppliedMsg:
		a.refreshFilteredColumns()
		return a, nil

	case SearchCancelledMsg:
		return a, nil

	case FilterChangedMsg:
		a.refreshFilteredColumns()
		return a, nil

	case LabelsLoadedMsg:
		a.availableLabels = msg.Labels
		return a, nil

	case EpicsLoadedMsg:
		a.availableEpics = msg.Epics
		// Refresh columns to show epic badges on tasks
		a.refreshColumnEpics()
		return a, nil

	case SubtaskCountsLoadedMsg:
		a.subtaskCounts = msg.Counts
		// Refresh columns to show subtask indicators on tasks
		a.refreshColumnSubtaskCounts()
		return a, nil

	case SubtasksExpandedMsg:
		if msg.Err != nil {
			return a, showStatus("Failed to load subtasks", true, 2*time.Second)
		}
		// Cache the loaded subtasks
		a.expandedSubtaskView.SetSubtasks(msg.ParentID, msg.Subtasks)
		return a, nil

	case ClearFiltersMsg:
		a.filterState.Clear()
		a.refreshFilteredColumns()
		return a, nil

	case RefreshFilteredTasksMsg:
		a.refreshFilteredColumns()
		return a, nil

	// =================================================================
	// Session Messages
	// =================================================================

	case SessionLoadedMsg:
		if msg.Session != nil {
			// Restore filter state from session
			if msg.Session.FilterState != nil {
				a.filterState = msg.Session.FilterState
				a.filterBar = NewFilterBar(a.filterState)
				a.searchOverlay = NewSearchOverlay(a.filterState)
			}
			// Restore focused column (clamped to valid range)
			if msg.Session.FocusedColumn >= 0 && msg.Session.FocusedColumn < len(a.columns) {
				// Update focus only if columns are initialized
				if len(a.columns) > 0 {
					a.columns[a.focusedCol].SetFocused(false)
					a.focusedCol = msg.Session.FocusedColumn
					a.columns[a.focusedCol].SetFocused(true)
				}
			}
			// Note: Board switching is handled separately by initialBoardRef
		}
		return a, nil

	// =================================================================
	// Keyboard Input
	// =================================================================

	case tea.KeyMsg:
		// Handle command palette first (it captures input when visible)
		if a.commandPalette != nil && a.commandPalette.IsVisible() {
			palette, cmd := a.commandPalette.Update(msg)
			a.commandPalette = palette
			return a, cmd
		}

		// Handle filter overlays (they capture input)
		if a.searchOverlay.IsActive() {
			overlay, cmd := a.searchOverlay.Update(msg)
			a.searchOverlay = overlay
			return a, cmd
		}

		if a.filterSelector.IsActive() {
			selector, cmd := a.filterSelector.Update(msg)
			a.filterSelector = selector
			return a, cmd
		}

		// Handle help overlay - only ? or Esc to close
		if a.helpOverlay != nil && a.helpOverlay.IsVisible() {
			switch msg.String() {
			case "?", "esc", "q":
				a.helpOverlay.Hide()
				return a, nil
			}
			return a, nil // Ignore all other keys when help is visible
		}

		switch a.view {
		case ViewBoard:
			return a.handleBoardKeys(msg)
		case ViewTaskDetail:
			return a.handleDetailKeys(msg)
		case ViewTaskForm:
			return a.handleFormKeys(msg)
		case ViewConfirm:
			return a.handleConfirmKeys(msg)
		case ViewBoardSelector:
			return a.handleBoardSelectorKeys(msg)
		}
	}

	return a, tea.Batch(cmds...)
}

// =============================================================================
// Key Handlers
// =============================================================================

// handleBoardKeys processes keyboard input when in board view.
func (a *App) handleBoardKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If not ready, ignore keys
	if !a.ready {
		return a, nil
	}

	// Handle pending filter key (second key in 'fp', 'ft', etc.)
	if a.pendingFilterKey {
		a.pendingFilterKey = false
		return a.handleFilterKey(msg.String())
	}

	// Handle pending bulk move key (column number after 'm')
	if a.pendingBulkMoveKey {
		a.pendingBulkMoveKey = false
		return a.handleBulkMoveKey(msg.String())
	}

	switch {
	// Quit
	case matchKey(msg, a.keys.Quit):
		// Save session state before exiting
		a.saveSession()

		// Clean up realtime connection before exiting
		if a.realtimeClient != nil {
			a.realtimeClient.Disconnect()
		}
		return a, tea.Quit

	// Column navigation - left
	case matchKey(msg, a.keys.Left):
		if a.focusedCol > 0 {
			a.columns[a.focusedCol].SetFocused(false)
			a.focusedCol--
			a.columns[a.focusedCol].SetFocused(true)
		}
		return a, nil

	// Column navigation - right
	case matchKey(msg, a.keys.Right):
		if a.focusedCol < len(a.columns)-1 {
			a.columns[a.focusedCol].SetFocused(false)
			a.focusedCol++
			a.columns[a.focusedCol].SetFocused(true)
		}
		return a, nil

	// Task navigation - up/down
	case matchKey(msg, a.keys.Up), matchKey(msg, a.keys.Down):
		// Pass to focused column
		if a.focusedCol >= 0 && a.focusedCol < len(a.columns) {
			col, cmd := a.columns[a.focusedCol].Update(msg)
			a.columns[a.focusedCol] = col
			return a, cmd
		}
		return a, nil

	// Enter - view task details
	case matchKey(msg, a.keys.Enter):
		task := a.getSelectedTask()
		if task != nil {
			return a, func() tea.Msg {
				return openTaskDetailMsg{task: *task}
			}
		}
		return a, nil

	// Help toggle
	case matchKey(msg, a.keys.Help):
		a.helpOverlay.Toggle()
		return a, nil
	}

	// Additional key handling by string
	switch msg.String() {
	case "/":
		// Open search overlay
		cmd := a.searchOverlay.Show()
		return a, cmd

	case "f":
		// Start filter key sequence
		a.pendingFilterKey = true
		return a, nil

	case "b":
		// Open board selector
		a.openBoardSelector()
		return a, nil

	case "n":
		// New task
		return a, func() tea.Msg {
			return openTaskFormMsg{mode: FormModeAdd}
		}

	case "e":
		// Edit task
		task := a.getSelectedTask()
		if task != nil {
			return a, func() tea.Msg {
				return openTaskFormMsg{
					mode:   FormModeEdit,
					taskID: task.ID,
					task:   task,
				}
			}
		}

	case "d":
		// Delete task(s) with confirmation
		if a.selectionState != nil && a.selectionState.IsActive() {
			// Bulk delete
			count := a.selectionState.Count()
			a.pendingBulkDelete = a.selectionState.GetSelected()
			a.confirmDialog = NewBulkDeleteConfirmDialog(count)
			a.view = ViewConfirm
		} else {
			// Single delete
			task := a.getSelectedTask()
			if task != nil {
				a.pendingDeleteTaskID = task.ID
				a.confirmDialog = NewDeleteConfirmDialog(task.TaskTitle)
				a.view = ViewConfirm
			}
		}
		return a, nil

	case "m":
		// Bulk move - start waiting for column number
		if a.selectionState != nil && a.selectionState.IsActive() {
			a.pendingBulkMove = a.selectionState.GetSelected()
			a.pendingBulkMoveKey = true
			return a, showStatus("Move to column: 1-6", false, 5*time.Second)
		}
		return a, nil

	case "H":
		// Move task to previous column
		return a.moveTaskLeft()

	case "L":
		// Move task to next column
		return a.moveTaskRight()

	case "K":
		// Move task up in column
		return a.reorderTaskUp()

	case "J":
		// Move task down in column
		return a.reorderTaskDown()

	case "1", "2", "3", "4", "5", "6":
		// Move task to column by number
		columnIndex := int(msg.String()[0] - '1')
		return a.moveTaskToColumnIndex(columnIndex)

	case "tab":
		// Toggle subtask expansion for current task
		return a.toggleSubtaskExpansion()

	case " ":
		// Space toggles task selection
		return a.toggleTaskSelection()

	case "ctrl+a":
		// Select all tasks in current column
		return a.selectAllInColumn()

	case "ctrl+k":
		// Open command palette
		if a.commandPalette != nil {
			a.commandPalette.Show()
		}
		return a, nil

	case "esc":
		// Clear selection if in selection mode
		if a.selectionState != nil && a.selectionState.IsActive() {
			a.selectionState.Clear()
			a.refreshColumnSelections()
			return a, showStatus("Selection cleared", false, 2*time.Second)
		}
	}

	return a, nil
}

// handleFilterKey handles the second key in a filter sequence (fp, ft, fl, fe, fc)
func (a *App) handleFilterKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "p":
		// Filter by priority
		cmd := a.filterSelector.ShowPriority()
		return a, cmd

	case "t":
		// Filter by type
		cmd := a.filterSelector.ShowType()
		return a, cmd

	case "l":
		// Filter by label
		if len(a.availableLabels) == 0 {
			return a, showStatus("No labels available", true, 2*time.Second)
		}
		cmd := a.filterSelector.ShowLabel(a.availableLabels)
		return a, cmd

	case "e":
		// Filter by epic
		if len(a.availableEpics) == 0 {
			return a, showStatus("No epics available", true, 2*time.Second)
		}
		cmd := a.filterSelector.ShowEpic(a.availableEpics)
		return a, cmd

	case "b":
		// Filter by blocked status
		cmd := a.filterSelector.ShowBlocked()
		return a, cmd

	case "c":
		// Clear all filters
		a.filterState.Clear()
		a.refreshFilteredColumns()
		return a, showStatus("Filters cleared", false, 2*time.Second)

	default:
		// Unknown second key - ignore
		return a, nil
	}
}

// handleBulkMoveKey handles the column number after 'm' for bulk move.
func (a *App) handleBulkMoveKey(key string) (tea.Model, tea.Cmd) {
	// Check if key is a valid column number
	if key >= "1" && key <= "6" {
		columnIndex := int(key[0] - '1')
		if columnIndex >= 0 && columnIndex < len(a.columnOrder) {
			targetColumn := a.columnOrder[columnIndex]
			taskIDs := a.pendingBulkMove
			a.pendingBulkMove = nil

			// Clear selection
			if a.selectionState != nil {
				a.selectionState.Clear()
				a.refreshColumnSelections()
			}

			// Get board ID
			boardID := ""
			if a.currentBoard != nil {
				boardID = a.currentBoard.Id
			}

			return a, BulkMove(a.pb, taskIDs, targetColumn, boardID)
		}
	}

	// Invalid key or cancelled
	a.pendingBulkMove = nil
	return a, showStatus("Bulk move cancelled", false, 2*time.Second)
}

// handleDetailKeys processes keyboard input when in task detail view.
func (a *App) handleDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.taskDetail == nil {
		a.view = ViewBoard
		return a, nil
	}

	// Let the detail component handle its own keys
	td, cmd := a.taskDetail.Update(msg)
	a.taskDetail = td
	return a, cmd
}

// handleFormKeys processes keyboard input when in task form view.
func (a *App) handleFormKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.taskForm == nil {
		a.view = ViewBoard
		return a, nil
	}

	// Let the form component handle its own keys
	tf, cmd := a.taskForm.Update(msg)
	a.taskForm = tf
	return a, cmd
}

// handleConfirmKeys processes keyboard input when in confirm dialog view.
func (a *App) handleConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.confirmDialog == nil {
		a.view = ViewBoard
		return a, nil
	}

	// Let the confirm dialog handle its own keys
	cd, cmd := a.confirmDialog.Update(msg)
	a.confirmDialog = cd
	return a, cmd
}

// handleBoardSelectorKeys processes keyboard input when in board selector view.
func (a *App) handleBoardSelectorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.boardSelector == nil {
		a.view = ViewBoard
		return a, nil
	}

	// Check for escape to close
	if msg.String() == "esc" || msg.String() == "b" {
		a.view = ViewBoard
		a.boardSelector = nil
		return a, nil
	}

	// Let the board selector handle its own keys
	bs, cmd := a.boardSelector.Update(msg)
	a.boardSelector = bs
	return a, cmd
}

// openBoardSelector opens the board selector overlay.
func (a *App) openBoardSelector() {
	if len(a.boards) == 0 {
		// Load boards first if not loaded
		return
	}

	// Get task counts for each board
	boardIDs := make([]string, len(a.boards))
	for i, b := range a.boards {
		boardIDs[i] = b.Id
	}
	taskCounts := getTaskCountsForBoards(a.pb, boardIDs)

	// Convert records to options
	options := BoardOptionsFromRecords(a.boards, taskCounts)

	// Get current board ID
	currentBoardID := ""
	if a.currentBoard != nil {
		currentBoardID = a.currentBoard.Id
	}

	// Create board selector
	a.boardSelector = NewBoardSelector(options, currentBoardID)
	a.boardSelector.SetSize(min(60, a.width-4), min(20, a.height-4))
	a.view = ViewBoardSelector
}

// updateHeaderInfo updates the header with current board info.
func (a *App) updateHeaderInfo() {
	if a.header == nil || a.currentBoard == nil {
		return
	}

	name := a.currentBoard.GetString("name")
	prefix := a.currentBoard.GetString("prefix")
	color := a.currentBoard.GetString("color")

	a.header.SetBoard(name, prefix, color)

	// Count tasks
	taskCount := 0
	for _, col := range a.columns {
		taskCount += len(col.Items())
	}
	a.header.SetTaskCount(taskCount)
}

// =============================================================================
// Task Movement Helpers
// =============================================================================

func (a *App) moveTaskLeft() (tea.Model, tea.Cmd) {
	task := a.getSelectedTask()
	if task == nil {
		return a, nil
	}

	currentIndex := a.getColumnIndex(task.Column)
	if currentIndex <= 0 {
		return a, showStatus("Already in first column", true, 2*time.Second)
	}

	targetColumn := a.columnOrder[currentIndex-1]
	return a, moveTaskToColumn(a.pb, task.ID, targetColumn)
}

func (a *App) moveTaskRight() (tea.Model, tea.Cmd) {
	task := a.getSelectedTask()
	if task == nil {
		return a, nil
	}

	currentIndex := a.getColumnIndex(task.Column)
	if currentIndex >= len(a.columnOrder)-1 {
		return a, showStatus("Already in last column", true, 2*time.Second)
	}

	targetColumn := a.columnOrder[currentIndex+1]
	return a, moveTaskToColumn(a.pb, task.ID, targetColumn)
}

func (a *App) reorderTaskUp() (tea.Model, tea.Cmd) {
	task := a.getSelectedTask()
	if task == nil {
		return a, nil
	}
	return a, reorderTaskInColumn(a.pb, task.ID, true)
}

func (a *App) reorderTaskDown() (tea.Model, tea.Cmd) {
	task := a.getSelectedTask()
	if task == nil {
		return a, nil
	}
	return a, reorderTaskInColumn(a.pb, task.ID, false)
}

func (a *App) moveTaskToColumnIndex(index int) (tea.Model, tea.Cmd) {
	if index < 0 || index >= len(a.columnOrder) {
		return a, nil
	}

	task := a.getSelectedTask()
	if task == nil {
		return a, nil
	}

	targetColumn := a.columnOrder[index]
	if targetColumn == task.Column {
		return a, nil // Already in this column
	}

	return a, moveTaskToColumn(a.pb, task.ID, targetColumn)
}

// createCommandActions creates CommandActions that wire to app methods.
func (a *App) createCommandActions() *CommandActions {
	return &CommandActions{
		NewTask: func() tea.Cmd {
			return func() tea.Msg {
				return openTaskFormMsg{mode: FormModeAdd}
			}
		},
		EditTask: func() tea.Cmd {
			task := a.getSelectedTask()
			if task == nil {
				return nil
			}
			return func() tea.Msg {
				return openTaskFormMsg{
					mode:   FormModeEdit,
					taskID: task.ID,
					task:   task,
				}
			}
		},
		DeleteTask: func() tea.Cmd {
			task := a.getSelectedTask()
			if task == nil {
				return nil
			}
			a.pendingDeleteTaskID = task.ID
			a.confirmDialog = NewDeleteConfirmDialog(task.TaskTitle)
			a.view = ViewConfirm
			return nil
		},
		ViewTask: func() tea.Cmd {
			task := a.getSelectedTask()
			if task == nil {
				return nil
			}
			return func() tea.Msg {
				return openTaskDetailMsg{task: *task}
			}
		},
		MoveTaskLeft: func() tea.Cmd {
			_, cmd := a.moveTaskLeft()
			return cmd
		},
		MoveTaskRight: func() tea.Cmd {
			_, cmd := a.moveTaskRight()
			return cmd
		},
		MoveToColumn: func(column string) func() tea.Cmd {
			return func() tea.Cmd {
				idx := a.getColumnIndex(column)
				if idx >= 0 {
					_, cmd := a.moveTaskToColumnIndex(idx)
					return cmd
				}
				return nil
			}
		},
		Search: func() tea.Cmd {
			return a.searchOverlay.Show()
		},
		FilterByPriority: func() tea.Cmd {
			return a.filterSelector.ShowPriority()
		},
		FilterByType: func() tea.Cmd {
			return a.filterSelector.ShowType()
		},
		FilterByEpic: func() tea.Cmd {
			if len(a.availableEpics) == 0 {
				return showStatus("No epics available", true, 2*time.Second)
			}
			return a.filterSelector.ShowEpic(a.availableEpics)
		},
		FilterByLabel: func() tea.Cmd {
			if len(a.availableLabels) == 0 {
				return showStatus("No labels available", true, 2*time.Second)
			}
			return a.filterSelector.ShowLabel(a.availableLabels)
		},
		ClearFilters: func() tea.Cmd {
			a.filterState.Clear()
			a.refreshFilteredColumns()
			return showStatus("Filters cleared", false, 2*time.Second)
		},
		SwitchBoard: func() tea.Cmd {
			a.openBoardSelector()
			return nil
		},
		Refresh: func() tea.Cmd {
			if a.currentBoard != nil {
				return tea.Batch(
					loadTasks(a.pb, a.currentBoard.Id),
					showStatus("Refreshing...", false, 2*time.Second),
				)
			}
			return nil
		},
		ToggleHelp: func() tea.Cmd {
			a.helpOverlay.Toggle()
			return nil
		},
		SelectAll: func() tea.Cmd {
			_, cmd := a.selectAllInColumn()
			return cmd
		},
		BulkMove: func() tea.Cmd {
			if a.selectionState != nil && a.selectionState.IsActive() {
				a.pendingBulkMove = a.selectionState.GetSelected()
				a.pendingBulkMoveKey = true
				return showStatus("Move to column: 1-6", false, 5*time.Second)
			}
			return showStatus("No tasks selected", true, 2*time.Second)
		},
		BulkDelete: func() tea.Cmd {
			if a.selectionState != nil && a.selectionState.IsActive() {
				count := a.selectionState.Count()
				a.pendingBulkDelete = a.selectionState.GetSelected()
				a.confirmDialog = NewBulkDeleteConfirmDialog(count)
				a.view = ViewConfirm
				return nil
			}
			return showStatus("No tasks selected", true, 2*time.Second)
		},
	}
}

func (a *App) getColumnIndex(column string) int {
	for i, col := range a.columnOrder {
		if col == column {
			return i
		}
	}
	return -1
}

func (a *App) getSelectedTask() *TaskItem {
	if a.focusedCol < 0 || a.focusedCol >= len(a.columns) {
		return nil
	}
	return a.columns[a.focusedCol].SelectedTask()
}

// toggleSubtaskExpansion expands or collapses subtasks for the selected task.
func (a *App) toggleSubtaskExpansion() (tea.Model, tea.Cmd) {
	task := a.getSelectedTask()
	if task == nil {
		return a, nil
	}

	// Only toggle if task has subtasks
	if !task.HasSubtasks || task.SubtaskCount == 0 {
		return a, showStatus("This task has no subtasks", false, 2*time.Second)
	}

	// Toggle expansion state
	a.expandedSubtaskView.Toggle(task.ID)
	isExpanded := a.expandedSubtaskView.IsExpanded(task.ID)

	// Update the task item in the column
	a.updateTaskSubtaskExpansion(task.ID, isExpanded)

	// If expanding and subtasks not loaded yet, load them
	if isExpanded && len(a.expandedSubtaskView.GetSubtasks(task.ID)) == 0 {
		boardPrefix := ""
		if a.currentBoard != nil {
			boardPrefix = a.currentBoard.GetString("prefix")
		}
		return a, CmdLoadSubtasks(a.pb, task.ID, boardPrefix)
	}

	return a, nil
}

// updateTaskSubtaskExpansion updates the SubtasksExpanded flag for a task in the columns.
func (a *App) updateTaskSubtaskExpansion(taskID string, expanded bool) {
	for colIdx := range a.columns {
		items := a.columns[colIdx].Items()
		for itemIdx, item := range items {
			if task, ok := item.(TaskItem); ok && task.ID == taskID {
				task.SubtasksExpanded = expanded
				items[itemIdx] = task
				a.columns[colIdx].SetItems(items)
				return
			}
		}
	}
}

// toggleTaskSelection toggles selection for the current task.
func (a *App) toggleTaskSelection() (tea.Model, tea.Cmd) {
	task := a.getSelectedTask()
	if task == nil {
		return a, nil
	}

	a.selectionState.ToggleTask(task.ID)
	// Update the selection visual in columns
	a.refreshColumnSelections()
	return a, nil
}

// selectAllInColumn selects all tasks in the current column.
func (a *App) selectAllInColumn() (tea.Model, tea.Cmd) {
	if a.focusedCol < 0 || a.focusedCol >= len(a.columns) {
		return a, nil
	}

	col := a.columns[a.focusedCol]
	items := col.Items()
	tasks := make([]TaskItem, 0, len(items))
	for _, item := range items {
		if task, ok := item.(TaskItem); ok {
			tasks = append(tasks, task)
		}
	}

	if len(tasks) == 0 {
		return a, showStatus("No tasks in column", false, 2*time.Second)
	}

	a.selectionState.SelectAllInColumn(tasks)
	// Update the selection visual in columns
	a.refreshColumnSelections()
	return a, showStatus(fmt.Sprintf("Selected %d tasks", len(tasks)), false, 2*time.Second)
}

// refreshColumnSelections updates IsSelected flag on all tasks based on selection state.
func (a *App) refreshColumnSelections() {
	if a.selectionState == nil {
		return
	}

	for colIdx := range a.columns {
		items := a.columns[colIdx].Items()
		changed := false
		for itemIdx, item := range items {
			if task, ok := item.(TaskItem); ok {
				isSelected := a.selectionState.IsSelected(task.ID)
				if task.IsSelected != isSelected {
					task.IsSelected = isSelected
					items[itemIdx] = task
					changed = true
				}
			}
		}
		if changed {
			a.columns[colIdx].SetItems(items)
		}
	}
}

// =============================================================================
// View Rendering
// =============================================================================

// View implements tea.Model.
// Returns the string to render to the terminal.
func (a *App) View() string {
	if !a.ready {
		return a.renderLoading()
	}

	if a.err != nil && a.currentBoard == nil {
		return a.renderError()
	}

	var sections []string

	// Header
	sections = append(sections, a.renderHeader())

	// Filter bar (only if filters active)
	if a.filterState != nil && a.filterState.HasActiveFilters() {
		sections = append(sections, a.filterBar.View())
	}

	// Main content depends on view state
	switch a.view {
	case ViewTaskDetail:
		// Board + detail panel side by side
		boardView := a.renderColumns()
		detailView := ""
		if a.taskDetail != nil {
			detailView = a.taskDetail.View()
		}
		sections = append(sections, lipgloss.JoinHorizontal(lipgloss.Top, boardView, detailView))

	case ViewTaskForm:
		// Board with form overlay
		boardView := a.renderColumns()
		if a.taskForm != nil {
			// Center the form over the board
			formView := a.taskForm.View()
			sections = append(sections, a.overlayCenter(boardView, formView))
		} else {
			sections = append(sections, boardView)
		}

	case ViewConfirm:
		// Board with confirm dialog overlay
		boardView := a.renderColumns()
		if a.confirmDialog != nil {
			// Center the dialog over the board
			dialogView := a.confirmDialog.View()
			sections = append(sections, a.overlayCenter(boardView, dialogView))
		} else {
			sections = append(sections, boardView)
		}

	case ViewBoardSelector:
		// Board with board selector overlay
		boardView := a.renderColumns()
		if a.boardSelector != nil {
			// Center the selector over the board
			selectorView := a.boardSelector.View()
			sections = append(sections, a.overlayCenter(boardView, selectorView))
		} else {
			sections = append(sections, boardView)
		}

	default:
		// Normal board view
		sections = append(sections, a.renderColumns())
	}

	// Status bar
	sections = append(sections, a.renderStatusBar())

	view := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Overlay search if active
	if a.searchOverlay.IsActive() {
		view = a.overlayCenter(view, a.searchOverlay.View())
	}

	// Overlay filter selector if active
	if a.filterSelector.IsActive() {
		view = a.overlayCenter(view, a.filterSelector.View())
	}

	// Overlay command palette if visible
	if a.commandPalette != nil && a.commandPalette.IsVisible() {
		a.commandPalette.SetSize(a.width, a.height)
		view = a.overlayCenter(view, a.commandPalette.View())
	}

	// Overlay help if visible
	if a.helpOverlay != nil && a.helpOverlay.IsVisible() {
		a.helpOverlay.SetSize(a.width, a.height)
		view = a.helpOverlay.View()
	}

	return view
}

// overlayCenter places the overlay in the center of the background
func (a *App) overlayCenter(background, overlay string) string {
	// Dim the entire background
	bgLines := strings.Split(background, "\n")
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	var dimmedLines []string
	for _, line := range bgLines {
		dimmedLines = append(dimmedLines, dimStyle.Render(line))
	}
	dimmedBg := strings.Join(dimmedLines, "\n")

	// Use lipgloss.Place to correctly position the overlay
	// This handles ANSI escape sequences properly
	bgHeight := lipgloss.Height(dimmedBg)
	bgWidth := a.width
	if bgWidth <= 0 {
		bgWidth = lipgloss.Width(dimmedBg)
	}

	return lipgloss.Place(
		bgWidth,
		bgHeight,
		lipgloss.Center,
		lipgloss.Center,
		overlay,
		lipgloss.WithWhitespaceBackground(lipgloss.NoColor{}),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("240")),
	)
}

// renderLoading shows a loading message while data is being fetched.
func (a *App) renderLoading() string {
	style := lipgloss.NewStyle().
		Width(a.width).
		Height(a.height).
		Align(lipgloss.Center, lipgloss.Center)

	return style.Render("Loading...")
}

// renderError shows an error message when loading fails.
func (a *App) renderError() string {
	style := lipgloss.NewStyle().
		Width(a.width).
		Height(a.height).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(lipgloss.Color("196"))

	return style.Render(fmt.Sprintf("Error: %v", a.err))
}

// renderHeader renders the board title and info with filter status.
func (a *App) renderHeader() string {
	if a.currentBoard == nil {
		return ""
	}

	boardName := a.currentBoard.GetString("name")
	boardPrefix := a.currentBoard.GetString("prefix")

	// Build header parts
	var parts []string

	// Board title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))
	parts = append(parts, titleStyle.Render(fmt.Sprintf("EgenSkriven - %s (%s)", boardName, boardPrefix)))

	// Task count with filter indicator
	totalTasks := a.getTotalTaskCount()
	filteredTasks := a.getFilteredTaskCount()

	countStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	if a.filterState != nil && a.filterState.HasActiveFilters() {
		// Show FILTERED indicator and filtered/total count
		filterIndicator := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true).
			Render(" FILTERED")

		countText := countStyle.Render(
			fmt.Sprintf(" (%d/%d tasks)", filteredTasks, totalTasks),
		)
		parts = append(parts, filterIndicator, countText)

		// Add filter summary
		summary := a.getFilterSummary()
		if summary != "" {
			summaryStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("62")).
				Italic(true)
			parts = append(parts, summaryStyle.Render(" | "+summary))
		}
	} else {
		// Show total only
		countText := countStyle.Render(
			fmt.Sprintf(" (%d tasks)", totalTasks),
		)
		parts = append(parts, countText)
	}

	header := lipgloss.JoinHorizontal(lipgloss.Center, parts...)
	return headerStyle.Width(a.width).Render(header)
}

// renderColumns renders all kanban columns horizontally.
func (a *App) renderColumns() string {
	if len(a.columns) == 0 {
		return ""
	}

	// Calculate available height for columns
	// Total height minus header (1) and status bar (1)
	colHeight := a.height - 3
	if colHeight < 5 {
		colHeight = 5
	}

	// Calculate column width (equal distribution)
	// When in detail view, use less width for columns
	totalWidth := a.width - 2
	if a.view == ViewTaskDetail {
		totalWidth = a.width / 2
	}
	colWidth := totalWidth / len(a.columns)
	if colWidth < 15 {
		colWidth = 15
	}

	// Render each column
	var cols []string
	for i := range a.columns {
		a.columns[i].SetSize(colWidth, colHeight)

		// Apply focused/blurred border style
		var style lipgloss.Style
		if a.columns[i].IsFocused() {
			style = focusedColumnStyle.Width(colWidth).Height(colHeight)
		} else {
			style = blurredColumnStyle.Width(colWidth).Height(colHeight)
		}

		cols = append(cols, style.Render(a.columns[i].View()))
	}

	// Join columns horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}

// renderStatusBar renders the bottom status bar with help hints.
func (a *App) renderStatusBar() string {
	// Left side: status message or help hints
	var left string
	if a.statusMessage != "" {
		style := statusBarStyle
		if a.statusIsError {
			style = style.Foreground(lipgloss.Color("196"))
		}
		left = style.Render(a.statusMessage)
	} else {
		// Show key hints based on current view and selection state
		var hints []string
		switch a.view {
		case ViewBoard:
			// Check if in selection mode
			if a.selectionState != nil && a.selectionState.IsActive() {
				hints = []string{
					"space: toggle",
					"ctrl+a: select all",
					"esc: clear",
					"d: delete selected",
				}
			} else {
				hints = []string{
					"h/l: columns",
					"j/k: tasks",
					"space: select",
					"n: new",
					"e: edit",
					"d: delete",
					"b: boards",
					"ctrl+k: cmds",
					"q: quit",
				}
			}
		case ViewTaskDetail:
			hints = []string{
				"j/k: scroll",
				"e: edit",
				"esc: close",
			}
		case ViewTaskForm:
			hints = []string{
				"tab: next field",
				"ctrl+s: save",
				"esc: cancel",
			}
		case ViewConfirm:
			hints = []string{
				"y: confirm",
				"n/esc: cancel",
			}
		case ViewBoardSelector:
			hints = []string{
				"j/k: navigate",
				"/: filter",
				"enter: select",
				"esc/b: cancel",
			}
		}
		left = statusBarStyle.Render(strings.Join(hints, " | "))
	}

	// Right side: selection count or board info
	var right string
	if a.selectionState != nil && a.selectionState.IsActive() {
		// Show selection count
		right = RenderSelectionCount(a.selectionState.Count())
	} else if a.currentBoard != nil {
		right = statusBarStyle.Render(a.currentBoard.GetString("prefix"))
	}

	// Calculate padding to right-align the right side
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	gap := a.width - leftWidth - rightWidth
	if gap < 0 {
		gap = 0
	}

	return left + strings.Repeat(" ", gap) + right
}

// =============================================================================
// Data Helpers
// =============================================================================

// initializeColumns creates columns from loaded tasks.
func (a *App) initializeColumns(tasks []*core.Record) {
	// Group tasks by column
	tasksByColumn := make(map[string][]*core.Record)
	for _, task := range tasks {
		col := task.GetString("column")
		tasksByColumn[col] = append(tasksByColumn[col], task)
	}

	// Get board prefix for display IDs
	boardPrefix := ""
	if a.currentBoard != nil {
		boardPrefix = a.currentBoard.GetString("prefix")
	}

	// Create columns
	a.columns = make([]Column, len(a.columnOrder))
	for i, status := range a.columnOrder {
		items := a.recordsToListItems(tasksByColumn[status], boardPrefix)
		focused := i == a.focusedCol
		a.columns[i] = NewColumn(status, items, focused)
	}

	a.updateColumnSizes()
}

// updateColumnsWithTasks refreshes column data from task records.
func (a *App) updateColumnsWithTasks(tasks []*core.Record) {
	// Group tasks by column
	tasksByColumn := make(map[string][]*core.Record)
	for _, task := range tasks {
		col := task.GetString("column")
		tasksByColumn[col] = append(tasksByColumn[col], task)
	}

	// Get board prefix for display IDs
	boardPrefix := ""
	if a.currentBoard != nil {
		boardPrefix = a.currentBoard.GetString("prefix")
	}

	// Update each column
	for i, status := range a.columnOrder {
		items := a.recordsToListItems(tasksByColumn[status], boardPrefix)
		a.columns[i].SetItems(items)
	}
}

// recordsToListItems converts PocketBase records to list items.
func (a *App) recordsToListItems(records []*core.Record, boardPrefix string) []list.Item {
	items := make([]list.Item, len(records))
	for i, record := range records {
		seq := record.GetInt("seq")
		displayID := board.FormatDisplayID(boardPrefix, seq)
		taskItem := NewTaskItemFromRecord(record, displayID)
		// Resolve epic information for badge display
		if taskItem.EpicID != "" && len(a.availableEpics) > 0 {
			for _, epic := range a.availableEpics {
				if epic.ID == taskItem.EpicID {
					taskItem.Epic = epic
					taskItem.EpicTitle = epic.Title
					break
				}
			}
		}
		// Apply subtask count if available
		if count, ok := a.subtaskCounts[taskItem.ID]; ok {
			taskItem.SubtaskCount = count
			taskItem.HasSubtasks = count > 0
			// Check if expanded
			if a.expandedSubtaskView != nil {
				taskItem.SubtasksExpanded = a.expandedSubtaskView.IsExpanded(taskItem.ID)
			}
		}
		// Apply selection state
		if a.selectionState != nil {
			taskItem.IsSelected = a.selectionState.IsSelected(taskItem.ID)
		}
		items[i] = taskItem
	}
	return items
}

// recordToTaskItem converts a single record to a TaskItem
func (a *App) recordToTaskItem(record *core.Record) TaskItem {
	boardPrefix := ""
	if a.currentBoard != nil {
		boardPrefix = a.currentBoard.GetString("prefix")
	}
	seq := record.GetInt("seq")
	displayID := board.FormatDisplayID(boardPrefix, seq)
	taskItem := NewTaskItemFromRecord(record, displayID)
	// Resolve epic information for badge display
	if taskItem.EpicID != "" && len(a.availableEpics) > 0 {
		for _, epic := range a.availableEpics {
			if epic.ID == taskItem.EpicID {
				taskItem.Epic = epic
				taskItem.EpicTitle = epic.Title
				break
			}
		}
	}
	return taskItem
}

// updateColumnSizes recalculates column dimensions after resize.
func (a *App) updateColumnSizes() {
	if len(a.columns) == 0 {
		return
	}

	// Available height for columns (minus header, filter bar, and status)
	colHeight := a.height - 3
	if a.filterState != nil && a.filterState.HasActiveFilters() {
		colHeight-- // Account for filter bar
	}
	if colHeight < 5 {
		colHeight = 5
	}

	// Calculate column width
	colWidth := (a.width - 2) / len(a.columns)
	if colWidth < 15 {
		colWidth = 15
	}

	for i := range a.columns {
		a.columns[i].SetSize(colWidth, colHeight)
	}
}

// refreshFilteredColumns triggers a reload of tasks with filters applied
// For simplicity, this just reloads tasks from the database
func (a *App) refreshFilteredColumns() {
	// The filtering is done at render time in View()
	// This method exists for consistency with the message pattern
	// and can trigger a reload if needed in the future
}

// refreshColumnEpics updates task items in all columns with resolved epic information.
// Called after epics are loaded to show epic badges on task cards.
func (a *App) refreshColumnEpics() {
	if len(a.availableEpics) == 0 {
		return
	}

	// Build epic lookup map
	epicMap := make(map[string]EpicOption, len(a.availableEpics))
	for _, epic := range a.availableEpics {
		epicMap[epic.ID] = epic
	}

	// Update each column's items
	for colIdx := range a.columns {
		items := a.columns[colIdx].Items()
		for itemIdx, item := range items {
			if task, ok := item.(TaskItem); ok && task.EpicID != "" {
				if epic, found := epicMap[task.EpicID]; found {
					task.Epic = epic
					task.EpicTitle = epic.Title
					items[itemIdx] = task
				}
			}
		}
		a.columns[colIdx].SetItems(items)
	}
}

// refreshColumnSubtaskCounts updates task items in all columns with subtask counts.
// Called after subtask counts are loaded to show [+N] indicators on task cards.
func (a *App) refreshColumnSubtaskCounts() {
	if len(a.subtaskCounts) == 0 {
		return
	}

	// Update each column's items
	for colIdx := range a.columns {
		items := a.columns[colIdx].Items()
		for itemIdx, item := range items {
			if task, ok := item.(TaskItem); ok {
				if count, found := a.subtaskCounts[task.ID]; found {
					task.SubtaskCount = count
					task.HasSubtasks = count > 0
					if a.expandedSubtaskView != nil {
						task.SubtasksExpanded = a.expandedSubtaskView.IsExpanded(task.ID)
					}
					items[itemIdx] = task
				}
			}
		}
		a.columns[colIdx].SetItems(items)
	}
}

// getTotalTaskCount returns the total number of tasks across all columns (unfiltered).
func (a *App) getTotalTaskCount() int {
	count := 0
	for _, col := range a.columns {
		count += col.TotalItemCount()
	}
	return count
}

// getFilteredTaskCount returns the number of tasks that match the current filters.
func (a *App) getFilteredTaskCount() int {
	if a.filterState == nil || !a.filterState.HasActiveFilters() {
		return a.getTotalTaskCount()
	}

	count := 0
	for _, col := range a.columns {
		// Get all items and apply filter
		items := col.AllTaskItems()
		filtered := a.filterState.Apply(items)
		count += len(filtered)
	}
	return count
}

// getFilterSummary returns a brief summary of active filters for display in the header.
func (a *App) getFilterSummary() string {
	if a.filterState == nil {
		return ""
	}

	var parts []string

	// Add search query if present
	if q := a.filterState.GetSearchQuery(); q != "" {
		parts = append(parts, fmt.Sprintf("search: %q", Truncate(q, 15)))
	}

	// Add filter descriptions (limit to 3 for space)
	filters := a.filterState.GetFilters()
	for i, f := range filters {
		if len(parts) >= 3 {
			remaining := len(filters) - i
			if a.filterState.GetSearchQuery() != "" {
				remaining = len(filters) - i
			}
			parts = append(parts, fmt.Sprintf("+%d more", remaining))
			break
		}
		parts = append(parts, f.String())
	}

	return strings.Join(parts, ", ")
}

// saveSession saves the current session state before quitting.
// Errors are silently ignored since session persistence is optional.
func (a *App) saveSession() {
	currentBoardID := ""
	if a.currentBoard != nil {
		currentBoardID = a.currentBoard.Id
	}

	_ = SaveSession(SessionState{
		FilterState:    a.filterState,
		CurrentBoardID: currentBoardID,
		FocusedColumn:  a.focusedCol,
	})
}

// =============================================================================
// Realtime Event Handlers
// =============================================================================

// handleRealtimeEvent processes a realtime event and returns appropriate commands.
func (a *App) handleRealtimeEvent(event RealtimeEvent) tea.Cmd {
	// Only handle task events for the current board
	if event.Collection != "tasks" {
		return nil
	}

	// Check if this event is for the current board
	boardID, ok := event.Record["board"].(string)
	if !ok || (a.currentBoard != nil && boardID != a.currentBoard.Id) {
		return nil
	}

	switch event.Action {
	case "create":
		return a.handleTaskCreated(event.Record)
	case "update":
		return a.handleTaskUpdated(event.Record)
	case "delete":
		return a.handleTaskDeleted(event.Record)
	}

	return nil
}

// handleTaskCreated handles a task creation event from realtime.
// Uses incremental update to insert the new task without reloading all tasks.
func (a *App) handleTaskCreated(record map[string]interface{}) tea.Cmd {
	if a.currentBoard == nil {
		return nil
	}

	// Convert map to TaskItem
	boardPrefix := a.currentBoard.GetString("prefix")
	task := NewTaskItemFromMap(record, boardPrefix)

	// Find the target column
	colIndex := a.columnIndexForStatus(task.Column)
	if colIndex < 0 || colIndex >= len(a.columns) {
		return nil
	}

	// Check if task already exists (avoid duplicates on reconnect)
	if existingIndex := a.columns[colIndex].FindTaskByID(task.ID); existingIndex >= 0 {
		// Already exists, treat as update
		a.columns[colIndex].UpdateTask(existingIndex, task)
		return showStatus("Task updated externally", false, 2*time.Second)
	}

	// Insert the new task
	a.columns[colIndex].InsertTask(task)

	return showStatus("Task created externally", false, 2*time.Second)
}

// handleTaskUpdated handles a task update event from realtime.
// Uses incremental update to modify the task in place or move it between columns.
func (a *App) handleTaskUpdated(record map[string]interface{}) tea.Cmd {
	if a.currentBoard == nil {
		return nil
	}

	// Convert map to TaskItem
	boardPrefix := a.currentBoard.GetString("prefix")
	task := NewTaskItemFromMap(record, boardPrefix)

	// Find current location of the task
	oldColIndex, oldItemIndex := a.findTaskInAllColumns(task.ID)

	// Find target column
	newColIndex := a.columnIndexForStatus(task.Column)
	if newColIndex < 0 || newColIndex >= len(a.columns) {
		return nil
	}

	if oldColIndex < 0 {
		// Task not found, treat as create
		a.columns[newColIndex].InsertTask(task)
		return showStatus("Task created externally", false, 2*time.Second)
	}

	if oldColIndex == newColIndex {
		// Same column, update in place
		a.columns[oldColIndex].UpdateTask(oldItemIndex, task)
	} else {
		// Different column, remove from old and insert into new
		a.columns[oldColIndex].RemoveTask(oldItemIndex)
		a.columns[newColIndex].InsertTask(task)
	}

	return showStatus("Task updated externally", false, 2*time.Second)
}

// handleTaskDeleted handles a task deletion event from realtime.
// Uses incremental update to remove the task without reloading all tasks.
func (a *App) handleTaskDeleted(record map[string]interface{}) tea.Cmd {
	taskID, ok := record["id"].(string)
	if !ok {
		return nil
	}

	// Find and remove the task
	colIndex, itemIndex := a.findTaskInAllColumns(taskID)
	if colIndex >= 0 {
		a.columns[colIndex].RemoveTask(itemIndex)
		return showStatus("Task deleted externally", false, 2*time.Second)
	}

	return nil
}

// findTaskInAllColumns searches for a task by ID across all columns.
// Returns (columnIndex, itemIndex) or (-1, -1) if not found.
func (a *App) findTaskInAllColumns(taskID string) (colIndex, itemIndex int) {
	for i := range a.columns {
		if idx := a.columns[i].FindTaskByID(taskID); idx >= 0 {
			return i, idx
		}
	}
	return -1, -1
}

// columnIndexForStatus returns the column index for a given status string.
// Returns -1 if the status is not found in columnOrder.
func (a *App) columnIndexForStatus(status string) int {
	for i, s := range a.columnOrder {
		if s == status {
			return i
		}
	}
	return -1
}

// matchKey checks if a key message matches a binding.
func matchKey(msg tea.KeyMsg, binding key.Binding) bool {
	for _, k := range binding.Keys() {
		if msg.String() == k {
			return true
		}
	}
	return false
}

// Run starts the TUI application.
// This is the main entry point called from the CLI command.
func Run(pb *pocketbase.PocketBase, boardRef string) error {
	app := NewApp(pb, boardRef)

	// Create program with alt screen (full terminal takeover)
	p := tea.NewProgram(app, tea.WithAltScreen())

	// Run the program
	_, err := p.Run()
	return err
}
