package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTaskItem_FilterValue(t *testing.T) {
	item := TaskItem{
		TaskTitle: "Implement feature",
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
				TaskTitle: "Simple task",
				DisplayID: "WRK-1",
				Priority:  "medium",
				Type:      "feature",
			},
			contains: []string{"WRK-1", "Simple task", "[feature]"},
		},
		{
			name: "urgent bug",
			item: TaskItem{
				TaskTitle: "Fix crash",
				DisplayID: "WRK-2",
				Priority:  "urgent",
				Type:      "bug",
			},
			contains: []string{"WRK-2", "Fix crash", "[bug]", "!!!"},
		},
		{
			name: "blocked task",
			item: TaskItem{
				TaskTitle: "Blocked task",
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
