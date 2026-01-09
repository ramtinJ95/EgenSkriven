import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ResumeModal } from './ResumeModal'
import type { ResumeResult } from '../hooks/useResume'

// Mock the useResume hook
vi.mock('../hooks/useResume', () => ({
  useResume: vi.fn(),
}))

import { useResume } from '../hooks/useResume'

const mockedUseResume = vi.mocked(useResume)

// Mock result data
const mockResumeResult: ResumeResult = {
  command: "claude --resume abc123 'Continue the task'",
  prompt: '## Task Context\n\nContinue working...',
  tool: 'claude-code',
  sessionRef: 'abc123-def456',
  workingDir: '/home/user/project',
}

describe('ResumeModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // Mock clipboard API
    Object.assign(navigator, {
      clipboard: {
        writeText: vi.fn().mockResolvedValue(undefined),
      },
    })
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  // Test: Modal is not rendered when not open
  describe('visibility', () => {
    it('returns null when isOpen is false', () => {
      mockedUseResume.mockReturnValue({
        resume: vi.fn(),
        loading: false,
        error: null,
      })

      const { container } = render(
        <ResumeModal
          isOpen={false}
          onClose={vi.fn()}
          taskId="task-123"
          displayId="WRK-123"
        />
      )

      expect(container.firstChild).toBeNull()
    })

    it('renders modal when isOpen is true', () => {
      mockedUseResume.mockReturnValue({
        resume: vi.fn(),
        loading: false,
        error: null,
      })

      render(
        <ResumeModal
          isOpen={true}
          onClose={vi.fn()}
          taskId="task-123"
          displayId="WRK-123"
        />
      )

      expect(screen.getByText('Resume Session for WRK-123')).toBeInTheDocument()
    })
  })

  // Test: Initial state shows generate button
  describe('initial state', () => {
    it('displays description text', () => {
      mockedUseResume.mockReturnValue({
        resume: vi.fn(),
        loading: false,
        error: null,
      })

      render(
        <ResumeModal
          isOpen={true}
          onClose={vi.fn()}
          taskId="task-123"
          displayId="WRK-123"
        />
      )

      expect(
        screen.getByText(/Generate the command to resume the agent session/)
      ).toBeInTheDocument()
    })

    it('shows generate button', () => {
      mockedUseResume.mockReturnValue({
        resume: vi.fn(),
        loading: false,
        error: null,
      })

      render(
        <ResumeModal
          isOpen={true}
          onClose={vi.fn()}
          taskId="task-123"
          displayId="WRK-123"
        />
      )

      expect(screen.getByText('Generate Resume Command')).toBeInTheDocument()
    })
  })

  // Test: Generate command functionality
  describe('generate command', () => {
    it('calls resume when generate button is clicked', async () => {
      const mockResume = vi.fn().mockResolvedValue(mockResumeResult)
      mockedUseResume.mockReturnValue({
        resume: mockResume,
        loading: false,
        error: null,
      })

      render(
        <ResumeModal
          isOpen={true}
          onClose={vi.fn()}
          taskId="task-123"
          displayId="WRK-123"
        />
      )

      const generateButton = screen.getByText('Generate Resume Command')
      await userEvent.click(generateButton)

      expect(mockResume).toHaveBeenCalledWith({
        taskId: 'task-123',
        displayId: 'WRK-123',
        exec: false,
      })
    })

    it('shows loading state while generating', () => {
      mockedUseResume.mockReturnValue({
        resume: vi.fn(),
        loading: true,
        error: null,
      })

      render(
        <ResumeModal
          isOpen={true}
          onClose={vi.fn()}
          taskId="task-123"
          displayId="WRK-123"
        />
      )

      expect(screen.getByText('Generating...')).toBeInTheDocument()
    })

    it('disables generate button while loading', () => {
      mockedUseResume.mockReturnValue({
        resume: vi.fn(),
        loading: true,
        error: null,
      })

      render(
        <ResumeModal
          isOpen={true}
          onClose={vi.fn()}
          taskId="task-123"
          displayId="WRK-123"
        />
      )

      const button = screen.getByText('Generating...').closest('button')
      expect(button).toBeDisabled()
    })

    it('shows error message when generation fails', async () => {
      const mockResume = vi.fn().mockRejectedValue(new Error('Network error'))
      mockedUseResume.mockReturnValue({
        resume: mockResume,
        loading: false,
        error: null,
      })

      render(
        <ResumeModal
          isOpen={true}
          onClose={vi.fn()}
          taskId="task-123"
          displayId="WRK-123"
        />
      )

      const generateButton = screen.getByText('Generate Resume Command')
      await userEvent.click(generateButton)

      await waitFor(() => {
        expect(screen.getByText('Network error')).toBeInTheDocument()
      })
    })
  })

  // Test: Result display
  describe('result display', () => {
    it('displays tool name after generation', async () => {
      const mockResume = vi.fn().mockResolvedValue(mockResumeResult)
      mockedUseResume.mockReturnValue({
        resume: mockResume,
        loading: false,
        error: null,
      })

      render(
        <ResumeModal
          isOpen={true}
          onClose={vi.fn()}
          taskId="task-123"
          displayId="WRK-123"
        />
      )

      await userEvent.click(screen.getByText('Generate Resume Command'))

      await waitFor(() => {
        expect(screen.getByText('Tool:')).toBeInTheDocument()
        expect(screen.getByText('claude-code')).toBeInTheDocument()
      })
    })

    it('displays working directory after generation', async () => {
      const mockResume = vi.fn().mockResolvedValue(mockResumeResult)
      mockedUseResume.mockReturnValue({
        resume: mockResume,
        loading: false,
        error: null,
      })

      render(
        <ResumeModal
          isOpen={true}
          onClose={vi.fn()}
          taskId="task-123"
          displayId="WRK-123"
        />
      )

      await userEvent.click(screen.getByText('Generate Resume Command'))

      await waitFor(() => {
        expect(screen.getByText('Working Dir:')).toBeInTheDocument()
        // Working dir appears twice (info row and instructions), use getAllByText
        const workingDirElements = screen.getAllByText('/home/user/project')
        expect(workingDirElements.length).toBeGreaterThanOrEqual(1)
      })
    })

    it('displays command after generation', async () => {
      const mockResume = vi.fn().mockResolvedValue(mockResumeResult)
      mockedUseResume.mockReturnValue({
        resume: mockResume,
        loading: false,
        error: null,
      })

      render(
        <ResumeModal
          isOpen={true}
          onClose={vi.fn()}
          taskId="task-123"
          displayId="WRK-123"
        />
      )

      await userEvent.click(screen.getByText('Generate Resume Command'))

      await waitFor(() => {
        expect(screen.getByText('Resume Command')).toBeInTheDocument()
        expect(screen.getByText(mockResumeResult.command)).toBeInTheDocument()
      })
    })

    it('displays instructions after generation', async () => {
      const mockResume = vi.fn().mockResolvedValue(mockResumeResult)
      mockedUseResume.mockReturnValue({
        resume: mockResume,
        loading: false,
        error: null,
      })

      render(
        <ResumeModal
          isOpen={true}
          onClose={vi.fn()}
          taskId="task-123"
          displayId="WRK-123"
        />
      )

      await userEvent.click(screen.getByText('Generate Resume Command'))

      await waitFor(() => {
        expect(screen.getByText('Instructions')).toBeInTheDocument()
        expect(screen.getByText('Copy the command above')).toBeInTheDocument()
        expect(screen.getByText('Paste and run the command')).toBeInTheDocument()
      })
    })
  })

  // Test: Copy to clipboard functionality
  describe('copy to clipboard', () => {
    it('shows copy button after generation', async () => {
      const mockResume = vi.fn().mockResolvedValue(mockResumeResult)
      mockedUseResume.mockReturnValue({
        resume: mockResume,
        loading: false,
        error: null,
      })

      render(
        <ResumeModal
          isOpen={true}
          onClose={vi.fn()}
          taskId="task-123"
          displayId="WRK-123"
        />
      )

      await userEvent.click(screen.getByText('Generate Resume Command'))

      await waitFor(() => {
        expect(screen.getByText('Copy')).toBeInTheDocument()
      })
    })

    it('copies command to clipboard when copy button is clicked', async () => {
      const mockResume = vi.fn().mockResolvedValue(mockResumeResult)
      mockedUseResume.mockReturnValue({
        resume: mockResume,
        loading: false,
        error: null,
      })

      render(
        <ResumeModal
          isOpen={true}
          onClose={vi.fn()}
          taskId="task-123"
          displayId="WRK-123"
        />
      )

      await userEvent.click(screen.getByText('Generate Resume Command'))

      await waitFor(() => {
        expect(screen.getByText('Copy')).toBeInTheDocument()
      })

      await userEvent.click(screen.getByText('Copy'))

      expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
        mockResumeResult.command
      )
    })

    it('shows "Copied!" after successful copy', async () => {
      const mockResume = vi.fn().mockResolvedValue(mockResumeResult)
      mockedUseResume.mockReturnValue({
        resume: mockResume,
        loading: false,
        error: null,
      })

      render(
        <ResumeModal
          isOpen={true}
          onClose={vi.fn()}
          taskId="task-123"
          displayId="WRK-123"
        />
      )

      await userEvent.click(screen.getByText('Generate Resume Command'))

      await waitFor(() => {
        expect(screen.getByText('Copy')).toBeInTheDocument()
      })

      await userEvent.click(screen.getByText('Copy'))

      await waitFor(() => {
        expect(screen.getByText('Copied!')).toBeInTheDocument()
      })
    })
  })

  // Test: Modal close functionality
  describe('close functionality', () => {
    it('calls onClose when close button is clicked', async () => {
      const onClose = vi.fn()
      mockedUseResume.mockReturnValue({
        resume: vi.fn(),
        loading: false,
        error: null,
      })

      render(
        <ResumeModal
          isOpen={true}
          onClose={onClose}
          taskId="task-123"
          displayId="WRK-123"
        />
      )

      // Find close button (it's the button in the header with X icon)
      const closeButton = document.querySelector('[class*="closeButton"]')
      expect(closeButton).toBeInTheDocument()

      fireEvent.click(closeButton!)

      expect(onClose).toHaveBeenCalledTimes(1)
    })

    it('calls onClose when clicking overlay', () => {
      const onClose = vi.fn()
      mockedUseResume.mockReturnValue({
        resume: vi.fn(),
        loading: false,
        error: null,
      })

      render(
        <ResumeModal
          isOpen={true}
          onClose={onClose}
          taskId="task-123"
          displayId="WRK-123"
        />
      )

      // Click on the overlay (the outer div)
      const overlay = document.querySelector('[class*="overlay"]')
      expect(overlay).toBeInTheDocument()

      fireEvent.click(overlay!)

      expect(onClose).toHaveBeenCalledTimes(1)
    })

    it('calls onClose when Escape key is pressed', () => {
      const onClose = vi.fn()
      mockedUseResume.mockReturnValue({
        resume: vi.fn(),
        loading: false,
        error: null,
      })

      render(
        <ResumeModal
          isOpen={true}
          onClose={onClose}
          taskId="task-123"
          displayId="WRK-123"
        />
      )

      fireEvent.keyDown(window, { key: 'Escape' })

      expect(onClose).toHaveBeenCalledTimes(1)
    })

    it('does not call onClose when clicking modal content', () => {
      const onClose = vi.fn()
      mockedUseResume.mockReturnValue({
        resume: vi.fn(),
        loading: false,
        error: null,
      })

      render(
        <ResumeModal
          isOpen={true}
          onClose={onClose}
          taskId="task-123"
          displayId="WRK-123"
        />
      )

      // Click on the modal content (not the overlay)
      const modal = document.querySelector('[class*="modal"]')
      expect(modal).toBeInTheDocument()

      fireEvent.click(modal!)

      // onClose should not be called when clicking inside the modal
      expect(onClose).not.toHaveBeenCalled()
    })
  })

  // Test: State reset when modal reopens
  describe('state reset', () => {
    it('resets result when modal reopens', async () => {
      const mockResume = vi.fn().mockResolvedValue(mockResumeResult)
      mockedUseResume.mockReturnValue({
        resume: mockResume,
        loading: false,
        error: null,
      })

      const { rerender } = render(
        <ResumeModal
          isOpen={true}
          onClose={vi.fn()}
          taskId="task-123"
          displayId="WRK-123"
        />
      )

      // Generate result
      await userEvent.click(screen.getByText('Generate Resume Command'))

      await waitFor(() => {
        expect(screen.getByText(mockResumeResult.command)).toBeInTheDocument()
      })

      // Close modal
      rerender(
        <ResumeModal
          isOpen={false}
          onClose={vi.fn()}
          taskId="task-123"
          displayId="WRK-123"
        />
      )

      // Reopen modal
      rerender(
        <ResumeModal
          isOpen={true}
          onClose={vi.fn()}
          taskId="task-123"
          displayId="WRK-123"
        />
      )

      // Should show generate button again, not the result
      expect(screen.getByText('Generate Resume Command')).toBeInTheDocument()
      expect(screen.queryByText(mockResumeResult.command)).not.toBeInTheDocument()
    })
  })

  // Test: Display ID in header
  describe('display id', () => {
    it('shows correct display ID in title', () => {
      mockedUseResume.mockReturnValue({
        resume: vi.fn(),
        loading: false,
        error: null,
      })

      render(
        <ResumeModal
          isOpen={true}
          onClose={vi.fn()}
          taskId="task-456"
          displayId="PROJ-789"
        />
      )

      expect(screen.getByText('Resume Session for PROJ-789')).toBeInTheDocument()
    })
  })
})
