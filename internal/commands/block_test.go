package commands

import (
	"encoding/json"
	"testing"

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

// setupTasksCollectionWithNeedInput creates tasks collection with need_input column for block tests
func setupTasksCollectionWithNeedInput(t *testing.T, app *pocketbase.PocketBase) {
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
		Values:   []string{"backlog", "todo", "in_progress", "need_input", "review", "done"},
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

// setupCommentsCollection creates comments collection for block tests
func setupCommentsCollection(t *testing.T, app *pocketbase.PocketBase) {
	t.Helper()

	_, err := app.FindCollectionByNameOrId("comments")
	if err == nil {
		return
	}

	collection := core.NewBaseCollection("comments")
	collection.Fields.Add(&core.TextField{Name: "task", Required: true})
	collection.Fields.Add(&core.TextField{Name: "content", Required: true})
	collection.Fields.Add(&core.SelectField{
		Name:     "author_type",
		Required: true,
		Values:   []string{"human", "agent"},
	})
	collection.Fields.Add(&core.TextField{Name: "author_id"})
	collection.Fields.Add(&core.JSONField{Name: "metadata"})

	if err := app.Save(collection); err != nil {
		t.Fatalf("failed to create comments collection: %v", err)
	}
}

// createBlockTestTask creates a task for block command testing
func createBlockTestTask(t *testing.T, app *pocketbase.PocketBase, title string, column string) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("tasks")
	require.NoError(t, err)

	record := core.NewRecord(collection)
	record.Set("title", title)
	record.Set("type", "feature")
	record.Set("priority", "medium")
	record.Set("column", column)
	record.Set("position", 1000.0)
	record.Set("labels", []string{})
	record.Set("blocked_by", []string{})
	record.Set("created_by", "cli")
	record.Set("history", []map[string]any{})

	require.NoError(t, app.Save(record))
	return record
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
	setupTasksCollectionWithNeedInput(t, app)
	setupCommentsCollection(t, app)

	// Create a test task in todo
	task := createBlockTestTask(t, app, "Test Task", "todo")
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
	setupTasksCollectionWithNeedInput(t, app)
	setupCommentsCollection(t, app)

	// Create a test task in in_progress
	task := createBlockTestTask(t, app, "In Progress Task", "in_progress")

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
	setupTasksCollectionWithNeedInput(t, app)
	setupCommentsCollection(t, app)

	// Create a test task in review column
	task := createBlockTestTask(t, app, "Review Task", "review")

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
	setupTasksCollectionWithNeedInput(t, app)
	setupCommentsCollection(t, app)

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
			task := createBlockTestTask(t, app, "Test for "+tt.name, "todo")

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
	setupTasksCollectionWithNeedInput(t, app)
	setupCommentsCollection(t, app)

	// Step 1: Create a task in todo state
	task := createBlockTestTask(t, app, "Implement authentication", "todo")
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
		"task = '"+task.Id+"'",
		"", // No sorting in tests - collection lacks autodate fields
		0, 0,
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
		"task = '"+task.Id+"'",
		"", // No sorting in tests - collection lacks autodate fields
		0, 0,
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
	setupTasksCollectionWithNeedInput(t, app)
	// Intentionally NOT setting up comments collection to simulate failure

	// Create a task
	task := createBlockTestTask(t, app, "Test atomic rollback", "in_progress")
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
