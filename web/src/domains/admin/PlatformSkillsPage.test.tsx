/**
 * @vitest-environment jsdom
 */
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import axios from 'axios';
import i18n from '../../i18n/config';
import PlatformSkillsPage from './PlatformSkillsPage';

vi.mock('axios');

const mockedAxios = vi.mocked(axios, true);
const mockedAxiosGet = vi.mocked(axios.get);
const mockedAxiosPost = vi.mocked(axios.post);
const mockedAxiosPut = vi.mocked(axios.put);

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

describe('PlatformSkillsPage', () => {
  beforeEach(async () => {
    localStorage.clear();
    localStorage.setItem('casdoor_token', 'mock-token');
    localStorage.setItem('i18nextLng', 'en');
    await i18n.changeLanguage('en');
    mockedAxios.get.mockReset();
    mockedAxiosPost.mockReset();
    mockedAxiosPut.mockReset();
    installDomMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('加载平台 skill 列表', async () => {
    mockedAxiosGet.mockImplementation((url) => {
      const target = String(url);
      if (target.includes('/api/admin/skills?view=governance')) {
        return Promise.resolve({
          data: [
            {
              id: 'skill-1',
              code: 'knowledge.search',
              name: '知识检索',
              description: 'desc',
              sourceType: 'builtin',
              providerType: 'native',
              trustLevel: 'platform_trusted',
              status: 'published',
              updatedAt: new Date().toISOString(),
            },
          ],
        } as any);
      }
      return Promise.resolve({ data: [] } as any);
    });

    render(<PlatformSkillsPage />);

    expect(await screen.findByText('knowledge.search')).toBeTruthy();
    expect(screen.getByText('知识检索')).toBeTruthy();
  });

  it('在 zh-CN 下展示中文文案', async () => {
    localStorage.setItem('i18nextLng', 'zh-CN');
    await i18n.changeLanguage('zh-CN');
    mockedAxiosGet.mockImplementation(() => Promise.resolve({ data: [] } as any));

    render(<PlatformSkillsPage />);

    await waitFor(() => expect(mockedAxiosGet).toHaveBeenCalled());
    expect(screen.getAllByRole('heading', { name: /Skill 管理/ }).length).toBeGreaterThan(0);
    expect(screen.getAllByRole('button', { name: /新建 Skill/ }).length).toBeGreaterThan(0);
  });

  it('创建 skill 时提交表单到后端', async () => {
    const user = userEvent.setup();
    mockedAxiosGet.mockImplementation(() => Promise.resolve({ data: [] } as any));
    mockedAxiosPost.mockResolvedValue({ data: { id: 'skill-1' } } as any);

    render(<PlatformSkillsPage />);

    await waitFor(() => expect(mockedAxiosGet).toHaveBeenCalled());
    await user.click(screen.getAllByRole('button', { name: /New Skill/i })[0]);
    await user.type(screen.getByPlaceholderText('knowledge.search'), 'knowledge.search');
    await user.type(screen.getByPlaceholderText('Knowledge Search'), '知识检索');
    const dialogs = await screen.findAllByRole('dialog');
    const dialog = dialogs[dialogs.length - 1];
    await user.click(within(dialog).getByRole('button', { name: 'Save' }));

    await waitFor(() => expect(mockedAxiosPost).toHaveBeenCalled());
    const [url, payload] = mockedAxiosPost.mock.calls[0];
    expect(url).toContain('/api/admin/skills');
    expect(payload).toMatchObject({
      code: 'knowledge.search',
      name: '知识检索',
      sourceType: 'builtin',
      providerType: 'native',
    });
  });

  it('可以从 import jobs 标签页发起导入', async () => {
    const user = userEvent.setup();
    mockedAxiosGet.mockImplementation((url) => {
      const target = String(url);
      if (target.includes('/api/admin/platform/skill-hubs')) {
        return Promise.resolve({
          data: [
            {
              id: 'hub-1',
              hubCode: 'partner-openapi',
              name: 'Partner Hub',
              hubType: 'openapi_hub',
              status: 'enabled',
              trustLevel: 'partner_verified',
              syncMode: 'manual',
              authScheme: 'none',
              updatedAt: new Date().toISOString(),
            },
          ],
        } as any);
      }
      return Promise.resolve({ data: [] } as any);
    });
    mockedAxiosPost.mockResolvedValue({ data: { id: 'job-1', jobStatus: 'completed' } } as any);

    render(<PlatformSkillsPage />);

    await waitFor(() => expect(mockedAxiosGet).toHaveBeenCalled());
    const importButtons = screen.getAllByRole('button', { name: /Import Jobs/i });
    await user.click(importButtons[importButtons.length - 1]);
    await user.click(screen.getByRole('button', { name: /Start Import/i }));

    const dialogs = await screen.findAllByRole('dialog');
    const dialog = dialogs[dialogs.length - 1];
    await user.click(within(dialog).getByRole('combobox'));
    await user.click(await screen.findByText(/Partner Hub/));
    await user.type(within(dialog).getByPlaceholderText('petstore/openapi.yaml'), 'petstore/openapi.yaml');
    await user.type(within(dialog).getByPlaceholderText('partner.petstore'), 'partner.petstore');
    await user.type(within(dialog).getByPlaceholderText('1.0.0'), '1.2.3');
    await user.click(within(dialog).getByRole('button', { name: 'Save' }));

    await waitFor(() => expect(mockedAxiosPost).toHaveBeenCalled());
    const [url, payload] = mockedAxiosPost.mock.calls[0];
    expect(url).toContain('/api/admin/platform/skill-import-jobs');
    expect(payload).toMatchObject({
      hubId: 'hub-1',
      sourceLocator: 'petstore/openapi.yaml',
      sourceNamespace: 'partner.petstore',
      sourceVersion: '1.2.3',
    });
  });
});
