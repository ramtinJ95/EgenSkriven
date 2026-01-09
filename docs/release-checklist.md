# AI Workflow Feature Release Checklist

Pre-release checklist for the Human-AI Collaborative Workflow feature
in EgenSkriven v1.0.0.

## Documentation

- [ ] SKILL.md (egenskriven) updated with collaborative workflow section
- [ ] SKILL.md (egenskriven-workflows) updated with resume modes
- [ ] prime.tmpl updated with workflow instructions
- [ ] AGENTS.md updated with workflow quick start
- [ ] collaborative-workflow-guide.md created and complete
- [ ] performance-notes.md created with benchmarks
- [ ] All command help text is accurate
- [ ] README mentions collaborative workflow feature

## Unit Tests

### Phase 0: Data Model
- [ ] Migration tests pass for `need_input` column
- [ ] Migration tests pass for `agent_session` field
- [ ] Migration tests pass for comments collection
- [ ] Migration tests pass for sessions collection
- [ ] Migration tests pass for `resume_mode` board field

### Phase 1: CLI Commands
- [ ] `block` command tests pass
- [ ] `comment` command tests pass
- [ ] `comments` command tests pass
- [ ] `list --need-input` filter tests pass

### Phase 2: Session Management
- [ ] `session link` command tests pass
- [ ] `session show` command tests pass
- [ ] `session history` command tests pass
- [ ] Session state transitions tested

### Phase 3: Resume Flow
- [ ] `resume` command tests pass (print mode)
- [ ] `resume --exec` tests pass
- [ ] Context prompt builder tests pass
- [ ] Resume command generation tests for all tools

### Phase 4: Tool Integrations
- [ ] `init --opencode` tests pass
- [ ] `init --claude-code` tests pass
- [ ] `init --codex` tests pass
- [ ] `init --all` tests pass
- [ ] `--force` flag overwrites correctly

### Phase 5: Web UI
- [ ] CommentsPanel component tests pass
- [ ] SessionInfo component tests pass
- [ ] Comments hook tests pass
- [ ] Real-time subscription tests pass
- [ ] Need input badge renders correctly

### Phase 6: Auto-Resume
- [ ] @agent mention detection tests pass
- [ ] Auto-resume trigger tests pass
- [ ] Board resume_mode configuration tests pass
- [ ] BoardSettingsModal tests pass

## Integration Tests

- [ ] Full workflow: block -> comment -> resume
- [ ] Session handoff between tools (link new session)
- [ ] Auto-resume with @agent mention
- [ ] Multiple comments before resume
- [ ] Resume with minimal context flag
- [ ] Error handling for missing session
- [ ] Error handling for wrong task state

## E2E Tests

### OpenCode
- [ ] Initialize integration: `egenskriven init --opencode`
- [ ] Verify tool file created: `.opencode/tool/egenskriven-session.ts`
- [ ] Start OpenCode session
- [ ] Call egenskriven-session tool
- [ ] Link session to task
- [ ] Block task with question
- [ ] Add human response
- [ ] Resume session
- [ ] Verify agent receives context

### Claude Code
- [ ] Initialize integration: `egenskriven init --claude-code`
- [ ] Verify hook created: `.claude/hooks/egenskriven-session.sh`
- [ ] Verify settings updated: `.claude/settings.json`
- [ ] Start Claude Code session
- [ ] Verify $CLAUDE_SESSION_ID is set
- [ ] Link session to task
- [ ] Block task with question
- [ ] Add human response
- [ ] Resume session
- [ ] Verify agent receives context

### Codex CLI
- [ ] Initialize integration: `egenskriven init --codex`
- [ ] Verify helper created: `.codex/get-session-id.sh`
- [ ] Verify script is executable
- [ ] Start Codex session
- [ ] Run helper script to get session ID
- [ ] Link session to task
- [ ] Block task with question
- [ ] Add human response
- [ ] Resume session
- [ ] Verify agent receives context

## Manual Verification

### UI Appearance
- [ ] Comments panel displays correctly in light mode
- [ ] Comments panel displays correctly in dark mode
- [ ] Session info displays correctly
- [ ] Need input badge animates (pulsing dot)
- [ ] Resume button is visible for blocked tasks
- [ ] Mobile responsive layout works

### UI Functionality
- [ ] Add comment via UI works
- [ ] Comments appear in real-time
- [ ] Resume button triggers resume flow
- [ ] Board settings modal opens
- [ ] Resume mode can be changed via UI
- [ ] Auto-resume indicator shows when enabled

### Error States
- [ ] "No session linked" error displays correctly
- [ ] "Task not in need_input" error displays correctly
- [ ] Invalid task reference error displays correctly
- [ ] Network error handling works

## Performance

- [ ] Block task completes in < 100ms
- [ ] Add comment completes in < 100ms
- [ ] List comments completes in < 200ms (100 comments)
- [ ] Resume command generation completes in < 50ms
- [ ] Auto-resume triggers within 1 second
- [ ] No noticeable slowdown in normal task operations
- [ ] Database size reasonable after extended use

## Cleanup

- [ ] No console.log statements in production code
- [ ] No TODO comments remaining in new code
- [ ] No debug flags enabled
- [ ] All code properly formatted (go fmt, prettier)
- [ ] No unused imports or variables
- [ ] License headers present where required
- [ ] Git history is clean (no WIP commits)

## Final Sign-off

- [ ] All automated tests pass
- [ ] All manual tests pass
- [ ] Documentation reviewed by human
- [ ] Feature demo completed
- [ ] Ready for release

---

## Notes

Record any issues found during testing:

| Issue | Status | Notes |
|-------|--------|-------|
| | | |

---

**Reviewer**: _______________

**Date**: _______________

**Approved for Release**: [ ] Yes  [ ] No
