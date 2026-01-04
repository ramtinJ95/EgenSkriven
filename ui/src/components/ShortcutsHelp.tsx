import { createPortal } from 'react-dom'
import { formatKeyCombo, type KeyCombo } from '../hooks/useKeyboard'
import styles from './ShortcutsHelp.module.css'

interface ShortcutGroup {
  title: string
  shortcuts: Array<{
    combo: KeyCombo
    description: string
  }>
}

interface ShortcutsHelpProps {
  isOpen: boolean
  onClose: () => void
}

const SHORTCUT_GROUPS: ShortcutGroup[] = [
  {
    title: 'Global',
    shortcuts: [
      { combo: { key: 'k', meta: true }, description: 'Command palette' },
      { combo: { key: '?' }, description: 'Show shortcuts help' },
    ],
  },
  {
    title: 'Task Actions',
    shortcuts: [
      { combo: { key: 'c' }, description: 'Create new task' },
      { combo: { key: 'Enter' }, description: 'Open selected task' },
      { combo: { key: ' ' }, description: 'Peek preview' },
      { combo: { key: 'e' }, description: 'Edit task (open detail)' },
      { combo: { key: 'Backspace' }, description: 'Delete task' },
    ],
  },
  {
    title: 'Task Properties',
    shortcuts: [
      { combo: { key: 's' }, description: 'Set status' },
      { combo: { key: 'p' }, description: 'Set priority' },
      { combo: { key: 't' }, description: 'Set type' },
    ],
  },
  {
    title: 'Navigation',
    shortcuts: [
      { combo: { key: 'j' }, description: 'Next task' },
      { combo: { key: 'k' }, description: 'Previous task' },
      { combo: { key: 'h' }, description: 'Previous column' },
      { combo: { key: 'l' }, description: 'Next column' },
      { combo: { key: 'ArrowDown' }, description: 'Next task (arrow)' },
      { combo: { key: 'ArrowUp' }, description: 'Previous task (arrow)' },
      { combo: { key: 'Escape' }, description: 'Close panel / deselect' },
    ],
  },
  {
    title: 'Selection',
    shortcuts: [
      { combo: { key: 'x' }, description: 'Toggle select task' },
      { combo: { key: 'x', shift: true }, description: 'Select range' },
      { combo: { key: 'a', meta: true }, description: 'Select all visible' },
    ],
  },
]

export function ShortcutsHelp({ isOpen, onClose }: ShortcutsHelpProps) {
  if (!isOpen) return null

  const modal = (
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
        <div className={styles.header}>
          <h2 className={styles.title}>Keyboard Shortcuts</h2>
          <button className={styles.closeButton} onClick={onClose}>
            <svg
              width="16"
              height="16"
              viewBox="0 0 16 16"
              fill="currentColor"
            >
              <path d="M4.646 4.646a.5.5 0 0 1 .708 0L8 7.293l2.646-2.647a.5.5 0 0 1 .708.708L8.707 8l2.647 2.646a.5.5 0 0 1-.708.708L8 8.707l-2.646 2.647a.5.5 0 0 1-.708-.708L7.293 8 4.646 5.354a.5.5 0 0 1 0-.708z" />
            </svg>
          </button>
        </div>

        <div className={styles.content}>
          {SHORTCUT_GROUPS.map((group) => (
            <div key={group.title} className={styles.group}>
              <h3 className={styles.groupTitle}>{group.title}</h3>
              <div className={styles.shortcuts}>
                {group.shortcuts.map((shortcut, index) => (
                  <div key={index} className={styles.shortcut}>
                    <span className={styles.description}>
                      {shortcut.description}
                    </span>
                    <kbd className={styles.key}>
                      {formatKeyCombo(shortcut.combo)}
                    </kbd>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>

        <div className={styles.footer}>
          <span className={styles.hint}>
            Press <kbd>Esc</kbd> to close
          </span>
        </div>
      </div>
    </div>
  )

  return createPortal(modal, document.body)
}
