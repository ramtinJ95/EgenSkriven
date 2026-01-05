import { useState, useEffect, useRef } from 'react'
import type { ParsedView } from '../hooks/useViews'
import { useViews } from '../hooks/useViews'
import { useFilterStore } from '../stores/filters'
import styles from './ViewsSidebar.module.css'

export interface ViewsSidebarProps {
  boardId: string | null
}

export function ViewsSidebar({ boardId }: ViewsSidebarProps) {
  const { views, loading, toggleFavorite, applyView, deleteView, saveCurrentAsView } = useViews(boardId)
  const currentViewId = useFilterStore((s) => s.currentViewId)
  const isModified = useFilterStore((s) => s.isModified)
  const filters = useFilterStore((s) => s.filters)
  const clearView = useFilterStore((s) => s.clearView)

  const [isCreating, setIsCreating] = useState(false)
  const [newViewName, setNewViewName] = useState('')
  const [savingView, setSavingView] = useState(false)
  
  // Confirmation dialog state
  const [deleteConfirmId, setDeleteConfirmId] = useState<string | null>(null)
  const confirmDialogRef = useRef<HTMLDivElement>(null)

  // Separate favorites and regular views
  const favoriteViews = views.filter((v) => v.is_favorite)
  const regularViews = views.filter((v) => !v.is_favorite)

  const handleSaveView = async () => {
    if (!newViewName.trim()) return

    setSavingView(true)
    try {
      const newView = await saveCurrentAsView(newViewName.trim())
      applyView(newView) // Apply it immediately so currentViewId is set
      setNewViewName('')
      setIsCreating(false)
    } catch (error) {
      console.error('Failed to save view:', error)
    } finally {
      setSavingView(false)
    }
  }

  const handleDeleteClick = (e: React.MouseEvent, viewId: string) => {
    e.stopPropagation()
    setDeleteConfirmId(viewId)
  }

  const handleConfirmDelete = async () => {
    if (!deleteConfirmId) return
    
    await deleteView(deleteConfirmId)
    if (currentViewId === deleteConfirmId) {
      clearView()
    }
    setDeleteConfirmId(null)
  }

  const handleCancelDelete = () => {
    setDeleteConfirmId(null)
  }

  // Close confirmation dialog on Escape key
  useEffect(() => {
    if (!deleteConfirmId) return

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        setDeleteConfirmId(null)
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [deleteConfirmId])

  // Focus the confirm dialog when it opens
  useEffect(() => {
    if (deleteConfirmId && confirmDialogRef.current) {
      confirmDialogRef.current.focus()
    }
  }, [deleteConfirmId])

  const handleToggleFavorite = async (e: React.MouseEvent, viewId: string) => {
    e.stopPropagation()
    await toggleFavorite(viewId)
  }

  const renderViewItem = (view: ParsedView) => {
    const isActive = currentViewId === view.id
    const showModified = isActive && isModified

    return (
      <div
        key={view.id}
        className={`${styles.viewItem} ${isActive ? styles.active : ''}`}
        onClick={() => applyView(view)}
      >
        <div className={styles.viewInfo}>
          <span className={styles.viewName}>
            {view.name}
            {showModified && <span className={styles.modifiedBadge}>Modified</span>}
          </span>
          <span className={styles.filterCount}>
            {view.filters.length} filter{view.filters.length !== 1 ? 's' : ''}
          </span>
        </div>
        <div className={styles.viewActions}>
          <button
            className={`${styles.actionButton} ${view.is_favorite ? styles.favorited : ''}`}
            onClick={(e) => handleToggleFavorite(e, view.id)}
            title={view.is_favorite ? 'Remove from favorites' : 'Add to favorites'}
          >
            {view.is_favorite ? '★' : '☆'}
          </button>
          <button
            className={styles.actionButton}
            onClick={(e) => handleDeleteClick(e, view.id)}
            title="Delete view"
          >
            ×
          </button>
        </div>
      </div>
    )
  }

  if (!boardId) {
    return (
      <div className={styles.sidebar}>
        <div className={styles.empty}>Select a board to see views</div>
      </div>
    )
  }

  return (
    <div className={styles.sidebar}>
      <div className={styles.header}>
        <span className={styles.title}>Views</span>
        {filters.length > 0 && !isCreating && (
          <button
            className={styles.saveButton}
            onClick={() => setIsCreating(true)}
            title="Save current filters as view"
          >
            + Save
          </button>
        )}
      </div>

      {isCreating && (
        <div className={styles.createForm}>
          <input
            type="text"
            className={styles.input}
            placeholder="View name..."
            value={newViewName}
            onChange={(e) => setNewViewName(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') handleSaveView()
              if (e.key === 'Escape') {
                setIsCreating(false)
                setNewViewName('')
              }
            }}
            autoFocus
          />
          <div className={styles.createActions}>
            <button
              className={styles.cancelButton}
              onClick={() => {
                setIsCreating(false)
                setNewViewName('')
              }}
            >
              Cancel
            </button>
            <button
              className={styles.confirmButton}
              onClick={handleSaveView}
              disabled={!newViewName.trim() || savingView}
            >
              {savingView ? 'Saving...' : 'Save'}
            </button>
          </div>
        </div>
      )}

      {loading ? (
        <div className={styles.loading}>Loading views...</div>
      ) : (
        <>
          {favoriteViews.length > 0 && (
            <div className={styles.section}>
              <div className={styles.sectionTitle}>Favorites</div>
              {favoriteViews.map(renderViewItem)}
            </div>
          )}

          {regularViews.length > 0 && (
            <div className={styles.section}>
              <div className={styles.sectionTitle}>All Views</div>
              {regularViews.map(renderViewItem)}
            </div>
          )}

          {views.length === 0 && !isCreating && (
            <div className={styles.empty}>
              <p>No saved views yet.</p>
              <p className={styles.hint}>
                Add filters and click "Save" to create a view.
              </p>
            </div>
          )}
        </>
      )}

      {/* Delete Confirmation Dialog */}
      {deleteConfirmId && (
        <div className={styles.confirmOverlay} onClick={handleCancelDelete}>
          <div
            ref={confirmDialogRef}
            className={styles.confirmDialog}
            onClick={(e) => e.stopPropagation()}
            tabIndex={-1}
            role="alertdialog"
            aria-labelledby="confirm-title"
            aria-describedby="confirm-description"
          >
            <p id="confirm-title" className={styles.confirmTitle}>
              Delete this view?
            </p>
            <p id="confirm-description" className={styles.confirmDescription}>
              This action cannot be undone.
            </p>
            <div className={styles.confirmActions}>
              <button
                className={styles.cancelButton}
                onClick={handleCancelDelete}
              >
                Cancel
              </button>
              <button
                className={styles.deleteButton}
                onClick={handleConfirmDelete}
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
