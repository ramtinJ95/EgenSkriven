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

// generateCodexIntegration generates Codex helper script for session discovery.
// Generated file: .codex/get-session-id.sh
func generateCodexIntegration(force bool) ([]string, error) {
	codexDir := ".codex"
	helperFile := filepath.Join(codexDir, "get-session-id.sh")

	// Create directory
	if err := os.MkdirAll(codexDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", codexDir, err)
	}

	// Check if file exists
	if _, err := os.Stat(helperFile); err == nil && !force {
		return nil, fmt.Errorf("file already exists: %s (use --force to overwrite)", helperFile)
	}

	// Write helper script with executable permissions (0755)
	if err := os.WriteFile(helperFile, []byte(codexHelperTemplate), 0755); err != nil {
		return nil, fmt.Errorf("failed to write %s: %w", helperFile, err)
	}

	return []string{helperFile}, nil
}
