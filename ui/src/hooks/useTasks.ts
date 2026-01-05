import { useEffect, useState, useCallback } from 'react'
import { pb } from '../lib/pb'
import type { Task, Column } from '../types/task'

// Debug logging - enabled via VITE_DEBUG_REALTIME=true
const DEBUG_REALTIME = import.meta.env.VITE_DEBUG_REALTIME === 'true'
const debugLog = (...args: unknown[]) => {
  if (DEBUG_REALTIME) {
    console.log('[useTasks]', ...args)
  }
}

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
    debugLog('Setting up subscription for board:', boardId)
    
    let isSubscribed = false

    // Subscribe to all task changes
    pb.collection('tasks')
      .subscribe<Task>('*', (event) => {
        debugLog('========== EVENT RECEIVED ==========')
        debugLog('Action:', event.action)
        debugLog('Record ID:', event.record.id)
        debugLog('Record Title:', event.record.title)
        debugLog('Record Board:', event.record.board)
        debugLog('Current Board Filter:', boardId)

        // Filter events by board if boardId is provided
        const taskBoardId = event.record.board
        const matchesBoard = !boardId || taskBoardId === boardId
        debugLog('Matches Board:', matchesBoard)

        switch (event.action) {
          case 'create':
            // Only add if task belongs to the current board (or no board filter)
            if (matchesBoard) {
              debugLog('Adding new task to state')
              setTasks((prev) => {
                const newTasks = [...prev, event.record].sort((a, b) => a.position - b.position)
                debugLog('New task count:', newTasks.length)
                return newTasks
              })
            } else {
              debugLog('Ignoring create - different board')
            }
            break

          case 'update':
            // If task was moved to a different board, remove it from this view
            if (boardId && taskBoardId !== boardId) {
              debugLog('Removing task - moved to different board')
              setTasks((prev) => prev.filter((t) => t.id !== event.record.id))
            } else {
              // Task belongs to this board - update or add it
              setTasks((prev) => {
                const existingIndex = prev.findIndex((t) => t.id === event.record.id)
                debugLog('Task exists in state:', existingIndex >= 0)
                if (existingIndex >= 0) {
                  // Task exists - replace it
                  debugLog('Replacing existing task')
                  const updated = [...prev]
                  updated[existingIndex] = event.record
                  return updated
                } else {
                  // Task moved INTO this board from another board - add it
                  debugLog('Task moved INTO this board - adding')
                  return [...prev, event.record].sort((a, b) => a.position - b.position)
                }
              })
            }
            break

          case 'delete':
            debugLog('Removing deleted task')
            // Remove deleted task from state
            setTasks((prev) => prev.filter((t) => t.id !== event.record.id))
            break

          default:
            debugLog('Unknown action:', event.action)
        }
        debugLog('=====================================')
      })
      .then(() => {
        isSubscribed = true
        debugLog('Subscription established successfully')
      })
      .catch((err) => {
        console.error('[useTasks] Subscription FAILED:', err)
      })

    // Cleanup subscription on unmount
    return () => {
      debugLog('Cleaning up subscription, was subscribed:', isSubscribed)
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

  // Update a task with optimistic update
  const updateTask = useCallback(
    async (id: string, data: Partial<Task>): Promise<Task> => {
      // Optimistic update - update local state immediately
      setTasks((prev) =>
        prev
          .map((t) => (t.id === id ? { ...t, ...data, updated: new Date().toISOString() } : t))
          .sort((a, b) => a.position - b.position)
      )

      try {
        // Send to server
        const updated = await pb.collection('tasks').update<Task>(id, data)

        // Update with server response (includes computed fields like 'updated' timestamp)
        setTasks((prev) => prev.map((t) => (t.id === id ? updated : t)))

        return updated
      } catch (err) {
        // Rollback on error - refetch all tasks
        console.error('[useTasks] Update failed, refetching:', err)
        const options: { sort: string; filter?: string } = { sort: 'position' }
        if (boardId) {
          options.filter = `board = "${boardId}"`
        }
        const records = await pb.collection('tasks').getFullList<Task>(options)
        setTasks(records)
        throw err
      }
    },
    [boardId]
  )

  // Delete a task
  const deleteTask = useCallback(async (id: string): Promise<void> => {
    await pb.collection('tasks').delete(id)
  }, [])

  // Move a task to a new column/position with optimistic update
  const moveTask = useCallback(
    async (id: string, column: Column, position: number): Promise<Task> => {
      // Optimistic update
      setTasks((prev) =>
        prev
          .map((t) => (t.id === id ? { ...t, column, position } : t))
          .sort((a, b) => a.position - b.position)
      )

      try {
        const updated = await pb.collection('tasks').update<Task>(id, {
          column,
          position,
        })

        // Sync with server response
        setTasks((prev) => prev.map((t) => (t.id === id ? updated : t)))

        return updated
      } catch (err) {
        // Rollback - refetch
        console.error('[useTasks] Move failed, refetching:', err)
        const options: { sort: string; filter?: string } = { sort: 'position' }
        if (boardId) {
          options.filter = `board = "${boardId}"`
        }
        const records = await pb.collection('tasks').getFullList<Task>(options)
        setTasks(records)
        throw err
      }
    },
    [boardId]
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
