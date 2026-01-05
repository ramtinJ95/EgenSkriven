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
