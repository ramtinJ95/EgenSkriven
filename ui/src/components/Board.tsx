import { useMemo, useState } from 'react'
import {
  DndContext,
  DragOverlay,
  closestCenter,
  PointerSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core'
import type { DragEndEvent, DragStartEvent } from '@dnd-kit/core'
import { Column } from './Column'
import { VirtualizedColumn } from './VirtualizedColumn'
import { TaskCard } from './TaskCard'
import { BoardSkeleton } from './Skeleton'
import { useCurrentBoard } from '../contexts'
import { type Task, type Column as ColumnType } from '../types/task'
import { DEFAULT_COLUMNS } from '../types/board'
import styles from './Board.module.css'

// Threshold for using virtualized columns
const VIRTUALIZATION_THRESHOLD = 50

interface BoardProps {
  tasks: Task[] // Filtered tasks passed from parent
  onTaskClick?: (task: Task) => void
  onTaskSelect?: (task: Task) => void
  selectedTaskId?: string | null
  multiSelectedIds?: Set<string>
  isSelected?: (taskId: string) => boolean
  moveTask: (id: string, column: ColumnType, position: number) => Promise<Task>
}

/**
 * Kanban board with columns and drag-and-drop.
 *
 * Features:
 * - Displays tasks grouped by column
 * - Filters tasks by current board and active filters
 * - Supports board-specific custom columns
 * - Drag tasks between columns
 * - Real-time updates from CLI changes
 * - Click task to open detail panel
 * - Selected task state for keyboard navigation
 */
export function Board({ tasks, onTaskClick, onTaskSelect, selectedTaskId, multiSelectedIds, isSelected, moveTask }: BoardProps) {
  const { currentBoard, loading: boardLoading } = useCurrentBoard()
  const [activeTask, setActiveTask] = useState<Task | null>(null)

  // Get columns from board or use defaults
  const columns = useMemo(() => {
    if (currentBoard?.columns && currentBoard.columns.length > 0) {
      return currentBoard.columns as ColumnType[]
    }
    return DEFAULT_COLUMNS as ColumnType[]
  }, [currentBoard])

  const loading = boardLoading

  // Configure drag sensors
  // PointerSensor requires a small movement before dragging starts
  // This prevents accidental drags when clicking
  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8, // 8px movement required to start drag
      },
    })
  )

  // Group tasks by column
  const tasksByColumn = useMemo(() => {
    const grouped: Record<string, Task[]> = {}

    // Initialize all columns with empty arrays
    columns.forEach((col) => {
      grouped[col] = []
    })

    tasks.forEach((task) => {
      if (grouped[task.column]) {
        grouped[task.column].push(task)
      }
    })

    // Sort each column by position
    Object.keys(grouped).forEach((col) => {
      grouped[col].sort((a, b) => a.position - b.position)
    })

    return grouped
  }, [tasks, columns])

  // Handle drag start - store the dragged task for overlay
  const handleDragStart = (event: DragStartEvent) => {
    const task = tasks.find((t) => t.id === event.active.id)
    if (task) {
      setActiveTask(task)
    }
  }

  // Handle drag end - move task to new column
  const handleDragEnd = async (event: DragEndEvent) => {
    setActiveTask(null)
    const { active, over } = event
    if (!over) return

    const taskId = active.id as string
    const task = tasks.find((t) => t.id === taskId)
    if (!task) return

    // Get the target column from the droppable area
    const targetColumn = over.data.current?.column as ColumnType | undefined
    if (!targetColumn) return

    // If dropped in same column, no change needed (position sorting is Phase 4)
    if (task.column === targetColumn) return

    // Calculate new position (append to end of target column)
    const targetTasks = tasksByColumn[targetColumn]
    const maxPosition = targetTasks.reduce((max, t) => Math.max(max, t.position), 0)
    const newPosition = maxPosition + 1000

    // Move task to new column
    try {
      await moveTask(taskId, targetColumn, newPosition)
    } catch (err) {
      console.error('Failed to move task:', err)
    }
  }

  if (loading) {
    return <BoardSkeleton />
  }

  if (!currentBoard) {
    return (
      <div className={styles.empty}>
        <span>No board selected</span>
        <p>Create a board using the sidebar to get started.</p>
      </div>
    )
  }

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
    >
      <div className={styles.board}>
        {columns.map((column) => {
          const columnTasks = tasksByColumn[column] || []
          
          // Use VirtualizedColumn for columns with many tasks
          if (columnTasks.length > VIRTUALIZATION_THRESHOLD) {
            return (
              <VirtualizedColumn
                key={column}
                column={column}
                tasks={columnTasks}
                onTaskClick={onTaskClick}
                onTaskSelect={onTaskSelect}
                selectedTaskId={selectedTaskId}
                isSelected={isSelected}
                currentBoard={currentBoard}
              />
            )
          }
          
          return (
            <Column
              key={column}
              column={column}
              tasks={columnTasks}
              onTaskClick={onTaskClick}
              onTaskSelect={onTaskSelect}
              selectedTaskId={selectedTaskId}
              isSelected={isSelected}
              currentBoard={currentBoard}
            />
          )
        })}
      </div>

      {/* Drag overlay - renders dragged card in a portal above everything */}
      <DragOverlay dropAnimation={null}>
        {activeTask ? (
          <div style={{ width: 'var(--column-width)' }}>
            <TaskCard task={activeTask} currentBoard={currentBoard} isDragOverlay />
          </div>
        ) : null}
      </DragOverlay>
    </DndContext>
  )
}
