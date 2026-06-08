import React, { useEffect, useMemo, useRef, useState } from 'react';
import {
  Button,
  Card,
  Drawer,
  Empty,
  Form,
  Input,
  Modal,
  Radio,
  Select,
  Space,
  Table,
  Tag,
  Typography,
  message,
} from 'antd';
import { ArrowRightOutlined, CloudDownloadOutlined, PlusOutlined, SafetyOutlined, SearchOutlined } from '@ant-design/icons';
import axios from 'axios';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { BACKEND_URL } from '../../config';
import { getLocalizedPath, resolveSupportedLanguage } from '../../i18n/config';
import { casdoorService } from '../identity/CasdoorService';

const { Title, Paragraph, Text } = Typography;
const { TextArea } = Input;
const ADMIN_SKILLS_API_BASE = `${BACKEND_URL}/api/admin/skills`;
const CURRENT_ENTERPRISE_STORAGE_KEY = 'dotblue_current_enterprise_id';
const scrollableModalStyles = {
  body: {
    maxHeight: '70vh',
    overflowY: 'auto' as const,
    paddingRight: 8,
  },
};

type PlatformTabKey = 'skills' | 'hubs' | 'imports';
type PlatformSkillsExperience = 'governance' | 'market';

interface PlatformSkillsPageProps {
  defaultTab?: PlatformTabKey;
  experience?: PlatformSkillsExperience;
  openFlow?: 'skill' | 'hub' | 'import';
}

interface MarketFilterOption {
  label: string;
  value: string;
}

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

interface SkillResourceReleaseItem {
  id: string;
  resourceType: string;
  resourceId: string;
  releaseScope: string;
  targetEnterpriseId?: string;
  releaseStatus: string;
  note?: string;
  operatedBy?: string;
  createdAt?: string;
  updatedAt?: string;
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

interface ResourceReleaseFormValues {
  releaseScope: 'global' | 'enterprise';
  targetEnterpriseId?: string;
  releaseStatus: 'enabled' | 'disabled';
  note?: string;
}

interface AgentInstallTarget {
  id: string;
  agentName: string;
  engineType: 'hermes' | 'nanobot';
  modelName?: string;
}

interface ReleaseTargetState {
  resourceId: string;
  resourceType: 'skill' | 'hub';
  resourceName: string;
}

interface InstallToAgentFormValues {
  agentId: string;
  entryAlias?: string;
  invokeVisibility: 'auto' | 'suggested' | 'manual';
}

interface InstallSuccessState {
  agentId: string;
  agentName: string;
  skillName: string;
}

interface ImportRolloutState {
  skillId: string;
  skillVersionId?: string;
  sourceLocator: string;
  sourceNamespace?: string;
  sourceVersion?: string;
}

interface QuickImportTemplate {
  key: string;
  label: string;
  description: string;
  locator: string;
  namespace?: string;
  version?: string;
}

function getAuthHeaders() {
  const token = casdoorService.getToken();
  const enterpriseId = localStorage.getItem(CURRENT_ENTERPRISE_STORAGE_KEY)?.trim();
  const headers: Record<string, string> = {};
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }
  if (enterpriseId) {
    headers['X-Enterprise-ID'] = enterpriseId;
  }
  return headers;
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

function renderStatusTag(status: string, label: string, fallbackLabel = '-') {
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
  return <Tag color={color}>{label || status || fallbackLabel}</Tag>;
}

function renderTrustTag(trustLevel: string, label: string, fallbackLabel = '-') {
  const normalized = (trustLevel || '').toLowerCase();
  const color = normalized.includes('trusted')
    ? 'success'
    : normalized.includes('verified')
      ? 'processing'
      : normalized.includes('blocked')
        ? 'error'
        : 'default';
  return <Tag color={color}>{label || trustLevel || fallbackLabel}</Tag>;
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
  return errorKey ? t(errorKey, { defaultValue: fallbackMessage }) : fallbackMessage;
}

function renderSummaryCards(items: Array<{ label: string; value: number; active?: boolean; onClick?: () => void }>) {
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
        <Card
          key={item.label}
          size="small"
          hoverable={Boolean(item.onClick)}
          onClick={item.onClick}
          style={{
            cursor: item.onClick ? 'pointer' : 'default',
            borderColor: item.active ? '#1677ff' : undefined,
            boxShadow: item.active ? '0 0 0 1px rgba(22, 119, 255, 0.15)' : undefined,
          }}
        >
          <Text type="secondary">{item.label}</Text>
          <Title level={4} style={{ margin: '8px 0 0' }}>{item.value}</Title>
        </Card>
      ))}
    </div>
  );
}

const PlatformSkillsPage: React.FC<PlatformSkillsPageProps> = ({
  defaultTab = 'skills',
  experience = 'governance',
  openFlow,
}) => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const currentLanguage = resolveSupportedLanguage(i18n?.resolvedLanguage || i18n?.language);
  const emptyPlaceholder = t('common_empty_placeholder');
  const platformSkillFormExamples = useMemo(() => ({
    skillCode: t('platform_skills_form_code_example'),
    hubCode: t('platform_skill_hubs_form_code_example'),
    hubName: t('platform_skill_hubs_form_name_example'),
    hubBaseUrl: t('platform_skill_hubs_form_base_url_example'),
    version: t('platform_skill_version_form_version_example'),
    referencesJson: t('platform_skill_version_form_references_example'),
  }), [t]);
  const [messageApi, contextHolder] = message.useMessage();
  const [activeTab, setActiveTab] = useState<PlatformTabKey>(defaultTab);
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
  const [releaseOpen, setReleaseOpen] = useState(false);
  const [installToAgentOpen, setInstallToAgentOpen] = useState(false);
  const [installSuccess, setInstallSuccess] = useState<InstallSuccessState | null>(null);
  const [importRollout, setImportRollout] = useState<ImportRolloutState | null>(null);
  const [saving, setSaving] = useState(false);
  const [referenceLoading, setReferenceLoading] = useState(false);
  const [publishingId, setPublishingId] = useState<string>('');
  const [selectedSkill, setSelectedSkill] = useState<SkillDetail | null>(null);
  const [installSkill, setInstallSkill] = useState<SkillItem | null>(null);
  const [installTargets, setInstallTargets] = useState<AgentInstallTarget[]>([]);
  const [installTargetsLoading, setInstallTargetsLoading] = useState(false);
  const [editingHub, setEditingHub] = useState<SkillHubItem | null>(null);
  const [releaseTarget, setReleaseTarget] = useState<ReleaseTargetState | null>(null);
  const [releaseRecords, setReleaseRecords] = useState<SkillResourceReleaseItem[]>([]);
  const [releaseRecordsLoading, setReleaseRecordsLoading] = useState(false);
  const [createForm] = Form.useForm<SkillFormValues>();
  const [versionForm] = Form.useForm<SkillVersionFormValues>();
  const [referenceEditorForm] = Form.useForm<ReferenceEditorFormValues>();
  const [hubForm] = Form.useForm<SkillHubFormValues>();
  const [importForm] = Form.useForm<ImportJobFormValues>();
  const [releaseForm] = Form.useForm<ResourceReleaseFormValues>();
  const [installToAgentForm] = Form.useForm<InstallToAgentFormValues>();
  const [editingVersion, setEditingVersion] = useState<SkillVersionItem | null>(null);
  const [searchText, setSearchText] = useState('');
  const [skillSourceFilter, setSkillSourceFilter] = useState<string>('all');
  const [skillStatusFilter, setSkillStatusFilter] = useState<string>('all');
  const [skillSummaryFilter, setSkillSummaryFilter] = useState<string>('all');
  const [hubTypeFilter, setHubTypeFilter] = useState<string>('all');
  const [hubSummaryFilter, setHubSummaryFilter] = useState<string>('all');
  const [importStatusFilter, setImportStatusFilter] = useState<string>('all');
  const [importSummaryFilter, setImportSummaryFilter] = useState<string>('all');
  const initialRouteBehaviorApplied = useRef(false);
  const selectedImportHubId = Form.useWatch('hubId', importForm);
  const selectedReleaseScope = Form.useWatch('releaseScope', releaseForm);

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

  useEffect(() => {
    if (initialRouteBehaviorApplied.current) {
      return;
    }

    setActiveTab(defaultTab);
    if (openFlow === 'skill') {
      openCreateSkill();
    } else if (openFlow === 'hub') {
      setActiveTab('hubs');
      openCreateHub();
    } else if (openFlow === 'import') {
      setActiveTab('imports');
      openImportJob();
    }

    initialRouteBehaviorApplied.current = true;
  }, [defaultTab, openFlow]);

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

  const normalizedSearch = searchText.trim().toLowerCase();
  const matchesSearch = (...values: Array<string | undefined>) => {
    if (!normalizedSearch) {
      return true;
    }
    return values.some((value) => (value || '').toLowerCase().includes(normalizedSearch));
  };

  const filteredSkills = useMemo(() => skills.filter((item) => {
    const matchesKeyword = matchesSearch(
      item.code,
      item.name,
      item.description,
      item.sourceType,
      item.providerType,
      item.status,
      item.trustLevel,
    );
    const matchesSource = skillSourceFilter === 'all' || item.sourceType === skillSourceFilter;
    const matchesStatus = skillStatusFilter === 'all' || item.status === skillStatusFilter;
    const matchesSummary = skillSummaryFilter === 'all' || item.status === skillSummaryFilter;
    return matchesKeyword && matchesSource && matchesStatus && matchesSummary;
  }), [skills, normalizedSearch, skillSourceFilter, skillStatusFilter, skillSummaryFilter]);

  const filteredHubs = useMemo(() => hubs.filter((item) => {
    const matchesKeyword = matchesSearch(
      item.hubCode,
      item.name,
      item.hubType,
      item.baseUrl,
      item.status,
      item.trustLevel,
    );
    const matchesType = hubTypeFilter === 'all' || item.hubType === hubTypeFilter;
    const matchesSummary = hubSummaryFilter === 'all'
      || (hubSummaryFilter === 'enabled' && item.status === 'enabled')
      || (hubSummaryFilter === 'openapi' && item.hubType === 'openapi_hub')
      || (hubSummaryFilter === 'mcp' && item.hubType === 'mcp_hub');
    return matchesKeyword && matchesType && matchesSummary;
  }), [hubs, normalizedSearch, hubTypeFilter, hubSummaryFilter]);

  const filteredImportJobs = useMemo(() => importJobs.filter((item) => {
    const matchesKeyword = matchesSearch(
      hubNameMap[item.hubId],
      item.sourceLocator,
      item.sourceNamespace,
      item.sourceVersion,
      item.jobStatus,
      item.errorMessage,
    );
    const matchesStatus = importStatusFilter === 'all' || item.jobStatus === importStatusFilter;
    const matchesSummary = importSummaryFilter === 'all'
      || (importSummaryFilter === 'completed' && item.jobStatus === 'completed')
      || (importSummaryFilter === 'running' && item.jobStatus === 'normalizing')
      || (importSummaryFilter === 'failed' && item.jobStatus === 'failed');
    return matchesKeyword && matchesStatus && matchesSummary;
  }), [hubNameMap, importJobs, normalizedSearch, importStatusFilter, importSummaryFilter]);

  const skillSourceOptions = useMemo<MarketFilterOption[]>(() => (
    [
      { label: t('platform_skill_market_filter_all'), value: 'all' },
      ...Array.from(new Set(skills.map((item) => item.sourceType).filter(Boolean))).map((value) => ({
        label: translateSourceType(value),
        value,
      })),
    ]
  ), [skills, t]);

  const skillStatusOptions = useMemo<MarketFilterOption[]>(() => (
    [
      { label: t('platform_skill_market_filter_all_status'), value: 'all' },
      ...Array.from(new Set(skills.map((item) => item.status).filter(Boolean))).map((value) => ({
        label: translateStatus(value),
        value,
      })),
    ]
  ), [skills, t]);

  const hubTypeOptions = useMemo<MarketFilterOption[]>(() => (
    [
      { label: t('platform_skill_market_filter_all_sources'), value: 'all' },
      ...Array.from(new Set(hubs.map((item) => item.hubType).filter(Boolean))).map((value) => ({
        label: translateHubType(value),
        value,
      })),
    ]
  ), [hubs, t]);

  const importStatusOptions = useMemo<MarketFilterOption[]>(() => (
    [
      { label: t('platform_skill_market_filter_all_imports'), value: 'all' },
      ...Array.from(new Set(importJobs.map((item) => item.jobStatus).filter(Boolean))).map((value) => ({
        label: translateStatus(value),
        value,
      })),
    ]
  ), [importJobs, t]);

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

  const openImportJob = (preset?: Partial<ImportJobFormValues>) => {
    const defaultHubId = preset?.hubId || (hubs.length === 1 ? hubs[0].id : '');
    importForm.resetFields();
    importForm.setFieldsValue({
      hubId: defaultHubId,
      sourceLocator: '',
      sourceNamespace: '',
      sourceVersion: '',
      ...preset,
    });
    setImportOpen(true);
  };

  const openReleaseSettings = (target: ReleaseTargetState) => {
    setReleaseTarget(target);
    setReleaseRecords([]);
    releaseForm.resetFields();
    releaseForm.setFieldsValue({
      releaseScope: 'global',
      targetEnterpriseId: '',
      releaseStatus: 'enabled',
      note: '',
    });
    setReleaseOpen(true);
    void fetchReleaseRecords(target);
  };

  const fetchReleaseRecords = async (target: ReleaseTargetState) => {
    setReleaseRecordsLoading(true);
    try {
      const res = await axios.get(`${BACKEND_URL}/api/admin/platform/resource-releases`, {
        headers: getAuthHeaders(),
        params: {
          resourceType: target.resourceType,
          resourceId: target.resourceId,
        },
      });
      setReleaseRecords(Array.isArray(res.data) ? res.data : []);
    } catch (error: any) {
      const errorText = error?.response?.data;
      messageApi.error(
        translatePlatformSkillError(errorText, t('platform_resource_release_load_failed'), t),
      );
      setReleaseRecords([]);
    } finally {
      setReleaseRecordsLoading(false);
    }
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
      messageApi.error(
        translatePlatformSkillError(errorText, t('platform_skill_hub_save_failed'), t),
      );
    } finally {
      setSaving(false);
    }
  };

  const handleImportSkill = async () => {
    const values = await importForm.validateFields();
    setSaving(true);
    try {
      const res = await axios.post(`${BACKEND_URL}/api/admin/platform/skill-import-jobs`, values, {
        headers: getAuthHeaders(),
      });
      messageApi.success(t('platform_skill_import_create_success'));
      setImportOpen(false);
      importForm.resetFields();
      if (res.data?.targetSkillId) {
        setImportRollout({
          skillId: res.data.targetSkillId,
          skillVersionId: res.data.targetSkillVersionId,
          sourceLocator: values.sourceLocator,
          sourceNamespace: values.sourceNamespace,
          sourceVersion: values.sourceVersion,
        });
      }
      await Promise.all([fetchImportJobs(), fetchSkills()]);
    } catch (error: any) {
      const errorText = error?.response?.data;
      messageApi.error(
        translatePlatformSkillError(errorText, t('platform_skill_import_create_failed'), t),
      );
    } finally {
      setSaving(false);
    }
  };

  const handleSaveRelease = async () => {
    if (!releaseTarget) {
      return;
    }
    const values = await releaseForm.validateFields();
    setSaving(true);
    try {
      await axios.post(`${BACKEND_URL}/api/admin/platform/resource-releases`, {
        resourceId: releaseTarget.resourceId,
        resourceType: releaseTarget.resourceType,
        releaseScope: values.releaseScope,
        targetEnterpriseId: values.releaseScope === 'enterprise' ? (values.targetEnterpriseId || '') : '',
        releaseStatus: values.releaseStatus,
        note: values.note || '',
      }, {
        headers: getAuthHeaders(),
      });
      messageApi.success(t('platform_resource_release_success'));
      await fetchReleaseRecords(releaseTarget);
      releaseForm.setFieldsValue({
        releaseScope: values.releaseScope,
        targetEnterpriseId: values.releaseScope === 'enterprise' ? (values.targetEnterpriseId || '') : '',
        releaseStatus: values.releaseStatus,
        note: values.note || '',
      });
    } catch (error: any) {
      const errorText = error?.response?.data;
      messageApi.error(
        translatePlatformSkillError(errorText, t('platform_resource_release_failed'), t),
      );
    } finally {
      setSaving(false);
    }
  };

  const fetchInstallTargets = async () => {
    setInstallTargetsLoading(true);
    try {
      const res = await axios.get(`${BACKEND_URL}/api/agents`, {
        headers: getAuthHeaders(),
      });
      setInstallTargets(Array.isArray(res.data) ? res.data : []);
    } catch {
      messageApi.error(t('platform_skill_install_agent_targets_load_failed'));
      setInstallTargets([]);
    } finally {
      setInstallTargetsLoading(false);
    }
  };

  const openInstallToAgent = async (skill: SkillItem) => {
    setInstallSkill(skill);
    setInstallToAgentOpen(true);
    installToAgentForm.resetFields();
    installToAgentForm.setFieldsValue({ invokeVisibility: 'auto' });
    await fetchInstallTargets();
  };

  const handleInstallToAgent = async () => {
    if (!installSkill?.id) {
      return;
    }
    const values = await installToAgentForm.validateFields();
    setSaving(true);
    try {
      await axios.post(`${BACKEND_URL}/api/admin/agents/${values.agentId}/skills/install`, {
        skillId: installSkill.id,
        entryAlias: values.entryAlias,
        invokeVisibility: values.invokeVisibility,
      }, {
        headers: getAuthHeaders(),
      });
      const targetAgent = installTargets.find((item) => item.id === values.agentId);
      messageApi.success(t('platform_skill_install_agent_success', { name: installSkill.name }));
      setInstallToAgentOpen(false);
      setInstallSuccess({
        agentId: values.agentId,
        agentName: targetAgent?.agentName || values.agentId,
        skillName: installSkill.name,
      });
    } catch (error: any) {
      const errorText = error?.response?.data;
      messageApi.error(
        translatePlatformSkillError(errorText, t('platform_skill_install_agent_failed'), t),
      );
    } finally {
      setSaving(false);
    }
  };

  const selectedImportHub = useMemo(
    () => hubs.find((item) => item.id === selectedImportHubId) || null,
    [hubs, selectedImportHubId],
  );

  const quickImportTemplates = useMemo<QuickImportTemplate[]>(() => {
    const hubType = selectedImportHub?.hubType || '';
    const tencentLocator = 'https://skillhub.cn/skills/weather';
    const openapiLocator = 'petstore/openapi.yaml';
    const privateLocator = 'internal/knowledge-search';

    if (hubType === 'tencent_skillhub') {
      return [
        {
          key: 'weather',
          label: t('platform_skill_import_template_weather_title'),
          description: t('platform_skill_import_template_weather_desc'),
          locator: tencentLocator,
          namespace: 'weather',
          version: 'latest',
        },
        {
          key: 'knowledge',
          label: t('platform_skill_import_template_knowledge_title'),
          description: t('platform_skill_import_template_knowledge_desc'),
          locator: 'https://skillhub.cn/skills/knowledge-search',
          namespace: 'knowledge-search',
          version: 'latest',
        },
      ];
    }

    if (hubType === 'enterprise_private_hub') {
      return [
        {
          key: 'private-knowledge',
          label: t('platform_skill_import_template_private_title'),
          description: t('platform_skill_import_template_private_desc'),
          locator: privateLocator,
          namespace: 'internal.knowledge',
          version: 'v1',
        },
      ];
    }

    return [
      {
        key: 'openapi-petstore',
        label: t('platform_skill_import_template_openapi_title'),
        description: t('platform_skill_import_template_openapi_desc'),
        locator: openapiLocator,
        namespace: 'partner.petstore',
        version: '1.0.0',
      },
      {
        key: 'weather',
        label: t('platform_skill_import_template_weather_title'),
        description: t('platform_skill_import_template_weather_desc'),
        locator: selectedImportHub ? tencentLocator : 'weather',
        namespace: 'weather',
        version: 'latest',
      },
    ];
  }, [selectedImportHub, t]);

  const applyImportTemplate = (template: QuickImportTemplate) => {
    importForm.setFieldsValue({
      sourceLocator: template.locator,
      sourceNamespace: template.namespace,
      sourceVersion: template.version,
    });
  };

  const renderImportGuideCard = () => (
    <Card
      size="small"
      style={{
        marginBottom: 16,
        borderRadius: 20,
        background: 'linear-gradient(135deg, rgba(22,119,255,0.08), rgba(22,119,255,0.02))',
        border: '1px solid rgba(22,119,255,0.12)',
      }}
    >
      <Space orientation="vertical" size={10} style={{ width: '100%' }}>
        <Text strong>{t('platform_skill_import_guide_title')}</Text>
        <Paragraph type="secondary" style={{ marginBottom: 0 }}>
          {t('platform_skill_import_guide_desc')}
        </Paragraph>
        <Space wrap size={[8, 8]}>
          <Tag color={hubSummary.total > 0 ? 'success' : 'default'}>{t('platform_skill_import_guide_step_hub')}</Tag>
          <Tag color="processing">{t('platform_skill_import_guide_step_search')}</Tag>
          <Tag>{t('platform_skill_import_guide_step_import')}</Tag>
        </Space>
        <Text type="secondary">
          {selectedImportHub
            ? t('platform_skill_import_guide_selected_hub', {
              name: selectedImportHub.name,
              type: translateHubType(selectedImportHub.hubType),
            })
            : t('platform_skill_import_guide_pick_hub')}
        </Text>
        {selectedImportHub ? (
          <Space wrap size={[8, 8]}>
            <Tag>{translateHubType(selectedImportHub.hubType)}</Tag>
            {renderTrustTag(selectedImportHub.trustLevel, translateTrustLevel(selectedImportHub.trustLevel))}
            {selectedImportHub.baseUrl ? <Tag>{selectedImportHub.baseUrl}</Tag> : null}
          </Space>
        ) : null}
        <Space wrap>
          {quickImportTemplates.map((template) => (
            <Button key={template.key} onClick={() => applyImportTemplate(template)}>
              {template.label}
            </Button>
          ))}
          <Button
            type="link"
            style={{ paddingInline: 0 }}
            onClick={() => {
              setImportOpen(false);
              setActiveTab('hubs');
              openCreateHub();
            }}
          >
            {t('platform_skill_import_guide_action_manage_sources')}
          </Button>
        </Space>
      </Space>
    </Card>
  );

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
        <Button type="primary" icon={<CloudDownloadOutlined />} onClick={() => openImportJob()}>
          {t('platform_skills_import_start')}
        </Button>
      );

  const isMarketExperience = experience === 'market';
  const pageTitle = isMarketExperience ? t('platform_skill_market_title') : t('platform_skill_governance_title');
  const pageDescription = isMarketExperience ? t('platform_skill_market_desc') : t('platform_skill_governance_desc');
  const searchPlaceholder = isMarketExperience
    ? t('platform_skill_market_search_placeholder')
    : t('platform_skill_governance_search_placeholder');
  const tabItems = isMarketExperience
    ? [
      { key: 'skills', label: `${t('platform_skill_market_tab_catalog')} (${skillSummary.total})` },
      { key: 'hubs', label: `${t('platform_skill_market_tab_sources')} (${hubSummary.total})` },
      { key: 'imports', label: `${t('platform_skill_market_tab_imports')} (${importSummary.total})` },
    ]
    : [
      { key: 'skills', label: `${t('platform_skills_tab_skills')} (${skillSummary.total})` },
      { key: 'hubs', label: `${t('platform_skills_tab_hubs')} (${hubSummary.total})` },
      { key: 'imports', label: `${t('platform_skills_tab_imports')} (${importSummary.total})` },
    ];

  const renderMarketEmpty = (
    title: string,
    description: string,
    actionLabel: string,
    action: () => void,
  ) => (
    <Card style={{ borderRadius: 20 }}>
      <Empty
        image={Empty.PRESENTED_IMAGE_SIMPLE}
        description={
          <Space orientation="vertical" size={4}>
            <Text strong>{title}</Text>
            <Text type="secondary">{description}</Text>
          </Space>
        }
      >
        <Button type="primary" onClick={action}>{actionLabel}</Button>
      </Empty>
    </Card>
  );

  const renderMarketSkillCards = () => {
    if (!filteredSkills.length) {
      return renderMarketEmpty(
        t('platform_skill_market_empty_skills_title'),
        t('platform_skill_market_empty_skills_desc'),
        t('platform_skill_market_action_import'),
        openImportJob,
      );
    }

    return (
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: 16 }}>
        {filteredSkills.map((item) => (
          <Card key={item.id} hoverable style={{ borderRadius: 20 }}>
            <Space orientation="vertical" size={14} style={{ width: '100%' }}>
              <Space align="start" style={{ width: '100%', justifyContent: 'space-between' }}>
                <div>
                  <Text type="secondary">{item.code}</Text>
                  <Title level={5} style={{ margin: '6px 0 0' }}>{item.name}</Title>
                </div>
                {renderStatusTag(item.status, translateStatus(item.status), emptyPlaceholder)}
              </Space>
              <Paragraph
                type="secondary"
                ellipsis={{ rows: 3, expandable: false }}
                style={{ minHeight: 66, marginBottom: 0 }}
              >
                {item.description || t('platform_skill_detail_no_description')}
              </Paragraph>
              <Space wrap size={[8, 8]}>
                {renderTrustTag(item.trustLevel, translateTrustLevel(item.trustLevel), emptyPlaceholder)}
                <Tag>{translateSourceType(item.sourceType)}</Tag>
                <Tag>{translateProviderType(item.providerType)}</Tag>
              </Space>
              <Text type="secondary">
                {t('platform_skill_market_card_updated_at', {
                  value: formatDateTime(item.updatedAt, i18n?.resolvedLanguage || i18n?.language),
                })}
              </Text>
              <Space wrap>
                <Button
                  type="primary"
                  ghost
                  disabled={item.status !== 'published'}
                  onClick={() => void openInstallToAgent(item)}
                >
                  {t('platform_skill_market_card_action_install_agent')}
                </Button>
                <Button onClick={() => void fetchDetail(item.id)}>
                  {t('platform_skills_view_detail')}
                </Button>
                <Button
                  type="link"
                  style={{ paddingInline: 0 }}
                  onClick={() => {
                    setActiveTab('imports');
                    openImportJob({ sourceNamespace: item.code });
                  }}
                >
                  {t('platform_skill_market_card_action_import_similar')}
                </Button>
              </Space>
            </Space>
          </Card>
        ))}
      </div>
    );
  };

  const renderMarketHubCards = () => {
    if (!filteredHubs.length) {
      return renderMarketEmpty(
        t('platform_skill_market_empty_sources_title'),
        t('platform_skill_market_empty_sources_desc'),
        t('platform_skill_market_action_manage_sources'),
        () => openCreateHub(),
      );
    }

    return (
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: 16 }}>
        {filteredHubs.map((item) => (
          <Card key={item.id} hoverable style={{ borderRadius: 20 }}>
            <Space orientation="vertical" size={14} style={{ width: '100%' }}>
              <Space align="start" style={{ width: '100%', justifyContent: 'space-between' }}>
                <div>
                  <Text type="secondary">{item.hubCode}</Text>
                  <Title level={5} style={{ margin: '6px 0 0' }}>{item.name}</Title>
                </div>
                {renderStatusTag(item.status, translateStatus(item.status), emptyPlaceholder)}
              </Space>
              <Space wrap size={[8, 8]}>
                <Tag>{translateHubType(item.hubType)}</Tag>
                {renderTrustTag(item.trustLevel, translateTrustLevel(item.trustLevel), emptyPlaceholder)}
              </Space>
              <Paragraph type="secondary" style={{ marginBottom: 0 }}>
                {item.baseUrl || emptyPlaceholder}
              </Paragraph>
              <Space orientation="vertical" size={2}>
                <Text type="secondary">{t('platform_skill_market_card_sync_mode', { value: translateSyncMode(item.syncMode) })}</Text>
                <Text type="secondary">{t('platform_skill_market_card_auth_scheme', { value: translateAuthScheme(item.authScheme) })}</Text>
              </Space>
              <Space wrap>
                <Button onClick={() => openCreateHub(item)}>{t('platform_skills_edit')}</Button>
                <Button
                  type="primary"
                  ghost
                  onClick={() => {
                    setActiveTab('imports');
                    openImportJob({ hubId: item.id });
                  }}
                >
                  {t('platform_skill_market_card_action_import_from_source')}
                </Button>
              </Space>
            </Space>
          </Card>
        ))}
      </div>
    );
  };

  const renderMarketImportCards = () => {
    if (!filteredImportJobs.length) {
      return renderMarketEmpty(
        t('platform_skill_market_empty_imports_title'),
        t('platform_skill_market_empty_imports_desc'),
        t('platform_skill_market_action_import'),
        openImportJob,
      );
    }

    return (
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: 16 }}>
        {filteredImportJobs.map((item) => (
          <Card key={item.id} hoverable style={{ borderRadius: 20 }}>
            <Space orientation="vertical" size={14} style={{ width: '100%' }}>
              <Space align="start" style={{ width: '100%', justifyContent: 'space-between' }}>
                <div>
                  <Text type="secondary">{hubNameMap[item.hubId] || emptyPlaceholder}</Text>
                  <Title level={5} style={{ margin: '6px 0 0' }}>{item.sourceLocator || emptyPlaceholder}</Title>
                </div>
                {renderStatusTag(item.jobStatus, translateStatus(item.jobStatus), emptyPlaceholder)}
              </Space>
              <Space orientation="vertical" size={2}>
                <Text type="secondary">{t('platform_skill_import_jobs_column_source_namespace')}: {item.sourceNamespace || emptyPlaceholder}</Text>
                <Text type="secondary">{t('platform_skill_import_jobs_column_source_version')}: {item.sourceVersion || emptyPlaceholder}</Text>
                <Text type="secondary">{t('platform_skill_import_jobs_column_target_skill')}: {item.targetSkillId || emptyPlaceholder}</Text>
              </Space>
              {item.errorMessage ? (
                <Paragraph type="danger" style={{ marginBottom: 0 }}>
                  {item.errorMessage}
                </Paragraph>
              ) : (
                <Paragraph type="secondary" style={{ marginBottom: 0 }}>
                  {t('platform_skill_market_card_finished_at', {
                    value: formatDateTime(item.finishedAt || item.startedAt || item.createdAt, i18n?.resolvedLanguage || i18n?.language),
                  })}
                </Paragraph>
              )}
              <Space wrap>
                <Button
                  onClick={() => openImportJob({
                    hubId: item.hubId,
                    sourceLocator: item.sourceLocator,
                    sourceNamespace: item.sourceNamespace,
                    sourceVersion: item.sourceVersion,
                  })}
                >
                  {t('platform_skill_market_card_action_retry_import')}
                </Button>
                {item.targetSkillId ? (
                  <Button type="link" style={{ paddingInline: 0 }} onClick={() => void fetchDetail(item.targetSkillId!)}>
                    {t('platform_skill_market_card_action_view_skill')}
                  </Button>
                ) : null}
              </Space>
            </Space>
          </Card>
        ))}
      </div>
    );
  };

  const renderMarketRecommendationCard = () => {
    const hasHubs = hubSummary.total > 0;
    const hasImports = importSummary.total > 0;
    const hasPublishedSkills = skillSummary.published > 0;

    let title = t('platform_skill_market_recommended_connect_title');
    let description = t('platform_skill_market_recommended_connect_desc');
    let primaryActionLabel = t('platform_skill_market_action_manage_sources');
    let primaryAction = () => {
      setActiveTab('hubs');
      openCreateHub();
    };
    let secondaryActionLabel = t('platform_skill_market_action_create');
    let secondaryAction = openCreateSkill;

    if (hasHubs && !hasImports) {
      title = t('platform_skill_market_recommended_import_title');
      description = t('platform_skill_market_recommended_import_desc');
      primaryActionLabel = t('platform_skill_market_action_import');
      primaryAction = () => {
        setActiveTab('imports');
        openImportJob();
      };
      secondaryActionLabel = t('platform_skill_market_action_manage_sources');
      secondaryAction = () => setActiveTab('hubs');
    } else if (hasImports && !hasPublishedSkills) {
      title = t('platform_skill_market_recommended_govern_title');
      description = t('platform_skill_market_recommended_govern_desc');
      primaryActionLabel = t('platform_skill_builder_next_steps_governance');
      primaryAction = () => setActiveTab('skills');
      secondaryActionLabel = t('platform_skill_market_tab_imports');
      secondaryAction = () => setActiveTab('imports');
    } else if (hasPublishedSkills) {
      title = t('platform_skill_market_recommended_install_title');
      description = t('platform_skill_market_recommended_install_desc');
      primaryActionLabel = t('platform_skill_market_tab_catalog');
      primaryAction = () => setActiveTab('skills');
      secondaryActionLabel = t('platform_skill_market_action_import');
      secondaryAction = () => {
        setActiveTab('imports');
        openImportJob();
      };
    }

    return (
      <Card
        style={{
          marginBottom: 16,
          borderRadius: 24,
          background: 'linear-gradient(135deg, rgba(22,119,255,0.08), rgba(22,119,255,0.02))',
          border: '1px solid rgba(22,119,255,0.12)',
        }}
      >
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            gap: 16,
            flexWrap: 'wrap',
          }}
        >
          <Space orientation="vertical" size={8} style={{ maxWidth: 760 }}>
            <Text strong>{t('platform_skill_market_recommended_title')}</Text>
            <Title level={4} style={{ margin: 0 }}>{title}</Title>
            <Paragraph type="secondary" style={{ marginBottom: 0 }}>
              {description}
            </Paragraph>
            <Text type="secondary">{t('platform_skill_market_search_hint')}</Text>
          </Space>
          <Space wrap>
            <Button onClick={secondaryAction}>{secondaryActionLabel}</Button>
            <Button type="primary" icon={<ArrowRightOutlined />} onClick={primaryAction}>
              {primaryActionLabel}
            </Button>
          </Space>
        </div>
      </Card>
    );
  };

  return (
    <div style={{ animation: 'fadeIn 0.5s ease-out' }}>
      {contextHolder}
      <div style={{ marginBottom: 24, display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 16, flexWrap: 'wrap' }}>
        <div>
          <Title level={3} style={{ marginBottom: 8 }}>
            <SafetyOutlined style={{ marginRight: 8 }} />
            {pageTitle}
          </Title>
          <Paragraph type="secondary" style={{ marginBottom: 0 }}>
            {pageDescription}
          </Paragraph>
        </div>
        <Space wrap>
          {isMarketExperience ? (
            <>
              <Button icon={<CloudDownloadOutlined />} onClick={() => openImportJob()}>
                {t('platform_skill_market_action_import')}
              </Button>
              <Button onClick={() => openCreateHub()}>
                {t('platform_skill_market_action_manage_sources')}
              </Button>
              <Button type="primary" icon={<PlusOutlined />} onClick={openCreateSkill}>
                {t('platform_skill_market_action_create')}
              </Button>
            </>
          ) : actionButton}
        </Space>
      </div>

      {isMarketExperience ? (
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))',
            gap: 16,
            marginBottom: 16,
          }}
        >
          <Card size="small" style={{ borderRadius: 16 }}>
            <Text strong>{t('platform_skill_market_quick_import_title')}</Text>
            <Paragraph type="secondary" style={{ margin: '8px 0 12px' }}>
              {t('platform_skill_market_quick_import_desc')}
            </Paragraph>
            <Button type="link" style={{ paddingInline: 0 }} onClick={() => openImportJob()}>
              {t('platform_skill_market_action_import')}
            </Button>
          </Card>
          <Card size="small" style={{ borderRadius: 16 }}>
            <Text strong>{t('platform_skill_market_quick_create_title')}</Text>
            <Paragraph type="secondary" style={{ margin: '8px 0 12px' }}>
              {t('platform_skill_market_quick_create_desc')}
            </Paragraph>
            <Button type="link" style={{ paddingInline: 0 }} onClick={openCreateSkill}>
              {t('platform_skill_market_action_create')}
            </Button>
          </Card>
          <Card size="small" style={{ borderRadius: 16 }}>
            <Text strong>{t('platform_skill_market_quick_hub_title')}</Text>
            <Paragraph type="secondary" style={{ margin: '8px 0 12px' }}>
              {t('platform_skill_market_quick_hub_desc')}
            </Paragraph>
            <Button type="link" style={{ paddingInline: 0 }} onClick={() => {
              setActiveTab('hubs');
              openCreateHub();
            }}>
              {t('platform_skill_market_action_manage_sources')}
            </Button>
          </Card>
        </div>
      ) : null}

      {isMarketExperience ? renderMarketRecommendationCard() : null}

      <Card size="small" style={{ marginBottom: 16, borderRadius: 16 }}>
        <Space orientation="vertical" size={12} style={{ width: '100%' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', gap: 12, flexWrap: 'wrap' }}>
            <Space wrap size={12}>
            {tabItems.map((item) => (
              <Button
                key={item.key}
                type={activeTab === item.key ? 'primary' : 'default'}
                onClick={() => setActiveTab(item.key as PlatformTabKey)}
              >
                {item.label}
              </Button>
            ))}
            </Space>
            <Input
              allowClear
              value={searchText}
              onChange={(event) => setSearchText(event.target.value)}
              prefix={<SearchOutlined />}
              placeholder={searchPlaceholder}
              style={{ width: 320, maxWidth: '100%' }}
            />
          </div>
          {isMarketExperience ? (
            <Text type="secondary">{t('platform_skill_market_search_hint')}</Text>
          ) : null}
        </Space>
      </Card>

      {activeTab === 'skills' ? (
        <>
          {renderSummaryCards([
            { label: t('platform_skills_summary_total'), value: skillSummary.total, active: skillSummaryFilter === 'all', onClick: () => setSkillSummaryFilter('all') },
            { label: t('platform_skills_summary_published'), value: skillSummary.published, active: skillSummaryFilter === 'published', onClick: () => setSkillSummaryFilter('published') },
            { label: t('platform_skills_summary_draft'), value: skillSummary.draft, active: skillSummaryFilter === 'draft', onClick: () => setSkillSummaryFilter('draft') },
            { label: t('platform_skills_summary_disabled'), value: skillSummary.disabled, active: skillSummaryFilter === 'disabled', onClick: () => setSkillSummaryFilter('disabled') },
          ])}

          {isMarketExperience ? (
            <>
              <Card size="small" style={{ marginBottom: 16, borderRadius: 16 }}>
                <Space wrap>
                  <Select value={skillSourceFilter} onChange={setSkillSourceFilter} options={skillSourceOptions} style={{ minWidth: 180 }} />
                  <Select value={skillStatusFilter} onChange={setSkillStatusFilter} options={skillStatusOptions} style={{ minWidth: 180 }} />
                </Space>
              </Card>
              {skillsLoading ? <Card loading style={{ borderRadius: 20 }} /> : renderMarketSkillCards()}
            </>
          ) : (
            <Card variant="borderless" style={{ borderRadius: 20 }}>
              <Table
                rowKey="id"
                loading={skillsLoading}
                dataSource={filteredSkills}
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
                    render: (value: string) => renderTrustTag(value, translateTrustLevel(value), emptyPlaceholder),
                  },
                  {
                    title: t('platform_skills_column_status'),
                    dataIndex: 'status',
                    key: 'status',
                    width: 120,
                    render: (value: string) => renderStatusTag(value, translateStatus(value), emptyPlaceholder),
                  },
                  {
                    title: t('platform_skills_column_latest_stable'),
                    dataIndex: 'latestStableVersionId',
                    key: 'latestStableVersionId',
                    width: 220,
                    ellipsis: true,
                    render: (value?: string) => value || emptyPlaceholder,
                  },
                  {
                    title: t('platform_skills_column_updated_at'),
                    dataIndex: 'updatedAt',
                    key: 'updatedAt',
                    width: 200,
                    render: (value: string) => formatDateTime(value, i18n?.resolvedLanguage || i18n?.language),
                  },
                  {
                    title: t('platform_skills_column_actions'),
                    key: 'actions',
                    width: 240,
                    render: (_: unknown, record: SkillItem) => (
                      <Space>
                        <Button
                          size="small"
                          onClick={(event) => {
                            event.stopPropagation();
                            void fetchDetail(record.id);
                          }}
                        >
                          {t('platform_skills_view_detail')}
                        </Button>
                        <Button
                          size="small"
                          onClick={(event) => {
                            event.stopPropagation();
                            openReleaseSettings({
                              resourceId: record.id,
                              resourceType: 'skill',
                              resourceName: record.name || record.code,
                            });
                          }}
                        >
                          {t('platform_resource_release_action')}
                        </Button>
                      </Space>
                    ),
                  },
                ]}
              />
            </Card>
          )}
        </>
      ) : null}

      {activeTab === 'hubs' ? (
        <>
          {renderSummaryCards([
            { label: t('platform_skill_hubs_summary_total'), value: hubSummary.total, active: hubSummaryFilter === 'all', onClick: () => setHubSummaryFilter('all') },
            { label: t('platform_skill_hubs_summary_enabled'), value: hubSummary.enabled, active: hubSummaryFilter === 'enabled', onClick: () => setHubSummaryFilter('enabled') },
            { label: t('platform_skill_hubs_summary_openapi'), value: hubSummary.openapi, active: hubSummaryFilter === 'openapi', onClick: () => setHubSummaryFilter('openapi') },
            { label: t('platform_skill_hubs_summary_mcp'), value: hubSummary.mcp, active: hubSummaryFilter === 'mcp', onClick: () => setHubSummaryFilter('mcp') },
          ])}

          {isMarketExperience ? (
            <>
              <Card size="small" style={{ marginBottom: 16, borderRadius: 16 }}>
                <Space wrap>
                  <Select value={hubTypeFilter} onChange={setHubTypeFilter} options={hubTypeOptions} style={{ minWidth: 220 }} />
                </Space>
              </Card>
              {hubsLoading ? <Card loading style={{ borderRadius: 20 }} /> : renderMarketHubCards()}
            </>
          ) : (
            <Card variant="borderless" style={{ borderRadius: 20 }}>
              <Table
                rowKey="id"
                loading={hubsLoading}
                dataSource={filteredHubs}
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
                    render: (value: string) => value || emptyPlaceholder,
                  },
                  {
                    title: t('platform_skills_column_trust_level'),
                    dataIndex: 'trustLevel',
                    key: 'trustLevel',
                    width: 150,
                    render: (value: string) => renderTrustTag(value, translateTrustLevel(value), emptyPlaceholder),
                  },
                  {
                    title: t('platform_skills_column_status'),
                    dataIndex: 'status',
                    key: 'status',
                    width: 120,
                    render: (value: string) => renderStatusTag(value, translateStatus(value), emptyPlaceholder),
                  },
                  {
                    title: t('platform_skills_column_updated_at'),
                    dataIndex: 'updatedAt',
                    key: 'updatedAt',
                    width: 180,
                    render: (value: string) => formatDateTime(value, i18n?.resolvedLanguage || i18n?.language),
                  },
                  {
                    title: t('platform_skills_column_actions'),
                    key: 'actions',
                    width: 240,
                    render: (_: unknown, record: SkillHubItem) => (
                      <Space>
                        <Button size="small" onClick={() => openCreateHub(record)}>
                          {t('platform_skills_edit')}
                        </Button>
                        <Button
                          size="small"
                          onClick={() => openReleaseSettings({
                            resourceId: record.id,
                            resourceType: 'hub',
                            resourceName: record.name || record.hubCode,
                          })}
                        >
                          {t('platform_resource_release_action')}
                        </Button>
                      </Space>
                    ),
                  },
                ]}
              />
            </Card>
          )}
        </>
      ) : null}

      {activeTab === 'imports' ? (
        <>
          {renderSummaryCards([
            { label: t('platform_skill_import_jobs_summary_total'), value: importSummary.total, active: importSummaryFilter === 'all', onClick: () => setImportSummaryFilter('all') },
            { label: t('platform_skill_import_jobs_summary_completed'), value: importSummary.completed, active: importSummaryFilter === 'completed', onClick: () => setImportSummaryFilter('completed') },
            { label: t('platform_skill_import_jobs_summary_running'), value: importSummary.running, active: importSummaryFilter === 'running', onClick: () => setImportSummaryFilter('running') },
            { label: t('platform_skill_import_jobs_summary_failed'), value: importSummary.failed, active: importSummaryFilter === 'failed', onClick: () => setImportSummaryFilter('failed') },
          ])}

          {isMarketExperience ? (
            <>
              {renderImportGuideCard()}
              <Card size="small" style={{ marginBottom: 16, borderRadius: 16 }}>
                <Space wrap>
                  <Select value={importStatusFilter} onChange={setImportStatusFilter} options={importStatusOptions} style={{ minWidth: 220 }} />
                </Space>
              </Card>
              {importsLoading ? <Card loading style={{ borderRadius: 20 }} /> : renderMarketImportCards()}
            </>
          ) : (
            <Card variant="borderless" style={{ borderRadius: 20 }}>
              <Table
                rowKey="id"
                loading={importsLoading}
                dataSource={filteredImportJobs}
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
                    render: (value: string) => hubNameMap[value] || value || emptyPlaceholder,
                  },
                  {
                    title: t('platform_skill_import_jobs_column_source_locator'),
                    dataIndex: 'sourceLocator',
                    key: 'sourceLocator',
                    width: 280,
                    ellipsis: true,
                    render: (value: string) => value || emptyPlaceholder,
                  },
                  {
                    title: t('platform_skill_import_jobs_column_source_namespace'),
                    dataIndex: 'sourceNamespace',
                    key: 'sourceNamespace',
                    width: 180,
                    render: (value: string) => value || emptyPlaceholder,
                  },
                  { title: t('platform_skill_import_jobs_column_source_version'), dataIndex: 'sourceVersion', key: 'sourceVersion', width: 120, render: (value: string) => value || emptyPlaceholder },
                  {
                    title: t('platform_skills_column_status'),
                    dataIndex: 'jobStatus',
                    key: 'jobStatus',
                    width: 120,
                    render: (value: string) => renderStatusTag(value, translateStatus(value), emptyPlaceholder),
                  },
                  {
                    title: t('platform_skill_import_jobs_column_target_skill'),
                    dataIndex: 'targetSkillId',
                    key: 'targetSkillId',
                    width: 160,
                    render: (value?: string) => value || emptyPlaceholder,
                  },
                  {
                    title: t('platform_skill_import_jobs_column_target_version'),
                    dataIndex: 'targetSkillVersionId',
                    key: 'targetSkillVersionId',
                    width: 160,
                    render: (value?: string) => value || emptyPlaceholder,
                  },
                  {
                    title: t('platform_skill_import_jobs_column_finished_at'),
                    dataIndex: 'finishedAt',
                    key: 'finishedAt',
                    width: 180,
                    render: (value?: string) => formatDateTime(value, i18n?.resolvedLanguage || i18n?.language),
                  },
                  {
                    title: t('platform_skill_import_jobs_column_error_message'),
                    dataIndex: 'errorMessage',
                    key: 'errorMessage',
                    width: 280,
                    ellipsis: true,
                    render: (value?: string) => value || emptyPlaceholder,
                  },
                ]}
              />
            </Card>
          )}
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
            <Input placeholder={platformSkillFormExamples.skillCode} />
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
            <Input placeholder={platformSkillFormExamples.hubCode} />
          </Form.Item>
          <Form.Item label={t('platform_skills_form_name')} name="name" rules={[{ required: true, message: t('platform_skills_form_name_required') }]}>
            <Input placeholder={platformSkillFormExamples.hubName} />
          </Form.Item>
          <Form.Item label={t('platform_skill_hubs_form_type')} name="hubType" rules={[{ required: true, message: t('platform_skill_hubs_form_type_required') }]}>
            <Select
              options={[
                { label: translateHubType('openapi_hub'), value: 'openapi_hub' },
                { label: translateHubType('mcp_hub'), value: 'mcp_hub' },
                { label: translateHubType('tencent_skillhub'), value: 'tencent_skillhub' },
                { label: translateHubType('builtin_hub'), value: 'builtin_hub' },
                { label: translateHubType('enterprise_private_hub'), value: 'enterprise_private_hub' },
              ]}
            />
          </Form.Item>
          <Form.Item label={t('platform_skill_hubs_form_base_url')} name="baseUrl">
            <Input placeholder={platformSkillFormExamples.hubBaseUrl} />
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
          {renderImportGuideCard()}
          <Form.Item label={t('platform_skill_import_jobs_form_hub')} name="hubId" rules={[{ required: true, message: t('platform_skill_import_jobs_form_hub_required') }]}>
            <Select
              showSearch
              optionFilterProp="label"
              placeholder={t('platform_skill_import_jobs_form_hub_placeholder')}
              options={hubs.map((item) => ({
                label: `${item.name} (${item.hubCode})`,
                value: item.id,
              }))}
            />
          </Form.Item>
          <Form.Item label={t('platform_skill_import_jobs_form_source_locator')} name="sourceLocator" rules={[{ required: true, message: t('platform_skill_import_jobs_form_source_locator_required') }]}>
            <Input placeholder={t('platform_skill_import_jobs_form_source_locator_placeholder')} />
          </Form.Item>
          <Form.Item label={t('platform_skill_import_jobs_form_source_namespace')} name="sourceNamespace">
            <Input placeholder={t('platform_skill_import_jobs_form_source_namespace_placeholder')} />
          </Form.Item>
          <Form.Item label={t('platform_skill_import_jobs_form_source_version')} name="sourceVersion">
            <Input placeholder={t('platform_skill_import_jobs_form_source_version_placeholder')} />
          </Form.Item>
          <Card size="small" style={{ borderRadius: 16 }}>
            <Space orientation="vertical" size={8} style={{ width: '100%' }}>
              <Text strong>{t('platform_skill_import_examples_title')}</Text>
              <Text type="secondary">{t('platform_skill_import_examples_desc')}</Text>
              <Space orientation="vertical" size={6} style={{ width: '100%' }}>
                {quickImportTemplates.map((template) => (
                  <Card key={template.key} size="small" style={{ borderRadius: 12 }}>
                    <Space orientation="vertical" size={4} style={{ width: '100%' }}>
                      <Space wrap style={{ justifyContent: 'space-between', width: '100%' }}>
                        <Text strong>{template.label}</Text>
                        <Button size="small" onClick={() => applyImportTemplate(template)}>
                          {t('platform_skill_import_examples_use')}
                        </Button>
                      </Space>
                      <Text type="secondary">{template.description}</Text>
                      <Text type="secondary">
                        {t('platform_skill_import_examples_locator', { value: template.locator })}
                      </Text>
                      {template.namespace ? (
                        <Text type="secondary">
                          {t('platform_skill_import_examples_namespace', { value: template.namespace })}
                        </Text>
                      ) : null}
                    </Space>
                  </Card>
                ))}
              </Space>
            </Space>
          </Card>
        </Form>
      </Modal>

      <Modal
        title={t('platform_skill_install_agent_modal_title')}
        open={installToAgentOpen}
        onCancel={() => {
          setInstallToAgentOpen(false);
          setInstallSkill(null);
        }}
        onOk={handleInstallToAgent}
        confirmLoading={saving}
        okText={t('platform_skill_market_card_action_install_agent')}
        cancelText={t('agent_cancel')}
        destroyOnHidden
      >
        <Space orientation="vertical" size={16} style={{ width: '100%' }}>
          <Card
            size="small"
            style={{
              borderRadius: 16,
              background: 'linear-gradient(135deg, rgba(22,119,255,0.08), rgba(22,119,255,0.02))',
              border: '1px solid rgba(22,119,255,0.12)',
            }}
          >
            <Space orientation="vertical" size={6} style={{ width: '100%' }}>
              <Text strong>{t('platform_skill_install_agent_selected_skill')}</Text>
              <Text>{installSkill ? `${installSkill.name} (${installSkill.code})` : emptyPlaceholder}</Text>
              <Text type="secondary">{t('platform_skill_install_agent_modal_desc')}</Text>
            </Space>
          </Card>

          {installTargets.length === 0 && !installTargetsLoading ? (
            <Card size="small" style={{ borderRadius: 16 }}>
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description={t('platform_skill_install_agent_modal_empty')}
              >
                <Button onClick={() => navigate(getLocalizedPath('/dashboard', currentLanguage))}>
                  {t('agent_create')}
                </Button>
              </Empty>
            </Card>
          ) : (
            <Form form={installToAgentForm} layout="vertical" initialValues={{ invokeVisibility: 'auto' }}>
              <Form.Item
                label={t('platform_skill_install_agent_form_agent')}
                name="agentId"
                rules={[{ required: true, message: t('platform_skill_install_agent_form_agent_required') }]}
              >
                <Select
                  loading={installTargetsLoading}
                  showSearch
                  optionFilterProp="label"
                  placeholder={t('platform_skill_install_agent_form_agent_placeholder')}
                  options={installTargets.map((item) => ({
                    label: `${item.agentName} · ${item.engineType}${item.modelName ? ` · ${item.modelName}` : ''}`,
                    value: item.id,
                  }))}
                />
              </Form.Item>
              <Form.Item label={t('platform_skill_install_agent_form_alias')} name="entryAlias">
                <Input placeholder={t('platform_skill_install_agent_form_alias_placeholder')} />
              </Form.Item>
              <Form.Item label={t('platform_skill_install_agent_form_visibility')} name="invokeVisibility">
                <Select
                  options={[
                    { label: t('agent_skill_panel_visibility_auto'), value: 'auto' },
                    { label: t('agent_skill_panel_visibility_suggested'), value: 'suggested' },
                    { label: t('agent_skill_panel_visibility_manual'), value: 'manual' },
                  ]}
                />
              </Form.Item>
              <Text type="secondary">{t('platform_skill_install_agent_helper')}</Text>
            </Form>
          )}
        </Space>
      </Modal>

      <Modal
        title={t('platform_skill_install_success_modal_title')}
        open={!!installSuccess}
        onCancel={() => setInstallSuccess(null)}
        footer={installSuccess ? [
          <Button key="close" onClick={() => setInstallSuccess(null)}>
            {t('agent_cancel')}
          </Button>,
          <Button
            key="agent-skills"
            onClick={() => {
              navigate(getLocalizedPath(`/dashboard/agents/${installSuccess.agentId}/skills`, currentLanguage));
              setInstallSuccess(null);
            }}
          >
            {t('platform_skill_install_success_action_agent')}
          </Button>,
          <Button
            key="chat-verify"
            type="primary"
            onClick={() => {
              navigate(getLocalizedPath('/chat', currentLanguage), {
                state: {
                  verifyAgentId: installSuccess.agentId,
                  verifyAgentName: installSuccess.agentName,
                  verifySkillName: installSuccess.skillName,
                  source: 'skill-market-install',
                },
              });
              setInstallSuccess(null);
            }}
          >
            {t('platform_skill_install_success_action_chat')}
          </Button>,
        ] : undefined}
        destroyOnHidden
      >
        {installSuccess ? (
          <Space orientation="vertical" size={12} style={{ width: '100%' }}>
            <Text strong>{t('platform_skill_install_success_selected_skill')}</Text>
            <Text>{installSuccess.skillName}</Text>
            <Paragraph type="secondary" style={{ marginBottom: 0 }}>
              {t('platform_skill_install_success_modal_desc', {
                agentName: installSuccess.agentName,
                skillName: installSuccess.skillName,
              })}
            </Paragraph>
          </Space>
        ) : null}
      </Modal>

      <Modal
        title={t('platform_skill_import_rollout_modal_title')}
        open={!!importRollout}
        onCancel={() => setImportRollout(null)}
        footer={importRollout ? [
          <Button key="close" onClick={() => setImportRollout(null)}>
            {t('agent_cancel')}
          </Button>,
          <Button
            key="detail"
            onClick={() => {
              void fetchDetail(importRollout.skillId);
              setImportRollout(null);
            }}
          >
            {t('platform_skill_import_rollout_action_review')}
          </Button>,
          <Button
            key="enterprise"
            type="primary"
            onClick={() => {
              navigate(getLocalizedPath(`/admin/enterprise?tab=skills&skillId=${encodeURIComponent(importRollout.skillId)}`, currentLanguage));
              setImportRollout(null);
            }}
          >
            {t('platform_skill_import_rollout_action_enterprise')}
          </Button>,
        ] : undefined}
        destroyOnHidden
      >
        {importRollout ? (
          <Space orientation="vertical" size={12} style={{ width: '100%' }}>
            <Paragraph type="secondary" style={{ marginBottom: 0 }}>
              {t('platform_skill_import_rollout_modal_desc')}
            </Paragraph>
            <Space wrap size={[8, 8]}>
              <Tag color="processing">{t('platform_skill_import_rollout_step_publish')}</Tag>
              <Tag color="gold">{t('platform_skill_import_rollout_step_enable')}</Tag>
              <Tag>{t('platform_skill_import_rollout_step_install')}</Tag>
            </Space>
            <Text strong>{t('platform_skill_import_rollout_locator_label')}</Text>
            <Text>{importRollout.sourceLocator}</Text>
            {importRollout.sourceNamespace ? (
              <>
                <Text strong>{t('platform_skill_import_rollout_namespace_label')}</Text>
                <Text>{importRollout.sourceNamespace}</Text>
              </>
            ) : null}
            {importRollout.sourceVersion ? (
              <>
                <Text strong>{t('platform_skill_import_rollout_version_label')}</Text>
                <Text>{importRollout.sourceVersion}</Text>
              </>
            ) : null}
          </Space>
        ) : null}
      </Modal>

      <Modal
        title={t(
          releaseTarget?.resourceType === 'hub'
            ? 'platform_resource_release_hub_title'
            : 'platform_resource_release_skill_title',
        )}
        open={releaseOpen}
        onCancel={() => {
          setReleaseOpen(false);
          setReleaseTarget(null);
          setReleaseRecords([]);
        }}
        onOk={handleSaveRelease}
        confirmLoading={saving}
        okText={t('agent_save')}
        cancelText={t('agent_cancel')}
        destroyOnHidden
      >
        <Space orientation="vertical" size={12} style={{ width: '100%' }}>
          <Card size="small" loading={releaseRecordsLoading} style={{ borderRadius: 16, background: '#fafafa' }}>
            <Space orientation="vertical" size={10} style={{ width: '100%' }}>
              <Text strong>{t('platform_resource_release_current_title')}</Text>
              {releaseRecords.length ? releaseRecords.map((item) => (
                <Card key={item.id} size="small" style={{ borderRadius: 12 }}>
                  <Space orientation="vertical" size={4} style={{ width: '100%' }}>
                    <Space wrap size={[8, 8]}>
                      <Tag color={item.releaseScope === 'global' ? 'blue' : 'purple'}>
                        {item.releaseScope === 'global'
                          ? t('platform_resource_release_scope_global')
                          : t('platform_resource_release_scope_enterprise')}
                      </Tag>
                      <Tag color={item.releaseStatus === 'enabled' ? 'success' : 'default'}>
                        {translateStatus(item.releaseStatus)}
                      </Tag>
                      {item.targetEnterpriseId ? <Tag>{item.targetEnterpriseId}</Tag> : null}
                    </Space>
                    <Text type="secondary">
                      {t('platform_resource_release_meta_line', {
                        operator: item.operatedBy || '-',
                        updatedAt: formatDateTime(item.updatedAt, currentLanguage),
                      })}
                    </Text>
                    {item.note ? <Paragraph style={{ marginBottom: 0 }}>{item.note}</Paragraph> : null}
                  </Space>
                </Card>
              )) : (
                <Text type="secondary">{t('platform_resource_release_current_empty')}</Text>
              )}
            </Space>
          </Card>
        <Form form={releaseForm} layout="vertical">
          <Paragraph type="secondary" style={{ marginBottom: 12 }}>
            {t('platform_resource_release_desc', { name: releaseTarget?.resourceName || '-' })}
          </Paragraph>
          <Form.Item label={t('platform_resource_release_scope')} name="releaseScope" rules={[{ required: true }]}>
            <Radio.Group
              optionType="button"
              buttonStyle="solid"
              options={[
                { label: t('platform_resource_release_scope_global'), value: 'global' },
                { label: t('platform_resource_release_scope_enterprise'), value: 'enterprise' },
              ]}
            />
          </Form.Item>
          {selectedReleaseScope === 'enterprise' ? (
            <Form.Item
              label={t('platform_resource_release_target_enterprise')}
              name="targetEnterpriseId"
              rules={[{ required: true, message: t('platform_resource_release_target_enterprise_required') }]}
            >
              <Input placeholder={t('platform_resource_release_target_enterprise_placeholder')} />
            </Form.Item>
          ) : null}
          <Form.Item label={t('platform_resource_release_status')} name="releaseStatus" rules={[{ required: true }]}>
            <Select
              options={[
                { label: translateStatus('enabled'), value: 'enabled' },
                { label: translateStatus('disabled'), value: 'disabled' },
              ]}
            />
          </Form.Item>
          <Form.Item label={t('platform_resource_release_note')} name="note">
            <TextArea rows={3} placeholder={t('platform_resource_release_note_placeholder')} />
          </Form.Item>
        </Form>
        </Space>
      </Modal>

      <Drawer
        title={selectedSkill?.skill?.name || t('platform_skill_detail_title')}
        open={detailOpen}
        onClose={() => setDetailOpen(false)}
        width={960}
        extra={
          <Space>
            {selectedSkill?.skill ? (
              <Button
                onClick={() => openReleaseSettings({
                  resourceId: selectedSkill.skill.id,
                  resourceType: 'skill',
                  resourceName: selectedSkill.skill.name || selectedSkill.skill.code,
                })}
              >
                {t('platform_resource_release_action')}
              </Button>
            ) : null}
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
          </Space>
        }
      >
        {selectedSkill?.skill ? (
          <Space orientation="vertical" size={16} style={{ width: '100%' }}>
            <Card size="small">
              <Space size={12} wrap>
                <Text strong>{selectedSkill.skill.code}</Text>
                {renderTrustTag(selectedSkill.skill.trustLevel, translateTrustLevel(selectedSkill.skill.trustLevel), emptyPlaceholder)}
                {renderStatusTag(selectedSkill.skill.status, translateStatus(selectedSkill.skill.status), emptyPlaceholder)}
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
                    render: (value: string) => renderStatusTag(value, translateStatus(value), emptyPlaceholder),
                  },
                  {
                    title: t('platform_skill_detail_version_column_change_log'),
                    dataIndex: 'changeLog',
                    key: 'changeLog',
                    render: (value: string) => value || emptyPlaceholder,
                  },
                  {
                    title: t('platform_skill_detail_version_column_published_at'),
                    dataIndex: 'publishedAt',
                    key: 'publishedAt',
                    width: 180,
                    render: (value?: string) => formatDateTime(value, i18n?.resolvedLanguage || i18n?.language),
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
                    render: (value: string) => value || emptyPlaceholder,
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
                    render: (value: string) => value || emptyPlaceholder,
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
            <Input placeholder={platformSkillFormExamples.version} />
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
            <TextArea rows={4} placeholder={platformSkillFormExamples.referencesJson} />
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
            <TextArea rows={8} placeholder={platformSkillFormExamples.referencesJson} />
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
