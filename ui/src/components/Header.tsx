import { SearchBar } from './SearchBar'
import styles from './Header.module.css'

interface HeaderProps {
  onDisplayOptionsClick?: () => void
  onSettingsClick?: () => void
}

/**
 * Application header with search and display options.
 */
export function Header({ onDisplayOptionsClick, onSettingsClick }: HeaderProps) {
  return (
    <header className={styles.header}>
      <div className={styles.title}>
        <span className={styles.logo}>EgenSkriven</span>
      </div>

      <div className={styles.center}>
        <SearchBar />
      </div>

      <div className={styles.actions}>
        <button
          className={styles.displayButton}
          onClick={onDisplayOptionsClick}
          title="Display options"
        >
          <DisplayIcon />
          <span>Display</span>
        </button>
        <button
          className={styles.iconButton}
          onClick={onSettingsClick}
          title="Settings (Ctrl+,)"
        >
          <SettingsIcon />
        </button>
        <span className={styles.shortcut}>
          <kbd>?</kbd> Help
        </span>
      </div>
    </header>
  )
}

function DisplayIcon() {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M2 4h12M2 8h12M2 12h12"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
      />
    </svg>
  )
}

function SettingsIcon() {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M8 10a2 2 0 100-4 2 2 0 000 4z"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M13.6 10a1.2 1.2 0 00.24 1.32l.04.04a1.46 1.46 0 11-2.06 2.06l-.04-.04a1.2 1.2 0 00-1.32-.24 1.2 1.2 0 00-.72 1.1v.12a1.46 1.46 0 11-2.92 0v-.06a1.2 1.2 0 00-.78-1.1 1.2 1.2 0 00-1.32.24l-.04.04a1.46 1.46 0 11-2.06-2.06l.04-.04a1.2 1.2 0 00.24-1.32 1.2 1.2 0 00-1.1-.72h-.12a1.46 1.46 0 110-2.92h.06a1.2 1.2 0 001.1-.78 1.2 1.2 0 00-.24-1.32l-.04-.04a1.46 1.46 0 112.06-2.06l.04.04a1.2 1.2 0 001.32.24h.06a1.2 1.2 0 00.72-1.1v-.12a1.46 1.46 0 012.92 0v.06a1.2 1.2 0 00.72 1.1 1.2 1.2 0 001.32-.24l.04-.04a1.46 1.46 0 112.06 2.06l-.04.04a1.2 1.2 0 00-.24 1.32v.06a1.2 1.2 0 001.1.72h.12a1.46 1.46 0 010 2.92h-.06a1.2 1.2 0 00-1.1.72z"
        stroke="currentColor"
        strokeWidth="1.2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  )
}
