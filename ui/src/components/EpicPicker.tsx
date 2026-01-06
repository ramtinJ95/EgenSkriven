import { useState, useRef, useEffect } from 'react'
import { useEpics } from '../hooks/useEpics'
import styles from './EpicPicker.module.css'

interface EpicPickerProps {
  value: string | null          // Epic ID or null
  onChange: (epicId: string | null) => void
  placeholder?: string
}

/**
 * Epic picker component for selecting an epic for a task.
 *
 * Features:
 * - Dropdown with search filter
 * - Color indicators for each epic
 * - "No epic" option to clear selection
 * - Checkmark for selected epic
 * - Click outside to close
 * - Escape key to close
 */
export function EpicPicker({ value, onChange, placeholder = 'Set epic' }: EpicPickerProps) {
  const { epics, loading } = useEpics()
  const [isOpen, setIsOpen] = useState(false)
  const [search, setSearch] = useState('')
  const containerRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  // Filter epics by search term
  const filteredEpics = epics.filter(epic =>
    epic.title.toLowerCase().includes(search.toLowerCase())
  )

  // Get currently selected epic
  const selectedEpic = value ? epics.find(e => e.id === value) : null

  // Close picker when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false)
        setSearch('')
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  // Close on Escape key
  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape' && isOpen) {
        event.stopPropagation()
        setIsOpen(false)
        setSearch('')
      }
    }
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [isOpen])

  // Focus search input when dropdown opens
  useEffect(() => {
    if (isOpen && inputRef.current) {
      inputRef.current.focus()
    }
  }, [isOpen])

  const handleSelect = (epicId: string | null) => {
    onChange(epicId)
    setIsOpen(false)
    setSearch('')
  }

  // Format display value
  const displayValue = selectedEpic ? selectedEpic.title : placeholder

  const triggerClasses = [
    styles.trigger,
    selectedEpic ? styles.hasValue : '',
  ].filter(Boolean).join(' ')

  return (
    <div className={styles.epicPicker} ref={containerRef}>
      {/* Trigger - using div with role="button" to allow nested clear button */}
      <div
        className={triggerClasses}
        onClick={() => setIsOpen(!isOpen)}
        role="button"
        tabIndex={0}
        aria-label="Select epic"
        aria-expanded={isOpen}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            setIsOpen(!isOpen)
          }
        }}
      >
        {selectedEpic && (
          <span
            className={styles.colorDot}
            style={{ backgroundColor: selectedEpic.color || '#5E6AD2' }}
          />
        )}
        <span className={styles.value}>{displayValue}</span>
        {selectedEpic && (
          <button
            className={styles.clear}
            onClick={(e) => {
              e.stopPropagation()
              onChange(null)
            }}
            aria-label="Clear epic"
            type="button"
          >
            &times;
          </button>
        )}
      </div>

      {/* Dropdown */}
      {isOpen && (
        <div className={styles.dropdown}>
          {/* Search input */}
          <div className={styles.searchContainer}>
            <input
              ref={inputRef}
              type="text"
              className={styles.searchInput}
              placeholder="Search epics..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>

          {/* Epic list */}
          <div className={styles.list}>
            {loading ? (
              <div className={styles.loading}>Loading epics...</div>
            ) : (
              <>
                {/* "No epic" option */}
                <button
                  className={`${styles.option} ${!value ? styles.selected : ''}`}
                  onClick={() => handleSelect(null)}
                  type="button"
                >
                  <span className={styles.colorDot} style={{ backgroundColor: '#666' }} />
                  <span className={styles.optionName}>No epic</span>
                  {!value && <span className={styles.checkmark}>&#10003;</span>}
                </button>

                {/* Epic options */}
                {filteredEpics.map(epic => (
                  <button
                    key={epic.id}
                    className={`${styles.option} ${value === epic.id ? styles.selected : ''}`}
                    onClick={() => handleSelect(epic.id)}
                    type="button"
                  >
                    <span
                      className={styles.colorDot}
                      style={{ backgroundColor: epic.color || '#5E6AD2' }}
                    />
                    <span className={styles.optionName}>{epic.title}</span>
                    {value === epic.id && <span className={styles.checkmark}>&#10003;</span>}
                  </button>
                ))}

                {/* Empty state when search has no matches */}
                {filteredEpics.length === 0 && search && (
                  <div className={styles.empty}>
                    No epics matching "{search}"
                  </div>
                )}

                {/* Empty state when no epics exist */}
                {epics.length === 0 && !search && (
                  <div className={styles.empty}>
                    No epics created yet.
                    <br />
                    <span className={styles.hint}>
                      Create one with: egenskriven epic add "Epic name"
                    </span>
                  </div>
                )}
              </>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
