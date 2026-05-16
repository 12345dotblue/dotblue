import React from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { HelmetProvider, Helmet } from 'react-helmet-async'
import { ConfigProvider, theme } from 'antd'
import { useTranslation } from 'react-i18next'
import './i18n/config'

import Login from './domains/identity/Login'
import LoginCallback from './domains/identity/LoginCallback'
import { casdoorService } from './domains/identity/CasdoorService'
import AgentList from './domains/agent/AgentList'
import ChatPage from './domains/chat/ChatPage'
import AdminSettings from './domains/admin/AdminSettings'
import PlatformSettingsPage from './domains/admin/PlatformSettingsPage'
import InviteAcceptPage from './domains/admin/InviteAcceptPage'
import SetupWizard from './domains/setup/SetupWizard'

import LandingLayout from './components/Layouts/LandingLayout'
import AppLayout from './components/Layouts/AppLayout'
import LandingPage from './domains/marketing/LandingPage'
import Terms from './domains/legal/Terms'
import Privacy from './domains/legal/Privacy'
import Refund from './domains/legal/Refund'

import { BACKEND_URL } from './config'

function PrivateRoute({ children }: { children: React.ReactNode }) {
  return casdoorService.isAuthenticated() ? <>{children}</> : <Navigate to="/login" replace />
}

function PlatformAdminRoute({ children }: { children: React.ReactNode }) {
  if (!casdoorService.isAuthenticated()) return <Navigate to="/login" replace />
  if (!casdoorService.isAdmin()) return <Navigate to="/admin/enterprise" replace />
  return <>{children}</>
}

// Cached setup status — only fetch once per session
let setupStatusCache: boolean | null = null

// Check if platform is initialized; redirect to /setup if not
function SetupGuard({ children }: { children: React.ReactNode }) {
  const [checked, setChecked] = React.useState(setupStatusCache !== null)
  const [initialized, setInitialized] = React.useState(setupStatusCache === true)

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
  if (!initialized) return <Navigate to="/setup" replace />
  return <>{children}</>
}

function App() {
  const { t } = useTranslation()

  return (
    <HelmetProvider>
      <ConfigProvider
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
            }
          },
        }}
      >
        <BrowserRouter>
          <Helmet>
            <title>{t('app_name')}</title>
            <meta name="description" content={t('description')} />
          </Helmet>
          <Routes>
            <Route path="/" element={<LandingLayout><LandingPage /></LandingLayout>} />
            <Route path="/terms" element={<LandingLayout><Terms /></LandingLayout>} />
            <Route path="/privacy" element={<LandingLayout><Privacy /></LandingLayout>} />
            <Route path="/refund" element={<LandingLayout><Refund /></LandingLayout>} />

            <Route path="/setup" element={<SetupWizard />} />
            <Route path="/login" element={<Login />} />
            <Route path="/callback" element={<LoginCallback />} />
            <Route path="/invite/:code" element={<InviteAcceptPage />} />
            <Route path="/admin/settings" element={<Navigate to="/admin/enterprise" replace />} />

            <Route
              path="/dashboard"
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
              path="/chat"
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
              path="/admin/enterprise"
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
              path="/admin/platform"
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

            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
      </ConfigProvider>
    </HelmetProvider>
  )
}

export default App
