# Phase 6: Advanced Features

**Goal**: Feature parity with web UI - epics, blocked tasks, due dates, subtasks, multi-select, and bulk operations.

**Duration Estimate**: 3-4 days

**Prerequisites**: Phase 5 (Filtering & Search) completed

**Deliverable**: Full feature parity with web UI including all advanced task features and bulk operations.

---

## Overview

Phase 6 transforms the TUI from a basic kanban interface into a fully-featured task management tool. This phase adds visual indicators for task dependencies and deadlines, expandable subtask hierarchies, and powerful multi-select operations for managing tasks in bulk.

### Why These Features Matter

1. **Epics** - Group related tasks visually with color-coded badges, filter by epic
2. **Blocked Indicator** - Immediately see which tasks can't be started yet
3. **Due Dates** - Surface overdue tasks with urgent red highlighting
4. **Subtasks** - View task hierarchies without leaving the board view
5. **Multi-Select** - Select multiple tasks for bulk operations (saves time)
6. **Bulk Operations** - Move or delete many tasks at once
7. **Help Overlay** - Discoverability for all keyboard shortcuts
8. **Command Palette** - Quick access to any action (optional but powerful)

### Data Model Reference

From the migrations, here are the relevant fields:

```
tasks collection:
  - blocked_by: JSONField (array of task IDs)
  - due_date: DateField (ISO 8601: YYYY-MM-DD)
  - parent: RelationField (self-reference to tasks.id)
  - epic: RelationField (to epics.id)

epics collection:
  - title: TextField (required, max 200)
  - description: TextField (max 5000)
  - color: TextField (hex color, e.g., "#3B82F6")
  - board: RelationField (to boards.id)
```

---

## Tasks

### 6.1 Epic Display and Filtering

**What**: Show epic badge on task cards with color coding, add epic filter.

**Why**: Epics group related tasks across columns. Visual color coding helps users quickly identify which tasks belong to which epic.

**Steps**:

1. Create `internal/tui/epic.go` with epic data structures
2. Add epic badge rendering to task items
3. Implement epic loading with board tasks
4. Add epic filter (fe key binding)
5. Show active epic filter in filter bar

**File**: `internal/tui/epic.go`

```go
package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// Epic represents an epic for display in the TUI.
type Epic struct {
	ID          string
	Title       string
	Description string
	Color       string // Hex color like "#3B82F6"
	BoardID     string
}

// EpicCache holds loaded epics for quick lookup.
type EpicCache struct {
	epics    map[string]Epic // keyed by epic ID
	byBoard  map[string][]Epic
}

// NewEpicCache creates an empty epic cache.
func NewEpicCache() *EpicCache {
	return &EpicCache{
		epics:   make(map[string]Epic),
		byBoard: make(map[string][]Epic),
	}
}

// Load fetches all epics from the database.
func (c *EpicCache) Load(app *pocketbase.PocketBase) error {
	records, err := app.FindAllRecords("epics")
	if err != nil {
		return fmt.Errorf("failed to load epics: %w", err)
	}

	c.epics = make(map[string]Epic)
	c.byBoard = make(map[string][]Epic)

	for _, r := range records {
		epic := epicFromRecord(r)
		c.epics[epic.ID] = epic

		boardID := epic.BoardID
		c.byBoard[boardID] = append(c.byBoard[boardID], epic)
	}

	return nil
}

// Get returns an epic by ID, or nil if not found.
func (c *EpicCache) Get(id string) *Epic {
	if epic, ok := c.epics[id]; ok {
		return &epic
	}
	return nil
}

// ForBoard returns all epics for a given board.
func (c *EpicCache) ForBoard(boardID string) []Epic {
	return c.byBoard[boardID]
}

// epicFromRecord converts a PocketBase record to Epic.
func epicFromRecord(r *core.Record) Epic {
	return Epic{
		ID:          r.Id,
		Title:       r.GetString("title"),
		Description: r.GetString("description"),
		Color:       r.GetString("color"),
		BoardID:     r.GetString("board"),
	}
}

// EpicBadgeStyle creates a styled badge for an epic.
// Uses the epic's color as background with contrasting text.
func EpicBadgeStyle(epic *Epic) lipgloss.Style {
	if epic == nil {
		return lipgloss.NewStyle()
	}

	// Parse hex color for background
	bgColor := lipgloss.Color(epic.Color)
	
	// Use white text for darker colors, black for lighter
	// Simple heuristic: if color starts with dark hex values
	fgColor := lipgloss.Color("#FFFFFF")
	if len(epic.Color) == 7 {
		// Very rough brightness check
		r := epic.Color[1:3]
		if r >= "AA" {
			fgColor = lipgloss.Color("#000000")
		}
	}

	return lipgloss.NewStyle().
		Background(bgColor).
		Foreground(fgColor).
		Padding(0, 1).
		Bold(true)
}

// RenderEpicBadge renders a compact epic badge for task cards.
func RenderEpicBadge(epic *Epic, maxWidth int) string {
	if epic == nil {
		return ""
	}

	title := epic.Title
	if len(title) > maxWidth-2 {
		title = title[:maxWidth-5] + "..."
	}

	return EpicBadgeStyle(epic).Render(title)
}

// EpicFilterOption represents an epic option in the filter selector.
type EpicFilterOption struct {
	Epic  Epic
	Count int // Number of tasks in this epic
}

// FilterValue implements list.Item for epic filtering.
func (o EpicFilterOption) FilterValue() string {
	return o.Epic.Title
}

// Title implements list.Item.
func (o EpicFilterOption) Title() string {
	return fmt.Sprintf("%s (%d)", o.Epic.Title, o.Count)
}

// Description implements list.Item.
func (o EpicFilterOption) Description() string {
	return o.Epic.Description
}
```

**Update TaskItem** to show epic badge:

**File**: `internal/tui/task_item.go` (additions)

```go
package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// TaskItem represents a task for display in the TUI.
// This extends the base structure with epic reference.
type TaskItem struct {
	ID             string
	DisplayID      string    // "WRK-123"
	Title          string
	Description    string
	Type           string    // bug, feature, chore
	Priority       string    // low, medium, high, urgent
	Column         string
	Position       float64
	Labels         []string
	BlockedBy      []string  // Task IDs that block this task
	DueDate        string    // ISO 8601: YYYY-MM-DD
	EpicID         string    // Epic ID reference
	Epic           *Epic     // Resolved epic (for display)
	ParentID       string    // Parent task ID (for subtasks)
	HasSubtasks    bool      // True if this task has children
	SubtaskCount   int       // Number of subtasks
	SubtasksExpanded bool    // True if subtasks are visible
	Subtasks       []TaskItem // Child tasks (when expanded)
	Selected       bool      // Multi-select state
}

// Implement list.Item interface

func (t TaskItem) FilterValue() string {
	parts := []string{t.Title, t.Description, t.DisplayID}
	if t.Epic != nil {
		parts = append(parts, t.Epic.Title)
	}
	parts = append(parts, t.Labels...)
	return strings.Join(parts, " ")
}

// renderTitle builds the task title with all indicators.
func (t TaskItem) renderTitle() string {
	var parts []string

	// Selection indicator (for multi-select)
	if t.Selected {
		checkmark := lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true).
			Render("[x]")
		parts = append(parts, checkmark)
	}

	// Priority indicator
	parts = append(parts, t.renderPriorityDot())

	// Display ID
	id := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(t.DisplayID)
	parts = append(parts, id)

	// Subtask indicator
	if t.HasSubtasks {
		indicator := "+"
		if t.SubtasksExpanded {
			indicator = "-"
		}
		subtaskBadge := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Render(fmt.Sprintf("[%s%d]", indicator, t.SubtaskCount))
		parts = append(parts, subtaskBadge)
	}

	// Title
	title := truncate(t.Title, 40)
	parts = append(parts, title)

	// Type badge
	if t.Type != "" {
		parts = append(parts, t.renderTypeBadge())
	}

	// Blocked indicator
	blocked := t.renderBlockedIndicator()
	if blocked != "" {
		parts = append(parts, blocked)
	}

	// Due date (if set)
	dueDate := t.renderDueDate()
	if dueDate != "" {
		parts = append(parts, dueDate)
	}

	return strings.Join(parts, " ")
}

// renderDescription builds the description line with epic and labels.
func (t TaskItem) renderDescription() string {
	var parts []string

	// Epic badge
	if t.Epic != nil {
		epicBadge := RenderEpicBadge(t.Epic, 15)
		parts = append(parts, epicBadge)
	}

	// Labels (show first 3)
	if len(t.Labels) > 0 {
		labelStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("213"))
		for i, label := range t.Labels {
			if i >= 3 {
				parts = append(parts, labelStyle.Render(fmt.Sprintf("+%d", len(t.Labels)-3)))
				break
			}
			parts = append(parts, labelStyle.Render("#"+label))
		}
	}

	// Blocking info
	if len(t.BlockedBy) > 0 {
		blockedStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Faint(true)
		parts = append(parts, blockedStyle.Render(
			fmt.Sprintf("blocked by %d task(s)", len(t.BlockedBy))))
	}

	return strings.Join(parts, " ")
}

func (t TaskItem) renderPriorityDot() string {
	colors := map[string]string{
		"urgent": "196", // red
		"high":   "208", // orange
		"medium": "226", // yellow
		"low":    "240", // gray
	}
	color, ok := colors[t.Priority]
	if !ok {
		color = "240"
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Render("*")
}

func (t TaskItem) renderTypeBadge() string {
	colors := map[string]string{
		"bug":     "196", // red
		"feature": "39",  // cyan
		"chore":   "240", // gray
	}
	color, ok := colors[t.Type]
	if !ok {
		color = "240"
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Render("[" + t.Type + "]")
}

// truncate shortens a string to max length with ellipsis.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
```

**Expected output**:

```
* WRK-42 [+3] Implement auth flow [feature] [BLOCKED] Due: 2025-01-15
  [Auth Epic] #security #api blocked by 2 task(s)
```

**Common mistakes**:
- Forgetting to load epics when loading tasks
- Not handling nil epic pointer when rendering
- Using wrong color format (must be hex with #)

---

### 6.2 Blocked Tasks Indicator

**What**: Show prominent [BLOCKED] indicator when task has blockers, show blocker list in description.

**Why**: Blocked tasks can't be worked on until dependencies complete. This visual indicator prevents wasted effort.

**Steps**:

1. Create `internal/tui/blocked.go` with blocking logic
2. Add renderBlockedIndicator to TaskItem
3. Show blocking task IDs in description
4. Add "blocked" filter option

**File**: `internal/tui/blocked.go`

```go
package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/pocketbase/pocketbase"
)

// BlockedStyle is used for all blocked indicators.
var BlockedStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("196")).
	Bold(true)

// BlockedFaintStyle is used for secondary blocked info.
var BlockedFaintStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("196")).
	Faint(true)

// IsBlocked returns true if the task has any blockers.
func (t TaskItem) IsBlocked() bool {
	return len(t.BlockedBy) > 0
}

// renderBlockedIndicator returns the [BLOCKED] badge if blocked.
func (t TaskItem) renderBlockedIndicator() string {
	if !t.IsBlocked() {
		return ""
	}
	return BlockedStyle.Render("[BLOCKED]")
}

// renderBlockedByList returns a formatted list of blocking task IDs.
func (t TaskItem) renderBlockedByList() string {
	if !t.IsBlocked() {
		return ""
	}

	return BlockedFaintStyle.Render(
		fmt.Sprintf("blocked by: %v", t.BlockedBy))
}

// BlockingInfo contains resolved information about blockers.
type BlockingInfo struct {
	BlockerID     string
	BlockerTitle  string
	BlockerColumn string
	IsDone        bool // True if blocker is in "done" column
}

// ResolveBlockers looks up blocking task information.
// Returns detailed info about each blocker for display.
func ResolveBlockers(app *pocketbase.PocketBase, blockedBy []string) ([]BlockingInfo, error) {
	if len(blockedBy) == 0 {
		return nil, nil
	}

	var result []BlockingInfo
	for _, id := range blockedBy {
		record, err := app.FindRecordById("tasks", id)
		if err != nil {
			// Blocker might have been deleted
			result = append(result, BlockingInfo{
				BlockerID:    id,
				BlockerTitle: "(deleted)",
				IsDone:       true, // Treat deleted as resolved
			})
			continue
		}

		result = append(result, BlockingInfo{
			BlockerID:     record.Id,
			BlockerTitle:  record.GetString("title"),
			BlockerColumn: record.GetString("column"),
			IsDone:        record.GetString("column") == "done",
		})
	}

	return result, nil
}

// RenderBlockersList creates a detailed blocker list for the task detail view.
func RenderBlockersList(blockers []BlockingInfo, width int) string {
	if len(blockers) == 0 {
		return ""
	}

	header := BlockedStyle.Render("Blocked By:")
	
	var lines []string
	lines = append(lines, header)

	for _, b := range blockers {
		status := BlockedStyle.Render("*")
		if b.IsDone {
			status = lipgloss.NewStyle().
				Foreground(lipgloss.Color("82")).
				Render("*")
		}

		title := truncate(b.BlockerTitle, width-10)
		column := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("[" + b.BlockerColumn + "]")

		lines = append(lines, fmt.Sprintf("  %s %s %s", status, title, column))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// FilterBlockedTasks filters a task list to only blocked tasks.
func FilterBlockedTasks(tasks []TaskItem) []TaskItem {
	var blocked []TaskItem
	for _, t := range tasks {
		if t.IsBlocked() {
			blocked = append(blocked, t)
		}
	}
	return blocked
}

// FilterReadyTasks filters a task list to only unblocked tasks.
func FilterReadyTasks(tasks []TaskItem) []TaskItem {
	var ready []TaskItem
	for _, t := range tasks {
		if !t.IsBlocked() {
			ready = append(ready, t)
		}
	}
	return ready
}
```

**Expected output** (task card):

```
* WRK-15 Fix login timeout [bug] [BLOCKED]
  blocked by: [abc123, def456]
```

**Expected output** (detail view):

```
Blocked By:
  * Implement session refresh [in_progress]
  * Add retry logic [todo]
```

**Common mistakes**:
- Not handling deleted blocker tasks gracefully
- Showing raw IDs instead of resolving to titles in detail view
- Not updating blocked status when blockers are completed

---

### 6.3 Due Date with Overdue Highlighting

**What**: Display due dates with color-coded urgency (red=overdue, yellow=due soon, gray=future).

**Why**: Overdue tasks need immediate attention. Visual highlighting ensures they don't get lost.

**Steps**:

1. Create `internal/tui/due_date.go` with date parsing and styling
2. Add renderDueDate to TaskItem
3. Implement due date filters (overdue, due-today, due-soon)

**File**: `internal/tui/due_date.go`

```go
package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// DueDateUrgency represents the urgency level of a due date.
type DueDateUrgency int

const (
	DueDateNone     DueDateUrgency = iota // No due date set
	DueDateFuture                         // More than 7 days away
	DueDateUpcoming                       // 2-7 days away
	DueDateSoon                           // Tomorrow or day after
	DueDateToday                          // Due today
	DueDateOverdue                        // Past due
)

// DueDateStyles maps urgency to styling.
var DueDateStyles = map[DueDateUrgency]lipgloss.Style{
	DueDateNone: lipgloss.NewStyle(),
	DueDateFuture: lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")), // gray
	DueDateUpcoming: lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")), // cyan
	DueDateSoon: lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")), // yellow
	DueDateToday: lipgloss.NewStyle().
		Foreground(lipgloss.Color("208")). // orange
		Bold(true),
	DueDateOverdue: lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")). // red
		Bold(true),
}

// ParseDueDate parses an ISO 8601 date string.
// Returns zero time if empty or invalid.
func ParseDueDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}

	// Try full datetime first
	t, err := time.Parse(time.RFC3339, dateStr)
	if err == nil {
		return t
	}

	// Try date only
	t, err = time.Parse("2006-01-02", dateStr)
	if err == nil {
		return t
	}

	// Try datetime without timezone
	t, err = time.Parse("2006-01-02 15:04:05", dateStr)
	if err == nil {
		return t
	}

	return time.Time{}
}

// GetDueDateUrgency calculates the urgency level of a due date.
func GetDueDateUrgency(dueDate time.Time) DueDateUrgency {
	if dueDate.IsZero() {
		return DueDateNone
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	due := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, dueDate.Location())

	days := int(due.Sub(today).Hours() / 24)

	switch {
	case days < 0:
		return DueDateOverdue
	case days == 0:
		return DueDateToday
	case days <= 2:
		return DueDateSoon
	case days <= 7:
		return DueDateUpcoming
	default:
		return DueDateFuture
	}
}

// FormatDueDate formats a due date for display with relative indicators.
func FormatDueDate(dueDate time.Time, urgency DueDateUrgency) string {
	if dueDate.IsZero() {
		return ""
	}

	dateStr := dueDate.Format("Jan 2")

	switch urgency {
	case DueDateOverdue:
		days := int(time.Since(dueDate).Hours() / 24)
		if days == 1 {
			return fmt.Sprintf("OVERDUE (1 day)")
		}
		return fmt.Sprintf("OVERDUE (%d days)", days)
	case DueDateToday:
		return "TODAY"
	case DueDateSoon:
		days := int(time.Until(dueDate).Hours() / 24)
		if days == 1 {
			return "Tomorrow"
		}
		return fmt.Sprintf("In %d days", days)
	case DueDateUpcoming:
		return dateStr
	default:
		return dateStr
	}
}

// renderDueDate renders the due date with appropriate styling.
func (t TaskItem) renderDueDate() string {
	dueDate := ParseDueDate(t.DueDate)
	if dueDate.IsZero() {
		return ""
	}

	urgency := GetDueDateUrgency(dueDate)
	style := DueDateStyles[urgency]
	text := FormatDueDate(dueDate, urgency)

	// Add icon based on urgency
	var icon string
	switch urgency {
	case DueDateOverdue:
		icon = "!! "
	case DueDateToday:
		icon = "! "
	case DueDateSoon:
		icon = "> "
	default:
		icon = ""
	}

	return style.Render(icon + text)
}

// DueDateFilter represents filter options for due dates.
type DueDateFilter int

const (
	DueDateFilterAll DueDateFilter = iota
	DueDateFilterOverdue
	DueDateFilterToday
	DueDateFilterThisWeek
	DueDateFilterHasDue
	DueDateFilterNoDue
)

// FilterByDueDate filters tasks by due date criteria.
func FilterByDueDate(tasks []TaskItem, filter DueDateFilter) []TaskItem {
	if filter == DueDateFilterAll {
		return tasks
	}

	var result []TaskItem
	for _, t := range tasks {
		dueDate := ParseDueDate(t.DueDate)
		urgency := GetDueDateUrgency(dueDate)

		switch filter {
		case DueDateFilterOverdue:
			if urgency == DueDateOverdue {
				result = append(result, t)
			}
		case DueDateFilterToday:
			if urgency == DueDateToday {
				result = append(result, t)
			}
		case DueDateFilterThisWeek:
			if urgency == DueDateToday || urgency == DueDateSoon || urgency == DueDateUpcoming {
				result = append(result, t)
			}
		case DueDateFilterHasDue:
			if !dueDate.IsZero() {
				result = append(result, t)
			}
		case DueDateFilterNoDue:
			if dueDate.IsZero() {
				result = append(result, t)
			}
		}
	}

	return result
}
```

**Expected output**:

```
# Overdue task
* WRK-10 Update docs [chore] !! OVERDUE (3 days)

# Due today
* WRK-11 Fix bug [bug] ! TODAY

# Due soon
* WRK-12 Add tests [chore] > Tomorrow

# Future date
* WRK-13 Refactor API [feature] Jan 20
```

**Common mistakes**:
- Not handling timezone differences correctly
- Parsing date strings in wrong format
- Not accounting for time portion of datetime

---

### 6.4 Subtask Display (Expandable)

**What**: Show subtask indicator on parent tasks, allow expanding to see children inline.

**Why**: Subtasks break down complex work. Users need to see the hierarchy without opening each task.

**Steps**:

1. Create `internal/tui/subtask.go` with subtask logic
2. Add subtask indicator [+N] to parent tasks
3. Implement expand/collapse (Enter or Tab on parent)
4. Render subtasks indented below parent
5. Handle navigation into subtask list

**File**: `internal/tui/subtask.go`

```go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
)

// SubtaskIndicator renders the expandable subtask indicator.
type SubtaskIndicator struct {
	Count    int
	Expanded bool
}

// Render creates the subtask indicator string.
func (s SubtaskIndicator) Render() string {
	if s.Count == 0 {
		return ""
	}

	icon := "+"
	if s.Expanded {
		icon = "-"
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true).
		Render(fmt.Sprintf("[%s%d]", icon, s.Count))
}

// LoadSubtasks loads child tasks for a parent task.
func LoadSubtasks(app *pocketbase.PocketBase, parentID string) ([]TaskItem, error) {
	records, err := app.FindAllRecords("tasks",
		dbx.NewExp("parent = {:parent}", dbx.Params{"parent": parentID}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load subtasks: %w", err)
	}

	var subtasks []TaskItem
	for _, r := range records {
		subtasks = append(subtasks, taskItemFromRecord(r, nil))
	}

	return subtasks, nil
}

// CountSubtasks counts direct children for a task.
func CountSubtasks(app *pocketbase.PocketBase, parentID string) (int, error) {
	records, err := app.FindAllRecords("tasks",
		dbx.NewExp("parent = {:parent}", dbx.Params{"parent": parentID}),
	)
	if err != nil {
		return 0, err
	}
	return len(records), nil
}

// GetSubtaskCounts returns a map of parentID -> subtask count for all tasks.
// This is more efficient than counting one at a time.
func GetSubtaskCounts(app *pocketbase.PocketBase, boardID string) (map[string]int, error) {
	counts := make(map[string]int)

	// Query all tasks with parents in this board
	records, err := app.FindAllRecords("tasks",
		dbx.NewExp("board = {:board} AND parent != '' AND parent IS NOT NULL",
			dbx.Params{"board": boardID}),
	)
	if err != nil {
		return nil, err
	}

	for _, r := range records {
		parentID := r.GetString("parent")
		if parentID != "" {
			counts[parentID]++
		}
	}

	return counts, nil
}

// SubtaskTreeNode represents a task in the subtask hierarchy.
type SubtaskTreeNode struct {
	Task     TaskItem
	Children []SubtaskTreeNode
	Depth    int
}

// BuildSubtaskTree builds a tree from flat task list.
func BuildSubtaskTree(tasks []TaskItem) []SubtaskTreeNode {
	// Create lookup map
	byID := make(map[string]TaskItem)
	for _, t := range tasks {
		byID[t.ID] = t
	}

	// Find root tasks (no parent or parent not in list)
	var roots []SubtaskTreeNode
	childrenOf := make(map[string][]TaskItem)

	for _, t := range tasks {
		if t.ParentID == "" {
			roots = append(roots, SubtaskTreeNode{Task: t, Depth: 0})
		} else {
			childrenOf[t.ParentID] = append(childrenOf[t.ParentID], t)
		}
	}

	// Build tree recursively
	var buildChildren func(parentID string, depth int) []SubtaskTreeNode
	buildChildren = func(parentID string, depth int) []SubtaskTreeNode {
		children := childrenOf[parentID]
		var nodes []SubtaskTreeNode
		for _, child := range children {
			node := SubtaskTreeNode{
				Task:     child,
				Depth:    depth,
				Children: buildChildren(child.ID, depth+1),
			}
			nodes = append(nodes, node)
		}
		return nodes
	}

	// Add children to roots
	for i := range roots {
		roots[i].Children = buildChildren(roots[i].Task.ID, 1)
	}

	return roots
}

// RenderSubtaskTree renders the subtask tree with indentation.
func RenderSubtaskTree(nodes []SubtaskTreeNode, indent int) string {
	var lines []string

	indentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	for _, node := range nodes {
		prefix := strings.Repeat("  ", indent+node.Depth)
		connector := indentStyle.Render("|-- ")

		taskLine := fmt.Sprintf("%s%s%s %s",
			prefix, connector,
			node.Task.DisplayID,
			truncate(node.Task.Title, 40))

		// Add status indicator
		statusColors := map[string]string{
			"done":        "82",  // green
			"in_progress": "214", // orange
			"review":      "205", // pink
			"todo":        "39",  // cyan
			"backlog":     "240", // gray
		}
		color := statusColors[node.Task.Column]
		if color == "" {
			color = "240"
		}
		status := lipgloss.NewStyle().
			Foreground(lipgloss.Color(color)).
			Render("[" + node.Task.Column + "]")

		lines = append(lines, taskLine+" "+status)

		// Render children recursively
		if len(node.Children) > 0 {
			childLines := RenderSubtaskTree(node.Children, indent)
			lines = append(lines, childLines)
		}
	}

	return strings.Join(lines, "\n")
}

// ExpandedSubtaskView tracks which parent tasks have expanded subtasks.
type ExpandedSubtaskView struct {
	expanded map[string]bool // parentID -> expanded
	subtasks map[string][]TaskItem // cached subtasks
}

// NewExpandedSubtaskView creates a new subtask view tracker.
func NewExpandedSubtaskView() *ExpandedSubtaskView {
	return &ExpandedSubtaskView{
		expanded: make(map[string]bool),
		subtasks: make(map[string][]TaskItem),
	}
}

// IsExpanded returns true if the task's subtasks are visible.
func (v *ExpandedSubtaskView) IsExpanded(taskID string) bool {
	return v.expanded[taskID]
}

// Toggle expands/collapses subtasks for a task.
func (v *ExpandedSubtaskView) Toggle(taskID string) {
	v.expanded[taskID] = !v.expanded[taskID]
}

// SetSubtasks caches loaded subtasks.
func (v *ExpandedSubtaskView) SetSubtasks(parentID string, subtasks []TaskItem) {
	v.subtasks[parentID] = subtasks
}

// GetSubtasks returns cached subtasks.
func (v *ExpandedSubtaskView) GetSubtasks(parentID string) []TaskItem {
	return v.subtasks[parentID]
}

// taskItemFromRecord helper - should be defined in task_item.go
// Adding signature here for reference
func taskItemFromRecord(r interface{}, epicCache *EpicCache) TaskItem {
	// Implementation would extract fields from PocketBase record
	// This is a placeholder - actual implementation depends on record type
	return TaskItem{}
}
```

**Expected output** (collapsed):

```
* WRK-42 [+3] Implement user auth [feature]
```

**Expected output** (expanded):

```
* WRK-42 [-3] Implement user auth [feature]
  |-- WRK-43 Design auth flow [done]
  |-- WRK-44 Add login endpoint [in_progress]
  |-- WRK-45 Add logout endpoint [todo]
```

**Common mistakes**:
- Infinite loops in tree building with circular parent references
- Not limiting depth to prevent excessive nesting
- Losing navigation state when collapsing

---

### 6.5 Multi-Select Tasks (Space Key)

**What**: Allow selecting multiple tasks with Space key, show selection count in status bar.

**Why**: Multi-select enables bulk operations like moving or deleting many tasks at once.

**Steps**:

1. Create `internal/tui/multi_select.go` with selection state
2. Add Space key handler to toggle selection
3. Update task rendering to show selection checkbox
4. Show selection count in status bar
5. Add "select all in column" (Ctrl+A)
6. Add "clear selection" (Escape when selecting)

**File**: `internal/tui/multi_select.go`

```go
package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// MultiSelect manages multi-selection state for tasks.
type MultiSelect struct {
	selected map[string]bool // taskID -> selected
	order    []string        // selection order (for ordered operations)
}

// NewMultiSelect creates a new multi-select state.
func NewMultiSelect() *MultiSelect {
	return &MultiSelect{
		selected: make(map[string]bool),
		order:    make([]string, 0),
	}
}

// Toggle adds or removes a task from selection.
func (m *MultiSelect) Toggle(taskID string) {
	if m.selected[taskID] {
		delete(m.selected, taskID)
		// Remove from order
		for i, id := range m.order {
			if id == taskID {
				m.order = append(m.order[:i], m.order[i+1:]...)
				break
			}
		}
	} else {
		m.selected[taskID] = true
		m.order = append(m.order, taskID)
	}
}

// Select adds a task to selection (without toggling).
func (m *MultiSelect) Select(taskID string) {
	if !m.selected[taskID] {
		m.selected[taskID] = true
		m.order = append(m.order, taskID)
	}
}

// Deselect removes a task from selection.
func (m *MultiSelect) Deselect(taskID string) {
	if m.selected[taskID] {
		delete(m.selected, taskID)
		for i, id := range m.order {
			if id == taskID {
				m.order = append(m.order[:i], m.order[i+1:]...)
				break
			}
		}
	}
}

// IsSelected returns true if task is selected.
func (m *MultiSelect) IsSelected(taskID string) bool {
	return m.selected[taskID]
}

// Count returns number of selected tasks.
func (m *MultiSelect) Count() int {
	return len(m.selected)
}

// HasSelection returns true if any tasks are selected.
func (m *MultiSelect) HasSelection() bool {
	return len(m.selected) > 0
}

// Clear removes all selections.
func (m *MultiSelect) Clear() {
	m.selected = make(map[string]bool)
	m.order = make([]string, 0)
}

// GetSelected returns all selected task IDs.
func (m *MultiSelect) GetSelected() []string {
	return m.order
}

// GetSelectedMap returns the selection map (for fast lookup).
func (m *MultiSelect) GetSelectedMap() map[string]bool {
	return m.selected
}

// SelectAll selects all tasks in the provided list.
func (m *MultiSelect) SelectAll(taskIDs []string) {
	for _, id := range taskIDs {
		m.Select(id)
	}
}

// SelectAllInColumn selects all tasks in a column.
func (m *MultiSelect) SelectAllInColumn(tasks []TaskItem) {
	for _, t := range tasks {
		m.Select(t.ID)
	}
}

// RenderSelectionIndicator renders the checkbox for a task.
func RenderSelectionIndicator(selected bool) string {
	if selected {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true).
			Render("[x]")
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("[ ]")
}

// RenderSelectionCount renders the selection count for status bar.
func RenderSelectionCount(count int) string {
	if count == 0 {
		return ""
	}

	style := lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("255")).
		Bold(true).
		Padding(0, 1)

	text := fmt.Sprintf("%d selected", count)
	if count == 1 {
		text = "1 selected"
	}

	return style.Render(text)
}

// SelectionMode represents the current selection mode.
type SelectionMode int

const (
	SelectionModeNone   SelectionMode = iota // Normal navigation
	SelectionModeActive                       // Multi-select active
	SelectionModeRange                        // Range selection (shift+click style)
)

// SelectionState tracks complete selection state.
type SelectionState struct {
	Mode        SelectionMode
	MultiSelect *MultiSelect
	RangeStart  string // For range selection
}

// NewSelectionState creates a new selection state.
func NewSelectionState() *SelectionState {
	return &SelectionState{
		Mode:        SelectionModeNone,
		MultiSelect: NewMultiSelect(),
	}
}

// EnterSelectionMode activates multi-select mode.
func (s *SelectionState) EnterSelectionMode() {
	s.Mode = SelectionModeActive
}

// ExitSelectionMode clears selection and returns to normal mode.
func (s *SelectionState) ExitSelectionMode() {
	s.Mode = SelectionModeNone
	s.MultiSelect.Clear()
}

// ToggleTask toggles selection, entering select mode if needed.
func (s *SelectionState) ToggleTask(taskID string) {
	if s.Mode == SelectionModeNone {
		s.Mode = SelectionModeActive
	}
	s.MultiSelect.Toggle(taskID)

	// Exit mode if nothing selected
	if s.MultiSelect.Count() == 0 {
		s.Mode = SelectionModeNone
	}
}
```

**Expected output** (status bar with selection):

```
+------------------+
| 3 selected       | Space: toggle | Esc: clear | m: move | d: delete
+------------------+
```

**Common mistakes**:
- Not clearing selection after bulk operation
- Losing selection when navigating between columns
- Not showing selection mode in status bar

---

### 6.6 Bulk Operations (Move, Delete)

**What**: Perform operations on all selected tasks at once.

**Why**: Moving or deleting tasks one-by-one is tedious. Bulk operations save significant time.

**Steps**:

1. Create `internal/tui/bulk_ops.go` with bulk operation commands
2. Add bulk move (m key when selected, then column number)
3. Add bulk delete (d key when selected, with confirmation)
4. Show progress feedback during operations
5. Handle partial failures gracefully

**File**: `internal/tui/bulk_ops.go`

```go
package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pocketbase/pocketbase"
)

// BulkOperationType represents the type of bulk operation.
type BulkOperationType int

const (
	BulkOpMove BulkOperationType = iota
	BulkOpDelete
	BulkOpSetPriority
	BulkOpSetType
	BulkOpAddLabel
)

// BulkOperation represents a bulk operation to perform.
type BulkOperation struct {
	Type     BulkOperationType
	TaskIDs  []string
	Target   string // target column, priority, etc.
}

// BulkResult represents the result of a bulk operation.
type BulkResult struct {
	Operation   BulkOperationType
	SuccessIDs  []string
	FailedIDs   []string
	Errors      []error
}

// BulkResultMsg is the tea.Msg for bulk operation completion.
type BulkResultMsg struct {
	Result BulkResult
}

// BulkProgressMsg reports progress during bulk operations.
type BulkProgressMsg struct {
	Current int
	Total   int
	TaskID  string
}

// BulkErrorMsg reports an error during bulk operations.
type BulkErrorMsg struct {
	Errors []error
}

// BulkMove moves multiple tasks to a target column.
func BulkMove(app *pocketbase.PocketBase, taskIDs []string, targetColumn string) tea.Cmd {
	return func() tea.Msg {
		result := BulkResult{
			Operation: BulkOpMove,
		}

		for _, id := range taskIDs {
			record, err := app.FindRecordById("tasks", id)
			if err != nil {
				result.FailedIDs = append(result.FailedIDs, id)
				result.Errors = append(result.Errors, fmt.Errorf("task %s: %w", id, err))
				continue
			}

			// Get next position in target column
			position := getNextPositionInColumn(app, targetColumn, record.GetString("board"))

			record.Set("column", targetColumn)
			record.Set("position", position)

			if err := app.Save(record); err != nil {
				result.FailedIDs = append(result.FailedIDs, id)
				result.Errors = append(result.Errors, fmt.Errorf("task %s: %w", id, err))
				continue
			}

			result.SuccessIDs = append(result.SuccessIDs, id)
		}

		return BulkResultMsg{Result: result}
	}
}

// BulkDelete deletes multiple tasks.
func BulkDelete(app *pocketbase.PocketBase, taskIDs []string) tea.Cmd {
	return func() tea.Msg {
		result := BulkResult{
			Operation: BulkOpDelete,
		}

		for _, id := range taskIDs {
			record, err := app.FindRecordById("tasks", id)
			if err != nil {
				result.FailedIDs = append(result.FailedIDs, id)
				result.Errors = append(result.Errors, fmt.Errorf("task %s: %w", id, err))
				continue
			}

			if err := app.Delete(record); err != nil {
				result.FailedIDs = append(result.FailedIDs, id)
				result.Errors = append(result.Errors, fmt.Errorf("task %s: %w", id, err))
				continue
			}

			result.SuccessIDs = append(result.SuccessIDs, id)
		}

		return BulkResultMsg{Result: result}
	}
}

// BulkSetPriority sets priority on multiple tasks.
func BulkSetPriority(app *pocketbase.PocketBase, taskIDs []string, priority string) tea.Cmd {
	return func() tea.Msg {
		result := BulkResult{
			Operation: BulkOpSetPriority,
		}

		for _, id := range taskIDs {
			record, err := app.FindRecordById("tasks", id)
			if err != nil {
				result.FailedIDs = append(result.FailedIDs, id)
				result.Errors = append(result.Errors, err)
				continue
			}

			record.Set("priority", priority)

			if err := app.Save(record); err != nil {
				result.FailedIDs = append(result.FailedIDs, id)
				result.Errors = append(result.Errors, err)
				continue
			}

			result.SuccessIDs = append(result.SuccessIDs, id)
		}

		return BulkResultMsg{Result: result}
	}
}

// ConfirmDialog represents a confirmation dialog.
type ConfirmDialog struct {
	Title   string
	Message string
	OnConfirm func() tea.Cmd
	OnCancel  func() tea.Cmd
	Width   int
}

// NewDeleteConfirmDialog creates a delete confirmation dialog.
func NewDeleteConfirmDialog(count int, onConfirm, onCancel func() tea.Cmd) ConfirmDialog {
	msg := fmt.Sprintf("Delete %d task(s)? This cannot be undone.", count)
	if count == 1 {
		msg = "Delete 1 task? This cannot be undone."
	}

	return ConfirmDialog{
		Title:     "Confirm Delete",
		Message:   msg,
		OnConfirm: onConfirm,
		OnCancel:  onCancel,
		Width:     50,
	}
}

// View renders the confirmation dialog.
func (d ConfirmDialog) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196"))

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	buttonStyle := lipgloss.NewStyle().
		Padding(0, 2)

	confirmBtn := buttonStyle.Copy().
		Background(lipgloss.Color("196")).
		Foreground(lipgloss.Color("255")).
		Render("Delete (Enter)")

	cancelBtn := buttonStyle.Copy().
		Background(lipgloss.Color("240")).
		Foreground(lipgloss.Color("255")).
		Render("Cancel (Esc)")

	content := lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Render(d.Title),
		"",
		messageStyle.Render(d.Message),
		"",
		lipgloss.JoinHorizontal(lipgloss.Center, cancelBtn, "  ", confirmBtn),
	)

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(1, 2).
		Width(d.Width)

	return dialogStyle.Render(content)
}

// RenderBulkResult formats the result of a bulk operation.
func RenderBulkResult(result BulkResult) string {
	var action string
	switch result.Operation {
	case BulkOpMove:
		action = "moved"
	case BulkOpDelete:
		action = "deleted"
	case BulkOpSetPriority:
		action = "updated"
	default:
		action = "processed"
	}

	if len(result.FailedIDs) == 0 {
		// All succeeded
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Render(fmt.Sprintf("%d task(s) %s successfully", len(result.SuccessIDs), action))
	}

	// Some failed
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Render(fmt.Sprintf("%d succeeded, %d failed",
			len(result.SuccessIDs), len(result.FailedIDs)))
}

// getNextPositionInColumn helper - would be imported from commands package
func getNextPositionInColumn(app *pocketbase.PocketBase, column, boardID string) float64 {
	// Implementation would query max position in column and add 1000
	// Simplified version:
	return 1000.0
}
```

**Expected output** (confirmation dialog):

```
+------------------------------------------+
|            Confirm Delete                |
|                                          |
|  Delete 3 task(s)? This cannot be undone.|
|                                          |
|     [Cancel (Esc)]    [Delete (Enter)]   |
+------------------------------------------+
```

**Expected output** (result message):

```
3 task(s) deleted successfully
```

**Common mistakes**:
- Not handling partial failures (some succeed, some fail)
- Not confirming destructive operations
- Not clearing selection after operation
- Not refreshing task list after bulk changes

---

### 6.7 Help Overlay (? Key)

**What**: Show all keyboard shortcuts in a toggleable overlay.

**Why**: Users need discoverability for keyboard shortcuts. The help overlay teaches the interface.

**Steps**:

1. Create `internal/tui/help.go` using bubbles/help
2. Define all keybindings with descriptions
3. Implement toggle with ? key
4. Style the overlay to not obscure too much

**File**: `internal/tui/help.go`

```go
package tui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// HelpKeyMap defines all keybindings for help display.
type HelpKeyMap struct {
	// Navigation
	Up         key.Binding
	Down       key.Binding
	Left       key.Binding
	Right      key.Binding
	FirstItem  key.Binding
	LastItem   key.Binding

	// Task Actions
	Enter      key.Binding
	NewTask    key.Binding
	EditTask   key.Binding
	DeleteTask key.Binding

	// Task Movement
	MoveLeft   key.Binding
	MoveRight  key.Binding
	MoveUp     key.Binding
	MoveDown   key.Binding
	MoveTo     key.Binding

	// Selection
	Select     key.Binding
	SelectAll  key.Binding
	ClearSelect key.Binding

	// Subtasks
	ExpandToggle key.Binding

	// Filtering
	Search     key.Binding
	FilterPriority key.Binding
	FilterType key.Binding
	FilterLabel key.Binding
	FilterEpic key.Binding
	ClearFilters key.Binding

	// Global
	Help       key.Binding
	Board      key.Binding
	Refresh    key.Binding
	Quit       key.Binding
	Escape     key.Binding
}

// DefaultHelpKeyMap returns the default keybindings.
func DefaultHelpKeyMap() HelpKeyMap {
	return HelpKeyMap{
		// Navigation
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/up", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/down", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("h/left", "prev column"),
		),
		Right: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("l/right", "next column"),
		),
		FirstItem: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "first task"),
		),
		LastItem: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "last task"),
		),

		// Task Actions
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "view details"),
		),
		NewTask: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new task"),
		),
		EditTask: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit task"),
		),
		DeleteTask: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete task"),
		),

		// Task Movement
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
		MoveTo: key.NewBinding(
			key.WithKeys("1", "2", "3", "4", "5"),
			key.WithHelp("1-5", "move to column"),
		),

		// Selection
		Select: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "select task"),
		),
		SelectAll: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "select all in column"),
		),
		ClearSelect: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "clear selection"),
		),

		// Subtasks
		ExpandToggle: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "expand/collapse subtasks"),
		),

		// Filtering
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		FilterPriority: key.NewBinding(
			key.WithKeys("f", "p"),
			key.WithHelp("fp", "filter by priority"),
		),
		FilterType: key.NewBinding(
			key.WithKeys("f", "t"),
			key.WithHelp("ft", "filter by type"),
		),
		FilterLabel: key.NewBinding(
			key.WithKeys("f", "l"),
			key.WithHelp("fl", "filter by label"),
		),
		FilterEpic: key.NewBinding(
			key.WithKeys("f", "e"),
			key.WithHelp("fe", "filter by epic"),
		),
		ClearFilters: key.NewBinding(
			key.WithKeys("f", "c"),
			key.WithHelp("fc", "clear filters"),
		),

		// Global
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Board: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "switch board"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "refresh"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel/close"),
		),
	}
}

// ShortHelp returns bindings shown in compact help view.
func (k HelpKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Help, k.Up, k.Down, k.Left, k.Right,
		k.Enter, k.NewTask, k.Quit,
	}
}

// FullHelp returns bindings shown in full help view.
func (k HelpKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		// Navigation column
		{k.Up, k.Down, k.Left, k.Right, k.FirstItem, k.LastItem},
		// Task Actions column
		{k.Enter, k.NewTask, k.EditTask, k.DeleteTask, k.ExpandToggle},
		// Movement column
		{k.MoveLeft, k.MoveRight, k.MoveUp, k.MoveDown, k.MoveTo},
		// Selection column
		{k.Select, k.SelectAll, k.ClearSelect},
		// Filtering column
		{k.Search, k.FilterPriority, k.FilterType, k.FilterEpic, k.ClearFilters},
		// Global column
		{k.Help, k.Board, k.Refresh, k.Quit},
	}
}

// HelpOverlay wraps bubbles/help with styling.
type HelpOverlay struct {
	model   help.Model
	keys    HelpKeyMap
	visible bool
	width   int
	height  int
}

// NewHelpOverlay creates a new help overlay.
func NewHelpOverlay() *HelpOverlay {
	h := help.New()
	h.ShowAll = true

	// Customize styles
	h.Styles.ShortKey = lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)
	h.Styles.ShortDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))
	h.Styles.ShortSeparator = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
	h.Styles.FullKey = lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)
	h.Styles.FullDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))
	h.Styles.FullSeparator = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	return &HelpOverlay{
		model:   h,
		keys:    DefaultHelpKeyMap(),
		visible: false,
	}
}

// Toggle shows/hides the help overlay.
func (h *HelpOverlay) Toggle() {
	h.visible = !h.visible
}

// Show makes the help overlay visible.
func (h *HelpOverlay) Show() {
	h.visible = true
}

// Hide hides the help overlay.
func (h *HelpOverlay) Hide() {
	h.visible = false
}

// IsVisible returns true if help is showing.
func (h *HelpOverlay) IsVisible() bool {
	return h.visible
}

// SetSize updates the help overlay dimensions.
func (h *HelpOverlay) SetSize(width, height int) {
	h.width = width
	h.height = height
	h.model.Width = width - 4
}

// View renders the help overlay.
func (h *HelpOverlay) View() string {
	if !h.visible {
		return ""
	}

	content := h.model.View(h.keys)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Margin(0, 0, 1, 0)

	overlayStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		Width(h.width - 4).
		MaxHeight(h.height - 4)

	return overlayStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("Keyboard Shortcuts"),
			content,
			"",
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Render("Press ? to close"),
		),
	)
}

// ShortHelpView returns a compact help string for status bar.
func (h *HelpOverlay) ShortHelpView() string {
	h.model.ShowAll = false
	defer func() { h.model.ShowAll = true }()
	return h.model.View(h.keys)
}
```

**Expected output** (help overlay):

```
+------------------------------------------------------------------+
| Keyboard Shortcuts                                               |
|                                                                  |
| Navigation         Task Actions       Movement                   |
| k/up   move up     enter view details H   move task left         |
| j/down move down   n     new task     L   move task right        |
| h/left prev column e     edit task    K   move task up           |
| l/right next column d    delete task  J   move task down         |
| g      first task  tab   expand sub   1-5 move to column         |
| G      last task                                                 |
|                                                                  |
| Selection          Filtering          Global                     |
| space  select task /     search       ?   toggle help            |
| ctrl+a select all  fp    by priority  b   switch board           |
| esc    clear       ft    by type      ctrl+r refresh             |
|                    fe    by epic      q   quit                   |
|                    fc    clear                                   |
|                                                                  |
| Press ? to close                                                 |
+------------------------------------------------------------------+
```

**Common mistakes**:
- Not sizing help overlay to fit terminal
- Help obscuring important content without dismiss option
- Incomplete keybinding documentation

---

### 6.8 Keyboard Shortcut Cheat Sheet

**What**: Compact quick-reference for keyboard shortcuts shown in status bar.

**Why**: Users shouldn't need to open full help for common actions. The cheat sheet provides context-sensitive hints.

**Steps**:

1. Add context-sensitive hints to status bar
2. Show different hints based on current mode
3. Update hints when entering selection mode

**File**: `internal/tui/status_bar.go` (additions)

```go
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// StatusBar renders the bottom status bar with context-sensitive hints.
type StatusBar struct {
	width        int
	mode         ViewMode
	selectionCount int
	filterActive bool
	boardName    string
	taskCount    int
}

// ViewMode represents the current view mode for context hints.
type ViewMode int

const (
	ViewModeBoard ViewMode = iota
	ViewModeDetail
	ViewModeForm
	ViewModeFilter
	ViewModeSelection
	ViewModeHelp
)

// NewStatusBar creates a new status bar.
func NewStatusBar() *StatusBar {
	return &StatusBar{
		mode: ViewModeBoard,
	}
}

// SetWidth updates the status bar width.
func (s *StatusBar) SetWidth(width int) {
	s.width = width
}

// SetMode updates the current view mode.
func (s *StatusBar) SetMode(mode ViewMode) {
	s.mode = mode
}

// SetSelectionCount updates the selection count.
func (s *StatusBar) SetSelectionCount(count int) {
	s.selectionCount = count
	if count > 0 {
		s.mode = ViewModeSelection
	} else if s.mode == ViewModeSelection {
		s.mode = ViewModeBoard
	}
}

// SetFilterActive sets whether filters are active.
func (s *StatusBar) SetFilterActive(active bool) {
	s.filterActive = active
}

// SetBoardInfo sets the current board name and task count.
func (s *StatusBar) SetBoardInfo(name string, taskCount int) {
	s.boardName = name
	s.taskCount = taskCount
}

// View renders the status bar.
func (s *StatusBar) View() string {
	style := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("252")).
		Width(s.width)

	leftStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	rightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)

	// Left side: context-sensitive hints
	var hints string
	switch s.mode {
	case ViewModeBoard:
		hints = keyStyle.Render("?") + " help  " +
			keyStyle.Render("n") + " new  " +
			keyStyle.Render("e") + " edit  " +
			keyStyle.Render("d") + " delete  " +
			keyStyle.Render("/") + " search  " +
			keyStyle.Render("b") + " boards"

	case ViewModeSelection:
		hints = RenderSelectionCount(s.selectionCount) + "  " +
			keyStyle.Render("m") + " move  " +
			keyStyle.Render("d") + " delete  " +
			keyStyle.Render("esc") + " clear"

	case ViewModeDetail:
		hints = keyStyle.Render("e") + " edit  " +
			keyStyle.Render("d") + " delete  " +
			keyStyle.Render("esc") + " close"

	case ViewModeForm:
		hints = keyStyle.Render("Tab") + " next field  " +
			keyStyle.Render("Enter") + " save  " +
			keyStyle.Render("Esc") + " cancel"

	case ViewModeFilter:
		hints = keyStyle.Render("Enter") + " apply  " +
			keyStyle.Render("Esc") + " cancel  " +
			keyStyle.Render("fc") + " clear all"

	case ViewModeHelp:
		hints = keyStyle.Render("?") + " close help"
	}

	left := leftStyle.Render(hints)

	// Right side: board info
	right := rightStyle.Render(s.boardName)
	if s.taskCount > 0 {
		right += rightStyle.Render(" * " + formatCount(s.taskCount, "task"))
	}
	if s.filterActive {
		right += "  " + lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Render("[filtered]")
	}

	// Calculate spacing
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	padding := s.width - leftWidth - rightWidth - 2
	if padding < 0 {
		padding = 0
	}

	content := left + spacer(padding) + right

	return style.Padding(0, 1).Render(content)
}

func formatCount(n int, singular string) string {
	if n == 1 {
		return "1 " + singular
	}
	return lipgloss.NewStyle().Render(
		string(rune('0'+n/1000)) + string(rune('0'+n%1000/100)) +
			string(rune('0'+n%100/10)) + string(rune('0'+n%10)) + " " + singular + "s")
}

func spacer(n int) string {
	if n <= 0 {
		return ""
	}
	s := make([]byte, n)
	for i := range s {
		s[i] = ' '
	}
	return string(s)
}
```

**Expected output** (normal mode):

```
? help  n new  e edit  d delete  / search  b boards        Work Board * 42 tasks
```

**Expected output** (selection mode):

```
3 selected  m move  d delete  esc clear                    Work Board * 42 tasks
```

---

### 6.9 Command Palette (Ctrl+K) - Optional

**What**: Quick command launcher with fuzzy search for all actions.

**Why**: Power users prefer typing commands over remembering shortcuts. The palette provides IDE-like productivity.

**Steps**:

1. Create `internal/tui/command_palette.go`
2. Define all available commands
3. Implement fuzzy matching on command names
4. Execute selected command

**File**: `internal/tui/command_palette.go`

```go
package tui

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Command represents an action in the command palette.
type Command struct {
	ID          string
	Name        string        // Display name
	Description string        // Longer description
	Shortcut    string        // Keyboard shortcut hint
	Category    string        // Grouping category
	Action      func() tea.Cmd // Action to execute
}

// CommandPalette is the quick command launcher.
type CommandPalette struct {
	visible     bool
	input       textinput.Model
	commands    []Command
	filtered    []Command
	selected    int
	width       int
	height      int
	maxResults  int
}

// NewCommandPalette creates a new command palette.
func NewCommandPalette(commands []Command) *CommandPalette {
	ti := textinput.New()
	ti.Placeholder = "Type a command..."
	ti.Focus()

	return &CommandPalette{
		visible:    false,
		input:      ti,
		commands:   commands,
		filtered:   commands,
		selected:   0,
		maxResults: 10,
	}
}

// DefaultCommands returns the standard command list.
func DefaultCommands(actions *CommandActions) []Command {
	return []Command{
		// Task Commands
		{ID: "new-task", Name: "New Task", Description: "Create a new task", Shortcut: "n", Category: "Tasks", Action: actions.NewTask},
		{ID: "edit-task", Name: "Edit Task", Description: "Edit selected task", Shortcut: "e", Category: "Tasks", Action: actions.EditTask},
		{ID: "delete-task", Name: "Delete Task", Description: "Delete selected task", Shortcut: "d", Category: "Tasks", Action: actions.DeleteTask},
		{ID: "view-task", Name: "View Task Details", Description: "Open task detail view", Shortcut: "Enter", Category: "Tasks", Action: actions.ViewTask},

		// Movement Commands
		{ID: "move-left", Name: "Move Task Left", Description: "Move task to previous column", Shortcut: "H", Category: "Movement", Action: actions.MoveTaskLeft},
		{ID: "move-right", Name: "Move Task Right", Description: "Move task to next column", Shortcut: "L", Category: "Movement", Action: actions.MoveTaskRight},
		{ID: "move-backlog", Name: "Move to Backlog", Description: "Move task to backlog", Shortcut: "1", Category: "Movement", Action: actions.MoveToColumn("backlog")},
		{ID: "move-todo", Name: "Move to Todo", Description: "Move task to todo", Shortcut: "2", Category: "Movement", Action: actions.MoveToColumn("todo")},
		{ID: "move-progress", Name: "Move to In Progress", Description: "Move task to in_progress", Shortcut: "3", Category: "Movement", Action: actions.MoveToColumn("in_progress")},
		{ID: "move-review", Name: "Move to Review", Description: "Move task to review", Shortcut: "4", Category: "Movement", Action: actions.MoveToColumn("review")},
		{ID: "move-done", Name: "Move to Done", Description: "Mark task as done", Shortcut: "5", Category: "Movement", Action: actions.MoveToColumn("done")},

		// Filter Commands
		{ID: "search", Name: "Search", Description: "Search tasks", Shortcut: "/", Category: "Filter", Action: actions.Search},
		{ID: "filter-priority", Name: "Filter by Priority", Description: "Filter tasks by priority", Shortcut: "fp", Category: "Filter", Action: actions.FilterByPriority},
		{ID: "filter-type", Name: "Filter by Type", Description: "Filter tasks by type", Shortcut: "ft", Category: "Filter", Action: actions.FilterByType},
		{ID: "filter-epic", Name: "Filter by Epic", Description: "Filter tasks by epic", Shortcut: "fe", Category: "Filter", Action: actions.FilterByEpic},
		{ID: "clear-filters", Name: "Clear Filters", Description: "Remove all active filters", Shortcut: "fc", Category: "Filter", Action: actions.ClearFilters},

		// View Commands
		{ID: "switch-board", Name: "Switch Board", Description: "Change to different board", Shortcut: "b", Category: "View", Action: actions.SwitchBoard},
		{ID: "refresh", Name: "Refresh", Description: "Reload all data", Shortcut: "Ctrl+R", Category: "View", Action: actions.Refresh},
		{ID: "toggle-help", Name: "Toggle Help", Description: "Show/hide keyboard shortcuts", Shortcut: "?", Category: "View", Action: actions.ToggleHelp},

		// Bulk Commands
		{ID: "select-all", Name: "Select All in Column", Description: "Select all tasks in current column", Shortcut: "Ctrl+A", Category: "Selection", Action: actions.SelectAll},
		{ID: "bulk-move", Name: "Bulk Move", Description: "Move selected tasks", Shortcut: "m", Category: "Selection", Action: actions.BulkMove},
		{ID: "bulk-delete", Name: "Bulk Delete", Description: "Delete selected tasks", Shortcut: "d", Category: "Selection", Action: actions.BulkDelete},
	}
}

// CommandActions holds action functions for commands.
// This would be populated by the main App model.
type CommandActions struct {
	NewTask          func() tea.Cmd
	EditTask         func() tea.Cmd
	DeleteTask       func() tea.Cmd
	ViewTask         func() tea.Cmd
	MoveTaskLeft     func() tea.Cmd
	MoveTaskRight    func() tea.Cmd
	MoveToColumn     func(column string) func() tea.Cmd
	Search           func() tea.Cmd
	FilterByPriority func() tea.Cmd
	FilterByType     func() tea.Cmd
	FilterByEpic     func() tea.Cmd
	ClearFilters     func() tea.Cmd
	SwitchBoard      func() tea.Cmd
	Refresh          func() tea.Cmd
	ToggleHelp       func() tea.Cmd
	SelectAll        func() tea.Cmd
	BulkMove         func() tea.Cmd
	BulkDelete       func() tea.Cmd
}

// Show opens the command palette.
func (p *CommandPalette) Show() {
	p.visible = true
	p.input.SetValue("")
	p.filtered = p.commands
	p.selected = 0
	p.input.Focus()
}

// Hide closes the command palette.
func (p *CommandPalette) Hide() {
	p.visible = false
	p.input.Blur()
}

// IsVisible returns true if palette is showing.
func (p *CommandPalette) IsVisible() bool {
	return p.visible
}

// SetSize updates the palette dimensions.
func (p *CommandPalette) SetSize(width, height int) {
	p.width = width
	p.height = height
	p.input.Width = width - 6
}

// Update handles input for the command palette.
func (p *CommandPalette) Update(msg tea.Msg) (*CommandPalette, tea.Cmd) {
	if !p.visible {
		return p, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			p.Hide()
			return p, nil

		case "enter":
			if len(p.filtered) > 0 && p.selected < len(p.filtered) {
				cmd := p.filtered[p.selected]
				p.Hide()
				if cmd.Action != nil {
					return p, cmd.Action()
				}
			}
			return p, nil

		case "up", "ctrl+p":
			if p.selected > 0 {
				p.selected--
			}
			return p, nil

		case "down", "ctrl+n":
			if p.selected < len(p.filtered)-1 {
				p.selected++
			}
			return p, nil
		}
	}

	// Handle text input
	var cmd tea.Cmd
	p.input, cmd = p.input.Update(msg)

	// Filter commands based on input
	p.filterCommands(p.input.Value())

	return p, cmd
}

// filterCommands filters commands based on search query.
func (p *CommandPalette) filterCommands(query string) {
	if query == "" {
		p.filtered = p.commands
		p.selected = 0
		return
	}

	query = strings.ToLower(query)
	var matched []struct {
		cmd   Command
		score int
	}

	for _, cmd := range p.commands {
		score := fuzzyScore(query, strings.ToLower(cmd.Name))
		if score > 0 {
			matched = append(matched, struct {
				cmd   Command
				score int
			}{cmd, score})
		}
	}

	// Sort by score descending
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].score > matched[j].score
	})

	p.filtered = make([]Command, len(matched))
	for i, m := range matched {
		p.filtered[i] = m.cmd
	}

	// Clamp selection
	if p.selected >= len(p.filtered) {
		p.selected = len(p.filtered) - 1
	}
	if p.selected < 0 {
		p.selected = 0
	}
}

// fuzzyScore returns a matching score (0 = no match).
func fuzzyScore(query, text string) int {
	if strings.Contains(text, query) {
		// Bonus for substring match
		return 100 + (100 - len(text))
	}

	// Simple fuzzy: all query chars must appear in order
	qi := 0
	score := 0
	for ti := 0; ti < len(text) && qi < len(query); ti++ {
		if text[ti] == query[qi] {
			qi++
			score += 10
			// Bonus for consecutive matches
			if ti > 0 && text[ti-1] == query[qi-2] {
				score += 5
			}
		}
	}

	if qi == len(query) {
		return score
	}
	return 0
}

// View renders the command palette.
func (p *CommandPalette) View() string {
	if !p.visible {
		return ""
	}

	// Input field
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1).
		Width(p.width - 4)

	input := inputStyle.Render(p.input.View())

	// Results list
	var results []string
	for i, cmd := range p.filtered {
		if i >= p.maxResults {
			remaining := len(p.filtered) - p.maxResults
			moreStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Italic(true)
			results = append(results, moreStyle.Render(
				strings.Repeat(" ", 2)+"...and "+string(rune('0'+remaining))+" more"))
			break
		}

		style := lipgloss.NewStyle().Width(p.width - 6)
		if i == p.selected {
			style = style.
				Background(lipgloss.Color("62")).
				Foreground(lipgloss.Color("255"))
		}

		nameStyle := lipgloss.NewStyle().Bold(true)
		shortcutStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39"))

		line := nameStyle.Render(cmd.Name)
		if cmd.Shortcut != "" {
			line += " " + shortcutStyle.Render("["+cmd.Shortcut+"]")
		}

		results = append(results, style.Render("  "+line))
	}

	resultList := strings.Join(results, "\n")

	// Container
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 1).
		Width(p.width - 2)

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("Command Palette")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		input,
		"",
		resultList,
	)

	return containerStyle.Render(content)
}
```

**Expected output** (command palette):

```
+------------------------------------------------------+
| Command Palette                                      |
|                                                      |
| +--------------------------------------------------+ |
| | move                                             | |
| +--------------------------------------------------+ |
|                                                      |
|   Move Task Left [H]                                 |
|   Move Task Right [L]                                |
|   Move to Backlog [1]                                |
|   Move to Todo [2]                                   |
|   Move to In Progress [3]                            |
|   Bulk Move [m]                                      |
|   ...and 2 more                                      |
+------------------------------------------------------+
```

**Common mistakes**:
- Not resetting input on close
- Not handling empty results state
- Fuzzy matching too strict or too loose
- Not showing keyboard shortcut hints

---

## Verification Checklist

Complete each section in order. Check off each item as you verify it.

### Epic Display

- [x] **Epic badge shows on tasks**
  - Create a task with an epic assigned
  - Verify colored badge appears in task card
  - Verify epic title is truncated if too long

- [x] **Epic filter works**
  - Press `fe` to open epic filter
  - Select an epic
  - Verify only tasks from that epic are shown

### Blocked Indicator

- [x] **Blocked tasks show indicator**
  - Create a task blocked by another task
  - Verify [BLOCKED] badge appears in red
  - Verify "blocked by X task(s)" in description

- [x] **Blocked filter works**
  - Filter for blocked tasks only
  - Verify only tasks with blockers appear

### Due Dates

- [ ] **Overdue tasks show red**
  - Create task with past due date
  - Verify "!! OVERDUE (X days)" in red

- [ ] **Today tasks show orange**
  - Create task due today
  - Verify "! TODAY" in orange/bold

- [ ] **Soon tasks show yellow**
  - Create task due in 1-2 days
  - Verify "> Tomorrow" or similar in yellow

### Subtasks

- [ ] **Parent shows subtask count**
  - Create task with children
  - Verify [+N] indicator on parent

- [ ] **Subtasks expand/collapse**
  - Press Tab on parent task
  - Verify children appear indented
  - Press Tab again to collapse

### Multi-Select

- [ ] **Space toggles selection**
  - Press Space on a task
  - Verify [x] checkbox appears
  - Verify selection count in status bar

- [ ] **Ctrl+A selects all in column**
  - Navigate to column with tasks
  - Press Ctrl+A
  - Verify all tasks in column selected

- [ ] **Esc clears selection**
  - With tasks selected, press Esc
  - Verify all selections cleared

### Bulk Operations

- [ ] **Bulk move works**
  - Select multiple tasks
  - Press m, then column number
  - Verify all tasks moved

- [ ] **Bulk delete with confirmation**
  - Select multiple tasks
  - Press d
  - Verify confirmation dialog appears
  - Confirm and verify deletion

### Help Overlay

- [ ] **? toggles help**
  - Press ?
  - Verify help overlay appears
  - Press ? again to close

- [ ] **All keybindings documented**
  - Verify all implemented shortcuts appear in help

### Status Bar

- [ ] **Context-sensitive hints**
  - In normal mode, verify standard hints
  - Select tasks, verify selection hints
  - Open detail view, verify detail hints

### Command Palette (if implemented)

- [ ] **Ctrl+K opens palette**
  - Press Ctrl+K
  - Verify palette appears with input focused

- [ ] **Fuzzy search works**
  - Type partial command name
  - Verify matching commands filtered

- [ ] **Execute command**
  - Select a command, press Enter
  - Verify action executes

---

## File Summary

| File | Lines | Purpose |
|------|-------|---------|
| `internal/tui/epic.go` | ~130 | Epic data, caching, badge rendering |
| `internal/tui/task_item.go` | ~150 | Extended task rendering with all indicators |
| `internal/tui/blocked.go` | ~120 | Blocked indicator and blocker resolution |
| `internal/tui/due_date.go` | ~180 | Due date parsing, urgency, styling |
| `internal/tui/subtask.go` | ~200 | Subtask tree, expand/collapse logic |
| `internal/tui/multi_select.go` | ~150 | Multi-selection state management |
| `internal/tui/bulk_ops.go` | ~220 | Bulk move, delete, confirmation dialog |
| `internal/tui/help.go` | ~250 | Help overlay with full keybinding docs |
| `internal/tui/status_bar.go` | ~130 | Context-sensitive status bar hints |
| `internal/tui/command_palette.go` | ~300 | Optional command palette with fuzzy search |

**Total new code**: ~1,830 lines

---

## What You Should Have Now

After completing Phase 6, your TUI should:

```
+------------------------------------------------------------------+
|  Work Board                                    [filtered by epic] |
+------------------------------------------------------------------+
|  Backlog (3)  |  Todo (5)   | In Progress(2) | Review (1) | Done |
+---------------+-------------+----------------+------------+------+
| * WRK-10      | [x] WRK-15  | * WRK-20       | * WRK-25   |      |
|   Add login   |   Fix bug   |   [Auth] API   |   Tests    |      |
|   [Auth]      |   [BLOCKED] |   !! OVERDUE   |            |      |
|               |   ! TODAY   |                |            |      |
| * WRK-11 [+3] |             |                |            |      |
|   User flow   | * WRK-16    |                |            |      |
|   [feature]   |   Update    |                |            |      |
+---------------+-------------+----------------+------------+------+
| 1 selected  m move  d delete  esc clear      Work Board * 11 tasks|
+------------------------------------------------------------------+
```

Features completed:
- Epic badges with color coding
- [BLOCKED] indicator with blocker info
- Due date with urgency colors (red/yellow/orange/gray)
- Subtask indicator [+N] with expand/collapse
- Multi-select with [x] checkboxes
- Selection count in status bar
- Bulk move and delete operations
- Full help overlay (? key)
- Context-sensitive keyboard hints
- Optional: Command palette (Ctrl+K)

---

## Next Phase

**Phase 7: Polish & Testing** will add:
- Graceful terminal resize handling
- Loading indicators and spinners
- Comprehensive error handling and display
- Edge case handling (empty boards, no tasks)
- Performance optimization for large boards
- Unit tests for all components
- Integration tests with PocketBase
- Documentation and usage guide

---

## Troubleshooting

### Epic colors not rendering correctly

**Problem**: Epic badges appear without color or wrong color.

**Solution**:
- Verify epic has valid hex color format (`#RRGGBB`)
- Check terminal supports 256 colors
- Ensure Lip Gloss color conversion is correct

```bash
# Check terminal color support
echo $TERM
# Should be xterm-256color or similar
```

### Blocked indicator not appearing

**Problem**: Tasks with blockers don't show [BLOCKED].

**Solution**:
- Verify `blocked_by` field is being loaded
- Check JSON array parsing works correctly
- Ensure blocked_by contains valid task IDs

```go
// Debug: print blocked_by value
log.Printf("Task %s blocked_by: %v", task.ID, task.BlockedBy)
```

### Due dates showing wrong urgency

**Problem**: Tasks show wrong color for due date urgency.

**Solution**:
- Check timezone handling in date parsing
- Verify date format matches ISO 8601
- Test with explicit dates to verify logic

```go
// Debug: print parsed dates
log.Printf("Due: %v, Urgency: %v", dueDate, urgency)
```

### Subtasks not loading

**Problem**: Parent tasks don't show subtask count.

**Solution**:
- Verify `parent` field relation is set up
- Check subtask query filter is correct
- Ensure subtask counts are loaded with tasks

```go
// Debug: check parent field
log.Printf("Task %s parent: %s", task.ID, task.ParentID)
```

### Multi-select not persisting across columns

**Problem**: Selection clears when navigating to different column.

**Solution**:
- Ensure selection state is at App level, not Column level
- Verify selection map uses task IDs, not indices
- Check state isn't being reset on column change

### Bulk operations failing silently

**Problem**: Bulk move/delete doesn't show errors.

**Solution**:
- Check error handling in bulk operation commands
- Ensure BulkResultMsg is being processed
- Verify result display logic handles failures

```go
// Debug: log each operation
log.Printf("Bulk op on %s: %v", id, err)
```

### Help overlay too large/small

**Problem**: Help doesn't fit terminal or is too cramped.

**Solution**:
- Use `tea.WindowSizeMsg` to track terminal size
- Set max height on help overlay
- Consider scrollable viewport for long help

### Command palette fuzzy match too strict

**Problem**: Typing doesn't find expected commands.

**Solution**:
- Adjust fuzzy score thresholds
- Add fallback to substring match
- Include command description in search

---

## Glossary

| Term | Definition |
|------|------------|
| **Epic** | A collection of related tasks, displayed with color badge |
| **Blocker** | A task that must be completed before another can start |
| **Subtask** | A child task nested under a parent task |
| **Multi-select** | Selecting multiple items for batch operations |
| **Command Palette** | Quick launcher for commands with fuzzy search |
| **Urgency** | Due date proximity category (overdue, today, soon, etc.) |
| **Fuzzy matching** | Finding matches even with typos or partial input |
| **Help overlay** | Modal showing all available keyboard shortcuts |
| **Status bar** | Bottom bar showing context-sensitive hints |
| **Bulk operation** | Action performed on multiple selected items at once |
