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

### Files Created

| File | Purpose |
|------|---------|
| `migrations/9_due_dates.go` | Adds `due_date` DateField to tasks collection |
| `migrations/10_subtasks.go` | Adds `parent` RelationField (self-reference) to tasks |
| `internal/commands/date_parser.go` | Date parsing utility supporting ISO 8601 and relative dates |
| `internal/commands/date_parser_test.go` | Unit tests for date parser |
| `internal/commands/show_test.go` | Unit tests for show command sub-task display |
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
| `internal/output/output.go` | Added `TaskDetailWithSubtasks()` method for displaying task with sub-tasks |
| `ui/src/components/TaskDetail.tsx` | Integrated DatePicker component for due_date editing |
| `ui/src/types/task.ts` | Added `parent?: string` field to Task interface |

---

## What Remains To Be Done

### Next Task to Work On

**p8-25**: Integrate SubtaskList into TaskDetail.tsx

This involves:
1. Import SubtaskList component into TaskDetail.tsx
2. Add SubtaskList after the properties section
3. Pass the tasks array and parent ID
4. Wire up onTaskClick and onToggleComplete handlers

### Remaining High Priority Tasks (7 items)

| ID | Task |
|----|------|
| p8-25 | Integrate SubtaskList into TaskDetail.tsx |
| p8-25-test | UI test for SubtaskList |
| p8-32 | Create export.go command |
| p8-33 | Create import.go command |
| p8-34 | Write import/export tests |
| p8-36 | Final verification: Run all CLI tests |
| p8-37 | Final verification: Run all UI tests |

### Remaining Medium Priority Tasks (19 items)

Epic-related UI (p8-12 through p8-20-test), TaskCard overdue highlighting (p8-21, p8-21-test), MarkdownEditor (p8-26 through p8-28-test), ActivityLog (p8-29 through p8-31-test), EpicPicker integration (p8-35).

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
  {"id": "p8-25", "content": "8.9 UI: Integrate SubtaskList into TaskDetail.tsx", "status": "pending", "priority": "high"},
  {"id": "p8-25-test", "content": "8.9 TEST: Use ui-test-engineer to test SubtaskList component functionality", "status": "pending", "priority": "high"},
  {"id": "p8-26", "content": "8.10 UI: Create MarkdownEditor.tsx component with toolbar, keyboard shortcuts", "status": "pending", "priority": "medium"},
  {"id": "p8-27", "content": "8.10 UI: Create MarkdownEditor.module.css styles", "status": "pending", "priority": "medium"},
  {"id": "p8-28", "content": "8.10 UI: Integrate MarkdownEditor into TaskDetail.tsx for description editing", "status": "pending", "priority": "medium"},
  {"id": "p8-28-test", "content": "8.10 TEST: Use ui-test-engineer to test MarkdownEditor functionality", "status": "pending", "priority": "medium"},
  {"id": "p8-29", "content": "8.11 UI: Create ActivityLog.tsx component for task history display", "status": "pending", "priority": "medium"},
  {"id": "p8-30", "content": "8.11 UI: Create ActivityLog.module.css styles", "status": "pending", "priority": "medium"},
  {"id": "p8-31", "content": "8.11 UI: Integrate ActivityLog into TaskDetail.tsx", "status": "pending", "priority": "medium"},
  {"id": "p8-31-test", "content": "8.11 TEST: Use ui-test-engineer to test ActivityLog display", "status": "pending", "priority": "medium"},
  {"id": "p8-32", "content": "8.12 CLI: Create internal/commands/export.go command with JSON and CSV formats", "status": "pending", "priority": "high"},
  {"id": "p8-33", "content": "8.12 CLI: Create internal/commands/import.go command with merge/replace strategies", "status": "pending", "priority": "high"},
  {"id": "p8-34", "content": "8.12 Tests: Write export_test.go and import_test.go integration tests", "status": "pending", "priority": "high"},
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
4. Start with task p8-25 (next pending high-priority task: integrate SubtaskList into TaskDetail)
5. After each task, commit with conventional commit format
6. For UI tasks, run ui-test-engineer after implementation (see `docs/ui-test-instructions.md`)

---

## Useful Commands

```bash
# Build and verify Go code
cd /home/ramtinj/personal-workspace/EgenSkriven && go build ./...

# Run Go tests
go test ./internal/commands/... -v

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
```

---

## Progress Summary

**Completed**: 17/37 tasks (46%)
- All CLI due date features (p8-1 through p8-4, p8-2-test)
- All CLI sub-task features (p8-5 through p8-8, p8-8-test)
- DatePicker UI component (p8-9 through p8-11-test)
- SubtaskList UI component (p8-22 through p8-24)

**Remaining High Priority**: 7 tasks
**Remaining Medium Priority**: 19 tasks
