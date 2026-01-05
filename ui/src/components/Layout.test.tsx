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

// Mock filter store
vi.mock('../stores/filters', () => ({
  useFilterStore: () => ({
    filters: [],
    matchMode: 'all',
    searchQuery: '',
    setSearchQuery: vi.fn(),
    clearFilters: vi.fn(),
    removeFilter: vi.fn(),
    setMatchMode: vi.fn(),
  }),
}))

// Helper to render Layout with required props
const renderLayout = (children: React.ReactNode) => {
  return render(
    <Layout
      totalTasks={10}
      filteredTasks={5}
      onOpenFilterBuilder={vi.fn()}
      onOpenDisplayOptions={vi.fn()}
    >
      {children}
    </Layout>
  )
}

describe('Layout', () => {
  // Test 6.1: Renders Header component
  it('renders Header component', () => {
    renderLayout(<div>Content</div>)
    // Both Header and Sidebar have "EgenSkriven" - get the first one
    expect(screen.getAllByText('EgenSkriven').length).toBeGreaterThanOrEqual(1)
  })

  it('renders header with display button', () => {
    renderLayout(<div>Content</div>)
    expect(screen.getByText('Display')).toBeInTheDocument()
  })

  // Test 6.2: Renders children in main area
  it('renders children in main area', () => {
    renderLayout(<div>Test Content</div>)
    expect(screen.getByText('Test Content')).toBeInTheDocument()
  })

  it('renders multiple children', () => {
    render(
      <Layout
        totalTasks={10}
        filteredTasks={5}
        onOpenFilterBuilder={vi.fn()}
        onOpenDisplayOptions={vi.fn()}
      >
        <div>First Child</div>
        <div>Second Child</div>
      </Layout>
    )
    expect(screen.getByText('First Child')).toBeInTheDocument()
    expect(screen.getByText('Second Child')).toBeInTheDocument()
  })

  it('renders complex children', () => {
    renderLayout(
      <article>
        <h1>Article Title</h1>
        <p>Article content</p>
      </article>
    )
    expect(screen.getByRole('heading', { name: 'Article Title' })).toBeInTheDocument()
    expect(screen.getByText('Article content')).toBeInTheDocument()
  })

  it('renders FilterBar with task counts', () => {
    renderLayout(<div>Content</div>)
    // FilterBar shows "5 of 10 tasks" when there are active filters
    expect(screen.getByText(/5 of 10/)).toBeInTheDocument()
  })
})
