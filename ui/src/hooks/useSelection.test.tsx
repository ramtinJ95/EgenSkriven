import { describe, it, expect } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import type { ReactNode } from 'react'
import { SelectionProvider, useSelection } from './useSelection'

// Wrapper for renderHook that provides the SelectionProvider
const wrapper = ({ children }: { children: ReactNode }) => (
  <SelectionProvider>{children}</SelectionProvider>
)

describe('useSelection', () => {
  it('starts with no selection', () => {
    const { result } = renderHook(() => useSelection(), { wrapper })

    expect(result.current.selectedTaskId).toBeNull()
    expect(result.current.multiSelectedIds.size).toBe(0)
    expect(result.current.focusedColumn).toBeNull()
  })

  it('throws error when used outside SelectionProvider', () => {
    // Suppress console.error for this test since we expect an error
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

    expect(() => {
      renderHook(() => useSelection())
    }).toThrow('useSelection must be used within a SelectionProvider')

    consoleSpy.mockRestore()
  })

  describe('selectTask', () => {
    it('selects a single task', () => {
      const { result } = renderHook(() => useSelection(), { wrapper })

      act(() => {
        result.current.selectTask('task-1')
      })

      expect(result.current.selectedTaskId).toBe('task-1')
    })

    it('clears previous selection when selecting new task', () => {
      const { result } = renderHook(() => useSelection(), { wrapper })

      act(() => {
        result.current.selectTask('task-1')
      })
      act(() => {
        result.current.selectTask('task-2')
      })

      expect(result.current.selectedTaskId).toBe('task-2')
    })

    it('clears multi-selection when selecting a single task', () => {
      const { result } = renderHook(() => useSelection(), { wrapper })

      // First, add some multi-selected items
      act(() => {
        result.current.toggleMultiSelect('task-1')
        result.current.toggleMultiSelect('task-2')
      })

      expect(result.current.multiSelectedIds.size).toBe(2)

      // Now select a single task
      act(() => {
        result.current.selectTask('task-3')
      })

      expect(result.current.selectedTaskId).toBe('task-3')
      expect(result.current.multiSelectedIds.size).toBe(0)
    })

    it('can deselect by passing null', () => {
      const { result } = renderHook(() => useSelection(), { wrapper })

      act(() => {
        result.current.selectTask('task-1')
      })
      act(() => {
        result.current.selectTask(null)
      })

      expect(result.current.selectedTaskId).toBeNull()
    })
  })

  describe('toggleMultiSelect', () => {
    it('adds a task to multi-selection', () => {
      const { result } = renderHook(() => useSelection(), { wrapper })

      act(() => {
        result.current.toggleMultiSelect('task-1')
      })

      expect(result.current.multiSelectedIds.has('task-1')).toBe(true)
    })

    it('removes a task from multi-selection when already selected', () => {
      const { result } = renderHook(() => useSelection(), { wrapper })

      act(() => {
        result.current.toggleMultiSelect('task-1')
      })
      act(() => {
        result.current.toggleMultiSelect('task-1')
      })

      expect(result.current.multiSelectedIds.has('task-1')).toBe(false)
    })

    it('can select multiple tasks', () => {
      const { result } = renderHook(() => useSelection(), { wrapper })

      act(() => {
        result.current.toggleMultiSelect('task-1')
        result.current.toggleMultiSelect('task-2')
        result.current.toggleMultiSelect('task-3')
      })

      expect(result.current.multiSelectedIds.size).toBe(3)
      expect(result.current.multiSelectedIds.has('task-1')).toBe(true)
      expect(result.current.multiSelectedIds.has('task-2')).toBe(true)
      expect(result.current.multiSelectedIds.has('task-3')).toBe(true)
    })
  })

  describe('selectRange', () => {
    const allTaskIds = ['task-1', 'task-2', 'task-3', 'task-4', 'task-5']

    it('selects a range of tasks from start to end', () => {
      const { result } = renderHook(() => useSelection(), { wrapper })

      act(() => {
        result.current.selectRange('task-2', 'task-4', allTaskIds)
      })

      expect(result.current.multiSelectedIds.size).toBe(3)
      expect(result.current.multiSelectedIds.has('task-2')).toBe(true)
      expect(result.current.multiSelectedIds.has('task-3')).toBe(true)
      expect(result.current.multiSelectedIds.has('task-4')).toBe(true)
    })

    it('selects a range in reverse order', () => {
      const { result } = renderHook(() => useSelection(), { wrapper })

      act(() => {
        result.current.selectRange('task-4', 'task-2', allTaskIds)
      })

      expect(result.current.multiSelectedIds.size).toBe(3)
      expect(result.current.multiSelectedIds.has('task-2')).toBe(true)
      expect(result.current.multiSelectedIds.has('task-3')).toBe(true)
      expect(result.current.multiSelectedIds.has('task-4')).toBe(true)
    })

    it('does nothing if fromId is not in allTaskIds', () => {
      const { result } = renderHook(() => useSelection(), { wrapper })

      act(() => {
        result.current.selectRange('nonexistent', 'task-4', allTaskIds)
      })

      expect(result.current.multiSelectedIds.size).toBe(0)
    })

    it('does nothing if toId is not in allTaskIds', () => {
      const { result } = renderHook(() => useSelection(), { wrapper })

      act(() => {
        result.current.selectRange('task-2', 'nonexistent', allTaskIds)
      })

      expect(result.current.multiSelectedIds.size).toBe(0)
    })

    it('selects a single task if fromId equals toId', () => {
      const { result } = renderHook(() => useSelection(), { wrapper })

      act(() => {
        result.current.selectRange('task-3', 'task-3', allTaskIds)
      })

      expect(result.current.multiSelectedIds.size).toBe(1)
      expect(result.current.multiSelectedIds.has('task-3')).toBe(true)
    })
  })

  describe('selectAll', () => {
    it('selects all provided task IDs', () => {
      const { result } = renderHook(() => useSelection(), { wrapper })

      const taskIds = ['task-1', 'task-2', 'task-3']
      act(() => {
        result.current.selectAll(taskIds)
      })

      expect(result.current.multiSelectedIds.size).toBe(3)
      taskIds.forEach((id) => {
        expect(result.current.multiSelectedIds.has(id)).toBe(true)
      })
    })

    it('replaces previous multi-selection', () => {
      const { result } = renderHook(() => useSelection(), { wrapper })

      act(() => {
        result.current.toggleMultiSelect('task-old')
      })
      act(() => {
        result.current.selectAll(['task-1', 'task-2'])
      })

      expect(result.current.multiSelectedIds.size).toBe(2)
      expect(result.current.multiSelectedIds.has('task-old')).toBe(false)
    })

    it('handles empty array', () => {
      const { result } = renderHook(() => useSelection(), { wrapper })

      act(() => {
        result.current.selectAll([])
      })

      expect(result.current.multiSelectedIds.size).toBe(0)
    })
  })

  describe('clearSelection', () => {
    it('clears single selection', () => {
      const { result } = renderHook(() => useSelection(), { wrapper })

      act(() => {
        result.current.selectTask('task-1')
      })
      act(() => {
        result.current.clearSelection()
      })

      expect(result.current.selectedTaskId).toBeNull()
    })

    it('clears multi-selection', () => {
      const { result } = renderHook(() => useSelection(), { wrapper })

      act(() => {
        result.current.toggleMultiSelect('task-1')
        result.current.toggleMultiSelect('task-2')
      })
      act(() => {
        result.current.clearSelection()
      })

      expect(result.current.multiSelectedIds.size).toBe(0)
    })

    it('clears both single and multi-selection', () => {
      const { result } = renderHook(() => useSelection(), { wrapper })

      act(() => {
        result.current.selectTask('task-single')
        result.current.toggleMultiSelect('task-multi-1')
        result.current.toggleMultiSelect('task-multi-2')
      })
      act(() => {
        result.current.clearSelection()
      })

      expect(result.current.selectedTaskId).toBeNull()
      expect(result.current.multiSelectedIds.size).toBe(0)
    })
  })

  describe('setFocusedColumn', () => {
    it('sets the focused column', () => {
      const { result } = renderHook(() => useSelection(), { wrapper })

      act(() => {
        result.current.setFocusedColumn('todo')
      })

      expect(result.current.focusedColumn).toBe('todo')
    })

    it('can clear focused column with null', () => {
      const { result } = renderHook(() => useSelection(), { wrapper })

      act(() => {
        result.current.setFocusedColumn('todo')
      })
      act(() => {
        result.current.setFocusedColumn(null)
      })

      expect(result.current.focusedColumn).toBeNull()
    })
  })

  describe('isSelected', () => {
    it('returns true for single-selected task', () => {
      const { result } = renderHook(() => useSelection(), { wrapper })

      act(() => {
        result.current.selectTask('task-1')
      })

      expect(result.current.isSelected('task-1')).toBe(true)
      expect(result.current.isSelected('task-2')).toBe(false)
    })

    it('returns true for multi-selected task', () => {
      const { result } = renderHook(() => useSelection(), { wrapper })

      act(() => {
        result.current.toggleMultiSelect('task-1')
        result.current.toggleMultiSelect('task-2')
      })

      expect(result.current.isSelected('task-1')).toBe(true)
      expect(result.current.isSelected('task-2')).toBe(true)
      expect(result.current.isSelected('task-3')).toBe(false)
    })

    it('returns false for unselected task', () => {
      const { result } = renderHook(() => useSelection(), { wrapper })

      expect(result.current.isSelected('any-task')).toBe(false)
    })
  })
})
