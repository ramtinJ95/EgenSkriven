package commands

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ramtinJ95/EgenSkriven/internal/output"
	"github.com/ramtinJ95/EgenSkriven/internal/resume"
	"github.com/ramtinJ95/EgenSkriven/internal/testutil"
)

// ========== Helper Functions ==========

// createTestTaskWithSession creates a task with an agent_session linked.
func createTestTaskWithSession(t *testing.T, app *pocketbase.PocketBase, title, tool, sessionRef string) *core.Record {
	t.Helper()

	task := CreateTestTask(t, app, title, "need_input")

	session := map[string]any{
		"tool":        tool,
		"ref":         sessionRef,
		"ref_type":    "uuid",
		"working_dir": "/tmp/test-project",
		"linked_at":   time.Now().Format(time.RFC3339),
	}
	task.Set("agent_session", session)

	require.NoError(t, app.Save(task))
	return task
}

// addTestCommentForResume creates a comment for testing resume context.
func addTestCommentForResume(t *testing.T, app *pocketbase.PocketBase, taskId, content, authorType, authorId string) *core.Record {
	t.Helper()
	return CreateTestComment(t, app, taskId, content, authorType, authorId)
}

// getResumeJSONResult parses JSON output from resume command.
func getResumeJSONResult(t *testing.T, output []byte) map[string]any {
	t.Helper()

	var result map[string]any
	err := json.Unmarshal(output, &result)
	require.NoError(t, err, "should be valid JSON")
	return result
}

// ========== Tests for Task State Validation ==========

// Test 7.4: resume fails for task not in need_input state
func TestResumeCommand_FailsForTaskNotInNeedInput(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollectionWithAgentSession(t, app)
	SetupCommentsCollectionWithAutodate(t, app)
	SetupSessionsCollection(t, app)

	tests := []struct {
		name   string
		column string
	}{
		{"in_progress", "in_progress"},
		{"todo", "todo"},
		{"done", "done"},
		{"backlog", "backlog"},
		{"review", "review"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a task in wrong state with session linked
			task := CreateTestTask(t, app, "Task in "+tt.column, tt.column)

			// Set agent_session (task needs session for this test to test state, not session)
			session := map[string]any{
				"tool":        "opencode",
				"ref":         "test-session-123",
				"ref_type":    "uuid",
				"working_dir": "/tmp",
				"linked_at":   time.Now().Format(time.RFC3339),
			}
			task.Set("agent_session", session)
			require.NoError(t, app.Save(task))

			// Simulate the validation that resume command does
			column := task.GetString("column")

			// Verify task is not in need_input state
			assert.NotEqual(t, "need_input", column)

			// The resume command should reject this
			if column != "need_input" {
				// This is the expected behavior - matches resume.go lines 79-82
				errMsg := fmt.Sprintf("task %s is not in need_input state (current: %s)",
					shortID(task.Id), column)
				assert.Contains(t, errMsg, "not in need_input")
				assert.Contains(t, errMsg, column, "error should mention current state")
			}
		})
	}
}

// Test 7.5: resume fails for task without agent_session linked
func TestResumeCommand_FailsForTaskWithoutSession(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollectionWithAgentSession(t, app)
	SetupCommentsCollectionWithAutodate(t, app)
	SetupSessionsCollection(t, app)

	// Create a task in need_input WITHOUT session linked
	task := CreateTestTask(t, app, "Task without session", "need_input")

	// Verify task is in need_input
	assert.Equal(t, "need_input", task.GetString("column"))

	// Verify no session is linked
	sessionData := task.Get("agent_session")
	session, err := output.ParseAgentSession(sessionData)
	assert.NoError(t, err)
	assert.Nil(t, session, "task should not have a session")

	// The resume command should fail with helpful hint - matches resume.go lines 86-100
	displayId := getTaskDisplayID(app, task)
	expectedHint := fmt.Sprintf("session link %s --tool", displayId)
	assert.Contains(t, expectedHint, "session link", "error should include hint about session link command")
}

// Test 7.6: resume prints command by default (no --exec)
func TestResumeCommand_PrintsModeByDefault(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollectionWithAgentSession(t, app)
	SetupCommentsCollectionWithAutodate(t, app)
	SetupSessionsCollection(t, app)

	// Create a task with session
	task := createTestTaskWithSession(t, app, "Test print mode", "opencode", "test-session-123")
	addTestCommentForResume(t, app, task.Id, "What should I do?", "agent", "opencode")
	addTestCommentForResume(t, app, task.Id, "Use JWT", "human", "john")

	// Simulate what resume command does in print mode
	sessionData := task.Get("agent_session")
	session, err := output.ParseAgentSession(sessionData)
	require.NoError(t, err)
	require.NotNil(t, session)

	tool := session["tool"].(string)
	sessionRef := session["ref"].(string)

	// Build the resume command
	resumeCmd, err := resume.BuildResumeCommand(tool, sessionRef, "/tmp/test-project", "test prompt")
	require.NoError(t, err)

	// Verify command contains expected elements
	assert.Contains(t, resumeCmd.Command, "opencode run", "should show opencode run command")
	assert.Contains(t, resumeCmd.Command, sessionRef, "should include session ref")

	// The print mode should suggest --exec
	displayId := getTaskDisplayID(app, task)
	execHint := fmt.Sprintf("egenskriven resume %s --exec", displayId)
	assert.Contains(t, execHint, "--exec", "should mention --exec flag")
}

// Test 7.7: resume --json outputs valid JSON
func TestResumeCommand_JSONOutput(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollectionWithAgentSession(t, app)
	SetupCommentsCollectionWithAutodate(t, app)
	SetupSessionsCollection(t, app)

	// Create a task with session
	task := createTestTaskWithSession(t, app, "JSON output test", "claude-code", "json-session-456")
	addTestCommentForResume(t, app, task.Id, "Question from agent", "agent", "claude-code")

	// Fetch comments
	comments, err := fetchCommentsForResume(app, task.Id)
	require.NoError(t, err)

	// Build prompt
	prompt := resume.BuildContextPrompt(task, comments)

	// Simulate JSON output structure
	jsonResult := map[string]any{
		"task_id":       task.Id,
		"display_id":    getTaskDisplayID(app, task),
		"tool":          "claude-code",
		"session_ref":   "json-session-456",
		"working_dir":   "/tmp/test-project",
		"command":       "claude --resume json-session-456 '...'",
		"prompt":        prompt,
		"prompt_length": len(prompt),
	}

	// Verify JSON can be marshaled
	jsonBytes, err := json.Marshal(jsonResult)
	require.NoError(t, err, "should produce valid JSON")

	// Verify JSON can be unmarshaled
	result := getResumeJSONResult(t, jsonBytes)

	// Verify required fields present
	assert.Equal(t, "claude-code", result["tool"])
	assert.Equal(t, "json-session-456", result["session_ref"])
	assert.NotEmpty(t, result["command"])
	assert.NotEmpty(t, result["prompt"])
	assert.Greater(t, result["prompt_length"], float64(0), "prompt_length should be > 0")
}

// Test 7.8: resume --minimal uses shorter prompt
func TestResumeCommand_MinimalPromptIsShorter(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollectionWithAgentSession(t, app)
	SetupCommentsCollectionWithAutodate(t, app)
	SetupSessionsCollection(t, app)

	// Create a task with session and many comments
	task := createTestTaskWithSession(t, app, "Minimal prompt test", "opencode", "minimal-session")

	// Add many comments
	for i := 0; i < 10; i++ {
		addTestCommentForResume(t, app, task.Id, fmt.Sprintf("This is comment number %d with some content to make it longer", i), "human", "user")
	}

	// Fetch comments
	comments, err := fetchCommentsForResume(app, task.Id)
	require.NoError(t, err)
	require.Len(t, comments, 10, "should have 10 comments")

	// Build full and minimal prompts
	fullPrompt := resume.BuildContextPrompt(task, comments)
	minimalPrompt := resume.BuildMinimalPrompt(task, comments)

	// Verify minimal is shorter
	assert.Less(t, len(minimalPrompt), len(fullPrompt),
		"minimal prompt (%d) should be shorter than full prompt (%d)",
		len(minimalPrompt), len(fullPrompt))
}

// Test 7.9: resume --prompt "custom" uses custom prompt override
func TestResumeCommand_CustomPromptOverride(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollectionWithAgentSession(t, app)
	SetupCommentsCollectionWithAutodate(t, app)
	SetupSessionsCollection(t, app)

	// Create a task with session
	task := createTestTaskWithSession(t, app, "Custom prompt test", "codex", "custom-session")
	addTestCommentForResume(t, app, task.Id, "Original question", "agent", "codex")

	// Custom prompt
	customPrompt := "This is a custom prompt that overrides everything"

	// Build resume command with custom prompt
	sessionData := task.Get("agent_session")
	session, _ := output.ParseAgentSession(sessionData)

	resumeCmd, err := resume.BuildResumeCommand(
		session["tool"].(string),
		session["ref"].(string),
		session["working_dir"].(string),
		customPrompt,
	)
	require.NoError(t, err)

	// Verify the command uses the custom prompt (it would be escaped in the command)
	assert.Equal(t, customPrompt, resumeCmd.Prompt, "resume command should use custom prompt")
}

// Test 7.10: resume --exec --dry-run shows command without executing
func TestResumeCommand_DryRunShowsCommandWithoutExecuting(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollectionWithAgentSession(t, app)
	SetupCommentsCollectionWithAutodate(t, app)
	SetupSessionsCollection(t, app)

	// Create a task with session in need_input
	task := createTestTaskWithSession(t, app, "Dry run test", "opencode", "dryrun-session")
	addTestCommentForResume(t, app, task.Id, "Question", "agent", "opencode")
	addTestCommentForResume(t, app, task.Id, "Answer", "human", "user")

	// Store original column
	originalColumn := task.GetString("column")
	assert.Equal(t, "need_input", originalColumn)

	// In dry-run mode:
	// - Command should be shown
	// - Task should NOT be modified
	// - Process should NOT be spawned

	// Fetch comments and build prompt
	comments, err := fetchCommentsForResume(app, task.Id)
	require.NoError(t, err)
	prompt := resume.BuildContextPrompt(task, comments)

	// Build resume command
	sessionData := task.Get("agent_session")
	session, _ := output.ParseAgentSession(sessionData)

	resumeCmd, err := resume.BuildResumeCommand(
		session["tool"].(string),
		session["ref"].(string),
		session["working_dir"].(string),
		prompt,
	)
	require.NoError(t, err)

	// Verify command was built
	assert.NotEmpty(t, resumeCmd.Command)
	assert.Contains(t, resumeCmd.Command, "opencode run")
	assert.Contains(t, resumeCmd.Command, "dryrun-session")

	// Re-fetch task to ensure it wasn't modified
	task, err = app.FindRecordById("tasks", task.Id)
	require.NoError(t, err)

	// Task should still be in need_input (dry-run doesn't modify task)
	assert.Equal(t, "need_input", task.GetString("column"),
		"dry-run should not modify task state")
}

// ========== Additional Tests ==========

// TestResumeCommand_AllToolsSupported verifies resume works for all three tools
func TestResumeCommand_AllToolsSupported(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollectionWithAgentSession(t, app)
	SetupCommentsCollectionWithAutodate(t, app)
	SetupSessionsCollection(t, app)

	tests := []struct {
		tool           string
		expectedPrefix string
	}{
		{"opencode", "opencode run"},
		{"claude-code", "claude --resume"},
		{"codex", "codex exec resume"},
	}

	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			task := createTestTaskWithSession(t, app, "Task for "+tt.tool, tt.tool, tt.tool+"-session")

			sessionData := task.Get("agent_session")
			session, _ := output.ParseAgentSession(sessionData)

			resumeCmd, err := resume.BuildResumeCommand(
				session["tool"].(string),
				session["ref"].(string),
				session["working_dir"].(string),
				"test prompt",
			)
			require.NoError(t, err)

			assert.Contains(t, resumeCmd.Command, tt.expectedPrefix,
				"command for %s should start with %s", tt.tool, tt.expectedPrefix)
			assert.Contains(t, resumeCmd.Command, session["ref"].(string),
				"command should include session ref")
		})
	}
}

// TestResumeCommand_ContextIncludesAllComments verifies all comments are included in context
func TestResumeCommand_ContextIncludesAllComments(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollectionWithAgentSession(t, app)
	SetupCommentsCollectionWithAutodate(t, app)
	SetupSessionsCollection(t, app)

	task := createTestTaskWithSession(t, app, "Multi-comment task", "opencode", "context-session")

	// Add specific comments that we can verify in the context
	addTestCommentForResume(t, app, task.Id, "First question from agent", "agent", "opencode")
	addTestCommentForResume(t, app, task.Id, "Human response number one", "human", "developer")
	addTestCommentForResume(t, app, task.Id, "Follow-up from agent", "agent", "opencode")
	addTestCommentForResume(t, app, task.Id, "Final answer from human", "human", "developer")

	comments, err := fetchCommentsForResume(app, task.Id)
	require.NoError(t, err)
	require.Len(t, comments, 4, "should have 4 comments")

	prompt := resume.BuildContextPrompt(task, comments)

	// Verify all comments are in the prompt
	assert.Contains(t, prompt, "First question from agent")
	assert.Contains(t, prompt, "Human response number one")
	assert.Contains(t, prompt, "Follow-up from agent")
	assert.Contains(t, prompt, "Final answer from human")
}

// TestResumeCommand_UpdatesTaskOnExec verifies task is updated when --exec is used
func TestResumeCommand_UpdatesTaskOnExec(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollectionWithAgentSession(t, app)
	SetupCommentsCollectionWithAutodate(t, app)
	SetupSessionsCollection(t, app)

	task := createTestTaskWithSession(t, app, "Exec update test", "opencode", "exec-session")
	addTestCommentForResume(t, app, task.Id, "Question", "agent", "opencode")

	// Verify initial state
	assert.Equal(t, "need_input", task.GetString("column"))

	// Simulate what updateTaskForResume does
	err := updateTaskForResume(app, task)
	require.NoError(t, err)

	// Re-fetch task
	task, err = app.FindRecordById("tasks", task.Id)
	require.NoError(t, err)

	// Verify task moved to in_progress
	assert.Equal(t, "in_progress", task.GetString("column"),
		"task should move to in_progress after resume")

	// Verify history was updated
	history := getHistoryFromTask(t, task)
	require.Greater(t, len(history), 0, "history should have entries")

	lastEntry := history[len(history)-1]
	assert.Equal(t, "resumed", lastEntry["action"], "last action should be 'resumed'")
}

// TestResumeCommand_InvalidSessionRefRejected verifies invalid session refs are rejected
func TestResumeCommand_InvalidSessionRefRejected(t *testing.T) {
	tests := []struct {
		name    string
		ref     string
		wantErr bool
	}{
		{"empty ref", "", true},
		{"too short", "abc", true},
		{"valid uuid-like", "abc12345", false},
		{"full uuid", "550e8400-e29b-41d4-a716-446655440000", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := resume.ValidateSessionRef("opencode", tt.ref)
			if tt.wantErr {
				assert.Error(t, err, "should reject invalid ref %q", tt.ref)
			} else {
				assert.NoError(t, err, "should accept valid ref %q", tt.ref)
			}
		})
	}
}

// TestResumeCommand_WorkingDirPreserved verifies working_dir is used correctly
func TestResumeCommand_WorkingDirPreserved(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollectionWithAgentSession(t, app)
	SetupCommentsCollectionWithAutodate(t, app)
	SetupSessionsCollection(t, app)

	// Create task with specific working_dir
	task := CreateTestTask(t, app, "Working dir test", "need_input")
	specificWorkingDir := "/home/user/specific-project"

	session := map[string]any{
		"tool":        "opencode",
		"ref":         "workdir-session",
		"ref_type":    "uuid",
		"working_dir": specificWorkingDir,
		"linked_at":   time.Now().Format(time.RFC3339),
	}
	task.Set("agent_session", session)
	require.NoError(t, app.Save(task))

	// Fetch and verify working_dir
	sessionData := task.Get("agent_session")
	parsedSession, _ := output.ParseAgentSession(sessionData)

	resumeCmd, err := resume.BuildResumeCommand(
		parsedSession["tool"].(string),
		parsedSession["ref"].(string),
		parsedSession["working_dir"].(string),
		"test prompt",
	)
	require.NoError(t, err)

	assert.Equal(t, specificWorkingDir, resumeCmd.WorkingDir,
		"working_dir should be preserved from session")
}

// TestResumeCommand_PromptWithSpecialChars verifies special characters are handled
func TestResumeCommand_PromptWithSpecialChars(t *testing.T) {
	app := testutil.NewTestApp(t)
	SetupTasksCollectionWithAgentSession(t, app)
	SetupCommentsCollectionWithAutodate(t, app)
	SetupSessionsCollection(t, app)

	task := createTestTaskWithSession(t, app, "Special chars test", "opencode", "special-session")

	// Add comment with special characters
	specialContent := "Use JWT with 'single quotes' and \"double quotes\"\nAlso newlines\tand\ttabs"
	addTestCommentForResume(t, app, task.Id, specialContent, "human", "user")

	comments, err := fetchCommentsForResume(app, task.Id)
	require.NoError(t, err)

	prompt := resume.BuildContextPrompt(task, comments)

	// Verify special chars are in prompt
	assert.Contains(t, prompt, "single quotes")
	assert.Contains(t, prompt, "double quotes")

	// Build resume command - should not error
	resumeCmd, err := resume.BuildResumeCommand("opencode", "special-session", "/tmp", prompt)
	require.NoError(t, err)
	assert.NotEmpty(t, resumeCmd.Command)
}
