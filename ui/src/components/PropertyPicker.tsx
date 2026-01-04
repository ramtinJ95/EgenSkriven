import {
  useState,
  useEffect,
  useRef,
  useCallback,
  useMemo,
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

/**
 * Inner component that handles the picker logic.
 * This is remounted when the picker opens, which resets all state automatically.
 */
function PropertyPickerInner<T extends string>({
  onClose,
  onSelect,
  options,
  currentValue,
  title,
  anchorElement,
}: Omit<PropertyPickerProps<T>, 'isOpen'>) {
  // Calculate initial selected index based on current value
  const initialIndex = useMemo(() => {
    const idx = options.findIndex((opt) => opt.value === currentValue)
    return idx >= 0 ? idx : 0
  }, [options, currentValue])

  const [selectedIndex, setSelectedIndex] = useState(initialIndex)
  const [query, setQuery] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)
  const pickerRef = useRef<HTMLDivElement>(null)

  // Filter options by query
  const filteredOptions = query
    ? options.filter((opt) =>
        opt.label.toLowerCase().includes(query.toLowerCase())
      )
    : options

  // Derive bounded selectedIndex to keep in bounds
  const boundedSelectedIndex = Math.min(
    selectedIndex,
    Math.max(0, filteredOptions.length - 1)
  )

  // Focus input on mount
  useEffect(() => {
    const timer = setTimeout(() => inputRef.current?.focus(), 0)
    return () => clearTimeout(timer)
  }, [])

  // Position the picker near the anchor element
  useEffect(() => {
    if (!anchorElement || !pickerRef.current) return

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
  }, [anchorElement])

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
          if (filteredOptions[boundedSelectedIndex]) {
            handleSelect(filteredOptions[boundedSelectedIndex])
          }
          break

        case 'Escape':
          event.preventDefault()
          onClose()
          break
      }
    },
    [filteredOptions, boundedSelectedIndex, handleSelect, onClose]
  )

  return (
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
                  index === boundedSelectedIndex ? styles.selected : ''
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
}

/**
 * PropertyPicker component - renders a picker for selecting property values.
 * Uses portal to render at document body level.
 */
export function PropertyPicker<T extends string>(props: PropertyPickerProps<T>) {
  const { isOpen, ...rest } = props

  if (!isOpen) return null

  // By conditionally rendering the inner component, we get automatic state reset
  // when the picker opens (component remounts with fresh state)
  const picker = <PropertyPickerInner {...rest} />

  return createPortal(picker, document.body)
}

// Re-export options from separate file for backwards compatibility
export { STATUS_OPTIONS, PRIORITY_OPTIONS, TYPE_OPTIONS } from './propertyOptions'
