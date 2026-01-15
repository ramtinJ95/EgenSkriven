package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HelpOverlay displays keyboard shortcuts in a toggleable overlay.
type HelpOverlay struct {
	visible bool
	width   int
	height  int
}

// NewHelpOverlay creates a new help overlay.
func NewHelpOverlay() *HelpOverlay {
	return &HelpOverlay{
		visible: false,
	}
}

// Toggle shows/hides the help overlay.
func (h *HelpOverlay) Toggle() {
	h.visible = !h.visible
}

// Show makes the help overlay visible.
func (h *HelpOverlay) Show() {
	h.visible = true
}

// Hide hides the help overlay.
func (h *HelpOverlay) Hide() {
	h.visible = false
}

// IsVisible returns true if help is showing.
func (h *HelpOverlay) IsVisible() bool {
	return h.visible
}

// SetSize updates the help overlay dimensions.
func (h *HelpOverlay) SetSize(width, height int) {
	h.width = width
	h.height = height
}

// HelpSection represents a group of related keybindings.
type HelpSection struct {
	Title    string
	Bindings []HelpBinding
}

// HelpBinding represents a single keybinding.
type HelpBinding struct {
	Key         string
	Description string
}

// GetHelpSections returns all keyboard shortcut sections.
func GetHelpSections() []HelpSection {
	return []HelpSection{
		{
			Title: "Navigation",
			Bindings: []HelpBinding{
				{Key: "j/↓", Description: "Move down"},
				{Key: "k/↑", Description: "Move up"},
				{Key: "h/←", Description: "Previous column"},
				{Key: "l/→", Description: "Next column"},
				{Key: "g", Description: "First task"},
				{Key: "G", Description: "Last task"},
			},
		},
		{
			Title: "Task Actions",
			Bindings: []HelpBinding{
				{Key: "Enter", Description: "View details"},
				{Key: "n", Description: "New task"},
				{Key: "e", Description: "Edit task"},
				{Key: "d", Description: "Delete task"},
			},
		},
		{
			Title: "Task Movement",
			Bindings: []HelpBinding{
				{Key: "H", Description: "Move task left"},
				{Key: "L", Description: "Move task right"},
				{Key: "K", Description: "Move task up"},
				{Key: "J", Description: "Move task down"},
				{Key: "1-5", Description: "Move to column"},
			},
		},
		{
			Title: "Filtering",
			Bindings: []HelpBinding{
				{Key: "/", Description: "Search tasks"},
				{Key: "fp", Description: "Filter by priority"},
				{Key: "ft", Description: "Filter by type"},
				{Key: "fl", Description: "Filter by label"},
				{Key: "fe", Description: "Filter by epic"},
				{Key: "fb", Description: "Filter by blocked"},
				{Key: "fc", Description: "Clear filters"},
			},
		},
		{
			Title: "Global",
			Bindings: []HelpBinding{
				{Key: "?", Description: "Toggle help"},
				{Key: "b", Description: "Switch board"},
				{Key: "r", Description: "Refresh"},
				{Key: "q", Description: "Quit"},
				{Key: "Esc", Description: "Cancel/close"},
			},
		},
	}
}

// View renders the help overlay.
func (h *HelpOverlay) View() string {
	if !h.visible {
		return ""
	}

	sections := GetHelpSections()

	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	sectionTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(0)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("82")).
		Bold(true).
		Width(8)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	// Render sections
	var columns []string
	for _, section := range sections {
		var lines []string
		lines = append(lines, sectionTitleStyle.Render(section.Title))

		for _, binding := range section.Bindings {
			line := keyStyle.Render(binding.Key) + " " + descStyle.Render(binding.Description)
			lines = append(lines, line)
		}

		columns = append(columns, strings.Join(lines, "\n"))
	}

	// Arrange in rows (2-3 columns per row)
	colWidth := 28
	var rows []string

	// First row: Navigation, Task Actions, Task Movement
	if len(columns) >= 3 {
		row1 := lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.NewStyle().Width(colWidth).Render(columns[0]),
			lipgloss.NewStyle().Width(colWidth).Render(columns[1]),
			lipgloss.NewStyle().Width(colWidth).Render(columns[2]),
		)
		rows = append(rows, row1)
	}

	// Second row: Filtering, Global
	if len(columns) >= 5 {
		row2 := lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.NewStyle().Width(colWidth).Render(columns[3]),
			lipgloss.NewStyle().Width(colWidth).Render(columns[4]),
		)
		rows = append(rows, row2)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)

	// Footer
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		MarginTop(1)
	footer := footerStyle.Render("Press ? to close")

	// Main container
	overlayStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2)

	overlay := overlayStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("Keyboard Shortcuts"),
			content,
			footer,
		),
	)

	// Center the overlay
	return lipgloss.Place(
		h.width,
		h.height,
		lipgloss.Center,
		lipgloss.Center,
		overlay,
	)
}
