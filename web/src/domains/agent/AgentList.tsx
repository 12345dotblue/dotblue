import React, { useState, useEffect } from 'react';
import { Card, Button, List, Modal, Form, Input, message, Typography, Space, Empty, Popconfirm } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, RobotOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import { BACKEND_URL } from '../../config';

const { Title, Paragraph, Text } = Typography;
const { TextArea } = Input;

interface AgentItem {
  id: string;
  agentName: string;
  systemPrompt: string;
  createdAt: string;
}

const AgentList: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [agents, setAgents] = useState<AgentItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingAgent, setEditingAgent] = useState<AgentItem | null>(null);
  const [form] = Form.useForm();
  const [saving, setSaving] = useState(false);

  const fetchAgents = () => {
    const token = localStorage.getItem('casdoor_token');
    axios.get(`${BACKEND_URL}/api/agents`, {
      headers: { Authorization: `Bearer ${token}` },
    }).then(res => {
      setAgents(res.data || []);
    }).catch(() => {
      setAgents([]);
    }).finally(() => {
      setLoading(false);
    });
  };

  useEffect(() => {
    fetchAgents();
  }, []);

  const openCreate = () => {
    setEditingAgent(null);
    form.resetFields();
    setModalOpen(true);
  };

  const openEdit = (agent: AgentItem) => {
    setEditingAgent(agent);
    form.setFieldsValue({
      agentName: agent.agentName,
      systemPrompt: agent.systemPrompt,
    });
    setModalOpen(true);
  };

  const handleSave = async () => {
    try {
      const values = await form.validateFields();
      setSaving(true);
      const token = localStorage.getItem('casdoor_token');

      if (editingAgent) {
        await axios.put(`${BACKEND_URL}/api/agents/${editingAgent.id}`, values, {
          headers: { Authorization: `Bearer ${token}` },
        });
        message.success(t('agent_update_success'));
      } else {
        await axios.post(`${BACKEND_URL}/api/agents`, values, {
          headers: { Authorization: `Bearer ${token}` },
        });
        message.success(t('agent_create_success'));
      }

      setModalOpen(false);
      fetchAgents();
    } catch {
      // validation error or API error
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (agentId: string) => {
    try {
      const token = localStorage.getItem('casdoor_token');
      await axios.delete(`${BACKEND_URL}/api/agents/${agentId}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      message.success(t('agent_delete_success'));
      fetchAgents();
    } catch {
      message.error(t('agent_delete_failed'));
    }
  };

  return (
    <div style={{ animation: 'fadeIn 0.5s ease-out' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24, maxWidth: 800 }}>
        <div>
          <Title level={4} style={{ marginBottom: 4 }}>
            <RobotOutlined style={{ marginRight: 8 }} />
            {t('agent_list_title')}
          </Title>
          <Text type="secondary">{t('agent_list_desc')}</Text>
        </div>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
          {t('agent_create')}
        </Button>
      </div>

      {agents.length === 0 && !loading ? (
        <Card bordered={false} style={{ borderRadius: 12, maxWidth: 800, textAlign: 'center', padding: '40px 0' }}>
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description={t('agent_no_agents')}
          >
            <Space>
              <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
                {t('agent_create_first')}
              </Button>
              <Button onClick={() => navigate('/chat')}>
                {t('agent_go_chat')}
              </Button>
            </Space>
          </Empty>
        </Card>
      ) : (
        <List
          loading={loading}
          dataSource={agents}
          style={{ maxWidth: 800 }}
          renderItem={(item) => (
            <Card
              bordered={false}
              style={{ borderRadius: 12, marginBottom: 16, boxShadow: '0 4px 20px rgba(0,0,0,0.03)' }}
            >
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <Space>
                    <RobotOutlined style={{ color: '#1677ff', fontSize: 20 }} />
                    <Title level={5} style={{ margin: 0 }}>{item.agentName}</Title>
                  </Space>
                  <Paragraph
                    type="secondary"
                    ellipsis={{ rows: 2 }}
                    style={{ marginTop: 8, marginBottom: 0 }}
                  >
                    {item.systemPrompt}
                  </Paragraph>
                </div>
                <Space style={{ marginLeft: 16, flexShrink: 0 }}>
                  <Button icon={<EditOutlined />} onClick={() => openEdit(item)}>
                    {t('agent_edit')}
                  </Button>
                  <Popconfirm
                    title={t('agent_confirm_delete')}
                    onConfirm={() => handleDelete(item.id)}
                    okText={t('agent_delete')}
                    cancelText={t('agent_cancel')}
                  >
                    <Button danger icon={<DeleteOutlined />}>
                      {t('agent_delete')}
                    </Button>
                  </Popconfirm>
                </Space>
              </div>
            </Card>
          )}
        />
      )}

      <Modal
        title={editingAgent ? t('agent_edit') : t('agent_create')}
        open={modalOpen}
        onOk={handleSave}
        onCancel={() => setModalOpen(false)}
        confirmLoading={saving}
        okText={editingAgent ? t('agent_save') : t('agent_create')}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          initialValues={{ agentName: '', systemPrompt: '' }}
        >
          <Form.Item
            label={t('agent_name')}
            name="agentName"
            rules={[{ required: true, message: t('agent_name_required') }]}
          >
            <Input placeholder={t('placeholder_agent_name')} />
          </Form.Item>
          <Form.Item
            label={t('system_prompt')}
            name="systemPrompt"
            rules={[{ required: true, message: t('system_prompt_required') }]}
          >
            <TextArea rows={6} placeholder={t('placeholder_system_prompt')} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default AgentList;
