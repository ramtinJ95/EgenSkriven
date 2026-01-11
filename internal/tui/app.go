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
)

// App is the main TUI application model.
// It implements tea.Model and manages the overall application state.
type App struct {
	// Core dependencies
	pb *pocketbase.PocketBase

	// Board state
	currentBoard *core.Record   // Currently displayed board
	boards       []*core.Record // All available boards

	// Column state - 6 columns for default workflow
	columns     []Column
	focusedCol  int
	columnOrder []string // Column status order

	// UI state
	width  int
	height int
	ready  bool // True once initial data is loaded
	view   ViewState

	// Overlays
	taskDetail    *TaskDetail
	taskForm      *TaskForm
	confirmDialog *ConfirmDialog

	// Pending operations
	pendingDeleteTaskID string

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
}

// NewApp creates a new TUI application.
// boardRef is optional - if empty, uses the first available board.
func NewApp(pb *pocketbase.PocketBase, boardRef string) *App {
	h := help.New()
	h.ShowAll = false

	return &App{
		pb:              pb,
		keys:            defaultKeyMap(),
		help:            h,
		focusedCol:      0,
		initialBoardRef: boardRef,
		columnOrder:     []string{"backlog", "todo", "in_progress", "need_input", "review", "done"},
		view:            ViewBoard,
	}
}

// Init implements tea.Model.
// Called once when the program starts. Returns initial commands.
func (a *App) Init() tea.Cmd {
	// Load board and tasks asynchronously
	return loadBoardAndTasks(a.pb, a.initialBoardRef)
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
		// Update overlay sizes
		if a.taskDetail != nil {
			a.taskDetail.SetSize(a.width/2, a.height-4)
		}
		if a.taskForm != nil {
			a.taskForm.SetSize(a.width/2, a.height-10)
		}
		return a, nil

	// =================================================================
	// Initial Load Messages
	// =================================================================

	case boardAndTasksLoadedMsg:
		a.currentBoard = msg.board
		a.initializeColumns(msg.tasks)
		a.ready = true
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
			if a.pendingDeleteTaskID != "" {
				cmds = append(cmds, deleteTask(a.pb, a.pendingDeleteTaskID))
				a.pendingDeleteTaskID = ""
			}
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
	// Keyboard Input
	// =================================================================

	case tea.KeyMsg:
		switch a.view {
		case ViewBoard:
			return a.handleBoardKeys(msg)
		case ViewTaskDetail:
			return a.handleDetailKeys(msg)
		case ViewTaskForm:
			return a.handleFormKeys(msg)
		case ViewConfirm:
			return a.handleConfirmKeys(msg)
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

	switch {
	// Quit
	case matchKey(msg, a.keys.Quit):
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
		a.help.ShowAll = !a.help.ShowAll
		return a, nil
	}

	// Additional key handling by string
	switch msg.String() {
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
		// Delete task (with confirmation)
		task := a.getSelectedTask()
		if task != nil {
			a.pendingDeleteTaskID = task.ID
			a.confirmDialog = NewDeleteConfirmDialog(task.TaskTitle)
			a.view = ViewConfirm
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
	}

	return a, nil
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

	default:
		// Normal board view
		sections = append(sections, a.renderColumns())
	}

	// Status bar
	sections = append(sections, a.renderStatusBar())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
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

// renderHeader renders the board title and info.
func (a *App) renderHeader() string {
	if a.currentBoard == nil {
		return ""
	}

	boardName := a.currentBoard.GetString("name")
	boardPrefix := a.currentBoard.GetString("prefix")

	title := fmt.Sprintf("EgenSkriven - %s (%s)", boardName, boardPrefix)
	return headerStyle.Width(a.width).Render(title)
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
		// Show key hints based on current view
		var hints []string
		switch a.view {
		case ViewBoard:
			hints = []string{
				"h/l: columns",
				"j/k: tasks",
				"n: new",
				"e: edit",
				"d: delete",
				"enter: details",
				"q: quit",
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
		}
		left = statusBarStyle.Render(strings.Join(hints, " | "))
	}

	// Right side: board info
	var right string
	if a.currentBoard != nil {
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
		items[i] = NewTaskItemFromRecord(record, displayID)
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
	return NewTaskItemFromRecord(record, displayID)
}

// updateColumnSizes recalculates column dimensions after resize.
func (a *App) updateColumnSizes() {
	if len(a.columns) == 0 {
		return
	}

	// Available height for columns (minus header and status)
	colHeight := a.height - 3
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
