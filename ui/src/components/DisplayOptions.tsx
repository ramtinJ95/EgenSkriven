import { useCallback, useEffect, useRef } from 'react'
import { useFilterStore } from '../stores/filters'
import styles from './DisplayOptions.module.css'

interface DisplayOptionsProps {
  isOpen: boolean
  onClose: () => void
}

const VISIBLE_FIELD_OPTIONS = [
  { key: 'priority', label: 'Priority' },
  { key: 'labels', label: 'Labels' },
  { key: 'due_date', label: 'Due Date' },
  { key: 'epic', label: 'Epic' },
  { key: 'type', label: 'Type' },
]

const GROUP_BY_OPTIONS = [
  { value: 'column', label: 'Status' },
  { value: 'priority', label: 'Priority' },
  { value: 'type', label: 'Type' },
  { value: 'epic', label: 'Epic' },
]

export function DisplayOptions({ isOpen, onClose }: DisplayOptionsProps) {
  const menuRef = useRef<HTMLDivElement>(null)
  const displayOptions = useFilterStore((s) => s.displayOptions)
  const setDisplayOptions = useFilterStore((s) => s.setDisplayOptions)

  // Close on click outside
  useEffect(() => {
    if (!isOpen) return

    const handleClickOutside = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        onClose()
      }
    }

    // Delay adding the listener to avoid immediate close
    const timeoutId = setTimeout(() => {
      document.addEventListener('mousedown', handleClickOutside)
    }, 0)

    return () => {
      clearTimeout(timeoutId)
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [isOpen, onClose])

  // Close on Escape
  useEffect(() => {
    if (!isOpen) return

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose()
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [isOpen, onClose])

  const handleViewModeChange = useCallback(
    (mode: 'board' | 'list') => {
      setDisplayOptions({ viewMode: mode })
    },
    [setDisplayOptions]
  )

  const handleDensityChange = useCallback(
    (density: 'compact' | 'comfortable') => {
      setDisplayOptions({ density })
    },
    [setDisplayOptions]
  )

  const handleFieldToggle = useCallback(
    (field: string, checked: boolean) => {
      const fields = checked
        ? [...displayOptions.visibleFields, field]
        : displayOptions.visibleFields.filter((f) => f !== field)
      setDisplayOptions({ visibleFields: fields })
    },
    [displayOptions.visibleFields, setDisplayOptions]
  )

  const handleGroupByChange = useCallback(
    (e: React.ChangeEvent<HTMLSelectElement>) => {
      const value = e.target.value as 'column' | 'priority' | 'type' | 'epic'
      setDisplayOptions({ groupBy: value })
    },
    [setDisplayOptions]
  )

  if (!isOpen) return null

  return (
    <div className={styles.overlay}>
      <div ref={menuRef} className={styles.menu}>
        <div className={styles.header}>
          <h3 className={styles.title}>Display options</h3>
        </div>

        {/* View Mode */}
        <div className={styles.section}>
          <label className={styles.sectionLabel}>View</label>
          <div className={styles.buttonGroup}>
            <button
              className={`${styles.toggleButton} ${displayOptions.viewMode === 'board' ? styles.active : ''}`}
              onClick={() => handleViewModeChange('board')}
            >
              <BoardIcon />
              Board
            </button>
            <button
              className={`${styles.toggleButton} ${displayOptions.viewMode === 'list' ? styles.active : ''}`}
              onClick={() => handleViewModeChange('list')}
            >
              <ListIcon />
              List
            </button>
          </div>
        </div>

        {/* Density */}
        <div className={styles.section}>
          <label className={styles.sectionLabel}>Density</label>
          <div className={styles.buttonGroup}>
            <button
              className={`${styles.toggleButton} ${displayOptions.density === 'compact' ? styles.active : ''}`}
              onClick={() => handleDensityChange('compact')}
            >
              Compact
            </button>
            <button
              className={`${styles.toggleButton} ${displayOptions.density === 'comfortable' ? styles.active : ''}`}
              onClick={() => handleDensityChange('comfortable')}
            >
              Comfortable
            </button>
          </div>
        </div>

        {/* Visible Fields */}
        <div className={styles.section}>
          <label className={styles.sectionLabel}>Show on cards</label>
          <div className={styles.checkboxGroup}>
            {VISIBLE_FIELD_OPTIONS.map((option) => (
              <label key={option.key} className={styles.checkboxLabel}>
                <input
                  type="checkbox"
                  className={styles.checkbox}
                  checked={displayOptions.visibleFields.includes(option.key)}
                  onChange={(e) => handleFieldToggle(option.key, e.target.checked)}
                />
                <span className={styles.checkboxText}>{option.label}</span>
              </label>
            ))}
          </div>
        </div>

        {/* Group By (Board mode only) */}
        {displayOptions.viewMode === 'board' && (
          <div className={styles.section}>
            <label className={styles.sectionLabel}>Group by</label>
            <select
              className={styles.select}
              value={displayOptions.groupBy || 'column'}
              onChange={handleGroupByChange}
            >
              {GROUP_BY_OPTIONS.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </div>
        )}

        {/* Keyboard shortcut hint */}
        <div className={styles.hint}>
          <kbd className={styles.kbd}>Cmd</kbd>
          <span>+</span>
          <kbd className={styles.kbd}>B</kbd>
          <span className={styles.hintText}>to toggle view</span>
        </div>
      </div>
    </div>
  )
}

function BoardIcon() {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <rect x="1" y="2" width="4" height="12" rx="1" fill="currentColor" opacity="0.6" />
      <rect x="6" y="2" width="4" height="8" rx="1" fill="currentColor" opacity="0.6" />
      <rect x="11" y="2" width="4" height="10" rx="1" fill="currentColor" opacity="0.6" />
    </svg>
  )
}

function ListIcon() {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <rect x="1" y="3" width="14" height="2" rx="0.5" fill="currentColor" opacity="0.6" />
      <rect x="1" y="7" width="14" height="2" rx="0.5" fill="currentColor" opacity="0.6" />
      <rect x="1" y="11" width="14" height="2" rx="0.5" fill="currentColor" opacity="0.6" />
    </svg>
  )
}
