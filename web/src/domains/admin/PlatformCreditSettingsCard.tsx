import React, { useEffect, useMemo, useState } from 'react';
import { Button, Card, DatePicker, Form, Input, InputNumber, Modal, Select, Space, Statistic, Table, Tag, Typography, message } from 'antd';
import { DollarOutlined, PlusOutlined } from '@ant-design/icons';
import axios from 'axios';
import type { Dayjs } from 'dayjs';
import { useTranslation } from 'react-i18next';
import { BACKEND_URL } from '../../config';
import { casdoorService } from '../identity/CasdoorService';

const { Paragraph } = Typography;

interface PlatformModel {
  id: string;
  displayName: string;
  model: string;
}

interface EnterpriseMembership {
  enterpriseId: string;
  name: string;
  role: string;
}

interface CreditOverviewItem {
  creditType: string;
  totalCredits: number;
  reservedCredits: number;
  availableCredits: number;
  grantedCredits: number;
  settledCredits: number;
  expiredCredits: number;
}

interface CreditPriceBook {
  id: string;
  enterpriseId: string;
  creditType: string;
  modelId: string;
  fundingType: string;
  modelSourceType: string;
  currency: string;
  costInputUsdPer1M: number;
  costOutputUsdPer1M: number;
  platformMultiplier: number;
  enterpriseMultiplier: number;
  inputCreditsPer1M: number;
  outputCreditsPer1M: number;
  effectiveAt: string;
  status: string;
}

interface GrantFormValues {
  enterpriseId: string;
  creditType: string;
  sourceType: string;
  credits: number;
  reasonCode?: string;
  effectiveAt?: Dayjs;
  expiresAt?: Dayjs;
  metadataJson?: string;
}

interface PriceBookFormValues {
  enterpriseId?: string;
  modelId: string;
  currency: string;
  costInputUsdPer1M: number;
  costOutputUsdPer1M: number;
  platformMultiplier: number;
  enterpriseMultiplier: number;
  effectiveAt?: Dayjs;
  status: string;
}

function getAuthHeaders() {
  const token = casdoorService.getToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

function formatDateTime(value?: string, locale = 'en') {
  if (!value) {
    return '-';
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return '-';
  }
  return new Intl.DateTimeFormat(locale.toLowerCase().startsWith('zh') ? 'zh-CN' : 'en-US', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date);
}

function getCreditTypeLabel(value: string, t: (key: string) => string) {
  return t(value === 'platform' ? 'credit_type_platform' : 'credit_type_enterprise');
}

function getFundingTypeLabel(value: string, t: (key: string) => string) {
  switch (value) {
    case 'platform_funded':
      return t('credit_funding_platform');
    case 'enterprise_funded':
      return t('credit_funding_enterprise');
    default:
      return value;
  }
}

function getStatusLabel(value: string, t: (key: string) => string) {
  switch (value) {
    case 'active':
      return t('credit_status_active');
    case 'disabled':
      return t('credit_status_disabled');
    default:
      return value;
  }
}

const PlatformCreditSettingsCard: React.FC = () => {
  const { t, i18n } = useTranslation();
  const [messageApi, contextHolder] = message.useMessage();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [enterpriseId, setEnterpriseId] = useState('');
  const [enterprises, setEnterprises] = useState<EnterpriseMembership[]>([]);
  const [overview, setOverview] = useState<CreditOverviewItem[]>([]);
  const [models, setModels] = useState<PlatformModel[]>([]);
  const [priceBooks, setPriceBooks] = useState<CreditPriceBook[]>([]);
  const [grantModalOpen, setGrantModalOpen] = useState(false);
  const [priceBookModalOpen, setPriceBookModalOpen] = useState(false);
  const [editingPriceBook, setEditingPriceBook] = useState<CreditPriceBook | null>(null);
  const [grantForm] = Form.useForm<GrantFormValues>();
  const [priceBookForm] = Form.useForm<PriceBookFormValues>();

  const modelNameMap = useMemo(() => new Map(models.map((item) => [item.id, `${item.displayName} (${item.model})`])), [models]);
  const enterpriseNameMap = useMemo(() => new Map(enterprises.map((item) => [item.enterpriseId, item.name])), [enterprises]);
  const globalPriceBooks = useMemo(() => priceBooks.filter((item) => !item.enterpriseId), [priceBooks]);
  const enterprisePriceBooks = useMemo(() => priceBooks.filter((item) => !!item.enterpriseId), [priceBooks]);
  const currentLocale = i18n.resolvedLanguage || i18n.language || 'en';
  const formatDisplayDateTime = (value?: string) => formatDateTime(value, currentLocale);
  const requiredMessage = (label: string) => t('credit_field_required', { field: label });
  const creditTypeOptions = [
    { label: t('credit_type_platform'), value: 'platform' },
    { label: t('credit_type_enterprise'), value: 'enterprise' },
  ];
  const grantSourceOptions = [
    { label: t('credit_source_type_manual_adjust'), value: 'manual_adjust' },
    { label: t('credit_source_type_trial_grant'), value: 'trial_grant' },
    { label: t('credit_source_type_subscription_grant'), value: 'subscription_grant' },
    { label: t('credit_source_type_topup_grant'), value: 'topup_grant' },
  ];
  const statusOptions = [
    { label: t('credit_status_active'), value: 'active' },
    { label: t('credit_status_disabled'), value: 'disabled' },
  ];

  const loadStaticData = async () => {
    try {
      const [modelsRes, priceBooksRes, enterprisesRes] = await Promise.all([
        axios.get(`${BACKEND_URL}/api/admin/platform/llm-models`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/admin/platform/credit-price-books`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/enterprises`, { headers: getAuthHeaders() }),
      ]);
      setModels(Array.isArray(modelsRes.data) ? modelsRes.data : []);
      setPriceBooks(Array.isArray(priceBooksRes.data) ? priceBooksRes.data : []);
      setEnterprises(Array.isArray(enterprisesRes.data) ? enterprisesRes.data : []);
    } catch {
      messageApi.error(t('platform_credit_config_failed'));
    }
  };

  const loadOverview = async (nextEnterpriseId?: string) => {
    const targetEnterpriseId = (nextEnterpriseId ?? enterpriseId).trim();
    if (!targetEnterpriseId) {
      setOverview([]);
      return;
    }
    try {
      const overviewRes = await axios.get(`${BACKEND_URL}/api/admin/platform/credits/overview`, {
        headers: getAuthHeaders(),
        params: { enterpriseId: targetEnterpriseId },
      });
      setOverview(Array.isArray(overviewRes.data?.wallets) ? overviewRes.data.wallets : []);
    } catch {
      messageApi.error(t('platform_credit_overview_failed'));
    }
  };

  const loadData = async () => {
    setLoading(true);
    try {
      await Promise.all([loadStaticData(), loadOverview()]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const openCreateGrant = () => {
    grantForm.setFieldsValue({
      enterpriseId,
      creditType: 'platform',
      sourceType: 'manual_adjust',
      credits: 0,
      metadataJson: '{}',
    });
    setGrantModalOpen(true);
  };

  const openCreatePriceBook = () => {
    setEditingPriceBook(null);
    priceBookForm.setFieldsValue({
      modelId: models[0]?.id,
      currency: 'USD',
      costInputUsdPer1M: 0,
      costOutputUsdPer1M: 0,
      platformMultiplier: 1,
      enterpriseMultiplier: 1,
      enterpriseId: undefined,
      status: 'active',
    });
    setPriceBookModalOpen(true);
  };

  const openEditPriceBook = (item: CreditPriceBook) => {
    setEditingPriceBook(item);
    priceBookForm.setFieldsValue({
      modelId: item.modelId,
      currency: item.currency || 'USD',
      costInputUsdPer1M: item.costInputUsdPer1M,
      costOutputUsdPer1M: item.costOutputUsdPer1M,
      platformMultiplier: item.platformMultiplier,
      enterpriseMultiplier: item.enterpriseMultiplier,
      enterpriseId: item.enterpriseId || undefined,
      status: item.status,
    });
    setPriceBookModalOpen(true);
  };

  const saveGrant = async () => {
    const values = await grantForm.validateFields();
    setSaving(true);
    try {
      await axios.post(`${BACKEND_URL}/api/admin/platform/credits/grants`, {
        enterpriseId: values.enterpriseId,
        creditType: values.creditType,
        sourceType: values.sourceType,
        credits: values.credits,
        reasonCode: values.reasonCode,
        metadataJson: values.metadataJson,
        effectiveAt: values.effectiveAt?.toISOString(),
        expiresAt: values.expiresAt?.toISOString(),
      }, { headers: getAuthHeaders() });
      messageApi.success(t('platform_credit_grant_created'));
      setGrantModalOpen(false);
      setEnterpriseId(values.enterpriseId);
      await loadOverview(values.enterpriseId);
    } catch {
      messageApi.error(t('platform_credit_grant_failed'));
    } finally {
      setSaving(false);
    }
  };

  const savePriceBook = async () => {
    const values = await priceBookForm.validateFields();
    setSaving(true);
    try {
      const payload = {
        creditType: 'platform',
        modelId: values.modelId,
        modelScope: 'platform',
        modelSourceType: 'platform_model',
        fundingType: 'platform_funded',
        currency: values.currency,
        creditUnitUsd: 0.0001,
        enterpriseId: values.enterpriseId || '',
        costInputUsdPer1M: values.costInputUsdPer1M,
        costOutputUsdPer1M: values.costOutputUsdPer1M,
        platformMultiplier: values.platformMultiplier,
        enterpriseMultiplier: values.enterpriseMultiplier,
        effectiveAt: values.effectiveAt?.toISOString(),
        status: values.status,
      };
      if (editingPriceBook) {
        await axios.put(`${BACKEND_URL}/api/admin/platform/credit-price-books/${editingPriceBook.id}`, payload, { headers: getAuthHeaders() });
        messageApi.success(t('platform_credit_price_book_updated'));
      } else {
        await axios.post(`${BACKEND_URL}/api/admin/platform/credit-price-books`, payload, { headers: getAuthHeaders() });
        messageApi.success(t('platform_credit_price_book_created'));
      }
      setPriceBookModalOpen(false);
      await loadStaticData();
    } catch {
      messageApi.error(t('platform_credit_price_book_save_failed'));
    } finally {
      setSaving(false);
    }
  };

  const deletePriceBook = async (id: string) => {
    const target = priceBooks.find((item) => item.id === id);
    try {
      await axios.delete(`${BACKEND_URL}/api/admin/platform/credit-price-books/${id}`, {
        headers: getAuthHeaders(),
        params: { enterpriseId: target?.enterpriseId || '' },
      });
      messageApi.success(t('platform_credit_price_book_deleted'));
      await loadStaticData();
    } catch {
      messageApi.error(t('platform_credit_price_book_delete_failed'));
    }
  };

  return (
    <Space orientation="vertical" size={16} style={{ width: '100%' }}>
      {contextHolder}

      <Card variant="borderless" style={{ borderRadius: 20 }} loading={loading}>
        <Paragraph type="secondary" style={{ marginBottom: 16 }}>
          {t('platform_credit_desc')}
        </Paragraph>
        <Space.Compact style={{ width: '100%', marginBottom: 16 }}>
          <Select
            showSearch
            allowClear
            placeholder={t('platform_credit_target_enterprise')}
            value={enterpriseId}
            onChange={(value) => setEnterpriseId(value || '')}
            options={enterprises.map((item) => ({
              label: `${item.name} (${item.enterpriseId})`,
              value: item.enterpriseId,
            }))}
            optionFilterProp="label"
            style={{ width: '100%' }}
          />
          <Button onClick={() => loadOverview()}>{t('platform_credit_load_overview')}</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={openCreateGrant}>{t('platform_credit_new_grant')}</Button>
        </Space.Compact>
        <Space size={16} wrap>
          {overview.map((item) => (
            <Statistic
              key={item.creditType}
              title={getCreditTypeLabel(item.creditType, t)}
              value={item.availableCredits || 0}
              prefix={<DollarOutlined />}
              suffix={t('credit_available_suffix')}
            />
          ))}
        </Space>
      </Card>

      <Card
        variant="borderless"
        title={t('platform_credit_global_price_books')}
        extra={<Button type="primary" icon={<PlusOutlined />} onClick={openCreatePriceBook}>{t('platform_credit_new_price_book')}</Button>}
        style={{ borderRadius: 20 }}
        loading={loading}
      >
        <Table
          rowKey="id"
          pagination={false}
          dataSource={globalPriceBooks}
          scroll={{ x: 1180 }}
          columns={[
            { title: t('credit_credit_type_label'), dataIndex: 'creditType', key: 'creditType', render: (value: string) => <Tag color="blue">{getCreditTypeLabel(value, t)}</Tag> },
            { title: t('credit_model_label'), dataIndex: 'modelId', key: 'modelId', render: (value: string) => modelNameMap.get(value) || value },
            { title: t('credit_funding_label'), dataIndex: 'fundingType', key: 'fundingType', render: (value: string) => getFundingTypeLabel(value, t) },
            { title: t('credit_source_label'), dataIndex: 'modelSourceType', key: 'modelSourceType' },
            { title: t('credit_currency_label'), dataIndex: 'currency', key: 'currency' },
            { title: t('credit_input_price_per_m_label'), dataIndex: 'costInputUsdPer1M', key: 'costInputUsdPer1M' },
            { title: t('credit_output_price_per_m_label'), dataIndex: 'costOutputUsdPer1M', key: 'costOutputUsdPer1M' },
            { title: t('credit_platform_multiplier_label'), dataIndex: 'platformMultiplier', key: 'platformMultiplier' },
            { title: t('credit_enterprise_multiplier_label'), dataIndex: 'enterpriseMultiplier', key: 'enterpriseMultiplier' },
            { title: t('credit_input_credits_per_m_label'), dataIndex: 'inputCreditsPer1M', key: 'inputCreditsPer1M' },
            { title: t('credit_output_credits_per_m_label'), dataIndex: 'outputCreditsPer1M', key: 'outputCreditsPer1M' },
            { title: t('credit_status_label'), dataIndex: 'status', key: 'status' },
            { title: t('credit_effective_at_label'), dataIndex: 'effectiveAt', key: 'effectiveAt', render: formatDisplayDateTime },
            {
              title: t('credit_actions_label'),
              key: 'actions',
              render: (_: unknown, item: CreditPriceBook) => (
                <Space>
                  <Button onClick={() => openEditPriceBook(item)}>{t('credit_edit_action')}</Button>
                  <Button danger onClick={() => deletePriceBook(item.id)}>{t('credit_delete_action')}</Button>
                </Space>
              ),
            },
          ]}
        />
      </Card>

      <Card
        variant="borderless"
        title={t('platform_credit_enterprise_overrides')}
        style={{ borderRadius: 20 }}
        loading={loading}
      >
        <Table
          rowKey="id"
          pagination={false}
          dataSource={enterprisePriceBooks}
          scroll={{ x: 1280 }}
          columns={[
            { title: t('credit_enterprise_label'), dataIndex: 'enterpriseId', key: 'enterpriseId', render: (value: string) => enterpriseNameMap.get(value) || value },
            { title: t('credit_model_label'), dataIndex: 'modelId', key: 'modelId', render: (value: string) => modelNameMap.get(value) || value },
            { title: t('credit_currency_label'), dataIndex: 'currency', key: 'currency' },
            { title: t('credit_platform_multiplier_label'), dataIndex: 'platformMultiplier', key: 'platformMultiplier' },
            { title: t('credit_enterprise_multiplier_label'), dataIndex: 'enterpriseMultiplier', key: 'enterpriseMultiplier' },
            { title: t('credit_input_credits_per_m_label'), dataIndex: 'inputCreditsPer1M', key: 'inputCreditsPer1M' },
            { title: t('credit_output_credits_per_m_label'), dataIndex: 'outputCreditsPer1M', key: 'outputCreditsPer1M' },
            { title: t('credit_status_label'), dataIndex: 'status', key: 'status', render: (value: string) => getStatusLabel(value, t) },
            { title: t('credit_effective_at_label'), dataIndex: 'effectiveAt', key: 'effectiveAt', render: formatDisplayDateTime },
            {
              title: t('credit_actions_label'),
              key: 'actions',
              render: (_: unknown, item: CreditPriceBook) => (
                <Space>
                  <Button onClick={() => openEditPriceBook(item)}>{t('credit_edit_action')}</Button>
                  <Button danger onClick={() => deletePriceBook(item.id)}>{t('credit_delete_action')}</Button>
                </Space>
              ),
            },
          ]}
        />
      </Card>

      <Modal
        title={t('platform_credit_create_grant')}
        open={grantModalOpen}
        onOk={saveGrant}
        onCancel={() => setGrantModalOpen(false)}
        confirmLoading={saving}
        destroyOnHidden
      >
        <Form form={grantForm} layout="vertical">
          <Form.Item label={t('platform_credit_target_enterprise')} name="enterpriseId" rules={[{ required: true, message: requiredMessage(t('platform_credit_target_enterprise')) }]}>
            <Select
              showSearch
              options={enterprises.map((item) => ({
                label: `${item.name} (${item.enterpriseId})`,
                value: item.enterpriseId,
              }))}
              optionFilterProp="label"
              placeholder={t('platform_credit_target_enterprise')}
            />
          </Form.Item>
          <Form.Item label={t('credit_credit_type_label')} name="creditType" rules={[{ required: true, message: requiredMessage(t('credit_credit_type_label')) }]}>
            <Select options={creditTypeOptions} />
          </Form.Item>
          <Form.Item label={t('credit_source_type_label')} name="sourceType" rules={[{ required: true, message: requiredMessage(t('credit_source_type_label')) }]}>
            <Select options={grantSourceOptions} />
          </Form.Item>
          <Form.Item label={t('credit_credits_label')} name="credits" rules={[{ required: true, message: requiredMessage(t('credit_credits_label')) }]}>
            <InputNumber controls={false} style={{ width: '100%' }} min={1} />
          </Form.Item>
          <Form.Item label={t('credit_reason_code_label')} name="reasonCode">
            <Input />
          </Form.Item>
          <Form.Item label={t('credit_effective_at_label')} name="effectiveAt">
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label={t('credit_expires_at_label')} name="expiresAt">
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label={t('credit_metadata_json_label')} name="metadataJson">
            <Input.TextArea rows={4} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={editingPriceBook ? t('platform_credit_edit_price_book') : t('platform_credit_create_price_book')}
        open={priceBookModalOpen}
        onOk={savePriceBook}
        onCancel={() => setPriceBookModalOpen(false)}
        confirmLoading={saving}
        destroyOnHidden
      >
        <Form form={priceBookForm} layout="vertical">
          <Form.Item label={t('platform_credit_override_scope')} name="enterpriseId">
            <Select
              allowClear
              showSearch
              placeholder={t('platform_credit_override_scope_placeholder')}
              options={enterprises.map((item) => ({
                label: `${item.name} (${item.enterpriseId})`,
                value: item.enterpriseId,
              }))}
              optionFilterProp="label"
            />
          </Form.Item>
          <Form.Item label={t('credit_model_label')} name="modelId" rules={[{ required: true, message: requiredMessage(t('credit_model_label')) }]}>
            <Select options={models.map((item) => ({ label: `${item.displayName} (${item.model})`, value: item.id }))} />
          </Form.Item>
          <Form.Item label={t('credit_currency_label')} name="currency" rules={[{ required: true, message: requiredMessage(t('credit_currency_label')) }]}>
            <Select options={[{ label: 'USD', value: 'USD' }, { label: 'CNY', value: 'CNY' }]} />
          </Form.Item>
          <Form.Item label={t('credit_input_price_per_m_tokens_label')} name="costInputUsdPer1M">
            <InputNumber controls={false} style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label={t('credit_output_price_per_m_tokens_label')} name="costOutputUsdPer1M">
            <InputNumber controls={false} style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label={t('credit_platform_multiplier_label')} name="platformMultiplier">
            <InputNumber controls={false} style={{ width: '100%' }} min={1} step={0.1} />
          </Form.Item>
          <Form.Item label={t('credit_enterprise_multiplier_label')} name="enterpriseMultiplier">
            <InputNumber controls={false} style={{ width: '100%' }} min={1} step={0.1} />
          </Form.Item>
          <Form.Item label={t('credit_effective_at_label')} name="effectiveAt">
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label={t('credit_status_label')} name="status">
            <Select options={statusOptions} />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
};

export default PlatformCreditSettingsCard;
