package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelpOverlay_Toggle(t *testing.T) {
	h := NewHelpOverlay()

	// Initially not visible
	assert.False(t, h.IsVisible())

	// Toggle on
	h.Toggle()
	assert.True(t, h.IsVisible())

	// Toggle off
	h.Toggle()
	assert.False(t, h.IsVisible())
}

func TestHelpOverlay_ShowHide(t *testing.T) {
	h := NewHelpOverlay()

	// Show
	h.Show()
	assert.True(t, h.IsVisible())

	// Hide
	h.Hide()
	assert.False(t, h.IsVisible())
}

func TestHelpOverlay_SetSize(t *testing.T) {
	h := NewHelpOverlay()
	h.SetSize(100, 50)

	assert.Equal(t, 100, h.width)
	assert.Equal(t, 50, h.height)
}

func TestHelpOverlay_ViewWhenHidden(t *testing.T) {
	h := NewHelpOverlay()

	// When hidden, View should return empty string
	assert.Empty(t, h.View())
}

func TestHelpOverlay_ViewWhenVisible(t *testing.T) {
	h := NewHelpOverlay()
	h.SetSize(120, 40)
	h.Show()

	view := h.View()

	// Should contain title
	assert.Contains(t, view, "Keyboard Shortcuts")

	// Should contain some key bindings
	assert.Contains(t, view, "j/↓")
	assert.Contains(t, view, "Move down")
	assert.Contains(t, view, "Navigation")
	assert.Contains(t, view, "Filtering")

	// Should contain close hint
	assert.Contains(t, view, "Press ? to close")
}

func TestGetHelpSections(t *testing.T) {
	sections := GetHelpSections()

	// Should have multiple sections
	assert.GreaterOrEqual(t, len(sections), 4)

	// Check section titles exist
	titles := make(map[string]bool)
	for _, s := range sections {
		titles[s.Title] = true
	}

	assert.True(t, titles["Navigation"])
	assert.True(t, titles["Task Actions"])
	assert.True(t, titles["Filtering"])
	assert.True(t, titles["Global"])

	// Each section should have bindings
	for _, s := range sections {
		assert.NotEmpty(t, s.Bindings, "section %s should have bindings", s.Title)
	}
}

func TestHelpBinding(t *testing.T) {
	binding := HelpBinding{
		Key:         "j/↓",
		Description: "Move down",
	}

	assert.Equal(t, "j/↓", binding.Key)
	assert.Equal(t, "Move down", binding.Description)
}
