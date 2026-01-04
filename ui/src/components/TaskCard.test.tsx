import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { TaskCard } from './TaskCard'
import type { Task } from '../types/task'
import type { Board } from '../types/board'

// Mock task for testing
const mockTask: Task = {
  id: 'test-123',
  title: 'Test Task Title',
  type: 'feature',
  priority: 'high',
  column: 'todo',
  position: 1000,
  labels: ['frontend', 'ui'],
  created_by: 'user',
  collectionId: 'tasks',
  collectionName: 'tasks',
  created: '2024-01-15T10:00:00Z',
  updated: '2024-01-15T10:00:00Z',
}

// Mock task with board relation
const mockTaskWithBoard: Task = {
  ...mockTask,
  id: 'task-with-board',
  board: 'board-123',
  seq: 42,
}

// Mock board for testing
const mockBoard: Board = {
  id: 'board-123',
  name: 'Work',
  prefix: 'WRK',
  columns: ['backlog', 'todo', 'in_progress', 'review', 'done'],
  color: '#3B82F6',
  collectionId: 'boards',
  collectionName: 'boards',
  created: '2024-01-15T10:00:00Z',
  updated: '2024-01-15T10:00:00Z',
}

describe('TaskCard', () => {
  it('renders task title', () => {
    render(<TaskCard task={mockTask} />)
    expect(screen.getByText('Test Task Title')).toBeInTheDocument()
  })

  it('renders task ID (truncated)', () => {
    render(<TaskCard task={mockTask} />)
    expect(screen.getByText('test-123')).toBeInTheDocument()
  })

  it('renders labels', () => {
    render(<TaskCard task={mockTask} />)
    expect(screen.getByText('frontend')).toBeInTheDocument()
    expect(screen.getByText('ui')).toBeInTheDocument()
  })

  it('renders priority indicator', () => {
    render(<TaskCard task={mockTask} />)
    expect(screen.getByText('High')).toBeInTheDocument()
  })

  it('renders task type', () => {
    render(<TaskCard task={mockTask} />)
    expect(screen.getByText('feature')).toBeInTheDocument()
  })

  describe('display ID', () => {
    it('shows display ID when board and seq are provided', () => {
      render(<TaskCard task={mockTaskWithBoard} currentBoard={mockBoard} />)
      expect(screen.getByText('WRK-42')).toBeInTheDocument()
    })

    it('falls back to short ID when no board provided', () => {
      render(<TaskCard task={mockTaskWithBoard} />)
      // Should show first 8 chars of ID
      expect(screen.getByText('task-wit')).toBeInTheDocument()
    })

    it('falls back to short ID when task has no seq', () => {
      const taskWithoutSeq: Task = { ...mockTask, board: 'board-123' }
      render(<TaskCard task={taskWithoutSeq} currentBoard={mockBoard} />)
      expect(screen.getByText('test-123')).toBeInTheDocument()
    })

    it('falls back to short ID when currentBoard is null', () => {
      render(<TaskCard task={mockTaskWithBoard} currentBoard={null} />)
      expect(screen.getByText('task-wit')).toBeInTheDocument()
    })
  })
})
