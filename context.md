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
19. `feat(ui): add EpicPicker component for task epic assignment` - Created Epic type, useEpics hook, EpicPicker component and styles, integrated into TaskDetail
20. `feat(ui): add EpicList component for sidebar epic filtering` - Created EpicList.tsx, EpicList.module.css, updated Sidebar.tsx, Layout.tsx, App.tsx
21. `feat(ui): add EpicDetail view with progress and task list` - Created useEpic.ts hook, EpicDetail.tsx, EpicDetail.module.css, integrated into App.tsx
22. `feat(ui): add overdue and due-today highlighting to TaskCard` - Updated TaskCard.tsx and TaskCard.module.css

### Files Created (This Session)

| File | Purpose |
|------|---------|
| `ui/src/components/EpicList.tsx` | Sidebar component showing epics with task counts, filtering support |
| `ui/src/components/EpicList.module.css` | CSS module styles for EpicList |
| `ui/src/hooks/useEpic.ts` | Hook for single epic CRUD operations with real-time updates |
| `ui/src/components/EpicDetail.tsx` | Full epic detail view with progress bar, task list by column, edit/delete |
| `ui/src/components/EpicDetail.module.css` | CSS module styles for EpicDetail |

### Files Modified (This Session)

| File | Changes |
|------|---------|
| `ui/src/components/Sidebar.tsx` | Added EpicList integration, new props for tasks/epic filtering |
| `ui/src/components/Layout.tsx` | Added props to pass tasks and epic filter state to Sidebar |
| `ui/src/App.tsx` | Added epic filter state, EpicDetail state, handleSelectEpic callback, EpicDetail component |
| `ui/src/components/TaskCard.tsx` | Added isOverdue/isDueToday logic, OverdueIcon, conditional styling |
| `ui/src/components/TaskCard.module.css` | Added .overdue and .dueToday styles for due date highlighting |

### All Files Created (Full Phase 8)

| File | Purpose |
|------|---------|
| `migrations/9_due_dates.go` | Adds `due_date` DateField to tasks collection |
| `migrations/10_subtasks.go` | Adds `parent` RelationField (self-reference) to tasks |
| `internal/commands/date_parser.go` | Date parsing utility supporting ISO 8601 and relative dates |
| `internal/commands/date_parser_test.go` | Unit tests for date parser |
| `internal/commands/show_test.go` | Unit tests for show command sub-task display |
| `internal/commands/export.go` | CLI command to export tasks/boards/epics to JSON or CSV |
| `internal/commands/import.go` | CLI command to import data with merge/replace strategies |
| `internal/commands/export_test.go` | Integration tests for export command |
| `internal/commands/import_test.go` | Integration tests for import command |
| `ui/src/components/DatePicker.tsx` | Calendar date picker with month navigation and quick shortcuts |
| `ui/src/components/DatePicker.module.css` | CSS module styles for DatePicker |
| `ui/src/components/SubtaskList.tsx` | Component to display sub-tasks with progress bar |
| `ui/src/components/SubtaskList.module.css` | CSS module styles for SubtaskList |
| `ui/src/types/epic.ts` | Epic type definition with EPIC_COLORS constant |
| `ui/src/hooks/useEpics.ts` | Hook for fetching all epics with real-time updates and CRUD |
| `ui/src/hooks/useEpic.ts` | Hook for single epic CRUD operations |
| `ui/src/components/EpicPicker.tsx` | Dropdown component for selecting epic on a task |
| `ui/src/components/EpicPicker.module.css` | CSS module styles for EpicPicker |
| `ui/src/components/EpicList.tsx` | Sidebar component for epic filtering with task counts |
| `ui/src/components/EpicList.module.css` | CSS module styles for EpicList |
| `ui/src/components/EpicDetail.tsx` | Epic detail view with progress and task list |
| `ui/src/components/EpicDetail.module.css` | CSS module styles for EpicDetail |

---

## What Remains To Be Done

### Next Task to Work On

**p8-20-test**: Test EpicDetail view functionality (using ui-test-engineer)

Then continue with:
- p8-21-test: Test TaskCard overdue highlighting
- p8-26: Create MarkdownEditor.tsx component

### Remaining Tasks (10 items)

| ID | Task | Category | Status |
|----|------|----------|--------|
| p8-20-test | Test EpicDetail view functionality | Epic UI (8.7) | pending |
| p8-21-test | Test TaskCard overdue highlighting | TaskCard (8.8) | pending |
| p8-26 | Create MarkdownEditor.tsx component with toolbar | MarkdownEditor (8.10) | pending |
| p8-27 | Create MarkdownEditor.module.css styles | MarkdownEditor (8.10) | pending |
| p8-28 | Integrate MarkdownEditor into TaskDetail.tsx | MarkdownEditor (8.10) | pending |
| p8-28-test | Test MarkdownEditor functionality | MarkdownEditor (8.10) | pending |
| p8-29 | Create ActivityLog.tsx component for task history | ActivityLog (8.11) | pending |
| p8-30 | Create ActivityLog.module.css styles | ActivityLog (8.11) | pending |
| p8-31 | Integrate ActivityLog into TaskDetail.tsx | ActivityLog (8.11) | pending |
| p8-31-test | Test ActivityLog display | ActivityLog (8.11) | pending |
| p8-36 | Final Verification: Run all CLI tests | Final | pending |
| p8-37 | Final Verification: Run all UI tests | Final | pending |

---

## Full Todo List (for next session)

Copy this JSON to recreate the todo list using TodoWrite:

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
  {"id": "p8-13", "content": "8.5 UI: Create useEpics hook in hooks/ for fetching all epics with real-time updates", "status": "completed", "priority": "medium"},
  {"id": "p8-12", "content": "8.5 UI: Create EpicPicker.tsx component for task epic assignment", "status": "completed", "priority": "medium"},
  {"id": "p8-14", "content": "8.5 UI: Create EpicPicker.module.css styles", "status": "completed", "priority": "medium"},
  {"id": "p8-35", "content": "8.5 UI: Integrate EpicPicker into TaskDetail.tsx for epic assignment", "status": "completed", "priority": "medium"},
  {"id": "p8-14-test", "content": "8.5 TEST: Use ui-test-engineer to test EpicPicker component functionality", "status": "completed", "priority": "medium"},
  {"id": "p8-15", "content": "8.6 UI: Create EpicList.tsx component for sidebar with task counts", "status": "completed", "priority": "medium"},
  {"id": "p8-16", "content": "8.6 UI: Create EpicList.module.css styles", "status": "completed", "priority": "medium"},
  {"id": "p8-17", "content": "8.6 UI: Integrate EpicList into Sidebar.tsx", "status": "completed", "priority": "medium"},
  {"id": "p8-17-test", "content": "8.6 TEST: Use ui-test-engineer to test EpicList sidebar integration", "status": "completed", "priority": "medium"},
  {"id": "p8-19", "content": "8.7 UI: Create useEpic(id) hook for single epic CRUD operations", "status": "completed", "priority": "medium"},
  {"id": "p8-18", "content": "8.7 UI: Create EpicDetail.tsx view for epic progress and task list", "status": "completed", "priority": "medium"},
  {"id": "p8-20", "content": "8.7 UI: Create EpicDetail.module.css styles", "status": "completed", "priority": "medium"},
  {"id": "p8-20-test", "content": "8.7 TEST: Use ui-test-engineer to test EpicDetail view functionality", "status": "pending", "priority": "medium"},
  {"id": "p8-21", "content": "8.8 UI: Update TaskCard.tsx - Add overdue highlighting logic for due dates", "status": "completed", "priority": "medium"},
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
- `due_date` already typed in TypeScript and displayed in TaskCard (now with overdue highlighting)
- ReactMarkdown already integrated for description preview
- Epics collection and CLI commands already exist

### Epic UI Implementation Details

**EpicList (8.6)** - Sidebar epic filtering:
- `ui/src/components/EpicList.tsx` - Shows all epics with task counts
- Uses `useEpics()` hook for epic data
- Computes task counts client-side using `useMemo`
- Supports "All Tasks", individual epics, and "No Epic" filter
- Integrates with filter store to add/remove epic filters
- Props passed through Layout -> Sidebar -> EpicList

**EpicDetail (8.7)** - Full epic view:
- `ui/src/hooks/useEpic.ts` - Single epic CRUD with real-time updates
- `ui/src/components/EpicDetail.tsx` - Modal overlay with:
  - Color bar, title, description
  - Edit mode with title/description/color picker
  - Progress bar (completed/total tasks)
  - Task list grouped by column
  - Delete with confirmation
- Integrated into App.tsx with `epicDetailId` state

**TaskCard Overdue (8.8)**:
- Added `isOverdue` and `isDueToday` computed values
- OverdueIcon SVG for visual indicator
- CSS classes `.overdue` (red) and `.dueToday` (orange)
- Only shows overdue if task is not done

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

---

## How to Continue

1. Read this context file
2. Read `docs/phase-8.md` for detailed implementation specs
3. Use the TodoWrite tool to recreate the todo list from the JSON above
4. Next tasks to work on:
   - p8-20-test: Test EpicDetail (start dev server first!)
   - p8-21-test: Test TaskCard overdue highlighting
   - p8-26: Create MarkdownEditor.tsx component
5. After each feature group, commit with conventional commit format
6. For UI tests, follow `docs/ui-test-instructions.md`

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

# Kill processes on common ports (if needed)
lsof -ti:5173 | xargs kill -9 2>/dev/null

# Test export command
./egenskriven export --format json
./egenskriven export --format csv

# Test import command (dry-run)
./egenskriven import backup.json --dry-run
```

---

## Progress Summary

**Completed**: 35/47 tasks (74%)

### Completed Features:
- **8.1 Due Dates Migration** (p8-1): migrations/9_due_dates.go
- **8.2 Due Dates CLI** (p8-2, p8-3, p8-4, p8-2-test): date_parser.go, --due flag, filter flags, tests
- **8.3 Sub-tasks Migration & CLI** (p8-5 through p8-8-test): migrations/10_subtasks.go, --parent flag, show command
- **8.4 DatePicker UI** (p8-9 through p8-11-test): DatePicker component, styles, integration, tested
- **8.5 EpicPicker UI** (p8-12, p8-13, p8-14, p8-35, p8-14-test): Epic type, useEpics hook, EpicPicker, integration, tested
- **8.6 EpicList UI** (p8-15, p8-16, p8-17, p8-17-test): EpicList component, styles, sidebar integration, tested
- **8.7 EpicDetail UI** (p8-18, p8-19, p8-20): useEpic hook, EpicDetail component, styles - NEEDS TESTING
- **8.8 TaskCard Overdue** (p8-21): Overdue highlighting logic - NEEDS TESTING
- **8.9 SubtaskList UI** (p8-22 through p8-25-test): SubtaskList component, styles, integration, tested
- **8.12 Export/Import CLI** (p8-32, p8-33, p8-34): export.go, import.go, tests

### Remaining:
- **8.7 EpicDetail Test** (p8-20-test): UI test needed - 1 task
- **8.8 TaskCard Test** (p8-21-test): UI test needed - 1 task
- **8.10 MarkdownEditor** (p8-26, p8-27, p8-28, p8-28-test): Rich text editing - 4 tasks
- **8.11 ActivityLog** (p8-29, p8-30, p8-31, p8-31-test): Task history display - 4 tasks
- **Final Verification** (p8-36, p8-37): Run all tests - 2 tasks

---

## Session Summary (2026-01-06)

### What Was Done This Session:

1. **Recreated todo list** from previous context.md

2. **Completed 8.6 EpicList UI** (sidebar epic filtering):
   - Created `ui/src/components/EpicList.tsx` with task counts per epic
   - Created `ui/src/components/EpicList.module.css` matching dark theme
   - Updated `ui/src/components/Sidebar.tsx` to include EpicList
   - Updated `ui/src/components/Layout.tsx` to pass tasks and epic props
   - Updated `ui/src/App.tsx` with epic filter state and handler
   - Committed: `feat(ui): add EpicList component for sidebar epic filtering`

3. **Completed 8.7 EpicDetail UI** (epic detail view):
   - Created `ui/src/hooks/useEpic.ts` for single epic CRUD
   - Created `ui/src/components/EpicDetail.tsx` with progress bar, task list, edit/delete
   - Created `ui/src/components/EpicDetail.module.css` with full styling
   - Integrated EpicDetail into App.tsx
   - Committed: `feat(ui): add EpicDetail view with progress and task list`

4. **Completed 8.8 TaskCard Overdue** (overdue highlighting):
   - Added `isOverdue` and `isDueToday` computed values to TaskCard.tsx
   - Added OverdueIcon SVG component
   - Added `.overdue` and `.dueToday` CSS classes
   - Committed: `feat(ui): add overdue and due-today highlighting to TaskCard`

### Current State:
- All implementation for 8.6, 8.7, 8.8 is complete
- UI tests for EpicDetail (p8-20-test) and TaskCard overdue (p8-21-test) still pending
- Next implementation work: 8.10 MarkdownEditor

### Git Status:
- Branch: `implement-phase-8`
- 4 commits ahead of origin
- All changes committed
