import { useState, useEffect, useRef } from 'react'
import { COLUMNS, COLUMN_NAMES, type Column } from '../types/task'
import styles from './QuickCreate.module.css'

interface QuickCreateProps {
  isOpen: boolean
  onClose: () => void
  onCreate: (title: string, column: Column, description?: string) => Promise<void>
}

/**
 * Modal for quickly creating a new task.
 * 
 * Features:
 * - Auto-focus on title input
 * - Column selector (defaults to "backlog")
 * - Optional description field (expandable)
 * - Enter to create, Esc to cancel
 * - Supports Markdown in description
 */
export function QuickCreate({ isOpen, onClose, onCreate }: QuickCreateProps) {
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [showDescription, setShowDescription] = useState(false)
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
      setDescription('')
      setShowDescription(false)
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
      await onCreate(title.trim(), column, description.trim() || undefined)
      onClose()
    } catch (err) {
      console.error('Failed to create task:', err)
      // Keep modal open so user can retry
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

          {/* Description field - expandable */}
          {!showDescription ? (
            <button
              type="button"
              className={styles.addDescriptionButton}
              onClick={() => setShowDescription(true)}
              disabled={isCreating}
            >
              + Add description
            </button>
          ) : (
            <div className={styles.field}>
              <label className={styles.label}>
                Description
                <span className={styles.labelHint}> (supports Markdown)</span>
              </label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Add a more detailed description..."
                className={styles.textarea}
                disabled={isCreating}
                rows={4}
              />
            </div>
          )}

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
