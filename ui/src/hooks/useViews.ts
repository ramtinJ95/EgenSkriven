import { useState, useEffect, useCallback } from 'react'
import { pb } from '../lib/pb'
import type { Filter, MatchMode, DisplayOptions } from '../stores/filters'
import { useFilterStore } from '../stores/filters'

// View record from PocketBase
export interface View {
  id: string
  collectionId: string
  collectionName: string
  created: string
  updated: string
  name: string
  board: string
  filters: Filter[] | string // JSON string or parsed array
  display: DisplayOptions | string // JSON string or parsed object
  is_favorite: boolean
  match_mode: MatchMode
}

// Parsed view with proper types
export interface ParsedView extends Omit<View, 'filters' | 'display'> {
  filters: Filter[]
  display: DisplayOptions
}

// Parse JSON fields from PocketBase
function parseView(view: View): ParsedView {
  return {
    ...view,
    filters: typeof view.filters === 'string' ? JSON.parse(view.filters) : view.filters,
    display: typeof view.display === 'string' ? JSON.parse(view.display) : view.display,
  }
}

export interface UseViewsReturn {
  views: ParsedView[]
  loading: boolean
  error: Error | null
  createView: (name: string) => Promise<ParsedView>
  updateView: (id: string, updates: Partial<{ name: string; is_favorite: boolean }>) => Promise<ParsedView>
  deleteView: (id: string) => Promise<void>
  toggleFavorite: (id: string) => Promise<void>
  applyView: (view: ParsedView) => void
  saveCurrentAsView: (name: string) => Promise<ParsedView>
}

export function useViews(boardId: string | null): UseViewsReturn {
  const [views, setViews] = useState<ParsedView[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  // Get filter store actions
  const loadView = useFilterStore((s) => s.loadView)
  const filters = useFilterStore((s) => s.filters)
  const matchMode = useFilterStore((s) => s.matchMode)
  const displayOptions = useFilterStore((s) => s.displayOptions)

  // Fetch views for the current board
  useEffect(() => {
    if (!boardId) {
      setViews([])
      setLoading(false)
      return
    }

    setLoading(true)
    setError(null)

    // boardId comes from trusted sources (PocketBase record IDs)
    pb.collection('views')
      .getFullList<View>({
        filter: `board = "${boardId}"`,
        sort: '-is_favorite,name',
      })
      .then((records) => {
        setViews(records.map(parseView))
      })
      .catch((err) => {
        console.error('Error fetching views:', err)
        setError(err instanceof Error ? err : new Error('Failed to fetch views'))
      })
      .finally(() => setLoading(false))
  }, [boardId])

  // Subscribe to realtime updates
  useEffect(() => {
    if (!boardId) return

    const handleEvent = (event: { action: string; record: View }) => {
      // Only handle views for the current board
      if (event.record.board !== boardId) return

      if (event.action === 'create') {
        setViews((prev) => [...prev, parseView(event.record)])
      } else if (event.action === 'update') {
        setViews((prev) =>
          prev.map((v) => (v.id === event.record.id ? parseView(event.record) : v))
        )
      } else if (event.action === 'delete') {
        setViews((prev) => prev.filter((v) => v.id !== event.record.id))
      }
    }

    pb.collection('views').subscribe<View>('*', handleEvent)

    return () => {
      pb.collection('views').unsubscribe('*')
    }
  }, [boardId])

  // Create a new view
  const createView = useCallback(
    async (name: string): Promise<ParsedView> => {
      if (!boardId) throw new Error('No board selected')

      const record = await pb.collection('views').create<View>({
        name,
        board: boardId,
        filters: JSON.stringify(filters),
        match_mode: matchMode,
        display: JSON.stringify(displayOptions),
        is_favorite: false,
      })

      return parseView(record)
    },
    [boardId, filters, matchMode, displayOptions]
  )

  // Update an existing view
  const updateView = useCallback(
    async (
      id: string,
      updates: Partial<{ name: string; is_favorite: boolean }>
    ): Promise<ParsedView> => {
      const record = await pb.collection('views').update<View>(id, updates)
      return parseView(record)
    },
    []
  )

  // Delete a view
  const deleteView = useCallback(async (id: string): Promise<void> => {
    await pb.collection('views').delete(id)
  }, [])

  // Toggle favorite status
  const toggleFavorite = useCallback(
    async (id: string): Promise<void> => {
      const view = views.find((v) => v.id === id)
      if (!view) return

      await pb.collection('views').update(id, {
        is_favorite: !view.is_favorite,
      })
    },
    [views]
  )

  // Apply a saved view to the filter state
  const applyView = useCallback(
    (view: ParsedView): void => {
      loadView(view.id, view.filters, view.match_mode, view.display)
    },
    [loadView]
  )

  // Save current filter state as a new view
  const saveCurrentAsView = useCallback(
    async (name: string): Promise<ParsedView> => {
      return createView(name)
    },
    [createView]
  )

  return {
    views,
    loading,
    error,
    createView,
    updateView,
    deleteView,
    toggleFavorite,
    applyView,
    saveCurrentAsView,
  }
}
