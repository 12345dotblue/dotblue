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
  const username = casdoorService.getUsername() || 'User';
  const currentLanguage = resolveSupportedLanguage(i18n.resolvedLanguage || i18n.language);
  const currentLanguageLabel = LANGUAGE_OPTIONS.find((item) => item.value === currentLanguage)?.shortLabel || 'EN';

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
    if (resolveSupportedLanguage(i18n.resolvedLanguage || i18n.language) !== resolved) {
      window.location.reload();
    }
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
            key: '/admin/platform/skills',
            label: t('platform_skills_title'),
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
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%', background: '#001529' }}>
      <div style={{
        height: 64,
        display: 'flex',
        alignItems: 'center',
        padding: '0 24px',
        justifyContent: collapsed && !mobileVisible ? 'center' : 'flex-start',
        overflow: 'hidden',
        transition: 'all 0.2s',
      }}>
        <img
          src="/brand/dotblue-logo.png"
          alt="dotblue"
          style={{
            width: collapsed && !mobileVisible ? 32 : 92,
            height: 32,
            objectFit: 'contain',
            flexShrink: 0,
          }}
        />
        {(!collapsed || mobileVisible) && (
          <Title level={4} style={{ margin: '0 0 0 12px', color: '#fff', letterSpacing: 1, whiteSpace: 'nowrap' }}>
            dotblue
          </Title>
        )}
      </div>
      <Menu
        theme="dark"
        mode="inline"
        inlineCollapsed={collapsed}
        selectedKeys={[normalizedPath]}
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
        theme="dark"
        className="desktop-sider"
        width={240}
        style={{
          position: 'fixed',
          height: '100vh',
          left: 0,
          top: 0,
          bottom: 0,
          boxShadow: '4px 0 10px rgba(0,0,0,0.1)',
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
        background: '#f4f7f9',
      }} className="main-layout">
        <Header style={{
          background: 'rgba(255, 255, 255, 0.7)',
          backdropFilter: 'blur(12px)',
          padding: '0 24px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          borderBottom: '1px solid rgba(0,0,0,0.05)',
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
            <Dropdown menu={{
              items: LANGUAGE_OPTIONS.map((item) => ({
                key: item.value,
                label: item.label,
              })),
              onClick: ({ key }) => changeLanguage(String(key)),
            }}>
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

        <Content style={{ padding: '24px', maxWidth: 1400, margin: '0 auto', width: '100%' }}>
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
          .desktop-sider .ant-menu-item-selected {
            background: rgba(22, 119, 255, 0.15) !important;
            border-radius: 8px !important;
            width: calc(100% - 16px) !important;
            margin-left: 8px !important;
          }
          .ant-menu-dark .ant-menu-item {
             margin-bottom: 8px;
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
