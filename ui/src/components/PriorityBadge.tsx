import { PRIORITY_NAMES, type Priority } from '../types/task'
import styles from './PriorityBadge.module.css'

interface PriorityBadgeProps {
  priority: Priority
  /** Show icon instead of text */
  iconOnly?: boolean
}

// Priority icons (simple arrows/symbols)
const priorityIcons: Record<Priority, string> = {
  urgent: '\u26A0\uFE0F', // Warning sign
  high: '\u2191',        // Up arrow
  medium: '\u2212',      // Minus
  low: '\u2193',         // Down arrow
}

/**
 * PriorityBadge displays the task priority with appropriate color.
 *
 * @example
 * ```tsx
 * <PriorityBadge priority="urgent" />
 * <PriorityBadge priority="high" iconOnly />
 * ```
 */
export function PriorityBadge({ priority, iconOnly = false }: PriorityBadgeProps) {
  const colorVar = `var(--priority-${priority})`

  if (iconOnly) {
    return (
      <span
        className={styles.icon}
        style={{ color: colorVar }}
        title={PRIORITY_NAMES[priority] || priority}
      >
        {priorityIcons[priority]}
      </span>
    )
  }

  return (
    <span className={styles.badge} style={{ color: colorVar }}>
      {PRIORITY_NAMES[priority] || priority}
    </span>
  )
}
