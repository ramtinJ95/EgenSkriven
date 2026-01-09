/**
 * Theme Validation Schema
 *
 * Uses Zod for runtime validation of custom themes.
 * Ensures user-provided theme JSON files are valid.
 */

import { z } from 'zod';

// Hex color validation (e.g., #RRGGBB)
const hexColor = z.string().regex(/^#[0-9A-Fa-f]{6}$/, {
  message: 'Must be a valid hex color (e.g., #FF5500)',
});

// CSS color value (for rgba() and other CSS values)
const cssColor = z.string().min(1, 'Color value is required');

// CSS shadow value
const cssShadow = z.string().min(1, 'Shadow value is required');

/**
 * Schema for theme colors.
 */
export const themeColorsSchema = z.object({
  // Backgrounds
  bgApp: hexColor,
  bgSidebar: hexColor,
  bgCard: hexColor,
  bgCardHover: hexColor,
  bgCardSelected: hexColor,
  bgInput: hexColor,
  bgOverlay: cssColor, // Can be rgba()

  // Text
  textPrimary: hexColor,
  textSecondary: hexColor,
  textMuted: hexColor,
  textDisabled: hexColor,

  // Borders
  borderSubtle: hexColor,
  borderDefault: hexColor,

  // Accent
  accent: hexColor,
  accentHover: hexColor,
  accentMuted: cssColor, // Can be rgba()

  // Shadows
  shadowSm: cssShadow,
  shadowMd: cssShadow,
  shadowLg: cssShadow,
  shadowDrag: cssShadow,

  // Status colors
  statusBacklog: hexColor,
  statusTodo: hexColor,
  statusInProgress: hexColor,
  statusNeedInput: hexColor,
  statusReview: hexColor,
  statusDone: hexColor,
  statusCanceled: hexColor,

  // Priority colors
  priorityUrgent: hexColor,
  priorityHigh: hexColor,
  priorityMedium: hexColor,
  priorityLow: hexColor,
  priorityNone: hexColor,

  // Type colors
  typeBug: hexColor,
  typeFeature: hexColor,
  typeChore: hexColor,
});

/**
 * Schema for a complete theme definition.
 */
export const themeSchema = z.object({
  name: z
    .string()
    .min(1, 'Theme name is required')
    .max(50, 'Theme name must be 50 characters or less'),
  appearance: z.enum(['light', 'dark'], {
    error: "Appearance must be 'light' or 'dark'",
  }),
  author: z.string().optional(),
  source: z.string().url('Source must be a valid URL').optional(),
  colors: themeColorsSchema,
});

/**
 * Type inferred from the theme schema (for custom theme input).
 */
export type ThemeInput = z.infer<typeof themeSchema>;

/**
 * Validation result type.
 */
export type ValidationResult =
  | { success: true; theme: ThemeInput }
  | { success: false; errors: string[] };

/**
 * Validates a theme object.
 *
 * @param input - The theme object to validate
 * @returns Validation result with either the parsed theme or error messages
 */
export function validateTheme(input: unknown): ValidationResult {
  const result = themeSchema.safeParse(input);

  if (result.success) {
    return { success: true, theme: result.data };
  }

  return {
    success: false,
    errors: result.error.issues.map(
      (e) => `${e.path.join('.')}: ${e.message}`
    ),
  };
}
