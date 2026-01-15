package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
)

// BulkOperationType represents the type of bulk operation.
type BulkOperationType int

const (
	BulkOpMove BulkOperationType = iota
	BulkOpDelete
	BulkOpSetPriority
	BulkOpSetType
	BulkOpAddLabel
)

// BulkOperation represents a bulk operation to perform.
type BulkOperation struct {
	Type    BulkOperationType
	TaskIDs []string
	Target  string // target column, priority, etc.
}

// BulkResult represents the result of a bulk operation.
type BulkResult struct {
	Operation  BulkOperationType
	SuccessIDs []string
	FailedIDs  []string
	Errors     []error
}

// BulkResultMsg is the tea.Msg for bulk operation completion.
type BulkResultMsg struct {
	Result BulkResult
}

// BulkProgressMsg reports progress during bulk operations.
type BulkProgressMsg struct {
	Current int
	Total   int
	TaskID  string
}

// BulkErrorMsg reports an error during bulk operations.
type BulkErrorMsg struct {
	Errors []error
}

// BulkMove moves multiple tasks to a target column.
func BulkMove(app *pocketbase.PocketBase, taskIDs []string, targetColumn, boardID string) tea.Cmd {
	return func() tea.Msg {
		result := BulkResult{
			Operation: BulkOpMove,
		}

		for _, id := range taskIDs {
			record, err := app.FindRecordById("tasks", id)
			if err != nil {
				result.FailedIDs = append(result.FailedIDs, id)
				result.Errors = append(result.Errors, fmt.Errorf("task %s: %w", id, err))
				continue
			}

			// Get next position in target column
			position := getNextPositionInColumn(app, targetColumn, boardID)

			record.Set("column", targetColumn)
			record.Set("position", position)

			if err := app.Save(record); err != nil {
				result.FailedIDs = append(result.FailedIDs, id)
				result.Errors = append(result.Errors, fmt.Errorf("task %s: %w", id, err))
				continue
			}

			result.SuccessIDs = append(result.SuccessIDs, id)
		}

		return BulkResultMsg{Result: result}
	}
}

// BulkDelete deletes multiple tasks.
func BulkDelete(app *pocketbase.PocketBase, taskIDs []string) tea.Cmd {
	return func() tea.Msg {
		result := BulkResult{
			Operation: BulkOpDelete,
		}

		for _, id := range taskIDs {
			record, err := app.FindRecordById("tasks", id)
			if err != nil {
				result.FailedIDs = append(result.FailedIDs, id)
				result.Errors = append(result.Errors, fmt.Errorf("task %s: %w", id, err))
				continue
			}

			if err := app.Delete(record); err != nil {
				result.FailedIDs = append(result.FailedIDs, id)
				result.Errors = append(result.Errors, fmt.Errorf("task %s: %w", id, err))
				continue
			}

			result.SuccessIDs = append(result.SuccessIDs, id)
		}

		return BulkResultMsg{Result: result}
	}
}

// getNextPositionInColumn returns the next position for a task in a column.
func getNextPositionInColumn(app *pocketbase.PocketBase, column, boardID string) float64 {
	// Find the highest position in the column
	records, err := app.FindAllRecords("tasks",
		dbx.NewExp("board = {:board} AND column = {:column}",
			dbx.Params{"board": boardID, "column": column}),
	)
	if err != nil || len(records) == 0 {
		return 1.0
	}

	maxPos := 0.0
	for _, r := range records {
		pos := r.GetFloat("position")
		if pos > maxPos {
			maxPos = pos
		}
	}

	return maxPos + 1.0
}

// FormatBulkResult formats a bulk result for display.
func FormatBulkResult(result BulkResult) string {
	successCount := len(result.SuccessIDs)
	failedCount := len(result.FailedIDs)

	var opName string
	switch result.Operation {
	case BulkOpMove:
		opName = "Moved"
	case BulkOpDelete:
		opName = "Deleted"
	case BulkOpSetPriority:
		opName = "Updated priority for"
	case BulkOpSetType:
		opName = "Updated type for"
	default:
		opName = "Processed"
	}

	if failedCount == 0 {
		if successCount == 1 {
			return fmt.Sprintf("%s 1 task", opName)
		}
		return fmt.Sprintf("%s %d tasks", opName, successCount)
	}

	return fmt.Sprintf("%s %d tasks, %d failed", opName, successCount, failedCount)
}
