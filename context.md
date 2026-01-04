# Phase 5: Multi-Board Support - Implementation Context

**Last Updated**: 2026-01-04
**Branch**: `implement-phase-5`
**Status**: Backend complete, Frontend pending

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

## Git Commits Made

```
49bb038 feat(migrations): add boards collection migration
18b66e3 feat(migrations): add board relation field to tasks
122c1bf feat(migrations): add seq field to tasks for display IDs
55bccc8 feat(board): add board service with CRUD operations
0911967 feat(cli): add board commands (list, add, show, use, delete)
f363453 feat(cli): add --board flag to add command with display IDs
e3dd56e feat(cli): add --board and --all-boards flags to list command
2884c90 feat(resolver): add display ID resolution (PREFIX-NUMBER)
```

## What Remains To Be Done

### Backend (Go)

| ID | Task | Priority | Notes |
|----|------|----------|-------|
| 5.10 | Create board service tests (`internal/board/board_test.go`) | Medium | Test Create, GetByNameOrPrefix, GetNextSequence, ParseDisplayID, Delete. May need testutil helpers. |

### Frontend (React/TypeScript)

| ID | Task | Priority | Notes |
|----|------|----------|-------|
| 5.12 | Create `useBoards` hook (`ui/src/hooks/useBoards.ts`) | High | Fetch all boards, subscribe to real-time changes, provide createBoard/deleteBoard functions. |
| 5.13 | Create `useCurrentBoard` hook (`ui/src/hooks/useCurrentBoard.ts`) | High | Track active board state, persist to localStorage, provide setCurrentBoard function. |
| 5.19 | Update `useTasks` hook (`ui/src/hooks/useTasks.ts`) | High | Add optional `boardId` parameter to filter tasks by board. |
| 5.11 | Create Sidebar component (`ui/src/components/Sidebar.tsx` + `.module.css`) | High | Board list, current board indicator, "New board" button, collapsible. Include NewBoardModal. |
| 5.18 | Update Layout component (`ui/src/components/Layout.tsx`) | High | Integrate Sidebar, manage collapsed state. |
| 5.14 | Update Board component (`ui/src/components/Board.tsx`) | High | Use `useCurrentBoard`, filter tasks by board, use board's custom columns if defined. |
| 5.15 | Update TaskCard component (`ui/src/components/TaskCard.tsx`) | High | Show display ID (PREFIX-SEQ) instead of internal ID substring. |
| 5.16 | Update CommandPalette (`ui/src/components/CommandPalette.tsx`) | Medium | Add board switching commands ("Go to board: X"). |

### Verification

| ID | Task | Priority | Notes |
|----|------|----------|-------|
| 5.20 | CLI verification checklist | Medium | Test: board create/list/use/delete, task creation with display IDs, --all-boards flag, resolver with display IDs. |
| 5.21 | UI verification checklist | Medium | Test: sidebar shows boards, board switching works, display IDs on cards, command palette board commands. |

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

### UI Architecture Notes

- PocketBase client is at `ui/src/lib/pb.ts` (simple export, not a hook)
- Components use CSS Modules (`.module.css` pattern)
- Existing hooks: `useTasks`, `useKeyboard`, `useSelection`
- No existing board-related UI code

## File Structure Reference

```
internal/
  board/
    board.go           # Board service (DONE)
    board_test.go      # Tests (TODO)
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
  hooks/
    useTasks.ts        # Needs boardId filter (TODO)
    useBoards.ts       # New hook (TODO)
    useCurrentBoard.ts # New hook (TODO)
  components/
    Sidebar.tsx        # New component (TODO)
    Sidebar.module.css # New styles (TODO)
    Layout.tsx         # Integrate sidebar (TODO)
    Board.tsx          # Filter by board (TODO)
    TaskCard.tsx       # Display IDs (TODO)
    CommandPalette.tsx # Board commands (TODO)
```

## How To Continue

1. Read this document and `docs/phase-5.md` for full context
2. Recreate the todo list from the "What Remains To Be Done" section
3. Start with the hooks (5.12, 5.13, 5.19) as they're dependencies for components
4. Then build the Sidebar (5.11) and integrate it into Layout (5.18)
5. Update Board and TaskCard components (5.14, 5.15)
6. Add CommandPalette board commands (5.16)
7. Run verification checklists (5.20, 5.21)
8. Write tests if time permits (5.10)

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
