import React from 'react'
import { BrowserRouter, Routes, Route, Navigate, Outlet, useLocation, useParams } from 'react-router-dom'
import { HelmetProvider, Helmet } from 'react-helmet-async'
import { ConfigProvider, theme } from 'antd'
import enUS from 'antd/locale/en_US'
import zhCN from 'antd/locale/zh_CN'
import jaJP from 'antd/locale/ja_JP'
import koKR from 'antd/locale/ko_KR'
import frFR from 'antd/locale/fr_FR'
import esES from 'antd/locale/es_ES'
import { useTranslation } from 'react-i18next'
import './i18n/config'
import { getLocalizedPath, getPreferredLanguage, resolveSupportedLanguage } from './i18n/config'

import Login from './domains/identity/Login'
import LoginCallback from './domains/identity/LoginCallback'
import { casdoorService } from './domains/identity/CasdoorService'
import AgentList from './domains/agent/AgentList'
import ChatPage from './domains/chat/ChatPage'
import AdminSettings from './domains/admin/AdminSettings'
import PlatformSettingsPage from './domains/admin/PlatformSettingsPage'
import PlatformSkillsPage from './domains/admin/PlatformSkillsPage'
import PlatformSkillBuilderPage from './domains/admin/PlatformSkillBuilderPage'
import InviteAcceptPage from './domains/admin/InviteAcceptPage'
import SetupWizard from './domains/setup/SetupWizard'
import AgentSkillManagementPage from './domains/agent/AgentSkillManagementPage'

import LandingLayout from './components/Layouts/LandingLayout'
import AppLayout from './components/Layouts/AppLayout'
import LandingPage from './domains/marketing/LandingPage'
import ProductDocsPage from './domains/marketing/ProductDocsPage'
import Terms from './domains/legal/Terms'
import Privacy from './domains/legal/Privacy'
import Refund from './domains/legal/Refund'

import { BACKEND_URL } from './config'

const ANTD_LOCALES = {
  en: enUS,
  'zh-CN': zhCN,
  ja: jaJP,
  ko: koKR,
  fr: frFR,
  es: esES,
} as const

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const { lng } = useParams()
  const language = resolveSupportedLanguage(lng || getPreferredLanguage())
  return casdoorService.isAuthenticated() ? <>{children}</> : <Navigate to={getLocalizedPath('/login', language)} replace />
}

function PlatformAdminRoute({ children }: { children: React.ReactNode }) {
  const { lng } = useParams()
  const language = resolveSupportedLanguage(lng || getPreferredLanguage())
  if (!casdoorService.isAuthenticated()) return <Navigate to={getLocalizedPath('/login', language)} replace />
  if (!casdoorService.isAdmin()) return <Navigate to={getLocalizedPath('/admin/enterprise', language)} replace />
  return <>{children}</>
}

// Cached setup status — only fetch once per session
let setupStatusCache: boolean | null = null

// Check if platform is initialized; redirect to /setup if not
function SetupGuard({ children }: { children: React.ReactNode }) {
  const [checked, setChecked] = React.useState(setupStatusCache !== null)
  const [initialized, setInitialized] = React.useState(setupStatusCache === true)
  const { lng } = useParams()
  const language = resolveSupportedLanguage(lng || getPreferredLanguage())

  React.useEffect(() => {
    if (setupStatusCache !== null) return

    fetch(`${BACKEND_URL}/api/setup/status`)
      .then(res => res.json())
      .then(data => {
        const init = data.initialized === true
        setupStatusCache = init
        setInitialized(init)
        setChecked(true)
      })
      .catch(() => {
        setupStatusCache = true
        setInitialized(true)
        setChecked(true)
      })
  }, [])

  if (!checked) return null
  if (!initialized) return <Navigate to={getLocalizedPath('/setup', language)} replace />
  return <>{children}</>
}

function RedirectToLocalized({ path }: { path?: string }) {
  const location = useLocation()
  const language = getPreferredLanguage()
  const targetPath = path || location.pathname
  const params = new URLSearchParams(location.search)
  params.delete('lng')
  const search = params.toString()
  return <Navigate to={`${getLocalizedPath(targetPath, language)}${search ? `?${search}` : ''}${location.hash}`} replace />
}

export function LocalizedRouteGuard() {
  const { lng } = useParams()
  const location = useLocation()
  const { i18n } = useTranslation()
  const resolved = resolveSupportedLanguage(lng || getPreferredLanguage())
  const currentLanguage = resolveSupportedLanguage(i18n.resolvedLanguage || i18n.language)

  React.useEffect(() => {
    if (currentLanguage === resolved) {
      return
    }

    try {
      localStorage.setItem('i18nextLng', resolved)
    } catch {
      // Ignore storage failures and still update in-memory language state.
    }

    void i18n.changeLanguage(resolved)
  }, [currentLanguage, i18n, resolved])

  if (lng !== resolved) {
    const segments = location.pathname.split('/').filter(Boolean)
    const restPath = segments.length > 1 ? `/${segments.slice(1).join('/')}` : '/'
    const params = new URLSearchParams(location.search)
    params.delete('lng')
    const search = params.toString()
    return <Navigate to={`${getLocalizedPath(restPath, resolved)}${search ? `?${search}` : ''}${location.hash}`} replace />
  }

  // Delay child route rendering until the route language has been applied,
  // otherwise pages can paint once with a stale locale and stay inconsistent.
  if (currentLanguage !== resolved) {
    return null
  }

  return <Outlet />
}

function LocalizedFallback() {
  const { lng } = useParams()
  const language = resolveSupportedLanguage(lng || getPreferredLanguage())
  return <Navigate to={getLocalizedPath('/', language)} replace />
}

function App() {
  const { t, i18n } = useTranslation()
  const currentLanguage = resolveSupportedLanguage(i18n.resolvedLanguage || i18n.language) as keyof typeof ANTD_LOCALES

  React.useEffect(() => {
    document.documentElement.lang = currentLanguage
  }, [currentLanguage])

  return (
    <HelmetProvider>
      <ConfigProvider
        locale={ANTD_LOCALES[currentLanguage]}
        modal={{
          closable: {
            'aria-label': t('common_close'),
          },
        }}
        drawer={{
          closable: {
            'aria-label': t('common_close'),
          },
        }}
        theme={{
          algorithm: theme.defaultAlgorithm,
          token: {
            colorPrimary: '#1677ff',
            borderRadius: 12,
            fontFamily: '"Inter", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
            colorBgLayout: '#f4f7f9',
          },
          components: {
            Layout: {
              headerBg: 'rgba(255, 255, 255, 0.7)',
              siderBg: '#001529',
            },
            Card: {
              boxShadowTertiary: '0 4px 20px rgba(0,0,0,0.03)',
            },
            Button: {
              controlHeight: 40,
              paddingContentHorizontal: 20,
            },
          },
        }}
      >
        <BrowserRouter>
          <Helmet>
            <title>{t('app_name')}</title>
            <meta name="description" content={t('description')} />
          </Helmet>
          <Routes>
            <Route path="/callback" element={<LoginCallback />} />
            <Route path="/" element={<RedirectToLocalized path="/" />} />
            <Route path="/docs" element={<RedirectToLocalized path="/docs" />} />
            <Route path="/docs/:sectionSlug/:docSlug" element={<RedirectToLocalized />} />
            <Route path="/terms" element={<RedirectToLocalized path="/terms" />} />
            <Route path="/privacy" element={<RedirectToLocalized path="/privacy" />} />
            <Route path="/refund" element={<RedirectToLocalized path="/refund" />} />
            <Route path="/setup" element={<RedirectToLocalized path="/setup" />} />
            <Route path="/login" element={<RedirectToLocalized path="/login" />} />
            <Route path="/invite/:code" element={<RedirectToLocalized />} />
            <Route path="/dashboard" element={<RedirectToLocalized path="/dashboard" />} />
            <Route path="/chat" element={<RedirectToLocalized path="/chat" />} />
            <Route path="/admin/settings" element={<RedirectToLocalized path="/admin/enterprise" />} />
            <Route path="/admin/enterprise" element={<RedirectToLocalized path="/admin/enterprise" />} />
            <Route path="/admin/platform" element={<RedirectToLocalized path="/admin/platform" />} />
            <Route path="/admin/platform/skills" element={<RedirectToLocalized path="/admin/platform/skills" />} />
            <Route path="/admin/platform/skill-market" element={<RedirectToLocalized path="/admin/platform/skill-market" />} />
            <Route path="/admin/platform/skills/new" element={<RedirectToLocalized path="/admin/platform/skills/new" />} />
            <Route path="/dashboard/agents/:agentId/skills" element={<RedirectToLocalized />} />

            <Route path="/:lng" element={<LocalizedRouteGuard />}>
              <Route index element={<LandingLayout><LandingPage /></LandingLayout>} />
              <Route path="docs" element={<LandingLayout><ProductDocsPage /></LandingLayout>} />
              <Route path="docs/:sectionSlug/:docSlug" element={<LandingLayout><ProductDocsPage /></LandingLayout>} />
              <Route path="terms" element={<LandingLayout><Terms /></LandingLayout>} />
              <Route path="privacy" element={<LandingLayout><Privacy /></LandingLayout>} />
              <Route path="refund" element={<LandingLayout><Refund /></LandingLayout>} />
              <Route path="setup" element={<SetupWizard />} />
              <Route path="login" element={<Login />} />
              <Route path="invite/:code" element={<InviteAcceptPage />} />
              <Route path="admin/settings" element={<Navigate to="../enterprise" replace relative="path" />} />
              <Route
                path="dashboard"
                element={
                  <SetupGuard>
                    <PrivateRoute>
                      <AppLayout>
                        <AgentList />
                      </AppLayout>
                    </PrivateRoute>
                  </SetupGuard>
                }
              />
              <Route
                path="dashboard/agents/:agentId/skills"
                element={
                  <SetupGuard>
                    <PrivateRoute>
                      <AppLayout>
                        <AgentSkillManagementPage />
                      </AppLayout>
                    </PrivateRoute>
                  </SetupGuard>
                }
              />
              <Route
                path="chat"
                element={
                  <SetupGuard>
                    <PrivateRoute>
                      <AppLayout>
                        <ChatPage />
                      </AppLayout>
                    </PrivateRoute>
                  </SetupGuard>
                }
              />
              <Route
                path="admin/enterprise"
                element={
                  <SetupGuard>
                    <PrivateRoute>
                      <AppLayout>
                        <AdminSettings />
                      </AppLayout>
                    </PrivateRoute>
                  </SetupGuard>
                }
              />
              <Route
                path="admin/platform"
                element={
                  <SetupGuard>
                    <PlatformAdminRoute>
                      <AppLayout>
                        <PlatformSettingsPage />
                      </AppLayout>
                    </PlatformAdminRoute>
                  </SetupGuard>
                }
              />
              <Route
                path="admin/platform/skill-market"
                element={
                  <SetupGuard>
                    <PlatformAdminRoute>
                      <AppLayout>
                        <PlatformSkillsPage experience="market" />
                      </AppLayout>
                    </PlatformAdminRoute>
                  </SetupGuard>
                }
              />
              <Route
                path="admin/platform/skills/new"
                element={
                  <SetupGuard>
                    <PlatformAdminRoute>
                      <AppLayout>
                        <PlatformSkillBuilderPage />
                      </AppLayout>
                    </PlatformAdminRoute>
                  </SetupGuard>
                }
              />
              <Route
                path="admin/platform/skills"
                element={
                  <SetupGuard>
                    <PlatformAdminRoute>
                      <AppLayout>
                        <PlatformSkillsPage />
                      </AppLayout>
                    </PlatformAdminRoute>
                  </SetupGuard>
                }
              />
              <Route
                path="admin/platform/skills/create"
                element={
                  <SetupGuard>
                    <PlatformAdminRoute>
                      <AppLayout>
                        <PlatformSkillsPage experience="market" defaultTab="skills" openFlow="skill" />
                      </AppLayout>
                    </PlatformAdminRoute>
                  </SetupGuard>
                }
              />
              <Route
                path="admin/platform/skills/import"
                element={
                  <SetupGuard>
                    <PlatformAdminRoute>
                      <AppLayout>
                        <PlatformSkillsPage experience="market" defaultTab="imports" openFlow="import" />
                      </AppLayout>
                    </PlatformAdminRoute>
                  </SetupGuard>
                }
              />
              <Route
                path="admin/platform/skill-hubs"
                element={
                  <SetupGuard>
                    <PlatformAdminRoute>
                      <AppLayout>
                        <PlatformSkillsPage defaultTab="hubs" />
                      </AppLayout>
                    </PlatformAdminRoute>
                  </SetupGuard>
                }
              />
              <Route path="*" element={<LocalizedFallback />} />
            </Route>

            <Route path="*" element={<RedirectToLocalized path="/" />} />
          </Routes>
        </BrowserRouter>
      </ConfigProvider>
    </HelmetProvider>
  )
}

export default App
