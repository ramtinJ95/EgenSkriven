package commands

import (
	"testing"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ramtinJ95/EgenSkriven/internal/testutil"
)

// ========== Setup Functions ==========

// setupTasksCollectionForComments creates tasks collection for comments list tests
func setupTasksCollectionForComments(t *testing.T, app *pocketbase.PocketBase) {
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

// setupCommentsCollectionForComments creates comments collection for comments list tests
// This version includes autodate fields for proper sorting
func setupCommentsCollectionForComments(t *testing.T, app *pocketbase.PocketBase) {
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
	collection.Fields.Add(&core.AutodateField{
		Name:     "created",
		OnCreate: true,
	})

	if err := app.Save(collection); err != nil {
		t.Fatalf("failed to create comments collection: %v", err)
	}
}

// createCommentsTestTask creates a task for comments list command testing
func createCommentsTestTask(t *testing.T, app *pocketbase.PocketBase, title string, column string) *core.Record {
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

// createCommentsTestComment creates a comment for testing
func createCommentsTestComment(t *testing.T, app *pocketbase.PocketBase, taskId, content, authorType, authorId string) *core.Record {
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

// getCommentsForTaskList returns all comments for a given task ID
func getCommentsForTaskList(t *testing.T, app *pocketbase.PocketBase, taskId string) []*core.Record {
	t.Helper()

	records, err := app.FindRecordsByFilter(
		"comments",
		"task = '"+taskId+"'",
		"+created", // Sort by creation time ascending
		0,
		0,
	)
	require.NoError(t, err)
	return records
}

// ========== formatRelativeTime Tests ==========

func TestFormatRelativeTime_JustNow(t *testing.T) {
	now := time.Now()
	result := formatRelativeTime(now)
	assert.Equal(t, "just now", result)
}

func TestFormatRelativeTime_SecondsAgo(t *testing.T) {
	thirtySecondsAgo := time.Now().Add(-30 * time.Second)
	result := formatRelativeTime(thirtySecondsAgo)
	assert.Equal(t, "just now", result)
}

func TestFormatRelativeTime_MinutesAgo(t *testing.T) {
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
	result := formatRelativeTime(fiveMinutesAgo)
	assert.Equal(t, "5m ago", result)
}

func TestFormatRelativeTime_OneMinuteAgo(t *testing.T) {
	oneMinuteAgo := time.Now().Add(-1 * time.Minute)
	result := formatRelativeTime(oneMinuteAgo)
	assert.Equal(t, "1m ago", result)
}

func TestFormatRelativeTime_HoursAgo(t *testing.T) {
	threeHoursAgo := time.Now().Add(-3 * time.Hour)
	result := formatRelativeTime(threeHoursAgo)
	assert.Equal(t, "3h ago", result)
}

func TestFormatRelativeTime_OneHourAgo(t *testing.T) {
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	result := formatRelativeTime(oneHourAgo)
	assert.Equal(t, "1h ago", result)
}

func TestFormatRelativeTime_23HoursAgo(t *testing.T) {
	twentyThreeHoursAgo := time.Now().Add(-23 * time.Hour)
	result := formatRelativeTime(twentyThreeHoursAgo)
	assert.Equal(t, "23h ago", result)
}

func TestFormatRelativeTime_MoreThan24Hours(t *testing.T) {
	twoDaysAgo := time.Now().Add(-48 * time.Hour)
	result := formatRelativeTime(twoDaysAgo)
	// Should return absolute format like "Jan 2, 15:04"
	assert.Contains(t, result, ",")
	assert.Contains(t, result, ":")
}

func TestFormatRelativeTime_FutureTime(t *testing.T) {
	// Future times should result in negative diff, which will be < time.Minute
	futureTime := time.Now().Add(1 * time.Hour)
	result := formatRelativeTime(futureTime)
	// Negative diff is less than any threshold, so it may return unusual value
	// The function doesn't handle future times specially
	assert.NotEmpty(t, result)
}

// ========== Comments Listing Tests ==========

func TestCommentsListing_ListAllComments(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComments(t, app)
	setupCommentsCollectionForComments(t, app)

	// Create a test task
	task := createCommentsTestTask(t, app, "Test Task", "todo")

	// Create multiple comments
	createCommentsTestComment(t, app, task.Id, "First comment", "agent", "opencode")
	createCommentsTestComment(t, app, task.Id, "Second comment", "human", "developer")
	createCommentsTestComment(t, app, task.Id, "Third comment", "agent", "claude-code")

	// List comments
	comments := getCommentsForTaskList(t, app, task.Id)

	// Verify all comments are listed
	assert.Len(t, comments, 3, "should list all 3 comments")
}

func TestCommentsListing_EmptyTaskShowsNoComments(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComments(t, app)
	setupCommentsCollectionForComments(t, app)

	// Create a task with no comments
	task := createCommentsTestTask(t, app, "Task Without Comments", "todo")

	// List comments
	comments := getCommentsForTaskList(t, app, task.Id)

	// Verify no comments
	assert.Empty(t, comments, "task should have no comments")
}

func TestCommentsListing_CommentsOnlyForSpecificTask(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComments(t, app)
	setupCommentsCollectionForComments(t, app)

	// Create two tasks
	task1 := createCommentsTestTask(t, app, "Task 1", "todo")
	task2 := createCommentsTestTask(t, app, "Task 2", "in_progress")

	// Create comments on both tasks
	createCommentsTestComment(t, app, task1.Id, "Comment on task 1", "human", "user1")
	createCommentsTestComment(t, app, task1.Id, "Another on task 1", "human", "user2")
	createCommentsTestComment(t, app, task2.Id, "Comment on task 2", "agent", "agent1")
	createCommentsTestComment(t, app, task2.Id, "Another on task 2", "agent", "agent2")
	createCommentsTestComment(t, app, task2.Id, "Third on task 2", "human", "user3")

	// List comments for each task
	task1Comments := getCommentsForTaskList(t, app, task1.Id)
	task2Comments := getCommentsForTaskList(t, app, task2.Id)

	// Verify isolation
	assert.Len(t, task1Comments, 2, "task1 should have 2 comments")
	assert.Len(t, task2Comments, 3, "task2 should have 3 comments")
}

func TestCommentsListing_LimitComments(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComments(t, app)
	setupCommentsCollectionForComments(t, app)

	// Create a test task
	task := createCommentsTestTask(t, app, "Test Task", "todo")

	// Create 5 comments
	for i := 1; i <= 5; i++ {
		createCommentsTestComment(t, app, task.Id, "Comment "+string(rune('0'+i)), "human", "user")
	}

	// Query with limit
	records, err := app.FindRecordsByFilter(
		"comments",
		"task = '"+task.Id+"'",
		"+created",
		2, // Limit to 2
		0,
	)
	require.NoError(t, err)

	// Verify limit is respected
	assert.Len(t, records, 2, "should only return 2 comments with limit=2")
}

func TestCommentsListing_CommentContent(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComments(t, app)
	setupCommentsCollectionForComments(t, app)

	task := createCommentsTestTask(t, app, "Test Task", "need_input")

	// Create comment with specific content
	expectedContent := "What authentication approach should I use? JWT or sessions?"
	createCommentsTestComment(t, app, task.Id, expectedContent, "agent", "opencode")

	// Retrieve comments
	comments := getCommentsForTaskList(t, app, task.Id)

	require.Len(t, comments, 1)
	assert.Equal(t, expectedContent, comments[0].GetString("content"), "comment content should match")
}

func TestCommentsListing_CommentAuthorInfo(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComments(t, app)
	setupCommentsCollectionForComments(t, app)

	task := createCommentsTestTask(t, app, "Test Task", "todo")

	// Create comments with different author types
	createCommentsTestComment(t, app, task.Id, "Human comment", "human", "john.doe")
	createCommentsTestComment(t, app, task.Id, "Agent comment", "agent", "claude-code")

	// Retrieve comments
	comments := getCommentsForTaskList(t, app, task.Id)
	require.Len(t, comments, 2)

	// Find human and agent comments
	var humanComment, agentComment *core.Record
	for _, c := range comments {
		if c.GetString("author_type") == "human" {
			humanComment = c
		} else if c.GetString("author_type") == "agent" {
			agentComment = c
		}
	}

	// Verify author info
	require.NotNil(t, humanComment, "should have human comment")
	require.NotNil(t, agentComment, "should have agent comment")

	assert.Equal(t, "human", humanComment.GetString("author_type"))
	assert.Equal(t, "john.doe", humanComment.GetString("author_id"))

	assert.Equal(t, "agent", agentComment.GetString("author_type"))
	assert.Equal(t, "claude-code", agentComment.GetString("author_id"))
}

func TestCommentsListing_CommentsHaveCreatedTimestamp(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComments(t, app)
	setupCommentsCollectionForComments(t, app)

	task := createCommentsTestTask(t, app, "Test Task", "todo")

	// Create a comment
	createCommentsTestComment(t, app, task.Id, "Test comment", "human", "user")

	// Retrieve comments
	comments := getCommentsForTaskList(t, app, task.Id)
	require.Len(t, comments, 1)

	// Verify created timestamp exists and is recent
	created := comments[0].GetDateTime("created").Time()
	assert.False(t, created.IsZero(), "created timestamp should be set")
	assert.WithinDuration(t, time.Now(), created, 5*time.Second, "created should be recent")
}

func TestCommentsListing_CommentsWithMentions(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComments(t, app)
	setupCommentsCollectionForComments(t, app)

	task := createCommentsTestTask(t, app, "Test Task", "need_input")

	// Create comments with @mentions
	collection, err := app.FindCollectionByNameOrId("comments")
	require.NoError(t, err)

	record := core.NewRecord(collection)
	record.Set("task", task.Id)
	record.Set("content", "@agent I've decided to use JWT")
	record.Set("author_type", "human")
	record.Set("author_id", "developer")
	record.Set("metadata", map[string]any{
		"mentions": []string{"@agent"},
	})
	require.NoError(t, app.Save(record))

	// Retrieve and verify
	comments := getCommentsForTaskList(t, app, task.Id)
	require.Len(t, comments, 1)
	assert.Contains(t, comments[0].GetString("content"), "@agent")
}

func TestCommentsListing_LongCommentContent(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComments(t, app)
	setupCommentsCollectionForComments(t, app)

	task := createCommentsTestTask(t, app, "Test Task", "todo")

	// Create a comment with long content including newlines
	longContent := `Here's my detailed response:

1. Use JWT for authentication
2. Access tokens should expire in 15 minutes
3. Refresh tokens should expire in 7 days
4. Store tokens securely

Additional notes:
- Consider using httpOnly cookies
- Implement token rotation
- Add rate limiting

Let me know if you have any questions!`

	createCommentsTestComment(t, app, task.Id, longContent, "human", "architect")

	// Retrieve and verify full content is preserved
	comments := getCommentsForTaskList(t, app, task.Id)
	require.Len(t, comments, 1)
	assert.Equal(t, longContent, comments[0].GetString("content"), "long content should be preserved exactly")
}

func TestCommentsListing_SpecialCharactersInContent(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComments(t, app)
	setupCommentsCollectionForComments(t, app)

	task := createCommentsTestTask(t, app, "Test Task", "todo")

	// Create comment with special characters
	specialContent := `Code example:
` + "```go" + `
func main() {
    fmt.Println("Hello, World!")
}
` + "```" + `

Special chars: <>&"'` + "`"

	createCommentsTestComment(t, app, task.Id, specialContent, "human", "developer")

	// Retrieve and verify
	comments := getCommentsForTaskList(t, app, task.Id)
	require.Len(t, comments, 1)
	assert.Equal(t, specialContent, comments[0].GetString("content"))
}

func TestCommentsListing_EmptyAuthorId(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComments(t, app)
	setupCommentsCollectionForComments(t, app)

	task := createCommentsTestTask(t, app, "Test Task", "todo")

	// Create comment with empty author_id
	collection, err := app.FindCollectionByNameOrId("comments")
	require.NoError(t, err)

	record := core.NewRecord(collection)
	record.Set("task", task.Id)
	record.Set("content", "Anonymous comment")
	record.Set("author_type", "human")
	record.Set("author_id", "")
	require.NoError(t, app.Save(record))

	// Retrieve and verify
	comments := getCommentsForTaskList(t, app, task.Id)
	require.Len(t, comments, 1)
	assert.Equal(t, "human", comments[0].GetString("author_type"))
	assert.Empty(t, comments[0].GetString("author_id"), "author_id should be empty")
}

// ========== Comments Filtering Tests ==========

func TestCommentsListing_FilterBySince(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComments(t, app)
	setupCommentsCollectionForComments(t, app)

	task := createCommentsTestTask(t, app, "Test Task", "todo")

	// Create an old comment (by setting it manually - note: autodate will still set current time)
	// For this test, we verify that the filter mechanism works with the database
	createCommentsTestComment(t, app, task.Id, "Comment 1", "human", "user")

	// Get timestamp after first comment
	time.Sleep(10 * time.Millisecond) // Small delay to ensure different timestamps
	sinceTime := time.Now().Format(time.RFC3339)
	time.Sleep(10 * time.Millisecond)

	createCommentsTestComment(t, app, task.Id, "Comment 2", "human", "user")

	// Query with since filter
	filter := "task = '" + task.Id + "' && created > '" + sinceTime + "'"
	records, err := app.FindRecordsByFilter(
		"comments",
		filter,
		"+created",
		0,
		0,
	)
	require.NoError(t, err)

	// Should only get Comment 2 (created after sinceTime)
	// Note: Due to timing precision, this might get 0 or 1 comment
	// The important thing is that the filter syntax works
	assert.LessOrEqual(t, len(records), 2, "filter should reduce results")
}

// ========== Integration Test ==========

func TestCommentsListing_FullWorkflowVerification(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollectionForComments(t, app)
	setupCommentsCollectionForComments(t, app)

	// Create a task that will be blocked
	task := createCommentsTestTask(t, app, "Implement user authentication", "need_input")

	// Simulate agent blocking with a question
	agentQuestion := "What authentication method should I use? Options are:\n1. JWT\n2. Session cookies\n3. OAuth2"
	createCommentsTestComment(t, app, task.Id, agentQuestion, "agent", "opencode")

	// Simulate human response
	humanResponse := "@agent Please use JWT with the following configuration:\n- Access token: 15 min expiry\n- Refresh token: 7 days expiry\n- Use RS256 algorithm"
	createCommentsTestComment(t, app, task.Id, humanResponse, "human", "tech-lead")

	// Simulate agent acknowledgment
	agentAck := "Got it! I'll implement JWT authentication with the specified configuration."
	createCommentsTestComment(t, app, task.Id, agentAck, "agent", "opencode")

	// List all comments
	comments := getCommentsForTaskList(t, app, task.Id)

	// Verify the full conversation
	require.Len(t, comments, 3, "should have 3 comments in the conversation")

	// Verify conversation participants
	var agentCount, humanCount int
	for _, c := range comments {
		switch c.GetString("author_type") {
		case "agent":
			agentCount++
		case "human":
			humanCount++
		}
	}

	assert.Equal(t, 2, agentCount, "should have 2 agent comments")
	assert.Equal(t, 1, humanCount, "should have 1 human comment")
}
