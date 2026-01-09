/**
 * Default Dark Theme
 *
 * Extracted from the original tokens.css dark mode defaults.
 */

import type { Theme } from '../types';

export const dark: Theme = {
  id: 'dark',
  name: 'Dark',
  appearance: 'dark',
  colors: {
    // Backgrounds
    bgApp: '#0D0D0D',
    bgSidebar: '#141414',
    bgCard: '#1A1A1A',
    bgCardHover: '#252525',
    bgCardSelected: '#2E2E2E',
    bgInput: '#1F1F1F',
    bgOverlay: 'rgba(0, 0, 0, 0.6)',

    // Text
    textPrimary: '#F5F5F5',
    textSecondary: '#A0A0A0',
    textMuted: '#666666',
    textDisabled: '#444444',

    // Borders
    borderSubtle: '#2A2A2A',
    borderDefault: '#333333',

    // Accent
    accent: '#5E6AD2',
    accentHover: '#6E7BE2',
    accentMuted: 'rgba(94, 106, 210, 0.2)',

    // Shadows
    shadowSm: '0 1px 2px rgba(0, 0, 0, 0.3)',
    shadowMd: '0 4px 6px rgba(0, 0, 0, 0.4)',
    shadowLg: '0 10px 15px rgba(0, 0, 0, 0.5)',
    shadowDrag: '0 12px 24px rgba(0, 0, 0, 0.6)',

    // Status colors
    statusBacklog: '#6B7280',
    statusTodo: '#E5E5E5',
    statusInProgress: '#F59E0B',
    statusNeedInput: '#F97316',
    statusReview: '#A855F7',
    statusDone: '#22C55E',
    statusCanceled: '#6B7280',

    // Priority colors
    priorityUrgent: '#EF4444',
    priorityHigh: '#F97316',
    priorityMedium: '#EAB308',
    priorityLow: '#6B7280',
    priorityNone: '#444444',

    // Type colors
    typeBug: '#EF4444',
    typeFeature: '#A855F7',
    typeChore: '#6B7280',
  },
};
