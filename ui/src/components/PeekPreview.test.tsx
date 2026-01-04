import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { PeekPreview } from './PeekPreview'
import type { Task } from '../types/task'

// Mock task for testing
const mockTask: Task = {
  id: 'test-1234-5678-90ab-cdef',
  title: 'Test Task Title',
  description: 'This is a test description for the task.',
  type: 'feature',
  priority: 'high',
  column: 'in_progress',
  position: 1000,
  labels: ['frontend', 'ui', 'urgent'],
  created_by: 'user',
  collectionId: 'tasks',
  collectionName: 'tasks',
  created: '2024-01-15T10:00:00Z',
  updated: '2024-01-15T10:00:00Z',
}

describe('PeekPreview', () => {
  const mockOnClose = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders nothing when isOpen is false', () => {
    render(
      <PeekPreview task={mockTask} isOpen={false} onClose={mockOnClose} />
    )

    expect(screen.queryByText('Test Task Title')).not.toBeInTheDocument()
  })

  it('renders nothing when task is null', () => {
    render(<PeekPreview task={null} isOpen={true} onClose={mockOnClose} />)

    expect(screen.queryByText('Test Task Title')).not.toBeInTheDocument()
  })

  it('renders nothing when both isOpen is false and task is null', () => {
    render(<PeekPreview task={null} isOpen={false} onClose={mockOnClose} />)

    expect(screen.queryByRole('heading')).not.toBeInTheDocument()
  })

  it('displays task ID (truncated to 8 chars)', () => {
    render(
      <PeekPreview task={mockTask} isOpen={true} onClose={mockOnClose} />
    )

    // ID should be truncated to first 8 characters
    expect(screen.getByText('test-123')).toBeInTheDocument()
  })

  it('displays task type with correct styling', () => {
    render(
      <PeekPreview task={mockTask} isOpen={true} onClose={mockOnClose} />
    )

    expect(screen.getByText('feature')).toBeInTheDocument()
  })

  it('displays task title', () => {
    render(
      <PeekPreview task={mockTask} isOpen={true} onClose={mockOnClose} />
    )

    expect(screen.getByText('Test Task Title')).toBeInTheDocument()
  })

  it('displays status using COLUMN_NAMES', () => {
    render(
      <PeekPreview task={mockTask} isOpen={true} onClose={mockOnClose} />
    )

    // 'in_progress' should be displayed as 'In Progress'
    expect(screen.getByText('In Progress')).toBeInTheDocument()
  })

  it('displays priority', () => {
    render(
      <PeekPreview task={mockTask} isOpen={true} onClose={mockOnClose} />
    )

    expect(screen.getByText('high')).toBeInTheDocument()
  })

  it('displays labels if present', () => {
    render(
      <PeekPreview task={mockTask} isOpen={true} onClose={mockOnClose} />
    )

    expect(screen.getByText('frontend')).toBeInTheDocument()
    expect(screen.getByText('ui')).toBeInTheDocument()
    expect(screen.getByText('urgent')).toBeInTheDocument()
  })

  it('does not display labels section when no labels', () => {
    const taskWithoutLabels: Task = {
      ...mockTask,
      labels: undefined,
    }

    render(
      <PeekPreview task={taskWithoutLabels} isOpen={true} onClose={mockOnClose} />
    )

    expect(screen.queryByText('Labels')).not.toBeInTheDocument()
  })

  it('does not display labels section when labels array is empty', () => {
    const taskWithEmptyLabels: Task = {
      ...mockTask,
      labels: [],
    }

    render(
      <PeekPreview task={taskWithEmptyLabels} isOpen={true} onClose={mockOnClose} />
    )

    expect(screen.queryByText('Labels')).not.toBeInTheDocument()
  })

  it('displays description', () => {
    render(
      <PeekPreview task={mockTask} isOpen={true} onClose={mockOnClose} />
    )

    expect(
      screen.getByText('This is a test description for the task.')
    ).toBeInTheDocument()
  })

  it('truncates description to 200 characters with ellipsis', () => {
    const longDescription = 'A'.repeat(250)
    const taskWithLongDesc: Task = {
      ...mockTask,
      description: longDescription,
    }

    render(
      <PeekPreview task={taskWithLongDesc} isOpen={true} onClose={mockOnClose} />
    )

    // Should be truncated to 200 chars + '...'
    const expectedText = 'A'.repeat(200) + '...'
    expect(screen.getByText(expectedText)).toBeInTheDocument()
  })

  it('displays full description when under 200 characters', () => {
    const shortDescription = 'Short description'
    const taskWithShortDesc: Task = {
      ...mockTask,
      description: shortDescription,
    }

    render(
      <PeekPreview task={taskWithShortDesc} isOpen={true} onClose={mockOnClose} />
    )

    expect(screen.getByText('Short description')).toBeInTheDocument()
    // Should not have ellipsis
    expect(screen.queryByText(/\.\.\.$/)).not.toBeInTheDocument()
  })

  it('does not display description section when no description', () => {
    const taskWithoutDesc: Task = {
      ...mockTask,
      description: undefined,
    }

    render(
      <PeekPreview task={taskWithoutDesc} isOpen={true} onClose={mockOnClose} />
    )

    expect(screen.queryByText('Description')).not.toBeInTheDocument()
  })

  it('closes on overlay click', async () => {
    const user = userEvent.setup()

    render(
      <PeekPreview task={mockTask} isOpen={true} onClose={mockOnClose} />
    )

    // Click the overlay (the outer div)
    const overlay = document.querySelector('[class*="overlay"]')
    expect(overlay).toBeInTheDocument()
    if (overlay) {
      await user.click(overlay)
    }

    expect(mockOnClose).toHaveBeenCalled()
  })

  it('does not close when clicking inside the preview', async () => {
    const user = userEvent.setup()

    render(
      <PeekPreview task={mockTask} isOpen={true} onClose={mockOnClose} />
    )

    // Click inside the preview content
    const title = screen.getByText('Test Task Title')
    await user.click(title)

    expect(mockOnClose).not.toHaveBeenCalled()
  })

  it('shows footer hint about Enter key', () => {
    render(
      <PeekPreview task={mockTask} isOpen={true} onClose={mockOnClose} />
    )

    expect(screen.getByText(/press/i)).toBeInTheDocument()
    // Find the kbd element with 'Enter'
    const footer = document.querySelector('[class*="footer"]')
    expect(footer).toBeInTheDocument()
    expect(footer?.textContent).toContain('Enter')
  })

  describe('different task types', () => {
    it('renders bug type correctly', () => {
      const bugTask: Task = {
        ...mockTask,
        type: 'bug',
      }

      render(<PeekPreview task={bugTask} isOpen={true} onClose={mockOnClose} />)

      expect(screen.getByText('bug')).toBeInTheDocument()
    })

    it('renders chore type correctly', () => {
      const choreTask: Task = {
        ...mockTask,
        type: 'chore',
      }

      render(
        <PeekPreview task={choreTask} isOpen={true} onClose={mockOnClose} />
      )

      expect(screen.getByText('chore')).toBeInTheDocument()
    })
  })

  describe('different statuses', () => {
    it('renders backlog status correctly', () => {
      const backlogTask: Task = {
        ...mockTask,
        column: 'backlog',
      }

      render(
        <PeekPreview task={backlogTask} isOpen={true} onClose={mockOnClose} />
      )

      expect(screen.getByText('Backlog')).toBeInTheDocument()
    })

    it('renders done status correctly', () => {
      const doneTask: Task = {
        ...mockTask,
        column: 'done',
      }

      render(
        <PeekPreview task={doneTask} isOpen={true} onClose={mockOnClose} />
      )

      expect(screen.getByText('Done')).toBeInTheDocument()
    })
  })

  describe('different priorities', () => {
    it('renders urgent priority correctly', () => {
      const urgentTask: Task = {
        ...mockTask,
        priority: 'urgent',
        labels: ['frontend'], // Use labels without 'urgent' to avoid duplicate text
      }

      render(
        <PeekPreview task={urgentTask} isOpen={true} onClose={mockOnClose} />
      )

      expect(screen.getByText('urgent')).toBeInTheDocument()
    })

    it('renders low priority correctly', () => {
      const lowTask: Task = {
        ...mockTask,
        priority: 'low',
      }

      render(
        <PeekPreview task={lowTask} isOpen={true} onClose={mockOnClose} />
      )

      expect(screen.getByText('low')).toBeInTheDocument()
    })
  })
})
