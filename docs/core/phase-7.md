# Phase 7: Polish

**Goal**: Complete visual polish including theming, animations, responsive design, and quality of life improvements.

**Duration Estimate**: 5-7 days

**Prerequisites**: Phase 6 complete (Filtering & Views)

**Deliverable**: A fully polished UI with light/dark themes, smooth animations, responsive layout for all screen sizes, and performant experience for large boards.

---

## Overview

Phase 7 transforms the functional UI from previous phases into a polished, professional application. This phase focuses on:

- **Theming**: Light mode, dark mode, system preference detection, and accent colors
- **Animations**: Smooth transitions that enhance usability without feeling slow
- **Responsive Design**: Full functionality on mobile, tablet, and desktop
- **Quality of Life**: Loading states, toast notifications, and visual feedback
- **Performance**: Virtualization and optimization for large datasets

### Why This Phase Matters

Polish is what separates a prototype from a product. Users perceive a polished UI as more trustworthy and easier to use. Key benefits:

- **Theme support** respects user preferences and reduces eye strain
- **Animations** provide feedback that makes the app feel responsive
- **Responsive design** ensures the app works everywhere
- **Loading states** communicate what's happening and reduce perceived wait time
- **Performance optimization** keeps the app fast as data grows

### Design Reference

Refer to `docs/ui-design.md` for complete design specifications including:
- Color tokens for dark and light modes
- Typography scale
- Spacing values
- Animation timing
- Responsive breakpoints

---

## Environment Requirements

Before starting, ensure you have completed Phase 6 and have:

| Requirement | Description |
|-------------|-------------|
| Phase 6 Complete | Filtering and saved views working |
| Node.js 18+ | `node --version` |
| React Dev Server | `cd ui && npm run dev` should work |

---

## Tasks

### 7.1 Implement Light Mode Color Tokens

**What**: Add CSS custom properties for light mode colors.

**Why**: Light mode is essential for users who prefer it or work in bright environments. Using CSS custom properties allows runtime theme switching without page reload.

**File**: `ui/src/styles/light.css`

```css
/*
 * Light mode color tokens
 * 
 * These override the dark mode defaults when applied to <html> element.
 * Applied via: document.documentElement.setAttribute('data-theme', 'light')
 */

[data-theme="light"] {
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
  --border-focus: var(--accent);

  /* Shadows - more visible in light mode */
  --shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
  --shadow-md: 0 4px 6px rgba(0, 0, 0, 0.07);
  --shadow-lg: 0 10px 15px rgba(0, 0, 0, 0.1);

  /* Drag shadow - needs more definition in light mode */
  --shadow-drag: 0 8px 16px rgba(0, 0, 0, 0.15);
}
```

**Steps**:

1. Create the file:
   ```bash
   touch ui/src/styles/light.css
   ```

2. Open in your editor and paste the CSS above.

3. Import in your main CSS file (`ui/src/styles/tokens.css` or `ui/src/index.css`):
   ```css
   @import './light.css';
   ```

4. Verify the file is loaded by checking the browser's DevTools:
   - Open DevTools > Elements
   - Select the `<html>` element
   - Add attribute `data-theme="light"`
   - Confirm the UI switches to light colors

**Common Mistakes**:
- Forgetting to import the light.css file
- Using incorrect CSS custom property names (must match dark mode exactly)
- Missing the `[data-theme="light"]` selector

---

### 7.2 Create Theme Context

**What**: Create a React context for managing theme state across the application.

**Why**: Theme preference needs to be accessible from any component (settings panel, command palette, etc.) and persisted across sessions.

**File**: `ui/src/contexts/ThemeContext.tsx`

```tsx
import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from 'react';

/**
 * Theme options:
 * - 'system': Follow OS preference
 * - 'light': Always light mode
 * - 'dark': Always dark mode
 */
type Theme = 'system' | 'light' | 'dark';

/**
 * Resolved theme is what's actually applied to the DOM.
 * When theme is 'system', this reflects the OS preference.
 */
type ResolvedTheme = 'light' | 'dark';

interface ThemeContextValue {
  /** Current theme setting (may be 'system') */
  theme: Theme;
  /** Actually applied theme (never 'system') */
  resolvedTheme: ResolvedTheme;
  /** Update theme setting */
  setTheme: (theme: Theme) => void;
}

const ThemeContext = createContext<ThemeContextValue | undefined>(undefined);

const STORAGE_KEY = 'egenskriven-theme';

/**
 * Get the system's color scheme preference.
 */
function getSystemTheme(): ResolvedTheme {
  if (typeof window === 'undefined') return 'dark';
  return window.matchMedia('(prefers-color-scheme: light)').matches
    ? 'light'
    : 'dark';
}

/**
 * Apply theme to DOM by setting data-theme attribute on <html>.
 */
function applyTheme(resolvedTheme: ResolvedTheme): void {
  document.documentElement.setAttribute('data-theme', resolvedTheme);
}

interface ThemeProviderProps {
  children: ReactNode;
}

export function ThemeProvider({ children }: ThemeProviderProps) {
  // Initialize from localStorage, defaulting to 'system'
  const [theme, setThemeState] = useState<Theme>(() => {
    if (typeof window === 'undefined') return 'system';
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored === 'light' || stored === 'dark' || stored === 'system') {
      return stored;
    }
    return 'system';
  });

  // Track the resolved (actual) theme
  const [resolvedTheme, setResolvedTheme] = useState<ResolvedTheme>(() => {
    if (theme === 'system') return getSystemTheme();
    return theme;
  });

  // Update resolved theme when theme setting changes
  useEffect(() => {
    const newResolved = theme === 'system' ? getSystemTheme() : theme;
    setResolvedTheme(newResolved);
    applyTheme(newResolved);
  }, [theme]);

  // Listen for system theme changes when using 'system' mode
  useEffect(() => {
    if (theme !== 'system') return;

    const mediaQuery = window.matchMedia('(prefers-color-scheme: light)');
    
    const handleChange = (e: MediaQueryListEvent) => {
      const newResolved = e.matches ? 'light' : 'dark';
      setResolvedTheme(newResolved);
      applyTheme(newResolved);
    };

    mediaQuery.addEventListener('change', handleChange);
    return () => mediaQuery.removeEventListener('change', handleChange);
  }, [theme]);

  // Persist theme to localStorage
  const setTheme = (newTheme: Theme) => {
    setThemeState(newTheme);
    localStorage.setItem(STORAGE_KEY, newTheme);
  };

  return (
    <ThemeContext.Provider value={{ theme, resolvedTheme, setTheme }}>
      {children}
    </ThemeContext.Provider>
  );
}

/**
 * Hook to access theme context.
 * 
 * @example
 * const { theme, resolvedTheme, setTheme } = useTheme();
 * 
 * // Toggle between light and dark
 * setTheme(resolvedTheme === 'light' ? 'dark' : 'light');
 * 
 * // Set to follow system
 * setTheme('system');
 */
export function useTheme(): ThemeContextValue {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }
  return context;
}
```

**Steps**:

1. Create the contexts directory if it doesn't exist:
   ```bash
   mkdir -p ui/src/contexts
   ```

2. Create the file:
   ```bash
   touch ui/src/contexts/ThemeContext.tsx
   ```

3. Open in your editor and paste the code above.

4. Wrap your app with the ThemeProvider. In `ui/src/App.tsx` or `ui/src/main.tsx`:
   ```tsx
   import { ThemeProvider } from './contexts/ThemeContext';

   function App() {
     return (
       <ThemeProvider>
         {/* rest of your app */}
       </ThemeProvider>
     );
   }
   ```

5. Verify it works:
   - Add a temporary button to test:
     ```tsx
     const { theme, setTheme } = useTheme();
     <button onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}>
       Toggle Theme
     </button>
     ```
   - Click the button and confirm colors change
   - Refresh the page and confirm the preference persisted

**Key Concepts Explained**:

| Concept | Explanation |
|---------|-------------|
| `theme` | User's preference: 'system', 'light', or 'dark' |
| `resolvedTheme` | Actual theme applied: always 'light' or 'dark' |
| `matchMedia` | Browser API to detect system color scheme preference |
| `localStorage` | Persists preference across browser sessions |

---

### 7.3 Create Settings Panel

**What**: Create a settings panel accessible via `Cmd+,` or settings icon.

**Why**: Users need a central place to configure appearance preferences like theme and accent color.

**File**: `ui/src/components/Settings.tsx`

```tsx
import { useEffect, useRef } from 'react';
import { useTheme } from '../contexts/ThemeContext';
import { useAccentColor } from '../hooks/useAccentColor';
import './Settings.css';

interface SettingsProps {
  /** Whether the settings panel is open */
  isOpen: boolean;
  /** Callback when settings should close */
  onClose: () => void;
}

/**
 * Available accent color options.
 * Each has a name and hex value.
 */
const ACCENT_COLORS = [
  { name: 'Blue', value: '#5E6AD2' },
  { name: 'Purple', value: '#9333EA' },
  { name: 'Green', value: '#22C55E' },
  { name: 'Orange', value: '#F97316' },
  { name: 'Pink', value: '#EC4899' },
  { name: 'Cyan', value: '#06B6D4' },
  { name: 'Red', value: '#EF4444' },
  { name: 'Yellow', value: '#EAB308' },
] as const;

export function Settings({ isOpen, onClose }: SettingsProps) {
  const { theme, setTheme } = useTheme();
  const { accentColor, setAccentColor } = useAccentColor();
  const panelRef = useRef<HTMLDivElement>(null);

  // Close on Escape key
  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, onClose]);

  // Close when clicking outside
  useEffect(() => {
    if (!isOpen) return;

    const handleClickOutside = (e: MouseEvent) => {
      if (panelRef.current && !panelRef.current.contains(e.target as Node)) {
        onClose();
      }
    };

    // Delay to prevent immediate close from the click that opened it
    const timeoutId = setTimeout(() => {
      document.addEventListener('click', handleClickOutside);
    }, 0);

    return () => {
      clearTimeout(timeoutId);
      document.removeEventListener('click', handleClickOutside);
    };
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  return (
    <div className="settings-overlay">
      <div className="settings-panel" ref={panelRef}>
        <header className="settings-header">
          <h2>Settings</h2>
          <button 
            className="settings-close" 
            onClick={onClose}
            aria-label="Close settings"
          >
            &times;
          </button>
        </header>

        <div className="settings-content">
          {/* Appearance Section */}
          <section className="settings-section">
            <h3>Appearance</h3>

            {/* Theme Selection */}
            <div className="settings-row">
              <label htmlFor="theme-select">Theme</label>
              <select
                id="theme-select"
                value={theme}
                onChange={(e) => setTheme(e.target.value as 'system' | 'light' | 'dark')}
                className="settings-select"
              >
                <option value="system">System</option>
                <option value="light">Light</option>
                <option value="dark">Dark</option>
              </select>
            </div>

            {/* Accent Color Selection */}
            <div className="settings-row">
              <label>Accent Color</label>
              <div className="accent-color-grid">
                {ACCENT_COLORS.map((color) => (
                  <button
                    key={color.value}
                    className={`accent-color-option ${
                      accentColor === color.value ? 'selected' : ''
                    }`}
                    style={{ backgroundColor: color.value }}
                    onClick={() => setAccentColor(color.value)}
                    title={color.name}
                    aria-label={`Set accent color to ${color.name}`}
                  >
                    {accentColor === color.value && (
                      <span className="accent-color-check">&#10003;</span>
                    )}
                  </button>
                ))}
              </div>
            </div>
          </section>

          {/* Keyboard Shortcuts Section */}
          <section className="settings-section">
            <h3>Keyboard Shortcuts</h3>
            <p className="settings-hint">
              Press <kbd>?</kbd> to view all keyboard shortcuts
            </p>
          </section>
        </div>
      </div>
    </div>
  );
}
```

**File**: `ui/src/components/Settings.css`

```css
/* Settings Panel Styles */

.settings-overlay {
  position: fixed;
  inset: 0;
  background: var(--bg-overlay);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  animation: fadeIn var(--duration-normal) ease-out;
}

.settings-panel {
  background: var(--bg-card);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-default);
  width: 100%;
  max-width: 480px;
  max-height: 80vh;
  overflow-y: auto;
  animation: scaleIn var(--duration-normal) ease-out;
}

.settings-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-4) var(--space-6);
  border-bottom: 1px solid var(--border-subtle);
}

.settings-header h2 {
  font-size: var(--text-xl);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  margin: 0;
}

.settings-close {
  background: none;
  border: none;
  font-size: 24px;
  color: var(--text-muted);
  cursor: pointer;
  padding: var(--space-1);
  line-height: 1;
  border-radius: var(--radius-sm);
  transition: color var(--duration-fast) ease;
}

.settings-close:hover {
  color: var(--text-primary);
}

.settings-content {
  padding: var(--space-6);
}

.settings-section {
  margin-bottom: var(--space-6);
}

.settings-section:last-child {
  margin-bottom: 0;
}

.settings-section h3 {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin: 0 0 var(--space-4) 0;
}

.settings-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-4);
}

.settings-row:last-child {
  margin-bottom: 0;
}

.settings-row label {
  font-size: var(--text-base);
  color: var(--text-primary);
}

.settings-select {
  background: var(--bg-input);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  padding: var(--space-2) var(--space-3);
  color: var(--text-primary);
  font-size: var(--text-base);
  cursor: pointer;
  min-width: 120px;
}

.settings-select:focus {
  outline: none;
  border-color: var(--accent);
}

/* Accent Color Grid */
.accent-color-grid {
  display: grid;
  grid-template-columns: repeat(4, 32px);
  gap: var(--space-2);
}

.accent-color-option {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: 2px solid transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform var(--duration-fast) ease,
              border-color var(--duration-fast) ease;
}

.accent-color-option:hover {
  transform: scale(1.1);
}

.accent-color-option.selected {
  border-color: var(--text-primary);
}

.accent-color-check {
  color: white;
  font-size: 14px;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.3);
}

.settings-hint {
  color: var(--text-muted);
  font-size: var(--text-sm);
  margin: 0;
}

.settings-hint kbd {
  background: var(--bg-input);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  padding: 2px 6px;
  font-family: var(--font-mono);
  font-size: var(--text-xs);
}

/* Animations */
@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes scaleIn {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}
```

**Steps**:

1. Create the component file:
   ```bash
   touch ui/src/components/Settings.tsx
   touch ui/src/components/Settings.css
   ```

2. Open in your editor and paste the code above.

3. Create the accent color hook (referenced by Settings). See task 7.4.

4. Add settings state to your main layout component:
   ```tsx
   const [settingsOpen, setSettingsOpen] = useState(false);
   
   // Add keyboard shortcut for Cmd+,
   useEffect(() => {
     const handleKeyDown = (e: KeyboardEvent) => {
       if ((e.metaKey || e.ctrlKey) && e.key === ',') {
         e.preventDefault();
         setSettingsOpen(true);
       }
     };
     window.addEventListener('keydown', handleKeyDown);
     return () => window.removeEventListener('keydown', handleKeyDown);
   }, []);
   
   return (
     <>
       {/* Your existing layout */}
       <Settings isOpen={settingsOpen} onClose={() => setSettingsOpen(false)} />
     </>
   );
   ```

5. Verify settings panel:
   - Press `Cmd+,` (or `Ctrl+,` on Windows/Linux)
   - Settings panel should appear with smooth animation
   - Theme dropdown should change the UI theme
   - Accent colors should be selectable
   - Press `Esc` or click outside to close

---

### 7.4 Implement Accent Colors

**What**: Create a hook for managing accent color with persistence.

**Why**: Accent colors personalize the UI and provide visual consistency for interactive elements like buttons, focus states, and selected items.

**File**: `ui/src/hooks/useAccentColor.ts`

```typescript
import { useState, useEffect } from 'react';

const STORAGE_KEY = 'egenskriven-accent';
const DEFAULT_ACCENT = '#5E6AD2'; // Blue

/**
 * Apply accent color to CSS custom properties.
 * This updates the --accent variable used throughout the app.
 */
function applyAccentColor(color: string): void {
  document.documentElement.style.setProperty('--accent', color);
  
  // Also set RGB values for transparency support
  // Convert hex to RGB
  const r = parseInt(color.slice(1, 3), 16);
  const g = parseInt(color.slice(3, 5), 16);
  const b = parseInt(color.slice(5, 7), 16);
  document.documentElement.style.setProperty('--accent-rgb', `${r}, ${g}, ${b}`);
}

interface UseAccentColorReturn {
  /** Current accent color (hex) */
  accentColor: string;
  /** Update accent color */
  setAccentColor: (color: string) => void;
}

/**
 * Hook to manage accent color preference.
 * 
 * @example
 * const { accentColor, setAccentColor } = useAccentColor();
 * setAccentColor('#22C55E'); // Set to green
 */
export function useAccentColor(): UseAccentColorReturn {
  const [accentColor, setAccentState] = useState<string>(() => {
    if (typeof window === 'undefined') return DEFAULT_ACCENT;
    return localStorage.getItem(STORAGE_KEY) || DEFAULT_ACCENT;
  });

  // Apply accent color on mount and when it changes
  useEffect(() => {
    applyAccentColor(accentColor);
  }, [accentColor]);

  const setAccentColor = (color: string) => {
    // Validate hex color format
    if (!/^#[0-9A-Fa-f]{6}$/.test(color)) {
      console.error('Invalid hex color:', color);
      return;
    }
    setAccentState(color);
    localStorage.setItem(STORAGE_KEY, color);
  };

  return { accentColor, setAccentColor };
}
```

**Steps**:

1. Create the file:
   ```bash
   touch ui/src/hooks/useAccentColor.ts
   ```

2. Open in your editor and paste the code above.

3. Update your CSS to use the accent variable. Ensure `ui/src/styles/tokens.css` has:
   ```css
   :root {
     /* Default accent color (blue) */
     --accent: #5E6AD2;
     --accent-rgb: 94, 106, 210;
   }
   ```

4. Use accent color in CSS for interactive elements:
   ```css
   /* Example: Focus states */
   button:focus-visible {
     outline: 2px solid var(--accent);
     outline-offset: 2px;
   }
   
   /* Example: Selected items */
   .task-card.selected {
     border-color: var(--accent);
   }
   
   /* Example: Primary buttons */
   .button-primary {
     background-color: var(--accent);
   }
   
   /* Example: Accent with transparency */
   .selection-highlight {
     background-color: rgba(var(--accent-rgb), 0.1);
   }
   ```

5. Verify accent colors work:
   - Open Settings panel
   - Click different accent colors
   - Confirm buttons, focus rings, and selected items update
   - Refresh the page and confirm preference persisted

---

### 7.5 Add CSS Animations

**What**: Add smooth transitions for panels, modals, hover states, and drag feedback.

**Why**: Animations provide visual feedback that makes the UI feel responsive and polished. They should enhance usability, not slow things down.

**File**: `ui/src/styles/animations.css`

```css
/*
 * Animation Tokens and Keyframes
 * 
 * Timing guidelines (from ui-design.md):
 * - Hover states: 100ms
 * - Panel slide-in: 150ms
 * - Modal appear: 150ms (fade + scale from 0.95)
 * - Drag feedback: immediate
 * - Dropdown open: 100ms (fade + slide)
 */

:root {
  /* Duration tokens */
  --duration-fast: 100ms;
  --duration-normal: 150ms;
  --duration-slow: 200ms;

  /* Easing tokens */
  --ease-default: cubic-bezier(0.4, 0, 0.2, 1);
  --ease-in: cubic-bezier(0.4, 0, 1, 1);
  --ease-out: cubic-bezier(0, 0, 0.2, 1);
  --ease-spring: cubic-bezier(0.34, 1.56, 0.64, 1);
}

/* =============================================================================
   Keyframe Animations
   ============================================================================= */

/* Fade in (modals, overlays) */
@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes fadeOut {
  from { opacity: 1; }
  to { opacity: 0; }
}

/* Scale + fade (modals, popovers) */
@keyframes scaleIn {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

@keyframes scaleOut {
  from {
    opacity: 1;
    transform: scale(1);
  }
  to {
    opacity: 0;
    transform: scale(0.95);
  }
}

/* Slide from right (detail panel) */
@keyframes slideInFromRight {
  from {
    opacity: 0;
    transform: translateX(100%);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

@keyframes slideOutToRight {
  from {
    opacity: 1;
    transform: translateX(0);
  }
  to {
    opacity: 0;
    transform: translateX(100%);
  }
}

/* Slide down (dropdowns) */
@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-8px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Slide up (mobile bottom sheet) */
@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(100%);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Skeleton loading pulse */
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

/* =============================================================================
   Utility Classes
   ============================================================================= */

/* Apply to elements that should transition on hover */
.transition-colors {
  transition: color var(--duration-fast) ease,
              background-color var(--duration-fast) ease,
              border-color var(--duration-fast) ease;
}

.transition-transform {
  transition: transform var(--duration-fast) ease;
}

.transition-opacity {
  transition: opacity var(--duration-fast) ease;
}

.transition-all {
  transition: all var(--duration-fast) ease;
}

/* Animation classes */
.animate-fade-in {
  animation: fadeIn var(--duration-normal) ease-out;
}

.animate-scale-in {
  animation: scaleIn var(--duration-normal) ease-out;
}

.animate-slide-in-right {
  animation: slideInFromRight var(--duration-normal) ease-out;
}

.animate-slide-down {
  animation: slideDown var(--duration-fast) ease-out;
}

.animate-slide-up {
  animation: slideUp var(--duration-normal) ease-out;
}

/* Skeleton loading */
.skeleton {
  background: linear-gradient(
    90deg,
    var(--bg-card) 25%,
    var(--bg-card-hover) 50%,
    var(--bg-card) 75%
  );
  background-size: 200% 100%;
  animation: pulse 1.5s ease-in-out infinite;
  border-radius: var(--radius-sm);
}

/* =============================================================================
   Component-Specific Animations
   ============================================================================= */

/* Task Card Hover */
.task-card {
  transition: background-color var(--duration-fast) ease,
              border-color var(--duration-fast) ease,
              box-shadow var(--duration-fast) ease;
}

/* Task Card Drag State */
.task-card.dragging {
  box-shadow: var(--shadow-drag, 0 8px 16px rgba(0, 0, 0, 0.25));
  transform: scale(1.02);
  cursor: grabbing;
}

/* Detail Panel */
.detail-panel {
  animation: slideInFromRight var(--duration-normal) ease-out;
}

.detail-panel.closing {
  animation: slideOutToRight var(--duration-normal) ease-out;
}

/* Modal/Dialog */
.modal-overlay {
  animation: fadeIn var(--duration-normal) ease-out;
}

.modal-content {
  animation: scaleIn var(--duration-normal) ease-out;
}

/* Dropdown Menu */
.dropdown-menu {
  animation: slideDown var(--duration-fast) ease-out;
  transform-origin: top;
}

/* Command Palette */
.command-palette {
  animation: scaleIn var(--duration-normal) ease-out;
}

/* Toast Notification */
.toast {
  animation: slideDown var(--duration-normal) ease-out;
}

.toast.exiting {
  animation: fadeOut var(--duration-fast) ease-out;
}

/* =============================================================================
   Reduced Motion
   ============================================================================= */

/*
 * Respect user's motion preferences.
 * When reduced motion is preferred, use instant transitions.
 */
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

**Steps**:

1. Create the file:
   ```bash
   touch ui/src/styles/animations.css
   ```

2. Open in your editor and paste the CSS above.

3. Import in your main CSS file:
   ```css
   @import './animations.css';
   ```

4. Apply animation classes to components:
   - Detail panel: Add `detail-panel` class
   - Modals: Add `modal-overlay` and `modal-content` classes
   - Dropdowns: Add `dropdown-menu` class
   - Task cards: Add `task-card` class

5. Verify animations work:
   - Open a task detail panel - should slide in from right
   - Open command palette - should scale in
   - Hover over task cards - should transition smoothly
   - Drag a task - should lift with shadow

**Important Notes**:
- The `prefers-reduced-motion` media query respects users who get motion sickness
- Animations should be subtle (150ms or less) to feel responsive
- Drag animations should be immediate (no delay) for direct manipulation feel

---

### 7.6 Implement Responsive Layout

**What**: Add responsive styles and behavior for mobile, tablet, and desktop.

**Why**: Users should have full functionality regardless of screen size. Mobile users are common, even for developer tools.

**File**: `ui/src/styles/responsive.css`

```css
/*
 * Responsive Breakpoints (from ui-design.md):
 * - Mobile: < 640px
 * - Tablet: 640px - 1024px
 * - Desktop: > 1024px
 */

/* =============================================================================
   Breakpoint Variables (for reference, use media queries directly)
   ============================================================================= */
/*
  --breakpoint-sm: 640px;   Tablet starts
  --breakpoint-md: 768px;   Tablet mid
  --breakpoint-lg: 1024px;  Desktop starts
  --breakpoint-xl: 1280px;  Large desktop
*/

/* =============================================================================
   Mobile Styles (< 640px)
   ============================================================================= */

@media (max-width: 639px) {
  /* Hide sidebar by default on mobile */
  .sidebar {
    position: fixed;
    left: 0;
    top: 0;
    bottom: 0;
    z-index: 50;
    transform: translateX(-100%);
    transition: transform var(--duration-normal) ease-out;
    width: 280px;
    background: var(--bg-sidebar);
    border-right: 1px solid var(--border-default);
  }

  .sidebar.open {
    transform: translateX(0);
  }

  /* Sidebar backdrop */
  .sidebar-backdrop {
    position: fixed;
    inset: 0;
    background: var(--bg-overlay);
    z-index: 49;
    opacity: 0;
    pointer-events: none;
    transition: opacity var(--duration-normal) ease;
  }

  .sidebar-backdrop.visible {
    opacity: 1;
    pointer-events: auto;
  }

  /* Board columns scroll horizontally */
  .board {
    overflow-x: auto;
    scroll-snap-type: x mandatory;
    -webkit-overflow-scrolling: touch;
    padding: var(--space-4);
    gap: var(--space-3);
  }

  .column {
    flex-shrink: 0;
    width: 85vw;
    max-width: 320px;
    scroll-snap-align: start;
  }

  /* Detail panel becomes full screen */
  .detail-panel {
    position: fixed;
    inset: 0;
    width: 100%;
    max-width: none;
    border-radius: 0;
  }

  /* Command palette becomes bottom sheet */
  .command-palette-container {
    align-items: flex-end;
  }

  .command-palette {
    width: 100%;
    max-width: none;
    max-height: 70vh;
    border-radius: var(--radius-lg) var(--radius-lg) 0 0;
    animation: slideUp var(--duration-normal) ease-out;
  }

  /* Larger touch targets */
  .task-card {
    padding: var(--space-4);
    min-height: 60px;
  }

  button, 
  .button,
  [role="button"] {
    min-height: 44px;
    min-width: 44px;
  }

  /* Mobile menu button (hamburger) */
  .mobile-menu-button {
    display: flex;
  }

  /* Hide desktop-only elements */
  .desktop-only {
    display: none !important;
  }
}

/* =============================================================================
   Tablet Styles (640px - 1024px)
   ============================================================================= */

@media (min-width: 640px) and (max-width: 1023px) {
  /* Collapsible sidebar */
  .sidebar {
    width: 60px;
    transition: width var(--duration-normal) ease;
    overflow: hidden;
  }

  .sidebar.expanded {
    width: 240px;
  }

  .sidebar.expanded .sidebar-label {
    opacity: 1;
  }

  .sidebar-label {
    opacity: 0;
    white-space: nowrap;
    transition: opacity var(--duration-fast) ease;
  }

  /* Show 2-3 columns */
  .board {
    overflow-x: auto;
    padding: var(--space-4);
  }

  .column {
    min-width: 260px;
    flex-shrink: 0;
  }

  /* Detail panel is narrower */
  .detail-panel {
    width: 350px;
  }

  /* Hide mobile-only elements */
  .mobile-only {
    display: none !important;
  }
}

/* =============================================================================
   Desktop Styles (>= 1024px)
   ============================================================================= */

@media (min-width: 1024px) {
  /* Full sidebar */
  .sidebar {
    width: 240px;
    flex-shrink: 0;
  }

  /* Board fills remaining space */
  .board {
    flex: 1;
    overflow-x: auto;
    padding: var(--space-4);
  }

  .column {
    width: 280px;
    flex-shrink: 0;
  }

  /* Detail panel */
  .detail-panel {
    width: 400px;
    min-width: 350px;
    max-width: 500px;
    resize: horizontal;
    overflow: auto;
  }

  /* Hide mobile/tablet-only elements */
  .mobile-only,
  .tablet-only,
  .mobile-menu-button {
    display: none !important;
  }
}

/* =============================================================================
   Large Desktop Styles (>= 1280px)
   ============================================================================= */

@media (min-width: 1280px) {
  .column {
    width: 300px;
  }

  .detail-panel {
    width: 450px;
  }
}

/* =============================================================================
   Touch Device Adjustments
   ============================================================================= */

@media (hover: none) {
  /* Remove hover effects on touch devices (they cause sticky hover states) */
  .task-card:hover {
    background-color: var(--bg-card);
  }

  /* Larger touch targets */
  .dropdown-item,
  .menu-item {
    padding: var(--space-3) var(--space-4);
    min-height: 48px;
  }
}

/* =============================================================================
   Print Styles
   ============================================================================= */

@media print {
  .sidebar,
  .detail-panel,
  .command-palette-container {
    display: none !important;
  }

  .board {
    overflow: visible;
    display: block;
  }

  .column {
    break-inside: avoid;
    margin-bottom: var(--space-6);
  }

  .task-card {
    break-inside: avoid;
  }
}
```

**Steps**:

1. Create the file:
   ```bash
   touch ui/src/styles/responsive.css
   ```

2. Open in your editor and paste the CSS above.

3. Import in your main CSS file:
   ```css
   @import './responsive.css';
   ```

4. Add necessary classes to your components:
   - Sidebar: `sidebar` class, toggle `open` class on mobile
   - Board: `board` class
   - Columns: `column` class
   - Detail panel: `detail-panel` class

5. Create mobile menu toggle. Add to your layout:
   ```tsx
   const [sidebarOpen, setSidebarOpen] = useState(false);
   
   return (
     <>
       {/* Mobile menu button */}
       <button 
         className="mobile-menu-button"
         onClick={() => setSidebarOpen(true)}
         aria-label="Open menu"
       >
         &#9776;
       </button>
       
       {/* Sidebar backdrop (mobile) */}
       <div 
         className={`sidebar-backdrop ${sidebarOpen ? 'visible' : ''}`}
         onClick={() => setSidebarOpen(false)}
       />
       
       {/* Sidebar */}
       <aside className={`sidebar ${sidebarOpen ? 'open' : ''}`}>
         {/* sidebar content */}
       </aside>
     </>
   );
   ```

6. Verify responsive behavior:
   - Resize browser window to different sizes
   - Mobile (<640px): Sidebar hidden, horizontal scroll, bottom sheet modals
   - Tablet (640-1024px): Collapsible sidebar, 2-3 columns visible
   - Desktop (>1024px): Full layout with all panels

7. Test on actual mobile device or emulator:
   - Chrome DevTools > Toggle device toolbar
   - Test touch interactions (swipe, tap)
   - Verify touch targets are at least 44px

---

### 7.7 Add Loading States

**What**: Create skeleton loaders and loading indicators for async operations.

**Why**: Loading states communicate that something is happening, reducing perceived wait time and preventing user frustration.

**File**: `ui/src/components/Skeleton.tsx`

```tsx
import './Skeleton.css';

interface SkeletonProps {
  /** Width of skeleton (CSS value) */
  width?: string;
  /** Height of skeleton (CSS value) */
  height?: string;
  /** Border radius (CSS value) */
  radius?: string;
  /** Additional class name */
  className?: string;
}

/**
 * Generic skeleton loader component.
 * 
 * @example
 * <Skeleton width="100%" height="20px" />
 */
export function Skeleton({ 
  width = '100%', 
  height = '16px', 
  radius = 'var(--radius-sm)',
  className = ''
}: SkeletonProps) {
  return (
    <div 
      className={`skeleton ${className}`}
      style={{ width, height, borderRadius: radius }}
      role="status"
      aria-label="Loading..."
    />
  );
}

/**
 * Skeleton for a task card.
 * Matches the visual structure of TaskCard component.
 */
export function TaskCardSkeleton() {
  return (
    <div className="task-card-skeleton">
      <div className="task-card-skeleton-header">
        <Skeleton width="12px" height="12px" radius="50%" />
        <Skeleton width="60px" height="12px" />
      </div>
      <Skeleton width="100%" height="16px" className="task-card-skeleton-title" />
      <Skeleton width="70%" height="14px" />
      <div className="task-card-skeleton-footer">
        <Skeleton width="50px" height="20px" radius="var(--radius-sm)" />
        <Skeleton width="24px" height="16px" />
      </div>
    </div>
  );
}

/**
 * Skeleton for a column of tasks.
 * Shows header and multiple task card skeletons.
 */
export function ColumnSkeleton({ cardCount = 3 }: { cardCount?: number }) {
  return (
    <div className="column-skeleton">
      <div className="column-skeleton-header">
        <Skeleton width="100px" height="16px" />
        <Skeleton width="24px" height="16px" radius="var(--radius-sm)" />
      </div>
      <div className="column-skeleton-cards">
        {Array.from({ length: cardCount }).map((_, i) => (
          <TaskCardSkeleton key={i} />
        ))}
      </div>
    </div>
  );
}

/**
 * Full board skeleton for initial load.
 */
export function BoardSkeleton() {
  return (
    <div className="board-skeleton" role="status" aria-label="Loading board...">
      <ColumnSkeleton cardCount={2} />
      <ColumnSkeleton cardCount={4} />
      <ColumnSkeleton cardCount={2} />
      <ColumnSkeleton cardCount={1} />
      <ColumnSkeleton cardCount={3} />
    </div>
  );
}
```

**File**: `ui/src/components/Skeleton.css`

```css
/* Skeleton Loading Styles */

.skeleton {
  background: linear-gradient(
    90deg,
    var(--bg-card-hover) 25%,
    var(--bg-card) 50%,
    var(--bg-card-hover) 75%
  );
  background-size: 200% 100%;
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

@keyframes skeleton-pulse {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

/* Task Card Skeleton */
.task-card-skeleton {
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: var(--space-3);
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.task-card-skeleton-header {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.task-card-skeleton-title {
  margin-top: var(--space-1);
}

.task-card-skeleton-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: var(--space-2);
}

/* Column Skeleton */
.column-skeleton {
  width: 280px;
  flex-shrink: 0;
}

.column-skeleton-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-2) var(--space-3);
  margin-bottom: var(--space-2);
}

.column-skeleton-cards {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

/* Board Skeleton */
.board-skeleton {
  display: flex;
  gap: var(--space-4);
  padding: var(--space-4);
  overflow-x: auto;
}
```

**Steps**:

1. Create the component files:
   ```bash
   touch ui/src/components/Skeleton.tsx
   touch ui/src/components/Skeleton.css
   ```

2. Open in your editor and paste the code above.

3. Use skeletons in your board component:
   ```tsx
   function Board() {
     const { tasks, loading } = useTasks();
     
     if (loading) {
       return <BoardSkeleton />;
     }
     
     return (
       // actual board content
     );
   }
   ```

4. Add inline skeleton for individual loading states:
   ```tsx
   {isUpdating ? (
     <Skeleton width="80px" height="20px" />
   ) : (
     <span>{task.priority}</span>
   )}
   ```

5. Verify skeletons work:
   - Add artificial delay to useTasks hook: `await new Promise(r => setTimeout(r, 2000))`
   - Refresh page and observe skeleton animation
   - Remove the delay when done testing

---

### 7.8 Add Toast Notifications

**What**: Create a toast notification system for user feedback.

**Why**: Toasts provide non-intrusive feedback for actions like creating tasks, errors, or successful operations.

**File**: `ui/src/components/Toast.tsx`

```tsx
import { createContext, useContext, useState, useCallback, type ReactNode } from 'react';
import './Toast.css';

type ToastType = 'success' | 'error' | 'info' | 'warning';

interface Toast {
  id: string;
  type: ToastType;
  message: string;
  action?: {
    label: string;
    onClick: () => void;
  };
}

interface ToastContextValue {
  toast: (type: ToastType, message: string, action?: Toast['action']) => void;
  success: (message: string) => void;
  error: (message: string) => void;
  info: (message: string) => void;
}

const ToastContext = createContext<ToastContextValue | undefined>(undefined);

const AUTO_DISMISS_MS = 3000;

interface ToastProviderProps {
  children: ReactNode;
}

export function ToastProvider({ children }: ToastProviderProps) {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const removeToast = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const toast = useCallback((
    type: ToastType, 
    message: string, 
    action?: Toast['action']
  ) => {
    const id = `${Date.now()}-${Math.random().toString(36).slice(2)}`;
    
    setToasts((prev) => [...prev, { id, type, message, action }]);

    // Auto dismiss (unless there's an action button)
    if (!action) {
      setTimeout(() => removeToast(id), AUTO_DISMISS_MS);
    }
  }, [removeToast]);

  const success = useCallback((message: string) => toast('success', message), [toast]);
  const error = useCallback((message: string) => toast('error', message), [toast]);
  const info = useCallback((message: string) => toast('info', message), [toast]);

  return (
    <ToastContext.Provider value={{ toast, success, error, info }}>
      {children}
      <ToastContainer toasts={toasts} onDismiss={removeToast} />
    </ToastContext.Provider>
  );
}

interface ToastContainerProps {
  toasts: Toast[];
  onDismiss: (id: string) => void;
}

function ToastContainer({ toasts, onDismiss }: ToastContainerProps) {
  if (toasts.length === 0) return null;

  return (
    <div className="toast-container" role="region" aria-label="Notifications">
      {toasts.map((toast) => (
        <ToastItem key={toast.id} toast={toast} onDismiss={onDismiss} />
      ))}
    </div>
  );
}

interface ToastItemProps {
  toast: Toast;
  onDismiss: (id: string) => void;
}

function ToastItem({ toast, onDismiss }: ToastItemProps) {
  const icons: Record<ToastType, string> = {
    success: '&#10003;', // checkmark
    error: '&#10007;',   // X
    info: 'i',
    warning: '!',
  };

  return (
    <div 
      className={`toast toast-${toast.type}`} 
      role="alert"
    >
      <span 
        className="toast-icon" 
        dangerouslySetInnerHTML={{ __html: icons[toast.type] }} 
      />
      <span className="toast-message">{toast.message}</span>
      {toast.action && (
        <button 
          className="toast-action"
          onClick={() => {
            toast.action?.onClick();
            onDismiss(toast.id);
          }}
        >
          {toast.action.label}
        </button>
      )}
      <button 
        className="toast-dismiss"
        onClick={() => onDismiss(toast.id)}
        aria-label="Dismiss"
      >
        &times;
      </button>
    </div>
  );
}

/**
 * Hook to show toast notifications.
 * 
 * @example
 * const { success, error } = useToast();
 * 
 * // Show success toast
 * success('Task created!');
 * 
 * // Show error toast
 * error('Failed to save changes');
 * 
 * // Show toast with action
 * toast('info', 'Task archived', {
 *   label: 'Undo',
 *   onClick: () => restoreTask(id)
 * });
 */
export function useToast(): ToastContextValue {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error('useToast must be used within a ToastProvider');
  }
  return context;
}
```

**File**: `ui/src/components/Toast.css`

```css
/* Toast Notification Styles */

.toast-container {
  position: fixed;
  top: var(--space-4);
  right: var(--space-4);
  z-index: 200;
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  max-width: 400px;
  pointer-events: none;
}

.toast {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  background: var(--bg-card);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
  animation: toast-enter var(--duration-normal) ease-out;
  pointer-events: auto;
}

@keyframes toast-enter {
  from {
    opacity: 0;
    transform: translateX(100%);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.toast.exiting {
  animation: toast-exit var(--duration-fast) ease-out forwards;
}

@keyframes toast-exit {
  to {
    opacity: 0;
    transform: translateX(100%);
  }
}

/* Toast Icons */
.toast-icon {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: bold;
  flex-shrink: 0;
}

.toast-success .toast-icon {
  background: var(--status-done);
  color: white;
}

.toast-error .toast-icon {
  background: var(--priority-urgent);
  color: white;
}

.toast-info .toast-icon {
  background: var(--accent);
  color: white;
}

.toast-warning .toast-icon {
  background: var(--priority-medium);
  color: white;
}

/* Toast Content */
.toast-message {
  flex: 1;
  color: var(--text-primary);
  font-size: var(--text-sm);
  line-height: var(--leading-normal);
}

.toast-action {
  background: none;
  border: none;
  color: var(--accent);
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  cursor: pointer;
  padding: var(--space-1) var(--space-2);
  border-radius: var(--radius-sm);
  transition: background-color var(--duration-fast) ease;
}

.toast-action:hover {
  background: rgba(var(--accent-rgb), 0.1);
}

.toast-dismiss {
  background: none;
  border: none;
  color: var(--text-muted);
  font-size: 18px;
  cursor: pointer;
  padding: var(--space-1);
  line-height: 1;
  border-radius: var(--radius-sm);
  transition: color var(--duration-fast) ease;
}

.toast-dismiss:hover {
  color: var(--text-primary);
}

/* Mobile positioning */
@media (max-width: 639px) {
  .toast-container {
    top: auto;
    bottom: var(--space-4);
    left: var(--space-4);
    right: var(--space-4);
    max-width: none;
  }

  .toast {
    animation-name: toast-enter-mobile;
  }

  @keyframes toast-enter-mobile {
    from {
      opacity: 0;
      transform: translateY(100%);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
}
```

**Steps**:

1. Create the component files:
   ```bash
   touch ui/src/components/Toast.tsx
   touch ui/src/components/Toast.css
   ```

2. Open in your editor and paste the code above.

3. Wrap your app with ToastProvider:
   ```tsx
   import { ToastProvider } from './components/Toast';
   
   function App() {
     return (
       <ThemeProvider>
         <ToastProvider>
           {/* rest of app */}
         </ToastProvider>
       </ThemeProvider>
     );
   }
   ```

4. Use toasts in your components:
   ```tsx
   const { success, error } = useToast();
   
   const handleCreateTask = async () => {
     try {
       await createTask(data);
       success('Task created!');
     } catch (e) {
       error('Failed to create task');
     }
   };
   ```

5. Verify toasts work:
   - Trigger a success toast - should appear top-right
   - Should auto-dismiss after 3 seconds
   - Click dismiss button - should close immediately
   - Test on mobile - should appear at bottom

---

### 7.9 Improve Drag and Drop

**What**: Enhance drag and drop with better visual feedback, auto-scroll, and touch support.

**Why**: Drag and drop is a core interaction. Good feedback makes it feel direct and responsive.

**File**: `ui/src/components/Board.tsx` (update existing)

Add these enhancements to your existing drag and drop implementation:

```tsx
import { useState, useRef, useCallback } from 'react';
import {
  DndContext,
  DragOverlay,
  closestCenter,
  PointerSensor,
  TouchSensor,
  useSensor,
  useSensors,
  type DragStartEvent,
  type DragEndEvent,
  type DragOverEvent,
} from '@dnd-kit/core';
import {
  SortableContext,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable';

// ... existing imports

export function Board() {
  const { tasks, moveTask } = useTasks();
  const [activeTask, setActiveTask] = useState<Task | null>(null);
  const [overId, setOverId] = useState<string | null>(null);

  // Configure sensors for both pointer and touch
  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        // Require 8px movement before starting drag
        // This prevents accidental drags when clicking
        distance: 8,
      },
    }),
    useSensor(TouchSensor, {
      activationConstraint: {
        // For touch, require 250ms hold OR 8px movement
        delay: 250,
        tolerance: 8,
      },
    })
  );

  const handleDragStart = useCallback((event: DragStartEvent) => {
    const task = tasks.find((t) => t.id === event.active.id);
    if (task) {
      setActiveTask(task);
      // Add dragging class to body for global styles
      document.body.classList.add('is-dragging');
    }
  }, [tasks]);

  const handleDragOver = useCallback((event: DragOverEvent) => {
    setOverId(event.over?.id as string | null);
  }, []);

  const handleDragEnd = useCallback(async (event: DragEndEvent) => {
    const { active, over } = event;
    
    // Clean up
    setActiveTask(null);
    setOverId(null);
    document.body.classList.remove('is-dragging');

    if (!over) return;

    const taskId = active.id as string;
    const overId = over.id as string;

    // Determine target column and position
    // This logic depends on your column/task structure
    const targetColumn = getColumnFromOverId(overId);
    const targetPosition = getPositionFromOverId(overId, tasks);

    if (targetColumn && targetPosition !== null) {
      await moveTask(taskId, targetColumn, targetPosition);
    }
  }, [tasks, moveTask]);

  const handleDragCancel = useCallback(() => {
    setActiveTask(null);
    setOverId(null);
    document.body.classList.remove('is-dragging');
  }, []);

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragStart={handleDragStart}
      onDragOver={handleDragOver}
      onDragEnd={handleDragEnd}
      onDragCancel={handleDragCancel}
    >
      <div className="board">
        {COLUMNS.map((column) => (
          <Column
            key={column}
            column={column}
            tasks={tasksByColumn[column]}
            isOver={overId === column}
          />
        ))}
      </div>

      {/* Drag overlay - renders the dragged item */}
      <DragOverlay>
        {activeTask ? (
          <TaskCard task={activeTask} isDragOverlay />
        ) : null}
      </DragOverlay>
    </DndContext>
  );
}
```

**File**: `ui/src/styles/drag-drop.css`

```css
/* Drag and Drop Styles */

/* Body state during drag */
body.is-dragging {
  cursor: grabbing !important;
  user-select: none;
}

body.is-dragging * {
  cursor: grabbing !important;
}

/* Task card grab cursor */
.task-card {
  cursor: grab;
}

.task-card:active {
  cursor: grabbing;
}

/* Dragging task (original position) */
.task-card.dragging {
  opacity: 0.5;
}

/* Drag overlay (the moving clone) */
.task-card.drag-overlay {
  box-shadow: var(--shadow-drag);
  transform: rotate(3deg) scale(1.02);
  cursor: grabbing;
}

/* Drop target indicator */
.column.drag-over {
  background: rgba(var(--accent-rgb), 0.05);
}

/* Drop position indicator */
.drop-indicator {
  height: 2px;
  background: var(--accent);
  border-radius: 1px;
  margin: var(--space-1) 0;
  animation: pulse-subtle 1s ease-in-out infinite;
}

@keyframes pulse-subtle {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

/* Empty column drop area */
.column-empty-drop {
  min-height: 100px;
  border: 2px dashed var(--border-subtle);
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
  font-size: var(--text-sm);
  transition: border-color var(--duration-fast) ease,
              background-color var(--duration-fast) ease;
}

.column-empty-drop.drag-over {
  border-color: var(--accent);
  background: rgba(var(--accent-rgb), 0.05);
}

/* Auto-scroll zones (visual indicator) */
.auto-scroll-zone {
  position: absolute;
  left: 0;
  right: 0;
  height: 60px;
  pointer-events: none;
  opacity: 0;
  transition: opacity var(--duration-fast) ease;
}

.auto-scroll-zone.top {
  top: 0;
  background: linear-gradient(
    to bottom,
    rgba(var(--accent-rgb), 0.1),
    transparent
  );
}

.auto-scroll-zone.bottom {
  bottom: 0;
  background: linear-gradient(
    to top,
    rgba(var(--accent-rgb), 0.1),
    transparent
  );
}

.auto-scroll-zone.active {
  opacity: 1;
}

/* Touch improvements */
@media (hover: none) {
  .task-card {
    /* Prevent long-press context menu on mobile */
    -webkit-touch-callout: none;
    -webkit-user-select: none;
  }

  /* Visual feedback for touch hold */
  .task-card.touch-active {
    transform: scale(0.98);
    transition: transform var(--duration-fast) ease;
  }
}
```

**Steps**:

1. Update your Board component with the enhanced drag handlers.

2. Create the drag-drop CSS file:
   ```bash
   touch ui/src/styles/drag-drop.css
   ```

3. Import in your main CSS:
   ```css
   @import './drag-drop.css';
   ```

4. Update TaskCard to accept and use `isDragOverlay` prop:
   ```tsx
   interface TaskCardProps {
     task: Task;
     isDragOverlay?: boolean;
   }
   
   export function TaskCard({ task, isDragOverlay }: TaskCardProps) {
     return (
       <div className={`task-card ${isDragOverlay ? 'drag-overlay' : ''}`}>
         {/* ... */}
       </div>
     );
   }
   ```

5. Verify drag and drop improvements:
   - Drag a task - should lift with shadow and slight rotation
   - Original position should fade to 50% opacity
   - Drop target column should highlight
   - Test on touch device - should work with long-press

---

### 7.10 Add Keyboard Navigation Visual Feedback

**What**: Add visible focus indicators and selection highlights for keyboard navigation.

**Why**: Keyboard users (and power users using shortcuts) need clear visual feedback about which element is focused or selected.

**File**: `ui/src/styles/focus.css`

```css
/*
 * Focus and Selection Styles
 * 
 * Accessibility requirements:
 * - Focus must be visible (WCAG 2.4.7)
 * - Focus should have sufficient contrast
 * - Selection state should be distinct from focus
 */

/* =============================================================================
   Global Focus Styles
   ============================================================================= */

/* Remove default focus outline, we'll add our own */
*:focus {
  outline: none;
}

/* Visible focus ring for keyboard navigation */
*:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

/* Special handling for inputs */
input:focus-visible,
textarea:focus-visible,
select:focus-visible {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(var(--accent-rgb), 0.2);
}

/* =============================================================================
   Task Card Focus/Selection States
   ============================================================================= */

.task-card:focus-visible {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 0 2px var(--accent);
}

/* Selected state (via keyboard or click) */
.task-card.selected {
  background: var(--bg-card-selected);
  border-color: var(--accent);
}

/* Multi-select state */
.task-card.multi-selected {
  background: var(--bg-card-selected);
  border-color: var(--accent);
}

.task-card.multi-selected::before {
  content: '';
  position: absolute;
  left: var(--space-2);
  top: 50%;
  transform: translateY(-50%);
  width: 16px;
  height: 16px;
  background: var(--accent);
  border-radius: 3px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.task-card.multi-selected::after {
  content: '\2713'; /* checkmark */
  position: absolute;
  left: calc(var(--space-2) + 3px);
  top: 50%;
  transform: translateY(-50%);
  color: white;
  font-size: 10px;
  font-weight: bold;
}

/* =============================================================================
   Column Focus State
   ============================================================================= */

.column.focused {
  background: rgba(var(--accent-rgb), 0.03);
}

.column.focused .column-header {
  color: var(--accent);
}

/* =============================================================================
   Navigation Indicator
   ============================================================================= */

/* Current column indicator during J/K navigation */
.column-nav-indicator {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: var(--accent);
  border-radius: 0 0 2px 2px;
  opacity: 0;
  transition: opacity var(--duration-fast) ease;
}

.column.focused .column-nav-indicator {
  opacity: 1;
}

/* =============================================================================
   Selection Count Indicator
   ============================================================================= */

.selection-count {
  position: fixed;
  bottom: var(--space-4);
  left: 50%;
  transform: translateX(-50%);
  background: var(--bg-card);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  padding: var(--space-2) var(--space-4);
  box-shadow: var(--shadow-lg);
  display: flex;
  align-items: center;
  gap: var(--space-3);
  z-index: 50;
  animation: slideUp var(--duration-normal) ease-out;
}

.selection-count-text {
  color: var(--text-primary);
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
}

.selection-count-actions {
  display: flex;
  gap: var(--space-2);
}

.selection-count-btn {
  background: var(--bg-input);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  padding: var(--space-1) var(--space-2);
  color: var(--text-primary);
  font-size: var(--text-xs);
  cursor: pointer;
  transition: background-color var(--duration-fast) ease;
}

.selection-count-btn:hover {
  background: var(--bg-card-hover);
}

/* =============================================================================
   Skip Link (Accessibility)
   ============================================================================= */

.skip-link {
  position: absolute;
  top: -100%;
  left: 0;
  background: var(--accent);
  color: white;
  padding: var(--space-2) var(--space-4);
  z-index: 1000;
  transition: top var(--duration-fast) ease;
}

.skip-link:focus {
  top: 0;
}

/* =============================================================================
   High Contrast Mode
   ============================================================================= */

@media (prefers-contrast: high) {
  *:focus-visible {
    outline-width: 3px;
    outline-color: currentColor;
  }

  .task-card.selected,
  .task-card:focus-visible {
    outline: 3px solid currentColor;
    outline-offset: 0;
  }
}
```

**Steps**:

1. Create the focus styles file:
   ```bash
   touch ui/src/styles/focus.css
   ```

2. Import in your main CSS:
   ```css
   @import './focus.css';
   ```

3. Add skip link to your layout for accessibility:
   ```tsx
   <a href="#main-content" className="skip-link">
     Skip to main content
   </a>
   <main id="main-content">
     {/* board content */}
   </main>
   ```

4. Implement selection state in your task store or context:
   ```tsx
   const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
   
   const selectTask = (id: string, multi = false) => {
     setSelectedIds(prev => {
       if (multi) {
         const next = new Set(prev);
         if (next.has(id)) {
           next.delete(id);
         } else {
           next.add(id);
         }
         return next;
       }
       return new Set([id]);
     });
   };
   ```

5. Apply selected class to TaskCard:
   ```tsx
   <div className={`task-card ${selectedIds.has(task.id) ? 'selected' : ''}`}>
   ```

6. Add selection count indicator when multiple selected:
   ```tsx
   {selectedIds.size > 1 && (
     <div className="selection-count">
       <span className="selection-count-text">
         {selectedIds.size} tasks selected
       </span>
       <div className="selection-count-actions">
         <button className="selection-count-btn" onClick={clearSelection}>
           Clear
         </button>
         <button className="selection-count-btn" onClick={deleteSelected}>
           Delete
         </button>
       </div>
     </div>
   )}
   ```

7. Verify focus and selection states:
   - Tab through the interface - focus rings should be visible
   - Press arrow keys to navigate tasks - current task should highlight
   - Press X to select multiple - selection indicator should appear
   - Test with keyboard only (no mouse)

---

### 7.11 Optimize Performance

**What**: Add virtualization for long lists and optimize re-renders.

**Why**: As users add more tasks, the board needs to stay performant. Virtualization renders only visible items, dramatically reducing DOM nodes.

**File**: `ui/src/components/VirtualizedColumn.tsx`

```tsx
import { useRef, useCallback } from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';
import { SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable';
import { TaskCard } from './TaskCard';
import type { Task } from '../hooks/usePocketBase';

interface VirtualizedColumnProps {
  column: string;
  tasks: Task[];
  selectedIds: Set<string>;
  onSelectTask: (id: string) => void;
}

/**
 * A column that virtualizes its task list.
 * Only renders visible tasks, keeping DOM size small.
 * 
 * Use this when a column might have many tasks (>50).
 */
export function VirtualizedColumn({ 
  column, 
  tasks, 
  selectedIds,
  onSelectTask 
}: VirtualizedColumnProps) {
  const parentRef = useRef<HTMLDivElement>(null);

  // Estimate task card height (adjust based on your design)
  const estimateSize = useCallback(() => 80, []);

  const virtualizer = useVirtualizer({
    count: tasks.length,
    getScrollElement: () => parentRef.current,
    estimateSize,
    overscan: 5, // Render 5 extra items above/below viewport
  });

  const virtualItems = virtualizer.getVirtualItems();

  return (
    <div className="column">
      <div className="column-header">
        <span className="column-name">{column}</span>
        <span className="column-count">{tasks.length}</span>
      </div>

      <div 
        ref={parentRef} 
        className="column-content"
        style={{ overflow: 'auto', height: '100%' }}
      >
        <div
          style={{
            height: `${virtualizer.getTotalSize()}px`,
            width: '100%',
            position: 'relative',
          }}
        >
          <SortableContext
            items={tasks.map((t) => t.id)}
            strategy={verticalListSortingStrategy}
          >
            {virtualItems.map((virtualItem) => {
              const task = tasks[virtualItem.index];
              return (
                <div
                  key={task.id}
                  style={{
                    position: 'absolute',
                    top: 0,
                    left: 0,
                    width: '100%',
                    transform: `translateY(${virtualItem.start}px)`,
                  }}
                >
                  <TaskCard
                    task={task}
                    isSelected={selectedIds.has(task.id)}
                    onSelect={() => onSelectTask(task.id)}
                  />
                </div>
              );
            })}
          </SortableContext>
        </div>
      </div>
    </div>
  );
}
```

**Additional Performance Optimizations**:

**1. Memoize expensive computations:**

```tsx
import { useMemo } from 'react';

function Board() {
  const { tasks } = useTasks();

  // Memoize task grouping
  const tasksByColumn = useMemo(() => {
    const grouped: Record<string, Task[]> = {};
    COLUMNS.forEach((col) => (grouped[col] = []));
    
    tasks.forEach((task) => {
      if (grouped[task.column]) {
        grouped[task.column].push(task);
      }
    });
    
    // Sort each column by position
    Object.keys(grouped).forEach((col) => {
      grouped[col].sort((a, b) => a.position - b.position);
    });
    
    return grouped;
  }, [tasks]);

  // ... rest of component
}
```

**2. Memoize TaskCard component:**

```tsx
import { memo } from 'react';

export const TaskCard = memo(function TaskCard({ 
  task, 
  isSelected,
  onSelect 
}: TaskCardProps) {
  // ... component implementation
});
```

**3. Use React.lazy for large components:**

```tsx
import { lazy, Suspense } from 'react';

const TaskDetail = lazy(() => import('./TaskDetail'));

function App() {
  return (
    <Suspense fallback={<div className="skeleton" />}>
      <TaskDetail />
    </Suspense>
  );
}
```

**Steps**:

1. Install the virtualization library:
   ```bash
   cd ui
   npm install @tanstack/react-virtual
   ```

2. Create the VirtualizedColumn component:
   ```bash
   touch ui/src/components/VirtualizedColumn.tsx
   ```

3. Use virtualized columns for columns with many tasks:
   ```tsx
   {tasks.length > 50 ? (
     <VirtualizedColumn column={column} tasks={tasks} {...props} />
   ) : (
     <Column column={column} tasks={tasks} {...props} />
   )}
   ```

4. Add React.memo to TaskCard and other frequently rendered components.

5. Profile performance using React DevTools:
   - Install React DevTools browser extension
   - Open DevTools > Profiler tab
   - Record while scrolling/dragging
   - Look for unnecessary re-renders

6. Verify performance improvements:
   - Create 100+ tasks in one column
   - Scroll should remain smooth (60fps)
   - Dragging should not lag
   - Memory usage should stay reasonable

---

## Verification Checklist

Complete each section in order. Check off each item as you verify it.

### Theme Verification

- [ ] **Dark mode is default**
  
  Load the app fresh - should show dark theme.

- [ ] **Light mode works**
  
  Open Settings > Theme > Light.
  All colors should update.

- [ ] **System preference works**
  
  Set Theme to System.
  Change OS theme.
  App should update automatically.

- [ ] **Theme persists**
  
  Set to Light mode, refresh page.
  Should still be Light mode.

- [ ] **Accent colors work**
  
  Open Settings > Accent Color.
  Click different colors.
  Buttons, focus rings, selections should update.

- [ ] **Accent color persists**
  
  Set accent to Green, refresh page.
  Should still be Green.

### Animation Verification

- [ ] **Panel slide-in**
  
  Click a task - detail panel should slide in from right.

- [ ] **Modal scale-in**
  
  Press `Cmd+K` - command palette should scale in.

- [ ] **Hover transitions**
  
  Hover over task cards - should smoothly change background.

- [ ] **Drag feedback**
  
  Drag a task - should lift with shadow, rotate slightly.

- [ ] **Reduced motion respected**
  
  Set OS to "Reduce motion".
  Animations should be instant.

### Responsive Verification

- [ ] **Mobile layout (<640px)**
  
  Resize browser to mobile width.
  - Sidebar should be hidden
  - Board should scroll horizontally
  - Detail panel should be full screen

- [ ] **Tablet layout (640-1024px)**
  
  Resize browser to tablet width.
  - Sidebar should be collapsible
  - 2-3 columns should be visible

- [ ] **Desktop layout (>1024px)**
  
  Full browser width.
  - Sidebar should be visible
  - All columns should fit
  - Detail panel should slide in

- [ ] **Touch targets**
  
  On touch device or emulator:
  - Buttons should be easy to tap (44px min)
  - No accidental taps on small elements

### Loading & Feedback Verification

- [ ] **Skeleton loaders appear**
  
  Add delay to useTasks hook, refresh.
  Board skeleton should animate.

- [ ] **Toast notifications work**
  
  Create a task.
  Success toast should appear.

- [ ] **Toast auto-dismisses**
  
  Toast should disappear after 3 seconds.

- [ ] **Toast action works**
  
  If you have an undo toast:
  Clicking "Undo" should work and dismiss toast.

### Keyboard & Focus Verification

- [ ] **Focus rings visible**
  
  Tab through the interface.
  Focus rings should be clear.

- [ ] **Task selection visible**
  
  Press Enter on a task.
  Selected task should be highlighted.

- [ ] **Multi-select indicator**
  
  Select multiple tasks with X.
  Selection count should appear at bottom.

- [ ] **Skip link works**
  
  Press Tab immediately after page load.
  "Skip to main content" should appear.

### Performance Verification

- [ ] **Large board scrolls smoothly**
  
  Create 100 tasks in one column.
  Scrolling should be smooth (60fps).

- [ ] **Dragging doesn't lag**
  
  With 100+ tasks, drag a task.
  Should move smoothly without lag.

- [ ] **No excessive re-renders**
  
  Open React DevTools Profiler.
  Scroll and interact.
  Check for unexpected re-renders.

---

## File Summary

| File | Lines (approx) | Purpose |
|------|----------------|---------|
| `ui/src/styles/light.css` | 30 | Light mode color tokens |
| `ui/src/contexts/ThemeContext.tsx` | 100 | Theme state management |
| `ui/src/components/Settings.tsx` | 120 | Settings panel component |
| `ui/src/components/Settings.css` | 120 | Settings panel styles |
| `ui/src/hooks/useAccentColor.ts` | 50 | Accent color hook |
| `ui/src/styles/animations.css` | 150 | Animation keyframes and utilities |
| `ui/src/styles/responsive.css` | 180 | Responsive breakpoint styles |
| `ui/src/components/Skeleton.tsx` | 80 | Loading skeleton components |
| `ui/src/components/Skeleton.css` | 60 | Skeleton styles |
| `ui/src/components/Toast.tsx` | 130 | Toast notification system |
| `ui/src/components/Toast.css` | 100 | Toast styles |
| `ui/src/styles/drag-drop.css` | 90 | Drag and drop styles |
| `ui/src/styles/focus.css` | 140 | Focus and selection styles |
| `ui/src/components/VirtualizedColumn.tsx` | 70 | Virtualized list column |

**Total new code**: ~1,420 lines

---

## What You Should Have Now

After completing Phase 7:

1. **Theming System**
   - Light/Dark mode with system preference support
   - 8 accent color options
   - Preferences persisted to localStorage

2. **Smooth Animations**
   - Panel slide-ins
   - Modal scale-ins
   - Hover transitions
   - Drag feedback
   - Respects reduced motion preference

3. **Responsive Layout**
   - Full mobile support with slide-out sidebar
   - Tablet layout with collapsible sidebar
   - Desktop layout with all panels

4. **User Feedback**
   - Skeleton loaders during initial load
   - Toast notifications for actions
   - Clear focus and selection states

5. **Performance**
   - Virtualization for long lists
   - Memoized components
   - Smooth 60fps interactions

---

## Next Phase

**Phase 8: Advanced Features** will add:
- Epics UI with visual grouping
- Due dates with calendar picker
- Sub-tasks with parent-child relationships
- Markdown editor for descriptions
- Activity log in task detail
- Import/Export functionality
- Task templates

---

## Troubleshooting

### Theme doesn't persist after refresh

**Problem**: Theme resets to dark mode on page refresh.

**Solution**: Check that localStorage is being set:
```javascript
// In browser console
localStorage.getItem('egenskriven-theme')
```
If undefined, verify the `setTheme` function calls `localStorage.setItem`.

### Animations feel janky

**Problem**: Animations are stuttering or not smooth.

**Solution**:
1. Check if animations are on properties that can be GPU-accelerated:
   - Use `transform` and `opacity` when possible
   - Avoid animating `width`, `height`, `top`, `left`
2. Add `will-change` for frequently animated elements:
   ```css
   .task-card {
     will-change: transform;
   }
   ```
3. Check for layout thrashing in DevTools Performance tab

### Mobile sidebar doesn't close on outside click

**Problem**: Clicking outside the sidebar doesn't close it.

**Solution**: Verify the backdrop element:
1. Has correct z-index (below sidebar, above content)
2. Has `pointer-events: auto` when visible
3. onClick handler calls `setSidebarOpen(false)`

### Virtualization breaks drag and drop

**Problem**: Can't drag items in virtualized columns.

**Solution**: 
1. Ensure SortableContext is inside the virtualized container
2. Items need stable keys (don't use index)
3. Make sure virtualized items have proper position styles

### Focus not visible on some elements

**Problem**: Tab navigation works but no visible focus ring.

**Solution**:
1. Check for `outline: none` without replacement
2. Verify `:focus-visible` styles are not overridden
3. Check contrast of focus ring color against background

### Performance still poor with virtualization

**Problem**: Even with virtualization, scrolling lags.

**Solution**:
1. Reduce `overscan` value in virtualizer config
2. Check for expensive renders in child components
3. Memoize TaskCard and other list items
4. Verify you're not creating new objects/arrays in render

---

## Glossary

| Term | Definition |
|------|------------|
| **CSS Custom Properties** | Variables defined with `--name` syntax, can be changed at runtime |
| **prefers-color-scheme** | Media query that detects OS light/dark mode preference |
| **prefers-reduced-motion** | Media query that detects if user wants reduced animation |
| **Virtualization** | Technique of only rendering visible items in a list |
| **Memoization** | Caching result of expensive computations |
| **DragOverlay** | @dnd-kit component that renders the dragged item clone |
| **focus-visible** | CSS pseudo-class that only shows focus for keyboard navigation |
| **Skeleton loader** | Placeholder that animates while content loads |
| **Toast** | Non-blocking notification that appears temporarily |
