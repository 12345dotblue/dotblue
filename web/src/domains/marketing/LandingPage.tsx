import React, { useMemo, useState } from 'react';
import { Typography, Button, Row, Col, Card, Space, Switch, Tag, theme } from 'antd';
import {
  SafetyCertificateOutlined,
  AppstoreOutlined,
  ThunderboltOutlined,
  ApiOutlined,
  AuditOutlined,
  ArrowRightOutlined,
  CheckOutlined,
  DeploymentUnitOutlined,
  ClusterOutlined,
  RadarChartOutlined,
  CustomerServiceOutlined,
  ShopOutlined,
  ReadOutlined,
  SolutionOutlined,
} from '@ant-design/icons';
import { Helmet } from 'react-helmet-async';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { useAuthState } from '../identity/useAuthState';
import { SUPPORTED_LANGUAGES, buildLocalizedUrl, getLocalizedPath, resolveSupportedLanguage } from '../../i18n/config';
import { useThemeMode } from '../../theme/themeMode';

const { Title, Paragraph, Text } = Typography;

type StyledIconElement = React.ReactElement<{ style?: React.CSSProperties }>;

const LandingPage: React.FC = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const { token } = theme.useToken();
  const { resolvedTheme } = useThemeMode();
  const [yearly, setYearly] = useState(false);
  const isAuthenticated = useAuthState();
  const isDark = resolvedTheme === 'dark';
  const currentLanguage = resolveSupportedLanguage(i18n?.resolvedLanguage || i18n?.language);
  const primaryTarget = getLocalizedPath(isAuthenticated ? '/dashboard' : '/login', currentLanguage);
  const homeTitle = t('home_seo_title');
  const homeDescription = t('home_seo_description');
  const homeKeywords = t('home_seo_keywords');
  const canonicalUrl = buildLocalizedUrl('/', currentLanguage);

  const heroProofs = useMemo(() => ([
    t('hero_proof_security'),
    t('hero_proof_models'),
    t('hero_proof_visibility'),
  ]), [t]);

  const heroPanelItems = useMemo(() => ([
    {
      label: t('hero_panel_item_build_label'),
      title: t('hero_panel_item_build_title'),
      desc: t('hero_panel_item_build_desc'),
      icon: <DeploymentUnitOutlined />,
      color: '#1677ff',
    },
    {
      label: t('hero_panel_item_govern_label'),
      title: t('hero_panel_item_govern_title'),
      desc: t('hero_panel_item_govern_desc'),
      icon: <ClusterOutlined />,
      color: '#13c2c2',
    },
    {
      label: t('hero_panel_item_runtime_label'),
      title: t('hero_panel_item_runtime_title'),
      desc: t('hero_panel_item_runtime_desc'),
      icon: <RadarChartOutlined />,
      color: '#722ed1',
    },
  ]), [t]);

  const valueProps = useMemo(() => ([
    {
      label: t('value_prop_build_label'),
      title: t('value_prop_build_title'),
      desc: t('value_prop_build_desc'),
      metric: t('value_prop_build_metric'),
      icon: <AppstoreOutlined />,
      color: '#1677ff',
    },
    {
      label: t('value_prop_private_label'),
      title: t('value_prop_private_title'),
      desc: t('value_prop_private_desc'),
      metric: t('value_prop_private_metric'),
      icon: <SafetyCertificateOutlined />,
      color: '#52c41a',
    },
    {
      label: t('value_prop_governance_label'),
      title: t('value_prop_governance_title'),
      desc: t('value_prop_governance_desc'),
      metric: t('value_prop_governance_metric'),
      icon: <ClusterOutlined />,
      color: '#722ed1',
    },
    {
      label: t('value_prop_cost_label'),
      title: t('value_prop_cost_title'),
      desc: t('value_prop_cost_desc'),
      metric: t('value_prop_cost_metric'),
      icon: <ThunderboltOutlined />,
      color: '#fa8c16',
    },
  ]), [t]);

  const capabilities = useMemo(() => ([
    { title: t('capability_builder_title'), desc: t('capability_builder_desc'), icon: <DeploymentUnitOutlined />, color: '#1677ff' },
    { title: t('capability_workspace_title'), desc: t('capability_workspace_desc'), icon: <AppstoreOutlined />, color: '#13c2c2' },
    { title: t('capability_agent_governance_title'), desc: t('capability_agent_governance_desc'), icon: <ClusterOutlined />, color: '#722ed1' },
    { title: t('capability_skill_governance_title'), desc: t('capability_skill_governance_desc'), icon: <ApiOutlined />, color: '#eb2f96' },
    { title: t('capability_runtime_title'), desc: t('capability_runtime_desc'), icon: <RadarChartOutlined />, color: '#2f54eb' },
    { title: t('capability_cost_title'), desc: t('capability_cost_desc'), icon: <AuditOutlined />, color: '#faad14' },
  ]), [t]);

  const deploymentModes = useMemo(() => ([
    {
      title: t('deployment_saas_title'),
      desc: t('deployment_saas_desc'),
      points: [t('deployment_saas_point_one'), t('deployment_saas_point_two')],
      icon: <ThunderboltOutlined />,
      color: '#1677ff',
    },
    {
      title: t('deployment_self_host_title'),
      desc: t('deployment_self_host_desc'),
      points: [t('deployment_self_host_point_one'), t('deployment_self_host_point_two')],
      icon: <DeploymentUnitOutlined />,
      color: '#13c2c2',
    },
    {
      title: t('deployment_private_title'),
      desc: t('deployment_private_desc'),
      points: [t('deployment_private_point_one'), t('deployment_private_point_two')],
      icon: <SafetyCertificateOutlined />,
      color: '#722ed1',
    },
  ]), [t]);

  const assistants = useMemo(() => ([
    {
      title: t('assistant_customer_title'),
      desc: t('assistant_customer_desc'),
      tag: t('assistant_customer_tag'),
      icon: <CustomerServiceOutlined />,
      color: '#1677ff',
    },
    {
      title: t('assistant_sales_title'),
      desc: t('assistant_sales_desc'),
      tag: t('assistant_sales_tag'),
      icon: <ShopOutlined />,
      color: '#13c2c2',
    },
    {
      title: t('assistant_knowledge_title'),
      desc: t('assistant_knowledge_desc'),
      tag: t('assistant_knowledge_tag'),
      icon: <ReadOutlined />,
      color: '#722ed1',
    },
    {
      title: t('assistant_ops_title'),
      desc: t('assistant_ops_desc'),
      tag: t('assistant_ops_tag'),
      icon: <SolutionOutlined />,
      color: '#fa8c16',
    },
  ]), [t]);

  const plans = useMemo(() => ([
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
  ]), [navigate, primaryTarget, t, yearly]);

  const surface = useMemo(() => ({
    heroBg: isDark
      ? 'radial-gradient(circle at top right, rgba(37, 99, 235, 0.24), transparent 34%), linear-gradient(180deg, #0b1220 0%, #101a2b 72%)'
      : 'radial-gradient(circle at top right, rgba(22,119,255,0.18), transparent 34%), linear-gradient(180deg, #f7fbff 0%, #ffffff 72%)',
    heroCardBg: isDark
      ? 'linear-gradient(180deg, #132033 0%, #101a2b 100%)'
      : 'linear-gradient(180deg, #ffffff 0%, #f7fbff 100%)',
    sectionBg: isDark ? '#0f1728' : '#ffffff',
    altSectionBg: isDark ? '#0c1423' : '#fafbfc',
    cardBg: isDark
      ? 'linear-gradient(180deg, #132033 0%, #101a2b 100%)'
      : 'linear-gradient(180deg, #ffffff 0%, #f8fbff 100%)',
    border: 'var(--app-shell-border)',
    mutedText: 'var(--app-panel-text-muted)',
    proofBg: isDark ? 'rgba(16, 26, 43, 0.92)' : '#fff',
    proofBorder: isDark ? 'rgba(96, 165, 250, 0.16)' : '#d9e8ff',
    proofColor: isDark ? '#dbeafe' : '#2f3a4a',
    successBadgeBg: isDark ? 'rgba(22, 163, 74, 0.14)' : '#f6ffed',
    successBadgeColor: isDark ? '#86efac' : '#389e0d',
    successBadgeBorder: isDark ? 'rgba(34, 197, 94, 0.24)' : '#b7eb8f',
    softPanelBg: isDark
      ? 'linear-gradient(135deg, rgba(22,119,255,0.18), rgba(19,194,194,0.16))'
      : 'linear-gradient(135deg, rgba(22,119,255,0.08), rgba(54,207,201,0.12))',
    softPanelBorder: isDark ? 'rgba(96, 165, 250, 0.16)' : 'rgba(22,119,255,0.12)',
    panelShadow: isDark ? '0 28px 80px rgba(2, 8, 23, 0.28)' : '0 28px 80px rgba(15, 52, 96, 0.14)',
    statShadow: isDark ? '0 12px 30px rgba(2, 8, 23, 0.2)' : '0 8px 30px rgba(15, 52, 96, 0.06)',
    cardShadow: isDark ? '0 18px 44px rgba(2, 8, 23, 0.22)' : '0 16px 40px rgba(15,52,96,0.08)',
    pricingShadow: isDark ? '0 10px 35px rgba(2, 8, 23, 0.22)' : '0 10px 35px rgba(15,52,96,0.06)',
  }), [isDark]);

  return (
    <div style={{ width: '100%', overflow: 'hidden' }}>
      <Helmet>
        <html lang={currentLanguage} />
        <title>{homeTitle}</title>
        <meta name="description" content={homeDescription} />
        <meta
          name="keywords"
          content={homeKeywords}
        />
        <meta name="robots" content="index,follow" />
        <link rel="canonical" href={canonicalUrl} />
        <link rel="alternate" href="https://dotblue.ai/" hrefLang="x-default" />
        {SUPPORTED_LANGUAGES.map((language) => (
          <link key={language} rel="alternate" hrefLang={language} href={buildLocalizedUrl('/', language)} />
        ))}
        <meta property="og:type" content="website" />
        <meta property="og:title" content={homeTitle} />
        <meta property="og:description" content={homeDescription} />
        <meta property="og:url" content={canonicalUrl} />
      </Helmet>
      <section style={{
        padding: '112px 5% 88px',
        background: surface.heroBg,
      }}>
        <div style={{ maxWidth: 1180, margin: '0 auto' }}>
          <Row gutter={[48, 48]} align="middle">
            <Col xs={24} lg={14}>
              <Space orientation="vertical" size={20} style={{ width: '100%' }}>
                <Space wrap size={[12, 12]}>
                  <Tag color="blue" style={{ borderRadius: 999, padding: '6px 14px', marginInlineEnd: 0 }}>
                    {t('hero_tagline')}
                  </Tag>
                  <Tag style={{ borderRadius: 999, padding: '6px 14px', marginInlineEnd: 0, background: surface.successBadgeBg, color: surface.successBadgeColor, borderColor: surface.successBadgeBorder }}>
                    {t('hero_badge_product')}
                  </Tag>
                </Space>
                <Title style={{ fontSize: 'clamp(38px, 5.4vw, 62px)', margin: 0, lineHeight: 1.08, whiteSpace: 'pre-line', letterSpacing: '-0.03em' }}>
                  {t('hero_title')}
                </Title>
                <Paragraph style={{ fontSize: 'clamp(17px, 2vw, 20px)', color: surface.mutedText, margin: 0, maxWidth: 720 }}>
                  {t('hero_subtitle')}
                </Paragraph>
                <Space size="middle" wrap>
                  <Button
                    type="primary"
                    size="large"
                    shape="round"
                    icon={<ArrowRightOutlined />}
                    onClick={() => navigate(primaryTarget)}
                    style={{ height: 52, padding: '0 36px', fontSize: 16, fontWeight: 600 }}
                  >
                    {isAuthenticated ? t('go_to_dashboard') : t('hero_cta_primary')}
                  </Button>
                  <Button
                    size="large"
                    shape="round"
                    onClick={() => {
                      document.getElementById('highlights')?.scrollIntoView({ behavior: 'smooth', block: 'start' });
                    }}
                    style={{ height: 52, padding: '0 36px', fontSize: 16 }}
                  >
                    {t('hero_cta_secondary')}
                  </Button>
                </Space>
                <Space size={[16, 12]} wrap>
                  <Button type="link" onClick={() => navigate(getLocalizedPath('/docs', currentLanguage))} style={{ paddingInline: 0 }}>
                    {t('landing_nav_docs')}
                  </Button>
                  <Button
                    type="link"
                    href="https://github.com/12345dotblue/dotblue"
                    target="_blank"
                    rel="noreferrer"
                    style={{ paddingInline: 0 }}
                  >
                    {t('landing_nav_github')}
                  </Button>
                </Space>
                <Space wrap size={[10, 10]}>
                  {heroProofs.map((item) => (
                    <Tag key={item} style={{ borderRadius: 999, padding: '8px 14px', background: surface.proofBg, borderColor: surface.proofBorder, color: surface.proofColor, marginInlineEnd: 0 }}>
                      {item}
                    </Tag>
                  ))}
                </Space>
              </Space>
            </Col>
            <Col xs={24} lg={10}>
              <Card
                variant="borderless"
                style={{
                  borderRadius: 28,
                  background: surface.heroCardBg,
                  border: `1px solid ${surface.border}`,
                  boxShadow: surface.panelShadow,
                }}
                styles={{ body: { padding: 28 } }}
              >
                <Space orientation="vertical" size={18} style={{ width: '100%' }}>
                  <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                    <Text strong style={{ fontSize: 16 }}>{t('assistants_panel_title')}</Text>
                    <Tag color="processing" style={{ marginInlineEnd: 0 }}>{t('assistants_panel_badge')}</Tag>
                  </div>
                  <div style={{ borderRadius: 20, padding: 20, background: surface.softPanelBg, border: `1px solid ${surface.softPanelBorder}` }}>
                    <Space orientation="vertical" size={10} style={{ width: '100%' }}>
                      <Text type="secondary">{t('assistants_panel_label')}</Text>
                      <Title level={4} style={{ margin: 0 }}>{t('assistants_panel_featured_title')}</Title>
                      <Paragraph style={{ margin: 0, color: surface.mutedText }}>{t('assistants_panel_featured_desc')}</Paragraph>
                    </Space>
                  </div>
                  <div style={{ display: 'grid', gap: 12 }}>
                    {heroPanelItems.map((item) => (
                      <div
                        key={item.title}
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'space-between',
                          gap: 16,
                          padding: '14px 16px',
                          borderRadius: 18,
                          background: 'var(--app-panel-bg)',
                          border: `1px solid ${surface.border}`,
                        }}
                      >
                        <Space size={12} align="start">
                          <div style={{ width: 40, height: 40, borderRadius: 12, background: `${item.color}18`, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                            {React.cloneElement(item.icon as StyledIconElement, { style: { fontSize: 18, color: item.color } })}
                          </div>
                          <div>
                            <div style={{ fontSize: 14, fontWeight: 600, color: token.colorText }}>{item.title}</div>
                            <Text type="secondary" style={{ fontSize: 12 }}>{item.desc}</Text>
                          </div>
                        </Space>
                        <Tag style={{ marginInlineEnd: 0, borderRadius: 999 }}>{item.label}</Tag>
                      </div>
                    ))}
                  </div>
                </Space>
              </Card>
            </Col>
          </Row>

          <Row justify="center" style={{ marginTop: 56 }} gutter={[24, 24]}>
            {[
              { label: t('hero_stat_security'), value: t('hero_stat_value_security') },
              { label: t('hero_stat_sandbox'), value: t('hero_stat_value_sandbox') },
              { label: t('hero_stat_uptime'), value: t('hero_stat_value_uptime') },
              { label: t('hero_stat_cost'), value: t('hero_stat_value_cost') },
            ].map((s) => (
              <Col key={s.label} xs={24} sm={12} lg={6}>
                <div style={{ borderRadius: 24, padding: '20px 24px', background: 'var(--app-panel-bg)', border: `1px solid ${surface.border}`, boxShadow: surface.statShadow, textAlign: 'center' }}>
                  <div style={{ fontSize: 30, fontWeight: 700, color: token.colorPrimary, lineHeight: 1.1 }}>{s.value}</div>
                  <Text type="secondary" style={{ fontSize: 13 }}>{s.label}</Text>
                </div>
              </Col>
            ))}
          </Row>
        </div>
      </section>

      <section id="highlights" style={{ padding: '96px 5%', background: surface.sectionBg }}>
        <div style={{ textAlign: 'center', maxWidth: 720, margin: '0 auto 56px' }}>
          <Title level={2} style={{ marginBottom: 12 }}>{t('highlights_title')}</Title>
          <Paragraph type="secondary" style={{ fontSize: 16, marginBottom: 0 }}>
            {t('highlights_subtitle')}
          </Paragraph>
        </div>
        <Row gutter={[24, 24]} justify="center" style={{ maxWidth: 1180, margin: '0 auto' }}>
          {valueProps.map((item) => (
            <Col xs={24} md={12} xl={6} key={item.title}>
              <Card
                variant="borderless"
                hoverable
                style={{
                  height: '100%',
                  borderRadius: 24,
                  background: surface.cardBg,
                  border: `1px solid ${surface.border}`,
                  boxShadow: surface.cardShadow,
                }}
                styles={{ body: { padding: 28 } }}
              >
                <Space orientation="vertical" size={18} style={{ width: '100%' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 16 }}>
                    <div style={{ width: 52, height: 52, borderRadius: 16, background: `${item.color}18`, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                      {React.cloneElement(item.icon as StyledIconElement, { style: { fontSize: 24, color: item.color } })}
                    </div>
                    <Tag color="default" style={{ marginInlineEnd: 0, borderRadius: 999 }}>{item.metric}</Tag>
                  </div>
                  <div>
                    <Text type="secondary" style={{ textTransform: 'uppercase', letterSpacing: 1.2, fontSize: 12 }}>
                      {item.label}
                    </Text>
                    <Title level={4} style={{ marginTop: 10, marginBottom: 10 }}>{item.title}</Title>
                    <Paragraph type="secondary" style={{ marginBottom: 0 }}>{item.desc}</Paragraph>
                  </div>
                </Space>
              </Card>
            </Col>
          ))}
        </Row>
      </section>

      <section style={{ padding: '100px 5%', background: surface.altSectionBg }}>
        <div style={{ textAlign: 'center', maxWidth: 680, margin: '0 auto 56px' }}>
          <Title level={2} style={{ marginBottom: 12 }}>{t('features_title')}</Title>
          <Paragraph type="secondary" style={{ fontSize: 16, marginBottom: 0 }}>{t('features_subtitle')}</Paragraph>
        </div>
        <Row gutter={[24, 24]} justify="center" style={{ maxWidth: 1180, margin: '0 auto' }}>
          {capabilities.map((feature) => (
            <Col xs={24} sm={12} lg={8} key={feature.title}>
              <Card hoverable variant="borderless" style={{ height: '100%', borderRadius: 20, background: 'var(--app-panel-bg)', border: `1px solid ${surface.border}`, boxShadow: surface.cardShadow }} styles={{ body: { padding: 28 } }}>
                <div style={{
                  width: 48, height: 48, borderRadius: 12,
                  background: `${feature.color}15`, display: 'flex', alignItems: 'center', justifyContent: 'center',
                  marginBottom: 20,
                }}>
                  {React.cloneElement(feature.icon as StyledIconElement, { style: { fontSize: 24, color: feature.color } })}
                </div>
                <Title level={4} style={{ marginBottom: 8 }}>{feature.title}</Title>
                <Paragraph type="secondary" style={{ marginBottom: 0 }}>{feature.desc}</Paragraph>
              </Card>
            </Col>
          ))}
        </Row>
      </section>

      <section style={{ padding: '96px 5%', background: surface.sectionBg }}>
        <div style={{ textAlign: 'center', maxWidth: 720, margin: '0 auto 56px' }}>
          <Title level={2} style={{ marginBottom: 12 }}>{t('deployment_title')}</Title>
          <Paragraph type="secondary" style={{ fontSize: 16, marginBottom: 0 }}>
            {t('deployment_subtitle')}
          </Paragraph>
        </div>
        <Row gutter={[24, 24]} justify="center" style={{ maxWidth: 1180, margin: '0 auto' }}>
          {deploymentModes.map((item) => (
            <Col xs={24} md={12} xl={8} key={item.title}>
              <Card
                variant="borderless"
                style={{
                  height: '100%',
                  borderRadius: 24,
                  background: surface.cardBg,
                  border: `1px solid ${surface.border}`,
                  boxShadow: surface.cardShadow,
                }}
                styles={{ body: { padding: 28 } }}
              >
                <Space orientation="vertical" size={18} style={{ width: '100%' }}>
                  <div style={{ width: 52, height: 52, borderRadius: 16, background: `${item.color}18`, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                    {React.cloneElement(item.icon as StyledIconElement, { style: { fontSize: 24, color: item.color } })}
                  </div>
                  <div>
                    <Title level={4} style={{ marginBottom: 10 }}>{item.title}</Title>
                    <Paragraph type="secondary" style={{ marginBottom: 0 }}>{item.desc}</Paragraph>
                  </div>
                  <Space wrap size={[10, 10]}>
                    {item.points.map((point) => (
                      <Tag key={point} style={{ marginInlineEnd: 0, borderRadius: 999 }}>{point}</Tag>
                    ))}
                  </Space>
                </Space>
              </Card>
            </Col>
          ))}
        </Row>
      </section>

      <section id="assistants" style={{ padding: '88px 5%', background: surface.altSectionBg }}>
        <div style={{ maxWidth: 1180, margin: '0 auto' }}>
          <div style={{ textAlign: 'center', maxWidth: 760, margin: '0 auto 56px' }}>
            <Title level={2} style={{ marginBottom: 12 }}>{t('assistants_title')}</Title>
            <Paragraph type="secondary" style={{ fontSize: 16, marginBottom: 0 }}>
              {t('assistants_subtitle')}
            </Paragraph>
          </div>
          <Row gutter={[24, 24]}>
            {assistants.map((assistant) => (
              <Col xs={24} md={12} xl={6} key={assistant.title}>
                <Card
                  hoverable
                  variant="borderless"
                  style={{
                    height: '100%',
                    borderRadius: 22,
                    background: surface.cardBg,
                    border: `1px solid ${surface.border}`,
                    boxShadow: surface.cardShadow,
                  }}
                  styles={{ body: { padding: 24 } }}
                >
                  <Space orientation="vertical" size={16} style={{ width: '100%' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 12 }}>
                      <div style={{ width: 48, height: 48, borderRadius: 14, background: `${assistant.color}18`, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                        {React.cloneElement(assistant.icon as StyledIconElement, { style: { fontSize: 22, color: assistant.color } })}
                      </div>
                      <Tag color="blue" style={{ marginInlineEnd: 0, borderRadius: 999 }}>{assistant.tag}</Tag>
                    </div>
                    <div>
                      <Title level={4} style={{ marginBottom: 8 }}>{assistant.title}</Title>
                      <Paragraph type="secondary" style={{ marginBottom: 0 }}>{assistant.desc}</Paragraph>
                    </div>
                  </Space>
                </Card>
              </Col>
            ))}
          </Row>
        </div>
      </section>

      <section id="pricing" style={{ padding: '96px 5%', background: surface.sectionBg }}>
        <div style={{ textAlign: 'center', maxWidth: 680, margin: '0 auto 48px' }}>
          <Title level={2} style={{ marginBottom: 12 }}>{t('pricing_title')}</Title>
          <Paragraph type="secondary" style={{ fontSize: 16, marginBottom: 0 }}>{t('pricing_subtitle')}</Paragraph>
          <div style={{ marginTop: 24, display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 12 }}>
            <Text type={yearly ? 'secondary' : undefined} strong={!yearly}>{t('pricing_monthly')}</Text>
            <Switch checked={yearly} onChange={setYearly} />
            <Text type={yearly ? undefined : 'secondary'} strong={yearly}>{t('pricing_yearly')}</Text>
            <Tag color="green">{t('pricing_yearly_discount')}</Tag>
          </div>
        </div>
        <Row gutter={[24, 24]} justify="center" style={{ maxWidth: 1180, margin: '0 auto' }}>
          {plans.map((plan) => (
            <Col xs={24} md={12} xl={8} key={plan.name}>
              <Card
                variant={plan.popular ? 'outlined' : 'borderless'}
                style={{
                  height: '100%',
                  borderRadius: 24,
                  position: 'relative',
                  background: 'var(--app-panel-bg)',
                  boxShadow: plan.popular ? `0 14px 50px ${token.colorPrimary}22` : surface.pricingShadow,
                  borderColor: plan.popular ? token.colorPrimary : surface.border,
                }}
                styles={{ body: { padding: 28 } }}
              >
                {plan.popular && (
                  <Tag color="blue" style={{ position: 'absolute', top: -12, left: '50%', transform: 'translateX(-50%)', fontSize: 12, borderRadius: 999 }}>
                    {t('pricing_pro_popular')}
                  </Tag>
                )}
                <div style={{ textAlign: 'center', paddingTop: plan.popular ? 8 : 0 }}>
                  <Title level={4} style={{ marginBottom: 4 }}>{plan.name}</Title>
                  <div style={{ fontSize: 42, fontWeight: 700, color: token.colorPrimary, margin: '16px 0 8px' }}>
                    {plan.price}
                    {plan.price !== t('pricing_enterprise_price') && <span style={{ fontSize: 16, fontWeight: 400, color: surface.mutedText }}>/{yearly ? t('pricing_yearly') : t('pricing_monthly')}</span>}
                  </div>
                  <Paragraph type="secondary" style={{ marginBottom: 0 }}>{plan.desc}</Paragraph>
                </div>
                <div style={{ margin: '24px 0 28px', display: 'flex', flexDirection: 'column', gap: 12 }}>
                  {plan.features.map((feature) => (
                    <div key={feature} style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                      <CheckOutlined style={{ color: token.colorPrimary }} />
                      <Text>{feature}</Text>
                    </div>
                  ))}
                </div>
                <Button
                  type={plan.popular ? 'primary' : 'default'}
                  block
                  size="large"
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

      <section style={{ padding: '80px 5% 96px' }}>
        <div style={{ maxWidth: 1180, margin: '0 auto' }}>
          <div style={{
            background: isDark
              ? 'linear-gradient(135deg, #1d4ed8 0%, #0f2f7a 100%)'
              : 'linear-gradient(135deg, #1677ff 0%, #003eb3 100%)',
            borderRadius: 32, padding: '72px 40px', textAlign: 'center', color: '#fff',
            boxShadow: isDark ? '0 20px 40px rgba(15, 23, 42, 0.36)' : '0 20px 40px rgba(22,119,255,0.2)',
          }}>
            <Title level={2} style={{ color: '#fff', marginBottom: 16 }}>{t('cta_title')}</Title>
            <Paragraph style={{ color: 'rgba(255,255,255,0.85)', fontSize: 18, margin: '0 auto 28px', maxWidth: 680 }}>
              {t('cta_subtitle')}
            </Paragraph>
            <Space wrap size={[10, 10]} style={{ justifyContent: 'center', marginBottom: 32 }}>
              {[t('cta_point_one'), t('cta_point_two'), t('cta_point_three')].map((item) => (
                <Tag key={item} style={{ borderRadius: 999, padding: '8px 14px', color: '#fff', background: 'rgba(255,255,255,0.14)', borderColor: 'rgba(255,255,255,0.24)', marginInlineEnd: 0 }}>
                  {item}
                </Tag>
              ))}
            </Space>
            <Button size="large" shape="round" ghost onClick={() => navigate(primaryTarget)} style={{ height: 52, padding: '0 48px', fontSize: 16, fontWeight: 600 }}>
              {isAuthenticated ? t('go_to_dashboard') : t('cta_button')}
            </Button>
          </div>
        </div>
      </section>
    </div>
  );
};

export default LandingPage;
