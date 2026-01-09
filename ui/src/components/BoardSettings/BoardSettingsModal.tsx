import { useState, useEffect, useCallback, useRef } from 'react'
import { pb } from '../../lib/pb'
import type { Board, ResumeMode } from '../../types/board'
import { BOARD_COLORS } from '../../types/board'
import { ResumeModeSelector } from './ResumeModeSelector'
import styles from './BoardSettingsModal.module.css'

interface BoardSettingsModalProps {
  /** Whether the modal is open */
  isOpen: boolean
  /** Callback when modal should close */
  onClose: () => void
  /** The board to edit */
  board: Board
  /** Callback when board is updated */
  onUpdate?: (updatedBoard: Board) => void
}

/**
 * Modal for editing board settings.
 *
 * Features:
 * - Edit board name and color
 * - Configure resume mode (manual/command/auto)
 * - Form validation and error handling
 */
export function BoardSettingsModal({
  isOpen,
  onClose,
  board,
  onUpdate,
}: BoardSettingsModalProps) {
  const [name, setName] = useState(board.name)
  const [color, setColor] = useState(board.color || BOARD_COLORS[0])
  const [resumeMode, setResumeMode] = useState<ResumeMode>(
    (board.resume_mode as ResumeMode) || 'command'
  )
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)
  const [hasChanges, setHasChanges] = useState(false)

  const modalRef = useRef<HTMLDivElement>(null)
  const closeButtonRef = useRef<HTMLButtonElement>(null)

  // Reset form when board changes or modal opens
  useEffect(() => {
    if (isOpen) {
      setName(board.name)
      setColor(board.color || BOARD_COLORS[0])
      setResumeMode((board.resume_mode as ResumeMode) || 'command')
      setError('')
      setHasChanges(false)
    }
  }, [isOpen, board])

  // Track changes
  useEffect(() => {
    const changed =
      name !== board.name ||
      color !== (board.color || BOARD_COLORS[0]) ||
      resumeMode !== ((board.resume_mode as ResumeMode) || 'command')
    setHasChanges(changed)
  }, [name, color, resumeMode, board])

  // Focus management
  useEffect(() => {
    if (isOpen) {
      setTimeout(() => closeButtonRef.current?.focus(), 0)
    }
  }, [isOpen])

  // Handle Escape key
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (e.key === 'Escape' && !saving) {
        onClose()
      }
    },
    [onClose, saving]
  )

  useEffect(() => {
    if (!isOpen) return
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [isOpen, handleKeyDown])

  const handleSave = async () => {
    setError('')

    if (!name.trim()) {
      setError('Name is required')
      return
    }

    setSaving(true)
    try {
      const updated = await pb.collection('boards').update<Board>(board.id, {
        name: name.trim(),
        color,
        resume_mode: resumeMode,
      })
      onUpdate?.(updated)
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update board')
    } finally {
      setSaving(false)
    }
  }

  if (!isOpen) return null

  return (
    <div className={styles.overlay} onClick={onClose}>
      <div
        ref={modalRef}
        className={styles.modal}
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-labelledby="board-settings-title"
      >
        <div className={styles.header}>
          <h2 id="board-settings-title" className={styles.title}>
            Board Settings
          </h2>
          <button
            ref={closeButtonRef}
            className={styles.closeButton}
            onClick={onClose}
            aria-label="Close"
            disabled={saving}
          >
            <CloseIcon />
          </button>
        </div>

        <div className={styles.content}>
          {/* Board Name */}
          <div className={styles.field}>
            <label htmlFor="board-settings-name">Name</label>
            <input
              id="board-settings-name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              disabled={saving}
            />
          </div>

          {/* Board Prefix (read-only) */}
          <div className={styles.field}>
            <label>Prefix</label>
            <div className={styles.readOnly}>
              {board.prefix}
              <span className={styles.hint}>Prefix cannot be changed</span>
            </div>
          </div>

          {/* Board Color */}
          <div className={styles.field}>
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
                  disabled={saving}
                  aria-label={`Select color ${c}`}
                />
              ))}
            </div>
          </div>

          {/* Divider */}
          <div className={styles.divider} />

          {/* Resume Mode */}
          <ResumeModeSelector
            value={resumeMode}
            onChange={setResumeMode}
            disabled={saving}
          />

          {/* Error */}
          {error && <div className={styles.error}>{error}</div>}
        </div>

        <div className={styles.footer}>
          <button
            type="button"
            className={styles.buttonSecondary}
            onClick={onClose}
            disabled={saving}
          >
            Cancel
          </button>
          <button
            type="button"
            className={styles.buttonPrimary}
            onClick={handleSave}
            disabled={saving || !hasChanges}
          >
            {saving ? 'Saving...' : 'Save Changes'}
          </button>
        </div>
      </div>
    </div>
  )
}

function CloseIcon() {
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
      <path d="M12 4L4 12M4 4l8 8" />
    </svg>
  )
}

export default BoardSettingsModal
