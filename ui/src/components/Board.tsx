import { useMemo } from 'react'
import {
  DndContext,
  DragOverlay,
  closestCenter,
  PointerSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core'
import type { DragEndEvent, DragStartEvent } from '@dnd-kit/core'
import { useState } from 'react'
import { Column } from './Column'
import { TaskCard } from './TaskCard'
import { useTasks } from '../hooks/useTasks'
import { COLUMNS, type Task, type Column as ColumnType } from '../types/task'
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
 * - Drag tasks between columns
 * - Real-time updates from CLI changes
 * - Click task to open detail panel
 * - Selected task state for keyboard navigation
 */
export function Board({ onTaskClick, onTaskSelect, selectedTaskId }: BoardProps) {
  const { tasks, loading, error, moveTask } = useTasks()
  const [activeTask, setActiveTask] = useState<Task | null>(null)

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
    const grouped: Record<ColumnType, Task[]> = {
      backlog: [],
      todo: [],
      in_progress: [],
      review: [],
      done: [],
    }

    tasks.forEach((task) => {
      if (grouped[task.column]) {
        grouped[task.column].push(task)
      }
    })

    // Sort each column by position
    Object.keys(grouped).forEach((col) => {
      grouped[col as ColumnType].sort((a, b) => a.position - b.position)
    })

    return grouped
  }, [tasks])

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
    const maxPosition = targetTasks.reduce(
      (max, t) => Math.max(max, t.position),
      0
    )
    const newPosition = maxPosition + 1000

    // Move task to new column
    await moveTask(taskId, targetColumn, newPosition)
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

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
    >
      <div className={styles.board}>
        {COLUMNS.map((column) => (
          <Column
            key={column}
            column={column}
            tasks={tasksByColumn[column]}
            onTaskClick={onTaskClick}
            onTaskSelect={onTaskSelect}
            selectedTaskId={selectedTaskId}
          />
        ))}
      </div>

      {/* Drag overlay shows the card being dragged */}
      <DragOverlay>
        {activeTask ? <TaskCard task={activeTask} isDragging /> : null}
      </DragOverlay>
    </DndContext>
  )
}
