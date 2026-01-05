import { createContext } from 'react'

export interface SelectionState {
  // Currently focused/selected task ID (single selection)
  selectedTaskId: string | null

  // Multi-selected task IDs (for bulk operations)
  multiSelectedIds: Set<string>

  // Currently focused column (for keyboard navigation)
  focusedColumn: string | null
}

export interface SelectionActions {
  // Select a single task (clears multi-selection)
  selectTask: (taskId: string | null) => void

  // Toggle a task in multi-selection
  toggleMultiSelect: (taskId: string) => void

  // Select a range of tasks (Shift+click behavior)
  selectRange: (fromId: string, toId: string, allTaskIds: string[]) => void

  // Select all visible tasks
  selectAll: (taskIds: string[]) => void

  // Clear all selection
  clearSelection: () => void

  // Set focused column (for H/L navigation)
  setFocusedColumn: (column: string | null) => void

  // Check if a task is selected (single or multi)
  isSelected: (taskId: string) => boolean
}

export interface SelectionContextValue extends SelectionState, SelectionActions {}

export const SelectionContext = createContext<SelectionContextValue | null>(null)
