# Phase 4: Interactive UI - Implementation Context

**Last Updated:** 2026-01-04
**Branch:** `implement-phase-4`
**Status:** In Progress (8 of 15 tasks completed)

---

## Overview

Phase 4 transforms the basic UI into a keyboard-first, Linear-inspired interface with:
- Command Palette (`Cmd+K`)
- Keyboard Navigation (`J/K/H/L`)
- Property Shortcuts (`S`, `P`, `T`, `L`)
- Real-time Updates (already implemented)
- Peek Preview (`Space`)

Reference document: `docs/phase-4.md`

---

## Completed Tasks

### 4.1 - useSelection Hook ✅
**File:** `ui/src/hooks/useSelection.tsx`
**Commit:** `feat(ui): add useSelection hook for task selection state`

Created a React context for selection state management:
- `SelectionProvider` - wraps the app to provide selection context
- `useSelection()` - hook to access selection state and actions
- Features:
  - Single selection (`selectedTaskId`)
  - Multi-selection (`multiSelectedIds` as `Set<string>`)
  - Column focus tracking (`focusedColumn`)
  - Actions: `selectTask`, `toggleMultiSelect`, `selectRange`, `selectAll`, `clearSelection`, `setFocusedColumn`
  - Helper: `isSelected(taskId)` returns boolean

**Note:** File is `.tsx` not `.ts` because it contains JSX for the Provider component.

---

### 4.2 - useKeyboard Hook ✅
**File:** `ui/src/hooks/useKeyboard.ts`
**Commit:** `feat(ui): add useKeyboard hook for keyboard shortcuts`

Created keyboard shortcut management system:
- `useKeyboardShortcuts(shortcuts)` - hook to register shortcuts
- `formatKeyCombo(combo)` - formats key combo for display (e.g., "Cmd+K")
- Features:
  - Input field detection (disables shortcuts when typing)
  - Cross-platform modifier handling (Cmd on Mac, Ctrl on Windows)
  - Conditional shortcuts via `when` callback
  - `allowInInput` option for shortcuts that should work in inputs (like Escape)

**Key Interfaces:**
```typescript
interface KeyCombo {
  key: string
  meta?: boolean    // Cmd/Ctrl
  ctrl?: boolean    // Ctrl specifically
  alt?: boolean     // Option/Alt
  shift?: boolean
}

interface ShortcutHandler {
  combo: KeyCombo
  handler: (event: KeyboardEvent) => void | boolean
  description?: string
  allowInInput?: boolean
  when?: () => boolean
}
```

---

### 4.3 - CommandPalette Component ✅
**Files:** 
- `ui/src/components/CommandPalette.tsx`
- `ui/src/components/CommandPalette.module.css`

**Commit:** `feat(ui): add CommandPalette component with fuzzy search`

Command palette with fuzzy search:
- Triggered by `Cmd+K` / `Ctrl+K`
- Fuzzy matching with scoring (prefers word boundaries, consecutive matches)
- Grouped sections: Actions, Navigation, Recent Tasks
- Keyboard navigation: `↑/↓` to move, `Enter` to execute, `Escape` to close
- Portal rendering (appears above all content)
- Smooth animations

**Key Interface:**
```typescript
interface Command {
  id: string
  label: string
  shortcut?: KeyCombo
  section: 'actions' | 'navigation' | 'recent'
  icon?: string
  action: () => void
  when?: () => boolean  // Conditional visibility
}
```

---

### 4.4 - PropertyPicker Component ✅
**Files:**
- `ui/src/components/PropertyPicker.tsx`
- `ui/src/components/PropertyPicker.module.css`

**Commit:** `feat(ui): add PropertyPicker component with anchor positioning`

Generic property picker with anchor positioning:
- Positions near the anchor element (selected task card)
- Keyboard navigation: `↑/↓`, `Enter`, `Escape`
- Filter input for searching options
- Current value highlighted with checkmark
- Pre-configured option sets exported:
  - `STATUS_OPTIONS` - backlog, todo, in_progress, review, done
  - `PRIORITY_OPTIONS` - urgent, high, medium, low
  - `TYPE_OPTIONS` - bug, feature, chore

**Key Interface:**
```typescript
interface PropertyOption<T> {
  value: T
  label: string
  icon?: ReactNode
  color?: string
}

interface PropertyPickerProps<T> {
  isOpen: boolean
  onClose: () => void
  onSelect: (value: T) => void
  options: PropertyOption<T>[]
  currentValue?: T
  title: string
  anchorElement?: HTMLElement | null  // For positioning
}
```

---

### 4.5 - ShortcutsHelp Component ✅
**Files:**
- `ui/src/components/ShortcutsHelp.tsx`
- `ui/src/components/ShortcutsHelp.module.css`

**Commit:** `feat(ui): add ShortcutsHelp modal for keyboard shortcuts`

Modal displaying all keyboard shortcuts:
- Triggered by `?` key
- 2-column grid layout
- Grouped by category: Global, Task Actions, Task Properties, Navigation, Selection
- Uses `formatKeyCombo` for consistent key display
- Close with `Escape` or click outside

**Shortcut Groups Defined:**
- Global: Cmd+K, /, Cmd+B, Cmd+\, ?
- Task Actions: C, Enter, Space, E, Backspace
- Task Properties: S, P, T, L, D
- Navigation: J, K, H, L, ↑, ↓, Escape
- Selection: X, Shift+X, Cmd+A

---

### 4.6 - PeekPreview Component ✅
**Files:**
- `ui/src/components/PeekPreview.tsx`
- `ui/src/components/PeekPreview.module.css`

**Commit:** `feat(ui): add PeekPreview component for quick task preview`

Quick task preview overlay:
- Triggered by `Space` key when task is selected
- Shows: task ID, type badge, title, status, priority, labels, description (truncated to 200 chars)
- Footer hint: "Press Enter to open full details"
- Close with `Space` again, `Escape`, or click outside
- Status/priority/type colors match design tokens

---

### 4.7 - Verify useTasks Real-time ✅
**File:** `ui/src/hooks/useTasks.ts` (existing, no changes needed)
**Status:** Verified working

The existing implementation already handles real-time subscriptions:
- Lines 59-88: Subscribes to `'*'` events
- Handles `create`, `update`, `delete` actions
- Properly cleans up subscription on unmount

No changes were needed - implementation was already complete from Phase 2.

---

### 4.8 - Update App.tsx ✅
**File:** `ui/src/App.tsx`
**Commit:** `feat(ui): integrate all Phase 4 components into App.tsx`

Major refactor to integrate all Phase 4 components:

**Structure Changes:**
- Split into `App` (wrapper with SelectionProvider) and `AppContent` (main logic)
- Replaced simple `selectedTaskId` state with `useSelection()` hook

**New State:**
```typescript
// Modal states
const [isCommandPaletteOpen, setIsCommandPaletteOpen] = useState(false)
const [isQuickCreateOpen, setIsQuickCreateOpen] = useState(false)
const [isShortcutsHelpOpen, setIsShortcutsHelpOpen] = useState(false)
const [isDetailOpen, setIsDetailOpen] = useState(false)
const [isPeekOpen, setIsPeekOpen] = useState(false)

// Property picker states
const [statusPickerOpen, setStatusPickerOpen] = useState(false)
const [priorityPickerOpen, setPriorityPickerOpen] = useState(false)
const [typePickerOpen, setTypePickerOpen] = useState(false)

// Anchor element ref for property pickers
const selectedCardRef = useRef<HTMLElement | null>(null)
```

**Keyboard Shortcuts Registered:**
- `Cmd+K` - Open command palette
- `?` - Show shortcuts help
- `Escape` - Close modals / deselect (cascading: peek → detail → selection)
- `C` - Create task
- `Enter` - Open selected task detail
- `Space` - Toggle peek preview
- `Backspace` - Delete task (with confirmation)
- `S` - Set status (when task selected)
- `P` - Set priority (when task selected)
- `T` - Set type (when task selected)
- `J` / `↓` - Navigate to next task
- `K` / `↑` - Navigate to previous task

**Commands in Palette:**
- Create task, Change status, Set priority, Set type, Delete task, Show shortcuts
- Recent tasks (last 5) for quick navigation

**Anchor Positioning:**
The `onTaskSelect` callback in Board stores a reference to the selected card element:
```typescript
onTaskSelect={(task) => {
  selectTask(task.id)
  const element = document.querySelector(`[data-task-id="${task.id}"]`)
  if (element instanceof HTMLElement) {
    selectedCardRef.current = element
  }
}}
```
This requires TaskCard to have `data-task-id` attribute (see task 4.9).

---

## Remaining Tasks

### 4.9 - Update TaskCard.tsx (PENDING)
**Priority:** Medium
**File:** `ui/src/components/TaskCard.tsx`

**What needs to be done:**
1. Add `data-task-id={task.id}` attribute to the card element for anchor positioning
2. Verify the existing `isSelected` prop styling works with the new SelectionProvider

**Current TaskCard props:**
```typescript
interface TaskCardProps {
  task: Task
  isSelected?: boolean
  onClick?: (task: Task) => void
  onSelect?: (task: Task) => void
}
```

**Required change (minimal):**
Add to the card's root div:
```tsx
<div
  className={...}
  data-task-id={task.id}  // ADD THIS
  ...
>
```

The `.selected` CSS class already exists in `TaskCard.module.css`.

---

### 4.10 - Write Tests for CommandPalette (PENDING)
**Priority:** Medium
**File to create:** `ui/src/components/CommandPalette.test.tsx`

**Test cases to cover:**
1. Renders nothing when `isOpen={false}`
2. Renders command list when `isOpen={true}`
3. Filters commands by search query (fuzzy matching)
4. Executes command on click
5. Closes on Escape key
6. Navigates with arrow keys (selection moves)
7. Executes selected command on Enter
8. Groups commands by section (Actions, Navigation, Recent)
9. Respects `when` condition for conditional commands

**Mock setup needed:**
- Mock commands array
- Mock `onClose` callback
- Use `@testing-library/react` and `@testing-library/user-event`

---

### 4.11 - Write Tests for useSelection (PENDING)
**Priority:** Medium
**File to create:** `ui/src/hooks/useSelection.test.tsx`

**Test cases to cover:**
1. Starts with no selection (`selectedTaskId` is null, `multiSelectedIds` is empty)
2. `selectTask` selects a single task
3. `selectTask` clears multi-selection when single selecting
4. `toggleMultiSelect` adds/removes tasks from multi-selection
5. `selectRange` selects a range of tasks
6. `selectAll` selects all provided task IDs
7. `clearSelection` clears both single and multi-selection
8. `isSelected` returns true for selected tasks (single or multi)
9. `setFocusedColumn` updates focused column

**Test setup:**
```typescript
const wrapper = ({ children }: { children: ReactNode }) => (
  <SelectionProvider>{children}</SelectionProvider>
)

const { result } = renderHook(() => useSelection(), { wrapper })
```

---

### 4.12 - Write Tests for PropertyPicker (PENDING)
**Priority:** Medium
**File to create:** `ui/src/components/PropertyPicker.test.tsx`

**Test cases to cover:**
1. Renders nothing when `isOpen={false}`
2. Renders options when `isOpen={true}`
3. Filters options by query input
4. Navigates with arrow keys
5. Selects option on Enter
6. Selects option on click
7. Closes on Escape
8. Shows checkmark for current value
9. Positions near anchor element (may need to mock getBoundingClientRect)

---

### 4.13 - Write Tests for ShortcutsHelp (PENDING)
**Priority:** Medium
**File to create:** `ui/src/components/ShortcutsHelp.test.tsx`

**Test cases to cover:**
1. Renders nothing when `isOpen={false}`
2. Renders all shortcut groups when open
3. Displays correct key combinations (uses `formatKeyCombo`)
4. Closes on overlay click
5. Closes on close button click
6. Contains all expected shortcut categories

---

### 4.14 - Write Tests for PeekPreview (PENDING)
**Priority:** Medium
**File to create:** `ui/src/components/PeekPreview.test.tsx`

**Test cases to cover:**
1. Renders nothing when `isOpen={false}` or `task={null}`
2. Displays task ID (truncated to 8 chars)
3. Displays task type with correct styling
4. Displays task title
5. Displays status and priority
6. Displays labels if present
7. Truncates description to 200 characters with ellipsis
8. Closes on overlay click
9. Shows footer hint about Enter key

---

### 4.15 - Final Verification (PENDING)
**Priority:** High

**Build verification:**
```bash
cd ui && npm run build   # Should complete without errors
cd ui && npm test        # All tests should pass
make build               # Full application build with embedded UI
```

**Manual verification checklist:**

**Keyboard Shortcuts:**
- [ ] `Cmd+K` / `Ctrl+K` opens command palette
- [ ] `?` opens shortcuts help modal
- [ ] `C` opens quick create
- [ ] `J` / `↓` moves to next task
- [ ] `K` / `↑` moves to previous task
- [ ] `Enter` opens selected task detail
- [ ] `Space` shows peek preview
- [ ] `Escape` closes panels / clears selection (cascading)
- [ ] `Backspace` deletes selected task (with confirmation)

**Property Pickers (with task selected):**
- [ ] `S` opens status picker
- [ ] `P` opens priority picker
- [ ] `T` opens type picker
- [ ] Arrow keys navigate options
- [ ] Enter selects option
- [ ] Filter input works
- [ ] Current value shows checkmark

**Command Palette:**
- [ ] Fuzzy search filters commands
- [ ] Sections display correctly
- [ ] Arrow keys move selection
- [ ] Enter executes command
- [ ] Escape closes palette
- [ ] Recent tasks appear in palette

**Selection State:**
- [ ] Clicking task selects it (visual highlight)
- [ ] Keyboard navigation changes selection
- [ ] Selection persists after closing detail panel

**Real-time Updates:**
```bash
# In one terminal, run UI
cd ui && npm run dev

# In another terminal, run backend
make dev

# In a third terminal, test CLI commands
./egenskriven add "Test task" --column todo      # Should appear in UI
./egenskriven move <id> in_progress              # Should move in UI
./egenskriven delete <id> --force                # Should disappear from UI
```

**Shortcuts Not Firing in Inputs:**
- [ ] Typing in quick create title doesn't trigger shortcuts
- [ ] Typing in command palette search doesn't trigger shortcuts
- [ ] Typing in property picker filter doesn't trigger shortcuts

---

## File Summary

| File | Status | Purpose |
|------|--------|---------|
| `ui/src/hooks/useSelection.tsx` | ✅ Created | Selection state context |
| `ui/src/hooks/useKeyboard.ts` | ✅ Created | Keyboard shortcut system |
| `ui/src/components/CommandPalette.tsx` | ✅ Created | Command palette UI |
| `ui/src/components/CommandPalette.module.css` | ✅ Created | Command palette styles |
| `ui/src/components/PropertyPicker.tsx` | ✅ Created | Property picker component |
| `ui/src/components/PropertyPicker.module.css` | ✅ Created | Property picker styles |
| `ui/src/components/ShortcutsHelp.tsx` | ✅ Created | Shortcuts help modal |
| `ui/src/components/ShortcutsHelp.module.css` | ✅ Created | Shortcuts help styles |
| `ui/src/components/PeekPreview.tsx` | ✅ Created | Peek preview overlay |
| `ui/src/components/PeekPreview.module.css` | ✅ Created | Peek preview styles |
| `ui/src/hooks/useTasks.ts` | ✅ Verified | Real-time already working |
| `ui/src/App.tsx` | ✅ Updated | Integrated all components |
| `ui/src/components/TaskCard.tsx` | ⏳ Pending | Needs data-task-id attribute |
| `ui/src/components/CommandPalette.test.tsx` | ⏳ Pending | Tests needed |
| `ui/src/hooks/useSelection.test.tsx` | ⏳ Pending | Tests needed |
| `ui/src/components/PropertyPicker.test.tsx` | ⏳ Pending | Tests needed |
| `ui/src/components/ShortcutsHelp.test.tsx` | ⏳ Pending | Tests needed |
| `ui/src/components/PeekPreview.test.tsx` | ⏳ Pending | Tests needed |

---

## Git Commits (in order)

1. `feat(ui): add useSelection hook for task selection state`
2. `feat(ui): add useKeyboard hook for keyboard shortcuts`
3. `feat(ui): add CommandPalette component with fuzzy search`
4. `feat(ui): add PropertyPicker component with anchor positioning`
5. `feat(ui): add ShortcutsHelp modal for keyboard shortcuts`
6. `feat(ui): add PeekPreview component for quick task preview`
7. `feat(ui): integrate all Phase 4 components into App.tsx`

---

## Todo List for Continuation

Copy this to recreate the todo list:

```
4.9  | Update TaskCard.tsx - Add data-task-id attribute for anchor positioning | pending | medium
4.10 | Write tests for CommandPalette.tsx - Fuzzy search, keyboard navigation, command execution, close behavior | pending | medium
4.11 | Write tests for useSelection.tsx - Single select, multi-select, range select, clear, isSelected helper | pending | medium
4.12 | Write tests for PropertyPicker.tsx - Option filtering, keyboard navigation, selection, anchor positioning | pending | medium
4.13 | Write tests for ShortcutsHelp.tsx - Render all shortcut groups, close on Escape/overlay click | pending | medium
4.14 | Write tests for PeekPreview.tsx - Task data display, close behavior, truncation | pending | medium
4.15 | Final verification - Run build, all tests, and manual verification of keyboard shortcuts and real-time updates | pending | high
```

---

## Technical Notes

### TypeScript Considerations
- `useSelection.tsx` uses `.tsx` extension due to JSX in SelectionProvider
- Type imports use `type` keyword for verbatimModuleSyntax compatibility
- `ReactNode` must be imported as `type ReactNode`

### Testing Stack
- Vitest with jsdom environment
- @testing-library/react + @testing-library/user-event
- @testing-library/jest-dom matchers
- Test setup in `ui/src/test/setup.ts`
- Config in `ui/vitest.config.ts`

### Design Tokens
All components use CSS custom properties from `ui/src/styles/tokens.css`:
- Colors: `--bg-*`, `--text-*`, `--border-*`, `--status-*`, `--priority-*`, `--type-*`
- Spacing: `--space-1` through `--space-10` (4px base)
- Typography: `--text-xs` through `--text-2xl`
- Radius: `--radius-sm`, `--radius-md`, `--radius-lg`
- Animations: `--duration-fast`, `--duration-normal`

### Portal Rendering
CommandPalette, PropertyPicker, ShortcutsHelp, and PeekPreview all use `createPortal` to render at document.body level with appropriate z-index layering:
- PeekPreview: z-index 999
- CommandPalette/ShortcutsHelp: z-index 1000
- PropertyPicker: z-index 1001

---

## How to Continue

1. Read this document to understand current state
2. Run `cd ui && npm run build` to verify everything compiles
3. Start with task 4.9 (TaskCard update) - it's a small change
4. Then proceed with tests (4.10-4.14)
5. Finally do verification (4.15)
6. Commit after each task with conventional commit format (max 70 chars)
