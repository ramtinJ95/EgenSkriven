# EgenSkriven

A local-first kanban task manager with CLI and web UI. Built with agentic workflows in mind from the ground up.

> **Early Access**: This project is functional but still in development. Feedback welcome!

## Technology Stack

| Component | Technology | Version |
|-----------|------------|---------|
| Backend | Go | 1.25.5+ |
| Framework | PocketBase | 0.35.0 |
| Database | SQLite | embedded |
| CLI | Cobra | 1.10.2 |
| Frontend | React | 19.2.0 |
| Build Tool | Vite | 7.2.4 |
| State | Zustand | 5.0.9 |
| Drag & Drop | dnd-kit | 6.3.1 |
| Testing | Vitest + Playwright | 4.0.16 / 1.57.0 |

## Features

### Core
- **Single binary** - Go backend with embedded React UI, no external dependencies
- **Real-time sync** - CLI changes appear instantly in the web UI via SSE subscriptions
- **Local-first** - All data stored locally in SQLite via PocketBase
- **Hybrid mode** - CLI automatically falls back to direct database access when server is unavailable

### CLI
- **Full task management** - Add, list, show, move, update, delete tasks
- **Batch operations** - Create multiple tasks from JSON via stdin or file
- **Advanced filtering** - Filter by column, type, priority, labels, search, and more
- **Flexible task references** - Reference tasks by ID, ID prefix, display ID (WRK-123), or title substring

### Multi-Board Support
- **Multiple boards** - Create and manage separate boards for different projects
- **Board-prefixed IDs** - Tasks get board-specific IDs (e.g., WRK-123, PER-456)
- **Default board** - Set a default board for CLI commands
- **Board switcher** - Quick switch between boards in UI sidebar

### Agent Integration (AI-Native)
- **First-class AI agent support** - Designed for Claude Code, OpenCode, Cursor, Codex, etc.
- **Human-AI collaboration** - Session linking, blocking flow, resume workflow
- **Prime command** - Generate context-aware instructions for agents
- **Configurable workflows** - Strict, light, or minimal enforcement modes
- **Agent modes** - Autonomous, collaborative, or supervised agent behavior
- **Suggest command** - AI-friendly task prioritization
- **Context command** - Project state summary for agents
- **Skills system** - On-demand instruction loading for token efficiency

### Human-AI Collaborative Workflow
- **Session linking** - Agents link their sessions to tasks for tracking
- **Blocking flow** - When stuck, agents block tasks with questions for human input
- **Resume flow** - Humans answer questions, then resume agent with full context
- **Auto-resume** - Optional automatic resume when human comments with @agent
- **Resume modes** - command (print command), manual (copy command), auto (trigger on comment)
- **Comments** - Human-AI conversation threads on tasks

### Task Dependencies
- **Blocking relationships** - Track task dependencies with `--blocked-by`
- **Circular detection** - Prevents invalid dependency chains
- **Ready filter** - Find unblocked tasks ready to work on (`--ready`)
- **Blocking filters** - `--is-blocked` and `--not-blocked` filters

### Due Dates
- **Task due dates** - Set optional due dates with `--due` flag
- **Natural language** - Supports "today", "tomorrow", "next week", "next month"
- **Date formats** - ISO 8601 (YYYY-MM-DD), "Jan 15", "1/15/2025"
- **Date filters** - `--due-before`, `--due-after`, `--has-due`, `--no-due`
- **UI highlighting** - Overdue tasks shown in red, due today in orange

### Sub-tasks
- **Task hierarchies** - Create parent-child relationships with `--parent`
- **Sub-task filters** - `--has-parent` (sub-tasks only), `--no-parent` (top-level only)
- **Progress tracking** - Sub-task completion shown in task detail view
- **Inheritance** - Sub-tasks inherit context from parent task

### Epics
- **Epic management** - Group related tasks into epics
- **Board-scoped** - Epics belong to a specific board (use `--board` flag)
- **Epic CRUD** - Create, list, show, delete epics via CLI
- **Task linking** - Link tasks to epics with `--epic` flag
- **UI features** - Epic picker, sidebar list, detail view with progress

### Web UI
- **Kanban board** - Drag and drop tasks between columns
- **List view** - Toggle between board and table view with `Ctrl+B`
- **Virtualized columns** - Handles large task lists (50+ tasks) efficiently
- **Command palette** - Quick actions with `Cmd+K` and fuzzy search
- **Keyboard-driven** - Full keyboard navigation (j/k/h/l, vim-style shortcuts)
- **Task detail panel** - View and edit task properties with Markdown description support
- **Quick create** - Press `C` to create tasks instantly
- **Peek preview** - Press `Space` for quick task preview
- **Property pickers** - Keyboard shortcuts for status (S), priority (P), type (T)
- **Filter builder** - Advanced filtering with `F` key, supports multiple conditions
- **Saved views** - Save filter configurations as reusable views with favorites
- **Comments panel** - Human-AI conversation threads on tasks
- **Session info** - Track linked AI agent sessions
- **Activity log** - Task history with relative timestamps and actor tracking
- **Markdown editor** - Rich text editing with toolbar and preview mode
- **Date picker** - Calendar picker with shortcuts (Today, Tomorrow, Next Week)
- **Epic management** - Epic picker, sidebar list, and detail view with progress

### Theming
- **7 built-in themes** - Dark, Light, Gruvbox Dark, Catppuccin Mocha, Nord, Tokyo Night, Purple Dream
- **Custom themes** - Import your own JSON theme files with full color customization
- **System mode** - Follow OS dark/light preference with configurable theme per mode
- **Accent colors** - 8 preset accent colors or use theme's default
- **Real-time preview** - Theme changes apply instantly

## Installation

### Quick Install (macOS/Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/ramtinJ95/EgenSkriven/main/install.sh | sh
```

### Manual Download

Download the latest release for your platform from [Releases](https://github.com/ramtinJ95/EgenSkriven/releases):

| Platform | Binary |
|----------|--------|
| macOS (Apple Silicon) | `egenskriven-darwin-arm64` |
| macOS (Intel) | `egenskriven-darwin-amd64` |
| Linux (64-bit) | `egenskriven-linux-amd64` |
| Linux (ARM64) | `egenskriven-linux-arm64` |
| Windows (64-bit) | `egenskriven-windows-amd64.exe` |

### Build from Source

**Prerequisites:**
- Go 1.25.5+
- Node.js 18+ (for building UI)

```bash
git clone https://github.com/ramtinJ95/EgenSkriven
cd EgenSkriven
make build
```

This creates a single `./egenskriven` binary with the UI embedded.

## Quick Start

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

# Add a task with due date
./egenskriven add "Submit report" --due tomorrow
./egenskriven add "Quarterly review" --due "2025-01-15"

# Add a sub-task
./egenskriven add "Write unit tests" --parent abc123

# List tasks
./egenskriven list
./egenskriven list --column todo
./egenskriven list --type bug --priority urgent

# Show task details (use ID, ID prefix, display ID, or title)
./egenskriven show abc123
./egenskriven show WRK-123
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
| `epic list` | List all epics (use `--board` to filter) |
| `epic add <title>` | Create a new epic (use `--board` to specify board) |
| `epic show <ref>` | Show epic details |
| `epic delete <ref>` | Delete an epic |

### Human-AI Collaboration

| Command | Description |
|---------|-------------|
| `block <task> "message"` | Block task with question for human input |
| `comment <task> "message"` | Add comment to task |
| `comments <task>` | List task comments |
| `session link <task>` | Link agent session to task |
| `session show <task>` | Show linked session for task |
| `session history <task>` | Show session history |
| `session unlink <task>` | Unlink session from task |
| `resume <task>` | Resume blocked task with context |

### Agent Integration

| Command | Description |
|---------|-------------|
| `init` | Initialize project configuration |
| `prime` | Output agent instructions |
| `context` | Show project state summary |
| `suggest` | Suggest next task to work on |
| `skill install` | Install skills for AI agents |
| `skill uninstall` | Remove installed skills |
| `skill status` | Show skill installation status |

### Data Management

| Command | Description |
|---------|-------------|
| `export` | Export tasks and boards to JSON or CSV |
| `import <file>` | Import from backup file |
| `backup` | Create database backup |

### Utilities

| Command | Description |
|---------|-------------|
| `version` | Show version info |
| `completion <shell>` | Generate shell completions (bash, zsh, fish, powershell) |
| `self-upgrade` | Update to latest version |

### Global Flags

- `--json`, `-j` - Output in JSON format (machine-readable)
- `--quiet`, `-q` - Suppress non-essential output
- `--verbose`, `-v` - Show detailed output including connection method
- `--direct` - Skip HTTP API, use direct database access (faster, works offline)

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

### Task Reference

Tasks can be referenced by:
1. Full ID (e.g., `abc123def456`)
2. ID prefix (e.g., `abc`) - must be unique
3. Display ID (e.g., `WRK-123`)
4. Title substring (e.g., `"dark mode"`) - must be unique

## Human-AI Collaborative Workflow

EgenSkriven provides a complete workflow for AI agents to collaborate with humans asynchronously.

### How It Works

1. **Agent starts work**: Agent links its session to a task
2. **Agent gets stuck**: Agent blocks the task with a question
3. **Human responds**: Human sees blocked task, adds a comment with the answer
4. **Agent resumes**: Human resumes the agent with full context injected

### Session Linking

When an AI agent starts working on a task, it links its session:

```bash
# Link current session to task
egenskriven session link <task> --tool opencode --ref <session-id>

# Show linked session
egenskriven session show <task>

# View session history
egenskriven session history <task>
```

### Blocking Flow

When an agent needs human input:

```bash
# Block task with a question
egenskriven block <task> "Should I use PostgreSQL or SQLite for this feature?"

# List tasks needing human input
egenskriven list --need-input
```

### Resume Flow

After human provides input:

```bash
# Add comment with answer
egenskriven comment <task> "Use SQLite for simplicity"

# Resume the agent (injects full context)
egenskriven resume <task>

# Or with auto-execution (if configured)
egenskriven resume <task> --exec
```

### Resume Modes

Resume modes control what happens when you respond to a blocked task. When an agent needs your input, it "blocks" the task and waits. Resume mode determines how work continues after you answer.

| Mode | Behavior | Use When |
|------|----------|----------|
| `manual` | Command is printed, you copy and run it yourself | Learning the workflow, debugging, you want to see exactly what happens |
| `command` (default) | You explicitly run `egenskriven resume <task> --exec` | Normal development, you want explicit control over when the agent resumes |
| `auto` | Agent resumes automatically when you add a comment with `@agent` | Pair programming with AI, rapid back-and-forth, you want responsiveness |

Configure per board:
```bash
egenskriven board update <board> --resume-mode auto
```

### Session ID Discovery

| Tool | Method |
|------|--------|
| OpenCode | Call `egenskriven-session` tool |
| Claude Code | Use `$CLAUDE_SESSION_ID` env var |
| Codex | Run `.codex/get-session-id.sh` |

## Agent Integration

EgenSkriven is designed to work seamlessly with AI coding agents like Claude Code, OpenCode, Cursor, Codex, etc.

### Initialize Configuration

```bash
# Create .egenskriven/config.json
./egenskriven init --workflow strict --mode autonomous

# Initialize for specific tools
./egenskriven init --opencode
./egenskriven init --claude-code
./egenskriven init --codex
```

### Workflow Modes

Workflow modes control how strictly the agent must track work in tasks.

| Mode | Behavior | Use When |
|------|----------|----------|
| `strict` | Agent **must** use task tracking for everything: create/claim before starting, update status while working, mark complete after finishing | You need an audit trail, working in a team, or want full visibility into what the agent did and why |
| `light` (default) | Agent **should** track significant work: create tasks for features/bugs/multi-step work, mark done when finished, skip tracking for trivial stuff | Solo development, you want tracking without ceremony |
| `minimal` | Agent **decides** when tracking is useful: no requirements, agent uses judgment on what's worth tracking | Exploratory work, quick fixes, you trust the agent's judgment |

### Agent Modes

Agent modes control how much autonomy the agent has over task operations.

| Mode | Behavior | Use When |
|------|----------|----------|
| `autonomous` | Agent **acts first, you review later**: creates tasks without asking, updates status and notes freely, completes and closes tasks on its own | You trust the agent, want fast iteration, will review async |
| `collaborative` | Agent **acts on small things, explains big decisions**: can read tasks and make minor updates, but for major actions (completing, deleting, changing priority to urgent) it explains why and waits | You want the agent to be productive but still catch important decisions |
| `supervised` | Agent **can only look, not touch**: read-only access to task data, outputs commands for you to run (e.g., "Run: `egenskriven move FIX-1 done`") | Sensitive projects, onboarding a new agent, you want full control |

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

# Find tasks needing human input
./egenskriven list --need-input --json

# Create task with agent tracking
./egenskriven add "Fix bug" --agent claude --json

# All commands support JSON output
./egenskriven list --json --fields id,title,column
```

### Skills System

EgenSkriven provides a skills system for AI agents that support on-demand instruction loading. Skills are more token-efficient than always-on instructions.

#### Install Skills

Skills are installed to both Claude Code and OpenCode directories automatically:

| Location | Claude Code | OpenCode |
|----------|-------------|----------|
| Global | `~/.claude/skills/` | `~/.config/opencode/skill/` |
| Project | `.claude/skills/` | `.opencode/skill/` |

```bash
# Interactive installation
egenskriven skill install

# Install globally (all projects)
egenskriven skill install --global

# Install to current project only
egenskriven skill install --project

# Update existing skills
egenskriven skill install --force
```

#### Available Skills

| Skill | Description | When Agent Loads |
|-------|-------------|------------------|
| `egenskriven` | Core commands and task management | User mentions tasks, kanban, boards |
| `egenskriven-workflows` | Workflow modes and agent behaviors | Configuring workflow/agent modes |
| `egenskriven-advanced` | Epics, dependencies, sub-tasks, batch ops | Complex task relationships |

### Example: Claude Code Hook

Add to your Claude Code settings (`.claude/settings.json`):

```json
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          { "type": "command", "command": "egenskriven prime --agent claude" }
        ]
      }
    ],
    "PreCompact": [
      {
        "hooks": [
          { "type": "command", "command": "egenskriven prime --agent claude" }
        ]
      }
    ]
  }
}
```

## Web UI

The web UI provides a full-featured kanban board accessible at `http://localhost:8090` when the server is running.

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Cmd+K` | Open command palette |
| `C` | Create new task |
| `Enter` | Open selected task detail |
| `Space` | Peek preview |
| `E` | Edit task (open detail) |
| `Backspace` | Delete task |
| `F` | Open filter builder |
| `Ctrl+B` | Toggle board/list view |
| `Ctrl+,` | Open settings |
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

**Markdown Editor:**
| Key | Action |
|-----|--------|
| `Ctrl+B` | Bold |
| `Ctrl+I` | Italic |
| `Ctrl+E` | Inline code |
| `Ctrl+K` | Insert link |
| `Ctrl+Enter` | Save description |
| `Escape` | Cancel editing |

### Themes

| Theme | Style | Description |
|-------|-------|-------------|
| Dark | Dark | Default dark theme with blue accent |
| Light | Light | Clean light theme for daytime use |
| Gruvbox Dark | Dark | Warm retro groove colors |
| Catppuccin Mocha | Dark | Soothing pastel colors |
| Nord | Dark | Cool arctic bluish tones |
| Tokyo Night | Dark | Purple-ish city lights inspired |
| Purple Dream | Dark | Vibrant purple aesthetic |

#### Custom Themes

Import your own themes via JSON files in Settings. Required format:

```json
{
  "name": "My Custom Theme",
  "appearance": "dark",
  "author": "Your Name",
  "colors": {
    "bgApp": "#1a1b26",
    "bgSidebar": "#16161e",
    "bgCard": "#1f2335",
    "bgCardHover": "#292e42",
    "bgCardSelected": "#33394b",
    "bgInput": "#1a1b26",
    "bgOverlay": "rgba(0, 0, 0, 0.6)",
    "textPrimary": "#c0caf5",
    "textSecondary": "#9aa5ce",
    "textMuted": "#565f89",
    "textDisabled": "#414868",
    "borderSubtle": "#292e42",
    "borderDefault": "#33394b",
    "accent": "#7aa2f7",
    "accentHover": "#89b4fa",
    "accentMuted": "rgba(122, 162, 247, 0.2)",
    "shadowSm": "0 1px 2px rgba(0, 0, 0, 0.3)",
    "shadowMd": "0 4px 6px rgba(0, 0, 0, 0.4)",
    "shadowLg": "0 10px 15px rgba(0, 0, 0, 0.5)",
    "shadowDrag": "0 12px 24px rgba(0, 0, 0, 0.6)",
    "statusBacklog": "#6B7280",
    "statusTodo": "#c0caf5",
    "statusInProgress": "#e0af68",
    "statusReview": "#bb9af7",
    "statusDone": "#9ece6a",
    "statusCanceled": "#6B7280",
    "priorityUrgent": "#f7768e",
    "priorityHigh": "#ff9e64",
    "priorityMedium": "#e0af68",
    "priorityLow": "#6B7280",
    "priorityNone": "#414868",
    "typeBug": "#f7768e",
    "typeFeature": "#bb9af7",
    "typeChore": "#6B7280"
  }
}
```

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
./egenskriven list --need-input      # Tasks blocked awaiting human input

# Due date filters
./egenskriven list --due-before "2025-01-15"
./egenskriven list --due-after tomorrow
./egenskriven list --has-due         # Only tasks with due dates
./egenskriven list --no-due          # Only tasks without due dates

# Sub-task filters
./egenskriven list --has-parent      # Show only sub-tasks
./egenskriven list --no-parent       # Show only top-level tasks

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

## Epics

Group related tasks into epics:

```bash
# Create an epic (board-scoped)
./egenskriven epic add "Q1 Launch" --board work --color "#3B82F6"

# List epics
./egenskriven epic list --board work

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

## Export/Import/Backup

### Export

```bash
# Export all data to JSON
./egenskriven export --format json > backup.json

# Export tasks only to CSV
./egenskriven export --format csv > tasks.csv

# Export specific board
./egenskriven export --board work -o work-backup.json
```

### Import

```bash
# Import with merge strategy (skip existing, add new)
./egenskriven import backup.json

# Import with replace strategy (overwrite existing)
./egenskriven import backup.json --strategy replace

# Preview import without making changes
./egenskriven import backup.json --dry-run
```

### Backup

```bash
# Create database backup
./egenskriven backup

# Backup to specific location
./egenskriven backup --output /path/to/backup/
```

## Hybrid Mode (Online/Offline)

EgenSkriven supports a hybrid mode that allows the CLI to work both when the server is running and when it's offline:

### How It Works

1. **Default behavior**: CLI commands first try to connect to the HTTP API (server)
2. **Automatic fallback**: If the server is unavailable, CLI automatically uses direct database access
3. **Seamless experience**: Same commands work in both modes with identical output

### Usage

```bash
# Uses HTTP API if server is running, falls back to direct DB if not
./egenskriven list

# Force direct database access (faster, no real-time updates to UI)
./egenskriven list --direct

# Useful for offline work or when server is down
./egenskriven add "Offline task" --direct
./egenskriven move abc123 done --direct
```

### Trade-offs

| Mode | Real-time UI Updates | Speed | Requires Server |
|------|---------------------|-------|-----------------|
| HTTP API (default) | Yes | Normal | Yes (with fallback) |
| Direct (`--direct`) | No | Faster | No |

## Shell Completions

Enable tab completion for your shell:

```bash
# Bash (Linux)
egenskriven completion bash > /etc/bash_completion.d/egenskriven

# Bash (macOS with bash-completion@2)
egenskriven completion bash > $(brew --prefix)/etc/bash_completion.d/egenskriven

# Zsh
egenskriven completion zsh > "${fpath[1]}/_egenskriven"

# Fish
egenskriven completion fish > ~/.config/fish/completions/egenskriven.fish

# PowerShell
egenskriven completion powershell >> $PROFILE
```

## Self-Upgrade

Update EgenSkriven to the latest version:

```bash
# Check for updates and install if available
egenskriven self-upgrade

# Only check for updates (don't install)
egenskriven self-upgrade --check

# Force reinstall current version
egenskriven self-upgrade --force

# JSON output for scripting
egenskriven self-upgrade --check --json
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
make test       # Run Go unit tests
make test-ui    # Run UI unit tests (Vitest)
```

E2E tests use Playwright and can be run from the `ui/` directory.

### Build

```bash
make build    # Production binary with embedded UI
make clean    # Remove build artifacts
```

## Project Structure

```
.
├── cmd/egenskriven/        # Main entry point
├── internal/
│   ├── autoresume/         # Auto-resume AI sessions on @agent comments
│   ├── board/              # Board management logic
│   ├── commands/           # CLI commands
│   │   └── skills/         # Embedded skill files for AI agents
│   ├── config/             # Configuration management
│   ├── db/                 # Database utilities
│   ├── hooks/              # PocketBase event hooks
│   ├── output/             # Output formatting (human/JSON)
│   ├── resolver/           # Task reference resolution
│   ├── resume/             # Resume command building
│   └── testutil/           # Test utilities
├── migrations/             # Database migrations (18 total)
├── ui/                     # React frontend
│   ├── src/
│   │   ├── components/     # React components
│   │   ├── contexts/       # React contexts
│   │   ├── hooks/          # Custom hooks
│   │   ├── stores/         # Zustand stores
│   │   ├── themes/         # Theme system
│   │   ├── types/          # TypeScript types
│   │   └── lib/            # PocketBase client
│   └── embed.go            # Embeds UI into Go binary
├── tests/
│   ├── e2e/                # End-to-end tests
│   └── performance/        # Performance benchmarks
└── docs/                   # Documentation
```

## Database Schema

| Collection | Purpose |
|------------|---------|
| `tasks` | Core task storage with history tracking |
| `boards` | Multi-board support with resume modes |
| `epics` | Task grouping (board-scoped) |
| `comments` | Task discussions (human-AI threads) |
| `sessions` | AI agent session tracking |
| `views` | Saved filter configurations |

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
  },
  "server": {
    "url": "http://localhost:8090"
  }
}
```

## Deployment Scenarios

EgenSkriven can be deployed in different ways depending on how you want to organize your tasks. The database location is determined by the `--dir` flag or defaults to `pb_data/` in the current working directory.

### Scenario 1: Project-Specific (Recommended for Teams)

Each project has its own isolated task management with separate database and configuration.

**Setup:**
```bash
# In each project directory
cd ~/my-project
egenskriven init --workflow light
egenskriven serve
```

This creates:
- `.egenskriven/config.json` - Project configuration
- `pb_data/data.db` - Project database

**When to use:**
- Working on distinct projects that don't share tasks
- Team projects where tasks are version-controlled with code
- Projects with different workflow requirements (strict vs minimal)
- When you want to `.gitignore` the `pb_data/` but commit `.egenskriven/config.json`

**Pros:**
- Natural project isolation
- Config lives with the project and can be version controlled
- Each project can have different workflow modes
- Works seamlessly - no extra flags needed when in project directory

**Cons:**
- Must start a server per project if you want the web UI
- Tasks don't aggregate across projects
- Multiple databases to manage

**Recommended `.gitignore`:**
```gitignore
pb_data/
```

### Scenario 2: Global/Centralized (Recommended for Personal Use)

One database for all tasks, accessible from anywhere. Use boards to organize tasks by project.

**Setup:**
```bash
# 1. Create global data directory
mkdir -p ~/.egenskriven

# 2. Add shell alias (~/.bashrc, ~/.zshrc, or ~/.config/fish/config.fish)
alias egs='egenskriven --dir ~/.egenskriven'

# 3. Start global server (in a separate terminal or background)
egs serve --http :8090 &

# 4. Create boards for each project (or use --direct without server)
egs board add "Work" --prefix WRK
egs board add "Personal" --prefix PER
egs board add "SideProject" --prefix SIDE

# 5. Use from anywhere
egs add "Fix login bug" --board WRK
egs add "Buy groceries" --board PER
egs list --all-boards
```

**When to use:**
- Personal task management across all your work
- You want one unified view of all tasks
- You prefer one server running in the background
- Quick capture of tasks without switching contexts

**Pros:**
- All tasks in one place with unified search
- Single server to manage
- Cross-project visibility and planning
- Great for personal productivity workflows

**Cons:**
- Must use alias or `--dir` flag consistently
- No project-specific configuration (workflow modes apply globally)
- Database not tied to any specific project

**Recommended shell configuration:**
```bash
# ~/.bashrc or ~/.zshrc
export EGENSKRIVEN_DIR="$HOME/.egenskriven"
alias egs='egenskriven --dir "$EGENSKRIVEN_DIR"'

# Optional: Start server on login (background)
# (egs serve --http :8090 &) 2>/dev/null
```

**Fish shell:**
```fish
# ~/.config/fish/config.fish
set -gx EGENSKRIVEN_DIR "$HOME/.egenskriven"
alias egs="egenskriven --dir $EGENSKRIVEN_DIR"
```

### Scenario 3: Hybrid Approach

Global database for task storage, but project-specific configs for workflow settings.

**Setup:**
```bash
# Global alias for database location
alias egs='egenskriven --dir ~/.egenskriven'

# Project config for workflow settings
cd ~/my-project
egs init --workflow strict  # Creates .egenskriven/config.json
```

In this setup:
- Tasks are stored in the global `~/.egenskriven` database
- The project's `.egenskriven/config.json` controls workflow mode and default board
- Agents read project-local config for behavior settings

**Note:** The CLI reads config from the current directory, so project-specific settings (like `defaultBoard`) work when you're in that project, even with a global database.

### Running Multiple Servers

You can run multiple servers on different ports for different databases:

```bash
# Terminal 1: Global database on default port
egenskriven serve --dir ~/.egenskriven --http :8090

# Terminal 2: Specific project database
egenskriven serve --dir ~/work-project/pb_data --http :8091
```

Access different UIs:
- Global: `http://localhost:8090`
- Work project: `http://localhost:8091`

### Offline CLI Usage

The `--direct` flag allows CLI access without a running server:

```bash
# Works even when server is not running
egenskriven --direct list
egenskriven --direct add "Quick task"
egenskriven --direct move WRK-123 done
```

This is useful for:
- Quick task capture without starting a server
- CI/CD pipelines
- Scripts and automation
- Systems with limited resources

**Trade-off:** Direct mode doesn't trigger real-time UI updates. Changes will appear in the web UI on next page load or refresh.

### Comparison Table

| Aspect | Project-Specific | Global | Hybrid |
|--------|-----------------|--------|--------|
| Task isolation | Per project | Shared (use boards) | Shared (use boards) |
| Config location | Per project | N/A | Per project |
| Server management | One per project | Single global | Single global |
| CLI usage | Natural (no flags) | Requires `--dir` or alias | Requires `--dir` or alias |
| Best for | Teams, distinct projects | Personal productivity | Personal with project workflows |
| Offline support | Natural | With `--dir` flag | With `--dir` flag |

### Environment Variables

| Variable | Purpose |
|----------|---------|
| `EGENSKRIVEN_AUTHOR` | Default author name for comments |
| `EGENSKRIVEN_AGENT` | Default agent name for block command |

> **Note:** There is currently no environment variable for database directory. Use the `--dir` flag or shell aliases instead.

## License

MIT
