/**
 * @vitest-environment jsdom
 */
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import axios from 'axios';
import i18n from '../../i18n/config';
import EnterpriseCreditSettingsTab from './EnterpriseCreditSettingsTab';

vi.mock('axios');

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

function mockEnterpriseCreditGets() {
  mockedAxiosGet.mockImplementation((url) => {
    const target = String(url);
    if (target.includes('/api/admin/credits/overview')) {
      return Promise.resolve({ data: { enterpriseId: 'ent-1', wallets: [{ creditType: 'enterprise', availableCredits: 120 }] } } as any);
    }
    if (target.includes('/api/admin/credits/wallets')) {
      return Promise.resolve({ data: [{ id: 'wallet-1', creditType: 'enterprise', totalCredits: 200, reservedCredits: 20, availableCredits: 180 }] } as any);
    }
    if (target.includes('/api/admin/credits/grants')) {
      return Promise.resolve({ data: [] } as any);
    }
    if (target.includes('/api/admin/credits/ledger')) {
      return Promise.resolve({ data: [] } as any);
    }
    if (target.includes('/api/admin/llm-models')) {
      return Promise.resolve({
        data: [
          { id: 'model-ent', displayName: 'Enterprise Model', model: 'ent-model', fundingType: 'enterprise_funded', modelSourceType: 'enterprise_custom_model' },
          { id: 'model-plat', displayName: 'Platform Model', model: 'plat-model', fundingType: 'platform_funded', modelSourceType: 'platform_model' },
        ],
      } as any);
    }
    if (target.includes('/api/admin/credit-price-books')) {
      return Promise.resolve({ data: [] } as any);
    }
    if (target.includes('/api/admin/credit-budget-policies')) {
      return Promise.resolve({ data: [] } as any);
    }
    if (target.includes('/api/agents')) {
      return Promise.resolve({ data: [] } as any);
    }
    if (target.includes('/api/admin/members')) {
      return Promise.resolve({ data: [] } as any);
    }
    return Promise.resolve({ data: [] } as any);
  });
}

describe('EnterpriseCreditSettingsTab', () => {
  beforeEach(async () => {
    localStorage.clear();
    localStorage.setItem('casdoor_token', 'mock-token');
    mockedAxiosGet.mockReset();
    mockedAxiosPost.mockReset();
    installDomMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('renders core credit sections in English', async () => {
    localStorage.setItem('i18nextLng', 'en');
    await i18n.changeLanguage('en');
    mockEnterpriseCreditGets();

    render(<EnterpriseCreditSettingsTab />);

    await waitFor(() => expect(mockedAxiosGet).toHaveBeenCalled());
    expect((await screen.findAllByText('Wallets')).length).toBeGreaterThan(0);
    expect(screen.getAllByText('Platform Model Credit Overrides').length).toBeGreaterThan(0);
    expect(screen.getAllByRole('button', { name: /New Grant/i }).length).toBeGreaterThan(0);
  });

  it('renders localized credit sections in zh-CN', async () => {
    localStorage.setItem('i18nextLng', 'zh-CN');
    await i18n.changeLanguage('zh-CN');
    mockEnterpriseCreditGets();

    render(<EnterpriseCreditSettingsTab />);

    await waitFor(() => expect(mockedAxiosGet).toHaveBeenCalled());
    expect((await screen.findAllByText('钱包')).length).toBeGreaterThan(0);
    expect(screen.getAllByText('平台模型点数特配').length).toBeGreaterThan(0);
    expect(screen.getAllByRole('button', { name: /新建发放/ }).length).toBeGreaterThan(0);
  });

  it('submits the selected settlement currency when creating an enterprise price book', async () => {
    const user = userEvent.setup();
    localStorage.setItem('i18nextLng', 'en');
    await i18n.changeLanguage('en');
    mockEnterpriseCreditGets();
    mockedAxiosPost.mockResolvedValue({ data: { id: 'pb-ent-1' } } as any);

    render(<EnterpriseCreditSettingsTab />);

    await waitFor(() => expect(mockedAxiosGet).toHaveBeenCalled());
    await user.click(screen.getAllByRole('button', { name: /new enterprise price book/i })[0]);

    const dialogs = await screen.findAllByRole('dialog');
    const dialog = dialogs[dialogs.length - 1];
    const currencyCombobox = within(dialog).getAllByRole('combobox')[2];
    await user.click(currencyCombobox);
    await user.click(screen.getAllByText('CNY').at(-1) as HTMLElement);
    await user.click(within(dialog).getByRole('button', { name: 'OK' }));

    await waitFor(() => expect(mockedAxiosPost).toHaveBeenCalled());
    expect(mockedAxiosPost.mock.calls[0][0]).toContain('/api/admin/credit-price-books');
    expect(mockedAxiosPost.mock.calls[0][1]).toMatchObject({ currency: 'CNY' });
  });
});
