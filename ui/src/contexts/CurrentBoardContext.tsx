import { useState, useEffect, useCallback, createContext, useContext, type ReactNode } from 'react'
import type { Board } from '../types/board'
import { useBoards } from '../hooks/useBoards'

const STORAGE_KEY = 'egenskriven-current-board'

interface CurrentBoardContextValue {
  currentBoard: Board | null
  setCurrentBoard: (board: Board) => void
  loading: boolean
  /** All boards - exposed to avoid duplicate useBoards() subscriptions */
  boards: Board[]
  /** Error from loading boards */
  boardsError: Error | null
  /** Create a new board */
  createBoard: (input: { name: string; prefix: string; color?: string; columns?: string[] }) => Promise<Board>
  /** Delete a board */
  deleteBoard: (id: string) => Promise<void>
}

const CurrentBoardContext = createContext<CurrentBoardContextValue | null>(null)

interface CurrentBoardProviderProps {
  children: ReactNode
}

/**
 * Provider component for current board state.
 * 
 * Wraps the application to provide current board context.
 * Persists the selected board to localStorage.
 * 
 * @example
 * ```tsx
 * function App() {
 *   return (
 *     <CurrentBoardProvider>
 *       <Layout />
 *     </CurrentBoardProvider>
 *   )
 * }
 * ```
 */
export function CurrentBoardProvider({ children }: CurrentBoardProviderProps) {
  const { boards, loading: boardsLoading, error: boardsError, createBoard, deleteBoard } = useBoards()
  const [currentBoard, setCurrentBoardState] = useState<Board | null>(null)
  const [initialized, setInitialized] = useState(false)

  // Initialize current board from localStorage or first board
  useEffect(() => {
    if (boardsLoading) return

    // Already initialized
    if (initialized) return

    // Handle case where there are no boards
    if (boards.length === 0) {
      setCurrentBoardState(null)
      setInitialized(true)
      return
    }

    const savedBoardId = localStorage.getItem(STORAGE_KEY)
    
    if (savedBoardId) {
      // Try to find the saved board
      const savedBoard = boards.find((b) => b.id === savedBoardId)
      if (savedBoard) {
        setCurrentBoardState(savedBoard)
        setInitialized(true)
        return
      }
      // Saved board not found - clear stale localStorage
      localStorage.removeItem(STORAGE_KEY)
    }

    // Default to first board
    setCurrentBoardState(boards[0])
    localStorage.setItem(STORAGE_KEY, boards[0].id)
    setInitialized(true)
  }, [boards, boardsLoading, initialized])

  // Keep current board in sync if it's updated
  useEffect(() => {
    if (!currentBoard) return

    const updatedBoard = boards.find((b) => b.id === currentBoard.id)
    if (updatedBoard && updatedBoard !== currentBoard) {
      setCurrentBoardState(updatedBoard)
    }
  }, [boards, currentBoard])

  // Handle board deletion - switch to another board or set to null
  useEffect(() => {
    if (!currentBoard) return

    const stillExists = boards.some((b) => b.id === currentBoard.id)
    if (!stillExists) {
      if (boards.length > 0) {
        // Current board was deleted, switch to first available
        setCurrentBoardState(boards[0])
        localStorage.setItem(STORAGE_KEY, boards[0].id)
      } else {
        // All boards deleted, set to null
        setCurrentBoardState(null)
        localStorage.removeItem(STORAGE_KEY)
      }
    }
  }, [boards, currentBoard])

  const setCurrentBoard = useCallback((board: Board) => {
    setCurrentBoardState(board)
    localStorage.setItem(STORAGE_KEY, board.id)
  }, [])

  const value: CurrentBoardContextValue = {
    currentBoard,
    setCurrentBoard,
    loading: boardsLoading || !initialized,
    boards,
    boardsError,
    createBoard,
    deleteBoard,
  }

  return (
    <CurrentBoardContext.Provider value={value}>
      {children}
    </CurrentBoardContext.Provider>
  )
}

/**
 * Hook to access and set the current board.
 * 
 * Must be used within a CurrentBoardProvider.
 * 
 * Features:
 * - Returns the currently selected board
 * - Provides setCurrentBoard to change the active board
 * - Persists selection to localStorage
 * - Automatically syncs with board updates/deletions
 * 
 * @example
 * ```tsx
 * function BoardHeader() {
 *   const { currentBoard, setCurrentBoard } = useCurrentBoard()
 *   
 *   return (
 *     <div>
 *       <h1>{currentBoard?.name}</h1>
 *       <span>{currentBoard?.prefix}</span>
 *     </div>
 *   )
 * }
 * ```
 */
export function useCurrentBoard(): CurrentBoardContextValue {
  const context = useContext(CurrentBoardContext)
  
  if (!context) {
    throw new Error('useCurrentBoard must be used within a CurrentBoardProvider')
  }
  
  return context
}
