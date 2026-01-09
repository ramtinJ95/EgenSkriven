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

// generateOpenCodeIntegration generates the OpenCode custom tool for session discovery.
// Generated file: .opencode/tool/egenskriven-session.ts
func generateOpenCodeIntegration(force bool) ([]string, error) {
	toolDir := ".opencode/tool"
	toolFile := filepath.Join(toolDir, "egenskriven-session.ts")

	// Create directory
	if err := os.MkdirAll(toolDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", toolDir, err)
	}

	// Check if file exists
	if _, err := os.Stat(toolFile); err == nil && !force {
		return nil, fmt.Errorf("file already exists: %s (use --force to overwrite)", toolFile)
	}

	// Write tool file
	if err := os.WriteFile(toolFile, []byte(openCodeToolTemplate), 0644); err != nil {
		return nil, fmt.Errorf("failed to write %s: %w", toolFile, err)
	}

	return []string{toolFile}, nil
}
