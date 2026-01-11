# Phase 7: Polish & Testing

**Goal**: Production-ready TUI with graceful error handling, loading states, performance optimization, and comprehensive test coverage.

**Duration Estimate**: 2-3 days

**Prerequisites**: Phase 6 (Advanced Features) completed, including help overlay, blocked tasks, due dates, and multi-select.

**Deliverable**: A polished, tested TUI ready for production use with:
- Graceful terminal resize handling
- Loading indicators for async operations
- Robust error handling and display
- Empty state displays for edge cases
- Performance optimizations for large boards
- Comprehensive unit and integration tests
- Updated CLI documentation and help text

---

## Overview

Phase 7 focuses on the final polish that transforms a functional TUI into a production-ready application. This phase addresses:

1. **Robustness**: Handle edge cases like empty boards, terminal resize, and network errors
2. **User Experience**: Provide visual feedback for loading states and errors
3. **Performance**: Optimize for boards with hundreds of tasks
4. **Quality Assurance**: Comprehensive testing to prevent regressions
5. **Documentation**: Integration with CLI help system

### Why This Matters

Users encounter edge cases in real-world usage:
- What happens when the terminal is resized mid-operation?
- How does the TUI behave with 500+ tasks?
- What feedback do users get when data is loading?
- What happens when there are no tasks at all?

A production-ready TUI handles all these gracefully, providing clear feedback and maintaining responsiveness.

### Key Principles

| Principle | Implementation |
|-----------|----------------|
| **Graceful Degradation** | Never crash; always show user-friendly errors |
| **Responsive Feedback** | Loading spinners, success/error messages |
| **Performance First** | Virtualization, lazy loading, pagination |
| **Test Coverage** | Unit tests for components, integration tests with real data |

---

## Tasks

### 7.1 Terminal Resize Handling

**What**: Implement graceful handling of terminal resize events to recalculate layout and update all components.

**Why**: Users frequently resize their terminal windows. Without proper handling, the TUI will render incorrectly or crash. The `tea.WindowSizeMsg` message is sent automatically by Bubble Tea when the terminal size changes.

**Steps**:

1. Create `internal/tui/resize.go` with resize handling logic
2. Update all components to respond to size changes
3. Implement minimum size constraints with warning display
4. Test with various terminal sizes

**Code** (`internal/tui/resize.go`):

```go
package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MinWidth is the minimum terminal width for proper rendering
const MinWidth = 60

// MinHeight is the minimum terminal height for proper rendering
const MinHeight = 15

// ResizeState tracks whether the terminal is too small
type ResizeState struct {
	Width       int
	Height      int
	TooSmall    bool
	MinWidth    int
	MinHeight   int
}

// NewResizeState creates a new resize state with default minimums
func NewResizeState() ResizeState {
	return ResizeState{
		MinWidth:  MinWidth,
		MinHeight: MinHeight,
	}
}

// Update handles window size messages
func (r *ResizeState) Update(msg tea.WindowSizeMsg) {
	r.Width = msg.Width
	r.Height = msg.Height
	r.TooSmall = msg.Width < r.MinWidth || msg.Height < r.MinHeight
}

// IsTooSmall returns true if the terminal is below minimum size
func (r *ResizeState) IsTooSmall() bool {
	return r.TooSmall
}

// TooSmallMessage returns a formatted message for small terminals
func (r *ResizeState) TooSmallMessage() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true).
		Align(lipgloss.Center)

	msg := fmt.Sprintf(
		"Terminal too small\n\n"+
			"Current: %dx%d\n"+
			"Minimum: %dx%d\n\n"+
			"Please resize your terminal",
		r.Width, r.Height,
		r.MinWidth, r.MinHeight,
	)

	return style.
		Width(r.Width).
		Height(r.Height).
		Render(msg)
}

// handleResize processes a window size message and updates all components
func (m *App) handleResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.resize.Update(msg)
	m.width = msg.Width
	m.height = msg.Height

	// Don't update components if terminal is too small
	if m.resize.IsTooSmall() {
		return m, nil
	}

	// Calculate available space for board
	// Account for: header (1), filter bar (1 if visible), status bar (1)
	headerHeight := 1
	filterHeight := 0
	if len(m.filters) > 0 || m.isSearching {
		filterHeight = 1
	}
	statusBarHeight := 1
	boardHeight := m.height - headerHeight - filterHeight - statusBarHeight

	// Update board dimensions
	if m.board != nil {
		m.board.SetSize(m.width, boardHeight)
	}

	// Update column widths
	// Each column gets equal width, minus borders/padding
	if len(m.columns) > 0 {
		// Calculate width per column (account for spacing between columns)
		spacing := len(m.columns) - 1 // 1 char between each column
		colWidth := (m.width - spacing) / len(m.columns)

		for i := range m.columns {
			m.columns[i].SetSize(colWidth, boardHeight-2) // -2 for column header/footer
		}
	}

	// Update overlay components if visible
	if m.taskDetail != nil {
		// Detail panel takes right 40% of screen
		detailWidth := m.width * 40 / 100
		if detailWidth < 40 {
			detailWidth = 40
		}
		m.taskDetail.SetSize(detailWidth, m.height-2)
	}

	if m.taskForm != nil {
		// Form takes center 60% of screen
		formWidth := m.width * 60 / 100
		if formWidth < 50 {
			formWidth = 50
		}
		formHeight := m.height * 70 / 100
		if formHeight < 20 {
			formHeight = 20
		}
		m.taskForm.SetSize(formWidth, formHeight)
	}

	if m.boardSelector != nil {
		// Board selector takes 40 columns
		m.boardSelector.SetSize(40, m.height/2)
	}

	return m, nil
}

// CalculateColumnWidth returns the width for each column
func CalculateColumnWidth(totalWidth int, numColumns int) int {
	if numColumns == 0 {
		return totalWidth
	}
	// Subtract 1 char per column gap (numColumns - 1)
	spacing := numColumns - 1
	return (totalWidth - spacing) / numColumns
}

// CalculateVisibleRows returns how many task rows can be displayed
func CalculateVisibleRows(columnHeight int, rowHeight int) int {
	if rowHeight == 0 {
		return 0
	}
	// Subtract header (title + count)
	available := columnHeight - 2
	if available < 0 {
		return 0
	}
	return available / rowHeight
}
```

**Expected output**:
- When terminal is resized, all columns adjust proportionally
- When terminal is too small, a warning message is shown
- No layout glitches during rapid resize events

**Common Mistakes**:
- Forgetting to update overlay components (detail, form) on resize
- Not handling zero-width or zero-height edge cases
- Calculating column widths without accounting for borders/spacing

---

### 7.2 Loading Indicators

**What**: Add loading spinner and status messages for async operations like task loading, saving, and board switching.

**Why**: Without visual feedback, users don't know if the app is working or frozen. Loading indicators provide confidence that operations are in progress.

**Steps**:

1. Create `internal/tui/loading.go` with loading state and spinner
2. Integrate spinner component from `bubbles/spinner`
3. Add loading messages for each operation type
4. Display loading state in status bar

**Code** (`internal/tui/loading.go`):

```go
package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LoadingState represents the current loading operation
type LoadingState int

const (
	LoadingIdle LoadingState = iota
	LoadingTasks
	LoadingBoards
	LoadingSaving
	LoadingDeleting
	LoadingMoving
	LoadingRefreshing
)

// loadingMessages maps loading states to user-friendly messages
var loadingMessages = map[LoadingState]string{
	LoadingIdle:       "",
	LoadingTasks:      "Loading tasks...",
	LoadingBoards:     "Loading boards...",
	LoadingSaving:     "Saving...",
	LoadingDeleting:   "Deleting...",
	LoadingMoving:     "Moving task...",
	LoadingRefreshing: "Refreshing...",
}

// LoadingIndicator wraps a spinner with loading state
type LoadingIndicator struct {
	state     LoadingState
	spinner   spinner.Model
	startTime time.Time
}

// NewLoadingIndicator creates a new loading indicator
func NewLoadingIndicator() LoadingIndicator {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	
	return LoadingIndicator{
		state:   LoadingIdle,
		spinner: s,
	}
}

// SetState changes the loading state
func (l *LoadingIndicator) SetState(state LoadingState) tea.Cmd {
	l.state = state
	if state != LoadingIdle {
		l.startTime = time.Now()
		return l.spinner.Tick
	}
	return nil
}

// IsLoading returns true if currently loading
func (l *LoadingIndicator) IsLoading() bool {
	return l.state != LoadingIdle
}

// State returns the current loading state
func (l *LoadingIndicator) State() LoadingState {
	return l.state
}

// Update handles spinner tick messages
func (l *LoadingIndicator) Update(msg tea.Msg) (LoadingIndicator, tea.Cmd) {
	if l.state == LoadingIdle {
		return *l, nil
	}

	var cmd tea.Cmd
	l.spinner, cmd = l.spinner.Update(msg)
	return *l, cmd
}

// View renders the loading indicator
func (l *LoadingIndicator) View() string {
	if l.state == LoadingIdle {
		return ""
	}

	message := loadingMessages[l.state]
	elapsed := time.Since(l.startTime)

	// Show elapsed time if loading takes more than 2 seconds
	elapsedStr := ""
	if elapsed > 2*time.Second {
		elapsedStr = " (" + elapsed.Truncate(time.Second).String() + ")"
	}

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205"))

	return style.Render(l.spinner.View() + " " + message + elapsedStr)
}

// loadingCmd is a message type for setting loading state
type loadingStateMsg struct {
	state LoadingState
}

// SetLoading returns a command that sets the loading state
func SetLoading(state LoadingState) tea.Cmd {
	return func() tea.Msg {
		return loadingStateMsg{state: state}
	}
}

// Integration with App model
// Add to App struct:
//   loading LoadingIndicator
//
// In App.Init():
//   m.loading = NewLoadingIndicator()
//
// In App.Update() for spinner ticks:
//   case spinner.TickMsg:
//       var cmd tea.Cmd
//       m.loading, cmd = m.loading.Update(msg)
//       return m, cmd
//
// In App.Update() for loading state changes:
//   case loadingStateMsg:
//       return m, m.loading.SetState(msg.state)
//
// In App.View() status bar:
//   if m.loading.IsLoading() {
//       statusContent = m.loading.View()
//   }
```

**Example Integration**:

```go
// When loading tasks
func (m *App) loadTasks() tea.Cmd {
	return tea.Batch(
		SetLoading(LoadingTasks),
		func() tea.Msg {
			tasks, err := m.fetchTasks()
			if err != nil {
				return errMsg{err}
			}
			return tasksLoadedMsg{tasks: tasks}
		},
	)
}

// When tasks are loaded
case tasksLoadedMsg:
	m.loading.SetState(LoadingIdle)
	m.populateColumns(msg.tasks)
	return m, nil
```

**Expected output**:
- Spinner appears during all async operations
- Loading message describes current operation
- Elapsed time shown for slow operations (>2s)
- Spinner stops when operation completes

**Common Mistakes**:
- Forgetting to set loading state to idle on success or error
- Not handling spinner tick messages causing animation to freeze
- Multiple concurrent operations overwriting each other's loading state

---

### 7.3 Error Handling and Display

**What**: Create a robust error display system with auto-dismiss and support for recoverable vs. fatal errors.

**Why**: Users need clear feedback when something goes wrong. Good error handling distinguishes between transient errors (retry) and fatal errors (exit).

**Steps**:

1. Create `internal/tui/error.go` with error display component
2. Implement auto-dismiss with configurable timeout
3. Support different error severity levels
4. Provide actionable error messages with suggestions

**Code** (`internal/tui/error.go`):

```go
package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ErrorSeverity indicates how serious an error is
type ErrorSeverity int

const (
	// ErrorInfo is informational (blue background)
	ErrorInfo ErrorSeverity = iota
	// ErrorWarning is a warning (yellow background)
	ErrorWarning
	// ErrorError is an error (red background)
	ErrorError
	// ErrorFatal is a fatal error requiring exit
	ErrorFatal
)

// Default error display duration
const DefaultErrorDuration = 5 * time.Second

// ErrorDisplay manages error message display
type ErrorDisplay struct {
	message    string
	severity   ErrorSeverity
	suggestion string
	timeout    time.Time
	visible    bool
}

// NewErrorDisplay creates a new error display
func NewErrorDisplay() ErrorDisplay {
	return ErrorDisplay{
		visible: false,
	}
}

// Show displays an error message
func (e *ErrorDisplay) Show(severity ErrorSeverity, message string) tea.Cmd {
	e.message = message
	e.severity = severity
	e.suggestion = ""
	e.timeout = time.Now().Add(DefaultErrorDuration)
	e.visible = true
	return e.startTimer()
}

// ShowWithSuggestion displays an error with a helpful suggestion
func (e *ErrorDisplay) ShowWithSuggestion(severity ErrorSeverity, message, suggestion string) tea.Cmd {
	e.message = message
	e.severity = severity
	e.suggestion = suggestion
	e.timeout = time.Now().Add(DefaultErrorDuration)
	e.visible = true
	return e.startTimer()
}

// ShowPersistent displays an error that won't auto-dismiss
func (e *ErrorDisplay) ShowPersistent(severity ErrorSeverity, message string) {
	e.message = message
	e.severity = severity
	e.suggestion = ""
	e.timeout = time.Time{} // Zero time = no timeout
	e.visible = true
}

// Dismiss hides the error display
func (e *ErrorDisplay) Dismiss() {
	e.visible = false
	e.message = ""
}

// IsVisible returns true if an error is being displayed
func (e *ErrorDisplay) IsVisible() bool {
	return e.visible
}

// IsFatal returns true if this is a fatal error
func (e *ErrorDisplay) IsFatal() bool {
	return e.severity == ErrorFatal
}

// Update handles tick messages for auto-dismiss
func (e *ErrorDisplay) Update(msg tea.Msg) (ErrorDisplay, tea.Cmd) {
	if !e.visible {
		return *e, nil
	}

	switch msg.(type) {
	case errorTimeoutMsg:
		if !e.timeout.IsZero() && time.Now().After(e.timeout) {
			e.Dismiss()
		}
	}
	return *e, nil
}

// View renders the error display
func (e *ErrorDisplay) View() string {
	if !e.visible {
		return ""
	}

	// Style based on severity
	var style lipgloss.Style
	switch e.severity {
	case ErrorInfo:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("27")).
			Padding(0, 1)
	case ErrorWarning:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("214")).
			Padding(0, 1)
	case ErrorError:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("196")).
			Padding(0, 1)
	case ErrorFatal:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("160")).
			Bold(true).
			Padding(0, 1)
	}

	content := e.message
	if e.suggestion != "" {
		content += " | Suggestion: " + e.suggestion
	}

	// Add dismiss hint for non-fatal persistent errors
	if e.timeout.IsZero() && e.severity != ErrorFatal {
		content += " (Press Esc to dismiss)"
	}

	return style.Render(content)
}

// Message types
type errorTimeoutMsg struct{}

func (e *ErrorDisplay) startTimer() tea.Cmd {
	if e.timeout.IsZero() {
		return nil
	}
	duration := time.Until(e.timeout)
	if duration < 0 {
		duration = 0
	}
	return tea.Tick(duration, func(t time.Time) tea.Msg {
		return errorTimeoutMsg{}
	})
}

// errMsg is the standard error message type from async operations
type errMsg struct {
	err error
}

// HandleError processes an error and returns appropriate commands
func HandleError(err error) tea.Cmd {
	return func() tea.Msg {
		return errMsg{err: err}
	}
}

// Error handling in App.Update():
//
// case errMsg:
//     // Determine severity based on error type
//     severity := ErrorError
//     suggestion := ""
//     
//     // Check for specific error types
//     if errors.Is(msg.err, ErrNotFound) {
//         severity = ErrorWarning
//         suggestion = "The task may have been deleted. Try refreshing."
//     } else if errors.Is(msg.err, ErrNetworkError) {
//         suggestion = "Check your connection and try again."
//     }
//     
//     // Stop loading and show error
//     m.loading.SetState(LoadingIdle)
//     return m, m.errorDisplay.Show(severity, msg.err.Error())
```

**Example Error Categories**:

```go
// Common error types for the TUI
var (
	ErrNotFound      = errors.New("not found")
	ErrNetworkError  = errors.New("network error")
	ErrDatabaseError = errors.New("database error")
	ErrValidation    = errors.New("validation error")
)

// Error messages for common scenarios
var errorSuggestions = map[string]string{
	"tasks collection not found": "Run 'egenskriven migrate' to initialize the database.",
	"database is locked":         "Close other egenskriven processes and try again.",
	"no tasks found":             "Create a task with 'n' or check your filters.",
}
```

**Expected output**:
- Errors appear in colored bar at top or bottom of screen
- Errors auto-dismiss after 5 seconds (configurable)
- Fatal errors show "Press q to quit" and disable other actions
- Suggestions help users resolve common issues

**Common Mistakes**:
- Not handling all error paths in async operations
- Leaving stale errors visible after recovery
- Not distinguishing between user errors and system errors
- Missing error handling for edge cases (empty results, timeouts)

---

### 7.4 Empty State Displays

**What**: Create informative empty state displays for boards without tasks, empty columns, and no search results.

**Why**: Empty states guide users on what to do next. Without them, users see blank screens and don't know if the app is broken or simply has no data.

**Steps**:

1. Create `internal/tui/empty.go` with empty state components
2. Implement context-aware messages (empty board vs. empty search)
3. Include actionable hints (keyboard shortcuts to add tasks)
4. Style consistently with the rest of the TUI

**Code** (`internal/tui/empty.go`):

```go
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// EmptyStateType identifies the type of empty state
type EmptyStateType int

const (
	EmptyBoard EmptyStateType = iota
	EmptyColumn
	EmptySearchResults
	EmptyFilterResults
	NoBoards
)

// EmptyState renders helpful messages when there's no data
type EmptyState struct {
	stateType EmptyStateType
	width     int
	height    int
}

// NewEmptyState creates a new empty state display
func NewEmptyState(stateType EmptyStateType) EmptyState {
	return EmptyState{
		stateType: stateType,
	}
}

// SetSize updates the dimensions
func (e *EmptyState) SetSize(width, height int) {
	e.width = width
	e.height = height
}

// View renders the empty state message
func (e *EmptyState) View() string {
	var title, subtitle, hint string

	switch e.stateType {
	case EmptyBoard:
		title = "No tasks yet"
		subtitle = "This board is empty"
		hint = "Press 'n' to create your first task"
	case EmptyColumn:
		title = "Empty"
		subtitle = ""
		hint = "Move tasks here with 'H' or 'L'"
	case EmptySearchResults:
		title = "No results"
		subtitle = "No tasks match your search"
		hint = "Press 'Esc' to clear search or try different terms"
	case EmptyFilterResults:
		title = "No matches"
		subtitle = "No tasks match the active filters"
		hint = "Press 'fc' to clear filters"
	case NoBoards:
		title = "No boards"
		subtitle = "No boards have been created yet"
		hint = "Create a board with 'egenskriven board add'"
	}

	// Styles
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Bold(true)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("62")).
		Italic(true)

	containerStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(e.width).
		Height(e.height)

	// Build content
	content := titleStyle.Render(title)
	if subtitle != "" {
		content += "\n" + subtitleStyle.Render(subtitle)
	}
	if hint != "" {
		content += "\n\n" + hintStyle.Render(hint)
	}

	return containerStyle.Render(content)
}

// renderEmptyBoard renders the full-board empty state
func (m *App) renderEmptyBoard() string {
	es := NewEmptyState(EmptyBoard)
	es.SetSize(m.width, m.height-3) // Account for header/status
	return es.View()
}

// renderEmptyColumn renders an empty column state
func (c *Column) renderEmpty() string {
	es := NewEmptyState(EmptyColumn)
	es.SetSize(c.width, c.height-2) // Account for column header
	return es.View()
}

// renderEmptySearchResults renders empty search results
func (m *App) renderEmptySearchResults() string {
	es := NewEmptyState(EmptySearchResults)
	es.SetSize(m.width, m.height-3)
	return es.View()
}

// renderEmptyFilterResults renders empty filter results
func (m *App) renderEmptyFilterResults() string {
	es := NewEmptyState(EmptyFilterResults)
	es.SetSize(m.width, m.height-3)
	return es.View()
}

// renderNoBoards renders the no boards state
func (m *App) renderNoBoards() string {
	es := NewEmptyState(NoBoards)
	es.SetSize(m.width, m.height-3)
	return es.View()
}

// isEmpty checks if the board has no tasks
func (m *App) isEmpty() bool {
	for _, col := range m.columns {
		if len(col.Items()) > 0 {
			return false
		}
	}
	return true
}

// isSearchEmpty checks if search returned no results
func (m *App) isSearchEmpty() bool {
	if m.searchQuery == "" {
		return false
	}
	for _, col := range m.columns {
		if len(col.FilteredItems()) > 0 {
			return false
		}
	}
	return true
}

// isFilterEmpty checks if filters returned no results
func (m *App) isFilterEmpty() bool {
	if len(m.filters) == 0 {
		return false
	}
	for _, col := range m.columns {
		if len(col.FilteredItems()) > 0 {
			return false
		}
	}
	return true
}
```

**Integration in Board View**:

```go
func (m *App) View() string {
	// Check for empty states first
	if len(m.boards) == 0 {
		return m.renderNoBoards()
	}
	
	if m.isEmpty() {
		return m.renderHeader() + "\n" + m.renderEmptyBoard() + "\n" + m.renderStatusBar()
	}
	
	if m.isSearchEmpty() {
		return m.renderHeader() + "\n" + m.renderEmptySearchResults() + "\n" + m.renderStatusBar()
	}
	
	if m.isFilterEmpty() {
		return m.renderHeader() + "\n" + m.renderEmptyFilterResults() + "\n" + m.renderStatusBar()
	}
	
	// Normal board rendering
	return m.renderBoard()
}
```

**Expected output**:
- Empty board shows "No tasks yet" with hint to press 'n'
- Empty search shows "No results" with hint to clear search
- Empty filter shows "No matches" with hint to clear filters
- All empty states are centered and styled consistently

**Common Mistakes**:
- Showing empty state during loading (check loading state first)
- Not distinguishing between "no data" and "no matching data"
- Forgetting to show empty state after filter/search clears results
- Inconsistent styling with the rest of the application

---

### 7.5 Performance Optimization

**What**: Optimize the TUI for large boards with hundreds of tasks using virtualization, lazy loading, and efficient updates.

**Why**: Without optimization, large boards cause lag, high memory usage, and poor user experience. The `bubbles/list` component has built-in virtualization, but we need to configure it properly.

**Steps**:

1. Create `internal/tui/performance.go` with optimization utilities
2. Configure list virtualization for large datasets
3. Implement lazy loading for done/backlog columns
4. Add pagination for very large columns
5. Optimize render cycles to avoid unnecessary work

**Code** (`internal/tui/performance.go`):

```go
package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// Performance thresholds
const (
	// MaxItemsPerColumn is the max items to load initially per column
	MaxItemsPerColumn = 100
	
	// LazyLoadThreshold is when to start lazy loading done/backlog
	LazyLoadThreshold = 50
	
	// VirtualizationThreshold enables virtualization when exceeded
	VirtualizationThreshold = 20
	
	// BatchSize for loading additional items
	BatchSize = 50
)

// PerformanceConfig holds performance tuning options
type PerformanceConfig struct {
	// EnableVirtualization uses list virtualization for long lists
	EnableVirtualization bool
	
	// LazyLoadInactive defers loading of done/backlog columns
	LazyLoadInactive bool
	
	// MaxItemsPerColumn limits initial load
	MaxItemsPerColumn int
	
	// BatchSize for loading more items
	BatchSize int
}

// DefaultPerformanceConfig returns sensible defaults
func DefaultPerformanceConfig() PerformanceConfig {
	return PerformanceConfig{
		EnableVirtualization: true,
		LazyLoadInactive:     true,
		MaxItemsPerColumn:    MaxItemsPerColumn,
		BatchSize:            BatchSize,
	}
}

// OptimizedColumn wraps Column with performance features
type OptimizedColumn struct {
	Column
	config       PerformanceConfig
	totalItems   int    // Total items in database (may be > loaded)
	loaded       bool   // Whether all items are loaded
	loading      bool   // Whether currently loading more
	isActive     bool   // Active columns (todo, in_progress, review) load first
}

// NewOptimizedColumn creates a performance-optimized column
func NewOptimizedColumn(status, title string, config PerformanceConfig, isActive bool) *OptimizedColumn {
	col := NewColumn(status, title, nil, false)
	
	// Configure virtualization
	if config.EnableVirtualization {
		// bubbles/list handles virtualization internally
		// We just need to provide reasonable height
		col.list.SetShowPagination(true)
	}
	
	return &OptimizedColumn{
		Column:   col,
		config:   config,
		isActive: isActive,
	}
}

// SetItems sets items with count tracking
func (c *OptimizedColumn) SetItems(items []list.Item, totalCount int) {
	c.totalItems = totalCount
	c.loaded = len(items) >= totalCount
	c.list.SetItems(items)
}

// HasMore returns true if there are more items to load
func (c *OptimizedColumn) HasMore() bool {
	return !c.loaded && len(c.list.Items()) < c.totalItems
}

// LoadMore loads additional items
func (c *OptimizedColumn) LoadMore(app *pocketbase.PocketBase, boardID string) tea.Cmd {
	if c.loading || c.loaded {
		return nil
	}
	
	c.loading = true
	currentCount := len(c.list.Items())
	
	return func() tea.Msg {
		// Fetch next batch
		records, err := app.FindRecordsByFilter(
			"tasks",
			"board = {:board} && column = {:column}",
			"+position", // Sort by position
			c.config.BatchSize,
			currentCount, // Offset
			map[string]any{
				"board":  boardID,
				"column": c.status,
			},
		)
		if err != nil {
			return errMsg{err}
		}
		
		return columnLoadedMsg{
			column: c.status,
			items:  recordsToItems(records),
			offset: currentCount,
		}
	}
}

// columnLoadedMsg indicates more items were loaded for a column
type columnLoadedMsg struct {
	column string
	items  []list.Item
	offset int
}

// recordsToItems converts database records to list items
func recordsToItems(records []*core.Record) []list.Item {
	items := make([]list.Item, len(records))
	for i, record := range records {
		items[i] = RecordToTaskItem(record)
	}
	return items
}

// LazyLoader manages deferred loading of inactive columns
type LazyLoader struct {
	pending  map[string]bool
	complete map[string]bool
}

// NewLazyLoader creates a new lazy loader
func NewLazyLoader() *LazyLoader {
	return &LazyLoader{
		pending:  make(map[string]bool),
		complete: make(map[string]bool),
	}
}

// ScheduleLoad marks a column for lazy loading
func (l *LazyLoader) ScheduleLoad(column string) {
	if !l.complete[column] {
		l.pending[column] = true
	}
}

// IsLoadPending returns true if column needs loading
func (l *LazyLoader) IsLoadPending(column string) bool {
	return l.pending[column]
}

// MarkComplete marks a column as fully loaded
func (l *LazyLoader) MarkComplete(column string) {
	l.pending[column] = false
	l.complete[column] = true
}

// GetPendingColumns returns columns that need loading
func (l *LazyLoader) GetPendingColumns() []string {
	var columns []string
	for col, pending := range l.pending {
		if pending {
			columns = append(columns, col)
		}
	}
	return columns
}

// LoadInactiveColumns loads done and backlog columns lazily
func (m *App) loadInactiveColumns() tea.Cmd {
	pending := m.lazyLoader.GetPendingColumns()
	if len(pending) == 0 {
		return nil
	}
	
	// Load one column at a time to avoid overwhelming the database
	column := pending[0]
	
	return func() tea.Msg {
		records, err := m.pb.FindRecordsByFilter(
			"tasks",
			"board = {:board} && column = {:column}",
			"+position",
			m.perfConfig.MaxItemsPerColumn,
			0,
			map[string]any{
				"board":  m.currentBoard.Id,
				"column": column,
			},
		)
		if err != nil {
			return errMsg{err}
		}
		
		return lazyLoadCompleteMsg{
			column: column,
			items:  recordsToItems(records),
		}
	}
}

type lazyLoadCompleteMsg struct {
	column string
	items  []list.Item
}

// OptimizeRender reduces unnecessary re-renders
type RenderCache struct {
	lastRender    string
	lastWidth     int
	lastHeight    int
	dirty         bool
}

// NewRenderCache creates a render cache
func NewRenderCache() *RenderCache {
	return &RenderCache{dirty: true}
}

// MarkDirty forces a re-render
func (c *RenderCache) MarkDirty() {
	c.dirty = true
}

// ShouldRender returns true if re-render is needed
func (c *RenderCache) ShouldRender(width, height int) bool {
	if c.dirty {
		return true
	}
	if width != c.lastWidth || height != c.lastHeight {
		return true
	}
	return false
}

// Cache stores the last render
func (c *RenderCache) Cache(content string, width, height int) {
	c.lastRender = content
	c.lastWidth = width
	c.lastHeight = height
	c.dirty = false
}

// Get returns cached content
func (c *RenderCache) Get() string {
	return c.lastRender
}
```

**Integration Example**:

```go
// In App initialization
func NewApp(pb *pocketbase.PocketBase, boardRef string) *App {
	m := &App{
		pb:          pb,
		perfConfig:  DefaultPerformanceConfig(),
		lazyLoader:  NewLazyLoader(),
		renderCache: NewRenderCache(),
	}
	
	// Schedule lazy loading for inactive columns
	if m.perfConfig.LazyLoadInactive {
		m.lazyLoader.ScheduleLoad("done")
		m.lazyLoader.ScheduleLoad("backlog")
	}
	
	return m
}

// In App.Init()
func (m *App) Init() tea.Cmd {
	return tea.Batch(
		m.loadActiveColumns(),      // Load todo, in_progress, review first
		m.tickLazyLoader(),         // Start lazy load timer
	)
}

// Lazy load timer
func (m *App) tickLazyLoader() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return lazyLoadTickMsg{}
	})
}

// Handle lazy load tick
case lazyLoadTickMsg:
	return m, m.loadInactiveColumns()

case lazyLoadCompleteMsg:
	m.lazyLoader.MarkComplete(msg.column)
	// Add items to column
	for i := range m.columns {
		if m.columns[i].status == msg.column {
			m.columns[i].SetItems(msg.items, len(msg.items))
			break
		}
	}
	// Continue lazy loading if more pending
	return m, m.loadInactiveColumns()
```

**Expected output**:
- Active columns (todo, in_progress, review) load immediately
- Inactive columns (done, backlog) load after 500ms delay
- Large columns use virtualization (smooth scrolling)
- No UI freeze with 500+ tasks

**Common Mistakes**:
- Loading all columns synchronously blocks the UI
- Not using list pagination for very long lists
- Re-rendering on every message even when nothing changed
- Forgetting to track total count for "load more" functionality

---

### 7.6 Unit Tests for Components

**What**: Write comprehensive unit tests for all TUI components to ensure correctness and prevent regressions.

**Why**: Unit tests catch bugs early, document expected behavior, and enable confident refactoring. Bubble Tea's architecture makes components easy to test in isolation.

**Steps**:

1. Create `internal/tui/app_test.go` for App model tests
2. Create `internal/tui/board_test.go` for Board tests
3. Create `internal/tui/column_test.go` for Column tests
4. Create tests for navigation, input handling, and rendering

**Code** (`internal/tui/app_test.go`):

```go
package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApp_Init(t *testing.T) {
	app := NewAppWithoutPB() // Test version without PocketBase
	
	cmd := app.Init()
	
	// Init should return commands to load boards
	require.NotNil(t, cmd, "Init should return a command")
}

func TestApp_HandleResize(t *testing.T) {
	app := NewAppWithoutPB()
	
	// Set initial size
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	model, _ := app.Update(msg)
	updatedApp := model.(*App)
	
	assert.Equal(t, 120, updatedApp.width)
	assert.Equal(t, 40, updatedApp.height)
	assert.False(t, updatedApp.resize.IsTooSmall())
}

func TestApp_HandleResize_TooSmall(t *testing.T) {
	app := NewAppWithoutPB()
	
	// Set size below minimum
	msg := tea.WindowSizeMsg{Width: 40, Height: 10}
	model, _ := app.Update(msg)
	updatedApp := model.(*App)
	
	assert.True(t, updatedApp.resize.IsTooSmall())
	
	// View should show warning message
	view := updatedApp.View()
	assert.Contains(t, view, "Terminal too small")
}

func TestApp_ColumnNavigation(t *testing.T) {
	app := NewAppWithoutPB()
	app.width = 120
	app.height = 40
	app.columns = make([]Column, 5)
	for i := range app.columns {
		app.columns[i] = NewColumn(defaultColumns[i], columnTitles[defaultColumns[i]], nil, i == 0)
	}
	app.focusedCol = 0
	
	// Move right
	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRight})
	updatedApp := model.(*App)
	assert.Equal(t, 1, updatedApp.focusedCol)
	
	// Move right again
	model, _ = updatedApp.Update(tea.KeyMsg{Type: tea.KeyRight})
	updatedApp = model.(*App)
	assert.Equal(t, 2, updatedApp.focusedCol)
	
	// Move left
	model, _ = updatedApp.Update(tea.KeyMsg{Type: tea.KeyLeft})
	updatedApp = model.(*App)
	assert.Equal(t, 1, updatedApp.focusedCol)
}

func TestApp_ColumnNavigation_Wrapping(t *testing.T) {
	app := NewAppWithoutPB()
	app.columns = make([]Column, 5)
	app.focusedCol = 4 // Last column
	
	// Move right from last column should stay at last
	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRight})
	updatedApp := model.(*App)
	assert.Equal(t, 4, updatedApp.focusedCol, "Should not wrap past last column")
	
	// Reset to first column
	updatedApp.focusedCol = 0
	
	// Move left from first column should stay at first
	model, _ = updatedApp.Update(tea.KeyMsg{Type: tea.KeyLeft})
	updatedApp = model.(*App)
	assert.Equal(t, 0, updatedApp.focusedCol, "Should not wrap before first column")
}

func TestApp_QuitKey(t *testing.T) {
	app := NewAppWithoutPB()
	
	model, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	
	// Should return quit command
	assert.NotNil(t, cmd)
	
	// Verify it's a quit command
	msg := cmd()
	_, isQuit := msg.(tea.QuitMsg)
	assert.True(t, isQuit, "Expected quit command")
	_ = model // Avoid unused variable
}

func TestApp_HelpToggle(t *testing.T) {
	app := NewAppWithoutPB()
	app.showHelp = false
	
	// Press ? to show help
	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	updatedApp := model.(*App)
	assert.True(t, updatedApp.showHelp)
	
	// Press ? again to hide help
	model, _ = updatedApp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	updatedApp = model.(*App)
	assert.False(t, updatedApp.showHelp)
}

func TestApp_EscapeClosesOverlays(t *testing.T) {
	app := NewAppWithoutPB()
	
	// Test help overlay
	app.showHelp = true
	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyEscape})
	updatedApp := model.(*App)
	assert.False(t, updatedApp.showHelp, "Esc should close help")
	
	// Test task detail overlay
	app.taskDetail = &TaskDetail{}
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyEscape})
	updatedApp = model.(*App)
	assert.Nil(t, updatedApp.taskDetail, "Esc should close task detail")
	
	// Test task form overlay
	app.taskForm = NewTaskForm(FormModeAdd)
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyEscape})
	updatedApp = model.(*App)
	assert.Nil(t, updatedApp.taskForm, "Esc should close task form")
}

func TestApp_ViewStates(t *testing.T) {
	app := NewAppWithoutPB()
	app.width = 120
	app.height = 40
	
	// Normal view
	app.view = ViewBoard
	view := app.View()
	assert.NotEmpty(t, view)
	
	// Help overlay should show help content
	app.showHelp = true
	view = app.View()
	assert.Contains(t, view, "Help") // Help title or content
}

// Helper to create App without PocketBase for testing
func NewAppWithoutPB() *App {
	return &App{
		resize:      NewResizeState(),
		loading:     NewLoadingIndicator(),
		errorDisplay: NewErrorDisplay(),
		columns:     []Column{},
		view:        ViewBoard,
	}
}

var defaultColumns = []string{"backlog", "todo", "in_progress", "review", "done"}
var columnTitles = map[string]string{
	"backlog":     "Backlog",
	"todo":        "Todo",
	"in_progress": "In Progress",
	"review":      "Review",
	"done":        "Done",
}
```

**Code** (`internal/tui/column_test.go`):

```go
package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestColumn_NewColumn(t *testing.T) {
	items := []list.Item{
		TaskItem{ID: "1", Title: "Task 1"},
		TaskItem{ID: "2", Title: "Task 2"},
	}
	
	col := NewColumn("todo", "Todo", items, true)
	
	assert.Equal(t, "todo", col.status)
	assert.Equal(t, "Todo", col.title)
	assert.True(t, col.focused)
	assert.Equal(t, 2, len(col.list.Items()))
}

func TestColumn_Navigation(t *testing.T) {
	items := []list.Item{
		TaskItem{ID: "1", Title: "Task 1"},
		TaskItem{ID: "2", Title: "Task 2"},
		TaskItem{ID: "3", Title: "Task 3"},
	}
	
	col := NewColumn("todo", "Todo", items, true)
	
	// Initial selection is first item
	assert.Equal(t, 0, col.list.Index())
	
	// Move down
	col.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, col.list.Index())
	
	// Move down again
	col.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 2, col.list.Index())
	
	// Move up
	col.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 1, col.list.Index())
}

func TestColumn_SetSize(t *testing.T) {
	col := NewColumn("todo", "Todo", nil, true)
	
	col.SetSize(30, 20)
	
	// list.Model doesn't expose dimensions directly,
	// so we verify it doesn't panic
	assert.NotPanics(t, func() {
		col.View()
	})
}

func TestColumn_SelectedItem(t *testing.T) {
	items := []list.Item{
		TaskItem{ID: "1", Title: "Task 1"},
		TaskItem{ID: "2", Title: "Task 2"},
	}
	
	col := NewColumn("todo", "Todo", items, true)
	
	// Select first item
	item := col.SelectedItem()
	assert.NotNil(t, item)
	
	task, ok := item.(TaskItem)
	assert.True(t, ok)
	assert.Equal(t, "1", task.ID)
}

func TestColumn_EmptyColumn(t *testing.T) {
	col := NewColumn("todo", "Todo", nil, true)
	
	// No items
	assert.Equal(t, 0, len(col.list.Items()))
	
	// SelectedItem should return nil
	item := col.SelectedItem()
	assert.Nil(t, item)
	
	// View should not panic
	assert.NotPanics(t, func() {
		col.View()
	})
}

func TestColumn_Focus(t *testing.T) {
	col := NewColumn("todo", "Todo", nil, false)
	
	assert.False(t, col.focused)
	
	col.Focus()
	assert.True(t, col.focused)
	
	col.Blur()
	assert.False(t, col.focused)
}

func TestTaskItem_FilterValue(t *testing.T) {
	task := TaskItem{
		ID:          "123",
		Title:       "Fix login bug",
		Description: "Users cannot login with email",
	}
	
	filterValue := task.FilterValue()
	
	// FilterValue should include title and description for search
	assert.Contains(t, filterValue, "Fix login bug")
	assert.Contains(t, filterValue, "Users cannot login")
}

func TestTaskItem_Rendering(t *testing.T) {
	task := TaskItem{
		DisplayID:  "WRK-123",
		Title:      "Test task",
		Priority:   "high",
		Type:       "bug",
		IsBlocked:  true,
	}
	
	// Title() should include key information
	title := task.Title()
	assert.Contains(t, title, "WRK-123")
	assert.Contains(t, title, "Test task")
	
	// With blocked indicator
	// Note: actual rendering depends on lipgloss styling
	assert.NotEmpty(t, title)
}

func TestColumn_Items(t *testing.T) {
	items := []list.Item{
		TaskItem{ID: "1", Title: "Task 1"},
		TaskItem{ID: "2", Title: "Task 2"},
	}
	
	col := NewColumn("todo", "Todo", items, true)
	
	gotItems := col.Items()
	assert.Equal(t, 2, len(gotItems))
}
```

**Code** (`internal/tui/board_test.go`):

```go
package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoard_NewBoard(t *testing.T) {
	board := NewBoard("Work", "WRK")
	
	assert.Equal(t, "Work", board.name)
	assert.Equal(t, "WRK", board.prefix)
	assert.Equal(t, 5, len(board.columns), "Should have 5 default columns")
}

func TestBoard_SetSize(t *testing.T) {
	board := NewBoard("Work", "WRK")
	board.SetSize(120, 40)
	
	assert.Equal(t, 120, board.width)
	assert.Equal(t, 40, board.height)
}

func TestBoard_View(t *testing.T) {
	board := NewBoard("Work", "WRK")
	board.SetSize(120, 40)
	
	view := board.View()
	
	// Should contain column headers
	assert.Contains(t, view, "Backlog")
	assert.Contains(t, view, "Todo")
	assert.Contains(t, view, "In Progress")
}

func TestBoard_Navigation(t *testing.T) {
	board := NewBoard("Work", "WRK")
	
	// Initial focus should be on first column (usually backlog or todo)
	assert.Equal(t, 0, board.focused)
	
	board.FocusNextColumn()
	assert.Equal(t, 1, board.focused)
	
	board.FocusPrevColumn()
	assert.Equal(t, 0, board.focused)
}

func TestBoard_FocusColumn(t *testing.T) {
	board := NewBoard("Work", "WRK")
	
	// Focus by index
	board.FocusColumn(2)
	assert.Equal(t, 2, board.focused)
	
	// Out of bounds should clamp
	board.FocusColumn(10)
	assert.Equal(t, 4, board.focused, "Should clamp to last column")
	
	board.FocusColumn(-1)
	assert.Equal(t, 0, board.focused, "Should clamp to first column")
}

func TestBoard_ColumnByStatus(t *testing.T) {
	board := NewBoard("Work", "WRK")
	
	col := board.ColumnByStatus("in_progress")
	assert.NotNil(t, col)
	assert.Equal(t, "in_progress", col.status)
	
	// Non-existent column
	col = board.ColumnByStatus("invalid")
	assert.Nil(t, col)
}
```

**Expected output**:
```
$ go test ./internal/tui/... -v
=== RUN   TestApp_Init
--- PASS: TestApp_Init (0.00s)
=== RUN   TestApp_HandleResize
--- PASS: TestApp_HandleResize (0.00s)
=== RUN   TestApp_HandleResize_TooSmall
--- PASS: TestApp_HandleResize_TooSmall (0.00s)
...
PASS
ok      github.com/ramtinJ95/EgenSkriven/internal/tui   0.5s
```

**Common Mistakes**:
- Testing rendering output too strictly (lipgloss adds escape codes)
- Not testing edge cases (empty lists, nil items)
- Testing internal state instead of observable behavior
- Forgetting to test keyboard navigation

---

### 7.7 Integration Tests with PocketBase

**What**: Write integration tests that verify the TUI works correctly with real PocketBase data.

**Why**: Unit tests verify components in isolation; integration tests verify the system works end-to-end. These tests catch issues like incorrect queries, data mapping errors, and state synchronization bugs.

**Steps**:

1. Create `internal/tui/integration_test.go`
2. Use `testutil.NewTestApp` for isolated test databases
3. Reuse setup helpers from `internal/commands/test_helpers_test.go`
4. Test complete workflows (load, create, move, delete)

**Code** (`internal/tui/integration_test.go`):

```go
package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ramtinJ95/EgenSkriven/internal/board"
	"github.com/ramtinJ95/EgenSkriven/internal/testutil"
)

// ========== Setup Helpers ==========

// setupTUITestCollections creates all collections needed for TUI tests
func setupTUITestCollections(t *testing.T, app *pocketbase.PocketBase) {
	t.Helper()
	
	// Boards collection
	_, err := app.FindCollectionByNameOrId("boards")
	if err != nil {
		boardsCollection := core.NewBaseCollection("boards")
		boardsCollection.Fields.Add(&core.TextField{Name: "name", Required: true})
		boardsCollection.Fields.Add(&core.TextField{Name: "prefix", Required: true})
		boardsCollection.Fields.Add(&core.JSONField{Name: "columns"})
		boardsCollection.Fields.Add(&core.NumberField{Name: "next_seq"})
		boardsCollection.Fields.Add(&core.TextField{Name: "color"})
		require.NoError(t, app.Save(boardsCollection))
	}
	
	// Tasks collection
	_, err = app.FindCollectionByNameOrId("tasks")
	if err != nil {
		tasksCollection := core.NewBaseCollection("tasks")
		tasksCollection.Fields.Add(&core.TextField{Name: "title", Required: true})
		tasksCollection.Fields.Add(&core.TextField{Name: "description"})
		tasksCollection.Fields.Add(&core.SelectField{
			Name:     "type",
			Required: true,
			Values:   []string{"bug", "feature", "chore"},
		})
		tasksCollection.Fields.Add(&core.SelectField{
			Name:     "priority",
			Required: true,
			Values:   []string{"low", "medium", "high", "urgent"},
		})
		tasksCollection.Fields.Add(&core.SelectField{
			Name:     "column",
			Required: true,
			Values:   []string{"backlog", "todo", "in_progress", "need_input", "review", "done"},
		})
		tasksCollection.Fields.Add(&core.NumberField{Name: "position", Required: true})
		tasksCollection.Fields.Add(&core.JSONField{Name: "labels"})
		tasksCollection.Fields.Add(&core.JSONField{Name: "blocked_by"})
		tasksCollection.Fields.Add(&core.SelectField{
			Name:     "created_by",
			Required: true,
			Values:   []string{"user", "agent", "cli"},
		})
		tasksCollection.Fields.Add(&core.TextField{Name: "board"})
		tasksCollection.Fields.Add(&core.NumberField{Name: "seq"})
		require.NoError(t, app.Save(tasksCollection))
	}
}

// createTUITestBoard creates a board for TUI tests
func createTUITestBoard(t *testing.T, app *pocketbase.PocketBase, name, prefix string) *core.Record {
	t.Helper()
	
	collection, err := app.FindCollectionByNameOrId("boards")
	require.NoError(t, err)
	
	record := core.NewRecord(collection)
	record.Set("name", name)
	record.Set("prefix", prefix)
	record.Set("columns", board.DefaultColumns)
	record.Set("next_seq", 1)
	record.Set("color", "#007bff")
	
	require.NoError(t, app.Save(record))
	return record
}

// createTUITestTask creates a task for TUI tests
func createTUITestTask(t *testing.T, app *pocketbase.PocketBase, boardID, title, column string, position float64, seq int) *core.Record {
	t.Helper()
	
	collection, err := app.FindCollectionByNameOrId("tasks")
	require.NoError(t, err)
	
	record := core.NewRecord(collection)
	record.Set("title", title)
	record.Set("type", "feature")
	record.Set("priority", "medium")
	record.Set("column", column)
	record.Set("position", position)
	record.Set("labels", []string{})
	record.Set("blocked_by", []string{})
	record.Set("created_by", "cli")
	record.Set("board", boardID)
	record.Set("seq", seq)
	
	require.NoError(t, app.Save(record))
	return record
}

// ========== Integration Tests ==========

func TestTUI_LoadTasksFromDatabase(t *testing.T) {
	pb := testutil.NewTestApp(t)
	setupTUITestCollections(t, pb)
	
	// Create test data
	testBoard := createTUITestBoard(t, pb, "Test Board", "TST")
	createTUITestTask(t, pb, testBoard.Id, "Task 1", "todo", 1000.0, 1)
	createTUITestTask(t, pb, testBoard.Id, "Task 2", "in_progress", 1000.0, 2)
	createTUITestTask(t, pb, testBoard.Id, "Task 3", "todo", 2000.0, 3)
	
	// Create TUI model
	tuiApp := NewApp(pb, testBoard.Id)
	
	// Simulate window size
	tuiApp.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	
	// Execute init command to load data
	cmd := tuiApp.Init()
	require.NotNil(t, cmd)
	
	// Process messages until tasks are loaded
	// In real tests, you'd use a test harness to process messages
	tasks, err := pb.FindAllRecords("tasks")
	require.NoError(t, err)
	require.Equal(t, 3, len(tasks))
}

func TestTUI_CountTasksPerColumn(t *testing.T) {
	pb := testutil.NewTestApp(t)
	setupTUITestCollections(t, pb)
	
	testBoard := createTUITestBoard(t, pb, "Work", "WRK")
	
	// Create tasks in different columns
	createTUITestTask(t, pb, testBoard.Id, "Backlog 1", "backlog", 1000.0, 1)
	createTUITestTask(t, pb, testBoard.Id, "Todo 1", "todo", 1000.0, 2)
	createTUITestTask(t, pb, testBoard.Id, "Todo 2", "todo", 2000.0, 3)
	createTUITestTask(t, pb, testBoard.Id, "In Progress 1", "in_progress", 1000.0, 4)
	
	// Query tasks by column
	todoTasks, err := pb.FindRecordsByFilter(
		"tasks",
		"board = {:board} && column = {:column}",
		"+position",
		0,
		0,
		map[string]any{
			"board":  testBoard.Id,
			"column": "todo",
		},
	)
	require.NoError(t, err)
	assert.Equal(t, 2, len(todoTasks), "Should have 2 todo tasks")
	
	inProgressTasks, err := pb.FindRecordsByFilter(
		"tasks",
		"board = {:board} && column = {:column}",
		"+position",
		0,
		0,
		map[string]any{
			"board":  testBoard.Id,
			"column": "in_progress",
		},
	)
	require.NoError(t, err)
	assert.Equal(t, 1, len(inProgressTasks), "Should have 1 in_progress task")
}

func TestTUI_TaskOrdering(t *testing.T) {
	pb := testutil.NewTestApp(t)
	setupTUITestCollections(t, pb)
	
	testBoard := createTUITestBoard(t, pb, "Work", "WRK")
	
	// Create tasks with specific positions
	createTUITestTask(t, pb, testBoard.Id, "Third", "todo", 3000.0, 1)
	createTUITestTask(t, pb, testBoard.Id, "First", "todo", 1000.0, 2)
	createTUITestTask(t, pb, testBoard.Id, "Second", "todo", 2000.0, 3)
	
	// Query tasks sorted by position
	tasks, err := pb.FindRecordsByFilter(
		"tasks",
		"board = {:board} && column = {:column}",
		"+position",
		0,
		0,
		map[string]any{
			"board":  testBoard.Id,
			"column": "todo",
		},
	)
	require.NoError(t, err)
	require.Equal(t, 3, len(tasks))
	
	// Verify order
	assert.Equal(t, "First", tasks[0].GetString("title"))
	assert.Equal(t, "Second", tasks[1].GetString("title"))
	assert.Equal(t, "Third", tasks[2].GetString("title"))
}

func TestTUI_MoveTaskBetweenColumns(t *testing.T) {
	pb := testutil.NewTestApp(t)
	setupTUITestCollections(t, pb)
	
	testBoard := createTUITestBoard(t, pb, "Work", "WRK")
	task := createTUITestTask(t, pb, testBoard.Id, "Moving Task", "todo", 1000.0, 1)
	
	// Move task to in_progress
	task.Set("column", "in_progress")
	require.NoError(t, pb.Save(task))
	
	// Verify task is now in in_progress
	updated, err := pb.FindRecordById("tasks", task.Id)
	require.NoError(t, err)
	assert.Equal(t, "in_progress", updated.GetString("column"))
	
	// Verify no task in todo
	todoTasks, err := pb.FindRecordsByFilter(
		"tasks",
		"board = {:board} && column = {:column}",
		"",
		0,
		0,
		map[string]any{
			"board":  testBoard.Id,
			"column": "todo",
		},
	)
	require.NoError(t, err)
	assert.Equal(t, 0, len(todoTasks))
}

func TestTUI_DeleteTask(t *testing.T) {
	pb := testutil.NewTestApp(t)
	setupTUITestCollections(t, pb)
	
	testBoard := createTUITestBoard(t, pb, "Work", "WRK")
	task := createTUITestTask(t, pb, testBoard.Id, "To Delete", "todo", 1000.0, 1)
	
	// Delete the task
	require.NoError(t, pb.Delete(task))
	
	// Verify task is gone
	_, err := pb.FindRecordById("tasks", task.Id)
	assert.Error(t, err, "Task should not be found after deletion")
}

func TestTUI_BoardsMap(t *testing.T) {
	pb := testutil.NewTestApp(t)
	setupTUITestCollections(t, pb)
	
	// Create multiple boards
	board1 := createTUITestBoard(t, pb, "Work", "WRK")
	board2 := createTUITestBoard(t, pb, "Personal", "PER")
	
	// Build boards map
	boards, err := pb.FindAllRecords("boards")
	require.NoError(t, err)
	
	boardsMap := make(map[string]*core.Record)
	for _, b := range boards {
		boardsMap[b.Id] = b
	}
	
	// Verify lookup
	assert.Equal(t, "WRK", boardsMap[board1.Id].GetString("prefix"))
	assert.Equal(t, "PER", boardsMap[board2.Id].GetString("prefix"))
}

func TestTUI_DisplayIDGeneration(t *testing.T) {
	pb := testutil.NewTestApp(t)
	setupTUITestCollections(t, pb)
	
	testBoard := createTUITestBoard(t, pb, "Work", "WRK")
	task := createTUITestTask(t, pb, testBoard.Id, "Test Task", "todo", 1000.0, 42)
	
	// Get task and board
	boardRecord, err := pb.FindRecordById("boards", testBoard.Id)
	require.NoError(t, err)
	
	// Build display ID
	prefix := boardRecord.GetString("prefix")
	seq := task.GetInt("seq")
	displayID := board.FormatDisplayID(prefix, seq)
	
	assert.Equal(t, "WRK-42", displayID)
}

func TestTUI_RecordsToTaskItems(t *testing.T) {
	pb := testutil.NewTestApp(t)
	setupTUITestCollections(t, pb)
	
	testBoard := createTUITestBoard(t, pb, "Work", "WRK")
	createTUITestTask(t, pb, testBoard.Id, "Task 1", "todo", 1000.0, 1)
	createTUITestTask(t, pb, testBoard.Id, "Task 2", "todo", 2000.0, 2)
	
	tasks, err := pb.FindAllRecords("tasks")
	require.NoError(t, err)
	
	// Convert to TaskItems
	items := make([]list.Item, len(tasks))
	for i, task := range tasks {
		items[i] = TaskItem{
			ID:        task.Id,
			Title:     task.GetString("title"),
			Column:    task.GetString("column"),
			Priority:  task.GetString("priority"),
			Type:      task.GetString("type"),
			Position:  task.GetFloat("position"),
		}
	}
	
	require.Equal(t, 2, len(items))
	
	// Verify items implement list.Item
	for _, item := range items {
		taskItem := item.(TaskItem)
		assert.NotEmpty(t, taskItem.FilterValue())
	}
}

func TestTUI_EmptyBoardState(t *testing.T) {
	pb := testutil.NewTestApp(t)
	setupTUITestCollections(t, pb)
	
	testBoard := createTUITestBoard(t, pb, "Empty Board", "EMP")
	
	// No tasks created
	tasks, err := pb.FindRecordsByFilter(
		"tasks",
		"board = {:board}",
		"",
		0,
		0,
		map[string]any{"board": testBoard.Id},
	)
	require.NoError(t, err)
	assert.Equal(t, 0, len(tasks), "Board should have no tasks")
}

func TestTUI_LargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}
	
	pb := testutil.NewTestApp(t)
	setupTUITestCollections(t, pb)
	
	testBoard := createTUITestBoard(t, pb, "Large Board", "LRG")
	
	// Create 100 tasks
	for i := 1; i <= 100; i++ {
		columns := []string{"backlog", "todo", "in_progress", "review", "done"}
		column := columns[i%5]
		createTUITestTask(t, pb, testBoard.Id, "Task "+string(rune('0'+i)), column, float64(i*1000), i)
	}
	
	// Verify all tasks created
	tasks, err := pb.FindAllRecords("tasks")
	require.NoError(t, err)
	assert.Equal(t, 100, len(tasks))
	
	// Verify column distribution
	for _, col := range []string{"backlog", "todo", "in_progress", "review", "done"} {
		colTasks, err := pb.FindRecordsByFilter(
			"tasks",
			"column = {:column}",
			"",
			0,
			0,
			map[string]any{"column": col},
		)
		require.NoError(t, err)
		assert.Equal(t, 20, len(colTasks), "Each column should have 20 tasks")
	}
}
```

**Expected output**:
```
$ go test ./internal/tui/... -v -run Integration
=== RUN   TestTUI_LoadTasksFromDatabase
--- PASS: TestTUI_LoadTasksFromDatabase (0.05s)
=== RUN   TestTUI_CountTasksPerColumn
--- PASS: TestTUI_CountTasksPerColumn (0.04s)
...
PASS
ok      github.com/ramtinJ95/EgenSkriven/internal/tui   1.2s
```

**Common Mistakes**:
- Not using isolated test databases (tests affect each other)
- Hardcoding IDs instead of using created record IDs
- Not testing error paths (what if collection doesn't exist?)
- Forgetting to clean up test data (handled by `t.Cleanup`)

---

### 7.8 Documentation and Usage Guide

**What**: Add comprehensive help text to the TUI command and update the main CLI help.

**Why**: Users need to know how to launch the TUI and what keybindings are available. Good documentation reduces support burden.

**Steps**:

1. Update `internal/commands/tui.go` with detailed help text
2. Add examples to command help
3. Create in-app help content
4. Update README if applicable

**Code** (Update `internal/commands/tui.go`):

```go
package commands

import (
	"fmt"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"

	"github.com/ramtinJ95/EgenSkriven/internal/config"
	"github.com/ramtinJ95/EgenSkriven/internal/output"
	"github.com/ramtinJ95/EgenSkriven/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func newTuiCmd(app *pocketbase.PocketBase) *cobra.Command {
	var boardRef string
	var serverURL string

	cmd := &cobra.Command{
		Use:   "tui",
		Aliases: []string{"ui", "board"},
		Short: "Open interactive kanban board",
		Long: `Launch the terminal user interface for managing tasks in a kanban board view.

The TUI provides a full-featured kanban board experience with:
  - 5 columns: Backlog, Todo, In Progress, Review, Done
  - Keyboard-driven navigation
  - Real-time sync when server is running
  - Task creation, editing, and movement

KEYBOARD SHORTCUTS:

  Navigation:
    h/Left     Move to previous column
    l/Right    Move to next column
    j/Down     Select next task
    k/Up       Select previous task
    g          Go to first task
    G          Go to last task

  Task Actions:
    Enter      View task details
    n          Create new task
    e          Edit selected task
    d          Delete task (with confirmation)
    Space      Toggle task selection

  Task Movement:
    H          Move task to previous column
    L          Move task to next column
    K          Move task up in column
    J          Move task down in column
    1-5        Move task to column by number

  Filtering:
    /          Search tasks
    fp         Filter by priority
    ft         Filter by type
    fl         Filter by label
    fc         Clear all filters

  Other:
    b          Switch boards
    ?          Toggle help
    q          Quit`,
		Example: `  # Open TUI with default board
  egenskriven tui

  # Open TUI with specific board
  egenskriven tui -b work

  # Open TUI with real-time sync from server
  egenskriven tui --server http://localhost:8090`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := output.New(false, false)

			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			// Load config for default board
			cfg, _ := config.LoadProjectConfig()
			if boardRef == "" && cfg != nil {
				boardRef = cfg.DefaultBoard
			}

			// Create TUI app
			tuiApp := tui.NewApp(app, boardRef)
			if serverURL != "" {
				tuiApp.SetServerURL(serverURL)
			}

			// Run TUI
			p := tea.NewProgram(tuiApp, tea.WithAltScreen())
			_, err := p.Run()
			if err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("TUI error: %v", err), nil)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&boardRef, "board", "b", "", "Board to open (name, prefix, or ID)")
	cmd.Flags().StringVarP(&serverURL, "server", "s", "", "Server URL for real-time sync")

	return cmd
}
```

**In-App Help Content** (add to `internal/tui/help.go`):

```go
package tui

const helpText = `
KEYBOARD SHORTCUTS

Navigation
  h/Left     Previous column
  l/Right    Next column
  j/Down     Next task
  k/Up       Previous task
  g          First task
  G          Last task
  Tab        Cycle columns

Actions
  Enter      View details
  n          New task
  e          Edit task
  d          Delete task
  Space      Select task

Movement
  H          Move left
  L          Move right
  K          Move up
  J          Move down
  1-5        Move to column

Filters
  /          Search
  fp         Filter priority
  ft         Filter type
  fl         Filter label
  fc         Clear filters

Other
  b          Switch board
  Ctrl+R     Refresh
  ?          Toggle help
  q          Quit
`
```

**Expected output**:
```
$ egenskriven tui --help
Launch the terminal user interface for managing tasks in a kanban board view.

The TUI provides a full-featured kanban board experience with:
  - 5 columns: Backlog, Todo, In Progress, Review, Done
  - Keyboard-driven navigation
  ...

KEYBOARD SHORTCUTS:

  Navigation:
    h/Left     Move to previous column
  ...

Examples:
  # Open TUI with default board
  egenskriven tui
  ...
```

**Common Mistakes**:
- Help text too long (users won't read it)
- Missing common use cases in examples
- Not documenting all keybindings
- Inconsistent formatting

---

### 7.9 Final CLI Integration

**What**: Ensure the TUI command is properly integrated into the main CLI help and accessible from the root command.

**Why**: The TUI should be discoverable from `egenskriven --help` and work seamlessly with other commands.

**Steps**:

1. Verify TUI command is registered in `root.go`
2. Update main CLI help to mention TUI
3. Test command discovery and execution

**Code** (Verify in `internal/commands/root.go`):

```go
package commands

import (
	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"
)

// Register adds all CLI commands to the PocketBase app
func Register(app *pocketbase.PocketBase) {
	// Task commands
	app.RootCmd.AddCommand(newAddCmd(app))
	app.RootCmd.AddCommand(newListCmd(app))
	app.RootCmd.AddCommand(newShowCmd(app))
	app.RootCmd.AddCommand(newMoveCmd(app))
	app.RootCmd.AddCommand(newUpdateCmd(app))
	app.RootCmd.AddCommand(newDeleteCmd(app))
	
	// Board commands
	app.RootCmd.AddCommand(newBoardCmd(app))
	
	// Session commands
	app.RootCmd.AddCommand(newSessionCmd(app))
	app.RootCmd.AddCommand(newBlockCmd(app))
	app.RootCmd.AddCommand(newResumeCmd(app))
	app.RootCmd.AddCommand(newCommentCmd(app))
	app.RootCmd.AddCommand(newCommentsCmd(app))
	
	// TUI command - interactive kanban board
	app.RootCmd.AddCommand(newTuiCmd(app))
	
	// Utility commands
	app.RootCmd.AddCommand(newExportCmd(app))
	app.RootCmd.AddCommand(newImportCmd(app))
	app.RootCmd.AddCommand(newBackupCmd(app))
	app.RootCmd.AddCommand(newContextCmd(app))
	app.RootCmd.AddCommand(newSuggestCmd(app))
	app.RootCmd.AddCommand(newPrimeCmd(app))
	
	// Configuration
	app.RootCmd.AddCommand(newConfigCmd(app))
	app.RootCmd.AddCommand(newInitCmd(app))
}
```

**Verification Commands**:

```bash
# Verify TUI appears in help
egenskriven --help | grep -A1 "tui"

# Verify TUI aliases work
egenskriven ui --help
egenskriven board --help  # Note: may conflict with board subcommand

# Test TUI launches
egenskriven tui
```

**Expected output**:
```
$ egenskriven --help
...
Available Commands:
  add         Create a new task
  board       Manage boards
  tui         Open interactive kanban board (aliases: ui)
...

$ egenskriven tui --help
Launch the terminal user interface...
```

---

## Verification Checklist

Complete each section in order. Check off each item as you verify it.

### Resize Handling

- [ ] **Terminal resize works**
  ```bash
  # Start TUI, then resize terminal window
  egenskriven tui
  ```
  Columns should adjust proportionally.

- [ ] **Small terminal shows warning**
  
  Resize to less than 60x15 characters.
  Should show "Terminal too small" message.

- [ ] **Overlay components resize**
  
  Open task detail, resize terminal.
  Detail panel should adjust.

### Loading States

- [ ] **Loading indicator appears**
  
  Launch TUI and observe loading spinner.
  Should show "Loading tasks...".

- [ ] **Loading stops on completion**
  
  Spinner should disappear when data loads.

- [ ] **Loading on save operations**
  
  Create a task (press 'n').
  Should show "Saving..." during save.

### Error Handling

- [ ] **Errors display correctly**
  
  Force an error (e.g., invalid board).
  Should show error message in colored bar.

- [ ] **Errors auto-dismiss**
  
  Wait 5 seconds.
  Error should disappear.

- [ ] **Escape dismisses errors**
  
  Press Esc while error is showing.
  Error should disappear immediately.

### Empty States

- [ ] **Empty board shows message**
  ```bash
  # Create empty board, open in TUI
  egenskriven board add "Empty" EMP
  egenskriven tui -b EMP
  ```
  Should show "No tasks yet" message.

- [ ] **Empty search shows message**
  
  Press '/' and search for "xyznonexistent".
  Should show "No results" message.

- [ ] **Empty filter shows message**
  
  Filter by a priority with no tasks.
  Should show "No matches" message.

### Performance

- [ ] **Large board loads quickly**
  ```bash
  # Create many tasks, then open TUI
  for i in {1..100}; do egenskriven add "Task $i" --type feature; done
  time egenskriven tui
  ```
  Should load in under 2 seconds.

- [ ] **Scrolling is smooth**
  
  Navigate up/down in large column.
  Should not lag or stutter.

- [ ] **Lazy loading works**
  
  Done column should load after a brief delay.

### Unit Tests

- [ ] **All unit tests pass**
  ```bash
  go test ./internal/tui/... -v
  ```
  All tests should pass.

- [ ] **Coverage is adequate**
  ```bash
  go test ./internal/tui/... -cover
  ```
  Should be >70% coverage.

### Integration Tests

- [ ] **Integration tests pass**
  ```bash
  go test ./internal/tui/... -v -run Integration
  ```
  All integration tests should pass.

- [ ] **No test database leaks**
  ```bash
  ls /tmp | grep egenskriven-test
  ```
  Should be empty after tests complete.

### Documentation

- [ ] **TUI help is complete**
  ```bash
  egenskriven tui --help
  ```
  Should show all keybindings.

- [ ] **Examples work**
  ```bash
  egenskriven tui
  egenskriven tui -b work
  ```
  Both should launch successfully.

- [ ] **In-app help works**
  
  Press '?' in TUI.
  Should show help overlay with all shortcuts.

### Final Integration

- [ ] **TUI appears in main help**
  ```bash
  egenskriven --help | grep tui
  ```
  Should list TUI command.

- [ ] **Aliases work**
  ```bash
  egenskriven ui
  ```
  Should launch TUI.

---

## File Summary

| File | Lines | Purpose |
|------|-------|---------|
| `internal/tui/resize.go` | ~100 | Terminal resize handling |
| `internal/tui/loading.go` | ~120 | Loading spinner and state |
| `internal/tui/error.go` | ~150 | Error display with auto-dismiss |
| `internal/tui/empty.go` | ~100 | Empty state displays |
| `internal/tui/performance.go` | ~200 | Optimization utilities |
| `internal/tui/app_test.go` | ~200 | Unit tests for App model |
| `internal/tui/column_test.go` | ~150 | Unit tests for Column |
| `internal/tui/board_test.go` | ~100 | Unit tests for Board |
| `internal/tui/integration_test.go` | ~300 | Integration tests with PocketBase |

**Total new code**: ~1,420 lines

---

## What You Should Have Now

After completing Phase 7, your TUI should:

```
internal/tui/
├── app.go              # Main application (from Phase 1)
├── app_test.go         # Unit tests
├── board.go            # Board model (from Phase 1)
├── board_test.go       # Unit tests
├── board_selector.go   # Board switching (from Phase 3)
├── column.go           # Column component (from Phase 1)
├── column_test.go      # Unit tests
├── commands.go         # Async commands (from Phase 1)
├── empty.go            # Empty state displays
├── error.go            # Error display component
├── filter_bar.go       # Filter controls (from Phase 5)
├── help.go             # Help overlay (from Phase 6)
├── integration_test.go # Integration tests
├── keys.go             # Keybindings (from Phase 1)
├── loading.go          # Loading indicator
├── messages.go         # Message types (from Phase 1)
├── performance.go      # Performance optimizations
├── realtime.go         # Real-time sync (from Phase 4)
├── resize.go           # Resize handling
├── styles.go           # Lipgloss styles (from Phase 1)
├── task_detail.go      # Task detail view (from Phase 2)
├── task_form.go        # Task form (from Phase 2)
└── task_item.go        # Task list item (from Phase 1)
```

**Production-Ready Features**:
- Graceful terminal resize
- Loading indicators for all async operations
- Error handling with suggestions
- Empty states for all scenarios
- Performance optimization for large boards
- Comprehensive test coverage
- Complete documentation

---

## Success Criteria

### MVP (Must Have)
- [ ] Terminal resize doesn't crash
- [ ] Loading indicator appears during data fetch
- [ ] Errors display and dismiss correctly
- [ ] Empty boards show helpful message
- [ ] Unit tests pass
- [ ] TUI launches from CLI

### Full Release (Should Have)
- [ ] All MVP features
- [ ] Resize updates all components
- [ ] Loading shows elapsed time for slow operations
- [ ] Different error severities (info, warning, error, fatal)
- [ ] Empty states for search and filters
- [ ] Lazy loading for large boards
- [ ] Integration tests pass
- [ ] In-app help overlay
- [ ] Complete CLI documentation

---

## Troubleshooting

### "Terminal too small" appears immediately

**Problem**: Minimum size check is failing even with large terminal.

**Solution**:
```go
// Verify window size message is being received
func (m *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        log.Printf("Window size: %dx%d", msg.Width, msg.Height)
        // ...
    }
}
```

### Loading spinner doesn't animate

**Problem**: Spinner tick messages aren't being processed.

**Solution**:
```go
// Ensure spinner ticks are handled in Update
case spinner.TickMsg:
    var cmd tea.Cmd
    m.loading, cmd = m.loading.Update(msg)
    return m, cmd
```

### Tests fail with "collection not found"

**Problem**: Test setup isn't creating collections.

**Solution**:
```go
// Call setup before any test operations
func TestSomething(t *testing.T) {
    app := testutil.NewTestApp(t)
    setupTUITestCollections(t, app)  // <-- Add this
    // ...
}
```

### Performance tests are slow

**Problem**: Creating 100+ records takes too long.

**Solution**:
```go
// Use batch inserts if available, or skip in short mode
func TestLargeDataset(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping in short mode")
    }
    // ...
}

// Run with: go test -short ./...
```

### Resize causes layout glitches

**Problem**: Components not updating dimensions correctly.

**Solution**:
```go
// Ensure all components get size updates
func (m *App) handleResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
    m.width = msg.Width
    m.height = msg.Height
    
    // Update EVERY component that cares about size
    for i := range m.columns {
        m.columns[i].SetSize(colWidth, colHeight)
    }
    if m.taskDetail != nil {
        m.taskDetail.SetSize(detailWidth, m.height-2)
    }
    // ... etc
}
```

---

## Glossary

| Term | Definition |
|------|------------|
| **Virtualization** | Rendering only visible items in a list for performance |
| **Lazy Loading** | Deferring data fetch until needed |
| **Integration Test** | Test that verifies multiple components work together |
| **Unit Test** | Test that verifies a single component in isolation |
| **Empty State** | UI shown when there's no data to display |
| **Auto-dismiss** | Automatically hiding a message after a timeout |
| **Render Cache** | Storing rendered output to avoid redundant work |
| **tea.WindowSizeMsg** | Bubble Tea message sent when terminal is resized |
| **spinner.TickMsg** | Bubble Tea message for spinner animation frames |

---

## Next Steps

Congratulations! With Phase 7 complete, you have a production-ready TUI. Consider these optional enhancements:

1. **Theme Support**: Allow users to customize colors
2. **Mouse Support**: Add mouse click handling for navigation
3. **Custom Columns**: Support user-defined column layouts
4. **Export/Import**: Export visible tasks to markdown/JSON
5. **Notifications**: Desktop notifications for blocked tasks
6. **Metrics**: Track usage patterns for optimization
