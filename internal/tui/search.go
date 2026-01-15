package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SearchOverlay provides a modal search input
type SearchOverlay struct {
	input       textinput.Model
	filterState *FilterState
	width       int
	height      int
	active      bool
}

// NewSearchOverlay creates a search overlay
func NewSearchOverlay(filterState *FilterState) SearchOverlay {
	ti := textinput.New()
	ti.Placeholder = "Type to search..."
	ti.CharLimit = 100
	ti.Width = 50
	ti.Prompt = "/ "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))

	return SearchOverlay{
		input:       ti,
		filterState: filterState,
		active:      false,
	}
}

// Show activates the search overlay
func (s *SearchOverlay) Show() tea.Cmd {
	s.active = true
	s.input.SetValue(s.filterState.GetSearchQuery())
	s.input.Focus()
	s.input.CursorEnd()
	return textinput.Blink
}

// Hide deactivates the search overlay
func (s *SearchOverlay) Hide() {
	s.active = false
	s.input.Blur()
}

// IsActive returns true if overlay is visible
func (s SearchOverlay) IsActive() bool {
	return s.active
}

// SetSize updates the overlay dimensions
func (s *SearchOverlay) SetSize(width, height int) {
	s.width = width
	s.height = height
	s.input.Width = min(width-10, 60)
}

// Init implements tea.Model
func (s SearchOverlay) Init() tea.Cmd {
	return nil
}

// Update handles input events
func (s SearchOverlay) Update(msg tea.Msg) (SearchOverlay, tea.Cmd) {
	if !s.active {
		return s, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Apply search
			s.filterState.SetSearchQuery(s.input.Value())
			s.Hide()
			return s, func() tea.Msg {
				return SearchAppliedMsg{Query: s.input.Value()}
			}

		case "esc":
			// Cancel without applying
			s.Hide()
			return s, func() tea.Msg {
				return SearchCancelledMsg{}
			}

		case "ctrl+u":
			// Clear input
			s.input.SetValue("")
			return s, nil

		default:
			var cmd tea.Cmd
			s.input, cmd = s.input.Update(msg)
			return s, cmd
		}
	}

	return s, nil
}

// View renders the overlay
func (s SearchOverlay) View() string {
	if !s.active {
		return ""
	}

	// Overlay box style
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Width(min(s.width-4, 70))

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1).
		Render("Search Tasks")

	// Input
	input := s.input.View()

	// Help text
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		MarginTop(1).
		Render("Enter to search | Esc to cancel | Ctrl+U to clear")

	// Combine
	content := lipgloss.JoinVertical(lipgloss.Left, title, input, help)
	box := boxStyle.Render(content)

	// Center the box
	return lipgloss.Place(
		s.width,
		s.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
	)
}

// SearchAppliedMsg is sent when search is confirmed
type SearchAppliedMsg struct {
	Query string
}

// SearchCancelledMsg is sent when search is cancelled
type SearchCancelledMsg struct{}
