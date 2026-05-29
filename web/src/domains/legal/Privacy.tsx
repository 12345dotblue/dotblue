import React from 'react';
import { Typography, Breadcrumb, Divider, Anchor, Row, Col } from 'antd';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { getLocalizedPath, resolveSupportedLanguage } from '../../i18n/config';

const { Title, Paragraph, Text } = Typography;

const sections = [
  { id: 'privacy_collect', titleKey: 'privacy_collect_title' },
  { id: 'privacy_use', titleKey: 'privacy_use_title' },
  { id: 'privacy_security', titleKey: 'privacy_security_title' },
  { id: 'privacy_sharing', titleKey: 'privacy_sharing_title' },
  { id: 'privacy_rights', titleKey: 'privacy_rights_title' },
  { id: 'privacy_retention', titleKey: 'privacy_retention_title' },
  { id: 'privacy_cookies', titleKey: 'privacy_cookies_title' },
  { id: 'privacy_contact', titleKey: 'privacy_contact_title' },
];

const Privacy: React.FC = () => {
  const { t, i18n } = useTranslation();
  const currentLanguage = resolveSupportedLanguage(i18n.resolvedLanguage || i18n.language);

  return (
    <div style={{ maxWidth: 1000, margin: '40px auto', padding: '0 24px' }}>
      <Breadcrumb items={[{ title: <Link to={getLocalizedPath('/', currentLanguage)}>{t('welcome')}</Link> }, { title: t('privacy') }]} />

      <Row gutter={48} style={{ marginTop: 40 }}>
        <Col xs={0} md={6}>
          <Anchor
            offsetTop={100}
            items={sections.map(s => ({ key: s.id, href: `#${s.id}`, title: t(s.titleKey) }))}
          />
        </Col>
        <Col xs={24} md={18}>
          <Title level={1}>{t('privacy')}</Title>
          <Text type="secondary">{t('last_updated')}</Text>
          <Divider />

          <Paragraph style={{ fontSize: 15, lineHeight: 1.8 }}>{t('privacy_intro')}</Paragraph>

          {sections.map(s => (
            <div key={s.id} id={s.id} style={{ marginTop: 36 }}>
              <Title level={3}>{t(s.titleKey)}</Title>
              <Paragraph style={{ lineHeight: 1.8 }}>{t(s.id)}</Paragraph>
            </div>
          ))}
        </Col>
      </Row>
    </div>
  );
};

export default Privacy;
