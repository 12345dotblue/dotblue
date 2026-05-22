function resolveBackendUrl() {
  const configuredBackendUrl = import.meta.env.VITE_BACKEND_URL?.trim();
  const isLocalBackendUrl = configuredBackendUrl === 'http://localhost:8000'
    || configuredBackendUrl === 'http://127.0.0.1:8000';

  // In local Vite dev, prefer same-origin `/api` unless the developer
  // explicitly points to a non-local backend endpoint.
  if (import.meta.env.DEV && (!configuredBackendUrl || isLocalBackendUrl)) {
    return '';
  }

  if (configuredBackendUrl) {
    return configuredBackendUrl;
  }

  return 'http://localhost:8000';
}

export const BACKEND_URL = resolveBackendUrl();
