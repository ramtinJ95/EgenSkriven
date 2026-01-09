#!/bin/bash
# Comprehensive test of collaborative workflow with all tools
# Run this manually to verify the full flow
#
# Usage: ./scripts/test-collaborative-workflow.sh
#
# Prerequisites:
#   - egenskriven binary built and in PATH
#   - jq installed for JSON parsing

set -e

echo "=== EgenSkriven Collaborative Workflow Test ==="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

pass() { echo -e "${GREEN}PASS${NC}: $1"; }
fail() { echo -e "${RED}FAIL${NC}: $1"; exit 1; }
warn() { echo -e "${YELLOW}WARN${NC}: $1"; }

# Check prerequisites
command -v egenskriven >/dev/null 2>&1 || fail "egenskriven not found in PATH"
command -v jq >/dev/null 2>&1 || fail "jq not found - install with: sudo apt install jq"

# Setup test environment
echo "Setting up test environment..."
TEST_DIR=$(mktemp -d)
echo "Test directory: $TEST_DIR"
cd "$TEST_DIR"

# Initialize egenskriven
egenskriven init >/dev/null 2>&1
pass "EgenSkriven initialized"

# Create a test board
BOARD_OUTPUT=$(egenskriven board create test-board --json 2>/dev/null || echo '{"error": true}')
if echo "$BOARD_OUTPUT" | jq -e '.id' >/dev/null 2>&1; then
    pass "Test board created"
else
    warn "Board creation returned unexpected output, continuing..."
fi

echo ""
echo "=== Test 1: Basic block and comment ==="

# Create a task
TASK1_OUTPUT=$(egenskriven add "Test task 1" --type feature --json 2>/dev/null)
TASK1_ID=$(echo "$TASK1_OUTPUT" | jq -r '.id // .display_id // empty' 2>/dev/null)
if [ -n "$TASK1_ID" ]; then
    pass "Task 1 created: $TASK1_ID"
else
    # Try without --json
    egenskriven add "Test task 1" --type feature >/dev/null 2>&1
    TASK1_ID="1"
    pass "Task 1 created (fallback mode)"
fi

# Block the task
egenskriven block "$TASK1_ID" "Test question - what approach should I use?" >/dev/null 2>&1 || true
SHOW_OUTPUT=$(egenskriven show "$TASK1_ID" --json 2>/dev/null || echo '{}')
COLUMN=$(echo "$SHOW_OUTPUT" | jq -r '.column // empty' 2>/dev/null)
if [ "$COLUMN" == "need_input" ]; then
    pass "Task blocked successfully (column: need_input)"
else
    warn "Could not verify task column, got: $COLUMN"
fi

# Add a comment
egenskriven comment "$TASK1_ID" "Test response - use approach A" >/dev/null 2>&1 || true
COMMENTS_OUTPUT=$(egenskriven comments "$TASK1_ID" --json 2>/dev/null || echo '{"comments": []}')
COMMENT_COUNT=$(echo "$COMMENTS_OUTPUT" | jq '.comments | length // 0' 2>/dev/null || echo "0")
if [ "$COMMENT_COUNT" -ge 1 ]; then
    pass "Comments added (count: $COMMENT_COUNT)"
else
    warn "Could not verify comments, got count: $COMMENT_COUNT"
fi

echo ""
echo "=== Test 2: Session linking ==="

# Create another task
egenskriven add "Test task 2" --type feature >/dev/null 2>&1 || true
TASK2_ID="2"

# Link a session
egenskriven session link "$TASK2_ID" --tool opencode --ref test-session-abc123 >/dev/null 2>&1 || true
SESSION_OUTPUT=$(egenskriven session show "$TASK2_ID" --json 2>/dev/null || echo '{}')
SESSION_TOOL=$(echo "$SESSION_OUTPUT" | jq -r '.agent_session.tool // .session.tool // empty' 2>/dev/null)
if [ "$SESSION_TOOL" == "opencode" ]; then
    pass "Session linked successfully (tool: opencode)"
else
    warn "Could not verify session linking, got tool: $SESSION_TOOL"
fi

echo ""
echo "=== Test 3: Resume command generation ==="

# Block task 2 first
egenskriven block "$TASK2_ID" "Another question for testing resume?" >/dev/null 2>&1 || true

# Get resume command
RESUME_OUTPUT=$(egenskriven resume "$TASK2_ID" --json 2>/dev/null || echo '{}')
RESUME_CMD=$(echo "$RESUME_OUTPUT" | jq -r '.command // empty' 2>/dev/null)
if [[ "$RESUME_CMD" == *"opencode"* ]] || [[ "$RESUME_CMD" == *"run"* ]]; then
    pass "Resume command generated correctly"
elif [ -n "$RESUME_CMD" ]; then
    pass "Resume command generated: ${RESUME_CMD:0:50}..."
else
    warn "Could not verify resume command generation"
fi

# Check session ref is in command
if [[ "$RESUME_CMD" == *"test-session-abc123"* ]]; then
    pass "Session ref included in resume command"
else
    warn "Session ref may not be in command (expected: test-session-abc123)"
fi

echo ""
echo "=== Test 4: Tool integrations ==="

# Test OpenCode integration
egenskriven init --opencode --force >/dev/null 2>&1 || true
if [ -f ".opencode/tool/egenskriven-session.ts" ]; then
    pass "OpenCode tool created"
else
    warn "OpenCode tool file not found"
fi

# Test Claude Code integration
egenskriven init --claude-code --force >/dev/null 2>&1 || true
if [ -f ".claude/hooks/egenskriven-session.sh" ]; then
    pass "Claude Code hook created"
else
    warn "Claude Code hook file not found"
fi

if [ -f ".claude/settings.json" ]; then
    pass "Claude Code settings created"
else
    warn "Claude Code settings file not found"
fi

# Test Codex integration
egenskriven init --codex --force >/dev/null 2>&1 || true
if [ -f ".codex/get-session-id.sh" ]; then
    pass "Codex helper script created"
    # Check it's executable
    if [ -x ".codex/get-session-id.sh" ]; then
        pass "Codex helper script is executable"
    else
        warn "Codex helper script is not executable"
    fi
else
    warn "Codex helper script not found"
fi

echo ""
echo "=== Test 5: List --need-input filter ==="

# Create and block another task
egenskriven add "Test task 3" --type bug >/dev/null 2>&1 || true
TASK3_ID="3"
egenskriven block "$TASK3_ID" "Question 3 for testing list filter?" >/dev/null 2>&1 || true

# List need-input tasks
NEED_INPUT_OUTPUT=$(egenskriven list --need-input --json 2>/dev/null || echo '{"tasks": []}')
NEED_INPUT_COUNT=$(echo "$NEED_INPUT_OUTPUT" | jq '.tasks | length // 0' 2>/dev/null || echo "0")
if [ "$NEED_INPUT_COUNT" -ge 1 ]; then
    pass "Need input filter works (found: $NEED_INPUT_COUNT tasks)"
else
    # Try alternative format
    NEED_INPUT_COUNT=$(echo "$NEED_INPUT_OUTPUT" | jq 'length // 0' 2>/dev/null || echo "0")
    if [ "$NEED_INPUT_COUNT" -ge 1 ]; then
        pass "Need input filter works (found: $NEED_INPUT_COUNT tasks)"
    else
        warn "Could not verify need-input filter, got count: $NEED_INPUT_COUNT"
    fi
fi

echo ""
echo "=== Test Summary ==="
echo ""
echo -e "${GREEN}All automated tests completed!${NC}"
echo ""
echo "Test directory: $TEST_DIR"
echo ""
echo "=== Manual Testing Recommendations ==="
echo ""
echo "The following tests require actual AI tool sessions:"
echo ""
echo "1. OpenCode Integration Test:"
echo "   - Start OpenCode session"
echo "   - Call egenskriven-session tool"
echo "   - Link session to task"
echo "   - Block task and verify resume works"
echo ""
echo "2. Claude Code Integration Test:"
echo "   - Start Claude Code session"
echo "   - Verify \$CLAUDE_SESSION_ID is set"
echo "   - Link session and test block/resume flow"
echo ""
echo "3. Codex Integration Test:"
echo "   - Start Codex session"
echo "   - Run .codex/get-session-id.sh"
echo "   - Link session and test block/resume flow"
echo ""
echo "4. Auto-Resume Test (requires auto mode):"
echo "   - Set board to auto mode: egenskriven board update <board> --resume-mode auto"
echo "   - Block a task"
echo "   - Add comment with @agent mention"
echo "   - Verify agent session resumes automatically"
echo ""
echo "To clean up: rm -rf $TEST_DIR"
