import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Header } from './Header'

// Mock filter store - Zustand stores use selector pattern
const mockFilterState = {
  searchQuery: '',
  setSearchQuery: vi.fn(),
}
vi.mock('../stores/filters', () => ({
  useFilterStore: (selector: (state: typeof mockFilterState) => unknown) => selector(mockFilterState),
}))

describe('Header', () => {
  // Test 7.1: Renders app title 'EgenSkriven'
  it('renders app title', () => {
    render(<Header />)
    expect(screen.getByText('EgenSkriven')).toBeInTheDocument()
  })

  // Test 7.2: Displays Display button and Help shortcut
  describe('header elements', () => {
    it('displays Display button', () => {
      render(<Header onDisplayOptionsClick={vi.fn()} />)
      expect(screen.getByText('Display')).toBeInTheDocument()
    })

    it('displays Help shortcut', () => {
      render(<Header />)
      expect(screen.getByText('?')).toBeInTheDocument()
      expect(screen.getByText('Help')).toBeInTheDocument()
    })

    it('renders search bar placeholder', () => {
      render(<Header />)
      expect(screen.getByPlaceholderText(/search tasks/i)).toBeInTheDocument()
    })
  })

  it('renders as header element', () => {
    render(<Header />)
    expect(screen.getByRole('banner')).toBeInTheDocument()
  })

  it('calls onDisplayOptionsClick when Display button is clicked', () => {
    const onDisplayOptionsClick = vi.fn()
    render(<Header onDisplayOptionsClick={onDisplayOptionsClick} />)
    
    const displayButton = screen.getByRole('button', { name: /display/i })
    displayButton.click()
    
    expect(onDisplayOptionsClick).toHaveBeenCalled()
  })
})
