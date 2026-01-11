# Phase 8: Advanced Features

**Goal**: Epic UI, due dates, sub-tasks, activity log, import/export, and other advanced features.

**Duration Estimate**: 7-10 days

**Prerequisites**: Phase 7 complete (polished UI base), Phase 3 complete (epic CLI commands exist).

**Deliverable**: A feature-complete kanban board with epics, due dates, sub-tasks, activity tracking, and data portability.

---

## Overview

Phase 8 adds the advanced features that transform EgenSkriven from a basic kanban board into a comprehensive task management system. These features build on top of the polished UI from Phase 7 and extend the CLI capabilities from earlier phases.

### What We're Building

| Feature | Description | Complexity |
|---------|-------------|------------|
| Epic UI | Visual management of epics in sidebar and task detail | Medium |
| Due Dates | Date picker, visual indicators, overdue highlighting | Medium |
| Sub-tasks | Nested tasks with parent-child relationships | High |
| Markdown Editor | Rich text editing for descriptions | Medium |
| Activity Log | Track and display task history | Medium |
| Import/Export | Backup and restore data | Medium |
| Task Templates | Predefined task structures (stretch) | Low |
| Timeline View | Gantt-style view (stretch) | High |

### Why These Features?

- **Epics** group related tasks and provide project-level visibility
- **Due dates** enable deadline tracking and prioritization
- **Sub-tasks** allow breaking down complex work
- **Activity log** provides audit trail and context
- **Import/Export** ensures data portability and backup

---

## Environment Requirements

Before starting, ensure you have:

| Tool | Version | Check Command |
|------|---------|---------------|
| Go | 1.21+ | `go version` |
| Node.js | 18+ | `node --version` |
| npm | 9+ | `npm --version` |

All previous phases must be complete:
- Phase 7: Polished UI with themes and animations
- Phase 3: Epic CLI commands (`egenskriven epic list/add/show/delete`)

---

## Tasks

### 8.1 Add Due Date Field to Tasks

**What**: Extend the tasks collection with a due_date field.

**Why**: Due dates are essential for deadline tracking. This migration adds the field to the database.

**File**: `migrations/2_due_dates.go`

```go
package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}

		// Add due_date field
		// DateField stores dates as ISO 8601 strings (YYYY-MM-DD)
		collection.Fields.Add(&core.DateField{
			Name:     "due_date",
			Required: false,
		})

		return app.Save(collection)
	}, func(app core.App) error {
		// Rollback: remove due_date field
		collection, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}

		collection.Fields.RemoveByName("due_date")
		return app.Save(collection)
	})
}
```

**Steps**:

1. Create the migration file:
   ```bash
   touch migrations/2_due_dates.go
   ```

2. Add the migration code above.

3. Verify migration runs:
   ```bash
   ./egenskriven migrate
   ```
   
   **Expected output**:
   ```
   Applied 1 migration(s)
   ```

4. Verify field exists in admin UI at `http://localhost:8090/_/`

---

### 8.2 Add Due Date CLI Flags

**What**: Add `--due` flag to add/update commands and date filters to list command.

**Why**: CLI users and agents need to set and filter by due dates.

**File**: Update `internal/commands/add.go`

```go
// Add to flags
var taskDue string

func init() {
	addCmd.Flags().StringVar(&taskDue, "due", "", 
		"Due date (ISO 8601 format: YYYY-MM-DD, or relative: 'tomorrow', 'next week')")
}

// In RunE, before creating task:
if taskDue != "" {
	parsedDate, err := parseDate(taskDue)
	if err != nil {
		return fmt.Errorf("invalid due date: %w", err)
	}
	input.DueDate = parsedDate
}
```

**File**: `internal/commands/date_parser.go`

```go
package commands

import (
	"fmt"
	"strings"
	"time"
)

// parseDate converts user input to ISO 8601 date string.
// Supports:
// - ISO 8601: "2025-01-15"
// - Relative: "today", "tomorrow", "next week", "next month"
// - Shorthand: "jan 15", "january 15"
func parseDate(input string) (string, error) {
	input = strings.TrimSpace(strings.ToLower(input))
	now := time.Now()

	// Handle relative dates
	switch input {
	case "today":
		return now.Format("2006-01-02"), nil
	case "tomorrow":
		return now.AddDate(0, 0, 1).Format("2006-01-02"), nil
	case "next week":
		return now.AddDate(0, 0, 7).Format("2006-01-02"), nil
	case "next month":
		return now.AddDate(0, 1, 0).Format("2006-01-02"), nil
	}

	// Try ISO 8601 format (YYYY-MM-DD)
	if t, err := time.Parse("2006-01-02", input); err == nil {
		return t.Format("2006-01-02"), nil
	}

	// Try common formats
	formats := []string{
		"Jan 2",           // "Jan 15"
		"January 2",       // "January 15"
		"Jan 2, 2006",     // "Jan 15, 2025"
		"January 2, 2006", // "January 15, 2025"
		"2/1/2006",        // "1/15/2025"
		"2/1",             // "1/15"
	}

	for _, format := range formats {
		if t, err := time.Parse(format, input); err == nil {
			// For formats without year, use current year
			// If date is in past, use next year
			if t.Year() == 0 {
				t = t.AddDate(now.Year(), 0, 0)
				if t.Before(now) {
					t = t.AddDate(1, 0, 0)
				}
			}
			return t.Format("2006-01-02"), nil
		}
	}

	return "", fmt.Errorf("could not parse date: %s", input)
}
```

**File**: Update `internal/commands/list.go`

```go
// Add to flags
var listDueBefore string
var listDueAfter string
var listHasDue bool
var listNoDue bool

func init() {
	listCmd.Flags().StringVar(&listDueBefore, "due-before", "", 
		"Tasks due before date (inclusive)")
	listCmd.Flags().StringVar(&listDueAfter, "due-after", "", 
		"Tasks due after date (inclusive)")
	listCmd.Flags().BoolVar(&listHasDue, "has-due", false, 
		"Only tasks with due date set")
	listCmd.Flags().BoolVar(&listNoDue, "no-due", false, 
		"Only tasks without due date")
}

// In filter building:
if listDueBefore != "" {
	date, err := parseDate(listDueBefore)
	if err != nil {
		return err
	}
	filters = append(filters, dbx.NewExp(
		"due_date <= {:due_before}", 
		dbx.Params{"due_before": date},
	))
}

if listDueAfter != "" {
	date, err := parseDate(listDueAfter)
	if err != nil {
		return err
	}
	filters = append(filters, dbx.NewExp(
		"due_date >= {:due_after}", 
		dbx.Params{"due_after": date},
	))
}

if listHasDue {
	filters = append(filters, dbx.NewExp("due_date != ''"))
}

if listNoDue {
	filters = append(filters, dbx.NewExp("due_date = '' OR due_date IS NULL"))
}
```

**Steps**:

1. Create the date parser file:
   ```bash
   touch internal/commands/date_parser.go
   ```

2. Update add.go and list.go with the flag code.

3. Write tests for date parsing.

4. Test the CLI:
   ```bash
   # Set due date on new task
   egenskriven add "Finish report" --due "2025-01-15"
   egenskriven add "Call client" --due "tomorrow"
   
   # Filter by due date
   egenskriven list --due-before "next week"
   egenskriven list --has-due
   ```

---

### 8.3 Add Parent Field for Sub-tasks

**What**: Add `parent` relation field to tasks collection for hierarchical tasks.

**Why**: Sub-tasks allow breaking down complex work into trackable pieces.

**File**: `migrations/3_subtasks.go`

```go
package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}

		// Add parent field - self-referential relation
		// A task can have one parent task
		collection.Fields.Add(&core.RelationField{
			Name:          "parent",
			Required:      false,
			CollectionId:  collection.Id, // Self-reference
			MaxSelect:     1,
			CascadeDelete: false, // Keep orphaned sub-tasks if parent deleted
		})

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}

		collection.Fields.RemoveByName("parent")
		return app.Save(collection)
	})
}
```

**Steps**:

1. Create the migration file:
   ```bash
   touch migrations/3_subtasks.go
   ```

2. Run the migration:
   ```bash
   ./egenskriven migrate
   ```

3. Update CLI commands to support parent:
   ```go
   // In add.go
   var taskParent string
   
   func init() {
       addCmd.Flags().StringVar(&taskParent, "parent", "", 
           "Parent task ID (creates sub-task)")
   }
   
   // In RunE:
   if taskParent != "" {
       // Resolve parent task
       resolution, err := resolver.ResolveTask(app, taskParent)
       if err != nil {
           return err
       }
       if resolution.Matches != nil {
           return fmt.Errorf("ambiguous parent reference")
       }
       input.Parent = resolution.Task.Id
   }
   ```

4. Update list.go with parent filters:
   ```go
   var listHasParent bool
   var listNoParent bool
   
   func init() {
       listCmd.Flags().BoolVar(&listHasParent, "has-parent", false,
           "Only show sub-tasks")
       listCmd.Flags().BoolVar(&listNoParent, "no-parent", false,
           "Only show top-level tasks (exclude sub-tasks)")
   }
   ```

5. Update show.go to display sub-tasks:
   ```go
   // After displaying task details, fetch and show sub-tasks
   subtasks, err := app.FindAllRecords("tasks",
       dbx.NewExp("parent = {:parent}", dbx.Params{"parent": task.Id}),
   )
   if err == nil && len(subtasks) > 0 {
       fmt.Println("\nSub-tasks:")
       for _, st := range subtasks {
           fmt.Printf("  [%s] %s (%s)\n", 
               st.Id[:7], 
               st.GetString("title"),
               st.GetString("column"),
           )
       }
   }
   ```

6. Test sub-task creation:
   ```bash
   # Create parent task
   egenskriven add "Implement login"
   # Output: Created task: Implement login [abc123]
   
   # Create sub-tasks
   egenskriven add "Create login form" --parent abc123
   egenskriven add "Add validation" --parent abc123
   egenskriven add "Connect to API" --parent abc123
   
   # View parent with sub-tasks
   egenskriven show abc123
   ```

---

### 8.4 Create Date Picker Component

**What**: Build a calendar date picker for the UI.

**Why**: Users need a visual way to set due dates on tasks.

**File**: `ui/src/components/DatePicker.tsx`

```tsx
import { useState, useRef, useEffect } from 'react';

interface DatePickerProps {
  value: string | null;           // ISO date string or null
  onChange: (date: string | null) => void;
  placeholder?: string;
}

// Days of the week headers
const DAYS = ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa'];

// Month names for header
const MONTHS = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December'
];

export function DatePicker({ value, onChange, placeholder = 'Set due date' }: DatePickerProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [viewDate, setViewDate] = useState(() => {
    // Start calendar on selected date or today
    return value ? new Date(value) : new Date();
  });
  const containerRef = useRef<HTMLDivElement>(null);

  // Close picker when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // Generate calendar grid for current view month
  const generateCalendarDays = () => {
    const year = viewDate.getFullYear();
    const month = viewDate.getMonth();
    
    // First day of month and total days
    const firstDay = new Date(year, month, 1).getDay();
    const daysInMonth = new Date(year, month + 1, 0).getDate();
    
    // Previous month days to show
    const prevMonthDays = new Date(year, month, 0).getDate();
    
    const days: { date: Date; isCurrentMonth: boolean; isToday: boolean; isSelected: boolean }[] = [];
    
    // Add previous month's trailing days
    for (let i = firstDay - 1; i >= 0; i--) {
      const date = new Date(year, month - 1, prevMonthDays - i);
      days.push({
        date,
        isCurrentMonth: false,
        isToday: false,
        isSelected: false,
      });
    }
    
    // Add current month's days
    const today = new Date();
    const selectedDate = value ? new Date(value) : null;
    
    for (let i = 1; i <= daysInMonth; i++) {
      const date = new Date(year, month, i);
      days.push({
        date,
        isCurrentMonth: true,
        isToday: date.toDateString() === today.toDateString(),
        isSelected: selectedDate ? date.toDateString() === selectedDate.toDateString() : false,
      });
    }
    
    // Add next month's leading days to fill grid
    const remaining = 42 - days.length; // 6 rows * 7 days
    for (let i = 1; i <= remaining; i++) {
      const date = new Date(year, month + 1, i);
      days.push({
        date,
        isCurrentMonth: false,
        isToday: false,
        isSelected: false,
      });
    }
    
    return days;
  };

  // Handle date selection
  const handleSelectDate = (date: Date) => {
    // Format as ISO date string (YYYY-MM-DD)
    const isoDate = date.toISOString().split('T')[0];
    onChange(isoDate);
    setIsOpen(false);
  };

  // Navigate months
  const prevMonth = () => {
    setViewDate(new Date(viewDate.getFullYear(), viewDate.getMonth() - 1, 1));
  };

  const nextMonth = () => {
    setViewDate(new Date(viewDate.getFullYear(), viewDate.getMonth() + 1, 1));
  };

  // Format display value
  const displayValue = value 
    ? new Date(value).toLocaleDateString('en-US', { 
        month: 'short', 
        day: 'numeric',
        year: 'numeric'
      })
    : placeholder;

  // Check if date is overdue
  const isOverdue = value && new Date(value) < new Date(new Date().toDateString());

  return (
    <div className="date-picker" ref={containerRef}>
      {/* Trigger button */}
      <button 
        className={`date-picker-trigger ${value ? 'has-value' : ''} ${isOverdue ? 'overdue' : ''}`}
        onClick={() => setIsOpen(!isOpen)}
        type="button"
      >
        <span className="date-icon">&#128197;</span>
        <span className="date-value">{displayValue}</span>
        {value && (
          <button 
            className="date-clear"
            onClick={(e) => {
              e.stopPropagation();
              onChange(null);
            }}
            aria-label="Clear date"
          >
            &times;
          </button>
        )}
      </button>

      {/* Calendar dropdown */}
      {isOpen && (
        <div className="date-picker-dropdown">
          {/* Month/year header with navigation */}
          <div className="date-picker-header">
            <button onClick={prevMonth} className="nav-btn" type="button">&lt;</button>
            <span className="month-year">
              {MONTHS[viewDate.getMonth()]} {viewDate.getFullYear()}
            </span>
            <button onClick={nextMonth} className="nav-btn" type="button">&gt;</button>
          </div>

          {/* Day of week headers */}
          <div className="date-picker-days-header">
            {DAYS.map(day => (
              <span key={day} className="day-header">{day}</span>
            ))}
          </div>

          {/* Calendar grid */}
          <div className="date-picker-grid">
            {generateCalendarDays().map((day, index) => (
              <button
                key={index}
                className={`day-cell 
                  ${!day.isCurrentMonth ? 'other-month' : ''} 
                  ${day.isToday ? 'today' : ''} 
                  ${day.isSelected ? 'selected' : ''}`
                }
                onClick={() => handleSelectDate(day.date)}
                type="button"
              >
                {day.date.getDate()}
              </button>
            ))}
          </div>

          {/* Quick select shortcuts */}
          <div className="date-picker-shortcuts">
            <button 
              onClick={() => handleSelectDate(new Date())} 
              type="button"
            >
              Today
            </button>
            <button 
              onClick={() => {
                const tomorrow = new Date();
                tomorrow.setDate(tomorrow.getDate() + 1);
                handleSelectDate(tomorrow);
              }} 
              type="button"
            >
              Tomorrow
            </button>
            <button 
              onClick={() => {
                const nextWeek = new Date();
                nextWeek.setDate(nextWeek.getDate() + 7);
                handleSelectDate(nextWeek);
              }} 
              type="button"
            >
              Next week
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
```

**File**: `ui/src/styles/date-picker.css`

```css
/* Date Picker Component Styles */

.date-picker {
  position: relative;
  display: inline-block;
}

.date-picker-trigger {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  background: var(--bg-input);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: var(--text-sm);
  cursor: pointer;
  transition: all var(--duration-fast) var(--ease-default);
}

.date-picker-trigger:hover {
  background: var(--bg-card-hover);
  border-color: var(--border-focus);
}

.date-picker-trigger.has-value {
  color: var(--text-primary);
}

.date-picker-trigger.overdue {
  color: var(--priority-urgent);
}

.date-picker-trigger.overdue .date-icon {
  color: var(--priority-urgent);
}

.date-icon {
  font-size: var(--text-base);
}

.date-clear {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 0 var(--space-1);
  font-size: var(--text-lg);
  line-height: 1;
}

.date-clear:hover {
  color: var(--text-primary);
}

.date-picker-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  z-index: 100;
  margin-top: var(--space-1);
  padding: var(--space-3);
  background: var(--bg-card);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-lg);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  min-width: 280px;
}

.date-picker-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-3);
}

.month-year {
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.nav-btn {
  background: none;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  padding: var(--space-1) var(--space-2);
  border-radius: var(--radius-sm);
}

.nav-btn:hover {
  background: var(--bg-card-hover);
  color: var(--text-primary);
}

.date-picker-days-header {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 0;
  margin-bottom: var(--space-2);
}

.day-header {
  text-align: center;
  font-size: var(--text-xs);
  color: var(--text-muted);
  font-weight: var(--font-medium);
  padding: var(--space-1);
}

.date-picker-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 2px;
}

.day-cell {
  aspect-ratio: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: none;
  border: none;
  color: var(--text-primary);
  font-size: var(--text-sm);
  cursor: pointer;
  border-radius: var(--radius-sm);
  transition: background var(--duration-fast);
}

.day-cell:hover {
  background: var(--bg-card-hover);
}

.day-cell.other-month {
  color: var(--text-muted);
}

.day-cell.today {
  font-weight: var(--font-semibold);
  color: var(--accent);
}

.day-cell.selected {
  background: var(--accent);
  color: white;
}

.day-cell.selected:hover {
  background: var(--accent);
}

.date-picker-shortcuts {
  display: flex;
  gap: var(--space-2);
  margin-top: var(--space-3);
  padding-top: var(--space-3);
  border-top: 1px solid var(--border-subtle);
}

.date-picker-shortcuts button {
  flex: 1;
  padding: var(--space-2);
  background: var(--bg-input);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  font-size: var(--text-xs);
  cursor: pointer;
}

.date-picker-shortcuts button:hover {
  background: var(--bg-card-hover);
  color: var(--text-primary);
}
```

**Steps**:

1. Create the component file:
   ```bash
   touch ui/src/components/DatePicker.tsx
   ```

2. Create the styles file:
   ```bash
   touch ui/src/styles/date-picker.css
   ```

3. Import styles in your main CSS or App.tsx.

4. Write component tests:
   ```typescript
   // ui/src/components/DatePicker.test.tsx
   import { render, screen, fireEvent } from '@testing-library/react';
   import { DatePicker } from './DatePicker';
   
   describe('DatePicker', () => {
     it('renders with placeholder when no value', () => {
       render(<DatePicker value={null} onChange={() => {}} />);
       expect(screen.getByText('Set due date')).toBeInTheDocument();
     });
     
     it('opens calendar on click', () => {
       render(<DatePicker value={null} onChange={() => {}} />);
       fireEvent.click(screen.getByRole('button'));
       expect(screen.getByText('Today')).toBeInTheDocument();
     });
     
     it('calls onChange when date selected', () => {
       const handleChange = vi.fn();
       render(<DatePicker value={null} onChange={handleChange} />);
       fireEvent.click(screen.getByRole('button'));
       fireEvent.click(screen.getByText('Today'));
       expect(handleChange).toHaveBeenCalled();
     });
   });
   ```

---

### 8.5 Create Epic Picker Component

**What**: Build a component to select/change epic for a task.

**Why**: Users need to visually assign tasks to epics from the task detail panel.

**File**: `ui/src/components/EpicPicker.tsx`

```tsx
import { useState, useEffect } from 'react';
import { useEpics } from '../hooks/usePocketBase';

interface EpicPickerProps {
  value: string | null;          // Epic ID or null
  onChange: (epicId: string | null) => void;
  onClose?: () => void;
}

export function EpicPicker({ value, onChange, onClose }: EpicPickerProps) {
  const { epics, loading } = useEpics();
  const [search, setSearch] = useState('');

  // Filter epics by search term
  const filteredEpics = epics.filter(epic =>
    epic.title.toLowerCase().includes(search.toLowerCase())
  );

  // Get currently selected epic
  const selectedEpic = value ? epics.find(e => e.id === value) : null;

  // Handle keyboard navigation
  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === 'Escape') {
        onClose?.();
      }
    }
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [onClose]);

  if (loading) {
    return <div className="epic-picker loading">Loading epics...</div>;
  }

  return (
    <div className="epic-picker">
      {/* Search input */}
      <div className="epic-picker-search">
        <input
          type="text"
          placeholder="Search epics..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          autoFocus
        />
      </div>

      {/* Epic list */}
      <div className="epic-picker-list">
        {/* "No epic" option */}
        <button
          className={`epic-option ${!value ? 'selected' : ''}`}
          onClick={() => {
            onChange(null);
            onClose?.();
          }}
        >
          <span className="epic-color" style={{ background: '#666' }} />
          <span className="epic-name">No epic</span>
          {!value && <span className="check-mark">&#10003;</span>}
        </button>

        {/* Epic options */}
        {filteredEpics.map(epic => (
          <button
            key={epic.id}
            className={`epic-option ${value === epic.id ? 'selected' : ''}`}
            onClick={() => {
              onChange(epic.id);
              onClose?.();
            }}
          >
            <span 
              className="epic-color" 
              style={{ background: epic.color || '#5E6AD2' }} 
            />
            <span className="epic-name">{epic.title}</span>
            {value === epic.id && <span className="check-mark">&#10003;</span>}
          </button>
        ))}

        {/* Empty state */}
        {filteredEpics.length === 0 && search && (
          <div className="epic-empty">
            No epics matching "{search}"
          </div>
        )}

        {epics.length === 0 && !search && (
          <div className="epic-empty">
            No epics created yet.
            <br />
            <span className="hint">Create one with: egenskriven epic add "Epic name"</span>
          </div>
        )}
      </div>
    </div>
  );
}
```

**File**: `ui/src/styles/epic-picker.css`

```css
/* Epic Picker Component Styles */

.epic-picker {
  background: var(--bg-card);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-lg);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  min-width: 240px;
  max-height: 300px;
  display: flex;
  flex-direction: column;
}

.epic-picker-search {
  padding: var(--space-2);
  border-bottom: 1px solid var(--border-subtle);
}

.epic-picker-search input {
  width: 100%;
  padding: var(--space-2);
  background: var(--bg-input);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  font-size: var(--text-sm);
}

.epic-picker-search input:focus {
  outline: none;
  border-color: var(--accent);
}

.epic-picker-search input::placeholder {
  color: var(--text-muted);
}

.epic-picker-list {
  overflow-y: auto;
  padding: var(--space-1);
}

.epic-option {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  width: 100%;
  padding: var(--space-2) var(--space-3);
  background: none;
  border: none;
  color: var(--text-primary);
  font-size: var(--text-sm);
  cursor: pointer;
  border-radius: var(--radius-sm);
  text-align: left;
}

.epic-option:hover {
  background: var(--bg-card-hover);
}

.epic-option.selected {
  background: var(--bg-card-selected);
}

.epic-color {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  flex-shrink: 0;
}

.epic-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.check-mark {
  color: var(--accent);
  font-size: var(--text-sm);
}

.epic-empty {
  padding: var(--space-4);
  text-align: center;
  color: var(--text-muted);
  font-size: var(--text-sm);
}

.epic-empty .hint {
  font-size: var(--text-xs);
  margin-top: var(--space-2);
  display: block;
  font-family: var(--font-mono);
}
```

**Steps**:

1. Create the component and styles files.

2. Add `useEpics` hook to `usePocketBase.ts`:
   ```typescript
   export function useEpics() {
     const [epics, setEpics] = useState<Epic[]>([]);
     const [loading, setLoading] = useState(true);
   
     useEffect(() => {
       pb.collection('epics')
         .getFullList<Epic>({ sort: 'title' })
         .then(setEpics)
         .finally(() => setLoading(false));
   
       // Subscribe to real-time updates
       pb.collection('epics').subscribe<Epic>('*', (e) => {
         if (e.action === 'create') {
           setEpics(prev => [...prev, e.record]);
         } else if (e.action === 'update') {
           setEpics(prev => prev.map(ep => ep.id === e.record.id ? e.record : ep));
         } else if (e.action === 'delete') {
           setEpics(prev => prev.filter(ep => ep.id !== e.record.id));
         }
       });
   
       return () => {
         pb.collection('epics').unsubscribe('*');
       };
     }, []);
   
     return { epics, loading };
   }
   ```

3. Integrate into TaskDetail panel.

---

### 8.6 Create Epic List Sidebar Component

**What**: Add epic listing to the sidebar with task counts.

**Why**: Users need quick access to view all tasks under an epic.

**File**: `ui/src/components/EpicList.tsx`

```tsx
import { useEpics, useTasks } from '../hooks/usePocketBase';

interface EpicListProps {
  selectedEpicId: string | null;
  onSelectEpic: (epicId: string | null) => void;
}

export function EpicList({ selectedEpicId, onSelectEpic }: EpicListProps) {
  const { epics, loading } = useEpics();
  const { tasks } = useTasks();

  // Count tasks per epic
  const taskCountByEpic = tasks.reduce((acc, task) => {
    if (task.epic) {
      acc[task.epic] = (acc[task.epic] || 0) + 1;
    }
    return acc;
  }, {} as Record<string, number>);

  if (loading) {
    return <div className="epic-list-loading">Loading...</div>;
  }

  if (epics.length === 0) {
    return null; // Don't show section if no epics
  }

  return (
    <div className="epic-list">
      <div className="sidebar-section-header">
        <span>Epics</span>
      </div>
      
      <div className="epic-list-items">
        {epics.map(epic => (
          <button
            key={epic.id}
            className={`epic-list-item ${selectedEpicId === epic.id ? 'active' : ''}`}
            onClick={() => onSelectEpic(selectedEpicId === epic.id ? null : epic.id)}
          >
            <span 
              className="epic-indicator" 
              style={{ background: epic.color || '#5E6AD2' }} 
            />
            <span className="epic-title">{epic.title}</span>
            <span className="epic-count">{taskCountByEpic[epic.id] || 0}</span>
          </button>
        ))}
      </div>
    </div>
  );
}
```

**File**: `ui/src/styles/epic-list.css`

```css
/* Epic List Sidebar Styles */

.epic-list {
  margin-top: var(--space-4);
}

.epic-list-items {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.epic-list-item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  background: none;
  border: none;
  color: var(--text-secondary);
  font-size: var(--text-sm);
  cursor: pointer;
  border-radius: var(--radius-sm);
  text-align: left;
  width: 100%;
  transition: all var(--duration-fast);
}

.epic-list-item:hover {
  background: var(--bg-card-hover);
  color: var(--text-primary);
}

.epic-list-item.active {
  background: var(--bg-card-selected);
  color: var(--text-primary);
}

.epic-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.epic-title {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.epic-count {
  color: var(--text-muted);
  font-size: var(--text-xs);
  background: var(--bg-input);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
}
```

**Steps**:

1. Create the component and styles.

2. Add to Sidebar.tsx:
   ```tsx
   // In Sidebar component
   const [selectedEpicId, setSelectedEpicId] = useState<string | null>(null);
   
   // Pass to EpicList
   <EpicList 
     selectedEpicId={selectedEpicId}
     onSelectEpic={(id) => {
       setSelectedEpicId(id);
       // Update filter state to filter by epic
       onFilterChange({ ...filters, epic: id });
     }}
   />
   ```

3. Test that clicking an epic filters the board.

---

### 8.7 Create Epic Detail View

**What**: Full-screen view for viewing/editing an epic and its tasks.

**Why**: Users need to see all tasks in an epic, track progress, and edit epic properties.

**File**: `ui/src/components/EpicDetail.tsx`

```tsx
import { useState } from 'react';
import { useEpic, useTasks } from '../hooks/usePocketBase';
import { TaskCard } from './TaskCard';

interface EpicDetailProps {
  epicId: string;
  onClose: () => void;
  onTaskClick: (taskId: string) => void;
}

export function EpicDetail({ epicId, onClose, onTaskClick }: EpicDetailProps) {
  const { epic, loading, updateEpic, deleteEpic } = useEpic(epicId);
  const { tasks } = useTasks();
  const [isEditing, setIsEditing] = useState(false);
  const [editTitle, setEditTitle] = useState('');
  const [editDescription, setEditDescription] = useState('');

  if (loading || !epic) {
    return <div className="epic-detail loading">Loading...</div>;
  }

  // Filter tasks belonging to this epic
  const epicTasks = tasks.filter(t => t.epic === epicId);
  
  // Calculate progress
  const completedCount = epicTasks.filter(t => t.column === 'done').length;
  const totalCount = epicTasks.length;
  const progressPercent = totalCount > 0 ? (completedCount / totalCount) * 100 : 0;

  // Group tasks by column
  const tasksByColumn = {
    backlog: epicTasks.filter(t => t.column === 'backlog'),
    todo: epicTasks.filter(t => t.column === 'todo'),
    in_progress: epicTasks.filter(t => t.column === 'in_progress'),
    review: epicTasks.filter(t => t.column === 'review'),
    done: epicTasks.filter(t => t.column === 'done'),
  };

  const handleStartEdit = () => {
    setEditTitle(epic.title);
    setEditDescription(epic.description || '');
    setIsEditing(true);
  };

  const handleSaveEdit = async () => {
    await updateEpic({
      title: editTitle,
      description: editDescription,
    });
    setIsEditing(false);
  };

  const handleDelete = async () => {
    if (confirm(`Delete epic "${epic.title}"? Tasks will remain but be unlinked.`)) {
      await deleteEpic();
      onClose();
    }
  };

  return (
    <div className="epic-detail">
      {/* Header */}
      <div className="epic-detail-header">
        <button className="back-btn" onClick={onClose}>
          &larr; Back
        </button>
        <div className="epic-actions">
          <button onClick={handleStartEdit}>Edit</button>
          <button className="danger" onClick={handleDelete}>Delete</button>
        </div>
      </div>

      {/* Epic info */}
      <div className="epic-detail-info">
        <div className="epic-color-bar" style={{ background: epic.color || '#5E6AD2' }} />
        
        {isEditing ? (
          <div className="epic-edit-form">
            <input
              type="text"
              value={editTitle}
              onChange={(e) => setEditTitle(e.target.value)}
              placeholder="Epic title"
              autoFocus
            />
            <textarea
              value={editDescription}
              onChange={(e) => setEditDescription(e.target.value)}
              placeholder="Description (optional)"
              rows={3}
            />
            <div className="edit-actions">
              <button onClick={handleSaveEdit}>Save</button>
              <button onClick={() => setIsEditing(false)}>Cancel</button>
            </div>
          </div>
        ) : (
          <>
            <h1 className="epic-title">{epic.title}</h1>
            {epic.description && (
              <p className="epic-description">{epic.description}</p>
            )}
          </>
        )}

        {/* Progress bar */}
        <div className="epic-progress">
          <div className="progress-header">
            <span>Progress</span>
            <span>{completedCount} / {totalCount} tasks</span>
          </div>
          <div className="progress-bar">
            <div 
              className="progress-fill" 
              style={{ width: `${progressPercent}%` }} 
            />
          </div>
        </div>
      </div>

      {/* Task list by status */}
      <div className="epic-tasks">
        {Object.entries(tasksByColumn).map(([column, columnTasks]) => (
          columnTasks.length > 0 && (
            <div key={column} className="epic-task-group">
              <h3 className="group-header">
                {column.replace('_', ' ')} ({columnTasks.length})
              </h3>
              <div className="group-tasks">
                {columnTasks.map(task => (
                  <TaskCard
                    key={task.id}
                    task={task}
                    onClick={() => onTaskClick(task.id)}
                    compact
                  />
                ))}
              </div>
            </div>
          )
        ))}

        {totalCount === 0 && (
          <div className="epic-empty">
            <p>No tasks in this epic yet.</p>
            <p className="hint">
              Link tasks to this epic from the task detail panel.
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
```

**Steps**:

1. Create the component.

2. Add routing/state to show epic detail when epic is clicked.

3. Add `useEpic` hook for single epic operations:
   ```typescript
   export function useEpic(epicId: string) {
     const [epic, setEpic] = useState<Epic | null>(null);
     const [loading, setLoading] = useState(true);
   
     useEffect(() => {
       pb.collection('epics')
         .getOne<Epic>(epicId)
         .then(setEpic)
         .finally(() => setLoading(false));
     }, [epicId]);
   
     const updateEpic = async (data: Partial<Epic>) => {
       const updated = await pb.collection('epics').update<Epic>(epicId, data);
       setEpic(updated);
     };
   
     const deleteEpic = async () => {
       await pb.collection('epics').delete(epicId);
     };
   
     return { epic, loading, updateEpic, deleteEpic };
   }
   ```

---

### 8.8 Add Due Date to Task Card

**What**: Display due date on task cards with overdue highlighting.

**Why**: Users need at-a-glance visibility of deadlines.

**File**: Update `ui/src/components/TaskCard.tsx`

```tsx
// Add to TaskCard component

// Helper to format due date display
function formatDueDate(dueDate: string | null): string | null {
  if (!dueDate) return null;
  
  const due = new Date(dueDate);
  const today = new Date();
  const tomorrow = new Date(today);
  tomorrow.setDate(tomorrow.getDate() + 1);
  
  // Format as "Today", "Tomorrow", or "Jan 15"
  if (due.toDateString() === today.toDateString()) {
    return 'Today';
  }
  if (due.toDateString() === tomorrow.toDateString()) {
    return 'Tomorrow';
  }
  
  return due.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
}

// Check if overdue
function isOverdue(dueDate: string | null): boolean {
  if (!dueDate) return false;
  return new Date(dueDate) < new Date(new Date().toDateString());
}

// In the TaskCard render:
{task.due_date && (
  <div className={`task-due-date ${isOverdue(task.due_date) ? 'overdue' : ''}`}>
    <span className="due-icon">&#128197;</span>
    <span>{formatDueDate(task.due_date)}</span>
  </div>
)}
```

**CSS additions**:

```css
/* Add to task card styles */

.task-due-date {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  font-size: var(--text-xs);
  color: var(--text-secondary);
}

.task-due-date.overdue {
  color: var(--priority-urgent);
}

.task-due-date .due-icon {
  font-size: var(--text-sm);
}
```

---

### 8.9 Create Sub-task List Component

**What**: Display and manage sub-tasks in the task detail panel.

**Why**: Users need to see and track sub-task completion.

**File**: `ui/src/components/SubtaskList.tsx`

```tsx
import { useState } from 'react';
import { useTasks } from '../hooks/usePocketBase';

interface SubtaskListProps {
  parentId: string;
  onTaskClick: (taskId: string) => void;
}

export function SubtaskList({ parentId, onTaskClick }: SubtaskListProps) {
  const { tasks, createTask, updateTask } = useTasks();
  const [isAdding, setIsAdding] = useState(false);
  const [newTitle, setNewTitle] = useState('');

  // Filter sub-tasks of this parent
  const subtasks = tasks.filter(t => t.parent === parentId);
  
  // Sort: incomplete first, then by position
  const sortedSubtasks = [...subtasks].sort((a, b) => {
    const aComplete = a.column === 'done';
    const bComplete = b.column === 'done';
    if (aComplete !== bComplete) return aComplete ? 1 : -1;
    return a.position - b.position;
  });

  // Calculate progress
  const completedCount = subtasks.filter(t => t.column === 'done').length;
  const totalCount = subtasks.length;

  // Toggle sub-task completion
  const toggleSubtask = async (subtask: Task) => {
    const newColumn = subtask.column === 'done' ? 'todo' : 'done';
    await updateTask(subtask.id, { column: newColumn });
  };

  // Add new sub-task
  const handleAddSubtask = async () => {
    if (!newTitle.trim()) return;
    
    await createTask({
      title: newTitle.trim(),
      parent: parentId,
      column: 'todo',
      type: 'chore', // Sub-tasks default to chore
      priority: 'medium',
    });
    
    setNewTitle('');
    setIsAdding(false);
  };

  return (
    <div className="subtask-list">
      <div className="subtask-header">
        <span className="subtask-label">Sub-tasks</span>
        {totalCount > 0 && (
          <span className="subtask-progress">
            {completedCount}/{totalCount}
          </span>
        )}
      </div>

      {/* Progress bar */}
      {totalCount > 0 && (
        <div className="subtask-progress-bar">
          <div 
            className="progress-fill" 
            style={{ width: `${(completedCount / totalCount) * 100}%` }} 
          />
        </div>
      )}

      {/* Sub-task list */}
      <div className="subtask-items">
        {sortedSubtasks.map(subtask => (
          <div 
            key={subtask.id} 
            className={`subtask-item ${subtask.column === 'done' ? 'completed' : ''}`}
          >
            <button
              className="subtask-checkbox"
              onClick={() => toggleSubtask(subtask)}
              aria-label={subtask.column === 'done' ? 'Mark incomplete' : 'Mark complete'}
            >
              {subtask.column === 'done' ? '&#10003;' : ''}
            </button>
            <button 
              className="subtask-title"
              onClick={() => onTaskClick(subtask.id)}
            >
              {subtask.title}
            </button>
          </div>
        ))}
      </div>

      {/* Add sub-task */}
      {isAdding ? (
        <div className="subtask-add-form">
          <input
            type="text"
            value={newTitle}
            onChange={(e) => setNewTitle(e.target.value)}
            placeholder="Sub-task title..."
            autoFocus
            onKeyDown={(e) => {
              if (e.key === 'Enter') handleAddSubtask();
              if (e.key === 'Escape') setIsAdding(false);
            }}
          />
          <div className="add-actions">
            <button onClick={handleAddSubtask}>Add</button>
            <button onClick={() => setIsAdding(false)}>Cancel</button>
          </div>
        </div>
      ) : (
        <button 
          className="subtask-add-btn"
          onClick={() => setIsAdding(true)}
        >
          + Add sub-task
        </button>
      )}
    </div>
  );
}
```

**File**: `ui/src/styles/subtask-list.css`

```css
/* Sub-task List Styles */

.subtask-list {
  margin-top: var(--space-4);
  padding-top: var(--space-4);
  border-top: 1px solid var(--border-subtle);
}

.subtask-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-2);
}

.subtask-label {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.subtask-progress {
  font-size: var(--text-xs);
  color: var(--text-muted);
}

.subtask-progress-bar {
  height: 4px;
  background: var(--bg-input);
  border-radius: 2px;
  margin-bottom: var(--space-3);
  overflow: hidden;
}

.subtask-progress-bar .progress-fill {
  height: 100%;
  background: var(--status-done);
  transition: width var(--duration-normal);
}

.subtask-items {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.subtask-item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-1) 0;
}

.subtask-item.completed .subtask-title {
  text-decoration: line-through;
  color: var(--text-muted);
}

.subtask-checkbox {
  width: 18px;
  height: 18px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--bg-input);
  color: var(--status-done);
  font-size: 12px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.subtask-checkbox:hover {
  border-color: var(--border-focus);
}

.subtask-item.completed .subtask-checkbox {
  background: var(--status-done);
  border-color: var(--status-done);
  color: white;
}

.subtask-title {
  flex: 1;
  background: none;
  border: none;
  color: var(--text-primary);
  font-size: var(--text-sm);
  text-align: left;
  cursor: pointer;
  padding: var(--space-1);
  border-radius: var(--radius-sm);
}

.subtask-title:hover {
  background: var(--bg-card-hover);
}

.subtask-add-btn {
  margin-top: var(--space-2);
  padding: var(--space-2);
  background: none;
  border: 1px dashed var(--border-default);
  border-radius: var(--radius-sm);
  color: var(--text-muted);
  font-size: var(--text-sm);
  cursor: pointer;
  width: 100%;
  text-align: left;
}

.subtask-add-btn:hover {
  border-color: var(--border-focus);
  color: var(--text-secondary);
}

.subtask-add-form {
  margin-top: var(--space-2);
}

.subtask-add-form input {
  width: 100%;
  padding: var(--space-2);
  background: var(--bg-input);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  font-size: var(--text-sm);
}

.subtask-add-form input:focus {
  outline: none;
  border-color: var(--accent);
}

.subtask-add-form .add-actions {
  display: flex;
  gap: var(--space-2);
  margin-top: var(--space-2);
}

.subtask-add-form button {
  padding: var(--space-1) var(--space-3);
  font-size: var(--text-xs);
  border-radius: var(--radius-sm);
  cursor: pointer;
}
```

---

### 8.10 Create Markdown Editor Component

**What**: Rich text editor for task descriptions with markdown support.

**Why**: Users need a good editing experience for longer task descriptions.

**File**: `ui/src/components/MarkdownEditor.tsx`

```tsx
import { useState, useRef, useEffect } from 'react';

interface MarkdownEditorProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
}

export function MarkdownEditor({ value, onChange, placeholder = 'Add a description...' }: MarkdownEditorProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [editValue, setEditValue] = useState(value);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  // Auto-resize textarea
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
      textareaRef.current.style.height = textareaRef.current.scrollHeight + 'px';
    }
  }, [editValue, isEditing]);

  // Focus textarea when entering edit mode
  useEffect(() => {
    if (isEditing && textareaRef.current) {
      textareaRef.current.focus();
      // Put cursor at end
      textareaRef.current.selectionStart = textareaRef.current.value.length;
    }
  }, [isEditing]);

  // Save changes
  const handleSave = () => {
    onChange(editValue);
    setIsEditing(false);
  };

  // Cancel editing
  const handleCancel = () => {
    setEditValue(value);
    setIsEditing(false);
  };

  // Handle keyboard shortcuts
  const handleKeyDown = (e: React.KeyboardEvent) => {
    // Cmd/Ctrl + Enter to save
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault();
      handleSave();
    }
    // Escape to cancel
    if (e.key === 'Escape') {
      handleCancel();
    }
    // Bold: Cmd/Ctrl + B
    if ((e.metaKey || e.ctrlKey) && e.key === 'b') {
      e.preventDefault();
      wrapSelection('**', '**');
    }
    // Italic: Cmd/Ctrl + I
    if ((e.metaKey || e.ctrlKey) && e.key === 'i') {
      e.preventDefault();
      wrapSelection('_', '_');
    }
    // Code: Cmd/Ctrl + `
    if ((e.metaKey || e.ctrlKey) && e.key === '`') {
      e.preventDefault();
      wrapSelection('`', '`');
    }
  };

  // Wrap selected text with markdown syntax
  const wrapSelection = (before: string, after: string) => {
    const textarea = textareaRef.current;
    if (!textarea) return;

    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const selectedText = editValue.substring(start, end);
    const newValue = 
      editValue.substring(0, start) + 
      before + selectedText + after + 
      editValue.substring(end);
    
    setEditValue(newValue);
    
    // Restore selection
    setTimeout(() => {
      textarea.focus();
      textarea.selectionStart = start + before.length;
      textarea.selectionEnd = end + before.length;
    }, 0);
  };

  // Simple markdown to HTML conversion for preview
  const renderMarkdown = (text: string): string => {
    if (!text) return '';
    
    return text
      // Headers
      .replace(/^### (.+)$/gm, '<h3>$1</h3>')
      .replace(/^## (.+)$/gm, '<h2>$1</h2>')
      .replace(/^# (.+)$/gm, '<h1>$1</h1>')
      // Bold
      .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
      // Italic
      .replace(/_(.+?)_/g, '<em>$1</em>')
      // Code
      .replace(/`(.+?)`/g, '<code>$1</code>')
      // Checkboxes
      .replace(/- \[x\] (.+)/g, '<div class="checkbox checked">&#10003; $1</div>')
      .replace(/- \[ \] (.+)/g, '<div class="checkbox">&#9633; $1</div>')
      // Bullet lists
      .replace(/^- (.+)$/gm, '<li>$1</li>')
      // Line breaks
      .replace(/\n/g, '<br/>');
  };

  if (!isEditing) {
    return (
      <div 
        className={`markdown-preview ${!value ? 'empty' : ''}`}
        onClick={() => setIsEditing(true)}
      >
        {value ? (
          <div 
            className="markdown-content"
            dangerouslySetInnerHTML={{ __html: renderMarkdown(value) }}
          />
        ) : (
          <span className="placeholder">{placeholder}</span>
        )}
      </div>
    );
  }

  return (
    <div className="markdown-editor">
      {/* Toolbar */}
      <div className="editor-toolbar">
        <button onClick={() => wrapSelection('**', '**')} title="Bold (Cmd+B)">
          <strong>B</strong>
        </button>
        <button onClick={() => wrapSelection('_', '_')} title="Italic (Cmd+I)">
          <em>I</em>
        </button>
        <button onClick={() => wrapSelection('`', '`')} title="Code (Cmd+`)">
          {'</>'}
        </button>
        <button onClick={() => wrapSelection('\n- [ ] ', '')} title="Checkbox">
          &#9744;
        </button>
        <button onClick={() => wrapSelection('\n## ', '')} title="Heading">
          H
        </button>
      </div>

      {/* Textarea */}
      <textarea
        ref={textareaRef}
        value={editValue}
        onChange={(e) => setEditValue(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder={placeholder}
        className="editor-textarea"
      />

      {/* Actions */}
      <div className="editor-actions">
        <span className="hint">Cmd+Enter to save, Esc to cancel</span>
        <div className="buttons">
          <button onClick={handleCancel} className="secondary">Cancel</button>
          <button onClick={handleSave} className="primary">Save</button>
        </div>
      </div>
    </div>
  );
}
```

**File**: `ui/src/styles/markdown-editor.css`

```css
/* Markdown Editor Styles */

.markdown-preview {
  padding: var(--space-3);
  background: var(--bg-input);
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  min-height: 80px;
  cursor: text;
  transition: border-color var(--duration-fast);
}

.markdown-preview:hover {
  border-color: var(--border-default);
}

.markdown-preview.empty {
  display: flex;
  align-items: center;
  justify-content: center;
}

.markdown-preview .placeholder {
  color: var(--text-muted);
  font-size: var(--text-sm);
}

.markdown-content {
  font-size: var(--text-sm);
  line-height: var(--leading-relaxed);
  color: var(--text-primary);
}

.markdown-content h1,
.markdown-content h2,
.markdown-content h3 {
  margin: var(--space-3) 0 var(--space-2);
  font-weight: var(--font-semibold);
}

.markdown-content h1 { font-size: var(--text-xl); }
.markdown-content h2 { font-size: var(--text-lg); }
.markdown-content h3 { font-size: var(--text-base); }

.markdown-content code {
  background: var(--bg-card);
  padding: 2px 4px;
  border-radius: var(--radius-sm);
  font-family: var(--font-mono);
  font-size: 0.9em;
}

.markdown-content .checkbox {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-1) 0;
}

.markdown-content .checkbox.checked {
  color: var(--text-muted);
  text-decoration: line-through;
}

.markdown-editor {
  border: 1px solid var(--border-focus);
  border-radius: var(--radius-md);
  background: var(--bg-input);
  overflow: hidden;
}

.editor-toolbar {
  display: flex;
  gap: var(--space-1);
  padding: var(--space-2);
  border-bottom: 1px solid var(--border-subtle);
  background: var(--bg-card);
}

.editor-toolbar button {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: none;
  border: none;
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  cursor: pointer;
  font-size: var(--text-sm);
}

.editor-toolbar button:hover {
  background: var(--bg-card-hover);
  color: var(--text-primary);
}

.editor-textarea {
  width: 100%;
  min-height: 120px;
  padding: var(--space-3);
  background: transparent;
  border: none;
  color: var(--text-primary);
  font-family: var(--font-mono);
  font-size: var(--text-sm);
  line-height: var(--leading-relaxed);
  resize: none;
}

.editor-textarea:focus {
  outline: none;
}

.editor-textarea::placeholder {
  color: var(--text-muted);
}

.editor-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-2) var(--space-3);
  border-top: 1px solid var(--border-subtle);
  background: var(--bg-card);
}

.editor-actions .hint {
  font-size: var(--text-xs);
  color: var(--text-muted);
}

.editor-actions .buttons {
  display: flex;
  gap: var(--space-2);
}

.editor-actions button {
  padding: var(--space-1) var(--space-3);
  border-radius: var(--radius-sm);
  font-size: var(--text-xs);
  cursor: pointer;
}

.editor-actions button.secondary {
  background: none;
  border: 1px solid var(--border-default);
  color: var(--text-secondary);
}

.editor-actions button.primary {
  background: var(--accent);
  border: none;
  color: white;
}
```

---

### 8.11 Create Activity Log Component

**What**: Display the history of changes to a task.

**Why**: Users and agents need to see what changed and when for context and audit.

**File**: `ui/src/components/ActivityLog.tsx`

```tsx
import { HistoryEntry } from '../hooks/usePocketBase';

interface ActivityLogProps {
  history: HistoryEntry[];
  created: string;
}

export function ActivityLog({ history, created }: ActivityLogProps) {
  // Sort history newest first
  const sortedHistory = [...history].sort(
    (a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
  );

  // Format timestamp relative to now
  const formatTime = (timestamp: string): string => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    
    return date.toLocaleDateString('en-US', { 
      month: 'short', 
      day: 'numeric',
      year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined
    });
  };

  // Format actor display
  const formatActor = (actor: string, actorDetail?: string): string => {
    if (actorDetail) {
      return actorDetail; // e.g., "claude", "opencode"
    }
    switch (actor) {
      case 'agent': return 'Agent';
      case 'cli': return 'CLI';
      case 'user': return 'You';
      default: return actor;
    }
  };

  // Format action description
  const formatAction = (entry: HistoryEntry): string => {
    switch (entry.action) {
      case 'created':
        return 'created this task';
      case 'moved':
        if (entry.changes?.field === 'column') {
          return `moved to ${entry.changes.to}`;
        }
        return 'moved this task';
      case 'updated':
        if (entry.changes) {
          const { field, from, to } = entry.changes;
          if (field === 'priority') {
            return `changed priority from ${from} to ${to}`;
          }
          if (field === 'title') {
            return `renamed to "${to}"`;
          }
          return `updated ${field}`;
        }
        return 'updated this task';
      case 'completed':
        return 'marked as done';
      default:
        return entry.action;
    }
  };

  // Get icon for actor type
  const getActorIcon = (actor: string): string => {
    switch (actor) {
      case 'agent': return '&#129302;'; // Robot emoji
      case 'cli': return '&#128187;';   // Computer emoji
      case 'user': return '&#128100;';  // Person emoji
      default: return '&#9679;';        // Bullet
    }
  };

  return (
    <div className="activity-log">
      <h3 className="activity-header">Activity</h3>
      
      <div className="activity-list">
        {sortedHistory.map((entry, index) => (
          <div key={index} className="activity-item">
            <span 
              className="activity-icon"
              dangerouslySetInnerHTML={{ __html: getActorIcon(entry.actor) }}
            />
            <div className="activity-content">
              <span className="activity-actor">
                {formatActor(entry.actor, entry.actor_detail)}
              </span>
              <span className="activity-action">
                {formatAction(entry)}
              </span>
            </div>
            <span className="activity-time">
              {formatTime(entry.timestamp)}
            </span>
          </div>
        ))}

        {/* Always show created entry at bottom */}
        <div className="activity-item">
          <span className="activity-icon">&#9679;</span>
          <div className="activity-content">
            <span className="activity-action">Created</span>
          </div>
          <span className="activity-time">
            {formatTime(created)}
          </span>
        </div>
      </div>
    </div>
  );
}
```

**File**: `ui/src/styles/activity-log.css`

```css
/* Activity Log Styles */

.activity-log {
  margin-top: var(--space-4);
  padding-top: var(--space-4);
  border-top: 1px solid var(--border-subtle);
}

.activity-header {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  margin-bottom: var(--space-3);
}

.activity-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.activity-item {
  display: flex;
  align-items: flex-start;
  gap: var(--space-2);
  font-size: var(--text-xs);
}

.activity-icon {
  width: 20px;
  text-align: center;
  color: var(--text-muted);
}

.activity-content {
  flex: 1;
  line-height: var(--leading-normal);
}

.activity-actor {
  font-weight: var(--font-medium);
  color: var(--text-primary);
  margin-right: var(--space-1);
}

.activity-action {
  color: var(--text-secondary);
}

.activity-time {
  color: var(--text-muted);
  white-space: nowrap;
}
```

---

### 8.12 Implement Import/Export Commands

**What**: CLI commands to export and import all data.

**Why**: Users need data backup and portability.

**File**: `internal/commands/export.go`

```go
package commands

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"
)

var exportFormat string
var exportBoard string

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export tasks and boards to a file",
	Long: `Export all data as JSON or CSV for backup or migration.

Examples:
  egenskriven export --format json > backup.json
  egenskriven export --format csv > tasks.csv
  egenskriven export --board work --format json > work-backup.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := app.Bootstrap(); err != nil {
			return err
		}

		switch exportFormat {
		case "json":
			return exportJSON(app, exportBoard)
		case "csv":
			return exportCSV(app, exportBoard)
		default:
			return fmt.Errorf("unsupported format: %s (use 'json' or 'csv')", exportFormat)
		}
	},
}

func init() {
	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "json", 
		"Output format: json, csv")
	exportCmd.Flags().StringVarP(&exportBoard, "board", "b", "", 
		"Export specific board only")
	
	rootCmd.AddCommand(exportCmd)
}

// ExportData represents the full export structure
type ExportData struct {
	Version  string        `json:"version"`
	Exported string        `json:"exported"`
	Boards   []ExportBoard `json:"boards"`
	Epics    []ExportEpic  `json:"epics"`
	Tasks    []ExportTask  `json:"tasks"`
}

type ExportBoard struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Prefix  string   `json:"prefix"`
	Columns []string `json:"columns"`
	Color   string   `json:"color,omitempty"`
}

type ExportEpic struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
}

type ExportTask struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type"`
	Priority    string   `json:"priority"`
	Column      string   `json:"column"`
	Position    float64  `json:"position"`
	Board       string   `json:"board,omitempty"`
	Epic        string   `json:"epic,omitempty"`
	Parent      string   `json:"parent,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	BlockedBy   []string `json:"blocked_by,omitempty"`
	DueDate     string   `json:"due_date,omitempty"`
	Created     string   `json:"created"`
	Updated     string   `json:"updated"`
}

func exportJSON(app *pocketbase.PocketBase, boardFilter string) error {
	data := ExportData{
		Version:  "1.0",
		Exported: time.Now().UTC().Format(time.RFC3339),
	}

	// Export boards
	boards, err := app.FindAllRecords("boards")
	if err == nil {
		for _, b := range boards {
			data.Boards = append(data.Boards, ExportBoard{
				ID:      b.Id,
				Name:    b.GetString("name"),
				Prefix:  b.GetString("prefix"),
				Columns: b.Get("columns").([]string),
				Color:   b.GetString("color"),
			})
		}
	}

	// Export epics
	epics, err := app.FindAllRecords("epics")
	if err == nil {
		for _, e := range epics {
			data.Epics = append(data.Epics, ExportEpic{
				ID:          e.Id,
				Title:       e.GetString("title"),
				Description: e.GetString("description"),
				Color:       e.GetString("color"),
			})
		}
	}

	// Export tasks (optionally filtered by board)
	var tasks []*core.Record
	if boardFilter != "" {
		// Find board by name or prefix
		board, err := findBoardByNameOrPrefix(app, boardFilter)
		if err != nil {
			return err
		}
		tasks, err = app.FindAllRecords("tasks",
			dbx.NewExp("board = {:board}", dbx.Params{"board": board.Id}),
		)
	} else {
		tasks, err = app.FindAllRecords("tasks")
	}
	if err != nil {
		return err
	}

	for _, t := range tasks {
		data.Tasks = append(data.Tasks, ExportTask{
			ID:          t.Id,
			Title:       t.GetString("title"),
			Description: t.GetString("description"),
			Type:        t.GetString("type"),
			Priority:    t.GetString("priority"),
			Column:      t.GetString("column"),
			Position:    t.GetFloat("position"),
			Board:       t.GetString("board"),
			Epic:        t.GetString("epic"),
			Parent:      t.GetString("parent"),
			Labels:      getStringSlice(t.Get("labels")),
			BlockedBy:   getStringSlice(t.Get("blocked_by")),
			DueDate:     t.GetString("due_date"),
			Created:     t.GetDateTime("created").String(),
			Updated:     t.GetDateTime("updated").String(),
		})
	}

	// Output JSON
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func exportCSV(app *pocketbase.PocketBase, boardFilter string) error {
	// Get tasks
	var tasks []*core.Record
	var err error
	if boardFilter != "" {
		board, err := findBoardByNameOrPrefix(app, boardFilter)
		if err != nil {
			return err
		}
		tasks, err = app.FindAllRecords("tasks",
			dbx.NewExp("board = {:board}", dbx.Params{"board": board.Id}),
		)
	} else {
		tasks, err = app.FindAllRecords("tasks")
	}
	if err != nil {
		return err
	}

	// Write CSV
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Header
	header := []string{
		"id", "title", "type", "priority", "column", "labels", 
		"epic", "parent", "due_date", "created", "updated",
	}
	writer.Write(header)

	// Rows
	for _, t := range tasks {
		row := []string{
			t.Id,
			t.GetString("title"),
			t.GetString("type"),
			t.GetString("priority"),
			t.GetString("column"),
			strings.Join(getStringSlice(t.Get("labels")), ";"),
			t.GetString("epic"),
			t.GetString("parent"),
			t.GetString("due_date"),
			t.GetDateTime("created").String(),
			t.GetDateTime("updated").String(),
		}
		writer.Write(row)
	}

	return nil
}
```

**File**: `internal/commands/import.go`

```go
package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
)

var importStrategy string
var importDryRun bool

var importCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import tasks and boards from a file",
	Long: `Import data from a JSON backup file.

Strategies:
  merge   - Skip existing records, add new ones (default)
  replace - Overwrite existing records with same ID

Examples:
  egenskriven import backup.json
  egenskriven import backup.json --strategy replace
  egenskriven import backup.json --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := app.Bootstrap(); err != nil {
			return err
		}

		filename := args[0]
		return runImport(app, filename, importStrategy, importDryRun)
	},
}

func init() {
	importCmd.Flags().StringVar(&importStrategy, "strategy", "merge",
		"Import strategy: merge, replace")
	importCmd.Flags().BoolVar(&importDryRun, "dry-run", false,
		"Preview changes without applying")

	rootCmd.AddCommand(importCmd)
}

func runImport(app *pocketbase.PocketBase, filename, strategy string, dryRun bool) error {
	// Read file
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Parse JSON
	var data ExportData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	fmt.Printf("Importing from %s (version %s, exported %s)\n", 
		filename, data.Version, data.Exported)
	fmt.Printf("Found: %d boards, %d epics, %d tasks\n",
		len(data.Boards), len(data.Epics), len(data.Tasks))

	if dryRun {
		fmt.Println("\n[DRY RUN - no changes will be made]")
	}

	stats := struct {
		BoardsCreated int
		BoardsSkipped int
		EpicsCreated  int
		EpicsSkipped  int
		TasksCreated  int
		TasksSkipped  int
		TasksUpdated  int
	}{}

	// Import boards
	boardsCollection, _ := app.FindCollectionByNameOrId("boards")
	for _, b := range data.Boards {
		existing, err := app.FindRecordById("boards", b.ID)
		
		if err == nil && existing != nil {
			if strategy == "replace" {
				if !dryRun {
					existing.Set("name", b.Name)
					existing.Set("prefix", b.Prefix)
					existing.Set("columns", b.Columns)
					existing.Set("color", b.Color)
					app.Save(existing)
				}
				stats.BoardsSkipped++ // count as updated
			} else {
				stats.BoardsSkipped++
			}
			continue
		}

		if !dryRun {
			record := core.NewRecord(boardsCollection)
			record.SetId(b.ID)
			record.Set("name", b.Name)
			record.Set("prefix", b.Prefix)
			record.Set("columns", b.Columns)
			record.Set("color", b.Color)
			if err := app.Save(record); err != nil {
				return fmt.Errorf("failed to import board %s: %w", b.Name, err)
			}
		}
		stats.BoardsCreated++
	}

	// Import epics
	epicsCollection, _ := app.FindCollectionByNameOrId("epics")
	for _, e := range data.Epics {
		existing, err := app.FindRecordById("epics", e.ID)
		
		if err == nil && existing != nil {
			if strategy == "replace" {
				if !dryRun {
					existing.Set("title", e.Title)
					existing.Set("description", e.Description)
					existing.Set("color", e.Color)
					app.Save(existing)
				}
			}
			stats.EpicsSkipped++
			continue
		}

		if !dryRun {
			record := core.NewRecord(epicsCollection)
			record.SetId(e.ID)
			record.Set("title", e.Title)
			record.Set("description", e.Description)
			record.Set("color", e.Color)
			if err := app.Save(record); err != nil {
				return fmt.Errorf("failed to import epic %s: %w", e.Title, err)
			}
		}
		stats.EpicsCreated++
	}

	// Import tasks
	tasksCollection, _ := app.FindCollectionByNameOrId("tasks")
	for _, t := range data.Tasks {
		existing, err := app.FindRecordById("tasks", t.ID)
		
		if err == nil && existing != nil {
			if strategy == "replace" {
				if !dryRun {
					existing.Set("title", t.Title)
					existing.Set("description", t.Description)
					existing.Set("type", t.Type)
					existing.Set("priority", t.Priority)
					existing.Set("column", t.Column)
					existing.Set("position", t.Position)
					existing.Set("board", t.Board)
					existing.Set("epic", t.Epic)
					existing.Set("parent", t.Parent)
					existing.Set("labels", t.Labels)
					existing.Set("blocked_by", t.BlockedBy)
					existing.Set("due_date", t.DueDate)
					app.Save(existing)
				}
				stats.TasksUpdated++
			} else {
				stats.TasksSkipped++
			}
			continue
		}

		if !dryRun {
			record := core.NewRecord(tasksCollection)
			record.SetId(t.ID)
			record.Set("title", t.Title)
			record.Set("description", t.Description)
			record.Set("type", t.Type)
			record.Set("priority", t.Priority)
			record.Set("column", t.Column)
			record.Set("position", t.Position)
			record.Set("board", t.Board)
			record.Set("epic", t.Epic)
			record.Set("parent", t.Parent)
			record.Set("labels", t.Labels)
			record.Set("blocked_by", t.BlockedBy)
			record.Set("due_date", t.DueDate)
			if err := app.Save(record); err != nil {
				return fmt.Errorf("failed to import task %s: %w", t.Title, err)
			}
		}
		stats.TasksCreated++
	}

	// Print summary
	fmt.Println("\nImport summary:")
	fmt.Printf("  Boards:  %d created, %d skipped\n", stats.BoardsCreated, stats.BoardsSkipped)
	fmt.Printf("  Epics:   %d created, %d skipped\n", stats.EpicsCreated, stats.EpicsSkipped)
	fmt.Printf("  Tasks:   %d created, %d updated, %d skipped\n", 
		stats.TasksCreated, stats.TasksUpdated, stats.TasksSkipped)

	return nil
}
```

**Steps**:

1. Create both command files.

2. Test export:
   ```bash
   # Export as JSON
   egenskriven export --format json > backup.json
   
   # Export specific board
   egenskriven export --board work > work-backup.json
   
   # Export as CSV
   egenskriven export --format csv > tasks.csv
   ```

3. Test import:
   ```bash
   # Preview import
   egenskriven import backup.json --dry-run
   
   # Import with merge (skip existing)
   egenskriven import backup.json
   
   # Import with replace (overwrite)
   egenskriven import backup.json --strategy replace
   ```

---

### 8.13 Write Date Parser Tests

**What**: Unit tests for the date parsing function.

**Why**: Date parsing has many edge cases that need coverage.

**File**: `internal/commands/date_parser_test.go`

```go
package commands

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDate_ISO8601(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2025-01-15", "2025-01-15"},
		{"2025-12-31", "2025-12-31"},
		{"2024-02-29", "2024-02-29"}, // Leap year
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result, err := parseDate(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseDate_Relative(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		input    string
		expected string
	}{
		{"today", now.Format("2006-01-02")},
		{"tomorrow", now.AddDate(0, 0, 1).Format("2006-01-02")},
		{"next week", now.AddDate(0, 0, 7).Format("2006-01-02")},
		{"next month", now.AddDate(0, 1, 0).Format("2006-01-02")},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result, err := parseDate(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseDate_CaseInsensitive(t *testing.T) {
	tests := []string{"TODAY", "Today", "toDay", "TOMORROW", "Tomorrow"}
	
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := parseDate(input)
			require.NoError(t, err, "should parse %q", input)
		})
	}
}

func TestParseDate_CommonFormats(t *testing.T) {
	// These tests depend on current date for year inference
	tests := []struct {
		input       string
		shouldParse bool
	}{
		{"Jan 15", true},
		{"January 15", true},
		{"Jan 15, 2025", true},
		{"1/15/2025", true},
		{"1/15", true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result, err := parseDate(tc.input)
			if tc.shouldParse {
				require.NoError(t, err)
				assert.NotEmpty(t, result)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestParseDate_InvalidInput(t *testing.T) {
	tests := []string{
		"",
		"invalid",
		"yesterday", // Not supported
		"2025-13-01", // Invalid month
		"2025-01-32", // Invalid day
		"someday",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := parseDate(input)
			assert.Error(t, err, "should fail to parse %q", input)
		})
	}
}

func TestParseDate_FutureYearInference(t *testing.T) {
	// When parsing "Jan 1" after Jan 1st, should infer next year
	now := time.Now()
	pastDate := now.AddDate(0, -1, 0) // One month ago
	
	input := pastDate.Format("Jan 2")
	result, err := parseDate(input)
	require.NoError(t, err)
	
	parsed, _ := time.Parse("2006-01-02", result)
	// Should be in the future (this year or next)
	assert.True(t, parsed.After(now) || parsed.Equal(now) || 
		parsed.Year() > now.Year(),
		"date should be in future or current year")
}
```

---

### 8.14 Write Import/Export Tests

**What**: Integration tests for import and export commands.

**Why**: Data portability is critical - must verify it works correctly.

**File**: `internal/commands/export_test.go`

```go
package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"egenskriven/internal/testutil"
)

func TestExportJSON_EmptyDatabase(t *testing.T) {
	app := testutil.NewTestApp(t)
	
	// Capture stdout
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	err := exportJSON(app, "")
	
	w.Close()
	os.Stdout = oldStdout
	buf.ReadFrom(r)
	
	require.NoError(t, err)
	
	var data ExportData
	err = json.Unmarshal(buf.Bytes(), &data)
	require.NoError(t, err)
	
	assert.Equal(t, "1.0", data.Version)
	assert.Empty(t, data.Tasks)
	assert.Empty(t, data.Boards)
	assert.Empty(t, data.Epics)
}

func TestExportJSON_WithData(t *testing.T) {
	app := testutil.NewTestApp(t)
	
	// Create test data
	setupTestData(t, app)
	
	var buf bytes.Buffer
	// ... capture output ...
	
	err := exportJSON(app, "")
	require.NoError(t, err)
	
	var data ExportData
	err = json.Unmarshal(buf.Bytes(), &data)
	require.NoError(t, err)
	
	assert.Len(t, data.Tasks, 2)
	assert.Len(t, data.Epics, 1)
}

func TestImport_Merge(t *testing.T) {
	app := testutil.NewTestApp(t)
	
	// Create a task that exists in import file
	existingTask := createTestTask(t, app, "Existing task")
	
	// Create import file with same task ID
	importData := ExportData{
		Version: "1.0",
		Tasks: []ExportTask{
			{
				ID:       existingTask.Id,
				Title:    "Updated title", // Different title
				Type:     "feature",
				Priority: "high",
				Column:   "todo",
			},
			{
				ID:       "new-task-id",
				Title:    "New task",
				Type:     "bug",
				Priority: "urgent",
				Column:   "backlog",
			},
		},
	}
	
	tmpFile := createTempImportFile(t, importData)
	
	err := runImport(app, tmpFile, "merge", false)
	require.NoError(t, err)
	
	// Existing task should NOT be updated (merge skips)
	task, _ := app.FindRecordById("tasks", existingTask.Id)
	assert.Equal(t, "Existing task", task.GetString("title"))
	
	// New task should be created
	newTask, err := app.FindRecordById("tasks", "new-task-id")
	require.NoError(t, err)
	assert.Equal(t, "New task", newTask.GetString("title"))
}

func TestImport_Replace(t *testing.T) {
	app := testutil.NewTestApp(t)
	
	existingTask := createTestTask(t, app, "Existing task")
	
	importData := ExportData{
		Version: "1.0",
		Tasks: []ExportTask{
			{
				ID:       existingTask.Id,
				Title:    "Updated title",
				Type:     "feature",
				Priority: "high",
				Column:   "todo",
			},
		},
	}
	
	tmpFile := createTempImportFile(t, importData)
	
	err := runImport(app, tmpFile, "replace", false)
	require.NoError(t, err)
	
	// Existing task SHOULD be updated (replace overwrites)
	task, _ := app.FindRecordById("tasks", existingTask.Id)
	assert.Equal(t, "Updated title", task.GetString("title"))
}

func TestImport_DryRun(t *testing.T) {
	app := testutil.NewTestApp(t)
	
	importData := ExportData{
		Version: "1.0",
		Tasks: []ExportTask{
			{
				ID:       "dry-run-task",
				Title:    "Should not be created",
				Type:     "feature",
				Priority: "medium",
				Column:   "backlog",
			},
		},
	}
	
	tmpFile := createTempImportFile(t, importData)
	
	err := runImport(app, tmpFile, "merge", true) // dry-run = true
	require.NoError(t, err)
	
	// Task should NOT be created
	_, err = app.FindRecordById("tasks", "dry-run-task")
	assert.Error(t, err, "task should not exist after dry run")
}

// Helper functions

func createTempImportFile(t *testing.T, data ExportData) string {
	t.Helper()
	
	file, err := os.CreateTemp("", "import-test-*.json")
	require.NoError(t, err)
	
	t.Cleanup(func() { os.Remove(file.Name()) })
	
	encoder := json.NewEncoder(file)
	err = encoder.Encode(data)
	require.NoError(t, err)
	
	file.Close()
	return file.Name()
}
```

---

## Verification Checklist

Complete each section in order. Check off each item as you verify it.

### Database Migration Verification

- [ ] **Due date migration runs**
  ```bash
  ./egenskriven migrate
  ```
  Should complete without errors.

- [ ] **Due date field exists**
  
  Check admin UI at `http://localhost:8090/_/` - tasks collection should have `due_date` field.

- [ ] **Parent field exists**
  
  Tasks collection should have `parent` relation field.

### CLI Due Date Verification

- [ ] **Create task with due date**
  ```bash
  egenskriven add "Test task" --due "2025-01-20"
  egenskriven add "Tomorrow task" --due tomorrow
  egenskriven add "Next week task" --due "next week"
  ```

- [ ] **Filter by due date**
  ```bash
  egenskriven list --due-before "next week"
  egenskriven list --has-due
  egenskriven list --no-due
  ```

### CLI Sub-task Verification

- [ ] **Create sub-task**
  ```bash
  # Create parent
  egenskriven add "Parent task"
  # Note the ID, then create sub-task
  egenskriven add "Sub-task 1" --parent <parent-id>
  egenskriven add "Sub-task 2" --parent <parent-id>
  ```

- [ ] **Show displays sub-tasks**
  ```bash
  egenskriven show <parent-id>
  ```
  Should list sub-tasks.

- [ ] **Filter by parent**
  ```bash
  egenskriven list --has-parent
  egenskriven list --no-parent
  ```

### UI Epic Verification

- [ ] **Epic picker works**
  
  Open task detail, click epic field, can select/change epic.

- [ ] **Epic list in sidebar**
  
  Epics appear in sidebar with task counts.

- [ ] **Epic detail view**
  
  Click epic to see progress and all tasks.

### UI Due Date Verification

- [ ] **Date picker opens**
  
  Click due date in task detail, calendar appears.

- [ ] **Can set date**
  
  Select a date, save, verify it persists.

- [ ] **Quick select works**
  
  "Today", "Tomorrow", "Next week" shortcuts work.

- [ ] **Due date shows on card**
  
  Task cards display due date when set.

- [ ] **Overdue highlighting**
  
  Tasks past due date show red indicator.

### UI Sub-task Verification

- [ ] **Sub-task list displays**
  
  Parent task shows sub-task list in detail panel.

- [ ] **Can add sub-task**
  
  "+ Add sub-task" button works.

- [ ] **Can toggle completion**
  
  Clicking checkbox marks sub-task done.

- [ ] **Progress bar updates**
  
  Progress reflects sub-task completion.

### UI Markdown Editor Verification

- [ ] **Click to edit**
  
  Clicking description enters edit mode.

- [ ] **Toolbar buttons work**
  
  Bold, italic, code buttons insert markdown.

- [ ] **Keyboard shortcuts work**
  
  Cmd+B for bold, Cmd+I for italic.

- [ ] **Preview renders markdown**
  
  After saving, markdown renders properly.

### UI Activity Log Verification

- [ ] **History displays**
  
  Task detail shows activity log section.

- [ ] **Actions are logged**
  
  Create, update, move actions appear in log.

- [ ] **Times are relative**
  
  Shows "2m ago", "1h ago", etc.

### Import/Export Verification

- [ ] **Export JSON works**
  ```bash
  egenskriven export --format json > backup.json
  cat backup.json
  ```
  Should produce valid JSON.

- [ ] **Export CSV works**
  ```bash
  egenskriven export --format csv > tasks.csv
  cat tasks.csv
  ```
  Should produce valid CSV.

- [ ] **Import dry-run works**
  ```bash
  egenskriven import backup.json --dry-run
  ```
  Should show what would be imported.

- [ ] **Import merge works**
  ```bash
  # Clear database first for clean test
  egenskriven import backup.json
  ```
  Should import data.

### Test Verification

- [ ] **All tests pass**
  ```bash
  make test
  ```
  All tests should pass.

- [ ] **Date parser tests pass**
  ```bash
  go test ./internal/commands/... -run TestParseDate -v
  ```

- [ ] **Import/export tests pass**
  ```bash
  go test ./internal/commands/... -run TestExport -v
  go test ./internal/commands/... -run TestImport -v
  ```

- [ ] **UI tests pass**
  ```bash
  cd ui && npm test
  ```

---

## File Summary

### CLI Files

| File | Lines | Purpose |
|------|-------|---------|
| `migrations/2_due_dates.go` | ~30 | Due date field migration |
| `migrations/3_subtasks.go` | ~25 | Parent field migration |
| `internal/commands/date_parser.go` | ~60 | Date parsing utility |
| `internal/commands/date_parser_test.go` | ~100 | Date parser tests |
| `internal/commands/export.go` | ~180 | Export command |
| `internal/commands/import.go` | ~160 | Import command |
| `internal/commands/export_test.go` | ~120 | Import/export tests |

### UI Files

| File | Lines | Purpose |
|------|-------|---------|
| `ui/src/components/DatePicker.tsx` | ~180 | Date picker component |
| `ui/src/components/EpicPicker.tsx` | ~80 | Epic selector |
| `ui/src/components/EpicList.tsx` | ~60 | Epic sidebar list |
| `ui/src/components/EpicDetail.tsx` | ~150 | Epic detail view |
| `ui/src/components/SubtaskList.tsx` | ~130 | Sub-task management |
| `ui/src/components/MarkdownEditor.tsx` | ~200 | Rich text editor |
| `ui/src/components/ActivityLog.tsx` | ~100 | Activity history |
| `ui/src/styles/*.css` | ~400 | Component styles |

**Total new code**: ~2,000+ lines

---

## What You Should Have Now

After completing Phase 8, your application should have:

```
Features:
  [x] Epic UI with sidebar, picker, and detail view
  [x] Due dates with date picker and overdue highlighting
  [x] Sub-tasks with progress tracking
  [x] Markdown editor for descriptions
  [x] Activity log showing task history
  [x] Import/export for data portability
  
CLI:
  egenskriven add "task" --due "2025-01-20"
  egenskriven add "task" --parent <id>
  egenskriven list --due-before "next week"
  egenskriven list --has-parent
  egenskriven export --format json > backup.json
  egenskriven import backup.json --dry-run
```

---

## Next Phase

**Phase 9: Release** will add:
- Cross-platform builds (macOS, Linux, Windows)
- Version embedding in binary
- GitHub release workflow
- Shell completion generation
- Installation documentation
- Final testing on all platforms

---

## Troubleshooting

### Migration fails with "field already exists"

**Problem**: Running migration again after it succeeded.

**Solution**: Migrations are idempotent by design. Check if field already exists in admin UI. If needed, reset database:
```bash
rm -rf pb_data
./egenskriven serve
./egenskriven migrate
```

### Date picker doesn't close

**Problem**: Click outside not detected.

**Solution**: Check the `useEffect` for click outside is properly attached. Ensure `containerRef` is set on the wrapper div.

### Sub-tasks not showing in parent

**Problem**: Parent task detail doesn't show sub-tasks.

**Solution**: Verify:
1. Sub-tasks have `parent` field set correctly
2. Filter query uses correct field name
3. Real-time subscription includes sub-task changes

### Import fails with "collection not found"

**Problem**: Importing into fresh database without collections.

**Solution**: Run migrations first:
```bash
./egenskriven migrate
./egenskriven import backup.json
```

### Activity log shows wrong time

**Problem**: Timestamps are in wrong timezone.

**Solution**: Activity log should use UTC internally and display relative times. Check:
1. History entries use ISO 8601 with Z suffix
2. Display function uses local time for formatting

### Export produces empty file

**Problem**: `egenskriven export` outputs nothing.

**Solution**: Check:
1. Database has data (run `egenskriven list` first)
2. If using `--board` flag, verify board exists
3. Check for errors in command output

---

## Glossary

| Term | Definition |
|------|------------|
| **Epic** | A collection of related tasks representing a larger goal |
| **Sub-task** | A task that is a child of another task |
| **Activity log** | Chronological history of changes to a task |
| **Due date** | Deadline for task completion |
| **ISO 8601** | International date format standard (YYYY-MM-DD) |
| **Migration** | Database schema change that adds/modifies fields |
| **Dry run** | Preview mode that shows what would happen without making changes |
