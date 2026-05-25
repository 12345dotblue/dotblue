import React, { useEffect, useMemo, useState } from 'react';
import { Button, Card, Form, InputNumber, Modal, Select, Space, Statistic, Switch, Table, Typography, message } from 'antd';
import { DollarOutlined, LineChartOutlined, PlusOutlined } from '@ant-design/icons';
import axios from 'axios';
import { BACKEND_URL } from '../../config';
import { casdoorService } from '../identity/CasdoorService';

const { Paragraph } = Typography;

interface UsageOverview {
  todayRequests: number;
  todayTokens: number;
  todayCost: number;
  todayCharge: number;
  monthRequests: number;
  monthTokens: number;
  monthCost: number;
  monthCharge: number;
}

interface TrendPoint {
  date: string;
  requestCount: number;
  totalTokens: number;
  costAmount: number;
  chargeAmount: number;
}

interface UsageEventItem {
  id: string;
  createdAt: string;
  agentId: string;
  modelNameSnapshot: string;
  sourceType: string;
  status: string;
  totalTokens: number;
  chargeAmount: number;
}

interface EnterpriseModel {
  id: string;
  displayName: string;
  model: string;
}

interface AgentItem {
  id: string;
  agentName: string;
}

interface ModelPrice {
  id: string;
  modelId: string;
  currency: string;
  costInputUnitPrice: number;
  costOutputUnitPrice: number;
  chargeInputUnitPrice: number;
  chargeOutputUnitPrice: number;
}

interface LimitPolicy {
  id: string;
  scopeType: string;
  scopeId: string;
  enabled: boolean;
  dailyTokenLimit: number;
  monthlyTokenLimit: number;
  dailyChargeLimit: number;
  monthlyChargeLimit: number;
  hardLimit: boolean;
}

interface PolicyFormValues extends LimitPolicy {}

function getAuthHeaders() {
  const token = casdoorService.getToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

const EnterpriseUsageSettingsTab: React.FC = () => {
  const [messageApi, contextHolder] = message.useMessage();
  const [loading, setLoading] = useState(true);
  const [overview, setOverview] = useState<UsageOverview | null>(null);
  const [trends, setTrends] = useState<TrendPoint[]>([]);
  const [events, setEvents] = useState<UsageEventItem[]>([]);
  const [models, setModels] = useState<EnterpriseModel[]>([]);
  const [agents, setAgents] = useState<AgentItem[]>([]);
  const [prices, setPrices] = useState<ModelPrice[]>([]);
  const [policies, setPolicies] = useState<LimitPolicy[]>([]);
  const [priceModalOpen, setPriceModalOpen] = useState(false);
  const [policyModalOpen, setPolicyModalOpen] = useState(false);
  const [editingPrice, setEditingPrice] = useState<ModelPrice | null>(null);
  const [editingPolicy, setEditingPolicy] = useState<LimitPolicy | null>(null);
  const [saving, setSaving] = useState(false);
  const [priceForm] = Form.useForm<ModelPrice>();
  const [policyForm] = Form.useForm<PolicyFormValues>();

  const modelOptions = useMemo(() => models.map((item) => ({
    label: `${item.displayName} (${item.model})`,
    value: item.id,
  })), [models]);

  const modelNameMap = useMemo(() => new Map(models.map((item) => [item.id, `${item.displayName} (${item.model})`])), [models]);
  const agentNameMap = useMemo(() => new Map(agents.map((item) => [item.id, item.agentName])), [agents]);

  const loadData = async () => {
    setLoading(true);
    try {
      const [overviewRes, trendsRes, eventsRes, modelsRes, agentsRes, pricesRes, enterprisePoliciesRes, agentPoliciesRes] = await Promise.all([
        axios.get(`${BACKEND_URL}/api/admin/usage/overview`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/admin/usage/trends?days=7`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/admin/usage/events?page=1&pageSize=10`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/admin/llm-models`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/agents`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/admin/llm-model-prices`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/admin/usage-limit-policies?scopeType=enterprise`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/admin/usage-limit-policies?scopeType=agent`, { headers: getAuthHeaders() }),
      ]);
      setOverview(overviewRes.data || null);
      setTrends(Array.isArray(trendsRes.data) ? trendsRes.data : []);
      setEvents(Array.isArray(eventsRes.data?.items) ? eventsRes.data.items : []);
      setModels(Array.isArray(modelsRes.data) ? modelsRes.data : []);
      setAgents(Array.isArray(agentsRes.data) ? agentsRes.data : []);
      setPrices(Array.isArray(pricesRes.data) ? pricesRes.data : []);
      setPolicies([
        ...(Array.isArray(enterprisePoliciesRes.data) ? enterprisePoliciesRes.data : []),
        ...(Array.isArray(agentPoliciesRes.data) ? agentPoliciesRes.data : []),
      ]);
    } catch {
      messageApi.error('加载企业用量数据失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const openCreatePrice = () => {
    setEditingPrice(null);
    priceForm.setFieldsValue({
      modelId: modelOptions[0]?.value,
      currency: 'USD',
      costInputUnitPrice: 0,
      costOutputUnitPrice: 0,
      chargeInputUnitPrice: 0,
      chargeOutputUnitPrice: 0,
    });
    setPriceModalOpen(true);
  };

  const openEditPrice = (item: ModelPrice) => {
    setEditingPrice(item);
    priceForm.setFieldsValue(item);
    setPriceModalOpen(true);
  };

  const savePrice = async () => {
    const values = await priceForm.validateFields();
    setSaving(true);
    try {
      if (editingPrice) {
        await axios.put(`${BACKEND_URL}/api/admin/llm-model-prices/${editingPrice.id}`, values, { headers: getAuthHeaders() });
        messageApi.success('企业模型内部计费价已更新');
      } else {
        await axios.post(`${BACKEND_URL}/api/admin/llm-model-prices`, values, { headers: getAuthHeaders() });
        messageApi.success('企业模型内部计费价已创建');
      }
      setPriceModalOpen(false);
      await loadData();
    } catch (error: any) {
      messageApi.error(typeof error?.response?.data === 'string' ? error.response.data : '保存企业模型价格失败');
    } finally {
      setSaving(false);
    }
  };

  const deletePrice = async (id: string) => {
    try {
      await axios.delete(`${BACKEND_URL}/api/admin/llm-model-prices/${id}`, { headers: getAuthHeaders() });
      messageApi.success('企业模型内部计费价已删除');
      await loadData();
    } catch {
      messageApi.error('删除企业模型价格失败');
    }
  };

  const openCreatePolicy = () => {
    setEditingPolicy(null);
    policyForm.setFieldsValue({
      scopeType: 'enterprise',
      enabled: true,
      dailyTokenLimit: 0,
      monthlyTokenLimit: 0,
      dailyChargeLimit: 0,
      monthlyChargeLimit: 0,
      hardLimit: true,
    });
    setPolicyModalOpen(true);
  };

  const openEditPolicy = (item: LimitPolicy) => {
    setEditingPolicy(item);
    policyForm.setFieldsValue(item);
    setPolicyModalOpen(true);
  };

  const savePolicy = async () => {
    const values = await policyForm.validateFields();
    setSaving(true);
    try {
      if (editingPolicy) {
        await axios.put(`${BACKEND_URL}/api/admin/usage-limit-policies/${editingPolicy.id}`, values, { headers: getAuthHeaders() });
        messageApi.success('企业限额策略已更新');
      } else {
        await axios.post(`${BACKEND_URL}/api/admin/usage-limit-policies`, values, { headers: getAuthHeaders() });
        messageApi.success('企业限额策略已创建');
      }
      setPolicyModalOpen(false);
      await loadData();
    } catch (error: any) {
      messageApi.error(typeof error?.response?.data === 'string' ? error.response.data : '保存企业限额策略失败');
    } finally {
      setSaving(false);
    }
  };

  const deletePolicy = async (item: LimitPolicy) => {
    try {
      await axios.delete(`${BACKEND_URL}/api/admin/usage-limit-policies/${item.id}?scopeType=${item.scopeType}&scopeId=${item.scopeId}`, { headers: getAuthHeaders() });
      messageApi.success('企业限额策略已删除');
      await loadData();
    } catch {
      messageApi.error('删除企业限额策略失败');
    }
  };

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      {contextHolder}
      <Card variant="borderless" style={{ borderRadius: 20 }} loading={loading}>
        <Paragraph type="secondary" style={{ marginBottom: 16 }}>
          企业可在这里查看 token 审计、配置企业模型内部计费价，并按企业或 Agent 设置消费限制。
        </Paragraph>
        <Space size={16} wrap>
          <Statistic title="今日请求" value={overview?.todayRequests || 0} />
          <Statistic title="今日 Tokens" value={overview?.todayTokens || 0} />
          <Statistic title="今日计费额" value={overview?.todayCharge || 0} precision={4} prefix={<DollarOutlined />} />
          <Statistic title="本月计费额" value={overview?.monthCharge || 0} precision={4} prefix={<DollarOutlined />} />
        </Space>
      </Card>

      <Card variant="borderless" title={<Space><LineChartOutlined />近 7 天趋势</Space>} style={{ borderRadius: 20 }} loading={loading}>
        <Table
          rowKey="date"
          pagination={false}
          dataSource={trends}
          columns={[
            { title: '日期', dataIndex: 'date', key: 'date', width: 120 },
            { title: '请求数', dataIndex: 'requestCount', key: 'requestCount', width: 100 },
            { title: 'Tokens', dataIndex: 'totalTokens', key: 'totalTokens', width: 140 },
            { title: '成本', dataIndex: 'costAmount', key: 'costAmount', width: 140, render: (value: number) => value.toFixed(4) },
            { title: '计费额', dataIndex: 'chargeAmount', key: 'chargeAmount', width: 140, render: (value: number) => value.toFixed(4) },
          ]}
        />
      </Card>

      <Card variant="borderless" title="最近调用事件" style={{ borderRadius: 20 }} loading={loading}>
        <Table
          rowKey="id"
          pagination={false}
          dataSource={events}
          scroll={{ x: 900 }}
          columns={[
            { title: '时间', dataIndex: 'createdAt', key: 'createdAt', width: 180 },
            { title: '模型', dataIndex: 'modelNameSnapshot', key: 'modelNameSnapshot', width: 220 },
            { title: 'Agent', dataIndex: 'agentId', key: 'agentId', width: 220, render: (value: string) => agentNameMap.get(value) || value },
            { title: '来源', dataIndex: 'sourceType', key: 'sourceType', width: 100 },
            { title: '状态', dataIndex: 'status', key: 'status', width: 110 },
            { title: 'Tokens', dataIndex: 'totalTokens', key: 'totalTokens', width: 120 },
            { title: '计费额', dataIndex: 'chargeAmount', key: 'chargeAmount', width: 120, render: (value: number) => value.toFixed(4) },
          ]}
        />
      </Card>

      <Card
        variant="borderless"
        title="企业模型内部计费价"
        extra={<Button type="primary" icon={<PlusOutlined />} onClick={openCreatePrice}>新增价格</Button>}
        style={{ borderRadius: 20 }}
        loading={loading}
      >
        <Table
          rowKey="id"
          pagination={false}
          dataSource={prices}
          columns={[
            { title: '模型', dataIndex: 'modelId', key: 'modelId', render: (value: string) => modelNameMap.get(value) || value },
            { title: '币种', dataIndex: 'currency', key: 'currency', width: 100 },
            { title: '输入成本价', dataIndex: 'costInputUnitPrice', key: 'costInputUnitPrice', width: 140 },
            { title: '输出成本价', dataIndex: 'costOutputUnitPrice', key: 'costOutputUnitPrice', width: 140 },
            { title: '输入计费价', dataIndex: 'chargeInputUnitPrice', key: 'chargeInputUnitPrice', width: 140 },
            { title: '输出计费价', dataIndex: 'chargeOutputUnitPrice', key: 'chargeOutputUnitPrice', width: 140 },
            {
              title: '操作',
              key: 'actions',
              width: 160,
              render: (_: unknown, item: ModelPrice) => (
                <Space>
                  <Button onClick={() => openEditPrice(item)}>编辑</Button>
                  <Button danger onClick={() => deletePrice(item.id)}>删除</Button>
                </Space>
              ),
            },
          ]}
        />
      </Card>

      <Card
        variant="borderless"
        title="企业消费限额"
        extra={<Button type="primary" icon={<PlusOutlined />} onClick={openCreatePolicy}>新增策略</Button>}
        style={{ borderRadius: 20 }}
        loading={loading}
      >
        <Table
          rowKey="id"
          pagination={false}
          dataSource={policies}
          columns={[
            { title: '作用域', dataIndex: 'scopeType', key: 'scopeType', width: 110, render: (value: string) => value === 'agent' ? 'Agent' : '企业' },
            { title: '目标', dataIndex: 'scopeId', key: 'scopeId', render: (value: string, item: LimitPolicy) => item.scopeType === 'agent' ? (agentNameMap.get(value) || value) : '当前企业' },
            { title: '启用', dataIndex: 'enabled', key: 'enabled', width: 90, render: (value: boolean) => value ? '是' : '否' },
            { title: '日 Token 限额', dataIndex: 'dailyTokenLimit', key: 'dailyTokenLimit', width: 140 },
            { title: '月 Token 限额', dataIndex: 'monthlyTokenLimit', key: 'monthlyTokenLimit', width: 140 },
            { title: '日计费限额', dataIndex: 'dailyChargeLimit', key: 'dailyChargeLimit', width: 140 },
            { title: '月计费限额', dataIndex: 'monthlyChargeLimit', key: 'monthlyChargeLimit', width: 140 },
            {
              title: '操作',
              key: 'actions',
              width: 160,
              render: (_: unknown, item: LimitPolicy) => (
                <Space>
                  <Button onClick={() => openEditPolicy(item)}>编辑</Button>
                  <Button danger onClick={() => deletePolicy(item)}>删除</Button>
                </Space>
              ),
            },
          ]}
        />
      </Card>

      <Modal
        title={editingPrice ? '编辑企业模型价格' : '新增企业模型价格'}
        open={priceModalOpen}
        onOk={savePrice}
        onCancel={() => setPriceModalOpen(false)}
        confirmLoading={saving}
        destroyOnClose
      >
        <Form form={priceForm} layout="vertical">
          <Form.Item label="模型" name="modelId" rules={[{ required: true, message: '请选择模型' }]}>
            <Select options={modelOptions} />
          </Form.Item>
          <Form.Item label="币种" name="currency">
            <Select options={[{ label: 'USD', value: 'USD' }, { label: 'CNY', value: 'CNY' }]} />
          </Form.Item>
          <Form.Item label="输入成本价 / 1M tokens" name="costInputUnitPrice">
            <InputNumber style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label="输出成本价 / 1M tokens" name="costOutputUnitPrice">
            <InputNumber style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label="输入计费价 / 1M tokens" name="chargeInputUnitPrice">
            <InputNumber style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label="输出计费价 / 1M tokens" name="chargeOutputUnitPrice">
            <InputNumber style={{ width: '100%' }} min={0} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={editingPolicy ? '编辑企业限额策略' : '新增企业限额策略'}
        open={policyModalOpen}
        onOk={savePolicy}
        onCancel={() => setPolicyModalOpen(false)}
        confirmLoading={saving}
        destroyOnClose
      >
        <Form form={policyForm} layout="vertical">
          <Form.Item label="作用域" name="scopeType" rules={[{ required: true, message: '请选择作用域' }]}>
            <Select options={[{ label: '企业', value: 'enterprise' }, { label: 'Agent', value: 'agent' }]} />
          </Form.Item>
          <Form.Item noStyle shouldUpdate={(prev, next) => prev.scopeType !== next.scopeType}>
            {({ getFieldValue }) => (
              getFieldValue('scopeType') === 'agent' ? (
                <Form.Item label="Agent" name="scopeId" rules={[{ required: true, message: '请选择 Agent' }]}>
                  <Select options={agents.map((item) => ({ label: item.agentName, value: item.id }))} />
                </Form.Item>
              ) : null
            )}
          </Form.Item>
          <Form.Item label="启用" name="enabled" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item label="日 Token 限额" name="dailyTokenLimit">
            <InputNumber style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label="月 Token 限额" name="monthlyTokenLimit">
            <InputNumber style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label="日计费限额" name="dailyChargeLimit">
            <InputNumber style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label="月计费限额" name="monthlyChargeLimit">
            <InputNumber style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label="硬限制" name="hardLimit" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
};

export default EnterpriseUsageSettingsTab;
