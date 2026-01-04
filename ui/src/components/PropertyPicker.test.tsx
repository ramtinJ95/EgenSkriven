import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import {
  PropertyPicker,
  STATUS_OPTIONS,
  PRIORITY_OPTIONS,
  TYPE_OPTIONS,
  type PropertyOption,
} from './PropertyPicker'

// Test options
const testOptions: PropertyOption<string>[] = [
  { value: 'option1', label: 'First Option', icon: '1' },
  { value: 'option2', label: 'Second Option', icon: '2' },
  { value: 'option3', label: 'Third Option', icon: '3' },
  { value: 'option4', label: 'Fourth Option', icon: '4', color: 'red' },
]

describe('PropertyPicker', () => {
  const mockOnClose = vi.fn()
  const mockOnSelect = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders nothing when isOpen is false', () => {
    render(
      <PropertyPicker
        isOpen={false}
        onClose={mockOnClose}
        onSelect={mockOnSelect}
        options={testOptions}
        title="Select Option"
      />
    )

    expect(screen.queryByText('Select Option')).not.toBeInTheDocument()
    expect(screen.queryByPlaceholderText('Filter...')).not.toBeInTheDocument()
  })

  it('renders options when isOpen is true', () => {
    render(
      <PropertyPicker
        isOpen={true}
        onClose={mockOnClose}
        onSelect={mockOnSelect}
        options={testOptions}
        title="Select Option"
      />
    )

    expect(screen.getByText('Select Option')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Filter...')).toBeInTheDocument()
    expect(screen.getByText('First Option')).toBeInTheDocument()
    expect(screen.getByText('Second Option')).toBeInTheDocument()
    expect(screen.getByText('Third Option')).toBeInTheDocument()
    expect(screen.getByText('Fourth Option')).toBeInTheDocument()
  })

  it('filters options by query input', async () => {
    const user = userEvent.setup()

    render(
      <PropertyPicker
        isOpen={true}
        onClose={mockOnClose}
        onSelect={mockOnSelect}
        options={testOptions}
        title="Select Option"
      />
    )

    const input = screen.getByPlaceholderText('Filter...')
    await user.type(input, 'first')

    expect(screen.getByText('First Option')).toBeInTheDocument()
    expect(screen.queryByText('Second Option')).not.toBeInTheDocument()
    expect(screen.queryByText('Third Option')).not.toBeInTheDocument()
    expect(screen.queryByText('Fourth Option')).not.toBeInTheDocument()
  })

  it('shows "No options found" when no matches', async () => {
    const user = userEvent.setup()

    render(
      <PropertyPicker
        isOpen={true}
        onClose={mockOnClose}
        onSelect={mockOnSelect}
        options={testOptions}
        title="Select Option"
      />
    )

    const input = screen.getByPlaceholderText('Filter...')
    await user.type(input, 'nonexistent')

    expect(screen.getByText('No options found')).toBeInTheDocument()
  })

  it('navigates with arrow keys', async () => {
    const user = userEvent.setup()

    render(
      <PropertyPicker
        isOpen={true}
        onClose={mockOnClose}
        onSelect={mockOnSelect}
        options={testOptions}
        title="Select Option"
      />
    )

    const input = screen.getByPlaceholderText('Filter...')

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

  it('selects option on Enter', async () => {
    const user = userEvent.setup()

    render(
      <PropertyPicker
        isOpen={true}
        onClose={mockOnClose}
        onSelect={mockOnSelect}
        options={testOptions}
        title="Select Option"
      />
    )

    const input = screen.getByPlaceholderText('Filter...')

    // Press Enter to select the first (selected) option
    await user.type(input, '{Enter}')

    expect(mockOnSelect).toHaveBeenCalledWith('option1')
    expect(mockOnClose).toHaveBeenCalled()
  })

  it('selects option on click', async () => {
    const user = userEvent.setup()

    render(
      <PropertyPicker
        isOpen={true}
        onClose={mockOnClose}
        onSelect={mockOnSelect}
        options={testOptions}
        title="Select Option"
      />
    )

    const secondOption = screen.getByText('Second Option')
    await user.click(secondOption)

    expect(mockOnSelect).toHaveBeenCalledWith('option2')
    expect(mockOnClose).toHaveBeenCalled()
  })

  it('closes on Escape key', async () => {
    const user = userEvent.setup()

    render(
      <PropertyPicker
        isOpen={true}
        onClose={mockOnClose}
        onSelect={mockOnSelect}
        options={testOptions}
        title="Select Option"
      />
    )

    const input = screen.getByPlaceholderText('Filter...')
    await user.type(input, '{Escape}')

    expect(mockOnClose).toHaveBeenCalled()
    expect(mockOnSelect).not.toHaveBeenCalled()
  })

  it('closes on overlay click', async () => {
    const user = userEvent.setup()

    render(
      <PropertyPicker
        isOpen={true}
        onClose={mockOnClose}
        onSelect={mockOnSelect}
        options={testOptions}
        title="Select Option"
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

  it('shows checkmark for current value', () => {
    render(
      <PropertyPicker
        isOpen={true}
        onClose={mockOnClose}
        onSelect={mockOnSelect}
        options={testOptions}
        currentValue="option2"
        title="Select Option"
      />
    )

    // The checkmark should appear next to the current value
    expect(screen.getByText('✓')).toBeInTheDocument()
  })

  it('displays option icons', () => {
    render(
      <PropertyPicker
        isOpen={true}
        onClose={mockOnClose}
        onSelect={mockOnSelect}
        options={testOptions}
        title="Select Option"
      />
    )

    expect(screen.getByText('1')).toBeInTheDocument()
    expect(screen.getByText('2')).toBeInTheDocument()
    expect(screen.getByText('3')).toBeInTheDocument()
    expect(screen.getByText('4')).toBeInTheDocument()
  })

  it('resets query when reopened', async () => {
    const user = userEvent.setup()

    const { rerender } = render(
      <PropertyPicker
        isOpen={true}
        onClose={mockOnClose}
        onSelect={mockOnSelect}
        options={testOptions}
        title="Select Option"
      />
    )

    // Type something
    const input = screen.getByPlaceholderText('Filter...')
    await user.type(input, 'test')

    // Close picker
    rerender(
      <PropertyPicker
        isOpen={false}
        onClose={mockOnClose}
        onSelect={mockOnSelect}
        options={testOptions}
        title="Select Option"
      />
    )

    // Reopen picker
    rerender(
      <PropertyPicker
        isOpen={true}
        onClose={mockOnClose}
        onSelect={mockOnSelect}
        options={testOptions}
        title="Select Option"
      />
    )

    // Query should be reset
    const newInput = screen.getByPlaceholderText('Filter...')
    expect(newInput).toHaveValue('')
  })

  it('sets initial selection to current value on open', () => {
    render(
      <PropertyPicker
        isOpen={true}
        onClose={mockOnClose}
        onSelect={mockOnSelect}
        options={testOptions}
        currentValue="option3"
        title="Select Option"
      />
    )

    // The third option (option3) should be selected
    const options = document.querySelectorAll('[class*="option"]')
    const selectedOption = document.querySelector('[class*="option"][class*="selected"]')
    expect(selectedOption).toBeInTheDocument()
    // The selected option should contain "Third Option"
    expect(selectedOption?.textContent).toContain('Third Option')
  })

  describe('pre-configured option sets', () => {
    it('exports STATUS_OPTIONS with correct values', () => {
      expect(STATUS_OPTIONS).toHaveLength(5)
      expect(STATUS_OPTIONS.map((o) => o.value)).toEqual([
        'backlog',
        'todo',
        'in_progress',
        'review',
        'done',
      ])
    })

    it('exports PRIORITY_OPTIONS with correct values', () => {
      expect(PRIORITY_OPTIONS).toHaveLength(4)
      expect(PRIORITY_OPTIONS.map((o) => o.value)).toEqual([
        'urgent',
        'high',
        'medium',
        'low',
      ])
    })

    it('exports TYPE_OPTIONS with correct values', () => {
      expect(TYPE_OPTIONS).toHaveLength(3)
      expect(TYPE_OPTIONS.map((o) => o.value)).toEqual([
        'bug',
        'feature',
        'chore',
      ])
    })
  })
})
