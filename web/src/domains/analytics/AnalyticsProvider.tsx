import { useEffect, useRef } from 'react';
import { useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { trackPageView } from './analytics';

export function AnalyticsProvider() {
  const location = useLocation();
  const { i18n } = useTranslation();
  const previousPath = useRef(location.pathname);

  useEffect(() => {
    if (location.pathname !== previousPath.current) {
      previousPath.current = location.pathname;

      trackPageView({
        path: location.pathname + location.search,
        title: document.title,
        language: i18n.language,
      });
    }
  }, [location, i18n.language]);

  return null;
}
