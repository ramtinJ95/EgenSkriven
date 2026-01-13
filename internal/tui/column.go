package tui

import (
	"fmt"
	"sort"

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
	l.SetShowStatusBar(false)    // We have our own status bar
	l.SetFilteringEnabled(false) // Disable for Phase 1
	l.SetShowHelp(false)         // We have global help
	l.SetShowTitle(false)        // We render title separately

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
	headerStyle := columnTitleStyle.Foreground(headerColor)
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

// =============================================================================
// Incremental Update Methods (for realtime sync)
// =============================================================================

// InsertTask adds a task to the column in sorted position order.
// Used for realtime task creation events.
func (c *Column) InsertTask(task TaskItem) {
	items := c.list.Items()

	// Find insertion point based on position
	insertAt := len(items)
	for i, item := range items {
		if t, ok := item.(TaskItem); ok && t.Position > task.Position {
			insertAt = i
			break
		}
	}

	// Insert at position
	newItems := make([]list.Item, 0, len(items)+1)
	newItems = append(newItems, items[:insertAt]...)
	newItems = append(newItems, task)
	newItems = append(newItems, items[insertAt:]...)

	c.list.SetItems(newItems)
}

// UpdateTask updates a task at the given index.
// Used for realtime task update events.
func (c *Column) UpdateTask(index int, task TaskItem) {
	items := c.list.Items()
	if index < 0 || index >= len(items) {
		return
	}

	items[index] = task

	// Check if position changed - if so, re-sort
	needsSort := false
	if index > 0 {
		if prev, ok := items[index-1].(TaskItem); ok && prev.Position > task.Position {
			needsSort = true
		}
	}
	if index < len(items)-1 {
		if next, ok := items[index+1].(TaskItem); ok && next.Position < task.Position {
			needsSort = true
		}
	}

	if needsSort {
		// Re-sort by position
		taskItems := make([]TaskItem, 0, len(items))
		for _, item := range items {
			if t, ok := item.(TaskItem); ok {
				taskItems = append(taskItems, t)
			}
		}
		sort.Slice(taskItems, func(i, j int) bool {
			return taskItems[i].Position < taskItems[j].Position
		})
		newItems := make([]list.Item, len(taskItems))
		for i, t := range taskItems {
			newItems[i] = t
		}
		items = newItems
	}

	c.list.SetItems(items)
}

// RemoveTask removes a task at the given index.
// Used for realtime task deletion events.
func (c *Column) RemoveTask(index int) {
	items := c.list.Items()
	if index < 0 || index >= len(items) {
		return
	}

	newItems := make([]list.Item, 0, len(items)-1)
	newItems = append(newItems, items[:index]...)
	newItems = append(newItems, items[index+1:]...)

	c.list.SetItems(newItems)
}

// FindTaskByID finds a task by its ID and returns its index.
// Returns -1 if not found.
func (c Column) FindTaskByID(taskID string) int {
	for i, item := range c.list.Items() {
		if task, ok := item.(TaskItem); ok && task.ID == taskID {
			return i
		}
	}
	return -1
}
