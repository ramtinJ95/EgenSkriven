# Phase 4: Interactive UI

**Goal**: Keyboard-driven UI with command palette, comprehensive shortcuts, and real-time subscriptions.

**Duration Estimate**: 5-7 days

**Prerequisites**: Phase 2 (Minimal UI) complete. The basic board view, task cards, drag-and-drop, and task detail panel must be working.

**Deliverable**: A fully keyboard-navigable UI with command palette, property pickers, real-time updates, and shortcuts help modal.

---

## Overview

Phase 4 transforms the basic UI from Phase 2 into a keyboard-first, Linear-inspired interface. Users can accomplish everything without touching the mouse:

- **Command Palette** (`Cmd+K`) - Quick access to all actions
- **Keyboard Navigation** - Move between tasks and columns with `J/K/H/L`
- **Property Shortcuts** - Change status (`S`), priority (`P`), type (`T`), labels (`L`)
- **Real-time Updates** - CLI changes appear instantly in the UI
- **Peek Preview** - Quick task preview with `Space`

### Why Keyboard-First?

Power users (developers, especially) work faster with keyboards:
- No context-switching between keyboard and mouse
- Muscle memory develops quickly for common actions
- Matches the CLI-first philosophy of EgenSkriven
- Linear, the inspiration for this UI, proves this approach works

### Architecture Overview

```
┌──────────────────────────────────────────────────────────────────────────┐
│                              App.tsx                                     │
│  ┌─────────────────────────────────────────────────────────────────────┐ │
│  │                     KeyboardProvider                                │ │
│  │  ┌──────────────┐ ┌───────────────────┐ ┌────────────────────────┐  │ │
│  │  │ useKeyboard  │ │ SelectionContext  │ │ CommandPaletteContext  │  │ │
│  │  │ (shortcuts)  │ │ (selected tasks)  │ │ (palette state)        │  │ │
│  │  └──────────────┘ └───────────────────┘ └────────────────────────┘  │ │
│  └─────────────────────────────────────────────────────────────────────┘ │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────────┐  │
│  │                         Board.tsx                                  │  │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐      │  │
│  │  │ Column  │ │ Column  │ │ Column  │ │ Column  │ │ Column  │      │  │
│  │  │         │ │         │ │         │ │         │ │         │      │  │
│  │  │ [Card]  │ │ [Card]* │ │ [Card]  │ │ [Card]  │ │ [Card]  │      │  │
│  │  │ [Card]  │ │ [Card]  │ │         │ │         │ │ [Card]  │      │  │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘      │  │
│  └────────────────────────────────────────────────────────────────────┘  │
│                              * = selected                                │
│  ┌────────────────────────────────────────────────────────────────────┐  │
│  │ CommandPalette (modal, triggered by Cmd+K)                         │  │
│  │ PropertyPickers (popovers, triggered by S/P/T/L)                   │  │
│  │ ShortcutsHelp (modal, triggered by ?)                              │  │
│  │ PeekPreview (overlay, triggered by Space)                          │  │
│  └────────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────────┘
```

---

## Environment Requirements

Before starting, ensure you have completed Phase 2 and have:

| Tool | Version | Check Command |
|------|---------|---------------|
| Node.js | 18+ | `node --version` |
| npm | 9+ | `npm --version` |
| Go | 1.21+ | `go version` |

Verify Phase 2 is working:
```bash
cd ui && npm run dev
# In another terminal:
make dev
# Open http://localhost:5173 - board should display with drag-and-drop working
```

---

## Tasks

### 4.1 Create Selection State Management

**What**: Create a React context to track which task(s) are selected.

**Why**: Selection state is needed by multiple components (Board, TaskCard, property pickers, keyboard handler). A context makes this state globally accessible without prop drilling.

**File**: `ui/src/hooks/useSelection.ts`

```typescript
import { createContext, useContext, useState, useCallback, ReactNode } from 'react';

export interface SelectionState {
  // Currently focused/selected task ID (single selection)
  selectedTaskId: string | null;
  
  // Multi-selected task IDs (for bulk operations)
  multiSelectedIds: Set<string>;
  
  // Currently focused column (for keyboard navigation)
  focusedColumn: string | null;
}

export interface SelectionActions {
  // Select a single task (clears multi-selection)
  selectTask: (taskId: string | null) => void;
  
  // Toggle a task in multi-selection
  toggleMultiSelect: (taskId: string) => void;
  
  // Select a range of tasks (Shift+click behavior)
  selectRange: (fromId: string, toId: string, allTaskIds: string[]) => void;
  
  // Select all visible tasks
  selectAll: (taskIds: string[]) => void;
  
  // Clear all selection
  clearSelection: () => void;
  
  // Set focused column (for H/L navigation)
  setFocusedColumn: (column: string | null) => void;
  
  // Check if a task is selected (single or multi)
  isSelected: (taskId: string) => boolean;
}

interface SelectionContextValue extends SelectionState, SelectionActions {}

const SelectionContext = createContext<SelectionContextValue | null>(null);

interface SelectionProviderProps {
  children: ReactNode;
}

export function SelectionProvider({ children }: SelectionProviderProps) {
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null);
  const [multiSelectedIds, setMultiSelectedIds] = useState<Set<string>>(new Set());
  const [focusedColumn, setFocusedColumn] = useState<string | null>(null);

  const selectTask = useCallback((taskId: string | null) => {
    setSelectedTaskId(taskId);
    // Single selection clears multi-selection
    setMultiSelectedIds(new Set());
  }, []);

  const toggleMultiSelect = useCallback((taskId: string) => {
    setMultiSelectedIds(prev => {
      const next = new Set(prev);
      if (next.has(taskId)) {
        next.delete(taskId);
      } else {
        next.add(taskId);
      }
      return next;
    });
  }, []);

  const selectRange = useCallback((fromId: string, toId: string, allTaskIds: string[]) => {
    const fromIndex = allTaskIds.indexOf(fromId);
    const toIndex = allTaskIds.indexOf(toId);
    
    if (fromIndex === -1 || toIndex === -1) return;
    
    const start = Math.min(fromIndex, toIndex);
    const end = Math.max(fromIndex, toIndex);
    
    const rangeIds = allTaskIds.slice(start, end + 1);
    setMultiSelectedIds(new Set(rangeIds));
  }, []);

  const selectAll = useCallback((taskIds: string[]) => {
    setMultiSelectedIds(new Set(taskIds));
  }, []);

  const clearSelection = useCallback(() => {
    setSelectedTaskId(null);
    setMultiSelectedIds(new Set());
  }, []);

  const isSelected = useCallback((taskId: string) => {
    return selectedTaskId === taskId || multiSelectedIds.has(taskId);
  }, [selectedTaskId, multiSelectedIds]);

  const value: SelectionContextValue = {
    selectedTaskId,
    multiSelectedIds,
    focusedColumn,
    selectTask,
    toggleMultiSelect,
    selectRange,
    selectAll,
    clearSelection,
    setFocusedColumn,
    isSelected,
  };

  return (
    <SelectionContext.Provider value={value}>
      {children}
    </SelectionContext.Provider>
  );
}

/**
 * Hook to access selection state and actions.
 * Must be used within a SelectionProvider.
 * 
 * @example
 * function TaskCard({ task }) {
 *   const { isSelected, selectTask } = useSelection();
 *   return (
 *     <div 
 *       className={isSelected(task.id) ? 'selected' : ''}
 *       onClick={() => selectTask(task.id)}
 *     >
 *       {task.title}
 *     </div>
 *   );
 * }
 */
export function useSelection(): SelectionContextValue {
  const context = useContext(SelectionContext);
  if (!context) {
    throw new Error('useSelection must be used within a SelectionProvider');
  }
  return context;
}
```

**Steps**:

1. Create the file:
   ```bash
   touch ui/src/hooks/useSelection.ts
   ```

2. Paste the code above into the file.

3. Verify it compiles:
   ```bash
   cd ui && npm run build
   ```
   
   **Expected**: Build succeeds without errors.

**Key Concepts**:

| Concept | Explanation |
|---------|-------------|
| `createContext` | Creates a React context for sharing state across components |
| `useCallback` | Memoizes functions to prevent unnecessary re-renders |
| `Set<string>` | Efficient data structure for tracking multi-selection |
| Context pattern | Avoids prop drilling through deeply nested components |

---

### 4.2 Create Keyboard Manager Hook

**What**: Create a hook that manages global keyboard shortcuts with proper handling for input fields.

**Why**: We need a centralized system for:
- Registering and unregistering shortcuts
- Preventing shortcuts when typing in inputs
- Supporting key sequences (e.g., `G then B`)
- Providing consistent modifier key handling across platforms

**File**: `ui/src/hooks/useKeyboard.ts`

```typescript
import { useEffect, useCallback, useRef } from 'react';

/**
 * Key combination descriptor.
 * Examples:
 *   { key: 'k', meta: true }        -> Cmd+K (Mac) / Ctrl+K (Windows)
 *   { key: 'Enter' }                -> Enter
 *   { key: 'Backspace', shift: true } -> Shift+Backspace
 */
export interface KeyCombo {
  key: string;
  meta?: boolean;    // Cmd on Mac, Ctrl on Windows
  ctrl?: boolean;    // Ctrl specifically (rarely needed)
  alt?: boolean;     // Option on Mac, Alt on Windows
  shift?: boolean;
}

export interface ShortcutHandler {
  // Key combination to trigger this shortcut
  combo: KeyCombo;
  
  // Handler function - return true to prevent default behavior
  handler: (event: KeyboardEvent) => void | boolean;
  
  // Optional description for help modal
  description?: string;
  
  // If true, shortcut works even when focus is in an input
  allowInInput?: boolean;
  
  // Optional: only active when this condition is true
  when?: () => boolean;
}

/**
 * Check if an element is an input where shortcuts should be disabled.
 */
function isInputElement(element: Element | null): boolean {
  if (!element) return false;
  
  const tagName = element.tagName.toLowerCase();
  
  // Standard input elements
  if (tagName === 'input' || tagName === 'textarea' || tagName === 'select') {
    return true;
  }
  
  // Contenteditable elements
  if (element.getAttribute('contenteditable') === 'true') {
    return true;
  }
  
  // Elements with role="textbox" (for rich text editors)
  if (element.getAttribute('role') === 'textbox') {
    return true;
  }
  
  return false;
}

/**
 * Check if a keyboard event matches a key combination.
 */
function matchesCombo(event: KeyboardEvent, combo: KeyCombo): boolean {
  // Normalize the key for comparison
  const eventKey = event.key.toLowerCase();
  const comboKey = combo.key.toLowerCase();
  
  // Check the key itself
  if (eventKey !== comboKey) {
    return false;
  }
  
  // Check modifiers
  // Note: event.metaKey is Cmd on Mac, we also accept ctrlKey for cross-platform
  const metaOrCtrl = event.metaKey || event.ctrlKey;
  
  if (combo.meta && !metaOrCtrl) return false;
  if (!combo.meta && metaOrCtrl && !combo.ctrl) return false;
  
  if (combo.shift && !event.shiftKey) return false;
  if (!combo.shift && event.shiftKey) return false;
  
  if (combo.alt && !event.altKey) return false;
  if (!combo.alt && event.altKey) return false;
  
  if (combo.ctrl && !event.ctrlKey) return false;
  
  return true;
}

/**
 * Hook for registering keyboard shortcuts.
 * 
 * @example
 * // In a component
 * useKeyboardShortcuts([
 *   {
 *     combo: { key: 'k', meta: true },
 *     handler: () => openCommandPalette(),
 *     description: 'Open command palette',
 *   },
 *   {
 *     combo: { key: 'c' },
 *     handler: () => createTask(),
 *     description: 'Create new task',
 *   },
 * ]);
 */
export function useKeyboardShortcuts(shortcuts: ShortcutHandler[]): void {
  // Use ref to always have latest shortcuts without re-adding listener
  const shortcutsRef = useRef(shortcuts);
  shortcutsRef.current = shortcuts;

  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      const activeElement = document.activeElement;
      const inInput = isInputElement(activeElement);
      
      for (const shortcut of shortcutsRef.current) {
        // Skip if in input and shortcut doesn't allow it
        if (inInput && !shortcut.allowInInput) {
          continue;
        }
        
        // Skip if condition not met
        if (shortcut.when && !shortcut.when()) {
          continue;
        }
        
        // Check if key matches
        if (matchesCombo(event, shortcut.combo)) {
          const result = shortcut.handler(event);
          
          // Prevent default unless handler explicitly returns false
          if (result !== false) {
            event.preventDefault();
            event.stopPropagation();
          }
          
          // Only trigger first matching shortcut
          return;
        }
      }
    }

    // Add listener with capture to handle before other listeners
    document.addEventListener('keydown', handleKeyDown, { capture: true });
    
    return () => {
      document.removeEventListener('keydown', handleKeyDown, { capture: true });
    };
  }, []); // Empty deps - we use ref for shortcuts
}

/**
 * Format a key combination for display in the UI.
 * 
 * @example
 * formatKeyCombo({ key: 'k', meta: true }) // Returns "Cmd+K" on Mac, "Ctrl+K" on Windows
 * formatKeyCombo({ key: 'Enter', shift: true }) // Returns "Shift+Enter"
 */
export function formatKeyCombo(combo: KeyCombo): string {
  const parts: string[] = [];
  
  // Detect platform for correct modifier display
  const isMac = navigator.platform.toLowerCase().includes('mac');
  
  if (combo.meta) {
    parts.push(isMac ? 'Cmd' : 'Ctrl');
  }
  if (combo.ctrl && !combo.meta) {
    parts.push('Ctrl');
  }
  if (combo.alt) {
    parts.push(isMac ? 'Option' : 'Alt');
  }
  if (combo.shift) {
    parts.push('Shift');
  }
  
  // Format key name
  let keyName = combo.key;
  if (keyName === ' ') keyName = 'Space';
  if (keyName === 'Escape') keyName = 'Esc';
  if (keyName === 'Backspace') keyName = isMac ? 'Delete' : 'Backspace';
  if (keyName === 'ArrowUp') keyName = '↑';
  if (keyName === 'ArrowDown') keyName = '↓';
  if (keyName === 'ArrowLeft') keyName = '←';
  if (keyName === 'ArrowRight') keyName = '→';
  
  // Capitalize single letters
  if (keyName.length === 1) {
    keyName = keyName.toUpperCase();
  }
  
  parts.push(keyName);
  
  return parts.join('+');
}
```

**Steps**:

1. Create the file:
   ```bash
   touch ui/src/hooks/useKeyboard.ts
   ```

2. Paste the code above.

3. Verify compilation:
   ```bash
   cd ui && npm run build
   ```

**Common Mistakes**:
- Forgetting to handle both `metaKey` (Mac Cmd) and `ctrlKey` (Windows Ctrl)
- Not preventing default behavior (causes browser shortcuts to fire)
- Not stopping propagation (causes multiple handlers to fire)

---

### 4.3 Create Command Palette Component

**What**: Build the command palette UI that opens with `Cmd+K`.

**Why**: The command palette provides quick access to all actions without memorizing shortcuts. It's the centerpiece of keyboard-first UIs.

**File**: `ui/src/components/CommandPalette.tsx`

```tsx
import { useState, useEffect, useMemo, useCallback, useRef } from 'react';
import { createPortal } from 'react-dom';
import { useKeyboardShortcuts, formatKeyCombo, KeyCombo } from '../hooks/useKeyboard';
import { useSelection } from '../hooks/useSelection';
import { useTasks } from '../hooks/useTasks';
import styles from './CommandPalette.module.css';

export interface Command {
  id: string;
  label: string;
  shortcut?: KeyCombo;
  section: 'actions' | 'navigation' | 'recent';
  icon?: string;
  action: () => void;
  // Optional: only show when condition is true
  when?: () => boolean;
}

interface CommandPaletteProps {
  isOpen: boolean;
  onClose: () => void;
  commands: Command[];
}

/**
 * Fuzzy match a query against a string.
 * Returns true if all characters in query appear in order in string.
 */
function fuzzyMatch(query: string, text: string): boolean {
  const queryLower = query.toLowerCase();
  const textLower = text.toLowerCase();
  
  let queryIndex = 0;
  
  for (let i = 0; i < textLower.length && queryIndex < queryLower.length; i++) {
    if (textLower[i] === queryLower[queryIndex]) {
      queryIndex++;
    }
  }
  
  return queryIndex === queryLower.length;
}

/**
 * Score a fuzzy match - higher is better.
 * Prefers matches at word boundaries and consecutive matches.
 */
function fuzzyScore(query: string, text: string): number {
  const queryLower = query.toLowerCase();
  const textLower = text.toLowerCase();
  
  let score = 0;
  let queryIndex = 0;
  let consecutiveMatches = 0;
  
  for (let i = 0; i < textLower.length && queryIndex < queryLower.length; i++) {
    if (textLower[i] === queryLower[queryIndex]) {
      // Bonus for match at start
      if (i === 0) score += 10;
      
      // Bonus for match after word boundary
      if (i > 0 && /\s/.test(text[i - 1])) score += 5;
      
      // Bonus for consecutive matches
      consecutiveMatches++;
      score += consecutiveMatches * 2;
      
      queryIndex++;
    } else {
      consecutiveMatches = 0;
    }
  }
  
  // Penalty for longer strings (prefer shorter matches)
  score -= text.length * 0.1;
  
  return score;
}

export function CommandPalette({ isOpen, onClose, commands }: CommandPaletteProps) {
  const [query, setQuery] = useState('');
  const [selectedIndex, setSelectedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLDivElement>(null);
  
  // Filter and sort commands based on query
  const filteredCommands = useMemo(() => {
    // Filter out commands that don't pass their condition
    let available = commands.filter(cmd => !cmd.when || cmd.when());
    
    if (!query) {
      return available;
    }
    
    // Filter by fuzzy match
    const matched = available.filter(cmd => fuzzyMatch(query, cmd.label));
    
    // Sort by score (best match first)
    matched.sort((a, b) => fuzzyScore(query, b.label) - fuzzyScore(query, a.label));
    
    return matched;
  }, [commands, query]);
  
  // Group commands by section
  const groupedCommands = useMemo(() => {
    const groups: Record<string, Command[]> = {
      actions: [],
      navigation: [],
      recent: [],
    };
    
    for (const cmd of filteredCommands) {
      groups[cmd.section]?.push(cmd);
    }
    
    return groups;
  }, [filteredCommands]);
  
  // Reset state when opened
  useEffect(() => {
    if (isOpen) {
      setQuery('');
      setSelectedIndex(0);
      // Focus input after a tick (portal needs to render first)
      setTimeout(() => inputRef.current?.focus(), 0);
    }
  }, [isOpen]);
  
  // Keep selected index in bounds
  useEffect(() => {
    if (selectedIndex >= filteredCommands.length) {
      setSelectedIndex(Math.max(0, filteredCommands.length - 1));
    }
  }, [filteredCommands.length, selectedIndex]);
  
  // Scroll selected item into view
  useEffect(() => {
    const selectedElement = listRef.current?.querySelector(`[data-index="${selectedIndex}"]`);
    selectedElement?.scrollIntoView({ block: 'nearest' });
  }, [selectedIndex]);
  
  const executeCommand = useCallback((command: Command) => {
    onClose();
    // Execute after close animation
    setTimeout(() => command.action(), 50);
  }, [onClose]);
  
  // Handle keyboard navigation within palette
  const handleKeyDown = useCallback((event: React.KeyboardEvent) => {
    switch (event.key) {
      case 'ArrowDown':
        event.preventDefault();
        setSelectedIndex(i => Math.min(i + 1, filteredCommands.length - 1));
        break;
        
      case 'ArrowUp':
        event.preventDefault();
        setSelectedIndex(i => Math.max(i - 1, 0));
        break;
        
      case 'Enter':
        event.preventDefault();
        if (filteredCommands[selectedIndex]) {
          executeCommand(filteredCommands[selectedIndex]);
        }
        break;
        
      case 'Escape':
        event.preventDefault();
        onClose();
        break;
    }
  }, [filteredCommands, selectedIndex, executeCommand, onClose]);
  
  if (!isOpen) return null;
  
  // Get flat index for a command (for keyboard navigation)
  let flatIndex = 0;
  const getFlatIndex = () => flatIndex++;
  
  const palette = (
    <div className={styles.overlay} onClick={onClose}>
      <div 
        className={styles.palette} 
        onClick={e => e.stopPropagation()}
        onKeyDown={handleKeyDown}
      >
        {/* Search input */}
        <div className={styles.inputWrapper}>
          <span className={styles.searchIcon}>
            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
              <path d="M11.742 10.344a6.5 6.5 0 1 0-1.397 1.398h-.001c.03.04.062.078.098.115l3.85 3.85a1 1 0 0 0 1.415-1.414l-3.85-3.85a1.007 1.007 0 0 0-.115-.1zM12 6.5a5.5 5.5 0 1 1-11 0 5.5 5.5 0 0 1 11 0z"/>
            </svg>
          </span>
          <input
            ref={inputRef}
            type="text"
            className={styles.input}
            placeholder="Type a command or search..."
            value={query}
            onChange={e => {
              setQuery(e.target.value);
              setSelectedIndex(0);
            }}
          />
        </div>
        
        {/* Command list */}
        <div className={styles.list} ref={listRef}>
          {filteredCommands.length === 0 ? (
            <div className={styles.empty}>No commands found</div>
          ) : (
            <>
              {/* Actions section */}
              {groupedCommands.actions.length > 0 && (
                <div className={styles.section}>
                  <div className={styles.sectionTitle}>ACTIONS</div>
                  {groupedCommands.actions.map(cmd => {
                    const index = getFlatIndex();
                    return (
                      <button
                        key={cmd.id}
                        data-index={index}
                        className={`${styles.item} ${index === selectedIndex ? styles.selected : ''}`}
                        onClick={() => executeCommand(cmd)}
                        onMouseEnter={() => setSelectedIndex(index)}
                      >
                        <span className={styles.icon}>{cmd.icon || '●'}</span>
                        <span className={styles.label}>{cmd.label}</span>
                        {cmd.shortcut && (
                          <span className={styles.shortcut}>
                            {formatKeyCombo(cmd.shortcut)}
                          </span>
                        )}
                      </button>
                    );
                  })}
                </div>
              )}
              
              {/* Navigation section */}
              {groupedCommands.navigation.length > 0 && (
                <div className={styles.section}>
                  <div className={styles.sectionTitle}>NAVIGATION</div>
                  {groupedCommands.navigation.map(cmd => {
                    const index = getFlatIndex();
                    return (
                      <button
                        key={cmd.id}
                        data-index={index}
                        className={`${styles.item} ${index === selectedIndex ? styles.selected : ''}`}
                        onClick={() => executeCommand(cmd)}
                        onMouseEnter={() => setSelectedIndex(index)}
                      >
                        <span className={styles.icon}>{cmd.icon || '→'}</span>
                        <span className={styles.label}>{cmd.label}</span>
                        {cmd.shortcut && (
                          <span className={styles.shortcut}>
                            {formatKeyCombo(cmd.shortcut)}
                          </span>
                        )}
                      </button>
                    );
                  })}
                </div>
              )}
              
              {/* Recent section */}
              {groupedCommands.recent.length > 0 && (
                <div className={styles.section}>
                  <div className={styles.sectionTitle}>RECENT TASKS</div>
                  {groupedCommands.recent.map(cmd => {
                    const index = getFlatIndex();
                    return (
                      <button
                        key={cmd.id}
                        data-index={index}
                        className={`${styles.item} ${index === selectedIndex ? styles.selected : ''}`}
                        onClick={() => executeCommand(cmd)}
                        onMouseEnter={() => setSelectedIndex(index)}
                      >
                        <span className={styles.label}>{cmd.label}</span>
                      </button>
                    );
                  })}
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
  
  // Use portal to render at document root (above all other content)
  return createPortal(palette, document.body);
}
```

**File**: `ui/src/components/CommandPalette.module.css`

```css
/* Backdrop overlay */
.overlay {
  position: fixed;
  inset: 0;
  background: var(--bg-overlay, rgba(0, 0, 0, 0.6));
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding-top: 15vh;
  z-index: 1000;
  animation: fadeIn 100ms ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

/* Main palette container */
.palette {
  width: 560px;
  max-height: 400px;
  background: var(--bg-card, #1A1A1A);
  border-radius: var(--radius-lg, 8px);
  border: 1px solid var(--border-default, #333333);
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.4);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  animation: slideIn 150ms ease-out;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: scale(0.95) translateY(-10px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

/* Search input area */
.inputWrapper {
  display: flex;
  align-items: center;
  padding: var(--space-3, 12px) var(--space-4, 16px);
  border-bottom: 1px solid var(--border-subtle, #2A2A2A);
}

.searchIcon {
  color: var(--text-muted, #666666);
  margin-right: var(--space-3, 12px);
  display: flex;
}

.input {
  flex: 1;
  background: transparent;
  border: none;
  color: var(--text-primary, #F5F5F5);
  font-size: var(--text-base, 13px);
  outline: none;
}

.input::placeholder {
  color: var(--text-muted, #666666);
}

/* Command list */
.list {
  flex: 1;
  overflow-y: auto;
  padding: var(--space-2, 8px);
}

.empty {
  padding: var(--space-6, 24px);
  text-align: center;
  color: var(--text-muted, #666666);
}

/* Section */
.section {
  margin-bottom: var(--space-2, 8px);
}

.sectionTitle {
  padding: var(--space-2, 8px) var(--space-3, 12px);
  font-size: var(--text-xs, 11px);
  font-weight: var(--font-semibold, 600);
  color: var(--text-muted, #666666);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

/* Command item */
.item {
  width: 100%;
  display: flex;
  align-items: center;
  padding: var(--space-2, 8px) var(--space-3, 12px);
  background: transparent;
  border: none;
  border-radius: var(--radius-sm, 4px);
  color: var(--text-primary, #F5F5F5);
  font-size: var(--text-base, 13px);
  text-align: left;
  cursor: pointer;
  transition: background-color 100ms;
}

.item:hover,
.item.selected {
  background: var(--bg-card-hover, #252525);
}

.icon {
  margin-right: var(--space-3, 12px);
  color: var(--text-secondary, #A0A0A0);
  font-size: var(--text-sm, 12px);
}

.label {
  flex: 1;
}

.shortcut {
  margin-left: var(--space-3, 12px);
  padding: var(--space-1, 4px) var(--space-2, 8px);
  background: var(--bg-input, #1F1F1F);
  border-radius: var(--radius-sm, 4px);
  font-size: var(--text-xs, 11px);
  font-family: var(--font-mono, monospace);
  color: var(--text-muted, #666666);
}
```

**Steps**:

1. Create the component file:
   ```bash
   touch ui/src/components/CommandPalette.tsx
   touch ui/src/components/CommandPalette.module.css
   ```

2. Paste the code into each file.

3. Verify compilation:
   ```bash
   cd ui && npm run build
   ```

**Key Features**:
- Fuzzy search with scoring (prefers matches at word boundaries)
- Keyboard navigation (`↑/↓`, `Enter`, `Escape`)
- Grouped sections (Actions, Navigation, Recent)
- Visual feedback for selected item
- Portal rendering (appears above all content)
- Smooth animations

---

### 4.4 Create Property Picker Components

**What**: Build popover components for changing task properties (status, priority, type, labels).

**Why**: Property pickers are triggered by keyboard shortcuts (`S`, `P`, `T`, `L`) when a task is selected. They provide quick property changes without opening the full detail panel.

**File**: `ui/src/components/PropertyPicker.tsx`

```tsx
import { useState, useEffect, useRef, useCallback, ReactNode } from 'react';
import { createPortal } from 'react-dom';
import styles from './PropertyPicker.module.css';

export interface PropertyOption<T> {
  value: T;
  label: string;
  icon?: ReactNode;
  color?: string;
}

interface PropertyPickerProps<T> {
  isOpen: boolean;
  onClose: () => void;
  onSelect: (value: T) => void;
  options: PropertyOption<T>[];
  currentValue?: T;
  title: string;
  // Position the picker near this element
  anchorElement?: HTMLElement | null;
}

export function PropertyPicker<T extends string>({
  isOpen,
  onClose,
  onSelect,
  options,
  currentValue,
  title,
  anchorElement,
}: PropertyPickerProps<T>) {
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [query, setQuery] = useState('');
  const inputRef = useRef<HTMLInputElement>(null);
  const pickerRef = useRef<HTMLDivElement>(null);
  
  // Filter options by query
  const filteredOptions = query
    ? options.filter(opt => 
        opt.label.toLowerCase().includes(query.toLowerCase())
      )
    : options;
  
  // Reset on open
  useEffect(() => {
    if (isOpen) {
      setQuery('');
      // Set initial selection to current value
      const currentIndex = options.findIndex(opt => opt.value === currentValue);
      setSelectedIndex(currentIndex >= 0 ? currentIndex : 0);
      setTimeout(() => inputRef.current?.focus(), 0);
    }
  }, [isOpen, currentValue, options]);
  
  // Keep selection in bounds
  useEffect(() => {
    if (selectedIndex >= filteredOptions.length) {
      setSelectedIndex(Math.max(0, filteredOptions.length - 1));
    }
  }, [filteredOptions.length, selectedIndex]);
  
  // Position the picker near the anchor element
  useEffect(() => {
    if (!isOpen || !anchorElement || !pickerRef.current) return;
    
    const anchorRect = anchorElement.getBoundingClientRect();
    const picker = pickerRef.current;
    
    // Position below the anchor, centered
    picker.style.top = `${anchorRect.bottom + 8}px`;
    picker.style.left = `${anchorRect.left + anchorRect.width / 2 - picker.offsetWidth / 2}px`;
    
    // Ensure it stays on screen
    const pickerRect = picker.getBoundingClientRect();
    if (pickerRect.right > window.innerWidth - 16) {
      picker.style.left = `${window.innerWidth - pickerRect.width - 16}px`;
    }
    if (pickerRect.left < 16) {
      picker.style.left = '16px';
    }
  }, [isOpen, anchorElement]);
  
  const handleSelect = useCallback((option: PropertyOption<T>) => {
    onSelect(option.value);
    onClose();
  }, [onSelect, onClose]);
  
  const handleKeyDown = useCallback((event: React.KeyboardEvent) => {
    switch (event.key) {
      case 'ArrowDown':
        event.preventDefault();
        setSelectedIndex(i => Math.min(i + 1, filteredOptions.length - 1));
        break;
        
      case 'ArrowUp':
        event.preventDefault();
        setSelectedIndex(i => Math.max(i - 1, 0));
        break;
        
      case 'Enter':
        event.preventDefault();
        if (filteredOptions[selectedIndex]) {
          handleSelect(filteredOptions[selectedIndex]);
        }
        break;
        
      case 'Escape':
        event.preventDefault();
        onClose();
        break;
    }
  }, [filteredOptions, selectedIndex, handleSelect, onClose]);
  
  if (!isOpen) return null;
  
  const picker = (
    <div className={styles.overlay} onClick={onClose}>
      <div 
        ref={pickerRef}
        className={styles.picker}
        onClick={e => e.stopPropagation()}
        onKeyDown={handleKeyDown}
      >
        <div className={styles.header}>
          <span className={styles.title}>{title}</span>
        </div>
        
        <div className={styles.inputWrapper}>
          <input
            ref={inputRef}
            type="text"
            className={styles.input}
            placeholder="Filter..."
            value={query}
            onChange={e => {
              setQuery(e.target.value);
              setSelectedIndex(0);
            }}
          />
        </div>
        
        <div className={styles.options}>
          {filteredOptions.length === 0 ? (
            <div className={styles.empty}>No options found</div>
          ) : (
            filteredOptions.map((option, index) => (
              <button
                key={option.value}
                className={`${styles.option} ${index === selectedIndex ? styles.selected : ''}`}
                onClick={() => handleSelect(option)}
                onMouseEnter={() => setSelectedIndex(index)}
              >
                {option.icon && (
                  <span 
                    className={styles.icon}
                    style={option.color ? { color: option.color } : undefined}
                  >
                    {option.icon}
                  </span>
                )}
                <span className={styles.label}>{option.label}</span>
                {option.value === currentValue && (
                  <span className={styles.check}>✓</span>
                )}
              </button>
            ))
          )}
        </div>
      </div>
    </div>
  );
  
  return createPortal(picker, document.body);
}

// Pre-configured pickers for common properties

export const STATUS_OPTIONS: PropertyOption<string>[] = [
  { value: 'backlog', label: 'Backlog', icon: '●', color: 'var(--status-backlog, #6B7280)' },
  { value: 'todo', label: 'Todo', icon: '●', color: 'var(--status-todo, #E5E5E5)' },
  { value: 'in_progress', label: 'In Progress', icon: '●', color: 'var(--status-in-progress, #F59E0B)' },
  { value: 'review', label: 'Review', icon: '●', color: 'var(--status-review, #A855F7)' },
  { value: 'done', label: 'Done', icon: '●', color: 'var(--status-done, #22C55E)' },
];

export const PRIORITY_OPTIONS: PropertyOption<string>[] = [
  { value: 'urgent', label: 'Urgent', icon: '🔴', color: 'var(--priority-urgent, #EF4444)' },
  { value: 'high', label: 'High', icon: '🟠', color: 'var(--priority-high, #F97316)' },
  { value: 'medium', label: 'Medium', icon: '🟡', color: 'var(--priority-medium, #EAB308)' },
  { value: 'low', label: 'Low', icon: '⚪', color: 'var(--priority-low, #6B7280)' },
];

export const TYPE_OPTIONS: PropertyOption<string>[] = [
  { value: 'bug', label: 'Bug', icon: '🐛', color: 'var(--type-bug, #EF4444)' },
  { value: 'feature', label: 'Feature', icon: '✨', color: 'var(--type-feature, #A855F7)' },
  { value: 'chore', label: 'Chore', icon: '🔧', color: 'var(--type-chore, #6B7280)' },
];
```

**File**: `ui/src/components/PropertyPicker.module.css`

```css
.overlay {
  position: fixed;
  inset: 0;
  z-index: 1001;
}

.picker {
  position: fixed;
  width: 220px;
  max-height: 320px;
  background: var(--bg-card, #1A1A1A);
  border-radius: var(--radius-md, 6px);
  border: 1px solid var(--border-default, #333333);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  animation: popIn 100ms ease-out;
}

@keyframes popIn {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

.header {
  padding: var(--space-2, 8px) var(--space-3, 12px);
  border-bottom: 1px solid var(--border-subtle, #2A2A2A);
}

.title {
  font-size: var(--text-xs, 11px);
  font-weight: var(--font-semibold, 600);
  color: var(--text-muted, #666666);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.inputWrapper {
  padding: var(--space-2, 8px);
  border-bottom: 1px solid var(--border-subtle, #2A2A2A);
}

.input {
  width: 100%;
  padding: var(--space-2, 8px);
  background: var(--bg-input, #1F1F1F);
  border: 1px solid var(--border-subtle, #2A2A2A);
  border-radius: var(--radius-sm, 4px);
  color: var(--text-primary, #F5F5F5);
  font-size: var(--text-sm, 12px);
  outline: none;
}

.input:focus {
  border-color: var(--border-focus, var(--accent, #5E6AD2));
}

.input::placeholder {
  color: var(--text-disabled, #444444);
}

.options {
  flex: 1;
  overflow-y: auto;
  padding: var(--space-1, 4px);
}

.empty {
  padding: var(--space-4, 16px);
  text-align: center;
  color: var(--text-muted, #666666);
  font-size: var(--text-sm, 12px);
}

.option {
  width: 100%;
  display: flex;
  align-items: center;
  padding: var(--space-2, 8px) var(--space-3, 12px);
  background: transparent;
  border: none;
  border-radius: var(--radius-sm, 4px);
  color: var(--text-primary, #F5F5F5);
  font-size: var(--text-sm, 12px);
  text-align: left;
  cursor: pointer;
  transition: background-color 100ms;
}

.option:hover,
.option.selected {
  background: var(--bg-card-hover, #252525);
}

.icon {
  margin-right: var(--space-2, 8px);
  font-size: var(--text-base, 13px);
}

.label {
  flex: 1;
}

.check {
  color: var(--accent, #5E6AD2);
  font-size: var(--text-xs, 11px);
}
```

**Steps**:

1. Create the files:
   ```bash
   touch ui/src/components/PropertyPicker.tsx
   touch ui/src/components/PropertyPicker.module.css
   ```

2. Paste the code into each file.

3. Verify compilation:
   ```bash
   cd ui && npm run build
   ```

---

### 4.5 Create Shortcuts Help Modal

**What**: Build a modal that displays all available keyboard shortcuts, triggered by `?`.

**Why**: Users need a way to discover and remember shortcuts. The help modal serves as a quick reference.

**File**: `ui/src/components/ShortcutsHelp.tsx`

```tsx
import { createPortal } from 'react-dom';
import { formatKeyCombo, KeyCombo } from '../hooks/useKeyboard';
import styles from './ShortcutsHelp.module.css';

interface ShortcutGroup {
  title: string;
  shortcuts: Array<{
    combo: KeyCombo;
    description: string;
  }>;
}

interface ShortcutsHelpProps {
  isOpen: boolean;
  onClose: () => void;
}

const SHORTCUT_GROUPS: ShortcutGroup[] = [
  {
    title: 'Global',
    shortcuts: [
      { combo: { key: 'k', meta: true }, description: 'Command palette' },
      { combo: { key: '/' }, description: 'Quick search' },
      { combo: { key: 'b', meta: true }, description: 'Toggle board/list view' },
      { combo: { key: '\\', meta: true }, description: 'Toggle sidebar' },
      { combo: { key: '?' }, description: 'Show shortcuts help' },
    ],
  },
  {
    title: 'Task Actions',
    shortcuts: [
      { combo: { key: 'c' }, description: 'Create new task' },
      { combo: { key: 'Enter' }, description: 'Open selected task' },
      { combo: { key: ' ' }, description: 'Peek preview' },
      { combo: { key: 'e' }, description: 'Edit title' },
      { combo: { key: 'Backspace' }, description: 'Delete task' },
    ],
  },
  {
    title: 'Task Properties',
    shortcuts: [
      { combo: { key: 's' }, description: 'Set status' },
      { combo: { key: 'p' }, description: 'Set priority' },
      { combo: { key: 't' }, description: 'Set type' },
      { combo: { key: 'l' }, description: 'Manage labels' },
      { combo: { key: 'd' }, description: 'Set due date' },
    ],
  },
  {
    title: 'Navigation',
    shortcuts: [
      { combo: { key: 'j' }, description: 'Next task' },
      { combo: { key: 'k' }, description: 'Previous task' },
      { combo: { key: 'h' }, description: 'Previous column' },
      { combo: { key: 'l' }, description: 'Next column' },
      { combo: { key: 'ArrowDown' }, description: 'Next task (arrow)' },
      { combo: { key: 'ArrowUp' }, description: 'Previous task (arrow)' },
      { combo: { key: 'Escape' }, description: 'Close panel / deselect' },
    ],
  },
  {
    title: 'Selection',
    shortcuts: [
      { combo: { key: 'x' }, description: 'Toggle select task' },
      { combo: { key: 'x', shift: true }, description: 'Select range' },
      { combo: { key: 'a', meta: true }, description: 'Select all visible' },
    ],
  },
];

export function ShortcutsHelp({ isOpen, onClose }: ShortcutsHelpProps) {
  if (!isOpen) return null;
  
  const modal = (
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.modal} onClick={e => e.stopPropagation()}>
        <div className={styles.header}>
          <h2 className={styles.title}>Keyboard Shortcuts</h2>
          <button className={styles.closeButton} onClick={onClose}>
            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
              <path d="M4.646 4.646a.5.5 0 0 1 .708 0L8 7.293l2.646-2.647a.5.5 0 0 1 .708.708L8.707 8l2.647 2.646a.5.5 0 0 1-.708.708L8 8.707l-2.646 2.647a.5.5 0 0 1-.708-.708L7.293 8 4.646 5.354a.5.5 0 0 1 0-.708z"/>
            </svg>
          </button>
        </div>
        
        <div className={styles.content}>
          {SHORTCUT_GROUPS.map(group => (
            <div key={group.title} className={styles.group}>
              <h3 className={styles.groupTitle}>{group.title}</h3>
              <div className={styles.shortcuts}>
                {group.shortcuts.map((shortcut, index) => (
                  <div key={index} className={styles.shortcut}>
                    <span className={styles.description}>{shortcut.description}</span>
                    <kbd className={styles.key}>{formatKeyCombo(shortcut.combo)}</kbd>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
        
        <div className={styles.footer}>
          <span className={styles.hint}>Press <kbd>Esc</kbd> to close</span>
        </div>
      </div>
    </div>
  );
  
  return createPortal(modal, document.body);
}
```

**File**: `ui/src/components/ShortcutsHelp.module.css`

```css
.overlay {
  position: fixed;
  inset: 0;
  background: var(--bg-overlay, rgba(0, 0, 0, 0.6));
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  animation: fadeIn 100ms ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.modal {
  width: 640px;
  max-height: 80vh;
  background: var(--bg-card, #1A1A1A);
  border-radius: var(--radius-lg, 8px);
  border: 1px solid var(--border-default, #333333);
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.4);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  animation: slideIn 150ms ease-out;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-4, 16px) var(--space-5, 20px);
  border-bottom: 1px solid var(--border-subtle, #2A2A2A);
}

.title {
  margin: 0;
  font-size: var(--text-lg, 14px);
  font-weight: var(--font-semibold, 600);
  color: var(--text-primary, #F5F5F5);
}

.closeButton {
  padding: var(--space-1, 4px);
  background: transparent;
  border: none;
  border-radius: var(--radius-sm, 4px);
  color: var(--text-muted, #666666);
  cursor: pointer;
  transition: color 100ms, background-color 100ms;
}

.closeButton:hover {
  color: var(--text-primary, #F5F5F5);
  background: var(--bg-card-hover, #252525);
}

.content {
  flex: 1;
  overflow-y: auto;
  padding: var(--space-4, 16px) var(--space-5, 20px);
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--space-5, 20px);
}

.group {
  display: flex;
  flex-direction: column;
}

.groupTitle {
  margin: 0 0 var(--space-3, 12px) 0;
  font-size: var(--text-xs, 11px);
  font-weight: var(--font-semibold, 600);
  color: var(--text-muted, #666666);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.shortcuts {
  display: flex;
  flex-direction: column;
  gap: var(--space-2, 8px);
}

.shortcut {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.description {
  font-size: var(--text-sm, 12px);
  color: var(--text-secondary, #A0A0A0);
}

.key {
  padding: var(--space-1, 4px) var(--space-2, 8px);
  background: var(--bg-input, #1F1F1F);
  border-radius: var(--radius-sm, 4px);
  font-size: var(--text-xs, 11px);
  font-family: var(--font-mono, monospace);
  color: var(--text-muted, #666666);
}

.footer {
  padding: var(--space-3, 12px) var(--space-5, 20px);
  border-top: 1px solid var(--border-subtle, #2A2A2A);
  text-align: center;
}

.hint {
  font-size: var(--text-xs, 11px);
  color: var(--text-muted, #666666);
}

.hint kbd {
  padding: var(--space-1, 4px) var(--space-2, 8px);
  background: var(--bg-input, #1F1F1F);
  border-radius: var(--radius-sm, 4px);
  font-family: var(--font-mono, monospace);
}
```

**Steps**:

1. Create the files:
   ```bash
   touch ui/src/components/ShortcutsHelp.tsx
   touch ui/src/components/ShortcutsHelp.module.css
   ```

2. Paste the code into each file.

3. Verify compilation:
   ```bash
   cd ui && npm run build
   ```

---

### 4.6 Create Peek Preview Component

**What**: Build an overlay that shows a quick preview of a task when `Space` is pressed.

**Why**: Peek preview allows users to glance at task details without fully opening the detail panel. Press Space again or Escape to dismiss.

**File**: `ui/src/components/PeekPreview.tsx`

```tsx
import { createPortal } from 'react-dom';
import { Task } from '../types/task';
import styles from './PeekPreview.module.css';

interface PeekPreviewProps {
  task: Task | null;
  isOpen: boolean;
  onClose: () => void;
}

export function PeekPreview({ task, isOpen, onClose }: PeekPreviewProps) {
  if (!isOpen || !task) return null;
  
  const preview = (
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.preview} onClick={e => e.stopPropagation()}>
        {/* Header with task ID and type */}
        <div className={styles.header}>
          <span className={styles.taskId}>{task.id.substring(0, 8)}</span>
          <span className={`${styles.type} ${styles[task.type]}`}>
            {task.type}
          </span>
        </div>
        
        {/* Title */}
        <h2 className={styles.title}>{task.title}</h2>
        
        {/* Properties row */}
        <div className={styles.properties}>
          <div className={styles.property}>
            <span className={styles.propertyLabel}>Status</span>
            <span className={`${styles.status} ${styles[task.column]}`}>
              {formatColumn(task.column)}
            </span>
          </div>
          
          <div className={styles.property}>
            <span className={styles.propertyLabel}>Priority</span>
            <span className={`${styles.priority} ${styles[task.priority]}`}>
              {task.priority}
            </span>
          </div>
          
          {task.labels && task.labels.length > 0 && (
            <div className={styles.property}>
              <span className={styles.propertyLabel}>Labels</span>
              <div className={styles.labels}>
                {task.labels.map(label => (
                  <span key={label} className={styles.label}>{label}</span>
                ))}
              </div>
            </div>
          )}
        </div>
        
        {/* Description preview */}
        {task.description && (
          <div className={styles.description}>
            <span className={styles.propertyLabel}>Description</span>
            <p className={styles.descriptionText}>
              {task.description.length > 200 
                ? task.description.substring(0, 200) + '...'
                : task.description
              }
            </p>
          </div>
        )}
        
        {/* Footer hint */}
        <div className={styles.footer}>
          <span className={styles.hint}>Press <kbd>Enter</kbd> to open full details</span>
        </div>
      </div>
    </div>
  );
  
  return createPortal(preview, document.body);
}

function formatColumn(column: string): string {
  return column.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
}
```

**File**: `ui/src/components/PeekPreview.module.css`

```css
.overlay {
  position: fixed;
  inset: 0;
  background: var(--bg-overlay, rgba(0, 0, 0, 0.4));
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 999;
  animation: fadeIn 100ms ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.preview {
  width: 480px;
  max-height: 400px;
  background: var(--bg-card, #1A1A1A);
  border-radius: var(--radius-lg, 8px);
  border: 1px solid var(--border-default, #333333);
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.4);
  overflow: hidden;
  animation: popIn 150ms ease-out;
}

@keyframes popIn {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

.header {
  display: flex;
  align-items: center;
  gap: var(--space-2, 8px);
  padding: var(--space-3, 12px) var(--space-4, 16px);
  border-bottom: 1px solid var(--border-subtle, #2A2A2A);
}

.taskId {
  font-size: var(--text-xs, 11px);
  font-family: var(--font-mono, monospace);
  color: var(--text-muted, #666666);
}

.type {
  padding: var(--space-1, 4px) var(--space-2, 8px);
  border-radius: var(--radius-sm, 4px);
  font-size: var(--text-xs, 11px);
  font-weight: var(--font-medium, 500);
}

.type.bug {
  background: rgba(239, 68, 68, 0.2);
  color: var(--type-bug, #EF4444);
}

.type.feature {
  background: rgba(168, 85, 247, 0.2);
  color: var(--type-feature, #A855F7);
}

.type.chore {
  background: rgba(107, 114, 128, 0.2);
  color: var(--type-chore, #6B7280);
}

.title {
  margin: 0;
  padding: var(--space-3, 12px) var(--space-4, 16px);
  font-size: var(--text-lg, 14px);
  font-weight: var(--font-semibold, 600);
  color: var(--text-primary, #F5F5F5);
  line-height: var(--leading-tight, 1.25);
}

.properties {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-4, 16px);
  padding: 0 var(--space-4, 16px) var(--space-3, 12px);
}

.property {
  display: flex;
  flex-direction: column;
  gap: var(--space-1, 4px);
}

.propertyLabel {
  font-size: var(--text-xs, 11px);
  color: var(--text-muted, #666666);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.status {
  font-size: var(--text-sm, 12px);
  font-weight: var(--font-medium, 500);
}

.status.backlog { color: var(--status-backlog, #6B7280); }
.status.todo { color: var(--status-todo, #E5E5E5); }
.status.in_progress { color: var(--status-in-progress, #F59E0B); }
.status.review { color: var(--status-review, #A855F7); }
.status.done { color: var(--status-done, #22C55E); }

.priority {
  font-size: var(--text-sm, 12px);
  font-weight: var(--font-medium, 500);
  text-transform: capitalize;
}

.priority.urgent { color: var(--priority-urgent, #EF4444); }
.priority.high { color: var(--priority-high, #F97316); }
.priority.medium { color: var(--priority-medium, #EAB308); }
.priority.low { color: var(--priority-low, #6B7280); }

.labels {
  display: flex;
  gap: var(--space-1, 4px);
}

.label {
  padding: var(--space-1, 4px) var(--space-2, 8px);
  background: var(--bg-input, #1F1F1F);
  border-radius: var(--radius-sm, 4px);
  font-size: var(--text-xs, 11px);
  color: var(--text-secondary, #A0A0A0);
}

.description {
  padding: 0 var(--space-4, 16px) var(--space-3, 12px);
}

.descriptionText {
  margin: var(--space-1, 4px) 0 0 0;
  font-size: var(--text-sm, 12px);
  color: var(--text-secondary, #A0A0A0);
  line-height: var(--leading-relaxed, 1.625);
}

.footer {
  padding: var(--space-3, 12px) var(--space-4, 16px);
  border-top: 1px solid var(--border-subtle, #2A2A2A);
  text-align: center;
}

.hint {
  font-size: var(--text-xs, 11px);
  color: var(--text-muted, #666666);
}

.hint kbd {
  padding: var(--space-1, 4px) var(--space-2, 8px);
  background: var(--bg-input, #1F1F1F);
  border-radius: var(--radius-sm, 4px);
  font-family: var(--font-mono, monospace);
}
```

**Steps**:

1. Create the files:
   ```bash
   touch ui/src/components/PeekPreview.tsx
   touch ui/src/components/PeekPreview.module.css
   ```

2. Paste the code into each file.

3. Verify compilation:
   ```bash
   cd ui && npm run build
   ```

---

### 4.7 Update useTasks Hook with Real-time Subscriptions

**What**: Enhance the existing `useTasks` hook to subscribe to real-time updates from PocketBase.

**Why**: When tasks are created, updated, or deleted via CLI, the UI should update immediately without requiring a refresh.

**File**: Update `ui/src/hooks/useTasks.ts`

```typescript
import { useEffect, useState, useCallback } from 'react';
import { pb } from '../lib/pb';
import { Task } from '../types/task';

export interface UseTasksResult {
  tasks: Task[];
  loading: boolean;
  error: Error | null;
  
  // CRUD operations
  createTask: (task: Partial<Task>) => Promise<Task>;
  updateTask: (id: string, updates: Partial<Task>) => Promise<Task>;
  deleteTask: (id: string) => Promise<void>;
  moveTask: (id: string, column: string, position: number) => Promise<Task>;
  
  // Refresh manually if needed
  refresh: () => Promise<void>;
}

export function useTasks(): UseTasksResult {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  // Fetch all tasks
  const fetchTasks = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      
      const records = await pb.collection('tasks').getFullList<Task>({
        sort: 'position',
      });
      
      setTasks(records);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to fetch tasks'));
      console.error('Failed to fetch tasks:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  // Initial fetch and real-time subscription
  useEffect(() => {
    fetchTasks();

    // Subscribe to all changes in the tasks collection
    // PocketBase sends events for create, update, and delete
    const unsubscribe = pb.collection('tasks').subscribe<Task>('*', (event) => {
      console.log('Real-time event:', event.action, event.record.id);
      
      switch (event.action) {
        case 'create':
          // Add new task to state
          setTasks(prev => [...prev, event.record].sort((a, b) => a.position - b.position));
          break;
          
        case 'update':
          // Update existing task in state
          setTasks(prev => 
            prev
              .map(t => t.id === event.record.id ? event.record : t)
              .sort((a, b) => a.position - b.position)
          );
          break;
          
        case 'delete':
          // Remove task from state
          setTasks(prev => prev.filter(t => t.id !== event.record.id));
          break;
      }
    });

    // Cleanup subscription on unmount
    return () => {
      pb.collection('tasks').unsubscribe('*');
    };
  }, [fetchTasks]);

  // Create a new task
  const createTask = useCallback(async (task: Partial<Task>): Promise<Task> => {
    const record = await pb.collection('tasks').create<Task>(task);
    // Note: Real-time subscription will update state automatically
    return record;
  }, []);

  // Update an existing task
  const updateTask = useCallback(async (id: string, updates: Partial<Task>): Promise<Task> => {
    const record = await pb.collection('tasks').update<Task>(id, updates);
    // Note: Real-time subscription will update state automatically
    return record;
  }, []);

  // Delete a task
  const deleteTask = useCallback(async (id: string): Promise<void> => {
    await pb.collection('tasks').delete(id);
    // Note: Real-time subscription will update state automatically
  }, []);

  // Move a task to a new column and/or position
  const moveTask = useCallback(async (id: string, column: string, position: number): Promise<Task> => {
    const record = await pb.collection('tasks').update<Task>(id, {
      column,
      position,
    });
    // Note: Real-time subscription will update state automatically
    return record;
  }, []);

  return {
    tasks,
    loading,
    error,
    createTask,
    updateTask,
    deleteTask,
    moveTask,
    refresh: fetchTasks,
  };
}
```

**Steps**:

1. Open the existing file (or create if it doesn't exist):
   ```bash
   # File should already exist from Phase 2
   # If not: touch ui/src/hooks/useTasks.ts
   ```

2. Replace the contents with the code above.

3. Verify compilation:
   ```bash
   cd ui && npm run build
   ```

**Testing Real-time Updates**:

1. Start the UI dev server: `cd ui && npm run dev`
2. Start the Go server: `make dev`
3. Open the UI in a browser
4. In another terminal, create a task via CLI:
   ```bash
   ./egenskriven add "Test real-time" --column todo
   ```
5. The task should appear in the UI immediately without refresh

---

### 4.8 Wire Everything Together in App.tsx

**What**: Update the main App component to integrate all the new features.

**Why**: All the individual components need to be connected with proper state management and keyboard shortcuts.

**File**: Update `ui/src/App.tsx`

```tsx
import { useState, useCallback, useMemo } from 'react';
import { SelectionProvider, useSelection } from './hooks/useSelection';
import { useKeyboardShortcuts } from './hooks/useKeyboard';
import { useTasks } from './hooks/useTasks';
import { Layout } from './components/Layout';
import { Board } from './components/Board';
import { TaskDetail } from './components/TaskDetail';
import { QuickCreate } from './components/QuickCreate';
import { CommandPalette, Command } from './components/CommandPalette';
import { PropertyPicker, STATUS_OPTIONS, PRIORITY_OPTIONS, TYPE_OPTIONS } from './components/PropertyPicker';
import { ShortcutsHelp } from './components/ShortcutsHelp';
import { PeekPreview } from './components/PeekPreview';
import { Task } from './types/task';

function AppContent() {
  const { tasks, loading, updateTask, createTask, deleteTask } = useTasks();
  const { selectedTaskId, selectTask, clearSelection } = useSelection();
  
  // Modal states
  const [isCommandPaletteOpen, setIsCommandPaletteOpen] = useState(false);
  const [isQuickCreateOpen, setIsQuickCreateOpen] = useState(false);
  const [isShortcutsHelpOpen, setIsShortcutsHelpOpen] = useState(false);
  const [isDetailOpen, setIsDetailOpen] = useState(false);
  const [isPeekOpen, setIsPeekOpen] = useState(false);
  
  // Property picker states
  const [statusPickerOpen, setStatusPickerOpen] = useState(false);
  const [priorityPickerOpen, setPriorityPickerOpen] = useState(false);
  const [typePickerOpen, setTypePickerOpen] = useState(false);
  
  // Get the currently selected task
  const selectedTask = useMemo(() => 
    tasks.find(t => t.id === selectedTaskId) || null,
    [tasks, selectedTaskId]
  );
  
  // Get sorted task IDs for navigation
  const sortedTaskIds = useMemo(() => 
    tasks.map(t => t.id),
    [tasks]
  );
  
  // Navigation helpers
  const navigateToNextTask = useCallback(() => {
    if (!selectedTaskId) {
      if (sortedTaskIds.length > 0) {
        selectTask(sortedTaskIds[0]);
      }
      return;
    }
    
    const currentIndex = sortedTaskIds.indexOf(selectedTaskId);
    if (currentIndex < sortedTaskIds.length - 1) {
      selectTask(sortedTaskIds[currentIndex + 1]);
    }
  }, [selectedTaskId, sortedTaskIds, selectTask]);
  
  const navigateToPrevTask = useCallback(() => {
    if (!selectedTaskId) {
      if (sortedTaskIds.length > 0) {
        selectTask(sortedTaskIds[sortedTaskIds.length - 1]);
      }
      return;
    }
    
    const currentIndex = sortedTaskIds.indexOf(selectedTaskId);
    if (currentIndex > 0) {
      selectTask(sortedTaskIds[currentIndex - 1]);
    }
  }, [selectedTaskId, sortedTaskIds, selectTask]);
  
  // Action handlers
  const openTaskDetail = useCallback(() => {
    if (selectedTaskId) {
      setIsDetailOpen(true);
      setIsPeekOpen(false);
    }
  }, [selectedTaskId]);
  
  const handleCreateTask = useCallback(async (taskData: Partial<Task>) => {
    const newTask = await createTask(taskData);
    setIsQuickCreateOpen(false);
    selectTask(newTask.id);
  }, [createTask, selectTask]);
  
  const handleDeleteTask = useCallback(async () => {
    if (selectedTaskId && window.confirm('Delete this task?')) {
      await deleteTask(selectedTaskId);
      clearSelection();
    }
  }, [selectedTaskId, deleteTask, clearSelection]);
  
  const handleStatusChange = useCallback(async (status: string) => {
    if (selectedTaskId) {
      await updateTask(selectedTaskId, { column: status });
    }
  }, [selectedTaskId, updateTask]);
  
  const handlePriorityChange = useCallback(async (priority: string) => {
    if (selectedTaskId) {
      await updateTask(selectedTaskId, { priority });
    }
  }, [selectedTaskId, updateTask]);
  
  const handleTypeChange = useCallback(async (type: string) => {
    if (selectedTaskId) {
      await updateTask(selectedTaskId, { type });
    }
  }, [selectedTaskId, updateTask]);
  
  // Build command palette commands
  const commands: Command[] = useMemo(() => [
    // Actions
    {
      id: 'create-task',
      label: 'Create task',
      shortcut: { key: 'c' },
      section: 'actions',
      icon: '+',
      action: () => setIsQuickCreateOpen(true),
    },
    {
      id: 'change-status',
      label: 'Change status',
      shortcut: { key: 's' },
      section: 'actions',
      icon: '●',
      action: () => setStatusPickerOpen(true),
      when: () => !!selectedTaskId,
    },
    {
      id: 'set-priority',
      label: 'Set priority',
      shortcut: { key: 'p' },
      section: 'actions',
      icon: '!',
      action: () => setPriorityPickerOpen(true),
      when: () => !!selectedTaskId,
    },
    {
      id: 'set-type',
      label: 'Set type',
      shortcut: { key: 't' },
      section: 'actions',
      icon: '◆',
      action: () => setTypePickerOpen(true),
      when: () => !!selectedTaskId,
    },
    {
      id: 'delete-task',
      label: 'Delete task',
      shortcut: { key: 'Backspace' },
      section: 'actions',
      icon: '×',
      action: handleDeleteTask,
      when: () => !!selectedTaskId,
    },
    {
      id: 'show-shortcuts',
      label: 'Show keyboard shortcuts',
      shortcut: { key: '?' },
      section: 'actions',
      icon: '?',
      action: () => setIsShortcutsHelpOpen(true),
    },
    
    // Navigation - add recent tasks
    ...tasks.slice(0, 5).map(task => ({
      id: `task-${task.id}`,
      label: `${task.id.substring(0, 8)} ${task.title}`,
      section: 'recent' as const,
      action: () => {
        selectTask(task.id);
        setIsDetailOpen(true);
      },
    })),
  ], [tasks, selectedTaskId, handleDeleteTask, selectTask]);
  
  // Register keyboard shortcuts
  useKeyboardShortcuts([
    // Global shortcuts
    {
      combo: { key: 'k', meta: true },
      handler: () => setIsCommandPaletteOpen(true),
      description: 'Open command palette',
    },
    {
      combo: { key: '?' },
      handler: () => setIsShortcutsHelpOpen(true),
      description: 'Show shortcuts help',
    },
    {
      combo: { key: 'Escape' },
      handler: () => {
        if (isPeekOpen) {
          setIsPeekOpen(false);
        } else if (isDetailOpen) {
          setIsDetailOpen(false);
        } else if (selectedTaskId) {
          clearSelection();
        }
      },
      description: 'Close/deselect',
      allowInInput: true,
    },
    
    // Task actions
    {
      combo: { key: 'c' },
      handler: () => setIsQuickCreateOpen(true),
      description: 'Create task',
    },
    {
      combo: { key: 'Enter' },
      handler: () => openTaskDetail(),
      description: 'Open selected task',
    },
    {
      combo: { key: ' ' },
      handler: () => {
        if (selectedTaskId) {
          setIsPeekOpen(prev => !prev);
        }
      },
      description: 'Peek preview',
    },
    {
      combo: { key: 'Backspace' },
      handler: () => handleDeleteTask(),
      description: 'Delete task',
    },
    
    // Property shortcuts (only when task selected)
    {
      combo: { key: 's' },
      handler: () => setStatusPickerOpen(true),
      when: () => !!selectedTaskId,
      description: 'Set status',
    },
    {
      combo: { key: 'p' },
      handler: () => setPriorityPickerOpen(true),
      when: () => !!selectedTaskId,
      description: 'Set priority',
    },
    {
      combo: { key: 't' },
      handler: () => setTypePickerOpen(true),
      when: () => !!selectedTaskId,
      description: 'Set type',
    },
    
    // Navigation
    {
      combo: { key: 'j' },
      handler: () => navigateToNextTask(),
      description: 'Next task',
    },
    {
      combo: { key: 'ArrowDown' },
      handler: () => navigateToNextTask(),
      description: 'Next task',
    },
    {
      combo: { key: 'k' },
      handler: () => navigateToPrevTask(),
      description: 'Previous task',
    },
    {
      combo: { key: 'ArrowUp' },
      handler: () => navigateToPrevTask(),
      description: 'Previous task',
    },
  ]);
  
  if (loading) {
    return <div className="loading">Loading...</div>;
  }
  
  return (
    <Layout>
      <Board 
        tasks={tasks}
        onTaskClick={(task) => {
          selectTask(task.id);
          setIsDetailOpen(true);
        }}
        onTaskSelect={(task) => selectTask(task.id)}
      />
      
      {/* Task Detail Panel */}
      <TaskDetail
        task={selectedTask}
        isOpen={isDetailOpen}
        onClose={() => setIsDetailOpen(false)}
        onUpdate={updateTask}
      />
      
      {/* Quick Create Modal */}
      <QuickCreate
        isOpen={isQuickCreateOpen}
        onClose={() => setIsQuickCreateOpen(false)}
        onCreate={handleCreateTask}
      />
      
      {/* Command Palette */}
      <CommandPalette
        isOpen={isCommandPaletteOpen}
        onClose={() => setIsCommandPaletteOpen(false)}
        commands={commands}
      />
      
      {/* Property Pickers */}
      <PropertyPicker
        isOpen={statusPickerOpen}
        onClose={() => setStatusPickerOpen(false)}
        onSelect={handleStatusChange}
        options={STATUS_OPTIONS}
        currentValue={selectedTask?.column}
        title="Set Status"
      />
      
      <PropertyPicker
        isOpen={priorityPickerOpen}
        onClose={() => setPriorityPickerOpen(false)}
        onSelect={handlePriorityChange}
        options={PRIORITY_OPTIONS}
        currentValue={selectedTask?.priority}
        title="Set Priority"
      />
      
      <PropertyPicker
        isOpen={typePickerOpen}
        onClose={() => setTypePickerOpen(false)}
        onSelect={handleTypeChange}
        options={TYPE_OPTIONS}
        currentValue={selectedTask?.type}
        title="Set Type"
      />
      
      {/* Shortcuts Help */}
      <ShortcutsHelp
        isOpen={isShortcutsHelpOpen}
        onClose={() => setIsShortcutsHelpOpen(false)}
      />
      
      {/* Peek Preview */}
      <PeekPreview
        task={selectedTask}
        isOpen={isPeekOpen}
        onClose={() => setIsPeekOpen(false)}
      />
    </Layout>
  );
}

export default function App() {
  return (
    <SelectionProvider>
      <AppContent />
    </SelectionProvider>
  );
}
```

**Steps**:

1. Open `ui/src/App.tsx` and replace with the code above.

2. Verify compilation:
   ```bash
   cd ui && npm run build
   ```

3. If you get import errors, ensure all component files exist from previous tasks.

---

### 4.9 Update TaskCard for Selection State

**What**: Update the TaskCard component to show visual feedback when selected.

**Why**: Users need to see which task is currently selected for keyboard actions to make sense.

**File**: Update `ui/src/components/TaskCard.tsx`

Add the `isSelected` prop and corresponding styles:

```tsx
// Add to existing TaskCard component

interface TaskCardProps {
  task: Task;
  isSelected?: boolean;
  onClick?: () => void;
  onSelect?: () => void;
}

export function TaskCard({ task, isSelected, onClick, onSelect }: TaskCardProps) {
  return (
    <div
      className={`${styles.card} ${isSelected ? styles.selected : ''}`}
      onClick={onClick}
      onMouseDown={onSelect}
      tabIndex={0}
    >
      {/* ... existing card content ... */}
    </div>
  );
}
```

**File**: Update `ui/src/components/TaskCard.module.css`

Add selected state styles:

```css
/* Add to existing styles */

.card.selected {
  background: var(--bg-card-selected, #2E2E2E);
  border: 1px solid var(--accent, #5E6AD2);
  box-shadow: 0 0 0 1px var(--accent, #5E6AD2);
}

.card:focus {
  outline: none;
  border-color: var(--accent, #5E6AD2);
}

.card:focus-visible {
  box-shadow: 0 0 0 2px var(--accent, #5E6AD2);
}
```

---

### 4.10 Write Component Tests

**What**: Write tests for the new interactive components.

**Why**: Tests ensure the keyboard shortcuts, command palette, and property pickers work correctly.

**File**: `ui/src/components/CommandPalette.test.tsx`

```tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { CommandPalette, Command } from './CommandPalette';

const mockCommands: Command[] = [
  {
    id: 'create',
    label: 'Create task',
    section: 'actions',
    action: vi.fn(),
  },
  {
    id: 'delete',
    label: 'Delete task',
    section: 'actions',
    action: vi.fn(),
  },
  {
    id: 'go-board',
    label: 'Go to board',
    section: 'navigation',
    action: vi.fn(),
  },
];

describe('CommandPalette', () => {
  it('renders nothing when closed', () => {
    render(
      <CommandPalette 
        isOpen={false} 
        onClose={() => {}} 
        commands={mockCommands} 
      />
    );
    
    expect(screen.queryByPlaceholderText(/type a command/i)).not.toBeInTheDocument();
  });
  
  it('renders command list when open', () => {
    render(
      <CommandPalette 
        isOpen={true} 
        onClose={() => {}} 
        commands={mockCommands} 
      />
    );
    
    expect(screen.getByPlaceholderText(/type a command/i)).toBeInTheDocument();
    expect(screen.getByText('Create task')).toBeInTheDocument();
    expect(screen.getByText('Delete task')).toBeInTheDocument();
  });
  
  it('filters commands by search query', () => {
    render(
      <CommandPalette 
        isOpen={true} 
        onClose={() => {}} 
        commands={mockCommands} 
      />
    );
    
    const input = screen.getByPlaceholderText(/type a command/i);
    fireEvent.change(input, { target: { value: 'create' } });
    
    expect(screen.getByText('Create task')).toBeInTheDocument();
    expect(screen.queryByText('Delete task')).not.toBeInTheDocument();
  });
  
  it('executes command on click', () => {
    const onClose = vi.fn();
    
    render(
      <CommandPalette 
        isOpen={true} 
        onClose={onClose} 
        commands={mockCommands} 
      />
    );
    
    fireEvent.click(screen.getByText('Create task'));
    
    expect(onClose).toHaveBeenCalled();
    // Action is called after a timeout
  });
  
  it('closes on Escape key', () => {
    const onClose = vi.fn();
    
    render(
      <CommandPalette 
        isOpen={true} 
        onClose={onClose} 
        commands={mockCommands} 
      />
    );
    
    const input = screen.getByPlaceholderText(/type a command/i);
    fireEvent.keyDown(input, { key: 'Escape' });
    
    expect(onClose).toHaveBeenCalled();
  });
  
  it('navigates with arrow keys', () => {
    render(
      <CommandPalette 
        isOpen={true} 
        onClose={() => {}} 
        commands={mockCommands} 
      />
    );
    
    const input = screen.getByPlaceholderText(/type a command/i);
    
    // First item should be selected initially
    // Press down to select second
    fireEvent.keyDown(input, { key: 'ArrowDown' });
    
    // Check that selection moved (by checking class or aria-selected)
    // This depends on your implementation
  });
});
```

**File**: `ui/src/hooks/useSelection.test.ts`

```typescript
import { describe, it, expect } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { SelectionProvider, useSelection } from './useSelection';
import { ReactNode } from 'react';

const wrapper = ({ children }: { children: ReactNode }) => (
  <SelectionProvider>{children}</SelectionProvider>
);

describe('useSelection', () => {
  it('starts with no selection', () => {
    const { result } = renderHook(() => useSelection(), { wrapper });
    
    expect(result.current.selectedTaskId).toBeNull();
    expect(result.current.multiSelectedIds.size).toBe(0);
  });
  
  it('selects a single task', () => {
    const { result } = renderHook(() => useSelection(), { wrapper });
    
    act(() => {
      result.current.selectTask('task-1');
    });
    
    expect(result.current.selectedTaskId).toBe('task-1');
    expect(result.current.isSelected('task-1')).toBe(true);
    expect(result.current.isSelected('task-2')).toBe(false);
  });
  
  it('clears multi-selection when single selecting', () => {
    const { result } = renderHook(() => useSelection(), { wrapper });
    
    act(() => {
      result.current.toggleMultiSelect('task-1');
      result.current.toggleMultiSelect('task-2');
    });
    
    expect(result.current.multiSelectedIds.size).toBe(2);
    
    act(() => {
      result.current.selectTask('task-3');
    });
    
    expect(result.current.multiSelectedIds.size).toBe(0);
    expect(result.current.selectedTaskId).toBe('task-3');
  });
  
  it('toggles multi-selection', () => {
    const { result } = renderHook(() => useSelection(), { wrapper });
    
    act(() => {
      result.current.toggleMultiSelect('task-1');
    });
    
    expect(result.current.isSelected('task-1')).toBe(true);
    
    act(() => {
      result.current.toggleMultiSelect('task-1');
    });
    
    expect(result.current.isSelected('task-1')).toBe(false);
  });
  
  it('selects a range of tasks', () => {
    const { result } = renderHook(() => useSelection(), { wrapper });
    const allIds = ['task-1', 'task-2', 'task-3', 'task-4', 'task-5'];
    
    act(() => {
      result.current.selectRange('task-2', 'task-4', allIds);
    });
    
    expect(result.current.isSelected('task-1')).toBe(false);
    expect(result.current.isSelected('task-2')).toBe(true);
    expect(result.current.isSelected('task-3')).toBe(true);
    expect(result.current.isSelected('task-4')).toBe(true);
    expect(result.current.isSelected('task-5')).toBe(false);
  });
  
  it('clears all selection', () => {
    const { result } = renderHook(() => useSelection(), { wrapper });
    
    act(() => {
      result.current.selectTask('task-1');
      result.current.toggleMultiSelect('task-2');
    });
    
    act(() => {
      result.current.clearSelection();
    });
    
    expect(result.current.selectedTaskId).toBeNull();
    expect(result.current.multiSelectedIds.size).toBe(0);
  });
});
```

**Steps**:

1. Create the test files:
   ```bash
   touch ui/src/components/CommandPalette.test.tsx
   touch ui/src/hooks/useSelection.test.ts
   ```

2. Paste the test code.

3. Run tests:
   ```bash
   cd ui && npm test
   ```

---

## Verification Checklist

Complete each section in order. Check off each item as you verify it.

### Build Verification

- [ ] **UI builds successfully**
  ```bash
  cd ui && npm run build
  ```
  Should complete without errors.

- [ ] **UI tests pass**
  ```bash
  cd ui && npm test
  ```
  All tests should pass.

- [ ] **Full application builds**
  ```bash
  make build
  ```
  Should produce binary with embedded UI.

### Keyboard Shortcut Verification

Run the UI (`cd ui && npm run dev`) and verify:

- [ ] **Command Palette** - `Cmd+K` (Mac) or `Ctrl+K` (Windows) opens palette
- [ ] **Shortcuts Help** - `?` opens help modal
- [ ] **Create Task** - `C` opens quick create
- [ ] **Navigation** - `J` moves to next task, `K` moves to previous
- [ ] **Arrow Keys** - `↓` and `↑` also navigate between tasks
- [ ] **Open Task** - `Enter` opens selected task's detail panel
- [ ] **Peek Preview** - `Space` shows quick preview of selected task
- [ ] **Close/Deselect** - `Escape` closes panels and clears selection
- [ ] **Delete Task** - `Backspace` prompts to delete selected task

### Property Picker Verification

With a task selected:

- [ ] **Status Picker** - `S` opens status picker, arrow keys navigate, Enter selects
- [ ] **Priority Picker** - `P` opens priority picker
- [ ] **Type Picker** - `T` opens type picker
- [ ] **Picker Filtering** - Typing filters options
- [ ] **Current Value** - Current value is highlighted with checkmark

### Command Palette Verification

- [ ] **Search** - Typing filters commands
- [ ] **Sections** - Commands grouped into Actions, Navigation, Recent
- [ ] **Keyboard Navigation** - Arrow keys move selection
- [ ] **Execute** - Enter executes selected command
- [ ] **Close** - Escape or click outside closes palette

### Selection State Verification

- [ ] **Visual Feedback** - Selected task has visible highlight/border
- [ ] **Click Select** - Clicking a task selects it
- [ ] **Keyboard Select** - Navigating with J/K changes selection
- [ ] **Clear Selection** - Escape clears selection

### Real-time Updates Verification

1. Start the UI in one terminal
2. In another terminal, run CLI commands:

- [ ] **Create via CLI** - `./egenskriven add "Test" --column todo` appears immediately in UI
- [ ] **Update via CLI** - `./egenskriven move <id> in_progress` moves card in UI
- [ ] **Delete via CLI** - `./egenskriven delete <id> --force` removes card from UI

### Shortcuts Not Firing in Inputs

- [ ] **In Quick Create** - Typing in title field doesn't trigger shortcuts
- [ ] **In Command Palette** - Typing doesn't trigger task shortcuts
- [ ] **In Property Picker** - Typing filters, doesn't trigger shortcuts

---

## File Summary

| File | Lines | Purpose |
|------|-------|---------|
| `ui/src/hooks/useSelection.ts` | ~120 | Selection state management |
| `ui/src/hooks/useKeyboard.ts` | ~120 | Keyboard shortcut system |
| `ui/src/components/CommandPalette.tsx` | ~200 | Command palette UI |
| `ui/src/components/CommandPalette.module.css` | ~100 | Command palette styles |
| `ui/src/components/PropertyPicker.tsx` | ~200 | Property picker component |
| `ui/src/components/PropertyPicker.module.css` | ~80 | Property picker styles |
| `ui/src/components/ShortcutsHelp.tsx` | ~120 | Shortcuts help modal |
| `ui/src/components/ShortcutsHelp.module.css` | ~100 | Shortcuts help styles |
| `ui/src/components/PeekPreview.tsx` | ~100 | Peek preview overlay |
| `ui/src/components/PeekPreview.module.css` | ~120 | Peek preview styles |
| `ui/src/hooks/useTasks.ts` | ~100 | Updated with real-time subscriptions |
| `ui/src/App.tsx` | ~250 | Updated with all integrations |
| `ui/src/components/CommandPalette.test.tsx` | ~100 | Command palette tests |
| `ui/src/hooks/useSelection.test.ts` | ~100 | Selection hook tests |

**Total new/modified code**: ~1,810 lines

---

## What You Should Have Now

After completing Phase 4, your UI should:

1. **Be fully keyboard-navigable** - All actions accessible via shortcuts
2. **Have a command palette** - Quick access to all commands with `Cmd+K`
3. **Show shortcuts help** - `?` displays all available shortcuts
4. **Support property pickers** - Quick status/priority/type changes
5. **Update in real-time** - CLI changes appear immediately
6. **Provide visual selection feedback** - Selected task is highlighted
7. **Support peek preview** - Quick task preview with `Space`

---

## Next Phase

**Phase 5: Multi-Board Support** will add:
- Multiple boards with board switcher
- Board-specific task ID prefixes (e.g., WRK-123)
- Sidebar with board list
- Board settings and customization
- CLI commands for board management

---

## Troubleshooting

### Keyboard shortcuts not working

**Problem**: Pressing shortcut keys does nothing.

**Solution**:
1. Make sure focus is not in an input field
2. Check browser console for errors
3. Verify the keyboard hook is registered in App.tsx
4. Check if another component is preventing event propagation

### Real-time updates not appearing

**Problem**: CLI changes don't show in UI.

**Solution**:
1. Verify PocketBase is running (`make dev`)
2. Check browser console for WebSocket/SSE connection errors
3. Ensure the collection name matches exactly ("tasks")
4. Try refreshing the page

### Command palette not filtering

**Problem**: Typing in palette doesn't filter commands.

**Solution**:
1. Check if the input is focused (click in the input field)
2. Verify the `query` state is being updated
3. Check console for JavaScript errors

### Property picker positioning wrong

**Problem**: Picker appears in wrong location or off-screen.

**Solution**:
1. Verify `anchorElement` is being passed correctly
2. Check the positioning logic in useEffect
3. The picker uses `position: fixed` - ensure parent doesn't have `transform`

### Selection visual not showing

**Problem**: Selected task doesn't look different.

**Solution**:
1. Verify `isSelected` prop is being passed to TaskCard
2. Check CSS module is being applied correctly
3. Inspect element to see if `.selected` class is present

---

## Glossary

| Term | Definition |
|------|------------|
| **Command Palette** | Modal with searchable list of all available commands |
| **Property Picker** | Popover for changing a single task property |
| **Peek Preview** | Quick overlay showing task details without opening full panel |
| **Selection State** | Tracking which task(s) are currently selected |
| **Real-time Subscription** | PocketBase SSE connection for live updates |
| **Key Combo** | Combination of keys that trigger a shortcut |
| **Fuzzy Search** | Search that matches partial/out-of-order characters |
| **Portal** | React pattern for rendering outside component hierarchy |
