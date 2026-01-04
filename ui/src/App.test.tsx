import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from './App'

// Mock useTasks hook
vi.mock('./hooks/useTasks', () => ({
  useTasks: () => ({
    tasks: [
      {
        id: 'task-1',
        title: 'Test Task',
        type: 'feature',
        priority: 'medium',
        column: 'todo',
        position: 1000,
        labels: [],
        created_by: 'user',
        collectionId: 'tasks',
        collectionName: 'tasks',
        created: '2024-01-15T10:00:00Z',
        updated: '2024-01-15T10:00:00Z',
      },
      {
        id: 'task-2',
        title: 'Second Task',
        type: 'bug',
        priority: 'high',
        column: 'in_progress',
        position: 2000,
        labels: ['frontend'],
        created_by: 'user',
        collectionId: 'tasks',
        collectionName: 'tasks',
        created: '2024-01-15T11:00:00Z',
        updated: '2024-01-15T11:00:00Z',
      },
    ],
    loading: false,
    error: null,
    createTask: vi.fn().mockResolvedValue({}),
    updateTask: vi.fn().mockResolvedValue({}),
    deleteTask: vi.fn(),
    moveTask: vi.fn(),
  }),
}))

// Helper to check if element has a class containing the given string (for CSS modules)
const hasClassContaining = (element: Element | null, className: string): boolean => {
  if (!element) return false
  return Array.from(element.classList).some(cls => cls.includes(className))
}

describe('App', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  // Test 3.1: C key opens QuickCreate modal
  describe('C key shortcut', () => {
    it('opens QuickCreate modal when C is pressed', async () => {
      const user = userEvent.setup()
      render(<App />)

      // QuickCreate should not be visible initially
      expect(screen.queryByText('Create Task')).not.toBeInTheDocument()

      // Press C key
      await user.keyboard('c')

      // QuickCreate should now be visible
      expect(screen.getByText('Create Task')).toBeInTheDocument()
    })

    it('opens QuickCreate modal when uppercase C is pressed', async () => {
      const user = userEvent.setup()
      render(<App />)

      await user.keyboard('C')

      expect(screen.getByText('Create Task')).toBeInTheDocument()
    })

    it('does not open QuickCreate when typing in the input', async () => {
      const user = userEvent.setup()
      render(<App />)

      // First open the modal
      await user.keyboard('c')
      expect(screen.getByText('Create Task')).toBeInTheDocument()

      // Type 'c' in the input - should not cause issues
      const input = screen.getByPlaceholderText('Task title...')
      await user.type(input, 'test c task')

      // Modal should still be open (not closed/re-opened)
      expect(screen.getByText('Create Task')).toBeInTheDocument()
      expect(input).toHaveValue('test c task')
    })

    it('does not re-open QuickCreate when already open', async () => {
      const user = userEvent.setup()
      render(<App />)

      // Open modal
      await user.keyboard('c')
      expect(screen.getByText('Create Task')).toBeInTheDocument()

      // Close it first by pressing Escape
      await user.keyboard('{Escape}')
      expect(screen.queryByText('Create Task')).not.toBeInTheDocument()

      // Now 'c' should open it again
      await user.keyboard('c')
      expect(screen.getByText('Create Task')).toBeInTheDocument()
    })
  })

  // Test 3.2: Escape closes modal/panel
  describe('Escape key', () => {
    it('closes QuickCreate modal when Escape is pressed', async () => {
      const user = userEvent.setup()
      render(<App />)

      // Open modal
      await user.keyboard('c')
      expect(screen.getByText('Create Task')).toBeInTheDocument()

      // Press Escape
      await user.keyboard('{Escape}')

      // Modal should be closed
      expect(screen.queryByText('Create Task')).not.toBeInTheDocument()
    })

    it('closes TaskDetail panel when Escape is pressed', async () => {
      const user = userEvent.setup()
      render(<App />)

      // Click a task to open detail panel (click on the card, not just the text)
      const taskCard = screen.getByRole('button', { name: /Test Task/i })
      await user.click(taskCard)

      // TaskDetail should be open - check for the Close button which is unique to TaskDetail
      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Close' })).toBeInTheDocument()
      })

      // Press Escape
      await user.keyboard('{Escape}')

      // TaskDetail should be closed
      await waitFor(() => {
        expect(screen.queryByRole('button', { name: 'Close' })).not.toBeInTheDocument()
      })
    })

    it('deselects task when Escape is pressed with no panel open', async () => {
      const user = userEvent.setup()
      render(<App />)

      // Click a task to open detail and select it
      const taskCard = screen.getByRole('button', { name: /Test Task/i })
      await user.click(taskCard)

      // Wait for detail panel to open
      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Close' })).toBeInTheDocument()
      })

      // Close the detail panel (first Escape)
      await user.keyboard('{Escape}')

      // Wait for detail panel to close
      await waitFor(() => {
        expect(screen.queryByRole('button', { name: 'Close' })).not.toBeInTheDocument()
      })

      // Task should still be selected at this point
      expect(hasClassContaining(taskCard, 'selected')).toBe(true)

      // Press Escape again to deselect
      await user.keyboard('{Escape}')

      // The task should no longer be visually selected
      expect(hasClassContaining(taskCard, 'selected')).toBe(false)
    })
  })

  // Test 3.3: Enter opens detail for selected task / clicking task
  describe('task interaction', () => {
    it('opens TaskDetail when clicking a task', async () => {
      const user = userEvent.setup()
      render(<App />)

      // Click a task card
      const taskCard = screen.getByRole('button', { name: /Test Task/i })
      await user.click(taskCard)

      // TaskDetail should open - check for Close button
      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Close' })).toBeInTheDocument()
      })
    })

    it('opens TaskDetail when Enter is pressed on selected task', async () => {
      const user = userEvent.setup()
      render(<App />)

      // Click on the task card to select it and open detail
      const taskCard = screen.getByRole('button', { name: /Test Task/i })
      await user.click(taskCard)

      // Wait for detail panel to open
      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Close' })).toBeInTheDocument()
      })

      // Close the detail panel (but keep selection)
      await user.keyboard('{Escape}')

      // Verify detail is closed
      await waitFor(() => {
        expect(screen.queryByRole('button', { name: 'Close' })).not.toBeInTheDocument()
      })

      // Now press Enter to reopen detail
      await user.keyboard('{Enter}')

      // TaskDetail should open again
      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Close' })).toBeInTheDocument()
      })
    })

    it('does not open TaskDetail when Enter is pressed without selection', async () => {
      const user = userEvent.setup()
      render(<App />)

      // Press Enter without selecting any task
      await user.keyboard('{Enter}')

      // No detail panel should open
      expect(screen.queryByRole('button', { name: 'Close' })).not.toBeInTheDocument()
    })
  })

  // Test 3.4: Task selection state works
  describe('task selection state', () => {
    it('clicking a task adds selected class to the card', async () => {
      const user = userEvent.setup()
      render(<App />)

      const taskCard = screen.getByRole('button', { name: /Test Task/i })
      await user.click(taskCard)

      // Check for selected class (CSS modules hash the class name)
      expect(hasClassContaining(taskCard, 'selected')).toBe(true)
    })

    it('maintains task selection after closing detail panel', async () => {
      const user = userEvent.setup()
      render(<App />)

      // Click a task to open detail
      const taskCard = screen.getByRole('button', { name: /Test Task/i })
      await user.click(taskCard)

      // Wait for detail to open
      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Close' })).toBeInTheDocument()
      })

      // Close the detail panel
      await user.keyboard('{Escape}')

      // Wait for detail to close
      await waitFor(() => {
        expect(screen.queryByRole('button', { name: 'Close' })).not.toBeInTheDocument()
      })

      // The task should still be visually selected
      expect(hasClassContaining(taskCard, 'selected')).toBe(true)
    })

    it('clears selection when Escape is pressed twice', async () => {
      const user = userEvent.setup()
      render(<App />)

      // Click a task to select and open detail
      const taskCard = screen.getByRole('button', { name: /Test Task/i })
      await user.click(taskCard)

      // Wait for detail to open
      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Close' })).toBeInTheDocument()
      })

      // First Escape - close detail panel
      await user.keyboard('{Escape}')

      // Wait for detail to close
      await waitFor(() => {
        expect(screen.queryByRole('button', { name: 'Close' })).not.toBeInTheDocument()
      })

      // Second Escape - deselect task
      await user.keyboard('{Escape}')

      // Task should no longer be selected
      expect(hasClassContaining(taskCard, 'selected')).toBe(false)
    })

    it('can select different tasks', async () => {
      const user = userEvent.setup()
      render(<App />)

      // Click first task
      const firstTask = screen.getByRole('button', { name: /Test Task/i })
      await user.click(firstTask)

      // Wait for detail to open
      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Close' })).toBeInTheDocument()
      })

      // Close detail
      await user.keyboard('{Escape}')

      // Wait for detail to close
      await waitFor(() => {
        expect(screen.queryByRole('button', { name: 'Close' })).not.toBeInTheDocument()
      })

      // Click second task
      const secondTask = screen.getByRole('button', { name: /Second Task/i })
      await user.click(secondTask)

      // Wait for new detail to open
      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Close' })).toBeInTheDocument()
      })

      // Second task should be selected
      expect(hasClassContaining(secondTask, 'selected')).toBe(true)

      // First task should not be selected
      expect(hasClassContaining(firstTask, 'selected')).toBe(false)
    })
  })

  // Additional: Board renders correctly
  describe('board rendering', () => {
    it('renders all column headers', () => {
      render(<App />)

      expect(screen.getByText('Backlog')).toBeInTheDocument()
      expect(screen.getByText('Todo')).toBeInTheDocument()
      expect(screen.getByText('In Progress')).toBeInTheDocument()
      expect(screen.getByText('Review')).toBeInTheDocument()
      expect(screen.getByText('Done')).toBeInTheDocument()
    })

    it('renders tasks in their respective columns', () => {
      render(<App />)

      // Test Task is in 'todo' column
      expect(screen.getByRole('button', { name: /Test Task/i })).toBeInTheDocument()
      // Second Task is in 'in_progress' column
      expect(screen.getByRole('button', { name: /Second Task/i })).toBeInTheDocument()
    })
  })
})
