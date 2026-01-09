import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { CommentsPanel } from './CommentsPanel'
import type { Comment } from '../types/comment'

// Mock the useComments hook
vi.mock('../hooks/useComments', () => ({
  useComments: vi.fn(),
  extractMentions: (text: string) => {
    const matches = text.match(/@\w+/g) || []
    return [...new Set(matches)]
  },
}))

import { useComments } from '../hooks/useComments'

const mockedUseComments = vi.mocked(useComments)

// Default mock return value with all required fields
const defaultMockReturn = {
  comments: [] as Comment[],
  loading: false,
  error: null,
  addComment: vi.fn(),
  adding: false,
  connectionError: null,
  reconnecting: false,
}

// Mock comments for testing
const mockComments: Comment[] = [
  {
    id: 'comment-1',
    collectionId: 'comments',
    collectionName: 'comments',
    task: 'task-123',
    content: 'First comment from agent',
    author_type: 'agent',
    author_id: 'claude',
    metadata: { mentions: [] },
    created: '2024-01-15T10:00:00Z',
    updated: '2024-01-15T10:00:00Z',
  },
  {
    id: 'comment-2',
    collectionId: 'comments',
    collectionName: 'comments',
    task: 'task-123',
    content: 'Reply from human with @agent mention',
    author_type: 'human',
    metadata: { mentions: ['@agent'] },
    created: '2024-01-15T11:00:00Z',
    updated: '2024-01-15T11:00:00Z',
  },
]

describe('CommentsPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  // 5.10.2: Test loading state displays correctly
  describe('loading state', () => {
    it('displays loading skeleton when loading', () => {
      mockedUseComments.mockReturnValue({
        ...defaultMockReturn,
        loading: true,
      })

      render(<CommentsPanel taskId="task-123" />)

      // Should show header
      expect(screen.getByText('Comments')).toBeInTheDocument()

      // Should show loading skeletons (they have the skeleton class)
      const container = document.querySelector('[class*="loading"]')
      expect(container).toBeInTheDocument()
    })
  })

  // 5.10.3: Test empty state shows when no comments
  describe('empty state', () => {
    it('shows empty message when no comments', () => {
      mockedUseComments.mockReturnValue({
        ...defaultMockReturn,
      })

      render(<CommentsPanel taskId="task-123" />)

      expect(screen.getByText('No comments yet')).toBeInTheDocument()
    })
  })

  // 5.10.4: Test comments display with author, time, content
  describe('comments display', () => {
    beforeEach(() => {
      mockedUseComments.mockReturnValue({
        ...defaultMockReturn,
        comments: mockComments,
      })
    })

    it('displays comment content', () => {
      render(<CommentsPanel taskId="task-123" />)

      expect(screen.getByText('First comment from agent')).toBeInTheDocument()
      expect(screen.getByText('Reply from human with @agent mention')).toBeInTheDocument()
    })

    it('displays author badges', () => {
      render(<CommentsPanel taskId="task-123" />)

      // Agent comment should show author_id
      expect(screen.getByText('claude')).toBeInTheDocument()
      // Human comment shows author_type
      expect(screen.getByText('human')).toBeInTheDocument()
    })

    it('shows comment count in header', () => {
      render(<CommentsPanel taskId="task-123" />)

      expect(screen.getByText('(2)')).toBeInTheDocument()
    })

    it('displays mentions in comment', () => {
      render(<CommentsPanel taskId="task-123" />)

      // The @agent mention from comment-2 metadata
      expect(screen.getByText('@agent')).toBeInTheDocument()
    })
  })

  // 5.10.5: Test add comment form works
  describe('add comment form', () => {
    it('renders textarea and submit button', () => {
      mockedUseComments.mockReturnValue({
        ...defaultMockReturn,
      })

      render(<CommentsPanel taskId="task-123" />)

      expect(screen.getByPlaceholderText('Add a comment...')).toBeInTheDocument()
      expect(screen.getByText('Add Comment')).toBeInTheDocument()
    })

    it('disables submit button when textarea is empty', () => {
      mockedUseComments.mockReturnValue({
        ...defaultMockReturn,
      })

      render(<CommentsPanel taskId="task-123" />)

      const submitButton = screen.getByText('Add Comment')
      expect(submitButton).toBeDisabled()
    })

    it('enables submit button when textarea has content', async () => {
      mockedUseComments.mockReturnValue({
        ...defaultMockReturn,
      })

      render(<CommentsPanel taskId="task-123" />)

      const textarea = screen.getByPlaceholderText('Add a comment...')
      await userEvent.type(textarea, 'Test comment')

      const submitButton = screen.getByText('Add Comment')
      expect(submitButton).not.toBeDisabled()
    })

    it('calls addComment when form is submitted', async () => {
      const mockAddComment = vi.fn().mockResolvedValue({})

      mockedUseComments.mockReturnValue({
        ...defaultMockReturn,
        addComment: mockAddComment,
      })

      render(<CommentsPanel taskId="task-123" />)

      const textarea = screen.getByPlaceholderText('Add a comment...')
      await userEvent.type(textarea, 'New comment')

      const submitButton = screen.getByText('Add Comment')
      await userEvent.click(submitButton)

      await waitFor(() => {
        expect(mockAddComment).toHaveBeenCalledWith({
          task: 'task-123',
          content: 'New comment',
          author_type: 'human',
        })
      })
    })

    it('clears textarea after successful submission', async () => {
      const mockAddComment = vi.fn().mockResolvedValue({})

      mockedUseComments.mockReturnValue({
        ...defaultMockReturn,
        addComment: mockAddComment,
      })

      render(<CommentsPanel taskId="task-123" />)

      const textarea = screen.getByPlaceholderText('Add a comment...') as HTMLTextAreaElement
      await userEvent.type(textarea, 'New comment')
      expect(textarea.value).toBe('New comment')

      const submitButton = screen.getByText('Add Comment')
      await userEvent.click(submitButton)

      await waitFor(() => {
        expect(textarea.value).toBe('')
      })
    })

    it('shows "Adding..." when submission is in progress', () => {
      mockedUseComments.mockReturnValue({
        ...defaultMockReturn,
        adding: true,
      })

      render(<CommentsPanel taskId="task-123" />)

      expect(screen.getByText('Adding...')).toBeInTheDocument()
    })
  })

  // 5.10.6: Test @agent warning shows in textarea
  describe('@agent mention warning', () => {
    beforeEach(() => {
      mockedUseComments.mockReturnValue({
        ...defaultMockReturn,
      })
    })

    it('shows warning when @agent is typed', async () => {
      render(<CommentsPanel taskId="task-123" />)

      const textarea = screen.getByPlaceholderText('Add a comment...')
      await userEvent.type(textarea, 'Hello @agent please help')

      expect(screen.getByText(/Will trigger auto-resume/)).toBeInTheDocument()
    })

    it('does not show warning without @agent', async () => {
      render(<CommentsPanel taskId="task-123" />)

      const textarea = screen.getByPlaceholderText('Add a comment...')
      await userEvent.type(textarea, 'Hello world')

      expect(screen.queryByText(/Will trigger auto-resume/)).not.toBeInTheDocument()
    })

    it('shows warning for @agent in any position', async () => {
      render(<CommentsPanel taskId="task-123" />)

      const textarea = screen.getByPlaceholderText('Add a comment...')
      await userEvent.type(textarea, 'Please @agent continue')

      expect(screen.getByText(/Will trigger auto-resume/)).toBeInTheDocument()
    })
  })

  // Error state
  describe('error state', () => {
    it('displays error message when there is an error', () => {
      mockedUseComments.mockReturnValue({
        ...defaultMockReturn,
        error: new Error('Failed to fetch'),
      })

      render(<CommentsPanel taskId="task-123" />)

      expect(screen.getByText('Failed to load comments')).toBeInTheDocument()
    })
  })

  // Connection error state (SSE)
  describe('connection error state', () => {
    it('displays connection warning when connectionError exists', () => {
      mockedUseComments.mockReturnValue({
        ...defaultMockReturn,
        connectionError: new Error('SSE connection failed'),
      })

      render(<CommentsPanel taskId="task-123" />)

      expect(screen.getByText(/Real-time updates unavailable/)).toBeInTheDocument()
    })

    it('displays reconnecting message when reconnecting', () => {
      mockedUseComments.mockReturnValue({
        ...defaultMockReturn,
        reconnecting: true,
      })

      render(<CommentsPanel taskId="task-123" />)

      expect(screen.getByText(/Reconnecting to real-time updates/)).toBeInTheDocument()
    })

    it('does not show connection warning when connected', () => {
      mockedUseComments.mockReturnValue({
        ...defaultMockReturn,
        connectionError: null,
        reconnecting: false,
      })

      render(<CommentsPanel taskId="task-123" />)

      expect(screen.queryByText(/Real-time updates/)).not.toBeInTheDocument()
      expect(screen.queryByText(/Reconnecting/)).not.toBeInTheDocument()
    })
  })
})
