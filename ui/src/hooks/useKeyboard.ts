import { useEffect, useRef } from 'react'

/**
 * Key combination descriptor.
 * Examples:
 *   { key: 'k', meta: true }        -> Cmd+K (Mac) / Ctrl+K (Windows)
 *   { key: 'Enter' }                -> Enter
 *   { key: 'Backspace', shift: true } -> Shift+Backspace
 */
export interface KeyCombo {
  key: string
  meta?: boolean // Cmd on Mac, Ctrl on Windows
  ctrl?: boolean // Ctrl specifically (rarely needed)
  alt?: boolean // Option on Mac, Alt on Windows
  shift?: boolean
}

export interface ShortcutHandler {
  // Key combination to trigger this shortcut
  combo: KeyCombo

  // Handler function - return true to prevent default behavior
  handler: (event: KeyboardEvent) => void | boolean

  // Optional description for help modal
  description?: string

  // If true, shortcut works even when focus is in an input
  allowInInput?: boolean

  // Optional: only active when this condition is true
  when?: () => boolean
}

/**
 * Check if an element is an input where shortcuts should be disabled.
 */
function isInputElement(element: Element | null): boolean {
  if (!element) return false

  const tagName = element.tagName.toLowerCase()

  // Standard input elements
  if (tagName === 'input' || tagName === 'textarea' || tagName === 'select') {
    return true
  }

  // Contenteditable elements
  if (element.getAttribute('contenteditable') === 'true') {
    return true
  }

  // Elements with role="textbox" (for rich text editors)
  if (element.getAttribute('role') === 'textbox') {
    return true
  }

  return false
}

/**
 * Check if a keyboard event matches a key combination.
 */
function matchesCombo(event: KeyboardEvent, combo: KeyCombo): boolean {
  // Normalize the key for comparison
  const eventKey = event.key.toLowerCase()
  const comboKey = combo.key.toLowerCase()

  // Check the key itself
  if (eventKey !== comboKey) {
    return false
  }

  // Check modifiers
  // Note: event.metaKey is Cmd on Mac, we also accept ctrlKey for cross-platform
  const metaOrCtrl = event.metaKey || event.ctrlKey

  if (combo.meta && !metaOrCtrl) return false
  if (!combo.meta && metaOrCtrl && !combo.ctrl) return false

  if (combo.shift && !event.shiftKey) return false
  if (!combo.shift && event.shiftKey) return false

  if (combo.alt && !event.altKey) return false
  if (!combo.alt && event.altKey) return false

  if (combo.ctrl && !event.ctrlKey) return false

  return true
}

/**
 * Hook for registering keyboard shortcuts.
 *
 * @example
 * // In a component
 * useKeyboardShortcuts([
 *   {
 *     combo: { key: 'k', meta: true },
 *     handler: () => openCommandPalette(),
 *     description: 'Open command palette',
 *   },
 *   {
 *     combo: { key: 'c' },
 *     handler: () => createTask(),
 *     description: 'Create new task',
 *   },
 * ]);
 */
export function useKeyboardShortcuts(shortcuts: ShortcutHandler[]): void {
  // Use ref to always have latest shortcuts without re-adding listener
  const shortcutsRef = useRef(shortcuts)
  shortcutsRef.current = shortcuts

  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      const activeElement = document.activeElement
      const inInput = isInputElement(activeElement)

      for (const shortcut of shortcutsRef.current) {
        // Skip if in input and shortcut doesn't allow it
        if (inInput && !shortcut.allowInInput) {
          continue
        }

        // Skip if condition not met
        if (shortcut.when && !shortcut.when()) {
          continue
        }

        // Check if key matches
        if (matchesCombo(event, shortcut.combo)) {
          const result = shortcut.handler(event)

          // Prevent default unless handler explicitly returns false
          if (result !== false) {
            event.preventDefault()
            event.stopPropagation()
          }

          // Only trigger first matching shortcut
          return
        }
      }
    }

    // Add listener with capture to handle before other listeners
    document.addEventListener('keydown', handleKeyDown, { capture: true })

    return () => {
      document.removeEventListener('keydown', handleKeyDown, { capture: true })
    }
  }, []) // Empty deps - we use ref for shortcuts
}

/**
 * Format a key combination for display in the UI.
 *
 * @example
 * formatKeyCombo({ key: 'k', meta: true }) // Returns "Cmd+K" on Mac, "Ctrl+K" on Windows
 * formatKeyCombo({ key: 'Enter', shift: true }) // Returns "Shift+Enter"
 */
export function formatKeyCombo(combo: KeyCombo): string {
  const parts: string[] = []

  // Detect platform for correct modifier display
  const isMac =
    typeof navigator !== 'undefined' &&
    navigator.platform.toLowerCase().includes('mac')

  if (combo.meta) {
    parts.push(isMac ? 'Cmd' : 'Ctrl')
  }
  if (combo.ctrl && !combo.meta) {
    parts.push('Ctrl')
  }
  if (combo.alt) {
    parts.push(isMac ? 'Option' : 'Alt')
  }
  if (combo.shift) {
    parts.push('Shift')
  }

  // Format key name
  let keyName = combo.key
  if (keyName === ' ') keyName = 'Space'
  if (keyName === 'Escape') keyName = 'Esc'
  if (keyName === 'Backspace') keyName = isMac ? 'Delete' : 'Backspace'
  if (keyName === 'ArrowUp') keyName = '↑'
  if (keyName === 'ArrowDown') keyName = '↓'
  if (keyName === 'ArrowLeft') keyName = '←'
  if (keyName === 'ArrowRight') keyName = '→'

  // Capitalize single letters
  if (keyName.length === 1) {
    keyName = keyName.toUpperCase()
  }

  parts.push(keyName)

  return parts.join('+')
}
