import { useState, useEffect } from 'react';

const STORAGE_KEY = 'egenskriven-accent';
const DEFAULT_ACCENT = '#5E6AD2'; // Blue

/**
 * Apply accent color to CSS custom properties.
 * This updates the --accent variable used throughout the app.
 */
function applyAccentColor(color: string): void {
  document.documentElement.style.setProperty('--accent', color);
  
  // Also set RGB values for transparency support
  // Convert hex to RGB
  const r = parseInt(color.slice(1, 3), 16);
  const g = parseInt(color.slice(3, 5), 16);
  const b = parseInt(color.slice(5, 7), 16);
  document.documentElement.style.setProperty('--accent-rgb', `${r}, ${g}, ${b}`);
}

interface UseAccentColorReturn {
  /** Current accent color (hex) */
  accentColor: string;
  /** Update accent color */
  setAccentColor: (color: string) => void;
}

/**
 * Hook to manage accent color preference.
 * 
 * @example
 * const { accentColor, setAccentColor } = useAccentColor();
 * setAccentColor('#22C55E'); // Set to green
 */
export function useAccentColor(): UseAccentColorReturn {
  const [accentColor, setAccentState] = useState<string>(() => {
    if (typeof window === 'undefined') return DEFAULT_ACCENT;
    return localStorage.getItem(STORAGE_KEY) || DEFAULT_ACCENT;
  });

  // Apply accent color on mount and when it changes
  useEffect(() => {
    applyAccentColor(accentColor);
  }, [accentColor]);

  const setAccentColor = (color: string) => {
    // Validate hex color format
    if (!/^#[0-9A-Fa-f]{6}$/.test(color)) {
      console.error('Invalid hex color:', color);
      return;
    }
    setAccentState(color);
    localStorage.setItem(STORAGE_KEY, color);
  };

  return { accentColor, setAccentColor };
}
