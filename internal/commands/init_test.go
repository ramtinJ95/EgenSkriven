package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateOpenCodeIntegration(t *testing.T) {
	// Create temp directory and change to it
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(originalDir) }()

	// Generate
	files, err := generateOpenCodeIntegration(false)
	require.NoError(t, err, "unexpected error")

	// Verify file created
	require.Len(t, files, 1, "expected 1 file")

	expectedFile := ".opencode/tool/egenskriven-session.ts"
	assert.Equal(t, expectedFile, files[0], "expected correct file path")

	// Verify file exists and has content
	content, err := os.ReadFile(expectedFile)
	require.NoError(t, err, "failed to read file")
	require.NotEmpty(t, content, "file should not be empty")

	// Verify key content
	contentStr := string(content)
	// The template destructures sessionID from context: const { sessionID, messageID, agent } = context
	assert.Contains(t, contentStr, "sessionID", "should reference sessionID")
	assert.Contains(t, contentStr, "context", "should use context object")
	assert.Contains(t, contentStr, "egenskriven session link", "should include link command")
	assert.Contains(t, contentStr, "@opencode-ai/plugin", "should import from plugin")
	assert.Contains(t, contentStr, "export default tool", "should export tool")
}

func TestGenerateOpenCodeIntegrationNoOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(originalDir) }()

	// Generate first time
	_, err = generateOpenCodeIntegration(false)
	require.NoError(t, err, "first generation should succeed")

	// Try to generate again without force
	_, err = generateOpenCodeIntegration(false)
	require.Error(t, err, "expected error when file exists")
	assert.Contains(t, err.Error(), "already exists", "error should mention file exists")
	assert.Contains(t, err.Error(), "--force", "error should mention --force flag")

	// Generate with force should succeed
	files, err := generateOpenCodeIntegration(true)
	require.NoError(t, err, "should succeed with force")
	require.Len(t, files, 1, "should return generated file")
}

func TestGenerateClaudeCodeIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(originalDir) }()

	files, err := generateClaudeCodeIntegration(false)
	require.NoError(t, err, "unexpected error")

	// Should create hook script and settings
	require.Len(t, files, 2, "expected 2 files")

	// Verify hook script
	hookScript := ".claude/hooks/egenskriven-session.sh"
	assert.Contains(t, files, hookScript, "should create hook script")

	content, err := os.ReadFile(hookScript)
	require.NoError(t, err, "failed to read hook script")
	contentStr := string(content)
	assert.Contains(t, contentStr, "CLAUDE_SESSION_ID", "hook should reference CLAUDE_SESSION_ID")
	assert.Contains(t, contentStr, "CLAUDE_ENV_FILE", "hook should reference CLAUDE_ENV_FILE")
	assert.Contains(t, contentStr, "jq", "hook should support jq")
	assert.Contains(t, contentStr, "python3", "hook should support python fallback")

	// Verify script is executable
	info, err := os.Stat(hookScript)
	require.NoError(t, err)
	assert.NotEqual(t, os.FileMode(0), info.Mode()&0111, "hook script should be executable")

	// Verify settings.json
	settingsFile := ".claude/settings.json"
	assert.Contains(t, files, settingsFile, "should create settings file")

	settingsContent, err := os.ReadFile(settingsFile)
	require.NoError(t, err, "failed to read settings")
	assert.Contains(t, string(settingsContent), "SessionStart", "settings should include SessionStart hook")
	assert.Contains(t, string(settingsContent), "egenskriven-session.sh", "settings should reference our hook")
}

func TestGenerateClaudeCodeIntegrationNoOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(originalDir) }()

	// Generate first time
	_, err = generateClaudeCodeIntegration(false)
	require.NoError(t, err, "first generation should succeed")

	// Try to generate again without force
	_, err = generateClaudeCodeIntegration(false)
	require.Error(t, err, "expected error when file exists")
	assert.Contains(t, err.Error(), "already exists", "error should mention file exists")

	// Generate with force should succeed
	files, err := generateClaudeCodeIntegration(true)
	require.NoError(t, err, "should succeed with force")
	require.Len(t, files, 2, "should return generated files")
}

func TestClaudeSettingsMerge(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(originalDir) }()

	// Create existing settings with other configuration
	require.NoError(t, os.MkdirAll(".claude", 0755))
	existingSettings := `{
  "someExistingSetting": true,
  "anotherSetting": "value",
  "hooks": {
    "SomeOtherHook": [{"matcher": "*", "hooks": []}]
  }
}`
	require.NoError(t, os.WriteFile(".claude/settings.json", []byte(existingSettings), 0644))

	// Generate integration
	_, err = generateClaudeCodeIntegration(true) // force to overwrite hook file
	require.NoError(t, err)

	// Verify settings were merged, not replaced
	content, err := os.ReadFile(".claude/settings.json")
	require.NoError(t, err)
	contentStr := string(content)

	assert.Contains(t, contentStr, "someExistingSetting", "should preserve existing settings")
	assert.Contains(t, contentStr, "anotherSetting", "should preserve other settings")
	assert.Contains(t, contentStr, "SomeOtherHook", "should preserve existing hooks")
	assert.Contains(t, contentStr, "SessionStart", "should add SessionStart hook")
	assert.Contains(t, contentStr, "egenskriven-session.sh", "should add our hook command")

	// Parse and verify structure
	var settings map[string]any
	require.NoError(t, json.Unmarshal(content, &settings))

	// Verify existing settings preserved
	assert.Equal(t, true, settings["someExistingSetting"])
	assert.Equal(t, "value", settings["anotherSetting"])

	// Verify hooks structure
	hooks, ok := settings["hooks"].(map[string]any)
	require.True(t, ok, "hooks should be a map")
	assert.Contains(t, hooks, "SomeOtherHook", "should have existing hook")
	assert.Contains(t, hooks, "SessionStart", "should have SessionStart hook")
}

func TestClaudeSettingsMergeIdempotent(t *testing.T) {
	// Test that running merge multiple times doesn't duplicate hooks
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(originalDir) }()

	// Generate once
	_, err = generateClaudeCodeIntegration(false)
	require.NoError(t, err)

	// Read settings after first generation
	content1, err := os.ReadFile(".claude/settings.json")
	require.NoError(t, err)

	// Generate again with force
	_, err = generateClaudeCodeIntegration(true)
	require.NoError(t, err)

	// Read settings after second generation
	content2, err := os.ReadFile(".claude/settings.json")
	require.NoError(t, err)

	// Count occurrences of our hook command
	count1 := strings.Count(string(content1), "egenskriven-session.sh")
	count2 := strings.Count(string(content2), "egenskriven-session.sh")

	assert.Equal(t, count1, count2, "should not duplicate hook on re-run")
	assert.Equal(t, 1, count2, "should have exactly one reference to our hook")
}

func TestGenerateCodexIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(originalDir) }()

	files, err := generateCodexIntegration(false)
	require.NoError(t, err, "unexpected error")

	require.Len(t, files, 1, "expected 1 file")

	// Verify helper script
	helperScript := ".codex/get-session-id.sh"
	assert.Equal(t, helperScript, files[0])

	content, err := os.ReadFile(helperScript)
	require.NoError(t, err, "failed to read helper script")

	contentStr := string(content)
	assert.Contains(t, contentStr, "CODEX_HOME", "should reference CODEX_HOME")
	assert.Contains(t, contentStr, "rollout-", "should look for rollout files")
	assert.Contains(t, contentStr, "sessions", "should reference sessions directory")
	assert.Contains(t, contentStr, "grep -oE", "should use grep with extended regex for macOS compatibility")

	// Verify executable
	info, err := os.Stat(helperScript)
	require.NoError(t, err)
	assert.NotEqual(t, os.FileMode(0), info.Mode()&0111, "helper script should be executable")
}

func TestGenerateCodexIntegrationNoOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(originalDir) }()

	// Generate first time
	_, err = generateCodexIntegration(false)
	require.NoError(t, err, "first generation should succeed")

	// Try to generate again without force
	_, err = generateCodexIntegration(false)
	require.Error(t, err, "expected error when file exists")
	assert.Contains(t, err.Error(), "already exists", "error should mention file exists")

	// Generate with force should succeed
	files, err := generateCodexIntegration(true)
	require.NoError(t, err, "should succeed with force")
	require.Len(t, files, 1, "should return generated file")
}

func TestInitIntegrationsAllFlag(t *testing.T) {
	// Test that calling all three generators creates all expected files
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(originalDir) }()

	// Simulate --all flag behavior by calling all generators
	var allGenerated []string

	files, err := generateOpenCodeIntegration(false)
	require.NoError(t, err)
	allGenerated = append(allGenerated, files...)

	files, err = generateClaudeCodeIntegration(false)
	require.NoError(t, err)
	allGenerated = append(allGenerated, files...)

	files, err = generateCodexIntegration(false)
	require.NoError(t, err)
	allGenerated = append(allGenerated, files...)

	// Verify all integrations were created
	expectedFiles := []string{
		".opencode/tool/egenskriven-session.ts",
		".claude/hooks/egenskriven-session.sh",
		".claude/settings.json",
		".codex/get-session-id.sh",
	}

	for _, f := range expectedFiles {
		assert.Contains(t, allGenerated, f, "should have generated %s", f)

		// Verify file actually exists
		_, err := os.Stat(f)
		assert.NoError(t, err, "expected file should exist: %s", f)
	}
}

func TestGenerateOpenCodeIntegrationDirectoryCreation(t *testing.T) {
	// Test that directories are created with correct permissions
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(originalDir) }()

	_, err = generateOpenCodeIntegration(false)
	require.NoError(t, err)

	// Verify directory structure
	info, err := os.Stat(".opencode")
	require.NoError(t, err)
	assert.True(t, info.IsDir(), ".opencode should be a directory")

	info, err = os.Stat(".opencode/tool")
	require.NoError(t, err)
	assert.True(t, info.IsDir(), ".opencode/tool should be a directory")
}

func TestGenerateClaudeCodeIntegrationDirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(originalDir) }()

	_, err = generateClaudeCodeIntegration(false)
	require.NoError(t, err)

	// Verify directory structure
	info, err := os.Stat(".claude")
	require.NoError(t, err)
	assert.True(t, info.IsDir(), ".claude should be a directory")

	info, err = os.Stat(".claude/hooks")
	require.NoError(t, err)
	assert.True(t, info.IsDir(), ".claude/hooks should be a directory")
}

func TestGenerateCodexIntegrationDirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(originalDir) }()

	_, err = generateCodexIntegration(false)
	require.NoError(t, err)

	// Verify directory structure
	info, err := os.Stat(".codex")
	require.NoError(t, err)
	assert.True(t, info.IsDir(), ".codex should be a directory")
}

func TestLoadClaudeSettings_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(originalDir) }()

	// Load non-existent file should return empty map
	settings := loadClaudeSettings("non-existent.json")
	assert.NotNil(t, settings, "should return non-nil map")
	assert.Empty(t, settings, "should return empty map for non-existent file")
}

func TestLoadClaudeSettings_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(originalDir) }()

	// Create file with invalid JSON
	require.NoError(t, os.WriteFile("invalid.json", []byte("not valid json{"), 0644))

	settings := loadClaudeSettings("invalid.json")
	assert.NotNil(t, settings, "should return non-nil map")
	assert.Empty(t, settings, "should return empty map for invalid JSON")
}

func TestLoadClaudeSettings_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(originalDir) }()

	// Create valid settings file
	validJSON := `{"key": "value", "number": 42}`
	require.NoError(t, os.WriteFile("valid.json", []byte(validJSON), 0644))

	settings := loadClaudeSettings("valid.json")
	assert.Equal(t, "value", settings["key"])
	assert.Equal(t, float64(42), settings["number"])
}

func TestMergeClaudeHooks_EmptySettings(t *testing.T) {
	settings := map[string]any{}
	result := mergeClaudeHooks(settings)

	hooks, ok := result["hooks"].(map[string]any)
	require.True(t, ok, "should have hooks section")

	sessionStart, ok := hooks["SessionStart"].([]any)
	require.True(t, ok, "should have SessionStart hooks")
	require.Len(t, sessionStart, 1, "should have one SessionStart hook")
}

func TestMergeClaudeHooks_ExistingDifferentHooks(t *testing.T) {
	settings := map[string]any{
		"hooks": map[string]any{
			"PreCommit": []any{
				map[string]any{"matcher": "*", "hooks": []any{}},
			},
		},
	}
	result := mergeClaudeHooks(settings)

	hooks := result["hooks"].(map[string]any)

	// Should preserve existing PreCommit hook
	_, ok := hooks["PreCommit"]
	assert.True(t, ok, "should preserve PreCommit hook")

	// Should add SessionStart hook
	_, ok = hooks["SessionStart"]
	assert.True(t, ok, "should add SessionStart hook")
}

func TestOpenCodeToolTemplateContent(t *testing.T) {
	// Verify template has correct structure
	assert.Contains(t, openCodeToolTemplate, "import { tool }", "should have tool import")
	assert.Contains(t, openCodeToolTemplate, "description:", "should have description")
	assert.Contains(t, openCodeToolTemplate, "args: {}", "should have empty args")
	assert.Contains(t, openCodeToolTemplate, "async execute", "should have async execute function")
	assert.Contains(t, openCodeToolTemplate, "JSON.stringify", "should stringify result")
}

func TestClaudeCodeHookTemplateContent(t *testing.T) {
	// Verify template has correct structure
	assert.Contains(t, claudeCodeHookTemplate, "#!/bin/bash", "should have bash shebang")
	assert.Contains(t, claudeCodeHookTemplate, "set -e", "should have error exit")
	assert.Contains(t, claudeCodeHookTemplate, "SessionStart", "should check for SessionStart event")
	assert.Contains(t, claudeCodeHookTemplate, "hookSpecificOutput", "should output hook specific data")
}

func TestCodexHelperTemplateContent(t *testing.T) {
	// Verify template has correct structure
	assert.Contains(t, codexHelperTemplate, "#!/bin/bash", "should have bash shebang")
	assert.Contains(t, codexHelperTemplate, "set -e", "should have error exit")
	assert.Contains(t, codexHelperTemplate, "CODEX_HOME", "should use CODEX_HOME env var")
	assert.Contains(t, codexHelperTemplate, "ls -t", "should sort by modification time")
	assert.Contains(t, codexHelperTemplate, "head -1", "should get most recent")
}

// TestFilePermissions verifies that generated scripts have correct permissions
func TestFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(originalDir) }()

	// Generate all integrations
	_, err = generateOpenCodeIntegration(false)
	require.NoError(t, err)
	_, err = generateClaudeCodeIntegration(false)
	require.NoError(t, err)
	_, err = generateCodexIntegration(false)
	require.NoError(t, err)

	// TypeScript file should be readable (0644)
	tsInfo, err := os.Stat(".opencode/tool/egenskriven-session.ts")
	require.NoError(t, err)
	tsPerms := tsInfo.Mode().Perm()
	// Check it's not executable
	assert.Equal(t, os.FileMode(0), tsPerms&0111, "TypeScript file should not be executable")

	// Bash scripts should be executable (0755)
	claudeInfo, err := os.Stat(".claude/hooks/egenskriven-session.sh")
	require.NoError(t, err)
	assert.NotEqual(t, os.FileMode(0), claudeInfo.Mode().Perm()&0111, "Claude hook should be executable")

	codexInfo, err := os.Stat(".codex/get-session-id.sh")
	require.NoError(t, err)
	assert.NotEqual(t, os.FileMode(0), codexInfo.Mode().Perm()&0111, "Codex script should be executable")
}

// TestFilepathJoin ensures filepath operations work correctly
func TestFilepathHandling(t *testing.T) {
	// Verify the expected file paths
	opencodePath := filepath.Join(".opencode", "tool", "egenskriven-session.ts")
	assert.Equal(t, ".opencode/tool/egenskriven-session.ts", opencodePath)

	claudePath := filepath.Join(".claude", "hooks", "egenskriven-session.sh")
	assert.Equal(t, ".claude/hooks/egenskriven-session.sh", claudePath)

	codexPath := filepath.Join(".codex", "get-session-id.sh")
	assert.Equal(t, ".codex/get-session-id.sh", codexPath)
}
