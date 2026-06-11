import React, { useEffect, useState } from 'react';
import { Button, Card, Col, Form, Input, InputNumber, Row, Select, Switch, Typography, message } from 'antd';
import { DatabaseOutlined, SettingOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import axios from 'axios';
import { BACKEND_URL } from '../../config';
import { casdoorService } from '../identity/CasdoorService';
import PlatformLLMSettingsCard from './PlatformLLMSettingsCard';
import PlatformCreditSettingsCard from './PlatformCreditSettingsCard';
import PlatformUsageSettingsCard from './PlatformUsageSettingsCard';
import { useThemeMode } from '../../theme/themeMode';

const { Paragraph, Title } = Typography;

const runtimeEngineDefaults: Record<'hermes' | 'nanobot', { enabled: boolean; image: string }> = {
  hermes: { enabled: true, image: 'nousresearch/hermes-agent:latest' },
  nanobot: { enabled: false, image: 'nanobot' },
};

function toRuntimeEngineMap(items?: RuntimeEngineConfig[]): PlatformFormValues['runtimeEngines'] {
  const result: PlatformFormValues['runtimeEngines'] = {
    hermes: { ...runtimeEngineDefaults.hermes },
    nanobot: { ...runtimeEngineDefaults.nanobot },
  };
  (items || []).forEach((item) => {
    if (item.engineType === 'hermes' || item.engineType === 'nanobot') {
      result[item.engineType] = {
        enabled: item.enabled,
        image: item.image || runtimeEngineDefaults[item.engineType].image,
      };
    }
  });
  return result;
}

function toRuntimeEngineList(values: PlatformFormValues['runtimeEngines']): RuntimeEngineConfig[] {
  return (['hermes', 'nanobot'] as const).map((engineType) => ({
    engineType,
    enabled: values?.[engineType]?.enabled ?? runtimeEngineDefaults[engineType].enabled,
    image: values?.[engineType]?.image || runtimeEngineDefaults[engineType].image,
  }));
}

interface PlatformConfig {
  dataBasePath: string;
  dataMountPath: string;
  containerPort: number;
  runtimeMode: 'auto' | 'host' | 'container';
  endpointMode: 'auto' | 'host_loopback' | 'docker_dns';
  dockerEndpoint: string;
  dockerNetwork: string;
  newEnterprisePlatformCredits: number;
  defaultCreditSettlementCurrency: 'USD' | 'CNY';
  runtimeEngines?: RuntimeEngineConfig[];
}

interface RuntimeEngineConfig {
  engineType: 'hermes' | 'nanobot';
  enabled: boolean;
  image: string;
}

interface PlatformFormValues extends Omit<PlatformConfig, 'runtimeEngines'> {
  runtimeEngines: Record<'hermes' | 'nanobot', { enabled: boolean; image: string }>;
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
  const { resolvedTheme } = useThemeMode();
  const [messageApi, contextHolder] = message.useMessage();
  const [fetching, setFetching] = useState(true);
  const [savingPlatform, setSavingPlatform] = useState(false);
  const [platformForm] = Form.useForm<PlatformFormValues>();
  const isDark = resolvedTheme === 'dark';

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
        newEnterprisePlatformCredits: data.platform?.newEnterprisePlatformCredits || 0,
        defaultCreditSettlementCurrency: data.platform?.defaultCreditSettlementCurrency || 'USD',
        runtimeEngines: toRuntimeEngineMap(data.platform?.runtimeEngines),
      });
    }).catch(() => {
      messageApi.error(t('platform_settings_load_failed'));
    }).finally(() => {
      setFetching(false);
    });
  }, [messageApi, platformForm, t]);

  const handlePlatformSave = async (values: PlatformFormValues) => {
    setSavingPlatform(true);
    try {
      const payload: PlatformConfig = {
        dataBasePath: values.dataBasePath,
        dataMountPath: values.dataMountPath,
        containerPort: values.containerPort,
        runtimeMode: values.runtimeMode,
        endpointMode: values.endpointMode,
        dockerEndpoint: values.dockerEndpoint,
        dockerNetwork: values.dockerNetwork,
        newEnterprisePlatformCredits: values.newEnterprisePlatformCredits,
        defaultCreditSettlementCurrency: values.defaultCreditSettlementCurrency,
        runtimeEngines: toRuntimeEngineList(values.runtimeEngines),
      };
      await axios.post(`${BACKEND_URL}/api/admin/settings`, { platform: payload }, {
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
                <Input placeholder={t('platform_settings_database_path_placeholder')} />
              </Form.Item>
              <Form.Item
                label={t('platform_settings_data_mount_path')}
                name="dataMountPath"
              >
                <Input placeholder={t('platform_settings_data_mount_path_placeholder')} />
              </Form.Item>
              <Form.Item
                label={t('platform_settings_container_port')}
                name="containerPort"
                rules={[{ required: true, message: t('platform_settings_container_port_required') }]}
              >
                <InputNumber controls={false} style={{ width: '100%' }} placeholder={t('platform_settings_container_port_placeholder')} />
              </Form.Item>
              <Form.Item
                label={t('platform_settings_new_enterprise_platform_credits')}
                name="newEnterprisePlatformCredits"
              >
                <InputNumber min={0} precision={0} controls={false} style={{ width: '100%' }} />
              </Form.Item>
              <Form.Item
                label={t('platform_settings_default_credit_currency')}
                name="defaultCreditSettlementCurrency"
                rules={[{ required: true, message: t('platform_settings_default_credit_currency_required') }]}
              >
                <Select
                  options={[
                    { value: 'USD', label: 'USD' },
                    { value: 'CNY', label: 'CNY' },
                  ]}
                />
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
                <Input placeholder={t('platform_settings_docker_endpoint_placeholder')} />
              </Form.Item>
              <Form.Item
                label={t('platform_settings_docker_network')}
                name="dockerNetwork"
              >
                <Input placeholder={t('platform_settings_docker_network_placeholder')} />
              </Form.Item>
              <Card
                size="small"
                style={{
                  marginBottom: 16,
                  borderRadius: 16,
                  background: isDark
                    ? 'linear-gradient(180deg, #132033 0%, #101a2b 100%)'
                    : 'linear-gradient(180deg, #fcfdff 0%, #f8fbff 100%)',
                  border: '1px solid var(--app-shell-border)',
                }}
                title={t('platform_settings_runtime_engines_title')}
              >
                <Paragraph type="secondary">
                  {t('platform_settings_runtime_engines_desc')}
                </Paragraph>
                {(['hermes', 'nanobot'] as const).map((engineType) => (
                  <div key={engineType} style={{ marginBottom: engineType === 'nanobot' ? 0 : 16 }}>
                    <Title level={5} style={{ marginBottom: 12 }}>
                      {engineType === 'hermes' ? t('platform_settings_runtime_engine_hermes') : t('platform_settings_runtime_engine_nanobot')}
                    </Title>
                    <Form.Item
                      label={t('platform_settings_runtime_engine_enabled')}
                      name={['runtimeEngines', engineType, 'enabled']}
                      valuePropName="checked"
                    >
                      <Switch />
                    </Form.Item>
                    <Form.Item
                      label={t('platform_settings_runtime_engine_image')}
                      name={['runtimeEngines', engineType, 'image']}
                      rules={[{ required: true, message: t('platform_settings_runtime_engine_image_required') }]}
                    >
                      <Input />
                    </Form.Item>
                  </div>
                ))}
              </Card>
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
          <PlatformCreditSettingsCard />
        </Col>
        <Col xs={24}>
          <PlatformUsageSettingsCard />
        </Col>
      </Row>
    </div>
  );
};

export default PlatformSettingsPage;

