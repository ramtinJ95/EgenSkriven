import { useState } from 'react'
import { useCurrentBoard } from '../contexts'
import { BOARD_COLORS } from '../types/board'
import type { Task } from '../types/task'
import { ViewsSidebar } from './ViewsSidebar'
import { EpicList } from './EpicList'
import styles from './Sidebar.module.css'

interface SidebarProps {
  collapsed: boolean
  onToggle: () => void
  /** Tasks for epic count calculation */
  tasks?: Task[]
  /** Currently selected epic filter */
  selectedEpicId?: string | null
  /** Callback when epic filter changes */
  onSelectEpic?: (epicId: string | null) => void
  /** Callback when epic detail view should be opened */
  onEpicDetailClick?: (epicId: string) => void
}

/**
 * Sidebar component for board navigation.
 *
 * Features:
 * - Lists all boards with current board indicator
 * - Collapsible for more screen space
 * - New board creation modal
 * - Board color indicators
 */
export function Sidebar({ collapsed, onToggle, tasks = [], selectedEpicId, onSelectEpic, onEpicDetailClick }: SidebarProps) {
  const { boards, loading, boardsError: error, createBoard, currentBoard, setCurrentBoard } = useCurrentBoard()
  const [showNewBoard, setShowNewBoard] = useState(false)

  if (collapsed) {
    return (
      <aside className={styles.sidebarCollapsed}>
        <button
          className={styles.toggleButton}
          onClick={onToggle}
          aria-label="Expand sidebar"
        >
          <ChevronRightIcon />
        </button>
      </aside>
    )
  }

  return (
    <aside className={styles.sidebar}>
      <div className={styles.header}>
        <h1 className={styles.title}>EgenSkriven</h1>
        <button
          className={styles.toggleButton}
          onClick={onToggle}
          aria-label="Collapse sidebar"
        >
          <ChevronLeftIcon />
        </button>
      </div>

      <nav className={styles.nav}>
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>BOARDS</h2>

          {loading ? (
            <div className={styles.loading}>Loading...</div>
          ) : error ? (
            <div className={styles.error}>Failed to load boards</div>
          ) : boards.length === 0 ? (
            <div className={styles.empty}>No boards yet</div>
          ) : (
            <ul className={styles.boardList}>
              {boards.map((board) => (
                <li key={board.id}>
                  <button
                    className={`${styles.boardItem} ${
                      currentBoard?.id === board.id ? styles.active : ''
                    }`}
                    onClick={() => setCurrentBoard(board)}
                  >
                    <span
                      className={styles.boardIndicator}
                      style={{ backgroundColor: board.color || BOARD_COLORS[0] }}
                    />
                    <span className={styles.boardName}>{board.name}</span>
                    <span className={styles.boardPrefix}>({board.prefix})</span>
                  </button>
                </li>
              ))}
            </ul>
          )}

          <button
            className={styles.newBoardButton}
            onClick={() => setShowNewBoard(true)}
          >
            + New board
          </button>
        </section>

        {/* Views section - only show when a board is selected */}
        <ViewsSidebar boardId={currentBoard?.id || null} />

        {/* Epics section - only show when a board is selected and handler provided */}
        {currentBoard && onSelectEpic && (
          <EpicList
            boardId={currentBoard.id}
            tasks={tasks}
            selectedEpicId={selectedEpicId}
            onSelectEpic={onSelectEpic}
            onEpicDetailClick={onEpicDetailClick}
          />
        )}
      </nav>

      {showNewBoard && (
        <NewBoardModal onClose={() => setShowNewBoard(false)} createBoard={createBoard} />
      )}
    </aside>
  )
}

interface NewBoardModalProps {
  onClose: () => void
  createBoard: (input: { name: string; prefix: string; color: string }) => Promise<unknown>
}

/**
 * Modal for creating a new board.
 */
function NewBoardModal({ onClose, createBoard }: NewBoardModalProps) {
  const [name, setName] = useState('')
  const [prefix, setPrefix] = useState('')
  const [color, setColor] = useState(BOARD_COLORS[0])
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (!name.trim()) {
      setError('Name is required')
      return
    }
    if (!prefix.trim()) {
      setError('Prefix is required')
      return
    }
    if (prefix.length > 10) {
      setError('Prefix must be 10 characters or less')
      return
    }
    if (!/^[A-Za-z0-9]+$/.test(prefix)) {
      setError('Prefix must be alphanumeric')
      return
    }

    setSubmitting(true)
    try {
      await createBoard({ name: name.trim(), prefix: prefix.toUpperCase(), color })
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create board')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className={styles.modalOverlay} onClick={onClose}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
        <h2 className={styles.modalTitle}>Create New Board</h2>
        <form onSubmit={handleSubmit}>
          <div className={styles.formField}>
            <label htmlFor="board-name">Name</label>
            <input
              id="board-name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g., Work, Personal"
              autoFocus
              disabled={submitting}
            />
          </div>

          <div className={styles.formField}>
            <label htmlFor="board-prefix">Prefix</label>
            <input
              id="board-prefix"
              type="text"
              value={prefix}
              onChange={(e) => setPrefix(e.target.value.toUpperCase())}
              placeholder="e.g., WRK, PER"
              maxLength={10}
              disabled={submitting}
            />
            <span className={styles.formHint}>
              Used in task IDs (e.g., {prefix || 'WRK'}-123)
            </span>
          </div>

          <div className={styles.formField}>
            <label>Color</label>
            <div className={styles.colorPicker}>
              {BOARD_COLORS.map((c) => (
                <button
                  key={c}
                  type="button"
                  className={`${styles.colorOption} ${
                    color === c ? styles.colorSelected : ''
                  }`}
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
              {submitting ? 'Creating...' : 'Create Board'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

// Simple SVG icons
function ChevronLeftIcon() {
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

function ChevronRightIcon() {
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
      <path d="M6 4L10 8L6 12" />
    </svg>
  )
}
