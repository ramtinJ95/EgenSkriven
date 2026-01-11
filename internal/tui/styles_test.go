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
