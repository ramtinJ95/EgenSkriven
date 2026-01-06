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

### Files Created

| File | Purpose |
|------|---------|
| `migrations/9_due_dates.go` | Adds `due_date` DateField to tasks collection |
| `migrations/10_subtasks.go` | Adds `parent` RelationField (self-reference) to tasks |
| `internal/commands/date_parser.go` | Date parsing utility supporting ISO 8601 and relative dates |
| `internal/commands/date_parser_test.go` | Unit tests for date parser |

### Files Modified

| File | Changes |
|------|---------|
| `internal/commands/add.go` | Added `--due` flag, `--parent` flag, `TaskInput.DueDate`, `TaskInput.Parent`, `resolveTaskByID()` function |
| `internal/commands/list.go` | Added `--due-before`, `--due-after`, `--has-due`, `--no-due`, `--has-parent`, `--no-parent` flags |

---

## What Remains To Be Done

### Full Todo List (for next session)

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
  {"id": "p8-8", "content": "8.3 CLI: Update show.go to display sub-tasks of a parent task", "status": "pending", "priority": "high"},
  {"id": "p8-9", "content": "8.4 UI: Create DatePicker.tsx component with calendar, navigation, quick shortcuts", "status": "pending", "priority": "high"},
  {"id": "p8-10", "content": "8.4 UI: Create DatePicker.module.css styles (following existing CSS Modules pattern)", "status": "pending", "priority": "high"},
  {"id": "p8-11", "content": "8.4 UI: Integrate DatePicker into TaskDetail.tsx for due_date editing", "status": "pending", "priority": "high"},
  {"id": "p8-11-test", "content": "8.4 TEST: Use ui-test-engineer to test DatePicker component functionality", "status": "pending", "priority": "high"},
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
  {"id": "p8-22", "content": "8.9 UI: Create SubtaskList.tsx component for task detail panel", "status": "pending", "priority": "high"},
  {"id": "p8-23", "content": "8.9 UI: Create SubtaskList.module.css styles", "status": "pending", "priority": "high"},
  {"id": "p8-24", "content": "8.9 UI: Update Task type in types/task.ts to include parent?: string field", "status": "pending", "priority": "high"},
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

### Next Task to Work On

**p8-8**: Update `show.go` to display sub-tasks of a parent task

This involves:
1. Reading `internal/commands/show.go`
2. After displaying task details, fetching sub-tasks where `parent = task.Id`
3. Displaying them in a "Sub-tasks:" section

### Remaining High Priority Tasks (9 items)

| ID | Task |
|----|------|
| p8-8 | Update show.go to display sub-tasks |
| p8-9 | Create DatePicker.tsx component |
| p8-10 | Create DatePicker.module.css styles |
| p8-11 | Integrate DatePicker into TaskDetail.tsx |
| p8-11-test | UI test for DatePicker |
| p8-22-25 | SubtaskList component (4 tasks) |
| p8-32 | Create export.go command |
| p8-33 | Create import.go command |
| p8-34 | Write import/export tests |

### Remaining Medium Priority Tasks (19 items)

Epic-related UI (p8-12 through p8-20-test), TaskCard overdue highlighting (p8-21, p8-21-test), MarkdownEditor (p8-26 through p8-28-test), ActivityLog (p8-29 through p8-31-test), EpicPicker integration (p8-35).

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

### Commit Convention
All commits should be conventional commits, max 70 characters:
- `feat(cli):` for CLI features
- `feat(db):` for migrations
- `feat(ui):` for UI features
- `test(cli):` or `test(ui):` for tests

---

## How to Continue

1. Read this context file
2. Read `docs/phase-8.md` for detailed implementation specs
3. Use the TodoWrite tool to recreate the todo list from the JSON above
4. Start with task p8-8 (next pending high-priority task)
5. After each task, commit with conventional commit format
6. For UI tasks, run ui-test-engineer after implementation

---

## Useful Commands

```bash
# Build and verify Go code
cd /home/ramtinj/personal-workspace/EgenSkriven && go build ./...

# Run Go tests
go test ./internal/commands/... -v

# Run specific test
go test ./internal/commands/... -run TestParseDate -v

# UI tests (from ui/ directory)
cd ui && npm test
```
