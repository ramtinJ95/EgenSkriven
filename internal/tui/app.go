package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"github.com/ramtinJ95/EgenSkriven/internal/board"
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
	switch msg := msg.(type) {

	// Window resize
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.updateColumnSizes()
		return a, nil

	// Keyboard input
	case tea.KeyMsg:
		return a.handleKeyMsg(msg)

	// Initial load complete
	case boardAndTasksLoadedMsg:
		a.currentBoard = msg.board
		a.initializeColumns(msg.tasks)
		a.ready = true
		return a, nil

	// Tasks loaded (refresh)
	case tasksLoadedMsg:
		a.updateColumnsWithTasks(msg.tasks)
		return a, nil

	// Error occurred
	case errMsg:
		a.err = msg.err
		a.statusMessage = msg.Error()
		a.statusIsError = true
		return a, nil

	// Status message
	case statusMsg:
		a.statusMessage = msg.message
		a.statusIsError = msg.isError
		return a, nil
	}

	return a, nil
}

// handleKeyMsg processes keyboard input.
func (a *App) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

	// Enter - view task details (placeholder for Phase 2)
	case matchKey(msg, a.keys.Enter):
		task := a.columns[a.focusedCol].SelectedTask()
		if task != nil {
			a.statusMessage = fmt.Sprintf("Selected: %s %s", task.DisplayID, task.TaskTitle)
			a.statusIsError = false
		}
		return a, nil

	// Help toggle (placeholder for Phase 2)
	case matchKey(msg, a.keys.Help):
		a.help.ShowAll = !a.help.ShowAll
		return a, nil
	}

	return a, nil
}

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

	// Columns (main content)
	sections = append(sections, a.renderColumns())

	// Status bar
	sections = append(sections, a.renderStatusBar())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
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
	colWidth := (a.width - 2) / len(a.columns) // -2 for outer margins
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
		// Show key hints
		hints := []string{
			"h/l: columns",
			"j/k: tasks",
			"enter: details",
			"q: quit",
			"?: help",
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
