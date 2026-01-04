# EgenSkriven UI Design

A Linear-inspired UI for a local-first kanban board. Dark mode first, keyboard-driven, fast.

---

## Design Principles

1. **Speed is everything** - 60fps, instant feedback, no loading spinners for local operations
2. **Keyboard-first** - Every action has a shortcut, command palette is central
3. **Dark mode primary** - Design dark-first, light mode as alternative
4. **Information density** - Show lots without clutter through hierarchy and spacing
5. **Minimal chrome** - Content is the focus, not UI decoration
6. **Consistency** - Same patterns everywhere

---

## Layout

### Three-Panel Structure

```
┌──────────────────────────────────────────────────────────────────────────┐
│  EgenSkriven                                              [Cmd+K] ⚙️     │
├────────────┬─────────────────────────────────────────────┬───────────────┤
│            │                                             │               │
│  BOARDS    │              MAIN CONTENT                   │    DETAIL     │
│  ─────     │           (Board/List View)                 │    PANEL      │
│  > Work    │                                             │   (slide-in)  │
│    Personal│                                             │               │
│    Side... │                                             │               │
│            │                                             │               │
│  ─────     │                                             │               │
│  FAVORITES │                                             │               │
│  ★ Urgent  │                                             │               │
│  ★ This w..│                                             │               │
│            │                                             │               │
│  ─────     │                                             │               │
│  VIEWS     │                                             │               │
│  All Tasks │                                             │               │
│  + New view│                                             │               │
│            │                                             │               │
├────────────┴─────────────────────────────────────────────┴───────────────┤
│  [C] Create   [F] Filter   [/] Search                    Board: Work     │
└──────────────────────────────────────────────────────────────────────────┘
```

### Panel Behavior

| Panel | Width | Behavior |
|-------|-------|----------|
| Sidebar | 240px | Collapsible (`Cmd+\`), resizable 200-300px |
| Main | Fluid | Fills remaining space |
| Detail | 400px | Slide-in from right, closable (`Esc`) |

---

## Board View

### Column Layout

```
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│ BACKLOG      (3)│ │ TODO         (5)│ │ IN PROGRESS  (2)│
├─────────────────┤ ├─────────────────┤ ├─────────────────┤
│                 │ │                 │ │                 │
│ ┌─────────────┐ │ │ ┌─────────────┐ │ │ ┌─────────────┐ │
│ │ Task card   │ │ │ │ Task card   │ │ │ │ Task card   │ │
│ └─────────────┘ │ │ └─────────────┘ │ │ └─────────────┘ │
│                 │ │                 │ │                 │
│ ┌─────────────┐ │ │ ┌─────────────┐ │ │ ┌─────────────┐ │
│ │ Task card   │ │ │ │ Task card   │ │ │ │ Task card   │ │
│ └─────────────┘ │ │ └─────────────┘ │ │ └─────────────┘ │
│                 │ │                 │ │                 │
│                 │ │ ┌─────────────┐ │ │                 │
│                 │ │ │ Task card   │ │ │                 │
│                 │ │ └─────────────┘ │ │                 │
│                 │ │                 │ │                 │
└─────────────────┘ └─────────────────┘ └─────────────────┘
```

**Column specs:**
- Fixed width: 280px
- Gap between columns: 16px
- Sticky headers when scrolling vertically
- Header shows: status name, task count
- Horizontal scroll when columns overflow viewport

### Task Card

```
┌───────────────────────────────────────┐
│ ● WRK-123                             │  ← Status dot + ID
│ Fix authentication timeout bug        │  ← Title (max 2 lines)
│                                       │
│ [Bug] [Auth]                          │  ← Labels (colored pills)
│                                       │
│ 🔴 Urgent              📅 Jan 15      │  ← Priority + Due date
└───────────────────────────────────────┘
```

**Card specs:**
- Padding: 12px
- Gap between cards: 8px
- Border radius: 6px
- Title: 14px, semibold, max 2 lines with ellipsis
- ID: 12px, muted color
- Labels: 11px pills with 4px radius

**Card states:**
- Default: `#1A1A1A` background
- Hover: `#252525` background
- Selected: `#2E2E2E` background + accent border
- Dragging: Elevated shadow, slight scale (1.02)

---

## List View

Alternative to board, toggled with `Cmd+B`:

```
┌──────────────────────────────────────────────────────────────────────────┐
│ ●  WRK-123  Fix authentication timeout bug        [Bug]    🔴 Urgent    │
├──────────────────────────────────────────────────────────────────────────┤
│ ●  WRK-124  Add dark mode toggle                  [Feature] ○ Medium    │
├──────────────────────────────────────────────────────────────────────────┤
│ ●  WRK-125  Update dependencies                   [Chore]   ○ Low       │
└──────────────────────────────────────────────────────────────────────────┘
```

**Row specs:**
- Height: 40px
- Hover background: `#1F1F1F`
- Selected: accent left border
- Columns: Status, ID, Title (fluid), Labels, Priority, Due date

---

## Task Detail Panel

Slides in from right when task is opened (`Enter` or click):

```
┌─────────────────────────────────────────┐
│ ← Back                        ⋮ Actions │
├─────────────────────────────────────────┤
│                                         │
│ Fix authentication timeout bug          │  ← Editable title
│ WRK-123                                 │  ← ID (copyable)
│                                         │
│ ┌─────────────────────────────────────┐ │
│ │ ## Approach                         │ │
│ │ Increase timeout, add retry logic   │ │
│ │                                     │ │
│ │ ## Checklist                        │ │
│ │ - [x] Reproduce issue               │ │
│ │ - [ ] Implement fix                 │ │
│ │ - [ ] Add tests                     │ │
│ │                                     │ │
│ │ ## Open Questions                   │ │
│ │ - Should we add exponential backoff?│ │
│ └─────────────────────────────────────┘ │
│                                         │
│ ─────────────────────────────────────── │
│                                         │
│ Status        ● In Progress             │
│ Priority      🔴 Urgent                 │
│ Type          Bug                       │
│ Labels        [Auth] [Backend]          │
│ Epic          Q1 Launch                 │
│ Due date      Jan 15, 2025              │
│                                         │
│ ─────────────────────────────────────── │
│                                         │
│ Activity                                │
│ Created Jan 10, 2025                    │
│ Updated 2 hours ago                     │
│                                         │
└─────────────────────────────────────────┘
```

**Panel specs:**
- Width: 400px (resizable 350-500px)
- Slide-in animation: 150ms ease
- Close: `Esc` or click outside
- Properties editable inline via click or shortcut

### Structured Description Sections

The description field supports structured markdown sections for agent workflows:

| Section | Purpose |
|---------|---------|
| `## Approach` | How the work will be implemented |
| `## Checklist` | Step-by-step tasks with checkboxes |
| `## Open Questions` | Uncertainties to resolve |
| `## Summary of Changes` | What was done (filled on completion) |
| `## Follow-up` | Related work identified during implementation |

**UI enhancements for sections:**
- Section headers rendered with slight background highlight
- Checkboxes (`- [ ]` / `- [x]`) rendered as interactive toggles
- Quick-add buttons for common sections
- Collapse/expand individual sections
- Progress indicator based on checkbox completion

---

## Command Palette

Opened with `Cmd+K`:

```
┌─────────────────────────────────────────────────────────────┐
│ 🔍  Type a command or search...                             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ ACTIONS                                                     │
│   ● Create task                                    C        │
│   ● Change status                                  S        │
│   ● Set priority                                   P        │
│   ● Add label                                      L        │
│                                                             │
│ NAVIGATION                                                  │
│   → Go to board: Work                              G B      │
│   → Go to board: Personal                          G B      │
│   → All tasks                                      G A      │
│                                                             │
│ RECENT TASKS                                                │
│   WRK-123 Fix authentication timeout bug                    │
│   WRK-120 Add user settings page                            │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Palette specs:**
- Width: 560px, centered
- Max height: 400px with scroll
- Fuzzy search matching
- Keyboard navigation: `↑/↓` to move, `Enter` to select
- Sections: Actions, Navigation, Recent Tasks, Search Results
- Backdrop: semi-transparent overlay

---

## Keyboard Shortcuts

### Global

| Shortcut | Action |
|----------|--------|
| `Cmd+K` | Command palette |
| `/` | Quick search |
| `Cmd+B` | Toggle board/list view |
| `Cmd+\` | Toggle sidebar |
| `Cmd+,` | Settings |
| `?` | Show shortcuts help |

### Task Actions

| Shortcut | Action |
|----------|--------|
| `C` | Create new task |
| `Enter` | Open selected task |
| `Space` | Peek preview |
| `E` | Edit title |
| `Backspace` | Delete task (with confirmation) |

### Task Properties (when task selected/open)

| Shortcut | Action |
|----------|--------|
| `S` | Set status |
| `P` | Set priority |
| `T` | Set type |
| `L` | Add/remove label |
| `D` | Set due date |

### Navigation

| Shortcut | Action |
|----------|--------|
| `J` / `↓` | Next task |
| `K` / `↑` | Previous task |
| `H` / `←` | Previous column |
| `L` / `→` | Next column |
| `Esc` | Close panel / deselect |
| `G then B` | Go to boards |

### Selection

| Shortcut | Action |
|----------|--------|
| `X` | Toggle select task |
| `Shift+X` | Select range |
| `Cmd+A` | Select all visible |
| `Esc` | Clear selection |

---

## Filtering

Accessed via `F` key or filter button:

```
┌─────────────────────────────────────────────────────────────┐
│ Filters                                          Clear all │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ ┌─────────────┐ ┌──────────────┐ ┌────────────────────────┐│
│ │ Status    ▼ │ │ is         ▼ │ │ In Progress          ▼ ││
│ └─────────────┘ └──────────────┘ └────────────────────────┘│
│                                                             │
│ + Add filter                                                │
│                                                             │
│ ─────────────────────────────────────────────────────────── │
│ Match: (●) All filters  ( ) Any filter                      │
└─────────────────────────────────────────────────────────────┘
```

**Filter options:**
- Status: is, is not, is any of
- Priority: is, is not, is any of
- Type: is, is not
- Labels: includes any, includes all, includes none
- Due date: before, after, is set, is not set
- Search: title contains

**Saved views:**
- Save current filter + display settings as named view
- Views appear in sidebar under "Views"
- Can favorite views (star icon)

---

## Quick Create

Pressing `C` opens inline task creation:

```
┌─────────────────────────────────────────────────────────────┐
│ + New task                                                  │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ Task title...                                           │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│ Status: Backlog ▼   Type: Feature ▼   Priority: None ▼     │
│                                                             │
│                                    [Cancel]  [Create ↵]     │
└─────────────────────────────────────────────────────────────┘
```

**Create specs:**
- Opens as modal or inline at top of current column
- Title is focused immediately
- `Enter` creates and closes
- `Cmd+Enter` creates and opens detail panel
- `Esc` cancels

---

## Multi-Board Support

### Board Switcher

In sidebar:

```
BOARDS
─────────────────
> Work            ← Active board (highlighted)
  Personal
  Side Projects
  
  + New board
```

### Board Settings

Each board has:
- Name
- ID prefix for tasks (e.g., "WRK", "PER")
- Custom columns/statuses
- Color theme (accent color)

---

## Theming

### Mode Toggle

Settings → Appearance:
- System (follows OS)
- Light
- Dark

### Accent Colors

Options:
- Blue (default) `#5E6AD2`
- Purple `#9333EA`
- Green `#22C55E`
- Orange `#F97316`
- Pink `#EC4899`
- Cyan `#06B6D4`
- Red `#EF4444`
- Yellow `#EAB308`

Accent color affects:
- Selected states
- Primary buttons
- Focus rings
- Active navigation items

### Custom Themes (Post-V1)

Theme file approach:
- Users drop a JSON file in `~/.egenskriven/themes/`
- Theme defines all color tokens (backgrounds, text, accents, status colors, etc.)
- Select theme in Settings → Appearance

Pre-packaged themes to include:
- Catppuccin (Mocha, Macchiato, Latte variants)
- Dracula
- Nord
- Gruvbox
- One Dark
- Solarized (Dark/Light)

Community themes can be shared as single JSON files.

---

## Color Tokens

### Dark Mode (Default)

```css
/* Backgrounds */
--bg-app: #0D0D0D;
--bg-sidebar: #141414;
--bg-card: #1A1A1A;
--bg-card-hover: #252525;
--bg-card-selected: #2E2E2E;
--bg-input: #1F1F1F;
--bg-overlay: rgba(0, 0, 0, 0.6);

/* Text */
--text-primary: #F5F5F5;
--text-secondary: #A0A0A0;
--text-muted: #666666;
--text-disabled: #444444;

/* Borders */
--border-subtle: #2A2A2A;
--border-default: #333333;
--border-focus: var(--accent);

/* Status colors */
--status-backlog: #6B7280;
--status-todo: #E5E5E5;
--status-in-progress: #F59E0B;
--status-review: #A855F7;
--status-done: #22C55E;
--status-canceled: #6B7280;

/* Priority colors */
--priority-urgent: #EF4444;
--priority-high: #F97316;
--priority-medium: #EAB308;
--priority-low: #6B7280;
--priority-none: #444444;

/* Type colors */
--type-bug: #EF4444;
--type-feature: #A855F7;
--type-chore: #6B7280;
```

### Light Mode

```css
/* Backgrounds */
--bg-app: #FFFFFF;
--bg-sidebar: #FAFAFA;
--bg-card: #FFFFFF;
--bg-card-hover: #F5F5F5;
--bg-card-selected: #F0F0F0;
--bg-input: #FFFFFF;
--bg-overlay: rgba(0, 0, 0, 0.4);

/* Text */
--text-primary: #171717;
--text-secondary: #525252;
--text-muted: #A3A3A3;
--text-disabled: #D4D4D4;

/* Borders */
--border-subtle: #E5E5E5;
--border-default: #D4D4D4;
```

---

## Typography

```css
/* Font family */
--font-sans: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
--font-mono: 'JetBrains Mono', 'Fira Code', monospace;

/* Font sizes */
--text-xs: 11px;
--text-sm: 12px;
--text-base: 13px;
--text-lg: 14px;
--text-xl: 16px;
--text-2xl: 20px;
--text-3xl: 24px;

/* Font weights */
--font-normal: 400;
--font-medium: 500;
--font-semibold: 600;

/* Line heights */
--leading-tight: 1.25;
--leading-normal: 1.5;
--leading-relaxed: 1.625;
```

### Usage

| Element | Size | Weight |
|---------|------|--------|
| Task title | text-lg (14px) | semibold |
| Task ID | text-sm (12px) | normal |
| Labels | text-xs (11px) | medium |
| Metadata | text-sm (12px) | normal |
| Section headers | text-sm (12px) | semibold, uppercase |
| Page titles | text-2xl (20px) | semibold |

---

## Spacing

```css
--space-1: 4px;
--space-2: 8px;
--space-3: 12px;
--space-4: 16px;
--space-5: 20px;
--space-6: 24px;
--space-8: 32px;
--space-10: 40px;
```

### Usage

| Element | Spacing |
|---------|---------|
| Card padding | space-3 (12px) |
| Card gap | space-2 (8px) |
| Column gap | space-4 (16px) |
| Section padding | space-4 (16px) |
| Modal padding | space-6 (24px) |

---

## Border Radius

```css
--radius-sm: 4px;   /* Labels, badges */
--radius-md: 6px;   /* Cards, inputs, buttons */
--radius-lg: 8px;   /* Modals, panels */
--radius-xl: 12px;  /* Large containers */
```

---

## Animations

```css
--duration-fast: 100ms;
--duration-normal: 150ms;
--duration-slow: 200ms;

--ease-default: cubic-bezier(0.4, 0, 0.2, 1);
--ease-in: cubic-bezier(0.4, 0, 1, 1);
--ease-out: cubic-bezier(0, 0, 0.2, 1);
```

### Motion Guidelines

- **Hover states**: 100ms, subtle background change
- **Panel slide-in**: 150ms, ease-out
- **Modal appear**: 150ms, fade + scale from 0.95
- **Drag feedback**: Immediate lift, smooth position updates
- **Dropdown open**: 100ms, fade + slide down

---

## Responsive Behavior

### Breakpoints

| Breakpoint | Width | Behavior |
|------------|-------|----------|
| Mobile | < 640px | Sidebar hidden, single column, bottom nav |
| Tablet | 640-1024px | Collapsible sidebar, 2-3 columns visible |
| Desktop | > 1024px | Full layout, all panels |

### Mobile Adaptations

- Sidebar becomes slide-over drawer
- Board view scrolls horizontally
- Detail panel becomes full-screen modal
- Command palette becomes bottom sheet
- Touch-friendly tap targets (44px min)

---

## Component Library

### Buttons

```
Primary:   [Create Task]     ← Accent bg, white text
Secondary: [Cancel]          ← Transparent, border
Ghost:     [Settings]        ← No border, hover bg
Danger:    [Delete]          ← Red accent
```

### Inputs

```
┌─────────────────────────────────────┐
│ Placeholder text...                 │
└─────────────────────────────────────┘

Focus: accent border ring
Error: red border + helper text below
```

### Dropdowns

```
┌─────────────────────┐
│ Selected value    ▼ │
├─────────────────────┤
│ Option 1          ✓ │
│ Option 2            │
│ Option 3            │
└─────────────────────┘
```

### Labels/Badges

```
[Bug]           ← Colored pill (red bg)
[Feature]       ← Colored pill (purple bg)
[Frontend]      ← Neutral pill (gray bg)
● In Progress   ← Status with dot
🔴 Urgent       ← Priority with icon
```

---

## Implementation Priority

### Phase 1: Core (MVP)

1. Board view with columns
2. Task cards (title, ID, priority)
3. Drag and drop between columns
4. Task detail panel
5. Quick create (`C`)
6. Basic keyboard navigation (`J/K`, `Enter`, `Esc`)
7. Dark mode

### Phase 2: Essential Features

1. Command palette (`Cmd+K`)
2. Filtering (`F`)
3. List view toggle
4. Labels
5. Light mode + theme toggle
6. Search (`/`)
7. Multi-board support
8. Board switcher in sidebar

### Phase 3: Polish

1. Saved/custom views
2. Peek preview (`Space`)
3. Full keyboard shortcuts
4. Accent color customization
5. Responsive/mobile layout
6. Animations and transitions
7. Keyboard shortcuts help modal

### Phase 4: Advanced

1. Epics/Projects grouping
2. Due dates with calendar picker
3. Sub-tasks
4. Timeline view
5. Custom themes
6. Import/export

---

*This document defines the UI design system for EgenSkriven. Implementation should follow Linear's design patterns while keeping the scope appropriate for a local-first, personal productivity tool.*
