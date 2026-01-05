# Phase 6: Filtering & Views

**Goal**: Implement advanced filtering, search, saved views, and display options for flexible task management.

**Duration Estimate**: 4-5 days

**Prerequisites**: Phase 4 (Interactive UI) and Phase 5 (Multi-Board) complete.

**Deliverable**: A fully functional filtering system with saved views, search, list view toggle, and customizable display options.

---

## Overview

Phase 6 transforms EgenSkriven from a basic kanban board into a powerful task management system:

- **Filter tasks** by status, priority, type, labels, due dates, and more
- **Search** across task titles and descriptions in real-time
- **Save views** with specific filter combinations for quick access
- **Toggle between board and list views**
- **Customize display options** like card density and visible properties

### Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Filter System                            │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────────┐   │
│  │ Filter State │───▶│ Filter Logic │───▶│ Filtered Tasks   │   │
│  │  (Zustand)   │    │  (useMemo)   │    │  (Display)       │   │
│  └──────────────┘    └──────────────┘    └──────────────────┘   │
│         │                                         │              │
│         ▼                                         ▼              │
│  ┌──────────────┐                        ┌──────────────────┐   │
│  │ Saved Views  │                        │ Board/List View  │   │
│  │ (PocketBase) │                        │                  │   │
│  └──────────────┘                        └──────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

---

## Tasks

### 6.1 Create Views Collection Migration

**What**: Add a `views` collection to PocketBase for storing saved views.

**File**: `migrations/3_views.go`

```go
package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection := core.NewBaseCollection("views")

		collection.Fields.Add(&core.TextField{
			Name: "name", Required: true, Min: 1, Max: 100,
		})
		collection.Fields.Add(&core.RelationField{
			Name: "board", CollectionId: "boards", Required: true, MaxSelect: 1,
		})
		collection.Fields.Add(&core.JSONField{
			Name: "filters", MaxSize: 10000,
		})
		collection.Fields.Add(&core.JSONField{
			Name: "display", MaxSize: 5000,
		})
		collection.Fields.Add(&core.BoolField{
			Name: "is_favorite",
		})
		collection.Fields.Add(&core.SelectField{
			Name: "match_mode", Values: []string{"all", "any"},
		})

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("views")
		if err != nil {
			return nil
		}
		return app.Delete(collection)
	})
}
```

**Steps**:
```bash
touch migrations/3_views.go
# Add code above
./egenskriven migrate up
```

---

### 6.2 Create Filter State Management

**What**: Create a Zustand store for managing filter state.

**File**: `ui/src/stores/filters.ts`

```typescript
import { create } from 'zustand';

export type FilterOperator =
  | 'is' | 'is_not' | 'is_any_of'
  | 'includes_any' | 'includes_all' | 'includes_none'
  | 'before' | 'after' | 'is_set' | 'is_not_set' | 'contains';

export type FilterField =
  | 'column' | 'priority' | 'type' | 'labels'
  | 'due_date' | 'epic' | 'created_by' | 'title';

export interface Filter {
  id: string;
  field: FilterField;
  operator: FilterOperator;
  value: string | string[] | null;
}

export interface DisplayOptions {
  viewMode: 'board' | 'list';
  density: 'compact' | 'comfortable';
  visibleFields: string[];
  groupBy: 'column' | 'priority' | 'type' | 'epic' | null;
}

export type MatchMode = 'all' | 'any';

interface FilterState {
  filters: Filter[];
  matchMode: MatchMode;
  searchQuery: string;
  displayOptions: DisplayOptions;
  currentViewId: string | null;
  isModified: boolean;
  
  addFilter: (filter: Omit<Filter, 'id'>) => void;
  removeFilter: (filterId: string) => void;
  updateFilter: (filterId: string, updates: Partial<Filter>) => void;
  clearFilters: () => void;
  setMatchMode: (mode: MatchMode) => void;
  setSearchQuery: (query: string) => void;
  setDisplayOptions: (options: Partial<DisplayOptions>) => void;
  loadView: (viewId: string, filters: Filter[], matchMode: MatchMode, display: DisplayOptions) => void;
  markAsModified: () => void;
}

const generateId = () => Math.random().toString(36).substring(2, 9);

const defaultDisplayOptions: DisplayOptions = {
  viewMode: 'board',
  density: 'comfortable',
  visibleFields: ['priority', 'labels', 'due_date'],
  groupBy: 'column',
};

export const useFilterStore = create<FilterState>((set) => ({
  filters: [],
  matchMode: 'all',
  searchQuery: '',
  displayOptions: defaultDisplayOptions,
  currentViewId: null,
  isModified: false,

  addFilter: (filter) => set((state) => ({
    filters: [...state.filters, { ...filter, id: generateId() }],
    isModified: true,
  })),

  removeFilter: (filterId) => set((state) => ({
    filters: state.filters.filter((f) => f.id !== filterId),
    isModified: true,
  })),

  updateFilter: (filterId, updates) => set((state) => ({
    filters: state.filters.map((f) => f.id === filterId ? { ...f, ...updates } : f),
    isModified: true,
  })),

  clearFilters: () => set({ filters: [], searchQuery: '', isModified: true }),

  setMatchMode: (mode) => set({ matchMode: mode, isModified: true }),

  setSearchQuery: (query) => set({ searchQuery: query }),

  setDisplayOptions: (options) => set((state) => ({
    displayOptions: { ...state.displayOptions, ...options },
    isModified: true,
  })),

  loadView: (viewId, filters, matchMode, display) => set({
    currentViewId: viewId, filters, matchMode, displayOptions: display, isModified: false,
  }),

  markAsModified: () => set({ isModified: true }),
}));
```

**Steps**:
```bash
cd ui && npm install zustand
mkdir -p src/stores
touch src/stores/filters.ts
```

---

### 6.3 Implement Filter Logic

**What**: Create a hook that applies filters to the task list.

**File**: `ui/src/hooks/useFilteredTasks.ts`

```typescript
import { useMemo } from 'react';
import { Task } from './usePocketBase';
import { Filter, useFilterStore } from '../stores/filters';

function matchesFilter(task: Task, filter: Filter): boolean {
  const { field, operator, value } = filter;

  switch (field) {
    case 'column':
    case 'priority':
    case 'type':
    case 'created_by':
      return matchSelectFilter(task[field], operator, value);
    case 'labels':
      return matchLabelsFilter(task.labels || [], operator, value as string[]);
    case 'due_date':
      return matchDateFilter(task.due_date, operator, value as string);
    case 'epic':
      return matchRelationFilter(task.epic, operator, value as string);
    case 'title':
      return matchTextFilter(task.title, operator, value as string);
    default:
      return true;
  }
}

function matchSelectFilter(taskValue: string | undefined, operator: string, filterValue: string | string[] | null): boolean {
  if (!taskValue) return operator === 'is_not_set';
  if (filterValue === null) return operator === 'is_set';

  switch (operator) {
    case 'is': return taskValue === filterValue;
    case 'is_not': return taskValue !== filterValue;
    case 'is_any_of': return Array.isArray(filterValue) && filterValue.includes(taskValue);
    case 'is_set': return !!taskValue;
    case 'is_not_set': return !taskValue;
    default: return true;
  }
}

function matchLabelsFilter(taskLabels: string[], operator: string, filterLabels: string[]): boolean {
  if (!filterLabels?.length) {
    return operator === 'is_set' ? taskLabels.length > 0 : operator === 'is_not_set' ? taskLabels.length === 0 : true;
  }
  switch (operator) {
    case 'includes_any': return filterLabels.some((l) => taskLabels.includes(l));
    case 'includes_all': return filterLabels.every((l) => taskLabels.includes(l));
    case 'includes_none': return !filterLabels.some((l) => taskLabels.includes(l));
    default: return true;
  }
}

function matchDateFilter(taskDate: string | undefined, operator: string, filterDate: string | null): boolean {
  if (operator === 'is_set') return !!taskDate;
  if (operator === 'is_not_set') return !taskDate;
  if (!taskDate || !filterDate) return false;

  const task = new Date(taskDate);
  const filter = new Date(filterDate);

  switch (operator) {
    case 'is': return task.toDateString() === filter.toDateString();
    case 'before': return task < filter;
    case 'after': return task > filter;
    default: return true;
  }
}

function matchRelationFilter(taskRelation: string | undefined, operator: string, filterValue: string | null): boolean {
  if (operator === 'is_set') return !!taskRelation;
  if (operator === 'is_not_set') return !taskRelation;
  return operator === 'is' ? taskRelation === filterValue : taskRelation !== filterValue;
}

function matchTextFilter(taskValue: string | undefined, operator: string, filterValue: string | null): boolean {
  if (!filterValue) return true;
  if (!taskValue) return false;
  const lower = taskValue.toLowerCase();
  const filterLower = filterValue.toLowerCase();
  return operator === 'contains' ? lower.includes(filterLower) : lower === filterLower;
}

function matchesSearch(task: Task, query: string): boolean {
  if (!query?.trim()) return true;
  const q = query.toLowerCase().trim();
  return (task.title || '').toLowerCase().includes(q) ||
         (task.description || '').toLowerCase().includes(q) ||
         task.id.toLowerCase().includes(q);
}

export function useFilteredTasks(tasks: Task[]): Task[] {
  const filters = useFilterStore((state) => state.filters);
  const matchMode = useFilterStore((state) => state.matchMode);
  const searchQuery = useFilterStore((state) => state.searchQuery);

  return useMemo(() => {
    return tasks.filter((task) => {
      if (!matchesSearch(task, searchQuery)) return false;
      if (filters.length === 0) return true;
      
      return matchMode === 'all'
        ? filters.every((f) => matchesFilter(task, f))
        : filters.some((f) => matchesFilter(task, f));
    });
  }, [tasks, filters, matchMode, searchQuery]);
}
```

**Test file**: `ui/src/hooks/useFilteredTasks.test.ts`

Write tests covering:
- Returns all tasks when no filters active
- Filters by priority, type, labels (includes_any, includes_all)
- Filters by due_date (is_set, is_not_set, before, after)
- Combines filters with AND (match all) and OR (match any)
- Search filters by title, description (case-insensitive)
- Combines search and filters

---

### 6.4 Create Filter Bar Component

**What**: Display active filters with quick actions.

**File**: `ui/src/components/FilterBar.tsx`

```typescript
import { useFilterStore, Filter } from '../stores/filters';

const FIELD_LABELS: Record<string, string> = {
  column: 'Status', priority: 'Priority', type: 'Type',
  labels: 'Labels', due_date: 'Due Date', epic: 'Epic',
};

const OPERATOR_LABELS: Record<string, string> = {
  is: 'is', is_not: 'is not', is_any_of: 'is any of',
  includes_any: 'includes any of', includes_all: 'includes all of',
  includes_none: 'includes none of', before: 'before', after: 'after',
  is_set: 'is set', is_not_set: 'is not set', contains: 'contains',
};

function FilterPill({ filter, onRemove }: { filter: Filter; onRemove: () => void }) {
  const formatValue = () => {
    if (filter.value === null) return '';
    return Array.isArray(filter.value) ? filter.value.join(', ') : String(filter.value);
  };

  return (
    <div className="filter-pill">
      <span className="filter-pill-field">{FIELD_LABELS[filter.field]}</span>
      <span className="filter-pill-operator">{OPERATOR_LABELS[filter.operator]}</span>
      {filter.value !== null && <span className="filter-pill-value">{formatValue()}</span>}
      <button className="filter-pill-remove" onClick={onRemove}>&times;</button>
    </div>
  );
}

interface FilterBarProps {
  totalTasks: number;
  filteredTasks: number;
  onOpenFilterBuilder: () => void;
}

export function FilterBar({ totalTasks, filteredTasks, onOpenFilterBuilder }: FilterBarProps) {
  const filters = useFilterStore((s) => s.filters);
  const matchMode = useFilterStore((s) => s.matchMode);
  const searchQuery = useFilterStore((s) => s.searchQuery);
  const clearFilters = useFilterStore((s) => s.clearFilters);
  const removeFilter = useFilterStore((s) => s.removeFilter);
  const setMatchMode = useFilterStore((s) => s.setMatchMode);
  const setSearchQuery = useFilterStore((s) => s.setSearchQuery);

  const hasActiveFilters = filters.length > 0 || searchQuery.trim() !== '';

  return (
    <div className="filter-bar">
      <button className="filter-bar-button" onClick={onOpenFilterBuilder}>
        Filter {filters.length > 0 && <span className="filter-count">{filters.length}</span>}
      </button>

      {hasActiveFilters && (
        <div className="filter-bar-pills">
          {searchQuery && (
            <div className="filter-pill">
              <span>Search: "{searchQuery}"</span>
              <button onClick={() => setSearchQuery('')}>&times;</button>
            </div>
          )}
          {filters.map((f) => (
            <FilterPill key={f.id} filter={f} onRemove={() => removeFilter(f.id)} />
          ))}
          {filters.length > 1 && (
            <div className="filter-match-mode">
              <button className={matchMode === 'all' ? 'active' : ''} onClick={() => setMatchMode('all')}>All</button>
              <button className={matchMode === 'any' ? 'active' : ''} onClick={() => setMatchMode('any')}>Any</button>
            </div>
          )}
          <button className="filter-bar-clear" onClick={clearFilters}>Clear all</button>
        </div>
      )}

      <div className="filter-bar-stats">
        {hasActiveFilters ? `${filteredTasks} of ${totalTasks}` : `${totalTasks} tasks`}
      </div>
    </div>
  );
}
```

**Styles**: Create `ui/src/components/FilterBar.css` with styles for `.filter-bar`, `.filter-pill`, `.filter-count`, `.filter-match-mode`, etc.

---

### 6.5 Create Filter Builder Component

**What**: UI for constructing new filters with field/operator/value selection.

**File**: `ui/src/components/FilterBuilder.tsx`

Key elements:
- Field selector dropdown (Status, Priority, Type, Labels, Due Date, Epic)
- Operator dropdown (changes based on field type)
- Value input (select, multi-select, date picker, or text based on field)
- Add filter button
- Display existing filters with AND/OR connectors
- Match mode toggle (All filters / Any filter)

```typescript
const FILTER_FIELDS = [
  { value: 'column', label: 'Status' },
  { value: 'priority', label: 'Priority' },
  { value: 'type', label: 'Type' },
  { value: 'labels', label: 'Labels' },
  { value: 'due_date', label: 'Due Date' },
  { value: 'epic', label: 'Epic' },
];

const OPERATORS_BY_FIELD = {
  column: [{ value: 'is', label: 'is' }, { value: 'is_not', label: 'is not' }, { value: 'is_any_of', label: 'is any of' }],
  priority: [{ value: 'is', label: 'is' }, { value: 'is_not', label: 'is not' }],
  type: [{ value: 'is', label: 'is' }, { value: 'is_not', label: 'is not' }],
  labels: [
    { value: 'includes_any', label: 'includes any of' },
    { value: 'includes_all', label: 'includes all of' },
    { value: 'is_set', label: 'is set' },
    { value: 'is_not_set', label: 'is not set' },
  ],
  due_date: [
    { value: 'before', label: 'before' },
    { value: 'after', label: 'after' },
    { value: 'is_set', label: 'is set' },
    { value: 'is_not_set', label: 'is not set' },
  ],
  epic: [{ value: 'is', label: 'is' }, { value: 'is_set', label: 'is set' }, { value: 'is_not_set', label: 'is not set' }],
};

const VALUE_OPTIONS = {
  column: ['backlog', 'todo', 'in_progress', 'review', 'done'],
  priority: ['urgent', 'high', 'medium', 'low'],
  type: ['bug', 'feature', 'chore'],
};
```

Implement:
- Modal overlay with close on backdrop click
- State for selectedField, selectedOperator, selectedValue
- Dynamic value input based on field type
- Add filter button that calls `addFilter` from store

**Styles**: Create `ui/src/components/FilterBuilder.css`

---

### 6.6 Create Search Component

**What**: Search input that filters tasks in real-time.

**File**: `ui/src/components/Search.tsx`

```typescript
import { useRef, useEffect } from 'react';
import { useFilterStore } from '../stores/filters';

export function SearchBar() {
  const inputRef = useRef<HTMLInputElement>(null);
  const searchQuery = useFilterStore((s) => s.searchQuery);
  const setSearchQuery = useFilterStore((s) => s.setSearchQuery);

  // Global "/" shortcut to focus
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === '/' && !['INPUT', 'TEXTAREA'].includes((e.target as HTMLElement).tagName)) {
        e.preventDefault();
        inputRef.current?.focus();
      }
    };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, []);

  return (
    <div className="search-bar">
      <input
        ref={inputRef}
        type="text"
        placeholder="Search... (/)"
        value={searchQuery}
        onChange={(e) => setSearchQuery(e.target.value)}
      />
      {searchQuery && (
        <button onClick={() => setSearchQuery('')}>&times;</button>
      )}
    </div>
  );
}
```

---

### 6.7 Implement Saved Views

**What**: Hook and components for saving, loading, and managing views.

**File**: `ui/src/hooks/useViews.ts`

```typescript
import { useState, useEffect, useCallback } from 'react';
import PocketBase from 'pocketbase';
import { Filter, MatchMode, DisplayOptions, useFilterStore } from '../stores/filters';

const pb = new PocketBase('/');

export interface View {
  id: string;
  name: string;
  board: string;
  filters: Filter[];
  display: DisplayOptions;
  is_favorite: boolean;
  match_mode: MatchMode;
}

export function useViews(boardId: string | null) {
  const [views, setViews] = useState<View[]>([]);
  const [loading, setLoading] = useState(true);
  const loadView = useFilterStore((s) => s.loadView);

  useEffect(() => {
    if (!boardId) { setViews([]); setLoading(false); return; }
    
    pb.collection('views')
      .getFullList<View>({ filter: `board = "${boardId}"`, sort: '-is_favorite,name' })
      .then(setViews)
      .finally(() => setLoading(false));
  }, [boardId]);

  // Subscribe to realtime updates
  useEffect(() => {
    if (!boardId) return;
    pb.collection('views').subscribe<View>('*', (e) => {
      if (e.record.board !== boardId) return;
      if (e.action === 'create') setViews((prev) => [...prev, e.record]);
      else if (e.action === 'update') setViews((prev) => prev.map((v) => v.id === e.record.id ? e.record : v));
      else if (e.action === 'delete') setViews((prev) => prev.filter((v) => v.id !== e.record.id));
    });
    return () => { pb.collection('views').unsubscribe('*'); };
  }, [boardId]);

  const createView = useCallback(async (name: string, filters: Filter[], matchMode: MatchMode, display: DisplayOptions) => {
    if (!boardId) throw new Error('No board');
    return pb.collection('views').create({ name, board: boardId, filters: JSON.stringify(filters), match_mode: matchMode, display: JSON.stringify(display), is_favorite: false });
  }, [boardId]);

  const deleteView = useCallback((id: string) => pb.collection('views').delete(id), []);
  
  const toggleFavorite = useCallback(async (id: string) => {
    const view = views.find((v) => v.id === id);
    if (view) await pb.collection('views').update(id, { is_favorite: !view.is_favorite });
  }, [views]);

  const applyView = useCallback((view: View) => {
    const filters = typeof view.filters === 'string' ? JSON.parse(view.filters) : view.filters;
    const display = typeof view.display === 'string' ? JSON.parse(view.display) : view.display;
    loadView(view.id, filters, view.match_mode, display);
  }, [loadView]);

  return { views, loading, createView, deleteView, toggleFavorite, applyView };
}
```

**File**: `ui/src/components/ViewsSidebar.tsx`

Implement:
- Favorites section (views with is_favorite=true)
- Views section (regular views)
- View item with click to apply, star to favorite, delete button
- "New view" button that saves current filters
- "Modified" indicator when current view has changes

---

### 6.8 Create List View Component

**What**: Alternative table/list view for tasks.

**File**: `ui/src/components/ListView.tsx`

```typescript
import { useMemo, useState, useCallback } from 'react';
import { Task } from '../hooks/usePocketBase';

const COLUMNS = [
  { key: 'column', label: 'Status', width: '100px', sortable: true },
  { key: 'id', label: 'ID', width: '100px', sortable: true },
  { key: 'title', label: 'Title', width: 'auto', sortable: true },
  { key: 'labels', label: 'Labels', width: '150px', sortable: false },
  { key: 'priority', label: 'Priority', width: '100px', sortable: true },
  { key: 'due_date', label: 'Due', width: '100px', sortable: true },
];

interface ListViewProps {
  tasks: Task[];
  onTaskClick: (task: Task) => void;
  selectedTaskId: string | null;
}

export function ListView({ tasks, onTaskClick, selectedTaskId }: ListViewProps) {
  const [sortColumn, setSortColumn] = useState('column');
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc');

  const sortedTasks = useMemo(() => {
    return [...tasks].sort((a, b) => {
      const aVal = a[sortColumn as keyof Task];
      const bVal = b[sortColumn as keyof Task];
      if (aVal == null) return sortDirection === 'asc' ? 1 : -1;
      if (bVal == null) return sortDirection === 'asc' ? -1 : 1;
      if (typeof aVal === 'string' && typeof bVal === 'string') {
        return sortDirection === 'asc' ? aVal.localeCompare(bVal) : bVal.localeCompare(aVal);
      }
      return 0;
    });
  }, [tasks, sortColumn, sortDirection]);

  const handleSort = useCallback((col: string) => {
    if (col === sortColumn) setSortDirection((d) => d === 'asc' ? 'desc' : 'asc');
    else { setSortColumn(col); setSortDirection('asc'); }
  }, [sortColumn]);

  return (
    <div className="list-view">
      <table>
        <thead>
          <tr>
            {COLUMNS.map((col) => (
              <th key={col.key} onClick={() => col.sortable && handleSort(col.key)}>
                {col.label} {sortColumn === col.key && (sortDirection === 'asc' ? '↑' : '↓')}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {sortedTasks.map((task) => (
            <tr key={task.id} className={selectedTaskId === task.id ? 'selected' : ''} onClick={() => onTaskClick(task)}>
              <td><StatusBadge status={task.column} /></td>
              <td>{task.id.slice(0, 8)}</td>
              <td>{task.title}</td>
              <td>{(task.labels || []).map((l) => <span key={l} className="label-pill">{l}</span>)}</td>
              <td><PriorityBadge priority={task.priority} /></td>
              <td>{task.due_date || '-'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const colors = { backlog: '#6B7280', todo: '#E5E5E5', in_progress: '#F59E0B', review: '#A855F7', done: '#22C55E' };
  return <span style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
    <span style={{ width: 8, height: 8, borderRadius: '50%', background: colors[status] }} />
    {status.replace('_', ' ')}
  </span>;
}

function PriorityBadge({ priority }: { priority: string }) {
  const colors = { urgent: '#EF4444', high: '#F97316', medium: '#EAB308', low: '#6B7280' };
  return <span style={{ color: colors[priority] }}>{priority}</span>;
}
```

---

### 6.9 Create Display Options Component

**What**: UI for configuring display settings.

**File**: `ui/src/components/DisplayOptions.tsx`

```typescript
import { useFilterStore, DisplayOptions } from '../stores/filters';

interface Props { isOpen: boolean; onClose: () => void; }

export function DisplayOptionsMenu({ isOpen, onClose }: Props) {
  const displayOptions = useFilterStore((s) => s.displayOptions);
  const setDisplayOptions = useFilterStore((s) => s.setDisplayOptions);

  if (!isOpen) return null;

  return (
    <div className="display-options-overlay" onClick={onClose}>
      <div className="display-options" onClick={(e) => e.stopPropagation()}>
        <h4>Display options</h4>
        
        {/* View mode */}
        <div className="section">
          <label>View</label>
          <div className="buttons">
            <button className={displayOptions.viewMode === 'board' ? 'active' : ''} onClick={() => setDisplayOptions({ viewMode: 'board' })}>Board</button>
            <button className={displayOptions.viewMode === 'list' ? 'active' : ''} onClick={() => setDisplayOptions({ viewMode: 'list' })}>List</button>
          </div>
        </div>

        {/* Density */}
        <div className="section">
          <label>Density</label>
          <div className="buttons">
            <button className={displayOptions.density === 'compact' ? 'active' : ''} onClick={() => setDisplayOptions({ density: 'compact' })}>Compact</button>
            <button className={displayOptions.density === 'comfortable' ? 'active' : ''} onClick={() => setDisplayOptions({ density: 'comfortable' })}>Comfortable</button>
          </div>
        </div>

        {/* Visible fields */}
        <div className="section">
          <label>Show on cards</label>
          {['priority', 'labels', 'due_date', 'epic', 'type'].map((field) => (
            <label key={field}>
              <input
                type="checkbox"
                checked={displayOptions.visibleFields.includes(field)}
                onChange={(e) => {
                  const fields = e.target.checked
                    ? [...displayOptions.visibleFields, field]
                    : displayOptions.visibleFields.filter((f) => f !== field);
                  setDisplayOptions({ visibleFields: fields });
                }}
              />
              {field}
            </label>
          ))}
        </div>

        {/* Group by (board only) */}
        {displayOptions.viewMode === 'board' && (
          <div className="section">
            <label>Group by</label>
            <select value={displayOptions.groupBy || 'column'} onChange={(e) => setDisplayOptions({ groupBy: e.target.value as any })}>
              <option value="column">Status</option>
              <option value="priority">Priority</option>
              <option value="type">Type</option>
              <option value="epic">Epic</option>
            </select>
          </div>
        )}
      </div>
    </div>
  );
}
```

---

### 6.10 Integrate Components and Keyboard Shortcuts

**What**: Wire everything together in the main App component.

**File**: Update `ui/src/App.tsx`

```typescript
import { useState, useEffect } from 'react';
import { useTasks } from './hooks/usePocketBase';
import { useFilteredTasks } from './hooks/useFilteredTasks';
import { useFilterStore } from './stores/filters';
import { Board } from './components/Board';
import { ListView } from './components/ListView';
import { FilterBar } from './components/FilterBar';
import { FilterBuilder } from './components/FilterBuilder';
import { SearchBar } from './components/Search';
import { DisplayOptionsMenu } from './components/DisplayOptions';
import { ViewsSidebar } from './components/ViewsSidebar';

export function App() {
  const { tasks, loading } = useTasks();
  const filteredTasks = useFilteredTasks(tasks);
  const displayOptions = useFilterStore((s) => s.displayOptions);

  const [isFilterBuilderOpen, setIsFilterBuilderOpen] = useState(false);
  const [isDisplayOptionsOpen, setIsDisplayOptionsOpen] = useState(false);
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null);
  const [currentBoardId] = useState<string | null>(null);

  // Keyboard shortcuts
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (['INPUT', 'TEXTAREA'].includes((e.target as HTMLElement).tagName)) return;

      if (e.key === 'f' && !e.metaKey && !e.ctrlKey) {
        e.preventDefault();
        setIsFilterBuilderOpen(true);
      }
      if (e.key === 'b' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        useFilterStore.getState().setDisplayOptions({
          viewMode: displayOptions.viewMode === 'board' ? 'list' : 'board',
        });
      }
      if (e.key === 'Escape') {
        setIsFilterBuilderOpen(false);
        setIsDisplayOptionsOpen(false);
      }
    };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [displayOptions.viewMode]);

  if (loading) return <div>Loading...</div>;

  return (
    <div className="app">
      <aside><ViewsSidebar boardId={currentBoardId} /></aside>
      <main>
        <header>
          <SearchBar />
          <button onClick={() => setIsDisplayOptionsOpen(!isDisplayOptionsOpen)}>Display</button>
        </header>
        <FilterBar totalTasks={tasks.length} filteredTasks={filteredTasks.length} onOpenFilterBuilder={() => setIsFilterBuilderOpen(true)} />
        {displayOptions.viewMode === 'board' 
          ? <Board tasks={filteredTasks} onTaskClick={(t) => setSelectedTaskId(t.id)} selectedTaskId={selectedTaskId} />
          : <ListView tasks={filteredTasks} onTaskClick={(t) => setSelectedTaskId(t.id)} selectedTaskId={selectedTaskId} />
        }
      </main>
      <FilterBuilder isOpen={isFilterBuilderOpen} onClose={() => setIsFilterBuilderOpen(false)} />
      <DisplayOptionsMenu isOpen={isDisplayOptionsOpen} onClose={() => setIsDisplayOptionsOpen(false)} />
    </div>
  );
}
```

---

## Verification Checklist

### Database
- [ ] Views collection exists in PocketBase with fields: name, board, filters, display, is_favorite, match_mode

### Filters
- [ ] Filter by status, priority, type works
- [ ] Filter by labels (includes_any, includes_all) works
- [ ] Multiple filters combine with AND/OR correctly
- [ ] Clear filters shows all tasks

### Search
- [ ] Search filters in real-time
- [ ] Case-insensitive
- [ ] Searches title and description

### Saved Views
- [ ] Create new view from current filters
- [ ] Apply saved view restores filters
- [ ] Favorite/unfavorite views
- [ ] Delete views
- [ ] "Modified" indicator when view changed

### Display
- [ ] `Cmd+B` toggles board/list view
- [ ] Density changes card size
- [ ] Toggle visible fields works

### List View
- [ ] Displays tasks in table
- [ ] Column sorting works
- [ ] Row selection works

### Keyboard Shortcuts
- [ ] `F` opens filter builder
- [ ] `Cmd+B` toggles view
- [ ] `/` focuses search
- [ ] `Esc` closes modals

### Tests
- [ ] `cd ui && npm test` passes

---

## Files Created

| File | Purpose |
|------|---------|
| `migrations/3_views.go` | Views collection migration |
| `ui/src/stores/filters.ts` | Filter state management |
| `ui/src/hooks/useFilteredTasks.ts` | Filter logic |
| `ui/src/hooks/useFilteredTasks.test.ts` | Filter tests |
| `ui/src/hooks/useViews.ts` | Saved views hook |
| `ui/src/components/FilterBar.tsx` | Active filters display |
| `ui/src/components/FilterBuilder.tsx` | Filter creation UI |
| `ui/src/components/Search.tsx` | Search input |
| `ui/src/components/ViewsSidebar.tsx` | Views list in sidebar |
| `ui/src/components/ListView.tsx` | Table view |
| `ui/src/components/DisplayOptions.tsx` | Display settings |
| `ui/src/components/*.css` | Styles for each component |

---

## Next Phase

**Phase 7: Polish** will add:
- Light mode and theme toggle
- Accent color customization
- Animations and transitions
- Responsive/mobile layout
- Toast notifications
- Loading states

---

## Troubleshooting

### Filters not working
1. Check `useFilteredTasks` is used in main component
2. Verify filter state updates (React DevTools)
3. Check console for errors

### Views not persisting
1. Verify views collection exists
2. Check `boardId` is passed to `useViews`
3. Check network tab for errors

### Keyboard shortcuts not working
1. Ensure focus is not in input field
2. Check event listener is registered
3. Verify component is mounted

### Search too slow
1. Add debouncing (300ms)
2. Limit to title only
3. Use virtualization for large lists
