# EgenSkriven

A local-first kanban task manager with CLI and web UI. Built with agentic workflows in mind from the ground up.

> **Early Access**: This project is functional but still in development. Feedback welcome!

## Features

### Core
- **Single binary** - Go backend with embedded React UI, no external dependencies
- **Real-time sync** - CLI changes appear instantly in the web UI via subscriptions
- **Local-first** - All data stored locally in SQLite via PocketBase

### CLI
- **Full task management** - Add, list, show, move, update, delete tasks
- **Batch operations** - Create multiple tasks from JSON via stdin or file
- **Advanced filtering** - Filter by column, type, priority, labels, search, and more
- **Flexible task references** - Reference tasks by ID, ID prefix, or title substring

### Multi-Board Support
- **Multiple boards** - Create and manage separate boards for different projects
- **Board-prefixed IDs** - Tasks get board-specific IDs (e.g., WRK-123, PER-456)
- **Default board** - Set a default board for CLI commands
- **Board switcher** - Quick switch between boards in UI sidebar

### Agent Integration (AI-Native)
- **First-class AI agent support** - Designed for Claude, GPT, OpenCode, Cursor, etc.
- **Prime command** - Generate context-aware instructions for agents
- **Configurable workflows** - Strict, light, or minimal enforcement modes
- **Agent modes** - Autonomous, collaborative, or supervised agent behavior
- **Suggest command** - AI-friendly task prioritization
- **Context command** - Project state summary for agents
- **Override TodoWrite** - Replace built-in agent task systems

### Task Dependencies
- **Blocking relationships** - Track task dependencies with `--blocked-by`
- **Circular detection** - Prevents invalid dependency chains
- **Ready filter** - Find unblocked tasks ready to work on (`--ready`)
- **Blocking filters** - `--is-blocked` and `--not-blocked` filters

### Epics
- **Epic management** - Group related tasks into epics
- **Epic CRUD** - Create, list, show, delete epics via CLI
- **Task linking** - Link tasks to epics with `--epic` flag

### Web UI
- **Kanban board** - Drag and drop tasks between columns
- **Command palette** - Quick actions with `Cmd+K` and fuzzy search
- **Keyboard-driven** - Full keyboard navigation (j/k/h/l, shortcuts)
- **Task detail panel** - View and edit task properties
- **Quick create** - Press `C` to create tasks instantly
- **Peek preview** - Press `Space` for quick task preview
- **Property pickers** - Keyboard shortcuts for status (S), priority (P), type (T)
- **Shortcuts help** - Press `?` to see all keyboard shortcuts
- **Sidebar** - Board navigation and creation

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+ (for building UI)

### Build

```bash
git clone https://github.com/ramtinJ95/EgenSkriven
cd EgenSkriven
make build
```

This creates a single `./egenskriven` binary with the UI embedded.

### Run

```bash
# Start the server (creates pb_data/ for database)
./egenskriven serve

# Web UI: http://localhost:8090
# Admin UI: http://localhost:8090/_/
```

### Basic CLI Usage

```bash
# Add a task
./egenskriven add "Implement dark mode"
./egenskriven add "Fix login crash" --type bug --priority urgent

# List tasks
./egenskriven list
./egenskriven list --column todo
./egenskriven list --type bug --priority urgent

# Show task details (use ID, ID prefix, or title)
./egenskriven show abc123
./egenskriven show "dark mode"

# Move between columns
./egenskriven move abc123 in_progress
./egenskriven move abc123 done

# Update task
./egenskriven update abc123 --priority high
./egenskriven update abc123 --blocked-by def456

# Delete task
./egenskriven delete abc123
```

## CLI Commands

### Task Management

| Command | Description |
|---------|-------------|
| `add <title>` | Create a new task |
| `add --stdin` | Batch create from JSON stdin |
| `add --file <path>` | Batch create from JSON file |
| `list` | List and filter tasks |
| `show <ref>` | Show task details |
| `move <ref> <column>` | Move task to column |
| `update <ref>` | Update task properties |
| `delete <ref>` | Delete a task |
| `version` | Show version info |

### Board Management

| Command | Description |
|---------|-------------|
| `board list` | List all boards |
| `board add <name> --prefix <PREFIX>` | Create a new board |
| `board show <ref>` | Show board details |
| `board use <ref>` | Set default board |
| `board delete <ref>` | Delete a board |

### Epic Management

| Command | Description |
|---------|-------------|
| `epic list` | List all epics |
| `epic add <title>` | Create a new epic |
| `epic show <ref>` | Show epic details |
| `epic delete <ref>` | Delete an epic |

### Agent Integration

| Command | Description |
|---------|-------------|
| `init` | Initialize project configuration |
| `prime` | Output agent instructions |
| `context` | Show project state summary |
| `suggest` | Suggest next task to work on |

### Global Flags

- `--json`, `-j` - Output in JSON format (machine-readable)
- `--quiet`, `-q` - Suppress non-essential output

### Task Reference

Tasks can be referenced by:
1. Full ID (e.g., `abc123def456`)
2. ID prefix (e.g., `abc`) - must be unique
3. Title substring (e.g., `"dark mode"`) - must be unique

## Task Properties

### Types
- `bug` - Something broken that needs fixing
- `feature` - New functionality
- `chore` - Maintenance, refactoring, docs

### Priorities
- `urgent` - Do immediately
- `high` - Do before normal work
- `medium` - Standard priority (default)
- `low` - Can be delayed

### Columns
- `backlog` - Not yet planned (default)
- `todo` - Ready to start
- `in_progress` - Currently working
- `review` - Awaiting review
- `done` - Completed

## Agent Integration

EgenSkriven is designed to work seamlessly with AI coding agents like Claude Code, OpenCode, Cursor, etc.

### Initialize Configuration

```bash
# Create .egenskriven/config.json
./egenskriven init --workflow strict --mode autonomous
```

**Workflow modes:**
- `strict` - Full enforcement: create before work, update during, summary after
- `light` - Basic tracking: create/complete tasks
- `minimal` - No enforcement: agent decides when to use

**Agent modes:**
- `autonomous` - Agent executes actions directly
- `collaborative` - Agent proposes major changes, executes minor ones
- `supervised` - Agent is read-only, outputs commands for human

### Get Agent Instructions

```bash
# Output instructions for your agent
./egenskriven prime --agent claude
```

The `prime` command outputs formatted instructions that tell the agent how to use EgenSkriven. Include this in your agent's system prompt or hooks.

### Agent-Friendly Commands

```bash
# Get project context (task counts by column/priority/type)
./egenskriven context --json

# Get task suggestions (prioritizes urgent, unblocking work)
./egenskriven suggest --json

# Find ready tasks (unblocked, in todo/backlog)
./egenskriven list --ready --json

# Create task with agent tracking
./egenskriven add "Fix bug" --agent claude --json

# All commands support JSON output
./egenskriven list --json --fields id,title,column
```

### Example: Claude Code Hook

Add to your Claude Code configuration:

```bash
# On session start
egenskriven prime --agent claude
```

This injects task tracking instructions into Claude's context.

## Web UI

The web UI provides a full-featured kanban board with:

- **Kanban board** - Tasks organized in columns (Backlog, Todo, In Progress, Review, Done)
- **Drag and drop** - Move tasks between columns with visual feedback
- **Real-time updates** - Changes from CLI appear instantly via subscriptions
- **Task details** - Click or press Enter to view/edit task properties
- **Quick create** - Press `C` to create new task with defaults
- **Peek preview** - Press `Space` on selected task for quick preview
- **Command palette** - Press `Cmd+K` for quick actions and search
- **Sidebar** - Board switcher with color indicators
- **Property pickers** - Quick property changes via keyboard

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Cmd+K` | Open command palette |
| `C` | Create new task |
| `Enter` | Open selected task detail |
| `Space` | Peek preview |
| `E` | Edit task (open detail) |
| `Backspace` | Delete task |
| `?` | Show shortcuts help |
| `Esc` | Close modal/deselect |

**Navigation:**
| Key | Action |
|-----|--------|
| `J` / `↓` | Next task |
| `K` / `↑` | Previous task |
| `H` / `←` | Previous column |
| `L` / `→` | Next column |

**Properties (when task selected):**
| Key | Action |
|-----|--------|
| `S` | Change status/column |
| `P` | Change priority |
| `T` | Change type |

**Selection:**
| Key | Action |
|-----|--------|
| `X` | Toggle select |
| `Shift+X` | Select range |
| `Cmd+A` | Select all visible |

## Filtering Tasks

The `list` command supports powerful filtering:

```bash
# Basic filters
./egenskriven list --column todo
./egenskriven list --type bug --priority urgent
./egenskriven list --label frontend --label ui

# Search by title
./egenskriven list --search "login"

# Agent/creator filters
./egenskriven list --created-by agent
./egenskriven list --agent claude

# Blocking filters
./egenskriven list --is-blocked      # Show blocked tasks
./egenskriven list --not-blocked     # Show unblocked tasks
./egenskriven list --ready           # Unblocked in todo/backlog

# Epic filter
./egenskriven list --epic "Q1 Launch"

# Board filter
./egenskriven list --board work
./egenskriven list --all-boards

# Sorting and limiting
./egenskriven list --sort "-priority,position"
./egenskriven list --limit 10

# JSON with selective fields
./egenskriven list --json --fields id,title,column
```

## Epics

Group related tasks into epics:

```bash
# Create an epic
./egenskriven epic add "Q1 Launch" --color "#3B82F6"

# List epics
./egenskriven epic list

# Link task to epic
./egenskriven add "Implement auth" --epic "Q1 Launch"

# Filter tasks by epic
./egenskriven list --epic "Q1 Launch"

# Show epic details
./egenskriven epic show "Q1 Launch"

# Delete epic
./egenskriven epic delete "Q1 Launch"
```

## Batch Operations

Create multiple tasks at once for agent workflows:

```bash
# From stdin (JSON lines)
echo '{"title":"Task 1"}
{"title":"Task 2","priority":"high"}' | ./egenskriven add --stdin

# From stdin (JSON array)
echo '[{"title":"Task 1"},{"title":"Task 2"}]' | ./egenskriven add --stdin

# From file
./egenskriven add --file tasks.json

# With agent tracking
echo '{"title":"Fix bug"}' | ./egenskriven add --stdin --agent claude
```

**JSON task format:**
```json
{
  "id": "optional15chars",
  "title": "Task title",
  "description": "Optional description",
  "type": "bug|feature|chore",
  "priority": "low|medium|high|urgent",
  "column": "backlog|todo|in_progress|review|done",
  "labels": ["label1", "label2"],
  "epic": "Epic title or ID"
}
```

## Multi-Board Support

Organize tasks across multiple boards:

```bash
# Create boards
./egenskriven board add "Work" --prefix WRK
./egenskriven board add "Personal" --prefix PER --color "#22C55E"

# List boards
./egenskriven board list

# Set default board
./egenskriven board use work

# Create task in specific board
./egenskriven add "Fix bug" --board work

# List tasks from specific board
./egenskriven list --board work

# List tasks from all boards
./egenskriven list --all-boards

# Show board details
./egenskriven board show work

# Delete board (with confirmation)
./egenskriven board delete work --force
```

Task IDs include board prefix (e.g., `WRK-123`, `PER-456`).

## Blocking Dependencies

Track task dependencies to identify parallelizable work:

```bash
# Mark task as blocked
./egenskriven update abc123 --blocked-by def456

# Add multiple blockers
./egenskriven update abc123 --add-blocked-by ghi789

# Remove blocker
./egenskriven update abc123 --remove-blocked-by def456

# Find blocked tasks
./egenskriven list --is-blocked

# Find unblocked tasks
./egenskriven list --not-blocked

# Find ready work (unblocked + in todo/backlog)
./egenskriven list --ready
```

## Development

### Setup

```bash
# Install Air for hot reload
go install github.com/air-verse/air@latest

# Install UI dependencies
cd ui && npm install && cd ..
```

### Development Servers

```bash
# Start both Go and React dev servers
make dev-all

# Or separately:
make dev      # Go server with hot reload (port 8090)
make dev-ui   # React dev server (port 5173)
```

### Testing

```bash
make test     # Run Go tests
make test-ui  # Run UI tests
```

### Build

```bash
make build    # Production binary with embedded UI
make clean    # Remove build artifacts
```

## Project Structure

```
.
├── cmd/egenskriven/     # Main entry point
├── internal/
│   ├── board/           # Board management logic
│   ├── commands/        # CLI commands (add, list, move, etc.)
│   ├── config/          # Configuration management
│   ├── output/          # Output formatting (human/JSON)
│   ├── resolver/        # Task reference resolution
│   └── testutil/        # Test utilities
├── migrations/          # Database migrations
├── ui/                  # React frontend
│   ├── src/
│   │   ├── components/  # React components (Board, TaskCard, CommandPalette, etc.)
│   │   ├── hooks/       # Custom hooks (useTasks, useKeyboard, useSelection, etc.)
│   │   ├── types/       # TypeScript types
│   │   └── lib/         # PocketBase client
│   └── ...
└── docs/                # Documentation
```

## Configuration

Project configuration is stored in `.egenskriven/config.json`:

```json
{
  "defaultBoard": "WRK",
  "agent": {
    "workflow": "strict",
    "mode": "autonomous",
    "overrideTodoWrite": true,
    "requireSummary": false,
    "structuredSections": false
  }
}
```

## License

MIT
