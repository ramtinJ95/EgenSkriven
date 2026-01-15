package tui

import (
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
	// Simple heuristic based on first byte of hex
	fgColor := lipgloss.Color("#FFFFFF")
	if len(bgColorStr) == 7 && bgColorStr[0] == '#' {
		// Very rough brightness check based on red channel
		r := bgColorStr[1:3]
		if r >= "AA" || r >= "aa" {
			fgColor = lipgloss.Color("#000000")
		}
	}

	return lipgloss.NewStyle().
		Background(bgColor).
		Foreground(fgColor).
		Padding(0, 1)
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
