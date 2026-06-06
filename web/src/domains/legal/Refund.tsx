import React from 'react';
import { Typography, Breadcrumb, Divider } from 'antd';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { getLocalizedPath, resolveSupportedLanguage } from '../../i18n/config';

const { Title, Paragraph, Text } = Typography;

const Refund: React.FC = () => {
  const { t, i18n } = useTranslation();
  const currentLanguage = resolveSupportedLanguage(i18n?.resolvedLanguage || i18n?.language);

  const eligibilityItems = [1, 2, 3, 4].map(i => t(`refund_eligibility_${i}`));

  return (
    <div style={{ maxWidth: 800, margin: '40px auto', padding: '0 24px' }}>
      <Breadcrumb items={[{ title: <Link to={getLocalizedPath('/', currentLanguage)}>{t('welcome')}</Link> }, { title: t('refund') }]} />

      <div style={{ marginTop: 40 }}>
        <Title level={1}>{t('refund_policy_title')}</Title>
        <Text type="secondary">{t('last_updated')}</Text>
        <Divider />

        <Paragraph style={{ fontSize: 15, lineHeight: 1.8 }}>{t('refund_intro')}</Paragraph>

        <div style={{ marginTop: 36 }}>
          <Title level={3}>{t('refund_trial_title')}</Title>
          <Paragraph style={{ lineHeight: 1.8 }}>{t('refund_trial')}</Paragraph>
        </div>

        <div style={{ marginTop: 36 }}>
          <Title level={3}>{t('refund_eligibility_title')}</Title>
          <ul style={{ paddingLeft: 20, lineHeight: 2 }}>
            {eligibilityItems.map((item, i) => (
              <li key={i}>{item}</li>
            ))}
          </ul>
        </div>

        <div style={{ marginTop: 36 }}>
          <Title level={3}>{t('refund_process_title')}</Title>
          <Paragraph style={{ lineHeight: 1.8 }}>{t('refund_process')}</Paragraph>
        </div>

        <div style={{ marginTop: 36 }}>
          <Title level={3}>{t('refund_exceptions_title')}</Title>
          <Paragraph style={{ lineHeight: 1.8 }}>{t('refund_exceptions')}</Paragraph>
        </div>

        <div style={{ marginTop: 36 }}>
          <Title level={3}>{t('refund_contact_title')}</Title>
          <Paragraph style={{ lineHeight: 1.8 }}>{t('refund_contact')}</Paragraph>
        </div>
      </div>
    </div>
  );
};

export default Refund;
