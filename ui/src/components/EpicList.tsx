import { useMemo, useState, useCallback, useEffect } from 'react'
import { useEpics } from '../hooks/useEpics'
import type { Task } from '../types/task'
import { EPIC_COLORS } from '../types/epic'
import styles from './EpicList.module.css'

interface EpicListProps {
  /** Current board ID to filter epics by */
  boardId?: string
  /** All tasks to calculate counts from */
  tasks: Task[]
  /** Currently selected epic ID */
  selectedEpicId?: string | null
  /** Callback when an epic is selected for filtering */
  onSelectEpic: (epicId: string | null) => void
  /** Callback when epic detail view should be opened */
  onEpicDetailClick?: (epicId: string) => void
}

/**
 * EpicList component for sidebar.
 *
 * Features:
 * - Lists epics for the current board only (board-scoped)
 * - Shows task count per epic
 * - Supports selecting an epic to filter tasks
 * - "All Tasks" option to clear filter
 * - Loading and error states
 */
export function EpicList({ boardId, tasks, selectedEpicId, onSelectEpic, onEpicDetailClick }: EpicListProps) {
  const { epics, loading, error, createEpic } = useEpics(boardId)
  const [showNewEpic, setShowNewEpic] = useState(false)

  // Calculate task counts per epic
  const epicCounts = useMemo(() => {
    const counts: Record<string, number> = {}
    tasks.forEach((task) => {
      if (task.epic) {
        counts[task.epic] = (counts[task.epic] || 0) + 1
      }
    })
    return counts
  }, [tasks])

  // Count tasks without an epic
  const noEpicCount = useMemo(() => {
    return tasks.filter((task) => !task.epic).length
  }, [tasks])

  // Total task count
  const totalCount = tasks.length

  if (loading) {
    return (
      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>EPICS</h2>
        <div className={styles.loading}>Loading...</div>
      </section>
    )
  }

  if (error) {
    return (
      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>EPICS</h2>
        <div className={styles.error}>Failed to load epics</div>
      </section>
    )
  }

  return (
    <section className={styles.section}>
      <h2 className={styles.sectionTitle}>EPICS</h2>

      <ul className={styles.epicList}>
        {/* All Epics option */}
        <li>
          <button
            className={`${styles.epicItem} ${selectedEpicId === null ? styles.active : ''}`}
            onClick={() => onSelectEpic(null)}
          >
            <span className={styles.epicIcon}>
              <AllIcon />
            </span>
            <span className={styles.epicName}>All Tasks</span>
            <span className={styles.epicCount}>{totalCount}</span>
          </button>
        </li>

        {/* Epic list */}
        {epics.map((epic) => (
          <li key={epic.id} className={styles.epicRow}>
            <button
              className={`${styles.epicItem} ${selectedEpicId === epic.id ? styles.active : ''}`}
              onClick={() => onSelectEpic(epic.id)}
            >
              <span
                className={styles.epicIndicator}
                style={{ backgroundColor: epic.color || EPIC_COLORS[0] }}
              />
              <span className={styles.epicName}>{epic.title}</span>
              <span className={styles.epicCount}>{epicCounts[epic.id] || 0}</span>
            </button>
            {onEpicDetailClick && (
              <button
                className={styles.epicDetailBtn}
                onClick={(e) => {
                  e.stopPropagation()
                  onEpicDetailClick(epic.id)
                }}
                aria-label={`View ${epic.title} details`}
                title="View epic details"
              >
                <DetailIcon />
              </button>
            )}
          </li>
        ))}

        {/* No Epic option (only show if there are tasks without epics) */}
        {noEpicCount > 0 && (
          <li>
            <button
              className={`${styles.epicItem} ${selectedEpicId === 'none' ? styles.active : ''}`}
              onClick={() => onSelectEpic('none')}
            >
              <span className={styles.epicIcon}>
                <NoEpicIcon />
              </span>
              <span className={styles.epicName}>No Epic</span>
              <span className={styles.epicCount}>{noEpicCount}</span>
            </button>
          </li>
        )}
      </ul>

      {epics.length === 0 && (
        <div className={styles.empty}>No epics yet</div>
      )}

      {/* New Epic Button */}
      {boardId && (
        <button
          className={styles.newEpicButton}
          onClick={() => setShowNewEpic(true)}
        >
          + New epic
        </button>
      )}

      {/* New Epic Modal */}
      {showNewEpic && boardId && (
        <NewEpicModal
          onClose={() => setShowNewEpic(false)}
          createEpic={createEpic}
          boardId={boardId}
        />
      )}
    </section>
  )
}

// Icon for "All Tasks" option
function AllIcon() {
  return (
    <svg
      width="14"
      height="14"
      viewBox="0 0 16 16"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <rect x="2" y="2" width="12" height="12" rx="2" />
      <path d="M2 6h12" />
      <path d="M6 2v12" />
    </svg>
  )
}

// Icon for "No Epic" option
function NoEpicIcon() {
  return (
    <svg
      width="14"
      height="14"
      viewBox="0 0 16 16"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <circle cx="8" cy="8" r="6" />
      <path d="M4 4l8 8" />
    </svg>
  )
}

// Icon for epic detail button
function DetailIcon() {
  return (
    <svg
      width="14"
      height="14"
      viewBox="0 0 16 16"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <circle cx="8" cy="8" r="6" />
      <path d="M8 6v4" />
      <circle cx="8" cy="11" r="0.5" fill="currentColor" />
    </svg>
  )
}

// Modal for creating a new epic
interface NewEpicModalProps {
  onClose: () => void
  createEpic: (input: { title: string; description?: string; color?: string; board: string }) => Promise<unknown>
  boardId: string
}

function NewEpicModal({ onClose, createEpic, boardId }: NewEpicModalProps) {
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [color, setColor] = useState(EPIC_COLORS[0])
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  // Handle Escape key to close modal
  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if (e.key === 'Escape' && !submitting) {
      onClose()
    }
  }, [onClose, submitting])

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (!title.trim()) {
      setError('Title is required')
      return
    }

    setSubmitting(true)
    try {
      await createEpic({
        title: title.trim(),
        description: description.trim() || undefined,
        color,
        board: boardId,
      })
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create epic')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className={styles.modalOverlay} onClick={onClose}>
      <div
        className={styles.modal}
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-labelledby="new-epic-modal-title"
      >
        <h2 id="new-epic-modal-title" className={styles.modalTitle}>Create New Epic</h2>
        <form onSubmit={handleSubmit}>
          <div className={styles.formField}>
            <label htmlFor="epic-title">Title</label>
            <input
              id="epic-title"
              type="text"
              value={title}
              onChange={(e) => {
                setTitle(e.target.value)
                if (error) setError('')
              }}
              placeholder="e.g., User Authentication"
              autoFocus
              disabled={submitting}
              maxLength={200}
            />
          </div>

          <div className={styles.formField}>
            <label htmlFor="epic-description">Description (optional)</label>
            <textarea
              id="epic-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Describe what this epic covers..."
              disabled={submitting}
              maxLength={5000}
              rows={3}
            />
          </div>

          <div className={styles.formField}>
            <label>Color</label>
            <div className={styles.colorPicker}>
              {EPIC_COLORS.map((c) => (
                <button
                  key={c}
                  type="button"
                  className={`${styles.colorOption} ${color === c ? styles.colorSelected : ''}`}
                  style={{ backgroundColor: c }}
                  onClick={() => setColor(c)}
                  disabled={submitting}
                  aria-label={`Select color ${c}`}
                />
              ))}
            </div>
          </div>

          {error && <div className={styles.formError}>{error}</div>}

          <div className={styles.modalActions}>
            <button
              type="button"
              className={styles.buttonSecondary}
              onClick={onClose}
              disabled={submitting}
            >
              Cancel
            </button>
            <button
              type="submit"
              className={styles.buttonPrimary}
              disabled={submitting}
            >
              {submitting ? 'Creating...' : 'Create Epic'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
