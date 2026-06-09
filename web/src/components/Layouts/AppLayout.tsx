import React, { useState, useMemo } from 'react';
import { Layout, Menu, Button, Dropdown, Space, Typography, theme, Drawer, Breadcrumb, Avatar, Select, Tag, message } from 'antd';
import {
  MessageOutlined,
  GlobalOutlined,
  LogoutOutlined,
  UserOutlined,
  AppstoreOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  SettingOutlined,
  TeamOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useNavigate, useLocation, Link } from 'react-router-dom';
import axios from 'axios';
import { casdoorService } from '../../domains/identity/CasdoorService';
import { BACKEND_URL } from '../../config';
import { LANGUAGE_OPTIONS, applyLanguagePreference, getLocalizedPath, resolveSupportedLanguage, stripLanguagePrefix } from '../../i18n/config';
import ThemeModeDropdown from '../ThemeModeDropdown';
import { useThemeMode } from '../../theme/themeMode';

const { Header, Content, Sider } = Layout;
const { Title, Text } = Typography;

interface AppLayoutProps {
  children: React.ReactNode;
}

interface EnterpriseMembership {
  enterpriseId: string;
  name: string;
  role: string;
}

const CURRENT_ENTERPRISE_STORAGE_KEY = 'dotblue_current_enterprise_id';

function getAuthHeaders() {
  const token = casdoorService.getToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

const AppLayout: React.FC<AppLayoutProps> = ({ children }) => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const { token } = theme.useToken();
  const { resolvedTheme } = useThemeMode();
  const [collapsed, setCollapsed] = useState(false);
  const [mobileVisible, setMobileVisible] = useState(false);
  const [enterpriseLoading, setEnterpriseLoading] = React.useState(false);
  const [switchingEnterprise, setSwitchingEnterprise] = React.useState(false);
  const [enterpriseList, setEnterpriseList] = React.useState<EnterpriseMembership[]>([]);
  const [currentEnterpriseId, setCurrentEnterpriseId] = React.useState<string>();
  const [openKeys, setOpenKeys] = React.useState<string[]>([]);
  const [messageApi, contextHolder] = message.useMessage();

  const isAdmin = casdoorService.isAdmin();
  const normalizedPath = stripLanguagePrefix(location.pathname);
  const isChatPage = normalizedPath === '/chat';
  const username = casdoorService.getUsername() || t('common_user');
  const currentLanguage = resolveSupportedLanguage(i18n.resolvedLanguage || i18n.language);
  const currentLanguageLabel = LANGUAGE_OPTIONS.find((item) => item.value === currentLanguage)?.shortLabel || currentLanguage.toUpperCase();

  React.useEffect(() => {
    if (!casdoorService.isAuthenticated()) {
      return;
    }
    setEnterpriseLoading(true);
    Promise.all([
      axios.get(`${BACKEND_URL}/api/enterprises`, { headers: getAuthHeaders() }),
      axios.get(`${BACKEND_URL}/api/enterprises/current`, { headers: getAuthHeaders() }),
    ]).then(([listRes, currentRes]) => {
      const rawList = Array.isArray(listRes.data) ? listRes.data : [];
      const list = rawList.filter((item: EnterpriseMembership, index: number, all: EnterpriseMembership[]) =>
        all.findIndex((candidate) => candidate.enterpriseId === item.enterpriseId) === index,
      );
      const resolvedEnterpriseId = currentRes.data?.enterpriseId || list[0]?.enterpriseId;
      setEnterpriseList(list);
      setCurrentEnterpriseId(resolvedEnterpriseId);
      if (resolvedEnterpriseId) {
        localStorage.setItem(CURRENT_ENTERPRISE_STORAGE_KEY, resolvedEnterpriseId);
      }
    }).catch(() => {
      setEnterpriseList([]);
      setCurrentEnterpriseId(undefined);
      localStorage.removeItem(CURRENT_ENTERPRISE_STORAGE_KEY);
    }).finally(() => {
      setEnterpriseLoading(false);
    });
  }, [location.pathname]);

  const changeLanguage = async (lng: string) => {
    const resolved = await applyLanguagePreference(lng);
    navigate(`${getLocalizedPath(normalizedPath, resolved)}${location.search}${location.hash}`, { replace: true });
  };

  const handleEnterpriseSwitch = async (enterpriseId: string) => {
    if (!enterpriseId || enterpriseId === currentEnterpriseId) {
      return;
    }
    setSwitchingEnterprise(true);
    try {
      await axios.post(`${BACKEND_URL}/api/enterprises/switch`, { enterpriseId }, {
        headers: getAuthHeaders(),
      });
      setCurrentEnterpriseId(enterpriseId);
      localStorage.setItem(CURRENT_ENTERPRISE_STORAGE_KEY, enterpriseId);
      messageApi.success(t('enterprise_switch_success'));
      window.location.reload();
    } catch {
      messageApi.error(t('enterprise_switch_failed'));
    } finally {
      setSwitchingEnterprise(false);
    }
  };

  React.useEffect(() => {
    if (normalizedPath.startsWith('/admin/platform')) {
      setOpenKeys(['platform-admin']);
      return;
    }
    setOpenKeys([]);
  }, [normalizedPath]);

  const currentEnterprise = enterpriseList.find((item) => item.enterpriseId === currentEnterpriseId);
  const canManageEnterprise = ['owner', 'admin'].includes((currentEnterprise?.role || '').toLowerCase());
  const selectedMenuKey = useMemo(() => {
    if (normalizedPath.startsWith('/admin/platform/skills/create')) return '/admin/platform/skills/new';
    if (normalizedPath.startsWith('/admin/platform/skills/import')) return '/admin/platform/skill-market';
    if (normalizedPath.startsWith('/admin/platform/skill-market')) return '/admin/platform/skill-market';
    if (normalizedPath.startsWith('/admin/platform/skill-hubs')) return '/admin/platform/skills';
    if (normalizedPath.startsWith('/dashboard/agents/') && normalizedPath.endsWith('/skills')) return '/dashboard';
    return normalizedPath;
  }, [normalizedPath]);

  const menuItems = useMemo(() => {
    const base = [
      {
        key: '/dashboard',
        icon: <AppstoreOutlined style={{ fontSize: '18px' }} />,
        label: t('agent_settings'),
      },
      {
        key: '/chat',
        icon: <MessageOutlined style={{ fontSize: '18px' }} />,
        label: t('chat'),
      },
    ];

    const adminItems = [];
    if (canManageEnterprise) {
      adminItems.push({
        key: '/admin/enterprise',
        icon: <TeamOutlined style={{ fontSize: '18px' }} />,
        label: t('enterprise_admin_nav'),
      });
    }
    if (isAdmin) {
      adminItems.push({
        key: 'platform-admin',
        icon: <SettingOutlined style={{ fontSize: '18px' }} />,
        label: t('platform_admin_nav'),
        children: [
          {
            key: '/admin/platform',
            label: t('platform_settings_nav'),
          },
          {
            key: '/admin/platform/skill-market',
            label: t('platform_skill_market_nav'),
          },
          {
            key: '/admin/platform/skills/new',
            label: t('platform_skill_builder_nav'),
          },
          {
            key: '/admin/platform/skills',
            label: t('platform_skill_governance_nav'),
          },
        ],
      });
    }

    return adminItems.length ? [...base, { type: 'divider' as const }, ...adminItems] : base;
  }, [t, isAdmin, canManageEnterprise]);

  const getPageTitle = () => {
    if (normalizedPath === '/dashboard') return t('agent_settings');
    if (normalizedPath === '/chat') return t('chat');
    if (normalizedPath === '/admin/settings' || normalizedPath === '/admin/enterprise') return t('enterprise_admin_nav');
    if (normalizedPath === '/admin/platform') return t('platform_settings_nav');
    if (normalizedPath === '/admin/platform/skill-market') return t('platform_skill_market_title');
    if (normalizedPath === '/admin/platform/skills/new') return t('platform_skill_builder_title');
    if (normalizedPath.startsWith('/dashboard/agents/') && normalizedPath.endsWith('/skills')) return t('agent_skill_page_title');
    if (normalizedPath.startsWith('/admin/platform/skills/create')) return t('platform_skill_builder_title');
    if (normalizedPath.startsWith('/admin/platform/skills/import')) return t('platform_skill_market_title');
    if (normalizedPath.startsWith('/admin/platform/skill-hubs')) return t('platform_skill_governance_title');
    if (normalizedPath.startsWith('/admin/platform/skills')) return t('platform_skills_title');
    return '';
  };
  const getPageSectionTitle = () => {
    if (normalizedPath === '/admin/platform' || normalizedPath.startsWith('/admin/platform/skills')) {
      return t('platform_admin_nav');
    }
    return undefined;
  };
  const hidePageHeading = normalizedPath === '/admin/enterprise';

  // Keep hook order stable across route changes, then branch the render.
  if (isChatPage) {
    return <>{children}</>;
  }

  const sideContent = (
    <div className="app-sider-shell" style={{ display: 'flex', flexDirection: 'column', height: '100%', background: 'var(--app-nav-bg)' }}>
      <div className={`app-sider-brand ${collapsed && !mobileVisible ? 'app-sider-brand--collapsed' : ''}`}>
        <img
          src={collapsed && !mobileVisible ? '/brand/dotblue-favicon.svg' : resolvedTheme === 'dark' ? '/brand/dotblue-logo-dark.svg' : '/brand/dotblue-logo-light.svg'}
          alt={t('app_name')}
          className="app-sider-brand-logo"
          onClick={() => {
            navigate(getLocalizedPath('/', currentLanguage));
            setMobileVisible(false);
          }}
        />
      </div>
      <Menu
        theme={resolvedTheme === 'dark' ? 'dark' : 'light'}
        mode="inline"
        inlineCollapsed={collapsed}
        selectedKeys={[selectedMenuKey]}
        openKeys={collapsed || mobileVisible ? openKeys : openKeys}
        onOpenChange={(keys) => setOpenKeys(keys as string[])}
        items={menuItems}
        onClick={({ key }) => {
          navigate(getLocalizedPath(String(key), currentLanguage));
          setMobileVisible(false);
        }}
        style={{ flex: 1, paddingTop: 16, borderRight: 0 }}
      />
    </div>
  );

  return (
    <Layout style={{ minHeight: '100vh' }}>
      {contextHolder}
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        breakpoint="lg"
        onBreakpoint={(broken) => setCollapsed(broken)}
        theme={resolvedTheme === 'dark' ? 'dark' : 'light'}
        className="desktop-sider"
        width={240}
        style={{
          position: 'fixed',
          height: '100vh',
          left: 0,
          top: 0,
          bottom: 0,
          boxShadow: '0 18px 48px rgba(2, 8, 23, 0.24)',
          zIndex: 20,
        }}
      >
        {sideContent}
      </Sider>

      <Drawer
        placement="left"
        onClose={() => setMobileVisible(false)}
        open={mobileVisible}
        styles={{ body: { padding: 0 } }}
        size="default"
        closable={false}
      >
        {sideContent}
      </Drawer>

      <Layout style={{
        marginLeft: collapsed ? 80 : 240,
        transition: 'all 0.2s cubic-bezier(0.645, 0.045, 0.355, 1)',
        background: 'var(--app-shell-bg)',
      }} className="main-layout">
        <Header style={{
          background: 'var(--app-header-bg)',
          backdropFilter: 'blur(12px)',
          padding: '0 24px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          borderBottom: '1px solid var(--app-shell-border)',
          position: 'sticky',
          top: 0,
          zIndex: 10,
          height: 64,
        }}>
          <Space size="large">
            <Button
              type="text"
              icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={() => {
                if (window.innerWidth < 992) setMobileVisible(true);
                else setCollapsed(!collapsed);
              }}
              style={{ fontSize: '18px' }}
            />
            {enterpriseList.length > 0 && (
              <Space size="small" align="start">
                <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
                  <Text type="secondary" style={{ fontSize: 12, lineHeight: 1 }}>
                    {t('enterprise_switch_label')}
                  </Text>
                  <Select
                    value={currentEnterpriseId}
                    loading={enterpriseLoading || switchingEnterprise}
                    onChange={handleEnterpriseSwitch}
                    style={{ minWidth: 260 }}
                    options={enterpriseList.map((item) => ({
                      label: `${item.name} · ${item.enterpriseId.slice(0, 8)}`,
                      value: item.enterpriseId,
                    }))}
                  />
                </div>
                {currentEnterprise?.role && (
                  <Tag color={currentEnterprise.role === 'owner' ? 'gold' : currentEnterprise.role === 'admin' ? 'blue' : 'default'}>
                    {t(`enterprise_role_${currentEnterprise.role}`)}
                  </Tag>
                )}
              </Space>
            )}
          </Space>

          <Space size="middle">
            <ThemeModeDropdown />
            <Dropdown menu={{
              items: LANGUAGE_OPTIONS.map((item) => ({
                key: item.value,
                label: item.label,
              })),
              onClick: ({ key }) => changeLanguage(String(key)),
            }} trigger={['click']}>
              <Button type="text" icon={<GlobalOutlined />}>
                {currentLanguageLabel}
              </Button>
            </Dropdown>
            <Dropdown menu={{
              items: [
                {
                  key: 'logout',
                  icon: <LogoutOutlined />,
                  label: t('logout'),
                  onClick: () => {
                    casdoorService.removeToken();
                    window.location.href = getLocalizedPath('/login', currentLanguage);
                  },
                },
              ],
            }}>
              <Button type="text" style={{ height: 40, padding: '4px 8px' }}>
                <Space>
                  <Avatar size="small" icon={<UserOutlined />} style={{ backgroundColor: token.colorPrimary }} />
                  <Text strong>{username}</Text>
                </Space>
              </Button>
            </Dropdown>
          </Space>
        </Header>

        <Content style={{ padding: '24px', maxWidth: 1400, margin: '0 auto', width: '100%', color: token.colorText }}>
          <div style={{ marginBottom: 24 }}>
            <Breadcrumb
              items={[
                { title: <Link to={getLocalizedPath('/dashboard', currentLanguage)}>{t('app_name')}</Link> },
                ...(getPageSectionTitle() ? [{ title: getPageSectionTitle() }] : []),
                { title: getPageTitle() },
              ]}
              style={{ marginBottom: hidePageHeading ? 0 : 8 }}
            />
            {!hidePageHeading && <Title level={2} style={{ margin: 0 }}>{getPageTitle()}</Title>}
          </div>

          <div className="page-content-wrapper" style={{ animation: 'contentFadeIn 0.4s ease-out' }}>
            {children}
          </div>
        </Content>
      </Layout>

      <style>
        {`
          .app-sider-shell {
            background: var(--app-nav-bg);
            color: var(--app-nav-text);
          }

          .app-sider-brand {
            min-height: 72px;
            display: flex;
            align-items: center;
            justify-content: flex-start;
            padding: 14px 20px 10px;
            overflow: hidden;
            transition: all 0.2s ease;
          }

          .app-sider-brand--collapsed {
            justify-content: center;
            padding-inline: 12px;
          }

          .app-sider-brand-logo {
            width: 124px;
            height: 40px;
            object-fit: contain;
            flex-shrink: 0;
            cursor: pointer;
          }

          .app-sider-brand--collapsed .app-sider-brand-logo {
            width: 38px;
            height: 38px;
          }

          .app-sider-shell .ant-menu {
            background: transparent !important;
          }

          .app-sider-shell .ant-menu-item,
          .app-sider-shell .ant-menu-submenu-title {
            color: var(--app-nav-text) !important;
            margin-bottom: 8px;
          }

          .app-sider-shell .ant-menu-submenu-selected > .ant-menu-submenu-title,
          .app-sider-shell .ant-menu-item:hover,
          .app-sider-shell .ant-menu-submenu-title:hover {
            color: var(--app-nav-text-strong) !important;
            background: var(--app-nav-item-hover) !important;
            border-radius: 10px !important;
          }

          .desktop-sider .ant-menu-item-selected {
            background: var(--app-nav-item-active) !important;
            border-radius: 8px !important;
            width: calc(100% - 16px) !important;
            margin-left: 8px !important;
          }

          .desktop-sider .ant-menu-item-selected::after {
            border-inline-end-color: transparent !important;
          }

          .main-layout .ant-layout-header,
          .main-layout .ant-layout-content,
          .main-layout .ant-breadcrumb,
          .main-layout .ant-select,
          .main-layout .ant-typography {
            color: var(--app-panel-text);
          }

          .main-layout .ant-breadcrumb a,
          .main-layout .ant-breadcrumb-separator {
            color: var(--app-panel-text-muted);
          }

          .main-layout .ant-select-selector,
          .main-layout .ant-btn:not(.ant-btn-primary):not(.ant-btn-link) {
            background: var(--app-panel-bg);
            border-color: var(--app-shell-border);
          }

          .main-layout .ant-card,
          .main-layout .ant-table-wrapper .ant-table,
          .main-layout .ant-table-wrapper .ant-table-container,
          .main-layout .ant-collapse,
          .main-layout .ant-tabs-nav,
          .main-layout .ant-list,
          .main-layout .ant-descriptions-view {
            background: var(--app-panel-bg);
            border-color: var(--app-shell-border);
            color: var(--app-panel-text);
          }

          .main-layout .ant-table-wrapper .ant-table-thead > tr > th,
          .main-layout .ant-table-wrapper .ant-table-tbody > tr > td,
          .main-layout .ant-collapse > .ant-collapse-item > .ant-collapse-header,
          .main-layout .ant-collapse-content-box,
          .main-layout .ant-tabs-tab,
          .main-layout .ant-card .ant-card-head,
          .main-layout .ant-card .ant-card-body {
            background: transparent;
            color: var(--app-panel-text);
            border-color: var(--app-shell-border);
          }

          .main-layout .ant-table-wrapper .ant-table-tbody > tr.ant-table-row:hover > td,
          .main-layout .ant-tabs-tab:hover,
          .main-layout .ant-list-item:hover {
            background: var(--app-panel-muted-bg);
          }

          .main-layout .ant-btn-text {
            border-color: transparent !important;
            background: transparent !important;
            color: var(--app-panel-text);
          }

          .main-layout .ant-btn-text:hover {
            background: var(--app-panel-muted-bg) !important;
          }

          @keyframes contentFadeIn {
            from { opacity: 0; transform: translateY(10px); }
            to { opacity: 1; transform: translateY(0); }
          }
          @media (max-width: 991px) {
            .desktop-sider { display: none !important; }
            .main-layout { margin-left: 0 !important; }
          }
          @media (max-width: 576px) {
            .main-layout .ant-layout-content { padding: 16px !important; }
          }
        `}
      </style>
    </Layout>
  );
};

export default AppLayout;
