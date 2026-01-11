# Phase 1: Core CLI

**Goal**: Fully functional CLI for basic task management with no UI. Tasks can be created, listed, viewed, moved, updated, and deleted.

**Duration Estimate**: 3-5 days

**Prerequisites**: Phase 0 complete (working build system, testing infrastructure).

**Deliverable**: A CLI application that manages tasks stored in PocketBase's SQLite database, with both human-readable and JSON output.

---

## Overview

In this phase, we build the core CLI commands that form the foundation of EgenSkriven. The CLI is designed to be:
- **Human-friendly** by default (readable output, helpful errors)
- **Machine-friendly** with `--json` flag (structured output for agents/scripts)
- **Flexible** in how tasks are referenced (by ID, ID prefix, or title)

By the end of this phase, you'll have a fully functional task manager accessible via command line.

### What We're Building

| Command | Purpose |
|---------|---------|
| `add` | Create new tasks |
| `list` | List and filter tasks |
| `show` | Show detailed task information |
| `move` | Move tasks between columns |
| `update` | Update task properties |
| `delete` | Delete tasks |

### Architecture Overview

```
User Input → Cobra Command → Task Resolver → PocketBase → Output Formatter → User
                  ↓                                              ↓
             Validation                                    JSON or Human
```

---

## Environment Requirements

Ensure Phase 0 is complete:

| Check | Command |
|-------|---------|
| Build works | `make build` |
| Tests run | `make test` |
| Server starts | `./egenskriven serve` |

---

## Tasks

### 1.1 Create Database Migration

**What**: Define the `tasks` collection schema using PocketBase migrations.

**Why**: PocketBase uses migrations to create and modify database collections. This ensures consistent schema across all installations.

**File**: `migrations/1_initial.go`

```go
package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection := core.NewBaseCollection("tasks")

		// Title - required, the main task identifier
		collection.Fields.Add(&core.TextField{
			Name:     "title",
			Required: true,
			Max:      500,
		})

		// Description - optional longer text
		collection.Fields.Add(&core.TextField{
			Name: "description",
			Max:  10000,
		})

		// Type - categorizes the task
		collection.Fields.Add(&core.SelectField{
			Name:     "type",
			Required: true,
			Values:   []string{"bug", "feature", "chore"},
		})

		// Priority - importance level
		collection.Fields.Add(&core.SelectField{
			Name:     "priority",
			Required: true,
			Values:   []string{"low", "medium", "high", "urgent"},
		})

		// Column - kanban board column (status)
		collection.Fields.Add(&core.SelectField{
			Name:     "column",
			Required: true,
			Values:   []string{"backlog", "todo", "in_progress", "review", "done"},
		})

		// Position - order within column (fractional for easy reordering)
		collection.Fields.Add(&core.NumberField{
			Name:     "position",
			Required: true,
			Min:      floatPtr(0),
		})

		// Labels - array of string tags
		collection.Fields.Add(&core.JSONField{
			Name:    "labels",
			MaxSize: 10000,
		})

		// Blocked by - array of task IDs that block this task
		collection.Fields.Add(&core.JSONField{
			Name:    "blocked_by",
			MaxSize: 10000,
		})

		// Created by - who created this task
		collection.Fields.Add(&core.SelectField{
			Name:     "created_by",
			Required: true,
			Values:   []string{"user", "agent", "cli"},
		})

		// Created by agent - optional agent identifier
		collection.Fields.Add(&core.TextField{
			Name: "created_by_agent",
			Max:  100,
		})

		// History - activity log as JSON array
		collection.Fields.Add(&core.JSONField{
			Name:    "history",
			MaxSize: 100000,
		})

		return app.Save(collection)
	}, func(app core.App) error {
		// Rollback: delete the collection
		collection, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}
		return app.Delete(collection)
	})
}

// Helper function for pointer to float
func floatPtr(f float64) *float64 {
	return &f
}
```

**Steps**:

1. Create the migrations directory (if not exists):
   ```bash
   mkdir -p migrations
   ```

2. Create the migration file:
   ```bash
   touch migrations/1_initial.go
   ```

3. Open in your editor and paste the code above.

4. Verify it compiles:
   ```bash
   go build ./migrations
   ```
   
   **Expected output**: No output means success!

5. Test the migration runs by starting the server:
   ```bash
   ./egenskriven serve
   ```
   
   Then visit `http://localhost:8090/_/` and check that the `tasks` collection exists.

**Schema Reference**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| id | string | auto | PocketBase auto-generated ID |
| title | string | yes | Task title (max 500 chars) |
| description | string | no | Longer description (max 10000 chars) |
| type | select | yes | bug, feature, chore |
| priority | select | yes | low, medium, high, urgent |
| column | select | yes | backlog, todo, in_progress, review, done |
| position | number | yes | Order within column (fractional) |
| labels | json | no | Array of label strings |
| blocked_by | json | no | Array of blocking task IDs |
| created_by | select | yes | user, agent, cli |
| created_by_agent | string | no | Agent identifier (e.g., "claude") |
| history | json | no | Activity log array |
| created | date | auto | Auto-generated timestamp |
| updated | date | auto | Auto-generated timestamp |

**Common Mistakes**:
- Forgetting to register the migration in `init()`
- Using wrong field types (e.g., `core.TextField` vs `core.SelectField`)
- Not handling the rollback function

---

### 1.2 Implement Output Formatter

**What**: Create a unified output formatter that handles both JSON and human-readable output.

**Why**: All CLI commands need consistent output formatting. The `--json` flag enables machine-readable output for scripts and AI agents.

**File**: `internal/output/output.go`

```go
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// Formatter handles output formatting for CLI commands.
// It supports both human-readable and JSON output modes.
type Formatter struct {
	// JSON enables JSON output mode
	JSON bool
	// Quiet suppresses non-essential output
	Quiet bool
}

// New creates a new Formatter with the given options.
func New(jsonMode, quiet bool) *Formatter {
	return &Formatter{
		JSON:  jsonMode,
		Quiet: quiet,
	}
}

// Task outputs a single task.
// Human mode: "Created task: Title [id]"
// JSON mode: Full task object
func (f *Formatter) Task(task *core.Record, action string) {
	if f.Quiet && !f.JSON {
		return
	}

	if f.JSON {
		f.writeJSON(taskToMap(task))
		return
	}

	fmt.Printf("%s task: %s [%s]\n", action, task.GetString("title"), shortID(task.Id))
}

// Tasks outputs a list of tasks.
// Human mode: Grouped by column
// JSON mode: Array with count
func (f *Formatter) Tasks(tasks []*core.Record) {
	if f.JSON {
		f.writeJSON(map[string]any{
			"tasks": tasksToMaps(tasks),
			"count": len(tasks),
		})
		return
	}

	// Group tasks by column
	grouped := groupByColumn(tasks)
	columns := []string{"backlog", "todo", "in_progress", "review", "done"}

	for _, col := range columns {
		colTasks := grouped[col]
		fmt.Printf("\n%s\n", strings.ToUpper(col))

		if len(colTasks) == 0 {
			fmt.Println("  (empty)")
			continue
		}

		for _, task := range colTasks {
			f.printTaskLine(task)
		}
	}
	fmt.Println()
}

// TaskDetail outputs detailed information about a task.
func (f *Formatter) TaskDetail(task *core.Record) {
	if f.JSON {
		f.writeJSON(taskToMap(task))
		return
	}

	fmt.Printf("\nTask: %s\n", task.Id)
	fmt.Printf("Title:       %s\n", task.GetString("title"))
	fmt.Printf("Type:        %s\n", task.GetString("type"))
	fmt.Printf("Priority:    %s\n", task.GetString("priority"))
	fmt.Printf("Column:      %s\n", task.GetString("column"))
	fmt.Printf("Position:    %.0f\n", task.GetFloat("position"))

	// Labels
	labels := getLabels(task)
	if len(labels) > 0 {
		fmt.Printf("Labels:      %s\n", strings.Join(labels, ", "))
	} else {
		fmt.Printf("Labels:      -\n")
	}

	// Blocked by
	blockedBy := getBlockedBy(task)
	if len(blockedBy) > 0 {
		fmt.Printf("Blocked by:  %s\n", strings.Join(blockedBy, ", "))
	}

	// Created by
	createdBy := task.GetString("created_by")
	if agent := task.GetString("created_by_agent"); agent != "" {
		fmt.Printf("Created by:  %s (%s)\n", createdBy, agent)
	} else {
		fmt.Printf("Created by:  %s\n", createdBy)
	}

	// Timestamps
	fmt.Printf("Created:     %s\n", formatTime(task.GetDateTime("created").Time()))
	fmt.Printf("Updated:     %s\n", formatTime(task.GetDateTime("updated").Time()))

	// Description
	if desc := task.GetString("description"); desc != "" {
		fmt.Printf("\nDescription:\n  %s\n", strings.ReplaceAll(desc, "\n", "\n  "))
	}

	fmt.Println()
}

// Success outputs a success message.
func (f *Formatter) Success(message string) {
	if f.Quiet {
		return
	}
	if f.JSON {
		f.writeJSON(map[string]any{
			"success": true,
			"message": message,
		})
		return
	}
	fmt.Println(message)
}

// Error outputs an error with optional data.
// Returns the error for convenient chaining.
func (f *Formatter) Error(code int, message string, data any) error {
	if f.JSON {
		errObj := map[string]any{
			"error": map[string]any{
				"code":    code,
				"message": message,
			},
		}
		if data != nil {
			errObj["error"].(map[string]any)["data"] = data
		}
		json.NewEncoder(os.Stderr).Encode(errObj)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", message)
	}
	return fmt.Errorf(message)
}

// AmbiguousError outputs an error for ambiguous task references.
func (f *Formatter) AmbiguousError(ref string, matches []*core.Record) error {
	data := map[string]any{
		"reference": ref,
		"matches":   tasksToMaps(matches),
	}

	if f.JSON {
		return f.Error(4, fmt.Sprintf("Ambiguous task reference: '%s' matches multiple tasks", ref), data)
	}

	fmt.Fprintf(os.Stderr, "Error: Ambiguous task reference: '%s' matches multiple tasks:\n", ref)
	for _, task := range matches {
		fmt.Fprintf(os.Stderr, "  [%s] %s\n", shortID(task.Id), task.GetString("title"))
	}
	return fmt.Errorf("ambiguous task reference")
}

// --- Helper functions ---

func (f *Formatter) writeJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func (f *Formatter) printTaskLine(task *core.Record) {
	priority := task.GetString("priority")
	priorityIndicator := ""
	switch priority {
	case "urgent":
		priorityIndicator = "!!!"
	case "high":
		priorityIndicator = "!!"
	case "medium":
		priorityIndicator = "!"
	}

	fmt.Printf("  [%s] %s (%s%s)\n",
		shortID(task.Id),
		task.GetString("title"),
		task.GetString("type"),
		func() string {
			if priorityIndicator != "" {
				return ", " + priorityIndicator + priority
			}
			return ""
		}(),
	)
}

func taskToMap(task *core.Record) map[string]any {
	return map[string]any{
		"id":               task.Id,
		"title":            task.GetString("title"),
		"description":      task.GetString("description"),
		"type":             task.GetString("type"),
		"priority":         task.GetString("priority"),
		"column":           task.GetString("column"),
		"position":         task.GetFloat("position"),
		"labels":           getLabels(task),
		"blocked_by":       getBlockedBy(task),
		"created_by":       task.GetString("created_by"),
		"created_by_agent": task.GetString("created_by_agent"),
		"created":          task.GetDateTime("created").Time().Format(time.RFC3339),
		"updated":          task.GetDateTime("updated").Time().Format(time.RFC3339),
	}
}

func tasksToMaps(tasks []*core.Record) []map[string]any {
	result := make([]map[string]any, len(tasks))
	for i, task := range tasks {
		result[i] = taskToMap(task)
	}
	return result
}

func groupByColumn(tasks []*core.Record) map[string][]*core.Record {
	grouped := make(map[string][]*core.Record)
	for _, task := range tasks {
		col := task.GetString("column")
		grouped[col] = append(grouped[col], task)
	}
	return grouped
}

func getLabels(task *core.Record) []string {
	raw := task.Get("labels")
	if raw == nil {
		return []string{}
	}
	if labels, ok := raw.([]any); ok {
		result := make([]string, len(labels))
		for i, l := range labels {
			result[i] = fmt.Sprint(l)
		}
		return result
	}
	return []string{}
}

func getBlockedBy(task *core.Record) []string {
	raw := task.Get("blocked_by")
	if raw == nil {
		return []string{}
	}
	if ids, ok := raw.([]any); ok {
		result := make([]string, len(ids))
		for i, id := range ids {
			result[i] = fmt.Sprint(id)
		}
		return result
	}
	return []string{}
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
```

**Steps**:

1. Create the file:
   ```bash
   touch internal/output/output.go
   ```

2. Open in your editor and paste the code above.

3. Verify it compiles:
   ```bash
   go build ./internal/output
   ```
   
   **Expected output**: No output means success!

**Key Concepts Explained**:

| Concept | Explanation |
|---------|-------------|
| `*Formatter` | Struct that holds output preferences (JSON mode, quiet mode) |
| `taskToMap()` | Converts PocketBase record to a plain map for JSON serialization |
| `shortID()` | Shows first 8 characters of ID for human readability |
| Exit codes | Standard codes: 0=success, 1=error, 3=not found, 4=ambiguous |

---

### 1.3 Implement Task Resolver

**What**: Create a resolver that finds tasks by ID, ID prefix, or title match.

**Why**: Users shouldn't need to type full 15-character IDs. The resolver allows flexible task references like `abc123` (ID prefix) or `"fix login"` (title search).

**File**: `internal/resolver/resolver.go`

```go
package resolver

import (
	"fmt"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// Resolution represents the result of resolving a task reference.
type Resolution struct {
	// Task is the resolved task (nil if not found or ambiguous)
	Task *core.Record
	// Matches contains all matching tasks (populated if ambiguous)
	Matches []*core.Record
}

// IsAmbiguous returns true if the resolution matched multiple tasks.
func (r *Resolution) IsAmbiguous() bool {
	return len(r.Matches) > 1
}

// IsNotFound returns true if no tasks matched.
func (r *Resolution) IsNotFound() bool {
	return r.Task == nil && len(r.Matches) == 0
}

// ResolveTask attempts to find a task by reference.
// Resolution order:
// 1. Exact ID match
// 2. ID prefix match (must be unique)
// 3. Title substring match (case-insensitive, must be unique)
func ResolveTask(app *pocketbase.PocketBase, ref string) (*Resolution, error) {
	// 1. Try exact ID match
	task, err := app.FindRecordById("tasks", ref)
	if err == nil {
		return &Resolution{Task: task}, nil
	}

	// 2. Try ID prefix match
	tasks, err := app.FindAllRecords("tasks",
		dbx.NewExp("id LIKE {:prefix}", dbx.Params{"prefix": ref + "%"}),
	)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if len(tasks) == 1 {
		return &Resolution{Task: tasks[0]}, nil
	}

	if len(tasks) > 1 {
		return &Resolution{Matches: tasks}, nil
	}

	// 3. Try title match (case-insensitive substring)
	tasks, err = app.FindAllRecords("tasks",
		dbx.NewExp("LOWER(title) LIKE {:title}",
			dbx.Params{"title": "%" + strings.ToLower(ref) + "%"}),
	)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	switch len(tasks) {
	case 0:
		return &Resolution{}, nil // not found
	case 1:
		return &Resolution{Task: tasks[0]}, nil
	default:
		return &Resolution{Matches: tasks}, nil // ambiguous
	}
}

// MustResolve resolves a task and returns an error if not found or ambiguous.
// This is a convenience wrapper for commands that need exactly one task.
func MustResolve(app *pocketbase.PocketBase, ref string) (*core.Record, error) {
	resolution, err := ResolveTask(app, ref)
	if err != nil {
		return nil, err
	}

	if resolution.IsNotFound() {
		return nil, fmt.Errorf("no task found matching: %s", ref)
	}

	if resolution.IsAmbiguous() {
		return nil, &AmbiguousError{
			Reference: ref,
			Matches:   resolution.Matches,
		}
	}

	return resolution.Task, nil
}

// AmbiguousError is returned when a reference matches multiple tasks.
type AmbiguousError struct {
	Reference string
	Matches   []*core.Record
}

func (e *AmbiguousError) Error() string {
	return fmt.Sprintf("ambiguous task reference: '%s' matches %d tasks", e.Reference, len(e.Matches))
}
```

**Steps**:

1. Create the file:
   ```bash
   touch internal/resolver/resolver.go
   ```

2. Open in your editor and paste the code above.

3. Verify it compiles:
   ```bash
   go build ./internal/resolver
   ```
   
   **Expected output**: No output means success!

**Resolution Examples**:

| Input | Resolution |
|-------|------------|
| `abc123def456` | Exact ID match |
| `abc123` | ID prefix match (if unique) |
| `"fix login"` | Title substring match (if unique) |
| `"task"` | Ambiguous (returns all matching tasks) |

**Common Mistakes**:
- Not handling the case where multiple tasks match by prefix
- Forgetting case-insensitive comparison for title search
- Not properly escaping SQL LIKE patterns

---

### 1.4 Implement Position Calculator

**What**: Create helper functions for calculating task positions within columns.

**Why**: When tasks are moved or created, we need to calculate their position in the column. Using fractional indexing avoids rebalancing all positions on every move.

**File**: `internal/commands/position.go`

```go
package commands

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

const (
	// DefaultPositionGap is the space between positions for new tasks
	DefaultPositionGap = 1000.0
	// MinPositionGap is the minimum gap before rebalancing should occur
	MinPositionGap = 0.001
)

// GetNextPosition returns the position for a new task at the end of a column.
// If the column is empty, returns DefaultPositionGap.
// Otherwise, returns the last position + DefaultPositionGap.
func GetNextPosition(app *pocketbase.PocketBase, column string) float64 {
	tasks, err := app.FindAllRecords("tasks",
		dbx.NewExp("column = {:col}", dbx.Params{"col": column}),
	)
	if err != nil || len(tasks) == 0 {
		return DefaultPositionGap
	}

	// Find the maximum position
	var maxPos float64
	for _, task := range tasks {
		pos := task.GetFloat("position")
		if pos > maxPos {
			maxPos = pos
		}
	}

	return maxPos + DefaultPositionGap
}

// GetPositionBetween calculates a position between two existing positions.
// This enables inserting tasks between existing tasks without rebalancing.
func GetPositionBetween(before, after float64) float64 {
	return (before + after) / 2.0
}

// GetPositionAtIndex returns the position for a task at a specific index in a column.
// index 0 = top of column
// index -1 = bottom of column (same as GetNextPosition)
func GetPositionAtIndex(app *pocketbase.PocketBase, column string, index int) float64 {
	tasks, err := app.FindAllRecords("tasks",
		dbx.NewExp("column = {:col}", dbx.Params{"col": column}),
	)
	if err != nil || len(tasks) == 0 {
		return DefaultPositionGap
	}

	// Sort tasks by position
	sortTasksByPosition(tasks)

	// Handle bottom of column
	if index < 0 || index >= len(tasks) {
		return tasks[len(tasks)-1].GetFloat("position") + DefaultPositionGap
	}

	// Handle top of column
	if index == 0 {
		return tasks[0].GetFloat("position") / 2.0
	}

	// Insert between two tasks
	before := tasks[index-1].GetFloat("position")
	after := tasks[index].GetFloat("position")
	return GetPositionBetween(before, after)
}

// GetPositionAfter returns a position after a specific task.
func GetPositionAfter(app *pocketbase.PocketBase, taskID string) (float64, error) {
	task, err := app.FindRecordById("tasks", taskID)
	if err != nil {
		return 0, err
	}

	column := task.GetString("column")
	targetPos := task.GetFloat("position")

	// Find the next task in the column
	tasks, err := app.FindAllRecords("tasks",
		dbx.NewExp("column = {:col} AND position > {:pos}",
			dbx.Params{"col": column, "pos": targetPos}),
	)
	if err != nil {
		return 0, err
	}

	if len(tasks) == 0 {
		// No task after, append to end
		return targetPos + DefaultPositionGap, nil
	}

	// Find minimum position among tasks after
	minPos := tasks[0].GetFloat("position")
	for _, t := range tasks {
		if pos := t.GetFloat("position"); pos < minPos {
			minPos = pos
		}
	}

	return GetPositionBetween(targetPos, minPos), nil
}

// GetPositionBefore returns a position before a specific task.
func GetPositionBefore(app *pocketbase.PocketBase, taskID string) (float64, error) {
	task, err := app.FindRecordById("tasks", taskID)
	if err != nil {
		return 0, err
	}

	column := task.GetString("column")
	targetPos := task.GetFloat("position")

	// Find the previous task in the column
	tasks, err := app.FindAllRecords("tasks",
		dbx.NewExp("column = {:col} AND position < {:pos}",
			dbx.Params{"col": column, "pos": targetPos}),
	)
	if err != nil {
		return 0, err
	}

	if len(tasks) == 0 {
		// No task before, put at top
		return targetPos / 2.0, nil
	}

	// Find maximum position among tasks before
	maxPos := tasks[0].GetFloat("position")
	for _, t := range tasks {
		if pos := t.GetFloat("position"); pos > maxPos {
			maxPos = pos
		}
	}

	return GetPositionBetween(maxPos, targetPos), nil
}

// sortTasksByPosition sorts tasks by their position field in ascending order.
func sortTasksByPosition(tasks []*core.Record) {
	for i := 0; i < len(tasks)-1; i++ {
		for j := i + 1; j < len(tasks); j++ {
			if tasks[i].GetFloat("position") > tasks[j].GetFloat("position") {
				tasks[i], tasks[j] = tasks[j], tasks[i]
			}
		}
	}
}
```

**Steps**:

1. Create the file:
   ```bash
   touch internal/commands/position.go
   ```

2. Open in your editor and paste the code above.

3. Verify it compiles:
   ```bash
   go build ./internal/commands
   ```
   
   **Expected output**: No output means success!

**How Fractional Indexing Works**:

```
Initial:        After insert between 1 and 2:
1. Task A (1000)    1. Task A (1000)
2. Task B (2000)    2. Task C (1500)  ← new task
3. Task C (3000)    3. Task B (2000)
                    4. Task D (3000)
```

No need to renumber existing tasks!

---

### 1.5 Implement Root Command and Global Flags

**What**: Create the root command that registers all subcommands and handles global flags.

**Why**: Cobra uses a root command as the entry point. Global flags like `--json` and `--quiet` are defined here and inherited by all subcommands.

**File**: `internal/commands/root.go`

```go
package commands

import (
	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"

	"github.com/yourusername/egenskriven/internal/output"
)

var (
	// Global flags
	jsonOutput bool
	quietMode  bool
	dataDir    string
)

// Register adds all CLI commands to the PocketBase app.
func Register(app *pocketbase.PocketBase) {
	// Create formatter that will be used by all commands
	// Note: The formatter is created fresh for each command execution
	// to respect the current flag values.

	// Add global flags to root command
	app.RootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false,
		"Output in JSON format")
	app.RootCmd.PersistentFlags().BoolVarP(&quietMode, "quiet", "q", false,
		"Suppress non-essential output")
	app.RootCmd.PersistentFlags().StringVar(&dataDir, "data", "",
		"Path to data directory (default: pb_data)")

	// Register all commands
	app.RootCmd.AddCommand(newAddCmd(app))
	app.RootCmd.AddCommand(newListCmd(app))
	app.RootCmd.AddCommand(newShowCmd(app))
	app.RootCmd.AddCommand(newMoveCmd(app))
	app.RootCmd.AddCommand(newUpdateCmd(app))
	app.RootCmd.AddCommand(newDeleteCmd(app))
}

// getFormatter creates a new output formatter with current flag values.
// This should be called at the start of each command's RunE function.
func getFormatter() *output.Formatter {
	return output.New(jsonOutput, quietMode)
}

// Exit codes for CLI commands
const (
	ExitSuccess          = 0
	ExitGeneralError     = 1
	ExitInvalidArguments = 2
	ExitNotFound         = 3
	ExitAmbiguous        = 4
	ExitValidation       = 5
)

// ValidColumns is the list of valid column values
var ValidColumns = []string{"backlog", "todo", "in_progress", "review", "done"}

// ValidTypes is the list of valid task types
var ValidTypes = []string{"bug", "feature", "chore"}

// ValidPriorities is the list of valid priority values
var ValidPriorities = []string{"low", "medium", "high", "urgent"}

// isValidColumn checks if a column name is valid.
func isValidColumn(col string) bool {
	for _, valid := range ValidColumns {
		if col == valid {
			return true
		}
	}
	return false
}

// isValidType checks if a type is valid.
func isValidType(t string) bool {
	for _, valid := range ValidTypes {
		if t == valid {
			return true
		}
	}
	return false
}

// isValidPriority checks if a priority is valid.
func isValidPriority(p string) bool {
	for _, valid := range ValidPriorities {
		if p == valid {
			return true
		}
	}
	return false
}
```

**Steps**:

1. Create the file:
   ```bash
   touch internal/commands/root.go
   ```

2. Open in your editor and paste the code above.

3. **Important**: Replace `github.com/yourusername/egenskriven` with your actual module path from `go.mod`.

4. Verify it compiles:
   ```bash
   go build ./internal/commands
   ```

**Global Flags**:

| Flag | Short | Description |
|------|-------|-------------|
| `--json` | `-j` | Output in JSON format |
| `--quiet` | `-q` | Suppress non-essential output |
| `--data` | | Path to data directory |

---

### 1.6 Implement `add` Command

**What**: Create the command to add new tasks.

**Why**: This is the primary way to create tasks. It supports various options for setting task properties at creation time.

**File**: `internal/commands/add.go`

```go
package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
)

func newAddCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		taskType   string
		priority   string
		column     string
		labels     []string
		customID   string
		createdBy  string
		agentName  string
	)

	cmd := &cobra.Command{
		Use:   "add <title>",
		Short: "Add a new task",
		Long: `Add a new task to the kanban board.

The task will be created with the specified properties and added to the 
end of the target column.

Examples:
  egenskriven add "Implement dark mode"
  egenskriven add "Fix login crash" --type bug --priority urgent
  egenskriven add "Setup CI" --id ci-setup-001
  egenskriven add "Refactor auth" --agent claude`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			// Bootstrap the app to ensure database is ready
			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
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
					// Detect if running in a TTY
					if fileInfo, _ := os.Stdin.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
						createdBy = "user"
					} else {
						createdBy = "cli"
					}
				}
			}

			// Find the tasks collection
			collection, err := app.FindCollectionByNameOrId("tasks")
			if err != nil {
				return out.Error(ExitGeneralError,
					"tasks collection not found - run migrations first", nil)
			}

			// Create the record
			record := core.NewRecord(collection)

			// Set custom ID if provided
			if customID != "" {
				// Check if task with this ID already exists (idempotency)
				existing, err := app.FindRecordById("tasks", customID)
				if err == nil {
					// Task exists, return it (idempotent behavior)
					out.Task(existing, "Existing")
					return nil
				}
				record.SetId(customID)
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
				return out.Error(ExitGeneralError,
					fmt.Sprintf("failed to create task: %v", err), nil)
			}

			out.Task(record, "Created")
			return nil
		},
	}

	// Define flags
	cmd.Flags().StringVarP(&taskType, "type", "t", "feature",
		"Task type (bug, feature, chore)")
	cmd.Flags().StringVarP(&priority, "priority", "p", "medium",
		"Priority (low, medium, high, urgent)")
	cmd.Flags().StringVarP(&column, "column", "c", "backlog",
		"Initial column")
	cmd.Flags().StringSliceVarP(&labels, "label", "l", nil,
		"Labels (repeatable)")
	cmd.Flags().StringVar(&customID, "id", "",
		"Custom ID for idempotency")
	cmd.Flags().StringVar(&createdBy, "created-by", "",
		"Creator type (user, agent, cli)")
	cmd.Flags().StringVar(&agentName, "agent", "",
		"Agent identifier (implies --created-by agent)")

	return cmd
}
```

**Steps**:

1. Create the file:
   ```bash
   touch internal/commands/add.go
   ```

2. Open in your editor and paste the code above.

3. **Important**: Replace `github.com/yourusername/egenskriven` with your actual module path.

4. Verify it compiles:
   ```bash
   go build ./internal/commands
   ```

**Usage Examples**:

```bash
# Basic task creation
egenskriven add "Implement dark mode"

# With all options
egenskriven add "Fix login crash" --type bug --priority urgent --column todo

# With labels
egenskriven add "Add user avatars" --label ui --label frontend

# Idempotent creation (safe for retries)
egenskriven add "Setup CI" --id ci-setup-001

# Agent creating a task
egenskriven add "Refactor auth module" --agent claude
```

**Flags**:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--type` | `-t` | feature | Task type (bug, feature, chore) |
| `--priority` | `-p` | medium | Priority (low, medium, high, urgent) |
| `--column` | `-c` | backlog | Initial column |
| `--label` | `-l` | | Labels (repeatable) |
| `--id` | | | Custom ID for idempotency |
| `--created-by` | | | Creator type (user, agent, cli) |
| `--agent` | | | Agent identifier |

---

### 1.7 Implement `list` Command

**What**: Create the command to list and filter tasks.

**Why**: Listing tasks is essential for seeing what's on the board. Filters allow narrowing down to specific tasks.

**File**: `internal/commands/list.go`

```go
package commands

import (
	"fmt"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
)

func newListCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		columns    []string
		types      []string
		priorities []string
		search     string
		createdBy  string
		agentName  string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		Long: `List and filter tasks on the kanban board.

By default, shows all tasks grouped by column. Use flags to filter.

Examples:
  egenskriven list
  egenskriven list --column todo
  egenskriven list --type bug --priority urgent
  egenskriven list --json`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			// Bootstrap the app
			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			// Build filter expressions
			var filters []dbx.Expression

			// Column filter
			if len(columns) > 0 {
				for _, col := range columns {
					if !isValidColumn(col) {
						return out.Error(ExitValidation,
							fmt.Sprintf("invalid column '%s', must be one of: %v", col, ValidColumns), nil)
					}
				}
				filters = append(filters, buildInFilter("column", columns))
			}

			// Type filter
			if len(types) > 0 {
				for _, t := range types {
					if !isValidType(t) {
						return out.Error(ExitValidation,
							fmt.Sprintf("invalid type '%s', must be one of: %v", t, ValidTypes), nil)
					}
				}
				filters = append(filters, buildInFilter("type", types))
			}

			// Priority filter
			if len(priorities) > 0 {
				for _, p := range priorities {
					if !isValidPriority(p) {
						return out.Error(ExitValidation,
							fmt.Sprintf("invalid priority '%s', must be one of: %v", p, ValidPriorities), nil)
					}
				}
				filters = append(filters, buildInFilter("priority", priorities))
			}

			// Search filter
			if search != "" {
				filters = append(filters, dbx.NewExp(
					"LOWER(title) LIKE {:search}",
					dbx.Params{"search": "%" + strings.ToLower(search) + "%"},
				))
			}

			// Created by filter
			if createdBy != "" {
				filters = append(filters, dbx.NewExp(
					"created_by = {:created_by}",
					dbx.Params{"created_by": createdBy},
				))
			}

			// Agent name filter
			if agentName != "" {
				filters = append(filters, dbx.NewExp(
					"created_by_agent = {:agent}",
					dbx.Params{"agent": agentName},
				))
			}

			// Execute query
			var tasks []*core.Record
			var err error

			if len(filters) > 0 {
				combined := dbx.And(filters...)
				tasks, err = app.FindAllRecords("tasks", combined)
			} else {
				tasks, err = app.FindAllRecords("tasks")
			}

			if err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to list tasks: %v", err), nil)
			}

			// Sort by position within each column
			sortTasksByPosition(tasks)

			out.Tasks(tasks)
			return nil
		},
	}

	// Define flags
	cmd.Flags().StringSliceVarP(&columns, "column", "c", nil,
		"Filter by column (repeatable)")
	cmd.Flags().StringSliceVarP(&types, "type", "t", nil,
		"Filter by type (repeatable)")
	cmd.Flags().StringSliceVarP(&priorities, "priority", "p", nil,
		"Filter by priority (repeatable)")
	cmd.Flags().StringVarP(&search, "search", "s", "",
		"Search title (case-insensitive)")
	cmd.Flags().StringVar(&createdBy, "created-by", "",
		"Filter by creator (user, agent, cli)")
	cmd.Flags().StringVar(&agentName, "agent", "",
		"Filter by agent name")

	return cmd
}

// buildInFilter creates a SQL IN expression for a list of values.
func buildInFilter(field string, values []string) dbx.Expression {
	if len(values) == 1 {
		return dbx.NewExp(
			fmt.Sprintf("%s = {:val}", field),
			dbx.Params{"val": values[0]},
		)
	}

	// Build IN clause
	placeholders := make([]string, len(values))
	params := dbx.Params{}
	for i, v := range values {
		key := fmt.Sprintf("val%d", i)
		placeholders[i] = "{:" + key + "}"
		params[key] = v
	}

	return dbx.NewExp(
		fmt.Sprintf("%s IN (%s)", field, strings.Join(placeholders, ", ")),
		params,
	)
}
```

**Steps**:

1. Create the file:
   ```bash
   touch internal/commands/list.go
   ```

2. Open in your editor and paste the code above.

3. Verify it compiles:
   ```bash
   go build ./internal/commands
   ```

**Usage Examples**:

```bash
# List all tasks
egenskriven list

# Filter by column
egenskriven list --column todo

# Multiple filters (AND logic)
egenskriven list --column todo --type bug --priority high

# Search by title
egenskriven list --search "login"

# JSON output
egenskriven list --json

# Filter by creator
egenskriven list --created-by agent --agent claude
```

**Output Formats**:

Human output (default):
```
BACKLOG
  [abc12345] Implement dark mode (feature)
  
TODO
  [def67890] Fix login crash (bug, !!!urgent)

IN_PROGRESS
  (empty)
```

JSON output (`--json`):
```json
{
  "tasks": [
    {
      "id": "abc12345def67890",
      "title": "Implement dark mode",
      "type": "feature",
      "priority": "medium",
      "column": "backlog"
    }
  ],
  "count": 1
}
```

---

### 1.8 Implement `show` Command

**What**: Create the command to display detailed information about a task.

**Why**: The list command shows minimal info. The show command displays all task details including description, labels, and history.

**File**: `internal/commands/show.go`

```go
package commands

import (
	"fmt"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"

	"github.com/yourusername/egenskriven/internal/resolver"
)

func newShowCmd(app *pocketbase.PocketBase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <task>",
		Short: "Show task details",
		Long: `Show detailed information about a task.

The task can be referenced by:
- Full ID: abc123def456
- ID prefix: abc123 (must be unique)
- Title: "fix login" (case-insensitive, must be unique)

Examples:
  egenskriven show abc123
  egenskriven show "login crash"
  egenskriven show abc123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			// Bootstrap the app
			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			ref := args[0]

			// Resolve the task
			resolution, err := resolver.ResolveTask(app, ref)
			if err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to resolve task: %v", err), nil)
			}

			if resolution.IsNotFound() {
				return out.Error(ExitNotFound, fmt.Sprintf("no task found matching: %s", ref), nil)
			}

			if resolution.IsAmbiguous() {
				return out.AmbiguousError(ref, resolution.Matches)
			}

			out.TaskDetail(resolution.Task)
			return nil
		},
	}

	return cmd
}
```

**Steps**:

1. Create the file:
   ```bash
   touch internal/commands/show.go
   ```

2. Open in your editor and paste the code above.

3. **Important**: Replace `github.com/yourusername/egenskriven` with your actual module path.

4. Verify it compiles:
   ```bash
   go build ./internal/commands
   ```

**Usage Examples**:

```bash
# By ID
egenskriven show abc123def456

# By ID prefix
egenskriven show abc123

# By title
egenskriven show "login crash"

# JSON output
egenskriven show abc123 --json
```

**Output Example** (human):
```
Task: abc123def456
Title:       Fix login crash
Type:        bug
Priority:    urgent
Column:      todo
Position:    1000
Labels:      auth, critical
Created by:  agent (claude)
Created:     2024-01-15 10:30:00
Updated:     2024-01-15 14:22:00

Description:
  Users are experiencing crashes when attempting to log in
  with SSO credentials. Stack trace attached.
```

---

### 1.9 Implement `move` Command

**What**: Create the command to move tasks between columns and positions.

**Why**: Moving tasks is a core kanban operation. This command handles changing columns and reordering within columns.

**File**: `internal/commands/move.go`

```go
package commands

import (
	"fmt"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"

	"github.com/yourusername/egenskriven/internal/resolver"
)

func newMoveCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		position int
		afterID  string
		beforeID string
	)

	cmd := &cobra.Command{
		Use:   "move <task> [column]",
		Short: "Move task to column/position",
		Long: `Move a task to a different column and/or position.

If a column is specified, the task is moved to that column.
Position can be controlled with --position, --after, or --before.

Position values:
  0  = top of column
  -1 = bottom of column (default)

Examples:
  egenskriven move abc123 in_progress
  egenskriven move abc123 todo --position 0
  egenskriven move abc123 --after def456
  egenskriven move abc123 --before ghi789`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			// Bootstrap the app
			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			taskRef := args[0]

			// Resolve the task
			task, err := resolver.MustResolve(app, taskRef)
			if err != nil {
				if ambErr, ok := err.(*resolver.AmbiguousError); ok {
					return out.AmbiguousError(taskRef, ambErr.Matches)
				}
				return out.Error(ExitNotFound, err.Error(), nil)
			}

			// Determine target column
			currentColumn := task.GetString("column")
			targetColumn := currentColumn

			if len(args) > 1 {
				targetColumn = args[1]
				if !isValidColumn(targetColumn) {
					return out.Error(ExitValidation,
						fmt.Sprintf("invalid column '%s', must be one of: %v", targetColumn, ValidColumns), nil)
				}
			}

			// Calculate new position
			var newPosition float64

			if afterID != "" {
				// Position after specific task
				pos, err := GetPositionAfter(app, afterID)
				if err != nil {
					return out.Error(ExitNotFound,
						fmt.Sprintf("task not found: %s", afterID), nil)
				}
				newPosition = pos

				// Get target column from reference task
				refTask, _ := app.FindRecordById("tasks", afterID)
				if refTask != nil && len(args) < 2 {
					targetColumn = refTask.GetString("column")
				}
			} else if beforeID != "" {
				// Position before specific task
				pos, err := GetPositionBefore(app, beforeID)
				if err != nil {
					return out.Error(ExitNotFound,
						fmt.Sprintf("task not found: %s", beforeID), nil)
				}
				newPosition = pos

				// Get target column from reference task
				refTask, _ := app.FindRecordById("tasks", beforeID)
				if refTask != nil && len(args) < 2 {
					targetColumn = refTask.GetString("column")
				}
			} else if position >= 0 {
				// Specific position index
				newPosition = GetPositionAtIndex(app, targetColumn, position)
			} else {
				// Default: end of column
				newPosition = GetNextPosition(app, targetColumn)
			}

			// Track changes for history
			oldColumn := currentColumn
			oldPosition := task.GetFloat("position")

			// Update the task
			task.Set("column", targetColumn)
			task.Set("position", newPosition)

			// Add to history
			addHistoryEntry(task, "moved", "", map[string]any{
				"column": map[string]any{
					"from": oldColumn,
					"to":   targetColumn,
				},
				"position": map[string]any{
					"from": oldPosition,
					"to":   newPosition,
				},
			})

			if err := app.Save(task); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to move task: %v", err), nil)
			}

			if targetColumn != oldColumn {
				out.Success(fmt.Sprintf("Moved task [%s] from %s to %s", task.Id[:8], oldColumn, targetColumn))
			} else {
				out.Success(fmt.Sprintf("Repositioned task [%s] in %s", task.Id[:8], targetColumn))
			}

			return nil
		},
	}

	// Define flags
	cmd.Flags().IntVarP(&position, "position", "", -1,
		"Position index (0=top, -1=bottom)")
	cmd.Flags().StringVar(&afterID, "after", "",
		"Position after this task")
	cmd.Flags().StringVar(&beforeID, "before", "",
		"Position before this task")

	return cmd
}

// addHistoryEntry appends an entry to the task's history.
func addHistoryEntry(task interface{ Get(string) any; Set(string, any) }, action, agent string, changes any) {
	var history []map[string]any

	// Get existing history
	raw := task.Get("history")
	if raw != nil {
		if h, ok := raw.([]any); ok {
			for _, entry := range h {
				if m, ok := entry.(map[string]any); ok {
					history = append(history, m)
				}
			}
		}
	}

	// Create new entry
	entry := map[string]any{
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"action":       action,
		"actor":        "cli",
		"actor_detail": agent,
		"changes":      changes,
	}

	history = append(history, entry)
	task.Set("history", history)
}
```

**Steps**:

1. Create the file:
   ```bash
   touch internal/commands/move.go
   ```

2. Open in your editor and paste the code above.

3. **Important**: Replace `github.com/yourusername/egenskriven` with your actual module path.

4. Verify it compiles:
   ```bash
   go build ./internal/commands
   ```

**Usage Examples**:

```bash
# Move to different column (appends to end)
egenskriven move abc123 in_progress

# Move to top of column
egenskriven move abc123 todo --position 0

# Move after another task
egenskriven move abc123 --after def456

# Move before another task
egenskriven move abc123 --before ghi789
```

---

### 1.10 Implement `update` Command

**What**: Create the command to update task properties.

**Why**: Tasks need to be modified after creation - changing priority, adding labels, updating description, etc.

**File**: `internal/commands/update.go`

```go
package commands

import (
	"fmt"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"

	"github.com/yourusername/egenskriven/internal/resolver"
)

func newUpdateCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		title        string
		description  string
		taskType     string
		priority     string
		addLabels    []string
		removeLabels []string
	)

	cmd := &cobra.Command{
		Use:   "update <task>",
		Short: "Update task properties",
		Long: `Update properties of an existing task.

Only specified fields will be updated. Use empty string to clear optional fields.

Examples:
  egenskriven update abc123 --title "New title"
  egenskriven update abc123 --priority urgent
  egenskriven update abc123 --add-label critical --remove-label backlog
  egenskriven update abc123 --description ""  # clears description`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			// Bootstrap the app
			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			taskRef := args[0]

			// Resolve the task
			task, err := resolver.MustResolve(app, taskRef)
			if err != nil {
				if ambErr, ok := err.(*resolver.AmbiguousError); ok {
					return out.AmbiguousError(taskRef, ambErr.Matches)
				}
				return out.Error(ExitNotFound, err.Error(), nil)
			}

			// Track changes
			changes := make(map[string]any)

			// Update title
			if cmd.Flags().Changed("title") {
				if title == "" {
					return out.Error(ExitValidation, "title cannot be empty", nil)
				}
				changes["title"] = map[string]any{
					"from": task.GetString("title"),
					"to":   title,
				}
				task.Set("title", title)
			}

			// Update description
			if cmd.Flags().Changed("description") {
				changes["description"] = map[string]any{
					"from": task.GetString("description"),
					"to":   description,
				}
				task.Set("description", description)
			}

			// Update type
			if cmd.Flags().Changed("type") {
				if !isValidType(taskType) {
					return out.Error(ExitValidation,
						fmt.Sprintf("invalid type '%s', must be one of: %v", taskType, ValidTypes), nil)
				}
				changes["type"] = map[string]any{
					"from": task.GetString("type"),
					"to":   taskType,
				}
				task.Set("type", taskType)
			}

			// Update priority
			if cmd.Flags().Changed("priority") {
				if !isValidPriority(priority) {
					return out.Error(ExitValidation,
						fmt.Sprintf("invalid priority '%s', must be one of: %v", priority, ValidPriorities), nil)
				}
				changes["priority"] = map[string]any{
					"from": task.GetString("priority"),
					"to":   priority,
				}
				task.Set("priority", priority)
			}

			// Update labels
			if len(addLabels) > 0 || len(removeLabels) > 0 {
				oldLabels := getTaskLabels(task)
				newLabels := updateLabels(oldLabels, addLabels, removeLabels)
				changes["labels"] = map[string]any{
					"from": oldLabels,
					"to":   newLabels,
				}
				task.Set("labels", newLabels)
			}

			// Check if any changes were made
			if len(changes) == 0 {
				return out.Error(ExitValidation, "no changes specified", nil)
			}

			// Add to history
			addHistoryEntry(task, "updated", "", changes)

			// Save the task
			if err := app.Save(task); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to update task: %v", err), nil)
			}

			out.Task(task, "Updated")
			return nil
		},
	}

	// Define flags
	cmd.Flags().StringVar(&title, "title", "", "New title")
	cmd.Flags().StringVar(&description, "description", "", "New description")
	cmd.Flags().StringVarP(&taskType, "type", "t", "", "New type (bug, feature, chore)")
	cmd.Flags().StringVarP(&priority, "priority", "p", "", "New priority (low, medium, high, urgent)")
	cmd.Flags().StringSliceVar(&addLabels, "add-label", nil, "Add label (repeatable)")
	cmd.Flags().StringSliceVar(&removeLabels, "remove-label", nil, "Remove label (repeatable)")

	return cmd
}

// getTaskLabels extracts labels from a task record.
func getTaskLabels(task interface{ Get(string) any }) []string {
	raw := task.Get("labels")
	if raw == nil {
		return []string{}
	}

	if labels, ok := raw.([]any); ok {
		result := make([]string, 0, len(labels))
		for _, l := range labels {
			if s, ok := l.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}

	return []string{}
}

// updateLabels adds and removes labels from a list.
func updateLabels(current, add, remove []string) []string {
	// Create a set of current labels
	labelSet := make(map[string]bool)
	for _, l := range current {
		labelSet[l] = true
	}

	// Remove labels
	for _, l := range remove {
		delete(labelSet, l)
	}

	// Add labels
	for _, l := range add {
		labelSet[l] = true
	}

	// Convert back to slice
	result := make([]string, 0, len(labelSet))
	for l := range labelSet {
		result = append(result, l)
	}

	return result
}
```

**Steps**:

1. Create the file:
   ```bash
   touch internal/commands/update.go
   ```

2. Open in your editor and paste the code above.

3. **Important**: Replace `github.com/yourusername/egenskriven` with your actual module path.

4. Verify it compiles:
   ```bash
   go build ./internal/commands
   ```

**Usage Examples**:

```bash
# Update single field
egenskriven update abc123 --title "New title"

# Update multiple fields
egenskriven update abc123 --type bug --priority urgent

# Manage labels
egenskriven update abc123 --add-label critical --remove-label backlog

# Clear description
egenskriven update abc123 --description ""
```

---

### 1.11 Implement `delete` Command

**What**: Create the command to delete tasks.

**Why**: Tasks need to be removed when they're no longer relevant or were created in error.

**File**: `internal/commands/delete.go`

```go
package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"

	"github.com/yourusername/egenskriven/internal/resolver"
)

func newDeleteCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		force bool
	)

	cmd := &cobra.Command{
		Use:   "delete <task> [task...]",
		Short: "Delete tasks",
		Long: `Delete one or more tasks.

By default, asks for confirmation before deleting.
Use --force to skip confirmation (useful for scripts).

Examples:
  egenskriven delete abc123
  egenskriven delete abc123 def456 ghi789
  egenskriven delete abc123 --force`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			// Bootstrap the app
			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			// Resolve all tasks first
			var tasksToDelete []struct {
				ref  string
				task interface{ GetId() string; GetString(string) string }
			}

			for _, ref := range args {
				task, err := resolver.MustResolve(app, ref)
				if err != nil {
					if ambErr, ok := err.(*resolver.AmbiguousError); ok {
						return out.AmbiguousError(ref, ambErr.Matches)
					}
					return out.Error(ExitNotFound, err.Error(), nil)
				}
				tasksToDelete = append(tasksToDelete, struct {
					ref  string
					task interface{ GetId() string; GetString(string) string }
				}{ref, task})
			}

			// Confirm deletion unless --force
			if !force && !jsonOutput {
				fmt.Printf("About to delete %d task(s):\n", len(tasksToDelete))
				for _, t := range tasksToDelete {
					fmt.Printf("  [%s] %s\n", t.task.GetId()[:8], t.task.GetString("title"))
				}
				fmt.Print("\nConfirm deletion? [y/N]: ")

				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				response = strings.TrimSpace(strings.ToLower(response))

				if response != "y" && response != "yes" {
					fmt.Println("Deletion cancelled.")
					return nil
				}
			}

			// Delete tasks
			var deleted int
			for _, t := range tasksToDelete {
				// Need to get the actual record for deletion
				record, err := app.FindRecordById("tasks", t.task.GetId())
				if err != nil {
					continue
				}

				if err := app.Delete(record); err != nil {
					return out.Error(ExitGeneralError,
						fmt.Sprintf("failed to delete task %s: %v", t.ref, err), nil)
				}
				deleted++
			}

			if jsonOutput {
				out.Success(fmt.Sprintf("Deleted %d task(s)", deleted))
			} else {
				fmt.Printf("Deleted %d task(s)\n", deleted)
			}

			return nil
		},
	}

	// Define flags
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")

	return cmd
}
```

**Steps**:

1. Create the file:
   ```bash
   touch internal/commands/delete.go
   ```

2. Open in your editor and paste the code above.

3. **Important**: Replace `github.com/yourusername/egenskriven` with your actual module path.

4. Verify it compiles:
   ```bash
   go build ./internal/commands
   ```

**Usage Examples**:

```bash
# Delete with confirmation
egenskriven delete abc123

# Delete multiple tasks
egenskriven delete abc123 def456 ghi789

# Delete without confirmation (for scripts)
egenskriven delete abc123 --force
```

---

### 1.12 Update Main Entry Point

**What**: Update `main.go` to register all CLI commands.

**Why**: The commands we created need to be registered with PocketBase's root command.

**File**: `cmd/egenskriven/main.go`

```go
package main

import (
	"log"

	"github.com/pocketbase/pocketbase"

	"github.com/yourusername/egenskriven/internal/commands"
	_ "github.com/yourusername/egenskriven/migrations" // Auto-register migrations
)

func main() {
	app := pocketbase.New()

	// Register custom CLI commands
	commands.Register(app)

	// Start the application
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
```

**Steps**:

1. Open `cmd/egenskriven/main.go` in your editor.

2. Replace the contents with the code above.

3. **Important**: Replace `github.com/yourusername/egenskriven` with your actual module path.

4. Build the application:
   ```bash
   make build
   ```
   
   **Expected output**:
   ```
   Building production binary...
   Built: ./egenskriven (35M)
   ```

5. Test the help command:
   ```bash
   ./egenskriven --help
   ```
   
   **Expected output**:
   ```
   Usage:
     egenskriven [command]

   Available Commands:
     add         Add a new task
     delete      Delete tasks
     list        List tasks
     migrate     Executes app DB migrations
     move        Move task to column/position
     serve       Starts the web server
     show        Show task details
     update      Update task properties
     ...
   ```

---

### 1.13 Write Unit Tests

**What**: Create unit tests for the resolver, output formatter, and position calculator.

**Why**: Unit tests ensure individual components work correctly in isolation. They run fast and catch bugs early.

**File**: `internal/resolver/resolver_test.go`

```go
package resolver

import (
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourusername/egenskriven/internal/testutil"
)

func TestResolveTask_ExactID(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	// Create a test task
	task := createTestTask(t, app, "Test Task", "feature", "medium", "backlog")

	// Resolve by exact ID
	resolution, err := ResolveTask(app, task.Id)

	require.NoError(t, err)
	assert.False(t, resolution.IsAmbiguous())
	assert.False(t, resolution.IsNotFound())
	assert.Equal(t, task.Id, resolution.Task.Id)
}

func TestResolveTask_IDPrefix(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	task := createTestTask(t, app, "Test Task", "feature", "medium", "backlog")

	// Resolve by ID prefix (first 6 characters)
	prefix := task.Id[:6]
	resolution, err := ResolveTask(app, prefix)

	require.NoError(t, err)
	assert.False(t, resolution.IsAmbiguous())
	assert.Equal(t, task.Id, resolution.Task.Id)
}

func TestResolveTask_TitleMatch(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	task := createTestTask(t, app, "Fix login authentication bug", "bug", "high", "todo")

	// Resolve by title substring (case-insensitive)
	resolution, err := ResolveTask(app, "login auth")

	require.NoError(t, err)
	assert.False(t, resolution.IsAmbiguous())
	assert.Equal(t, task.Id, resolution.Task.Id)
}

func TestResolveTask_Ambiguous(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	// Create multiple tasks with similar titles
	createTestTask(t, app, "Fix login bug", "bug", "high", "todo")
	createTestTask(t, app, "Fix login crash", "bug", "urgent", "todo")

	// Resolve with ambiguous reference
	resolution, err := ResolveTask(app, "login")

	require.NoError(t, err)
	assert.True(t, resolution.IsAmbiguous())
	assert.Len(t, resolution.Matches, 2)
}

func TestResolveTask_NotFound(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	resolution, err := ResolveTask(app, "nonexistent")

	require.NoError(t, err)
	assert.True(t, resolution.IsNotFound())
	assert.Nil(t, resolution.Task)
}

func TestMustResolve_Success(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	task := createTestTask(t, app, "Test Task", "feature", "medium", "backlog")

	resolved, err := MustResolve(app, task.Id)

	require.NoError(t, err)
	assert.Equal(t, task.Id, resolved.Id)
}

func TestMustResolve_NotFound(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	_, err := MustResolve(app, "nonexistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no task found")
}

func TestMustResolve_Ambiguous(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	createTestTask(t, app, "Task A", "feature", "medium", "backlog")
	createTestTask(t, app, "Task B", "feature", "medium", "backlog")

	_, err := MustResolve(app, "Task")

	require.Error(t, err)
	_, ok := err.(*AmbiguousError)
	assert.True(t, ok, "expected AmbiguousError")
}

// Helper functions

func setupTasksCollection(t *testing.T, app interface {
	FindCollectionByNameOrId(string) (*core.Collection, error)
	Save(any) error
}) {
	t.Helper()

	// Check if collection exists
	_, err := app.FindCollectionByNameOrId("tasks")
	if err == nil {
		return // Collection already exists
	}

	// Create tasks collection
	collection := core.NewBaseCollection("tasks")

	collection.Fields.Add(&core.TextField{Name: "title", Required: true})
	collection.Fields.Add(&core.TextField{Name: "description"})
	collection.Fields.Add(&core.SelectField{
		Name:     "type",
		Required: true,
		Values:   []string{"bug", "feature", "chore"},
	})
	collection.Fields.Add(&core.SelectField{
		Name:     "priority",
		Required: true,
		Values:   []string{"low", "medium", "high", "urgent"},
	})
	collection.Fields.Add(&core.SelectField{
		Name:     "column",
		Required: true,
		Values:   []string{"backlog", "todo", "in_progress", "review", "done"},
	})
	collection.Fields.Add(&core.NumberField{Name: "position", Required: true})
	collection.Fields.Add(&core.JSONField{Name: "labels"})
	collection.Fields.Add(&core.JSONField{Name: "blocked_by"})
	collection.Fields.Add(&core.SelectField{
		Name:     "created_by",
		Required: true,
		Values:   []string{"user", "agent", "cli"},
	})
	collection.Fields.Add(&core.TextField{Name: "created_by_agent"})
	collection.Fields.Add(&core.JSONField{Name: "history"})

	if err := app.Save(collection); err != nil {
		t.Fatalf("failed to create tasks collection: %v", err)
	}
}

func createTestTask(t *testing.T, app interface {
	FindCollectionByNameOrId(string) (*core.Collection, error)
	Save(any) error
}, title, taskType, priority, column string) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("tasks")
	if err != nil {
		t.Fatalf("tasks collection not found: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("title", title)
	record.Set("type", taskType)
	record.Set("priority", priority)
	record.Set("column", column)
	record.Set("position", 1000.0)
	record.Set("labels", []string{})
	record.Set("blocked_by", []string{})
	record.Set("created_by", "cli")
	record.Set("history", []map[string]any{})

	if err := app.Save(record); err != nil {
		t.Fatalf("failed to create test task: %v", err)
	}

	return record
}
```

**File**: `internal/output/output_test.go`

```go
package output

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatter_JSON_Mode(t *testing.T) {
	f := New(true, false)
	assert.True(t, f.JSON)
	assert.False(t, f.Quiet)
}

func TestFormatter_Quiet_Mode(t *testing.T) {
	f := New(false, true)
	assert.False(t, f.JSON)
	assert.True(t, f.Quiet)
}

func TestTaskToMap(t *testing.T) {
	// This test would require a mock record
	// For now, just test the helper functions
	t.Skip("Requires mock PocketBase record")
}

func TestShortID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"abc123def456", "abc123de"},
		{"short", "short"},
		{"12345678", "12345678"},
		{"123456789", "12345678"},
	}

	for _, tt := range tests {
		result := shortID(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestGroupByColumn(t *testing.T) {
	// Would require mock records
	t.Skip("Requires mock PocketBase records")
}

func TestFormatter_Error_Output(t *testing.T) {
	// Capture stderr
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	f := New(true, false)
	_ = f.Error(1, "test error", nil)

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var result map[string]any
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	errorObj := result["error"].(map[string]any)
	assert.Equal(t, float64(1), errorObj["code"])
	assert.Equal(t, "test error", errorObj["message"])
}
```

**File**: `internal/commands/position_test.go`

```go
package commands

import (
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"

	"github.com/yourusername/egenskriven/internal/testutil"
)

func TestGetNextPosition_EmptyColumn(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	pos := GetNextPosition(app, "backlog")

	assert.Equal(t, DefaultPositionGap, pos)
}

func TestGetNextPosition_WithExistingTasks(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	// Create a task at position 1000
	createTestTask(t, app, "Task 1", "feature", "medium", "backlog", 1000)

	pos := GetNextPosition(app, "backlog")

	assert.Equal(t, 2000.0, pos)
}

func TestGetPositionBetween(t *testing.T) {
	tests := []struct {
		before   float64
		after    float64
		expected float64
	}{
		{1000, 2000, 1500},
		{0, 1000, 500},
		{500, 600, 550},
	}

	for _, tt := range tests {
		result := GetPositionBetween(tt.before, tt.after)
		assert.Equal(t, tt.expected, result)
	}
}

func TestGetPositionAtIndex_Top(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	createTestTask(t, app, "Task 1", "feature", "medium", "backlog", 1000)
	createTestTask(t, app, "Task 2", "feature", "medium", "backlog", 2000)

	pos := GetPositionAtIndex(app, "backlog", 0)

	// Should be half of the first task's position
	assert.Equal(t, 500.0, pos)
}

func TestGetPositionAtIndex_Bottom(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	createTestTask(t, app, "Task 1", "feature", "medium", "backlog", 1000)

	pos := GetPositionAtIndex(app, "backlog", -1)

	assert.Equal(t, 2000.0, pos)
}

func TestSortTasksByPosition(t *testing.T) {
	// Create mock records with different positions
	// This would require creating actual core.Record instances
	t.Skip("Requires mock PocketBase records")
}

// Helper functions for position tests

func setupTasksCollection(t *testing.T, app interface {
	FindCollectionByNameOrId(string) (*core.Collection, error)
	Save(any) error
}) {
	t.Helper()

	_, err := app.FindCollectionByNameOrId("tasks")
	if err == nil {
		return
	}

	collection := core.NewBaseCollection("tasks")
	collection.Fields.Add(&core.TextField{Name: "title", Required: true})
	collection.Fields.Add(&core.TextField{Name: "description"})
	collection.Fields.Add(&core.SelectField{
		Name:     "type",
		Required: true,
		Values:   []string{"bug", "feature", "chore"},
	})
	collection.Fields.Add(&core.SelectField{
		Name:     "priority",
		Required: true,
		Values:   []string{"low", "medium", "high", "urgent"},
	})
	collection.Fields.Add(&core.SelectField{
		Name:     "column",
		Required: true,
		Values:   []string{"backlog", "todo", "in_progress", "review", "done"},
	})
	collection.Fields.Add(&core.NumberField{Name: "position", Required: true})
	collection.Fields.Add(&core.JSONField{Name: "labels"})
	collection.Fields.Add(&core.JSONField{Name: "blocked_by"})
	collection.Fields.Add(&core.SelectField{
		Name:     "created_by",
		Required: true,
		Values:   []string{"user", "agent", "cli"},
	})
	collection.Fields.Add(&core.TextField{Name: "created_by_agent"})
	collection.Fields.Add(&core.JSONField{Name: "history"})

	if err := app.Save(collection); err != nil {
		t.Fatalf("failed to create tasks collection: %v", err)
	}
}

func createTestTask(t *testing.T, app interface {
	FindCollectionByNameOrId(string) (*core.Collection, error)
	Save(any) error
}, title, taskType, priority, column string, position float64) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("tasks")
	if err != nil {
		t.Fatalf("tasks collection not found: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("title", title)
	record.Set("type", taskType)
	record.Set("priority", priority)
	record.Set("column", column)
	record.Set("position", position)
	record.Set("labels", []string{})
	record.Set("blocked_by", []string{})
	record.Set("created_by", "cli")
	record.Set("history", []map[string]any{})

	if err := app.Save(record); err != nil {
		t.Fatalf("failed to create test task: %v", err)
	}

	return record
}
```

**Steps**:

1. Create the test files:
   ```bash
   touch internal/resolver/resolver_test.go
   touch internal/output/output_test.go
   touch internal/commands/position_test.go
   ```

2. Open each file in your editor and paste the corresponding code.

3. **Important**: Replace `github.com/yourusername/egenskriven` with your actual module path.

4. Run the tests:
   ```bash
   make test
   ```
   
   **Expected output**: All tests pass (some may be skipped).

---

## Verification Checklist

Complete each section in order. Check off each item as you verify it.

### Build Verification

- [ ] **Project compiles**
  ```bash
  make build
  ```
  Should produce `egenskriven` binary.

- [ ] **All tests pass**
  ```bash
  make test
  ```
  Should show all tests passing.

### CLI Command Verification

- [ ] **Server starts and migrations run**
  ```bash
  ./egenskriven serve
  ```
  Visit `http://localhost:8090/_/` and verify `tasks` collection exists.

- [ ] **Add command works**
  ```bash
  ./egenskriven add "Test task"
  ```
  Should output: `Created task: Test task [xxxxxxxx]`

- [ ] **Add with options works**
  ```bash
  ./egenskriven add "Bug fix" --type bug --priority urgent --column todo
  ```
  Should create task with specified properties.

- [ ] **Add with custom ID works (idempotency)**
  ```bash
  ./egenskriven add "Setup CI" --id ci-setup-001
  ./egenskriven add "Setup CI" --id ci-setup-001
  ```
  Second command should return existing task.

- [ ] **List command works**
  ```bash
  ./egenskriven list
  ```
  Should show tasks grouped by column.

- [ ] **List with filters works**
  ```bash
  ./egenskriven list --column todo
  ./egenskriven list --type bug --priority urgent
  ```
  Should filter results correctly.

- [ ] **List with JSON output works**
  ```bash
  ./egenskriven list --json
  ```
  Should output valid JSON with tasks array.

- [ ] **Show command works**
  ```bash
  ./egenskriven show "Test task"
  ```
  Should display task details.

- [ ] **Show by ID prefix works**
  Create a task and note its ID, then:
  ```bash
  ./egenskriven show <first-6-chars>
  ```
  Should find the task.

- [ ] **Move command works**
  ```bash
  ./egenskriven move <task-id> in_progress
  ```
  Should move task to new column.

- [ ] **Update command works**
  ```bash
  ./egenskriven update <task-id> --priority high --add-label important
  ```
  Should update task properties.

- [ ] **Delete command works**
  ```bash
  ./egenskriven delete <task-id>
  ```
  Should prompt for confirmation and delete.

- [ ] **Delete with --force works**
  ```bash
  ./egenskriven delete <task-id> --force
  ```
  Should delete without confirmation.

### Error Handling Verification

- [ ] **Invalid column shows error**
  ```bash
  ./egenskriven add "Test" --column invalid
  ```
  Should show validation error.

- [ ] **Not found shows error**
  ```bash
  ./egenskriven show nonexistent
  ```
  Should show "no task found" error.

- [ ] **Ambiguous reference shows matches**
  Create two tasks with "test" in title, then:
  ```bash
  ./egenskriven show test
  ```
  Should list matching tasks.

- [ ] **JSON errors are structured**
  ```bash
  ./egenskriven show nonexistent --json
  ```
  Should output JSON with error object containing code and message.

### Exit Code Verification

- [ ] **Success returns 0**
  ```bash
  ./egenskriven list; echo $?
  ```
  Should print `0`.

- [ ] **Validation error returns 5**
  ```bash
  ./egenskriven add "Test" --column invalid; echo $?
  ```
  Should print `5`.

---

## File Summary

| File | Lines | Purpose |
|------|-------|---------|
| `migrations/1_initial.go` | ~80 | Tasks collection schema |
| `internal/output/output.go` | ~180 | Output formatting (JSON/human) |
| `internal/resolver/resolver.go` | ~80 | Task ID/title resolution |
| `internal/commands/root.go` | ~80 | Root command, global flags |
| `internal/commands/position.go` | ~120 | Position calculation |
| `internal/commands/add.go` | ~130 | Add task command |
| `internal/commands/list.go` | ~130 | List tasks command |
| `internal/commands/show.go` | ~60 | Show task details command |
| `internal/commands/move.go` | ~150 | Move task command |
| `internal/commands/update.go` | ~150 | Update task command |
| `internal/commands/delete.go` | ~100 | Delete task command |
| `cmd/egenskriven/main.go` | ~20 | Entry point |
| `internal/resolver/resolver_test.go` | ~150 | Resolver unit tests |
| `internal/output/output_test.go` | ~80 | Output formatter tests |
| `internal/commands/position_test.go` | ~100 | Position calculation tests |

**Total new code**: ~1,610 lines

---

## What You Should Have Now

After completing Phase 1, your project should:

```
egenskriven/
├── cmd/
│   └── egenskriven/
│       └── main.go              ✓ Updated
├── internal/
│   ├── commands/
│   │   ├── root.go              ✓ Created
│   │   ├── add.go               ✓ Created
│   │   ├── list.go              ✓ Created
│   │   ├── show.go              ✓ Created
│   │   ├── move.go              ✓ Created
│   │   ├── update.go            ✓ Created
│   │   ├── delete.go            ✓ Created
│   │   ├── position.go          ✓ Created
│   │   └── position_test.go     ✓ Created
│   ├── output/
│   │   ├── output.go            ✓ Created
│   │   └── output_test.go       ✓ Created
│   ├── resolver/
│   │   ├── resolver.go          ✓ Created
│   │   └── resolver_test.go     ✓ Created
│   ├── config/                  ✓ Empty (Phase 1.5)
│   ├── hooks/                   ✓ Empty (Phase 1.5)
│   └── testutil/
│       ├── testutil.go          ✓ Created (Phase 0)
│       └── testutil_test.go     ✓ Created (Phase 0)
├── migrations/
│   └── 1_initial.go             ✓ Created
├── ui/
│   └── embed.go                 ✓ Created (Phase 0)
├── .air.toml                    ✓ Created (Phase 0)
├── .gitignore                   ✓ Created (Phase 0)
├── go.mod                       ✓ Updated
├── go.sum                       ✓ Updated
└── Makefile                     ✓ Created (Phase 0)
```

---

## Next Phase

**Phase 1.5: Agent Integration** will add:
- Blocking relationships between tasks
- Per-project configuration (`.egenskriven/config.json`)
- Prime command for AI agent instructions
- Ready/blocked filters for `list` command
- Context and suggest commands
- OpenCode and Claude Code integration hooks

---

## Troubleshooting

### "tasks collection not found"

**Problem**: Migration hasn't run.

**Solution**:
```bash
# Start the server to run migrations
./egenskriven serve

# Or run migrations explicitly
./egenskriven migrate up
```

### "invalid type/priority/column"

**Problem**: Using a value not in the allowed list.

**Solution**: Use one of the valid values:
- Types: `bug`, `feature`, `chore`
- Priorities: `low`, `medium`, `high`, `urgent`
- Columns: `backlog`, `todo`, `in_progress`, `review`, `done`

### "ambiguous task reference"

**Problem**: Multiple tasks match your search.

**Solution**: Use a more specific reference:
- Full task ID
- Longer ID prefix
- More specific title match

### Tests fail with import errors

**Problem**: Module path mismatch.

**Solution**: Ensure all imports use the same module path as defined in `go.mod`. Replace `github.com/yourusername/egenskriven` with your actual module path.

### "failed to bootstrap"

**Problem**: Database initialization failed.

**Solution**:
```bash
# Remove existing data and try again
rm -rf pb_data/
./egenskriven serve
```

### Position calculation issues

**Problem**: Tasks appear in wrong order.

**Solution**: Positions use fractional indexing. If positions become too close (< 0.001), a rebalancing may be needed. This is a rare edge case after many insertions.

---

## Glossary

| Term | Definition |
|------|------------|
| **Cobra** | Go library for creating CLI applications (used by PocketBase) |
| **Task Resolver** | Component that finds tasks by ID, ID prefix, or title |
| **Fractional Indexing** | Position system using floats to avoid renumbering on inserts |
| **Idempotent** | Operation that can be repeated without different results |
| **Exit Code** | Numeric code returned by CLI indicating success/failure type |
| **History** | JSON array tracking changes to a task over time |
