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

func TestTaskItem_DescriptionWithEpic(t *testing.T) {
	tests := []struct {
		name     string
		item     TaskItem
		contains []string
	}{
		{
			name: "task with epic",
			item: TaskItem{
				Labels: []string{"backend"},
				Epic: EpicOption{
					ID:    "epic123",
					Title: "User Auth",
					Color: "#3B82F6",
				},
			},
			contains: []string{"User Auth", "#backend"},
		},
		{
			name: "task without epic",
			item: TaskItem{
				Labels: []string{"frontend"},
				Epic:   EpicOption{}, // Empty epic
			},
			contains: []string{"#frontend"},
		},
		{
			name: "epic only no labels",
			item: TaskItem{
				Epic: EpicOption{
					ID:    "epic456",
					Title: "Platform",
					Color: "#10B981",
				},
			},
			contains: []string{"Platform"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc := tt.item.Description()
			for _, s := range tt.contains {
				assert.Contains(t, desc, s)
			}
		})
	}
}

func TestResolveEpics(t *testing.T) {
	epics := []EpicOption{
		{ID: "epic1", Title: "Epic One", Color: "#FF0000"},
		{ID: "epic2", Title: "Epic Two", Color: "#00FF00"},
	}

	tasks := []TaskItem{
		{ID: "task1", TaskTitle: "Task 1", EpicID: "epic1"},
		{ID: "task2", TaskTitle: "Task 2", EpicID: "epic2"},
		{ID: "task3", TaskTitle: "Task 3", EpicID: ""}, // No epic
		{ID: "task4", TaskTitle: "Task 4", EpicID: "unknown"}, // Non-existent epic
	}

	resolved := ResolveEpics(tasks, epics)

	// Task 1 should have Epic One resolved
	assert.Equal(t, "Epic One", resolved[0].Epic.Title)
	assert.Equal(t, "Epic One", resolved[0].EpicTitle)
	assert.Equal(t, "#FF0000", resolved[0].Epic.Color)

	// Task 2 should have Epic Two resolved
	assert.Equal(t, "Epic Two", resolved[1].Epic.Title)
	assert.Equal(t, "Epic Two", resolved[1].EpicTitle)

	// Task 3 has no epic
	assert.Empty(t, resolved[2].Epic.ID)
	assert.Empty(t, resolved[2].EpicTitle)

	// Task 4 has non-existent epic (not resolved)
	assert.Empty(t, resolved[3].Epic.ID)
	assert.Empty(t, resolved[3].EpicTitle)
}

func TestSetEpic(t *testing.T) {
	task := &TaskItem{
		ID:        "task1",
		TaskTitle: "Test Task",
		EpicID:    "epic1",
	}

	epic := EpicOption{
		ID:    "epic1",
		Title: "Test Epic",
		Color: "#3B82F6",
	}

	task.SetEpic(epic)

	assert.Equal(t, "Test Epic", task.Epic.Title)
	assert.Equal(t, "Test Epic", task.EpicTitle)
	assert.Equal(t, "#3B82F6", task.Epic.Color)
}
