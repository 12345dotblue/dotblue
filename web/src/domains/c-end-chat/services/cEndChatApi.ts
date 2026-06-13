import axios from 'axios';
import { BACKEND_URL } from '../../../config';
import { casdoorService } from '../../identity/CasdoorService';

type Headers = Record<string, string>;

export type AgentSummary = {
  id: string;
  agentName: string;
  description?: string;
};

export type AgentEntryConfig = {
  id?: string;
  enabled: boolean;
  defaultAccessMode: 'standalone' | 'share' | 'embed';
  allowAnonymous: boolean;
  allowFileUpload: boolean;
  themeMode: 'auto' | 'light' | 'dark';
  compactHeader: boolean;
  sessionTtlSeconds: number;
  refreshBeforeSeconds: number;
};

export type EmbedConfig = {
  id?: string;
  allowedOrigins: string[];
  themeMode: 'auto' | 'light' | 'dark';
  compactHeader: boolean;
  allowFileUpload: boolean;
};

export type ShareLink = {
  id: string;
  shareCode: string;
  status: string;
  hasPassword?: boolean;
  allowContinueChat: boolean;
  allowAnonymous: boolean;
  conversationId?: string;
  expiresAt?: string;
};

function authHeaders(): Headers {
  const token = casdoorService.getToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

function sessionHeaders(sessionToken?: string): Headers {
  return sessionToken ? { Authorization: `Bearer ${sessionToken}` } : {};
}

function unwrapApiError(error: unknown): never {
  if (axios.isAxiosError(error)) {
    const responseMessage = typeof error.response?.data === 'string'
      ? error.response.data
      : error.response?.data?.message;
    throw new Error(responseMessage || error.message);
  }
  throw error instanceof Error ? error : new Error('request failed');
}

export async function listAgents() {
  try {
    const res = await axios.get<AgentSummary[]>(`${BACKEND_URL}/api/agents`, { headers: authHeaders() });
    return res.data ?? [];
  } catch (error) {
    return unwrapApiError(error);
  }
}

export async function getAgentEntry(agentId: string) {
  try {
    const res = await axios.get<{ config: AgentEntryConfig; embedConfig: EmbedConfig | null; shareLinks: ShareLink[] }>(
      `${BACKEND_URL}/api/admin/c-end-chat/agents/${agentId}`,
      { headers: authHeaders() },
    );
    return res.data;
  } catch (error) {
    return unwrapApiError(error);
  }
}

export async function saveAgentEntry(agentId: string, payload: AgentEntryConfig) {
  try {
    const res = await axios.put<AgentEntryConfig>(
      `${BACKEND_URL}/api/admin/c-end-chat/agents/${agentId}`,
      payload,
      { headers: authHeaders() },
    );
    return res.data;
  } catch (error) {
    return unwrapApiError(error);
  }
}

export async function saveEmbedConfig(agentId: string, payload: EmbedConfig) {
  try {
    const res = await axios.put<EmbedConfig>(
      `${BACKEND_URL}/api/admin/c-end-chat/agents/${agentId}/embed-config`,
      payload,
      { headers: authHeaders() },
    );
    return res.data;
  } catch (error) {
    return unwrapApiError(error);
  }
}

export async function createShareLink(payload: {
  agentId: string;
  conversationId?: string;
  password?: string;
  allowContinueChat: boolean;
  allowAnonymous: boolean;
  maxAccessCount?: number;
}) {
  try {
    const res = await axios.post<{ shareLink: ShareLink; shareUrl: string }>(
      `${BACKEND_URL}/api/admin/c-end-chat/share-links`,
      payload,
      { headers: authHeaders() },
    );
    return res.data;
  } catch (error) {
    return unwrapApiError(error);
  }
}

export async function revokeShareLink(id: string) {
  try {
    await axios.post(`${BACKEND_URL}/api/admin/c-end-chat/share-links/${id}/revoke`, {}, { headers: authHeaders() });
  } catch (error) {
    return unwrapApiError(error);
  }
}

export async function createEmbedToken(agentId: string, origin: string) {
  try {
    const res = await axios.post<{ embedToken: string; expiresInSeconds: number }>(
      `${BACKEND_URL}/api/admin/c-end-chat/agents/${agentId}/embed-token`,
      { origin },
      { headers: authHeaders() },
    );
    return res.data;
  } catch (error) {
    return unwrapApiError(error);
  }
}

export async function createStandaloneSession(agentId: string) {
  try {
    const res = await axios.post<{ sessionToken: string; expiresInSeconds: number; refreshBeforeSeconds: number; allowFileUpload: boolean; agentName: string }>(
      `${BACKEND_URL}/api/public/c-end-chat/agents/${agentId}/session`,
      {},
    );
    return res.data;
  } catch (error) {
    return unwrapApiError(error);
  }
}

export async function resolveShare(shareCode: string) {
  try {
    const res = await axios.get(`${BACKEND_URL}/api/public/c-end-chat/share-links/${shareCode}`);
    return res.data;
  } catch (error) {
    return unwrapApiError(error);
  }
}

export async function verifyShare(shareCode: string, password: string) {
  try {
    const res = await axios.post<{ sessionToken: string; expiresInSeconds: number; allowFileUpload: boolean }>(
      `${BACKEND_URL}/api/public/c-end-chat/share-links/${shareCode}/verify`,
      { password },
    );
    return res.data;
  } catch (error) {
    return unwrapApiError(error);
  }
}

export async function exchangeEmbedSession(embedToken: string, agentId: string) {
  try {
    const res = await axios.post<{ sessionToken: string; expiresInSeconds: number; allowFileUpload: boolean }>(
      `${BACKEND_URL}/api/public/c-end-chat/embed/session`,
      { embedToken, agentId },
    );
    return res.data;
  } catch (error) {
    return unwrapApiError(error);
  }
}

export async function createConversation(sessionToken: string) {
  try {
    const res = await axios.post(
      `${BACKEND_URL}/api/public/c-end-chat/conversations`,
      {},
      { headers: sessionHeaders(sessionToken) },
    );
    return res.data;
  } catch (error) {
    return unwrapApiError(error);
  }
}

export async function getConversationMessages(sessionToken: string, conversationId: string) {
  try {
    const res = await axios.get(
      `${BACKEND_URL}/api/public/c-end-chat/conversations/${conversationId}/messages?limit=50`,
      { headers: sessionHeaders(sessionToken) },
    );
    return res.data ?? [];
  } catch (error) {
    return unwrapApiError(error);
  }
}

export async function uploadPublicFile(
  sessionToken: string,
  conversationId: string,
  file: File,
  kind: 'image' | 'file',
) {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('conversationId', conversationId);
  formData.append('kind', kind);
  try {
    const res = await axios.post(
      `${BACKEND_URL}/api/public/c-end-chat/files`,
      formData,
      { headers: sessionHeaders(sessionToken) },
    );
    return res.data;
  } catch (error) {
    return unwrapApiError(error);
  }
}

export function getPublicSessionHeaders(sessionToken: string): { headers: Headers } {
  return {
    headers: sessionHeaders(sessionToken),
  };
}
