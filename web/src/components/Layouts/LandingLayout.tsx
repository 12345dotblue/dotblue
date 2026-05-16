import React from 'react';
import { Layout, Button, Space, Typography, Row, Col, Divider, theme } from 'antd';
import { GlobalOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { Link, useNavigate } from 'react-router-dom';

const { Header, Content, Footer } = Layout;
const { Title, Text, Paragraph } = Typography;

const LandingLayout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const { token } = theme.useToken();

  const changeLanguage = (lng: string) => {
    i18n.changeLanguage(lng);
  };

  return (
    <Layout style={{ minHeight: '100vh', background: '#fff' }}>
      <Header style={{
        background: 'rgba(255, 255, 255, 0.85)',
        backdropFilter: 'blur(10px)',
        padding: '0 5%',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        position: 'sticky',
        top: 0,
        zIndex: 100,
        borderBottom: '1px solid #f0f0f0',
        height: 72,
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 32 }}>
          <div style={{ display: 'flex', alignItems: 'center', cursor: 'pointer' }} onClick={() => navigate('/')}>
            <div style={{ width: 32, height: 32, background: token.colorPrimary, borderRadius: 8, marginRight: 12 }} />
            <Title level={4} style={{ margin: 0 }}>dotblue</Title>
          </div>
          <Space size="middle" style={{ display: 'flex' }}>
            <Button type="text" onClick={() => navigate('/#pricing')} style={{ padding: '0 12px' }}>
              {t('view_pricing')}
            </Button>
            <Button type="text" onClick={() => navigate('/terms')} style={{ padding: '0 12px' }}>
              {t('terms')}
            </Button>
            <Button type="text" onClick={() => navigate('/privacy')} style={{ padding: '0 12px' }}>
              {t('privacy')}
            </Button>
          </Space>
        </div>

        <Space size="middle">
          <Button type="text" onClick={() => changeLanguage(i18n.language === 'en' ? 'zh-CN' : 'en')} icon={<GlobalOutlined />}>
            {i18n.language === 'en' ? '中文' : 'English'}
          </Button>
          <Button type="primary" shape="round" onClick={() => navigate('/login')}>
            {t('login')}
          </Button>
        </Space>
      </Header>

      <Content>{children}</Content>

      <Footer style={{ background: '#f9fafb', padding: '64px 5% 32px' }}>
        <Row gutter={[32, 32]}>
          <Col xs={24} md={8}>
            <div style={{ display: 'flex', alignItems: 'center', marginBottom: 16 }}>
              <div style={{ width: 24, height: 24, background: token.colorPrimary, borderRadius: 6, marginRight: 8 }} />
              <Title level={5} style={{ margin: 0 }}>dotblue</Title>
            </div>
            <Paragraph type="secondary">{t('footer_desc')}</Paragraph>
            <Text type="secondary" style={{ fontSize: 12 }}>
              &copy; 2026 Dotblue Tech Co., Ltd.
            </Text>
          </Col>
          <Col xs={12} md={4}>
            <Title level={5}>{t('footer_product')}</Title>
            <Space direction="vertical">
              <Link to="/">{t('welcome')}</Link>
              <Link to="/login">{t('get_started')}</Link>
              <a href="/#pricing">{t('view_pricing')}</a>
            </Space>
          </Col>
          <Col xs={12} md={4}>
            <Title level={5}>{t('footer_legal')}</Title>
            <Space direction="vertical">
              <Link to="/terms">{t('terms')}</Link>
              <Link to="/privacy">{t('privacy')}</Link>
              <Link to="/refund">{t('refund')}</Link>
            </Space>
          </Col>
          <Col xs={24} md={8}>
            <Title level={5}>{t('contact_us')}</Title>
            <Paragraph type="secondary">
              Email: support@dotblue.ai
            </Paragraph>
            <Divider style={{ margin: '16px 0' }} />
            <div style={{ display: 'flex', gap: 12, opacity: 0.5 }}>
              <div style={{ width: 40, height: 24, background: '#e0e0e0', borderRadius: 4, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 10 }}>VISA</div>
              <div style={{ width: 40, height: 24, background: '#e0e0e0', borderRadius: 4, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 10 }}>MC</div>
              <div style={{ width: 40, height: 24, background: '#e0e0e0', borderRadius: 4, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 10 }}>AX</div>
            </div>
          </Col>
        </Row>
      </Footer>
    </Layout>
  );
};

export default LandingLayout;
