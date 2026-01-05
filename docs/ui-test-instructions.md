# UI Test Instructions

This document contains instructions for running UI tests using the `ui-test-engineer` agent.

---

## Prerequisites

### Step 1: Start the Dev Server FIRST

Before calling the `ui-test-engineer` agent, you **MUST** start the development server:

```bash
cd /home/ramtinj/personal-workspace/EgenSkriven/ui && npm run dev &
sleep 3
curl -s -o /dev/null -w "%{http_code}" http://localhost:5173
# Should return: 200
```

**Important Notes:**
- The server may use a different port if 5173 is busy (5174, 5175, 5176, etc.)
- Check the output of `npm run dev` to see which port is being used
- Wait for the server to be ready before running tests

---

## How to Call ui-test-engineer

Use the `task` tool with `subagent_type: ui-test-engineer`.

### Required Information in Prompt

1. **App URL**: The URL where the app is running (e.g., `http://localhost:5173`)
2. **Test Steps**: Numbered, specific actions to perform
3. **Expected Behavior**: What should happen for each step
4. **Required Output**: Request a structured report format

### Example Task Call

```
task(
  description="Test Settings panel",
  prompt="Test the Settings panel at http://localhost:5173...",
  subagent_type="ui-test-engineer"
)
```

---

## Prompt Template

Use this template when calling `ui-test-engineer`:

```
Perform [TYPE OF TEST] at http://localhost:[PORT]

## Test Areas

### 1. [Feature Name]
- Step 1: [Action to perform]
- Step 2: [Action to perform]
- Expected: [What should happen]

### 2. [Feature Name]
- Step 1: [Action to perform]
- Expected: [What should happen]

## Required Output

Return a detailed report with:
1. PASS/FAIL status for each test area
2. Specific notes on any failures
3. Screenshots or evidence where relevant
4. Console error check results
```

---

## Report Output Format

Request the agent to return reports in this format:

```markdown
# [Test Name] Report

## Test Summary
[Brief description of what was tested]

## Environment
- **URL**: http://localhost:[PORT]
- **Browser**: Chromium (Playwright)
- **Date**: [Date]

---

## Test Results

### 1. [Feature Name] [STATUS EMOJI] [PASS/FAIL]

| Test Case | Status | Notes |
|-----------|--------|-------|
| 1.1 [Test case] | [EMOJI] [STATUS] | [Details] |
| 1.2 [Test case] | [EMOJI] [STATUS] | [Details] |

### 2. [Feature Name] [STATUS EMOJI] [PASS/FAIL]

| Test Case | Status | Notes |
|-----------|--------|-------|
| 2.1 [Test case] | [EMOJI] [STATUS] | [Details] |

---

## Console Errors
[List any console errors, or "No console errors"]

---

## Final Verdict

| Category | Status |
|----------|--------|
| [Category 1] | [EMOJI] [STATUS] |
| [Category 2] | [EMOJI] [STATUS] |

**Overall: [EMOJI] [PASS/FAIL]**
```

### Status Indicators
- Use checkmark emoji for pass
- Use warning emoji for partial pass
- Use X emoji for fail

---

## Common Test Scenarios

### Theme Testing
```
## Theme System
- Open settings with Ctrl+, (or Cmd+,)
- Switch between System, Light, and Dark themes
- Verify colors change correctly for each theme
- Refresh the page - verify theme persists (localStorage)
```

### Keyboard Navigation Testing
```
## Keyboard Navigation
- Use Tab key to navigate through the interface
- Verify focus rings appear on focusable elements
- Verify interactive elements can be focused
```

### Drag and Drop Testing
```
## Drag and Drop
- Drag an item
- Verify original item fades (opacity change)
- Verify drag overlay appears with visual effects
- Drop the item and verify normal state
```

### Settings Panel Testing
```
## Settings Panel
- Open with Ctrl+, (or Cmd+,)
- Verify it closes with Escape key
- Verify clicking outside the panel closes it
- Verify X button closes it
```

---

## Full E2E Test Template

For comprehensive end-to-end testing, use this template:

```
Perform a comprehensive end-to-end test of the application at http://localhost:[PORT]

## 1. Settings Panel
- Open settings with Ctrl+, (or Cmd+,)
- Verify the panel opens correctly
- Click outside the panel - it should close
- Open again, press Escape - verify it closes
- Open again, click X button - verify it closes

## 2. Theme Switching
- Open settings
- Switch between System, Light, and Dark themes
- Verify colors change correctly
- Refresh page and verify theme persists

## 3. Accent Colors
- In settings, select different accent colors
- Verify the accent color changes visually
- Verify selected color shows a checkmark

## 4. Task/Item Interaction
- Click on an item - verify detail panel opens
- Close the detail panel
- Click on a different item - verify it works

## 5. Drag and Drop
- Drag an item to a different position/column
- Verify original fades during drag
- Verify drag overlay has visual effects
- Drop and verify item moves correctly

## 6. Keyboard Navigation
- Use Tab to navigate through interface
- Verify focus rings appear
- Verify items can be focused

## 7. General Functionality
- Verify all UI elements are visible
- Verify counts/labels are correct
- Verify no console errors

Return a detailed PASS/FAIL report for each test area.
```

---

## Tips

1. **Always check the port**: The dev server may use different ports if the default is busy
2. **Wait for server**: Give the server 2-3 seconds to start before testing
3. **Be specific**: Provide exact actions and expected outcomes
4. **Request console check**: Always ask for console error verification
5. **Use tables**: Request tabular format for easy scanning of results

---

## Troubleshooting

### Server not starting
```bash
# Kill any existing processes on common ports
lsof -ti:5173 | xargs kill -9 2>/dev/null
lsof -ti:5174 | xargs kill -9 2>/dev/null
lsof -ti:5175 | xargs kill -9 2>/dev/null
```

### Wrong port
Check the npm run dev output for the actual port being used:
```
VITE v7.3.0  ready in 154 ms
  Local:   http://localhost:5176/   <-- Use this port
```

### Tests timing out
Increase the timeout or ensure the server is fully ready before testing.
