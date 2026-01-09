import { memo } from 'react'
import { useDraggable } from '@dnd-kit/core'
import type { Task } from '../types/task'
import type { Board } from '../types/board'
import { formatDisplayId } from '../types/board'
import styles from './TaskCard.module.css'

interface TaskCardProps {
  task: Task
  isSelected?: boolean
  onClick?: (task: Task) => void
  onSelect?: (task: Task) => void
  currentBoard?: Board | null
  /** When true, this card is rendered in the DragOverlay (the moving clone) */
  isDragOverlay?: boolean
}

/**
 * A draggable task card.
 * 
 * Displays:
 * - Status dot and task ID (display ID if board available, otherwise short ID)
 * - Title (truncated to 2 lines)
 * - Labels (if any)
 * - Priority indicator
 * - Due date (if set)
 * 
 * Clicking opens the task detail panel.
 * 
 * Memoized to prevent unnecessary re-renders when other tasks change.
 */
function TaskCardComponent({ task, isSelected = false, onClick, onSelect, currentBoard, isDragOverlay = false }: TaskCardProps) {
  // Make this card draggable (skip if this is the drag overlay)
  const { attributes, listeners, setNodeRef, isDragging: isCurrentlyDragging } = useDraggable({
    id: task.id,
    disabled: isDragOverlay,
  })

  // When dragging, reduce opacity (the DragOverlay shows the visual)
  const style: React.CSSProperties | undefined = isCurrentlyDragging
    ? { opacity: 0.3 }
    : undefined

  // Priority indicator
  const getPriorityIndicator = (priority: string) => {
    switch (priority) {
      case 'urgent':
        return { color: 'var(--priority-urgent)', label: 'Urgent' }
      case 'high':
        return { color: 'var(--priority-high)', label: 'High' }
      case 'medium':
        return { color: 'var(--priority-medium)', label: 'Medium' }
      case 'low':
        return { color: 'var(--priority-low)', label: 'Low' }
      default:
        return null
    }
  }

  const priority = getPriorityIndicator(task.priority)

  // Check if task is overdue (due date in the past and not done)
  const isOverdue = (() => {
    if (!task.due_date || task.column === 'done') return false
    // Compare timestamps at start of day without mutating Date objects
    const dueDateStart = new Date(task.due_date).setHours(0, 0, 0, 0)
    const todayStart = new Date().setHours(0, 0, 0, 0)
    return dueDateStart < todayStart
  })()

  // Check if task is due today
  const isDueToday = (() => {
    if (!task.due_date || task.column === 'done') return false
    const dueDate = new Date(task.due_date)
    const today = new Date()
    return (
      dueDate.getFullYear() === today.getFullYear() &&
      dueDate.getMonth() === today.getMonth() &&
      dueDate.getDate() === today.getDate()
    )
  })()

  // Handle click to open detail panel
  // Only trigger if not dragging (to avoid opening panel after drag)
  const handleClick = () => {
    if (!isCurrentlyDragging && onClick) {
      onClick(task)
    }
  }

  // Handle focus/selection (for keyboard navigation)
  const handleFocus = () => {
    if (onSelect) {
      onSelect(task)
    }
  }

  // Check if task needs input (blocked)
  const needsInput = task.column === 'need_input'

  // Build class names - include global 'task-card' for drag-drop.css styles
  const classNames = [
    'task-card', // Global class for drag-drop.css
    styles.card,
    isSelected && styles.selected,
    isCurrentlyDragging && 'dragging',
    isDragOverlay && 'drag-overlay',
    needsInput && styles.needsInput,
  ].filter(Boolean).join(' ')

  return (
    <div
      ref={isDragOverlay ? undefined : setNodeRef}
      {...(isDragOverlay ? {} : listeners)}
      {...(isDragOverlay ? {} : attributes)}
      style={style}
      className={classNames}
      data-task-id={task.id}
      onClick={handleClick}
      onFocus={handleFocus}
      tabIndex={isDragOverlay ? -1 : 0}
      role="button"
      aria-pressed={isSelected}
    >
      {/* Header: Status dot + ID + Needs Input badge */}
      <div className={styles.header}>
        <span
          className={`${styles.statusDot} ${needsInput ? styles.pulsingDot : ''}`}
          style={{
            backgroundColor: `var(--status-${task.column.replace('_', '-')})`,
          }}
        />
        <span className={styles.id}>
          {currentBoard && task.seq
            ? formatDisplayId(currentBoard.prefix, task.seq)
            : task.id.slice(0, 8)}
        </span>
        {needsInput && (
          <span className={styles.needsInputBadge}>Needs Input</span>
        )}
      </div>

      {/* Title */}
      <h3 className={styles.title}>{task.title}</h3>

      {/* Labels */}
      {task.labels && task.labels.length > 0 && (
        <div className={styles.labels}>
          {task.labels.map((label) => (
            <span key={label} className={styles.label}>
              {label}
            </span>
          ))}
        </div>
      )}

      {/* Footer: Priority + Due Date + Type + Description indicator */}
      <div className={styles.footer}>
        {priority && (
          <span className={styles.priority} title={priority.label} style={{ color: priority.color }}>
            {priority.label}
          </span>
        )}
        {task.due_date && (
          <span 
            className={`${styles.dueDate} ${isOverdue ? styles.overdue : ''} ${isDueToday ? styles.dueToday : ''}`}
            title={isOverdue ? 'Overdue' : isDueToday ? 'Due today' : undefined}
          >
            {isOverdue && <OverdueIcon />}
            {new Date(task.due_date).toLocaleDateString('en-US', { 
              month: 'short', 
              day: 'numeric' 
            })}
          </span>
        )}
        {task.description && (
          <span className={styles.descriptionIndicator} title="Has description">
            <DescriptionIcon />
          </span>
        )}
        <span
          className={styles.type}
          style={{ color: `var(--type-${task.type})` }}
        >
          {task.type}
        </span>
      </div>
    </div>
  )
}

// Small warning icon for overdue tasks
function OverdueIcon() {
  return (
    <svg
      width="12"
      height="12"
      viewBox="0 0 16 16"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      style={{ marginRight: '4px' }}
    >
      <path
        d="M8 5v4M8 11h.01"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
      />
      <circle
        cx="8"
        cy="8"
        r="6"
        stroke="currentColor"
        strokeWidth="1.5"
      />
    </svg>
  )
}

// Small icon to indicate task has a description
function DescriptionIcon() {
  return (
    <svg
      width="14"
      height="14"
      viewBox="0 0 16 16"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M2 4h12M2 8h8M2 12h10"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
      />
    </svg>
  )
}

// Memoize to prevent unnecessary re-renders when other tasks change
// Uses default shallow comparison which handles all props including callbacks
export const TaskCard = memo(TaskCardComponent)
