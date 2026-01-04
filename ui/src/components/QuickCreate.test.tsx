import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QuickCreate } from './QuickCreate'

describe('QuickCreate', () => {
  const defaultProps = {
    isOpen: false,
    onClose: vi.fn(),
    onCreate: vi.fn(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  // Test 1.1: Renders when open, doesn't render when closed
  describe('visibility', () => {
    it('does not render when closed', () => {
      render(<QuickCreate {...defaultProps} isOpen={false} />)
      expect(screen.queryByText('Create Task')).not.toBeInTheDocument()
    })

    it('renders when open', () => {
      render(<QuickCreate {...defaultProps} isOpen={true} />)
      expect(screen.getByText('Create Task')).toBeInTheDocument()
    })
  })

  // Test 1.2: Form submission calls onCreate
  describe('form submission', () => {
    it('calls onCreate with title and column on submit', async () => {
      const user = userEvent.setup()
      const onCreate = vi.fn().mockResolvedValue(undefined)

      render(<QuickCreate isOpen={true} onClose={vi.fn()} onCreate={onCreate} />)

      const input = screen.getByPlaceholderText('Task title...')
      await user.type(input, 'New test task')

      const createButton = screen.getByRole('button', { name: /create/i })
      await user.click(createButton)

      expect(onCreate).toHaveBeenCalledWith('New test task', 'backlog')
    })

    it('trims whitespace from title before calling onCreate', async () => {
      const user = userEvent.setup()
      const onCreate = vi.fn().mockResolvedValue(undefined)

      render(<QuickCreate isOpen={true} onClose={vi.fn()} onCreate={onCreate} />)

      const input = screen.getByPlaceholderText('Task title...')
      await user.type(input, '  Trimmed title  ')

      const createButton = screen.getByRole('button', { name: /create/i })
      await user.click(createButton)

      expect(onCreate).toHaveBeenCalledWith('Trimmed title', 'backlog')
    })

    it('calls onClose after successful creation', async () => {
      const user = userEvent.setup()
      const onCreate = vi.fn().mockResolvedValue(undefined)
      const onClose = vi.fn()

      render(<QuickCreate isOpen={true} onClose={onClose} onCreate={onCreate} />)

      const input = screen.getByPlaceholderText('Task title...')
      await user.type(input, 'New task')

      const createButton = screen.getByRole('button', { name: /create/i })
      await user.click(createButton)

      expect(onClose).toHaveBeenCalled()
    })
  })

  // Test 1.3: Escape key closes modal
  describe('escape key', () => {
    it('calls onClose when Escape is pressed', async () => {
      const user = userEvent.setup()
      const onClose = vi.fn()

      render(<QuickCreate isOpen={true} onClose={onClose} onCreate={vi.fn()} />)

      await user.keyboard('{Escape}')

      expect(onClose).toHaveBeenCalled()
    })

    it('does not call onCreate when Escape is pressed', async () => {
      const user = userEvent.setup()
      const onCreate = vi.fn()

      render(<QuickCreate isOpen={true} onClose={vi.fn()} onCreate={onCreate} />)

      const input = screen.getByPlaceholderText('Task title...')
      await user.type(input, 'Some title')
      await user.keyboard('{Escape}')

      expect(onCreate).not.toHaveBeenCalled()
    })
  })

  // Test 1.4: Column selector works
  describe('column selector', () => {
    it('shows all 5 columns in the dropdown', () => {
      render(<QuickCreate isOpen={true} onClose={vi.fn()} onCreate={vi.fn()} />)

      const columnSelect = screen.getByRole('combobox')
      const options = columnSelect.querySelectorAll('option')

      expect(options).toHaveLength(5)
      expect(screen.getByText('Backlog')).toBeInTheDocument()
      expect(screen.getByText('Todo')).toBeInTheDocument()
      expect(screen.getByText('In Progress')).toBeInTheDocument()
      expect(screen.getByText('Review')).toBeInTheDocument()
      expect(screen.getByText('Done')).toBeInTheDocument()
    })

    it('defaults to backlog column', () => {
      render(<QuickCreate isOpen={true} onClose={vi.fn()} onCreate={vi.fn()} />)

      const columnSelect = screen.getByRole('combobox')
      expect(columnSelect).toHaveValue('backlog')
    })

    it('allows selecting a different column', async () => {
      const user = userEvent.setup()
      const onCreate = vi.fn().mockResolvedValue(undefined)

      render(<QuickCreate isOpen={true} onClose={vi.fn()} onCreate={onCreate} />)

      const input = screen.getByPlaceholderText('Task title...')
      await user.type(input, 'Task in todo')

      const columnSelect = screen.getByRole('combobox')
      await user.selectOptions(columnSelect, 'todo')

      const createButton = screen.getByRole('button', { name: /create/i })
      await user.click(createButton)

      expect(onCreate).toHaveBeenCalledWith('Task in todo', 'todo')
    })

    it('allows selecting in_progress column', async () => {
      const user = userEvent.setup()
      const onCreate = vi.fn().mockResolvedValue(undefined)

      render(<QuickCreate isOpen={true} onClose={vi.fn()} onCreate={onCreate} />)

      const input = screen.getByPlaceholderText('Task title...')
      await user.type(input, 'In progress task')

      const columnSelect = screen.getByRole('combobox')
      await user.selectOptions(columnSelect, 'in_progress')

      const createButton = screen.getByRole('button', { name: /create/i })
      await user.click(createButton)

      expect(onCreate).toHaveBeenCalledWith('In progress task', 'in_progress')
    })
  })

  // Test 1.5: Validation - empty title should not submit
  describe('validation', () => {
    it('disables create button when title is empty', () => {
      render(<QuickCreate isOpen={true} onClose={vi.fn()} onCreate={vi.fn()} />)

      const createButton = screen.getByRole('button', { name: /create/i })
      expect(createButton).toBeDisabled()
    })

    it('disables create button when title is whitespace only', async () => {
      const user = userEvent.setup()
      const onCreate = vi.fn()

      render(<QuickCreate isOpen={true} onClose={vi.fn()} onCreate={onCreate} />)

      const input = screen.getByPlaceholderText('Task title...')
      await user.type(input, '   ')

      const createButton = screen.getByRole('button', { name: /create/i })
      expect(createButton).toBeDisabled()
      expect(onCreate).not.toHaveBeenCalled()
    })

    it('enables create button when title has content', async () => {
      const user = userEvent.setup()

      render(<QuickCreate isOpen={true} onClose={vi.fn()} onCreate={vi.fn()} />)

      const input = screen.getByPlaceholderText('Task title...')
      await user.type(input, 'Valid title')

      const createButton = screen.getByRole('button', { name: /create/i })
      expect(createButton).toBeEnabled()
    })
  })

  // Additional: overlay click closes modal
  describe('overlay click', () => {
    it('calls onClose when clicking overlay', async () => {
      const user = userEvent.setup()
      const onClose = vi.fn()

      const { container } = render(
        <QuickCreate isOpen={true} onClose={onClose} onCreate={vi.fn()} />
      )

      const overlay = container.querySelector('[class*="overlay"]') as HTMLElement
      await user.click(overlay)

      expect(onClose).toHaveBeenCalled()
    })

    it('does not close when clicking modal content', async () => {
      const user = userEvent.setup()
      const onClose = vi.fn()

      render(<QuickCreate isOpen={true} onClose={onClose} onCreate={vi.fn()} />)

      // Click on the modal title (inside the modal)
      await user.click(screen.getByText('Create Task'))

      expect(onClose).not.toHaveBeenCalled()
    })
  })
})
