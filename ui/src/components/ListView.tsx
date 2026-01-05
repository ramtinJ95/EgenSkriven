import { useMemo, useState, useCallback } from 'react'
import type { Task, Column, Priority } from '../types/task'
import { COLUMN_NAMES, PRIORITY_NAMES } from '../types/task'
import styles from './ListView.module.css'

interface ColumnDef {
  key: string
  label: string
  width: string
  sortable: boolean
}

const COLUMNS: ColumnDef[] = [
  { key: 'column', label: 'Status', width: '120px', sortable: true },
  { key: 'seq', label: 'ID', width: '80px', sortable: true },
  { key: 'title', label: 'Title', width: 'auto', sortable: true },
  { key: 'labels', label: 'Labels', width: '180px', sortable: false },
  { key: 'priority', label: 'Priority', width: '100px', sortable: true },
  { key: 'due_date', label: 'Due', width: '110px', sortable: true },
]

interface ListViewProps {
  tasks: Task[]
  onTaskClick: (task: Task) => void
  selectedTaskId: string | null
}

export function ListView({ tasks, onTaskClick, selectedTaskId }: ListViewProps) {
  const [sortColumn, setSortColumn] = useState('column')
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc')

  const sortedTasks = useMemo(() => {
    return [...tasks].sort((a, b) => {
      const aVal = a[sortColumn as keyof Task]
      const bVal = b[sortColumn as keyof Task]

      // Handle null/undefined values
      if (aVal == null && bVal == null) return 0
      if (aVal == null) return sortDirection === 'asc' ? 1 : -1
      if (bVal == null) return sortDirection === 'asc' ? -1 : 1

      // Compare strings
      if (typeof aVal === 'string' && typeof bVal === 'string') {
        const comparison = aVal.localeCompare(bVal)
        return sortDirection === 'asc' ? comparison : -comparison
      }

      // Compare numbers
      if (typeof aVal === 'number' && typeof bVal === 'number') {
        return sortDirection === 'asc' ? aVal - bVal : bVal - aVal
      }

      return 0
    })
  }, [tasks, sortColumn, sortDirection])

  const handleSort = useCallback(
    (col: string) => {
      if (col === sortColumn) {
        setSortDirection((d) => (d === 'asc' ? 'desc' : 'asc'))
      } else {
        setSortColumn(col)
        setSortDirection('asc')
      }
    },
    [sortColumn]
  )

  const handleRowClick = useCallback(
    (task: Task) => {
      onTaskClick(task)
    },
    [onTaskClick]
  )

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent, task: Task) => {
      if (e.key === 'Enter' || e.key === ' ') {
        e.preventDefault()
        onTaskClick(task)
      }
    },
    [onTaskClick]
  )

  if (tasks.length === 0) {
    return (
      <div className={styles.emptyState}>
        <p>No tasks match the current filters</p>
      </div>
    )
  }

  return (
    <div className={styles.listView}>
      <table className={styles.table}>
        <thead className={styles.thead}>
          <tr>
            {COLUMNS.map((col) => (
              <th
                key={col.key}
                className={`${styles.th} ${col.sortable ? styles.sortable : ''}`}
                style={{ width: col.width }}
                onClick={() => col.sortable && handleSort(col.key)}
              >
                <span className={styles.thContent}>
                  {col.label}
                  {sortColumn === col.key && (
                    <span className={styles.sortIndicator}>
                      {sortDirection === 'asc' ? '↑' : '↓'}
                    </span>
                  )}
                </span>
              </th>
            ))}
          </tr>
        </thead>
        <tbody className={styles.tbody}>
          {sortedTasks.map((task) => (
            <tr
              key={task.id}
              className={`${styles.tr} ${selectedTaskId === task.id ? styles.selected : ''}`}
              onClick={() => handleRowClick(task)}
              onKeyDown={(e) => handleKeyDown(e, task)}
              tabIndex={0}
              role="button"
              aria-selected={selectedTaskId === task.id}
            >
              <td className={styles.td}>
                <StatusBadge status={task.column} />
              </td>
              <td className={styles.td}>
                <span className={styles.taskId}>#{task.seq ?? task.id.slice(0, 6)}</span>
              </td>
              <td className={styles.td}>
                <span className={styles.title}>{task.title}</span>
              </td>
              <td className={styles.td}>
                <div className={styles.labels}>
                  {(task.labels || []).slice(0, 3).map((label) => (
                    <span key={label} className={styles.labelPill}>
                      {label}
                    </span>
                  ))}
                  {(task.labels || []).length > 3 && (
                    <span className={styles.labelMore}>+{task.labels!.length - 3}</span>
                  )}
                </div>
              </td>
              <td className={styles.td}>
                <PriorityBadge priority={task.priority} />
              </td>
              <td className={styles.td}>
                <DueDate date={task.due_date} />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function StatusBadge({ status }: { status: Column }) {
  return (
    <span className={`${styles.statusBadge} ${styles[`status_${status}`]}`}>
      <span className={styles.statusDot} />
      {COLUMN_NAMES[status] || status.replace('_', ' ')}
    </span>
  )
}

function PriorityBadge({ priority }: { priority: Priority }) {
  return (
    <span className={`${styles.priorityBadge} ${styles[`priority_${priority}`]}`}>
      {PRIORITY_NAMES[priority] || priority}
    </span>
  )
}

function DueDate({ date }: { date?: string }) {
  if (!date) {
    return <span className={styles.noDate}>-</span>
  }

  const dueDate = new Date(date)
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  dueDate.setHours(0, 0, 0, 0)

  const diffDays = Math.ceil((dueDate.getTime() - today.getTime()) / (1000 * 60 * 60 * 24))
  const isOverdue = diffDays < 0
  const isDueSoon = diffDays >= 0 && diffDays <= 2

  const formatted = dueDate.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
  })

  return (
    <span
      className={`${styles.dueDate} ${isOverdue ? styles.overdue : ''} ${isDueSoon ? styles.dueSoon : ''}`}
    >
      {formatted}
    </span>
  )
}
