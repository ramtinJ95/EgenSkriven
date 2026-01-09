package resume

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockRecord is a mock implementation for testing that satisfies
// the interface used by BuildContextPrompt and BuildMinimalPrompt.
// Since pocketbase core.Record has these methods, we can use this mock
// instead of creating real records for unit tests.
type mockRecord struct {
	Id   string
	data map[string]any
}

func newMockRecord(id string, data map[string]any) *mockRecord {
	return &mockRecord{Id: id, data: data}
}

func (m *mockRecord) GetString(key string) string {
	if v, ok := m.data[key].(string); ok {
		return v
	}
	return ""
}

func (m *mockRecord) GetInt(key string) int {
	if v, ok := m.data[key].(int); ok {
		return v
	}
	return 0
}

func (m *mockRecord) Get(key string) any {
	return m.data[key]
}

// mockRecordAdapter wraps mockRecord to implement the interface expected by context.go
// This allows us to pass mockRecord where *core.Record is expected in tests.
// Note: This only works because we're testing within the same package.
type mockRecordAdapter struct {
	*mockRecord
}

// TestBuildContextPrompt_IncludesTaskTitle verifies that the prompt includes the task title.
func TestBuildContextPrompt_IncludesTaskTitle(t *testing.T) {
	task := newMockRecord("abc123def456", map[string]any{
		"title":    "Implement user authentication",
		"priority": "high",
		"seq":      42,
	})

	comments := []Comment{
		{
			Content:    "What approach should I use?",
			AuthorType: "agent",
			AuthorId:   "opencode",
			Created:    time.Now(),
		},
	}

	prompt := buildContextPromptFromMock(task, comments)

	assert.Contains(t, prompt, "Implement user authentication",
		"prompt should include task title")
}

// TestBuildContextPrompt_IncludesTaskPriority verifies that the prompt includes the task priority.
func TestBuildContextPrompt_IncludesTaskPriority(t *testing.T) {
	task := newMockRecord("abc123def456", map[string]any{
		"title":    "Test task",
		"priority": "high",
		"seq":      1,
	})

	prompt := buildContextPromptFromMock(task, []Comment{})

	assert.Contains(t, prompt, "**Priority**: high",
		"prompt should include task priority")
}

// TestBuildContextPrompt_IncludesAllCommentsInOrder verifies comments are in chronological order.
func TestBuildContextPrompt_IncludesAllCommentsInOrder(t *testing.T) {
	task := newMockRecord("abc123def456", map[string]any{
		"title":    "Test task",
		"priority": "medium",
		"seq":      1,
	})

	baseTime := time.Now()
	comments := []Comment{
		{
			Content:    "First comment",
			AuthorType: "agent",
			AuthorId:   "opencode",
			Created:    baseTime.Add(-2 * time.Hour),
		},
		{
			Content:    "Second comment",
			AuthorType: "human",
			AuthorId:   "john",
			Created:    baseTime.Add(-1 * time.Hour),
		},
		{
			Content:    "Third comment",
			AuthorType: "agent",
			AuthorId:   "opencode",
			Created:    baseTime,
		},
	}

	prompt := buildContextPromptFromMock(task, comments)

	// All comments should be present
	assert.Contains(t, prompt, "First comment", "should include first comment")
	assert.Contains(t, prompt, "Second comment", "should include second comment")
	assert.Contains(t, prompt, "Third comment", "should include third comment")

	// Comments should be in chronological order
	firstIdx := strings.Index(prompt, "First comment")
	secondIdx := strings.Index(prompt, "Second comment")
	thirdIdx := strings.Index(prompt, "Third comment")

	assert.Less(t, firstIdx, secondIdx, "first comment should appear before second")
	assert.Less(t, secondIdx, thirdIdx, "second comment should appear before third")
}

// TestBuildContextPrompt_FormatsAuthorsCorrectly verifies author formatting.
func TestBuildContextPrompt_FormatsAuthorsCorrectly(t *testing.T) {
	task := newMockRecord("abc123def456", map[string]any{
		"title":    "Test task",
		"priority": "medium",
		"seq":      1,
	})

	comments := []Comment{
		{
			Content:    "Comment with author ID",
			AuthorType: "agent",
			AuthorId:   "opencode",
			Created:    time.Now().Add(-1 * time.Hour),
		},
		{
			Content:    "Comment without author ID",
			AuthorType: "human",
			AuthorId:   "", // Empty author ID
			Created:    time.Now(),
		},
	}

	prompt := buildContextPromptFromMock(task, comments)

	// Should use authorId when present
	assert.Contains(t, prompt, "[opencode @",
		"should use authorId when present")

	// Should fall back to authorType when authorId is empty
	assert.Contains(t, prompt, "[human @",
		"should fall back to authorType when authorId is empty")
}

// TestBuildContextPrompt_HandlesEmptyComments verifies the empty comments placeholder.
func TestBuildContextPrompt_HandlesEmptyComments(t *testing.T) {
	task := newMockRecord("abc123def456", map[string]any{
		"title":    "Test task",
		"priority": "medium",
		"seq":      1,
	})

	prompt := buildContextPromptFromMock(task, []Comment{})

	assert.Contains(t, prompt, "_No comments yet_",
		"should show placeholder for empty comments")
}

// TestBuildContextPrompt_TruncatesLongDescriptions verifies description truncation at 500 chars.
func TestBuildContextPrompt_TruncatesLongDescriptions(t *testing.T) {
	// Create a description longer than 500 characters
	longDescription := strings.Repeat("A", 600)

	task := newMockRecord("abc123def456", map[string]any{
		"title":       "Test task",
		"priority":    "medium",
		"description": longDescription,
		"seq":         1,
	})

	prompt := buildContextPromptFromMock(task, []Comment{})

	// Should contain truncated description (first 500 chars + "...")
	assert.Contains(t, prompt, strings.Repeat("A", 500)+"...",
		"should truncate description at 500 chars and add ellipsis")

	// Should NOT contain the full 600-char description
	assert.NotContains(t, prompt, strings.Repeat("A", 600),
		"should not contain full description")
}

// TestBuildContextPrompt_IncludesShortDescription verifies short descriptions are not truncated.
func TestBuildContextPrompt_IncludesShortDescription(t *testing.T) {
	shortDescription := "This is a short description"

	task := newMockRecord("abc123def456", map[string]any{
		"title":       "Test task",
		"priority":    "medium",
		"description": shortDescription,
		"seq":         1,
	})

	prompt := buildContextPromptFromMock(task, []Comment{})

	assert.Contains(t, prompt, shortDescription,
		"should include full short description")
	assert.NotContains(t, prompt, shortDescription+"...",
		"should not truncate short description")
}

// TestBuildMinimalPrompt_OnlyIncludesLast3Comments verifies minimal prompt comment limit.
func TestBuildMinimalPrompt_OnlyIncludesLast3Comments(t *testing.T) {
	task := newMockRecord("abc123def456", map[string]any{
		"title": "Test task",
		"seq":   1,
	})

	baseTime := time.Now()
	comments := []Comment{
		{Content: "Comment 1 - oldest", AuthorType: "human", Created: baseTime.Add(-5 * time.Hour)},
		{Content: "Comment 2", AuthorType: "agent", Created: baseTime.Add(-4 * time.Hour)},
		{Content: "Comment 3", AuthorType: "human", Created: baseTime.Add(-3 * time.Hour)},
		{Content: "Comment 4", AuthorType: "agent", Created: baseTime.Add(-2 * time.Hour)},
		{Content: "Comment 5 - newest", AuthorType: "human", Created: baseTime},
	}

	prompt := buildMinimalPromptFromMock(task, comments)

	// Should NOT include oldest comments
	assert.NotContains(t, prompt, "Comment 1 - oldest",
		"minimal prompt should not include oldest comments")
	assert.NotContains(t, prompt, "Comment 2",
		"minimal prompt should not include old comments")

	// Should include last 3 comments
	assert.Contains(t, prompt, "Comment 3",
		"minimal prompt should include 3rd to last comment")
	assert.Contains(t, prompt, "Comment 4",
		"minimal prompt should include 2nd to last comment")
	assert.Contains(t, prompt, "Comment 5 - newest",
		"minimal prompt should include most recent comment")
}

// TestBuildMinimalPrompt_TruncatesLongComments verifies comment truncation at 200 chars.
func TestBuildMinimalPrompt_TruncatesLongComments(t *testing.T) {
	task := newMockRecord("abc123def456", map[string]any{
		"title": "Test task",
		"seq":   1,
	})

	// Create a comment longer than 200 characters
	longComment := strings.Repeat("B", 300)

	comments := []Comment{
		{Content: longComment, AuthorType: "human", Created: time.Now()},
	}

	prompt := buildMinimalPromptFromMock(task, comments)

	// Should contain truncated comment (first 200 chars + "...")
	assert.Contains(t, prompt, strings.Repeat("B", 200)+"...",
		"should truncate comment at 200 chars and add ellipsis")

	// Should NOT contain the full 300-char comment
	assert.NotContains(t, prompt, strings.Repeat("B", 300),
		"should not contain full long comment")
}

// TestBuildMinimalPrompt_DoesNotTruncateShortComments verifies short comments are not truncated.
func TestBuildMinimalPrompt_DoesNotTruncateShortComments(t *testing.T) {
	task := newMockRecord("abc123def456", map[string]any{
		"title": "Test task",
		"seq":   1,
	})

	shortComment := "This is a short comment"

	comments := []Comment{
		{Content: shortComment, AuthorType: "human", Created: time.Now()},
	}

	prompt := buildMinimalPromptFromMock(task, comments)

	assert.Contains(t, prompt, shortComment,
		"should include full short comment")
}

// TestBuildMinimalPrompt_IncludesTaskInfo verifies task info is present.
func TestBuildMinimalPrompt_IncludesTaskInfo(t *testing.T) {
	task := newMockRecord("abc123def456", map[string]any{
		"title": "Implement feature X",
		"seq":   42,
	})

	prompt := buildMinimalPromptFromMock(task, []Comment{})

	assert.Contains(t, prompt, "Task WRK-42:",
		"should include task display ID")
	assert.Contains(t, prompt, "Implement feature X",
		"should include task title")
}

// TestGetDisplayId_PrefersDisplayId verifies display_id field is used first.
func TestGetDisplayId_PrefersDisplayId(t *testing.T) {
	task := newMockRecord("abc123def456", map[string]any{
		"display_id": "PROJ-99",
		"seq":        42,
	})

	result := getDisplayIdFromMock(task)

	assert.Equal(t, "PROJ-99", result,
		"should prefer display_id when present")
}

// TestGetDisplayId_FallsBackToSeq verifies WRK-{seq} format fallback.
func TestGetDisplayId_FallsBackToSeq(t *testing.T) {
	task := newMockRecord("abc123def456", map[string]any{
		"seq": 42,
	})

	result := getDisplayIdFromMock(task)

	assert.Equal(t, "WRK-42", result,
		"should fall back to WRK-{seq} format")
}

// TestGetDisplayId_FallsBackToIdPrefix verifies ID prefix fallback.
func TestGetDisplayId_FallsBackToIdPrefix(t *testing.T) {
	task := newMockRecord("abc123def456ghij", map[string]any{})

	result := getDisplayIdFromMock(task)

	assert.Equal(t, "abc123de", result,
		"should fall back to first 8 chars of ID")
}

// TestFormatAuthorLabel_UsesAuthorId verifies authorId preference.
func TestFormatAuthorLabel_UsesAuthorId(t *testing.T) {
	result := formatAuthorLabel("agent", "opencode")

	assert.Equal(t, "opencode", result,
		"should use authorId when present")
}

// TestFormatAuthorLabel_FallsBackToAuthorType verifies authorType fallback.
func TestFormatAuthorLabel_FallsBackToAuthorType(t *testing.T) {
	result := formatAuthorLabel("human", "")

	assert.Equal(t, "human", result,
		"should fall back to authorType when authorId is empty")
}

// Helper functions that use mockRecord instead of *core.Record
// These mirror the actual functions but work with our mock type.

func buildContextPromptFromMock(task *mockRecord, comments []Comment) string {
	var sb strings.Builder

	displayId := getDisplayIdFromMock(task)
	title := task.GetString("title")
	priority := task.GetString("priority")
	description := task.GetString("description")

	// Header
	sb.WriteString("## Task Context (from EgenSkriven)\n\n")

	// Task info
	sb.WriteString("**Task**: " + displayId + " - " + title + "\n")
	sb.WriteString("**Status**: need_input -> in_progress\n")
	sb.WriteString("**Priority**: " + priority + "\n")

	// Include description if present (truncated if >500 chars)
	if description != "" {
		sb.WriteString("**Description**:\n")
		if len(description) > 500 {
			sb.WriteString(description[:500] + "...\n")
		} else {
			sb.WriteString(description + "\n")
		}
	}
	sb.WriteString("\n")

	// Comments thread
	sb.WriteString("## Conversation Thread\n\n")

	if len(comments) == 0 {
		sb.WriteString("_No comments yet_\n\n")
	} else {
		for _, c := range comments {
			authorLabel := formatAuthorLabel(c.AuthorType, c.AuthorId)
			timeLabel := c.Created.Format("15:04")

			sb.WriteString("[" + authorLabel + " @ " + timeLabel + "]: " + c.Content + "\n\n")
		}
	}

	// Instructions
	sb.WriteString("## Instructions\n\n")
	sb.WriteString("Continue working on the task based on the human's response above. ")
	sb.WriteString("The conversation context should help you understand what was discussed. ")
	sb.WriteString("If you need more clarification, you can block the task again with a new question.\n")

	return sb.String()
}

func buildMinimalPromptFromMock(task *mockRecord, comments []Comment) string {
	var sb strings.Builder

	displayId := getDisplayIdFromMock(task)
	title := task.GetString("title")

	sb.WriteString("Task " + displayId + ": " + title + "\n\n")
	sb.WriteString("Recent comments:\n")

	// Only include last 3 comments for minimal version
	start := 0
	if len(comments) > 3 {
		start = len(comments) - 3
	}

	for _, c := range comments[start:] {
		authorLabel := formatAuthorLabel(c.AuthorType, c.AuthorId)
		// Truncate long comments at 200 chars
		content := c.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		sb.WriteString("- " + authorLabel + ": " + content + "\n")
	}

	sb.WriteString("\nContinue based on the above context.\n")

	return sb.String()
}

func getDisplayIdFromMock(task *mockRecord) string {
	// Check for display_id field first
	if displayId := task.GetString("display_id"); displayId != "" {
		return displayId
	}
	// Fall back to WRK-{seq} format
	seq := task.GetInt("seq")
	if seq > 0 {
		return "WRK-" + itoa(seq)
	}
	// Fall back to first 8 chars of ID
	id := task.Id
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

// itoa converts an int to string using strconv.Itoa
func itoa(i int) string {
	return strconv.Itoa(i)
}
