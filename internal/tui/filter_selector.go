package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FilterSelectorType identifies which filter is being selected
type FilterSelectorType int

const (
	FilterSelectorNone FilterSelectorType = iota
	FilterSelectorPriority
	FilterSelectorTypeFilter
	FilterSelectorLabel
	FilterSelectorEpic
	FilterSelectorBlocked
)

// FilterOption represents a selectable filter value
type FilterOption struct {
	Value   string
	Display string
	Color   string
}

// Implement list.Item interface
func (f FilterOption) FilterValue() string { return f.Display }
func (f FilterOption) Title() string       { return f.Display }
func (f FilterOption) Description() string { return "" }

// FilterSelector is a modal for selecting filter values
type FilterSelector struct {
	selectorType FilterSelectorType
	list         list.Model
	width        int
	height       int
	active       bool
}

// NewFilterSelector creates a filter selector
func NewFilterSelector() FilterSelector {
	return FilterSelector{
		selectorType: FilterSelectorNone,
		active:       false,
	}
}

// ShowPriority opens selector for priority filter
func (s *FilterSelector) ShowPriority() tea.Cmd {
	options := []list.Item{
		FilterOption{Value: "urgent", Display: "Urgent", Color: "196"},
		FilterOption{Value: "high", Display: "High", Color: "208"},
		FilterOption{Value: "medium", Display: "Medium", Color: "226"},
		FilterOption{Value: "low", Display: "Low", Color: "240"},
	}
	return s.show(FilterSelectorPriority, "Filter by Priority", options)
}

// ShowType opens selector for type filter
func (s *FilterSelector) ShowType() tea.Cmd {
	options := []list.Item{
		FilterOption{Value: "bug", Display: "Bug", Color: "196"},
		FilterOption{Value: "feature", Display: "Feature", Color: "39"},
		FilterOption{Value: "chore", Display: "Chore", Color: "240"},
	}
	return s.show(FilterSelectorTypeFilter, "Filter by Type", options)
}

// ShowLabel opens selector for label filter with available labels
func (s *FilterSelector) ShowLabel(labels []string) tea.Cmd {
	options := make([]list.Item, len(labels))
	for i, label := range labels {
		options[i] = FilterOption{Value: label, Display: label, Color: "205"}
	}
	return s.show(FilterSelectorLabel, "Filter by Label", options)
}

// ShowEpic opens selector for epic filter with available epics
func (s *FilterSelector) ShowEpic(epics []EpicOption) tea.Cmd {
	options := make([]list.Item, len(epics))
	for i, epic := range epics {
		options[i] = FilterOption{
			Value:   epic.ID,
			Display: epic.Title,
			Color:   "99",
		}
	}
	return s.show(FilterSelectorEpic, "Filter by Epic", options)
}

// ShowBlocked opens selector for blocked status filter
func (s *FilterSelector) ShowBlocked() tea.Cmd {
	options := []list.Item{
		FilterOption{Value: "yes", Display: "Blocked Tasks", Color: "196"},
		FilterOption{Value: "no", Display: "Unblocked Tasks", Color: "82"},
	}
	return s.show(FilterSelectorBlocked, "Filter by Blocked Status", options)
}

func (s *FilterSelector) show(sType FilterSelectorType, title string, options []list.Item) tea.Cmd {
	s.selectorType = sType
	s.active = true

	// Create custom delegate with colored items
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetHeight(1)
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("62")).
		Bold(true)

	l := list.New(options, delegate, s.width-4, s.height-6)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowPagination(len(options) > 10)

	s.list = l

	return nil
}

// Hide closes the selector
func (s *FilterSelector) Hide() {
	s.active = false
	s.selectorType = FilterSelectorNone
}

// IsActive returns true if selector is visible
func (s FilterSelector) IsActive() bool {
	return s.active
}

// GetType returns the current selector type
func (s FilterSelector) GetType() FilterSelectorType {
	return s.selectorType
}

// SetSize updates dimensions
func (s *FilterSelector) SetSize(width, height int) {
	s.width = width
	s.height = height
	if s.active {
		s.list.SetSize(width-4, height-6)
	}
}

// Update handles input
func (s FilterSelector) Update(msg tea.Msg) (FilterSelector, tea.Cmd) {
	if !s.active {
		return s, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Get selected option
			if selected, ok := s.list.SelectedItem().(FilterOption); ok {
				currentType := s.selectorType
				s.Hide()
				return s, func() tea.Msg {
					return FilterSelectedMsg{
						Type:   currentType,
						Filter: createFilterFromOption(currentType, selected),
					}
				}
			}
			return s, nil

		case "esc", "q":
			s.Hide()
			return s, func() tea.Msg {
				return FilterCancelledMsg{}
			}

		default:
			var cmd tea.Cmd
			s.list, cmd = s.list.Update(msg)
			return s, cmd
		}
	}

	// Forward other messages to list
	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return s, cmd
}

func createFilterFromOption(sType FilterSelectorType, option FilterOption) Filter {
	var field string
	switch sType {
	case FilterSelectorPriority:
		field = "priority"
	case FilterSelectorTypeFilter:
		field = "type"
	case FilterSelectorLabel:
		field = "label"
	case FilterSelectorEpic:
		field = "epic"
	case FilterSelectorBlocked:
		field = "blocked"
	}

	return Filter{
		Field:    field,
		Operator: "is",
		Value:    option.Value,
		Display:  fmt.Sprintf("%s: %s", capitalizeFirst(field), option.Display),
	}
}

// View renders the selector
func (s FilterSelector) View() string {
	if !s.active {
		return ""
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1).
		Width(min(s.width-8, 40)).
		Height(min(s.height-4, 15))

	content := s.list.View()

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("j/k to move | Enter to select | Esc to cancel")

	box := boxStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		"",
		help,
	))

	return lipgloss.Place(
		s.width,
		s.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}

// FilterSelectedMsg is sent when a filter is chosen
type FilterSelectedMsg struct {
	Type   FilterSelectorType
	Filter Filter
}

// FilterCancelledMsg is sent when selector is cancelled
type FilterCancelledMsg struct{}

// EpicOption for epic filter selection
type EpicOption struct {
	ID    string
	Title string
	Color string
}

// capitalizeFirst capitalizes the first letter of a string
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	if s[0] >= 'a' && s[0] <= 'z' {
		return string(s[0]-32) + s[1:]
	}
	return s
}
