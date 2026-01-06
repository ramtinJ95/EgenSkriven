# AI-Native UX Design for EgenSkriven

A comprehensive guide for implementing AI-native features including skills, AGENTS.md, and cross-agent integration. This document captures research findings and design decisions for making EgenSkriven the primary task tracker for AI coding agents.

---

## Table of Contents

1. [Philosophy & Goals](#1-philosophy--goals)
2. [The Skills Approach](#2-the-skills-approach)
3. [Skill Architecture for EgenSkriven](#3-skill-architecture-for-egenskriven)
4. [SKILL.md Format Specification](#4-skillmd-format-specification)
5. [Skill Content Outlines](#5-skill-content-outlines)
6. [AGENTS.md Integration](#6-agentsmd-integration)
7. [The `skill install` Command](#7-the-skill-install-command)
8. [Context Window Optimization](#8-context-window-optimization)
9. [Prime Command Relationship](#9-prime-command-relationship)
10. [Real-World Patterns](#10-real-world-patterns-research-findings)
11. [Implementation Checklist](#11-implementation-checklist)
12. [Future Considerations](#12-future-considerations)

---

## 1. Philosophy & Goals

### Core Philosophy

EgenSkriven is designed to be **agent-native from the ground up**. This means AI coding assistants should naturally use EgenSkriven as their primary task tracker, replacing built-in systems like `TodoWrite`.

### Primary Goals

1. **Replace Built-in Agent Todo Systems**
   - Agents like Claude Code have internal todo tools (TodoWrite) that lack persistence
   - EgenSkriven provides persistent, structured task management
   - The goal is for agents to automatically prefer EgenSkriven over ephemeral alternatives

2. **On-Demand Context Loading**
   - Instructions should not bloat every conversation
   - Skills load only when the agent determines they're needed
   - Minimal always-in-context overhead

3. **Token Efficiency**
   - Modern context windows are large (128k-200k tokens) but performance degrades with bloat
   - Target: core skill under 1k tokens, total skills under 5k tokens
   - Follow the 80% rule: optimize before hitting 80% context utilization

4. **Cross-Agent Compatibility**
   - Support Claude Code, OpenCode, Cursor, Codex, Jules, and future agents
   - Use universal standards (AGENTS.md) where possible
   - Single skill location that works for multiple agents

### Design Principles

| Principle | Description |
|-----------|-------------|
| **Lazy Loading** | Load instructions only when needed |
| **Progressive Disclosure** | Start simple, expand on demand |
| **Universal Formats** | Use standards that work across agents |
| **Graceful Degradation** | Work even if skills aren't loaded |
| **Token Consciousness** | Every token has a cost |

---

## 2. The Skills Approach

### Why Skills Over Plugins/Hooks

We evaluated three approaches for agent integration:

| Approach | Pros | Cons |
|----------|------|------|
| **Plugins** | Deep integration, event hooks | Agent-specific, requires separate implementations |
| **Hooks (prime command)** | Always injected, guaranteed context | Token overhead on every session |
| **Skills** | On-demand, cross-platform, discoverable | Agent must decide to load |

**Decision: Skills as primary, prime as fallback**

Skills provide the best balance of token efficiency and cross-platform support. The existing `prime` command remains for hook-based injection when needed.

### Cross-Platform Compatibility

A key discovery from research: **OpenCode reads from `.claude/skills/` for Claude compatibility**.

This means a single skill location works for both major agents:

```
.claude/skills/egenskriven/SKILL.md
```

| Agent | Reads From | Notes |
|-------|------------|-------|
| Claude Code | `.claude/skills/` | Native location |
| OpenCode | `.opencode/skill/` AND `.claude/skills/` | Claude-compatible |
| Global (Claude) | `~/.claude/skills/` | User-wide installation |
| Global (OpenCode) | `~/.config/opencode/skill/` | User-wide installation |

### How Skills Work

1. **Discovery**: Agent sees available skills listed in tool description
2. **Selection**: Agent decides when a skill is relevant based on description
3. **Loading**: Agent calls `skill({ name: "egenskriven" })` to load full content
4. **Context**: Skill content becomes part of conversation context
5. **Persistence**: Content remains until session ends or compaction

### Skill Discovery Example

When an agent starts, it sees something like:

```xml
<available_skills>
  <skill>
    <name>egenskriven</name>
    <description>Local-first kanban task manager for AI agent workflows. 
    Use when managing tasks, tracking work, or when user mentions 
    tasks, kanban, boards, or egenskriven.</description>
  </skill>
  <skill>
    <name>egenskriven-workflows</name>
    <description>Workflow modes and agent behaviors for EgenSkriven. 
    Use when configuring strict/light/minimal modes or understanding 
    agent autonomous/collaborative/supervised behaviors.</description>
  </skill>
</available_skills>
```

The agent then decides when to load each skill based on the conversation.

---

## 3. Skill Architecture for EgenSkriven

### Three-Skill Structure

Based on research into best practices (Metabase, PyTorch, Anthropic guidelines), we use a **three-skill architecture**:

```
.claude/skills/
├── egenskriven/
│   └── SKILL.md              # Core (~800-1k tokens)
├── egenskriven-workflows/
│   └── SKILL.md              # Workflow modes (~1.5-2k tokens)
└── egenskriven-advanced/
    └── SKILL.md              # Advanced features (~1.5-2k tokens)
```

### Rationale for Split

| Consideration | Single Skill | Three Skills |
|---------------|--------------|--------------|
| Token efficiency | All or nothing | Load what's needed |
| Maintenance | One large file | Focused, smaller files |
| Agent selection | Always loads everything | Targeted loading |
| User needs | Simple users get bloat | Appropriate complexity |

**Research finding**: Anthropic recommends keeping SKILL.md under 500 lines for optimal performance. Splitting allows each skill to stay well under this limit.

### Skill Overview

#### `egenskriven` (Core Skill)

**Purpose**: Essential commands for basic task management

**Target size**: ~800-1,000 tokens (~150-200 lines)

**When agent loads**: 
- User mentions "task", "kanban", "board", "egenskriven"
- User wants to track, create, or manage work items
- Starting a new coding session with task context

**Content focus**:
- What is EgenSkriven (brief)
- Quick start commands
- Essential CRUD operations
- Task reference formats
- JSON output basics

#### `egenskriven-workflows` (Workflow Skill)

**Purpose**: Understanding and configuring workflow modes

**Target size**: ~1,500-2,000 tokens (~250-350 lines)

**When agent loads**:
- User mentions "workflow", "mode", "strict", "light", "minimal"
- User wants to configure agent behavior
- Agent needs to understand autonomous vs collaborative modes
- Using the `prime` command

**Content focus**:
- Workflow modes (strict/light/minimal) explained
- Agent modes (autonomous/collaborative/supervised)
- When to create vs complete tasks
- Session patterns and best practices
- Prime command usage

#### `egenskriven-advanced` (Advanced Skill)

**Purpose**: Complex features beyond basic task management

**Target size**: ~1,500-2,000 tokens (~250-350 lines)

**When agent loads**:
- User mentions "epic", "blocked", "dependency", "sub-task"
- Working with task relationships
- Batch operations needed
- Views and filtering required

**Content focus**:
- Epics and epic management
- Task dependencies (blocked_by, blocking)
- Sub-tasks and parent relationships
- Batch operations (--stdin, --file)
- Views, filters, and search
- Import/export operations

### Token Budget Summary

| Skill | Tokens | % of 200k Context |
|-------|--------|-------------------|
| Core only | ~1,000 | 0.5% |
| Core + Workflows | ~3,000 | 1.5% |
| All three | ~5,000 | 2.5% |

This is well within the recommended 1-5% budget for instructions.

---

## 4. SKILL.md Format Specification

### Required YAML Frontmatter

Every SKILL.md file must begin with YAML frontmatter:

```yaml
---
name: skill-name
description: What this skill does and when to use it (max 1024 chars)
---
```

### Optional Frontmatter Fields

```yaml
---
name: skill-name
description: Description with trigger keywords
license: MIT
version: 1.0.0
compatibility: opencode
metadata:
  category: task-management
  agent-native: true
allowed-tools: Read, Grep, Bash
---
```

### Name Validation Rules

The `name` field has strict requirements:

| Rule | Valid | Invalid |
|------|-------|---------|
| Length | 1-64 characters | Empty or >64 chars |
| Characters | Lowercase alphanumeric + hyphens | Uppercase, underscores, spaces |
| Hyphens | Single hyphens between words | Starting/ending hyphens, double hyphens |
| Match directory | `name` must match parent directory name | Mismatched names |

**Regex**: `^[a-z0-9]+(-[a-z0-9]+)*$`

**Examples**:
- `egenskriven` - Valid
- `egenskriven-workflows` - Valid
- `Egenskriven` - Invalid (uppercase)
- `egen_skriven` - Invalid (underscore)
- `egenskriven--advanced` - Invalid (double hyphen)

### Description Best Practices

The description is critical because agents use it to decide when to load the skill.

**Formula**: `[What it does] + [When to use it] + [Trigger keywords]`

**Good example**:
```yaml
description: Local-first kanban task manager for AI agent workflows. 
  Use when managing tasks, tracking work, creating issues, or when 
  user mentions tasks, kanban, boards, backlog, or egenskriven.
```

**Bad example**:
```yaml
description: Task management tool
```

**Include trigger keywords**:
- Action words: "create task", "list tasks", "track work"
- Domain words: "kanban", "board", "backlog", "sprint"
- Product name: "egenskriven"

### Content Structure

After the frontmatter, structure content for clarity:

```markdown
---
name: example-skill
description: ...
---

## Overview
Brief introduction (2-3 sentences)

## When to Use
- Bullet points of scenarios
- Trigger conditions

## Quick Reference
Essential commands or patterns

## Detailed Instructions
Step-by-step guidance

## Examples
Concrete usage examples

## Common Issues
Troubleshooting tips

## Checklist
Verification steps
```

### File Size Guidelines

| Metric | Recommended | Maximum |
|--------|-------------|---------|
| Lines | 150-350 | 500 |
| Tokens | 1,000-3,000 | 5,000 |
| File size | 5-15 KB | 25 KB |

---

## 5. Skill Content Outlines

### 5.1 `egenskriven` (Core Skill)

**Frontmatter**:
```yaml
---
name: egenskriven
description: Local-first kanban task manager for AI agent workflows. 
  Use when managing tasks, tracking work, creating issues, or when 
  user mentions tasks, kanban, boards, backlog, or egenskriven.
---
```

**Content outline**:

- **What is EgenSkriven**
  - Local-first kanban with CLI and web UI
  - Designed for AI agents from the ground up
  - Replaces ephemeral TodoWrite with persistent tracking
  - Single binary, no external dependencies

- **Quick Start Commands**
  - `egenskriven context --json` - Get project state summary
  - `egenskriven suggest --json` - Get recommended next task
  - `egenskriven list --ready` - List actionable (unblocked) tasks
  - `egenskriven add "title"` - Create a new task

- **Essential CRUD Operations**
  - `add` - Create tasks with --type, --priority, --column flags
  - `list` - List tasks with filtering options
  - `show <ref>` - Display task details
  - `move <ref> <column>` - Move task between columns
  - `update <ref>` - Modify task properties
  - `delete <ref>` - Remove task (with --force to skip confirmation)

- **Task Reference Formats**
  - Full ID: `abc123def456`
  - ID prefix: `abc` (minimum unique prefix)
  - Display ID: `WRK-123` (board prefix + sequence)
  - Title substring: `"dark mode"` (case-insensitive)

- **Columns (Workflow States)**
  - `backlog` - Ideas and future work
  - `todo` - Ready to start
  - `in_progress` - Currently being worked on
  - `review` - Awaiting review
  - `done` - Completed

- **JSON Output**
  - All commands support `--json` flag
  - Use `--fields id,title,column` to limit output fields
  - JSON output is token-efficient for agent processing

- **Exit Codes**
  - 0: Success
  - 1: General error
  - 2: Invalid arguments
  - 3: Task not found
  - 4: Ambiguous reference
  - 5: Validation error

### 5.2 `egenskriven-workflows` (Workflow Skill)

**Frontmatter**:
```yaml
---
name: egenskriven-workflows
description: Workflow modes and agent behaviors for EgenSkriven task management. 
  Use when configuring strict/light/minimal workflow modes, understanding 
  autonomous/collaborative/supervised agent behaviors, or using the prime command.
---
```

**Content outline**:

- **Workflow Modes Overview**
  - Configured in `.egenskriven/config.json`
  - Controls how strictly agents should use task tracking
  - Three modes: strict, light, minimal

- **Strict Workflow Mode**
  - Full enforcement of task tracking
  - Before starting work: Create or claim a task
  - During work: Update task status, add notes
  - After completion: Mark done, create follow-up tasks
  - Best for: Complex projects, team coordination, audit trails

- **Light Workflow Mode**
  - Basic tracking without ceremony
  - Create tasks for significant work
  - Complete tasks when done
  - No structured sections required
  - Best for: Solo development, rapid iteration

- **Minimal Workflow Mode**
  - Agent decides when to use task tracking
  - No enforcement or requirements
  - EgenSkriven available but optional
  - Best for: Exploratory work, quick fixes

- **Agent Modes**
  - `autonomous` - Execute actions directly, human reviews async
  - `collaborative` - Propose major changes, execute minor ones
  - `supervised` - Read-only, output commands for human to run

- **Autonomous Mode Details**
  - Agent creates, updates, completes tasks without asking
  - Human reviews via activity history
  - Ideal for trusted agents with clear scope

- **Collaborative Mode Details**
  - Agent can read tasks and make minor updates
  - Major changes (complete, delete) require explanation
  - Agent states intent, waits for human confirmation
  - Balance of autonomy and oversight

- **Supervised Mode Details**
  - Agent can only read task data
  - Outputs CLI commands for human to execute
  - Maximum control, minimum agent autonomy
  - Good for sensitive projects or new agents

- **Prime Command**
  - `egenskriven prime` - Output full agent instructions
  - `egenskriven prime --workflow strict` - Override workflow mode
  - `egenskriven prime --agent claude` - Identify agent
  - Used in session hooks (SessionStart, PreCompact)

- **Session Patterns**
  - Session start: Check `context`, review `suggest`
  - During work: Update task status as progress is made
  - Before completion: Verify all work is tracked
  - Session end: Summary of completed and remaining work

- **When to Create Tasks**
  - Multi-step work spanning multiple exchanges
  - Work that might be interrupted
  - Features with dependencies
  - Bugs that need tracking
  - NOT: Simple questions, one-off commands

### 5.3 `egenskriven-advanced` (Advanced Skill)

**Frontmatter**:
```yaml
---
name: egenskriven-advanced
description: Advanced EgenSkriven features including epics, task dependencies, 
  sub-tasks, batch operations, and views. Use when working with blocked tasks, 
  epic management, sub-tasks, batch imports, or complex filtering.
---
```

**Content outline**:

- **Epics**
  - Group related tasks under a theme
  - `egenskriven epic add "Epic title" --color "#hex"`
  - `egenskriven epic list`
  - `egenskriven epic show <ref>`
  - Link tasks: `egenskriven add "Task" --epic <epic-ref>`
  - Filter by epic: `egenskriven list --epic <ref>`

- **Task Dependencies (Blocking)**
  - Tasks can block other tasks
  - Add blocker: `egenskriven update <task> --blocked-by <blocker>`
  - Remove blocker: `egenskriven update <task> --remove-blocked-by <blocker>`
  - Circular dependencies are prevented automatically
  - Self-blocking is prevented

- **Filtering by Block Status**
  - `--ready` - Unblocked tasks in todo/backlog (agent-friendly)
  - `--is-blocked` - Only tasks blocked by others
  - `--not-blocked` - Only tasks with no blockers
  - Combine with other filters for precise queries

- **Sub-tasks**
  - Create sub-task: `egenskriven add "Sub-task" --parent <parent-ref>`
  - Parent shows progress based on sub-task completion
  - Sub-tasks inherit board from parent
  - Filter: `--has-parent`, `--no-parent`

- **Batch Operations**
  - Add multiple tasks from JSON lines:
    ```bash
    echo '{"title":"Task 1"}\n{"title":"Task 2"}' | egenskriven add --stdin
    ```
  - Add from file: `egenskriven add --file tasks.json`
  - Delete multiple: `egenskriven delete id1 id2 id3`
  - Delete from stdin: `echo "id1\nid2" | egenskriven delete --stdin`

- **Due Dates**
  - Set due date: `egenskriven add "Task" --due 2024-03-15`
  - Natural language: `--due tomorrow`, `--due "next friday"`
  - Filter: `--due-before`, `--due-after`, `--has-due`, `--no-due`
  - Overdue tasks highlighted in UI

- **Views and Filters**
  - Filter by type: `--type bug,feature`
  - Filter by priority: `--priority high,urgent`
  - Filter by label: `--label frontend`
  - Search: `--search "login"`
  - Limit results: `--limit 10`
  - Sort: `--sort priority`, `--sort -created` (descending)

- **Boards**
  - Multiple boards for different contexts
  - `egenskriven board list`
  - `egenskriven board add "Board Name" --prefix WRK`
  - `egenskriven board use <name>` - Set default
  - `--board <name>` flag on task commands
  - `--all-boards` to see all tasks

- **Import/Export**
  - Export JSON: `egenskriven export --format json > backup.json`
  - Export CSV: `egenskriven export --format csv > tasks.csv`
  - Import: `egenskriven import backup.json`
  - Import strategies: `--strategy merge` or `--strategy replace`

- **Field Selection for JSON Output**
  - `--fields id,title,column` - Only include specified fields
  - Reduces token usage when full details not needed
  - Combine with `--json` for machine-readable output

---

## 6. AGENTS.md Integration

### The Universal Standard

`AGENTS.md` is an emerging standard adopted by 60k+ open-source projects and officially maintained by the Agentic AI Foundation under the Linux Foundation.

### Supported Agents

| Agent | AGENTS.md Support |
|-------|------------------|
| Claude Code | Native (also reads CLAUDE.md) |
| OpenCode | Native |
| Cursor | Native |
| OpenAI Codex | Native |
| Google Jules | Native |
| GitHub Copilot | Native |
| VS Code | Native |
| Windsurf | Native |
| Aider | Native |
| Zed | Native |
| Warp | Native |
| Factory | Native |
| Devin | Native |

### Role in EgenSkriven Integration

AGENTS.md serves as a **lightweight pointer** to skills, not a comprehensive guide.

**Purpose**:
- Provide minimal always-loaded context (~200-500 tokens)
- Point agents to skills for detailed information
- Universal compatibility across all agents
- Quick reference for essential commands

**What NOT to include**:
- Full CLI documentation (put in skills)
- Detailed workflows (put in egenskriven-workflows skill)
- Advanced features (put in egenskriven-advanced skill)

### AGENTS.md Template for EgenSkriven Projects

The following template should be placed at the project root:

```markdown
# AGENTS.md

## Task Management

This project uses **EgenSkriven** for task tracking - a local-first kanban designed for AI agents.

### Quick Commands

```bash
# Get project status
egenskriven context --json

# Get suggested next task
egenskriven suggest --json

# List ready (unblocked) tasks
egenskriven list --ready --json

# Create a task
egenskriven add "Task title" --type feature --priority medium

# Complete a task
egenskriven move <task-ref> done
```

### Task References

Tasks can be referenced by:
- Full ID: `abc123def456`
- ID prefix: `abc`
- Display ID: `WRK-123`
- Title: `"dark mode"`

### For More Information

Load the appropriate skill for detailed guidance:
- `egenskriven` - Core commands and task management
- `egenskriven-workflows` - Workflow modes (strict/light/minimal) and agent behaviors
- `egenskriven-advanced` - Epics, dependencies, sub-tasks, batch operations

### Workflow

This project uses **light** workflow mode:
- Create tasks for significant work
- Update status as you progress
- Complete tasks when done
- Use `egenskriven suggest` to find next work
```

### Customization Points

Projects should customize the template:

1. **Workflow mode**: Change "light" to "strict" or "minimal" as appropriate
2. **Quick commands**: Add project-specific common operations
3. **Board info**: If using multiple boards, mention the default
4. **Project conventions**: Add any project-specific task conventions

---

## 7. The `skill install` Command

### Purpose

Provide a simple way for users to install EgenSkriven skills to their system, making the skills available to AI agents.

### User Flow

```
$ egenskriven skill install

EgenSkriven Skill Installation

Where would you like to install the skills?

  1. Global (~/.claude/skills/) - Available in all projects
  2. Project (.claude/skills/) - Only this project

Enter choice [1/2]: 1

Installing skills to ~/.claude/skills/...

Created:
  ~/.claude/skills/egenskriven/SKILL.md
  ~/.claude/skills/egenskriven-workflows/SKILL.md
  ~/.claude/skills/egenskriven-advanced/SKILL.md

Skills installed successfully!

Next steps:
  1. Restart your AI agent (Claude Code, OpenCode, etc.)
  2. The agent will automatically discover the new skills
  3. Skills load on-demand when relevant to your task

To uninstall: egenskriven skill uninstall
To update: egenskriven skill install --force
```

### Directory Structure Created

**Global installation** (`~/.claude/skills/`):
```
~/.claude/skills/
├── egenskriven/
│   └── SKILL.md
├── egenskriven-workflows/
│   └── SKILL.md
└── egenskriven-advanced/
    └── SKILL.md
```

**Project installation** (`.claude/skills/`):
```
.claude/skills/
├── egenskriven/
│   └── SKILL.md
├── egenskriven-workflows/
│   └── SKILL.md
└── egenskriven-advanced/
    └── SKILL.md
```

### Command Flags

| Flag | Description |
|------|-------------|
| `--global` | Install to ~/.claude/skills/ without prompting |
| `--project` | Install to .claude/skills/ without prompting |
| `--force` | Overwrite existing skills |
| `--json` | Output result as JSON |

### Uninstall and Update

```bash
# Remove installed skills
egenskriven skill uninstall

# Update to latest version (overwrites existing)
egenskriven skill install --force

# Check installed version
egenskriven skill status
```

### Implementation Notes

- Skills content should be embedded in the Go binary (go:embed)
- Use same content as would be in .claude/skills/ in the repo
- Respect XDG_CONFIG_HOME for non-standard home directories
- Create parent directories if they don't exist
- Warn but don't fail if skills already exist (unless --force)

---

## 8. Context Window Optimization

### Context Window Sizes (2025)

| Model | Context Window |
|-------|---------------|
| Claude Opus/Sonnet/Haiku | 200,000 tokens |
| GPT-4o | 128,000 tokens |
| GPT-4.1 | ~1M tokens |
| Gemini 2.0 | 1M+ tokens |

### The 80% Rule

Research shows model performance degrades before the context window is full due to "context distraction". Best practice:

> **Trigger optimization when approaching 80% context utilization**

For a 200k context window, this means optimizing around 160k tokens.

### Budget Allocation

Recommended allocation for a 200k context window:

| Category | % | Tokens |
|----------|---|--------|
| System instructions | 1-5% | 2,000-10,000 |
| Tool definitions | 2-5% | 4,000-10,000 |
| Retrieved knowledge | 10-30% | 20,000-60,000 |
| Conversation history | 30-50% | 60,000-100,000 |
| Reserved buffer | 15-20% | 30,000-40,000 |

**EgenSkriven skills target**: 1-2.5% (2,000-5,000 tokens)

### Tiered Loading Strategy

| Tier | Token Range | Use Case |
|------|-------------|----------|
| **Compact** | 500-1,000 | Session hooks, quick reminders |
| **Medium** | 2,000-5,000 | Full skill content, most use cases |
| **Comprehensive** | 5,000-10,000 | Complex domains, rich examples |

EgenSkriven uses the **medium tier** as default.

### Why Smaller is Better

1. **Context distraction**: More content = more places for model to lose focus
2. **Tool selection accuracy**: Research shows >30 tools causes errors
3. **Processing time**: Larger context = slower responses
4. **Cost**: Many APIs charge per token

### Lazy Loading Patterns

1. **Skills on demand**: Agent loads skill only when task matches
2. **Field filtering**: Use `--fields` to reduce JSON output
3. **Pagination**: Use `--limit` to cap results
4. **Progressive detail**: Start with list, drill into show

### Beads Reference

The Beads project achieves excellent token efficiency:

| Mode | Tokens | Use Case |
|------|--------|----------|
| MCP mode | ~50 | Brief workflow reminders |
| CLI mode | ~1-2k | Full command reference |
| Full tool scan | ~10.5k | Avoided via lazy loading |

EgenSkriven targets similar efficiency with skills.

---

## 9. Prime Command Relationship

### Prime vs Skills: When to Use Each

| Scenario | Use Prime | Use Skills |
|----------|-----------|------------|
| Session hooks (SessionStart) | Yes | No |
| Pre-compaction injection | Yes | No |
| Agent actively working with tasks | No | Yes |
| Environments without skill support | Yes | No |
| Maximum token efficiency | No | Yes |
| Guaranteed context injection | Yes | No |

### Prime Command Purpose

The `prime` command exists for scenarios where skills aren't available or hooks are needed:

```bash
# Basic usage - outputs full instructions
egenskriven prime

# Override workflow mode
egenskriven prime --workflow strict

# Identify the agent
egenskriven prime --agent claude
```

### Hook Integration

For agents that support hooks (Claude Code), prime can be auto-injected:

**Claude Code** (`.claude/settings.json`):
```json
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          { "type": "command", "command": "egenskriven prime" }
        ]
      }
    ],
    "PreCompact": [
      {
        "hooks": [
          { "type": "command", "command": "egenskriven prime" }
        ]
      }
    ]
  }
}
```

### Keeping Both Approaches

Skills and prime serve complementary purposes:

| Feature | Prime | Skills |
|---------|-------|--------|
| Token efficiency | Medium (~1-2k always) | High (on-demand) |
| Guaranteed injection | Yes (via hooks) | No (agent decides) |
| Cross-platform | Requires hook support | Universal |
| Customization | Via config file | Via skill content |
| Updates | Automatic (from binary) | Requires reinstall |

**Recommendation**: Use skills as primary, keep prime for backward compatibility and hook-based workflows.

### Backward Compatibility

Existing users with `egenskriven prime` in their hooks will continue to work. The skill system is additive, not a replacement.

Migration path:
1. Install skills: `egenskriven skill install`
2. Skills provide on-demand detail
3. Prime provides hook-based injection
4. Both can coexist

---

## 10. Real-World Patterns (Research Findings)

### Metabase Pattern (6 Skills + Shared)

Metabase splits skills by language and activity:

```
.claude/skills/
├── _shared/
│   ├── development-workflow.md
│   ├── clojure-style-guide.md
│   └── typescript-commands.md
├── clojure-write/SKILL.md
├── clojure-review/SKILL.md
├── typescript-write/SKILL.md
├── typescript-review/SKILL.md
├── docs-write/SKILL.md
└── docs-review/SKILL.md
```

**Key insight**: Separate write and review skills per domain, with shared content in `_shared/`.

**Naming pattern**: `[language]-[action]`

### PyTorch Pattern (Task-Specific)

PyTorch uses very focused, task-specific skills:

```
.claude/skills/
├── at-dispatch-v2/SKILL.md    # Convert dispatch macros
├── add-uint-support/SKILL.md  # Add unsigned int support
├── docstring/SKILL.md         # Write docstrings
└── skill-writer/SKILL.md      # Meta: create new skills
```

**Key insight**: Each skill handles ONE specific technical task.

**Naming pattern**: `[task-name]`

### Beads Pattern (Prime + AGENTS.md)

Beads uses a hybrid approach:

1. **`bd prime`**: Injected via hooks (~1-2k tokens)
2. **AGENTS.md**: Universal instructions
3. **CLAUDE.md**: Claude-specific details
4. **MCP server**: Fallback for non-CLI environments

**Key insight**: CLI + hooks uses 99% fewer tokens than full MCP tool scan.

**Token comparison**:
- MCP full scan: ~10,500 tokens
- CLI prime: ~1-2k tokens
- MCP minimal: ~50 tokens

### Ghost CMS Pattern (Workflow-Based)

Ghost organizes by development workflow:

```
.claude/skills/
├── add-admin-api-endpoint/SKILL.md
└── create-database-migration/SKILL.md
```

**Key insight**: Skills map to specific development tasks users commonly need.

**Naming pattern**: `[action]-[domain]`

### Common Patterns Summary

| Pattern | When to Use |
|---------|-------------|
| Language-Activity (Metabase) | Multi-language projects |
| Task-Specific (PyTorch) | Narrow, repeatable tasks |
| Workflow-Based (Ghost) | Common development patterns |
| Hybrid (Beads) | Maximum compatibility |

EgenSkriven uses a **domain-split** pattern: core, workflows, advanced.

---

## 11. Implementation Checklist

### Phase 1: Create Skill Files

- [ ] Create `.claude/skills/egenskriven/SKILL.md`
  - Frontmatter with name and description
  - Core commands and task management
  - ~150-200 lines, ~800-1k tokens

- [ ] Create `.claude/skills/egenskriven-workflows/SKILL.md`
  - Workflow modes documentation
  - Agent modes documentation
  - ~250-350 lines, ~1.5-2k tokens

- [ ] Create `.claude/skills/egenskriven-advanced/SKILL.md`
  - Epics, dependencies, sub-tasks
  - Batch operations, views, filters
  - ~250-350 lines, ~1.5-2k tokens

### Phase 2: Implement `skill install` Command

- [ ] Add `skill` command group to CLI
- [ ] Implement `skill install` with prompt
- [ ] Implement `--global` and `--project` flags
- [ ] Implement `--force` for overwrite
- [ ] Implement `skill uninstall`
- [ ] Implement `skill status`
- [ ] Embed skill content in binary (go:embed)

### Phase 3: Create AGENTS.md Template

- [ ] Create template in documentation
- [ ] Add to project root as example
- [ ] Document customization points

### Phase 4: Update Documentation

- [ ] Update README with skill installation
- [ ] Add skills section to README
- [ ] Document AGENTS.md usage
- [ ] Add troubleshooting guide

### Phase 5: Testing

- [ ] Test skill install (global)
- [ ] Test skill install (project)
- [ ] Test with Claude Code
- [ ] Test with OpenCode
- [ ] Test skill discovery
- [ ] Test skill loading
- [ ] Verify token counts

### Phase 6: Integration

- [ ] Ensure prime command still works
- [ ] Document migration from prime-only
- [ ] Update hook examples in docs

---

## 12. Future Considerations

### MCP Server as Fallback

For environments without shell access (Claude Desktop, some IDEs):

- Implement MCP server exposing EgenSkriven operations
- Higher token cost but universal compatibility
- Lower priority than skills approach

### Skill Versioning

As EgenSkriven evolves:

- Add version field to SKILL.md frontmatter
- `skill install` checks for updates
- `skill status` shows installed vs latest version
- Consider semantic versioning for skills

### Community Skill Contributions

Enable community extensions:

- Document skill authoring guidelines
- Create skill template generator
- Consider skill registry/discovery
- Allow custom skills alongside official ones

### Automated Skill Updates

Future enhancements:

- `egenskriven skill update` - Check and update skills
- Notification when skills are outdated
- Automatic updates (opt-in)
- Changelog for skill updates

### Cross-Agent Testing

Expand testing coverage:

- Cursor integration testing
- Codex integration testing
- Document agent-specific quirks
- Maintain compatibility matrix

---

## References

### Research Sources

- OpenCode Skills Documentation
- Claude Code Skills Specification (agentskills.io)
- Anthropic Best Practices for Context Engineering
- Microsoft AI Agents Course
- Beads Project (github.com/steveyegge/beads)
- Metabase Skills Implementation
- PyTorch Skills Implementation

### Related Documentation

- `docs/plan.md` - Phase 1.5 Agent Integration
- `docs/kanban-architecture.md` - Overall architecture
- `internal/commands/prime.go` - Prime command implementation
- `internal/commands/prime.tmpl` - Prime template content

---

*This document captures research conducted in January 2026 on AI agent integration patterns. Update as the ecosystem evolves.*
