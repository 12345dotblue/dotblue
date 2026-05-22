/**
 * @vitest-environment jsdom
 */
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { casdoorService, casdoorConfig } from './CasdoorService';

describe('CasdoorService', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should set and get token correctly', () => {
    casdoorService.setToken('mock_token_123');
    expect(casdoorService.getToken()).toBe('mock_token_123');
    expect(casdoorService.isAuthenticated()).toBe(true);
  });

  it('should remove token correctly', () => {
    casdoorService.setToken('mock_token_123');
    casdoorService.removeToken();
    expect(casdoorService.getToken()).toBeNull();
    expect(casdoorService.isAuthenticated()).toBe(false);
  });

  it('should generate correct sign-in url', () => {
    const url = casdoorService.getSigninUrl();
    expect(url).toContain(casdoorConfig.clientId);
  });

  it('should decode JWT and extract groups', () => {
    const payload = JSON.stringify({ owner: 'dotblue', name: 'admin', groups: ['admin'], isAdmin: true });
    const fakePayload = btoa(payload);
    const fakeToken = `header.${fakePayload}.signature`;
    casdoorService.setToken(fakeToken);
    expect(casdoorService.getOrganization()).toBe('dotblue');
    expect(casdoorService.getUsername()).toBe('admin');
    expect(casdoorService.getGroups()).toEqual(['admin']);
    expect(casdoorService.isAdmin()).toBe(true);
  });

  it('should detect admin via groups when isAdmin flag is false', () => {
    const payload = JSON.stringify({ owner: 'dotblue', name: 'groupadmin', groups: ['admin'], isAdmin: false });
    const fakePayload = btoa(payload);
    const fakeToken = `header.${fakePayload}.signature`;
    casdoorService.setToken(fakeToken);
    expect(casdoorService.isAdmin()).toBe(true);
  });

  it('should not be admin for regular user', () => {
    const payload = JSON.stringify({ owner: 'dotblue', name: 'user1', groups: [], isAdmin: false });
    const fakePayload = btoa(payload);
    const fakeToken = `header.${fakePayload}.signature`;
    casdoorService.setToken(fakeToken);
    expect(casdoorService.isAdmin()).toBe(false);
    expect(casdoorService.getGroups()).toEqual([]);
  });

  it('should return empty values when not authenticated', () => {
    expect(casdoorService.getOrganization()).toBe('');
    expect(casdoorService.getUsername()).toBe('');
    expect(casdoorService.getGroups()).toEqual([]);
    expect(casdoorService.isAdmin()).toBe(false);
  });

  it('should treat malformed token as authenticated but return empty claims', () => {
    casdoorService.setToken('malformed-token');
    expect(casdoorService.isAuthenticated()).toBe(true);
    expect(casdoorService.getOrganization()).toBe('');
    expect(casdoorService.getUsername()).toBe('');
  });
});
