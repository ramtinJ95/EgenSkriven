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

## Human-AI Collaborative Workflow

This project supports a collaborative workflow for AI agents using EgenSkriven
as a control plane.

### Quick Start

1. **Initialize tool integration**:
   ```bash
   egenskriven init --opencode    # For OpenCode
   egenskriven init --claude-code # For Claude Code
   egenskriven init --codex       # For Codex CLI
   ```

2. **Link your session when starting work**:
   ```bash
   egenskriven session link <task> --tool <tool> --ref <session-id>
   ```

3. **Block when you need human input**:
   ```bash
   egenskriven block <task> "Your question"
   ```

4. **Human responds**, then resumes you with full context.

### Session ID Discovery

| Tool | Method |
|------|--------|
| OpenCode | Call `egenskriven-session` tool |
| Claude Code | Use `$CLAUDE_SESSION_ID` env var |
| Codex | Run `.codex/get-session-id.sh` |

### Commands Reference

| Command | Description |
|---------|-------------|
| `egenskriven block <task> "msg"` | Block task with question |
| `egenskriven comment <task> "msg"` | Add comment to task |
| `egenskriven comments <task>` | List comments |
| `egenskriven session link <task>` | Link session to task |
| `egenskriven session show <task>` | Show linked session |
| `egenskriven resume <task>` | Resume blocked task |
| `egenskriven list --need-input` | List blocked tasks |

### Resume Modes

This project uses **command** resume mode:
- Run `egenskriven resume <task> --exec` to resume agent sessions
- For manual mode: command is printed for you to copy
- For auto mode: add comment with `@agent` to trigger auto-resume
