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
