import { useEffect, useRef, useCallback, useState } from 'react'
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
import type { Board } from '../types/board'
import { formatDisplayId } from '../types/board'
import { DatePicker } from './DatePicker'
import { EpicPicker } from './EpicPicker'
import { SubtaskList } from './SubtaskList'
import { MarkdownEditor } from './MarkdownEditor'
import { ActivityLog } from './ActivityLog'
import { CommentsPanel } from './CommentsPanel'
import { SessionInfo } from './SessionInfo'
import { ResumeModal } from './ResumeModal'
import styles from './TaskDetail.module.css'

interface TaskDetailProps {
  task: Task | null
  tasks: Task[]
  onClose: () => void
  onUpdate: (id: string, data: Partial<Task>) => Promise<void>
  onTaskClick?: (task: Task) => void
  currentBoard?: Board | null
}

/**
 * Slide-in panel showing full task details.
 *
 * Features:
 * - Close with Esc or click outside
 * - Editable properties via dropdowns (status, priority, type)
 * - Editable description with Markdown rendering
 * - Display all task metadata
 * - Comments panel for agent/human conversation
 * - Session info for linked agent sessions
 * - Resume modal for blocked tasks
 */
export function TaskDetail({ task, tasks, onClose, onUpdate, onTaskClick, currentBoard }: TaskDetailProps) {
  const panelRef = useRef<HTMLDivElement>(null)
  const [isResumeModalOpen, setIsResumeModalOpen] = useState(false)

  // Handle keyboard shortcuts
  useEffect(() => {
    // Don't add listener when no task is selected
    if (!task) return

    const handleKeyDown = (e: KeyboardEvent) => {
      // Only close on Escape if not inside the MarkdownEditor (which handles its own Escape)
      if (e.key === 'Escape') {
        // Check if we're inside an active editor
        const activeElement = document.activeElement
        const isInEditor = activeElement?.closest('[class*="MarkdownEditor"]')
        if (!isInEditor) {
          onClose()
        }
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

  // Description change handler
  const handleDescriptionChange = useCallback(async (newDescription: string) => {
    if (!task) return
    await onUpdate(task.id, { description: newDescription || undefined })
  }, [task, onUpdate])

  // Early return after all hooks are defined
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

  const handleDueDateChange = async (newDueDate: string | null) => {
    await onUpdate(task.id, { due_date: newDueDate || undefined })
  }

  const handleEpicChange = async (newEpicId: string | null) => {
    await onUpdate(task.id, { epic: newEpicId || undefined })
  }

  // Toggle subtask completion (moves between todo and done)
  const handleToggleSubtaskComplete = async (subtask: Task) => {
    const newColumn = subtask.column === 'done' ? 'todo' : 'done'
    await onUpdate(subtask.id, { column: newColumn })
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

  // Safely format a date string, returning fallback if invalid
  const formatDate = (dateStr: string | undefined): string => {
    if (!dateStr) return '-'
    const date = new Date(dateStr)
    // Check if date is valid and not the zero date (0001-01-01)
    if (isNaN(date.getTime()) || date.getFullYear() < 1970) {
      return '-'
    }
    return date.toLocaleDateString()
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

          {/* Description - editable with Markdown support */}
          <div className={styles.description}>
            <h3 className={styles.sectionTitle}>Description</h3>
            <MarkdownEditor
              value={task.description || ''}
              onChange={handleDescriptionChange}
              placeholder="Click to add description..."
              maxLength={10000}
            />
          </div>

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

            <div className={styles.property}>
              <span className={styles.propertyLabel}>Due Date</span>
              <DatePicker
                value={task.due_date || null}
                onChange={handleDueDateChange}
                placeholder="Set due date"
              />
            </div>

            <div className={styles.property}>
              <span className={styles.propertyLabel}>Epic</span>
              <EpicPicker
                boardId={task.board}
                value={task.epic || null}
                onChange={handleEpicChange}
                placeholder="Set epic"
              />
            </div>

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

          {/* Sub-tasks */}
          <SubtaskList
            parentId={task.id}
            tasks={tasks}
            onTaskClick={onTaskClick}
            onToggleComplete={handleToggleSubtaskComplete}
          />

          {/* Metadata */}
          <div className={styles.metadata}>
            <div className={styles.metaItem}>
              <span className={styles.metaLabel}>Created</span>
              <span className={styles.metaValue}>
                {formatDate(task.created)}
              </span>
            </div>
            <div className={styles.metaItem}>
              <span className={styles.metaLabel}>Updated</span>
              <span className={styles.metaValue}>
                {formatDate(task.updated)}
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

          {/* Activity Log */}
          <ActivityLog
            history={task.history || []}
            created={task.created}
          />

          {/* Agent Session Info */}
          <SessionInfo
            session={task.agent_session}
            taskColumn={task.column}
            onResume={() => setIsResumeModalOpen(true)}
          />

          {/* Comments Panel */}
          <CommentsPanel taskId={task.id} />
        </div>
      </div>

      {/* Resume Modal */}
      <ResumeModal
        isOpen={isResumeModalOpen}
        onClose={() => setIsResumeModalOpen(false)}
        taskId={task.id}
        displayId={
          currentBoard && task.seq
            ? formatDisplayId(currentBoard.prefix, task.seq)
            : task.id.slice(0, 8)
        }
      />
    </div>
  )
}
