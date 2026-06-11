import React, { useEffect, useMemo, useState } from 'react';
import { Button, Card, DatePicker, Form, Input, InputNumber, Modal, Select, Space, Statistic, Switch, Table, Tag, Typography, message } from 'antd';
import { DollarOutlined, PlusOutlined } from '@ant-design/icons';
import axios from 'axios';
import dayjs, { Dayjs } from 'dayjs';
import { useTranslation } from 'react-i18next';
import { BACKEND_URL } from '../../config';
import { casdoorService } from '../identity/CasdoorService';

const { Paragraph } = Typography;

interface CreditOverviewItem {
  creditType: string;
  totalCredits: number;
  reservedCredits: number;
  availableCredits: number;
  grantedCredits: number;
  settledCredits: number;
  expiredCredits: number;
}

interface CreditOverview {
  enterpriseId: string;
  wallets: CreditOverviewItem[];
}

interface CreditWallet {
  id: string;
  enterpriseId: string;
  creditType: string;
  totalCredits: number;
  reservedCredits: number;
  availableCredits: number;
}

interface CreditGrant {
  id: string;
  creditType: string;
  sourceType: string;
  grantedCredits: number;
  remainingCredits: number;
  effectiveAt: string;
  expiresAt?: string;
  createdAt: string;
}

interface CreditLedgerEntry {
  id: string;
  creditType: string;
  entryType: string;
  direction: string;
  credits: number;
  balanceAfter: number;
  reservedAfter: number;
  memberUserId?: string;
  agentId?: string;
  createdAt: string;
  reasonCode: string;
}

interface EnterpriseModel {
  id: string;
  displayName: string;
  model: string;
  fundingType: string;
  modelSourceType: string;
}

interface CreditPriceBook {
  id: string;
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

interface CreditBudgetPolicy {
  id: string;
  creditType: string;
  scopeType: string;
  scopeId: string;
  enabled: boolean;
  dailyCreditLimit: number;
  monthlyCreditLimit: number;
  dailyTokenLimit: number;
  monthlyTokenLimit: number;
  dailyUsdLimit: number;
  monthlyUsdLimit: number;
  hardLimit: boolean;
}

interface AgentItem {
  id: string;
  agentName: string;
}

interface MemberItem {
  userId: string;
  email: string;
  displayName: string;
}

interface GrantFormValues {
  creditType: string;
  sourceType: string;
  credits: number;
  reasonCode?: string;
  effectiveAt?: Dayjs;
  expiresAt?: Dayjs;
  metadataJson?: string;
}

interface PriceBookFormValues {
  creditType: string;
  modelId: string;
  fundingType: string;
  modelSourceType: string;
  modelScope: string;
  currency: string;
  costInputUsdPer1M: number;
  costOutputUsdPer1M: number;
  enterpriseMultiplier: number;
  effectiveAt?: Dayjs;
  status: string;
}

interface PolicyFormValues {
  creditType: string;
  scopeType: string;
  scopeId?: string;
  enabled: boolean;
  dailyCreditLimit: number;
  monthlyCreditLimit: number;
  dailyTokenLimit: number;
  monthlyTokenLimit: number;
  dailyUsdLimit: number;
  monthlyUsdLimit: number;
  hardLimit: boolean;
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

function getGrantSourceTypeLabel(value: string, t: (key: string) => string) {
  switch (value) {
    case 'manual_adjust':
      return t('credit_source_type_manual_adjust');
    case 'trial_grant':
      return t('credit_source_type_trial_grant');
    case 'subscription_grant':
      return t('credit_source_type_subscription_grant');
    case 'topup_grant':
      return t('credit_source_type_topup_grant');
    default:
      return value;
  }
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

function getScopeTypeLabel(value: string, t: (key: string) => string) {
  switch (value) {
    case 'enterprise':
      return t('credit_scope_enterprise');
    case 'member':
      return t('credit_scope_member');
    case 'agent':
      return t('credit_scope_agent');
    default:
      return value;
  }
}

function getEntryTypeLabel(value: string, t: (key: string) => string) {
  switch (value) {
    case 'grant':
      return t('credit_entry_type_grant');
    case 'reserve':
      return t('credit_entry_type_reserve');
    case 'settle':
      return t('credit_entry_type_settle');
    case 'release':
      return t('credit_entry_type_release');
    case 'expire':
      return t('credit_entry_type_expire');
    default:
      return value;
  }
}

function getDirectionLabel(value: string, t: (key: string) => string) {
  switch (value) {
    case 'credit':
      return t('credit_direction_credit');
    case 'debit':
      return t('credit_direction_debit');
    default:
      return value;
  }
}

const EnterpriseCreditSettingsTab: React.FC = () => {
  const { t, i18n } = useTranslation();
  const [messageApi, contextHolder] = message.useMessage();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [overview, setOverview] = useState<CreditOverview | null>(null);
  const [wallets, setWallets] = useState<CreditWallet[]>([]);
  const [grants, setGrants] = useState<CreditGrant[]>([]);
  const [ledger, setLedger] = useState<CreditLedgerEntry[]>([]);
  const [models, setModels] = useState<EnterpriseModel[]>([]);
  const [priceBooks, setPriceBooks] = useState<CreditPriceBook[]>([]);
  const [policies, setPolicies] = useState<CreditBudgetPolicy[]>([]);
  const [agents, setAgents] = useState<AgentItem[]>([]);
  const [members, setMembers] = useState<MemberItem[]>([]);
  const [grantModalOpen, setGrantModalOpen] = useState(false);
  const [priceBookModalOpen, setPriceBookModalOpen] = useState(false);
  const [policyModalOpen, setPolicyModalOpen] = useState(false);
  const [editingPriceBook, setEditingPriceBook] = useState<CreditPriceBook | null>(null);
  const [editingPolicy, setEditingPolicy] = useState<CreditBudgetPolicy | null>(null);
  const [grantForm] = Form.useForm<GrantFormValues>();
  const [priceBookForm] = Form.useForm<PriceBookFormValues>();
  const [policyForm] = Form.useForm<PolicyFormValues>();

  const modelNameMap = useMemo(() => new Map(models.map((item) => [item.id, `${item.displayName} (${item.model})`])), [models]);
  const agentNameMap = useMemo(() => new Map(agents.map((item) => [item.id, item.agentName])), [agents]);
  const memberNameMap = useMemo(() => new Map(members.map((item) => [item.userId, item.displayName || item.email])), [members]);
  const enterpriseFundedModels = useMemo(() => models.filter((item) => item.fundingType === 'enterprise_funded'), [models]);
  const platformFundedModels = useMemo(() => models.filter((item) => item.fundingType === 'platform_funded'), [models]);
  const enterprisePriceBooks = useMemo(() => priceBooks.filter((item) => item.creditType === 'enterprise'), [priceBooks]);
  const platformOverridePriceBooks = useMemo(() => priceBooks.filter((item) => item.creditType === 'platform'), [priceBooks]);
  const currentLocale = i18n.resolvedLanguage || i18n.language || 'en';
  const formatDisplayDateTime = (value?: string) => formatDateTime(value, currentLocale);
  const requiredMessage = (label: string) => t('credit_field_required', { field: label });
  const creditTypeOptions = [
    { label: t('credit_type_enterprise'), value: 'enterprise' },
    { label: t('credit_type_platform'), value: 'platform' },
  ];
  const grantSourceOptions = [
    { label: t('credit_source_type_manual_adjust'), value: 'manual_adjust' },
    { label: t('credit_source_type_trial_grant'), value: 'trial_grant' },
    { label: t('credit_source_type_subscription_grant'), value: 'subscription_grant' },
    { label: t('credit_source_type_topup_grant'), value: 'topup_grant' },
  ];
  const scopeTypeOptions = [
    { label: t('credit_scope_enterprise'), value: 'enterprise' },
    { label: t('credit_scope_member'), value: 'member' },
    { label: t('credit_scope_agent'), value: 'agent' },
  ];
  const statusOptions = [
    { label: t('credit_status_active'), value: 'active' },
    { label: t('credit_status_disabled'), value: 'disabled' },
  ];

  const loadData = async () => {
    setLoading(true);
    try {
      const [
        overviewRes,
        walletsRes,
        grantsRes,
        ledgerRes,
        modelsRes,
        priceBooksRes,
        enterprisePoliciesRes,
        memberPoliciesRes,
        agentPoliciesRes,
        agentsRes,
        membersRes,
      ] = await Promise.all([
        axios.get(`${BACKEND_URL}/api/admin/credits/overview`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/admin/credits/wallets`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/admin/credits/grants`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/admin/credits/ledger`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/admin/llm-models`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/admin/credit-price-books`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/admin/credit-budget-policies?scopeType=enterprise`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/admin/credit-budget-policies?scopeType=member`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/admin/credit-budget-policies?scopeType=agent`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/agents`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/admin/members`, { headers: getAuthHeaders() }),
      ]);
      setOverview(overviewRes.data || null);
      setWallets(Array.isArray(walletsRes.data) ? walletsRes.data : []);
      setGrants(Array.isArray(grantsRes.data) ? grantsRes.data : []);
      setLedger(Array.isArray(ledgerRes.data) ? ledgerRes.data : []);
      setModels(Array.isArray(modelsRes.data) ? modelsRes.data : []);
      setPriceBooks(Array.isArray(priceBooksRes.data) ? priceBooksRes.data : []);
      setPolicies([
        ...(Array.isArray(enterprisePoliciesRes.data) ? enterprisePoliciesRes.data : []),
        ...(Array.isArray(memberPoliciesRes.data) ? memberPoliciesRes.data : []),
        ...(Array.isArray(agentPoliciesRes.data) ? agentPoliciesRes.data : []),
      ]);
      setAgents(Array.isArray(agentsRes.data) ? agentsRes.data : []);
      setMembers(Array.isArray(membersRes.data) ? membersRes.data : []);
    } catch {
      messageApi.error(t('enterprise_credit_load_failed'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const openCreateGrant = () => {
    grantForm.setFieldsValue({
      creditType: 'enterprise',
      sourceType: 'manual_adjust',
      credits: 0,
      metadataJson: '{}',
    });
    setGrantModalOpen(true);
  };

  const openCreatePriceBook = () => {
    setEditingPriceBook(null);
    priceBookForm.setFieldsValue({
      creditType: 'enterprise',
      modelId: enterpriseFundedModels[0]?.id,
      fundingType: 'enterprise_funded',
      modelSourceType: 'enterprise_custom_model',
      modelScope: 'enterprise',
      currency: 'USD',
      costInputUsdPer1M: 0,
      costOutputUsdPer1M: 0,
      enterpriseMultiplier: 1,
      status: 'active',
    });
    setPriceBookModalOpen(true);
  };

  const openEditPriceBook = (item: CreditPriceBook) => {
    setEditingPriceBook(item);
    priceBookForm.setFieldsValue({
      creditType: item.creditType,
      modelId: item.modelId,
      fundingType: item.fundingType,
      modelSourceType: item.modelSourceType,
      modelScope: item.fundingType === 'platform_funded' ? 'platform' : 'enterprise',
      currency: item.currency || 'USD',
      costInputUsdPer1M: item.costInputUsdPer1M,
      costOutputUsdPer1M: item.costOutputUsdPer1M,
      enterpriseMultiplier: item.enterpriseMultiplier,
      effectiveAt: item.effectiveAt ? dayjs(item.effectiveAt) : undefined,
      status: item.status,
    });
    setPriceBookModalOpen(true);
  };

  const openCreatePlatformOverride = () => {
    setEditingPriceBook(null);
    priceBookForm.setFieldsValue({
      creditType: 'platform',
      modelId: platformFundedModels[0]?.id,
      fundingType: 'platform_funded',
      modelSourceType: 'platform_model',
      modelScope: 'platform',
      currency: 'USD',
      costInputUsdPer1M: 0,
      costOutputUsdPer1M: 0,
      enterpriseMultiplier: 1,
      status: 'active',
    });
    setPriceBookModalOpen(true);
  };

  const handlePriceBookCreditTypeChange = (value: string) => {
    const nextIsPlatform = value === 'platform';
    const nextModel = nextIsPlatform ? platformFundedModels[0]?.id : enterpriseFundedModels[0]?.id;
    priceBookForm.setFieldsValue({
      modelId: nextModel,
      fundingType: nextIsPlatform ? 'platform_funded' : 'enterprise_funded',
      modelSourceType: nextIsPlatform ? 'platform_model' : 'enterprise_custom_model',
      modelScope: nextIsPlatform ? 'platform' : 'enterprise',
    });
  };

  const openCreatePolicy = () => {
    setEditingPolicy(null);
    policyForm.setFieldsValue({
      creditType: 'enterprise',
      scopeType: 'enterprise',
      enabled: true,
      dailyCreditLimit: 0,
      monthlyCreditLimit: 0,
      dailyTokenLimit: 0,
      monthlyTokenLimit: 0,
      dailyUsdLimit: 0,
      monthlyUsdLimit: 0,
      hardLimit: true,
    });
    setPolicyModalOpen(true);
  };

  const openEditPolicy = (item: CreditBudgetPolicy) => {
    setEditingPolicy(item);
    policyForm.setFieldsValue(item);
    setPolicyModalOpen(true);
  };

  const saveGrant = async () => {
    const values = await grantForm.validateFields();
    setSaving(true);
    try {
      await axios.post(`${BACKEND_URL}/api/admin/credits/grants`, {
        creditType: values.creditType,
        sourceType: values.sourceType,
        credits: values.credits,
        reasonCode: values.reasonCode,
        metadataJson: values.metadataJson,
        effectiveAt: values.effectiveAt?.toISOString(),
        expiresAt: values.expiresAt?.toISOString(),
      }, { headers: getAuthHeaders() });
      messageApi.success(t('enterprise_credit_grant_created'));
      setGrantModalOpen(false);
      await loadData();
    } catch {
      messageApi.error(t('enterprise_credit_grant_failed'));
    } finally {
      setSaving(false);
    }
  };

  const savePriceBook = async () => {
    const values = await priceBookForm.validateFields();
    setSaving(true);
    try {
      const payload = {
        creditType: values.creditType,
        modelId: values.modelId,
        modelScope: values.modelScope,
        modelSourceType: values.modelSourceType,
        fundingType: values.fundingType,
        currency: values.currency,
        creditUnitUsd: 0.0001,
        costInputUsdPer1M: values.costInputUsdPer1M,
        costOutputUsdPer1M: values.costOutputUsdPer1M,
        platformMultiplier: 1,
        enterpriseMultiplier: values.enterpriseMultiplier,
        effectiveAt: values.effectiveAt?.toISOString(),
        status: values.status,
      };
      if (editingPriceBook) {
        await axios.put(`${BACKEND_URL}/api/admin/credit-price-books/${editingPriceBook.id}`, payload, { headers: getAuthHeaders() });
        messageApi.success(t('enterprise_credit_price_book_updated'));
      } else {
        await axios.post(`${BACKEND_URL}/api/admin/credit-price-books`, payload, { headers: getAuthHeaders() });
        messageApi.success(t('enterprise_credit_price_book_created'));
      }
      setPriceBookModalOpen(false);
      await loadData();
    } catch {
      messageApi.error(t('enterprise_credit_price_book_save_failed'));
    } finally {
      setSaving(false);
    }
  };

  const deletePriceBook = async (id: string) => {
    try {
      await axios.delete(`${BACKEND_URL}/api/admin/credit-price-books/${id}`, { headers: getAuthHeaders() });
      messageApi.success(t('enterprise_credit_price_book_deleted'));
      await loadData();
    } catch {
      messageApi.error(t('enterprise_credit_price_book_delete_failed'));
    }
  };

  const savePolicy = async () => {
    const values = await policyForm.validateFields();
    setSaving(true);
    try {
      const payload = {
        ...values,
        scopeId: values.scopeType === 'enterprise' ? 'enterprise' : values.scopeId,
      };
      if (editingPolicy) {
        await axios.put(`${BACKEND_URL}/api/admin/credit-budget-policies/${editingPolicy.id}`, payload, { headers: getAuthHeaders() });
        messageApi.success(t('enterprise_credit_policy_updated'));
      } else {
        await axios.post(`${BACKEND_URL}/api/admin/credit-budget-policies`, payload, { headers: getAuthHeaders() });
        messageApi.success(t('enterprise_credit_policy_created'));
      }
      setPolicyModalOpen(false);
      await loadData();
    } catch {
      messageApi.error(t('enterprise_credit_policy_save_failed'));
    } finally {
      setSaving(false);
    }
  };

  const deletePolicy = async (id: string) => {
    try {
      await axios.delete(`${BACKEND_URL}/api/admin/credit-budget-policies/${id}`, { headers: getAuthHeaders() });
      messageApi.success(t('enterprise_credit_policy_deleted'));
      await loadData();
    } catch {
      messageApi.error(t('enterprise_credit_policy_delete_failed'));
    }
  };

  return (
    <Space orientation="vertical" size={16} style={{ width: '100%' }}>
      {contextHolder}

      <Card variant="borderless" style={{ borderRadius: 20 }} loading={loading}>
        <Paragraph type="secondary" style={{ marginBottom: 16 }}>
          {t('enterprise_credit_desc')}
        </Paragraph>
        <Space size={16} wrap>
          {(overview?.wallets || []).map((item) => (
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

      <Card variant="borderless" title={t('credit_wallets_title')} style={{ borderRadius: 20 }} loading={loading}>
        <Table
          rowKey="id"
          pagination={false}
          dataSource={wallets}
          columns={[
            { title: t('credit_credit_type_label'), dataIndex: 'creditType', key: 'creditType', render: (value: string) => <Tag color={value === 'platform' ? 'blue' : 'purple'}>{getCreditTypeLabel(value, t)}</Tag> },
            { title: t('credit_total_label'), dataIndex: 'totalCredits', key: 'totalCredits' },
            { title: t('credit_reserved_label'), dataIndex: 'reservedCredits', key: 'reservedCredits' },
            { title: t('credit_available_label'), dataIndex: 'availableCredits', key: 'availableCredits' },
          ]}
        />
      </Card>

      <Card
        variant="borderless"
        title={t('credit_grants_title')}
        extra={<Button type="primary" icon={<PlusOutlined />} onClick={openCreateGrant}>{t('credit_new_grant')}</Button>}
        style={{ borderRadius: 20 }}
        loading={loading}
      >
        <Table<CreditGrant>
          rowKey="id"
          pagination={false}
          dataSource={grants}
          columns={[
            { title: t('credit_credit_type_label'), dataIndex: 'creditType', key: 'creditType', render: (value: string) => getCreditTypeLabel(value, t) },
            { title: t('credit_source_type_label'), dataIndex: 'sourceType', key: 'sourceType', render: (value: string) => getGrantSourceTypeLabel(value, t) },
            { title: t('credit_granted_label'), dataIndex: 'grantedCredits', key: 'grantedCredits' },
            { title: t('credit_remaining_label'), dataIndex: 'remainingCredits', key: 'remainingCredits' },
            { title: t('credit_effective_at_label'), dataIndex: 'effectiveAt', key: 'effectiveAt', render: formatDisplayDateTime },
            { title: t('credit_expires_at_label'), dataIndex: 'expiresAt', key: 'expiresAt', render: formatDisplayDateTime },
          ]}
        />
      </Card>

      <Card variant="borderless" title={t('credit_ledger_title')} style={{ borderRadius: 20 }} loading={loading}>
        <Table<CreditLedgerEntry>
          rowKey="id"
          pagination={{ pageSize: 10 }}
          dataSource={ledger}
          scroll={{ x: 980 }}
          columns={[
            { title: t('credit_credit_type_label'), dataIndex: 'creditType', key: 'creditType', render: (value: string) => getCreditTypeLabel(value, t) },
            { title: t('credit_entry_type_label'), dataIndex: 'entryType', key: 'entryType', render: (value: string) => getEntryTypeLabel(value, t) },
            { title: t('credit_direction_label'), dataIndex: 'direction', key: 'direction', render: (value: string) => getDirectionLabel(value, t) },
            { title: t('credit_credits_label'), dataIndex: 'credits', key: 'credits' },
            { title: t('credit_balance_after_label'), dataIndex: 'balanceAfter', key: 'balanceAfter' },
            { title: t('credit_reserved_after_label'), dataIndex: 'reservedAfter', key: 'reservedAfter' },
            { title: t('credit_member_label'), dataIndex: 'memberUserId', key: 'memberUserId', render: (value?: string) => value ? (memberNameMap.get(value) || value) : t('common_empty_placeholder') },
            { title: t('credit_agent_label'), dataIndex: 'agentId', key: 'agentId', render: (value?: string) => value ? (agentNameMap.get(value) || value) : t('common_empty_placeholder') },
            { title: t('credit_reason_label'), dataIndex: 'reasonCode', key: 'reasonCode' },
            { title: t('credit_created_at_label'), dataIndex: 'createdAt', key: 'createdAt', render: formatDisplayDateTime },
          ]}
        />
      </Card>

      <Card
        variant="borderless"
        title={t('credit_price_books_title')}
        extra={<Button type="primary" icon={<PlusOutlined />} onClick={openCreatePriceBook}>{t('credit_new_enterprise_price_book')}</Button>}
        style={{ borderRadius: 20 }}
        loading={loading}
      >
        <Table
          rowKey="id"
          pagination={false}
          dataSource={enterprisePriceBooks}
          columns={[
            { title: t('credit_credit_type_label'), dataIndex: 'creditType', key: 'creditType', render: (value: string) => getCreditTypeLabel(value, t) },
            { title: t('credit_model_label'), dataIndex: 'modelId', key: 'modelId', render: (value: string) => modelNameMap.get(value) || value },
            { title: t('credit_funding_label'), dataIndex: 'fundingType', key: 'fundingType', render: (value: string) => getFundingTypeLabel(value, t) },
            { title: t('credit_currency_label'), dataIndex: 'currency', key: 'currency' },
            { title: t('credit_input_price_per_m_label'), dataIndex: 'costInputUsdPer1M', key: 'costInputUsdPer1M' },
            { title: t('credit_output_price_per_m_label'), dataIndex: 'costOutputUsdPer1M', key: 'costOutputUsdPer1M' },
            { title: t('credit_multiplier_label'), dataIndex: 'enterpriseMultiplier', key: 'enterpriseMultiplier' },
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

      <Card
        variant="borderless"
        title={t('credit_platform_override_title')}
        extra={<Button type="primary" icon={<PlusOutlined />} onClick={openCreatePlatformOverride}>{t('credit_new_platform_override')}</Button>}
        style={{ borderRadius: 20 }}
        loading={loading}
      >
        <Table
          rowKey="id"
          pagination={false}
          dataSource={platformOverridePriceBooks}
          columns={[
            { title: t('credit_credit_type_label'), dataIndex: 'creditType', key: 'creditType', render: (value: string) => getCreditTypeLabel(value, t) },
            { title: t('credit_model_label'), dataIndex: 'modelId', key: 'modelId', render: (value: string) => modelNameMap.get(value) || value },
            { title: t('credit_funding_label'), dataIndex: 'fundingType', key: 'fundingType', render: (value: string) => getFundingTypeLabel(value, t) },
            { title: t('credit_currency_label'), dataIndex: 'currency', key: 'currency' },
            { title: t('credit_input_price_per_m_label'), dataIndex: 'costInputUsdPer1M', key: 'costInputUsdPer1M' },
            { title: t('credit_output_price_per_m_label'), dataIndex: 'costOutputUsdPer1M', key: 'costOutputUsdPer1M' },
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

      <Card
        variant="borderless"
        title={t('credit_budget_policies_title')}
        extra={<Button type="primary" icon={<PlusOutlined />} onClick={openCreatePolicy}>{t('credit_new_policy')}</Button>}
        style={{ borderRadius: 20 }}
        loading={loading}
      >
        <Table
          rowKey="id"
          pagination={false}
          dataSource={policies}
          columns={[
            { title: t('credit_credit_type_label'), dataIndex: 'creditType', key: 'creditType', render: (value: string) => getCreditTypeLabel(value, t) },
            { title: t('credit_scope_label'), dataIndex: 'scopeType', key: 'scopeType', render: (value: string) => getScopeTypeLabel(value, t) },
            {
              title: t('credit_target_label'),
              dataIndex: 'scopeId',
              key: 'scopeId',
              render: (value: string, item: CreditBudgetPolicy) => {
                if (item.scopeType === 'agent') {
                  return agentNameMap.get(value) || value;
                }
                if (item.scopeType === 'member') {
                  return memberNameMap.get(value) || value;
                }
                return t('credit_current_enterprise');
              },
            },
            { title: t('credit_enabled_label'), dataIndex: 'enabled', key: 'enabled', render: (value: boolean) => value ? t('credit_yes') : t('credit_no') },
            { title: t('credit_daily_credits_label'), dataIndex: 'dailyCreditLimit', key: 'dailyCreditLimit' },
            { title: t('credit_monthly_credits_label'), dataIndex: 'monthlyCreditLimit', key: 'monthlyCreditLimit' },
            { title: t('credit_daily_tokens_label'), dataIndex: 'dailyTokenLimit', key: 'dailyTokenLimit' },
            { title: t('credit_monthly_tokens_label'), dataIndex: 'monthlyTokenLimit', key: 'monthlyTokenLimit' },
            { title: t('credit_daily_usd_label'), dataIndex: 'dailyUsdLimit', key: 'dailyUsdLimit' },
            { title: t('credit_monthly_usd_label'), dataIndex: 'monthlyUsdLimit', key: 'monthlyUsdLimit' },
            { title: t('credit_hard_limit_label'), dataIndex: 'hardLimit', key: 'hardLimit', render: (value: boolean) => value ? t('credit_yes') : t('credit_no') },
            {
              title: t('credit_actions_label'),
              key: 'actions',
              render: (_: unknown, item: CreditBudgetPolicy) => (
                <Space>
                  <Button onClick={() => openEditPolicy(item)}>{t('credit_edit_action')}</Button>
                  <Button danger onClick={() => deletePolicy(item.id)}>{t('credit_delete_action')}</Button>
                </Space>
              ),
            },
          ]}
        />
      </Card>

      <Modal
        title={t('credit_create_grant')}
        open={grantModalOpen}
        onOk={saveGrant}
        onCancel={() => setGrantModalOpen(false)}
        confirmLoading={saving}
        destroyOnHidden
      >
        <Form form={grantForm} layout="vertical">
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
        title={editingPriceBook ? t('credit_edit_price_book') : t('credit_create_price_book')}
        open={priceBookModalOpen}
        onOk={savePriceBook}
        onCancel={() => setPriceBookModalOpen(false)}
        confirmLoading={saving}
        destroyOnHidden
      >
        <Form form={priceBookForm} layout="vertical">
          <Form.Item label={t('credit_credit_type_label')} name="creditType" rules={[{ required: true, message: requiredMessage(t('credit_credit_type_label')) }]}>
            <Select
              onChange={handlePriceBookCreditTypeChange}
              options={[
                ...creditTypeOptions,
              ]}
            />
          </Form.Item>
          <Form.Item noStyle shouldUpdate={(prev, next) => prev.creditType !== next.creditType}>
            {({ getFieldValue }) => {
              const creditType = getFieldValue('creditType');
              const options = creditType === 'platform'
                ? platformFundedModels.map((item) => ({ label: `${item.displayName} (${item.model})`, value: item.id }))
                : enterpriseFundedModels.map((item) => ({ label: `${item.displayName} (${item.model})`, value: item.id }));
              return (
                <Form.Item label={t('credit_model_label')} name="modelId" rules={[{ required: true, message: requiredMessage(t('credit_model_label')) }]}>
                  <Select options={options} />
                </Form.Item>
              );
            }}
          </Form.Item>
          <Form.Item name="fundingType" hidden>
            <Input />
          </Form.Item>
          <Form.Item name="modelSourceType" hidden>
            <Input />
          </Form.Item>
          <Form.Item name="modelScope" hidden>
            <Input />
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

      <Modal
        title={editingPolicy ? t('credit_edit_budget_policy') : t('credit_create_budget_policy')}
        open={policyModalOpen}
        onOk={savePolicy}
        onCancel={() => setPolicyModalOpen(false)}
        confirmLoading={saving}
        destroyOnHidden
      >
        <Form form={policyForm} layout="vertical">
          <Form.Item label={t('credit_credit_type_label')} name="creditType" rules={[{ required: true, message: requiredMessage(t('credit_credit_type_label')) }]}>
            <Select options={creditTypeOptions} />
          </Form.Item>
          <Form.Item label={t('credit_scope_label')} name="scopeType" rules={[{ required: true, message: requiredMessage(t('credit_scope_label')) }]}>
            <Select options={scopeTypeOptions} />
          </Form.Item>
          <Form.Item noStyle shouldUpdate={(prev, next) => prev.scopeType !== next.scopeType}>
            {({ getFieldValue }) => {
              const scopeType = getFieldValue('scopeType');
              if (scopeType === 'member') {
                return (
                  <Form.Item label={t('credit_member_label')} name="scopeId" rules={[{ required: true, message: requiredMessage(t('credit_member_label')) }]}>
                    <Select options={members.map((item) => ({ label: item.displayName || item.email, value: item.userId }))} />
                  </Form.Item>
                );
              }
              if (scopeType === 'agent') {
                return (
                  <Form.Item label={t('credit_agent_label')} name="scopeId" rules={[{ required: true, message: requiredMessage(t('credit_agent_label')) }]}>
                    <Select options={agents.map((item) => ({ label: item.agentName, value: item.id }))} />
                  </Form.Item>
                );
              }
              return null;
            }}
          </Form.Item>
          <Form.Item label={t('credit_enabled_label')} name="enabled" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item label={t('credit_daily_credit_limit_label')} name="dailyCreditLimit">
            <InputNumber controls={false} style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label={t('credit_monthly_credit_limit_label')} name="monthlyCreditLimit">
            <InputNumber controls={false} style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label={t('credit_daily_token_limit_label')} name="dailyTokenLimit">
            <InputNumber controls={false} style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label={t('credit_monthly_token_limit_label')} name="monthlyTokenLimit">
            <InputNumber controls={false} style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label={t('credit_daily_usd_limit_label')} name="dailyUsdLimit">
            <InputNumber controls={false} style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label={t('credit_monthly_usd_limit_label')} name="monthlyUsdLimit">
            <InputNumber controls={false} style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label={t('credit_hard_limit_label')} name="hardLimit" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
};

export default EnterpriseCreditSettingsTab;
