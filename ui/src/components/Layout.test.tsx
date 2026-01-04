import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Layout } from './Layout'

// Mock boards data
const mockBoards = [
  { id: 'board-1', name: 'Work', prefix: 'WRK', columns: [], color: '#3B82F6', collectionId: 'boards', collectionName: 'boards' },
]

// Mock the hooks used by Layout and Sidebar
vi.mock('../hooks/useCurrentBoard', () => ({
  CurrentBoardProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  useCurrentBoard: () => ({
    currentBoard: mockBoards[0],
    setCurrentBoard: vi.fn(),
    loading: false,
    boards: mockBoards,
    boardsError: null,
    createBoard: vi.fn(),
    deleteBoard: vi.fn(),
  }),
}))

describe('Layout', () => {
  // Test 6.1: Renders Header component
  it('renders Header component', () => {
    render(<Layout><div>Content</div></Layout>)
    // Both Header and Sidebar have "EgenSkriven" - get the first one
    expect(screen.getAllByText('EgenSkriven').length).toBeGreaterThanOrEqual(1)
  })

  it('renders header with keyboard shortcuts', () => {
    render(<Layout><div>Content</div></Layout>)
    expect(screen.getByText('Create')).toBeInTheDocument()
    expect(screen.getByText('Open')).toBeInTheDocument()
    expect(screen.getByText('Close')).toBeInTheDocument()
  })

  // Test 6.2: Renders children in main area
  it('renders children in main area', () => {
    render(<Layout><div>Test Content</div></Layout>)
    expect(screen.getByText('Test Content')).toBeInTheDocument()
  })

  it('renders multiple children', () => {
    render(
      <Layout>
        <div>First Child</div>
        <div>Second Child</div>
      </Layout>
    )
    expect(screen.getByText('First Child')).toBeInTheDocument()
    expect(screen.getByText('Second Child')).toBeInTheDocument()
  })

  it('renders complex children', () => {
    render(
      <Layout>
        <article>
          <h1>Article Title</h1>
          <p>Article content</p>
        </article>
      </Layout>
    )
    expect(screen.getByRole('heading', { name: 'Article Title' })).toBeInTheDocument()
    expect(screen.getByText('Article content')).toBeInTheDocument()
  })
})
