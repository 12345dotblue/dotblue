import React from 'react';
import { Layout, Button, Space, Typography, Row, Col, Divider, Dropdown } from 'antd';
import { AppstoreOutlined, GithubOutlined, GlobalOutlined, LogoutOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { Link, useNavigate } from 'react-router-dom';
import { casdoorService } from '../../domains/identity/CasdoorService';
import { useAuthState } from '../../domains/identity/useAuthState';
import { LANGUAGE_OPTIONS, applyLanguagePreference, getLocalizedPath, resolveSupportedLanguage, stripLanguagePrefix } from '../../i18n/config';

const { Header, Content, Footer } = Layout;
const { Title, Text, Paragraph } = Typography;

const LandingLayout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const isAuthenticated = useAuthState();
  const currentLanguage = resolveSupportedLanguage(i18n.resolvedLanguage || i18n.language);
  const currentLanguageLabel = LANGUAGE_OPTIONS.find((item) => item.value === currentLanguage)?.shortLabel || 'EN';

  const changeLanguage = async (lng: string) => {
    const resolved = await applyLanguagePreference(lng);
    if (resolveSupportedLanguage(i18n.resolvedLanguage || i18n.language) !== resolved) {
      window.location.reload();
    }
  };

  const scrollToSection = (sectionId: string) => {
    if (stripLanguagePrefix(window.location.pathname) !== '/') {
      window.location.href = `${getLocalizedPath('/', currentLanguage)}#${sectionId}`;
      return;
    }

    window.history.replaceState(null, '', `${getLocalizedPath('/', currentLanguage)}#${sectionId}`);
    document.getElementById(sectionId)?.scrollIntoView({ behavior: 'smooth', block: 'start' });
  };

  React.useEffect(() => {
    if (stripLanguagePrefix(window.location.pathname) === '/' && window.location.hash) {
      const targetId = window.location.hash.replace('#', '');
      window.setTimeout(() => {
        document.getElementById(targetId)?.scrollIntoView({ behavior: 'smooth', block: 'start' });
      }, 120);
    }
  }, []);

  return (
    <Layout style={{ minHeight: '100vh', background: '#fff' }}>
      <Header style={{
        background: 'rgba(255, 255, 255, 0.72)',
        backdropFilter: 'blur(16px)',
        padding: '0 5%',
        position: 'sticky',
        top: 0,
        zIndex: 100,
        borderBottom: '1px solid rgba(15, 23, 42, 0.05)',
        height: 88,
      }}>
        <div className="landing-header-shell">
          <div className="landing-header-main">
            <div
              className="landing-brand-block"
              onClick={() => navigate(getLocalizedPath('/', currentLanguage))}
            >
              <img
                src="/brand/dotblue-logo.png"
                alt="dotblue"
                style={{ width: 132, height: 42, objectFit: 'contain', flexShrink: 0 }}
              />
              <div
                style={{
                  width: 1,
                  height: 22,
                  background: 'linear-gradient(180deg, rgba(22,119,255,0.04) 0%, rgba(22,119,255,0.24) 50%, rgba(22,119,255,0.04) 100%)',
                  flexShrink: 0,
                }}
              />
              <div style={{ minWidth: 0, display: 'flex', flexDirection: 'column', gap: 1 }}>
                <Text
                  style={{
                    fontSize: 13,
                    fontWeight: 600,
                    color: '#0f172a',
                    lineHeight: 1.2,
                    whiteSpace: 'nowrap',
                  }}
                >
                  {t('brand_header_badge')}
                </Text>
                <Text
                  type="secondary"
                  style={{
                    fontSize: 11,
                    lineHeight: 1.2,
                    whiteSpace: 'nowrap',
                  }}
                >
                  {t('brand_header_subtitle')}
                </Text>
              </div>
            </div>
            <Space size={4} className="landing-nav-group">
              <Button className="landing-nav-button" type="text" onClick={() => scrollToSection('assistants')}>
                {t('landing_nav_assistants')}
              </Button>
              <Button className="landing-nav-button" type="text" onClick={() => scrollToSection('highlights')}>
                {t('landing_nav_highlights')}
              </Button>
              <Button className="landing-nav-button" type="text" onClick={() => scrollToSection('pricing')}>
                {t('view_pricing')}
              </Button>
              <Button className="landing-nav-button" type="text" onClick={() => navigate(getLocalizedPath('/docs', currentLanguage))}>
                {t('landing_nav_docs')}
              </Button>
              <Button
                className="landing-nav-button"
                type="text"
                icon={<GithubOutlined />}
                href="https://github.com/12345dotblue/dotblue"
                target="_blank"
                rel="noreferrer"
              >
                {t('landing_nav_github')}
              </Button>
              <Button className="landing-nav-button landing-nav-button--secondary" type="text" onClick={() => navigate(getLocalizedPath('/terms', currentLanguage))}>
                {t('terms')}
              </Button>
              <Button className="landing-nav-button landing-nav-button--secondary" type="text" onClick={() => navigate(getLocalizedPath('/privacy', currentLanguage))}>
                {t('privacy')}
              </Button>
            </Space>
          </div>

          <Space size="small" className="landing-header-actions">
            <Dropdown
              menu={{
                items: LANGUAGE_OPTIONS.map((item) => ({
                  key: item.value,
                  label: item.label,
                })),
                onClick: ({ key }) => changeLanguage(String(key)),
              }}
              trigger={['click']}
            >
              <Button className="landing-utility-button" type="text" icon={<GlobalOutlined />}>
                {currentLanguageLabel}
              </Button>
            </Dropdown>
            {isAuthenticated ? (
              <>
                <Button className="landing-secondary-button" shape="round" icon={<AppstoreOutlined />} onClick={() => navigate(getLocalizedPath('/dashboard', currentLanguage))}>
                  {t('go_to_dashboard')}
                </Button>
                <Button
                  className="landing-primary-button"
                  type="primary"
                  shape="round"
                  icon={<LogoutOutlined />}
                  onClick={() => {
                    casdoorService.removeToken();
                    navigate(getLocalizedPath('/login', currentLanguage));
                  }}
                >
                  {t('logout')}
                </Button>
              </>
            ) : (
              <Button className="landing-primary-button" type="primary" shape="round" onClick={() => navigate(getLocalizedPath('/login', currentLanguage))}>
                {t('login')}
              </Button>
            )}
          </Space>
        </div>
      </Header>

      <Content>{children}</Content>

      <Footer style={{ background: '#f9fafb', padding: '64px 5% 32px' }}>
        <Row gutter={[32, 32]}>
          <Col xs={24} md={8}>
            <div style={{ display: 'flex', alignItems: 'center', marginBottom: 16 }}>
              <img
                src="/brand/dotblue-logo.png"
                alt="dotblue"
                style={{ width: 104, height: 32, objectFit: 'contain', marginRight: 10 }}
              />
            </div>
            <Paragraph type="secondary">{t('footer_desc')}</Paragraph>
            <Text type="secondary" style={{ fontSize: 12 }}>
              &copy; 2026 Dotblue Tech Co., Ltd.
            </Text>
          </Col>
          <Col xs={12} md={4}>
            <Title level={5}>{t('footer_product')}</Title>
            <Space orientation="vertical">
              <Link to={getLocalizedPath('/', currentLanguage)}>{t('welcome')}</Link>
              <Link to={getLocalizedPath('/docs', currentLanguage)}>{t('landing_nav_docs')}</Link>
              <a href={`${getLocalizedPath('/', currentLanguage)}#highlights`}>{t('landing_nav_highlights')}</a>
              <a href={`${getLocalizedPath('/', currentLanguage)}#pricing`}>{t('view_pricing')}</a>
            </Space>
          </Col>
          <Col xs={12} md={4}>
            <Title level={5}>{t('footer_legal')}</Title>
            <Space orientation="vertical">
              <Link to={getLocalizedPath('/terms', currentLanguage)}>{t('terms')}</Link>
              <Link to={getLocalizedPath('/privacy', currentLanguage)}>{t('privacy')}</Link>
              <Link to={getLocalizedPath('/refund', currentLanguage)}>{t('refund')}</Link>
            </Space>
          </Col>
          <Col xs={24} md={8}>
            <Title level={5}>{t('contact_us')}</Title>
            <Paragraph type="secondary">
              Email: support@dotblue.ai
            </Paragraph>
            <Paragraph style={{ marginBottom: 0 }}>
              <a href="https://github.com/12345dotblue/dotblue" target="_blank" rel="noreferrer">
                <Space size={8}>
                  <GithubOutlined />
                  <span>{t('landing_nav_github')}</span>
                </Space>
              </a>
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
      <style>
        {`
          .landing-header-shell {
            max-width: 1240px;
            margin: 12px auto;
            height: 64px;
            padding: 0 18px 0 20px;
            border-radius: 20px;
            border: 1px solid rgba(15, 23, 42, 0.06);
            background: rgba(255, 255, 255, 0.86);
            box-shadow: 0 12px 32px rgba(15, 23, 42, 0.06);
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 20px;
          }

          .landing-header-main {
            display: flex;
            align-items: center;
            gap: 20px;
            min-width: 0;
            flex: 1 1 auto;
            overflow: hidden;
          }

          .landing-brand-block {
            display: flex;
            align-items: center;
            gap: 14px;
            cursor: pointer;
            min-width: 0;
            flex: 0 0 auto;
          }

          .landing-nav-group {
            display: flex;
            flex-wrap: nowrap !important;
            min-width: 0;
            overflow-x: auto;
            overflow-y: hidden;
            scrollbar-width: none;
            -ms-overflow-style: none;
            white-space: nowrap;
            flex: 1 1 auto;
          }

          .landing-nav-group::-webkit-scrollbar {
            display: none;
          }

          .landing-nav-group .ant-space-item {
            flex: 0 0 auto;
          }

          .landing-nav-group .landing-nav-button {
            height: 36px;
            padding: 0 14px;
            border-radius: 999px;
            color: #334155;
            font-weight: 500;
          }

          .landing-nav-group .landing-nav-button .anticon {
            font-size: 14px;
          }

          .landing-nav-group .landing-nav-button:hover {
            color: #1677ff !important;
            background: #eff6ff !important;
          }

          .landing-utility-button {
            height: 38px;
            padding: 0 14px;
            border-radius: 999px;
            border: 1px solid rgba(148, 163, 184, 0.22);
            background: #fff;
            color: #334155;
            font-weight: 500;
          }

          .landing-utility-button:hover {
            color: #1677ff !important;
            border-color: rgba(22, 119, 255, 0.22) !important;
            background: #f8fbff !important;
          }

          .landing-secondary-button {
            height: 40px;
            padding: 0 18px;
            border-radius: 999px;
            border-color: rgba(148, 163, 184, 0.28);
            color: #0f172a;
            font-weight: 600;
            box-shadow: none;
          }

          .landing-secondary-button:hover {
            color: #1677ff !important;
            border-color: rgba(22, 119, 255, 0.24) !important;
            background: #f8fbff !important;
          }

          .landing-primary-button {
            height: 40px;
            padding: 0 18px;
            border-radius: 999px;
            font-weight: 600;
            box-shadow: 0 10px 24px rgba(22, 119, 255, 0.18);
          }

          .landing-header-actions {
            flex: 0 0 auto;
            white-space: nowrap;
          }

          @media (max-width: 1320px) {
            .landing-nav-button--secondary {
              display: none;
            }
          }

          @media (max-width: 1180px) {
            .landing-brand-block > div:last-child {
              display: none !important;
            }
          }
        `}
      </style>
    </Layout>
  );
};

export default LandingLayout;
