import React, { useEffect, useMemo, useState } from 'react';
import {
  Button,
  Card,
  Drawer,
  Empty,
  Form,
  Input,
  Modal,
  Select,
  Space,
  Table,
  Tag,
  Typography,
  message,
} from 'antd';
import { CloudDownloadOutlined, PlusOutlined, SafetyOutlined } from '@ant-design/icons';
import axios from 'axios';
import { useTranslation } from 'react-i18next';
import { BACKEND_URL } from '../../config';
import { casdoorService } from '../identity/CasdoorService';

const { Title, Paragraph, Text } = Typography;
const { TextArea } = Input;
const ADMIN_SKILLS_API_BASE = `${BACKEND_URL}/api/admin/skills`;
const scrollableModalStyles = {
  body: {
    maxHeight: '70vh',
    overflowY: 'auto' as const,
    paddingRight: 8,
  },
};

type PlatformTabKey = 'skills' | 'hubs' | 'imports';

interface SkillItem {
  id: string;
  code: string;
  name: string;
  description: string;
  sourceType: string;
  providerType: string;
  trustLevel: string;
  status: string;
  latestVersionId?: string;
  latestPublishedVersionId?: string;
  latestStableVersionId?: string;
  createdAt: string;
  updatedAt: string;
}

interface SkillVersionItem {
  id: string;
  version: string;
  releaseChannel: string;
  releaseStatus: string;
  changeLog: string;
  createdAt: string;
  publishedAt?: string;
}

interface SkillReferenceItem {
  id: string;
  fromSkillVersionId: string;
  toSkillVersionId: string;
  invokeMode: string;
  conditionExpr: string;
  contextPassthrough: boolean;
  resultPassthrough: boolean;
  sortOrder: number;
}

interface SkillDetail {
  skill: SkillItem;
  versions: SkillVersionItem[];
  references: SkillReferenceItem[];
}

interface SkillHubItem {
  id: string;
  hubCode: string;
  name: string;
  hubType: string;
  baseUrl: string;
  status: string;
  trustLevel: string;
  syncMode: string;
  authScheme: string;
  updatedAt: string;
}

interface SkillImportJobItem {
  id: string;
  hubId: string;
  requestedBy: string;
  sourceLocator: string;
  sourceNamespace: string;
  sourceVersion: string;
  jobStatus: string;
  targetSkillId?: string;
  targetSkillVersionId?: string;
  errorMessage?: string;
  createdAt: string;
  startedAt?: string;
  finishedAt?: string;
}

interface SkillFormValues {
  code: string;
  name: string;
  description: string;
  sourceType: string;
  providerType: string;
}

interface SkillVersionFormValues {
  version: string;
  manifest: string;
  inputSchema: string;
  outputSchema: string;
  defaultPolicy: string;
  runtimeContract: string;
  references: string;
  changeLog: string;
}

interface ReferenceEditorFormValues {
  references: string;
}

interface SkillHubFormValues {
  hubCode: string;
  name: string;
  hubType: string;
  baseUrl: string;
  status: string;
  trustLevel: string;
  syncMode: string;
  authScheme: string;
  config: string;
  secret: string;
  importPolicy: string;
  allowedNamespaces: string;
  networkPolicy: string;
  signaturePolicy: string;
}

interface ImportJobFormValues {
  hubId: string;
  sourceLocator: string;
  sourceNamespace: string;
  sourceVersion: string;
}

function getAuthHeaders() {
  const token = casdoorService.getToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

function formatDateTime(value?: string, locale?: string) {
  if (!value) {
    return '-';
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return '-';
  }
  return new Intl.DateTimeFormat(locale || 'en', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date);
}

function renderStatusTag(status: string, label: string) {
  const normalized = (status || '').toLowerCase();
  const color = normalized === 'published' || normalized === 'completed' || normalized === 'enabled'
    ? 'success'
    : normalized === 'reviewing' || normalized === 'normalizing'
      ? 'processing'
      : normalized === 'disabled'
        ? 'default'
        : normalized === 'deprecated'
          ? 'warning'
          : normalized === 'failed' || normalized === 'blocked'
            ? 'error'
            : 'blue';
  return <Tag color={color}>{label || status || '-'}</Tag>;
}

function renderTrustTag(trustLevel: string, label: string) {
  const normalized = (trustLevel || '').toLowerCase();
  const color = normalized.includes('trusted')
    ? 'success'
    : normalized.includes('verified')
      ? 'processing'
      : normalized.includes('blocked')
        ? 'error'
        : 'default';
  return <Tag color={color}>{label || trustLevel || '-'}</Tag>;
}

function translatePlatformSkillError(
  errorText: unknown,
  fallbackMessage: string,
  t: any,
) {
  if (typeof errorText !== 'string' || !errorText.trim()) {
    return fallbackMessage;
  }

  const normalized = errorText.trim().toLowerCase();
  const errorKeyMap: Record<string, string> = {
    'skill not found': 'platform_skill_error_not_found',
    'skill version not found': 'platform_skill_error_version_not_found',
    'skill code already exists': 'platform_skill_error_code_exists',
    'skill is not published': 'platform_skill_error_not_published',
    'skill version is not ready for publish': 'platform_skill_error_version_not_ready',
    'skill cannot be enabled': 'platform_skill_error_enablement_denied',
    'skill cannot be installed': 'platform_skill_error_install_denied',
    'skill trust level does not allow this operation': 'platform_skill_error_trust_denied',
    'skill reference cycle detected': 'platform_skill_error_cycle_detected',
    'skill hub not found': 'platform_skill_error_hub_not_found',
  };

  const errorKey = errorKeyMap[normalized];
  return errorKey ? t(errorKey, { defaultValue: errorText }) : errorText;
}

function renderSummaryCards(items: Array<{ label: string; value: number }>) {
  return (
    <div
      style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
        gap: 16,
        marginBottom: 16,
      }}
    >
      {items.map((item) => (
        <Card key={item.label} size="small">
          <Text type="secondary">{item.label}</Text>
          <Title level={4} style={{ margin: '8px 0 0' }}>{item.value}</Title>
        </Card>
      ))}
    </div>
  );
}

const PlatformSkillsPage: React.FC = () => {
  const { t, i18n } = useTranslation();
  const [messageApi, contextHolder] = message.useMessage();
  const [activeTab, setActiveTab] = useState<PlatformTabKey>('skills');
  const [skills, setSkills] = useState<SkillItem[]>([]);
  const [hubs, setHubs] = useState<SkillHubItem[]>([]);
  const [importJobs, setImportJobs] = useState<SkillImportJobItem[]>([]);
  const [skillsLoading, setSkillsLoading] = useState(true);
  const [hubsLoading, setHubsLoading] = useState(true);
  const [importsLoading, setImportsLoading] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);
  const [versionOpen, setVersionOpen] = useState(false);
  const [referenceEditorOpen, setReferenceEditorOpen] = useState(false);
  const [detailOpen, setDetailOpen] = useState(false);
  const [hubOpen, setHubOpen] = useState(false);
  const [importOpen, setImportOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const [referenceLoading, setReferenceLoading] = useState(false);
  const [publishingId, setPublishingId] = useState<string>('');
  const [selectedSkill, setSelectedSkill] = useState<SkillDetail | null>(null);
  const [editingHub, setEditingHub] = useState<SkillHubItem | null>(null);
  const [createForm] = Form.useForm<SkillFormValues>();
  const [versionForm] = Form.useForm<SkillVersionFormValues>();
  const [referenceEditorForm] = Form.useForm<ReferenceEditorFormValues>();
  const [hubForm] = Form.useForm<SkillHubFormValues>();
  const [importForm] = Form.useForm<ImportJobFormValues>();
  const [editingVersion, setEditingVersion] = useState<SkillVersionItem | null>(null);

  const translateStatus = (status?: string) => {
    if (!status) {
      return '-';
    }
    return t(`platform_skills_status_${status}`, status);
  };

  const translateTrustLevel = (trustLevel?: string) => {
    if (!trustLevel) {
      return '-';
    }
    return t(`platform_skills_trust_${trustLevel}`, trustLevel);
  };

  const translateSourceType = (sourceType?: string) => {
    if (!sourceType) {
      return '-';
    }
    return t(`platform_skills_source_${sourceType}`, sourceType);
  };

  const translateProviderType = (providerType?: string) => {
    if (!providerType) {
      return '-';
    }
    return t(`platform_skills_provider_${providerType}`, providerType);
  };

  const translateHubType = (hubType?: string) => {
    if (!hubType) {
      return '-';
    }
    return t(`platform_skills_hub_type_${hubType}`, hubType);
  };

  const translateSyncMode = (syncMode?: string) => {
    if (!syncMode) {
      return '-';
    }
    return t(`platform_skills_sync_mode_${syncMode}`, syncMode);
  };

  const translateAuthScheme = (authScheme?: string) => {
    if (!authScheme) {
      return '-';
    }
    return t(`platform_skills_auth_scheme_${authScheme}`, authScheme);
  };

  const translateInvokeMode = (invokeMode?: string) => {
    if (!invokeMode) {
      return t('platform_skills_invoke_mode_sync');
    }
    return t(`platform_skills_invoke_mode_${invokeMode}`, invokeMode);
  };

  const fetchSkills = async () => {
    setSkillsLoading(true);
    try {
      const res = await axios.get(`${ADMIN_SKILLS_API_BASE}?view=governance`, {
        headers: getAuthHeaders(),
      });
      setSkills(Array.isArray(res.data) ? res.data : []);
    } catch {
      messageApi.error(t('platform_skills_load_failed'));
      setSkills([]);
    } finally {
      setSkillsLoading(false);
    }
  };

  const fetchHubs = async () => {
    setHubsLoading(true);
    try {
      const res = await axios.get(`${BACKEND_URL}/api/admin/platform/skill-hubs`, {
        headers: getAuthHeaders(),
      });
      setHubs(Array.isArray(res.data) ? res.data : []);
    } catch {
      messageApi.error(t('platform_skill_hubs_load_failed'));
      setHubs([]);
    } finally {
      setHubsLoading(false);
    }
  };

  const fetchImportJobs = async () => {
    setImportsLoading(true);
    try {
      const res = await axios.get(`${BACKEND_URL}/api/admin/platform/skill-import-jobs`, {
        headers: getAuthHeaders(),
      });
      setImportJobs(Array.isArray(res.data) ? res.data : []);
    } catch {
      messageApi.error(t('platform_skill_import_jobs_load_failed'));
      setImportJobs([]);
    } finally {
      setImportsLoading(false);
    }
  };

  const fetchDetail = async (skillId: string) => {
    try {
      const res = await axios.get(`${ADMIN_SKILLS_API_BASE}/${skillId}`, {
        headers: getAuthHeaders(),
      });
      setSelectedSkill(res.data || null);
      setDetailOpen(true);
    } catch {
      messageApi.error(t('platform_skill_detail_load_failed'));
    }
  };

  useEffect(() => {
    void Promise.all([fetchSkills(), fetchHubs(), fetchImportJobs()]);
  }, [t]);

  const skillSummary = useMemo(() => ({
    total: skills.length,
    published: skills.filter((item) => item.status === 'published').length,
    draft: skills.filter((item) => item.status === 'draft').length,
    disabled: skills.filter((item) => item.status === 'disabled').length,
  }), [skills]);

  const hubSummary = useMemo(() => ({
    total: hubs.length,
    enabled: hubs.filter((item) => item.status === 'enabled').length,
    openapi: hubs.filter((item) => item.hubType === 'openapi_hub').length,
    mcp: hubs.filter((item) => item.hubType === 'mcp_hub').length,
  }), [hubs]);

  const importSummary = useMemo(() => ({
    total: importJobs.length,
    completed: importJobs.filter((item) => item.jobStatus === 'completed').length,
    running: importJobs.filter((item) => item.jobStatus === 'normalizing').length,
    failed: importJobs.filter((item) => item.jobStatus === 'failed').length,
  }), [importJobs]);

  const hubNameMap = useMemo(() => {
    return hubs.reduce<Record<string, string>>((acc, hub) => {
      acc[hub.id] = hub.name;
      return acc;
    }, {});
  }, [hubs]);

  const openCreateSkill = () => {
    createForm.resetFields();
    createForm.setFieldsValue({ sourceType: 'builtin', providerType: 'native' });
    setCreateOpen(true);
  };

  const openCreateHub = (hub?: SkillHubItem) => {
    setEditingHub(hub || null);
    hubForm.resetFields();
    hubForm.setFieldsValue({
      hubCode: hub?.hubCode || '',
      name: hub?.name || '',
      hubType: hub?.hubType || 'openapi_hub',
      baseUrl: hub?.baseUrl || '',
      status: hub?.status || 'enabled',
      trustLevel: hub?.trustLevel || 'partner_verified',
      syncMode: hub?.syncMode || 'manual',
      authScheme: hub?.authScheme || 'none',
      config: '{}',
      secret: '{}',
      importPolicy: '{}',
      allowedNamespaces: '[]',
      networkPolicy: '{}',
      signaturePolicy: '{}',
    });
    setHubOpen(true);
  };

  const openImportJob = () => {
    importForm.resetFields();
    setImportOpen(true);
  };

  const handleCreateSkill = async () => {
    const values = await createForm.validateFields();
    setSaving(true);
    try {
      await axios.post(ADMIN_SKILLS_API_BASE, values, {
        headers: getAuthHeaders(),
      });
      messageApi.success(t('platform_skill_create_success'));
      setCreateOpen(false);
      createForm.resetFields();
      await fetchSkills();
    } catch (error: any) {
      const errorText = error?.response?.data;
      messageApi.error(
        translatePlatformSkillError(errorText, t('platform_skill_create_failed'), t),
      );
    } finally {
      setSaving(false);
    }
  };

  const handleCreateVersion = async () => {
    if (!selectedSkill?.skill?.id) {
      return;
    }
    const values = await versionForm.validateFields();
    setSaving(true);
    try {
      const references = parseJSONArrayOrThrow(values.references);
      await axios.post(`${ADMIN_SKILLS_API_BASE}/${selectedSkill.skill.id}/versions`, {
        version: values.version,
        manifest: safeParseJSON(values.manifest),
        inputSchema: safeParseJSON(values.inputSchema),
        outputSchema: safeParseJSON(values.outputSchema),
        defaultPolicy: safeParseJSON(values.defaultPolicy),
        runtimeContract: safeParseJSON(values.runtimeContract),
        references,
        changeLog: values.changeLog,
      }, {
        headers: getAuthHeaders(),
      });
      messageApi.success(t('platform_skill_version_create_success'));
      setVersionOpen(false);
      versionForm.resetFields();
      await fetchDetail(selectedSkill.skill.id);
      await fetchSkills();
    } catch (error: any) {
      if (error instanceof SyntaxError || error?.message === 'array') {
        messageApi.error(t('platform_skill_references_invalid_json'));
        return;
      }
      const errorText = error?.response?.data;
      messageApi.error(
        translatePlatformSkillError(errorText, t('platform_skill_version_create_failed'), t),
      );
    } finally {
      setSaving(false);
    }
  };

  const handleSubmitReview = async (versionId: string) => {
    if (!selectedSkill?.skill?.id) {
      return;
    }
    try {
      await axios.post(`${ADMIN_SKILLS_API_BASE}/${selectedSkill.skill.id}/submit-review`, {
        skillVersionId: versionId,
      }, {
        headers: getAuthHeaders(),
      });
      messageApi.success(t('platform_skill_submit_review_success'));
      await fetchDetail(selectedSkill.skill.id);
      await fetchSkills();
    } catch (error: any) {
      const errorText = error?.response?.data;
      messageApi.error(
        translatePlatformSkillError(errorText, t('platform_skill_submit_review_failed'), t),
      );
    }
  };

  const handlePublish = async (versionId: string) => {
    if (!selectedSkill?.skill?.id) {
      return;
    }
    setPublishingId(versionId);
    try {
      await axios.post(`${ADMIN_SKILLS_API_BASE}/${selectedSkill.skill.id}/publish`, {
        skillVersionId: versionId,
        releaseChannel: 'stable',
      }, {
        headers: getAuthHeaders(),
      });
      messageApi.success(t('platform_skill_publish_success'));
      await fetchDetail(selectedSkill.skill.id);
      await fetchSkills();
    } catch (error: any) {
      const errorText = error?.response?.data;
      messageApi.error(
        translatePlatformSkillError(errorText, t('platform_skill_publish_failed'), t),
      );
    } finally {
      setPublishingId('');
    }
  };

  const openReferenceEditor = async (version: SkillVersionItem) => {
    if (!selectedSkill?.skill?.id) {
      return;
    }
    setEditingVersion(version);
    setReferenceEditorOpen(true);
    setReferenceLoading(true);
    referenceEditorForm.setFieldsValue({ references: '[]' });
    try {
      const res = await axios.get(`${ADMIN_SKILLS_API_BASE}/${selectedSkill.skill.id}/versions/${version.id}/references`, {
        headers: getAuthHeaders(),
      });
      const currentReferences = Array.isArray(res.data) ? res.data : [];
      referenceEditorForm.setFieldsValue({
        references: JSON.stringify(currentReferences, null, 2),
      });
    } catch {
      messageApi.error(t('platform_skill_references_load_failed'));
      setReferenceEditorOpen(false);
      setEditingVersion(null);
      referenceEditorForm.resetFields();
    } finally {
      setReferenceLoading(false);
    }
  };

  const handleUpdateReferences = async () => {
    if (!selectedSkill?.skill?.id || !editingVersion?.id) {
      return;
    }
    const values = await referenceEditorForm.validateFields();
    setSaving(true);
    try {
      const references = parseJSONArrayOrThrow(values.references);
      await axios.post(`${ADMIN_SKILLS_API_BASE}/${selectedSkill.skill.id}/references`, {
        skillVersionId: editingVersion.id,
        references,
      }, {
        headers: getAuthHeaders(),
      });
      messageApi.success(t('platform_skill_references_update_success'));
      setReferenceEditorOpen(false);
      setEditingVersion(null);
      referenceEditorForm.resetFields();
      await fetchDetail(selectedSkill.skill.id);
    } catch (error: any) {
      if (error instanceof SyntaxError || error?.message === 'array') {
        messageApi.error(t('platform_skill_references_invalid_json'));
        return;
      }
      const errorText = error?.response?.data;
      messageApi.error(
        translatePlatformSkillError(errorText, t('platform_skill_references_update_failed'), t),
      );
    } finally {
      setSaving(false);
    }
  };

  const handleSaveHub = async () => {
    const values = await hubForm.validateFields();
    setSaving(true);
    try {
      const payload = {
        hubCode: values.hubCode,
        name: values.name,
        hubType: values.hubType,
        baseUrl: values.baseUrl,
        status: values.status,
        trustLevel: values.trustLevel,
        syncMode: values.syncMode,
        authScheme: values.authScheme,
        config: safeParseJSON(values.config),
        secret: safeParseJSON(values.secret),
        importPolicy: safeParseJSON(values.importPolicy),
        allowedNamespaces: safeParseJSONArray(values.allowedNamespaces),
        networkPolicy: safeParseJSON(values.networkPolicy),
        signaturePolicy: safeParseJSON(values.signaturePolicy),
      };
      if (editingHub?.id) {
        await axios.put(`${BACKEND_URL}/api/admin/platform/skill-hubs/${editingHub.id}`, payload, {
          headers: getAuthHeaders(),
        });
      } else {
        await axios.post(`${BACKEND_URL}/api/admin/platform/skill-hubs`, payload, {
          headers: getAuthHeaders(),
        });
      }
      messageApi.success(editingHub?.id ? t('platform_skill_hub_update_success') : t('platform_skill_hub_create_success'));
      setHubOpen(false);
      setEditingHub(null);
      await fetchHubs();
    } catch (error: any) {
      const errorText = error?.response?.data;
      messageApi.error(typeof errorText === 'string' ? errorText : t('platform_skill_hub_save_failed'));
    } finally {
      setSaving(false);
    }
  };

  const handleImportSkill = async () => {
    const values = await importForm.validateFields();
    setSaving(true);
    try {
      await axios.post(`${BACKEND_URL}/api/admin/platform/skill-import-jobs`, values, {
        headers: getAuthHeaders(),
      });
      messageApi.success(t('platform_skill_import_create_success'));
      setImportOpen(false);
      importForm.resetFields();
      await Promise.all([fetchImportJobs(), fetchSkills()]);
    } catch (error: any) {
      const errorText = error?.response?.data;
      messageApi.error(typeof errorText === 'string' ? errorText : t('platform_skill_import_create_failed'));
    } finally {
      setSaving(false);
    }
  };

  const actionButton = activeTab === 'skills'
    ? (
      <Button type="primary" icon={<PlusOutlined />} onClick={openCreateSkill}>
        {t('platform_skills_new_skill')}
      </Button>
    )
    : activeTab === 'hubs'
      ? (
        <Button type="primary" icon={<PlusOutlined />} onClick={() => openCreateHub()}>
          {t('platform_skills_new_hub')}
        </Button>
      )
      : (
        <Button type="primary" icon={<CloudDownloadOutlined />} onClick={openImportJob}>
          {t('platform_skills_import_start')}
        </Button>
      );

  return (
    <div style={{ animation: 'fadeIn 0.5s ease-out' }}>
      {contextHolder}
      <div style={{ marginBottom: 24, display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 16, flexWrap: 'wrap' }}>
        <div>
          <Title level={3} style={{ marginBottom: 8 }}>
            <SafetyOutlined style={{ marginRight: 8 }} />
            {t('platform_skills_title')}
          </Title>
          <Paragraph type="secondary" style={{ marginBottom: 0 }}>
            {t('platform_skills_desc')}
          </Paragraph>
        </div>
        {actionButton}
      </div>

      <Card size="small" style={{ marginBottom: 16, borderRadius: 16 }}>
        <Space wrap size={12}>
          {[
            { key: 'skills', label: `${t('platform_skills_tab_skills')} (${skillSummary.total})` },
            { key: 'hubs', label: `${t('platform_skills_tab_hubs')} (${hubSummary.total})` },
            { key: 'imports', label: `${t('platform_skills_tab_imports')} (${importSummary.total})` },
          ].map((item) => (
            <Button
              key={item.key}
              type={activeTab === item.key ? 'primary' : 'default'}
              onClick={() => setActiveTab(item.key as PlatformTabKey)}
            >
              {item.label}
            </Button>
          ))}
        </Space>
      </Card>

      {activeTab === 'skills' ? (
        <>
          {renderSummaryCards([
            { label: t('platform_skills_summary_total'), value: skillSummary.total },
            { label: t('platform_skills_summary_published'), value: skillSummary.published },
            { label: t('platform_skills_summary_draft'), value: skillSummary.draft },
            { label: t('platform_skills_summary_disabled'), value: skillSummary.disabled },
          ])}

          <Card variant="borderless" style={{ borderRadius: 20 }}>
            <Table
              rowKey="id"
              loading={skillsLoading}
              dataSource={skills}
              onRow={(record) => ({
                onClick: () => fetchDetail(record.id),
                style: { cursor: 'pointer' },
              })}
              pagination={{ pageSize: 10 }}
              scroll={{ x: 1480 }}
              locale={{
                emptyText: <Empty description={t('platform_skills_empty')} image={Empty.PRESENTED_IMAGE_SIMPLE} />,
              }}
              columns={[
                { title: t('platform_skills_column_code'), dataIndex: 'code', key: 'code', width: 220, ellipsis: true },
                { title: t('platform_skills_column_name'), dataIndex: 'name', key: 'name', width: 220 },
                {
                  title: t('platform_skills_column_source'),
                  dataIndex: 'sourceType',
                  key: 'sourceType',
                  width: 150,
                  render: (value: string) => translateSourceType(value),
                },
                {
                  title: t('platform_skills_column_provider'),
                  dataIndex: 'providerType',
                  key: 'providerType',
                  width: 150,
                  render: (value: string) => translateProviderType(value),
                },
                {
                  title: t('platform_skills_column_trust_level'),
                  dataIndex: 'trustLevel',
                  key: 'trustLevel',
                  width: 150,
                  render: (value: string) => renderTrustTag(value, translateTrustLevel(value)),
                },
                {
                  title: t('platform_skills_column_status'),
                  dataIndex: 'status',
                  key: 'status',
                  width: 120,
                  render: (value: string) => renderStatusTag(value, translateStatus(value)),
                },
                {
                  title: t('platform_skills_column_latest_stable'),
                  dataIndex: 'latestStableVersionId',
                  key: 'latestStableVersionId',
                  width: 220,
                  ellipsis: true,
                  render: (value?: string) => value || '-',
                },
                {
                  title: t('platform_skills_column_updated_at'),
                  dataIndex: 'updatedAt',
                  key: 'updatedAt',
                  width: 200,
                  render: (value: string) => formatDateTime(value, i18n.resolvedLanguage || i18n.language),
                },
                {
                  title: t('platform_skills_column_actions'),
                  key: 'actions',
                  width: 140,
                  render: (_: unknown, record: SkillItem) => (
                    <Button
                      size="small"
                      onClick={(event) => {
                        event.stopPropagation();
                        void fetchDetail(record.id);
                      }}
                    >
                      {t('platform_skills_view_detail')}
                    </Button>
                  ),
                },
              ]}
            />
          </Card>
        </>
      ) : null}

      {activeTab === 'hubs' ? (
        <>
          {renderSummaryCards([
            { label: t('platform_skill_hubs_summary_total'), value: hubSummary.total },
            { label: t('platform_skill_hubs_summary_enabled'), value: hubSummary.enabled },
            { label: t('platform_skill_hubs_summary_openapi'), value: hubSummary.openapi },
            { label: t('platform_skill_hubs_summary_mcp'), value: hubSummary.mcp },
          ])}

          <Card variant="borderless" style={{ borderRadius: 20 }}>
            <Table
              rowKey="id"
              loading={hubsLoading}
              dataSource={hubs}
              pagination={{ pageSize: 10 }}
              scroll={{ x: 1500 }}
              locale={{
                emptyText: <Empty description={t('platform_skill_hubs_empty')} image={Empty.PRESENTED_IMAGE_SIMPLE} />,
              }}
              columns={[
                { title: t('platform_skills_column_code'), dataIndex: 'hubCode', key: 'hubCode', width: 200, ellipsis: true },
                { title: t('platform_skills_column_name'), dataIndex: 'name', key: 'name', width: 220 },
                {
                  title: t('platform_skill_hubs_column_type'),
                  dataIndex: 'hubType',
                  key: 'hubType',
                  width: 180,
                  render: (value: string) => translateHubType(value),
                },
                {
                  title: t('platform_skill_hubs_column_base_url'),
                  dataIndex: 'baseUrl',
                  key: 'baseUrl',
                  width: 320,
                  ellipsis: true,
                  render: (value: string) => value || '-',
                },
                {
                  title: t('platform_skills_column_trust_level'),
                  dataIndex: 'trustLevel',
                  key: 'trustLevel',
                  width: 150,
                  render: (value: string) => renderTrustTag(value, translateTrustLevel(value)),
                },
                {
                  title: t('platform_skills_column_status'),
                  dataIndex: 'status',
                  key: 'status',
                  width: 120,
                  render: (value: string) => renderStatusTag(value, translateStatus(value)),
                },
                {
                  title: t('platform_skills_column_updated_at'),
                  dataIndex: 'updatedAt',
                  key: 'updatedAt',
                  width: 180,
                  render: (value: string) => formatDateTime(value, i18n.resolvedLanguage || i18n.language),
                },
                {
                  title: t('platform_skills_column_actions'),
                  key: 'actions',
                  width: 120,
                  render: (_: unknown, record: SkillHubItem) => (
                    <Button size="small" onClick={() => openCreateHub(record)}>
                      {t('platform_skills_edit')}
                    </Button>
                  ),
                },
              ]}
            />
          </Card>
        </>
      ) : null}

      {activeTab === 'imports' ? (
        <>
          {renderSummaryCards([
            { label: t('platform_skill_import_jobs_summary_total'), value: importSummary.total },
            { label: t('platform_skill_import_jobs_summary_completed'), value: importSummary.completed },
            { label: t('platform_skill_import_jobs_summary_running'), value: importSummary.running },
            { label: t('platform_skill_import_jobs_summary_failed'), value: importSummary.failed },
          ])}

          <Card variant="borderless" style={{ borderRadius: 20 }}>
            <Table
              rowKey="id"
              loading={importsLoading}
              dataSource={importJobs}
              pagination={{ pageSize: 10 }}
              scroll={{ x: 1800 }}
              locale={{
                emptyText: <Empty description={t('platform_skill_import_jobs_empty')} image={Empty.PRESENTED_IMAGE_SIMPLE} />,
              }}
              columns={[
                {
                  title: t('platform_skill_import_jobs_column_hub'),
                  dataIndex: 'hubId',
                  key: 'hubId',
                  width: 180,
                  render: (value: string) => hubNameMap[value] || value || '-',
                },
                {
                  title: t('platform_skill_import_jobs_column_source_locator'),
                  dataIndex: 'sourceLocator',
                  key: 'sourceLocator',
                  width: 280,
                  ellipsis: true,
                  render: (value: string) => value || '-',
                },
                {
                  title: t('platform_skill_import_jobs_column_source_namespace'),
                  dataIndex: 'sourceNamespace',
                  key: 'sourceNamespace',
                  width: 180,
                  render: (value: string) => value || '-',
                },
                { title: t('platform_skill_import_jobs_column_source_version'), dataIndex: 'sourceVersion', key: 'sourceVersion', width: 120, render: (value: string) => value || '-' },
                {
                  title: t('platform_skills_column_status'),
                  dataIndex: 'jobStatus',
                  key: 'jobStatus',
                  width: 120,
                  render: (value: string) => renderStatusTag(value, translateStatus(value)),
                },
                {
                  title: t('platform_skill_import_jobs_column_target_skill'),
                  dataIndex: 'targetSkillId',
                  key: 'targetSkillId',
                  width: 160,
                  render: (value?: string) => value || '-',
                },
                {
                  title: t('platform_skill_import_jobs_column_target_version'),
                  dataIndex: 'targetSkillVersionId',
                  key: 'targetSkillVersionId',
                  width: 160,
                  render: (value?: string) => value || '-',
                },
                {
                  title: t('platform_skill_import_jobs_column_finished_at'),
                  dataIndex: 'finishedAt',
                  key: 'finishedAt',
                  width: 180,
                  render: (value?: string) => formatDateTime(value, i18n.resolvedLanguage || i18n.language),
                },
                {
                  title: t('platform_skill_import_jobs_column_error_message'),
                  dataIndex: 'errorMessage',
                  key: 'errorMessage',
                  width: 280,
                  ellipsis: true,
                  render: (value?: string) => value || '-',
                },
              ]}
            />
          </Card>
        </>
      ) : null}

      <Modal
        title={t('platform_skills_new_skill')}
        open={createOpen}
        onCancel={() => setCreateOpen(false)}
        onOk={handleCreateSkill}
        confirmLoading={saving}
        okText={t('agent_save')}
        cancelText={t('agent_cancel')}
        styles={scrollableModalStyles}
        destroyOnHidden
      >
        <Form
          layout="vertical"
          form={createForm}
          initialValues={{ sourceType: 'builtin', providerType: 'native' }}
        >
          <Form.Item label={t('platform_skills_form_code')} name="code" rules={[{ required: true, message: t('platform_skills_form_code_required') }]}>
            <Input placeholder="knowledge.search" />
          </Form.Item>
          <Form.Item label={t('platform_skills_form_name')} name="name" rules={[{ required: true, message: t('platform_skills_form_name_required') }]}>
            <Input placeholder={t('platform_skills_form_name_placeholder')} />
          </Form.Item>
          <Form.Item label={t('platform_skills_form_description')} name="description">
            <TextArea rows={3} placeholder={t('platform_skills_form_description_placeholder')} />
          </Form.Item>
          <Form.Item label={t('platform_skills_form_source_type')} name="sourceType">
            <Select
              options={[
                { label: translateSourceType('builtin'), value: 'builtin' },
                { label: translateSourceType('partner'), value: 'partner' },
                { label: translateSourceType('openapi_catalog'), value: 'openapi_catalog' },
                { label: translateSourceType('mcp_catalog'), value: 'mcp_catalog' },
              ]}
            />
          </Form.Item>
          <Form.Item label={t('platform_skills_form_provider_type')} name="providerType">
            <Select
              options={[
                { label: translateProviderType('native'), value: 'native' },
                { label: translateProviderType('openapi'), value: 'openapi' },
                { label: translateProviderType('mcp'), value: 'mcp' },
                { label: translateProviderType('remote_hosted'), value: 'remote_hosted' },
              ]}
            />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={editingHub ? t('platform_skills_edit_hub') : t('platform_skills_new_hub')}
        open={hubOpen}
        onCancel={() => {
          setHubOpen(false);
          setEditingHub(null);
        }}
        onOk={handleSaveHub}
        confirmLoading={saving}
        width={900}
        okText={t('agent_save')}
        cancelText={t('agent_cancel')}
        styles={scrollableModalStyles}
        destroyOnHidden
      >
        <Form layout="vertical" form={hubForm}>
          <Form.Item label={t('platform_skill_hubs_form_code')} name="hubCode" rules={[{ required: true, message: t('platform_skill_hubs_form_code_required') }]}>
            <Input placeholder="partner-openapi" />
          </Form.Item>
          <Form.Item label={t('platform_skills_form_name')} name="name" rules={[{ required: true, message: t('platform_skills_form_name_required') }]}>
            <Input placeholder="Partner OpenAPI Hub" />
          </Form.Item>
          <Form.Item label={t('platform_skill_hubs_form_type')} name="hubType" rules={[{ required: true, message: t('platform_skill_hubs_form_type_required') }]}>
            <Select
              options={[
                { label: translateHubType('openapi_hub'), value: 'openapi_hub' },
                { label: translateHubType('mcp_hub'), value: 'mcp_hub' },
                { label: translateHubType('builtin_hub'), value: 'builtin_hub' },
                { label: translateHubType('enterprise_private_hub'), value: 'enterprise_private_hub' },
              ]}
            />
          </Form.Item>
          <Form.Item label={t('platform_skill_hubs_form_base_url')} name="baseUrl">
            <Input placeholder="https://hub.example.com" />
          </Form.Item>
          <Space size={12} style={{ width: '100%' }} wrap>
            <Form.Item label={t('platform_skills_column_status')} name="status" style={{ minWidth: 180 }}>
              <Select
                options={[
                  { label: translateStatus('enabled'), value: 'enabled' },
                  { label: translateStatus('disabled'), value: 'disabled' },
                ]}
              />
            </Form.Item>
            <Form.Item label={t('platform_skills_column_trust_level')} name="trustLevel" style={{ minWidth: 220 }}>
              <Select
                options={[
                  { label: translateTrustLevel('platform_trusted'), value: 'platform_trusted' },
                  { label: translateTrustLevel('partner_verified'), value: 'partner_verified' },
                  { label: translateTrustLevel('enterprise_verified'), value: 'enterprise_verified' },
                  { label: translateTrustLevel('unverified'), value: 'unverified' },
                  { label: translateTrustLevel('blocked'), value: 'blocked' },
                ]}
              />
            </Form.Item>
            <Form.Item label={t('platform_skill_hubs_form_sync_mode')} name="syncMode" style={{ minWidth: 180 }}>
              <Select
                options={[
                  { label: translateSyncMode('manual'), value: 'manual' },
                  { label: translateSyncMode('scheduled'), value: 'scheduled' },
                ]}
              />
            </Form.Item>
            <Form.Item label={t('platform_skill_hubs_form_auth_scheme')} name="authScheme" style={{ minWidth: 180 }}>
              <Select
                options={[
                  { label: translateAuthScheme('none'), value: 'none' },
                  { label: translateAuthScheme('api_key'), value: 'api_key' },
                  { label: translateAuthScheme('oauth2'), value: 'oauth2' },
                  { label: translateAuthScheme('oidc'), value: 'oidc' },
                ]}
              />
            </Form.Item>
          </Space>
          <Form.Item label={t('platform_skill_hubs_form_config_json')} name="config">
            <TextArea rows={3} />
          </Form.Item>
          <Form.Item label={t('platform_skill_hubs_form_secret_json')} name="secret">
            <TextArea rows={3} />
          </Form.Item>
          <Form.Item label={t('platform_skill_hubs_form_import_policy_json')} name="importPolicy">
            <TextArea rows={3} />
          </Form.Item>
          <Form.Item label={t('platform_skill_hubs_form_allowed_namespaces_json')} name="allowedNamespaces">
            <TextArea rows={3} />
          </Form.Item>
          <Form.Item label={t('platform_skill_hubs_form_network_policy_json')} name="networkPolicy">
            <TextArea rows={3} />
          </Form.Item>
          <Form.Item label={t('platform_skill_hubs_form_signature_policy_json')} name="signaturePolicy">
            <TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={t('platform_skills_import_start')}
        open={importOpen}
        onCancel={() => setImportOpen(false)}
        onOk={handleImportSkill}
        confirmLoading={saving}
        okText={t('agent_save')}
        cancelText={t('agent_cancel')}
        styles={scrollableModalStyles}
        destroyOnHidden
      >
        <Form layout="vertical" form={importForm}>
          <Form.Item label={t('platform_skill_import_jobs_form_hub')} name="hubId" rules={[{ required: true, message: t('platform_skill_import_jobs_form_hub_required') }]}>
            <Select
              placeholder={t('platform_skill_import_jobs_form_hub_placeholder')}
              options={hubs.map((item) => ({
                label: `${item.name} (${item.hubCode})`,
                value: item.id,
              }))}
            />
          </Form.Item>
          <Form.Item label={t('platform_skill_import_jobs_form_source_locator')} name="sourceLocator" rules={[{ required: true, message: t('platform_skill_import_jobs_form_source_locator_required') }]}>
            <Input placeholder="petstore/openapi.yaml" />
          </Form.Item>
          <Form.Item label={t('platform_skill_import_jobs_form_source_namespace')} name="sourceNamespace">
            <Input placeholder="partner.petstore" />
          </Form.Item>
          <Form.Item label={t('platform_skill_import_jobs_form_source_version')} name="sourceVersion">
            <Input placeholder="1.0.0" />
          </Form.Item>
        </Form>
      </Modal>

      <Drawer
        title={selectedSkill?.skill?.name || t('platform_skill_detail_title')}
        open={detailOpen}
        onClose={() => setDetailOpen(false)}
        width={960}
        extra={
          <Button
            icon={<PlusOutlined />}
            onClick={() => {
              versionForm.setFieldsValue({
                manifest: '{}',
                inputSchema: '{}',
                outputSchema: '{}',
                defaultPolicy: '{}',
                runtimeContract: '{}',
                references: '[]',
              });
              setVersionOpen(true);
            }}
          >
            {t('platform_skills_new_version')}
          </Button>
        }
      >
        {selectedSkill?.skill ? (
          <Space orientation="vertical" size={16} style={{ width: '100%' }}>
            <Card size="small">
              <Space size={12} wrap>
                <Text strong>{selectedSkill.skill.code}</Text>
                {renderTrustTag(selectedSkill.skill.trustLevel, translateTrustLevel(selectedSkill.skill.trustLevel))}
                {renderStatusTag(selectedSkill.skill.status, translateStatus(selectedSkill.skill.status))}
              </Space>
              <Paragraph type="secondary" style={{ marginTop: 12, marginBottom: 0 }}>
                {selectedSkill.skill.description || t('platform_skill_detail_no_description')}
              </Paragraph>
            </Card>
            <Card size="small" title={t('platform_skill_detail_versions')}>
              <Table
                rowKey="id"
                pagination={false}
                dataSource={selectedSkill.versions}
                locale={{
                  emptyText: <Empty description={t('platform_skill_detail_versions_empty')} image={Empty.PRESENTED_IMAGE_SIMPLE} />,
                }}
                columns={[
                  { title: t('platform_skill_detail_version_column_version'), dataIndex: 'version', key: 'version', width: 120 },
                  {
                    title: t('platform_skill_detail_version_column_release_channel'),
                    dataIndex: 'releaseChannel',
                    key: 'releaseChannel',
                    width: 120,
                    render: (value: string) => t(`platform_skills_release_channel_${value}`, value),
                  },
                  {
                    title: t('platform_skills_column_status'),
                    dataIndex: 'releaseStatus',
                    key: 'releaseStatus',
                    width: 120,
                    render: (value: string) => renderStatusTag(value, translateStatus(value)),
                  },
                  {
                    title: t('platform_skill_detail_version_column_change_log'),
                    dataIndex: 'changeLog',
                    key: 'changeLog',
                    render: (value: string) => value || '-',
                  },
                  {
                    title: t('platform_skill_detail_version_column_published_at'),
                    dataIndex: 'publishedAt',
                    key: 'publishedAt',
                    width: 180,
                    render: (value?: string) => formatDateTime(value, i18n.resolvedLanguage || i18n.language),
                  },
                  {
                    title: t('platform_skills_column_actions'),
                    key: 'actions',
                    width: 220,
                    render: (_: unknown, record: SkillVersionItem) => (
                      <Space>
                        {record.releaseStatus === 'draft' ? (
                          <Button size="small" onClick={() => handleSubmitReview(record.id)}>
                            {t('platform_skills_submit_review')}
                          </Button>
                        ) : null}
                        {record.releaseStatus === 'draft' || record.releaseStatus === 'reviewing' ? (
                          <Button size="small" onClick={() => void openReferenceEditor(record)}>
                            {t('platform_skill_references_edit')}
                          </Button>
                        ) : null}
                        {record.releaseStatus === 'draft' || record.releaseStatus === 'reviewing' ? (
                          <Button size="small" type="primary" loading={publishingId === record.id} onClick={() => handlePublish(record.id)}>
                            {t('platform_skills_publish')}
                          </Button>
                        ) : null}
                      </Space>
                    ),
                  },
                ]}
              />
            </Card>
            <Card size="small" title={t('platform_skill_detail_references')}>
              <Table
                rowKey="id"
                pagination={false}
                dataSource={selectedSkill.references || []}
                locale={{
                  emptyText: <Empty description={t('platform_skill_detail_references_empty')} image={Empty.PRESENTED_IMAGE_SIMPLE} />,
                }}
                columns={[
                  {
                    title: t('platform_skill_detail_reference_column_target_version'),
                    dataIndex: 'toSkillVersionId',
                    key: 'toSkillVersionId',
                    width: 220,
                    render: (value: string) => value || '-',
                  },
                  {
                    title: t('platform_skill_detail_reference_column_invoke_mode'),
                    dataIndex: 'invokeMode',
                    key: 'invokeMode',
                    width: 120,
                    render: (value: string) => translateInvokeMode(value),
                  },
                  {
                    title: t('platform_skill_detail_reference_column_condition_expr'),
                    dataIndex: 'conditionExpr',
                    key: 'conditionExpr',
                    render: (value: string) => value || '-',
                  },
                  {
                    title: t('platform_skill_detail_reference_column_passthrough'),
                    key: 'passthrough',
                    width: 180,
                    render: (_: unknown, record: SkillReferenceItem) => (
                      <Space wrap>
                        <Tag color={record.contextPassthrough ? 'processing' : 'default'}>{t('platform_skill_detail_reference_passthrough_context')}</Tag>
                        <Tag color={record.resultPassthrough ? 'processing' : 'default'}>{t('platform_skill_detail_reference_passthrough_result')}</Tag>
                      </Space>
                    ),
                  },
                ]}
              />
            </Card>
          </Space>
        ) : null}
      </Drawer>

      <Modal
        title={t('platform_skills_new_version')}
        open={versionOpen}
        onCancel={() => setVersionOpen(false)}
        onOk={handleCreateVersion}
        confirmLoading={saving}
        width={820}
        okText={t('agent_save')}
        cancelText={t('agent_cancel')}
        styles={scrollableModalStyles}
        destroyOnHidden
      >
        <Form layout="vertical" form={versionForm}>
          <Form.Item label={t('platform_skill_version_form_version')} name="version" rules={[{ required: true, message: t('platform_skill_version_form_version_required') }]}>
            <Input placeholder="1.0.0" />
          </Form.Item>
          <Form.Item label={t('platform_skill_version_form_manifest')} name="manifest">
            <TextArea rows={4} />
          </Form.Item>
          <Form.Item label={t('platform_skill_version_form_input_schema')} name="inputSchema">
            <TextArea rows={4} />
          </Form.Item>
          <Form.Item label={t('platform_skill_version_form_output_schema')} name="outputSchema">
            <TextArea rows={4} />
          </Form.Item>
          <Form.Item label={t('platform_skill_version_form_default_policy')} name="defaultPolicy">
            <TextArea rows={4} />
          </Form.Item>
          <Form.Item label={t('platform_skill_version_form_runtime_contract')} name="runtimeContract">
            <TextArea rows={4} />
          </Form.Item>
          <Form.Item label={t('platform_skill_version_form_references')} name="references">
            <TextArea rows={4} placeholder='[{"toSkillVersionId":"version-2","invokeMode":"sync"}]' />
          </Form.Item>
          <Form.Item label={t('platform_skill_version_form_change_log')} name="changeLog">
            <TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={t('platform_skill_references_edit')}
        open={referenceEditorOpen}
        onCancel={() => {
          setReferenceEditorOpen(false);
          setEditingVersion(null);
          referenceEditorForm.resetFields();
        }}
        onOk={handleUpdateReferences}
        confirmLoading={saving || referenceLoading}
        okText={t('agent_save')}
        cancelText={t('agent_cancel')}
        styles={scrollableModalStyles}
        destroyOnHidden
      >
        <Form layout="vertical" form={referenceEditorForm}>
          <Form.Item label={t('platform_skill_version_form_references')} name="references">
            <TextArea rows={8} placeholder='[{"toSkillVersionId":"version-2","invokeMode":"sync"}]' />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

function safeParseJSON(value?: string) {
  if (!value || !value.trim()) {
    return {};
  }
  try {
    return JSON.parse(value);
  } catch {
    return {};
  }
}

function safeParseJSONArray(value?: string) {
  if (!value || !value.trim()) {
    return [];
  }
  try {
    const parsed = JSON.parse(value);
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

function parseJSONArrayOrThrow(value?: string) {
  if (!value || !value.trim()) {
    return [];
  }
  const parsed = JSON.parse(value);
  if (!Array.isArray(parsed)) {
    throw new Error('array');
  }
  return parsed;
}

export default PlatformSkillsPage;
