import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Header } from './Header'

describe('Header', () => {
  // Test 7.1: Renders app title 'EgenSkriven'
  it('renders app title', () => {
    render(<Header />)
    expect(screen.getByText('EgenSkriven')).toBeInTheDocument()
  })

  // Test 7.2: Displays keyboard shortcuts (C, Enter, Esc)
  describe('keyboard shortcuts', () => {
    it('displays C shortcut for Create', () => {
      render(<Header />)
      expect(screen.getByText('C')).toBeInTheDocument()
      expect(screen.getByText('Create')).toBeInTheDocument()
    })

    it('displays Enter shortcut for Open', () => {
      render(<Header />)
      expect(screen.getByText('Enter')).toBeInTheDocument()
      expect(screen.getByText('Open')).toBeInTheDocument()
    })

    it('displays Esc shortcut for Close', () => {
      render(<Header />)
      expect(screen.getByText('Esc')).toBeInTheDocument()
      expect(screen.getByText('Close')).toBeInTheDocument()
    })

    it('renders keyboard shortcuts in kbd elements', () => {
      render(<Header />)
      
      const kbdElements = document.querySelectorAll('kbd')
      expect(kbdElements.length).toBe(3)
      
      const kbdTexts = Array.from(kbdElements).map(el => el.textContent)
      expect(kbdTexts).toContain('C')
      expect(kbdTexts).toContain('Enter')
      expect(kbdTexts).toContain('Esc')
    })
  })

  it('renders as header element', () => {
    render(<Header />)
    expect(screen.getByRole('banner')).toBeInTheDocument()
  })
})
