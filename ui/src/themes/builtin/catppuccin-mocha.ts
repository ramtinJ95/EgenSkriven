/**
 * Catppuccin Mocha Theme
 *
 * Soothing pastel theme - Mocha is the darkest variant.
 * Source: https://github.com/catppuccin/catppuccin
 */

import type { Theme } from '../types';

export const catppuccinMocha: Theme = {
  id: 'catppuccin-mocha',
  name: 'Catppuccin Mocha',
  appearance: 'dark',
  author: 'Catppuccin',
  source: 'https://github.com/catppuccin/catppuccin',
  colors: {
    // Backgrounds - Catppuccin Mocha palette
    bgApp: '#1e1e2e', // base
    bgSidebar: '#181825', // mantle
    bgCard: '#313244', // surface0
    bgCardHover: '#45475a', // surface1
    bgCardSelected: '#585b70', // surface2
    bgInput: '#181825', // mantle
    bgOverlay: 'rgba(17, 17, 27, 0.8)', // crust with alpha

    // Text - Catppuccin Mocha palette
    textPrimary: '#cdd6f4', // text
    textSecondary: '#bac2de', // subtext1
    textMuted: '#a6adc8', // subtext0
    textDisabled: '#6c7086', // overlay0

    // Borders
    borderSubtle: '#313244', // surface0
    borderDefault: '#45475a', // surface1

    // Accent - Catppuccin Mocha mauve (purple)
    accent: '#cba6f7', // mauve
    accentHover: '#f5c2e7', // pink
    accentMuted: 'rgba(203, 166, 247, 0.2)',

    // Shadows
    shadowSm: '0 1px 2px rgba(0, 0, 0, 0.3)',
    shadowMd: '0 4px 6px rgba(0, 0, 0, 0.4)',
    shadowLg: '0 10px 15px rgba(0, 0, 0, 0.5)',
    shadowDrag: '0 12px 24px rgba(0, 0, 0, 0.6)',

    // Status colors - Catppuccin Mocha palette
    statusBacklog: '#6c7086', // overlay0
    statusTodo: '#cdd6f4', // text
    statusInProgress: '#f9e2af', // yellow
    statusReview: '#cba6f7', // mauve
    statusDone: '#a6e3a1', // green
    statusCanceled: '#6c7086', // overlay0

    // Priority colors - Catppuccin Mocha palette
    priorityUrgent: '#f38ba8', // red
    priorityHigh: '#fab387', // peach
    priorityMedium: '#f9e2af', // yellow
    priorityLow: '#6c7086', // overlay0
    priorityNone: '#45475a', // surface1

    // Type colors - Catppuccin Mocha palette
    typeBug: '#f38ba8', // red
    typeFeature: '#cba6f7', // mauve
    typeChore: '#6c7086', // overlay0
  },
};
