# Phase 2: Minimal UI

**Goal**: Basic web UI with board view, drag-and-drop, and real-time sync with CLI changes.

**Duration Estimate**: 5-7 days

**Prerequisites**: Phase 1 complete (Core CLI with working task CRUD operations).

**Deliverable**: A functional React-based kanban board embedded in the Go binary, showing tasks created via CLI and allowing drag-and-drop between columns.

---

## Overview

Phase 2 adds the web-based UI to EgenSkriven. The React frontend is embedded directly into the Go binary using `go:embed`, meaning the final distribution remains a single file. Users can run `./egenskriven serve` and access a full kanban board at `http://localhost:8090`.

### Why Embedded React?

- **Single binary distribution**: No separate frontend deployment
- **Offline-first**: Works without internet after initial download
- **Version consistency**: UI and CLI always match
- **Simple deployment**: Copy one file, run it

### How Real-time Sync Works

1. CLI creates/updates task → writes to SQLite via PocketBase
2. PocketBase detects change → broadcasts SSE (Server-Sent Events)
3. React frontend subscribes to changes → receives event
4. UI updates → React state updates, board re-renders

### What We're Building

| Component | Purpose |
|-----------|---------|
| Vite + React + TypeScript | Modern frontend toolchain |
| PocketBase JS SDK | API communication and real-time subscriptions |
| @dnd-kit | Accessible drag-and-drop |
| CSS Design Tokens | Consistent dark-mode styling |
| Embedded UI | Single binary with `go:embed` |

---

## Scope & Deferred Features

Phase 2 focuses on a **minimal but functional** UI. Several features from the full design are intentionally deferred to later phases:

### Included in Phase 2
- Board view with 5 columns
- Task cards with title, ID, priority, type, labels, due date
- Drag-and-drop between columns
- Task detail panel with editable status, priority, type
- Quick create modal (`C` key)
- Selected task state with keyboard navigation (`Enter` to open)
- Real-time sync with CLI changes
- Dark mode styling with design tokens
- Basic component tests

### Deferred to Phase 4 (Interactive UI)
- Command palette (`Cmd+K`)
- Full keyboard navigation (`J/K` for up/down, `H/L` for columns)
- Property picker popovers (inline `S`, `P`, `T`, `L` shortcuts)
- Peek preview (`Space`)
- Keyboard shortcuts help modal (`?`)

### Deferred to Phase 5 (Multi-Board)
- Sidebar with board list
- Board switcher
- Board-specific column configuration

### Deferred to Phase 6 (Filtering & Views)
- Filter bar and filter builder
- Search (`/`)
- List view toggle (`Cmd+B`)
- Saved views

### Deferred to Phase 7 (Polish)
- Light mode
- Accent color customization
- Settings panel
- Responsive/mobile layout
- Toast notifications
- Animations refinements

### Deferred to Phase 8 (Advanced Features)
- Inline title editing
- Markdown description editor with structured sections
- Activity/history display
- Due date picker
- Sub-tasks display
- Epic integration in UI

---

## Environment Requirements

Before starting, ensure you have:

| Tool | Version | Check Command |
|------|---------|---------------|
| Node.js | 18+ | `node --version` |
| npm | 9+ | `npm --version` |
| Go | 1.21+ | `go version` |

**Install Node.js** (if needed):
- macOS: `brew install node`
- Linux: `curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash - && sudo apt-get install -y nodejs`
- Windows: Download installer from https://nodejs.org/

---

## Tasks

### 2.1 Initialize React Project

**What**: Create a new Vite + React + TypeScript project in the `ui/` directory.

**Why**: Vite provides fast development builds and optimized production bundles. TypeScript catches errors at compile time.

**Steps**:

1. Navigate to the ui directory:
   ```bash
   cd ui
   ```

2. Create the Vite project (using current directory):
   ```bash
   npm create vite@latest . -- --template react-ts
   ```
   
   When prompted "Current directory is not empty", select "Ignore files and continue".
   
   **Expected output**:
   ```
   Scaffolding project in /path/to/egenskriven/ui...
   Done. Now run:
     npm install
     npm run dev
   ```

3. Install base dependencies:
   ```bash
   npm install
   ```
   
   **Expected output**:
   ```
   added 200+ packages in Xs
   ```

4. Install additional dependencies:
   ```bash
   npm install pocketbase @dnd-kit/core @dnd-kit/sortable @dnd-kit/utilities
   ```
   
   **Expected output**:
   ```
   added 4 packages in Xs
   ```

5. Verify the project runs:
   ```bash
   npm run dev
   ```
   
   **Expected output**:
   ```
   VITE v5.x.x  ready in XXX ms
   
   ➜  Local:   http://localhost:5173/
   ➜  Network: use --host to expose
   ```

6. Stop the dev server with `Ctrl+C`

**Common Mistakes**:
- Running commands from wrong directory (must be in `ui/`)
- Node.js version too old (requires 18+)
- Existing files conflicting (safe to ignore `embed.go`)

---

### 2.2 Setup Vite Configuration

**What**: Configure Vite to proxy API requests to PocketBase and output to the correct directory.

**Why**: During development, the React dev server runs on port 5173 while PocketBase runs on port 8090. The proxy forwards API requests to PocketBase. The build output goes to `dist/` for embedding.

**File**: `ui/vite.config.ts`

```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  
  build: {
    // Output to dist/ directory (will be embedded in Go binary)
    outDir: 'dist',
    // Clean the output directory before building
    emptyOutDir: true,
  },
  
  server: {
    // Proxy API requests to PocketBase during development
    proxy: {
      // Forward /api/* requests to PocketBase
      '/api': {
        target: 'http://localhost:8090',
        changeOrigin: true,
      },
      // Forward /_/* requests (PocketBase admin UI) to PocketBase
      '/_': {
        target: 'http://localhost:8090',
        changeOrigin: true,
      },
    },
  },
})
```

**Steps**:

1. Open `ui/vite.config.ts` in your editor.

2. Replace the contents with the configuration above.

3. Verify the configuration is valid:
   ```bash
   npm run build
   ```
   
   **Expected output**:
   ```
   vite v5.x.x building for production...
   ✓ X modules transformed.
   dist/index.html                   0.xx kB │ gzip: 0.xx kB
   dist/assets/index-XXXXXXXX.css    X.xx kB │ gzip: X.xx kB
   dist/assets/index-XXXXXXXX.js     XXX.xx kB │ gzip: XX.xx kB
   ✓ built in Xs
   ```

4. Verify `dist/` directory was created:
   ```bash
   ls dist/
   ```
   
   **Expected output**:
   ```
   assets  index.html
   ```

**Key Configuration Explained**:

| Option | Purpose |
|--------|---------|
| `outDir: 'dist'` | Build output location (embedded by Go) |
| `emptyOutDir: true` | Clean build directory before each build |
| `proxy./api` | Forward API requests to PocketBase |
| `proxy./_` | Forward admin UI requests to PocketBase |

---

### 2.3 Update embed.go for Production

**What**: Update the placeholder `embed.go` to embed the built React app.

**Why**: Go's `//go:embed` directive includes files in the binary at compile time. This makes distribution simple - one file contains everything.

**File**: `ui/embed.go`

```go
package ui

import (
	"embed"
	"io/fs"
)

// DistFS holds the embedded React build output.
//
// During development (when dist/ doesn't exist), this will cause a build error.
// Run `cd ui && npm run build` first to create the dist/ directory.
//
// The "all:" prefix includes files starting with "." or "_" which Vite may create.

//go:embed all:dist
var distDir embed.FS

// DistFS is the filesystem containing the built React application.
// It's a sub-filesystem rooted at "dist" for cleaner path handling.
var DistFS, _ = fs.Sub(distDir, "dist")
```

**Steps**:

1. Open `ui/embed.go` in your editor.

2. Replace the entire contents with the code above.

3. Ensure the React app is built (required for embed to work):
   ```bash
   cd ui && npm run build
   ```

4. Verify the Go code compiles:
   ```bash
   cd .. && go build ./ui
   ```
   
   **Expected output**: No output means success.

**Important Notes**:
- The `//go:embed all:dist` directive MUST have no space between `//` and `go:`
- The `dist/` directory MUST exist before building the Go binary
- If you see "pattern dist: no matching files found", run `npm run build` first

---

### 2.4 Update main.go for UI Serving

**What**: Configure PocketBase to serve the embedded React app for non-API routes.

**Why**: PocketBase handles its own routes (`/api/*`, `/_/*`). We need to serve the React SPA for all other routes, with a fallback to `index.html` for client-side routing.

**File**: `cmd/egenskriven/main.go`

Update the existing `main.go` to serve the embedded UI:

```go
package main

import (
	"log"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"github.com/ramtinj/egenskriven/internal/commands"
	"github.com/ramtinj/egenskriven/migrations"
	"github.com/ramtinj/egenskriven/ui"
)

func main() {
	app := pocketbase.New()

	// Register database migrations
	migrations.Register(app)

	// Register CLI commands
	commands.Register(app)

	// Serve embedded React UI for non-API routes
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		// Catch-all route for the React SPA
		// This runs AFTER PocketBase's built-in routes (/api/*, /_/*)
		e.Router.GET("/{path...}", func(re *core.RequestEvent) error {
			path := re.Request.PathValue("path")

			// Skip API and admin routes (handled by PocketBase)
			if strings.HasPrefix(path, "api/") || strings.HasPrefix(path, "_/") {
				return re.Next()
			}

			// Try to serve the exact file from embedded filesystem
			// This handles JS, CSS, images, etc.
			if f, err := ui.DistFS.Open(path); err == nil {
				f.Close()
				return re.FileFS(ui.DistFS, path)
			}

			// For all other paths, serve index.html (SPA client-side routing)
			// This enables React Router to handle /board, /task/123, etc.
			return re.FileFS(ui.DistFS, "index.html")
		})

		return e.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
```

**Steps**:

1. Open `cmd/egenskriven/main.go` in your editor.

2. Replace the contents with the code above.

3. Build the full application:
   ```bash
   cd ui && npm run build && cd .. && make build
   ```

4. Test the server serves the UI:
   ```bash
   ./egenskriven serve
   ```

5. Open `http://localhost:8090` in your browser.
   
   **Expected**: You should see the default Vite React welcome page ("Vite + React").

6. Verify PocketBase admin still works at `http://localhost:8090/_/`

7. Stop the server with `Ctrl+C`

**Route Handling Explained**:

| Route | Handler |
|-------|---------|
| `/api/*` | PocketBase REST API |
| `/_/*` | PocketBase Admin UI |
| `/assets/*` | Static files from embedded UI |
| `/*` | React SPA (index.html) |

---

### 2.5 Update Makefile for UI Development

**What**: Add Make targets for UI development and building.

**Why**: Consistent commands make development easier. Parallel development of Go and React is essential for productivity.

**File**: `Makefile`

Add these targets to your existing Makefile:

```makefile
# =============================================================================
# UI Development
# =============================================================================

# Build the React UI for production
build-ui:
	@echo "Building React UI..."
	cd ui && npm ci && npm run build
	@echo "UI built: ui/dist/"

# Start React development server (port 5173)
# Use this alongside 'make dev' for hot reload on both frontend and backend
dev-ui:
	@echo "Starting React dev server..."
	@echo "Note: Run 'make dev' in another terminal for the Go backend"
	cd ui && npm run dev

# Run UI tests
test-ui:
	@echo "Running UI tests..."
	cd ui && npm test

# Clean UI build artifacts
clean-ui:
	@echo "Cleaning UI artifacts..."
	rm -rf ui/dist ui/node_modules

# =============================================================================
# Combined Development
# =============================================================================

# Development: run both React and Go with hot reload
# This runs both servers in parallel using Make's -j flag
dev-all:
	@echo "Starting development servers..."
	@echo "  React: http://localhost:5173 (with proxy to :8090)"
	@echo "  Go:    http://localhost:8090"
	@$(MAKE) -j2 dev-ui dev

# =============================================================================
# Production Build
# =============================================================================

# Build everything for production
# First builds UI, then embeds it into Go binary
build: build-ui
	@echo "Building Go binary with embedded UI..."
	CGO_ENABLED=0 go build -o egenskriven ./cmd/egenskriven
	@echo "Built: ./egenskriven ($(shell du -h egenskriven | cut -f1))"
```

**Steps**:

1. Open `Makefile` in your editor.

2. Add the new targets above (you can add them after the existing targets).

3. Test the UI build:
   ```bash
   make build-ui
   ```
   
   **Expected output**:
   ```
   Building React UI...
   npm warn ... (possible warnings, okay to ignore)
   
   > ui@0.0.0 build
   > tsc && vite build
   
   vite v5.x.x building for production...
   ✓ X modules transformed.
   ...
   UI built: ui/dist/
   ```

4. Test the full build:
   ```bash
   make build
   ```
   
   **Expected output**:
   ```
   Building React UI...
   ...
   Building Go binary with embedded UI...
   Built: ./egenskriven (35M)
   ```

5. Verify the binary serves the UI:
   ```bash
   ./egenskriven serve
   ```
   
   Open `http://localhost:8090` - you should see the React welcome page.

**Common Mistakes**:
- Makefile using spaces instead of tabs (causes "missing separator" error)
- Running `make build-ui` before `npm install`

---

### 2.6 Setup Design Tokens

**What**: Create CSS custom properties (variables) for consistent styling.

**Why**: Design tokens provide a single source of truth for colors, spacing, and typography. This makes theming and consistency easy.

**File**: `ui/src/styles/tokens.css`

```css
/* =============================================================================
   EgenSkriven Design Tokens
   
   Based on ui-design.md specifications.
   Dark mode is the default; light mode overrides in a separate file.
   ============================================================================= */

:root {
  /* -------------------------------------------------------------------------
     Colors: Backgrounds
     ------------------------------------------------------------------------- */
  --bg-app: #0D0D0D;
  --bg-sidebar: #141414;
  --bg-card: #1A1A1A;
  --bg-card-hover: #252525;
  --bg-card-selected: #2E2E2E;
  --bg-input: #1F1F1F;
  --bg-overlay: rgba(0, 0, 0, 0.6);

  /* -------------------------------------------------------------------------
     Colors: Text
     ------------------------------------------------------------------------- */
  --text-primary: #F5F5F5;
  --text-secondary: #A0A0A0;
  --text-muted: #666666;
  --text-disabled: #444444;

  /* -------------------------------------------------------------------------
     Colors: Borders
     ------------------------------------------------------------------------- */
  --border-subtle: #2A2A2A;
  --border-default: #333333;
  --border-focus: var(--accent);

  /* -------------------------------------------------------------------------
     Colors: Accent (Blue default, can be customized)
     ------------------------------------------------------------------------- */
  --accent: #5E6AD2;
  --accent-hover: #6E7BE2;
  --accent-muted: rgba(94, 106, 210, 0.2);

  /* -------------------------------------------------------------------------
     Colors: Status (Column/task status)
     ------------------------------------------------------------------------- */
  --status-backlog: #6B7280;
  --status-todo: #E5E5E5;
  --status-in-progress: #F59E0B;
  --status-review: #A855F7;
  --status-done: #22C55E;
  --status-canceled: #6B7280;

  /* -------------------------------------------------------------------------
     Colors: Priority
     ------------------------------------------------------------------------- */
  --priority-urgent: #EF4444;
  --priority-high: #F97316;
  --priority-medium: #EAB308;
  --priority-low: #6B7280;
  --priority-none: #444444;

  /* -------------------------------------------------------------------------
     Colors: Task Type
     ------------------------------------------------------------------------- */
  --type-bug: #EF4444;
  --type-feature: #A855F7;
  --type-chore: #6B7280;

  /* -------------------------------------------------------------------------
     Typography
     ------------------------------------------------------------------------- */
  --font-sans: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  --font-mono: 'JetBrains Mono', 'Fira Code', ui-monospace, monospace;

  /* Font sizes */
  --text-xs: 11px;
  --text-sm: 12px;
  --text-base: 13px;
  --text-lg: 14px;
  --text-xl: 16px;
  --text-2xl: 20px;
  --text-3xl: 24px;

  /* Font weights */
  --font-normal: 400;
  --font-medium: 500;
  --font-semibold: 600;

  /* Line heights */
  --leading-tight: 1.25;
  --leading-normal: 1.5;
  --leading-relaxed: 1.625;

  /* -------------------------------------------------------------------------
     Spacing
     ------------------------------------------------------------------------- */
  --space-1: 4px;
  --space-2: 8px;
  --space-3: 12px;
  --space-4: 16px;
  --space-5: 20px;
  --space-6: 24px;
  --space-8: 32px;
  --space-10: 40px;

  /* -------------------------------------------------------------------------
     Border Radius
     ------------------------------------------------------------------------- */
  --radius-sm: 4px;
  --radius-md: 6px;
  --radius-lg: 8px;
  --radius-xl: 12px;

  /* -------------------------------------------------------------------------
     Shadows
     ------------------------------------------------------------------------- */
  --shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.3);
  --shadow-md: 0 4px 6px rgba(0, 0, 0, 0.4);
  --shadow-lg: 0 10px 15px rgba(0, 0, 0, 0.5);
  --shadow-drag: 0 12px 24px rgba(0, 0, 0, 0.6);

  /* -------------------------------------------------------------------------
     Animations
     ------------------------------------------------------------------------- */
  --duration-fast: 100ms;
  --duration-normal: 150ms;
  --duration-slow: 200ms;

  --ease-default: cubic-bezier(0.4, 0, 0.2, 1);
  --ease-in: cubic-bezier(0.4, 0, 1, 1);
  --ease-out: cubic-bezier(0, 0, 0.2, 1);

  /* -------------------------------------------------------------------------
     Layout
     ------------------------------------------------------------------------- */
  --sidebar-width: 240px;
  --column-width: 280px;
  --column-gap: 16px;
  --detail-panel-width: 400px;
}

/* =============================================================================
   Base Styles
   ============================================================================= */

* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

html, body, #root {
  height: 100%;
  width: 100%;
}

body {
  font-family: var(--font-sans);
  font-size: var(--text-base);
  line-height: var(--leading-normal);
  color: var(--text-primary);
  background-color: var(--bg-app);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* Focus visible for accessibility */
:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

/* Remove default focus for mouse users */
:focus:not(:focus-visible) {
  outline: none;
}
```

**Steps**:

1. Create the styles directory:
   ```bash
   mkdir -p ui/src/styles
   ```

2. Create the file `ui/src/styles/tokens.css` with the content above.

3. Import the tokens in your main entry point. Open `ui/src/main.tsx` and add:
   ```typescript
   import './styles/tokens.css'
   ```
   
   The file should look like:
   ```typescript
   import { StrictMode } from 'react'
   import { createRoot } from 'react-dom/client'
   import './styles/tokens.css'
   import App from './App.tsx'

   createRoot(document.getElementById('root')!).render(
     <StrictMode>
       <App />
     </StrictMode>,
   )
   ```

4. Delete the default Vite CSS files (we'll use our own):
   ```bash
   rm ui/src/index.css ui/src/App.css
   ```

5. Remove the CSS import from `ui/src/App.tsx`:
   - Open `ui/src/App.tsx`
   - Remove the line `import './App.css'`

6. Verify the app still runs:
   ```bash
   cd ui && npm run dev
   ```
   
   Open `http://localhost:5173` - you should see the React app with dark background.

---

### 2.7 Create PocketBase Client

**What**: Create a typed PocketBase client for API communication.

**Why**: A centralized client provides consistent API access and TypeScript types for tasks.

**File**: `ui/src/lib/pb.ts`

```typescript
import PocketBase from 'pocketbase'

// Create a single PocketBase client instance
// In production, the UI and API are served from the same origin
// In development, Vite proxies /api requests to localhost:8090
export const pb = new PocketBase('/')

// Disable auto-cancellation to prevent issues with React strict mode
pb.autoCancellation(false)
```

**File**: `ui/src/types/task.ts`

```typescript
import type { RecordModel } from 'pocketbase'

// Task type values
export type TaskType = 'bug' | 'feature' | 'chore'

// Priority levels (ordered from highest to lowest)
export type Priority = 'urgent' | 'high' | 'medium' | 'low'

// Column/status values (ordered by workflow)
export type Column = 'backlog' | 'todo' | 'in_progress' | 'review' | 'done'

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
// All fields align with kanban-architecture.md data model
export interface Task extends RecordModel {
  title: string
  description?: string
  type: TaskType
  priority: Priority
  column: Column
  position: number
  board?: string              // Link to boards collection (multi-board, Phase 5)
  epic?: string               // Optional link to epics collection (Phase 3)
  parent?: string             // Optional parent task for sub-tasks (Phase 8)
  labels?: string[]
  blocked_by?: string[]       // Array of task IDs that block this task
  due_date?: string           // Optional due date (Phase 8)
  created_by?: CreatedBy
  created_by_agent?: string   // Agent identifier (e.g., "claude", "opencode")
  history?: HistoryEntry[]    // Activity tracking array
}

// All possible columns in display order
export const COLUMNS: Column[] = [
  'backlog',
  'todo',
  'in_progress',
  'review',
  'done',
]

// Human-readable column names
export const COLUMN_NAMES: Record<Column, string> = {
  backlog: 'Backlog',
  todo: 'Todo',
  in_progress: 'In Progress',
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
```

**Steps**:

1. Create the lib and types directories:
   ```bash
   mkdir -p ui/src/lib ui/src/types
   ```

2. Create `ui/src/lib/pb.ts` with the content above.

3. Create `ui/src/types/task.ts` with the content above.

4. Verify TypeScript is happy:
   ```bash
   cd ui && npm run build
   ```
   
   **Expected**: Build succeeds without type errors.

---

### 2.8 Create PocketBase Hooks

**What**: Create React hooks for fetching tasks and subscribing to real-time updates.

**Why**: Hooks encapsulate data fetching logic and make components cleaner. Real-time subscriptions keep the UI in sync with CLI changes.

**File**: `ui/src/hooks/useTasks.ts`

```typescript
import { useEffect, useState, useCallback } from 'react'
import { pb } from '../lib/pb'
import type { Task, Column } from '../types/task'

interface UseTasksReturn {
  tasks: Task[]
  loading: boolean
  error: Error | null
  createTask: (title: string, column?: Column) => Promise<Task>
  updateTask: (id: string, data: Partial<Task>) => Promise<Task>
  deleteTask: (id: string) => Promise<void>
  moveTask: (id: string, column: Column, position: number) => Promise<Task>
}

/**
 * Hook for managing tasks with real-time updates.
 * 
 * Features:
 * - Fetches all tasks on mount
 * - Subscribes to real-time create/update/delete events
 * - Provides CRUD operations
 * - Automatically updates local state on changes
 * 
 * @example
 * ```tsx
 * function Board() {
 *   const { tasks, loading, moveTask } = useTasks()
 *   
 *   if (loading) return <div>Loading...</div>
 *   
 *   return <BoardView tasks={tasks} onMove={moveTask} />
 * }
 * ```
 */
export function useTasks(): UseTasksReturn {
  const [tasks, setTasks] = useState<Task[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  // Fetch all tasks on mount
  useEffect(() => {
    const fetchTasks = async () => {
      try {
        const records = await pb.collection('tasks').getFullList<Task>({
          sort: 'position',
        })
        setTasks(records)
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Failed to fetch tasks'))
      } finally {
        setLoading(false)
      }
    }

    fetchTasks()
  }, [])

  // Subscribe to real-time updates
  useEffect(() => {
    // Subscribe to all task changes
    pb.collection('tasks').subscribe<Task>('*', (event) => {
      switch (event.action) {
        case 'create':
          // Add new task to state
          setTasks((prev) => [...prev, event.record])
          break
          
        case 'update':
          // Replace updated task in state
          setTasks((prev) =>
            prev.map((t) => (t.id === event.record.id ? event.record : t))
          )
          break
          
        case 'delete':
          // Remove deleted task from state
          setTasks((prev) => prev.filter((t) => t.id !== event.record.id))
          break
      }
    })

    // Cleanup subscription on unmount
    return () => {
      pb.collection('tasks').unsubscribe('*')
    }
  }, [])

  // Create a new task
  const createTask = useCallback(
    async (title: string, column: Column = 'backlog'): Promise<Task> => {
      // Get next position in column
      const columnTasks = tasks.filter((t) => t.column === column)
      const maxPosition = columnTasks.reduce(
        (max, t) => Math.max(max, t.position),
        0
      )
      const position = maxPosition + 1000

      const task = await pb.collection('tasks').create<Task>({
        title,
        column,
        position,
        type: 'feature',
        priority: 'medium',
        labels: [],
        created_by: 'user',
      })

      return task
    },
    [tasks]
  )

  // Update a task
  const updateTask = useCallback(
    async (id: string, data: Partial<Task>): Promise<Task> => {
      return pb.collection('tasks').update<Task>(id, data)
    },
    []
  )

  // Delete a task
  const deleteTask = useCallback(async (id: string): Promise<void> => {
    await pb.collection('tasks').delete(id)
  }, [])

  // Move a task to a new column/position
  const moveTask = useCallback(
    async (id: string, column: Column, position: number): Promise<Task> => {
      return pb.collection('tasks').update<Task>(id, {
        column,
        position,
      })
    },
    []
  )

  return {
    tasks,
    loading,
    error,
    createTask,
    updateTask,
    deleteTask,
    moveTask,
  }
}
```

**Steps**:

1. Create the hooks directory:
   ```bash
   mkdir -p ui/src/hooks
   ```

2. Create `ui/src/hooks/useTasks.ts` with the content above.

3. Verify TypeScript compiles:
   ```bash
   cd ui && npm run build
   ```

---

### 2.9 Create Layout Components

**What**: Create the main layout shell and header components.

**Why**: A consistent layout structure makes the app feel cohesive. The header provides navigation and actions.

**File**: `ui/src/components/Layout.tsx`

```tsx
import type { ReactNode } from 'react'
import { Header } from './Header'
import styles from './Layout.module.css'

interface LayoutProps {
  children: ReactNode
}

/**
 * Main application layout.
 * 
 * Structure:
 * - Header: App title and actions
 * - Main: Content area (board/list view)
 * 
 * Note: Sidebar will be added in Phase 5 (Multi-board support)
 */
export function Layout({ children }: LayoutProps) {
  return (
    <div className={styles.layout}>
      <Header />
      <main className={styles.main}>{children}</main>
    </div>
  )
}
```

**File**: `ui/src/components/Layout.module.css`

```css
.layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
  background-color: var(--bg-app);
}

.main {
  flex: 1;
  overflow: hidden;
  display: flex;
}
```

**File**: `ui/src/components/Header.tsx`

```tsx
import styles from './Header.module.css'

/**
 * Application header with title and shortcuts hint.
 */
export function Header() {
  return (
    <header className={styles.header}>
      <div className={styles.title}>
        <span className={styles.logo}>EgenSkriven</span>
      </div>
      
      <div className={styles.actions}>
        <span className={styles.shortcut}>
          <kbd>C</kbd> Create
        </span>
        <span className={styles.shortcut}>
          <kbd>Enter</kbd> Open
        </span>
        <span className={styles.shortcut}>
          <kbd>Esc</kbd> Close
        </span>
      </div>
    </header>
  )
}
```

**File**: `ui/src/components/Header.module.css`

```css
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 48px;
  padding: 0 var(--space-4);
  background-color: var(--bg-sidebar);
  border-bottom: 1px solid var(--border-subtle);
  flex-shrink: 0;
}

.title {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.logo {
  font-size: var(--text-lg);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.actions {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.shortcut {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  font-size: var(--text-sm);
  color: var(--text-muted);
}

.shortcut kbd {
  display: inline-block;
  padding: 2px 6px;
  font-family: var(--font-mono);
  font-size: var(--text-xs);
  background-color: var(--bg-card);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
}
```

**Steps**:

1. Create the components directory:
   ```bash
   mkdir -p ui/src/components
   ```

2. Create all four files above.

3. Verify there are no TypeScript errors:
   ```bash
   cd ui && npm run build
   ```

---

### 2.10 Create Board Components

**What**: Create the kanban board with columns and task cards.

**Why**: This is the core visual interface - displaying tasks grouped by column with drag-and-drop support.

**File**: `ui/src/components/Board.tsx`

```tsx
import { useMemo } from 'react'
import {
  DndContext,
  DragEndEvent,
  DragOverlay,
  DragStartEvent,
  closestCenter,
  PointerSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core'
import { useState } from 'react'
import { Column } from './Column'
import { TaskCard } from './TaskCard'
import { useTasks } from '../hooks/useTasks'
import { COLUMNS, type Task, type Column as ColumnType } from '../types/task'
import styles from './Board.module.css'

interface BoardProps {
  onTaskClick?: (task: Task) => void
  onTaskSelect?: (task: Task) => void
  selectedTaskId?: string | null
}

/**
 * Kanban board with columns and drag-and-drop.
 * 
 * Features:
 * - Displays tasks grouped by column
 * - Drag tasks between columns
 * - Real-time updates from CLI changes
 * - Click task to open detail panel
 * - Selected task state for keyboard navigation
 */
export function Board({ onTaskClick, onTaskSelect, selectedTaskId }: BoardProps) {
  const { tasks, loading, error, moveTask } = useTasks()
  const [activeTask, setActiveTask] = useState<Task | null>(null)

  // Configure drag sensors
  // PointerSensor requires a small movement before dragging starts
  // This prevents accidental drags when clicking
  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8, // 8px movement required to start drag
      },
    })
  )

  // Group tasks by column
  const tasksByColumn = useMemo(() => {
    const grouped: Record<ColumnType, Task[]> = {
      backlog: [],
      todo: [],
      in_progress: [],
      review: [],
      done: [],
    }

    tasks.forEach((task) => {
      if (grouped[task.column]) {
        grouped[task.column].push(task)
      }
    })

    // Sort each column by position
    Object.keys(grouped).forEach((col) => {
      grouped[col as ColumnType].sort((a, b) => a.position - b.position)
    })

    return grouped
  }, [tasks])

  // Handle drag start - store the dragged task for overlay
  const handleDragStart = (event: DragStartEvent) => {
    const task = tasks.find((t) => t.id === event.active.id)
    if (task) {
      setActiveTask(task)
    }
  }

  // Handle drag end - move task to new column
  const handleDragEnd = async (event: DragEndEvent) => {
    setActiveTask(null)
    
    const { active, over } = event
    if (!over) return

    const taskId = active.id as string
    const task = tasks.find((t) => t.id === taskId)
    if (!task) return

    // Get the target column from the droppable area
    const targetColumn = over.data.current?.column as ColumnType | undefined
    if (!targetColumn) return

    // If dropped in same column, no change needed (position sorting is Phase 4)
    if (task.column === targetColumn) return

    // Calculate new position (append to end of target column)
    const targetTasks = tasksByColumn[targetColumn]
    const maxPosition = targetTasks.reduce(
      (max, t) => Math.max(max, t.position),
      0
    )
    const newPosition = maxPosition + 1000

    // Move task to new column
    await moveTask(taskId, targetColumn, newPosition)
  }

  if (loading) {
    return (
      <div className={styles.loading}>
        <span>Loading tasks...</span>
      </div>
    )
  }

  if (error) {
    return (
      <div className={styles.error}>
        <span>Error: {error.message}</span>
      </div>
    )
  }

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
    >
      <div className={styles.board}>
        {COLUMNS.map((column) => (
          <Column
            key={column}
            column={column}
            tasks={tasksByColumn[column]}
            onTaskClick={onTaskClick}
            onTaskSelect={onTaskSelect}
            selectedTaskId={selectedTaskId}
          />
        ))}
      </div>

      {/* Drag overlay shows the card being dragged */}
      <DragOverlay>
        {activeTask ? <TaskCard task={activeTask} isDragging /> : null}
      </DragOverlay>
    </DndContext>
  )
}
```

**File**: `ui/src/components/Board.module.css`

```css
.board {
  display: flex;
  gap: var(--column-gap);
  padding: var(--space-4);
  height: 100%;
  overflow-x: auto;
  overflow-y: hidden;
}

.loading,
.error {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  color: var(--text-secondary);
  font-size: var(--text-lg);
}

.error {
  color: var(--priority-urgent);
}
```

**File**: `ui/src/components/Column.tsx`

```tsx
import { useDroppable } from '@dnd-kit/core'
import { TaskCard } from './TaskCard'
import { COLUMN_NAMES, type Column as ColumnType, type Task } from '../types/task'
import styles from './Column.module.css'

interface ColumnProps {
  column: ColumnType
  tasks: Task[]
  onTaskClick?: (task: Task) => void
  onTaskSelect?: (task: Task) => void
  selectedTaskId?: string | null
}

/**
 * A single column in the kanban board.
 * 
 * Acts as a droppable target for drag-and-drop.
 * Displays column header with name and count.
 */
export function Column({ column, tasks, onTaskClick, onTaskSelect, selectedTaskId }: ColumnProps) {
  // Make this column a droppable target
  const { setNodeRef, isOver } = useDroppable({
    id: `column-${column}`,
    data: {
      column, // Pass column info to drag handlers
    },
  })

  return (
    <div
      ref={setNodeRef}
      className={`${styles.column} ${isOver ? styles.over : ''}`}
    >
      <div className={styles.header}>
        <div className={styles.headerContent}>
          <span
            className={styles.statusDot}
            style={{ backgroundColor: `var(--status-${column.replace('_', '-')})` }}
          />
          <span className={styles.name}>{COLUMN_NAMES[column]}</span>
          <span className={styles.count}>{tasks.length}</span>
        </div>
      </div>

      <div className={styles.tasks}>
        {tasks.map((task) => (
          <TaskCard 
            key={task.id} 
            task={task} 
            onClick={onTaskClick}
            onSelect={onTaskSelect}
            isSelected={selectedTaskId === task.id}
          />
        ))}
      </div>
    </div>
  )
}
```

**File**: `ui/src/components/Column.module.css`

```css
.column {
  display: flex;
  flex-direction: column;
  width: var(--column-width);
  min-width: var(--column-width);
  height: 100%;
  background-color: transparent;
  border-radius: var(--radius-lg);
  transition: background-color var(--duration-fast) var(--ease-default);
}

.column.over {
  background-color: var(--accent-muted);
}

.header {
  position: sticky;
  top: 0;
  padding: var(--space-2) var(--space-3);
  background-color: var(--bg-app);
  z-index: 1;
}

.headerContent {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.statusDot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.name {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.count {
  font-size: var(--text-sm);
  color: var(--text-muted);
}

.tasks {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  padding: var(--space-2);
  overflow-y: auto;
  flex: 1;
}
```

**File**: `ui/src/components/TaskCard.tsx`

```tsx
import { useDraggable } from '@dnd-kit/core'
import { CSS } from '@dnd-kit/utilities'
import type { Task } from '../types/task'
import styles from './TaskCard.module.css'

interface TaskCardProps {
  task: Task
  isDragging?: boolean
  isSelected?: boolean
  onClick?: (task: Task) => void
  onSelect?: (task: Task) => void
}

/**
 * A draggable task card.
 * 
 * Displays:
 * - Status dot and task ID
 * - Title (truncated to 2 lines)
 * - Labels (if any)
 * - Priority indicator
 * - Due date (if set)
 * 
 * Clicking opens the task detail panel.
 */
export function TaskCard({ task, isDragging = false, isSelected = false, onClick, onSelect }: TaskCardProps) {
  // Make this card draggable
  const { attributes, listeners, setNodeRef, transform, isDragging: isCurrentlyDragging } = useDraggable({
    id: task.id,
  })

  // Apply transform during drag
  const style = transform
    ? {
        transform: CSS.Transform.toString(transform),
      }
    : undefined

  // Priority indicator
  const getPriorityIndicator = (priority: string) => {
    switch (priority) {
      case 'urgent':
        return { emoji: '🔴', label: 'Urgent' }
      case 'high':
        return { emoji: '🟠', label: 'High' }
      case 'medium':
        return { emoji: '🟡', label: 'Medium' }
      case 'low':
        return { emoji: '⚪', label: 'Low' }
      default:
        return null
    }
  }

  const priority = getPriorityIndicator(task.priority)

  // Handle click to open detail panel
  // Only trigger if not dragging (to avoid opening panel after drag)
  const handleClick = () => {
    if (!isCurrentlyDragging && onClick) {
      onClick(task)
    }
  }

  // Handle focus/selection (for keyboard navigation)
  const handleFocus = () => {
    if (onSelect) {
      onSelect(task)
    }
  }

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={`${styles.card} ${isDragging ? styles.dragging : ''} ${isSelected ? styles.selected : ''}`}
      onClick={handleClick}
      onFocus={handleFocus}
      tabIndex={0}
      role="button"
      aria-pressed={isSelected}
      {...listeners}
      {...attributes}
    >
      {/* Header: Status dot + ID */}
      <div className={styles.header}>
        <span
          className={styles.statusDot}
          style={{
            backgroundColor: `var(--status-${task.column.replace('_', '-')})`,
          }}
        />
        <span className={styles.id}>{task.id.slice(0, 8)}</span>
      </div>

      {/* Title */}
      <h3 className={styles.title}>{task.title}</h3>

      {/* Labels */}
      {task.labels && task.labels.length > 0 && (
        <div className={styles.labels}>
          {task.labels.map((label) => (
            <span key={label} className={styles.label}>
              {label}
            </span>
          ))}
        </div>
      )}

      {/* Footer: Priority + Due Date + Type */}
      <div className={styles.footer}>
        {priority && (
          <span className={styles.priority} title={priority.label}>
            {priority.emoji} {priority.label}
          </span>
        )}
        {task.due_date && (
          <span className={styles.dueDate}>
            📅 {new Date(task.due_date).toLocaleDateString('en-US', { 
              month: 'short', 
              day: 'numeric' 
            })}
          </span>
        )}
        <span
          className={styles.type}
          style={{ color: `var(--type-${task.type})` }}
        >
          {task.type}
        </span>
      </div>
    </div>
  )
}
```

**File**: `ui/src/components/TaskCard.module.css`

```css
.card {
  padding: var(--space-3);
  background-color: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  cursor: grab;
  transition: 
    background-color var(--duration-fast) var(--ease-default),
    box-shadow var(--duration-fast) var(--ease-default),
    transform var(--duration-fast) var(--ease-default);
}

.card:hover {
  background-color: var(--bg-card-hover);
}

.card:active {
  cursor: grabbing;
}

.card.dragging {
  background-color: var(--bg-card-selected);
  box-shadow: var(--shadow-drag);
  transform: scale(1.02);
  opacity: 0.9;
}

.card.selected {
  background-color: var(--bg-card-selected);
  border-color: var(--accent);
}

.card:focus {
  outline: none;
  border-color: var(--accent);
}

.header {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin-bottom: var(--space-2);
}

.statusDot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}

.id {
  font-family: var(--font-mono);
  font-size: var(--text-xs);
  color: var(--text-muted);
}

.title {
  font-size: var(--text-lg);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  line-height: var(--leading-tight);
  margin-bottom: var(--space-2);
  
  /* Truncate to 2 lines */
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.labels {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-1);
  margin-bottom: var(--space-2);
}

.label {
  padding: 2px 6px;
  font-size: var(--text-xs);
  font-weight: var(--font-medium);
  color: var(--text-secondary);
  background-color: var(--bg-input);
  border-radius: var(--radius-sm);
}

.footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
}

.priority {
  font-size: var(--text-xs);
  color: var(--text-secondary);
}

.dueDate {
  font-size: var(--text-xs);
  color: var(--text-secondary);
}

.type {
  font-size: var(--text-xs);
  font-weight: var(--font-medium);
  text-transform: capitalize;
}
```

**Steps**:

1. Create all the component files above.

2. Verify everything compiles:
   ```bash
   cd ui && npm run build
   ```

---

### 2.11 Create Quick Create Modal

**What**: Create a modal for quickly creating new tasks.

**Why**: Users need a fast way to add tasks without leaving the board view. The `C` keyboard shortcut opens this modal.

**File**: `ui/src/components/QuickCreate.tsx`

```tsx
import { useState, useEffect, useRef } from 'react'
import { COLUMNS, COLUMN_NAMES, type Column } from '../types/task'
import styles from './QuickCreate.module.css'

interface QuickCreateProps {
  isOpen: boolean
  onClose: () => void
  onCreate: (title: string, column: Column) => Promise<void>
}

/**
 * Modal for quickly creating a new task.
 * 
 * Features:
 * - Auto-focus on title input
 * - Column selector (defaults to "backlog")
 * - Enter to create, Esc to cancel
 */
export function QuickCreate({ isOpen, onClose, onCreate }: QuickCreateProps) {
  const [title, setTitle] = useState('')
  const [column, setColumn] = useState<Column>('backlog')
  const [isCreating, setIsCreating] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  // Auto-focus input when modal opens
  useEffect(() => {
    if (isOpen && inputRef.current) {
      inputRef.current.focus()
    }
  }, [isOpen])

  // Reset form when modal closes
  useEffect(() => {
    if (!isOpen) {
      setTitle('')
      setColumn('backlog')
    }
  }, [isOpen])

  // Handle keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (!isOpen) return

      if (e.key === 'Escape') {
        onClose()
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [isOpen, onClose])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!title.trim() || isCreating) return

    setIsCreating(true)
    try {
      await onCreate(title.trim(), column)
      onClose()
    } finally {
      setIsCreating(false)
    }
  }

  if (!isOpen) return null

  return (
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
        <h2 className={styles.title}>Create Task</h2>
        
        <form onSubmit={handleSubmit}>
          <div className={styles.field}>
            <input
              ref={inputRef}
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Task title..."
              className={styles.input}
              disabled={isCreating}
            />
          </div>

          <div className={styles.field}>
            <label className={styles.label}>Column</label>
            <select
              value={column}
              onChange={(e) => setColumn(e.target.value as Column)}
              className={styles.select}
              disabled={isCreating}
            >
              {COLUMNS.map((col) => (
                <option key={col} value={col}>
                  {COLUMN_NAMES[col]}
                </option>
              ))}
            </select>
          </div>

          <div className={styles.actions}>
            <button
              type="button"
              onClick={onClose}
              className={styles.cancelButton}
              disabled={isCreating}
            >
              Cancel
            </button>
            <button
              type="submit"
              className={styles.createButton}
              disabled={!title.trim() || isCreating}
            >
              {isCreating ? 'Creating...' : 'Create'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
```

**File**: `ui/src/components/QuickCreate.module.css`

```css
.overlay {
  position: fixed;
  inset: 0;
  background-color: var(--bg-overlay);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  animation: fadeIn var(--duration-fast) var(--ease-out);
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.modal {
  width: 100%;
  max-width: 480px;
  padding: var(--space-6);
  background-color: var(--bg-sidebar);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  animation: slideIn var(--duration-normal) var(--ease-out);
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: scale(0.95) translateY(-10px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.title {
  font-size: var(--text-xl);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  margin-bottom: var(--space-5);
}

.field {
  margin-bottom: var(--space-4);
}

.label {
  display: block;
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--text-secondary);
  margin-bottom: var(--space-1);
}

.input,
.select {
  width: 100%;
  padding: var(--space-3);
  font-family: var(--font-sans);
  font-size: var(--text-base);
  color: var(--text-primary);
  background-color: var(--bg-input);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  outline: none;
  transition: border-color var(--duration-fast) var(--ease-default);
}

.input:focus,
.select:focus {
  border-color: var(--accent);
}

.input::placeholder {
  color: var(--text-muted);
}

.select {
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%23A0A0A0' d='M6 8L2 4h8z'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  padding-right: 36px;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-2);
  margin-top: var(--space-5);
}

.cancelButton,
.createButton {
  padding: var(--space-2) var(--space-4);
  font-family: var(--font-sans);
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: 
    background-color var(--duration-fast) var(--ease-default),
    opacity var(--duration-fast) var(--ease-default);
}

.cancelButton {
  color: var(--text-secondary);
  background-color: transparent;
  border: 1px solid var(--border-default);
}

.cancelButton:hover {
  background-color: var(--bg-card);
}

.createButton {
  color: white;
  background-color: var(--accent);
  border: none;
}

.createButton:hover:not(:disabled) {
  background-color: var(--accent-hover);
}

.createButton:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
```

**Steps**:

1. Create both files above.

2. Verify compilation:
   ```bash
   cd ui && npm run build
   ```

---

### 2.12 Create Task Detail Panel

**What**: Create a slide-in panel that shows full task details.

**Why**: Users need to view and edit task properties beyond what fits on a card. The panel slides in from the right.

**File**: `ui/src/components/TaskDetail.tsx`

```tsx
import { useEffect, useRef } from 'react'
import { 
  COLUMN_NAMES, 
  PRIORITY_NAMES, 
  TYPE_NAMES,
  COLUMNS,
  PRIORITIES,
  TYPES,
  type Task, 
  type Column,
  type Priority,
  type TaskType,
} from '../types/task'
import styles from './TaskDetail.module.css'

interface TaskDetailProps {
  task: Task | null
  onClose: () => void
  onUpdate: (id: string, data: Partial<Task>) => Promise<void>
}

/**
 * Slide-in panel showing full task details.
 * 
 * Features:
 * - Close with Esc or click outside
 * - Editable properties via dropdowns (status, priority, type)
 * - Display all task metadata
 */
export function TaskDetail({ task, onClose, onUpdate }: TaskDetailProps) {
  const panelRef = useRef<HTMLDivElement>(null)

  // Handle keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && task) {
        onClose()
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [task, onClose])

  // Handle click outside
  const handleOverlayClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose()
    }
  }

  if (!task) return null

  const handleColumnChange = async (newColumn: Column) => {
    await onUpdate(task.id, { column: newColumn })
  }

  const handlePriorityChange = async (newPriority: Priority) => {
    await onUpdate(task.id, { priority: newPriority })
  }

  const handleTypeChange = async (newType: TaskType) => {
    await onUpdate(task.id, { type: newType })
  }

  // Get priority indicator emoji
  const getPriorityEmoji = (priority: Priority): string => {
    switch (priority) {
      case 'urgent': return '🔴'
      case 'high': return '🟠'
      case 'medium': return '🟡'
      case 'low': return '⚪'
      default: return ''
    }
  }

  return (
    <div className={styles.overlay} onClick={handleOverlayClick}>
      <div ref={panelRef} className={styles.panel}>
        {/* Header */}
        <div className={styles.header}>
          <button onClick={onClose} className={styles.closeButton}>
            ← Back
          </button>
        </div>

        {/* Content */}
        <div className={styles.content}>
          {/* Title */}
          <h1 className={styles.title}>{task.title}</h1>
          <span className={styles.id}>{task.id}</span>

          {/* Description */}
          {task.description && (
            <div className={styles.description}>
              <h3 className={styles.sectionTitle}>Description</h3>
              <p>{task.description}</p>
            </div>
          )}

          {/* Properties */}
          <div className={styles.properties}>
            <div className={styles.property}>
              <span className={styles.propertyLabel}>Status</span>
              <select
                value={task.column}
                onChange={(e) => handleColumnChange(e.target.value as Column)}
                className={styles.propertySelect}
              >
                {COLUMNS.map((col) => (
                  <option key={col} value={col}>
                    {COLUMN_NAMES[col]}
                  </option>
                ))}
              </select>
            </div>

            <div className={styles.property}>
              <span className={styles.propertyLabel}>Type</span>
              <select
                value={task.type}
                onChange={(e) => handleTypeChange(e.target.value as TaskType)}
                className={styles.propertySelect}
              >
                {TYPES.map((type) => (
                  <option key={type} value={type}>
                    {TYPE_NAMES[type]}
                  </option>
                ))}
              </select>
            </div>

            <div className={styles.property}>
              <span className={styles.propertyLabel}>Priority</span>
              <select
                value={task.priority}
                onChange={(e) => handlePriorityChange(e.target.value as Priority)}
                className={styles.propertySelect}
              >
                {PRIORITIES.map((priority) => (
                  <option key={priority} value={priority}>
                    {getPriorityEmoji(priority)} {PRIORITY_NAMES[priority]}
                  </option>
                ))}
              </select>
            </div>

            {task.labels && task.labels.length > 0 && (
              <div className={styles.property}>
                <span className={styles.propertyLabel}>Labels</span>
                <div className={styles.labels}>
                  {task.labels.map((label) => (
                    <span key={label} className={styles.label}>
                      {label}
                    </span>
                  ))}
                </div>
              </div>
            )}

            {task.due_date && (
              <div className={styles.property}>
                <span className={styles.propertyLabel}>Due Date</span>
                <span className={styles.propertyValue}>
                  {new Date(task.due_date).toLocaleDateString()}
                </span>
              </div>
            )}

            {task.blocked_by && task.blocked_by.length > 0 && (
              <div className={styles.property}>
                <span className={styles.propertyLabel}>Blocked By</span>
                <div className={styles.blockedList}>
                  {task.blocked_by.map((id) => (
                    <span key={id} className={styles.blockedId}>
                      {id.slice(0, 8)}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>

          {/* Metadata */}
          <div className={styles.metadata}>
            <div className={styles.metaItem}>
              <span className={styles.metaLabel}>Created</span>
              <span className={styles.metaValue}>
                {new Date(task.created).toLocaleDateString()}
              </span>
            </div>
            <div className={styles.metaItem}>
              <span className={styles.metaLabel}>Updated</span>
              <span className={styles.metaValue}>
                {new Date(task.updated).toLocaleDateString()}
              </span>
            </div>
            {task.created_by && (
              <div className={styles.metaItem}>
                <span className={styles.metaLabel}>Created By</span>
                <span className={styles.metaValue}>
                  {task.created_by}
                  {task.created_by_agent && ` (${task.created_by_agent})`}
                </span>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
```

**File**: `ui/src/components/TaskDetail.module.css`

```css
.overlay {
  position: fixed;
  inset: 0;
  background-color: var(--bg-overlay);
  z-index: 50;
  animation: fadeIn var(--duration-fast) var(--ease-out);
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.panel {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  width: var(--detail-panel-width);
  max-width: 100%;
  background-color: var(--bg-sidebar);
  border-left: 1px solid var(--border-default);
  overflow-y: auto;
  animation: slideIn var(--duration-normal) var(--ease-out);
}

@keyframes slideIn {
  from {
    transform: translateX(100%);
  }
  to {
    transform: translateX(0);
  }
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-4);
  border-bottom: 1px solid var(--border-subtle);
}

.closeButton {
  padding: var(--space-2) var(--space-3);
  font-family: var(--font-sans);
  font-size: var(--text-sm);
  color: var(--text-secondary);
  background: none;
  border: none;
  cursor: pointer;
  border-radius: var(--radius-md);
  transition: background-color var(--duration-fast) var(--ease-default);
}

.closeButton:hover {
  background-color: var(--bg-card);
}

.content {
  padding: var(--space-5);
}

.title {
  font-size: var(--text-2xl);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  line-height: var(--leading-tight);
  margin-bottom: var(--space-2);
}

.id {
  font-family: var(--font-mono);
  font-size: var(--text-sm);
  color: var(--text-muted);
  display: block;
  margin-bottom: var(--space-5);
}

.description {
  margin-bottom: var(--space-5);
}

.sectionTitle {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: var(--space-2);
}

.description p {
  font-size: var(--text-base);
  color: var(--text-primary);
  line-height: var(--leading-relaxed);
}

.properties {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  padding: var(--space-4) 0;
  border-top: 1px solid var(--border-subtle);
  border-bottom: 1px solid var(--border-subtle);
}

.property {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.propertyLabel {
  font-size: var(--text-sm);
  color: var(--text-secondary);
}

.propertyValue {
  font-size: var(--text-sm);
  color: var(--text-primary);
  text-transform: capitalize;
}

.propertySelect {
  padding: var(--space-1) var(--space-2);
  font-family: var(--font-sans);
  font-size: var(--text-sm);
  color: var(--text-primary);
  background-color: var(--bg-input);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  cursor: pointer;
}

.labels {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-1);
}

.label {
  padding: 2px 8px;
  font-size: var(--text-xs);
  font-weight: var(--font-medium);
  color: var(--text-secondary);
  background-color: var(--bg-input);
  border-radius: var(--radius-sm);
}

.blockedList {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-1);
}

.blockedId {
  padding: 2px 8px;
  font-family: var(--font-mono);
  font-size: var(--text-xs);
  color: var(--priority-urgent);
  background-color: rgba(239, 68, 68, 0.1);
  border-radius: var(--radius-sm);
}

.metadata {
  padding-top: var(--space-4);
}

.metaItem {
  display: flex;
  justify-content: space-between;
  margin-bottom: var(--space-2);
}

.metaLabel {
  font-size: var(--text-sm);
  color: var(--text-muted);
}

.metaValue {
  font-size: var(--text-sm);
  color: var(--text-secondary);
}
```

**Steps**:

1. Create both files above.

2. Verify compilation:
   ```bash
   cd ui && npm run build
   ```

---

### 2.13 Wire Everything Together in App.tsx

**What**: Update the main App component to use all the components we created.

**Why**: This connects all the pieces: layout, board, quick create modal, task detail panel, and keyboard shortcuts.

**File**: `ui/src/App.tsx`

```tsx
import { useState, useEffect, useCallback } from 'react'
import { Layout } from './components/Layout'
import { Board } from './components/Board'
import { QuickCreate } from './components/QuickCreate'
import { TaskDetail } from './components/TaskDetail'
import { useTasks } from './hooks/useTasks'
import type { Task, Column } from './types/task'

/**
 * Main application component.
 * 
 * Manages:
 * - Quick create modal (opened with 'C' key)
 * - Task detail panel (opened by clicking a task or pressing Enter)
 * - Selected task state (for keyboard navigation)
 * - Global keyboard shortcuts
 */
function App() {
  const { tasks, createTask, updateTask } = useTasks()
  const [isQuickCreateOpen, setIsQuickCreateOpen] = useState(false)
  const [selectedTask, setSelectedTask] = useState<Task | null>(null)
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null)

  // Global keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Don't trigger shortcuts when typing in inputs
      if (
        e.target instanceof HTMLInputElement ||
        e.target instanceof HTMLTextAreaElement ||
        e.target instanceof HTMLSelectElement
      ) {
        return
      }

      // 'C' to open quick create
      if (e.key === 'c' || e.key === 'C') {
        e.preventDefault()
        setIsQuickCreateOpen(true)
      }

      // 'Enter' to open selected task detail
      if (e.key === 'Enter' && selectedTaskId && !selectedTask) {
        e.preventDefault()
        const task = tasks.find((t) => t.id === selectedTaskId)
        if (task) {
          setSelectedTask(task)
        }
      }

      // 'Escape' to deselect task (when detail panel is closed)
      if (e.key === 'Escape' && !selectedTask && selectedTaskId) {
        e.preventDefault()
        setSelectedTaskId(null)
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [selectedTaskId, selectedTask, tasks])

  // Handle task creation
  const handleCreate = useCallback(
    async (title: string, column: Column) => {
      await createTask(title, column)
    },
    [createTask]
  )

  // Handle task update
  const handleUpdate = useCallback(
    async (id: string, data: Partial<Task>) => {
      await updateTask(id, data)
    },
    [updateTask]
  )

  // Handle task click to open detail panel
  const handleTaskClick = useCallback((task: Task) => {
    setSelectedTaskId(task.id)
    setSelectedTask(task)
  }, [])

  // Handle task selection (without opening detail)
  const handleTaskSelect = useCallback((task: Task) => {
    setSelectedTaskId(task.id)
  }, [])

  // Handle closing detail panel
  const handleCloseDetail = useCallback(() => {
    setSelectedTask(null)
    // Keep selectedTaskId so user can press Enter to reopen
  }, [])

  return (
    <Layout>
      <Board 
        onTaskClick={handleTaskClick} 
        selectedTaskId={selectedTaskId}
        onTaskSelect={handleTaskSelect}
      />

      {/* Quick Create Modal */}
      <QuickCreate
        isOpen={isQuickCreateOpen}
        onClose={() => setIsQuickCreateOpen(false)}
        onCreate={handleCreate}
      />

      {/* Task Detail Panel */}
      <TaskDetail
        task={selectedTask}
        onClose={handleCloseDetail}
        onUpdate={handleUpdate}
      />
    </Layout>
  )
}

export default App
```

**Steps**:

1. Replace the contents of `ui/src/App.tsx` with the code above.

2. Clean up the default Vite assets (no longer needed):
   ```bash
   rm -f ui/src/assets/react.svg ui/public/vite.svg
   ```

3. Build and test:
   ```bash
   cd ui && npm run build && cd .. && make build
   ./egenskriven serve
   ```

4. Open `http://localhost:8090` in your browser.
   
   **Expected**: Dark-themed kanban board with 5 columns.

5. Press `C` to open the quick create modal.

6. Create a test task via CLI in another terminal:
   ```bash
   ./egenskriven add "Test task from CLI" --column todo
   ```
   
   **Expected**: Task appears in the UI without refresh (real-time sync).

---

### 2.14 Setup UI Testing Infrastructure

**What**: Configure Vitest for testing React components.

**Why**: Automated tests catch bugs early and provide confidence when making changes.

**Steps**:

1. Install testing dependencies:
   ```bash
   cd ui
   npm install -D vitest @testing-library/react @testing-library/jest-dom @testing-library/user-event jsdom @types/testing-library__jest-dom
   ```

2. Create Vitest configuration file `ui/vitest.config.ts`:
   ```typescript
   import { defineConfig } from 'vitest/config'
   import react from '@vitejs/plugin-react'

   export default defineConfig({
     plugins: [react()],
     test: {
       environment: 'jsdom',
       setupFiles: './src/test/setup.ts',
       globals: true,
     },
   })
   ```

3. Create test setup file `ui/src/test/setup.ts`:
   ```typescript
   import '@testing-library/jest-dom'
   ```

4. Create the test directory:
   ```bash
   mkdir -p ui/src/test
   ```

5. Update `ui/package.json` to add the test script:
   ```json
   {
     "scripts": {
       "dev": "vite",
       "build": "tsc && vite build",
       "lint": "eslint . --ext ts,tsx --report-unused-disable-directives --max-warnings 0",
       "preview": "vite preview",
       "test": "vitest",
       "test:run": "vitest run"
     }
   }
   ```

6. Create a sample component test `ui/src/components/TaskCard.test.tsx`:
   ```tsx
   import { describe, it, expect } from 'vitest'
   import { render, screen } from '@testing-library/react'
   import { TaskCard } from './TaskCard'
   import type { Task } from '../types/task'

   // Mock task for testing
   const mockTask: Task = {
     id: 'test-123',
     title: 'Test Task Title',
     type: 'feature',
     priority: 'high',
     column: 'todo',
     position: 1000,
     labels: ['frontend', 'ui'],
     collectionId: 'tasks',
     collectionName: 'tasks',
     created: '2024-01-15T10:00:00Z',
     updated: '2024-01-15T10:00:00Z',
   }

   describe('TaskCard', () => {
     it('renders task title', () => {
       render(<TaskCard task={mockTask} />)
       expect(screen.getByText('Test Task Title')).toBeInTheDocument()
     })

     it('renders task ID (truncated)', () => {
       render(<TaskCard task={mockTask} />)
       expect(screen.getByText('test-123')).toBeInTheDocument()
     })

     it('renders labels', () => {
       render(<TaskCard task={mockTask} />)
       expect(screen.getByText('frontend')).toBeInTheDocument()
       expect(screen.getByText('ui')).toBeInTheDocument()
     })

     it('renders priority indicator', () => {
       render(<TaskCard task={mockTask} />)
       expect(screen.getByText('🟠 High')).toBeInTheDocument()
     })

     it('renders task type', () => {
       render(<TaskCard task={mockTask} />)
       expect(screen.getByText('feature')).toBeInTheDocument()
     })
   })
   ```

7. Create a test for the Board component `ui/src/components/Board.test.tsx`:
   ```tsx
   import { describe, it, expect, vi } from 'vitest'
   import { render, screen } from '@testing-library/react'
   import { Board } from './Board'
   import { COLUMNS, COLUMN_NAMES } from '../types/task'

   // Mock the useTasks hook
   vi.mock('../hooks/useTasks', () => ({
     useTasks: () => ({
       tasks: [
         {
           id: 'task-1',
           title: 'Task in Backlog',
           type: 'feature',
           priority: 'medium',
           column: 'backlog',
           position: 1000,
           labels: [],
           collectionId: 'tasks',
           collectionName: 'tasks',
           created: '2024-01-15T10:00:00Z',
           updated: '2024-01-15T10:00:00Z',
         },
         {
           id: 'task-2',
           title: 'Task in Todo',
           type: 'bug',
           priority: 'high',
           column: 'todo',
           position: 1000,
           labels: [],
           collectionId: 'tasks',
           collectionName: 'tasks',
           created: '2024-01-15T10:00:00Z',
           updated: '2024-01-15T10:00:00Z',
         },
         {
           id: 'task-3',
           title: 'Task in Progress',
           type: 'chore',
           priority: 'low',
           column: 'in_progress',
           position: 1000,
           labels: [],
           collectionId: 'tasks',
           collectionName: 'tasks',
           created: '2024-01-15T10:00:00Z',
           updated: '2024-01-15T10:00:00Z',
         },
       ],
       loading: false,
       error: null,
       moveTask: vi.fn(),
       createTask: vi.fn(),
       updateTask: vi.fn(),
       deleteTask: vi.fn(),
     }),
   }))

   describe('Board', () => {
     it('renders all five columns', () => {
       render(<Board />)
       
       // Check all column headers are present
       COLUMNS.forEach((column) => {
         expect(screen.getByText(COLUMN_NAMES[column])).toBeInTheDocument()
       })
     })

     it('groups tasks into correct columns', () => {
       render(<Board />)
       
       // Tasks should be rendered
       expect(screen.getByText('Task in Backlog')).toBeInTheDocument()
       expect(screen.getByText('Task in Todo')).toBeInTheDocument()
       expect(screen.getByText('Task in Progress')).toBeInTheDocument()
     })

     it('shows task count in column headers', () => {
       render(<Board />)
       
       // Backlog has 1 task, Todo has 1 task, In Progress has 1 task
       // The count should appear in the column header
       expect(screen.getByText('1')).toBeInTheDocument() // At least one count visible
     })

     it('shows loading state', () => {
       // Override mock for this test
       vi.doMock('../hooks/useTasks', () => ({
         useTasks: () => ({
           tasks: [],
           loading: true,
           error: null,
           moveTask: vi.fn(),
           createTask: vi.fn(),
           updateTask: vi.fn(),
           deleteTask: vi.fn(),
         }),
       }))
       
       // Re-import to get new mock
       // Note: In practice, you may need to restructure this test
     })

     it('shows error state when fetch fails', () => {
       // Similar pattern to loading state test
     })
   })
   ```

8. Create a test for the useTasks hook `ui/src/hooks/useTasks.test.ts`:
   ```typescript
   import { describe, it, expect, vi, beforeEach } from 'vitest'
   import { renderHook, waitFor } from '@testing-library/react'
   import { useTasks } from './useTasks'

   // Mock PocketBase
   vi.mock('../lib/pb', () => ({
     pb: {
       collection: vi.fn(() => ({
         getFullList: vi.fn().mockResolvedValue([]),
         subscribe: vi.fn(),
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
   ```

8. Run the tests:
   ```bash
   cd ui && npm run test:run
   ```
   
   **Expected output**:
   ```
   ✓ src/components/TaskCard.test.tsx (5)
      ✓ TaskCard (5)
        ✓ renders task title
        ✓ renders task ID (truncated)
        ✓ renders labels
        ✓ renders priority indicator
        ✓ renders task type
   ✓ src/hooks/useTasks.test.ts (3)
      ✓ useTasks (3)
        ✓ starts with loading state
        ✓ fetches tasks on mount
        ✓ provides CRUD operations

   Test Files  2 passed (2)
        Tests  8 passed (8)
   ```

---

## Verification Checklist

Complete each section in order. Check off each item as you verify it.

### Build Verification

- [ ] **UI builds successfully**
  ```bash
  cd ui && npm run build
  ```
  Should produce `ui/dist/` directory.

- [ ] **Go binary builds with embedded UI**
  ```bash
  make build
  ```
  Should produce `egenskriven` binary (~35-50MB).

- [ ] **Full build command works**
  ```bash
  make clean && make build
  ```
  Should build from scratch successfully.

### Runtime Verification

- [ ] **Server starts and serves UI**
  ```bash
  ./egenskriven serve
  ```
  Open `http://localhost:8090` - should show kanban board.

- [ ] **PocketBase admin UI still works**
  
  Open `http://localhost:8090/_/` - should show admin login.

- [ ] **CLI still works**
  ```bash
  ./egenskriven add "Test task"
  ./egenskriven list
  ```
  Should create and list tasks.

### UI Feature Verification

- [ ] **Board displays all columns**
  
  Should see: Backlog, Todo, In Progress, Review, Done

- [ ] **Quick create opens with 'C' key**
  
  Press `C` - modal should open.

- [ ] **Tasks can be created via modal**
  
  Open modal, enter title, select column, click Create.

- [ ] **Tasks can be dragged between columns**
  
  Drag a task from one column to another.

- [ ] **Real-time sync works**
  
  In one terminal, run: `./egenskriven add "CLI task" --column todo`
  
  Task should appear in UI without refresh.

- [ ] **Task detail panel opens on click**
  
  Click a task - detail panel should slide in from right.

- [ ] **Task detail panel opens on Enter key**
  
  Focus a task (tab or click), then press `Enter` - detail panel should open.

- [ ] **Panel closes with Esc**
  
  Press `Esc` - panel should close.

- [ ] **Selected task has visual indicator**
  
  Click a task without opening detail - should show selected border color.

- [ ] **Task properties are editable**
  
  In detail panel, change status, priority, and type via dropdowns.

- [ ] **Due date displays on cards**
  
  Create a task with due date via CLI:
  ```bash
  ./egenskriven add "Task with due" --due "2025-01-15"
  ```
  Card should show `📅 Jan 15`.

- [ ] **Dark mode styling applied**
  
  UI should have dark background (#0D0D0D) and light text.

### Test Verification

- [ ] **UI tests pass**
  ```bash
  cd ui && npm run test:run
  ```
  All tests should pass (TaskCard, Board, useTasks).

- [ ] **Go tests still pass**
  ```bash
  make test
  ```
  All existing tests should pass.

### Accessibility Verification

- [ ] **Focus visible on keyboard navigation**
  
  Tab through tasks - focus ring should be visible.

- [ ] **Task cards have proper ARIA attributes**
  
  Cards should have `role="button"` and `tabIndex={0}`.

---

## File Summary

| File | Lines (approx) | Purpose |
|------|----------------|---------|
| `ui/vite.config.ts` | 25 | Vite configuration with proxy |
| `ui/embed.go` | 15 | Go embed directive for UI |
| `cmd/egenskriven/main.go` | 50 | Updated to serve embedded UI |
| `ui/src/styles/tokens.css` | 150 | Design system tokens |
| `ui/src/lib/pb.ts` | 10 | PocketBase client |
| `ui/src/types/task.ts` | 80 | Task TypeScript types (all fields from architecture) |
| `ui/src/hooks/useTasks.ts` | 120 | Task data hook with real-time |
| `ui/src/components/Layout.tsx` | 20 | Layout wrapper |
| `ui/src/components/Layout.module.css` | 15 | Layout styles |
| `ui/src/components/Header.tsx` | 25 | Header component |
| `ui/src/components/Header.module.css` | 40 | Header styles |
| `ui/src/components/Board.tsx` | 110 | Kanban board with DnD and selection |
| `ui/src/components/Board.module.css` | 25 | Board styles |
| `ui/src/components/Column.tsx` | 55 | Column component with selection props |
| `ui/src/components/Column.module.css` | 60 | Column styles |
| `ui/src/components/TaskCard.tsx` | 100 | Draggable task card with selection + due date |
| `ui/src/components/TaskCard.module.css` | 115 | Card styles with selected state |
| `ui/src/components/QuickCreate.tsx` | 100 | Quick create modal |
| `ui/src/components/QuickCreate.module.css` | 120 | Modal styles |
| `ui/src/components/TaskDetail.tsx` | 160 | Task detail panel with editable props |
| `ui/src/components/TaskDetail.module.css` | 145 | Panel styles |
| `ui/src/App.tsx` | 100 | Main app component with selection state |
| `ui/vitest.config.ts` | 15 | Vitest configuration |
| `ui/src/test/setup.ts` | 5 | Test setup |
| `ui/src/components/TaskCard.test.tsx` | 50 | TaskCard component tests |
| `ui/src/components/Board.test.tsx` | 80 | Board component tests |
| `ui/src/hooks/useTasks.test.ts` | 50 | useTasks hook tests |

**Total new code**: ~1,900+ lines (including CSS)

---

## What You Should Have Now

After completing Phase 2, your project should have:

```
egenskriven/
├── cmd/
│   └── egenskriven/
│       └── main.go              ✓ Updated to serve UI
├── internal/
│   └── ...                      (unchanged from Phase 1)
├── migrations/
│   └── 1_initial.go             (unchanged)
├── ui/
│   ├── src/
│   │   ├── components/
│   │   │   ├── Board.tsx        ✓ Created
│   │   │   ├── Board.module.css ✓ Created
│   │   │   ├── Column.tsx       ✓ Created
│   │   │   ├── Column.module.css ✓ Created
│   │   │   ├── Header.tsx       ✓ Created
│   │   │   ├── Header.module.css ✓ Created
│   │   │   ├── Layout.tsx       ✓ Created
│   │   │   ├── Layout.module.css ✓ Created
│   │   │   ├── QuickCreate.tsx  ✓ Created
│   │   │   ├── QuickCreate.module.css ✓ Created
│   │   │   ├── TaskCard.tsx     ✓ Created (with selection + due date)
│   │   │   ├── TaskCard.module.css ✓ Created (with selected state)
│   │   │   ├── TaskCard.test.tsx ✓ Created
│   │   │   ├── TaskDetail.tsx   ✓ Created (with editable props)
│   │   │   ├── TaskDetail.module.css ✓ Created
│   │   │   └── Board.test.tsx   ✓ Created
│   │   ├── hooks/
│   │   │   ├── useTasks.ts      ✓ Created
│   │   │   └── useTasks.test.ts ✓ Created
│   │   ├── lib/
│   │   │   └── pb.ts            ✓ Created
│   │   ├── styles/
│   │   │   └── tokens.css       ✓ Created
│   │   ├── test/
│   │   │   └── setup.ts         ✓ Created
│   │   ├── types/
│   │   │   └── task.ts          ✓ Created
│   │   ├── App.tsx              ✓ Updated
│   │   └── main.tsx             ✓ Updated
│   ├── dist/                    ✓ Build output (embedded)
│   ├── embed.go                 ✓ Updated with go:embed
│   ├── vite.config.ts           ✓ Created
│   ├── vitest.config.ts         ✓ Created
│   └── package.json             ✓ Updated
├── Makefile                     ✓ Updated with UI targets
└── ...
```

---

## Next Phase

**Phase 3: Full CLI** will add:
- Batch input support (`--stdin`, `--file`)
- Advanced filters (`--label`, `--search`, `--limit`)
- Epic commands and epic support in tasks
- Batch delete operations
- Improved error messages

**Phase 4: Interactive UI** will add:
- Command palette (`Cmd+K`)
- Full keyboard navigation (`J/K`, `H/L`)
- Property picker popovers (`S`, `P`, `T`, `L`)
- Peek preview (`Space`)
- Real-time subscriptions improvements

---

## Troubleshooting

### "pattern dist: no matching files found" when building Go

**Problem**: The `go:embed` directive can't find the `dist/` directory.

**Solution**: Build the React app first:
```bash
cd ui && npm run build && cd ..
go build ./...
```

### Vite dev server shows "Failed to fetch" errors

**Problem**: The proxy isn't forwarding to PocketBase correctly.

**Solution**: Make sure PocketBase is running:
```bash
# Terminal 1
./egenskriven serve

# Terminal 2
cd ui && npm run dev
```

### Tasks don't appear in the UI

**Problem**: The tasks collection doesn't exist or has no records.

**Solution**: 
1. Check if tasks collection exists at `http://localhost:8090/_/`
2. Create a task via CLI: `./egenskriven add "Test task"`
3. Refresh the browser

### CSS styles not applying

**Problem**: The tokens.css file isn't being imported.

**Solution**: Check `ui/src/main.tsx` has the import:
```typescript
import './styles/tokens.css'
```

### Drag and drop not working

**Problem**: @dnd-kit not installed or configured correctly.

**Solution**:
```bash
cd ui
npm install @dnd-kit/core @dnd-kit/sortable @dnd-kit/utilities
```

### "Cannot find module 'pocketbase'" error

**Problem**: PocketBase SDK not installed.

**Solution**:
```bash
cd ui && npm install pocketbase
```

### Real-time updates not working

**Problem**: SSE connection not established.

**Solution**:
1. Check browser console for errors
2. Verify the proxy configuration in `vite.config.ts`
3. Ensure you're accessing via the Vite dev server (port 5173) or the Go server (port 8090), not mixing them

### Build fails with "module not found" for CSS modules

**Problem**: TypeScript doesn't recognize CSS module imports.

**Solution**: Create or update `ui/src/vite-env.d.ts`:
```typescript
/// <reference types="vite/client" />

declare module '*.module.css' {
  const classes: { [key: string]: string }
  export default classes
}
```

---

## Glossary

| Term | Definition |
|------|------------|
| **Vite** | Fast build tool for modern web development |
| **go:embed** | Go directive to include files in the binary at compile time |
| **@dnd-kit** | React drag-and-drop library |
| **CSS Modules** | CSS files where class names are scoped locally |
| **SSE** | Server-Sent Events - one-way real-time updates from server |
| **Design Tokens** | Reusable values for colors, spacing, typography |
| **SPA** | Single Page Application - client-side routing |
| **Proxy** | Forwarding requests from one server to another |
