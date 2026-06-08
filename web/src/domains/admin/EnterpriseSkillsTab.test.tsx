/**
 * @vitest-environment jsdom
 */
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { cleanup, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import axios from 'axios';
import { MemoryRouter } from 'react-router-dom';
import EnterpriseSkillsTab from './EnterpriseSkillsTab';

vi.mock('axios');
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

const mockedAxiosGet = vi.mocked(axios.get);
const mockedAxiosPost = vi.mocked(axios.post);

function renderInRouter(initialEntry = '/admin/enterprise?tab=skills') {
  return render(
    <MemoryRouter initialEntries={[initialEntry]}>
      <EnterpriseSkillsTab createSignal={0} />
    </MemoryRouter>,
  );
}

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

describe('EnterpriseSkillsTab', () => {
  beforeEach(() => {
    localStorage.clear();
    localStorage.setItem('casdoor_token', 'mock-token');
    mockedAxiosGet.mockReset();
    mockedAxiosPost.mockReset();
    installDomMocks();
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it('默认展示企业治理视图并加载企业自有 skill', async () => {
    mockedAxiosGet.mockImplementation((url) => {
      const target = String(url);
      if (target.includes('/api/admin/skills?view=governance')) {
        return Promise.resolve({
          data: [
            {
              id: 'skill-ent-1',
              code: 'enterprise.knowledge',
              name: '企业知识库',
              sourceType: 'builtin',
              trustLevel: 'enterprise_verified',
              status: 'draft',
              latestPublishedVersion: '',
              enablementStatus: '',
            },
          ],
        } as any);
      }
      if (target.includes('/api/admin/skills?view=catalog')) {
        return Promise.resolve({ data: [] } as any);
      }
      throw new Error(`unexpected get ${target}`);
    });

    renderInRouter();

    expect(await screen.findByText('enterprise_admin_skills_governance_title')).toBeTruthy();
    expect(await screen.findByText('enterprise.knowledge')).toBeTruthy();
  });

  it('开放 skill 时调用企业开放接口', async () => {
    const user = userEvent.setup();
    mockedAxiosGet.mockImplementation((url) => {
      const target = String(url);
      if (target.includes('/api/admin/skills?view=governance')) {
        return Promise.resolve({ data: [] } as any);
      }
      if (target.includes('/api/admin/skills?view=catalog')) {
        return Promise.resolve({
          data: [
            {
              id: 'skill-1',
              code: 'knowledge.search',
              name: '知识检索',
              sourceType: 'builtin',
              trustLevel: 'platform_trusted',
              status: 'published',
              latestPublishedVersion: '1.0.0',
              enablementStatus: '',
            },
          ],
        } as any);
      }
      throw new Error(`unexpected get ${target}`);
    });
    mockedAxiosPost.mockResolvedValue({ data: {} } as any);

    renderInRouter();

    expect(screen.getAllByText('enterprise_admin_skills_scope_tag').length).toBeGreaterThan(0);
    const catalogSwitch = await screen.findByRole('button', { name: /enterprise_admin_skills_view_catalog \(1\)/i });
    await user.click(catalogSwitch);
    expect(await screen.findByText('knowledge.search')).toBeTruthy();
    const row = screen.getByText('knowledge.search').closest('tr');
    expect(row).toBeTruthy();
    await user.click(within(row as HTMLElement).getByRole('button', { name: 'enterprise_admin_skills_action_enable' }));
    const dialog = await screen.findByRole('dialog');
    await user.click(within(dialog).getByRole('button', { name: 'enterprise_admin_skills_action_enable' }));

    await waitFor(() => expect(mockedAxiosPost).toHaveBeenCalled());
    expect(mockedAxiosPost.mock.calls[0][0]).toContain('/api/admin/skills/skill-1/enable');
  }, 10000);

  it('createSignal 触发时打开企业自有 skill 创建弹窗', async () => {
    mockedAxiosGet.mockImplementation((url) => {
      const target = String(url);
      if (target.includes('/api/admin/skills?view=governance') || target.includes('/api/admin/skills?view=catalog')) {
        return Promise.resolve({ data: [] } as any);
      }
      throw new Error(`unexpected get ${target}`);
    });

    render(
      <MemoryRouter initialEntries={['/admin/enterprise?tab=skills']}>
        <EnterpriseSkillsTab createSignal={1} />
      </MemoryRouter>,
    );

    expect((await screen.findAllByText('enterprise_admin_skills_action_create')).length).toBeGreaterThan(0);
    expect(await screen.findByRole('dialog')).toBeTruthy();
  });

  it('带 skillId 进入企业页时自动打开目录视图并预选启用目标 skill', async () => {
    mockedAxiosPost.mockResolvedValue({ data: {} } as any);
    mockedAxiosGet.mockImplementation((url) => {
      const target = String(url);
      if (target.includes('/api/admin/skills?view=governance')) {
        return Promise.resolve({ data: [] } as any);
      }
      if (target.includes('/api/admin/skills?view=catalog')) {
        return Promise.resolve({
          data: [
            {
              id: 'skill-rollout-1',
              code: 'weather.tencent',
              name: '天气 Skill',
              sourceType: 'partner',
              trustLevel: 'partner_verified',
              status: 'published',
              latestPublishedVersion: '1.0.0',
              enablementStatus: '',
            },
          ],
        } as any);
      }
      throw new Error(`unexpected get ${target}`);
    });

    renderInRouter('/admin/enterprise?tab=skills&skillId=skill-rollout-1');

    expect(await screen.findByText('enterprise_admin_skills_catalog_title')).toBeTruthy();
    const dialog = await screen.findByRole('dialog');
    await userEvent.setup().click(within(dialog).getByRole('button', { name: 'enterprise_admin_skills_action_enable' }));

    await waitFor(() => expect(mockedAxiosPost).toHaveBeenCalled());
    expect(mockedAxiosPost.mock.calls[0][0]).toContain('/api/admin/skills/skill-rollout-1/enable');
  });
});
