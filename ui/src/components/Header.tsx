import styles from './Header.module.css'

/**
 * Application header with title and shortcuts hint.
 */
export function Header() {
  return (
    <header className={styles.header}>
      <div className={styles.title}>
        <span className={styles.logo}>EgenSkriven</span>
      </div>
      
      <div className={styles.actions}>
        <span className={styles.shortcut}>
          <kbd>C</kbd> Create
        </span>
        <span className={styles.shortcut}>
          <kbd>Enter</kbd> Open
        </span>
        <span className={styles.shortcut}>
          <kbd>Esc</kbd> Close
        </span>
      </div>
    </header>
  )
}
