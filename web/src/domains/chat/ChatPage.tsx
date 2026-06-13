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
import ThemeModeDropdown from '../../components/ThemeModeDropdown';
import { useThemeMode } from '../../theme/themeMode';
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

interface ConversationListResponse {
  items: any[];
  hasMore: boolean;
  nextCursor?: string;
}

type ChatAuthHeaders = () => { headers: Record<string, string> };
type ChatProviderFactory = (agentId: string, events: { onTitleUpdated?: (conversationId: string, title: string) => void }) => ReturnType<typeof getOrCreateProvider>;
type ChatUploadResponse = {
  id: string;
  previewUrl?: string;
  downloadUrl?: string;
  width?: number;
  height?: number;
};

export interface ChatPageProps {
  getJwt?: () => string | null;
  authHeaders?: ChatAuthHeaders;
  listAgents?: () => Promise<AgentOption[]>;
  listConversations?: (cursor?: string) => Promise<ConversationListResponse>;
  createConversation?: (agentId: string) => Promise<any>;
  deleteConversation?: (conversationId: string) => Promise<void>;
  loadMessages?: (conversationId: string) => Promise<any[]>;
  uploadFile?: (conversationId: string, file: File, kind: 'image' | 'file') => Promise<ChatUploadResponse>;
  createProvider?: ChatProviderFactory;
  fixedAgentId?: string | null;
  fixedAgentName?: string;
  showSidebar?: boolean;
  showDashboardButton?: boolean;
  showLanguageSwitcher?: boolean;
  showUserMenu?: boolean;
  allowDeleteConversation?: boolean;
  allowFileUpload?: boolean;
  brandNavigatePath?: string;
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
const CHAT_SIDEBAR_WIDTH = 296;
const CHAT_SIDEBAR_COLLAPSED_WIDTH = 92;

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
          border: '1px solid var(--chat-card-border)',
          borderRadius: 14,
          padding: 8,
          background: 'var(--chat-card-bg)',
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
              background: 'var(--chat-muted-bg)',
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
              background: 'var(--chat-muted-bg)',
              color: 'var(--app-panel-text-muted)',
              fontSize: 12,
            }}
          >
            {loading || previewLoading ? t('chat_image_loading') : t('chat_image_unavailable')}
          </div>
        )}
        <div style={{ marginTop: 8, fontSize: 12, color: 'var(--app-panel-text)', wordBreak: 'break-all' }}>
          {name}
        </div>
        <div style={{ marginTop: 4, fontSize: 12, color: 'var(--app-panel-text-muted)' }}>
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
        color: '#60a5fa',
        background: 'rgba(59, 130, 246, 0.14)',
        border: 'rgba(96, 165, 250, 0.28)',
      };
    case 'done':
      return {
        label: t('chat_tool_status_done'),
        color: '#4ade80',
        background: 'rgba(34, 197, 94, 0.14)',
        border: 'rgba(74, 222, 128, 0.26)',
      };
    default:
      return {
        label: t('chat_tool_status_recorded'),
        color: 'var(--app-panel-text-muted)',
        background: 'var(--app-panel-muted-bg)',
        border: 'var(--app-shell-border)',
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
  getJwt: () => string | null;
  loadMessages: (conversationId: string) => Promise<any[]>;
  uploadFile?: (conversationId: string, file: File, kind: 'image' | 'file') => Promise<ChatUploadResponse>;
  allowFileUpload: boolean;
  t: any;
  bubbleRole: any;
}

const ConversationPane: React.FC<ConversationPaneProps> = ({
  conversationId,
  selectedAgentId,
  selectedAgentName,
  runtimeFooter,
  provider,
  getJwt,
  loadMessages,
  uploadFile,
  allowFileUpload,
  t,
  bubbleRole,
}) => {
  const senderRef = useRef<any>(null);
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const { token } = theme.useToken();
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
            const items = await loadMessages(conversationId);
            return (items || []).map((m: any, i: number) => ({
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
    if (!conversationId || list.length === 0 || !uploadFile) return;
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
      try {
        const res = await uploadFile(conversationId, currentFile, item.kind);
        setPendingUploads((prev) => prev.map((upload) => (
          upload.uid === item.uid
            ? {
                ...upload,
                fileId: res.id,
                previewUrl: res.previewUrl,
                downloadUrl: res.downloadUrl,
                width: res.width,
                height: res.height,
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
  }, [conversationId, getJwt, t, uploadFile]);

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
    <div
      className="chat-conversation-shell"
      style={{
        flex: 1,
        minHeight: 0,
        margin: 20,
        borderRadius: 24,
        border: '1px solid var(--chat-card-border)',
        background: 'var(--chat-card-bg)',
        boxShadow: 'var(--chat-card-shadow)',
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden',
      }}
    >
      <div
        className="chat-conversation-body"
        style={{
          flex: 1,
          minHeight: 0,
          background: parsedMessages.length === 0 ? 'var(--chat-card-bg-soft)' : 'var(--chat-card-bg)',
        }}
      >
        {parsedMessages.length === 0 ? (
          <div style={{ height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: '32px 24px' }}>
            <div style={{ width: '100%', maxWidth: 760, borderRadius: 24, border: '1px solid var(--chat-card-border)', background: 'var(--chat-card-bg-soft)', padding: '32px 28px' }}>
              <Welcome
                variant="borderless"
                icon={<ThunderboltOutlined style={{ color: token.colorPrimary, fontSize: 48 }} />}
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
            style={{ height: '100%', padding: '24px' }}
            styles={{ root: { maxWidth: 940 } }}
          />
        )}
      </div>

      <div style={{ padding: '16px 24px', borderTop: '1px solid var(--chat-input-border)', background: 'var(--chat-card-bg)' }}>
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
          onPasteFile={(files) => { if (allowFileUpload) void uploadFiles(files); }}
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
                  disabled={!conversationId || isRequesting || !allowFileUpload}
                  style={{ display: allowFileUpload ? undefined : 'none' }}
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
    </div>
  );
};

// --- Component ---

const ChatPage: React.FC<ChatPageProps> = ({
  getJwt: getJwtProp,
  authHeaders: authHeadersProp,
  listAgents: listAgentsProp,
  listConversations: listConversationsProp,
  createConversation: createConversationProp,
  deleteConversation: deleteConversationProp,
  loadMessages: loadMessagesProp,
  uploadFile: uploadFileProp,
  createProvider: createProviderProp,
  fixedAgentId = null,
  fixedAgentName,
  showSidebar = true,
  showDashboardButton = true,
  showLanguageSwitcher = true,
  showUserMenu = true,
  allowDeleteConversation = true,
  allowFileUpload = true,
  brandNavigatePath = '/',
}) => {
  const { t, i18n } = useTranslation();
  const { token } = theme.useToken();
  const { resolvedTheme } = useThemeMode();
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

  const getJwt = useCallback(() => (getJwtProp ? getJwtProp() : localStorage.getItem('casdoor_token')), [getJwtProp]);
  const authHeaders = useCallback(() => (
    authHeadersProp ? authHeadersProp() : { headers: getChatAuthHeaderMap(getJwt()) }
  ), [authHeadersProp, getJwt]);
  const assetToken = getJwt();
  const listAgents = useCallback(async () => {
    if (listAgentsProp) {
      return listAgentsProp();
    }
    const res = await axios.get(`${BACKEND_URL}/api/agents`, authHeaders());
    return (res.data || [])
      .map((a: any) => ({
        id: a.id,
        agentName: a.agentName,
        engineType: a.engineType || 'hermes',
      }))
      .filter((agent: AgentOption, index: number, all: AgentOption[]) =>
        all.findIndex((candidate) => candidate.id === agent.id) === index,
      );
  }, [authHeaders, listAgentsProp]);
  const listConversations = useCallback(async (cursor?: string) => {
    if (listConversationsProp) {
      return listConversationsProp(cursor);
    }
    const params = new URLSearchParams();
    params.set('limit', '20');
    if (cursor) params.set('cursor', cursor);
    const res = await axios.get(`${BACKEND_URL}/api/conversations?${params}`, authHeaders());
    return res.data;
  }, [authHeaders, listConversationsProp]);
  const createConversation = useCallback(async (agentId: string) => {
    if (createConversationProp) {
      return createConversationProp(agentId);
    }
    const res = await axios.post(`${BACKEND_URL}/api/conversations`, { agentId }, authHeaders());
    return res.data;
  }, [authHeaders, createConversationProp]);
  const deleteConversation = useCallback(async (conversationId: string) => {
    if (deleteConversationProp) {
      await deleteConversationProp(conversationId);
      return;
    }
    await axios.delete(`${BACKEND_URL}/api/conversations/${conversationId}`, authHeaders());
  }, [authHeaders, deleteConversationProp]);
  const loadMessages = useCallback(async (conversationId: string) => {
    if (loadMessagesProp) {
      return loadMessagesProp(conversationId);
    }
    const res = await axios.get(`${BACKEND_URL}/api/conversations/${conversationId}/messages?limit=50`, authHeaders());
    return res.data || [];
  }, [authHeaders, loadMessagesProp]);
  const uploadFile = useCallback(async (conversationId: string, file: File, kind: 'image' | 'file') => {
    if (uploadFileProp) {
      return uploadFileProp(conversationId, file, kind);
    }
    const formData = new FormData();
    formData.append('file', file);
    formData.append('conversationId', conversationId);
    formData.append('kind', kind);
    const res = await axios.post(`${BACKEND_URL}/api/files`, formData, {
      headers: getChatAuthHeaderMap(getJwt()),
    });
    return res.data;
  }, [getJwt, uploadFileProp]);
  const createProvider = useCallback<ChatProviderFactory>((agentId, events) => (
    createProviderProp ? createProviderProp(agentId, events) : getOrCreateProvider(agentId, events)
  ), [createProviderProp]);

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
    ? createProvider(selectedAgentId, { onTitleUpdated })
    : undefined;

  // --- Fetch agents ---
  useEffect(() => {
    const jwt = getJwt();
    if (!jwt) { setAgentsLoading(false); return; }
    listAgents()
      .then((items) => {
        let next = items;
        if (fixedAgentId && !items.find((item: AgentOption) => item.id === fixedAgentId)) {
          next = [{
            id: fixedAgentId,
            agentName: fixedAgentName || fixedAgentId,
            engineType: 'hermes',
          }, ...items];
        }
        setAgents(next);
        if (fixedAgentId) {
          setSelectedAgentId(fixedAgentId);
        }
      })
      .catch(() => {
        setAgents(fixedAgentId ? [{
          id: fixedAgentId,
          agentName: fixedAgentName || fixedAgentId,
          engineType: 'hermes',
        }] : []);
      })
      .finally(() => setAgentsLoading(false));
  }, [fixedAgentId, fixedAgentName, getJwt, listAgents]);

  // --- Fetch conversations from server ---
  const fetchConversations = useCallback((cursor?: string) => {
    const jwt = getJwt();
    if (!jwt) return;
    setConvLoading(true);
    listConversations(cursor).then((res) => {
      const currentConvs = conversationsRef.current;
      const currentActiveId = curConvIdRef.current;
      const { items, hasMore, nextCursor } = res;
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
  }, [getJwt, listConversations, t]);

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
    createConversation(agentId).then(async (data) => {
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
        const listRes = await listConversations();
        const latestItems: ConversationItem[] = (listRes?.items || [])
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
  }, [createConversation, getJwt, listConversations, t, upsertConversation]);

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

  useEffect(() => {
    if (agentsLoading || !fixedAgentId || curConvId || verifyHint?.verifyAgentId) {
      return;
    }
    handleNewConversation(fixedAgentId);
  }, [agentsLoading, curConvId, fixedAgentId, handleNewConversation, verifyHint]);

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
    deleteConversation(convId).then(() => {
      setConversations(prev => prev.filter(conv => conv.key !== convId));
      if (nextActive !== undefined) {
        setCurConvId(nextActive);
      }
    }).catch(() => {});
  }, [deleteConversation, getJwt]);

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
      <div style={{ height: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'var(--chat-card-bg-soft)', padding: 24 }}>
        <div style={{ width: '100%', maxWidth: 560, borderRadius: 24, background: 'var(--chat-card-bg)', border: '1px solid var(--chat-card-border)', boxShadow: 'var(--chat-card-shadow)', padding: 32, textAlign: 'center' }}>
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
        border: '1px solid var(--chat-tool-border)',
        borderRadius: 14,
        background: 'var(--chat-tool-shell-bg)',
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
                  background: 'var(--chat-tool-bg)',
                  border: '1px solid var(--chat-tool-border)',
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
                    background: 'var(--app-panel-muted-bg)',
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
                  label: <Space><BulbOutlined style={{ color: '#fbbf24' }} /><Text type="secondary" style={{ fontSize: 12 }}>{t('chat_thinking_process')}</Text></Space>,
                  children: (
                    <div style={{ fontSize: 13, color: 'var(--app-panel-text-muted)' }}>
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
      avatar: <Avatar size={32} icon={<UserOutlined />} style={{ background: token.colorPrimary }} />,
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
    <Layout style={{ height: '100vh', background: token.colorBgLayout }}>
      {/* Top bar */}
      <Layout.Header style={{
        height: 64, padding: '0 16px', display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        background: 'var(--app-panel-bg)', borderBottom: '1px solid var(--app-shell-border)', zIndex: 10,
      }} data-testid="chat-topbar">
        <Space size={12} style={{ minWidth: 0 }}>
          <div className="chat-header-brand">
            <img
              src={resolvedTheme === 'dark' ? '/brand/dotblue-logo-dark.svg' : '/brand/dotblue-logo-light.svg'}
              alt={t('app_name')}
              className="chat-header-brand-logo"
              onClick={() => navigate(getLocalizedPath(brandNavigatePath, currentLanguage))}
            />
          </div>
        </Space>
        <Space size="middle">
          <ThemeModeDropdown />
          {showDashboardButton ? <Button
            type="default"
            size="small"
            className="chat-header-dashboard-button"
            icon={<AppstoreOutlined />}
            onClick={() => navigate(getLocalizedPath('/dashboard', currentLanguage))}
          >
            {t('go_to_dashboard')}
          </Button> : null}
          {showLanguageSwitcher ? <Dropdown menu={{
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
          </Dropdown> : null}
          {showUserMenu ? <Dropdown menu={{ items: [
            { key: 'agents', label: t('agent_settings'), icon: <AppstoreOutlined />, onClick: () => navigate(getLocalizedPath('/dashboard', currentLanguage)) },
            { type: 'divider' as const },
            { key: 'logout', label: t('logout'), icon: <LogoutOutlined />, onClick: () => { casdoorService.removeToken(); window.location.href = getLocalizedPath('/login', currentLanguage); } },
          ]}}>
            <Avatar size="small" icon={<UserOutlined />} style={{ background: token.colorPrimary, cursor: 'pointer' }} />
          </Dropdown> : null}
        </Space>
      </Layout.Header>

      <Layout>
        {/* Sidebar — Conversations */}
        {showSidebar ? <Layout.Sider
          width={CHAT_SIDEBAR_WIDTH}
          collapsedWidth={CHAT_SIDEBAR_COLLAPSED_WIDTH}
          collapsed={sidebarCollapsed}
          trigger={null}
          theme={resolvedTheme === 'dark' ? 'dark' : 'light'}
          style={{
            height: 'calc(100vh - 64px)',
            overflow: 'visible',
            boxShadow: 'inset -1px 0 0 var(--chat-sidebar-border)',
          }}
        >
          <div className={`chat-sidebar-shell ${sidebarCollapsed ? 'chat-sidebar-shell--collapsed' : ''}`}>
            <div className={`chat-sidebar-panel ${sidebarCollapsed ? 'chat-sidebar-panel--collapsed' : ''}`} data-testid="chat-sidebar-panel">
              <Tooltip title={sidebarCollapsed ? t('chat_expand_sidebar') : t('chat_collapse_sidebar')}>
                <Button
                  type="text"
                  size="small"
                  className="chat-sidebar-toggle chat-sidebar-toggle--attached"
                  data-testid="chat-sidebar-toggle"
                  aria-label={sidebarCollapsed ? t('chat_expand_sidebar') : t('chat_collapse_sidebar')}
                  icon={sidebarCollapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
                  onClick={() => setSidebarCollapsed((value) => !value)}
                />
              </Tooltip>
              {!sidebarCollapsed && (
                <>
                <div className="chat-sidebar-controls">
                  <Dropdown menu={newChatMenu} trigger={['click']}>
                    <Button type="primary" icon={<PlusOutlined />} block>{t('chat_new_conversation')}</Button>
                  </Dropdown>
                  <Input
                    prefix={<SearchOutlined style={{ color: 'var(--app-nav-text-muted)' }} />}
                    placeholder={t('chat_search_conversations')}
                    value={searchText}
                    onChange={e => setSearchText(e.target.value)}
                    allowClear
                    size="small"
                  />
                </div>

                <div
                  className="chat-sidebar-conversations"
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
                    menu={allowDeleteConversation ? (conv) => ({
                      items: [
                        { key: 'delete', label: t('chat_delete_conversation'), icon: <DeleteOutlined />, danger: true },
                      ],
                      onClick: ({ key: actionKey }) => {
                        if (actionKey === 'delete') {
                          handleDeleteConversation(conv.key as string);
                        }
                      },
                    }) : undefined}
                  />
                  {convLoading && <div style={{ textAlign: 'center', padding: 12 }}><Text style={{ fontSize: 12, color: 'var(--app-nav-text-muted)' }}>{t('chat_loading_messages')}</Text></div>}
                </div>
                </>
              )}
            </div>
          </div>
        </Layout.Sider> : null}

        {/* Chat area */}
        <Layout.Content style={{ display: 'flex', flexDirection: 'column', height: 'calc(100vh - 64px)', background: 'var(--app-panel-bg)' }}>
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
            getJwt={getJwt}
            loadMessages={loadMessages}
            uploadFile={uploadFile}
            allowFileUpload={allowFileUpload}
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
            justify-content: center;
            flex-shrink: 0;
          }

          .chat-header-brand-logo {
            width: 112px;
            height: 34px;
            object-fit: contain;
            display: block;
            cursor: pointer;
          }

          .chat-sidebar-shell {
            display: flex;
            flex-direction: column;
            height: 100%;
            background: var(--chat-sidebar-bg);
            overflow: visible;
          }

          .chat-sidebar-shell--collapsed {
            align-items: center;
          }

          .chat-sidebar-panel {
            position: relative;
            display: flex;
            flex: 1;
            min-height: 0;
            flex-direction: column;
          }

          .chat-sidebar-panel--collapsed {
            align-items: center;
          }

          .chat-sidebar-toggle {
            color: var(--app-nav-text-strong) !important;
            border: 1px solid var(--chat-sidebar-border) !important;
            background: var(--chat-sidebar-surface) !important;
            box-shadow: 0 10px 24px rgba(15, 23, 42, 0.08);
          }

          .chat-sidebar-toggle:hover {
            color: var(--app-nav-text-strong) !important;
            border-color: rgba(96, 165, 250, 0.4) !important;
            background: var(--chat-sidebar-item-hover) !important;
          }

          .chat-sidebar-toggle--attached {
            position: absolute !important;
            top: 18px;
            right: -14px;
            width: 28px;
            height: 36px;
            border-radius: 0 12px 12px 0 !important;
            z-index: 2;
          }

          .chat-sidebar-controls {
            padding: 16px 12px 8px;
            display: flex;
            flex-direction: column;
            gap: 8px;
          }

          .chat-sidebar-conversations {
            flex: 1;
            overflow-y: auto;
            padding-bottom: 12px;
          }

          .chat-sidebar-shell .ant-input-affix-wrapper {
            background: var(--chat-sidebar-surface);
            border-color: var(--chat-sidebar-border);
            box-shadow: none;
          }

          .chat-sidebar-shell .ant-input-affix-wrapper .ant-input {
            background: transparent;
            color: var(--app-nav-text-strong);
          }

          .chat-sidebar-shell .ant-input-affix-wrapper .ant-input::placeholder {
            color: var(--app-nav-text-muted);
          }

          .chat-sidebar-shell .ant-input-affix-wrapper .ant-input-clear-icon {
            color: var(--app-nav-text-muted);
          }

          .chat-sidebar-shell .ant-conversations {
            background: transparent;
          }

          .chat-sidebar-shell .ant-conversations-group-title {
            color: var(--app-nav-text-muted) !important;
            padding-inline: 16px !important;
          }

          .chat-sidebar-shell .ant-conversations-item {
            margin: 4px 8px !important;
            border-radius: 12px !important;
            background: transparent !important;
            color: var(--app-nav-text) !important;
          }

          .chat-sidebar-shell .ant-conversations-item:hover {
            background: var(--chat-sidebar-item-hover) !important;
          }

          .chat-sidebar-shell .ant-conversations-item-active,
          .chat-sidebar-shell .ant-conversations-item-selected {
            background: var(--chat-sidebar-item-active) !important;
            box-shadow: inset 0 0 0 1px rgba(96, 165, 250, 0.16);
          }

          .chat-sidebar-shell .ant-conversations-item .ant-typography,
          .chat-sidebar-shell .ant-conversations-item .ant-dropdown-trigger,
          .chat-sidebar-shell .ant-conversations-item .anticon {
            color: inherit !important;
          }

          .chat-sidebar-shell .ant-conversations-item .ant-avatar {
            background: var(--chat-sidebar-item-hover) !important;
          }

          @media (max-width: 900px) {
            .chat-header-brand-logo {
              width: 92px;
            }

            .chat-header-dashboard-button {
              display: none !important;
            }
          }

          @media (max-width: 640px) {
            .chat-sidebar-toggle--attached {
              top: 14px;
            }
          }
        `}
      </style>
    </Layout>
  );
};

export default ChatPage;
