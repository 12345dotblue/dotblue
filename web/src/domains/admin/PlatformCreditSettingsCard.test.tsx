/**
 * @vitest-environment jsdom
 */
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import axios from 'axios';
import i18n from '../../i18n/config';
import PlatformCreditSettingsCard from './PlatformCreditSettingsCard';

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

function mockPlatformCreditGets() {
  mockedAxiosGet.mockImplementation((url) => {
    const target = String(url);
    if (target.includes('/api/admin/platform/llm-models')) {
      return Promise.resolve({ data: [{ id: 'model-1', displayName: 'Platform Model', model: 'plat-model' }] } as any);
    }
    if (target.includes('/api/admin/platform/credit-price-books')) {
      return Promise.resolve({ data: [] } as any);
    }
    if (target.includes('/api/enterprises')) {
      return Promise.resolve({ data: [{ enterpriseId: 'ent-1', name: 'Acme', role: 'owner' }] } as any);
    }
    return Promise.resolve({ data: [] } as any);
  });
}

describe('PlatformCreditSettingsCard', () => {
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

  it('renders platform credit sections in English', async () => {
    localStorage.setItem('i18nextLng', 'en');
    await i18n.changeLanguage('en');
    mockPlatformCreditGets();

    render(<PlatformCreditSettingsCard />);

    await waitFor(() => expect(mockedAxiosGet).toHaveBeenCalled());
    expect((await screen.findAllByText('Global Platform Credit Price Books')).length).toBeGreaterThan(0);
    expect(screen.getAllByRole('button', { name: /Load overview/i }).length).toBeGreaterThan(0);
    expect(screen.getAllByText('Enterprise Platform Credit Overrides').length).toBeGreaterThan(0);
  });

  it('renders localized platform credit sections in zh-CN', async () => {
    localStorage.setItem('i18nextLng', 'zh-CN');
    await i18n.changeLanguage('zh-CN');
    mockPlatformCreditGets();

    render(<PlatformCreditSettingsCard />);

    await waitFor(() => expect(mockedAxiosGet).toHaveBeenCalled());
    expect((await screen.findAllByText('平台全局点数价格书')).length).toBeGreaterThan(0);
    expect(screen.getAllByRole('button', { name: /加载概览/ }).length).toBeGreaterThan(0);
    expect(screen.getAllByText('企业平台点数特配').length).toBeGreaterThan(0);
  });

  it('submits the selected settlement currency when creating a platform price book', async () => {
    const user = userEvent.setup();
    localStorage.setItem('i18nextLng', 'en');
    await i18n.changeLanguage('en');
    mockPlatformCreditGets();
    mockedAxiosPost.mockResolvedValue({ data: { id: 'pb-1' } } as any);

    render(<PlatformCreditSettingsCard />);

    await waitFor(() => expect(mockedAxiosGet).toHaveBeenCalled());
    await user.click(screen.getAllByRole('button', { name: /new price book/i })[0]);

    const dialogs = await screen.findAllByRole('dialog');
    const dialog = dialogs[dialogs.length - 1];
    const currencyCombobox = within(dialog).getAllByRole('combobox')[2];
    await user.click(currencyCombobox);
    await user.click(screen.getAllByText('CNY').at(-1) as HTMLElement);
    await user.click(within(dialog).getByRole('button', { name: 'OK' }));

    await waitFor(() => expect(mockedAxiosPost).toHaveBeenCalled());
    expect(mockedAxiosPost.mock.calls[0][0]).toContain('/api/admin/platform/credit-price-books');
    expect(mockedAxiosPost.mock.calls[0][1]).toMatchObject({ currency: 'CNY' });
  });
});
