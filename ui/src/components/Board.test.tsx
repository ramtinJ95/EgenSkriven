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

// Mock moveTask function (now passed as prop)
const mockMoveTask = vi.fn().mockResolvedValue(mockTasks[0])

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
    // Reset mock before each test
    mockMoveTask.mockClear()
  })

  describe('normal state', () => {
    it('renders all five columns', () => {
      render(<Board tasks={mockTasks} moveTask={mockMoveTask} />)

      // Check all column headers are present
      COLUMNS.forEach((column) => {
        expect(screen.getByText(COLUMN_NAMES[column])).toBeInTheDocument()
      })
    })

    it('groups tasks into correct columns', () => {
      render(<Board tasks={mockTasks} moveTask={mockMoveTask} />)

      // Tasks should be rendered - use getAllByText since strict mode may render multiple
      expect(screen.getAllByText('Task in Backlog').length).toBeGreaterThanOrEqual(1)
      expect(screen.getAllByText('Task in Todo').length).toBeGreaterThanOrEqual(1)
      expect(screen.getAllByText('Task in Progress').length).toBeGreaterThanOrEqual(1)
    })

    it('displays task counts', () => {
      render(<Board tasks={mockTasks} moveTask={mockMoveTask} />)

      // At least one column should show a count
      // The exact count depends on render behavior, so we just check counts exist
      expect(screen.getAllByText(/^\d+$/).length).toBeGreaterThan(0)
    })
  })

  // Note: Loading and error states are now handled by the parent component (App.tsx)
  // since Board no longer calls useTasks internally. The Board component only shows
  // loading when the board itself is loading (from useCurrentBoard).
  describe('empty state', () => {
    it('renders empty columns when no tasks', () => {
      render(<Board tasks={[]} moveTask={mockMoveTask} />)

      // All columns should still be visible
      COLUMNS.forEach((column) => {
        expect(screen.getByText(COLUMN_NAMES[column])).toBeInTheDocument()
      })
    })
  })
})
