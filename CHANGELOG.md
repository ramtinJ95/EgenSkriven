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

## [0.2.4] - 2026-01-11

### Added
- **UI**: Inline title editing in TaskDetail panel - click on task title to edit directly
- **UI**: Keyboard support for title editing (Enter/Space to activate, Enter to save, Escape to cancel)
- **UI**: Hover and focus styles for clickable task title

### Changed
- **Docs**: Reorganized documentation files into subdirectories (`docs/ai-workflow/`, `docs/core/`, `docs/tui-implementation/`)

## [0.2.3] - 2026-01-11

### Added
- **Config**: Global configuration support at `~/.config/egenskriven/config.json` for user-wide settings
- **Config**: `GlobalConfig` type for data directory, default author/agent names, and server URL
- **Config**: `MergedConfig` type that combines global and project configs with project taking precedence
- **Config**: Tilde expansion support in `data_dir` (e.g., `~/data` expands to `/home/user/data`)
- **CLI**: New `config` command with `show` and `path` subcommands to inspect configuration
- **CLI**: `config show` displays merged, global, or project config (`--global`, `--project` flags)
- **CLI**: `config path` shows config file locations
- **Tests**: Comprehensive test suite for global config loading, merging, and caching

### Changed
- **Config**: Data directory now configured via `data_dir` in global config instead of `EGENSKRIVEN_DIR` environment variable
- **Config**: Default agent name now configured via `defaults.agent` in global config instead of `EGENSKRIVEN_AGENT` environment variable
- **Config**: Default author name now configured via `defaults.author` in global config instead of environment variable
- **Config**: Server URL can now be set in global config and overridden per-project

### Improved
- **Performance**: Global config is cached using `sync.Once` to avoid repeated disk reads
- **Documentation**: Improved README with better explanations of workflow modes, agent behaviors, and resume modes

## [0.2.2] - 2026-01-10

### Fixed
- **CLI**: Fix `self-upgrade` command failing with "text file busy" error on Linux. The command now correctly renames the running binary before replacing it, which is allowed by the OS even while the binary is executing.

## [0.2.1] - 2026-01-10

### Added
- **UI**: Epic creation modal in sidebar - Create epics directly from the UI with title, description, and color picker
- **UI**: "+ New epic" button in EPICS section with real-time updates via existing subscription

### Fixed
- **UI**: Task detail panel width increased from 400px to 500px for better content spacing
- **UI**: DatePicker dropdown now right-aligns to prevent overflow outside panel bounds
- **UI**: EpicPicker dropdown now right-aligns to prevent overflow outside panel bounds

### Documentation
- **README**: Updated project documentation

## [0.2.0] - 2026-01-10

### Added

#### Human-AI Collaborative Workflow
- **CLI**: `block` command for AI agents to pause tasks and request human input
- **CLI**: `comment` command to add comments to tasks
- **CLI**: `comments` command to list task comments
- **CLI**: `resume` command to continue blocked tasks with context
- **CLI**: `session link` and `session show` commands for agent session tracking
- **CLI**: `--need-input` flag for list command to filter blocked tasks
- **CLI**: Agent session display in `show` command
- **CLI**: `board update` command with `--resume-mode` configuration
- **CLI**: `backup` command for database migrations
- **Schema**: Comments collection for task discussions
- **Schema**: Sessions collection for agent session tracking
- **Schema**: `need_input` column for blocked tasks requiring human input
- **Schema**: `agent_session` JSON field on tasks
- **Schema**: `resume_mode` field on boards (command, manual, auto)
- **Init**: `--opencode`, `--claude-code`, `--codex` flags for tool integration setup
- **Auto-resume**: Service to automatically resume agent sessions on comment with `@agent`
- **Hooks**: Comment hook registration for auto-resume trigger

#### Web UI - Collaborative Workflow
- **UI**: CommentsPanel component for viewing/adding task comments
- **UI**: SessionInfo component displaying linked agent sessions
- **UI**: ResumeModal component for generating resume commands
- **UI**: `need_input` column in kanban board with visual indicators
- **UI**: BoardSettingsModal for configuring resume mode
- **UI**: ResumeModeSelector component for board settings
- **UI**: Auto-resume indicator in CommentsPanel
- **UI**: Real-time comment subscriptions via PocketBase SSE
- **UI**: Reusable StatusBadge and PriorityBadge components

#### Performance & Testing
- **Database**: Performance indexes migration for improved query speed
- **Benchmark**: Command benchmark test infrastructure
- **Benchmark**: Resume context building benchmarks
- **Benchmark**: Auto-resume and session link benchmarks
- **Benchmark**: Scaling benchmarks for all operations
- **Tests**: E2E performance test suite
- **Tests**: Memory usage verification tests
- **Tests**: Query performance verification tests
- **Tests**: Index verification tests
- **Tests**: Comprehensive unit tests for block, comment, resume commands
- **Tests**: Auto-resume service unit and E2E tests
- **Tests**: UI component tests (CommentsPanel, SessionInfo, ResumeModal)

#### Documentation
- **Docs**: Human-AI Collaborative Workflow user guide
- **Docs**: Performance tuning guide
- **Docs**: Release checklist for collaborative workflow
- **Docs**: Performance notes documentation
- **Docs**: Updated AGENTS.md with collaborative workflow section
- **Skills**: Updated egenskriven, egenskriven-workflows skills with workflow documentation
- **Skills**: `need_input` column documentation in workflow states

### Changed
- Migration files now use timestamp prefixes for correct ordering
- Prime command template includes Tool field for agent identification
- Resume command accepts displayId as parameter (refactored from duplicate getDisplayId)

### Fixed
- **CLI**: Include `due_date` field in API task creation
- **CLI**: Resolve edge case bugs found in E2E testing (buildInFilter parameter collision, display ID resolution in move command, self-blocking validation)
- **CLI**: Handle multiple types for labels in `recordToTaskData`
- **CLI**: Add Tool field to PrimeTemplateData for prime command
- **CLI**: Use proper type handling for `blocked_by` in API calls
- **CLI**: Include parent field in API task data for sub-tasks
- **CLI**: Support display IDs in `--parent` and `--blocked-by` flags
- **Resume**: Add 'no comments' indicator in BuildMinimalPrompt
- **Resume**: Log error when session status update fails
- **Resume**: Use empty actor_detail for consistency with move command
- **Init**: Address code review issues for tool integrations
- **UI**: Pass currentBoard prop chain for auto-resume indicator
- **UI**: Add escape key handler and accessibility improvements to NewBoardModal
- **UI**: Align CSS variables with theme system for dark mode
- **UI**: Add need_input column to kanban board
- **UI**: Handle PocketBase SSE connection errors gracefully
- **Schema**: Add updated field to collections, use named constants
- **Schema**: Add idempotency check and nil pointer guard

### Improved
- **UI**: ARIA labels for accessibility
- **UI**: Keyboard navigation support (arrow keys, Enter, Escape)
- **UI**: React.memo optimization for CommentItem render performance

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
