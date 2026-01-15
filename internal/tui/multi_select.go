package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// MultiSelect manages multi-selection state for tasks.
type MultiSelect struct {
	selected map[string]bool // taskID -> selected
	order    []string        // selection order (for ordered operations)
}

// NewMultiSelect creates a new multi-select state.
func NewMultiSelect() *MultiSelect {
	return &MultiSelect{
		selected: make(map[string]bool),
		order:    make([]string, 0),
	}
}

// Toggle adds or removes a task from selection.
func (m *MultiSelect) Toggle(taskID string) {
	if m.selected[taskID] {
		delete(m.selected, taskID)
		// Remove from order
		for i, id := range m.order {
			if id == taskID {
				m.order = append(m.order[:i], m.order[i+1:]...)
				break
			}
		}
	} else {
		m.selected[taskID] = true
		m.order = append(m.order, taskID)
	}
}

// Select adds a task to selection (without toggling).
func (m *MultiSelect) Select(taskID string) {
	if !m.selected[taskID] {
		m.selected[taskID] = true
		m.order = append(m.order, taskID)
	}
}

// Deselect removes a task from selection.
func (m *MultiSelect) Deselect(taskID string) {
	if m.selected[taskID] {
		delete(m.selected, taskID)
		for i, id := range m.order {
			if id == taskID {
				m.order = append(m.order[:i], m.order[i+1:]...)
				break
			}
		}
	}
}

// IsSelected returns true if task is selected.
func (m *MultiSelect) IsSelected(taskID string) bool {
	return m.selected[taskID]
}

// Count returns number of selected tasks.
func (m *MultiSelect) Count() int {
	return len(m.selected)
}

// HasSelection returns true if any tasks are selected.
func (m *MultiSelect) HasSelection() bool {
	return len(m.selected) > 0
}

// Clear removes all selections.
func (m *MultiSelect) Clear() {
	m.selected = make(map[string]bool)
	m.order = make([]string, 0)
}

// GetSelected returns all selected task IDs in selection order.
func (m *MultiSelect) GetSelected() []string {
	return m.order
}

// GetSelectedMap returns the selection map (for fast lookup).
func (m *MultiSelect) GetSelectedMap() map[string]bool {
	return m.selected
}

// SelectAll selects all tasks in the provided list.
func (m *MultiSelect) SelectAll(taskIDs []string) {
	for _, id := range taskIDs {
		m.Select(id)
	}
}

// SelectAllInColumn selects all tasks in a column.
func (m *MultiSelect) SelectAllInColumn(tasks []TaskItem) {
	for _, t := range tasks {
		m.Select(t.ID)
	}
}

// RenderSelectionIndicator renders the checkbox for a task.
func RenderSelectionIndicator(selected bool) string {
	if selected {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true).
			Render("[x]")
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("[ ]")
}

// RenderSelectionCount renders the selection count for status bar.
func RenderSelectionCount(count int) string {
	if count == 0 {
		return ""
	}

	style := lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("255")).
		Bold(true).
		Padding(0, 1)

	text := fmt.Sprintf("%d selected", count)
	if count == 1 {
		text = "1 selected"
	}

	return style.Render(text)
}

// SelectionMode represents the current selection mode.
type SelectionMode int

const (
	SelectionModeNone   SelectionMode = iota // Normal navigation
	SelectionModeActive                      // Multi-select active
)

// SelectionState tracks complete selection state.
type SelectionState struct {
	Mode        SelectionMode
	MultiSelect *MultiSelect
}

// NewSelectionState creates a new selection state.
func NewSelectionState() *SelectionState {
	return &SelectionState{
		Mode:        SelectionModeNone,
		MultiSelect: NewMultiSelect(),
	}
}

// EnterSelectionMode activates multi-select mode.
func (s *SelectionState) EnterSelectionMode() {
	s.Mode = SelectionModeActive
}

// ExitSelectionMode clears selection and returns to normal mode.
func (s *SelectionState) ExitSelectionMode() {
	s.Mode = SelectionModeNone
	s.MultiSelect.Clear()
}

// ToggleTask toggles selection, entering select mode if needed.
func (s *SelectionState) ToggleTask(taskID string) {
	if s.Mode == SelectionModeNone {
		s.Mode = SelectionModeActive
	}
	s.MultiSelect.Toggle(taskID)

	// Exit mode if nothing selected
	if s.MultiSelect.Count() == 0 {
		s.Mode = SelectionModeNone
	}
}

// IsActive returns true if selection mode is active.
func (s *SelectionState) IsActive() bool {
	return s.Mode == SelectionModeActive
}

// Count returns the number of selected items.
func (s *SelectionState) Count() int {
	return s.MultiSelect.Count()
}

// HasSelection returns true if any items are selected.
func (s *SelectionState) HasSelection() bool {
	return s.MultiSelect.HasSelection()
}

// IsSelected returns true if a task is selected.
func (s *SelectionState) IsSelected(taskID string) bool {
	return s.MultiSelect.IsSelected(taskID)
}

// GetSelected returns all selected task IDs.
func (s *SelectionState) GetSelected() []string {
	return s.MultiSelect.GetSelected()
}

// SelectAllInColumn selects all tasks in a column.
func (s *SelectionState) SelectAllInColumn(tasks []TaskItem) {
	if s.Mode == SelectionModeNone {
		s.Mode = SelectionModeActive
	}
	s.MultiSelect.SelectAllInColumn(tasks)
}

// Clear clears all selections without exiting selection mode.
func (s *SelectionState) Clear() {
	s.MultiSelect.Clear()
	s.Mode = SelectionModeNone
}
