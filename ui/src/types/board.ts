import type { RecordModel } from 'pocketbase'

/**
 * Board record from PocketBase
 * All fields align with migrations/4_boards.go schema
 */
export interface Board extends RecordModel {
  name: string
  prefix: string
  columns: string[]
  color?: string
}

// Default columns for new boards
export const DEFAULT_COLUMNS = ['backlog', 'todo', 'in_progress', 'review', 'done']

// Default board colors for the color picker
export const BOARD_COLORS = [
  '#5E6AD2', // Indigo (default)
  '#9333EA', // Purple
  '#22C55E', // Green
  '#F97316', // Orange
  '#EC4899', // Pink
  '#06B6D4', // Cyan
  '#EF4444', // Red
  '#EAB308', // Yellow
]

/**
 * Format a display ID from board prefix and sequence number
 * @example formatDisplayId('WRK', 123) returns 'WRK-123'
 */
export function formatDisplayId(prefix: string, seq: number): string {
  return `${prefix}-${seq}`
}

/**
 * Parse a display ID into prefix and sequence number
 * @example parseDisplayId('WRK-123') returns { prefix: 'WRK', seq: 123 }
 */
export function parseDisplayId(displayId: string): { prefix: string; seq: number } | null {
  const match = displayId.match(/^([A-Z0-9]+)-(\d+)$/i)
  if (!match) return null
  return {
    prefix: match[1].toUpperCase(),
    seq: parseInt(match[2], 10),
  }
}
