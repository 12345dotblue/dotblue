import React from 'react';
import { Typography, Breadcrumb, Divider, Anchor, Row, Col } from 'antd';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { getLocalizedPath, resolveSupportedLanguage } from '../../i18n/config';

const { Title, Paragraph, Text } = Typography;

const sections = [
  { id: 'terms_s1', titleKey: 'terms_s1_title', key: 'terms_s1' },
  { id: 'terms_s2', titleKey: 'terms_s2_title', key: 'terms_s2' },
  { id: 'terms_s3', titleKey: 'terms_s3_title', key: 'terms_s3' },
  { id: 'terms_s4', titleKey: 'terms_s4_title', key: 'terms_s4' },
  { id: 'terms_s5', titleKey: 'terms_s5_title', key: 'terms_s5' },
  { id: 'terms_s6', titleKey: 'terms_s6_title', key: 'terms_s6' },
  { id: 'terms_s7', titleKey: 'terms_s7_title', key: 'terms_s7' },
  { id: 'terms_s8', titleKey: 'terms_s8_title', key: 'terms_s8' },
  { id: 'terms_s9', titleKey: 'terms_s9_title', key: 'terms_s9' },
];

const Terms: React.FC = () => {
  const { t, i18n } = useTranslation();
  const currentLanguage = resolveSupportedLanguage(i18n.resolvedLanguage || i18n.language);

  return (
    <div style={{ maxWidth: 1000, margin: '40px auto', padding: '0 24px' }}>
      <Breadcrumb items={[{ title: <Link to={getLocalizedPath('/', currentLanguage)}>{t('welcome')}</Link> }, { title: t('terms') }]} />

      <Row gutter={48} style={{ marginTop: 40 }}>
        <Col xs={0} md={6}>
          <Anchor
            offsetTop={100}
            items={sections.map(s => ({ key: s.id, href: `#${s.id}`, title: t(s.titleKey) }))}
          />
        </Col>
        <Col xs={24} md={18}>
          <Title level={1}>{t('terms')}</Title>
          <Text type="secondary">{t('last_updated')}</Text>
          <Divider />

          <Paragraph style={{ fontSize: 15, lineHeight: 1.8 }}>{t('terms_intro')}</Paragraph>

          {sections.map(s => (
            <div key={s.id} id={s.id} style={{ marginTop: 36 }}>
              <Title level={3}>{t(s.titleKey)}</Title>
              <Paragraph style={{ lineHeight: 1.8 }}>{t(s.key)}</Paragraph>
            </div>
          ))}
        </Col>
      </Row>
    </div>
  );
};

export default Terms;
