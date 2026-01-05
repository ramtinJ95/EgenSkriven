import { useRef } from 'react'
import { useDroppable } from '@dnd-kit/core'
import { useVirtualizer } from '@tanstack/react-virtual'
import { TaskCard } from './TaskCard'
import { COLUMN_NAMES, type Task } from '../types/task'
import type { Board } from '../types/board'
import styles from './Column.module.css'

// Estimated task card height - used as initial estimate before measurement
const ESTIMATED_TASK_HEIGHT = 88

interface VirtualizedColumnProps {
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
 * A virtualized column for the kanban board.
 * 
 * Use this when a column has many tasks (>50) to improve performance.
 * Only renders visible tasks, keeping DOM size small.
 * 
 * Acts as a droppable target for drag-and-drop.
 * Displays column header with name and count.
 */
export function VirtualizedColumn({ 
  column, 
  tasks, 
  onTaskClick, 
  onTaskSelect, 
  selectedTaskId, 
  isSelected,
  currentBoard 
}: VirtualizedColumnProps) {
  const parentRef = useRef<HTMLDivElement>(null)

  // Make this column a droppable target
  const { setNodeRef, isOver } = useDroppable({
    id: `column-${column}`,
    data: {
      column, // Pass column info to drag handlers
    },
  })

  const virtualizer = useVirtualizer({
    count: tasks.length,
    getScrollElement: () => parentRef.current,
    // Initial estimate - actual sizes measured dynamically
    estimateSize: () => ESTIMATED_TASK_HEIGHT,
    overscan: 5, // Render 5 extra items above/below viewport
    gap: 8, // Gap between items (matches --space-2)
  })

  const virtualItems = virtualizer.getVirtualItems()

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

      <div 
        ref={parentRef}
        className={styles.tasks}
        style={{ overflow: 'auto', height: '100%' }}
      >
        <div
          style={{
            height: `${virtualizer.getTotalSize()}px`,
            width: '100%',
            position: 'relative',
          }}
        >
          {virtualItems.map((virtualItem) => {
            const task = tasks[virtualItem.index]
            return (
              <div
                key={task.id}
                data-index={virtualItem.index}
                ref={virtualizer.measureElement}
                style={{
                  position: 'absolute',
                  top: 0,
                  left: 0,
                  width: '100%',
                  transform: `translateY(${virtualItem.start}px)`,
                }}
              >
                <TaskCard
                  task={task}
                  onClick={onTaskClick}
                  onSelect={onTaskSelect}
                  isSelected={isSelected ? isSelected(task.id) : selectedTaskId === task.id}
                  currentBoard={currentBoard}
                />
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}
