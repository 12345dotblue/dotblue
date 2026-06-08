/**
 * @vitest-environment jsdom
 */
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { cleanup, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import axios from 'axios';
import AgentSkillsPanel from './AgentSkillsPanel';

vi.mock('axios');
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, options?: Record<string, unknown>) => {
      if (options && typeof options.count !== 'undefined') {
        return `${key} ${options.count}`;
      }
      return key;
    },
  }),
}));

const mockedAxiosGet = vi.mocked(axios.get);

function installDomMocks() {
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
}

describe('AgentSkillsPanel', () => {
  beforeEach(() => {
    mockedAxiosGet.mockReset();
    installDomMocks();
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it('以单页方式展示目录和已安装清单，并支持搜索过滤', async () => {
    const user = userEvent.setup();
    mockedAxiosGet.mockImplementation((url) => {
      const target = String(url);
      if (target.includes('/api/admin/agents/agent-1/skills')) {
        return Promise.resolve({
          data: [
            {
              id: 'binding-1',
              skillId: 'skill-1',
              skillVersionId: 'version-1',
              bindingStatus: 'installed',
              entryAlias: '知识助手',
              invokeVisibility: 'auto',
              skillCode: 'knowledge.search',
              skillName: 'Knowledge Search',
              version: '1.0.0',
            },
          ],
        } as any);
      }
      if (target.includes('/api/admin/agents/agent-1/skill-catalog')) {
        return Promise.resolve({
          data: [
            {
              id: 'skill-1',
              code: 'knowledge.search',
              name: 'Knowledge Search',
              sourceType: 'builtin',
              providerType: 'native',
              trustLevel: 'platform_trusted',
              latestPublishedVersion: '1.0.0',
              enablementStatus: 'enabled',
              agentInstalled: true,
              installedVersion: '1.0.0',
              displayStatus: 'installed',
              recommendedAction: 'none',
              blockReason: '',
              blockMessage: '',
            },
            {
              id: 'skill-2',
              code: 'weather.skill',
              name: 'Weather Skill',
              sourceType: 'partner',
              providerType: 'remote_hosted',
              trustLevel: 'partner_verified',
              latestPublishedVersion: '2.1.0',
              enablementStatus: '',
              agentInstalled: false,
              installedVersion: '',
              displayStatus: 'imported_pending_enable',
              recommendedAction: 'enable_install',
              blockReason: '',
              blockMessage: '',
            },
          ],
        } as any);
      }
      throw new Error(`unexpected get ${target}`);
    });

    render(
      <AgentSkillsPanel
        agentId="agent-1"
        authHeaders={{ Authorization: 'Bearer mock-token', 'X-Enterprise-ID': 'ent-1' }}
      />,
    );

    expect(await screen.findByText('agent_skill_panel_catalog_title')).toBeTruthy();
    expect(screen.getByText('agent_skill_panel_installed_title')).toBeTruthy();
    expect(screen.getByText('agent_skill_catalog_action_enable_install')).toBeTruthy();
    expect(screen.getAllByText('knowledge.search')).toHaveLength(2);
    expect(screen.getByText('weather.skill')).toBeTruthy();

    await user.type(screen.getByPlaceholderText('agent_skill_panel_search_placeholder'), 'weather');

    await waitFor(() => expect(screen.queryAllByText('knowledge.search')).toHaveLength(0));
    expect(screen.getByText('weather.skill')).toBeTruthy();
  }, 10000);
});
