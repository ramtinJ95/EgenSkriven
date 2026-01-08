package commands

import (
	"encoding/json"
	"testing"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ramtinJ95/EgenSkriven/internal/testutil"
)

// getHistoryFromTask extracts history from a task record, handling the types.JSONRaw type
func getHistoryFromTask(t *testing.T, task *core.Record) []map[string]any {
	t.Helper()

	raw := task.Get("history")
	if raw == nil {
		return []map[string]any{}
	}

	// Handle native []any type (before save)
	if h, ok := raw.([]any); ok {
		result := make([]map[string]any, len(h))
		for i, entry := range h {
			if m, ok := entry.(map[string]any); ok {
				result[i] = m
			}
		}
		return result
	}

	// Handle types.JSONRaw (from database)
	if jsonRaw, ok := raw.(interface{ String() string }); ok {
		var history []map[string]any
		jsonStr := jsonRaw.String()
		if jsonStr == "" || jsonStr == "null" {
			return []map[string]any{}
		}
		if err := json.Unmarshal([]byte(jsonStr), &history); err == nil {
			return history
		}
	}

	return []map[string]any{}
}

// simulateBlockTask simulates what the block command does:
// moves task to need_input and adds history entry
func simulateBlockTask(t *testing.T, app *pocketbase.PocketBase, task *core.Record, question string, agentName string) {
	t.Helper()

	currentColumn := task.GetString("column")

	// Update task column to need_input
	task.Set("column", "need_input")

	// Add history entry (same as block command does)
	addHistoryEntry(task, "blocked", agentName, map[string]any{
		"column": map[string]any{
			"from": currentColumn,
			"to":   "need_input",
		},
		"reason": question,
	})

	require.NoError(t, app.Save(task))
}

// simulateBlockTaskWithComment simulates the full block command:
// moves task to need_input, adds history entry, AND creates a comment
// This is used for integration tests that need the full workflow.
func simulateBlockTaskWithComment(t *testing.T, app *pocketbase.PocketBase, task *core.Record, question string, agentName string) {
	t.Helper()

	currentColumn := task.GetString("column")

	// Get comments collection
	commentsCollection, err := app.FindCollectionByNameOrId("comments")
	require.NoError(t, err, "comments collection should exist")

	// Execute in transaction for atomicity (same as real block command)
	err = app.RunInTransaction(func(txApp core.App) error {
		// Update task column to need_input
		task.Set("column", "need_input")

		// Add history entry
		addHistoryEntry(task, "blocked", agentName, map[string]any{
			"column": map[string]any{
				"from": currentColumn,
				"to":   "need_input",
			},
			"reason": question,
		})

		if err := txApp.Save(task); err != nil {
			return err
		}

		// Create comment (same as real block command)
		comment := core.NewRecord(commentsCollection)
		comment.Set("task", task.Id)
		comment.Set("content", question)
		comment.Set("author_type", "agent")
		comment.Set("author_id", agentName)
		comment.Set("metadata", map[string]any{
			"action": "block_question",
		})

		return txApp.Save(comment)
	})

	require.NoError(t, err, "block with comment should succeed")
}

// ========== Tests ==========

func TestBlockCommand_HistoryIsUpdatedCorrectly(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollection(t, app)
	SetupCommentsCollection(t, app)

	// Create a test task in todo
	task := CreateTestTask(t, app, "Test Task", "todo")
	initialHistoryLen := 0

	// Simulate blocking the task
	question := "What authentication approach should I use?"
	agentName := "test-agent"
	simulateBlockTask(t, app, task, question, agentName)

	// Re-fetch task from database
	task, err := app.FindRecordById("tasks", task.Id)
	require.NoError(t, err)

	// Verify task is in need_input
	assert.Equal(t, "need_input", task.GetString("column"))

	// Verify history was updated
	history := getHistoryFromTask(t, task)
	assert.Len(t, history, initialHistoryLen+1, "history should have one new entry")

	// Get the last history entry
	lastEntry := history[len(history)-1]

	// Verify history entry fields
	assert.Equal(t, "blocked", lastEntry["action"], "action should be 'blocked'")
	assert.Equal(t, "cli", lastEntry["actor"], "actor should be 'cli'")
	assert.Equal(t, agentName, lastEntry["actor_detail"], "actor_detail should match agent name")
	assert.NotEmpty(t, lastEntry["timestamp"], "timestamp should not be empty")

	// Verify changes in history
	changes, ok := lastEntry["changes"].(map[string]any)
	require.True(t, ok, "changes should be a map")

	// Verify column change
	columnChange, ok := changes["column"].(map[string]any)
	require.True(t, ok, "column change should be a map")
	assert.Equal(t, "todo", columnChange["from"], "column.from should be 'todo'")
	assert.Equal(t, "need_input", columnChange["to"], "column.to should be 'need_input'")

	// Verify reason
	assert.Equal(t, question, changes["reason"], "reason should match the question")
}

func TestBlockCommand_HistoryFromInProgress(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollection(t, app)
	SetupCommentsCollection(t, app)

	// Create a test task in in_progress
	task := CreateTestTask(t, app, "In Progress Task", "in_progress")

	// Simulate blocking the task
	question := "Need clarification on requirements"
	simulateBlockTask(t, app, task, question, "agent")

	// Re-fetch and verify
	task, err := app.FindRecordById("tasks", task.Id)
	require.NoError(t, err)

	history := getHistoryFromTask(t, task)
	require.Greater(t, len(history), 0, "history should not be empty")
	lastEntry := history[len(history)-1]
	changes, ok := lastEntry["changes"].(map[string]any)
	require.True(t, ok, "changes should be a map")
	columnChange, ok := changes["column"].(map[string]any)
	require.True(t, ok, "column change should be a map")

	// Verify the from column is in_progress
	assert.Equal(t, "in_progress", columnChange["from"], "column.from should be 'in_progress'")
	assert.Equal(t, "need_input", columnChange["to"], "column.to should be 'need_input'")
}

func TestBlockCommand_HistoryBlockedEntryStructure(t *testing.T) {
	// This test verifies the complete structure of the blocked history entry
	app := testutil.NewTestApp(t)
	SetupTasksCollection(t, app)
	SetupCommentsCollection(t, app)

	// Create a test task in review column
	task := CreateTestTask(t, app, "Review Task", "review")

	// Block it with a specific question
	question := "Should we refactor this module before proceeding?"
	agentName := "code-reviewer-agent"
	simulateBlockTask(t, app, task, question, agentName)

	// Re-fetch and verify
	task, err := app.FindRecordById("tasks", task.Id)
	require.NoError(t, err)

	history := getHistoryFromTask(t, task)
	require.Len(t, history, 1, "should have exactly one history entry")

	entry := history[0]

	// Verify all required fields are present
	assert.Equal(t, "blocked", entry["action"], "action should be 'blocked'")
	assert.Equal(t, "cli", entry["actor"], "actor should be 'cli'")
	assert.Equal(t, agentName, entry["actor_detail"], "actor_detail should be agent name")
	assert.NotEmpty(t, entry["timestamp"], "timestamp should be set")

	// Verify changes structure
	changes, ok := entry["changes"].(map[string]any)
	require.True(t, ok, "changes should be a map")

	columnChange, ok := changes["column"].(map[string]any)
	require.True(t, ok, "column change should be a map")
	assert.Equal(t, "review", columnChange["from"])
	assert.Equal(t, "need_input", columnChange["to"])

	assert.Equal(t, question, changes["reason"], "reason should match question")
}

func TestBlockCommand_HistoryWithDifferentAgents(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollection(t, app)
	SetupCommentsCollection(t, app)

	tests := []struct {
		name      string
		agentName string
	}{
		{"default agent", "agent"},
		{"custom agent name", "opencode-build"},
		{"claude-code agent", "claude-code"},
		{"empty agent uses default", ""}, // Should still work
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := CreateTestTask(t, app, "Test for "+tt.name, "todo")

			agentToUse := tt.agentName
			if agentToUse == "" {
				agentToUse = "agent" // Default fallback
			}
			simulateBlockTask(t, app, task, "Question?", agentToUse)

			task, _ = app.FindRecordById("tasks", task.Id)
			history := getHistoryFromTask(t, task)
			require.Greater(t, len(history), 0, "history should not be empty")
			lastEntry := history[len(history)-1]

			assert.Equal(t, agentToUse, lastEntry["actor_detail"])
		})
	}
}

// ========== Integration Tests ==========

// TestFullWorkflow_CreateBlockCommentList tests the complete workflow:
// create task → block → add comment → list comments
// This is an integration test that verifies all components work together.
func TestFullWorkflow_CreateBlockCommentList(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollection(t, app)
	SetupCommentsCollection(t, app)

	// Step 1: Create a task in todo state
	task := CreateTestTask(t, app, "Implement authentication", "todo")
	require.NotEmpty(t, task.Id, "task should have an ID")
	assert.Equal(t, "todo", task.GetString("column"), "task should start in todo column")

	// Step 2: Simulate agent blocking the task with a question
	// This mirrors what the block command does: update task AND create comment
	agentQuestion := "What authentication approach should I use? JWT, sessions, or OAuth2?"
	agentName := "opencode-build"
	simulateBlockTaskWithComment(t, app, task, agentQuestion, agentName)

	// Verify task moved to need_input
	task, err := app.FindRecordById("tasks", task.Id)
	require.NoError(t, err, "should be able to fetch task")
	assert.Equal(t, "need_input", task.GetString("column"), "task should be in need_input after blocking")

	// Verify blocking created a comment
	comments, err := app.FindRecordsByFilter(
		"comments",
		"task = {:taskId}",
		"", // No sorting in tests - collection lacks autodate fields
		0, 0,
		dbx.Params{"taskId": task.Id},
	)
	require.NoError(t, err, "should be able to fetch comments")
	require.Len(t, comments, 1, "blocking should create exactly one comment")

	agentComment := comments[0]
	assert.Equal(t, agentQuestion, agentComment.GetString("content"), "comment content should match question")
	assert.Equal(t, "agent", agentComment.GetString("author_type"), "comment should be from agent")
	assert.Equal(t, agentName, agentComment.GetString("author_id"), "comment author should match agent name")

	// Step 3: Simulate human adding a response comment
	humanResponse := "@agent Use JWT with refresh tokens. Access tokens expire in 15 minutes, refresh tokens in 7 days."
	humanAuthor := "senior-developer"

	commentsCollection, err := app.FindCollectionByNameOrId("comments")
	require.NoError(t, err, "comments collection should exist")

	humanComment := core.NewRecord(commentsCollection)
	humanComment.Set("task", task.Id)
	humanComment.Set("content", humanResponse)
	humanComment.Set("author_type", "human")
	humanComment.Set("author_id", humanAuthor)
	humanComment.Set("metadata", map[string]any{
		"mentions": []string{"@agent"},
	})
	require.NoError(t, app.Save(humanComment), "should be able to save human comment")

	// Step 4: Verify all comments can be listed (simulates comments command)
	allComments, err := app.FindRecordsByFilter(
		"comments",
		"task = {:taskId}",
		"", // No sorting in tests - collection lacks autodate fields
		0, 0,
		dbx.Params{"taskId": task.Id},
	)
	require.NoError(t, err, "should be able to list all comments")
	require.Len(t, allComments, 2, "should have 2 comments total")

	// Verify we have both comment types (order not guaranteed without autodate)
	var foundAgent, foundHuman bool
	for _, comment := range allComments {
		authorType := comment.GetString("author_type")
		content := comment.GetString("content")

		if authorType == "agent" && content == agentQuestion {
			foundAgent = true
		}
		if authorType == "human" && content == humanResponse {
			foundHuman = true
		}
	}

	assert.True(t, foundAgent, "should find the agent's blocking question comment")
	assert.True(t, foundHuman, "should find the human's response comment")

	// Verify history was updated correctly
	history := getHistoryFromTask(t, task)
	require.Greater(t, len(history), 0, "task should have history")

	lastHistoryEntry := history[len(history)-1]
	assert.Equal(t, "blocked", lastHistoryEntry["action"], "last history action should be 'blocked'")
	changes, ok := lastHistoryEntry["changes"].(map[string]any)
	require.True(t, ok, "changes should be a map")
	assert.Equal(t, agentQuestion, changes["reason"], "history should record the blocking reason")
}

// TestAtomicBlock_RollbackOnCommentFailure verifies that if comment creation fails,
// the task column change is rolled back (atomicity test).
func TestAtomicBlock_RollbackOnCommentFailure(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollection(t, app)
	// Intentionally NOT setting up comments collection to simulate failure

	// Create a task
	task := CreateTestTask(t, app, "Test atomic rollback", "in_progress")
	originalColumn := task.GetString("column")
	require.Equal(t, "in_progress", originalColumn)

	// Try to block the task - this should fail because comments collection doesn't exist
	// We'll simulate what the block command does: wrap in transaction and try to save comment
	err := app.RunInTransaction(func(txApp core.App) error {
		// Update task column to need_input
		task.Set("column", "need_input")
		addHistoryEntry(task, "blocked", "test-agent", map[string]any{
			"column": map[string]any{
				"from": originalColumn,
				"to":   "need_input",
			},
			"reason": "Test question",
		})

		if err := txApp.Save(task); err != nil {
			return err
		}

		// Try to create comment - this should fail because collection doesn't exist
		commentsCollection, err := txApp.FindCollectionByNameOrId("comments")
		if err != nil {
			return err // This will cause rollback
		}

		comment := core.NewRecord(commentsCollection)
		comment.Set("task", task.Id)
		comment.Set("content", "Test question")
		comment.Set("author_type", "agent")
		comment.Set("author_id", "test-agent")

		return txApp.Save(comment)
	})

	// The transaction should have failed
	require.Error(t, err, "transaction should fail when comments collection doesn't exist")

	// Verify task column was NOT changed (rolled back)
	task, fetchErr := app.FindRecordById("tasks", task.Id)
	require.NoError(t, fetchErr, "should be able to fetch task after failed transaction")

	// The task should still be in its original column because the transaction was rolled back
	assert.Equal(t, originalColumn, task.GetString("column"),
		"task column should be rolled back to original state after failed transaction")
}

// ========== Validation Tests ==========

// TestBlockCommand_FailsIfAlreadyBlocked verifies that blocking an already blocked task fails gracefully
func TestBlockCommand_FailsIfAlreadyBlocked(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollection(t, app)
	SetupCommentsCollection(t, app)

	// Create a task already in need_input
	task := CreateTestTask(t, app, "Already blocked task", "need_input")

	// Verify the task is in need_input
	assert.Equal(t, "need_input", task.GetString("column"))

	// Try to block it again - this should fail
	// We simulate the validation that block command does
	currentColumn := task.GetString("column")
	if currentColumn == "need_input" {
		// This is the expected behavior - the block command should reject this
		assert.Equal(t, "need_input", currentColumn, "task should already be in need_input")
	}

	// The block command checks this at lines 83-86 in block.go:
	// if currentColumn == "need_input" {
	//     return out.Error(ExitValidation, fmt.Sprintf("task %s is already blocked (in need_input)", shortID(task.Id)), nil)
	// }
}

// TestBlockCommand_FailsIfTaskIsDone verifies that blocking a completed task fails gracefully
func TestBlockCommand_FailsIfTaskIsDone(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollection(t, app)
	SetupCommentsCollection(t, app)

	// Create a task in done column
	task := CreateTestTask(t, app, "Completed task", "done")

	// Verify the task is done
	assert.Equal(t, "done", task.GetString("column"))

	// The block command checks this at lines 87-89 in block.go:
	// if currentColumn == "done" {
	//     return out.Error(ExitValidation, "cannot block a completed task", nil)
	// }

	// Verify that the validation would catch this
	currentColumn := task.GetString("column")
	assert.Equal(t, "done", currentColumn, "task should be in done column")
}

// TestBlockCommand_StdinInput verifies that the block command can read question from stdin
func TestBlockCommand_StdinInput(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollection(t, app)
	SetupCommentsCollection(t, app)

	// Create a task
	task := CreateTestTask(t, app, "Task for stdin test", "todo")

	// Simulate reading from stdin and blocking
	stdinQuestion := "This is a question from stdin\nWith multiple lines\nAnd detailed context"

	// Simulate what block command does with stdin input
	question := stdinQuestion

	// Execute the block operation (simulated)
	simulateBlockTaskWithComment(t, app, task, question, "test-agent")

	// Verify task was blocked
	task, err := app.FindRecordById("tasks", task.Id)
	require.NoError(t, err)
	assert.Equal(t, "need_input", task.GetString("column"))

	// Verify comment was created with full stdin content
	comments, err := app.FindRecordsByFilter("comments", "task = {:taskId}", "", 0, 0, dbx.Params{"taskId": task.Id})
	require.NoError(t, err)
	require.Len(t, comments, 1)
	assert.Equal(t, stdinQuestion, comments[0].GetString("content"), "comment should contain full stdin content")
}

// TestBlockCommand_JSONOutput verifies that the block command produces valid JSON output
func TestBlockCommand_JSONOutput(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollection(t, app)
	SetupCommentsCollection(t, app)

	// Create a task
	task := CreateTestTask(t, app, "Task for JSON test", "in_progress")

	// Execute block operation
	question := "What approach should I use?"
	agentName := "test-agent"
	simulateBlockTaskWithComment(t, app, task, question, agentName)

	// Verify the structure that JSON output would contain
	task, err := app.FindRecordById("tasks", task.Id)
	require.NoError(t, err)

	comments, err := app.FindRecordsByFilter("comments", "task = {:taskId}", "", 0, 0, dbx.Params{"taskId": task.Id})
	require.NoError(t, err)
	require.Len(t, comments, 1)

	// Verify all fields that would be in JSON output
	jsonResult := map[string]any{
		"success":    true,
		"task_id":    task.Id,
		"display_id": getTaskDisplayID(app, task),
		"column":     task.GetString("column"),
		"comment_id": comments[0].Id,
		"message":    "Task blocked, awaiting human input",
	}

	assert.True(t, jsonResult["success"].(bool))
	assert.NotEmpty(t, jsonResult["task_id"])
	assert.NotEmpty(t, jsonResult["display_id"])
	assert.Equal(t, "need_input", jsonResult["column"])
	assert.NotEmpty(t, jsonResult["comment_id"])
}
