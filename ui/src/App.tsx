import { useState, useCallback, useMemo } from 'react'
import { SelectionProvider, CurrentBoardProvider, useCurrentBoard, ThemeProvider } from './contexts'
import { useSelection } from './hooks/useSelection'
import { useKeyboardShortcuts } from './hooks/useKeyboard'
import { useTasks } from './hooks/useTasks'
import { useBoards } from './hooks/useBoards'
import { useFilteredTasks, useSearchDebounce } from './hooks/useFilteredTasks'
import { useFilterStore } from './stores/filters'
import { Layout } from './components/Layout'
import { Board } from './components/Board'
import { ListView } from './components/ListView'
import { TaskDetail } from './components/TaskDetail'
import { QuickCreate } from './components/QuickCreate'
import { CommandPalette, type Command } from './components/CommandPalette'
import {
  PropertyPicker,
  STATUS_OPTIONS,
  PRIORITY_OPTIONS,
  TYPE_OPTIONS,
} from './components/PropertyPicker'
import { ShortcutsHelp } from './components/ShortcutsHelp'
import { PeekPreview } from './components/PeekPreview'
import { FilterBuilder } from './components/FilterBuilder'
import { DisplayOptions } from './components/DisplayOptions'
import { Settings } from './components/Settings'
import { COLUMNS, type Task, type Column } from './types/task'

/**
 * Inner app content that uses selection context.
 * Separated from App to allow useSelection hook to work within SelectionProvider.
 */
function AppContent() {
  const { boards } = useBoards()
  const { currentBoard, setCurrentBoard } = useCurrentBoard()
  const { tasks, loading, createTask, updateTask, deleteTask, moveTask } = useTasks(currentBoard?.id)
  const {
    selectedTaskId,
    selectTask,
    clearSelection,
    toggleMultiSelect,
    selectRange,
    selectAll,
    setFocusedColumn,
    focusedColumn,
    multiSelectedIds,
    isSelected,
  } = useSelection()

  // Filter state and hooks
  const displayOptions = useFilterStore((s) => s.displayOptions)
  const setDisplayOptions = useFilterStore((s) => s.setDisplayOptions)
  const clearView = useFilterStore((s) => s.clearView)

  // Apply search debouncing (updates debouncedSearchQuery in store)
  useSearchDebounce()

  // Filter tasks based on active filters and search
  // This is used for ListView and navigation; Board uses its own internal filtering
  const filteredTasks = useFilteredTasks(tasks)

  // Modal states
  const [isCommandPaletteOpen, setIsCommandPaletteOpen] = useState(false)
  const [isQuickCreateOpen, setIsQuickCreateOpen] = useState(false)
  const [isShortcutsHelpOpen, setIsShortcutsHelpOpen] = useState(false)
  const [isDetailOpen, setIsDetailOpen] = useState(false)
  const [isPeekOpen, setIsPeekOpen] = useState(false)
  const [isFilterBuilderOpen, setIsFilterBuilderOpen] = useState(false)
  const [isDisplayOptionsOpen, setIsDisplayOptionsOpen] = useState(false)
  const [isSettingsOpen, setIsSettingsOpen] = useState(false)

  // Property picker states
  const [statusPickerOpen, setStatusPickerOpen] = useState(false)
  const [priorityPickerOpen, setPriorityPickerOpen] = useState(false)
  const [typePickerOpen, setTypePickerOpen] = useState(false)

  // Anchor element for property pickers (the selected task card)
  // Using state instead of ref to avoid accessing ref.current during render
  const [anchorElement, setAnchorElement] = useState<HTMLElement | null>(null)

  // Get the currently selected task
  const selectedTask = useMemo(
    () => tasks.find((t) => t.id === selectedTaskId) || null,
    [tasks, selectedTaskId]
  )

  // Get sorted task IDs for navigation (use filtered tasks for navigation)
  const sortedTaskIds = useMemo(() => filteredTasks.map((t) => t.id), [filteredTasks])

  // Get tasks grouped by column for column navigation
  const tasksByColumn = useMemo(() => {
    const grouped: Record<Column, Task[]> = {
      backlog: [],
      todo: [],
      in_progress: [],
      review: [],
      done: [],
    }
    for (const task of filteredTasks) {
      grouped[task.column]?.push(task)
    }
    return grouped
  }, [filteredTasks])

  // Navigation helpers
  const navigateToNextTask = useCallback(() => {
    if (!selectedTaskId) {
      if (sortedTaskIds.length > 0) {
        selectTask(sortedTaskIds[0])
      }
      return
    }

    const currentIndex = sortedTaskIds.indexOf(selectedTaskId)
    if (currentIndex < sortedTaskIds.length - 1) {
      selectTask(sortedTaskIds[currentIndex + 1])
    }
  }, [selectedTaskId, sortedTaskIds, selectTask])

  const navigateToPrevTask = useCallback(() => {
    if (!selectedTaskId) {
      if (sortedTaskIds.length > 0) {
        selectTask(sortedTaskIds[sortedTaskIds.length - 1])
      }
      return
    }

    const currentIndex = sortedTaskIds.indexOf(selectedTaskId)
    if (currentIndex > 0) {
      selectTask(sortedTaskIds[currentIndex - 1])
    }
  }, [selectedTaskId, sortedTaskIds, selectTask])

  // Column navigation helpers
  const navigateToNextColumn = useCallback(() => {
    // Find current column
    const currentColumn: Column = selectedTask?.column || (focusedColumn as Column) || COLUMNS[0]
    const currentIndex = COLUMNS.indexOf(currentColumn)

    // Find the next column with tasks (skip empty columns)
    for (let i = currentIndex + 1; i < COLUMNS.length; i++) {
      const nextColumn = COLUMNS[i]
      const tasksInColumn = tasksByColumn[nextColumn]
      if (tasksInColumn.length > 0) {
        setFocusedColumn(nextColumn)
        selectTask(tasksInColumn[0].id)
        return
      }
    }
    // If no column with tasks found to the right, just move to the next column
    // (for visual column focus even if empty)
    if (currentIndex < COLUMNS.length - 1) {
      setFocusedColumn(COLUMNS[currentIndex + 1])
    }
  }, [selectedTask, focusedColumn, tasksByColumn, setFocusedColumn, selectTask])

  const navigateToPrevColumn = useCallback(() => {
    // Find current column
    const currentColumn: Column = selectedTask?.column || (focusedColumn as Column) || COLUMNS[0]
    const currentIndex = COLUMNS.indexOf(currentColumn)

    // Find the previous column with tasks (skip empty columns)
    for (let i = currentIndex - 1; i >= 0; i--) {
      const prevColumn = COLUMNS[i]
      const tasksInColumn = tasksByColumn[prevColumn]
      if (tasksInColumn.length > 0) {
        setFocusedColumn(prevColumn)
        selectTask(tasksInColumn[0].id)
        return
      }
    }
    // If no column with tasks found to the left, just move to the previous column
    // (for visual column focus even if empty)
    if (currentIndex > 0) {
      setFocusedColumn(COLUMNS[currentIndex - 1])
    }
  }, [selectedTask, focusedColumn, tasksByColumn, setFocusedColumn, selectTask])

  // Action handlers
  const openTaskDetail = useCallback(() => {
    if (selectedTaskId) {
      setIsDetailOpen(true)
      setIsPeekOpen(false)
    }
  }, [selectedTaskId])

  const handleCreateTask = useCallback(
    async (title: string, column: Column, description?: string) => {
      const newTask = await createTask(title, column, description)
      setIsQuickCreateOpen(false)
      selectTask(newTask.id)
    },
    [createTask, selectTask]
  )

  const handleDeleteTask = useCallback(async () => {
    if (selectedTaskId && window.confirm('Delete this task?')) {
      await deleteTask(selectedTaskId)
      clearSelection()
    }
  }, [selectedTaskId, deleteTask, clearSelection])

  const handleStatusChange = useCallback(
    async (status: string) => {
      if (selectedTaskId) {
        await updateTask(selectedTaskId, { column: status as Column })
      }
    },
    [selectedTaskId, updateTask]
  )

  const handlePriorityChange = useCallback(
    async (priority: string) => {
      if (selectedTaskId) {
        await updateTask(selectedTaskId, {
          priority: priority as Task['priority'],
        })
      }
    },
    [selectedTaskId, updateTask]
  )

  const handleTypeChange = useCallback(
    async (type: string) => {
      if (selectedTaskId) {
        await updateTask(selectedTaskId, { type: type as Task['type'] })
      }
    },
    [selectedTaskId, updateTask]
  )

  const handleTaskClick = useCallback(
    (task: Task) => {
      selectTask(task.id)
      setIsDetailOpen(true)
    },
    [selectTask]
  )

  const handleTaskSelect = useCallback(
    (task: Task) => {
      selectTask(task.id)
      // Store reference to the task card element for anchor positioning
      const element = document.querySelector(`[data-task-id="${task.id}"]`)
      if (element instanceof HTMLElement) {
        setAnchorElement(element)
      }
    },
    [selectTask]
  )

  // Build command palette commands
  const commands: Command[] = useMemo(
    () => [
      // Actions
      {
        id: 'create-task',
        label: 'Create task',
        shortcut: { key: 'c' },
        section: 'actions',
        icon: '+',
        action: () => setIsQuickCreateOpen(true),
      },
      {
        id: 'change-status',
        label: 'Change status',
        shortcut: { key: 's' },
        section: 'actions',
        icon: '●',
        action: () => setStatusPickerOpen(true),
        when: () => !!selectedTaskId,
      },
      {
        id: 'set-priority',
        label: 'Set priority',
        shortcut: { key: 'p' },
        section: 'actions',
        icon: '!',
        action: () => setPriorityPickerOpen(true),
        when: () => !!selectedTaskId,
      },
      {
        id: 'set-type',
        label: 'Set type',
        shortcut: { key: 't' },
        section: 'actions',
        icon: '◆',
        action: () => setTypePickerOpen(true),
        when: () => !!selectedTaskId,
      },
      {
        id: 'delete-task',
        label: 'Delete task',
        shortcut: { key: 'Backspace' },
        section: 'actions',
        icon: '×',
        action: handleDeleteTask,
        when: () => !!selectedTaskId,
      },
      {
        id: 'open-filter',
        label: 'Open filter builder',
        shortcut: { key: 'f' },
        section: 'actions',
        icon: '⊛',
        action: () => setIsFilterBuilderOpen(true),
      },
      {
        id: 'toggle-view',
        label: 'Toggle board/list view',
        shortcut: { key: 'b', meta: true },
        section: 'actions',
        icon: '☰',
        action: () =>
          setDisplayOptions({
            viewMode: displayOptions.viewMode === 'board' ? 'list' : 'board',
          }),
      },
      {
        id: 'show-shortcuts',
        label: 'Show keyboard shortcuts',
        shortcut: { key: '?' },
        section: 'actions',
        icon: '?',
        action: () => setIsShortcutsHelpOpen(true),
      },

      // Board switching commands
      ...boards.map((board) => ({
        id: `board-${board.id}`,
        label: `${board.name} (${board.prefix})`,
        section: 'boards' as const,
        icon: currentBoard?.id === board.id ? '●' : '○',
        action: () => {
          setCurrentBoard(board)
          clearSelection()
          // Clear filters when switching boards to prevent stale filter state
          clearView()
        },
      })),

      // Navigation - add recent tasks
      ...tasks.slice(0, 5).map((task) => ({
        id: `task-${task.id}`,
        label: `${task.id.substring(0, 8)} ${task.title}`,
        section: 'recent' as const,
        action: () => {
          selectTask(task.id)
          setIsDetailOpen(true)
        },
      })),
    ],
    [
      tasks,
      selectedTaskId,
      handleDeleteTask,
      selectTask,
      boards,
      currentBoard,
      setCurrentBoard,
      clearSelection,
      clearView,
      displayOptions.viewMode,
      setDisplayOptions,
    ]
  )

  // Register keyboard shortcuts
  useKeyboardShortcuts([
    // Global shortcuts
    {
      combo: { key: 'k', meta: true },
      handler: () => setIsCommandPaletteOpen(true),
      description: 'Open command palette',
    },
    {
      combo: { key: ',', meta: true },
      handler: () => setIsSettingsOpen(true),
      description: 'Open settings',
    },
    {
      combo: { key: '?' },
      handler: () => setIsShortcutsHelpOpen(true),
      description: 'Show shortcuts help',
    },
    {
      combo: { key: 'Escape' },
      handler: () => {
        // Handle Escape in priority order
        // Return false if nothing was done, allowing event to propagate
        if (isSettingsOpen) {
          setIsSettingsOpen(false)
          return true
        } else if (isFilterBuilderOpen) {
          setIsFilterBuilderOpen(false)
          return true
        } else if (isDisplayOptionsOpen) {
          setIsDisplayOptionsOpen(false)
          return true
        } else if (isPeekOpen) {
          setIsPeekOpen(false)
          return true
        } else if (isDetailOpen) {
          setIsDetailOpen(false)
          return true
        } else if (isQuickCreateOpen) {
          setIsQuickCreateOpen(false)
          return true
        } else if (isCommandPaletteOpen) {
          setIsCommandPaletteOpen(false)
          return true
        } else if (isShortcutsHelpOpen) {
          setIsShortcutsHelpOpen(false)
          return true
        } else if (statusPickerOpen) {
          setStatusPickerOpen(false)
          return true
        } else if (priorityPickerOpen) {
          setPriorityPickerOpen(false)
          return true
        } else if (typePickerOpen) {
          setTypePickerOpen(false)
          return true
        } else if (selectedTaskId) {
          clearSelection()
          return true
        }
        // Nothing was closed, allow event to propagate
        return false
      },
      description: 'Close/deselect',
      allowInInput: true,
    },

    // Task actions
    {
      combo: { key: 'c' },
      handler: () => setIsQuickCreateOpen(true),
      description: 'Create task',
    },
    {
      combo: { key: 'Enter' },
      handler: () => openTaskDetail(),
      description: 'Open selected task',
    },
    {
      combo: { key: ' ' },
      handler: () => {
        if (selectedTaskId) {
          setIsPeekOpen((prev) => !prev)
        }
      },
      description: 'Peek preview',
    },
    {
      combo: { key: 'Backspace' },
      handler: () => {
        handleDeleteTask()
      },
      description: 'Delete task',
    },

    // Filter shortcuts
    {
      combo: { key: 'f' },
      handler: () => setIsFilterBuilderOpen(true),
      description: 'Open filter builder',
    },
    {
      combo: { key: 'b', meta: true },
      handler: () => {
        setDisplayOptions({
          viewMode: displayOptions.viewMode === 'board' ? 'list' : 'board',
        })
      },
      description: 'Toggle board/list view',
    },

    // Property shortcuts (only when task selected)
    {
      combo: { key: 's' },
      handler: () => setStatusPickerOpen(true),
      when: () => !!selectedTaskId,
      description: 'Set status',
    },
    {
      combo: { key: 'p' },
      handler: () => setPriorityPickerOpen(true),
      when: () => !!selectedTaskId,
      description: 'Set priority',
    },
    {
      combo: { key: 't' },
      handler: () => setTypePickerOpen(true),
      when: () => !!selectedTaskId,
      description: 'Set type',
    },

    // Navigation
    {
      combo: { key: 'j' },
      handler: () => navigateToNextTask(),
      description: 'Next task',
    },
    {
      combo: { key: 'ArrowDown' },
      handler: () => navigateToNextTask(),
      description: 'Next task',
    },
    {
      combo: { key: 'k' },
      handler: () => navigateToPrevTask(),
      description: 'Previous task',
    },
    {
      combo: { key: 'ArrowUp' },
      handler: () => navigateToPrevTask(),
      description: 'Previous task',
    },
    {
      combo: { key: 'h' },
      handler: () => navigateToPrevColumn(),
      description: 'Previous column',
    },
    {
      combo: { key: 'l' },
      handler: () => navigateToNextColumn(),
      description: 'Next column',
    },

    // Selection shortcuts
    {
      combo: { key: 'x' },
      handler: () => {
        if (selectedTaskId) {
          toggleMultiSelect(selectedTaskId)
        }
      },
      when: () => !!selectedTaskId,
      description: 'Toggle select task',
    },
    {
      combo: { key: 'x', shift: true },
      handler: () => {
        // Range select from first multi-selected to current
        if (selectedTaskId && sortedTaskIds.length > 0) {
          const firstSelected = sortedTaskIds[0]
          selectRange(firstSelected, selectedTaskId, sortedTaskIds)
        }
      },
      when: () => !!selectedTaskId,
      description: 'Select range',
    },
    {
      combo: { key: 'a', meta: true },
      handler: () => {
        selectAll(sortedTaskIds)
      },
      description: 'Select all visible',
    },

    // Edit title shortcut
    {
      combo: { key: 'e' },
      handler: () => {
        if (selectedTaskId) {
          setIsDetailOpen(true)
          setIsPeekOpen(false)
        }
      },
      when: () => !!selectedTaskId,
      description: 'Edit title (open detail)',
    },
  ])

  if (loading) {
    return (
      <Layout
        totalTasks={0}
        filteredTasks={0}
        onOpenFilterBuilder={() => {}}
        onOpenDisplayOptions={() => {}}
        onOpenSettings={() => setIsSettingsOpen(true)}
      >
        <div
          style={{
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            height: '100%',
          }}
        >
          Loading...
        </div>
      </Layout>
    )
  }

  return (
    <Layout
      totalTasks={tasks.length}
      filteredTasks={filteredTasks.length}
      onOpenFilterBuilder={() => setIsFilterBuilderOpen(true)}
      onOpenDisplayOptions={() => setIsDisplayOptionsOpen(true)}
      onOpenSettings={() => setIsSettingsOpen(true)}
    >
      {/* Selection count indicator - shows when multiple tasks are selected */}
      {multiSelectedIds.size > 0 && (
        <div style={{
          position: 'fixed',
          bottom: 'var(--space-4)',
          left: '50%',
          transform: 'translateX(-50%)',
          padding: 'var(--space-2) var(--space-4)',
          backgroundColor: 'var(--accent)',
          color: 'white',
          borderRadius: 'var(--radius-md)',
          fontSize: 'var(--text-sm)',
          fontWeight: 'var(--font-medium)',
          boxShadow: 'var(--shadow-lg)',
          zIndex: 100,
          display: 'flex',
          alignItems: 'center',
          gap: 'var(--space-2)',
        }}>
          <span>{multiSelectedIds.size} task{multiSelectedIds.size > 1 ? 's' : ''} selected</span>
          <button
            onClick={() => clearSelection()}
            style={{
              background: 'rgba(255,255,255,0.2)',
              border: 'none',
              borderRadius: 'var(--radius-sm)',
              padding: '2px 8px',
              color: 'white',
              cursor: 'pointer',
              fontSize: 'var(--text-xs)',
            }}
          >
            Clear
          </button>
        </div>
      )}

      {/* Conditional rendering: Board or List view */}
      {displayOptions.viewMode === 'board' ? (
        <Board
          tasks={filteredTasks}
          onTaskClick={handleTaskClick}
          onTaskSelect={handleTaskSelect}
          selectedTaskId={selectedTaskId}
          isSelected={isSelected}
          moveTask={moveTask}
        />
      ) : (
        <ListView
          tasks={filteredTasks}
          onTaskClick={handleTaskClick}
          selectedTaskId={selectedTaskId}
        />
      )}

      {/* Task Detail Panel */}
      <TaskDetail
        task={isDetailOpen ? selectedTask : null}
        tasks={tasks}
        onClose={() => setIsDetailOpen(false)}
        onUpdate={async (id, data) => {
          await updateTask(id, data)
        }}
        onTaskClick={(task) => {
          selectTask(task.id)
        }}
      />

      {/* Quick Create Modal */}
      <QuickCreate
        isOpen={isQuickCreateOpen}
        onClose={() => setIsQuickCreateOpen(false)}
        onCreate={handleCreateTask}
      />

      {/* Command Palette */}
      <CommandPalette
        isOpen={isCommandPaletteOpen}
        onClose={() => setIsCommandPaletteOpen(false)}
        commands={commands}
      />

      {/* Filter Builder Modal */}
      <FilterBuilder
        isOpen={isFilterBuilderOpen}
        onClose={() => setIsFilterBuilderOpen(false)}
        tasks={tasks}
      />

      {/* Display Options Menu */}
      <DisplayOptions
        isOpen={isDisplayOptionsOpen}
        onClose={() => setIsDisplayOptionsOpen(false)}
      />

      {/* Property Pickers */}
      <PropertyPicker
        isOpen={statusPickerOpen}
        onClose={() => setStatusPickerOpen(false)}
        onSelect={handleStatusChange}
        options={STATUS_OPTIONS}
        currentValue={selectedTask?.column}
        title="Set Status"
        anchorElement={anchorElement}
      />

      <PropertyPicker
        isOpen={priorityPickerOpen}
        onClose={() => setPriorityPickerOpen(false)}
        onSelect={handlePriorityChange}
        options={PRIORITY_OPTIONS}
        currentValue={selectedTask?.priority}
        title="Set Priority"
        anchorElement={anchorElement}
      />

      <PropertyPicker
        isOpen={typePickerOpen}
        onClose={() => setTypePickerOpen(false)}
        onSelect={handleTypeChange}
        options={TYPE_OPTIONS}
        currentValue={selectedTask?.type}
        title="Set Type"
        anchorElement={anchorElement}
      />

      {/* Shortcuts Help */}
      <ShortcutsHelp
        isOpen={isShortcutsHelpOpen}
        onClose={() => setIsShortcutsHelpOpen(false)}
      />

      {/* Peek Preview */}
      <PeekPreview
        task={selectedTask}
        isOpen={isPeekOpen}
        onClose={() => setIsPeekOpen(false)}
      />

      {/* Settings Panel */}
      <Settings
        isOpen={isSettingsOpen}
        onClose={() => setIsSettingsOpen(false)}
      />
    </Layout>
  )
}

/**
 * Main application component.
 *
 * Wraps AppContent in providers:
 * - ThemeProvider: Theme state management (light/dark/system)
 * - CurrentBoardProvider: Board selection state (must be outermost since AppContent uses it)
 * - SelectionProvider: Task selection state management
 */
export default function App() {
  return (
    <ThemeProvider>
      <CurrentBoardProvider>
        <SelectionProvider>
          <AppContent />
        </SelectionProvider>
      </CurrentBoardProvider>
    </ThemeProvider>
  )
}
