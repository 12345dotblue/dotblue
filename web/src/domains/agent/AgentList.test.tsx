/**
 * @vitest-environment jsdom
 */
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import axios from 'axios';
import AgentList from './AgentList';

vi.mock('axios');
vi.mock('react-router-dom', () => ({
  useNavigate: () => vi.fn(),
}));
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, options?: Record<string, unknown>) => {
      if (options?.count !== undefined) return `${key}:${String(options.count)}`;
      return key;
    },
  }),
}));

const mockedAxios = vi.mocked(axios, true);
const mockedAxiosPost = vi.mocked(axios.post);
const mockedAxiosPut = vi.mocked(axios.put);
const mockedAxiosDelete = vi.mocked(axios.delete);

function setupAxiosGet() {
  mockedAxios.get.mockImplementation((url: string) => {
    if (url.includes('/api/agents/model-options')) {
      return Promise.resolve({
        data: {
          modelOptions: [
            {
              label: '平台模型',
              options: [
                { label: 'GPT-4o', value: 'platform:model-1' },
              ],
            },
          ],
          runtimeOptions: [
            { value: 'hermes' },
            { value: 'nanobot' },
          ],
        },
      });
    }
    if (url.includes('/usage/overview')) {
      return Promise.resolve({ data: null });
    }
    if (url.includes('/api/agents')) {
      return Promise.resolve({ data: [] });
    }
    return Promise.resolve({ data: [] });
  });
}

describe('AgentList', () => {
  beforeEach(() => {
    localStorage.clear();
    localStorage.setItem('casdoor_token', 'mock-token');
    mockedAxios.get.mockReset();
    mockedAxiosPost.mockReset();
    mockedAxiosPut.mockReset();
    mockedAxiosDelete.mockReset();
    setupAxiosGet();
    (globalThis as any).matchMedia = (query: string) => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    });
    (globalThis as any).ResizeObserver = class {
      observe() {}
      unobserve() {}
      disconnect() {}
    };
    Element.prototype.scrollIntoView = vi.fn();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('创建 agent 时提交 engineType=nanobot', async () => {
    const user = userEvent.setup();
    mockedAxiosPost.mockResolvedValue({ data: { id: 'agent-1' } } as any);

    render(<AgentList />);

    await waitFor(() => expect(mockedAxios.get).toHaveBeenCalled());
    await user.click(screen.getByText('agent_create').closest('button')!);

    await user.type(screen.getByPlaceholderText('placeholder_agent_name'), 'Nanobot Agent');
    await user.type(screen.getByPlaceholderText('placeholder_system_prompt'), 'You are nanobot');

    await user.click(await screen.findByText('agent_engine_nanobot'));

    await user.click(screen.getAllByRole('button', { name: 'agent_create' }).at(-1)!);

    await waitFor(() => expect(mockedAxiosPost).toHaveBeenCalled());
    const [, payload] = mockedAxiosPost.mock.calls[0];
    expect(payload).toMatchObject({
      agentName: 'Nanobot Agent',
      systemPrompt: 'You are nanobot',
      modelScope: 'platform',
      modelId: 'model-1',
      engineType: 'nanobot',
    });
  });

  it('列表展示返回的 engineType 标签', async () => {
    mockedAxios.get.mockImplementation((url: string) => {
      if (url.includes('/api/agents/model-options')) {
        return Promise.resolve({ data: { modelOptions: [], runtimeOptions: [{ value: 'nanobot' }] } });
      }
      if (url.includes('/usage/overview')) {
        return Promise.resolve({ data: null });
      }
      if (url.endsWith('/api/agents')) {
        return Promise.resolve({
          data: [{
            id: 'agent-1',
            agentName: 'Bot',
            systemPrompt: 'prompt',
            modelScope: 'platform',
            modelId: 'model-1',
            modelName: 'GPT-4o',
            engineType: 'nanobot',
            createdAt: new Date().toISOString(),
          }],
        });
      }
      return Promise.resolve({ data: [] });
    });

    render(<AgentList />);

    expect(await screen.findByText('Bot')).toBeTruthy();
    const matches = await screen.findAllByText('agent_engine_nanobot');
    expect(matches.length).toBeGreaterThan(0);
  });
});
