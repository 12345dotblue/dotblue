import React, { useEffect, useMemo, useState } from 'react';
import { Button, Card, Form, InputNumber, Modal, Select, Space, Statistic, Switch, Table, Typography, message } from 'antd';
import { DollarOutlined, LineChartOutlined, PlusOutlined } from '@ant-design/icons';
import axios from 'axios';
import { useTranslation } from 'react-i18next';
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
  const { t } = useTranslation();
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
      messageApi.error(t('enterprise_usage_load_failed'));
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
        messageApi.success(t('enterprise_usage_price_updated'));
      } else {
        await axios.post(`${BACKEND_URL}/api/admin/llm-model-prices`, values, { headers: getAuthHeaders() });
        messageApi.success(t('enterprise_usage_price_created'));
      }
      setPriceModalOpen(false);
      await loadData();
    } catch (error: any) {
      messageApi.error(t('enterprise_usage_price_save_failed'));
    } finally {
      setSaving(false);
    }
  };

  const deletePrice = async (id: string) => {
    try {
      await axios.delete(`${BACKEND_URL}/api/admin/llm-model-prices/${id}`, { headers: getAuthHeaders() });
      messageApi.success(t('enterprise_usage_price_deleted'));
      await loadData();
    } catch {
      messageApi.error(t('enterprise_usage_price_delete_failed'));
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
        messageApi.success(t('enterprise_usage_policy_updated'));
      } else {
        await axios.post(`${BACKEND_URL}/api/admin/usage-limit-policies`, values, { headers: getAuthHeaders() });
        messageApi.success(t('enterprise_usage_policy_created'));
      }
      setPolicyModalOpen(false);
      await loadData();
    } catch (error: any) {
      messageApi.error(t('enterprise_usage_policy_save_failed'));
    } finally {
      setSaving(false);
    }
  };

  const deletePolicy = async (item: LimitPolicy) => {
    try {
      await axios.delete(`${BACKEND_URL}/api/admin/usage-limit-policies/${item.id}?scopeType=${item.scopeType}&scopeId=${item.scopeId}`, { headers: getAuthHeaders() });
      messageApi.success(t('enterprise_usage_policy_deleted'));
      await loadData();
    } catch {
      messageApi.error(t('enterprise_usage_policy_delete_failed'));
    }
  };

  return (
    <Space orientation="vertical" size={16} style={{ width: '100%' }}>
      {contextHolder}
      <Card variant="borderless" style={{ borderRadius: 20 }} loading={loading}>
        <Paragraph type="secondary" style={{ marginBottom: 16 }}>
          {t('enterprise_usage_summary_description')}
        </Paragraph>
        <Space size={16} wrap>
          <Statistic title={t('enterprise_usage_today_requests')} value={overview?.todayRequests || 0} />
          <Statistic title={t('enterprise_usage_today_tokens')} value={overview?.todayTokens || 0} />
          <Statistic title={t('enterprise_usage_today_charge')} value={overview?.todayCharge || 0} precision={4} prefix={<DollarOutlined />} />
          <Statistic title={t('enterprise_usage_month_charge')} value={overview?.monthCharge || 0} precision={4} prefix={<DollarOutlined />} />
        </Space>
      </Card>

      <Card variant="borderless" title={<Space><LineChartOutlined />{t('enterprise_usage_seven_day_trend')}</Space>} style={{ borderRadius: 20 }} loading={loading}>
        <Table
          rowKey="date"
          pagination={false}
          dataSource={trends}
          columns={[
            { title: t('enterprise_usage_date'), dataIndex: 'date', key: 'date', width: 120 },
            { title: t('enterprise_usage_request_count'), dataIndex: 'requestCount', key: 'requestCount', width: 100 },
            { title: t('enterprise_usage_tokens_label'), dataIndex: 'totalTokens', key: 'totalTokens', width: 140 },
            { title: t('enterprise_usage_cost'), dataIndex: 'costAmount', key: 'costAmount', width: 140, render: (value: number) => value.toFixed(4) },
            { title: t('enterprise_usage_charge'), dataIndex: 'chargeAmount', key: 'chargeAmount', width: 140, render: (value: number) => value.toFixed(4) },
          ]}
        />
      </Card>

      <Card variant="borderless" title={t('enterprise_usage_recent_events')} style={{ borderRadius: 20 }} loading={loading}>
        <Table
          rowKey="id"
          pagination={false}
          dataSource={events}
          scroll={{ x: 900 }}
          columns={[
            { title: t('enterprise_usage_time'), dataIndex: 'createdAt', key: 'createdAt', width: 180 },
            { title: t('enterprise_usage_model'), dataIndex: 'modelNameSnapshot', key: 'modelNameSnapshot', width: 220 },
            { title: t('enterprise_usage_agent_label'), dataIndex: 'agentId', key: 'agentId', width: 220, render: (value: string) => agentNameMap.get(value) || value },
            { title: t('enterprise_usage_source'), dataIndex: 'sourceType', key: 'sourceType', width: 100 },
            { title: t('enterprise_usage_status'), dataIndex: 'status', key: 'status', width: 110 },
            { title: t('enterprise_usage_tokens_label'), dataIndex: 'totalTokens', key: 'totalTokens', width: 120 },
            { title: t('enterprise_usage_charge'), dataIndex: 'chargeAmount', key: 'chargeAmount', width: 120, render: (value: number) => value.toFixed(4) },
          ]}
        />
      </Card>

      <Card
        variant="borderless"
        title={t('enterprise_usage_enterprise_model_pricing')}
        extra={<Button type="primary" icon={<PlusOutlined />} onClick={openCreatePrice}>{t('enterprise_usage_add_price')}</Button>}
        style={{ borderRadius: 20 }}
        loading={loading}
      >
        <Table
          rowKey="id"
          pagination={false}
          dataSource={prices}
          columns={[
            { title: t('enterprise_usage_model'), dataIndex: 'modelId', key: 'modelId', render: (value: string) => modelNameMap.get(value) || value },
            { title: t('enterprise_usage_currency'), dataIndex: 'currency', key: 'currency', width: 100 },
            { title: t('enterprise_usage_input_cost'), dataIndex: 'costInputUnitPrice', key: 'costInputUnitPrice', width: 140 },
            { title: t('enterprise_usage_output_cost'), dataIndex: 'costOutputUnitPrice', key: 'costOutputUnitPrice', width: 140 },
            { title: t('enterprise_usage_input_charge'), dataIndex: 'chargeInputUnitPrice', key: 'chargeInputUnitPrice', width: 140 },
            { title: t('enterprise_usage_output_charge'), dataIndex: 'chargeOutputUnitPrice', key: 'chargeOutputUnitPrice', width: 140 },
            {
              title: t('enterprise_usage_actions'),
              key: 'actions',
              width: 160,
              render: (_: unknown, item: ModelPrice) => (
                <Space>
                  <Button onClick={() => openEditPrice(item)}>{t('enterprise_usage_edit')}</Button>
                  <Button danger onClick={() => deletePrice(item.id)}>{t('enterprise_usage_delete')}</Button>
                </Space>
              ),
            },
          ]}
        />
      </Card>

      <Card
        variant="borderless"
        title={t('enterprise_usage_enterprise_limit_policy')}
        extra={<Button type="primary" icon={<PlusOutlined />} onClick={openCreatePolicy}>{t('enterprise_usage_add_policy')}</Button>}
        style={{ borderRadius: 20 }}
        loading={loading}
      >
        <Table
          rowKey="id"
          pagination={false}
          dataSource={policies}
          columns={[
            { title: t('enterprise_usage_scope'), dataIndex: 'scopeType', key: 'scopeType', width: 110, render: (value: string) => value === 'agent' ? t('enterprise_usage_agent_label') : t('enterprise_usage_enterprise_scope') },
            { title: t('enterprise_usage_target'), dataIndex: 'scopeId', key: 'scopeId', render: (value: string, item: LimitPolicy) => item.scopeType === 'agent' ? (agentNameMap.get(value) || value) : t('enterprise_usage_current_enterprise') },
            { title: t('enterprise_usage_enabled'), dataIndex: 'enabled', key: 'enabled', width: 90, render: (value: boolean) => value ? t('enterprise_usage_yes') : t('enterprise_usage_no') },
            { title: t('enterprise_usage_daily_token_limit'), dataIndex: 'dailyTokenLimit', key: 'dailyTokenLimit', width: 140 },
            { title: t('enterprise_usage_monthly_token_limit'), dataIndex: 'monthlyTokenLimit', key: 'monthlyTokenLimit', width: 140 },
            { title: t('enterprise_usage_daily_charge_limit'), dataIndex: 'dailyChargeLimit', key: 'dailyChargeLimit', width: 140 },
            { title: t('enterprise_usage_monthly_charge_limit'), dataIndex: 'monthlyChargeLimit', key: 'monthlyChargeLimit', width: 140 },
            {
              title: t('enterprise_usage_actions'),
              key: 'actions',
              width: 160,
              render: (_: unknown, item: LimitPolicy) => (
                <Space>
                  <Button onClick={() => openEditPolicy(item)}>{t('enterprise_usage_edit')}</Button>
                  <Button danger onClick={() => deletePolicy(item)}>{t('enterprise_usage_delete')}</Button>
                </Space>
              ),
            },
          ]}
        />
      </Card>

      <Modal
        title={editingPrice ? t('enterprise_usage_edit_model_price') : t('enterprise_usage_create_model_price')}
        open={priceModalOpen}
        onOk={savePrice}
        onCancel={() => setPriceModalOpen(false)}
        confirmLoading={saving}
        destroyOnHidden
      >
        <Form form={priceForm} layout="vertical">
          <Form.Item label={t('enterprise_usage_model')} name="modelId" rules={[{ required: true, message: t('enterprise_usage_select_model') }]}>
            <Select options={modelOptions} />
          </Form.Item>
          <Form.Item label={t('enterprise_usage_currency')} name="currency">
            <Select options={[{ label: t('enterprise_usage_currency_usd'), value: 'USD' }, { label: t('enterprise_usage_currency_cny'), value: 'CNY' }]} />
          </Form.Item>
          <Form.Item label={`${t('enterprise_usage_input_cost')} ${t('enterprise_usage_per_million_tokens')}`} name="costInputUnitPrice">
            <InputNumber controls={false} style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label={`${t('enterprise_usage_output_cost')} ${t('enterprise_usage_per_million_tokens')}`} name="costOutputUnitPrice">
            <InputNumber controls={false} style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label={`${t('enterprise_usage_input_charge')} ${t('enterprise_usage_per_million_tokens')}`} name="chargeInputUnitPrice">
            <InputNumber controls={false} style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label={`${t('enterprise_usage_output_charge')} ${t('enterprise_usage_per_million_tokens')}`} name="chargeOutputUnitPrice">
            <InputNumber controls={false} style={{ width: '100%' }} min={0} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={editingPolicy ? t('enterprise_usage_edit_limit_policy') : t('enterprise_usage_create_limit_policy')}
        open={policyModalOpen}
        onOk={savePolicy}
        onCancel={() => setPolicyModalOpen(false)}
        confirmLoading={saving}
        destroyOnHidden
      >
        <Form form={policyForm} layout="vertical">
          <Form.Item label={t('enterprise_usage_scope')} name="scopeType" rules={[{ required: true, message: t('enterprise_usage_select_scope') }]}>
            <Select options={[{ label: t('enterprise_usage_enterprise_scope'), value: 'enterprise' }, { label: t('enterprise_usage_agent_label'), value: 'agent' }]} />
          </Form.Item>
          <Form.Item noStyle shouldUpdate={(prev, next) => prev.scopeType !== next.scopeType}>
            {({ getFieldValue }) => (
              getFieldValue('scopeType') === 'agent' ? (
                <Form.Item label={t('enterprise_usage_agent_label')} name="scopeId" rules={[{ required: true, message: t('enterprise_usage_select_agent') }]}>
                  <Select options={agents.map((item) => ({ label: item.agentName, value: item.id }))} />
                </Form.Item>
              ) : null
            )}
          </Form.Item>
          <Form.Item label={t('enterprise_usage_enabled')} name="enabled" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item label={t('enterprise_usage_daily_token_limit')} name="dailyTokenLimit">
            <InputNumber controls={false} style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label={t('enterprise_usage_monthly_token_limit')} name="monthlyTokenLimit">
            <InputNumber controls={false} style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label={t('enterprise_usage_daily_charge_limit')} name="dailyChargeLimit">
            <InputNumber controls={false} style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label={t('enterprise_usage_monthly_charge_limit')} name="monthlyChargeLimit">
            <InputNumber controls={false} style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label={t('enterprise_usage_hard_limit')} name="hardLimit" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
};

export default EnterpriseUsageSettingsTab;


