import React, { useState } from 'react';
import { Typography, Button, Row, Col, Card, Space, Switch, Tag, theme } from 'antd';
import {
  SafetyCertificateOutlined,
  AppstoreOutlined,
  ThunderboltOutlined,
  ApiOutlined,
  AuditOutlined,
  CodeOutlined,
  ArrowRightOutlined,
  CheckOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { useAuthState } from '../identity/useAuthState';

const { Title, Paragraph, Text } = Typography;

type StyledIconElement = React.ReactElement<{ style?: React.CSSProperties }>;

const LandingPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { token } = theme.useToken();
  const [yearly, setYearly] = useState(false);
  const isAuthenticated = useAuthState();
  const primaryTarget = isAuthenticated ? '/dashboard' : '/login';

  const features = [
    { title: t('feat_security_title'), desc: t('feat_security_desc'), icon: <SafetyCertificateOutlined />, color: '#52c41a' },
    { title: t('feat_multi_title'), desc: t('feat_multi_desc'), icon: <AppstoreOutlined />, color: '#1677ff' },
    { title: t('feat_stream_title'), desc: t('feat_stream_desc'), icon: <ThunderboltOutlined />, color: '#faad14' },
    { title: t('feat_config_title'), desc: t('feat_config_desc'), icon: <ApiOutlined />, color: '#722ed1' },
    { title: t('feat_enterprise_title'), desc: t('feat_enterprise_desc'), icon: <AuditOutlined />, color: '#eb2f96' },
    { title: t('feat_api_title'), desc: t('feat_api_desc'), icon: <CodeOutlined />, color: '#13c2c2' },
  ];

  const plans = [
    {
      name: t('pricing_starter'),
      price: yearly ? t('pricing_starter_price_y') : t('pricing_starter_price_m'),
      desc: t('pricing_starter_desc'),
      cta: t('pricing_starter_cta'),
      popular: false,
      features: [t('pricing_starter_f1'), t('pricing_starter_f2'), t('pricing_starter_f3'), t('pricing_starter_f4')],
      onClick: () => navigate(primaryTarget),
    },
    {
      name: t('pricing_pro'),
      price: yearly ? t('pricing_pro_price_y') : t('pricing_pro_price_m'),
      desc: t('pricing_pro_desc'),
      cta: t('pricing_pro_cta'),
      popular: true,
      features: [t('pricing_pro_f1'), t('pricing_pro_f2'), t('pricing_pro_f3'), t('pricing_pro_f4'), t('pricing_pro_f5'), t('pricing_pro_f6')],
      onClick: () => navigate(primaryTarget),
    },
    {
      name: t('pricing_enterprise'),
      price: t('pricing_enterprise_price'),
      desc: t('pricing_enterprise_desc'),
      cta: t('pricing_enterprise_cta'),
      popular: false,
      features: [t('pricing_enterprise_f1'), t('pricing_enterprise_f2'), t('pricing_enterprise_f3'), t('pricing_enterprise_f4'), t('pricing_enterprise_f5'), t('pricing_enterprise_f6'), t('pricing_enterprise_f7')],
      onClick: () => window.open('mailto:sales@dotblue.ai'),
    },
  ];

  return (
    <div style={{ width: '100%' }}>
      {/* Hero */}
      <section style={{
        padding: '120px 5% 80px',
        background: 'radial-gradient(ellipse at 70% 20%, #e6f4ff 0%, #fff 60%)',
        textAlign: 'center',
      }}>
        <div style={{ maxWidth: 900, margin: '0 auto' }}>
          <Text strong style={{ color: token.colorPrimary, textTransform: 'uppercase', letterSpacing: 3, fontSize: 13 }}>
            {t('hero_tagline')}
          </Text>
          <Title style={{ fontSize: 'clamp(32px, 5vw, 56px)', marginTop: 16, marginBottom: 24, lineHeight: 1.15, whiteSpace: 'pre-line' }}>
            {t('hero_title')}
          </Title>
          <Paragraph style={{ fontSize: 'clamp(16px, 2vw, 20px)', color: '#666', marginBottom: 40, maxWidth: 680, margin: '0 auto 40px' }}>
            {t('hero_subtitle')}
          </Paragraph>
          <Space size="middle" wrap>
            <Button type="primary" size="large" shape="round" icon={<ArrowRightOutlined />} onClick={() => navigate(primaryTarget)} style={{ height: 52, padding: '0 36px', fontSize: 16 }}>
              {isAuthenticated ? 'Dashboard' : t('hero_cta_primary')}
            </Button>
            <Button size="large" shape="round" onClick={() => {
              document.getElementById('pricing')?.scrollIntoView({ behavior: 'smooth' });
            }} style={{ height: 52, padding: '0 36px', fontSize: 16 }}>
              {t('hero_cta_secondary')}
            </Button>
          </Space>

          {/* Stats bar */}
          <Row justify="center" style={{ marginTop: 64 }} gutter={[48, 24]}>
            {[
              { label: t('hero_stat_security'), value: 'gVisor' },
              { label: t('hero_stat_sandbox'), value: 'Per-User' },
              { label: t('hero_stat_uptime'), value: '99.9%' },
            ].map((s, i) => (
              <Col key={i} style={{ textAlign: 'center' }}>
                <div style={{ fontSize: 28, fontWeight: 700, color: token.colorPrimary }}>{s.value}</div>
                <Text type="secondary" style={{ fontSize: 13 }}>{s.label}</Text>
              </Col>
            ))}
          </Row>
        </div>
      </section>

      {/* Features */}
      <section style={{ padding: '100px 5%', background: '#fafbfc' }}>
        <div style={{ textAlign: 'center', marginBottom: 64, maxWidth: 640, margin: '0 auto 64px' }}>
          <Title level={2}>{t('features_title')}</Title>
          <Paragraph type="secondary" style={{ fontSize: 16 }}>{t('features_subtitle')}</Paragraph>
        </div>
        <Row gutter={[24, 24]} justify="center" style={{ maxWidth: 1100, margin: '0 auto' }}>
          {features.map((f, i) => (
            <Col xs={24} sm={12} lg={8} key={i}>
              <Card hoverable bordered={false} style={{ height: '100%', borderRadius: 16, padding: '8px 0' }}>
                <div style={{
                  width: 48, height: 48, borderRadius: 12,
                  background: `${f.color}15`, display: 'flex', alignItems: 'center', justifyContent: 'center',
                  marginBottom: 20,
                }}>
                  {React.cloneElement(f.icon as StyledIconElement, { style: { fontSize: 24, color: f.color } })}
                </div>
                <Title level={4} style={{ marginBottom: 8 }}>{f.title}</Title>
                <Paragraph type="secondary" style={{ marginBottom: 0 }}>{f.desc}</Paragraph>
              </Card>
            </Col>
          ))}
        </Row>
      </section>

      {/* Pricing */}
      <section id="pricing" style={{ padding: '100px 5%', background: '#fff' }}>
        <div style={{ textAlign: 'center', marginBottom: 48, maxWidth: 640, margin: '0 auto 48px' }}>
          <Title level={2}>{t('pricing_title')}</Title>
          <Paragraph type="secondary" style={{ fontSize: 16 }}>{t('pricing_subtitle')}</Paragraph>
          <div style={{ marginTop: 24, display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 12 }}>
            <Text type={yearly ? 'secondary' : undefined} strong={!yearly}>{t('pricing_monthly')}</Text>
            <Switch checked={yearly} onChange={setYearly} />
            <Text type={yearly ? undefined : 'secondary'} strong={yearly}>{t('pricing_yearly')}</Text>
            <Tag color="green">{t('pricing_yearly_discount')}</Tag>
          </div>
        </div>
        <Row gutter={[24, 24]} justify="center" style={{ maxWidth: 1100, margin: '0 auto' }}>
          {plans.map((plan, i) => (
            <Col xs={24} sm={12} lg={8} key={i}>
              <Card
                bordered={plan.popular}
                style={{
                  height: '100%', borderRadius: 16, position: 'relative',
                  boxShadow: plan.popular ? `0 8px 30px ${token.colorPrimary}22` : '0 2px 12px rgba(0,0,0,0.04)',
                  borderColor: plan.popular ? token.colorPrimary : undefined,
                }}
              >
                {plan.popular && (
                  <Tag color="blue" style={{ position: 'absolute', top: -12, left: '50%', transform: 'translateX(-50%)', fontSize: 12 }}>
                    {t('pricing_pro_popular')}
                  </Tag>
                )}
                <div style={{ textAlign: 'center', paddingTop: plan.popular ? 8 : 0 }}>
                  <Title level={4} style={{ marginBottom: 4 }}>{plan.name}</Title>
                  <div style={{ fontSize: 40, fontWeight: 700, color: token.colorPrimary, margin: '16px 0 8px' }}>
                    {plan.price}
                    {plan.price !== t('pricing_enterprise_price') && <span style={{ fontSize: 16, fontWeight: 400, color: '#999' }}>/{yearly ? t('pricing_yearly') : t('pricing_monthly')}</span>}
                  </div>
                  <Paragraph type="secondary">{plan.desc}</Paragraph>
                </div>
                <div style={{ margin: '24px 0' }}>
                  {plan.features.map((f, fi) => (
                    <div key={fi} style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 10 }}>
                      <CheckOutlined style={{ color: token.colorPrimary }} />
                      <Text>{f}</Text>
                    </div>
                  ))}
                </div>
                <Button
                  type={plan.popular ? 'primary' : 'default'}
                  block size="large"
                  shape="round"
                  onClick={plan.onClick}
                >
                  {plan.cta}
                </Button>
              </Card>
            </Col>
          ))}
        </Row>
      </section>

      {/* CTA */}
      <section style={{ padding: '80px 5%' }}>
        <div style={{
          background: 'linear-gradient(135deg, #1677ff 0%, #003eb3 100%)',
          borderRadius: 32, padding: '72px 40px', textAlign: 'center', color: '#fff',
          boxShadow: '0 20px 40px rgba(22,119,255,0.2)',
        }}>
          <Title level={2} style={{ color: '#fff', marginBottom: 16 }}>{t('cta_title')}</Title>
          <Paragraph style={{ color: 'rgba(255,255,255,0.85)', fontSize: 18, marginBottom: 40, maxWidth: 600, margin: '0 auto 40px' }}>
            {t('cta_subtitle')}
          </Paragraph>
          <Button size="large" shape="round" ghost onClick={() => navigate(primaryTarget)} style={{ height: 52, padding: '0 48px', fontSize: 16, fontWeight: 600 }}>
            {isAuthenticated ? 'Dashboard' : t('cta_button')}
          </Button>
        </div>
      </section>
    </div>
  );
};

export default LandingPage;
