import { useEffect, useRef, useState, useCallback } from 'react';
import { useTheme } from '../contexts/ThemeContext';
import { useAccentColor } from '../hooks/useAccentColor';
import {
  getThemesByAppearance,
  getCustomThemes,
  registerCustomTheme,
  removeCustomTheme,
  validateTheme,
  type ThemeId,
  type Theme,
} from '../themes';
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
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [importError, setImportError] = useState<string | null>(null);
  const [customThemes, setCustomThemes] = useState<Theme[]>(() =>
    getCustomThemes()
  );

  // Handle custom theme file import
  const handleImportTheme = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      const file = event.target.files?.[0];
      if (!file) return;

      setImportError(null);

      const reader = new FileReader();
      reader.onload = (e) => {
        try {
          const content = e.target?.result as string;
          const parsed = JSON.parse(content);

          // Validate the theme
          const result = validateTheme(parsed);
          if (!result.success) {
            setImportError(`Invalid theme: ${result.errors.join(', ')}`);
            return;
          }

          // Generate a unique ID from the file name
          const themeId = file.name
            .replace(/\.json$/i, '')
            .toLowerCase()
            .replace(/[^a-z0-9-]/g, '-');

          // Register the custom theme
          const theme: Theme = {
            id: themeId,
            name: result.theme.name,
            appearance: result.theme.appearance,
            colors: result.theme.colors,
            author: result.theme.author,
            source: result.theme.source,
          };

          registerCustomTheme(theme);
          setCustomThemes(getCustomThemes());
        } catch {
          setImportError('Failed to parse JSON file');
        }
      };

      reader.onerror = () => {
        setImportError('Failed to read file');
      };

      reader.readAsText(file);

      // Reset the input so the same file can be imported again
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
    },
    []
  );

  // Handle custom theme removal
  const handleRemoveCustomTheme = useCallback((themeId: string) => {
    removeCustomTheme(themeId);
    setCustomThemes(getCustomThemes());
  }, []);

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

          {/* Custom Themes Section */}
          <section className={styles.section}>
            <h3>Custom Themes</h3>

            {/* Import button */}
            <div className={styles.row}>
              <input
                ref={fileInputRef}
                type="file"
                accept=".json"
                onChange={handleImportTheme}
                className={styles.hiddenInput}
                id="theme-import"
              />
              <button
                className={styles.importButton}
                onClick={() => fileInputRef.current?.click()}
              >
                Import Theme
              </button>
            </div>

            {/* Import error message */}
            {importError && (
              <div className={styles.errorMessage}>{importError}</div>
            )}

            {/* List of custom themes */}
            {customThemes.length > 0 ? (
              <div className={styles.customThemesList}>
                {customThemes.map((theme) => (
                  <div key={theme.id} className={styles.customThemeItem}>
                    <div className={styles.customThemeInfo}>
                      <div
                        className={styles.customThemeSwatch}
                        style={{ backgroundColor: theme.colors.bgApp }}
                      >
                        <div
                          className={styles.customThemeAccent}
                          style={{ backgroundColor: theme.colors.accent }}
                        />
                      </div>
                      <span className={styles.customThemeName}>
                        {theme.name}
                      </span>
                    </div>
                    <button
                      className={styles.removeButton}
                      onClick={() => handleRemoveCustomTheme(theme.id)}
                      title="Remove theme"
                      aria-label={`Remove ${theme.name} theme`}
                    >
                      &times;
                    </button>
                  </div>
                ))}
              </div>
            ) : (
              <p className={styles.hint}>
                No custom themes imported. Import a JSON theme file to add one.
              </p>
            )}
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
