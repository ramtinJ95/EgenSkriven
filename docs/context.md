# Phase 3 Implementation Context

**Date**: 2026-01-04
**Branch**: `implement-phase-3`
**Status**: In Progress (11/16 tasks completed)

## Overview

Phase 3 extends the Core CLI with professional-grade features including batch operations, epic management, advanced filtering, and a version command. This document provides context for continuing the implementation.

## What Has Been Done

### 1. Migrations

**Files created:**
- `migrations/2_epics.go` - Creates the `epics` collection with fields:
  - `title` (required, max 200 chars)
  - `description` (max 5000 chars)
  - `color` (hex format pattern `^#[0-9A-Fa-f]{6}$`)
  - Auto timestamps (created, updated)

- `migrations/3_epic_relation.go` - Adds `epic` relation field to `tasks` collection:
  - Single relation to epics collection
  - `CascadeDelete: false` (tasks remain when epic is deleted)

**Pattern used**: `m.Register()` + `init()` pattern matching `1_initial.go`

### 2. Epic Commands

**File**: `internal/commands/epic.go`

Implements the `epic` command with subcommands:
- `epic list` - List all epics with task counts
- `epic add <title>` - Create epic with optional `--color` and `--description`
- `epic show <epic>` - Show epic details and linked tasks
- `epic delete <epic>` - Delete epic with confirmation (supports `--force`)

**Helper functions added:**
- `resolveEpic(app, ref)` - Resolves epic by ID, ID prefix, or title (case-insensitive)
- `getEpicTaskCount(app, epicID)` - Returns count of tasks linked to an epic
- `isValidHexColor(color)` - Validates `#RRGGBB` format

### 3. Task-Epic Integration

**Modified files:**

- `internal/commands/add.go`:
  - Added `--epic` / `-e` flag to link tasks to epics on creation
  - Added batch input support via `--stdin` and `--file` flags
  - New `TaskInput` struct for batch JSON parsing
  - New `addBatch()` function handling JSON lines and JSON array formats
  - New `defaultString()` helper function

- `internal/commands/list.go`:
  - Added `--epic` / `-e` filter to filter tasks by epic
  - Added `--label` / `-l` filter (repeatable) for label filtering
  - Added `--limit` flag for maximum results
  - Added `--sort` flag for custom sort order (e.g., `-priority,position`)
  - Refactored to use `RecordQuery` for proper DB-level limit/sort

- `internal/commands/delete.go`:
  - Added `--stdin` flag to read task references from stdin
  - Changed `Args` to `MinimumNArgs(0)` to support stdin-only usage

### 4. Version Command

**File**: `internal/commands/version.go`

- Displays version, build date, git commit, Go version, OS/arch
- Variables `Version`, `BuildDate`, `GitCommit` set via ldflags at build time
- Supports `--json` output

### 5. Output Enhancements

**File**: `internal/output/output.go`

- Added `ErrorWithSuggestion(code, message, suggestion, data)` method
- Outputs helpful suggestions in both human and JSON formats

### 6. Command Registration

**File**: `internal/commands/root.go`

Added registration for Phase 3 commands:
```go
// Phase 3 commands
app.RootCmd.AddCommand(newEpicCmd(app))
app.RootCmd.AddCommand(newVersionCmd())
```

## Commits Made

1. `feat(migrations): add epics collection migration`
2. `feat(migrations): add epic relation field to tasks`
3. `feat(commands): add epic management commands`
4. `feat(add): add --epic flag to link tasks to epics`
5. `feat(list): add --epic filter to filter tasks by epic`
6. `feat(add): add batch input support via --stdin and --file`
7. `feat(delete): add --stdin flag for batch deletion`
8. `feat(list): add --label, --limit, --sort flags`
9. `feat(commands): add version command`
10. `feat(output): add ErrorWithSuggestion method`
11. `feat(commands): register epic and version commands`

## What Remains To Be Done

### Task 3.12: Makefile ldflags (Low Priority)

**File**: `Makefile`

Add version embedding via ldflags to the build target:

```makefile
VERSION ?= dev
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LDFLAGS := -X github.com/ramtinJ95/EgenSkriven/internal/commands.Version=$(VERSION)
LDFLAGS += -X github.com/ramtinJ95/EgenSkriven/internal/commands.BuildDate=$(BUILD_DATE)
LDFLAGS += -X github.com/ramtinJ95/EgenSkriven/internal/commands.GitCommit=$(GIT_COMMIT)

build: build-ui
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o egenskriven ./cmd/egenskriven
```

### Task 3.13: Epic Tests (Medium Priority)

**File to create**: `internal/commands/epic_test.go`

Tests needed:
- `TestIsValidHexColor` - Test valid/invalid hex colors (`#3B82F6`, `#aabbcc`, `red`, `#FFF`, etc.)
- `TestResolveEpic` - Test resolution by exact ID, ID prefix, title, partial title, not found, ambiguous
- `TestGetEpicTaskCount` - Test count with 0 tasks, multiple tasks

Use `testutil.NewTestApp(t)` and `testutil.CreateTestCollection()` for test setup.

### Task 3.14: Batch Operation Tests (Medium Priority)

**File to create**: `internal/commands/add_batch_test.go`

Tests needed:
- `TestParseBatchInput_JSONLines` - Test parsing JSON lines format
- `TestParseBatchInput_JSONArray` - Test parsing JSON array format
- `TestDefaultString` - Test the defaultString helper function

### Task 3.15: Run Tests and Fix Issues (High Priority)

```bash
make test
```

Fix any compilation errors or test failures that arise.

### Task 3.16: Build and Manual Verification (High Priority)

```bash
make build
```

Manual verification checklist:
- [ ] `./egenskriven epic list`
- [ ] `./egenskriven epic add "Test Epic" --color "#3B82F6"`
- [ ] `./egenskriven epic show "Test Epic"`
- [ ] `./egenskriven epic delete "Test Epic" --force`
- [ ] `./egenskriven add "Task" --epic "Test Epic"`
- [ ] `./egenskriven list --epic "Test Epic"`
- [ ] `echo '{"title":"Batch Task"}' | ./egenskriven add --stdin`
- [ ] `./egenskriven add --file <json-file>`
- [ ] `./egenskriven list --label <label> --limit 5 --sort "-priority"`
- [ ] `./egenskriven version`
- [ ] `./egenskriven version --json`

## File Summary

| File | Status | Purpose |
|------|--------|---------|
| `migrations/2_epics.go` | DONE | Epics collection |
| `migrations/3_epic_relation.go` | DONE | Epic relation in tasks |
| `internal/commands/epic.go` | DONE | Epic CRUD commands |
| `internal/commands/version.go` | DONE | Version command |
| `internal/commands/add.go` | DONE | Added --epic, --stdin, --file |
| `internal/commands/delete.go` | DONE | Added --stdin |
| `internal/commands/list.go` | DONE | Added --epic, --label, --limit, --sort |
| `internal/commands/root.go` | DONE | Registered epic, version commands |
| `internal/output/output.go` | DONE | Added ErrorWithSuggestion |
| `Makefile` | TODO | Add version ldflags |
| `internal/commands/epic_test.go` | TODO | Epic tests |
| `internal/commands/add_batch_test.go` | TODO | Batch operation tests |

## Key Implementation Details

### Batch Input Formats

**JSON Lines** (one object per line):
```json
{"title": "Task 1", "type": "bug", "priority": "high"}
{"title": "Task 2", "column": "todo"}
```

**JSON Array**:
```json
[
  {"title": "Task 1", "type": "bug"},
  {"title": "Task 2", "priority": "high"}
]
```

### TaskInput Struct

```go
type TaskInput struct {
    ID          string   `json:"id,omitempty"`
    Title       string   `json:"title"`
    Description string   `json:"description,omitempty"`
    Type        string   `json:"type,omitempty"`
    Priority    string   `json:"priority,omitempty"`
    Column      string   `json:"column,omitempty"`
    Labels      []string `json:"labels,omitempty"`
    Epic        string   `json:"epic,omitempty"`
}
```

### Epic Resolution Order

1. Exact ID match
2. ID prefix match (must be unique)
3. Title substring match (case-insensitive, must be unique)

### List Command Query Changes

Changed from `FindAllRecords` to `RecordQuery` to support:
- `query.OrderBy()` for custom sorting
- `query.Limit()` for result limiting
- Proper DB-level filtering instead of in-memory

## How to Continue

1. Read this document to understand current state
2. Check out the `implement-phase-3` branch
3. Run `git log --oneline -15` to see recent commits
4. Pick up from Task 3.12 (Makefile) or any remaining task
5. After each task, commit with conventional commit format (max 70 chars)
6. Run `make test` after completing tests
7. Run `make build` for final verification

## References

- Phase 3 specification: `docs/phase-3.md`
- Test utilities: `internal/testutil/testutil.go`
- Existing tests for patterns: `internal/commands/*_test.go`
