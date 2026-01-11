# Phase 0: Schema Foundation

> **Parent Document**: [ai-workflow-plan.md](./ai-workflow-plan.md)  
> **Phase**: 0 of 7  
> **Status**: Not Started  
> **Estimated Effort**: 1-2 days  
> **Prerequisites**: None

## Overview

This phase establishes the database schema foundation for the AI-Human collaborative workflow feature. All subsequent phases depend on these migrations being completed successfully.

**What we're building:**
- Add `need_input` to the task column enum
- Add `agent_session` JSON field to tasks
- Create new `comments` collection
- Create new `sessions` collection
- Add `resume_mode` field to boards

**What we're NOT building yet:**
- CLI commands (Phase 1)
- Session linking logic (Phase 2)
- Resume functionality (Phase 3)
- UI components (Phase 5)

---

## Prerequisites

Before starting this phase:

1. Understand PocketBase migration system
2. Review existing migrations in `migrations/` directory
3. Ensure you can run `egenskriven serve` successfully
4. Have a test database you can reset

---

## Tasks

### Task 0.1: Create Migration for `need_input` Column

**File**: `migrations/12_need_input_column.go`

**Description**: Update the `column` field on the `tasks` collection to include `need_input` as a valid value.

**Implementation**:

```go
package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Find tasks collection
		tasks, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}

		// Get the column field and update its values
		columnField := tasks.Fields.GetByName("column")
		if columnField == nil {
			return fmt.Errorf("column field not found on tasks collection")
		}

		selectField, ok := columnField.(*core.SelectField)
		if !ok {
			return fmt.Errorf("column field is not a select field")
		}

		// Add need_input to valid values
		selectField.Values = []string{
			"backlog",
			"todo",
			"in_progress",
			"need_input", // NEW
			"review",
			"done",
		}

		return app.Save(tasks)
	}, func(app core.App) error {
		// Rollback: remove need_input from valid values
		tasks, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}

		columnField := tasks.Fields.GetByName("column")
		if columnField == nil {
			return nil // Field doesn't exist, nothing to rollback
		}

		selectField, ok := columnField.(*core.SelectField)
		if !ok {
			return nil
		}

		// Restore original values (without need_input)
		selectField.Values = []string{
			"backlog",
			"todo",
			"in_progress",
			"review",
			"done",
		}

		// Note: If any tasks are in need_input state, this rollback will fail
		// That's intentional - don't rollback if data exists
		return app.Save(tasks)
	})
}
```

**Verification**:
```bash
# After running migrations
egenskriven move <task-id> need_input
# Should succeed without error

# Verify in database
sqlite3 ~/.local/share/egenskriven/data.db "SELECT column FROM tasks WHERE id='<task-id>'"
# Should show: need_input
```

---

### Task 0.2: Create Migration for `agent_session` Field

**File**: `migrations/13_agent_session_field.go`

**Description**: Add `agent_session` JSON field to tasks collection to store the current linked agent session.

**Schema for `agent_session`**:
```typescript
interface AgentSession {
    tool: "opencode" | "claude-code" | "codex";
    ref: string;           // Session/thread ID (UUID or path)
    ref_type: "uuid" | "path";
    working_dir: string;   // Absolute path to project directory
    linked_at: string;     // ISO 8601 timestamp
}
```

**Example value**:
```json
{
    "tool": "opencode",
    "ref": "550e8400-e29b-41d4-a716-446655440000",
    "ref_type": "uuid",
    "working_dir": "/home/user/my-project",
    "linked_at": "2026-01-07T10:30:00Z"
}
```

**Implementation**:

```go
package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		tasks, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}

		// Check if field already exists
		if tasks.Fields.GetByName("agent_session") != nil {
			return nil // Already exists, skip
		}

		// Add agent_session JSON field
		tasks.Fields.Add(&core.JSONField{
			Name:    "agent_session",
			MaxSize: 10000, // 10KB should be plenty for session metadata
		})

		return app.Save(tasks)
	}, func(app core.App) error {
		// Rollback: remove agent_session field
		tasks, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}

		field := tasks.Fields.GetByName("agent_session")
		if field == nil {
			return nil // Field doesn't exist, nothing to rollback
		}

		tasks.Fields.RemoveByName("agent_session")
		return app.Save(tasks)
	})
}
```

**Verification**:
```bash
# Check field exists in schema
sqlite3 ~/.local/share/egenskriven/data.db ".schema tasks" | grep agent_session

# Test setting value via API (or direct SQL for now)
sqlite3 ~/.local/share/egenskriven/data.db \
  "UPDATE tasks SET agent_session = '{\"tool\":\"opencode\",\"ref\":\"test-123\",\"ref_type\":\"uuid\",\"working_dir\":\"/tmp\",\"linked_at\":\"2026-01-07T10:00:00Z\"}' WHERE id='<task-id>'"
```

---

### Task 0.3: Create Migration for `comments` Collection

**File**: `migrations/14_comments_collection.go`

**Description**: Create new `comments` collection for storing task comments/discussions.

**Schema**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | auto | yes | PocketBase auto-generated ID |
| `task` | relation | yes | Relation to tasks collection |
| `content` | text | yes | Comment text (max 50,000 chars) |
| `author_type` | select | yes | "human" or "agent" |
| `author_id` | text | no | Username, agent name, or empty |
| `metadata` | json | no | Additional data (mentions, session ref) |
| `created` | autodate | yes | Auto-set on creation |

**Implementation**:

```go
package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Check if collection already exists
		existing, _ := app.FindCollectionByNameOrId("comments")
		if existing != nil {
			return nil // Already exists, skip
		}

		// Find tasks collection for relation
		tasks, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return fmt.Errorf("tasks collection not found: %w", err)
		}

		// Create comments collection
		collection := core.NewBaseCollection("comments")

		// Task relation (required, cascade delete)
		collection.Fields.Add(&core.RelationField{
			Name:          "task",
			Required:      true,
			MaxSelect:     1,
			CollectionId:  tasks.Id,
			CascadeDelete: true,
		})

		// Comment content (required, max 50KB)
		collection.Fields.Add(&core.TextField{
			Name:     "content",
			Required: true,
			Max:      50000,
		})

		// Author type (required)
		collection.Fields.Add(&core.SelectField{
			Name:     "author_type",
			Required: true,
			Values:   []string{"human", "agent"},
		})

		// Author identifier (optional)
		collection.Fields.Add(&core.TextField{
			Name:     "author_id",
			Required: false,
			Max:      200,
		})

		// Metadata JSON (optional)
		// Used for: mentions array, session reference, etc.
		collection.Fields.Add(&core.JSONField{
			Name:    "metadata",
			MaxSize: 50000,
		})

		// Auto-timestamp on creation
		collection.Fields.Add(&core.AutodateField{
			Name:     "created",
			OnCreate: true,
		})

		// Add indexes for common queries
		collection.Indexes = []string{
			"CREATE INDEX idx_comments_task ON comments (task)",
			"CREATE INDEX idx_comments_created ON comments (created)",
		}

		// Set collection rules (adjust based on your auth setup)
		// For now, allow all operations (CLI-first)
		collection.ListRule = nil   // Anyone can list
		collection.ViewRule = nil   // Anyone can view
		collection.CreateRule = nil // Anyone can create
		collection.UpdateRule = nil // Anyone can update
		collection.DeleteRule = nil // Anyone can delete

		return app.Save(collection)
	}, func(app core.App) error {
		// Rollback: delete comments collection
		collection, err := app.FindCollectionByNameOrId("comments")
		if err != nil {
			return nil // Collection doesn't exist, nothing to rollback
		}
		return app.Delete(collection)
	})
}
```

**Verification**:
```bash
# Check collection exists
sqlite3 ~/.local/share/egenskriven/data.db ".tables" | grep comments

# Check schema
sqlite3 ~/.local/share/egenskriven/data.db ".schema comments"

# Test creating a comment via API
curl -X POST http://localhost:8090/api/collections/comments/records \
  -H "Content-Type: application/json" \
  -d '{
    "task": "<task-id>",
    "content": "Test comment",
    "author_type": "human",
    "author_id": "test-user"
  }'
```

---

### Task 0.4: Create Migration for `sessions` Collection

**File**: `migrations/15_sessions_collection.go`

**Description**: Create new `sessions` collection for tracking agent session history.

**Schema**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | auto | yes | PocketBase auto-generated ID |
| `task` | relation | yes | Relation to tasks collection |
| `tool` | select | yes | "opencode", "claude-code", or "codex" |
| `external_ref` | text | yes | Session/thread ID from the tool |
| `ref_type` | select | yes | "uuid" or "path" |
| `working_dir` | text | yes | Project directory path |
| `status` | select | yes | "active", "paused", "completed", "abandoned" |
| `created` | autodate | yes | When session was linked |
| `ended_at` | date | no | When session ended (if applicable) |

**Implementation**:

```go
package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Check if collection already exists
		existing, _ := app.FindCollectionByNameOrId("sessions")
		if existing != nil {
			return nil
		}

		// Find tasks collection for relation
		tasks, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return fmt.Errorf("tasks collection not found: %w", err)
		}

		collection := core.NewBaseCollection("sessions")

		// Task relation
		collection.Fields.Add(&core.RelationField{
			Name:          "task",
			Required:      true,
			MaxSelect:     1,
			CollectionId:  tasks.Id,
			CascadeDelete: true,
		})

		// Tool identifier
		collection.Fields.Add(&core.SelectField{
			Name:     "tool",
			Required: true,
			Values:   []string{"opencode", "claude-code", "codex"},
		})

		// External session reference (UUID or path)
		collection.Fields.Add(&core.TextField{
			Name:     "external_ref",
			Required: true,
			Max:      500,
		})

		// Reference type
		collection.Fields.Add(&core.SelectField{
			Name:     "ref_type",
			Required: true,
			Values:   []string{"uuid", "path"},
		})

		// Working directory
		collection.Fields.Add(&core.TextField{
			Name:     "working_dir",
			Required: true,
			Max:      1000,
		})

		// Session status
		collection.Fields.Add(&core.SelectField{
			Name:     "status",
			Required: true,
			Values:   []string{"active", "paused", "completed", "abandoned"},
		})

		// Auto-timestamp on creation
		collection.Fields.Add(&core.AutodateField{
			Name:     "created",
			OnCreate: true,
		})

		// End timestamp (optional)
		collection.Fields.Add(&core.DateField{
			Name: "ended_at",
		})

		// Indexes
		collection.Indexes = []string{
			"CREATE INDEX idx_sessions_task ON sessions (task)",
			"CREATE INDEX idx_sessions_status ON sessions (status)",
			"CREATE INDEX idx_sessions_external_ref ON sessions (external_ref)",
		}

		// Rules (open for CLI-first approach)
		collection.ListRule = nil
		collection.ViewRule = nil
		collection.CreateRule = nil
		collection.UpdateRule = nil
		collection.DeleteRule = nil

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("sessions")
		if err != nil {
			return nil
		}
		return app.Delete(collection)
	})
}
```

**Verification**:
```bash
# Check collection exists
sqlite3 ~/.local/share/egenskriven/data.db ".tables" | grep sessions

# Check schema
sqlite3 ~/.local/share/egenskriven/data.db ".schema sessions"

# Test creating a session
curl -X POST http://localhost:8090/api/collections/sessions/records \
  -H "Content-Type: application/json" \
  -d '{
    "task": "<task-id>",
    "tool": "opencode",
    "external_ref": "test-session-123",
    "ref_type": "uuid",
    "working_dir": "/home/user/project",
    "status": "active"
  }'
```

---

### Task 0.5: Create Migration for `resume_mode` on Boards

**File**: `migrations/16_board_resume_mode.go`

**Description**: Add `resume_mode` field to boards collection for configuring how blocked tasks are resumed.

**Schema**:

| Field | Type | Required | Default | Values |
|-------|------|----------|---------|--------|
| `resume_mode` | select | no | "command" | "manual", "command", "auto" |

**Implementation**:

```go
package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		boards, err := app.FindCollectionByNameOrId("boards")
		if err != nil {
			return err
		}

		// Check if field already exists
		if boards.Fields.GetByName("resume_mode") != nil {
			return nil
		}

		// Add resume_mode select field
		boards.Fields.Add(&core.SelectField{
			Name:     "resume_mode",
			Required: false,
			Values:   []string{"manual", "command", "auto"},
		})

		return app.Save(boards)
	}, func(app core.App) error {
		boards, err := app.FindCollectionByNameOrId("boards")
		if err != nil {
			return err
		}

		field := boards.Fields.GetByName("resume_mode")
		if field == nil {
			return nil
		}

		boards.Fields.RemoveByName("resume_mode")
		return app.Save(boards)
	})
}
```

**Verification**:
```bash
# Check field exists
sqlite3 ~/.local/share/egenskriven/data.db ".schema boards" | grep resume_mode

# Test setting value
sqlite3 ~/.local/share/egenskriven/data.db \
  "UPDATE boards SET resume_mode = 'auto' WHERE id='<board-id>'"
```

---

### Task 0.6: Update `ValidColumns` Constant

**File**: `internal/commands/root.go`

**Description**: Update the `ValidColumns` variable to include `need_input`.

**Before**:
```go
var ValidColumns = []string{"backlog", "todo", "in_progress", "review", "done"}
```

**After**:
```go
var ValidColumns = []string{"backlog", "todo", "in_progress", "need_input", "review", "done"}
```

**Find the exact location**:
```bash
grep -n "ValidColumns" internal/commands/root.go
```

**Verification**:
```bash
# After change, this should work
egenskriven move <task-id> need_input

# And list should show it
egenskriven list --column need_input
```

---

## Testing Checklist

Before considering this phase complete, verify:

### Schema Tests

- [ ] All 5 migrations run without errors
- [ ] Migrations are idempotent (running twice doesn't fail)
- [ ] Rollbacks work correctly (test each one)
- [ ] `need_input` is valid column value for tasks
- [ ] `agent_session` field accepts valid JSON
- [ ] `comments` collection exists with correct schema
- [ ] `sessions` collection exists with correct schema
- [ ] `resume_mode` field exists on boards

### Relation Tests

- [ ] Creating a comment with valid task ID works
- [ ] Creating a comment with invalid task ID fails
- [ ] Deleting a task cascades to delete its comments
- [ ] Creating a session with valid task ID works
- [ ] Deleting a task cascades to delete its sessions

### Index Tests

- [ ] Query comments by task is fast (check query plan)
- [ ] Query sessions by task is fast
- [ ] Query sessions by external_ref is fast

### CLI Tests

- [ ] `egenskriven move <task> need_input` works
- [ ] `egenskriven list` shows tasks in need_input column
- [ ] `egenskriven show <task>` displays agent_session if set

---

## Rollback Procedure

If you need to rollback this phase:

```bash
# 1. Stop the server
pkill egenskriven

# 2. Backup current database
cp ~/.local/share/egenskriven/data.db ~/.local/share/egenskriven/data.db.backup

# 3. Rollback migrations in reverse order
# This requires manual intervention or a rollback command
# PocketBase doesn't have automatic rollback, so you may need to:

# Option A: Restore from backup before migrations
cp ~/.local/share/egenskriven/data.db.pre-phase0 ~/.local/share/egenskriven/data.db

# Option B: Manually run rollback SQL
sqlite3 ~/.local/share/egenskriven/data.db << 'EOF'
-- Remove resume_mode from boards
-- (This is complex with PocketBase, may need to recreate collection)

-- Drop sessions collection
DROP TABLE IF EXISTS sessions;

-- Drop comments collection  
DROP TABLE IF EXISTS comments;

-- Remove agent_session from tasks
-- (Complex with PocketBase schema)

-- Update column enum to remove need_input
-- (Complex with PocketBase schema)
EOF

# 4. Revert code changes
git checkout internal/commands/root.go
git checkout migrations/
```

**Note**: PocketBase schema changes are stored in the database, so rollback requires either restoring a backup or carefully modifying the `_collections` table.

---

## Files Changed

| File | Change Type | Description |
|------|-------------|-------------|
| `migrations/12_need_input_column.go` | New | Add need_input to column enum |
| `migrations/13_agent_session_field.go` | New | Add agent_session JSON field |
| `migrations/14_comments_collection.go` | New | Create comments collection |
| `migrations/15_sessions_collection.go` | New | Create sessions collection |
| `migrations/16_board_resume_mode.go` | New | Add resume_mode to boards |
| `internal/commands/root.go` | Modified | Update ValidColumns |

---

## Next Phase

Once all tests pass, proceed to [Phase 1: Core CLI Commands](./ai-workflow-phase-1.md).

Phase 1 will implement:
- `egenskriven block` command
- `egenskriven comment` command
- `egenskriven comments` command
- `egenskriven list --need-input` flag

---

## Notes for Implementer

1. **Migration order matters**: Run migrations in numerical order (12, 13, 14, 15, 16)

2. **Check existing migrations**: Look at existing migration files to match the project's style

3. **PocketBase version**: Ensure you're using the same PocketBase version as the project

4. **Test database**: Consider using a separate test database during development:
   ```bash
   EGENSKRIVEN_DATA_DIR=/tmp/egenskriven-test egenskriven serve
   ```

5. **Collection IDs**: When creating relations, you need the actual collection ID, not just the name. The code above uses `app.FindCollectionByNameOrId()` to get this dynamically.
