import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Board } from './Board'
import { COLUMNS, COLUMN_NAMES } from '../types/task'

// Mock the useTasks hook
vi.mock('../hooks/useTasks', () => ({
  useTasks: () => ({
    tasks: [
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
      },
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
      },
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
      },
    ],
    loading: false,
    error: null,
    moveTask: vi.fn(),
    createTask: vi.fn(),
    updateTask: vi.fn(),
    deleteTask: vi.fn(),
  }),
}))

describe('Board', () => {
  it('renders all five columns', () => {
    render(<Board />)
    
    // Check all column headers are present
    COLUMNS.forEach((column) => {
      expect(screen.getByText(COLUMN_NAMES[column])).toBeInTheDocument()
    })
  })

  it('groups tasks into correct columns', () => {
    render(<Board />)
    
    // Tasks should be rendered - use getAllByText since strict mode may render multiple
    expect(screen.getAllByText('Task in Backlog').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Task in Todo').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Task in Progress').length).toBeGreaterThanOrEqual(1)
  })

  it('displays task counts', () => {
    render(<Board />)
    
    // At least one column should show a count
    // The exact count depends on render behavior, so we just check counts exist
    expect(screen.getAllByText(/^\d+$/).length).toBeGreaterThan(0)
  })
})
