# EgenSkriven Collaborative Workflow Guide

This guide explains how to use EgenSkriven's collaborative workflow feature
for human-AI pair programming.

## Introduction

The collaborative workflow enables AI coding assistants (OpenCode, Claude Code,
Codex) to request human input when they encounter decisions or blockers. This
creates a seamless back-and-forth communication loop where:

- AI agents can signal when they need human guidance
- Humans can respond via comments on tasks
- Agents can be resumed with full context preserved

This workflow transforms EgenSkriven into a **control plane** for AI agent
development, enabling structured collaboration regardless of which AI coding
tool you use.

## Getting Started

### Prerequisites

- EgenSkriven installed and configured
- One of: OpenCode, Claude Code, or Codex CLI
- A project with EgenSkriven initialized (`egenskriven init`)

### Setup

1. **Initialize tool integration**:

   ```bash
   cd your-project
   egenskriven init --all  # Or --opencode, --claude-code, --codex
   ```

   This creates the necessary files for session ID discovery:
   - OpenCode: `.opencode/tool/egenskriven-session.ts`
   - Claude Code: `.claude/hooks/egenskriven-session.sh` and settings
   - Codex: `.codex/get-session-id.sh`

2. **Configure resume mode** (optional):

   ```bash
   # View current board settings
   egenskriven board show main

   # Set resume mode (manual, command, or auto)
   egenskriven board update main --resume-mode command
   ```

## Workflow

### Step 1: Agent Starts Work

The agent picks up a task and links its session:

```
Agent: I'll work on WRK-42. Let me link my session first.
```

**For OpenCode:**
```bash
# Agent calls the egenskriven-session tool to get session ID
# Then runs:
egenskriven session link WRK-42 --tool opencode --ref <session-id>
egenskriven move WRK-42 in_progress
```

**For Claude Code:**
```bash
# Session ID is automatically set by the hook
egenskriven session link WRK-42 --tool claude-code --ref $CLAUDE_SESSION_ID
egenskriven move WRK-42 in_progress
```

**For Codex:**
```bash
SESSION_ID=$(.codex/get-session-id.sh)
egenskriven session link WRK-42 --tool codex --ref $SESSION_ID
egenskriven move WRK-42 in_progress
```

### Step 2: Agent Gets Blocked

The agent encounters a decision point that requires human input:

```
Agent: I need guidance on the authentication approach.
       I'll block the task with my question.
```

```bash
egenskriven block WRK-42 "Should I use JWT or session-based authentication?
JWT is stateless but requires token refresh logic.
Sessions are simpler but require server-side storage."
```

This command:
1. Moves the task to the `need_input` column
2. Adds the question as a comment
3. Preserves the agent session for later resume

The agent session ends here, but the context is preserved.

### Step 3: Human Responds

You see the blocked task in the UI or CLI:

```bash
$ egenskriven list --need-input
ID       TITLE              COLUMN       PRIORITY
WRK-42   Implement auth     need_input   high

$ egenskriven comments WRK-42
[agent @ 10:30]: Should I use JWT or session-based authentication?
                 JWT is stateless but requires token refresh logic.
                 Sessions are simpler but require server-side storage.
```

Add your response:

```bash
$ egenskriven comment WRK-42 "Use JWT with refresh tokens. Set access token
expiry to 15 minutes and refresh token to 7 days. Store refresh tokens
in HttpOnly cookies for security."
```

### Step 4: Resume the Agent

Depending on your configured resume mode:

**Command mode** (default):
```bash
$ egenskriven resume WRK-42 --exec
Resuming session for WRK-42...
Tool: opencode
Session: abc-123
Working directory: /home/user/project

[Agent session starts with full context]
```

**Manual mode**:
```bash
$ egenskriven resume WRK-42
Resume command for WRK-42:

  opencode run '## Task Context...' --session abc-123

Working directory: /home/user/project
Copy and run the command above.
```

**Auto mode**:
```bash
# Include @agent in your comment to trigger auto-resume
$ egenskriven comment WRK-42 "@agent Use JWT with refresh tokens..."
# Agent automatically resumes with full context
```

### Step 5: Agent Continues

The agent receives full context when resumed:

```markdown
## Task Context (from EgenSkriven)

**Task**: WRK-42 - Implement authentication
**Status**: need_input -> in_progress
**Priority**: high

## Conversation Thread

[agent @ 10:30]: Should I use JWT or session-based authentication?
                 JWT is stateless but requires token refresh logic.
                 Sessions are simpler but require server-side storage.

[human @ 10:45]: Use JWT with refresh tokens. Set access token expiry to
                 15 minutes and refresh token to 7 days. Store refresh
                 tokens in HttpOnly cookies for security.

## Instructions

Continue working on the task based on the human's response above.
The conversation context should help you understand what was discussed.
```

The agent now has everything it needs to continue working.

## Configuration

### Resume Modes

| Mode | Trigger | Best For |
|------|---------|----------|
| `manual` | Copy printed command | Learning, debugging |
| `command` | `egenskriven resume --exec` | Most workflows |
| `auto` | Comment with `@agent` | Quick back-and-forth |

### Setting Resume Mode

```bash
# Per-board configuration
egenskriven board update <board> --resume-mode <mode>

# Example: Enable auto-resume
egenskriven board update main --resume-mode auto
```

### Tool Integration Files

| Tool | Files Created |
|------|---------------|
| OpenCode | `.opencode/tool/egenskriven-session.ts` |
| Claude Code | `.claude/hooks/egenskriven-session.sh`, `.claude/settings.json` |
| Codex | `.codex/get-session-id.sh` |

Regenerate with `--force`:
```bash
egenskriven init --opencode --force
```

## Tips

1. **Be specific in questions**: Include context about what you've considered
   and what trade-offs you're weighing. This helps humans provide better answers.

2. **List options**: If you have multiple approaches in mind, list them with
   pros/cons. This speeds up the human's decision process.

3. **Link early**: Link your session at the start of work, before you might
   need to block. This ensures the resume flow works correctly.

4. **Check for updates**: If working on multiple tasks, periodically check
   for responses using `egenskriven comments <task> --json`.

5. **Use JSON output**: For automation and scripting, use `--json` flags:
   ```bash
   egenskriven list --need-input --json
   egenskriven comments WRK-42 --json
   egenskriven resume WRK-42 --json
   ```

## Troubleshooting

### "No session linked to task"

**Problem**: You tried to resume a task that doesn't have a linked session.

**Solution**: The agent must link its session before blocking:
```bash
egenskriven session link <task> --tool <tool> --ref <session-id>
```

### "Task is not in need_input state"

**Problem**: You tried to resume a task that isn't blocked.

**Solution**: Check the current task state:
```bash
egenskriven show <task>
```

Only tasks in the `need_input` column can be resumed.

### "Session ID not found" (Claude Code)

**Problem**: The `$CLAUDE_SESSION_ID` environment variable is not set.

**Solution**: Ensure the hook is properly configured:
```bash
# Regenerate the hook
egenskriven init --claude-code --force

# Verify settings.json has the hook configured
cat .claude/settings.json
```

### "Cannot find egenskriven-session tool" (OpenCode)

**Problem**: The custom tool file is missing or not recognized.

**Solution**: Regenerate the tool:
```bash
egenskriven init --opencode --force

# Verify the file exists
ls -la .opencode/tool/egenskriven-session.ts
```

### "Permission denied" (Codex helper script)

**Problem**: The helper script is not executable.

**Solution**: Make it executable:
```bash
chmod +x .codex/get-session-id.sh
```

### Comments not appearing

**Problem**: Comments you added aren't showing up.

**Solution**: Ensure you're using the correct task reference:
```bash
# Use display ID
egenskriven comments WRK-42

# Or use full ID
egenskriven comments abc123def456

# Check task exists
egenskriven show WRK-42
```
