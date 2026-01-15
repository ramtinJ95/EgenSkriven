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
	ID              string
	TaskTitle       string // renamed to avoid conflict with Title() method
	TaskDescription string // renamed to avoid conflict with Description() method
	Type            string // bug, feature, chore
	Priority        string // low, medium, high, urgent
	Column          string // backlog, todo, in_progress, need_input, review, done
	Labels          []string
	Position        float64
	DueDate         string // due date in YYYY-MM-DD format
	EpicID          string // ID of parent epic
	EpicTitle       string // title of parent epic (for display)

	// Display fields
	DisplayID string // e.g., "WRK-123"

	// Computed fields
	IsBlocked bool
	BlockedBy []string

	// Resolved epic information (for badge display)
	Epic EpicOption
}

// FilterValue returns the string used for filtering in the list.
// When the user types to filter, this value is searched.
func (t TaskItem) FilterValue() string {
	// Include title and display ID for filtering
	return t.TaskTitle + " " + t.DisplayID
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
// Format: [PRIORITY] DISPLAY_ID Title [TYPE] [BLOCKED] [DUE DATE]
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
	parts = append(parts, t.TaskTitle)

	// Type badge
	if t.Type != "" {
		parts = append(parts, GetTypeIndicator(t.Type))
	}

	// Blocked indicator
	if t.IsBlocked {
		blocked := blockedIndicatorStyle.Render("[BLOCKED]")
		parts = append(parts, blocked)
	}

	// Due date with urgency highlighting
	if t.DueDate != "" {
		dueStr := RenderDueDate(t.DueDate)
		if dueStr != "" {
			parts = append(parts, dueStr)
		}
	}

	return strings.Join(parts, " ")
}

// renderDescription creates the secondary info line.
// Shows epic badge, labels, blocked info, and other metadata.
func (t TaskItem) renderDescription() string {
	var parts []string

	// Epic badge (color-coded)
	if t.Epic.ID != "" && t.Epic.Title != "" {
		epicBadge := RenderEpicBadge(t.Epic, 15)
		parts = append(parts, epicBadge)
	}

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

	// Blocked by info (show count of blocking tasks)
	if t.IsBlocked && len(t.BlockedBy) > 0 {
		blockedStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Faint(true)
		blockerText := fmt.Sprintf("blocked by %d task(s)", len(t.BlockedBy))
		if len(t.BlockedBy) == 1 {
			blockerText = "blocked by 1 task"
		}
		parts = append(parts, blockedStyle.Render(blockerText))
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
		ID:              record.Id,
		TaskTitle:       record.GetString("title"),
		TaskDescription: record.GetString("description"),
		Type:            record.GetString("type"),
		Priority:        record.GetString("priority"),
		Column:          record.GetString("column"),
		Labels:          labels,
		Position:        record.GetFloat("position"),
		DueDate:         record.GetString("due_date"),
		EpicID:          record.GetString("epic"),
		DisplayID:       displayID,
		IsBlocked:       isBlocked,
		BlockedBy:       blockedBy,
	}
}

// NewTaskItemFromMap creates a TaskItem from a map (used for realtime events).
// The map comes from PocketBase SSE events which provide record data as JSON.
func NewTaskItemFromMap(m map[string]interface{}, boardPrefix string) TaskItem {
	getString := func(key string) string {
		if v, ok := m[key].(string); ok {
			return v
		}
		return ""
	}

	getFloat := func(key string) float64 {
		switch v := m[key].(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int64:
			return float64(v)
		default:
			return 0
		}
	}

	getInt := func(key string) int {
		switch v := m[key].(type) {
		case float64:
			return int(v)
		case int:
			return v
		case int64:
			return int(v)
		default:
			return 0
		}
	}

	getStringSlice := func(key string) []string {
		switch v := m[key].(type) {
		case []interface{}:
			result := make([]string, 0, len(v))
			for _, item := range v {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result
		case []string:
			return v
		default:
			return nil
		}
	}

	// Build display ID from prefix and seq
	seq := getInt("seq")
	displayID := fmt.Sprintf("%s-%d", boardPrefix, seq)

	// Extract blocked_by to determine if task is blocked
	blockedBy := getStringSlice("blocked_by")
	isBlocked := len(blockedBy) > 0

	return TaskItem{
		ID:              getString("id"),
		TaskTitle:       getString("title"),
		TaskDescription: getString("description"),
		Type:            getString("type"),
		Priority:        getString("priority"),
		Column:          getString("column"),
		Labels:          getStringSlice("labels"),
		Position:        getFloat("position"),
		DueDate:         getString("due_date"),
		EpicID:          getString("epic"),
		DisplayID:       displayID,
		IsBlocked:       isBlocked,
		BlockedBy:       blockedBy,
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

// SetEpic sets the resolved epic information for badge display.
func (t *TaskItem) SetEpic(epic EpicOption) {
	t.Epic = epic
	t.EpicTitle = epic.Title
}

// ResolveEpics sets epic information on tasks from available epics.
// This is used after loading tasks to add color-coded epic badges.
func ResolveEpics(tasks []TaskItem, epics []EpicOption) []TaskItem {
	// Build lookup map
	epicMap := make(map[string]EpicOption, len(epics))
	for _, epic := range epics {
		epicMap[epic.ID] = epic
	}

	// Resolve epics on tasks
	for i := range tasks {
		if tasks[i].EpicID != "" {
			if epic, ok := epicMap[tasks[i].EpicID]; ok {
				tasks[i].Epic = epic
				tasks[i].EpicTitle = epic.Title
			}
		}
	}

	return tasks
}
