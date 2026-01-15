package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatBulkResult_MoveSuccess(t *testing.T) {
	result := BulkResult{
		Operation:  BulkOpMove,
		SuccessIDs: []string{"task-1", "task-2", "task-3"},
	}

	formatted := FormatBulkResult(result)
	assert.Contains(t, formatted, "Moved")
	assert.Contains(t, formatted, "3 tasks")
}

func TestFormatBulkResult_MoveSingleSuccess(t *testing.T) {
	result := BulkResult{
		Operation:  BulkOpMove,
		SuccessIDs: []string{"task-1"},
	}

	formatted := FormatBulkResult(result)
	assert.Contains(t, formatted, "Moved")
	assert.Contains(t, formatted, "1 task")
}

func TestFormatBulkResult_DeleteSuccess(t *testing.T) {
	result := BulkResult{
		Operation:  BulkOpDelete,
		SuccessIDs: []string{"task-1", "task-2"},
	}

	formatted := FormatBulkResult(result)
	assert.Contains(t, formatted, "Deleted")
	assert.Contains(t, formatted, "2 tasks")
}

func TestFormatBulkResult_WithFailures(t *testing.T) {
	result := BulkResult{
		Operation:  BulkOpMove,
		SuccessIDs: []string{"task-1", "task-2"},
		FailedIDs:  []string{"task-3"},
	}

	formatted := FormatBulkResult(result)
	assert.Contains(t, formatted, "2 tasks")
	assert.Contains(t, formatted, "1 failed")
}

func TestFormatBulkResult_PriorityUpdate(t *testing.T) {
	result := BulkResult{
		Operation:  BulkOpSetPriority,
		SuccessIDs: []string{"task-1"},
	}

	formatted := FormatBulkResult(result)
	assert.Contains(t, formatted, "Updated priority")
}

func TestBulkResultMsg(t *testing.T) {
	result := BulkResult{
		Operation:  BulkOpDelete,
		SuccessIDs: []string{"task-1", "task-2"},
	}

	msg := BulkResultMsg{Result: result}
	assert.Equal(t, BulkOpDelete, msg.Result.Operation)
	assert.Len(t, msg.Result.SuccessIDs, 2)
}
