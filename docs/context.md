# EgenSkriven UI Issues - Debugging Context

This document contains detailed information about issues discovered during E2E testing of the web UI. Each issue includes reproduction steps, relevant code references, and debugging suggestions.

---

## Issue 1: ViewsSidebar Not Rendering - RESOLVED

**Resolution**: The Go binary with embedded UI wasn't rebuilt after the code change. Always rebuild after UI changes:
```bash
cd ui && npm run build && cd .. && CGO_ENABLED=0 go build -o egenskriven ./cmd/egenskriven
```

---

## Issue 2: Real-time CLI Subscriptions Not Working - RESOLVED (Dev Mode Only)

### Description
Tasks created via the CLI didn't automatically appear in the web UI. The UI required a page refresh to see CLI-created tasks.

### Root Cause
The CLI and web server ran as **separate processes**, each with their own `SubscriptionsBroker`. The CLI's direct database writes never triggered SSE events to the browser.

### Solution Implemented
**Hybrid CLI approach**: CLI tries HTTP API first (enables real-time), falls back to direct database access if server isn't running.

| Server Status | CLI Behavior | Real-time Updates |
|---------------|--------------|-------------------|
| Running | Uses HTTP API | Yes |
| Offline | Falls back to direct DB | No (but CLI still works) |

### What Was Implemented

1. **`internal/commands/client.go`** - API client for HTTP requests
2. **`internal/commands/root.go`** - Added `--direct` and `--verbose` flags
3. **Modified CLI commands** (`add.go`, `update.go`, `delete.go`, `move.go`) to use hybrid approach

### UI Fixes Implemented

1. **Diagnostic logging** in `useTasks.ts` (enabled via `VITE_DEBUG_REALTIME=true`)
2. **Bug fix for tasks moving INTO a board** - `update` handler now adds tasks that weren't previously in state
3. **Optimistic updates** for `updateTask`, `moveTask`, and `deleteTask` functions

### Verification (Dev Server - Working)
```bash
# Terminal 1: Start dev server
cd ui && VITE_DEBUG_REALTIME=true npm run dev

# Terminal 2: Start backend
./egenskriven serve

# Terminal 3: Create task
./egenskriven add "Test task" --verbose
# Output: [verbose] Server is running, using HTTP API for real-time updates

# Browser console shows:
# [useTasks] ========== EVENT RECEIVED ==========
# [useTasks] Action: create
# [useTasks] Adding new task to state
# Task appears immediately in UI without refresh
```

---

## Issue 3: Peek Preview Shows Stale Data - RESOLVED (Dev Mode Only)

### Description
After changing a task's properties (status/priority/type), the peek preview showed old values.

### Solution Implemented
**Optimistic updates** in `useTasks.ts` - local state is updated immediately before the server request completes.

### Verification (Dev Server - Working)
1. Select a task, press `S` to change status
2. Task card moves immediately to new column
3. Press `Space` for peek preview - shows correct updated status

---

## Issue 4: Real-time Not Working in Production Build (CURRENT - UNDER INVESTIGATION)

### Description
Real-time updates work correctly in the **Vite dev server** (port 5173) but do **NOT** work in the **production build** embedded in the Go binary (port 8090).

### Expected Behavior
Tasks created via CLI should appear in the UI immediately, just like they do with the dev server.

### Actual Behavior
- **Dev server (port 5173)**: Real-time works - tasks appear immediately
- **Production server (port 8090)**: Real-time does NOT work - requires page refresh

### Test Results (January 5, 2026)

**Dev Server Test:**
```bash
# Start dev server with debug logging
cd ui && VITE_DEBUG_REALTIME=true npm run dev

# Create task via CLI
./egenskriven add "Test task" --verbose

# Result: Task appears immediately in UI
# Console shows: [useTasks] Adding new task to state
```

**Production Server Test:**
```bash
# Rebuild and start production server
cd ui && npm run build
cd .. && CGO_ENABLED=0 go build -o egenskriven ./cmd/egenskriven
./egenskriven serve

# Create task via CLI
./egenskriven add "Test task" --verbose

# Result: Task does NOT appear until page refresh
# CLI confirms: [verbose] Task created via API (real-time updates enabled)
```

---

### Deep Investigation Results (January 5, 2026 - 16:00)

#### What Was Verified Using Playwright Browser Testing

| Check | Dev Mode (5173) | Production (8090) | Status |
|-------|-----------------|-------------------|--------|
| SSE connection established | ✅ Yes | ✅ Yes | Both work |
| GET `/api/realtime` returns 200 | ✅ Yes | ✅ Yes | Both work |
| POST `/api/realtime` subscription returns 204 | ✅ Yes | ✅ Yes | Both work |
| CLI uses HTTP API | ✅ Yes | ✅ Yes | Both work |
| Task created in database | ✅ Yes | ✅ Yes | Both work |
| SSE event received by browser | ✅ Yes | ❌ **NO** | **BROKEN** |
| Task appears without refresh | ✅ Yes | ❌ **NO** | **BROKEN** |

#### Console Log Comparison

**Dev Mode Console (working):**
```
[useTasks] ========== EVENT RECEIVED ==========
[useTasks] Action: create
[useTasks] Record ID: 9yst986ze3usc5v
[useTasks] Record Title: DEV SERVER REALTIME TEST 16:01:20
[useTasks] Record Board: ljb4ihazichxgx6
[useTasks] Matches Board: true
[useTasks] Adding new task to state
[useTasks] New task count: 37
```

**Production Mode Console:**
*(No events received - debug logging disabled, but even with manual checks, no events arrive)*

#### Production Bundle Verification

The production bundle (`ui/dist/assets/index-*.js`) was verified to contain:
- ✅ PocketBase SDK realtime code
- ✅ `subscribe` method calls
- ✅ `/api/realtime` endpoint references
- ✅ EventSource handling code

```bash
# Verified subscription code exists in bundle:
grep -c 'subscribe' ui/dist/assets/index-Cio8-ywD.js  # Returns: 1
grep -o '"/api/realtime"' ui/dist/assets/index-Cio8-ywD.js  # Returns: "/api/realtime" (twice)
```

#### SSE Connection Test via curl

```bash
# SSE connection establishes correctly:
curl -N -H "Accept: text/event-stream" http://127.0.0.1:8090/api/realtime

# Output:
id:FqffxMzXfTJBl0DJELIHhPbQi2N97ed7nstyWWxk
event:PB_CONNECT
data:{"clientId":"FqffxMzXfTJBl0DJELIHhPbQi2N97ed7nstyWWxk"}

# Note: PB_CONNECT event is received, proving SSE works at protocol level
```

---

### Root Cause Analysis

#### Confirmed Facts
1. **SSE connection works** - Both dev and prod establish SSE connections successfully
2. **Subscriptions are sent** - POST to `/api/realtime` returns 204 in both modes
3. **Events are generated** - Server creates events when CLI adds tasks
4. **Events reach dev mode** - Console logs prove events arrive in dev
5. **Events DON'T reach production** - Despite identical network setup

#### Most Likely Causes (Ranked by Probability)

##### 1. Browser HTTP/1.1 Connection Limits (HIGH PROBABILITY)
When both UI and API are served from the same origin (port 8090), browsers limit connections to ~6 per origin for HTTP/1.1. If the app has many concurrent requests, the SSE connection might get queued, interrupted, or multiplexed incorrectly.

**Evidence:**
- In dev mode, Vite (port 5173) is a DIFFERENT origin from API (port 8090)
- In production, both are SAME origin (port 8090)
- This creates connection contention in production only

##### 2. Duplicate Subscriptions Causing Conflicts (MEDIUM PROBABILITY)
Found that `useTasks` hook is called in TWO places:
- `ui/src/App.tsx` line 35: `const { tasks, loading, createTask, updateTask, deleteTask } = useTasks(currentBoard?.id)`
- `ui/src/components/Board.tsx` line 40: `const { tasks: allTasks, loading: tasksLoading, error, moveTask } = useTasks(currentBoard?.id)`

This creates **duplicate SSE subscriptions** to the same collection. While React handles this in dev mode, production bundling/optimization might cause race conditions.

##### 3. Response Buffering Difference (MEDIUM PROBABILITY)
The Go server embedded file serving might interact differently with SSE streaming compared to Vite's proxy. While PocketBase sets proper headers (`X-Accel-Buffering: no`), there could be subtle buffering differences.

##### 4. EventSource Reconnection Timing (LOW PROBABILITY)
The browser's EventSource might be reconnecting with a new clientId after initial subscriptions are sent, causing events to go to the old connection.

##### 5. Service Worker Interference (LOW PROBABILITY)
If Vite's build includes a service worker, it might be intercepting/caching SSE requests incorrectly. (Need to verify if service worker is present)

---

### Recommended Solutions

#### Option A: Enable HTTP/2 (Fixes Connection Limit Issue)
HTTP/2 multiplexes all connections over a single TCP connection, eliminating browser connection limits.

```bash
# Generate self-signed cert for testing
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes

# Start with HTTPS (enables HTTP/2)
./egenskriven serve --http=0.0.0.0:443 --cert=cert.pem --key=key.pem
```

**Pros:** Fixes root cause if it's connection limits
**Cons:** Requires HTTPS setup, self-signed certs cause browser warnings

#### Option B: Remove Duplicate useTasks Subscription (Quick Fix)
Remove the redundant `useTasks` call in `Board.tsx`:

```typescript
// Current (Board.tsx line 40):
const { tasks: allTasks, loading: tasksLoading, error, moveTask } = useTasks(currentBoard?.id)

// Change to:
const { moveTask } = useTasks(currentBoard?.id)
// OR remove entirely and pass moveTask as prop from App.tsx
```

**Pros:** Simple code change, reduces subscription complexity
**Cons:** Might not be the root cause

#### Option C: Enable Production Debug Logging (Diagnostic)
Modify `pb.ts` to enable logging in production for diagnosis:

```typescript
// In ui/src/lib/pb.ts, change:
const DEBUG_REALTIME = import.meta.env.VITE_DEBUG_REALTIME === 'true'

// To:
const DEBUG_REALTIME = true  // Force enable for debugging
// OR:
const DEBUG_REALTIME = localStorage.getItem('DEBUG_REALTIME') === 'true'
```

Then rebuild and test to see if events are received but not processed.

**Pros:** Gives visibility into production behavior
**Cons:** Adds console noise, not a fix

#### Option D: Add Polling Fallback (Workaround)
Add periodic polling as fallback when SSE fails:

```typescript
// In useTasks.ts, add:
useEffect(() => {
  const pollInterval = setInterval(async () => {
    const records = await pb.collection('tasks').getFullList<Task>({ sort: 'position' })
    setTasks(records)
  }, 30000) // Poll every 30 seconds
  
  return () => clearInterval(pollInterval)
}, [])
```

**Pros:** Ensures eventual consistency even if SSE fails
**Cons:** Not true real-time, adds server load

#### Option E: Test with Vite Preview (Diagnostic)
Test the production build served by Vite instead of Go:

```bash
cd ui && npm run build && npm run preview
# Open http://localhost:4173 and test with CLI
```

If this works, the issue is with Go's static file serving.
If this fails, the issue is with the production build itself.

---

### Recommended Next Steps

1. **First:** Try Option E (Vite Preview) - This will narrow down if the issue is Go serving or production build
2. **Second:** Implement Option B (Remove duplicate subscription) - Low risk, might fix it
3. **Third:** Implement Option C (Debug logging) - Get production diagnostics
4. **Fourth:** If still broken, implement Option A (HTTP/2) or Option D (Polling fallback)

---

### Key Differences: Dev vs Production

| Aspect | Dev Server (5173) | Production (8090) |
|--------|-------------------|-------------------|
| Build mode | Development | Production |
| Source maps | Full | Minified |
| Hot reload | Yes | No |
| Debug logging | Enabled via env var | Not available |
| Origin separation | UI on 5173, API on 8090 | Same origin (8090) |
| Connection pooling | Separate per origin | Shared (6 conn limit) |
| SSE events | ✅ Received | ❌ NOT received |
| API calls | ✅ Work | ✅ Work |
| Initial data load | ✅ Works | ✅ Works |

### Files Involved

| File | Purpose | Investigation Notes |
|------|---------|---------------------|
| `ui/vite.config.ts` | Build configuration | Proxy config only affects dev mode |
| `ui/src/lib/pb.ts` | PocketBase client | Initialized with `'/'` as base URL |
| `ui/src/hooks/useTasks.ts` | Subscription setup | Called from 2 components (potential issue) |
| `ui/src/App.tsx` | Main app component | Calls useTasks at line 35 |
| `ui/src/components/Board.tsx` | Board component | Also calls useTasks at line 40 (duplicate!) |
| `ui/embed.go` | Embeds UI in Go binary | Simple embed, unlikely to be issue |
| `cmd/egenskriven/main.go` | Go server setup | Standard PocketBase setup |

### Workaround (Until Fixed)

Use the **dev server** for real-time functionality:
```bash
# Terminal 1: Backend
./egenskriven serve

# Terminal 2: Frontend (with real-time working)
cd ui && npm run dev
# Open http://localhost:5173 instead of :8090
```

---

## Architecture Reference

### How Real-time Works

```
Browser ──SSE──► PocketBase Server ◄──HTTP API── CLI
                      │
                      ▼
                  SQLite DB
```

1. Browser establishes SSE connection to `/api/realtime`
2. Browser subscribes to `tasks` collection changes
3. CLI creates task via HTTP API to same server
4. Server broadcasts event to all SSE clients
5. Browser receives event and updates React state

### Key Files

| File | Purpose |
|------|---------|
| `ui/src/hooks/useTasks.ts` | Task state management and real-time subscriptions |
| `ui/src/lib/pb.ts` | PocketBase client singleton |
| `internal/commands/client.go` | CLI HTTP API client |
| `internal/commands/add.go` | Hybrid task creation (API + fallback) |

### Useful Commands

```bash
# Rebuild everything after UI changes
cd ui && npm run build && cd .. && CGO_ENABLED=0 go build -o egenskriven ./cmd/egenskriven

# Run dev server with debug logging
cd ui && VITE_DEBUG_REALTIME=true npm run dev

# Test CLI with verbose output
./egenskriven add "Test" --verbose

# Check SSE connection
curl -N http://localhost:8090/api/realtime
```

---

*Last updated: January 5, 2026 - 16:10 (Deep investigation completed)*
