# EgenSkriven

A local-first kanban task manager with CLI and web UI. Built with agentic workflows in mind from the ground up.

> **Early Access**: This project is functional but still in development. Feedback welcome!

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

> **Breaking Change**: Epics are now board-scoped instead of global. Existing global epics were removed during migration.

### Web UI
- **Kanban board** - Drag and drop tasks between columns
- **List view** - Toggle between board and table view with `Ctrl+B`
- **Command palette** - Quick actions with `Cmd+K` and fuzzy search
- **Keyboard-driven** - Full keyboard navigation (j/k/h/l, shortcuts)
- **Task detail panel** - View and edit task properties with Markdown description support
- **Quick create** - Press `C` to create tasks instantly
- **Peek preview** - Press `Space` for quick task preview
- **Property pickers** - Keyboard shortcuts for status (S), priority (P), type (T)
- **Filter builder** - Advanced filtering with `F` key, supports multiple conditions
- **Saved views** - Save filter configurations as reusable views with favorites
- **Shortcuts help** - Press `?` to see all keyboard shortcuts
- **Sidebar** - Board navigation, saved views, epic list, and board creation
- **Markdown editor** - Rich text editing with toolbar and preview mode
- **Activity log** - Task history with relative timestamps and actor tracking
- **Date picker** - Calendar picker with shortcuts (Today, Tomorrow, Next Week)
- **Epic management** - Epic picker, sidebar list, and detail view with progress

### Theming
- **Multiple built-in themes** - Dark, Light, Gruvbox Dark, Catppuccin Mocha, Nord, Tokyo Night
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
- Go 1.21+
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
| `completion <shell>` | Generate shell completions |
| `self-upgrade` | Upgrade to latest version |

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

### Export/Import

| Command | Description |
|---------|-------------|
| `export` | Export tasks and boards to JSON or CSV |
| `import <file>` | Import from backup file |

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

### Global Flags

- `--json`, `-j` - Output in JSON format (machine-readable)
- `--quiet`, `-q` - Suppress non-essential output
- `--verbose`, `-v` - Show detailed output including connection method
- `--direct` - Skip HTTP API, use direct database access (faster, works offline)

### Shell Completions

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

Restart your shell or source the completion file after installing.

### Self-Upgrade

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

This injects task tracking instructions into Claude's context at session start and before context compaction.

**Recommended setup (skills + prime):**
1. Install skills globally: `egenskriven skill install --global`
2. Add the prime hook above for guaranteed context injection
3. Skills provide on-demand detailed instructions, prime provides baseline context

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

#### Check Installation Status

```bash
egenskriven skill status
```

#### Uninstall Skills

```bash
# Remove from all locations
egenskriven skill uninstall

# Remove from specific location
egenskriven skill uninstall --global
egenskriven skill uninstall --project
```

### AGENTS.md

EgenSkriven includes an `AGENTS.md` file that provides minimal always-loaded context for AI agents. This file is automatically read by agents supporting the AGENTS.md standard (Claude Code, OpenCode, Cursor, Codex, etc.).

**Customize AGENTS.md for your project:**

1. **Workflow mode**: Change "light" to "strict" or "minimal" in the Workflow section
2. **Quick commands**: Add project-specific common operations
3. **Board info**: If using multiple boards, mention the default board
4. **Project conventions**: Add any project-specific task naming or labeling conventions

The AGENTS.md file points agents to the skills system for detailed documentation, keeping always-loaded context minimal.

### Skills vs Prime Command

| Feature | Skills | Prime |
|---------|--------|-------|
| Token efficiency | High (on-demand) | Medium (~1-2k always) |
| Agent support | Claude Code, OpenCode | Any agent with hooks |
| Loading | Agent decides | Hook-injected |
| Best for | Modern agents | Legacy/hook workflows |

Use both together: Skills provide on-demand detail, prime provides hook-based injection.

### Migrating from Prime-Only to Skills

If you were using only the `prime` command before:

1. **Install skills**: `egenskriven skill install --global`
2. **Keep your prime hook**: The prime command continues to work alongside skills
3. **Benefits of adding skills**:
   - On-demand loading reduces context usage
   - More detailed instructions when needed
   - Agent can load specific skill based on task type

**No changes required to existing setups** - skills are additive, not a replacement.

### Skills Troubleshooting

**Skills not discovered by agent:**
1. Verify installation with `egenskriven skill status`
2. Check the agent is reading from the correct location (global vs project)
3. Restart the agent after installing skills
4. Ensure skill files have correct YAML frontmatter with `name` and `description`

**Skills installed but not loading:**
- Agent loads skills on-demand based on keywords in conversation
- Mention "task", "kanban", "egenskriven" to trigger skill loading
- Some agents show available skills in their tool descriptions

**Permission errors during install:**
- Global install requires write access to `~/.claude/skills/` and `~/.config/opencode/skill/`
- Try project install (`--project`) if global fails
- Check directory permissions: `ls -la ~/.claude/` and `ls -la ~/.config/opencode/`

**Updating skills after EgenSkriven upgrade:**
- Run `egenskriven skill install --force` to update embedded skills
- This overwrites existing skill files with the latest versions

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

**Markdown Editor (in description field):**
| Key | Action |
|-----|--------|
| `Ctrl+B` | Bold |
| `Ctrl+I` | Italic |
| `Ctrl+E` | Inline code |
| `Ctrl+K` | Insert link |
| `Ctrl+Enter` | Save description |
| `Escape` | Cancel editing |

**Selection:**
| Key | Action |
|-----|--------|
| `X` | Toggle select |
| `Shift+X` | Select range |
| `Cmd+A` | Select all visible |

### Saved Views

Save commonly used filter configurations as reusable views:

1. **Create a view**: Apply filters using the filter builder (`F`), then click "+ Save" in the sidebar
2. **Name the view**: Enter a descriptive name for your filter configuration
3. **Access views**: Saved views appear in the sidebar under "Views"
4. **Favorite views**: Star important views to pin them to the "Favorites" section
5. **Delete views**: Click the `×` button on any saved view to remove it

Views are board-specific and persist across sessions.

### Themes

EgenSkriven supports a comprehensive theme system with built-in themes and custom theme support.

#### Built-in Themes

| Theme | Style | Description |
|-------|-------|-------------|
| Dark | Dark | Default dark theme with blue accent |
| Light | Light | Clean light theme for daytime use |
| Gruvbox Dark | Dark | Warm retro groove colors |
| Catppuccin Mocha | Dark | Soothing pastel colors |
| Nord | Dark | Cool arctic bluish tones |
| Tokyo Night | Dark | Purple-ish city lights inspired |

#### Theme Selection

1. Open Settings with `Ctrl+,` or click the settings icon
2. Choose a theme from the grid (includes System option)
3. Theme applies instantly with all colors updating

#### System Mode

When "System" is selected, the app follows your OS dark/light preference:
- Configure which theme to use for dark mode (default: Dark)
- Configure which theme to use for light mode (default: Light)
- Automatically switches when your OS preference changes

#### Accent Colors

Customize the accent color used for buttons, links, and highlights:
- **Theme Default** - Use the theme's carefully chosen accent color (indicated by "T" badge)
- **8 preset colors** - Blue, Purple, Green, Orange, Pink, Cyan, Red, Yellow
- Accent color resets to theme default when switching themes

#### Custom Themes

Import your own themes via JSON files:

1. Click "Import Theme" in Settings
2. Select a valid JSON theme file
3. Theme appears in the selection grid and can be applied
4. Remove custom themes with the × button

**Custom Theme JSON Format:**

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

All color properties are required. Use hex colors (`#RRGGBB`) for most values, and `rgba()` for overlay/muted colors.

### Filter Builder (Web UI)

Press `F` to open the filter builder with support for:

- **Multiple conditions**: Add multiple filters that combine with AND logic
- **Filter types**: Status, Priority, Type, Labels, Due Date, Epic, Title
- **Operators**: is, is not, is any of, is set, is not set
- **Active filter pills**: Visual indicators showing active filters with quick remove
- **Clear all**: Remove all filters with one click

## Filtering Tasks (CLI)

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

### When to Use `--direct`

- **Offline work**: When you don't have the server running
- **Performance**: Direct DB access is slightly faster (no HTTP overhead)
- **Scripting**: When you don't need real-time UI updates
- **CI/CD**: In automated pipelines where server may not be available

### Trade-offs

| Mode | Real-time UI Updates | Speed | Requires Server |
|------|---------------------|-------|-----------------|
| HTTP API (default) | Yes | Normal | Yes (with fallback) |
| Direct (`--direct`) | No | Faster | No |

> **Note**: Changes made via `--direct` mode are persisted to the database and will be visible in the UI when you refresh or restart the server.

## Export/Import

Backup and migrate your data:

```bash
# Export all data to JSON
./egenskriven export --format json > backup.json

# Export tasks only to CSV
./egenskriven export --format csv > tasks.csv

# Export specific board
./egenskriven export --board work -o work-backup.json

# Import with merge strategy (skip existing, add new)
./egenskriven import backup.json

# Import with replace strategy (overwrite existing)
./egenskriven import backup.json --strategy replace

# Preview import without making changes
./egenskriven import backup.json --dry-run
```

**Export formats:**
- `json` - Full data export (boards, epics, tasks) - suitable for backup/restore
- `csv` - Tasks only - suitable for spreadsheet analysis

**Import strategies:**
- `merge` (default) - Skip records that already exist, add new ones
- `replace` - Overwrite existing records with same ID

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
│   │   ├── contexts/    # React contexts (ThemeContext, BoardContext, etc.)
│   │   ├── hooks/       # Custom hooks (useTasks, useKeyboard, useSelection, etc.)
│   │   ├── themes/      # Theme system (types, builtin themes, validation)
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
