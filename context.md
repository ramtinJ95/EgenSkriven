# Config System Refactor: Environment Variables → Config File

## Overview

This document describes the refactoring of EgenSkriven's configuration system from environment variables to a unified config file approach. The goal is to have all configuration in `~/.config/egenskriven/config.json` (global) with project-level overrides in `.egenskriven/config.json`.

## Current State

### Environment Variables (to be removed)

| Variable | Location | Purpose |
|----------|----------|---------|
| `EGENSKRIVEN_DIR` | `cmd/egenskriven/main.go:21` | Data directory path |
| `EGENSKRIVEN_AGENT` | `internal/commands/block.go:182` | Default agent name |
| `EGENSKRIVEN_AUTHOR` | `internal/commands/comment.go:158` | Default comment author |
| `USER` (fallback) | `internal/commands/comment.go:161` | Fallback for author |

### Current Config Structure

**Project config** (`.egenskriven/config.json`):
```json
{
  "agent": {
    "workflow": "light",
    "mode": "autonomous",
    "overrideTodoWrite": true,
    "requireSummary": false,
    "structuredSections": false
  },
  "server": {
    "url": "http://localhost:8090"
  },
  "default_board": "WRK"
}
```

### Resume Mode
- Currently stored per-board in database (`resume_mode` field)
- Valid values: `manual`, `command`, `auto`
- Default: `command`

---

## Target State

### Config File Locations

| File | Purpose | Priority |
|------|---------|----------|
| `~/.config/egenskriven/config.json` | Global user preferences | Base |
| `.egenskriven/config.json` | Project-specific overrides | Override |

### Global Config Structure (`~/.config/egenskriven/config.json`)

```json
{
  "data_dir": "~/.egenskriven",
  "defaults": {
    "author": "username",
    "agent": "claude"
  },
  "agent": {
    "workflow": "light",
    "mode": "autonomous",
    "resume_mode": "command"
  },
  "server": {
    "url": "http://localhost:8090"
  }
}
```

### Project Config Structure (`.egenskriven/config.json`)

```json
{
  "default_board": "WRK",
  "agent": {
    "workflow": "strict",
    "mode": "collaborative",
    "resume_mode": "auto",
    "overrideTodoWrite": true,
    "requireSummary": false,
    "structuredSections": false
  },
  "server": {
    "url": "http://localhost:9090"
  }
}
```

### Merge Behavior

Values are merged with project config taking precedence:

| Setting | Global | Project | Effective |
|---------|--------|---------|-----------|
| `data_dir` | `~/.egenskriven` | - | `~/.egenskriven` (global only) |
| `defaults.author` | `ramtinj` | - | `ramtinj` (global only) |
| `defaults.agent` | `claude` | - | `claude` (global only) |
| `agent.workflow` | `light` | `strict` | `strict` (project wins) |
| `agent.mode` | `autonomous` | `collaborative` | `collaborative` (project wins) |
| `agent.resume_mode` | `command` | `auto` | `auto` (project wins) |
| `server.url` | `http://localhost:8090` | `http://localhost:9090` | `http://localhost:9090` (project wins) |
| `default_board` | - | `WRK` | `WRK` (project only) |

### Settings Classification

**Global-only settings** (cannot be overridden per-project):
- `data_dir` - Database location must be consistent
- `defaults.author` - User identity
- `defaults.agent` - User's preferred AI agent

**Overridable settings** (global default, project can override):
- `agent.workflow` - Workflow strictness
- `agent.mode` - Agent autonomy level
- `agent.resume_mode` - Default resume mode for new boards
- `agent.overrideTodoWrite` - TodoWrite behavior
- `agent.requireSummary` - Summary requirement
- `agent.structuredSections` - Section structure
- `server.url` - Server connection

**Project-only settings**:
- `default_board` - Only meaningful per-project

---

## Implementation Plan

### Phase 1: Update Config Types

**File: `internal/config/config.go`**

1. Add new types for global config:

```go
// GlobalConfig represents user-wide configuration.
// Location: ~/.config/egenskriven/config.json
type GlobalConfig struct {
    DataDir  string         `json:"data_dir,omitempty"`
    Defaults DefaultsConfig `json:"defaults,omitempty"`
    Agent    AgentConfig    `json:"agent,omitempty"`
    Server   ServerConfig   `json:"server,omitempty"`
}

// DefaultsConfig contains default values for commands.
type DefaultsConfig struct {
    Author string `json:"author,omitempty"`
    Agent  string `json:"agent,omitempty"`
}
```

2. Add `ResumeMode` to `AgentConfig`:

```go
type AgentConfig struct {
    Workflow           string `json:"workflow,omitempty"`
    Mode               string `json:"mode,omitempty"`
    ResumeMode         string `json:"resume_mode,omitempty"`  // NEW
    OverrideTodoWrite  bool   `json:"overrideTodoWrite,omitempty"`
    RequireSummary     bool   `json:"requireSummary,omitempty"`
    StructuredSections bool   `json:"structuredSections,omitempty"`
}
```

3. Add validation for resume mode:

```go
var ValidResumeModes = []string{"manual", "command", "auto"}

func ValidateResumeMode(mode string) error {
    for _, valid := range ValidResumeModes {
        if mode == valid {
            return nil
        }
    }
    return fmt.Errorf("invalid resume_mode '%s', must be one of: %v", mode, ValidResumeModes)
}
```

### Phase 2: Add Global Config Loading

**File: `internal/config/config.go`**

1. Add global config path helper:

```go
// GlobalConfigPath returns the path to the global config file.
// Returns ~/.config/egenskriven/config.json
func GlobalConfigPath() (string, error) {
    home, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(home, ".config", "egenskriven", "config.json"), nil
}
```

2. Add global config loader:

```go
// LoadGlobalConfig loads configuration from ~/.config/egenskriven/config.json.
// Returns default config if file doesn't exist.
func LoadGlobalConfig() (*GlobalConfig, error) {
    configPath, err := GlobalConfigPath()
    if err != nil {
        return DefaultGlobalConfig(), nil
    }

    data, err := os.ReadFile(configPath)
    if os.IsNotExist(err) {
        return DefaultGlobalConfig(), nil
    }
    if err != nil {
        return nil, err
    }

    cfg := DefaultGlobalConfig()
    if err := json.Unmarshal(data, cfg); err != nil {
        return nil, err
    }

    // Expand ~ in data_dir
    if strings.HasPrefix(cfg.DataDir, "~/") {
        home, _ := os.UserHomeDir()
        cfg.DataDir = filepath.Join(home, cfg.DataDir[2:])
    }

    return cfg, nil
}

// DefaultGlobalConfig returns global configuration with default values.
func DefaultGlobalConfig() *GlobalConfig {
    return &GlobalConfig{
        DataDir: "", // Empty means use PocketBase default (pb_data in cwd)
        Defaults: DefaultsConfig{
            Author: "",
            Agent:  "agent",
        },
        Agent: AgentConfig{
            Workflow:   "light",
            Mode:       "autonomous",
            ResumeMode: "command",
        },
        Server: ServerConfig{
            URL: "http://localhost:8090",
        },
    }
}
```

3. Add save function for global config:

```go
// SaveGlobalConfig saves configuration to ~/.config/egenskriven/config.json.
func SaveGlobalConfig(cfg *GlobalConfig) error {
    configPath, err := GlobalConfigPath()
    if err != nil {
        return err
    }

    // Create directory if needed
    if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
        return err
    }

    data, err := json.MarshalIndent(cfg, "", "  ")
    if err != nil {
        return err
    }

    return os.WriteFile(configPath, data, 0644)
}
```

### Phase 3: Add Merged Config Loading

**File: `internal/config/config.go`**

1. Add merged config type and loader:

```go
// MergedConfig represents the effective configuration after merging
// global and project configs.
type MergedConfig struct {
    // From global only
    DataDir  string
    Defaults DefaultsConfig

    // Merged (project overrides global)
    Agent  AgentConfig
    Server ServerConfig

    // From project only
    DefaultBoard string
}

// Load loads and merges global and project configuration.
// Project config values override global config where both are set.
func Load() (*MergedConfig, error) {
    global, err := LoadGlobalConfig()
    if err != nil {
        return nil, fmt.Errorf("loading global config: %w", err)
    }

    project, err := LoadProjectConfig()
    if err != nil {
        return nil, fmt.Errorf("loading project config: %w", err)
    }

    return merge(global, project), nil
}

// merge combines global and project configs with project taking precedence.
func merge(global *GlobalConfig, project *Config) *MergedConfig {
    merged := &MergedConfig{
        // Global-only values
        DataDir:  global.DataDir,
        Defaults: global.Defaults,

        // Start with global agent config
        Agent: global.Agent,

        // Start with global server config
        Server: global.Server,

        // Project-only values
        DefaultBoard: project.DefaultBoard,
    }

    // Override agent settings from project if set
    if project.Agent.Workflow != "" {
        merged.Agent.Workflow = project.Agent.Workflow
    }
    if project.Agent.Mode != "" {
        merged.Agent.Mode = project.Agent.Mode
    }
    if project.Agent.ResumeMode != "" {
        merged.Agent.ResumeMode = project.Agent.ResumeMode
    }
    // Boolean fields: project always wins if project config was loaded
    merged.Agent.OverrideTodoWrite = project.Agent.OverrideTodoWrite
    merged.Agent.RequireSummary = project.Agent.RequireSummary
    merged.Agent.StructuredSections = project.Agent.StructuredSections

    // Override server URL from project if set
    if project.Server.URL != "" {
        merged.Server.URL = project.Server.URL
    }

    return merged
}
```

### Phase 4: Update main.go

**File: `cmd/egenskriven/main.go`**

Remove environment variable usage and use config instead:

```go
package main

import (
    "log"
    "strings"

    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/core"

    "github.com/ramtinJ95/EgenSkriven/internal/board"
    "github.com/ramtinJ95/EgenSkriven/internal/commands"
    "github.com/ramtinJ95/EgenSkriven/internal/config"
    "github.com/ramtinJ95/EgenSkriven/internal/hooks"
    _ "github.com/ramtinJ95/EgenSkriven/migrations"
    "github.com/ramtinJ95/EgenSkriven/ui"
)

func main() {
    // Load global config for data directory
    globalCfg, err := config.LoadGlobalConfig()
    if err != nil {
        log.Printf("Warning: failed to load global config: %v", err)
        globalCfg = config.DefaultGlobalConfig()
    }

    // Create PocketBase with configured data directory
    var app *pocketbase.PocketBase
    if globalCfg.DataDir != "" {
        app = pocketbase.NewWithConfig(pocketbase.Config{
            DefaultDataDir: globalCfg.DataDir,
        })
    } else {
        app = pocketbase.New()
    }

    // ... rest unchanged
}
```

### Phase 5: Update block.go

**File: `internal/commands/block.go`**

Replace environment variable with config:

```go
// getDefaultAgentName returns the default agent name from config.
func getDefaultAgentName() string {
    cfg, err := config.LoadGlobalConfig()
    if err != nil {
        return "agent"
    }
    if cfg.Defaults.Agent != "" {
        return cfg.Defaults.Agent
    }
    return "agent"
}
```

Update `getAgentName()` function (around line 180):

```go
// getAgentName returns the agent name, using config default if not specified.
func getAgentName(flagValue string) string {
    if flagValue != "" {
        return flagValue
    }
    return getDefaultAgentName()
}
```

### Phase 6: Update comment.go

**File: `internal/commands/comment.go`**

Replace environment variable with config:

```go
// getDefaultAuthor returns the default author from config or system user.
func getDefaultAuthor() string {
    cfg, err := config.LoadGlobalConfig()
    if err == nil && cfg.Defaults.Author != "" {
        return cfg.Defaults.Author
    }
    // Fallback to system user
    if user := os.Getenv("USER"); user != "" {
        return user
    }
    return ""
}
```

Update `getAuthorName()` function (around line 153):

```go
// getAuthorName returns the author name from flag or config default.
func getAuthorName(flagValue string) string {
    if flagValue != "" {
        return flagValue
    }
    return getDefaultAuthor()
}
```

Also update the help text that mentions `EGENSKRIVEN_AGENT` (around line 174):

```go
// Remove the environment variable hints from help text
```

### Phase 7: Update client.go

**File: `internal/commands/client.go`**

The client already uses `config.LoadProjectConfig()` for server URL. Update to use merged config:

```go
// NewAPIClient creates a new API client with the configured server URL.
func NewAPIClient() *APIClient {
    baseURL := DefaultServerURL

    // Use merged config to get effective server URL
    if cfg, err := config.Load(); err == nil && cfg.Server.URL != "" {
        baseURL = cfg.Server.URL
    }

    return &APIClient{
        baseURL: baseURL,
        // ... rest unchanged
    }
}
```

### Phase 8: Update init.go (if exists)

Update the `init` command to optionally create/update global config.

Add new flags:
- `--global` - Initialize global config instead of project config

### Phase 9: Add config command (optional but recommended)

**File: `internal/commands/config.go`**

Add a new `config` command for managing configuration:

```go
// egenskriven config show          - Show effective (merged) config
// egenskriven config show --global - Show global config only
// egenskriven config show --project - Show project config only
// egenskriven config set <key> <value> --global - Set global config value
// egenskriven config set <key> <value> - Set project config value
// egenskriven config path --global - Show global config path
// egenskriven config path - Show project config path
```

Minimum implementation:

```go
func newConfigCmd(app *pocketbase.PocketBase) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "config",
        Short: "Manage configuration",
    }

    cmd.AddCommand(newConfigShowCmd())
    cmd.AddCommand(newConfigPathCmd())

    return cmd
}

func newConfigShowCmd() *cobra.Command {
    var showGlobal, showProject bool

    cmd := &cobra.Command{
        Use:   "show",
        Short: "Show configuration",
        RunE: func(cmd *cobra.Command, args []string) error {
            if showGlobal {
                cfg, err := config.LoadGlobalConfig()
                if err != nil {
                    return err
                }
                return json.NewEncoder(os.Stdout).Encode(cfg)
            }
            if showProject {
                cfg, err := config.LoadProjectConfig()
                if err != nil {
                    return err
                }
                return json.NewEncoder(os.Stdout).Encode(cfg)
            }
            // Default: show merged config
            cfg, err := config.Load()
            if err != nil {
                return err
            }
            return json.NewEncoder(os.Stdout).Encode(cfg)
        },
    }

    cmd.Flags().BoolVar(&showGlobal, "global", false, "Show global config only")
    cmd.Flags().BoolVar(&showProject, "project", false, "Show project config only")

    return cmd
}

func newConfigPathCmd() *cobra.Command {
    var showGlobal bool

    cmd := &cobra.Command{
        Use:   "path",
        Short: "Show config file path",
        RunE: func(cmd *cobra.Command, args []string) error {
            if showGlobal {
                path, err := config.GlobalConfigPath()
                if err != nil {
                    return err
                }
                fmt.Println(path)
                return nil
            }
            fmt.Println(".egenskriven/config.json")
            return nil
        },
    }

    cmd.Flags().BoolVar(&showGlobal, "global", false, "Show global config path")

    return cmd
}
```

---

## Files to Modify

| File | Changes |
|------|---------|
| `internal/config/config.go` | Add GlobalConfig, LoadGlobalConfig, Load (merged), SaveGlobalConfig |
| `internal/config/config_test.go` | Add tests for new functions |
| `cmd/egenskriven/main.go` | Remove `os.Getenv("EGENSKRIVEN_DIR")`, use config |
| `internal/commands/block.go` | Remove `os.Getenv("EGENSKRIVEN_AGENT")`, use config |
| `internal/commands/comment.go` | Remove `os.Getenv("EGENSKRIVEN_AUTHOR")`, use config |
| `internal/commands/client.go` | Use `config.Load()` instead of `config.LoadProjectConfig()` |
| `internal/commands/root.go` | Register new `config` command |
| `internal/commands/init.go` | Add `--global` flag for global config init |
| `README.md` | Update configuration documentation |

## Files to Create

| File | Purpose |
|------|---------|
| `internal/commands/config.go` | New `config` command (show, path) |

## Code to Remove

1. **main.go:21** - Remove `os.Getenv("EGENSKRIVEN_DIR")` block
2. **block.go:180-185** - Remove `os.Getenv("EGENSKRIVEN_AGENT")` block
3. **comment.go:153-177** - Remove environment variable logic, simplify

---

## Testing Plan

### Unit Tests

1. **config_test.go** - Add tests for:
   - `LoadGlobalConfig` with missing file (returns defaults)
   - `LoadGlobalConfig` with valid file
   - `LoadGlobalConfig` with `~` expansion in data_dir
   - `Load` (merged) with only global config
   - `Load` (merged) with only project config
   - `Load` (merged) with both configs (verify override behavior)
   - `SaveGlobalConfig` creates directory if needed
   - `ValidateResumeMode` validation

2. **Integration test** - Verify:
   - Global config at `~/.config/egenskriven/config.json` is respected
   - Project config overrides global config
   - `data_dir` from global config is used for database

### Manual Testing

```bash
# 1. Create global config
mkdir -p ~/.config/egenskriven
cat > ~/.config/egenskriven/config.json << 'EOF'
{
  "data_dir": "~/.egenskriven",
  "defaults": {
    "author": "testuser",
    "agent": "claude"
  },
  "agent": {
    "workflow": "light",
    "mode": "autonomous",
    "resume_mode": "command"
  }
}
EOF

# 2. Verify data_dir is used
./egenskriven board list  # Should use ~/.egenskriven/data.db

# 3. Verify defaults are used
./egenskriven block TASK-1 "test"  # Should use "claude" as agent
./egenskriven comment TASK-1 "test"  # Should use "testuser" as author

# 4. Test project override
mkdir test-project && cd test-project
mkdir .egenskriven
echo '{"agent":{"workflow":"strict"}}' > .egenskriven/config.json
./egenskriven prime  # Should show "strict" workflow

# 5. Test config command
./egenskriven config show
./egenskriven config show --global
./egenskriven config path --global
```

---

## Migration Guide for Users

### Before (Environment Variables)

```bash
# ~/.bashrc or ~/.zshrc
export EGENSKRIVEN_DIR="$HOME/.egenskriven"
export EGENSKRIVEN_AUTHOR="myname"
export EGENSKRIVEN_AGENT="claude"
```

### After (Config File)

```bash
# Create config directory
mkdir -p ~/.config/egenskriven

# Create config file
cat > ~/.config/egenskriven/config.json << 'EOF'
{
  "data_dir": "~/.egenskriven",
  "defaults": {
    "author": "myname",
    "agent": "claude"
  },
  "agent": {
    "workflow": "light",
    "mode": "autonomous",
    "resume_mode": "command"
  }
}
EOF

# Remove old environment variables from shell profile
# (EGENSKRIVEN_DIR, EGENSKRIVEN_AUTHOR, EGENSKRIVEN_AGENT)
```

---

## Implementation Order

1. **Phase 1-3**: Config types and loading (can be done together)
2. **Phase 4**: Update main.go (depends on Phase 1-3)
3. **Phase 5-6**: Update block.go and comment.go (depends on Phase 1-3)
4. **Phase 7**: Update client.go (depends on Phase 1-3)
5. **Phase 8-9**: Add config command (optional, can be done last)
6. **Update README.md** (after all code changes)
7. **Write tests** (can be done in parallel with implementation)

Each phase should be a separate commit for clean history.

---

## Commit Plan

```
1. feat(config): add GlobalConfig type and LoadGlobalConfig function
2. feat(config): add Load function for merged config
3. refactor(cli): use config.LoadGlobalConfig for data directory
4. refactor(cli): use config for default agent name in block command
5. refactor(cli): use config for default author in comment command
6. refactor(cli): use merged config for API client server URL
7. feat(cli): add config command for showing configuration
8. docs: update README with new configuration system
9. test: add tests for global and merged config loading
```
