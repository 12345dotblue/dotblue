/**
 * @vitest-environment jsdom
 */
import '@testing-library/jest-dom/vitest';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ThemeModeProvider, useThemeMode } from './themeMode';

function ThemeProbe() {
  const { themeMode, resolvedTheme, setThemeMode } = useThemeMode();

  return (
    <div>
      <div data-testid="theme-mode">{themeMode}</div>
      <div data-testid="resolved-theme">{resolvedTheme}</div>
      <button type="button" onClick={() => setThemeMode('dark')}>dark</button>
      <button type="button" onClick={() => setThemeMode('system')}>system</button>
    </div>
  );
}

describe('ThemeModeProvider', () => {
  beforeEach(() => {
    localStorage.clear();
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: vi.fn().mockImplementation((query: string) => ({
        matches: query.includes('prefers-color-scheme') ? false : false,
        media: query,
        onchange: null,
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
      })),
    });
  });

  it('同步主题模式到根节点并持久化用户选择', async () => {
    const user = userEvent.setup();

    render(
      <ThemeModeProvider>
        <ThemeProbe />
      </ThemeModeProvider>,
    );

    expect(screen.getByTestId('theme-mode')).toHaveTextContent('system');
    expect(screen.getByTestId('resolved-theme')).toHaveTextContent('light');
    expect(document.documentElement.dataset.theme).toBe('light');

    await user.click(screen.getByRole('button', { name: 'dark' }));

    expect(screen.getByTestId('theme-mode')).toHaveTextContent('dark');
    expect(screen.getByTestId('resolved-theme')).toHaveTextContent('dark');
    expect(document.documentElement.dataset.theme).toBe('dark');
    expect(localStorage.getItem('dotblue_theme_mode')).toBe('dark');

    await user.click(screen.getByRole('button', { name: 'system' }));

    expect(screen.getByTestId('theme-mode')).toHaveTextContent('system');
    expect(screen.getByTestId('resolved-theme')).toHaveTextContent('light');
    expect(document.documentElement.dataset.theme).toBe('light');
  });
});
