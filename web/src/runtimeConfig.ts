type RuntimeConfigSource = Partial<{
  VITE_CASDOOR_SERVER_URL: string;
  VITE_CASDOOR_CLIENT_ID: string;
  VITE_CASDOOR_ORG_NAME: string;
  VITE_CASDOOR_APP_NAME: string;
  VITE_CASDOOR_REDIRECT_URL: string;
  VITE_BACKEND_URL: string;
}>;

declare global {
  interface Window {
    __DOTBLUE_CONFIG__?: RuntimeConfigSource;
  }
}

function readValue(runtimeValue: string | undefined, buildValue: string | undefined) {
  const trimmedRuntimeValue = runtimeValue?.trim();
  if (trimmedRuntimeValue) {
    return trimmedRuntimeValue;
  }

  const trimmedBuildValue = buildValue?.trim();
  if (trimmedBuildValue) {
    return trimmedBuildValue;
  }

  return '';
}

const runtimeConfigSource = typeof window === 'undefined'
  ? {}
  : (window.__DOTBLUE_CONFIG__ ?? {});

export const runtimeConfig = {
  casdoorServerUrl: readValue(
    runtimeConfigSource.VITE_CASDOOR_SERVER_URL,
    import.meta.env.VITE_CASDOOR_SERVER_URL,
  ),
  casdoorClientId: readValue(
    runtimeConfigSource.VITE_CASDOOR_CLIENT_ID,
    import.meta.env.VITE_CASDOOR_CLIENT_ID,
  ),
  casdoorOrgName: readValue(
    runtimeConfigSource.VITE_CASDOOR_ORG_NAME,
    import.meta.env.VITE_CASDOOR_ORG_NAME,
  ),
  casdoorAppName: readValue(
    runtimeConfigSource.VITE_CASDOOR_APP_NAME,
    import.meta.env.VITE_CASDOOR_APP_NAME,
  ),
  casdoorRedirectUrl: readValue(
    runtimeConfigSource.VITE_CASDOOR_REDIRECT_URL,
    import.meta.env.VITE_CASDOOR_REDIRECT_URL,
  ),
  backendUrl: readValue(
    runtimeConfigSource.VITE_BACKEND_URL,
    import.meta.env.VITE_BACKEND_URL,
  ),
};
