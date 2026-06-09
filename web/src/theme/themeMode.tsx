/* eslint-disable react-refresh/only-export-components */
import React from 'react';
import { theme as antdTheme } from 'antd';

export type ThemeMode = 'light' | 'dark' | 'system';
export type ResolvedTheme = 'light' | 'dark';

const STORAGE_KEY = 'dotblue_theme_mode';

interface ThemeModeContextValue {
  themeMode: ThemeMode;
  resolvedTheme: ResolvedTheme;
  setThemeMode: (mode: ThemeMode) => void;
}

const ThemeModeContext = React.createContext<ThemeModeContextValue | null>(null);

function isThemeMode(value: string | null): value is ThemeMode {
  return value === 'light' || value === 'dark' || value === 'system';
}

function readStoredThemeMode(): ThemeMode {
  if (typeof window === 'undefined') {
    return 'system';
  }

  try {
    const stored = window.localStorage.getItem(STORAGE_KEY);
    return isThemeMode(stored) ? stored : 'system';
  } catch {
    return 'system';
  }
}

function resolveSystemTheme(): ResolvedTheme {
  if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') {
    return 'light';
  }
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

export function buildAntdThemeConfig(resolvedTheme: ResolvedTheme) {
  const isDark = resolvedTheme === 'dark';

  return {
    algorithm: isDark ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm,
    token: {
      colorPrimary: '#1677ff',
      borderRadius: 12,
      fontFamily: '"Inter", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
      colorBgLayout: isDark ? '#0b1220' : '#f4f7f9',
      colorBgContainer: isDark ? '#101a2b' : '#ffffff',
      colorText: isDark ? '#e2e8f0' : '#0f172a',
      colorTextSecondary: isDark ? '#94a3b8' : '#64748b',
      colorBorder: isDark ? '#243244' : '#d9e2ec',
      colorBorderSecondary: isDark ? '#1b2838' : '#e6edf5',
      colorFillSecondary: isDark ? 'rgba(148, 163, 184, 0.14)' : '#f8fafc',
      boxShadow: isDark
        ? '0 16px 40px rgba(2, 8, 23, 0.36)'
        : '0 16px 40px rgba(15, 23, 42, 0.08)',
      boxShadowSecondary: isDark
        ? '0 10px 28px rgba(2, 8, 23, 0.3)'
        : '0 10px 28px rgba(15, 23, 42, 0.06)',
    },
    components: {
      Layout: {
        headerBg: isDark ? 'rgba(11, 18, 32, 0.78)' : 'rgba(255, 255, 255, 0.72)',
        siderBg: isDark ? '#08111f' : '#eef4fb',
        bodyBg: isDark ? '#0b1220' : '#f4f7f9',
      },
      Card: {
        boxShadowTertiary: isDark
          ? '0 12px 32px rgba(2, 8, 23, 0.22)'
          : '0 4px 20px rgba(15, 23, 42, 0.03)',
      },
      Button: {
        controlHeight: 40,
        paddingContentHorizontal: 20,
      },
      Menu: {
        darkItemBg: isDark ? '#08111f' : '#001529',
        darkSubMenuItemBg: isDark ? '#08111f' : '#001529',
        darkItemSelectedBg: 'rgba(22, 119, 255, 0.18)',
        itemBg: isDark ? '#08111f' : '#eef4fb',
        itemColor: isDark ? 'rgba(226, 232, 240, 0.86)' : '#334155',
        itemSelectedBg: isDark ? 'rgba(59, 130, 246, 0.24)' : 'rgba(22, 119, 255, 0.14)',
        itemSelectedColor: isDark ? '#f8fafc' : '#0f172a',
        itemHoverColor: isDark ? '#f8fafc' : '#0f172a',
      },
    },
  };
}

export const ThemeModeProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [themeMode, setThemeMode] = React.useState<ThemeMode>(() => readStoredThemeMode());
  const [systemTheme, setSystemTheme] = React.useState<ResolvedTheme>(() => resolveSystemTheme());

  React.useEffect(() => {
    if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') {
      return undefined;
    }

    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const updateSystemTheme = (event?: MediaQueryListEvent) => {
      setSystemTheme(event?.matches ?? mediaQuery.matches ? 'dark' : 'light');
    };

    updateSystemTheme();
    mediaQuery.addEventListener('change', updateSystemTheme);
    return () => mediaQuery.removeEventListener('change', updateSystemTheme);
  }, []);

  const resolvedTheme = themeMode === 'system' ? systemTheme : themeMode;

  React.useEffect(() => {
    if (typeof document === 'undefined') {
      return;
    }

    document.documentElement.dataset.theme = resolvedTheme;
    document.documentElement.style.colorScheme = resolvedTheme;

    try {
      window.localStorage.setItem(STORAGE_KEY, themeMode);
    } catch {
      // Ignore storage failures and keep the in-memory mode.
    }
  }, [resolvedTheme, themeMode]);

  const contextValue = React.useMemo<ThemeModeContextValue>(() => ({
    themeMode,
    resolvedTheme,
    setThemeMode,
  }), [themeMode, resolvedTheme]);

  return (
    <ThemeModeContext.Provider value={contextValue}>
      {children}
    </ThemeModeContext.Provider>
  );
};

export function useThemeMode(): ThemeModeContextValue {
  const context = React.useContext(ThemeModeContext);
  if (!context) {
    throw new Error('useThemeMode must be used within ThemeModeProvider');
  }
  return context;
}
