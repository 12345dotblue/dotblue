import React from 'react';
import { Button, Card, Space, Typography } from 'antd';
import { CloudDownloadOutlined, PlusOutlined, SafetyOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { getLocalizedPath, resolveSupportedLanguage } from '../../i18n/config';

const { Title, Paragraph, Text } = Typography;

const PlatformSkillBuilderPage: React.FC = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const currentLanguage = resolveSupportedLanguage(i18n.resolvedLanguage || i18n.language);

  const go = (path: string) => navigate(getLocalizedPath(path, currentLanguage));

  return (
    <div style={{ animation: 'fadeIn 0.5s ease-out' }}>
      <div style={{ marginBottom: 24 }}>
        <Title level={3} style={{ marginBottom: 8 }}>
          <PlusOutlined style={{ marginRight: 8 }} />
          {t('platform_skill_builder_title')}
        </Title>
        <Paragraph type="secondary" style={{ marginBottom: 0, maxWidth: 760 }}>
          {t('platform_skill_builder_desc')}
        </Paragraph>
      </div>

      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(260px, 1fr))',
          gap: 16,
          marginBottom: 24,
        }}
      >
        <Card style={{ borderRadius: 20 }}>
          <Space direction="vertical" size={10}>
            <Text strong>{t('platform_skill_builder_card_create_title')}</Text>
            <Paragraph type="secondary" style={{ marginBottom: 0 }}>
              {t('platform_skill_builder_card_create_desc')}
            </Paragraph>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => go('/admin/platform/skills/create')}>
              {t('platform_skill_builder_card_create_action')}
            </Button>
          </Space>
        </Card>

        <Card style={{ borderRadius: 20 }}>
          <Space direction="vertical" size={10}>
            <Text strong>{t('platform_skill_builder_card_import_title')}</Text>
            <Paragraph type="secondary" style={{ marginBottom: 0 }}>
              {t('platform_skill_builder_card_import_desc')}
            </Paragraph>
            <Button icon={<CloudDownloadOutlined />} onClick={() => go('/admin/platform/skills/import')}>
              {t('platform_skill_builder_card_import_action')}
            </Button>
          </Space>
        </Card>

        <Card style={{ borderRadius: 20 }}>
          <Space direction="vertical" size={10}>
            <Text strong>{t('platform_skill_builder_card_hubs_title')}</Text>
            <Paragraph type="secondary" style={{ marginBottom: 0 }}>
              {t('platform_skill_builder_card_hubs_desc')}
            </Paragraph>
            <Button icon={<SafetyOutlined />} onClick={() => go('/admin/platform/skill-hubs')}>
              {t('platform_skill_builder_card_hubs_action')}
            </Button>
          </Space>
        </Card>
      </div>

      <Card style={{ borderRadius: 20 }}>
        <Space direction="vertical" size={8}>
          <Text strong>{t('platform_skill_builder_next_steps_title')}</Text>
          <Paragraph type="secondary" style={{ marginBottom: 0 }}>
            {t('platform_skill_builder_next_steps_desc')}
          </Paragraph>
          <Space wrap>
            <Button onClick={() => go('/admin/platform/skill-market')}>
              {t('platform_skill_builder_next_steps_market')}
            </Button>
            <Button onClick={() => go('/admin/platform/skills')}>
              {t('platform_skill_builder_next_steps_governance')}
            </Button>
          </Space>
        </Space>
      </Card>
    </div>
  );
};

export default PlatformSkillBuilderPage;
