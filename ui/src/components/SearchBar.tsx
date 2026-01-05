import { useRef, useEffect, forwardRef, useImperativeHandle } from 'react'
import { useFilterStore } from '../stores/filters'
import styles from './SearchBar.module.css'

export interface SearchBarHandle {
  focus: () => void
}

export interface SearchBarProps {
  placeholder?: string
}

export const SearchBar = forwardRef<SearchBarHandle, SearchBarProps>(
  function SearchBar({ placeholder = 'Search tasks... (/)' }, ref) {
    const inputRef = useRef<HTMLInputElement>(null)
    const searchQuery = useFilterStore((s) => s.searchQuery)
    const setSearchQuery = useFilterStore((s) => s.setSearchQuery)

    // Expose focus method via ref
    useImperativeHandle(ref, () => ({
      focus: () => inputRef.current?.focus(),
    }))

    // Global "/" shortcut to focus search
    useEffect(() => {
      const handleKeyDown = (e: KeyboardEvent) => {
        // Don't trigger if user is typing in an input/textarea
        const target = e.target as HTMLElement
        if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA') {
          return
        }

        // Focus on "/" key
        if (e.key === '/') {
          e.preventDefault()
          inputRef.current?.focus()
        }
      }

      document.addEventListener('keydown', handleKeyDown)
      return () => document.removeEventListener('keydown', handleKeyDown)
    }, [])

    // Handle escape to blur and clear
    const handleKeyDown = (e: React.KeyboardEvent) => {
      if (e.key === 'Escape') {
        if (searchQuery) {
          setSearchQuery('')
        } else {
          inputRef.current?.blur()
        }
      }
    }

    return (
      <div className={styles.searchBar}>
        <svg
          className={styles.searchIcon}
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
        >
          <circle cx="11" cy="11" r="8" />
          <path d="m21 21-4.35-4.35" />
        </svg>
        <input
          ref={inputRef}
          type="text"
          className={styles.input}
          placeholder={placeholder}
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          onKeyDown={handleKeyDown}
        />
        {searchQuery && (
          <button
            className={styles.clearButton}
            onClick={() => setSearchQuery('')}
            aria-label="Clear search"
          >
            &times;
          </button>
        )}
        <kbd className={styles.shortcut}>/</kbd>
      </div>
    )
  }
)
