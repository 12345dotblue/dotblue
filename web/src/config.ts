const LOCAL_WSL_BACKEND_URL = 'http://172.22.3.181:8000';
const LOCALHOST_BACKEND_URLS = new Set([
  'http://localhost:8000',
  'http://127.0.0.1:8000',
]);

function resolveBackendUrl() {
  const configuredBackendUrl = import.meta.env.VITE_BACKEND_URL?.trim();
  const runningOnLocalhost = typeof window !== 'undefined' && ['localhost', '127.0.0.1'].includes(window.location.hostname);

  if (configuredBackendUrl && (!runningOnLocalhost || !LOCALHOST_BACKEND_URLS.has(configuredBackendUrl))) {
    return configuredBackendUrl;
  }

  if (runningOnLocalhost) {
    return LOCAL_WSL_BACKEND_URL;
  }

  return 'http://localhost:8000';
}

export const BACKEND_URL = resolveBackendUrl();
