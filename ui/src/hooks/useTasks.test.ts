import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { useTasks } from './useTasks'

// Create mock functions we can inspect
const mockCreate = vi.fn().mockResolvedValue({ id: 'new-task', title: 'New Task' })
const mockUpdate = vi.fn().mockResolvedValue({ id: 'task-1', title: 'Updated' })
const mockDelete = vi.fn().mockResolvedValue(undefined)
const mockGetFullList = vi.fn().mockResolvedValue([])
const mockSubscribe = vi.fn().mockResolvedValue(undefined)
const mockUnsubscribe = vi.fn()

// Mock PocketBase
vi.mock('../lib/pb', () => ({
  pb: {
    collection: vi.fn(() => ({
      getFullList: mockGetFullList,
      subscribe: mockSubscribe,
      unsubscribe: mockUnsubscribe,
      create: mockCreate,
      update: mockUpdate,
      delete: mockDelete,
    })),
  },
}))

describe('useTasks', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockGetFullList.mockResolvedValue([])
    mockSubscribe.mockResolvedValue(undefined)
  })

  describe('initial state', () => {
    it('starts with loading state', () => {
      const { result } = renderHook(() => useTasks())
      expect(result.current.loading).toBe(true)
      expect(result.current.tasks).toEqual([])
    })

    it('fetches tasks on mount', async () => {
      const { result } = renderHook(() => useTasks())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })
    })

    it('provides CRUD operations', () => {
      const { result } = renderHook(() => useTasks())

      expect(typeof result.current.createTask).toBe('function')
      expect(typeof result.current.updateTask).toBe('function')
      expect(typeof result.current.deleteTask).toBe('function')
      expect(typeof result.current.moveTask).toBe('function')
    })

    it('fetches tasks with position sorting', async () => {
      renderHook(() => useTasks())

      await waitFor(() => {
        expect(mockGetFullList).toHaveBeenCalledWith({ sort: 'position' })
      })
    })

    it('sets error state when fetch fails', async () => {
      mockGetFullList.mockRejectedValue(new Error('Network error'))

      const { result } = renderHook(() => useTasks())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
        expect(result.current.error).toBeInstanceOf(Error)
        expect(result.current.error?.message).toBe('Network error')
      })
    })
  })

  // Test 4.1: createTask calls pb.collection().create with correct params
  describe('createTask', () => {
    it('sends correct data to PocketBase', async () => {
      mockCreate.mockResolvedValue({
        id: 'new-task',
        title: 'New Task Title',
        column: 'todo',
        type: 'feature',
        priority: 'medium',
        labels: [],
        created_by: 'user',
        position: 1000,
      })

      const { result } = renderHook(() => useTasks())

      await waitFor(() => expect(result.current.loading).toBe(false))

      await act(async () => {
        await result.current.createTask('New Task Title', 'todo')
      })

      expect(mockCreate).toHaveBeenCalledWith(
        expect.objectContaining({
          title: 'New Task Title',
          column: 'todo',
          type: 'feature',
          priority: 'medium',
          labels: [],
          created_by: 'user',
        })
      )
    })

    it('uses backlog as default column', async () => {
      mockCreate.mockResolvedValue({ id: 'new-task', title: 'Test' })

      const { result } = renderHook(() => useTasks())

      await waitFor(() => expect(result.current.loading).toBe(false))

      await act(async () => {
        await result.current.createTask('Test Task')
      })

      expect(mockCreate).toHaveBeenCalledWith(
        expect.objectContaining({
          column: 'backlog',
        })
      )
    })

    it('calculates correct position based on existing tasks', async () => {
      mockGetFullList.mockResolvedValue([
        { id: 'task-1', column: 'todo', position: 1000 },
        { id: 'task-2', column: 'todo', position: 2000 },
        { id: 'task-3', column: 'backlog', position: 500 },
      ])
      mockCreate.mockResolvedValue({ id: 'new-task', title: 'Test', position: 3000 })

      const { result } = renderHook(() => useTasks())

      await waitFor(() => expect(result.current.loading).toBe(false))

      await act(async () => {
        await result.current.createTask('New Task', 'todo')
      })

      // Position should be max position in todo column (2000) + 1000 = 3000
      expect(mockCreate).toHaveBeenCalledWith(
        expect.objectContaining({
          position: 3000,
        })
      )
    })

    it('returns the created task', async () => {
      const createdTask = {
        id: 'new-task',
        title: 'New Task',
        column: 'backlog',
        position: 1000,
      }
      mockCreate.mockResolvedValue(createdTask)

      const { result } = renderHook(() => useTasks())

      await waitFor(() => expect(result.current.loading).toBe(false))

      let returnedTask
      await act(async () => {
        returnedTask = await result.current.createTask('New Task')
      })

      expect(returnedTask).toEqual(createdTask)
    })
  })

  // Test 4.2: updateTask calls pb.collection().update
  describe('updateTask', () => {
    it('sends correct data to PocketBase', async () => {
      mockUpdate.mockResolvedValue({ id: 'task-123', priority: 'urgent' })

      const { result } = renderHook(() => useTasks())

      await waitFor(() => expect(result.current.loading).toBe(false))

      await act(async () => {
        await result.current.updateTask('task-123', { priority: 'urgent' })
      })

      expect(mockUpdate).toHaveBeenCalledWith('task-123', { priority: 'urgent' })
    })

    it('can update multiple fields at once', async () => {
      mockUpdate.mockResolvedValue({
        id: 'task-123',
        title: 'Updated Title',
        priority: 'high',
        type: 'bug',
      })

      const { result } = renderHook(() => useTasks())

      await waitFor(() => expect(result.current.loading).toBe(false))

      await act(async () => {
        await result.current.updateTask('task-123', {
          title: 'Updated Title',
          priority: 'high',
          type: 'bug',
        })
      })

      expect(mockUpdate).toHaveBeenCalledWith('task-123', {
        title: 'Updated Title',
        priority: 'high',
        type: 'bug',
      })
    })

    it('returns the updated task', async () => {
      const updatedTask = { id: 'task-123', title: 'Updated' }
      mockUpdate.mockResolvedValue(updatedTask)

      const { result } = renderHook(() => useTasks())

      await waitFor(() => expect(result.current.loading).toBe(false))

      let returnedTask
      await act(async () => {
        returnedTask = await result.current.updateTask('task-123', { title: 'Updated' })
      })

      expect(returnedTask).toEqual(updatedTask)
    })
  })

  // Test 4.3: deleteTask calls pb.collection().delete
  describe('deleteTask', () => {
    it('calls PocketBase delete with correct ID', async () => {
      mockDelete.mockResolvedValue(undefined)

      const { result } = renderHook(() => useTasks())

      await waitFor(() => expect(result.current.loading).toBe(false))

      await act(async () => {
        await result.current.deleteTask('task-123')
      })

      expect(mockDelete).toHaveBeenCalledWith('task-123')
    })
  })

  // Test 4.4: moveTask updates column and position
  describe('moveTask', () => {
    it('updates column and position', async () => {
      mockUpdate.mockResolvedValue({
        id: 'task-123',
        column: 'in_progress',
        position: 5000,
      })

      const { result } = renderHook(() => useTasks())

      await waitFor(() => expect(result.current.loading).toBe(false))

      await act(async () => {
        await result.current.moveTask('task-123', 'in_progress', 5000)
      })

      expect(mockUpdate).toHaveBeenCalledWith('task-123', {
        column: 'in_progress',
        position: 5000,
      })
    })

    it('returns the updated task', async () => {
      const movedTask = { id: 'task-123', column: 'done', position: 1000 }
      mockUpdate.mockResolvedValue(movedTask)

      const { result } = renderHook(() => useTasks())

      await waitFor(() => expect(result.current.loading).toBe(false))

      let returnedTask
      await act(async () => {
        returnedTask = await result.current.moveTask('task-123', 'done', 1000)
      })

      expect(returnedTask).toEqual(movedTask)
    })
  })

  // Test 4.5: Real-time subscription updates state correctly
  describe('real-time subscription', () => {
    it('subscribes to task updates on mount', async () => {
      renderHook(() => useTasks())

      await waitFor(() => {
        expect(mockSubscribe).toHaveBeenCalledWith('*', expect.any(Function))
      })
    })

    it('unsubscribes on unmount', async () => {
      const { unmount } = renderHook(() => useTasks())

      await waitFor(() => {
        expect(mockSubscribe).toHaveBeenCalled()
      })

      unmount()

      expect(mockUnsubscribe).toHaveBeenCalledWith('*')
    })

    it('adds new task to state when create event is received', async () => {
      let subscribeCallback: ((event: { action: string; record: unknown }) => void) | null = null

      mockSubscribe.mockImplementation((_, callback) => {
        subscribeCallback = callback
        return Promise.resolve()
      })

      const { result } = renderHook(() => useTasks())

      await waitFor(() => expect(result.current.loading).toBe(false))

      // Simulate receiving a create event
      act(() => {
        subscribeCallback?.({
          action: 'create',
          record: { id: 'new-task', title: 'New Task', column: 'backlog', position: 1000 },
        })
      })

      expect(result.current.tasks).toContainEqual(
        expect.objectContaining({ id: 'new-task', title: 'New Task' })
      )
    })

    it('updates task in state when update event is received', async () => {
      mockGetFullList.mockResolvedValue([
        { id: 'task-1', title: 'Original Title', column: 'todo', position: 1000 },
      ])

      let subscribeCallback: ((event: { action: string; record: unknown }) => void) | null = null

      mockSubscribe.mockImplementation((_, callback) => {
        subscribeCallback = callback
        return Promise.resolve()
      })

      const { result } = renderHook(() => useTasks())

      await waitFor(() => expect(result.current.tasks.length).toBe(1))

      // Simulate receiving an update event
      act(() => {
        subscribeCallback?.({
          action: 'update',
          record: { id: 'task-1', title: 'Updated Title', column: 'todo', position: 1000 },
        })
      })

      expect(result.current.tasks[0].title).toBe('Updated Title')
    })

    it('removes task from state when delete event is received', async () => {
      mockGetFullList.mockResolvedValue([
        { id: 'task-1', title: 'To Delete', column: 'backlog', position: 1000 },
      ])

      let subscribeCallback: ((event: { action: string; record: unknown }) => void) | null = null

      mockSubscribe.mockImplementation((_, callback) => {
        subscribeCallback = callback
        return Promise.resolve()
      })

      const { result } = renderHook(() => useTasks())

      await waitFor(() => expect(result.current.tasks.length).toBe(1))

      // Simulate receiving a delete event
      act(() => {
        subscribeCallback?.({
          action: 'delete',
          record: { id: 'task-1' },
        })
      })

      expect(result.current.tasks).toHaveLength(0)
    })

    it('handles multiple create events correctly', async () => {
      let subscribeCallback: ((event: { action: string; record: unknown }) => void) | null = null

      mockSubscribe.mockImplementation((_, callback) => {
        subscribeCallback = callback
        return Promise.resolve()
      })

      const { result } = renderHook(() => useTasks())

      await waitFor(() => expect(result.current.loading).toBe(false))

      // Simulate receiving multiple create events
      act(() => {
        subscribeCallback?.({
          action: 'create',
          record: { id: 'task-1', title: 'Task 1', column: 'backlog', position: 1000 },
        })
      })

      act(() => {
        subscribeCallback?.({
          action: 'create',
          record: { id: 'task-2', title: 'Task 2', column: 'todo', position: 2000 },
        })
      })

      expect(result.current.tasks).toHaveLength(2)
      expect(result.current.tasks).toContainEqual(expect.objectContaining({ id: 'task-1' }))
      expect(result.current.tasks).toContainEqual(expect.objectContaining({ id: 'task-2' }))
    })
  })
})
