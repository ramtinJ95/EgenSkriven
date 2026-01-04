import { useDraggable } from '@dnd-kit/core'
import { CSS } from '@dnd-kit/utilities'
import type { Task } from '../types/task'
import styles from './TaskCard.module.css'

interface TaskCardProps {
  task: Task
  isDragging?: boolean
  isSelected?: boolean
  onClick?: (task: Task) => void
  onSelect?: (task: Task) => void
}

/**
 * A draggable task card.
 * 
 * Displays:
 * - Status dot and task ID
 * - Title (truncated to 2 lines)
 * - Labels (if any)
 * - Priority indicator
 * - Due date (if set)
 * 
 * Clicking opens the task detail panel.
 */
export function TaskCard({ task, isDragging = false, isSelected = false, onClick, onSelect }: TaskCardProps) {
  // Make this card draggable
  const { attributes, listeners, setNodeRef, transform, isDragging: isCurrentlyDragging } = useDraggable({
    id: task.id,
  })

  // Apply transform during drag
  const style = transform
    ? {
        transform: CSS.Transform.toString(transform),
      }
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

  return (
    <div
      ref={setNodeRef}
      {...listeners}
      {...attributes}
      style={style}
      className={`${styles.card} ${isDragging ? styles.dragging : ''} ${isSelected ? styles.selected : ''}`}
      onClick={handleClick}
      onFocus={handleFocus}
      tabIndex={0}
      role="button"
      aria-pressed={isSelected}
    >
      {/* Header: Status dot + ID */}
      <div className={styles.header}>
        <span
          className={styles.statusDot}
          style={{
            backgroundColor: `var(--status-${task.column.replace('_', '-')})`,
          }}
        />
        <span className={styles.id}>{task.id.slice(0, 8)}</span>
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

      {/* Footer: Priority + Due Date + Type */}
      <div className={styles.footer}>
        {priority && (
          <span className={styles.priority} title={priority.label} style={{ color: priority.color }}>
            {priority.label}
          </span>
        )}
        {task.due_date && (
          <span className={styles.dueDate}>
            {new Date(task.due_date).toLocaleDateString('en-US', { 
              month: 'short', 
              day: 'numeric' 
            })}
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
