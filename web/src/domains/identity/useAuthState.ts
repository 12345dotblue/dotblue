import { useSyncExternalStore } from 'react';
import { CASDOOR_AUTH_CHANGED_EVENT, casdoorService } from './CasdoorService';

function subscribe(onStoreChange: () => void) {
  window.addEventListener('storage', onStoreChange);
  window.addEventListener('focus', onStoreChange);
  window.addEventListener('pageshow', onStoreChange);
  window.addEventListener('visibilitychange', onStoreChange);
  window.addEventListener(CASDOOR_AUTH_CHANGED_EVENT, onStoreChange);

  return () => {
    window.removeEventListener('storage', onStoreChange);
    window.removeEventListener('focus', onStoreChange);
    window.removeEventListener('pageshow', onStoreChange);
    window.removeEventListener('visibilitychange', onStoreChange);
    window.removeEventListener(CASDOOR_AUTH_CHANGED_EVENT, onStoreChange);
  };
}

function getSnapshot() {
  return casdoorService.isAuthenticated();
}

export function useAuthState() {
  return useSyncExternalStore(subscribe, getSnapshot, getSnapshot);
}
