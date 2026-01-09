import { useEffect, useState, useCallback } from 'react'
import { pb } from '../lib/pb'
import type { Board } from '../types/board'
import { DEFAULT_COLUMNS } from '../types/board'

interface UseBoardsReturn {
  boards: Board[]
  loading: boolean
  error: Error | null
  createBoard: (input: CreateBoardInput) => Promise<Board>
  deleteBoard: (id: string) => Promise<void>
}

interface CreateBoardInput {
  name: string
  prefix: string
  color?: string
  columns?: string[]
}

// Note: DEFAULT_COLUMNS is imported from types/board.ts to ensure consistency

/**
 * Hook for managing boards with real-time updates.
 *
 * Features:
 * - Fetches all boards on mount
 * - Subscribes to real-time create/update/delete events
 * - Provides CRUD operations
 * - Automatically updates local state on changes
 *
 * @example
 * ```tsx
 * function Sidebar() {
 *   const { boards, loading, createBoard } = useBoards()
 *
 *   if (loading) return <div>Loading...</div>
 *
 *   return <BoardList boards={boards} onCreate={createBoard} />
 * }
 * ```
 */
export function useBoards(): UseBoardsReturn {
  const [boards, setBoards] = useState<Board[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  // Fetch all boards on mount
  useEffect(() => {
    const fetchBoards = async () => {
      try {
        const records = await pb.collection('boards').getFullList<Board>({
          sort: 'name',
        })
        setBoards(records)
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Failed to fetch boards'))
      } finally {
        setLoading(false)
      }
    }

    fetchBoards()
  }, [])

  // Subscribe to real-time updates
  useEffect(() => {
    // Subscribe to all board changes
    pb.collection('boards')
      .subscribe<Board>('*', (event) => {
        switch (event.action) {
          case 'create':
            // Add new board to state and sort by name
            setBoards((prev) =>
              [...prev, event.record].sort((a, b) => a.name.localeCompare(b.name))
            )
            break

          case 'update':
            // Update or add board in state
            setBoards((prev) => {
              const existingIndex = prev.findIndex((b) => b.id === event.record.id)
              if (existingIndex >= 0) {
                // Board exists - replace it
                const updated = [...prev]
                updated[existingIndex] = event.record
                return updated.sort((a, b) => a.name.localeCompare(b.name))
              } else {
                // Board somehow appeared - add it
                return [...prev, event.record].sort((a, b) => a.name.localeCompare(b.name))
              }
            })
            break

          case 'delete':
            // Remove deleted board from state
            setBoards((prev) => prev.filter((b) => b.id !== event.record.id))
            break
        }
      })
      .catch((err) => {
        console.error('Failed to subscribe to board updates:', err)
      })

    // Cleanup subscription on unmount
    return () => {
      pb.collection('boards').unsubscribe('*')
    }
  }, [])

  // Create a new board
  const createBoard = useCallback(
    async (input: CreateBoardInput): Promise<Board> => {
      const board = await pb.collection('boards').create<Board>({
        name: input.name.trim(),
        prefix: input.prefix.toUpperCase().trim(),
        columns: input.columns || DEFAULT_COLUMNS,
        color: input.color || '',
      })

      return board
    },
    []
  )

  // Delete a board
  const deleteBoard = useCallback(async (id: string): Promise<void> => {
    await pb.collection('boards').delete(id)
  }, [])

  return {
    boards,
    loading,
    error,
    createBoard,
    deleteBoard,
  }
}
