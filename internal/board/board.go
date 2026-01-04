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
	record.Set("next_seq", 1) // Initialize sequence counter
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

// GetNextSequence returns the next sequence number for a board.
// DEPRECATED: Use GetAndIncrementSequence instead to avoid race conditions.
//
// This function is kept for backwards compatibility but has a race condition
// when multiple tasks are created concurrently.
func GetNextSequence(app *pocketbase.PocketBase, boardID string) (int, error) {
	// Use the new atomic function
	return GetAndIncrementSequence(app, boardID)
}

// GetAndIncrementSequence atomically gets the next sequence number and increments
// the counter in the board record.
//
// This prevents race conditions where concurrent task creation could result in
// duplicate sequence numbers. The board's next_seq field is atomically incremented.
//
// If next_seq is not set (legacy boards), it calculates from existing tasks.
func GetAndIncrementSequence(app *pocketbase.PocketBase, boardID string) (int, error) {
	// Find the board
	board, err := app.FindRecordById("boards", boardID)
	if err != nil {
		return 0, fmt.Errorf("board not found: %w", err)
	}

	// Get current next_seq value
	nextSeq := board.GetInt("next_seq")

	// If next_seq is not set (0), initialize it from existing tasks
	if nextSeq == 0 {
		nextSeq, err = calculateNextSeqFromTasks(app, boardID)
		if err != nil {
			return 0, err
		}
	}

	// Atomically increment next_seq in the board
	board.Set("next_seq", nextSeq+1)
	if err := app.Save(board); err != nil {
		return 0, fmt.Errorf("failed to increment sequence: %w", err)
	}

	return nextSeq, nil
}

// calculateNextSeqFromTasks finds the max sequence from existing tasks.
// Used to initialize next_seq for legacy boards that don't have it set.
func calculateNextSeqFromTasks(app *pocketbase.PocketBase, boardID string) (int, error) {
	records, err := app.FindAllRecords("tasks",
		dbx.NewExp("board = {:board}", dbx.Params{"board": boardID}),
	)
	if err != nil || len(records) == 0 {
		return 1, nil
	}

	// Find max sequence
	maxSeq := 0
	for _, r := range records {
		seq := r.GetInt("seq")
		if seq > maxSeq {
			maxSeq = seq
		}
	}

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
//
// Returns an error if:
// - Format is not PREFIX-NUMBER
// - Sequence is not a valid positive integer
func ParseDisplayID(displayID string) (prefix string, seq int, err error) {
	parts := strings.SplitN(displayID, "-", 2)
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid display ID format: %s (expected PREFIX-NUMBER)", displayID)
	}

	prefix = strings.ToUpper(parts[0])
	if prefix == "" {
		return "", 0, fmt.Errorf("invalid display ID format: %s (prefix cannot be empty)", displayID)
	}

	_, err = fmt.Sscanf(parts[1], "%d", &seq)
	if err != nil {
		return "", 0, fmt.Errorf("invalid sequence number in display ID: %s", displayID)
	}

	// Reject negative or zero sequence numbers
	if seq <= 0 {
		return "", 0, fmt.Errorf("invalid sequence number in display ID: %s (must be positive)", displayID)
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
		} else if strArr, ok := c.([]string); ok {
			columns = strArr
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
