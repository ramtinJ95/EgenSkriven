package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FilterBar displays active filters and search input
type FilterBar struct {
	filterState  *FilterState
	searchInput  textinput.Model
	isSearching  bool
	selectedChip int // Index of selected chip for removal (-1 = none)
	width        int
	focused      bool
}

// NewFilterBar creates a new filter bar
func NewFilterBar(filterState *FilterState) FilterBar {
	ti := textinput.New()
	ti.Placeholder = "Search tasks..."
	ti.CharLimit = 100
	ti.Width = 30

	return FilterBar{
		filterState:  filterState,
		searchInput:  ti,
		isSearching:  false,
		selectedChip: -1,
		focused:      false,
	}
}

// Init implements tea.Model
func (f FilterBar) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (f FilterBar) Update(msg tea.Msg) (FilterBar, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if f.isSearching {
			return f.handleSearchInput(msg)
		}
		return f.handleChipNavigation(msg)
	}

	return f, cmd
}

func (f FilterBar) handleSearchInput(msg tea.KeyMsg) (FilterBar, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Apply search and exit search mode
		f.filterState.SetSearchQuery(f.searchInput.Value())
		f.isSearching = false
		f.searchInput.Blur()
		return f, nil

	case "esc":
		// Cancel search
		f.isSearching = false
		f.searchInput.SetValue(f.filterState.GetSearchQuery())
		f.searchInput.Blur()
		return f, nil

	default:
		// Forward to text input
		var cmd tea.Cmd
		f.searchInput, cmd = f.searchInput.Update(msg)
		return f, cmd
	}
}

func (f FilterBar) handleChipNavigation(msg tea.KeyMsg) (FilterBar, tea.Cmd) {
	filters := f.filterState.GetFilters()
	chipCount := len(filters)
	if f.filterState.GetSearchQuery() != "" {
		chipCount++ // Search query counts as a chip
	}

	switch msg.String() {
	case "left", "h":
		if chipCount > 0 {
			f.selectedChip--
			if f.selectedChip < 0 {
				f.selectedChip = chipCount - 1
			}
		}
		return f, nil

	case "right", "l":
		if chipCount > 0 {
			f.selectedChip++
			if f.selectedChip >= chipCount {
				f.selectedChip = 0
			}
		}
		return f, nil

	case "x", "backspace", "delete":
		// Remove selected chip
		if f.selectedChip >= 0 && f.selectedChip < chipCount {
			return f.removeSelectedChip()
		}
		return f, nil
	}

	return f, nil
}

func (f FilterBar) removeSelectedChip() (FilterBar, tea.Cmd) {
	filters := f.filterState.GetFilters()
	searchActive := f.filterState.GetSearchQuery() != ""

	// Determine what's at the selected index
	if searchActive {
		if f.selectedChip == 0 {
			// Remove search query
			f.filterState.SetSearchQuery("")
			f.searchInput.SetValue("")
		} else {
			// Remove filter (offset by 1 for search)
			filterIdx := f.selectedChip - 1
			if filterIdx < len(filters) {
				f.filterState.RemoveFilter(filters[filterIdx])
			}
		}
	} else {
		// No search, just filters
		if f.selectedChip < len(filters) {
			f.filterState.RemoveFilter(filters[f.selectedChip])
		}
	}

	// Adjust selected chip
	newCount := len(f.filterState.GetFilters())
	if f.filterState.GetSearchQuery() != "" {
		newCount++
	}
	if f.selectedChip >= newCount {
		f.selectedChip = newCount - 1
	}

	return f, nil
}

// StartSearch enters search mode
func (f *FilterBar) StartSearch() tea.Cmd {
	f.isSearching = true
	f.searchInput.SetValue(f.filterState.GetSearchQuery())
	f.searchInput.Focus()
	return textinput.Blink
}

// StopSearch exits search mode without applying
func (f *FilterBar) StopSearch() {
	f.isSearching = false
	f.searchInput.Blur()
}

// IsSearching returns true if search input is active
func (f FilterBar) IsSearching() bool {
	return f.isSearching
}

// SetWidth sets the available width for the filter bar
func (f *FilterBar) SetWidth(width int) {
	f.width = width
	f.searchInput.Width = min(width-20, 40)
}

// Focus sets focus state
func (f *FilterBar) Focus() {
	f.focused = true
}

// Blur removes focus
func (f *FilterBar) Blur() {
	f.focused = false
	f.selectedChip = -1
}

// View renders the filter bar
func (f FilterBar) View() string {
	if !f.filterState.HasActiveFilters() && !f.isSearching {
		return ""
	}

	var parts []string

	// Search input or search chip
	if f.isSearching {
		searchStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)
		parts = append(parts, searchStyle.Render("/ "+f.searchInput.View()))
	} else if f.filterState.GetSearchQuery() != "" {
		parts = append(parts, f.renderSearchChip())
	}

	// Filter chips
	filters := f.filterState.GetFilters()
	for i, filter := range filters {
		chipIdx := i
		if f.filterState.GetSearchQuery() != "" {
			chipIdx++ // Offset for search chip
		}
		parts = append(parts, f.renderFilterChip(filter, chipIdx))
	}

	// Clear all hint
	if len(parts) > 0 && !f.isSearching {
		hint := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render(" [fc to clear]")
		parts = append(parts, hint)
	}

	// Container style
	containerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Width(f.width).
		Padding(0, 1)

	return containerStyle.Render(strings.Join(parts, " "))
}

func (f FilterBar) renderSearchChip() string {
	isSelected := f.selectedChip == 0

	style := f.chipStyle(isSelected)
	icon := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Render("Q")

	content := icon + " " + Truncate(f.filterState.GetSearchQuery(), 20)

	if isSelected {
		content += " x"
	}

	return style.Render(content)
}

func (f FilterBar) renderFilterChip(filter Filter, idx int) string {
	isSelected := f.selectedChip == idx

	// Choose color based on field
	var color string
	switch filter.Field {
	case "priority":
		color = f.priorityColor(filter.Value)
	case "type":
		color = f.typeColor(filter.Value)
	case "label":
		color = "205" // Pink for labels
	case "epic":
		color = "99" // Purple for epics
	default:
		color = "62" // Default blue
	}

	style := f.chipStyle(isSelected)
	style = style.Background(lipgloss.Color(color))

	content := filter.String()
	if isSelected {
		content += " x"
	}

	return style.Render(content)
}

func (f FilterBar) chipStyle(selected bool) lipgloss.Style {
	base := lipgloss.NewStyle().
		Foreground(lipgloss.Color("0")).
		Padding(0, 1).
		MarginRight(1)

	if selected {
		return base.
			Bold(true).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("255"))
	}

	return base
}

func (f FilterBar) priorityColor(priority string) string {
	switch strings.ToLower(priority) {
	case "urgent":
		return "196" // Red
	case "high":
		return "208" // Orange
	case "medium":
		return "226" // Yellow
	case "low":
		return "240" // Gray
	default:
		return "62"
	}
}

func (f FilterBar) typeColor(taskType string) string {
	switch strings.ToLower(taskType) {
	case "bug":
		return "196" // Red
	case "feature":
		return "39" // Cyan
	case "chore":
		return "240" // Gray
	default:
		return "62"
	}
}
