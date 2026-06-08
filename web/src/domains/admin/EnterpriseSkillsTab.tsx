import React, { useEffect, useMemo, useRef, useState } from 'react';
import { Button, Card, Empty, Form, Input, Modal, Select, Space, Table, Tag, Typography, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import axios from 'axios';
import { useTranslation } from 'react-i18next';
import { useSearchParams } from 'react-router-dom';
import { BACKEND_URL } from '../../config';
import { casdoorService } from '../identity/CasdoorService';

const { Paragraph, Text, Title } = Typography;
const { TextArea } = Input;

interface SkillItem {
  id: string;
  code: string;
  name: string;
  description: string;
  sourceType: string;
  providerType?: string;
  trustLevel: string;
  status: string;
  latestPublishedVersion: string;
  enablementStatus: string;
}

interface SkillVersionItem {
  id: string;
  version: string;
  releaseChannel: string;
  releaseStatus: string;
  changeLog: string;
  createdAt?: string;
  publishedAt?: string;
}

interface SkillReferenceItem {
  id: string;
  toSkillVersionId: string;
}

interface GovernedSkillDetail {
  skill: SkillItem;
  versions: SkillVersionItem[];
  references: SkillReferenceItem[];
}

interface EnterpriseSkillsTabProps {
  createSignal: number;
}

interface EnableSkillFormValues {
  skillId: string;
  channelScope?: string[];
}

interface CreateSkillFormValues {
  code: string;
  name: string;
  description?: string;
  sourceType: string;
  providerType: string;
}

interface CreateVersionFormValues {
  version: string;
  changeLog?: string;
}

type EnterpriseSkillView = 'governance' | 'catalog';
const CURRENT_ENTERPRISE_STORAGE_KEY = 'dotblue_current_enterprise_id';

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

function renderStatusTag(status?: string, t?: (key: string, defaultValue?: string) => string) {
  if (!status) {
    return <Tag>{t ? t('enterprise_admin_skills_status_not_enabled') : 'enterprise_admin_skills_status_not_enabled'}</Tag>;
  }
  const normalized = status.toLowerCase();
  const color = normalized === 'enabled' || normalized === 'published'
    ? 'success'
    : normalized === 'reviewing'
      ? 'processing'
      : normalized === 'suspended'
        ? 'warning'
        : 'default';
  const label = t ? t(`platform_skills_status_${normalized}`, status) : status;
  return <Tag color={color}>{label}</Tag>;
}

function renderTrustTag(value?: string, t?: (key: string, defaultValue?: string) => string) {
  if (!value) {
    return <Tag>-</Tag>;
  }
  const normalized = value.toLowerCase();
  const color = normalized.includes('trusted')
    ? 'success'
    : normalized.includes('verified')
      ? 'processing'
      : normalized.includes('blocked')
        ? 'error'
        : 'default';
  return <Tag color={color}>{t ? t(`platform_skills_trust_${normalized}`, value) : value}</Tag>;
}

function getEnterpriseSkillErrorMessage(fallbackMessage: string) {
  return fallbackMessage;
}

const EnterpriseSkillsTab: React.FC<EnterpriseSkillsTabProps> = ({ createSignal }) => {
  const { t } = useTranslation();
  const [searchParams, setSearchParams] = useSearchParams();
  const [messageApi, contextHolder] = message.useMessage();
  const [activeView, setActiveView] = useState<EnterpriseSkillView>('governance');
  const [catalogLoading, setCatalogLoading] = useState(true);
  const [governanceLoading, setGovernanceLoading] = useState(true);
  const [catalogSkills, setCatalogSkills] = useState<SkillItem[]>([]);
  const [governedSkills, setGovernedSkills] = useState<SkillItem[]>([]);
  const [enableModalOpen, setEnableModalOpen] = useState(false);
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [detailModalOpen, setDetailModalOpen] = useState(false);
  const [versionModalOpen, setVersionModalOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);
  const [publishingId, setPublishingId] = useState('');
  const [pendingEnableSkillId, setPendingEnableSkillId] = useState('');
  const [selectedSkill, setSelectedSkill] = useState<GovernedSkillDetail | null>(null);
  const [enableForm] = Form.useForm<EnableSkillFormValues>();
  const [createForm] = Form.useForm<CreateSkillFormValues>();
  const [versionForm] = Form.useForm<CreateVersionFormValues>();
  const consumedSkillIdRef = useRef('');

  const translateWithFallback = (key: string, fallback?: string) => (fallback ? t(key, fallback) : t(key));

  const fetchCatalogSkills = async () => {
    setCatalogLoading(true);
    try {
      const res = await axios.get(`${BACKEND_URL}/api/admin/skills?view=catalog`, {
        headers: getAuthHeaders(),
      });
      setCatalogSkills(Array.isArray(res.data) ? res.data : []);
    } catch {
      messageApi.error(t('enterprise_admin_skills_catalog_load_failed'));
      setCatalogSkills([]);
    } finally {
      setCatalogLoading(false);
    }
  };

  const fetchGovernedSkills = async () => {
    setGovernanceLoading(true);
    try {
      const res = await axios.get(`${BACKEND_URL}/api/admin/skills?view=governance`, {
        headers: getAuthHeaders(),
      });
      setGovernedSkills(Array.isArray(res.data) ? res.data : []);
    } catch {
      messageApi.error(t('enterprise_admin_skills_governance_load_failed'));
      setGovernedSkills([]);
    } finally {
      setGovernanceLoading(false);
    }
  };

  const fetchSkillDetail = async (skillId: string) => {
    setDetailLoading(true);
    try {
      const res = await axios.get(`${BACKEND_URL}/api/admin/skills/${skillId}`, {
        headers: getAuthHeaders(),
      });
      setSelectedSkill(res.data || null);
      setDetailModalOpen(true);
    } catch {
      messageApi.error(t('enterprise_admin_skills_detail_load_failed'));
    } finally {
      setDetailLoading(false);
    }
  };

  useEffect(() => {
    void Promise.all([fetchCatalogSkills(), fetchGovernedSkills()]);
  }, []);

  useEffect(() => {
    if (createSignal > 0) {
      setCreateModalOpen(true);
      setActiveView('governance');
    }
  }, [createSignal]);

  useEffect(() => {
    if (!createModalOpen) {
      return;
    }
    createForm.resetFields();
    createForm.setFieldsValue({
      sourceType: 'builtin',
      providerType: 'native',
    });
  }, [createModalOpen, createForm]);

  useEffect(() => {
    if (!versionModalOpen) {
      return;
    }
    versionForm.resetFields();
  }, [versionModalOpen, versionForm]);

  useEffect(() => {
    if (!enableModalOpen) {
      return;
    }
    enableForm.resetFields();
    if (pendingEnableSkillId) {
      enableForm.setFieldsValue({
        skillId: pendingEnableSkillId,
        channelScope: [],
      });
    }
  }, [enableModalOpen, enableForm, pendingEnableSkillId]);

  useEffect(() => {
    if (!searchParams.get('skillId')) {
      consumedSkillIdRef.current = '';
    }
  }, [searchParams]);

  useEffect(() => {
    const targetSkillId = searchParams.get('skillId')?.trim() || '';
    if (!targetSkillId || catalogLoading || !catalogSkills.length) {
      return;
    }
    if (consumedSkillIdRef.current === targetSkillId) {
      return;
    }
    const matchedSkill = catalogSkills.find((item) => item.id === targetSkillId);
    if (!matchedSkill) {
      consumedSkillIdRef.current = targetSkillId;
      clearPendingSkillId();
      return;
    }
    consumedSkillIdRef.current = targetSkillId;
    setActiveView('catalog');
    if (matchedSkill.enablementStatus !== 'enabled') {
      openEnableModal(matchedSkill.id);
      return;
    }
    messageApi.success(t('enterprise_admin_skills_target_already_enabled'));
    clearPendingSkillId();
  }, [catalogLoading, catalogSkills, enableForm, messageApi, searchParams, setSearchParams, t]);

  const enableableSkills = useMemo(
    () => catalogSkills.filter((item) => item.enablementStatus !== 'enabled'),
    [catalogSkills],
  );

  const clearPendingSkillId = () => {
    if (!searchParams.get('skillId')) {
      return;
    }
    const nextSearchParams = new URLSearchParams(searchParams);
    nextSearchParams.delete('skillId');
    setSearchParams(nextSearchParams, { replace: true });
  };

  const openEnableModal = (skillId?: string) => {
    setPendingEnableSkillId(skillId?.trim() || '');
    setEnableModalOpen(true);
  };

  const handleEnable = async () => {
    const values = await enableForm.validateFields();
    setSaving(true);
    try {
      await axios.post(`${BACKEND_URL}/api/admin/skills/${values.skillId}/enable`, {
        channelScope: values.channelScope || [],
      }, {
        headers: getAuthHeaders(),
      });
      messageApi.success(t('enterprise_admin_skills_enable_success'));
      setEnableModalOpen(false);
      setPendingEnableSkillId('');
      enableForm.resetFields();
      clearPendingSkillId();
      await fetchCatalogSkills();
    } catch (error: any) {
      messageApi.error(getEnterpriseSkillErrorMessage(t('enterprise_admin_skills_enable_failed')));
    } finally {
      setSaving(false);
    }
  };

  const handleDisable = async (skillId: string) => {
    try {
      await axios.post(`${BACKEND_URL}/api/admin/skills/${skillId}/disable`, {}, {
        headers: getAuthHeaders(),
      });
      messageApi.success(t('enterprise_admin_skills_disable_success'));
      await fetchCatalogSkills();
    } catch (error: any) {
      messageApi.error(getEnterpriseSkillErrorMessage(t('enterprise_admin_skills_disable_failed')));
    }
  };

  const handleCreateSkill = async () => {
    const values = await createForm.validateFields();
    setSaving(true);
    try {
      await axios.post(`${BACKEND_URL}/api/admin/skills`, values, {
        headers: getAuthHeaders(),
      });
      messageApi.success(t('enterprise_admin_skills_create_success'));
      setCreateModalOpen(false);
      createForm.resetFields();
      await fetchGovernedSkills();
    } catch (error: any) {
      messageApi.error(getEnterpriseSkillErrorMessage(t('enterprise_admin_skills_create_failed')));
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
      await axios.post(`${BACKEND_URL}/api/admin/skills/${selectedSkill.skill.id}/versions`, {
        version: values.version,
        changeLog: values.changeLog || '',
      }, {
        headers: getAuthHeaders(),
      });
      messageApi.success(t('enterprise_admin_skills_version_create_success'));
      setVersionModalOpen(false);
      versionForm.resetFields();
      await Promise.all([
        fetchSkillDetail(selectedSkill.skill.id),
        fetchGovernedSkills(),
      ]);
    } catch (error: any) {
      messageApi.error(getEnterpriseSkillErrorMessage(t('enterprise_admin_skills_version_create_failed')));
    } finally {
      setSaving(false);
    }
  };

  const handleSubmitReview = async (versionId: string) => {
    if (!selectedSkill?.skill?.id) {
      return;
    }
    try {
      await axios.post(`${BACKEND_URL}/api/admin/skills/${selectedSkill.skill.id}/submit-review`, {
        skillVersionId: versionId,
      }, {
        headers: getAuthHeaders(),
      });
      messageApi.success(t('enterprise_admin_skills_submit_review_success'));
      await Promise.all([
        fetchSkillDetail(selectedSkill.skill.id),
        fetchGovernedSkills(),
      ]);
    } catch (error: any) {
      messageApi.error(getEnterpriseSkillErrorMessage(t('enterprise_admin_skills_submit_review_failed')));
    }
  };

  const handlePublish = async (versionId: string) => {
    if (!selectedSkill?.skill?.id) {
      return;
    }
    setPublishingId(versionId);
    try {
      await axios.post(`${BACKEND_URL}/api/admin/skills/${selectedSkill.skill.id}/publish`, {
        skillVersionId: versionId,
        releaseChannel: 'stable',
      }, {
        headers: getAuthHeaders(),
      });
      messageApi.success(t('enterprise_admin_skills_publish_success'));
      await Promise.all([
        fetchSkillDetail(selectedSkill.skill.id),
        fetchGovernedSkills(),
        fetchCatalogSkills(),
      ]);
    } catch (error: any) {
      messageApi.error(getEnterpriseSkillErrorMessage(t('enterprise_admin_skills_publish_failed')));
    } finally {
      setPublishingId('');
    }
  };

  return (
    <div>
      {contextHolder}
      <Card variant="borderless" style={{ borderRadius: 20 }}>
        <Space orientation="vertical" size={12} style={{ width: '100%', marginBottom: 16 }}>
          <Tag color="blue" style={{ width: 'fit-content', borderRadius: 999, paddingInline: 12 }}>
            {t('enterprise_admin_skills_scope_tag')}
          </Tag>
          <div>
            <Text strong style={{ display: 'block', fontSize: 16, marginBottom: 4 }}>
              {t('enterprise_admin_skills_title')}
            </Text>
            <Paragraph type="secondary" style={{ marginBottom: 0 }}>
              {t('enterprise_admin_skills_desc')}
            </Paragraph>
          </div>
        </Space>
        <Card size="small" style={{ marginBottom: 16, borderRadius: 16 }}>
          <Space wrap size={12}>
            <Button
              type={activeView === 'governance' ? 'primary' : 'default'}
              onClick={() => setActiveView('governance')}
            >
              {t('enterprise_admin_skills_view_governance')} ({governedSkills.length})
            </Button>
            <Button
              type={activeView === 'catalog' ? 'primary' : 'default'}
              onClick={() => setActiveView('catalog')}
            >
              {t('enterprise_admin_skills_view_catalog')} ({catalogSkills.length})
            </Button>
          </Space>
        </Card>

        {activeView === 'governance' ? (
          <Space orientation="vertical" size={16} style={{ width: '100%' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>
              <div>
                <Title level={5} style={{ marginBottom: 4 }}>
                  {t('enterprise_admin_skills_governance_title')}
                </Title>
                <Paragraph type="secondary" style={{ marginBottom: 0 }}>
                  {t('enterprise_admin_skills_governance_desc')}
                </Paragraph>
              </div>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => {
                  setCreateModalOpen(true);
                }}
              >
                {t('enterprise_admin_skills_action_create')}
              </Button>
            </div>
            <Table
              rowKey="id"
              loading={governanceLoading}
              dataSource={governedSkills}
              locale={{
                emptyText: <Empty description={t('enterprise_admin_skills_governance_empty')} image={Empty.PRESENTED_IMAGE_SIMPLE} />,
              }}
              columns={[
                { title: t('platform_skills_column_code'), dataIndex: 'code', key: 'code', width: 220 },
                { title: t('platform_skills_column_name'), dataIndex: 'name', key: 'name', width: 220 },
                { title: t('platform_skills_column_source'), dataIndex: 'sourceType', key: 'sourceType', width: 140 },
                {
                  title: t('platform_skills_column_trust_level'),
                  dataIndex: 'trustLevel',
                  key: 'trustLevel',
                  width: 160,
                  render: (value: string) => renderTrustTag(value, translateWithFallback),
                },
                {
                  title: t('platform_skills_column_status'),
                  dataIndex: 'status',
                  key: 'status',
                  width: 140,
                  render: (value: string) => renderStatusTag(value, translateWithFallback),
                },
                {
                  title: t('enterprise_admin_skills_column_published_version'),
                  dataIndex: 'latestPublishedVersion',
                  key: 'latestPublishedVersion',
                  width: 140,
                  render: (value?: string) => value || '-',
                },
                {
                  title: t('platform_skills_column_actions'),
                  key: 'actions',
                  width: 160,
                  render: (_: unknown, item: SkillItem) => (
                    <Button size="small" onClick={() => void fetchSkillDetail(item.id)}>
                      {t('platform_skills_view_detail')}
                    </Button>
                  ),
                },
              ]}
            />
          </Space>
        ) : (
          <Space orientation="vertical" size={16} style={{ width: '100%' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>
              <div>
                <Title level={5} style={{ marginBottom: 4 }}>
                  {t('enterprise_admin_skills_catalog_title')}
                </Title>
                <Paragraph type="secondary" style={{ marginBottom: 0 }}>
                  {t('enterprise_admin_skills_catalog_desc')}
                </Paragraph>
              </div>
              <Button
                type="primary"
                onClick={() => {
                  openEnableModal();
                }}
              >
                {t('enterprise_admin_skills_action_enable')}
              </Button>
            </div>
            <Table
              rowKey="id"
              loading={catalogLoading}
              dataSource={catalogSkills}
              locale={{
                emptyText: <Empty description={t('enterprise_admin_skills_catalog_empty')} image={Empty.PRESENTED_IMAGE_SIMPLE} />,
              }}
              columns={[
                { title: t('platform_skills_column_code'), dataIndex: 'code', key: 'code', width: 220 },
                { title: t('platform_skills_column_name'), dataIndex: 'name', key: 'name', width: 220 },
                { title: t('platform_skills_column_source'), dataIndex: 'sourceType', key: 'sourceType', width: 140 },
                {
                  title: t('platform_skills_column_trust_level'),
                  dataIndex: 'trustLevel',
                  key: 'trustLevel',
                  width: 160,
                  render: (value: string) => renderTrustTag(value, translateWithFallback),
                },
                {
                  title: t('enterprise_admin_skills_column_enablement'),
                  dataIndex: 'enablementStatus',
                  key: 'enablementStatus',
                  width: 140,
                  render: (value: string) => renderStatusTag(value, translateWithFallback),
                },
                {
                  title: t('enterprise_admin_skills_column_published_version'),
                  dataIndex: 'latestPublishedVersion',
                  key: 'latestPublishedVersion',
                  width: 140,
                  render: (value?: string) => value || '-',
                },
                {
                  title: t('platform_skills_column_actions'),
                  key: 'actions',
                  width: 160,
                  render: (_: unknown, item: SkillItem) => (
                    item.enablementStatus === 'enabled' ? (
                      <Button danger size="small" onClick={() => handleDisable(item.id)}>
                        {t('enterprise_admin_skills_action_disable')}
                      </Button>
                    ) : (
                      <Button size="small" type="primary" onClick={() => openEnableModal(item.id)}>
                        {t('enterprise_admin_skills_action_enable')}
                      </Button>
                    )
                  ),
                },
              ]}
            />
          </Space>
        )}
      </Card>

      <Modal
        title={t('enterprise_admin_skills_action_enable')}
        open={enableModalOpen}
        onCancel={() => {
          setEnableModalOpen(false);
          setPendingEnableSkillId('');
          clearPendingSkillId();
        }}
        onOk={handleEnable}
        confirmLoading={saving}
        okText={t('enterprise_admin_skills_action_enable')}
        destroyOnHidden
      >
        <Form form={enableForm} layout="vertical">
          <Form.Item label={t('enterprise_admin_skills_select')} name="skillId" rules={[{ required: true, message: t('enterprise_admin_skills_select_required') }]}>
            <Select
              showSearch
              optionFilterProp="label"
              options={enableableSkills.map((item) => ({
                label: `${item.code} · ${item.name}`,
                value: item.id,
              }))}
            />
          </Form.Item>
          <Form.Item label={t('enterprise_admin_skills_channel_scope')} name="channelScope">
            <Select
              mode="multiple"
              placeholder={t('enterprise_admin_skills_channel_scope_placeholder')}
              options={[
                { label: t('enterprise_admin_skills_channel_web'), value: 'web' },
                { label: t('enterprise_admin_skills_channel_im'), value: 'im' },
                { label: t('enterprise_admin_skills_channel_api'), value: 'api' },
              ]}
            />
          </Form.Item>
          <Space orientation="vertical" size={4}>
            <Text type="secondary">{t('enterprise_admin_skills_enable_hint')}</Text>
          </Space>
        </Form>
      </Modal>

      <Modal
        title={t('enterprise_admin_skills_action_create')}
        open={createModalOpen}
        onCancel={() => setCreateModalOpen(false)}
        onOk={handleCreateSkill}
        confirmLoading={saving}
        okText={t('agent_save')}
        cancelText={t('agent_cancel')}
        destroyOnHidden
      >
        <Form
          layout="vertical"
          form={createForm}
          initialValues={{ sourceType: 'builtin', providerType: 'native' }}
        >
          <Form.Item label={t('platform_skills_form_code')} name="code" rules={[{ required: true, message: t('platform_skills_form_code_required') }]}>
            <Input placeholder={t('enterprise_admin_skills_form_code_example')} />
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
                { label: t('platform_skills_source_builtin'), value: 'builtin' },
                { label: t('platform_skills_source_partner'), value: 'partner' },
                { label: t('platform_skills_source_openapi_catalog'), value: 'openapi_catalog' },
                { label: t('platform_skills_source_mcp_catalog'), value: 'mcp_catalog' },
              ]}
            />
          </Form.Item>
          <Form.Item label={t('platform_skills_form_provider_type')} name="providerType">
            <Select
              options={[
                { label: t('platform_skills_provider_native'), value: 'native' },
                { label: t('platform_skills_provider_openapi'), value: 'openapi' },
                { label: t('platform_skills_provider_mcp'), value: 'mcp' },
                { label: t('platform_skills_provider_remote_hosted'), value: 'remote_hosted' },
              ]}
            />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={selectedSkill?.skill?.name || t('enterprise_admin_skills_detail_title')}
        open={detailModalOpen}
        onCancel={() => setDetailModalOpen(false)}
        footer={null}
        width={960}
        destroyOnHidden
      >
        {selectedSkill?.skill ? (
          <Space orientation="vertical" size={16} style={{ width: '100%' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>
              <Space wrap size={12}>
                <Text strong>{selectedSkill.skill.code}</Text>
                {renderTrustTag(selectedSkill.skill.trustLevel, translateWithFallback)}
                {renderStatusTag(selectedSkill.skill.status, translateWithFallback)}
              </Space>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => {
                  setVersionModalOpen(true);
                }}
              >
                {t('enterprise_admin_skills_action_create_version')}
              </Button>
            </div>
            <Paragraph type="secondary" style={{ marginBottom: 0 }}>
              {selectedSkill.skill.description || t('enterprise_admin_skills_detail_empty_desc')}
            </Paragraph>
            <Table
              rowKey="id"
              loading={detailLoading}
              pagination={false}
              dataSource={selectedSkill.versions}
              locale={{
                emptyText: <Empty description={t('enterprise_admin_skills_versions_empty')} image={Empty.PRESENTED_IMAGE_SIMPLE} />,
              }}
              columns={[
                { title: t('enterprise_admin_skills_column_version'), dataIndex: 'version', key: 'version', width: 140 },
                {
                  title: t('enterprise_admin_skills_column_release_channel'),
                  dataIndex: 'releaseChannel',
                  key: 'releaseChannel',
                  width: 140,
                  render: (value: string) => t(`platform_skills_release_channel_${value}`, value || '-'),
                },
                {
                  title: t('platform_skills_column_status'),
                  dataIndex: 'releaseStatus',
                  key: 'releaseStatus',
                  width: 140,
                  render: (value: string) => renderStatusTag(value, translateWithFallback),
                },
                {
                  title: t('enterprise_admin_skills_column_change_log'),
                  dataIndex: 'changeLog',
                  key: 'changeLog',
                  render: (value?: string) => value || '-',
                },
                {
                  title: t('platform_skills_column_actions'),
                  key: 'actions',
                  width: 260,
                  render: (_: unknown, item: SkillVersionItem) => (
                    <Space>
                      {item.releaseStatus === 'draft' ? (
                        <Button size="small" onClick={() => void handleSubmitReview(item.id)}>
                          {t('platform_skills_submit_review')}
                        </Button>
                      ) : null}
                      {item.releaseStatus === 'draft' || item.releaseStatus === 'reviewing' ? (
                        <Button
                          size="small"
                          type="primary"
                          loading={publishingId === item.id}
                          onClick={() => void handlePublish(item.id)}
                        >
                          {t('platform_skills_publish')}
                        </Button>
                      ) : null}
                    </Space>
                  ),
                },
              ]}
            />
          </Space>
        ) : null}
      </Modal>

      <Modal
        title={t('enterprise_admin_skills_action_create_version')}
        open={versionModalOpen}
        onCancel={() => setVersionModalOpen(false)}
        onOk={handleCreateVersion}
        confirmLoading={saving}
        okText={t('agent_save')}
        cancelText={t('agent_cancel')}
        destroyOnHidden
      >
        <Form layout="vertical" form={versionForm}>
          <Form.Item label={t('enterprise_admin_skills_column_version')} name="version" rules={[{ required: true, message: t('enterprise_admin_skills_version_required') }]}>
            <Input placeholder={t('platform_skill_version_form_version_example')} />
          </Form.Item>
          <Form.Item label={t('enterprise_admin_skills_column_change_log')} name="changeLog">
            <TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default EnterpriseSkillsTab;
