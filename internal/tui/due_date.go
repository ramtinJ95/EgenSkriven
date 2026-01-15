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

// ParseDueDate parses a date string in various formats.
// Returns zero time if empty or invalid.
func ParseDueDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}

	// Try full datetime first (RFC3339)
	t, err := time.Parse(time.RFC3339, dateStr)
	if err == nil {
		return t
	}

	// Try date only (YYYY-MM-DD)
	t, err = time.Parse("2006-01-02", dateStr)
	if err == nil {
		return t
	}

	// Try datetime without timezone
	t, err = time.Parse("2006-01-02 15:04:05", dateStr)
	if err == nil {
		return t
	}

	// Try datetime with Z suffix
	t, err = time.Parse("2006-01-02T15:04:05Z", dateStr)
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

	switch urgency {
	case DueDateOverdue:
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		due := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, dueDate.Location())
		days := int(today.Sub(due).Hours() / 24)
		if days == 1 {
			return "OVERDUE (1 day)"
		}
		return fmt.Sprintf("OVERDUE (%d days)", days)
	case DueDateToday:
		return "TODAY"
	case DueDateSoon:
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		due := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, dueDate.Location())
		days := int(due.Sub(today).Hours() / 24)
		if days == 1 {
			return "Tomorrow"
		}
		return fmt.Sprintf("In %d days", days)
	case DueDateUpcoming:
		return dueDate.Format("Jan 2")
	default:
		return dueDate.Format("Jan 2")
	}
}

// RenderDueDate renders the due date with appropriate styling.
func RenderDueDate(dateStr string) string {
	dueDate := ParseDueDate(dateStr)
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
