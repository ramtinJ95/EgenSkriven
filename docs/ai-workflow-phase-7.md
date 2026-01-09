# Phase 7: Documentation & Polish

> **Parent Document**: [ai-workflow-plan.md](./ai-workflow-plan.md)
> **Phase**: 7 of 7
> **Status**: Not Started
> **Prerequisites**: [Phase 6](./ai-workflow-phase-6.md) completed

---

## Phase 7 Todo List

### Documentation Updates

- [x] **Task 7.1: Update `egenskriven` Skill**
  - [x] 7.1.1: Add "Human-AI Collaborative Workflow" section header
  - [x] 7.1.2: Document workflow overview diagram
  - [x] 7.1.3: Document tool integration setup for OpenCode
  - [x] 7.1.4: Document tool integration setup for Claude Code
  - [x] 7.1.5: Document tool integration setup for Codex
  - [x] 7.1.6: Document session linking commands for each tool
  - [x] 7.1.7: Document `egenskriven block` command with examples
  - [x] 7.1.8: Document `egenskriven comment` command with examples
  - [x] 7.1.9: Document `egenskriven comments` command with examples
  - [x] 7.1.10: Document `egenskriven list --need-input` command
  - [x] 7.1.11: Document resume modes table (manual, command, auto)
  - [x] 7.1.12: Document `egenskriven resume` command with examples
  - [x] 7.1.13: Add complete workflow example section
  - [x] 7.1.14: Add tips section for best practices

- [x] **Task 7.2: Update `egenskriven-workflows` Skill**
  - [x] 7.2.1: Add "Resume Modes" section header
  - [x] 7.2.2: Document Manual Mode with example output
  - [x] 7.2.3: Document Command Mode with example output
  - [x] 7.2.4: Document Auto Mode with example output
  - [x] 7.2.5: Document `egenskriven board update --resume-mode` command
  - [x] 7.2.6: Add workflow recommendations table
  - [x] 7.2.7: Document collaborative mode integration with agent modes
  - [x] 7.2.8: Add example JSON configuration

- [x] **Task 7.3: Update Prime Template**
  - [x] 7.3.1: Add "Human-AI Collaborative Workflow" section
  - [x] 7.3.2: Add "Before Starting Work" subsection with session linking
  - [x] 7.3.3: Add tool-specific session linking using `.Tool` template variable
  - [x] 7.3.4: Document "When You Need Human Input" workflow
  - [x] 7.3.5: Document context format agent receives after resume
  - [x] 7.3.6: Add "Best Practices" subsection
  - [x] 7.3.7: Add complete workflow example with bash commands
  - [x] 7.3.8: Add "Checking for Responses" subsection

- [x] **Task 7.4: Update AGENTS.md**
  - [x] 7.4.1: Add "Human-AI Collaborative Workflow" section header
  - [x] 7.4.2: Add "Quick Start" subsection with numbered steps
  - [x] 7.4.3: Document tool initialization commands
  - [x] 7.4.4: Document session linking command pattern
  - [x] 7.4.5: Document block command pattern
  - [x] 7.4.6: Add "Session ID Discovery" table
  - [x] 7.4.7: Add "Commands Reference" table
  - [x] 7.4.8: Add "Resume Modes" section with mode-specific behavior

### New Documentation Files

- [x] **Task 7.5: Create User Guide**
  - [x] 7.5.1: Create `docs/collaborative-workflow-guide.md` file
  - [x] 7.5.2: Add "Introduction" section
  - [x] 7.5.3: Add "Getting Started" section with prerequisites
  - [x] 7.5.4: Document setup steps (init, configure resume mode)
  - [x] 7.5.5: Add "Workflow" section - Step 1: Agent Starts Work
  - [x] 7.5.6: Add "Workflow" section - Step 2: Agent Gets Blocked
  - [x] 7.5.7: Add "Workflow" section - Step 3: Human Responds
  - [x] 7.5.8: Add "Workflow" section - Step 4: Resume the Agent
  - [x] 7.5.9: Add "Workflow" section - Step 5: Agent Continues
  - [x] 7.5.10: Add "Configuration" section with resume modes table
  - [x] 7.5.11: Add "Tips" section with best practices
  - [x] 7.5.12: Add "Troubleshooting" section with common errors

### Testing & Verification

- [x] **Task 7.6: Create Final Test Script**
  - [x] 7.6.1: Create `scripts/test-collaborative-workflow.sh` file
  - [x] 7.6.2: Add color output functions (pass/fail)
  - [x] 7.6.3: Add setup section (temp dir, init egenskriven)
  - [x] 7.6.4: Implement Test 1: Basic block and comment
  - [x] 7.6.5: Implement Test 2: Session linking
  - [x] 7.6.6: Implement Test 3: Resume command generation
  - [x] 7.6.7: Implement Test 4: Tool integrations (OpenCode, Claude Code, Codex)
  - [x] 7.6.8: Implement Test 5: List --need-input filter
  - [x] 7.6.9: Add manual testing recommendations output
  - [x] 7.6.10: Make script executable

- [x] **Task 7.7: Performance Review**
  - [x] 7.7.1: Create `docs/performance-notes.md` file
  - [x] 7.7.2: Add "Expected Performance" table with targets
  - [x] 7.7.3: Document database indexes
  - [x] 7.7.4: Document optimizations

- [x] **Task 7.8: Create Release Checklist**
  - [x] 7.8.1: Create `docs/release-checklist.md` file
  - [x] 7.8.2: Add "Documentation" section checklist
  - [x] 7.8.3: Add "Unit Tests" section checklist
  - [x] 7.8.4: Add "Integration Tests" section checklist
  - [x] 7.8.5: Add "E2E Tests" section checklist
  - [x] 7.8.6: Add "Manual Verification" section checklist
  - [x] 7.8.7: Add "Performance" section checklist
  - [x] 7.8.8: Add "Cleanup" section checklist

### Final Verification Checklist

#### Documentation
- [x] 7.V.1: All skills documentation is accurate and complete
- [x] 7.V.2: Prime template works for all tools (OpenCode, Claude Code, Codex)
- [x] 7.V.3: AGENTS.md accurately reflects implementation
- [x] 7.V.4: User guide is comprehensive
- [x] 7.V.5: All code examples are tested and work

#### Testing
- [x] 7.V.6: `scripts/test-collaborative-workflow.sh` passes all tests
- [ ] 7.V.7: Manual test with OpenCode completed
- [x] 7.V.8: Manual test with Claude Code completed
- [ ] 7.V.9: Manual test with Codex completed
- [ ] 7.V.10: Auto-resume mode works with @agent mention

#### Polish
- [x] 7.V.11: No typos in documentation
- [x] 7.V.12: Consistent terminology across all docs
- [x] 7.V.13: Clear error messages throughout
- [x] 7.V.14: Helpful tips and best practices included

---

## Overview

This final phase focuses on documentation, polish, and final testing. All features are implemented - now we ensure they're well-documented, polished, and thoroughly tested with all three AI tools.

**What we're doing:**
- Update all skills documentation
- Update prime template for collaborative workflow
- Update AGENTS.md
- Final end-to-end testing with OpenCode, Claude Code, and Codex
- Performance review and optimization
- User documentation

---

## Prerequisites

Before starting this phase:

1. All previous phases (0-6) are complete
2. All features work individually
3. Basic testing has passed

---

## Tasks

### Task 7.1: Update `egenskriven` Skill

**File**: `internal/commands/skills/egenskriven/SKILL.md`

Add comprehensive documentation for the collaborative workflow:

```markdown
## Human-AI Collaborative Workflow

EgenSkriven supports a collaborative workflow where AI agents can request human input
and resume work once they receive a response. This enables seamless back-and-forth
communication between humans and AI coding assistants.

### Workflow Overview

```
Agent starts work on task
        │
        ▼
Agent encounters decision point
        │
        ▼
Agent blocks task: egenskriven block <task> "question"
        │
        ▼
Human sees blocked task in UI or CLI
        │
        ▼
Human adds response: egenskriven comment <task> "response"
        │
        ▼
Resume (manual, command, or auto)
        │
        ▼
Agent continues with full context
```

### Setting Up Tool Integration

Before using the collaborative workflow, set up the integration for your AI tool:

#### OpenCode
```bash
egenskriven init --opencode
```
Creates `.opencode/tool/egenskriven-session.ts`. Call the `egenskriven-session` tool to get your session ID.

#### Claude Code
```bash
egenskriven init --claude-code
```
Creates hooks that set `$CLAUDE_SESSION_ID` automatically on session start.

#### Codex CLI
```bash
egenskriven init --codex
```
Creates `.codex/get-session-id.sh` helper script.

### Linking Your Session

Before blocking a task, link your session:

```bash
# OpenCode: Call the egenskriven-session tool first, then:
egenskriven session link <task-ref> --tool opencode --ref <session-id>

# Claude Code:
egenskriven session link <task-ref> --tool claude-code --ref $CLAUDE_SESSION_ID

# Codex:
SESSION_ID=$(.codex/get-session-id.sh)
egenskriven session link <task-ref> --tool codex --ref $SESSION_ID
```

### Blocking a Task

When you need human input:

```bash
# Block with a question (atomic operation)
egenskriven block <task-ref> "What authentication approach should I use?"

# For longer questions, use stdin
echo "I need to decide between several approaches..." | egenskriven block <task-ref> --stdin
```

This will:
1. Move the task to the `need_input` column
2. Add your question as a comment
3. Preserve your session for later resume

### Adding Comments

```bash
# Add a response comment
egenskriven comment <task-ref> "Use JWT with refresh tokens"

# Add with explicit author
egenskriven comment <task-ref> "Approved" --author "jane"

# Add from stdin for longer responses
cat response.txt | egenskriven comment <task-ref> --stdin
```

### Viewing Comments

```bash
# List all comments
egenskriven comments <task-ref>

# As JSON
egenskriven comments <task-ref> --json

# Only recent comments
egenskriven comments <task-ref> --since "2026-01-07T10:00:00Z"
```

### Listing Blocked Tasks

```bash
# List all tasks needing input
egenskriven list --need-input

# As JSON
egenskriven list --need-input --json
```

### Resuming Work

Depending on the board's resume mode:

```bash
# Print the resume command (manual mode)
egenskriven resume <task-ref>

# Execute the resume directly (command mode)
egenskriven resume <task-ref> --exec

# Use minimal prompt (fewer tokens)
egenskriven resume <task-ref> --exec --minimal
```

For auto mode, add a comment with `@agent` and the session will resume automatically.

### Resume Modes

Configure per-board:

| Mode | Behavior |
|------|----------|
| `manual` | Print command for user to copy |
| `command` | User runs `egenskriven resume --exec` |
| `auto` | Auto-resume on `@agent` comment |

```bash
# Set resume mode
egenskriven board update <board> --resume-mode auto
```

### Example: Complete Workflow

```bash
# 1. Agent links session
egenskriven session link WRK-42 --tool opencode --ref abc-123

# 2. Agent works, then gets blocked
egenskriven block WRK-42 "Should I implement REST or GraphQL first?"

# 3. Human responds
egenskriven comment WRK-42 "Start with REST, we'll add GraphQL later"

# 4. Human resumes agent
egenskriven resume WRK-42 --exec

# Agent continues with full context...
```

### Tips

- Always link your session BEFORE blocking a task
- Be specific in your blocking questions
- Use the `--json` flag for scripting and automation
- Check comments periodically if working on multiple tasks
```

---

### Task 7.2: Update `egenskriven-workflows` Skill

**File**: `internal/commands/skills/egenskriven-workflows/SKILL.md`

Add section on collaborative workflow modes:

```markdown
## Resume Modes

Each board can be configured with a resume mode that controls how blocked tasks
are resumed after human input.

### Available Modes

#### Manual Mode (`manual`)
The resume command is printed for the user to copy and run manually.

```bash
$ egenskriven resume WRK-42
Resume command for WRK-42:

  opencode run '## Task Context...' --session abc-123

Working directory: /home/user/project
```

Best for: Maximum control, debugging, understanding the resume process.

#### Command Mode (`command`) - Default
User explicitly runs the resume command with `--exec` flag.

```bash
$ egenskriven resume WRK-42 --exec
Resuming session for WRK-42...
Tool: opencode
Working directory: /home/user/project

[Agent session starts]
```

Best for: Most workflows, explicit control over when resume happens.

#### Auto Mode (`auto`)
Session resumes automatically when human adds a comment containing `@agent`.

```bash
# Human adds comment:
$ egenskriven comment WRK-42 "@agent I've decided to use JWT auth"

# Session automatically resumes with full context
```

Best for: Responsive workflows, quick back-and-forth communication.

### Configuring Resume Mode

```bash
# View current mode
egenskriven board show <board>

# Change mode
egenskriven board update <board> --resume-mode auto
```

### Workflow Recommendations

| Scenario | Recommended Mode |
|----------|-----------------|
| Learning the workflow | `manual` |
| Normal development | `command` |
| Pair programming with AI | `auto` |
| Sensitive/critical tasks | `command` |
| High-frequency interaction | `auto` |

### Collaborative Mode Integration

The collaborative workflow works with all agent modes:

| Agent Mode | Blocking Behavior |
|------------|-------------------|
| `autonomous` | Block only for critical decisions |
| `collaborative` | Block for significant decisions |
| `supervised` | Block frequently for confirmation |

Example configuration:
```json
{
  "agent": {
    "mode": "collaborative"
  }
}
```

With `collaborative` mode, agents are expected to:
1. Execute minor updates directly
2. Block for major decisions requiring human input
3. Provide clear, specific questions when blocking
```

---

### Task 7.3: Update Prime Template

**File**: `internal/commands/prime.tmpl`

Add comprehensive section on collaborative workflow:

```markdown
## Human-AI Collaborative Workflow

When you need human input to proceed with a task, use the collaborative workflow.

### Before Starting Work

1. **Link your session** to the task:
   {{if eq .Tool "opencode"}}
   - Call the `egenskriven-session` tool to get your session ID
   - Run: `egenskriven session link <task-ref> --tool opencode --ref <session-id>`
   {{else if eq .Tool "claude-code"}}
   - Your session ID is available as `$CLAUDE_SESSION_ID`
   - Run: `egenskriven session link <task-ref> --tool claude-code --ref $CLAUDE_SESSION_ID`
   {{else if eq .Tool "codex"}}
   - Run: `SESSION_ID=$(.codex/get-session-id.sh)`
   - Run: `egenskriven session link <task-ref> --tool codex --ref $SESSION_ID`
   {{end}}

2. **Move task to in_progress**:
   ```bash
   egenskriven move <task-ref> in_progress
   ```

### When You Need Human Input

1. **Block the task** with your question:
   ```bash
   egenskriven block <task-ref> "Your specific question here"
   ```

2. **Your session is preserved**. The human will respond via comments and resume you.

### After Being Resumed

You will receive context that includes:
- Task information (title, priority, description)
- Full comment thread (your question + human responses)
- Instructions to continue

Example context you'll receive:
```
## Task Context (from EgenSkriven)

**Task**: WRK-42 - Implement authentication
**Status**: need_input -> in_progress
**Priority**: high

## Conversation Thread

[agent @ 10:30]: What authentication approach should I use?
[human @ 11:45]: Use JWT with refresh tokens. The refresh token should
have a 7-day expiry, and the access token should be 15 minutes.

## Instructions

Continue working on the task based on the human's response above.
```

### Best Practices

1. **Be specific** in your blocking questions
2. **Provide context** about what you've already considered
3. **List options** if you have multiple approaches in mind
4. **Link your session early** before you might need to block

### Example

```bash
# 1. Start work
egenskriven session link WRK-42 --tool {{.Tool}} --ref <your-session-id>
egenskriven move WRK-42 in_progress

# 2. Work on the task...

# 3. Encounter a decision point
egenskriven block WRK-42 "The API could use either REST or GraphQL.
REST is simpler but GraphQL offers more flexibility for the frontend.
Which approach should I implement first?"

# 4. Wait for human response (session ends here)

# 5. When resumed, you'll have full context to continue
```

### Checking for Responses

If you want to check for responses without waiting for resume:

```bash
egenskriven comments <task-ref> --json
```
```

---

### Task 7.4: Update AGENTS.md

**File**: `AGENTS.md`

Add comprehensive section:

```markdown
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

This project uses **{{.ResumeMode}}** resume mode:
{{if eq .ResumeMode "manual"}}
- Print resume command for manual execution
{{else if eq .ResumeMode "command"}}
- Run `egenskriven resume <task> --exec` to resume
{{else if eq .ResumeMode "auto"}}
- Add comment with `@agent` to trigger auto-resume
{{end}}
```

---

### Task 7.5: Create User Guide

**File**: `docs/collaborative-workflow-guide.md`

```markdown
# EgenSkriven Collaborative Workflow Guide

This guide explains how to use EgenSkriven's collaborative workflow feature
for human-AI pair programming.

## Introduction

The collaborative workflow enables AI coding assistants (OpenCode, Claude Code,
Codex) to request human input when they encounter decisions or blockers. This
creates a seamless back-and-forth communication loop.

## Getting Started

### Prerequisites

- EgenSkriven installed and configured
- One of: OpenCode, Claude Code, or Codex CLI
- A project with EgenSkriven initialized

### Setup

1. **Initialize tool integration**:
   ```bash
   cd your-project
   egenskriven init --all  # Or --opencode, --claude-code, --codex
   ```

2. **Configure resume mode** (optional):
   ```bash
   egenskriven board update main --resume-mode command  # or auto
   ```

## Workflow

### 1. Agent Starts Work

The agent picks up a task and links its session:

```
Agent: I'll work on WRK-42. Let me link my session first.
[Agent calls egenskriven-session tool]
[Agent runs: egenskriven session link WRK-42 --tool opencode --ref abc-123]
[Agent runs: egenskriven move WRK-42 in_progress]
```

### 2. Agent Gets Blocked

The agent encounters a decision point:

```
Agent: I need guidance on the authentication approach.
[Agent runs: egenskriven block WRK-42 "Should I use JWT or sessions?"]

Task moves to "need_input" column.
Agent session ends (but is preserved).
```

### 3. Human Responds

You see the blocked task in the UI or CLI:

```bash
$ egenskriven list --need-input
WRK-42  Implement auth  need_input  high

$ egenskriven comments WRK-42
[agent @ 10:30]: Should I use JWT or sessions?
```

Add your response:

```bash
$ egenskriven comment WRK-42 "Use JWT with refresh tokens"
```

### 4. Resume the Agent

Depending on your resume mode:

**Command mode** (default):
```bash
$ egenskriven resume WRK-42 --exec
Resuming session for WRK-42...
[Agent continues]
```

**Auto mode**:
```bash
$ egenskriven comment WRK-42 "@agent Use JWT with refresh tokens"
# Agent automatically resumes
```

### 5. Agent Continues

The agent receives full context:

```
## Task Context (from EgenSkriven)
Task: WRK-42 - Implement auth
Priority: high

## Conversation Thread
[agent @ 10:30]: Should I use JWT or sessions?
[human @ 10:45]: Use JWT with refresh tokens

## Instructions
Continue working on the task based on the above.
```

## Configuration

### Resume Modes

| Mode | Trigger | Best For |
|------|---------|----------|
| manual | Copy command | Learning, debugging |
| command | `resume --exec` | Most workflows |
| auto | `@agent` comment | Quick interaction |

### Setting Resume Mode

```bash
# Per-board configuration
egenskriven board update <board> --resume-mode <mode>
```

## Tips

1. **Be specific in questions**: Help the human understand the context
2. **List options**: If you have alternatives, list them
3. **Link early**: Link session before you might need to block
4. **Check for updates**: Use `egenskriven comments <task>` to check responses

## Troubleshooting

### "No session linked"

Link your session before blocking:
```bash
egenskriven session link <task> --tool <tool> --ref <id>
```

### "Task not in need_input"

Only blocked tasks can be resumed. Current column:
```bash
egenskriven show <task>
```

### Session ID not found

Re-run tool integration setup:
```bash
egenskriven init --<tool> --force
```
```

---

### Task 7.6: Final Testing with All Tools

Create a comprehensive test script:

**File**: `scripts/test-collaborative-workflow.sh`

```bash
#!/bin/bash
# Comprehensive test of collaborative workflow with all tools
# Run this manually to verify the full flow

set -e

echo "=== EgenSkriven Collaborative Workflow Test ==="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

pass() { echo -e "${GREEN}PASS${NC}: $1"; }
fail() { echo -e "${RED}FAIL${NC}: $1"; exit 1; }

# Setup
echo "Setting up test environment..."
TEST_DIR=$(mktemp -d)
cd "$TEST_DIR"
egenskriven init

# Create board with auto mode
egenskriven board create test-board
egenskriven board update test-board --resume-mode auto

echo ""
echo "=== Test 1: Basic block and comment ==="
egenskriven add "Test task 1" --board test-board --type feature
egenskriven block WRK-1 "Test question?"
COLUMN=$(egenskriven show WRK-1 --json | jq -r '.column')
[ "$COLUMN" == "need_input" ] && pass "Task blocked" || fail "Task not blocked"

egenskriven comment WRK-1 "Test response"
COMMENTS=$(egenskriven comments WRK-1 --json | jq '.count')
[ "$COMMENTS" -eq 2 ] && pass "Comments added" || fail "Comments not added"

echo ""
echo "=== Test 2: Session linking ==="
egenskriven add "Test task 2" --board test-board --type feature
egenskriven session link WRK-2 --tool opencode --ref test-session-123
SESSION=$(egenskriven session show WRK-2 --json | jq -r '.session.tool')
[ "$SESSION" == "opencode" ] && pass "Session linked" || fail "Session not linked"

echo ""
echo "=== Test 3: Resume command generation ==="
egenskriven block WRK-2 "Another question?"
RESUME_CMD=$(egenskriven resume WRK-2 --json | jq -r '.command')
[[ "$RESUME_CMD" == *"opencode run"* ]] && pass "Resume command correct" || fail "Resume command wrong"
[[ "$RESUME_CMD" == *"test-session-123"* ]] && pass "Session ref in command" || fail "Session ref missing"

echo ""
echo "=== Test 4: Tool integrations ==="
egenskriven init --opencode --force
[ -f ".opencode/tool/egenskriven-session.ts" ] && pass "OpenCode tool created" || fail "OpenCode tool missing"

egenskriven init --claude-code --force
[ -f ".claude/hooks/egenskriven-session.sh" ] && pass "Claude Code hook created" || fail "Claude Code hook missing"
[ -f ".claude/settings.json" ] && pass "Claude Code settings created" || fail "Claude Code settings missing"

egenskriven init --codex --force
[ -f ".codex/get-session-id.sh" ] && pass "Codex helper created" || fail "Codex helper missing"

echo ""
echo "=== Test 5: List --need-input ==="
egenskriven add "Test task 3" --board test-board
egenskriven block WRK-3 "Question 3?"
NEED_INPUT=$(egenskriven list --need-input --json | jq '.count')
[ "$NEED_INPUT" -ge 2 ] && pass "Need input filter works" || fail "Need input filter broken"

echo ""
echo "=== All tests passed! ==="
echo "Test directory: $TEST_DIR"
echo ""
echo "Manual testing recommended:"
echo "1. Test with actual OpenCode session"
echo "2. Test with actual Claude Code session"
echo "3. Test with actual Codex session"
echo "4. Test auto-resume with @agent mention"
```

---

### Task 7.7: Performance Review

Review and document performance characteristics:

**Areas to check:**

1. **Comment loading**: Verify comments load quickly (< 200ms for 100 comments)
2. **Resume command generation**: Should be instant
3. **Auto-resume trigger**: Should fire within 1 second of comment creation
4. **Real-time updates**: Verify SSE subscriptions work smoothly

**File**: `docs/performance-notes.md`

```markdown
# Performance Notes

## Collaborative Workflow Performance

### Expected Performance

| Operation | Target | Notes |
|-----------|--------|-------|
| Block task | < 100ms | Atomic transaction |
| Add comment | < 100ms | With mention extraction |
| List comments | < 200ms | Up to 100 comments |
| Resume generation | < 50ms | Context building |
| Auto-resume trigger | < 1s | Background execution |

### Database Indexes

The following indexes are used:
- `idx_comments_task` - Fast comment lookup by task
- `idx_comments_created` - Chronological ordering
- `idx_sessions_task` - Session lookup by task
- `idx_sessions_external_ref` - Session lookup by ref

### Optimizations

1. Comments are fetched with ascending sort (oldest first) for chronological display
2. Context prompt is built in-memory, no additional DB queries
3. Auto-resume runs in goroutine to not block comment creation
4. Real-time subscriptions use PocketBase's SSE implementation
```

---

### Task 7.8: Final Checklist

Create a release checklist:

**File**: `docs/release-checklist.md`

```markdown
# AI Workflow Feature Release Checklist

## Documentation

- [ ] SKILL.md (egenskriven) updated
- [ ] SKILL.md (egenskriven-workflows) updated
- [ ] prime.tmpl updated
- [ ] AGENTS.md updated
- [ ] collaborative-workflow-guide.md created
- [ ] All command help text accurate

## Testing

### Unit Tests
- [ ] Phase 0: Migration tests pass
- [ ] Phase 1: Command tests pass
- [ ] Phase 2: Session tests pass
- [ ] Phase 3: Resume tests pass
- [ ] Phase 4: Init tests pass
- [ ] Phase 5: UI component tests pass
- [ ] Phase 6: Auto-resume tests pass

### Integration Tests
- [ ] Full workflow: block -> comment -> resume
- [ ] Session handoff between tools
- [ ] Auto-resume with @agent mention

### E2E Tests
- [ ] OpenCode: Full workflow tested
- [ ] Claude Code: Full workflow tested
- [ ] Codex: Full workflow tested

## Manual Verification

- [ ] UI looks correct in light mode
- [ ] UI looks correct in dark mode
- [ ] Mobile responsive
- [ ] Real-time updates work
- [ ] Error states display correctly

## Performance

- [ ] No noticeable slowdown in task operations
- [ ] Comments load quickly
- [ ] Resume command generates quickly

## Cleanup

- [ ] No console.log statements left
- [ ] No TODO comments left
- [ ] No debug flags enabled
- [ ] All code properly formatted
```

---

## Testing Checklist

Before considering this phase complete:

### Documentation

- [ ] All skills documentation accurate and complete
- [ ] Prime template works for all tools
- [ ] AGENTS.md is accurate
- [ ] User guide is comprehensive
- [ ] All examples work

### Testing

- [ ] Test script passes
- [ ] Manual test with OpenCode
- [ ] Manual test with Claude Code
- [ ] Manual test with Codex
- [ ] Auto-resume works

### Polish

- [ ] No typos in documentation
- [ ] Consistent terminology
- [ ] Clear error messages
- [ ] Helpful tips included

---

## Files Changed/Created

| File | Change Type | Description |
|------|-------------|-------------|
| `internal/commands/skills/egenskriven/SKILL.md` | Modified | Add workflow docs |
| `internal/commands/skills/egenskriven-workflows/SKILL.md` | Modified | Add resume modes |
| `internal/commands/prime.tmpl` | Modified | Add workflow section |
| `AGENTS.md` | Modified | Add workflow section |
| `docs/collaborative-workflow-guide.md` | New | User guide |
| `docs/performance-notes.md` | New | Performance docs |
| `docs/release-checklist.md` | New | Release checklist |
| `scripts/test-collaborative-workflow.sh` | New | Test script |

---

## Completion

After this phase:

1. **All features implemented** and tested
2. **Documentation complete** for users and developers
3. **Ready for release** as part of EgenSkriven v1.0.0

Congratulations on completing the AI-Human Collaborative Workflow feature!
