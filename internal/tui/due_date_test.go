package tui

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDueDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool // true if should parse successfully
	}{
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "date only",
			input:    "2025-01-15",
			expected: true,
		},
		{
			name:     "RFC3339",
			input:    "2025-01-15T10:30:00Z",
			expected: true,
		},
		{
			name:     "datetime without timezone",
			input:    "2025-01-15 10:30:00",
			expected: true,
		},
		{
			name:     "invalid format",
			input:    "Jan 15, 2025",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseDueDate(tt.input)
			if tt.expected {
				assert.False(t, result.IsZero(), "expected valid date")
			} else {
				assert.True(t, result.IsZero(), "expected zero time")
			}
		})
	}
}

func TestGetDueDateUrgency(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	tests := []struct {
		name     string
		dueDate  time.Time
		expected DueDateUrgency
	}{
		{
			name:     "zero time",
			dueDate:  time.Time{},
			expected: DueDateNone,
		},
		{
			name:     "overdue by 3 days",
			dueDate:  today.AddDate(0, 0, -3),
			expected: DueDateOverdue,
		},
		{
			name:     "overdue by 1 day",
			dueDate:  today.AddDate(0, 0, -1),
			expected: DueDateOverdue,
		},
		{
			name:     "due today",
			dueDate:  today,
			expected: DueDateToday,
		},
		{
			name:     "due tomorrow",
			dueDate:  today.AddDate(0, 0, 1),
			expected: DueDateSoon,
		},
		{
			name:     "due in 2 days",
			dueDate:  today.AddDate(0, 0, 2),
			expected: DueDateSoon,
		},
		{
			name:     "due in 5 days",
			dueDate:  today.AddDate(0, 0, 5),
			expected: DueDateUpcoming,
		},
		{
			name:     "due in 7 days",
			dueDate:  today.AddDate(0, 0, 7),
			expected: DueDateUpcoming,
		},
		{
			name:     "due in 10 days",
			dueDate:  today.AddDate(0, 0, 10),
			expected: DueDateFuture,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDueDateUrgency(tt.dueDate)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatDueDate(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	tests := []struct {
		name     string
		dueDate  time.Time
		urgency  DueDateUrgency
		contains string
	}{
		{
			name:     "overdue 3 days",
			dueDate:  today.AddDate(0, 0, -3),
			urgency:  DueDateOverdue,
			contains: "OVERDUE (3 days)",
		},
		{
			name:     "overdue 1 day",
			dueDate:  today.AddDate(0, 0, -1),
			urgency:  DueDateOverdue,
			contains: "OVERDUE (1 day)",
		},
		{
			name:     "due today",
			dueDate:  today,
			urgency:  DueDateToday,
			contains: "TODAY",
		},
		{
			name:     "due tomorrow",
			dueDate:  today.AddDate(0, 0, 1),
			urgency:  DueDateSoon,
			contains: "Tomorrow",
		},
		{
			name:     "due in 2 days",
			dueDate:  today.AddDate(0, 0, 2),
			urgency:  DueDateSoon,
			contains: "In 2 days",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDueDate(tt.dueDate, tt.urgency)
			assert.Contains(t, result, tt.contains)
		})
	}
}

func TestRenderDueDate(t *testing.T) {
	tests := []struct {
		name     string
		dateStr  string
		contains string
		empty    bool
	}{
		{
			name:    "empty string",
			dateStr: "",
			empty:   true,
		},
		{
			name:     "today",
			dateStr:  time.Now().Format("2006-01-02"),
			contains: "TODAY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderDueDate(tt.dateStr)
			if tt.empty {
				assert.Empty(t, result)
			} else {
				assert.Contains(t, result, tt.contains)
			}
		})
	}
}

func TestDueDateStyles(t *testing.T) {
	// Verify all urgency levels have styles
	urgencies := []DueDateUrgency{
		DueDateNone,
		DueDateFuture,
		DueDateUpcoming,
		DueDateSoon,
		DueDateToday,
		DueDateOverdue,
	}

	for _, urgency := range urgencies {
		style, ok := DueDateStyles[urgency]
		assert.True(t, ok, "style should exist for urgency %d", urgency)
		assert.NotNil(t, style)
	}
}
