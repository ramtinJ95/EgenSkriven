# Custom Themes Implementation Plan

## Summary

Implement a custom theme system that:
1. Supports popular dark themes out of the box (Gruvbox, Catppuccin Mocha, Nord, Tokyo Night)
2. Allows users to easily create custom themes via JSON configuration files
3. Resets accent color to theme default when switching themes
4. Supports system mode with configurable light/dark theme preferences

---

## Current Architecture Analysis

### Existing Theme System
The application currently uses a CSS-first approach for theming:

1. **Dark mode as default** - `tokens.css` defines all CSS variables under `:root`
2. **Light mode via selector** - `theme-light.css` overrides variables using `[data-theme="light"]`
3. **React Context** - `ThemeContext.tsx` manages theme state (`system | light | dark`)
4. **DOM attribute** - Theme is applied via `data-theme` attribute on `<html>`
5. **Accent color hook** - `useAccentColor.ts` allows runtime accent color customization

### Key Files
- `ui/src/styles/tokens.css` - Design tokens (dark mode defaults)
- `ui/src/styles/theme-light.css` - Light mode overrides
- `ui/src/styles/index.css` - CSS entry point
- `ui/src/contexts/ThemeContext.tsx` - Theme state management
- `ui/src/hooks/useAccentColor.ts` - Accent color customization
- `ui/src/components/Settings.tsx` - Theme UI settings

### CSS Variables That Need Theming
| Category | Variables |
|----------|-----------|
| Backgrounds | `--bg-app`, `--bg-sidebar`, `--bg-card`, `--bg-card-hover`, `--bg-card-selected`, `--bg-input`, `--bg-overlay` |
| Text | `--text-primary`, `--text-secondary`, `--text-muted`, `--text-disabled` |
| Borders | `--border-subtle`, `--border-default` |
| Accent | `--accent`, `--accent-hover`, `--accent-muted` |
| Shadows | `--shadow-sm`, `--shadow-md`, `--shadow-lg`, `--shadow-drag` |
| Status | `--status-backlog`, `--status-todo`, `--status-in-progress`, `--status-review`, `--status-done`, `--status-canceled` |
| Priority | `--priority-urgent`, `--priority-high`, `--priority-medium`, `--priority-low`, `--priority-none` |
| Type | `--type-bug`, `--type-feature`, `--type-chore` |

---

## Research: How Other Tools Handle Themes

### Alacritty (Terminal Emulator)
- Uses **TOML** format for configuration
- Themes defined as color tables with semantic names
- Supports importing theme files: `import = ["~/.config/alacritty/themes/gruvbox_dark.toml"]`
- Community themes available at: https://github.com/alacritty/alacritty-theme
- Color structure:
  ```toml
  [colors.primary]
  foreground = "#d8d8d8"
  background = "#181818"
  
  [colors.normal]
  black = "#181818"
  red = "#ac4242"
  ...
  ```

### Kitty (Terminal Emulator)
- Uses simple **conf** format (key=value)
- Themes via `include` directive: `include themes/gruvbox_dark.conf`
- Built-in theme kitten: `kitty +kitten themes`
- Community themes: https://github.com/kovidgoyal/kitty-themes
- Simple color definitions:
  ```conf
  foreground #dddddd
  background #000000
  color0 #000000
  ...
  ```

### VS Code
- Uses **JSON** for color themes
- Themes defined in `package.json` and separate JSON files
- Semantic color tokens with workbench and syntax colors
- Example structure:
  ```json
  {
    "type": "dark",
    "colors": {
      "editor.background": "#1e1e1e",
      "editor.foreground": "#d4d4d4"
    }
  }
  ```

### Recommended Approach for EgenSkriven
Based on research, use **JSON** format for custom themes because:
1. Native to web/JavaScript ecosystem
2. Easy to parse and validate with TypeScript/Zod
3. Familiar to web developers
4. Can be easily exported/imported
5. Supports comments via JSONC if needed

---

## Implementation Plan

### Phase 1: Theme Type System

Create TypeScript interfaces for themes:

```typescript
// ui/src/themes/types.ts

export interface ThemeColors {
  // Backgrounds
  bgApp: string;
  bgSidebar: string;
  bgCard: string;
  bgCardHover: string;
  bgCardSelected: string;
  bgInput: string;
  bgOverlay: string;

  // Text
  textPrimary: string;
  textSecondary: string;
  textMuted: string;
  textDisabled: string;

  // Borders
  borderSubtle: string;
  borderDefault: string;

  // Accent
  accent: string;
  accentHover: string;
  accentMuted: string;

  // Shadows
  shadowSm: string;
  shadowMd: string;
  shadowLg: string;
  shadowDrag: string;

  // Status colors
  statusBacklog: string;
  statusTodo: string;
  statusInProgress: string;
  statusReview: string;
  statusDone: string;
  statusCanceled: string;

  // Priority colors
  priorityUrgent: string;
  priorityHigh: string;
  priorityMedium: string;
  priorityLow: string;
  priorityNone: string;

  // Type colors
  typeBug: string;
  typeFeature: string;
  typeChore: string;
}

export interface Theme {
  id: string;                      // Unique identifier (e.g., 'gruvbox-dark')
  name: string;                    // Display name (e.g., 'Gruvbox Dark')
  appearance: 'light' | 'dark';    // For system preference matching
  colors: ThemeColors;
  author?: string;                 // Optional theme author
  source?: string;                 // Optional source URL
}

// Built-in theme IDs
export type BuiltinThemeId = 
  | 'dark'              // Built-in dark (current default)
  | 'light'             // Built-in light
  | 'gruvbox-dark'
  | 'catppuccin-mocha'
  | 'nord'
  | 'tokyo-night';

// Custom themes use string IDs prefixed with 'custom-'
export type ThemeId = BuiltinThemeId | `custom-${string}`;
```

### Phase 2: Custom Theme JSON Schema

Define a JSON schema for custom themes that users can create:

```json
// Example: ~/.config/egenskriven/themes/my-theme.json
{
  "$schema": "https://egenskriven.app/schemas/theme.json",
  "name": "My Custom Theme",
  "appearance": "dark",
  "author": "Your Name",
  "colors": {
    "bgApp": "#1a1b26",
    "bgSidebar": "#16161e",
    "bgCard": "#1f2335",
    "bgCardHover": "#292e42",
    "bgCardSelected": "#33394b",
    "bgInput": "#1a1b26",
    "bgOverlay": "rgba(0, 0, 0, 0.6)",
    
    "textPrimary": "#c0caf5",
    "textSecondary": "#9aa5ce",
    "textMuted": "#565f89",
    "textDisabled": "#414868",
    
    "borderSubtle": "#292e42",
    "borderDefault": "#33394b",
    
    "accent": "#7aa2f7",
    "accentHover": "#89b4fa",
    "accentMuted": "rgba(122, 162, 247, 0.2)",
    
    "shadowSm": "0 1px 2px rgba(0, 0, 0, 0.3)",
    "shadowMd": "0 4px 6px rgba(0, 0, 0, 0.4)",
    "shadowLg": "0 10px 15px rgba(0, 0, 0, 0.5)",
    "shadowDrag": "0 12px 24px rgba(0, 0, 0, 0.6)",
    
    "statusBacklog": "#6B7280",
    "statusTodo": "#c0caf5",
    "statusInProgress": "#e0af68",
    "statusReview": "#bb9af7",
    "statusDone": "#9ece6a",
    "statusCanceled": "#6B7280",
    
    "priorityUrgent": "#f7768e",
    "priorityHigh": "#ff9e64",
    "priorityMedium": "#e0af68",
    "priorityLow": "#6B7280",
    "priorityNone": "#414868",
    
    "typeBug": "#f7768e",
    "typeFeature": "#bb9af7",
    "typeChore": "#6B7280"
  }
}
```

### Phase 3: Theme Registry and Loader

```typescript
// ui/src/themes/index.ts

import type { Theme, ThemeId, BuiltinThemeId } from './types';
import { dark } from './builtin/dark';
import { light } from './builtin/light';
import { gruvboxDark } from './builtin/gruvbox-dark';
import { catppuccinMocha } from './builtin/catppuccin-mocha';
import { nord } from './builtin/nord';
import { tokyoNight } from './builtin/tokyo-night';

// Built-in themes registry
export const builtinThemes: Record<BuiltinThemeId, Theme> = {
  'dark': dark,
  'light': light,
  'gruvbox-dark': gruvboxDark,
  'catppuccin-mocha': catppuccinMocha,
  'nord': nord,
  'tokyo-night': tokyoNight,
};

// Custom themes loaded from localStorage/files
let customThemes: Map<string, Theme> = new Map();

export function getTheme(id: ThemeId): Theme | undefined {
  if (id.startsWith('custom-')) {
    return customThemes.get(id);
  }
  return builtinThemes[id as BuiltinThemeId];
}

export function getAllThemes(): Theme[] {
  return [
    ...Object.values(builtinThemes),
    ...Array.from(customThemes.values()),
  ];
}

export function getThemesByAppearance(appearance: 'light' | 'dark'): Theme[] {
  return getAllThemes().filter(t => t.appearance === appearance);
}

export function registerCustomTheme(theme: Theme): void {
  const id = `custom-${theme.id}`;
  customThemes.set(id, { ...theme, id });
  // Persist to localStorage
  saveCustomThemes();
}

export function removeCustomTheme(id: string): void {
  customThemes.delete(id);
  saveCustomThemes();
}

// Persist/load custom themes from localStorage
const CUSTOM_THEMES_KEY = 'egenskriven-custom-themes';

function saveCustomThemes(): void {
  const themes = Array.from(customThemes.values());
  localStorage.setItem(CUSTOM_THEMES_KEY, JSON.stringify(themes));
}

export function loadCustomThemes(): void {
  try {
    const stored = localStorage.getItem(CUSTOM_THEMES_KEY);
    if (stored) {
      const themes: Theme[] = JSON.parse(stored);
      themes.forEach(theme => {
        customThemes.set(theme.id, theme);
      });
    }
  } catch (e) {
    console.error('Failed to load custom themes:', e);
  }
}
```

### Phase 4: Theme Application Function

```typescript
// ui/src/themes/apply.ts

import type { Theme } from './types';

/**
 * Apply theme colors to CSS custom properties.
 * Resets any previously set accent color override.
 */
export function applyTheme(theme: Theme): void {
  const root = document.documentElement;
  const { colors } = theme;

  // Set data attributes for CSS fallbacks
  root.setAttribute('data-theme', theme.appearance);
  root.setAttribute('data-theme-id', theme.id);

  // Apply all color variables
  root.style.setProperty('--bg-app', colors.bgApp);
  root.style.setProperty('--bg-sidebar', colors.bgSidebar);
  root.style.setProperty('--bg-card', colors.bgCard);
  root.style.setProperty('--bg-card-hover', colors.bgCardHover);
  root.style.setProperty('--bg-card-selected', colors.bgCardSelected);
  root.style.setProperty('--bg-input', colors.bgInput);
  root.style.setProperty('--bg-overlay', colors.bgOverlay);

  root.style.setProperty('--text-primary', colors.textPrimary);
  root.style.setProperty('--text-secondary', colors.textSecondary);
  root.style.setProperty('--text-muted', colors.textMuted);
  root.style.setProperty('--text-disabled', colors.textDisabled);

  root.style.setProperty('--border-subtle', colors.borderSubtle);
  root.style.setProperty('--border-default', colors.borderDefault);

  root.style.setProperty('--accent', colors.accent);
  root.style.setProperty('--accent-hover', colors.accentHover);
  root.style.setProperty('--accent-muted', colors.accentMuted);

  root.style.setProperty('--shadow-sm', colors.shadowSm);
  root.style.setProperty('--shadow-md', colors.shadowMd);
  root.style.setProperty('--shadow-lg', colors.shadowLg);
  root.style.setProperty('--shadow-drag', colors.shadowDrag);

  root.style.setProperty('--status-backlog', colors.statusBacklog);
  root.style.setProperty('--status-todo', colors.statusTodo);
  root.style.setProperty('--status-in-progress', colors.statusInProgress);
  root.style.setProperty('--status-review', colors.statusReview);
  root.style.setProperty('--status-done', colors.statusDone);
  root.style.setProperty('--status-canceled', colors.statusCanceled);

  root.style.setProperty('--priority-urgent', colors.priorityUrgent);
  root.style.setProperty('--priority-high', colors.priorityHigh);
  root.style.setProperty('--priority-medium', colors.priorityMedium);
  root.style.setProperty('--priority-low', colors.priorityLow);
  root.style.setProperty('--priority-none', colors.priorityNone);

  root.style.setProperty('--type-bug', colors.typeBug);
  root.style.setProperty('--type-feature', colors.typeFeature);
  root.style.setProperty('--type-chore', colors.typeChore);

  // Set accent RGB for transparency support
  const accentRgb = hexToRgb(colors.accent);
  if (accentRgb) {
    root.style.setProperty('--accent-rgb', `${accentRgb.r}, ${accentRgb.g}, ${accentRgb.b}`);
  }
}

function hexToRgb(hex: string): { r: number; g: number; b: number } | null {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  return result
    ? {
        r: parseInt(result[1], 16),
        g: parseInt(result[2], 16),
        b: parseInt(result[3], 16),
      }
    : null;
}
```

### Phase 5: Updated ThemeContext

```typescript
// ui/src/contexts/ThemeContext.tsx (refactored)

import { createContext, useContext, useEffect, useState, type ReactNode } from 'react';
import { 
  getAllThemes, 
  getTheme, 
  getThemesByAppearance, 
  loadCustomThemes,
  type Theme, 
  type ThemeId 
} from '../themes';
import { applyTheme } from '../themes/apply';

type ThemeMode = 'system' | ThemeId;

interface ThemeContextValue {
  /** Current theme mode setting (may be 'system') */
  themeMode: ThemeMode;
  /** Actually applied theme object */
  activeTheme: Theme;
  /** Update theme mode */
  setThemeMode: (mode: ThemeMode) => void;
  /** All available themes */
  availableThemes: Theme[];
  /** Preferred theme for system dark mode */
  preferredDarkTheme: ThemeId;
  /** Preferred theme for system light mode */
  preferredLightTheme: ThemeId;
  /** Set preferred dark theme for system mode */
  setPreferredDarkTheme: (id: ThemeId) => void;
  /** Set preferred light theme for system mode */
  setPreferredLightTheme: (id: ThemeId) => void;
}

const ThemeContext = createContext<ThemeContextValue | undefined>(undefined);

const STORAGE_KEYS = {
  mode: 'egenskriven-theme-mode',
  darkPref: 'egenskriven-theme-dark',
  lightPref: 'egenskriven-theme-light',
};

function getSystemAppearance(): 'light' | 'dark' {
  if (typeof window === 'undefined') return 'dark';
  return window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark';
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  // Load custom themes on mount
  useEffect(() => {
    loadCustomThemes();
  }, []);

  // Initialize theme mode from localStorage
  const [themeMode, setThemeModeState] = useState<ThemeMode>(() => {
    if (typeof window === 'undefined') return 'system';
    return (localStorage.getItem(STORAGE_KEYS.mode) as ThemeMode) || 'system';
  });

  // Initialize preferred themes for system mode
  const [preferredDarkTheme, setPreferredDarkState] = useState<ThemeId>(() => {
    if (typeof window === 'undefined') return 'dark';
    return (localStorage.getItem(STORAGE_KEYS.darkPref) as ThemeId) || 'dark';
  });

  const [preferredLightTheme, setPreferredLightState] = useState<ThemeId>(() => {
    if (typeof window === 'undefined') return 'light';
    return (localStorage.getItem(STORAGE_KEYS.lightPref) as ThemeId) || 'light';
  });

  // Resolve which theme to actually apply
  const [activeTheme, setActiveTheme] = useState<Theme>(() => {
    if (themeMode === 'system') {
      const appearance = getSystemAppearance();
      const themeId = appearance === 'dark' ? preferredDarkTheme : preferredLightTheme;
      return getTheme(themeId) || getTheme('dark')!;
    }
    return getTheme(themeMode) || getTheme('dark')!;
  });

  // Apply theme when it changes
  useEffect(() => {
    applyTheme(activeTheme);
  }, [activeTheme]);

  // Listen for system preference changes when in system mode
  useEffect(() => {
    if (themeMode !== 'system') return;

    const mediaQuery = window.matchMedia('(prefers-color-scheme: light)');
    
    const handleChange = (e: MediaQueryListEvent) => {
      const themeId = e.matches ? preferredLightTheme : preferredDarkTheme;
      const theme = getTheme(themeId) || getTheme('dark')!;
      setActiveTheme(theme);
    };

    mediaQuery.addEventListener('change', handleChange);
    return () => mediaQuery.removeEventListener('change', handleChange);
  }, [themeMode, preferredDarkTheme, preferredLightTheme]);

  // Update active theme when mode or preferences change
  useEffect(() => {
    let themeId: ThemeId;
    if (themeMode === 'system') {
      const appearance = getSystemAppearance();
      themeId = appearance === 'dark' ? preferredDarkTheme : preferredLightTheme;
    } else {
      themeId = themeMode;
    }
    const theme = getTheme(themeId) || getTheme('dark')!;
    setActiveTheme(theme);
  }, [themeMode, preferredDarkTheme, preferredLightTheme]);

  const setThemeMode = (mode: ThemeMode) => {
    setThemeModeState(mode);
    localStorage.setItem(STORAGE_KEYS.mode, mode);
  };

  const setPreferredDarkTheme = (id: ThemeId) => {
    setPreferredDarkState(id);
    localStorage.setItem(STORAGE_KEYS.darkPref, id);
  };

  const setPreferredLightTheme = (id: ThemeId) => {
    setPreferredLightState(id);
    localStorage.setItem(STORAGE_KEYS.lightPref, id);
  };

  return (
    <ThemeContext.Provider
      value={{
        themeMode,
        activeTheme,
        setThemeMode,
        availableThemes: getAllThemes(),
        preferredDarkTheme,
        preferredLightTheme,
        setPreferredDarkTheme,
        setPreferredLightTheme,
      }}
    >
      {children}
    </ThemeContext.Provider>
  );
}

export function useTheme(): ThemeContextValue {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }
  return context;
}
```

### Phase 6: Theme Validation with Zod

```typescript
// ui/src/themes/schema.ts

import { z } from 'zod';

const hexColor = z.string().regex(/^#[0-9A-Fa-f]{6}$/);
const cssColor = z.string(); // For rgba() values

export const themeColorsSchema = z.object({
  bgApp: hexColor,
  bgSidebar: hexColor,
  bgCard: hexColor,
  bgCardHover: hexColor,
  bgCardSelected: hexColor,
  bgInput: hexColor,
  bgOverlay: cssColor,

  textPrimary: hexColor,
  textSecondary: hexColor,
  textMuted: hexColor,
  textDisabled: hexColor,

  borderSubtle: hexColor,
  borderDefault: hexColor,

  accent: hexColor,
  accentHover: hexColor,
  accentMuted: cssColor,

  shadowSm: z.string(),
  shadowMd: z.string(),
  shadowLg: z.string(),
  shadowDrag: z.string(),

  statusBacklog: hexColor,
  statusTodo: hexColor,
  statusInProgress: hexColor,
  statusReview: hexColor,
  statusDone: hexColor,
  statusCanceled: hexColor,

  priorityUrgent: hexColor,
  priorityHigh: hexColor,
  priorityMedium: hexColor,
  priorityLow: hexColor,
  priorityNone: hexColor,

  typeBug: hexColor,
  typeFeature: hexColor,
  typeChore: hexColor,
});

export const themeSchema = z.object({
  name: z.string().min(1).max(50),
  appearance: z.enum(['light', 'dark']),
  author: z.string().optional(),
  source: z.string().url().optional(),
  colors: themeColorsSchema,
});

export type ThemeInput = z.infer<typeof themeSchema>;

export function validateTheme(input: unknown): { success: true; theme: ThemeInput } | { success: false; errors: string[] } {
  const result = themeSchema.safeParse(input);
  if (result.success) {
    return { success: true, theme: result.data };
  }
  return {
    success: false,
    errors: result.error.errors.map(e => `${e.path.join('.')}: ${e.message}`),
  };
}
```

---

## Built-in Themes to Implement (Dark Only)

### 1. Default Dark (Current)
Extract current dark theme colors from `tokens.css`.

### 2. Default Light (Current)
Extract current light theme colors from `theme-light.css`.

### 3. Gruvbox Dark
Warm, retro groove colors with dark background.
- Source: https://github.com/morhetz/gruvbox

### 4. Catppuccin Mocha
Soothing pastel theme with dark background (darkest Catppuccin variant).
- Source: https://github.com/catppuccin/catppuccin

### 5. Nord
Cool, bluish arctic colors with dark background.
- Source: https://www.nordtheme.com

### 6. Tokyo Night
Purple-ish dark theme inspired by Tokyo city lights.
- Source: https://github.com/enkia/tokyo-night-vscode-theme

---

## Implementation Steps (Task Order)

### Core Infrastructure
1. Create theme type definitions (`ui/src/themes/types.ts`)
2. Create theme validation schema (`ui/src/themes/schema.ts`)
3. Extract current dark theme to `ui/src/themes/builtin/dark.ts`
4. Extract current light theme to `ui/src/themes/builtin/light.ts`
5. Create theme registry (`ui/src/themes/index.ts`)
6. Create theme application utility (`ui/src/themes/apply.ts`)
7. Refactor ThemeContext to use new theme system

### Built-in Themes
8. Implement Gruvbox Dark theme
9. Implement Catppuccin Mocha theme
10. Implement Nord theme
11. Implement Tokyo Night theme

### UI Updates
12. Update Settings UI with theme selection grid
13. Add theme preview swatches
14. Add system mode preferences (preferred dark/light themes)
15. Add custom theme import functionality

### Cleanup
16. Remove color variables from `tokens.css` (keep non-color tokens)
17. Remove `theme-light.css` (replaced by JS theme system)
18. Test all themes across all components

---

## Key Decisions

### Accent Color Behavior
- **Decision**: Reset accent color to theme's default when switching themes
- **Rationale**: Each theme has carefully chosen accent colors that work with the palette. User can still override after switching.

### System Mode Behavior
- **Decision**: Allow users to configure which specific theme is used for system dark/light modes
- **Example**: User can set "Catppuccin Mocha" for dark mode and "Light" for light mode
- **Storage**:
  - `egenskriven-theme-mode` - 'system' | ThemeId
  - `egenskriven-theme-dark` - ThemeId for system dark preference
  - `egenskriven-theme-light` - ThemeId for system light preference

### Custom Theme Format
- **Decision**: Use JSON format for custom themes
- **Rationale**: Native to web ecosystem, easy to validate, familiar format

---

## File Changes Summary

### New Files
```
ui/src/themes/
  types.ts              # Type definitions
  schema.ts             # Zod validation schema
  index.ts              # Theme registry and exports
  apply.ts              # Theme application utility
  builtin/
    dark.ts             # Current dark theme
    light.ts            # Current light theme
    gruvbox-dark.ts     # Gruvbox Dark
    catppuccin-mocha.ts # Catppuccin Mocha
    nord.ts             # Nord
    tokyo-night.ts      # Tokyo Night
```

### Modified Files
- `ui/src/contexts/ThemeContext.tsx` - Major refactor to new system
- `ui/src/contexts/index.ts` - Update exports
- `ui/src/components/Settings.tsx` - Add theme selection UI
- `ui/src/styles/tokens.css` - Remove color variables (keep spacing, typography, etc.)
- `ui/src/styles/index.css` - Remove theme-light.css import
- `ui/src/hooks/useAccentColor.ts` - Integrate with theme system

### Removed Files
- `ui/src/styles/theme-light.css` - Replaced by JS theme system

---

## Settings UI Design

```
Appearance
-----------
Theme: [Dropdown: System / Dark / Light / Gruvbox Dark / Catppuccin Mocha / Nord / Tokyo Night]

[If System selected:]
  Dark mode theme:  [Dropdown of dark themes]
  Light mode theme: [Dropdown of light themes]

Theme Preview:
  [Grid of 4-5 color swatches showing bg, text, accent, status colors]

Custom Themes
-------------
[Import Theme] button -> Opens file picker for JSON
[List of imported custom themes with delete option]
```

---

## Estimated Effort

| Phase | Effort |
|-------|--------|
| Type system + schema | 1-2 hours |
| Theme registry + loader | 1-2 hours |
| Theme application utility | 1 hour |
| ThemeContext refactor | 2-3 hours |
| Settings UI update | 3-4 hours |
| Each built-in theme | 30-45 min each (x4 = 2-3 hours) |
| Custom theme import UI | 1-2 hours |
| Testing + polish | 2-3 hours |
| **Total** | ~14-20 hours |
