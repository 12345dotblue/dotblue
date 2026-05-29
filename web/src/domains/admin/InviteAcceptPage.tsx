import React, { useEffect, useMemo, useState } from 'react';
import { Button, Card, Result, Space, Spin, Typography, message } from 'antd';
import { CheckCircleOutlined, MailOutlined, TeamOutlined } from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import axios from 'axios';
import { BACKEND_URL } from '../../config';
import { casdoorService } from '../identity/CasdoorService';
import { getLocalizedPath, resolveSupportedLanguage } from '../../i18n/config';

const { Paragraph, Title } = Typography;

const PENDING_INVITE_CODE_KEY = 'pending_invite_code';

function getAuthHeaders() {
  const token = casdoorService.getToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

const InviteAcceptPage: React.FC = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const { code = '' } = useParams();
  const currentLanguage = resolveSupportedLanguage(i18n.resolvedLanguage || i18n.language);
  const [messageApi, contextHolder] = message.useMessage();
  const [loading, setLoading] = useState(false);
  const [acceptedEnterpriseName, setAcceptedEnterpriseName] = useState('');

  const inviteCode = useMemo(() => code.trim(), [code]);
  const authenticated = casdoorService.isAuthenticated();

  useEffect(() => {
    if (!inviteCode || !authenticated) {
      return;
    }

    setLoading(true);
    axios.post(`${BACKEND_URL}/api/invitations/${inviteCode}/accept`, {}, {
      headers: getAuthHeaders(),
    }).then((res) => {
      const enterpriseName = res.data?.enterprise?.name || '';
      setAcceptedEnterpriseName(enterpriseName);
      messageApi.success(t('invite_accept_success'));
    }).catch((error) => {
      const errorText = error?.response?.data || t('invite_accept_failed');
      messageApi.error(typeof errorText === 'string' ? errorText : t('invite_accept_failed'));
    }).finally(() => {
      setLoading(false);
    });
  }, [authenticated, inviteCode, messageApi, t]);

  const rememberInviteAndGo = (target: 'signin' | 'signup') => {
    if (inviteCode) {
      localStorage.setItem(PENDING_INVITE_CODE_KEY, inviteCode);
    }
    window.location.href = target === 'signup'
      ? casdoorService.getSignUpUrl()
      : casdoorService.getSigninUrl();
  };

  if (!inviteCode) {
    return (
      <div style={{ minHeight: '100vh', display: 'grid', placeItems: 'center', padding: 24 }}>
        {contextHolder}
        <Result
          status="404"
          title={t('invite_not_found_title')}
          subTitle={t('invite_not_found_desc')}
          extra={<Button type="primary" onClick={() => navigate(getLocalizedPath('/', currentLanguage))}>{t('invite_back_home')}</Button>}
        />
      </div>
    );
  }

  if (!authenticated) {
    return (
      <div style={{ minHeight: '100vh', display: 'grid', placeItems: 'center', padding: 24, background: '#f4f7f9' }}>
        {contextHolder}
        <Card variant="borderless" style={{ width: '100%', maxWidth: 560, borderRadius: 20 }}>
          <Space orientation="vertical" size={16} style={{ width: '100%' }}>
            <Space align="start">
              <TeamOutlined style={{ fontSize: 28, color: '#1677ff', marginTop: 6 }} />
              <div>
                <Title level={3} style={{ marginBottom: 8 }}>{t('invite_join_title')}</Title>
                <Paragraph type="secondary" style={{ marginBottom: 0 }}>
                  {t('invite_join_desc')}
                </Paragraph>
              </div>
            </Space>

            <Card size="small" style={{ borderRadius: 14, background: '#fafcff' }}>
              <Space orientation="vertical" size={4}>
                <Paragraph type="secondary" style={{ marginBottom: 0 }}>{t('invite_code_label')}</Paragraph>
                <Paragraph copyable={{ text: inviteCode }} style={{ marginBottom: 0, fontFamily: 'monospace' }}>
                  {inviteCode}
                </Paragraph>
              </Space>
            </Card>

            <Space wrap>
              <Button type="primary" size="large" onClick={() => rememberInviteAndGo('signin')}>
                {t('invite_login_to_join')}
              </Button>
              <Button size="large" icon={<MailOutlined />} onClick={() => rememberInviteAndGo('signup')}>
                {t('invite_create_account')}
              </Button>
            </Space>
          </Space>
        </Card>
      </div>
    );
  }

  if (loading) {
    return (
      <div style={{ minHeight: '100vh', display: 'grid', placeItems: 'center', background: '#f4f7f9' }}>
        {contextHolder}
        <Space orientation="vertical" size={16} align="center">
          <Spin size="large" />
          <Paragraph type="secondary" style={{ marginBottom: 0 }}>
            {t('invite_accepting')}
          </Paragraph>
        </Space>
      </div>
    );
  }

  if (acceptedEnterpriseName) {
    return (
      <div style={{ minHeight: '100vh', display: 'grid', placeItems: 'center', padding: 24, background: '#f4f7f9' }}>
        {contextHolder}
        <Result
          status="success"
          icon={<CheckCircleOutlined />}
          title={t('invite_accepted_title')}
          subTitle={t('invite_accepted_desc', { enterpriseName: acceptedEnterpriseName })}
          extra={[
            <Button type="primary" key="admin" onClick={() => navigate(getLocalizedPath('/admin/enterprise', currentLanguage))}>
              {t('invite_open_admin_console')}
            </Button>,
            <Button key="chat" onClick={() => navigate(getLocalizedPath('/chat', currentLanguage))}>
              {t('invite_go_chat')}
            </Button>,
          ]}
        />
      </div>
    );
  }

  return (
    <div style={{ minHeight: '100vh', display: 'grid', placeItems: 'center', padding: 24, background: '#f4f7f9' }}>
      {contextHolder}
      <Result
        status="warning"
        title={t('invite_accept_failed_title')}
        subTitle={t('invite_accept_failed_desc')}
        extra={[
          <Button type="primary" key="retry" onClick={() => window.location.reload()}>
            {t('invite_retry')}
          </Button>,
          <Button key="dashboard" onClick={() => navigate(getLocalizedPath('/dashboard', currentLanguage))}>
            {t('invite_back_dashboard')}
          </Button>,
        ]}
      />
    </div>
  );
};

export default InviteAcceptPage;
