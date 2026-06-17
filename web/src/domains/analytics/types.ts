export interface PageViewParams {
  path: string;
  title: string;
  language: string;
  referrer?: string;
}

export interface EventParams {
  [key: string]: string | number | boolean | undefined;
}

export interface AnalyticsAdapter {
  init: (projectId: string) => void;
  trackPageView: (params: PageViewParams) => void;
  trackEvent: (eventName: string, params?: EventParams) => void;
  identify?: (userId: string, traits?: Record<string, unknown>) => void;
}
