import { useState, useRef, useEffect, useCallback } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import styles from './MarkdownEditor.module.css'

interface MarkdownEditorProps {
  /** Current value */
  value: string
  /** Called when value changes (on save) */
  onChange: (value: string) => void
  /** Placeholder text when empty */
  placeholder?: string
  /** Maximum character count */
  maxLength?: number
  /** Auto-save on blur */
  autoSaveOnBlur?: boolean
}

/**
 * MarkdownEditor component with toolbar and keyboard shortcuts.
 *
 * Features:
 * - Toggle between edit and preview modes
 * - Toolbar with bold, italic, code, heading, list, link, and quote buttons
 * - Keyboard shortcuts (Ctrl+B, Ctrl+I, Ctrl+K, etc.)
 * - Character count
 * - Auto-resize textarea
 * - Full markdown preview with ReactMarkdown + remark-gfm
 */
export function MarkdownEditor({
  value,
  onChange,
  placeholder = 'Add a description... (supports Markdown)',
  maxLength = 10000,
  autoSaveOnBlur = true,
}: MarkdownEditorProps) {
  const [isEditing, setIsEditing] = useState(false)
  const [editValue, setEditValue] = useState(value)
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  // Sync editValue when value prop changes (external updates)
  useEffect(() => {
    if (!isEditing) {
      setEditValue(value)
    }
  }, [value, isEditing])

  // Auto-resize textarea
  useEffect(() => {
    if (textareaRef.current && isEditing) {
      textareaRef.current.style.height = 'auto'
      textareaRef.current.style.height = Math.max(120, textareaRef.current.scrollHeight) + 'px'
    }
  }, [editValue, isEditing])

  // Focus textarea when entering edit mode
  useEffect(() => {
    if (isEditing && textareaRef.current) {
      textareaRef.current.focus()
      // Put cursor at end
      const len = textareaRef.current.value.length
      textareaRef.current.setSelectionRange(len, len)
    }
  }, [isEditing])

  // Save changes
  const handleSave = useCallback(() => {
    onChange(editValue)
    setIsEditing(false)
  }, [editValue, onChange])

  // Cancel editing
  const handleCancel = useCallback(() => {
    setEditValue(value)
    setIsEditing(false)
  }, [value])

  // Wrap selected text with markdown syntax
  const wrapSelection = useCallback((before: string, after: string, onNewLine = false) => {
    const textarea = textareaRef.current
    if (!textarea) return

    const start = textarea.selectionStart
    const end = textarea.selectionEnd
    const selectedText = editValue.substring(start, end)

    // If onNewLine, add newline before if not at start of line
    let prefix = ''
    if (onNewLine && start > 0 && editValue[start - 1] !== '\n') {
      prefix = '\n'
    }

    const newValue =
      editValue.substring(0, start) +
      prefix +
      before +
      selectedText +
      after +
      editValue.substring(end)

    setEditValue(newValue)

    // Restore selection
    setTimeout(() => {
      textarea.focus()
      const newStart = start + prefix.length + before.length
      textarea.setSelectionRange(newStart, newStart + selectedText.length)
    }, 0)
  }, [editValue])

  // Insert text at cursor
  const insertAtCursor = useCallback((text: string) => {
    const textarea = textareaRef.current
    if (!textarea) return

    const start = textarea.selectionStart
    const end = textarea.selectionEnd

    // If at start of line or at beginning, just insert
    // Otherwise, add newline first
    let prefix = ''
    if (start > 0 && editValue[start - 1] !== '\n') {
      prefix = '\n'
    }

    const newValue = editValue.substring(0, start) + prefix + text + editValue.substring(end)
    setEditValue(newValue)

    setTimeout(() => {
      textarea.focus()
      const newPos = start + prefix.length + text.length
      textarea.setSelectionRange(newPos, newPos)
    }, 0)
  }, [editValue])

  // Handle keyboard shortcuts
  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    // Cmd/Ctrl + Enter to save
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault()
      handleSave()
      return
    }
    // Escape to cancel
    if (e.key === 'Escape') {
      e.preventDefault()
      e.stopPropagation()
      handleCancel()
      return
    }
    // Bold: Cmd/Ctrl + B
    if ((e.metaKey || e.ctrlKey) && e.key === 'b') {
      e.preventDefault()
      wrapSelection('**', '**')
      return
    }
    // Italic: Cmd/Ctrl + I
    if ((e.metaKey || e.ctrlKey) && e.key === 'i') {
      e.preventDefault()
      wrapSelection('_', '_')
      return
    }
    // Code: Cmd/Ctrl + E (common in many editors)
    if ((e.metaKey || e.ctrlKey) && e.key === 'e') {
      e.preventDefault()
      wrapSelection('`', '`')
      return
    }
    // Link: Cmd/Ctrl + K
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      e.preventDefault()
      wrapSelection('[', '](url)')
      return
    }
  }, [handleSave, handleCancel, wrapSelection])

  // Handle blur
  const handleBlur = useCallback((e: React.FocusEvent) => {
    // Don't save if clicking on toolbar buttons
    if (e.relatedTarget?.closest(`.${styles.toolbar}`)) {
      return
    }
    if (autoSaveOnBlur) {
      handleSave()
    }
  }, [autoSaveOnBlur, handleSave])

  // Start editing
  const startEditing = useCallback(() => {
    setEditValue(value)
    setIsEditing(true)
  }, [value])

  // Render preview mode
  if (!isEditing) {
    return (
      <div className={styles.container}>
        {value ? (
          <div
            className={styles.preview}
            onClick={startEditing}
            role="button"
            tabIndex={0}
            onKeyDown={(e) => e.key === 'Enter' && startEditing()}
          >
            <ReactMarkdown remarkPlugins={[remarkGfm]}>{value}</ReactMarkdown>
          </div>
        ) : (
          <button className={styles.addButton} onClick={startEditing}>
            {placeholder}
          </button>
        )}
      </div>
    )
  }

  // Render edit mode
  return (
    <div className={styles.container}>
      <div className={styles.editor}>
        {/* Toolbar */}
        <div className={styles.toolbar}>
          <button
            type="button"
            className={styles.toolbarButton}
            onClick={() => wrapSelection('**', '**')}
            title="Bold (Ctrl+B)"
            aria-label="Bold"
          >
            <BoldIcon />
          </button>
          <button
            type="button"
            className={styles.toolbarButton}
            onClick={() => wrapSelection('_', '_')}
            title="Italic (Ctrl+I)"
            aria-label="Italic"
          >
            <ItalicIcon />
          </button>
          <button
            type="button"
            className={styles.toolbarButton}
            onClick={() => wrapSelection('`', '`')}
            title="Code (Ctrl+E)"
            aria-label="Inline Code"
          >
            <CodeIcon />
          </button>
          <div className={styles.toolbarDivider} />
          <button
            type="button"
            className={styles.toolbarButton}
            onClick={() => insertAtCursor('## ')}
            title="Heading"
            aria-label="Heading"
          >
            <HeadingIcon />
          </button>
          <button
            type="button"
            className={styles.toolbarButton}
            onClick={() => insertAtCursor('- ')}
            title="Bullet List"
            aria-label="Bullet List"
          >
            <ListIcon />
          </button>
          <button
            type="button"
            className={styles.toolbarButton}
            onClick={() => insertAtCursor('- [ ] ')}
            title="Checkbox"
            aria-label="Checkbox"
          >
            <CheckboxIcon />
          </button>
          <div className={styles.toolbarDivider} />
          <button
            type="button"
            className={styles.toolbarButton}
            onClick={() => wrapSelection('[', '](url)')}
            title="Link (Ctrl+K)"
            aria-label="Link"
          >
            <LinkIcon />
          </button>
          <button
            type="button"
            className={styles.toolbarButton}
            onClick={() => insertAtCursor('> ')}
            title="Quote"
            aria-label="Quote"
          >
            <QuoteIcon />
          </button>
        </div>

        {/* Textarea */}
        <textarea
          ref={textareaRef}
          value={editValue}
          onChange={(e) => setEditValue(e.target.value)}
          onKeyDown={handleKeyDown}
          onBlur={handleBlur}
          placeholder={placeholder}
          className={styles.textarea}
          maxLength={maxLength}
        />

        {/* Actions */}
        <div className={styles.actions}>
          <span className={styles.hint}>
            Ctrl+Enter to save, Esc to cancel
          </span>
          <span className={styles.charCount}>
            {editValue.length.toLocaleString()} / {maxLength.toLocaleString()}
          </span>
          <div className={styles.buttons}>
            <button
              type="button"
              className={styles.cancelButton}
              onClick={handleCancel}
            >
              Cancel
            </button>
            <button
              type="button"
              className={styles.saveButton}
              onClick={handleSave}
            >
              Save
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

// Toolbar Icons
function BoldIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
      <path d="M4 2h5.5c1.93 0 3.5 1.57 3.5 3.5 0 1.12-.53 2.12-1.35 2.76A3.5 3.5 0 0 1 10 14H4V2zm2 5h3.5c.83 0 1.5-.67 1.5-1.5S10.33 4 9.5 4H6v3zm0 2v3h4c.83 0 1.5-.67 1.5-1.5S10.83 9 10 9H6z" />
    </svg>
  )
}

function ItalicIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
      <path d="M6 2h6v2H9.7l-2 8H10v2H4v-2h2.3l2-8H6V2z" />
    </svg>
  )
}

function CodeIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <path d="M10 4l4 4-4 4M6 4l-4 4 4 4" />
    </svg>
  )
}

function HeadingIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
      <path d="M3 2v12h2V9h6v5h2V2h-2v5H5V2H3z" />
    </svg>
  )
}

function ListIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
      <path d="M2 4a1 1 0 1 1 0-2 1 1 0 0 1 0 2zm3-1h9v1H5V3zm-3 5a1 1 0 1 1 0-2 1 1 0 0 1 0 2zm3-1h9v1H5V7zm-3 5a1 1 0 1 1 0-2 1 1 0 0 1 0 2zm3-1h9v1H5v-1z" />
    </svg>
  )
}

function CheckboxIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <rect x="2" y="2" width="12" height="12" rx="2" />
      <path d="M5 8l2 2 4-4" />
    </svg>
  )
}

function LinkIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <path d="M6.5 9.5l3-3M7 11a3 3 0 0 1-4.24-4.24l2-2a3 3 0 0 1 4.24 0M9 5a3 3 0 0 1 4.24 4.24l-2 2a3 3 0 0 1-4.24 0" />
    </svg>
  )
}

function QuoteIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
      <path d="M3 3h4v4H5c0 1.1.9 2 2 2v2c-2.2 0-4-1.8-4-4V3zm8 0h4v4h-2c0 1.1.9 2 2 2v2c-2.2 0-4-1.8-4-4V3z" />
    </svg>
  )
}
