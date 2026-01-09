# Phase 4: Tool Integrations

> **Parent Document**: [ai-workflow-plan.md](./ai-workflow-plan.md)  
> **Phase**: 4 of 7  
> **Status**: Not Started  
> **Estimated Effort**: 2-3 days  
> **Prerequisites**: [Phase 3](./ai-workflow-phase-3.md) completed

## Overview

This phase implements the tool-specific integrations that allow AI agents to discover their session IDs. Each tool has a different mechanism for this, and we'll generate the necessary files locally in the user's project.

**What we're building:**
- `egenskriven init --opencode` - Generate OpenCode custom tool
- `egenskriven init --claude-code` - Generate Claude Code hooks
- `egenskriven init --codex` - Generate Codex helper script
- `egenskriven init --all` - Generate all integrations

**What we're NOT building yet (out of scope for this phase):**
- Web UI (comments panel, resume button, session info display)
- Auto-resume on @agent mention

---

## Prerequisites

Before starting this phase:

1. Phase 3 is complete (resume command works)
2. Understand how each tool exposes session IDs:
   - OpenCode: Custom tools receive `context.sessionID`
   - Claude Code: Hooks receive `session_id` in stdin JSON
   - Codex: Parse from rollout filename

---

## Session Discovery Mechanisms

### OpenCode

OpenCode custom tools receive a context object with `sessionID`:

```typescript
export default tool({
  async execute(args, context) {
    const sessionId = context.sessionID;
    // ...
  },
})
```

### Claude Code

Claude Code hooks receive JSON via stdin with `session_id`:

```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "hook_event_name": "SessionStart",
  ...
}
```

The hook can persist this to `CLAUDE_ENV_FILE` for use by the agent.

### Codex

Codex doesn't expose session ID directly. Workaround: parse from rollout filename:

```
~/.codex/sessions/rollout-2025-05-07T17-24-21-5973b6c0-94b8-487b-a530-2aeb6098ae0e.jsonl
                                               ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
                                               This is the session ID (UUID)
```

---

## Tasks

### Task 4.1: Update `init` Command with Tool Flags

**File**: `internal/commands/init.go` (modify existing)

**Description**: Add flags to generate tool-specific integrations.

**New Flags**:
| Flag | Description |
|------|-------------|
| `--opencode` | Generate OpenCode custom tool |
| `--claude-code` | Generate Claude Code hooks |
| `--codex` | Generate Codex helper script |
| `--all` | Generate all tool integrations |

**Implementation** (add to existing init command):

```go
package commands

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"
)

// Embed the template files
//go:embed templates/*
var templateFS embed.FS

func newInitCmd(app *pocketbase.PocketBase) *cobra.Command {
	var opencode, claudeCode, codex, all bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize EgenSkriven in current directory",
		Long: `Initialize EgenSkriven configuration and optionally generate
tool-specific integrations for session discovery.

Tool integrations enable AI agents to discover their session IDs,
which is required for the resume workflow.`,
		Example: `  # Initialize with default config
  egenskriven init
  
  # Initialize with OpenCode integration
  egenskriven init --opencode
  
  # Initialize with all tool integrations
  egenskriven init --all
  
  # Add integrations to existing project
  egenskriven init --claude-code --codex`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Existing init logic (if any)...
			
			// Handle tool integrations
			if all {
				opencode, claudeCode, codex = true, true, true
			}

			generated := []string{}

			if opencode {
				files, err := generateOpenCodeIntegration()
				if err != nil {
					return fmt.Errorf("failed to generate OpenCode integration: %w", err)
				}
				generated = append(generated, files...)
			}

			if claudeCode {
				files, err := generateClaudeCodeIntegration()
				if err != nil {
					return fmt.Errorf("failed to generate Claude Code integration: %w", err)
				}
				generated = append(generated, files...)
			}

			if codex {
				files, err := generateCodexIntegration()
				if err != nil {
					return fmt.Errorf("failed to generate Codex integration: %w", err)
				}
				generated = append(generated, files...)
			}

			// Output results
			if len(generated) > 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "Generated tool integrations:")
				for _, f := range generated {
					fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", f)
				}
				fmt.Fprintln(cmd.OutOrStdout())
				fmt.Fprintln(cmd.OutOrStdout(), "See each file for usage instructions.")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&opencode, "opencode", false, "Generate OpenCode custom tool")
	cmd.Flags().BoolVar(&claudeCode, "claude-code", false, "Generate Claude Code hooks")
	cmd.Flags().BoolVar(&codex, "codex", false, "Generate Codex helper script")
	cmd.Flags().BoolVar(&all, "all", false, "Generate all tool integrations")

	return cmd
}
```

---

### Task 4.2: Implement OpenCode Integration Generator

**File**: `internal/commands/init_opencode.go`

**Description**: Generate the OpenCode custom tool for session discovery.

**Generated file**: `.opencode/tool/egenskriven-session.ts`

**Implementation**:

```go
package commands

import (
	"fmt"
	"os"
	"path/filepath"
)

const openCodeToolTemplate = `import { tool } from "@opencode-ai/plugin"

/**
 * EgenSkriven Session Discovery Tool
 * 
 * This tool allows the AI agent to discover its OpenCode session ID,
 * which is required for linking sessions to EgenSkriven tasks.
 * 
 * Usage:
 * 1. Call this tool to get your session ID
 * 2. Run the provided command to link to a task
 * 
 * Example:
 *   Agent calls: egenskriven-session
 *   Response: { session_id: "abc-123", link_command: "egenskriven session link <task> --tool opencode --ref abc-123" }
 *   Agent runs: egenskriven session link WRK-42 --tool opencode --ref abc-123
 */
export default tool({
  description: ` + "`" + `Get current OpenCode session ID for EgenSkriven task tracking.

Call this tool when you need to:
- Link this session to an EgenSkriven task before starting work
- Get your session ID for any tracking purpose

Returns the session ID and a command to link it to a task.` + "`" + `,
  args: {},
  async execute(args, context) {
    const { sessionID, messageID, agent } = context
    
    return JSON.stringify({
      tool: "opencode",
      session_id: sessionID,
      message_id: messageID,
      agent: agent,
      link_command: ` + "`" + `egenskriven session link <task-ref> --tool opencode --ref ${sessionID}` + "`" + `,
      instructions: "Run the link_command with your task reference to link this session."
    }, null, 2)
  },
})
`

func generateOpenCodeIntegration() ([]string, error) {
	toolDir := ".opencode/tool"
	toolFile := filepath.Join(toolDir, "egenskriven-session.ts")

	// Create directory
	if err := os.MkdirAll(toolDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", toolDir, err)
	}

	// Check if file exists
	if _, err := os.Stat(toolFile); err == nil {
		// File exists - ask before overwriting? For now, skip
		return nil, fmt.Errorf("file already exists: %s (use --force to overwrite)", toolFile)
	}

	// Write tool file
	if err := os.WriteFile(toolFile, []byte(openCodeToolTemplate), 0644); err != nil {
		return nil, fmt.Errorf("failed to write %s: %w", toolFile, err)
	}

	return []string{toolFile}, nil
}
```

---

### Task 4.3: Implement Claude Code Integration Generator

**File**: `internal/commands/init_claude.go`

**Description**: Generate Claude Code hooks for session discovery.

**Generated files**:
- `.claude/hooks/egenskriven-session.sh` - Hook script
- `.claude/settings.json` - Updated settings (merged)

**Implementation**:

```go
package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

func generateClaudeCodeIntegration() ([]string, error) {
	hooksDir := ".claude/hooks"
	hookFile := filepath.Join(hooksDir, "egenskriven-session.sh")
	settingsFile := ".claude/settings.json"
	
	generatedFiles := []string{}

	// Create hooks directory
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", hooksDir, err)
	}

	// Write hook script
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

func mergeClaudeHooks(settings map[string]any) map[string]any {
	// Get or create hooks section
	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		hooks = map[string]any{}
	}

	// Add SessionStart hook
	sessionStartHooks := []map[string]any{
		{
			"matcher": "startup|resume",
			"hooks": []map[string]any{
				{
					"type":    "command",
					"command": `bash "$CLAUDE_PROJECT_DIR/.claude/hooks/egenskriven-session.sh"`,
				},
			},
		},
	}

	// Check if we already have SessionStart hooks
	existing, hasExisting := hooks["SessionStart"].([]any)
	if hasExisting {
		// Append our hook if not already present
		alreadyPresent := false
		for _, h := range existing {
			if hook, ok := h.(map[string]any); ok {
				if hooks, ok := hook["hooks"].([]any); ok {
					for _, hh := range hooks {
						if hhMap, ok := hh.(map[string]any); ok {
							if cmd, ok := hhMap["command"].(string); ok {
								if contains([]string{cmd}, "egenskriven-session.sh") {
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
			// Convert sessionStartHooks to []any
			for _, h := range sessionStartHooks {
				existing = append(existing, h)
			}
			hooks["SessionStart"] = existing
		}
	} else {
		hooks["SessionStart"] = sessionStartHooks
	}

	settings["hooks"] = hooks
	return settings
}
```

---

### Task 4.4: Implement Codex Integration Generator

**File**: `internal/commands/init_codex.go`

**Description**: Generate Codex helper script for session discovery.

**Generated file**: `.codex/get-session-id.sh`

**Implementation**:

```go
package commands

import (
	"fmt"
	"os"
	"path/filepath"
)

const codexHelperTemplate = `#!/bin/bash
# EgenSkriven Session ID Discovery for Codex CLI
#
# Codex CLI does not expose the session/thread ID via environment variable.
# This script extracts it from the most recent rollout file.
#
# Usage:
#   SESSION_ID=$(.codex/get-session-id.sh)
#   egenskriven session link <task-ref> --tool codex --ref $SESSION_ID
#
# Note: This works by finding the most recently modified rollout file.
# In rare cases with multiple concurrent Codex instances, it may return
# the wrong session. Use explicit session management if this is a concern.

set -e

# Codex data directory
CODEX_DIR="${CODEX_HOME:-$HOME/.codex}"
SESSIONS_DIR="$CODEX_DIR/sessions"

# Check if sessions directory exists
if [ ! -d "$SESSIONS_DIR" ]; then
    echo "ERROR: Codex sessions directory not found: $SESSIONS_DIR" >&2
    echo "Make sure Codex CLI has been run at least once." >&2
    exit 1
fi

# Find the most recently modified rollout file
LATEST_ROLLOUT=$(ls -t "$SESSIONS_DIR"/rollout-*.jsonl 2>/dev/null | head -1)

if [ -z "$LATEST_ROLLOUT" ]; then
    echo "ERROR: No Codex session files found in $SESSIONS_DIR" >&2
    echo "Start a Codex session first, then run this script." >&2
    exit 1
fi

# Extract UUID from filename
# Format: rollout-2025-05-07T17-24-21-5973b6c0-94b8-487b-a530-2aeb6098ae0e.jsonl
SESSION_ID=$(basename "$LATEST_ROLLOUT" | grep -oP '[0-9a-f]{8}(-[0-9a-f]{4}){3}-[0-9a-f]{12}')

if [ -z "$SESSION_ID" ]; then
    echo "ERROR: Could not extract session ID from filename: $LATEST_ROLLOUT" >&2
    exit 1
fi

# Output just the session ID
echo "$SESSION_ID"
`

func generateCodexIntegration() ([]string, error) {
	codexDir := ".codex"
	helperFile := filepath.Join(codexDir, "get-session-id.sh")

	// Create directory
	if err := os.MkdirAll(codexDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", codexDir, err)
	}

	// Write helper script
	if err := os.WriteFile(helperFile, []byte(codexHelperTemplate), 0755); err != nil {
		return nil, fmt.Errorf("failed to write %s: %w", helperFile, err)
	}

	return []string{helperFile}, nil
}
```

---

### Task 4.5: Add --force Flag for Overwriting

**Description**: Add a `--force` flag to overwrite existing files.

**Update init command**:

```go
var force bool

// In command definition
cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing files")

// Update generators to accept force parameter
func generateOpenCodeIntegration(force bool) ([]string, error) {
	// ...
	
	// Check if file exists
	if _, err := os.Stat(toolFile); err == nil && !force {
		return nil, fmt.Errorf("file already exists: %s (use --force to overwrite)", toolFile)
	}
	
	// ...
}
```

---

### Task 4.6: Write Unit Tests

**File**: `internal/commands/init_test.go`

```go
package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateOpenCodeIntegration(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Generate
	files, err := generateOpenCodeIntegration(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file created
	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
	}

	expectedFile := ".opencode/tool/egenskriven-session.ts"
	if files[0] != expectedFile {
		t.Errorf("expected %s, got %s", expectedFile, files[0])
	}

	// Verify file exists and has content
	content, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if len(content) == 0 {
		t.Error("file is empty")
	}

	// Verify key content
	contentStr := string(content)
	if !strings.Contains(contentStr, "context.sessionID") {
		t.Error("should reference context.sessionID")
	}
	if !strings.Contains(contentStr, "egenskriven session link") {
		t.Error("should include link command")
	}
}

func TestGenerateOpenCodeIntegrationNoOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Generate first time
	generateOpenCodeIntegration(false)

	// Try to generate again without force
	_, err := generateOpenCodeIntegration(false)
	if err == nil {
		t.Error("expected error when file exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention file exists: %v", err)
	}

	// Generate with force
	files, err := generateOpenCodeIntegration(true)
	if err != nil {
		t.Fatalf("should succeed with force: %v", err)
	}
	if len(files) != 1 {
		t.Error("should return generated file")
	}
}

func TestGenerateClaudeCodeIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	files, err := generateClaudeCodeIntegration(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should create hook script and settings
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}

	// Verify hook script
	hookScript := ".claude/hooks/egenskriven-session.sh"
	content, err := os.ReadFile(hookScript)
	if err != nil {
		t.Fatalf("failed to read hook script: %v", err)
	}
	if !strings.Contains(string(content), "CLAUDE_SESSION_ID") {
		t.Error("hook should reference CLAUDE_SESSION_ID")
	}

	// Verify script is executable
	info, _ := os.Stat(hookScript)
	if info.Mode()&0111 == 0 {
		t.Error("hook script should be executable")
	}

	// Verify settings.json
	settingsFile := ".claude/settings.json"
	settingsContent, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatalf("failed to read settings: %v", err)
	}
	if !strings.Contains(string(settingsContent), "SessionStart") {
		t.Error("settings should include SessionStart hook")
	}
}

func TestClaudeSettingsMerge(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create existing settings
	os.MkdirAll(".claude", 0755)
	existingSettings := `{
  "someExistingSetting": true,
  "hooks": {
    "SomeOtherHook": [{"matcher": "*", "hooks": []}]
  }
}`
	os.WriteFile(".claude/settings.json", []byte(existingSettings), 0644)

	// Generate integration
	generateClaudeCodeIntegration(true)

	// Verify settings were merged, not replaced
	content, _ := os.ReadFile(".claude/settings.json")
	contentStr := string(content)

	if !strings.Contains(contentStr, "someExistingSetting") {
		t.Error("should preserve existing settings")
	}
	if !strings.Contains(contentStr, "SomeOtherHook") {
		t.Error("should preserve existing hooks")
	}
	if !strings.Contains(contentStr, "SessionStart") {
		t.Error("should add SessionStart hook")
	}
}

func TestGenerateCodexIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	files, err := generateCodexIntegration(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
	}

	// Verify helper script
	helperScript := ".codex/get-session-id.sh"
	content, err := os.ReadFile(helperScript)
	if err != nil {
		t.Fatalf("failed to read helper script: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "CODEX_HOME") {
		t.Error("should reference CODEX_HOME")
	}
	if !strings.Contains(contentStr, "rollout-") {
		t.Error("should look for rollout files")
	}

	// Verify executable
	info, _ := os.Stat(helperScript)
	if info.Mode()&0111 == 0 {
		t.Error("helper script should be executable")
	}
}

func TestInitCommandAllFlag(t *testing.T) {
	app := setupTestApp(t)
	defer cleanupTestApp(t, app)

	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	cmd := newInitCmd(app)
	cmd.SetArgs([]string{"--all"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify all integrations were created
	expectedFiles := []string{
		".opencode/tool/egenskriven-session.ts",
		".claude/hooks/egenskriven-session.sh",
		".claude/settings.json",
		".codex/get-session-id.sh",
	}

	for _, f := range expectedFiles {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			t.Errorf("expected file not created: %s", f)
		}
	}
}
```

---

### Task 4.7: Update Skills Documentation

**File**: `internal/commands/skills/egenskriven/SKILL.md` (update)

Add section about tool integrations:

```markdown
## Tool Integrations

Before using the collaborative workflow, initialize the tool integration for your AI tool:

### OpenCode

\`\`\`bash
egenskriven init --opencode
\`\`\`

This creates `.opencode/tool/egenskriven-session.ts`. When working, call the
`egenskriven-session` tool to get your session ID:

\`\`\`
Agent: [calls egenskriven-session tool]
Response: { session_id: "abc-123", link_command: "egenskriven session link <task> --tool opencode --ref abc-123" }
Agent: [runs link command]
\`\`\`

### Claude Code

\`\`\`bash
egenskriven init --claude-code
\`\`\`

This creates a hook that automatically sets `$CLAUDE_SESSION_ID` on session start.
Link your session with:

\`\`\`bash
egenskriven session link <task-ref> --tool claude-code --ref $CLAUDE_SESSION_ID
\`\`\`

### Codex CLI

\`\`\`bash
egenskriven init --codex
\`\`\`

This creates a helper script. Get your session ID with:

\`\`\`bash
SESSION_ID=$(.codex/get-session-id.sh)
egenskriven session link <task-ref> --tool codex --ref $SESSION_ID
\`\`\`
```

---

## Files Changed/Created

| File | Change Type | Description |
|------|-------------|-------------|
| `internal/commands/init.go` | Modified | Add tool integration flags |
| `internal/commands/init_opencode.go` | New | OpenCode generator |
| `internal/commands/init_claude.go` | New | Claude Code generator |
| `internal/commands/init_codex.go` | New | Codex generator |
| `internal/commands/init_test.go` | New/Modified | Generator tests |
| Skills documentation | Modified | Add tool integration docs |

### Generated Files (in user's project)

| File | Tool | Description |
|------|------|-------------|
| `.opencode/tool/egenskriven-session.ts` | OpenCode | Custom tool |
| `.claude/hooks/egenskriven-session.sh` | Claude Code | Hook script |
| `.claude/settings.json` | Claude Code | Hook configuration |
| `.codex/get-session-id.sh` | Codex | Helper script |

---

## Phase 4 Task Checklist

This section provides a detailed checklist of all tasks required to complete Phase 4.

### Task 4.1: Update `init` Command with Tool Flags

- [x] Add `--opencode` flag to init command
- [x] Add `--claude-code` flag to init command
- [x] Add `--codex` flag to init command
- [x] Add `--all` flag to init command
- [x] Implement flag handling logic (when `--all` is set, enable all three)
- [x] Add command examples to help text
- [x] Display list of generated files on completion

### Task 4.2: Implement OpenCode Integration Generator

- [x] Create `internal/commands/init_opencode.go` file
- [x] Define `openCodeToolTemplate` constant with TypeScript template
- [x] Implement `generateOpenCodeIntegration(force bool)` function
- [x] Create `.opencode/tool/` directory if not exists
- [x] Write `egenskriven-session.ts` file with correct permissions (0644)
- [x] Handle existing file detection (error without `--force`)
- [x] Verify template includes `context.sessionID` reference
- [x] Verify template includes `link_command` instructions

### Task 4.3: Implement Claude Code Integration Generator

- [x] Create `internal/commands/init_claude.go` file
- [x] Define `claudeCodeHookTemplate` constant with bash script template
- [x] Implement `generateClaudeCodeIntegration(force bool)` function
- [x] Implement `loadClaudeSettings(path string)` helper function
- [x] Implement `mergeClaudeHooks(settings map[string]any)` function
- [x] Create `.claude/hooks/` directory if not exists
- [x] Write `egenskriven-session.sh` hook script with executable permissions (0755)
- [x] Create or update `.claude/settings.json` with hook configuration
- [x] Ensure settings merge preserves existing configuration
- [x] Hook script should support both `jq` and Python fallback for JSON parsing
- [x] Hook script should persist `CLAUDE_SESSION_ID` to `$CLAUDE_ENV_FILE`

### Task 4.4: Implement Codex Integration Generator

- [x] Create `internal/commands/init_codex.go` file (implemented in init_integrations.go)
- [x] Define `codexHelperTemplate` constant with bash script template
- [x] Implement `generateCodexIntegration(force bool)` function
- [x] Create `.codex/` directory if not exists
- [x] Write `get-session-id.sh` helper script with executable permissions (0755)
- [x] Script should handle `CODEX_HOME` environment variable
- [x] Script should find most recent rollout file
- [x] Script should extract UUID from rollout filename using regex
- [x] Script should provide clear error messages when no session found

### Task 4.5: Add --force Flag for Overwriting

- [x] Add `--force` / `-f` flag to init command (completed in Task 4.1)
- [x] Pass `force` parameter to all generator functions (completed in Task 4.1)
- [x] Update `generateOpenCodeIntegration` to accept `force` parameter (completed in Task 4.2)
- [x] Update `generateClaudeCodeIntegration` to accept `force` parameter (completed in Task 4.3)
- [x] Update `generateCodexIntegration` to accept `force` parameter (completed in Task 4.4)
- [x] When file exists and `force=false`, return appropriate error message (verified for OpenCode, Claude Code, Codex)
- [x] When file exists and `force=true`, overwrite without error (verified for OpenCode, Claude Code, Codex)

### Task 4.6: Write Unit Tests

- [x] Create `internal/commands/init_test.go` file
- [x] Add `strings` import for content verification
- [x] Implement `TestGenerateOpenCodeIntegration` test
  - [x] Test file creation in temp directory
  - [x] Verify correct file path
  - [x] Verify file is not empty
  - [x] Verify file contains `sessionID` (via context destructuring)
  - [x] Verify file contains `egenskriven session link`
- [x] Implement `TestGenerateOpenCodeIntegrationNoOverwrite` test
  - [x] Test first generation succeeds
  - [x] Test second generation without force fails
  - [x] Verify error message mentions "already exists"
  - [x] Test generation with force succeeds
- [x] Implement `TestGenerateClaudeCodeIntegration` test
  - [x] Verify 2 files created (hook script + settings)
  - [x] Verify hook script content contains `CLAUDE_SESSION_ID`
  - [x] Verify hook script is executable
  - [x] Verify settings.json contains `SessionStart` hook
- [x] Implement `TestClaudeSettingsMerge` test
  - [x] Create existing settings.json with other settings
  - [x] Generate integration
  - [x] Verify existing settings preserved
  - [x] Verify new SessionStart hook added
- [x] Implement `TestGenerateCodexIntegration` test
  - [x] Verify 1 file created
  - [x] Verify script contains `CODEX_HOME`
  - [x] Verify script looks for `rollout-` files
  - [x] Verify script is executable
- [x] Implement `TestInitCommandAllFlag` test (named `TestInitIntegrationsAllFlag`)
  - [x] Test `--all` flag creates all integrations
  - [x] Verify all expected files exist
- [x] Additional tests implemented:
  - [x] `TestClaudeSettingsMergeIdempotent` - ensures hooks aren't duplicated
  - [x] `TestLoadClaudeSettings_NonExistent` - handles missing files
  - [x] `TestLoadClaudeSettings_InvalidJSON` - handles invalid JSON
  - [x] `TestLoadClaudeSettings_Valid` - loads valid settings
  - [x] `TestMergeClaudeHooks_EmptySettings` - merges into empty settings
  - [x] `TestMergeClaudeHooks_ExistingDifferentHooks` - preserves other hooks
  - [x] `TestOpenCodeToolTemplateContent` - validates template structure
  - [x] `TestClaudeCodeHookTemplateContent` - validates template structure
  - [x] `TestCodexHelperTemplateContent` - validates template structure
  - [x] `TestFilePermissions` - validates script permissions
  - [x] `Test*DirectoryCreation` - validates directory structure

### Task 4.7: Update Skills Documentation

- [ ] Update `internal/commands/skills/egenskriven/SKILL.md`
- [ ] Add "Tool Integrations" section
- [ ] Document OpenCode integration setup and usage
- [ ] Document Claude Code integration setup and usage
- [ ] Document Codex CLI integration setup and usage
- [ ] Include example commands for each tool

### Testing Checklist

#### OpenCode Integration Testing

- [x] `init --opencode` creates `.opencode/tool/egenskriven-session.ts`
- [x] Generated tool file is valid TypeScript (no syntax errors)
- [x] Tool correctly references `context.sessionID`
- [x] Tool provides helpful usage instructions
- [x] Re-running without `--force` fails with appropriate error
- [x] Re-running with `--force` overwrites successfully

#### Claude Code Integration Testing

- [x] `init --claude-code` creates hook script at `.claude/hooks/egenskriven-session.sh`
- [x] `init --claude-code` creates/updates `.claude/settings.json`
- [x] Hook script is executable (has execute permissions)
- [x] Hook script handles JSON parsing with jq
- [ ] Hook script handles JSON parsing with Python fallback (not tested - jq available)
- [x] Settings merge preserves existing configuration
- [x] Settings include correct SessionStart hook configuration

#### Codex Integration Testing

- [x] `init --codex` creates `.codex/get-session-id.sh`
- [x] Helper script is executable (has execute permissions)
- [x] Helper script correctly parses rollout filenames
- [x] Helper script handles missing sessions directory gracefully
- [x] Helper script handles no rollout files gracefully

#### Combined Testing

- [x] `init --all` creates all three integrations
- [x] `init --force` overwrites all existing files
- [x] Error messages are helpful and actionable
- [x] Output lists all generated files clearly

#### Manual Integration Testing

- [ ] OpenCode: Start session, call tool, verify session ID returned correctly
- [ ] Claude Code: Start session, verify `$CLAUDE_SESSION_ID` is set in env
- [ ] Codex: Start session, run helper script, verify correct session ID returned

---

## Notes for Implementer

1. **File permissions**: Hook scripts must be executable (0755).

2. **JSON merging**: When updating settings.json, carefully merge to preserve existing configuration.

3. **Template escaping**: Be careful with backticks and template literals in the embedded templates.

4. **Cross-platform**: The helper scripts use bash. Consider Windows compatibility if needed.

5. **Idempotency**: Running the generators multiple times should either fail gracefully or use --force.

6. **Testing with real tools**: Manual testing with actual OpenCode/Claude Code/Codex is essential to verify the integrations work.
