import React, { useEffect, useMemo, useState } from 'react';
import { Button, Empty, Form, Input, Modal, Select, Space, Table, Tag, Typography, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import axios from 'axios';
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
}

const AgentSkillsPanel: React.FC<AgentSkillsPanelProps> = ({ agentId, authHeaders }) => {
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
        messageApi.warning('只有企业管理员可以管理 Agent 技能安装');
      } else {
        messageApi.error('加载 Agent 技能失败');
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

  const installableSkills = useMemo(() => {
    const installedIds = new Set(installedSkills.map((item) => item.skillId));
    return publishedSkills.filter((item) => item.enablementStatus === 'enabled' && !installedIds.has(item.id));
  }, [installedSkills, publishedSkills]);

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
      messageApi.success('技能安装成功');
      setInstallOpen(false);
      installForm.resetFields();
      await loadData();
    } catch (error: any) {
      const errorText = error?.response?.data;
      messageApi.error(typeof errorText === 'string' ? errorText : '技能安装失败');
    } finally {
      setSaving(false);
    }
  };

  const handleUninstall = async (skillId: string) => {
    try {
      await axios.post(`${BACKEND_URL}/api/admin/agents/${agentId}/skills/${skillId}/uninstall`, {}, {
        headers: authHeaders,
      });
      messageApi.success('技能已卸载');
      await loadData();
    } catch (error: any) {
      const errorText = error?.response?.data;
      messageApi.error(typeof errorText === 'string' ? errorText : '技能卸载失败');
    }
  };

  return (
    <div>
      {contextHolder}
      <Space orientation="vertical" size={16} style={{ width: '100%' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>
          <Paragraph type="secondary" style={{ marginBottom: 0 }}>
            已安装技能会在当前 Agent 上生效。企业未启用的 Skill 不可安装。
          </Paragraph>
          <Button icon={<PlusOutlined />} type="primary" onClick={() => {
            installForm.setFieldsValue({ invokeVisibility: 'auto' });
            setInstallOpen(true);
          }}>
            安装 Skill
          </Button>
        </div>
        <Table
          rowKey="id"
          loading={loading}
          dataSource={installedSkills}
          locale={{
            emptyText: <Empty description="当前 Agent 暂未安装 Skill" image={Empty.PRESENTED_IMAGE_SIMPLE} />,
          }}
          pagination={false}
          columns={[
            { title: 'Code', dataIndex: 'skillCode', key: 'skillCode', width: 220 },
            { title: '名称', dataIndex: 'skillName', key: 'skillName', width: 220 },
            { title: '版本', dataIndex: 'version', key: 'version', width: 140 },
            {
              title: '调用方式',
              dataIndex: 'invokeVisibility',
              key: 'invokeVisibility',
              width: 120,
              render: (value: string) => <Tag color={value === 'manual' ? 'default' : value === 'suggested' ? 'processing' : 'success'}>{value}</Tag>,
            },
            {
              title: '别名',
              dataIndex: 'entryAlias',
              key: 'entryAlias',
              width: 160,
              render: (value?: string) => value || <Text type="secondary">-</Text>,
            },
            {
              title: '状态',
              dataIndex: 'bindingStatus',
              key: 'bindingStatus',
              width: 120,
              render: (value: string) => <Tag color={value === 'installed' ? 'success' : 'default'}>{value}</Tag>,
            },
            {
              title: '操作',
              key: 'actions',
              width: 120,
              render: (_: unknown, record: InstalledSkillItem) => (
                <Button danger size="small" onClick={() => handleUninstall(record.skillId)}>
                  卸载
                </Button>
              ),
            },
          ]}
        />
      </Space>

      <Modal
        title="安装 Skill"
        open={installOpen}
        onCancel={() => setInstallOpen(false)}
        onOk={handleInstall}
        confirmLoading={saving}
        okText="安装"
        destroyOnHidden
      >
        <Form form={installForm} layout="vertical" initialValues={{ invokeVisibility: 'auto' }}>
          <Form.Item label="Skill" name="skillId" rules={[{ required: true, message: '请选择 Skill' }]}>
            <Select
              showSearch
              optionFilterProp="label"
              options={installableSkills.map((item) => ({
                label: `${item.code} · ${item.name}`,
                value: item.id,
              }))}
            />
          </Form.Item>
          <Form.Item label="Agent 内显示别名" name="entryAlias">
            <Input placeholder="可选，例如：知识助手" />
          </Form.Item>
          <Form.Item label="调用方式" name="invokeVisibility">
            <Select
              options={[
                { label: '自动', value: 'auto' },
                { label: '建议调用', value: 'suggested' },
                { label: '手动', value: 'manual' },
              ]}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default AgentSkillsPanel;
