import type { AgentSession, SessionTool } from '../types/session'
import type { Column } from '../types/task'
import styles from './SessionInfo.module.css'

interface SessionInfoProps {
  session: AgentSession | null | undefined
  taskColumn: Column
  onResume?: () => void
}

// Tool configuration with icons and display names
const toolConfig: Record<SessionTool, { icon: string; name: string }> = {
  opencode: { icon: '\u26A1', name: 'OpenCode' },       // Lightning bolt
  'claude-code': { icon: '\u{1F916}', name: 'Claude Code' }, // Robot
  codex: { icon: '\u{1F52E}', name: 'Codex' },          // Crystal ball
}

/**
 * SessionInfo displays linked agent session information for a task.
 *
 * Features:
 * - Tool icon and name display (opencode, claude-code, codex)
 * - Session reference with truncation
 * - Linked timestamp
 * - Blocked/Active status indicator
 * - Resume button for need_input tasks
 */
export function SessionInfo({ session, taskColumn, onResume }: SessionInfoProps) {
  // Don't render if no session
  if (!session) {
    return null
  }

  const tool = toolConfig[session.tool] || { icon: '\u{1F527}', name: session.tool }
  const linkedAgo = formatTimeAgo(session.linked_at)
  const isBlocked = taskColumn === 'need_input'

  return (
    <section className={styles.sessionInfo} aria-labelledby="session-heading">
      <h4 id="session-heading" className={styles.header}>Agent Session</h4>

      <div className={styles.content}>
        {/* Tool icon */}
        <span className={styles.toolIcon} data-tool={session.tool} aria-hidden="true">
          {tool.icon}
        </span>

        {/* Session details */}
        <div className={styles.details}>
          <div className={styles.toolName}>{tool.name}</div>
          <div className={styles.sessionRef} title={session.ref} aria-label={`Session reference: ${session.ref}`}>
            {truncateMiddle(session.ref, 24)}
          </div>
          <time className={styles.linkedTime} dateTime={session.linked_at}>Linked {linkedAgo}</time>
        </div>

        {/* Status indicator */}
        <div
          className={`${styles.status} ${isBlocked ? styles.blocked : styles.active}`}
          role="status"
          aria-label={`Session status: ${isBlocked ? 'Blocked' : 'Active'}`}
        >
          {isBlocked ? 'Blocked' : 'Active'}
        </div>
      </div>

      {/* Working directory */}
      <div className={styles.workingDir}>
        <span className={styles.workingDirLabel} id="working-dir-label">Working Dir:</span>
        <span className={styles.workingDirPath} title={session.working_dir} aria-labelledby="working-dir-label">
          {truncatePath(session.working_dir, 40)}
        </span>
      </div>

      {/* Resume button */}
      {isBlocked && onResume && (
        <button
          onClick={onResume}
          className={styles.resumeButton}
          aria-label={`Resume ${tool.name} agent session`}
        >
          <PlayIcon className={styles.playIcon} />
          Resume Agent Session
        </button>
      )}
    </section>
  )
}

/**
 * Simple play icon SVG component.
 */
function PlayIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
      <path
        fillRule="evenodd"
        d="M10 18a8 8 0 100-16 8 8 0 000 16zM9.555 7.168A1 1 0 008 8v4a1 1 0 001.555.832l3-2a1 1 0 000-1.664l-3-2z"
        clipRule="evenodd"
      />
    </svg>
  )
}

/**
 * Truncate string in the middle, preserving start and end.
 */
function truncateMiddle(str: string, maxLen: number): string {
  if (str.length <= maxLen) return str
  const half = Math.floor((maxLen - 3) / 2)
  return `${str.slice(0, half)}...${str.slice(-half)}`
}

/**
 * Truncate path from the beginning if too long.
 */
function truncatePath(path: string, maxLen: number): string {
  if (path.length <= maxLen) return path
  return '...' + path.slice(-(maxLen - 3))
}

/**
 * Format timestamp as relative time (e.g., "2h ago").
 */
function formatTimeAgo(timestamp: string): string {
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
    year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined,
  })
}
