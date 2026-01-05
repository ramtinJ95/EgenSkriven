# Real-Time Updates Test Plan & Implementation Guide

**Document Version**: 1.0  
**Created**: January 5, 2026  
**Status**: Ready for Implementation

This document is a self-contained guide for diagnosing and fixing the real-time update issues between the CLI and UI in EgenSkriven. It includes all findings, root cause analysis, code references, and step-by-step implementation instructions.

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Background & Architecture](#background--architecture)
3. [Issue Analysis](#issue-analysis)
4. [Phase 1: Diagnostic Logging](#phase-1-diagnostic-logging)
5. [Phase 2: Fix Task Board Movement Bug](#phase-2-fix-task-board-movement-bug)
6. [Phase 3: Implement Optimistic Updates](#phase-3-implement-optimistic-updates)
7. [Phase 4: SSE Connection Monitoring](#phase-4-sse-connection-monitoring)
8. [Testing Procedures](#testing-procedures)
9. [Appendix: Code References](#appendix-code-references)

---

## Executive Summary

### The Problem

Tasks created or modified via the CLI do not appear in the web UI in real-time. Users must refresh the page to see changes made through the CLI.

### Root Causes Identified

1. **Bug in `useTasks.ts`**: The `update` event handler doesn't add tasks that move INTO the current board from another board
2. **No debug logging**: Impossible to verify if SSE events are being received
3. **No optimistic updates**: UI-initiated changes wait for server round-trip before updating
4. **Subscription not awaited**: Potential race condition where events fire before subscription is ready

### What's Working

- CLI hybrid approach (HTTP API with direct DB fallback) is correctly implemented
- PocketBase collection rules allow public access (`ListRule = ""`)
- SSE connection appears to be established (verified via curl tests)
- PocketBase server hooks are correctly configured

---

## Background & Architecture

### How PocketBase Real-Time Works

PocketBase uses **Server-Sent Events (SSE)** for real-time updates, NOT WebSocket.

```
┌─────────────────────────────────────────────────────────────────┐
│                     PocketBase Server (port 8090)                │
│                                                                  │
│  ┌──────────────────┐    ┌─────────────────────────────────┐   │
│  │ HTTP API Handler │    │ SubscriptionsBroker             │   │
│  │                  │    │  - Browser Client 1 (SSE)       │   │
│  │ OnRecordCreate ──┼───►│  - Browser Client 2 (SSE)       │   │
│  │ hooks broadcast  │    │  - ...                          │   │
│  └──────────────────┘    └─────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

**Key Points:**
- SSE events are sent for `create`, `update`, and `delete` record operations
- Wildcard subscriptions (`'*'`) require the collection's `ListRule` to pass
- Events ARE sent back to the client that initiated the change
- The `SubscriptionsBroker` is per-process (in-memory)

### CLI Hybrid Approach

The CLI was modified to use HTTP API calls when the server is running, enabling real-time broadcasts:

```
CLI Command → IsServerRunning()? 
    ├─ Yes → HTTP API → Server processes → Broadcasts SSE → UI receives
    └─ No  → Direct DB → No broadcast (server not running)
```

**Relevant Files:**
- `internal/commands/client.go` - API client implementation
- `internal/commands/add.go` - `saveRecordHybrid()` function
- `internal/commands/root.go` - `--direct` and `--verbose` flags

### UI Subscription Setup

The UI subscribes to real-time events in `useTasks.ts`:

```typescript
pb.collection('tasks').subscribe<Task>('*', (event) => {
  switch (event.action) {
    case 'create': // Add task to state
    case 'update': // Update task in state  
    case 'delete': // Remove task from state
  }
})
```

---

## Issue Analysis

### Issue 1: Task Moving INTO Board Not Handled (CRITICAL)

**File**: `ui/src/hooks/useTasks.ts`, Lines 87-96

**Current Code:**
```typescript
case 'update':
  // If task was moved to a different board, remove it from this view
  if (boardId && taskBoardId !== boardId) {
    setTasks((prev) => prev.filter((t) => t.id !== event.record.id))
  } else {
    // BUG: map() only replaces existing items, doesn't add new ones!
    setTasks((prev) =>
      prev.map((t) => (t.id === event.record.id ? event.record : t))
    )
  }
```

**The Bug:**
When a task is updated and its `board` field changes TO the current board (from another board), the task won't appear because:
1. The task wasn't in the `tasks` array before (it belonged to a different board)
2. `map()` iterates over existing items and replaces matches
3. Since there's no match, the array is returned unchanged
4. The task never appears in the UI

**Impact:**
- Tasks moved between boards via CLI don't appear
- Multi-board workflows are broken
- Users must refresh to see cross-board moves

### Issue 2: No Debug Logging

**The Problem:**
The subscription callback has no logging, making it impossible to verify:
- If SSE events are being received at all
- What data is in the events
- Why events might be filtered out

**Current Code (no logging):**
```typescript
pb.collection('tasks').subscribe<Task>('*', (event) => {
  const taskBoardId = event.record.board
  const matchesBoard = !boardId || taskBoardId === boardId
  // No way to know if we got here
})
```

### Issue 3: Subscription Not Awaited

**The Problem:**
`subscribe()` returns a Promise, but it's not awaited. This could cause:
- Race conditions where tasks are created before subscription is ready
- Missed events during initial page load

**Current Code:**
```typescript
useEffect(() => {
  // NOT awaited - subscription might not be ready
  pb.collection('tasks').subscribe<Task>('*', (event) => {
    // ...
  }).catch((err) => {
    console.error('Failed to subscribe:', err)
  })
}, [boardId])
```

### Issue 4: No Optimistic Updates

**The Problem:**
UI-initiated changes (via property pickers, drag-drop, etc.) don't update local state immediately. They wait for:
1. HTTP request to complete
2. Server to process
3. SSE event to broadcast
4. UI to receive and process event

This causes:
- Sluggish UI feel
- Stale data in peek preview
- Task cards not moving immediately when status is changed

**Current Code:**
```typescript
const updateTask = useCallback(
  async (id: string, data: Partial<Task>): Promise<Task> => {
    // No local state update - just sends to server
    return pb.collection('tasks').update<Task>(id, data)
  },
  []
)
```

---

## Phase 1: Diagnostic Logging

**Goal:** Add comprehensive logging to verify where events are being lost.

### Step 1.1: Add Logging to useTasks.ts

Replace the subscription effect in `ui/src/hooks/useTasks.ts`:

```typescript
// Subscribe to real-time updates
useEffect(() => {
  console.log('[useTasks] Setting up subscription for board:', boardId)
  
  let isSubscribed = false
  
  pb.collection('tasks')
    .subscribe<Task>('*', (event) => {
      console.log('[useTasks] ========== EVENT RECEIVED ==========')
      console.log('[useTasks] Action:', event.action)
      console.log('[useTasks] Record ID:', event.record.id)
      console.log('[useTasks] Record Title:', event.record.title)
      console.log('[useTasks] Record Board:', event.record.board)
      console.log('[useTasks] Current Board Filter:', boardId)
      
      const taskBoardId = event.record.board
      const matchesBoard = !boardId || taskBoardId === boardId
      console.log('[useTasks] Matches Board:', matchesBoard)

      switch (event.action) {
        case 'create':
          if (matchesBoard) {
            console.log('[useTasks] Adding new task to state')
            setTasks((prev) => {
              const newTasks = [...prev, event.record].sort((a, b) => a.position - b.position)
              console.log('[useTasks] New task count:', newTasks.length)
              return newTasks
            })
          } else {
            console.log('[useTasks] Ignoring create - different board')
          }
          break
          
        case 'update':
          if (boardId && taskBoardId !== boardId) {
            console.log('[useTasks] Removing task - moved to different board')
            setTasks((prev) => prev.filter((t) => t.id !== event.record.id))
          } else {
            console.log('[useTasks] Updating task in state')
            setTasks((prev) => {
              const exists = prev.some((t) => t.id === event.record.id)
              console.log('[useTasks] Task exists in state:', exists)
              if (exists) {
                return prev.map((t) => (t.id === event.record.id ? event.record : t))
              } else {
                // Task moved INTO this board - add it
                console.log('[useTasks] Task moved INTO this board - adding')
                return [...prev, event.record].sort((a, b) => a.position - b.position)
              }
            })
          }
          break
          
        case 'delete':
          console.log('[useTasks] Removing deleted task')
          setTasks((prev) => prev.filter((t) => t.id !== event.record.id))
          break
          
        default:
          console.log('[useTasks] Unknown action:', event.action)
      }
      console.log('[useTasks] =====================================')
    })
    .then(() => {
      isSubscribed = true
      console.log('[useTasks] Subscription established successfully')
    })
    .catch((err) => {
      console.error('[useTasks] Subscription FAILED:', err)
    })

  return () => {
    console.log('[useTasks] Cleaning up subscription, was subscribed:', isSubscribed)
    pb.collection('tasks').unsubscribe('*')
  }
}, [boardId])
```

### Step 1.2: Add SSE Connection Monitoring

Add to `ui/src/lib/pb.ts`:

```typescript
import PocketBase from 'pocketbase'

export const pb = new PocketBase('/')
pb.autoCancellation(false)

// Debug: Log SSE connection events
if (typeof window !== 'undefined') {
  // Monitor connection state
  const checkConnection = () => {
    console.log('[PB] Realtime client ID:', pb.realtime.clientId)
    console.log('[PB] Is connected:', !!pb.realtime.clientId)
  }
  
  // Check periodically during development
  setInterval(checkConnection, 30000)
  
  // Log initial state after a short delay
  setTimeout(checkConnection, 1000)
}
```

### Step 1.3: Test the Logging

1. Start the server:
   ```bash
   ./egenskriven serve
   ```

2. Open browser to `http://localhost:8090`

3. Open browser DevTools Console

4. Create a task via CLI:
   ```bash
   ./egenskriven add "Test real-time" --verbose
   ```

5. Check console for:
   - `[useTasks] Subscription established successfully` - confirms subscription is working
   - `[useTasks] ========== EVENT RECEIVED ==========` - confirms events are being received
   - The full event details

**Expected Outcomes:**
- If NO events appear: SSE connection issue
- If events appear but task doesn't show: Board filtering or state update issue
- If events appear and task shows: Real-time is working!

---

## Phase 2: Fix Task Board Movement Bug

**Goal:** Fix the `update` handler to properly add tasks that move INTO the current board.

### Step 2.1: Update useTasks.ts

**File:** `ui/src/hooks/useTasks.ts`

**Find this code (around lines 87-96):**
```typescript
case 'update':
  // If task was moved to a different board, remove it from this view
  if (boardId && taskBoardId !== boardId) {
    setTasks((prev) => prev.filter((t) => t.id !== event.record.id))
  } else {
    // Replace updated task in state
    setTasks((prev) =>
      prev.map((t) => (t.id === event.record.id ? event.record : t))
    )
  }
  break
```

**Replace with:**
```typescript
case 'update':
  // If task was moved to a different board, remove it from this view
  if (boardId && taskBoardId !== boardId) {
    setTasks((prev) => prev.filter((t) => t.id !== event.record.id))
  } else {
    // Task belongs to this board - update or add it
    setTasks((prev) => {
      const existingIndex = prev.findIndex((t) => t.id === event.record.id)
      if (existingIndex >= 0) {
        // Task exists - replace it
        const updated = [...prev]
        updated[existingIndex] = event.record
        return updated
      } else {
        // Task moved INTO this board from another board - add it
        return [...prev, event.record].sort((a, b) => a.position - b.position)
      }
    })
  }
  break
```

### Step 2.2: Apply Same Fix to useBoards.ts

**File:** `ui/src/hooks/useBoards.ts`

The same pattern exists but is less critical for boards (boards rarely move between... boards). However, for consistency:

**Find (around line 79-83):**
```typescript
case 'update':
  // Replace updated board in state
  setBoards((prev) =>
    prev.map((b) => (b.id === event.record.id ? event.record : b))
  )
  break
```

**Replace with:**
```typescript
case 'update':
  setBoards((prev) => {
    const existingIndex = prev.findIndex((b) => b.id === event.record.id)
    if (existingIndex >= 0) {
      const updated = [...prev]
      updated[existingIndex] = event.record
      return updated.sort((a, b) => a.name.localeCompare(b.name))
    } else {
      // Board somehow appeared - add it
      return [...prev, event.record].sort((a, b) => a.name.localeCompare(b.name))
    }
  })
  break
```

### Step 2.3: Apply Same Fix to useViews.ts

**File:** `ui/src/hooks/useViews.ts`

**Find (around line 96-99):**
```typescript
} else if (event.action === 'update') {
  setViews((prev) =>
    prev.map((v) => (v.id === event.record.id ? parseView(event.record) : v))
  )
}
```

**Replace with:**
```typescript
} else if (event.action === 'update') {
  setViews((prev) => {
    const existingIndex = prev.findIndex((v) => v.id === event.record.id)
    if (existingIndex >= 0) {
      const updated = [...prev]
      updated[existingIndex] = parseView(event.record)
      return updated
    } else {
      return [...prev, parseView(event.record)]
    }
  })
}
```

---

## Phase 3: Implement Optimistic Updates

**Goal:** Update local state immediately for UI-initiated changes, providing instant feedback.

### Step 3.1: Update useTasks.ts updateTask Function

**File:** `ui/src/hooks/useTasks.ts`

**Find the updateTask function (around line 150-156):**
```typescript
const updateTask = useCallback(
  async (id: string, data: Partial<Task>): Promise<Task> => {
    return pb.collection('tasks').update<Task>(id, data)
  },
  []
)
```

**Replace with:**
```typescript
const updateTask = useCallback(
  async (id: string, data: Partial<Task>): Promise<Task> => {
    // Optimistic update - update local state immediately
    setTasks((prev) =>
      prev.map((t) => (t.id === id ? { ...t, ...data, updated: new Date().toISOString() } : t))
        .sort((a, b) => a.position - b.position)
    )
    
    try {
      // Send to server
      const updated = await pb.collection('tasks').update<Task>(id, data)
      
      // Update with server response (includes computed fields like 'updated' timestamp)
      setTasks((prev) =>
        prev.map((t) => (t.id === id ? updated : t))
      )
      
      return updated
    } catch (err) {
      // Rollback on error - refetch all tasks
      console.error('[useTasks] Update failed, refetching:', err)
      const options: { sort: string; filter?: string } = { sort: 'position' }
      if (boardId) {
        options.filter = `board = "${boardId}"`
      }
      const records = await pb.collection('tasks').getFullList<Task>(options)
      setTasks(records)
      throw err
    }
  },
  [boardId]
)
```

### Step 3.2: Update moveTask Function

**Find the moveTask function (around line 164-172):**
```typescript
const moveTask = useCallback(
  async (id: string, column: Column, position: number): Promise<Task> => {
    return pb.collection('tasks').update<Task>(id, {
      column,
      position,
    })
  },
  []
)
```

**Replace with:**
```typescript
const moveTask = useCallback(
  async (id: string, column: Column, position: number): Promise<Task> => {
    // Optimistic update
    setTasks((prev) =>
      prev.map((t) => (t.id === id ? { ...t, column, position } : t))
        .sort((a, b) => a.position - b.position)
    )
    
    try {
      const updated = await pb.collection('tasks').update<Task>(id, {
        column,
        position,
      })
      
      // Sync with server response
      setTasks((prev) =>
        prev.map((t) => (t.id === id ? updated : t))
      )
      
      return updated
    } catch (err) {
      // Rollback - refetch
      console.error('[useTasks] Move failed, refetching:', err)
      const options: { sort: string; filter?: string } = { sort: 'position' }
      if (boardId) {
        options.filter = `board = "${boardId}"`
      }
      const records = await pb.collection('tasks').getFullList<Task>(options)
      setTasks(records)
      throw err
    }
  },
  [boardId]
)
```

### Step 3.3: Update deleteTask Function

**Find the deleteTask function (around line 158-161):**
```typescript
const deleteTask = useCallback(async (id: string): Promise<void> => {
  await pb.collection('tasks').delete(id)
}, [])
```

**Replace with:**
```typescript
const deleteTask = useCallback(async (id: string): Promise<void> => {
  // Store task for potential rollback
  const taskToDelete = tasks.find((t) => t.id === id)
  
  // Optimistic delete
  setTasks((prev) => prev.filter((t) => t.id !== id))
  
  try {
    await pb.collection('tasks').delete(id)
  } catch (err) {
    // Rollback - add task back
    console.error('[useTasks] Delete failed, rolling back:', err)
    if (taskToDelete) {
      setTasks((prev) => [...prev, taskToDelete].sort((a, b) => a.position - b.position))
    }
    throw err
  }
}, [tasks])
```

**Note:** The dependency array now includes `tasks` to access the task for rollback.

---

## Phase 4: SSE Connection Monitoring

**Goal:** Add visibility into the SSE connection state for debugging and user feedback.

### Step 4.1: Create SSE Monitor Hook

**Create new file:** `ui/src/hooks/useRealtimeStatus.ts`

```typescript
import { useState, useEffect } from 'react'
import { pb } from '../lib/pb'

export interface RealtimeStatus {
  isConnected: boolean
  clientId: string | null
  lastConnected: Date | null
  reconnectCount: number
}

/**
 * Hook to monitor PocketBase real-time connection status.
 * Useful for debugging and showing connection state to users.
 */
export function useRealtimeStatus(): RealtimeStatus {
  const [status, setStatus] = useState<RealtimeStatus>({
    isConnected: false,
    clientId: null,
    lastConnected: null,
    reconnectCount: 0,
  })

  useEffect(() => {
    // Subscribe to connection events
    const handleConnect = (clientId: string) => {
      console.log('[Realtime] Connected with client ID:', clientId)
      setStatus((prev) => ({
        isConnected: true,
        clientId,
        lastConnected: new Date(),
        reconnectCount: prev.lastConnected ? prev.reconnectCount + 1 : 0,
      }))
    }

    // PocketBase sends PB_CONNECT event when SSE connection is established
    pb.realtime.subscribe('PB_CONNECT', handleConnect)

    // Handle disconnect
    const originalOnDisconnect = pb.realtime.onDisconnect
    pb.realtime.onDisconnect = (activeSubscriptions) => {
      console.log('[Realtime] Disconnected. Active subscriptions:', activeSubscriptions)
      setStatus((prev) => ({
        ...prev,
        isConnected: false,
      }))
      originalOnDisconnect?.(activeSubscriptions)
    }

    // Check initial state
    if (pb.realtime.clientId) {
      setStatus({
        isConnected: true,
        clientId: pb.realtime.clientId,
        lastConnected: new Date(),
        reconnectCount: 0,
      })
    }

    return () => {
      pb.realtime.unsubscribe('PB_CONNECT')
    }
  }, [])

  return status
}
```

### Step 4.2: Add Connection Indicator (Optional)

**Create new file:** `ui/src/components/ConnectionStatus.tsx`

```typescript
import { useRealtimeStatus } from '../hooks/useRealtimeStatus'
import styles from './ConnectionStatus.module.css'

export function ConnectionStatus() {
  const { isConnected, reconnectCount } = useRealtimeStatus()

  // Only show if disconnected or recently reconnected
  if (isConnected && reconnectCount === 0) {
    return null
  }

  return (
    <div className={`${styles.status} ${isConnected ? styles.connected : styles.disconnected}`}>
      <span className={styles.dot} />
      <span className={styles.text}>
        {isConnected 
          ? `Reconnected (${reconnectCount})`
          : 'Connecting...'}
      </span>
    </div>
  )
}
```

**Create:** `ui/src/components/ConnectionStatus.module.css`

```css
.status {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.connected {
  background: var(--color-success-bg, #1a2e1a);
  color: var(--color-success, #4ade80);
}

.connected .dot {
  background: var(--color-success, #4ade80);
}

.disconnected {
  background: var(--color-warning-bg, #2e2a1a);
  color: var(--color-warning, #facc15);
}

.disconnected .dot {
  background: var(--color-warning, #facc15);
  animation: pulse 1s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}
```

### Step 4.3: Add to Layout (Optional)

In `ui/src/components/Layout.tsx`, import and add the component:

```typescript
import { ConnectionStatus } from './ConnectionStatus'

// In the header area:
<ConnectionStatus />
```

---

## Testing Procedures

### Test 1: Verify SSE Events Are Received

**Setup:**
1. Ensure Phase 1 logging is implemented
2. Start server: `./egenskriven serve`
3. Open browser to `http://localhost:8090`
4. Open DevTools Console

**Test Steps:**
1. Create task via CLI:
   ```bash
   ./egenskriven add "CLI Test Task" --verbose
   ```
2. Check console for `[useTasks] ========== EVENT RECEIVED ==========`

**Expected:** Event should be logged with action `create`

**If No Events:**
- Check Network tab for `/api/realtime` SSE connection
- Verify `[useTasks] Subscription established successfully` appears
- Check for errors in console

### Test 2: Verify Task Appears After CLI Create

**Setup:**
1. Ensure Phase 2 fix is implemented
2. Start server and open browser

**Test Steps:**
1. Note current task count in UI
2. Create task via CLI:
   ```bash
   ./egenskriven add "Should Appear Immediately"
   ```
3. Observe UI without refreshing

**Expected:** Task appears in backlog column immediately

### Test 3: Verify Cross-Board Movement

**Setup:**
1. Create two boards via CLI or UI
2. Create a task in Board A

**Test Steps:**
1. Open Board A in browser
2. Via CLI, move task to Board B:
   ```bash
   ./egenskriven update <task-id> --board "Board B"
   ```
3. Observe: Task should disappear from Board A view
4. Switch to Board B in UI
5. Task should be visible

**Expected:** Task moves between boards in real-time

### Test 4: Verify Optimistic Updates

**Setup:**
1. Ensure Phase 3 is implemented
2. Open browser with DevTools Network tab

**Test Steps:**
1. Select a task
2. Press `S` to open status picker
3. Change status to "In Progress"
4. Observe: Task should move BEFORE network request completes

**Expected:** 
- Task moves immediately (optimistic)
- Network request fires in background
- No visible delay

### Test 5: Verify Update via CLI Reflects in UI

**Test Steps:**
1. Open UI with a task visible
2. Via CLI, update the task:
   ```bash
   ./egenskriven update <task-id> --priority urgent
   ```
3. Observe task card in UI

**Expected:** Priority indicator changes to urgent immediately

### Test 6: Verify Delete via CLI Reflects in UI

**Test Steps:**
1. Open UI with a task visible
2. Note the task ID
3. Via CLI, delete the task:
   ```bash
   ./egenskriven delete <task-id> --force
   ```
4. Observe UI

**Expected:** Task disappears immediately without refresh

### Test 7: Stress Test - Rapid CLI Operations

**Test Steps:**
1. Open UI
2. Run rapid CLI commands:
   ```bash
   for i in {1..5}; do ./egenskriven add "Rapid $i"; done
   ```
3. Observe UI

**Expected:** All 5 tasks appear in UI (may appear rapidly one after another)

---

## Appendix: Code References

### File Locations

| File | Purpose |
|------|---------|
| `ui/src/hooks/useTasks.ts` | Task state management and real-time subscriptions |
| `ui/src/hooks/useBoards.ts` | Board state management and real-time subscriptions |
| `ui/src/hooks/useViews.ts` | Saved views state management |
| `ui/src/lib/pb.ts` | PocketBase client singleton |
| `internal/commands/client.go` | CLI HTTP API client |
| `internal/commands/add.go` | Task creation with hybrid approach |
| `internal/commands/root.go` | CLI global flags (`--direct`, `--verbose`) |
| `cmd/egenskriven/main.go` | Server setup and hooks |
| `migrations/1_initial.go` | Tasks collection with API rules |
| `migrations/7_boards_api_rules.go` | Boards collection API rules |

### Collection API Rules

Both `tasks` and `boards` collections have public access:

```go
collection.ListRule = func() *string { s := ""; return &s }()   // "" = public
collection.ViewRule = func() *string { s := ""; return &s }()   // "" = public
collection.CreateRule = func() *string { s := ""; return &s }() // "" = public
collection.UpdateRule = func() *string { s := ""; return &s }() // "" = public
collection.DeleteRule = func() *string { s := ""; return &s }() // "" = public
```

### CLI Hybrid Functions

**saveRecordHybrid** (in `add.go`):
```go
func saveRecordHybrid(app *pocketbase.PocketBase, record *core.Record, out *output.Formatter) error {
    if isDirectMode() {
        return app.Save(record)
    }
    
    client := NewAPIClient()
    if client.IsServerRunning() {
        taskData := recordToTaskData(record)
        _, err := client.CreateTask(taskData)
        if err != nil {
            if apiErr, ok := IsAPIError(err); ok && apiErr.IsValidationError() {
                return fmt.Errorf("validation error: %s", apiErr.Message)
            }
            // Fallback to direct
            return app.Save(record)
        }
        return nil
    }
    
    return app.Save(record)
}
```

### PocketBase SDK Subscription

From the JS SDK:
```typescript
// Subscribe to all changes
pb.collection('tasks').subscribe('*', callback)

// Subscribe to single record
pb.collection('tasks').subscribe('RECORD_ID', callback)

// Unsubscribe
pb.collection('tasks').unsubscribe('*')
```

**Event Object Structure:**
```typescript
interface RecordSubscription<T> {
  action: 'create' | 'update' | 'delete'
  record: T  // Full record data
}
```

### Vite Proxy Configuration

**File:** `ui/vite.config.ts`

```typescript
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8090',
      changeOrigin: true,
    },
    '/_': {
      target: 'http://localhost:8090',
      changeOrigin: true,
    },
  },
}
```

---

## Implementation Checklist

### Phase 1: Diagnostic Logging
- [ ] Add logging to `useTasks.ts` subscription
- [ ] Add SSE connection logging to `pb.ts`
- [ ] Test and verify events are received
- [ ] Document findings

### Phase 2: Fix Board Movement Bug
- [ ] Update `useTasks.ts` update handler
- [ ] Update `useBoards.ts` update handler
- [ ] Update `useViews.ts` update handler
- [ ] Test cross-board movement

### Phase 3: Optimistic Updates
- [ ] Update `updateTask` function
- [ ] Update `moveTask` function
- [ ] Update `deleteTask` function
- [ ] Test UI responsiveness

### Phase 4: SSE Monitoring
- [ ] Create `useRealtimeStatus` hook
- [ ] Create `ConnectionStatus` component (optional)
- [ ] Add to layout (optional)
- [ ] Test reconnection scenarios

### Final Testing
- [ ] Run all test procedures
- [ ] Verify CLI → UI real-time works
- [ ] Verify UI → UI real-time works (multi-tab)
- [ ] Remove debug logging (or make conditional)
- [ ] Update documentation

---

## Notes

### Why curl Works But CLI Didn't (Historical)

Initially, the CLI used direct database access (`app.Save()`), which only triggered hooks in the CLI process. The web server's `SubscriptionsBroker` never received these events.

The hybrid approach was implemented to have the CLI make HTTP API calls to the running server, ensuring events are broadcast to all connected clients.

### PocketBase SSE vs WebSocket

PocketBase uses SSE (Server-Sent Events), not WebSocket. Key differences:
- SSE is unidirectional (server → client)
- SSE automatically reconnects
- SSE works through HTTP/1.1 proxies
- SSE connection shows as `/api/realtime` in Network tab

### React Strict Mode

`pb.autoCancellation(false)` is set to prevent issues with React Strict Mode's double-mounting in development.

---

*Document maintained as part of EgenSkriven development. Last updated: January 5, 2026*
