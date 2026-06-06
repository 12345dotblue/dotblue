import React from 'react';
import {
  Alert,
  Button,
  Card,
  Col,
  Descriptions,
  Empty,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Row,
  Select,
  Space,
  Switch,
  Table,
  Tag,
  Tabs,
  Typography,
  message,
} from 'antd';
import {
  ApiOutlined,
  LinkOutlined,
  PlusOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import axios from 'axios';
import { useTranslation } from 'react-i18next';
import { BACKEND_URL } from '../../config';
import { casdoorService } from '../identity/CasdoorService';

const { Paragraph, Text } = Typography;
const TELEGRAM_PLATFORM = 'telegram';
const QQ_PLATFORM = 'qq';
const MATRIX_PLATFORM = 'matrix';
const DISCORD_PLATFORM = 'discord';
const SLACK_PLATFORM = 'slack';
const FEISHU_PLATFORM = 'feishu';
const WEB_PLATFORM = 'web';

interface IMSettingsTabProps {
  createSignal?: number;
}

interface ConnectionItem {
  id: string;
  enterpriseId: string;
  platform: string;
  name: string;
  status: string;
  connectionMode: string;
  config?: Record<string, any>;
  callbackPath?: string;
  lastConnectedAt?: string;
  lastError?: string;
  createdBy?: string;
}

interface AgentItem {
  id: string;
  agentName: string;
}

interface BindingItem {
  id: string;
  agentId: string;
  connectionId: string;
  status: string;
  triggerMode: string;
  triggerConfig?: Record<string, any>;
  sessionStrategy: string;
  replyMode: string;
  allowGroup: boolean;
  allowDm: boolean;
  priority: number;
}

interface ConnectionEventItem {
  id: string;
  event_id: string;
  external_message_id: string;
  direction: string;
  status: string;
  error_message?: string;
  created_at?: string;
}

interface DeliveryLogItem {
  id: string;
  conversation_id?: string;
  message_id?: string;
  attempt: number;
  status: string;
  error_message?: string;
  created_at?: string;
}

function getAuthHeaders() {
  const token = casdoorService.getToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

function formatDateTime(value?: string) {
  if (!value) return '-';
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? '-' : date.toLocaleString();
}

function normalizeKeywordString(binding?: BindingItem) {
  const keywords = binding?.triggerConfig?.keywords;
  if (Array.isArray(keywords)) {
    return keywords.join(', ');
  }
  return '';
}

function formatErrorPreview(value?: string) {
  if (!value) return '-';
  return value.length > 48 ? `${value.slice(0, 48)}...` : value;
}

function isFixtureConnection(connection?: ConnectionItem | null) {
  if (!connection) return false;
  return connection.createdBy === 'integration-test' || connection.config?.appId === 'cli_integration';
}

function formatConnectionSource(connection?: ConnectionItem | null) {
  if (!connection?.createdBy) return '';
  return connection.createdBy;
}

function getResponseErrorText(error: any) {
  return typeof error?.response?.data === 'string' ? error.response.data.toLowerCase() : '';
}

const IMSettingsTab: React.FC<IMSettingsTabProps> = ({ createSignal = 0 }) => {
  const { t } = useTranslation();
  const [messageApi, contextHolder] = message.useMessage();
  const [loading, setLoading] = React.useState(true);
  const [agentsLoading, setAgentsLoading] = React.useState(false);
  const [connections, setConnections] = React.useState<ConnectionItem[]>([]);
  const [agents, setAgents] = React.useState<AgentItem[]>([]);
  const [selectedConnectionId, setSelectedConnectionId] = React.useState<string>();
  const [bindingsLoading, setBindingsLoading] = React.useState(false);
  const [bindings, setBindings] = React.useState<BindingItem[]>([]);
  const [eventsLoading, setEventsLoading] = React.useState(false);
  const [deliveriesLoading, setDeliveriesLoading] = React.useState(false);
  const [events, setEvents] = React.useState<ConnectionEventItem[]>([]);
  const [deliveries, setDeliveries] = React.useState<DeliveryLogItem[]>([]);
  const [eventFilters, setEventFilters] = React.useState({ direction: '', status: '' });
  const [deliveryFilters, setDeliveryFilters] = React.useState({ status: '' });
  const [detailState, setDetailState] = React.useState<{ title: string; payload: Record<string, any> } | null>(null);
  const [connectionModalOpen, setConnectionModalOpen] = React.useState(false);
  const [bindingModalOpen, setBindingModalOpen] = React.useState(false);
  const [editingConnection, setEditingConnection] = React.useState<ConnectionItem | null>(null);
  const [editingBinding, setEditingBinding] = React.useState<BindingItem | null>(null);
  const [savingConnection, setSavingConnection] = React.useState(false);
  const [savingBinding, setSavingBinding] = React.useState(false);
  const [connectionForm] = Form.useForm();
  const [bindingForm] = Form.useForm();
  const connectionFormPlatform = Form.useWatch('platform', connectionForm) || editingConnection?.platform || TELEGRAM_PLATFORM;

  const selectedConnection = React.useMemo(
    () => connections.find((item) => item.id === selectedConnectionId) || null,
    [connections, selectedConnectionId],
  );
  const selectedConnectionIsFixture = React.useMemo(
    () => isFixtureConnection(selectedConnection),
    [selectedConnection],
  );
  const getPlatformLabel = React.useCallback((value?: string) => {
    switch (value) {
      case TELEGRAM_PLATFORM:
        return t('im_settings_platform_telegram');
      case QQ_PLATFORM:
        return t('im_settings_platform_qq');
      case MATRIX_PLATFORM:
        return t('im_settings_platform_matrix');
      case DISCORD_PLATFORM:
        return t('im_settings_platform_discord');
      case SLACK_PLATFORM:
        return t('im_settings_platform_slack');
      case FEISHU_PLATFORM:
        return t('im_settings_platform_feishu');
      case WEB_PLATFORM:
        return t('im_settings_platform_web');
      default:
        return value || '-';
    }
  }, [t]);
  const getConnectionModeLabel = React.useCallback((value?: string) => {
    switch (value) {
      case 'direct':
        return t('im_settings_connection_mode_direct');
      case 'polling':
        return t('im_settings_connection_mode_polling');
      case 'sync':
        return t('im_settings_connection_mode_sync');
      case 'gateway':
        return t('im_settings_connection_mode_gateway');
      case 'socket_mode':
        return t('im_settings_connection_mode_socket_mode');
      default:
        return value || '-';
    }
  }, [t]);
  const getDirectionLabel = React.useCallback((value?: string) => {
    switch (value) {
      case 'inbound':
        return t('im_settings_inbound_label');
      case 'outbound':
        return t('im_settings_outbound_label');
      default:
        return value || '-';
    }
  }, [t]);
  const getStatusLabel = React.useCallback((value?: string) => {
    switch (value) {
      case 'received':
        return t('im_settings_received_status');
      case 'processed':
        return t('im_settings_processed_status');
      case 'no_binding':
        return t('im_settings_no_binding_status');
      case 'error':
        return t('im_settings_error_status');
      case 'pending':
        return t('im_settings_pending_status');
      case 'accepted':
        return t('im_settings_accepted_status');
      case 'failed':
        return t('im_settings_failed_status');
      case 'active':
        return t('im_settings_active_status');
      case 'disabled':
        return t('im_settings_disabled_status');
      default:
        return value || '-';
    }
  }, [t]);
  const getTriggerModeLabel = React.useCallback((value?: string) => {
    switch (value) {
      case 'mention_only':
        return t('im_settings_mention_only_trigger');
      case 'all_messages':
        return t('im_settings_all_messages_trigger');
      case 'keyword':
        return t('im_settings_keyword_trigger');
      case 'command':
        return t('im_settings_command_trigger');
      case 'dm_only':
        return t('im_settings_dm_only_trigger');
      case 'group_only':
        return t('im_settings_group_only_trigger');
      default:
        return value || '-';
    }
  }, [t]);
  const getSessionStrategyLabel = React.useCallback((value?: string) => {
    switch (value) {
      case 'per_user':
        return t('im_settings_per_user_session');
      case 'per_chat':
        return t('im_settings_per_chat_session');
      case 'per_thread':
        return t('im_settings_per_thread_session');
      case 'per_chat_per_user':
        return t('im_settings_per_chat_per_user_session');
      default:
        return value || '-';
    }
  }, [t]);
  const getSourceLabel = React.useCallback((connection?: ConnectionItem | null) => {
    const source = formatConnectionSource(connection);
    if (!source) {
      return t('im_settings_unknown_source');
    }
    if (source === 'integration-test') {
      return t('im_settings_fixture_tag');
    }
    return source;
  }, [t]);
  const connectionModeOptions = React.useMemo(
    () => {
      if (connectionFormPlatform === WEB_PLATFORM) {
        return [{ label: t('im_settings_connection_mode_direct'), value: 'direct' }];
      }
      if (connectionFormPlatform === TELEGRAM_PLATFORM) {
        return [{ label: t('im_settings_connection_mode_polling'), value: 'polling' }];
      }
      if (connectionFormPlatform === MATRIX_PLATFORM) {
        return [{ label: t('im_settings_connection_mode_sync'), value: 'sync' }];
      }
      if (connectionFormPlatform === QQ_PLATFORM) {
        return [{ label: t('im_settings_connection_mode_gateway'), value: 'gateway' }];
      }
      if (connectionFormPlatform === DISCORD_PLATFORM) {
        return [{ label: t('im_settings_connection_mode_gateway'), value: 'gateway' }];
      }
      return [{ label: t('im_settings_connection_mode_socket_mode'), value: 'socket_mode' }];
    },
    [connectionFormPlatform, t],
  );

  const loadConnections = React.useCallback(async () => {
    setLoading(true);
    try {
      const res = await axios.get(`${BACKEND_URL}/api/admin/im/connections`, {
        headers: getAuthHeaders(),
      });
      const items = Array.isArray(res.data) ? res.data : [];
      setConnections(items);
      setSelectedConnectionId((current) => {
        if (current && items.some((item: ConnectionItem) => item.id === current)) {
          return current;
        }
        return items[0]?.id;
      });
    } catch {
      messageApi.error(t('im_settings_load_connections_failed'));
      setConnections([]);
    } finally {
      setLoading(false);
    }
  }, [messageApi]);

  const loadAgents = React.useCallback(async () => {
    setAgentsLoading(true);
    try {
      const res = await axios.get(`${BACKEND_URL}/api/agents`, {
        headers: getAuthHeaders(),
      });
      setAgents(Array.isArray(res.data) ? res.data : []);
    } catch {
      setAgents([]);
    } finally {
      setAgentsLoading(false);
    }
  }, []);

  const loadBindings = React.useCallback(async (connectionID?: string) => {
    if (!connectionID) {
      setBindings([]);
      return;
    }
    setBindingsLoading(true);
    try {
      const res = await axios.get(`${BACKEND_URL}/api/admin/im/connections/${connectionID}/bindings`, {
        headers: getAuthHeaders(),
      });
      setBindings(Array.isArray(res.data) ? res.data : []);
    } catch {
      messageApi.error(t('im_settings_load_bindings_failed'));
      setBindings([]);
    } finally {
      setBindingsLoading(false);
    }
  }, [messageApi]);

  const loadEvents = React.useCallback(async (connectionID?: string) => {
    if (!connectionID) {
      setEvents([]);
      return;
    }
    setEventsLoading(true);
    try {
      const res = await axios.get(`${BACKEND_URL}/api/admin/im/connections/${connectionID}/events`, {
        headers: getAuthHeaders(),
        params: {
          limit: 20,
          ...(eventFilters.direction ? { direction: eventFilters.direction } : {}),
          ...(eventFilters.status ? { status: eventFilters.status } : {}),
        },
      });
      setEvents(Array.isArray(res.data?.items) ? res.data.items : []);
    } catch {
      messageApi.error(t('im_settings_load_events_failed'));
      setEvents([]);
    } finally {
      setEventsLoading(false);
    }
  }, [eventFilters.direction, eventFilters.status, messageApi]);

  const loadDeliveries = React.useCallback(async (connectionID?: string) => {
    if (!connectionID) {
      setDeliveries([]);
      return;
    }
    setDeliveriesLoading(true);
    try {
      const res = await axios.get(`${BACKEND_URL}/api/admin/im/connections/${connectionID}/deliveries`, {
        headers: getAuthHeaders(),
        params: {
          limit: 20,
          ...(deliveryFilters.status ? { status: deliveryFilters.status } : {}),
        },
      });
      setDeliveries(Array.isArray(res.data?.items) ? res.data.items : []);
    } catch {
      messageApi.error(t('im_settings_load_deliveries_failed'));
      setDeliveries([]);
    } finally {
      setDeliveriesLoading(false);
    }
  }, [deliveryFilters.status, messageApi]);

  React.useEffect(() => {
    loadConnections();
    loadAgents();
  }, [loadConnections, loadAgents]);

  React.useEffect(() => {
    loadBindings(selectedConnectionId);
    loadEvents(selectedConnectionId);
    loadDeliveries(selectedConnectionId);
  }, [selectedConnectionId, loadBindings, loadEvents, loadDeliveries]);

  React.useEffect(() => {
    if (createSignal > 0) {
      openCreateConnection();
    }
  }, [createSignal]);

  const openDetail = (title: string, payload: Record<string, any>) => {
    setDetailState({ title, payload });
  };

  const openCreateConnection = () => {
    setEditingConnection(null);
    connectionForm.setFieldsValue({
      platform: TELEGRAM_PLATFORM,
      name: '',
      connectionMode: 'polling',
      token: '',
      pollTimeoutSeconds: 30,
      qqAppId: '',
      qqAppSecret: '',
      matrixHomeserver: '',
      matrixUserId: '',
      matrixAccessToken: '',
      matrixSyncTimeoutSeconds: 30,
      discordToken: '',
      botToken: '',
      appToken: '',
      appId: '',
      appSecret: '',
      domain: 'feishu',
    });
    setConnectionModalOpen(true);
  };

  const openEditConnection = (connection: ConnectionItem) => {
    setEditingConnection(connection);
    connectionForm.setFieldsValue({
      platform: connection.platform,
      name: connection.name,
      connectionMode: connection.connectionMode || (
        connection.platform === WEB_PLATFORM
          ? 'direct'
          : connection.platform === TELEGRAM_PLATFORM
            ? 'polling'
            : connection.platform === QQ_PLATFORM
              ? 'gateway'
            : connection.platform === MATRIX_PLATFORM
              ? 'sync'
            : connection.platform === DISCORD_PLATFORM
              ? 'gateway'
            : 'socket_mode'
      ),
      token: '',
      pollTimeoutSeconds: connection.config?.pollTimeoutSeconds || 30,
      qqAppId: connection.config?.appId || '',
      qqAppSecret: '',
      matrixHomeserver: connection.config?.homeserver || '',
      matrixUserId: connection.config?.userId || '',
      matrixAccessToken: '',
      matrixSyncTimeoutSeconds: connection.config?.syncTimeoutSeconds || 30,
      discordToken: '',
      botToken: '',
      appToken: '',
      appId: connection.config?.appId || '',
      appSecret: '',
      domain: connection.config?.domain || 'feishu',
    });
    setConnectionModalOpen(true);
  };

  const handleSaveConnection = async () => {
    const values = await connectionForm.validateFields();
    const payload = values.platform === WEB_PLATFORM
      ? {
          platform: values.platform,
          name: values.name,
          connectionMode: values.connectionMode,
          config: {
            channel: 'web_chat',
          },
        }
      : values.platform === TELEGRAM_PLATFORM
        ? {
            platform: values.platform,
            name: values.name,
            connectionMode: values.connectionMode,
            config: {
              pollTimeoutSeconds: values.pollTimeoutSeconds,
            },
            ...(values.token
              ? { secrets: { token: values.token } }
              : editingConnection
                ? {}
                : { secrets: { token: values.token } }),
          }
        : values.platform === QQ_PLATFORM
          ? {
              platform: values.platform,
              name: values.name,
              connectionMode: values.connectionMode,
              config: {
                appId: values.qqAppId,
              },
              ...(values.qqAppSecret
                ? { secrets: { appSecret: values.qqAppSecret } }
                : editingConnection
                  ? {}
                  : { secrets: { appSecret: values.qqAppSecret } }),
            }
        : values.platform === MATRIX_PLATFORM
          ? {
              platform: values.platform,
              name: values.name,
              connectionMode: values.connectionMode,
              config: {
                homeserver: values.matrixHomeserver,
                userId: values.matrixUserId,
                syncTimeoutSeconds: values.matrixSyncTimeoutSeconds,
              },
              ...(values.matrixAccessToken
                ? { secrets: { accessToken: values.matrixAccessToken } }
                : editingConnection
                  ? {}
                  : { secrets: { accessToken: values.matrixAccessToken } }),
            }
        : values.platform === DISCORD_PLATFORM
          ? {
              platform: values.platform,
              name: values.name,
              connectionMode: values.connectionMode,
              config: {},
              ...(values.discordToken
                ? { secrets: { token: values.discordToken } }
                : editingConnection
                  ? {}
                  : { secrets: { token: values.discordToken } }),
            }
        : values.platform === SLACK_PLATFORM
          ? {
              platform: values.platform,
              name: values.name,
              connectionMode: values.connectionMode,
              config: {},
              ...((values.botToken || values.appToken)
                ? { secrets: { botToken: values.botToken, appToken: values.appToken } }
                : editingConnection
                  ? {}
                  : { secrets: { botToken: values.botToken, appToken: values.appToken } }),
            }
        : {
            platform: values.platform,
            name: values.name,
            connectionMode: values.connectionMode,
            config: {
              appId: values.appId,
              domain: values.domain,
            },
            ...(values.appSecret
              ? { secrets: { appSecret: values.appSecret } }
              : editingConnection
                ? {}
                : { secrets: { appSecret: values.appSecret } }),
          };

    setSavingConnection(true);
    try {
      if (editingConnection) {
        await axios.put(`${BACKEND_URL}/api/admin/im/connections/${editingConnection.id}`, payload, {
          headers: getAuthHeaders(),
        });
        messageApi.success(t('im_settings_connection_updated'));
      } else {
        await axios.post(`${BACKEND_URL}/api/admin/im/connections`, payload, {
          headers: getAuthHeaders(),
        });
        messageApi.success(t('im_settings_connection_created'));
      }
      setConnectionModalOpen(false);
      await loadConnections();
    } catch (error: any) {
      messageApi.error(t('im_settings_save_connection_failed'));
    } finally {
      setSavingConnection(false);
    }
  };

  const handleTestConnection = async (connection: ConnectionItem) => {
    try {
      await axios.post(`${BACKEND_URL}/api/admin/im/connections/${connection.id}/test`, {}, {
        headers: getAuthHeaders(),
      });
      messageApi.success(t('im_settings_connection_test_passed'));
    } catch (error: any) {
      messageApi.error(t('im_settings_connection_test_failed'));
    }
  };

  const handleToggleConnection = async (connection: ConnectionItem, enable: boolean) => {
    try {
      await axios.post(`${BACKEND_URL}/api/admin/im/connections/${connection.id}/${enable ? 'enable' : 'disable'}`, {}, {
        headers: getAuthHeaders(),
      });
      messageApi.success(enable ? t('im_settings_connection_enabled') : t('im_settings_connection_disabled'));
      await loadConnections();
    } catch (error: any) {
      const errorText = getResponseErrorText(error);
      if (enable && isFixtureConnection(connection) && errorText.includes('invalid param')) {
        messageApi.error(t('im_settings_fixture_enable_failed'));
        return;
      }
      messageApi.error(t('im_settings_connection_status_update_failed'));
    }
  };

  const openCreateBinding = () => {
    if (!selectedConnection) {
      messageApi.warning(t('im_settings_select_connection_warning'));
      return;
    }
    setEditingBinding(null);
    bindingForm.setFieldsValue({
      agentId: undefined,
      status: 'active',
      triggerMode: 'mention_only',
      keywords: '',
      sessionStrategy: 'per_chat_per_user',
      replyMode: 'default',
      allowGroup: true,
      allowDm: true,
      priority: 100,
    });
    setBindingModalOpen(true);
  };

  const openEditBinding = (binding: BindingItem) => {
    setEditingBinding(binding);
    bindingForm.setFieldsValue({
      agentId: binding.agentId,
      status: binding.status,
      triggerMode: binding.triggerMode,
      keywords: normalizeKeywordString(binding),
      sessionStrategy: binding.sessionStrategy,
      replyMode: binding.replyMode,
      allowGroup: binding.allowGroup,
      allowDm: binding.allowDm,
      priority: binding.priority,
    });
    setBindingModalOpen(true);
  };

  const handleSaveBinding = async () => {
    if (!selectedConnection) {
      return;
    }
    const values = await bindingForm.validateFields();
    const payload: Record<string, any> = {
      agentId: values.agentId,
      status: values.status,
      triggerMode: values.triggerMode,
      sessionStrategy: values.sessionStrategy,
      replyMode: values.replyMode,
      allowGroup: values.allowGroup,
      allowDm: values.allowDm,
      priority: values.priority,
    };
    if (values.triggerMode === 'keyword') {
      payload.triggerConfig = {
        keywords: String(values.keywords || '')
          .split(',')
          .map((item) => item.trim())
          .filter(Boolean),
      };
    } else if (!editingBinding) {
      payload.triggerConfig = {};
    }

    setSavingBinding(true);
    try {
      if (editingBinding) {
        await axios.put(`${BACKEND_URL}/api/admin/im/bindings/${editingBinding.id}`, payload, {
          headers: getAuthHeaders(),
        });
        messageApi.success(t('im_settings_binding_updated'));
      } else {
        await axios.post(`${BACKEND_URL}/api/admin/im/connections/${selectedConnection.id}/bindings`, payload, {
          headers: getAuthHeaders(),
        });
        messageApi.success(t('im_settings_binding_created'));
      }
      setBindingModalOpen(false);
      await loadBindings(selectedConnection.id);
    } catch (error: any) {
      messageApi.error(t('im_settings_save_binding_failed'));
    } finally {
      setSavingBinding(false);
    }
  };

  const handleDeleteBinding = async (binding: BindingItem) => {
    try {
      await axios.delete(`${BACKEND_URL}/api/admin/im/bindings/${binding.id}`, {
        headers: getAuthHeaders(),
      });
      messageApi.success(t('im_settings_binding_deleted'));
      await loadBindings(selectedConnection?.id);
    } catch (error: any) {
      messageApi.error(t('im_settings_binding_delete_failed'));
    }
  };

  const connectionColumns = [
    {
      title: t('im_settings_connection_column'),
      key: 'name',
      render: (_: unknown, item: ConnectionItem) => (
        <Space orientation="vertical" size={0}>
          <Space wrap size={[8, 0]}>
            <Text strong>{item.name}</Text>
            {isFixtureConnection(item) ? <Tag color="gold">{t('im_settings_fixture_tag')}</Tag> : null}
            {item.lastError ? <Tag color="error">{t('im_settings_recent_failure_tag')}</Tag> : null}
          </Space>
          <Text type="secondary">
            {getPlatformLabel(item.platform)} · {getConnectionModeLabel(item.connectionMode)} · {t('im_settings_source_prefix')}: {getSourceLabel(item)}
          </Text>
          {item.lastError ? (
            <Text type="danger">{formatErrorPreview(item.lastError)}</Text>
          ) : null}
        </Space>
      ),
    },
    {
      title: t('im_settings_status_column'),
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (value: string) => (
        <Tag color={value === 'active' ? 'success' : value === 'error' ? 'error' : 'default'}>
          {getStatusLabel(value)}
        </Tag>
      ),
    },
    {
      title: t('im_settings_callback_path_column'),
      dataIndex: 'callbackPath',
      key: 'callbackPath',
      render: (value?: string) => <Text code>{value || '-'}</Text>,
    },
    {
      title: t('im_settings_last_connected_column'),
      dataIndex: 'lastConnectedAt',
      key: 'lastConnectedAt',
      width: 180,
      render: (value?: string) => formatDateTime(value),
    },
    {
      title: t('im_settings_actions_column'),
      key: 'actions',
      width: 320,
      render: (_: unknown, item: ConnectionItem) => (
        <Space wrap>
          <Button size="small" onClick={() => openEditConnection(item)}>{t('im_settings_edit')}</Button>
          <Button size="small" onClick={() => handleTestConnection(item)}>{t('im_settings_test')}</Button>
          {item.status === 'active' ? (
            <Button size="small" onClick={() => handleToggleConnection(item, false)}>{t('im_settings_disable')}</Button>
          ) : (
            <Button size="small" type="primary" onClick={() => handleToggleConnection(item, true)}>{t('im_settings_enable')}</Button>
          )}
        </Space>
      ),
    },
  ];

  const bindingColumns = [
    {
      title: t('im_settings_agent_label'),
      key: 'agentId',
      render: (_: unknown, item: BindingItem) => {
        const agent = agents.find((candidate) => candidate.id === item.agentId);
        return (
          <Space orientation="vertical" size={0}>
            <Text strong>{agent?.agentName || item.agentId}</Text>
            <Text type="secondary">{item.agentId}</Text>
          </Space>
        );
      },
    },
    {
      title: t('im_settings_trigger_column'),
      key: 'triggerMode',
      render: (_: unknown, item: BindingItem) => (
        <Space orientation="vertical" size={0}>
          <Tag color="blue">{getTriggerModeLabel(item.triggerMode)}</Tag>
          {item.triggerMode === 'keyword' && (
            <Text type="secondary">
              {(item.triggerConfig?.keywords || []).join(', ') || '-'}
            </Text>
          )}
        </Space>
      ),
    },
    {
      title: t('im_settings_session_column'),
      dataIndex: 'sessionStrategy',
      key: 'sessionStrategy',
      width: 180,
      render: (value: string) => getSessionStrategyLabel(value),
    },
    {
      title: t('im_settings_priority_column'),
      dataIndex: 'priority',
      key: 'priority',
      width: 100,
    },
    {
      title: t('im_settings_status_column'),
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (value: string) => <Tag color={value === 'active' ? 'success' : 'default'}>{getStatusLabel(value)}</Tag>,
    },
    {
      title: t('im_settings_actions_column'),
      key: 'actions',
      width: 180,
      render: (_: unknown, item: BindingItem) => (
        <Space>
          <Button size="small" onClick={() => openEditBinding(item)}>{t('im_settings_edit')}</Button>
          <Popconfirm title={t('im_settings_confirm_delete_binding')} onConfirm={() => handleDeleteBinding(item)}>
            <Button size="small" danger>{t('im_settings_delete')}</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const eventColumns = [
    {
      title: t('im_settings_event_id_column'),
      dataIndex: 'event_id',
      key: 'event_id',
      render: (value: string) => <Text code>{value || '-'}</Text>,
    },
    {
      title: t('im_settings_direction_column'),
      dataIndex: 'direction',
      key: 'direction',
      width: 100,
      render: (value: string) => <Tag>{getDirectionLabel(value)}</Tag>,
    },
    {
      title: t('im_settings_status_column'),
      dataIndex: 'status',
      key: 'status',
      width: 110,
      render: (value: string) => (
        <Tag color={value === 'received' ? 'processing' : value === 'no_binding' ? 'warning' : value === 'error' ? 'error' : 'default'}>
          {getStatusLabel(value)}
        </Tag>
      ),
    },
    {
      title: t('im_settings_external_message_column'),
      dataIndex: 'external_message_id',
      key: 'external_message_id',
      render: (value?: string) => value ? <Text code>{value}</Text> : '-',
    },
    {
      title: t('im_settings_time_column'),
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (value?: string) => formatDateTime(value),
    },
    {
      title: t('im_settings_error_column'),
      dataIndex: 'error_message',
      key: 'error_message',
      render: (value?: string) => value ? <Text type="danger">{formatErrorPreview(value)}</Text> : '-',
    },
    {
      title: t('im_settings_detail_column'),
      key: 'actions',
      width: 100,
      render: (_: unknown, item: ConnectionEventItem) => (
        <Button size="small" onClick={() => openDetail(`${t('im_settings_event_detail_title')} · ${item.event_id || item.id}`, item)}>
          {t('im_settings_view')}
        </Button>
      ),
    },
  ];

  const deliveryColumns = [
    {
      title: t('im_settings_status_column'),
      dataIndex: 'status',
      key: 'status',
      width: 110,
      render: (value: string) => (
        <Tag color={value === 'accepted' ? 'success' : value === 'pending' ? 'processing' : value === 'failed' ? 'error' : 'default'}>
          {getStatusLabel(value)}
        </Tag>
      ),
    },
    {
      title: t('im_settings_attempts_column'),
      dataIndex: 'attempt',
      key: 'attempt',
      width: 80,
    },
    {
      title: t('im_settings_message_id_column'),
      dataIndex: 'message_id',
      key: 'message_id',
      render: (value?: string) => value ? <Text code>{value}</Text> : '-',
    },
    {
      title: t('im_settings_conversation_id_column'),
      dataIndex: 'conversation_id',
      key: 'conversation_id',
      render: (value?: string) => value ? <Text code>{value}</Text> : '-',
    },
    {
      title: t('im_settings_time_column'),
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (value?: string) => formatDateTime(value),
    },
    {
      title: t('im_settings_error_column'),
      dataIndex: 'error_message',
      key: 'error_message',
      render: (value?: string) => value ? <Text type="danger">{formatErrorPreview(value)}</Text> : '-',
    },
    {
      title: t('im_settings_detail_column'),
      key: 'actions',
      width: 100,
      render: (_: unknown, item: DeliveryLogItem) => (
        <Button size="small" onClick={() => openDetail(`${t('im_settings_delivery_detail_title')} · ${item.id}`, item)}>
          {t('im_settings_view')}
        </Button>
      ),
    },
  ];

  return (
    <div>
      {contextHolder}
      <Space orientation="vertical" size={16} style={{ width: '100%' }}>
        <Card variant="borderless" style={{ borderRadius: 20 }}>
          <Paragraph type="secondary" style={{ marginBottom: 0 }}>
            {t('im_settings_summary_description')}
          </Paragraph>
        </Card>

        <Row gutter={[16, 16]}>
          <Col xs={24} xl={14}>
            <Card
              variant="borderless"
              style={{ borderRadius: 20 }}
              title={<Space><ApiOutlined />{t('im_settings_connections_card_title')}</Space>}
              extra={
                <Space>
                  <Button icon={<ReloadOutlined />} onClick={loadConnections}>{t('im_settings_refresh')}</Button>
                  <Button type="primary" icon={<PlusOutlined />} onClick={openCreateConnection}>{t('im_settings_new_connection')}</Button>
                </Space>
              }
            >
              <Table
                rowKey="id"
                loading={loading}
                dataSource={connections}
                columns={connectionColumns}
                pagination={false}
                locale={{ emptyText: <Empty description={t('im_settings_no_connections')} image={Empty.PRESENTED_IMAGE_SIMPLE} /> }}
                rowClassName={(record) => record.id === selectedConnectionId ? 'im-selected-row' : ''}
                onRow={(record) => ({
                  onClick: () => setSelectedConnectionId(record.id),
                })}
                scroll={{ x: 960 }}
              />
            </Card>
          </Col>

          <Col xs={24} xl={10}>
            <Card
              variant="borderless"
              style={{ borderRadius: 20, minHeight: 420 }}
              title={<Space><LinkOutlined />{t('im_settings_bindings_card_title')}</Space>}
              extra={
                <Space>
                  <Button icon={<ReloadOutlined />} disabled={!selectedConnection} onClick={() => loadBindings(selectedConnection?.id)}>{t('im_settings_refresh')}</Button>
                  <Button type="primary" icon={<PlusOutlined />} disabled={!selectedConnection} onClick={openCreateBinding}>{t('im_settings_new_binding')}</Button>
                </Space>
              }
            >
              {selectedConnection ? (
                <Space orientation="vertical" size={16} style={{ width: '100%' }}>
                  <Card size="small" style={{ borderRadius: 16, background: '#fafcff' }}>
                    <Space orientation="vertical" size={4}>
                      <Space wrap size={[8, 0]}>
                        <Text strong>{selectedConnection.name}</Text>
                      {selectedConnectionIsFixture ? <Tag color="gold">{t('im_settings_fixture_connection_tag')}</Tag> : null}
                      {selectedConnection.status === 'error' ? <Tag color="error">{t('im_settings_enable_error_tag')}</Tag> : null}
                      </Space>
                      <Text type="secondary">{getPlatformLabel(selectedConnection.platform)} · {getConnectionModeLabel(selectedConnection.connectionMode)}</Text>
                    <Text type="secondary">{t('im_settings_source_prefix')}: {getSourceLabel(selectedConnection)}</Text>
                      <Text type="secondary">
                        {selectedConnection.platform === WEB_PLATFORM
                          ? `${t('im_settings_channel_label')}: ${
                              !selectedConnection.config?.channel || selectedConnection.config.channel === 'web_chat'
                                ? t('im_settings_channel_web_chat')
                                : selectedConnection.config.channel
                            }`
                          : selectedConnection.platform === TELEGRAM_PLATFORM
                            ? `${t('im_settings_poll_timeout_label')}: ${selectedConnection.config?.pollTimeoutSeconds || 30}s`
                            : selectedConnection.platform === QQ_PLATFORM
                              ? `${t('im_settings_app_id_label')}: ${selectedConnection.config?.appId || '-'}`
                            : selectedConnection.platform === MATRIX_PLATFORM
                                ? `${t('im_settings_homeserver_label')}: ${selectedConnection.config?.homeserver || '-'}`
                            : selectedConnection.platform === DISCORD_PLATFORM
                                ? t('im_settings_gateway_mode_label')
                            : selectedConnection.platform === SLACK_PLATFORM
                                  ? t('im_settings_socket_mode_label')
                                  : `${t('im_settings_app_id_label')}: ${selectedConnection.config?.appId || '-'}`}
                      </Text>
                      <Text type="secondary">{t('im_settings_callback_path_label')}: {selectedConnection.callbackPath || '-'}</Text>
                      {selectedConnection.lastError ? <Text type="danger">{selectedConnection.lastError}</Text> : null}
                    </Space>
                  </Card>
                  {selectedConnectionIsFixture ? (
                    <Alert
                      type="warning"
                      showIcon
                      message={t('im_settings_fixture_warning_title')}
                      description={t('im_settings_fixture_warning_description')}
                    />
                  ) : null}
                  <Tabs
                    items={[
                      {
                        key: 'bindings',
                        label: `${t('im_settings_bindings_tab_label')} (${bindings.length})`,
                        children: (
                          <Table
                            rowKey="id"
                            loading={bindingsLoading}
                            dataSource={bindings}
                            columns={bindingColumns}
                            pagination={false}
                            locale={{ emptyText: <Empty description={t('im_settings_no_bindings')} image={Empty.PRESENTED_IMAGE_SIMPLE} /> }}
                            scroll={{ x: 760 }}
                          />
                        ),
                      },
                      {
                        key: 'events',
                        label: `${t('im_settings_events_tab_label')} (${events.length})`,
                        children: (
                          <Space orientation="vertical" size={12} style={{ width: '100%' }}>
                            <Space wrap style={{ justifyContent: 'space-between', width: '100%' }}>
                              <Space wrap>
                                <Select
                                  allowClear
                                  placeholder={t('im_settings_direction_column')}
                                  style={{ width: 120 }}
                                  value={eventFilters.direction || undefined}
                                  options={[
                                    { label: t('im_settings_inbound_label'), value: 'inbound' },
                                    { label: t('im_settings_outbound_label'), value: 'outbound' },
                                  ]}
                                  onChange={(value) => setEventFilters((current) => ({ ...current, direction: value || '' }))}
                                />
                                <Select
                                  allowClear
                                  placeholder={t('im_settings_status_column')}
                                  style={{ width: 160 }}
                                  value={eventFilters.status || undefined}
                                  options={[
                                    { label: t('im_settings_received_status'), value: 'received' },
                                    { label: t('im_settings_processed_status'), value: 'processed' },
                                    { label: t('im_settings_no_binding_status'), value: 'no_binding' },
                                    { label: t('im_settings_error_status'), value: 'error' },
                                  ]}
                                  onChange={(value) => setEventFilters((current) => ({ ...current, status: value || '' }))}
                                />
                              </Space>
                              <Button icon={<ReloadOutlined />} onClick={() => loadEvents(selectedConnection.id)}>{t('im_settings_refresh_events')}</Button>
                            </Space>
                            <Table
                              rowKey="id"
                              loading={eventsLoading}
                              dataSource={events}
                              columns={eventColumns}
                              pagination={false}
                              locale={{ emptyText: <Empty description={t('im_settings_no_events')} image={Empty.PRESENTED_IMAGE_SIMPLE} /> }}
                              scroll={{ x: 960 }}
                            />
                          </Space>
                        ),
                      },
                      {
                        key: 'deliveries',
                        label: `${t('im_settings_deliveries_tab_label')} (${deliveries.length})`,
                        children: (
                          <Space orientation="vertical" size={12} style={{ width: '100%' }}>
                            <Space wrap style={{ justifyContent: 'space-between', width: '100%' }}>
                              <Select
                                allowClear
                                placeholder={t('im_settings_status_column')}
                                style={{ width: 160 }}
                                value={deliveryFilters.status || undefined}
                                options={[
                                  { label: t('im_settings_pending_status'), value: 'pending' },
                                  { label: t('im_settings_accepted_status'), value: 'accepted' },
                                  { label: t('im_settings_failed_status'), value: 'failed' },
                                ]}
                                onChange={(value) => setDeliveryFilters({ status: value || '' })}
                              />
                              <Button icon={<ReloadOutlined />} onClick={() => loadDeliveries(selectedConnection.id)}>{t('im_settings_refresh_deliveries')}</Button>
                            </Space>
                            <Table
                              rowKey="id"
                              loading={deliveriesLoading}
                              dataSource={deliveries}
                              columns={deliveryColumns}
                              pagination={false}
                              locale={{ emptyText: <Empty description={t('im_settings_no_deliveries')} image={Empty.PRESENTED_IMAGE_SIMPLE} /> }}
                              scroll={{ x: 960 }}
                            />
                          </Space>
                        ),
                      },
                    ]}
                  />
                </Space>
              ) : (
                <Empty description={t('im_settings_select_connection_first')} image={Empty.PRESENTED_IMAGE_SIMPLE} />
              )}
            </Card>
          </Col>
        </Row>
      </Space>

      <Modal
        title={editingConnection ? t('im_settings_edit_connection_title') : t('im_settings_create_connection_title')}
        open={connectionModalOpen}
        destroyOnHidden
        onCancel={() => setConnectionModalOpen(false)}
        onOk={handleSaveConnection}
        confirmLoading={savingConnection}
        okText={editingConnection ? t('im_settings_save') : t('im_settings_create')}
      >
        <Form form={connectionForm} layout="vertical">
          <Form.Item label={t('im_settings_platform_label')} name="platform" rules={[{ required: true, message: t('im_settings_select_platform') }]}>
            <Select
              disabled={Boolean(editingConnection)}
              options={[
                { label: t('im_settings_platform_telegram'), value: TELEGRAM_PLATFORM },
                { label: t('im_settings_platform_qq'), value: QQ_PLATFORM },
                { label: t('im_settings_platform_matrix'), value: MATRIX_PLATFORM },
                { label: t('im_settings_platform_discord'), value: DISCORD_PLATFORM },
                { label: t('im_settings_platform_slack'), value: SLACK_PLATFORM },
                { label: t('im_settings_platform_feishu'), value: FEISHU_PLATFORM },
                { label: t('im_settings_platform_web'), value: WEB_PLATFORM },
              ]}
            />
          </Form.Item>
          <Form.Item label={t('im_settings_connection_name_label')} name="name" rules={[{ required: true, message: t('im_settings_connection_name_required') }]}>
            <Input placeholder={t('im_settings_connection_name_placeholder')} />
          </Form.Item>
          <Form.Item label={t('im_settings_connection_mode_label')} name="connectionMode" rules={[{ required: true, message: t('im_settings_select_connection_mode') }]}>
            <Select options={connectionModeOptions} />
          </Form.Item>
          {connectionFormPlatform === WEB_PLATFORM ? (
            <Alert
              type="info"
              showIcon
              message={t('im_settings_web_chat_connection_title')}
              description={t('im_settings_web_chat_connection_description')}
            />
          ) : connectionFormPlatform === TELEGRAM_PLATFORM ? (
            <>
              <Alert
                type="info"
                showIcon
                message={t('im_settings_telegram_connection_title')}
                description={t('im_settings_telegram_connection_description')}
              />
              <Form.Item label={editingConnection ? t('im_settings_bot_token_optional_label') : t('im_settings_bot_token_label')} name="token" rules={editingConnection ? [] : [{ required: true, message: t('im_settings_bot_token_required') }]}>
                <Input.Password placeholder={t('im_settings_bot_token_example')} />
              </Form.Item>
              <Form.Item label={t('im_settings_poll_timeout_label')} name="pollTimeoutSeconds" rules={[{ required: true, message: t('im_settings_poll_timeout_label') }]}>
                <InputNumber controls={false} min={10} max={60} style={{ width: '100%' }} />
              </Form.Item>
            </>
          ) : connectionFormPlatform === QQ_PLATFORM ? (
            <>
              <Alert
                type="info"
                showIcon
                message={t('im_settings_qq_connection_title')}
                description={t('im_settings_qq_connection_description')}
              />
              <Form.Item label={t('im_settings_app_id_label')} name="qqAppId" rules={[{ required: true, message: t('im_settings_app_id_required') }]}>
                <Input placeholder={t('im_settings_qq_app_id_placeholder')} />
              </Form.Item>
              <Form.Item label={editingConnection ? t('im_settings_app_secret_optional_label') : t('im_settings_app_secret_label')} name="qqAppSecret" rules={editingConnection ? [] : [{ required: true, message: t('im_settings_app_secret_required') }]}>
                <Input.Password placeholder={t('im_settings_qq_app_secret_placeholder')} />
              </Form.Item>
            </>
          ) : connectionFormPlatform === MATRIX_PLATFORM ? (
            <>
              <Alert
                type="info"
                showIcon
                message={t('im_settings_matrix_connection_title')}
                description={t('im_settings_matrix_connection_description')}
              />
              <Form.Item label={t('im_settings_homeserver_label')} name="matrixHomeserver" rules={[{ required: true, message: t('im_settings_homeserver_required') }]}>
                <Input placeholder={t('im_settings_homeserver_placeholder')} />
              </Form.Item>
              <Form.Item label={t('im_settings_user_id_label')} name="matrixUserId" rules={[{ required: true, message: t('im_settings_user_id_required') }]}>
                <Input placeholder={t('im_settings_user_id_placeholder')} />
              </Form.Item>
              <Form.Item label={editingConnection ? t('im_settings_access_token_optional_label') : t('im_settings_access_token_label')} name="matrixAccessToken" rules={editingConnection ? [] : [{ required: true, message: t('im_settings_access_token_required') }]}>
                <Input.Password placeholder={t('im_settings_access_token_label')} />
              </Form.Item>
              <Form.Item label={t('im_settings_sync_timeout_label')} name="matrixSyncTimeoutSeconds" rules={[{ required: true, message: t('im_settings_sync_timeout_required') }]}>
                <InputNumber controls={false} min={10} max={60} style={{ width: '100%' }} />
              </Form.Item>
            </>
          ) : connectionFormPlatform === DISCORD_PLATFORM ? (
            <>
              <Alert
                type="info"
                showIcon
                message={t('im_settings_discord_connection_title')}
                description={t('im_settings_discord_connection_description')}
              />
              <Form.Item label={editingConnection ? t('im_settings_bot_token_optional_label') : t('im_settings_bot_token_label')} name="discordToken" rules={editingConnection ? [] : [{ required: true, message: t('im_settings_bot_token_required') }]}>
                <Input.Password placeholder={t('im_settings_bot_token_label')} />
              </Form.Item>
            </>
          ) : connectionFormPlatform === SLACK_PLATFORM ? (
            <>
              <Alert
                type="info"
                showIcon
                message={t('im_settings_slack_connection_title')}
                description={t('im_settings_slack_connection_description')}
              />
              <Form.Item label={editingConnection ? t('im_settings_bot_token_optional_label') : t('im_settings_bot_token_label')} name="botToken" rules={editingConnection ? [] : [{ required: true, message: t('im_settings_bot_token_required') }]}>
                <Input.Password placeholder={t('im_settings_slack_bot_token_placeholder')} />
              </Form.Item>
              <Form.Item label={editingConnection ? t('im_settings_app_token_optional_label') : t('im_settings_app_token_label')} name="appToken" rules={editingConnection ? [] : [{ required: true, message: t('im_settings_app_token_required') }]}>
                <Input.Password placeholder={t('im_settings_slack_app_token_placeholder')} />
              </Form.Item>
            </>
          ) : (
            <>
              <Form.Item label={t('im_settings_app_id_label')} name="appId" rules={[{ required: true, message: t('im_settings_app_id_required') }]}>
                <Input />
              </Form.Item>
              <Form.Item label={editingConnection ? t('im_settings_app_secret_optional_label') : t('im_settings_app_secret_label')} name="appSecret" rules={editingConnection ? [] : [{ required: true, message: t('im_settings_app_secret_required') }]}>
                <Input.Password />
              </Form.Item>
              <Form.Item label={t('im_settings_domain_label')} name="domain">
                <Select options={[{ label: t('im_settings_domain_feishu'), value: 'feishu' }, { label: t('im_settings_domain_lark'), value: 'lark' }]} />
              </Form.Item>
            </>
          )}
        </Form>
      </Modal>

      <Modal
        title={editingBinding ? t('im_settings_edit_binding_title') : t('im_settings_create_binding_title')}
        open={bindingModalOpen}
        destroyOnHidden
        onCancel={() => setBindingModalOpen(false)}
        onOk={handleSaveBinding}
        confirmLoading={savingBinding}
        okText={editingBinding ? t('im_settings_save') : t('im_settings_create')}
      >
        <Form form={bindingForm} layout="vertical" initialValues={{ allowGroup: true, allowDm: true }}>
          <Form.Item label={t('im_settings_agent_label')} name="agentId" rules={[{ required: true, message: t('im_settings_select_agent') }]}>
            <Select
              loading={agentsLoading}
              showSearch
              optionFilterProp="label"
              options={agents.map((item) => ({ label: item.agentName, value: item.id }))}
            />
          </Form.Item>
          <Form.Item label={t('im_settings_status_column')} name="status" rules={[{ required: true, message: t('im_settings_select_status') }]}>
            <Select options={[{ label: t('im_settings_active_status'), value: 'active' }, { label: t('im_settings_disabled_status'), value: 'disabled' }]} />
          </Form.Item>
          <Form.Item label={t('im_settings_trigger_column')} name="triggerMode" rules={[{ required: true, message: t('im_settings_select_trigger_mode') }]}>
            <Select
              options={[
                { label: t('im_settings_mention_only_trigger'), value: 'mention_only' },
                { label: t('im_settings_all_messages_trigger'), value: 'all_messages' },
                { label: t('im_settings_keyword_trigger'), value: 'keyword' },
                { label: t('im_settings_command_trigger'), value: 'command' },
                { label: t('im_settings_dm_only_trigger'), value: 'dm_only' },
                { label: t('im_settings_group_only_trigger'), value: 'group_only' },
              ]}
            />
          </Form.Item>
          <Form.Item noStyle shouldUpdate={(prev, current) => prev.triggerMode !== current.triggerMode}>
            {({ getFieldValue }) => getFieldValue('triggerMode') === 'keyword' ? (
              <Form.Item
                label={t('im_settings_keyword_label')}
                name="keywords"
                rules={[{ required: true, message: t('im_settings_keyword_required') }]}
              >
                <Input placeholder={t('im_settings_keyword_placeholder')} />
              </Form.Item>
            ) : null}
          </Form.Item>
          <Form.Item label={t('im_settings_session_strategy_label')} name="sessionStrategy" rules={[{ required: true, message: t('im_settings_select_session_strategy') }]}>
            <Select
              options={[
                { label: t('im_settings_per_user_session'), value: 'per_user' },
                { label: t('im_settings_per_chat_session'), value: 'per_chat' },
                { label: t('im_settings_per_thread_session'), value: 'per_thread' },
                { label: t('im_settings_per_chat_per_user_session'), value: 'per_chat_per_user' },
              ]}
            />
          </Form.Item>
          <Form.Item label={t('im_settings_reply_mode_label')} name="replyMode" rules={[{ required: true, message: t('im_settings_reply_mode_required') }]}>
            <Input />
          </Form.Item>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item label={t('im_settings_allow_group')} name="allowGroup" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label={t('im_settings_allow_dm')} name="allowDm" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item label={t('im_settings_priority_column')} name="priority" rules={[{ required: true, message: t('im_settings_priority_required') }]}>
            <InputNumber controls={false} min={0} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={detailState?.title || t('im_settings_detail_dialog_title')}
        open={Boolean(detailState)}
        footer={null}
        width={720}
        onCancel={() => setDetailState(null)}
      >
        {detailState ? (
          <Space orientation="vertical" size={16} style={{ width: '100%' }}>
            <Descriptions bordered size="small" column={1}>
              {Object.entries(detailState.payload).map(([key, value]) => (
                <Descriptions.Item key={key} label={key}>
                  {typeof value === 'object' && value !== null ? (
                    <pre style={{ margin: 0, whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
                      {JSON.stringify(value, null, 2)}
                    </pre>
                  ) : (
                    String(value ?? '-')
                  )}
                </Descriptions.Item>
              ))}
            </Descriptions>
          </Space>
        ) : null}
      </Modal>

      <style>
        {`
          .im-selected-row > td {
            background: #f0f7ff !important;
          }
        `}
      </style>
    </div>
  );
};

export default IMSettingsTab;

