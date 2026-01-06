/**
 * Purple Dream Theme
 *
 * A dreamy vaporwave aesthetic theme featuring pastel pinks, purples, and cyans
 * with a retro-futuristic atmosphere. Based on the Omarchy theme.
 * Source: https://github.com/ramtinJ95/purple-dream
 */

import type { Theme } from '../types';

export const purpleDream: Theme = {
  id: 'purple-dream',
  name: 'Purple Dream',
  appearance: 'dark',
  author: 'Ramtin Javanmardi',
  source: 'https://github.com/ramtinJ95/purple-dream',
  colors: {
    // Backgrounds - Purple Dream palette
    bgApp: '#1a0d2e', // deep dark purple
    bgSidebar: '#1a0d2e', // main background
    bgCard: '#2d1b4e', // elevated surfaces
    bgCardHover: '#3d2a5e', // slightly lighter for hover
    bgCardSelected: '#543a6e', // muted purple for selection
    bgInput: '#2d1b4e', // elevated background
    bgOverlay: 'rgba(26, 13, 46, 0.85)',

    // Text - Purple Dream palette
    textPrimary: '#d4a5ff', // lavender foreground
    textSecondary: '#a37acc', // dim purple text
    textMuted: '#8b72a3', // comments
    textDisabled: '#543a6e', // muted purple

    // Borders
    borderSubtle: '#543a6e', // muted purple
    borderDefault: '#ff6ec7', // hot pink accent

    // Accent - Hot pink
    accent: '#ff6ec7', // hot pink
    accentHover: '#ff9adc', // soft pink
    accentMuted: 'rgba(255, 110, 199, 0.2)',

    // Shadows
    shadowSm: '0 1px 2px rgba(26, 13, 46, 0.4)',
    shadowMd: '0 4px 6px rgba(26, 13, 46, 0.5)',
    shadowLg: '0 10px 15px rgba(26, 13, 46, 0.6)',
    shadowDrag: '0 12px 24px rgba(26, 13, 46, 0.7)',

    // Status colors - Purple Dream palette
    statusBacklog: '#8b72a3', // comments purple
    statusTodo: '#d4a5ff', // lavender
    statusInProgress: '#f9f871', // neon yellow
    statusReview: '#8b9aff', // periwinkle blue
    statusDone: '#5ffbf1', // bright cyan
    statusCanceled: '#543a6e', // muted purple

    // Priority colors - Purple Dream palette
    priorityUrgent: '#ff6ec7', // hot pink
    priorityHigh: '#ffb3d9', // pastel pink
    priorityMedium: '#f9f871', // neon yellow
    priorityLow: '#8b72a3', // muted purple
    priorityNone: '#543a6e', // dark muted purple

    // Type colors - Purple Dream palette
    typeBug: '#ff6ec7', // hot pink
    typeFeature: '#f4a5ff', // soft magenta
    typeChore: '#8b72a3', // muted purple
  },
};
