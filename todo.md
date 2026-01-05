# Task Description Field Implementation Plan

This document outlines the plan to add a description text field to tasks, similar to Linear and Trello.

---

## Overview

Add a `description` field to tasks that allows users to:
- Add a description when creating a task (optional)
- View the description in the task detail panel
- Edit the description inline in the detail panel
- See an indicator on task cards when a description exists

---

## Research Summary

### Industry Best Practices (Linear & Trello)

| Aspect | Linear | Trello | Our Approach |
|--------|--------|--------|--------------|
| Editor type | Rich text + Markdown | Rich text + Markdown | Plain text (Phase 1), Markdown (Phase 2) |
| Quick create | Title only, description in detail | Title only | Include optional description field |
| Card preview | Not shown | Badge indicator only | Show indicator icon |
| Detail view | Inline editable | Modal editable | Inline editable |
| Auto-save | Yes | Yes | Yes (on blur) |

### Current Codebase State

**Good news**: The description field already exists in the backend!

| Layer | Status | Notes |
|-------|--------|-------|
| Type definition | Exists | `ui/src/types/task.ts` - `description?: string` |
| Database schema | Exists | `migrations/1_initial.go` - 10,000 char limit |
| Backend API | Exists | Go CLI supports `--description` flag |
| Frontend API | Partial | `useTasks.ts` needs to pass description |
| UI Components | Missing | Need to add input/display fields |

---

## Implementation Plan

### Phase 1: Core Functionality (Required)

#### Task 1.0: Add Markdown rendering library
**Action**: Install `react-markdown` and `remark-gfm` for GitHub Flavored Markdown
**Changes**:
- Add dependencies to package.json
- Support for bold, italic, strikethrough, links, lists, code blocks, etc.

**Estimated effort**: Small

#### Task 1.1: Update useTasks hook
**File**: `ui/src/hooks/useTasks.ts`
**Changes**:
- Modify `createTask` function signature: `createTask(title, column, description?)`
- Add description to the taskData object sent to PocketBase
- Ensure updateTask already handles description (verify)

**Estimated effort**: Small

#### Task 1.2: Update QuickCreate component
**File**: `ui/src/components/QuickCreate.tsx`
**Changes**:
- Add `description` state variable
- Add collapsible/expandable description textarea
- Update `onCreate` prop type to include description
- Add "Add description" toggle/expansion button
- Submit description with form

**File**: `ui/src/components/QuickCreate.module.css`
**Changes**:
- Add `.textarea` styles (multi-line input)
- Add `.descriptionToggle` styles
- Add `.descriptionField` collapsible container styles

**Estimated effort**: Medium

#### Task 1.3: Update TaskDetail component
**File**: `ui/src/components/TaskDetail.tsx`
**Changes**:
- Add view/edit mode toggle for description
- View mode: Render description as Markdown using `react-markdown`
- Edit mode: Show textarea for editing raw Markdown
- Add empty state: "Click to add description" placeholder
- Add auto-save on blur (call `onUpdate` with new description)
- Add character count indicator (optional)
- Add subtle hint about Markdown support

**File**: `ui/src/components/TaskDetail.module.css`
**Changes**:
- Add `.descriptionTextarea` styles
- Add `.descriptionPlaceholder` styles for empty state
- Add `.descriptionMarkdown` styles for rendered Markdown
- Add focus/editing state styles
- Style Markdown elements (headings, lists, code blocks, links, etc.)

**Estimated effort**: Medium

#### Task 1.4: Update App.tsx
**File**: `ui/src/App.tsx`
**Changes**:
- Update `handleCreateTask` signature to accept description
- Pass description to `createTask` hook

**Estimated effort**: Small

#### Task 1.5: Add description indicator to TaskCard
**File**: `ui/src/components/TaskCard.tsx`
**Changes**:
- Add small icon/indicator when `task.description` exists
- Position in footer area near type badge
- Tooltip showing "Has description" on hover

**File**: `ui/src/components/TaskCard.module.css`
**Changes**:
- Add `.descriptionIndicator` styles

**Estimated effort**: Small

---

### Phase 2: Enhanced UX (Optional Future)

#### Task 2.1: Markdown formatting toolbar
- Add toolbar above textarea (bold, italic, code, links buttons)
- Keyboard shortcuts for formatting (Ctrl+B, Ctrl+I, etc.)

#### Task 2.2: Rich text editor
- Consider TipTap or Slate.js for WYSIWYG editing
- Add slash command menu (`/` to insert elements)

#### Task 2.3: Description preview on card
- Show first line of description truncated on TaskCard
- Expandable on hover

---

## File Changes Summary

### Required Changes (Phase 1)

| File | Type | Description |
|------|------|-------------|
| `ui/package.json` | Modify | Add react-markdown and remark-gfm |
| `ui/src/hooks/useTasks.ts` | Modify | Add description param to createTask |
| `ui/src/components/QuickCreate.tsx` | Modify | Add description textarea |
| `ui/src/components/QuickCreate.module.css` | Modify | Add textarea styles |
| `ui/src/components/TaskDetail.tsx` | Modify | Make description editable |
| `ui/src/components/TaskDetail.module.css` | Modify | Add editable styles |
| `ui/src/App.tsx` | Modify | Update handleCreateTask |
| `ui/src/components/TaskCard.tsx` | Modify | Add description indicator |
| `ui/src/components/TaskCard.module.css` | Modify | Add indicator styles |

### Test Files to Update

| File | Description |
|------|-------------|
| `ui/src/components/QuickCreate.test.tsx` | Test description input |
| Add new tests for TaskDetail description editing |

---

## UI/UX Design

### QuickCreate Modal
```
┌─────────────────────────────────────────┐
│ Create Task                          X  │
├─────────────────────────────────────────┤
│ Title                                   │
│ ┌─────────────────────────────────────┐ │
│ │ Enter task title...                 │ │
│ └─────────────────────────────────────┘ │
│                                         │
│ [+ Add description]  <- Click to expand │
│                                         │
│ ┌─────────────────────────────────────┐ │
│ │ Description (optional)              │ │ <- Expanded
│ │                                     │ │
│ └─────────────────────────────────────┘ │
│                                         │
│ Column                                  │
│ [Backlog ▼]                             │
│                                         │
│            [Cancel]  [Create Task]      │
└─────────────────────────────────────────┘
```

### TaskDetail Panel - Description Section
```
┌─────────────────────────────────────────┐
│ Description                             │
├─────────────────────────────────────────┤
│ ┌─────────────────────────────────────┐ │
│ │ Click to add description...         │ │ <- Empty state
│ └─────────────────────────────────────┘ │
│                                         │
│ OR when populated:                      │
│                                         │
│ ┌─────────────────────────────────────┐ │
│ │ This task involves implementing     │ │
│ │ the new feature for user profiles.  │ │
│ │ Key requirements:                   │ │
│ │ - Add avatar upload                 │ │
│ │ - Add bio field                     │ │
│ └─────────────────────────────────────┘ │
│                        123 / 10000 chars│
└─────────────────────────────────────────┘
```

### TaskCard - Description Indicator
```
┌─────────────────────────────────────┐
│ ● ABC-123                           │
│ Fix login button alignment          │
│                                     │
│ High    Jan 15    feature    ≡     │ <- ≡ indicates description
└─────────────────────────────────────┘
```

---

## Testing Plan

### Manual Testing
1. Create a task with description via QuickCreate
2. Create a task without description
3. Open task detail and verify description displays
4. Edit description in detail panel (add, modify, clear)
5. Verify auto-save on blur
6. Verify description indicator appears on cards with descriptions
7. Verify indicator doesn't appear on cards without descriptions

### Automated Testing
1. Unit tests for QuickCreate with description
2. Unit tests for TaskDetail description editing
3. Integration test for create task with description flow

---

## Acceptance Criteria

- [x] Can create a task with an optional description
- [x] Description displays in task detail panel
- [x] Description is editable inline in detail panel
- [x] Changes auto-save when focus leaves textarea
- [x] Task cards show indicator when description exists
- [x] Empty state shows "Click to add description" placeholder
- [x] Description supports up to 10,000 characters
- [x] All existing functionality continues to work
- [x] Markdown rendering works (bold, italic, lists, code, links, tables)

---

## Notes

- The backend (Go CLI and PocketBase) already fully supports the description field
- No database migrations needed
- Focus is entirely on frontend UI implementation
- Phase 1 includes Markdown rendering for descriptions
- Users write Markdown in a textarea, it renders as formatted text in view mode

---

*Created: Mon Jan 05 2026*
*Completed: Mon Jan 05 2026*

---

## Implementation Status: COMPLETE

All Phase 1 tasks have been implemented and tested successfully.

### Files Modified:
- `ui/package.json` - Added react-markdown and remark-gfm
- `ui/src/hooks/useTasks.ts` - Added description to createTask
- `ui/src/components/QuickCreate.tsx` - Added description textarea
- `ui/src/components/QuickCreate.module.css` - Added textarea styles
- `ui/src/components/TaskDetail.tsx` - Editable description with Markdown
- `ui/src/components/TaskDetail.module.css` - Description styles + Markdown styles
- `ui/src/components/TaskCard.tsx` - Description indicator icon
- `ui/src/components/TaskCard.module.css` - Indicator styles
- `ui/src/components/Board.tsx` - Updated props interface
- `ui/src/App.tsx` - Updated handleCreateTask signature
