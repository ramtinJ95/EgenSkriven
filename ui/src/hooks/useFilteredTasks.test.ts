import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { filterHelpers, useFilteredTasks, useSearchDebounce } from './useFilteredTasks'
import { useFilterStore } from '../stores/filters'
import type { Task } from '../types/task'

const {
  matchesFilter,
  matchSelectFilter,
  matchLabelsFilter,
  matchDateFilter,
  matchRelationFilter,
  matchTextFilter,
  matchesSearch,
} = filterHelpers

// Create a mock task factory
function createMockTask(overrides: Partial<Task> = {}): Task {
  return {
    id: 'task-1',
    collectionId: 'tasks',
    collectionName: 'tasks',
    created: '2024-01-01T00:00:00.000Z',
    updated: '2024-01-01T00:00:00.000Z',
    title: 'Test Task',
    description: 'Test description',
    type: 'feature',
    priority: 'medium',
    column: 'todo',
    position: 1000,
    labels: [],
    created_by: 'user',
    ...overrides,
  } as Task
}

describe('matchSelectFilter', () => {
  describe('is operator', () => {
    it('returns true when values match', () => {
      expect(matchSelectFilter('high', 'is', 'high')).toBe(true)
    })

    it('returns false when values do not match', () => {
      expect(matchSelectFilter('high', 'is', 'low')).toBe(false)
    })

    it('returns false when task value is undefined', () => {
      expect(matchSelectFilter(undefined, 'is', 'high')).toBe(false)
    })
  })

  describe('is_not operator', () => {
    it('returns true when values do not match', () => {
      expect(matchSelectFilter('high', 'is_not', 'low')).toBe(true)
    })

    it('returns false when values match', () => {
      expect(matchSelectFilter('high', 'is_not', 'high')).toBe(false)
    })
  })

  describe('is_any_of operator', () => {
    it('returns true when value is in array', () => {
      expect(matchSelectFilter('high', 'is_any_of', ['high', 'urgent'])).toBe(true)
    })

    it('returns false when value is not in array', () => {
      expect(matchSelectFilter('low', 'is_any_of', ['high', 'urgent'])).toBe(false)
    })
  })

  describe('is_set operator', () => {
    it('returns true when value exists', () => {
      expect(matchSelectFilter('high', 'is_set', null)).toBe(true)
    })

    it('returns false when value is undefined', () => {
      expect(matchSelectFilter(undefined, 'is_set', null)).toBe(false)
    })
  })

  describe('is_not_set operator', () => {
    it('returns true when value is undefined', () => {
      expect(matchSelectFilter(undefined, 'is_not_set', null)).toBe(true)
    })

    it('returns false when value exists', () => {
      expect(matchSelectFilter('high', 'is_not_set', null)).toBe(false)
    })
  })
})

describe('matchLabelsFilter', () => {
  describe('includes_any operator', () => {
    it('returns true when task has at least one filter label', () => {
      expect(matchLabelsFilter(['bug', 'frontend'], 'includes_any', ['bug', 'backend'])).toBe(true)
    })

    it('returns false when task has none of the filter labels', () => {
      expect(matchLabelsFilter(['frontend'], 'includes_any', ['bug', 'backend'])).toBe(false)
    })

    it('returns true when filter labels is empty', () => {
      expect(matchLabelsFilter(['bug'], 'includes_any', [])).toBe(true)
    })
  })

  describe('includes_all operator', () => {
    it('returns true when task has all filter labels', () => {
      expect(matchLabelsFilter(['bug', 'frontend', 'urgent'], 'includes_all', ['bug', 'frontend'])).toBe(true)
    })

    it('returns false when task is missing some filter labels', () => {
      expect(matchLabelsFilter(['bug'], 'includes_all', ['bug', 'frontend'])).toBe(false)
    })
  })

  describe('includes_none operator', () => {
    it('returns true when task has none of the filter labels', () => {
      expect(matchLabelsFilter(['docs'], 'includes_none', ['bug', 'frontend'])).toBe(true)
    })

    it('returns false when task has any filter label', () => {
      expect(matchLabelsFilter(['bug', 'docs'], 'includes_none', ['bug', 'frontend'])).toBe(false)
    })
  })

  describe('is_set operator', () => {
    it('returns true when task has labels', () => {
      expect(matchLabelsFilter(['bug'], 'is_set', [])).toBe(true)
    })

    it('returns false when task has no labels', () => {
      expect(matchLabelsFilter([], 'is_set', [])).toBe(false)
    })
  })

  describe('is_not_set operator', () => {
    it('returns true when task has no labels', () => {
      expect(matchLabelsFilter([], 'is_not_set', [])).toBe(true)
    })

    it('returns false when task has labels', () => {
      expect(matchLabelsFilter(['bug'], 'is_not_set', [])).toBe(false)
    })
  })
})

describe('matchDateFilter', () => {
  describe('is operator', () => {
    it('returns true when dates are the same day', () => {
      expect(matchDateFilter('2024-03-15T10:00:00Z', 'is', '2024-03-15')).toBe(true)
    })

    it('returns false when dates are different days', () => {
      expect(matchDateFilter('2024-03-15', 'is', '2024-03-16')).toBe(false)
    })
  })

  describe('before operator', () => {
    it('returns true when task date is before filter date', () => {
      expect(matchDateFilter('2024-03-14', 'before', '2024-03-15')).toBe(true)
    })

    it('returns false when task date is after filter date', () => {
      expect(matchDateFilter('2024-03-16', 'before', '2024-03-15')).toBe(false)
    })

    it('returns false when dates are the same', () => {
      expect(matchDateFilter('2024-03-15', 'before', '2024-03-15')).toBe(false)
    })
  })

  describe('after operator', () => {
    it('returns true when task date is after filter date', () => {
      expect(matchDateFilter('2024-03-16', 'after', '2024-03-15')).toBe(true)
    })

    it('returns false when task date is before filter date', () => {
      expect(matchDateFilter('2024-03-14', 'after', '2024-03-15')).toBe(false)
    })
  })

  describe('is_set operator', () => {
    it('returns true when due_date exists', () => {
      expect(matchDateFilter('2024-03-15', 'is_set', null)).toBe(true)
    })

    it('returns false when due_date is undefined', () => {
      expect(matchDateFilter(undefined, 'is_set', null)).toBe(false)
    })
  })

  describe('is_not_set operator', () => {
    it('returns true when due_date is undefined', () => {
      expect(matchDateFilter(undefined, 'is_not_set', null)).toBe(true)
    })

    it('returns false when due_date exists', () => {
      expect(matchDateFilter('2024-03-15', 'is_not_set', null)).toBe(false)
    })
  })
})

describe('matchRelationFilter', () => {
  describe('is operator', () => {
    it('returns true when relation matches', () => {
      expect(matchRelationFilter('epic-1', 'is', 'epic-1')).toBe(true)
    })

    it('returns false when relation does not match', () => {
      expect(matchRelationFilter('epic-1', 'is', 'epic-2')).toBe(false)
    })
  })

  describe('is_not operator', () => {
    it('returns true when relation does not match', () => {
      expect(matchRelationFilter('epic-1', 'is_not', 'epic-2')).toBe(true)
    })

    it('returns false when relation matches', () => {
      expect(matchRelationFilter('epic-1', 'is_not', 'epic-1')).toBe(false)
    })
  })

  describe('is_set operator', () => {
    it('returns true when relation exists', () => {
      expect(matchRelationFilter('epic-1', 'is_set', null)).toBe(true)
    })

    it('returns false when relation is undefined', () => {
      expect(matchRelationFilter(undefined, 'is_set', null)).toBe(false)
    })
  })

  describe('is_not_set operator', () => {
    it('returns true when relation is undefined', () => {
      expect(matchRelationFilter(undefined, 'is_not_set', null)).toBe(true)
    })

    it('returns false when relation exists', () => {
      expect(matchRelationFilter('epic-1', 'is_not_set', null)).toBe(false)
    })
  })
})

describe('matchTextFilter', () => {
  describe('contains operator', () => {
    it('returns true when text contains filter value (case-insensitive)', () => {
      expect(matchTextFilter('Fix Login Bug', 'contains', 'login')).toBe(true)
    })

    it('returns false when text does not contain filter value', () => {
      expect(matchTextFilter('Fix Login Bug', 'contains', 'signup')).toBe(false)
    })

    it('returns true when filter value is empty', () => {
      expect(matchTextFilter('Fix Login Bug', 'contains', '')).toBe(true)
    })
  })

  describe('is operator', () => {
    it('returns true when text exactly matches (case-insensitive)', () => {
      expect(matchTextFilter('Fix Bug', 'is', 'fix bug')).toBe(true)
    })

    it('returns false when text does not exactly match', () => {
      expect(matchTextFilter('Fix Bug', 'is', 'fix')).toBe(false)
    })
  })

  describe('is_not operator', () => {
    it('returns true when text does not match', () => {
      expect(matchTextFilter('Fix Bug', 'is_not', 'Add Feature')).toBe(true)
    })

    it('returns false when text matches', () => {
      expect(matchTextFilter('Fix Bug', 'is_not', 'fix bug')).toBe(false)
    })
  })
})

describe('matchesSearch', () => {
  it('returns true for empty query', () => {
    const task = createMockTask({ title: 'Test' })
    expect(matchesSearch(task, '')).toBe(true)
    expect(matchesSearch(task, '   ')).toBe(true)
  })

  it('matches title (case-insensitive)', () => {
    const task = createMockTask({ title: 'Fix Login Bug' })
    expect(matchesSearch(task, 'login')).toBe(true)
    expect(matchesSearch(task, 'LOGIN')).toBe(true)
  })

  it('matches description (case-insensitive)', () => {
    const task = createMockTask({
      title: 'Task',
      description: 'Users cannot authenticate',
    })
    expect(matchesSearch(task, 'authenticate')).toBe(true)
  })

  it('matches task ID', () => {
    const task = createMockTask({ id: 'abc123xyz' })
    expect(matchesSearch(task, 'abc123')).toBe(true)
  })

  it('matches labels', () => {
    const task = createMockTask({ labels: ['frontend', 'urgent'] })
    expect(matchesSearch(task, 'front')).toBe(true)
  })

  it('returns false when no match found', () => {
    const task = createMockTask({
      title: 'Fix Bug',
      description: 'Something broken',
      labels: ['backend'],
    })
    expect(matchesSearch(task, 'frontend')).toBe(false)
  })
})

describe('matchesFilter (integration)', () => {
  it('matches priority filter', () => {
    const task = createMockTask({ priority: 'high' })
    const filter = { id: '1', field: 'priority' as const, operator: 'is' as const, value: 'high' }
    expect(matchesFilter(task, filter)).toBe(true)
  })

  it('matches column filter', () => {
    const task = createMockTask({ column: 'in_progress' })
    const filter = { id: '1', field: 'column' as const, operator: 'is' as const, value: 'in_progress' }
    expect(matchesFilter(task, filter)).toBe(true)
  })

  it('matches type filter', () => {
    const task = createMockTask({ type: 'bug' })
    const filter = { id: '1', field: 'type' as const, operator: 'is' as const, value: 'bug' }
    expect(matchesFilter(task, filter)).toBe(true)
  })

  it('matches labels filter with includes_any', () => {
    const task = createMockTask({ labels: ['frontend', 'urgent'] })
    const filter = { id: '1', field: 'labels' as const, operator: 'includes_any' as const, value: ['urgent', 'backend'] }
    expect(matchesFilter(task, filter)).toBe(true)
  })

  it('matches due_date filter', () => {
    const task = createMockTask({ due_date: '2024-03-20' })
    const filter = { id: '1', field: 'due_date' as const, operator: 'before' as const, value: '2024-03-25' }
    expect(matchesFilter(task, filter)).toBe(true)
  })

  it('matches title filter', () => {
    const task = createMockTask({ title: 'Fix Critical Bug' })
    const filter = { id: '1', field: 'title' as const, operator: 'contains' as const, value: 'critical' }
    expect(matchesFilter(task, filter)).toBe(true)
  })
})

describe('filter combination (AND/OR logic)', () => {
  const tasks = [
    createMockTask({ id: '1', priority: 'high', type: 'bug', labels: ['frontend'] }),
    createMockTask({ id: '2', priority: 'high', type: 'feature', labels: ['backend'] }),
    createMockTask({ id: '3', priority: 'low', type: 'bug', labels: ['frontend'] }),
    createMockTask({ id: '4', priority: 'low', type: 'feature', labels: ['docs'] }),
  ]

  it('AND: all filters must match', () => {
    const filters = [
      { id: '1', field: 'priority' as const, operator: 'is' as const, value: 'high' },
      { id: '2', field: 'type' as const, operator: 'is' as const, value: 'bug' },
    ]

    const result = tasks.filter((task) => filters.every((f) => matchesFilter(task, f)))

    expect(result).toHaveLength(1)
    expect(result[0].id).toBe('1')
  })

  it('OR: at least one filter must match', () => {
    const filters = [
      { id: '1', field: 'priority' as const, operator: 'is' as const, value: 'high' },
      { id: '2', field: 'labels' as const, operator: 'includes_any' as const, value: ['docs'] },
    ]

    const result = tasks.filter((task) => filters.some((f) => matchesFilter(task, f)))

    expect(result).toHaveLength(3) // tasks 1, 2 (high priority), and 4 (docs label)
  })

  it('returns all tasks when no filters', () => {
    const result = tasks.filter(() => true)
    expect(result).toHaveLength(4)
  })

  it('combines search and filters correctly', () => {
    const task = createMockTask({
      title: 'Fix Login Bug',
      priority: 'high',
      labels: ['frontend'],
    })

    // Both search and filter should pass
    const searchMatch = matchesSearch(task, 'login')
    const filterMatch = matchesFilter(task, {
      id: '1',
      field: 'priority',
      operator: 'is',
      value: 'high',
    })

    expect(searchMatch && filterMatch).toBe(true)
  })
})

describe('edge cases', () => {
  it('handles undefined labels gracefully', () => {
    const task = createMockTask({ title: 'No Match', labels: undefined })
    // Should not crash when labels are undefined and searching for a label value
    expect(matchesSearch(task, 'nonexistent-label-xyz')).toBe(false)
  })

  it('handles undefined description gracefully', () => {
    const task = createMockTask({ description: undefined })
    expect(matchesSearch(task, 'description')).toBe(false)
  })

  it('handles null filter value gracefully', () => {
    expect(matchSelectFilter('high', 'is', null)).toBe(true)
  })

  it('handles empty string task value', () => {
    expect(matchTextFilter('', 'contains', 'test')).toBe(false)
  })
})

// Integration tests for hooks with Zustand store
describe('useFilteredTasks hook integration', () => {
  const mockTasks = [
    createMockTask({ id: '1', title: 'High Priority Bug', priority: 'high', type: 'bug', labels: ['frontend'] }),
    createMockTask({ id: '2', title: 'Medium Feature', priority: 'medium', type: 'feature', labels: ['backend'] }),
    createMockTask({ id: '3', title: 'Low Bug', priority: 'low', type: 'bug', labels: ['frontend'] }),
    createMockTask({ id: '4', title: 'Urgent Chore', priority: 'urgent', type: 'chore', labels: ['docs'] }),
  ]

  beforeEach(() => {
    // Reset the store to initial state before each test
    useFilterStore.setState({
      filters: [],
      matchMode: 'all',
      searchQuery: '',
      debouncedSearchQuery: '',
      displayOptions: {
        viewMode: 'board',
        density: 'comfortable',
        visibleFields: ['priority', 'labels', 'due_date'],
        groupBy: 'column',
      },
      currentViewId: null,
      isModified: false,
    })
  })

  it('returns all tasks when no filters are applied', () => {
    const { result } = renderHook(() => useFilteredTasks(mockTasks))
    expect(result.current).toHaveLength(4)
  })

  it('filters tasks by single filter with "all" match mode', () => {
    // Add a filter for high priority
    act(() => {
      useFilterStore.getState().addFilter({
        field: 'priority',
        operator: 'is',
        value: 'high',
      })
    })

    const { result } = renderHook(() => useFilteredTasks(mockTasks))
    expect(result.current).toHaveLength(1)
    expect(result.current[0].id).toBe('1')
  })

  it('filters tasks with multiple filters using "all" match mode (AND logic)', () => {
    act(() => {
      useFilterStore.getState().addFilter({
        field: 'type',
        operator: 'is',
        value: 'bug',
      })
      useFilterStore.getState().addFilter({
        field: 'labels',
        operator: 'includes_any',
        value: ['frontend'],
      })
    })

    const { result } = renderHook(() => useFilteredTasks(mockTasks))
    // Should match tasks that are bugs AND have frontend label
    expect(result.current).toHaveLength(2) // task 1 and task 3
    expect(result.current.map((t) => t.id)).toEqual(['1', '3'])
  })

  it('filters tasks with multiple filters using "any" match mode (OR logic)', () => {
    act(() => {
      useFilterStore.getState().setMatchMode('any')
      useFilterStore.getState().addFilter({
        field: 'priority',
        operator: 'is',
        value: 'urgent',
      })
      useFilterStore.getState().addFilter({
        field: 'type',
        operator: 'is',
        value: 'feature',
      })
    })

    const { result } = renderHook(() => useFilteredTasks(mockTasks))
    // Should match tasks that are urgent OR features
    expect(result.current).toHaveLength(2) // task 2 (feature) and task 4 (urgent)
    expect(result.current.map((t) => t.id)).toEqual(['2', '4'])
  })

  it('filters tasks by debounced search query', () => {
    act(() => {
      // Directly set the debounced search query (simulating the debounce completion)
      useFilterStore.setState({ debouncedSearchQuery: 'bug' })
    })

    const { result } = renderHook(() => useFilteredTasks(mockTasks))
    // Should match tasks with "bug" in title
    expect(result.current).toHaveLength(2)
    expect(result.current.map((t) => t.id)).toEqual(['1', '3'])
  })

  it('combines search and filters correctly', () => {
    act(() => {
      // Set search query (debounced)
      useFilterStore.setState({ debouncedSearchQuery: 'bug' })
      // Add filter for high priority
      useFilterStore.getState().addFilter({
        field: 'priority',
        operator: 'is',
        value: 'high',
      })
    })

    const { result } = renderHook(() => useFilteredTasks(mockTasks))
    // Should match high priority tasks with "bug" in title
    expect(result.current).toHaveLength(1)
    expect(result.current[0].id).toBe('1')
  })

  it('returns empty array when no tasks match', () => {
    act(() => {
      useFilterStore.getState().addFilter({
        field: 'priority',
        operator: 'is',
        value: 'critical', // non-existent priority
      })
    })

    const { result } = renderHook(() => useFilteredTasks(mockTasks))
    expect(result.current).toHaveLength(0)
  })

  it('updates filtered results when store changes', () => {
    const { result, rerender } = renderHook(() => useFilteredTasks(mockTasks))
    
    // Initially all tasks
    expect(result.current).toHaveLength(4)

    // Add a filter
    act(() => {
      useFilterStore.getState().addFilter({
        field: 'type',
        operator: 'is',
        value: 'bug',
      })
    })

    // Rerender to get updated result
    rerender()
    expect(result.current).toHaveLength(2)

    // Remove the filter
    act(() => {
      const filters = useFilterStore.getState().filters
      if (filters.length > 0) {
        useFilterStore.getState().removeFilter(filters[0].id)
      }
    })

    rerender()
    expect(result.current).toHaveLength(4)
  })
})

describe('useSearchDebounce hook integration', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    // Reset the store
    useFilterStore.setState({
      searchQuery: '',
      debouncedSearchQuery: '',
    })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('debounces search query updates', async () => {
    renderHook(() => useSearchDebounce())

    // Set search query
    act(() => {
      useFilterStore.getState().setSearchQuery('test')
    })

    // Immediately after, debounced query should still be empty
    expect(useFilterStore.getState().debouncedSearchQuery).toBe('')

    // Fast forward past debounce delay (300ms)
    act(() => {
      vi.advanceTimersByTime(300)
    })

    // Now debounced query should be updated
    expect(useFilterStore.getState().debouncedSearchQuery).toBe('test')
  })

  it('cancels pending debounce when query changes', async () => {
    renderHook(() => useSearchDebounce())

    // Set initial search query
    act(() => {
      useFilterStore.getState().setSearchQuery('first')
    })

    // Advance only 100ms (not enough for debounce)
    act(() => {
      vi.advanceTimersByTime(100)
    })

    // Change query before debounce completes
    act(() => {
      useFilterStore.getState().setSearchQuery('second')
    })

    // Advance past original debounce time
    act(() => {
      vi.advanceTimersByTime(200)
    })

    // Should still be empty because the second query restarted the timer
    expect(useFilterStore.getState().debouncedSearchQuery).toBe('')

    // Complete the debounce for second query
    act(() => {
      vi.advanceTimersByTime(100)
    })

    // Now it should have the second value
    expect(useFilterStore.getState().debouncedSearchQuery).toBe('second')
  })
})
