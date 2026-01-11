# Phase 1: TUI Foundation

**Goal**: Basic kanban board with column/task navigation in the terminal.

**Duration Estimate**: 3-4 days

**Prerequisites**: 
- Completed core CLI phases (working `egenskriven` binary)
- PocketBase running with `tasks` and `boards` collections
- At least one board created (`egenskriven board create`)
- Go 1.21+ installed

**Deliverable**: Can launch TUI with `egenskriven tui`, see tasks in columns, navigate with keyboard (h/j/k/l).

---

## Overview

This phase builds the foundation for EgenSkriven's Terminal User Interface (TUI). We're using the Elm-inspired Model-View-Update architecture provided by Bubble Tea, the most popular Go TUI framework.

### Why Bubble Tea?

- **Functional Architecture**: The Model-View-Update pattern makes state management predictable and testable
- **Rich Component Library**: `bubbles` provides pre-built components (lists, text inputs, viewports)
- **Beautiful Styling**: Lip Gloss enables sophisticated terminal styling
- **Async Support**: Commands handle async operations cleanly without callbacks

### What We're Building

```
+-----------------------------------------------------------------------+
| EgenSkriven - Work Board (WRK)                                        |
+-----------------------------------------------------------------------+
| BACKLOG (3)    | TODO (2)       | IN_PROGRESS (1)| REVIEW (0) | DONE |
|----------------|----------------|----------------|------------|------|
| * WRK-1 Setup  | * WRK-4 Auth   | > WRK-6 API    |            | ...  |
| * WRK-2 Config | * WRK-5 Tests  |                |            |      |
| * WRK-3 Deps   |                |                |            |      |
+-----------------------------------------------------------------------+
| h/l: columns | j/k: tasks | enter: details | q: quit | ?: help       |
+-----------------------------------------------------------------------+
```

The `>` indicates the currently selected task. Navigation uses vim-style keys.

---

## Environment Requirements

Before starting, ensure you have the Bubble Tea dependencies available:

| Dependency | Purpose | Check Command |
|------------|---------|---------------|
| Go 1.21+ | Language runtime | `go version` |
| Bubble Tea | TUI framework | Added in this phase |
| Bubbles | UI components | Added in this phase |
| Lip Gloss | Styling | Added in this phase |

---

## Tasks

### 1.1 Create Directory Structure [COMPLETED]

**What**: Create the `internal/tui/` directory with all necessary Go files.

**Why**: Organizing TUI code in its own package keeps it separate from CLI commands and allows for clean imports. Each file has a specific responsibility.

**Steps**:

1. Create the TUI directory and files:
   ```bash
   mkdir -p internal/tui
   ```

2. Create empty Go files:
   ```bash
   touch internal/tui/app.go
   touch internal/tui/board.go
   touch internal/tui/column.go
   touch internal/tui/task_item.go
   touch internal/tui/keys.go
   touch internal/tui/styles.go
   touch internal/tui/messages.go
   touch internal/tui/commands.go
   ```

3. Verify the structure:
   ```bash
   ls -la internal/tui/
   ```
   
   **Expected output**:
   ```
   total 0
   drwxr-xr-x  app.go
   drwxr-xr-x  board.go
   drwxr-xr-x  column.go
   drwxr-xr-x  task_item.go
   drwxr-xr-x  keys.go
   drwxr-xr-x  styles.go
   drwxr-xr-x  messages.go
   drwxr-xr-x  commands.go
   ```

**Directory Structure Explained**:

| File | Purpose |
|------|---------|
| `app.go` | Main application model, Init/Update/View, tea.Program entry |
| `board.go` | Board view component containing columns |
| `column.go` | Single column wrapping bubbles/list |
| `task_item.go` | TaskItem implementing list.Item interface |
| `keys.go` | Keybinding definitions using bubbles/key |
| `styles.go` | Lip Gloss style definitions |
| `messages.go` | Custom message types for Bubble Tea |
| `commands.go` | Async command functions (database operations) |

**Common Mistakes**:
- Creating files in wrong directory (must be `internal/tui/`, not `tui/`)
- Missing package declaration in files

---

### 1.2 Add Bubble Tea Dependencies [COMPLETED]

**What**: Add the Charm libraries to `go.mod`.

**Why**: These libraries provide the TUI framework, pre-built components, and styling capabilities.

**Steps**:

1. Add the required dependencies:
   ```bash
   go get github.com/charmbracelet/bubbletea@latest
   go get github.com/charmbracelet/bubbles@latest
   go get github.com/charmbracelet/lipgloss@latest
   ```

2. Tidy the module:
   ```bash
   go mod tidy
   ```

3. Verify the dependencies were added:
   ```bash
   grep charmbracelet go.mod
   ```
   
   **Expected output** (versions may vary):
   ```
   github.com/charmbracelet/bubbletea v0.25.0
   github.com/charmbracelet/bubbles v0.18.0
   github.com/charmbracelet/lipgloss v0.9.1
   ```

**Package Overview**:

| Package | What It Provides |
|---------|------------------|
| `bubbletea` | Core framework: `tea.Model`, `tea.Cmd`, `tea.Program` |
| `bubbles/list` | Scrollable list component with filtering |
| `bubbles/key` | Keybinding definitions with help text |
| `bubbles/help` | Auto-generated help from keybindings |
| `lipgloss` | CSS-like styling for terminal output |

**Common Mistakes**:
- Forgetting to run `go mod tidy`
- Version conflicts with existing dependencies

---

### 1.3 Implement Messages [COMPLETED]

**What**: Define custom message types for communication between components.

**Why**: Bubble Tea uses messages to communicate state changes. Custom messages allow components to signal events like "tasks loaded" or "error occurred".

**File**: `internal/tui/messages.go`

```go
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
```

**Key Concepts**:

- Messages are sent via `tea.Cmd` functions that return a `tea.Msg`
- The `Update` function receives messages and updates model state
- Messages should be simple structs with the data needed to update state

**Common Mistakes**:
- Making messages too complex (they should just carry data)
- Forgetting to handle messages in the Update function

---

### 1.4 Implement Key Bindings [COMPLETED]

**What**: Define all keyboard shortcuts for TUI navigation and actions.

**Why**: Centralized keybinding definitions enable consistent help text generation and make it easy to change keys later.

**File**: `internal/tui/keys.go`

```go
package tui

import (
	"github.com/charmbracelet/bubbles/key"
)

// keyMap defines all keybindings for the TUI.
// Each binding includes the keys and help text.
type keyMap struct {
	// Navigation - moving between columns and tasks
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding

	// Actions - interacting with tasks
	Enter key.Binding

	// Global - application-level controls
	Quit   key.Binding
	Help   key.Binding
	Escape key.Binding
}

// defaultKeyMap returns the default keybindings.
// Uses vim-style navigation (h/j/k/l) with arrow key alternatives.
func defaultKeyMap() keyMap {
	return keyMap{
		// Up moves selection up within a column (previous task)
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("k/up", "up"),
		),
		// Down moves selection down within a column (next task)
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("j/down", "down"),
		),
		// Left moves focus to the previous column
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("h/left", "prev column"),
		),
		// Right moves focus to the next column
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("l/right", "next column"),
		),
		// Enter opens task details (Phase 2)
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "view details"),
		),
		// Quit exits the TUI
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		// Help toggles the help overlay (Phase 2)
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		// Escape closes overlays or cancels operations
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

// ShortHelp returns the keybindings to show in the short help view.
// These are displayed in the status bar.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Left, k.Right, k.Up, k.Down, k.Enter, k.Quit, k.Help}
}

// FullHelp returns all keybindings grouped for the full help view.
// Used when the user presses '?' to see all available keys.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},     // Navigation
		{k.Enter, k.Quit, k.Help, k.Escape}, // Actions
	}
}
```

**Key Concepts**:

- `key.NewBinding` creates a binding with keys and help text
- `WithKeys` accepts multiple keys for the same action
- `ShortHelp` and `FullHelp` implement the `help.KeyMap` interface

**Common Mistakes**:
- Using key combinations that conflict with terminal shortcuts
- Forgetting to add new bindings to help methods

---

### 1.5 Implement Styles [COMPLETED]

**What**: Define Lip Gloss styles for consistent visual appearance.

**Why**: Centralized styles ensure visual consistency and make it easy to adjust colors and spacing.

**File**: `internal/tui/styles.go`

```go
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette - using ANSI 256 colors for terminal compatibility.
// These colors work well on both light and dark terminal backgrounds.
var (
	// Brand colors
	primaryColor   = lipgloss.Color("62")  // Blue - used for focused elements
	secondaryColor = lipgloss.Color("205") // Pink - used for accents

	// Priority colors - visual indicators for task urgency
	priorityUrgent = lipgloss.Color("196") // Red
	priorityHigh   = lipgloss.Color("208") // Orange
	priorityMedium = lipgloss.Color("226") // Yellow
	priorityLow    = lipgloss.Color("240") // Gray

	// Type colors - distinguish task types visually
	typeBug     = lipgloss.Color("196") // Red - bugs stand out
	typeFeature = lipgloss.Color("39")  // Cyan - features are positive
	typeChore   = lipgloss.Color("240") // Gray - chores are neutral

	// Column header colors - each column has a distinct color
	columnBacklog    = lipgloss.Color("240") // Gray
	columnTodo       = lipgloss.Color("39")  // Cyan
	columnInProgress = lipgloss.Color("214") // Orange
	columnNeedInput  = lipgloss.Color("196") // Red (needs attention)
	columnReview     = lipgloss.Color("205") // Pink
	columnDone       = lipgloss.Color("82")  // Green

	// UI element colors
	borderColor      = lipgloss.Color("240") // Gray border
	focusBorderColor = lipgloss.Color("62")  // Blue border for focused
	mutedColor       = lipgloss.Color("240") // Gray for secondary text
	textColor        = lipgloss.Color("252") // Light gray for text
)

// Column styles - different border colors for focused vs unfocused columns.
var (
	// focusedColumnStyle is applied to the currently selected column.
	// It has a highlighted border and slightly different background.
	focusedColumnStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(focusBorderColor).
				Padding(0, 1)

	// blurredColumnStyle is applied to non-selected columns.
	// Uses a dimmer border color.
	blurredColumnStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(borderColor).
				Padding(0, 1)
)

// Header and title styles.
var (
	// headerStyle is used for the main application header.
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Padding(0, 1)

	// boardTitleStyle is used for the board name in the header.
	boardTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(secondaryColor)

	// columnTitleStyle is the base style for column headers.
	// The foreground color is set dynamically based on column type.
	columnTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Padding(0, 1)
)

// Status bar styles.
var (
	// statusBarStyle is the background style for the status bar.
	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(textColor).
			Padding(0, 1)

	// statusErrorStyle is used for error messages in the status bar.
	statusErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true)

	// statusSuccessStyle is used for success messages.
	statusSuccessStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("82"))
)

// Task item styles.
var (
	// selectedTaskStyle highlights the currently selected task.
	selectedTaskStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Background(lipgloss.Color("236")).
				Bold(true)

	// normalTaskStyle is used for non-selected tasks.
	normalTaskStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	// blockedIndicatorStyle styles the [BLOCKED] indicator.
	blockedIndicatorStyle = lipgloss.NewStyle().
				Foreground(priorityUrgent).
				Bold(true)
)

// GetColumnHeaderColor returns the appropriate color for a column header.
// Each column status has a distinct color for quick visual identification.
func GetColumnHeaderColor(status string) lipgloss.Color {
	switch status {
	case "backlog":
		return columnBacklog
	case "todo":
		return columnTodo
	case "in_progress":
		return columnInProgress
	case "need_input":
		return columnNeedInput
	case "review":
		return columnReview
	case "done":
		return columnDone
	default:
		return mutedColor
	}
}

// GetPriorityIndicator returns a styled priority indicator string.
// Higher priority = more prominent indicator.
func GetPriorityIndicator(priority string) string {
	var color lipgloss.Color
	var indicator string

	switch priority {
	case "urgent":
		color = priorityUrgent
		indicator = "!!!"
	case "high":
		color = priorityHigh
		indicator = "!!"
	case "medium":
		color = priorityMedium
		indicator = "!"
	default: // low
		color = priorityLow
		indicator = ""
	}

	if indicator == "" {
		return ""
	}

	return lipgloss.NewStyle().Foreground(color).Render(indicator)
}

// GetTypeIndicator returns a styled type badge.
func GetTypeIndicator(taskType string) string {
	var color lipgloss.Color

	switch taskType {
	case "bug":
		color = typeBug
	case "feature":
		color = typeFeature
	case "chore":
		color = typeChore
	default:
		color = mutedColor
	}

	return lipgloss.NewStyle().Foreground(color).Render("[" + taskType + "]")
}
```

**Key Concepts**:

- Use ANSI 256 colors for broad terminal compatibility
- Define styles as package-level variables for reuse
- Group related colors and styles together
- Provide helper functions for dynamic styling

**Common Mistakes**:
- Using true color (24-bit) which may not work in all terminals
- Hardcoding style values instead of using variables
- Not testing on light/dark terminal backgrounds

---

### 1.6 Implement TaskItem [COMPLETED]

**What**: Create the `TaskItem` type that implements `list.Item` interface.

**Why**: The bubbles/list component requires items to implement `list.Item`. This allows the list to render tasks with titles and descriptions.

**File**: `internal/tui/task_item.go`

```go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pocketbase/pocketbase/core"
)

// TaskItem represents a task in the kanban board.
// It implements the list.Item interface required by bubbles/list.
type TaskItem struct {
	// Core fields from database
	ID          string
	Title       string
	Description string
	Type        string   // bug, feature, chore
	Priority    string   // low, medium, high, urgent
	Column      string   // backlog, todo, in_progress, need_input, review, done
	Labels      []string
	Position    float64

	// Display fields
	DisplayID string // e.g., "WRK-123"

	// Computed fields
	IsBlocked bool
	BlockedBy []string
}

// FilterValue returns the string used for filtering in the list.
// When the user types to filter, this value is searched.
func (t TaskItem) FilterValue() string {
	// Include title and display ID for filtering
	return t.Title + " " + t.DisplayID
}

// Title returns the primary display string for the list item.
// This is rendered as the main line in the list.
func (t TaskItem) Title() string {
	return t.renderTitle()
}

// Description returns the secondary display string.
// Rendered below the title in a dimmer color.
func (t TaskItem) Description() string {
	return t.renderDescription()
}

// renderTitle creates the formatted title line for display.
// Format: [PRIORITY] DISPLAY_ID Title [TYPE] [BLOCKED]
func (t TaskItem) renderTitle() string {
	var parts []string

	// Priority indicator (colored dot or exclamation marks)
	if indicator := GetPriorityIndicator(t.Priority); indicator != "" {
		parts = append(parts, indicator)
	}

	// Display ID in muted color
	idStyle := lipgloss.NewStyle().Foreground(mutedColor)
	parts = append(parts, idStyle.Render(t.DisplayID))

	// Task title (main content)
	parts = append(parts, t.Title)

	// Type badge
	if t.Type != "" {
		parts = append(parts, GetTypeIndicator(t.Type))
	}

	// Blocked indicator
	if t.IsBlocked {
		blocked := blockedIndicatorStyle.Render("[BLOCKED]")
		parts = append(parts, blocked)
	}

	return strings.Join(parts, " ")
}

// renderDescription creates the secondary info line.
// Shows labels and other metadata.
func (t TaskItem) renderDescription() string {
	var parts []string

	// Labels (show first 3)
	if len(t.Labels) > 0 {
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		maxLabels := 3
		if len(t.Labels) < maxLabels {
			maxLabels = len(t.Labels)
		}
		for _, label := range t.Labels[:maxLabels] {
			parts = append(parts, labelStyle.Render("#"+label))
		}
		if len(t.Labels) > 3 {
			parts = append(parts, labelStyle.Render(fmt.Sprintf("+%d", len(t.Labels)-3)))
		}
	}

	// If no parts, return empty
	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, " ")
}

// NewTaskItemFromRecord creates a TaskItem from a PocketBase record.
// This is the primary way to create TaskItems from database queries.
func NewTaskItemFromRecord(record *core.Record, displayID string) TaskItem {
	// Extract labels from the record
	var labels []string
	if rawLabels := record.Get("labels"); rawLabels != nil {
		labels = record.GetStringSlice("labels")
	}

	// Extract blocked_by to determine if task is blocked
	blockedBy := record.GetStringSlice("blocked_by")
	isBlocked := len(blockedBy) > 0

	return TaskItem{
		ID:          record.Id,
		Title:       record.GetString("title"),
		Description: record.GetString("description"),
		Type:        record.GetString("type"),
		Priority:    record.GetString("priority"),
		Column:      record.GetString("column"),
		Labels:      labels,
		Position:    record.GetFloat("position"),
		DisplayID:   displayID,
		IsBlocked:   isBlocked,
		BlockedBy:   blockedBy,
	}
}

// Truncate truncates a string to maxLen, adding "..." if truncated.
// Used to fit long titles in limited space.
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
```

**Key Concepts**:

- `list.Item` interface requires `FilterValue()`, `Title()`, and `Description()`
- The `Title()` return value is what appears in the list
- `NewTaskItemFromRecord` handles conversion from database records

**Expected output** when viewing a task:
```
!! WRK-123 Implement user authentication [feature]
#backend #auth +2
```

**Common Mistakes**:
- Returning very long strings from `Title()` (truncate to fit column width)
- Not handling nil/empty values from the database
- Forgetting to check if labels slice is empty before indexing

---

### 1.7 Implement Column Component [COMPLETED]

**What**: Create the `Column` component that wraps `bubbles/list` for a single kanban column.

**Why**: Each column (backlog, todo, etc.) is a scrollable list of tasks. The `Column` component manages the list state and handles per-column navigation.

**File**: `internal/tui/column.go`

```go
package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Column represents a single column in the kanban board.
// It wraps a bubbles/list for task display and navigation.
type Column struct {
	status  string     // Column identifier: "backlog", "todo", etc.
	title   string     // Display title: "Backlog", "Todo", etc.
	list    list.Model // The bubbles list component
	focused bool       // Whether this column has focus
	width   int        // Current width
	height  int        // Current height
}

// columnTitles maps status values to display titles.
var columnTitles = map[string]string{
	"backlog":     "Backlog",
	"todo":        "Todo",
	"in_progress": "In Progress",
	"need_input":  "Need Input",
	"review":      "Review",
	"done":        "Done",
}

// NewColumn creates a new Column with the given status and items.
// The status determines the column's role (backlog, todo, etc.).
func NewColumn(status string, items []list.Item, focused bool) Column {
	// Get display title from status
	title := columnTitles[status]
	if title == "" {
		title = status // Fallback to status if unknown
	}

	// Create a custom delegate for rendering list items
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.SetHeight(2) // Title + description

	// Customize delegate styles
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color("236")).
		Bold(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("236"))

	// Create the list
	l := list.New(items, delegate, 0, 0)
	l.Title = title
	l.SetShowStatusBar(false) // We have our own status bar
	l.SetFilteringEnabled(false) // Disable for Phase 1
	l.SetShowHelp(false) // We have global help
	l.SetShowTitle(false) // We render title separately

	// Disable keybindings that conflict with our navigation
	l.KeyMap.Quit.SetEnabled(false)

	return Column{
		status:  status,
		title:   title,
		list:    l,
		focused: focused,
	}
}

// Init implements tea.Model (no-op for Column).
func (c Column) Init() tea.Cmd {
	return nil
}

// Update handles messages for this column.
// Most navigation is handled by passing messages to the embedded list.
func (c Column) Update(msg tea.Msg) (Column, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Only process keys if this column is focused
		if !c.focused {
			return c, nil
		}

		// Handle up/down navigation
		switch msg.String() {
		case "k", "up":
			c.list.CursorUp()
		case "j", "down":
			c.list.CursorDown()
		}
	}

	// Pass other messages to the list
	c.list, cmd = c.list.Update(msg)
	return c, cmd
}

// View renders the column.
func (c Column) View() string {
	// Column header with count
	headerColor := GetColumnHeaderColor(c.status)
	headerStyle := columnTitleStyle.Copy().Foreground(headerColor)
	if c.focused {
		headerStyle = headerStyle.Underline(true)
	}

	count := len(c.list.Items())
	header := headerStyle.Render(fmt.Sprintf("%s (%d)", c.title, count))

	// List content
	listView := c.list.View()

	// Empty state
	if count == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true).
			Padding(1, 0)
		listView = emptyStyle.Render("(empty)")
	}

	// Combine header and list
	return lipgloss.JoinVertical(lipgloss.Left, header, listView)
}

// SetSize updates the column dimensions.
// Called when the terminal is resized.
func (c *Column) SetSize(width, height int) {
	c.width = width
	c.height = height

	// Account for header (1 line) and some padding
	listHeight := height - 3
	if listHeight < 1 {
		listHeight = 1
	}

	c.list.SetSize(width-2, listHeight) // -2 for borders
}

// SetFocused updates the column's focus state.
func (c *Column) SetFocused(focused bool) {
	c.focused = focused
}

// IsFocused returns whether this column has focus.
func (c Column) IsFocused() bool {
	return c.focused
}

// Status returns the column's status identifier.
func (c Column) Status() string {
	return c.status
}

// Items returns the items in this column's list.
func (c Column) Items() []list.Item {
	return c.list.Items()
}

// SetItems replaces all items in the column's list.
func (c *Column) SetItems(items []list.Item) {
	c.list.SetItems(items)
}

// SelectedItem returns the currently selected item, or nil if empty.
func (c Column) SelectedItem() list.Item {
	if len(c.list.Items()) == 0 {
		return nil
	}
	return c.list.SelectedItem()
}

// SelectedTask returns the selected TaskItem, or nil if empty.
func (c Column) SelectedTask() *TaskItem {
	item := c.SelectedItem()
	if item == nil {
		return nil
	}
	task, ok := item.(TaskItem)
	if !ok {
		return nil
	}
	return &task
}
```

**Key Concepts**:

- Each `Column` wraps a `list.Model` from bubbles
- The `focused` state determines if the column processes keyboard input
- `SetSize` must be called when terminal dimensions change
- `SelectedTask` returns the currently highlighted task

**Common Mistakes**:
- Not disabling conflicting keybindings in the list
- Forgetting to set size after creation
- Not handling empty list case

---

### 1.8 Implement Async Commands [COMPLETED]

**What**: Create command functions for loading data asynchronously.

**Why**: Bubble Tea uses commands to perform async operations. Commands return messages that update state without blocking the UI.

**File**: `internal/tui/commands.go`

```go
package tui

import (
	"fmt"
	"sort"

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

// boardAndTasksLoadedMsg is sent when both board and tasks are loaded.
// Used for the initial load sequence.
type boardAndTasksLoadedMsg struct {
	board *core.Record
	tasks []*core.Record
}

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
		3*1000000000, // 3 seconds in nanoseconds
		func(_ interface{}) tea.Msg {
			return statusMsg{message: "", isError: false}
		},
	)
}
```

**Key Concepts**:

- Commands are functions that return `tea.Cmd`
- `tea.Cmd` is a function that returns `tea.Msg`
- Commands run asynchronously; the returned message is sent to `Update`
- Combine multiple loads into single commands when possible

**Common Mistakes**:
- Blocking in the Init function instead of using commands
- Not handling errors in commands
- Forgetting to sort tasks by position

---

### 1.9 Implement Main App Model

**What**: Create the main `App` model that ties everything together.

**Why**: The `App` is the root of the Bubble Tea application. It manages columns, handles input routing, and coordinates the overall view.

**File**: `internal/tui/app.go`

```go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
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
	columns      []Column
	focusedCol   int
	columnOrder  []string // Column status order

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
			a.statusMessage = fmt.Sprintf("Selected: %s %s", task.DisplayID, task.Title)
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
```

**Key Concepts**:

- `Init()` returns the first command to execute (loading data)
- `Update()` handles all messages and returns updated model + commands
- `View()` returns the string to render
- Column focus is tracked with `focusedCol` index
- Layout is calculated based on terminal dimensions

**Expected output** (in terminal):
```
EgenSkriven - Work Board (WRK)
+------------+------------+-------------+-------------+--------+------+
| Backlog(3) | Todo (2)   | In Progress | Need Input  | Review | Done |
|            |            | (1)         | (0)         | (0)    | (5)  |
|------------|------------|-------------|-------------|--------|------|
| !! WRK-1   | > WRK-4    | WRK-6       |             |        | ...  |
| Setup      |   Auth     |   API       |             |        |      |
|            |            |             |             |        |      |
| WRK-2      | WRK-5      |             |             |        |      |
| Config     |   Tests    |             |             |        |      |
+------------+------------+-------------+-------------+--------+------+
h/l: columns | j/k: tasks | enter: details | q: quit | ?: help    WRK
```

**Common Mistakes**:
- Not updating column sizes on resize
- Forgetting to handle the not-ready state
- Not returning `tea.Quit` command for quit

---

### 1.10 Implement CLI Command

**What**: Create the `egenskriven tui` command that launches the TUI.

**Why**: The TUI needs to be accessible from the main CLI with proper integration.

**File**: `internal/commands/tui.go`

```go
package commands

import (
	"fmt"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"

	"github.com/ramtinJ95/EgenSkriven/internal/config"
	"github.com/ramtinJ95/EgenSkriven/internal/tui"
)

// newTuiCmd creates the 'tui' command for launching the terminal UI.
func newTuiCmd(app *pocketbase.PocketBase) *cobra.Command {
	var boardRef string

	cmd := &cobra.Command{
		Use:     "tui",
		Aliases: []string{"ui", "board"},
		Short:   "Open interactive kanban board",
		Long: `Launch the terminal user interface for managing tasks in a kanban board view.

The TUI provides a full kanban board experience with:
- Column-based task view (backlog, todo, in_progress, need_input, review, done)
- Vim-style navigation (h/j/k/l)
- Task details view
- Real-time updates when server is running

Examples:
  egenskriven tui
  egenskriven tui --board work
  egenskriven tui -b WRK

Navigation:
  h/l       Move between columns
  j/k       Move between tasks within a column
  Enter     View task details
  q         Quit`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Bootstrap the app to ensure database is ready
			if err := app.Bootstrap(); err != nil {
				return fmt.Errorf("failed to bootstrap: %w", err)
			}

			// Load config for default board if not specified
			if boardRef == "" {
				cfg, _ := config.LoadProjectConfig()
				if cfg != nil && cfg.DefaultBoard != "" {
					boardRef = cfg.DefaultBoard
				}
			}

			// Run the TUI
			return tui.Run(app, boardRef)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&boardRef, "board", "b", "",
		"Board to open (name or prefix)")

	return cmd
}
```

**Register the command** in `internal/commands/root.go`:

Add this line in the `Register` function after the other commands:

```go
// TUI command
app.RootCmd.AddCommand(newTuiCmd(app))
```

**Steps**:

1. Create `internal/commands/tui.go` with the code above.

2. Edit `internal/commands/root.go` to register the command:
   
   Find the `Register` function and add:
   ```go
   // TUI command
   app.RootCmd.AddCommand(newTuiCmd(app))
   ```

3. Build and test:
   ```bash
   go build -o egenskriven ./cmd/egenskriven
   ./egenskriven tui --help
   ```
   
   **Expected output**:
   ```
   Launch the terminal user interface for managing tasks in a kanban board view.
   
   The TUI provides a full kanban board experience with:
   ...
   
   Usage:
     egenskriven tui [flags]
   
   Aliases:
     tui, ui, board
   
   Flags:
     -b, --board string   Board to open (name or prefix)
     -h, --help           help for tui
   ```

4. Run the TUI:
   ```bash
   ./egenskriven tui
   ```
   
   **Expected**: Full-screen TUI with columns and tasks.

**Common Mistakes**:
- Forgetting to register the command in `root.go`
- Not bootstrapping the app before running TUI
- Not checking for terminal support

---

### 1.11 Write Unit Tests

**What**: Create tests for the TUI components.

**Why**: Tests ensure components work correctly and catch regressions.

**File**: `internal/tui/task_item_test.go`

```go
package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTaskItem_FilterValue(t *testing.T) {
	item := TaskItem{
		Title:     "Implement feature",
		DisplayID: "WRK-123",
	}

	value := item.FilterValue()

	assert.Contains(t, value, "Implement feature")
	assert.Contains(t, value, "WRK-123")
}

func TestTaskItem_Title(t *testing.T) {
	tests := []struct {
		name     string
		item     TaskItem
		contains []string
	}{
		{
			name: "basic task",
			item: TaskItem{
				Title:     "Simple task",
				DisplayID: "WRK-1",
				Priority:  "medium",
				Type:      "feature",
			},
			contains: []string{"WRK-1", "Simple task", "[feature]"},
		},
		{
			name: "urgent bug",
			item: TaskItem{
				Title:     "Fix crash",
				DisplayID: "WRK-2",
				Priority:  "urgent",
				Type:      "bug",
			},
			contains: []string{"WRK-2", "Fix crash", "[bug]", "!!!"},
		},
		{
			name: "blocked task",
			item: TaskItem{
				Title:     "Blocked task",
				DisplayID: "WRK-3",
				Priority:  "low",
				Type:      "chore",
				IsBlocked: true,
			},
			contains: []string{"WRK-3", "Blocked task", "[BLOCKED]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title := tt.item.Title()
			for _, s := range tt.contains {
				assert.Contains(t, title, s)
			}
		})
	}
}

func TestTaskItem_Description(t *testing.T) {
	tests := []struct {
		name     string
		item     TaskItem
		expected string
	}{
		{
			name: "no labels",
			item: TaskItem{
				Labels: []string{},
			},
			expected: "",
		},
		{
			name: "with labels",
			item: TaskItem{
				Labels: []string{"backend", "auth"},
			},
			expected: "#backend #auth",
		},
		{
			name: "too many labels truncated",
			item: TaskItem{
				Labels: []string{"one", "two", "three", "four", "five"},
			},
			expected: "#one #two #three +2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc := tt.item.Description()
			if tt.expected == "" {
				assert.Empty(t, desc)
			} else {
				for _, part := range []string{"#one", "#two", "#three"} {
					if len(tt.item.Labels) >= 3 {
						assert.Contains(t, desc, part)
					}
				}
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "he..."},
		{"hello world", 8, "hello..."},
		{"hi", 3, "hi"},
		{"abc", 3, "abc"},
		{"abcd", 3, "abc"}, // edge case: maxLen <= 3
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := Truncate(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}
```

**File**: `internal/tui/column_test.go`

```go
package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/stretchr/testify/assert"
)

func TestNewColumn(t *testing.T) {
	items := []list.Item{
		TaskItem{ID: "1", Title: "Task 1", DisplayID: "WRK-1"},
		TaskItem{ID: "2", Title: "Task 2", DisplayID: "WRK-2"},
	}

	col := NewColumn("todo", items, true)

	assert.Equal(t, "todo", col.Status())
	assert.True(t, col.IsFocused())
	assert.Len(t, col.Items(), 2)
}

func TestColumn_SetFocused(t *testing.T) {
	col := NewColumn("backlog", nil, false)

	assert.False(t, col.IsFocused())

	col.SetFocused(true)
	assert.True(t, col.IsFocused())

	col.SetFocused(false)
	assert.False(t, col.IsFocused())
}

func TestColumn_SelectedTask(t *testing.T) {
	// Empty column
	col := NewColumn("todo", nil, true)
	assert.Nil(t, col.SelectedTask())

	// Column with tasks
	items := []list.Item{
		TaskItem{ID: "1", Title: "Task 1", DisplayID: "WRK-1"},
	}
	col = NewColumn("todo", items, true)

	task := col.SelectedTask()
	assert.NotNil(t, task)
	assert.Equal(t, "1", task.ID)
}

func TestColumn_View(t *testing.T) {
	items := []list.Item{
		TaskItem{ID: "1", Title: "Task 1", DisplayID: "WRK-1"},
	}
	col := NewColumn("todo", items, true)
	col.SetSize(30, 20)

	view := col.View()

	// Should contain column title with count
	assert.Contains(t, view, "Todo")
	assert.Contains(t, view, "(1)")
}

func TestColumn_EmptyView(t *testing.T) {
	col := NewColumn("review", nil, false)
	col.SetSize(30, 20)

	view := col.View()

	assert.Contains(t, view, "Review")
	assert.Contains(t, view, "(0)")
	assert.Contains(t, view, "(empty)")
}
```

**File**: `internal/tui/styles_test.go`

```go
package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetColumnHeaderColor(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"backlog", "240"},
		{"todo", "39"},
		{"in_progress", "214"},
		{"need_input", "196"},
		{"review", "205"},
		{"done", "82"},
		{"unknown", "240"}, // defaults to muted
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			color := GetColumnHeaderColor(tt.status)
			assert.Equal(t, tt.expected, string(color))
		})
	}
}

func TestGetPriorityIndicator(t *testing.T) {
	tests := []struct {
		priority string
		contains string
	}{
		{"urgent", "!!!"},
		{"high", "!!"},
		{"medium", "!"},
		{"low", ""},
	}

	for _, tt := range tests {
		t.Run(tt.priority, func(t *testing.T) {
			indicator := GetPriorityIndicator(tt.priority)
			if tt.contains == "" {
				assert.Empty(t, indicator)
			} else {
				assert.Contains(t, indicator, tt.contains)
			}
		})
	}
}

func TestGetTypeIndicator(t *testing.T) {
	tests := []struct {
		taskType string
		contains string
	}{
		{"bug", "[bug]"},
		{"feature", "[feature]"},
		{"chore", "[chore]"},
	}

	for _, tt := range tests {
		t.Run(tt.taskType, func(t *testing.T) {
			indicator := GetTypeIndicator(tt.taskType)
			assert.Contains(t, indicator, tt.contains)
		})
	}
}
```

**Run the tests**:

```bash
go test ./internal/tui/... -v
```

**Expected output**:
```
=== RUN   TestTaskItem_FilterValue
--- PASS: TestTaskItem_FilterValue (0.00s)
=== RUN   TestTaskItem_Title
=== RUN   TestTaskItem_Title/basic_task
=== RUN   TestTaskItem_Title/urgent_bug
=== RUN   TestTaskItem_Title/blocked_task
--- PASS: TestTaskItem_Title (0.00s)
...
PASS
ok      github.com/ramtinJ95/EgenSkriven/internal/tui
```

**Common Mistakes**:
- Testing private functions (test through public interface)
- Not testing edge cases (empty lists, nil values)

---

## Verification Checklist

Complete each section in order. Check off each item as you verify it.

### Build Verification

- [ ] **Dependencies installed**
  ```bash
  go mod tidy
  ```
  Should complete without errors.

- [ ] **Project compiles**
  ```bash
  go build ./...
  ```
  Should complete without errors.

- [ ] **TUI package compiles**
  ```bash
  go build ./internal/tui
  ```
  Should complete without errors.

### Command Verification

- [ ] **TUI command available**
  ```bash
  ./egenskriven --help
  ```
  Should show `tui` in the command list.

- [ ] **TUI help works**
  ```bash
  ./egenskriven tui --help
  ```
  Should show usage and flags.

### TUI Verification

- [ ] **TUI launches**
  ```bash
  ./egenskriven tui
  ```
  Should show full-screen kanban board.

- [ ] **Columns displayed**
  
  Should see 6 columns: Backlog, Todo, In Progress, Need Input, Review, Done.

- [ ] **Tasks displayed**
  
  Tasks should appear in their respective columns with display IDs.

- [ ] **Column navigation works**
  
  Press `h` and `l` (or arrow keys) to move focus between columns.
  The focused column should have a different border color.

- [ ] **Task navigation works**
  
  Press `j` and `k` (or arrow keys) to move selection within a column.
  The selected task should be highlighted.

- [ ] **Quit works**
  
  Press `q` or `Ctrl+C` to exit the TUI.
  Should return to shell cleanly.

- [ ] **Board flag works**
  ```bash
  ./egenskriven tui --board WRK
  ```
  Should open the specified board.

### Test Verification

- [ ] **Tests pass**
  ```bash
  go test ./internal/tui/... -v
  ```
  All tests should pass.

- [ ] **No race conditions**
  ```bash
  go test ./internal/tui/... -race
  ```
  Should complete without race warnings.

---

## File Summary

| File | Lines | Purpose |
|------|-------|---------|
| `internal/tui/messages.go` | ~50 | Custom message types |
| `internal/tui/keys.go` | ~80 | Keybinding definitions |
| `internal/tui/styles.go` | ~120 | Lip Gloss style definitions |
| `internal/tui/task_item.go` | ~100 | TaskItem list.Item implementation |
| `internal/tui/column.go` | ~150 | Column component |
| `internal/tui/commands.go` | ~100 | Async command functions |
| `internal/tui/app.go` | ~350 | Main application model |
| `internal/commands/tui.go` | ~60 | CLI command |
| `internal/tui/task_item_test.go` | ~80 | TaskItem tests |
| `internal/tui/column_test.go` | ~60 | Column tests |
| `internal/tui/styles_test.go` | ~50 | Style tests |

**Total new code**: ~1200 lines

---

## What You Should Have Now

After completing Phase 1, your project should:

```
internal/
├── tui/
│   ├── app.go              ✓ Main application model
│   ├── board.go            ✓ Empty (Phase 3)
│   ├── column.go           ✓ Column component
│   ├── task_item.go        ✓ TaskItem implementation
│   ├── keys.go             ✓ Keybindings
│   ├── styles.go           ✓ Lip Gloss styles
│   ├── messages.go         ✓ Message types
│   ├── commands.go         ✓ Async commands
│   ├── task_item_test.go   ✓ Tests
│   ├── column_test.go      ✓ Tests
│   └── styles_test.go      ✓ Tests
├── commands/
│   ├── tui.go              ✓ CLI command
│   └── root.go             ✓ Updated to register TUI
```

You should be able to:

1. Run `egenskriven tui` and see a full-screen kanban board
2. Navigate between columns with `h` and `l`
3. Navigate between tasks with `j` and `k`
4. See tasks with their display IDs, types, and priorities
5. See blocked task indicators
6. Exit cleanly with `q`

---

## Next Phase

**Phase 2: Task Operations** will add:

- Task detail view (Enter to open)
- Task creation form (n key)
- Task editing (e key)
- Task deletion with confirmation (d key)
- Task movement between columns (H/L, 1-5)
- Task reordering within column (Shift+J/K)
- Success/error feedback messages

---

## Troubleshooting

### "panic: terminal too small"

**Problem**: Terminal window is too small for the TUI.

**Solution**: Resize your terminal to at least 80x24 characters. The TUI requires minimum dimensions to display columns properly.

### "no boards found"

**Problem**: No boards exist in the database.

**Solution**: Create a board first:
```bash
./egenskriven board create "Work" --prefix WRK
```

### "tasks collection not found"

**Problem**: Database hasn't been migrated.

**Solution**: Run migrations:
```bash
./egenskriven serve  # Start once to run migrations
# Then Ctrl+C to stop
./egenskriven tui    # Try TUI again
```

### Keys not working

**Problem**: Key presses don't navigate or respond.

**Solution**: 
- Ensure the terminal supports keyboard input
- Check if running in a Docker container (may need `-it` flag)
- Try different terminal emulator

### Colors look wrong

**Problem**: Colors appear incorrect or not at all.

**Solution**:
- Ensure `TERM` environment variable is set correctly
- Try `export TERM=xterm-256color`
- Some minimal terminals don't support 256 colors

### Tests fail with import errors

**Problem**: Tests can't find packages.

**Solution**:
```bash
go mod tidy
go mod verify
```

### High CPU usage

**Problem**: TUI uses excessive CPU.

**Solution**: This is usually caused by an infinite loop in the Update function. Check that:
- You're not sending commands that trigger themselves
- Tick commands have proper intervals

---

## Glossary

| Term | Definition |
|------|------------|
| **Bubble Tea** | Go framework for building terminal UIs using Elm-style architecture |
| **bubbles** | Pre-built components for Bubble Tea (list, textinput, viewport) |
| **Lip Gloss** | Go library for terminal styling (colors, borders, padding) |
| **Model** | The data structure representing application state |
| **Update** | Function that handles messages and updates the Model |
| **View** | Function that renders the Model to a string |
| **Cmd** | A function that produces a Msg, used for async operations |
| **Msg** | A message that triggers state updates |
| **tea.Program** | The main Bubble Tea runtime that drives the application |
| **Alt Screen** | Terminal mode that provides full-screen canvas |
| **list.Item** | Interface that items must implement to be displayed in bubbles/list |
