import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { SessionInfo } from './SessionInfo'
import type { AgentSession } from '../types/session'
import type { Column } from '../types/task'

// Mock session data for testing
const mockSession: AgentSession = {
  tool: 'claude-code',
  ref: 'abc123-def456-ghi789-jkl012',
  ref_type: 'uuid',
  working_dir: '/home/user/project',
  linked_at: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(), // 2 hours ago
}

const mockOpencodeSession: AgentSession = {
  tool: 'opencode',
  ref: 'session-xyz-123',
  ref_type: 'uuid',
  working_dir: '/workspace/app',
  linked_at: new Date(Date.now() - 30 * 60 * 1000).toISOString(), // 30 mins ago
}

const mockCodexSession: AgentSession = {
  tool: 'codex',
  ref: 'codex-thread-456',
  ref_type: 'uuid',
  working_dir: '/projects/codex-app',
  linked_at: new Date(Date.now() - 5 * 60 * 1000).toISOString(), // 5 mins ago
}

describe('SessionInfo', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  // Test: Shows nothing when no session
  describe('empty state', () => {
    it('returns null when session is null', () => {
      const { container } = render(
        <SessionInfo session={null} taskColumn="in_progress" />
      )
      expect(container.firstChild).toBeNull()
    })

    it('returns null when session is undefined', () => {
      const { container } = render(
        <SessionInfo session={undefined} taskColumn="in_progress" />
      )
      expect(container.firstChild).toBeNull()
    })
  })

  // Test: Displays tool icon and name
  describe('tool display', () => {
    it('displays Claude Code icon and name', () => {
      render(<SessionInfo session={mockSession} taskColumn="in_progress" />)

      expect(screen.getByText('Claude Code')).toBeInTheDocument()
      // Robot emoji for claude-code
      expect(screen.getByText('🤖')).toBeInTheDocument()
    })

    it('displays OpenCode icon and name', () => {
      render(<SessionInfo session={mockOpencodeSession} taskColumn="in_progress" />)

      expect(screen.getByText('OpenCode')).toBeInTheDocument()
      // Lightning bolt for opencode
      expect(screen.getByText('⚡')).toBeInTheDocument()
    })

    it('displays Codex icon and name', () => {
      render(<SessionInfo session={mockCodexSession} taskColumn="in_progress" />)

      expect(screen.getByText('Codex')).toBeInTheDocument()
      // Crystal ball for codex
      expect(screen.getByText('🔮')).toBeInTheDocument()
    })
  })

  // Test: Shows session ref (truncated)
  describe('session reference display', () => {
    it('displays truncated session ref for long refs', () => {
      render(<SessionInfo session={mockSession} taskColumn="in_progress" />)

      // Should truncate the middle of long refs
      // Original: abc123-def456-ghi789-jkl012 (28 chars)
      // With maxLen 24, should show something like: abc123-def...789-jkl012
      const refElement = screen.getByTitle(mockSession.ref)
      expect(refElement).toBeInTheDocument()
    })

    it('displays full session ref when short', () => {
      const shortSession: AgentSession = {
        ...mockSession,
        ref: 'short-ref',
      }
      render(<SessionInfo session={shortSession} taskColumn="in_progress" />)

      expect(screen.getByText('short-ref')).toBeInTheDocument()
    })
  })

  // Test: Shows linked time
  describe('linked time display', () => {
    it('displays relative linked time', () => {
      render(<SessionInfo session={mockSession} taskColumn="in_progress" />)

      // Should show "2h ago" for session linked 2 hours ago
      expect(screen.getByText(/Linked.*ago/)).toBeInTheDocument()
    })

    it('displays "just now" for recent sessions', () => {
      const recentSession: AgentSession = {
        ...mockSession,
        linked_at: new Date().toISOString(),
      }
      render(<SessionInfo session={recentSession} taskColumn="in_progress" />)

      expect(screen.getByText(/Linked just now/)).toBeInTheDocument()
    })
  })

  // Test: Shows blocked/active status
  describe('status indicator', () => {
    it('shows Active status when task is in_progress', () => {
      render(<SessionInfo session={mockSession} taskColumn="in_progress" />)

      expect(screen.getByText('Active')).toBeInTheDocument()
    })

    it('shows Blocked status when task is need_input', () => {
      render(<SessionInfo session={mockSession} taskColumn="need_input" />)

      expect(screen.getByText('Blocked')).toBeInTheDocument()
    })

    it('shows Active status for other column states', () => {
      const columns: Column[] = ['backlog', 'todo', 'review', 'done']

      columns.forEach((column) => {
        const { unmount } = render(
          <SessionInfo session={mockSession} taskColumn={column} />
        )
        expect(screen.getByText('Active')).toBeInTheDocument()
        unmount()
      })
    })
  })

  // Test: Resume button appears for need_input tasks
  describe('resume button', () => {
    it('shows resume button when task is need_input and onResume provided', () => {
      const onResume = vi.fn()
      render(
        <SessionInfo
          session={mockSession}
          taskColumn="need_input"
          onResume={onResume}
        />
      )

      expect(screen.getByText('Resume Agent Session')).toBeInTheDocument()
    })

    it('calls onResume when resume button is clicked', () => {
      const onResume = vi.fn()
      render(
        <SessionInfo
          session={mockSession}
          taskColumn="need_input"
          onResume={onResume}
        />
      )

      const resumeButton = screen.getByText('Resume Agent Session')
      fireEvent.click(resumeButton)

      expect(onResume).toHaveBeenCalledTimes(1)
    })

    it('hides resume button when task is not need_input', () => {
      const onResume = vi.fn()
      render(
        <SessionInfo
          session={mockSession}
          taskColumn="in_progress"
          onResume={onResume}
        />
      )

      expect(screen.queryByText('Resume Agent Session')).not.toBeInTheDocument()
    })

    it('hides resume button when onResume is not provided', () => {
      render(
        <SessionInfo
          session={mockSession}
          taskColumn="need_input"
        />
      )

      expect(screen.queryByText('Resume Agent Session')).not.toBeInTheDocument()
    })
  })

  // Test: Working directory display
  describe('working directory', () => {
    it('displays working directory', () => {
      render(<SessionInfo session={mockSession} taskColumn="in_progress" />)

      expect(screen.getByText('Working Dir:')).toBeInTheDocument()
      expect(screen.getByText(mockSession.working_dir)).toBeInTheDocument()
    })

    it('truncates long working directory paths', () => {
      const longPathSession: AgentSession = {
        ...mockSession,
        working_dir: '/very/long/path/to/some/deeply/nested/project/directory/here',
      }
      render(<SessionInfo session={longPathSession} taskColumn="in_progress" />)

      // The path should be truncated from the beginning
      const pathElement = screen.getByTitle(longPathSession.working_dir)
      expect(pathElement).toBeInTheDocument()
    })
  })

  // Test: Header display
  describe('header', () => {
    it('displays "Agent Session" header', () => {
      render(<SessionInfo session={mockSession} taskColumn="in_progress" />)

      expect(screen.getByText('Agent Session')).toBeInTheDocument()
    })
  })
})
