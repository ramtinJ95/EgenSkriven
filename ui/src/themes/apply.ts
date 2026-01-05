/**
 * Theme Application Utility
 *
 * Applies theme colors to CSS custom properties on the document root.
 * Resets any previously set accent color override when switching themes.
 */

import type { Theme } from './types';

/**
 * Convert hex color to RGB components.
 *
 * @param hex - Hex color string (e.g., '#5E6AD2')
 * @returns RGB components or null if invalid
 */
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

/**
 * Apply a theme to the document.
 *
 * Sets all CSS custom properties from the theme colors.
 * Also sets data attributes for CSS fallbacks and debugging.
 *
 * @param theme - Theme to apply
 */
export function applyTheme(theme: Theme): void {
  const root = document.documentElement;
  const { colors } = theme;

  // Set data attributes for CSS fallbacks and debugging
  root.setAttribute('data-theme', theme.appearance);
  root.setAttribute('data-theme-id', theme.id);

  // Apply background colors
  root.style.setProperty('--bg-app', colors.bgApp);
  root.style.setProperty('--bg-sidebar', colors.bgSidebar);
  root.style.setProperty('--bg-card', colors.bgCard);
  root.style.setProperty('--bg-card-hover', colors.bgCardHover);
  root.style.setProperty('--bg-card-selected', colors.bgCardSelected);
  root.style.setProperty('--bg-input', colors.bgInput);
  root.style.setProperty('--bg-overlay', colors.bgOverlay);

  // Apply text colors
  root.style.setProperty('--text-primary', colors.textPrimary);
  root.style.setProperty('--text-secondary', colors.textSecondary);
  root.style.setProperty('--text-muted', colors.textMuted);
  root.style.setProperty('--text-disabled', colors.textDisabled);

  // Apply border colors
  root.style.setProperty('--border-subtle', colors.borderSubtle);
  root.style.setProperty('--border-default', colors.borderDefault);

  // Apply accent colors
  root.style.setProperty('--accent', colors.accent);
  root.style.setProperty('--accent-hover', colors.accentHover);
  root.style.setProperty('--accent-muted', colors.accentMuted);

  // Set accent RGB for transparency support
  const accentRgb = hexToRgb(colors.accent);
  if (accentRgb) {
    root.style.setProperty(
      '--accent-rgb',
      `${accentRgb.r}, ${accentRgb.g}, ${accentRgb.b}`
    );
  }

  // Apply shadows
  root.style.setProperty('--shadow-sm', colors.shadowSm);
  root.style.setProperty('--shadow-md', colors.shadowMd);
  root.style.setProperty('--shadow-lg', colors.shadowLg);
  root.style.setProperty('--shadow-drag', colors.shadowDrag);

  // Apply status colors
  root.style.setProperty('--status-backlog', colors.statusBacklog);
  root.style.setProperty('--status-todo', colors.statusTodo);
  root.style.setProperty('--status-in-progress', colors.statusInProgress);
  root.style.setProperty('--status-review', colors.statusReview);
  root.style.setProperty('--status-done', colors.statusDone);
  root.style.setProperty('--status-canceled', colors.statusCanceled);

  // Apply priority colors
  root.style.setProperty('--priority-urgent', colors.priorityUrgent);
  root.style.setProperty('--priority-high', colors.priorityHigh);
  root.style.setProperty('--priority-medium', colors.priorityMedium);
  root.style.setProperty('--priority-low', colors.priorityLow);
  root.style.setProperty('--priority-none', colors.priorityNone);

  // Apply type colors
  root.style.setProperty('--type-bug', colors.typeBug);
  root.style.setProperty('--type-feature', colors.typeFeature);
  root.style.setProperty('--type-chore', colors.typeChore);
}

/**
 * Apply a custom accent color override.
 *
 * Used when the user wants to override just the accent color
 * without changing the entire theme.
 *
 * @param color - Hex color for the accent
 */
export function applyAccentColor(color: string): void {
  const root = document.documentElement;

  root.style.setProperty('--accent', color);

  // Calculate hover color (slightly lighter)
  const rgb = hexToRgb(color);
  if (rgb) {
    const hoverR = Math.min(255, rgb.r + 16);
    const hoverG = Math.min(255, rgb.g + 16);
    const hoverB = Math.min(255, rgb.b + 16);
    const hoverHex = `#${hoverR.toString(16).padStart(2, '0')}${hoverG.toString(16).padStart(2, '0')}${hoverB.toString(16).padStart(2, '0')}`;
    root.style.setProperty('--accent-hover', hoverHex);

    // Set muted with alpha
    root.style.setProperty(
      '--accent-muted',
      `rgba(${rgb.r}, ${rgb.g}, ${rgb.b}, 0.2)`
    );

    // Set RGB for transparency support
    root.style.setProperty('--accent-rgb', `${rgb.r}, ${rgb.g}, ${rgb.b}`);
  }
}
