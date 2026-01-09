import { useEffect, useState, useCallback } from 'react'
import { pb } from '../lib/pb'
import type { Comment, CreateCommentInput } from '../types/comment'

// Debug logging - enabled via VITE_DEBUG_REALTIME=true
const DEBUG_REALTIME = import.meta.env.VITE_DEBUG_REALTIME === 'true'
const debugLog = (...args: unknown[]) => {
  if (DEBUG_REALTIME) {
    console.log('[useComments]', ...args)
  }
}

interface UseCommentsReturn {
  comments: Comment[]
  loading: boolean
  error: Error | null
  addComment: (input: CreateCommentInput) => Promise<Comment>
  adding: boolean
  /** SSE connection error - null means connected, Error means connection failed */
  connectionError: Error | null
  /** Whether the SSE subscription is attempting to reconnect */
  reconnecting: boolean
}

/**
 * Hook for fetching and managing comments for a specific task.
 *
 * Features:
 * - Fetches comments on mount filtered by taskId
 * - Returns comments sorted by creation date (oldest first)
 * - Provides loading and error states
 * - Provides addComment mutation with optimistic updates
 * - Real-time subscription for live updates (create/update/delete)
 *
 * Real-time updates:
 * - Automatically subscribes to PocketBase SSE for comment changes
 * - Handles comments added via CLI or other clients
 * - Intelligently deduplicates with optimistic updates
 * - Cleans up subscription on unmount
 *
 * @param taskId - The task ID to fetch comments for
 *
 * @example
 * ```tsx
 * function CommentsPanel({ taskId }: { taskId: string }) {
 *   const { comments, loading, error, addComment, adding } = useComments(taskId)
 *
 *   if (loading) return <div>Loading...</div>
 *   if (error) return <div>Error: {error.message}</div>
 *
 *   const handleSubmit = async (content: string) => {
 *     await addComment({ task: taskId, content, author_type: 'human' })
 *   }
 *
 *   return (
 *     <div>
 *       {comments.map(comment => (
 *         <Comment key={comment.id} comment={comment} />
 *       ))}
 *       <AddCommentForm onSubmit={handleSubmit} disabled={adding} />
 *     </div>
 *   )
 * }
 * ```
 */
// Maximum retry attempts for SSE reconnection
const MAX_RETRY_ATTEMPTS = 5
// Base delay for exponential backoff (ms)
const BASE_RETRY_DELAY = 1000

export function useComments(taskId: string): UseCommentsReturn {
  const [comments, setComments] = useState<Comment[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)
  const [adding, setAdding] = useState(false)
  const [connectionError, setConnectionError] = useState<Error | null>(null)
  const [reconnecting, setReconnecting] = useState(false)

  // Fetch comments on mount or when taskId changes
  useEffect(() => {
    // Don't fetch if no taskId
    if (!taskId) {
      setComments([])
      setLoading(false)
      return
    }

    const fetchComments = async () => {
      setLoading(true)
      setError(null)

      try {
        debugLog('Fetching comments for task:', taskId)

        const records = await pb.collection('comments').getFullList<Comment>({
          filter: `task = "${taskId}"`,
          sort: '+created', // Oldest first for chronological reading
        })

        debugLog('Fetched', records.length, 'comments')
        setComments(records)
      } catch (err) {
        console.error('[useComments] Failed to fetch comments:', err)
        setError(err instanceof Error ? err : new Error('Failed to fetch comments'))
      } finally {
        setLoading(false)
      }
    }

    fetchComments()
  }, [taskId])

  // Subscribe to real-time comment updates with retry logic
  useEffect(() => {
    // Don't subscribe if no taskId
    if (!taskId) {
      return
    }

    debugLog('Setting up real-time subscription for task:', taskId)

    let isSubscribed = false
    let retryAttempt = 0
    let retryTimeoutId: ReturnType<typeof setTimeout> | null = null
    let isCancelled = false

    // Event handler for SSE events
    const handleEvent = (event: { action: string; record: Comment }) => {
      debugLog('========== COMMENT EVENT RECEIVED ==========')
      debugLog('Action:', event.action)
      debugLog('Record ID:', event.record.id)
      debugLog('Record Task:', event.record.task)
      debugLog('Current Task Filter:', taskId)

      // Only process events for this task
      const commentTaskId = event.record.task
      const matchesTask = commentTaskId === taskId
      debugLog('Matches Task:', matchesTask)

      if (!matchesTask) {
        debugLog('Ignoring event - different task')
        return
      }

      switch (event.action) {
        case 'create':
          debugLog('Adding new comment to state')
          setComments((prev) => {
            // Check if this comment already exists (from optimistic update)
            const exists = prev.some((c) => c.id === event.record.id)
            if (exists) {
              debugLog('Comment already exists, skipping')
              return prev
            }
            // Also check if we have an optimistic version (temp-*) that should be replaced
            const hasOptimistic = prev.some(
              (c) =>
                c.id.startsWith('temp-') &&
                c.content === event.record.content &&
                c.author_type === event.record.author_type
            )
            if (hasOptimistic) {
              debugLog('Replacing optimistic comment with server version')
              return prev.map((c) =>
                c.id.startsWith('temp-') &&
                c.content === event.record.content &&
                c.author_type === event.record.author_type
                  ? event.record
                  : c
              )
            }
            // New comment from elsewhere (CLI, another user) - add it
            const newComments = [...prev, event.record].sort(
              (a, b) => new Date(a.created).getTime() - new Date(b.created).getTime()
            )
            debugLog('New comment count:', newComments.length)
            return newComments
          })
          break

        case 'update':
          debugLog('Updating comment in state')
          setComments((prev) =>
            prev.map((c) => (c.id === event.record.id ? event.record : c))
          )
          break

        case 'delete':
          debugLog('Removing deleted comment')
          setComments((prev) => prev.filter((c) => c.id !== event.record.id))
          break

        default:
          debugLog('Unknown action:', event.action)
      }
      debugLog('=============================================')
    }

    // Subscribe with retry logic
    const subscribe = async () => {
      if (isCancelled) return

      try {
        setReconnecting(retryAttempt > 0)

        await pb.collection('comments').subscribe<Comment>('*', handleEvent)

        isSubscribed = true
        retryAttempt = 0
        setConnectionError(null)
        setReconnecting(false)
        debugLog('Comments subscription established successfully')
      } catch (err) {
        console.error('[useComments] Subscription FAILED:', err)
        const error = err instanceof Error ? err : new Error('SSE connection failed')
        setConnectionError(error)

        // Retry with exponential backoff
        if (retryAttempt < MAX_RETRY_ATTEMPTS && !isCancelled) {
          const delay = BASE_RETRY_DELAY * Math.pow(2, retryAttempt)
          debugLog(`Retrying subscription in ${delay}ms (attempt ${retryAttempt + 1}/${MAX_RETRY_ATTEMPTS})`)
          retryAttempt++
          setReconnecting(true)
          retryTimeoutId = setTimeout(subscribe, delay)
        } else {
          setReconnecting(false)
          if (retryAttempt >= MAX_RETRY_ATTEMPTS) {
            console.error('[useComments] Max retry attempts reached, giving up')
          }
        }
      }
    }

    subscribe()

    // Cleanup subscription on unmount or taskId change
    return () => {
      isCancelled = true
      if (retryTimeoutId) {
        clearTimeout(retryTimeoutId)
      }
      debugLog('Cleaning up comments subscription, was subscribed:', isSubscribed)
      pb.collection('comments').unsubscribe('*')
    }
  }, [taskId])

  // Add a new comment with optimistic update
  const addComment = useCallback(
    async (input: CreateCommentInput): Promise<Comment> => {
      setAdding(true)

      // Extract @mentions from content for metadata
      const mentions = extractMentions(input.content)

      // Create optimistic comment for immediate UI feedback
      const optimisticComment: Comment = {
        id: `temp-${Date.now()}`, // Temporary ID
        collectionId: 'comments',
        collectionName: 'comments',
        task: input.task,
        content: input.content,
        author_type: input.author_type,
        author_id: input.author_id,
        metadata: { mentions },
        created: new Date().toISOString(),
        updated: new Date().toISOString(),
      }

      // Optimistic update - add to local state immediately
      setComments((prev) => [...prev, optimisticComment])
      debugLog('Added optimistic comment:', optimisticComment.id)

      try {
        // Create comment on server
        const created = await pb.collection('comments').create<Comment>({
          task: input.task,
          content: input.content,
          author_type: input.author_type,
          author_id: input.author_id,
          metadata: { mentions },
        })

        debugLog('Comment created on server:', created.id)

        // Replace optimistic comment with server response
        setComments((prev) =>
          prev.map((c) => (c.id === optimisticComment.id ? created : c))
        )

        return created
      } catch (err) {
        console.error('[useComments] Failed to add comment:', err)

        // Rollback - remove optimistic comment
        setComments((prev) => prev.filter((c) => c.id !== optimisticComment.id))

        throw err instanceof Error ? err : new Error('Failed to add comment')
      } finally {
        setAdding(false)
      }
    },
    []
  )

  return {
    comments,
    loading,
    error,
    addComment,
    adding,
    connectionError,
    reconnecting,
  }
}

/**
 * Extract @mentions from comment text.
 *
 * @param text - The comment content to extract mentions from
 * @returns Array of unique mentions (e.g., ['@agent', '@user'])
 *
 * @example
 * extractMentions("Hello @agent, please continue @agent")
 * // Returns: ['@agent']
 */
export function extractMentions(text: string): string[] {
  const matches = text.match(/@\w+/g) || []
  return [...new Set(matches)] // Dedupe
}
