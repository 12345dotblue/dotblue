import { library as en } from './en';
import { library as zhCN } from './zh-CN';
import { localizedMeta as jaMeta, localizedArticles as jaArticles } from './ja';
import { localizedMeta as koMeta, localizedArticles as koArticles } from './ko';
import { localizedMeta as frMeta, localizedArticles as frArticles } from './fr';
import { localizedMeta as esMeta, localizedArticles as esArticles } from './es';
import type { DocsArticle, DocsLibrary, LocalizedDocsMeta } from '../schema';

export const DIRECT_DOCS_LIBRARIES: Record<string, DocsLibrary> = {
  en,
  'zh-CN': zhCN,
};

export const LOCALIZED_DOCS_META: Record<string, LocalizedDocsMeta> = {
  ja: jaMeta,
  ko: koMeta,
  fr: frMeta,
  es: esMeta,
};

export const LOCALIZED_DOCS_ARTICLES: Record<string, Record<string, DocsArticle>> = {
  ja: jaArticles,
  ko: koArticles,
  fr: frArticles,
  es: esArticles,
};
