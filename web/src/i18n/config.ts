import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';
import { TRANSLATION_RESOURCES as resources } from './locales';

export const LANGUAGE_OPTIONS = [
  { value: 'en', label: 'English', shortLabel: 'EN' },
  { value: 'zh-CN', label: '简体中文', shortLabel: '中文' },
  { value: 'ja', label: '日本語', shortLabel: 'JA' },
  { value: 'ko', label: '한국어', shortLabel: 'KO' },
  { value: 'fr', label: 'Français', shortLabel: 'FR' },
  { value: 'es', label: 'Español', shortLabel: 'ES' },
] as const;

export const SUPPORTED_LANGUAGES = LANGUAGE_OPTIONS.map((item) => item.value);
export const SEO_BASE_URL = 'https://dotblue.ai';
export const DEFAULT_LANGUAGE = 'en';

export function resolveSupportedLanguage(language?: string) {
  if (!language) {
    return DEFAULT_LANGUAGE;
  }

  if (SUPPORTED_LANGUAGES.includes(language as typeof SUPPORTED_LANGUAGES[number])) {
    return language;
  }

  const normalized = language.toLowerCase();
  if (normalized.startsWith('zh')) {
    return 'zh-CN';
  }

  const base = normalized.split('-')[0];
  return SUPPORTED_LANGUAGES.find((item) => item === base) || DEFAULT_LANGUAGE;
}

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    supportedLngs: [...SUPPORTED_LANGUAGES],
    fallbackLng: DEFAULT_LANGUAGE,
    load: 'currentOnly',
    nonExplicitSupportedLngs: false,
    detection: {
      order: ['querystring', 'path', 'localStorage', 'navigator'],
      lookupQuerystring: 'lng',
      lookupFromPathIndex: 0,
      caches: ['localStorage'],
    },
    interpolation: {
      escapeValue: false
    }
  });

export async function applyLanguagePreference(language: string) {
  const resolved = resolveSupportedLanguage(language);

  try {
    localStorage.setItem('i18nextLng', resolved);
  } catch {
    // Ignore storage failures and still attempt to switch language in memory.
  }

  await i18n.changeLanguage(resolved);
  syncLanguagePath(resolved);
  return resolved;
}

export function getLanguageFromPath(pathname: string) {
  const firstSegment = pathname.split('/').filter(Boolean)[0];
  if (!firstSegment) {
    return null;
  }

  return SUPPORTED_LANGUAGES.includes(firstSegment as typeof SUPPORTED_LANGUAGES[number]) ? firstSegment : null;
}

export function stripLanguagePrefix(pathname: string) {
  const segments = pathname.split('/').filter(Boolean);
  if (segments.length === 0) {
    return '/';
  }

  const [firstSegment, ...rest] = segments;
  if (!SUPPORTED_LANGUAGES.includes(firstSegment as typeof SUPPORTED_LANGUAGES[number])) {
    return pathname || '/';
  }

  return rest.length > 0 ? `/${rest.join('/')}` : '/';
}

export function getLocalizedPath(path: string, language: string) {
  const resolved = resolveSupportedLanguage(language);
  const url = new URL(path, SEO_BASE_URL);
  const strippedPath = stripLanguagePrefix(url.pathname);
  const normalizedPath = strippedPath === '/' ? '' : strippedPath;
  return `/${resolved}${normalizedPath}${url.search}${url.hash}`;
}

export function buildLocalizedUrl(path: string, language: string) {
  return new URL(getLocalizedPath(path, language), SEO_BASE_URL).toString();
}

export function getPreferredLanguage(language?: string) {
  if (language) {
    return resolveSupportedLanguage(language);
  }

  if (typeof window !== 'undefined') {
    const pathLanguage = getLanguageFromPath(window.location.pathname);
    if (pathLanguage) {
      return pathLanguage;
    }
  }

  try {
    const storedLanguage = localStorage.getItem('i18nextLng');
    if (storedLanguage) {
      return resolveSupportedLanguage(storedLanguage);
    }
  } catch {
    // Ignore storage failures and fall back to runtime language.
  }

  return resolveSupportedLanguage(i18n.resolvedLanguage || i18n.language);
}

export function syncLanguagePath(language: string) {
  if (typeof window === 'undefined') {
    return;
  }

  const resolved = resolveSupportedLanguage(language);
  const url = new URL(window.location.href);
  url.searchParams.delete('lng');
  const localizedPath = getLocalizedPath(`${stripLanguagePrefix(url.pathname)}${url.search}${url.hash}`, resolved);
  window.history.replaceState(null, '', localizedPath);
}

export default i18n;
