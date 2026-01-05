/**
 * Gruvbox Dark Theme
 *
 * Warm, retro groove colors with dark background.
 * Source: https://github.com/morhetz/gruvbox
 */

import type { Theme } from '../types';

export const gruvboxDark: Theme = {
  id: 'gruvbox-dark',
  name: 'Gruvbox Dark',
  appearance: 'dark',
  author: 'morhetz',
  source: 'https://github.com/morhetz/gruvbox',
  colors: {
    // Backgrounds - Gruvbox dark palette
    bgApp: '#1d2021', // bg0_h (hard)
    bgSidebar: '#282828', // bg0
    bgCard: '#3c3836', // bg1
    bgCardHover: '#504945', // bg2
    bgCardSelected: '#665c54', // bg3
    bgInput: '#32302f', // bg0_s (soft)
    bgOverlay: 'rgba(29, 32, 33, 0.8)',

    // Text - Gruvbox fg colors
    textPrimary: '#ebdbb2', // fg1
    textSecondary: '#d5c4a1', // fg2
    textMuted: '#a89984', // fg4
    textDisabled: '#665c54', // bg3

    // Borders
    borderSubtle: '#3c3836', // bg1
    borderDefault: '#504945', // bg2

    // Accent - Gruvbox orange
    accent: '#fe8019', // orange
    accentHover: '#fabd2f', // yellow
    accentMuted: 'rgba(254, 128, 25, 0.2)',

    // Shadows
    shadowSm: '0 1px 2px rgba(0, 0, 0, 0.3)',
    shadowMd: '0 4px 6px rgba(0, 0, 0, 0.4)',
    shadowLg: '0 10px 15px rgba(0, 0, 0, 0.5)',
    shadowDrag: '0 12px 24px rgba(0, 0, 0, 0.6)',

    // Status colors - Gruvbox palette
    statusBacklog: '#928374', // gray
    statusTodo: '#ebdbb2', // fg1
    statusInProgress: '#fabd2f', // yellow
    statusReview: '#d3869b', // purple
    statusDone: '#b8bb26', // green
    statusCanceled: '#928374', // gray

    // Priority colors - Gruvbox palette
    priorityUrgent: '#fb4934', // red
    priorityHigh: '#fe8019', // orange
    priorityMedium: '#fabd2f', // yellow
    priorityLow: '#928374', // gray
    priorityNone: '#665c54', // bg3

    // Type colors - Gruvbox palette
    typeBug: '#fb4934', // red
    typeFeature: '#d3869b', // purple
    typeChore: '#928374', // gray
  },
};
