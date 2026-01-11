package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/stretchr/testify/assert"
)

func TestNewColumn(t *testing.T) {
	items := []list.Item{
		TaskItem{ID: "1", TaskTitle: "Task 1", DisplayID: "WRK-1"},
		TaskItem{ID: "2", TaskTitle: "Task 2", DisplayID: "WRK-2"},
	}

	col := NewColumn("todo", items, true)

	assert.Equal(t, "todo", col.Status())
	assert.True(t, col.IsFocused())
	assert.Len(t, col.Items(), 2)
}

func TestColumn_SetFocused(t *testing.T) {
	col := NewColumn("backlog", nil, false)

	assert.False(t, col.IsFocused())

	col.SetFocused(true)
	assert.True(t, col.IsFocused())

	col.SetFocused(false)
	assert.False(t, col.IsFocused())
}

func TestColumn_SelectedTask(t *testing.T) {
	// Empty column
	col := NewColumn("todo", nil, true)
	assert.Nil(t, col.SelectedTask())

	// Column with tasks
	items := []list.Item{
		TaskItem{ID: "1", TaskTitle: "Task 1", DisplayID: "WRK-1"},
	}
	col = NewColumn("todo", items, true)

	task := col.SelectedTask()
	assert.NotNil(t, task)
	assert.Equal(t, "1", task.ID)
}

func TestColumn_View(t *testing.T) {
	items := []list.Item{
		TaskItem{ID: "1", TaskTitle: "Task 1", DisplayID: "WRK-1"},
	}
	col := NewColumn("todo", items, true)
	col.SetSize(30, 20)

	view := col.View()

	// Should contain column title with count
	assert.Contains(t, view, "Todo")
	assert.Contains(t, view, "(1)")
}

func TestColumn_EmptyView(t *testing.T) {
	col := NewColumn("review", nil, false)
	col.SetSize(30, 20)

	view := col.View()

	assert.Contains(t, view, "Review")
	assert.Contains(t, view, "(0)")
	assert.Contains(t, view, "(empty)")
}
