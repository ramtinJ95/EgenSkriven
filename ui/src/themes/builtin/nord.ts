/**
 * Nord Theme
 *
 * Cool, bluish arctic colors with dark background.
 * Source: https://www.nordtheme.com
 */

import type { Theme } from '../types';

export const nord: Theme = {
  id: 'nord',
  name: 'Nord',
  appearance: 'dark',
  author: 'Arctic Ice Studio',
  source: 'https://www.nordtheme.com',
  colors: {
    // Backgrounds - Nord Polar Night palette
    bgApp: '#2e3440', // nord0
    bgSidebar: '#3b4252', // nord1
    bgCard: '#434c5e', // nord2
    bgCardHover: '#4c566a', // nord3
    bgCardSelected: '#5e6779', // lighter nord3
    bgInput: '#3b4252', // nord1
    bgOverlay: 'rgba(46, 52, 64, 0.8)', // nord0 with alpha

    // Text - Nord Snow Storm palette
    textPrimary: '#eceff4', // nord6
    textSecondary: '#e5e9f0', // nord5
    textMuted: '#d8dee9', // nord4
    textDisabled: '#4c566a', // nord3

    // Borders
    borderSubtle: '#434c5e', // nord2
    borderDefault: '#4c566a', // nord3

    // Accent - Nord Frost (blue)
    accent: '#88c0d0', // nord8
    accentHover: '#8fbcbb', // nord7
    accentMuted: 'rgba(136, 192, 208, 0.2)',

    // Shadows
    shadowSm: '0 1px 2px rgba(0, 0, 0, 0.3)',
    shadowMd: '0 4px 6px rgba(0, 0, 0, 0.4)',
    shadowLg: '0 10px 15px rgba(0, 0, 0, 0.5)',
    shadowDrag: '0 12px 24px rgba(0, 0, 0, 0.6)',

    // Status colors - Nord Aurora palette
    statusBacklog: '#4c566a', // nord3
    statusTodo: '#eceff4', // nord6
    statusInProgress: '#ebcb8b', // nord13 (yellow)
    statusReview: '#b48ead', // nord15 (purple)
    statusDone: '#a3be8c', // nord14 (green)
    statusCanceled: '#4c566a', // nord3

    // Priority colors - Nord Aurora palette
    priorityUrgent: '#bf616a', // nord11 (red)
    priorityHigh: '#d08770', // nord12 (orange)
    priorityMedium: '#ebcb8b', // nord13 (yellow)
    priorityLow: '#4c566a', // nord3
    priorityNone: '#434c5e', // nord2

    // Type colors - Nord Aurora palette
    typeBug: '#bf616a', // nord11 (red)
    typeFeature: '#b48ead', // nord15 (purple)
    typeChore: '#4c566a', // nord3
  },
};
