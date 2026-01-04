import { useDroppable } from '@dnd-kit/core'
import { TaskCard } from './TaskCard'
import { COLUMN_NAMES, type Column as ColumnType, type Task } from '../types/task'
import styles from './Column.module.css'

interface ColumnProps {
  column: ColumnType
  tasks: Task[]
  onTaskClick?: (task: Task) => void
  onTaskSelect?: (task: Task) => void
  selectedTaskId?: string | null
}

/**
 * A single column in the kanban board.
 * 
 * Acts as a droppable target for drag-and-drop.
 * Displays column header with name and count.
 */
export function Column({ column, tasks, onTaskClick, onTaskSelect, selectedTaskId }: ColumnProps) {
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
          <span className={styles.name}>{COLUMN_NAMES[column]}</span>
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
            isSelected={selectedTaskId === task.id}
          />
        ))}
      </div>
    </div>
  )
}
