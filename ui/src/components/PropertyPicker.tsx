import {
  useState,
  useEffect,
  useRef,
  useCallback,
  type ReactNode,
} from 'react'
import { createPortal } from 'react-dom'
import styles from './PropertyPicker.module.css'

export interface PropertyOption<T> {
  value: T
  label: string
  icon?: ReactNode
  color?: string
}

interface PropertyPickerProps<T> {
  isOpen: boolean
  onClose: () => void
  onSelect: (value: T) => void
  options: PropertyOption<T>[]
  currentValue?: T
  title: string
  // Position the picker near this element
  anchorElement?: HTMLElement | null
}

export function PropertyPicker<T extends string>({
  isOpen,
  onClose,
  onSelect,
  options,
  currentValue,
  title,
  anchorElement,
}: PropertyPickerProps<T>) {
  const [selectedIndex, setSelectedIndex] = useState(0)
  const [query, setQuery] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)
  const pickerRef = useRef<HTMLDivElement>(null)

  // Filter options by query
  const filteredOptions = query
    ? options.filter((opt) =>
        opt.label.toLowerCase().includes(query.toLowerCase())
      )
    : options

  // Reset on open
  useEffect(() => {
    if (isOpen) {
      setQuery('')
      // Set initial selection to current value
      const currentIndex = options.findIndex(
        (opt) => opt.value === currentValue
      )
      setSelectedIndex(currentIndex >= 0 ? currentIndex : 0)
      setTimeout(() => inputRef.current?.focus(), 0)
    }
  }, [isOpen, currentValue, options])

  // Keep selection in bounds
  useEffect(() => {
    if (selectedIndex >= filteredOptions.length) {
      setSelectedIndex(Math.max(0, filteredOptions.length - 1))
    }
  }, [filteredOptions.length, selectedIndex])

  // Position the picker near the anchor element
  useEffect(() => {
    if (!isOpen || !anchorElement || !pickerRef.current) return

    const anchorRect = anchorElement.getBoundingClientRect()
    const picker = pickerRef.current

    // Position below the anchor, centered
    picker.style.top = `${anchorRect.bottom + 8}px`
    picker.style.left = `${
      anchorRect.left + anchorRect.width / 2 - picker.offsetWidth / 2
    }px`

    // Ensure it stays on screen
    const pickerRect = picker.getBoundingClientRect()
    if (pickerRect.right > window.innerWidth - 16) {
      picker.style.left = `${window.innerWidth - pickerRect.width - 16}px`
    }
    if (pickerRect.left < 16) {
      picker.style.left = '16px'
    }
    // If goes below viewport, position above anchor
    if (pickerRect.bottom > window.innerHeight - 16) {
      picker.style.top = `${anchorRect.top - pickerRect.height - 8}px`
    }
  }, [isOpen, anchorElement])

  const handleSelect = useCallback(
    (option: PropertyOption<T>) => {
      onSelect(option.value)
      onClose()
    },
    [onSelect, onClose]
  )

  const handleKeyDown = useCallback(
    (event: React.KeyboardEvent) => {
      switch (event.key) {
        case 'ArrowDown':
          event.preventDefault()
          setSelectedIndex((i) => Math.min(i + 1, filteredOptions.length - 1))
          break

        case 'ArrowUp':
          event.preventDefault()
          setSelectedIndex((i) => Math.max(i - 1, 0))
          break

        case 'Enter':
          event.preventDefault()
          if (filteredOptions[selectedIndex]) {
            handleSelect(filteredOptions[selectedIndex])
          }
          break

        case 'Escape':
          event.preventDefault()
          onClose()
          break
      }
    },
    [filteredOptions, selectedIndex, handleSelect, onClose]
  )

  if (!isOpen) return null

  const picker = (
    <div className={styles.overlay} onClick={onClose}>
      <div
        ref={pickerRef}
        className={styles.picker}
        onClick={(e) => e.stopPropagation()}
        onKeyDown={handleKeyDown}
      >
        <div className={styles.header}>
          <span className={styles.title}>{title}</span>
        </div>

        <div className={styles.inputWrapper}>
          <input
            ref={inputRef}
            type="text"
            className={styles.input}
            placeholder="Filter..."
            value={query}
            onChange={(e) => {
              setQuery(e.target.value)
              setSelectedIndex(0)
            }}
          />
        </div>

        <div className={styles.options}>
          {filteredOptions.length === 0 ? (
            <div className={styles.empty}>No options found</div>
          ) : (
            filteredOptions.map((option, index) => (
              <button
                key={option.value}
                className={`${styles.option} ${
                  index === selectedIndex ? styles.selected : ''
                }`}
                onClick={() => handleSelect(option)}
                onMouseEnter={() => setSelectedIndex(index)}
              >
                {option.icon && (
                  <span
                    className={styles.icon}
                    style={option.color ? { color: option.color } : undefined}
                  >
                    {option.icon}
                  </span>
                )}
                <span className={styles.label}>{option.label}</span>
                {option.value === currentValue && (
                  <span className={styles.check}>✓</span>
                )}
              </button>
            ))
          )}
        </div>
      </div>
    </div>
  )

  return createPortal(picker, document.body)
}

// Pre-configured pickers for common properties

export const STATUS_OPTIONS: PropertyOption<string>[] = [
  {
    value: 'backlog',
    label: 'Backlog',
    icon: '●',
    color: 'var(--status-backlog, #6B7280)',
  },
  {
    value: 'todo',
    label: 'Todo',
    icon: '●',
    color: 'var(--status-todo, #E5E5E5)',
  },
  {
    value: 'in_progress',
    label: 'In Progress',
    icon: '●',
    color: 'var(--status-in-progress, #F59E0B)',
  },
  {
    value: 'review',
    label: 'Review',
    icon: '●',
    color: 'var(--status-review, #A855F7)',
  },
  {
    value: 'done',
    label: 'Done',
    icon: '●',
    color: 'var(--status-done, #22C55E)',
  },
]

export const PRIORITY_OPTIONS: PropertyOption<string>[] = [
  {
    value: 'urgent',
    label: 'Urgent',
    icon: '🔴',
    color: 'var(--priority-urgent, #EF4444)',
  },
  {
    value: 'high',
    label: 'High',
    icon: '🟠',
    color: 'var(--priority-high, #F97316)',
  },
  {
    value: 'medium',
    label: 'Medium',
    icon: '🟡',
    color: 'var(--priority-medium, #EAB308)',
  },
  {
    value: 'low',
    label: 'Low',
    icon: '⚪',
    color: 'var(--priority-low, #6B7280)',
  },
]

export const TYPE_OPTIONS: PropertyOption<string>[] = [
  {
    value: 'bug',
    label: 'Bug',
    icon: '🐛',
    color: 'var(--type-bug, #EF4444)',
  },
  {
    value: 'feature',
    label: 'Feature',
    icon: '✨',
    color: 'var(--type-feature, #A855F7)',
  },
  {
    value: 'chore',
    label: 'Chore',
    icon: '🔧',
    color: 'var(--type-chore, #6B7280)',
  },
]
