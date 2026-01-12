package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Header displays board information at the top of the TUI
type Header struct {
	boardName   string
	boardPrefix string
	boardColor  string
	taskCount   int
	width       int
	showHelp    bool // Show keybinding hints
}

// NewHeader creates a new header component
func NewHeader() *Header {
	return &Header{
		showHelp: true,
	}
}

// SetBoard updates the header with board information
func (h *Header) SetBoard(name, prefix, color string) {
	h.boardName = name
	h.boardPrefix = prefix
	h.boardColor = color
}

// SetTaskCount updates the task count display
func (h *Header) SetTaskCount(count int) {
	h.taskCount = count
}

// SetWidth updates the header width
func (h *Header) SetWidth(width int) {
	h.width = width
}

// SetShowHelp toggles help hint visibility
func (h *Header) SetShowHelp(show bool) {
	h.showHelp = show
}

// View renders the header
func (h *Header) View() string {
	if h.boardName == "" {
		return ""
	}

	// Base styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		MarginBottom(1)

	// Board prefix with color
	prefixColor := lipgloss.Color("39") // Default cyan
	if h.boardColor != "" {
		prefixColor = lipgloss.Color(h.boardColor)
	}
	prefix := lipgloss.NewStyle().
		Foreground(prefixColor).
		Bold(true).
		Render("[" + h.boardPrefix + "]")

	// Board name
	name := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Bold(true).
		Render(h.boardName)

	// Task count
	countStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	var countText string
	if h.taskCount == 0 {
		countText = "No tasks"
	} else if h.taskCount == 1 {
		countText = "1 task"
	} else {
		countText = fmt.Sprintf("%d tasks", h.taskCount)
	}
	count := countStyle.Render(countText)

	// Left side: board info
	leftContent := fmt.Sprintf("%s %s  %s", prefix, name, count)

	// Right side: keybinding hints (if enabled)
	var rightContent string
	if h.showHelp {
		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Faint(true)
		rightContent = helpStyle.Render("b:switch board  ?:help  q:quit")
	}

	// Calculate spacing
	leftLen := lipgloss.Width(leftContent)
	rightLen := lipgloss.Width(rightContent)
	spacerLen := h.width - leftLen - rightLen - 4 // Account for padding
	if spacerLen < 1 {
		spacerLen = 1
	}

	// Build the header line
	spacer := lipgloss.NewStyle().Width(spacerLen).Render("")
	content := leftContent + spacer + rightContent

	// Add bottom border
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color("240")).
		Width(h.width)

	return headerStyle.Render(borderStyle.Render(content))
}

// Height returns the height of the header in lines
func (h *Header) Height() int {
	return 2 // Header line + border
}
