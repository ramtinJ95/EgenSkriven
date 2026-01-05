import type { Filter, FilterField } from '../stores/filters'
import { useFilterStore } from '../stores/filters'
import styles from './FilterBar.module.css'

// Human-readable field labels
const FIELD_LABELS: Record<FilterField, string> = {
  column: 'Status',
  priority: 'Priority',
  type: 'Type',
  labels: 'Labels',
  due_date: 'Due Date',
  epic: 'Epic',
  created_by: 'Created By',
  title: 'Title',
}

// Human-readable operator labels
const OPERATOR_LABELS: Record<string, string> = {
  is: 'is',
  is_not: 'is not',
  is_any_of: 'is any of',
  includes_any: 'includes any of',
  includes_all: 'includes all of',
  includes_none: 'includes none of',
  before: 'before',
  after: 'after',
  is_set: 'is set',
  is_not_set: 'is not set',
  contains: 'contains',
}

interface FilterPillProps {
  filter: Filter
  onRemove: () => void
}

function FilterPill({ filter, onRemove }: FilterPillProps) {
  const formatValue = () => {
    if (filter.value === null) return ''
    if (Array.isArray(filter.value)) {
      return filter.value.join(', ')
    }
    return String(filter.value)
  }

  const showValue = filter.operator !== 'is_set' && filter.operator !== 'is_not_set'

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Delete' || e.key === 'Backspace') {
      e.preventDefault()
      onRemove()
    }
  }

  const filterDescription = `${FIELD_LABELS[filter.field]} ${OPERATOR_LABELS[filter.operator]}${showValue && filter.value !== null ? ` ${formatValue()}` : ''}`

  return (
    <div
      className={styles.pill}
      tabIndex={0}
      role="button"
      aria-label={`Filter: ${filterDescription}. Press Delete or Backspace to remove.`}
      onKeyDown={handleKeyDown}
    >
      <span className={styles.pillField}>{FIELD_LABELS[filter.field]}</span>
      <span className={styles.pillOperator}>{OPERATOR_LABELS[filter.operator]}</span>
      {showValue && filter.value !== null && (
        <span className={styles.pillValue}>{formatValue()}</span>
      )}
      <button
        className={styles.pillRemove}
        onClick={onRemove}
        aria-label={`Remove ${FIELD_LABELS[filter.field]} filter`}
        tabIndex={-1}
      >
        ×
      </button>
    </div>
  )
}

export interface FilterBarProps {
  totalTasks: number
  filteredTasks: number
  onOpenFilterBuilder: () => void
}

export function FilterBar({
  totalTasks,
  filteredTasks,
  onOpenFilterBuilder,
}: FilterBarProps) {
  const filters = useFilterStore((s) => s.filters)
  const matchMode = useFilterStore((s) => s.matchMode)
  const searchQuery = useFilterStore((s) => s.searchQuery)
  const clearFilters = useFilterStore((s) => s.clearFilters)
  const removeFilter = useFilterStore((s) => s.removeFilter)
  const setMatchMode = useFilterStore((s) => s.setMatchMode)
  const setSearchQuery = useFilterStore((s) => s.setSearchQuery)

  const hasActiveFilters = filters.length > 0 || searchQuery.trim() !== ''

  return (
    <div className={styles.filterBar}>
      <button className={styles.filterButton} onClick={onOpenFilterBuilder}>
        <svg
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
        >
          <polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3" />
        </svg>
        Filter
        {filters.length > 0 && (
          <span className={styles.filterCount}>{filters.length}</span>
        )}
      </button>

      {hasActiveFilters && (
        <div className={styles.pillsContainer}>
          {searchQuery && (
            <div className={styles.pill}>
              <span className={styles.pillField}>Search</span>
              <span className={styles.pillValue}>"{searchQuery}"</span>
              <button
                className={styles.pillRemove}
                onClick={() => setSearchQuery('')}
                aria-label="Clear search"
              >
                ×
              </button>
            </div>
          )}

          {filters.map((f, index) => (
            <div key={f.id} className={styles.pillWrapper}>
              {index > 0 && (
                <span className={styles.connector}>
                  {matchMode === 'all' ? 'and' : 'or'}
                </span>
              )}
              <FilterPill filter={f} onRemove={() => removeFilter(f.id)} />
            </div>
          ))}

          {filters.length > 1 && (
            <div className={styles.matchModeToggle}>
              <button
                className={`${styles.matchModeButton} ${matchMode === 'all' ? styles.active : ''}`}
                onClick={() => setMatchMode('all')}
              >
                All
              </button>
              <button
                className={`${styles.matchModeButton} ${matchMode === 'any' ? styles.active : ''}`}
                onClick={() => setMatchMode('any')}
              >
                Any
              </button>
            </div>
          )}

          <button className={styles.clearButton} onClick={clearFilters}>
            Clear all
          </button>
        </div>
      )}

      <div className={styles.stats}>
        {hasActiveFilters
          ? `${filteredTasks} of ${totalTasks} tasks`
          : `${totalTasks} tasks`}
      </div>
    </div>
  )
}
