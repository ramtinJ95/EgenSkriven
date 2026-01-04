import { useState, useEffect, useRef } from 'react'
import { COLUMNS, COLUMN_NAMES, type Column } from '../types/task'
import styles from './QuickCreate.module.css'

interface QuickCreateProps {
  isOpen: boolean
  onClose: () => void
  onCreate: (title: string, column: Column) => Promise<void>
}

/**
 * Modal for quickly creating a new task.
 * 
 * Features:
 * - Auto-focus on title input
 * - Column selector (defaults to "backlog")
 * - Enter to create, Esc to cancel
 */
export function QuickCreate({ isOpen, onClose, onCreate }: QuickCreateProps) {
  const [title, setTitle] = useState('')
  const [column, setColumn] = useState<Column>('backlog')
  const [isCreating, setIsCreating] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  // Auto-focus input when modal opens
  useEffect(() => {
    if (isOpen && inputRef.current) {
      inputRef.current.focus()
    }
  }, [isOpen])

  // Reset form when modal closes
  useEffect(() => {
    if (!isOpen) {
      setTitle('')
      setColumn('backlog')
    }
  }, [isOpen])

  // Handle keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (!isOpen) return

      if (e.key === 'Escape') {
        onClose()
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [isOpen, onClose])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!title.trim() || isCreating) return

    setIsCreating(true)
    try {
      await onCreate(title.trim(), column)
      onClose()
    } finally {
      setIsCreating(false)
    }
  }

  if (!isOpen) return null

  return (
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
        <h2 className={styles.title}>Create Task</h2>
        
        <form onSubmit={handleSubmit}>
          <div className={styles.field}>
            <input
              ref={inputRef}
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Task title..."
              className={styles.input}
              disabled={isCreating}
            />
          </div>

          <div className={styles.field}>
            <label className={styles.label}>Column</label>
            <select
              value={column}
              onChange={(e) => setColumn(e.target.value as Column)}
              className={styles.select}
              disabled={isCreating}
            >
              {COLUMNS.map((col) => (
                <option key={col} value={col}>
                  {COLUMN_NAMES[col]}
                </option>
              ))}
            </select>
          </div>

          <div className={styles.actions}>
            <button
              type="button"
              onClick={onClose}
              className={styles.cancelButton}
              disabled={isCreating}
            >
              Cancel
            </button>
            <button
              type="submit"
              className={styles.createButton}
              disabled={!title.trim() || isCreating}
            >
              {isCreating ? 'Creating...' : 'Create'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
