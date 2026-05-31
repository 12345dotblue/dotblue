/**
 * @vitest-environment jsdom
 */
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import axios from 'axios';
import EnterpriseSkillsTab from './EnterpriseSkillsTab';

vi.mock('axios');
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

const mockedAxiosGet = vi.mocked(axios.get);
const mockedAxiosPost = vi.mocked(axios.post);

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

    render(<EnterpriseSkillsTab createSignal={0} />);

    expect(await screen.findByText('enterprise_admin_skills_governance_title')).toBeTruthy();
    expect(await screen.findByText('enterprise.knowledge')).toBeTruthy();
  });

  it('启用 skill 时调用企业启用接口', async () => {
    const user = userEvent.setup();
    mockedAxiosGet.mockResolvedValue({
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
    mockedAxiosPost.mockResolvedValue({ data: {} } as any);

    render(<EnterpriseSkillsTab createSignal={0} />);

    expect(screen.getAllByText('enterprise_admin_skills_scope_tag').length).toBeGreaterThan(0);
    const catalogSwitch = await screen.findByRole('button', { name: /enterprise_admin_skills_view_catalog \(1\)/i });
    await user.click(catalogSwitch);
    expect(await screen.findByText('knowledge.search')).toBeTruthy();
    const row = screen.getByText('knowledge.search').closest('tr');
    expect(row).toBeTruthy();
    await user.click(within(row as HTMLElement).getByRole('button'));
    await user.click(screen.getAllByRole('combobox')[0]);
    await user.click((await screen.findAllByText('knowledge.search · 知识检索')).at(-1)!);
    await user.click(screen.getAllByRole('button').at(-1)!);

    await waitFor(() => expect(mockedAxiosPost).toHaveBeenCalled());
    expect(mockedAxiosPost.mock.calls[0][0]).toContain('/api/admin/skills/skill-1/enable');
  });

  it('createSignal 触发时打开企业自有 skill 创建弹窗', async () => {
    mockedAxiosGet.mockImplementation((url) => {
      const target = String(url);
      if (target.includes('/api/admin/skills?view=governance') || target.includes('/api/admin/skills?view=catalog')) {
        return Promise.resolve({ data: [] } as any);
      }
      throw new Error(`unexpected get ${target}`);
    });

    render(<EnterpriseSkillsTab createSignal={1} />);

    expect((await screen.findAllByText('enterprise_admin_skills_action_create')).length).toBeGreaterThan(0);
    expect(screen.getByPlaceholderText('enterprise.knowledge')).toBeTruthy();
  });
});
