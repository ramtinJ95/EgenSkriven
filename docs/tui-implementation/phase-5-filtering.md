# Phase 5: Filtering & Search

**Goal**: Filter and search tasks in the TUI kanban board

**Duration Estimate**: 2-3 days

**Prerequisites**: Phase 4 (Real-Time Sync) completed

**Deliverable**: Can filter/search tasks like web UI with keyboard-driven filter controls

---

## Overview

Phase 5 adds powerful filtering and search capabilities to the TUI. Users need to quickly find specific tasks among potentially hundreds across multiple columns. This phase implements:

- **Search**: Full-text search across task titles and descriptions
- **Filters**: Filter by priority, type, label, and epic
- **Filter Bar**: Visual indicator of active filters
- **Quick Keys**: Two-key sequences for rapid filter access

### Why Filtering Matters

In a real kanban workflow:
- Developers might want to see only their urgent bugs
- Product managers might filter to a specific epic
- Team leads might look for tasks needing review

Without filtering, users must visually scan all columns - slow and error-prone. With filtering, they press two keys and see exactly what they need.

### Filter Architecture

Filters work as a pipeline:

```
All Tasks → Search Filter → Priority Filter → Type Filter → Label Filter → Epic Filter → Displayed Tasks
```

Each filter is additive (AND logic). Active filters are shown as "chips" in the FilterBar, providing visual feedback and quick removal.

---

## Tasks

### 5.1 Define Filter Types and State ✅ COMPLETED

**What**: Create the core filter data structures and state management.

**Why**: A clean data model makes filter logic predictable and testable. Separating filter state from UI state allows independent testing.

**Steps**:

1. Create `internal/tui/filter.go`
2. Define `Filter` struct with field, operator, and value
3. Define `FilterState` struct to hold all active filters
4. Implement `Apply()` method to filter task slices
5. Implement `matches()` helper for individual task matching

**Code**:

**File**: `internal/tui/filter.go`

```go
package tui

import (
	"encoding/json"
	"strings"
)

// Filter represents a single filter condition
type Filter struct {
	Field    string // "priority", "type", "label", "epic"
	Operator string // "is", "is_not", "includes"
	Value    string // The filter value
	Display  string // Human-readable display (e.g., "Priority: High")
}

// String returns a display-friendly representation of the filter
func (f Filter) String() string {
	if f.Display != "" {
		return f.Display
	}
	return f.Field + ":" + f.Value
}

// FilterState holds all active filters and search query
type FilterState struct {
	filters     []Filter
	searchQuery string
}

// NewFilterState creates an empty filter state
func NewFilterState() *FilterState {
	return &FilterState{
		filters: make([]Filter, 0),
	}
}

// AddFilter adds a new filter, replacing any existing filter for the same field
// (except for labels, which can have multiple values)
func (f *FilterState) AddFilter(filter Filter) {
	// For most fields, replace existing filter
	if filter.Field != "label" {
		f.RemoveFilterByField(filter.Field)
	}
	f.filters = append(f.filters, filter)
}

// RemoveFilter removes a specific filter
func (f *FilterState) RemoveFilter(filter Filter) {
	newFilters := make([]Filter, 0, len(f.filters))
	for _, existing := range f.filters {
		if existing.Field != filter.Field || existing.Value != filter.Value {
			newFilters = append(newFilters, existing)
		}
	}
	f.filters = newFilters
}

// RemoveFilterByField removes all filters for a given field
func (f *FilterState) RemoveFilterByField(field string) {
	newFilters := make([]Filter, 0, len(f.filters))
	for _, existing := range f.filters {
		if existing.Field != field {
			newFilters = append(newFilters, existing)
		}
	}
	f.filters = newFilters
}

// Clear removes all filters and search query
func (f *FilterState) Clear() {
	f.filters = make([]Filter, 0)
	f.searchQuery = ""
}

// SetSearchQuery sets the search query
func (f *FilterState) SetSearchQuery(query string) {
	f.searchQuery = query
}

// GetSearchQuery returns the current search query
func (f *FilterState) GetSearchQuery() string {
	return f.searchQuery
}

// GetFilters returns a copy of the active filters
func (f *FilterState) GetFilters() []Filter {
	result := make([]Filter, len(f.filters))
	copy(result, f.filters)
	return result
}

// HasActiveFilters returns true if any filters or search is active
func (f *FilterState) HasActiveFilters() bool {
	return len(f.filters) > 0 || f.searchQuery != ""
}

// Apply filters a slice of tasks and returns matching tasks
func (f *FilterState) Apply(tasks []TaskItem) []TaskItem {
	if !f.HasActiveFilters() {
		return tasks
	}

	result := make([]TaskItem, 0, len(tasks))
	for _, task := range tasks {
		if f.matches(task) {
			result = append(result, task)
		}
	}
	return result
}

// matches checks if a single task passes all filters
func (f *FilterState) matches(task TaskItem) bool {
	// Check search query first (searches title and description)
	if f.searchQuery != "" {
		query := strings.ToLower(f.searchQuery)
		titleMatch := strings.Contains(strings.ToLower(task.Title), query)
		descMatch := strings.Contains(strings.ToLower(task.Description), query)
		idMatch := strings.Contains(strings.ToLower(task.DisplayID), query)
		
		if !titleMatch && !descMatch && !idMatch {
			return false
		}
	}

	// Check each filter
	for _, filter := range f.filters {
		if !f.matchesFilter(task, filter) {
			return false
		}
	}

	return true
}

// matchesFilter checks if a task matches a single filter
func (f *FilterState) matchesFilter(task TaskItem, filter Filter) bool {
	switch filter.Field {
	case "priority":
		return f.matchPriority(task, filter)
	case "type":
		return f.matchType(task, filter)
	case "label":
		return f.matchLabel(task, filter)
	case "epic":
		return f.matchEpic(task, filter)
	default:
		return true // Unknown filter field - pass through
	}
}

func (f *FilterState) matchPriority(task TaskItem, filter Filter) bool {
	switch filter.Operator {
	case "is":
		return strings.EqualFold(task.Priority, filter.Value)
	case "is_not":
		return !strings.EqualFold(task.Priority, filter.Value)
	default:
		return strings.EqualFold(task.Priority, filter.Value)
	}
}

func (f *FilterState) matchType(task TaskItem, filter Filter) bool {
	switch filter.Operator {
	case "is":
		return strings.EqualFold(task.Type, filter.Value)
	case "is_not":
		return !strings.EqualFold(task.Type, filter.Value)
	default:
		return strings.EqualFold(task.Type, filter.Value)
	}
}

func (f *FilterState) matchLabel(task TaskItem, filter Filter) bool {
	// Labels use "includes" logic - task must have the label
	for _, label := range task.Labels {
		if strings.EqualFold(label, filter.Value) {
			return filter.Operator != "is_not"
		}
	}
	// Label not found
	return filter.Operator == "is_not"
}

func (f *FilterState) matchEpic(task TaskItem, filter Filter) bool {
	switch filter.Operator {
	case "is":
		// Match by epic ID or title
		return task.Epic == filter.Value || 
		       strings.EqualFold(task.EpicTitle, filter.Value)
	case "is_not":
		return task.Epic != filter.Value && 
		       !strings.EqualFold(task.EpicTitle, filter.Value)
	default:
		return task.Epic == filter.Value || 
		       strings.EqualFold(task.EpicTitle, filter.Value)
	}
}

// ToJSON serializes filter state for persistence
func (f *FilterState) ToJSON() ([]byte, error) {
	data := struct {
		Filters     []Filter `json:"filters"`
		SearchQuery string   `json:"searchQuery"`
	}{
		Filters:     f.filters,
		SearchQuery: f.searchQuery,
	}
	return json.Marshal(data)
}

// FromJSON deserializes filter state
func (f *FilterState) FromJSON(data []byte) error {
	var parsed struct {
		Filters     []Filter `json:"filters"`
		SearchQuery string   `json:"searchQuery"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}
	f.filters = parsed.Filters
	f.searchQuery = parsed.SearchQuery
	return nil
}
```

**Expected output**: The file compiles without errors:
```bash
go build ./internal/tui/...
```

**Common Mistakes**:
- Forgetting to handle case-insensitive matching (users expect "HIGH" == "high")
- Not handling empty label slices (nil vs empty slice)
- Mutating the original slice instead of returning a copy

---

### 5.2 Create FilterBar Component ✅ COMPLETED

**What**: Build the visual component that displays active filters and search.

**Why**: Users need visual feedback about what filters are active. Filter "chips" are a familiar UI pattern that also allow easy removal.

**Steps**:

1. Create `internal/tui/filter_bar.go`
2. Implement the Bubble Tea Model interface
3. Create styled "chips" for each active filter
4. Integrate with textinput for search
5. Handle chip removal interaction

**Code**:

**File**: `internal/tui/filter_bar.go`

```go
package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FilterBar displays active filters and search input
type FilterBar struct {
	filterState   *FilterState
	searchInput   textinput.Model
	isSearching   bool
	selectedChip  int // Index of selected chip for removal (-1 = none)
	width         int
	focused       bool
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

	content := icon + " " + truncate(f.filterState.GetSearchQuery(), 20)
	
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

// Helper function to truncate strings
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "..."
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
```

**Expected output**: Component renders properly with styled chips:
```
[Q: dark mode x] [Priority: High] [Type: Bug] [fc to clear]
```

**Common Mistakes**:
- Not handling empty filter state (renders empty bar instead of nothing)
- Forgetting to handle keyboard navigation when searching
- Chip selection getting out of sync when filters are removed

---

### 5.3 Create Search Overlay ✅ COMPLETED

**What**: Implement the search input overlay that appears when pressing `/`.

**Why**: The search overlay provides a focused input experience, separate from the main board view. This pattern is familiar from vim, browsers, and many other tools.

**Steps**:

1. Create `internal/tui/search.go`
2. Build a centered overlay with text input
3. Handle real-time search preview (optional)
4. Integrate with FilterState

**Code**:

**File**: `internal/tui/search.go`

```go
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
```

**Expected output**: Centered overlay appears when triggered:
```
                    +--------------------+
                    | Search Tasks       |
                    | / dark mode_       |
                    |                    |
                    | Enter | Esc | ^U   |
                    +--------------------+
```

**Common Mistakes**:
- Not preserving existing search query when opening overlay
- Forgetting to blur the input when closing
- Not centering properly on different terminal sizes

---

### 5.4 Create Filter Selector Overlays ✅ COMPLETED

**What**: Build reusable selector overlays for priority, type, label, and epic filters.

**Why**: Each filter type needs a selection UI. A reusable component reduces code duplication and ensures consistent behavior.

**Steps**:

1. Create `internal/tui/filter_selector.go`
2. Build a list-based selector component
3. Implement factory functions for each filter type
4. Handle selection and cancellation

**Code**:

**File**: `internal/tui/filter_selector.go`

```go
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
	FilterSelectorType
	FilterSelectorLabel
	FilterSelectorEpic
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
	return s.show(FilterSelectorType, "Filter by Type", options)
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
			Display: fmt.Sprintf("%s", epic.Title),
			Color:   "99",
		}
	}
	return s.show(FilterSelectorEpic, "Filter by Epic", options)
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
				s.Hide()
				return s, func() tea.Msg {
					return FilterSelectedMsg{
						Type:   s.selectorType,
						Filter: s.createFilter(selected),
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

func (s FilterSelector) createFilter(option FilterOption) Filter {
	var field string
	switch s.selectorType {
	case FilterSelectorPriority:
		field = "priority"
	case FilterSelectorType:
		field = "type"
	case FilterSelectorLabel:
		field = "label"
	case FilterSelectorEpic:
		field = "epic"
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

// Helper to capitalize first letter
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}
```

**Expected output**: Selector appears centered with navigable options:
```
        +---------------------------+
        | Filter by Priority        |
        |                           |
        |   Urgent                  |
        | > High                    |
        |   Medium                  |
        |   Low                     |
        |                           |
        | j/k | Enter | Esc         |
        +---------------------------+
```

**Common Mistakes**:
- Not handling empty options list (e.g., no labels defined)
- Forgetting to reset selection when reopening
- List height calculation issues causing scrolling problems

---

### 5.5 Add Filter Message Types ✅ COMPLETED

**What**: Define all message types needed for filter operations.

**Why**: Bubble Tea uses messages for all state changes. Well-defined message types make the flow clear and enable type-safe handling.

**Steps**:

1. Update `internal/tui/messages.go` with filter messages
2. Add command creators for filter operations

**Code**:

**File**: `internal/tui/messages.go` (add to existing file)

```go
package tui

// =============================================================================
// Filter Messages
// =============================================================================

// FilterChangedMsg indicates filters have been updated
type FilterChangedMsg struct {
	FilterState *FilterState
}

// QuickFilterMsg triggers a quick filter key sequence
type QuickFilterMsg struct {
	Prefix rune // 'f' for filter commands
	Key    rune // 'p', 't', 'l', 'e', 'c'
}

// ShowSearchMsg triggers the search overlay
type ShowSearchMsg struct{}

// HideSearchMsg closes the search overlay
type HideSearchMsg struct{}

// ShowFilterSelectorMsg opens a filter selector
type ShowFilterSelectorMsg struct {
	Type FilterSelectorType
}

// ClearFiltersMsg clears all active filters
type ClearFiltersMsg struct{}

// ToggleFilterBarFocusMsg toggles focus on the filter bar
type ToggleFilterBarFocusMsg struct{}

// RefreshFilteredTasksMsg triggers re-filtering of tasks
type RefreshFilteredTasksMsg struct{}

// LoadLabelsMsg requests loading available labels
type LoadLabelsMsg struct{}

// LabelsLoadedMsg contains available labels
type LabelsLoadedMsg struct {
	Labels []string
}

// LoadEpicsMsg requests loading available epics
type LoadEpicsMsg struct{}

// EpicsLoadedMsg contains available epics
type EpicsLoadedMsg struct {
	Epics []EpicOption
}
```

**File**: `internal/tui/commands.go` (add to existing file)

```go
package tui

import (
	"sort"

	"github.com/pocketbase/pocketbase"
	tea "github.com/charmbracelet/bubbletea"
)

// =============================================================================
// Filter Commands
// =============================================================================

// CmdApplyFilter adds a filter and refreshes the view
func CmdApplyFilter(filterState *FilterState, filter Filter) tea.Cmd {
	return func() tea.Msg {
		filterState.AddFilter(filter)
		return FilterChangedMsg{FilterState: filterState}
	}
}

// CmdRemoveFilter removes a filter and refreshes the view
func CmdRemoveFilter(filterState *FilterState, filter Filter) tea.Cmd {
	return func() tea.Msg {
		filterState.RemoveFilter(filter)
		return FilterChangedMsg{FilterState: filterState}
	}
}

// CmdClearFilters clears all filters
func CmdClearFilters(filterState *FilterState) tea.Cmd {
	return func() tea.Msg {
		filterState.Clear()
		return FilterChangedMsg{FilterState: filterState}
	}
}

// CmdSetSearchQuery sets the search query
func CmdSetSearchQuery(filterState *FilterState, query string) tea.Cmd {
	return func() tea.Msg {
		filterState.SetSearchQuery(query)
		return FilterChangedMsg{FilterState: filterState}
	}
}

// CmdLoadLabels loads all unique labels from tasks
func CmdLoadLabels(app *pocketbase.PocketBase, boardID string) tea.Cmd {
	return func() tea.Msg {
		labels := make(map[string]bool)

		records, err := app.FindAllRecords("tasks")
		if err != nil {
			return LabelsLoadedMsg{Labels: []string{}}
		}

		for _, record := range records {
			// Filter by board if specified
			if boardID != "" && record.GetString("board") != boardID {
				continue
			}

			// Extract labels (stored as JSON array)
			taskLabels := record.Get("labels")
			if labelSlice, ok := taskLabels.([]interface{}); ok {
				for _, l := range labelSlice {
					if label, ok := l.(string); ok && label != "" {
						labels[label] = true
					}
				}
			}
		}

		// Convert to sorted slice
		result := make([]string, 0, len(labels))
		for label := range labels {
			result = append(result, label)
		}
		sort.Strings(result)

		return LabelsLoadedMsg{Labels: result}
	}
}

// CmdLoadEpics loads all epics for a board
func CmdLoadEpics(app *pocketbase.PocketBase, boardID string) tea.Cmd {
	return func() tea.Msg {
		records, err := app.FindAllRecords("epics")
		if err != nil {
			return EpicsLoadedMsg{Epics: []EpicOption{}}
		}

		epics := make([]EpicOption, 0, len(records))
		for _, record := range records {
			// Filter by board if specified
			if boardID != "" && record.GetString("board") != boardID {
				continue
			}

			epics = append(epics, EpicOption{
				ID:    record.Id,
				Title: record.GetString("title"),
				Color: record.GetString("color"),
			})
		}

		// Sort by title
		sort.Slice(epics, func(i, j int) bool {
			return epics[i].Title < epics[j].Title
		})

		return EpicsLoadedMsg{Epics: epics}
	}
}
```

**Expected output**: Messages compile and can be used in Update handlers:
```go
case FilterSelectedMsg:
    return m, CmdApplyFilter(m.filterState, msg.Filter)
```

**Common Mistakes**:
- Circular imports between messages.go and other files
- Forgetting to handle message types in the main Update loop
- Not returning the FilterState in messages (causes stale state)

---

### 5.6 Integrate Filtering into App Model ✅ COMPLETED

**What**: Wire up filtering components into the main TUI application.

**Why**: The App model orchestrates all components. Proper integration ensures filters apply correctly and UI state stays synchronized.

**Steps**:

1. Add filter state and components to App struct
2. Handle filter key bindings in Update
3. Apply filters when rendering columns
4. Handle filter-related messages

**Code**:

**File**: `internal/tui/app.go` (partial update - add to existing App struct and methods)

```go
package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pocketbase/pocketbase"
)

// App is the main TUI application model
type App struct {
	// ... existing fields ...
	pb           *pocketbase.PocketBase
	board        *Board
	columns      []Column
	focusedCol   int
	width        int
	height       int
	ready        bool

	// Filter state (add these)
	filterState    *FilterState
	filterBar      FilterBar
	searchOverlay  SearchOverlay
	filterSelector FilterSelector
	
	// Quick filter key state
	pendingFilterKey bool // True after 'f' is pressed
	
	// Cached filter data
	availableLabels []string
	availableEpics  []EpicOption
	
	// ... rest of existing fields ...
}

// NewApp creates a new TUI application
func NewApp(pb *pocketbase.PocketBase, boardRef string) *App {
	filterState := NewFilterState()
	
	return &App{
		pb:             pb,
		filterState:    filterState,
		filterBar:      NewFilterBar(filterState),
		searchOverlay:  NewSearchOverlay(filterState),
		filterSelector: NewFilterSelector(),
		pendingFilterKey: false,
	}
}

// Init implements tea.Model
func (m *App) Init() tea.Cmd {
	return tea.Batch(
		// ... existing init commands ...
		CmdLoadLabels(m.pb, m.getCurrentBoardID()),
		CmdLoadEpics(m.pb, m.getCurrentBoardID()),
	)
}

// Update implements tea.Model
func (m *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle overlays first (they capture input)
	if m.searchOverlay.IsActive() {
		overlay, cmd := m.searchOverlay.Update(msg)
		m.searchOverlay = overlay
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		// Check if search completed
		if !m.searchOverlay.IsActive() {
			cmds = append(cmds, m.refreshFilteredTasks)
		}
		return m, tea.Batch(cmds...)
	}

	if m.filterSelector.IsActive() {
		selector, cmd := m.filterSelector.Update(msg)
		m.filterSelector = selector
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.filterBar.SetWidth(msg.Width)
		m.searchOverlay.SetSize(msg.Width, msg.Height)
		m.filterSelector.SetSize(msg.Width, msg.Height)
		m.ready = true
		return m, nil

	case FilterSelectedMsg:
		m.filterState.AddFilter(msg.Filter)
		return m, m.refreshFilteredTasks

	case FilterCancelledMsg:
		return m, nil

	case SearchAppliedMsg:
		return m, m.refreshFilteredTasks

	case SearchCancelledMsg:
		return m, nil

	case FilterChangedMsg:
		return m, m.refreshFilteredTasks

	case LabelsLoadedMsg:
		m.availableLabels = msg.Labels
		return m, nil

	case EpicsLoadedMsg:
		m.availableEpics = msg.Epics
		return m, nil

	case ClearFiltersMsg:
		m.filterState.Clear()
		return m, m.refreshFilteredTasks
	}

	// ... rest of existing Update logic ...

	return m, tea.Batch(cmds...)
}

func (m *App) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Handle pending filter key (second key in 'fp', 'ft', etc.)
	if m.pendingFilterKey {
		m.pendingFilterKey = false
		return m.handleFilterKey(key)
	}

	switch key {
	case "/":
		// Open search
		cmd := m.searchOverlay.Show()
		return m, cmd

	case "f":
		// Start filter key sequence
		m.pendingFilterKey = true
		return m, nil

	// ... existing key handlers ...
	}

	return m, nil
}

func (m *App) handleFilterKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "p":
		// Filter by priority
		cmd := m.filterSelector.ShowPriority()
		return m, cmd

	case "t":
		// Filter by type
		cmd := m.filterSelector.ShowType()
		return m, cmd

	case "l":
		// Filter by label
		if len(m.availableLabels) == 0 {
			// No labels available
			return m, nil
		}
		cmd := m.filterSelector.ShowLabel(m.availableLabels)
		return m, cmd

	case "e":
		// Filter by epic
		if len(m.availableEpics) == 0 {
			// No epics available
			return m, nil
		}
		cmd := m.filterSelector.ShowEpic(m.availableEpics)
		return m, cmd

	case "c":
		// Clear all filters
		m.filterState.Clear()
		return m, m.refreshFilteredTasks

	default:
		// Unknown second key - ignore
		return m, nil
	}
}

// refreshFilteredTasks reapplies filters to columns
func (m *App) refreshFilteredTasks() tea.Msg {
	return RefreshFilteredTasksMsg{}
}

// getFilteredTasksForColumn returns tasks for a column with filters applied
func (m *App) getFilteredTasksForColumn(column string) []TaskItem {
	// Get all tasks for this column
	allTasks := m.getTasksForColumn(column)
	
	// Apply filters
	return m.filterState.Apply(allTasks)
}

// getCurrentBoardID returns the current board ID
func (m *App) getCurrentBoardID() string {
	if m.board == nil {
		return ""
	}
	return m.board.ID
}

// View implements tea.Model
func (m *App) View() string {
	if !m.ready {
		return "Loading..."
	}

	var sections []string

	// Header
	sections = append(sections, m.renderHeader())

	// Filter bar (only if filters active)
	if m.filterState.HasActiveFilters() {
		sections = append(sections, m.filterBar.View())
	}

	// Columns with filtered tasks
	sections = append(sections, m.renderColumns())

	// Status bar
	sections = append(sections, m.renderStatusBar())

	// Base view
	view := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Overlay search if active
	if m.searchOverlay.IsActive() {
		searchView := m.searchOverlay.View()
		view = m.overlayOnTop(view, searchView)
	}

	// Overlay filter selector if active
	if m.filterSelector.IsActive() {
		selectorView := m.filterSelector.View()
		view = m.overlayOnTop(view, selectorView)
	}

	return view
}

func (m *App) overlayOnTop(base, overlay string) string {
	// Simple overlay - just return overlay for now
	// A proper implementation would composite the views
	return overlay
}

func (m *App) renderStatusBar() string {
	style := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("252")).
		Width(m.width).
		Padding(0, 1)

	var hints []string

	if m.filterState.HasActiveFilters() {
		count := len(m.filterState.GetFilters())
		if m.filterState.GetSearchQuery() != "" {
			count++
		}
		hints = append(hints, lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Render("Filtering ("+string(rune('0'+count))+" active)"))
	}

	hints = append(hints, "/ search", "fp priority", "ft type", "fl label", "fe epic", "fc clear", "? help")

	return style.Render(strings.Join(hints, " | "))
}
```

**Expected output**: Filters work with keyboard shortcuts:
- `/` opens search overlay
- `fp` opens priority selector
- `ft` opens type selector
- `fl` opens label selector
- `fe` opens epic selector
- `fc` clears all filters

**Common Mistakes**:
- Forgetting to refresh columns after filter changes
- Not passing filterState by reference (causes stale data)
- Overlay not receiving key events (check active state order)

---

### 5.7 Add Keybindings for Filters ✅ COMPLETED

**What**: Define and document all filter-related keybindings.

**Why**: Consistent keybindings improve usability. Documenting them in the keys.go file keeps them discoverable.

**Steps**:

1. Update `internal/tui/keys.go` with filter bindings
2. Add to help text

**Code**:

**File**: `internal/tui/keys.go` (add to existing keyMap)

```go
package tui

import (
	"github.com/charmbracelet/bubbles/key"
)

// KeyMap defines all keybindings for the TUI
type KeyMap struct {
	// Navigation
	Up        key.Binding
	Down      key.Binding
	Left      key.Binding
	Right     key.Binding
	FirstItem key.Binding
	LastItem  key.Binding

	// Actions
	Enter  key.Binding
	New    key.Binding
	Edit   key.Binding
	Delete key.Binding
	Select key.Binding

	// Movement
	MoveLeft  key.Binding
	MoveRight key.Binding
	MoveUp    key.Binding
	MoveDown  key.Binding

	// Filtering (add these)
	Search         key.Binding
	FilterPriority key.Binding
	FilterType     key.Binding
	FilterLabel    key.Binding
	FilterEpic     key.Binding
	ClearFilters   key.Binding

	// Global
	Quit    key.Binding
	Help    key.Binding
	Board   key.Binding
	Refresh key.Binding
	Escape  key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Navigation
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("k/up", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("j/down", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("h/left", "prev column"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("l/right", "next column"),
		),
		FirstItem: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "first item"),
		),
		LastItem: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "last item"),
		),

		// Actions
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "view details"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new task"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit task"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete task"),
		),
		Select: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "select"),
		),

		// Movement
		MoveLeft: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "move task left"),
		),
		MoveRight: key.NewBinding(
			key.WithKeys("L"),
			key.WithHelp("L", "move task right"),
		),
		MoveUp: key.NewBinding(
			key.WithKeys("K"),
			key.WithHelp("K", "move task up"),
		),
		MoveDown: key.NewBinding(
			key.WithKeys("J"),
			key.WithHelp("J", "move task down"),
		),

		// Filtering
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		FilterPriority: key.NewBinding(
			key.WithKeys("f p"),
			key.WithHelp("fp", "filter priority"),
		),
		FilterType: key.NewBinding(
			key.WithKeys("f t"),
			key.WithHelp("ft", "filter type"),
		),
		FilterLabel: key.NewBinding(
			key.WithKeys("f l"),
			key.WithHelp("fl", "filter label"),
		),
		FilterEpic: key.NewBinding(
			key.WithKeys("f e"),
			key.WithHelp("fe", "filter epic"),
		),
		ClearFilters: key.NewBinding(
			key.WithKeys("f c"),
			key.WithHelp("fc", "clear filters"),
		),

		// Global
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Board: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "switch board"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "refresh"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel/close"),
		),
	}
}

// ShortHelp returns a short help string for the filter commands
func (k KeyMap) FilterHelp() []key.Binding {
	return []key.Binding{
		k.Search,
		k.FilterPriority,
		k.FilterType,
		k.FilterLabel,
		k.FilterEpic,
		k.ClearFilters,
	}
}

// FullHelp returns the full help for all bindings
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right, k.FirstItem, k.LastItem},
		{k.Enter, k.New, k.Edit, k.Delete, k.Select},
		{k.MoveLeft, k.MoveRight, k.MoveUp, k.MoveDown},
		{k.Search, k.FilterPriority, k.FilterType, k.FilterLabel, k.FilterEpic, k.ClearFilters},
		{k.Board, k.Help, k.Refresh, k.Quit},
	}
}
```

**Expected output**: Keybindings appear in help overlay:
```
Filter Commands:
  /   search       fp  filter priority   ft  filter type
  fl  filter label fe  filter epic       fc  clear filters
```

**Common Mistakes**:
- Two-key sequences require special handling (can't use `key.WithKeys("fp")`)
- Help text getting too long for screen width
- Forgetting to add new bindings to FullHelp()

---

### 5.8 Show Active Filter Count in Header ✅ COMPLETED

**What**: Display the number of active filters in the board header.

**Why**: Visual feedback helps users understand why they might see fewer tasks than expected.

**Steps**:

1. Update header rendering to include filter indicator
2. Show count and optional summary

**Code**:

**File**: `internal/tui/app.go` (update renderHeader method)

```go
func (m *App) renderHeader() string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Padding(0, 1)

	var parts []string

	// Board name
	boardName := "Kanban Board"
	if m.board != nil {
		boardName = m.board.Prefix + " - " + m.board.Name
	}
	parts = append(parts, headerStyle.Render(boardName))

	// Task count
	totalTasks := m.getTotalTaskCount()
	filteredTasks := m.getFilteredTaskCount()

	countStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	if m.filterState.HasActiveFilters() {
		// Show filtered/total
		filterIndicator := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true).
			Render("FILTERED")

		countText := countStyle.Render(
			fmt.Sprintf("(%d/%d tasks)", filteredTasks, totalTasks),
		)
		parts = append(parts, filterIndicator, countText)
	} else {
		// Show total only
		countText := countStyle.Render(
			fmt.Sprintf("(%d tasks)", totalTasks),
		)
		parts = append(parts, countText)
	}

	// Filter summary (what's active)
	if m.filterState.HasActiveFilters() {
		summary := m.getFilterSummary()
		if summary != "" {
			summaryStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("62")).
				Italic(true)
			parts = append(parts, summaryStyle.Render("| "+summary))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, parts...)
}

func (m *App) getTotalTaskCount() int {
	count := 0
	for _, col := range m.columns {
		count += len(col.AllItems()) // Unfiltered count
	}
	return count
}

func (m *App) getFilteredTaskCount() int {
	count := 0
	for _, col := range m.columns {
		count += len(col.Items()) // Filtered count
	}
	return count
}

func (m *App) getFilterSummary() string {
	var parts []string

	if q := m.filterState.GetSearchQuery(); q != "" {
		parts = append(parts, fmt.Sprintf("search: %q", truncate(q, 15)))
	}

	filters := m.filterState.GetFilters()
	for _, f := range filters {
		if len(parts) >= 3 {
			remaining := len(filters) - 3
			if m.filterState.GetSearchQuery() != "" {
				remaining++
			}
			parts = append(parts, fmt.Sprintf("+%d more", remaining))
			break
		}
		parts = append(parts, f.String())
	}

	return strings.Join(parts, ", ")
}
```

**Expected output**: Header shows filter state:
```
WRK - Work Board  FILTERED (5/23 tasks) | Priority: High, Type: Bug
```

**Common Mistakes**:
- Confusing filtered count with total count
- Summary getting too long and wrapping

---

### 5.9 Persist Filter State During Session

**What**: Keep filters active when switching between views (but not between sessions).

**Why**: Users shouldn't lose their filter context when viewing a task detail or switching boards.

**Steps**:

1. Store filter state in App (already done)
2. Restore filter state when returning from overlays
3. Optionally save to session file for persistence across restarts

**Code**:

**File**: `internal/tui/session.go`

```go
package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// SessionState holds state that persists during a TUI session
type SessionState struct {
	FilterState    *FilterState `json:"filterState"`
	CurrentBoardID string       `json:"currentBoardID"`
	FocusedColumn  int          `json:"focusedColumn"`
}

// sessionFilePath returns the path to the session state file
func sessionFilePath() string {
	// Store in temp directory - cleared on reboot
	return filepath.Join(os.TempDir(), "egenskriven-tui-session.json")
}

// SaveSession saves current session state
func SaveSession(state SessionState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(sessionFilePath(), data, 0644)
}

// LoadSession loads session state if available
func LoadSession() (*SessionState, error) {
	data, err := os.ReadFile(sessionFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No session to load
		}
		return nil, err
	}

	var state SessionState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// ClearSession removes saved session state
func ClearSession() error {
	path := sessionFilePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // Nothing to clear
	}
	return os.Remove(path)
}
```

**File**: `internal/tui/app.go` (add session handling)

```go
// In Init(), load previous session
func (m *App) Init() tea.Cmd {
	return tea.Batch(
		m.loadSession,
		// ... other init commands ...
	)
}

func (m *App) loadSession() tea.Msg {
	session, err := LoadSession()
	if err != nil || session == nil {
		return nil
	}
	return SessionLoadedMsg{Session: session}
}

// Handle session loaded
case SessionLoadedMsg:
	if msg.Session != nil {
		if msg.Session.FilterState != nil {
			m.filterState = msg.Session.FilterState
			m.filterBar = NewFilterBar(m.filterState)
			m.searchOverlay = NewSearchOverlay(m.filterState)
		}
		m.focusedCol = msg.Session.FocusedColumn
		// Board switch would need separate handling
	}
	return m, nil

// On quit, save session
case tea.KeyMsg:
	if key.Matches(msg, m.keys.Quit) {
		_ = SaveSession(SessionState{
			FilterState:    m.filterState,
			CurrentBoardID: m.getCurrentBoardID(),
			FocusedColumn:  m.focusedCol,
		})
		return m, tea.Quit
	}
```

**Expected output**: Filters persist when you quit and relaunch the TUI (within the same system session).

**Common Mistakes**:
- Saving to a location that requires elevated permissions
- Not handling missing session file gracefully
- Storing sensitive data in session (filter values are generally safe)

---

### 5.10 Write Tests for Filter Logic

**What**: Comprehensive tests for all filter operations.

**Why**: Filters are critical for usability. Bugs here cause users to miss important tasks.

**Steps**:

1. Create `internal/tui/filter_test.go`
2. Test each filter type
3. Test combinations
4. Test edge cases

**Code**:

**File**: `internal/tui/filter_test.go`

```go
package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterState_SearchQuery(t *testing.T) {
	fs := NewFilterState()

	// Initially empty
	assert.Empty(t, fs.GetSearchQuery())
	assert.False(t, fs.HasActiveFilters())

	// Set query
	fs.SetSearchQuery("test")
	assert.Equal(t, "test", fs.GetSearchQuery())
	assert.True(t, fs.HasActiveFilters())

	// Clear
	fs.Clear()
	assert.Empty(t, fs.GetSearchQuery())
	assert.False(t, fs.HasActiveFilters())
}

func TestFilterState_AddRemoveFilter(t *testing.T) {
	fs := NewFilterState()

	// Add filter
	filter := Filter{Field: "priority", Operator: "is", Value: "high"}
	fs.AddFilter(filter)

	filters := fs.GetFilters()
	require.Len(t, filters, 1)
	assert.Equal(t, "priority", filters[0].Field)

	// Add same field replaces
	filter2 := Filter{Field: "priority", Operator: "is", Value: "urgent"}
	fs.AddFilter(filter2)

	filters = fs.GetFilters()
	require.Len(t, filters, 1)
	assert.Equal(t, "urgent", filters[0].Value)

	// Remove filter
	fs.RemoveFilter(filter2)
	assert.Empty(t, fs.GetFilters())
}

func TestFilterState_LabelFiltersStack(t *testing.T) {
	fs := NewFilterState()

	// Labels can have multiple values
	fs.AddFilter(Filter{Field: "label", Operator: "is", Value: "frontend"})
	fs.AddFilter(Filter{Field: "label", Operator: "is", Value: "urgent"})

	filters := fs.GetFilters()
	assert.Len(t, filters, 2) // Both should exist
}

func TestFilterState_Apply_SearchQuery(t *testing.T) {
	fs := NewFilterState()
	fs.SetSearchQuery("dark mode")

	tasks := []TaskItem{
		{ID: "1", Title: "Implement dark mode", Description: "Add theme support"},
		{ID: "2", Title: "Fix login bug", Description: "Users can't login"},
		{ID: "3", Title: "Update docs", Description: "Document dark mode feature"},
	}

	result := fs.Apply(tasks)

	assert.Len(t, result, 2)
	assert.Equal(t, "1", result[0].ID)
	assert.Equal(t, "3", result[1].ID)
}

func TestFilterState_Apply_SearchQuery_CaseInsensitive(t *testing.T) {
	fs := NewFilterState()
	fs.SetSearchQuery("DARK MODE")

	tasks := []TaskItem{
		{ID: "1", Title: "implement dark mode", Description: ""},
	}

	result := fs.Apply(tasks)
	assert.Len(t, result, 1)
}

func TestFilterState_Apply_SearchQuery_MatchesDisplayID(t *testing.T) {
	fs := NewFilterState()
	fs.SetSearchQuery("WRK-123")

	tasks := []TaskItem{
		{ID: "1", DisplayID: "WRK-123", Title: "Some task"},
		{ID: "2", DisplayID: "WRK-456", Title: "Another task"},
	}

	result := fs.Apply(tasks)
	assert.Len(t, result, 1)
	assert.Equal(t, "WRK-123", result[0].DisplayID)
}

func TestFilterState_Apply_PriorityFilter(t *testing.T) {
	fs := NewFilterState()
	fs.AddFilter(Filter{Field: "priority", Operator: "is", Value: "high"})

	tasks := []TaskItem{
		{ID: "1", Priority: "high"},
		{ID: "2", Priority: "medium"},
		{ID: "3", Priority: "High"}, // Case variation
		{ID: "4", Priority: "low"},
	}

	result := fs.Apply(tasks)

	assert.Len(t, result, 2)
	assert.Equal(t, "1", result[0].ID)
	assert.Equal(t, "3", result[1].ID)
}

func TestFilterState_Apply_PriorityFilter_IsNot(t *testing.T) {
	fs := NewFilterState()
	fs.AddFilter(Filter{Field: "priority", Operator: "is_not", Value: "low"})

	tasks := []TaskItem{
		{ID: "1", Priority: "high"},
		{ID: "2", Priority: "low"},
		{ID: "3", Priority: "medium"},
	}

	result := fs.Apply(tasks)

	assert.Len(t, result, 2)
	assert.Equal(t, "1", result[0].ID)
	assert.Equal(t, "3", result[1].ID)
}

func TestFilterState_Apply_TypeFilter(t *testing.T) {
	fs := NewFilterState()
	fs.AddFilter(Filter{Field: "type", Operator: "is", Value: "bug"})

	tasks := []TaskItem{
		{ID: "1", Type: "bug"},
		{ID: "2", Type: "feature"},
		{ID: "3", Type: "bug"},
	}

	result := fs.Apply(tasks)

	assert.Len(t, result, 2)
}

func TestFilterState_Apply_LabelFilter(t *testing.T) {
	fs := NewFilterState()
	fs.AddFilter(Filter{Field: "label", Operator: "is", Value: "frontend"})

	tasks := []TaskItem{
		{ID: "1", Labels: []string{"frontend", "urgent"}},
		{ID: "2", Labels: []string{"backend"}},
		{ID: "3", Labels: []string{"frontend"}},
		{ID: "4", Labels: nil}, // No labels
	}

	result := fs.Apply(tasks)

	assert.Len(t, result, 2)
	assert.Equal(t, "1", result[0].ID)
	assert.Equal(t, "3", result[1].ID)
}

func TestFilterState_Apply_LabelFilter_MultiplLabels(t *testing.T) {
	fs := NewFilterState()
	// Must have BOTH labels
	fs.AddFilter(Filter{Field: "label", Operator: "is", Value: "frontend"})
	fs.AddFilter(Filter{Field: "label", Operator: "is", Value: "urgent"})

	tasks := []TaskItem{
		{ID: "1", Labels: []string{"frontend", "urgent"}},
		{ID: "2", Labels: []string{"frontend"}},
		{ID: "3", Labels: []string{"urgent"}},
	}

	result := fs.Apply(tasks)

	assert.Len(t, result, 1)
	assert.Equal(t, "1", result[0].ID)
}

func TestFilterState_Apply_EpicFilter(t *testing.T) {
	fs := NewFilterState()
	fs.AddFilter(Filter{Field: "epic", Operator: "is", Value: "epic-123"})

	tasks := []TaskItem{
		{ID: "1", Epic: "epic-123", EpicTitle: "Q1 Launch"},
		{ID: "2", Epic: "epic-456", EpicTitle: "Tech Debt"},
		{ID: "3", Epic: "", EpicTitle: ""},
	}

	result := fs.Apply(tasks)

	assert.Len(t, result, 1)
	assert.Equal(t, "1", result[0].ID)
}

func TestFilterState_Apply_EpicFilter_ByTitle(t *testing.T) {
	fs := NewFilterState()
	fs.AddFilter(Filter{Field: "epic", Operator: "is", Value: "Q1 Launch"})

	tasks := []TaskItem{
		{ID: "1", Epic: "epic-123", EpicTitle: "Q1 Launch"},
		{ID: "2", Epic: "epic-456", EpicTitle: "Tech Debt"},
	}

	result := fs.Apply(tasks)

	assert.Len(t, result, 1)
}

func TestFilterState_Apply_CombinedFilters(t *testing.T) {
	fs := NewFilterState()
	fs.SetSearchQuery("login")
	fs.AddFilter(Filter{Field: "priority", Operator: "is", Value: "high"})
	fs.AddFilter(Filter{Field: "type", Operator: "is", Value: "bug"})

	tasks := []TaskItem{
		{ID: "1", Title: "Fix login bug", Priority: "high", Type: "bug"},
		{ID: "2", Title: "Fix login bug", Priority: "low", Type: "bug"},
		{ID: "3", Title: "Fix login bug", Priority: "high", Type: "feature"},
		{ID: "4", Title: "Add logout", Priority: "high", Type: "bug"},
	}

	result := fs.Apply(tasks)

	// Only task 1 matches all criteria
	assert.Len(t, result, 1)
	assert.Equal(t, "1", result[0].ID)
}

func TestFilterState_Apply_NoFilters(t *testing.T) {
	fs := NewFilterState()

	tasks := []TaskItem{
		{ID: "1"},
		{ID: "2"},
		{ID: "3"},
	}

	result := fs.Apply(tasks)

	// All tasks should be returned
	assert.Len(t, result, 3)
}

func TestFilterState_Apply_EmptyTasks(t *testing.T) {
	fs := NewFilterState()
	fs.SetSearchQuery("test")

	result := fs.Apply([]TaskItem{})

	assert.Empty(t, result)
}

func TestFilterState_ToJSON_FromJSON(t *testing.T) {
	original := NewFilterState()
	original.SetSearchQuery("test query")
	original.AddFilter(Filter{Field: "priority", Operator: "is", Value: "high"})
	original.AddFilter(Filter{Field: "type", Operator: "is", Value: "bug"})

	// Serialize
	data, err := original.ToJSON()
	require.NoError(t, err)

	// Deserialize
	restored := NewFilterState()
	err = restored.FromJSON(data)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, "test query", restored.GetSearchQuery())
	assert.Len(t, restored.GetFilters(), 2)
}

func TestFilter_String(t *testing.T) {
	tests := []struct {
		filter   Filter
		expected string
	}{
		{
			filter:   Filter{Field: "priority", Value: "high", Display: "Priority: High"},
			expected: "Priority: High",
		},
		{
			filter:   Filter{Field: "priority", Value: "high"},
			expected: "priority:high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.filter.String())
		})
	}
}

func TestFilterState_RemoveFilterByField(t *testing.T) {
	fs := NewFilterState()
	fs.AddFilter(Filter{Field: "priority", Value: "high"})
	fs.AddFilter(Filter{Field: "type", Value: "bug"})
	fs.AddFilter(Filter{Field: "label", Value: "frontend"})
	fs.AddFilter(Filter{Field: "label", Value: "urgent"})

	// Remove all labels
	fs.RemoveFilterByField("label")

	filters := fs.GetFilters()
	assert.Len(t, filters, 2)
	for _, f := range filters {
		assert.NotEqual(t, "label", f.Field)
	}
}
```

**Expected output**:
```bash
go test ./internal/tui/... -v
=== RUN   TestFilterState_SearchQuery
--- PASS: TestFilterState_SearchQuery (0.00s)
=== RUN   TestFilterState_AddRemoveFilter
--- PASS: TestFilterState_AddRemoveFilter (0.00s)
...
PASS
```

**Common Mistakes**:
- Not testing case insensitivity
- Missing nil/empty slice tests
- Forgetting to test filter combinations

---

## Verification Checklist

Complete each verification in order. Check off each item.

### Build Verification

- [ ] **All files compile**
  ```bash
  go build ./internal/tui/...
  ```
  Should complete without errors.

- [ ] **Tests pass**
  ```bash
  go test ./internal/tui/... -v
  ```
  All tests should pass.

### Functional Verification

- [ ] **Search works**
  - Press `/` - search overlay appears
  - Type "bug" - input shows text
  - Press Enter - overlay closes, tasks filtered
  - Only tasks with "bug" in title/description visible

- [ ] **Priority filter works**
  - Press `fp` - priority selector appears
  - Navigate to "High" with j/k
  - Press Enter - selector closes
  - Only high priority tasks visible
  - Filter chip appears in filter bar

- [ ] **Type filter works**
  - Press `ft` - type selector appears
  - Select "Bug"
  - Only bug-type tasks visible

- [ ] **Label filter works**
  - Press `fl` - label selector appears
  - Select a label
  - Only tasks with that label visible

- [ ] **Epic filter works**
  - Press `fe` - epic selector appears
  - Select an epic
  - Only tasks in that epic visible

- [ ] **Clear filters works**
  - Have active filters
  - Press `fc`
  - All tasks visible again
  - Filter bar empty

- [ ] **Filter chips show correctly**
  - Apply multiple filters
  - Filter bar shows all as chips
  - Chips have correct colors

- [ ] **Filter chips can be removed**
  - Focus filter bar
  - Navigate to chip with h/l
  - Press x or backspace
  - That filter removed

- [ ] **Combined filters work**
  - Apply search + priority + type
  - Only tasks matching ALL criteria visible
  - Header shows correct count

### Edge Cases

- [ ] **Empty search**
  - Press `/`, immediately press Enter
  - No filter applied

- [ ] **No labels available**
  - Press `fl` when no tasks have labels
  - Nothing happens (or shows "No labels" message)

- [ ] **No matching tasks**
  - Apply filter that matches nothing
  - Columns show "No tasks" or are empty
  - No errors

- [ ] **Filter persists across views**
  - Apply filter
  - Open task detail
  - Close task detail
  - Filter still active

- [ ] **Escape closes overlays**
  - Open search, press Esc
  - Search closes, no filter applied
  - Open selector, press Esc
  - Selector closes, no filter applied

### Performance

- [ ] **Filtering is fast**
  - With 100+ tasks, filtering feels instant
  - No visible lag on keypress

---

## File Summary

| File | Lines | Purpose |
|------|-------|---------|
| `internal/tui/filter.go` | ~200 | Filter state and matching logic |
| `internal/tui/filter_bar.go` | ~250 | Filter bar UI component |
| `internal/tui/search.go` | ~150 | Search overlay component |
| `internal/tui/filter_selector.go` | ~250 | Reusable filter selector |
| `internal/tui/messages.go` | +50 | Filter message types |
| `internal/tui/commands.go` | +100 | Filter commands |
| `internal/tui/keys.go` | +30 | Filter keybindings |
| `internal/tui/session.go` | ~80 | Session persistence |
| `internal/tui/filter_test.go` | ~300 | Comprehensive tests |
| `internal/tui/app.go` | +150 | Integration updates |

**Total new code**: ~1,500 lines

---

## What You Should Have Now

After completing Phase 5, your TUI should:

```
internal/tui/
├── app.go              # Updated with filter integration
├── filter.go           # ✓ Created
├── filter_bar.go       # ✓ Created
├── filter_selector.go  # ✓ Created
├── filter_test.go      # ✓ Created
├── search.go           # ✓ Created
├── session.go          # ✓ Created
├── messages.go         # Updated with filter messages
├── commands.go         # Updated with filter commands
├── keys.go             # Updated with filter bindings
└── ... (existing files)
```

### Feature Summary

| Feature | Status |
|---------|--------|
| Search by title/description | Complete |
| Filter by priority | Complete |
| Filter by type | Complete |
| Filter by label | Complete |
| Filter by epic | Complete |
| Clear all filters | Complete |
| Visual filter chips | Complete |
| Filter count in header | Complete |
| Keyboard navigation | Complete |
| Session persistence | Complete |
| Comprehensive tests | Complete |

---

## Next Phase

**Phase 6: Advanced Features** will add:
- Epic display and epic-level views
- Blocked tasks indicator and dependency visualization
- Due date display with overdue highlighting
- Subtask display (expandable tree)
- Multi-select tasks with Space
- Bulk operations (move, delete multiple)
- Help overlay with full keybinding reference
- Command palette (Ctrl+K) for quick actions

---

## Troubleshooting

### Filter key sequence not working

**Problem**: Pressing `fp` does nothing.

**Solution**: The two-key sequence requires proper state tracking:
1. First key `f` sets `pendingFilterKey = true`
2. Second key `p` is handled in `handleFilterKey()`
3. If you handle keys elsewhere, the state might reset

Check that no other key handler intercepts `f` before setting the pending state.

### Search overlay doesn't close

**Problem**: After pressing Enter, overlay stays open.

**Solution**: Ensure `Hide()` is called and `IsActive()` returns false:
```go
case "enter":
    s.Hide()  // Must be called
    return s, ...
```

### Filter chips not updating

**Problem**: Applying a filter doesn't show the chip.

**Solution**: 
1. Verify `FilterChangedMsg` is sent
2. Verify `filterState` is passed by pointer
3. Check `HasActiveFilters()` returns true
4. Ensure `filterBar.View()` is called in render

### Labels/Epics selector is empty

**Problem**: Pressing `fl` or `fe` shows empty selector.

**Solution**:
1. Verify `CmdLoadLabels` and `CmdLoadEpics` are called in `Init()`
2. Check that tasks actually have labels/epics set
3. Verify the board filter in load commands matches current board

### Filter matching is wrong

**Problem**: Filter shows/hides wrong tasks.

**Solution**: Check the filter logic in `matchesFilter()`:
1. Are you comparing case-insensitively?
2. Is the operator being respected (`is` vs `is_not`)?
3. For labels, are you checking the slice correctly?

Run the tests to isolate the issue:
```bash
go test ./internal/tui/... -run TestFilterState -v
```

### Session not persisting

**Problem**: Filters lost after quitting and relaunching.

**Solution**:
1. Check temp directory permissions: `ls -la /tmp/egenskriven-*`
2. Verify `SaveSession` is called before quit
3. Verify `LoadSession` is called in `Init()`
4. Check for JSON marshaling errors

---

## Glossary

| Term | Definition |
|------|------------|
| **Filter** | A condition that hides non-matching tasks |
| **Filter State** | The collection of all active filters and search query |
| **Filter Bar** | UI component showing active filters as chips |
| **Filter Chip** | Visual indicator for a single active filter |
| **Filter Selector** | Modal overlay for choosing filter values |
| **Quick Filter Keys** | Two-key sequences like `fp` for fast filtering |
| **Search Query** | Text searched in task title, description, and ID |
| **AND Logic** | Multiple filters must all match (not OR) |
| **Session State** | Filter settings persisted during TUI usage |
| **Overlay** | Modal component that appears on top of main view |
