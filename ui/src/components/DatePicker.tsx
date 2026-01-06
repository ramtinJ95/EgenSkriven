import { useState, useRef, useEffect } from 'react'
import styles from './DatePicker.module.css'

interface DatePickerProps {
  value: string | null           // ISO date string or null
  onChange: (date: string | null) => void
  placeholder?: string
}

// Days of the week headers
const DAYS = ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa']

// Month names for header
const MONTHS = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December'
]

/**
 * Calendar date picker component.
 * 
 * Features:
 * - Calendar grid with month navigation
 * - Quick shortcuts (Today, Tomorrow, Next week)
 * - Overdue highlighting
 * - Clear button
 * - Click outside to close
 */
export function DatePicker({ value, onChange, placeholder = 'Set due date' }: DatePickerProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [viewDate, setViewDate] = useState(() => {
    // Start calendar on selected date or today
    return value ? new Date(value) : new Date()
  })
  const containerRef = useRef<HTMLDivElement>(null)

  // Close picker when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false)
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
      }
    }
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [isOpen])

  // Update view date when value changes externally
  useEffect(() => {
    if (value) {
      setViewDate(new Date(value))
    }
  }, [value])

  // Generate calendar grid for current view month
  const generateCalendarDays = () => {
    const year = viewDate.getFullYear()
    const month = viewDate.getMonth()
    
    // First day of month and total days
    const firstDay = new Date(year, month, 1).getDay()
    const daysInMonth = new Date(year, month + 1, 0).getDate()
    
    // Previous month days to show
    const prevMonthDays = new Date(year, month, 0).getDate()
    
    const days: { date: Date; isCurrentMonth: boolean; isToday: boolean; isSelected: boolean }[] = []
    
    // Add previous month's trailing days
    for (let i = firstDay - 1; i >= 0; i--) {
      const date = new Date(year, month - 1, prevMonthDays - i)
      days.push({
        date,
        isCurrentMonth: false,
        isToday: false,
        isSelected: false,
      })
    }
    
    // Add current month's days
    const today = new Date()
    const selectedDate = value ? new Date(value) : null
    
    for (let i = 1; i <= daysInMonth; i++) {
      const date = new Date(year, month, i)
      days.push({
        date,
        isCurrentMonth: true,
        isToday: date.toDateString() === today.toDateString(),
        isSelected: selectedDate ? date.toDateString() === selectedDate.toDateString() : false,
      })
    }
    
    // Add next month's leading days to fill grid (6 rows * 7 days = 42)
    const remaining = 42 - days.length
    for (let i = 1; i <= remaining; i++) {
      const date = new Date(year, month + 1, i)
      days.push({
        date,
        isCurrentMonth: false,
        isToday: false,
        isSelected: false,
      })
    }
    
    return days
  }

  // Handle date selection
  const handleSelectDate = (date: Date) => {
    // Format as ISO date string (YYYY-MM-DD) using local date values
    // Note: Using toISOString() would convert to UTC and could shift the date
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    const isoDate = `${year}-${month}-${day}`
    onChange(isoDate)
    setIsOpen(false)
  }

  // Navigate months
  const prevMonth = () => {
    setViewDate(new Date(viewDate.getFullYear(), viewDate.getMonth() - 1, 1))
  }

  const nextMonth = () => {
    setViewDate(new Date(viewDate.getFullYear(), viewDate.getMonth() + 1, 1))
  }

  // Format display value
  const displayValue = value 
    ? new Date(value).toLocaleDateString('en-US', { 
        month: 'short', 
        day: 'numeric',
        year: 'numeric'
      })
    : placeholder

  // Check if date is overdue (before today)
  const isOverdue = value && new Date(value) < new Date(new Date().toDateString())

  const triggerClasses = [
    styles.trigger,
    value ? styles.hasValue : '',
    isOverdue ? styles.overdue : '',
  ].filter(Boolean).join(' ')

  return (
    <div className={styles.datePicker} ref={containerRef}>
      {/* Trigger - using div with role="button" to allow nested clear button */}
      <div 
        className={triggerClasses}
        onClick={() => setIsOpen(!isOpen)}
        role="button"
        tabIndex={0}
        aria-label="Select due date"
        aria-expanded={isOpen}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            setIsOpen(!isOpen)
          }
        }}
      >
        <span className={styles.icon}>&#128197;</span>
        <span className={styles.value}>{displayValue}</span>
        {value && (
          <button 
            className={styles.clear}
            onClick={(e) => {
              e.stopPropagation()
              onChange(null)
            }}
            aria-label="Clear date"
            type="button"
          >
            &times;
          </button>
        )}
      </div>

      {/* Calendar dropdown */}
      {isOpen && (
        <div className={styles.dropdown}>
          {/* Month/year header with navigation */}
          <div className={styles.header}>
            <button 
              onClick={prevMonth} 
              className={styles.navBtn} 
              type="button"
              aria-label="Previous month"
            >
              &lt;
            </button>
            <span className={styles.monthYear}>
              {MONTHS[viewDate.getMonth()]} {viewDate.getFullYear()}
            </span>
            <button 
              onClick={nextMonth} 
              className={styles.navBtn} 
              type="button"
              aria-label="Next month"
            >
              &gt;
            </button>
          </div>

          {/* Day of week headers */}
          <div className={styles.daysHeader}>
            {DAYS.map(day => (
              <span key={day} className={styles.dayHeader}>{day}</span>
            ))}
          </div>

          {/* Calendar grid */}
          <div className={styles.grid}>
            {generateCalendarDays().map((day, index) => {
              const cellClasses = [
                styles.dayCell,
                !day.isCurrentMonth ? styles.otherMonth : '',
                day.isToday ? styles.today : '',
                day.isSelected ? styles.selected : '',
              ].filter(Boolean).join(' ')

              return (
                <button
                  key={index}
                  className={cellClasses}
                  onClick={() => handleSelectDate(day.date)}
                  type="button"
                  aria-label={day.date.toLocaleDateString()}
                >
                  {day.date.getDate()}
                </button>
              )
            })}
          </div>

          {/* Quick select shortcuts */}
          <div className={styles.shortcuts}>
            <button 
              onClick={() => handleSelectDate(new Date())} 
              type="button"
              className={styles.shortcut}
            >
              Today
            </button>
            <button 
              onClick={() => {
                const tomorrow = new Date()
                tomorrow.setDate(tomorrow.getDate() + 1)
                handleSelectDate(tomorrow)
              }} 
              type="button"
              className={styles.shortcut}
            >
              Tomorrow
            </button>
            <button 
              onClick={() => {
                const nextWeek = new Date()
                nextWeek.setDate(nextWeek.getDate() + 7)
                handleSelectDate(nextWeek)
              }} 
              type="button"
              className={styles.shortcut}
            >
              Next week
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
