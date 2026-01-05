# Phase 6: Filtering & Views - Implementation Context

**Last Updated:** 2026-01-05
**Branch:** `implement-phase-6`
**Status:** Nearly Complete (~95%) - Only manual testing remains

---

## Overview

Phase 6 adds advanced filtering, search, saved views, and display options to the EgenSkriven kanban board application. This transforms it from a basic kanban board into a powerful task management system.

### Key Features Implemented
- Filter tasks by status, priority, type, labels, due dates, and more
- Real-time search across task titles and descriptions
- Save filter combinations as reusable "views"
- Toggle between board and list views
- Customize display options (density, visible fields, grouping)

---

## What Has Been Done

### 1. Dependencies (Task 6.1) - COMPLETE
- **Zustand** installed for state management
- Commit: `feat(ui): add zustand for filter state management`

### 2. Database Migration (Task 6.2) - COMPLETE
- Created `migrations/8_views.go` for the views collection
- **Note:** The phase-6.md doc incorrectly says `3_views.go` but migrations 3-7 already exist
- Fields: `name`, `board` (relation), `filters` (JSON), `display` (JSON), `is_favorite` (bool), `match_mode` (select: all/any)
- Commit: `feat(db): add views collection migration for saved filters`

### 3. Filter State Store (Task 6.3) - COMPLETE
- Created `ui/src/stores/filters.ts`
- Zustand store with localStorage persistence via `persist` middleware
- Persists: filters, matchMode, displayOptions, currentViewId
- Does NOT persist: searchQuery, debouncedSearchQuery (intentional)
- Includes helper functions: `getOperatorsForField()`, `operatorRequiresValue()`
- Commit: `feat(ui): add zustand filter store with localStorage persist`

### 4. Filter Logic Hook (Task 6.4) - COMPLETE
- Created `ui/src/hooks/useFilteredTasks.ts`
- Implements all filter matching logic:
  - `matchSelectFilter` - for column, priority, type, created_by
  - `matchLabelsFilter` - for labels (includes_any, includes_all, includes_none)
  - `matchDateFilter` - for due_date (is, before, after, is_set, is_not_set)
  - `matchRelationFilter` - for epic
  - `matchTextFilter` - for title (contains, is, is_not)
  - `matchesSearch` - searches title, description, id, labels
- Includes `useSearchDebounce` hook with 300ms debounce
- Exports `filterHelpers` for testing
- Commit: `feat(ui): add useFilteredTasks hook with debounced search`

### 5. Filter Tests (Task 6.5) - COMPLETE
- Created `ui/src/hooks/useFilteredTasks.test.ts`
- 68 tests covering all filter operations
- Tests for: select filters, labels filters, date filters, relation filters, text filters, search, AND/OR combination, edge cases
- All tests passing
- Commit: `test(ui): add comprehensive tests for filter logic`

### 6. FilterBar Component (Task 6.6) - COMPLETE
- Created `ui/src/components/FilterBar.tsx` + `.module.css`
- Displays active filters as removable "pills"
- Shows search query as a pill
- AND/OR toggle for multiple filters
- "Clear all" button
- Task count stats (e.g., "5 of 12 tasks")
- Filter button to open FilterBuilder
- Commit: `feat(ui): add FilterBar component with filter pills`

### 7. FilterBuilder Component (Task 6.7) - COMPLETE
- Created `ui/src/components/FilterBuilder.tsx` + `.module.css`
- Modal overlay for creating filters
- Dynamic field/operator/value selection
- Fetches epics from PocketBase for epic filter dropdown
- Includes hardcoded common labels (frontend, backend, bug, etc.)
- Shows existing filters with ability to remove
- AND/OR mode toggle
- Commit: `feat(ui): add FilterBuilder modal for creating filters`

### 8. SearchBar Component (Task 6.8) - COMPLETE
- Created `ui/src/components/SearchBar.tsx` + `.module.css`
- Global "/" keyboard shortcut to focus
- Escape to clear or blur
- Clear button when search has value
- Shows "/" shortcut hint (hidden when focused)
- Uses `forwardRef` for external focus control
- Commit: `feat(ui): add SearchBar component with / shortcut`

### 9. useViews Hook (Task 6.9) - COMPLETE
- Created `ui/src/hooks/useViews.ts`
- Fetches views for current board from PocketBase
- Realtime subscription for create/update/delete
- CRUD operations: createView, updateView, deleteView, toggleFavorite
- `applyView()` - loads a view into the filter store
- `saveCurrentAsView()` - saves current filters as new view
- Parses JSON fields (filters, display) from PocketBase
- Commit: `feat(ui): add useViews hook for saved filter views`

### 10. ViewsSidebar Component (Task 6.10) - COMPLETE
- Created `ui/src/components/ViewsSidebar.tsx` + `.module.css`
- Displays saved views in sidebar
- Separates favorites and regular views
- Click to apply view
- Star button to toggle favorite
- Delete button with confirmation
- "Modified" badge when current view has changes
- "Save" button appears when filters are active
- Inline form for naming new views
- Commit: `feat(ui): add ViewsSidebar component for saved views`

### 11. ListView Component (Task 6.11) - COMPLETE
- Created `ui/src/components/ListView.tsx` + `.module.css`
- Table view alternative to the kanban board
- Sortable columns (click header to sort asc/desc)
- Columns: Status, ID, Title, Labels, Priority, Due Date
- Row selection (highlight selected task)
- StatusBadge, PriorityBadge, DueDate sub-components
- Empty state when no tasks match filters
- Keyboard accessible rows (Enter/Space to select)
- Commit: `feat(ui): add ListView component with sortable columns`

### 12. DisplayOptions Component (Task 6.12) - COMPLETE
- Created `ui/src/components/DisplayOptions.tsx` + `.module.css`
- Dropdown/popover menu for display settings
- View mode toggle: Board / List (with icons)
- Density toggle: Compact / Comfortable
- Visible fields checkboxes: priority, labels, due_date, epic, type
- Group by dropdown (board mode only): Status, Priority, Type, Epic
- Keyboard shortcut hint (Cmd+B to toggle view)
- Click outside to close
- Commit: `feat(ui): add DisplayOptions menu for view settings`

### 13. App Integration (Task 6.13) - COMPLETE
- Updated `ui/src/App.tsx` to wire everything together
- Imports and uses `useFilteredTasks(tasks)` hook
- Added `useSearchDebounce()` call
- Board component now uses filtered tasks internally
- Added state for: isFilterBuilderOpen, isDisplayOptionsOpen
- Updated Header to include SearchBar and Display button
- Added FilterBar below Header in Layout
- Conditionally renders Board or ListView based on displayOptions.viewMode
- Updated Layout props to accept filter-related callbacks
- Commit: `feat(ui): integrate Phase 6 filter components into App`

### 14. Keyboard Shortcuts (Task 6.14) - COMPLETE
- `F` - Opens filter builder modal
- `Cmd+B` / `Ctrl+B` - Toggles between board and list view
- `/` - Focuses search input (implemented in SearchBar)
- `Escape` - Closes modals (FilterBuilder, DisplayOptions, etc.)
- Added to useKeyboardShortcuts in App.tsx
- Also added to command palette commands
- (Implemented as part of Task 6.13 commit)

### 15. Run Migration (Task 6.15) - COMPLETE
- Server starts and auto-runs migrations
- Views collection verified working with CRUD operations
- Tested: create view, list views, delete view
- Collection has all required fields: name, board, filters, display, is_favorite, match_mode

### 16. Run Tests (Task 6.16) - COMPLETE
- All 346 tests passing
- Fixed Layout.test.tsx and Header.test.tsx for new component structure
- Filter tests (68) all pass
- Build succeeds without errors
- Commit: `test(ui): fix Layout and Header tests for Phase 6 changes`

---

## What Remains To Be Done

### 17. Manual Testing (Task 6.17) - PENDING
- Start the server: `./egenskriven serve`
- Start the UI: `cd ui && npm run dev`
- Verify checklist from phase-6.md (lines 874-910):
  - [ ] Views collection exists with correct fields
  - [ ] Filter by status, priority, type works
  - [ ] Filter by labels (includes_any, includes_all) works
  - [ ] Multiple filters combine with AND/OR correctly
  - [ ] Clear filters shows all tasks
  - [ ] Search filters in real-time (case-insensitive)
  - [ ] Search works on title and description
  - [ ] Create new view from current filters
  - [ ] Apply saved view restores filters
  - [ ] Favorite/unfavorite views
  - [ ] Delete views
  - [ ] "Modified" indicator when view changed
  - [ ] Cmd+B toggles board/list view
  - [ ] Density changes card size
  - [ ] Toggle visible fields works
  - [ ] List view displays tasks in table
  - [ ] Column sorting works
  - [ ] Row selection works
  - [ ] F opens filter builder
  - [ ] / focuses search
  - [ ] Esc closes modals

---

## Files Created/Modified

### New Files
```
migrations/8_views.go                           # Views collection migration
ui/src/stores/filters.ts                        # Zustand filter store
ui/src/hooks/useFilteredTasks.ts               # Filter logic hook
ui/src/hooks/useFilteredTasks.test.ts          # Filter tests (68 tests)
ui/src/hooks/useViews.ts                       # Saved views hook
ui/src/components/FilterBar.tsx                # Active filters display
ui/src/components/FilterBar.module.css
ui/src/components/FilterBuilder.tsx            # Filter creation modal
ui/src/components/FilterBuilder.module.css
ui/src/components/SearchBar.tsx                # Search input
ui/src/components/SearchBar.module.css
ui/src/components/ViewsSidebar.tsx             # Saved views list
ui/src/components/ViewsSidebar.module.css
ui/src/components/ListView.tsx                 # Table view for tasks
ui/src/components/ListView.module.css
ui/src/components/DisplayOptions.tsx           # Display settings menu
ui/src/components/DisplayOptions.module.css
```

### Modified Files
```
ui/package.json                                 # Added zustand dependency
ui/src/App.tsx                                  # Full Phase 6 integration
ui/src/components/Layout.tsx                    # Added FilterBar and callbacks
ui/src/components/Header.tsx                    # Added SearchBar and Display button
ui/src/components/Header.module.css             # Updated styles for new layout
ui/src/components/Board.tsx                     # Now uses useFilteredTasks internally
ui/src/components/Layout.test.tsx               # Updated for new Layout props
ui/src/components/Header.test.tsx               # Updated for new Header structure
```

---

## Key Implementation Details

### Filter Store Structure
```typescript
interface FilterState {
  filters: Filter[]                    // Active filters
  matchMode: 'all' | 'any'            // AND/OR mode
  searchQuery: string                  // Immediate search input
  debouncedSearchQuery: string         // Debounced for filtering
  displayOptions: DisplayOptions       // View settings
  currentViewId: string | null         // Active saved view
  isModified: boolean                  // View has unsaved changes
}

interface DisplayOptions {
  viewMode: 'board' | 'list'
  density: 'compact' | 'comfortable'
  visibleFields: string[]              // ['priority', 'labels', 'due_date']
  groupBy: 'column' | 'priority' | 'type' | 'epic' | null
}
```

### Filter Types
```typescript
type FilterField = 'column' | 'priority' | 'type' | 'labels' | 'due_date' | 'epic' | 'created_by' | 'title'

type FilterOperator = 'is' | 'is_not' | 'is_any_of' | 'includes_any' | 'includes_all' | 'includes_none' | 'before' | 'after' | 'is_set' | 'is_not_set' | 'contains'

interface Filter {
  id: string
  field: FilterField
  operator: FilterOperator
  value: string | string[] | null
}
```

### localStorage Key
- `egenskriven-filters` - Persisted filter state

---

## Git Commits (in order)
1. `feat(ui): add zustand for filter state management`
2. `feat(db): add views collection migration for saved filters`
3. `feat(ui): add zustand filter store with localStorage persist`
4. `feat(ui): add useFilteredTasks hook with debounced search`
5. `test(ui): add comprehensive tests for filter logic`
6. `feat(ui): add FilterBar component with filter pills`
7. `feat(ui): add FilterBuilder modal for creating filters`
8. `feat(ui): add SearchBar component with / shortcut`
9. `feat(ui): add useViews hook for saved filter views`
10. `feat(ui): add ViewsSidebar component for saved views`
11. `feat(ui): add ListView component with sortable columns`
12. `feat(ui): add DisplayOptions menu for view settings`
13. `feat(ui): integrate Phase 6 filter components into App`
14. `test(ui): fix Layout and Header tests for Phase 6 changes`

---

## Todo List for Remaining Task

Copy this to recreate the todo list:

```
6.17 - Manual testing: Verify all checklist items from phase-6.md [PENDING, MEDIUM]
```

---

## How to Continue

1. Start the backend server:
   ```bash
   ./egenskriven serve
   ```

2. Start the UI dev server:
   ```bash
   cd ui && npm run dev
   ```

3. Open http://localhost:5173 in browser

4. Go through the manual testing checklist in Task 6.17

5. Fix any issues found during testing

6. Once all items verified, Phase 6 is complete!

---

## Reference Documents
- `docs/phase-6.md` - Full implementation spec with code examples
- `ui/src/types/task.ts` - Task type definitions (Column, Priority, TaskType, etc.)
- `ui/src/styles/tokens.css` - CSS variables for styling
