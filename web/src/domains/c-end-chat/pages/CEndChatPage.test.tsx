/**
 * @vitest-environment jsdom
 */
import '@testing-library/jest-dom/vitest';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { CEndChatPage } from './CEndChatPage';
import { saveCEndSessionToken } from '../services/cEndSession';

vi.mock('../../chat/ChatPage', () => ({
  __esModule: true,
  default: () => <div>chat-page-mock</div>,
}));

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

const createStandaloneSessionMock = vi.fn();

vi.mock('../services/cEndChatApi', async () => {
  const actual = await vi.importActual<typeof import('../services/cEndChatApi')>('../services/cEndChatApi');
  return {
    ...actual,
    createStandaloneSession: (...args: Parameters<typeof actual.createStandaloneSession>) => createStandaloneSessionMock(...args),
  };
});

describe('CEndChatPage', () => {
  beforeEach(() => {
    createStandaloneSessionMock.mockReset();
  });

  afterEach(() => {
    sessionStorage.clear();
    cleanup();
  });

  it('在 sessionToken 存在时显示成功状态', () => {
    saveCEndSessionToken('agent-1', {
      token: 'abc',
      allowFileUpload: true,
      agentName: 'Agent One',
    }, 'standalone');

    render(
      <MemoryRouter initialEntries={['/zh-CN/agents/agent-1/chat']}>
        <Routes>
          <Route path="/:lng/agents/:agentId/chat" element={<CEndChatPage />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByText('chat-page-mock')).toBeInTheDocument();
  });

  it('在 sessionToken 缺失时自动创建 standalone session', async () => {
    createStandaloneSessionMock.mockResolvedValue({
      sessionToken: 'standalone-token',
      allowFileUpload: false,
      agentName: 'Agent Two',
      expiresInSeconds: 3600,
      refreshBeforeSeconds: 300,
    });

    render(
      <MemoryRouter initialEntries={['/zh-CN/agents/agent-2/chat']}>
        <Routes>
          <Route path="/:lng/agents/:agentId/chat" element={<CEndChatPage />} />
        </Routes>
      </MemoryRouter>,
    );

    await waitFor(() => {
      expect(screen.getByText('chat-page-mock')).toBeInTheDocument();
    });
  });

  it('在 standalone session 创建失败时显示错误', async () => {
    createStandaloneSessionMock.mockRejectedValue(new Error('anonymous access is disabled'));

    render(
      <MemoryRouter initialEntries={['/zh-CN/agents/agent-3/chat']}>
        <Routes>
          <Route path="/:lng/agents/:agentId/chat" element={<CEndChatPage />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByText('c_end_chat_title')).toBeInTheDocument();
    expect(await screen.findByText('anonymous access is disabled')).toBeInTheDocument();
  });
});
