import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { useBoards } from './useBoards'

// Create mock functions we can inspect
const mockCreate = vi.fn()
const mockDelete = vi.fn()
const mockGetFullList = vi.fn()
const mockSubscribe = vi.fn()
const mockUnsubscribe = vi.fn()

// Mock PocketBase
vi.mock('../lib/pb', () => ({
  pb: {
    collection: vi.fn(() => ({
      getFullList: mockGetFullList,
      subscribe: mockSubscribe,
      unsubscribe: mockUnsubscribe,
      create: mockCreate,
      delete: mockDelete,
    })),
  },
}))

describe('useBoards', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockGetFullList.mockResolvedValue([])
    mockSubscribe.mockResolvedValue(undefined)
  })

  describe('initial state', () => {
    it('starts with loading state', () => {
      const { result } = renderHook(() => useBoards())
      expect(result.current.loading).toBe(true)
      expect(result.current.boards).toEqual([])
    })

    it('fetches boards on mount', async () => {
      const { result } = renderHook(() => useBoards())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })
    })

    it('provides CRUD operations', () => {
      const { result } = renderHook(() => useBoards())

      expect(typeof result.current.createBoard).toBe('function')
      expect(typeof result.current.deleteBoard).toBe('function')
    })

    it('fetches boards sorted by name', async () => {
      renderHook(() => useBoards())

      await waitFor(() => {
        expect(mockGetFullList).toHaveBeenCalledWith({ sort: 'name' })
      })
    })

    it('sets error state when fetch fails', async () => {
      mockGetFullList.mockRejectedValue(new Error('Network error'))

      const { result } = renderHook(() => useBoards())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
        expect(result.current.error).toBeInstanceOf(Error)
        expect(result.current.error?.message).toBe('Network error')
      })
    })
  })

  describe('createBoard', () => {
    it('sends correct data to PocketBase', async () => {
      mockCreate.mockResolvedValue({
        id: 'new-board',
        name: 'Work',
        prefix: 'WRK',
        columns: ['backlog', 'todo', 'in_progress', 'review', 'done'],
        color: '#3B82F6',
      })

      const { result } = renderHook(() => useBoards())

      await waitFor(() => expect(result.current.loading).toBe(false))

      await act(async () => {
        await result.current.createBoard({
          name: 'Work',
          prefix: 'wrk', // lowercase should be uppercased
          color: '#3B82F6',
        })
      })

      expect(mockCreate).toHaveBeenCalledWith(
        expect.objectContaining({
          name: 'Work',
          prefix: 'WRK',
          color: '#3B82F6',
        })
      )
    })

    it('uses default columns when not provided', async () => {
      mockCreate.mockResolvedValue({ id: 'new-board', name: 'Test', prefix: 'TST' })

      const { result } = renderHook(() => useBoards())

      await waitFor(() => expect(result.current.loading).toBe(false))

      await act(async () => {
        await result.current.createBoard({ name: 'Test', prefix: 'TST' })
      })

      expect(mockCreate).toHaveBeenCalledWith(
        expect.objectContaining({
          columns: ['backlog', 'todo', 'in_progress', 'review', 'done'],
        })
      )
    })

    it('returns the created board', async () => {
      const createdBoard = {
        id: 'new-board',
        name: 'Work',
        prefix: 'WRK',
        columns: ['backlog', 'todo', 'done'],
        color: '#3B82F6',
      }
      mockCreate.mockResolvedValue(createdBoard)

      const { result } = renderHook(() => useBoards())

      await waitFor(() => expect(result.current.loading).toBe(false))

      let returnedBoard
      await act(async () => {
        returnedBoard = await result.current.createBoard({
          name: 'Work',
          prefix: 'WRK',
        })
      })

      expect(returnedBoard).toEqual(createdBoard)
    })
  })

  describe('deleteBoard', () => {
    it('calls PocketBase delete with correct ID', async () => {
      mockDelete.mockResolvedValue(undefined)

      const { result } = renderHook(() => useBoards())

      await waitFor(() => expect(result.current.loading).toBe(false))

      await act(async () => {
        await result.current.deleteBoard('board-123')
      })

      expect(mockDelete).toHaveBeenCalledWith('board-123')
    })
  })

  describe('real-time subscription', () => {
    it('subscribes to board updates on mount', async () => {
      renderHook(() => useBoards())

      await waitFor(() => {
        expect(mockSubscribe).toHaveBeenCalledWith('*', expect.any(Function))
      })
    })

    it('unsubscribes on unmount', async () => {
      const { unmount } = renderHook(() => useBoards())

      await waitFor(() => {
        expect(mockSubscribe).toHaveBeenCalled()
      })

      unmount()

      expect(mockUnsubscribe).toHaveBeenCalledWith('*')
    })

    it('adds new board to state when create event is received', async () => {
      let subscribeCallback: ((event: { action: string; record: unknown }) => void) | null = null

      mockSubscribe.mockImplementation((_, callback) => {
        subscribeCallback = callback
        return Promise.resolve()
      })

      const { result } = renderHook(() => useBoards())

      await waitFor(() => expect(result.current.loading).toBe(false))

      act(() => {
        subscribeCallback?.({
          action: 'create',
          record: { id: 'new-board', name: 'Work', prefix: 'WRK', columns: [] },
        })
      })

      expect(result.current.boards).toContainEqual(
        expect.objectContaining({ id: 'new-board', name: 'Work' })
      )
    })

    it('updates board in state when update event is received', async () => {
      mockGetFullList.mockResolvedValue([
        { id: 'board-1', name: 'Original Name', prefix: 'ORG', columns: [] },
      ])

      let subscribeCallback: ((event: { action: string; record: unknown }) => void) | null = null

      mockSubscribe.mockImplementation((_, callback) => {
        subscribeCallback = callback
        return Promise.resolve()
      })

      const { result } = renderHook(() => useBoards())

      await waitFor(() => expect(result.current.boards.length).toBe(1))

      act(() => {
        subscribeCallback?.({
          action: 'update',
          record: { id: 'board-1', name: 'Updated Name', prefix: 'ORG', columns: [] },
        })
      })

      expect(result.current.boards[0].name).toBe('Updated Name')
    })

    it('removes board from state when delete event is received', async () => {
      mockGetFullList.mockResolvedValue([
        { id: 'board-1', name: 'To Delete', prefix: 'DEL', columns: [] },
      ])

      let subscribeCallback: ((event: { action: string; record: unknown }) => void) | null = null

      mockSubscribe.mockImplementation((_, callback) => {
        subscribeCallback = callback
        return Promise.resolve()
      })

      const { result } = renderHook(() => useBoards())

      await waitFor(() => expect(result.current.boards.length).toBe(1))

      act(() => {
        subscribeCallback?.({
          action: 'delete',
          record: { id: 'board-1' },
        })
      })

      expect(result.current.boards).toHaveLength(0)
    })

    it('keeps boards sorted by name after create event', async () => {
      mockGetFullList.mockResolvedValue([
        { id: 'board-1', name: 'Zebra', prefix: 'ZEB', columns: [] },
      ])

      let subscribeCallback: ((event: { action: string; record: unknown }) => void) | null = null

      mockSubscribe.mockImplementation((_, callback) => {
        subscribeCallback = callback
        return Promise.resolve()
      })

      const { result } = renderHook(() => useBoards())

      await waitFor(() => expect(result.current.boards.length).toBe(1))

      act(() => {
        subscribeCallback?.({
          action: 'create',
          record: { id: 'board-2', name: 'Apple', prefix: 'APL', columns: [] },
        })
      })

      expect(result.current.boards[0].name).toBe('Apple')
      expect(result.current.boards[1].name).toBe('Zebra')
    })
  })
})
