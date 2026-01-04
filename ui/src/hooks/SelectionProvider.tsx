import { useState, useCallback, type ReactNode } from 'react'
import {
  SelectionContext,
  type SelectionContextValue,
} from './selectionContext'

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
