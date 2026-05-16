import React, { useState, useEffect, useCallback, useRef } from 'react';
import { Bubble, Sender, Welcome, Conversations } from '@ant-design/x';
import {
  Typography, Space, theme, Tooltip, Avatar, Button, Empty, Collapse,
  Input, Dropdown, Layout, Tag,
} from 'antd';
import {
  RobotOutlined, UserOutlined, ThunderboltOutlined, PlusOutlined, BulbOutlined,
  DeleteOutlined, SearchOutlined, MenuFoldOutlined,
  MenuUnfoldOutlined, AppstoreOutlined, GlobalOutlined, LogoutOutlined,
} from '@ant-design/icons';
import Markdown from '@ant-design/x-markdown';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import { useXChat } from '@ant-design/x-sdk';
import { BACKEND_URL } from '../../config';
import { casdoorService } from '../identity/CasdoorService';
import { getOrCreateProvider } from './SSEChatProvider';
import type { ChatInput, ChatMessage, SSEChunk, ToolCallItem } from './SSEChatProvider';

const { Text } = Typography;

// --- Types ---

interface AgentOption {
  id: string;
  agentName: string;
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
  provider: ReturnType<typeof getOrCreateProvider> | undefined;
  authHeaders: () => { headers: { Authorization: string } };
  getJwt: () => string | null;
  t: any;
  bubbleRole: any;
}

const ConversationPane: React.FC<ConversationPaneProps> = ({
  conversationId,
  selectedAgentId,
  selectedAgentName,
  provider,
  authHeaders,
  getJwt,
  t,
  bubbleRole,
}) => {
  const senderRef = useRef<any>(null);

  const { onRequest, isRequesting, abort, parsedMessages } = useXChat<ChatMessage, ChatMessage, ChatInput, SSEChunk>({
    provider,
    conversationKey: conversationId,
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

  const handleSend = useCallback((content: string) => {
    if (!content.trim() || !selectedAgentId || !conversationId) return;
    onRequest({ content, agentId: selectedAgentId, conversationId });
    senderRef.current?.clear();
  }, [selectedAgentId, conversationId, onRequest]);

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
        <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <Welcome
            variant="borderless"
            icon={<ThunderboltOutlined style={{ color: '#1677ff', fontSize: 48 }} />}
            title={selectedAgentName || t('welcome')}
            description={t('hero_subtitle')}
          />
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
        <Sender
          ref={senderRef}
          loading={isRequesting}
          onSubmit={handleSend}
          onCancel={() => abort()}
          placeholder={conversationId ? t('chat_placeholder') : t('chat_select_conversation_first')}
          disabled={!conversationId}
          allowSpeech
        />
        <div style={{ textAlign: 'center', marginTop: 8 }}>
          <Text type="secondary" style={{ fontSize: 12 }}>
            Powered by Hermes Engine &bull; gVisor Isolated
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

  const getJwt = useCallback(() => localStorage.getItem('casdoor_token'), []);
  const authHeaders = useCallback(() => ({
    headers: { Authorization: `Bearer ${getJwt()}` },
  }), [getJwt]);

  const [conversations, setConversations] = useState<ConversationItem[]>([]);
  const [curConvId, setCurConvId] = useState('');

  const upsertConversation = useCallback((conversation: ConversationItem, placement: 'prepend' | 'append' = 'append') => {
    setConversations(prev => {
      const next = prev.filter(item => item.key !== conversation.key);
      return placement === 'prepend' ? [conversation, ...next] : [...next, conversation];
    });
  }, []);

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

  // --- No agents ---
  if (!agentsLoading && agents.length === 0) {
    return (
      <div style={{ height: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#f4f7f9' }}>
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={t('chat_no_agents')}>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/dashboard')}>
            {t('chat_create_first_agent')}
          </Button>
        </Empty>
      </div>
    );
  }

  const selectedAgent = agents.find(a => a.id === selectedAgentId);

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

  // --- Bubble role config ---
  const bubbleRole = {
    assistant: {
      placement: 'start' as const,
      avatar: <Avatar size={32} icon={<RobotOutlined />} style={{ background: token.colorPrimary }} />,
      name: selectedAgent?.agentName || 'Assistant',
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
        height: 56, padding: '0 16px', display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        background: '#fff', borderBottom: '1px solid #f0f0f0', zIndex: 10,
      }}>
        <Space>
          <Button type="text" icon={sidebarCollapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setSidebarCollapsed(!sidebarCollapsed)} />
          <div style={{ width: 28, height: 28, background: `linear-gradient(135deg, ${token.colorPrimary} 0%, #36cfc9 100%)`, borderRadius: 6 }} />
          <Text strong style={{ fontSize: 16 }}>dotblue</Text>
        </Space>
        <Space size="middle">
          <Space size={6}>
            <RobotOutlined style={{ color: token.colorPrimary }} />
            <Text style={{ fontSize: 13 }}>
              {curConvId ? (selectedAgent?.agentName || t('chat_select_agent')) : t('chat_select_conversation_first')}
            </Text>
          </Space>
          <Dropdown menu={{ items: [
            { key: 'en', label: 'English', onClick: () => i18n.changeLanguage('en') },
            { key: 'zh-CN', label: '简体中文', onClick: () => i18n.changeLanguage('zh-CN') },
          ]}}>
            <Button type="text" size="small" icon={<GlobalOutlined />} />
          </Dropdown>
          <Dropdown menu={{ items: [
            { key: 'agents', label: t('agent_settings'), icon: <AppstoreOutlined />, onClick: () => navigate('/dashboard') },
            { type: 'divider' as const },
            { key: 'logout', label: t('logout'), icon: <LogoutOutlined />, onClick: () => { casdoorService.removeToken(); window.location.href = '/login'; } },
          ]}}>
            <Avatar size="small" icon={<UserOutlined />} style={{ background: token.colorPrimary, cursor: 'pointer' }} />
          </Dropdown>
        </Space>
      </Layout.Header>

      <Layout>
        {/* Sidebar — Conversations */}
        {!sidebarCollapsed && (
          <Layout.Sider width={280} theme="light" style={{
            borderRight: '1px solid #f0f0f0', height: 'calc(100vh - 56px)', overflow: 'hidden',
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
              {convLoading && <div style={{ textAlign: 'center', padding: 12 }}><Text type="secondary" style={{ fontSize: 12 }}>...</Text></div>}
            </div>
          </Layout.Sider>
        )}

        {/* Chat area */}
        <Layout.Content style={{ display: 'flex', flexDirection: 'column', height: 'calc(100vh - 56px)', background: '#fff' }}>
          <ConversationPane
            key={curConvId || 'empty'}
            conversationId={curConvId}
            selectedAgentId={selectedAgentId}
            selectedAgentName={selectedAgent?.agentName}
            provider={provider}
            authHeaders={authHeaders}
            getJwt={getJwt}
            t={t}
            bubbleRole={bubbleRole}
          />
        </Layout.Content>
      </Layout>
    </Layout>
  );
};

export default ChatPage;
