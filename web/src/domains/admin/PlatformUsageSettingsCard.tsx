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

interface PlatformModel {
  id: string;
  displayName: string;
  model: string;
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
  enabled: boolean;
  dailyTokenLimit: number;
  monthlyTokenLimit: number;
  dailyChargeLimit: number;
  monthlyChargeLimit: number;
  hardLimit: boolean;
}

function getAuthHeaders() {
  const token = casdoorService.getToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

const PlatformUsageSettingsCard: React.FC = () => {
  const { t } = useTranslation();
  const [messageApi, contextHolder] = message.useMessage();
  const [loading, setLoading] = useState(true);
  const [overview, setOverview] = useState<UsageOverview | null>(null);
  const [trends, setTrends] = useState<TrendPoint[]>([]);
  const [events, setEvents] = useState<UsageEventItem[]>([]);
  const [models, setModels] = useState<PlatformModel[]>([]);
  const [prices, setPrices] = useState<ModelPrice[]>([]);
  const [policies, setPolicies] = useState<LimitPolicy[]>([]);
  const [priceModalOpen, setPriceModalOpen] = useState(false);
  const [policyModalOpen, setPolicyModalOpen] = useState(false);
  const [editingPrice, setEditingPrice] = useState<ModelPrice | null>(null);
  const [editingPolicy, setEditingPolicy] = useState<LimitPolicy | null>(null);
  const [saving, setSaving] = useState(false);
  const [priceForm] = Form.useForm<ModelPrice>();
  const [policyForm] = Form.useForm<LimitPolicy>();

  const modelOptions = useMemo(() => models.map((item) => ({
    label: `${item.displayName} (${item.model})`,
    value: item.id,
  })), [models]);

  const modelNameMap = useMemo(() => new Map(models.map((item) => [item.id, `${item.displayName} (${item.model})`])), [models]);

  const loadData = async () => {
    setLoading(true);
    try {
      const [overviewRes, trendsRes, eventsRes, modelsRes, pricesRes, policiesRes] = await Promise.all([
        axios.get(`${BACKEND_URL}/api/admin/platform/usage/overview`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/admin/platform/usage/trends?days=7`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/admin/platform/usage/events?page=1&pageSize=10`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/admin/platform/llm-models`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/admin/platform/model-prices`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/admin/platform/usage-limit-policies`, { headers: getAuthHeaders() }),
      ]);
      setOverview(overviewRes.data || null);
      setTrends(Array.isArray(trendsRes.data) ? trendsRes.data : []);
      setEvents(Array.isArray(eventsRes.data?.items) ? eventsRes.data.items : []);
      setModels(Array.isArray(modelsRes.data) ? modelsRes.data : []);
      setPrices(Array.isArray(pricesRes.data) ? pricesRes.data : []);
      setPolicies(Array.isArray(policiesRes.data) ? policiesRes.data : []);
    } catch {
      messageApi.error(t('platform_usage_load_failed'));
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
        await axios.put(`${BACKEND_URL}/api/admin/platform/model-prices/${editingPrice.id}`, values, { headers: getAuthHeaders() });
        messageApi.success(t('platform_usage_price_updated'));
      } else {
        await axios.post(`${BACKEND_URL}/api/admin/platform/model-prices`, values, { headers: getAuthHeaders() });
        messageApi.success(t('platform_usage_price_created'));
      }
      setPriceModalOpen(false);
      await loadData();
    } catch (error: any) {
      messageApi.error(t('platform_usage_price_save_failed'));
    } finally {
      setSaving(false);
    }
  };

  const deletePrice = async (id: string) => {
    try {
      await axios.delete(`${BACKEND_URL}/api/admin/platform/model-prices/${id}`, { headers: getAuthHeaders() });
      messageApi.success(t('platform_usage_price_deleted'));
      await loadData();
    } catch {
      messageApi.error(t('platform_usage_price_delete_failed'));
    }
  };

  const openCreatePolicy = () => {
    setEditingPolicy(null);
    policyForm.setFieldsValue({
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
        await axios.put(`${BACKEND_URL}/api/admin/platform/usage-limit-policies/${editingPolicy.id}`, values, { headers: getAuthHeaders() });
        messageApi.success(t('platform_usage_policy_updated'));
      } else {
        await axios.post(`${BACKEND_URL}/api/admin/platform/usage-limit-policies`, values, { headers: getAuthHeaders() });
        messageApi.success(t('platform_usage_policy_created'));
      }
      setPolicyModalOpen(false);
      await loadData();
    } catch (error: any) {
      messageApi.error(t('platform_usage_policy_save_failed'));
    } finally {
      setSaving(false);
    }
  };

  const deletePolicy = async (id: string) => {
    try {
      await axios.delete(`${BACKEND_URL}/api/admin/platform/usage-limit-policies/${id}`, { headers: getAuthHeaders() });
      messageApi.success(t('platform_usage_policy_deleted'));
      await loadData();
    } catch {
      messageApi.error(t('platform_usage_policy_delete_failed'));
    }
  };

  return (
    <Space orientation="vertical" size={16} style={{ width: '100%' }}>
      {contextHolder}
      <Card variant="borderless" style={{ borderRadius: 20 }} loading={loading}>
        <Paragraph type="secondary" style={{ marginBottom: 16 }}>
          {t('platform_usage_summary_description')}
        </Paragraph>
        <Space size={16} wrap>
          <Statistic title={t('platform_usage_today_requests')} value={overview?.todayRequests || 0} />
          <Statistic title={t('platform_usage_today_tokens')} value={overview?.todayTokens || 0} />
          <Statistic title={t('platform_usage_today_charge')} value={overview?.todayCharge || 0} precision={4} prefix={<DollarOutlined />} />
          <Statistic title={t('platform_usage_month_charge')} value={overview?.monthCharge || 0} precision={4} prefix={<DollarOutlined />} />
        </Space>
      </Card>

      <Card
        variant="borderless"
        title={<Space><LineChartOutlined />{t('platform_usage_seven_day_trend')}</Space>}
        style={{ borderRadius: 20 }}
        loading={loading}
      >
        <Table
          rowKey="date"
          pagination={false}
          dataSource={trends}
          columns={[
            { title: t('platform_usage_date'), dataIndex: 'date', key: 'date', width: 120 },
            { title: t('platform_usage_request_count'), dataIndex: 'requestCount', key: 'requestCount', width: 100 },
            { title: t('platform_usage_tokens_label'), dataIndex: 'totalTokens', key: 'totalTokens', width: 140 },
            { title: t('platform_usage_cost'), dataIndex: 'costAmount', key: 'costAmount', width: 140, render: (value: number) => value.toFixed(4) },
            { title: t('platform_usage_charge'), dataIndex: 'chargeAmount', key: 'chargeAmount', width: 140, render: (value: number) => value.toFixed(4) },
          ]}
        />
      </Card>

      <Card variant="borderless" title={t('platform_usage_recent_events')} style={{ borderRadius: 20 }} loading={loading}>
        <Table
          rowKey="id"
          pagination={false}
          dataSource={events}
          scroll={{ x: 840 }}
          columns={[
            { title: t('platform_usage_time'), dataIndex: 'createdAt', key: 'createdAt', width: 180 },
            { title: t('platform_usage_model'), dataIndex: 'modelNameSnapshot', key: 'modelNameSnapshot', width: 220 },
            { title: t('platform_usage_source'), dataIndex: 'sourceType', key: 'sourceType', width: 100 },
            { title: t('platform_usage_status'), dataIndex: 'status', key: 'status', width: 110 },
            { title: t('platform_usage_tokens_label'), dataIndex: 'totalTokens', key: 'totalTokens', width: 120 },
            { title: t('platform_usage_charge'), dataIndex: 'chargeAmount', key: 'chargeAmount', width: 120, render: (value: number) => value.toFixed(4) },
          ]}
        />
      </Card>

      <Card
        variant="borderless"
        title={t('platform_usage_platform_model_pricing')}
        extra={<Button type="primary" icon={<PlusOutlined />} onClick={openCreatePrice}>{t('platform_usage_add_price')}</Button>}
        style={{ borderRadius: 20 }}
        loading={loading}
      >
        <Table
          rowKey="id"
          pagination={false}
          dataSource={prices}
          scroll={{ x: 980 }}
          columns={[
            {
              title: t('platform_usage_model'),
              dataIndex: 'modelId',
              key: 'modelId',
              render: (value: string) => modelNameMap.get(value) || value,
            },
            { title: t('platform_usage_currency'), dataIndex: 'currency', key: 'currency', width: 100 },
            { title: t('platform_usage_input_cost'), dataIndex: 'costInputUnitPrice', key: 'costInputUnitPrice', width: 140 },
            { title: t('platform_usage_output_cost'), dataIndex: 'costOutputUnitPrice', key: 'costOutputUnitPrice', width: 140 },
            { title: t('platform_usage_input_charge'), dataIndex: 'chargeInputUnitPrice', key: 'chargeInputUnitPrice', width: 140 },
            { title: t('platform_usage_output_charge'), dataIndex: 'chargeOutputUnitPrice', key: 'chargeOutputUnitPrice', width: 140 },
            {
              title: t('platform_usage_actions'),
              key: 'actions',
              width: 160,
              render: (_: unknown, item: ModelPrice) => (
                <Space>
                  <Button onClick={() => openEditPrice(item)}>{t('platform_usage_edit')}</Button>
                  <Button danger onClick={() => deletePrice(item.id)}>{t('platform_usage_delete')}</Button>
                </Space>
              ),
            },
          ]}
        />
      </Card>

      <Card
        variant="borderless"
        title={t('platform_usage_platform_limit_policy')}
        extra={<Button type="primary" icon={<PlusOutlined />} onClick={openCreatePolicy}>{t('platform_usage_add_policy')}</Button>}
        style={{ borderRadius: 20 }}
        loading={loading}
      >
        <Table
          rowKey="id"
          pagination={false}
          dataSource={policies}
          columns={[
            { title: t('platform_usage_enabled'), dataIndex: 'enabled', key: 'enabled', width: 90, render: (value: boolean) => value ? t('platform_usage_yes') : t('platform_usage_no') },
            { title: t('platform_usage_daily_token_limit'), dataIndex: 'dailyTokenLimit', key: 'dailyTokenLimit', width: 140 },
            { title: t('platform_usage_monthly_token_limit'), dataIndex: 'monthlyTokenLimit', key: 'monthlyTokenLimit', width: 140 },
            { title: t('platform_usage_daily_charge_limit'), dataIndex: 'dailyChargeLimit', key: 'dailyChargeLimit', width: 140 },
            { title: t('platform_usage_monthly_charge_limit'), dataIndex: 'monthlyChargeLimit', key: 'monthlyChargeLimit', width: 140 },
            { title: t('platform_usage_hard_limit'), dataIndex: 'hardLimit', key: 'hardLimit', width: 90, render: (value: boolean) => value ? t('platform_usage_yes') : t('platform_usage_no') },
            {
              title: t('platform_usage_actions'),
              key: 'actions',
              width: 160,
              render: (_: unknown, item: LimitPolicy) => (
                <Space>
                  <Button onClick={() => openEditPolicy(item)}>{t('platform_usage_edit')}</Button>
                  <Button danger onClick={() => deletePolicy(item.id)}>{t('platform_usage_delete')}</Button>
                </Space>
              ),
            },
          ]}
        />
      </Card>

      <Modal
        title={editingPrice ? t('platform_usage_edit_model_price') : t('platform_usage_create_model_price')}
        open={priceModalOpen}
        onOk={savePrice}
        onCancel={() => setPriceModalOpen(false)}
        confirmLoading={saving}
        destroyOnHidden
      >
        <Form form={priceForm} layout="vertical">
          <Form.Item label={t('platform_usage_model')} name="modelId" rules={[{ required: true, message: t('platform_usage_select_model') }]}>
            <Select options={modelOptions} />
          </Form.Item>
          <Form.Item label={t('platform_usage_currency')} name="currency">
            <Select options={[{ label: t('platform_usage_currency_usd'), value: 'USD' }, { label: t('platform_usage_currency_cny'), value: 'CNY' }]} />
          </Form.Item>
          <Form.Item label={`${t('platform_usage_input_cost')} ${t('platform_usage_per_million_tokens')}`} name="costInputUnitPrice">
            <InputNumber controls={false} style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label={`${t('platform_usage_output_cost')} ${t('platform_usage_per_million_tokens')}`} name="costOutputUnitPrice">
            <InputNumber controls={false} style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label={`${t('platform_usage_input_charge')} ${t('platform_usage_per_million_tokens')}`} name="chargeInputUnitPrice">
            <InputNumber controls={false} style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label={`${t('platform_usage_output_charge')} ${t('platform_usage_per_million_tokens')}`} name="chargeOutputUnitPrice">
            <InputNumber controls={false} style={{ width: '100%' }} min={0} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={editingPolicy ? t('platform_usage_edit_limit_policy') : t('platform_usage_create_limit_policy')}
        open={policyModalOpen}
        onOk={savePolicy}
        onCancel={() => setPolicyModalOpen(false)}
        confirmLoading={saving}
        destroyOnHidden
      >
        <Form form={policyForm} layout="vertical">
          <Form.Item label={t('platform_usage_enabled')} name="enabled" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item label={t('platform_usage_daily_token_limit')} name="dailyTokenLimit">
            <InputNumber controls={false} style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label={t('platform_usage_monthly_token_limit')} name="monthlyTokenLimit">
            <InputNumber controls={false} style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label={t('platform_usage_daily_charge_limit')} name="dailyChargeLimit">
            <InputNumber controls={false} style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label={t('platform_usage_monthly_charge_limit')} name="monthlyChargeLimit">
            <InputNumber controls={false} style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label={t('platform_usage_hard_limit')} name="hardLimit" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
};

export default PlatformUsageSettingsCard;


