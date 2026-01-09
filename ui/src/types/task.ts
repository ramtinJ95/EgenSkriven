import type { RecordModel } from 'pocketbase'
import type { AgentSession } from './session'

// Task type values
export type TaskType = 'bug' | 'feature' | 'chore'

// Priority levels (ordered from highest to lowest)
export type Priority = 'urgent' | 'high' | 'medium' | 'low'

// Column/status values (ordered by workflow)
export type Column = 'backlog' | 'todo' | 'in_progress' | 'need_input' | 'review' | 'done'

// Creator type
export type CreatedBy = 'user' | 'agent' | 'cli'

// History entry for activity tracking
export interface HistoryEntry {
  timestamp: string
  action: 'created' | 'updated' | 'moved' | 'completed' | 'deleted'
  actor: CreatedBy
  actor_detail?: string
  changes?: {
    field: string
    from: unknown
    to: unknown
  }
}

// Task record from PocketBase
// All fields align with migrations/1_initial.go schema
export interface Task extends RecordModel {
  title: string
  description?: string
  type: TaskType
  priority: Priority
  column: Column
  position: number
  labels?: string[]
  blocked_by?: string[]       // Array of task IDs that block this task
  due_date?: string           // Optional due date
  parent?: string             // Parent task ID for sub-tasks
  epic?: string               // Epic ID (relation to epics collection)
  created_by: CreatedBy
  created_by_agent?: string   // Agent identifier (e.g., "claude", "opencode")
  history?: HistoryEntry[]    // Activity tracking array
  board?: string              // Board ID (relation to boards collection)
  seq?: number                // Per-board sequence number for display IDs
  agent_session?: AgentSession // Current linked agent session (JSON field)
}

// All possible columns in display order
export const COLUMNS: Column[] = [
  'backlog',
  'todo',
  'in_progress',
  'need_input',
  'review',
  'done',
]

// Human-readable column names
export const COLUMN_NAMES: Record<Column, string> = {
  backlog: 'Backlog',
  todo: 'Todo',
  in_progress: 'In Progress',
  need_input: 'Need Input',
  review: 'Review',
  done: 'Done',
}

// Human-readable priority names
export const PRIORITY_NAMES: Record<Priority, string> = {
  urgent: 'Urgent',
  high: 'High',
  medium: 'Medium',
  low: 'Low',
}

// All possible priorities in order (highest to lowest)
export const PRIORITIES: Priority[] = ['urgent', 'high', 'medium', 'low']

// Human-readable type names
export const TYPE_NAMES: Record<TaskType, string> = {
  bug: 'Bug',
  feature: 'Feature',
  chore: 'Chore',
}

// All possible types
export const TYPES: TaskType[] = ['bug', 'feature', 'chore']
