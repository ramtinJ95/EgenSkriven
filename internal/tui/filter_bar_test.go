package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestFilterBar_New(t *testing.T) {
	fs := NewFilterState()
	fb := NewFilterBar(fs)

	assert.NotNil(t, fb.filterState)
	assert.False(t, fb.isSearching)
	assert.Equal(t, -1, fb.selectedChip)
	assert.False(t, fb.focused)
}

func TestFilterBar_View_Empty(t *testing.T) {
	fs := NewFilterState()
	fb := NewFilterBar(fs)
	fb.SetWidth(80)

	// Empty filter state should render nothing
	view := fb.View()
	assert.Empty(t, view)
}

func TestFilterBar_View_WithSearchQuery(t *testing.T) {
	fs := NewFilterState()
	fs.SetSearchQuery("test query")
	fb := NewFilterBar(fs)
	fb.SetWidth(80)

	view := fb.View()
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "test query")
	assert.Contains(t, view, "fc to clear")
}

func TestFilterBar_View_WithFilters(t *testing.T) {
	fs := NewFilterState()
	fs.AddFilter(Filter{Field: "priority", Value: "high", Display: "Priority: High"})
	fb := NewFilterBar(fs)
	fb.SetWidth(80)

	view := fb.View()
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Priority: High")
}

func TestFilterBar_StartSearch(t *testing.T) {
	fs := NewFilterState()
	fb := NewFilterBar(fs)

	assert.False(t, fb.IsSearching())

	fb.StartSearch()

	assert.True(t, fb.IsSearching())
}

func TestFilterBar_StopSearch(t *testing.T) {
	fs := NewFilterState()
	fb := NewFilterBar(fs)

	fb.StartSearch()
	assert.True(t, fb.IsSearching())

	fb.StopSearch()
	assert.False(t, fb.IsSearching())
}

func TestFilterBar_SetWidth(t *testing.T) {
	fs := NewFilterState()
	fb := NewFilterBar(fs)

	fb.SetWidth(100)
	assert.Equal(t, 100, fb.width)
}

func TestFilterBar_Focus_Blur(t *testing.T) {
	fs := NewFilterState()
	fb := NewFilterBar(fs)

	fb.Focus()
	assert.True(t, fb.focused)

	// Set selected chip
	fs.AddFilter(Filter{Field: "priority", Value: "high"})
	fb.selectedChip = 0

	fb.Blur()
	assert.False(t, fb.focused)
	assert.Equal(t, -1, fb.selectedChip)
}

func TestFilterBar_ChipNavigation(t *testing.T) {
	fs := NewFilterState()
	fs.SetSearchQuery("test")
	fs.AddFilter(Filter{Field: "priority", Value: "high"})
	fs.AddFilter(Filter{Field: "type", Value: "bug"})
	fb := NewFilterBar(fs)
	fb.Focus()

	// Initial state
	assert.Equal(t, -1, fb.selectedChip)

	// Navigate right
	fb, _ = fb.handleChipNavigation(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	assert.Equal(t, 0, fb.selectedChip)

	fb, _ = fb.handleChipNavigation(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	assert.Equal(t, 1, fb.selectedChip)

	fb, _ = fb.handleChipNavigation(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	assert.Equal(t, 2, fb.selectedChip)

	// Wrap around
	fb, _ = fb.handleChipNavigation(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	assert.Equal(t, 0, fb.selectedChip)

	// Navigate left
	fb, _ = fb.handleChipNavigation(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	assert.Equal(t, 2, fb.selectedChip)
}

func TestFilterBar_RemoveChip_SearchQuery(t *testing.T) {
	fs := NewFilterState()
	fs.SetSearchQuery("test")
	fb := NewFilterBar(fs)
	fb.selectedChip = 0

	assert.NotEmpty(t, fs.GetSearchQuery())

	fb, _ = fb.removeSelectedChip()

	assert.Empty(t, fs.GetSearchQuery())
}

func TestFilterBar_RemoveChip_Filter(t *testing.T) {
	fs := NewFilterState()
	fs.AddFilter(Filter{Field: "priority", Value: "high"})
	fs.AddFilter(Filter{Field: "type", Value: "bug"})
	fb := NewFilterBar(fs)
	fb.selectedChip = 0

	assert.Len(t, fs.GetFilters(), 2)

	fb, _ = fb.removeSelectedChip()

	assert.Len(t, fs.GetFilters(), 1)
	assert.Equal(t, "type", fs.GetFilters()[0].Field)
}

func TestFilterBar_RemoveChip_WithSearchAndFilter(t *testing.T) {
	fs := NewFilterState()
	fs.SetSearchQuery("test")
	fs.AddFilter(Filter{Field: "priority", Value: "high"})
	fb := NewFilterBar(fs)

	// Select the filter (index 1 because search is index 0)
	fb.selectedChip = 1

	fb, _ = fb.removeSelectedChip()

	// Search should remain, filter should be gone
	assert.NotEmpty(t, fs.GetSearchQuery())
	assert.Empty(t, fs.GetFilters())
}

func TestFilterBar_HandleSearchInput_Enter(t *testing.T) {
	fs := NewFilterState()
	fb := NewFilterBar(fs)
	fb.StartSearch()
	fb.searchInput.SetValue("new search")

	fb, _ = fb.handleSearchInput(tea.KeyMsg{Type: tea.KeyEnter})

	assert.False(t, fb.IsSearching())
	assert.Equal(t, "new search", fs.GetSearchQuery())
}

func TestFilterBar_HandleSearchInput_Escape(t *testing.T) {
	fs := NewFilterState()
	fs.SetSearchQuery("original")
	fb := NewFilterBar(fs)
	fb.StartSearch()
	fb.searchInput.SetValue("modified")

	fb, _ = fb.handleSearchInput(tea.KeyMsg{Type: tea.KeyEsc})

	assert.False(t, fb.IsSearching())
	// Should restore original value
	assert.Equal(t, "original", fs.GetSearchQuery())
}

func TestFilterBar_PriorityColor(t *testing.T) {
	fb := FilterBar{}

	assert.Equal(t, "196", fb.priorityColor("urgent"))
	assert.Equal(t, "208", fb.priorityColor("high"))
	assert.Equal(t, "226", fb.priorityColor("medium"))
	assert.Equal(t, "240", fb.priorityColor("low"))
	assert.Equal(t, "62", fb.priorityColor("unknown"))
}

func TestFilterBar_TypeColor(t *testing.T) {
	fb := FilterBar{}

	assert.Equal(t, "196", fb.typeColor("bug"))
	assert.Equal(t, "39", fb.typeColor("feature"))
	assert.Equal(t, "240", fb.typeColor("chore"))
	assert.Equal(t, "62", fb.typeColor("unknown"))
}

func TestFilterBar_ChipNavigation_Empty(t *testing.T) {
	fs := NewFilterState()
	fb := NewFilterBar(fs)

	// With no chips, navigation should do nothing
	fb, _ = fb.handleChipNavigation(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	assert.Equal(t, -1, fb.selectedChip)

	fb, _ = fb.handleChipNavigation(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	assert.Equal(t, -1, fb.selectedChip)
}

func TestFilterBar_RemoveChip_AdjustsSelection(t *testing.T) {
	fs := NewFilterState()
	fs.AddFilter(Filter{Field: "priority", Value: "high"})
	fs.AddFilter(Filter{Field: "type", Value: "bug"})
	fb := NewFilterBar(fs)

	// Select the last chip
	fb.selectedChip = 1

	fb, _ = fb.removeSelectedChip()

	// Selection should adjust to the last remaining chip
	assert.Equal(t, 0, fb.selectedChip)
}
