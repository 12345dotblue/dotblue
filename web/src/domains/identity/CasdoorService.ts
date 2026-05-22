import Sdk from 'casdoor-js-sdk';

const configuredRedirect = (import.meta.env.VITE_CASDOOR_REDIRECT_URL ?? '').trim();

export const casdoorConfig = {
  serverUrl: import.meta.env.VITE_CASDOOR_SERVER_URL,
  clientId: import.meta.env.VITE_CASDOOR_CLIENT_ID,
  organizationName: import.meta.env.VITE_CASDOOR_ORG_NAME,
  appName: import.meta.env.VITE_CASDOOR_APP_NAME,
  redirectPath: configuredRedirect || "/callback",
};

class CasdoorService {
  private sdk: Sdk;
  private static readonly AUTH_CHANGED_EVENT = 'casdoor-auth-changed';

  constructor() {
    this.sdk = new Sdk(casdoorConfig);
  }

  getSignInUrl() {
    return this.sdk.getSigninUrl();
  }

  getSignUpUrl() {
    return this.sdk.getSignupUrl();
  }

  getSigninUrl() {
    return this.sdk.getSigninUrl();
  }

  async signin(url: string) {
    return this.sdk.signin(url).then((res: any) => {
      if (res.status === 'ok') {
        this.setToken(res.data);
        return res;
      }
      throw new Error(res.msg);
    });
  }

  setToken(token: string) {
    localStorage.setItem('casdoor_token', token);
    window.dispatchEvent(new Event(CasdoorService.AUTH_CHANGED_EVENT));
  }

  getToken(): string | null {
    return localStorage.getItem('casdoor_token');
  }

  removeToken() {
    localStorage.removeItem('casdoor_token');
    window.dispatchEvent(new Event(CasdoorService.AUTH_CHANGED_EVENT));
  }

  isAuthenticated(): boolean {
    return !!this.getToken();
  }

  // Decode JWT payload (without verification) to extract user info.
  private decodeToken(): { owner?: string; name?: string; isAdmin?: boolean; groups?: string[]; exp?: number; [key: string]: any } | null {
    const token = this.getToken();
    if (!token) return null;
    try {
      const parts = token.split('.');
      if (parts.length !== 3) return null;
      const payload = JSON.parse(this.decodeBase64Url(parts[1]));
      return payload;
    } catch {
      return null;
    }
  }

  private decodeBase64Url(value: string): string {
    const normalized = value
      .replace(/-/g, '+')
      .replace(/_/g, '/')
      .padEnd(Math.ceil(value.length / 4) * 4, '=');
    return atob(normalized);
  }

  private getValidTokenPayload(): { owner?: string; name?: string; isAdmin?: boolean; groups?: string[]; exp?: number; [key: string]: any } | null {
    const decoded = this.decodeToken();
    if (!decoded) {
      return null;
    }
    return decoded;
  }

  getOrganization(): string {
    const decoded = this.getValidTokenPayload();
    return decoded?.owner || '';
  }

  getUsername(): string {
    const decoded = this.getValidTokenPayload();
    return decoded?.name || '';
  }

  getGroups(): string[] {
    const decoded = this.getValidTokenPayload();
    return decoded?.groups || [];
  }

  isAdmin(): boolean {
    const decoded = this.getValidTokenPayload();
    if (decoded?.isAdmin) return true;
    const groups = decoded?.groups || [];
    return groups.includes('admin');
  }
}

export const casdoorService = new CasdoorService();
export const CASDOOR_AUTH_CHANGED_EVENT = 'casdoor-auth-changed';
