import React, { useState, useEffect, useCallback, useRef } from 'react';
import { Bubble, Sender, Welcome, Conversations, FileCard } from '@ant-design/x';
import {
  Typography, Space, theme, Tooltip, Avatar, Button, Empty, Collapse,
  Input, Dropdown, Layout, Tag, Alert,
} from 'antd';
import {
  RobotOutlined, UserOutlined, ThunderboltOutlined, PlusOutlined, BulbOutlined,
  DeleteOutlined, SearchOutlined, MenuFoldOutlined,
  MenuUnfoldOutlined, AppstoreOutlined, GlobalOutlined, LogoutOutlined,
  PaperClipOutlined, CloseCircleOutlined,
} from '@ant-design/icons';
import Markdown from '@ant-design/x-markdown';
import { useTranslation } from 'react-i18next';
import { useLocation, useNavigate } from 'react-router-dom';
import axios from 'axios';
import { useXChat } from '@ant-design/x-sdk';
import { BACKEND_URL } from '../../config';
import { LANGUAGE_OPTIONS, applyLanguagePreference, getLocalizedPath, resolveSupportedLanguage, stripLanguagePrefix } from '../../i18n/config';
import { casdoorService } from '../identity/CasdoorService';
import { getOrCreateProvider } from './SSEChatProvider';
import { resolveUploadErrorMessage } from './uploadError';
import type {
  AttachmentItem,
  ChatInput,
  ChatMessage,
  MessagePart,
  SSEChunk,
  ToolCallItem,
} from './SSEChatProvider';

const { Text } = Typography;

// --- Types ---

interface AgentOption {
  id: string;
  agentName: string;
  engineType?: 'hermes' | 'nanobot';
}

interface ConversationItem {
  key: string;
  title: string;
  label: React.ReactNode;
  group: string;
  agentId: string;
  agentName: string;
  updatedAt: string;
}

interface VerifyChatState {
  verifyAgentId?: string;
  verifyAgentName?: string;
  verifySkillName?: string;
  source?: string;
}

interface PendingUpload {
  uid: string;
  fileId?: string;
  name: string;
  mimeType: string;
  size: number;
  kind: 'image' | 'file';
  previewUrl?: string;
  downloadUrl?: string;
  width?: number;
  height?: number;
  status: 'uploading' | 'uploaded' | 'failed';
  error?: string;
}

const renderConversationLabel = (title: string | undefined, agentName: string | undefined, t: any): React.ReactNode => {
  const base = title || t('chat_untitled');
  const agent = ((agentName || '').trim()) || t('chat_select_agent');
  return (
    <Space size={8}>
      <Tooltip title={agent}>
        <Avatar size={20} icon={<RobotOutlined />} />
      </Tooltip>
      <span>{base}</span>
    </Space>
  );
};

const mapConversationFromApi = (c: any, getGroupLabel: (dateStr: string) => string, t: any): ConversationItem => ({
  key: c.id,
  title: c.title || t('chat_untitled'),
  label: renderConversationLabel(c.title, c.agentName, t),
  group: getGroupLabel(c.updatedAt),
  agentId: c.agentId,
  agentName: c.agentName || '',
  updatedAt: c.updatedAt,
});

function formatEngineLabel(engineType: string | undefined, t: any): string {
  return engineType === 'nanobot' ? t('agent_engine_nanobot') : t('agent_engine_hermes');
}

function formatMessageTime(value?: string, locale = 'zh-CN'): string {
  if (!value) return '';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '';

  const now = new Date();
  const isToday = date.toDateString() === now.toDateString();
  return new Intl.DateTimeFormat(locale, isToday
    ? { hour: '2-digit', minute: '2-digit' }
    : { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' }).format(date);
}

function resolveBackendAssetUrl(url?: string): string | undefined {
  if (!url) return undefined;
  return /^https?:\/\//i.test(url) ? url : `${BACKEND_URL}${url}`;
}

const CURRENT_ENTERPRISE_STORAGE_KEY = 'dotblue_current_enterprise_id';

function getChatAuthHeaderMap(tokenOverride?: string | null): Record<string, string> {
  const token = tokenOverride ?? localStorage.getItem('casdoor_token');
  const enterpriseId = localStorage.getItem(CURRENT_ENTERPRISE_STORAGE_KEY)?.trim();
  const headers: Record<string, string> = {};
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }
  if (enterpriseId) {
    headers['X-Enterprise-ID'] = enterpriseId;
  }
  return headers;
}

async function fetchAuthorizedBlob(url: string, token: string): Promise<Blob> {
  const response = await axios.get(url, {
    responseType: 'blob',
    headers: getChatAuthHeaderMap(token),
  });
  return response.data as Blob;
}

function revokeObjectUrlLater(objectUrl: string): void {
  window.setTimeout(() => {
    URL.revokeObjectURL(objectUrl);
  }, 60_000);
}

function downloadBlob(blob: Blob, fileName: string, fallbackFileName: string): void {
  const objectUrl = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = objectUrl;
  link.download = fileName || fallbackFileName;
  link.rel = 'noopener noreferrer';
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  revokeObjectUrlLater(objectUrl);
}

interface AuthenticatedFileCardProps {
  name: string;
  byte: number;
  type: 'image' | 'file';
  previewUrl?: string;
  downloadUrl?: string;
  description?: string;
  loading?: boolean;
  token?: string | null;
}

const AuthenticatedFileCard: React.FC<AuthenticatedFileCardProps> = ({
  name,
  byte,
  type,
  previewUrl,
  downloadUrl,
  description,
  loading,
  token,
}) => {
  const { t } = useTranslation();
  const [previewSrc, setPreviewSrc] = useState<string>();
  const [previewLoading, setPreviewLoading] = useState(false);
  const [previewError, setPreviewError] = useState<string>();

  useEffect(() => {
    let disposed = false;
    let currentObjectUrl: string | undefined;
    const jwt = (token || '').trim();
    const targetUrl = type === 'image' ? resolveBackendAssetUrl(previewUrl || downloadUrl) : undefined;

    setPreviewSrc(undefined);
    setPreviewError(undefined);
    if (!targetUrl || !jwt) {
      setPreviewLoading(false);
      return undefined;
    }

    setPreviewLoading(true);
    void fetchAuthorizedBlob(targetUrl, jwt)
      .then((blob) => {
        const objectUrl = URL.createObjectURL(blob);
        if (disposed) {
          URL.revokeObjectURL(objectUrl);
          return;
        }
        currentObjectUrl = objectUrl;
        setPreviewSrc(objectUrl);
      })
      .catch(() => {
        if (!disposed) {
          setPreviewError(t('chat_preview_load_failed'));
        }
      })
      .finally(() => {
        if (!disposed) {
          setPreviewLoading(false);
        }
      });

    return () => {
      disposed = true;
      if (currentObjectUrl) {
        URL.revokeObjectURL(currentObjectUrl);
      }
    };
  }, [type, previewUrl, downloadUrl, token]);

  const handleOpen = useCallback(async () => {
    const jwt = (token || '').trim();
    const targetUrl = resolveBackendAssetUrl(type === 'image' ? (previewUrl || downloadUrl) : downloadUrl);
    if (!targetUrl || !jwt) return;
    const blob = await fetchAuthorizedBlob(targetUrl, jwt);
    if (type === 'image') {
      const objectUrl = URL.createObjectURL(blob);
      window.open(objectUrl, '_blank', 'noopener,noreferrer');
      revokeObjectUrlLater(objectUrl);
      return;
    }
    downloadBlob(blob, name, t('common_download'));
  }, [type, previewUrl, downloadUrl, token, name, t]);

  if (type === 'image') {
    return (
      <div
        onClick={() => { void handleOpen(); }}
        style={{
          width: '100%',
          maxWidth: 360,
          border: '1px solid #f0f0f0',
          borderRadius: 14,
          padding: 8,
          background: '#fff',
          cursor: 'pointer',
        }}
      >
        {previewSrc ? (
          <img
            src={previewSrc}
            alt={name}
            style={{
              display: 'block',
              width: '100%',
              maxHeight: 220,
              objectFit: 'contain',
              borderRadius: 10,
              background: '#fafafa',
            }}
          />
        ) : (
          <div
            style={{
              height: 160,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              borderRadius: 10,
              background: '#fafafa',
              color: '#8c8c8c',
              fontSize: 12,
            }}
          >
            {loading || previewLoading ? t('chat_image_loading') : t('chat_image_unavailable')}
          </div>
        )}
        <div style={{ marginTop: 8, fontSize: 12, color: '#262626', wordBreak: 'break-all' }}>
          {name}
        </div>
        <div style={{ marginTop: 4, fontSize: 12, color: '#8c8c8c' }}>
          {previewError || description}
        </div>
      </div>
    );
  }

  return (
    <FileCard
      name={name}
      byte={byte}
      type={type}
      loading={loading || previewLoading}
      description={previewError || description}
      onClick={() => { void handleOpen(); }}
    />
  );
}

function pendingUploadsToParts(items: PendingUpload[], content: string): MessagePart[] {
  const parts: MessagePart[] = [];
  const trimmed = content.trim();
  if (trimmed) {
    parts.push({ type: 'text', text: trimmed });
  }
  items
    .filter((item) => item.status === 'uploaded' && item.fileId)
    .forEach((item) => {
      parts.push({
        type: item.kind,
        fileId: item.fileId,
        name: item.name,
        mimeType: item.mimeType,
        size: item.size,
        previewUrl: item.previewUrl,
        downloadUrl: item.downloadUrl,
        width: item.width,
        height: item.height,
      });
    });
  return parts;
}

function messageAttachments(msg?: ChatMessage): AttachmentItem[] {
  if (!msg) return [];
  if (msg.attachments?.length) return msg.attachments;
  return (msg.parts || [])
    .filter((part) => !!part.fileId && (part.type === 'image' || part.type === 'file'))
    .map((part) => ({
      fileId: part.fileId!,
      kind: part.type === 'image' ? 'image' : 'file',
      name: part.name || '',
      mimeType: part.mimeType || '',
      size: part.size || 0,
      previewUrl: part.previewUrl,
      downloadUrl: part.downloadUrl,
      width: part.width,
      height: part.height,
      status: 'uploaded',
    }));
}

function getToolStatusMeta(status: string, t: any): { label: string; color: string; background: string; border: string } {
  switch (status) {
    case 'running':
      return {
        label: t('chat_tool_status_running'),
        color: '#1677ff',
        background: '#e6f4ff',
        border: '#91caff',
      };
    case 'done':
      return {
        label: t('chat_tool_status_done'),
        color: '#389e0d',
        background: '#f6ffed',
        border: '#b7eb8f',
      };
    default:
      return {
        label: t('chat_tool_status_recorded'),
        color: '#595959',
        background: '#fafafa',
        border: '#d9d9d9',
      };
  }
}

function normalizeFinishedToolCalls(toolCalls?: ToolCallItem[]): ToolCallItem[] | undefined {
  if (!toolCalls?.length) return undefined;
  return toolCalls.map((toolCall) => (
    toolCall.status === 'running'
      ? { ...toolCall, status: 'done' }
      : toolCall
  ));
}

interface ConversationPaneProps {
  conversationId: string;
  selectedAgentId: string | null;
  selectedAgentName?: string;
  runtimeFooter: string;
  provider: ReturnType<typeof getOrCreateProvider> | undefined;
  authHeaders: () => { headers: Record<string, string> };
  getJwt: () => string | null;
  t: any;
  bubbleRole: any;
}

const ConversationPane: React.FC<ConversationPaneProps> = ({
  conversationId,
  selectedAgentId,
  selectedAgentName,
  runtimeFooter,
  provider,
  authHeaders,
  getJwt,
  t,
  bubbleRole,
}) => {
  const senderRef = useRef<any>(null);
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const [pendingUploads, setPendingUploads] = useState<PendingUpload[]>([]);
  const [reloadSeq, setReloadSeq] = useState(0);
  const previousRequestingRef = useRef(false);
  const uploadingCount = pendingUploads.filter((item) => item.status === 'uploading').length;
  const assetToken = getJwt();

  const { onRequest, isRequesting, abort, parsedMessages } = useXChat<ChatMessage, ChatMessage, ChatInput, SSEChunk>({
    provider,
    conversationKey: `${conversationId}:${reloadSeq}`,
    defaultMessages: conversationId
      ? async () => {
          const jwt = getJwt();
          if (!jwt) return [];
          try {
            const res = await axios.get(`${BACKEND_URL}/api/conversations/${conversationId}/messages?limit=50`, authHeaders());
            return (res.data || []).map((m: any, i: number) => ({
              id: m.id || `hist_${String(i).padStart(6, '0')}`,
              message: {
                role: m.role,
                content: m.content,
                ...(m.parts ? { parts: m.parts } : {}),
                ...(m.attachments ? { attachments: m.attachments } : {}),
                ...(m.thinking ? { thinking: m.thinking } : {}),
                ...(m.toolCalls ? { toolCalls: normalizeFinishedToolCalls(m.toolCalls) } : {}),
                ...(m.createdAt ? { createdAt: m.createdAt } : {}),
              } as ChatMessage,
              status: 'local' as const,
            }));
          } catch {
            return [];
          }
        }
      : [],
    requestPlaceholder: { role: 'assistant', content: '' },
    requestFallback: { role: 'assistant', content: t('chat_thinking') },
  });

  useEffect(() => {
    setPendingUploads([]);
  }, [conversationId]);

  useEffect(() => {
    setReloadSeq(0);
  }, [conversationId]);

  useEffect(() => {
    if (previousRequestingRef.current && !isRequesting && conversationId) {
      // Reload persisted messages after each request so the UI converges
      // even when the streaming transport closes without a clean finalize step.
      setReloadSeq((value) => value + 1);
    }
    previousRequestingRef.current = isRequesting;
  }, [conversationId, isRequesting]);

  const uploadFiles = useCallback(async (files: FileList | File[]) => {
    const list = Array.from(files || []);
    if (!conversationId || list.length === 0) return;
    const jwt = getJwt();
    if (!jwt) return;

    const seedItems: PendingUpload[] = list.map((file) => ({
      uid: `${file.name}-${file.size}-${file.lastModified}-${Math.random().toString(36).slice(2, 8)}`,
      name: file.name,
      mimeType: file.type || 'application/octet-stream',
      size: file.size,
      kind: file.type.startsWith('image/') ? 'image' : 'file',
      status: 'uploading',
    }));
    setPendingUploads((prev) => [...prev, ...seedItems]);

    await Promise.all(seedItems.map(async (item, index) => {
      const currentFile = list[index];
      const formData = new FormData();
      formData.append('file', currentFile);
      formData.append('conversationId', conversationId);
      formData.append('kind', item.kind);
      try {
        const res = await axios.post(`${BACKEND_URL}/api/files`, formData, {
          headers: getChatAuthHeaderMap(jwt),
        });
        setPendingUploads((prev) => prev.map((upload) => (
          upload.uid === item.uid
            ? {
                ...upload,
                fileId: res.data.id,
                previewUrl: res.data.previewUrl,
                downloadUrl: res.data.downloadUrl,
                width: res.data.width,
                height: res.data.height,
                status: 'uploaded',
              }
            : upload
        )));
      } catch (error: any) {
        setPendingUploads((prev) => prev.map((upload) => (
          upload.uid === item.uid
            ? {
                ...upload,
                status: 'failed',
                error: resolveUploadErrorMessage(error, t('chat_upload_failed')),
              }
            : upload
        )));
      }
    }));
  }, [conversationId, getJwt, t]);

  const handleSend = useCallback((content: string) => {
    if (!selectedAgentId || !conversationId || uploadingCount > 0) return;
    const parts = pendingUploadsToParts(pendingUploads, content);
    if (parts.length === 0) return;
    onRequest({
      content: content.trim(),
      agentId: selectedAgentId,
      conversationId,
      parts,
    });
    setPendingUploads([]);
    senderRef.current?.clear();
  }, [selectedAgentId, conversationId, onRequest, pendingUploads, uploadingCount]);

  const handleOpenFilePicker = useCallback(() => {
    fileInputRef.current?.click();
  }, []);

  const handleRemoveUpload = useCallback((uid: string) => {
    setPendingUploads((prev) => prev.filter((item) => item.uid !== uid));
  }, []);

  const bubbleItems = [...parsedMessages].map((item) => {
    const msg = item.message as ChatMessage;
    const isUpdating = item.status === 'updating' || item.status === 'loading';
    return {
      key: item.id,
      content: msg.content,
      role: msg.role,
      loading: msg.role === 'assistant' && isUpdating && !msg.content && !msg.thinking,
      extraInfo: { chatMsg: msg, isUpdating },
    };
  });

  return (
    <>
      {parsedMessages.length === 0 ? (
        <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', padding: '32px 24px' }}>
          <div style={{ width: '100%', maxWidth: 760, borderRadius: 24, border: '1px solid #eef2f6', background: 'linear-gradient(180deg, #ffffff 0%, #f7fbff 100%)', boxShadow: '0 20px 60px rgba(15,52,96,0.08)', padding: '32px 28px' }}>
            <Welcome
              variant="borderless"
              icon={<ThunderboltOutlined style={{ color: '#1677ff', fontSize: 48 }} />}
              title={selectedAgentName || t('welcome')}
              description={t('hero_subtitle')}
            />
            <Space wrap size={[8, 8]} style={{ marginTop: 20 }}>
              <Tag color="blue">{t('hero_stat_security')}</Tag>
              <Tag color="gold">{t('highlight_runtime_metric')}</Tag>
              <Tag color="purple">{t('feat_api_title')}</Tag>
            </Space>
          </div>
        </div>
      ) : (
        <Bubble.List
          items={bubbleItems}
          role={bubbleRole}
          autoScroll
          style={{ flex: 1, padding: '24px' }}
          styles={{ root: { maxWidth: 940 } }}
        />
      )}

      <div style={{ padding: '16px 24px', borderTop: '1px solid #f0f0f0' }}>
        <input
          ref={fileInputRef}
          type="file"
          multiple
          style={{ display: 'none' }}
          onChange={(event) => {
            if (event.target.files?.length) {
              void uploadFiles(event.target.files);
              event.target.value = '';
            }
          }}
        />
        <Sender
          ref={senderRef}
          loading={isRequesting}
          onSubmit={handleSend}
          onCancel={() => abort()}
          onPasteFile={(files) => { void uploadFiles(files); }}
          placeholder={conversationId ? t('chat_placeholder') : t('chat_select_conversation_first')}
          disabled={!conversationId || uploadingCount > 0}
          header={pendingUploads.length > 0 ? (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 8 }}>
              {pendingUploads.map((item) => (
                <div key={item.uid} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <AuthenticatedFileCard
                      name={item.name}
                      byte={item.size}
                      type={item.kind === 'image' ? 'image' : 'file'}
                      previewUrl={item.previewUrl}
                      downloadUrl={item.downloadUrl}
                      loading={item.status === 'uploading'}
                      description={item.error || item.mimeType}
                      token={assetToken}
                    />
                  </div>
                  <Button
                    type="text"
                    icon={<CloseCircleOutlined />}
                    onClick={() => handleRemoveUpload(item.uid)}
                    disabled={item.status === 'uploading'}
                  />
                </div>
              ))}
            </div>
          ) : null}
          suffix={(oriNode) => (
            <Space size={4}>
              <Tooltip title={t('chat_upload_attachment')}>
                <Button
                  type="text"
                  icon={<PaperClipOutlined />}
                  onClick={handleOpenFilePicker}
                  disabled={!conversationId || isRequesting}
                />
              </Tooltip>
              {oriNode}
            </Space>
          )}
          allowSpeech
        />
        <div style={{ textAlign: 'center', marginTop: 8 }}>
          <Text type="secondary" style={{ fontSize: 12 }}>
            {runtimeFooter}
          </Text>
        </div>
      </div>
    </>
  );
};

// --- Component ---

const ChatPage: React.FC = () => {
  const { t, i18n } = useTranslation();
  const { token } = theme.useToken();
  const navigate = useNavigate();
  const location = useLocation();
  const currentLanguage = resolveSupportedLanguage(i18n?.resolvedLanguage || i18n?.language);
  const initialVerifyState = ((location.state as VerifyChatState | null) || null);

  // Agents
  const [agents, setAgents] = useState<AgentOption[]>([]);
  const [agentsLoading, setAgentsLoading] = useState(true);
  const [selectedAgentId, setSelectedAgentId] = useState<string | null>(null);

  // UI-only state
  const [searchText, setSearchText] = useState('');
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [convLoading, setConvLoading] = useState(false);
  const [convHasMore, setConvHasMore] = useState(true);
  const [convNextCursor, setConvNextCursor] = useState('');
  const [verifyHint, setVerifyHint] = useState<VerifyChatState | null>(
    initialVerifyState?.verifyAgentId ? initialVerifyState : null,
  );
  const verifyFlowHandledRef = useRef(false);

  const getJwt = useCallback(() => localStorage.getItem('casdoor_token'), []);
  const authHeaders = useCallback(() => ({
    headers: getChatAuthHeaderMap(getJwt()),
  }), [getJwt]);
  const assetToken = getJwt();

  const [conversations, setConversations] = useState<ConversationItem[]>([]);
  const [curConvId, setCurConvId] = useState('');

  const upsertConversation = useCallback((conversation: ConversationItem, placement: 'prepend' | 'append' = 'append') => {
    setConversations(prev => {
      const next = prev.filter(item => item.key !== conversation.key);
      return placement === 'prepend' ? [conversation, ...next] : [...next, conversation];
    });
  }, []);

  useEffect(() => {
    const state = (location.state as VerifyChatState | null) || null;
    if (!state?.verifyAgentId) {
      return;
    }
    setVerifyHint(state);
    verifyFlowHandledRef.current = false;
    navigate(getLocalizedPath('/chat', currentLanguage), { replace: true, state: null });
  }, [location.state, navigate, currentLanguage]);

  // Provider events — title updates from SSE stream
  const onTitleUpdated = useCallback((convId: string, title: string) => {
    setConversations(prev => prev.map(conv => (
      conv.key === convId
        ? { ...conv, title, label: renderConversationLabel(title, conv.agentName, t) }
        : conv
    )));
  }, [t]);

  // X-SDK: Chat messages
  const provider = selectedAgentId
    ? getOrCreateProvider(selectedAgentId, { onTitleUpdated })
    : undefined;

  // --- Fetch agents ---
  useEffect(() => {
    const jwt = getJwt();
    if (!jwt) { setAgentsLoading(false); return; }
    axios.get(`${BACKEND_URL}/api/agents`, authHeaders()).then(res => {
      const list: AgentOption[] = (res.data || [])
        .map((a: any) => ({
          id: a.id,
          agentName: a.agentName,
          engineType: a.engineType || 'hermes',
        }))
        .filter((agent: AgentOption, index: number, all: AgentOption[]) =>
          all.findIndex((candidate) => candidate.id === agent.id) === index,
        );
      setAgents(list);
    }).catch(() => setAgents([])).finally(() => setAgentsLoading(false));
  }, []);

  // --- Fetch conversations from server ---
  const fetchConversations = useCallback((cursor?: string) => {
    const jwt = getJwt();
    if (!jwt) return;
    setConvLoading(true);
    const params = new URLSearchParams();
    params.set('limit', '20');
    if (cursor) params.set('cursor', cursor);
    axios.get(`${BACKEND_URL}/api/conversations?${params}`, authHeaders()).then(res => {
      const currentConvs = conversationsRef.current;
      const currentActiveId = curConvIdRef.current;
      const { items, hasMore, nextCursor } = res.data;
      const mapped = (items || []).map((c: any) => mapConversationFromApi(c, getGroupLabel, t));
      if (cursor) {
        const existingIds = new Set(currentConvs.map(c => c.key));
        const newItems = mapped.filter((m: ConversationItem) => !existingIds.has(m.key));
        setConversations([...currentConvs, ...newItems]);
      } else {
        const serverIds = new Set(mapped.map((c: ConversationItem) => c.key));
        const localOnly = currentConvs.filter(c => !serverIds.has(c.key));
        setConversations([...localOnly, ...mapped]);
        if (!currentActiveId && mapped.length > 0) {
          setCurConvId(mapped[0].key);
        }
      }
      setConvHasMore(hasMore);
      setConvNextCursor(nextCursor || '');
    }).catch(() => {}).finally(() => setConvLoading(false));
  }, [getJwt, authHeaders, t]);

  useEffect(() => { fetchConversations(); }, []);

  // Update agent selection when switching conversations
  useEffect(() => {
    if (!curConvId || conversations.length === 0) return;
    const conv = conversations.find(c => c.key === curConvId);
    if (conv && conv.agentId && conv.agentId !== selectedAgentId) {
      setSelectedAgentId(conv.agentId);
    }
  }, [curConvId, conversations, selectedAgentId]);

  // --- New conversation ---
  const handleNewConversation = useCallback((agentId: string) => {
    const jwt = getJwt();
    if (!jwt) return;
    const knownIds = new Set(conversationsRef.current.map(conv => conv.key));
    axios.post(`${BACKEND_URL}/api/conversations`, { agentId }, authHeaders()).then(async (res) => {
      const data = res.data;
      const fallbackAgentId = data.agentId || agentId;
      const fallbackAgentName = data.agentName || '';
      const fallbackConversation: ConversationItem = {
        key: data.id,
        title: data.title || t('chat_untitled'),
        label: renderConversationLabel(data.title, fallbackAgentName, t),
        group: getGroupLabel(new Date().toISOString()),
        agentId: fallbackAgentId,
        agentName: fallbackAgentName,
        updatedAt: new Date().toISOString(),
      };

      try {
        const listRes = await axios.get(`${BACKEND_URL}/api/conversations?limit=20`, authHeaders());
        const latestItems: ConversationItem[] = (listRes.data?.items || [])
          .map((item: any) => mapConversationFromApi(item, getGroupLabel, t));

        const resolvedConversation = latestItems.find(item =>
          !knownIds.has(item.key) && item.agentId === fallbackAgentId,
        ) || latestItems.find(item => item.key === data.id) || fallbackConversation;

        setConversations(prev => {
          const serverIds = new Set(latestItems.map(item => item.key));
          const localOnly = prev.filter(item => !serverIds.has(item.key));
          const merged = [...localOnly, ...latestItems].filter(item => item.key !== resolvedConversation.key);
          return [resolvedConversation, ...merged];
        });
        setCurConvId(resolvedConversation.key);
        setSelectedAgentId(resolvedConversation.agentId);
      } catch {
        upsertConversation(fallbackConversation, 'prepend');
        setCurConvId(fallbackConversation.key);
        setSelectedAgentId(fallbackConversation.agentId);
      }
    }).catch(() => {});
  }, [getJwt, authHeaders, t, upsertConversation]);

  useEffect(() => {
    if (agentsLoading || !verifyHint?.verifyAgentId || verifyFlowHandledRef.current) {
      return;
    }
    const targetAgent = agents.find((item) => item.id === verifyHint.verifyAgentId);
    if (!targetAgent) {
      verifyFlowHandledRef.current = true;
      return;
    }
    verifyFlowHandledRef.current = true;
    setSelectedAgentId(targetAgent.id);
    handleNewConversation(targetAgent.id);
  }, [agents, agentsLoading, handleNewConversation, verifyHint]);

  // --- Delete conversation ---
  const curConvIdRef = useRef(curConvId);
  curConvIdRef.current = curConvId;
  const conversationsRef = useRef(conversations);
  conversationsRef.current = conversations;

  const handleDeleteConversation = useCallback((convId: string) => {
    const jwt = getJwt();
    if (!jwt) return;
    // Determine next active conversation before async
    const isActive = convId === curConvIdRef.current;
    const remaining = conversationsRef.current.filter(c => c.key !== convId);
    const nextActive = isActive && remaining.length > 0 ? remaining[0].key : (isActive ? '' : undefined);
    axios.delete(`${BACKEND_URL}/api/conversations/${convId}`, authHeaders()).then(() => {
      setConversations(prev => prev.filter(conv => conv.key !== convId));
      if (nextActive !== undefined) {
        setCurConvId(nextActive);
      }
    }).catch(() => {});
  }, [getJwt, authHeaders]);

  // --- Load more conversations ---
  const handleLoadMoreConvs = useCallback(() => {
    if (convHasMore && convNextCursor && !convLoading) {
      fetchConversations(convNextCursor);
    }
  }, [convHasMore, convNextCursor, convLoading, fetchConversations]);

  // --- Date grouping ---
  const getGroupLabel = (dateStr: string) => {
    const date = new Date(dateStr);
    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const yesterday = new Date(today.getTime() - 86400000);
    const weekAgo = new Date(today.getTime() - 7 * 86400000);
    if (date >= today) return t('chat_today');
    if (date >= yesterday) return t('chat_yesterday');
    if (date >= weekAgo) return t('chat_last_7_days');
    return t('chat_earlier');
  };

  // --- Filter ---
  const filteredConversations = searchText
    ? conversations.filter(c =>
        (c.title || '').toLowerCase().includes(searchText.toLowerCase()) ||
        (c.agentName || '').toLowerCase().includes(searchText.toLowerCase()),
      )
    : conversations;

  const currentLanguageLabel = LANGUAGE_OPTIONS.find((option) => option.value === currentLanguage)?.shortLabel || currentLanguage.toUpperCase();
  const verifyAgent = verifyHint?.verifyAgentId ? agents.find((item) => item.id === verifyHint.verifyAgentId) : null;

  // --- No agents ---
  if (!agentsLoading && agents.length === 0) {
    return (
      <div style={{ height: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'linear-gradient(180deg, #f4f7f9 0%, #ffffff 100%)', padding: 24 }}>
        <div style={{ width: '100%', maxWidth: 560, borderRadius: 24, background: '#fff', border: '1px solid #eef2f6', boxShadow: '0 20px 60px rgba(15,52,96,0.08)', padding: 32, textAlign: 'center' }}>
          <Space wrap size={[8, 8]} style={{ justifyContent: 'center', marginBottom: 16 }}>
            <Tag color="blue">{t('hero_stat_security')}</Tag>
            <Tag color="cyan">{t('feat_multi_title')}</Tag>
            <Tag color="purple">{t('feat_api_title')}</Tag>
          </Space>
          <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={t('chat_no_agents')}>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate(getLocalizedPath('/dashboard', currentLanguage))}>
              {t('chat_create_first_agent')}
            </Button>
          </Empty>
        </div>
      </div>
    );
  }

  const selectedAgent = agents.find(a => a.id === selectedAgentId);
  const runtimeFooter = t('chat_runtime_footer', {
    engine: formatEngineLabel(selectedAgent?.engineType, t),
  });
  const renderTimestamp = (value?: string, align: 'left' | 'right' = 'left') => {
    const formatted = formatMessageTime(value, i18n.language || 'zh-CN');
    if (!formatted) return null;
    return (
      <div style={{ marginTop: 8, textAlign: align }}>
        <Text type="secondary" style={{ fontSize: 11, letterSpacing: 0.2 }}>
          {formatted}
        </Text>
      </div>
    );
  };

  const renderToolCalls = (toolCalls: ToolCallItem[]) => {
    if (!toolCalls.length) return null;
    return (
      <div style={{
        marginBottom: 12,
        padding: '10px 12px',
        border: '1px solid #eef2f6',
        borderRadius: 14,
        background: 'linear-gradient(180deg, #fcfdff 0%, #f7faff 100%)',
        boxShadow: 'inset 0 1px 0 rgba(255,255,255,0.8)',
      }}
      >
        <div style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          marginBottom: 8,
        }}
        >
          <Text type="secondary" style={{ fontSize: 12, fontWeight: 600 }}>
            {t('chat_tool_calls')}
          </Text>
          <Text type="secondary" style={{ fontSize: 11 }}>
            {t('chat_tool_steps', { count: toolCalls.length })}
          </Text>
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          {toolCalls.map((tc: ToolCallItem, i: number) => {
            const statusMeta = getToolStatusMeta(tc.status, t);
            return (
              <div
                key={`${tc.tool}-${i}`}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  gap: 12,
                  padding: '8px 10px',
                  borderRadius: 12,
                  background: '#fff',
                  border: '1px solid #f0f0f0',
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: 10, minWidth: 0 }}>
                  <div style={{
                    width: 28,
                    height: 28,
                    borderRadius: 10,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    background: '#f5f7fa',
                    fontSize: 15,
                    flexShrink: 0,
                  }}
                  >
                    {tc.emoji}
                  </div>
                  <div style={{ minWidth: 0 }}>
                    <div style={{ fontSize: 13, fontWeight: 600, color: token.colorText }}>
                      {tc.label}
                    </div>
                    <Text type="secondary" style={{ fontSize: 11 }}>
                      {tc.tool}
                    </Text>
                  </div>
                </div>
                <Tag
                  bordered={false}
                  style={{
                    marginInlineEnd: 0,
                    borderRadius: 999,
                    paddingInline: 10,
                    color: statusMeta.color,
                    background: statusMeta.background,
                    border: `1px solid ${statusMeta.border}`,
                  }}
                >
                  {statusMeta.label}
                </Tag>
              </div>
            );
          })}
        </div>
      </div>
    );
  };

  const renderAttachments = (attachments: AttachmentItem[], align: 'left' | 'right') => {
    if (!attachments.length) return null;
    return (
      <div style={{
        display: 'flex',
        flexDirection: 'column',
        gap: 8,
        alignItems: align === 'right' ? 'flex-end' : 'flex-start',
        marginTop: 10,
      }}
      >
        {attachments.map((attachment, index) => (
          <div key={`${attachment.fileId}-${index}`} style={{ maxWidth: 360, width: '100%' }}>
            <AuthenticatedFileCard
              name={attachment.name}
              byte={attachment.size}
              type={attachment.kind === 'image' ? 'image' : 'file'}
              previewUrl={attachment.previewUrl}
              downloadUrl={attachment.downloadUrl}
              description={attachment.mimeType}
              token={assetToken}
            />
          </div>
        ))}
      </div>
    );
  };

  // --- Bubble role config ---
  const bubbleRole = {
    assistant: {
      placement: 'start' as const,
      avatar: <Avatar size={32} icon={<RobotOutlined />} style={{ background: token.colorPrimary }} />,
      name: selectedAgent?.agentName || t('chat_assistant_name'),
      contentRender: (_: string, { extraInfo }: { extraInfo?: any }) => {
        const msg: ChatMessage = extraInfo?.chatMsg;
        if (!msg) return null;
        return (
          <div>
            {msg.toolCalls && msg.toolCalls.length > 0 && renderToolCalls(msg.toolCalls)}
            {msg.thinking && (
              <Collapse
                size="small"
                defaultActiveKey={['thinking']}
                style={{ marginBottom: 10, background: 'transparent' }}
                items={[{
                  key: 'thinking',
                  label: <Space><BulbOutlined style={{ color: '#faad14' }} /><Text type="secondary" style={{ fontSize: 12 }}>{t('chat_thinking_process')}</Text></Space>,
                  children: (
                    <div style={{ fontSize: 13, color: '#595959' }}>
                      <Markdown>{msg.thinking}</Markdown>
                    </div>
                  ),
                }]}
              />
            )}
            <div style={{ color: token.colorText, lineHeight: 1.7 }}>
              {msg.content ? <Markdown>{msg.content}</Markdown> : t('chat_thinking')}
            </div>
            {renderAttachments(messageAttachments(msg), 'left')}
            {renderTimestamp(msg.createdAt, 'left')}
          </div>
        );
      },
    },
    user: {
      placement: 'end' as const,
      avatar: <Avatar size={32} icon={<UserOutlined />} style={{ background: '#87d068' }} />,
      contentRender: (_: string, { extraInfo }: { extraInfo?: any }) => {
        const msg: ChatMessage = extraInfo?.chatMsg;
        if (!msg) return null;
        return (
          <div>
            <div style={{ whiteSpace: 'pre-wrap', lineHeight: 1.7, color: token.colorText }}>
              {msg.content}
            </div>
            {renderAttachments(messageAttachments(msg), 'right')}
            {renderTimestamp(msg.createdAt, 'right')}
          </div>
        );
      },
    },
  };

  // --- New chat menu ---
  const newChatMenu = {
    items: agents.map(a => ({
      key: a.id,
      label: a.agentName,
      icon: <RobotOutlined />,
    })),
    onClick: ({ key }: { key: string }) => handleNewConversation(key),
  };

  return (
    <Layout style={{ height: '100vh', background: '#fff' }}>
      {/* Top bar */}
      <Layout.Header style={{
        height: 64, padding: '0 16px', display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        background: '#fff', borderBottom: '1px solid #f0f0f0', zIndex: 10,
      }}>
        <Space size={12} style={{ minWidth: 0 }}>
          <Button type="text" icon={sidebarCollapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setSidebarCollapsed(!sidebarCollapsed)} />
          <div className="chat-header-brand">
            <img
              src="/brand/dotblue-logo.png"
              alt={t('app_name')}
              className="chat-header-brand-logo"
            />
            <div className="chat-header-brand-copy">
              <Text className="chat-header-brand-title">{t('brand_header_badge')}</Text>
              <Text className="chat-header-brand-subtitle">{t('brand_header_subtitle')}</Text>
            </div>
          </div>
        </Space>
        <Space size="middle">
          <Space size={6}>
            <RobotOutlined style={{ color: token.colorPrimary }} />
            <Text style={{ fontSize: 13 }}>
              {curConvId ? (selectedAgent?.agentName || t('chat_select_agent')) : t('chat_select_conversation_first')}
            </Text>
          </Space>
          <Button
            type="default"
            size="small"
            className="chat-header-dashboard-button"
            icon={<AppstoreOutlined />}
            onClick={() => navigate(getLocalizedPath('/dashboard', currentLanguage))}
          >
            {t('go_to_dashboard')}
          </Button>
          <Dropdown menu={{
            items: LANGUAGE_OPTIONS.map((option) => ({
              key: option.value,
              label: option.label,
            })),
            onClick: async ({ key }) => {
              const resolved = await applyLanguagePreference(String(key));
              const normalizedPath = stripLanguagePrefix(location.pathname);
              navigate(`${getLocalizedPath(normalizedPath, resolved)}${location.search}${location.hash}`, { replace: true });
            },
          }} trigger={['click']}>
            <Button type="text" size="small" icon={<GlobalOutlined />}>
              {currentLanguageLabel}
            </Button>
          </Dropdown>
          <Dropdown menu={{ items: [
            { key: 'agents', label: t('agent_settings'), icon: <AppstoreOutlined />, onClick: () => navigate(getLocalizedPath('/dashboard', currentLanguage)) },
            { type: 'divider' as const },
            { key: 'logout', label: t('logout'), icon: <LogoutOutlined />, onClick: () => { casdoorService.removeToken(); window.location.href = getLocalizedPath('/login', currentLanguage); } },
          ]}}>
            <Avatar size="small" icon={<UserOutlined />} style={{ background: token.colorPrimary, cursor: 'pointer' }} />
          </Dropdown>
        </Space>
      </Layout.Header>

      <Layout>
        {/* Sidebar — Conversations */}
        {!sidebarCollapsed && (
          <Layout.Sider width={280} theme="light" style={{
            borderRight: '1px solid #f0f0f0', height: 'calc(100vh - 64px)', overflow: 'hidden',
            display: 'flex', flexDirection: 'column',
          }}>
            <div style={{ padding: '12px 12px 8px' }}>
              <Dropdown menu={newChatMenu} trigger={['click']}>
                <Button type="primary" icon={<PlusOutlined />} block>{t('chat_new_conversation')}</Button>
              </Dropdown>
              <Input
                prefix={<SearchOutlined style={{ color: '#bfbfbf' }} />}
                placeholder={t('chat_search_conversations')}
                value={searchText}
                onChange={e => setSearchText(e.target.value)}
                allowClear
                style={{ marginTop: 8 }}
                size="small"
              />
            </div>

            <div style={{ flex: 1, overflowY: 'auto' }}
              onScroll={(e) => {
                const el = e.currentTarget;
                if (el.scrollHeight - el.scrollTop - el.clientHeight < 50 && convHasMore && !convLoading) {
                  handleLoadMoreConvs();
                }
              }}
            >
              <Conversations
                items={filteredConversations}
                activeKey={curConvId || undefined}
                onActiveChange={(key) => setCurConvId(key as string)}
                groupable
                menu={(conv) => ({
                  items: [
                    { key: 'delete', label: t('chat_delete_conversation'), icon: <DeleteOutlined />, danger: true },
                  ],
                  onClick: ({ key: actionKey }) => {
                    if (actionKey === 'delete') {
                      handleDeleteConversation(conv.key as string);
                    }
                  },
                })}
              />
              {convLoading && <div style={{ textAlign: 'center', padding: 12 }}><Text type="secondary" style={{ fontSize: 12 }}>{t('chat_loading_messages')}</Text></div>}
            </div>
          </Layout.Sider>
        )}

        {/* Chat area */}
        <Layout.Content style={{ display: 'flex', flexDirection: 'column', height: 'calc(100vh - 56px)', background: '#fff' }}>
          {verifyHint ? (
            <div style={{ padding: '16px 24px 0' }}>
              <Alert
                showIcon
                closable
                type="info"
                message={t('chat_verify_banner_title', {
                  skillName: verifyHint.verifySkillName || t('platform_skill_market_card_action_install_agent'),
                })}
                description={t('chat_verify_banner_desc', {
                  agentName: verifyAgent?.agentName || verifyHint.verifyAgentName || t('chat_select_agent'),
                  skillName: verifyHint.verifySkillName || t('platform_skill_market_card_action_install_agent'),
                })}
                onClose={() => setVerifyHint(null)}
              />
            </div>
          ) : null}
          <ConversationPane
            key={curConvId || 'empty'}
            conversationId={curConvId}
            selectedAgentId={selectedAgentId}
            selectedAgentName={selectedAgent?.agentName}
            runtimeFooter={runtimeFooter}
            provider={provider}
            authHeaders={authHeaders}
            getJwt={getJwt}
            t={t}
            bubbleRole={bubbleRole}
          />
        </Layout.Content>
      </Layout>
      <style>
        {`
          .chat-header-brand {
            display: flex;
            align-items: center;
            gap: 12px;
            min-width: 0;
          }

          .chat-header-brand-logo {
            width: 118px;
            height: 36px;
            object-fit: contain;
            flex-shrink: 0;
          }

          .chat-header-brand-copy {
            display: flex;
            flex-direction: column;
            gap: 1px;
            min-width: 0;
          }

          .chat-header-brand-title {
            color: #0f172a !important;
            font-size: 12px;
            font-weight: 600;
            line-height: 1.2;
            white-space: nowrap;
          }

          .chat-header-brand-subtitle {
            color: #64748b !important;
            font-size: 11px;
            line-height: 1.2;
            white-space: nowrap;
          }

          @media (max-width: 900px) {
            .chat-header-brand-copy,
            .chat-header-dashboard-button {
              display: none !important;
            }
          }

          @media (max-width: 640px) {
            .chat-header-brand-logo {
              width: 108px;
              height: 34px;
            }
          }
        `}
      </style>
    </Layout>
  );
};

export default ChatPage;

