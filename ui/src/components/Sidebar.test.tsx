import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Sidebar } from './Sidebar'

// Mock boards data
const mockBoards = [
  { id: 'board-1', name: 'Work', prefix: 'WRK', columns: [], color: '#3B82F6' },
  { id: 'board-2', name: 'Personal', prefix: 'PER', columns: [], color: '#22C55E' },
]

// Mock state
let mockBoardsLoading = false
let mockCurrentBoard = mockBoards[0]
const mockSetCurrentBoard = vi.fn()
const mockCreateBoard = vi.fn()

// Mock useBoards hook
vi.mock('../hooks/useBoards', () => ({
  useBoards: () => ({
    boards: mockBoards,
    loading: mockBoardsLoading,
    error: null,
    createBoard: mockCreateBoard,
    deleteBoard: vi.fn(),
  }),
}))

// Mock useCurrentBoard hook
vi.mock('../hooks/useCurrentBoard', () => ({
  useCurrentBoard: () => ({
    currentBoard: mockCurrentBoard,
    setCurrentBoard: mockSetCurrentBoard,
    loading: false,
  }),
}))

describe('Sidebar', () => {
  const mockOnToggle = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    mockBoardsLoading = false
    mockCurrentBoard = mockBoards[0]
  })

  describe('expanded state', () => {
    it('renders title', () => {
      render(<Sidebar collapsed={false} onToggle={mockOnToggle} />)
      expect(screen.getByText('EgenSkriven')).toBeInTheDocument()
    })

    it('renders BOARDS section title', () => {
      render(<Sidebar collapsed={false} onToggle={mockOnToggle} />)
      expect(screen.getByText('BOARDS')).toBeInTheDocument()
    })

    it('renders all boards', () => {
      render(<Sidebar collapsed={false} onToggle={mockOnToggle} />)
      expect(screen.getByText('Work')).toBeInTheDocument()
      expect(screen.getByText('(WRK)')).toBeInTheDocument()
      expect(screen.getByText('Personal')).toBeInTheDocument()
      expect(screen.getByText('(PER)')).toBeInTheDocument()
    })

    it('renders "New board" button', () => {
      render(<Sidebar collapsed={false} onToggle={mockOnToggle} />)
      expect(screen.getByText('+ New board')).toBeInTheDocument()
    })

    it('renders collapse button with correct aria-label', () => {
      render(<Sidebar collapsed={false} onToggle={mockOnToggle} />)
      expect(screen.getByLabelText('Collapse sidebar')).toBeInTheDocument()
    })

    it('calls onToggle when collapse button is clicked', () => {
      render(<Sidebar collapsed={false} onToggle={mockOnToggle} />)
      fireEvent.click(screen.getByLabelText('Collapse sidebar'))
      expect(mockOnToggle).toHaveBeenCalledTimes(1)
    })
  })

  describe('collapsed state', () => {
    it('only renders expand button', () => {
      render(<Sidebar collapsed={true} onToggle={mockOnToggle} />)
      expect(screen.queryByText('EgenSkriven')).not.toBeInTheDocument()
      expect(screen.queryByText('BOARDS')).not.toBeInTheDocument()
      expect(screen.getByLabelText('Expand sidebar')).toBeInTheDocument()
    })

    it('calls onToggle when expand button is clicked', () => {
      render(<Sidebar collapsed={true} onToggle={mockOnToggle} />)
      fireEvent.click(screen.getByLabelText('Expand sidebar'))
      expect(mockOnToggle).toHaveBeenCalledTimes(1)
    })
  })

  describe('loading state', () => {
    it('shows loading message when boards are loading', () => {
      mockBoardsLoading = true
      render(<Sidebar collapsed={false} onToggle={mockOnToggle} />)
      expect(screen.getByText('Loading...')).toBeInTheDocument()
    })
  })

  describe('board selection', () => {
    it('highlights current board', () => {
      render(<Sidebar collapsed={false} onToggle={mockOnToggle} />)
      const workButton = screen.getByRole('button', { name: /Work.*WRK/ })
      // CSS modules add hash to class names, so we check for partial match
      expect(workButton.className).toMatch(/active/)
    })

    it('calls setCurrentBoard when a board is clicked', () => {
      render(<Sidebar collapsed={false} onToggle={mockOnToggle} />)
      const personalButton = screen.getByRole('button', { name: /Personal.*PER/ })
      fireEvent.click(personalButton)
      expect(mockSetCurrentBoard).toHaveBeenCalledWith(mockBoards[1])
    })
  })

  describe('new board modal', () => {
    it('opens modal when "New board" button is clicked', async () => {
      render(<Sidebar collapsed={false} onToggle={mockOnToggle} />)
      fireEvent.click(screen.getByText('+ New board'))
      expect(screen.getByText('Create New Board')).toBeInTheDocument()
    })

    it('closes modal when Cancel is clicked', async () => {
      render(<Sidebar collapsed={false} onToggle={mockOnToggle} />)
      fireEvent.click(screen.getByText('+ New board'))
      fireEvent.click(screen.getByText('Cancel'))
      expect(screen.queryByText('Create New Board')).not.toBeInTheDocument()
    })

    it('validates empty name', async () => {
      const user = userEvent.setup()
      render(<Sidebar collapsed={false} onToggle={mockOnToggle} />)
      fireEvent.click(screen.getByText('+ New board'))

      await user.type(screen.getByLabelText('Prefix'), 'TST')
      fireEvent.click(screen.getByText('Create Board'))

      expect(screen.getByText('Name is required')).toBeInTheDocument()
    })

    it('validates empty prefix', async () => {
      const user = userEvent.setup()
      render(<Sidebar collapsed={false} onToggle={mockOnToggle} />)
      fireEvent.click(screen.getByText('+ New board'))

      await user.type(screen.getByLabelText('Name'), 'Test Board')
      fireEvent.click(screen.getByText('Create Board'))

      expect(screen.getByText('Prefix is required')).toBeInTheDocument()
    })

    it('validates non-alphanumeric prefix', async () => {
      const user = userEvent.setup()
      render(<Sidebar collapsed={false} onToggle={mockOnToggle} />)
      fireEvent.click(screen.getByText('+ New board'))

      await user.type(screen.getByLabelText('Name'), 'Test Board')
      // Type special characters (they'll be uppercased but still invalid)
      await user.type(screen.getByLabelText('Prefix'), 'TST@#')
      fireEvent.click(screen.getByText('Create Board'))

      expect(screen.getByText('Prefix must be alphanumeric')).toBeInTheDocument()
    })

    it('calls createBoard with correct data', async () => {
      const user = userEvent.setup()
      mockCreateBoard.mockResolvedValue({ id: 'new-board', name: 'Test', prefix: 'TST' })

      render(<Sidebar collapsed={false} onToggle={mockOnToggle} />)
      fireEvent.click(screen.getByText('+ New board'))

      await user.type(screen.getByLabelText('Name'), 'Test Board')
      await user.type(screen.getByLabelText('Prefix'), 'tst') // lowercase should be uppercased

      fireEvent.click(screen.getByText('Create Board'))

      await waitFor(() => {
        expect(mockCreateBoard).toHaveBeenCalledWith({
          name: 'Test Board',
          prefix: 'TST',
          color: expect.any(String),
        })
      })
    })

    it('shows preview of display ID format', async () => {
      const user = userEvent.setup()
      render(<Sidebar collapsed={false} onToggle={mockOnToggle} />)
      fireEvent.click(screen.getByText('+ New board'))

      await user.type(screen.getByLabelText('Prefix'), 'WRK')

      expect(screen.getByText(/WRK-123/)).toBeInTheDocument()
    })

    it('uppercase prefix input as user types', async () => {
      const user = userEvent.setup()
      render(<Sidebar collapsed={false} onToggle={mockOnToggle} />)
      fireEvent.click(screen.getByText('+ New board'))

      const prefixInput = screen.getByLabelText('Prefix') as HTMLInputElement
      await user.type(prefixInput, 'abc')

      expect(prefixInput.value).toBe('ABC')
    })
  })
})
