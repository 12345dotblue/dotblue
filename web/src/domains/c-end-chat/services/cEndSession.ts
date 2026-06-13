const PREFIX = 'dotblue_c_end_session_token:';

export type CEndAccessMode = 'standalone' | 'share' | 'embed';

export function getCEndSessionStorageKey(agentId: string, accessMode: CEndAccessMode = 'standalone'): string {
  return `${PREFIX}${accessMode}:${agentId}`;
}

export type CEndSessionState = {
  token: string;
  allowFileUpload?: boolean;
  conversationId?: string;
  agentName?: string;
  accessMode?: CEndAccessMode;
};

export function saveCEndSessionToken(agentId: string, state: CEndSessionState, accessMode: CEndAccessMode = 'standalone'): void {
  sessionStorage.setItem(getCEndSessionStorageKey(agentId, accessMode), JSON.stringify({ ...state, accessMode }));
}

export function loadCEndSessionState(agentId: string, accessMode: CEndAccessMode = 'standalone'): CEndSessionState | null {
  const raw = sessionStorage.getItem(getCEndSessionStorageKey(agentId, accessMode));
  if (!raw) {
    return null;
  }
  try {
    return JSON.parse(raw) as CEndSessionState;
  } catch {
    sessionStorage.removeItem(getCEndSessionStorageKey(agentId, accessMode));
    return null;
  }
}

export function removeCEndSessionToken(agentId: string, accessMode: CEndAccessMode = 'standalone'): void {
  sessionStorage.removeItem(getCEndSessionStorageKey(agentId, accessMode));
}
