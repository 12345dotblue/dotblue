import React, { useEffect, useState } from 'react';
import { Button, Card, Checkbox, Form, Input, Modal, Popconfirm, Select, Space, Table, Tag, Typography, message } from 'antd';
import { CloudServerOutlined, DeleteOutlined, EditOutlined, PlusOutlined, StarOutlined } from '@ant-design/icons';
import axios from 'axios';
import { useTranslation } from 'react-i18next';
import { BACKEND_URL } from '../../config';
import { casdoorService } from '../identity/CasdoorService';

const { Paragraph, Text } = Typography;

interface PlatformLLMModel {
  id: string;
  displayName: string;
  type: 'openai' | 'anthropic';
  apiBase: string;
  apiKey: string;
  model: string;
  isDefault: boolean;
}

function getAuthHeaders() {
  const token = casdoorService.getToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

const PlatformLLMSettingsCard: React.FC = () => {
  const { t } = useTranslation();
  const [messageApi, contextHolder] = message.useMessage();
  const [models, setModels] = useState<PlatformLLMModel[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingModel, setEditingModel] = useState<PlatformLLMModel | null>(null);
  const [form] = Form.useForm<PlatformLLMModel>();
  const providerType = Form.useWatch('type', form);

  const loadModels = async () => {
    setLoading(true);
    try {
      const res = await axios.get(`${BACKEND_URL}/api/admin/platform/llm-models`, { headers: getAuthHeaders() });
      setModels(Array.isArray(res.data) ? res.data : []);
    } catch {
      messageApi.error(t('platform_admin_llm_load_failed'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadModels();
  }, []);

  const openCreate = () => {
    setEditingModel(null);
    form.setFieldsValue({
      displayName: '',
      type: 'openai',
      apiBase: 'https://api.openai.com/v1',
      apiKey: '',
      model: '',
      isDefault: models.length === 0,
    });
    setModalOpen(true);
  };

  const openEdit = (item: PlatformLLMModel) => {
    setEditingModel(item);
    form.setFieldsValue(item);
    setModalOpen(true);
  };

  const handleSave = async () => {
    const values = await form.validateFields();
    setSaving(true);
    try {
      if (editingModel) {
        await axios.put(`${BACKEND_URL}/api/admin/platform/llm-models/${editingModel.id}`, values, { headers: getAuthHeaders() });
        messageApi.success(t('platform_admin_llm_updated'));
      } else {
        await axios.post(`${BACKEND_URL}/api/admin/platform/llm-models`, values, { headers: getAuthHeaders() });
        messageApi.success(t('platform_admin_llm_created'));
      }
      setModalOpen(false);
      await loadModels();
    } catch (error: any) {
      const errorText = error?.response?.data;
      messageApi.error(typeof errorText === 'string' ? errorText : t('platform_admin_llm_save_failed'));
    } finally {
      setSaving(false);
    }
  };

  const handleSetDefault = async (item: PlatformLLMModel) => {
    try {
      await axios.put(`${BACKEND_URL}/api/admin/platform/llm-models/${item.id}`, { ...item, isDefault: true }, { headers: getAuthHeaders() });
      messageApi.success(t('platform_admin_llm_default_set'));
      await loadModels();
    } catch {
      messageApi.error(t('platform_admin_llm_default_failed'));
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await axios.delete(`${BACKEND_URL}/api/admin/platform/llm-models/${id}`, { headers: getAuthHeaders() });
      messageApi.success(t('platform_admin_llm_deleted'));
      await loadModels();
    } catch {
      messageApi.error(t('platform_admin_llm_delete_failed'));
    }
  };

  return (
    <Card
      variant="borderless"
      title={<span><CloudServerOutlined style={{ color: '#52c41a', marginRight: 8 }} />{t('platform_admin_llm_title')}</span>}
      extra={<Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>{t('platform_admin_llm_create')}</Button>}
      style={{ borderRadius: 16 }}
      loading={loading}
    >
      {contextHolder}
      <Paragraph type="secondary">
        {t('platform_admin_llm_desc')}
      </Paragraph>
      <Table
        rowKey="id"
        pagination={false}
        dataSource={models}
        scroll={{ x: 920 }}
        columns={[
          {
            title: t('platform_admin_llm_name'),
            dataIndex: 'displayName',
            key: 'displayName',
            render: (value: string, item: PlatformLLMModel) => (
              <Space direction="vertical" size={0}>
                <Space size={8}>
                  <Text strong>{value}</Text>
                  {item.isDefault ? <Tag color="gold">{t('platform_admin_llm_default')}</Tag> : null}
                </Space>
                <Text type="secondary">{item.model}</Text>
              </Space>
            ),
          },
          { title: t('platform_admin_llm_provider'), dataIndex: 'type', key: 'type', width: 140 },
          { title: t('platform_admin_llm_api_base'), dataIndex: 'apiBase', key: 'apiBase', ellipsis: true },
          {
            title: t('platform_admin_llm_actions'),
            key: 'actions',
            width: 260,
            render: (_: unknown, item: PlatformLLMModel) => (
              <Space>
                <Button icon={<EditOutlined />} onClick={() => openEdit(item)}>
                  {t('platform_admin_llm_edit')}
                </Button>
                <Button
                  icon={<StarOutlined />}
                  disabled={item.isDefault}
                  onClick={() => handleSetDefault(item)}
                >
                  {t('platform_admin_llm_set_default')}
                </Button>
                <Popconfirm
                  title={t('platform_admin_llm_delete_confirm')}
                  onConfirm={() => handleDelete(item.id)}
                  okText={t('platform_admin_llm_delete')}
                  cancelText={t('agent_cancel')}
                >
                  <Button danger icon={<DeleteOutlined />}>
                    {t('platform_admin_llm_delete')}
                  </Button>
                </Popconfirm>
              </Space>
            ),
          },
        ]}
      />

      <Modal
        title={editingModel ? t('platform_admin_llm_edit') : t('platform_admin_llm_create')}
        open={modalOpen}
        onOk={handleSave}
        onCancel={() => setModalOpen(false)}
        confirmLoading={saving}
        okText={editingModel ? t('agent_save') : t('platform_admin_llm_create')}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          initialValues={{ type: 'openai', apiBase: 'https://api.openai.com/v1', displayName: '', apiKey: '', model: '', isDefault: models.length === 0 }}
        >
          <Form.Item
            label={t('platform_admin_llm_name')}
            name="displayName"
            rules={[{ required: true, message: t('platform_admin_llm_name_required') }]}
          >
            <Input placeholder={t('platform_admin_llm_name_placeholder')} />
          </Form.Item>
          <Form.Item
            label={t('platform_admin_llm_provider')}
            name="type"
            rules={[{ required: true, message: t('platform_admin_llm_provider_required') }]}
          >
            <Select
              options={[
                { label: 'OpenAI', value: 'openai' },
                { label: 'Anthropic', value: 'anthropic' },
              ]}
            />
          </Form.Item>
          <Form.Item label={t('platform_admin_llm_api_base')} name="apiBase">
            <Input placeholder={providerType === 'anthropic' ? 'https://api.anthropic.com/v1' : 'https://api.openai.com/v1'} />
          </Form.Item>
          <Form.Item label={t('platform_admin_llm_api_key')} name="apiKey">
            <Input.Password placeholder={providerType === 'anthropic' ? 'sk-ant-...' : 'sk-...'} />
          </Form.Item>
          <Form.Item
            label={t('platform_admin_llm_model')}
            name="model"
            rules={[{ required: true, message: t('platform_admin_llm_model_required') }]}
          >
            <Input placeholder={providerType === 'anthropic' ? 'claude-sonnet-4-20250514' : 'gpt-4o'} />
          </Form.Item>
          <Form.Item name="isDefault" valuePropName="checked">
            <Checkbox>{t('platform_admin_llm_default_checkbox')}</Checkbox>
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
};

export default PlatformLLMSettingsCard;

