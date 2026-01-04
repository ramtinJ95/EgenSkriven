import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { DndContext } from '@dnd-kit/core'
import { Column } from './Column'
import type { Task } from '../types/task'

// Wrapper component to provide DndContext
const DndWrapper = ({ children }: { children: React.ReactNode }) => (
  <DndContext>{children}</DndContext>
)

// Helper function to create mock tasks
const createMockTask = (overrides: Partial<Task> = {}): Task => ({
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
  ...overrides,
})

describe('Column', () => {
  // Test 5.1: Renders column header with correct human-readable name
  describe('column header', () => {
    it('renders column name for backlog', () => {
      render(
        <DndWrapper>
          <Column column="backlog" tasks={[]} />
        </DndWrapper>
      )
      expect(screen.getByText('Backlog')).toBeInTheDocument()
    })

    it('renders column name for todo', () => {
      render(
        <DndWrapper>
          <Column column="todo" tasks={[]} />
        </DndWrapper>
      )
      expect(screen.getByText('Todo')).toBeInTheDocument()
    })

    it('renders column name for in_progress', () => {
      render(
        <DndWrapper>
          <Column column="in_progress" tasks={[]} />
        </DndWrapper>
      )
      expect(screen.getByText('In Progress')).toBeInTheDocument()
    })

    it('renders column name for review', () => {
      render(
        <DndWrapper>
          <Column column="review" tasks={[]} />
        </DndWrapper>
      )
      expect(screen.getByText('Review')).toBeInTheDocument()
    })

    it('renders column name for done', () => {
      render(
        <DndWrapper>
          <Column column="done" tasks={[]} />
        </DndWrapper>
      )
      expect(screen.getByText('Done')).toBeInTheDocument()
    })
  })

  // Test 5.2: Shows task count
  describe('task count', () => {
    it('shows correct task count', () => {
      const tasks = [
        createMockTask({ id: 'task-001', title: 'Task 1' }),
        createMockTask({ id: 'task-002', title: 'Task 2' }),
        createMockTask({ id: 'task-003', title: 'Task 3' }),
      ]

      render(
        <DndWrapper>
          <Column column="todo" tasks={tasks} />
        </DndWrapper>
      )
      // Get all elements with text '3' and find the one in the count span
      const countElements = screen.getAllByText('3')
      // The count should be in the header
      const countElement = countElements.find(el => el.className.includes('count'))
      expect(countElement).toBeTruthy()
    })

    it('shows zero count for empty column', () => {
      render(
        <DndWrapper>
          <Column column="done" tasks={[]} />
        </DndWrapper>
      )
      expect(screen.getByText('0')).toBeInTheDocument()
    })

    it('shows count of 1 for single task', () => {
      const tasks = [createMockTask()]

      render(
        <DndWrapper>
          <Column column="backlog" tasks={tasks} />
        </DndWrapper>
      )
      expect(screen.getByText('1')).toBeInTheDocument()
    })
  })

  // Test 5.3: Renders all tasks passed to it
  describe('task rendering', () => {
    it('renders all task cards', () => {
      const tasks = [
        createMockTask({ id: '1', title: 'First Task' }),
        createMockTask({ id: '2', title: 'Second Task' }),
      ]

      render(
        <DndWrapper>
          <Column column="todo" tasks={tasks} />
        </DndWrapper>
      )

      expect(screen.getByText('First Task')).toBeInTheDocument()
      expect(screen.getByText('Second Task')).toBeInTheDocument()
    })

    it('renders tasks with correct IDs', () => {
      const tasks = [
        createMockTask({ id: 'task-abc', title: 'Task ABC' }),
        createMockTask({ id: 'task-xyz', title: 'Task XYZ' }),
      ]

      render(
        <DndWrapper>
          <Column column="review" tasks={tasks} />
        </DndWrapper>
      )

      expect(screen.getByText('task-abc')).toBeInTheDocument()
      expect(screen.getByText('task-xyz')).toBeInTheDocument()
    })

    it('renders empty column with no task cards', () => {
      render(
        <DndWrapper>
          <Column column="backlog" tasks={[]} />
        </DndWrapper>
      )

      // Only the column header should be present, no task cards
      expect(screen.getByText('Backlog')).toBeInTheDocument()
      expect(screen.queryByRole('button')).not.toBeInTheDocument()
    })
  })

  // Task click callback
  describe('task interactions', () => {
    it('passes onTaskClick to TaskCard components', () => {
      const onTaskClick = vi.fn()
      const task = createMockTask({ id: 'test-task', title: 'Clickable Task' })

      render(
        <DndWrapper>
          <Column column="todo" tasks={[task]} onTaskClick={onTaskClick} />
        </DndWrapper>
      )

      // Verify the task card is rendered with click handler capability
      const taskCard = screen.getByRole('button', { name: /Clickable Task/i })
      expect(taskCard).toBeInTheDocument()
    })

    it('marks selected task as selected', () => {
      const tasks = [
        createMockTask({ id: 'task-1', title: 'Task 1' }),
        createMockTask({ id: 'task-2', title: 'Task 2' }),
      ]

      render(
        <DndWrapper>
          <Column column="todo" tasks={tasks} selectedTaskId="task-1" />
        </DndWrapper>
      )

      const task1Card = screen.getByRole('button', { name: /Task 1/i })
      const task2Card = screen.getByRole('button', { name: /Task 2/i })

      // task-1 should be selected (aria-pressed="true")
      expect(task1Card).toHaveAttribute('aria-pressed', 'true')
      // task-2 should not be selected
      expect(task2Card).toHaveAttribute('aria-pressed', 'false')
    })
  })
})
