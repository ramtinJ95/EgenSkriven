import { createPortal } from 'react-dom'
import { type Task, COLUMN_NAMES } from '../types/task'
import styles from './PeekPreview.module.css'

interface PeekPreviewProps {
  task: Task | null
  isOpen: boolean
  onClose: () => void
}

export function PeekPreview({ task, isOpen, onClose }: PeekPreviewProps) {
  if (!isOpen || !task) return null

  const preview = (
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.preview} onClick={(e) => e.stopPropagation()}>
        {/* Header with task ID and type */}
        <div className={styles.header}>
          <span className={styles.taskId}>{task.id.substring(0, 8)}</span>
          <span className={`${styles.type} ${styles[task.type]}`}>
            {task.type}
          </span>
        </div>

        {/* Title */}
        <h2 className={styles.title}>{task.title}</h2>

        {/* Properties row */}
        <div className={styles.properties}>
          <div className={styles.property}>
            <span className={styles.propertyLabel}>Status</span>
            <span className={`${styles.status} ${styles[task.column]}`}>
              {COLUMN_NAMES[task.column]}
            </span>
          </div>

          <div className={styles.property}>
            <span className={styles.propertyLabel}>Priority</span>
            <span className={`${styles.priority} ${styles[task.priority]}`}>
              {task.priority}
            </span>
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
        </div>

        {/* Description preview */}
        {task.description && (
          <div className={styles.description}>
            <span className={styles.propertyLabel}>Description</span>
            <p className={styles.descriptionText}>
              {task.description.length > 200
                ? task.description.substring(0, 200) + '...'
                : task.description}
            </p>
          </div>
        )}

        {/* Footer hint */}
        <div className={styles.footer}>
          <span className={styles.hint}>
            Press <kbd>Enter</kbd> to open full details
          </span>
        </div>
      </div>
    </div>
  )

  return createPortal(preview, document.body)
}
