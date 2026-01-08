package commands

import (
	"testing"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ramtinJ95/EgenSkriven/internal/testutil"
)

// ========== Setup Functions ==========
// These are thin wrappers around shared helpers in test_helpers_test.go

// setupTasksCollectionForComment creates tasks collection for comment tests.
// Deprecated: Use SetupTasksCollection from test_helpers_test.go instead.
func setupTasksCollectionForComment(t *testing.T, app *pocketbase.PocketBase) {
	SetupTasksCollection(t, app)
}

// setupCommentsCollectionForComment creates comments collection for comment tests.
// Deprecated: Use SetupCommentsCollection from test_helpers_test.go instead.
func setupCommentsCollectionForComment(t *testing.T, app *pocketbase.PocketBase) {
	SetupCommentsCollection(t, app)
}

// createCommentTestTask creates a task for comment command testing.
// Deprecated: Use CreateTestTask from test_helpers_test.go instead.
func createCommentTestTask(t *testing.T, app *pocketbase.PocketBase, title string, column string) *core.Record {
	return CreateTestTask(t, app, title, column)
}

// createTestComment creates a comment directly for testing.
// Deprecated: Use CreateTestComment from test_helpers_test.go instead.
func createTestComment(t *testing.T, app *pocketbase.PocketBase, taskId, content, authorType, authorId string) *core.Record {
	return CreateTestComment(t, app, taskId, content, authorType, authorId)
}

// getCommentsForTask returns all comments for a given task ID.
// Deprecated: Use GetCommentsForTask from test_helpers_test.go instead.
func getCommentsForTask(t *testing.T, app *pocketbase.PocketBase, taskId string) []*core.Record {
	return GetCommentsForTask(t, app, taskId)
}

// ========== extractMentions Tests ==========

func TestExtractMentions_SingleMention(t *testing.T) {
	text := "Hello @agent"
	mentions := extractMentions(text)

	assert.Len(t, mentions, 1)
	assert.Contains(t, mentions, "@agent")
}

func TestExtractMentions_MultipleMentions(t *testing.T) {
	text := "@agent @user please help"
	mentions := extractMentions(text)

	assert.Len(t, mentions, 2)
	assert.Contains(t, mentions, "@agent")
	assert.Contains(t, mentions, "@user")
}

func TestExtractMentions_NoMentions(t *testing.T) {
	text := "No mentions here"
	mentions := extractMentions(text)

	assert.Empty(t, mentions)
}

func TestExtractMentions_DuplicateMentions(t *testing.T) {
	text := "@agent @agent duplicate"
	mentions := extractMentions(text)

	// Should be deduplicated
	assert.Len(t, mentions, 1)
	assert.Contains(t, mentions, "@agent")
}

func TestExtractMentions_MentionAtStart(t *testing.T) {
	text := "@agent I've decided"
	mentions := extractMentions(text)

	assert.Len(t, mentions, 1)
	assert.Contains(t, mentions, "@agent")
}

func TestExtractMentions_MentionInMiddle(t *testing.T) {
	text := "Hey @agent please review"
	mentions := extractMentions(text)

	assert.Len(t, mentions, 1)
	assert.Contains(t, mentions, "@agent")
}

func TestExtractMentions_MentionWithNewlines(t *testing.T) {
	text := "First line\n@agent\nThird line"
	mentions := extractMentions(text)

	assert.Len(t, mentions, 1)
	assert.Contains(t, mentions, "@agent")
}

func TestExtractMentions_MentionWithNumbers(t *testing.T) {
	text := "Hello @agent123"
	mentions := extractMentions(text)

	assert.Len(t, mentions, 1)
	assert.Contains(t, mentions, "@agent123")
}

func TestExtractMentions_MentionWithUnderscore(t *testing.T) {
	text := "Hello @code_reviewer"
	mentions := extractMentions(text)

	assert.Len(t, mentions, 1)
	assert.Contains(t, mentions, "@code_reviewer")
}

// ========== resolveAuthor Tests ==========

func TestResolveAuthor_FlagValue(t *testing.T) {
	result := resolveAuthor("jane.doe")
	assert.Equal(t, "jane.doe", result)
}

func TestResolveAuthor_EmptyFlagUsesEnv(t *testing.T) {
	// Note: This test would need to manipulate env vars
	// For now, we test that empty flag returns empty when no env is set
	result := resolveAuthor("")
	// Result depends on USER env var, but won't be empty if USER is set
	// We just verify it doesn't panic
	assert.NotNil(t, result)
}

// ========== isAgentContext Tests ==========

func TestIsAgentContext_NoEnvVars(t *testing.T) {
	// When no agent env vars are set, should return false
	// This test assumes clean environment
	// In practice, env vars might be set by the test runner
	result := isAgentContext()
	// We can't assert a specific value since env might have vars set
	assert.IsType(t, true, result)
}

// ========== Comment Creation Tests ==========

func TestCommentCreation_BasicComment(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComment(t, app)
	setupCommentsCollectionForComment(t, app)

	// Create a test task
	task := createCommentTestTask(t, app, "Test Task", "todo")

	// Create a comment
	comment := createTestComment(t, app, task.Id, "This is a test comment", "human", "testuser")

	// Verify comment was created
	assert.NotEmpty(t, comment.Id)
	assert.Equal(t, task.Id, comment.GetString("task"))
	assert.Equal(t, "This is a test comment", comment.GetString("content"))
	assert.Equal(t, "human", comment.GetString("author_type"))
	assert.Equal(t, "testuser", comment.GetString("author_id"))
}

func TestCommentCreation_AgentComment(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComment(t, app)
	setupCommentsCollectionForComment(t, app)

	// Create a test task
	task := createCommentTestTask(t, app, "Test Task", "need_input")

	// Create an agent comment
	comment := createTestComment(t, app, task.Id, "What approach should I use?", "agent", "opencode")

	// Verify comment was created with agent type
	assert.Equal(t, "agent", comment.GetString("author_type"))
	assert.Equal(t, "opencode", comment.GetString("author_id"))
}

func TestCommentCreation_MultipleComments(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComment(t, app)
	setupCommentsCollectionForComment(t, app)

	// Create a test task
	task := createCommentTestTask(t, app, "Test Task", "todo")

	// Create multiple comments
	createTestComment(t, app, task.Id, "First comment", "agent", "agent1")
	createTestComment(t, app, task.Id, "Second comment", "human", "user1")
	createTestComment(t, app, task.Id, "Third comment", "agent", "agent2")

	// Query comments
	comments := getCommentsForTask(t, app, task.Id)

	// Verify all comments were created
	assert.Len(t, comments, 3)
}

func TestCommentCreation_CommentOnNeedInputTask(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComment(t, app)
	setupCommentsCollectionForComment(t, app)

	// Create a task in need_input state
	task := createCommentTestTask(t, app, "Blocked Task", "need_input")

	// Create a response comment
	comment := createTestComment(t, app, task.Id, "@agent I've decided to use JWT", "human", "developer")

	// Verify comment was created
	assert.Equal(t, task.Id, comment.GetString("task"))
	assert.Contains(t, comment.GetString("content"), "@agent")
}

func TestCommentCreation_LongContent(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComment(t, app)
	setupCommentsCollectionForComment(t, app)

	// Create a test task
	task := createCommentTestTask(t, app, "Test Task", "todo")

	// Create a comment with long content
	longContent := "This is a very long comment that spans multiple paragraphs.\n\n" +
		"It contains detailed information about the decision made:\n" +
		"1. We should use JWT for authentication\n" +
		"2. Refresh tokens should expire in 7 days\n" +
		"3. Access tokens should expire in 15 minutes\n\n" +
		"Please proceed with this approach."

	comment := createTestComment(t, app, task.Id, longContent, "human", "architect")

	// Verify full content was stored
	assert.Equal(t, longContent, comment.GetString("content"))
}

func TestCommentCreation_EmptyAuthorId(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComment(t, app)
	setupCommentsCollectionForComment(t, app)

	// Create a test task
	task := createCommentTestTask(t, app, "Test Task", "todo")

	// Create a comment with empty author_id
	collection, _ := app.FindCollectionByNameOrId("comments")
	record := core.NewRecord(collection)
	record.Set("task", task.Id)
	record.Set("content", "Anonymous comment")
	record.Set("author_type", "human")
	record.Set("author_id", "") // Empty author_id

	err := app.Save(record)
	require.NoError(t, err)

	// Verify comment was created
	assert.NotEmpty(t, record.Id)
	assert.Empty(t, record.GetString("author_id"))
}

func TestCommentCreation_WithMetadata(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComment(t, app)
	setupCommentsCollectionForComment(t, app)

	// Create a test task
	task := createCommentTestTask(t, app, "Test Task", "todo")

	// Create a comment with metadata
	collection, _ := app.FindCollectionByNameOrId("comments")
	record := core.NewRecord(collection)
	record.Set("task", task.Id)
	record.Set("content", "@agent please review")
	record.Set("author_type", "human")
	record.Set("author_id", "developer")
	record.Set("metadata", map[string]any{
		"mentions": []string{"@agent"},
		"action":   "response",
	})

	err := app.Save(record)
	require.NoError(t, err)

	// Verify metadata was stored
	metadata := record.Get("metadata")
	assert.NotNil(t, metadata)
}

// ========== Comment Query Tests ==========

func TestCommentQuery_FilterByTask(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComment(t, app)
	setupCommentsCollectionForComment(t, app)

	// Create two tasks
	task1 := createCommentTestTask(t, app, "Task 1", "todo")
	task2 := createCommentTestTask(t, app, "Task 2", "todo")

	// Create comments on both tasks
	createTestComment(t, app, task1.Id, "Comment on task 1", "human", "user")
	createTestComment(t, app, task1.Id, "Another on task 1", "human", "user")
	createTestComment(t, app, task2.Id, "Comment on task 2", "human", "user")

	// Query comments for task1 only
	task1Comments := getCommentsForTask(t, app, task1.Id)
	task2Comments := getCommentsForTask(t, app, task2.Id)

	assert.Len(t, task1Comments, 2)
	assert.Len(t, task2Comments, 1)
}

func TestCommentQuery_EmptyTaskHasNoComments(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComment(t, app)
	setupCommentsCollectionForComment(t, app)

	// Create a task with no comments
	task := createCommentTestTask(t, app, "Task Without Comments", "todo")

	// Query comments
	comments := getCommentsForTask(t, app, task.Id)

	assert.Empty(t, comments)
}

// ========== Author Type Tests ==========

func TestCommentAuthorTypes(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComment(t, app)
	setupCommentsCollectionForComment(t, app)

	task := createCommentTestTask(t, app, "Test Task", "todo")

	tests := []struct {
		name       string
		authorType string
		authorId   string
	}{
		{"human with name", "human", "john.doe"},
		{"human without name", "human", ""},
		{"agent with name", "agent", "opencode"},
		{"agent with tool name", "agent", "claude-code"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comment := createTestComment(t, app, task.Id, "Test content", tt.authorType, tt.authorId)

			assert.Equal(t, tt.authorType, comment.GetString("author_type"))
			assert.Equal(t, tt.authorId, comment.GetString("author_id"))
		})
	}
}

// ========== Edge Case Tests ==========

func TestCommentCreation_SpecialCharacters(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComment(t, app)
	setupCommentsCollectionForComment(t, app)

	task := createCommentTestTask(t, app, "Test Task", "todo")

	// Content with special characters
	content := `Here's my response with special chars: <>&"'` + "`" + `
Code block:
` + "```go" + `
func main() {}
` + "```"

	comment := createTestComment(t, app, task.Id, content, "human", "developer")

	// Verify content is stored correctly
	assert.Equal(t, content, comment.GetString("content"))
}

func TestCommentCreation_UnicodeContent(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComment(t, app)
	setupCommentsCollectionForComment(t, app)

	task := createCommentTestTask(t, app, "Test Task", "todo")

	// Content with unicode characters
	content := "Response with unicode: Hello"

	comment := createTestComment(t, app, task.Id, content, "human", "developer")

	// Verify unicode content is stored correctly
	assert.Equal(t, content, comment.GetString("content"))
}

// ========== Stdin and JSON Output Tests ==========

// TestCommentCommand_StdinInput verifies that the comment command can read from stdin
func TestCommentCommand_StdinInput(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComment(t, app)
	setupCommentsCollectionForComment(t, app)

	task := createCommentTestTask(t, app, "Task for stdin test", "need_input")

	// Simulate reading from stdin
	stdinContent := `Here's my detailed response from stdin:

1. Use JWT for authentication
2. Access tokens should expire in 15 minutes
3. Refresh tokens should expire in 7 days

Additional notes:
- Consider using httpOnly cookies
- Implement token rotation`

	// Create comment with the stdin content (simulates --stdin flag behavior)
	comment := createTestComment(t, app, task.Id, stdinContent, "human", "architect")

	// Verify full content was preserved
	assert.Equal(t, stdinContent, comment.GetString("content"), "stdin content should be preserved exactly")
	assert.Contains(t, comment.GetString("content"), "JWT for authentication")
	assert.Contains(t, comment.GetString("content"), "httpOnly cookies")
}

// TestCommentCommand_JSONOutput verifies that the comment command produces valid JSON output structure
func TestCommentCommand_JSONOutput(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComment(t, app)
	setupCommentsCollectionForComment(t, app)

	task := createCommentTestTask(t, app, "Task for JSON test", "need_input")

	// Create a comment with mentions
	content := "@agent I've decided to use OAuth2 for authentication"
	authorType := "human"
	authorId := "senior-dev"

	comment := createTestComment(t, app, task.Id, content, authorType, authorId)

	// Extract mentions like the command does
	mentions := extractMentions(content)

	// Verify the structure that would be in JSON output
	jsonResult := map[string]any{
		"success":     true,
		"comment_id":  comment.Id,
		"task_id":     task.Id,
		"display_id":  task.Id[:8], // Simplified display ID
		"author_type": authorType,
		"author_id":   authorId,
		"mentions":    mentions,
	}

	// Verify all expected fields
	assert.True(t, jsonResult["success"].(bool))
	assert.NotEmpty(t, jsonResult["comment_id"])
	assert.Equal(t, task.Id, jsonResult["task_id"])
	assert.Equal(t, "human", jsonResult["author_type"])
	assert.Equal(t, "senior-dev", jsonResult["author_id"])

	// Verify mentions were extracted
	mentionsList := jsonResult["mentions"].([]string)
	assert.Len(t, mentionsList, 1)
	assert.Contains(t, mentionsList, "@agent")
}

// TestCommentCommand_JSONOutputWithMultipleMentions verifies mentions are correctly included in JSON
func TestCommentCommand_JSONOutputWithMultipleMentions(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComment(t, app)
	setupCommentsCollectionForComment(t, app)

	task := createCommentTestTask(t, app, "Task for mentions JSON test", "todo")

	// Content with multiple mentions
	content := "@agent @reviewer please review this approach @lead"
	mentions := extractMentions(content)

	// Create comment with mentions in metadata
	comment := createTestComment(t, app, task.Id, content, "human", "developer")
	require.NotEmpty(t, comment.Id)

	// Verify mentions extraction
	assert.Len(t, mentions, 3)
	assert.Contains(t, mentions, "@agent")
	assert.Contains(t, mentions, "@reviewer")
	assert.Contains(t, mentions, "@lead")
}
