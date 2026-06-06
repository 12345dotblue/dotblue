import React, { useState, useEffect } from 'react';
import { Form, Input, Button, Card, Typography, Result, message } from 'antd';
import { LockOutlined, UserOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import axios from 'axios';
import { BACKEND_URL } from '../../config';
import { getLocalizedPath, resolveSupportedLanguage } from '../../i18n/config';

const { Title, Text } = Typography;
const { Password } = Input;

interface AdminInfo {
  adminUsername: string;
  adminPassword: string;
  adminEmail: string;
}

const SetupWizard: React.FC = () => {
  const { t, i18n } = useTranslation();
  const [initialized, setInitialized] = useState<boolean | null>(null);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const currentLanguage = resolveSupportedLanguage(i18n?.resolvedLanguage || i18n?.language);

  const [form] = Form.useForm<AdminInfo>();

  useEffect(() => {
    axios.get(`${BACKEND_URL}/api/setup/status`)
      .then(res => {
        setInitialized(res.data.initialized === true);
      })
      .catch(() => {
        setInitialized(true);
      });
  }, []);

  const onFinish = async (values: AdminInfo) => {
    setLoading(true);
    try {
      await axios.post(`${BACKEND_URL}/api/setup/install`, values);
      message.success(t('setup_install_success'));
      navigate(getLocalizedPath('/login', currentLanguage));
    } catch (err: unknown) {
      if (axios.isAxiosError(err) && err.response?.status === 409) {
        message.error(t('setup_user_exists'));
      } else {
        const msg = err instanceof Error ? err.message : '';
        if (msg.includes('403') || msg.includes('already initialized')) {
          setInitialized(true);
        }
        message.error(t('setup_install_failed'));
      }
    } finally {
      setLoading(false);
    }
  };

  if (initialized === null) {
    return null;
  }

  if (initialized) {
    return (
      <div style={{
        minHeight: '100vh',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        padding: 24,
      }}>
        <div style={{
          background: '#fff',
          borderRadius: 20,
          padding: '48px 40px',
          maxWidth: 480,
          width: '100%',
          boxShadow: '0 20px 60px rgba(0,0,0,0.15)',
        }}>
          <Result
            status="warning"
            icon={<LockOutlined style={{ color: '#faad14' }} />}
            title={t('setup_locked_title')}
            subTitle={t('setup_locked_desc')}
            extra={
              <Button type="primary" size="large" onClick={() => navigate(getLocalizedPath('/login', currentLanguage))}>
                {t('setup_go_login')}
              </Button>
            }
          />
        </div>
      </div>
    );
  }

  return (
    <div style={{
      minHeight: '100vh',
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      padding: 24,
    }}>
      <div style={{
        background: '#fff',
        borderRadius: 20,
        padding: '48px 40px',
        maxWidth: 480,
        width: '100%',
        boxShadow: '0 20px 60px rgba(0,0,0,0.15)',
      }}>
        <div style={{ textAlign: 'center', marginBottom: 40 }}>
          <div style={{
            width: 56, height: 56, borderRadius: 14,
            background: 'linear-gradient(135deg, #1677ff 0%, #36cfc9 100%)',
            display: 'inline-block', marginBottom: 16,
            boxShadow: '0 8px 24px rgba(22,119,255,0.3)',
          }} />
          <Title level={3} style={{ margin: 0 }}>{t('setup_welcome')}</Title>
          <Text type="secondary">{t('setup_subtitle')}</Text>
        </div>

        <Card variant="borderless" style={{ borderRadius: 12 }}>
          <Typography.Paragraph type="secondary" style={{ marginBottom: 24 }}>
            {t('setup_admin_desc')}
          </Typography.Paragraph>
          <Form
            form={form}
            layout="vertical"
            onFinish={onFinish}
            initialValues={{ adminUsername: 'admin' }}
          >
            <Form.Item label={t('setup_username')} name="adminUsername" rules={[{ required: true }]}>
              <Input prefix={<UserOutlined />} placeholder={t('setup_username_placeholder')} />
            </Form.Item>
            <Form.Item label={t('setup_password')} name="adminPassword" rules={[{ required: true, min: 6 }]}>
              <Password placeholder={t('setup_password_placeholder')} />
            </Form.Item>
            <Form.Item label={t('setup_email')} name="adminEmail" rules={[{ required: true, type: 'email' }]}>
              <Input placeholder={t('setup_email_placeholder')} />
            </Form.Item>
            <Form.Item style={{ marginBottom: 0 }}>
              <Button type="primary" htmlType="submit" block loading={loading}>
                {t('setup_complete')}
              </Button>
            </Form.Item>
          </Form>
        </Card>

        <div style={{ textAlign: 'center', marginTop: 24 }}>
          <Text type="secondary" style={{ fontSize: 12 }}>
            {t('setup_provider_hint')}
          </Text>
        </div>
      </div>
    </div>
  );
};

export default SetupWizard;
