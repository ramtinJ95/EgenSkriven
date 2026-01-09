import { useState, useCallback } from 'react'
import { pb } from '../lib/pb'
import type { Task } from '../types/task'
import type { Comment } from '../types/comment'
import type { AgentSession, SessionTool } from '../types/session'

// Debug logging - enabled via VITE_DEBUG_REALTIME=true
const DEBUG_RESUME = import.meta.env.VITE_DEBUG_REALTIME === 'true'
const debugLog = (...args: unknown[]) => {
  if (DEBUG_RESUME) {
    console.log('[useResume]', ...args)
  }
}

// Input for resume operation
export interface ResumeInput {
  taskId: string
  displayId: string
  exec?: boolean // Whether to update task status (actual execution is client-side)
}

// Result from resume operation
export interface ResumeResult {
  command: string
  prompt: string
  tool: SessionTool
  sessionRef: string
  workingDir: string
}

interface UseResumeReturn {
  resume: (input: ResumeInput) => Promise<ResumeResult>
  loading: boolean
  error: Error | null
}

/**
 * Hook for generating resume commands for blocked tasks.
 *
 * Features:
 * - Fetches task, session, and comments data
 * - Builds context prompt with full conversation thread
 * - Generates tool-specific resume command
 * - Optionally updates task status to in_progress
 *
 * @example
 * ```tsx
 * function ResumeButton({ task }: { task: Task }) {
 *   const { resume, loading, error } = useResume()
 *
 *   const handleResume = async () => {
 *     const result = await resume({
 *       taskId: task.id,
 *       displayId: `WRK-${task.seq}`,
 *     })
 *     // Copy result.command to clipboard or display it
 *   }
 *
 *   return (
 *     <button onClick={handleResume} disabled={loading}>
 *       {loading ? 'Generating...' : 'Resume Session'}
 *     </button>
 *   )
 * }
 * ```
 */
export function useResume(): UseResumeReturn {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)

  const resume = useCallback(async (input: ResumeInput): Promise<ResumeResult> => {
    setLoading(true)
    setError(null)

    try {
      debugLog('Starting resume for task:', input.taskId)

      // Fetch the task with session info
      const task = await pb.collection('tasks').getOne<Task>(input.taskId)
      debugLog('Fetched task:', task.title)

      // Get the agent session
      const session = task.agent_session as AgentSession | undefined
      if (!session) {
        throw new Error('No agent session linked to this task')
      }
      debugLog('Session found:', session.tool, session.ref)

      // Fetch comments for context
      const comments = await pb.collection('comments').getFullList<Comment>({
        filter: `task = "${input.taskId}"`,
        sort: '+created', // Oldest first for chronological reading
      })
      debugLog('Fetched', comments.length, 'comments')

      // Build the context prompt
      const prompt = buildContextPrompt(task, input.displayId, comments)
      debugLog('Built prompt, length:', prompt.length)

      // Build the resume command
      const command = buildResumeCommand(session.tool, session.ref, prompt)
      debugLog('Built command for tool:', session.tool)

      // Optionally update task status
      if (input.exec) {
        debugLog('Updating task status to in_progress')
        await pb.collection('tasks').update(input.taskId, {
          column: 'in_progress',
        })
      }

      return {
        command,
        prompt,
        tool: session.tool,
        sessionRef: session.ref,
        workingDir: session.working_dir,
      }
    } catch (err) {
      console.error('[useResume] Failed to generate resume:', err)
      const resumeError = err instanceof Error ? err : new Error('Failed to generate resume command')
      setError(resumeError)
      throw resumeError
    } finally {
      setLoading(false)
    }
  }, [])

  return {
    resume,
    loading,
    error,
  }
}

/**
 * Build context prompt for resuming an agent session.
 *
 * The prompt includes:
 * - Task metadata (ID, title, priority)
 * - Full conversation thread from comments
 * - Instructions for the agent to continue
 *
 * @param task - The task record
 * @param displayId - Human-readable task ID (e.g., "WRK-123")
 * @param comments - Comments on the task, sorted by created date
 * @returns Formatted markdown prompt for the agent
 */
export function buildContextPrompt(
  task: Task,
  displayId: string,
  comments: Comment[]
): string {
  const lines: string[] = []

  // Task context section
  lines.push('## Task Context (from EgenSkriven)')
  lines.push('')
  lines.push(`**Task**: ${displayId} - ${task.title}`)
  lines.push(`**Status**: need_input -> in_progress`)
  lines.push(`**Priority**: ${task.priority}`)
  lines.push('')

  // Description if present
  if (task.description) {
    lines.push('**Description**:')
    lines.push(task.description)
    lines.push('')
  }

  // Conversation thread section
  lines.push('## Conversation Thread')
  lines.push('')

  if (comments.length === 0) {
    lines.push('_No comments yet_')
    lines.push('')
  } else {
    for (const comment of comments) {
      const author = comment.author_id || comment.author_type
      const timestamp = new Date(comment.created).toLocaleTimeString('en-US', {
        hour: '2-digit',
        minute: '2-digit',
      })
      lines.push(`[${author} @ ${timestamp}]: ${comment.content}`)
      lines.push('')
    }
  }

  // Instructions section
  lines.push('## Instructions')
  lines.push('')
  lines.push(
    "Continue working on the task based on the human's response above. " +
      'The conversation context should help you understand what was discussed.'
  )
  lines.push('')

  return lines.join('\n')
}

/**
 * Build the shell command to resume an agent session.
 *
 * Generates tool-specific commands:
 * - OpenCode: `opencode run "<prompt>" --session <ref>`
 * - Claude Code: `claude --resume <ref> "<prompt>"`
 * - Codex: `codex exec resume <ref> "<prompt>"`
 *
 * @param tool - The AI tool (opencode, claude-code, codex)
 * @param ref - The session/thread reference ID
 * @param prompt - The context prompt to inject
 * @returns Shell command string
 */
export function buildResumeCommand(
  tool: SessionTool,
  ref: string,
  prompt: string
): string {
  // Escape prompt for shell - use single quotes and escape embedded single quotes
  const escaped = prompt.replace(/'/g, "'\\''")

  switch (tool) {
    case 'opencode':
      return `opencode run '${escaped}' --session ${ref}`
    case 'claude-code':
      return `claude --resume ${ref} '${escaped}'`
    case 'codex':
      return `codex exec resume ${ref} '${escaped}'`
    default:
      // TypeScript exhaustiveness check
      const _exhaustive: never = tool
      return `# Unknown tool: ${_exhaustive}`
  }
}
