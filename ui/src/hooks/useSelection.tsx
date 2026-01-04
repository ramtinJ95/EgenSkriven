import {
  createContext,
  useContext,
  useState,
  useCallback,
  type ReactNode,
} from 'react'

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

interface SelectionContextValue extends SelectionState, SelectionActions {}

const SelectionContext = createContext<SelectionContextValue | null>(null)

interface SelectionProviderProps {
  children: ReactNode
}

export function SelectionProvider({ children }: SelectionProviderProps) {
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null)
  const [multiSelectedIds, setMultiSelectedIds] = useState<Set<string>>(
    new Set()
  )
  const [focusedColumn, setFocusedColumn] = useState<string | null>(null)

  const selectTask = useCallback((taskId: string | null) => {
    setSelectedTaskId(taskId)
    // Single selection clears multi-selection
    setMultiSelectedIds(new Set())
  }, [])

  const toggleMultiSelect = useCallback((taskId: string) => {
    setMultiSelectedIds((prev) => {
      const next = new Set(prev)
      if (next.has(taskId)) {
        next.delete(taskId)
      } else {
        next.add(taskId)
      }
      return next
    })
  }, [])

  const selectRange = useCallback(
    (fromId: string, toId: string, allTaskIds: string[]) => {
      const fromIndex = allTaskIds.indexOf(fromId)
      const toIndex = allTaskIds.indexOf(toId)

      if (fromIndex === -1 || toIndex === -1) return

      const start = Math.min(fromIndex, toIndex)
      const end = Math.max(fromIndex, toIndex)

      const rangeIds = allTaskIds.slice(start, end + 1)
      setMultiSelectedIds(new Set(rangeIds))
    },
    []
  )

  const selectAll = useCallback((taskIds: string[]) => {
    setMultiSelectedIds(new Set(taskIds))
  }, [])

  const clearSelection = useCallback(() => {
    setSelectedTaskId(null)
    setMultiSelectedIds(new Set())
  }, [])

  const isSelected = useCallback(
    (taskId: string) => {
      return selectedTaskId === taskId || multiSelectedIds.has(taskId)
    },
    [selectedTaskId, multiSelectedIds]
  )

  const value: SelectionContextValue = {
    selectedTaskId,
    multiSelectedIds,
    focusedColumn,
    selectTask,
    toggleMultiSelect,
    selectRange,
    selectAll,
    clearSelection,
    setFocusedColumn,
    isSelected,
  }

  return (
    <SelectionContext.Provider value={value}>
      {children}
    </SelectionContext.Provider>
  )
}

/**
 * Hook to access selection state and actions.
 * Must be used within a SelectionProvider.
 *
 * @example
 * function TaskCard({ task }) {
 *   const { isSelected, selectTask } = useSelection();
 *   return (
 *     <div
 *       className={isSelected(task.id) ? 'selected' : ''}
 *       onClick={() => selectTask(task.id)}
 *     >
 *       {task.title}
 *     </div>
 *   );
 * }
 */
export function useSelection(): SelectionContextValue {
  const context = useContext(SelectionContext)
  if (!context) {
    throw new Error('useSelection must be used within a SelectionProvider')
  }
  return context
}
