import { type Task, COLUMN_NAMES } from '../types/task'
import styles from './SubtaskList.module.css'

interface SubtaskListProps {
  parentId: string
  tasks: Task[]
  onTaskClick?: (task: Task) => void
  onToggleComplete?: (task: Task) => void
}

/**
 * SubtaskList displays sub-tasks of a parent task.
 * 
 * Features:
 * - Shows list of sub-tasks with completion status
 * - Click to toggle completion (moves between todo/done)
 * - Click on title to view task details
 * - Progress indicator showing completed vs total
 */
export function SubtaskList({ parentId, tasks, onTaskClick, onToggleComplete }: SubtaskListProps) {
  // Filter tasks to only show direct children of this parent
  const subtasks = tasks.filter(task => task.parent === parentId)
  
  // Calculate completion stats
  const completedCount = subtasks.filter(t => t.column === 'done').length
  const totalCount = subtasks.length
  const progressPercent = totalCount > 0 ? Math.round((completedCount / totalCount) * 100) : 0

  if (subtasks.length === 0) {
    return null
  }

  const handleCheckboxClick = (e: React.MouseEvent, task: Task) => {
    e.stopPropagation()
    if (onToggleComplete) {
      onToggleComplete(task)
    }
  }

  return (
    <div className={styles.subtaskList}>
      <div className={styles.header}>
        <h3 className={styles.title}>Sub-tasks</h3>
        <span className={styles.count}>
          {completedCount}/{totalCount}
        </span>
      </div>

      {/* Progress bar */}
      <div className={styles.progressBar}>
        <div 
          className={styles.progressFill} 
          style={{ width: `${progressPercent}%` }}
        />
      </div>

      {/* Subtask items */}
      <ul className={styles.list}>
        {subtasks.map(task => {
          const isCompleted = task.column === 'done'
          
          return (
            <li 
              key={task.id} 
              className={`${styles.item} ${isCompleted ? styles.completed : ''}`}
            >
              <button
                className={styles.checkbox}
                onClick={(e) => handleCheckboxClick(e, task)}
                aria-label={isCompleted ? 'Mark as incomplete' : 'Mark as complete'}
                type="button"
              >
                {isCompleted ? (
                  <span className={styles.checkmark}>&#10003;</span>
                ) : (
                  <span className={styles.unchecked} />
                )}
              </button>
              
              <button
                className={styles.taskButton}
                onClick={() => onTaskClick?.(task)}
                type="button"
              >
                <span className={styles.taskTitle}>{task.title}</span>
                <span className={styles.taskStatus}>
                  {COLUMN_NAMES[task.column]}
                </span>
              </button>
            </li>
          )
        })}
      </ul>
    </div>
  )
}
