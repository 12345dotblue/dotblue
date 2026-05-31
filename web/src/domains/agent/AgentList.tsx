import React, { useState, useEffect } from 'react';
import { Card, Button, Modal, Form, Input, message, Typography, Space, Empty, Popconfirm, Select, Tag, Statistic, Table, Radio, Tabs } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, RobotOutlined, LineChartOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import { BACKEND_URL } from '../../config';
import { getLocalizedPath, getPreferredLanguage } from '../../i18n/config';
import { casdoorService } from '../identity/CasdoorService';
import AgentSkillsPanel from './AgentSkillsPanel';

const { Title, Paragraph, Text } = Typography;
const { TextArea } = Input;

interface AgentItem {
  id: string;
  agentName: string;
  systemPrompt: string;
  modelScope: 'platform' | 'enterprise';
  modelId?: string;
  modelName?: string;
  engineType: 'hermes' | 'nanobot';
  todayTokens?: number;
  todayCharge?: number;
  monthTokens?: number;
  monthCharge?: number;
  createdAt: string;
}

interface AgentUsageOverview {
  todayRequests: number;
  todayTokens: number;
  todayCost: number;
  todayCharge: number;
  monthRequests: number;
  monthTokens: number;
  monthCost: number;
  monthCharge: number;
}

interface AgentTrendPoint {
  date: string;
  requestCount: number;
  totalTokens: number;
  costAmount: number;
  chargeAmount: number;
}

type AgentOverviewMap = Record<string, AgentUsageOverview>;

interface ModelOptionItem {
  label: string;
  value: string;
}

interface ModelOptionGroup {
  label: string;
  options: ModelOptionItem[];
}

interface RuntimeOptionItem {
  value: 'hermes' | 'nanobot';
}

interface AgentOptionsResponse {
  modelOptions: ModelOptionGroup[];
  runtimeOptions: RuntimeOptionItem[];
}

interface AgentFormValues {
  agentName: string;
  systemPrompt: string;
  modelSelection: string;
  engineType: 'hermes' | 'nanobot';
}

function formatEngineLabel(engineType: string, t: (key: string) => string): string {
  return engineType === 'nanobot' ? t('agent_engine_nanobot') : t('agent_engine_hermes');
}

const AgentList: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const currentLanguage = getPreferredLanguage();
  const [agents, setAgents] = useState<AgentItem[]>([]);
  const [modelOptions, setModelOptions] = useState<ModelOptionGroup[]>([]);
  const [runtimeOptions, setRuntimeOptions] = useState<RuntimeOptionItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingAgent, setEditingAgent] = useState<AgentItem | null>(null);
  const [usageModalOpen, setUsageModalOpen] = useState(false);
  const [usageLoading, setUsageLoading] = useState(false);
  const [usageAgent, setUsageAgent] = useState<AgentItem | null>(null);
  const [usageOverview, setUsageOverview] = useState<AgentUsageOverview | null>(null);
  const [usageTrends, setUsageTrends] = useState<AgentTrendPoint[]>([]);
  const [agentOverviewMap, setAgentOverviewMap] = useState<AgentOverviewMap>({});
  const [form] = Form.useForm<AgentFormValues>();
  const [saving, setSaving] = useState(false);

  const fetchAgentOverviews = async (items: AgentItem[]) => {
    const token = localStorage.getItem('casdoor_token');
    try {
      const results = await Promise.all(items.map(async (item) => {
        const res = await axios.get(`${BACKEND_URL}/api/agents/${item.id}/usage/overview`, {
          headers: { Authorization: `Bearer ${token}` },
        });
        return [item.id, res.data || null] as const;
      }));
      setAgentOverviewMap(Object.fromEntries(results.filter(([, value]) => value)));
    } catch {
      setAgentOverviewMap({});
    }
  };

  const fetchAgents = () => {
    const token = localStorage.getItem('casdoor_token');
    axios.get(`${BACKEND_URL}/api/agents`, {
      headers: { Authorization: `Bearer ${token}` },
    }).then(res => {
      const list = res.data || [];
      setAgents(list);
      fetchAgentOverviews(list);
    }).catch(() => {
      setAgents([]);
      setAgentOverviewMap({});
    }).finally(() => {
      setLoading(false);
    });
  };

  const resolvedRuntimeOptions = runtimeOptions.length > 0 ? runtimeOptions : [{ value: 'hermes' as const }];

  const defaultRuntimeEngine = () => resolvedRuntimeOptions[0]?.value || 'hermes';

  const fetchAgentOptions = () => {
    const token = localStorage.getItem('casdoor_token');
    axios.get(`${BACKEND_URL}/api/agents/model-options`, {
      headers: { Authorization: `Bearer ${token}` },
    }).then(res => {
      const data: AgentOptionsResponse | ModelOptionGroup[] = res.data;
      if (Array.isArray(data)) {
        setModelOptions(data);
        setRuntimeOptions([{ value: 'hermes' }]);
        return;
      }
      setModelOptions(Array.isArray(data?.modelOptions) ? data.modelOptions : []);
      setRuntimeOptions(Array.isArray(data?.runtimeOptions) && data.runtimeOptions.length > 0 ? data.runtimeOptions : [{ value: 'hermes' }]);
    }).catch(() => {
      setModelOptions([]);
      setRuntimeOptions([{ value: 'hermes' }]);
    });
  };

  const getDefaultModelSelection = () => modelOptions[0]?.options?.[0]?.value;

  const toModelSelection = (agent?: Pick<AgentItem, 'modelScope' | 'modelId'> | null) => {
    if (agent?.modelScope === 'enterprise' && agent.modelId) {
      return `enterprise:${agent.modelId}`;
    }
    if (agent?.modelScope === 'platform' && agent.modelId) {
      return `platform:${agent.modelId}`;
    }
    return getDefaultModelSelection();
  };

  const parseModelSelection = (value: string) => {
    if (value.startsWith('enterprise:')) {
      return {
        modelScope: 'enterprise',
        modelId: value.slice('enterprise:'.length),
      };
    }
    if (value.startsWith('platform:')) {
      return {
        modelScope: 'platform',
        modelId: value.slice('platform:'.length),
      };
    }
    return {
      modelScope: 'platform',
      modelId: value,
    };
  };

  useEffect(() => {
    fetchAgents();
    fetchAgentOptions();
  }, []);

  const openCreate = () => {
    setEditingAgent(null);
    form.resetFields();
    form.setFieldValue('modelSelection', getDefaultModelSelection());
    form.setFieldValue('engineType', defaultRuntimeEngine());
    setModalOpen(true);
  };

  const openEdit = (agent: AgentItem) => {
    setEditingAgent(agent);
    form.setFieldsValue({
      agentName: agent.agentName,
      systemPrompt: agent.systemPrompt,
      modelSelection: toModelSelection(agent),
      engineType: agent.engineType || 'hermes',
    });
    setModalOpen(true);
  };

  const handleSave = async () => {
    try {
      const values = await form.validateFields();
      const { modelScope, modelId } = parseModelSelection(values.modelSelection);
      setSaving(true);
      const token = localStorage.getItem('casdoor_token');
      const payload = {
        agentName: values.agentName,
        systemPrompt: values.systemPrompt,
        modelScope,
        modelId,
        engineType: values.engineType,
      };

      if (editingAgent) {
        await axios.put(`${BACKEND_URL}/api/agents/${editingAgent.id}`, payload, {
          headers: { Authorization: `Bearer ${token}` },
        });
        message.success(t('agent_update_success'));
      } else {
        await axios.post(`${BACKEND_URL}/api/agents`, payload, {
          headers: { Authorization: `Bearer ${token}` },
        });
        message.success(t('agent_create_success'));
      }

      setModalOpen(false);
      fetchAgents();
    } catch {
      // validation error or API error
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (agentId: string) => {
    try {
      const token = localStorage.getItem('casdoor_token');
      await axios.delete(`${BACKEND_URL}/api/agents/${agentId}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      message.success(t('agent_delete_success'));
      fetchAgents();
    } catch {
      message.error(t('agent_delete_failed'));
    }
  };

  const openUsage = async (agent: AgentItem) => {
    const token = localStorage.getItem('casdoor_token');
    setUsageAgent(agent);
    setUsageModalOpen(true);
    setUsageLoading(true);
    try {
      const [overviewRes, trendsRes] = await Promise.all([
        axios.get(`${BACKEND_URL}/api/agents/${agent.id}/usage/overview`, {
          headers: { Authorization: `Bearer ${token}` },
        }),
        axios.get(`${BACKEND_URL}/api/agents/${agent.id}/usage/trends?days=7`, {
          headers: { Authorization: `Bearer ${token}` },
        }),
      ]);
      setUsageOverview(overviewRes.data || null);
      setUsageTrends(Array.isArray(trendsRes.data) ? trendsRes.data : []);
    } catch {
      message.error('加载 Agent 用量失败');
      setUsageOverview(null);
      setUsageTrends([]);
    } finally {
      setUsageLoading(false);
    }
  };

  return (
    <div style={{ animation: 'fadeIn 0.5s ease-out' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24, maxWidth: 800 }}>
        <div>
          <Title level={4} style={{ marginBottom: 4 }}>
            <RobotOutlined style={{ marginRight: 8 }} />
            {t('agent_list_title')}
          </Title>
          <Text type="secondary">{t('agent_list_desc')}</Text>
        </div>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
          {t('agent_create')}
        </Button>
      </div>

      {agents.length === 0 && !loading ? (
        <Card variant="borderless" style={{ borderRadius: 12, maxWidth: 800, textAlign: 'center', padding: '40px 0' }}>
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description={t('agent_no_agents')}
          >
            <Space>
              <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
                {t('agent_create_first')}
              </Button>
              <Button onClick={() => navigate(getLocalizedPath('/chat', currentLanguage))}>
                {t('agent_go_chat')}
              </Button>
            </Space>
          </Empty>
        </Card>
      ) : (
        <div style={{ maxWidth: 800 }}>
          <Space orientation="vertical" size={16} style={{ width: '100%' }}>
            {agents.map((item) => (
              <Card
                key={item.id}
                variant="borderless"
                style={{ borderRadius: 12, boxShadow: '0 4px 20px rgba(0,0,0,0.03)' }}
              >
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <Space>
                    <RobotOutlined style={{ color: '#1677ff', fontSize: 20 }} />
                    <Title level={5} style={{ margin: 0 }}>{item.agentName}</Title>
                  </Space>
                  <Paragraph
                    type="secondary"
                    ellipsis={{ rows: 2 }}
                    style={{ marginTop: 8, marginBottom: 0 }}
                  >
                    {item.systemPrompt}
                  </Paragraph>
                  <div style={{ marginTop: 12 }}>
                    <Tag color={item.modelScope === 'enterprise' ? 'blue' : 'green'}>
                      {item.modelScope === 'enterprise' ? t('agent_model_scope_enterprise') : t('agent_model_scope_platform')}
                    </Tag>
                    <Tag color={item.engineType === 'nanobot' ? 'purple' : 'gold'}>
                      {formatEngineLabel(item.engineType, t)}
                    </Tag>
                    {item.modelName ? <Text type="secondary">{item.modelName}</Text> : null}
                  </div>
                  <div style={{ marginTop: 12, display: 'flex', gap: 24, flexWrap: 'wrap' }}>
                    <Text type="secondary">今日 Tokens: {agentOverviewMap[item.id]?.todayTokens || 0}</Text>
                    <Text type="secondary">今日计费: {(agentOverviewMap[item.id]?.todayCharge || 0).toFixed(4)}</Text>
                    <Text type="secondary">本月 Tokens: {agentOverviewMap[item.id]?.monthTokens || 0}</Text>
                    <Text type="secondary">本月计费: {(agentOverviewMap[item.id]?.monthCharge || 0).toFixed(4)}</Text>
                  </div>
                </div>
                <Space style={{ marginLeft: 16, flexShrink: 0 }}>
                  <Button icon={<LineChartOutlined />} onClick={() => openUsage(item)}>
                    用量
                  </Button>
                  <Button icon={<EditOutlined />} onClick={() => openEdit(item)}>
                    {t('agent_edit')}
                  </Button>
                  <Popconfirm
                    title={t('agent_confirm_delete')}
                    onConfirm={() => handleDelete(item.id)}
                    okText={t('agent_delete')}
                    cancelText={t('agent_cancel')}
                  >
                    <Button danger icon={<DeleteOutlined />}>
                      {t('agent_delete')}
                    </Button>
                  </Popconfirm>
                </Space>
              </div>
              </Card>
            ))}
          </Space>
        </div>
      )}

      <Modal
        title={editingAgent ? t('agent_edit') : t('agent_create')}
        open={modalOpen}
        onOk={handleSave}
        onCancel={() => setModalOpen(false)}
        confirmLoading={saving}
        okText={editingAgent ? t('agent_save') : t('agent_create')}
        destroyOnHidden
      >
        <Tabs
          defaultActiveKey="basic"
          items={[
            {
              key: 'basic',
              label: '基础配置',
              children: (
                <Form
                  form={form}
                  layout="vertical"
                  initialValues={{ agentName: '', systemPrompt: '', modelSelection: getDefaultModelSelection(), engineType: 'hermes' }}
                >
                  <Form.Item
                    label={t('agent_name')}
                    name="agentName"
                    rules={[{ required: true, message: t('agent_name_required') }]}
                  >
                    <Input placeholder={t('placeholder_agent_name')} />
                  </Form.Item>
                  <Form.Item
                    label={t('system_prompt')}
                    name="systemPrompt"
                    rules={[{ required: true, message: t('system_prompt_required') }]}
                  >
                    <TextArea rows={6} placeholder={t('placeholder_system_prompt')} />
                  </Form.Item>
                  <Form.Item
                    label={t('agent_model')}
                    name="modelSelection"
                    rules={[{ required: true, message: t('agent_model_required') }]}
                  >
                    <Select
                      placeholder={t('agent_model_placeholder')}
                      options={modelOptions}
                      notFoundContent={t('agent_model_empty')}
                    />
                  </Form.Item>
                  <Form.Item
                    label={t('agent_engine')}
                    name="engineType"
                    rules={[{ required: true, message: t('agent_engine_required') }]}
                  >
                    <Radio.Group
                      options={resolvedRuntimeOptions.map((item) => ({
                        label: formatEngineLabel(item.value, t),
                        value: item.value,
                      }))}
                      optionType="button"
                      buttonStyle="solid"
                    />
                  </Form.Item>
                </Form>
              ),
            },
            {
              key: 'skills',
              label: '已安装 Skills',
              children: editingAgent ? (
                <AgentSkillsPanel
                  agentId={editingAgent.id}
                  authHeaders={(() => {
                    const token = casdoorService.getToken();
                    return token ? { Authorization: `Bearer ${token}` } : {};
                  })()}
                />
              ) : (
                <Empty description="请先创建 Agent，再安装 Skill" image={Empty.PRESENTED_IMAGE_SIMPLE} />
              ),
            },
          ]}
        />
      </Modal>

      <Modal
        title={usageAgent ? `${usageAgent.agentName} 用量` : 'Agent 用量'}
        open={usageModalOpen}
        onCancel={() => setUsageModalOpen(false)}
        footer={null}
        width={880}
        destroyOnHidden
      >
        <Space orientation="vertical" size={16} style={{ width: '100%' }}>
          <Card variant="borderless" loading={usageLoading} style={{ background: '#fafafa' }}>
            <Space size={24} wrap>
              <Statistic title="今日请求" value={usageOverview?.todayRequests || 0} />
              <Statistic title="今日 Tokens" value={usageOverview?.todayTokens || 0} />
              <Statistic title="今日计费" value={usageOverview?.todayCharge || 0} precision={4} />
              <Statistic title="本月请求" value={usageOverview?.monthRequests || 0} />
              <Statistic title="本月 Tokens" value={usageOverview?.monthTokens || 0} />
              <Statistic title="本月计费" value={usageOverview?.monthCharge || 0} precision={4} />
            </Space>
          </Card>
          <Card variant="borderless" title="近 7 天趋势">
            <Table
              rowKey="date"
              loading={usageLoading}
              pagination={false}
              dataSource={usageTrends}
              columns={[
                { title: '日期', dataIndex: 'date', key: 'date', width: 120 },
                { title: '请求数', dataIndex: 'requestCount', key: 'requestCount', width: 120 },
                { title: 'Tokens', dataIndex: 'totalTokens', key: 'totalTokens', width: 140 },
                { title: '成本', dataIndex: 'costAmount', key: 'costAmount', width: 140, render: (value: number) => value.toFixed(4) },
                { title: '计费', dataIndex: 'chargeAmount', key: 'chargeAmount', width: 140, render: (value: number) => value.toFixed(4) },
              ]}
            />
          </Card>
        </Space>
      </Modal>
    </div>
  );
};

export default AgentList;
