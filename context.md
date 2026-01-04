# Phase 5: Multi-Board Support - Implementation Context

**Last Updated**: 2026-01-05
**Branch**: `implement-phase-5`
**Status**: Backend complete, Frontend complete, Verification pending

## Overview

This document captures the implementation progress of Phase 5 (Multi-Board Support) to enable seamless handoff. The goal is to allow users to organize tasks into multiple boards with unique prefixes, custom columns, and display IDs like `WRK-123`.

## What Has Been Done

### 1. Database Migrations (Complete)

Three new migrations were created to extend the schema:

| File | Purpose |
|------|---------|
| `migrations/4_boards.go` | Creates `boards` collection with fields: `name`, `prefix`, `columns` (JSON), `color`. Includes unique index on `prefix`. |
| `migrations/5_tasks_board_relation.go` | Adds `board` RelationField to `tasks` collection linking to `boards`. Includes index for performance. |
| `migrations/6_tasks_sequence.go` | Adds `seq` NumberField to `tasks` for per-board sequence numbers used in display IDs. Includes compound index on `(board, seq)`. |

### 2. Board Service (Complete)

Created `internal/board/board.go` with the following functionality:

- **Types**: `Board` struct, `CreateInput` struct
- **Constants**: `DefaultColumns` = `["backlog", "todo", "in_progress", "review", "done"]`
- **Functions**:
  - `Create(app, input)` - Creates a board with validation (prefix uppercase, alphanumeric, unique)
  - `GetByNameOrPrefix(app, ref)` - Flexible lookup by ID, prefix, or name (case-insensitive)
  - `GetAll(app)` - Returns all boards
  - `GetNextSequence(app, boardID)` - Returns next sequence number for a board
  - `FormatDisplayID(prefix, seq)` - Formats display ID (e.g., "WRK-123")
  - `ParseDisplayID(displayID)` - Parses display ID into prefix and sequence
  - `RecordToBoard(record)` - Converts PocketBase record to Board struct
  - `Delete(app, boardID, deleteTasks)` - Deletes board with optional task cleanup

### 3. Config Updates (Complete)

Updated `internal/config/config.go`:

- Added `DefaultBoard string` field to `Config` struct
- This stores the user's preferred default board prefix for CLI commands

### 4. CLI Commands (Complete)

#### Board Commands (`internal/commands/board.go`)

New `board` command with subcommands:

| Command | Description |
|---------|-------------|
| `board list` | Lists all boards, marks current default with `>` |
| `board add "Name" --prefix PRE` | Creates a new board (supports `--color`, `--columns`) |
| `board show <ref>` | Shows board details and task count |
| `board use <ref>` | Sets default board in `.egenskriven/config.json` |
| `board delete <ref>` | Deletes board and tasks (supports `--force`) |

#### Updated Add Command (`internal/commands/add.go`)

- Added `--board` / `-b` flag to specify target board
- Added `resolveBoard()` helper that checks: explicit flag > config default > first board > create "Default" board
- Tasks now get `board` and `seq` fields set on creation
- Output shows display IDs (e.g., "Created: Task title [WRK-1]")
- Batch creation also supports board assignment

#### Updated List Command (`internal/commands/list.go`)

- Added `--board` / `-b` flag to filter by specific board
- Added `--all-boards` flag to show tasks from all boards
- Tasks display with board-prefixed IDs instead of internal IDs
- Uses new `TasksWithBoards()` and `TasksWithFieldsAndBoards()` output methods

#### Updated Root Command (`internal/commands/root.go`)

- Registered `newBoardCmd(app)` in the command tree

### 5. Task Resolver (Complete)

Updated `internal/resolver/resolver.go`:

- Added display ID resolution as **first** resolution attempt
- Resolution order is now:
  1. Display ID (e.g., "WRK-123") - parses prefix, finds board, queries by board+seq
  2. Exact internal ID
  3. ID prefix match
  4. Title substring match

### 6. Output Formatter (Complete)

Updated `internal/output/output.go` with new methods:

- `TasksWithBoards(tasks, boardsMap)` - Outputs tasks with display IDs
- `TasksWithFieldsAndBoards(tasks, fields, boardsMap)` - JSON output with field selection and display IDs
- `printTaskLineWithBoard(task, boardsMap)` - Helper for human-readable output
- `taskToMapWithBoard(task, boardsMap)` - Adds `display_id`, `board`, `seq` to JSON
- `getDisplayID(task, boardsMap)` - Returns display ID or falls back to short ID

### 7. Frontend Hooks (Complete)

#### useBoards Hook (`ui/src/hooks/useBoards.ts`)

- Fetches all boards from PocketBase on mount
- Subscribes to real-time create/update/delete events
- Provides `createBoard(input)` and `deleteBoard(id)` functions
- Returns `{ boards, loading, error, createBoard, deleteBoard }`

#### useCurrentBoard Hook (`ui/src/hooks/useCurrentBoard.tsx`)

- Context provider pattern (`CurrentBoardProvider`)
- Tracks active board state
- Persists selection to localStorage (`egenskriven-current-board` key)
- Auto-syncs when boards are updated or deleted
- Returns `{ currentBoard, setCurrentBoard, loading }`

#### Updated useTasks Hook (`ui/src/hooks/useTasks.ts`)

- Added optional `boardId` parameter to filter tasks
- Re-fetches when boardId changes
- Real-time updates filter by board
- `createTask()` now includes board and seq fields

### 8. Frontend Types (Complete)

#### Board Type (`ui/src/types/board.ts`)

- `Board` interface extending PocketBase `RecordModel`
- `DEFAULT_COLUMNS` constant
- `BOARD_COLORS` array for color picker
- `formatDisplayId(prefix, seq)` helper
- `parseDisplayId(displayId)` helper

#### Updated Task Type (`ui/src/types/task.ts`)

- Added `board?: string` field (board ID relation)
- Added `seq?: number` field (per-board sequence)

### 9. Frontend Components (Complete)

#### Sidebar Component (`ui/src/components/Sidebar.tsx` + `.module.css`)

- Collapsible sidebar with board list
- Current board indicator (highlighted)
- Board color dots
- "New board" button opens modal
- `NewBoardModal` with name, prefix, color picker
- Persists collapsed state to localStorage

#### Updated Layout Component (`ui/src/components/Layout.tsx`)

- Wraps app in `CurrentBoardProvider`
- Integrates Sidebar with collapsed state management
- Updated CSS for sidebar + content layout

#### Updated Board Component (`ui/src/components/Board.tsx`)

- Uses `useCurrentBoard()` to get current board
- Passes `boardId` to `useTasks()` for filtering
- Supports board's custom columns (falls back to defaults)
- Shows "No board selected" message when appropriate
- Passes `currentBoard` to Column and TaskCard

#### Updated Column Component (`ui/src/components/Column.tsx`)

- Accepts `column` as `string` (not just ColumnType) for custom columns
- Added `getColumnDisplayName()` for formatting custom column names
- Passes `currentBoard` to TaskCard

#### Updated TaskCard Component (`ui/src/components/TaskCard.tsx`)

- Accepts `currentBoard` prop
- Shows display ID (e.g., "WRK-123") when board and seq available
- Falls back to short internal ID if no board

#### Updated CommandPalette Component (`ui/src/components/CommandPalette.tsx`)

- Added `'boards'` section type
- Added "SWITCH BOARD" section in UI
- Board commands show current board indicator

#### Updated App Component (`ui/src/App.tsx`)

- Imports and uses `useBoards` and `useCurrentBoard`
- Passes `currentBoard?.id` to `useTasks()`
- Builds board switching commands for CommandPalette
- Clears selection when switching boards

## Git Commits Made

### Backend (Previous Session)
```
49bb038 feat(migrations): add boards collection migration
18b66e3 feat(migrations): add board relation field to tasks
122c1bf feat(migrations): add seq field to tasks for display IDs
55bccc8 feat(board): add board service with CRUD operations
0911967 feat(cli): add board commands (list, add, show, use, delete)
f363453 feat(cli): add --board flag to add command with display IDs
e3dd56e feat(cli): add --board and --all-boards flags to list command
2884c90 feat(resolver): add display ID resolution (PREFIX-NUMBER)
1dd6cd2 docs: add implementation context for phase-5 handoff
```

### Frontend (Current Session)
```
8f388e8 feat(ui): add useBoards hook and Board type
f632c81 feat(ui): add useCurrentBoard hook with localStorage
1c54e75 feat(ui): add boardId filter to useTasks hook
69ad395 feat(ui): add Sidebar component with board list
7726286 feat(ui): integrate Sidebar into Layout component
5de29e7 feat(ui): update Board to use current board context
1d14a11 feat(ui): show display ID on TaskCard component
d8c56f7 feat(ui): add board switching to CommandPalette
```

## What Remains To Be Done

### To Recreate Todo List

Use the following to recreate the todo list for continuity:

```json
[
  {"id": "5.10", "content": "Create board service tests (internal/board/board_test.go) - Test Create, GetByNameOrPrefix, GetNextSequence, ParseDisplayID, Delete", "status": "pending", "priority": "medium"},
  {"id": "5.20", "content": "CLI verification checklist - Test board CRUD, task creation with display IDs, --all-boards, resolver", "status": "pending", "priority": "medium"},
  {"id": "5.21", "content": "UI verification checklist - Test sidebar, board switching, display IDs on cards, command palette", "status": "pending", "priority": "medium"}
]
```

### Remaining Tasks

| ID | Task | Priority | Notes |
|----|------|----------|-------|
| 5.10 | Create board service tests (`internal/board/board_test.go`) | Medium | Test Create, GetByNameOrPrefix, GetNextSequence, ParseDisplayID, Delete. May need testutil helpers. |
| 5.20 | CLI verification checklist | Medium | Test: board create/list/use/delete, task creation with display IDs, --all-boards flag, resolver with display IDs. |
| 5.21 | UI verification checklist | Medium | Test: sidebar shows boards, board switching works, display IDs on cards, command palette board commands. |

### CLI Verification Checklist (5.20)

Run these commands to verify CLI functionality:

```bash
# Build the application
go build -o egenskriven ./cmd/egenskriven

# Start the server (in background or separate terminal)
./egenskriven serve &

# Board CRUD
./egenskriven board add "Work" --prefix WRK --color "#3B82F6"
./egenskriven board add "Personal" --prefix PER --color "#22C55E"
./egenskriven board list                    # Should show both boards
./egenskriven board show WRK                # Should show board details
./egenskriven board use WRK                 # Set default board

# Task creation with display IDs
./egenskriven add "Test task 1"             # Should create WRK-1
./egenskriven add "Test task 2" -b PER      # Should create PER-1
./egenskriven add "Test task 3"             # Should create WRK-2

# List with board filtering
./egenskriven list                          # Show only WRK tasks
./egenskriven list -b PER                   # Show only PER tasks
./egenskriven list --all-boards             # Show all tasks

# Resolver with display IDs
./egenskriven show WRK-1                    # Should resolve and show task
./egenskriven done PER-1                    # Should move task to done

# Cleanup
./egenskriven board delete PER --force      # Delete board and its tasks
```

### UI Verification Checklist (5.21)

1. **Sidebar**
   - [ ] Sidebar appears on left side of screen
   - [ ] Board list shows all boards with colors
   - [ ] Current board is highlighted
   - [ ] Clicking board switches to it
   - [ ] Collapse/expand button works
   - [ ] "New board" button opens modal
   - [ ] Modal validates prefix (uppercase, alphanumeric, unique)
   - [ ] Color picker works

2. **Board View**
   - [ ] Shows only tasks from current board
   - [ ] Columns match board's custom columns (or defaults)
   - [ ] "No board selected" shows when appropriate

3. **Task Cards**
   - [ ] Display ID shows (e.g., "WRK-1") instead of internal ID
   - [ ] Falls back to short ID if no board/seq

4. **Command Palette (Cmd+K)**
   - [ ] "SWITCH BOARD" section appears
   - [ ] All boards listed with names and prefixes
   - [ ] Current board has filled dot indicator
   - [ ] Selecting board switches to it

5. **Real-time Updates**
   - [ ] Creating board via CLI appears in sidebar
   - [ ] Creating task via CLI appears on board
   - [ ] Deleting board switches to another

## Key Implementation Details

### Display ID Format

- Format: `PREFIX-NUMBER` (e.g., `WRK-123`, `PER-1`)
- Prefix: 1-10 uppercase alphanumeric characters, unique per board
- Number: Per-board incrementing sequence starting at 1

### Board Resolution (CLI)

The `board.GetByNameOrPrefix()` function accepts:
- Exact board ID
- Prefix (case-insensitive): "WRK" or "wrk"
- Name (case-insensitive): "Work" or "work"
- Partial name match: "wor" matches "Work" if unique

### Default Board Logic

When creating tasks without `--board` flag:
1. Check `--board` flag
2. Check `config.DefaultBoard` from `.egenskriven/config.json`
3. Use first existing board
4. Create "Default" board with prefix "DEF" if none exist

### UI State Management

- **CurrentBoardProvider**: Wraps Layout to provide board context
- **localStorage keys**:
  - `egenskriven-current-board`: Selected board ID
  - `egenskriven-sidebar-collapsed`: Sidebar collapsed state

### Sequence Number Generation (UI)

When creating tasks from UI, the `useTasks.createTask()` function:
1. Fetches all tasks for the board sorted by `-seq`
2. Takes max seq + 1 (or 1 if no tasks)
3. Includes `board` and `seq` in create payload

## File Structure Reference

```
internal/
  board/
    board.go           # Board service (DONE)
    board_test.go      # Tests (TODO - 5.10)
  commands/
    board.go           # Board CLI commands (DONE)
    add.go             # Updated with --board (DONE)
    list.go            # Updated with --board, --all-boards (DONE)
    root.go            # Registered board command (DONE)
  config/
    config.go          # Added DefaultBoard field (DONE)
  output/
    output.go          # Added board-aware output methods (DONE)
  resolver/
    resolver.go        # Added display ID resolution (DONE)

migrations/
  4_boards.go                  # Boards collection (DONE)
  5_tasks_board_relation.go    # Board field on tasks (DONE)
  6_tasks_sequence.go          # Seq field on tasks (DONE)

ui/src/
  types/
    board.ts           # Board type and helpers (DONE)
    task.ts            # Added board, seq fields (DONE)
  hooks/
    useBoards.ts       # Board CRUD hook (DONE)
    useCurrentBoard.tsx # Current board context (DONE)
    useTasks.ts        # Added boardId filter (DONE)
  components/
    Sidebar.tsx        # Board navigation (DONE)
    Sidebar.module.css # Sidebar styles (DONE)
    Layout.tsx         # Integrated sidebar (DONE)
    Layout.module.css  # Updated layout styles (DONE)
    Board.tsx          # Board filtering (DONE)
    Board.module.css   # Added empty state (DONE)
    Column.tsx         # Custom column names (DONE)
    TaskCard.tsx       # Display IDs (DONE)
    CommandPalette.tsx # Board switching (DONE)
  App.tsx              # Board context usage (DONE)
```

## How To Continue

1. Read this document for full context
2. Recreate the todo list using the JSON in "To Recreate Todo List" section
3. Run verification checklists (5.20, 5.21)
4. Write board service tests if time permits (5.10)
5. Update this document with any issues found

## Verification Commands

```bash
# Build and verify Go code compiles
go build ./...

# Run existing tests
go test ./...

# Build UI
cd ui && npm run build

# Start the application
./egenskriven serve
```
