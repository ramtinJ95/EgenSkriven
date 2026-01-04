import { describe, it, expect, vi, beforeEach, beforeAll } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { CommandPalette, type Command } from './CommandPalette'

// Mock scrollIntoView which is not implemented in jsdom
beforeAll(() => {
  Element.prototype.scrollIntoView = vi.fn()
})

// Mock commands for testing
const mockCommands: Command[] = [
  {
    id: 'create-task',
    label: 'Create Task',
    section: 'actions',
    icon: '+',
    shortcut: { key: 'c' },
    action: vi.fn(),
  },
  {
    id: 'delete-task',
    label: 'Delete Task',
    section: 'actions',
    icon: '×',
    shortcut: { key: 'Backspace' },
    action: vi.fn(),
  },
  {
    id: 'go-backlog',
    label: 'Go to Backlog',
    section: 'navigation',
    icon: '→',
    action: vi.fn(),
  },
  {
    id: 'recent-1',
    label: 'Fix login bug',
    section: 'recent',
    action: vi.fn(),
  },
  {
    id: 'conditional-cmd',
    label: 'Conditional Command',
    section: 'actions',
    action: vi.fn(),
    when: () => false, // This command should not appear
  },
]

describe('CommandPalette', () => {
  const mockOnClose = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    // Reset all command action mocks
    mockCommands.forEach((cmd) => {
      if (typeof cmd.action === 'function') {
        vi.mocked(cmd.action).mockClear()
      }
    })
  })

  it('renders nothing when isOpen is false', () => {
    render(
      <CommandPalette
        isOpen={false}
        onClose={mockOnClose}
        commands={mockCommands}
      />
    )

    expect(screen.queryByPlaceholderText(/type a command/i)).not.toBeInTheDocument()
  })

  it('renders command list when isOpen is true', () => {
    render(
      <CommandPalette
        isOpen={true}
        onClose={mockOnClose}
        commands={mockCommands}
      />
    )

    expect(screen.getByPlaceholderText(/type a command/i)).toBeInTheDocument()
    expect(screen.getByText('Create Task')).toBeInTheDocument()
    expect(screen.getByText('Delete Task')).toBeInTheDocument()
    expect(screen.getByText('Go to Backlog')).toBeInTheDocument()
    expect(screen.getByText('Fix login bug')).toBeInTheDocument()
  })

  it('groups commands by section', () => {
    render(
      <CommandPalette
        isOpen={true}
        onClose={mockOnClose}
        commands={mockCommands}
      />
    )

    expect(screen.getByText('ACTIONS')).toBeInTheDocument()
    expect(screen.getByText('NAVIGATION')).toBeInTheDocument()
    expect(screen.getByText('RECENT TASKS')).toBeInTheDocument()
  })

  it('filters commands by search query (fuzzy matching)', async () => {
    const user = userEvent.setup()

    render(
      <CommandPalette
        isOpen={true}
        onClose={mockOnClose}
        commands={mockCommands}
      />
    )

    const input = screen.getByPlaceholderText(/type a command/i)
    await user.type(input, 'crt')

    // "Create Task" should match "crt" (fuzzy)
    expect(screen.getByText('Create Task')).toBeInTheDocument()
    // "Delete Task" should not match "crt"
    expect(screen.queryByText('Delete Task')).not.toBeInTheDocument()
  })

  it('shows "No commands found" when no matches', async () => {
    const user = userEvent.setup()

    render(
      <CommandPalette
        isOpen={true}
        onClose={mockOnClose}
        commands={mockCommands}
      />
    )

    const input = screen.getByPlaceholderText(/type a command/i)
    await user.type(input, 'xyznonexistent')

    expect(screen.getByText('No commands found')).toBeInTheDocument()
  })

  it('executes command on click', async () => {
    const user = userEvent.setup()
    vi.useFakeTimers({ shouldAdvanceTime: true })

    render(
      <CommandPalette
        isOpen={true}
        onClose={mockOnClose}
        commands={mockCommands}
      />
    )

    const createTaskButton = screen.getByText('Create Task')
    await user.click(createTaskButton)

    expect(mockOnClose).toHaveBeenCalled()

    // Action is executed after a timeout
    vi.advanceTimersByTime(100)
    expect(mockCommands[0].action).toHaveBeenCalled()

    vi.useRealTimers()
  })

  it('closes on Escape key', async () => {
    const user = userEvent.setup()

    render(
      <CommandPalette
        isOpen={true}
        onClose={mockOnClose}
        commands={mockCommands}
      />
    )

    const input = screen.getByPlaceholderText(/type a command/i)
    await user.type(input, '{Escape}')

    expect(mockOnClose).toHaveBeenCalled()
  })

  it('closes on overlay click', async () => {
    const user = userEvent.setup()

    render(
      <CommandPalette
        isOpen={true}
        onClose={mockOnClose}
        commands={mockCommands}
      />
    )

    // Click the overlay (the outer div)
    const overlay = document.querySelector('[class*="overlay"]')
    expect(overlay).toBeInTheDocument()
    if (overlay) {
      await user.click(overlay)
    }

    expect(mockOnClose).toHaveBeenCalled()
  })

  it('navigates with arrow keys', async () => {
    const user = userEvent.setup()

    render(
      <CommandPalette
        isOpen={true}
        onClose={mockOnClose}
        commands={mockCommands}
      />
    )

    const input = screen.getByPlaceholderText(/type a command/i)

    // First item should be selected by default
    let selectedItems = document.querySelectorAll('[class*="selected"]')
    expect(selectedItems.length).toBeGreaterThan(0)

    // Press ArrowDown to move selection
    await user.type(input, '{ArrowDown}')

    // Press ArrowUp to move back
    await user.type(input, '{ArrowUp}')

    // Selection should still work (no errors)
    selectedItems = document.querySelectorAll('[class*="selected"]')
    expect(selectedItems.length).toBeGreaterThan(0)
  })

  it('executes selected command on Enter', async () => {
    const user = userEvent.setup()
    vi.useFakeTimers({ shouldAdvanceTime: true })

    render(
      <CommandPalette
        isOpen={true}
        onClose={mockOnClose}
        commands={mockCommands}
      />
    )

    const input = screen.getByPlaceholderText(/type a command/i)

    // Press Enter to execute the first (selected) command
    await user.type(input, '{Enter}')

    expect(mockOnClose).toHaveBeenCalled()

    // Action is executed after a timeout
    vi.advanceTimersByTime(100)
    expect(mockCommands[0].action).toHaveBeenCalled()

    vi.useRealTimers()
  })

  it('respects when condition for conditional commands', () => {
    render(
      <CommandPalette
        isOpen={true}
        onClose={mockOnClose}
        commands={mockCommands}
      />
    )

    // The conditional command (when: () => false) should not appear
    expect(screen.queryByText('Conditional Command')).not.toBeInTheDocument()
  })

  it('shows conditional command when condition is true', () => {
    const commandsWithVisibleConditional = mockCommands.map((cmd) =>
      cmd.id === 'conditional-cmd' ? { ...cmd, when: () => true } : cmd
    )

    render(
      <CommandPalette
        isOpen={true}
        onClose={mockOnClose}
        commands={commandsWithVisibleConditional}
      />
    )

    expect(screen.getByText('Conditional Command')).toBeInTheDocument()
  })

  it('displays keyboard shortcuts', () => {
    render(
      <CommandPalette
        isOpen={true}
        onClose={mockOnClose}
        commands={mockCommands}
      />
    )

    // Check that shortcut keys are displayed
    expect(screen.getByText('C')).toBeInTheDocument()
  })

  it('resets query and selection when reopened', async () => {
    const user = userEvent.setup()

    const { rerender } = render(
      <CommandPalette
        isOpen={true}
        onClose={mockOnClose}
        commands={mockCommands}
      />
    )

    // Type something
    const input = screen.getByPlaceholderText(/type a command/i)
    await user.type(input, 'test')

    // Close palette
    rerender(
      <CommandPalette
        isOpen={false}
        onClose={mockOnClose}
        commands={mockCommands}
      />
    )

    // Reopen palette
    rerender(
      <CommandPalette
        isOpen={true}
        onClose={mockOnClose}
        commands={mockCommands}
      />
    )

    // Query should be reset
    const newInput = screen.getByPlaceholderText(/type a command/i)
    expect(newInput).toHaveValue('')
  })
})
