# Phase 7: Polish - Implementation Context

This document contains all context needed to continue Phase 7 implementation in a new session.

**Last Updated**: Session ended after completing tasks 7.0 through 7.1-test

## Current Progress Summary

### ✅ COMPLETED TASKS

| Task | Description | Commit |
|------|-------------|--------|
| 7.0 | CSS refactor - split tokens.css into tokens.css + base.css, create index.css | `refactor(ui): split CSS into modular files with index entry point` |
| 7.0-test | Verified build passes, no visual regressions | (verified via build) |
| 7.0b | Created `ui/src/contexts/` directory, moved SelectionProvider, selectionContext, useCurrentBoard | `refactor(ui): move context files to dedicated contexts/ directory` |
| 7.0b-test | Verified context migration - build passes, tests pass | (verified via build + tests) |
| 7.1 | Created theme-light.css with light mode color overrides | `feat(ui): add light mode theme CSS with color token overrides` |
| 7.1-test | Verified light mode CSS via ui-test-engineer - ALL PASS | (verified via ui-test-engineer) |
| 7.11 | Updated main.tsx to import index.css | (done as part of 7.0) |
| 7.14 | Fixed all import paths after context migration | (done as part of 7.0b) |
| 7.14-test | Verified imports - build passes | (verified via build) |

### 🔜 NEXT TASK TO DO

**Task 7.2**: Create ThemeContext.tsx in contexts/
- Theme state management (system/light/dark)
- localStorage persistence
- System preference detection via `prefers-color-scheme` media query
- See `docs/phase-7.md` for full implementation code

### 📋 REMAINING TASKS

| Task | Description | Priority |
|------|-------------|----------|
| **7.2** | **Create ThemeContext.tsx** | **High** |
| 7.2-test | Verify ThemeContext with ui-test-engineer | High |
| 7.3 | Create Settings.tsx & Settings.module.css | High |
| 7.3-test | Verify Settings panel with ui-test-engineer | High |
| 7.4 | Create useAccentColor.ts hook | High |
| 7.4-test | Verify accent colors with ui-test-engineer | High |
| 7.5 | Create animations.css | Medium |
| 7.5-test | Verify animations with ui-test-engineer | Medium |
| 7.6 | Create Skeleton.tsx & Skeleton.module.css | Medium |
| 7.6-test | Verify skeletons with ui-test-engineer | Medium |
| 7.7 | Create focus.css | Medium |
| 7.7-test | Verify focus styles with ui-test-engineer | Medium |
| 7.8 | Create drag-drop.css | Medium |
| 7.8-test | Verify drag-drop with ui-test-engineer | Medium |
| 7.9 | Install @tanstack/react-virtual, create VirtualizedColumn.tsx | Medium |
| 7.9-test | Verify virtualization with ui-test-engineer | Medium |
| 7.10 | Update App.tsx (ThemeProvider, Cmd+, shortcut) | High |
| 7.10-test | Verify App.tsx with ui-test-engineer | High |
| 7.12 | Update Board.tsx (skeleton, virtualization) | Medium |
| 7.12-test | Verify Board.tsx with ui-test-engineer | Medium |
| 7.13 | Update TaskCard.tsx (memo, drag overlay) | Medium |
| 7.13-test | Verify TaskCard.tsx with ui-test-engineer | Medium |
| 7.15 | FINAL regression test with ui-test-engineer | High |

---

## Current Codebase State (After Completed Tasks)

### File Structure NOW
```
ui/src/
├── styles/
│   ├── index.css           # ✅ Entry point (imports tokens, base, theme-light)
│   ├── tokens.css          # ✅ CSS variables only (dark mode defaults)
│   ├── base.css            # ✅ Reset, body, typography
│   ├── theme-light.css     # ✅ Light mode overrides [data-theme="light"]
│   ├── animations.css      # ❌ TODO
│   ├── focus.css           # ❌ TODO
│   └── drag-drop.css       # ❌ TODO
├── contexts/
│   ├── index.ts            # ✅ Re-exports all contexts
│   ├── SelectionContext.ts # ✅ Moved from hooks/
│   ├── SelectionProvider.tsx # ✅ Moved from hooks/
│   ├── CurrentBoardContext.tsx # ✅ Moved from hooks/
│   └── ThemeContext.tsx    # ❌ TODO (next task)
├── hooks/
│   ├── useSelection.ts     # ✅ Updated import path
│   ├── useAccentColor.ts   # ❌ TODO
│   └── ...existing hooks
├── components/
│   ├── Board.tsx           # ✅ Updated import path
│   ├── Sidebar.tsx         # ✅ Updated import path
│   ├── Settings.tsx        # ❌ TODO
│   ├── Skeleton.tsx        # ❌ TODO
│   ├── VirtualizedColumn.tsx # ❌ TODO
│   └── ...existing components
├── App.tsx                 # ✅ Updated import paths (needs ThemeProvider wrap)
└── main.tsx                # ✅ Imports index.css
```

### What's Working
- ✅ Dark mode (default) - all CSS variables in tokens.css
- ✅ Light mode CSS - theme-light.css overrides when `data-theme="light"` is set
- ✅ Context architecture - contexts/ directory with proper exports
- ✅ All existing functionality (selection, board switching, tasks, etc.)
- ✅ Build passes, all tests pass

### What's NOT Working Yet
- ❌ No way for users to switch themes (needs ThemeContext + Settings panel)
- ❌ No accent color customization
- ❌ No loading skeletons
- ❌ No virtualization for large lists
- ❌ No enhanced animations

---

## Todo List JSON

Copy this JSON to recreate the todo list using the `todowrite` tool:

```json
[
  {
    "id": "7.0",
    "content": "7.0 PREP: Refactor CSS - Split tokens.css into tokens.css (variables only) + base.css (reset/base styles), create index.css entry point",
    "status": "completed",
    "priority": "high"
  },
  {
    "id": "7.0-test",
    "content": "7.0-TEST: Verify CSS refactor - App still renders correctly, all existing styles work, no visual regressions",
    "status": "completed",
    "priority": "high"
  },
  {
    "id": "7.0b",
    "content": "7.0b PREP: Create ui/src/contexts/ directory and move existing context files (SelectionProvider, selectionContext, useCurrentBoard) there",
    "status": "completed",
    "priority": "high"
  },
  {
    "id": "7.0b-test",
    "content": "7.0b-TEST: Verify context migration - Selection still works, board switching works, no runtime errors",
    "status": "completed",
    "priority": "high"
  },
  {
    "id": "7.1",
    "content": "7.1 Create theme-light.css - Light mode color token overrides with [data-theme=\"light\"] selector",
    "status": "completed",
    "priority": "high"
  },
  {
    "id": "7.1-test",
    "content": "7.1-TEST: Verify light mode CSS - Manually set data-theme=\"light\" on html element, verify all colors change correctly",
    "status": "completed",
    "priority": "high"
  },
  {
    "id": "7.2",
    "content": "7.2 Create ThemeContext.tsx in contexts/ - Theme state (system/light/dark), localStorage persistence, system preference detection",
    "status": "pending",
    "priority": "high"
  },
  {
    "id": "7.2-test",
    "content": "7.2-TEST: Verify ThemeContext - Theme switches between dark/light/system, persists after refresh, responds to OS preference changes",
    "status": "pending",
    "priority": "high"
  },
  {
    "id": "7.3",
    "content": "7.3 Create Settings.tsx & Settings.module.css - Settings panel with theme dropdown and accent color grid (Cmd+, shortcut)",
    "status": "pending",
    "priority": "high"
  },
  {
    "id": "7.3-test",
    "content": "7.3-TEST: Verify Settings panel - Opens with Cmd+,, closes with Escape/outside click, theme dropdown works, accent colors selectable",
    "status": "pending",
    "priority": "high"
  },
  {
    "id": "7.4",
    "content": "7.4 Create useAccentColor.ts hook - Accent color management with 8 colors, --accent-rgb variable, localStorage persistence",
    "status": "pending",
    "priority": "high"
  },
  {
    "id": "7.4-test",
    "content": "7.4-TEST: Verify accent colors - All 8 colors apply correctly, --accent and --accent-rgb update, persists after refresh",
    "status": "pending",
    "priority": "high"
  },
  {
    "id": "7.5",
    "content": "7.5 Create animations.css - Centralized keyframes (fadeIn, scaleIn, slideInFromRight, slideDown, slideUp) + utility classes + reduced motion support",
    "status": "pending",
    "priority": "medium"
  },
  {
    "id": "7.5-test",
    "content": "7.5-TEST: Verify animations - Keyframes work on modals/panels, reduced motion media query disables animations",
    "status": "pending",
    "priority": "medium"
  },
  {
    "id": "7.6",
    "content": "7.6 Create Skeleton.tsx & Skeleton.module.css - Loading skeleton components (Skeleton, TaskCardSkeleton, ColumnSkeleton, BoardSkeleton)",
    "status": "pending",
    "priority": "medium"
  },
  {
    "id": "7.6-test",
    "content": "7.6-TEST: Verify skeletons - BoardSkeleton displays during loading, animation pulses correctly, matches card/column dimensions",
    "status": "pending",
    "priority": "medium"
  },
  {
    "id": "7.7",
    "content": "7.7 Create focus.css - Enhanced focus styles for accessibility, selection states, skip link, high contrast mode support",
    "status": "pending",
    "priority": "medium"
  },
  {
    "id": "7.7-test",
    "content": "7.7-TEST: Verify focus styles - Tab navigation shows focus rings, skip link appears on focus, selection states visible",
    "status": "pending",
    "priority": "medium"
  },
  {
    "id": "7.8",
    "content": "7.8 Create drag-drop.css - Drag visual feedback (shadow, rotation, cursor states), drop indicators",
    "status": "pending",
    "priority": "medium"
  },
  {
    "id": "7.8-test",
    "content": "7.8-TEST: Verify drag-drop styles - Dragged card has shadow/rotation, drop zones highlight, cursor changes appropriately",
    "status": "pending",
    "priority": "medium"
  },
  {
    "id": "7.9",
    "content": "7.9 Install @tanstack/react-virtual and create VirtualizedColumn.tsx - Virtualize columns with >50 tasks for performance",
    "status": "pending",
    "priority": "medium"
  },
  {
    "id": "7.9-test",
    "content": "7.9-TEST: Verify virtualization - Create 100+ tasks in column, scrolling smooth, only visible items rendered in DOM",
    "status": "pending",
    "priority": "medium"
  },
  {
    "id": "7.10",
    "content": "7.10 Update App.tsx - Wrap with ThemeProvider, add Cmd+, keyboard shortcut for settings panel",
    "status": "pending",
    "priority": "high"
  },
  {
    "id": "7.10-test",
    "content": "7.10-TEST: Verify App.tsx changes - ThemeProvider works, Cmd+, opens settings, no breaking changes to existing functionality",
    "status": "pending",
    "priority": "high"
  },
  {
    "id": "7.11",
    "content": "7.11 Update main.tsx - Change import from tokens.css to index.css",
    "status": "completed",
    "priority": "high"
  },
  {
    "id": "7.12",
    "content": "7.12 Update Board.tsx - Add skeleton loading state, use VirtualizedColumn for columns with >50 tasks",
    "status": "pending",
    "priority": "medium"
  },
  {
    "id": "7.12-test",
    "content": "7.12-TEST: Verify Board.tsx changes - Skeleton shows during load, virtualization kicks in at 50+ tasks",
    "status": "pending",
    "priority": "medium"
  },
  {
    "id": "7.13",
    "content": "7.13 Update TaskCard.tsx - Add isDragOverlay prop, memoization with React.memo, drag state classes",
    "status": "pending",
    "priority": "medium"
  },
  {
    "id": "7.13-test",
    "content": "7.13-TEST: Verify TaskCard.tsx changes - Drag overlay renders correctly, memo prevents unnecessary re-renders",
    "status": "pending",
    "priority": "medium"
  },
  {
    "id": "7.14",
    "content": "7.14 Update import paths - Fix all imports after moving context files to contexts/ directory",
    "status": "completed",
    "priority": "high"
  },
  {
    "id": "7.14-test",
    "content": "7.14-TEST: Verify import paths - No import errors, app compiles and runs, all features still work",
    "status": "completed",
    "priority": "high"
  },
  {
    "id": "7.15",
    "content": "7.15 FINAL TEST: Full regression test - Theme persistence, animations, keyboard navigation, skeleton loading, virtualization, drag-drop",
    "status": "pending",
    "priority": "high"
  }
]
```

---

## Key Decisions Made

1. **Contexts Directory**: Created `ui/src/contexts/` directory (industry standard) and moved existing context files there
2. **CSS Organization**: Split into logical files with `index.css` entry point
3. **Virtualization**: Will use `@tanstack/react-virtual` for columns with >50 tasks
4. **Testing**: Use `ui-test-engineer` agent after each implementation task (see detailed instructions below)
5. **Deferred**: Toast notifications, mobile/responsive layouts

---

## UI Testing Instructions

**IMPORTANT**: All `-test` tasks MUST use the `ui-test-engineer` agent for visual/functional verification, not just unit tests. The ui-test-engineer agent can interact with the running application in a real browser to verify UI behavior.

### Step-by-Step: How to Run UI Tests

#### Step 1: Start the Dev Server FIRST

**CRITICAL**: The dev server MUST be running and verified BEFORE calling ui-test-engineer. The agent cannot start servers itself.

```bash
# Start the dev server and verify it's running
cd /home/ramtinj/personal-workspace/EgenSkriven/ui && npm run dev &
echo "Server starting..."
sleep 5
netstat -tlnp 2>/dev/null | grep 5173 || ss -tlnp | grep 5173
```

You should see output like:
```
VITE v7.3.0  ready in 160 ms
  ➜  Local:   http://localhost:5173/
```

And the port check should show:
```
LISTEN ... :5173 ...
```

If the server is already running, you can verify with:
```bash
curl -s -o /dev/null -w "%{http_code}" http://localhost:5173
# Should return: 200
```

#### Step 2: Call ui-test-engineer with Detailed Prompt

Use the `task` tool with `subagent_type: ui-test-engineer`. The prompt MUST include:
1. **App URL**: Always `http://localhost:5173`
2. **Test Steps**: Numbered, specific actions to perform
3. **Expected Behavior**: What should happen at each step
4. **Required Output**: Specify the exact format you want back

#### Step 3: Review the Test Report

The ui-test-engineer will return a detailed report. Check:
- All steps passed
- No console errors
- Overall verdict is PASS

If FAIL, fix the issues and re-run the test.

### Required Output from ui-test-engineer

**IMPORTANT**: Always specify what output you want back from the ui-test-engineer agent. Include this at the end of your test prompt:

```
**Required Output**:
Return a detailed test report including:
1. PASS/FAIL status for each test step
2. Any errors or issues found (with details)
3. Console errors observed (if any)
4. Overall verdict: PASS or FAIL
5. Recommendations for fixes if FAIL
```

### Complete Example: Running a UI Test

Here's the full workflow for testing (example from 7.1-test):

**1. First, start and verify the dev server:**
```bash
cd /home/ramtinj/personal-workspace/EgenSkriven/ui && npm run dev &
sleep 5
ss -tlnp | grep 5173
```

**2. Then call ui-test-engineer with this prompt:**

```
Test the light mode CSS implementation for the EgenSkriven app.

**App URL**: http://localhost:5173

**Test Steps**:
1. Navigate to the app at http://localhost:5173
2. Observe the initial dark mode state - the app should have dark backgrounds
3. Using Playwright, execute JavaScript to add data-theme="light" attribute:
   document.documentElement.setAttribute('data-theme', 'light')
4. Verify the following color changes occur:
   - App background should become white/light (#FFFFFF)
   - Sidebar should become light gray (#FAFAFA)
   - Text should become dark colored (#171717)
   - Cards should have white backgrounds
5. Execute JavaScript to remove the attribute:
   document.documentElement.removeAttribute('data-theme')
6. Verify dark mode colors return

**Expected Behavior**:
- Light mode: White/light backgrounds, dark text
- Dark mode: Dark backgrounds (#0D0D0D), light text (#F5F5F5)
- Instant transition with no delay

**Required Output**:
Return a detailed test report with the following format:

## Test Report: [Feature Name]

### Step 1: [Step description]
- Status: PASS/FAIL
- Observation: [what you saw]

### Step 2: [Step description]
- Status: PASS/FAIL
- Observation: [what you saw]

[...continue for all steps...]

### Console Errors
- [List any console errors, or "None"]

### Overall Verdict
- **PASS** or **FAIL**
- Summary: [brief summary]

### Recommendations
- [Any fixes needed, or "None - all tests passed"]
```

**3. The ui-test-engineer will return a detailed report.**

---

## Reference Documents

- `docs/phase-7.md` - Full phase 7 specification with code examples for each task
- `docs/ui-design.md` - Design specifications (colors, typography, spacing, animations)

---

## How to Start New Session

1. **Read this file** (`docs/context.md`) - you're doing this now!
2. **Recreate the todo list** using the JSON above with the `todowrite` tool
3. **Check the "NEXT TASK TO DO" section** above - currently **Task 7.2**
4. **Read `docs/phase-7.md`** for the implementation code for the next task
5. **Start the dev server** before running any ui-test-engineer tests
6. **After each task**: commit with conventional commit format (max 70 chars)
7. **After each task**: run the corresponding `-test` task with ui-test-engineer

### Git Commits Made This Session

```
refactor(ui): split CSS into modular files with index entry point
refactor(ui): move context files to dedicated contexts/ directory
feat(ui): add light mode theme CSS with color token overrides
docs: add ui-test-engineer instructions to context.md
docs: add detailed ui-test-engineer workflow instructions
```

### Branch

Working on branch: `implement-phase-7`
