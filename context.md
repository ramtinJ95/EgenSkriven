# Phase 8 Implementation Context

**Last Updated**: 2026-01-06
**Branch**: `implement-phase-8`
**Phase Document**: `docs/phase-8.md`

## Summary

Phase 8 adds advanced features to EgenSkriven: Epic UI, Due Dates, Sub-tasks, Markdown Editor, Activity Log, and Import/Export functionality.

---

## What Has Been Completed

### Commits Made (in order)

1. `feat(db): add due_date field to tasks collection` - Created `migrations/9_due_dates.go`
2. `feat(cli): add date parser utility for due dates` - Created `internal/commands/date_parser.go`
3. `feat(cli): add --due flag to add command` - Updated `internal/commands/add.go`
4. `feat(cli): add due date filter flags to list command` - Updated `internal/commands/list.go` with `--due-before`, `--due-after`, `--has-due`, `--no-due`
5. `test(cli): add date parser unit tests` - Created `internal/commands/date_parser_test.go`
6. `feat(db): add parent field to tasks for sub-tasks` - Created `migrations/10_subtasks.go`
7. `feat(cli): add --parent flag for creating sub-tasks` - Updated `internal/commands/add.go` with `--parent` flag and `resolveTaskByID` function
8. `feat(cli): add --has-parent, --no-parent flags to list` - Updated `internal/commands/list.go`
9. `feat(cli): add sub-task display to show command` - Updated `internal/commands/show.go` and `internal/output/output.go` with `TaskDetailWithSubtasks` method
10. `test(cli): add show command sub-task display tests` - Created `internal/commands/show_test.go`
11. `feat(ui): add DatePicker component with calendar and shortcuts` - Created `ui/src/components/DatePicker.tsx` and `DatePicker.module.css`
12. `feat(ui): integrate DatePicker into TaskDetail for due date editing` - Updated `ui/src/components/TaskDetail.tsx`
13. `fix(ui): fix DatePicker nested button and escape key propagation` - Fixed HTML nesting issue and event propagation
14. `feat(ui): add SubtaskList component and parent field to Task type` - Created `ui/src/components/SubtaskList.tsx`, `SubtaskList.module.css`, updated `ui/src/types/task.ts`
15. `feat(ui): integrate SubtaskList into TaskDetail panel` - Updated `ui/src/components/TaskDetail.tsx` and `ui/src/App.tsx` to pass tasks array and handlers
16. `feat(cli): add export command with JSON and CSV formats` - Created `internal/commands/export.go`
17. `feat(cli): add import command with merge/replace strategies` - Created `internal/commands/import.go`
18. `test(cli): add export and import integration tests` - Created `internal/commands/export_test.go` and `internal/commands/import_test.go`

### Files Created

| File | Purpose |
|------|---------|
| `migrations/9_due_dates.go` | Adds `due_date` DateField to tasks collection |
| `migrations/10_subtasks.go` | Adds `parent` RelationField (self-reference) to tasks |
| `internal/commands/date_parser.go` | Date parsing utility supporting ISO 8601 and relative dates |
| `internal/commands/date_parser_test.go` | Unit tests for date parser |
| `internal/commands/show_test.go` | Unit tests for show command sub-task display |
| `internal/commands/export.go` | CLI command to export tasks/boards/epics to JSON or CSV |
| `internal/commands/import.go` | CLI command to import data with merge/replace strategies |
| `internal/commands/export_test.go` | Integration tests for export command (JSON, CSV, board filtering) |
| `internal/commands/import_test.go` | Integration tests for import command (merge, replace, dry-run) |
| `ui/src/components/DatePicker.tsx` | Calendar date picker with month navigation and quick shortcuts |
| `ui/src/components/DatePicker.module.css` | CSS module styles for DatePicker |
| `ui/src/components/SubtaskList.tsx` | Component to display sub-tasks with progress bar |
| `ui/src/components/SubtaskList.module.css` | CSS module styles for SubtaskList |

### Files Modified

| File | Changes |
|------|---------|
| `internal/commands/add.go` | Added `--due` flag, `--parent` flag, `TaskInput.DueDate`, `TaskInput.Parent`, `resolveTaskByID()` function |
| `internal/commands/list.go` | Added `--due-before`, `--due-after`, `--has-due`, `--no-due`, `--has-parent`, `--no-parent` flags |
| `internal/commands/show.go` | Added sub-task query and display logic using `TaskDetailWithSubtasks` |
| `internal/commands/root.go` | Registered `newExportCmd` and `newImportCmd` commands |
| `internal/output/output.go` | Added `TaskDetailWithSubtasks()` method for displaying task with sub-tasks |
| `ui/src/components/TaskDetail.tsx` | Integrated DatePicker and SubtaskList components, added `tasks` and `onTaskClick` props |
| `ui/src/components/TaskDetail.test.tsx` | Updated tests to include new required `tasks` prop |
| `ui/src/App.tsx` | Pass `tasks` array and `onTaskClick` handler to TaskDetail |
| `ui/src/types/task.ts` | Added `parent?: string` field to Task interface |

---

## What Remains To Be Done

### Next Task to Work On

**p8-12**: Create EpicPicker.tsx component for task epic assignment

This is the start of the Epic UI features (8.5, 8.6, 8.7). The recommended order is:
1. p8-13: Create useEpics hook first (data fetching)
2. p8-12: Create EpicPicker component
3. p8-14: Create EpicPicker styles
4. p8-14-test: Test EpicPicker

### Remaining High Priority Tasks (2 items)

| ID | Task |
|----|------|
| p8-36 | Final verification: Run all CLI tests |
| p8-37 | Final verification: Run all UI tests |

### Remaining Medium Priority Tasks (17 items)

| ID | Task | Category |
|----|------|----------|
| p8-12 | Create EpicPicker.tsx component for task epic assignment | Epic UI (8.5) |
| p8-13 | Create useEpics hook in hooks/ for fetching all epics | Epic UI (8.5) |
| p8-14 | Create EpicPicker.module.css styles | Epic UI (8.5) |
| p8-14-test | Test EpicPicker component functionality | Epic UI (8.5) |
| p8-15 | Create EpicList.tsx component for sidebar with task counts | Epic UI (8.6) |
| p8-16 | Create EpicList.module.css styles | Epic UI (8.6) |
| p8-17 | Integrate EpicList into Sidebar.tsx | Epic UI (8.6) |
| p8-17-test | Test EpicList sidebar integration | Epic UI (8.6) |
| p8-18 | Create EpicDetail.tsx view for epic progress and task list | Epic UI (8.7) |
| p8-19 | Create useEpic(id) hook for single epic CRUD operations | Epic UI (8.7) |
| p8-20 | Create EpicDetail.module.css styles | Epic UI (8.7) |
| p8-20-test | Test EpicDetail view functionality | Epic UI (8.7) |
| p8-21 | Update TaskCard.tsx - Add overdue highlighting logic | TaskCard (8.8) |
| p8-21-test | Test TaskCard overdue highlighting | TaskCard (8.8) |
| p8-26 | Create MarkdownEditor.tsx component with toolbar | MarkdownEditor (8.10) |
| p8-27 | Create MarkdownEditor.module.css styles | MarkdownEditor (8.10) |
| p8-28 | Integrate MarkdownEditor into TaskDetail.tsx | MarkdownEditor (8.10) |
| p8-28-test | Test MarkdownEditor functionality | MarkdownEditor (8.10) |
| p8-29 | Create ActivityLog.tsx component for task history display | ActivityLog (8.11) |
| p8-30 | Create ActivityLog.module.css styles | ActivityLog (8.11) |
| p8-31 | Integrate ActivityLog into TaskDetail.tsx | ActivityLog (8.11) |
| p8-31-test | Test ActivityLog display | ActivityLog (8.11) |
| p8-35 | Integrate EpicPicker into TaskDetail.tsx for epic assignment | EpicPicker Integration (8.13) |

---

## Full Todo List (for next session)

Copy this JSON to recreate the todo list:

```json
[
  {"id": "p8-1", "content": "8.1 Migration: Create migrations/9_due_dates.go - Add due_date DateField to tasks collection", "status": "completed", "priority": "high"},
  {"id": "p8-2", "content": "8.2 CLI: Create internal/commands/date_parser.go - Date parsing utility for ISO 8601, relative dates (today, tomorrow, next week)", "status": "completed", "priority": "high"},
  {"id": "p8-3", "content": "8.2 CLI: Add --due flag to add.go command", "status": "completed", "priority": "high"},
  {"id": "p8-4", "content": "8.2 CLI: Add --due-before, --due-after, --has-due, --no-due flags to list.go", "status": "completed", "priority": "high"},
  {"id": "p8-2-test", "content": "8.2 Tests: Write date_parser_test.go unit tests", "status": "completed", "priority": "high"},
  {"id": "p8-5", "content": "8.3 Migration: Create migrations/10_subtasks.go - Add parent RelationField to tasks (self-reference)", "status": "completed", "priority": "high"},
  {"id": "p8-6", "content": "8.3 CLI: Add --parent flag to add.go for creating sub-tasks", "status": "completed", "priority": "high"},
  {"id": "p8-7", "content": "8.3 CLI: Add --has-parent, --no-parent flags to list.go", "status": "completed", "priority": "high"},
  {"id": "p8-8", "content": "8.3 CLI: Update show.go to display sub-tasks of a parent task", "status": "completed", "priority": "high"},
  {"id": "p8-8-test", "content": "8.3 Tests: Write show_test.go for sub-task display functionality", "status": "completed", "priority": "high"},
  {"id": "p8-9", "content": "8.4 UI: Create DatePicker.tsx component with calendar, navigation, quick shortcuts", "status": "completed", "priority": "high"},
  {"id": "p8-10", "content": "8.4 UI: Create DatePicker.module.css styles (following existing CSS Modules pattern)", "status": "completed", "priority": "high"},
  {"id": "p8-11", "content": "8.4 UI: Integrate DatePicker into TaskDetail.tsx for due_date editing", "status": "completed", "priority": "high"},
  {"id": "p8-11-test", "content": "8.4 TEST: Use ui-test-engineer to test DatePicker component functionality", "status": "completed", "priority": "high"},
  {"id": "p8-12", "content": "8.5 UI: Create EpicPicker.tsx component for task epic assignment", "status": "pending", "priority": "medium"},
  {"id": "p8-13", "content": "8.5 UI: Create useEpics hook in hooks/ for fetching all epics with real-time updates", "status": "pending", "priority": "medium"},
  {"id": "p8-14", "content": "8.5 UI: Create EpicPicker.module.css styles", "status": "pending", "priority": "medium"},
  {"id": "p8-14-test", "content": "8.5 TEST: Use ui-test-engineer to test EpicPicker component functionality", "status": "pending", "priority": "medium"},
  {"id": "p8-15", "content": "8.6 UI: Create EpicList.tsx component for sidebar with task counts", "status": "pending", "priority": "medium"},
  {"id": "p8-16", "content": "8.6 UI: Create EpicList.module.css styles", "status": "pending", "priority": "medium"},
  {"id": "p8-17", "content": "8.6 UI: Integrate EpicList into Sidebar.tsx", "status": "pending", "priority": "medium"},
  {"id": "p8-17-test", "content": "8.6 TEST: Use ui-test-engineer to test EpicList sidebar integration", "status": "pending", "priority": "medium"},
  {"id": "p8-18", "content": "8.7 UI: Create EpicDetail.tsx view for epic progress and task list", "status": "pending", "priority": "medium"},
  {"id": "p8-19", "content": "8.7 UI: Create useEpic(id) hook for single epic CRUD operations", "status": "pending", "priority": "medium"},
  {"id": "p8-20", "content": "8.7 UI: Create EpicDetail.module.css styles", "status": "pending", "priority": "medium"},
  {"id": "p8-20-test", "content": "8.7 TEST: Use ui-test-engineer to test EpicDetail view functionality", "status": "pending", "priority": "medium"},
  {"id": "p8-21", "content": "8.8 UI: Update TaskCard.tsx - Add overdue highlighting logic for due dates", "status": "pending", "priority": "medium"},
  {"id": "p8-21-test", "content": "8.8 TEST: Use ui-test-engineer to test TaskCard overdue highlighting", "status": "pending", "priority": "medium"},
  {"id": "p8-22", "content": "8.9 UI: Create SubtaskList.tsx component for task detail panel", "status": "completed", "priority": "high"},
  {"id": "p8-23", "content": "8.9 UI: Create SubtaskList.module.css styles", "status": "completed", "priority": "high"},
  {"id": "p8-24", "content": "8.9 UI: Update Task type in types/task.ts to include parent?: string field", "status": "completed", "priority": "high"},
  {"id": "p8-25", "content": "8.9 UI: Integrate SubtaskList into TaskDetail.tsx", "status": "completed", "priority": "high"},
  {"id": "p8-25-test", "content": "8.9 TEST: Use ui-test-engineer to test SubtaskList component functionality", "status": "completed", "priority": "high"},
  {"id": "p8-26", "content": "8.10 UI: Create MarkdownEditor.tsx component with toolbar, keyboard shortcuts", "status": "pending", "priority": "medium"},
  {"id": "p8-27", "content": "8.10 UI: Create MarkdownEditor.module.css styles", "status": "pending", "priority": "medium"},
  {"id": "p8-28", "content": "8.10 UI: Integrate MarkdownEditor into TaskDetail.tsx for description editing", "status": "pending", "priority": "medium"},
  {"id": "p8-28-test", "content": "8.10 TEST: Use ui-test-engineer to test MarkdownEditor functionality", "status": "pending", "priority": "medium"},
  {"id": "p8-29", "content": "8.11 UI: Create ActivityLog.tsx component for task history display", "status": "pending", "priority": "medium"},
  {"id": "p8-30", "content": "8.11 UI: Create ActivityLog.module.css styles", "status": "pending", "priority": "medium"},
  {"id": "p8-31", "content": "8.11 UI: Integrate ActivityLog into TaskDetail.tsx", "status": "pending", "priority": "medium"},
  {"id": "p8-31-test", "content": "8.11 TEST: Use ui-test-engineer to test ActivityLog display", "status": "pending", "priority": "medium"},
  {"id": "p8-32", "content": "8.12 CLI: Create internal/commands/export.go command with JSON and CSV formats", "status": "completed", "priority": "high"},
  {"id": "p8-33", "content": "8.12 CLI: Create internal/commands/import.go command with merge/replace strategies", "status": "completed", "priority": "high"},
  {"id": "p8-34", "content": "8.12 Tests: Write export_test.go and import_test.go integration tests", "status": "completed", "priority": "high"},
  {"id": "p8-35", "content": "8.13 UI: Integrate EpicPicker into TaskDetail.tsx for epic assignment", "status": "pending", "priority": "medium"},
  {"id": "p8-36", "content": "Final Verification: Run all CLI tests (go test ./...)", "status": "pending", "priority": "high"},
  {"id": "p8-37", "content": "Final Verification: Run all UI tests (npm test in ui/)", "status": "pending", "priority": "high"}
]
```

---

## Key Implementation Notes

### CSS Pattern
The UI uses **CSS Modules** (`.module.css` files), NOT plain `.css` files as shown in the phase-8.md document. Follow the existing pattern in `ui/src/components/`.

### Migration Numbering
Migrations are numbered sequentially. Existing: 1-8. New: 9 (due_dates), 10 (subtasks).

### Existing Functionality to Leverage
- `history` field already exists in tasks (JSON array) - just needs UI display
- `due_date` already typed in TypeScript and displayed in TaskCard (read-only)
- ReactMarkdown already integrated for description preview
- Epics collection and CLI commands already exist

### Export/Import Command Details

**Export Command** (`internal/commands/export.go`):
- Supports JSON and CSV formats (`--format json` or `--format csv`)
- Can filter by board (`--board <name>`)
- Can output to file (`--output <file>`) or stdout
- JSON includes boards, epics, and tasks with full metadata
- CSV exports only tasks in flat table format

**Import Command** (`internal/commands/import.go`):
- Two strategies: `merge` (skip existing, default) and `replace` (overwrite)
- Supports `--dry-run` to preview without making changes
- Imports boards, epics, and tasks from JSON backup
- Preserves original IDs for referential integrity

### Export/Import Test Notes
- PocketBase requires record IDs to be at least 15 characters
- Test IDs use format like `board1testid001`, `task1testid0001`
- `out.Error()` calls `os.Exit()` which terminates the test process, so error path tests verify the underlying operations directly

### Testing Approach
After each UI feature, use `ui-test-engineer` agent to create/run Playwright tests.

**IMPORTANT**: Before calling ui-test-engineer:
1. Start the dev server first: `cd ui && npm run dev &`
2. Wait for server to be ready (check port 5173, 5174, etc.)
3. Include the correct URL in the test prompt
4. Follow the template in `docs/ui-test-instructions.md`

### Commit Convention
All commits should be conventional commits, max 70 characters:
- `feat(cli):` for CLI features
- `feat(db):` for migrations
- `feat(ui):` for UI features
- `test(cli):` or `test(ui):` for tests
- `fix(ui):` for bug fixes

### Known Issues Found During Testing
1. **DatePicker database issue**: The `due_date` field migration needs to be applied to the running database instance. The migration file exists but may not be applied if using an existing database.

---

## How to Continue

1. Read this context file
2. Read `docs/phase-8.md` for detailed implementation specs
3. Use the TodoWrite tool to recreate the todo list from the JSON above
4. Start with task p8-12 (next pending task: EpicPicker component)
5. Recommended order for Epic UI:
   - p8-13 (useEpics hook) -> p8-12 (EpicPicker) -> p8-14 (styles) -> p8-14-test
   - p8-15 (EpicList) -> p8-16 (styles) -> p8-17 (sidebar integration) -> p8-17-test
   - p8-19 (useEpic hook) -> p8-18 (EpicDetail) -> p8-20 (styles) -> p8-20-test
6. After each task, commit with conventional commit format
7. For UI tasks, run ui-test-engineer after implementation (see `docs/ui-test-instructions.md`)

---

## Useful Commands

```bash
# Build and verify Go code
cd /home/ramtinj/personal-workspace/EgenSkriven && go build ./...

# Run Go tests
go test ./internal/commands/... -v

# Run all Go tests
go test ./...

# Run specific test
go test ./internal/commands/... -run TestShowCommand -v

# UI development (from ui/ directory)
cd ui && npm run dev

# UI build
cd ui && npm run build

# UI tests (from ui/ directory)
cd ui && npm test

# Check dev server status
curl -s -o /dev/null -w "%{http_code}" http://localhost:5173

# Test export command
./egenskriven export --format json
./egenskriven export --format csv

# Test import command (dry-run)
./egenskriven import backup.json --dry-run
```

---

## Progress Summary

**Completed**: 22/37 tasks (59%)
- All CLI due date features (p8-1 through p8-4, p8-2-test)
- All CLI sub-task features (p8-5 through p8-8, p8-8-test)
- DatePicker UI component (p8-9 through p8-11-test)
- SubtaskList UI component (p8-22 through p8-25-test)
- Export/Import CLI commands with tests (p8-32, p8-33, p8-34)

**Remaining High Priority**: 2 tasks (p8-36, p8-37) - Final verification
**Remaining Medium Priority**: 17 tasks
- Epic UI (p8-12 through p8-20-test): 9 tasks
- TaskCard overdue (p8-21, p8-21-test): 2 tasks
- MarkdownEditor (p8-26 through p8-28-test): 4 tasks
- ActivityLog (p8-29 through p8-31-test): 4 tasks
- EpicPicker integration (p8-35): 1 task

---

## Recent Session Summary (This Session)

In this session we:
1. Recreated the todo list from context.md
2. Completed p8-34: Created export_test.go and import_test.go integration tests
   - Tests for JSON and CSV export formats
   - Tests for board filtering in export
   - Tests for merge and replace import strategies
   - Tests for dry-run mode
   - Tests for helper functions (getExportStringSlice, findExportBoardByNameOrPrefix)
   - Fixed PocketBase ID length requirement (15+ chars)
3. All CLI tests pass (`go test ./...`)

**Current State**: All high-priority CLI tasks are complete. The remaining work is primarily UI components for Epic management, TaskCard overdue highlighting, MarkdownEditor, and ActivityLog.
