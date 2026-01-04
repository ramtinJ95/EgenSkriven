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
 * - Fetches all tasks on mount
 * - Subscribes to real-time create/update/delete events
 * - Provides CRUD operations
 * - Automatically updates local state on changes
 * 
 * @example
 * ```tsx
 * function Board() {
 *   const { tasks, loading, moveTask } = useTasks()
 *   
 *   if (loading) return <div>Loading...</div>
 *   
 *   return <BoardView tasks={tasks} onMove={moveTask} />
 * }
 * ```
 */
export function useTasks(): UseTasksReturn {
  const [tasks, setTasks] = useState<Task[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  // Fetch all tasks on mount
  useEffect(() => {
    const fetchTasks = async () => {
      try {
        const records = await pb.collection('tasks').getFullList<Task>({
          sort: 'position',
        })
        setTasks(records)
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Failed to fetch tasks'))
      } finally {
        setLoading(false)
      }
    }

    fetchTasks()
  }, [])

  // Subscribe to real-time updates
  useEffect(() => {
    // Subscribe to all task changes
    pb.collection('tasks').subscribe<Task>('*', (event) => {
      switch (event.action) {
        case 'create':
          // Add new task to state
          setTasks((prev) => [...prev, event.record])
          break
          
        case 'update':
          // Replace updated task in state
          setTasks((prev) =>
            prev.map((t) => (t.id === event.record.id ? event.record : t))
          )
          break
          
        case 'delete':
          // Remove deleted task from state
          setTasks((prev) => prev.filter((t) => t.id !== event.record.id))
          break
      }
    })

    // Cleanup subscription on unmount
    return () => {
      pb.collection('tasks').unsubscribe('*')
    }
  }, [])

  // Create a new task
  const createTask = useCallback(
    async (title: string, column: Column = 'backlog'): Promise<Task> => {
      // Get next position in column
      const columnTasks = tasks.filter((t) => t.column === column)
      const maxPosition = columnTasks.reduce(
        (max, t) => Math.max(max, t.position),
        0
      )
      const position = maxPosition + 1000

      const task = await pb.collection('tasks').create<Task>({
        title,
        column,
        position,
        type: 'feature',
        priority: 'medium',
        labels: [],
        created_by: 'user',
      })

      return task
    },
    [tasks]
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
