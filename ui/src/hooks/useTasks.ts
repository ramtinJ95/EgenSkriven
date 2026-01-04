import { useEffect, useState, useCallback } from 'react'
import { pb } from '../lib/pb'
import type { Task, Column } from '../types/task'

interface UseTasksReturn {
  tasks: Task[]
  loading: boolean
  error: Error | null
  createTask: (title: string, column?: Column) => Promise<Task>
  updateTask: (id: string, data: Partial<Task>) => Promise<Task>
  deleteTask: (id: string) => Promise<void>
  moveTask: (id: string, column: Column, position: number) => Promise<Task>
}

/**
 * Hook for managing tasks with real-time updates.
 * 
 * Features:
 * - Fetches tasks on mount (optionally filtered by board)
 * - Subscribes to real-time create/update/delete events
 * - Provides CRUD operations
 * - Automatically updates local state on changes
 * 
 * @param boardId - Optional board ID to filter tasks. If not provided, fetches all tasks.
 * 
 * @example
 * ```tsx
 * function Board() {
 *   const { currentBoard } = useCurrentBoard()
 *   const { tasks, loading, moveTask } = useTasks(currentBoard?.id)
 *   
 *   if (loading) return <div>Loading...</div>
 *   
 *   return <BoardView tasks={tasks} onMove={moveTask} />
 * }
 * ```
 */
export function useTasks(boardId?: string): UseTasksReturn {
  const [tasks, setTasks] = useState<Task[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  // Fetch tasks on mount or when boardId changes
  useEffect(() => {
    const fetchTasks = async () => {
      setLoading(true)
      try {
        const options: { sort: string; filter?: string } = {
          sort: 'position',
        }
        
        // Filter by board if boardId is provided
        if (boardId) {
          options.filter = `board = "${boardId}"`
        }
        
        const records = await pb.collection('tasks').getFullList<Task>(options)
        setTasks(records)
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Failed to fetch tasks'))
      } finally {
        setLoading(false)
      }
    }

    fetchTasks()
  }, [boardId])

  // Subscribe to real-time updates
  useEffect(() => {
    // Subscribe to all task changes
    pb.collection('tasks').subscribe<Task>('*', (event) => {
      // Filter events by board if boardId is provided
      const taskBoardId = event.record.board
      const matchesBoard = !boardId || taskBoardId === boardId

      switch (event.action) {
        case 'create':
          // Only add if task belongs to the current board (or no board filter)
          if (matchesBoard) {
            setTasks((prev) => 
              [...prev, event.record].sort((a, b) => a.position - b.position)
            )
          }
          break
          
        case 'update':
          // If task was moved to a different board, remove it from this view
          if (boardId && taskBoardId !== boardId) {
            setTasks((prev) => prev.filter((t) => t.id !== event.record.id))
          } else {
            // Replace updated task in state
            setTasks((prev) =>
              prev.map((t) => (t.id === event.record.id ? event.record : t))
            )
          }
          break
          
        case 'delete':
          // Remove deleted task from state
          setTasks((prev) => prev.filter((t) => t.id !== event.record.id))
          break
      }
    }).catch((err) => {
      console.error('Failed to subscribe to task updates:', err)
    })

    // Cleanup subscription on unmount
    return () => {
      pb.collection('tasks').unsubscribe('*')
    }
  }, [boardId])

  // Create a new task
  // Note: The backend assigns the sequence number atomically to avoid race conditions.
  // The UI only needs to provide the board ID; seq is assigned server-side.
  const createTask = useCallback(
    async (title: string, column: Column = 'backlog'): Promise<Task> => {
      // Get next position in column
      const columnTasks = tasks.filter((t) => t.column === column)
      const maxPosition = columnTasks.reduce(
        (max, t) => Math.max(max, t.position),
        0
      )
      const position = maxPosition + 1000

      const taskData: Record<string, unknown> = {
        title,
        column,
        position,
        type: 'feature',
        priority: 'medium',
        labels: [],
        created_by: 'user',
      }

      // Add board if boardId is provided
      // Note: seq is assigned by the backend via PocketBase hooks
      if (boardId) {
        taskData.board = boardId
      }

      const task = await pb.collection('tasks').create<Task>(taskData)

      return task
    },
    [tasks, boardId]
  )

  // Update a task
  const updateTask = useCallback(
    async (id: string, data: Partial<Task>): Promise<Task> => {
      return pb.collection('tasks').update<Task>(id, data)
    },
    []
  )

  // Delete a task
  const deleteTask = useCallback(async (id: string): Promise<void> => {
    await pb.collection('tasks').delete(id)
  }, [])

  // Move a task to a new column/position
  const moveTask = useCallback(
    async (id: string, column: Column, position: number): Promise<Task> => {
      return pb.collection('tasks').update<Task>(id, {
        column,
        position,
      })
    },
    []
  )

  return {
    tasks,
    loading,
    error,
    createTask,
    updateTask,
    deleteTask,
    moveTask,
  }
}
