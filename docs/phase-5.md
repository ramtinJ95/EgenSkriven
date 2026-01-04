# Phase 5: Multi-Board Support

**Goal**: Support multiple boards with board-specific task organization, CLI board management, and UI board switching.

**Duration Estimate**: 3-4 days

**Prerequisites**: 
- Phase 2 complete (UI exists)
- Phase 3 complete (full CLI)
- Phase 4 complete (interactive UI with command palette)

**Deliverable**: Users can create and manage multiple boards (e.g., "Work", "Personal"), switch between them in both CLI and UI, and tasks display board-prefixed IDs.

---

## Overview

EgenSkriven currently operates with a single implicit board. This phase adds multi-board support, allowing users to organize tasks into separate boards with distinct prefixes, columns, and colors.

### Why Multi-Board?

Different contexts need different organization:
- **Work**: Bug tracking with custom columns (Backlog, Sprint, Review, QA, Done)
- **Personal**: Simple kanban (Todo, Doing, Done)
- **Side Projects**: Standard flow with different priorities

Each board gets:
- A unique **prefix** for task IDs (e.g., `WRK-123`, `PER-456`)
- Optional **custom columns** (or use defaults)
- An **accent color** for visual distinction in the UI
- Isolated task lists

### Architecture Changes

```
Before (Single Board):
┌─────────────────────────────────────────────┐
│ tasks                                       │
│ - id, title, column, position, ...          │
└─────────────────────────────────────────────┘

After (Multi-Board):
┌─────────────────────────────────────────────┐
│ boards                                      │
│ - id, name, prefix, columns, color          │
├─────────────────────────────────────────────┤
│ tasks                                       │
│ - id, title, column, position, ...          │
│ - board (relation to boards)                │
└─────────────────────────────────────────────┘
```

### Display ID vs Storage ID

- **Storage**: Tasks still use auto-generated PocketBase IDs (`abc123def456`)
- **Display**: Tasks show board-prefixed IDs (`WRK-123`)
- **Resolution**: CLI accepts both formats; UI always shows display format

---

## Environment Requirements

Before starting, ensure you have completed:

| Phase | Verification |
|-------|--------------|
| Phase 2 | `make build && ./egenskriven serve` shows UI at localhost:8090 |
| Phase 3 | `egenskriven epic list` works (full CLI) |
| Phase 4 | `Cmd+K` opens command palette in UI |

---

## Tasks

### 5.1 Create Boards Collection Migration

**What**: Add a `boards` collection to store board metadata.

**Why**: Each board needs persistent configuration: name, prefix, columns, and styling.

**File**: `migrations/2_boards.go`

```go
package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Create boards collection
		collection := core.NewBaseCollection("boards")

		// Name: Human-readable board name
		// Examples: "Work", "Personal", "Side Projects"
		collection.Fields.Add(&core.TextField{
			Name:     "name",
			Required: true,
			Min:      1,
			Max:      100,
		})

		// Prefix: Short uppercase identifier for task IDs
		// Examples: "WRK", "PER", "SIDE"
		// Must be unique across all boards
		collection.Fields.Add(&core.TextField{
			Name:     "prefix",
			Required: true,
			Min:      1,
			Max:      10,
			// Validated as uppercase in application code
		})

		// Columns: JSON array of column definitions
		// Default: ["backlog", "todo", "in_progress", "review", "done"]
		// Allows boards to have custom workflows
		collection.Fields.Add(&core.JSONField{
			Name:    "columns",
			MaxSize: 10000,
		})

		// Color: Hex color for board accent
		// Used in UI for board identification
		// Examples: "#3B82F6" (blue), "#22C55E" (green)
		collection.Fields.Add(&core.TextField{
			Name: "color",
			Max:  7, // #RRGGBB format
		})

		// Add unique index on prefix
		collection.Indexes = []string{
			"CREATE UNIQUE INDEX idx_boards_prefix ON boards(prefix)",
		}

		if err := app.Save(collection); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// Rollback: drop the boards collection
		collection, err := app.FindCollectionByNameOrId("boards")
		if err != nil {
			return err
		}
		return app.Delete(collection)
	})
}
```

**Steps**:

1. Create the file:
   ```bash
   touch migrations/2_boards.go
   ```

2. Open in your editor and paste the code above.

3. Verify it compiles:
   ```bash
   go build ./migrations
   ```
   
   **Expected output**: No output means success!

**Important Notes**:
- The `prefix` field has a unique index to prevent duplicates
- `columns` is JSON to allow flexible column configurations
- Migration includes rollback function for safety

---

### 5.2 Update Tasks Collection with Board Relation

**What**: Add a `board` relation field to the tasks collection.

**Why**: Each task must belong to a board. The relation links tasks to their parent board.

**File**: `migrations/3_tasks_board_relation.go`

```go
package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Get existing tasks collection
		collection, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}

		// Add board relation field
		// Each task belongs to exactly one board
		collection.Fields.Add(&core.RelationField{
			Name:          "board",
			CollectionId:  "", // Will be set to boards collection ID
			Required:      false, // Initially false for migration of existing tasks
			MaxSelect:     1,
			CascadeDelete: false, // Don't delete tasks when board is deleted (handle manually)
		})

		// We need to find the boards collection ID
		boardsCollection, err := app.FindCollectionByNameOrId("boards")
		if err != nil {
			// Boards collection doesn't exist yet - this is OK during initial setup
			// The field will be updated when boards collection is created
			return app.Save(collection)
		}

		// Update the relation to point to boards collection
		boardField := collection.Fields.GetByName("board")
		if boardField != nil {
			if relationField, ok := boardField.(*core.RelationField); ok {
				relationField.CollectionId = boardsCollection.Id
			}
		}

		// Add index for faster queries by board
		// This significantly improves performance when listing tasks for a specific board
		collection.Indexes = append(collection.Indexes,
			"CREATE INDEX idx_tasks_board ON tasks(board)",
		)

		return app.Save(collection)
	}, func(app core.App) error {
		// Rollback: remove board field from tasks
		collection, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}

		// Remove the board field
		collection.Fields.RemoveByName("board")

		// Remove the index
		// Note: PocketBase handles index cleanup automatically when field is removed

		return app.Save(collection)
	})
}
```

**Steps**:

1. Create the file:
   ```bash
   touch migrations/3_tasks_board_relation.go
   ```

2. Open in your editor and paste the code above.

3. Verify it compiles:
   ```bash
   go build ./migrations
   ```

**Migration Order**:
- Migration 1: Creates `tasks` collection (from Phase 1)
- Migration 2: Creates `boards` collection (this phase)
- Migration 3: Adds `board` relation to `tasks` (this phase)

---

### 5.3 Add Task Sequence Number Field

**What**: Add a `seq` (sequence number) field to tasks for display IDs.

**Why**: Display IDs like `WRK-123` require a per-board incrementing counter. This field stores the sequence number.

**File**: `migrations/4_tasks_sequence.go`

```go
package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Get existing tasks collection
		collection, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}

		// Add sequence number field
		// This is the numeric part of the display ID (e.g., 123 in WRK-123)
		// Auto-incremented per board when creating tasks
		collection.Fields.Add(&core.NumberField{
			Name: "seq",
			Min:  floatPtr(1),
		})

		// Add compound index for efficient sequence queries
		// Used to find max sequence for a board when creating new tasks
		collection.Indexes = append(collection.Indexes,
			"CREATE INDEX idx_tasks_board_seq ON tasks(board, seq)",
		)

		return app.Save(collection)
	}, func(app core.App) error {
		// Rollback: remove seq field
		collection, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}
		collection.Fields.RemoveByName("seq")
		return app.Save(collection)
	})
}

// Helper to create a pointer to float64
func floatPtr(f float64) *float64 {
	return &f
}
```

**Steps**:

1. Create the file:
   ```bash
   touch migrations/4_tasks_sequence.go
   ```

2. Open in your editor and paste the code above.

3. Verify it compiles:
   ```bash
   go build ./migrations
   ```

---

### 5.4 Create Board Service

**What**: Create a service layer for board operations including display ID generation.

**Why**: Centralizes board-related logic: creating boards, generating display IDs, resolving board references.

**File**: `internal/board/board.go`

```go
package board

import (
	"fmt"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// DefaultColumns are used when creating a board without custom columns
var DefaultColumns = []string{"backlog", "todo", "in_progress", "review", "done"}

// Board represents a board with its metadata
type Board struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Prefix  string   `json:"prefix"`
	Columns []string `json:"columns"`
	Color   string   `json:"color,omitempty"`
}

// CreateInput contains the data needed to create a board
type CreateInput struct {
	Name    string   // Required: Human-readable name
	Prefix  string   // Required: Uppercase prefix for task IDs
	Columns []string // Optional: Custom columns (defaults to DefaultColumns)
	Color   string   // Optional: Hex color code
}

// Create creates a new board with the given input
//
// The prefix is automatically uppercased and validated:
// - Must be 1-10 characters
// - Must be alphanumeric (no spaces or special characters)
// - Must be unique across all boards
func Create(app *pocketbase.PocketBase, input CreateInput) (*Board, error) {
	// Validate and normalize prefix
	prefix := strings.ToUpper(strings.TrimSpace(input.Prefix))
	if len(prefix) == 0 {
		return nil, fmt.Errorf("prefix is required")
	}
	if len(prefix) > 10 {
		return nil, fmt.Errorf("prefix must be 10 characters or less")
	}
	if !isAlphanumeric(prefix) {
		return nil, fmt.Errorf("prefix must be alphanumeric (letters and numbers only)")
	}

	// Validate name
	name := strings.TrimSpace(input.Name)
	if len(name) == 0 {
		return nil, fmt.Errorf("name is required")
	}

	// Check prefix uniqueness
	existing, _ := app.FindFirstRecordByData("boards", "prefix", prefix)
	if existing != nil {
		return nil, fmt.Errorf("prefix '%s' is already in use by board '%s'",
			prefix, existing.GetString("name"))
	}

	// Use default columns if none provided
	columns := input.Columns
	if len(columns) == 0 {
		columns = DefaultColumns
	}

	// Get boards collection
	collection, err := app.FindCollectionByNameOrId("boards")
	if err != nil {
		return nil, fmt.Errorf("boards collection not found: %w", err)
	}

	// Create record
	record := core.NewRecord(collection)
	record.Set("name", name)
	record.Set("prefix", prefix)
	record.Set("columns", columns)
	if input.Color != "" {
		record.Set("color", input.Color)
	}

	if err := app.Save(record); err != nil {
		return nil, fmt.Errorf("failed to create board: %w", err)
	}

	return &Board{
		ID:      record.Id,
		Name:    name,
		Prefix:  prefix,
		Columns: columns,
		Color:   input.Color,
	}, nil
}

// GetByNameOrPrefix finds a board by name or prefix (case-insensitive)
//
// This allows flexible board references in the CLI:
//   - "Work" (name)
//   - "work" (name, case-insensitive)
//   - "WRK" (prefix)
//   - "wrk" (prefix, case-insensitive)
func GetByNameOrPrefix(app *pocketbase.PocketBase, ref string) (*core.Record, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil, fmt.Errorf("board reference is required")
	}

	// Try exact ID match first
	if record, err := app.FindRecordById("boards", ref); err == nil {
		return record, nil
	}

	// Try case-insensitive prefix match
	records, err := app.FindAllRecords("boards",
		dbx.NewExp("LOWER(prefix) = {:prefix}", dbx.Params{"prefix": strings.ToLower(ref)}),
	)
	if err == nil && len(records) == 1 {
		return records[0], nil
	}

	// Try case-insensitive name match
	records, err = app.FindAllRecords("boards",
		dbx.NewExp("LOWER(name) = {:name}", dbx.Params{"name": strings.ToLower(ref)}),
	)
	if err == nil && len(records) == 1 {
		return records[0], nil
	}

	// Try partial name match (for convenience)
	records, err = app.FindAllRecords("boards",
		dbx.NewExp("LOWER(name) LIKE {:name}", dbx.Params{"name": "%" + strings.ToLower(ref) + "%"}),
	)
	if err == nil && len(records) == 1 {
		return records[0], nil
	}
	if len(records) > 1 {
		return nil, fmt.Errorf("ambiguous board reference '%s' matches multiple boards", ref)
	}

	return nil, fmt.Errorf("board not found: %s", ref)
}

// GetAll returns all boards
func GetAll(app *pocketbase.PocketBase) ([]*core.Record, error) {
	return app.FindAllRecords("boards", dbx.NewExp("1=1"))
}

// GetNextSequence returns the next sequence number for a board
//
// This is used when creating tasks to generate display IDs.
// Uses SELECT MAX(seq) + 1 for efficiency.
func GetNextSequence(app *pocketbase.PocketBase, boardID string) (int, error) {
	// Find the maximum sequence number for this board
	records, err := app.FindAllRecords("tasks",
		dbx.NewExp("board = {:board}", dbx.Params{"board": boardID}),
		dbx.OrderBy("seq DESC"),
		dbx.Limit(1),
	)
	if err != nil {
		return 1, nil // First task in board
	}

	if len(records) == 0 {
		return 1, nil // First task in board
	}

	maxSeq := records[0].GetInt("seq")
	return maxSeq + 1, nil
}

// FormatDisplayID creates a display ID from board prefix and sequence
//
// Example: FormatDisplayID("WRK", 123) returns "WRK-123"
func FormatDisplayID(prefix string, seq int) string {
	return fmt.Sprintf("%s-%d", prefix, seq)
}

// ParseDisplayID extracts the prefix and sequence from a display ID
//
// Example: ParseDisplayID("WRK-123") returns ("WRK", 123, nil)
func ParseDisplayID(displayID string) (prefix string, seq int, err error) {
	parts := strings.SplitN(displayID, "-", 2)
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid display ID format: %s (expected PREFIX-NUMBER)", displayID)
	}

	prefix = strings.ToUpper(parts[0])
	_, err = fmt.Sscanf(parts[1], "%d", &seq)
	if err != nil {
		return "", 0, fmt.Errorf("invalid sequence number in display ID: %s", displayID)
	}

	return prefix, seq, nil
}

// RecordToBoard converts a PocketBase record to a Board struct
func RecordToBoard(record *core.Record) *Board {
	columns := DefaultColumns
	if c := record.Get("columns"); c != nil {
		if arr, ok := c.([]interface{}); ok {
			columns = make([]string, len(arr))
			for i, v := range arr {
				columns[i] = fmt.Sprint(v)
			}
		}
	}

	return &Board{
		ID:      record.Id,
		Name:    record.GetString("name"),
		Prefix:  record.GetString("prefix"),
		Columns: columns,
		Color:   record.GetString("color"),
	}
}

// Delete removes a board and optionally its tasks
//
// If deleteTasks is false, tasks are orphaned (board field cleared).
// If deleteTasks is true, all tasks in the board are deleted.
func Delete(app *pocketbase.PocketBase, boardID string, deleteTasks bool) error {
	board, err := app.FindRecordById("boards", boardID)
	if err != nil {
		return fmt.Errorf("board not found: %w", err)
	}

	if deleteTasks {
		// Delete all tasks in this board
		tasks, err := app.FindAllRecords("tasks",
			dbx.NewExp("board = {:board}", dbx.Params{"board": boardID}),
		)
		if err == nil {
			for _, task := range tasks {
				if err := app.Delete(task); err != nil {
					return fmt.Errorf("failed to delete task %s: %w", task.Id, err)
				}
			}
		}
	} else {
		// Clear board reference from tasks (orphan them)
		tasks, err := app.FindAllRecords("tasks",
			dbx.NewExp("board = {:board}", dbx.Params{"board": boardID}),
		)
		if err == nil {
			for _, task := range tasks {
				task.Set("board", "")
				if err := app.Save(task); err != nil {
					return fmt.Errorf("failed to update task %s: %w", task.Id, err)
				}
			}
		}
	}

	return app.Delete(board)
}

// isAlphanumeric checks if a string contains only letters and numbers
func isAlphanumeric(s string) bool {
	for _, r := range s {
		if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}
```

**Steps**:

1. Create the directory and file:
   ```bash
   mkdir -p internal/board
   touch internal/board/board.go
   ```

2. Open in your editor and paste the code above.

3. Verify it compiles:
   ```bash
   go build ./internal/board
   ```

---

### 5.5 Implement Board CLI Commands

**What**: Add CLI commands for board management.

**Why**: Users need to create, list, switch, and delete boards from the command line.

**File**: `internal/commands/board.go`

```go
package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"

	"github.com/yourusername/egenskriven/internal/board"
	"github.com/yourusername/egenskriven/internal/config"
	"github.com/yourusername/egenskriven/internal/output"
)

// NewBoardCmd creates the board command and its subcommands
func NewBoardCmd(app *pocketbase.PocketBase, out *output.Formatter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "board",
		Short: "Manage boards",
		Long: `Manage multiple boards for organizing tasks.

Each board has:
- A unique name (e.g., "Work", "Personal")
- A prefix for task IDs (e.g., "WRK", "PER")
- Optional custom columns
- An optional accent color`,
		Example: `  egenskriven board list
  egenskriven board add "Work" --prefix WRK
  egenskriven board show work
  egenskriven board use work
  egenskriven board delete work --force`,
	}

	cmd.AddCommand(newBoardListCmd(app, out))
	cmd.AddCommand(newBoardAddCmd(app, out))
	cmd.AddCommand(newBoardShowCmd(app, out))
	cmd.AddCommand(newBoardUseCmd(app, out))
	cmd.AddCommand(newBoardDeleteCmd(app, out))

	return cmd
}

// newBoardListCmd creates the 'board list' subcommand
func newBoardListCmd(app *pocketbase.PocketBase, out *output.Formatter) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all boards",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := app.Bootstrap(); err != nil {
				return err
			}

			records, err := board.GetAll(app)
			if err != nil {
				return err
			}

			boards := make([]*board.Board, len(records))
			for i, r := range records {
				boards[i] = board.RecordToBoard(r)
			}

			// Get current board from config
			cfg, _ := config.LoadProjectConfig()
			currentBoard := ""
			if cfg != nil {
				currentBoard = cfg.DefaultBoard
			}

			if out.JSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
					"boards":        boards,
					"current_board": currentBoard,
					"count":         len(boards),
				})
			}

			if len(boards) == 0 {
				fmt.Println("No boards found. Create one with: egenskriven board add \"Name\" --prefix PREFIX")
				return nil
			}

			fmt.Println("BOARDS")
			fmt.Println("------")
			for _, b := range boards {
				marker := "  "
				if b.Prefix == currentBoard || b.Name == currentBoard {
					marker = "> "
				}
				fmt.Printf("%s%s (%s)\n", marker, b.Name, b.Prefix)
			}

			return nil
		},
	}
}

// newBoardAddCmd creates the 'board add' subcommand
func newBoardAddCmd(app *pocketbase.PocketBase, out *output.Formatter) *cobra.Command {
	var (
		prefix  string
		color   string
		columns []string
	)

	cmd := &cobra.Command{
		Use:   "add [name]",
		Short: "Create a new board",
		Long: `Create a new board with a unique prefix.

The prefix is used in task IDs (e.g., WRK-123) and must be:
- 1-10 characters
- Alphanumeric (letters and numbers only)
- Unique across all boards`,
		Args: cobra.ExactArgs(1),
		Example: `  egenskriven board add "Work" --prefix WRK
  egenskriven board add "Personal" --prefix PER --color "#22C55E"
  egenskriven board add "Sprint" --prefix SPR --columns "backlog,ready,doing,review,done"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := app.Bootstrap(); err != nil {
				return err
			}

			name := args[0]
			if prefix == "" {
				return fmt.Errorf("--prefix is required")
			}

			b, err := board.Create(app, board.CreateInput{
				Name:    name,
				Prefix:  prefix,
				Columns: columns,
				Color:   color,
			})
			if err != nil {
				return err
			}

			if out.JSON {
				return json.NewEncoder(os.Stdout).Encode(b)
			}

			fmt.Printf("Created board: %s (%s)\n", b.Name, b.Prefix)
			return nil
		},
	}

	cmd.Flags().StringVarP(&prefix, "prefix", "p", "", "Task ID prefix (required, e.g., WRK)")
	cmd.Flags().StringVarP(&color, "color", "c", "", "Accent color (hex, e.g., #3B82F6)")
	cmd.Flags().StringSliceVar(&columns, "columns", nil, "Custom columns (comma-separated)")
	cmd.MarkFlagRequired("prefix")

	return cmd
}

// newBoardShowCmd creates the 'board show' subcommand
func newBoardShowCmd(app *pocketbase.PocketBase, out *output.Formatter) *cobra.Command {
	return &cobra.Command{
		Use:   "show [name-or-prefix]",
		Short: "Show board details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := app.Bootstrap(); err != nil {
				return err
			}

			record, err := board.GetByNameOrPrefix(app, args[0])
			if err != nil {
				return err
			}

			b := board.RecordToBoard(record)

			// Count tasks in this board
			tasks, _ := app.FindAllRecords("tasks",
				dbx.NewExp("board = {:board}", dbx.Params{"board": record.Id}),
			)
			taskCount := len(tasks)

			if out.JSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
					"id":         b.ID,
					"name":       b.Name,
					"prefix":     b.Prefix,
					"columns":    b.Columns,
					"color":      b.Color,
					"task_count": taskCount,
				})
			}

			fmt.Printf("Board: %s\n", b.Name)
			fmt.Printf("Prefix: %s\n", b.Prefix)
			fmt.Printf("Columns: %v\n", b.Columns)
			if b.Color != "" {
				fmt.Printf("Color: %s\n", b.Color)
			}
			fmt.Printf("Tasks: %d\n", taskCount)

			return nil
		},
	}
}

// newBoardUseCmd creates the 'board use' subcommand
func newBoardUseCmd(app *pocketbase.PocketBase, out *output.Formatter) *cobra.Command {
	return &cobra.Command{
		Use:   "use [name-or-prefix]",
		Short: "Set the default board for CLI commands",
		Long: `Set the default board used by CLI commands.

When set, commands like 'add', 'list', 'show' will operate on this board
unless overridden with the --board flag.

The setting is stored in .egenskriven/config.json in the current directory.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := app.Bootstrap(); err != nil {
				return err
			}

			record, err := board.GetByNameOrPrefix(app, args[0])
			if err != nil {
				return err
			}

			// Update project config
			cfg, err := config.LoadProjectConfig()
			if err != nil {
				cfg = &config.Config{}
			}

			cfg.DefaultBoard = record.GetString("prefix")

			if err := config.SaveProjectConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			if out.JSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]string{
					"default_board": cfg.DefaultBoard,
				})
			}

			fmt.Printf("Default board set to: %s (%s)\n",
				record.GetString("name"),
				record.GetString("prefix"))

			return nil
		},
	}
}

// newBoardDeleteCmd creates the 'board delete' subcommand
func newBoardDeleteCmd(app *pocketbase.PocketBase, out *output.Formatter) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [name-or-prefix]",
		Short: "Delete a board",
		Long: `Delete a board and all its tasks.

WARNING: This permanently deletes all tasks in the board!
Use --force to skip the confirmation prompt.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := app.Bootstrap(); err != nil {
				return err
			}

			record, err := board.GetByNameOrPrefix(app, args[0])
			if err != nil {
				return err
			}

			boardName := record.GetString("name")
			boardPrefix := record.GetString("prefix")

			// Count tasks that will be deleted
			tasks, _ := app.FindAllRecords("tasks",
				dbx.NewExp("board = {:board}", dbx.Params{"board": record.Id}),
			)
			taskCount := len(tasks)

			// Confirm deletion
			if !force && taskCount > 0 {
				fmt.Printf("WARNING: Board '%s' contains %d task(s).\n", boardName, taskCount)
				fmt.Print("Delete board and all its tasks? [y/N]: ")

				var confirm string
				fmt.Scanln(&confirm)
				if confirm != "y" && confirm != "Y" {
					fmt.Println("Cancelled.")
					return nil
				}
			}

			// Delete board and tasks
			if err := board.Delete(app, record.Id, true); err != nil {
				return err
			}

			if out.JSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
					"deleted":       true,
					"board":         boardName,
					"prefix":        boardPrefix,
					"tasks_deleted": taskCount,
				})
			}

			fmt.Printf("Deleted board '%s' and %d task(s)\n", boardName, taskCount)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}
```

**Steps**:

1. Create the file:
   ```bash
   touch internal/commands/board.go
   ```

2. Open in your editor and paste the code above.

3. Add missing import for dbx:
   ```go
   import (
       "github.com/pocketbase/dbx"
       // ... other imports
   )
   ```

4. Verify it compiles:
   ```bash
   go build ./internal/commands
   ```

---

### 5.6 Update Config to Store Default Board

**What**: Extend the config structure to persist the default board setting.

**Why**: Users should be able to set a default board that persists across CLI sessions.

**File**: Update `internal/config/config.go`

Add these fields and functions to the existing config:

```go
// Add to Config struct
type Config struct {
	Agent        AgentConfig `json:"agent"`
	DefaultBoard string      `json:"default_board,omitempty"` // ADD THIS
}

// Add SaveProjectConfig function
func SaveProjectConfig(cfg *Config) error {
	configDir := ".egenskriven"
	configPath := filepath.Join(configDir, "config.json")

	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
```

**Steps**:

1. Open `internal/config/config.go`.

2. Add the `DefaultBoard` field to the `Config` struct.

3. Add the `SaveProjectConfig` function.

4. Verify it compiles:
   ```bash
   go build ./internal/config
   ```

---

### 5.7 Update Add Command for Multi-Board

**What**: Modify the `add` command to support boards and generate display IDs.

**Why**: Tasks must be created in a specific board with a proper sequence number for display IDs.

**File**: Update `internal/commands/add.go`

Key changes to implement:

```go
// Add --board flag
var boardFlag string

func init() {
	addCmd.Flags().StringVarP(&boardFlag, "board", "b", "", "Board to create task in")
}

// In RunE function, resolve the board:
func createTask(app *pocketbase.PocketBase, input TaskInput) (*core.Record, error) {
	// Determine which board to use
	boardRef := input.Board
	if boardRef == "" {
		// Check config for default board
		cfg, _ := config.LoadProjectConfig()
		if cfg != nil && cfg.DefaultBoard != "" {
			boardRef = cfg.DefaultBoard
		}
	}

	var boardRecord *core.Record
	var err error

	if boardRef != "" {
		boardRecord, err = board.GetByNameOrPrefix(app, boardRef)
		if err != nil {
			return nil, fmt.Errorf("invalid board: %w", err)
		}
	} else {
		// Get or create default board if none specified
		boards, _ := board.GetAll(app)
		if len(boards) == 0 {
			// Create a default board
			b, err := board.Create(app, board.CreateInput{
				Name:   "Default",
				Prefix: "DEF",
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create default board: %w", err)
			}
			boardRecord, _ = app.FindRecordById("boards", b.ID)
		} else {
			boardRecord = boards[0]
		}
	}

	// Get next sequence number for this board
	seq, err := board.GetNextSequence(app, boardRecord.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get sequence: %w", err)
	}

	// Create the task record
	collection, err := app.FindCollectionByNameOrId("tasks")
	if err != nil {
		return nil, err
	}

	record := core.NewRecord(collection)
	record.Set("title", input.Title)
	record.Set("board", boardRecord.Id)
	record.Set("seq", seq)
	// ... set other fields

	if err := app.Save(record); err != nil {
		return nil, err
	}

	return record, nil
}
```

**Steps**:

1. Open `internal/commands/add.go`.

2. Add the `--board` flag.

3. Update the `createTask` function to:
   - Resolve the board from flag, config, or default
   - Get the next sequence number
   - Store `board` and `seq` fields

4. Update the output to show display ID:
   ```go
   displayID := board.FormatDisplayID(boardRecord.GetString("prefix"), seq)
   fmt.Printf("Created task: %s [%s]\n", input.Title, displayID)
   ```

5. Verify it compiles:
   ```bash
   go build ./internal/commands
   ```

---

### 5.8 Update List Command for Multi-Board

**What**: Modify the `list` command to filter by board and support `--all-boards`.

**Why**: Users should see only tasks from the current board by default, with an option to see all.

**File**: Update `internal/commands/list.go`

Key changes to implement:

```go
// Add flags
var (
	boardFlag    string
	allBoardsFlag bool
)

func init() {
	listCmd.Flags().StringVarP(&boardFlag, "board", "b", "", "Filter by board")
	listCmd.Flags().BoolVar(&allBoardsFlag, "all-boards", false, "Show tasks from all boards")
}

// In RunE, add board filtering:
func listTasks(app *pocketbase.PocketBase, filter Filter) ([]*core.Record, error) {
	var conditions []dbx.Expression

	// Board filtering
	if !filter.AllBoards {
		boardRef := filter.Board
		if boardRef == "" {
			cfg, _ := config.LoadProjectConfig()
			if cfg != nil && cfg.DefaultBoard != "" {
				boardRef = cfg.DefaultBoard
			}
		}

		if boardRef != "" {
			boardRecord, err := board.GetByNameOrPrefix(app, boardRef)
			if err != nil {
				return nil, err
			}
			conditions = append(conditions,
				dbx.NewExp("board = {:board}", dbx.Params{"board": boardRecord.Id}))
		}
	}

	// ... existing filter conditions

	return app.FindAllRecords("tasks", conditions...)
}

// Update output to show display IDs
func formatTask(task *core.Record, boardsMap map[string]*core.Record) string {
	boardRecord := boardsMap[task.GetString("board")]
	prefix := "???"
	if boardRecord != nil {
		prefix = boardRecord.GetString("prefix")
	}
	seq := task.GetInt("seq")
	displayID := board.FormatDisplayID(prefix, seq)

	return fmt.Sprintf("[%s] %s", displayID, task.GetString("title"))
}
```

**Steps**:

1. Open `internal/commands/list.go`.

2. Add `--board` and `--all-boards` flags.

3. Update the filtering logic.

4. Update output to show display IDs instead of internal IDs.

5. Verify it compiles:
   ```bash
   go build ./internal/commands
   ```

---

### 5.9 Update Task Resolver for Display IDs

**What**: Extend the resolver to accept display IDs (e.g., `WRK-123`).

**Why**: Users should be able to reference tasks by their human-friendly display IDs.

**File**: Update `internal/resolver/resolver.go`

Add display ID resolution:

```go
// ResolveTask now handles multiple ID formats:
// 1. Display ID: "WRK-123"
// 2. Internal ID: "abc123def456"
// 3. ID prefix: "abc123"
// 4. Title substring: "fix login"
func ResolveTask(app *pocketbase.PocketBase, ref string) (*Resolution, error) {
	ref = strings.TrimSpace(ref)

	// 1. Try display ID format (PREFIX-NUMBER)
	if prefix, seq, err := board.ParseDisplayID(ref); err == nil {
		// Find board by prefix
		boardRecord, err := board.GetByNameOrPrefix(app, prefix)
		if err == nil {
			// Find task by board + sequence
			tasks, err := app.FindAllRecords("tasks",
				dbx.NewExp("board = {:board} AND seq = {:seq}",
					dbx.Params{"board": boardRecord.Id, "seq": seq}),
			)
			if err == nil && len(tasks) == 1 {
				return &Resolution{Task: tasks[0]}, nil
			}
		}
	}

	// 2. Try exact internal ID match
	if task, err := app.FindRecordById("tasks", ref); err == nil {
		return &Resolution{Task: task}, nil
	}

	// 3. Try ID prefix match
	tasks, err := app.FindAllRecords("tasks",
		dbx.NewExp("id LIKE {:prefix}", dbx.Params{"prefix": ref + "%"}),
	)
	if err == nil && len(tasks) == 1 {
		return &Resolution{Task: tasks[0]}, nil
	}

	// 4. Try title match (existing logic)
	// ... keep existing title matching code
}
```

**Steps**:

1. Open `internal/resolver/resolver.go`.

2. Add display ID parsing as the first resolution attempt.

3. Import the `board` package.

4. Verify it compiles:
   ```bash
   go build ./internal/resolver
   ```

---

### 5.10 Create Board Service Tests

**What**: Write tests for the board service.

**Why**: Board creation, resolution, and sequence generation must work correctly.

**File**: `internal/board/board_test.go`

```go
package board

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourusername/egenskriven/internal/testutil"
)

func TestCreate_Success(t *testing.T) {
	app := testutil.NewTestApp(t)
	
	// Create boards collection (normally done by migration)
	testutil.SetupBoardsCollection(t, app)

	b, err := Create(app, CreateInput{
		Name:   "Work",
		Prefix: "WRK",
		Color:  "#3B82F6",
	})

	require.NoError(t, err)
	assert.Equal(t, "Work", b.Name)
	assert.Equal(t, "WRK", b.Prefix)
	assert.Equal(t, "#3B82F6", b.Color)
	assert.Equal(t, DefaultColumns, b.Columns)
}

func TestCreate_PrefixUppercased(t *testing.T) {
	app := testutil.NewTestApp(t)
	testutil.SetupBoardsCollection(t, app)

	b, err := Create(app, CreateInput{
		Name:   "Work",
		Prefix: "wrk", // lowercase
	})

	require.NoError(t, err)
	assert.Equal(t, "WRK", b.Prefix) // should be uppercased
}

func TestCreate_DuplicatePrefixFails(t *testing.T) {
	app := testutil.NewTestApp(t)
	testutil.SetupBoardsCollection(t, app)

	// Create first board
	_, err := Create(app, CreateInput{Name: "Work", Prefix: "WRK"})
	require.NoError(t, err)

	// Try to create second board with same prefix
	_, err = Create(app, CreateInput{Name: "Work 2", Prefix: "WRK"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already in use")
}

func TestCreate_InvalidPrefixFails(t *testing.T) {
	app := testutil.NewTestApp(t)
	testutil.SetupBoardsCollection(t, app)

	tests := []struct {
		name   string
		prefix string
	}{
		{"empty", ""},
		{"too long", "VERYLONGPREFIX"},
		{"special chars", "WRK@#"},
		{"spaces", "WR K"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Create(app, CreateInput{
				Name:   "Test",
				Prefix: tc.prefix,
			})
			assert.Error(t, err)
		})
	}
}

func TestGetByNameOrPrefix_ByPrefix(t *testing.T) {
	app := testutil.NewTestApp(t)
	testutil.SetupBoardsCollection(t, app)

	created, _ := Create(app, CreateInput{Name: "Work", Prefix: "WRK"})

	// Find by exact prefix
	found, err := GetByNameOrPrefix(app, "WRK")
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.Id)

	// Find by lowercase prefix
	found, err = GetByNameOrPrefix(app, "wrk")
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.Id)
}

func TestGetByNameOrPrefix_ByName(t *testing.T) {
	app := testutil.NewTestApp(t)
	testutil.SetupBoardsCollection(t, app)

	created, _ := Create(app, CreateInput{Name: "My Work Board", Prefix: "WRK"})

	// Find by exact name
	found, err := GetByNameOrPrefix(app, "My Work Board")
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.Id)

	// Find by partial name
	found, err = GetByNameOrPrefix(app, "work")
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.Id)
}

func TestGetNextSequence(t *testing.T) {
	app := testutil.NewTestApp(t)
	testutil.SetupBoardsCollection(t, app)
	testutil.SetupTasksCollection(t, app)

	b, _ := Create(app, CreateInput{Name: "Work", Prefix: "WRK"})

	// First task should get sequence 1
	seq1, err := GetNextSequence(app, b.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, seq1)

	// Create a task with that sequence
	testutil.CreateTask(t, app, map[string]interface{}{
		"title": "Task 1",
		"board": b.ID,
		"seq":   seq1,
	})

	// Next sequence should be 2
	seq2, err := GetNextSequence(app, b.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, seq2)
}

func TestFormatDisplayID(t *testing.T) {
	assert.Equal(t, "WRK-1", FormatDisplayID("WRK", 1))
	assert.Equal(t, "WRK-123", FormatDisplayID("WRK", 123))
	assert.Equal(t, "PER-9999", FormatDisplayID("PER", 9999))
}

func TestParseDisplayID(t *testing.T) {
	prefix, seq, err := ParseDisplayID("WRK-123")
	require.NoError(t, err)
	assert.Equal(t, "WRK", prefix)
	assert.Equal(t, 123, seq)

	// Invalid formats
	_, _, err = ParseDisplayID("invalid")
	assert.Error(t, err)

	_, _, err = ParseDisplayID("WRK-abc")
	assert.Error(t, err)
}

func TestDelete_DeletesTasks(t *testing.T) {
	app := testutil.NewTestApp(t)
	testutil.SetupBoardsCollection(t, app)
	testutil.SetupTasksCollection(t, app)

	b, _ := Create(app, CreateInput{Name: "Work", Prefix: "WRK"})

	// Create some tasks
	for i := 1; i <= 3; i++ {
		testutil.CreateTask(t, app, map[string]interface{}{
			"title": fmt.Sprintf("Task %d", i),
			"board": b.ID,
			"seq":   i,
		})
	}

	// Delete board with tasks
	err := Delete(app, b.ID, true)
	require.NoError(t, err)

	// Board should be gone
	_, err = app.FindRecordById("boards", b.ID)
	assert.Error(t, err)

	// Tasks should be gone too
	tasks, _ := app.FindAllRecords("tasks",
		dbx.NewExp("board = {:board}", dbx.Params{"board": b.ID}))
	assert.Len(t, tasks, 0)
}
```

**Steps**:

1. Create the file:
   ```bash
   touch internal/board/board_test.go
   ```

2. Open in your editor and paste the code above.

3. Add helper functions to `internal/testutil/testutil.go`:
   ```go
   func SetupBoardsCollection(t *testing.T, app *pocketbase.PocketBase) {
       // Create boards collection for testing
       // Similar to the migration but simplified
   }

   func SetupTasksCollection(t *testing.T, app *pocketbase.PocketBase) {
       // Create tasks collection with board relation
   }

   func CreateTask(t *testing.T, app *pocketbase.PocketBase, data map[string]interface{}) *core.Record {
       // Helper to create test tasks
   }
   ```

4. Run the tests:
   ```bash
   go test ./internal/board -v
   ```

---

### 5.11 Create Sidebar Component

**What**: Create a UI sidebar showing boards and enabling board switching.

**Why**: Users need to see and switch between boards in the UI.

**File**: `ui/src/components/Sidebar.tsx`

```tsx
import { useState, useEffect } from 'react';
import { useBoards, useCurrentBoard, Board } from '../hooks/usePocketBase';
import './Sidebar.css';

interface SidebarProps {
  collapsed: boolean;
  onToggle: () => void;
}

export function Sidebar({ collapsed, onToggle }: SidebarProps) {
  const { boards, loading } = useBoards();
  const { currentBoard, setCurrentBoard } = useCurrentBoard();
  const [showNewBoard, setShowNewBoard] = useState(false);

  if (collapsed) {
    return (
      <div className="sidebar collapsed">
        <button className="sidebar-toggle" onClick={onToggle} aria-label="Expand sidebar">
          {/* Chevron right icon */}
        </button>
      </div>
    );
  }

  return (
    <aside className="sidebar">
      <div className="sidebar-header">
        <h1 className="sidebar-title">EgenSkriven</h1>
        <button className="sidebar-toggle" onClick={onToggle} aria-label="Collapse sidebar">
          {/* Chevron left icon */}
        </button>
      </div>

      <nav className="sidebar-nav">
        {/* Boards Section */}
        <section className="sidebar-section">
          <h2 className="sidebar-section-title">BOARDS</h2>
          
          {loading ? (
            <div className="sidebar-loading">Loading...</div>
          ) : (
            <ul className="board-list">
              {boards.map((board) => (
                <li key={board.id}>
                  <button
                    className={`board-item ${currentBoard?.id === board.id ? 'active' : ''}`}
                    onClick={() => setCurrentBoard(board)}
                    style={{ '--board-color': board.color || 'var(--accent)' } as React.CSSProperties}
                  >
                    <span 
                      className="board-indicator"
                      style={{ backgroundColor: board.color || 'var(--accent)' }}
                    />
                    <span className="board-name">{board.name}</span>
                    <span className="board-prefix">({board.prefix})</span>
                  </button>
                </li>
              ))}
            </ul>
          )}

          <button className="new-board-btn" onClick={() => setShowNewBoard(true)}>
            + New board
          </button>
        </section>

        {/* Views Section (placeholder for Phase 6) */}
        <section className="sidebar-section">
          <h2 className="sidebar-section-title">VIEWS</h2>
          <button className="new-view-btn" disabled>
            + New view
          </button>
        </section>
      </nav>

      {showNewBoard && (
        <NewBoardModal onClose={() => setShowNewBoard(false)} />
      )}
    </aside>
  );
}

// New Board Modal Component
function NewBoardModal({ onClose }: { onClose: () => void }) {
  const [name, setName] = useState('');
  const [prefix, setPrefix] = useState('');
  const [color, setColor] = useState('#5E6AD2');
  const [error, setError] = useState('');
  const { createBoard } = useBoards();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!name.trim()) {
      setError('Name is required');
      return;
    }
    if (!prefix.trim()) {
      setError('Prefix is required');
      return;
    }

    try {
      await createBoard({ name, prefix: prefix.toUpperCase(), color });
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create board');
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <h2>Create New Board</h2>
        <form onSubmit={handleSubmit}>
          <div className="form-field">
            <label htmlFor="board-name">Name</label>
            <input
              id="board-name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g., Work, Personal"
              autoFocus
            />
          </div>

          <div className="form-field">
            <label htmlFor="board-prefix">Prefix</label>
            <input
              id="board-prefix"
              type="text"
              value={prefix}
              onChange={(e) => setPrefix(e.target.value.toUpperCase())}
              placeholder="e.g., WRK, PER"
              maxLength={10}
            />
            <span className="form-hint">Used in task IDs (e.g., {prefix || 'WRK'}-123)</span>
          </div>

          <div className="form-field">
            <label>Color</label>
            <div className="color-picker">
              {['#5E6AD2', '#9333EA', '#22C55E', '#F97316', '#EC4899', '#06B6D4', '#EF4444', '#EAB308'].map((c) => (
                <button
                  key={c}
                  type="button"
                  className={`color-option ${color === c ? 'selected' : ''}`}
                  style={{ backgroundColor: c }}
                  onClick={() => setColor(c)}
                />
              ))}
            </div>
          </div>

          {error && <div className="form-error">{error}</div>}

          <div className="modal-actions">
            <button type="button" className="btn-secondary" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="btn-primary">
              Create Board
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
```

**Steps**:

1. Create the file:
   ```bash
   touch ui/src/components/Sidebar.tsx
   ```

2. Open in your editor and paste the code above.

3. Create the CSS file `ui/src/components/Sidebar.css`:
   ```css
   .sidebar {
     width: 240px;
     min-width: 240px;
     height: 100vh;
     background: var(--bg-sidebar);
     border-right: 1px solid var(--border-subtle);
     display: flex;
     flex-direction: column;
   }

   .sidebar.collapsed {
     width: 48px;
     min-width: 48px;
   }

   .sidebar-header {
     padding: var(--space-4);
     display: flex;
     justify-content: space-between;
     align-items: center;
   }

   .sidebar-title {
     font-size: var(--text-lg);
     font-weight: var(--font-semibold);
     color: var(--text-primary);
   }

   .sidebar-section {
     padding: var(--space-3) var(--space-4);
   }

   .sidebar-section-title {
     font-size: var(--text-xs);
     font-weight: var(--font-semibold);
     color: var(--text-muted);
     text-transform: uppercase;
     letter-spacing: 0.05em;
     margin-bottom: var(--space-2);
   }

   .board-list {
     list-style: none;
     padding: 0;
     margin: 0;
   }

   .board-item {
     width: 100%;
     display: flex;
     align-items: center;
     gap: var(--space-2);
     padding: var(--space-2);
     border: none;
     background: transparent;
     color: var(--text-secondary);
     font-size: var(--text-base);
     text-align: left;
     cursor: pointer;
     border-radius: var(--radius-md);
   }

   .board-item:hover {
     background: var(--bg-card-hover);
     color: var(--text-primary);
   }

   .board-item.active {
     background: var(--bg-card-selected);
     color: var(--text-primary);
   }

   .board-indicator {
     width: 8px;
     height: 8px;
     border-radius: 50%;
   }

   .board-prefix {
     color: var(--text-muted);
     font-size: var(--text-sm);
   }

   .new-board-btn,
   .new-view-btn {
     width: 100%;
     padding: var(--space-2);
     border: none;
     background: transparent;
     color: var(--text-muted);
     font-size: var(--text-sm);
     text-align: left;
     cursor: pointer;
   }

   .new-board-btn:hover {
     color: var(--text-primary);
   }
   ```

4. Verify it compiles:
   ```bash
   cd ui && npm run build
   ```

---

### 5.12 Create useBoards Hook

**What**: Create a React hook for board data and operations.

**Why**: Centralizes board state management and provides CRUD operations.

**File**: Update `ui/src/hooks/usePocketBase.ts`

Add these exports:

```typescript
// Board types
export interface Board {
  id: string;
  name: string;
  prefix: string;
  columns: string[];
  color?: string;
}

// useBoards hook
export function useBoards() {
  const [boards, setBoards] = useState<Board[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Fetch all boards
    pb.collection('boards')
      .getFullList<Board>({ sort: 'name' })
      .then(setBoards)
      .finally(() => setLoading(false));

    // Subscribe to board changes
    pb.collection('boards').subscribe<Board>('*', (e) => {
      if (e.action === 'create') {
        setBoards((prev) => [...prev, e.record]);
      } else if (e.action === 'update') {
        setBoards((prev) => prev.map((b) => (b.id === e.record.id ? e.record : b)));
      } else if (e.action === 'delete') {
        setBoards((prev) => prev.filter((b) => b.id !== e.record.id));
      }
    });

    return () => {
      pb.collection('boards').unsubscribe('*');
    };
  }, []);

  const createBoard = async (input: { name: string; prefix: string; color?: string }) => {
    const board = await pb.collection('boards').create({
      name: input.name,
      prefix: input.prefix.toUpperCase(),
      columns: ['backlog', 'todo', 'in_progress', 'review', 'done'],
      color: input.color,
    });
    return board;
  };

  const deleteBoard = async (id: string) => {
    await pb.collection('boards').delete(id);
  };

  return { boards, loading, createBoard, deleteBoard };
}

// useCurrentBoard hook - manages the active board
export function useCurrentBoard() {
  const [currentBoard, setCurrentBoardState] = useState<Board | null>(null);
  const { boards } = useBoards();

  useEffect(() => {
    // Initialize from localStorage or use first board
    const savedBoardId = localStorage.getItem('currentBoard');
    if (savedBoardId) {
      const board = boards.find((b) => b.id === savedBoardId);
      if (board) {
        setCurrentBoardState(board);
        return;
      }
    }
    // Default to first board
    if (boards.length > 0 && !currentBoard) {
      setCurrentBoardState(boards[0]);
    }
  }, [boards]);

  const setCurrentBoard = (board: Board) => {
    setCurrentBoardState(board);
    localStorage.setItem('currentBoard', board.id);
  };

  return { currentBoard, setCurrentBoard };
}
```

**Steps**:

1. Open `ui/src/hooks/usePocketBase.ts`.

2. Add the Board interface and hooks.

3. Verify it compiles:
   ```bash
   cd ui && npm run build
   ```

---

### 5.13 Update Board Component for Multi-Board

**What**: Modify the Board component to filter tasks by current board.

**Why**: Each board should only show its own tasks.

**File**: Update `ui/src/components/Board.tsx`

Key changes:

```tsx
import { useCurrentBoard } from '../hooks/usePocketBase';

export function Board() {
  const { currentBoard } = useCurrentBoard();
  const { tasks, loading, moveTask } = useTasks(currentBoard?.id);

  // Filter tasks by current board
  const tasksByColumn = useMemo(() => {
    const grouped: Record<string, Task[]> = {};
    
    // Use board's columns or defaults
    const columns = currentBoard?.columns || ['backlog', 'todo', 'in_progress', 'review', 'done'];
    columns.forEach((col) => (grouped[col] = []));

    tasks.forEach((task) => {
      if (grouped[task.column]) {
        grouped[task.column].push(task);
      }
    });

    // Sort by position within each column
    Object.keys(grouped).forEach((col) => {
      grouped[col].sort((a, b) => a.position - b.position);
    });

    return grouped;
  }, [tasks, currentBoard]);

  if (!currentBoard) {
    return <div className="board-empty">No board selected</div>;
  }

  // ... rest of component
}
```

**Steps**:

1. Open `ui/src/components/Board.tsx`.

2. Import and use `useCurrentBoard`.

3. Update `useTasks` to accept a board ID filter.

4. Use the board's custom columns if available.

5. Verify it compiles:
   ```bash
   cd ui && npm run build
   ```

---

### 5.14 Update TaskCard for Display IDs

**What**: Show board-prefixed display IDs on task cards.

**Why**: Tasks should display `WRK-123` instead of internal IDs.

**File**: Update `ui/src/components/TaskCard.tsx`

```tsx
import { useCurrentBoard } from '../hooks/usePocketBase';

export function TaskCard({ task }: { task: Task }) {
  const { currentBoard } = useCurrentBoard();

  // Format display ID
  const displayId = currentBoard
    ? `${currentBoard.prefix}-${task.seq}`
    : task.id.substring(0, 8);

  return (
    <div className="task-card">
      <div className="task-header">
        <span className="task-status-dot" style={{ backgroundColor: getStatusColor(task.column) }} />
        <span className="task-id">{displayId}</span>
      </div>
      <h3 className="task-title">{task.title}</h3>
      {/* ... rest of card */}
    </div>
  );
}
```

**Steps**:

1. Open `ui/src/components/TaskCard.tsx`.

2. Import `useCurrentBoard`.

3. Format the display ID from board prefix and task sequence.

4. Verify it compiles:
   ```bash
   cd ui && npm run build
   ```

---

### 5.15 Add Board Switcher to Command Palette

**What**: Add board switching capability to the command palette.

**Why**: Power users should be able to switch boards via keyboard.

**File**: Update `ui/src/components/CommandPalette.tsx`

Add board-related commands:

```tsx
// Add board commands to the command list
const boardCommands = boards.map((board) => ({
  id: `board-${board.id}`,
  title: `Go to board: ${board.name}`,
  shortcut: 'G B',
  section: 'navigation',
  action: () => {
    setCurrentBoard(board);
    onClose();
  },
}));

// Merge with existing commands
const allCommands = [...actionCommands, ...boardCommands, ...recentTaskCommands];
```

**Steps**:

1. Open `ui/src/components/CommandPalette.tsx`.

2. Import `useBoards` and `useCurrentBoard`.

3. Add board navigation commands.

4. Verify it compiles:
   ```bash
   cd ui && npm run build
   ```

---

### 5.16 Register Board Commands in CLI

**What**: Add the board command to the root command.

**Why**: Makes `egenskriven board` available in the CLI.

**File**: Update `internal/commands/root.go`

```go
func Register(app *pocketbase.PocketBase) {
	out := &output.Formatter{}

	rootCmd := &cobra.Command{...}

	// ... existing commands

	// Add board command
	rootCmd.AddCommand(NewBoardCmd(app, out))

	app.RootCmd.AddCommand(rootCmd)
}
```

**Steps**:

1. Open `internal/commands/root.go`.

2. Import the board command.

3. Add it to the root command.

4. Verify it compiles:
   ```bash
   go build ./cmd/egenskriven
   ```

---

## Verification Checklist

Complete each section in order. Check off each item as you verify it.

### CLI Verification

- [ ] **Board creation works**
  ```bash
  ./egenskriven board add "Work" --prefix WRK
  ```
  Should output: `Created board: Work (WRK)`

- [ ] **Board list shows boards**
  ```bash
  ./egenskriven board list
  ```
  Should show the created board.

- [ ] **Duplicate prefix rejected**
  ```bash
  ./egenskriven board add "Work 2" --prefix WRK
  ```
  Should show error about prefix being in use.

- [ ] **Board use sets default**
  ```bash
  ./egenskriven board use Work
  cat .egenskriven/config.json
  ```
  Should show `"default_board": "WRK"`.

- [ ] **Task creation uses board**
  ```bash
  ./egenskriven add "Test task"
  ```
  Should output: `Created task: Test task [WRK-1]`

- [ ] **Task list shows display IDs**
  ```bash
  ./egenskriven list
  ```
  Should show `[WRK-1] Test task` format.

- [ ] **Task resolution by display ID works**
  ```bash
  ./egenskriven show WRK-1
  ```
  Should show task details.

- [ ] **All-boards flag works**
  ```bash
  ./egenskriven board add "Personal" --prefix PER
  ./egenskriven add "Personal task" --board PER
  ./egenskriven list --all-boards
  ```
  Should show tasks from both boards.

- [ ] **Board deletion works**
  ```bash
  ./egenskriven board delete Personal --force
  ```
  Should delete board and its tasks.

### UI Verification

- [ ] **Sidebar shows boards**
  
  Open `http://localhost:8090` and verify sidebar lists boards.

- [ ] **Board switching works**
  
  Click a different board in sidebar; task list updates.

- [ ] **New board modal works**
  
  Click "+ New board", fill form, verify board appears.

- [ ] **Task cards show display IDs**
  
  Verify tasks show `WRK-123` format, not internal IDs.

- [ ] **Command palette has board commands**
  
  Press `Cmd+K`, type "board", verify board navigation options.

- [ ] **Board color applies**
  
  Create board with color, verify indicator shows that color.

### Test Verification

- [ ] **All tests pass**
  ```bash
  make test
  ```
  Should pass all tests including new board tests.

- [ ] **Board tests specifically pass**
  ```bash
  go test ./internal/board -v
  ```
  Should pass all board service tests.

---

## File Summary

| File | Lines | Purpose |
|------|-------|---------|
| `migrations/2_boards.go` | ~50 | Creates boards collection |
| `migrations/3_tasks_board_relation.go` | ~50 | Adds board relation to tasks |
| `migrations/4_tasks_sequence.go` | ~40 | Adds sequence field to tasks |
| `internal/board/board.go` | ~200 | Board service layer |
| `internal/board/board_test.go` | ~150 | Board service tests |
| `internal/commands/board.go` | ~200 | Board CLI commands |
| `internal/config/config.go` (update) | ~20 | Add default board config |
| `internal/commands/add.go` (update) | ~50 | Board support in add |
| `internal/commands/list.go` (update) | ~30 | Board filtering in list |
| `internal/resolver/resolver.go` (update) | ~30 | Display ID resolution |
| `ui/src/components/Sidebar.tsx` | ~150 | Board sidebar UI |
| `ui/src/components/Sidebar.css` | ~100 | Sidebar styling |
| `ui/src/hooks/usePocketBase.ts` (update) | ~80 | Board hooks |
| `ui/src/components/Board.tsx` (update) | ~20 | Board filtering |
| `ui/src/components/TaskCard.tsx` (update) | ~10 | Display IDs |
| `ui/src/components/CommandPalette.tsx` (update) | ~20 | Board commands |

**Total new/modified code**: ~1,200 lines

---

## What You Should Have Now

After completing Phase 5, your application should:

1. **Support multiple boards** with unique prefixes
2. **Display human-friendly task IDs** like `WRK-123`
3. **Allow CLI board management**: create, list, use, delete
4. **Show board switcher in UI sidebar**
5. **Filter tasks by board** in both CLI and UI
6. **Store default board preference** in config
7. **Resolve tasks by display ID** in CLI commands

---

## Next Phase

**Phase 6: Filtering & Views** will add:
- Advanced filter UI with multiple conditions
- Search functionality
- Saved views with persisted filters
- List view alternative to board view
- Display options per view

---

## Troubleshooting

### "boards collection not found"

**Problem**: Migrations haven't run.

**Solution**:
```bash
./egenskriven migrate
```

### "prefix already in use"

**Problem**: Trying to create a board with a duplicate prefix.

**Solution**: Choose a different prefix or delete the existing board.

### Tasks don't show display IDs

**Problem**: Tasks created before multi-board support don't have `seq` field.

**Solution**: Run a migration to backfill sequence numbers for existing tasks:
```bash
# Manual SQL in PocketBase admin or create a migration
UPDATE tasks SET seq = rowid WHERE seq IS NULL;
```

### Board not found when using display ID

**Problem**: Display ID parsing expects exact `PREFIX-NUMBER` format.

**Solution**: Ensure you're using uppercase prefix: `WRK-123` not `wrk-123`.

### UI shows wrong tasks after switching boards

**Problem**: Tasks hook not filtering by board ID.

**Solution**: Verify `useTasks` hook accepts and uses board ID parameter.

### Sidebar not collapsible

**Problem**: Missing keyboard shortcut handler for `Cmd+\`.

**Solution**: Add to `useKeyboard` hook in Phase 4.

---

## Glossary

| Term | Definition |
|------|------------|
| **Board** | A container for tasks with its own prefix, columns, and color |
| **Prefix** | Short uppercase identifier used in display IDs (e.g., "WRK") |
| **Display ID** | Human-readable task identifier combining prefix and sequence (e.g., "WRK-123") |
| **Sequence** | Per-board incrementing number for task IDs |
| **Internal ID** | Auto-generated PocketBase ID used for storage |
| **Default Board** | Board used when no `--board` flag is specified |
