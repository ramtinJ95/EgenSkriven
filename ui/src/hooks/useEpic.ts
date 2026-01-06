import { useEffect, useState, useCallback } from 'react'
import { pb } from '../lib/pb'
import type { Epic } from '../types/epic'

interface UseEpicReturn {
  epic: Epic | null
  loading: boolean
  error: Error | null
  updateEpic: (data: Partial<Epic>) => Promise<Epic>
  deleteEpic: () => Promise<void>
  refetch: () => Promise<void>
}

/**
 * Hook for managing a single epic with CRUD operations.
 *
 * Features:
 * - Fetches a single epic by ID
 * - Provides update and delete operations
 * - Subscribes to real-time updates for this epic
 * - Handles loading and error states
 *
 * @example
 * ```tsx
 * function EpicDetail({ epicId }: { epicId: string }) {
 *   const { epic, loading, updateEpic, deleteEpic } = useEpic(epicId)
 *
 *   if (loading) return <div>Loading...</div>
 *   if (!epic) return <div>Epic not found</div>
 *
 *   return (
 *     <div>
 *       <h1>{epic.title}</h1>
 *       <button onClick={() => updateEpic({ title: 'New Title' })}>
 *         Rename
 *       </button>
 *       <button onClick={deleteEpic}>Delete</button>
 *     </div>
 *   )
 * }
 * ```
 */
export function useEpic(epicId: string | null): UseEpicReturn {
  const [epic, setEpic] = useState<Epic | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  // Fetch epic on mount or when epicId changes
  const fetchEpic = useCallback(async () => {
    if (!epicId) {
      setEpic(null)
      setLoading(false)
      return
    }

    setLoading(true)
    setError(null)

    try {
      const record = await pb.collection('epics').getOne<Epic>(epicId)
      setEpic(record)
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to fetch epic'))
      setEpic(null)
    } finally {
      setLoading(false)
    }
  }, [epicId])

  useEffect(() => {
    fetchEpic()
  }, [fetchEpic])

  // Subscribe to real-time updates for this specific epic
  useEffect(() => {
    if (!epicId) return

    pb.collection('epics')
      .subscribe<Epic>(epicId, (event) => {
        if (event.action === 'update') {
          setEpic(event.record)
        } else if (event.action === 'delete') {
          setEpic(null)
        }
      })
      .catch((err) => {
        console.error('Failed to subscribe to epic updates:', err)
      })

    // Cleanup subscription on unmount or epicId change
    return () => {
      pb.collection('epics').unsubscribe(epicId)
    }
  }, [epicId])

  // Update epic
  const updateEpic = useCallback(
    async (data: Partial<Epic>): Promise<Epic> => {
      if (!epicId) {
        throw new Error('No epic ID provided')
      }

      const updateData: Record<string, unknown> = {}
      if (data.title !== undefined) updateData.title = data.title.trim()
      if (data.description !== undefined) updateData.description = data.description.trim()
      if (data.color !== undefined) updateData.color = data.color

      const updated = await pb.collection('epics').update<Epic>(epicId, updateData)
      setEpic(updated)
      return updated
    },
    [epicId]
  )

  // Delete epic
  const deleteEpic = useCallback(async (): Promise<void> => {
    if (!epicId) {
      throw new Error('No epic ID provided')
    }

    await pb.collection('epics').delete(epicId)
    setEpic(null)
  }, [epicId])

  return {
    epic,
    loading,
    error,
    updateEpic,
    deleteEpic,
    refetch: fetchEpic,
  }
}
