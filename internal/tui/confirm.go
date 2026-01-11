package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmDialog presents a yes/no confirmation to the user
type ConfirmDialog struct {
	title    string
	message  string
	yesLabel string
	noLabel  string
	focused  bool // true = Yes focused, false = No focused
	width    int
	keys     confirmKeyMap
}

type confirmKeyMap struct {
	Yes   key.Binding
	No    key.Binding
	Left  key.Binding
	Right key.Binding
	Tab   key.Binding
}

func defaultConfirmKeyMap() confirmKeyMap {
	return confirmKeyMap{
		Yes: key.NewBinding(
			key.WithKeys("y", "enter"),
			key.WithHelp("y/enter", "confirm"),
		),
		No: key.NewBinding(
			key.WithKeys("n", "esc"),
			key.WithHelp("n/esc", "cancel"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("h/left", "switch"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("l/right", "switch"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch"),
		),
	}
}

// NewConfirmDialog creates a new confirmation dialog
func NewConfirmDialog(title, message string) *ConfirmDialog {
	return &ConfirmDialog{
		title:    title,
		message:  message,
		yesLabel: "Yes",
		noLabel:  "No",
		focused:  false, // Default to No for safety
		width:    50,
		keys:     defaultConfirmKeyMap(),
	}
}

// NewDeleteConfirmDialog creates a confirmation dialog for deleting a task
func NewDeleteConfirmDialog(taskTitle string) *ConfirmDialog {
	return &ConfirmDialog{
		title:    "Delete Task?",
		message:  "Delete \"" + truncateString(taskTitle, 30) + "\"?\nThis action cannot be undone.",
		yesLabel: "Delete",
		noLabel:  "Cancel",
		focused:  false, // Default to Cancel for safety
		width:    50,
		keys:     defaultConfirmKeyMap(),
	}
}

// Init initializes the dialog
func (d *ConfirmDialog) Init() tea.Cmd {
	return nil
}

// Update handles messages for the dialog
func (d *ConfirmDialog) Update(msg tea.Msg) (*ConfirmDialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, d.keys.Yes):
			if d.focused {
				return d, func() tea.Msg {
					return closeConfirmDialogMsg{confirmed: true}
				}
			}
			// If on No button and pressed enter, treat as No
			if msg.String() == "enter" && !d.focused {
				return d, func() tea.Msg {
					return closeConfirmDialogMsg{confirmed: false}
				}
			}
			// Just 'y' always confirms
			if msg.String() == "y" {
				return d, func() tea.Msg {
					return closeConfirmDialogMsg{confirmed: true}
				}
			}

		case key.Matches(msg, d.keys.No):
			return d, func() tea.Msg {
				return closeConfirmDialogMsg{confirmed: false}
			}

		case key.Matches(msg, d.keys.Left), key.Matches(msg, d.keys.Right), key.Matches(msg, d.keys.Tab):
			d.focused = !d.focused
			return d, nil
		}
	}

	return d, nil
}

// View renders the dialog
func (d *ConfirmDialog) View() string {
	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(warningColor).
		MarginBottom(1).
		Render(d.title)

	// Message
	message := lipgloss.NewStyle().
		Width(d.width - 4).
		Render(d.message)

	// Buttons
	yesStyle := buttonStyle
	noStyle := buttonStyle

	if d.focused {
		yesStyle = buttonDangerStyle
	} else {
		noStyle = buttonFocusedStyle
	}

	yesBtn := yesStyle.Render("[ " + d.yesLabel + " ]")
	noBtn := noStyle.Render("[ " + d.noLabel + " ]")

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, yesBtn, "  ", noBtn)

	// Combine all parts
	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		message,
		"",
		buttons,
	)

	return confirmDialogStyle.
		Width(d.width).
		Render(content)
}

// SetWidth sets the dialog width
func (d *ConfirmDialog) SetWidth(width int) {
	d.width = width
}

// truncateString truncates a string to a maximum length, adding "..." if truncated
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return "..."
	}
	return s[:maxLen-3] + "..."
}
