# EgenSkriven

A local-first kanban task manager with CLI and web UI. Built with agentic workflows in mind from the ground up.

> **Early Access**: This project is functional but still in development. Feedback welcome!

## Features

- **Single binary** - Go backend with embedded React UI, no external dependencies
- **Real-time sync** - CLI changes appear instantly in the web UI
- **Agent-native** - First-class support for AI agents (Claude, GPT, etc.) with configurable workflows
- **Local-first** - All data stored locally in SQLite via PocketBase
- **Blocking dependencies** - Track task dependencies with circular detection

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
| `list` | List and filter tasks |
| `show <ref>` | Show task details |
| `move <ref> <column>` | Move task to column |
| `update <ref>` | Update task properties |
| `delete <ref>` | Delete a task |

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

The web UI provides a kanban board with:

- **Drag and drop** - Move tasks between columns
- **Real-time updates** - Changes from CLI appear instantly
- **Task details** - Click to view/edit task
- **Quick create** - Press `C` to create new task

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `C` | Open quick create modal |
| `Enter` | Open selected task detail |
| `Esc` | Close modal/deselect |

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
│   ├── commands/        # CLI commands (add, list, move, etc.)
│   ├── config/          # Configuration management
│   ├── output/          # Output formatting (human/JSON)
│   └── resolver/        # Task reference resolution
├── migrations/          # Database migrations
├── ui/                  # React frontend
│   ├── src/
│   │   ├── components/  # React components
│   │   ├── hooks/       # Custom hooks
│   │   └── lib/         # PocketBase client
│   └── ...
└── docs/                # Documentation
```

## Configuration

Project configuration is stored in `.egenskriven/config.json`:

```json
{
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
