import { useCallback } from 'react';
import { trackEvent, identify } from './analytics';
import type { EventParams } from './types';

export function useAnalytics() {
  const track = useCallback((eventName: string, params?: EventParams) => {
    trackEvent(eventName, params);
  }, []);

  const setIdentify = useCallback((userId: string, traits?: Record<string, unknown>) => {
    identify(userId, traits);
  }, []);

  return { track, identify: setIdentify };
}
