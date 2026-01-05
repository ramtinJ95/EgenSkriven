/**
 * Tokyo Night Theme
 *
 * Purple-ish dark theme inspired by Tokyo city lights.
 * Source: https://github.com/enkia/tokyo-night-vscode-theme
 */

import type { Theme } from '../types';

export const tokyoNight: Theme = {
  id: 'tokyo-night',
  name: 'Tokyo Night',
  appearance: 'dark',
  author: 'enkia',
  source: 'https://github.com/enkia/tokyo-night-vscode-theme',
  colors: {
    // Backgrounds - Tokyo Night palette
    bgApp: '#1a1b26', // background
    bgSidebar: '#16161e', // background dark
    bgCard: '#1f2335', // background highlight
    bgCardHover: '#292e42', // terminal black
    bgCardSelected: '#33394b', // line highlight
    bgInput: '#1a1b26', // background
    bgOverlay: 'rgba(22, 22, 30, 0.8)',

    // Text - Tokyo Night palette
    textPrimary: '#c0caf5', // foreground
    textSecondary: '#9aa5ce', // foreground secondary
    textMuted: '#565f89', // comment
    textDisabled: '#414868', // terminal bright black

    // Borders
    borderSubtle: '#292e42', // terminal black
    borderDefault: '#33394b', // line highlight

    // Accent - Tokyo Night blue
    accent: '#7aa2f7', // blue
    accentHover: '#89b4fa', // lighter blue
    accentMuted: 'rgba(122, 162, 247, 0.2)',

    // Shadows
    shadowSm: '0 1px 2px rgba(0, 0, 0, 0.3)',
    shadowMd: '0 4px 6px rgba(0, 0, 0, 0.4)',
    shadowLg: '0 10px 15px rgba(0, 0, 0, 0.5)',
    shadowDrag: '0 12px 24px rgba(0, 0, 0, 0.6)',

    // Status colors - Tokyo Night palette
    statusBacklog: '#565f89', // comment
    statusTodo: '#c0caf5', // foreground
    statusInProgress: '#e0af68', // yellow
    statusReview: '#bb9af7', // purple
    statusDone: '#9ece6a', // green
    statusCanceled: '#565f89', // comment

    // Priority colors - Tokyo Night palette
    priorityUrgent: '#f7768e', // red
    priorityHigh: '#ff9e64', // orange
    priorityMedium: '#e0af68', // yellow
    priorityLow: '#565f89', // comment
    priorityNone: '#414868', // terminal bright black

    // Type colors - Tokyo Night palette
    typeBug: '#f7768e', // red
    typeFeature: '#bb9af7', // purple
    typeChore: '#565f89', // comment
  },
};
