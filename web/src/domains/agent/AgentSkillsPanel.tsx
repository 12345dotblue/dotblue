import React, { useEffect, useMemo, useState } from 'react';
import { Button, Card, Empty, Form, Input, Modal, Select, Space, Table, Tabs, Tag, Typography, message } from 'antd';
import axios from 'axios';
import { useTranslation } from 'react-i18next';
import { BACKEND_URL } from '../../config';

const { Paragraph, Text } = Typography;

interface PublishedSkillItem {
  id: string;
  code: string;
  name: string;
  sourceType: string;
  providerType: string;
  trustLevel: string;
  latestPublishedVersion: string;
  enablementStatus: string;
  agentInstalled: boolean;
  installedVersion: string;
  displayStatus: string;
  recommendedAction: string;
  blockReason: string;
  blockMessage: string;
}

interface InstalledSkillItem {
  id: string;
  skillId: string;
  skillVersionId: string;
  bindingStatus: string;
  entryAlias: string;
  invokeVisibility: string;
  skillCode: string;
  skillName: string;
  version: string;
}

interface InstallSkillFormValues {
  skillId: string;
  invokeVisibility: 'auto' | 'suggested' | 'manual';
  entryAlias?: string;
}

interface AgentSkillsPanelProps {
  agentId: string;
  authHeaders: Record<string, string | undefined>;
}

const AgentSkillsPanel: React.FC<AgentSkillsPanelProps> = ({ agentId, authHeaders }) => {
  const { t } = useTranslation();
  const [messageApi, contextHolder] = message.useMessage();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [installOpen, setInstallOpen] = useState(false);
  const [selectedCatalogSkill, setSelectedCatalogSkill] = useState<PublishedSkillItem | null>(null);
  const [installedSkills, setInstalledSkills] = useState<InstalledSkillItem[]>([]);
  const [catalogSkills, setCatalogSkills] = useState<PublishedSkillItem[]>([]);
  const [installForm] = Form.useForm<InstallSkillFormValues>();

  const loadData = async () => {
    if (!agentId) {
      return;
    }
    setLoading(true);
    try {
      const [installedRes, catalogRes] = await Promise.all([
        axios.get(`${BACKEND_URL}/api/admin/agents/${agentId}/skills`, { headers: authHeaders }),
        axios.get(`${BACKEND_URL}/api/admin/agents/${agentId}/skill-catalog`, { headers: authHeaders }),
      ]);
      setInstalledSkills(Array.isArray(installedRes.data) ? installedRes.data : []);
      setCatalogSkills(Array.isArray(catalogRes.data) ? catalogRes.data : []);
    } catch (error: any) {
      if (error?.response?.status === 403) {
        messageApi.warning(t('agent_skills_panel_enterprise_admin_only'));
      } else {
        messageApi.error(t('agent_skills_panel_load_failed'));
      }
      setInstalledSkills([]);
      setCatalogSkills([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [agentId]);

  useEffect(() => {
    if (!installOpen) {
      setSelectedCatalogSkill(null);
      return;
    }
    installForm.resetFields();
    installForm.setFieldsValue({ invokeVisibility: 'auto' });
  }, [installOpen, installForm]);

  const installableSkills = useMemo(() => {
    return catalogSkills.filter((item) => item.displayStatus === 'enabled_installable');
  }, [catalogSkills]);

  const pendingEnableSkills = useMemo(
    () => catalogSkills.filter((item) => item.displayStatus === 'imported_pending_enable'),
    [catalogSkills],
  );

  const catalogInstalledCount = useMemo(
    () => catalogSkills.filter((item) => item.displayStatus === 'installed').length,
    [catalogSkills],
  );

  const openInstallFlow = (skill: PublishedSkillItem) => {
    setSelectedCatalogSkill(skill);
    setInstallOpen(true);
    installForm.resetFields();
    installForm.setFieldsValue({ invokeVisibility: 'auto' });
  };

  const handleInstall = async () => {
    if (!selectedCatalogSkill?.id) {
      return;
    }
    const values = await installForm.validateFields();
    setSaving(true);
    try {
      await axios.post(`${BACKEND_URL}/api/admin/agents/${agentId}/skills/ensure-installed`, {
        skillId: selectedCatalogSkill.id,
        entryAlias: values.entryAlias,
        invokeVisibility: values.invokeVisibility,
      }, {
        headers: authHeaders,
      });
      messageApi.success(
        selectedCatalogSkill.displayStatus === 'imported_pending_enable'
          ? t('agent_skill_catalog_enable_install_success')
          : t('agent_skills_panel_install_success'),
      );
      setInstallOpen(false);
      setSelectedCatalogSkill(null);
      installForm.resetFields();
      await loadData();
    } catch (error: any) {
      if (error?.response?.status === 403) {
        messageApi.error(t('agent_skill_catalog_action_denied'));
      } else {
        messageApi.error(t('agent_skills_panel_install_failed'));
      }
    } finally {
      setSaving(false);
    }
  };

  const handleUninstall = async (skillId: string) => {
    try {
      await axios.post(`${BACKEND_URL}/api/admin/agents/${agentId}/skills/${skillId}/uninstall`, {}, {
        headers: authHeaders,
      });
      messageApi.success(t('agent_skills_panel_uninstall_success'));
      await loadData();
    } catch (error: any) {
      messageApi.error(t('agent_skills_panel_uninstall_failed'));
    }
  };

  const translateSourceType = (value?: string) => {
    if (!value) {
      return '-';
    }
    return t(`platform_skills_source_${value}`, value);
  };

  const translateStatus = (value: string) => {
    switch (value) {
      case 'installed':
        return t('agent_skill_catalog_status_installed');
      case 'enabled_installable':
        return t('agent_skill_catalog_status_installable');
      case 'imported_pending_enable':
        return t('agent_skill_catalog_status_pending_enable');
      case 'blocked':
        return t('agent_skill_catalog_status_blocked');
      default:
        return t('agent_skill_catalog_status_unavailable');
    }
  };

  const renderCatalogStatusTag = (value: string) => {
    let color = 'default';
    if (value === 'installed') {
      color = 'success';
    } else if (value === 'enabled_installable') {
      color = 'processing';
    } else if (value === 'imported_pending_enable') {
      color = 'gold';
    } else if (value === 'blocked') {
      color = 'error';
    }
    return <Tag color={color}>{translateStatus(value)}</Tag>;
  };

  const getCatalogHint = (record: PublishedSkillItem) => {
    if (record.displayStatus === 'installed') {
      return t('agent_skill_catalog_hint_installed', {
        version: record.installedVersion || record.latestPublishedVersion || '-',
      });
    }
    if (record.displayStatus === 'enabled_installable') {
      return t('agent_skill_catalog_hint_installable');
    }
    if (record.displayStatus === 'imported_pending_enable') {
      return t('agent_skill_catalog_hint_pending_enable');
    }
    if (record.blockReason === 'skill_blocked') {
      return t('agent_skill_catalog_hint_blocked');
    }
    if (record.blockReason === 'skill_not_published') {
      return t('agent_skill_catalog_hint_not_published');
    }
    return record.blockMessage || t('agent_skill_catalog_hint_unavailable');
  };

  const renderCatalogAction = (item: PublishedSkillItem) => {
    if (item.displayStatus === 'installed') {
      return <Tag color="success">{t('agent_skill_catalog_action_installed')}</Tag>;
    }
    if (item.displayStatus === 'enabled_installable') {
      return (
        <Button size="small" type="primary" onClick={() => openInstallFlow(item)}>
          {t('agent_skill_catalog_action_install')}
        </Button>
      );
    }
    if (item.displayStatus === 'imported_pending_enable') {
      return (
        <Button size="small" type="primary" onClick={() => openInstallFlow(item)}>
          {t('agent_skill_catalog_action_enable_install')}
        </Button>
      );
    }
    return (
      <Text type="secondary">
        {t('agent_skill_catalog_action_unavailable')}
      </Text>
    );
  };

  return (
    <div>
      {contextHolder}
      <Space orientation="vertical" size={16} style={{ width: '100%' }}>
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
            gap: 12,
          }}
        >
          <Card size="small" style={{ borderRadius: 16 }}>
            <Text type="secondary">{t('agent_skill_panel_summary_installed')}</Text>
            <Paragraph style={{ fontSize: 24, fontWeight: 600, margin: '8px 0 0' }}>{installedSkills.length}</Paragraph>
          </Card>
          <Card size="small" style={{ borderRadius: 16 }}>
            <Text type="secondary">{t('agent_skill_panel_summary_installable')}</Text>
            <Paragraph style={{ fontSize: 24, fontWeight: 600, margin: '8px 0 0' }}>{installableSkills.length}</Paragraph>
          </Card>
          <Card size="small" style={{ borderRadius: 16 }}>
            <Text type="secondary">{t('agent_skill_panel_summary_pending_enable')}</Text>
            <Paragraph style={{ fontSize: 24, fontWeight: 600, margin: '8px 0 0' }}>{pendingEnableSkills.length}</Paragraph>
          </Card>
          <Card size="small" style={{ borderRadius: 16 }}>
            <Text type="secondary">{t('agent_skill_panel_summary_catalog_installed')}</Text>
            <Paragraph style={{ fontSize: 24, fontWeight: 600, margin: '8px 0 0' }}>{catalogInstalledCount}</Paragraph>
          </Card>
        </div>
        <Card
          size="small"
          style={{
            borderRadius: 20,
            background: 'linear-gradient(135deg, rgba(22,119,255,0.08), rgba(22,119,255,0.02))',
            border: '1px solid rgba(22,119,255,0.12)',
          }}
        >
          <Space direction="vertical" size={10} style={{ width: '100%' }}>
            <Text strong>{t('agent_skill_panel_guide_title')}</Text>
            <Paragraph type="secondary" style={{ marginBottom: 0 }}>
              {t('agent_skill_panel_guide_desc')}
            </Paragraph>
            <Space wrap size={[8, 8]}>
              <Tag color={pendingEnableSkills.length > 0 ? 'gold' : 'default'}>{t('agent_skill_panel_step_publish')}</Tag>
              <Tag color={pendingEnableSkills.length > 0 ? 'processing' : 'default'}>{t('agent_skill_panel_step_enable')}</Tag>
              <Tag color={installableSkills.length > 0 ? 'success' : 'default'}>{t('agent_skill_panel_step_install')}</Tag>
              <Tag>{t('agent_skill_panel_step_verify')}</Tag>
            </Space>
          </Space>
        </Card>
        <Paragraph type="secondary" style={{ marginBottom: 0, maxWidth: 860 }}>
          {t('agent_skill_panel_helper_text')}
        </Paragraph>
        <Tabs
          defaultActiveKey="catalog"
          items={[
            {
              key: 'catalog',
              label: t('agent_skill_panel_tab_catalog'),
              children: (
                <Table
                  rowKey="id"
                  loading={loading}
                  dataSource={catalogSkills}
                  locale={{
                    emptyText: <Empty description={t('agent_skill_catalog_empty')} image={Empty.PRESENTED_IMAGE_SIMPLE} />,
                  }}
                  pagination={false}
                  columns={[
                    { title: t('agent_skills_panel_code_column'), dataIndex: 'code', key: 'code', width: 220 },
                    { title: t('agent_skills_panel_name_column'), dataIndex: 'name', key: 'name', width: 220 },
                    {
                      title: t('agent_skill_catalog_source_column'),
                      dataIndex: 'sourceType',
                      key: 'sourceType',
                      width: 140,
                      render: (value: string) => translateSourceType(value),
                    },
                    {
                      title: t('agent_skill_catalog_version_column'),
                      dataIndex: 'latestPublishedVersion',
                      key: 'latestPublishedVersion',
                      width: 120,
                      render: (value: string) => value || <Text type="secondary">-</Text>,
                    },
                    {
                      title: t('agent_skill_catalog_status_column'),
                      dataIndex: 'displayStatus',
                      key: 'displayStatus',
                      width: 160,
                      render: (value: string) => renderCatalogStatusTag(value),
                    },
                    {
                      title: t('agent_skill_catalog_hint_column'),
                      dataIndex: 'blockMessage',
                      key: 'blockMessage',
                      render: (_value: string, record: PublishedSkillItem) => (
                        <Text type="secondary">
                          {getCatalogHint(record)}
                        </Text>
                      ),
                    },
                    {
                      title: t('agent_skills_panel_actions_column'),
                      key: 'actions',
                      width: 180,
                      render: (_: unknown, record: PublishedSkillItem) => renderCatalogAction(record),
                    },
                  ]}
                />
              ),
            },
            {
              key: 'installed',
              label: t('agent_skill_panel_tab_installed'),
              children: (
                <Table
                  rowKey="id"
                  loading={loading}
                  dataSource={installedSkills}
                  locale={{
                    emptyText: <Empty description={t('agent_skill_panel_table_empty')} image={Empty.PRESENTED_IMAGE_SIMPLE} />,
                  }}
                  pagination={false}
                  columns={[
                    { title: t('agent_skills_panel_code_column'), dataIndex: 'skillCode', key: 'skillCode', width: 220 },
                    { title: t('agent_skills_panel_name_column'), dataIndex: 'skillName', key: 'skillName', width: 220 },
                    { title: t('agent_skills_panel_version_column'), dataIndex: 'version', key: 'version', width: 140 },
                    {
                      title: t('agent_skills_panel_invoke_mode_column'),
                      dataIndex: 'invokeVisibility',
                      key: 'invokeVisibility',
                      width: 120,
                      render: (value: string) => (
                        <Tag color={value === 'manual' ? 'default' : value === 'suggested' ? 'processing' : 'success'}>
                          {value === 'manual'
                            ? t('agent_skill_panel_visibility_manual')
                            : value === 'suggested'
                              ? t('agent_skill_panel_visibility_suggested')
                              : t('agent_skill_panel_visibility_auto')}
                        </Tag>
                      ),
                    },
                    {
                      title: t('agent_skills_panel_alias'),
                      dataIndex: 'entryAlias',
                      key: 'entryAlias',
                      width: 160,
                      render: (value?: string) => value || <Text type="secondary">-</Text>,
                    },
                    {
                      title: t('agent_skills_panel_status_column'),
                      dataIndex: 'bindingStatus',
                      key: 'bindingStatus',
                      width: 120,
                      render: (value: string) => <Tag color={value === 'installed' ? 'success' : 'default'}>{value === 'installed' ? t('agent_skills_panel_installed_status') : value}</Tag>,
                    },
                    {
                      title: t('agent_skills_panel_actions_column'),
                      key: 'actions',
                      width: 120,
                      render: (_: unknown, record: InstalledSkillItem) => (
                        <Button danger size="small" onClick={() => handleUninstall(record.skillId)}>
                          {t('agent_skills_panel_uninstall_action')}
                        </Button>
                      ),
                    },
                  ]}
                />
              ),
            },
          ]}
        />
      </Space>

      <Modal
        title={selectedCatalogSkill?.displayStatus === 'imported_pending_enable'
          ? t('agent_skill_catalog_enable_install_modal_title')
          : t('agent_skill_panel_modal_title')}
        open={installOpen}
        onCancel={() => {
          setInstallOpen(false);
          setSelectedCatalogSkill(null);
        }}
        onOk={handleInstall}
        confirmLoading={saving}
        okText={selectedCatalogSkill?.displayStatus === 'imported_pending_enable'
          ? t('agent_skill_catalog_action_enable_install')
          : t('agent_skill_panel_action_install')}
        destroyOnHidden
      >
        <Form form={installForm} layout="vertical" initialValues={{ invokeVisibility: 'auto' }}>
          <Form.Item label={t('agent_skill_panel_form_skill')}>
            <Input value={selectedCatalogSkill ? `${selectedCatalogSkill.code} · ${selectedCatalogSkill.name}` : ''} disabled />
          </Form.Item>
          {selectedCatalogSkill?.displayStatus === 'imported_pending_enable' ? (
            <Paragraph type="secondary" style={{ marginTop: 0 }}>
              {t('agent_skill_catalog_enable_install_modal_desc')}
            </Paragraph>
          ) : null}
          <Form.Item label={t('agent_skill_panel_form_alias')} name="entryAlias">
            <Input placeholder={t('agent_skill_panel_form_alias_placeholder')} />
          </Form.Item>
          <Form.Item label={t('agent_skill_panel_form_visibility')} name="invokeVisibility">
            <Select
              options={[
                { label: t('agent_skill_panel_visibility_auto'), value: 'auto' },
                { label: t('agent_skill_panel_visibility_suggested'), value: 'suggested' },
                { label: t('agent_skill_panel_visibility_manual'), value: 'manual' },
              ]}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default AgentSkillsPanel;

