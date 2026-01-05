import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Layout } from './Layout'

// Mock boards data
const mockBoards = [
  { id: 'board-1', name: 'Work', prefix: 'WRK', columns: [], color: '#3B82F6', collectionId: 'boards', collectionName: 'boards' },
]

// Mock the hooks used by Layout and Sidebar (imported via ../contexts which re-exports from CurrentBoardContext)
vi.mock('../contexts/CurrentBoardContext', () => ({
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

// Mock ThemeContext to avoid matchMedia issues
vi.mock('../contexts/ThemeContext', () => ({
  ThemeProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  useTheme: () => ({
    themeName: 'dark',
    setThemeName: vi.fn(),
    resolvedTheme: {
      name: 'Dark',
      appearance: 'dark' as const,
      colors: {},
    },
    activeTheme: {
      id: 'dark',
      name: 'Dark',
      appearance: 'dark' as const,
      colors: { accent: '#5E6AD2' },
    },
    allThemes: [],
    customThemes: [],
    importTheme: vi.fn(),
    removeCustomTheme: vi.fn(),
    darkTheme: 'dark',
    setDarkTheme: vi.fn(),
    lightTheme: 'light',
    setLightTheme: vi.fn(),
  }),
}))

// Mock filter store - Zustand stores use selector pattern
const mockFilterState = {
  filters: [],
  matchMode: 'all',
  searchQuery: '',
  setSearchQuery: vi.fn(),
  clearFilters: vi.fn(),
  removeFilter: vi.fn(),
  setMatchMode: vi.fn(),
}
vi.mock('../stores/filters', () => ({
  useFilterStore: (selector: (state: typeof mockFilterState) => unknown) => selector(mockFilterState),
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

  it('renders header with search bar', () => {
    renderLayout(<div>Content</div>)
    expect(screen.getByPlaceholderText(/search tasks/i)).toBeInTheDocument()
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
    // FilterBar shows total task count when no filters active
    expect(screen.getByText('10 tasks')).toBeInTheDocument()
  })
})
