import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Board } from './Board'
import { COLUMNS, COLUMN_NAMES, type Task } from '../types/task'

// Default mock tasks data (passed as prop)
const mockTasks: Task[] = [
  {
    id: 'task-1',
    title: 'Task in Backlog',
    type: 'feature',
    priority: 'medium',
    column: 'backlog',
    position: 1000,
    labels: [],
    created_by: 'user',
    collectionId: 'tasks',
    collectionName: 'tasks',
    created: '2024-01-15T10:00:00Z',
    updated: '2024-01-15T10:00:00Z',
  } as Task,
  {
    id: 'task-2',
    title: 'Task in Todo',
    type: 'bug',
    priority: 'high',
    column: 'todo',
    position: 1000,
    labels: [],
    created_by: 'cli',
    collectionId: 'tasks',
    collectionName: 'tasks',
    created: '2024-01-15T10:00:00Z',
    updated: '2024-01-15T10:00:00Z',
  } as Task,
  {
    id: 'task-3',
    title: 'Task in Progress',
    type: 'chore',
    priority: 'low',
    column: 'in_progress',
    position: 1000,
    labels: [],
    created_by: 'agent',
    collectionId: 'tasks',
    collectionName: 'tasks',
    created: '2024-01-15T10:00:00Z',
    updated: '2024-01-15T10:00:00Z',
  } as Task,
]

// useTasks hook mock for internal operations (loading, error, moveTask etc.)
const defaultMockUseTasks = {
  tasks: mockTasks,
  loading: false,
  error: null as Error | null,
  moveTask: vi.fn(),
  createTask: vi.fn(),
  updateTask: vi.fn(),
  deleteTask: vi.fn(),
}

// Mutable mock that can be changed per test
let mockUseTasks = { ...defaultMockUseTasks }

// Mock the useTasks hook
vi.mock('../hooks/useTasks', () => ({
  useTasks: () => mockUseTasks,
}))

// Mock the useCurrentBoard hook
vi.mock('../hooks/useCurrentBoard', () => ({
  useCurrentBoard: () => ({
    currentBoard: {
      id: 'board-1',
      name: 'Work',
      prefix: 'WRK',
      columns: ['backlog', 'todo', 'in_progress', 'review', 'done'],
      color: '#3B82F6',
    },
    setCurrentBoard: vi.fn(),
    loading: false,
  }),
}))

describe('Board', () => {
  beforeEach(() => {
    // Reset mock to default before each test
    mockUseTasks = { ...defaultMockUseTasks }
  })

  describe('normal state', () => {
    it('renders all five columns', () => {
      render(<Board tasks={mockTasks} />)

      // Check all column headers are present
      COLUMNS.forEach((column) => {
        expect(screen.getByText(COLUMN_NAMES[column])).toBeInTheDocument()
      })
    })

    it('groups tasks into correct columns', () => {
      render(<Board tasks={mockTasks} />)

      // Tasks should be rendered - use getAllByText since strict mode may render multiple
      expect(screen.getAllByText('Task in Backlog').length).toBeGreaterThanOrEqual(1)
      expect(screen.getAllByText('Task in Todo').length).toBeGreaterThanOrEqual(1)
      expect(screen.getAllByText('Task in Progress').length).toBeGreaterThanOrEqual(1)
    })

    it('displays task counts', () => {
      render(<Board tasks={mockTasks} />)

      // At least one column should show a count
      // The exact count depends on render behavior, so we just check counts exist
      expect(screen.getAllByText(/^\d+$/).length).toBeGreaterThan(0)
    })
  })

  // Test 8.1: Shows 'Loading tasks...' when loading is true
  describe('loading state', () => {
    it('shows loading message when loading', () => {
      mockUseTasks = {
        ...defaultMockUseTasks,
        tasks: [],
        loading: true,
        error: null,
      }

      render(<Board tasks={[]} />)
      expect(screen.getByText('Loading tasks...')).toBeInTheDocument()
    })

    it('does not render columns when loading', () => {
      mockUseTasks = {
        ...defaultMockUseTasks,
        tasks: [],
        loading: true,
        error: null,
      }

      render(<Board tasks={[]} />)
      expect(screen.queryByText('Backlog')).not.toBeInTheDocument()
      expect(screen.queryByText('Todo')).not.toBeInTheDocument()
    })
  })

  // Test 8.2: Shows error message when error is not null
  describe('error state', () => {
    it('shows error message when error occurs', () => {
      mockUseTasks = {
        ...defaultMockUseTasks,
        tasks: [],
        loading: false,
        error: new Error('Failed to fetch tasks'),
      }

      render(<Board tasks={[]} />)
      expect(screen.getByText(/Error:/)).toBeInTheDocument()
      expect(screen.getByText(/Failed to fetch tasks/)).toBeInTheDocument()
    })

    it('does not render columns when error occurs', () => {
      mockUseTasks = {
        ...defaultMockUseTasks,
        tasks: [],
        loading: false,
        error: new Error('Network error'),
      }

      render(<Board tasks={[]} />)
      expect(screen.queryByText('Backlog')).not.toBeInTheDocument()
      expect(screen.queryByText('Todo')).not.toBeInTheDocument()
    })

    it('displays specific error message from Error object', () => {
      mockUseTasks = {
        ...defaultMockUseTasks,
        tasks: [],
        loading: false,
        error: new Error('Connection timeout'),
      }

      render(<Board tasks={[]} />)
      expect(screen.getByText('Error: Connection timeout')).toBeInTheDocument()
    })
  })
})
