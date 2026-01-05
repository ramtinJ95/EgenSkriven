import { useState, useEffect } from 'react'
import { pb } from '../lib/pb'
import type { FilterField, FilterOperator } from '../stores/filters'
import {
  useFilterStore,
  getOperatorsForField,
  operatorRequiresValue,
} from '../stores/filters'
import { COLUMNS, PRIORITIES, TYPES, COLUMN_NAMES, PRIORITY_NAMES, TYPE_NAMES } from '../types/task'
import styles from './FilterBuilder.module.css'

// Filter field options
const FILTER_FIELDS: { value: FilterField; label: string }[] = [
  { value: 'column', label: 'Status' },
  { value: 'priority', label: 'Priority' },
  { value: 'type', label: 'Type' },
  { value: 'labels', label: 'Labels' },
  { value: 'due_date', label: 'Due Date' },
  { value: 'epic', label: 'Epic' },
  { value: 'title', label: 'Title' },
]

// Operator labels for display
const OPERATOR_LABELS: Record<FilterOperator, string> = {
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

// Epic type from PocketBase
interface Epic {
  id: string
  title: string
  color?: string
}

export interface FilterBuilderProps {
  isOpen: boolean
  onClose: () => void
}

export function FilterBuilder({ isOpen, onClose }: FilterBuilderProps) {
  const addFilter = useFilterStore((s) => s.addFilter)
  const filters = useFilterStore((s) => s.filters)
  const matchMode = useFilterStore((s) => s.matchMode)
  const setMatchMode = useFilterStore((s) => s.setMatchMode)
  const removeFilter = useFilterStore((s) => s.removeFilter)

  // Form state
  const [selectedField, setSelectedField] = useState<FilterField>('priority')
  const [selectedOperator, setSelectedOperator] = useState<FilterOperator>('is')
  const [selectedValue, setSelectedValue] = useState<string | string[]>('')
  const [textValue, setTextValue] = useState('')
  const [dateValue, setDateValue] = useState('')

  // Epics data
  const [epics, setEpics] = useState<Epic[]>([])

  // Available labels from localStorage or default
  const [availableLabels] = useState<string[]>(() => {
    // In a real app, you might fetch these from the backend
    // For now, use some common labels
    return ['frontend', 'backend', 'bug', 'feature', 'urgent', 'docs', 'ui', 'api', 'testing']
  })

  // Fetch epics on mount
  useEffect(() => {
    pb.collection('epics')
      .getFullList<Epic>()
      .then(setEpics)
      .catch(console.error)
  }, [])

  // Reset operator and value when field changes
  useEffect(() => {
    const operators = getOperatorsForField(selectedField)
    setSelectedOperator(operators[0])
    setSelectedValue('')
    setTextValue('')
    setDateValue('')
  }, [selectedField])

  // Get available operators for the selected field
  const availableOperators = getOperatorsForField(selectedField)

  // Handle adding a filter
  const handleAddFilter = () => {
    let value: string | string[] | null = null

    if (operatorRequiresValue(selectedOperator)) {
      switch (selectedField) {
        case 'title':
          value = textValue
          break
        case 'due_date':
          value = dateValue
          break
        case 'labels':
          value = Array.isArray(selectedValue) ? selectedValue : [selectedValue].filter(Boolean)
          break
        default:
          value = selectedValue as string
      }

      // Don't add if no value provided
      if (!value || (Array.isArray(value) && value.length === 0)) return
    }

    addFilter({
      field: selectedField,
      operator: selectedOperator,
      value,
    })

    // Reset form
    setSelectedValue('')
    setTextValue('')
    setDateValue('')
  }

  // Render value input based on field type
  const renderValueInput = () => {
    if (!operatorRequiresValue(selectedOperator)) {
      return null
    }

    switch (selectedField) {
      case 'column':
        return (
          <select
            className={styles.select}
            value={selectedValue as string}
            onChange={(e) => setSelectedValue(e.target.value)}
          >
            <option value="">Select status...</option>
            {COLUMNS.map((col) => (
              <option key={col} value={col}>
                {COLUMN_NAMES[col]}
              </option>
            ))}
          </select>
        )

      case 'priority':
        return (
          <select
            className={styles.select}
            value={selectedValue as string}
            onChange={(e) => setSelectedValue(e.target.value)}
          >
            <option value="">Select priority...</option>
            {PRIORITIES.map((p) => (
              <option key={p} value={p}>
                {PRIORITY_NAMES[p]}
              </option>
            ))}
          </select>
        )

      case 'type':
        return (
          <select
            className={styles.select}
            value={selectedValue as string}
            onChange={(e) => setSelectedValue(e.target.value)}
          >
            <option value="">Select type...</option>
            {TYPES.map((t) => (
              <option key={t} value={t}>
                {TYPE_NAMES[t]}
              </option>
            ))}
          </select>
        )

      case 'labels':
        return (
          <div className={styles.multiSelect}>
            {availableLabels.map((label) => {
              const isSelected = Array.isArray(selectedValue) && selectedValue.includes(label)
              return (
                <button
                  key={label}
                  type="button"
                  className={`${styles.labelChip} ${isSelected ? styles.selected : ''}`}
                  onClick={() => {
                    const current = Array.isArray(selectedValue) ? selectedValue : []
                    if (isSelected) {
                      setSelectedValue(current.filter((l) => l !== label))
                    } else {
                      setSelectedValue([...current, label])
                    }
                  }}
                >
                  {label}
                </button>
              )
            })}
          </div>
        )

      case 'due_date':
        return (
          <input
            type="date"
            className={styles.input}
            value={dateValue}
            onChange={(e) => setDateValue(e.target.value)}
          />
        )

      case 'epic':
        return (
          <select
            className={styles.select}
            value={selectedValue as string}
            onChange={(e) => setSelectedValue(e.target.value)}
          >
            <option value="">Select epic...</option>
            {epics.map((epic) => (
              <option key={epic.id} value={epic.id}>
                {epic.title}
              </option>
            ))}
          </select>
        )

      case 'title':
        return (
          <input
            type="text"
            className={styles.input}
            placeholder="Enter text..."
            value={textValue}
            onChange={(e) => setTextValue(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleAddFilter()}
          />
        )

      default:
        return null
    }
  }

  if (!isOpen) return null

  return (
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
        <div className={styles.header}>
          <h3>Filter Tasks</h3>
          <button className={styles.closeButton} onClick={onClose}>
            &times;
          </button>
        </div>

        <div className={styles.content}>
          {/* Existing filters */}
          {filters.length > 0 && (
            <div className={styles.existingFilters}>
              <div className={styles.filtersHeader}>
                <span className={styles.filtersLabel}>Active filters</span>
                {filters.length > 1 && (
                  <div className={styles.matchModeToggle}>
                    <span className={styles.matchModeLabel}>Match:</span>
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
              </div>
              <div className={styles.filtersList}>
                {filters.map((filter, index) => (
                  <div key={filter.id} className={styles.filterItem}>
                    {index > 0 && (
                      <span className={styles.connector}>
                        {matchMode === 'all' ? 'AND' : 'OR'}
                      </span>
                    )}
                    <div className={styles.filterPill}>
                      <span className={styles.filterField}>
                        {FILTER_FIELDS.find((f) => f.value === filter.field)?.label}
                      </span>
                      <span className={styles.filterOperator}>
                        {OPERATOR_LABELS[filter.operator]}
                      </span>
                      {filter.value !== null && (
                        <span className={styles.filterValue}>
                          {Array.isArray(filter.value)
                            ? filter.value.join(', ')
                            : String(filter.value)}
                        </span>
                      )}
                      <button
                        className={styles.filterRemove}
                        onClick={() => removeFilter(filter.id)}
                      >
                        &times;
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Add new filter */}
          <div className={styles.addFilter}>
            <div className={styles.addFilterLabel}>Add filter</div>
            <div className={styles.filterForm}>
              <select
                className={styles.select}
                value={selectedField}
                onChange={(e) => setSelectedField(e.target.value as FilterField)}
              >
                {FILTER_FIELDS.map((field) => (
                  <option key={field.value} value={field.value}>
                    {field.label}
                  </option>
                ))}
              </select>

              <select
                className={styles.select}
                value={selectedOperator}
                onChange={(e) => setSelectedOperator(e.target.value as FilterOperator)}
              >
                {availableOperators.map((op) => (
                  <option key={op} value={op}>
                    {OPERATOR_LABELS[op]}
                  </option>
                ))}
              </select>

              {renderValueInput()}

              <button
                className={styles.addButton}
                onClick={handleAddFilter}
                disabled={
                  operatorRequiresValue(selectedOperator) &&
                  !selectedValue &&
                  !textValue &&
                  !dateValue
                }
              >
                Add
              </button>
            </div>
          </div>
        </div>

        <div className={styles.footer}>
          <button className={styles.doneButton} onClick={onClose}>
            Done
          </button>
        </div>
      </div>
    </div>
  )
}
