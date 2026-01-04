# Phase 2 Implementation Context

**Last Updated**: 2026-01-04
**Current Status**: Tasks 2.1-2.7 completed, paused for review before continuing with 2.8-2.15

---

## Overview

EgenSkriven is a local-first kanban board application with a CLI-first design. It uses PocketBase as the backend (SQLite + REST API + real-time subscriptions) and an embedded React UI served from a single Go binary.

**Phase 2 Goal**: Basic web UI with board view, drag-and-drop, and real-time sync with CLI changes.

---

## What Has Been Completed

### Phase 0-1.5 (Prior Work)
- Full CLI implementation with all commands (add, list, show, move, update, delete, init, prime, context, suggest)
- PocketBase integration with task collection schema
- Output formatter with JSON/human modes
- Task resolver (ID, prefix, title matching)
- Position calculator (fractional indexing)
- Blocking relationship support
- Agent integration features (prime, context, suggest commands)

### Phase 2 Progress (Tasks 2.1-2.7)

| Task | Status | Description |
|------|--------|-------------|
| 2.1 | Done | Initialized Vite + React + TypeScript project with dependencies |
| 2.2 | Done | Configured Vite proxy and build output |
| 2.3 | Done | Updated embed.go with go:embed directive |
| 2.4 | Done | Updated main.go with OnServe hook for SPA serving |
| 2.5 | Done | Updated Makefile with UI build targets |
| 2.6 | Done | Created design tokens (CSS variables for dark mode) |
| 2.7 | Done | Created PocketBase client and TypeScript types |

---

## Git Commits Made

```
ec25ae5 feat(ui): initialize Vite React TypeScript project
f91a38b feat(ui): configure Vite proxy and build output
19077db feat(ui): add go:embed directive for React dist
e8d0a5f feat(ui): add OnServe hook to serve embedded React SPA
407a9b4 feat(build): add UI build targets to Makefile
f4f7b3b feat(ui): add design tokens with dark mode styling
2854fc9 feat(ui): add PocketBase client and Task types
```

---

## Current Project Structure

```
EgenSkriven/
├── cmd/egenskriven/
│   └── main.go                    # Updated - serves embedded UI
├── internal/
│   ├── commands/                  # CLI commands (unchanged)
│   ├── config/                    # Config loading (unchanged)
│   ├── output/                    # Output formatting (unchanged)
│   ├── resolver/                  # Task resolution (unchanged)
│   └── testutil/                  # Test helpers (unchanged)
├── migrations/
│   └── 1_initial.go               # Task collection schema
├── ui/
│   ├── dist/                      # Build output (gitignored, embedded)
│   ├── node_modules/              # Dependencies (gitignored)
│   ├── public/
│   │   └── vite.svg
│   ├── src/
│   │   ├── assets/
│   │   │   └── react.svg
│   │   ├── lib/
│   │   │   └── pb.ts              # NEW - PocketBase client
│   │   ├── styles/
│   │   │   └── tokens.css         # NEW - Design tokens
│   │   ├── types/
│   │   │   └── task.ts            # NEW - TypeScript types
│   │   ├── App.tsx                # Updated - placeholder
│   │   └── main.tsx               # Updated - imports tokens.css
│   ├── embed.go                   # Updated - go:embed directive
│   ├── eslint.config.js
│   ├── index.html
│   ├── package.json
│   ├── package-lock.json
│   ├── tsconfig.app.json
│   ├── tsconfig.json
│   ├── tsconfig.node.json
│   └── vite.config.ts             # Updated - proxy config
├── docs/
│   ├── phase-2.md                 # Full implementation guide
│   └── context.md                 # THIS FILE
├── Makefile                       # Updated - UI targets
├── go.mod
└── go.sum
```

---

## Key Files to Understand

### 1. `ui/src/types/task.ts`
Defines TypeScript types matching the database schema:
- `Task` interface with all fields from PocketBase
- `Column`, `Priority`, `TaskType`, `CreatedBy` types
- Helper constants: `COLUMNS`, `COLUMN_NAMES`, `PRIORITIES`, `PRIORITY_NAMES`, `TYPES`, `TYPE_NAMES`

### 2. `ui/src/lib/pb.ts`
PocketBase client initialization:
```typescript
export const pb = new PocketBase('/')
pb.autoCancellation(false)
```

### 3. `ui/src/styles/tokens.css`
CSS custom properties for:
- Background colors (app, sidebar, card, overlay)
- Text colors (primary, secondary, muted, disabled)
- Border colors
- Accent colors (blue default)
- Status colors (backlog, todo, in_progress, review, done)
- Priority colors (urgent, high, medium, low)
- Type colors (bug, feature, chore)
- Typography (font families, sizes, weights, line heights)
- Spacing scale (space-1 through space-10)
- Border radius, shadows, animations
- Layout dimensions (sidebar, column, detail panel widths)

### 4. `ui/vite.config.ts`
Vite configuration with:
- React plugin
- Build output to `dist/`
- Proxy for `/api` and `/_` routes to `localhost:8090`

### 5. `cmd/egenskriven/main.go`
Go entry point that:
- Registers CLI commands
- Auto-imports migrations
- Serves embedded React SPA via `OnServe` hook
- Handles static files and SPA fallback to index.html

### 6. `ui/embed.go`
Go embed directive:
```go
//go:embed all:dist
var distDir embed.FS
var DistFS, _ = fs.Sub(distDir, "dist")
```

### 7. `Makefile`
New UI targets:
- `make build-ui` - Build React UI
- `make dev-ui` - Start Vite dev server (port 5173)
- `make test-ui` - Run UI tests
- `make clean-ui` - Remove UI artifacts
- `make dev-all` - Run both React and Go dev servers
- `make build` - Now depends on `build-ui`

---

## Database Schema (tasks collection)

| Field | Type | Required | Values |
|-------|------|----------|--------|
| title | TextField | Yes | Max 500 chars |
| description | TextField | No | Max 10000 chars |
| type | SelectField | Yes | bug, feature, chore |
| priority | SelectField | Yes | low, medium, high, urgent |
| column | SelectField | Yes | backlog, todo, in_progress, review, done |
| position | NumberField | Yes | Min 0 (fractional) |
| labels | JSONField | No | String array |
| blocked_by | JSONField | No | Task ID array |
| created_by | SelectField | Yes | user, agent, cli |
| created_by_agent | TextField | No | Max 100 chars |
| history | JSONField | No | Activity log array |

---

## Dependencies Installed

### Go
- `github.com/pocketbase/pocketbase` v0.35.0
- `github.com/pocketbase/dbx` v1.11.0
- `github.com/spf13/cobra` v1.10.2
- `github.com/stretchr/testify` v1.11.1

### Node.js (ui/package.json)
- `react` ^19.1.0
- `react-dom` ^19.1.0
- `pocketbase` (latest)
- `@dnd-kit/core` (latest)
- `@dnd-kit/sortable` (latest)
- `@dnd-kit/utilities` (latest)
- Dev: `typescript`, `vite`, `@vitejs/plugin-react`, `eslint`

---

## How to Test Current State

```bash
# Build everything
make build

# Run server
./egenskriven serve

# Open http://localhost:8090 - should show placeholder "EgenSkriven" page with dark background

# Create a task via CLI
./egenskriven add "Test task" --column todo

# Check PocketBase admin at http://localhost:8090/_/
```

---

## Implementation Reference

The full implementation guide is in `docs/phase-2.md`. It contains:
- Complete code for all components
- Step-by-step instructions
- Verification checklists
- Troubleshooting guide

---

## Remaining Tasks

### Task 2.8: Create PocketBase Hooks
**Priority**: High
**File**: `ui/src/hooks/useTasks.ts`

Create a React hook for fetching tasks and subscribing to real-time updates:
- Fetch all tasks on mount with `pb.collection('tasks').getFullList()`
- Subscribe to real-time create/update/delete events
- Provide CRUD operations: `createTask`, `updateTask`, `deleteTask`, `moveTask`
- Handle loading and error states

Reference implementation in `docs/phase-2.md` section 2.8.

---

### Task 2.9: Create Layout Components
**Priority**: Medium
**Files**: 
- `ui/src/components/Layout.tsx`
- `ui/src/components/Layout.module.css`
- `ui/src/components/Header.tsx`
- `ui/src/components/Header.module.css`

Create the main layout shell:
- `Layout`: Wrapper with header and main content area
- `Header`: App title and keyboard shortcut hints (C, Enter, Esc)

Reference implementation in `docs/phase-2.md` section 2.9.

---

### Task 2.10: Create Board Components
**Priority**: High
**Files**:
- `ui/src/components/Board.tsx` + `.module.css`
- `ui/src/components/Column.tsx` + `.module.css`
- `ui/src/components/TaskCard.tsx` + `.module.css`

Create the kanban board:
- `Board`: DnD context, groups tasks by column, handles drag events
- `Column`: Droppable area, displays column header with count
- `TaskCard`: Draggable card with status dot, title, labels, priority, type

Key features:
- Use `@dnd-kit/core` for drag-and-drop
- Group tasks by `column` field
- Sort by `position` within columns
- Show loading/error states

Reference implementation in `docs/phase-2.md` section 2.10.

---

### Task 2.11: Create Quick Create Modal
**Priority**: Medium
**Files**:
- `ui/src/components/QuickCreate.tsx`
- `ui/src/components/QuickCreate.module.css`

Create a modal for quickly creating tasks:
- Opens with `C` key
- Title input (auto-focused)
- Column selector dropdown
- Enter to create, Esc to cancel
- Closes after successful creation

Reference implementation in `docs/phase-2.md` section 2.11.

---

### Task 2.12: Create Task Detail Panel
**Priority**: Medium
**Files**:
- `ui/src/components/TaskDetail.tsx`
- `ui/src/components/TaskDetail.module.css`

Create a slide-in panel for viewing/editing task details:
- Opens when clicking a task or pressing Enter on selected task
- Displays all task properties
- Editable dropdowns for status, priority, type
- Shows labels, blocked_by, metadata (created, updated, created_by)
- Closes with Esc or click outside

Reference implementation in `docs/phase-2.md` section 2.12.

---

### Task 2.13: Wire Everything in App.tsx
**Priority**: High
**File**: `ui/src/App.tsx`

Connect all components:
- Import and use `Layout`, `Board`, `QuickCreate`, `TaskDetail`
- Use `useTasks` hook for data
- Manage state: `isQuickCreateOpen`, `selectedTask`, `selectedTaskId`
- Implement keyboard shortcuts:
  - `C` - Open quick create modal
  - `Enter` - Open detail panel for selected task
  - `Esc` - Close panel/deselect task
- Pass callbacks to components for task operations

Reference implementation in `docs/phase-2.md` section 2.13.

---

### Task 2.14: Setup UI Testing
**Priority**: Medium
**Files**:
- `ui/vitest.config.ts`
- `ui/src/test/setup.ts`
- `ui/src/components/TaskCard.test.tsx`
- `ui/src/components/Board.test.tsx`
- `ui/src/hooks/useTasks.test.ts`

Setup testing infrastructure:
1. Install dependencies:
   ```bash
   npm install -D vitest @testing-library/react @testing-library/jest-dom @testing-library/user-event jsdom @types/testing-library__jest-dom
   ```

2. Create vitest config with jsdom environment

3. Create test setup file importing jest-dom

4. Update package.json scripts:
   ```json
   "test": "vitest",
   "test:run": "vitest run"
   ```

5. Write tests for TaskCard, Board, and useTasks

Reference implementation in `docs/phase-2.md` section 2.14.

---

### Task 2.15: Verification
**Priority**: High

Run through the complete verification checklist from `docs/phase-2.md`:

**Build Verification**:
- [ ] `cd ui && npm run build` produces `ui/dist/`
- [ ] `make build` produces ~35-50MB binary
- [ ] `make clean && make build` works from scratch

**Runtime Verification**:
- [ ] `./egenskriven serve` serves UI at http://localhost:8090
- [ ] PocketBase admin works at http://localhost:8090/_/
- [ ] CLI commands still work

**UI Feature Verification**:
- [ ] Board displays 5 columns (Backlog, Todo, In Progress, Review, Done)
- [ ] `C` key opens quick create modal
- [ ] Tasks can be created via modal
- [ ] Tasks can be dragged between columns
- [ ] Real-time sync: CLI-created tasks appear in UI without refresh
- [ ] Clicking task opens detail panel
- [ ] `Enter` opens detail for selected task
- [ ] `Esc` closes panel
- [ ] Selected task has visual indicator
- [ ] Task properties are editable in detail panel
- [ ] Dark mode styling applied

**Test Verification**:
- [ ] `cd ui && npm run test:run` passes all tests
- [ ] `make test` passes Go tests

**Accessibility Verification**:
- [ ] Focus visible on keyboard navigation
- [ ] Task cards have proper ARIA attributes

---

## Notes for Next Implementer

1. **Follow `docs/phase-2.md`** - It has complete code for each component
2. **Build order matters** - Always build UI before Go binary (`make build` handles this)
3. **Test with CLI** - Create tasks via CLI to verify real-time sync works
4. **Commit after each task** - Use conventional commits, max 70 chars
5. **CSS Modules** - All component styles use `.module.css` for scoped class names
6. **Design tokens** - Use CSS variables from `tokens.css` for all styling

---

## Useful Commands

```bash
# Development
make dev          # Go server with hot reload (Air)
make dev-ui       # Vite dev server on port 5173
make dev-all      # Both servers in parallel

# Building
make build-ui     # Build React only
make build        # Build UI + Go binary

# Testing
make test         # Go tests
make test-ui      # UI tests (after 2.14 is done)

# Cleaning
make clean-ui     # Remove ui/dist and node_modules
make clean        # Full clean including pb_data
```
