import { SearchBar } from './SearchBar'
import styles from './Header.module.css'

interface HeaderProps {
  onDisplayOptionsClick?: () => void
}

/**
 * Application header with search and display options.
 */
export function Header({ onDisplayOptionsClick }: HeaderProps) {
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
