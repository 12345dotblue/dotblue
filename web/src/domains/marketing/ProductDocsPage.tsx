import React, { useMemo } from 'react';
import { Anchor, Breadcrumb, Button, Card, Col, Divider, Row, Space, Tag, Typography } from 'antd';
import { ArrowLeftOutlined, ArrowRightOutlined, BookOutlined, GithubOutlined, LoginOutlined, RocketOutlined } from '@ant-design/icons';
import { Helmet } from 'react-helmet-async';
import { Link, Navigate, useNavigate, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { SUPPORTED_LANGUAGES, buildLocalizedUrl, getLocalizedPath, resolveSupportedLanguage } from '../../i18n/config';
import { useAuthState } from '../identity/useAuthState';
import { DOCS_REPO_URL, findDocsArticle, flattenDocsArticles, getDocsHref, getDocsLibrary } from './productDocsLibrary';

const { Title, Paragraph, Text } = Typography;

const ProductDocsPage: React.FC = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const { sectionSlug, docSlug } = useParams();
  const isAuthenticated = useAuthState();
  const currentLanguage = resolveSupportedLanguage(i18n.resolvedLanguage || i18n.language);
  const docs = getDocsLibrary(currentLanguage);
  const articleMatch = sectionSlug && docSlug ? findDocsArticle(currentLanguage, sectionSlug, docSlug) : null;
  const article = articleMatch?.article;
  const articleSection = articleMatch?.section;
  const allArticles = useMemo(() => flattenDocsArticles(docs), [docs]);
  const currentArticleIndex = article ? allArticles.findIndex((item) => item.slug === article.slug && item.sectionSlug === article.sectionSlug) : -1;
  const previousArticle = currentArticleIndex > 0 ? allArticles[currentArticleIndex - 1] : null;
  const nextArticle = currentArticleIndex >= 0 && currentArticleIndex < allArticles.length - 1 ? allArticles[currentArticleIndex + 1] : null;
  const currentPath = article ? `/docs/${article.sectionSlug}/${article.slug}` : '/docs';
  const canonicalUrl = buildLocalizedUrl(currentPath, currentLanguage);
  const localizedHomePath = getLocalizedPath('/', currentLanguage);
  const localizedDocsHomePath = getLocalizedPath('/docs', currentLanguage);
  const tocTitle = currentLanguage === 'zh-CN' ? '本页目录' : 'On this page';
  const previousLabel = currentLanguage === 'zh-CN' ? '上一篇' : 'Previous';
  const nextLabel = currentLanguage === 'zh-CN' ? '下一篇' : 'Next';

  const anchorItems = useMemo(
    () => (article ? article.sections.map((section) => ({ key: section.id, href: `#${section.id}`, title: section.title })) : []),
    [article],
  );

  if ((sectionSlug || docSlug) && !article) {
    return <Navigate to={localizedDocsHomePath} replace />;
  }

  const breadcrumbJsonLd = {
    '@context': 'https://schema.org',
    '@type': 'BreadcrumbList',
    itemListElement: [
      { '@type': 'ListItem', position: 1, name: t('welcome'), item: buildLocalizedUrl('/', currentLanguage) },
      { '@type': 'ListItem', position: 2, name: t('landing_nav_docs'), item: buildLocalizedUrl('/docs', currentLanguage) },
      ...(article && articleSection
        ? [
            { '@type': 'ListItem', position: 3, name: articleSection.title, item: buildLocalizedUrl(`/docs/${articleSection.slug}/${article.slug}`, currentLanguage) },
            { '@type': 'ListItem', position: 4, name: article.title, item: canonicalUrl },
          ]
        : []),
    ],
  };

  const articleJsonLd = article
    ? {
        '@context': 'https://schema.org',
        '@type': 'TechArticle',
        headline: article.seoTitle,
        description: article.seoDescription,
        inLanguage: currentLanguage,
        url: canonicalUrl,
        about: articleSection?.title || docs.title,
      }
    : null;

  return (
    <>
      <Helmet>
        <html lang={currentLanguage} />
        <title>{article ? article.seoTitle : docs.homeSeoTitle}</title>
        <meta name="description" content={article ? article.seoDescription : docs.homeSeoDescription} />
        {!article && <meta name="keywords" content={docs.homeSeoKeywords} />}
        <meta name="robots" content="index,follow" />
        <link rel="canonical" href={canonicalUrl} />
        <link rel="alternate" href={`https://dotblue.ai${currentPath}`} hrefLang="x-default" />
        {SUPPORTED_LANGUAGES.map((language) => (
          <link key={language} rel="alternate" hrefLang={language} href={buildLocalizedUrl(currentPath, language)} />
        ))}
        <meta property="og:type" content="article" />
        <meta property="og:title" content={article ? article.seoTitle : docs.homeSeoTitle} />
        <meta property="og:description" content={article ? article.seoDescription : docs.homeSeoDescription} />
        <meta property="og:url" content={canonicalUrl} />
        <script type="application/ld+json">{JSON.stringify(breadcrumbJsonLd)}</script>
        {articleJsonLd && <script type="application/ld+json">{JSON.stringify(articleJsonLd)}</script>}
      </Helmet>

      <div style={{ maxWidth: 1240, margin: '0 auto', padding: '40px 24px 96px' }}>
        <Breadcrumb
          items={[
            { title: <Link to={localizedHomePath}>{t('welcome')}</Link> },
            { title: <Link to={localizedDocsHomePath}>{t('landing_nav_docs')}</Link> },
            ...(articleSection ? [{ title: articleSection.title }] : []),
            ...(article ? [{ title: article.title }] : []),
          ]}
        />

        {!article ? (
          <>
            <section style={{ padding: '40px 0 24px' }}>
              <Space direction="vertical" size={18} style={{ width: '100%' }}>
                <Tag color="blue" style={{ width: 'fit-content', borderRadius: 999, padding: '6px 12px' }}>
                  {docs.eyebrow}
                </Tag>
                <Title level={1} style={{ margin: 0, maxWidth: 920 }}>
                  {docs.title}
                </Title>
                <Paragraph style={{ fontSize: 17, color: '#5b6673', marginBottom: 0, maxWidth: 980 }}>
                  {docs.subtitle}
                </Paragraph>
                <Space wrap size={[12, 12]}>
                  <Button type="primary" shape="round" icon={<LoginOutlined />} onClick={() => navigate(getLocalizedPath('/login', currentLanguage))}>
                    {t('login')}
                  </Button>
                  <Button shape="round" icon={<RocketOutlined />} onClick={() => navigate(getLocalizedPath(isAuthenticated ? '/dashboard' : '/login', currentLanguage))}>
                    {t('go_to_dashboard')}
                  </Button>
                  <Button shape="round" icon={<GithubOutlined />} href={DOCS_REPO_URL} target="_blank" rel="noreferrer">
                    GitHub
                  </Button>
                </Space>
              </Space>
            </section>

            <Row gutter={[32, 32]} align="top">
              <Col xs={24} lg={17}>
                <Space direction="vertical" size={28} style={{ width: '100%' }}>
                  {docs.sections.map((section) => (
                    <Card key={section.slug} variant="borderless" style={{ borderRadius: 24, boxShadow: '0 16px 48px rgba(15, 23, 42, 0.08)' }}>
                      <Space direction="vertical" size={18} style={{ width: '100%' }}>
                        <div>
                          <Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>
                            {docs.categoriesLabel}
                          </Text>
                          <Title level={2} style={{ margin: 0 }}>
                            {section.title}
                          </Title>
                        </div>
                        <Paragraph style={{ marginBottom: 0, color: '#475569', fontSize: 16 }}>
                          {section.description}
                        </Paragraph>
                        <div>
                          <Text strong style={{ display: 'block', marginBottom: 12 }}>
                            {docs.sectionDescriptionLabel}
                          </Text>
                          <Row gutter={[16, 16]}>
                            {section.articles.map((item) => (
                              <Col xs={24} md={12} key={`${section.slug}-${item.slug}`}>
                                <Card
                                  size="small"
                                  hoverable
                                  variant="borderless"
                                  style={{ borderRadius: 18, background: '#f8fbff', boxShadow: '0 8px 24px rgba(15, 23, 42, 0.04)' }}
                                >
                                  <Space direction="vertical" size={8} style={{ width: '100%' }}>
                                    <Tag color="blue" style={{ width: 'fit-content', borderRadius: 999 }}>
                                      {item.readingTime}
                                    </Tag>
                                    <Link to={getDocsHref(currentLanguage, section.slug, item.slug)}>
                                      <Text strong style={{ fontSize: 16 }}>
                                        {item.title}
                                      </Text>
                                    </Link>
                                    <Paragraph style={{ marginBottom: 0, color: '#64748b' }}>{item.summary}</Paragraph>
                                  </Space>
                                </Card>
                              </Col>
                            ))}
                          </Row>
                        </div>
                      </Space>
                    </Card>
                  ))}

                  <Card variant="borderless" style={{ borderRadius: 24, boxShadow: '0 16px 48px rgba(15, 23, 42, 0.08)' }}>
                    <Title level={2}>{docs.repoTitle}</Title>
                    <Paragraph style={{ color: '#475569', marginBottom: 20 }}>{docs.repoDescription}</Paragraph>
                    <Button type="primary" shape="round" icon={<GithubOutlined />} href={DOCS_REPO_URL} target="_blank" rel="noreferrer">
                      {DOCS_REPO_URL}
                    </Button>
                  </Card>
                </Space>
              </Col>

              <Col xs={24} lg={7}>
                <div style={{ position: 'sticky', top: 108 }}>
                  <Card variant="borderless" style={{ borderRadius: 20, boxShadow: '0 12px 32px rgba(15, 23, 42, 0.06)' }}>
                    <Title level={5} style={{ marginTop: 0 }}>{docs.quickLinksTitle}</Title>
                    <Space direction="vertical" size={12} style={{ width: '100%' }}>
                      {docs.quickLinks.map((item) => (
                        <Card key={`${item.label}-${item.url}`} size="small" variant="borderless" style={{ borderRadius: 16, background: '#f8fbff' }}>
                          <Space direction="vertical" size={4} style={{ width: '100%' }}>
                            {item.url.startsWith('http') ? (
                              <a href={item.url} target="_blank" rel="noreferrer">{item.label}</a>
                            ) : (
                              <Link to={item.url}>{item.label}</Link>
                            )}
                            {item.description && <Text type="secondary" style={{ lineHeight: 1.7 }}>{item.description}</Text>}
                          </Space>
                        </Card>
                      ))}
                    </Space>
                    <Divider />
                    <Title level={5}>{docs.popularLabel}</Title>
                    <Space direction="vertical" size={10} style={{ width: '100%' }}>
                      {allArticles.slice(0, 5).map((item) => (
                        <Link key={`${item.sectionSlug}-${item.slug}`} to={getDocsHref(currentLanguage, item.sectionSlug, item.slug)}>
                          {item.title}
                        </Link>
                      ))}
                    </Space>
                  </Card>
                </div>
              </Col>
            </Row>
          </>
        ) : (
          <Row gutter={[32, 32]} align="top" style={{ paddingTop: 32 }}>
            <Col xs={24} lg={6}>
              <div style={{ position: 'sticky', top: 108 }}>
                <Card variant="borderless" style={{ borderRadius: 20, boxShadow: '0 12px 32px rgba(15, 23, 42, 0.06)' }}>
                  <Title level={5} style={{ marginTop: 0 }}>{docs.categoriesLabel}</Title>
                  <Space direction="vertical" size={18} style={{ width: '100%' }}>
                    {docs.sections.map((section) => (
                      <div key={section.slug}>
                        <Text strong style={{ display: 'block', marginBottom: 8 }}>{section.title}</Text>
                        <Space direction="vertical" size={8} style={{ width: '100%' }}>
                          {section.articles.map((item) => {
                            const active = item.slug === article.slug && item.sectionSlug === article.sectionSlug;
                            return (
                              <Link
                                key={`${section.slug}-${item.slug}`}
                                to={getDocsHref(currentLanguage, section.slug, item.slug)}
                                style={{ color: active ? '#1677ff' : '#334155', fontWeight: active ? 600 : 400 }}
                              >
                                {item.title}
                              </Link>
                            );
                          })}
                        </Space>
                      </div>
                    ))}
                  </Space>
                </Card>
              </div>
            </Col>

            <Col xs={24} lg={13}>
              <Card variant="borderless" style={{ borderRadius: 24, boxShadow: '0 16px 48px rgba(15, 23, 42, 0.08)' }}>
                <Space direction="vertical" size={18} style={{ width: '100%' }}>
                  <Tag color="blue" style={{ width: 'fit-content', borderRadius: 999, padding: '6px 12px' }}>
                    {articleSection?.title}
                  </Tag>
                  <Title level={1} style={{ margin: 0 }}>{article.title}</Title>
                  <Paragraph style={{ fontSize: 17, color: '#475569', marginBottom: 0 }}>{article.summary}</Paragraph>
                  <Space split={<Divider type="vertical" />} wrap>
                    <Text type="secondary">{article.readingTime}</Text>
                    <Link to={localizedDocsHomePath}>{t('landing_nav_docs')}</Link>
                  </Space>
                </Space>

                <Divider style={{ marginBlock: 28 }} />

                <Space direction="vertical" size={40} style={{ width: '100%' }}>
                  {article.sections.map((section) => (
                    <section id={section.id} key={section.id}>
                      <Title level={2}>{section.title}</Title>
                      {section.intro && <Paragraph style={{ fontSize: 16, color: '#475569' }}>{section.intro}</Paragraph>}
                      {section.paragraphs?.map((paragraph) => (
                        <Paragraph key={paragraph} style={{ lineHeight: 1.9 }}>{paragraph}</Paragraph>
                      ))}
                      {section.bullets && (
                        <ul style={{ paddingLeft: 20, marginBottom: 0 }}>
                          {section.bullets.map((bullet) => (
                            <li key={bullet} style={{ marginBottom: 12, color: '#334155', lineHeight: 1.9 }}>{bullet}</li>
                          ))}
                        </ul>
                      )}
                      {section.steps && (
                        <Space direction="vertical" size={16} style={{ width: '100%', marginTop: 16 }}>
                          {section.steps.map((item, index) => (
                            <Space key={`${section.id}-${item.title}-${index}`} align="start" size={16}>
                              <div style={{ width: 32, height: 32, borderRadius: 999, background: '#eaf3ff', color: '#1677ff', display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 700, flexShrink: 0, marginTop: 4 }}>
                                {index + 1}
                              </div>
                              <div>
                                <Text strong style={{ display: 'block', fontSize: 16, marginBottom: 4 }}>{item.title}</Text>
                                <Paragraph type="secondary" style={{ marginBottom: 0 }}>{item.desc}</Paragraph>
                              </div>
                            </Space>
                          ))}
                        </Space>
                      )}
                      {section.note && (
                        <Card size="small" variant="borderless" style={{ marginTop: 20, borderRadius: 18, background: '#f8fbff' }}>
                          <Text>{section.note}</Text>
                        </Card>
                      )}
                      {section.links && (
                        <Space direction="vertical" size={12} style={{ width: '100%', marginTop: 20 }}>
                          {section.links.map((link) => (
                            <Card key={link.url} size="small" variant="borderless" style={{ borderRadius: 16, background: '#f8fbff', boxShadow: '0 8px 24px rgba(15, 23, 42, 0.04)' }}>
                              <Space direction="vertical" size={4} style={{ width: '100%' }}>
                                <a href={link.url} target="_blank" rel="noreferrer">{link.label}</a>
                                {link.description && <Text type="secondary" style={{ lineHeight: 1.7 }}>{link.description}</Text>}
                              </Space>
                            </Card>
                          ))}
                        </Space>
                      )}
                      {section.code && (
                        <pre style={{ marginTop: 20, padding: 20, borderRadius: 18, background: '#0f172a', color: '#e2e8f0', overflowX: 'auto', fontSize: 13, lineHeight: 1.7 }}>
                          <code>{section.code.value}</code>
                        </pre>
                      )}
                    </section>
                  ))}
                </Space>

                <Divider style={{ marginBlock: 32 }} />

                <Row gutter={[16, 16]}>
                  <Col xs={24} md={12}>
                    {previousArticle && (
                      <Card variant="borderless" style={{ borderRadius: 18, background: '#f8fbff' }}>
                        <Space direction="vertical" size={8}>
                          <Text type="secondary"><ArrowLeftOutlined /> {previousLabel}</Text>
                          <Link to={getDocsHref(currentLanguage, previousArticle.sectionSlug, previousArticle.slug)}>{previousArticle.title}</Link>
                        </Space>
                      </Card>
                    )}
                  </Col>
                  <Col xs={24} md={12}>
                    {nextArticle && (
                      <Card variant="borderless" style={{ borderRadius: 18, background: '#f8fbff' }}>
                        <Space direction="vertical" size={8} style={{ width: '100%', textAlign: 'right' }}>
                          <Text type="secondary">{nextLabel} <ArrowRightOutlined /></Text>
                          <Link to={getDocsHref(currentLanguage, nextArticle.sectionSlug, nextArticle.slug)}>{nextArticle.title}</Link>
                        </Space>
                      </Card>
                    )}
                  </Col>
                </Row>
              </Card>
            </Col>

            <Col xs={24} lg={5}>
              <div style={{ position: 'sticky', top: 108 }}>
                <Card variant="borderless" style={{ borderRadius: 20, boxShadow: '0 12px 32px rgba(15, 23, 42, 0.06)' }}>
                  <Title level={5} style={{ marginTop: 0 }}>{tocTitle}</Title>
                  <Anchor offsetTop={120} items={anchorItems} />
                  <Divider />
                  <Space direction="vertical" size={10} style={{ width: '100%' }}>
                    <Button block icon={<BookOutlined />} onClick={() => navigate(localizedDocsHomePath)}>
                      {t('landing_nav_docs')}
                    </Button>
                    <Button block icon={<GithubOutlined />} href={DOCS_REPO_URL} target="_blank" rel="noreferrer">
                      GitHub
                    </Button>
                  </Space>
                </Card>
              </div>
            </Col>
          </Row>
        )}
      </div>
    </>
  );
};

export default ProductDocsPage;
