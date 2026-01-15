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
		titleMatch := strings.Contains(strings.ToLower(task.TaskTitle), query)
		descMatch := strings.Contains(strings.ToLower(task.TaskDescription), query)
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
		return task.EpicID == filter.Value ||
			strings.EqualFold(task.EpicTitle, filter.Value)
	case "is_not":
		return task.EpicID != filter.Value &&
			!strings.EqualFold(task.EpicTitle, filter.Value)
	default:
		return task.EpicID == filter.Value ||
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
