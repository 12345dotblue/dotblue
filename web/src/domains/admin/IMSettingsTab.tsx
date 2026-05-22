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
import { BACKEND_URL } from '../../config';
import { casdoorService } from '../identity/CasdoorService';

const { Paragraph, Text } = Typography;
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
  if (!connection?.createdBy) return '未知来源';
  return connection.createdBy === 'integration-test' ? 'integration-test' : connection.createdBy;
}

const IMSettingsTab: React.FC<IMSettingsTabProps> = ({ createSignal = 0 }) => {
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
  const connectionFormPlatform = Form.useWatch('platform', connectionForm) || editingConnection?.platform || FEISHU_PLATFORM;

  const selectedConnection = React.useMemo(
    () => connections.find((item) => item.id === selectedConnectionId) || null,
    [connections, selectedConnectionId],
  );
  const selectedConnectionIsFixture = React.useMemo(
    () => isFixtureConnection(selectedConnection),
    [selectedConnection],
  );
  const connectionModeOptions = React.useMemo(
    () => connectionFormPlatform === WEB_PLATFORM
      ? [{ label: 'direct', value: 'direct' }]
      : [{ label: 'socket_mode', value: 'socket_mode' }],
    [connectionFormPlatform],
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
      messageApi.error('加载 IM 连接失败');
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
      messageApi.error('加载绑定失败');
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
      messageApi.error('加载事件日志失败');
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
      messageApi.error('加载投递日志失败');
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
      platform: FEISHU_PLATFORM,
      name: '',
      connectionMode: 'socket_mode',
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
      connectionMode: connection.connectionMode || (connection.platform === WEB_PLATFORM ? 'direct' : 'socket_mode'),
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
        messageApi.success('连接已更新');
      } else {
        await axios.post(`${BACKEND_URL}/api/admin/im/connections`, payload, {
          headers: getAuthHeaders(),
        });
        messageApi.success('连接已创建');
      }
      setConnectionModalOpen(false);
      await loadConnections();
    } catch (error: any) {
      messageApi.error(typeof error?.response?.data === 'string' ? error.response.data : '保存连接失败');
    } finally {
      setSavingConnection(false);
    }
  };

  const handleTestConnection = async (connection: ConnectionItem) => {
    try {
      const res = await axios.post(`${BACKEND_URL}/api/admin/im/connections/${connection.id}/test`, {}, {
        headers: getAuthHeaders(),
      });
      messageApi.success(res.data?.detail || '连接测试通过');
    } catch (error: any) {
      messageApi.error(typeof error?.response?.data === 'string' ? error.response.data : '连接测试失败');
    }
  };

  const handleToggleConnection = async (connection: ConnectionItem, enable: boolean) => {
    try {
      await axios.post(`${BACKEND_URL}/api/admin/im/connections/${connection.id}/${enable ? 'enable' : 'disable'}`, {}, {
        headers: getAuthHeaders(),
      });
      messageApi.success(enable ? '连接已启用' : '连接已停用');
      await loadConnections();
    } catch (error: any) {
      const errorMessage = typeof error?.response?.data === 'string' ? error.response.data : '更新连接状态失败';
      if (enable && isFixtureConnection(connection) && errorMessage.includes('invalid param')) {
        messageApi.error('当前连接是集成测试夹具数据，使用了无效的飞书 App ID / Secret，无法真实启用。请改成真实飞书应用配置后再启用。');
        return;
      }
      messageApi.error(errorMessage);
    }
  };

  const openCreateBinding = () => {
    if (!selectedConnection) {
      messageApi.warning('请先选择一个连接');
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
        messageApi.success('绑定已更新');
      } else {
        await axios.post(`${BACKEND_URL}/api/admin/im/connections/${selectedConnection.id}/bindings`, payload, {
          headers: getAuthHeaders(),
        });
        messageApi.success('绑定已创建');
      }
      setBindingModalOpen(false);
      await loadBindings(selectedConnection.id);
    } catch (error: any) {
      messageApi.error(typeof error?.response?.data === 'string' ? error.response.data : '保存绑定失败');
    } finally {
      setSavingBinding(false);
    }
  };

  const handleDeleteBinding = async (binding: BindingItem) => {
    try {
      await axios.delete(`${BACKEND_URL}/api/admin/im/bindings/${binding.id}`, {
        headers: getAuthHeaders(),
      });
      messageApi.success('绑定已删除');
      await loadBindings(selectedConnection?.id);
    } catch (error: any) {
      messageApi.error(typeof error?.response?.data === 'string' ? error.response.data : '删除绑定失败');
    }
  };

  const connectionColumns = [
    {
      title: '连接',
      key: 'name',
      render: (_: unknown, item: ConnectionItem) => (
        <Space orientation="vertical" size={0}>
          <Space wrap size={[8, 0]}>
            <Text strong>{item.name}</Text>
            {isFixtureConnection(item) ? <Tag color="gold">测试夹具</Tag> : null}
            {item.lastError ? <Tag color="error">最近失败</Tag> : null}
          </Space>
          <Text type="secondary">
            {item.platform} · {item.connectionMode} · 来源：{formatConnectionSource(item)}
          </Text>
          {item.lastError ? (
            <Text type="danger">{formatErrorPreview(item.lastError)}</Text>
          ) : null}
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (value: string) => (
        <Tag color={value === 'active' ? 'success' : value === 'error' ? 'error' : 'default'}>
          {value}
        </Tag>
      ),
    },
    {
      title: '回调路径',
      dataIndex: 'callbackPath',
      key: 'callbackPath',
      render: (value?: string) => <Text code>{value || '-'}</Text>,
    },
    {
      title: '最近连接',
      dataIndex: 'lastConnectedAt',
      key: 'lastConnectedAt',
      width: 180,
      render: (value?: string) => formatDateTime(value),
    },
    {
      title: '操作',
      key: 'actions',
      width: 320,
      render: (_: unknown, item: ConnectionItem) => (
        <Space wrap>
          <Button size="small" onClick={() => openEditConnection(item)}>编辑</Button>
          <Button size="small" onClick={() => handleTestConnection(item)}>测试</Button>
          {item.status === 'active' ? (
            <Button size="small" onClick={() => handleToggleConnection(item, false)}>停用</Button>
          ) : (
            <Button size="small" type="primary" onClick={() => handleToggleConnection(item, true)}>启用</Button>
          )}
        </Space>
      ),
    },
  ];

  const bindingColumns = [
    {
      title: 'Agent',
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
      title: '触发',
      key: 'triggerMode',
      render: (_: unknown, item: BindingItem) => (
        <Space orientation="vertical" size={0}>
          <Tag color="blue">{item.triggerMode}</Tag>
          {item.triggerMode === 'keyword' && (
            <Text type="secondary">
              {(item.triggerConfig?.keywords || []).join(', ') || '-'}
            </Text>
          )}
        </Space>
      ),
    },
    {
      title: '会话',
      dataIndex: 'sessionStrategy',
      key: 'sessionStrategy',
      width: 180,
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
      width: 100,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (value: string) => <Tag color={value === 'active' ? 'success' : 'default'}>{value}</Tag>,
    },
    {
      title: '操作',
      key: 'actions',
      width: 180,
      render: (_: unknown, item: BindingItem) => (
        <Space>
          <Button size="small" onClick={() => openEditBinding(item)}>编辑</Button>
          <Popconfirm title="确认删除这个绑定？" onConfirm={() => handleDeleteBinding(item)}>
            <Button size="small" danger>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const eventColumns = [
    {
      title: '事件 ID',
      dataIndex: 'event_id',
      key: 'event_id',
      render: (value: string) => <Text code>{value || '-'}</Text>,
    },
    {
      title: '方向',
      dataIndex: 'direction',
      key: 'direction',
      width: 100,
      render: (value: string) => <Tag>{value || '-'}</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 110,
      render: (value: string) => (
        <Tag color={value === 'received' ? 'processing' : value === 'no_binding' ? 'warning' : value === 'error' ? 'error' : 'default'}>
          {value}
        </Tag>
      ),
    },
    {
      title: '外部消息',
      dataIndex: 'external_message_id',
      key: 'external_message_id',
      render: (value?: string) => value ? <Text code>{value}</Text> : '-',
    },
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (value?: string) => formatDateTime(value),
    },
    {
      title: '错误',
      dataIndex: 'error_message',
      key: 'error_message',
      render: (value?: string) => value ? <Text type="danger">{formatErrorPreview(value)}</Text> : '-',
    },
    {
      title: '详情',
      key: 'actions',
      width: 100,
      render: (_: unknown, item: ConnectionEventItem) => (
        <Button size="small" onClick={() => openDetail(`事件详情 · ${item.event_id || item.id}`, item)}>
          查看
        </Button>
      ),
    },
  ];

  const deliveryColumns = [
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 110,
      render: (value: string) => (
        <Tag color={value === 'accepted' ? 'success' : value === 'pending' ? 'processing' : value === 'failed' ? 'error' : 'default'}>
          {value}
        </Tag>
      ),
    },
    {
      title: '尝试',
      dataIndex: 'attempt',
      key: 'attempt',
      width: 80,
    },
    {
      title: '消息 ID',
      dataIndex: 'message_id',
      key: 'message_id',
      render: (value?: string) => value ? <Text code>{value}</Text> : '-',
    },
    {
      title: '会话 ID',
      dataIndex: 'conversation_id',
      key: 'conversation_id',
      render: (value?: string) => value ? <Text code>{value}</Text> : '-',
    },
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (value?: string) => formatDateTime(value),
    },
    {
      title: '错误',
      dataIndex: 'error_message',
      key: 'error_message',
      render: (value?: string) => value ? <Text type="danger">{formatErrorPreview(value)}</Text> : '-',
    },
    {
      title: '详情',
      key: 'actions',
      width: 100,
      render: (_: unknown, item: DeliveryLogItem) => (
        <Button size="small" onClick={() => openDetail(`投递详情 · ${item.id}`, item)}>
          查看
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
            管理企业级 IM 渠道连接，并为每个连接配置 Agent 绑定、触发规则和会话策略。
          </Paragraph>
        </Card>

        <Row gutter={[16, 16]}>
          <Col xs={24} xl={14}>
            <Card
              variant="borderless"
              style={{ borderRadius: 20 }}
              title={<Space><ApiOutlined />IM 连接</Space>}
              extra={
                <Space>
                  <Button icon={<ReloadOutlined />} onClick={loadConnections}>刷新</Button>
                  <Button type="primary" icon={<PlusOutlined />} onClick={openCreateConnection}>新建连接</Button>
                </Space>
              }
            >
              <Table
                rowKey="id"
                loading={loading}
                dataSource={connections}
                columns={connectionColumns}
                pagination={false}
                locale={{ emptyText: <Empty description="暂无 IM 连接" image={Empty.PRESENTED_IMAGE_SIMPLE} /> }}
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
              title={<Space><LinkOutlined />Agent 绑定</Space>}
              extra={
                <Space>
                  <Button icon={<ReloadOutlined />} disabled={!selectedConnection} onClick={() => loadBindings(selectedConnection?.id)}>刷新</Button>
                  <Button type="primary" icon={<PlusOutlined />} disabled={!selectedConnection} onClick={openCreateBinding}>新增绑定</Button>
                </Space>
              }
            >
              {selectedConnection ? (
                <Space orientation="vertical" size={16} style={{ width: '100%' }}>
                  <Card size="small" style={{ borderRadius: 16, background: '#fafcff' }}>
                    <Space orientation="vertical" size={4}>
                      <Space wrap size={[8, 0]}>
                        <Text strong>{selectedConnection.name}</Text>
                        {selectedConnectionIsFixture ? <Tag color="gold">测试夹具连接</Tag> : null}
                        {selectedConnection.status === 'error' ? <Tag color="error">启用异常</Tag> : null}
                      </Space>
                      <Text type="secondary">{selectedConnection.platform} · {selectedConnection.connectionMode}</Text>
                      <Text type="secondary">来源：{formatConnectionSource(selectedConnection)}</Text>
                      <Text type="secondary">
                        {selectedConnection.platform === WEB_PLATFORM
                          ? `通道：${selectedConnection.config?.channel || 'web_chat'}`
                          : `App ID：${selectedConnection.config?.appId || '-'}`}
                      </Text>
                      <Text type="secondary">回调路径：{selectedConnection.callbackPath || '-'}</Text>
                      {selectedConnection.lastError ? <Text type="danger">{selectedConnection.lastError}</Text> : null}
                    </Space>
                  </Card>
                  {selectedConnectionIsFixture ? (
                    <Alert
                      type="warning"
                      showIcon
                      message="当前连接来自集成测试夹具"
                      description="这条连接使用测试用 App ID / Secret，测试连接或启用时会命中真实飞书 API 的 invalid param。若要验证真实启用链路，请新建或编辑为真实飞书应用配置。"
                    />
                  ) : null}
                  <Tabs
                    items={[
                      {
                        key: 'bindings',
                        label: `绑定 (${bindings.length})`,
                        children: (
                          <Table
                            rowKey="id"
                            loading={bindingsLoading}
                            dataSource={bindings}
                            columns={bindingColumns}
                            pagination={false}
                            locale={{ emptyText: <Empty description="当前连接暂无绑定" image={Empty.PRESENTED_IMAGE_SIMPLE} /> }}
                            scroll={{ x: 760 }}
                          />
                        ),
                      },
                      {
                        key: 'events',
                        label: `事件 (${events.length})`,
                        children: (
                          <Space orientation="vertical" size={12} style={{ width: '100%' }}>
                            <Space wrap style={{ justifyContent: 'space-between', width: '100%' }}>
                              <Space wrap>
                                <Select
                                  allowClear
                                  placeholder="方向"
                                  style={{ width: 120 }}
                                  value={eventFilters.direction || undefined}
                                  options={[
                                    { label: 'inbound', value: 'inbound' },
                                    { label: 'outbound', value: 'outbound' },
                                  ]}
                                  onChange={(value) => setEventFilters((current) => ({ ...current, direction: value || '' }))}
                                />
                                <Select
                                  allowClear
                                  placeholder="状态"
                                  style={{ width: 160 }}
                                  value={eventFilters.status || undefined}
                                  options={[
                                    { label: 'received', value: 'received' },
                                    { label: 'processed', value: 'processed' },
                                    { label: 'no_binding', value: 'no_binding' },
                                    { label: 'error', value: 'error' },
                                  ]}
                                  onChange={(value) => setEventFilters((current) => ({ ...current, status: value || '' }))}
                                />
                              </Space>
                              <Button icon={<ReloadOutlined />} onClick={() => loadEvents(selectedConnection.id)}>刷新事件</Button>
                            </Space>
                            <Table
                              rowKey="id"
                              loading={eventsLoading}
                              dataSource={events}
                              columns={eventColumns}
                              pagination={false}
                              locale={{ emptyText: <Empty description="当前连接暂无事件日志" image={Empty.PRESENTED_IMAGE_SIMPLE} /> }}
                              scroll={{ x: 960 }}
                            />
                          </Space>
                        ),
                      },
                      {
                        key: 'deliveries',
                        label: `投递 (${deliveries.length})`,
                        children: (
                          <Space orientation="vertical" size={12} style={{ width: '100%' }}>
                            <Space wrap style={{ justifyContent: 'space-between', width: '100%' }}>
                              <Select
                                allowClear
                                placeholder="状态"
                                style={{ width: 160 }}
                                value={deliveryFilters.status || undefined}
                                options={[
                                  { label: 'pending', value: 'pending' },
                                  { label: 'accepted', value: 'accepted' },
                                  { label: 'failed', value: 'failed' },
                                ]}
                                onChange={(value) => setDeliveryFilters({ status: value || '' })}
                              />
                              <Button icon={<ReloadOutlined />} onClick={() => loadDeliveries(selectedConnection.id)}>刷新投递</Button>
                            </Space>
                            <Table
                              rowKey="id"
                              loading={deliveriesLoading}
                              dataSource={deliveries}
                              columns={deliveryColumns}
                              pagination={false}
                              locale={{ emptyText: <Empty description="当前连接暂无投递日志" image={Empty.PRESENTED_IMAGE_SIMPLE} /> }}
                              scroll={{ x: 960 }}
                            />
                          </Space>
                        ),
                      },
                    ]}
                  />
                </Space>
              ) : (
                <Empty description="请先选择一个 IM 连接" image={Empty.PRESENTED_IMAGE_SIMPLE} />
              )}
            </Card>
          </Col>
        </Row>
      </Space>

      <Modal
        title={editingConnection ? '编辑 IM 连接' : '新建 IM 连接'}
        open={connectionModalOpen}
        destroyOnHidden
        onCancel={() => setConnectionModalOpen(false)}
        onOk={handleSaveConnection}
        confirmLoading={savingConnection}
        okText={editingConnection ? '保存' : '创建'}
      >
        <Form form={connectionForm} layout="vertical">
          <Form.Item label="平台" name="platform" rules={[{ required: true, message: '请选择平台' }]}>
            <Select
              disabled={Boolean(editingConnection)}
              options={[
                { label: 'Feishu', value: FEISHU_PLATFORM },
                { label: 'Web Chat', value: WEB_PLATFORM },
              ]}
            />
          </Form.Item>
          <Form.Item label="连接名称" name="name" rules={[{ required: true, message: '请输入连接名称' }]}>
            <Input placeholder="例如：企业飞书主通道" />
          </Form.Item>
          <Form.Item label="连接模式" name="connectionMode" rules={[{ required: true, message: '请选择连接模式' }]}>
            <Select options={connectionModeOptions} />
          </Form.Item>
          {connectionFormPlatform === WEB_PLATFORM ? (
            <Alert
              type="info"
              showIcon
              message="Web Chat 连接"
              description="Web Chat 连接不需要第三方 App ID / Secret。系统会把网页聊天请求接入 IM 框架，并通过该连接记录绑定、事件和投递。"
            />
          ) : (
            <>
              <Form.Item label="App ID" name="appId" rules={[{ required: true, message: '请输入 App ID' }]}>
                <Input />
              </Form.Item>
              <Form.Item label={editingConnection ? 'App Secret（留空表示保持不变）' : 'App Secret'} name="appSecret" rules={editingConnection ? [] : [{ required: true, message: '请输入 App Secret' }]}>
                <Input.Password />
              </Form.Item>
              <Form.Item label="域名" name="domain">
                <Select options={[{ label: 'feishu', value: 'feishu' }, { label: 'lark', value: 'lark' }]} />
              </Form.Item>
            </>
          )}
        </Form>
      </Modal>

      <Modal
        title={editingBinding ? '编辑绑定' : '新增绑定'}
        open={bindingModalOpen}
        destroyOnHidden
        onCancel={() => setBindingModalOpen(false)}
        onOk={handleSaveBinding}
        confirmLoading={savingBinding}
        okText={editingBinding ? '保存' : '创建'}
      >
        <Form form={bindingForm} layout="vertical" initialValues={{ allowGroup: true, allowDm: true }}>
          <Form.Item label="Agent" name="agentId" rules={[{ required: true, message: '请选择 Agent' }]}>
            <Select
              loading={agentsLoading}
              showSearch
              optionFilterProp="label"
              options={agents.map((item) => ({ label: item.agentName, value: item.id }))}
            />
          </Form.Item>
          <Form.Item label="状态" name="status" rules={[{ required: true, message: '请选择状态' }]}>
            <Select options={[{ label: 'active', value: 'active' }, { label: 'disabled', value: 'disabled' }]} />
          </Form.Item>
          <Form.Item label="触发模式" name="triggerMode" rules={[{ required: true, message: '请选择触发模式' }]}>
            <Select
              options={[
                { label: 'mention_only', value: 'mention_only' },
                { label: 'all_messages', value: 'all_messages' },
                { label: 'keyword', value: 'keyword' },
                { label: 'command', value: 'command' },
                { label: 'dm_only', value: 'dm_only' },
                { label: 'group_only', value: 'group_only' },
              ]}
            />
          </Form.Item>
          <Form.Item noStyle shouldUpdate={(prev, current) => prev.triggerMode !== current.triggerMode}>
            {({ getFieldValue }) => getFieldValue('triggerMode') === 'keyword' ? (
              <Form.Item
                label="关键词"
                name="keywords"
                rules={[{ required: true, message: '请输入关键词，多个关键词用逗号分隔' }]}
              >
                <Input placeholder="例如：报销,审批,日报" />
              </Form.Item>
            ) : null}
          </Form.Item>
          <Form.Item label="会话策略" name="sessionStrategy" rules={[{ required: true, message: '请选择会话策略' }]}>
            <Select
              options={[
                { label: 'per_user', value: 'per_user' },
                { label: 'per_chat', value: 'per_chat' },
                { label: 'per_thread', value: 'per_thread' },
                { label: 'per_chat_per_user', value: 'per_chat_per_user' },
              ]}
            />
          </Form.Item>
          <Form.Item label="回复模式" name="replyMode" rules={[{ required: true, message: '请输入回复模式' }]}>
            <Input />
          </Form.Item>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item label="允许群聊" name="allowGroup" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="允许私聊" name="allowDm" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item label="优先级" name="priority" rules={[{ required: true, message: '请输入优先级' }]}>
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={detailState?.title || '日志详情'}
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
