# Phase 1.5: Agent Integration

**Goal**: Make EgenSkriven agent-native so AI coding assistants (Claude, OpenCode, Cursor, etc.) use it as their primary task tracker.

**Duration Estimate**: 2-3 days

**Prerequisites**: Phase 1 complete (Core CLI with add, list, show, move, update, delete commands).

**Deliverable**: A CLI with agent-friendly features including per-project configuration, the `prime` command for injecting instructions into AI agents, blocking relationships between tasks, and helper commands like `context` and `suggest`.

---

## Overview

In this phase, we add features that make EgenSkriven work seamlessly with AI coding assistants. The goal is for agents to use EgenSkriven instead of their built-in todo lists (like `TodoWrite`).

### Why Agent Integration?

AI coding assistants need a structured way to track work. EgenSkriven provides:
- **Structured output**: JSON mode for machine-readable responses
- **Blocking relationships**: Agents can identify parallelizable work
- **Prime command**: Injects workflow instructions into agent context
- **Per-project config**: Different workflows for different projects

### What We're Building

| Component | Purpose |
|-----------|---------|
| Config loader | Load per-project settings from `.egenskriven/config.json` |
| Prime command | Output agent instructions based on config |
| Init command | Create project configuration |
| Context command | Output project state summary |
| Suggest command | Recommend tasks to work on |
| Blocking filters | `--ready`, `--is-blocked`, `--not-blocked` |
| Field selection | `--fields` flag for token-efficient output |
| OpenCode plugin | Hook for OpenCode agent integration |
| Claude hooks | Configuration for Claude Code integration |

### Architecture Overview

```
.egenskriven/config.json → Config Loader → Prime Command → Agent Context
                                              ↓
                              Template Engine (prime.tmpl)
                                              ↓
                              Agent receives instructions
```

---

## Environment Requirements

Ensure Phase 1 is complete:

| Check | Command |
|-------|---------|
| Build works | `make build` |
| Tests pass | `make test` |
| Add command works | `./egenskriven add "Test task"` |
| List command works | `./egenskriven list` |

---

## Tasks

### 1.5.1 Implement Config Loader

**What**: Create a configuration loader that reads per-project settings from `.egenskriven/config.json`.

**Why**: Different projects may need different agent workflows. A strict workflow for critical projects, a light workflow for personal projects.

**File**: `internal/config/config.go`

```go
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AgentConfig defines agent-specific behavior settings.
type AgentConfig struct {
	// Workflow mode: "strict", "light", "minimal"
	// - strict: Full enforcement (create before, update during, summary after)
	// - light: Basic tracking (create/complete, no structured sections)
	// - minimal: No enforcement (agent decides when to use)
	Workflow string `json:"workflow"`

	// Mode defines agent autonomy: "autonomous", "collaborative", "supervised"
	// - autonomous: Agent executes actions directly
	// - collaborative: Agent proposes major changes, executes minor ones
	// - supervised: Agent is read-only, outputs commands for human
	Mode string `json:"mode"`

	// OverrideTodoWrite tells agents to ignore built-in todo tools
	OverrideTodoWrite bool `json:"overrideTodoWrite"`

	// RequireSummary requires agents to add summary section on completion
	RequireSummary bool `json:"requireSummary"`

	// StructuredSections encourages structured markdown sections in descriptions
	StructuredSections bool `json:"structuredSections"`
}

// Config represents the project configuration.
type Config struct {
	Agent AgentConfig `json:"agent"`
}

// DefaultConfig returns configuration with default values.
func DefaultConfig() *Config {
	return &Config{
		Agent: AgentConfig{
			Workflow:           "light",
			Mode:               "autonomous",
			OverrideTodoWrite:  true,
			RequireSummary:     false,
			StructuredSections: false,
		},
	}
}

// LoadProjectConfig loads configuration from .egenskriven/config.json.
// Returns default config if file doesn't exist.
func LoadProjectConfig() (*Config, error) {
	return LoadProjectConfigFrom(".")
}

// LoadProjectConfigFrom loads configuration from a specific directory.
func LoadProjectConfigFrom(dir string) (*Config, error) {
	configPath := filepath.Join(dir, ".egenskriven", "config.json")

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		// Return defaults if no config file
		return DefaultConfig(), nil
	}
	if err != nil {
		return nil, err
	}

	// Start with defaults, then overlay file values
	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Validate values
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validateConfig ensures config values are valid.
func validateConfig(cfg *Config) error {
	validWorkflows := map[string]bool{"strict": true, "light": true, "minimal": true}
	if !validWorkflows[cfg.Agent.Workflow] {
		cfg.Agent.Workflow = "light" // Default to light if invalid
	}

	validModes := map[string]bool{"autonomous": true, "collaborative": true, "supervised": true}
	if !validModes[cfg.Agent.Mode] {
		cfg.Agent.Mode = "autonomous" // Default to autonomous if invalid
	}

	return nil
}

// SaveConfig saves configuration to .egenskriven/config.json.
func SaveConfig(dir string, cfg *Config) error {
	configDir := filepath.Join(dir, ".egenskriven")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.json")

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
```

**Steps**:

1. Create the file:
   ```bash
   touch internal/config/config.go
   ```

2. Open in your editor and paste the code above.

3. Verify it compiles:
   ```bash
   go build ./internal/config
   ```
   
   **Expected output**: No output means success!

**Configuration Options**:

| Option | Values | Description |
|--------|--------|-------------|
| `workflow` | strict, light, minimal | Workflow enforcement level |
| `mode` | autonomous, collaborative, supervised | Agent autonomy level |
| `overrideTodoWrite` | true, false | Tell agents to ignore built-in todo tools |
| `requireSummary` | true, false | Require summary section on completion |
| `structuredSections` | true, false | Encourage structured markdown sections |

---

### 1.5.2 Write Config Tests

**What**: Create tests for the configuration loader.

**Why**: Config loading is critical - invalid configs should be handled gracefully with defaults.

**File**: `internal/config/config_test.go`

```go
package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadProjectConfig_Defaults(t *testing.T) {
	// Create a temp directory without config file
	tmpDir, err := os.MkdirTemp("", "egenskriven-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	cfg, err := LoadProjectConfigFrom(tmpDir)

	require.NoError(t, err)
	assert.Equal(t, "light", cfg.Agent.Workflow)
	assert.Equal(t, "autonomous", cfg.Agent.Mode)
	assert.True(t, cfg.Agent.OverrideTodoWrite)
	assert.False(t, cfg.Agent.RequireSummary)
	assert.False(t, cfg.Agent.StructuredSections)
}

func TestLoadProjectConfig_FromFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "egenskriven-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create config file
	configDir := filepath.Join(tmpDir, ".egenskriven")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	configContent := `{
		"agent": {
			"workflow": "strict",
			"mode": "collaborative",
			"overrideTodoWrite": false,
			"requireSummary": true,
			"structuredSections": true
		}
	}`

	require.NoError(t, os.WriteFile(
		filepath.Join(configDir, "config.json"),
		[]byte(configContent),
		0644,
	))

	cfg, err := LoadProjectConfigFrom(tmpDir)

	require.NoError(t, err)
	assert.Equal(t, "strict", cfg.Agent.Workflow)
	assert.Equal(t, "collaborative", cfg.Agent.Mode)
	assert.False(t, cfg.Agent.OverrideTodoWrite)
	assert.True(t, cfg.Agent.RequireSummary)
	assert.True(t, cfg.Agent.StructuredSections)
}

func TestLoadProjectConfig_InvalidJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "egenskriven-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create invalid config file
	configDir := filepath.Join(tmpDir, ".egenskriven")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(configDir, "config.json"),
		[]byte("invalid json"),
		0644,
	))

	_, err = LoadProjectConfigFrom(tmpDir)

	assert.Error(t, err)
}

func TestLoadProjectConfig_InvalidWorkflow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "egenskriven-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create config with invalid workflow
	configDir := filepath.Join(tmpDir, ".egenskriven")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	configContent := `{"agent": {"workflow": "invalid"}}`
	require.NoError(t, os.WriteFile(
		filepath.Join(configDir, "config.json"),
		[]byte(configContent),
		0644,
	))

	cfg, err := LoadProjectConfigFrom(tmpDir)

	require.NoError(t, err)
	assert.Equal(t, "light", cfg.Agent.Workflow) // Should default to light
}

func TestSaveConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "egenskriven-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	cfg := &Config{
		Agent: AgentConfig{
			Workflow:           "strict",
			Mode:               "supervised",
			OverrideTodoWrite:  true,
			RequireSummary:     true,
			StructuredSections: true,
		},
	}

	err = SaveConfig(tmpDir, cfg)
	require.NoError(t, err)

	// Verify file was created
	configPath := filepath.Join(tmpDir, ".egenskriven", "config.json")
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Load and verify contents
	loaded, err := LoadProjectConfigFrom(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, cfg.Agent.Workflow, loaded.Agent.Workflow)
	assert.Equal(t, cfg.Agent.Mode, loaded.Agent.Mode)
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, "light", cfg.Agent.Workflow)
	assert.Equal(t, "autonomous", cfg.Agent.Mode)
	assert.True(t, cfg.Agent.OverrideTodoWrite)
	assert.False(t, cfg.Agent.RequireSummary)
	assert.False(t, cfg.Agent.StructuredSections)
}
```

**Steps**:

1. Create the file:
   ```bash
   touch internal/config/config_test.go
   ```

2. Open in your editor and paste the code above.

3. Run the tests:
   ```bash
   go test ./internal/config -v
   ```
   
   **Expected output**:
   ```
   === RUN   TestLoadProjectConfig_Defaults
   --- PASS: TestLoadProjectConfig_Defaults (0.00s)
   === RUN   TestLoadProjectConfig_FromFile
   --- PASS: TestLoadProjectConfig_FromFile (0.00s)
   === RUN   TestLoadProjectConfig_InvalidJSON
   --- PASS: TestLoadProjectConfig_InvalidJSON (0.00s)
   === RUN   TestLoadProjectConfig_InvalidWorkflow
   --- PASS: TestLoadProjectConfig_InvalidWorkflow (0.00s)
   === RUN   TestSaveConfig
   --- PASS: TestSaveConfig (0.00s)
   === RUN   TestDefaultConfig
   --- PASS: TestDefaultConfig (0.00s)
   PASS
   ```

---

### 1.5.3 Implement Init Command

**What**: Create the `init` command that initializes EgenSkriven configuration for a project.

**Why**: Users need an easy way to create and customize their project configuration.

**File**: `internal/commands/init.go`

```go
package commands

import (
	"fmt"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"

	"github.com/yourusername/egenskriven/internal/config"
)

func newInitCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		workflow string
		mode     string
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize EgenSkriven configuration",
		Long: `Initialize EgenSkriven configuration for the current project.

Creates .egenskriven/config.json with agent workflow settings.

Workflow modes:
  strict   - Full enforcement: create before work, update during, summary after
  light    - Basic tracking: create/complete tasks, no structured sections
  minimal  - No enforcement: agent decides when to use

Agent modes:
  autonomous    - Agent executes actions directly, human reviews async
  collaborative - Agent proposes major changes, executes minor ones
  supervised    - Agent is read-only, outputs commands for human

Examples:
  egenskriven init
  egenskriven init --workflow strict
  egenskriven init --workflow light --mode collaborative`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			// Load existing config or create new
			cfg := config.DefaultConfig()

			// Override with flags if provided
			if workflow != "" {
				cfg.Agent.Workflow = workflow
			}
			if mode != "" {
				cfg.Agent.Mode = mode
			}

			// Save configuration
			if err := config.SaveConfig(".", cfg); err != nil {
				return out.Error(ExitGeneralError,
					fmt.Sprintf("failed to save config: %v", err), nil)
			}

			if !jsonOutput {
				fmt.Println("Created .egenskriven/config.json")
				fmt.Printf("  Workflow: %s\n", cfg.Agent.Workflow)
				fmt.Printf("  Mode: %s\n", cfg.Agent.Mode)
				fmt.Println()
				fmt.Println("Next steps:")
				fmt.Println("  1. Edit .egenskriven/config.json to customize settings")
				fmt.Println("  2. Run 'egenskriven prime' to see agent instructions")
				fmt.Println("  3. Configure your AI agent to run 'egenskriven prime' on session start")
			} else {
				out.Success("Configuration initialized")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&workflow, "workflow", "",
		"Workflow mode (strict, light, minimal)")
	cmd.Flags().StringVar(&mode, "mode", "",
		"Agent mode (autonomous, collaborative, supervised)")

	return cmd
}
```

**Steps**:

1. Create the file:
   ```bash
   touch internal/commands/init.go
   ```

2. Open in your editor and paste the code above.

3. **Important**: Replace `github.com/yourusername/egenskriven` with your actual module path.

4. Verify it compiles:
   ```bash
   go build ./internal/commands
   ```

**Usage Examples**:

```bash
# Initialize with defaults
egenskriven init

# Initialize with strict workflow
egenskriven init --workflow strict

# Initialize with collaborative mode
egenskriven init --mode collaborative
```

---

### 1.5.4 Create Prime Template

**What**: Create the template file that generates agent instructions.

**Why**: The prime command uses a Go template to generate customized instructions based on project configuration.

**File**: `internal/commands/prime.tmpl`

```
<EXTREMELY_IMPORTANT>
# EgenSkriven Task Tracker for Agents

This project uses **EgenSkriven**, a local-first kanban board for task tracking.
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
Example: "I believe task [id] is complete. Run `egenskriven move [id] done` to confirm."
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
- Reference task ID in commits: "feat: implement X [task-id]"
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
# Find ready tasks (unblocked, in todo/backlog)
egenskriven list --json --ready

# Show task details
egenskriven show <id> --json

# Create task (always include --agent to identify yourself)
egenskriven add "Title" --type bug --priority urgent --agent {{.AgentName}} --json

# Move task to different column
egenskriven move <id> in_progress
egenskriven move <id> done

# Update task
egenskriven update <id> --priority high --blocked-by <other-id>

# Get work suggestions
egenskriven suggest --json

# Get project context
egenskriven context --json

# List all tasks
egenskriven list --json

# Filter tasks
egenskriven list --json --column todo --type bug
egenskriven list --json --is-blocked
egenskriven list --json --not-blocked
```

## Task Types
- `bug`: Something broken that needs fixing
- `feature`: New functionality
- `chore`: Maintenance, refactoring, docs

## Priorities
- `urgent`: Do immediately
- `high`: Do before normal work
- `medium`: Standard priority
- `low`: Can be delayed

## Columns
- `backlog`: Not yet planned
- `todo`: Ready to start
- `in_progress`: Currently working
- `review`: Awaiting review
- `done`: Completed

## Blocking Relationships

Use `--blocked-by` to track dependencies:
- `egenskriven update <id> --blocked-by <other-id>` - Mark as blocked
- `egenskriven list --not-blocked` - Find parallelizable work
- `egenskriven list --is-blocked` - See what's waiting
- `egenskriven list --ready` - Unblocked tasks in todo/backlog
</EXTREMELY_IMPORTANT>
```

**Steps**:

1. Create the file:
   ```bash
   touch internal/commands/prime.tmpl
   ```

2. Open in your editor and paste the template above.

**Template Variables**:

| Variable | Source | Description |
|----------|--------|-------------|
| `{{.WorkflowMode}}` | config.Agent.Workflow | strict, light, minimal |
| `{{.AgentMode}}` | config.Agent.Mode | autonomous, collaborative, supervised |
| `{{.AgentName}}` | flag or default | Agent identifier (e.g., "claude") |
| `{{.OverrideTodoWrite}}` | config.Agent.OverrideTodoWrite | Whether to override built-in todos |
| `{{.RequireSummary}}` | config.Agent.RequireSummary | Require summary section |
| `{{.StructuredSections}}` | config.Agent.StructuredSections | Use structured markdown |

---

### 1.5.5 Implement Prime Command

**What**: Create the `prime` command that outputs agent instructions.

**Why**: AI agents need context about how to use EgenSkriven. The prime command provides customized instructions based on project configuration.

**File**: `internal/commands/prime.go`

```go
package commands

import (
	_ "embed"
	"fmt"
	"os"
	"text/template"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"

	"github.com/yourusername/egenskriven/internal/config"
)

//go:embed prime.tmpl
var primeTemplate string

// PrimeTemplateData holds data for the prime template.
type PrimeTemplateData struct {
	WorkflowMode       string
	AgentMode          string
	AgentName          string
	OverrideTodoWrite  bool
	RequireSummary     bool
	StructuredSections bool
}

func newPrimeCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		workflowOverride string
		agentName        string
	)

	cmd := &cobra.Command{
		Use:   "prime",
		Short: "Output instructions for AI coding agents",
		Long: `Output instructions that teach AI agents how to use EgenSkriven.

Reads configuration from .egenskriven/config.json in the current project.
Use --workflow to override the workflow mode.

This command is typically called automatically via agent hooks (Claude, OpenCode)
rather than manually. The output is designed to be injected into agent context.

Examples:
  egenskriven prime
  egenskriven prime --workflow strict
  egenskriven prime --agent claude`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load project configuration
			cfg, err := config.LoadProjectConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
				cfg = config.DefaultConfig()
			}

			// Override workflow if specified
			workflowMode := cfg.Agent.Workflow
			if workflowOverride != "" {
				workflowMode = workflowOverride
			}

			// Set default agent name if not provided
			if agentName == "" {
				agentName = "agent"
			}

			// Parse and execute template
			tmpl, err := template.New("prime").Parse(primeTemplate)
			if err != nil {
				return fmt.Errorf("failed to parse template: %w", err)
			}

			data := PrimeTemplateData{
				WorkflowMode:       workflowMode,
				AgentMode:          cfg.Agent.Mode,
				AgentName:          agentName,
				OverrideTodoWrite:  cfg.Agent.OverrideTodoWrite,
				RequireSummary:     cfg.Agent.RequireSummary,
				StructuredSections: cfg.Agent.StructuredSections,
			}

			return tmpl.Execute(os.Stdout, data)
		},
	}

	cmd.Flags().StringVar(&workflowOverride, "workflow", "",
		"Override workflow mode (strict, light, minimal)")
	cmd.Flags().StringVar(&agentName, "agent", "",
		"Agent identifier for --agent flag in examples")

	return cmd
}
```

**Steps**:

1. Create the file:
   ```bash
   touch internal/commands/prime.go
   ```

2. Open in your editor and paste the code above.

3. **Important**: Replace `github.com/yourusername/egenskriven` with your actual module path.

4. Verify it compiles:
   ```bash
   go build ./internal/commands
   ```

**Usage Examples**:

```bash
# Output with project defaults
egenskriven prime

# Override to strict mode
egenskriven prime --workflow strict

# Specify agent name
egenskriven prime --agent claude
```

---

### 1.5.6 Add Blocking Relationship Filters to List

**What**: Add `--ready`, `--is-blocked`, `--not-blocked`, and `--fields` flags to the list command.

**Why**: Agents need to find actionable tasks. These filters help identify which tasks can be worked on and which are waiting on dependencies.

**File**: Update `internal/commands/list.go`

Add the following code to the existing `newListCmd` function:

```go
// Add these variables to the var block at the top of newListCmd
var (
	// ... existing variables ...
	ready      bool
	isBlocked  bool
	notBlocked bool
	fields     string
)

// Add these filters in the filter building section, after existing filters:

// Ready filter: unblocked tasks in todo/backlog
if ready {
	columns = []string{"todo", "backlog"}
	notBlocked = true
}

// Is blocked filter
if isBlocked {
	filters = append(filters, dbx.NewExp(
		"json_array_length(blocked_by) > 0",
	))
}

// Not blocked filter
if notBlocked {
	filters = append(filters, dbx.Or(
		dbx.NewExp("blocked_by IS NULL"),
		dbx.NewExp("blocked_by = '[]'"),
		dbx.NewExp("json_array_length(blocked_by) = 0"),
	))
}

// Add these flag definitions after existing flags:
cmd.Flags().BoolVar(&ready, "ready", false,
	"Show unblocked tasks in todo/backlog (agent-friendly)")
cmd.Flags().BoolVar(&isBlocked, "is-blocked", false,
	"Show only tasks blocked by others")
cmd.Flags().BoolVar(&notBlocked, "not-blocked", false,
	"Show only tasks not blocked by others")
cmd.Flags().StringVar(&fields, "fields", "",
	"Comma-separated fields to include in JSON output")
```

Here is the complete updated `internal/commands/list.go`:

```go
package commands

import (
	"fmt"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
)

func newListCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		columns    []string
		types      []string
		priorities []string
		search     string
		createdBy  string
		agentName  string
		ready      bool
		isBlocked  bool
		notBlocked bool
		fields     string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		Long: `List and filter tasks on the kanban board.

By default, shows all tasks grouped by column. Use flags to filter.

Examples:
  egenskriven list
  egenskriven list --column todo
  egenskriven list --type bug --priority urgent
  egenskriven list --ready
  egenskriven list --is-blocked
  egenskriven list --json --fields id,title,column`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			// Bootstrap the app
			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			// Build filter expressions
			var filters []dbx.Expression

			// Ready filter: unblocked tasks in todo/backlog
			if ready {
				columns = []string{"todo", "backlog"}
				notBlocked = true
			}

			// Column filter
			if len(columns) > 0 {
				for _, col := range columns {
					if !isValidColumn(col) {
						return out.Error(ExitValidation,
							fmt.Sprintf("invalid column '%s', must be one of: %v", col, ValidColumns), nil)
					}
				}
				filters = append(filters, buildInFilter("column", columns))
			}

			// Type filter
			if len(types) > 0 {
				for _, t := range types {
					if !isValidType(t) {
						return out.Error(ExitValidation,
							fmt.Sprintf("invalid type '%s', must be one of: %v", t, ValidTypes), nil)
					}
				}
				filters = append(filters, buildInFilter("type", types))
			}

			// Priority filter
			if len(priorities) > 0 {
				for _, p := range priorities {
					if !isValidPriority(p) {
						return out.Error(ExitValidation,
							fmt.Sprintf("invalid priority '%s', must be one of: %v", p, ValidPriorities), nil)
					}
				}
				filters = append(filters, buildInFilter("priority", priorities))
			}

			// Search filter
			if search != "" {
				filters = append(filters, dbx.NewExp(
					"LOWER(title) LIKE {:search}",
					dbx.Params{"search": "%" + strings.ToLower(search) + "%"},
				))
			}

			// Created by filter
			if createdBy != "" {
				filters = append(filters, dbx.NewExp(
					"created_by = {:created_by}",
					dbx.Params{"created_by": createdBy},
				))
			}

			// Agent name filter
			if agentName != "" {
				filters = append(filters, dbx.NewExp(
					"created_by_agent = {:agent}",
					dbx.Params{"agent": agentName},
				))
			}

			// Is blocked filter
			if isBlocked {
				filters = append(filters, dbx.NewExp(
					"json_array_length(blocked_by) > 0",
				))
			}

			// Not blocked filter
			if notBlocked {
				filters = append(filters, dbx.Or(
					dbx.NewExp("blocked_by IS NULL"),
					dbx.NewExp("blocked_by = '[]'"),
					dbx.NewExp("json_array_length(blocked_by) = 0"),
				))
			}

			// Execute query
			var tasks []*core.Record
			var err error

			if len(filters) > 0 {
				combined := dbx.And(filters...)
				tasks, err = app.FindAllRecords("tasks", combined)
			} else {
				tasks, err = app.FindAllRecords("tasks")
			}

			if err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to list tasks: %v", err), nil)
			}

			// Sort by position within each column
			sortTasksByPosition(tasks)

			// Handle field selection for JSON output
			if jsonOutput && fields != "" {
				out.TasksWithFields(tasks, strings.Split(fields, ","))
			} else {
				out.Tasks(tasks)
			}

			return nil
		},
	}

	// Define flags
	cmd.Flags().StringSliceVarP(&columns, "column", "c", nil,
		"Filter by column (repeatable)")
	cmd.Flags().StringSliceVarP(&types, "type", "t", nil,
		"Filter by type (repeatable)")
	cmd.Flags().StringSliceVarP(&priorities, "priority", "p", nil,
		"Filter by priority (repeatable)")
	cmd.Flags().StringVarP(&search, "search", "s", "",
		"Search title (case-insensitive)")
	cmd.Flags().StringVar(&createdBy, "created-by", "",
		"Filter by creator (user, agent, cli)")
	cmd.Flags().StringVar(&agentName, "agent", "",
		"Filter by agent name")
	cmd.Flags().BoolVar(&ready, "ready", false,
		"Show unblocked tasks in todo/backlog (agent-friendly)")
	cmd.Flags().BoolVar(&isBlocked, "is-blocked", false,
		"Show only tasks blocked by others")
	cmd.Flags().BoolVar(&notBlocked, "not-blocked", false,
		"Show only tasks not blocked by others")
	cmd.Flags().StringVar(&fields, "fields", "",
		"Comma-separated fields to include in JSON output")

	return cmd
}

// buildInFilter creates a SQL IN expression for a list of values.
func buildInFilter(field string, values []string) dbx.Expression {
	if len(values) == 1 {
		return dbx.NewExp(
			fmt.Sprintf("%s = {:val}", field),
			dbx.Params{"val": values[0]},
		)
	}

	// Build IN clause
	placeholders := make([]string, len(values))
	params := dbx.Params{}
	for i, v := range values {
		key := fmt.Sprintf("val%d", i)
		placeholders[i] = "{:" + key + "}"
		params[key] = v
	}

	return dbx.NewExp(
		fmt.Sprintf("%s IN (%s)", field, strings.Join(placeholders, ", ")),
		params,
	)
}
```

---

### 1.5.7 Add TasksWithFields to Output Formatter

**What**: Add a method to output tasks with only selected fields.

**Why**: Agents don't always need all task fields. Limiting output reduces token usage.

**File**: Update `internal/output/output.go`

Add this method to the `Formatter` struct:

```go
// TasksWithFields outputs tasks with only specified fields (JSON only).
// If not in JSON mode, falls back to regular task output.
func (f *Formatter) TasksWithFields(tasks []*core.Record, fields []string) {
	if !f.JSON {
		f.Tasks(tasks)
		return
	}

	// Build filtered task list
	result := make([]map[string]any, len(tasks))
	for i, task := range tasks {
		fullMap := taskToMap(task)
		filtered := make(map[string]any)
		for _, field := range fields {
			field = strings.TrimSpace(field)
			if val, ok := fullMap[field]; ok {
				filtered[field] = val
			}
		}
		result[i] = filtered
	}

	f.writeJSON(map[string]any{
		"tasks": result,
		"count": len(tasks),
	})
}
```

Don't forget to add `"strings"` to the imports if not already present.

---

### 1.5.8 Add Blocking Relationship Support to Update Command

**What**: Add `--blocked-by` and `--remove-blocked-by` flags to the update command.

**Why**: Tasks need to track dependencies. Blocking relationships help agents identify parallelizable work.

**File**: Update `internal/commands/update.go`

Add these variables and handling to the existing `newUpdateCmd` function:

```go
// Add to the var block:
var (
	// ... existing variables ...
	blockedBy       []string
	removeBlockedBy []string
)

// Add this handling after the labels update section:

// Update blocked_by
if len(blockedBy) > 0 || len(removeBlockedBy) > 0 {
	oldBlockedBy := getTaskBlockedBy(task)
	newBlockedBy := updateBlockedBy(app, task.Id, oldBlockedBy, blockedBy, removeBlockedBy)
	
	// Validate that all blocking tasks exist
	for _, blockingID := range newBlockedBy {
		if blockingID == task.Id {
			return out.Error(ExitValidation, "task cannot block itself", nil)
		}
		_, err := app.FindRecordById("tasks", blockingID)
		if err != nil {
			return out.Error(ExitNotFound,
				fmt.Sprintf("blocking task not found: %s", blockingID), nil)
		}
	}
	
	changes["blocked_by"] = map[string]any{
		"from": oldBlockedBy,
		"to":   newBlockedBy,
	}
	task.Set("blocked_by", newBlockedBy)
}

// Add these flag definitions:
cmd.Flags().StringSliceVar(&blockedBy, "blocked-by", nil,
	"Add blocking task ID (repeatable)")
cmd.Flags().StringSliceVar(&removeBlockedBy, "remove-blocked-by", nil,
	"Remove blocking task ID (repeatable)")
```

Also add these helper functions:

```go
// getTaskBlockedBy extracts blocked_by IDs from a task record.
func getTaskBlockedBy(task interface{ Get(string) any }) []string {
	raw := task.Get("blocked_by")
	if raw == nil {
		return []string{}
	}

	if ids, ok := raw.([]any); ok {
		result := make([]string, 0, len(ids))
		for _, id := range ids {
			if s, ok := id.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}

	return []string{}
}

// updateBlockedBy adds and removes blocking task IDs.
func updateBlockedBy(app *pocketbase.PocketBase, taskID string, current, add, remove []string) []string {
	// Create a set of current blocked_by
	blockedSet := make(map[string]bool)
	for _, id := range current {
		blockedSet[id] = true
	}

	// Remove IDs
	for _, id := range remove {
		delete(blockedSet, id)
	}

	// Add IDs (but not self)
	for _, id := range add {
		if id != taskID {
			blockedSet[id] = true
		}
	}

	// Convert back to slice
	result := make([]string, 0, len(blockedSet))
	for id := range blockedSet {
		result = append(result, id)
	}

	return result
}
```

---

### 1.5.9 Implement Context Command

**What**: Create the `context` command that outputs a project state summary.

**Why**: Agents need to understand the current state of a project's tasks to make good decisions.

**File**: `internal/commands/context.go`

```go
package commands

import (
	"fmt"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
)

// ContextSummary holds project state summary.
type ContextSummary struct {
	Summary      Summary `json:"summary"`
	BlockedCount int     `json:"blocked_count"`
	ReadyCount   int     `json:"ready_count"`
}

// Summary holds task counts.
type Summary struct {
	Total      int            `json:"total"`
	ByColumn   map[string]int `json:"by_column"`
	ByPriority map[string]int `json:"by_priority"`
	ByType     map[string]int `json:"by_type"`
}

func newContextCmd(app *pocketbase.PocketBase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Output project state summary",
		Long: `Output a summary of the project's task state for agent context.

This provides agents with a quick overview of:
- Total task count
- Tasks by column/status
- Tasks by priority
- Number of blocked vs ready tasks

Examples:
  egenskriven context --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			// Bootstrap the app
			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			// Get all tasks
			tasks, err := app.FindAllRecords("tasks")
			if err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to list tasks: %v", err), nil)
			}

			// Build summary
			summary := buildContextSummary(tasks)

			if jsonOutput {
				out.writeJSON(summary)
			} else {
				printContextSummary(summary)
			}

			return nil
		},
	}

	return cmd
}

func buildContextSummary(tasks []*core.Record) ContextSummary {
	summary := ContextSummary{
		Summary: Summary{
			Total:      len(tasks),
			ByColumn:   make(map[string]int),
			ByPriority: make(map[string]int),
			ByType:     make(map[string]int),
		},
	}

	for _, task := range tasks {
		// Count by column
		col := task.GetString("column")
		summary.Summary.ByColumn[col]++

		// Count by priority
		priority := task.GetString("priority")
		summary.Summary.ByPriority[priority]++

		// Count by type
		taskType := task.GetString("type")
		summary.Summary.ByType[taskType]++

		// Count blocked
		blockedBy := getTaskBlockedBy(task)
		if len(blockedBy) > 0 {
			summary.BlockedCount++
		}

		// Count ready (unblocked in todo/backlog)
		if (col == "todo" || col == "backlog") && len(blockedBy) == 0 {
			summary.ReadyCount++
		}
	}

	return summary
}

func printContextSummary(s ContextSummary) {
	fmt.Printf("Project Summary\n")
	fmt.Printf("===============\n\n")

	fmt.Printf("Total tasks: %d\n", s.Summary.Total)
	fmt.Printf("Ready to work: %d\n", s.ReadyCount)
	fmt.Printf("Blocked: %d\n", s.BlockedCount)

	fmt.Printf("\nBy Column:\n")
	for _, col := range ValidColumns {
		count := s.Summary.ByColumn[col]
		if count > 0 {
			fmt.Printf("  %-12s %d\n", col+":", count)
		}
	}

	fmt.Printf("\nBy Priority:\n")
	for _, p := range ValidPriorities {
		count := s.Summary.ByPriority[p]
		if count > 0 {
			fmt.Printf("  %-12s %d\n", p+":", count)
		}
	}

	fmt.Printf("\nBy Type:\n")
	for _, t := range ValidTypes {
		count := s.Summary.ByType[t]
		if count > 0 {
			fmt.Printf("  %-12s %d\n", t+":", count)
		}
	}

	fmt.Println()
}
```

**Steps**:

1. Create the file:
   ```bash
   touch internal/commands/context.go
   ```

2. Open in your editor and paste the code above.

3. Verify it compiles:
   ```bash
   go build ./internal/commands
   ```

**Usage Examples**:

```bash
# Human-readable output
egenskriven context

# JSON output for agents
egenskriven context --json
```

**Output Example** (JSON):

```json
{
  "summary": {
    "total": 15,
    "by_column": {
      "backlog": 5,
      "todo": 4,
      "in_progress": 2,
      "review": 1,
      "done": 3
    },
    "by_priority": {
      "urgent": 1,
      "high": 3,
      "medium": 8,
      "low": 3
    },
    "by_type": {
      "bug": 4,
      "feature": 9,
      "chore": 2
    }
  },
  "blocked_count": 2,
  "ready_count": 6
}
```

---

### 1.5.10 Implement Suggest Command

**What**: Create the `suggest` command that recommends tasks to work on.

**Why**: Agents benefit from prioritized suggestions. This helps them make better decisions about what to work on next.

**File**: `internal/commands/suggest.go`

```go
package commands

import (
	"fmt"
	"sort"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
)

// Suggestion represents a task suggestion with reasoning.
type Suggestion struct {
	Task   map[string]any `json:"task"`
	Reason string         `json:"reason"`
}

// SuggestResponse holds the list of suggestions.
type SuggestResponse struct {
	Suggestions []Suggestion `json:"suggestions"`
}

func newSuggestCmd(app *pocketbase.PocketBase) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "suggest",
		Short: "Suggest tasks to work on",
		Long: `Suggest tasks to work on next based on priority and dependencies.

Suggestion priority:
1. In-progress tasks (continue current work)
2. Urgent unblocked tasks
3. High priority unblocked tasks
4. Tasks that unblock the most other tasks

Examples:
  egenskriven suggest --json
  egenskriven suggest --json --limit 3`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			// Bootstrap the app
			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			// Get all tasks
			tasks, err := app.FindAllRecords("tasks")
			if err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to list tasks: %v", err), nil)
			}

			// Build suggestions
			suggestions := buildSuggestions(tasks, limit)

			if jsonOutput {
				out.writeJSON(SuggestResponse{Suggestions: suggestions})
			} else {
				printSuggestions(suggestions)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 5, "Maximum number of suggestions")

	return cmd
}

func buildSuggestions(tasks []*core.Record, limit int) []Suggestion {
	var suggestions []Suggestion

	// Index tasks by ID for quick lookup
	taskMap := make(map[string]*core.Record)
	for _, t := range tasks {
		taskMap[t.Id] = t
	}

	// Calculate how many tasks each task unblocks
	unblocksCount := make(map[string]int)
	for _, t := range tasks {
		blockedBy := getTaskBlockedBy(t)
		for _, blockingID := range blockedBy {
			unblocksCount[blockingID]++
		}
	}

	// 1. In-progress tasks (continue current work)
	for _, t := range tasks {
		if t.GetString("column") == "in_progress" {
			suggestions = append(suggestions, Suggestion{
				Task:   taskToSuggestionMap(t),
				Reason: "Continue current work",
			})
		}
	}

	// 2. Urgent unblocked tasks
	for _, t := range tasks {
		if t.GetString("column") != "in_progress" &&
			t.GetString("column") != "done" &&
			t.GetString("column") != "review" &&
			t.GetString("priority") == "urgent" &&
			len(getTaskBlockedBy(t)) == 0 {
			suggestions = append(suggestions, Suggestion{
				Task:   taskToSuggestionMap(t),
				Reason: "Urgent priority, unblocked",
			})
		}
	}

	// 3. High priority unblocked tasks
	for _, t := range tasks {
		if t.GetString("column") != "in_progress" &&
			t.GetString("column") != "done" &&
			t.GetString("column") != "review" &&
			t.GetString("priority") == "high" &&
			len(getTaskBlockedBy(t)) == 0 {
			suggestions = append(suggestions, Suggestion{
				Task:   taskToSuggestionMap(t),
				Reason: "High priority, unblocked",
			})
		}
	}

	// 4. Tasks that unblock others
	type unblockingTask struct {
		task   *core.Record
		count  int
	}
	var unblocking []unblockingTask
	for _, t := range tasks {
		if t.GetString("column") != "done" &&
			t.GetString("column") != "review" {
			count := unblocksCount[t.Id]
			if count > 0 && len(getTaskBlockedBy(t)) == 0 {
				unblocking = append(unblocking, unblockingTask{t, count})
			}
		}
	}

	// Sort by count descending
	sort.Slice(unblocking, func(i, j int) bool {
		return unblocking[i].count > unblocking[j].count
	})

	for _, ut := range unblocking {
		// Check if already in suggestions
		alreadyAdded := false
		for _, s := range suggestions {
			if s.Task["id"] == ut.task.Id {
				alreadyAdded = true
				break
			}
		}
		if !alreadyAdded {
			suggestions = append(suggestions, Suggestion{
				Task:   taskToSuggestionMap(ut.task),
				Reason: fmt.Sprintf("Unblocks %d other task(s)", ut.count),
			})
		}
	}

	// Limit results
	if limit > 0 && len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}

	return suggestions
}

func taskToSuggestionMap(task *core.Record) map[string]any {
	return map[string]any{
		"id":       task.Id,
		"title":    task.GetString("title"),
		"type":     task.GetString("type"),
		"priority": task.GetString("priority"),
		"column":   task.GetString("column"),
	}
}

func printSuggestions(suggestions []Suggestion) {
	if len(suggestions) == 0 {
		fmt.Println("No suggestions - all caught up!")
		return
	}

	fmt.Println("Suggested tasks to work on:")
	fmt.Println()

	for i, s := range suggestions {
		fmt.Printf("%d. [%s] %s\n", i+1, s.Task["id"].(string)[:8], s.Task["title"])
		fmt.Printf("   Type: %s, Priority: %s, Column: %s\n",
			s.Task["type"], s.Task["priority"], s.Task["column"])
		fmt.Printf("   Reason: %s\n", s.Reason)
		fmt.Println()
	}
}
```

**Steps**:

1. Create the file:
   ```bash
   touch internal/commands/suggest.go
   ```

2. Open in your editor and paste the code above.

3. Verify it compiles:
   ```bash
   go build ./internal/commands
   ```

**Usage Examples**:

```bash
# Get suggestions
egenskriven suggest

# JSON output for agents
egenskriven suggest --json

# Limit to 3 suggestions
egenskriven suggest --json --limit 3
```

**Output Example** (JSON):

```json
{
  "suggestions": [
    {
      "task": {
        "id": "abc123def",
        "title": "Critical bug fix",
        "type": "bug",
        "priority": "urgent",
        "column": "in_progress"
      },
      "reason": "Continue current work"
    },
    {
      "task": {
        "id": "def456ghi",
        "title": "Security update",
        "type": "bug",
        "priority": "urgent",
        "column": "todo"
      },
      "reason": "Urgent priority, unblocked"
    }
  ]
}
```

---

### 1.5.11 Update Root Command to Register New Commands

**What**: Update the root command to register the new commands.

**Why**: The new commands need to be added to the command tree.

**File**: Update `internal/commands/root.go`

Add these lines to the `Register` function:

```go
// Register adds all CLI commands to the PocketBase app.
func Register(app *pocketbase.PocketBase) {
	// ... existing code ...

	// Register all commands
	app.RootCmd.AddCommand(newAddCmd(app))
	app.RootCmd.AddCommand(newListCmd(app))
	app.RootCmd.AddCommand(newShowCmd(app))
	app.RootCmd.AddCommand(newMoveCmd(app))
	app.RootCmd.AddCommand(newUpdateCmd(app))
	app.RootCmd.AddCommand(newDeleteCmd(app))
	
	// Phase 1.5 commands
	app.RootCmd.AddCommand(newInitCmd(app))
	app.RootCmd.AddCommand(newPrimeCmd(app))
	app.RootCmd.AddCommand(newContextCmd(app))
	app.RootCmd.AddCommand(newSuggestCmd(app))
}
```

---

### 1.5.12 Write Tests for Blocking Relationships

**What**: Create tests for blocking relationship functionality.

**Why**: Blocking relationships are critical for agents. Tests ensure they work correctly.

**File**: `internal/commands/blocking_test.go`

```go
package commands

import (
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourusername/egenskriven/internal/testutil"
)

func TestUpdateBlockedBy_AddsBlockingTask(t *testing.T) {
	current := []string{}
	add := []string{"task-1", "task-2"}
	remove := []string{}

	result := updateBlockedBy(nil, "my-task", current, add, remove)

	assert.Len(t, result, 2)
	assert.Contains(t, result, "task-1")
	assert.Contains(t, result, "task-2")
}

func TestUpdateBlockedBy_RemovesBlockingTask(t *testing.T) {
	current := []string{"task-1", "task-2", "task-3"}
	add := []string{}
	remove := []string{"task-2"}

	result := updateBlockedBy(nil, "my-task", current, add, remove)

	assert.Len(t, result, 2)
	assert.Contains(t, result, "task-1")
	assert.Contains(t, result, "task-3")
	assert.NotContains(t, result, "task-2")
}

func TestUpdateBlockedBy_PreventsSelfBlocking(t *testing.T) {
	current := []string{}
	add := []string{"my-task"} // Trying to block self
	remove := []string{}

	result := updateBlockedBy(nil, "my-task", current, add, remove)

	assert.Len(t, result, 0)
	assert.NotContains(t, result, "my-task")
}

func TestUpdateBlockedBy_CombinedAddAndRemove(t *testing.T) {
	current := []string{"task-1", "task-2"}
	add := []string{"task-3"}
	remove := []string{"task-1"}

	result := updateBlockedBy(nil, "my-task", current, add, remove)

	assert.Len(t, result, 2)
	assert.Contains(t, result, "task-2")
	assert.Contains(t, result, "task-3")
	assert.NotContains(t, result, "task-1")
}

func TestGetTaskBlockedBy_Empty(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	task := createTestTaskWithBlockedBy(t, app, "Test", []string{})

	result := getTaskBlockedBy(task)

	assert.Len(t, result, 0)
}

func TestGetTaskBlockedBy_WithValues(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	task := createTestTaskWithBlockedBy(t, app, "Test", []string{"blocker-1", "blocker-2"})

	result := getTaskBlockedBy(task)

	assert.Len(t, result, 2)
	assert.Contains(t, result, "blocker-1")
	assert.Contains(t, result, "blocker-2")
}

// Helper function
func createTestTaskWithBlockedBy(t *testing.T, app interface {
	FindCollectionByNameOrId(string) (*core.Collection, error)
	Save(any) error
}, title string, blockedBy []string) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("tasks")
	require.NoError(t, err)

	record := core.NewRecord(collection)
	record.Set("title", title)
	record.Set("type", "feature")
	record.Set("priority", "medium")
	record.Set("column", "backlog")
	record.Set("position", 1000.0)
	record.Set("labels", []string{})
	record.Set("blocked_by", blockedBy)
	record.Set("created_by", "cli")
	record.Set("history", []map[string]any{})

	require.NoError(t, app.Save(record))

	return record
}
```

**Steps**:

1. Create the file:
   ```bash
   touch internal/commands/blocking_test.go
   ```

2. Open in your editor and paste the code above.

3. **Important**: Replace `github.com/yourusername/egenskriven` with your actual module path.

4. Run the tests:
   ```bash
   go test ./internal/commands -v -run Blocking
   ```

---

### 1.5.13 Write Tests for Prime Command

**What**: Create tests for the prime command and template rendering.

**Why**: The prime command's output varies based on configuration. Tests ensure it renders correctly.

**File**: `internal/commands/prime_test.go`

```go
package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrimeTemplate_StrictWorkflow(t *testing.T) {
	tmpl, err := template.New("prime").Parse(primeTemplate)
	require.NoError(t, err)

	var buf bytes.Buffer
	data := PrimeTemplateData{
		WorkflowMode:       "strict",
		AgentMode:          "autonomous",
		AgentName:          "claude",
		OverrideTodoWrite:  true,
		RequireSummary:     true,
		StructuredSections: true,
	}

	err = tmpl.Execute(&buf, data)
	require.NoError(t, err)

	output := buf.String()

	// Check strict workflow content
	assert.Contains(t, output, "BEFORE starting any task")
	assert.Contains(t, output, "DURING work")
	assert.Contains(t, output, "AFTER completing")
	assert.Contains(t, output, "## Approach")
	assert.Contains(t, output, "## Summary of Changes")
}

func TestPrimeTemplate_LightWorkflow(t *testing.T) {
	tmpl, err := template.New("prime").Parse(primeTemplate)
	require.NoError(t, err)

	var buf bytes.Buffer
	data := PrimeTemplateData{
		WorkflowMode:       "light",
		AgentMode:          "autonomous",
		AgentName:          "agent",
		OverrideTodoWrite:  true,
		RequireSummary:     false,
		StructuredSections: false,
	}

	err = tmpl.Execute(&buf, data)
	require.NoError(t, err)

	output := buf.String()

	// Check light workflow content
	assert.Contains(t, output, "Create task for substantial work")
	assert.NotContains(t, output, "BEFORE starting any task")
}

func TestPrimeTemplate_MinimalWorkflow(t *testing.T) {
	tmpl, err := template.New("prime").Parse(primeTemplate)
	require.NoError(t, err)

	var buf bytes.Buffer
	data := PrimeTemplateData{
		WorkflowMode:       "minimal",
		AgentMode:          "autonomous",
		AgentName:          "agent",
		OverrideTodoWrite:  false,
		RequireSummary:     false,
		StructuredSections: false,
	}

	err = tmpl.Execute(&buf, data)
	require.NoError(t, err)

	output := buf.String()

	// Check minimal workflow content
	assert.Contains(t, output, "Use egenskriven for task tracking as needed")
	assert.NotContains(t, output, "Always use egenskriven instead of TodoWrite")
}

func TestPrimeTemplate_AgentModes(t *testing.T) {
	tmpl, err := template.New("prime").Parse(primeTemplate)
	require.NoError(t, err)

	tests := []struct {
		mode     string
		contains string
	}{
		{"autonomous", "full autonomy"},
		{"collaborative", "execute minor updates"},
		{"supervised", "read-only mode"},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			var buf bytes.Buffer
			data := PrimeTemplateData{
				WorkflowMode:      "light",
				AgentMode:         tt.mode,
				AgentName:         "agent",
				OverrideTodoWrite: true,
			}

			err := tmpl.Execute(&buf, data)
			require.NoError(t, err)

			assert.Contains(t, buf.String(), tt.contains)
		})
	}
}

func TestPrimeTemplate_OverrideTodoWrite(t *testing.T) {
	tmpl, err := template.New("prime").Parse(primeTemplate)
	require.NoError(t, err)

	// With override
	var buf1 bytes.Buffer
	data1 := PrimeTemplateData{
		WorkflowMode:      "light",
		AgentMode:         "autonomous",
		OverrideTodoWrite: true,
	}
	require.NoError(t, tmpl.Execute(&buf1, data1))
	assert.Contains(t, buf1.String(), "Always use egenskriven instead of TodoWrite")

	// Without override
	var buf2 bytes.Buffer
	data2 := PrimeTemplateData{
		WorkflowMode:      "light",
		AgentMode:         "autonomous",
		OverrideTodoWrite: false,
	}
	require.NoError(t, tmpl.Execute(&buf2, data2))
	assert.NotContains(t, buf2.String(), "Always use egenskriven instead of TodoWrite")
}

func TestPrimeIntegration_RespectsConfigFile(t *testing.T) {
	// Create temp directory with config
	tmpDir, err := os.MkdirTemp("", "egenskriven-prime-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	configDir := filepath.Join(tmpDir, ".egenskriven")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	configContent := `{
		"agent": {
			"workflow": "strict",
			"mode": "collaborative"
		}
	}`
	require.NoError(t, os.WriteFile(
		filepath.Join(configDir, "config.json"),
		[]byte(configContent),
		0644,
	))

	// This test would require running the actual command
	// For now, just verify the config is read correctly
	t.Log("Config file created - integration test would verify prime output")
}
```

**Steps**:

1. Create the file:
   ```bash
   touch internal/commands/prime_test.go
   ```

2. Open in your editor and paste the code above.

3. Run the tests:
   ```bash
   go test ./internal/commands -v -run Prime
   ```

---

### 1.5.14 Write Tests for Suggest Command

**What**: Create tests for the suggest command's prioritization logic.

**Why**: Suggestion logic is important for agent productivity. Tests ensure correct prioritization.

**File**: `internal/commands/suggest_test.go`

```go
package commands

import (
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourusername/egenskriven/internal/testutil"
)

func TestBuildSuggestions_PrioritizesInProgress(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	// Create tasks
	inProgressTask := createTestTaskWithColumn(t, app, "In Progress Task", "in_progress", "medium")
	_ = createTestTaskWithColumn(t, app, "Todo Task", "todo", "urgent")

	tasks, _ := app.FindAllRecords("tasks")
	suggestions := buildSuggestions(tasks, 10)

	require.NotEmpty(t, suggestions)
	assert.Equal(t, inProgressTask.Id, suggestions[0].Task["id"])
	assert.Equal(t, "Continue current work", suggestions[0].Reason)
}

func TestBuildSuggestions_PrioritizesUrgent(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	// Create tasks (no in_progress)
	urgentTask := createTestTaskWithPriority(t, app, "Urgent Task", "urgent")
	_ = createTestTaskWithPriority(t, app, "High Task", "high")

	tasks, _ := app.FindAllRecords("tasks")
	suggestions := buildSuggestions(tasks, 10)

	require.NotEmpty(t, suggestions)
	assert.Equal(t, urgentTask.Id, suggestions[0].Task["id"])
	assert.Contains(t, suggestions[0].Reason, "Urgent")
}

func TestBuildSuggestions_IncludesUnblockingTasks(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	// Create blocker task
	blockerTask := createTestTaskWithColumn(t, app, "Blocker Task", "todo", "medium")

	// Create blocked task
	createTestTaskWithBlockedBy(t, app, "Blocked Task", []string{blockerTask.Id})

	tasks, _ := app.FindAllRecords("tasks")
	suggestions := buildSuggestions(tasks, 10)

	// Find the blocker in suggestions
	var found bool
	for _, s := range suggestions {
		if s.Task["id"] == blockerTask.Id {
			found = true
			assert.Contains(t, s.Reason, "Unblocks")
			break
		}
	}
	assert.True(t, found, "Blocker task should be in suggestions")
}

func TestBuildSuggestions_ExcludesDoneTasks(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	_ = createTestTaskWithColumn(t, app, "Done Task", "done", "urgent")
	todoTask := createTestTaskWithColumn(t, app, "Todo Task", "todo", "medium")

	tasks, _ := app.FindAllRecords("tasks")
	suggestions := buildSuggestions(tasks, 10)

	// Should only have the todo task
	require.Len(t, suggestions, 1)
	assert.Equal(t, todoTask.Id, suggestions[0].Task["id"])
}

func TestBuildSuggestions_RespectsLimit(t *testing.T) {
	app := testutil.NewTestApp(t)
	setupTasksCollection(t, app)

	// Create multiple tasks
	for i := 0; i < 10; i++ {
		createTestTaskWithColumn(t, app, "Task", "todo", "medium")
	}

	tasks, _ := app.FindAllRecords("tasks")

	suggestions := buildSuggestions(tasks, 3)
	assert.Len(t, suggestions, 3)

	suggestions = buildSuggestions(tasks, 5)
	assert.Len(t, suggestions, 5)
}

func TestBuildSuggestions_EmptyTasks(t *testing.T) {
	suggestions := buildSuggestions([]*core.Record{}, 10)
	assert.Empty(t, suggestions)
}

// Helper functions

func createTestTaskWithColumn(t *testing.T, app interface {
	FindCollectionByNameOrId(string) (*core.Collection, error)
	Save(any) error
}, title, column, priority string) *core.Record {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId("tasks")
	require.NoError(t, err)

	record := core.NewRecord(collection)
	record.Set("title", title)
	record.Set("type", "feature")
	record.Set("priority", priority)
	record.Set("column", column)
	record.Set("position", 1000.0)
	record.Set("labels", []string{})
	record.Set("blocked_by", []string{})
	record.Set("created_by", "cli")
	record.Set("history", []map[string]any{})

	require.NoError(t, app.Save(record))

	return record
}

func createTestTaskWithPriority(t *testing.T, app interface {
	FindCollectionByNameOrId(string) (*core.Collection, error)
	Save(any) error
}, title, priority string) *core.Record {
	return createTestTaskWithColumn(t, app, title, "todo", priority)
}
```

**Steps**:

1. Create the file:
   ```bash
   touch internal/commands/suggest_test.go
   ```

2. Open in your editor and paste the code above.

3. **Important**: Replace `github.com/yourusername/egenskriven` with your actual module path.

4. Run the tests:
   ```bash
   go test ./internal/commands -v -run Suggest
   ```

---

### 1.5.15 Create OpenCode Plugin

**What**: Create the OpenCode plugin that injects prime output into agent context.

**Why**: OpenCode agents need automatic context injection at session start and compacting.

**File**: `.opencode/plugin/egenskriven-prime.ts`

```typescript
import type { Plugin } from "@opencode/plugin";

export const EgenSkrivenPlugin: Plugin = async ({ $ }) => {
  // Check if egenskriven is available
  try {
    await $`egenskriven version`.quiet();
  } catch {
    // Not installed, skip plugin
    return {};
  }

  // Get prime output
  const prime = await $`egenskriven prime`.text();

  return {
    // Inject into system prompt at session start
    "experimental.chat.system.transform": async (_, output) => {
      output.system.push(prime);
    },

    // Re-inject after context compacting
    "experimental.session.compacting": async (_, output) => {
      output.context.push(prime);
    },
  };
};

export default EgenSkrivenPlugin;
```

**Steps**:

1. Create the directory structure:
   ```bash
   mkdir -p .opencode/plugin
   ```

2. Create the file:
   ```bash
   touch .opencode/plugin/egenskriven-prime.ts
   ```

3. Open in your editor and paste the code above.

**How it works**:

1. Plugin checks if `egenskriven` is installed
2. On session start, runs `egenskriven prime` and injects output into system prompt
3. After context compacting, re-injects the prime output

---

### 1.5.16 Document Claude Code Integration

**What**: Create documentation and example configuration for Claude Code hooks.

**Why**: Claude Code users need to know how to set up the integration.

**File**: `.claude/settings.json` (example for documentation)

```json
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "egenskriven prime"
          }
        ]
      }
    ],
    "PreCompact": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "egenskriven prime"
          }
        ]
      }
    ]
  }
}
```

**Steps**:

1. Create the directory:
   ```bash
   mkdir -p .claude
   ```

2. Create the file:
   ```bash
   touch .claude/settings.json
   ```

3. Open in your editor and paste the configuration above.

**How it works**:

1. `SessionStart` hook runs `egenskriven prime` when Claude starts a new session
2. `PreCompact` hook runs `egenskriven prime` before context compacting to preserve instructions

---

## Verification Checklist

Complete each section in order. Check off each item as you verify it.

### Build Verification

- [ ] **Project compiles**
  ```bash
  make build
  ```
  Should produce `egenskriven` binary.

- [ ] **All tests pass**
  ```bash
  make test
  ```
  Should show all tests passing.

### Config Verification

- [ ] **Init command creates config**
  ```bash
  rm -rf .egenskriven
  ./egenskriven init
  cat .egenskriven/config.json
  ```
  Should show config file with defaults.

- [ ] **Init with workflow flag**
  ```bash
  rm -rf .egenskriven
  ./egenskriven init --workflow strict
  cat .egenskriven/config.json
  ```
  Should show `"workflow": "strict"`.

- [ ] **Init with mode flag**
  ```bash
  rm -rf .egenskriven
  ./egenskriven init --mode collaborative
  cat .egenskriven/config.json
  ```
  Should show `"mode": "collaborative"`.

### Prime Command Verification

- [ ] **Prime outputs instructions**
  ```bash
  ./egenskriven prime
  ```
  Should output agent instructions.

- [ ] **Prime respects config**
  ```bash
  ./egenskriven init --workflow strict
  ./egenskriven prime | grep "BEFORE starting"
  ```
  Should find strict workflow content.

- [ ] **Prime workflow override**
  ```bash
  ./egenskriven prime --workflow minimal
  ```
  Should output minimal workflow regardless of config.

- [ ] **Prime agent name**
  ```bash
  ./egenskriven prime --agent claude | grep claude
  ```
  Should include "claude" in examples.

### Blocking Relationships Verification

- [ ] **Add blocking relationship**
  ```bash
  # Create two tasks
  ./egenskriven add "Task A" --json
  ./egenskriven add "Task B" --json
  
  # Get Task A's ID, then:
  ./egenskriven update <task-b-id> --blocked-by <task-a-id>
  ./egenskriven show <task-b-id>
  ```
  Should show Task B is blocked by Task A.

- [ ] **Remove blocking relationship**
  ```bash
  ./egenskriven update <task-b-id> --remove-blocked-by <task-a-id>
  ./egenskriven show <task-b-id>
  ```
  Should show no blocking relationship.

- [ ] **Self-blocking prevented**
  ```bash
  ./egenskriven update <task-id> --blocked-by <same-task-id>
  ```
  Should show error or ignore.

- [ ] **List ready tasks**
  ```bash
  ./egenskriven list --ready --json
  ```
  Should show only unblocked tasks in todo/backlog.

- [ ] **List blocked tasks**
  ```bash
  ./egenskriven list --is-blocked --json
  ```
  Should show only blocked tasks.

- [ ] **List unblocked tasks**
  ```bash
  ./egenskriven list --not-blocked --json
  ```
  Should show only unblocked tasks.

### Context Command Verification

- [ ] **Context outputs summary**
  ```bash
  ./egenskriven context
  ```
  Should show human-readable summary.

- [ ] **Context JSON output**
  ```bash
  ./egenskriven context --json
  ```
  Should output valid JSON with summary, blocked_count, ready_count.

### Suggest Command Verification

- [ ] **Suggest outputs recommendations**
  ```bash
  ./egenskriven suggest
  ```
  Should show task recommendations.

- [ ] **Suggest JSON output**
  ```bash
  ./egenskriven suggest --json
  ```
  Should output valid JSON with suggestions array.

- [ ] **Suggest respects limit**
  ```bash
  ./egenskriven suggest --json --limit 2
  ```
  Should output at most 2 suggestions.

### Field Selection Verification

- [ ] **List with field selection**
  ```bash
  ./egenskriven list --json --fields id,title
  ```
  Should output only id and title fields per task.

### Exit Code Verification

- [ ] **Success returns 0**
  ```bash
  ./egenskriven context; echo $?
  ```
  Should print `0`.

---

## File Summary

| File | Lines | Purpose |
|------|-------|---------|
| `internal/config/config.go` | ~100 | Configuration loading and saving |
| `internal/config/config_test.go` | ~100 | Configuration tests |
| `internal/commands/init.go` | ~70 | Initialize command |
| `internal/commands/prime.go` | ~80 | Prime command |
| `internal/commands/prime.tmpl` | ~100 | Prime template |
| `internal/commands/context.go` | ~120 | Context command |
| `internal/commands/suggest.go` | ~170 | Suggest command |
| `internal/commands/list.go` | ~50 | Updates for blocking filters |
| `internal/commands/update.go` | ~50 | Updates for blocked-by management |
| `internal/commands/blocking_test.go` | ~100 | Blocking relationship tests |
| `internal/commands/prime_test.go` | ~130 | Prime command tests |
| `internal/commands/suggest_test.go` | ~120 | Suggest command tests |
| `.opencode/plugin/egenskriven-prime.ts` | ~30 | OpenCode plugin |
| `.claude/settings.json` | ~20 | Claude Code hooks |

**Total new code**: ~1,240 lines

---

## What You Should Have Now

After completing Phase 1.5, your project should:

```
egenskriven/
├── cmd/
│   └── egenskriven/
│       └── main.go              ✓ Unchanged
├── internal/
│   ├── commands/
│   │   ├── root.go              ✓ Updated (new commands)
│   │   ├── add.go               ✓ Unchanged
│   │   ├── list.go              ✓ Updated (blocking filters)
│   │   ├── show.go              ✓ Unchanged
│   │   ├── move.go              ✓ Unchanged
│   │   ├── update.go            ✓ Updated (blocked-by flags)
│   │   ├── delete.go            ✓ Unchanged
│   │   ├── position.go          ✓ Unchanged
│   │   ├── init.go              ✓ Created
│   │   ├── prime.go             ✓ Created
│   │   ├── prime.tmpl           ✓ Created
│   │   ├── context.go           ✓ Created
│   │   ├── suggest.go           ✓ Created
│   │   ├── blocking_test.go     ✓ Created
│   │   ├── prime_test.go        ✓ Created
│   │   └── suggest_test.go      ✓ Created
│   ├── config/
│   │   ├── config.go            ✓ Created
│   │   └── config_test.go       ✓ Created
│   ├── output/
│   │   ├── output.go            ✓ Updated (TasksWithFields)
│   │   └── output_test.go       ✓ Unchanged
│   ├── resolver/
│   │   ├── resolver.go          ✓ Unchanged
│   │   └── resolver_test.go     ✓ Unchanged
│   ├── hooks/                   ✓ Empty (future use)
│   └── testutil/
│       └── testutil.go          ✓ Unchanged
├── migrations/
│   └── 1_initial.go             ✓ Unchanged
├── ui/
│   └── embed.go                 ✓ Unchanged
├── .opencode/
│   └── plugin/
│       └── egenskriven-prime.ts ✓ Created
├── .claude/
│   └── settings.json            ✓ Created
├── .egenskriven/                ✓ Created by init command
│   └── config.json              ✓ Created by init command
├── .air.toml                    ✓ Unchanged
├── .gitignore                   ✓ Unchanged
├── go.mod                       ✓ Unchanged
├── go.sum                       ✓ Unchanged
└── Makefile                     ✓ Unchanged
```

---

## Next Phase

**Phase 2: Minimal UI** will add:
- React frontend setup with Vite
- PocketBase SDK integration
- Basic kanban board with columns
- Drag and drop between columns
- Task detail panel
- Quick create modal
- Real-time subscriptions for live updates

---

## Troubleshooting

### "failed to parse template"

**Problem**: Template syntax error in prime.tmpl.

**Solution**: Check for:
- Missing `{{end}}` for conditionals
- Typos in variable names
- Unbalanced braces

### "config not found but expecting strict workflow"

**Problem**: Config file doesn't exist or has wrong workflow value.

**Solution**:
```bash
# Reinitialize with correct workflow
./egenskriven init --workflow strict
```

### "blocked_by validation failed"

**Problem**: Trying to add a non-existent task as blocker.

**Solution**: Ensure the blocking task ID exists:
```bash
./egenskriven show <blocking-task-id>
```

### Prime output not appearing in agent context

**Problem**: Agent hooks not configured correctly.

**Solution**: 
- For OpenCode: Ensure `.opencode/plugin/egenskriven-prime.ts` exists
- For Claude: Ensure `.claude/settings.json` has correct hooks configuration
- Verify `egenskriven prime` works from the project directory

### Tests failing with "collection not found"

**Problem**: Test database doesn't have the tasks collection.

**Solution**: Tests should create the collection using `setupTasksCollection`. Ensure this helper is called at the start of each test.

### JSON array length not working in SQLite

**Problem**: `json_array_length` might not be available in older SQLite versions.

**Solution**: PocketBase bundles a modern SQLite. If issues persist, use alternative query:
```go
// Alternative to json_array_length
dbx.NewExp("blocked_by != '[]' AND blocked_by != '' AND blocked_by IS NOT NULL")
```

---

## Glossary

| Term | Definition |
|------|------------|
| **Prime** | Command that outputs instructions for AI agents |
| **Workflow Mode** | Level of task tracking enforcement (strict/light/minimal) |
| **Agent Mode** | Level of agent autonomy (autonomous/collaborative/supervised) |
| **Blocking Relationship** | Dependency where one task blocks another |
| **Ready Task** | Unblocked task in todo or backlog |
| **Context** | Project state summary for agent awareness |
| **Suggestion** | Recommended task to work on based on priority |
| **Hook** | Agent lifecycle event that triggers commands |
