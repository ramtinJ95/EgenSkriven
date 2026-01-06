import { useMemo } from 'react'
import { useEpics } from '../hooks/useEpics'
import type { Task } from '../types/task'
import { EPIC_COLORS } from '../types/epic'
import styles from './EpicList.module.css'

interface EpicListProps {
  /** All tasks to calculate counts from */
  tasks: Task[]
  /** Currently selected epic ID */
  selectedEpicId?: string | null
  /** Callback when an epic is selected */
  onSelectEpic: (epicId: string | null) => void
}

/**
 * EpicList component for sidebar.
 *
 * Features:
 * - Lists all epics with color indicators
 * - Shows task count per epic
 * - Supports selecting an epic to filter tasks
 * - "All Epics" option to clear filter
 * - Loading and error states
 */
export function EpicList({ tasks, selectedEpicId, onSelectEpic }: EpicListProps) {
  const { epics, loading, error } = useEpics()

  // Calculate task counts per epic
  const epicCounts = useMemo(() => {
    const counts: Record<string, number> = {}
    tasks.forEach((task) => {
      if (task.epic) {
        counts[task.epic] = (counts[task.epic] || 0) + 1
      }
    })
    return counts
  }, [tasks])

  // Count tasks without an epic
  const noEpicCount = useMemo(() => {
    return tasks.filter((task) => !task.epic).length
  }, [tasks])

  // Total task count
  const totalCount = tasks.length

  if (loading) {
    return (
      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>EPICS</h2>
        <div className={styles.loading}>Loading...</div>
      </section>
    )
  }

  if (error) {
    return (
      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>EPICS</h2>
        <div className={styles.error}>Failed to load epics</div>
      </section>
    )
  }

  return (
    <section className={styles.section}>
      <h2 className={styles.sectionTitle}>EPICS</h2>

      <ul className={styles.epicList}>
        {/* All Epics option */}
        <li>
          <button
            className={`${styles.epicItem} ${selectedEpicId === null ? styles.active : ''}`}
            onClick={() => onSelectEpic(null)}
          >
            <span className={styles.epicIcon}>
              <AllIcon />
            </span>
            <span className={styles.epicName}>All Tasks</span>
            <span className={styles.epicCount}>{totalCount}</span>
          </button>
        </li>

        {/* Epic list */}
        {epics.map((epic) => (
          <li key={epic.id}>
            <button
              className={`${styles.epicItem} ${selectedEpicId === epic.id ? styles.active : ''}`}
              onClick={() => onSelectEpic(epic.id)}
            >
              <span
                className={styles.epicIndicator}
                style={{ backgroundColor: epic.color || EPIC_COLORS[0] }}
              />
              <span className={styles.epicName}>{epic.title}</span>
              <span className={styles.epicCount}>{epicCounts[epic.id] || 0}</span>
            </button>
          </li>
        ))}

        {/* No Epic option (only show if there are tasks without epics) */}
        {noEpicCount > 0 && (
          <li>
            <button
              className={`${styles.epicItem} ${selectedEpicId === 'none' ? styles.active : ''}`}
              onClick={() => onSelectEpic('none')}
            >
              <span className={styles.epicIcon}>
                <NoEpicIcon />
              </span>
              <span className={styles.epicName}>No Epic</span>
              <span className={styles.epicCount}>{noEpicCount}</span>
            </button>
          </li>
        )}
      </ul>

      {epics.length === 0 && (
        <div className={styles.empty}>No epics yet</div>
      )}
    </section>
  )
}

// Icon for "All Tasks" option
function AllIcon() {
  return (
    <svg
      width="14"
      height="14"
      viewBox="0 0 16 16"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <rect x="2" y="2" width="12" height="12" rx="2" />
      <path d="M2 6h12" />
      <path d="M6 2v12" />
    </svg>
  )
}

// Icon for "No Epic" option
function NoEpicIcon() {
  return (
    <svg
      width="14"
      height="14"
      viewBox="0 0 16 16"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <circle cx="8" cy="8" r="6" />
      <path d="M4 4l8 8" />
    </svg>
  )
}
