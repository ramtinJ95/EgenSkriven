import { useEffect, useState, useCallback } from 'react'
import { pb } from '../lib/pb'
import type { Epic } from '../types/epic'

interface UseEpicsReturn {
  epics: Epic[]
  loading: boolean
  error: Error | null
  createEpic: (input: CreateEpicInput) => Promise<Epic>
  updateEpic: (id: string, input: Partial<CreateEpicInput>) => Promise<Epic>
  deleteEpic: (id: string) => Promise<void>
}

interface CreateEpicInput {
  title: string
  description?: string
  color?: string
}

/**
 * Hook for managing epics with real-time updates.
 *
 * Features:
 * - Fetches all epics on mount
 * - Subscribes to real-time create/update/delete events
 * - Provides CRUD operations
 * - Automatically updates local state on changes
 *
 * @example
 * ```tsx
 * function EpicPicker() {
 *   const { epics, loading } = useEpics()
 *
 *   if (loading) return <div>Loading...</div>
 *
 *   return <select>{epics.map(e => <option key={e.id}>{e.title}</option>)}</select>
 * }
 * ```
 */
export function useEpics(): UseEpicsReturn {
  const [epics, setEpics] = useState<Epic[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  // Fetch all epics on mount
  useEffect(() => {
    const fetchEpics = async () => {
      try {
        const records = await pb.collection('epics').getFullList<Epic>({
          sort: 'title',
        })
        setEpics(records)
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Failed to fetch epics'))
      } finally {
        setLoading(false)
      }
    }

    fetchEpics()
  }, [])

  // Subscribe to real-time updates
  useEffect(() => {
    // Subscribe to all epic changes
    pb.collection('epics')
      .subscribe<Epic>('*', (event) => {
        switch (event.action) {
          case 'create':
            // Add new epic to state and sort by title
            setEpics((prev) =>
              [...prev, event.record].sort((a, b) => a.title.localeCompare(b.title))
            )
            break

          case 'update':
            // Update or add epic in state
            setEpics((prev) => {
              const existingIndex = prev.findIndex((e) => e.id === event.record.id)
              if (existingIndex >= 0) {
                // Epic exists - replace it
                const updated = [...prev]
                updated[existingIndex] = event.record
                return updated.sort((a, b) => a.title.localeCompare(b.title))
              } else {
                // Epic somehow appeared - add it
                return [...prev, event.record].sort((a, b) => a.title.localeCompare(b.title))
              }
            })
            break

          case 'delete':
            // Remove deleted epic from state
            setEpics((prev) => prev.filter((e) => e.id !== event.record.id))
            break
        }
      })
      .catch((err) => {
        console.error('Failed to subscribe to epic updates:', err)
        // Surface subscription errors to the component
        setError(err instanceof Error ? err : new Error('Failed to subscribe to epic updates'))
      })

    // Cleanup subscription on unmount
    return () => {
      pb.collection('epics').unsubscribe('*')
    }
  }, [])

  // Create a new epic
  const createEpic = useCallback(
    async (input: CreateEpicInput): Promise<Epic> => {
      const epic = await pb.collection('epics').create<Epic>({
        title: input.title.trim(),
        description: input.description?.trim() || '',
        color: input.color || '#5E6AD2',
      })

      return epic
    },
    []
  )

  // Update an epic
  const updateEpic = useCallback(
    async (id: string, input: Partial<CreateEpicInput>): Promise<Epic> => {
      const data: Record<string, unknown> = {}
      if (input.title !== undefined) data.title = input.title.trim()
      if (input.description !== undefined) data.description = input.description.trim()
      if (input.color !== undefined) data.color = input.color

      const epic = await pb.collection('epics').update<Epic>(id, data)
      return epic
    },
    []
  )

  // Delete an epic
  const deleteEpic = useCallback(async (id: string): Promise<void> => {
    await pb.collection('epics').delete(id)
  }, [])

  return {
    epics,
    loading,
    error,
    createEpic,
    updateEpic,
    deleteEpic,
  }
}
