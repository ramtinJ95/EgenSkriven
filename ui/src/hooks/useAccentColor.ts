import { useState, useEffect, useCallback, useRef } from 'react';
import { applyAccentColor } from '../themes/apply';
import { useTheme } from '../contexts/ThemeContext';

const STORAGE_KEY = 'egenskriven-accent';

interface UseAccentColorReturn {
  /** Current accent color (hex) - null means using theme default */
  accentColor: string;
  /** Whether a custom accent is set (not using theme default) */
  isCustomAccent: boolean;
  /** Update accent color */
  setAccentColor: (color: string) => void;
  /** Reset to theme's default accent color */
  resetToThemeDefault: () => void;
}

/**
 * Hook to manage accent color preference.
 *
 * The accent color can either be:
 * 1. The theme's default accent color (when no custom color is set)
 * 2. A user-selected custom accent color
 *
 * @example
 * const { accentColor, setAccentColor, resetToThemeDefault } = useAccentColor();
 * setAccentColor('#22C55E'); // Set to green
 * resetToThemeDefault(); // Reset to theme's accent
 */
export function useAccentColor(): UseAccentColorReturn {
  const { activeTheme } = useTheme();

  // Track the previous theme ID to detect theme changes
  const prevThemeIdRef = useRef<string>(activeTheme.id);

  // Get stored custom accent (or null if none)
  const [customAccent, setCustomAccent] = useState<string | null>(() => {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem(STORAGE_KEY);
  });

  // Effective accent is either custom or theme default
  const accentColor = customAccent || activeTheme.colors.accent;
  const isCustomAccent = customAccent !== null;

  // Reset custom accent when theme changes (per design decision in context.md)
  useEffect(() => {
    if (prevThemeIdRef.current !== activeTheme.id) {
      // Theme changed - reset to theme's default accent
      setCustomAccent(null);
      localStorage.removeItem(STORAGE_KEY);
      prevThemeIdRef.current = activeTheme.id;
    }
  }, [activeTheme.id]);

  // Apply custom accent color when it changes
  useEffect(() => {
    if (customAccent) {
      applyAccentColor(customAccent);
    }
    // Theme accent is already applied by ThemeContext via applyTheme()
  }, [customAccent]);

  const setAccentColor = useCallback((color: string) => {
    // Validate hex color format
    if (!/^#[0-9A-Fa-f]{6}$/.test(color)) {
      console.error('Invalid hex color:', color);
      return;
    }
    setCustomAccent(color);
    localStorage.setItem(STORAGE_KEY, color);
    applyAccentColor(color);
  }, []);

  const resetToThemeDefault = useCallback(() => {
    setCustomAccent(null);
    localStorage.removeItem(STORAGE_KEY);
    // Theme accent will be applied by the next theme render
  }, []);

  return { accentColor, isCustomAccent, setAccentColor, resetToThemeDefault };
}
