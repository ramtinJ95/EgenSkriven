import { COLUMN_NAMES, type Column } from '../types/task'
import styles from './StatusBadge.module.css'

interface StatusBadgeProps {
  status: Column
  /** Show as compact dot-only version */
  compact?: boolean
}

/**
 * StatusBadge displays the task status with a colored dot.
 *
 * Supports all column statuses including need_input for blocked tasks.
 *
 * @example
 * ```tsx
 * <StatusBadge status="in_progress" />
 * <StatusBadge status="need_input" compact />
 * ```
 */
export function StatusBadge({ status, compact = false }: StatusBadgeProps) {
  if (compact) {
    return (
      <span
        className={styles.dot}
        style={{ backgroundColor: `var(--status-${status.replace('_', '-')})` }}
        title={COLUMN_NAMES[status] || status.replace('_', ' ')}
      />
    )
  }

  return (
    <span className={`${styles.badge} ${styles[`status_${status}`]}`}>
      <span
        className={styles.dot}
        style={{ backgroundColor: `var(--status-${status.replace('_', '-')})` }}
      />
      {COLUMN_NAMES[status] || status.replace('_', ' ')}
    </span>
  )
}
