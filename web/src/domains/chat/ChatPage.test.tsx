/**
 * @vitest-environment jsdom
 */
import React from 'react';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import '@testing-library/jest-dom/vitest';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import axios from 'axios';
import ChatPage from './ChatPage';
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
vi.mock('../identity/CasdoorService', () => ({
  casdoorService: {
    removeToken: vi.fn(),
  },
}));
vi.mock('./SSEChatProvider', () => ({
  getOrCreateProvider: () => ({}),
}));
vi.mock('@ant-design/x-sdk', () => ({
  useXChat: () => ({
    onRequest: vi.fn(),
    isRequesting: false,
    abort: vi.fn(),
    parsedMessages: [],
  }),
}));
vi.mock('@ant-design/x-markdown', () => ({
  default: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));
vi.mock('@ant-design/x', async () => {
  const ReactModule = await vi.importActual<typeof import('react')>('react');

  return {
    Bubble: {
      List: ({ items }: { items: Array<{ key: string }> }) => <div data-testid="bubble-list">{items.length}</div>,
    },
    Sender: ReactModule.forwardRef(function MockSender() {
      return <div data-testid="sender" />;
    }),
    Welcome: ({ title, description }: { title: React.ReactNode; description: React.ReactNode }) => (
      <div data-testid="welcome">
        <div>{title}</div>
        <div>{description}</div>
      </div>
    ),
    Conversations: ({
      items,
      onActiveChange,
    }: {
      items: Array<{ key: string; title: string }>;
      onActiveChange?: (key: string) => void;
    }) => (
      <div data-testid="conversations">
        {items.map((item) => (
          <button key={item.key} type="button" onClick={() => onActiveChange?.(item.key)}>
            {item.title}
          </button>
        ))}
      </div>
    ),
    FileCard: ({ name }: { name: string }) => <div>{name}</div>,
  };
});

const mockedAxios = vi.mocked(axios, true);

describe('ChatPage', () => {
  beforeEach(() => {
    mockedAxios.get.mockReset();
    mockedAxios.post.mockReset();
    mockedAxios.delete.mockReset();

    localStorage.clear();
    localStorage.setItem('casdoor_token', 'mock-token');

    mockedAxios.get.mockImplementation((url: string) => {
      if (url.includes('/api/agents')) {
        return Promise.resolve({
          data: [
            {
              id: 'agent-1',
              agentName: 'Demo Agent',
              engineType: 'hermes',
            },
          ],
        });
      }

      if (url.includes('/api/conversations?')) {
        return Promise.resolve({
          data: {
            items: [
              {
                id: 'conv-1',
                title: 'First conversation',
                agentId: 'agent-1',
                agentName: 'Demo Agent',
                updatedAt: new Date().toISOString(),
              },
            ],
            hasMore: false,
            nextCursor: '',
          },
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

  it('将侧栏折叠按钮保留在侧边会话面板且不再渲染多余的侧栏标题', async () => {
    const user = userEvent.setup();

    render(
      <ThemeModeProvider>
        <MemoryRouter initialEntries={['/zh-CN/chat']}>
          <Routes>
            <Route path="/:lng/chat" element={<ChatPage />} />
          </Routes>
        </MemoryRouter>
      </ThemeModeProvider>,
    );

    await waitFor(() => {
      expect(mockedAxios.get).toHaveBeenCalled();
    });

    const topbar = screen.getByTestId('chat-topbar');
    const sidebarPanel = screen.getByTestId('chat-sidebar-panel');
    const sidebarToggle = within(sidebarPanel).getByTestId('chat-sidebar-toggle');

    expect(within(topbar).queryByTestId('chat-sidebar-toggle')).toBeNull();
    expect(within(topbar).getByAltText('app_name')).toBeInTheDocument();
    expect(screen.queryByText('chat')).toBeNull();
    expect(sidebarToggle).toHaveAttribute('aria-label', 'chat_collapse_sidebar');
    expect(screen.getByText('chat_new_conversation')).toBeInTheDocument();

    await user.click(sidebarToggle);

    await waitFor(() => {
      expect(screen.queryByText('chat_new_conversation')).not.toBeInTheDocument();
    });
    expect(screen.getByRole('button', { name: 'chat_expand_sidebar' })).toBeInTheDocument();
  });
});
