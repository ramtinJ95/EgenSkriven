import { type HistoryEntry } from '../types/task'
import styles from './ActivityLog.module.css'

interface ActivityLogProps {
  history: HistoryEntry[]
  created: string
}

/**
 * ActivityLog displays the history of changes to a task.
 * 
 * Features:
 * - Shows all history entries sorted newest first
 * - Relative timestamps (e.g., "2h ago", "3d ago")
 * - Actor icons (user, agent, CLI)
 * - Human-readable action descriptions
 * - Always shows task creation at the bottom
 */
export function ActivityLog({ history, created }: ActivityLogProps) {
  // Sort history newest first
  const sortedHistory = [...(history || [])].sort(
    (a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
  )

  // Format timestamp relative to now
  const formatTime = (timestamp: string): string => {
    const date = new Date(timestamp)
    if (isNaN(date.getTime())) return '-'
    
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMins = Math.floor(diffMs / 60000)
    const diffHours = Math.floor(diffMs / 3600000)
    const diffDays = Math.floor(diffMs / 86400000)

    if (diffMins < 1) return 'just now'
    if (diffMins < 60) return `${diffMins}m ago`
    if (diffHours < 24) return `${diffHours}h ago`
    if (diffDays < 7) return `${diffDays}d ago`
    
    return date.toLocaleDateString('en-US', { 
      month: 'short', 
      day: 'numeric',
      year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined
    })
  }

  // Format actor display
  const formatActor = (actor: string, actorDetail?: string): string => {
    if (actorDetail) {
      return actorDetail // e.g., "claude", "opencode"
    }
    switch (actor) {
      case 'agent': return 'Agent'
      case 'cli': return 'CLI'
      case 'user': return 'You'
      default: return actor
    }
  }

  // Format action description
  const formatAction = (entry: HistoryEntry): string => {
    switch (entry.action) {
      case 'created':
        return 'created this task'
      case 'moved':
        if (entry.changes?.field === 'column') {
          return `moved to ${formatColumnName(String(entry.changes.to))}`
        }
        return 'moved this task'
      case 'updated':
        if (entry.changes) {
          const { field, from, to } = entry.changes
          if (field === 'priority') {
            return `changed priority from ${from} to ${to}`
          }
          if (field === 'title') {
            return `renamed to "${to}"`
          }
          if (field === 'description') {
            return 'updated description'
          }
          if (field === 'due_date') {
            if (!from && to) return `set due date to ${formatDate(String(to))}`
            if (from && !to) return 'removed due date'
            return `changed due date to ${formatDate(String(to))}`
          }
          if (field === 'epic') {
            if (!from && to) return 'assigned to epic'
            if (from && !to) return 'removed from epic'
            return 'changed epic'
          }
          if (field === 'type') {
            return `changed type from ${from} to ${to}`
          }
          return `updated ${field}`
        }
        return 'updated this task'
      case 'completed':
        return 'marked as done'
      case 'deleted':
        return 'deleted this task'
      default:
        return entry.action
    }
  }

  // Helper to format column names
  const formatColumnName = (column: string): string => {
    const names: Record<string, string> = {
      backlog: 'Backlog',
      todo: 'Todo',
      in_progress: 'In Progress',
      review: 'Review',
      done: 'Done'
    }
    return names[column] || column
  }

  // Helper to format dates
  const formatDate = (dateStr: string): string => {
    const date = new Date(dateStr)
    if (isNaN(date.getTime())) return dateStr
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
  }

  // Get icon for actor type
  const getActorIcon = (actor: string): string => {
    switch (actor) {
      case 'agent': return '\u{1F916}' // Robot emoji
      case 'cli': return '\u{1F4BB}'   // Computer emoji  
      case 'user': return '\u{1F464}'  // Person emoji
      default: return '\u{2022}'       // Bullet
    }
  }

  // Don't show if no history and no created date
  if ((!history || history.length === 0) && !created) {
    return null
  }

  return (
    <div className={styles.activityLog}>
      <h3 className={styles.header}>Activity</h3>
      
      <div className={styles.list}>
        {sortedHistory.map((entry, index) => (
          <div key={index} className={styles.item}>
            <span className={styles.icon}>
              {getActorIcon(entry.actor)}
            </span>
            <div className={styles.content}>
              <span className={styles.actor}>
                {formatActor(entry.actor, entry.actor_detail)}
              </span>{' '}
              <span className={styles.action}>
                {formatAction(entry)}
              </span>
            </div>
            <span className={styles.time}>
              {formatTime(entry.timestamp)}
            </span>
          </div>
        ))}

        {/* Always show created entry at bottom */}
        {created && (
          <div className={styles.item}>
            <span className={styles.icon}>{'\u{2728}'}</span>
            <div className={styles.content}>
              <span className={styles.action}>Task created</span>
            </div>
            <span className={styles.time}>
              {formatTime(created)}
            </span>
          </div>
        )}
      </div>
    </div>
  )
}
