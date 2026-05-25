import React, { useEffect, useState } from 'react';
import { Button, Card, Col, Form, Input, InputNumber, Row, Select, Typography, message } from 'antd';
import { DatabaseOutlined, SettingOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import axios from 'axios';
import { BACKEND_URL } from '../../config';
import { casdoorService } from '../identity/CasdoorService';
import PlatformLLMSettingsCard from './PlatformLLMSettingsCard';
import PlatformUsageSettingsCard from './PlatformUsageSettingsCard';

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

interface AdminSettingsData {
  platform: PlatformConfig | null;
}

function getAuthHeaders() {
  const token = casdoorService.getToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

const PlatformSettingsPage: React.FC = () => {
  const { t } = useTranslation();
  const [messageApi, contextHolder] = message.useMessage();
  const [fetching, setFetching] = useState(true);
  const [savingPlatform, setSavingPlatform] = useState(false);
  const [platformForm] = Form.useForm<PlatformConfig>();

  useEffect(() => {
    axios.get(`${BACKEND_URL}/api/admin/settings`, {
      headers: getAuthHeaders(),
    }).then((res) => {
      const data: AdminSettingsData = res.data;
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
  }, [messageApi, platformForm, t]);

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
          <PlatformLLMSettingsCard />
        </Col>
        <Col xs={24}>
          <PlatformUsageSettingsCard />
        </Col>
      </Row>
    </div>
  );
};

export default PlatformSettingsPage;
