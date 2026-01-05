/**
 * Theme Registry
 *
 * Central registry for all themes (built-in and custom).
 * Provides functions to get, list, and manage themes.
 */

import type { Theme, BuiltinThemeId, ThemeId } from './types';
import { dark } from './builtin/dark';
import { light } from './builtin/light';
import { gruvboxDark } from './builtin/gruvbox-dark';
import { catppuccinMocha } from './builtin/catppuccin-mocha';
import { nord } from './builtin/nord';
import { tokyoNight } from './builtin/tokyo-night';

// Re-export types
export type { Theme, ThemeColors, BuiltinThemeId, ThemeId } from './types';
export type { ThemeInput, ValidationResult } from './schema';
export { validateTheme } from './schema';

/**
 * Built-in themes registry
 */
export const builtinThemes: Record<BuiltinThemeId, Theme> = {
  dark,
  light,
  'gruvbox-dark': gruvboxDark,
  'catppuccin-mocha': catppuccinMocha,
  nord,
  'tokyo-night': tokyoNight,
};

/**
 * Custom themes loaded from localStorage
 */
const customThemes: Map<string, Theme> = new Map();

/**
 * Storage key for custom themes
 */
const CUSTOM_THEMES_KEY = 'egenskriven-custom-themes';

/**
 * Get a theme by ID.
 *
 * @param id - Theme ID (builtin or custom)
 * @returns Theme if found, undefined otherwise
 */
export function getTheme(id: ThemeId): Theme | undefined {
  if (id.startsWith('custom-')) {
    return customThemes.get(id);
  }
  return builtinThemes[id as BuiltinThemeId];
}

/**
 * Get all available themes (built-in and custom).
 *
 * @returns Array of all themes
 */
export function getAllThemes(): Theme[] {
  return [
    ...Object.values(builtinThemes),
    ...Array.from(customThemes.values()),
  ];
}

/**
 * Get themes filtered by appearance.
 *
 * @param appearance - 'light' or 'dark'
 * @returns Array of themes matching the appearance
 */
export function getThemesByAppearance(appearance: 'light' | 'dark'): Theme[] {
  return getAllThemes().filter((t) => t.appearance === appearance);
}

/**
 * Register a custom theme.
 *
 * @param theme - Theme to register
 */
export function registerCustomTheme(theme: Theme): void {
  const id = theme.id.startsWith('custom-') ? theme.id : `custom-${theme.id}`;
  customThemes.set(id, { ...theme, id });
  saveCustomThemes();
}

/**
 * Remove a custom theme.
 *
 * @param id - Theme ID to remove
 */
export function removeCustomTheme(id: string): void {
  customThemes.delete(id);
  saveCustomThemes();
}

/**
 * Save custom themes to localStorage.
 */
function saveCustomThemes(): void {
  try {
    const themes = Array.from(customThemes.values());
    localStorage.setItem(CUSTOM_THEMES_KEY, JSON.stringify(themes));
  } catch (e) {
    console.error('Failed to save custom themes:', e);
  }
}

/**
 * Load custom themes from localStorage.
 * Should be called on app initialization.
 */
export function loadCustomThemes(): void {
  try {
    const stored = localStorage.getItem(CUSTOM_THEMES_KEY);
    if (stored) {
      const themes: Theme[] = JSON.parse(stored);
      themes.forEach((theme) => {
        customThemes.set(theme.id, theme);
      });
    }
  } catch (e) {
    console.error('Failed to load custom themes:', e);
  }
}

/**
 * Get all custom themes.
 *
 * @returns Array of custom themes
 */
export function getCustomThemes(): Theme[] {
  return Array.from(customThemes.values());
}

/**
 * Check if a theme is a built-in theme.
 *
 * @param id - Theme ID to check
 * @returns true if built-in, false if custom
 */
export function isBuiltinTheme(id: string): id is BuiltinThemeId {
  return id in builtinThemes;
}
