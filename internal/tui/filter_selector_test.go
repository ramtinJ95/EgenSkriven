package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestFilterSelector_New(t *testing.T) {
	fs := NewFilterSelector()

	assert.False(t, fs.active)
	assert.Equal(t, FilterSelectorNone, fs.selectorType)
}

func TestFilterSelector_ShowPriority(t *testing.T) {
	fs := NewFilterSelector()
	fs.SetSize(80, 24)

	fs.ShowPriority()

	assert.True(t, fs.IsActive())
	assert.Equal(t, FilterSelectorPriority, fs.GetType())
}

func TestFilterSelector_ShowType(t *testing.T) {
	fs := NewFilterSelector()
	fs.SetSize(80, 24)

	fs.ShowType()

	assert.True(t, fs.IsActive())
	assert.Equal(t, FilterSelectorTypeFilter, fs.GetType())
}

func TestFilterSelector_ShowLabel(t *testing.T) {
	fs := NewFilterSelector()
	fs.SetSize(80, 24)

	labels := []string{"frontend", "backend", "urgent"}
	fs.ShowLabel(labels)

	assert.True(t, fs.IsActive())
	assert.Equal(t, FilterSelectorLabel, fs.GetType())
}

func TestFilterSelector_ShowLabel_Empty(t *testing.T) {
	fs := NewFilterSelector()
	fs.SetSize(80, 24)

	// Should handle empty labels
	fs.ShowLabel([]string{})

	assert.True(t, fs.IsActive())
	assert.Equal(t, FilterSelectorLabel, fs.GetType())
}

func TestFilterSelector_ShowEpic(t *testing.T) {
	fs := NewFilterSelector()
	fs.SetSize(80, 24)

	epics := []EpicOption{
		{ID: "epic-1", Title: "Q1 Launch", Color: "blue"},
		{ID: "epic-2", Title: "Tech Debt", Color: "red"},
	}
	fs.ShowEpic(epics)

	assert.True(t, fs.IsActive())
	assert.Equal(t, FilterSelectorEpic, fs.GetType())
}

func TestFilterSelector_Hide(t *testing.T) {
	fs := NewFilterSelector()
	fs.SetSize(80, 24)
	fs.ShowPriority()

	assert.True(t, fs.IsActive())

	fs.Hide()

	assert.False(t, fs.IsActive())
	assert.Equal(t, FilterSelectorNone, fs.GetType())
}

func TestFilterSelector_SetSize(t *testing.T) {
	fs := NewFilterSelector()

	fs.SetSize(100, 50)

	assert.Equal(t, 100, fs.width)
	assert.Equal(t, 50, fs.height)
}

func TestFilterSelector_View_Inactive(t *testing.T) {
	fs := NewFilterSelector()
	fs.SetSize(80, 24)

	view := fs.View()
	assert.Empty(t, view)
}

func TestFilterSelector_View_Active(t *testing.T) {
	fs := NewFilterSelector()
	fs.SetSize(80, 24)
	fs.ShowPriority()

	view := fs.View()
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Filter by Priority")
	assert.Contains(t, view, "Enter to select")
}

func TestFilterSelector_Update_Escape(t *testing.T) {
	fs := NewFilterSelector()
	fs.SetSize(80, 24)
	fs.ShowPriority()

	fs, cmd := fs.Update(tea.KeyMsg{Type: tea.KeyEsc})

	assert.False(t, fs.IsActive())
	assert.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(FilterCancelledMsg)
	assert.True(t, ok)
}

func TestFilterSelector_Update_Q(t *testing.T) {
	fs := NewFilterSelector()
	fs.SetSize(80, 24)
	fs.ShowPriority()

	fs, cmd := fs.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	assert.False(t, fs.IsActive())
	assert.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(FilterCancelledMsg)
	assert.True(t, ok)
}

func TestFilterSelector_Update_Inactive(t *testing.T) {
	fs := NewFilterSelector()
	fs.SetSize(80, 24)
	// Don't call Show

	fs, cmd := fs.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.False(t, fs.IsActive())
	assert.Nil(t, cmd)
}

func TestFilterSelector_Update_Enter(t *testing.T) {
	fs := NewFilterSelector()
	fs.SetSize(80, 24)
	fs.ShowPriority()

	// Select by pressing Enter on the first item (Urgent)
	fs, cmd := fs.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.False(t, fs.IsActive())
	assert.NotNil(t, cmd)

	msg := cmd()
	selectedMsg, ok := msg.(FilterSelectedMsg)
	assert.True(t, ok)
	assert.Equal(t, FilterSelectorPriority, selectedMsg.Type)
	assert.Equal(t, "priority", selectedMsg.Filter.Field)
	assert.Equal(t, "urgent", selectedMsg.Filter.Value)
}

func TestFilterOption_ListItem(t *testing.T) {
	opt := FilterOption{
		Value:   "high",
		Display: "High Priority",
		Color:   "208",
	}

	assert.Equal(t, "High Priority", opt.FilterValue())
	assert.Equal(t, "High Priority", opt.Title())
	assert.Empty(t, opt.Description())
}

func TestCapitalizeFirst(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"priority", "Priority"},
		{"type", "Type"},
		{"label", "Label"},
		{"epic", "Epic"},
		{"", ""},
		{"Already", "Already"},
		{"123abc", "123abc"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := capitalizeFirst(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateFilterFromOption(t *testing.T) {
	tests := []struct {
		sType    FilterSelectorType
		option   FilterOption
		expected Filter
	}{
		{
			sType:  FilterSelectorPriority,
			option: FilterOption{Value: "high", Display: "High"},
			expected: Filter{
				Field:    "priority",
				Operator: "is",
				Value:    "high",
				Display:  "Priority: High",
			},
		},
		{
			sType:  FilterSelectorTypeFilter,
			option: FilterOption{Value: "bug", Display: "Bug"},
			expected: Filter{
				Field:    "type",
				Operator: "is",
				Value:    "bug",
				Display:  "Type: Bug",
			},
		},
		{
			sType:  FilterSelectorLabel,
			option: FilterOption{Value: "frontend", Display: "frontend"},
			expected: Filter{
				Field:    "label",
				Operator: "is",
				Value:    "frontend",
				Display:  "Label: frontend",
			},
		},
		{
			sType:  FilterSelectorEpic,
			option: FilterOption{Value: "epic-123", Display: "Q1 Launch"},
			expected: Filter{
				Field:    "epic",
				Operator: "is",
				Value:    "epic-123",
				Display:  "Epic: Q1 Launch",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected.Field, func(t *testing.T) {
			result := createFilterFromOption(tt.sType, tt.option)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFilterSelectedMsg(t *testing.T) {
	msg := FilterSelectedMsg{
		Type: FilterSelectorPriority,
		Filter: Filter{
			Field: "priority",
			Value: "high",
		},
	}

	assert.Equal(t, FilterSelectorPriority, msg.Type)
	assert.Equal(t, "priority", msg.Filter.Field)
}

func TestFilterCancelledMsg(t *testing.T) {
	msg := FilterCancelledMsg{}
	assert.NotNil(t, msg)
}

func TestEpicOption(t *testing.T) {
	epic := EpicOption{
		ID:    "epic-123",
		Title: "Q1 Launch",
		Color: "blue",
	}

	assert.Equal(t, "epic-123", epic.ID)
	assert.Equal(t, "Q1 Launch", epic.Title)
	assert.Equal(t, "blue", epic.Color)
}
