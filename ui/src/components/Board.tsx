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
import { TaskCard } from './TaskCard'
import { useTasks } from '../hooks/useTasks'
import { useCurrentBoard } from '../hooks/useCurrentBoard'
import { useFilteredTasks } from '../hooks/useFilteredTasks'
import { type Task, type Column as ColumnType } from '../types/task'
import { DEFAULT_COLUMNS } from '../types/board'
import styles from './Board.module.css'

interface BoardProps {
  onTaskClick?: (task: Task) => void
  onTaskSelect?: (task: Task) => void
  selectedTaskId?: string | null
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
export function Board({ onTaskClick, onTaskSelect, selectedTaskId }: BoardProps) {
  const { currentBoard, loading: boardLoading } = useCurrentBoard()
  const { tasks: allTasks, loading: tasksLoading, error, moveTask } = useTasks(currentBoard?.id)
  const [activeTask, setActiveTask] = useState<Task | null>(null)

  // Apply filters to tasks
  const tasks = useFilteredTasks(allTasks)

  // Get columns from board or use defaults
  const columns = useMemo(() => {
    if (currentBoard?.columns && currentBoard.columns.length > 0) {
      return currentBoard.columns as ColumnType[]
    }
    return DEFAULT_COLUMNS as ColumnType[]
  }, [currentBoard])

  const loading = boardLoading || tasksLoading

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
    const task = allTasks.find((t) => t.id === event.active.id)
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
    const task = allTasks.find((t) => t.id === taskId)
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
    return (
      <div className={styles.loading}>
        <span>Loading tasks...</span>
      </div>
    )
  }

  if (error) {
    return (
      <div className={styles.error}>
        <span>Error: {error.message}</span>
      </div>
    )
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
        {columns.map((column) => (
          <Column
            key={column}
            column={column}
            tasks={tasksByColumn[column] || []}
            onTaskClick={onTaskClick}
            onTaskSelect={onTaskSelect}
            selectedTaskId={selectedTaskId}
            currentBoard={currentBoard}
          />
        ))}
      </div>

      {/* Drag overlay - renders dragged card in a portal above everything */}
      <DragOverlay dropAnimation={null}>
        {activeTask ? (
          <div style={{ width: 'var(--column-width)' }}>
            <TaskCard task={activeTask} currentBoard={currentBoard} />
          </div>
        ) : null}
      </DragOverlay>
    </DndContext>
  )
}
