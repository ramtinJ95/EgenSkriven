# Changelog

All notable changes to EgenSkriven will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Nothing yet

### Changed
- Nothing yet

### Fixed
- Nothing yet

## [0.1.1] - 2026-01-07

### Added
- **CLI**: OpenCode skill directory support for seamless integration
- **CLI**: New `skill` command with install/uninstall/status subcommands
- **Skills**: Core `egenskriven` skill file for AI agents
- **Skills**: `egenskriven-workflows` skill file documenting workflow modes
- **Skills**: `egenskriven-advanced` skill file for epics, dependencies, and batch operations
- **Docs**: AGENTS.md template for AI agent integration
- **Docs**: Skills system documentation
- **Docs**: Skills troubleshooting guide
- **Docs**: Migration guide and updated hook examples

### Changed
- Embedded skills are now the source of truth for skill content
- JSON mode defaults to global scope

### Fixed
- Handle multi-line YAML descriptions in skill files
- Log warning when skill removal fails instead of silent failure

### Removed
- Unused `ListEmbeddedSkills` function

## [0.1.0] - 2026-01-07

### Added

#### Core
- Single binary distribution with embedded React UI
- PocketBase backend with SQLite database
- Real-time sync between CLI and web UI via SSE subscriptions
- Hybrid mode: CLI works both online (HTTP API) and offline (direct DB access)

#### CLI Commands
- `add` - Create tasks with support for batch operations via stdin/file
- `list` - List and filter tasks with advanced filtering options
- `show` - Display detailed task information
- `move` - Move tasks between columns
- `update` - Update task properties
- `delete` - Delete tasks
- `board` - Multi-board management (add, list, show, use, delete)
- `epic` - Epic management (add, list, show, delete)
- `export` - Export to JSON or CSV format
- `import` - Import with merge or replace strategies
- `init` - Initialize project configuration
- `prime` - Generate agent instructions
- `context` - Show project state summary
- `suggest` - AI-friendly task prioritization
- `version` - Display version and build info
- `completion` - Shell completions for bash, zsh, fish, powershell
- `self-upgrade` - Update binary to latest version

#### Task Features
- Task types: bug, feature, chore
- Priority levels: urgent, high, medium, low
- Kanban columns: backlog, todo, in_progress, review, done
- Labels for categorization
- Blocking dependencies with circular detection
- Due dates with natural language support
- Sub-tasks with parent-child relationships
- Board-prefixed IDs (e.g., WRK-123, PER-456)

#### Agent Integration
- First-class AI agent support for Claude, GPT, OpenCode, Cursor
- Configurable workflows: strict, light, minimal
- Agent modes: autonomous, collaborative, supervised
- Agent tracking on task creation
- Override TodoWrite functionality

#### Web UI
- Kanban board with drag-and-drop
- List view toggle (Ctrl+B)
- Command palette (Cmd+K) with fuzzy search
- Full keyboard navigation (vim-style j/k/h/l)
- Task detail panel with Markdown support
- Quick create (C key)
- Peek preview (Space key)
- Property pickers (S/P/T keys)
- Filter builder with multiple conditions
- Saved views with favorites
- Multi-select with batch operations
- Activity log with relative timestamps

#### Theming
- 6 built-in themes: Dark, Light, Gruvbox Dark, Catppuccin Mocha, Nord, Tokyo Night
- Custom theme import via JSON
- System mode following OS preference
- 8 accent color presets

#### Distribution
- Cross-platform binaries (macOS, Linux, Windows)
- One-line installation script
- GitHub Actions release workflow
- Self-upgrade command
- SHA-256 checksums for verification

### Technical
- Go 1.21+ backend
- React 18 frontend with TypeScript
- PocketBase for data persistence
- Vite for frontend build
- Air for hot reload development
