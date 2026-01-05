/**
 * Theme Context
 *
 * Manages theme state and provides theme selection functionality.
 * Supports built-in themes, custom themes, and system mode with
 * configurable light/dark preferences.
 */

import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from 'react';
import {
  getAllThemes,
  getTheme,
  type Theme,
  type ThemeId,
} from '../themes';
import { applyTheme } from '../themes/apply';

/**
 * Theme mode can be 'system' or a specific theme ID.
 */
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

/**
 * Get the system's color scheme preference.
 */
function getSystemAppearance(): 'light' | 'dark' {
  if (typeof window === 'undefined') return 'dark';
  return window.matchMedia('(prefers-color-scheme: light)').matches
    ? 'light'
    : 'dark';
}

/**
 * Resolve which theme to use based on mode and preferences.
 */
function resolveTheme(
  mode: ThemeMode,
  preferredDark: ThemeId,
  preferredLight: ThemeId
): Theme {
  if (mode === 'system') {
    const appearance = getSystemAppearance();
    const themeId = appearance === 'dark' ? preferredDark : preferredLight;
    return getTheme(themeId) || getTheme('dark')!;
  }
  return getTheme(mode) || getTheme('dark')!;
}

interface ThemeProviderProps {
  children: ReactNode;
}

export function ThemeProvider({ children }: ThemeProviderProps) {
  // Note: Custom themes are loaded synchronously at module initialization
  // in themes/index.ts, so they're available before this component renders.

  // Initialize theme mode from localStorage
  const [themeMode, setThemeModeState] = useState<ThemeMode>(() => {
    if (typeof window === 'undefined') return 'system';
    const stored = localStorage.getItem(STORAGE_KEYS.mode);
    return (stored as ThemeMode) || 'system';
  });

  // Initialize preferred themes for system mode
  const [preferredDarkTheme, setPreferredDarkState] = useState<ThemeId>(() => {
    if (typeof window === 'undefined') return 'dark';
    return (localStorage.getItem(STORAGE_KEYS.darkPref) as ThemeId) || 'dark';
  });

  const [preferredLightTheme, setPreferredLightState] = useState<ThemeId>(
    () => {
      if (typeof window === 'undefined') return 'light';
      return (
        (localStorage.getItem(STORAGE_KEYS.lightPref) as ThemeId) || 'light'
      );
    }
  );

  // Track the actually applied theme
  const [activeTheme, setActiveTheme] = useState<Theme>(() => {
    return resolveTheme(themeMode, preferredDarkTheme, preferredLightTheme);
  });

  // Apply theme to DOM when it changes
  useEffect(() => {
    applyTheme(activeTheme);
  }, [activeTheme]);

  // Update active theme when mode or preferences change
  useEffect(() => {
    const newTheme = resolveTheme(
      themeMode,
      preferredDarkTheme,
      preferredLightTheme
    );
    setActiveTheme(newTheme);
  }, [themeMode, preferredDarkTheme, preferredLightTheme]);

  // Listen for system preference changes when in system mode
  useEffect(() => {
    if (themeMode !== 'system') return;

    const mediaQuery = window.matchMedia('(prefers-color-scheme: light)');

    const handleChange = () => {
      const newTheme = resolveTheme(
        'system',
        preferredDarkTheme,
        preferredLightTheme
      );
      setActiveTheme(newTheme);
    };

    mediaQuery.addEventListener('change', handleChange);
    return () => mediaQuery.removeEventListener('change', handleChange);
  }, [themeMode, preferredDarkTheme, preferredLightTheme]);

  // Setters with persistence
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

/**
 * Hook to access theme context.
 *
 * @example
 * const { themeMode, activeTheme, setThemeMode, availableThemes } = useTheme();
 *
 * // Switch to a specific theme
 * setThemeMode('gruvbox-dark');
 *
 * // Set to follow system
 * setThemeMode('system');
 *
 * // Get available themes
 * availableThemes.map(t => t.name);
 */
export function useTheme(): ThemeContextValue {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }
  return context;
}
