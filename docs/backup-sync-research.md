# Backup and Sync Feature Research

> **Date**: January 2026
> **Status**: Research complete, implementation pending
> **Goal**: Enable easy data portability between machines via export/import commands

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Current State Analysis](#current-state-analysis)
3. [Option 1: Enhanced Export/Import](#option-1-enhanced-exportimport-commands)
4. [Option 2: Direct Database Copy](#option-2-direct-database-copy)
5. [Option 3: Git-based Sync](#option-3-git-based-sync)
6. [Option 4: PocketBase S3/Cloud Backup](#option-4-pocketbase-s3cloud-backup)
7. [Option 5: Custom Portable Format (.egs)](#option-5-custom-portable-format-egs)
8. [Data Model Reference](#data-model-reference)
9. [Implementation Patterns](#implementation-patterns)
10. [Recommendations](#recommendations)

---

## Executive Summary

### Problem Statement
Users need to sync EgenSkriven data (projects, tasks, boards, etc.) across different machines easily.

### Proposed Solutions
Five approaches were researched:

| Option | Approach | Effort | Best For |
|--------|----------|--------|----------|
| 1 | Enhanced Export/Import | Low | Quick fix, human-readable backups |
| 2 | Direct Database Copy | Low | Complete backups, disaster recovery |
| 3 | Git-based Sync | High | Version history, collaboration |
| 4 | S3/Cloud Backup | Config only | Automated cloud redundancy |
| 5 | Custom .egs Format | Medium | Portable single-file with conflict detection |

### Recommended Approach
**Hybrid**: Fix Option 1 + Add Option 2 for Phase 1, then consider Option 5 for advanced sync.

---

## Current State Analysis

### Existing Commands

EgenSkriven already has three data-related commands:

#### `egenskriven backup`
**File**: `internal/commands/backup.go`

- Creates timestamped SQLite database file copies
- Supports custom filename and output directory
- Lists existing backups with `--list` flag
- Direct file copy (preserves all data)

```bash
egenskriven backup                      # Auto-timestamped
egenskriven backup my-backup.db         # Custom name
egenskriven backup -o /path/to/backup   # Custom directory
egenskriven backup --list               # List backups
```

#### `egenskriven export`
**File**: `internal/commands/export.go`

- Exports to JSON (complete) or CSV (tasks only)
- Board filtering support
- Output to file or stdout

```bash
egenskriven export                      # JSON to stdout
egenskriven export --format json -o backup.json
egenskriven export --format csv -o tasks.csv
egenskriven export --board work
```

#### `egenskriven import`
**File**: `internal/commands/import.go`

- Import from JSON files
- Merge (skip existing) or Replace strategies
- Dry-run preview support

```bash
egenskriven import backup.json
egenskriven import backup.json --strategy replace
egenskriven import backup.json --dry-run
```

### Database Structure

**Location**: `pb_data/data.db` (SQLite via PocketBase)

**Collections**:
- `tasks` - Core task data
- `boards` - Board configurations
- `epics` - Epic groupings
- `comments` - Task discussions
- `sessions` - AI agent session tracking
- `views` - Saved filter configurations

---

## Option 1: Enhanced Export/Import Commands

### Current Export Data Structure

```go
type ExportData struct {
    Version  string        `json:"version"`  // "1.0"
    Exported string        `json:"exported"` // RFC3339 timestamp
    Boards   []ExportBoard `json:"boards"`
    Epics    []ExportEpic  `json:"epics"`
    Tasks    []ExportTask  `json:"tasks"`
}
```

### What's Currently Exported vs Missing

| Collection | Exported Fields | Missing Fields |
|------------|-----------------|----------------|
| **Tasks** | id, title, description, type, priority, column, position, board, epic, parent, labels, blocked_by, due_date, created_by, created, updated | `seq`, `history`, `created_by_agent`, `agent_session` |
| **Boards** | id, name, prefix, columns, color | `next_seq` (critical!), `resume_mode` |
| **Epics** | id, title, description, color | `board` relation (orphaned on import!) |
| **Comments** | — | Entire collection not exported |
| **Sessions** | — | Entire collection not exported |
| **Views** | — | Entire collection not exported |

### Critical Bug: Board Column Defaults Mismatch

```go
// Import defaults (backup.go):
[]string{"backlog", "todo", "in_progress", "review", "done"}

// Application defaults (board/board.go):
[]string{"backlog", "todo", "in_progress", "need_input", "review", "done"}

// MISSING: "need_input" column in import!
```

### Critical Bug: Missing `next_seq` Counter

The `next_seq` field on boards is not exported. This counter generates unique task sequence numbers (e.g., WRK-42). Without it:
- Importing to a fresh database starts from 1
- Could create duplicate task IDs if original board continues

### Import Strategy Limitations

**Merge Strategy**:
- Skips existing records by ID
- Creates new records
- No field-level merge

**Replace Strategy**:
- Overwrites entire records
- Empty values in export CLEAR fields (destructive!)
- No conflict detection

### Enhancements Needed

1. **Add missing task fields**: `seq`, `history`, `created_by_agent`, `agent_session`
2. **Add missing board fields**: `next_seq`, `resume_mode`
3. **Fix epic export**: Include `board` relation
4. **Add collections**: comments, sessions, views
5. **Fix column defaults**: Match application defaults
6. **Add conflict detection**: Compare `updated` timestamps
7. **Add schema version**: For future migrations

### Estimated Effort: 2-4 hours

### Pros
- Already 90% implemented
- Human-readable JSON output
- Dry-run support exists
- Merge/replace strategies exist

### Cons
- No conflict detection
- No incremental sync
- Manual file transfer required

---

## Option 2: Direct Database Copy

### Recommended Method: VACUUM INTO

`VACUUM INTO` is the safest method for online SQLite backup:

```go
func BackupDatabase(db *sql.DB, backupPath string) error {
    // Remove existing file (VACUUM INTO requires non-existent destination)
    os.Remove(backupPath)

    // Escape path for SQL injection prevention
    safePath := strings.ReplaceAll(backupPath, "'", "''")

    // Create consistent snapshot
    _, err := db.Exec(fmt.Sprintf("VACUUM INTO '%s'", safePath))
    if err != nil {
        return fmt.Errorf("VACUUM INTO failed: %w", err)
    }

    // Verify integrity
    return VerifyBackup(backupPath)
}

func VerifyBackup(backupPath string) error {
    backupDB, err := sql.Open("sqlite3", backupPath)
    if err != nil {
        return err
    }
    defer backupDB.Close()

    var result string
    if err := backupDB.QueryRow("PRAGMA integrity_check(1)").Scan(&result); err != nil {
        return err
    }
    if result != "ok" {
        os.Remove(backupPath) // Delete corrupt backup
        return fmt.Errorf("integrity check failed: %s", result)
    }
    return nil
}
```

### Why VACUUM INTO Over File Copy

| Method | Online Safe | Handles WAL | Consistent | Defragments |
|--------|-------------|-------------|------------|-------------|
| File copy | ❌ | ❌ | ❌ | ❌ |
| `.backup` command | ✅ | ✅ | ✅ | ❌ |
| `VACUUM INTO` | ✅ | ✅ | ✅ | ✅ |

### WAL Mode Considerations

When SQLite is in WAL (Write-Ahead Logging) mode:

```
database.db      # Main database file
database.db-wal  # Write-ahead log (uncommitted changes)
database.db-shm  # Shared memory file (NEVER copy this)
```

**VACUUM INTO** consolidates WAL into the backup automatically.

### pb_data Directory Structure

```
pb_data/
├── data.db          ✅ REQUIRED - Main database
├── data.db-wal      ❌ NOT NEEDED with VACUUM INTO
├── data.db-shm      ❌ NEVER COPY - Memory mapped, regenerated
├── storage/         ✅ REQUIRED - Uploaded files
│   └── {collection_id}/
│       └── {record_id}/
│           └── {filename}
├── backups/         ❌ EXCLUDE - Local backups
└── auxiliary.db     ⚠️ OPTIONAL - Logs and temp data
```

### Compression (Optional)

```go
func CreateCompressedBackup(pbDataDir, destPath string) error {
    // Create tar.gz archive
    file, err := os.Create(destPath)
    if err != nil {
        return err
    }
    defer file.Close()

    gzWriter := gzip.NewWriter(file)
    defer gzWriter.Close()

    tarWriter := tar.NewWriter(gzWriter)
    defer tarWriter.Close()

    // Add data.db (via VACUUM INTO to temp file first)
    // Add storage/ directory
    // ...
}
```

### Estimated Effort: 1-2 hours

### Pros
- Zero data loss (complete database)
- Simple implementation
- Works while app is running
- Includes everything (comments, sessions, history)
- Established patterns from many projects

### Cons
- Binary format (can't diff/merge)
- Larger file size than JSON
- Version coupling (same PocketBase version needed)
- No selective export

---

## Option 3: Git-based Sync

### Pattern 1: Password-Store Style (Simple)

From `zx2c4/password-store`:

```bash
# Initialize git in data directory
egenskriven git init

# Auto-commit on every change
egenskriven git add && egenskriven git commit -m "Update tasks"

# Manual sync
egenskriven git push
egenskriven git pull
```

```go
func gitCommit(repoPath, message string) error {
    cmd := exec.Command("git", "-C", repoPath, "add", "-A")
    if err := cmd.Run(); err != nil {
        return err
    }

    // Check if there are changes
    status := exec.Command("git", "-C", repoPath, "status", "--porcelain")
    output, _ := status.Output()
    if len(output) == 0 {
        return nil // Nothing to commit
    }

    cmd = exec.Command("git", "-C", repoPath, "commit", "-m", message)
    return cmd.Run()
}
```

### Pattern 2: Chezmoi Style (Configurable)

```toml
# .egenskriven/config.toml
[git]
enabled = true
autoAdd = true      # Auto-stage changes
autoCommit = true   # Auto-commit after changes
autoPush = false    # Manual push (safer)
commitMessageTemplate = "{{.action}}: {{.summary}}"
```

### Conflict Resolution Strategies

#### 1. Git's merge=union (Line-Based)

For append-only or line-based files:

```gitattributes
# .gitattributes
tasks.txt merge=union
```

#### 2. Last-Write-Wins

Simple but may lose data:

```go
func mergeLastWriteWins(local, remote *Task) *Task {
    if local.UpdatedAt.After(remote.UpdatedAt) {
        return local
    }
    return remote
}
```

#### 3. Three-Way Merge (Sophisticated)

From `jupyter/nbdime`:

```go
func threeWayMerge(base, local, remote *ExportData) (*ExportData, []Conflict) {
    localDiff := diff(base, local)
    remoteDiff := diff(base, remote)

    // Apply non-conflicting changes
    // Detect conflicts where both modified same field
    // ...
}
```

#### 4. Conflict Files (Obsidian Style)

```go
// When conflict detected, create:
// task.sync-conflict-20260115-143022.json
func createConflictFile(original string, conflictData []byte) string {
    timestamp := time.Now().Format("20060102-150405")
    ext := filepath.Ext(original)
    base := strings.TrimSuffix(original, ext)
    return fmt.Sprintf("%s.sync-conflict-%s%s", base, timestamp, ext)
}
```

### Custom Merge Driver for Tasks

```bash
# Register custom merge driver
git config merge.egenskriven-json.name "EgenSkriven JSON merger"
git config merge.egenskriven-json.driver "egenskriven merge-json %O %A %B"
```

```go
// egenskriven merge-json implementation
func mergeJSON(basePath, oursPath, theirsPath string) error {
    base := loadExport(basePath)
    ours := loadExport(oursPath)
    theirs := loadExport(theirsPath)

    merged, conflicts := threeWayMerge(base, ours, theirs)

    if len(conflicts) > 0 {
        // Write conflict markers or create conflict files
        return fmt.Errorf("conflicts detected: %d", len(conflicts))
    }

    return writeExport(oursPath, merged)
}
```

### Recommended Pull Strategy

```bash
# From chezmoi and oh-my-zsh patterns
git pull --rebase --autostash
```

This:
1. Stashes local uncommitted changes
2. Rebases local commits on top of remote
3. Re-applies stashed changes

### Estimated Effort: 8-16 hours (High)

### Pros
- Full version history
- Works with any git remote
- Leverage existing git knowledge
- Can review changes before sync
- Collaboration-ready

### Cons
- Requires git setup on each machine
- Conflict resolution can be confusing
- Not suitable for binary attachments
- Learning curve for non-git users
- Need to export to file format git can track

---

## Option 4: PocketBase S3/Cloud Backup

### Native PocketBase S3 Configuration

PocketBase has built-in S3 backup support:

```bash
curl -X PATCH http://localhost:8090/api/settings \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "backups": {
      "cron": "0 2 * * *",
      "cronMaxKeep": 7,
      "s3": {
        "enabled": true,
        "bucket": "my-pocketbase-backups",
        "region": "us-west-004",
        "endpoint": "s3.us-west-004.backblazeb2.com",
        "accessKey": "'$ACCESS_KEY'",
        "secret": "'$SECRET_KEY'",
        "forcePathStyle": false
      }
    }
  }'
```

### API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/backups` | GET | List all backups |
| `/api/backups` | POST | Create new backup |
| `/api/backups/{key}` | GET | Download backup |
| `/api/backups/{key}` | DELETE | Delete backup |
| `/api/backups/{key}/restore` | POST | Restore from backup |
| `/api/backups/upload` | POST | Upload backup file |

All endpoints require superuser authentication.

### Cron Expression Examples

| Expression | Schedule |
|------------|----------|
| `0 0 * * *` | Daily at midnight |
| `0 2 * * *` | Daily at 2 AM |
| `0 0 * * 0` | Weekly on Sunday |
| `0 0 1 * *` | Monthly on 1st |

### S3-Compatible Providers

| Provider | Endpoint Format | Notes |
|----------|-----------------|-------|
| AWS S3 | `s3.{region}.amazonaws.com` | Standard |
| Backblaze B2 | `s3.{region}.backblazeb2.com` | Lowest cost |
| Cloudflare R2 | `{account-id}.r2.cloudflarestorage.com` | Zero egress |
| DigitalOcean Spaces | `{region}.digitaloceanspaces.com` | Simple |
| MinIO | `http://localhost:9000` | Self-hosted |
| Wasabi | `s3.{region}.wasabisys.com` | 90-day minimum |

### Cost Comparison (10TB storage, 1TB egress/month)

| Provider | Storage/TB/mo | Egress | Monthly Total |
|----------|---------------|--------|---------------|
| **Cloudflare R2** | $15 | Free | ~$150 |
| **Backblaze B2** | $6 | 3x free, then $0.01/GB | ~$60 |
| **Wasabi** | $6.99 | Free | ~$70 |
| **AWS S3** | $23 | $0.09/GB | ~$320 |

### Restore Limitations

From PocketBase source code:

```go
// RestoreBackup is experimental and currently works only on UNIX based systems.
// Requires free disk space of at least 2x the size of the backup.
```

**Critical Limitations**:
- Unix-only (uses `execve` for process restart)
- Requires 2x backup size in free disk space
- May fail with custom network mounts
- S3 collection files NOT included in backup

### Alternative: rclone for Any Cloud

```bash
#!/bin/bash
# backup-to-cloud.sh

# Create local backup first
egenskriven backup -o /tmp/pb_backup/

# Sync to encrypted remote
rclone sync /tmp/pb_backup/ encrypted-remote:egenskriven/

# Cleanup
rm -rf /tmp/pb_backup/
```

### Estimated Effort: Configuration only (< 1 hour)

### Pros
- Automatic scheduled backups
- Cloud redundancy
- Native PocketBase integration
- No code changes needed
- Supports multiple providers

### Cons
- Requires S3 account setup
- Internet dependency
- Restore limitations on Windows/non-Unix
- Not a true "sync" - one-way backup
- S3 collection files not included

---

## Option 5: Custom Portable Format (.egs)

### Proposed File Format

```go
type EgsExportFormat struct {
    FormatVersion uint              `json:"format_version"`
    AppVersion    string            `json:"app_version"`
    ExportedAt    time.Time         `json:"exported_at"`
    SourceInfo    ExportSourceInfo  `json:"source"`
    Checksum      string            `json:"checksum,omitempty"`
    Data          ExportData        `json:"data"`
}

type ExportSourceInfo struct {
    Hostname    string `json:"hostname"`
    Username    string `json:"username"`
    BoardPath   string `json:"board_path"`
    ExportType  string `json:"export_type"` // "full", "incremental", "selective"
}

type ExportData struct {
    Boards     []BoardExport     `json:"boards"`
    Epics      []EpicExport      `json:"epics"`
    Tasks      []TaskExport      `json:"tasks"`
    Comments   []CommentExport   `json:"comments,omitempty"`
    Sessions   []SessionExport   `json:"sessions,omitempty"`
    Views      []ViewExport      `json:"views,omitempty"`
    Statistics ExportStatistics  `json:"statistics"`
}
```

### File Extension and Detection

```go
const (
    EgsFileExtension     = ".egs"
    CurrentFormatVersion = 1
)

// Detect format by magic bytes
func DetectFormat(filename string) (string, error) {
    f, err := os.Open(filename)
    if err != nil {
        return "", err
    }
    defer f.Close()

    header := make([]byte, 2)
    io.ReadFull(f, header)

    // Gzip magic: 0x1f 0x8b
    if header[0] == 0x1f && header[1] == 0x8b {
        return "gzip", nil
    }

    // JSON starts with { or [
    if header[0] == '{' || header[0] == '[' {
        return "json", nil
    }

    return "unknown", nil
}
```

### Atomic File Operations

From Tailscale's atomicfile pattern:

```go
func WriteFileAtomic(filename string, data []byte, perm os.FileMode) error {
    // Create temp file in same directory (ensures same filesystem)
    dir := filepath.Dir(filename)
    f, err := os.CreateTemp(dir, ".tmp-")
    if err != nil {
        return err
    }
    tmpName := f.Name()

    // Cleanup on failure
    defer func() {
        if err != nil {
            os.Remove(tmpName)
        }
    }()

    // Write data
    if _, err = f.Write(data); err != nil {
        f.Close()
        return err
    }

    // Set permissions
    if err = f.Chmod(perm); err != nil {
        f.Close()
        return err
    }

    // Sync to disk
    if err = f.Sync(); err != nil {
        f.Close()
        return err
    }

    // Close before rename
    if err = f.Close(); err != nil {
        return err
    }

    // Atomic rename
    return os.Rename(tmpName, filename)
}
```

### Compression (gzip)

```go
func ExportCompressed(filename string, export *EgsExportFormat) error {
    tmpFile, err := os.CreateTemp(filepath.Dir(filename), ".egs-*")
    if err != nil {
        return err
    }
    tmpName := tmpFile.Name()
    defer func() {
        if err != nil {
            os.Remove(tmpName)
        }
    }()

    gw := gzip.NewWriter(tmpFile)
    enc := json.NewEncoder(gw)
    enc.SetIndent("", "  ")

    if err = enc.Encode(export); err != nil {
        gw.Close()
        tmpFile.Close()
        return err
    }

    if err = gw.Close(); err != nil {
        tmpFile.Close()
        return err
    }

    if err = tmpFile.Sync(); err != nil {
        tmpFile.Close()
        return err
    }

    if err = tmpFile.Close(); err != nil {
        return err
    }

    return os.Rename(tmpName, filename)
}
```

### Checksum Verification

```go
import (
    "crypto/sha256"
    "encoding/hex"
)

func ComputeChecksum(data []byte) string {
    hash := sha256.Sum256(data)
    return hex.EncodeToString(hash[:])
}

func VerifyChecksum(export *EgsExportFormat, dataBytes []byte) error {
    if export.Checksum == "" {
        return nil
    }

    computed := ComputeChecksum(dataBytes)
    if computed != export.Checksum {
        return fmt.Errorf("checksum mismatch: expected %s, got %s",
            export.Checksum, computed)
    }
    return nil
}
```

### Conflict Detection with Lamport Clocks

```go
type LamportClock struct {
    counter uint64
}

func (l *LamportClock) Time() uint64 {
    return atomic.LoadUint64(&l.counter)
}

func (l *LamportClock) Increment() uint64 {
    return atomic.AddUint64(&l.counter, 1)
}

func (l *LamportClock) Witness(v uint64) {
    for {
        cur := atomic.LoadUint64(&l.counter)
        if v <= cur {
            return
        }
        if atomic.CompareAndSwapUint64(&l.counter, cur, v+1) {
            return
        }
    }
}

// Task version for conflict detection
type TaskVersion struct {
    ID          string    `json:"id"`
    UpdatedAt   time.Time `json:"updated_at"`
    LamportTime uint64    `json:"lamport_time"`
    ContentHash string    `json:"content_hash"`
}
```

### Merge Strategies

```go
type MergeStrategy string

const (
    MergeStrategyLocalWins  MergeStrategy = "local_wins"
    MergeStrategyRemoteWins MergeStrategy = "remote_wins"
    MergeStrategyNewest     MergeStrategy = "newest"
    MergeStrategyManual     MergeStrategy = "manual"
)

func ResolveConflict(conflict Conflict, strategy MergeStrategy) (*TaskExport, error) {
    switch strategy {
    case MergeStrategyLocalWins:
        return &conflict.LocalTask, nil
    case MergeStrategyRemoteWins:
        return &conflict.RemoteTask, nil
    case MergeStrategyNewest:
        if conflict.LocalTask.UpdatedAt.After(conflict.RemoteTask.UpdatedAt) {
            return &conflict.LocalTask, nil
        }
        return &conflict.RemoteTask, nil
    case MergeStrategyManual:
        return nil, fmt.Errorf("manual resolution required")
    }
    return nil, fmt.Errorf("unknown strategy")
}
```

### Version Migration

```go
type Migrator struct {
    migrations map[uint]MigrationFunc
}

type MigrationFunc func(data json.RawMessage) (json.RawMessage, error)

func (m *Migrator) Migrate(data []byte) (*EgsExportFormat, error) {
    var header struct {
        FormatVersion uint `json:"format_version"`
    }
    if err := json.Unmarshal(data, &header); err != nil {
        header.FormatVersion = 0 // Legacy format
    }

    raw := json.RawMessage(data)

    // Apply migrations sequentially
    for v := header.FormatVersion; v < CurrentFormatVersion; v++ {
        migration, ok := m.migrations[v+1]
        if !ok {
            return nil, fmt.Errorf("no migration for v%d to v%d", v, v+1)
        }

        var err error
        raw, err = migration(raw)
        if err != nil {
            return nil, err
        }
    }

    var export EgsExportFormat
    return &export, json.Unmarshal(raw, &export)
}
```

### CLI Commands

```bash
# Export
egenskriven db-export backup.egs
egenskriven db-export backup.egs --format=json --pretty
egenskriven db-export backup.egs --include-archived

# Import
egenskriven db-import backup.egs --dry-run
egenskriven db-import backup.egs --merge=newest
egenskriven db-import backup.egs --force --backup
```

### Import Preview (Dry Run)

```go
type ImportPreview struct {
    SourceFile    string
    ExportedAt    time.Time
    TotalTasks    int
    NewTasks      int
    UpdatedTasks  int
    SkippedTasks  int
    ConflictTasks int
    Conflicts     []Conflict
}

func PreviewImport(filename string, board *Board) (*ImportPreview, error) {
    export, err := ImportCompressed(filename)
    if err != nil {
        return nil, err
    }

    preview := &ImportPreview{
        SourceFile: filename,
        ExportedAt: export.ExportedAt,
        TotalTasks: len(export.Data.Tasks),
    }

    // Analyze each task...
    return preview, nil
}

func PrintPreview(preview *ImportPreview) {
    fmt.Printf("\nImport Preview\n")
    fmt.Printf("==============\n")
    fmt.Printf("Source:     %s\n", preview.SourceFile)
    fmt.Printf("Exported:   %s\n", preview.ExportedAt.Format(time.RFC3339))
    fmt.Printf("\n")
    fmt.Printf("Tasks:      %d total\n", preview.TotalTasks)
    fmt.Printf("  New:      %d\n", preview.NewTasks)
    fmt.Printf("  Updated:  %d\n", preview.UpdatedTasks)
    fmt.Printf("  Skipped:  %d (unchanged)\n", preview.SkippedTasks)
    fmt.Printf("  Conflicts: %d\n", preview.ConflictTasks)
}
```

### Estimated Effort: 4-8 hours

### Pros
- Single portable file
- Human-inspectable (JSON inside)
- Version migration support
- Conflict detection built-in
- Dry-run preview
- Works offline
- Compression reduces size

### Cons
- New code to write and maintain
- Need to design format carefully
- No automatic sync (manual transfer)

---

## Data Model Reference

### Tasks Collection

| Field | Type | Required | Sync Notes |
|-------|------|----------|------------|
| id | UUID | Yes | Primary key |
| title | Text (500) | Yes | |
| description | Text (10000) | No | |
| type | Select | Yes | bug, feature, chore |
| priority | Select | Yes | low, medium, high, urgent |
| column | Select | Yes | backlog, todo, in_progress, need_input, review, done |
| position | Number | Yes | Fractional ordering |
| board | Relation | No | References boards |
| epic | Relation | No | References epics |
| parent | Relation | No | Self-reference for subtasks |
| labels | JSON array | No | |
| blocked_by | JSON array | No | Task IDs |
| due_date | Date | No | |
| created_by | Select | Yes | user, agent, cli |
| created_by_agent | Text (100) | No | Agent identifier |
| agent_session | JSON | No | Current session metadata |
| history | JSON array | No | Audit trail |
| seq | Number | No | Board-specific sequence |
| created | DateTime | Yes | Auto-generated |
| updated | DateTime | Yes | Auto-updated |

### Boards Collection

| Field | Type | Required | Sync Notes |
|-------|------|----------|------------|
| id | UUID | Yes | Primary key |
| name | Text (100) | Yes | |
| prefix | Text (10) | Yes | Unique, e.g., "WRK" |
| columns | JSON array | No | Custom workflow columns |
| color | Text (7) | No | Hex color |
| next_seq | Number | No | **Critical for sync!** |
| resume_mode | Select | No | manual, command, auto |

### Epics Collection

| Field | Type | Required | Sync Notes |
|-------|------|----------|------------|
| id | UUID | Yes | Primary key |
| title | Text (200) | Yes | |
| description | Text (5000) | No | |
| color | Text (7) | No | |
| board | Relation | Yes | **Currently not exported!** |
| created | DateTime | Yes | |
| updated | DateTime | Yes | |

### Comments Collection

| Field | Type | Required | Sync Notes |
|-------|------|----------|------------|
| id | UUID | Yes | Primary key |
| task | Relation | Yes | Cascade delete |
| content | Text (50000) | Yes | |
| author_type | Select | Yes | human, agent |
| author_id | Text (200) | No | |
| metadata | JSON | No | Session refs, mentions |
| created | DateTime | Yes | |
| updated | DateTime | Yes | |

### Sessions Collection

| Field | Type | Required | Sync Notes |
|-------|------|----------|------------|
| id | UUID | Yes | Primary key |
| task | Relation | Yes | Cascade delete |
| tool | Select | Yes | opencode, claude-code, codex |
| external_ref | Text (500) | Yes | Session ID from AI tool |
| ref_type | Select | Yes | uuid, path |
| working_dir | Text (1000) | Yes | **Machine-specific!** |
| status | Select | Yes | active, paused, completed, abandoned |
| created | DateTime | Yes | |
| updated | DateTime | Yes | |
| ended_at | Date | No | |

### Views Collection

| Field | Type | Required | Sync Notes |
|-------|------|----------|------------|
| id | UUID | Yes | Primary key |
| board | Relation | Yes | |
| name | Text (100) | Yes | |
| filters | JSON array | No | |
| display | JSON | No | viewMode, density, etc. |
| match_mode | Select | Yes | all, any |
| is_favorite | Boolean | No | |

### Machine-Specific vs Shareable Data

**DO NOT SYNC (Machine-specific)**:
- `sessions.working_dir` - Absolute paths differ per machine
- `~/.config/egenskriven/config.json` - User preferences
- `/tmp/egenskriven-tui-session.json` - Ephemeral UI state

**SHOULD SYNC**:
- All task/board/epic data
- Comments (with relative session references)
- Views and filters
- Project config (`.egenskriven/config.json`)

---

## Implementation Patterns

### From Real Projects

#### Tailscale - Atomic File Writes
```go
// temp file → sync → rename pattern
```

#### k0sproject/k0s - VACUUM INTO
```go
func (db *sqliteDB) Backup(path string) error {
    _, err := db.Exec("VACUUM INTO ?", path)
    return err
}
```

#### Password-store - Git Integration
```bash
# Simple pass-through for git commands
git -C "$DATA_DIR" "$@"
```

#### Chezmoi - Configurable Auto-Sync
```toml
[git]
autoAdd = true
autoCommit = true
autoPush = false
```

#### nbdime - Three-Way JSON Merge
```python
def decide_merge(base, local, remote, strategies=None):
    local_diff = diff(base, local)
    remote_diff = diff(base, remote)
    return decide_merge_with_diff(base, local, remote, local_diff, remote_diff, strategies)
```

#### HashiCorp Serf - Lamport Clocks
```go
type LamportClock struct {
    counter uint64
}
```

---

## Recommendations

### Phase 1: Quick Wins (2-4 hours)

1. **Fix existing export/import**:
   - Add missing fields: `seq`, `history`, `created_by_agent`, `next_seq`
   - Fix epic→board relation export
   - Fix board column defaults

2. **Add `db-copy` command**:
   - Use VACUUM INTO for safe online backup
   - Include integrity verification
   - Optional compression

### Phase 2: Enhanced Sync (4-8 hours)

Choose one based on user preference:

- **Option 5 (.egs format)** if you want:
  - Portable single-file backups
  - Built-in conflict detection
  - Version migration support

- **Option 3 (Git sync)** if you want:
  - Full version history
  - Collaboration support
  - Leverage existing git knowledge

### Phase 3: Automation (Optional)

- **Option 4 (S3 backup)** for:
  - Automated cloud backups
  - Zero-maintenance redundancy
  - Disaster recovery

---

## Comparison Matrix

| Criteria | Option 1 | Option 2 | Option 3 | Option 4 | Option 5 |
|----------|----------|----------|----------|----------|----------|
| **Effort** | Low | Low | High | Config | Medium |
| **Data completeness** | Partial | 100% | Depends | 100% | 100% |
| **Human readable** | ✅ JSON | ❌ Binary | ✅ JSON | ❌ Binary | ✅ JSON |
| **Conflict handling** | ❌ | ❌ | ✅ Git | ❌ | ✅ Built-in |
| **Version history** | ❌ | ❌ | ✅ Full | ❌ | ❌ |
| **Offline capable** | ✅ | ✅ | ✅ | ❌ | ✅ |
| **Auto sync** | ❌ | ❌ | ✅ | ✅ Cron | ❌ |
| **Self-hosted** | ✅ | ✅ | ✅ | ❌ | ✅ |
| **File size** | Small | Large | Small | Large | Small |

---

## References

### PocketBase Documentation
- Backup API: https://pocketbase.io/docs/api-backups
- Settings: https://pocketbase.io/docs/api-settings
- Going to Production: https://pocketbase.io/docs/going-to-production

### SQLite Documentation
- VACUUM INTO: https://sqlite.org/lang_vacuum.html
- WAL Mode: https://sqlite.org/wal.html

### Git Sync Projects
- Password-store: https://github.com/zx2c4/password-store
- Chezmoi: https://github.com/twpayne/chezmoi
- Obsidian Git: https://github.com/Vinzent03/obsidian-git

### Task Management Tools
- Taskwarrior Sync RFC: https://github.com/gothenburgbitfactory/taskwarrior/blob/develop/doc/devel/rfcs/sync.md
- Linear Export: https://linear.app/docs/exporting-data

### Implementation Patterns
- Tailscale atomicfile: https://github.com/tailscale/tailscale
- nbdime (JSON merge): https://github.com/jupyter/nbdime
- HashiCorp Serf (Lamport clocks): https://github.com/hashicorp/serf
