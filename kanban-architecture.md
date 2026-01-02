# EgenSkriven: PocketBase + React + CLI Architecture

## Overview

A local-first kanban board with CLI-first design. The CLI is designed to be agent-friendly - any AI coding assistant can invoke it directly without needing a special "AI layer". PocketBase provides the infrastructure: SQLite database, REST API, real-time subscriptions via SSE, and admin UI. The React frontend is embedded in the binary for single-file distribution.

```
┌─────────────────────────────────────────────────────────────────┐
│                     Single Go Binary                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────────────┐│
│  │   CLI        │  │  PocketBase  │  │   Embedded React UI   ││
│  │   Commands   │  │    Core      │  │   (go:embed)          ││
│  │              │  │              │  │                        ││
│  │  • add       │  │  • REST API  │  │  • Kanban board       ││
│  │  • move      │  │  • Realtime  │  │  • Real-time sync     ││
│  │  • list      │  │  • Auth      │  │  • Drag & drop        ││
│  │  • update    │  │  • Admin UI  │  │                        ││
│  │  • delete    │  │              │  │                        ││
│  └──────┬───────┘  └──────┬───────┘  └───────────┬────────────┘│
│         │                 │                       │             │
│         └─────────────────┼───────────────────────┘             │
│                           │                                     │
│                    ┌──────▼──────┐                              │
│                    │   SQLite    │                              │
│                    │  (pb_data)  │                              │
│                    └─────────────┘                              │
└─────────────────────────────────────────────────────────────────┘
```

## Design Principles

### Agent-Friendly CLI

The CLI is designed so that AI agents (Claude, GPT, Cursor, etc.) can use it directly without any special integration layer:

1. **Structured output**: All commands support `--json` for machine-readable output
2. **Flexible identification**: Tasks can be referenced by ID or partial title match
3. **Batch operations**: Commands accept multiple items via stdin or file
4. **Rich filtering**: Precise queries reduce context needed by agents
5. **Idempotent operations**: `--id` flag allows safe retries
6. **Clear errors**: Error messages include enough context for agents to self-correct

### Human-First Defaults

Despite being agent-friendly, the CLI defaults to human-readable output. Agents can opt into JSON mode.

## Project Structure

```
egenskriven/
├── cmd/
│   └── egenskriven/
│       └── main.go              # Entry point, registers CLI commands
├── internal/
│   ├── commands/
│   │   ├── root.go              # Root command, global flags
│   │   ├── add.go               # kanban add "task title"
│   │   ├── move.go              # kanban move <task> <column>
│   │   ├── update.go            # kanban update <task> [fields]
│   │   ├── delete.go            # kanban delete <task>
│   │   ├── list.go              # kanban list [filters]
│   │   ├── show.go              # kanban show <task>
│   │   └── epic.go              # kanban epic [subcommands]
│   ├── output/
│   │   └── output.go            # JSON/human output formatting
│   ├── resolver/
│   │   └── resolver.go          # ID/title task resolution
│   └── hooks/
│       └── hooks.go             # PocketBase event hooks
├── ui/
│   ├── src/
│   │   ├── App.tsx
│   │   ├── components/
│   │   │   ├── Board.tsx        # Main kanban board
│   │   │   ├── Column.tsx       # Single column
│   │   │   └── TaskCard.tsx     # Draggable task card
│   │   ├── hooks/
│   │   │   └── usePocketBase.ts # PocketBase SDK wrapper
│   │   └── lib/
│   │       └── pb.ts            # PocketBase client instance
│   ├── dist/                    # Vite build output
│   ├── embed.go                 # go:embed directive
│   ├── package.json
│   └── vite.config.ts
├── migrations/                  # PocketBase migrations
│   └── 1234567890_initial.go
├── pb_data/                     # SQLite + uploads (gitignored)
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Data Model (PocketBase Collections)

### tasks
| Field       | Type     | Description                              |
|-------------|----------|------------------------------------------|
| id          | string   | Auto-generated (or user-provided for idempotency) |
| title       | string   | Task title                               |
| description | string   | Optional longer description              |
| type        | select   | bug, feature, chore                      |
| priority    | select   | low, medium, high, urgent                |
| column      | select   | backlog, todo, in_progress, review, done |
| position    | number   | Order within column (fractional allowed) |
| epic        | relation | Optional link to epics collection        |
| labels      | json     | Array of label strings                   |
| created     | date     | Auto-generated                           |
| updated     | date     | Auto-generated                           |

### epics
| Field       | Type     | Description                              |
|-------------|----------|------------------------------------------|
| id          | string   | Auto-generated                           |
| title       | string   | Epic title                               |
| description | string   | Epic description                         |
| color       | string   | Hex color for visual grouping            |

## CLI Interface

### Global Flags

All commands support these flags:

| Flag       | Short | Description                          |
|------------|-------|--------------------------------------|
| `--json`   | `-j`  | Output in JSON format                |
| `--quiet`  | `-q`  | Suppress non-essential output        |
| `--data`   |       | Path to pb_data directory            |

### Task Resolution

Commands that accept `<task>` resolve it in this order:
1. **Exact ID match**: If input matches a task ID exactly
2. **ID prefix match**: If input is a unique prefix of a task ID
3. **Title match**: Case-insensitive substring match (fails if ambiguous)

```bash
# All of these could reference the same task
egenskriven show abc123def456
egenskriven show abc123
egenskriven show "fix login"
```

If ambiguous, the command fails with a list of matching tasks.

### Commands

#### `egenskriven add`

Create one or more tasks.

```bash
# Basic usage
egenskriven add "Implement dark mode"

# With options
egenskriven add "Fix login crash" --type bug --priority urgent --column todo

# With custom ID (idempotent)
egenskriven add "Setup CI pipeline" --id ci-setup-001

# With labels and epic
egenskriven add "Add user avatars" --label ui --label frontend --epic abc123

# Batch: from stdin (one JSON per line)
echo '{"title":"Task 1"}
{"title":"Task 2","priority":"high"}' | egenskriven add --stdin

# Batch: from file
egenskriven add --file tasks.json
```

**Flags:**

| Flag         | Short | Default   | Description                        |
|--------------|-------|-----------|------------------------------------|
| `--type`     | `-t`  | feature   | bug, feature, chore                |
| `--priority` | `-p`  | medium    | low, medium, high, urgent          |
| `--column`   | `-c`  | backlog   | Initial column                     |
| `--label`    | `-l`  |           | Add label (repeatable)             |
| `--epic`     | `-e`  |           | Link to epic                       |
| `--id`       |       |           | Custom ID (for idempotency)        |
| `--stdin`    |       |           | Read tasks from stdin (JSON lines) |
| `--file`     | `-f`  |           | Read tasks from JSON file          |

**Output (human):**
```
✓ Created task: Implement dark mode [abc123]
```

**Output (JSON):**
```json
{
  "id": "abc123def456",
  "title": "Implement dark mode",
  "type": "feature",
  "priority": "medium",
  "column": "backlog",
  "position": 1000,
  "labels": [],
  "created": "2024-01-15T10:30:00Z"
}
```

#### `egenskriven list`

List and filter tasks.

```bash
# All tasks
egenskriven list

# Filter by column
egenskriven list --column todo
egenskriven list -c in_progress -c review  # multiple columns

# Filter by type, priority
egenskriven list --type bug --priority urgent

# Filter by label
egenskriven list --label frontend

# Filter by epic
egenskriven list --epic abc123

# Combine filters (AND logic)
egenskriven list --column todo --type bug --priority high

# Search by title
egenskriven list --search "login"

# Output as JSON (for agents)
egenskriven list --json
```

**Flags:**

| Flag         | Short | Description                           |
|--------------|-------|---------------------------------------|
| `--column`   | `-c`  | Filter by column (repeatable)         |
| `--type`     | `-t`  | Filter by type (repeatable)           |
| `--priority` | `-p`  | Filter by priority (repeatable)       |
| `--label`    | `-l`  | Filter by label (repeatable)          |
| `--epic`     | `-e`  | Filter by epic                        |
| `--search`   | `-s`  | Search title (case-insensitive)       |
| `--limit`    |       | Max results (default: no limit)       |
| `--sort`     |       | Sort field (default: position)        |

**Output (human):**
```
BACKLOG
  [abc123] Implement dark mode (feature, medium)
  [def456] Add user avatars (feature, low)

TODO
  [ghi789] Fix login crash (bug, urgent)

IN_PROGRESS
  (empty)
```

**Output (JSON):**
```json
{
  "tasks": [
    {
      "id": "abc123",
      "title": "Implement dark mode",
      "type": "feature",
      "priority": "medium",
      "column": "backlog",
      "position": 1000,
      "labels": [],
      "created": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 1
}
```

#### `egenskriven show`

Show detailed information about a task.

```bash
egenskriven show abc123
egenskriven show "login crash"  # title match
```

**Output (human):**
```
Task: abc123def456
Title:       Fix login crash
Type:        bug
Priority:    urgent
Column:      todo
Position:    1000
Labels:      auth, critical
Epic:        -
Created:     2024-01-15 10:30:00
Updated:     2024-01-15 14:22:00

Description:
  Users are experiencing crashes when attempting to log in
  with SSO credentials. Stack trace attached.
```

#### `egenskriven move`

Move task to a different column and/or position.

```bash
# Move to column (appends to end)
egenskriven move abc123 in_progress

# Move to column at specific position
egenskriven move abc123 in_progress --position 0  # top
egenskriven move abc123 in_progress --position -1 # bottom (default)

# Move to position within current column
egenskriven move abc123 --position 0

# Move relative to another task
egenskriven move abc123 --after def456
egenskriven move abc123 --before def456
```

**Flags:**

| Flag         | Description                              |
|--------------|------------------------------------------|
| `--position` | Numeric position (0=top, -1=bottom)      |
| `--after`    | Position after this task                 |
| `--before`   | Position before this task                |

#### `egenskriven update`

Update task fields.

```bash
# Update single field
egenskriven update abc123 --title "New title"
egenskriven update abc123 --priority urgent

# Update multiple fields
egenskriven update abc123 --type bug --priority high --label critical

# Clear optional fields
egenskriven update abc123 --description ""
egenskriven update abc123 --epic ""

# Add/remove labels
egenskriven update abc123 --add-label urgent --remove-label backlog
```

**Flags:**

| Flag             | Description                        |
|------------------|------------------------------------|
| `--title`        | New title                          |
| `--description`  | New description                    |
| `--type`         | New type                           |
| `--priority`     | New priority                       |
| `--epic`         | Link to epic (empty to clear)      |
| `--add-label`    | Add label (repeatable)             |
| `--remove-label` | Remove label (repeatable)          |

#### `egenskriven delete`

Delete one or more tasks.

```bash
# Single task
egenskriven delete abc123

# Multiple tasks
egenskriven delete abc123 def456 ghi789

# From stdin
echo -e "abc123\ndef456" | egenskriven delete --stdin

# Skip confirmation (for scripts/agents)
egenskriven delete abc123 --force
```

**Flags:**

| Flag      | Short | Description                    |
|-----------|-------|--------------------------------|
| `--force` | `-f`  | Skip confirmation prompt       |
| `--stdin` |       | Read task IDs from stdin       |

#### `egenskriven epic`

Manage epics.

```bash
# List epics
egenskriven epic list

# Create epic
egenskriven epic add "Q1 Launch" --color "#3B82F6"

# Show epic with linked tasks
egenskriven epic show abc123

# Delete epic (tasks remain, unlinked)
egenskriven epic delete abc123
```

### Batch Input Format

For `--stdin` and `--file`, tasks are specified as JSON lines (one JSON object per line):

```json
{"title": "Task 1", "type": "bug", "priority": "high"}
{"title": "Task 2", "column": "todo"}
{"title": "Task 3", "id": "custom-id-001"}
```

Or as a JSON array:

```json
[
  {"title": "Task 1", "type": "bug"},
  {"title": "Task 2", "column": "todo"}
]
```

### Exit Codes

| Code | Meaning                                    |
|------|--------------------------------------------|
| 0    | Success                                    |
| 1    | General error                              |
| 2    | Invalid arguments/flags                    |
| 3    | Task not found                             |
| 4    | Ambiguous task reference                   |
| 5    | Validation error (invalid type, etc.)      |

### Error Output (JSON mode)

```json
{
  "error": {
    "code": 4,
    "message": "Ambiguous task reference: 'login' matches multiple tasks",
    "matches": [
      {"id": "abc123", "title": "Fix login crash"},
      {"id": "def456", "title": "Add login analytics"}
    ]
  }
}
```

## Key Implementation Details

### 1. Main Entry Point (cmd/kanban/main.go)

```go
package main

import (
    "log"

    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/core"
    
    "egenskriven/internal/commands"
    "egenskriven/ui"
)

func main() {
    app := pocketbase.New()

    // Register custom CLI commands
    commands.Register(app)

    // Serve embedded React frontend
    app.OnServe().BindFunc(func(e *core.ServeEvent) error {
        e.Router.GET("/{path...}", func(re *core.RequestEvent) error {
            path := re.Request.PathValue("path")
            
            if f, err := ui.DistFS.Open(path); err == nil {
                f.Close()
                return re.FileFS(ui.DistFS, path)
            }
            
            return re.FileFS(ui.DistFS, "index.html")
        })
        
        return e.Next()
    })

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```

### 2. Task Resolver (internal/resolver/resolver.go)

```go
package resolver

import (
    "fmt"
    "strings"

    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/core"
)

type Resolution struct {
    Task    *core.Record
    Matches []*core.Record // populated if ambiguous
}

func ResolveTask(app *pocketbase.PocketBase, ref string) (*Resolution, error) {
    // 1. Try exact ID match
    if task, err := app.FindRecordById("tasks", ref); err == nil {
        return &Resolution{Task: task}, nil
    }

    // 2. Try ID prefix match
    tasks, err := app.FindAllRecords("tasks", 
        dbx.NewExp("id LIKE {:prefix}", dbx.Params{"prefix": ref + "%"}),
    )
    if err == nil && len(tasks) == 1 {
        return &Resolution{Task: tasks[0]}, nil
    }

    // 3. Try title match (case-insensitive substring)
    tasks, err = app.FindAllRecords("tasks",
        dbx.NewExp("LOWER(title) LIKE {:title}", 
            dbx.Params{"title": "%" + strings.ToLower(ref) + "%"}),
    )
    if err != nil {
        return nil, err
    }

    switch len(tasks) {
    case 0:
        return nil, fmt.Errorf("no task found matching: %s", ref)
    case 1:
        return &Resolution{Task: tasks[0]}, nil
    default:
        return &Resolution{Matches: tasks}, nil // ambiguous
    }
}
```

### 3. Output Formatter (internal/output/output.go)

```go
package output

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/pocketbase/pocketbase/core"
)

type Formatter struct {
    JSON  bool
    Quiet bool
}

func (f *Formatter) Task(task *core.Record) {
    if f.JSON {
        json.NewEncoder(os.Stdout).Encode(taskToMap(task))
        return
    }

    fmt.Printf("✓ Created task: %s [%s]\n", 
        task.GetString("title"), 
        task.Id)
}

func (f *Formatter) Error(code int, message string, data any) {
    if f.JSON {
        json.NewEncoder(os.Stderr).Encode(map[string]any{
            "error": map[string]any{
                "code":    code,
                "message": message,
                "data":    data,
            },
        })
        return
    }

    fmt.Fprintf(os.Stderr, "Error: %s\n", message)
}

func taskToMap(task *core.Record) map[string]any {
    return map[string]any{
        "id":          task.Id,
        "title":       task.GetString("title"),
        "description": task.GetString("description"),
        "type":        task.GetString("type"),
        "priority":    task.GetString("priority"),
        "column":      task.GetString("column"),
        "position":    task.GetFloat("position"),
        "labels":      task.Get("labels"),
        "epic":        task.GetString("epic"),
        "created":     task.GetDateTime("created"),
        "updated":     task.GetDateTime("updated"),
    }
}
```

### 4. Add Command (internal/commands/add.go)

```go
package commands

import (
    "bufio"
    "encoding/json"
    "fmt"
    "os"

    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/core"
    "github.com/spf13/cobra"

    "egenskriven/internal/output"
)

func NewAddCmd(app *pocketbase.PocketBase, out *output.Formatter) *cobra.Command {
    var (
        taskType string
        priority string
        column   string
        labels   []string
        epic     string
        customID string
        stdin    bool
        file     string
    )

    cmd := &cobra.Command{
        Use:   "add [title]",
        Short: "Add a new task",
        Long: `Add a new task to the kanban board.

Supports batch creation via --stdin or --file for agent workflows.`,
        Example: `  egenskriven add "Implement dark mode"
  egenskriven add "Fix bug" --type bug --priority urgent
  egenskriven add "Setup CI" --id ci-setup-001
  echo '{"title":"Task 1"}' | egenskriven add --stdin`,
        RunE: func(cmd *cobra.Command, args []string) error {
            if err := app.Bootstrap(); err != nil {
                return err
            }

            // Handle batch input
            if stdin || file != "" {
                return addBatch(app, out, stdin, file)
            }

            if len(args) == 0 {
                return fmt.Errorf("title is required")
            }

            task, err := createTask(app, TaskInput{
                ID:       customID,
                Title:    args[0],
                Type:     taskType,
                Priority: priority,
                Column:   column,
                Labels:   labels,
                Epic:     epic,
            })
            if err != nil {
                return err
            }

            out.Task(task)
            return nil
        },
    }

    cmd.Flags().StringVarP(&taskType, "type", "t", "feature", "Task type (bug, feature, chore)")
    cmd.Flags().StringVarP(&priority, "priority", "p", "medium", "Priority (low, medium, high, urgent)")
    cmd.Flags().StringVarP(&column, "column", "c", "backlog", "Initial column")
    cmd.Flags().StringSliceVarP(&labels, "label", "l", nil, "Labels (repeatable)")
    cmd.Flags().StringVarP(&epic, "epic", "e", "", "Link to epic")
    cmd.Flags().StringVar(&customID, "id", "", "Custom ID for idempotency")
    cmd.Flags().BoolVar(&stdin, "stdin", false, "Read tasks from stdin")
    cmd.Flags().StringVarP(&file, "file", "f", "", "Read tasks from file")

    return cmd
}

type TaskInput struct {
    ID          string   `json:"id,omitempty"`
    Title       string   `json:"title"`
    Description string   `json:"description,omitempty"`
    Type        string   `json:"type,omitempty"`
    Priority    string   `json:"priority,omitempty"`
    Column      string   `json:"column,omitempty"`
    Labels      []string `json:"labels,omitempty"`
    Epic        string   `json:"epic,omitempty"`
}

func createTask(app *pocketbase.PocketBase, input TaskInput) (*core.Record, error) {
    collection, err := app.FindCollectionByNameOrId("tasks")
    if err != nil {
        return nil, err
    }

    record := core.NewRecord(collection)
    
    if input.ID != "" {
        record.SetId(input.ID)
    }
    
    record.Set("title", input.Title)
    record.Set("type", defaultString(input.Type, "feature"))
    record.Set("priority", defaultString(input.Priority, "medium"))
    record.Set("column", defaultString(input.Column, "backlog"))
    record.Set("position", getNextPosition(app, input.Column))
    
    if len(input.Labels) > 0 {
        record.Set("labels", input.Labels)
    }
    if input.Epic != "" {
        record.Set("epic", input.Epic)
    }
    if input.Description != "" {
        record.Set("description", input.Description)
    }

    if err := app.Save(record); err != nil {
        return nil, err
    }

    return record, nil
}

func addBatch(app *pocketbase.PocketBase, out *output.Formatter, useStdin bool, file string) error {
    var reader *bufio.Scanner

    if useStdin {
        reader = bufio.NewScanner(os.Stdin)
    } else {
        f, err := os.Open(file)
        if err != nil {
            return err
        }
        defer f.Close()
        reader = bufio.NewScanner(f)
    }

    var created []*core.Record
    for reader.Scan() {
        line := reader.Text()
        if line == "" {
            continue
        }

        var input TaskInput
        if err := json.Unmarshal([]byte(line), &input); err != nil {
            return fmt.Errorf("invalid JSON: %s", line)
        }

        task, err := createTask(app, input)
        if err != nil {
            return err
        }
        created = append(created, task)
    }

    for _, task := range created {
        out.Task(task)
    }
    return nil
}
```

### 5. Position Management

Using fractional indexing for positions to avoid rebalancing:

```go
func getNextPosition(app *pocketbase.PocketBase, column string) float64 {
    tasks, _ := app.FindAllRecords("tasks",
        dbx.NewExp("column = {:col}", dbx.Params{"col": column}),
        dbx.OrderBy("position DESC"),
        dbx.Limit(1),
    )
    
    if len(tasks) == 0 {
        return 1000.0
    }
    
    return tasks[0].GetFloat("position") + 1000.0
}

func getPositionBetween(before, after float64) float64 {
    return (before + after) / 2.0
}
```

### 6. React Frontend - PocketBase Hook (ui/src/hooks/usePocketBase.ts)

```typescript
import { useEffect, useState } from 'react';
import PocketBase, { RecordModel } from 'pocketbase';

const pb = new PocketBase('/');

export interface Task extends RecordModel {
  title: string;
  description?: string;
  type: 'bug' | 'feature' | 'chore';
  priority: 'low' | 'medium' | 'high' | 'urgent';
  column: 'backlog' | 'todo' | 'in_progress' | 'review' | 'done';
  position: number;
  epic?: string;
  labels?: string[];
}

export function useTasks() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    pb.collection('tasks')
      .getFullList<Task>({ sort: 'position' })
      .then(setTasks)
      .finally(() => setLoading(false));

    pb.collection('tasks').subscribe<Task>('*', (e) => {
      if (e.action === 'create') {
        setTasks(prev => [...prev, e.record]);
      } else if (e.action === 'update') {
        setTasks(prev => prev.map(t => t.id === e.record.id ? e.record : t));
      } else if (e.action === 'delete') {
        setTasks(prev => prev.filter(t => t.id !== e.record.id));
      }
    });

    return () => {
      pb.collection('tasks').unsubscribe('*');
    };
  }, []);

  const moveTask = async (taskId: string, newColumn: string, newPosition: number) => {
    await pb.collection('tasks').update(taskId, {
      column: newColumn,
      position: newPosition,
    });
  };

  const createTask = async (task: Partial<Task>) => {
    return pb.collection('tasks').create(task);
  };

  return { tasks, loading, moveTask, createTask };
}
```

### 7. Kanban Board Component (ui/src/components/Board.tsx)

```tsx
import { useMemo } from 'react';
import {
  DndContext,
  DragEndEvent,
  closestCenter,
} from '@dnd-kit/core';
import { useTasks, Task } from '../hooks/usePocketBase';
import { Column } from './Column';

const COLUMNS = ['backlog', 'todo', 'in_progress', 'review', 'done'] as const;

export function Board() {
  const { tasks, loading, moveTask } = useTasks();

  const tasksByColumn = useMemo(() => {
    const grouped: Record<string, Task[]> = {};
    COLUMNS.forEach(col => grouped[col] = []);
    
    tasks.forEach(task => {
      if (grouped[task.column]) {
        grouped[task.column].push(task);
      }
    });
    
    Object.keys(grouped).forEach(col => {
      grouped[col].sort((a, b) => a.position - b.position);
    });
    
    return grouped;
  }, [tasks]);

  const handleDragEnd = async (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over) return;

    const taskId = active.id as string;
    const newColumn = over.data.current?.column as string;
    const newPosition = over.data.current?.position as number;

    if (newColumn) {
      await moveTask(taskId, newColumn, newPosition);
    }
  };

  if (loading) return <div>Loading...</div>;

  return (
    <DndContext collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
      <div className="flex gap-4 p-4 h-screen bg-gray-100">
        {COLUMNS.map(column => (
          <Column
            key={column}
            name={column}
            tasks={tasksByColumn[column]}
          />
        ))}
      </div>
    </DndContext>
  );
}
```

## Build & Development

### Makefile

```makefile
.PHONY: dev build clean

# Development: run React dev server + Go with Air
dev:
	@$(MAKE) -j2 dev-ui dev-go

dev-ui:
	cd ui && npm run dev

dev-go:
	air

# Build production binary
build: build-ui build-go

build-ui:
	cd ui && npm ci && npm run build

build-go:
	CGO_ENABLED=0 go build -o egenskriven ./cmd/egenskriven

# Cross-compile for all platforms
release: build-ui
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o dist/egenskriven-darwin-arm64 ./cmd/egenskriven
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o dist/egenskriven-darwin-amd64 ./cmd/egenskriven
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/egenskriven-linux-amd64 ./cmd/egenskriven
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o dist/egenskriven-windows-amd64.exe ./cmd/egenskriven

clean:
	rm -rf egenskriven dist/ ui/dist/
```

### Vite Config (ui/vite.config.ts)

```typescript
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
  server: {
    proxy: {
      '/api': 'http://localhost:8090',
      '/_': 'http://localhost:8090',
    },
  },
});
```

## CLI Usage Examples

```bash
# Start the server (web UI + API)
./egenskriven serve

# Basic task management
./egenskriven add "Implement dark mode" --type feature --priority medium
./egenskriven add "Fix login crash" --type bug --priority urgent --column todo
./egenskriven move abc123 in_progress
./egenskriven update abc123 --priority high --add-label critical
./egenskriven delete abc123

# Querying
./egenskriven list
./egenskriven list --column todo --type bug
./egenskriven list --priority urgent --json
./egenskriven show abc123

# Batch operations (agent-friendly)
./egenskriven add --stdin <<EOF
{"title": "Task 1", "type": "bug"}
{"title": "Task 2", "priority": "high"}
EOF

# JSON output for agents
./egenskriven list --json | jq '.tasks[] | select(.priority == "urgent")'

# Idempotent creation (safe for retries)
./egenskriven add "Setup CI" --id ci-setup-001
./egenskriven add "Setup CI" --id ci-setup-001  # no-op if exists

# View board in browser
open http://localhost:8090
```

## Real-time Sync: How It Works

1. **CLI creates task** → writes directly to SQLite via PocketBase's Go API
2. **PocketBase detects change** → broadcasts SSE event to subscribed clients
3. **React frontend** → receives event via `pb.collection('tasks').subscribe('*', ...)`
4. **UI updates** → React state updates, board re-renders

**Note**: CLI commands run in a separate process from `serve`. They share the SQLite database but not the in-memory event bus. Real-time works because PocketBase's SSE is database-driven.

## Future Enhancements

1. **Offline CLI mode**: Queue operations when server isn't running, sync on next `serve`
2. **TUI mode**: Full terminal UI with Bubble Tea (`egenskriven board` opens interactive view)
3. **Git integration**: Auto-create tasks from commit messages, link branches to tasks
4. **Templates**: `egenskriven add --template bug-report` for predefined task structures
5. **Archiving**: `egenskriven archive` to move done tasks to archive, keeping board clean

## Why This Architecture?

| Decision | Rationale |
|----------|-----------|
| PocketBase | Free real-time, admin UI, auth, REST API. Why rebuild? |
| Embedded React | Single binary distribution, no separate frontend deploy |
| SQLite | Perfect for local-first, no external DB needed |
| Cobra CLI | Industry standard, same as kubectl/docker/gh |
| @dnd-kit | Best React DnD library, accessible, small bundle |
| Agent-friendly CLI | Any AI can use it directly, no special integration needed |
| JSON output | Structured data for programmatic use |
| Fractional positions | Avoids rebalancing on every move |
