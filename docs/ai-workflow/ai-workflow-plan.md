# AI-Human Collaborative Workflow Implementation Plan

> **Version**: 1.0.0  
> **Status**: Planning  
> **Created**: 2026-01-07  
> **Target Release**: EgenSkriven v1.0.0

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Vision & Goals](#2-vision--goals)
3. [Research Findings](#3-research-findings)
4. [Design Decisions](#4-design-decisions)
5. [Architecture Overview](#5-architecture-overview)
6. [Data Model Changes](#6-data-model-changes)
7. [CLI Commands](#7-cli-commands)
8. [Tool-Specific Integrations](#8-tool-specific-integrations)
9. [Resume Flow Implementation](#9-resume-flow-implementation)
10. [Web UI Changes](#10-web-ui-changes)
11. [Skills & Documentation](#11-skills--documentation)
12. [Implementation Phases](#12-implementation-phases)
13. [Testing Strategy](#13-testing-strategy)
14. [Appendix](#14-appendix)

---

## 1. Executive Summary

### Problem Statement

Currently, AI coding agents (OpenCode, Claude Code, Codex) work in isolation. When an agent gets stuck or needs human input, there's no structured way to:
- Signal that human input is needed
- Capture the human's response
- Resume the agent's work with full context preserved

### Solution

Extend EgenSkriven to serve as a **control plane** for AI agent workflows, enabling:
1. Agents to move tasks through a workflow (including a `need_input` blocked state)
2. Humans to provide input via a comments system on tasks
3. Agents to resume work with injected context from comments
4. Session tracking across multiple AI tools

### Supported Tools (v1.0.0)

| Tool | Session Support | Integration Method |
|------|----------------|-------------------|
| OpenCode | Full | Custom Tool |
| Claude Code | Full | Hooks System |
| Codex CLI | Workaround | Rollout File Parsing |

---

## 2. Vision & Goals

### Vision

Make EgenSkriven the universal control plane for human-AI collaborative coding workflows, regardless of which AI coding tool is used.

### Goals for v1.0.0

1. **Blocked State Workflow**: Agent can signal it's blocked and needs human input
2. **Comments System**: Humans can respond to agent questions on tasks
3. **Session Tracking**: Link AI tool sessions to EgenSkriven tasks
4. **Context-Preserving Resume**: Resume agent work with full conversation context
5. **Multi-Tool Support**: Work with OpenCode, Claude Code, and Codex CLI

### Non-Goals for v1.0.0

- MCP server implementation (nice-to-have for future)
- Aider support (limited session capabilities)
- Automatic workflow triggers beyond `@agent` mention
- Real-time notifications/webhooks

---

## 3. Research Findings

### 3.1 EgenSkriven Current Architecture

**Technology Stack:**
- Backend: Go with PocketBase (embedded SQLite)
- Frontend: React with TypeScript (embedded in Go binary)
- CLI: Cobra (via PocketBase)
- Real-time: PocketBase SSE subscriptions

**Current Task Schema:**
```go
Fields:
- id, title, description, type, priority
- column        // backlog, todo, in_progress, review, done
- position, labels, blocked_by
- created_by, created_by_agent
- history       // JSON array of activity entries
- epic, board, seq, due_date, parent
- created, updated
```

**Gaps Identified:**
- No `need_input` column/state
- No comments system (only `history` for structural changes)
- No session tracking for AI tools

### 3.2 AI Tool Session Capabilities

#### OpenCode

| Feature | Support | Details |
|---------|---------|---------|
| Sessions | Yes | Persisted, UUID-based |
| Resume | Yes | `opencode run "prompt" --session <id>` |
| Session ID Access | Via Custom Tool | `context.sessionID` in tool execute |
| Hooks/Plugins | Yes | Full plugin system with event hooks |

**Resume Command:**
```bash
opencode run "Continue working on the task" --session <session-id>
```

**Session ID Discovery:**
Custom tools receive session ID in context:
```typescript
// .opencode/tool/get-session.ts
import { tool } from "@opencode-ai/plugin"

export default tool({
  description: "Get current OpenCode session ID",
  args: {},
  async execute(args, context) {
    return context.sessionID
  },
})
```

#### Claude Code

| Feature | Support | Details |
|---------|---------|---------|
| Sessions | Yes | Persisted in `~/.claude/projects/` |
| Resume | Yes | `claude --resume <id> "prompt"` |
| Session ID Access | Via Hooks | `session_id` in stdin JSON |
| Hooks | Yes | SessionStart, SessionEnd, etc. |

**Resume Command:**
```bash
claude --resume <session-id> "Continue working on the task"
```

**Session ID Discovery:**
Hooks receive session ID; can persist to env var:
```bash
#!/bin/bash
# SessionStart hook - persist session ID
SESSION_ID=$(cat | jq -r '.session_id')
if [ -n "$CLAUDE_ENV_FILE" ]; then
  echo "export CLAUDE_SESSION_ID=$SESSION_ID" >> "$CLAUDE_ENV_FILE"
fi
```

#### Codex CLI

| Feature | Support | Details |
|---------|---------|---------|
| Sessions | Yes | Called "threads", UUID v7 |
| Resume | Yes | `codex exec resume <id> "prompt"` |
| Session ID Access | Workaround | Parse rollout filename |
| Hooks | No | Limited extension points |

**Resume Command:**
```bash
codex exec resume <thread-id> "Continue working on the task"
```

**Session ID Discovery (Workaround):**
```bash
# Parse from most recent rollout file
ls -t ~/.codex/sessions/rollout-*.jsonl | head -1 | \
  grep -oP '[0-9a-f]{8}(-[0-9a-f]{4}){3}-[0-9a-f]{12}'
```

### 3.3 Resume Command Patterns

All three tools support passing a prompt when resuming:

| Tool | Command Pattern |
|------|----------------|
| OpenCode | `opencode run "<prompt>" --session <id>` |
| Claude Code | `claude --resume <id> "<prompt>"` |
| Codex | `codex exec resume <id> "<prompt>"` |

This enables context injection when resuming blocked tasks.

---

## 4. Design Decisions

### Decision Log

| # | Topic | Decision | Rationale |
|---|-------|----------|-----------|
| 1 | Comments storage | New `comments` collection | Proper schema, scalable, queryable |
| 2 | Session linking | Task field + sessions table | Current session on task, history in table |
| 3 | Blocked state | New `need_input` column | Explicit, visible on kanban board |
| 4 | Resume mechanism | All 3 modes, configurable | Flexibility per project |
| 5 | Field naming | `agent_session` (tool-agnostic) | Support multiple tools uniformly |
| 6 | v1 tool support | OpenCode, Claude Code, Codex | Best session support |
| 7 | MCP server | Not in v1 | Skills + CLI is more portable |
| 8 | Comments threading | Flat list | Keep simple for v1 |
| 9 | Resume command | Print + `--exec` flag | Safe default, power user option |
| 10 | Auto-resume trigger | `@agent` mention | Prevent accidental triggers |
| 11 | Context format | Rich (full thread) | More context = better agent performance |
| 12 | Tool handoff | Allow, mark old abandoned | Flexibility with clear state |
| 13 | Status on resume | Per mode behavior | `command` → explicit, `auto` → on pickup |
| 14 | Web UI comments | Include in v1 | Essential for human interaction |
| 15 | Author identification | Config → env → "human" | Flexible identification |
| 16 | Block command | Atomic operation | Single command for consistency |
| 17 | Session registration | Explicit CLI call | No auto-detection (unreliable) |
| 18 | Tool integrations | Auto-generated locally | Works with tool permissions |

### Resume Mode Configuration

Per-board configuration for resume behavior:

| Mode | Behavior | Trigger |
|------|----------|---------|
| `manual` | Print command for user to run | User copies command |
| `command` | `egenskriven resume <task>` spawns tool | User runs command |
| `auto` | Auto-resume on `@agent` comment | Comment with mention |

**Default**: `command`

---

## 5. Architecture Overview

### System Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              AI Agent                                    │
│  (OpenCode / Claude Code / Codex)                                       │
│                                                                          │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │  Agent detects it's blocked                                       │   │
│  │  ↓                                                                │   │
│  │  Calls: egenskriven block WRK-123 "What auth approach to use?"   │   │
│  │  ↓                                                                │   │
│  │  Task moves to need_input, comment added                         │   │
│  │  ↓                                                                │   │
│  │  Agent session preserved (can resume later)                      │   │
│  └──────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                           EgenSkriven                                    │
│                                                                          │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────────┐  │
│  │     Tasks       │  │    Comments     │  │       Sessions          │  │
│  │                 │  │                 │  │                         │  │
│  │ - column        │  │ - task (rel)    │  │ - task (rel)            │  │
│  │ - agent_session │  │ - content       │  │ - tool                  │  │
│  │   - tool        │  │ - author_type   │  │ - external_ref          │  │
│  │   - ref         │  │ - author_id     │  │ - status                │  │
│  │   - working_dir │  │ - created       │  │ - working_dir           │  │
│  └────────┬────────┘  └────────┬────────┘  └─────────────────────────┘  │
│           │                    │                                         │
│           └────────────────────┴─────────────────────────────────────┐  │
│                                                                       │  │
│  ┌────────────────────────────────────────────────────────────────┐  │  │
│  │  CLI Commands                                                   │  │  │
│  │  - egenskriven block <task> "question"                         │  │  │
│  │  - egenskriven comment <task> "response"                       │  │  │
│  │  - egenskriven resume <task> [--exec]                          │  │  │
│  │  - egenskriven session link <task> --tool X --ref Y            │  │  │
│  └────────────────────────────────────────────────────────────────┘  │  │
│                                                                       │  │
│  ┌────────────────────────────────────────────────────────────────┐  │  │
│  │  Web UI                                                         │  │  │
│  │  - Comments panel on task detail                                │  │  │
│  │  - Resume button for need_input tasks                          │  │  │
│  │  - Session info display                                         │  │  │
│  └────────────────────────────────────────────────────────────────┘  │  │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                              Human User                                  │
│                                                                          │
│  1. Sees task in "need_input" column                                    │
│  2. Reads agent's question in comments                                  │
│  3. Adds response comment (with @agent if auto mode)                    │
│  4. Runs: egenskriven resume WRK-123 (or auto-triggered)               │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                           Resume Flow                                    │
│                                                                          │
│  egenskriven resume WRK-123 --exec                                      │
│  ↓                                                                       │
│  1. Fetch task + comments + session info                                │
│  2. Build context injection prompt                                      │
│  3. Spawn: opencode run "<context>" --session <ref>                    │
│  4. Move task to in_progress                                            │
│  5. Agent continues with full context                                   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Workflow State Machine

```
                    ┌──────────────────┐
                    │     backlog      │
                    └────────┬─────────┘
                             │ agent picks up
                             ▼
                    ┌──────────────────┐
                    │       todo       │
                    └────────┬─────────┘
                             │ agent starts work
                             ▼
                    ┌──────────────────┐
        ┌──────────│   in_progress    │──────────┐
        │          └────────┬─────────┘          │
        │                   │                    │
        │ blocked           │ done               │ needs review
        ▼                   │                    ▼
┌──────────────────┐        │           ┌──────────────────┐
│   need_input     │        │           │      review      │
└────────┬─────────┘        │           └────────┬─────────┘
         │                  │                    │
         │ human responds   │                    │ approved
         │ + resume         │                    │
         │                  ▼                    │
         │         ┌──────────────────┐          │
         └────────►│       done       │◄─────────┘
                   └──────────────────┘
```

---

## 6. Data Model Changes

### 6.1 Tasks Collection Updates

#### New Column Value

Add `need_input` to the `column` field's allowed values:

```go
// internal/commands/root.go
var ValidColumns = []string{
    "backlog", 
    "todo", 
    "in_progress", 
    "need_input",    // NEW
    "review", 
    "done",
}
```

**Migration file: `migrations/12_need_input_column.go`**

```go
package migrations

import (
    "github.com/pocketbase/pocketbase/core"
    m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
    m.Register(func(app core.App) error {
        tasks, err := app.FindCollectionByNameOrId("tasks")
        if err != nil {
            return err
        }

        columnField := tasks.Fields.GetByName("column").(*core.SelectField)
        columnField.Values = []string{
            "backlog", "todo", "in_progress", "need_input", "review", "done",
        }

        return app.Save(tasks)
    }, func(app core.App) error {
        // Rollback
        tasks, err := app.FindCollectionByNameOrId("tasks")
        if err != nil {
            return err
        }

        columnField := tasks.Fields.GetByName("column").(*core.SelectField)
        columnField.Values = []string{
            "backlog", "todo", "in_progress", "review", "done",
        }

        return app.Save(tasks)
    })
}
```

#### New `agent_session` Field

Add JSON field for current agent session:

```go
// Add to tasks collection
collection.Fields.Add(&core.JSONField{
    Name:    "agent_session",
    MaxSize: 10000,
})
```

**Schema for `agent_session` JSON:**

```typescript
interface AgentSession {
    tool: "opencode" | "claude-code" | "codex";
    ref: string;           // Session/thread ID
    ref_type: "uuid" | "path";
    working_dir: string;   // Project directory
    linked_at: string;     // ISO timestamp
}
```

**Example:**
```json
{
    "tool": "opencode",
    "ref": "550e8400-e29b-41d4-a716-446655440000",
    "ref_type": "uuid",
    "working_dir": "/home/user/my-project",
    "linked_at": "2026-01-07T10:30:00Z"
}
```

### 6.2 New Comments Collection

**Migration file: `migrations/13_comments_collection.go`**

```go
package migrations

import (
    "github.com/pocketbase/pocketbase/core"
    m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
    m.Register(func(app core.App) error {
        collection := core.NewBaseCollection("comments")

        // Task relation (required)
        collection.Fields.Add(&core.RelationField{
            Name:          "task",
            Required:      true,
            MaxSelect:     1,
            CollectionId:  "tasks",  // Will be resolved
            CascadeDelete: true,
        })

        // Comment content
        collection.Fields.Add(&core.TextField{
            Name:     "content",
            Required: true,
            Max:      50000,
        })

        // Author type (human or agent)
        collection.Fields.Add(&core.SelectField{
            Name:     "author_type",
            Required: true,
            Values:   []string{"human", "agent"},
        })

        // Author identifier
        collection.Fields.Add(&core.TextField{
            Name:     "author_id",
            Required: false,
            Max:      200,
        })

        // Metadata (for session context, etc.)
        collection.Fields.Add(&core.JSONField{
            Name:    "metadata",
            MaxSize: 50000,
        })

        // Auto timestamps
        collection.Fields.Add(&core.AutodateField{
            Name:     "created",
            OnCreate: true,
        })

        collection.Indexes = []string{
            "CREATE INDEX idx_comments_task ON comments (task)",
            "CREATE INDEX idx_comments_created ON comments (created)",
        }

        return app.Save(collection)
    }, func(app core.App) error {
        collection, err := app.FindCollectionByNameOrId("comments")
        if err != nil {
            return nil // Collection doesn't exist, nothing to rollback
        }
        return app.Delete(collection)
    })
}
```

**Comment Schema:**

```typescript
interface Comment {
    id: string;
    task: string;           // Task ID (relation)
    content: string;        // Comment text
    author_type: "human" | "agent";
    author_id?: string;     // Username, agent name, or null
    metadata?: {
        session_ref?: string;    // If from agent session
        mentions?: string[];     // @agent, @user mentions
    };
    created: string;        // ISO timestamp
}
```

### 6.3 New Sessions Collection (History)

**Migration file: `migrations/14_sessions_collection.go`**

```go
package migrations

import (
    "github.com/pocketbase/pocketbase/core"
    m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
    m.Register(func(app core.App) error {
        collection := core.NewBaseCollection("sessions")

        // Task relation
        collection.Fields.Add(&core.RelationField{
            Name:          "task",
            Required:      true,
            MaxSelect:     1,
            CollectionId:  "tasks",
            CascadeDelete: true,
        })

        // Tool identifier
        collection.Fields.Add(&core.SelectField{
            Name:     "tool",
            Required: true,
            Values:   []string{"opencode", "claude-code", "codex"},
        })

        // External session reference
        collection.Fields.Add(&core.TextField{
            Name:     "external_ref",
            Required: true,
            Max:      500,
        })

        // Reference type
        collection.Fields.Add(&core.SelectField{
            Name:     "ref_type",
            Required: true,
            Values:   []string{"uuid", "path"},
        })

        // Working directory
        collection.Fields.Add(&core.TextField{
            Name:     "working_dir",
            Required: true,
            Max:      1000,
        })

        // Session status
        collection.Fields.Add(&core.SelectField{
            Name:     "status",
            Required: true,
            Values:   []string{"active", "paused", "completed", "abandoned"},
        })

        // Timestamps
        collection.Fields.Add(&core.AutodateField{
            Name:     "created",
            OnCreate: true,
        })

        collection.Fields.Add(&core.DateField{
            Name: "ended_at",
        })

        collection.Indexes = []string{
            "CREATE INDEX idx_sessions_task ON sessions (task)",
            "CREATE INDEX idx_sessions_status ON sessions (status)",
            "CREATE INDEX idx_sessions_external_ref ON sessions (external_ref)",
        }

        return app.Save(collection)
    }, func(app core.App) error {
        collection, err := app.FindCollectionByNameOrId("sessions")
        if err != nil {
            return nil
        }
        return app.Delete(collection)
    })
}
```

### 6.4 Boards Collection Update

Add resume mode configuration:

```go
// Add to boards collection
collection.Fields.Add(&core.SelectField{
    Name:     "resume_mode",
    Required: false,
    Values:   []string{"manual", "command", "auto"},
})
```

**Default value**: `command`

---

## 7. CLI Commands

### 7.1 `egenskriven block` Command

Atomic command to move task to `need_input` and add a comment.

**Usage:**
```bash
egenskriven block <task-ref> "question for human"
egenskriven block <task-ref> --stdin  # Read from stdin for longer questions
```

**Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--stdin` | | Read question from stdin |
| `--json` | `-j` | Output result as JSON |

**Implementation file: `internal/commands/block.go`**

```go
package commands

import (
    "github.com/spf13/cobra"
)

func newBlockCmd(app *pocketbase.PocketBase) *cobra.Command {
    var useStdin bool
    var jsonOutput bool

    cmd := &cobra.Command{
        Use:   "block <task-ref> [question]",
        Short: "Block a task and request human input",
        Long: `Move a task to the need_input column and add a comment with 
your question. This is an atomic operation that ensures the task 
state and comment are created together.`,
        Example: `  egenskriven block WRK-123 "What authentication approach should I use?"
  egenskriven block abc "Should I use JWT or session cookies?" --json
  echo "Long question here" | egenskriven block WRK-123 --stdin`,
        Args: cobra.RangeArgs(1, 2),
        RunE: func(cmd *cobra.Command, args []string) error {
            taskRef := args[0]
            
            var question string
            if useStdin {
                // Read from stdin
                data, err := io.ReadAll(os.Stdin)
                if err != nil {
                    return fmt.Errorf("failed to read from stdin: %w", err)
                }
                question = strings.TrimSpace(string(data))
            } else if len(args) > 1 {
                question = args[1]
            } else {
                return fmt.Errorf("question required: provide as argument or use --stdin")
            }

            // Resolve task
            task, err := resolver.MustResolve(app, taskRef)
            if err != nil {
                return err
            }

            // Start transaction
            return app.RunInTransaction(func(txApp core.App) error {
                // 1. Move task to need_input
                task.Set("column", "need_input")
                
                // 2. Add to history
                addHistoryEntry(task, "blocked", getAgentName(), map[string]any{
                    "reason": question,
                })
                
                if err := txApp.Save(task); err != nil {
                    return err
                }

                // 3. Create comment
                comment := core.NewRecord(commentsCollection)
                comment.Set("task", task.Id)
                comment.Set("content", question)
                comment.Set("author_type", "agent")
                comment.Set("author_id", getAgentName())
                
                if err := txApp.Save(comment); err != nil {
                    return err
                }

                // Output result
                if jsonOutput {
                    return outputJSON(map[string]any{
                        "task":    task.Id,
                        "column":  "need_input",
                        "comment": comment.Id,
                    })
                }
                
                fmt.Printf("Task %s blocked. Awaiting human input.\n", 
                    task.GetString("display_id"))
                return nil
            })
        },
    }

    cmd.Flags().BoolVar(&useStdin, "stdin", false, "Read question from stdin")
    cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")

    return cmd
}
```

### 7.2 `egenskriven comment` Command

Add a comment to a task.

**Usage:**
```bash
egenskriven comment <task-ref> "comment text"
egenskriven comment <task-ref> --stdin
```

**Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--stdin` | | Read comment from stdin |
| `--json` | `-j` | Output result as JSON |
| `--author` | `-a` | Override author identifier |

**Implementation file: `internal/commands/comment.go`**

```go
package commands

func newCommentCmd(app *pocketbase.PocketBase) *cobra.Command {
    var useStdin bool
    var jsonOutput bool
    var author string

    cmd := &cobra.Command{
        Use:   "comment <task-ref> [text]",
        Short: "Add a comment to a task",
        Example: `  egenskriven comment WRK-123 "Use JWT with refresh tokens"
  egenskriven comment WRK-123 "@agent I've decided to use OAuth2" 
  echo "Detailed response..." | egenskriven comment WRK-123 --stdin`,
        Args: cobra.RangeArgs(1, 2),
        RunE: func(cmd *cobra.Command, args []string) error {
            taskRef := args[0]
            
            var text string
            if useStdin {
                data, err := io.ReadAll(os.Stdin)
                if err != nil {
                    return err
                }
                text = strings.TrimSpace(string(data))
            } else if len(args) > 1 {
                text = args[1]
            } else {
                return fmt.Errorf("comment text required")
            }

            task, err := resolver.MustResolve(app, taskRef)
            if err != nil {
                return err
            }

            // Determine author
            authorId := resolveAuthor(author) // config → env → "human"
            authorType := "human"
            if isAgentContext() {
                authorType = "agent"
            }

            // Check for @agent mention
            mentions := extractMentions(text) // ["@agent"]

            // Create comment
            comment := core.NewRecord(commentsCollection)
            comment.Set("task", task.Id)
            comment.Set("content", text)
            comment.Set("author_type", authorType)
            comment.Set("author_id", authorId)
            comment.Set("metadata", map[string]any{
                "mentions": mentions,
            })

            if err := app.Save(comment); err != nil {
                return err
            }

            // Output
            if jsonOutput {
                return outputJSON(map[string]any{
                    "id":       comment.Id,
                    "task":     task.Id,
                    "mentions": mentions,
                })
            }

            fmt.Printf("Comment added to %s\n", task.GetString("display_id"))
            return nil
        },
    }

    cmd.Flags().BoolVar(&useStdin, "stdin", false, "Read from stdin")
    cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
    cmd.Flags().StringVarP(&author, "author", "a", "", "Author identifier")

    return cmd
}
```

### 7.3 `egenskriven comments` Command

List comments for a task.

**Usage:**
```bash
egenskriven comments <task-ref>
egenskriven comments <task-ref> --since "2026-01-07T10:00:00Z"
egenskriven comments <task-ref> --json
```

**Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--since` | | Only show comments after timestamp |
| `--limit` | `-n` | Limit number of comments |
| `--json` | `-j` | Output as JSON |

### 7.4 `egenskriven session` Command

Manage agent sessions linked to tasks.

**Subcommands:**

#### `egenskriven session link`

Link an agent session to a task.

```bash
egenskriven session link <task-ref> --tool opencode --ref <session-id>
egenskriven session link WRK-123 --tool claude-code --ref abc-123
egenskriven session link WRK-123 --tool codex --ref 550e8400-e29b-41d4-a716-446655440000
```

**Flags:**
| Flag | Description | Required |
|------|-------------|----------|
| `--tool` | Tool name (opencode, claude-code, codex) | Yes |
| `--ref` | Session/thread ID | Yes |
| `--working-dir` | Working directory (defaults to cwd) | No |

**Implementation:**
```go
func newSessionLinkCmd(app *pocketbase.PocketBase) *cobra.Command {
    var tool, ref, workingDir string

    cmd := &cobra.Command{
        Use:   "link <task-ref>",
        Short: "Link an agent session to a task",
        RunE: func(cmd *cobra.Command, args []string) error {
            taskRef := args[0]

            task, err := resolver.MustResolve(app, taskRef)
            if err != nil {
                return err
            }

            if workingDir == "" {
                workingDir, _ = os.Getwd()
            }

            // Validate tool
            validTools := []string{"opencode", "claude-code", "codex"}
            if !contains(validTools, tool) {
                return fmt.Errorf("invalid tool: %s", tool)
            }

            // Check for existing session, mark as abandoned
            existingSession := task.Get("agent_session")
            if existingSession != nil {
                // Create history record for old session
                oldSession := existingSession.(map[string]any)
                createSessionRecord(app, task.Id, oldSession, "abandoned")
            }

            // Set new session
            refType := "uuid"
            if strings.HasPrefix(ref, "/") || strings.HasPrefix(ref, ".") {
                refType = "path"
            }

            session := map[string]any{
                "tool":        tool,
                "ref":         ref,
                "ref_type":    refType,
                "working_dir": workingDir,
                "linked_at":   time.Now().Format(time.RFC3339),
            }

            task.Set("agent_session", session)
            
            // Also create session record
            createSessionRecord(app, task.Id, session, "active")

            return app.Save(task)
        },
    }

    cmd.Flags().StringVar(&tool, "tool", "", "Tool name (required)")
    cmd.Flags().StringVar(&ref, "ref", "", "Session reference (required)")
    cmd.Flags().StringVar(&workingDir, "working-dir", "", "Working directory")
    cmd.MarkFlagRequired("tool")
    cmd.MarkFlagRequired("ref")

    return cmd
}
```

#### `egenskriven session show`

Show session info for a task.

```bash
egenskriven session show <task-ref>
egenskriven session show WRK-123 --json
```

#### `egenskriven session history`

Show session history for a task.

```bash
egenskriven session history <task-ref>
```

### 7.5 `egenskriven resume` Command

Resume work on a blocked task.

**Usage:**
```bash
egenskriven resume <task-ref>           # Print command to run
egenskriven resume <task-ref> --exec    # Actually spawn the tool
egenskriven resume <task-ref> --json    # Output command as JSON
```

**Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--exec` | `-e` | Execute the resume command |
| `--json` | `-j` | Output as JSON |
| `--prompt` | `-p` | Override the context prompt |

**Implementation file: `internal/commands/resume.go`**

```go
package commands

func newResumeCmd(app *pocketbase.PocketBase) *cobra.Command {
    var execFlag bool
    var jsonOutput bool
    var customPrompt string

    cmd := &cobra.Command{
        Use:   "resume <task-ref>",
        Short: "Resume work on a blocked task",
        Long: `Generate or execute the command to resume an AI agent session 
for a blocked task. This injects context from comments into the session.`,
        Example: `  egenskriven resume WRK-123           # Print command
  egenskriven resume WRK-123 --exec    # Execute resume
  egenskriven resume WRK-123 --json    # JSON output`,
        Args: cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            taskRef := args[0]

            task, err := resolver.MustResolve(app, taskRef)
            if err != nil {
                return err
            }

            // Validate task state
            if task.GetString("column") != "need_input" {
                return fmt.Errorf("task %s is not in need_input state", 
                    task.GetString("display_id"))
            }

            // Get session info
            sessionData := task.Get("agent_session")
            if sessionData == nil {
                return fmt.Errorf("no agent session linked to task %s", 
                    task.GetString("display_id"))
            }
            session := sessionData.(map[string]any)

            // Fetch comments
            comments, err := fetchComments(app, task.Id)
            if err != nil {
                return err
            }

            // Build context prompt
            prompt := customPrompt
            if prompt == "" {
                prompt = buildContextPrompt(task, comments)
            }

            // Build resume command
            resumeCmd := buildResumeCommand(
                session["tool"].(string),
                session["ref"].(string),
                prompt,
            )

            if jsonOutput {
                return outputJSON(map[string]any{
                    "task":        task.Id,
                    "tool":        session["tool"],
                    "session_ref": session["ref"],
                    "command":     resumeCmd,
                    "prompt":      prompt,
                })
            }

            if execFlag {
                // Move task to in_progress first
                task.Set("column", "in_progress")
                addHistoryEntry(task, "resumed", "user", nil)
                if err := app.Save(task); err != nil {
                    return err
                }

                // Execute the command
                fmt.Printf("Resuming session for %s...\n", 
                    task.GetString("display_id"))
                return executeResumeCommand(resumeCmd, session["working_dir"].(string))
            }

            // Just print the command
            fmt.Println("Run this command to resume:")
            fmt.Println()
            fmt.Printf("  %s\n", resumeCmd)
            fmt.Println()
            fmt.Println("Or use --exec to run it directly.")
            return nil
        },
    }

    cmd.Flags().BoolVarP(&execFlag, "exec", "e", false, "Execute the command")
    cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
    cmd.Flags().StringVarP(&customPrompt, "prompt", "p", "", "Custom prompt")

    return cmd
}

func buildResumeCommand(tool, ref, prompt string) string {
    // Escape prompt for shell
    escapedPrompt := shellescape.Quote(prompt)
    
    switch tool {
    case "opencode":
        return fmt.Sprintf("opencode run %s --session %s", escapedPrompt, ref)
    case "claude-code":
        return fmt.Sprintf("claude --resume %s %s", ref, escapedPrompt)
    case "codex":
        return fmt.Sprintf("codex exec resume %s %s", ref, escapedPrompt)
    default:
        return fmt.Sprintf("# Unknown tool: %s", tool)
    }
}

func buildContextPrompt(task *core.Record, comments []Comment) string {
    var sb strings.Builder
    
    sb.WriteString("## Task Context (from EgenSkriven)\n\n")
    sb.WriteString(fmt.Sprintf("**Task**: %s - %s\n", 
        task.GetString("display_id"), 
        task.GetString("title")))
    sb.WriteString(fmt.Sprintf("**Status**: need_input -> in_progress\n"))
    sb.WriteString(fmt.Sprintf("**Priority**: %s\n\n", 
        task.GetString("priority")))
    
    sb.WriteString("## Conversation Thread\n\n")
    for _, c := range comments {
        authorLabel := c.AuthorId
        if authorLabel == "" {
            authorLabel = c.AuthorType
        }
        sb.WriteString(fmt.Sprintf("[%s @ %s]: %s\n\n", 
            authorLabel, 
            c.Created.Format("15:04"),
            c.Content))
    }
    
    sb.WriteString("## Instructions\n\n")
    sb.WriteString("Continue working on the task based on the human's response above. ")
    sb.WriteString("The conversation context should help you understand what was discussed.\n")
    
    return sb.String()
}
```

### 7.6 `egenskriven init` Command Updates

Add tool integration initialization.

**New subcommands:**

```bash
egenskriven init --opencode      # Generate OpenCode custom tool
egenskriven init --claude-code   # Generate Claude Code hooks
egenskriven init --codex         # Generate Codex helper script
egenskriven init --all           # Generate all integrations
```

**Implementation:**

```go
func newInitCmd(app *pocketbase.PocketBase) *cobra.Command {
    var opencode, claudeCode, codex, all bool

    cmd := &cobra.Command{
        Use:   "init",
        Short: "Initialize EgenSkriven in current directory",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Existing init logic...
            
            if all {
                opencode, claudeCode, codex = true, true, true
            }
            
            if opencode {
                if err := generateOpenCodeTool(); err != nil {
                    return err
                }
                fmt.Println("Generated .opencode/tool/egenskriven-session.ts")
            }
            
            if claudeCode {
                if err := generateClaudeCodeHooks(); err != nil {
                    return err
                }
                fmt.Println("Generated .claude/hooks/session-register.sh")
                fmt.Println("Updated .claude/settings.json")
            }
            
            if codex {
                if err := generateCodexHelper(); err != nil {
                    return err
                }
                fmt.Println("Generated .codex/get-session-id.sh")
            }
            
            return nil
        },
    }

    cmd.Flags().BoolVar(&opencode, "opencode", false, "Generate OpenCode integration")
    cmd.Flags().BoolVar(&claudeCode, "claude-code", false, "Generate Claude Code integration")
    cmd.Flags().BoolVar(&codex, "codex", false, "Generate Codex integration")
    cmd.Flags().BoolVar(&all, "all", false, "Generate all tool integrations")

    return cmd
}
```

### 7.7 Updated `egenskriven list` Command

Add `--need-input` filter.

```bash
egenskriven list --need-input          # Tasks needing input
egenskriven list --need-input --json   # As JSON
```

**New flag:**
```go
cmd.Flags().BoolVar(&needInput, "need-input", false, "Show only tasks needing input")
```

---

## 8. Tool-Specific Integrations

### 8.1 OpenCode Integration

#### Custom Tool: `.opencode/tool/egenskriven-session.ts`

```typescript
import { tool } from "@opencode-ai/plugin"

export default tool({
  description: `Get current OpenCode session information for EgenSkriven task tracking.
  
Call this tool when you need to:
- Link this session to an EgenSkriven task
- Register your session before starting work
- Get your session ID for any reason`,
  args: {},
  async execute(args, context) {
    const { sessionID, messageID, agent } = context
    
    return JSON.stringify({
      tool: "opencode",
      session_id: sessionID,
      message_id: messageID,
      agent: agent,
      instructions: `To link this session to a task, run:
egenskriven session link <task-ref> --tool opencode --ref ${sessionID}`
    }, null, 2)
  },
})
```

#### Generation Function

```go
func generateOpenCodeTool() error {
    toolDir := ".opencode/tool"
    if err := os.MkdirAll(toolDir, 0755); err != nil {
        return err
    }

    toolContent := `import { tool } from "@opencode-ai/plugin"

export default tool({
  description: \`Get current OpenCode session ID for EgenSkriven task tracking.
  
Call this when you need to link this session to an EgenSkriven task.\`,
  args: {},
  async execute(args, context) {
    return JSON.stringify({
      tool: "opencode",
      session_id: context.sessionID,
      link_command: \`egenskriven session link <task> --tool opencode --ref \${context.sessionID}\`
    }, null, 2)
  },
})
`
    return os.WriteFile(
        filepath.Join(toolDir, "egenskriven-session.ts"),
        []byte(toolContent),
        0644,
    )
}
```

### 8.2 Claude Code Integration

#### Hook Script: `.claude/hooks/session-register.sh`

```bash
#!/bin/bash
# EgenSkriven session registration hook for Claude Code
# This hook persists the session ID as an environment variable

# Read JSON input from stdin
INPUT=$(cat)

# Extract session_id using jq (or Python fallback)
if command -v jq &> /dev/null; then
    SESSION_ID=$(echo "$INPUT" | jq -r '.session_id')
    HOOK_EVENT=$(echo "$INPUT" | jq -r '.hook_event_name')
else
    SESSION_ID=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('session_id',''))")
    HOOK_EVENT=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('hook_event_name',''))")
fi

# Only process SessionStart events
if [ "$HOOK_EVENT" != "SessionStart" ]; then
    exit 0
fi

# Persist session ID for subsequent Bash tool calls
if [ -n "$CLAUDE_ENV_FILE" ] && [ -n "$SESSION_ID" ]; then
    echo "export CLAUDE_SESSION_ID=$SESSION_ID" >> "$CLAUDE_ENV_FILE"
fi

# Output success (optional context for Claude)
echo '{"status": "registered", "session_id": "'$SESSION_ID'"}'
exit 0
```

#### Settings Update: `.claude/settings.json`

```json
{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "startup|resume",
        "hooks": [
          {
            "type": "command",
            "command": "bash \"$CLAUDE_PROJECT_DIR/.claude/hooks/session-register.sh\""
          }
        ]
      }
    ]
  }
}
```

#### Generation Function

```go
func generateClaudeCodeHooks() error {
    hooksDir := ".claude/hooks"
    if err := os.MkdirAll(hooksDir, 0755); err != nil {
        return err
    }

    // Write hook script
    hookScript := `#!/bin/bash
# EgenSkriven session registration hook for Claude Code

INPUT=$(cat)

if command -v jq &> /dev/null; then
    SESSION_ID=$(echo "$INPUT" | jq -r '.session_id')
else
    SESSION_ID=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('session_id',''))")
fi

if [ -n "$CLAUDE_ENV_FILE" ] && [ -n "$SESSION_ID" ]; then
    echo "export CLAUDE_SESSION_ID=$SESSION_ID" >> "$CLAUDE_ENV_FILE"
fi

exit 0
`
    hookPath := filepath.Join(hooksDir, "session-register.sh")
    if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
        return err
    }

    // Update settings.json
    settingsPath := ".claude/settings.json"
    settings := loadOrCreateSettings(settingsPath)
    
    // Add hook configuration
    settings["hooks"] = map[string]any{
        "SessionStart": []map[string]any{
            {
                "matcher": "startup|resume",
                "hooks": []map[string]any{
                    {
                        "type":    "command",
                        "command": `bash "$CLAUDE_PROJECT_DIR/.claude/hooks/session-register.sh"`,
                    },
                },
            },
        },
    }

    return saveSettings(settingsPath, settings)
}
```

### 8.3 Codex Integration

#### Helper Script: `.codex/get-session-id.sh`

```bash
#!/bin/bash
# EgenSkriven session ID discovery for Codex CLI
# 
# Codex does not expose session ID via environment variable.
# This script extracts it from the most recent rollout file.

CODEX_DIR="${CODEX_HOME:-$HOME/.codex}"
SESSIONS_DIR="$CODEX_DIR/sessions"

# Find most recent rollout file
LATEST_ROLLOUT=$(ls -t "$SESSIONS_DIR"/rollout-*.jsonl 2>/dev/null | head -1)

if [ -z "$LATEST_ROLLOUT" ]; then
    echo "ERROR: No Codex session found" >&2
    exit 1
fi

# Extract UUID from filename
# Format: rollout-2025-05-07T17-24-21-5973b6c0-94b8-487b-a530-2aeb6098ae0e.jsonl
SESSION_ID=$(basename "$LATEST_ROLLOUT" | grep -oP '[0-9a-f]{8}(-[0-9a-f]{4}){3}-[0-9a-f]{12}')

if [ -z "$SESSION_ID" ]; then
    echo "ERROR: Could not extract session ID" >&2
    exit 1
fi

echo "$SESSION_ID"
```

#### Generation Function

```go
func generateCodexHelper() error {
    codexDir := ".codex"
    if err := os.MkdirAll(codexDir, 0755); err != nil {
        return err
    }

    script := `#!/bin/bash
# EgenSkriven session ID discovery for Codex CLI

CODEX_DIR="${CODEX_HOME:-$HOME/.codex}"
SESSIONS_DIR="$CODEX_DIR/sessions"

LATEST_ROLLOUT=$(ls -t "$SESSIONS_DIR"/rollout-*.jsonl 2>/dev/null | head -1)

if [ -z "$LATEST_ROLLOUT" ]; then
    echo "ERROR: No Codex session found" >&2
    exit 1
fi

SESSION_ID=$(basename "$LATEST_ROLLOUT" | grep -oP '[0-9a-f]{8}(-[0-9a-f]{4}){3}-[0-9a-f]{12}')

if [ -z "$SESSION_ID" ]; then
    echo "ERROR: Could not extract session ID" >&2
    exit 1
fi

echo "$SESSION_ID"
`
    return os.WriteFile(
        filepath.Join(codexDir, "get-session-id.sh"),
        []byte(script),
        0755,
    )
}
```

---

## 9. Resume Flow Implementation

### 9.1 Context Injection Format

When resuming, the following context is injected:

```markdown
## Task Context (from EgenSkriven)

**Task**: WRK-123 - Implement user authentication
**Status**: need_input -> in_progress  
**Priority**: high

## Conversation Thread

[agent @ 10:30]: I'm implementing authentication and need guidance. What approach should I use - JWT tokens with refresh tokens, or session-based authentication with cookies?

[human @ 11:45]: Use JWT with refresh tokens. The refresh token should have a 7-day expiry, and the access token should be 15 minutes. Store refresh tokens in HttpOnly cookies.

[human @ 11:47]: Also make sure to implement token rotation on refresh.

## Instructions

Continue working on the task based on the human's response above. The conversation context should help you understand what was discussed.
```

### 9.2 Resume Mode Behaviors

#### Manual Mode

```
User runs: egenskriven resume WRK-123

Output:
  Run this command to resume:

    opencode run "## Task Context..." --session abc-123

  Or use --exec to run it directly.

[User copies and runs command manually]
```

#### Command Mode (Default)

```
User runs: egenskriven resume WRK-123 --exec

1. Task moves to in_progress
2. History entry added: "resumed by user"
3. Tool process spawned with context prompt
4. Agent continues work
```

#### Auto Mode

```
User adds comment: "@agent I've decided to use JWT"

1. Comment saved with mention detected
2. System detects @agent mention on need_input task
3. Automatically triggers resume flow
4. Task moves to in_progress
5. Agent session resumed with context
```

### 9.3 Auto-Resume Implementation

For `auto` mode, implement a watcher or hook:

```go
// In comment creation, check for auto-resume trigger
func onCommentCreated(app core.App, comment *core.Record) error {
    task, err := app.FindRecordById("tasks", comment.GetString("task"))
    if err != nil {
        return err
    }

    // Check conditions for auto-resume
    if task.GetString("column") != "need_input" {
        return nil
    }

    board, err := app.FindRecordById("boards", task.GetString("board"))
    if err != nil {
        return err
    }

    if board.GetString("resume_mode") != "auto" {
        return nil
    }

    // Check for @agent mention
    metadata := comment.Get("metadata").(map[string]any)
    mentions, ok := metadata["mentions"].([]string)
    if !ok || !contains(mentions, "@agent") {
        return nil
    }

    // Trigger resume
    return triggerAutoResume(app, task, comment)
}

func triggerAutoResume(app core.App, task *core.Record, triggerComment *core.Record) error {
    session := task.Get("agent_session").(map[string]any)
    
    comments, err := fetchComments(app, task.Id)
    if err != nil {
        return err
    }

    prompt := buildContextPrompt(task, comments)
    resumeCmd := buildResumeCommand(
        session["tool"].(string),
        session["ref"].(string),
        prompt,
    )

    // Move task
    task.Set("column", "in_progress")
    addHistoryEntry(task, "auto_resumed", "system", map[string]any{
        "trigger_comment": triggerComment.Id,
    })
    if err := app.Save(task); err != nil {
        return err
    }

    // Execute in background
    go executeResumeCommand(resumeCmd, session["working_dir"].(string))
    
    return nil
}
```

---

## 10. Web UI Changes

### 10.1 Task Detail - Comments Panel

Add a comments section to the task detail view.

**File: `ui/src/components/TaskDetail/CommentsPanel.tsx`**

```tsx
import React, { useState } from 'react';
import { useComments, useAddComment } from '@/hooks/useComments';
import { formatDistanceToNow } from 'date-fns';

interface CommentsPanelProps {
  taskId: string;
}

export function CommentsPanel({ taskId }: CommentsPanelProps) {
  const { comments, isLoading } = useComments(taskId);
  const addComment = useAddComment();
  const [newComment, setNewComment] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newComment.trim()) return;
    
    await addComment.mutateAsync({
      taskId,
      content: newComment,
      authorType: 'human',
    });
    setNewComment('');
  };

  if (isLoading) {
    return <div className="p-4">Loading comments...</div>;
  }

  return (
    <div className="border-t border-gray-200 dark:border-gray-700">
      <h3 className="px-4 py-2 font-semibold text-sm text-gray-600 dark:text-gray-400">
        Comments ({comments.length})
      </h3>
      
      {/* Comments list */}
      <div className="max-h-64 overflow-y-auto">
        {comments.map((comment) => (
          <div 
            key={comment.id}
            className={`px-4 py-3 border-b border-gray-100 dark:border-gray-800 ${
              comment.author_type === 'agent' 
                ? 'bg-blue-50 dark:bg-blue-900/20' 
                : ''
            }`}
          >
            <div className="flex items-center gap-2 mb-1">
              <span className={`text-xs font-medium px-2 py-0.5 rounded ${
                comment.author_type === 'agent'
                  ? 'bg-blue-100 text-blue-700 dark:bg-blue-800 dark:text-blue-200'
                  : 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-200'
              }`}>
                {comment.author_type === 'agent' ? 'Agent' : 'Human'}
              </span>
              {comment.author_id && (
                <span className="text-xs text-gray-500">
                  {comment.author_id}
                </span>
              )}
              <span className="text-xs text-gray-400">
                {formatDistanceToNow(new Date(comment.created), { addSuffix: true })}
              </span>
            </div>
            <p className="text-sm whitespace-pre-wrap">{comment.content}</p>
          </div>
        ))}
        
        {comments.length === 0 && (
          <div className="px-4 py-8 text-center text-gray-500 text-sm">
            No comments yet
          </div>
        )}
      </div>

      {/* Add comment form */}
      <form onSubmit={handleSubmit} className="p-4 border-t border-gray-200 dark:border-gray-700">
        <textarea
          value={newComment}
          onChange={(e) => setNewComment(e.target.value)}
          placeholder="Add a comment... (use @agent to trigger auto-resume)"
          className="w-full px-3 py-2 text-sm border rounded-lg resize-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:border-gray-700"
          rows={3}
        />
        <div className="flex justify-end mt-2">
          <button
            type="submit"
            disabled={!newComment.trim() || addComment.isPending}
            className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50"
          >
            {addComment.isPending ? 'Adding...' : 'Add Comment'}
          </button>
        </div>
      </form>
    </div>
  );
}
```

### 10.2 Task Detail - Session Info

Display linked session information.

**File: `ui/src/components/TaskDetail/SessionInfo.tsx`**

```tsx
import React from 'react';
import { Task } from '@/types';

interface SessionInfoProps {
  task: Task;
}

export function SessionInfo({ task }: SessionInfoProps) {
  const session = task.agent_session;
  
  if (!session) {
    return null;
  }

  const toolIcons: Record<string, string> = {
    'opencode': '⚡',
    'claude-code': '🤖',
    'codex': '🔮',
  };

  return (
    <div className="px-4 py-3 bg-gray-50 dark:bg-gray-800/50 border-t border-gray-200 dark:border-gray-700">
      <h4 className="text-xs font-semibold text-gray-500 uppercase mb-2">
        Agent Session
      </h4>
      <div className="flex items-center gap-3">
        <span className="text-lg">{toolIcons[session.tool] || '🔧'}</span>
        <div>
          <div className="text-sm font-medium">
            {session.tool}
          </div>
          <div className="text-xs text-gray-500 font-mono">
            {session.ref.slice(0, 8)}...
          </div>
        </div>
      </div>
    </div>
  );
}
```

### 10.3 Task Card - Need Input Badge

Show visual indicator for tasks needing input.

**Update: `ui/src/components/TaskCard.tsx`**

```tsx
// Add to task card component
{task.column === 'need_input' && (
  <div className="absolute top-2 right-2">
    <span className="flex h-3 w-3">
      <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-orange-400 opacity-75"></span>
      <span className="relative inline-flex rounded-full h-3 w-3 bg-orange-500"></span>
    </span>
  </div>
)}
```

### 10.4 Resume Button

Add resume button to task detail for `need_input` tasks.

```tsx
{task.column === 'need_input' && task.agent_session && (
  <button
    onClick={handleResume}
    className="w-full mt-4 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 flex items-center justify-center gap-2"
  >
    <PlayIcon className="w-4 h-4" />
    Resume Agent Session
  </button>
)}
```

### 10.5 Hooks for Comments

**File: `ui/src/hooks/useComments.ts`**

```typescript
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { pb } from '@/lib/pocketbase';

export interface Comment {
  id: string;
  task: string;
  content: string;
  author_type: 'human' | 'agent';
  author_id?: string;
  metadata?: {
    mentions?: string[];
  };
  created: string;
}

export function useComments(taskId: string) {
  return useQuery({
    queryKey: ['comments', taskId],
    queryFn: async () => {
      const records = await pb.collection('comments').getFullList<Comment>({
        filter: `task = "${taskId}"`,
        sort: 'created',
      });
      return records;
    },
  });
}

export function useAddComment() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (data: {
      taskId: string;
      content: string;
      authorType: 'human' | 'agent';
    }) => {
      // Extract mentions
      const mentions = data.content.match(/@\w+/g) || [];
      
      return pb.collection('comments').create({
        task: data.taskId,
        content: data.content,
        author_type: data.authorType,
        metadata: { mentions },
      });
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['comments', variables.taskId] });
    },
  });
}

// Real-time subscription for comments
export function useCommentsSubscription(taskId: string) {
  const queryClient = useQueryClient();
  
  useEffect(() => {
    const unsubscribe = pb.collection('comments').subscribe('*', (e) => {
      if (e.record.task === taskId) {
        queryClient.invalidateQueries({ queryKey: ['comments', taskId] });
      }
    });
    
    return () => {
      unsubscribe.then(unsub => unsub());
    };
  }, [taskId, queryClient]);
}
```

---

## 11. Skills & Documentation

### 11.1 Update `egenskriven` Skill

**File: `internal/commands/skills/egenskriven/SKILL.md`**

Add new section:

```markdown
## Human-AI Collaborative Workflow

EgenSkriven supports a collaborative workflow where you can request human input
and resume work once you receive a response.

### Requesting Human Input

When you're blocked and need human guidance:

\`\`\`bash
# Block task with question (atomic operation)
egenskriven block <task-ref> "Your question for the human"

# Example
egenskriven block WRK-123 "Should I use JWT or session-based authentication?"
\`\`\`

This will:
1. Move the task to the `need_input` column
2. Add your question as a comment
3. Preserve your session for later resume

### Linking Your Session

Before you can be resumed, link your session to the task:

**For OpenCode:**
\`\`\`bash
# Call the egenskriven-session tool to get your session ID
# Then link it:
egenskriven session link <task-ref> --tool opencode --ref <session-id>
\`\`\`

**For Claude Code:**
\`\`\`bash
# Session ID is available as $CLAUDE_SESSION_ID after the hook runs
egenskriven session link <task-ref> --tool claude-code --ref $CLAUDE_SESSION_ID
\`\`\`

**For Codex:**
\`\`\`bash
# Get session ID from helper script
SESSION_ID=$(.codex/get-session-id.sh)
egenskriven session link <task-ref> --tool codex --ref $SESSION_ID
\`\`\`

### Viewing Comments

Check for human responses:

\`\`\`bash
egenskriven comments <task-ref>
egenskriven comments <task-ref> --json
\`\`\`

### Resume Flow

When the human responds, they will either:
1. Run `egenskriven resume <task-ref>` to continue your session
2. Add a comment with `@agent` to auto-trigger resume (if configured)

Your session will be resumed with full context from the comment thread.
```

### 11.2 Update `egenskriven-workflows` Skill

**File: `internal/commands/skills/egenskriven-workflows/SKILL.md`**

Add new section:

```markdown
## Resume Modes

Each board can be configured with a resume mode that controls how blocked tasks
are resumed after human input.

### Available Modes

| Mode | Behavior |
|------|----------|
| `manual` | Print resume command for user to copy |
| `command` | User runs `egenskriven resume --exec` |
| `auto` | Auto-resume when comment contains `@agent` |

### Configuration

Resume mode is set per-board:

\`\`\`bash
# Set resume mode for a board
egenskriven board update <board> --resume-mode auto
\`\`\`

### Collaborative Workflow Best Practices

1. **Always link your session** before blocking a task
2. **Be specific** in your blocking question
3. **Check comments** periodically if working on multiple tasks
4. **Use atomic operations** - prefer `egenskriven block` over separate move + comment
```

### 11.3 Update Prime Template

**File: `internal/commands/prime.tmpl`**

Add to collaborative mode section:

```markdown
## Requesting Human Input

When you encounter a decision that requires human judgment or you're blocked:

1. **Link your session first** (if not already linked):
   - OpenCode: Call the `egenskriven-session` tool, then run the link command
   - Claude Code: Use `$CLAUDE_SESSION_ID` environment variable
   - Codex: Run `.codex/get-session-id.sh`

2. **Block the task with your question**:
   \`\`\`bash
   egenskriven block <task-ref> "Your specific question"
   \`\`\`

3. **Wait for human response** - your session will be preserved

4. **You will be resumed** with full context when the human responds

### Example Workflow

\`\`\`bash
# 1. Get and link session (OpenCode example)
# [Call egenskriven-session tool]
egenskriven session link WRK-123 --tool opencode --ref <your-session-id>

# 2. Work on task...
egenskriven move WRK-123 in_progress

# 3. Hit a decision point - block with question
egenskriven block WRK-123 "The API supports both REST and GraphQL. Which should I implement first?"

# 4. Session ends, human will respond and resume you
\`\`\`

When resumed, you'll receive context like:
\`\`\`
## Task Context (from EgenSkriven)
Task: WRK-123 - Implement API endpoints
...
## Conversation Thread
[agent]: The API supports both REST and GraphQL. Which should I implement first?
[human]: Start with REST endpoints, we can add GraphQL later.
\`\`\`
```

### 11.4 Update AGENTS.md

**File: `AGENTS.md`**

Add section:

```markdown
## Human-AI Collaborative Workflow

This project supports a collaborative workflow for AI agents.

### When Blocked

If you need human input:

\`\`\`bash
# 1. Link your session (one-time per task)
egenskriven session link <task> --tool <your-tool> --ref <session-id>

# 2. Block with your question
egenskriven block <task> "Your question here"
\`\`\`

### Session ID Discovery

- **OpenCode**: Use the `egenskriven-session` custom tool
- **Claude Code**: `$CLAUDE_SESSION_ID` environment variable
- **Codex**: Run `.codex/get-session-id.sh`

### Resume

The human will resume you with context. Check:
\`\`\`bash
egenskriven comments <task>
\`\`\`
```

---

## 12. Implementation Phases

### Phase 1: Foundation (Week 1)

**Goal**: Core data model and basic CLI commands

**Tasks**:
1. [ ] Create migration: Add `need_input` to column values
2. [ ] Create migration: Add `agent_session` field to tasks
3. [ ] Create migration: Create `comments` collection
4. [ ] Create migration: Create `sessions` collection
5. [ ] Update `ValidColumns` in root.go
6. [ ] Implement `egenskriven block` command
7. [ ] Implement `egenskriven comment` command
8. [ ] Implement `egenskriven comments` command
9. [ ] Update `egenskriven list` with `--need-input` flag
10. [ ] Write unit tests for new commands

**Deliverables**:
- Database schema updated
- Basic CLI workflow functional
- Agent can block task and add comment

### Phase 2: Session Management (Week 2)

**Goal**: Session linking and resume command

**Tasks**:
1. [ ] Implement `egenskriven session link` command
2. [ ] Implement `egenskriven session show` command
3. [ ] Implement `egenskriven session history` command
4. [ ] Implement `egenskriven resume` command (print mode)
5. [ ] Implement `egenskriven resume --exec` functionality
6. [ ] Implement context prompt builder
7. [ ] Add resume mode to boards schema
8. [ ] Write unit tests for session commands

**Deliverables**:
- Full session management via CLI
- Resume command working for all three tools

### Phase 3: Tool Integrations (Week 3)

**Goal**: Auto-generated tool integrations

**Tasks**:
1. [ ] Implement `egenskriven init --opencode`
2. [ ] Implement `egenskriven init --claude-code`
3. [ ] Implement `egenskriven init --codex`
4. [ ] Implement `egenskriven init --all`
5. [ ] Test OpenCode custom tool
6. [ ] Test Claude Code hooks
7. [ ] Test Codex helper script
8. [ ] Update skills documentation

**Deliverables**:
- One-command setup for each tool
- Verified session ID discovery for all tools

### Phase 4: Web UI (Week 4)

**Goal**: Comments panel and UI enhancements

**Tasks**:
1. [ ] Create `CommentsPanel` component
2. [ ] Create `SessionInfo` component
3. [ ] Add comments hooks with real-time subscription
4. [ ] Add need_input visual indicator to task cards
5. [ ] Add resume button to task detail
6. [ ] Update kanban board to show need_input column
7. [ ] Style need_input column distinctively
8. [ ] Write component tests

**Deliverables**:
- Full Web UI support for collaborative workflow
- Real-time comment updates

### Phase 5: Auto-Resume & Polish (Week 5)

**Goal**: Auto-resume mode and final polish

**Tasks**:
1. [ ] Implement `@agent` mention detection
2. [ ] Implement auto-resume trigger on comment creation
3. [ ] Add board resume_mode configuration UI
4. [ ] Update prime template
5. [ ] Update all skills documentation
6. [ ] Update AGENTS.md
7. [ ] End-to-end testing with all three tools
8. [ ] Performance testing
9. [ ] Documentation review

**Deliverables**:
- Complete feature set
- Production-ready implementation

---

## 13. Testing Strategy

### 13.1 Unit Tests

**Commands**:
```go
// internal/commands/block_test.go
func TestBlockCommand(t *testing.T) {
    // Test successful block
    // Test missing question
    // Test invalid task ref
    // Test already blocked task
}

// internal/commands/comment_test.go
func TestCommentCommand(t *testing.T) {
    // Test add comment
    // Test stdin input
    // Test mention extraction
}

// internal/commands/resume_test.go
func TestResumeCommand(t *testing.T) {
    // Test context building
    // Test command generation for each tool
    // Test invalid state errors
}
```

### 13.2 Integration Tests

**Workflow tests**:
```go
func TestBlockAndResumeWorkflow(t *testing.T) {
    // 1. Create task
    // 2. Link session
    // 3. Block with question
    // 4. Verify task in need_input
    // 5. Add human comment
    // 6. Resume
    // 7. Verify task in in_progress
    // 8. Verify context prompt contains comments
}
```

### 13.3 End-to-End Tests

**For each tool (OpenCode, Claude Code, Codex)**:
1. Initialize tool integration
2. Start session
3. Link session to task
4. Block task
5. Add human response
6. Resume session
7. Verify agent receives context

### 13.4 UI Tests

**Playwright tests**:
```typescript
test('comments panel displays and updates', async ({ page }) => {
    // Navigate to task detail
    // Verify comments panel
    // Add comment
    // Verify comment appears
    // Verify real-time update
});

test('resume button triggers resume', async ({ page }) => {
    // Create blocked task
    // Navigate to task
    // Click resume button
    // Verify task moves to in_progress
});
```

---

## 14. Appendix

### 14.1 Full Command Reference

| Command | Description |
|---------|-------------|
| `egenskriven block <task> "question"` | Block task and add question |
| `egenskriven comment <task> "text"` | Add comment to task |
| `egenskriven comments <task>` | List comments |
| `egenskriven session link <task> --tool X --ref Y` | Link session |
| `egenskriven session show <task>` | Show session info |
| `egenskriven session history <task>` | Show session history |
| `egenskriven resume <task>` | Print resume command |
| `egenskriven resume <task> --exec` | Execute resume |
| `egenskriven list --need-input` | List blocked tasks |
| `egenskriven init --opencode` | Generate OpenCode integration |
| `egenskriven init --claude-code` | Generate Claude Code integration |
| `egenskriven init --codex` | Generate Codex integration |

### 14.2 Environment Variables

| Variable | Description |
|----------|-------------|
| `EGENSKRIVEN_AUTHOR` | Default comment author |
| `CLAUDE_SESSION_ID` | Claude Code session (set by hook) |

### 14.3 Configuration Options

**Board settings**:
```json
{
    "resume_mode": "command"  // "manual" | "command" | "auto"
}
```

**Agent config**:
```json
{
    "agent": {
        "workflow": "light",
        "mode": "collaborative"
    }
}
```

### 14.4 API Endpoints (PocketBase)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/collections/comments/records` | List comments |
| POST | `/api/collections/comments/records` | Create comment |
| GET | `/api/collections/sessions/records` | List sessions |
| POST | `/api/collections/sessions/records` | Create session |

### 14.5 Glossary

| Term | Definition |
|------|------------|
| **Block** | Move task to `need_input` state |
| **Resume** | Continue agent work on blocked task |
| **Session** | AI tool's conversation context |
| **Context Injection** | Providing comments as prompt on resume |
| **Mention** | `@agent` or `@user` in comment text |

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2026-01-07 | AI + Human | Initial comprehensive plan |
