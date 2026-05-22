import React, { useEffect, useState } from 'react';
import { Button, Card, Col, Divider, Form, Input, InputNumber, Row, Select, Typography, message } from 'antd';
import { CloudServerOutlined, DatabaseOutlined, SettingOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import axios from 'axios';
import { BACKEND_URL } from '../../config';
import { casdoorService } from '../identity/CasdoorService';

const { Paragraph, Title } = Typography;

interface PlatformConfig {
  dataBasePath: string;
  dataMountPath: string;
  containerPort: number;
  runtimeMode: 'auto' | 'host' | 'container';
  endpointMode: 'auto' | 'host_loopback' | 'docker_dns';
  dockerEndpoint: string;
  dockerNetwork: string;
}

interface ProviderConfig {
  type: 'openai' | 'anthropic';
  apiBase: string;
  apiKey: string;
  model: string;
}

interface AdminSettingsData {
  platform: PlatformConfig | null;
  provider: ProviderConfig | null;
}

function getAuthHeaders() {
  const token = casdoorService.getToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

const PlatformSettingsPage: React.FC = () => {
  const { t } = useTranslation();
  const [messageApi, contextHolder] = message.useMessage();
  const [fetching, setFetching] = useState(true);
  const [savingProvider, setSavingProvider] = useState(false);
  const [savingPlatform, setSavingPlatform] = useState(false);
  const [providerForm] = Form.useForm<ProviderConfig>();
  const [platformForm] = Form.useForm<PlatformConfig>();
  const providerType = Form.useWatch('type', providerForm);

  useEffect(() => {
    axios.get(`${BACKEND_URL}/api/admin/settings`, {
      headers: getAuthHeaders(),
    }).then((res) => {
      const data: AdminSettingsData = res.data;
      providerForm.setFieldsValue({
        type: data.provider?.type || 'openai',
        apiBase: data.provider?.apiBase || '',
        apiKey: data.provider?.apiKey || '',
        model: data.provider?.model || 'gpt-4o',
      });
      platformForm.setFieldsValue({
        dataBasePath: data.platform?.dataBasePath || '',
        dataMountPath: data.platform?.dataMountPath || '',
        containerPort: data.platform?.containerPort || 8642,
        runtimeMode: data.platform?.runtimeMode || 'auto',
        endpointMode: data.platform?.endpointMode || 'auto',
        dockerEndpoint: data.platform?.dockerEndpoint || '',
        dockerNetwork: data.platform?.dockerNetwork || '',
      });
    }).catch(() => {
      messageApi.error(t('platform_settings_load_failed'));
    }).finally(() => {
      setFetching(false);
    });
  }, [messageApi, platformForm, providerForm, t]);

  const handleProviderSave = async (values: ProviderConfig) => {
    setSavingProvider(true);
    try {
      await axios.post(`${BACKEND_URL}/api/admin/settings`, { provider: values }, {
        headers: getAuthHeaders(),
      });
      messageApi.success(t('platform_settings_provider_saved'));
    } catch {
      messageApi.error(t('platform_settings_provider_save_failed'));
    } finally {
      setSavingProvider(false);
    }
  };

  const handlePlatformSave = async (values: PlatformConfig) => {
    setSavingPlatform(true);
    try {
      await axios.post(`${BACKEND_URL}/api/admin/settings`, { platform: values }, {
        headers: getAuthHeaders(),
      });
      messageApi.success(t('platform_settings_runtime_saved'));
    } catch {
      messageApi.error(t('platform_settings_runtime_save_failed'));
    } finally {
      setSavingPlatform(false);
    }
  };

  return (
    <div style={{ animation: 'fadeIn 0.5s ease-out' }}>
      {contextHolder}
      <div style={{ marginBottom: 24 }}>
        <Title level={3} style={{ marginBottom: 8 }}>
          <SettingOutlined style={{ marginRight: 8 }} />
          {t('platform_settings_title')}
        </Title>
        <Paragraph type="secondary" style={{ marginBottom: 0, maxWidth: 760 }}>
          {t('platform_settings_desc')}
        </Paragraph>
      </div>

      <Row gutter={[16, 16]}>
        <Col xs={24} xl={10}>
          <Card
            variant="borderless"
            title={<span><DatabaseOutlined style={{ color: '#1890ff', marginRight: 8 }} />{t('platform_settings_runtime_title')}</span>}
            style={{ borderRadius: 16 }}
            loading={fetching}
          >
            <Paragraph type="secondary">
              {t('platform_settings_runtime_desc')}
            </Paragraph>
            <Form form={platformForm} layout="vertical" onFinish={handlePlatformSave}>
              <Form.Item
                label={t('platform_settings_database_path')}
                name="dataBasePath"
                rules={[{ required: true, message: t('platform_settings_database_path_required') }]}
              >
                <Input placeholder="/var/lib/dotblue/agents" />
              </Form.Item>
              <Form.Item
                label={t('platform_settings_data_mount_path')}
                name="dataMountPath"
              >
                <Input placeholder="/runtime-data" />
              </Form.Item>
              <Form.Item
                label={t('platform_settings_container_port')}
                name="containerPort"
                rules={[{ required: true, message: t('platform_settings_container_port_required') }]}
              >
                <InputNumber style={{ width: '100%' }} placeholder="8642" />
              </Form.Item>
              <Form.Item
                label={t('platform_settings_runtime_mode')}
                name="runtimeMode"
                rules={[{ required: true, message: t('platform_settings_runtime_mode_required') }]}
              >
                <Select
                  options={[
                    { value: 'auto', label: t('platform_settings_runtime_mode_auto') },
                    { value: 'host', label: t('platform_settings_runtime_mode_host') },
                    { value: 'container', label: t('platform_settings_runtime_mode_container') },
                  ]}
                />
              </Form.Item>
              <Form.Item
                label={t('platform_settings_endpoint_mode')}
                name="endpointMode"
                rules={[{ required: true, message: t('platform_settings_endpoint_mode_required') }]}
              >
                <Select
                  options={[
                    { value: 'auto', label: t('platform_settings_endpoint_mode_auto') },
                    { value: 'host_loopback', label: t('platform_settings_endpoint_mode_host_loopback') },
                    { value: 'docker_dns', label: t('platform_settings_endpoint_mode_docker_dns') },
                  ]}
                />
              </Form.Item>
              <Form.Item
                label={t('platform_settings_docker_endpoint')}
                name="dockerEndpoint"
              >
                <Input placeholder="unix:///var/run/docker.sock" />
              </Form.Item>
              <Form.Item
                label={t('platform_settings_docker_network')}
                name="dockerNetwork"
              >
                <Input placeholder="dotblue_default" />
              </Form.Item>
              <Button type="primary" htmlType="submit" loading={savingPlatform}>
                {t('platform_settings_save_runtime')}
              </Button>
            </Form>
          </Card>
        </Col>

        <Col xs={24} xl={14}>
          <Card
            variant="borderless"
            title={<span><CloudServerOutlined style={{ color: '#52c41a', marginRight: 8 }} />{t('platform_settings_provider_title')}</span>}
            style={{ borderRadius: 16 }}
            loading={fetching}
          >
            <Paragraph type="secondary">
              {t('platform_settings_provider_desc')}
            </Paragraph>
            <Form
              form={providerForm}
              layout="vertical"
              onFinish={handleProviderSave}
              initialValues={{ type: 'openai', model: 'gpt-4o', apiBase: 'https://api.openai.com/v1' }}
            >
              <Form.Item label={t('platform_settings_provider_type')} name="type" hidden rules={[{ required: true }]}>
                <Input />
              </Form.Item>
              <Form.Item name="type" noStyle>
                <div style={{ marginBottom: 24 }}>
                  <Button
                    type={providerType === 'openai' ? 'primary' : 'default'}
                    onClick={() => providerForm.setFieldValue('type', 'openai')}
                    style={{ marginRight: 8 }}
                  >
                    OpenAI
                  </Button>
                  <Button
                    type={providerType === 'anthropic' ? 'primary' : 'default'}
                    onClick={() => providerForm.setFieldValue('type', 'anthropic')}
                  >
                    Anthropic
                  </Button>
                </div>
              </Form.Item>
              <Form.Item label={t('platform_settings_api_base')} name="apiBase">
                <Input placeholder={providerType === 'anthropic' ? 'https://api.anthropic.com/v1' : 'https://api.openai.com/v1'} />
              </Form.Item>
              <Form.Item
                label={t('platform_settings_api_key')}
                name="apiKey"
                rules={[{ required: true, message: t('platform_settings_api_key_required') }]}
              >
                <Input.Password placeholder={providerType === 'anthropic' ? 'sk-ant-...' : 'sk-...'} />
              </Form.Item>
              <Form.Item label={t('platform_settings_model')} name="model">
                <Input placeholder={providerType === 'anthropic' ? 'claude-sonnet-4-20250514' : 'gpt-4o'} />
              </Form.Item>
              <Divider style={{ margin: '8px 0 16px' }} />
              <Button type="primary" htmlType="submit" loading={savingProvider}>
                {t('platform_settings_save_provider')}
              </Button>
            </Form>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default PlatformSettingsPage;
