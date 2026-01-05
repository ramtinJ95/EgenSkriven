import { useMemo, useEffect, useRef } from 'react'
import type { Task } from '../types/task'
import type { Filter } from '../stores/filters'
import { useFilterStore } from '../stores/filters'

// Debounce delay for search (ms)
const SEARCH_DEBOUNCE_MS = 300

/**
 * Hook to debounce the search query
 * Updates debouncedSearchQuery in the store after delay
 */
export function useSearchDebounce(): void {
  const searchQuery = useFilterStore((state) => state.searchQuery)
  const setDebouncedSearchQuery = useFilterStore(
    (state) => state.setDebouncedSearchQuery
  )
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    // Set new timeout to update debounced value
    const timeoutId = setTimeout(() => {
      setDebouncedSearchQuery(searchQuery)
    }, SEARCH_DEBOUNCE_MS)

    // Cleanup on unmount or when searchQuery changes
    return () => {
      clearTimeout(timeoutId)
    }
    // Note: setDebouncedSearchQuery is a stable Zustand action, but included for correctness
  }, [searchQuery, setDebouncedSearchQuery])
}

/**
 * Check if a task matches a single filter
 */
function matchesFilter(task: Task, filter: Filter): boolean {
  const { field, operator, value } = filter

  switch (field) {
    case 'column':
    case 'priority':
    case 'type':
    case 'created_by':
      return matchSelectFilter(task[field], operator, value)
    case 'labels':
      return matchLabelsFilter(task.labels || [], operator, value as string[])
    case 'due_date':
      return matchDateFilter(task.due_date, operator, value as string)
    case 'epic':
      return matchRelationFilter(
        (task as Task & { epic?: string }).epic,
        operator,
        value as string
      )
    case 'title':
      return matchTextFilter(task.title, operator, value as string)
    default:
      return true
  }
}

/**
 * Match single-select fields (column, priority, type, created_by)
 */
function matchSelectFilter(
  taskValue: string | undefined,
  operator: string,
  filterValue: string | string[] | null
): boolean {
  // Handle is_set / is_not_set operators
  if (operator === 'is_set') return !!taskValue
  if (operator === 'is_not_set') return !taskValue

  // If task value is empty and we're not checking set/not_set
  if (!taskValue) return false

  // If filter value is null, treat as checking for existence
  if (filterValue === null) return true

  switch (operator) {
    case 'is':
      return taskValue === filterValue
    case 'is_not':
      return taskValue !== filterValue
    case 'is_any_of':
      return Array.isArray(filterValue) && filterValue.includes(taskValue)
    default:
      return true
  }
}

/**
 * Match array fields (labels)
 */
function matchLabelsFilter(
  taskLabels: string[],
  operator: string,
  filterLabels: string[]
): boolean {
  // Handle is_set / is_not_set operators
  if (operator === 'is_set') return taskLabels.length > 0
  if (operator === 'is_not_set') return taskLabels.length === 0

  // If no filter labels specified, consider it a match
  if (!filterLabels?.length) return true

  switch (operator) {
    case 'includes_any':
      // Task has at least one of the filter labels
      return filterLabels.some((l) => taskLabels.includes(l))
    case 'includes_all':
      // Task has all of the filter labels
      return filterLabels.every((l) => taskLabels.includes(l))
    case 'includes_none':
      // Task has none of the filter labels
      return !filterLabels.some((l) => taskLabels.includes(l))
    default:
      return true
  }
}

/**
 * Match date fields (due_date)
 */
function matchDateFilter(
  taskDate: string | undefined,
  operator: string,
  filterDate: string | null
): boolean {
  // Handle is_set / is_not_set operators
  if (operator === 'is_set') return !!taskDate
  if (operator === 'is_not_set') return !taskDate

  // If either date is missing, can't compare
  if (!taskDate || !filterDate) return false

  const taskDateObj = new Date(taskDate)
  const filterDateObj = new Date(filterDate)

  // Reset to start of day for comparison
  taskDateObj.setHours(0, 0, 0, 0)
  filterDateObj.setHours(0, 0, 0, 0)

  switch (operator) {
    case 'is':
      return taskDateObj.getTime() === filterDateObj.getTime()
    case 'before':
      return taskDateObj < filterDateObj
    case 'after':
      return taskDateObj > filterDateObj
    default:
      return true
  }
}

/**
 * Match relation fields (epic)
 */
function matchRelationFilter(
  taskRelation: string | undefined,
  operator: string,
  filterValue: string | null
): boolean {
  // Handle is_set / is_not_set operators
  if (operator === 'is_set') return !!taskRelation
  if (operator === 'is_not_set') return !taskRelation

  // If filter value is null, treat as any
  if (filterValue === null) return true

  switch (operator) {
    case 'is':
      return taskRelation === filterValue
    case 'is_not':
      return taskRelation !== filterValue
    default:
      return true
  }
}

/**
 * Match text fields (title)
 */
function matchTextFilter(
  taskValue: string | undefined,
  operator: string,
  filterValue: string | null
): boolean {
  // If no filter value, consider it a match
  if (!filterValue) return true

  // If task value is empty
  if (!taskValue) return false

  const lowerTaskValue = taskValue.toLowerCase()
  const lowerFilterValue = filterValue.toLowerCase()

  switch (operator) {
    case 'contains':
      return lowerTaskValue.includes(lowerFilterValue)
    case 'is':
      return lowerTaskValue === lowerFilterValue
    case 'is_not':
      return lowerTaskValue !== lowerFilterValue
    default:
      return true
  }
}

/**
 * Check if a task matches the search query
 * Searches in title, description, and task ID
 */
function matchesSearch(task: Task, query: string): boolean {
  // Empty query matches everything
  if (!query?.trim()) return true

  const lowerQuery = query.toLowerCase().trim()

  // Search in title
  if ((task.title || '').toLowerCase().includes(lowerQuery)) return true

  // Search in description
  if ((task.description || '').toLowerCase().includes(lowerQuery)) return true

  // Search in task ID (supports partial match)
  if (task.id.toLowerCase().includes(lowerQuery)) return true

  // Search in labels
  if (task.labels?.some((label) => label.toLowerCase().includes(lowerQuery))) {
    return true
  }

  return false
}

/**
 * Main hook: Filter tasks based on current filter state
 * Uses useMemo for performance - only recalculates when dependencies change
 */
export function useFilteredTasks(tasks: Task[]): Task[] {
  const filters = useFilterStore((state) => state.filters)
  const matchMode = useFilterStore((state) => state.matchMode)
  const debouncedSearchQuery = useFilterStore(
    (state) => state.debouncedSearchQuery
  )

  return useMemo(() => {
    return tasks.filter((task) => {
      // First check search query
      if (!matchesSearch(task, debouncedSearchQuery)) {
        return false
      }

      // If no filters, task passes
      if (filters.length === 0) {
        return true
      }

      // Apply filters based on match mode
      if (matchMode === 'all') {
        // AND: all filters must match
        return filters.every((f) => matchesFilter(task, f))
      } else {
        // OR: at least one filter must match
        return filters.some((f) => matchesFilter(task, f))
      }
    })
  }, [tasks, filters, matchMode, debouncedSearchQuery])
}

/**
 * Export filter matching functions for testing
 */
export const filterHelpers = {
  matchesFilter,
  matchSelectFilter,
  matchLabelsFilter,
  matchDateFilter,
  matchRelationFilter,
  matchTextFilter,
  matchesSearch,
}
