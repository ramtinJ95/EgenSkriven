---
name: egenskriven
description: Local-first kanban task manager for AI agent workflows. Use when managing tasks, tracking work, creating issues, or when user mentions tasks, kanban, boards, backlog, or egenskriven.
---

## Overview

EgenSkriven is a local-first kanban task manager with CLI and web UI, designed for AI agents from the ground up. It replaces ephemeral todo systems (like TodoWrite) with persistent, structured task tracking.

## Quick Start Commands

```bash
# Get project status summary
egenskriven context --json

# Get recommended next task
egenskriven suggest --json

# List actionable (unblocked) tasks
egenskriven list --ready --json

# Create a new task
egenskriven add "Task title" --type feature --priority medium
```

## Essential CRUD Operations

| Command | Description |
|---------|-------------|
| `add "title"` | Create task (flags: `--type`, `--priority`, `--column`) |
| `list` | List tasks (flags: `--ready`, `--column`, `--type`) |
| `show <ref>` | Display task details |
| `move <ref> <column>` | Move task between columns |
| `update <ref>` | Modify task properties |
| `delete <ref>` | Remove task (`--force` skips confirmation) |

## Task Reference Formats

Tasks can be referenced by:

- **Full ID**: `abc123def456`
- **ID prefix**: `abc` (minimum unique prefix)
- **Display ID**: `WRK-123` (board prefix + sequence)
- **Title substring**: `"dark mode"` (case-insensitive match)

## Columns (Workflow States)

| Column | Description |
|--------|-------------|
| `backlog` | Ideas and future work |
| `todo` | Ready to start |
| `in_progress` | Currently being worked on |
| `review` | Awaiting review |
| `done` | Completed |

## JSON Output

All commands support `--json` for machine-readable output:

```bash
# Get specific fields only (reduces tokens)
egenskriven list --json --fields id,title,column

# Full task details
egenskriven show <ref> --json
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |
| 3 | Task not found |
| 4 | Ambiguous reference |
| 5 | Validation error |

## When to Use EgenSkriven

- Multi-step work spanning multiple exchanges
- Work that might be interrupted
- Features with dependencies
- Bugs that need tracking

**Not needed for**: Simple questions, one-off commands, trivial changes.

## Tool Integrations

Before using the collaborative workflow (blocking tasks, resume flow), initialize the tool integration for your AI coding tool:

### OpenCode

```bash
egenskriven init --opencode
```

This creates `.opencode/tool/egenskriven-session.ts`. When working on a task, call the `egenskriven-session` tool to get your session ID:

```
Agent: [calls egenskriven-session tool]
Response: { session_id: "abc-123", link_command: "egenskriven session link <task> --tool opencode --ref abc-123" }
Agent: [runs link command to associate session with task]
```

### Claude Code

```bash
egenskriven init --claude-code
```

This creates:
- `.claude/hooks/egenskriven-session.sh` - Hook script that runs on SessionStart
- `.claude/settings.json` - Hook configuration (merged with existing settings)

After the hook runs, your session ID is available as `$CLAUDE_SESSION_ID`. Link your session with:

```bash
egenskriven session link <task-ref> --tool claude-code --ref $CLAUDE_SESSION_ID
```

### Codex CLI

```bash
egenskriven init --codex
```

This creates `.codex/get-session-id.sh`. Get your session ID with:

```bash
SESSION_ID=$(.codex/get-session-id.sh)
egenskriven session link <task-ref> --tool codex --ref $SESSION_ID
```

### All Integrations

Generate all tool integrations at once:

```bash
egenskriven init --all
egenskriven init --all --force  # Overwrite existing files
```

## Related Skills

- `egenskriven-workflows` - Workflow modes (strict/light/minimal) and agent behaviors
- `egenskriven-advanced` - Epics, dependencies, sub-tasks, batch operations
