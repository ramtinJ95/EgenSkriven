# Phase 3: Multi-Board Support

**Goal**: Board switching and management with persistent configuration

**Duration Estimate**: 2-3 days

**Prerequisites**: Phase 2 (Task Operations) completed - full CRUD for tasks working

**Deliverable**: Can switch between boards, displays board info in header, remembers last used board

---

## Overview

EgenSkriven supports multiple boards, each with its own tasks, prefix, and optionally custom columns. This phase adds the ability to:

1. **Switch between boards** using a modal selector (triggered by 'b' key)
2. **Display current board info** in the header (name, prefix, task count)
3. **Persist the last-used board** so it opens automatically next time
4. **Support board-specific columns** when configured

### Why Multi-Board Support?

Users often organize work across multiple contexts:
- "Work" board for professional tasks
- "Personal" board for personal projects
- "Learning" board for courses and study goals

Without board switching, users would need to restart the TUI with `--board` flag each time. With this phase, switching is instant and seamless.

### Key Concepts

| Concept | Description |
|---------|-------------|
| **Board** | A collection of tasks with a unique prefix (e.g., "WRK") |
| **Prefix** | Uppercase identifier used in display IDs (e.g., WRK-123) |
| **Columns** | Board-specific workflow stages (defaults to 6 standard columns) |
| **Default Board** | The board loaded on startup, persisted in config |

---

## Tasks

### 3.1 Create Board-Related Message Types ✅ COMPLETED

**What**: Define tea.Msg types for board operations.

**Why**: Bubble Tea uses messages to communicate between components and async operations. We need messages for loading boards, switching boards, and handling errors.

**File**: `internal/tui/messages.go`

**Steps**:

1. Add board-related message types to the existing messages.go file
2. Include all necessary fields for board data transfer
3. Define error message types for board operations

**Code**:

```go
// internal/tui/messages.go

package tui

import (
	"github.com/pocketbase/pocketbase/core"
)

// Board-related messages

// boardsLoadedMsg is sent when all boards have been loaded from the database
type boardsLoadedMsg struct {
	boards []*core.Record
}

// boardSwitchedMsg is sent when the user selects a different board
type boardSwitchedMsg struct {
	boardID string
}

// boardTasksLoadedMsg is sent when tasks for the current board are loaded
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

// errMsg represents an error from any async operation
type errMsg struct {
	err     error
	context string // Optional: describes what operation failed
}

func (e errMsg) Error() string {
	if e.context != "" {
		return e.context + ": " + e.err.Error()
	}
	return e.err.Error()
}
```

**Expected output**: No output - file compiles successfully.

**Common mistakes**:
- Forgetting to export message types (use uppercase first letter if needed externally)
- Not including enough context in error messages

---

### 3.2 Implement BoardSelector Component ✅ COMPLETED

**What**: Create a modal/overlay component for selecting boards using `bubbles/list`.

**Why**: The board selector provides a searchable, keyboard-navigable list of all available boards. It appears as an overlay when the user presses 'b'.

**File**: `internal/tui/board_selector.go`

**Steps**:

1. Create BoardOption struct implementing list.Item interface
2. Create BoardSelector struct wrapping bubbles/list
3. Implement Init, Update, View methods
4. Add filtering support for quick board search

**Code**:

```go
// internal/tui/board_selector.go

package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pocketbase/pocketbase/core"

	"github.com/ramtinJ95/EgenSkriven/internal/board"
)

// BoardOption represents a board in the selector list
type BoardOption struct {
	ID         string
	Name       string
	Prefix     string
	Color      string
	TaskCount  int
	Columns    []string
	IsSelected bool // Currently active board
}

// Implement list.Item interface for BoardOption

// FilterValue returns the string used for filtering/searching
func (b BoardOption) FilterValue() string {
	return b.Name + " " + b.Prefix
}

// Title returns the display title for the list item
func (b BoardOption) Title() string {
	prefix := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true).
		Render(b.Prefix)
	
	name := b.Name
	if b.IsSelected {
		name = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Render(name + " (current)")
	}
	
	return fmt.Sprintf("%s - %s", prefix, name)
}

// Description returns additional info shown below the title
func (b BoardOption) Description() string {
	if b.TaskCount == 0 {
		return "No tasks"
	}
	if b.TaskCount == 1 {
		return "1 task"
	}
	return fmt.Sprintf("%d tasks", b.TaskCount)
}

// BoardSelector is the modal component for selecting boards
type BoardSelector struct {
	list          list.Model
	boards        []BoardOption
	currentBoard  string // ID of currently selected board
	width         int
	height        int
	keys          boardSelectorKeyMap
}

// boardSelectorKeyMap defines keybindings for the board selector
type boardSelectorKeyMap struct {
	Select key.Binding
	Cancel key.Binding
}

// defaultBoardSelectorKeys returns the default keybindings
func defaultBoardSelectorKeys() boardSelectorKeyMap {
	return boardSelectorKeyMap{
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select board"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

// NewBoardSelector creates a new board selector component
func NewBoardSelector(boards []BoardOption, currentBoardID string) *BoardSelector {
	// Create list items from board options
	items := make([]list.Item, len(boards))
	for i, b := range boards {
		b.IsSelected = (b.ID == currentBoardID)
		boards[i] = b
		items[i] = b
	}

	// Create delegate with custom styling
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.SetHeight(2)
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("62")).
		BorderLeftForeground(lipgloss.Color("62"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("240"))

	// Create list model
	l := list.New(items, delegate, 0, 0)
	l.Title = "Switch Board"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	// Find and select current board in list
	for i, b := range boards {
		if b.ID == currentBoardID {
			l.Select(i)
			break
		}
	}

	return &BoardSelector{
		list:         l,
		boards:       boards,
		currentBoard: currentBoardID,
		keys:         defaultBoardSelectorKeys(),
	}
}

// Init initializes the board selector (required by tea.Model interface pattern)
func (s *BoardSelector) Init() tea.Cmd {
	return nil
}

// Update handles messages for the board selector
func (s *BoardSelector) Update(msg tea.Msg) (*BoardSelector, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle selection
		if key.Matches(msg, s.keys.Select) {
			if item, ok := s.list.SelectedItem().(BoardOption); ok {
				return s, func() tea.Msg {
					return boardSwitchedMsg{boardID: item.ID}
				}
			}
		}

		// Handle cancellation
		if key.Matches(msg, s.keys.Cancel) {
			// Return nil command to signal cancellation
			return s, nil
		}
	}

	// Delegate to list for navigation and filtering
	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return s, cmd
}

// View renders the board selector
func (s *BoardSelector) View() string {
	// Modal container style
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		Width(s.width).
		Height(s.height)

	return modalStyle.Render(s.list.View())
}

// SetSize updates the dimensions of the board selector
func (s *BoardSelector) SetSize(width, height int) {
	s.width = width
	s.height = height
	
	// Account for modal padding and border
	listWidth := width - 6
	listHeight := height - 4
	
	s.list.SetSize(listWidth, listHeight)
}

// SelectedBoard returns the currently highlighted board option
func (s *BoardSelector) SelectedBoard() (BoardOption, bool) {
	if item, ok := s.list.SelectedItem().(BoardOption); ok {
		return item, true
	}
	return BoardOption{}, false
}

// BoardOptionsFromRecords converts PocketBase records to BoardOption slice
func BoardOptionsFromRecords(records []*core.Record, taskCounts map[string]int) []BoardOption {
	options := make([]BoardOption, len(records))
	for i, record := range records {
		b := board.RecordToBoard(record)
		count := 0
		if taskCounts != nil {
			count = taskCounts[record.Id]
		}
		options[i] = BoardOption{
			ID:        record.Id,
			Name:      b.Name,
			Prefix:    b.Prefix,
			Color:     b.Color,
			TaskCount: count,
			Columns:   b.Columns,
		}
	}
	return options
}
```

**Expected output**: Component compiles and can be instantiated.

**Common mistakes**:
- Not implementing list.Item interface correctly (FilterValue, Title, Description)
- Forgetting to set proper list dimensions
- Not handling keyboard events for selection/cancellation

---

### 3.3 Implement Board Header Component ✅ COMPLETED

**What**: Create a header component that displays the current board's name, prefix, and task count.

**Why**: Users need visual confirmation of which board they're viewing. The header provides context at a glance.

**File**: `internal/tui/header.go`

**Steps**:

1. Create Header struct with board info fields
2. Implement View method with styled output
3. Include task count per column or total
4. Add visual indicator for board color (if set)

**Code**:

```go
// internal/tui/header.go

package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Header displays board information at the top of the TUI
type Header struct {
	boardName   string
	boardPrefix string
	boardColor  string
	taskCount   int
	columnCount int
	width       int
	showHelp    bool // Show keybinding hints
}

// NewHeader creates a new header component
func NewHeader() *Header {
	return &Header{
		showHelp: true,
	}
}

// SetBoard updates the header with board information
func (h *Header) SetBoard(name, prefix, color string) {
	h.boardName = name
	h.boardPrefix = prefix
	h.boardColor = color
}

// SetTaskCount updates the task count display
func (h *Header) SetTaskCount(count int) {
	h.taskCount = count
}

// SetColumnCount updates the column count
func (h *Header) SetColumnCount(count int) {
	h.columnCount = count
}

// SetWidth updates the header width
func (h *Header) SetWidth(width int) {
	h.width = width
}

// SetShowHelp toggles help hint visibility
func (h *Header) SetShowHelp(show bool) {
	h.showHelp = show
}

// View renders the header
func (h *Header) View() string {
	if h.boardName == "" {
		return ""
	}

	// Base styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		MarginBottom(1)

	// Board prefix with color
	prefixColor := lipgloss.Color("39") // Default cyan
	if h.boardColor != "" {
		prefixColor = lipgloss.Color(h.boardColor)
	}
	prefix := lipgloss.NewStyle().
		Foreground(prefixColor).
		Bold(true).
		Render("[" + h.boardPrefix + "]")

	// Board name
	name := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Bold(true).
		Render(h.boardName)

	// Task count
	countStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
	
	var countText string
	if h.taskCount == 0 {
		countText = "No tasks"
	} else if h.taskCount == 1 {
		countText = "1 task"
	} else {
		countText = fmt.Sprintf("%d tasks", h.taskCount)
	}
	count := countStyle.Render(countText)

	// Left side: board info
	leftContent := fmt.Sprintf("%s %s  %s", prefix, name, count)

	// Right side: keybinding hints (if enabled)
	var rightContent string
	if h.showHelp {
		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Faint(true)
		rightContent = helpStyle.Render("b:switch board  ?:help  q:quit")
	}

	// Calculate spacing
	leftLen := lipgloss.Width(leftContent)
	rightLen := lipgloss.Width(rightContent)
	spacerLen := h.width - leftLen - rightLen - 4 // Account for padding
	if spacerLen < 1 {
		spacerLen = 1
	}

	// Build the header line
	spacer := lipgloss.NewStyle().Width(spacerLen).Render("")
	content := leftContent + spacer + rightContent

	// Add bottom border
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color("240")).
		Width(h.width)

	return headerStyle.Render(borderStyle.Render(content))
}

// Height returns the height of the header in lines
func (h *Header) Height() int {
	return 2 // Header line + border
}
```

**Expected output**: Header renders with board name, prefix, and task count.

**Common mistakes**:
- Not accounting for terminal width when calculating spacing
- Hardcoding colors instead of using board's color setting
- Not handling empty board name gracefully

---

### 3.4 Implement Board Loading Command ✅ COMPLETED

**What**: Create tea.Cmd functions to load boards from the database.

**Why**: Bubble Tea uses commands for async operations. Loading boards should happen asynchronously to avoid blocking the UI.

**File**: `internal/tui/commands.go`

**Steps**:

1. Create loadBoards command to fetch all boards
2. Create loadBoardTasks command to fetch tasks for a specific board
3. Create getTaskCounts helper for board selector
4. Handle errors appropriately

**Code**:

```go
// internal/tui/commands.go

package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"github.com/ramtinJ95/EgenSkriven/internal/board"
	"github.com/ramtinJ95/EgenSkriven/internal/config"
)

// loadBoards returns a command that loads all boards from the database
func loadBoards(app *pocketbase.PocketBase) tea.Cmd {
	return func() tea.Msg {
		records, err := board.GetAll(app)
		if err != nil {
			return errMsg{err: err, context: "loading boards"}
		}
		return boardsLoadedMsg{boards: records}
	}
}

// loadBoardTasks returns a command that loads tasks for a specific board
func loadBoardTasks(app *pocketbase.PocketBase, boardID string) tea.Cmd {
	return func() tea.Msg {
		records, err := app.FindAllRecords("tasks",
			dbx.NewExp("board = {:board}", dbx.Params{"board": boardID}),
		)
		if err != nil {
			return errMsg{err: err, context: "loading tasks"}
		}
		return boardTasksLoadedMsg{tasks: records}
	}
}

// loadBoardColumns returns a command that loads columns for a specific board
func loadBoardColumns(app *pocketbase.PocketBase, boardID string) tea.Cmd {
	return func() tea.Msg {
		record, err := app.FindRecordById("boards", boardID)
		if err != nil {
			return errMsg{err: err, context: "loading board columns"}
		}

		b := board.RecordToBoard(record)
		return boardColumnsMsg{columns: b.Columns}
	}
}

// getTaskCountsForBoards returns task counts per board
// This is a sync function used during board selector initialization
func getTaskCountsForBoards(app *pocketbase.PocketBase, boardIDs []string) map[string]int {
	counts := make(map[string]int)
	
	for _, boardID := range boardIDs {
		records, err := app.FindAllRecords("tasks",
			dbx.NewExp("board = {:board}", dbx.Params{"board": boardID}),
		)
		if err == nil {
			counts[boardID] = len(records)
		}
	}
	
	return counts
}

// saveLastBoard returns a command that persists the last-used board to config
func saveLastBoard(boardID string) tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.LoadProjectConfig()
		if err != nil {
			// If we can't load, create a new config
			cfg = config.DefaultConfig()
		}
		
		cfg.DefaultBoard = boardID
		
		if err := config.SaveConfig(".", cfg); err != nil {
			return errMsg{err: err, context: "saving last board"}
		}
		
		return lastBoardSavedMsg{boardID: boardID}
	}
}

// loadDefaultBoard returns a command that loads the default board from config
func loadDefaultBoard() tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.LoadProjectConfig()
		if err != nil || cfg.DefaultBoard == "" {
			// No default board configured
			return nil
		}
		return boardSwitchedMsg{boardID: cfg.DefaultBoard}
	}
}

// switchBoard returns a command sequence for switching to a new board
// It loads the board's tasks and columns, then saves it as the last-used board
func switchBoard(app *pocketbase.PocketBase, boardID string) tea.Cmd {
	return tea.Batch(
		loadBoardTasks(app, boardID),
		loadBoardColumns(app, boardID),
		saveLastBoard(boardID),
	)
}

// findBoardByRef finds a board by name, prefix, or ID
func findBoardByRef(app *pocketbase.PocketBase, ref string) tea.Cmd {
	return func() tea.Msg {
		record, err := board.GetByNameOrPrefix(app, ref)
		if err != nil {
			return errMsg{err: err, context: "finding board"}
		}
		return boardSwitchedMsg{boardID: record.Id}
	}
}

// BoardData holds computed board information for display
type BoardData struct {
	ID         string
	Name       string
	Prefix     string
	Color      string
	Columns    []string
	TaskCount  int
	Tasks      []*core.Record
	TasksByCol map[string][]*core.Record
}

// computeBoardData processes raw records into display-ready data
func computeBoardData(boardRecord *core.Record, tasks []*core.Record) *BoardData {
	b := board.RecordToBoard(boardRecord)
	
	// Group tasks by column
	tasksByCol := make(map[string][]*core.Record)
	for _, col := range b.Columns {
		tasksByCol[col] = []*core.Record{}
	}
	
	for _, task := range tasks {
		col := task.GetString("column")
		if col == "" {
			col = "backlog" // Default column
		}
		tasksByCol[col] = append(tasksByCol[col], task)
	}
	
	return &BoardData{
		ID:         boardRecord.Id,
		Name:       b.Name,
		Prefix:     b.Prefix,
		Color:      b.Color,
		Columns:    b.Columns,
		TaskCount:  len(tasks),
		Tasks:      tasks,
		TasksByCol: tasksByCol,
	}
}
```

**Expected output**: Commands execute without errors and return appropriate messages.

**Common mistakes**:
- Not handling database errors
- Blocking the UI with synchronous operations
- Not batching related commands together

---

### 3.5 Update App Model for Board Support ✅ COMPLETED

**What**: Modify the main App model to support board switching and management.

**Why**: The App model is the central state container. It needs to track the current board, handle board-related messages, and coordinate board switching.

**File**: `internal/tui/app.go`

**Steps**:

1. Add board-related fields to App struct
2. Update Init to load boards and default board
3. Add message handlers for board operations
4. Integrate BoardSelector as an overlay
5. Update View to show header and handle selector overlay

**Code**:

```go
// internal/tui/app.go

package tui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
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
	ViewHelp
	ViewBoardSelector
)

// App is the main TUI application model
type App struct {
	// PocketBase instance
	pb *pocketbase.PocketBase

	// Board state
	boards        []*core.Record     // All available boards
	currentBoard  *core.Record       // Currently active board
	boardData     *BoardData         // Computed data for current board
	columns       []string           // Current board's columns
	initialBoard  string             // Board to load on startup (from --board flag)

	// UI components
	header        *Header
	boardSelector *BoardSelector
	help          help.Model
	keys          keyMap

	// View state
	view       ViewState
	focusedCol int

	// Dimensions
	width  int
	height int
	ready  bool

	// Error state
	lastError error
}

// NewApp creates a new TUI application
func NewApp(pb *pocketbase.PocketBase, initialBoard string) *App {
	return &App{
		pb:           pb,
		header:       NewHeader(),
		help:         help.New(),
		keys:         defaultKeyMap(),
		view:         ViewBoard,
		focusedCol:   0,
		initialBoard: initialBoard,
	}
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	// Load all boards first, then determine which one to show
	return loadBoards(a.pb)
}

// Update handles messages and updates the model
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return a.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.ready = true
		a.header.SetWidth(msg.Width)
		if a.boardSelector != nil {
			a.boardSelector.SetSize(min(50, msg.Width-4), min(20, msg.Height-4))
		}
		return a, nil

	case boardsLoadedMsg:
		a.boards = msg.boards
		
		// Determine which board to show
		var targetBoardID string
		
		// Priority 1: Board from --board flag
		if a.initialBoard != "" {
			if record, err := board.GetByNameOrPrefix(a.pb, a.initialBoard); err == nil {
				targetBoardID = record.Id
			}
		}
		
		// Priority 2: Default board from config
		if targetBoardID == "" {
			cmds = append(cmds, loadDefaultBoard())
		}
		
		// Priority 3: First board in list
		if targetBoardID == "" && len(a.boards) > 0 {
			targetBoardID = a.boards[0].Id
		}
		
		if targetBoardID != "" {
			// Find the board record
			for _, b := range a.boards {
				if b.Id == targetBoardID {
					a.currentBoard = b
					break
				}
			}
			cmds = append(cmds, switchBoard(a.pb, targetBoardID))
		}
		
		return a, tea.Batch(cmds...)

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
		
		// Load the new board's data
		return a, switchBoard(a.pb, msg.boardID)

	case boardTasksLoadedMsg:
		if a.currentBoard != nil {
			a.boardData = computeBoardData(a.currentBoard, msg.tasks)
			a.updateHeader()
		}
		return a, nil

	case boardColumnsMsg:
		a.columns = msg.columns
		return a, nil

	case lastBoardSavedMsg:
		// Silently acknowledge save
		return a, nil

	case errMsg:
		a.lastError = msg.err
		return a, nil
	}

	// Handle component updates based on view
	switch a.view {
	case ViewBoardSelector:
		if a.boardSelector != nil {
			var cmd tea.Cmd
			a.boardSelector, cmd = a.boardSelector.Update(msg)
			return a, cmd
		}
	}

	return a, nil
}

// handleKeyMsg processes keyboard input
func (a *App) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys work in any view
	switch {
	case key.Matches(msg, a.keys.Quit):
		return a, tea.Quit

	case key.Matches(msg, a.keys.Help):
		if a.view == ViewHelp {
			a.view = ViewBoard
		} else {
			a.view = ViewHelp
		}
		return a, nil

	case key.Matches(msg, a.keys.Board):
		if a.view == ViewBoardSelector {
			// Close selector
			a.view = ViewBoard
			a.boardSelector = nil
		} else {
			// Open selector
			a.openBoardSelector()
		}
		return a, nil

	case key.Matches(msg, a.keys.Escape):
		// Close any overlay
		switch a.view {
		case ViewBoardSelector:
			a.view = ViewBoard
			a.boardSelector = nil
		case ViewHelp:
			a.view = ViewBoard
		case ViewTaskDetail, ViewTaskForm:
			a.view = ViewBoard
		}
		return a, nil
	}

	// View-specific keys
	switch a.view {
	case ViewBoardSelector:
		if a.boardSelector != nil {
			var cmd tea.Cmd
			a.boardSelector, cmd = a.boardSelector.Update(msg)
			return a, cmd
		}

	case ViewBoard:
		return a.handleBoardKeys(msg)
	}

	return a, nil
}

// handleBoardKeys processes keys when in board view
func (a *App) handleBoardKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.Left):
		if a.focusedCol > 0 {
			a.focusedCol--
		}
	case key.Matches(msg, a.keys.Right):
		if a.focusedCol < len(a.columns)-1 {
			a.focusedCol++
		}
	}
	return a, nil
}

// openBoardSelector initializes and opens the board selector
func (a *App) openBoardSelector() {
	if len(a.boards) == 0 {
		return
	}

	// Get task counts for each board
	boardIDs := make([]string, len(a.boards))
	for i, b := range a.boards {
		boardIDs[i] = b.Id
	}
	taskCounts := getTaskCountsForBoards(a.pb, boardIDs)

	// Create board options
	options := BoardOptionsFromRecords(a.boards, taskCounts)

	// Get current board ID
	currentBoardID := ""
	if a.currentBoard != nil {
		currentBoardID = a.currentBoard.Id
	}

	// Create selector
	a.boardSelector = NewBoardSelector(options, currentBoardID)
	a.boardSelector.SetSize(min(50, a.width-4), min(20, a.height-4))
	a.view = ViewBoardSelector
}

// updateHeader updates the header with current board info
func (a *App) updateHeader() {
	if a.boardData == nil {
		return
	}
	a.header.SetBoard(a.boardData.Name, a.boardData.Prefix, a.boardData.Color)
	a.header.SetTaskCount(a.boardData.TaskCount)
}

// View renders the TUI
func (a *App) View() string {
	if !a.ready {
		return "Loading..."
	}

	var content string

	// Render base content (header + board)
	headerView := a.header.View()
	headerHeight := a.header.Height()

	// Board content area
	boardHeight := a.height - headerHeight - 1 // -1 for status bar
	boardView := a.renderBoardView(boardHeight)

	content = lipgloss.JoinVertical(lipgloss.Left, headerView, boardView)

	// Render overlay if active
	switch a.view {
	case ViewBoardSelector:
		if a.boardSelector != nil {
			overlay := a.boardSelector.View()
			content = a.renderOverlay(content, overlay)
		}
	case ViewHelp:
		overlay := a.renderHelpView()
		content = a.renderOverlay(content, overlay)
	}

	return content
}

// renderBoardView renders the kanban columns
func (a *App) renderBoardView(height int) string {
	if a.boardData == nil || len(a.columns) == 0 {
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(2)
		return style.Render("No board selected. Press 'b' to select a board.")
	}

	// Calculate column width
	colWidth := (a.width - 2) / len(a.columns)
	if colWidth < 20 {
		colWidth = 20
	}

	// Render each column
	cols := make([]string, len(a.columns))
	for i, colName := range a.columns {
		focused := i == a.focusedCol
		tasks := a.boardData.TasksByCol[colName]
		cols[i] = a.renderColumn(colName, tasks, colWidth, height, focused)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}

// renderColumn renders a single kanban column
func (a *App) renderColumn(name string, tasks []*core.Record, width, height int, focused bool) string {
	// Column style
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(0, 1)

	if focused {
		style = style.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))
	} else {
		style = style.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))
	}

	// Column header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(a.columnColor(name))
	header := headerStyle.Render(formatColumnName(name) + " (" + formatCount(len(tasks)) + ")")

	// Task list (simplified for this phase)
	var taskLines []string
	for _, task := range tasks {
		title := task.GetString("title")
		if len(title) > width-4 {
			title = title[:width-7] + "..."
		}
		taskLines = append(taskLines, "  "+title)
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		append([]string{header, ""}, taskLines...)...,
	)

	return style.Render(content)
}

// renderOverlay renders content with an overlay on top
func (a *App) renderOverlay(base, overlay string) string {
	// Center the overlay
	overlayWidth := lipgloss.Width(overlay)
	overlayHeight := lipgloss.Height(overlay)
	
	x := (a.width - overlayWidth) / 2
	y := (a.height - overlayHeight) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	// For simplicity, just return the overlay
	// A more sophisticated implementation would dim the background
	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, overlay)
}

// renderHelpView renders the help overlay
func (a *App) renderHelpView() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2)

	content := `Keyboard Shortcuts

Navigation:
  h/l, Left/Right  Move between columns
  j/k, Up/Down     Move between tasks
  Tab              Cycle columns

Boards:
  b                Switch board
  
Tasks:
  Enter            View task details
  n                New task
  e                Edit task
  d                Delete task

General:
  ?                Toggle help
  q, Ctrl+C        Quit`

	return style.Render(content)
}

// columnColor returns the color for a column name
func (a *App) columnColor(name string) lipgloss.Color {
	colors := map[string]lipgloss.Color{
		"backlog":     lipgloss.Color("240"),
		"todo":        lipgloss.Color("39"),
		"in_progress": lipgloss.Color("214"),
		"need_input":  lipgloss.Color("196"),
		"review":      lipgloss.Color("205"),
		"done":        lipgloss.Color("82"),
	}
	if c, ok := colors[name]; ok {
		return c
	}
	return lipgloss.Color("255")
}

// formatColumnName converts snake_case to Title Case
func formatColumnName(name string) string {
	switch name {
	case "backlog":
		return "Backlog"
	case "todo":
		return "Todo"
	case "in_progress":
		return "In Progress"
	case "need_input":
		return "Need Input"
	case "review":
		return "Review"
	case "done":
		return "Done"
	default:
		return name
	}
}

// formatCount returns a string representation of a count
func formatCount(n int) string {
	if n == 0 {
		return "0"
	}
	return fmt.Sprintf("%d", n)
}

// Helper for Go 1.21+
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Add missing import
import "fmt"
```

**Expected output**: App initializes, loads boards, and displays header with board info.

**Common mistakes**:
- Not handling the case when no boards exist
- Not properly updating header when board changes
- Memory leaks from not cleaning up board selector

---

### 3.6 Add KeyMap for Board Operations ✅ COMPLETED

**What**: Define keybindings for board operations.

**Why**: Consistent keybindings across the application. The keyMap struct centralizes all key definitions.

**File**: `internal/tui/keys.go`

**Steps**:

1. Add Board key binding to the keyMap struct
2. Update defaultKeyMap function
3. Add help text for board switching

**Code**:

```go
// internal/tui/keys.go

package tui

import (
	"github.com/charmbracelet/bubbles/key"
)

// keyMap defines all keybindings for the TUI
type keyMap struct {
	// Navigation
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding

	// Task actions
	Enter  key.Binding
	New    key.Binding
	Edit   key.Binding
	Delete key.Binding

	// Task movement
	MoveLeft  key.Binding
	MoveRight key.Binding
	MoveUp    key.Binding
	MoveDown  key.Binding

	// Global
	Quit   key.Binding
	Help   key.Binding
	Board  key.Binding
	Search key.Binding
	Escape key.Binding

	// Quick column jumps
	Column1 key.Binding
	Column2 key.Binding
	Column3 key.Binding
	Column4 key.Binding
	Column5 key.Binding
	Column6 key.Binding
}

// defaultKeyMap returns the default keybindings
func defaultKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("k/up", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("j/down", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("h/left", "previous column"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("l/right", "next column"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "view details"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new task"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit task"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete task"),
		),
		MoveLeft: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "move task left"),
		),
		MoveRight: key.NewBinding(
			key.WithKeys("L"),
			key.WithHelp("L", "move task right"),
		),
		MoveUp: key.NewBinding(
			key.WithKeys("K"),
			key.WithHelp("K", "move task up"),
		),
		MoveDown: key.NewBinding(
			key.WithKeys("J"),
			key.WithHelp("J", "move task down"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Board: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "switch board"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel/close"),
		),
		Column1: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "jump to column 1"),
		),
		Column2: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "jump to column 2"),
		),
		Column3: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "jump to column 3"),
		),
		Column4: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "jump to column 4"),
		),
		Column5: key.NewBinding(
			key.WithKeys("5"),
			key.WithHelp("5", "jump to column 5"),
		),
		Column6: key.NewBinding(
			key.WithKeys("6"),
			key.WithHelp("6", "jump to column 6"),
		),
	}
}

// ShortHelp returns keybindings for the short help view
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Board, k.Quit}
}

// FullHelp returns keybindings for the full help view
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Enter, k.New, k.Edit, k.Delete},
		{k.MoveLeft, k.MoveRight, k.MoveUp, k.MoveDown},
		{k.Help, k.Board, k.Search, k.Quit},
	}
}
```

**Expected output**: Key bindings work as expected.

**Common mistakes**:
- Conflicting keybindings
- Missing help text
- Not implementing help.KeyMap interface

---

### 3.7 Implement Board Persistence Tests

**What**: Write tests for board switching and persistence functionality.

**Why**: Tests ensure board switching works correctly and the last-used board is properly saved/loaded.

**File**: `internal/tui/board_test.go`

**Steps**:

1. Create test helpers for board operations
2. Test board selector creation and navigation
3. Test board switching messages
4. Test config persistence

**Code**:

```go
// internal/tui/board_test.go

package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ramtinJ95/EgenSkriven/internal/board"
	"github.com/ramtinJ95/EgenSkriven/internal/config"
	"github.com/ramtinJ95/EgenSkriven/internal/testutil"
)

func TestBoardOptionsFromRecords(t *testing.T) {
	app := testutil.NewTestApp(t)

	// Create boards collection
	testutil.EnsureBoardsCollection(t, app)

	// Create test boards
	b1, err := board.Create(app, board.CreateInput{
		Name:   "Work",
		Prefix: "WRK",
	})
	require.NoError(t, err)

	b2, err := board.Create(app, board.CreateInput{
		Name:   "Personal",
		Prefix: "PER",
		Color:  "#FF5733",
	})
	require.NoError(t, err)

	// Get records
	records, err := board.GetAll(app)
	require.NoError(t, err)
	assert.Len(t, records, 2)

	// Convert to options
	taskCounts := map[string]int{
		b1.ID: 5,
		b2.ID: 3,
	}
	options := BoardOptionsFromRecords(records, taskCounts)

	assert.Len(t, options, 2)

	// Find options by ID
	var workOpt, persOpt BoardOption
	for _, opt := range options {
		if opt.ID == b1.ID {
			workOpt = opt
		}
		if opt.ID == b2.ID {
			persOpt = opt
		}
	}

	// Verify Work board
	assert.Equal(t, "Work", workOpt.Name)
	assert.Equal(t, "WRK", workOpt.Prefix)
	assert.Equal(t, 5, workOpt.TaskCount)

	// Verify Personal board
	assert.Equal(t, "Personal", persOpt.Name)
	assert.Equal(t, "PER", persOpt.Prefix)
	assert.Equal(t, "#FF5733", persOpt.Color)
	assert.Equal(t, 3, persOpt.TaskCount)
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

func TestSaveLastBoard(t *testing.T) {
	// Create temp directory for config
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
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
	_, err := os.Stat(configPath)
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
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	// Reset config cache
	config.ResetGlobalConfigCache()

	// Create config with default board
	cfg := config.DefaultConfig()
	cfg.DefaultBoard = "my-default-board"
	err := config.SaveConfig(tmpDir, cfg)
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
	os.Chdir(tmpDir)
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
	header := NewHeader()
	header.SetWidth(80)
	header.SetBoard("Test", "TST", "")

	tests := []struct {
		count    int
		expected string
	}{
		{0, "No tasks"},
		{1, "1 task"},
		{100, "100 tasks"},
	}

	for _, tt := range tests {
		header.SetTaskCount(tt.count)
		view := header.View()
		assert.Contains(t, view, tt.expected)
	}
}

func TestComputeBoardData(t *testing.T) {
	app := testutil.NewTestApp(t)

	// Create collections
	testutil.EnsureBoardsCollection(t, app)
	testutil.EnsureTasksCollection(t, app)

	// Create board
	b, err := board.Create(app, board.CreateInput{
		Name:   "Test",
		Prefix: "TST",
	})
	require.NoError(t, err)

	// Create tasks in different columns
	testutil.CreateTestTask(t, app, b.ID, "Task 1", "todo")
	testutil.CreateTestTask(t, app, b.ID, "Task 2", "todo")
	testutil.CreateTestTask(t, app, b.ID, "Task 3", "in_progress")

	// Get board record and tasks
	boardRecord, err := app.FindRecordById("boards", b.ID)
	require.NoError(t, err)

	tasks, err := app.FindAllRecords("tasks")
	require.NoError(t, err)

	// Compute board data
	data := computeBoardData(boardRecord, tasks)

	assert.Equal(t, b.ID, data.ID)
	assert.Equal(t, "Test", data.Name)
	assert.Equal(t, "TST", data.Prefix)
	assert.Equal(t, 3, data.TaskCount)
	assert.Len(t, data.TasksByCol["todo"], 2)
	assert.Len(t, data.TasksByCol["in_progress"], 1)
}
```

**Expected output**: All tests pass.

**Common mistakes**:
- Not cleaning up temp directories
- Not resetting config cache between tests
- Missing test for edge cases (empty boards, no tasks)

---

## Verification Checklist

### Board Loading

- [ ] **Boards load on startup**
  ```bash
  egenskriven tui
  ```
  Should display first board (or default if configured).

- [ ] **Board info shows in header**
  Header should display: `[PREFIX] Board Name  X tasks`

- [ ] **Multiple boards available**
  Create at least 2 boards before testing selector.

### Board Switching

- [ ] **Board selector opens**
  Press 'b' key. Modal should appear with board list.

- [ ] **Board selector is searchable**
  Start typing board name. List should filter.

- [ ] **Can navigate board list**
  Use j/k or up/down arrows to move selection.

- [ ] **Can select board**
  Press Enter to switch to selected board.

- [ ] **Can cancel selector**
  Press Esc or q to close without switching.

- [ ] **Tasks reload on switch**
  After switching, tasks from new board should appear.

### Persistence

- [ ] **Last board is saved**
  ```bash
  cat .egenskriven/config.json
  ```
  Should contain `"default_board": "board-id"`.

- [ ] **Default board loads on startup**
  Close and reopen TUI. Same board should be selected.

- [ ] **--board flag overrides default**
  ```bash
  egenskriven tui --board OtherBoard
  ```
  Should open specified board, not default.

### Custom Columns

- [ ] **Board-specific columns display**
  Create board with custom columns:
  ```bash
  egenskriven board add "Custom" --prefix CUS --columns "draft,active,done"
  ```
  Switch to this board. Only 3 columns should appear.

### Tests

- [ ] **All tests pass**
  ```bash
  go test ./internal/tui/... -v
  ```

---

## File Summary

| File | Lines | Purpose |
|------|-------|---------|
| `internal/tui/messages.go` | ~50 | Message type definitions |
| `internal/tui/board_selector.go` | ~200 | Board selection overlay component |
| `internal/tui/header.go` | ~120 | Board header display component |
| `internal/tui/commands.go` | ~150 | Async command functions for board ops |
| `internal/tui/app.go` | ~400 | Updated main application model |
| `internal/tui/keys.go` | ~130 | Keybinding definitions |
| `internal/tui/board_test.go` | ~250 | Tests for board functionality |

**Total new/modified code**: ~1300 lines

---

## What You Should Have Now

After completing Phase 3, your TUI should:

```
internal/tui/
├── app.go              # Updated: board state, selector integration
├── board_selector.go   # New: board switching modal
├── board_test.go       # New: tests for board functionality
├── commands.go         # Updated: board loading commands
├── header.go           # New: board info header
├── keys.go             # Updated: board keybinding
├── messages.go         # Updated: board message types
├── styles.go           # Existing
├── column.go           # Existing from Phase 1
├── task_item.go        # Existing from Phase 1
├── task_detail.go      # Existing from Phase 2
└── task_form.go        # Existing from Phase 2
```

And in `.egenskriven/config.json`:
```json
{
  "agent": {
    "workflow": "light",
    "mode": "autonomous"
  },
  "default_board": "abc123def456"
}
```

---

## Next Phase

**Phase 4: Real-Time Sync** will add:
- SSE client for PocketBase real-time events
- Subscribe to tasks collection on startup
- Handle create/update/delete events
- Update UI without full refresh
- Connection status indicator
- Automatic reconnection

---

## Troubleshooting

### "No boards found" on startup

**Problem**: No boards exist in the database.

**Solution**: Create a board first:
```bash
egenskriven board add "Work" --prefix WRK
```

### Board selector doesn't open

**Problem**: 'b' key doesn't trigger selector.

**Solution**: Check you're in board view (not task detail or form). Press Esc first to ensure you're in base view.

### Config file not saved

**Problem**: `.egenskriven/config.json` not created or updated.

**Solution**: Check directory permissions:
```bash
ls -la .egenskriven/
# If missing:
mkdir -p .egenskriven
chmod 755 .egenskriven
```

### Wrong columns after switching boards

**Problem**: Old board's columns still showing.

**Solution**: Ensure `boardColumnsMsg` is handled correctly in Update. Check that columns are being set from the board record, not cached.

### Tests fail with "collection not found"

**Problem**: Test database doesn't have required collections.

**Solution**: Ensure `testutil.EnsureBoardsCollection` and `testutil.EnsureTasksCollection` are called in test setup.

### Board selector shows no task counts

**Problem**: All boards show "0 tasks" or "No tasks".

**Solution**: Verify `getTaskCountsForBoards` is being called with correct board IDs. Check that tasks have correct `board` field set.

---

## Glossary

| Term | Definition |
|------|------------|
| **tea.Cmd** | A function that returns a tea.Msg, used for async operations |
| **tea.Msg** | An event or message in Bubble Tea's event loop |
| **tea.Batch** | Combines multiple commands to run concurrently |
| **list.Item** | Interface for items in bubbles/list (FilterValue, Title, Description) |
| **Overlay** | A component rendered on top of the base content |
| **BoardData** | Computed display data for a board (tasks grouped by column) |
| **DefaultBoard** | The board ID saved in config to load on startup |
