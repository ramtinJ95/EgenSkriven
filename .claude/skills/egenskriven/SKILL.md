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

## Related Skills

- `egenskriven-workflows` - Workflow modes (strict/light/minimal) and agent behaviors
- `egenskriven-advanced` - Epics, dependencies, sub-tasks, batch operations
