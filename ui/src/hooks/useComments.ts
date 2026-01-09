import { useEffect, useState } from 'react'
import { pb } from '../lib/pb'
import type { Comment } from '../types/comment'

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
}

/**
 * Hook for fetching comments for a specific task.
 *
 * Features:
 * - Fetches comments on mount filtered by taskId
 * - Returns comments sorted by creation date (oldest first)
 * - Provides loading and error states
 *
 * @param taskId - The task ID to fetch comments for
 *
 * @example
 * ```tsx
 * function CommentsPanel({ taskId }: { taskId: string }) {
 *   const { comments, loading, error } = useComments(taskId)
 *
 *   if (loading) return <div>Loading...</div>
 *   if (error) return <div>Error: {error.message}</div>
 *
 *   return (
 *     <div>
 *       {comments.map(comment => (
 *         <Comment key={comment.id} comment={comment} />
 *       ))}
 *     </div>
 *   )
 * }
 * ```
 */
export function useComments(taskId: string): UseCommentsReturn {
  const [comments, setComments] = useState<Comment[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

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

  return {
    comments,
    loading,
    error,
  }
}
