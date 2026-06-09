/**
 * @vitest-environment jsdom
 */
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import axios from 'axios';
import AppLayout from './AppLayout';
import { ThemeModeProvider } from '../../theme/themeMode';

vi.mock('axios');
vi.mock('react-i18next', async () => {
  const actual = await vi.importActual<typeof import('react-i18next')>('react-i18next');
  return {
    ...actual,
    useTranslation: () => ({
      t: (key: string) => key,
      i18n: {
        language: 'zh-CN',
        resolvedLanguage: 'zh-CN',
      },
    }),
  };
});
vi.mock('../../domains/identity/CasdoorService', () => ({
  casdoorService: {
    isAdmin: () => true,
    isAuthenticated: () => true,
    getUsername: () => 'admin',
    getToken: () => 'mock-token',
    removeToken: vi.fn(),
  },
}));

const mockedAxios = vi.mocked(axios, true);

function LocationProbe() {
  const location = useLocation();
  return <div data-testid="location-path">{location.pathname}</div>;
}

describe('AppLayout', () => {
  beforeEach(() => {
    mockedAxios.get.mockReset();
    mockedAxios.post.mockReset();
    mockedAxios.get.mockImplementation((url: string) => {
      if (url.includes('/api/enterprises/current')) {
        return Promise.resolve({
          data: {
            enterpriseId: 'ent-1',
          },
        });
      }
      if (url.includes('/api/enterprises')) {
        return Promise.resolve({
          data: [
            {
              enterpriseId: 'ent-1',
              name: 'Acme',
              role: 'owner',
            },
          ],
        });
      }
      return Promise.resolve({ data: [] });
    });
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: (query: string) => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
      }),
    });
    class MockResizeObserver {
      observe() {}
      unobserve() {}
      disconnect() {}
    }
    Object.defineProperty(window, 'ResizeObserver', {
      writable: true,
      value: MockResizeObserver,
    });
    Element.prototype.scrollIntoView = vi.fn();
  });

  it('将技能管理归到平台管理员分组，并保持本地化导航路径', async () => {
    const user = userEvent.setup();

    render(
      <ThemeModeProvider>
        <MemoryRouter initialEntries={['/zh-CN/admin/platform']}>
          <Routes>
            <Route
              path="/:lng/*"
              element={(
                <>
                  <LocationProbe />
                  <AppLayout>
                    <div>content</div>
                  </AppLayout>
                </>
              )}
            />
          </Routes>
        </MemoryRouter>
      </ThemeModeProvider>,
    );

    await waitFor(() => expect(mockedAxios.get).toHaveBeenCalled());

    expect(screen.getAllByText('platform_admin_nav').length).toBeGreaterThan(0);
    expect(screen.getAllByText('platform_settings_nav').length).toBeGreaterThan(0);
    expect(screen.getAllByText('platform_skill_governance_nav').length).toBeGreaterThan(0);

    await user.click(screen.getAllByText('platform_skill_governance_nav')[0]);

    await waitFor(() => {
      expect(screen.getByTestId('location-path').textContent).toBe('/zh-CN/admin/platform/skills');
    });
  });
});
