# Phase 7: Polish - Implementation Context

This document contains all context needed to continue Phase 7 implementation in a new session.

## Overview

Phase 7 transforms the functional UI into a polished, professional application. This phase focuses on:
- **Theming**: Light mode, dark mode, system preference detection, and accent colors
- **Animations**: Smooth transitions that enhance usability
- **Quality of Life**: Loading states (skeletons) and visual feedback
- **Performance**: Virtualization for large datasets (>50 tasks per column)

**Note**: Mobile/responsive design and toast notifications are deferred to later phases. This is a desktop-first product.

## Key Decisions Made

1. **Contexts Directory**: Create `ui/src/contexts/` directory (industry standard) and move existing context files there
2. **CSS Organization**: Split into logical files with `index.css` entry point:
   - `tokens.css` - CSS variables only
   - `base.css` - Reset, body, typography
   - `theme-light.css` - Light mode overrides
   - `animations.css` - Keyframes + utilities
   - `focus.css` - Accessibility styles
   - `drag-drop.css` - Drag feedback styles
3. **Virtualization**: Use `@tanstack/react-virtual` for columns with >50 tasks
4. **Testing**: Use `ui-test-engineer` agent after each implementation task
5. **Deferred**: Toast notifications, mobile/responsive layouts

## Current Codebase State

### Existing Structure
```
ui/src/
├── styles/
│   └── tokens.css          # Global design tokens + base reset (needs splitting)
├── hooks/
│   ├── SelectionProvider.tsx  # Needs to move to contexts/
│   ├── selectionContext.ts    # Needs to move to contexts/
│   ├── useCurrentBoard.tsx    # Needs to move to contexts/
│   ├── useKeyboard.ts
│   ├── useTasks.ts
│   └── ...
├── components/
│   ├── Board.tsx           # Has basic dnd-kit drag-drop
│   ├── TaskCard.tsx        # Needs memo, drag overlay prop
│   ├── Column.tsx
│   └── ... (all use .module.css)
├── App.tsx
└── main.tsx               # Imports tokens.css
```

### What Exists
- Design tokens in `tokens.css` (colors, typography, spacing, animation timing)
- CSS Modules architecture for all components
- dnd-kit packages installed (`@dnd-kit/core`, `@dnd-kit/sortable`, `@dnd-kit/utilities`)
- Basic keyboard navigation (`useKeyboard.ts`)
- Selection context for multi-select
- Basic transitions on cards, buttons, overlays
- Basic focus styles (`:focus-visible`)

### What Needs Creation
- `ui/src/contexts/` directory
- `ui/src/styles/index.css` (entry point)
- `ui/src/styles/base.css` (extracted from tokens.css)
- `ui/src/styles/theme-light.css`
- `ui/src/styles/animations.css`
- `ui/src/styles/focus.css`
- `ui/src/styles/drag-drop.css`
- `ui/src/contexts/ThemeContext.tsx`
- `ui/src/hooks/useAccentColor.ts`
- `ui/src/components/Settings.tsx` + `Settings.module.css`
- `ui/src/components/Skeleton.tsx` + `Skeleton.module.css`
- `ui/src/components/VirtualizedColumn.tsx`

## File Structure After Phase 7

```
ui/src/
├── styles/
│   ├── index.css           # Entry point (imports all)
│   ├── tokens.css          # CSS variables only
│   ├── base.css            # Reset, body, typography
│   ├── theme-light.css     # Light mode overrides
│   ├── animations.css      # Keyframes + utilities
│   ├── focus.css           # Accessibility styles
│   └── drag-drop.css       # Drag feedback styles
├── contexts/
│   ├── ThemeContext.tsx    # Theme state + provider
│   ├── SelectionContext.tsx # Moved from hooks/
│   └── BoardContext.tsx    # Moved from hooks/
├── hooks/
│   ├── useAccentColor.ts   # Accent color hook
│   ├── useTheme.ts         # Re-export from context
│   └── ...existing hooks
└── components/
    ├── Settings.tsx        # Settings panel
    ├── Settings.module.css
    ├── Skeleton.tsx        # Loading skeletons
    ├── Skeleton.module.css
    ├── VirtualizedColumn.tsx # For large lists
    └── ...existing components
```

## Todo List

Copy this JSON to recreate the todo list using the `todowrite` tool:

```json
[
  {
    "id": "7.0",
    "content": "7.0 PREP: Refactor CSS - Split tokens.css into tokens.css (variables only) + base.css (reset/base styles), create index.css entry point",
    "status": "pending",
    "priority": "high"
  },
  {
    "id": "7.0-test",
    "content": "7.0-TEST: Verify CSS refactor - App still renders correctly, all existing styles work, no visual regressions",
    "status": "pending",
    "priority": "high"
  },
  {
    "id": "7.0b",
    "content": "7.0b PREP: Create ui/src/contexts/ directory and move existing context files (SelectionProvider, selectionContext, useCurrentBoard) there",
    "status": "pending",
    "priority": "high"
  },
  {
    "id": "7.0b-test",
    "content": "7.0b-TEST: Verify context migration - Selection still works, board switching works, no runtime errors",
    "status": "pending",
    "priority": "high"
  },
  {
    "id": "7.1",
    "content": "7.1 Create theme-light.css - Light mode color token overrides with [data-theme=\"light\"] selector",
    "status": "pending",
    "priority": "high"
  },
  {
    "id": "7.1-test",
    "content": "7.1-TEST: Verify light mode CSS - Manually set data-theme=\"light\" on html element, verify all colors change correctly",
    "status": "pending",
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
    "status": "pending",
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
    "status": "pending",
    "priority": "high"
  },
  {
    "id": "7.14-test",
    "content": "7.14-TEST: Verify import paths - No import errors, app compiles and runs, all features still work",
    "status": "pending",
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

## Todo List Summary Table

| ID | Task | Priority |
|----|------|----------|
| **Prep** |||
| 7.0 | Refactor CSS structure (tokens.css + base.css + index.css) | High |
| 7.0-test | Verify CSS refactor - no visual regressions | High |
| 7.0b | Create contexts/ directory, move existing contexts | High |
| 7.0b-test | Verify context migration works | High |
| **Theming** |||
| 7.1 | Create theme-light.css | High |
| 7.1-test | Verify light mode colors | High |
| 7.2 | Create ThemeContext.tsx | High |
| 7.2-test | Verify theme switching & persistence | High |
| 7.3 | Create Settings.tsx & Settings.module.css | High |
| 7.3-test | Verify Settings panel UI & interactions | High |
| 7.4 | Create useAccentColor.ts hook | High |
| 7.4-test | Verify accent colors apply & persist | High |
| **Animations & Styles** |||
| 7.5 | Create animations.css | Medium |
| 7.5-test | Verify animations & reduced motion | Medium |
| 7.6 | Create Skeleton.tsx & Skeleton.module.css | Medium |
| 7.6-test | Verify skeleton loading states | Medium |
| 7.7 | Create focus.css | Medium |
| 7.7-test | Verify focus styles & accessibility | Medium |
| 7.8 | Create drag-drop.css | Medium |
| 7.8-test | Verify drag-drop visual feedback | Medium |
| **Performance** |||
| 7.9 | Install @tanstack/react-virtual, create VirtualizedColumn.tsx | Medium |
| 7.9-test | Verify virtualization with 100+ tasks | Medium |
| **Integration** |||
| 7.10 | Update App.tsx (ThemeProvider, Cmd+, shortcut) | High |
| 7.10-test | Verify App.tsx integration | High |
| 7.11 | Update main.tsx (import index.css) | High |
| 7.12 | Update Board.tsx (skeleton, virtualization) | Medium |
| 7.12-test | Verify Board.tsx changes | Medium |
| 7.13 | Update TaskCard.tsx (memo, drag overlay) | Medium |
| 7.13-test | Verify TaskCard.tsx changes | Medium |
| 7.14 | Fix import paths after context migration | High |
| 7.14-test | Verify imports work, no errors | High |
| **Final** |||
| 7.15 | Full regression test | High |

## UI Testing Instructions

After each implementation task, use the `ui-test-engineer` agent to verify:

```
Example prompt for 7.1-test:
"Test the light mode CSS implementation. The app is at http://localhost:5173. 
Open browser DevTools, add data-theme='light' attribute to the <html> element, 
and verify all colors change correctly - backgrounds should be light, text should be dark."
```

## Reference Documents

- `docs/phase-7.md` - Full phase 7 specification with code examples
- `docs/ui-design.md` - Design specifications (colors, typography, spacing, animations)

## How to Start New Session

1. Read this file (`docs/context.md`)
2. Read `docs/phase-7.md` for detailed implementation code
3. Use the JSON above with `todowrite` tool to recreate the todo list
4. Check which tasks are already completed by examining the codebase
5. Continue from the next pending task
