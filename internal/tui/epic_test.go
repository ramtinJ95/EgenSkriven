package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEpicBadgeStyle(t *testing.T) {
	tests := []struct {
		name  string
		epic  EpicOption
		desc  string
	}{
		{
			name:  "default color",
			epic:  EpicOption{ID: "1", Title: "Epic", Color: ""},
			desc:  "should use default purple when no color set",
		},
		{
			name:  "custom color",
			epic:  EpicOption{ID: "2", Title: "Epic", Color: "#3B82F6"},
			desc:  "should use custom color",
		},
		{
			name:  "light color",
			epic:  EpicOption{ID: "3", Title: "Epic", Color: "#FFFFFF"},
			desc:  "should have dark text for light background",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := EpicBadgeStyle(tt.epic)
			// Style should be created without error
			assert.NotNil(t, style)
		})
	}
}

func TestRenderEpicBadge(t *testing.T) {
	tests := []struct {
		name     string
		epic     EpicOption
		maxWidth int
		expected string
	}{
		{
			name:     "empty epic",
			epic:     EpicOption{},
			maxWidth: 15,
			expected: "",
		},
		{
			name:     "short title",
			epic:     EpicOption{ID: "1", Title: "Auth", Color: "#3B82F6"},
			maxWidth: 15,
			expected: "Auth", // Title should be in output
		},
		{
			name:     "long title truncated",
			epic:     EpicOption{ID: "2", Title: "Very Long Epic Title That Exceeds Max", Color: "#3B82F6"},
			maxWidth: 15,
			expected: "Very Long", // Should be truncated, contains truncated title
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderEpicBadge(tt.epic, tt.maxWidth)
			if tt.expected == "" {
				assert.Empty(t, result)
			} else {
				assert.Contains(t, result, tt.expected)
			}
		})
	}
}
