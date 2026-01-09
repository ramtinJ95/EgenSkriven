package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const claudeCodeHookTemplate = `#!/bin/bash
# EgenSkriven Session Registration Hook for Claude Code
#
# This hook runs on SessionStart and persists the session ID as an
# environment variable for use by the AI agent.
#
# After this hook runs, the agent can access $CLAUDE_SESSION_ID
# in any Bash tool calls.
#
# Usage by agent:
#   egenskriven session link <task-ref> --tool claude-code --ref $CLAUDE_SESSION_ID

set -e

# Read JSON input from stdin
INPUT=$(cat)

# Extract session_id (try jq first, fall back to Python)
if command -v jq &> /dev/null; then
    SESSION_ID=$(echo "$INPUT" | jq -r '.session_id // empty')
    HOOK_EVENT=$(echo "$INPUT" | jq -r '.hook_event_name // empty')
else
    SESSION_ID=$(echo "$INPUT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('session_id',''))" 2>/dev/null || echo "")
    HOOK_EVENT=$(echo "$INPUT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('hook_event_name',''))" 2>/dev/null || echo "")
fi

# Only process SessionStart events
if [ "$HOOK_EVENT" != "SessionStart" ]; then
    exit 0
fi

# Persist session ID for subsequent Bash tool calls
if [ -n "$CLAUDE_ENV_FILE" ] && [ -n "$SESSION_ID" ]; then
    echo "export CLAUDE_SESSION_ID=$SESSION_ID" >> "$CLAUDE_ENV_FILE"
fi

# Output success context (optional, will be shown to Claude)
cat << EOF
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "status": "registered",
    "session_id": "$SESSION_ID",
    "message": "Session ID available as \$CLAUDE_SESSION_ID"
  }
}
EOF

exit 0
`

// generateClaudeCodeIntegration generates Claude Code hooks for session discovery.
// Generated files: .claude/hooks/egenskriven-session.sh, .claude/settings.json
func generateClaudeCodeIntegration(force bool) ([]string, error) {
	hooksDir := ".claude/hooks"
	hookFile := filepath.Join(hooksDir, "egenskriven-session.sh")
	settingsFile := ".claude/settings.json"

	generatedFiles := []string{}

	// Create hooks directory
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", hooksDir, err)
	}

	// Check if hook file exists
	if _, err := os.Stat(hookFile); err == nil && !force {
		return nil, fmt.Errorf("file already exists: %s (use --force to overwrite)", hookFile)
	}

	// Write hook script with executable permissions
	if err := os.WriteFile(hookFile, []byte(claudeCodeHookTemplate), 0755); err != nil {
		return nil, fmt.Errorf("failed to write %s: %w", hookFile, err)
	}
	generatedFiles = append(generatedFiles, hookFile)

	// Update or create settings.json
	settings := loadClaudeSettings(settingsFile)
	settings = mergeClaudeHooks(settings)

	settingsJSON, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsFile, settingsJSON, 0644); err != nil {
		return nil, fmt.Errorf("failed to write %s: %w", settingsFile, err)
	}
	generatedFiles = append(generatedFiles, settingsFile)

	return generatedFiles, nil
}

// loadClaudeSettings loads existing Claude settings from the given path,
// or returns an empty map if the file doesn't exist.
func loadClaudeSettings(path string) map[string]any {
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]any{}
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		return map[string]any{}
	}

	return settings
}

// mergeClaudeHooks merges the EgenSkriven SessionStart hook into the settings,
// preserving any existing hooks and settings.
func mergeClaudeHooks(settings map[string]any) map[string]any {
	// Get or create hooks section
	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		hooks = map[string]any{}
	}

	// Define our SessionStart hook
	egenskrivenHook := map[string]any{
		"matcher": "startup|resume",
		"hooks": []map[string]any{
			{
				"type":    "command",
				"command": `bash "$CLAUDE_PROJECT_DIR/.claude/hooks/egenskriven-session.sh"`,
			},
		},
	}

	// Check if we already have SessionStart hooks
	existing, hasExisting := hooks["SessionStart"].([]any)
	if hasExisting {
		// Check if our hook is already present
		alreadyPresent := false
		for _, h := range existing {
			if hook, ok := h.(map[string]any); ok {
				if hooksList, ok := hook["hooks"].([]any); ok {
					for _, hh := range hooksList {
						if hhMap, ok := hh.(map[string]any); ok {
							if cmd, ok := hhMap["command"].(string); ok {
								if strings.Contains(cmd, "egenskriven-session.sh") {
									alreadyPresent = true
									break
								}
							}
						}
					}
				}
			}
		}

		if !alreadyPresent {
			// Append our hook to existing SessionStart hooks
			existing = append(existing, egenskrivenHook)
			hooks["SessionStart"] = existing
		}
	} else {
		// No existing SessionStart hooks, create new
		hooks["SessionStart"] = []any{egenskrivenHook}
	}

	settings["hooks"] = hooks
	return settings
}
