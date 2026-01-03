# EgenSkriven: PocketBase + React + CLI Architecture

## Overview

A local-first kanban board with CLI-first design. The CLI is designed to be agent-friendly - any AI coding assistant can invoke it directly without needing a special "AI layer". PocketBase provides the infrastructure: SQLite database, REST API, real-time subscriptions via SSE, and admin UI. The React frontend is embedded in the binary for single-file distribution.

```
┌─────────────────────────────────────────────────────────────────┐
│                     Single Go Binary                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────────────┐│
│  │   CLI        │  │  PocketBase  │  │   Embedded React UI   ││
│  │   Commands   │  │    Core      │  │   (go:embed)          ││
│  │              │  │              │  │                        ││
│  │  • add       │  │  • REST API  │  │  • Kanban board       ││
│  │  • move      │  │  • Realtime  │  │  • Real-time sync     ││
│  │  • list      │  │  • Auth      │  │  • Drag & drop        ││
│  │  • update    │  │  • Admin UI  │  │                        ││
│  │  • delete    │  │              │  │                        ││
│  └──────┬───────┘  └──────┬───────┘  └───────────┬────────────┘│
│         │                 │                       │             │
│         └─────────────────┼───────────────────────┘             │
│                           │                                     │
│                    ┌──────▼──────┐                              │
│                    │   SQLite    │                              │
│                    │  (pb_data)  │                              │
│                    └─────────────┘                              │
└─────────────────────────────────────────────────────────────────┘
```

## Design Principles

### Agent-Friendly CLI

The CLI is designed so that AI agents (Claude, GPT, Cursor, etc.) can use it directly without any special integration layer:

1. **Structured output**: All commands support `--json` for machine-readable output
2. **Flexible identification**: Tasks can be referenced by ID or partial title match
3. **Batch operations**: Commands accept multiple items via stdin or file
4. **Rich filtering**: Precise queries reduce context needed by agents
5. **Idempotent operations**: `--id` flag allows safe retries
6. **Clear errors**: Error messages include enough context for agents to self-correct

### Human-First Defaults

Despite being agent-friendly, the CLI defaults to human-readable output. Agents can opt into JSON mode.

## Project Structure

```
egenskriven/
├── cmd/
│   └── egenskriven/
│       └── main.go              # Entry point, registers CLI commands
├── internal/
│   ├── commands/
│   │   ├── root.go              # Root command, global flags
│   │   ├── add.go               # egenskriven add "task title"
│   │   ├── move.go              # egenskriven move <task> <column>
│   │   ├── update.go            # egenskriven update <task> [fields]
│   │   ├── delete.go            # egenskriven delete <task>
│   │   ├── list.go              # egenskriven list [filters]
│   │   ├── show.go              # egenskriven show <task>
│   │   ├── epic.go              # egenskriven epic [subcommands]
│   │   ├── board.go             # egenskriven board [subcommands]
│   │   ├── prime.go             # egenskriven prime (agent instructions)
│   │   ├── prime.tmpl           # Agent instructions template
│   │   ├── context.go           # egenskriven context (project summary)
│   │   ├── suggest.go           # egenskriven suggest (task suggestions)
│   │   ├── init.go              # egenskriven init (config setup)
│   │   ├── template.go          # egenskriven template [subcommands]
│   │   ├── export.go            # egenskriven export
│   │   ├── import.go            # egenskriven import
│   │   ├── version.go           # egenskriven version
│   │   ├── completion.go        # egenskriven completion
│   │   ├── position.go          # Position calculation helpers
│   │   ├── *_test.go            # Command tests
│   │   └── position_test.go     # Position calculation tests
│   ├── config/
│   │   ├── config.go            # Project config loading
│   │   └── config_test.go       # Config tests
│   ├── output/
│   │   ├── output.go            # JSON/human output formatting
│   │   └── output_test.go       # Output tests
│   ├── resolver/
│   │   ├── resolver.go          # ID/title task resolution
│   │   └── resolver_test.go     # Resolver tests
│   ├── testutil/
│   │   └── testutil.go          # Shared test helpers
│   └── hooks/
│       └── hooks.go             # PocketBase event hooks
├── ui/
│   ├── src/
│   │   ├── App.tsx
│   │   ├── components/
│   │   │   ├── Board.tsx        # Main kanban board
│   │   │   ├── Column.tsx       # Single column
│   │   │   ├── TaskCard.tsx     # Draggable task card
│   │   │   ├── TaskDetail.tsx   # Task detail panel
│   │   │   ├── QuickCreate.tsx  # Quick create modal
│   │   │   ├── CommandPalette.tsx # Command palette (Cmd+K)
│   │   │   ├── Sidebar.tsx      # Board/view sidebar
│   │   │   ├── FilterBar.tsx    # Filter controls
│   │   │   ├── FilterBuilder.tsx # Filter construction UI
│   │   │   ├── ListView.tsx     # List view alternative
│   │   │   ├── Search.tsx       # Search input component
│   │   │   ├── Settings.tsx     # Settings panel
│   │   │   ├── BoardSettings.tsx # Board configuration
│   │   │   ├── ShortcutsHelp.tsx # Keyboard shortcuts modal
│   │   │   ├── Toast.tsx        # Toast notifications
│   │   │   ├── DisplayOptions.tsx # View display settings
│   │   │   ├── MarkdownEditor.tsx # Description editor
│   │   │   ├── StatusPicker.tsx  # Status selector popover
│   │   │   ├── PriorityPicker.tsx # Priority selector popover
│   │   │   ├── TypePicker.tsx    # Type selector popover
│   │   │   ├── LabelPicker.tsx   # Label selector popover
│   │   │   ├── EpicList.tsx     # Epic listing
│   │   │   ├── EpicDetail.tsx   # Epic detail view
│   │   │   └── EpicPicker.tsx   # Epic selector
│   │   ├── hooks/
│   │   │   ├── usePocketBase.ts # PocketBase SDK wrapper
│   │   │   └── useKeyboard.ts   # Keyboard shortcut manager
│   │   ├── stores/
│   │   │   ├── selection.ts     # Task selection state
│   │   │   └── filters.ts       # Filter state
│   │   ├── styles/
│   │   │   ├── tokens.css       # Design tokens
│   │   │   └── light.css        # Light mode overrides
│   │   ├── lib/
│   │   │   └── pb.ts            # PocketBase client instance
│   │   └── test/
│   │       └── setup.ts         # Test setup
│   ├── dist/                    # Vite build output
│   ├── embed.go                 # go:embed directive
│   ├── package.json
│   ├── vite.config.ts
│   └── vitest.config.ts         # Vitest configuration
├── migrations/                  # PocketBase migrations
│   └── 1_initial.go
├── .egenskriven/                # Project-specific config (created by init)
│   └── config.json              # Agent workflow configuration
├── .opencode/                   # OpenCode agent integration
│   └── plugin/
│       └── egenskriven-prime.ts # OpenCode plugin for prime injection
├── .claude/                     # Claude Code agent integration
│   └── settings.json            # Claude hooks configuration
├── pb_data/                     # SQLite + uploads (gitignored)
├── go.mod
├── go.sum
├── Makefile
├── .air.toml                    # Hot reload configuration
├── .gitignore
└── README.md
```

## Data Model (PocketBase Collections)

### tasks
| Field            | Type     | Description                              |
|------------------|----------|------------------------------------------|
| id               | string   | Auto-generated (or user-provided for idempotency) |
| title            | string   | Task title                               |
| description      | string   | Optional longer description              |
| type             | select   | bug, feature, chore                      |
| priority         | select   | low, medium, high, urgent                |
| column           | select   | backlog, todo, in_progress, review, done |
| position         | number   | Order within column (fractional allowed) |
| board            | relation | Link to boards collection (multi-board)  |
| epic             | relation | Optional link to epics collection        |
| parent           | relation | Optional parent task (for sub-tasks)     |
| labels           | json     | Array of label strings                   |
| blocked_by       | json     | Array of task IDs that block this task   |
| due_date         | date     | Optional due date                        |
| created_by       | select   | user, agent, cli - who created this task |
| created_by_agent | string   | Agent identifier (e.g., "claude", "opencode") |
| history          | json     | Array of activity entries (see below)    |
| created          | date     | Auto-generated                           |
| updated          | date     | Auto-generated                           |

### history (JSON array within task)

Each entry in the `history` array tracks a change:

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "action": "created",
  "actor": "agent",
  "actor_detail": "claude",
  "changes": null
}
```

| Field        | Description                                    |
|--------------|------------------------------------------------|
| timestamp    | ISO 8601 timestamp                             |
| action       | created, updated, moved, completed, deleted    |
| actor        | user, agent, cli                               |
| actor_detail | Optional identifier (agent name, username)     |
| changes      | Object with `field`, `from`, `to` for updates  |

### epics
| Field       | Type     | Description                              |
|-------------|----------|------------------------------------------|
| id          | string   | Auto-generated                           |
| title       | string   | Epic title                               |
| description | string   | Epic description                         |
| color       | string   | Hex color for visual grouping            |

### boards
| Field       | Type     | Description                              |
|-------------|----------|------------------------------------------|
| id          | string   | Auto-generated                           |
| name        | string   | Board name (e.g., "Work", "Personal")    |
| prefix      | string   | Task ID prefix (e.g., "WRK"), unique     |
| columns     | json     | Array of column definitions              |
| color       | string   | Accent color for board                   |
| created     | date     | Auto-generated                           |
| updated     | date     | Auto-generated                           |

### views
| Field       | Type     | Description                              |
|-------------|----------|------------------------------------------|
| id          | string   | Auto-generated                           |
| name        | string   | View name                                |
| board       | relation | Link to boards collection                |
| filters     | json     | Saved filter state                       |
| display     | json     | View preferences (board/list, fields)    |
| is_favorite | boolean  | Whether view is starred                  |
| created     | date     | Auto-generated                           |
| updated     | date     | Auto-generated                           |

### templates
| Field       | Type     | Description                              |
|-------------|----------|------------------------------------------|
| id          | string   | Auto-generated or user-provided name     |
| name        | string   | Template name (e.g., "bug-report")       |
| type        | select   | Default task type                        |
| priority    | select   | Default priority                         |
| column      | select   | Default column                           |
| labels      | json     | Default labels array                     |
| epic        | relation | Default epic link                        |
| description | string   | Default description template             |
| created     | date     | Auto-generated                           |

**Note:** When multi-board support is enabled, tasks have a `board` relation field linking them to a board. Task IDs are displayed with board prefix (e.g., "WRK-123") but stored as auto-generated IDs internally.

## CLI Interface

### Global Flags

All commands support these flags:

| Flag       | Short | Description                          |
|------------|-------|--------------------------------------|
| `--json`   | `-j`  | Output in JSON format                |
| `--quiet`  | `-q`  | Suppress non-essential output        |
| `--data`   |       | Path to pb_data directory            |
| `--board`  | `-b`  | Specify board by name or prefix (multi-board mode) |

### Task Resolution

Commands that accept `<task>` resolve it in this order:
1. **Exact ID match**: If input matches a task ID exactly
2. **ID prefix match**: If input is a unique prefix of a task ID
3. **Title match**: Case-insensitive substring match (fails if ambiguous)

```bash
# All of these could reference the same task
egenskriven show abc123def456
egenskriven show abc123
egenskriven show "fix login"
```

If ambiguous, the command fails with a list of matching tasks.

### Commands

#### `egenskriven add`

Create one or more tasks.

```bash
# Basic usage
egenskriven add "Implement dark mode"

# With options
egenskriven add "Fix login crash" --type bug --priority urgent --column todo

# With custom ID (idempotent)
egenskriven add "Setup CI pipeline" --id ci-setup-001

# With labels and epic
egenskriven add "Add user avatars" --label ui --label frontend --epic abc123

# Agent creating a task (identifies itself for activity tracking)
egenskriven add "Refactor auth module" --type chore --agent claude

# Batch: from stdin (one JSON per line)
echo '{"title":"Task 1"}
{"title":"Task 2","priority":"high"}' | egenskriven add --stdin

# Batch: from file
egenskriven add --file tasks.json
```

**Flags:**

| Flag         | Short | Default   | Description                        |
|--------------|-------|-----------|------------------------------------|
| `--type`     | `-t`  | feature   | bug, feature, chore                |
| `--priority` | `-p`  | medium    | low, medium, high, urgent          |
| `--column`   | `-c`  | backlog   | Initial column                     |
| `--label`    | `-l`  |           | Add label (repeatable)             |
| `--epic`     | `-e`  |           | Link to epic                       |
| `--id`       |       |           | Custom ID (for idempotency)        |
| `--stdin`    |       |           | Read tasks from stdin (JSON lines) |
| `--file`     | `-f`  |           | Read tasks from JSON file          |
| `--created-by` |     |           | Creator type: user, agent, cli     |
| `--agent`    |       |           | Agent identifier (e.g., "claude", "opencode") |
| `--template` |       |           | Use a task template                |
| `--due`      |       |           | Due date (ISO 8601 or natural language) |
| `--parent`   |       |           | Parent task ID (creates sub-task)  |
| `--blocked-by` |     |           | Blocking task ID (repeatable)      |

**Output (human):**
```
✓ Created task: Implement dark mode [abc123]
```

**Output (JSON):**
```json
{
  "id": "abc123def456",
  "title": "Implement dark mode",
  "type": "feature",
  "priority": "medium",
  "column": "backlog",
  "position": 1000,
  "labels": [],
  "created": "2024-01-15T10:30:00Z"
}
```

#### `egenskriven list`

List and filter tasks.

```bash
# All tasks
egenskriven list

# Filter by column
egenskriven list --column todo
egenskriven list -c in_progress -c review  # multiple columns

# Filter by type, priority
egenskriven list --type bug --priority urgent

# Filter by label
egenskriven list --label frontend

# Filter by epic
egenskriven list --epic abc123

# Combine filters (AND logic)
egenskriven list --column todo --type bug --priority high

# Search by title
egenskriven list --search "login"

# Output as JSON (for agents)
egenskriven list --json

# List all boards' tasks
egenskriven list --all-boards
```

**Flags:**

| Flag           | Short | Description                                      |
|----------------|-------|--------------------------------------------------|
| `--column`     | `-c`  | Filter by column (repeatable)                    |
| `--type`       | `-t`  | Filter by type (repeatable)                      |
| `--priority`   | `-p`  | Filter by priority (repeatable)                  |
| `--label`      | `-l`  | Filter by label (repeatable)                     |
| `--epic`       | `-e`  | Filter by epic                                   |
| `--search`     | `-s`  | Search title (case-insensitive)                  |
| `--limit`      |       | Max results (default: no limit)                  |
| `--sort`       |       | Sort field (default: position)                   |
| `--ready`      |       | Unblocked tasks in todo/backlog (agent-friendly) |
| `--is-blocked` |       | Only show tasks blocked by others                |
| `--not-blocked`|       | Only show tasks not blocked by others            |
| `--fields`     |       | Comma-separated fields to include in JSON output |
| `--created-by` |       | Filter by creator: user, agent, cli              |
| `--agent`      |       | Filter by specific agent name                    |
| `--all-boards` |       | Show tasks from all boards (multi-board mode)    |
| `--due-before` |       | Tasks due before date                            |
| `--due-after`  |       | Tasks due after date                             |
| `--has-parent` |       | Show only sub-tasks                              |
| `--no-parent`  |       | Show only top-level tasks (exclude sub-tasks)    |

**Output (human):**
```
BACKLOG
  [abc123] Implement dark mode (feature, medium)
  [def456] Add user avatars (feature, low)

TODO
  [ghi789] Fix login crash (bug, urgent)

IN_PROGRESS
  (empty)
```

**Output (JSON):**
```json
{
  "tasks": [
    {
      "id": "abc123",
      "title": "Implement dark mode",
      "type": "feature",
      "priority": "medium",
      "column": "backlog",
      "position": 1000,
      "labels": [],
      "created": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 1
}
```

#### `egenskriven show`

Show detailed information about a task.

```bash
egenskriven show abc123
egenskriven show "login crash"  # title match
```

**Output (human):**
```
Task: abc123def456
Title:       Fix login crash
Type:        bug
Priority:    urgent
Column:      todo
Position:    1000
Labels:      auth, critical
Epic:        -
Created:     2024-01-15 10:30:00
Updated:     2024-01-15 14:22:00

Description:
  Users are experiencing crashes when attempting to log in
  with SSO credentials. Stack trace attached.
```

#### `egenskriven move`

Move task to a different column and/or position.

```bash
# Move to column (appends to end)
egenskriven move abc123 in_progress

# Move to column at specific position
egenskriven move abc123 in_progress --position 0  # top
egenskriven move abc123 in_progress --position -1 # bottom (default)

# Move to position within current column
egenskriven move abc123 --position 0

# Move relative to another task
egenskriven move abc123 --after def456
egenskriven move abc123 --before def456
```

**Flags:**

| Flag         | Description                              |
|--------------|------------------------------------------|
| `--position` | Numeric position (0=top, -1=bottom)      |
| `--after`    | Position after this task                 |
| `--before`   | Position before this task                |

#### `egenskriven update`

Update task fields.

```bash
# Update single field
egenskriven update abc123 --title "New title"
egenskriven update abc123 --priority urgent

# Update multiple fields
egenskriven update abc123 --type bug --priority high --label critical

# Clear optional fields
egenskriven update abc123 --description ""
egenskriven update abc123 --epic ""

# Add/remove labels
egenskriven update abc123 --add-label urgent --remove-label backlog

# Blocking relationships
egenskriven update abc123 --blocked-by def456
egenskriven update abc123 --remove-blocked-by def456
```

**Flags:**

| Flag                | Description                            |
|---------------------|----------------------------------------|
| `--title`           | New title                              |
| `--description`     | New description                        |
| `--type`            | New type                               |
| `--priority`        | New priority                           |
| `--epic`            | Link to epic (empty to clear)          |
| `--add-label`       | Add label (repeatable)                 |
| `--remove-label`    | Remove label (repeatable)              |
| `--blocked-by`      | Add blocking task ID (repeatable)      |
| `--remove-blocked-by` | Remove blocking task ID (repeatable) |
| `--due`             | Set due date (ISO 8601, or "" to clear)|
| `--parent`          | Set parent task (for sub-tasks)        |

#### `egenskriven delete`

Delete one or more tasks.

```bash
# Single task
egenskriven delete abc123

# Multiple tasks
egenskriven delete abc123 def456 ghi789

# From stdin
echo -e "abc123\ndef456" | egenskriven delete --stdin

# Skip confirmation (for scripts/agents)
egenskriven delete abc123 --force
```

**Flags:**

| Flag      | Short | Description                    |
|-----------|-------|--------------------------------|
| `--force` | `-f`  | Skip confirmation prompt       |
| `--stdin` |       | Read task IDs from stdin       |

#### `egenskriven epic`

Manage epics.

```bash
# List epics
egenskriven epic list

# Create epic
egenskriven epic add "Q1 Launch" --color "#3B82F6"

# Show epic with linked tasks
egenskriven epic show abc123

# Delete epic (tasks remain, unlinked)
egenskriven epic delete abc123
```

#### `egenskriven board`

Manage multiple boards.

```bash
# List all boards
egenskriven board list

# Create a new board
egenskriven board add "Work" --prefix WRK --color "#3B82F6"

# Show board details
egenskriven board show work

# Delete a board (tasks are also deleted!)
egenskriven board delete work --force

# Set default board for CLI
egenskriven board use work
```

**Flags for `board add`:**

| Flag       | Short | Default   | Description                        |
|------------|-------|-----------|------------------------------------|
| `--prefix` | `-p`  | Required  | Task ID prefix (e.g., "WRK")       |
| `--color`  | `-c`  |           | Accent color (hex)                 |
| `--columns`|       |           | Custom columns (JSON array)        |

**Output (JSON):**
```json
{
  "id": "abc123",
  "name": "Work",
  "prefix": "WRK",
  "columns": ["backlog", "todo", "in_progress", "review", "done"],
  "color": "#3B82F6",
  "task_count": 15
}
```

#### `egenskriven init`

Initialize EgenSkriven configuration for a project.

```bash
# Initialize with defaults
egenskriven init

# Initialize with specific workflow mode
egenskriven init --workflow strict

# Initialize with specific agent mode
egenskriven init --mode collaborative
```

Creates `.egenskriven/config.json` with project-specific agent configuration.

**Flags:**

| Flag         | Default   | Description                              |
|--------------|-----------|------------------------------------------|
| `--workflow` | light     | Workflow mode: strict, light, minimal    |
| `--mode`     | autonomous| Agent mode: autonomous, collaborative, supervised |

#### `egenskriven version`

Display version and build information.

```bash
egenskriven version
```

**Output (human):**
```
EgenSkriven v1.0.0
Build date: 2025-01-03T10:00:00Z
Go version: go1.21.0
```

**Output (JSON):**
```json
{
  "version": "1.0.0",
  "build_date": "2025-01-03T10:00:00Z",
  "go_version": "go1.21.0"
}
```

#### `egenskriven export`

Export tasks and boards to a file.

```bash
# Export all data as JSON
egenskriven export --format json > backup.json

# Export tasks only as CSV
egenskriven export --format csv > tasks.csv

# Export specific board
egenskriven export --board work --format json > work-backup.json
```

**Flags:**

| Flag       | Short | Default | Description                        |
|------------|-------|---------|------------------------------------|
| `--format` | `-f`  | json    | Output format: json, csv           |
| `--board`  | `-b`  |         | Export specific board only         |

#### `egenskriven import`

Import tasks and boards from a file.

```bash
# Import from JSON backup
egenskriven import backup.json

# Import with merge strategy (skip existing)
egenskriven import backup.json --strategy merge

# Import with replace strategy (overwrite existing)
egenskriven import backup.json --strategy replace
```

**Flags:**

| Flag         | Default | Description                              |
|--------------|---------|------------------------------------------|
| `--strategy` | merge   | Import strategy: merge, replace          |
| `--dry-run`  | false   | Preview changes without applying         |

#### `egenskriven template`

Manage task templates for quick creation.

```bash
# List templates
egenskriven template list

# Create a template
egenskriven template add bug-report --type bug --priority high --label bug

# Use template when creating task
egenskriven add "Login crashes on Safari" --template bug-report

# Delete template
egenskriven template delete bug-report
```

**Flags for `template add`:**

| Flag         | Description                              |
|--------------|------------------------------------------|
| `--type`     | Default task type                        |
| `--priority` | Default priority                         |
| `--column`   | Default column                           |
| `--label`    | Default labels (repeatable)              |
| `--epic`     | Default epic                             |

#### `egenskriven completion`

Generate shell completion scripts.

```bash
# Bash
egenskriven completion bash > /etc/bash_completion.d/egenskriven

# Zsh
egenskriven completion zsh > "${fpath[1]}/_egenskriven"

# Fish
egenskriven completion fish > ~/.config/fish/completions/egenskriven.fish

# PowerShell
egenskriven completion powershell > egenskriven.ps1
```

#### `egenskriven prime`

Output instructions for AI coding agents.

```bash
# Output agent instructions (for hook integration)
egenskriven prime

# Typically used in agent hooks, not directly
```

This command outputs a complete guide that teaches AI agents how to use EgenSkriven, including workflow patterns, CLI reference, and available options.

#### `egenskriven context`

Output project state summary for agent context.

```bash
# Get project context
egenskriven context --json
```

**Output (JSON):**
```json
{
  "boards": [
    {"id": "work", "name": "Work", "prefix": "WRK", "task_count": 15}
  ],
  "current_board": "work",
  "summary": {
    "total": 15,
    "by_column": {"todo": 5, "in_progress": 2, "review": 1, "done": 7},
    "by_priority": {"urgent": 1, "high": 3, "medium": 8, "low": 3}
  },
  "blocked_count": 2,
  "ready_count": 3
}
```

#### `egenskriven suggest`

Suggest tasks to work on next.

```bash
# Get work suggestions
egenskriven suggest --json

# Limit suggestions
egenskriven suggest --json --limit 3
```

**Flags:**

| Flag      | Short | Default | Description                        |
|-----------|-------|---------|------------------------------------|
| `--limit` | `-l`  | 5       | Maximum number of suggestions      |

**Output (JSON):**
```json
{
  "suggestions": [
    {
      "task": {
        "id": "abc123",
        "title": "Critical security bug",
        "type": "bug",
        "priority": "urgent",
        "column": "todo"
      },
      "reason": "Highest priority unblocked task"
    },
    {
      "task": {
        "id": "def456",
        "title": "Continue feature work",
        "type": "feature",
        "priority": "high",
        "column": "in_progress"
      },
      "reason": "Already in progress"
    }
  ]
}
```

**Suggestion logic:**
1. In-progress tasks (continue current work)
2. Urgent unblocked tasks
3. High priority unblocked tasks
4. Tasks that unblock the most other tasks

### Batch Input Format

For `--stdin` and `--file`, tasks are specified as JSON lines (one JSON object per line):

```json
{"title": "Task 1", "type": "bug", "priority": "high"}
{"title": "Task 2", "column": "todo"}
{"title": "Task 3", "id": "custom-id-001"}
```

Or as a JSON array:

```json
[
  {"title": "Task 1", "type": "bug"},
  {"title": "Task 2", "column": "todo"}
]
```

### Exit Codes

| Code | Meaning                                    |
|------|--------------------------------------------|
| 0    | Success                                    |
| 1    | General error                              |
| 2    | Invalid arguments/flags                    |
| 3    | Task not found                             |
| 4    | Ambiguous task reference                   |
| 5    | Validation error (invalid type, etc.)      |

### Error Output (JSON mode)

```json
{
  "error": {
    "code": 4,
    "message": "Ambiguous task reference: 'login' matches multiple tasks",
    "matches": [
      {"id": "abc123", "title": "Fix login crash"},
      {"id": "def456", "title": "Add login analytics"}
    ]
  }
}
```

## Key Implementation Details

### 1. Main Entry Point (cmd/kanban/main.go)

```go
package main

import (
    "log"

    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/core"
    
    "egenskriven/internal/commands"
    "egenskriven/ui"
)

func main() {
    app := pocketbase.New()

    // Register custom CLI commands
    commands.Register(app)

    // Serve embedded React frontend
    app.OnServe().BindFunc(func(e *core.ServeEvent) error {
        e.Router.GET("/{path...}", func(re *core.RequestEvent) error {
            path := re.Request.PathValue("path")
            
            if f, err := ui.DistFS.Open(path); err == nil {
                f.Close()
                return re.FileFS(ui.DistFS, path)
            }
            
            return re.FileFS(ui.DistFS, "index.html")
        })
        
        return e.Next()
    })

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```

### 2. Task Resolver (internal/resolver/resolver.go)

```go
package resolver

import (
    "fmt"
    "strings"

    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/core"
)

type Resolution struct {
    Task    *core.Record
    Matches []*core.Record // populated if ambiguous
}

func ResolveTask(app *pocketbase.PocketBase, ref string) (*Resolution, error) {
    // 1. Try exact ID match
    if task, err := app.FindRecordById("tasks", ref); err == nil {
        return &Resolution{Task: task}, nil
    }

    // 2. Try ID prefix match
    tasks, err := app.FindAllRecords("tasks", 
        dbx.NewExp("id LIKE {:prefix}", dbx.Params{"prefix": ref + "%"}),
    )
    if err == nil && len(tasks) == 1 {
        return &Resolution{Task: tasks[0]}, nil
    }

    // 3. Try title match (case-insensitive substring)
    tasks, err = app.FindAllRecords("tasks",
        dbx.NewExp("LOWER(title) LIKE {:title}", 
            dbx.Params{"title": "%" + strings.ToLower(ref) + "%"}),
    )
    if err != nil {
        return nil, err
    }

    switch len(tasks) {
    case 0:
        return nil, fmt.Errorf("no task found matching: %s", ref)
    case 1:
        return &Resolution{Task: tasks[0]}, nil
    default:
        return &Resolution{Matches: tasks}, nil // ambiguous
    }
}
```

### 3. Output Formatter (internal/output/output.go)

```go
package output

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/pocketbase/pocketbase/core"
)

type Formatter struct {
    JSON  bool
    Quiet bool
}

func (f *Formatter) Task(task *core.Record) {
    if f.JSON {
        json.NewEncoder(os.Stdout).Encode(taskToMap(task))
        return
    }

    fmt.Printf("✓ Created task: %s [%s]\n", 
        task.GetString("title"), 
        task.Id)
}

func (f *Formatter) Error(code int, message string, data any) {
    if f.JSON {
        json.NewEncoder(os.Stderr).Encode(map[string]any{
            "error": map[string]any{
                "code":    code,
                "message": message,
                "data":    data,
            },
        })
        return
    }

    fmt.Fprintf(os.Stderr, "Error: %s\n", message)
}

func taskToMap(task *core.Record) map[string]any {
    return map[string]any{
        "id":          task.Id,
        "title":       task.GetString("title"),
        "description": task.GetString("description"),
        "type":        task.GetString("type"),
        "priority":    task.GetString("priority"),
        "column":      task.GetString("column"),
        "position":    task.GetFloat("position"),
        "labels":      task.Get("labels"),
        "epic":        task.GetString("epic"),
        "created":     task.GetDateTime("created"),
        "updated":     task.GetDateTime("updated"),
    }
}
```

### 4. Add Command (internal/commands/add.go)

```go
package commands

import (
    "bufio"
    "encoding/json"
    "fmt"
    "os"

    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/core"
    "github.com/spf13/cobra"

    "egenskriven/internal/output"
)

func NewAddCmd(app *pocketbase.PocketBase, out *output.Formatter) *cobra.Command {
    var (
        taskType string
        priority string
        column   string
        labels   []string
        epic     string
        customID string
        stdin    bool
        file     string
    )

    cmd := &cobra.Command{
        Use:   "add [title]",
        Short: "Add a new task",
        Long: `Add a new task to the kanban board.

Supports batch creation via --stdin or --file for agent workflows.`,
        Example: `  egenskriven add "Implement dark mode"
  egenskriven add "Fix bug" --type bug --priority urgent
  egenskriven add "Setup CI" --id ci-setup-001
  echo '{"title":"Task 1"}' | egenskriven add --stdin`,
        RunE: func(cmd *cobra.Command, args []string) error {
            if err := app.Bootstrap(); err != nil {
                return err
            }

            // Handle batch input
            if stdin || file != "" {
                return addBatch(app, out, stdin, file)
            }

            if len(args) == 0 {
                return fmt.Errorf("title is required")
            }

            task, err := createTask(app, TaskInput{
                ID:       customID,
                Title:    args[0],
                Type:     taskType,
                Priority: priority,
                Column:   column,
                Labels:   labels,
                Epic:     epic,
            })
            if err != nil {
                return err
            }

            out.Task(task)
            return nil
        },
    }

    cmd.Flags().StringVarP(&taskType, "type", "t", "feature", "Task type (bug, feature, chore)")
    cmd.Flags().StringVarP(&priority, "priority", "p", "medium", "Priority (low, medium, high, urgent)")
    cmd.Flags().StringVarP(&column, "column", "c", "backlog", "Initial column")
    cmd.Flags().StringSliceVarP(&labels, "label", "l", nil, "Labels (repeatable)")
    cmd.Flags().StringVarP(&epic, "epic", "e", "", "Link to epic")
    cmd.Flags().StringVar(&customID, "id", "", "Custom ID for idempotency")
    cmd.Flags().BoolVar(&stdin, "stdin", false, "Read tasks from stdin")
    cmd.Flags().StringVarP(&file, "file", "f", "", "Read tasks from file")

    return cmd
}

type TaskInput struct {
    ID          string   `json:"id,omitempty"`
    Title       string   `json:"title"`
    Description string   `json:"description,omitempty"`
    Type        string   `json:"type,omitempty"`
    Priority    string   `json:"priority,omitempty"`
    Column      string   `json:"column,omitempty"`
    Labels      []string `json:"labels,omitempty"`
    Epic        string   `json:"epic,omitempty"`
}

func createTask(app *pocketbase.PocketBase, input TaskInput) (*core.Record, error) {
    collection, err := app.FindCollectionByNameOrId("tasks")
    if err != nil {
        return nil, err
    }

    record := core.NewRecord(collection)
    
    if input.ID != "" {
        record.SetId(input.ID)
    }
    
    record.Set("title", input.Title)
    record.Set("type", defaultString(input.Type, "feature"))
    record.Set("priority", defaultString(input.Priority, "medium"))
    record.Set("column", defaultString(input.Column, "backlog"))
    record.Set("position", getNextPosition(app, input.Column))
    
    if len(input.Labels) > 0 {
        record.Set("labels", input.Labels)
    }
    if input.Epic != "" {
        record.Set("epic", input.Epic)
    }
    if input.Description != "" {
        record.Set("description", input.Description)
    }

    if err := app.Save(record); err != nil {
        return nil, err
    }

    return record, nil
}

func addBatch(app *pocketbase.PocketBase, out *output.Formatter, useStdin bool, file string) error {
    var reader *bufio.Scanner

    if useStdin {
        reader = bufio.NewScanner(os.Stdin)
    } else {
        f, err := os.Open(file)
        if err != nil {
            return err
        }
        defer f.Close()
        reader = bufio.NewScanner(f)
    }

    var created []*core.Record
    for reader.Scan() {
        line := reader.Text()
        if line == "" {
            continue
        }

        var input TaskInput
        if err := json.Unmarshal([]byte(line), &input); err != nil {
            return fmt.Errorf("invalid JSON: %s", line)
        }

        task, err := createTask(app, input)
        if err != nil {
            return err
        }
        created = append(created, task)
    }

    for _, task := range created {
        out.Task(task)
    }
    return nil
}
```

### 5. Position Management

Using fractional indexing for positions to avoid rebalancing:

```go
func getNextPosition(app *pocketbase.PocketBase, column string) float64 {
    tasks, _ := app.FindAllRecords("tasks",
        dbx.NewExp("column = {:col}", dbx.Params{"col": column}),
        dbx.OrderBy("position DESC"),
        dbx.Limit(1),
    )
    
    if len(tasks) == 0 {
        return 1000.0
    }
    
    return tasks[0].GetFloat("position") + 1000.0
}

func getPositionBetween(before, after float64) float64 {
    return (before + after) / 2.0
}
```

### 6. React Frontend - PocketBase Hook (ui/src/hooks/usePocketBase.ts)

```typescript
import { useEffect, useState } from 'react';
import PocketBase, { RecordModel } from 'pocketbase';

const pb = new PocketBase('/');

export interface Task extends RecordModel {
  title: string;
  description?: string;
  type: 'bug' | 'feature' | 'chore';
  priority: 'low' | 'medium' | 'high' | 'urgent';
  column: 'backlog' | 'todo' | 'in_progress' | 'review' | 'done';
  position: number;
  board?: string;
  epic?: string;
  parent?: string;
  labels?: string[];
  blocked_by?: string[];
  due_date?: string;
  created_by?: 'user' | 'agent' | 'cli';
  created_by_agent?: string;
  history?: HistoryEntry[];
}

export interface HistoryEntry {
  timestamp: string;
  action: 'created' | 'updated' | 'moved' | 'completed' | 'deleted';
  actor: 'user' | 'agent' | 'cli';
  actor_detail?: string;
  changes?: {
    field: string;
    from: any;
    to: any;
  };
}

export function useTasks() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    pb.collection('tasks')
      .getFullList<Task>({ sort: 'position' })
      .then(setTasks)
      .finally(() => setLoading(false));

    pb.collection('tasks').subscribe<Task>('*', (e) => {
      if (e.action === 'create') {
        setTasks(prev => [...prev, e.record]);
      } else if (e.action === 'update') {
        setTasks(prev => prev.map(t => t.id === e.record.id ? e.record : t));
      } else if (e.action === 'delete') {
        setTasks(prev => prev.filter(t => t.id !== e.record.id));
      }
    });

    return () => {
      pb.collection('tasks').unsubscribe('*');
    };
  }, []);

  const moveTask = async (taskId: string, newColumn: string, newPosition: number) => {
    await pb.collection('tasks').update(taskId, {
      column: newColumn,
      position: newPosition,
    });
  };

  const createTask = async (task: Partial<Task>) => {
    return pb.collection('tasks').create(task);
  };

  return { tasks, loading, moveTask, createTask };
}
```

### 7. Kanban Board Component (ui/src/components/Board.tsx)

```tsx
import { useMemo } from 'react';
import {
  DndContext,
  DragEndEvent,
  closestCenter,
} from '@dnd-kit/core';
import { useTasks, Task } from '../hooks/usePocketBase';
import { Column } from './Column';

const COLUMNS = ['backlog', 'todo', 'in_progress', 'review', 'done'] as const;

export function Board() {
  const { tasks, loading, moveTask } = useTasks();

  const tasksByColumn = useMemo(() => {
    const grouped: Record<string, Task[]> = {};
    COLUMNS.forEach(col => grouped[col] = []);
    
    tasks.forEach(task => {
      if (grouped[task.column]) {
        grouped[task.column].push(task);
      }
    });
    
    Object.keys(grouped).forEach(col => {
      grouped[col].sort((a, b) => a.position - b.position);
    });
    
    return grouped;
  }, [tasks]);

  const handleDragEnd = async (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over) return;

    const taskId = active.id as string;
    const newColumn = over.data.current?.column as string;
    const newPosition = over.data.current?.position as number;

    if (newColumn) {
      await moveTask(taskId, newColumn, newPosition);
    }
  };

  if (loading) return <div>Loading...</div>;

  return (
    <DndContext collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
      <div className="flex gap-4 p-4 h-screen bg-gray-100">
        {COLUMNS.map(column => (
          <Column
            key={column}
            name={column}
            tasks={tasksByColumn[column]}
          />
        ))}
      </div>
    </DndContext>
  );
}
```

## Build & Development

### Makefile

```makefile
.PHONY: dev build clean

# Development: run React dev server + Go with Air
dev:
	@$(MAKE) -j2 dev-ui dev-go

dev-ui:
	cd ui && npm run dev

dev-go:
	air

# Build production binary
build: build-ui build-go

build-ui:
	cd ui && npm ci && npm run build

build-go:
	CGO_ENABLED=0 go build -o egenskriven ./cmd/egenskriven

# Cross-compile for all platforms
release: build-ui
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o dist/egenskriven-darwin-arm64 ./cmd/egenskriven
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o dist/egenskriven-darwin-amd64 ./cmd/egenskriven
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/egenskriven-linux-amd64 ./cmd/egenskriven
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o dist/egenskriven-windows-amd64.exe ./cmd/egenskriven

clean:
	rm -rf egenskriven dist/ ui/dist/
```

### Vite Config (ui/vite.config.ts)

```typescript
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
  server: {
    proxy: {
      '/api': 'http://localhost:8090',
      '/_': 'http://localhost:8090',
    },
  },
});
```

## CLI Usage Examples

```bash
# Start the server (web UI + API)
./egenskriven serve

# Basic task management
./egenskriven add "Implement dark mode" --type feature --priority medium
./egenskriven add "Fix login crash" --type bug --priority urgent --column todo
./egenskriven move abc123 in_progress
./egenskriven update abc123 --priority high --add-label critical
./egenskriven delete abc123

# Querying
./egenskriven list
./egenskriven list --column todo --type bug
./egenskriven list --priority urgent --json
./egenskriven show abc123

# Batch operations (agent-friendly)
./egenskriven add --stdin <<EOF
{"title": "Task 1", "type": "bug"}
{"title": "Task 2", "priority": "high"}
EOF

# JSON output for agents
./egenskriven list --json | jq '.tasks[] | select(.priority == "urgent")'

# Idempotent creation (safe for retries)
./egenskriven add "Setup CI" --id ci-setup-001
./egenskriven add "Setup CI" --id ci-setup-001  # no-op if exists

# Multi-board operations
./egenskriven board add "Work" --prefix WRK
./egenskriven board add "Personal" --prefix PER
./egenskriven board use work
./egenskriven add "Task in work board" --board work
./egenskriven list --all-boards

# Project initialization for agents
./egenskriven init --workflow strict --mode collaborative

# Templates
./egenskriven template add bug-report --type bug --priority high
./egenskriven add "Safari crash" --template bug-report

# Backup and restore
./egenskriven export --format json > backup.json
./egenskriven import backup.json

# View board in browser
open http://localhost:8090
```

## Real-time Sync: How It Works

1. **CLI creates task** → writes directly to SQLite via PocketBase's Go API
2. **PocketBase detects change** → broadcasts SSE event to subscribed clients
3. **React frontend** → receives event via `pb.collection('tasks').subscribe('*', ...)`
4. **UI updates** → React state updates, board re-renders

**Note**: CLI commands run in a separate process from `serve`. They share the SQLite database but not the in-memory event bus. Real-time works because PocketBase's SSE is database-driven.

## Future Enhancements

### Planned for V1 (see plan.md for phases)
- **Multi-board support**: Multiple boards with prefixed task IDs
- **Saved views**: Filter + display settings saved and accessible from sidebar
- **Due dates**: Task deadlines with overdue indicators
- **Sub-tasks**: Nested tasks with parent-child relationships
- **Epics UI**: Visual epic management in the web interface
- **Import/Export**: Backup and restore functionality
- **Task templates**: Predefined task structures for common patterns

### Post-V1
1. **Custom themes**: User-defined color themes via JSON files
2. **TUI mode**: Full terminal UI with Bubble Tea (`egenskriven tui` opens interactive view)
3. **Git integration**: Auto-create tasks from commit messages, link branches to tasks
4. **Archiving**: `egenskriven archive` to move done tasks to archive, keeping board clean
5. **Sync & Collaboration**: Optional cloud sync and multi-user support

## AI Agent Integration

EgenSkriven is designed to be **agent-native** - AI coding assistants can use it as their primary task tracking system, replacing built-in todo lists.

### Per-Project Configuration

Agent behavior is configurable per project via `.egenskriven/config.json`:

```json
{
  "agent": {
    "workflow": "strict",
    "mode": "autonomous",
    "overrideTodoWrite": true,
    "requireSummary": true,
    "structuredSections": true
  }
}
```

**Workflow modes:**

| Mode | Description |
|------|-------------|
| `strict` | Full workflow enforcement: create before work, update during, summary after |
| `light` | Basic tracking: create/complete tasks, no structured sections required |
| `minimal` | Just tracking: no workflow enforcement, agent decides when to use |

**Agent modes:**

| Mode | Description |
|------|-------------|
| `autonomous` | Agent executes actions directly. Human reviews asynchronously via activity history. Best for experienced users who trust the agent. |
| `collaborative` | Agent proposes major changes (complete, delete) and explains intent, but executes minor updates (status, priority). Human confirms major actions. |
| `supervised` | Agent is read-only. Can only query tasks and suggest actions. Outputs CLI commands for human to run. Best for learning or sensitive projects. |

**Configuration options:**

| Option | Default | Description |
|--------|---------|-------------|
| `workflow` | `light` | Workflow enforcement level |
| `mode` | `autonomous` | Agent autonomy level |
| `overrideTodoWrite` | `true` | Tell agents to ignore built-in todo tools |
| `requireSummary` | `false` | Require `## Summary of Changes` on completion |
| `structuredSections` | `false` | Encourage structured markdown sections in descriptions |

The `prime` command reads this config and adjusts its output accordingly.

### The Prime Command

The `egenskriven prime` command outputs a complete guide for AI agents:

```bash
# Output agent instructions (reads .egenskriven/config.json)
egenskriven prime

# Override workflow mode
egenskriven prime --mode strict
egenskriven prime --mode light
egenskriven prime --mode minimal
```

This outputs structured instructions that teach agents:
- How to discover available work
- How to create and update tasks
- Workflow patterns (based on configured mode)
- CLI reference with examples
- Available types, priorities, and columns

### Agent Workflow (Strict Mode)

```
1. DISCOVER WORK
   egenskriven list --json --ready

2. BEFORE STARTING
   - Check for existing task: egenskriven list --json --search "keyword"
   - Create if needed: egenskriven add "Title" --type feature --column todo --json
   - Move to in_progress: egenskriven move <id> in_progress

3. DURING WORK
   - Update description with progress notes
   - Add structured sections (## Approach, ## Open Questions)
   - Reference blocking tasks if parallelizable work is identified

4. AFTER COMPLETING
   - Add ## Summary of Changes section to description
   - Move to done: egenskriven move <id> done
   - Reference task ID in commit message

5. COMMIT
   - Include task ID: "feat: implement X [WRK-123]"
```

### Agent Workflow (Light Mode)

```
1. For substantial work, create a task
2. Update status when done
3. Reference task ID in commits when relevant
```

### Structured Description Sections

When `structuredSections` is enabled, agents are encouraged to use these markdown sections in task descriptions:

```markdown
## Approach
Brief description of how this will be implemented.

## Open Questions
- Question 1?
- Question 2?

## Checklist
- [ ] Step 1
- [ ] Step 2
- [x] Completed step

## Summary of Changes
What was actually done (filled in on completion).

## Follow-up
- Related work identified during implementation
```

Agents can update these sections as work progresses. The `## Summary of Changes` section is particularly valuable for project history and context.

### Hook Integration

Agents automatically receive context via lifecycle hooks:

**Claude Code** (`.claude/settings.json`):
```json
{
  "hooks": {
    "SessionStart": [
      { "hooks": [{ "type": "command", "command": "egenskriven prime" }] }
    ],
    "PreCompact": [
      { "hooks": [{ "type": "command", "command": "egenskriven prime" }] }
    ]
  }
}
```

**OpenCode** (`.opencode/plugin/egenskriven-prime.ts`):
```typescript
import type { Plugin } from "@opencode/plugin";

export const EgenSkrivenPlugin: Plugin = async ({ $ }) => {
  const prime = await $`egenskriven prime`.text();

  return {
    "experimental.chat.system.transform": async (_, output) => {
      output.system.push(prime);
    },
    "experimental.session.compacting": async (_, output) => {
      output.context.push(prime);
    },
  };
};

export default EgenSkrivenPlugin;
```

### Agent-Optimized CLI Features

#### JSON Output

All commands support `--json` for structured output:

```bash
egenskriven list --json
egenskriven show WRK-123 --json
egenskriven add "Title" --json
```

#### Field Selection

Reduce token usage by requesting only needed fields:

```bash
egenskriven list --json --fields id,title,column
```

Output:
```json
{
  "tasks": [
    {"id": "abc", "title": "Fix bug", "column": "todo"}
  ]
}
```

#### Ready Filter

Find actionable tasks (not blocked, not in progress/done):

```bash
egenskriven list --json --ready
```

#### Blocking Relationships

Track task dependencies for parallelization:

```bash
# Mark task as blocked by another
egenskriven update WRK-123 --blocked-by WRK-100

# Find blocked tasks
egenskriven list --json --is-blocked

# Find unblocked tasks (ready for parallel work)
egenskriven list --json --not-blocked --column todo
```

#### Context Summary

Get project state for agent context:

```bash
egenskriven context --json
```

Output:
```json
{
  "current_board": "work",
  "summary": {
    "total": 15,
    "by_column": {"todo": 5, "in_progress": 2, "done": 8},
    "by_priority": {"urgent": 1, "high": 3}
  },
  "blocked_count": 2,
  "ready_count": 3
}
```

#### Suggest Command

Get recommendations for what to work on:

```bash
egenskriven suggest --json
```

Output:
```json
{
  "suggestions": [
    {
      "task": {"id": "WRK-101", "title": "Critical bug", "priority": "urgent"},
      "reason": "Highest priority in todo"
    },
    {
      "task": {"id": "WRK-105", "title": "Unblocked task"},
      "reason": "Ready for parallel work"
    }
  ]
}
```

### Prime Command Template

The prime command uses an embedded template (`internal/commands/prime.tmpl`) that adapts based on project configuration:

```markdown
<EXTREMELY_IMPORTANT>
# EgenSkriven Task Tracker for Agents

This project uses **EgenSkriven**, a local-first kanban board.
{{if .OverrideTodoWrite}}
**Always use egenskriven instead of TodoWrite to manage your work and tasks.**
**Always use egenskriven instead of writing todo lists.**
{{end}}

All commands support `--json` for machine-readable output.

## Agent Mode: {{.AgentMode}}

{{if eq .AgentMode "autonomous"}}
You have full autonomy to create, update, and complete tasks directly.
Always identify yourself when making changes: use `--agent <your-name>` flag.
The human will review your actions asynchronously via the activity history.
{{else if eq .AgentMode "collaborative"}}
You can execute minor updates (status, priority, labels) directly.
For major actions (completing tasks, deleting), explain your intent and let the human confirm.
Example: "I believe task WRK-123 is complete. Run `egenskriven move WRK-123 done` to confirm."
{{else}}
You are in supervised/read-only mode. You can query tasks and make suggestions.
Output CLI commands for the human to execute. Do not run commands that modify tasks.
{{end}}

## Workflow Mode: {{.WorkflowMode}}

{{if eq .WorkflowMode "strict"}}
### BEFORE starting any task:
1. Check for existing task: `egenskriven list --json --search "keyword"`
2. Create if needed: `egenskriven add "Title" --type <type> --column todo --agent {{.AgentName}} --json`
3. Start work: `egenskriven move <id> in_progress`

### DURING work:
- Keep task description updated with progress
{{if .StructuredSections}}
- Use structured sections in description:
  - `## Approach` - How you plan to implement
  - `## Open Questions` - Uncertainties to resolve
  - `## Checklist` - Steps with checkboxes
{{end}}
- Identify blocking relationships for parallel work

### AFTER completing:
{{if .RequireSummary}}
- Add `## Summary of Changes` section describing what was done
{{end}}
- Mark complete: `egenskriven move <id> done`
- Reference task ID in commits: "feat: implement X [WRK-123]"
{{else if eq .WorkflowMode "light"}}
### Workflow
- Create task for substantial work: `egenskriven add "Title" --type <type> --agent {{.AgentName}} --json`
- Update status when done: `egenskriven move <id> done`
- Reference task ID in commits when relevant
{{else}}
### Workflow
- Use egenskriven for task tracking as needed
- All commands support `--json` for structured output
{{end}}

## Quick Reference

```bash
# Find ready tasks
egenskriven list --json --ready

# Show task details
egenskriven show <id> --json

# Create task (always include --agent to identify yourself)
egenskriven add "Title" --type bug --priority urgent --agent <your-name> --json

# Move task
egenskriven move <id> in_progress

# Update task
egenskriven update <id> --priority high --blocked-by <other-id>

# Get suggestions
egenskriven suggest --json

# Get project context
egenskriven context --json

# View activity history for a task
egenskriven show <id> --json  # includes history array
```

## Task Types
- bug: Something broken that needs fixing
- feature: New functionality
- chore: Maintenance, refactoring, docs

## Priorities
- urgent: Do immediately
- high: Do before normal work
- medium: Standard priority
- low: Can be delayed

## Columns
- backlog: Not yet planned
- todo: Ready to start
- in_progress: Currently working
- review: Awaiting review
- done: Completed

## Blocking Relationships

Use `--blocked-by` to track dependencies:
- `egenskriven update <id> --blocked-by <other-id>` - Mark as blocked
- `egenskriven list --not-blocked` - Find parallelizable work
- `egenskriven list --is-blocked` - See what's waiting
</EXTREMELY_IMPORTANT>
```

## Why This Architecture?

| Decision | Rationale |
|----------|-----------|
| PocketBase | Free real-time, admin UI, auth, REST API. Why rebuild? |
| Embedded React | Single binary distribution, no separate frontend deploy |
| SQLite | Perfect for local-first, no external DB needed |
| Cobra CLI | Industry standard, same as kubectl/docker/gh |
| @dnd-kit | Best React DnD library, accessible, small bundle |
| Agent-native CLI | Any AI can use it directly, replaces built-in todo lists |
| JSON output | Structured data for programmatic use |
| Fractional positions | Avoids rebalancing on every move |
| Blocking relationships | Enable agents to identify parallelizable work |
| Per-project config | Different workflows for different projects |
| Multi-board support | Separate contexts (work, personal, etc.) |
| Activity history | Track who changed what, when (human or agent) |
| Templates | Quick creation of common task patterns |
| Sub-tasks | Break down complex work into trackable pieces |
