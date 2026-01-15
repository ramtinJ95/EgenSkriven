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
		{ID: "1", TaskTitle: "Implement dark mode", TaskDescription: "Add theme support"},
		{ID: "2", TaskTitle: "Fix login bug", TaskDescription: "Users can't login"},
		{ID: "3", TaskTitle: "Update docs", TaskDescription: "Document dark mode feature"},
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
		{ID: "1", TaskTitle: "implement dark mode", TaskDescription: ""},
	}

	result := fs.Apply(tasks)
	assert.Len(t, result, 1)
}

func TestFilterState_Apply_SearchQuery_MatchesDisplayID(t *testing.T) {
	fs := NewFilterState()
	fs.SetSearchQuery("WRK-123")

	tasks := []TaskItem{
		{ID: "1", DisplayID: "WRK-123", TaskTitle: "Some task"},
		{ID: "2", DisplayID: "WRK-456", TaskTitle: "Another task"},
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

func TestFilterState_Apply_LabelFilter_MultipleLabels(t *testing.T) {
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
		{ID: "1", EpicID: "epic-123", EpicTitle: "Q1 Launch"},
		{ID: "2", EpicID: "epic-456", EpicTitle: "Tech Debt"},
		{ID: "3", EpicID: "", EpicTitle: ""},
	}

	result := fs.Apply(tasks)

	assert.Len(t, result, 1)
	assert.Equal(t, "1", result[0].ID)
}

func TestFilterState_Apply_EpicFilter_ByTitle(t *testing.T) {
	fs := NewFilterState()
	fs.AddFilter(Filter{Field: "epic", Operator: "is", Value: "Q1 Launch"})

	tasks := []TaskItem{
		{ID: "1", EpicID: "epic-123", EpicTitle: "Q1 Launch"},
		{ID: "2", EpicID: "epic-456", EpicTitle: "Tech Debt"},
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
		{ID: "1", TaskTitle: "Fix login bug", Priority: "high", Type: "bug"},
		{ID: "2", TaskTitle: "Fix login bug", Priority: "low", Type: "bug"},
		{ID: "3", TaskTitle: "Fix login bug", Priority: "high", Type: "feature"},
		{ID: "4", TaskTitle: "Add logout", Priority: "high", Type: "bug"},
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
