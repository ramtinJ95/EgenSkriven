import { useEffect, useRef } from 'react'
import { 
  COLUMN_NAMES, 
  PRIORITY_NAMES, 
  TYPE_NAMES,
  COLUMNS,
  PRIORITIES,
  TYPES,
  type Task, 
  type Column,
  type Priority,
  type TaskType,
} from '../types/task'
import styles from './TaskDetail.module.css'

interface TaskDetailProps {
  task: Task | null
  onClose: () => void
  onUpdate: (id: string, data: Partial<Task>) => Promise<void>
}

/**
 * Slide-in panel showing full task details.
 * 
 * Features:
 * - Close with Esc or click outside
 * - Editable properties via dropdowns (status, priority, type)
 * - Display all task metadata
 */
export function TaskDetail({ task, onClose, onUpdate }: TaskDetailProps) {
  const panelRef = useRef<HTMLDivElement>(null)

  // Handle keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && task) {
        onClose()
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [task, onClose])

  // Handle click outside
  const handleOverlayClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose()
    }
  }

  if (!task) return null

  const handleColumnChange = async (newColumn: Column) => {
    await onUpdate(task.id, { column: newColumn })
  }

  const handlePriorityChange = async (newPriority: Priority) => {
    await onUpdate(task.id, { priority: newPriority })
  }

  const handleTypeChange = async (newType: TaskType) => {
    await onUpdate(task.id, { type: newType })
  }

  // Get priority color
  const getPriorityColor = (priority: Priority): string => {
    switch (priority) {
      case 'urgent': return 'var(--priority-urgent)'
      case 'high': return 'var(--priority-high)'
      case 'medium': return 'var(--priority-medium)'
      case 'low': return 'var(--priority-low)'
      default: return ''
    }
  }

  return (
    <div className={styles.overlay} onClick={handleOverlayClick}>
      <div ref={panelRef} className={styles.panel}>
        {/* Header */}
        <div className={styles.header}>
          <button onClick={onClose} className={styles.closeButton}>
            Close
          </button>
        </div>

        {/* Content */}
        <div className={styles.content}>
          {/* Title */}
          <h1 className={styles.title}>{task.title}</h1>
          <span className={styles.id}>{task.id}</span>

          {/* Description */}
          {task.description && (
            <div className={styles.description}>
              <h3 className={styles.sectionTitle}>Description</h3>
              <p>{task.description}</p>
            </div>
          )}

          {/* Properties */}
          <div className={styles.properties}>
            <div className={styles.property}>
              <span className={styles.propertyLabel}>Status</span>
              <select
                value={task.column}
                onChange={(e) => handleColumnChange(e.target.value as Column)}
                className={styles.propertySelect}
              >
                {COLUMNS.map((col) => (
                  <option key={col} value={col}>
                    {COLUMN_NAMES[col]}
                  </option>
                ))}
              </select>
            </div>

            <div className={styles.property}>
              <span className={styles.propertyLabel}>Type</span>
              <select
                value={task.type}
                onChange={(e) => handleTypeChange(e.target.value as TaskType)}
                className={styles.propertySelect}
              >
                {TYPES.map((type) => (
                  <option key={type} value={type}>
                    {TYPE_NAMES[type]}
                  </option>
                ))}
              </select>
            </div>

            <div className={styles.property}>
              <span className={styles.propertyLabel}>Priority</span>
              <select
                value={task.priority}
                onChange={(e) => handlePriorityChange(e.target.value as Priority)}
                className={styles.propertySelect}
                style={{ color: getPriorityColor(task.priority) }}
              >
                {PRIORITIES.map((priority) => (
                  <option key={priority} value={priority}>
                    {PRIORITY_NAMES[priority]}
                  </option>
                ))}
              </select>
            </div>

            {task.labels && task.labels.length > 0 && (
              <div className={styles.property}>
                <span className={styles.propertyLabel}>Labels</span>
                <div className={styles.labels}>
                  {task.labels.map((label) => (
                    <span key={label} className={styles.label}>
                      {label}
                    </span>
                  ))}
                </div>
              </div>
            )}

            {task.due_date && (
              <div className={styles.property}>
                <span className={styles.propertyLabel}>Due Date</span>
                <span className={styles.propertyValue}>
                  {new Date(task.due_date).toLocaleDateString()}
                </span>
              </div>
            )}

            {task.blocked_by && task.blocked_by.length > 0 && (
              <div className={styles.property}>
                <span className={styles.propertyLabel}>Blocked By</span>
                <div className={styles.blockedList}>
                  {task.blocked_by.map((id) => (
                    <span key={id} className={styles.blockedId}>
                      {id.slice(0, 8)}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>

          {/* Metadata */}
          <div className={styles.metadata}>
            <div className={styles.metaItem}>
              <span className={styles.metaLabel}>Created</span>
              <span className={styles.metaValue}>
                {new Date(task.created).toLocaleDateString()}
              </span>
            </div>
            <div className={styles.metaItem}>
              <span className={styles.metaLabel}>Updated</span>
              <span className={styles.metaValue}>
                {new Date(task.updated).toLocaleDateString()}
              </span>
            </div>
            {task.created_by && (
              <div className={styles.metaItem}>
                <span className={styles.metaLabel}>Created By</span>
                <span className={styles.metaValue}>
                  {task.created_by}
                  {task.created_by_agent && ` (${task.created_by_agent})`}
                </span>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
