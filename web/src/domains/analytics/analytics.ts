import type { AnalyticsAdapter, EventParams, PageViewParams } from './types';

let adapter: AnalyticsAdapter | null = null;
let enabled = false;

export function initAnalytics(projectId: string) {
  if (!projectId) {
    return;
  }

  import('./clarity-adapter').then((mod) => {
    adapter = mod.createClarityAdapter();
    adapter.init(projectId);
    enabled = true;
  });
}

export function trackPageView(params: PageViewParams) {
  if (!enabled || !adapter) return;
  adapter.trackPageView(params);
}

export function trackEvent(eventName: string, params?: EventParams) {
  if (!enabled || !adapter) return;
  adapter.trackEvent(eventName, params);
}

export function identify(userId: string, traits?: Record<string, unknown>) {
  if (!enabled || !adapter?.identify) return;
  adapter.identify(userId, traits);
}

export function isEnabled() {
  return enabled;
}
