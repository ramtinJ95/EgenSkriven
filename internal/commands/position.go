package commands

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"github.com/ramtinJ95/EgenSkriven/internal/position"
)

const (
	// DefaultPositionGap is the space between positions for new tasks
	// Deprecated: Use position.DefaultGap instead
	DefaultPositionGap = position.DefaultGap
	// MinPositionGap is the minimum gap before rebalancing should occur
	// Deprecated: Use position.MinGap instead
	MinPositionGap = position.MinGap
)

// GetNextPosition returns the position for a new task at the end of a column.
// If the column is empty, returns DefaultPositionGap.
// Otherwise, returns the last position + DefaultPositionGap.
func GetNextPosition(app *pocketbase.PocketBase, column string) float64 {
	return position.GetNext(app, column)
}

// GetPositionBetween calculates a position between two existing positions.
// This enables inserting tasks between existing tasks without rebalancing.
func GetPositionBetween(before, after float64) float64 {
	return position.GetBetween(before, after)
}

// GetPositionAtIndex returns the position for a task at a specific index in a column.
// index 0 = top of column
// index -1 = bottom of column (same as GetNextPosition)
func GetPositionAtIndex(app *pocketbase.PocketBase, column string, index int) float64 {
	return position.GetAtIndex(app, column, index)
}

// GetPositionAfter returns a position after a specific task.
func GetPositionAfter(app *pocketbase.PocketBase, taskID string) (float64, error) {
	return position.GetAfter(app, taskID)
}

// GetPositionBefore returns a position before a specific task.
func GetPositionBefore(app *pocketbase.PocketBase, taskID string) (float64, error) {
	return position.GetBefore(app, taskID)
}

// sortTasksByPosition sorts tasks by their position field in ascending order.
// Wrapper for backward compatibility.
func sortTasksByPosition(tasks []*core.Record) {
	position.SortByPosition(tasks)
}
