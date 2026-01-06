import type { RecordModel } from 'pocketbase'

/**
 * Epic record from PocketBase
 * All fields align with migrations/2_epics.go and 11_epics_board_relation.go schema
 */
export interface Epic extends RecordModel {
  title: string
  description?: string
  color?: string // Hex color (e.g., "#3B82F6")
  board: string // Board ID this epic belongs to
}

/**
 * Default colors for epics
 * Same palette as board colors for consistency
 */
export const EPIC_COLORS = [
  '#5E6AD2', // Indigo (default)
  '#9333EA', // Purple
  '#22C55E', // Green
  '#F97316', // Orange
  '#EC4899', // Pink
  '#06B6D4', // Cyan
  '#EF4444', // Red
  '#EAB308', // Yellow
]
