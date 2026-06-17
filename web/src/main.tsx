import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import { runtimeConfig } from './runtimeConfig.ts'
import { initAnalytics } from './domains/analytics/analytics.ts'

if (runtimeConfig.clarityProjectId) {
  initAnalytics(runtimeConfig.clarityProjectId)
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
