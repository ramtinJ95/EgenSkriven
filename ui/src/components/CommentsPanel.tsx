import { useState, useRef, useEffect } from 'react'
import { useComments, extractMentions } from '../hooks/useComments'
import type { Comment } from '../types/comment'
import styles from './CommentsPanel.module.css'

interface CommentsPanelProps {
  taskId: string
}

/**
 * CommentsPanel displays and manages comments for a task.
 *
 * Features:
 * - Lists all comments for the task with real-time updates
 * - Shows agent vs human comments with different styling
 * - Loading and empty states
 * - Comment submission form
 * - @agent mention warning indicator
 */
export function CommentsPanel({ taskId }: CommentsPanelProps) {
  const { comments, loading, error, addComment, adding, connectionError, reconnecting } = useComments(taskId)
  const [newComment, setNewComment] = useState('')
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  // Auto-resize textarea
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
      textareaRef.current.style.height = `${textareaRef.current.scrollHeight}px`
    }
  }, [newComment])

  // Handle form submission
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    const trimmed = newComment.trim()
    if (!trimmed || adding) return

    try {
      await addComment({
        task: taskId,
        content: trimmed,
        author_type: 'human',
      })
      setNewComment('')
    } catch (err) {
      console.error('Failed to add comment:', err)
    }
  }

  // Handle Ctrl+Enter to submit
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
      e.preventDefault()
      handleSubmit(e)
    }
  }

  // Check for @agent mention
  const mentions = extractMentions(newComment)
  const hasAgentMention = mentions.includes('@agent')

  // Loading state
  if (loading) {
    return (
      <section className={styles.commentsPanel} aria-labelledby="comments-heading" aria-busy="true">
        <h3 id="comments-heading" className={styles.header}>Comments</h3>
        <div className={styles.loading} role="status" aria-label="Loading comments">
          <div className={styles.skeleton} />
          <div className={styles.skeleton} />
          <div className={styles.skeleton} />
        </div>
      </section>
    )
  }

  // Error state
  if (error) {
    return (
      <section className={styles.commentsPanel} aria-labelledby="comments-heading-error">
        <h3 id="comments-heading-error" className={styles.header}>Comments</h3>
        <div className={styles.error} role="alert">Failed to load comments</div>
      </section>
    )
  }

  return (
    <section className={styles.commentsPanel} aria-labelledby="comments-heading-main">
      <h3 id="comments-heading-main" className={styles.header}>
        Comments
        {comments.length > 0 && (
          <span className={styles.count} aria-label={`${comments.length} comments`}>({comments.length})</span>
        )}
      </h3>

      {/* Connection warning */}
      {(connectionError || reconnecting) && (
        <div className={styles.connectionWarning} role="alert">
          <span className={styles.connectionIcon} aria-hidden="true">
            {reconnecting ? '\u{1F504}' : '\u26A0\uFE0F'}
          </span>
          <span>
            {reconnecting
              ? 'Reconnecting to real-time updates...'
              : 'Real-time updates unavailable. New comments may not appear automatically.'}
          </span>
        </div>
      )}

      {/* Comments list */}
      <div className={styles.list} role="list" aria-label="Comment thread">
        {comments.length === 0 ? (
          <div className={styles.empty} role="status">No comments yet</div>
        ) : (
          comments.map((comment) => (
            <CommentItem key={comment.id} comment={comment} />
          ))
        )}
      </div>

      {/* Add comment form */}
      <form onSubmit={handleSubmit} className={styles.form} aria-label="Add comment">
        <textarea
          ref={textareaRef}
          value={newComment}
          onChange={(e) => setNewComment(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Add a comment..."
          className={styles.textarea}
          rows={2}
          disabled={adding}
          aria-label="Comment text"
          aria-describedby={hasAgentMention ? 'agent-mention-warning' : undefined}
        />

        {/* @agent mention warning */}
        {hasAgentMention && (
          <div id="agent-mention-warning" className={styles.mentionWarning} role="alert">
            <span className={styles.warningIcon} aria-hidden="true">{'\u26A0\uFE0F'}</span>
            Will trigger auto-resume (if enabled)
          </div>
        )}

        <div className={styles.formFooter}>
          <span className={styles.hint} aria-hidden="true">Ctrl+Enter to submit</span>
          <button
            type="submit"
            disabled={!newComment.trim() || adding}
            className={styles.submitButton}
            aria-busy={adding}
          >
            {adding ? 'Adding...' : 'Add Comment'}
          </button>
        </div>
      </form>
    </section>
  )
}

/**
 * Individual comment item with agent/human styling.
 */
function CommentItem({ comment }: { comment: Comment }) {
  const isAgent = comment.author_type === 'agent'
  const author = comment.author_id || comment.author_type
  const timeAgo = formatTimeAgo(comment.created)

  // Extract and highlight mentions in content
  const mentions = comment.metadata?.mentions || []

  return (
    <article
      className={`${styles.item} ${isAgent ? styles.agentItem : styles.humanItem}`}
      role="listitem"
      aria-label={`Comment by ${author}`}
      tabIndex={0}
    >
      {/* Header with author and time */}
      <div className={styles.itemHeader}>
        <span className={styles.authorBadge} data-type={comment.author_type}>
          <span className={styles.authorIcon} aria-hidden="true">
            {isAgent ? '\u{1F916}' : '\u{1F464}'}
          </span>
          {author}
        </span>
        <time className={styles.time} dateTime={comment.created}>{timeAgo}</time>
        {mentions.length > 0 && (
          <span className={styles.mentions} aria-label="Mentions">
            {mentions.map((m) => (
              <span key={m} className={styles.mention}>
                {m}
              </span>
            ))}
          </span>
        )}
      </div>

      {/* Content */}
      <div className={styles.itemContent}>{comment.content}</div>
    </article>
  )
}

/**
 * Format timestamp as relative time (e.g., "2h ago").
 */
function formatTimeAgo(timestamp: string): string {
  const date = new Date(timestamp)
  if (isNaN(date.getTime())) return '-'

  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)

  if (diffMins < 1) return 'just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  if (diffDays < 7) return `${diffDays}d ago`

  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined,
  })
}
