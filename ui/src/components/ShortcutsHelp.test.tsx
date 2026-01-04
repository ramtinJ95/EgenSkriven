import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ShortcutsHelp } from './ShortcutsHelp'

describe('ShortcutsHelp', () => {
  const mockOnClose = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders nothing when isOpen is false', () => {
    render(<ShortcutsHelp isOpen={false} onClose={mockOnClose} />)

    expect(screen.queryByText('Keyboard Shortcuts')).not.toBeInTheDocument()
  })

  it('renders the modal when isOpen is true', () => {
    render(<ShortcutsHelp isOpen={true} onClose={mockOnClose} />)

    expect(screen.getByText('Keyboard Shortcuts')).toBeInTheDocument()
  })

  it('renders all shortcut groups', () => {
    render(<ShortcutsHelp isOpen={true} onClose={mockOnClose} />)

    // Check for all group titles
    expect(screen.getByText('Global')).toBeInTheDocument()
    expect(screen.getByText('Task Actions')).toBeInTheDocument()
    expect(screen.getByText('Task Properties')).toBeInTheDocument()
    expect(screen.getByText('Navigation')).toBeInTheDocument()
    expect(screen.getByText('Selection')).toBeInTheDocument()
  })

  it('renders shortcut descriptions', () => {
    render(<ShortcutsHelp isOpen={true} onClose={mockOnClose} />)

    // Check for some shortcut descriptions
    expect(screen.getByText('Command palette')).toBeInTheDocument()
    expect(screen.getByText('Create new task')).toBeInTheDocument()
    expect(screen.getByText('Set status')).toBeInTheDocument()
    expect(screen.getByText('Next task')).toBeInTheDocument()
    expect(screen.getByText('Toggle select task')).toBeInTheDocument()
  })

  it('displays correct key combinations', () => {
    render(<ShortcutsHelp isOpen={true} onClose={mockOnClose} />)

    // Check for some key combinations (formatted)
    // Note: formatKeyCombo formats these based on platform
    // We just check that some kbd elements exist
    const kbdElements = document.querySelectorAll('kbd')
    expect(kbdElements.length).toBeGreaterThan(0)
  })

  it('closes on overlay click', async () => {
    const user = userEvent.setup()

    render(<ShortcutsHelp isOpen={true} onClose={mockOnClose} />)

    // Click the overlay (the outer div)
    const overlay = document.querySelector('[class*="overlay"]')
    expect(overlay).toBeInTheDocument()
    if (overlay) {
      await user.click(overlay)
    }

    expect(mockOnClose).toHaveBeenCalled()
  })

  it('does not close when clicking inside the modal', async () => {
    const user = userEvent.setup()

    render(<ShortcutsHelp isOpen={true} onClose={mockOnClose} />)

    // Click inside the modal content
    const modalContent = screen.getByText('Keyboard Shortcuts')
    await user.click(modalContent)

    expect(mockOnClose).not.toHaveBeenCalled()
  })

  it('closes on close button click', async () => {
    const user = userEvent.setup()

    render(<ShortcutsHelp isOpen={true} onClose={mockOnClose} />)

    // Find the close button (it contains an SVG)
    const closeButton = document.querySelector('[class*="closeButton"]')
    expect(closeButton).toBeInTheDocument()
    if (closeButton) {
      await user.click(closeButton)
    }

    expect(mockOnClose).toHaveBeenCalled()
  })

  it('contains footer hint about Escape key', () => {
    render(<ShortcutsHelp isOpen={true} onClose={mockOnClose} />)

    // The hint text should mention pressing Esc to close
    expect(screen.getByText(/press/i)).toBeInTheDocument()
    // Find the kbd element in the footer
    const footer = document.querySelector('[class*="footer"]')
    expect(footer).toBeInTheDocument()
    expect(footer?.textContent).toContain('Esc')
  })

  describe('shortcut content', () => {
    it('contains Global shortcuts', () => {
      render(<ShortcutsHelp isOpen={true} onClose={mockOnClose} />)

      expect(screen.getByText('Command palette')).toBeInTheDocument()
      expect(screen.getByText('Quick search')).toBeInTheDocument()
      expect(screen.getByText('Toggle board/list view')).toBeInTheDocument()
      expect(screen.getByText('Toggle sidebar')).toBeInTheDocument()
      expect(screen.getByText('Show shortcuts help')).toBeInTheDocument()
    })

    it('contains Task Actions shortcuts', () => {
      render(<ShortcutsHelp isOpen={true} onClose={mockOnClose} />)

      expect(screen.getByText('Create new task')).toBeInTheDocument()
      expect(screen.getByText('Open selected task')).toBeInTheDocument()
      expect(screen.getByText('Peek preview')).toBeInTheDocument()
      expect(screen.getByText('Edit title')).toBeInTheDocument()
      expect(screen.getByText('Delete task')).toBeInTheDocument()
    })

    it('contains Task Properties shortcuts', () => {
      render(<ShortcutsHelp isOpen={true} onClose={mockOnClose} />)

      expect(screen.getByText('Set status')).toBeInTheDocument()
      expect(screen.getByText('Set priority')).toBeInTheDocument()
      expect(screen.getByText('Set type')).toBeInTheDocument()
      expect(screen.getByText('Manage labels')).toBeInTheDocument()
      expect(screen.getByText('Set due date')).toBeInTheDocument()
    })

    it('contains Navigation shortcuts', () => {
      render(<ShortcutsHelp isOpen={true} onClose={mockOnClose} />)

      expect(screen.getByText('Next task')).toBeInTheDocument()
      expect(screen.getByText('Previous task')).toBeInTheDocument()
      expect(screen.getByText('Previous column')).toBeInTheDocument()
      expect(screen.getByText('Next column')).toBeInTheDocument()
      expect(screen.getByText('Next task (arrow)')).toBeInTheDocument()
      expect(screen.getByText('Previous task (arrow)')).toBeInTheDocument()
      expect(screen.getByText('Close panel / deselect')).toBeInTheDocument()
    })

    it('contains Selection shortcuts', () => {
      render(<ShortcutsHelp isOpen={true} onClose={mockOnClose} />)

      expect(screen.getByText('Toggle select task')).toBeInTheDocument()
      expect(screen.getByText('Select range')).toBeInTheDocument()
      expect(screen.getByText('Select all visible')).toBeInTheDocument()
    })
  })
})
