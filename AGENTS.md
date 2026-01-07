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
