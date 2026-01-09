import type { ResumeMode } from '../../types/board'
import styles from './ResumeModeSelector.module.css'

interface ResumeModeSelectorProps {
  /** Current selected resume mode */
  value: ResumeMode
  /** Callback when mode changes */
  onChange: (mode: ResumeMode) => void
  /** Whether the selector is disabled */
  disabled?: boolean
}

/** Mode option configuration */
interface ModeOption {
  value: ResumeMode
  label: string
  description: string
}

/** Available resume mode options with descriptions */
const modes: ModeOption[] = [
  {
    value: 'manual',
    label: 'Manual',
    description: 'Print resume command for user to copy and run',
  },
  {
    value: 'command',
    label: 'Command',
    description: 'User runs "egenskriven resume --exec" to resume',
  },
  {
    value: 'auto',
    label: 'Auto',
    description: 'Automatically resume when human adds @agent comment',
  },
]

/**
 * ResumeModeSelector provides a UI for selecting how blocked tasks
 * should be resumed after human input.
 *
 * @example
 * ```tsx
 * <ResumeModeSelector
 *   value={boardResumeMode}
 *   onChange={(mode) => updateBoard({ resume_mode: mode })}
 * />
 * ```
 */
export function ResumeModeSelector({
  value,
  onChange,
  disabled = false,
}: ResumeModeSelectorProps) {
  return (
    <div className={styles.container}>
      <label className={styles.label}>Resume Mode</label>
      <p className={styles.hint}>
        How blocked tasks should be resumed after human input
      </p>

      <div className={styles.options}>
        {modes.map((mode) => (
          <label
            key={mode.value}
            className={`${styles.option} ${
              value === mode.value ? styles.selected : ''
            } ${disabled ? styles.disabled : ''}`}
          >
            <input
              type="radio"
              name="resumeMode"
              value={mode.value}
              checked={value === mode.value}
              onChange={() => onChange(mode.value)}
              disabled={disabled}
              className={styles.radio}
            />
            <div className={styles.optionContent}>
              <span className={styles.optionLabel}>{mode.label}</span>
              <span className={styles.optionDescription}>
                {mode.description}
              </span>
            </div>
          </label>
        ))}
      </div>

      {value === 'auto' && (
        <div className={styles.warning}>
          <WarningIcon />
          <p>
            <strong>Note:</strong> Auto-resume will trigger when a human adds a
            comment containing <code className={styles.code}>@agent</code> to a
            blocked task with a linked session.
          </p>
        </div>
      )}
    </div>
  )
}

/** Warning icon for auto mode */
function WarningIcon() {
  return (
    <svg
      className={styles.warningIcon}
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M8 6v4M8 12h.01" />
      <path d="M6.86 2.857l-5.65 9.974A1 1 0 002.07 14h11.86a1 1 0 00.86-1.514L9.14 2.857a1 1 0 00-1.72 0z" />
    </svg>
  )
}

export default ResumeModeSelector
