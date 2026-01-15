package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestSearchOverlay_New(t *testing.T) {
	fs := NewFilterState()
	so := NewSearchOverlay(fs)

	assert.NotNil(t, so.filterState)
	assert.False(t, so.active)
	assert.Equal(t, "/ ", so.input.Prompt)
}

func TestSearchOverlay_Show(t *testing.T) {
	fs := NewFilterState()
	fs.SetSearchQuery("existing query")
	so := NewSearchOverlay(fs)

	assert.False(t, so.IsActive())

	so.Show()

	assert.True(t, so.IsActive())
	// Should preserve existing query
	assert.Equal(t, "existing query", so.input.Value())
}

func TestSearchOverlay_Hide(t *testing.T) {
	fs := NewFilterState()
	so := NewSearchOverlay(fs)

	so.Show()
	assert.True(t, so.IsActive())

	so.Hide()
	assert.False(t, so.IsActive())
}

func TestSearchOverlay_SetSize(t *testing.T) {
	fs := NewFilterState()
	so := NewSearchOverlay(fs)

	so.SetSize(100, 50)

	assert.Equal(t, 100, so.width)
	assert.Equal(t, 50, so.height)
}

func TestSearchOverlay_View_Inactive(t *testing.T) {
	fs := NewFilterState()
	so := NewSearchOverlay(fs)
	so.SetSize(80, 24)

	// Inactive overlay should render nothing
	view := so.View()
	assert.Empty(t, view)
}

func TestSearchOverlay_View_Active(t *testing.T) {
	fs := NewFilterState()
	so := NewSearchOverlay(fs)
	so.SetSize(80, 24)
	so.Show()

	view := so.View()
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Search Tasks")
	assert.Contains(t, view, "Enter to search")
	assert.Contains(t, view, "Esc to cancel")
}

func TestSearchOverlay_Update_Enter(t *testing.T) {
	fs := NewFilterState()
	so := NewSearchOverlay(fs)
	so.Show()
	so.input.SetValue("new search")

	so, cmd := so.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.False(t, so.IsActive())
	assert.Equal(t, "new search", fs.GetSearchQuery())
	assert.NotNil(t, cmd)

	// Execute the command and check the message
	msg := cmd()
	appliedMsg, ok := msg.(SearchAppliedMsg)
	assert.True(t, ok)
	assert.Equal(t, "new search", appliedMsg.Query)
}

func TestSearchOverlay_Update_Escape(t *testing.T) {
	fs := NewFilterState()
	fs.SetSearchQuery("original")
	so := NewSearchOverlay(fs)
	so.Show()
	so.input.SetValue("modified")

	so, cmd := so.Update(tea.KeyMsg{Type: tea.KeyEsc})

	assert.False(t, so.IsActive())
	// Original query should remain unchanged
	assert.Equal(t, "original", fs.GetSearchQuery())
	assert.NotNil(t, cmd)

	// Execute the command and check the message
	msg := cmd()
	_, ok := msg.(SearchCancelledMsg)
	assert.True(t, ok)
}

func TestSearchOverlay_Update_CtrlU(t *testing.T) {
	fs := NewFilterState()
	so := NewSearchOverlay(fs)
	so.Show()
	so.input.SetValue("some text")

	so, _ = so.Update(tea.KeyMsg{Type: tea.KeyCtrlU})

	// Input should be cleared but overlay still active
	assert.True(t, so.IsActive())
	assert.Empty(t, so.input.Value())
}

func TestSearchOverlay_Update_Inactive(t *testing.T) {
	fs := NewFilterState()
	so := NewSearchOverlay(fs)
	// Don't call Show()

	so, cmd := so.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should not process when inactive
	assert.False(t, so.IsActive())
	assert.Nil(t, cmd)
}

func TestSearchOverlay_Update_TextInput(t *testing.T) {
	fs := NewFilterState()
	so := NewSearchOverlay(fs)
	so.Show()

	// Type a character
	so, _ = so.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	// Input should have the character
	assert.Contains(t, so.input.Value(), "a")
}

func TestSearchAppliedMsg(t *testing.T) {
	msg := SearchAppliedMsg{Query: "test query"}
	assert.Equal(t, "test query", msg.Query)
}

func TestSearchCancelledMsg(t *testing.T) {
	msg := SearchCancelledMsg{}
	assert.NotNil(t, msg)
}
