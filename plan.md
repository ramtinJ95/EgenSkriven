# EgenSkriven Implementation Plan

A phased implementation plan from zero to full product. Each phase ends with a testable, working version of the application.

---

## Overview

| Phase | Name | Duration Estimate | Deliverable |
|-------|------|-------------------|-------------|
| 0 | Project Setup | 1-2 days | Build system, dev environment |
| 1 | Core CLI | 3-5 days | Fully functional CLI (no UI) |
| 1.5 | Agent Integration | 2-3 days | Prime command, hooks, agent-friendly CLI |
| 2 | Minimal UI | 5-7 days | Basic board view, single board |
| 3 | Full CLI | 3-4 days | All CLI commands, batch ops |
| 4 | Interactive UI | 5-7 days | Keyboard shortcuts, command palette |
| 5 | Multi-Board | 3-4 days | Multiple boards, board switcher |
| 6 | Filtering & Views | 4-5 days | Filters, saved views, search |
| 7 | Polish | 5-7 days | Theming, animations, mobile |
| 8 | Advanced Features | 7-10 days | Epics, due dates, sub-tasks |
| 9 | Release | 2-3 days | Cross-compilation, distribution |

**Total estimated time: 7-11 weeks**

---

## Phase 0: Project Setup

**Goal:** Working development environment with build system.

### Tasks

#### 0.1 Initialize Go Module
```bash
go mod init github.com/yourusername/egenskriven
```

**Dependencies to add:**
- `github.com/pocketbase/pocketbase`
- `github.com/spf13/cobra` (comes with PocketBase)

#### 0.2 Create Project Structure
```
egenskriven/
├── cmd/
│   └── egenskriven/
│       └── main.go
├── internal/
│   ├── commands/
│   │   └── *_test.go     # Command integration tests
│   ├── output/
│   │   └── output_test.go
│   ├── resolver/
│   │   └── resolver_test.go
│   ├── testutil/
│   │   └── testutil.go   # Shared test helpers
│   └── hooks/
├── ui/
│   └── embed.go          # Placeholder for now
├── migrations/
├── go.mod
├── go.sum
├── Makefile
├── .gitignore
└── .air.toml
```

#### 0.3 Create Makefile
```makefile
.PHONY: dev build clean

dev:
	air

build:
	CGO_ENABLED=0 go build -o egenskriven ./cmd/egenskriven

clean:
	rm -rf egenskriven pb_data/
```

#### 0.4 Setup Air for Hot Reload
Create `.air.toml` for Go hot reloading during development.

#### 0.5 Create Minimal main.go
Bare PocketBase app that starts and creates `pb_data/`.

#### 0.6 Setup .gitignore
```
pb_data/
egenskriven
dist/
ui/dist/
ui/node_modules/
.air/
```

#### 0.7 Setup Testing Infrastructure

**Go testing setup:**
- Use Go's built-in `testing` package
- Add `github.com/stretchr/testify` for assertions and mocks
- Create `internal/testutil/` for shared test helpers

**Test helper file:** `internal/testutil/testutil.go`
```go
package testutil

import (
    "os"
    "testing"
    
    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/tests"
)

// NewTestApp creates a PocketBase instance with a temporary database
func NewTestApp(t *testing.T) *pocketbase.PocketBase {
    t.Helper()
    
    // Create temp directory for test database
    tmpDir, err := os.MkdirTemp("", "egenskriven-test-*")
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { os.RemoveAll(tmpDir) })
    
    app := pocketbase.New()
    // Configure for testing...
    return app
}
```

**Update Makefile:**
```makefile
.PHONY: dev build clean test test-coverage

test:
	go test ./... -v

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
```

### Test Criteria
- [ ] `make build` produces a binary
- [ ] `./egenskriven serve` starts PocketBase on port 8090
- [ ] Admin UI accessible at `http://localhost:8090/_/`
- [ ] `make dev` hot-reloads on Go file changes
- [ ] `make test` runs successfully (no tests yet, but infrastructure works)
- [ ] Test helper creates isolated test database

### Dependencies
None - this is the foundation.

---

## Phase 1: Core CLI

**Goal:** Functional CLI for basic task management. No UI yet.

### Tasks

#### 1.1 Create Database Migration
Define `tasks` collection with fields:
- id (string)
- title (string, required)
- description (string)
- type (select: bug, feature, chore)
- priority (select: low, medium, high, urgent)
- column (select: backlog, todo, in_progress, review, done)
- position (number)
- labels (json)
- blocked_by (json) - array of task IDs that block this task
- created_by (select: user, agent, cli) - who created this task
- created_by_agent (string, optional) - agent identifier (e.g., "claude", "opencode", "cursor")
- history (json) - array of activity entries for tracking changes

**History entry format:**
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "action": "created|updated|moved|completed",
  "actor": "user|agent|cli",
  "actor_detail": "claude",  // optional
  "changes": {
    "field": "column",
    "from": "todo",
    "to": "in_progress"
  }
}
```

**File:** `migrations/1_initial.go`

#### 1.2 Implement Output Formatter
**File:** `internal/output/output.go`

- `Formatter` struct with `JSON` and `Quiet` flags
- `Task()` - format single task
- `Tasks()` - format task list
- `Error()` - format error with code
- `Success()` - format success message

#### 1.3 Implement Task Resolver
**File:** `internal/resolver/resolver.go`

- `ResolveTask(app, ref)` - resolve by ID, ID prefix, or title
- Return `Resolution` with task or ambiguous matches
- Proper error types for not found / ambiguous

#### 1.4 Implement Root Command
**File:** `internal/commands/root.go`

- Global flags: `--json`, `--quiet`, `--data`
- Pass `Formatter` to subcommands
- Version command

#### 1.5 Implement `add` Command
**File:** `internal/commands/add.go`

- Basic: `egenskriven add "title"`
- Flags: `--type`, `--priority`, `--column`, `--label`, `--id`
- Flags for actor tracking: `--created-by` (user|agent|cli), `--agent` (agent identifier)
- Position management (append to column)
- Human and JSON output
- Record creation in history array

#### 1.6 Implement `list` Command
**File:** `internal/commands/list.go`

- List all tasks grouped by column
- Flags: `--column`, `--type`, `--priority` (basic filters)
- Flags for actor filtering: `--created-by` (user|agent|cli), `--agent` (filter by agent name)
- Human output: grouped by column
- JSON output: flat array with count

#### 1.7 Implement `show` Command
**File:** `internal/commands/show.go`

- Show single task details
- Use resolver for flexible task reference
- All fields displayed

#### 1.8 Implement `move` Command
**File:** `internal/commands/move.go`

- Move task to column: `egenskriven move <task> <column>`
- Position flags: `--position`, `--after`, `--before`
- Fractional position calculation

#### 1.9 Implement `update` Command
**File:** `internal/commands/update.go`

- Update any field: `--title`, `--type`, `--priority`, `--description`
- Label management: `--add-label`, `--remove-label`

#### 1.10 Implement `delete` Command
**File:** `internal/commands/delete.go`

- Delete single task with confirmation
- `--force` to skip confirmation
- Support multiple task IDs

#### 1.11 Define Exit Codes
Implement consistent exit codes:
- 0: Success
- 1: General error
- 2: Invalid arguments
- 3: Task not found
- 4: Ambiguous reference
- 5: Validation error

#### 1.12 Write Unit Tests

**File:** `internal/resolver/resolver_test.go`
```go
func TestResolveTask_ExactID(t *testing.T) {
    app := testutil.NewTestApp(t)
    // Create test task
    // Assert exact ID resolution works
}

func TestResolveTask_IDPrefix(t *testing.T) {
    // Test ID prefix matching
}

func TestResolveTask_TitleMatch(t *testing.T) {
    // Test title substring matching
}

func TestResolveTask_Ambiguous(t *testing.T) {
    // Test ambiguous reference returns matches
}

func TestResolveTask_NotFound(t *testing.T) {
    // Test not found error
}
```

**File:** `internal/output/output_test.go`
```go
func TestFormatter_Task_JSON(t *testing.T) {
    // Test JSON output format
}

func TestFormatter_Task_Human(t *testing.T) {
    // Test human-readable output
}

func TestFormatter_Error_JSON(t *testing.T) {
    // Test error JSON format with code
}
```

**File:** `internal/commands/position_test.go`
```go
func TestGetNextPosition_EmptyColumn(t *testing.T) {
    // Should return 1000.0
}

func TestGetNextPosition_WithExistingTasks(t *testing.T) {
    // Should return last position + 1000
}

func TestGetPositionBetween(t *testing.T) {
    // Test fractional position calculation
}
```

#### 1.13 Write Integration Tests

**File:** `internal/commands/commands_test.go`
```go
func TestAddCommand(t *testing.T) {
    app := testutil.NewTestApp(t)
    
    // Test basic add
    err := runCommand(app, "add", "Test task")
    assert.NoError(t, err)
    
    // Verify task exists
    tasks, _ := app.FindAllRecords("tasks")
    assert.Len(t, tasks, 1)
    assert.Equal(t, "Test task", tasks[0].GetString("title"))
}

func TestAddCommand_WithFlags(t *testing.T) {
    // Test --type, --priority, --column flags
}

func TestAddCommand_CustomID(t *testing.T) {
    // Test --id flag for idempotency
}

func TestListCommand_GroupedByColumn(t *testing.T) {
    // Create tasks in different columns
    // Verify output groups correctly
}

func TestMoveCommand(t *testing.T) {
    // Create task, move to different column
    // Verify column and position updated
}

func TestDeleteCommand_WithConfirmation(t *testing.T) {
    // Test delete requires confirmation
}

func TestDeleteCommand_Force(t *testing.T) {
    // Test --force skips confirmation
}
```

### Test Criteria

**Automated Tests (must pass):**
- [ ] `make test` passes with all unit tests
- [ ] Resolver tests cover: exact ID, prefix, title match, ambiguous, not found
- [ ] Output formatter tests cover: JSON and human formats
- [ ] Position calculation tests cover: empty column, existing tasks, between positions
- [ ] Integration tests cover: add, list, show, move, update, delete commands

**Manual Verification:**
- [ ] `egenskriven add "Test task"` creates a task
- [ ] `egenskriven list` shows tasks grouped by column
- [ ] `egenskriven show <id>` displays task details (including blocked_by)
- [ ] `egenskriven move <id> in_progress` moves task
- [ ] `egenskriven update <id> --priority urgent` updates task
- [ ] `egenskriven delete <id>` deletes task (with confirmation)
- [ ] `egenskriven list --json` outputs valid JSON
- [ ] Task resolution works by ID prefix and title substring
- [ ] Ambiguous references return error with matches

### Dependencies
- Phase 0 complete

---

## Phase 1.5: Agent Integration

**Goal:** Make EgenSkriven agent-native so AI coding assistants use it as their primary task tracker.

### Philosophy

EgenSkriven should replace built-in agent todo systems (like `TodoWrite`). Agents should:
1. Automatically receive context about the task system via hooks
2. Use CLI commands for all task management
3. Be able to identify parallelizable work via blocking relationships
4. Get structured JSON output optimized for token efficiency

### Per-Project Configuration

Agent behavior is configurable per project, allowing different workflows for different contexts:

**Configuration file:** `.egenskriven/config.json`

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
- `strict`: Full enforcement (create before, update during, summary after)
- `light`: Basic tracking (create/complete, no structured sections)
- `minimal`: No enforcement (agent decides when to use)

**Agent modes:**
- `autonomous`: Agent executes actions directly (create, update, complete tasks). Human reviews asynchronously via activity history.
- `collaborative`: Agent proposes actions and explains intent, but waits for human to confirm before major changes (completing tasks, deleting). Minor updates allowed.
- `supervised`: Agent can only read tasks and make suggestions. Human executes all changes. Agent outputs commands for human to run.

### Tasks

#### 1.5.1 Add Blocking Relationships to Data Model

Update `tasks` collection migration:
```go
// Add blocked_by field
record.Set("blocked_by", []string{}) // Array of task IDs
```

**Validation rules:**
- Cannot block self
- Cannot create circular dependencies
- Referenced tasks must exist

#### 1.5.2 Implement Config Loading

**File:** `internal/config/config.go`

```go
package config

import (
    "encoding/json"
    "os"
    "path/filepath"
)

type AgentConfig struct {
    Workflow           string `json:"workflow"`           // strict, light, minimal
    Mode               string `json:"mode"`               // autonomous, collaborative, supervised
    OverrideTodoWrite  bool   `json:"overrideTodoWrite"`
    RequireSummary     bool   `json:"requireSummary"`
    StructuredSections bool   `json:"structuredSections"`
}

type Config struct {
    Agent AgentConfig `json:"agent"`
}

func LoadProjectConfig() (*Config, error) {
    // Look for .egenskriven/config.json in current directory
    configPath := filepath.Join(".egenskriven", "config.json")
    
    data, err := os.ReadFile(configPath)
    if os.IsNotExist(err) {
        // Return defaults
        return &Config{
            Agent: AgentConfig{
                Workflow:           "light",
                Mode:               "autonomous",
                OverrideTodoWrite:  true,
                RequireSummary:     false,
                StructuredSections: false,
            },
        }, nil
    }
    if err != nil {
        return nil, err
    }
    
    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }
    
    return &cfg, nil
}
```

#### 1.5.3 Implement Prime Command

**File:** `internal/commands/prime.go`

```go
//go:embed prime.tmpl
var agentPromptTemplate string

var primeMode string

var primeCmd = &cobra.Command{
    Use:   "prime",
    Short: "Output instructions for AI coding agents",
    Long: `Outputs a complete guide for AI agents to use EgenSkriven.
    
Reads configuration from .egenskriven/config.json in the current project.
Use --mode to override the workflow mode.

This is typically called automatically via agent hooks (Claude, OpenCode)
rather than manually.`,
    RunE: func(cmd *cobra.Command, args []string) error {
        cfg, err := config.LoadProjectConfig()
        if err != nil {
            return err
        }
        
        // Override mode if specified
        workflowMode := cfg.Agent.Workflow
        if primeMode != "" {
            workflowMode = primeMode
        }
        
        tmpl, err := template.New("prime").Parse(agentPromptTemplate)
        if err != nil {
            return err
        }
        
        return tmpl.Execute(os.Stdout, TemplateData{
            WorkflowMode:       workflowMode,
            AgentMode:          cfg.Agent.Mode,
            OverrideTodoWrite:  cfg.Agent.OverrideTodoWrite,
            RequireSummary:     cfg.Agent.RequireSummary,
            StructuredSections: cfg.Agent.StructuredSections,
            Types:              []string{"bug", "feature", "chore"},
            Priorities:         []string{"low", "medium", "high", "urgent"},
            Columns:            []string{"backlog", "todo", "in_progress", "review", "done"},
        })
    },
}

func init() {
    primeCmd.Flags().StringVar(&primeMode, "mode", "", 
        "Override workflow mode (strict, light, minimal)")
}
```

#### 1.5.4 Create Prime Template

**File:** `internal/commands/prime.tmpl`

Template content (see kanban-architecture.md for full template).

Key sections:
- `<EXTREMELY_IMPORTANT>` wrapper for priority
- Conditional workflow instructions based on workflow mode (strict/light/minimal)
- Conditional behavior based on agent mode (autonomous/collaborative/supervised)
- Override TodoWrite instruction (configurable)
- Structured sections guidance (when enabled)
- CLI quick reference
- Blocking relationship usage
- Types, priorities, columns reference
- Actor identification (agent must identify itself via `--created-by-agent` flag)

The template uses Go template conditionals to adapt output:
```go
{{if eq .WorkflowMode "strict"}}
// Strict workflow instructions
{{else if eq .WorkflowMode "light"}}
// Light workflow instructions
{{else}}
// Minimal instructions
{{end}}

{{if eq .AgentMode "autonomous"}}
// Execute actions directly, human reviews async
{{else if eq .AgentMode "collaborative"}}
// Propose major changes, execute minor ones
// For completing/deleting: explain intent, let human confirm
{{else}}
// Supervised: read-only, output commands for human to run
{{end}}
```

#### 1.5.5 Add `--ready` Filter to List

```go
// In list.go
var listReady bool

func init() {
    listCmd.Flags().BoolVar(&listReady, "ready", false, 
        "Show unblocked tasks in todo/backlog (agent-friendly)")
}

// In RunE:
if listReady {
    // Filter: column in (todo, backlog) AND not blocked
    filter.Column = []string{"todo", "backlog"}
    filter.NotBlocked = true
}
```

#### 1.5.6 Add Blocking Filters to List

```go
var listIsBlocked bool
var listNotBlocked bool

func init() {
    listCmd.Flags().BoolVar(&listIsBlocked, "is-blocked", false, 
        "Show only tasks blocked by others")
    listCmd.Flags().BoolVar(&listNotBlocked, "not-blocked", false, 
        "Show only tasks not blocked by others")
}
```

#### 1.5.7 Add `--fields` Flag for Selective Output

```go
var listFields string

func init() {
    listCmd.Flags().StringVar(&listFields, "fields", "", 
        "Comma-separated fields to include in JSON output")
}

// In output formatting:
if listFields != "" {
    fields := strings.Split(listFields, ",")
    // Filter output to only include specified fields
}
```

#### 1.5.8 Add Blocking Management to Update

```go
var updateBlockedBy []string
var updateRemoveBlockedBy []string

func init() {
    updateCmd.Flags().StringArrayVar(&updateBlockedBy, "blocked-by", nil,
        "Add blocking task ID (repeatable)")
    updateCmd.Flags().StringArrayVar(&updateRemoveBlockedBy, "remove-blocked-by", nil,
        "Remove blocking task ID (repeatable)")
}
```

#### 1.5.9 Implement Context Command

**File:** `internal/commands/context.go`

```go
var contextCmd = &cobra.Command{
    Use:   "context",
    Short: "Output project state summary for agents",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Gather statistics
        tasks, _ := getAllTasks(app)
        
        summary := ContextSummary{
            CurrentBoard: getCurrentBoard(),
            Summary: Summary{
                Total:      len(tasks),
                ByColumn:   countByColumn(tasks),
                ByPriority: countByPriority(tasks),
            },
            BlockedCount: countBlocked(tasks),
            ReadyCount:   countReady(tasks),
        }
        
        return outputJSON(summary)
    },
}
```

#### 1.5.10 Implement Suggest Command

**File:** `internal/commands/suggest.go`

```go
var suggestCmd = &cobra.Command{
    Use:   "suggest",
    Short: "Suggest tasks to work on next",
    RunE: func(cmd *cobra.Command, args []string) error {
        suggestions := []Suggestion{}
        
        // 1. In-progress tasks (continue current work)
        inProgress := getTasksByColumn("in_progress")
        for _, t := range inProgress {
            suggestions = append(suggestions, Suggestion{
                Task:   t,
                Reason: "Continue current work",
            })
        }
        
        // 2. Urgent unblocked tasks
        urgent := getUnblockedByPriority("urgent")
        for _, t := range urgent {
            suggestions = append(suggestions, Suggestion{
                Task:   t,
                Reason: "Urgent priority, unblocked",
            })
        }
        
        // 3. High priority unblocked
        high := getUnblockedByPriority("high")
        for _, t := range high {
            suggestions = append(suggestions, Suggestion{
                Task:   t,
                Reason: "High priority, unblocked",
            })
        }
        
        // 4. Tasks that unblock others
        unblocking := getTasksThatUnblockMost()
        for _, t := range unblocking {
            suggestions = append(suggestions, Suggestion{
                Task:   t,
                Reason: fmt.Sprintf("Unblocks %d other tasks", t.UnblocksCount),
            })
        }
        
        return outputJSON(SuggestResponse{Suggestions: suggestions[:limit]})
    },
}
```

#### 1.5.11 Create OpenCode Plugin

**File:** `.opencode/plugin/egenskriven-prime.ts`

```typescript
import type { Plugin } from "@opencode/plugin";

export const EgenSkrivenPlugin: Plugin = async ({ $ }) => {
  // Check if egenskriven is available
  try {
    await $`egenskriven version`.quiet();
  } catch {
    return {}; // Not installed, skip
  }

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

#### 1.5.12 Document Claude Code Integration

Add to README and create example `.claude/settings.json`:

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

#### 1.5.13 Update Show Command for Blocking Info

Display blocking relationships in show output:

```
Task: abc123def456
Title:       Fix login crash
...
Blocked by:  def456 (Setup auth system)
Blocks:      ghi789 (Add user dashboard)
```

#### 1.5.14 Implement Init Command for Config

**File:** `internal/commands/init.go`

```go
var initCmd = &cobra.Command{
    Use:   "init",
    Short: "Initialize EgenSkriven configuration for a project",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Create .egenskriven directory
        if err := os.MkdirAll(".egenskriven", 0755); err != nil {
            return err
        }
        
        // Create default config
        cfg := Config{
            Agent: AgentConfig{
                Workflow:           "light",
                OverrideTodoWrite:  true,
                RequireSummary:     false,
                StructuredSections: false,
            },
        }
        
        data, _ := json.MarshalIndent(cfg, "", "  ")
        return os.WriteFile(".egenskriven/config.json", data, 0644)
    },
}
```

Usage:
```bash
# Initialize with defaults
egenskriven init

# Initialize with strict mode
egenskriven init --workflow strict
```

#### 1.5.15 Write Unit Tests for Agent Features

**File:** `internal/config/config_test.go`
```go
func TestLoadProjectConfig_Defaults(t *testing.T) {
    // No config file should return defaults
}

func TestLoadProjectConfig_FromFile(t *testing.T) {
    // Create temp config file, verify it loads
}

func TestLoadProjectConfig_InvalidJSON(t *testing.T) {
    // Invalid JSON should return error
}
```

**File:** `internal/commands/blocking_test.go`
```go
func TestBlockedBy_PreventsSelfBlocking(t *testing.T) {
    // Task cannot block itself
}

func TestBlockedBy_PreventsCircularDependency(t *testing.T) {
    // A blocks B, B cannot block A
}

func TestBlockedBy_ReferencedTaskMustExist(t *testing.T) {
    // Cannot block by non-existent task
}

func TestListReady_FiltersCorrectly(t *testing.T) {
    // --ready shows only unblocked todo/backlog tasks
}

func TestListIsBlocked_FiltersCorrectly(t *testing.T) {
    // --is-blocked shows only blocked tasks
}
```

**File:** `internal/commands/prime_test.go`
```go
func TestPrime_OutputChangesWithWorkflowMode(t *testing.T) {
    modes := []string{"strict", "light", "minimal"}
    for _, mode := range modes {
        // Verify output differs per mode
    }
}

func TestPrime_RespectsConfigFile(t *testing.T) {
    // Create config with strict mode
    // Verify prime output matches
}
```

**File:** `internal/commands/suggest_test.go`
```go
func TestSuggest_PrioritizesInProgress(t *testing.T) {
    // In-progress tasks should come first
}

func TestSuggest_PrioritizesUrgent(t *testing.T) {
    // Urgent unblocked tasks should come before high priority
}

func TestSuggest_IncludesUnblockingTasks(t *testing.T) {
    // Tasks that unblock others should be suggested
}
```

### Test Criteria

**Automated Tests (must pass):**
- [ ] `make test` passes with all new unit tests
- [ ] Config loading tests: defaults, from file, invalid JSON
- [ ] Blocking validation tests: self-blocking, circular deps, missing refs
- [ ] List filter tests: --ready, --is-blocked, --not-blocked
- [ ] Prime command tests: output varies by mode, respects config
- [ ] Suggest command tests: prioritization logic

**Manual Verification:**
- [ ] `egenskriven init` creates `.egenskriven/config.json`
- [ ] `egenskriven prime` outputs complete agent instructions
- [ ] `egenskriven prime` reads config from `.egenskriven/config.json`
- [ ] `egenskriven prime --mode strict` overrides config
- [ ] Prime output changes based on workflow mode (strict/light/minimal)
- [ ] `egenskriven list --ready` shows unblocked todo/backlog tasks
- [ ] `egenskriven list --is-blocked` shows blocked tasks
- [ ] `egenskriven list --not-blocked` shows unblocked tasks
- [ ] `egenskriven list --json --fields id,title` outputs only those fields
- [ ] `egenskriven update <id> --blocked-by <other>` adds blocking relationship
- [ ] `egenskriven update <id> --remove-blocked-by <other>` removes it
- [ ] `egenskriven context --json` outputs project summary
- [ ] `egenskriven suggest --json` outputs prioritized suggestions
- [ ] OpenCode plugin correctly injects prime output
- [ ] Claude hooks documentation is complete
- [ ] Circular dependency detection works
- [ ] Self-blocking is prevented

### Dependencies
- Phase 1 complete (Core CLI exists)

---

## Phase 2: Minimal UI

**Goal:** Basic web UI with board view. Single board, no advanced features.

### Tasks

#### 2.1 Initialize React Project
```bash
cd ui
npm create vite@latest . -- --template react-ts
npm install
```

**Dependencies:**
- `pocketbase` (JS SDK)
- `@dnd-kit/core`, `@dnd-kit/sortable`
- CSS: Tailwind or vanilla CSS (decide based on preference)

#### 2.2 Setup Vite Config
**File:** `ui/vite.config.ts`

- Proxy `/api` and `/_` to PocketBase
- Output to `dist/`

#### 2.3 Create embed.go
**File:** `ui/embed.go`

```go
package ui

import (
    "embed"
    "io/fs"
)

//go:embed all:dist
var distDir embed.FS

var DistFS, _ = fs.Sub(distDir, "dist")
```

#### 2.4 Update main.go for UI Serving
Serve embedded React app for non-API routes:
- `/api/*` → PocketBase API
- `/_/*` → PocketBase Admin
- `/*` → React SPA

#### 2.5 Update Makefile
```makefile
dev:
	@$(MAKE) -j2 dev-ui dev-go

dev-ui:
	cd ui && npm run dev

dev-go:
	air

build: build-ui build-go

build-ui:
	cd ui && npm ci && npm run build
```

#### 2.6 Setup Design Tokens
**File:** `ui/src/styles/tokens.css`

- All CSS custom properties from ui-design.md
- Dark mode colors (default)
- Typography scale
- Spacing scale

#### 2.7 Create PocketBase Hook
**File:** `ui/src/hooks/usePocketBase.ts`

- `useTasks()` - fetch all tasks, subscribe to changes
- `useTask(id)` - fetch single task
- CRUD operations: `createTask`, `updateTask`, `deleteTask`, `moveTask`

#### 2.8 Create Layout Components
**Files:**
- `ui/src/components/Layout.tsx` - main app shell
- `ui/src/components/Header.tsx` - top bar with title

Minimal layout without sidebar for now.

#### 2.9 Create Board Components
**Files:**
- `ui/src/components/Board.tsx` - column container
- `ui/src/components/Column.tsx` - single column with header
- `ui/src/components/TaskCard.tsx` - draggable task card

**TaskCard displays:**
- Status dot (colored by column)
- Task ID
- Title (max 2 lines)
- Priority indicator

#### 2.10 Implement Drag and Drop
Using @dnd-kit:
- Drag tasks between columns
- Update position on drop
- Visual feedback during drag

#### 2.11 Create Task Detail Panel
**File:** `ui/src/components/TaskDetail.tsx`

- Slide-in panel from right
- Display all task fields
- Editable title (inline)
- Editable properties via dropdowns
- Close with Esc or click outside

#### 2.12 Create Quick Create Modal
**File:** `ui/src/components/QuickCreate.tsx`

- Modal with title input
- Default column/type/priority dropdowns
- Create on Enter
- Cancel on Esc

#### 2.13 Basic Keyboard Navigation
- `C` - open quick create
- `Esc` - close panels/modals
- `Enter` - open selected task
- Click task to select/open

#### 2.14 Setup UI Testing

**Install Vitest:**
```bash
cd ui
npm install -D vitest @testing-library/react @testing-library/jest-dom jsdom
```

**File:** `ui/vitest.config.ts`
```typescript
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
  },
})
```

**File:** `ui/src/hooks/usePocketBase.test.ts`
```typescript
import { renderHook, waitFor } from '@testing-library/react'
import { useTasks } from './usePocketBase'

describe('useTasks', () => {
  it('fetches tasks on mount', async () => {
    // Mock PocketBase client
    // Verify tasks are loaded
  })
  
  it('updates state on realtime events', async () => {
    // Simulate create/update/delete events
  })
})
```

**File:** `ui/src/components/TaskCard.test.tsx`
```typescript
import { render, screen } from '@testing-library/react'
import { TaskCard } from './TaskCard'

describe('TaskCard', () => {
  it('renders task title', () => {
    render(<TaskCard task={mockTask} />)
    expect(screen.getByText('Test Task')).toBeInTheDocument()
  })
  
  it('shows priority indicator', () => {
    // Test urgent/high/medium/low indicators
  })
})
```

**Update package.json:**
```json
{
  "scripts": {
    "test": "vitest",
    "test:coverage": "vitest --coverage"
  }
}
```

### Test Criteria

**Automated Tests (must pass):**
- [ ] `cd ui && npm test` passes all component tests
- [ ] usePocketBase hook tests: fetch, create, update, delete, realtime
- [ ] TaskCard tests: renders title, priority, labels
- [ ] Board tests: renders columns, groups tasks correctly

**Manual Verification:**
- [ ] `make build && ./egenskriven serve` serves the UI
- [ ] Board displays columns: Backlog, Todo, In Progress, Review, Done
- [ ] Tasks created via CLI appear in UI (after refresh)
- [ ] Tasks can be dragged between columns
- [ ] Clicking a task opens detail panel
- [ ] Task properties can be edited in detail panel
- [ ] `C` key opens quick create modal
- [ ] New tasks appear in correct column
- [ ] UI uses dark mode colors from design spec

### Dependencies
- Phase 1 complete (CLI for testing data)

---

## Phase 3: Full CLI

**Goal:** Complete CLI with all features from architecture doc.

### Tasks

#### 3.1 Add Batch Input Support
Update `add` command:
- `--stdin` flag for JSON lines input
- `--file` flag for JSON file input
- Support both JSON lines and JSON array format

#### 3.2 Add Advanced Filters to `list`
- `--label` filter (repeatable)
- `--search` for title search
- `--limit` for max results
- `--sort` for sort field
- Multiple values per filter (OR within filter, AND between filters)

Note: `--ready`, `--is-blocked`, `--not-blocked`, and `--fields` filters are implemented in Phase 1.5.

#### 3.3 Add Batch Delete
Update `delete` command:
- Multiple task IDs as arguments
- `--stdin` flag for IDs from stdin

#### 3.4 Create Epic Commands
**File:** `internal/commands/epic.go`

Add `epics` collection to migrations:
- id, title, description, color

Subcommands:
- `egenskriven epic list`
- `egenskriven epic add "title" --color "#hex"`
- `egenskriven epic show <id>`
- `egenskriven epic delete <id>`

#### 3.5 Add Epic Support to Tasks
- `--epic` flag on `add` command
- `--epic` filter on `list` command
- Display epic in `show` output

#### 3.6 Improve Error Messages
- Detailed validation errors
- Suggestions for common mistakes
- Context in ambiguous reference errors

#### 3.7 Add `version` Command
Display version, build date, Go version.

### Test Criteria
- [ ] Batch add: `echo '{"title":"T1"}\n{"title":"T2"}' | egenskriven add --stdin`
- [ ] Batch delete: `egenskriven delete id1 id2 id3`
- [ ] Filter by label: `egenskriven list --label frontend`
- [ ] Search: `egenskriven list --search "login"`
- [ ] Epic CRUD works
- [ ] Tasks can be linked to epics
- [ ] `egenskriven version` displays version info

### Dependencies
- Phase 1 complete

---

## Phase 4: Interactive UI

**Goal:** Keyboard-driven UI with command palette and shortcuts.

### Tasks

#### 4.1 Implement Command Palette
**File:** `ui/src/components/CommandPalette.tsx`

- Open with `Cmd+K` / `Ctrl+K`
- Fuzzy search input
- Sections: Actions, Navigation, Recent Tasks
- Keyboard navigation with arrow keys
- Execute action on Enter

**Actions to support:**
- Create task
- Change status (when task selected)
- Set priority (when task selected)
- Go to board (placeholder for multi-board)

#### 4.2 Create Keyboard Manager
**File:** `ui/src/hooks/useKeyboard.ts`

- Global keyboard event listener
- Shortcut registration system
- Prevent shortcuts when typing in inputs
- Support key sequences (e.g., `G then B`)

#### 4.3 Implement Task Selection State
**File:** `ui/src/stores/selection.ts` (or context)

- Track selected task ID
- Track multi-selection for bulk operations
- Selection visual feedback on cards

#### 4.4 Implement All Keyboard Shortcuts

**Global:**
- `Cmd+K` - Command palette
- `/` - Focus search (placeholder)
- `Cmd+B` - Toggle board/list view (placeholder)
- `?` - Show shortcuts help

**Task actions:**
- `C` - Create task
- `Enter` - Open selected task
- `E` - Edit title (when task open)
- `Backspace` - Delete with confirmation

**Properties (when task selected/open):**
- `S` - Change status
- `P` - Change priority
- `T` - Change type
- `L` - Manage labels

**Navigation:**
- `J` / `↓` - Next task
- `K` / `↑` - Previous task
- `H` / `←` - Previous column
- `L` / `→` - Next column (note: conflicts with Labels)
- `Esc` - Close panel, deselect

**Selection:**
- `X` - Toggle select
- `Shift+X` - Select range
- `Cmd+A` - Select all visible

#### 4.5 Create Property Popovers
**Files:**
- `ui/src/components/StatusPicker.tsx`
- `ui/src/components/PriorityPicker.tsx`
- `ui/src/components/TypePicker.tsx`
- `ui/src/components/LabelPicker.tsx`

Triggered by keyboard shortcuts or clicking property in detail panel.

#### 4.6 Implement Keyboard Shortcuts Help Modal
**File:** `ui/src/components/ShortcutsHelp.tsx`

- Triggered by `?`
- Grouped by category
- Shows all available shortcuts

#### 4.7 Implement Peek Preview
- `Space` on selected task shows quick preview
- Overlay without full panel slide-in
- Press again or Esc to dismiss

#### 4.8 Add Real-time Subscriptions
Update `usePocketBase.ts`:
- Subscribe to task changes
- Handle create/update/delete events
- Update local state without refetch

### Test Criteria
- [ ] `Cmd+K` opens command palette
- [ ] Typing in palette filters actions/tasks
- [ ] `C` opens quick create from anywhere
- [ ] `J/K` navigate between tasks
- [ ] `Enter` opens task detail
- [ ] `S` opens status picker on selected task
- [ ] `?` shows shortcuts help
- [ ] `Space` shows peek preview
- [ ] Real-time: CLI changes reflect in UI immediately

### Dependencies
- Phase 2 complete

---

## Phase 5: Multi-Board Support

**Goal:** Support multiple boards, global access from CLI.

### Tasks

#### 5.1 Create Boards Collection
**Migration update:**

`boards` collection:
- id (string)
- name (string, required)
- prefix (string, required, unique) - e.g., "WRK"
- columns (json) - array of column definitions
- color (string) - accent color for board

Update `tasks` collection:
- Add `board` relation field

#### 5.2 Update CLI for Multi-Board

**Global flag:**
- `--board` / `-b` - specify board by name or prefix

**New commands:**
- `egenskriven board list`
- `egenskriven board add "name" --prefix WRK`
- `egenskriven board show <name>`
- `egenskriven board delete <name>`
- `egenskriven board use <name>` - set default board

**Default board:**
- Store in `~/.egenskriven/config.json`
- Or use first board if none set

#### 5.3 Update Task ID Format
Task IDs now include board prefix:
- Display: `WRK-123`
- Storage: still auto-generated ID, prefix is display only

Or alternative: auto-increment per board.

#### 5.4 Update Task Commands
- `add` respects `--board` flag
- `list` filters by board (default: current board)
- `list --all-boards` shows all
- Resolver scopes to board unless cross-board reference

#### 5.5 Create Sidebar Component
**File:** `ui/src/components/Sidebar.tsx`

- Board list with active indicator
- Click to switch boards
- "New board" button
- Collapsible with `Cmd+\`

#### 5.6 Create Board Switcher
- Quick switch via command palette
- Keyboard shortcut `G then B`
- Remember last used board

#### 5.7 Update Board Component
- Load tasks filtered by current board
- Board-specific column configuration
- Board accent color applied

#### 5.8 Create Board Settings
**File:** `ui/src/components/BoardSettings.tsx`

- Edit board name
- Change prefix
- Configure columns (add/remove/rename)
- Set accent color
- Delete board

### Test Criteria
- [ ] `egenskriven board add "Work" --prefix WRK` creates board
- [ ] `egenskriven add "task" --board Work` creates task in Work board
- [ ] `egenskriven list` shows only current board's tasks
- [ ] `egenskriven list --all-boards` shows all tasks
- [ ] UI sidebar shows all boards
- [ ] Clicking board in sidebar switches view
- [ ] Tasks display board-prefixed IDs (WRK-123)
- [ ] Board settings allow column customization

### Dependencies
- Phase 2 complete (UI exists)
- Phase 3 complete (full CLI)

---

## Phase 6: Filtering & Views

**Goal:** Advanced filtering, search, and saved views.

### Tasks

#### 6.1 Create Filter State Management
**File:** `ui/src/stores/filters.ts`

- Active filters array
- Filter operators (is, is not, includes any, etc.)
- AND/OR mode toggle

#### 6.2 Create Filter UI
**File:** `ui/src/components/FilterBar.tsx`

- Filter button in header
- Active filter pills
- Clear all button

**File:** `ui/src/components/FilterBuilder.tsx`

- Property selector dropdown
- Operator dropdown
- Value selector (depends on property type)
- Add filter button
- AND/OR toggle

#### 6.3 Implement Filter Logic
Filter properties:
- Status (column)
- Priority
- Type
- Labels (includes any/all/none)
- Due date (before/after/is set)
- Title search

Apply filters to task list before rendering.

#### 6.4 Create Search Component
**File:** `ui/src/components/Search.tsx`

- Triggered by `/` key
- Search input in header area
- Real-time filtering as you type
- Search title and description

#### 6.5 Create Views Collection
**Migration:**

`views` collection:
- id
- name (string)
- board (relation)
- filters (json) - saved filter state
- display (json) - view preferences (board/list, visible fields)
- is_favorite (boolean)

#### 6.6 Implement Save View
- "Save as view" in filter bar
- Name input modal
- Saved view appears in sidebar

#### 6.7 Create Views Sidebar Section
**Update:** `ui/src/components/Sidebar.tsx`

- "Views" section below boards
- List saved views
- Star/unstar to favorite
- Favorites appear at top

#### 6.8 Implement View Loading
- Click view to apply filters
- View name shown in header
- "Modified" indicator if filters changed

#### 6.9 Add List View
**File:** `ui/src/components/ListView.tsx`

- Table/row layout
- Columns: Status, ID, Title, Labels, Priority, Due
- Sortable columns
- Toggle with `Cmd+B`

#### 6.10 Add Display Options
**File:** `ui/src/components/DisplayOptions.tsx`

- Toggle visible card properties
- Grouping options (by status, priority, type)
- Density setting (compact/comfortable)

### Test Criteria
- [ ] `F` opens filter builder
- [ ] Can filter by status, priority, type, labels
- [ ] Multiple filters combine with AND/OR
- [ ] `/` opens search, filters in real-time
- [ ] Can save current filters as named view
- [ ] Saved views appear in sidebar
- [ ] Clicking view applies its filters
- [ ] `Cmd+B` toggles between board and list view
- [ ] Display options persist per view

### Dependencies
- Phase 4 complete (keyboard shortcuts)
- Phase 5 complete (multi-board for views per board)

---

## Phase 7: Polish

**Goal:** Theming, animations, responsive design, quality of life.

### Tasks

#### 7.1 Implement Light Mode
**File:** `ui/src/styles/light.css`

- All light mode color tokens
- System preference detection
- Manual toggle in settings

#### 7.2 Create Settings Panel
**File:** `ui/src/components/Settings.tsx`

- Appearance section: Theme (System/Light/Dark)
- Accent color picker
- Display density
- Keyboard shortcut customization (stretch)

#### 7.3 Implement Accent Colors
- 8 accent color options
- Apply to selected states, buttons, focus rings
- Store preference in localStorage/config

#### 7.4 Add Animations
Implement transitions from ui-design.md:
- Panel slide-in (150ms ease-out)
- Modal appear (150ms fade + scale)
- Hover states (100ms)
- Dropdown open (100ms fade + slide)
- Drag feedback (immediate lift)

Use CSS transitions and/or Framer Motion.

#### 7.5 Implement Responsive Layout
**Breakpoints:**
- Mobile (<640px): Sidebar drawer, single column, bottom sheet modals
- Tablet (640-1024px): Collapsible sidebar, 2-3 columns
- Desktop (>1024px): Full layout

**Mobile adaptations:**
- Touch-friendly tap targets (44px min)
- Swipe gestures for drawer
- Bottom sheet for command palette

#### 7.6 Add Loading States
- Skeleton loaders for initial load
- Optimistic updates for fast feedback
- Error states with retry

#### 7.7 Add Toast Notifications
**File:** `ui/src/components/Toast.tsx`

- Success/error/info toasts
- Auto-dismiss after 3s
- Action button option
- Stack multiple toasts

#### 7.8 Improve Drag and Drop
- Better drop indicators
- Auto-scroll when dragging near edges
- Touch support for mobile

#### 7.9 Add Keyboard Navigation Visual Feedback
- Visible focus indicators
- Selected task highlight
- Current column indicator when navigating

#### 7.10 Performance Optimization
- Virtualize long task lists
- Memoize expensive computations
- Lazy load detail panel content

### Test Criteria
- [ ] Light/Dark/System theme toggle works
- [ ] Accent color changes apply immediately
- [ ] Animations are smooth at 60fps
- [ ] Mobile layout works on narrow screens
- [ ] Sidebar becomes drawer on mobile
- [ ] Toast notifications appear for actions
- [ ] Loading states show during data fetch
- [ ] Keyboard focus is always visible
- [ ] Large boards (100+ tasks) remain performant

### Dependencies
- Phase 6 complete (all features exist to polish)

---

## Phase 8: Advanced Features

**Goal:** Epics, due dates, sub-tasks, and other advanced features.

### Tasks

#### 8.1 Implement Epic UI
**Files:**
- `ui/src/components/EpicList.tsx`
- `ui/src/components/EpicDetail.tsx`
- `ui/src/components/EpicPicker.tsx`

- Epic section in sidebar
- Epic detail view showing linked tasks
- Epic picker in task detail panel
- Epic badge on task cards (optional display)

#### 8.2 Add Due Dates
**Migration:** Add `due_date` field to tasks

**CLI:**
- `--due` flag on add/update
- `--due-before`, `--due-after` filters

**UI:**
- Date picker component
- Due date in task card
- Due date filter
- Overdue visual indicator

#### 8.3 Implement Sub-tasks
**Migration:** Add `parent` relation field to tasks

**CLI:**
- `egenskriven add "subtask" --parent <task>`
- Sub-tasks listed under parent in `show`

**UI:**
- Sub-task list in detail panel
- Add sub-task button
- Progress indicator on parent
- Expand/collapse sub-tasks in board view

#### 8.4 Add Description Markdown Editor
**File:** `ui/src/components/MarkdownEditor.tsx`

- Rich text editing for descriptions
- Markdown preview
- Basic formatting toolbar
- Keyboard shortcuts (bold, italic, etc.)

#### 8.5 Create Timeline View (Stretch)
**File:** `ui/src/components/Timeline.tsx`

- Gantt-style view for tasks with due dates
- Zoom levels: Week, Month, Quarter
- Drag to adjust dates

#### 8.6 Add Activity Log
Track and display:
- Task created
- Status changed
- Priority changed
- Moved between columns
- Description edited

Show in task detail panel.

#### 8.7 Implement Import/Export
**CLI:**
- `egenskriven export --format json > backup.json`
- `egenskriven export --format csv > tasks.csv`
- `egenskriven import backup.json`

**UI:**
- Export button in settings
- Import via file picker

#### 8.8 Add Task Templates (Stretch)
**CLI:**
- `egenskriven template add bug-report --type bug --labels bug`
- `egenskriven add "title" --template bug-report`

**UI:**
- Template picker in quick create
- Template management in settings

### Test Criteria
- [ ] Epics display in sidebar
- [ ] Tasks can be linked to epics via UI
- [ ] Due dates can be set and display on cards
- [ ] Overdue tasks show visual indicator
- [ ] Sub-tasks nest under parent tasks
- [ ] Parent shows sub-task progress
- [ ] Markdown editor works for descriptions
- [ ] Export produces valid JSON/CSV
- [ ] Import restores data correctly

### Dependencies
- Phase 7 complete (polished base)
- Phase 3 complete (epic CLI commands)

---

## Phase 9: Release

**Goal:** Production-ready release with distribution.

### Tasks

#### 9.1 Cross-Platform Build
Update Makefile for all platforms (5 builds):
```makefile
release: build-ui
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o dist/egenskriven-darwin-arm64 ./cmd/egenskriven
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o dist/egenskriven-darwin-amd64 ./cmd/egenskriven
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/egenskriven-linux-amd64 ./cmd/egenskriven
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o dist/egenskriven-linux-arm64 ./cmd/egenskriven
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o dist/egenskriven-windows-amd64.exe ./cmd/egenskriven
```

**Platforms:**
- macOS (Apple Silicon / ARM64)
- macOS (Intel / AMD64)
- Linux (AMD64)
- Linux (ARM64 - for Raspberry Pi, ARM servers)
- Windows (AMD64)

#### 9.2 Add Version Embedding
```go
var (
    Version   = "dev"
    BuildDate = "unknown"
)
```

Build with:
```bash
go build -ldflags "-X main.Version=1.0.0 -X main.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

#### 9.3 Create GitHub Release Workflow
`.github/workflows/release.yml`:
- Trigger on tag push
- Build all platforms
- Create GitHub release
- Upload binaries

#### 9.4 Shell Completions
Implement shell completion generation:
- `egenskriven completion bash`
- `egenskriven completion zsh`
- `egenskriven completion fish`

#### 9.5 Create Installation Script
```bash
curl -fsSL https://egenskriven.app/install.sh | sh
```

Or via package managers if feasible.

#### 9.6 Write Documentation
- README.md with quick start
- CLI reference (auto-generated from Cobra)
- UI guide with screenshots
- Configuration documentation

#### 9.7 Create Landing Page (Optional)
Simple site at egenskriven.app:
- Feature overview
- Download links
- Documentation links

#### 9.8 Final Testing
- Test on macOS (Intel + ARM)
- Test on Linux (Ubuntu, Arch)
- Test on Windows
- Test fresh install experience
- Test upgrade path

#### 9.9 Create Changelog
Document all features for v1.0.0.

### Test Criteria
- [ ] Binaries build for all 5 platform/arch combinations
- [ ] Each binary is a single file, no external dependencies
- [ ] `./egenskriven version` shows correct version
- [ ] Fresh install works: download, run `serve`, use CLI and UI
- [ ] GitHub release workflow succeeds
- [ ] Shell completions work
- [ ] Documentation is complete and accurate

### Dependencies
- All previous phases complete

---

## Post-V1: Future Phases

### Phase 10: Custom Themes
- Theme JSON file format
- Theme loader from `~/.egenskriven/themes/`
- Pre-packaged themes: Catppuccin, Dracula, Nord, Gruvbox, One Dark, Solarized
- Theme preview in settings
- Community theme sharing

### Phase 11: Git Integration
- `egenskriven git link <task> <branch>`
- Auto-detect task ID from branch name
- Show linked branches in task detail
- Create task from commit message pattern

### Phase 12: TUI Mode
- `egenskriven board` opens terminal UI
- Bubble Tea implementation
- Full keyboard navigation
- Works over SSH

### Phase 13: Sync & Collaboration (Major)
- Optional cloud sync
- Share boards with others
- Real-time collaboration
- Requires auth system

---

## Dependency Graph

```
Phase 0 (Setup)
    │
    ▼
Phase 1 (Core CLI)
    │
    ├───────────────────────────────────┐
    ▼                                   ▼
Phase 1.5 (Agent Integration)     Phase 2 (Minimal UI)
    │                                   │
    │                                   ├───────────────┐
    │                                   ▼               ▼
    │                               Phase 4         Phase 3 (Full CLI)
    │                               (Interactive)       │
    │                                   │               │
    └───────────────────────────────────┼───────────────┘
                                        ▼
                                Phase 5 (Multi-Board)
                                        │
                                        ▼
                                Phase 6 (Filtering & Views)
                                        │
                                        ▼
                                Phase 7 (Polish)
                                        │
                                        ▼
                                Phase 8 (Advanced)
                                        │
                                        ▼
                                Phase 9 (Release)
```

**Parallel work opportunities:**
- Phase 1.5 (Agent Integration) can run parallel to Phase 2 (Minimal UI)
- Phase 3 (Full CLI) can run parallel to Phase 2 (Minimal UI)
- Phase 4 (Interactive UI) requires Phase 2
- Phase 5 (Multi-Board) requires Phase 3, Phase 4, and benefits from Phase 1.5

---

## Testing Strategy

### Philosophy

Tests are written **alongside implementation**, not after. Each phase includes specific test tasks and automated tests must pass before the phase is considered complete.

### Test Pyramid

```
        /\
       /  \      E2E Tests (Phase 2+)
      /----\     - Playwright for UI
     /      \    - Full user flows
    /--------\   
   /          \  Integration Tests (Phase 1+)
  /------------\ - CLI command flows
 /              \- Database operations
/----------------\
|  Unit Tests    | - Resolver logic
|  (Phase 1+)    | - Position calculation
|                | - Filter logic
|                | - Output formatting
------------------
```

### Tools

| Tool | Purpose | Phase Introduced |
|------|---------|------------------|
| Go `testing` | Unit & integration tests | Phase 0 |
| `testify` | Assertions, mocks | Phase 0 |
| `testutil.NewTestApp()` | Isolated test database | Phase 0 |
| Vitest | React component tests | Phase 2 |
| Playwright | E2E browser tests | Phase 4 (optional) |

### Test Requirements Per Phase

| Phase | Requirement |
|-------|-------------|
| 0 | Test infrastructure set up, `make test` works |
| 1 | Unit tests for resolver, output, position; integration tests for commands |
| 1.5 | Unit tests for config, blocking, prime, suggest |
| 2 | React component tests with Vitest |
| 3+ | Tests for new features before merging |

### Running Tests

```bash
# Run all Go tests
make test

# Run with coverage report
make test-coverage

# Run specific package
go test ./internal/resolver/... -v

# Run UI tests (Phase 2+)
cd ui && npm test
```

### Coverage Goals

- **Phase 1-1.5**: 80%+ coverage on `internal/resolver`, `internal/output`, `internal/config`
- **Phase 2+**: Component tests for critical UI paths
- Focus on logic-heavy code, not boilerplate

### Manual Testing Checklist

Each phase includes specific test criteria that must pass before moving to next phase. These are split into:
- **Automated Tests**: Must pass in CI/locally via `make test`
- **Manual Verification**: Human testing for UX and edge cases

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| PocketBase API changes | Pin to specific version, test upgrades |
| Large binary size | Monitor bundle size, lazy load UI chunks |
| SQLite performance at scale | Test with 1000+ tasks, add indexes if needed |
| Complex drag-drop edge cases | Extensive manual testing, fallback behaviors |
| Cross-platform issues | CI testing on all platforms |

---

*This plan provides a structured path from zero to a full-featured product. Each phase builds on previous work and ends with a testable deliverable. Adjust timeline estimates based on your velocity after completing Phase 0 and 1.*
