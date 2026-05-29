import React, { useMemo } from 'react';
import { Anchor, Breadcrumb, Button, Card, Col, Divider, List, Row, Space, Tag, Typography } from 'antd';
import { BookOutlined, GithubOutlined, LoginOutlined, RocketOutlined } from '@ant-design/icons';
import { Helmet } from 'react-helmet-async';
import { Link, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { SUPPORTED_LANGUAGES, buildLocalizedUrl, getLocalizedPath, resolveSupportedLanguage } from '../../i18n/config';
import { useAuthState } from '../identity/useAuthState';
import { getDocsContent } from './productDocsContent';

const { Title, Paragraph, Text } = Typography;

const REPO_URL = 'https://github.com/12345dotblue/dotblue';

const ProductDocsPage: React.FC = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const isAuthenticated = useAuthState();
  const currentLanguage = resolveSupportedLanguage(i18n.resolvedLanguage || i18n.language);
  const content = getDocsContent(currentLanguage);
  const canonicalUrl = buildLocalizedUrl('/docs', currentLanguage);
  const localizedHomePath = getLocalizedPath('/', currentLanguage);

  const anchorItems = useMemo(
    () => content.sections.map((section) => ({ key: section.id, href: `#${section.id}`, title: section.title })),
    [content.sections],
  );

  return (
    <>
      <Helmet>
        <html lang={currentLanguage} />
        <title>{content.seoTitle}</title>
        <meta name="description" content={content.seoDescription} />
        <meta name="keywords" content={content.seoKeywords} />
        <meta name="robots" content="index,follow" />
        <link rel="canonical" href={canonicalUrl} />
        <link rel="alternate" href="https://dotblue.ai/docs" hrefLang="x-default" />
        {SUPPORTED_LANGUAGES.map((language) => (
          <link key={language} rel="alternate" hrefLang={language} href={buildLocalizedUrl('/docs', language)} />
        ))}
        <meta property="og:type" content="article" />
        <meta property="og:title" content={content.seoTitle} />
        <meta property="og:description" content={content.seoDescription} />
        <meta property="og:url" content={canonicalUrl} />
      </Helmet>

      <div style={{ maxWidth: 1180, margin: '0 auto', padding: '40px 24px 96px' }}>
        <Breadcrumb
          items={[
            { title: <Link to={localizedHomePath}>{t('welcome')}</Link> },
            { title: t('landing_nav_docs') },
          ]}
        />

        <section style={{ padding: '40px 0 32px' }}>
          <Space direction="vertical" size={18} style={{ width: '100%' }}>
            <Tag color="blue" style={{ width: 'fit-content', borderRadius: 999, padding: '6px 12px' }}>
              {content.eyebrow}
            </Tag>
            <Title level={1} style={{ margin: 0, maxWidth: 880 }}>
              {content.title}
            </Title>
            <Paragraph style={{ fontSize: 17, color: '#5b6673', marginBottom: 0, maxWidth: 920 }}>
              {content.subtitle}
            </Paragraph>
            <Space wrap size={[12, 12]}>
              <Button type="primary" shape="round" icon={<LoginOutlined />} onClick={() => navigate(getLocalizedPath('/login', currentLanguage))}>
                {content.primaryCta}
              </Button>
              <Button shape="round" icon={<RocketOutlined />} onClick={() => navigate(getLocalizedPath(isAuthenticated ? '/dashboard' : '/login', currentLanguage))}>
                {content.secondaryCta}
              </Button>
              <Button shape="round" icon={<GithubOutlined />} href={REPO_URL} target="_blank" rel="noreferrer">
                {content.githubCta}
              </Button>
            </Space>
          </Space>
        </section>

        <Row gutter={[40, 32]} align="top">
          <Col xs={24} lg={17}>
            <Space direction="vertical" size={40} style={{ width: '100%' }}>
              {content.sections.map((section) => (
                <section id={section.id} key={section.id}>
                  <Title level={2}>{section.title}</Title>
                  {section.intro && <Paragraph style={{ fontSize: 16, color: '#475569' }}>{section.intro}</Paragraph>}

                  {section.cards && (
                    <Row gutter={[16, 16]} style={{ marginTop: 8 }}>
                      {section.cards.map((card) => (
                        <Col xs={24} md={12} key={card.title}>
                          <Card bordered={false} style={{ height: '100%', borderRadius: 20, boxShadow: '0 12px 32px rgba(15, 23, 42, 0.06)' }}>
                            <Space direction="vertical" size={10} style={{ width: '100%' }}>
                              {card.tag && <Tag style={{ width: 'fit-content', borderRadius: 999 }}>{card.tag}</Tag>}
                              <Title level={4} style={{ margin: 0 }}>
                                {card.title}
                              </Title>
                              <Paragraph type="secondary" style={{ marginBottom: 0 }}>
                                {card.desc}
                              </Paragraph>
                            </Space>
                          </Card>
                        </Col>
                      ))}
                    </Row>
                  )}

                  {section.steps && (
                    <List
                      style={{ marginTop: 8 }}
                      dataSource={section.steps}
                      renderItem={(item, index) => (
                        <List.Item style={{ paddingInline: 0 }}>
                          <Space align="start" size={16}>
                            <div
                              style={{
                                width: 32,
                                height: 32,
                                borderRadius: 999,
                                background: '#eaf3ff',
                                color: '#1677ff',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                fontWeight: 700,
                                flexShrink: 0,
                                marginTop: 4,
                              }}
                            >
                              {index + 1}
                            </div>
                            <div>
                              <Text strong style={{ display: 'block', fontSize: 16, marginBottom: 4 }}>
                                {item.title}
                              </Text>
                              <Paragraph type="secondary" style={{ marginBottom: 0 }}>
                                {item.desc}
                              </Paragraph>
                            </div>
                          </Space>
                        </List.Item>
                      )}
                    />
                  )}

                  {section.paragraphs?.map((paragraph) => (
                    <Paragraph key={paragraph} style={{ lineHeight: 1.85 }}>
                      {paragraph}
                    </Paragraph>
                  ))}

                  {section.bullets && (
                    <ul style={{ paddingLeft: 20, marginBottom: 0 }}>
                      {section.bullets.map((bullet) => (
                        <li key={bullet} style={{ marginBottom: 12, color: '#334155', lineHeight: 1.8 }}>
                          {bullet}
                        </li>
                      ))}
                    </ul>
                  )}

                  {section.links && (
                    <Space direction="vertical" size={12} style={{ width: '100%', marginTop: 20 }}>
                      {section.links.map((link) => (
                        <Card
                          key={link.url}
                          size="small"
                          bordered={false}
                          style={{ borderRadius: 16, background: '#f8fbff', boxShadow: '0 8px 24px rgba(15, 23, 42, 0.04)' }}
                        >
                          <Space direction="vertical" size={4} style={{ width: '100%' }}>
                            <a href={link.url} target="_blank" rel="noreferrer">
                              {link.label}
                            </a>
                            {link.description && (
                              <Text type="secondary" style={{ lineHeight: 1.7 }}>
                                {link.description}
                              </Text>
                            )}
                          </Space>
                        </Card>
                      ))}
                    </Space>
                  )}

                  {section.code && (
                    <pre
                      style={{
                        marginTop: 20,
                        padding: 20,
                        borderRadius: 18,
                        background: '#0f172a',
                        color: '#e2e8f0',
                        overflowX: 'auto',
                        fontSize: 13,
                        lineHeight: 1.7,
                      }}
                    >
                      <code>{section.code}</code>
                    </pre>
                  )}

                  <Divider style={{ marginTop: 32 }} />
                </section>
              ))}

              <section id="github">
                <Title level={2}>{content.repoTitle}</Title>
                <Paragraph style={{ color: '#475569' }}>{content.repoDescription}</Paragraph>
                <ul style={{ paddingLeft: 20 }}>
                  {content.repoPoints.map((point) => (
                    <li key={point} style={{ marginBottom: 12, color: '#334155', lineHeight: 1.8 }}>
                      {point}
                    </li>
                  ))}
                </ul>
                <Button type="primary" shape="round" icon={<GithubOutlined />} href={REPO_URL} target="_blank" rel="noreferrer">
                  {REPO_URL}
                </Button>
              </section>
            </Space>
          </Col>

          <Col xs={24} lg={7}>
            <div style={{ position: 'sticky', top: 108 }}>
              <Card bordered={false} style={{ borderRadius: 20, boxShadow: '0 12px 32px rgba(15, 23, 42, 0.06)' }}>
                <Title level={5} style={{ marginTop: 0 }}>
                  {content.anchorTitle}
                </Title>
                <Anchor offsetTop={120} items={[...anchorItems, { key: 'github', href: '#github', title: content.repoTitle }]} />
                <Divider />
                <Space direction="vertical" size={10} style={{ width: '100%' }}>
                  <Button block icon={<BookOutlined />} onClick={() => navigate(localizedHomePath)}>
                    {t('welcome')}
                  </Button>
                  <Button block icon={<GithubOutlined />} href={REPO_URL} target="_blank" rel="noreferrer">
                    GitHub
                  </Button>
                </Space>
              </Card>
            </div>
          </Col>
        </Row>
      </div>
    </>
  );
};

export default ProductDocsPage;
