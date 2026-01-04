import { useState, useEffect, useCallback } from 'react'
import { Layout } from './components/Layout'
import { Board } from './components/Board'
import { QuickCreate } from './components/QuickCreate'
import { TaskDetail } from './components/TaskDetail'
import { useTasks } from './hooks/useTasks'
import type { Task, Column } from './types/task'

/**
 * Main application component.
 * 
 * Manages:
 * - Quick create modal (opened with 'C' key)
 * - Task detail panel (opened by clicking a task or pressing Enter)
 * - Selected task state (for keyboard navigation)
 * - Global keyboard shortcuts
 */
function App() {
  const { tasks, createTask, updateTask } = useTasks()
  const [isQuickCreateOpen, setIsQuickCreateOpen] = useState(false)
  const [selectedTask, setSelectedTask] = useState<Task | null>(null)
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null)

  // Global keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Don't trigger shortcuts when typing in inputs
      if (
        e.target instanceof HTMLInputElement ||
        e.target instanceof HTMLTextAreaElement ||
        e.target instanceof HTMLSelectElement
      ) {
        return
      }

      // 'C' to open quick create
      if (e.key === 'c' || e.key === 'C') {
        e.preventDefault()
        setIsQuickCreateOpen(true)
      }

      // 'Enter' to open selected task detail
      if (e.key === 'Enter' && selectedTaskId && !selectedTask) {
        e.preventDefault()
        const task = tasks.find((t) => t.id === selectedTaskId)
        if (task) {
          setSelectedTask(task)
        }
      }

      // 'Escape' to deselect task (when detail panel is closed)
      if (e.key === 'Escape' && !selectedTask && selectedTaskId) {
        e.preventDefault()
        setSelectedTaskId(null)
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [selectedTaskId, selectedTask, tasks])

  // Handle task creation
  const handleCreate = useCallback(
    async (title: string, column: Column) => {
      await createTask(title, column)
    },
    [createTask]
  )

  // Handle task update
  const handleUpdate = useCallback(
    async (id: string, data: Partial<Task>) => {
      await updateTask(id, data)
    },
    [updateTask]
  )

  // Handle task click to open detail panel
  const handleTaskClick = useCallback((task: Task) => {
    setSelectedTaskId(task.id)
    setSelectedTask(task)
  }, [])

  // Handle task selection (without opening detail)
  const handleTaskSelect = useCallback((task: Task) => {
    setSelectedTaskId(task.id)
  }, [])

  // Handle closing detail panel
  const handleCloseDetail = useCallback(() => {
    setSelectedTask(null)
    // Keep selectedTaskId so user can press Enter to reopen
  }, [])

  return (
    <Layout>
      <Board 
        onTaskClick={handleTaskClick} 
        selectedTaskId={selectedTaskId}
        onTaskSelect={handleTaskSelect}
      />

      {/* Quick Create Modal */}
      <QuickCreate
        isOpen={isQuickCreateOpen}
        onClose={() => setIsQuickCreateOpen(false)}
        onCreate={handleCreate}
      />

      {/* Task Detail Panel */}
      <TaskDetail
        task={selectedTask}
        onClose={handleCloseDetail}
        onUpdate={handleUpdate}
      />
    </Layout>
  )
}

export default App
