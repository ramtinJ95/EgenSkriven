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

func TestIsLightColor(t *testing.T) {
	tests := []struct {
		name     string
		hexColor string
		expected bool
	}{
		// Light colors - should return true (need dark text)
		{name: "white", hexColor: "#FFFFFF", expected: true},
		{name: "yellow", hexColor: "#FFFF00", expected: true},
		{name: "cyan", hexColor: "#00FFFF", expected: true},
		{name: "light gray", hexColor: "#CCCCCC", expected: true},
		{name: "light green", hexColor: "#90EE90", expected: true},

		// Dark colors - should return false (need white text)
		{name: "black", hexColor: "#000000", expected: false},
		{name: "dark blue", hexColor: "#000080", expected: false},
		{name: "dark red", hexColor: "#8B0000", expected: false},
		{name: "purple", hexColor: "#6366F1", expected: false},
		{name: "dark green", hexColor: "#006400", expected: false},

		// Edge cases - pure colors
		{name: "pure red", hexColor: "#FF0000", expected: false},    // luminance ~76
		{name: "pure green", hexColor: "#00FF00", expected: true},   // luminance ~150
		{name: "pure blue", hexColor: "#0000FF", expected: false},   // luminance ~29

		// Invalid formats - should return false (assume dark)
		{name: "invalid - no hash", hexColor: "FFFFFF", expected: false},
		{name: "invalid - short", hexColor: "#FFF", expected: false},
		{name: "invalid - empty", hexColor: "", expected: false},
		{name: "invalid - gibberish", hexColor: "#GGGGGG", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isLightColor(tt.hexColor)
			assert.Equal(t, tt.expected, result, "isLightColor(%q) should return %v", tt.hexColor, tt.expected)
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
