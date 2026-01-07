---
name: egenskriven-workflows
description: Workflow modes and agent behaviors for EgenSkriven task management. Use when configuring strict/light/minimal workflow modes, understanding autonomous/collaborative/supervised agent behaviors, or using the prime command.
---

## Overview

EgenSkriven supports configurable workflow modes that control how strictly agents should use task tracking, and agent modes that define the level of autonomy.

## Workflow Modes

Configured in `.egenskriven/config.json` under `workflow_mode`.

### Strict Mode

Full enforcement of task tracking. Best for complex projects, team coordination, and audit trails.

**Requirements:**
- Before starting work: Create or claim a task
- During work: Update task status, add notes
- After completion: Mark done, create follow-up tasks

```bash
# Check current context before starting
egenskriven context --json

# Claim a task before working
egenskriven move <task-ref> in_progress

# Complete when done
egenskriven move <task-ref> done
```

### Light Mode (Default)

Basic tracking without ceremony. Best for solo development and rapid iteration.

**Requirements:**
- Create tasks for significant work
- Complete tasks when done
- No structured sections required

```bash
# Create task for significant work
egenskriven add "Implement feature X" --type feature

# Complete when done
egenskriven move <task-ref> done
```

### Minimal Mode

Agent decides when to use task tracking. Best for exploratory work and quick fixes.

**Behavior:**
- No enforcement or requirements
- EgenSkriven available but optional
- Agent uses judgment on when tracking adds value

## Agent Modes

Controls the level of agent autonomy in task operations.

### Autonomous Mode

Agent executes actions directly; human reviews asynchronously.

**Behavior:**
- Creates, updates, completes tasks without asking
- Human reviews via activity history
- Ideal for trusted agents with clear scope

**Example session:**
```bash
# Agent automatically creates task
egenskriven add "Fix login bug" --type bug --priority high

# Agent works and updates status
egenskriven move FIX-1 in_progress
egenskriven update FIX-1 --note "Root cause: session timeout"

# Agent completes when done
egenskriven move FIX-1 done
```

### Collaborative Mode

Agent proposes major changes, executes minor ones.

**Behavior:**
- Can read tasks and make minor updates
- Major changes (complete, delete) require explanation
- Agent states intent, human confirms if needed
- Balance of autonomy and oversight

**Major actions requiring explanation:**
- Completing tasks
- Deleting tasks
- Changing priority to urgent
- Reassigning tasks

### Supervised Mode

Agent can only read task data; outputs commands for human to execute.

**Behavior:**
- Read-only access to task data
- Outputs CLI commands as suggestions
- Maximum control, minimum agent autonomy
- Good for sensitive projects or new agents

**Example output:**
```
I recommend completing this task. Run:
  egenskriven move FIX-1 done
```

## Prime Command

The `prime` command outputs full agent instructions for hook-based injection.

```bash
# Basic usage - outputs full instructions
egenskriven prime

# Override workflow mode
egenskriven prime --workflow strict

# Identify the agent
egenskriven prime --agent claude
```

### Hook Integration

For agents supporting hooks (Claude Code), prime can be auto-injected:

```json
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          { "type": "command", "command": "egenskriven prime" }
        ]
      }
    ]
  }
}
```

## Session Patterns

### Starting a Session

1. Run `egenskriven context --json` to get project state
2. Run `egenskriven suggest --json` to get recommended next task
3. Review ready tasks with `egenskriven list --ready`

### During Work

1. Move task to `in_progress` when starting
2. Add notes for significant progress or blockers
3. Update status as work progresses

### Completing Work

1. Verify all acceptance criteria are met
2. Move task to `done` or `review` as appropriate
3. Create follow-up tasks if needed

## When to Create Tasks

**Create tasks for:**
- Multi-step work spanning multiple exchanges
- Work that might be interrupted
- Features with dependencies
- Bugs that need tracking

**Skip task creation for:**
- Simple questions or explanations
- One-off commands
- Trivial changes (typos, formatting)
- Information gathering

## Configuration

Check current configuration:

```bash
egenskriven config show
```

View workflow mode in context:

```bash
egenskriven context --json | jq '.workflow_mode'
```

## Related Skills

- `egenskriven` - Core commands and task management
- `egenskriven-advanced` - Epics, dependencies, sub-tasks, batch operations
