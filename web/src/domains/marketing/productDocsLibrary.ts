import i18n, { getLocalizedPath, resolveSupportedLanguage } from '../../i18n/config';
import { DIRECT_DOCS_LIBRARIES, LOCALIZED_DOCS_ARTICLES, LOCALIZED_DOCS_META } from './docs/locales';
import type { DocsLibrary, DocsLink } from './docs/schema';

export type {
  DocsArticle,
  DocsArticleSection,
  DocsCodeBlock,
  DocsLibrary,
  DocsLink,
  DocsSection,
  DocsStep,
} from './docs/schema';

export const DOCS_REPO_URL = 'https://github.com/12345dotblue/dotblue';

function buildQuickLinks(language: string): DocsLink[] {
  const docsT = i18n.getFixedT(resolveSupportedLanguage(language));
  return [
    { label: docsT('docs_quick_link_home'), url: getLocalizedPath('/', language) },
    { label: docsT('docs_quick_link_login'), url: getLocalizedPath('/login', language) },
    { label: docsT('docs_github_label'), url: DOCS_REPO_URL, description: docsT('docs_quick_link_repo_description') },
  ];
}

function applyLocalizedDocsMeta(base: DocsLibrary, language: string): DocsLibrary {
  const meta = LOCALIZED_DOCS_META[language];
  if (!meta) return base;
  return {
    ...base,
    sections: base.sections.map((section) => {
      const sectionMeta = meta.sections[section.slug];
      return {
        ...section,
        title: sectionMeta?.title || section.title,
        description: sectionMeta?.description || section.description,
        articles: section.articles.map((article) => {
          const articleMeta = meta.articles[article.slug];
          if (!articleMeta) return article;
          return {
            ...article,
            title: articleMeta.title,
            summary: articleMeta.summary,
            seoTitle: `${articleMeta.title} | dotblue Docs`,
            seoDescription: articleMeta.summary,
          };
        }),
      };
    }),
  };
}

function applyLocalizedDocsContent(base: DocsLibrary, language: string): DocsLibrary {
  const articleMap = LOCALIZED_DOCS_ARTICLES[language];
  if (!articleMap) return base;
  return {
    ...base,
    sections: base.sections.map((section) => ({
      ...section,
      articles: section.articles.map((article) => {
        const localizedArticle = articleMap[article.slug];
        if (!localizedArticle) return article;
        return {
          ...localizedArticle,
          contentLanguage: language,
          isFallbackContent: false,
        };
      }),
    })),
  };
}

function finalizeDocsLocalizationState(library: DocsLibrary, language: string): DocsLibrary {
  const hasFallbackArticle = library.sections.some((section) =>
    section.articles.some((article) => article.isFallbackContent ?? true),
  );
  if (hasFallbackArticle) {
    return library;
  }
  return {
    ...library,
    contentLanguage: language,
    isFallbackContent: false,
  };
}

export function getDocsLibrary(language: string): DocsLibrary {
  const resolved = resolveSupportedLanguage(language);
  const direct = DIRECT_DOCS_LIBRARIES[resolved];
  if (direct) {
    return {
      ...direct,
      requestedLanguage: resolved,
      contentLanguage: direct.contentLanguage,
      isFallbackContent: false,
    };
  }
  const docsT = i18n.getFixedT(resolved);
  const fallbackLibrary = {
    ...DIRECT_DOCS_LIBRARIES.en,
    requestedLanguage: resolved,
    contentLanguage: 'en',
    isFallbackContent: true,
    homeSeoTitle: docsT('docs_fallback_home_seo_title'),
    homeSeoDescription: docsT('docs_fallback_home_seo_description'),
    homeSeoKeywords: docsT('docs_fallback_home_seo_keywords'),
    eyebrow: docsT('docs_fallback_eyebrow'),
    title: docsT('docs_fallback_title'),
    subtitle: docsT('docs_fallback_subtitle'),
    categoriesLabel: docsT('docs_fallback_categories_label'),
    popularLabel: docsT('docs_fallback_popular_label'),
    sectionDescriptionLabel: docsT('docs_fallback_section_description_label'),
    quickLinksTitle: docsT('docs_fallback_quick_links_title'),
    repoTitle: docsT('docs_fallback_repo_title'),
    repoDescription: docsT('docs_fallback_repo_description'),
    quickLinks: buildQuickLinks(resolved),
  };
  return finalizeDocsLocalizationState(
    applyLocalizedDocsContent(applyLocalizedDocsMeta(fallbackLibrary, resolved), resolved),
    resolved,
  );
}

export function flattenDocsArticles(library: DocsLibrary) {
  return library.sections.flatMap((section) => section.articles.map((article) => ({ ...article, sectionTitle: section.title })));
}

export function findDocsArticle(language: string, sectionSlug: string, articleSlug: string) {
  const library = getDocsLibrary(language);
  for (const section of library.sections) {
    if (section.slug !== sectionSlug) continue;
    for (const article of section.articles) {
      if (article.slug === articleSlug) {
        return { section, article };
      }
    }
  }
  return null;
}

export function getDocsHref(language: string, sectionSlug?: string, articleSlug?: string) {
  if (!sectionSlug || !articleSlug) {
    return getLocalizedPath('/docs', language);
  }
  return getLocalizedPath(`/docs/${sectionSlug}/${articleSlug}`, language);
}
