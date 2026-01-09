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
}

/**
 * Hook for fetching and managing comments for a specific task.
 *
 * Features:
 * - Fetches comments on mount filtered by taskId
 * - Returns comments sorted by creation date (oldest first)
 * - Provides loading and error states
 * - Provides addComment mutation with optimistic updates
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
export function useComments(taskId: string): UseCommentsReturn {
  const [comments, setComments] = useState<Comment[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)
  const [adding, setAdding] = useState(false)

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
