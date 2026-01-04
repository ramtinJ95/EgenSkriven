import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { TaskDetail } from './TaskDetail'
import type { Task } from '../types/task'

describe('TaskDetail', () => {
  const mockTask: Task = {
    id: 'task-123',
    title: 'Test Task',
    description: 'Test description',
    type: 'feature',
    priority: 'high',
    column: 'todo',
    position: 1000,
    labels: ['frontend'],
    created_by: 'user',
    collectionId: 'tasks',
    collectionName: 'tasks',
    created: '2024-01-15T10:00:00Z',
    updated: '2024-01-15T10:00:00Z',
  }

  const defaultProps = {
    task: null as Task | null,
    onClose: vi.fn(),
    onUpdate: vi.fn(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  // Test 2.1: Renders task title and metadata
  describe('rendering', () => {
    it('renders nothing when task is null', () => {
      render(<TaskDetail {...defaultProps} task={null} />)
      expect(screen.queryByRole('heading')).not.toBeInTheDocument()
    })

    it('renders task title', () => {
      render(<TaskDetail {...defaultProps} task={mockTask} />)
      expect(screen.getByRole('heading', { name: 'Test Task' })).toBeInTheDocument()
    })

    it('renders task ID', () => {
      render(<TaskDetail {...defaultProps} task={mockTask} />)
      expect(screen.getByText('task-123')).toBeInTheDocument()
    })

    it('renders task description', () => {
      render(<TaskDetail {...defaultProps} task={mockTask} />)
      expect(screen.getByText('Test description')).toBeInTheDocument()
    })

    it('does not render description section when description is empty', () => {
      const taskWithoutDescription = { ...mockTask, description: undefined }
      render(<TaskDetail {...defaultProps} task={taskWithoutDescription} />)
      expect(screen.queryByText('Description')).not.toBeInTheDocument()
    })

    it('renders labels', () => {
      render(<TaskDetail {...defaultProps} task={mockTask} />)
      expect(screen.getByText('frontend')).toBeInTheDocument()
    })

    it('renders multiple labels', () => {
      const taskWithMultipleLabels = { ...mockTask, labels: ['frontend', 'urgent', 'api'] }
      render(<TaskDetail {...defaultProps} task={taskWithMultipleLabels} />)
      expect(screen.getByText('frontend')).toBeInTheDocument()
      expect(screen.getByText('urgent')).toBeInTheDocument()
      expect(screen.getByText('api')).toBeInTheDocument()
    })

    it('does not render labels section when labels are empty', () => {
      const taskWithoutLabels = { ...mockTask, labels: [] }
      render(<TaskDetail {...defaultProps} task={taskWithoutLabels} />)
      expect(screen.queryByText('Labels')).not.toBeInTheDocument()
    })

    it('renders close button', () => {
      render(<TaskDetail {...defaultProps} task={mockTask} />)
      expect(screen.getByRole('button', { name: 'Close' })).toBeInTheDocument()
    })
  })

  // Test 2.2: Status/priority/type dropdowns call onUpdate
  describe('dropdown interactions', () => {
    it('calls onUpdate when status is changed', async () => {
      const user = userEvent.setup()
      const onUpdate = vi.fn().mockResolvedValue(undefined)

      render(<TaskDetail task={mockTask} onClose={vi.fn()} onUpdate={onUpdate} />)

      // Find the status select (first combobox - Status)
      const selects = screen.getAllByRole('combobox')
      const statusSelect = selects[0]
      await user.selectOptions(statusSelect, 'in_progress')

      expect(onUpdate).toHaveBeenCalledWith('task-123', { column: 'in_progress' })
    })

    it('calls onUpdate when type is changed', async () => {
      const user = userEvent.setup()
      const onUpdate = vi.fn().mockResolvedValue(undefined)

      render(<TaskDetail task={mockTask} onClose={vi.fn()} onUpdate={onUpdate} />)

      // Find the type select (second combobox - Type)
      const selects = screen.getAllByRole('combobox')
      const typeSelect = selects[1]
      await user.selectOptions(typeSelect, 'bug')

      expect(onUpdate).toHaveBeenCalledWith('task-123', { type: 'bug' })
    })

    it('calls onUpdate when priority is changed', async () => {
      const user = userEvent.setup()
      const onUpdate = vi.fn().mockResolvedValue(undefined)

      render(<TaskDetail task={mockTask} onClose={vi.fn()} onUpdate={onUpdate} />)

      // Find the priority select (third combobox - Priority)
      const selects = screen.getAllByRole('combobox')
      const prioritySelect = selects[2]
      await user.selectOptions(prioritySelect, 'urgent')

      expect(onUpdate).toHaveBeenCalledWith('task-123', { priority: 'urgent' })
    })

    it('shows all status options', () => {
      render(<TaskDetail {...defaultProps} task={mockTask} />)

      const selects = screen.getAllByRole('combobox')
      const statusSelect = selects[0]
      const options = statusSelect.querySelectorAll('option')

      expect(options).toHaveLength(5)
    })

    it('shows all type options', () => {
      render(<TaskDetail {...defaultProps} task={mockTask} />)

      const selects = screen.getAllByRole('combobox')
      const typeSelect = selects[1]
      const options = typeSelect.querySelectorAll('option')

      expect(options).toHaveLength(3) // bug, feature, chore
    })

    it('shows all priority options', () => {
      render(<TaskDetail {...defaultProps} task={mockTask} />)

      const selects = screen.getAllByRole('combobox')
      const prioritySelect = selects[2]
      const options = prioritySelect.querySelectorAll('option')

      expect(options).toHaveLength(4) // urgent, high, medium, low
    })

    it('displays current status value', () => {
      render(<TaskDetail {...defaultProps} task={mockTask} />)

      const selects = screen.getAllByRole('combobox')
      const statusSelect = selects[0]
      expect(statusSelect).toHaveValue('todo')
    })

    it('displays current type value', () => {
      render(<TaskDetail {...defaultProps} task={mockTask} />)

      const selects = screen.getAllByRole('combobox')
      const typeSelect = selects[1]
      expect(typeSelect).toHaveValue('feature')
    })

    it('displays current priority value', () => {
      render(<TaskDetail {...defaultProps} task={mockTask} />)

      const selects = screen.getAllByRole('combobox')
      const prioritySelect = selects[2]
      expect(prioritySelect).toHaveValue('high')
    })
  })

  // Test 2.3: Escape key closes panel
  describe('escape key', () => {
    it('calls onClose when Escape is pressed', async () => {
      const user = userEvent.setup()
      const onClose = vi.fn()

      render(<TaskDetail task={mockTask} onClose={onClose} onUpdate={vi.fn()} />)

      await user.keyboard('{Escape}')

      expect(onClose).toHaveBeenCalled()
    })

    it('does not call onClose via Escape when task is null', async () => {
      const user = userEvent.setup()
      const onClose = vi.fn()

      render(<TaskDetail task={null} onClose={onClose} onUpdate={vi.fn()} />)

      await user.keyboard('{Escape}')

      expect(onClose).not.toHaveBeenCalled()
    })
  })

  // Test 2.4: Click outside closes panel
  describe('click interactions', () => {
    it('calls onClose when clicking overlay', async () => {
      const user = userEvent.setup()
      const onClose = vi.fn()

      const { container } = render(
        <TaskDetail task={mockTask} onClose={onClose} onUpdate={vi.fn()} />
      )

      const overlay = container.querySelector('[class*="overlay"]') as HTMLElement
      await user.click(overlay)

      expect(onClose).toHaveBeenCalled()
    })

    it('does not close when clicking panel content', async () => {
      const user = userEvent.setup()
      const onClose = vi.fn()

      render(<TaskDetail task={mockTask} onClose={onClose} onUpdate={vi.fn()} />)

      // Click on the task title (inside the panel)
      await user.click(screen.getByRole('heading', { name: 'Test Task' }))

      expect(onClose).not.toHaveBeenCalled()
    })

    it('calls onClose when clicking close button', async () => {
      const user = userEvent.setup()
      const onClose = vi.fn()

      render(<TaskDetail task={mockTask} onClose={onClose} onUpdate={vi.fn()} />)

      await user.click(screen.getByRole('button', { name: 'Close' }))

      expect(onClose).toHaveBeenCalled()
    })
  })

  // Additional: metadata rendering
  describe('metadata', () => {
    it('renders created date', () => {
      render(<TaskDetail {...defaultProps} task={mockTask} />)
      // The date format depends on locale, check for some part of it
      expect(screen.getByText('Created')).toBeInTheDocument()
    })

    it('renders updated date', () => {
      render(<TaskDetail {...defaultProps} task={mockTask} />)
      expect(screen.getByText('Updated')).toBeInTheDocument()
    })

    it('renders created by', () => {
      render(<TaskDetail {...defaultProps} task={mockTask} />)
      expect(screen.getByText('Created By')).toBeInTheDocument()
      expect(screen.getByText('user')).toBeInTheDocument()
    })

    it('renders due date when present', () => {
      const taskWithDueDate = { ...mockTask, due_date: '2024-02-15T10:00:00Z' }
      render(<TaskDetail {...defaultProps} task={taskWithDueDate} />)
      expect(screen.getByText('Due Date')).toBeInTheDocument()
    })

    it('does not render due date section when not present', () => {
      render(<TaskDetail {...defaultProps} task={mockTask} />)
      expect(screen.queryByText('Due Date')).not.toBeInTheDocument()
    })

    it('renders blocked by when present', () => {
      const taskWithBlockedBy = { ...mockTask, blocked_by: ['task-456', 'task-789'] }
      render(<TaskDetail {...defaultProps} task={taskWithBlockedBy} />)
      expect(screen.getByText('Blocked By')).toBeInTheDocument()
      // IDs are truncated to 8 chars
      expect(screen.getByText('task-456')).toBeInTheDocument()
      expect(screen.getByText('task-789')).toBeInTheDocument()
    })

    it('does not render blocked by section when not present', () => {
      render(<TaskDetail {...defaultProps} task={mockTask} />)
      expect(screen.queryByText('Blocked By')).not.toBeInTheDocument()
    })
  })
})
