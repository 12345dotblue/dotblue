import React, { useEffect, useMemo, useState } from 'react';
import { Button, Card, Empty, Form, Input, Modal, Select, Space, Table, Tag, Typography, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import axios from 'axios';
import { useTranslation } from 'react-i18next';
import { BACKEND_URL } from '../../config';

const { Paragraph, Text } = Typography;

interface PublishedSkillItem {
  id: string;
  code: string;
  name: string;
  latestPublishedVersion: string;
  enablementStatus: string;
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
  marketHref?: string;
  builderHref?: string;
}

const AgentSkillsPanel: React.FC<AgentSkillsPanelProps> = ({ agentId, authHeaders, marketHref, builderHref }) => {
  const { t } = useTranslation();
  const [messageApi, contextHolder] = message.useMessage();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [installOpen, setInstallOpen] = useState(false);
  const [installedSkills, setInstalledSkills] = useState<InstalledSkillItem[]>([]);
  const [publishedSkills, setPublishedSkills] = useState<PublishedSkillItem[]>([]);
  const [installForm] = Form.useForm<InstallSkillFormValues>();

  const loadData = async () => {
    if (!agentId) {
      return;
    }
    setLoading(true);
    try {
      const [installedRes, publishedRes] = await Promise.all([
        axios.get(`${BACKEND_URL}/api/admin/agents/${agentId}/skills`, { headers: authHeaders }),
        axios.get(`${BACKEND_URL}/api/admin/skills?view=catalog`, { headers: authHeaders }),
      ]);
      setInstalledSkills(Array.isArray(installedRes.data) ? installedRes.data : []);
      setPublishedSkills(Array.isArray(publishedRes.data) ? publishedRes.data : []);
    } catch (error: any) {
      if (error?.response?.status === 403) {
        messageApi.warning(t('agent_skills_panel_enterprise_admin_only'));
      } else {
        messageApi.error(t('agent_skills_panel_load_failed'));
      }
      setInstalledSkills([]);
      setPublishedSkills([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [agentId]);

  useEffect(() => {
    if (!installOpen) {
      return;
    }
    installForm.resetFields();
    installForm.setFieldsValue({ invokeVisibility: 'auto' });
  }, [installOpen, installForm]);

  const installableSkills = useMemo(() => {
    const installedIds = new Set(installedSkills.map((item) => item.skillId));
    return publishedSkills.filter((item) => item.enablementStatus === 'enabled' && !installedIds.has(item.id));
  }, [installedSkills, publishedSkills]);

  const enabledSkillsCount = useMemo(
    () => publishedSkills.filter((item) => item.enablementStatus === 'enabled').length,
    [publishedSkills],
  );

  const handleInstall = async () => {
    const values = await installForm.validateFields();
    setSaving(true);
    try {
      await axios.post(`${BACKEND_URL}/api/admin/agents/${agentId}/skills/install`, {
        skillId: values.skillId,
        entryAlias: values.entryAlias,
        invokeVisibility: values.invokeVisibility,
      }, {
        headers: authHeaders,
      });
      messageApi.success(t('agent_skills_panel_install_success'));
      setInstallOpen(false);
      installForm.resetFields();
      await loadData();
    } catch (error: any) {
      messageApi.error(t('agent_skills_panel_install_failed'));
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
            <Text type="secondary">{t('agent_skill_panel_summary_enabled')}</Text>
            <Paragraph style={{ fontSize: 24, fontWeight: 600, margin: '8px 0 0' }}>{enabledSkillsCount}</Paragraph>
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
              <Tag color={enabledSkillsCount > 0 ? 'processing' : 'default'}>{t('agent_skill_panel_step_enable')}</Tag>
              <Tag color={installableSkills.length > 0 ? 'success' : 'default'}>{t('agent_skill_panel_step_install')}</Tag>
              <Tag>{t('agent_skill_panel_step_verify')}</Tag>
            </Space>
            <Space wrap>
              <Button icon={<PlusOutlined />} type="primary" onClick={() => setInstallOpen(true)}>
                {t('agent_skill_panel_action_install')}
              </Button>
              {marketHref ? (
                <Button href={marketHref}>{t('agent_skill_panel_action_market')}</Button>
              ) : null}
              {builderHref ? (
                <Button href={builderHref}>{t('agent_skill_panel_action_builder')}</Button>
              ) : null}
            </Space>
          </Space>
        </Card>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>
          <Paragraph type="secondary" style={{ marginBottom: 0, maxWidth: 720 }}>
            {t('agent_skill_panel_helper_text')}
          </Paragraph>
          <Button icon={<PlusOutlined />} type="primary" onClick={() => setInstallOpen(true)}>
            {t('agent_skill_panel_action_install')}
          </Button>
        </div>
        {!installedSkills.length ? (
          <Card size="small" style={{ borderRadius: 16 }}>
            <Empty
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              description={
                <Space direction="vertical" size={4}>
                  <Text strong>{t('agent_skill_panel_empty_title')}</Text>
                  <Text type="secondary">{t('agent_skill_panel_empty_desc')}</Text>
                </Space>
              }
            >
              <Space wrap>
                {installableSkills.length ? (
                  <Button type="primary" onClick={() => setInstallOpen(true)}>
                    {t('agent_skill_panel_action_install')}
                  </Button>
                ) : null}
                {marketHref ? (
                  <Button href={marketHref}>{t('agent_skill_panel_action_market')}</Button>
                ) : null}
                {builderHref ? (
                  <Button href={builderHref}>{t('agent_skill_panel_action_builder')}</Button>
                ) : null}
              </Space>
            </Empty>
          </Card>
        ) : null}
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
      </Space>

      <Modal
        title={t('agent_skill_panel_modal_title')}
        open={installOpen}
        onCancel={() => setInstallOpen(false)}
        onOk={handleInstall}
        confirmLoading={saving}
        okText={t('agent_skill_panel_action_install')}
        destroyOnHidden
      >
        <Form form={installForm} layout="vertical" initialValues={{ invokeVisibility: 'auto' }}>
          <Form.Item label={t('agent_skill_panel_form_skill')} name="skillId" rules={[{ required: true, message: t('agent_skill_panel_form_skill_required') }]}>
            <Select
              showSearch
              optionFilterProp="label"
              notFoundContent={t('agent_skill_panel_form_skill_empty')}
              options={installableSkills.map((item) => ({
                label: `${item.code} · ${item.name}`,
                value: item.id,
              }))}
            />
          </Form.Item>
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

