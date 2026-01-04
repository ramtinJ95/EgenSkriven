import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { useTasks } from './useTasks'

// Mock PocketBase
vi.mock('../lib/pb', () => ({
  pb: {
    collection: vi.fn(() => ({
      getFullList: vi.fn().mockResolvedValue([]),
      subscribe: vi.fn().mockReturnValue(Promise.resolve()),
      unsubscribe: vi.fn(),
      create: vi.fn(),
      update: vi.fn(),
      delete: vi.fn(),
    })),
  },
}))

describe('useTasks', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('starts with loading state', () => {
    const { result } = renderHook(() => useTasks())
    expect(result.current.loading).toBe(true)
    expect(result.current.tasks).toEqual([])
  })

  it('fetches tasks on mount', async () => {
    const { result } = renderHook(() => useTasks())
    
    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })
  })

  it('provides CRUD operations', () => {
    const { result } = renderHook(() => useTasks())
    
    expect(typeof result.current.createTask).toBe('function')
    expect(typeof result.current.updateTask).toBe('function')
    expect(typeof result.current.deleteTask).toBe('function')
    expect(typeof result.current.moveTask).toBe('function')
  })
})
