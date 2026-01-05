import { useDroppable } from '@dnd-kit/core'
import { TaskCard } from './TaskCard'
import { COLUMN_NAMES, type Task } from '../types/task'
import type { Board } from '../types/board'
import styles from './Column.module.css'

interface ColumnProps {
  column: string
  tasks: Task[]
  onTaskClick?: (task: Task) => void
  onTaskSelect?: (task: Task) => void
  selectedTaskId?: string | null
  isSelected?: (taskId: string) => boolean
  currentBoard?: Board | null
}

/**
 * Format column name for display.
 * Uses predefined names if available, otherwise formats the key.
 */
function getColumnDisplayName(column: string): string {
  // Check if we have a predefined name
  if (column in COLUMN_NAMES) {
    return COLUMN_NAMES[column as keyof typeof COLUMN_NAMES]
  }
  // Format custom column name: replace underscores with spaces, title case
  return column
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}

/**
 * A single column in the kanban board.
 * 
 * Acts as a droppable target for drag-and-drop.
 * Displays column header with name and count.
 */
export function Column({ column, tasks, onTaskClick, onTaskSelect, selectedTaskId, isSelected, currentBoard }: ColumnProps) {
  // Make this column a droppable target
  const { setNodeRef, isOver } = useDroppable({
    id: `column-${column}`,
    data: {
      column, // Pass column info to drag handlers
    },
  })

  return (
    <div
      ref={setNodeRef}
      className={`${styles.column} ${isOver ? styles.over : ''}`}
    >
      <div className={styles.header}>
        <div className={styles.headerContent}>
          <span
            className={styles.statusDot}
            style={{ backgroundColor: `var(--status-${column.replace('_', '-')})` }}
          />
          <span className={styles.name}>{getColumnDisplayName(column)}</span>
          <span className={styles.count}>{tasks.length}</span>
        </div>
      </div>

      <div className={styles.tasks}>
        {tasks.map((task) => (
          <TaskCard 
            key={task.id} 
            task={task} 
            onClick={onTaskClick}
            onSelect={onTaskSelect}
            isSelected={isSelected ? isSelected(task.id) : selectedTaskId === task.id}
            currentBoard={currentBoard}
          />
        ))}
      </div>
    </div>
  )
}
