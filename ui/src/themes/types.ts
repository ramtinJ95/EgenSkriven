/**
 * Theme Type Definitions
 *
 * Defines the structure for themes in EgenSkriven.
 * Themes can be built-in or custom (user-created via JSON).
 */

/**
 * All themeable color properties.
 * These map directly to CSS custom properties (--variable-name).
 */
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
  statusNeedInput: string;
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

/**
 * Complete theme definition.
 */
export interface Theme {
  /** Unique identifier (e.g., 'gruvbox-dark') */
  id: string;
  /** Display name (e.g., 'Gruvbox Dark') */
  name: string;
  /** For system preference matching */
  appearance: 'light' | 'dark';
  /** Theme color palette */
  colors: ThemeColors;
  /** Optional theme author */
  author?: string;
  /** Optional source URL */
  source?: string;
}

/**
 * Built-in theme IDs
 */
export type BuiltinThemeId =
  | 'dark' // Built-in dark (current default)
  | 'light' // Built-in light
  | 'gruvbox-dark'
  | 'catppuccin-mocha'
  | 'nord'
  | 'tokyo-night'
  | 'purple-dream';

/**
 * Custom themes use string IDs prefixed with 'custom-'
 */
export type ThemeId = BuiltinThemeId | `custom-${string}`;
