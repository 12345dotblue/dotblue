import React, { useEffect, useState } from 'react';
import { Button, Card, Form, Input, Modal, Popconfirm, Select, Space, Table, Typography, message } from 'antd';
import { CloudServerOutlined, DeleteOutlined, EditOutlined, PlusOutlined } from '@ant-design/icons';
import axios from 'axios';
import { useTranslation } from 'react-i18next';
import { BACKEND_URL } from '../../config';
import { casdoorService } from '../identity/CasdoorService';

const { Paragraph, Text } = Typography;

interface EnterpriseLLMModel {
  id: string;
  displayName: string;
  type: 'openai' | 'anthropic';
  apiBase: string;
  apiKey: string;
  model: string;
}

interface Props {
  createSignal?: number;
}

function getAuthHeaders() {
  const token = casdoorService.getToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

const EnterpriseLLMSettingsTab: React.FC<Props> = ({ createSignal = 0 }) => {
  const { t } = useTranslation();
  const [messageApi, contextHolder] = message.useMessage();
  const [models, setModels] = useState<EnterpriseLLMModel[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingModel, setEditingModel] = useState<EnterpriseLLMModel | null>(null);
  const [form] = Form.useForm<EnterpriseLLMModel>();
  const providerType = Form.useWatch('type', form);

  const loadModels = async () => {
    setLoading(true);
    try {
      const res = await axios.get(`${BACKEND_URL}/api/admin/llm-models`, {
        headers: getAuthHeaders(),
      });
      setModels(Array.isArray(res.data) ? res.data : []);
    } catch {
      messageApi.error(t('enterprise_admin_llm_load_failed'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadModels();
  }, []);

  useEffect(() => {
    if (createSignal > 0) {
      setEditingModel(null);
      setModalOpen(true);
    }
  }, [createSignal]);

  useEffect(() => {
    if (!modalOpen) {
      return;
    }
    form.setFieldsValue(editingModel || {
      displayName: '',
      type: 'openai',
      apiBase: 'https://api.openai.com/v1',
      apiKey: '',
      model: '',
    });
  }, [modalOpen, editingModel, form]);

  const openCreate = () => {
    setEditingModel(null);
    setModalOpen(true);
  };

  const openEdit = (item: EnterpriseLLMModel) => {
    setEditingModel(item);
    setModalOpen(true);
  };

  const handleSave = async () => {
    const values = await form.validateFields();
    setSaving(true);
    try {
      if (editingModel) {
        await axios.put(`${BACKEND_URL}/api/admin/llm-models/${editingModel.id}`, values, {
          headers: getAuthHeaders(),
        });
        messageApi.success(t('enterprise_admin_llm_updated'));
      } else {
        await axios.post(`${BACKEND_URL}/api/admin/llm-models`, values, {
          headers: getAuthHeaders(),
        });
        messageApi.success(t('enterprise_admin_llm_created'));
      }
      setModalOpen(false);
      await loadModels();
    } catch (error: any) {
      const errorText = error?.response?.data;
      messageApi.error(typeof errorText === 'string' ? errorText : t('enterprise_admin_llm_save_failed'));
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await axios.delete(`${BACKEND_URL}/api/admin/llm-models/${id}`, {
        headers: getAuthHeaders(),
      });
      messageApi.success(t('enterprise_admin_llm_deleted'));
      await loadModels();
    } catch {
      messageApi.error(t('enterprise_admin_llm_delete_failed'));
    }
  };

  return (
    <Space orientation="vertical" size={16} style={{ width: '100%' }}>
      {contextHolder}
      <Card variant="borderless" style={{ borderRadius: 20 }}>
        <Paragraph type="secondary" style={{ marginBottom: 0 }}>
          {t('enterprise_admin_llm_desc')}
        </Paragraph>
      </Card>
      <Card
        variant="borderless"
        title={<Space><CloudServerOutlined />{t('enterprise_admin_llm_title')}</Space>}
        extra={<Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>{t('enterprise_admin_llm_create')}</Button>}
        style={{ borderRadius: 20 }}
      >
        <Table
          rowKey="id"
          loading={loading}
          pagination={false}
          dataSource={models}
          scroll={{ x: 920 }}
          columns={[
            {
              title: t('enterprise_admin_llm_name'),
              dataIndex: 'displayName',
              key: 'displayName',
              render: (value: string, item: EnterpriseLLMModel) => (
                <Space orientation="vertical" size={0}>
                  <Text strong>{value}</Text>
                  <Text type="secondary">{item.model}</Text>
                </Space>
              ),
            },
            {
              title: t('enterprise_admin_llm_provider'),
              dataIndex: 'type',
              key: 'type',
              width: 140,
            },
            {
              title: t('enterprise_admin_llm_api_base'),
              dataIndex: 'apiBase',
              key: 'apiBase',
              ellipsis: true,
            },
            {
              title: t('enterprise_admin_llm_actions'),
              key: 'actions',
              width: 180,
              render: (_: unknown, item: EnterpriseLLMModel) => (
                <Space>
                  <Button icon={<EditOutlined />} onClick={() => openEdit(item)}>
                    {t('enterprise_admin_llm_edit')}
                  </Button>
                  <Popconfirm
                    title={t('enterprise_admin_llm_delete_confirm')}
                    onConfirm={() => handleDelete(item.id)}
                    okText={t('enterprise_admin_llm_delete')}
                    cancelText={t('agent_cancel')}
                  >
                    <Button danger icon={<DeleteOutlined />}>
                      {t('enterprise_admin_llm_delete')}
                    </Button>
                  </Popconfirm>
                </Space>
              ),
            },
          ]}
        />
      </Card>

      <Modal
        title={editingModel ? t('enterprise_admin_llm_edit') : t('enterprise_admin_llm_create')}
        open={modalOpen}
        onOk={handleSave}
        onCancel={() => setModalOpen(false)}
        confirmLoading={saving}
        okText={editingModel ? t('agent_save') : t('enterprise_admin_llm_create')}
        destroyOnHidden
      >
        <Form
          form={form}
          layout="vertical"
          initialValues={{ type: 'openai', apiBase: 'https://api.openai.com/v1', displayName: '', apiKey: '', model: '' }}
        >
          <Form.Item
            label={t('enterprise_admin_llm_name')}
            name="displayName"
            rules={[{ required: true, message: t('enterprise_admin_llm_name_required') }]}
          >
            <Input placeholder={t('enterprise_admin_llm_name_placeholder')} />
          </Form.Item>
          <Form.Item
            label={t('enterprise_admin_llm_provider')}
            name="type"
            rules={[{ required: true, message: t('enterprise_admin_llm_provider_required') }]}
          >
            <Select
              options={[
                { label: 'OpenAI', value: 'openai' },
                { label: 'Anthropic', value: 'anthropic' },
              ]}
            />
          </Form.Item>
          <Form.Item label={t('enterprise_admin_llm_api_base')} name="apiBase">
            <Input placeholder={providerType === 'anthropic' ? 'https://api.anthropic.com/v1' : 'https://api.openai.com/v1'} />
          </Form.Item>
          <Form.Item label={t('enterprise_admin_llm_api_key')} name="apiKey">
            <Input.Password placeholder={providerType === 'anthropic' ? 'sk-ant-...' : 'sk-...'} />
          </Form.Item>
          <Form.Item
            label={t('enterprise_admin_llm_model')}
            name="model"
            rules={[{ required: true, message: t('enterprise_admin_llm_model_required') }]}
          >
            <Input placeholder={providerType === 'anthropic' ? 'claude-sonnet-4-20250514' : 'gpt-4o'} />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
};

export default EnterpriseLLMSettingsTab;
