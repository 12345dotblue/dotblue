import Sdk from 'casdoor-js-sdk';

export const casdoorConfig = {
  serverUrl: import.meta.env.VITE_CASDOOR_SERVER_URL,
  clientId: import.meta.env.VITE_CASDOOR_CLIENT_ID,
  organizationName: import.meta.env.VITE_CASDOOR_ORG_NAME,
  appName: import.meta.env.VITE_CASDOOR_APP_NAME,
  redirectPath: "/callback",
};

class CasdoorService {
  private sdk: Sdk;

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
  }

  getToken(): string | null {
    return localStorage.getItem('casdoor_token');
  }

  removeToken() {
    localStorage.removeItem('casdoor_token');
  }

  isAuthenticated(): boolean {
    return !!this.getToken();
  }

  // Decode JWT payload (without verification) to extract user info
  private decodeToken(): { owner?: string; name?: string; isAdmin?: boolean; groups?: string[]; [key: string]: any } | null {
    const token = this.getToken();
    if (!token) return null;
    try {
      const parts = token.split('.');
      if (parts.length !== 3) return null;
      const payload = JSON.parse(atob(parts[1]));
      return payload;
    } catch {
      return null;
    }
  }

  getOrganization(): string {
    const decoded = this.decodeToken();
    return decoded?.owner || '';
  }

  getUsername(): string {
    const decoded = this.decodeToken();
    return decoded?.name || '';
  }

  getGroups(): string[] {
    const decoded = this.decodeToken();
    return decoded?.groups || [];
  }

  isAdmin(): boolean {
    const decoded = this.decodeToken();
    if (decoded?.isAdmin) return true;
    const groups = decoded?.groups || [];
    return groups.includes('admin');
  }
}

export const casdoorService = new CasdoorService();
