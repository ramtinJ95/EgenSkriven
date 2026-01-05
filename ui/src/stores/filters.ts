import { create } from 'zustand'
import { persist } from 'zustand/middleware'

// Filter operators for different field types
export type FilterOperator =
  | 'is'
  | 'is_not'
  | 'is_any_of'
  | 'includes_any'
  | 'includes_all'
  | 'includes_none'
  | 'before'
  | 'after'
  | 'is_set'
  | 'is_not_set'
  | 'contains'

// Fields that can be filtered
export type FilterField =
  | 'column'
  | 'priority'
  | 'type'
  | 'labels'
  | 'due_date'
  | 'epic'
  | 'created_by'
  | 'title'

// Single filter definition
export interface Filter {
  id: string
  field: FilterField
  operator: FilterOperator
  value: string | string[] | null
}

// Display options for board/list view
export interface DisplayOptions {
  viewMode: 'board' | 'list'
  density: 'compact' | 'comfortable'
  visibleFields: string[]
  groupBy: 'column' | 'priority' | 'type' | 'epic' | null
}

// How multiple filters are combined
export type MatchMode = 'all' | 'any'

// Filter state interface
interface FilterState {
  // State
  filters: Filter[]
  matchMode: MatchMode
  searchQuery: string
  debouncedSearchQuery: string
  displayOptions: DisplayOptions
  currentViewId: string | null
  isModified: boolean

  // Actions
  addFilter: (filter: Omit<Filter, 'id'>) => void
  removeFilter: (filterId: string) => void
  updateFilter: (filterId: string, updates: Partial<Filter>) => void
  clearFilters: () => void
  setMatchMode: (mode: MatchMode) => void
  setSearchQuery: (query: string) => void
  setDebouncedSearchQuery: (query: string) => void
  setDisplayOptions: (options: Partial<DisplayOptions>) => void
  loadView: (
    viewId: string,
    filters: Filter[],
    matchMode: MatchMode,
    display: DisplayOptions
  ) => void
  clearView: () => void
  markAsModified: () => void
}

// Generate random ID for filters
const generateId = () => Math.random().toString(36).substring(2, 9)

// Default display options
const defaultDisplayOptions: DisplayOptions = {
  viewMode: 'board',
  density: 'comfortable',
  visibleFields: ['priority', 'labels', 'due_date'],
  groupBy: 'column',
}

// Create the store with localStorage persistence
export const useFilterStore = create<FilterState>()(
  persist(
    (set) => ({
      // Initial state
      filters: [],
      matchMode: 'all',
      searchQuery: '',
      debouncedSearchQuery: '',
      displayOptions: defaultDisplayOptions,
      currentViewId: null,
      isModified: false,

      // Add a new filter
      addFilter: (filter) =>
        set((state) => ({
          filters: [...state.filters, { ...filter, id: generateId() }],
          isModified: true,
        })),

      // Remove a filter by ID
      removeFilter: (filterId) =>
        set((state) => ({
          filters: state.filters.filter((f) => f.id !== filterId),
          isModified: true,
        })),

      // Update an existing filter
      updateFilter: (filterId, updates) =>
        set((state) => ({
          filters: state.filters.map((f) =>
            f.id === filterId ? { ...f, ...updates } : f
          ),
          isModified: true,
        })),

      // Clear all filters and search
      clearFilters: () =>
        set({
          filters: [],
          searchQuery: '',
          debouncedSearchQuery: '',
          isModified: true,
        }),

      // Set the match mode (all/any)
      setMatchMode: (mode) => set({ matchMode: mode, isModified: true }),

      // Set search query (immediate, for UI display)
      setSearchQuery: (query) => set({ searchQuery: query }),

      // Set debounced search query (for filtering, called after debounce)
      setDebouncedSearchQuery: (query) => set({ debouncedSearchQuery: query }),

      // Update display options
      setDisplayOptions: (options) =>
        set((state) => ({
          displayOptions: { ...state.displayOptions, ...options },
          isModified: true,
        })),

      // Load a saved view
      loadView: (viewId, filters, matchMode, display) =>
        set({
          currentViewId: viewId,
          filters,
          matchMode,
          displayOptions: display,
          isModified: false,
        }),

      // Clear the current view (reset to defaults)
      clearView: () =>
        set({
          currentViewId: null,
          filters: [],
          matchMode: 'all',
          searchQuery: '',
          debouncedSearchQuery: '',
          displayOptions: defaultDisplayOptions,
          isModified: false,
        }),

      // Mark the current view as modified
      markAsModified: () => set({ isModified: true }),
    }),
    {
      name: 'egenskriven-filters',
      // Only persist these fields to localStorage
      partialize: (state) => ({
        filters: state.filters,
        matchMode: state.matchMode,
        displayOptions: state.displayOptions,
        currentViewId: state.currentViewId,
      }),
    }
  )
)

// Helper: Get operators available for a field type
export function getOperatorsForField(field: FilterField): FilterOperator[] {
  switch (field) {
    case 'column':
    case 'priority':
    case 'type':
    case 'created_by':
      return ['is', 'is_not', 'is_any_of', 'is_set', 'is_not_set']
    case 'labels':
      return ['includes_any', 'includes_all', 'includes_none', 'is_set', 'is_not_set']
    case 'due_date':
      return ['is', 'before', 'after', 'is_set', 'is_not_set']
    case 'epic':
      return ['is', 'is_not', 'is_set', 'is_not_set']
    case 'title':
      return ['contains', 'is', 'is_not']
    default:
      return ['is', 'is_not']
  }
}

// Helper: Check if an operator requires a value
export function operatorRequiresValue(operator: FilterOperator): boolean {
  return operator !== 'is_set' && operator !== 'is_not_set'
}
