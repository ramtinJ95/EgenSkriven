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
 */
export function TaskCard({ task, isSelected = false, onClick, onSelect, currentBoard }: TaskCardProps) {
  // Make this card draggable
  const { attributes, listeners, setNodeRef, isDragging: isCurrentlyDragging } = useDraggable({
    id: task.id,
  })

  // When dragging, hide this card (the DragOverlay shows the visual)
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
      className={`${styles.card} ${isSelected ? styles.selected : ''}`}
      data-task-id={task.id}
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
        <span className={styles.id}>
          {currentBoard && task.seq
            ? formatDisplayId(currentBoard.prefix, task.seq)
            : task.id.slice(0, 8)}
        </span>
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
