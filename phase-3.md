# Phase 3: Full CLI

**Goal**: Complete CLI with all features including batch operations, epics, advanced filtering, and improved error handling.

**Duration Estimate**: 3-4 days

**Prerequisites**: Phase 1 (Core CLI) complete. Phase 1.5 and Phase 2 can run in parallel with this phase.

**Deliverable**: A fully-featured CLI with batch operations, epic management, advanced filtering, and a version command.

---

## Overview

Phase 3 extends the Core CLI from Phase 1 with professional-grade features:

- **Batch Operations**: Create and delete multiple tasks efficiently via stdin or file input
- **Epics**: Group related tasks under larger initiatives with color-coded organization
- **Advanced Filtering**: Labels, limits, and sorting for precise task queries
- **Error Handling**: Better messages with suggestions and context
- **Version Command**: Display build and version information

These features make the CLI suitable for both human users and AI agents that need to manage tasks programmatically.

### Why These Features?

| Feature | Purpose |
|---------|---------|
| Batch input | AI agents can create many tasks in one command |
| Epics | Organize work into larger initiatives (e.g., "Q1 Launch") |
| Advanced filters | Reduce context needed when querying tasks |
| Better errors | Help users (and agents) self-correct mistakes |
| Version command | Standard CLI practice for debugging |

---

## Already Implemented

Based on the current codebase, the following features from Phase 3 are **already partially implemented**:

### List Command Filters (Partially Done)
**File**: `internal/commands/list.go`

| Flag | Status | Notes |
|------|--------|-------|
| `--column` | Done | Repeatable, multi-value |
| `--type` | Done | Repeatable, multi-value |
| `--priority` | Done | Repeatable, multi-value |
| `--search` | Done | Case-insensitive title search |
| `--created-by` | Done | Filter by creator type |
| `--agent` | Done | Filter by agent name |
| `--ready` | Done | Unblocked tasks in todo/backlog |
| `--is-blocked` | Done | Blocked tasks only |
| `--not-blocked` | Done | Unblocked tasks only |
| `--fields` | Done | JSON field selection |

### Delete Command (Partially Done)
**File**: `internal/commands/delete.go`

- Multiple task IDs as arguments - **Done**
- `--force` flag to skip confirmation - **Done**
- Confirmation prompt with task listing - **Done**

### Error Infrastructure (Partially Done)
**Files**: `internal/output/output.go`, `internal/commands/root.go`, `internal/resolver/resolver.go`

- Exit codes defined - **Done**
- `AmbiguousError` struct - **Done**
- JSON error format - **Done**

---

## Tasks (Remaining Work)

### 3.1 Add Epics Collection Migration

**What**: Create the `epics` database collection.

**Why**: Epics group related tasks into larger initiatives. They have their own identity (title, description, color) separate from tasks.

**File**: `migrations/2_epics.go`

```go
package migrations

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterEpicsMigration registers the epics collection migration.
// Call this from your main.go after registering the tasks migration.
func RegisterEpicsMigration(app *pocketbase.PocketBase) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		// Check if epics collection already exists
		_, err := app.FindCollectionByNameOrId("epics")
		if err == nil {
			// Collection exists, skip creation
			return e.Next()
		}

		// Create epics collection
		collection := core.NewBaseCollection("epics")

		// id is automatically created by PocketBase

		// title - Epic name (required)
		collection.Fields.Add(&core.TextField{
			Name:     "title",
			Required: true,
			Min:      1,
			Max:      200,
		})

		// description - Longer description of the epic
		collection.Fields.Add(&core.TextField{
			Name: "description",
			Max:  5000,
		})

		// color - Hex color for visual grouping (e.g., "#3B82F6")
		collection.Fields.Add(&core.TextField{
			Name:    "color",
			Pattern: `^#[0-9A-Fa-f]{6}$`,
			Max:     7,
		})

		if err := app.Save(collection); err != nil {
			return err
		}

		return e.Next()
	})
}
```

**Steps**:

1. Create the file:
   ```bash
   touch migrations/2_epics.go
   ```

2. Add the code above.

3. Register in `main.go` (or wherever migrations are registered):
   ```go
   migrations.RegisterEpicsMigration(app)
   ```

4. Verify migration runs:
   ```bash
   make run
   ```
   Check admin UI at `http://localhost:8090/_/` - epics collection should exist.

---

### 3.2 Add Epic Relation to Tasks

**What**: Add an `epic` relation field to the tasks collection.

**Why**: Tasks need to link to epics so we can group and filter them.

**File**: `migrations/3_epic_relation.go`

```go
package migrations

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterEpicRelationMigration adds the epic relation field to tasks.
func RegisterEpicRelationMigration(app *pocketbase.PocketBase) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		tasks, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return e.Next() // Tasks collection doesn't exist yet
		}

		// Check if field already exists
		if tasks.Fields.GetByName("epic") != nil {
			return e.Next()
		}

		epics, err := app.FindCollectionByNameOrId("epics")
		if err != nil {
			return e.Next() // Epics collection doesn't exist yet
		}

		// Add epic relation field
		tasks.Fields.Add(&core.RelationField{
			Name:          "epic",
			CollectionId:  epics.Id,
			MaxSelect:     1,
			CascadeDelete: false, // Tasks remain when epic is deleted
		})

		if err := app.Save(tasks); err != nil {
			return err
		}

		return e.Next()
	})
}
```

**Steps**:

1. Create the file:
   ```bash
   touch migrations/3_epic_relation.go
   ```

2. Add the code above.

3. Register in main.go **after** the epics migration.

4. Verify: In admin UI, check tasks collection has "epic" field.

---

### 3.3 Implement Epic Commands

**What**: Create the epic command with subcommands for list, add, show, delete.

**Why**: Epic operations (list, add, show, delete) are grouped under `egenskriven epic`.

**File**: `internal/commands/epic.go`

```go
package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
)

func newEpicCmd(app *pocketbase.PocketBase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "epic",
		Short: "Manage epics",
		Long: `Manage epics for grouping related tasks.

Epics are larger initiatives that contain multiple tasks.
Examples: "Q1 Launch", "Auth Refactor", "Performance Sprint"`,
	}

	// Add subcommands
	cmd.AddCommand(newEpicListCmd(app))
	cmd.AddCommand(newEpicAddCmd(app))
	cmd.AddCommand(newEpicShowCmd(app))
	cmd.AddCommand(newEpicDeleteCmd(app))

	return cmd
}

// ========== Epic List ==========

func newEpicListCmd(app *pocketbase.PocketBase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all epics",
		Long:  `List all epics with their task counts.`,
		Example: `  egenskriven epic list
  egenskriven epic list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			// Find all epics
			records, err := app.FindAllRecords("epics")
			if err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to list epics: %v", err), nil)
			}

			// Output
			if out.JSON {
				epics := make([]map[string]any, 0, len(records))
				for _, record := range records {
					taskCount := getEpicTaskCount(app, record.Id)
					epics = append(epics, map[string]any{
						"id":          record.Id,
						"title":       record.GetString("title"),
						"description": record.GetString("description"),
						"color":       record.GetString("color"),
						"task_count":  taskCount,
						"created":     record.GetDateTime("created").String(),
						"updated":     record.GetDateTime("updated").String(),
					})
				}
				return json.NewEncoder(os.Stdout).Encode(map[string]any{
					"epics": epics,
					"count": len(epics),
				})
			}

			// Human-readable output
			if len(records) == 0 {
				fmt.Println("No epics found. Create one with: egenskriven epic add \"Epic title\"")
				return nil
			}

			fmt.Println("EPICS")
			fmt.Println(strings.Repeat("-", 40))
			for _, record := range records {
				taskCount := getEpicTaskCount(app, record.Id)
				colorIndicator := ""
				if color := record.GetString("color"); color != "" {
					colorIndicator = fmt.Sprintf(" %s", color)
				}
				fmt.Printf("  [%s] %s%s (%d tasks)\n",
					shortID(record.Id), record.GetString("title"), colorIndicator, taskCount)
			}
			fmt.Printf("\nTotal: %d epics\n", len(records))

			return nil
		},
	}

	return cmd
}

// ========== Epic Add ==========

func newEpicAddCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		color       string
		description string
	)

	cmd := &cobra.Command{
		Use:   "add <title>",
		Short: "Create a new epic",
		Long: `Create a new epic to group related tasks.

Epics can have a color for visual identification in the UI.
Color must be a valid hex code (e.g., #3B82F6).`,
		Example: `  egenskriven epic add "Q1 Launch"
  egenskriven epic add "Auth Refactor" --color "#22C55E"
  egenskriven epic add "Tech Debt" --description "Clean up legacy code"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			title := args[0]

			// Validate color format if provided
			if color != "" && !isValidHexColor(color) {
				return out.Error(ExitValidation,
					fmt.Sprintf("invalid color format '%s' (must be hex like #3B82F6)", color), nil)
			}

			// Find epics collection
			collection, err := app.FindCollectionByNameOrId("epics")
			if err != nil {
				return out.Error(ExitGeneralError, "epics collection not found - run migrations first", nil)
			}

			// Create record
			record := core.NewRecord(collection)
			record.Set("title", title)
			if description != "" {
				record.Set("description", description)
			}
			if color != "" {
				record.Set("color", color)
			}

			if err := app.Save(record); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to create epic: %v", err), nil)
			}

			// Output
			if out.JSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]any{
					"id":          record.Id,
					"title":       record.GetString("title"),
					"description": record.GetString("description"),
					"color":       record.GetString("color"),
					"task_count":  0,
					"created":     record.GetDateTime("created").String(),
				})
			}

			colorDisplay := ""
			if color != "" {
				colorDisplay = fmt.Sprintf(" %s", color)
			}
			fmt.Printf("Created epic: %s%s [%s]\n", title, colorDisplay, shortID(record.Id))

			return nil
		},
	}

	cmd.Flags().StringVarP(&color, "color", "c", "", "Epic color (hex, e.g., #3B82F6)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Epic description")

	return cmd
}

// ========== Epic Show ==========

func newEpicShowCmd(app *pocketbase.PocketBase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <epic>",
		Short: "Show epic details",
		Long: `Show detailed information about an epic including linked tasks.

You can reference an epic by:
- Full ID: abc123def456
- Partial ID: abc123
- Title (case-insensitive): "q1 launch"`,
		Example: `  egenskriven epic show abc123
  egenskriven epic show "Q1 Launch"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			ref := args[0]

			// Resolve epic reference
			record, err := resolveEpic(app, ref)
			if err != nil {
				return out.Error(ExitNotFound, err.Error(), nil)
			}

			// Get linked tasks
			linkedTasks, err := app.FindAllRecords("tasks",
				dbx.NewExp("epic = {:epicId}", dbx.Params{"epicId": record.Id}),
			)
			if err != nil {
				linkedTasks = []*core.Record{}
			}

			// Output
			if out.JSON {
				tasks := make([]map[string]any, 0, len(linkedTasks))
				for _, task := range linkedTasks {
					tasks = append(tasks, map[string]any{
						"id":       task.Id,
						"title":    task.GetString("title"),
						"column":   task.GetString("column"),
						"priority": task.GetString("priority"),
					})
				}

				return json.NewEncoder(os.Stdout).Encode(map[string]any{
					"epic": map[string]any{
						"id":          record.Id,
						"title":       record.GetString("title"),
						"description": record.GetString("description"),
						"color":       record.GetString("color"),
						"task_count":  len(linkedTasks),
						"created":     record.GetDateTime("created").String(),
						"updated":     record.GetDateTime("updated").String(),
					},
					"tasks": tasks,
				})
			}

			// Human-readable output
			fmt.Printf("Epic: %s\n", record.Id)
			fmt.Printf("Title:       %s\n", record.GetString("title"))
			if color := record.GetString("color"); color != "" {
				fmt.Printf("Color:       %s\n", color)
			}
			if desc := record.GetString("description"); desc != "" {
				fmt.Printf("Description: %s\n", desc)
			}
			fmt.Printf("Created:     %s\n", record.GetDateTime("created").String())
			fmt.Printf("Updated:     %s\n", record.GetDateTime("updated").String())
			fmt.Printf("\nLinked Tasks (%d):\n", len(linkedTasks))

			if len(linkedTasks) == 0 {
				fmt.Println("  (no tasks)")
			} else {
				for _, task := range linkedTasks {
					fmt.Printf("  [%s] %s (%s, %s)\n",
						shortID(task.Id),
						task.GetString("title"),
						task.GetString("column"),
						task.GetString("priority"),
					)
				}
			}

			return nil
		},
	}

	return cmd
}

// ========== Epic Delete ==========

func newEpicDeleteCmd(app *pocketbase.PocketBase) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <epic>",
		Short: "Delete an epic",
		Long: `Delete an epic.

Tasks linked to the epic will remain but will no longer be associated with the epic.
Use --force to skip the confirmation prompt.`,
		Example: `  egenskriven epic delete abc123
  egenskriven epic delete "Q1 Launch" --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			ref := args[0]

			// Resolve epic
			record, err := resolveEpic(app, ref)
			if err != nil {
				return out.Error(ExitNotFound, err.Error(), nil)
			}

			// Count linked tasks
			taskCount := getEpicTaskCount(app, record.Id)

			// Confirm deletion
			if !force && !out.Quiet && !out.JSON {
				fmt.Printf("Delete epic '%s'?", record.GetString("title"))
				if taskCount > 0 {
					fmt.Printf(" (%d tasks will be unlinked)", taskCount)
				}
				fmt.Print("\nType 'yes' to confirm: ")

				var response string
				fmt.Scanln(&response)
				if response != "yes" {
					fmt.Println("Cancelled")
					return nil
				}
			}

			// Delete the epic
			if err := app.Delete(record); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to delete epic: %v", err), nil)
			}

			// Output
			if out.JSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]any{
					"deleted":        record.Id,
					"title":          record.GetString("title"),
					"tasks_unlinked": taskCount,
				})
			}

			fmt.Printf("Deleted epic: %s [%s]\n", record.GetString("title"), shortID(record.Id))
			if taskCount > 0 {
				fmt.Printf("  %d tasks unlinked\n", taskCount)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

// ========== Helper Functions ==========

// resolveEpic finds an epic by ID, ID prefix, or title
func resolveEpic(app *pocketbase.PocketBase, ref string) (*core.Record, error) {
	// Try exact ID match
	record, err := app.FindRecordById("epics", ref)
	if err == nil {
		return record, nil
	}

	// Try ID prefix match
	records, err := app.FindAllRecords("epics",
		dbx.NewExp("id LIKE {:prefix}", dbx.Params{"prefix": ref + "%"}),
	)
	if err == nil && len(records) == 1 {
		return records[0], nil
	}

	// Try title match (case-insensitive)
	records, err = app.FindAllRecords("epics",
		dbx.NewExp("LOWER(title) LIKE {:title}",
			dbx.Params{"title": "%" + strings.ToLower(ref) + "%"}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search epics: %w", err)
	}

	switch len(records) {
	case 0:
		return nil, fmt.Errorf("no epic found matching: %s", ref)
	case 1:
		return records[0], nil
	default:
		// Multiple matches - ambiguous
		var matches []string
		for _, r := range records {
			matches = append(matches, fmt.Sprintf("[%s] %s", shortID(r.Id), r.GetString("title")))
		}
		return nil, fmt.Errorf("ambiguous epic reference '%s' matches multiple epics:\n  %s",
			ref, strings.Join(matches, "\n  "))
	}
}

// getEpicTaskCount returns the number of tasks linked to an epic
func getEpicTaskCount(app *pocketbase.PocketBase, epicID string) int {
	tasks, err := app.FindAllRecords("tasks",
		dbx.NewExp("epic = {:epicId}", dbx.Params{"epicId": epicID}),
	)
	if err != nil {
		return 0
	}
	return len(tasks)
}

// isValidHexColor validates a hex color string (#RRGGBB format)
func isValidHexColor(color string) bool {
	if len(color) != 7 || color[0] != '#' {
		return false
	}
	for i := 1; i < 7; i++ {
		c := color[i]
		valid := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
		if !valid {
			return false
		}
	}
	return true
}
```

**Steps**:

1. Create the file:
   ```bash
   touch internal/commands/epic.go
   ```

2. Add the code above.

3. Register in `root.go` by adding to the command registration:
   ```go
   rootCmd.AddCommand(newEpicCmd(app))
   ```

4. Verify:
   ```bash
   make build
   ./egenskriven epic --help
   ```

---

### 3.4 Add Epic Flag to Add Command

**What**: Add `--epic` flag to the task add command.

**Why**: Users need to link tasks to epics when creating them.

**Update**: `internal/commands/add.go`

Add the following changes:

```go
// Add to flag declarations (around line 20)
var (
	// ... existing flags ...
	epic string
)

// Add flag definition (around line 144)
cmd.Flags().StringVarP(&epic, "epic", "e", "", "Link task to epic (ID or title)")

// Add in RunE, after validations and before creating the record (around line 85):
// Handle epic linkage
var epicID string
if epic != "" {
	epicRecord, err := resolveEpic(app, epic)
	if err != nil {
		return out.Error(ExitValidation, fmt.Sprintf("invalid epic: %v", err), nil)
	}
	epicID = epicRecord.Id
}

// Then when setting record fields (around line 107):
if epicID != "" {
	record.Set("epic", epicID)
}
```

**Usage**:

```bash
# Link task to epic by ID
egenskriven add "Implement login" --epic abc123

# Link task to epic by title
egenskriven add "Add logout button" --epic "Auth Refactor"
```

---

### 3.5 Add Epic Filter to List Command

**What**: Add `--epic` filter to the list command.

**Why**: Users need to see all tasks in a specific epic.

**Update**: `internal/commands/list.go`

Add the following changes:

```go
// Add to flag declarations (around line 23)
var epicFilter string

// Add flag definition (around line 186)
cmd.Flags().StringVarP(&epicFilter, "epic", "e", "", "Filter by epic (ID or title)")

// Add filter logic in RunE (around line 120, after other filters):
// Epic filter
if epicFilter != "" {
	epicRecord, err := resolveEpic(app, epicFilter)
	if err != nil {
		return out.Error(ExitValidation, fmt.Sprintf("invalid epic filter: %v", err), nil)
	}
	filters = append(filters, dbx.NewExp(
		"epic = {:epic}",
		dbx.Params{"epic": epicRecord.Id},
	))
}
```

**Usage**:

```bash
# List tasks in an epic
egenskriven list --epic "Q1 Launch"
egenskriven list --epic abc123 --json
```

---

### 3.6 Add Batch Input to Add Command

**What**: Allow creating multiple tasks from stdin or a file.

**Why**: AI agents and scripts need to create many tasks efficiently in one operation.

**Update**: `internal/commands/add.go`

This requires significant changes. Here's the updated file:

```go
package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
)

// TaskInput represents a task for batch creation
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

func newAddCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		taskType  string
		priority  string
		column    string
		labels    []string
		customID  string
		createdBy string
		agentName string
		epic      string
		stdin     bool
		file      string
	)

	cmd := &cobra.Command{
		Use:   "add [title]",
		Short: "Add a new task",
		Long: `Add a new task to the kanban board.

Supports batch creation via --stdin or --file for agent workflows.
Batch input accepts JSON lines (one JSON object per line) or a JSON array.

Examples:
  egenskriven add "Implement dark mode"
  egenskriven add "Fix bug" --type bug --priority urgent
  egenskriven add "Setup CI" --id ci-setup-001
  egenskriven add "Refactor auth" --agent claude
  egenskriven add "Add login" --epic "Auth Refactor"
  
  # Batch from stdin (JSON lines)
  echo '{"title":"Task 1"}
{"title":"Task 2","priority":"high"}' | egenskriven add --stdin
  
  # Batch from file
  egenskriven add --file tasks.json`,
		Args: cobra.MaximumNArgs(1), // Changed from ExactArgs(1)
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			// Handle batch input
			if stdin || file != "" {
				return addBatch(app, out, stdin, file, agentName)
			}

			// Single task creation requires title argument
			if len(args) == 0 {
				return out.Error(ExitInvalidArguments,
					"title is required\n\nUsage: egenskriven add <title>\n       egenskriven add --stdin < tasks.json", nil)
			}

			title := args[0]

			// Validate inputs
			if !isValidType(taskType) {
				return out.Error(ExitValidation,
					fmt.Sprintf("invalid type '%s', must be one of: %v", taskType, ValidTypes), nil)
			}
			if !isValidPriority(priority) {
				return out.Error(ExitValidation,
					fmt.Sprintf("invalid priority '%s', must be one of: %v", priority, ValidPriorities), nil)
			}
			if !isValidColumn(column) {
				return out.Error(ExitValidation,
					fmt.Sprintf("invalid column '%s', must be one of: %v", column, ValidColumns), nil)
			}

			// Determine creator
			if createdBy == "" {
				if agentName != "" {
					createdBy = "agent"
				} else {
					if fileInfo, _ := os.Stdin.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
						createdBy = "user"
					} else {
						createdBy = "cli"
					}
				}
			}

			// Resolve epic if provided
			var epicID string
			if epic != "" {
				epicRecord, err := resolveEpic(app, epic)
				if err != nil {
					return out.Error(ExitValidation, fmt.Sprintf("invalid epic: %v", err), nil)
				}
				epicID = epicRecord.Id
			}

			// Find the tasks collection
			collection, err := app.FindCollectionByNameOrId("tasks")
			if err != nil {
				return out.Error(ExitGeneralError, "tasks collection not found - run migrations first", nil)
			}

			// Create the record
			record := core.NewRecord(collection)

			// Set custom ID if provided (idempotency)
			if customID != "" {
				existing, err := app.FindRecordById("tasks", customID)
				if err == nil {
					out.Task(existing, "Existing")
					return nil
				}
				record.Id = customID
			}

			// Set task fields
			record.Set("title", title)
			record.Set("type", taskType)
			record.Set("priority", priority)
			record.Set("column", column)
			record.Set("position", GetNextPosition(app, column))
			record.Set("labels", labels)
			record.Set("blocked_by", []string{})
			record.Set("created_by", createdBy)
			if agentName != "" {
				record.Set("created_by_agent", agentName)
			}
			if epicID != "" {
				record.Set("epic", epicID)
			}

			// Initialize history
			history := []map[string]any{
				{
					"timestamp":    time.Now().UTC().Format(time.RFC3339),
					"action":       "created",
					"actor":        createdBy,
					"actor_detail": agentName,
					"changes":      nil,
				},
			}
			record.Set("history", history)

			// Save the record
			if err := app.Save(record); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to create task: %v", err), nil)
			}

			out.Task(record, "Created")
			return nil
		},
	}

	// Define flags
	cmd.Flags().StringVarP(&taskType, "type", "t", "feature", "Task type (bug, feature, chore)")
	cmd.Flags().StringVarP(&priority, "priority", "p", "medium", "Priority (low, medium, high, urgent)")
	cmd.Flags().StringVarP(&column, "column", "c", "backlog", "Initial column")
	cmd.Flags().StringSliceVarP(&labels, "label", "l", nil, "Labels (repeatable)")
	cmd.Flags().StringVar(&customID, "id", "", "Custom ID for idempotency")
	cmd.Flags().StringVar(&createdBy, "created-by", "", "Creator type (user, agent, cli)")
	cmd.Flags().StringVar(&agentName, "agent", "", "Agent identifier (implies --created-by agent)")
	cmd.Flags().StringVarP(&epic, "epic", "e", "", "Link task to epic (ID or title)")
	cmd.Flags().BoolVar(&stdin, "stdin", false, "Read tasks from stdin (JSON lines or array)")
	cmd.Flags().StringVarP(&file, "file", "f", "", "Read tasks from JSON file")

	return cmd
}

// addBatch handles batch task creation from stdin or file
func addBatch(app *pocketbase.PocketBase, out *Formatter, useStdin bool, filePath string, agent string) error {
	var reader io.Reader

	if useStdin {
		reader = os.Stdin
	} else {
		f, err := os.Open(filePath)
		if err != nil {
			return out.Error(ExitGeneralError, fmt.Sprintf("failed to open file: %v", err), nil)
		}
		defer f.Close()
		reader = f
	}

	// Read all content to detect format
	content, err := io.ReadAll(reader)
	if err != nil {
		return out.Error(ExitGeneralError, fmt.Sprintf("failed to read input: %v", err), nil)
	}

	trimmed := strings.TrimSpace(string(content))
	if trimmed == "" {
		return out.Error(ExitInvalidArguments, "empty input", nil)
	}

	var inputs []TaskInput

	// Detect format: JSON array or JSON lines
	if strings.HasPrefix(trimmed, "[") {
		// JSON array format
		if err := json.Unmarshal([]byte(trimmed), &inputs); err != nil {
			return out.Error(ExitInvalidArguments, fmt.Sprintf("invalid JSON array: %v", err), nil)
		}
	} else {
		// JSON lines format
		scanner := bufio.NewScanner(strings.NewReader(trimmed))
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			var input TaskInput
			if err := json.Unmarshal([]byte(line), &input); err != nil {
				return out.Error(ExitInvalidArguments,
					fmt.Sprintf("line %d: invalid JSON: %v", lineNum, err), nil)
			}
			inputs = append(inputs, input)
		}
		if err := scanner.Err(); err != nil {
			return out.Error(ExitGeneralError, fmt.Sprintf("failed to read input: %v", err), nil)
		}
	}

	if len(inputs) == 0 {
		return out.Error(ExitInvalidArguments, "no tasks found in input", nil)
	}

	// Find the tasks collection
	collection, err := app.FindCollectionByNameOrId("tasks")
	if err != nil {
		return out.Error(ExitGeneralError, "tasks collection not found - run migrations first", nil)
	}

	// Create all tasks
	var created []*core.Record
	var errors []string

	for i, input := range inputs {
		if input.Title == "" {
			errors = append(errors, fmt.Sprintf("task %d: title is required", i+1))
			continue
		}

		record := core.NewRecord(collection)

		if input.ID != "" {
			// Check idempotency
			existing, err := app.FindRecordById("tasks", input.ID)
			if err == nil {
				created = append(created, existing)
				continue
			}
			record.Id = input.ID
		}

		// Set fields with defaults
		record.Set("title", input.Title)
		record.Set("type", defaultString(input.Type, "feature"))
		record.Set("priority", defaultString(input.Priority, "medium"))
		record.Set("column", defaultString(input.Column, "backlog"))
		record.Set("position", GetNextPosition(app, defaultString(input.Column, "backlog")))
		record.Set("labels", input.Labels)
		record.Set("blocked_by", []string{})

		if input.Description != "" {
			record.Set("description", input.Description)
		}

		// Handle epic
		if input.Epic != "" {
			epicRecord, err := resolveEpic(app, input.Epic)
			if err != nil {
				errors = append(errors, fmt.Sprintf("task %d (%s): invalid epic '%s': %v",
					i+1, input.Title, input.Epic, err))
				continue
			}
			record.Set("epic", epicRecord.Id)
		}

		// Set creator info
		createdBy := "cli"
		if agent != "" {
			createdBy = "agent"
			record.Set("created_by_agent", agent)
		}
		record.Set("created_by", createdBy)

		// Initialize history
		history := []map[string]any{
			{
				"timestamp":    time.Now().UTC().Format(time.RFC3339),
				"action":       "created",
				"actor":        createdBy,
				"actor_detail": agent,
				"changes":      nil,
			},
		}
		record.Set("history", history)

		if err := app.Save(record); err != nil {
			errors = append(errors, fmt.Sprintf("task %d (%s): failed to save: %v", i+1, input.Title, err))
			continue
		}
		created = append(created, record)
	}

	// Output results
	if out.JSON {
		tasks := make([]map[string]any, 0, len(created))
		for _, record := range created {
			tasks = append(tasks, map[string]any{
				"id":       record.Id,
				"title":    record.GetString("title"),
				"type":     record.GetString("type"),
				"priority": record.GetString("priority"),
				"column":   record.GetString("column"),
			})
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"created": len(created),
			"failed":  len(errors),
			"tasks":   tasks,
			"errors":  errors,
		})
	}

	// Human output
	for _, record := range created {
		fmt.Printf("Created: %s [%s]\n", record.GetString("title"), shortID(record.Id))
	}

	if len(errors) > 0 {
		fmt.Println("\nErrors:")
		for _, e := range errors {
			fmt.Printf("  %s\n", e)
		}
	}

	fmt.Printf("\nCreated %d tasks", len(created))
	if len(errors) > 0 {
		fmt.Printf(", %d failed", len(errors))
	}
	fmt.Println()

	return nil
}

// defaultString returns the value if non-empty, otherwise the default
func defaultString(value, defaultVal string) string {
	if value == "" {
		return defaultVal
	}
	return value
}
```

**Batch Input Formats**:

**JSON Lines** (one object per line):
```json
{"title": "Task 1", "type": "bug", "priority": "high"}
{"title": "Task 2", "column": "todo"}
{"title": "Task 3", "epic": "Q1 Launch"}
```

**JSON Array**:
```json
[
  {"title": "Task 1", "type": "bug"},
  {"title": "Task 2", "priority": "high"}
]
```

**Usage**:

```bash
# From stdin (JSON lines)
echo '{"title":"Task 1"}
{"title":"Task 2","priority":"high"}' | egenskriven add --stdin

# From file
egenskriven add --file tasks.json

# JSON output for verification
egenskriven add --file tasks.json --json
```

---

### 3.7 Add Stdin Support to Delete Command

**What**: Add `--stdin` flag to read task references from stdin.

**Why**: Batch cleanup operations need to delete many tasks efficiently.

**Update**: `internal/commands/delete.go`

Add the following changes:

```go
// Add to flag declarations (around line 17)
var (
	force bool
	stdin bool  // ADD THIS
)

// Add flag definition (around line 101)
cmd.Flags().BoolVar(&stdin, "stdin", false, "Read task references from stdin (one per line)")

// Update Args to allow zero arguments when using stdin (around line 33)
Args: cobra.MinimumNArgs(0), // Changed from MinimumNArgs(1)

// Update RunE to handle stdin (replace lines 42-58):
// Collect task references
var refs []string

if stdin {
	// Read from stdin
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		ref := strings.TrimSpace(scanner.Text())
		if ref != "" {
			refs = append(refs, ref)
		}
	}
	if err := scanner.Err(); err != nil {
		return out.Error(ExitGeneralError, fmt.Sprintf("failed to read stdin: %v", err), nil)
	}
} else {
	if len(args) == 0 {
		return out.Error(ExitInvalidArguments,
			"at least one task reference is required\n\nUsage: egenskriven delete <task> [task...]\n       echo 'task-id' | egenskriven delete --stdin", nil)
	}
	refs = args
}

if len(refs) == 0 {
	return out.Error(ExitInvalidArguments, "no tasks to delete", nil)
}

// Rest of the function continues with 'refs' instead of 'args'
```

**Usage**:

```bash
# From stdin (one ID per line)
echo -e "abc123\ndef456" | egenskriven delete --stdin --force

# Multiple arguments (already works)
egenskriven delete abc123 def456 --force
```

---

### 3.8 Add Missing List Filters

**What**: Add `--label`, `--limit`, and `--sort` flags to the list command.

**Why**: These filters enable precise queries that reduce token usage for AI agents.

**Update**: `internal/commands/list.go`

Add the following changes:

```go
// Add to flag declarations (around line 24)
var (
	// ... existing flags ...
	labels []string  // ADD THIS
	limit  int       // ADD THIS
	sort   string    // ADD THIS
)

// Add flag definitions (around line 186)
cmd.Flags().StringSliceVarP(&labels, "label", "l", nil, "Filter by label (repeatable)")
cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of results (0 = no limit)")
cmd.Flags().StringVar(&sort, "sort", "", "Sort order (e.g., '-priority,position')")

// Add label filter in RunE (around line 120):
// Label filter
if len(labels) > 0 {
	for _, label := range labels {
		filters = append(filters, dbx.NewExp(
			"labels LIKE {:label}",
			dbx.Params{"label": "%" + label + "%"},
		))
	}
}

// Replace the query execution section (around line 138-147):
// Execute query with optional limit
var tasks []*core.Record
var err error

query := app.RecordQuery("tasks")

if len(filters) > 0 {
	combined := dbx.And(filters...)
	query = query.AndWhere(combined)
}

// Apply custom sort if specified
if sort != "" {
	// Parse sort string (e.g., "-priority,position")
	sortFields := strings.Split(sort, ",")
	for _, field := range sortFields {
		field = strings.TrimSpace(field)
		if strings.HasPrefix(field, "-") {
			query = query.OrderBy(field[1:] + " DESC")
		} else {
			query = query.OrderBy(field + " ASC")
		}
	}
} else {
	query = query.OrderBy("column ASC", "position ASC")
}

// Apply limit
if limit > 0 {
	query = query.Limit(int64(limit))
}

err = query.All(&tasks)
if err != nil {
	return out.Error(ExitGeneralError, fmt.Sprintf("failed to list tasks: %v", err), nil)
}
```

**Note**: The current codebase uses `FindAllRecords` which doesn't support limit/sort directly. You may need to switch to `RecordQuery` for more control, or fetch all and filter in Go.

**Simpler Alternative** (if keeping FindAllRecords):

```go
// After getting all tasks, apply limit in Go:
if limit > 0 && len(tasks) > limit {
	tasks = tasks[:limit]
}
```

**Usage**:

```bash
# Filter by label
egenskriven list --label frontend --label ui

# Limit results
egenskriven list --limit 10

# Custom sort
egenskriven list --sort "-priority,position"

# Combined
egenskriven list --label critical --limit 5 --sort "-priority"
```

---

### 3.9 Implement Version Command

**What**: Display version and build information.

**Why**: Standard practice for CLI tools. Helps with debugging and support.

**File**: `internal/commands/version.go`

```go
package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

// These variables are set at build time via ldflags:
// go build -ldflags "-X github.com/ramtinJ95/EgenSkriven/internal/commands.Version=1.0.0"
var (
	Version   = "dev"
	BuildDate = "unknown"
	GitCommit = "unknown"
)

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Long:  `Display version, build date, and runtime information.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			info := map[string]string{
				"version":    Version,
				"build_date": BuildDate,
				"git_commit": GitCommit,
				"go_version": runtime.Version(),
				"os":         runtime.GOOS,
				"arch":       runtime.GOARCH,
			}

			if out.JSON {
				return json.NewEncoder(os.Stdout).Encode(info)
			}

			fmt.Printf("EgenSkriven %s\n", Version)
			fmt.Printf("Build date: %s\n", BuildDate)
			fmt.Printf("Git commit: %s\n", GitCommit)
			fmt.Printf("Go version: %s\n", runtime.Version())
			fmt.Printf("OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)

			return nil
		},
	}

	return cmd
}
```

**Register in root.go**:

```go
rootCmd.AddCommand(newVersionCmd())
```

**Update Makefile** for version embedding:

```makefile
VERSION ?= dev
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LDFLAGS := -X github.com/ramtinJ95/EgenSkriven/internal/commands.Version=$(VERSION)
LDFLAGS += -X github.com/ramtinJ95/EgenSkriven/internal/commands.BuildDate=$(BUILD_DATE)
LDFLAGS += -X github.com/ramtinJ95/EgenSkriven/internal/commands.GitCommit=$(GIT_COMMIT)

build:
	@echo "Building $(VERSION)..."
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o egenskriven ./cmd/egenskriven
```

**Usage**:

```bash
./egenskriven version
# Output:
# EgenSkriven dev
# Build date: 2024-01-15T10:00:00Z
# Git commit: abc1234
# Go version: go1.21.0
# OS/Arch:    linux/amd64

./egenskriven version --json
```

---

### 3.10 Improve Error Messages with Suggestions

**What**: Enhance error output to include helpful suggestions.

**Why**: Good error messages help users and AI agents self-correct mistakes.

**Update**: `internal/output/output.go`

The current implementation has basic error handling. Enhance it by adding a suggestion field:

```go
// Add to the Error method or create a new ErrorWithSuggestion method:

// ErrorWithSuggestion outputs an error with a helpful suggestion
func (f *Formatter) ErrorWithSuggestion(code int, message, suggestion string, data any) error {
	if f.JSON {
		errOutput := map[string]any{
			"error": map[string]any{
				"code":    code,
				"message": message,
			},
		}
		if suggestion != "" {
			errOutput["error"].(map[string]any)["suggestion"] = suggestion
		}
		if data != nil {
			errOutput["error"].(map[string]any)["data"] = data
		}
		json.NewEncoder(os.Stderr).Encode(errOutput)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", message)
		if suggestion != "" {
			fmt.Fprintf(os.Stderr, "\nSuggestion: %s\n", suggestion)
		}
	}
	os.Exit(code)
	return nil
}
```

**Use throughout commands**:

```go
// In resolver or commands:
return out.ErrorWithSuggestion(ExitNotFound,
	fmt.Sprintf("no task found matching: %s", ref),
	"Use 'egenskriven list' to see available tasks",
	nil)

return out.ErrorWithSuggestion(ExitAmbiguous,
	fmt.Sprintf("ambiguous reference '%s' matches multiple tasks", ref),
	"Use a more specific ID or the full task ID",
	matches)

return out.ErrorWithSuggestion(ExitValidation,
	fmt.Sprintf("invalid priority '%s'", priority),
	fmt.Sprintf("Valid priorities: %v", ValidPriorities),
	nil)
```

---

### 3.11 Write Tests for Epic Commands

**File**: `internal/commands/epic_test.go`

```go
package commands

import (
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ramtinJ95/EgenSkriven/internal/testutil"
)

func TestIsValidHexColor(t *testing.T) {
	tests := []struct {
		color string
		valid bool
	}{
		{"#3B82F6", true},
		{"#aabbcc", true},
		{"#AABBCC", true},
		{"red", false},
		{"#FFF", false},
		{"3B82F6", false},
		{"#GGGGGG", false},
		{"", false},
		{"#12345", false},
		{"#1234567", false},
	}

	for _, tt := range tests {
		t.Run(tt.color, func(t *testing.T) {
			result := isValidHexColor(tt.color)
			assert.Equal(t, tt.valid, result, "isValidHexColor(%q)", tt.color)
		})
	}
}

func TestResolveEpic(t *testing.T) {
	app := testutil.NewTestApp(t)

	// Create epics collection
	collection := testutil.CreateTestCollection(t, app, "epics",
		&core.TextField{Name: "title", Required: true},
		&core.TextField{Name: "description"},
		&core.TextField{Name: "color"},
	)

	// Create a test epic
	epic := core.NewRecord(collection)
	epic.Set("title", "Q1 Launch")
	epic.Set("color", "#3B82F6")
	require.NoError(t, app.Save(epic))

	t.Run("resolves by exact ID", func(t *testing.T) {
		resolved, err := resolveEpic(app, epic.Id)
		require.NoError(t, err)
		assert.Equal(t, epic.Id, resolved.Id)
	})

	t.Run("resolves by ID prefix", func(t *testing.T) {
		resolved, err := resolveEpic(app, epic.Id[:8])
		require.NoError(t, err)
		assert.Equal(t, epic.Id, resolved.Id)
	})

	t.Run("resolves by title", func(t *testing.T) {
		resolved, err := resolveEpic(app, "Q1 Launch")
		require.NoError(t, err)
		assert.Equal(t, epic.Id, resolved.Id)
	})

	t.Run("resolves by partial title (case-insensitive)", func(t *testing.T) {
		resolved, err := resolveEpic(app, "q1")
		require.NoError(t, err)
		assert.Equal(t, epic.Id, resolved.Id)
	})

	t.Run("fails on not found", func(t *testing.T) {
		_, err := resolveEpic(app, "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no epic found")
	})

	t.Run("fails on ambiguous", func(t *testing.T) {
		// Create another epic with similar title
		epic2 := core.NewRecord(collection)
		epic2.Set("title", "Q1 Launch Prep")
		require.NoError(t, app.Save(epic2))

		_, err := resolveEpic(app, "Q1 Launch")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ambiguous")
	})
}

func TestGetEpicTaskCount(t *testing.T) {
	app := testutil.NewTestApp(t)

	// Create epics collection
	epicsCollection := testutil.CreateTestCollection(t, app, "epics",
		&core.TextField{Name: "title", Required: true},
	)

	// Create tasks collection with epic relation
	tasksCollection := testutil.CreateTestCollection(t, app, "tasks",
		&core.TextField{Name: "title", Required: true},
		&core.RelationField{Name: "epic", CollectionId: epicsCollection.Id, MaxSelect: 1},
	)

	// Create an epic
	epic := core.NewRecord(epicsCollection)
	epic.Set("title", "Test Epic")
	require.NoError(t, app.Save(epic))

	t.Run("returns 0 for epic with no tasks", func(t *testing.T) {
		count := getEpicTaskCount(app, epic.Id)
		assert.Equal(t, 0, count)
	})

	t.Run("returns correct count for epic with tasks", func(t *testing.T) {
		// Create tasks linked to the epic
		for i := 0; i < 3; i++ {
			task := core.NewRecord(tasksCollection)
			task.Set("title", fmt.Sprintf("Task %d", i))
			task.Set("epic", epic.Id)
			require.NoError(t, app.Save(task))
		}

		count := getEpicTaskCount(app, epic.Id)
		assert.Equal(t, 3, count)
	})
}
```

---

### 3.12 Write Tests for Batch Operations

**File**: `internal/commands/add_batch_test.go`

```go
package commands

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBatchInput_JSONLines(t *testing.T) {
	input := `{"title":"Task 1","type":"bug"}
{"title":"Task 2","priority":"high"}
{"title":"Task 3","column":"todo"}`

	reader := strings.NewReader(input)
	
	// This tests the parsing logic - you may need to extract it into a testable function
	var inputs []TaskInput
	
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var input TaskInput
		err := json.Unmarshal([]byte(line), &input)
		require.NoError(t, err)
		inputs = append(inputs, input)
	}

	assert.Len(t, inputs, 3)
	assert.Equal(t, "Task 1", inputs[0].Title)
	assert.Equal(t, "bug", inputs[0].Type)
	assert.Equal(t, "Task 2", inputs[1].Title)
	assert.Equal(t, "high", inputs[1].Priority)
	assert.Equal(t, "Task 3", inputs[2].Title)
	assert.Equal(t, "todo", inputs[2].Column)
}

func TestParseBatchInput_JSONArray(t *testing.T) {
	input := `[
		{"title":"Task 1"},
		{"title":"Task 2","labels":["frontend","ui"]}
	]`

	var inputs []TaskInput
	err := json.Unmarshal([]byte(input), &inputs)
	require.NoError(t, err)

	assert.Len(t, inputs, 2)
	assert.Equal(t, "Task 1", inputs[0].Title)
	assert.Equal(t, "Task 2", inputs[1].Title)
	assert.Len(t, inputs[1].Labels, 2)
	assert.Contains(t, inputs[1].Labels, "frontend")
	assert.Contains(t, inputs[1].Labels, "ui")
}

func TestDefaultString(t *testing.T) {
	assert.Equal(t, "default", defaultString("", "default"))
	assert.Equal(t, "value", defaultString("value", "default"))
}
```

---

## Verification Checklist

### Epic Commands

- [ ] **Epic list works**
  ```bash
  egenskriven epic list
  egenskriven epic list --json
  ```

- [ ] **Epic add works**
  ```bash
  egenskriven epic add "Test Epic"
  egenskriven epic add "Colored Epic" --color "#3B82F6"
  egenskriven epic add "With Description" --description "Test description"
  ```

- [ ] **Epic show works**
  ```bash
  egenskriven epic show "Test Epic"
  egenskriven epic show <id> --json
  ```

- [ ] **Epic delete works**
  ```bash
  egenskriven epic delete "Test Epic" --force
  ```

- [ ] **Task-epic linking works**
  ```bash
  egenskriven add "Task in Epic" --epic "Test Epic"
  egenskriven list --epic "Test Epic"
  ```

### Batch Operations

- [ ] **Batch add from stdin (JSON lines)**
  ```bash
  echo '{"title":"Task 1"}
  {"title":"Task 2"}' | egenskriven add --stdin
  ```

- [ ] **Batch add from stdin (JSON array)**
  ```bash
  echo '[{"title":"Task 1"},{"title":"Task 2"}]' | egenskriven add --stdin
  ```

- [ ] **Batch add from file**
  ```bash
  echo '[{"title":"File Task"}]' > /tmp/tasks.json
  egenskriven add --file /tmp/tasks.json
  ```

- [ ] **Batch delete from stdin** (NEW)
  ```bash
  echo -e "abc123\ndef456" | egenskriven delete --stdin --force
  ```

### Advanced Filters (New)

- [ ] **Filter by label**
  ```bash
  egenskriven add "Labeled Task" --label frontend --label ui
  egenskriven list --label frontend
  ```

- [ ] **Limit results**
  ```bash
  egenskriven list --limit 5
  ```

- [ ] **Custom sort**
  ```bash
  egenskriven list --sort "-priority,position"
  ```

### Version Command

- [ ] **Version displays correctly**
  ```bash
  egenskriven version
  egenskriven version --json
  ```

### Tests

- [ ] **All tests pass**
  ```bash
  make test
  ```

---

## File Summary

| File | Status | Lines (approx) | Purpose |
|------|--------|----------------|---------|
| `migrations/2_epics.go` | NEW | ~50 | Epics collection migration |
| `migrations/3_epic_relation.go` | NEW | ~40 | Epic relation in tasks |
| `internal/commands/epic.go` | NEW | ~350 | Epic CRUD commands |
| `internal/commands/epic_test.go` | NEW | ~120 | Epic tests |
| `internal/commands/version.go` | NEW | ~50 | Version command |
| `internal/commands/add_batch_test.go` | NEW | ~60 | Batch operation tests |
| `internal/commands/add.go` | MODIFY | +150 | Add batch input, epic flag |
| `internal/commands/delete.go` | MODIFY | +30 | Add stdin flag |
| `internal/commands/list.go` | MODIFY | +50 | Add label, limit, sort, epic flags |
| `internal/commands/root.go` | MODIFY | +5 | Register epic, version commands |
| `internal/output/output.go` | MODIFY | +20 | Add suggestion to errors |
| `Makefile` | MODIFY | +10 | Add ldflags for version |

**Total new code**: ~900 lines
**Total modifications**: ~265 lines

---

## What You Should Have Now

After completing Phase 3:

```
egenskriven/
├── migrations/
│   ├── 1_initial.go           (existing - tasks)
│   ├── 2_epics.go             NEW - epics collection
│   └── 3_epic_relation.go     NEW - epic field in tasks
├── internal/
│   └── commands/
│       ├── root.go            MODIFY - register epic, version
│       ├── add.go             MODIFY - batch input, epic flag
│       ├── delete.go          MODIFY - stdin flag
│       ├── list.go            MODIFY - label, limit, sort, epic flags
│       ├── epic.go            NEW - epic commands
│       ├── epic_test.go       NEW - epic tests
│       ├── version.go         NEW - version command
│       └── add_batch_test.go  NEW - batch tests
│   └── output/
│       └── output.go          MODIFY - error suggestions
└── Makefile                   MODIFY - version ldflags
```

---

## Next Phase

**Phase 4: Interactive UI** will add:
- Command palette (Cmd+K)
- Keyboard shortcuts for all actions
- Task selection state
- Property picker popovers
- Peek preview
- Real-time updates from CLI changes

---

## Troubleshooting

### "epics collection not found"

**Problem**: Epic migration hasn't run yet.

**Solution**: Restart the server to trigger migrations:
```bash
make run
```

### Batch input fails with "invalid JSON"

**Problem**: JSON format is incorrect.

**Solution**: Validate your JSON:
```bash
# Check if valid JSON
echo '{"title":"test"}' | jq .

# Common issues:
# - Missing quotes around strings
# - Trailing commas
# - Single quotes instead of double quotes
```

### Task-epic relation not working

**Problem**: The relation field wasn't properly created.

**Solution**: Check admin UI at `http://localhost:8090/_/` and verify:
1. Tasks collection has "epic" field
2. Field type is "Relation"
3. Relation points to "epics" collection

### Tests fail with "collection not found"

**Problem**: Test helpers aren't creating collections.

**Solution**: Ensure collections are created in tests before creating records. Use `testutil.CreateTestCollection`.

### Color validation too strict/loose

**Problem**: Hex color validation isn't working as expected.

**Solution**: The `isValidHexColor` function expects exactly `#RRGGBB` format:
- Valid: `#3B82F6`, `#aabbcc`
- Invalid: `red`, `#FFF`, `3B82F6`, `#GGGGGG`

### List --limit not working

**Problem**: `FindAllRecords` doesn't support limit.

**Solution**: Either:
1. Switch to `RecordQuery` for more control
2. Apply limit in Go after fetching all records (simpler but less efficient)

---

## Glossary

| Term | Definition |
|------|------------|
| **Epic** | A large initiative grouping multiple related tasks |
| **Batch input** | Creating/deleting multiple items in one command |
| **JSON lines** | Format with one JSON object per line (not an array) |
| **ldflags** | Linker flags to embed values at build time |
| **Field selection** | Requesting only specific fields in JSON output |
