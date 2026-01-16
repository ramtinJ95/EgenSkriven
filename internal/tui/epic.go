package tui

import (
	"strconv"

	"github.com/charmbracelet/lipgloss"
)

// EpicBadgeStyle creates a styled badge for an epic option.
// Uses the epic's color as background with contrasting text.
func EpicBadgeStyle(epic EpicOption) lipgloss.Style {
	// Default color if not set
	bgColorStr := epic.Color
	if bgColorStr == "" {
		bgColorStr = "#6366F1" // Default purple
	}

	// Parse hex color for background
	bgColor := lipgloss.Color(bgColorStr)

	// Use white text for darker colors, black for lighter
	// Based on relative luminance calculation
	fgColor := lipgloss.Color("#FFFFFF")
	if isLightColor(bgColorStr) {
		fgColor = lipgloss.Color("#000000")
	}

	return lipgloss.NewStyle().
		Background(bgColor).
		Foreground(fgColor).
		Padding(0, 1)
}

// isLightColor determines if a hex color is light enough to need dark text.
// Uses the relative luminance formula: 0.299*R + 0.587*G + 0.114*B
// Returns true if the color is light (luminance > 128).
func isLightColor(hexColor string) bool {
	if len(hexColor) != 7 || hexColor[0] != '#' {
		return false // Invalid format, assume dark
	}

	r, err := strconv.ParseInt(hexColor[1:3], 16, 64)
	if err != nil {
		return false
	}
	g, err := strconv.ParseInt(hexColor[3:5], 16, 64)
	if err != nil {
		return false
	}
	b, err := strconv.ParseInt(hexColor[5:7], 16, 64)
	if err != nil {
		return false
	}

	// Calculate relative luminance using standard coefficients
	// These weights account for human perception of brightness
	luminance := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
	return luminance > 128
}

// RenderEpicBadge renders a compact epic badge for task cards.
func RenderEpicBadge(epic EpicOption, maxWidth int) string {
	if epic.Title == "" {
		return ""
	}

	title := epic.Title
	if maxWidth > 5 && len(title) > maxWidth-2 {
		title = title[:maxWidth-5] + "..."
	}

	return EpicBadgeStyle(epic).Render(title)
}
