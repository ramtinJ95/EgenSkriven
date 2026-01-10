# UI Test Plan

This document outlines a comprehensive UI testing plan for EgenSkriven, organized into phases that can be executed sequentially using the `ui-test-engineer` agent.

---

## Overview

The EgenSkriven UI is a local-first kanban task manager with the following major feature areas:

- **Layout & Navigation**: Collapsible sidebar, board list, header with search
- **Theme System**: 7+ built-in themes, accent colors, system mode, custom themes
- **Task Management**: Cards, columns, detail panel, drag-and-drop
- **Filtering & Search**: Real-time search, filter builder, saved views
- **Keyboard Navigation**: Full keyboard operability with vim-style shortcuts
- **Modals**: Command palette, quick create, settings, display options

---

## Prerequisites

Before running any test phase, ensure the dev server is running:

```bash
cd /home/ramtinj/personal-workspace/EgenSkriven/ui && npm run dev &
sleep 3
curl -s -o /dev/null -w "%{http_code}" http://localhost:5173
# Should return: 200
```

**Note**: The server may use ports 5174, 5175, or 5176 if 5173 is busy.

---

## Test Phases

### Phase 0: Start Dev Server

**Status**: Prerequisite

**Actions**:
1. Start the development server
2. Verify it responds with HTTP 200
3. Note the actual port being used

---

### Phase 1: Core Layout & Navigation Testing

**Subtasks**:
- 1.1 Sidebar collapse/expand and persistence
- 1.2 Board list navigation and selection
- 1.3 Header elements (logo, search, buttons)

**ui-test-engineer Prompt**:

```
Perform Core Layout & Navigation testing at http://localhost:5173

## Test Areas

### 1.1 Sidebar Collapse/Expand
- Locate the sidebar on the left side of the screen
- Find and click the collapse/expand toggle button (chevron)
- Verify the sidebar collapses (width decreases significantly)
- Click the toggle again to expand
- Refresh the page - verify the collapse state persists (localStorage)

### 1.2 Board List Navigation
- In the sidebar, locate the board list section
- Verify boards are displayed with color dots and prefixes
- Click on a different board - verify it becomes active (highlighted)
- Verify the main content area updates to show the selected board
- Verify the active board has different styling than inactive boards

### 1.3 Header Elements
- Verify the header contains the logo/title "EgenSkriven"
- Locate the search bar in the center of the header
- Verify there are settings and help buttons in the header
- Click the settings button - verify settings panel opens
- Close settings, click help button - verify shortcuts help opens

## Required Output

Return a detailed report with:
1. PASS/FAIL status for each test area (1.1, 1.2, 1.3)
2. Specific notes on any failures
3. Console error check results
4. Screenshots or evidence where relevant
```

---

### Phase 2: Settings & Theme System Testing

**Subtasks**:
- 2.1 Settings panel open/close (keyboard, click, X button)
- 2.2 Theme switching (all 7+ themes)
- 2.3 Accent color selection and persistence
- 2.4 System theme mode preferences

**ui-test-engineer Prompt**:

```
Perform Settings & Theme System testing at http://localhost:5173

## Test Areas

### 2.1 Settings Panel Open/Close
- Press Ctrl+, (or Cmd+, on Mac) - verify settings panel opens
- Press Escape - verify panel closes
- Open settings again using the header settings button
- Click the X button - verify panel closes
- Open settings again, click outside the panel - verify it closes

### 2.2 Theme Switching
- Open settings panel
- Locate the theme selection grid
- Test switching to each theme and verify colors change:
  - Light theme - verify light background
  - Dark theme - verify dark background
  - Gruvbox Dark - verify retro colors
  - Catppuccin Mocha - verify pastel dark colors
  - Nord - verify arctic blue palette
  - Tokyo Night - verify purple/blue tones
  - Purple Dream - verify purple focus
- Refresh the page - verify selected theme persists

### 2.3 Accent Color Selection
- In settings, locate the accent color grid
- Click on different accent colors
- Verify the accent color changes visually (buttons, highlights)
- Verify the selected color shows a checkmark indicator
- Refresh page - verify accent color persists
- Click "Theme Default" - verify accent resets

### 2.4 System Theme Mode
- Select "System" as the theme mode
- Locate the system mode preferences section
- Verify you can select preferred light and dark themes
- Verify the app follows system dark/light preference

## Required Output

Return a detailed report with:
1. PASS/FAIL status for each test area
2. List of all themes tested with results
3. Accent color persistence verification
4. Console error check results
```

---

### Phase 3: Task Cards & Board View Testing

**Subtasks**:
- 3.1 Task card display (ID, title, labels, priority, type, due)
- 3.2 Column display (headers, counts, status dots)
- 3.3 Task selection and focus states
- 3.4 Board loading and empty states

**ui-test-engineer Prompt**:

```
Perform Task Cards & Board View testing at http://localhost:5173

## Test Areas

### 3.1 Task Card Display
- Locate task cards in the board columns
- For each visible task card, verify it displays:
  - Task ID (format: PREFIX-123 or short ID)
  - Title (may be truncated to 2 lines)
  - Status dot with appropriate color
  - Priority badge if set (color-coded)
  - Type badge if visible (feature, bug, chore, etc.)
  - Labels as colored chips if present
  - Due date if set (check for overdue styling)
- Verify cards in "need_input" column have special styling (pulsing dot)

### 3.2 Column Display
- Verify columns are displayed for different statuses
- Each column should have:
  - Column header with name
  - Task count badge
  - Status dot matching the column type
- Scroll within a column if there are many tasks
- Verify empty columns display appropriately

### 3.3 Task Selection and Focus
- Click on a task card - verify it shows selected state
- Verify the selected card has different styling (border/shadow)
- Click on a different task - verify selection moves
- Use keyboard (j/k or arrow keys) to navigate between tasks
- Verify focus ring appears on focused tasks

### 3.4 Board Loading and Empty States
- If possible, trigger a loading state (refresh page quickly)
- Verify skeleton loaders appear during load
- If there's an empty board/column, verify empty state message

## Required Output

Return a detailed report with:
1. PASS/FAIL status for each test area
2. List of card elements verified
3. Column structure verification
4. Selection behavior notes
5. Console error check results
```

---

### Phase 4: Task Detail Panel Testing

**Subtasks**:
- 4.1 Panel open/close (click, outside, escape, X)
- 4.2 Editable properties (status, type, priority, due, epic)
- 4.3 Description markdown editor
- 4.4 Sub-tasks, activity log, comments

**ui-test-engineer Prompt**:

```
Perform Task Detail Panel testing at http://localhost:5173

## Test Areas

### 4.1 Panel Open/Close
- Click on a task card - verify detail panel slides in from right
- Locate and click the X button - verify panel closes
- Open panel again by clicking a task
- Press Escape - verify panel closes
- Open panel again
- Click outside the panel (on the overlay) - verify panel closes

### 4.2 Editable Properties
- Open a task detail panel
- Locate and test each property:
  - Status dropdown - change status, verify it updates
  - Type dropdown - change type, verify badge updates
  - Priority dropdown - change priority, verify color changes
  - Due date picker - open calendar, select date, verify display
  - Epic picker (if available) - select/change epic
- Verify changes persist (refresh and check)

### 4.3 Description Markdown Editor
- In task detail, locate the description section
- Click to edit (or find edit button)
- Verify textarea appears with toolbar (Bold, Italic, Code, etc.)
- Type some text with markdown (e.g., **bold**, *italic*)
- Test keyboard shortcuts: Ctrl+B for bold, Ctrl+I for italic
- Save changes (Ctrl+Enter or blur)
- Verify markdown renders correctly in preview mode

### 4.4 Sub-tasks, Activity Log, Comments
- Locate sub-tasks section (if task has sub-tasks)
- Verify sub-tasks display with completion checkboxes
- Toggle a sub-task completion - verify state changes
- Find activity log section - verify history entries display
- Locate comments section
- Add a new comment using the input field
- Verify comment appears in the list

## Required Output

Return a detailed report with:
1. PASS/FAIL status for each test area
2. List of all properties tested
3. Markdown editor functionality notes
4. Comments/activity verification
5. Console error check results
```

---

### Phase 5: Drag and Drop Testing

**Subtasks**:
- 5.1 Drag initiation (8px threshold)
- 5.2 Visual feedback (opacity, overlay, column highlight)
- 5.3 Drop operation and state verification

**ui-test-engineer Prompt**:

```
Perform Drag and Drop testing at http://localhost:5173

## Test Areas

### 5.1 Drag Initiation
- Locate a task card in any column
- Click and hold on the task card
- Move the mouse slightly (more than 8px) to initiate drag
- Verify the drag operation starts
- Release without moving to a new column (should cancel)

### 5.2 Visual Feedback During Drag
- Start dragging a task card
- Verify the original card fades (opacity reduces)
- Verify a drag overlay/clone appears following the cursor
- Drag over different columns
- Verify columns highlight when hovering over them
- Verify visual effects on the drag overlay (shadow, scale, rotation)

### 5.3 Drop Operation
- Drag a task from one column to another
- Drop the task in the target column
- Verify:
  - Original card returns to normal opacity
  - Task appears in the new column
  - Task no longer appears in the original column
  - Task status updates to match the new column
- Open the task detail and verify status changed

## Required Output

Return a detailed report with:
1. PASS/FAIL status for each test area
2. Drag initiation behavior notes
3. Visual feedback descriptions
4. Drop state verification
5. Console error check results
```

---

### Phase 6: Search, Filters & Views Testing

**Subtasks**:
- 6.1 Search bar (focus with /, clear, real-time)
- 6.2 Filter builder (add, remove, match mode)
- 6.3 Filter bar and active filter pills
- 6.4 Views sidebar (save, favorite, delete)

**ui-test-engineer Prompt**:

```
Perform Search, Filters & Views testing at http://localhost:5173

## Test Areas

### 6.1 Search Bar
- Press "/" key - verify search bar focuses
- Type a search query - verify tasks filter in real-time
- Verify matching tasks remain visible, non-matching hide
- Click the clear button (X) - verify search clears
- Press Escape while focused - verify focus leaves or clears
- Verify search query appears as a pill in filter bar

### 6.2 Filter Builder
- Click the filter button in the header/filter bar
- Verify filter builder modal opens
- Add a filter:
  - Select a field (e.g., Priority)
  - Select an operator (e.g., "is")
  - Select a value (e.g., "high")
  - Click Add button
- Verify filter appears in active filters
- Add another filter
- Toggle match mode between "All" and "Any"
- Remove a filter using X button
- Click Done to close

### 6.3 Filter Bar and Pills
- With active filters, verify filter bar appears below header
- Verify each filter shows as a pill with field, operator, value
- Verify task count displays (e.g., "5 of 20 tasks")
- Click X on a filter pill - verify it removes
- Click "Clear all" - verify all filters removed

### 6.4 Views Sidebar (if available)
- Locate views section in sidebar
- With active filters, find "Save View" option
- Enter a view name and save
- Verify view appears in the list
- Click a saved view - verify filters apply
- Toggle favorite on a view
- Delete a view (with confirmation)

## Required Output

Return a detailed report with:
1. PASS/FAIL status for each test area
2. Search functionality notes
3. Filter builder usability
4. Views management verification
5. Console error check results
```

---

### Phase 7: Modals & Command Palette Testing

**Subtasks**:
- 7.1 Quick create modal (fields, validation)
- 7.2 Command palette (search, keyboard nav, execution)
- 7.3 Display options (view toggle, density, fields)
- 7.4 Board settings and new board modals

**ui-test-engineer Prompt**:

```
Perform Modals & Command Palette testing at http://localhost:5173

## Test Areas

### 7.1 Quick Create Modal
- Press "c" key or find the create button
- Verify quick create modal opens
- Test the form fields:
  - Title input (should auto-focus)
  - Description field (expandable)
  - Column/status selector
- Try to submit with empty title - verify disabled/validation
- Enter a title and submit
- Verify modal closes and task appears
- Press Escape to cancel - verify modal closes without creating

### 7.2 Command Palette
- Press Cmd+K (or Ctrl+K) - verify command palette opens
- Verify search input is focused
- Type to filter commands - verify fuzzy matching
- Use arrow keys to navigate - verify selection moves
- Press Enter on a command - verify it executes
- Press Escape - verify palette closes
- Verify sections: Actions, Navigation, Boards, Recent

### 7.3 Display Options
- Find and click the display options button (near view toggle)
- Verify dropdown/popover opens
- Test options:
  - View mode toggle (Board/List)
  - Density toggle (Compact/Comfortable)
  - Visible fields checkboxes (Priority, Labels, Due, etc.)
  - Group by dropdown
- Verify changes apply immediately
- Click outside - verify dropdown closes

### 7.4 Board Settings and New Board Modals
- In sidebar, click settings icon on a board
- Verify board settings modal opens with:
  - Name input
  - Prefix display (read-only)
  - Color picker
  - Resume mode selector
- Make a change and save
- Click "New Board" button
- Verify new board modal with name, prefix, color fields
- Test validation (prefix format, required fields)
- Create a new board - verify it appears in list

## Required Output

Return a detailed report with:
1. PASS/FAIL status for each test area
2. Modal behavior and validation notes
3. Command palette functionality
4. Display options verification
5. Console error check results
```

---

### Phase 8: Keyboard Navigation & Shortcuts Testing

**Subtasks**:
- 8.1 Global shortcuts (Cmd+K, Cmd+,, ?, Escape)
- 8.2 Task shortcuts (c, e, Enter, Space, Backspace)
- 8.3 Navigation shortcuts (j/k/h/l, arrows)
- 8.4 Focus management and tab order

**ui-test-engineer Prompt**:

```
Perform Keyboard Navigation & Shortcuts testing at http://localhost:5173

## Test Areas

### 8.1 Global Shortcuts
- Press Cmd+K (Ctrl+K on Linux) - verify command palette opens
- Press Escape - verify it closes
- Press Cmd+, (Ctrl+,) - verify settings opens
- Press Escape - verify it closes
- Press ? - verify shortcuts help opens
- Press Escape - verify it closes
- Press / - verify search bar focuses

### 8.2 Task Shortcuts
- Select a task (click on it)
- Press "c" - verify quick create modal opens (close it)
- With task selected, press Enter - verify detail panel opens
- Close panel, press Space - verify peek preview appears
- Press Enter to open full detail, then close
- Press "e" - verify task opens for editing
- Press Backspace/Delete - verify delete confirmation (cancel it)

### 8.3 Navigation Shortcuts
- Click on a task to select it
- Press "j" or ArrowDown - verify next task in column selected
- Press "k" or ArrowUp - verify previous task selected
- Press "h" - verify focus moves to previous column
- Press "l" - verify focus moves to next column
- Verify navigation wraps or stops at boundaries

### 8.4 Focus Management
- Use Tab key to navigate through the interface
- Verify focus rings appear on focusable elements
- Verify focus order is logical (left to right, top to bottom)
- Open a modal - verify focus is trapped inside
- Close modal - verify focus returns appropriately
- Verify no elements are skipped in tab order

## Required Output

Return a detailed report with:
1. PASS/FAIL status for each shortcut category
2. Complete list of shortcuts tested
3. Navigation behavior notes
4. Focus management verification
5. Console error check results
```

---

### Phase 9: List View Testing

**Subtasks**:
- 9.1 Table display and column rendering
- 9.2 Column sorting (click header, direction toggle)
- 9.3 Row selection and keyboard navigation
- 9.4 Empty and filtered states

**ui-test-engineer Prompt**:

```
Perform List View testing at http://localhost:5173

## Test Areas

### 9.1 Table Display
- Switch to List view (Cmd+B or display options)
- Verify table layout appears with columns:
  - Status (with color badge)
  - ID (display ID)
  - Title
  - Labels (up to 3 with overflow indicator)
  - Priority (color-coded)
  - Due Date (with overdue styling if applicable)
- Verify all tasks are displayed as rows
- Verify column headers are visible

### 9.2 Column Sorting
- Click on a column header (e.g., Status)
- Verify tasks sort by that column
- Click same header again - verify sort direction toggles
- Verify sort indicator (arrow) shows direction
- Test sorting on multiple columns:
  - Priority (high to low, low to high)
  - Due Date (earliest first, latest first)
  - Title (alphabetical)

### 9.3 Row Selection
- Click on a table row - verify it highlights
- Verify the selected row has different styling
- Use keyboard (arrow keys or j/k) to move selection
- Press Enter on selected row - verify detail panel opens
- Press Space - verify peek preview appears

### 9.4 Empty and Filtered States
- Apply a filter that matches no tasks
- Verify "No tasks match filters" message appears
- Clear filters - verify tasks reappear
- If board is empty, verify appropriate empty state

## Required Output

Return a detailed report with:
1. PASS/FAIL status for each test area
2. Table column verification
3. Sorting behavior notes
4. Row interaction verification
5. Console error check results
```

---

### Phase 10: Epic Management Testing

**Subtasks**:
- 10.1 Epic list display and filtering
- 10.2 Epic detail view and editing
- 10.3 Epic picker in task detail

**ui-test-engineer Prompt**:

```
Perform Epic Management testing at http://localhost:5173

## Test Areas

### 10.1 Epic List Display
- Locate epic list in the sidebar (may need to expand section)
- Verify epics display with:
  - Color indicator dot
  - Epic name
  - Task count
- Click "All Tasks" option - verify all tasks show
- Click "No Epic" option - verify only unassigned tasks show
- Click on an epic - verify only that epic's tasks show
- Verify active filter is highlighted

### 10.2 Epic Detail View
- Click the detail button on an epic (may be an icon)
- Verify epic detail view opens with:
  - Epic title (editable)
  - Epic description (editable)
  - Color bar showing epic color
  - Progress bar showing completion
  - Tasks grouped by status
- Edit the epic title - verify it updates
- Edit description - verify it saves
- Use color picker to change epic color
- Navigate back to board

### 10.3 Epic Picker in Task Detail
- Open a task's detail panel
- Locate the Epic field/picker
- Click to open epic dropdown
- Verify available epics are listed
- Select an epic - verify it assigns
- Open dropdown again - verify checkmark on selected
- Click "Clear" or "No Epic" - verify epic removed
- Use search/filter to find an epic by name

## Required Output

Return a detailed report with:
1. PASS/FAIL status for each test area
2. Epic list functionality notes
3. Epic detail editing verification
4. Epic picker behavior
5. Console error check results
```

---

### Phase 11: Integration & Error States Testing

**Subtasks**:
- 11.1 End-to-end task workflow
- 11.2 Loading and skeleton states
- 11.3 Error states and console errors
- 11.4 State persistence (localStorage)

**ui-test-engineer Prompt**:

```
Perform Integration & Error States testing at http://localhost:5173

## Test Areas

### 11.1 End-to-End Task Workflow
- Create a new task using quick create (press "c")
- Set title, description, and column
- Verify task appears in the correct column
- Click on the task to open detail panel
- Change the priority to "high"
- Add a due date
- Close detail panel
- Drag the task to a different column
- Verify status updates
- Use search to find the task by title
- Open detail panel and add a comment
- Verify comment appears

### 11.2 Loading and Skeleton States
- Refresh the page
- Observe loading states (may be brief):
  - Board skeleton loaders
  - "Loading..." text in sidebar
  - Comments loading skeleton
- Verify all loading states resolve properly

### 11.3 Error States and Console Errors
- Open browser DevTools console before testing
- Perform various actions and monitor for errors
- Check for:
  - JavaScript errors
  - Network request failures
  - React warnings/errors
  - Unhandled promise rejections
- If possible, simulate network issues:
  - Throttle network in DevTools
  - Verify graceful degradation

### 11.4 State Persistence (localStorage)
- Test persistence of:
  - Theme selection (change theme, refresh, verify)
  - Sidebar collapse state (collapse, refresh, verify)
  - Current board selection (select board, refresh, verify)
  - Display options (change view/density, refresh, verify)
- Clear localStorage and verify app still works
- Verify default values are applied

## Required Output

Return a detailed report with:
1. PASS/FAIL status for each test area
2. End-to-end workflow completion notes
3. List of any console errors found
4. Persistence verification results
5. Overall application stability assessment
```

---

## Running the Tests

To execute each phase:

1. **Start the dev server** (Phase 0)
2. **For each phase**, create a new `ui-test-engineer` session using the Task tool:

```
Task(
  description="Phase X: [Phase Name]",
  prompt="[Copy the ui-test-engineer prompt for the phase]",
  subagent_type="ui-test-engineer"
)
```

3. **Review the results** and mark subtasks as complete
4. **Proceed to the next phase** sequentially

---

## Test Results Tracking

| Phase | Status | Pass | Fail | Notes |
|-------|--------|------|------|-------|
| 0 | ✅ DONE | - | - | Dev server running on ports 5173-5176 |
| 1 | ✅ DONE | 3 | 0 | All subtasks (1.1, 1.2, 1.3) passed. Screenshots in .playwright-mcp/ |
| 2 | ✅ DONE | 4 | 0 | All subtasks passed (2.1-2.4). Themes, accents, persistence verified |
| 3 | Pending | - | - | - |
| 4 | Pending | - | - | - |
| 5 | Pending | - | - | - |
| 6 | Pending | - | - | - |
| 7 | Pending | - | - | - |
| 8 | Pending | - | - | - |
| 9 | Pending | - | - | - |
| 10 | Pending | - | - | - |
| 11 | Pending | - | - | - |

---

## Troubleshooting

### Server Issues
```bash
# Kill existing processes on common ports
lsof -ti:5173 | xargs kill -9 2>/dev/null
lsof -ti:5174 | xargs kill -9 2>/dev/null
```

### Port Detection
Check `npm run dev` output for actual port:
```
VITE v7.3.0  ready in 154 ms
  Local:   http://localhost:5176/   <-- Use this port
```

### Test Timeouts
- Ensure server is fully ready before testing
- Increase timeout for slower operations
