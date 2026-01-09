/**
 * Default Light Theme
 *
 * Extracted from the original theme-light.css overrides.
 * Uses the same status/priority/type colors as dark theme for consistency.
 */

import type { Theme } from '../types';

export const light: Theme = {
  id: 'light',
  name: 'Light',
  appearance: 'light',
  colors: {
    // Backgrounds
    bgApp: '#FFFFFF',
    bgSidebar: '#FAFAFA',
    bgCard: '#FFFFFF',
    bgCardHover: '#F5F5F5',
    bgCardSelected: '#F0F0F0',
    bgInput: '#FFFFFF',
    bgOverlay: 'rgba(0, 0, 0, 0.4)',

    // Text
    textPrimary: '#171717',
    textSecondary: '#525252',
    textMuted: '#A3A3A3',
    textDisabled: '#D4D4D4',

    // Borders
    borderSubtle: '#E5E5E5',
    borderDefault: '#D4D4D4',

    // Accent (same as dark)
    accent: '#5E6AD2',
    accentHover: '#6E7BE2',
    accentMuted: 'rgba(94, 106, 210, 0.2)',

    // Shadows - lighter for light mode
    shadowSm: '0 1px 2px rgba(0, 0, 0, 0.05)',
    shadowMd: '0 4px 6px rgba(0, 0, 0, 0.07)',
    shadowLg: '0 10px 15px rgba(0, 0, 0, 0.1)',
    shadowDrag: '0 8px 16px rgba(0, 0, 0, 0.15)',

    // Status colors (same as dark)
    statusBacklog: '#6B7280',
    statusTodo: '#E5E5E5',
    statusInProgress: '#F59E0B',
    statusNeedInput: '#F97316',
    statusReview: '#A855F7',
    statusDone: '#22C55E',
    statusCanceled: '#6B7280',

    // Priority colors (same as dark)
    priorityUrgent: '#EF4444',
    priorityHigh: '#F97316',
    priorityMedium: '#EAB308',
    priorityLow: '#6B7280',
    priorityNone: '#444444',

    // Type colors (same as dark)
    typeBug: '#EF4444',
    typeFeature: '#A855F7',
    typeChore: '#6B7280',
  },
};
