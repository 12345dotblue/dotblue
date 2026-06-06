export type DocsLink = {
  label: string;
  url: string;
  description?: string;
};

export type DocsStep = {
  title: string;
  desc: string;
};

export type DocsCodeBlock = {
  language: string;
  value: string;
};

export type DocsArticleSection = {
  id: string;
  title: string;
  intro?: string;
  paragraphs?: string[];
  bullets?: string[];
  steps?: DocsStep[];
  code?: DocsCodeBlock;
  links?: DocsLink[];
  note?: string;
};

export type DocsArticle = {
  sectionSlug: string;
  slug: string;
  title: string;
  summary: string;
  seoTitle: string;
  seoDescription: string;
  readingTime: string;
  sections: DocsArticleSection[];
  contentLanguage?: string;
  isFallbackContent?: boolean;
};

export type DocsSection = {
  slug: string;
  title: string;
  description: string;
  articles: DocsArticle[];
};

export type DocsLibrary = {
  requestedLanguage: string;
  contentLanguage: string;
  isFallbackContent: boolean;
  homeSeoTitle: string;
  homeSeoDescription: string;
  homeSeoKeywords: string;
  eyebrow: string;
  title: string;
  subtitle: string;
  categoriesLabel: string;
  popularLabel: string;
  sectionDescriptionLabel: string;
  quickLinksTitle: string;
  quickLinks: DocsLink[];
  repoTitle: string;
  repoDescription: string;
  sections: DocsSection[];
};

export type DocsSectionMeta = {
  title: string;
  description: string;
};

export type DocsArticleMeta = {
  title: string;
  summary: string;
};

export type LocalizedDocsMeta = {
  sections: Record<string, DocsSectionMeta>;
  articles: Record<string, DocsArticleMeta>;
};
