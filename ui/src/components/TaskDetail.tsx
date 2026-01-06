import { useEffect, useRef, useState, useCallback } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
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
import { DatePicker } from './DatePicker'
import { SubtaskList } from './SubtaskList'
import styles from './TaskDetail.module.css'

interface TaskDetailProps {
  task: Task | null
  tasks: Task[]
  onClose: () => void
  onUpdate: (id: string, data: Partial<Task>) => Promise<void>
  onTaskClick?: (task: Task) => void
}

/**
 * Slide-in panel showing full task details.
 * 
 * Features:
 * - Close with Esc or click outside
 * - Editable properties via dropdowns (status, priority, type)
 * - Editable description with Markdown rendering
 * - Display all task metadata
 */
export function TaskDetail({ task, tasks, onClose, onUpdate, onTaskClick }: TaskDetailProps) {
  const panelRef = useRef<HTMLDivElement>(null)
  const [isEditingDescription, setIsEditingDescription] = useState(false)
  const [descriptionDraft, setDescriptionDraft] = useState('')
  const descriptionTextareaRef = useRef<HTMLTextAreaElement>(null)

  // Handle keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && task) {
        // If editing description, save and exit edit mode first
        if (isEditingDescription) {
          handleDescriptionBlur()
          return
        }
        onClose()
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [task, onClose, isEditingDescription])

  // Reset edit state when task changes
  useEffect(() => {
    setIsEditingDescription(false)
    setDescriptionDraft(task?.description || '')
  }, [task?.id])

  // Focus textarea when entering edit mode
  useEffect(() => {
    if (isEditingDescription && descriptionTextareaRef.current) {
      descriptionTextareaRef.current.focus()
      // Move cursor to end
      const len = descriptionTextareaRef.current.value.length
      descriptionTextareaRef.current.setSelectionRange(len, len)
    }
  }, [isEditingDescription])

  // Handle click outside
  const handleOverlayClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose()
    }
  }

  // Description editing handlers - must be defined before early return to follow hooks rules
  const handleDescriptionBlur = useCallback(async () => {
    if (!task) return
    
    const trimmedDescription = descriptionDraft.trim()
    const currentDescription = task.description || ''
    
    // Only update if changed
    if (trimmedDescription !== currentDescription) {
      await onUpdate(task.id, { description: trimmedDescription || undefined })
    }
    
    setIsEditingDescription(false)
  }, [task, descriptionDraft, onUpdate])

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

  // Toggle subtask completion (moves between todo and done)
  const handleToggleSubtaskComplete = async (subtask: Task) => {
    const newColumn = subtask.column === 'done' ? 'todo' : 'done'
    await onUpdate(subtask.id, { column: newColumn })
  }

  // Description click handler (defined after early return is OK, it's not a hook)
  const handleDescriptionClick = () => {
    setDescriptionDraft(task.description || '')
    setIsEditingDescription(true)
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
            <h3 className={styles.sectionTitle}>
              Description
              {!isEditingDescription && task.description && (
                <button 
                  className={styles.editButton}
                  onClick={handleDescriptionClick}
                  title="Edit description"
                >
                  Edit
                </button>
              )}
            </h3>
            
            {isEditingDescription ? (
              <div className={styles.descriptionEdit}>
                <textarea
                  ref={descriptionTextareaRef}
                  value={descriptionDraft}
                  onChange={(e) => setDescriptionDraft(e.target.value)}
                  onBlur={handleDescriptionBlur}
                  className={styles.descriptionTextarea}
                  placeholder="Add a description... (supports Markdown)"
                  rows={6}
                />
                <div className={styles.descriptionHint}>
                  <span>Markdown supported</span>
                  <span className={styles.charCount}>
                    {descriptionDraft.length} / 10,000
                  </span>
                </div>
              </div>
            ) : task.description ? (
              <div 
                className={styles.descriptionContent}
                onClick={handleDescriptionClick}
                role="button"
                tabIndex={0}
                onKeyDown={(e) => e.key === 'Enter' && handleDescriptionClick()}
              >
                <ReactMarkdown remarkPlugins={[remarkGfm]}>
                  {task.description}
                </ReactMarkdown>
              </div>
            ) : (
              <button 
                className={styles.addDescriptionButton}
                onClick={handleDescriptionClick}
              >
                Click to add description...
              </button>
            )}
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
        </div>
      </div>
    </div>
  )
}
