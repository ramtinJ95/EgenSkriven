import { useEffect, useRef } from 'react';
import { useTheme } from '../contexts/ThemeContext';
import { useAccentColor } from '../hooks/useAccentColor';
import { getThemesByAppearance, type ThemeId } from '../themes';
import styles from './Settings.module.css';

interface SettingsProps {
  /** Whether the settings panel is open */
  isOpen: boolean;
  /** Callback when settings should close */
  onClose: () => void;
}

/**
 * Available accent color options.
 * Each has a name and hex value.
 */
const ACCENT_COLORS = [
  { name: 'Blue', value: '#5E6AD2' },
  { name: 'Purple', value: '#9333EA' },
  { name: 'Green', value: '#22C55E' },
  { name: 'Orange', value: '#F97316' },
  { name: 'Pink', value: '#EC4899' },
  { name: 'Cyan', value: '#06B6D4' },
  { name: 'Red', value: '#EF4444' },
  { name: 'Yellow', value: '#EAB308' },
] as const;

export function Settings({ isOpen, onClose }: SettingsProps) {
  const {
    themeMode,
    setThemeMode,
    availableThemes,
    preferredDarkTheme,
    preferredLightTheme,
    setPreferredDarkTheme,
    setPreferredLightTheme,
  } = useTheme();
  const { accentColor, setAccentColor } = useAccentColor();
  const panelRef = useRef<HTMLDivElement>(null);
  const justOpenedRef = useRef(false);

  // Track when panel just opened to prevent immediate close
  useEffect(() => {
    if (isOpen) {
      justOpenedRef.current = true;
    }
  }, [isOpen]);

  // Note: Escape key is handled centrally in App.tsx to avoid duplicate handlers

  // Close when clicking outside (using mousedown for more reliable detection)
  useEffect(() => {
    if (!isOpen) return;

    const handleClickOutside = (e: MouseEvent) => {
      // Skip if panel just opened (prevents the opening click from closing it)
      if (justOpenedRef.current) {
        justOpenedRef.current = false;
        return;
      }
      if (panelRef.current && !panelRef.current.contains(e.target as Node)) {
        onClose();
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  // Get themes by appearance for system mode preferences
  const darkThemes = getThemesByAppearance('dark');
  const lightThemes = getThemesByAppearance('light');

  return (
    <div className={styles.overlay}>
      <div className={styles.panel} ref={panelRef}>
        <header className={styles.header}>
          <h2>Settings</h2>
          <button
            className={styles.closeButton}
            onClick={onClose}
            aria-label="Close settings"
          >
            &times;
          </button>
        </header>

        <div className={styles.content}>
          {/* Appearance Section */}
          <section className={styles.section}>
            <h3>Appearance</h3>

            {/* Theme Selection Grid */}
            <div className={styles.row}>
              <label>Theme</label>
              <div className={styles.themeGrid}>
                {/* System option */}
                <button
                  className={`${styles.themeOption} ${
                    themeMode === 'system' ? styles.selected : ''
                  }`}
                  onClick={() => setThemeMode('system')}
                  title="Follow system preference"
                >
                  <div className={styles.themePreview}>
                    <div
                      className={styles.themeSwatch}
                      style={{
                        background:
                          'linear-gradient(135deg, #1a1a1a 50%, #ffffff 50%)',
                      }}
                    />
                  </div>
                  <span className={styles.themeName}>System</span>
                </button>

                {/* All available themes */}
                {availableThemes.map((theme) => (
                  <button
                    key={theme.id}
                    className={`${styles.themeOption} ${
                      themeMode === theme.id ? styles.selected : ''
                    }`}
                    onClick={() => setThemeMode(theme.id as ThemeId)}
                    title={theme.name}
                  >
                    <div className={styles.themePreview}>
                      <div
                        className={styles.themeSwatch}
                        style={{ backgroundColor: theme.colors.bgApp }}
                      >
                        <div
                          className={styles.themeAccentDot}
                          style={{ backgroundColor: theme.colors.accent }}
                        />
                      </div>
                    </div>
                    <span className={styles.themeName}>{theme.name}</span>
                  </button>
                ))}
              </div>
            </div>

            {/* System Mode Preferences - only show when system mode is selected */}
            {themeMode === 'system' && (
              <>
                <div className={styles.row}>
                  <label htmlFor="dark-theme-select">Dark mode theme</label>
                  <select
                    id="dark-theme-select"
                    value={preferredDarkTheme}
                    onChange={(e) =>
                      setPreferredDarkTheme(e.target.value as ThemeId)
                    }
                    className={styles.select}
                  >
                    {darkThemes.map((theme) => (
                      <option key={theme.id} value={theme.id}>
                        {theme.name}
                      </option>
                    ))}
                  </select>
                </div>

                <div className={styles.row}>
                  <label htmlFor="light-theme-select">Light mode theme</label>
                  <select
                    id="light-theme-select"
                    value={preferredLightTheme}
                    onChange={(e) =>
                      setPreferredLightTheme(e.target.value as ThemeId)
                    }
                    className={styles.select}
                  >
                    {lightThemes.map((theme) => (
                      <option key={theme.id} value={theme.id}>
                        {theme.name}
                      </option>
                    ))}
                  </select>
                </div>
              </>
            )}

            {/* Accent Color Selection */}
            <div className={styles.row}>
              <label>Accent Color</label>
              <div className={styles.accentColorGrid}>
                {ACCENT_COLORS.map((color) => (
                  <button
                    key={color.value}
                    className={`${styles.accentColorOption} ${
                      accentColor === color.value ? styles.selected : ''
                    }`}
                    style={{ backgroundColor: color.value }}
                    onClick={() => setAccentColor(color.value)}
                    title={color.name}
                    aria-label={`Set accent color to ${color.name}`}
                  >
                    {accentColor === color.value && (
                      <span className={styles.accentColorCheck}>&#10003;</span>
                    )}
                  </button>
                ))}
              </div>
            </div>
          </section>

          {/* Keyboard Shortcuts Section */}
          <section className={styles.section}>
            <h3>Keyboard Shortcuts</h3>
            <p className={styles.hint}>
              Press <kbd className={styles.kbd}>?</kbd> to view all keyboard
              shortcuts
            </p>
          </section>
        </div>
      </div>
    </div>
  );
}
