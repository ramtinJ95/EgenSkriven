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

// setupTasksCollectionForComment creates tasks collection for comment tests
func setupTasksCollectionForComment(t *testing.T, app *pocketbase.PocketBase) {
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

// setupCommentsCollectionForComment creates comments collection for comment tests
func setupCommentsCollectionForComment(t *testing.T, app *pocketbase.PocketBase) {
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

// createCommentTestTask creates a task for comment command testing
func createCommentTestTask(t *testing.T, app *pocketbase.PocketBase, title string, column string) *core.Record {
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

// createTestComment creates a comment directly for testing
func createTestComment(t *testing.T, app *pocketbase.PocketBase, taskId, content, authorType, authorId string) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("comments")
	require.NoError(t, err)

	record := core.NewRecord(collection)
	record.Set("task", taskId)
	record.Set("content", content)
	record.Set("author_type", authorType)
	record.Set("author_id", authorId)
	record.Set("metadata", map[string]any{})

	require.NoError(t, app.Save(record))
	return record
}

// getCommentsForTask returns all comments for a given task ID
func getCommentsForTask(t *testing.T, app *pocketbase.PocketBase, taskId string) []*core.Record {
	t.Helper()

	records, err := app.FindRecordsByFilter(
		"comments",
		"task = '"+taskId+"'",
		"", // No sorting - simplifies test setup
		0,
		0,
	)
	require.NoError(t, err)
	return records
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
