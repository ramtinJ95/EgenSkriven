import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { useComments, extractMentions } from './useComments'
import type { Comment } from '../types/comment'

// Create mock functions we can inspect
const mockCreate = vi.fn()
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
    })),
  },
}))

describe('useComments', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockGetFullList.mockResolvedValue([])
    mockSubscribe.mockResolvedValue(undefined)
  })

  describe('initial state', () => {
    it('starts with loading state', () => {
      const { result } = renderHook(() => useComments('task-123'))
      expect(result.current.loading).toBe(true)
      expect(result.current.comments).toEqual([])
    })

    it('fetches comments on mount', async () => {
      const { result } = renderHook(() => useComments('task-123'))

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })
    })

    it('does not fetch when taskId is empty', async () => {
      const { result } = renderHook(() => useComments(''))

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(mockGetFullList).not.toHaveBeenCalled()
    })

    it('fetches comments with correct filter', async () => {
      renderHook(() => useComments('task-123'))

      await waitFor(() => {
        expect(mockGetFullList).toHaveBeenCalledWith({
          filter: 'task = "task-123"',
          sort: '+created',
        })
      })
    })

    it('sets error state when fetch fails', async () => {
      mockGetFullList.mockRejectedValue(new Error('Network error'))

      const { result } = renderHook(() => useComments('task-123'))

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
        expect(result.current.error).toBeInstanceOf(Error)
      })
    })
  })

  describe('addComment', () => {
    it('creates comment with correct data', async () => {
      const createdComment: Comment = {
        id: 'comment-123',
        collectionId: 'comments',
        collectionName: 'comments',
        task: 'task-123',
        content: 'Test comment',
        author_type: 'human',
        metadata: { mentions: [] },
        created: new Date().toISOString(),
        updated: new Date().toISOString(),
      }
      mockCreate.mockResolvedValue(createdComment)

      const { result } = renderHook(() => useComments('task-123'))

      await waitFor(() => expect(result.current.loading).toBe(false))

      await act(async () => {
        await result.current.addComment({
          task: 'task-123',
          content: 'Test comment',
          author_type: 'human',
        })
      })

      expect(mockCreate).toHaveBeenCalledWith({
        task: 'task-123',
        content: 'Test comment',
        author_type: 'human',
        author_id: undefined,
        metadata: { mentions: [] },
      })
    })

    it('extracts mentions from content', async () => {
      mockCreate.mockResolvedValue({
        id: 'comment-123',
        collectionId: 'comments',
        collectionName: 'comments',
        task: 'task-123',
        content: 'Hello @agent please help',
        author_type: 'human',
        metadata: { mentions: ['@agent'] },
        created: new Date().toISOString(),
        updated: new Date().toISOString(),
      })

      const { result } = renderHook(() => useComments('task-123'))

      await waitFor(() => expect(result.current.loading).toBe(false))

      await act(async () => {
        await result.current.addComment({
          task: 'task-123',
          content: 'Hello @agent please help',
          author_type: 'human',
        })
      })

      expect(mockCreate).toHaveBeenCalledWith(
        expect.objectContaining({
          metadata: { mentions: ['@agent'] },
        })
      )
    })

    it('performs optimistic update', async () => {
      // Make create slow to observe optimistic update
      mockCreate.mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve({
          id: 'server-comment-123',
          collectionId: 'comments',
          collectionName: 'comments',
          task: 'task-123',
          content: 'Test',
          author_type: 'human',
          metadata: { mentions: [] },
          created: new Date().toISOString(),
          updated: new Date().toISOString(),
        }), 100))
      )

      const { result } = renderHook(() => useComments('task-123'))

      await waitFor(() => expect(result.current.loading).toBe(false))

      // Start adding comment (don't await)
      act(() => {
        result.current.addComment({
          task: 'task-123',
          content: 'Test',
          author_type: 'human',
        })
      })

      // Optimistic comment should appear immediately with temp ID
      expect(result.current.comments).toHaveLength(1)
      expect(result.current.comments[0].id).toMatch(/^temp-/)
      expect(result.current.comments[0].content).toBe('Test')
    })

    it('rolls back optimistic update on error', async () => {
      mockCreate.mockRejectedValue(new Error('Server error'))

      const { result } = renderHook(() => useComments('task-123'))

      await waitFor(() => expect(result.current.loading).toBe(false))

      // Try to add comment (will fail)
      await act(async () => {
        try {
          await result.current.addComment({
            task: 'task-123',
            content: 'Test',
            author_type: 'human',
          })
        } catch {
          // Expected error
        }
      })

      // Optimistic comment should be rolled back
      expect(result.current.comments).toHaveLength(0)
    })
  })

  // Test 5.12.1: Real-time subscription updates comments correctly
  describe('real-time subscription', () => {
    it('subscribes to comments updates on mount', async () => {
      renderHook(() => useComments('task-123'))

      await waitFor(() => {
        expect(mockSubscribe).toHaveBeenCalledWith('*', expect.any(Function))
      })
    })

    it('unsubscribes on unmount', async () => {
      const { unmount } = renderHook(() => useComments('task-123'))

      await waitFor(() => {
        expect(mockSubscribe).toHaveBeenCalled()
      })

      unmount()

      expect(mockUnsubscribe).toHaveBeenCalledWith('*')
    })

    it('does not subscribe when taskId is empty', async () => {
      renderHook(() => useComments(''))

      await waitFor(() => {
        expect(mockSubscribe).not.toHaveBeenCalled()
      })
    })

    it('adds new comment to state when create event is received', async () => {
      let subscribeCallback: ((event: { action: string; record: Comment }) => void) | null = null

      mockSubscribe.mockImplementation((_, callback) => {
        subscribeCallback = callback
        return Promise.resolve()
      })

      const { result } = renderHook(() => useComments('task-123'))

      await waitFor(() => expect(result.current.loading).toBe(false))

      // Simulate receiving a create event from another client (e.g., CLI)
      act(() => {
        subscribeCallback?.({
          action: 'create',
          record: {
            id: 'cli-comment',
            collectionId: 'comments',
            collectionName: 'comments',
            task: 'task-123',
            content: 'Comment from CLI',
            author_type: 'agent',
            author_id: 'claude',
            metadata: { mentions: [] },
            created: new Date().toISOString(),
            updated: new Date().toISOString(),
          },
        })
      })

      expect(result.current.comments).toContainEqual(
        expect.objectContaining({ id: 'cli-comment', content: 'Comment from CLI' })
      )
    })

    it('ignores events for different tasks', async () => {
      let subscribeCallback: ((event: { action: string; record: Comment }) => void) | null = null

      mockSubscribe.mockImplementation((_, callback) => {
        subscribeCallback = callback
        return Promise.resolve()
      })

      const { result } = renderHook(() => useComments('task-123'))

      await waitFor(() => expect(result.current.loading).toBe(false))

      // Simulate receiving a create event for a different task
      act(() => {
        subscribeCallback?.({
          action: 'create',
          record: {
            id: 'other-comment',
            collectionId: 'comments',
            collectionName: 'comments',
            task: 'different-task', // Different task
            content: 'Comment for another task',
            author_type: 'human',
            metadata: { mentions: [] },
            created: new Date().toISOString(),
            updated: new Date().toISOString(),
          },
        })
      })

      // Should not add the comment since it's for a different task
      expect(result.current.comments).toHaveLength(0)
    })

    it('updates comment in state when update event is received', async () => {
      const existingComment: Comment = {
        id: 'comment-1',
        collectionId: 'comments',
        collectionName: 'comments',
        task: 'task-123',
        content: 'Original content',
        author_type: 'human',
        metadata: { mentions: [] },
        created: new Date().toISOString(),
        updated: new Date().toISOString(),
      }
      mockGetFullList.mockResolvedValue([existingComment])

      let subscribeCallback: ((event: { action: string; record: Comment }) => void) | null = null

      mockSubscribe.mockImplementation((_, callback) => {
        subscribeCallback = callback
        return Promise.resolve()
      })

      const { result } = renderHook(() => useComments('task-123'))

      await waitFor(() => expect(result.current.comments.length).toBe(1))

      // Simulate receiving an update event
      act(() => {
        subscribeCallback?.({
          action: 'update',
          record: {
            ...existingComment,
            content: 'Updated content',
          },
        })
      })

      expect(result.current.comments[0].content).toBe('Updated content')
    })

    it('removes comment from state when delete event is received', async () => {
      const existingComment: Comment = {
        id: 'comment-1',
        collectionId: 'comments',
        collectionName: 'comments',
        task: 'task-123',
        content: 'To be deleted',
        author_type: 'human',
        metadata: { mentions: [] },
        created: new Date().toISOString(),
        updated: new Date().toISOString(),
      }
      mockGetFullList.mockResolvedValue([existingComment])

      let subscribeCallback: ((event: { action: string; record: Comment }) => void) | null = null

      mockSubscribe.mockImplementation((_, callback) => {
        subscribeCallback = callback
        return Promise.resolve()
      })

      const { result } = renderHook(() => useComments('task-123'))

      await waitFor(() => expect(result.current.comments.length).toBe(1))

      // Simulate receiving a delete event
      act(() => {
        subscribeCallback?.({
          action: 'delete',
          record: existingComment,
        })
      })

      expect(result.current.comments).toHaveLength(0)
    })

    it('deduplicates with optimistic updates on create event', async () => {
      let subscribeCallback: ((event: { action: string; record: Comment }) => void) | null = null

      mockSubscribe.mockImplementation((_, callback) => {
        subscribeCallback = callback
        return Promise.resolve()
      })

      // Make create return after a delay
      const serverComment: Comment = {
        id: 'server-123',
        collectionId: 'comments',
        collectionName: 'comments',
        task: 'task-123',
        content: 'My comment',
        author_type: 'human',
        metadata: { mentions: [] },
        created: new Date().toISOString(),
        updated: new Date().toISOString(),
      }
      mockCreate.mockResolvedValue(serverComment)

      const { result } = renderHook(() => useComments('task-123'))

      await waitFor(() => expect(result.current.loading).toBe(false))

      // Add a comment (creates optimistic update)
      await act(async () => {
        await result.current.addComment({
          task: 'task-123',
          content: 'My comment',
          author_type: 'human',
        })
      })

      // Now simulate receiving the SSE event for the same comment
      act(() => {
        subscribeCallback?.({
          action: 'create',
          record: serverComment,
        })
      })

      // Should still only have 1 comment (deduplicated)
      expect(result.current.comments).toHaveLength(1)
      expect(result.current.comments[0].id).toBe('server-123')
    })

    it('sorts comments by creation date when adding via SSE', async () => {
      const existingComment: Comment = {
        id: 'comment-1',
        collectionId: 'comments',
        collectionName: 'comments',
        task: 'task-123',
        content: 'First comment',
        author_type: 'human',
        metadata: { mentions: [] },
        created: '2024-01-15T10:00:00Z',
        updated: '2024-01-15T10:00:00Z',
      }
      mockGetFullList.mockResolvedValue([existingComment])

      let subscribeCallback: ((event: { action: string; record: Comment }) => void) | null = null

      mockSubscribe.mockImplementation((_, callback) => {
        subscribeCallback = callback
        return Promise.resolve()
      })

      const { result } = renderHook(() => useComments('task-123'))

      await waitFor(() => expect(result.current.comments.length).toBe(1))

      // Add an older comment via SSE (should be sorted before the existing one)
      act(() => {
        subscribeCallback?.({
          action: 'create',
          record: {
            id: 'older-comment',
            collectionId: 'comments',
            collectionName: 'comments',
            task: 'task-123',
            content: 'Older comment',
            author_type: 'agent',
            metadata: { mentions: [] },
            created: '2024-01-15T09:00:00Z', // Earlier timestamp
            updated: '2024-01-15T09:00:00Z',
          },
        })
      })

      expect(result.current.comments).toHaveLength(2)
      // Older comment should be first (sorted by created)
      expect(result.current.comments[0].id).toBe('older-comment')
      expect(result.current.comments[1].id).toBe('comment-1')
    })
  })

  describe('connection error handling', () => {
    it('sets connectionError when subscription fails', async () => {
      mockSubscribe.mockRejectedValue(new Error('SSE connection failed'))

      const { result } = renderHook(() => useComments('task-123'))

      await waitFor(() => {
        expect(result.current.connectionError).toBeInstanceOf(Error)
      })
    })

    it('sets reconnecting state during retry', async () => {
      // First call fails, subsequent calls succeed
      let callCount = 0
      mockSubscribe.mockImplementation(() => {
        callCount++
        if (callCount === 1) {
          return Promise.reject(new Error('SSE connection failed'))
        }
        return Promise.resolve()
      })

      const { result } = renderHook(() => useComments('task-123'))

      // Should eventually start reconnecting
      await waitFor(() => {
        expect(result.current.reconnecting).toBe(true)
      }, { timeout: 2000 })
    })
  })
})

describe('extractMentions', () => {
  it('extracts single mention', () => {
    expect(extractMentions('Hello @agent')).toEqual(['@agent'])
  })

  it('extracts multiple mentions', () => {
    const result = extractMentions('Hello @agent and @user')
    expect(result).toContain('@agent')
    expect(result).toContain('@user')
  })

  it('deduplicates mentions', () => {
    expect(extractMentions('@agent please help @agent')).toEqual(['@agent'])
  })

  it('returns empty array when no mentions', () => {
    expect(extractMentions('No mentions here')).toEqual([])
  })

  it('handles empty string', () => {
    expect(extractMentions('')).toEqual([])
  })

  it('extracts mentions with underscores', () => {
    expect(extractMentions('Hello @user_name')).toEqual(['@user_name'])
  })
})
