import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { CurrentBoardProvider, useCurrentBoard } from './useCurrentBoard'
import type { ReactNode } from 'react'

// Mock boards data
const mockBoards = [
  { id: 'board-1', name: 'Work', prefix: 'WRK', columns: [], color: '#3B82F6', collectionId: 'boards', collectionName: 'boards' },
  { id: 'board-2', name: 'Personal', prefix: 'PER', columns: [], color: '#22C55E', collectionId: 'boards', collectionName: 'boards' },
]

// Mock useBoards hook
let mockBoardsLoading = false
let mockBoardsData = [...mockBoards]

vi.mock('./useBoards', () => ({
  useBoards: () => ({
    boards: mockBoardsData,
    loading: mockBoardsLoading,
    error: null,
    createBoard: vi.fn(),
    deleteBoard: vi.fn(),
  }),
}))

// localStorage is mocked globally in test/setup.ts

// Wrapper component for testing
function wrapper({ children }: { children: ReactNode }) {
  return <CurrentBoardProvider>{children}</CurrentBoardProvider>
}

describe('useCurrentBoard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    mockBoardsLoading = false
    mockBoardsData = [...mockBoards]
  })

  describe('initial state', () => {
    it('throws error when used outside provider', () => {
      // Suppress console.error for this test
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

      expect(() => {
        renderHook(() => useCurrentBoard())
      }).toThrow('useCurrentBoard must be used within a CurrentBoardProvider')

      consoleSpy.mockRestore()
    })

    it('defaults to first board when no localStorage value', async () => {
      const { result } = renderHook(() => useCurrentBoard(), { wrapper })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.currentBoard?.id).toBe('board-1')
    })

    it('restores board from localStorage', async () => {
      localStorage.setItem('egenskriven-current-board', 'board-2')

      const { result } = renderHook(() => useCurrentBoard(), { wrapper })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.currentBoard?.id).toBe('board-2')
    })

    it('falls back to first board if saved board not found', async () => {
      localStorage.setItem('egenskriven-current-board', 'non-existent-board')

      const { result } = renderHook(() => useCurrentBoard(), { wrapper })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.currentBoard?.id).toBe('board-1')
    })

    it('returns loading true while boards are loading', () => {
      mockBoardsLoading = true

      const { result } = renderHook(() => useCurrentBoard(), { wrapper })

      expect(result.current.loading).toBe(true)
    })
  })

  describe('setCurrentBoard', () => {
    it('updates current board', async () => {
      const { result } = renderHook(() => useCurrentBoard(), { wrapper })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      act(() => {
        result.current.setCurrentBoard(mockBoards[1])
      })

      expect(result.current.currentBoard?.id).toBe('board-2')
    })

    it('persists selection to localStorage', async () => {
      const { result } = renderHook(() => useCurrentBoard(), { wrapper })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      act(() => {
        result.current.setCurrentBoard(mockBoards[1])
      })

      expect(localStorage.getItem('egenskriven-current-board')).toBe('board-2')
    })
  })

  describe('board sync', () => {
    it('updates current board when it is modified', async () => {
      const { result, rerender } = renderHook(() => useCurrentBoard(), { wrapper })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      // Simulate board update
      mockBoardsData = [
        { id: 'board-1', name: 'Work Updated', prefix: 'WRK', columns: [], color: '#3B82F6', collectionId: 'boards', collectionName: 'boards' },
        { id: 'board-2', name: 'Personal', prefix: 'PER', columns: [], color: '#22C55E', collectionId: 'boards', collectionName: 'boards' },
      ]

      rerender()

      await waitFor(() => {
        expect(result.current.currentBoard?.name).toBe('Work Updated')
      })
    })

    it('switches to another board when current board is deleted', async () => {
      const { result, rerender } = renderHook(() => useCurrentBoard(), { wrapper })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.currentBoard?.id).toBe('board-1')

      // Simulate board deletion
      mockBoardsData = [
        { id: 'board-2', name: 'Personal', prefix: 'PER', columns: [], color: '#22C55E', collectionId: 'boards', collectionName: 'boards' },
      ]

      rerender()

      await waitFor(() => {
        expect(result.current.currentBoard?.id).toBe('board-2')
      })
    })

    it('sets currentBoard to null when all boards are deleted', async () => {
      const { result, rerender } = renderHook(() => useCurrentBoard(), { wrapper })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.currentBoard?.id).toBe('board-1')

      // Simulate all boards deleted
      mockBoardsData = []

      rerender()

      await waitFor(() => {
        expect(result.current.currentBoard).toBeNull()
      })
    })

    it('clears localStorage when saved board not found', async () => {
      localStorage.setItem('egenskriven-current-board', 'deleted-board')

      renderHook(() => useCurrentBoard(), { wrapper })

      await waitFor(() => {
        // Should have cleared the stale localStorage entry
        expect(localStorage.getItem('egenskriven-current-board')).toBe('board-1')
      })
    })
  })
})
