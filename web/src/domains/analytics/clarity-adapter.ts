import type { AnalyticsAdapter, EventParams, PageViewParams } from './types';

declare global {
  interface Window {
    clarity: (...args: unknown[]) => void;
  }
}

function loadClarityScript(projectId: string) {
  (function (c: Document, l: string, a: string, _r: string, i: string) {
    const w = window as unknown as Record<string, unknown>;
    w[a] = w[a] || function () {
      const q = ((w[a] as Record<string, unknown>)?.q as unknown[]) || [];
      q.push(arguments);
      ((w[a] as Record<string, unknown>).q as unknown[]) = q;
    };
    const script = c.createElement(l) as HTMLScriptElement;
    script.async = true;
    script.src = `https://www.clarity.ms/tag/${i}`;
    const firstScript = c.getElementsByTagName(l)[0];
    firstScript.parentNode?.insertBefore(script, firstScript);
  })(document, 'script', 'clarity', 'script', projectId);
}

export function createClarityAdapter(): AnalyticsAdapter {
  return {
    init(projectId: string) {
      loadClarityScript(projectId);
    },

    trackPageView(params: PageViewParams) {
      window.clarity('set', 'page_language', params.language);
      if (params.referrer) {
        window.clarity('set', 'referrer', params.referrer);
      }
    },

    trackEvent(eventName: string, params?: EventParams) {
      window.clarity('event', eventName, params);
    },

    identify(userId: string, traits?: Record<string, unknown>) {
      window.clarity('identify', userId, traits);
    },
  };
}
