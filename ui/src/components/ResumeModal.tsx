import { useState, useEffect } from 'react'
import { useResume, type ResumeResult } from '../hooks/useResume'
import styles from './ResumeModal.module.css'

interface ResumeModalProps {
  isOpen: boolean
  onClose: () => void
  taskId: string
  displayId: string
}

/**
 * ResumeModal generates and displays the resume command for an agent session.
 *
 * Features:
 * - Generate resume command with full context
 * - Display tool name and session info
 * - Copy command to clipboard
 * - Display working directory
 * - Instructions for resuming
 */
export function ResumeModal({ isOpen, onClose, taskId, displayId }: ResumeModalProps) {
  const { resume, loading } = useResume()
  const [result, setResult] = useState<ResumeResult | null>(null)
  const [copied, setCopied] = useState(false)
  const [generateError, setGenerateError] = useState<string | null>(null)

  // Reset state when modal opens
  useEffect(() => {
    if (isOpen) {
      setResult(null)
      setCopied(false)
      setGenerateError(null)
    }
  }, [isOpen])

  // Handle keyboard shortcuts
  useEffect(() => {
    if (!isOpen) return

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose()
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [isOpen, onClose])

  if (!isOpen) return null

  const handleGenerate = async () => {
    setGenerateError(null)
    try {
      const res = await resume({ taskId, displayId, exec: false })
      setResult(res)
    } catch (err) {
      console.error('Failed to generate resume command:', err)
      setGenerateError(err instanceof Error ? err.message : 'Failed to generate command')
    }
  }

  const handleCopy = async () => {
    if (!result?.command) return

    try {
      await navigator.clipboard.writeText(result.command)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error('Failed to copy to clipboard:', err)
    }
  }

  const handleOverlayClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose()
    }
  }

  return (
    <div
      className={styles.overlay}
      onClick={handleOverlayClick}
      role="dialog"
      aria-modal="true"
      aria-labelledby="resume-modal-title"
    >
      <div className={styles.modal}>
        {/* Header */}
        <div className={styles.header}>
          <h2 id="resume-modal-title" className={styles.title}>Resume Session for {displayId}</h2>
          <button onClick={onClose} className={styles.closeButton} aria-label="Close modal">
            <CloseIcon />
          </button>
        </div>

        {/* Content */}
        <div className={styles.content}>
          {!result ? (
            // Initial state - generate button
            <div className={styles.generateSection}>
              <p className={styles.description}>
                Generate the command to resume the agent session with full context
                from comments and task details.
              </p>

              {generateError && (
                <div className={styles.error} role="alert">{generateError}</div>
              )}

              <button
                onClick={handleGenerate}
                disabled={loading}
                className={styles.generateButton}
                aria-busy={loading}
              >
                {loading ? (
                  <>
                    <SpinnerIcon className={styles.spinner} />
                    Generating...
                  </>
                ) : (
                  'Generate Resume Command'
                )}
              </button>
            </div>
          ) : (
            // Result state - show command
            <div className={styles.resultSection}>
              {/* Tool info */}
              <div className={styles.infoRow}>
                <span className={styles.infoLabel}>Tool:</span>
                <span className={styles.infoValue}>{result.tool}</span>
              </div>

              {/* Working directory */}
              <div className={styles.infoRow}>
                <span className={styles.infoLabel}>Working Dir:</span>
                <span className={styles.infoValueMono} title={result.workingDir}>
                  {result.workingDir}
                </span>
              </div>

              {/* Command */}
              <div className={styles.commandSection}>
                <label id="command-label" className={styles.commandLabel}>Resume Command</label>
                <div className={styles.commandWrapper}>
                  <pre className={styles.command} aria-labelledby="command-label" tabIndex={0}>
                    {result.command}
                  </pre>
                  <button
                    onClick={handleCopy}
                    className={styles.copyButton}
                    aria-label={copied ? 'Command copied to clipboard' : 'Copy command to clipboard'}
                  >
                    {copied ? (
                      <>
                        <CheckIcon />
                        Copied!
                      </>
                    ) : (
                      <>
                        <CopyIcon />
                        Copy
                      </>
                    )}
                  </button>
                </div>
              </div>

              {/* Instructions */}
              <div className={styles.instructions}>
                <h4 className={styles.instructionsTitle}>Instructions</h4>
                <ol className={styles.instructionsList}>
                  <li>Copy the command above</li>
                  <li>
                    Open a terminal in:{' '}
                    <code className={styles.inlineCode}>{result.workingDir}</code>
                  </li>
                  <li>Paste and run the command</li>
                  <li>The agent will resume with full conversation context</li>
                </ol>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

// Icon components
function CloseIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true">
      <path d="M4.28 3.22a.75.75 0 00-1.06 1.06L6.94 8l-3.72 3.72a.75.75 0 101.06 1.06L8 9.06l3.72 3.72a.75.75 0 101.06-1.06L9.06 8l3.72-3.72a.75.75 0 00-1.06-1.06L8 6.94 4.28 3.22z" />
    </svg>
  )
}

function CopyIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true">
      <path d="M0 6.75C0 5.784.784 5 1.75 5h1.5a.75.75 0 010 1.5h-1.5a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-1.5a.75.75 0 011.5 0v1.5A1.75 1.75 0 019.25 16h-7.5A1.75 1.75 0 010 14.25v-7.5z" />
      <path d="M5 1.75C5 .784 5.784 0 6.75 0h7.5C15.216 0 16 .784 16 1.75v7.5A1.75 1.75 0 0114.25 11h-7.5A1.75 1.75 0 015 9.25v-7.5zm1.75-.25a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-7.5a.25.25 0 00-.25-.25h-7.5z" />
    </svg>
  )
}

function CheckIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true">
      <path d="M13.78 4.22a.75.75 0 010 1.06l-7.25 7.25a.75.75 0 01-1.06 0L2.22 9.28a.75.75 0 011.06-1.06L6 10.94l6.72-6.72a.75.75 0 011.06 0z" />
    </svg>
  )
}

function SpinnerIcon({ className }: { className?: string }) {
  return (
    <svg className={className} width="16" height="16" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true">
      <path d="M8 0a8 8 0 100 16A8 8 0 008 0zm0 14a6 6 0 110-12 6 6 0 010 12z" opacity="0.2" />
      <path d="M8 2a6 6 0 016 6h-2a4 4 0 00-4-4V2z" />
    </svg>
  )
}
