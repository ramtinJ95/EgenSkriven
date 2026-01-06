import { useState, useMemo, useEffect, useCallback } from 'react'
import { useEpic } from '../hooks/useEpic'
import type { Task, Column } from '../types/task'
import { COLUMN_NAMES } from '../types/task'
import { EPIC_COLORS } from '../types/epic'
import styles from './EpicDetail.module.css'

interface EpicDetailProps {
  /** Epic ID to display, or null if no epic selected */
  epicId: string | null
  /** All tasks to filter for this epic */
  tasks: Task[]
  /** Callback when close button is clicked */
  onClose: () => void
  /** Callback when a task is clicked */
  onTaskClick: (task: Task) => void
}

/**
 * EpicDetail component displays a full view of an epic.
 *
 * Features:
 * - Shows epic title, description, and color
 * - Displays progress bar with task completion
 * - Lists all tasks grouped by column/status
 * - Allows editing epic title and description
 * - Supports deleting the epic
 */
export function EpicDetail({ epicId, tasks, onClose, onTaskClick }: EpicDetailProps) {
  const { epic, loading, error, updateEpic, deleteEpic } = useEpic(epicId)
  const [isEditing, setIsEditing] = useState(false)
  const [editTitle, setEditTitle] = useState('')
  const [editDescription, setEditDescription] = useState('')
  const [editColor, setEditColor] = useState(EPIC_COLORS[0])
  const [saving, setSaving] = useState(false)

  // Filter tasks belonging to this epic
  const epicTasks = useMemo(() => {
    return tasks.filter((t) => t.epic === epicId)
  }, [tasks, epicId])

  // Calculate progress
  const { completedCount, totalCount, progressPercent } = useMemo(() => {
    const completed = epicTasks.filter((t) => t.column === 'done').length
    const total = epicTasks.length
    const percent = total > 0 ? (completed / total) * 100 : 0
    return { completedCount: completed, totalCount: total, progressPercent: percent }
  }, [epicTasks])

  // Group tasks by column
  const tasksByColumn = useMemo(() => {
    const grouped: Record<Column, Task[]> = {
      backlog: [],
      todo: [],
      in_progress: [],
      review: [],
      done: [],
    }
    for (const task of epicTasks) {
      grouped[task.column]?.push(task)
    }
    return grouped
  }, [epicTasks])

  // Reset edit state when epic changes
  useEffect(() => {
    if (epic) {
      setEditTitle(epic.title)
      setEditDescription(epic.description || '')
      setEditColor(epic.color || EPIC_COLORS[0])
    }
    setIsEditing(false)
  }, [epic])

  const handleStartEdit = useCallback(() => {
    if (epic) {
      setEditTitle(epic.title)
      setEditDescription(epic.description || '')
      setEditColor(epic.color || EPIC_COLORS[0])
      setIsEditing(true)
    }
  }, [epic])

  const handleCancelEdit = useCallback(() => {
    setIsEditing(false)
  }, [])

  const handleSaveEdit = useCallback(async () => {
    if (!editTitle.trim()) return

    setSaving(true)
    try {
      await updateEpic({
        title: editTitle.trim(),
        description: editDescription.trim(),
        color: editColor,
      })
      setIsEditing(false)
    } catch (err) {
      console.error('Failed to update epic:', err)
    } finally {
      setSaving(false)
    }
  }, [editTitle, editDescription, editColor, updateEpic])

  const handleDelete = useCallback(async () => {
    if (!epic) return

    const confirmed = window.confirm(
      `Delete epic "${epic.title}"? Tasks will remain but be unlinked from this epic.`
    )

    if (confirmed) {
      try {
        await deleteEpic()
        onClose()
      } catch (err) {
        console.error('Failed to delete epic:', err)
      }
    }
  }, [epic, deleteEpic, onClose])

  // Handle keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        if (isEditing) {
          handleCancelEdit()
        } else {
          onClose()
        }
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [isEditing, handleCancelEdit, onClose])

  // Don't render if no epicId
  if (!epicId) {
    return null
  }

  if (loading) {
    return (
      <div className={styles.overlay}>
        <div className={styles.panel}>
          <div className={styles.loading}>Loading epic...</div>
        </div>
      </div>
    )
  }

  if (error || !epic) {
    return (
      <div className={styles.overlay}>
        <div className={styles.panel}>
          <div className={styles.error}>
            <p>Epic not found</p>
            <button className={styles.closeButton} onClick={onClose}>
              Close
            </button>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.panel} onClick={(e) => e.stopPropagation()}>
        {/* Header */}
        <div className={styles.header}>
          <button className={styles.backButton} onClick={onClose}>
            <BackIcon />
            <span>Back</span>
          </button>
          <div className={styles.actions}>
            {!isEditing && (
              <>
                <button className={styles.actionButton} onClick={handleStartEdit}>
                  Edit
                </button>
                <button className={`${styles.actionButton} ${styles.danger}`} onClick={handleDelete}>
                  Delete
                </button>
              </>
            )}
          </div>
        </div>

        {/* Epic Info */}
        <div className={styles.info}>
          {/* Color bar */}
          <div
            className={styles.colorBar}
            style={{ backgroundColor: epic.color || EPIC_COLORS[0] }}
          />

          {isEditing ? (
            <div className={styles.editForm}>
              <input
                type="text"
                className={styles.titleInput}
                value={editTitle}
                onChange={(e) => setEditTitle(e.target.value)}
                placeholder="Epic title"
                autoFocus
                disabled={saving}
              />
              <textarea
                className={styles.descriptionInput}
                value={editDescription}
                onChange={(e) => setEditDescription(e.target.value)}
                placeholder="Description (optional)"
                rows={3}
                disabled={saving}
              />
              <div className={styles.colorPicker}>
                <span className={styles.colorLabel}>Color</span>
                <div className={styles.colorOptions}>
                  {EPIC_COLORS.map((color) => (
                    <button
                      key={color}
                      type="button"
                      className={`${styles.colorOption} ${editColor === color ? styles.selected : ''}`}
                      style={{ backgroundColor: color }}
                      onClick={() => setEditColor(color)}
                      disabled={saving}
                      aria-label={`Select color ${color}`}
                    />
                  ))}
                </div>
              </div>
              <div className={styles.editActions}>
                <button
                  className={styles.saveButton}
                  onClick={handleSaveEdit}
                  disabled={saving || !editTitle.trim()}
                >
                  {saving ? 'Saving...' : 'Save'}
                </button>
                <button
                  className={styles.cancelButton}
                  onClick={handleCancelEdit}
                  disabled={saving}
                >
                  Cancel
                </button>
              </div>
            </div>
          ) : (
            <>
              <h1 className={styles.title}>{epic.title}</h1>
              {epic.description && (
                <p className={styles.description}>{epic.description}</p>
              )}
            </>
          )}

          {/* Progress bar */}
          <div className={styles.progress}>
            <div className={styles.progressHeader}>
              <span>Progress</span>
              <span>
                {completedCount} / {totalCount} tasks
              </span>
            </div>
            <div className={styles.progressBar}>
              <div
                className={styles.progressFill}
                style={{ width: `${progressPercent}%` }}
              />
            </div>
          </div>
        </div>

        {/* Task list by status */}
        <div className={styles.tasks}>
          {(['backlog', 'todo', 'in_progress', 'review', 'done'] as Column[]).map((column) => {
            const columnTasks = tasksByColumn[column]
            if (columnTasks.length === 0) return null

            return (
              <div key={column} className={styles.taskGroup}>
                <h3 className={styles.groupHeader}>
                  {COLUMN_NAMES[column]} ({columnTasks.length})
                </h3>
                <ul className={styles.groupTasks}>
                  {columnTasks.map((task) => (
                    <li key={task.id}>
                      <button
                        className={styles.taskItem}
                        onClick={() => onTaskClick(task)}
                      >
                        <span className={styles.taskStatus} data-column={task.column} />
                        <span className={styles.taskTitle}>{task.title}</span>
                        {task.priority && (
                          <span className={styles.taskPriority} data-priority={task.priority}>
                            {task.priority === 'urgent' && '!!!'}
                            {task.priority === 'high' && '!!'}
                            {task.priority === 'medium' && '!'}
                          </span>
                        )}
                      </button>
                    </li>
                  ))}
                </ul>
              </div>
            )
          })}

          {totalCount === 0 && (
            <div className={styles.empty}>
              <p>No tasks in this epic yet.</p>
              <p className={styles.hint}>
                Link tasks to this epic from the task detail panel.
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

// Back arrow icon
function BackIcon() {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M10 12L6 8L10 4" />
    </svg>
  )
}
