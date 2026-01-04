import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { TaskCard } from './TaskCard'
import type { Task } from '../types/task'

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
})
