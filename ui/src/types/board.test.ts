import { describe, it, expect } from 'vitest'
import { formatDisplayId, parseDisplayId, DEFAULT_COLUMNS, BOARD_COLORS } from './board'

describe('board type helpers', () => {
  describe('formatDisplayId', () => {
    it('formats prefix and sequence into display ID', () => {
      expect(formatDisplayId('WRK', 1)).toBe('WRK-1')
      expect(formatDisplayId('WRK', 123)).toBe('WRK-123')
      expect(formatDisplayId('PER', 9999)).toBe('PER-9999')
    })

    it('handles single character prefixes', () => {
      expect(formatDisplayId('A', 1)).toBe('A-1')
    })

    it('handles long prefixes', () => {
      expect(formatDisplayId('LONGPREFIX', 1)).toBe('LONGPREFIX-1')
    })
  })

  describe('parseDisplayId', () => {
    it('parses valid display IDs', () => {
      expect(parseDisplayId('WRK-123')).toEqual({ prefix: 'WRK', seq: 123 })
      expect(parseDisplayId('PER-1')).toEqual({ prefix: 'PER', seq: 1 })
      expect(parseDisplayId('A-9999')).toEqual({ prefix: 'A', seq: 9999 })
    })

    it('uppercases prefix from lowercase input', () => {
      expect(parseDisplayId('wrk-123')).toEqual({ prefix: 'WRK', seq: 123 })
      expect(parseDisplayId('per-456')).toEqual({ prefix: 'PER', seq: 456 })
    })

    it('handles alphanumeric prefixes', () => {
      expect(parseDisplayId('WRK2-123')).toEqual({ prefix: 'WRK2', seq: 123 })
      expect(parseDisplayId('ABC123-1')).toEqual({ prefix: 'ABC123', seq: 1 })
    })

    it('returns null for invalid formats', () => {
      expect(parseDisplayId('invalid')).toBeNull()
      expect(parseDisplayId('WRK')).toBeNull()
      expect(parseDisplayId('WRK-')).toBeNull()
      expect(parseDisplayId('-123')).toBeNull()
      expect(parseDisplayId('')).toBeNull()
      expect(parseDisplayId('WRK-abc')).toBeNull()
    })

    it('returns null for special characters in prefix', () => {
      expect(parseDisplayId('WRK@-123')).toBeNull()
      expect(parseDisplayId('WRK#-123')).toBeNull()
      expect(parseDisplayId('WRK -123')).toBeNull()
    })
  })

  describe('DEFAULT_COLUMNS', () => {
    it('has standard kanban columns', () => {
      expect(DEFAULT_COLUMNS).toEqual([
        'backlog',
        'todo',
        'in_progress',
        'review',
        'done',
      ])
    })

    it('has 5 columns', () => {
      expect(DEFAULT_COLUMNS).toHaveLength(5)
    })
  })

  describe('BOARD_COLORS', () => {
    it('has 8 color options', () => {
      expect(BOARD_COLORS).toHaveLength(8)
    })

    it('all colors are valid hex codes', () => {
      const hexColorRegex = /^#[0-9A-Fa-f]{6}$/
      BOARD_COLORS.forEach((color) => {
        expect(color).toMatch(hexColorRegex)
      })
    })

    it('includes indigo as first color (default)', () => {
      expect(BOARD_COLORS[0]).toBe('#5E6AD2')
    })
  })
})
