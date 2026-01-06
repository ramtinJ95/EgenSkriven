# Global Database Support

Design document for enabling a single EgenSkriven database accessible from any directory on the system.

---

## Table of Contents

1. [Overview](#overview)
2. [Use Cases](#use-cases)
3. [Environment Variable Approach](#environment-variable-approach)
4. [Behavior Specification](#behavior-specification)
5. [Implementation Details](#implementation-details)
6. [Shell Configuration Examples](#shell-configuration-examples)
7. [Testing Considerations](#testing-considerations)
8. [Migration Path](#migration-path)
9. [Future Extensions](#future-extensions)

---

## Overview

### Current Behavior

EgenSkriven currently creates a **per-project database**:

- PocketBase data stored in `pb_data/` in the current working directory
- Each project has its own isolated database
- Boards, tasks, and epics are scoped to that project
- Running `egenskriven serve` from different directories creates separate databases

**Example**:
```
~/project-a/
└── pb_data/          # Database for project-a

~/project-b/
└── pb_data/          # Separate database for project-b
```

### Goal

Enable a **single global database** accessible from any directory:

- All boards visible regardless of current directory
- Unified task management across all projects
- Single server instance serving all work
- Optional per-project databases still supported

### Chosen Approach

**Environment variable**: `EGENSKRIVEN_DATA`

This approach was chosen because:
- Minimal code changes required
- Backward compatible (env var is optional)
- Standard Unix pattern
- Easy to configure in shell profiles
- Works with existing PocketBase infrastructure

---

## Use Cases

### 1. Access All Boards from Any Project

**Scenario**: You have multiple boards (Work, Personal, Open Source) and want to access all of them regardless of which project directory you're in.

**Current limitation**: Each directory has its own database, so boards are isolated.

**With global database**:
```bash
# From any directory
cd ~/random-project
egenskriven board list    # Shows all boards: Work, Personal, Open Source
egenskriven list --board Work --ready
```

### 2. Unified Task Management

**Scenario**: Track all your work in one place while working across multiple repositories.

**With global database**:
```bash
# Working on frontend repo
cd ~/work/frontend
egenskriven add "Fix login bug" --board Work

# Switch to backend repo
cd ~/work/backend
egenskriven add "Add auth endpoint" --board Work --blocked-by <frontend-task>

# Both tasks visible from either location
egenskriven list --board Work
```

### 3. Single Server Instance

**Scenario**: Run one EgenSkriven server that serves your entire workflow.

**With global database**:
```bash
# Start server once (from any directory)
EGENSKRIVEN_DATA=~/.egenskriven/data egenskriven serve

# Access from anywhere
# CLI commands work from any directory
# Web UI shows all boards at http://localhost:8090
```

### 4. Shared Team Database

**Scenario**: Team members access a shared database on a network drive or server.

**With global database**:
```bash
# Point to shared location
export EGENSKRIVEN_DATA=/mnt/team-share/egenskriven/data
egenskriven serve
```

---

## Environment Variable Approach

### Variable Name

```
EGENSKRIVEN_DATA
```

### Value

Absolute path to the directory where PocketBase should store its data.

**Examples**:
```bash
# User's home directory
EGENSKRIVEN_DATA=~/.egenskriven/data

# Absolute path
EGENSKRIVEN_DATA=/home/username/.egenskriven/data

# Custom location
EGENSKRIVEN_DATA=/var/lib/egenskriven

# Network share
EGENSKRIVEN_DATA=/mnt/shared/egenskriven
```

### Recommended Default Location

For users who want a global database, the recommended location is:

```
~/.egenskriven/data/
```

This follows common Unix conventions:
- User-specific data in home directory
- Hidden directory (dot prefix)
- Separate `data/` subdirectory for future config files

**Full structure**:
```
~/.egenskriven/
├── data/                 # PocketBase data (pb_data equivalent)
│   ├── data.db          # SQLite database
│   └── storage/         # File uploads (if any)
└── config.json          # Future: global config file
```

---

## Behavior Specification

### Priority Order

1. If `EGENSKRIVEN_DATA` is set and non-empty: Use that path
2. If not set: Use `pb_data/` in current working directory (existing behavior)

### Detailed Behavior

| EGENSKRIVEN_DATA | Result |
|------------------|--------|
| Not set | `./pb_data/` (current behavior) |
| Empty string (`""`) | `./pb_data/` (current behavior) |
| `~/.egenskriven/data` | `~/.egenskriven/data/` |
| `/absolute/path` | `/absolute/path/` |
| Relative path (`./custom`) | Resolved relative to cwd |

### Directory Creation

- If the specified directory doesn't exist, PocketBase creates it automatically
- Parent directories must exist (e.g., `~/.egenskriven/` must exist for `~/.egenskriven/data/`)
- Consider: Should the command create parent directories? (Recommendation: Yes, with warning)

### Error Handling

| Condition | Behavior |
|-----------|----------|
| Path exists and is a file | Error: "EGENSKRIVEN_DATA must be a directory" |
| Path not writable | Error: "Cannot write to EGENSKRIVEN_DATA: permission denied" |
| Parent doesn't exist | Create parent directories (with info message) |
| Invalid path characters | Error: "Invalid path in EGENSKRIVEN_DATA" |

---

## Implementation Details

### Location

**File**: `cmd/egenskriven/main.go`

### Current Code

```go
func main() {
    app := pocketbase.New()  // Uses default pb_data/
    // ...
}
```

### Modified Code

```go
func main() {
    app := createApp()
    // ...
}

func createApp() *pocketbase.PocketBase {
    dataDir := os.Getenv("EGENSKRIVEN_DATA")
    
    if dataDir == "" {
        // Default behavior: pb_data in current directory
        return pocketbase.New()
    }
    
    // Expand ~ to home directory
    if strings.HasPrefix(dataDir, "~/") {
        home, err := os.UserHomeDir()
        if err == nil {
            dataDir = filepath.Join(home, dataDir[2:])
        }
    }
    
    // Create parent directories if needed
    parentDir := filepath.Dir(dataDir)
    if err := os.MkdirAll(parentDir, 0755); err != nil {
        log.Printf("Warning: Could not create parent directory %s: %v", parentDir, err)
    }
    
    return pocketbase.NewWithConfig(pocketbase.Config{
        DefaultDataDir: dataDir,
    })
}
```

### Key Implementation Notes

1. **Tilde expansion**: Go doesn't automatically expand `~`, so handle it explicitly
2. **Parent directory creation**: Use `os.MkdirAll` to create parent dirs
3. **Existing pattern**: `internal/testutil/testutil.go` already uses `NewWithConfig` successfully
4. **No breaking changes**: Empty or unset env var preserves current behavior

---

## Shell Configuration Examples

### Bash

Add to `~/.bashrc` or `~/.bash_profile`:

```bash
# EgenSkriven global database
export EGENSKRIVEN_DATA="$HOME/.egenskriven/data"
```

Reload:
```bash
source ~/.bashrc
```

### Zsh

Add to `~/.zshrc`:

```zsh
# EgenSkriven global database
export EGENSKRIVEN_DATA="$HOME/.egenskriven/data"
```

Reload:
```bash
source ~/.zshrc
```

### Fish

Add to `~/.config/fish/config.fish`:

```fish
# EgenSkriven global database
set -gx EGENSKRIVEN_DATA "$HOME/.egenskriven/data"
```

Reload:
```bash
source ~/.config/fish/config.fish
```

### PowerShell (Windows)

Add to PowerShell profile:

```powershell
# EgenSkriven global database
$env:EGENSKRIVEN_DATA = "$env:USERPROFILE\.egenskriven\data"
```

### Temporary Usage

For one-off commands without permanent configuration:

```bash
# Single command
EGENSKRIVEN_DATA=~/.egenskriven/data egenskriven list

# Subshell
(export EGENSKRIVEN_DATA=~/.egenskriven/data; egenskriven serve)
```

---

## Testing Considerations

### Existing Test Pattern

The test infrastructure already uses custom data directories:

```go
// internal/testutil/testutil.go
app := pocketbase.NewWithConfig(pocketbase.Config{
    DefaultDataDir: tmpDir,
})
```

This confirms the approach works.

### Test Requirements

1. **Unit tests for env var parsing**
   - Empty string returns default
   - Valid path returns that path
   - Tilde expansion works
   - Invalid paths error appropriately

2. **Integration tests**
   - Server starts with custom data dir
   - CLI commands use correct database
   - Data persists across commands

3. **Existing tests unchanged**
   - Tests use temp directories via testutil
   - No env var set during tests
   - Isolation maintained

### Test Environment

Tests should explicitly NOT set `EGENSKRIVEN_DATA` to avoid interference:

```go
func TestSomething(t *testing.T) {
    // Ensure clean environment
    os.Unsetenv("EGENSKRIVEN_DATA")
    
    // Use testutil for isolated database
    app := testutil.NewTestApp(t)
    // ...
}
```

---

## Migration Path

### For New Users

1. Set environment variable in shell config
2. Start server: `egenskriven serve`
3. Create boards and tasks
4. Access from any directory

### For Existing Users (Per-Project to Global)

**Option A: Start Fresh**
1. Set `EGENSKRIVEN_DATA`
2. Create new boards in global database
3. Recreate tasks or leave old databases

**Option B: Migrate Data**
1. Export from project database:
   ```bash
   cd ~/project-with-data
   egenskriven export --format json > backup.json
   ```

2. Set environment variable:
   ```bash
   export EGENSKRIVEN_DATA=~/.egenskriven/data
   ```

3. Start server and import:
   ```bash
   egenskriven serve &
   egenskriven import backup.json --strategy merge
   ```

4. Repeat for other project databases

### Keeping Per-Project Databases

Users can still use per-project databases by:

1. Not setting `EGENSKRIVEN_DATA` globally
2. Or unsetting it temporarily:
   ```bash
   unset EGENSKRIVEN_DATA
   cd ~/isolated-project
   egenskriven serve  # Uses ./pb_data/
   ```

---

## Future Extensions

### 1. `--data` Command Line Flag

Higher priority than env var, allows per-command override:

```bash
# Use specific database for this command
egenskriven --data ~/.egenskriven/data list

# Override env var
EGENSKRIVEN_DATA=/default/path egenskriven --data /other/path list
# Uses /other/path
```

**Priority order** (highest first):
1. `--data` flag
2. `EGENSKRIVEN_DATA` env var
3. Default (`./pb_data/`)

**Implementation**: Add to `internal/commands/root.go`:
```go
var dataDir string
rootCmd.PersistentFlags().StringVar(&dataDir, "data", "",
    "Path to data directory (overrides EGENSKRIVEN_DATA)")
```

### 2. Global Config File

Store default settings in `~/.egenskriven/config.json`:

```json
{
  "data_dir": "~/.egenskriven/data",
  "default_board": "Work",
  "server": {
    "port": 8090
  }
}
```

**Priority order** (highest first):
1. `--data` flag
2. `EGENSKRIVEN_DATA` env var
3. `~/.egenskriven/config.json` data_dir
4. Default (`./pb_data/`)

### 3. Per-Project Override

Allow `.egenskriven/config.json` in a project to override global:

```json
{
  "data_dir": "./pb_data",  // Use local database for this project
  "default_board": "ProjectSpecific"
}
```

**Use case**: Most work uses global database, but one project needs isolation.

### 4. Database Selection UI

In the web UI, show which database is active:

- Display path in header/footer
- Warning if using non-default location
- Easy way to see/change in settings

---

## Summary

| Aspect | Specification |
|--------|---------------|
| **Variable** | `EGENSKRIVEN_DATA` |
| **Default** | `./pb_data/` (no change) |
| **Recommended** | `~/.egenskriven/data` |
| **Implementation** | `cmd/egenskriven/main.go` |
| **Pattern** | `pocketbase.NewWithConfig()` |
| **Breaking changes** | None |
| **Future** | `--data` flag, global config |

---

## Implementation Checklist

- [ ] Add env var check in `main.go`
- [ ] Implement tilde expansion
- [ ] Create parent directories if needed
- [ ] Add error handling for invalid paths
- [ ] Update README with env var documentation
- [ ] Add shell configuration examples to docs
- [ ] Write unit tests for env var parsing
- [ ] Write integration test with custom data dir
- [ ] Document migration path for existing users

---

*This document describes the first phase of global database support. The `--data` flag and global config file are planned as future extensions.*
