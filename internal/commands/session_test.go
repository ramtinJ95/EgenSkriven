package commands

import (
	"fmt"
	"testing"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ramtinJ95/EgenSkriven/internal/output"
	"github.com/ramtinJ95/EgenSkriven/internal/testutil"
)

// ========== Setup Functions ==========

// SetupSessionsCollection creates the sessions collection for session tests.
func SetupSessionsCollection(t *testing.T, app *pocketbase.PocketBase) {
	t.Helper()

	_, err := app.FindCollectionByNameOrId("sessions")
	if err == nil {
		return
	}

	// Get tasks collection for relation
	tasks, err := app.FindCollectionByNameOrId("tasks")
	require.NoError(t, err, "tasks collection must exist before sessions")

	collection := core.NewBaseCollection("sessions")

	// Task relation (required)
	collection.Fields.Add(&core.RelationField{
		Name:          "task",
		CollectionId:  tasks.Id,
		MaxSelect:     1,
		Required:      true,
		CascadeDelete: true,
	})

	// Tool identifier
	collection.Fields.Add(&core.SelectField{
		Name:     "tool",
		Required: true,
		Values:   []string{"opencode", "claude-code", "codex"},
	})

	// External session reference
	collection.Fields.Add(&core.TextField{
		Name:     "external_ref",
		Required: true,
		Max:      500,
	})

	// Reference type
	collection.Fields.Add(&core.SelectField{
		Name:     "ref_type",
		Required: true,
		Values:   []string{"uuid", "path"},
	})

	// Working directory
	collection.Fields.Add(&core.TextField{
		Name:     "working_dir",
		Required: true,
		Max:      1000,
	})

	// Session status
	collection.Fields.Add(&core.SelectField{
		Name:     "status",
		Required: true,
		Values:   []string{"active", "paused", "completed", "abandoned"},
	})

	// Auto-timestamp on creation
	collection.Fields.Add(&core.AutodateField{
		Name:     "created",
		OnCreate: true,
	})

	// End timestamp (optional)
	collection.Fields.Add(&core.DateField{
		Name: "ended_at",
	})

	require.NoError(t, app.Save(collection), "failed to create sessions collection")
}

// SetupTasksCollectionWithAgentSession creates tasks collection with agent_session field.
func SetupTasksCollectionWithAgentSession(t *testing.T, app *pocketbase.PocketBase) {
	t.Helper()

	// First create basic tasks collection
	SetupTasksCollection(t, app)

	// Then add agent_session field if not present
	tasks, err := app.FindCollectionByNameOrId("tasks")
	require.NoError(t, err)

	if tasks.Fields.GetByName("agent_session") == nil {
		tasks.Fields.Add(&core.JSONField{
			Name:    "agent_session",
			MaxSize: 10000,
		})
		require.NoError(t, app.Save(tasks), "failed to add agent_session field")
	}
}

// CreateSessionTestTask creates a task for session command testing.
func CreateSessionTestTask(t *testing.T, app *pocketbase.PocketBase, title string, column string) *core.Record {
	return CreateTestTask(t, app, title, column)
}

// CreateTestSession creates a session record directly for testing.
func CreateTestSession(t *testing.T, app *pocketbase.PocketBase, taskId, tool, externalRef, refType, workingDir, status string) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("sessions")
	require.NoError(t, err)

	record := core.NewRecord(collection)
	record.Set("task", taskId)
	record.Set("tool", tool)
	record.Set("external_ref", externalRef)
	record.Set("ref_type", refType)
	record.Set("working_dir", workingDir)
	record.Set("status", status)

	require.NoError(t, app.Save(record))
	return record
}

// GetSessionsForTask returns all sessions for a given task ID.
func GetSessionsForTask(t *testing.T, app *pocketbase.PocketBase, taskId string) []*core.Record {
	t.Helper()

	records, err := app.FindRecordsByFilter(
		"sessions",
		"task = {:taskId}",
		"-created",
		0,
		0,
		dbx.Params{"taskId": taskId},
	)
	require.NoError(t, err)
	return records
}

// SimulateSessionLink simulates what the session link command does:
// sets agent_session on task and creates a session record.
func SimulateSessionLink(t *testing.T, app *pocketbase.PocketBase, task *core.Record, tool, ref, workingDir string) *core.Record {
	t.Helper()

	refType := determineRefType(ref)

	// Set agent_session on task
	session := map[string]any{
		"tool":        tool,
		"ref":         ref,
		"ref_type":    refType,
		"working_dir": workingDir,
		"linked_at":   "2026-01-08T12:00:00Z",
	}
	task.Set("agent_session", session)

	// Add to history
	addHistoryEntry(task, "session_linked", "agent", map[string]any{
		"tool":        tool,
		"session_ref": ref,
	})

	require.NoError(t, app.Save(task))

	// Create session record
	return CreateTestSession(t, app, task.Id, tool, ref, refType, workingDir, "active")
}

// ========== determineRefType Tests ==========

func TestDetermineRefType_UUID(t *testing.T) {
	tests := []struct {
		ref      string
		expected string
	}{
		{"550e8400-e29b-41d4-a716-446655440000", "uuid"},
		{"abc123def456", "uuid"},
		{"simple-session-id", "uuid"},
		{"test-123", "uuid"},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			got := determineRefType(tt.ref)
			assert.Equal(t, tt.expected, got, "determineRefType(%q)", tt.ref)
		})
	}
}

func TestDetermineRefType_Path(t *testing.T) {
	tests := []struct {
		ref      string
		expected string
	}{
		{"/home/user/.aider/history.md", "path"},
		{"./relative/path", "path"},
		{"/absolute/path/to/session", "path"},
		{"path/with/slash", "path"},
		{"C:\\Users\\path", "path"},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			got := determineRefType(tt.ref)
			assert.Equal(t, tt.expected, got, "determineRefType(%q)", tt.ref)
		})
	}
}

// ========== Session Link Tests ==========

func TestSessionLink_CreatesSessionOnTask(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollectionWithAgentSession(t, app)
	SetupSessionsCollection(t, app)

	// Create a test task
	task := CreateSessionTestTask(t, app, "Test Task", "todo")
	require.NotEmpty(t, task.Id)

	// Simulate linking a session
	tool := "opencode"
	ref := "test-session-123"
	workingDir := "/home/user/project"
	SimulateSessionLink(t, app, task, tool, ref, workingDir)

	// Re-fetch task from database
	task, err := app.FindRecordById("tasks", task.Id)
	require.NoError(t, err)

	// Verify agent_session was set
	sessionData := task.Get("agent_session")
	require.NotNil(t, sessionData, "agent_session should be set")

	session, err := output.ParseAgentSession(sessionData)
	require.NoError(t, err)
	require.NotNil(t, session)

	assert.Equal(t, tool, session["tool"])
	assert.Equal(t, ref, session["ref"])
	assert.Equal(t, "uuid", session["ref_type"])
	assert.Equal(t, workingDir, session["working_dir"])
}

func TestSessionLink_CreatesRecordInSessionsTable(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollectionWithAgentSession(t, app)
	SetupSessionsCollection(t, app)

	task := CreateSessionTestTask(t, app, "Test Task", "in_progress")

	// Link a session
	tool := "claude-code"
	ref := "session-456"
	workingDir := "/home/user/myproject"
	SimulateSessionLink(t, app, task, tool, ref, workingDir)

	// Verify session record was created
	sessions := GetSessionsForTask(t, app, task.Id)
	require.Len(t, sessions, 1, "should have 1 session record")

	sessionRecord := sessions[0]
	assert.Equal(t, task.Id, sessionRecord.GetString("task"))
	assert.Equal(t, tool, sessionRecord.GetString("tool"))
	assert.Equal(t, ref, sessionRecord.GetString("external_ref"))
	assert.Equal(t, "uuid", sessionRecord.GetString("ref_type"))
	assert.Equal(t, workingDir, sessionRecord.GetString("working_dir"))
	assert.Equal(t, "active", sessionRecord.GetString("status"))
}

func TestSessionLink_ReplacesExistingSession(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollectionWithAgentSession(t, app)
	SetupSessionsCollection(t, app)

	task := CreateSessionTestTask(t, app, "Test Task", "todo")

	// Link first session
	SimulateSessionLink(t, app, task, "opencode", "session-1", "/home/user/project")

	// Re-fetch task
	task, _ = app.FindRecordById("tasks", task.Id)

	// Mark old session as abandoned (simulating what the real link command does)
	oldSessions, _ := app.FindRecordsByFilter(
		"sessions",
		fmt.Sprintf("task = '%s' && external_ref = 'session-1' && status = 'active'", task.Id),
		"",
		1,
		0,
	)
	if len(oldSessions) > 0 {
		oldSessions[0].Set("status", "abandoned")
		app.Save(oldSessions[0])
	}

	// Link second session
	SimulateSessionLink(t, app, task, "claude-code", "session-2", "/home/user/project")

	// Re-fetch task
	task, err := app.FindRecordById("tasks", task.Id)
	require.NoError(t, err)

	// Verify current session is session-2
	sessionData := task.Get("agent_session")
	session, _ := output.ParseAgentSession(sessionData)
	assert.Equal(t, "session-2", session["ref"])
	assert.Equal(t, "claude-code", session["tool"])

	// Verify session-1 is in history as abandoned
	sessions := GetSessionsForTask(t, app, task.Id)
	require.Len(t, sessions, 2, "should have 2 session records")

	var foundAbandoned, foundActive bool
	for _, s := range sessions {
		if s.GetString("external_ref") == "session-1" && s.GetString("status") == "abandoned" {
			foundAbandoned = true
		}
		if s.GetString("external_ref") == "session-2" && s.GetString("status") == "active" {
			foundActive = true
		}
	}
	assert.True(t, foundAbandoned, "session-1 should be abandoned")
	assert.True(t, foundActive, "session-2 should be active")
}

func TestSessionLink_InvalidTool(t *testing.T) {
	// Test that invalid tool name is rejected
	validTools := []string{"opencode", "claude-code", "codex"}

	invalidTools := []string{"invalid", "aider", "copilot", ""}
	for _, tool := range invalidTools {
		t.Run("invalid_"+tool, func(t *testing.T) {
			isValid := containsString(validTools, tool)
			assert.False(t, isValid, "tool %q should be invalid", tool)
		})
	}
}

func TestSessionLink_ValidTools(t *testing.T) {
	validTools := []string{"opencode", "claude-code", "codex"}

	for _, tool := range validTools {
		t.Run("valid_"+tool, func(t *testing.T) {
			isValid := containsString(validSessionTools, tool)
			assert.True(t, isValid, "tool %q should be valid", tool)
		})
	}
}

func TestSessionLink_PathTypeRef(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollectionWithAgentSession(t, app)
	SetupSessionsCollection(t, app)

	task := CreateSessionTestTask(t, app, "Test Task", "todo")

	// Link a session with a path-type reference
	pathRef := "/home/user/.aider/history.md"
	SimulateSessionLink(t, app, task, "opencode", pathRef, "/home/user/project")

	// Re-fetch task
	task, _ = app.FindRecordById("tasks", task.Id)

	sessionData := task.Get("agent_session")
	session, _ := output.ParseAgentSession(sessionData)

	assert.Equal(t, pathRef, session["ref"])
	assert.Equal(t, "path", session["ref_type"], "ref_type should be 'path' for path references")
}

// ========== Session Show Tests ==========

func TestSessionShow_NoSessionLinked(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollectionWithAgentSession(t, app)
	SetupSessionsCollection(t, app)

	task := CreateSessionTestTask(t, app, "Task Without Session", "todo")

	// Verify task has no session (output.ParseAgentSession handles empty/nil properly)
	sessionData := task.Get("agent_session")
	session, err := output.ParseAgentSession(sessionData)
	assert.NoError(t, err)
	assert.Nil(t, session, "task should not have a session")
}

func TestSessionShow_DisplaysSessionDetails(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollectionWithAgentSession(t, app)
	SetupSessionsCollection(t, app)

	task := CreateSessionTestTask(t, app, "Task With Session", "in_progress")

	// Link a session
	tool := "codex"
	ref := "my-codex-session-xyz"
	workingDir := "/home/user/myproject"
	SimulateSessionLink(t, app, task, tool, ref, workingDir)

	// Re-fetch task
	task, _ = app.FindRecordById("tasks", task.Id)

	// Verify session details can be retrieved
	sessionData := task.Get("agent_session")
	session, err := output.ParseAgentSession(sessionData)
	require.NoError(t, err)
	require.NotNil(t, session)

	assert.Equal(t, tool, session["tool"])
	assert.Equal(t, ref, session["ref"])
	assert.Equal(t, "uuid", session["ref_type"])
	assert.Equal(t, workingDir, session["working_dir"])
	assert.NotEmpty(t, session["linked_at"])
}

func TestSessionShow_NeedInputTaskShowsResumeHint(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollectionWithAgentSession(t, app)
	SetupSessionsCollection(t, app)

	// Create a task in need_input state
	task := CreateSessionTestTask(t, app, "Blocked Task", "need_input")

	// Link a session
	SimulateSessionLink(t, app, task, "opencode", "blocked-session", "/home/user/project")

	// Re-fetch task
	task, _ = app.FindRecordById("tasks", task.Id)

	// Verify task is in need_input
	assert.Equal(t, "need_input", task.GetString("column"))

	// Verify session is linked
	sessionData := task.Get("agent_session")
	assert.NotNil(t, sessionData, "session should be linked")
}

// ========== Session History Tests ==========

func TestSessionHistory_ShowsAllSessionsForTask(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollectionWithAgentSession(t, app)
	SetupSessionsCollection(t, app)

	task := CreateSessionTestTask(t, app, "Task With History", "todo")

	// Create multiple session records directly
	CreateTestSession(t, app, task.Id, "opencode", "session-1", "uuid", "/home/user/project", "abandoned")
	CreateTestSession(t, app, task.Id, "claude-code", "session-2", "uuid", "/home/user/project", "abandoned")
	CreateTestSession(t, app, task.Id, "codex", "session-3", "uuid", "/home/user/project", "active")

	// Fetch all sessions
	sessions := GetSessionsForTask(t, app, task.Id)
	assert.Len(t, sessions, 3, "should have 3 sessions in history")
}

func TestSessionHistory_CorrectStatus(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollectionWithAgentSession(t, app)
	SetupSessionsCollection(t, app)

	task := CreateSessionTestTask(t, app, "Test Task", "todo")

	// Create sessions with different statuses
	CreateTestSession(t, app, task.Id, "opencode", "s1", "uuid", "/home", "active")
	CreateTestSession(t, app, task.Id, "opencode", "s2", "uuid", "/home", "abandoned")
	CreateTestSession(t, app, task.Id, "opencode", "s3", "uuid", "/home", "completed")
	CreateTestSession(t, app, task.Id, "opencode", "s4", "uuid", "/home", "paused")

	sessions := GetSessionsForTask(t, app, task.Id)
	require.Len(t, sessions, 4)

	statuses := make(map[string]bool)
	for _, s := range sessions {
		statuses[s.GetString("status")] = true
	}

	assert.True(t, statuses["active"])
	assert.True(t, statuses["abandoned"])
	assert.True(t, statuses["completed"])
	assert.True(t, statuses["paused"])
}

func TestSessionHistory_EmptyWhenNoSessions(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollectionWithAgentSession(t, app)
	SetupSessionsCollection(t, app)

	task := CreateSessionTestTask(t, app, "Task Without Sessions", "todo")

	sessions := GetSessionsForTask(t, app, task.Id)
	assert.Empty(t, sessions, "should have no sessions")
}

// ========== Session Unlink Tests ==========

func TestSessionUnlink_ClearsAgentSession(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollectionWithAgentSession(t, app)
	SetupSessionsCollection(t, app)

	task := CreateSessionTestTask(t, app, "Test Task", "in_progress")

	// Link a session first
	SimulateSessionLink(t, app, task, "opencode", "session-to-unlink", "/home/user/project")

	// Re-fetch task
	task, _ = app.FindRecordById("tasks", task.Id)

	// Verify session is linked
	sessionData := task.Get("agent_session")
	require.NotNil(t, sessionData)

	// Simulate unlink
	task.Set("agent_session", nil)
	addHistoryEntry(task, "session_unlinked", "user", map[string]any{
		"tool":         "opencode",
		"session_ref":  "session-to-unlink",
		"final_status": "abandoned",
	})
	require.NoError(t, app.Save(task))

	// Re-fetch and verify session is cleared
	task, _ = app.FindRecordById("tasks", task.Id)
	sessionData = task.Get("agent_session")
	// output.ParseAgentSession handles empty/nil properly - returns nil for empty JSON
	session, err := output.ParseAgentSession(sessionData)
	assert.NoError(t, err)
	assert.Nil(t, session, "agent_session should be cleared")
}

func TestSessionUnlink_ValidStatuses(t *testing.T) {
	validStatuses := []string{"abandoned", "completed"}

	for _, status := range validStatuses {
		t.Run(status, func(t *testing.T) {
			isValid := containsString(validStatuses, status)
			assert.True(t, isValid)
		})
	}

	invalidStatuses := []string{"active", "paused", "invalid", ""}
	for _, status := range invalidStatuses {
		t.Run("invalid_"+status, func(t *testing.T) {
			isValid := containsString(validStatuses, status)
			assert.False(t, isValid)
		})
	}
}

// ========== output.ParseAgentSession Tests ==========

func TestParseSessionData_NilReturnsNil(t *testing.T) {
	result, err := output.ParseAgentSession(nil)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestParseSessionData_EmptyStringReturnsNil(t *testing.T) {
	result, err := output.ParseAgentSession("")
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestParseSessionData_MapReturnsMap(t *testing.T) {
	input := map[string]any{
		"tool":        "opencode",
		"ref":         "test-123",
		"ref_type":    "uuid",
		"working_dir": "/home/user",
		"linked_at":   "2026-01-08T12:00:00Z",
	}

	result, err := output.ParseAgentSession(input)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "opencode", result["tool"])
	assert.Equal(t, "test-123", result["ref"])
}

func TestParseSessionData_JSONStringParsesCorrectly(t *testing.T) {
	jsonStr := `{"tool":"claude-code","ref":"abc-def","ref_type":"uuid","working_dir":"/home"}`

	result, err := output.ParseAgentSession(jsonStr)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "claude-code", result["tool"])
	assert.Equal(t, "abc-def", result["ref"])
}

func TestParseSessionData_NullStringReturnsNil(t *testing.T) {
	result, err := output.ParseAgentSession("null")
	assert.NoError(t, err)
	assert.Nil(t, result)
}

// ========== output.TruncateMiddle Tests ==========

func TestTruncateSessionRef_ShortString(t *testing.T) {
	short := "abc123"
	result := output.TruncateMiddle(short, 40)
	assert.Equal(t, short, result)
}

func TestTruncateSessionRef_LongString(t *testing.T) {
	long := "550e8400-e29b-41d4-a716-446655440000-extra-long-id"
	result := output.TruncateMiddle(long, 20)
	// Result will be at most maxLen, with "..." in the middle
	assert.LessOrEqual(t, len(result), 20)
	assert.Contains(t, result, "...")
}

func TestTruncateSessionRef_ExactLength(t *testing.T) {
	exact := "12345678901234567890"
	result := output.TruncateMiddle(exact, 20)
	assert.Equal(t, exact, result)
}

// ========== containsString Tests ==========

func TestContainsString_Found(t *testing.T) {
	slice := []string{"opencode", "claude-code", "codex"}
	assert.True(t, containsString(slice, "opencode"))
	assert.True(t, containsString(slice, "claude-code"))
	assert.True(t, containsString(slice, "codex"))
}

func TestContainsString_NotFound(t *testing.T) {
	slice := []string{"opencode", "claude-code", "codex"}
	assert.False(t, containsString(slice, "aider"))
	assert.False(t, containsString(slice, ""))
	assert.False(t, containsString(slice, "OPENCODE")) // Case sensitive
}

func TestContainsString_EmptySlice(t *testing.T) {
	var slice []string
	assert.False(t, containsString(slice, "anything"))
}

// ========== getSessionStatusIcon Tests ==========

func TestGetSessionStatusIcon(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"active", "[ACTIVE]"},
		{"paused", "[PAUSED]"},
		{"completed", "[DONE]  "},
		{"abandoned", "[OLD]   "},
		{"unknown", "[?]     "},
		{"", "[?]     "},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := getSessionStatusIcon(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ========== Integration Tests ==========

func TestFullSessionWorkflow(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollectionWithAgentSession(t, app)
	SetupSessionsCollection(t, app)

	// Step 1: Create a task
	task := CreateSessionTestTask(t, app, "Full Workflow Task", "todo")
	require.NotEmpty(t, task.Id)

	// Step 2: Link first session
	SimulateSessionLink(t, app, task, "opencode", "session-1", "/home/user/project")
	task, _ = app.FindRecordById("tasks", task.Id)

	// Verify session is linked
	sessionData := task.Get("agent_session")
	session, _ := output.ParseAgentSession(sessionData)
	assert.Equal(t, "session-1", session["ref"])

	// Step 3: Mark old session as abandoned and link new one
	oldSessions, _ := app.FindRecordsByFilter(
		"sessions",
		fmt.Sprintf("task = '%s' && status = 'active'", task.Id),
		"",
		10,
		0,
	)
	for _, s := range oldSessions {
		s.Set("status", "abandoned")
		app.Save(s)
	}

	SimulateSessionLink(t, app, task, "claude-code", "session-2", "/home/user/project")
	task, _ = app.FindRecordById("tasks", task.Id)

	// Verify current session is session-2
	sessionData = task.Get("agent_session")
	session, _ = output.ParseAgentSession(sessionData)
	assert.Equal(t, "session-2", session["ref"])
	assert.Equal(t, "claude-code", session["tool"])

	// Step 4: Link a third session
	oldSessions, _ = app.FindRecordsByFilter(
		"sessions",
		fmt.Sprintf("task = '%s' && status = 'active'", task.Id),
		"",
		10,
		0,
	)
	for _, s := range oldSessions {
		s.Set("status", "abandoned")
		app.Save(s)
	}

	SimulateSessionLink(t, app, task, "codex", "session-3", "/home/user/project")
	task, _ = app.FindRecordById("tasks", task.Id)

	// Step 5: Verify history has all 3 sessions
	sessions := GetSessionsForTask(t, app, task.Id)
	assert.Len(t, sessions, 3, "should have 3 sessions in history")

	// Count by status
	activeCount := 0
	abandonedCount := 0
	for _, s := range sessions {
		if s.GetString("status") == "active" {
			activeCount++
		} else if s.GetString("status") == "abandoned" {
			abandonedCount++
		}
	}
	assert.Equal(t, 1, activeCount, "should have 1 active session")
	assert.Equal(t, 2, abandonedCount, "should have 2 abandoned sessions")
}

func TestSessionWithNeedInputTask(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollectionWithAgentSession(t, app)
	SetupCommentsCollection(t, app)
	SetupSessionsCollection(t, app)

	// Create task, link session, block it
	task := CreateSessionTestTask(t, app, "Blocked Task", "in_progress")
	SimulateSessionLink(t, app, task, "opencode", "blocked-session", "/home/user/project")

	// Move to need_input (simulating block command)
	task, _ = app.FindRecordById("tasks", task.Id)
	task.Set("column", "need_input")
	require.NoError(t, app.Save(task))

	// Verify task state
	task, _ = app.FindRecordById("tasks", task.Id)
	assert.Equal(t, "need_input", task.GetString("column"))

	// Verify session is still linked
	sessionData := task.Get("agent_session")
	session, _ := output.ParseAgentSession(sessionData)
	assert.Equal(t, "blocked-session", session["ref"])
	assert.Equal(t, "opencode", session["tool"])
}
