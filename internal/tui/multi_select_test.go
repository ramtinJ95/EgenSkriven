package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMultiSelect_Toggle(t *testing.T) {
	ms := NewMultiSelect()

	// Initially empty
	assert.False(t, ms.IsSelected("task-1"))
	assert.Equal(t, 0, ms.Count())

	// Toggle on
	ms.Toggle("task-1")
	assert.True(t, ms.IsSelected("task-1"))
	assert.Equal(t, 1, ms.Count())

	// Toggle off
	ms.Toggle("task-1")
	assert.False(t, ms.IsSelected("task-1"))
	assert.Equal(t, 0, ms.Count())
}

func TestMultiSelect_SelectDeselect(t *testing.T) {
	ms := NewMultiSelect()

	// Select
	ms.Select("task-1")
	ms.Select("task-2")
	assert.Equal(t, 2, ms.Count())

	// Select same task again (no change)
	ms.Select("task-1")
	assert.Equal(t, 2, ms.Count())

	// Deselect
	ms.Deselect("task-1")
	assert.Equal(t, 1, ms.Count())
	assert.False(t, ms.IsSelected("task-1"))
	assert.True(t, ms.IsSelected("task-2"))
}

func TestMultiSelect_Clear(t *testing.T) {
	ms := NewMultiSelect()
	ms.Select("task-1")
	ms.Select("task-2")
	ms.Select("task-3")

	assert.Equal(t, 3, ms.Count())
	assert.True(t, ms.HasSelection())

	ms.Clear()

	assert.Equal(t, 0, ms.Count())
	assert.False(t, ms.HasSelection())
}

func TestMultiSelect_GetSelected(t *testing.T) {
	ms := NewMultiSelect()
	ms.Select("task-1")
	ms.Select("task-2")
	ms.Select("task-3")

	selected := ms.GetSelected()
	assert.Len(t, selected, 3)
	// Order should be preserved
	assert.Equal(t, []string{"task-1", "task-2", "task-3"}, selected)
}

func TestMultiSelect_SelectAllInColumn(t *testing.T) {
	ms := NewMultiSelect()
	tasks := []TaskItem{
		{ID: "task-1", TaskTitle: "Task 1"},
		{ID: "task-2", TaskTitle: "Task 2"},
		{ID: "task-3", TaskTitle: "Task 3"},
	}

	ms.SelectAllInColumn(tasks)

	assert.Equal(t, 3, ms.Count())
	assert.True(t, ms.IsSelected("task-1"))
	assert.True(t, ms.IsSelected("task-2"))
	assert.True(t, ms.IsSelected("task-3"))
}

func TestSelectionState_ToggleTask(t *testing.T) {
	ss := NewSelectionState()

	// Initially not active
	assert.False(t, ss.IsActive())

	// Toggle first task - enters selection mode
	ss.ToggleTask("task-1")
	assert.True(t, ss.IsActive())
	assert.Equal(t, 1, ss.Count())

	// Toggle second task
	ss.ToggleTask("task-2")
	assert.True(t, ss.IsActive())
	assert.Equal(t, 2, ss.Count())

	// Toggle first task off
	ss.ToggleTask("task-1")
	assert.True(t, ss.IsActive())
	assert.Equal(t, 1, ss.Count())

	// Toggle last task off - exits selection mode
	ss.ToggleTask("task-2")
	assert.False(t, ss.IsActive())
	assert.Equal(t, 0, ss.Count())
}

func TestSelectionState_ExitSelectionMode(t *testing.T) {
	ss := NewSelectionState()
	ss.ToggleTask("task-1")
	ss.ToggleTask("task-2")

	assert.True(t, ss.IsActive())
	assert.Equal(t, 2, ss.Count())

	ss.ExitSelectionMode()

	assert.False(t, ss.IsActive())
	assert.Equal(t, 0, ss.Count())
}

func TestSelectionState_SelectAllInColumn(t *testing.T) {
	ss := NewSelectionState()
	tasks := []TaskItem{
		{ID: "task-1", TaskTitle: "Task 1"},
		{ID: "task-2", TaskTitle: "Task 2"},
	}

	// Should enter selection mode
	assert.False(t, ss.IsActive())
	ss.SelectAllInColumn(tasks)
	assert.True(t, ss.IsActive())
	assert.Equal(t, 2, ss.Count())
}

func TestSelectionState_IsSelected(t *testing.T) {
	ss := NewSelectionState()

	assert.False(t, ss.IsSelected("task-1"))

	ss.ToggleTask("task-1")
	assert.True(t, ss.IsSelected("task-1"))
	assert.False(t, ss.IsSelected("task-2"))
}

func TestSelectionState_GetSelected(t *testing.T) {
	ss := NewSelectionState()
	ss.ToggleTask("task-1")
	ss.ToggleTask("task-2")

	selected := ss.GetSelected()
	assert.Len(t, selected, 2)
	assert.Contains(t, selected, "task-1")
	assert.Contains(t, selected, "task-2")
}

func TestSelectionState_Clear(t *testing.T) {
	ss := NewSelectionState()
	ss.ToggleTask("task-1")
	ss.ToggleTask("task-2")

	assert.True(t, ss.IsActive())
	assert.Equal(t, 2, ss.Count())

	ss.Clear()

	assert.False(t, ss.IsActive())
	assert.Equal(t, 0, ss.Count())
}

func TestRenderSelectionIndicator(t *testing.T) {
	selected := RenderSelectionIndicator(true)
	unselected := RenderSelectionIndicator(false)

	assert.Contains(t, selected, "x")
	assert.Contains(t, unselected, " ")
}

func TestRenderSelectionCount(t *testing.T) {
	// Zero count returns empty
	assert.Empty(t, RenderSelectionCount(0))

	// Single item
	single := RenderSelectionCount(1)
	assert.Contains(t, single, "1 selected")

	// Multiple items
	multiple := RenderSelectionCount(5)
	assert.Contains(t, multiple, "5 selected")
}
