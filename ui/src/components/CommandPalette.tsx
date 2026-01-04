import { useState, useEffect, useMemo, useCallback, useRef } from 'react'
import { createPortal } from 'react-dom'
import { formatKeyCombo, type KeyCombo } from '../hooks/useKeyboard'
import styles from './CommandPalette.module.css'

export interface Command {
  id: string
  label: string
  shortcut?: KeyCombo
  section: 'actions' | 'navigation' | 'recent'
  icon?: string
  action: () => void
  // Optional: only show when condition is true
  when?: () => boolean
}

interface CommandPaletteProps {
  isOpen: boolean
  onClose: () => void
  commands: Command[]
}

/**
 * Fuzzy match a query against a string.
 * Returns true if all characters in query appear in order in string.
 */
function fuzzyMatch(query: string, text: string): boolean {
  const queryLower = query.toLowerCase()
  const textLower = text.toLowerCase()

  let queryIndex = 0

  for (let i = 0; i < textLower.length && queryIndex < queryLower.length; i++) {
    if (textLower[i] === queryLower[queryIndex]) {
      queryIndex++
    }
  }

  return queryIndex === queryLower.length
}

/**
 * Score a fuzzy match - higher is better.
 * Prefers matches at word boundaries and consecutive matches.
 */
function fuzzyScore(query: string, text: string): number {
  const queryLower = query.toLowerCase()
  const textLower = text.toLowerCase()

  let score = 0
  let queryIndex = 0
  let consecutiveMatches = 0

  for (let i = 0; i < textLower.length && queryIndex < queryLower.length; i++) {
    if (textLower[i] === queryLower[queryIndex]) {
      // Bonus for match at start
      if (i === 0) score += 10

      // Bonus for match after word boundary
      if (i > 0 && /\s/.test(text[i - 1])) score += 5

      // Bonus for consecutive matches
      consecutiveMatches++
      score += consecutiveMatches * 2

      queryIndex++
    } else {
      consecutiveMatches = 0
    }
  }

  // Penalty for longer strings (prefer shorter matches)
  score -= text.length * 0.1

  return score
}

export function CommandPalette({
  isOpen,
  onClose,
  commands,
}: CommandPaletteProps) {
  const [query, setQuery] = useState('')
  const [selectedIndex, setSelectedIndex] = useState(0)
  const inputRef = useRef<HTMLInputElement>(null)
  const listRef = useRef<HTMLDivElement>(null)

  // Filter and sort commands based on query
  const filteredCommands = useMemo(() => {
    // Filter out commands that don't pass their condition
    const available = commands.filter((cmd) => !cmd.when || cmd.when())

    if (!query) {
      return available
    }

    // Filter by fuzzy match
    const matched = available.filter((cmd) => fuzzyMatch(query, cmd.label))

    // Sort by score (best match first)
    matched.sort((a, b) => fuzzyScore(query, b.label) - fuzzyScore(query, a.label))

    return matched
  }, [commands, query])

  // Group commands by section
  const groupedCommands = useMemo(() => {
    const groups: Record<string, Command[]> = {
      actions: [],
      navigation: [],
      recent: [],
    }

    for (const cmd of filteredCommands) {
      groups[cmd.section]?.push(cmd)
    }

    return groups
  }, [filteredCommands])

  // Reset state when opened
  useEffect(() => {
    if (isOpen) {
      setQuery('')
      setSelectedIndex(0)
      // Focus input after a tick (portal needs to render first)
      setTimeout(() => inputRef.current?.focus(), 0)
    }
  }, [isOpen])

  // Keep selected index in bounds
  useEffect(() => {
    if (selectedIndex >= filteredCommands.length) {
      setSelectedIndex(Math.max(0, filteredCommands.length - 1))
    }
  }, [filteredCommands.length, selectedIndex])

  // Scroll selected item into view
  useEffect(() => {
    const selectedElement = listRef.current?.querySelector(
      `[data-index="${selectedIndex}"]`
    )
    selectedElement?.scrollIntoView({ block: 'nearest' })
  }, [selectedIndex])

  const executeCommand = useCallback(
    (command: Command) => {
      onClose()
      // Execute after close animation
      setTimeout(() => command.action(), 50)
    },
    [onClose]
  )

  // Handle keyboard navigation within palette
  const handleKeyDown = useCallback(
    (event: React.KeyboardEvent) => {
      switch (event.key) {
        case 'ArrowDown':
          event.preventDefault()
          setSelectedIndex((i) => Math.min(i + 1, filteredCommands.length - 1))
          break

        case 'ArrowUp':
          event.preventDefault()
          setSelectedIndex((i) => Math.max(i - 1, 0))
          break

        case 'Enter':
          event.preventDefault()
          if (filteredCommands[selectedIndex]) {
            executeCommand(filteredCommands[selectedIndex])
          }
          break

        case 'Escape':
          event.preventDefault()
          onClose()
          break
      }
    },
    [filteredCommands, selectedIndex, executeCommand, onClose]
  )

  if (!isOpen) return null

  // Get flat index for a command (for keyboard navigation)
  let flatIndex = 0
  const getFlatIndex = () => flatIndex++

  const palette = (
    <div className={styles.overlay} onClick={onClose}>
      <div
        className={styles.palette}
        onClick={(e) => e.stopPropagation()}
        onKeyDown={handleKeyDown}
      >
        {/* Search input */}
        <div className={styles.inputWrapper}>
          <span className={styles.searchIcon}>
            <svg
              width="16"
              height="16"
              viewBox="0 0 16 16"
              fill="currentColor"
            >
              <path d="M11.742 10.344a6.5 6.5 0 1 0-1.397 1.398h-.001c.03.04.062.078.098.115l3.85 3.85a1 1 0 0 0 1.415-1.414l-3.85-3.85a1.007 1.007 0 0 0-.115-.1zM12 6.5a5.5 5.5 0 1 1-11 0 5.5 5.5 0 0 1 11 0z" />
            </svg>
          </span>
          <input
            ref={inputRef}
            type="text"
            className={styles.input}
            placeholder="Type a command or search..."
            value={query}
            onChange={(e) => {
              setQuery(e.target.value)
              setSelectedIndex(0)
            }}
          />
        </div>

        {/* Command list */}
        <div className={styles.list} ref={listRef}>
          {filteredCommands.length === 0 ? (
            <div className={styles.empty}>No commands found</div>
          ) : (
            <>
              {/* Actions section */}
              {groupedCommands.actions.length > 0 && (
                <div className={styles.section}>
                  <div className={styles.sectionTitle}>ACTIONS</div>
                  {groupedCommands.actions.map((cmd) => {
                    const index = getFlatIndex()
                    return (
                      <button
                        key={cmd.id}
                        data-index={index}
                        className={`${styles.item} ${
                          index === selectedIndex ? styles.selected : ''
                        }`}
                        onClick={() => executeCommand(cmd)}
                        onMouseEnter={() => setSelectedIndex(index)}
                      >
                        <span className={styles.icon}>{cmd.icon || '●'}</span>
                        <span className={styles.label}>{cmd.label}</span>
                        {cmd.shortcut && (
                          <span className={styles.shortcut}>
                            {formatKeyCombo(cmd.shortcut)}
                          </span>
                        )}
                      </button>
                    )
                  })}
                </div>
              )}

              {/* Navigation section */}
              {groupedCommands.navigation.length > 0 && (
                <div className={styles.section}>
                  <div className={styles.sectionTitle}>NAVIGATION</div>
                  {groupedCommands.navigation.map((cmd) => {
                    const index = getFlatIndex()
                    return (
                      <button
                        key={cmd.id}
                        data-index={index}
                        className={`${styles.item} ${
                          index === selectedIndex ? styles.selected : ''
                        }`}
                        onClick={() => executeCommand(cmd)}
                        onMouseEnter={() => setSelectedIndex(index)}
                      >
                        <span className={styles.icon}>{cmd.icon || '→'}</span>
                        <span className={styles.label}>{cmd.label}</span>
                        {cmd.shortcut && (
                          <span className={styles.shortcut}>
                            {formatKeyCombo(cmd.shortcut)}
                          </span>
                        )}
                      </button>
                    )
                  })}
                </div>
              )}

              {/* Recent section */}
              {groupedCommands.recent.length > 0 && (
                <div className={styles.section}>
                  <div className={styles.sectionTitle}>RECENT TASKS</div>
                  {groupedCommands.recent.map((cmd) => {
                    const index = getFlatIndex()
                    return (
                      <button
                        key={cmd.id}
                        data-index={index}
                        className={`${styles.item} ${
                          index === selectedIndex ? styles.selected : ''
                        }`}
                        onClick={() => executeCommand(cmd)}
                        onMouseEnter={() => setSelectedIndex(index)}
                      >
                        <span className={styles.label}>{cmd.label}</span>
                      </button>
                    )
                  })}
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  )

  // Use portal to render at document root (above all other content)
  return createPortal(palette, document.body)
}
