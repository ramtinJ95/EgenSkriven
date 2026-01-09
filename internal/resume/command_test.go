package resume

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildResumeCommand_OpenCode verifies the correct command for opencode.
// Format: opencode run "<prompt>" --session <session-id>
func TestBuildResumeCommand_OpenCode(t *testing.T) {
	rc, err := BuildResumeCommand(ToolOpenCode, "abc-123-session", "/tmp/project", "Continue working")

	require.NoError(t, err, "should not return error for opencode")
	assert.NotNil(t, rc, "should return ResumeCommand")

	// Verify command contains expected parts
	assert.Contains(t, rc.Command, "opencode run",
		"command should contain 'opencode run'")
	assert.Contains(t, rc.Command, "abc-123-session",
		"command should contain session ref")
	assert.Contains(t, rc.Command, "--session",
		"command should contain '--session' flag")

	// Verify struct fields
	assert.Equal(t, ToolOpenCode, rc.Tool, "tool should be opencode")
	assert.Equal(t, "abc-123-session", rc.SessionRef, "session ref should match")
	assert.Equal(t, "/tmp/project", rc.WorkingDir, "working dir should match")
	assert.Equal(t, "Continue working", rc.Prompt, "prompt should match")

	// Verify Args for exec.Command
	assert.Equal(t, "opencode", rc.Args[0], "first arg should be 'opencode'")
	assert.Equal(t, "run", rc.Args[1], "second arg should be 'run'")
	assert.Equal(t, "Continue working", rc.Args[2], "third arg should be prompt")
	assert.Equal(t, "--session", rc.Args[3], "fourth arg should be '--session'")
	assert.Equal(t, "abc-123-session", rc.Args[4], "fifth arg should be session ref")
}

// TestBuildResumeCommand_ClaudeCode verifies the correct command for claude-code.
// Format: claude --resume <session-id> "<prompt>"
func TestBuildResumeCommand_ClaudeCode(t *testing.T) {
	rc, err := BuildResumeCommand(ToolClaudeCode, "def-456-session", "/home/user/project", "Continue working")

	require.NoError(t, err, "should not return error for claude-code")
	assert.NotNil(t, rc, "should return ResumeCommand")

	// Verify command contains expected parts
	assert.Contains(t, rc.Command, "claude --resume",
		"command should contain 'claude --resume'")
	assert.Contains(t, rc.Command, "def-456-session",
		"command should contain session ref")

	// Verify struct fields
	assert.Equal(t, ToolClaudeCode, rc.Tool, "tool should be claude-code")
	assert.Equal(t, "def-456-session", rc.SessionRef, "session ref should match")
	assert.Equal(t, "/home/user/project", rc.WorkingDir, "working dir should match")
	assert.Equal(t, "Continue working", rc.Prompt, "prompt should match")

	// Verify Args for exec.Command
	assert.Equal(t, "claude", rc.Args[0], "first arg should be 'claude'")
	assert.Equal(t, "--resume", rc.Args[1], "second arg should be '--resume'")
	assert.Equal(t, "def-456-session", rc.Args[2], "third arg should be session ref")
	assert.Equal(t, "Continue working", rc.Args[3], "fourth arg should be prompt")
}

// TestBuildResumeCommand_Codex verifies the correct command for codex.
// Format: codex exec resume <session-id> "<prompt>"
func TestBuildResumeCommand_Codex(t *testing.T) {
	rc, err := BuildResumeCommand(ToolCodex, "ghi-789-session", "/var/project", "Continue working")

	require.NoError(t, err, "should not return error for codex")
	assert.NotNil(t, rc, "should return ResumeCommand")

	// Verify command contains expected parts
	assert.Contains(t, rc.Command, "codex exec resume",
		"command should contain 'codex exec resume'")
	assert.Contains(t, rc.Command, "ghi-789-session",
		"command should contain session ref")

	// Verify struct fields
	assert.Equal(t, ToolCodex, rc.Tool, "tool should be codex")
	assert.Equal(t, "ghi-789-session", rc.SessionRef, "session ref should match")
	assert.Equal(t, "/var/project", rc.WorkingDir, "working dir should match")
	assert.Equal(t, "Continue working", rc.Prompt, "prompt should match")

	// Verify Args for exec.Command
	assert.Equal(t, "codex", rc.Args[0], "first arg should be 'codex'")
	assert.Equal(t, "exec", rc.Args[1], "second arg should be 'exec'")
	assert.Equal(t, "resume", rc.Args[2], "third arg should be 'resume'")
	assert.Equal(t, "ghi-789-session", rc.Args[3], "fourth arg should be session ref")
	assert.Equal(t, "Continue working", rc.Args[4], "fifth arg should be prompt")
}

// TestBuildResumeCommand_UnknownTool verifies that an error is returned for unknown tools.
func TestBuildResumeCommand_UnknownTool(t *testing.T) {
	rc, err := BuildResumeCommand("unknown-tool", "some-session", "/tmp", "Continue")

	assert.Error(t, err, "should return error for unknown tool")
	assert.Nil(t, rc, "should not return ResumeCommand for unknown tool")
	assert.Contains(t, err.Error(), "unsupported tool",
		"error should mention 'unsupported tool'")
	assert.Contains(t, err.Error(), "unknown-tool",
		"error should include the invalid tool name")
}

// TestBuildResumeCommand_EmptyTool verifies that an error is returned for empty tool.
func TestBuildResumeCommand_EmptyTool(t *testing.T) {
	rc, err := BuildResumeCommand("", "some-session", "/tmp", "Continue")

	assert.Error(t, err, "should return error for empty tool")
	assert.Nil(t, rc, "should not return ResumeCommand for empty tool")
}

// TestBuildResumeCommand_EscapesSingleQuotes verifies prompts with single quotes are escaped.
func TestBuildResumeCommand_EscapesSingleQuotes(t *testing.T) {
	prompt := "What's the status? Use 'single quotes' here"
	rc, err := BuildResumeCommand(ToolOpenCode, "abc-123", "/tmp", prompt)

	require.NoError(t, err, "should not return error")

	// Command should be safe to execute - single quotes should be escaped
	cmd := rc.Command

	// The escaped form should contain the escape sequence '\''
	// Single quotes in shell are escaped as: end quote, backslash-quote, start quote: '\''
	assert.Contains(t, cmd, `'\''`,
		"command should contain escape sequence for single quotes")

	// Original single quotes should NOT appear without escaping
	// The prompt has "What's" - so we should NOT see unescaped "What's"
	// Instead we should see "What'\''s"
	assert.Contains(t, cmd, "What'\\''s",
		"single quote in 'What's' should be escaped")
}

// TestBuildResumeCommand_EscapesDoubleQuotes verifies prompts with double quotes are handled.
func TestBuildResumeCommand_EscapesDoubleQuotes(t *testing.T) {
	prompt := `Use "double quotes" in the prompt`
	rc, err := BuildResumeCommand(ToolOpenCode, "abc-123", "/tmp", prompt)

	require.NoError(t, err, "should not return error")

	// The prompt should be in the command (double quotes inside single quotes are safe)
	assert.Contains(t, rc.Command, `"double quotes"`,
		"command should contain the double quotes from prompt")
}

// TestBuildResumeCommand_EscapesNewlines verifies prompts with newlines are escaped.
func TestBuildResumeCommand_EscapesNewlines(t *testing.T) {
	prompt := "Line 1\nLine 2\nLine 3"
	rc, err := BuildResumeCommand(ToolOpenCode, "abc-123", "/tmp", prompt)

	require.NoError(t, err, "should not return error")

	// Newlines should be preserved inside quoted string
	assert.Contains(t, rc.Command, "\n",
		"command should preserve newlines in prompt")
}

// TestBuildResumeCommand_EscapesSpecialShellChars verifies prompts with special shell characters.
func TestBuildResumeCommand_EscapesSpecialShellChars(t *testing.T) {
	prompt := "Test $VAR and `command` and $(subshell) and !history"
	rc, err := BuildResumeCommand(ToolOpenCode, "abc-123", "/tmp", prompt)

	require.NoError(t, err, "should not return error")

	// Inside single quotes, these special chars are literal
	// The command should contain these characters
	assert.Contains(t, rc.Command, "$VAR",
		"command should contain $VAR")
	assert.Contains(t, rc.Command, "`command`",
		"command should contain backtick command")
}

// TestBuildResumeCommand_EmptyPrompt verifies handling of empty prompt.
func TestBuildResumeCommand_EmptyPrompt(t *testing.T) {
	rc, err := BuildResumeCommand(ToolOpenCode, "abc-123", "/tmp", "")

	require.NoError(t, err, "should not return error for empty prompt")
	assert.NotNil(t, rc, "should return ResumeCommand")

	// Empty prompt should be quoted as ''
	assert.Contains(t, rc.Command, "''",
		"command should contain empty quoted string")
}

// TestValidateSessionRef_RejectsEmptyRef verifies empty session refs are rejected.
func TestValidateSessionRef_RejectsEmptyRef(t *testing.T) {
	err := ValidateSessionRef(ToolOpenCode, "")

	assert.Error(t, err, "should return error for empty ref")
	assert.Contains(t, err.Error(), "empty",
		"error should mention empty")
}

// TestValidateSessionRef_RejectsTooShortRef verifies refs shorter than 8 chars are rejected.
func TestValidateSessionRef_RejectsTooShortRef(t *testing.T) {
	shortRefs := []string{"abc", "1234567", "short"}

	for _, ref := range shortRefs {
		t.Run(ref, func(t *testing.T) {
			err := ValidateSessionRef(ToolOpenCode, ref)

			assert.Error(t, err, "should return error for short ref %q", ref)
			assert.Contains(t, err.Error(), "too short",
				"error should mention 'too short'")
		})
	}
}

// TestValidateSessionRef_AcceptsValidRef verifies valid session refs are accepted.
func TestValidateSessionRef_AcceptsValidRef(t *testing.T) {
	validRefs := []string{
		"12345678",                             // Exactly 8 chars
		"550e8400-e29b-41d4-a716-446655440000", // UUID format
		"abc123def456",                         // 12 chars
		"session-id-that-is-long-enough",       // Long enough
	}

	for _, ref := range validRefs {
		t.Run(ref, func(t *testing.T) {
			err := ValidateSessionRef(ToolOpenCode, ref)

			assert.NoError(t, err, "should not return error for valid ref %q", ref)
		})
	}
}

// TestValidateSessionRef_AllTools verifies validation works for all supported tools.
func TestValidateSessionRef_AllTools(t *testing.T) {
	tools := []string{ToolOpenCode, ToolClaudeCode, ToolCodex}
	validRef := "valid-session-ref-12345"

	for _, tool := range tools {
		t.Run(tool, func(t *testing.T) {
			err := ValidateSessionRef(tool, validRef)

			assert.NoError(t, err, "should not return error for %s with valid ref", tool)
		})
	}
}

// TestShellQuote_BasicString verifies basic string quoting.
func TestShellQuote_BasicString(t *testing.T) {
	result := ShellQuote("hello world")

	assert.Equal(t, "'hello world'", result,
		"basic string should be wrapped in single quotes")
}

// TestShellQuote_EmptyString verifies empty string quoting.
func TestShellQuote_EmptyString(t *testing.T) {
	result := ShellQuote("")

	assert.Equal(t, "''", result,
		"empty string should be quoted as ''")
}

// TestShellQuote_StringWithSingleQuotes verifies single quote escaping.
func TestShellQuote_StringWithSingleQuotes(t *testing.T) {
	result := ShellQuote("it's a test")

	// Single quotes are escaped as '\''
	assert.Equal(t, "'it'\\''s a test'", result,
		"single quotes should be escaped with '\\''")
}

// TestShellQuote_MultipleSpecialChars verifies complex string escaping.
func TestShellQuote_MultipleSpecialChars(t *testing.T) {
	// String with various special characters
	result := ShellQuote(`What's the "status" of $VAR?`)

	// Should be wrapped in single quotes
	assert.True(t, strings.HasPrefix(result, "'"),
		"result should start with single quote")
	assert.True(t, strings.HasSuffix(result, "'"),
		"result should end with single quote")

	// Single quote in "What's" should be escaped
	assert.Contains(t, result, `'\''`,
		"single quote should be escaped")
}

// TestIsValidTool_ValidTools verifies IsValidTool returns true for valid tools.
func TestIsValidTool_ValidTools(t *testing.T) {
	validTools := []string{ToolOpenCode, ToolClaudeCode, ToolCodex}

	for _, tool := range validTools {
		t.Run(tool, func(t *testing.T) {
			assert.True(t, IsValidTool(tool),
				"IsValidTool should return true for %s", tool)
		})
	}
}

// TestIsValidTool_InvalidTools verifies IsValidTool returns false for invalid tools.
func TestIsValidTool_InvalidTools(t *testing.T) {
	invalidTools := []string{"", "unknown", "vscode", "vim", "OPENCODE"}

	for _, tool := range invalidTools {
		t.Run(tool, func(t *testing.T) {
			assert.False(t, IsValidTool(tool),
				"IsValidTool should return false for %q", tool)
		})
	}
}

// TestBuildResumeCommand_CommandLineFormat verifies the generated command format is correct.
func TestBuildResumeCommand_CommandLineFormat(t *testing.T) {
	tests := []struct {
		name       string
		tool       string
		sessionRef string
		prompt     string
		wantPrefix string
	}{
		{
			name:       "opencode format",
			tool:       ToolOpenCode,
			sessionRef: "session-123",
			prompt:     "Continue",
			wantPrefix: "opencode run '",
		},
		{
			name:       "claude-code format",
			tool:       ToolClaudeCode,
			sessionRef: "session-456",
			prompt:     "Continue",
			wantPrefix: "claude --resume session-456 '",
		},
		{
			name:       "codex format",
			tool:       ToolCodex,
			sessionRef: "session-789",
			prompt:     "Continue",
			wantPrefix: "codex exec resume session-789 '",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc, err := BuildResumeCommand(tt.tool, tt.sessionRef, "/tmp", tt.prompt)

			require.NoError(t, err)
			assert.True(t, strings.HasPrefix(rc.Command, tt.wantPrefix),
				"command should start with %q, got %q", tt.wantPrefix, rc.Command)
		})
	}
}
