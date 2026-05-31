import React, { useEffect, useState } from 'react';
import { Button, Card, Empty, Space, Typography } from 'antd';
import { ArrowLeftOutlined, AppstoreOutlined } from '@ant-design/icons';
import axios from 'axios';
import { useNavigate, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { BACKEND_URL } from '../../config';
import { casdoorService } from '../identity/CasdoorService';
import { getLocalizedPath, resolveSupportedLanguage } from '../../i18n/config';
import AgentSkillsPanel from './AgentSkillsPanel';

const { Title, Paragraph, Text } = Typography;
const CURRENT_ENTERPRISE_STORAGE_KEY = 'dotblue_current_enterprise_id';

function getAgentAuthHeaders() {
  const token = casdoorService.getToken();
  const enterpriseId = localStorage.getItem(CURRENT_ENTERPRISE_STORAGE_KEY)?.trim();
  const headers: Record<string, string> = {};
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }
  if (enterpriseId) {
    headers['X-Enterprise-ID'] = enterpriseId;
  }
  return headers;
}

const AgentSkillManagementPage: React.FC = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const { agentId = '' } = useParams();
  const currentLanguage = resolveSupportedLanguage(i18n.resolvedLanguage || i18n.language);
  const marketPath = getLocalizedPath('/admin/platform/skill-market', currentLanguage);
  const builderPath = getLocalizedPath('/admin/platform/skills/new', currentLanguage);
  const [agentName, setAgentName] = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!agentId) {
      setLoading(false);
      return;
    }
    setLoading(true);
    axios.get(`${BACKEND_URL}/api/agents`, {
      headers: getAgentAuthHeaders(),
    }).then((res) => {
      const agents = Array.isArray(res.data) ? res.data : [];
      const current = agents.find((item: { id: string }) => item.id === agentId);
      setAgentName(current?.agentName || '');
    }).catch(() => {
      setAgentName('');
    }).finally(() => {
      setLoading(false);
    });
  }, [agentId]);

  return (
    <div style={{ animation: 'fadeIn 0.5s ease-out' }}>
      <Space direction="vertical" size={16} style={{ width: '100%' }}>
        <Button
          icon={<ArrowLeftOutlined />}
          onClick={() => navigate(getLocalizedPath('/dashboard', currentLanguage))}
          style={{ width: 'fit-content' }}
        >
          {t('agent_skill_page_back')}
        </Button>

        <div>
          <Title level={3} style={{ marginBottom: 8 }}>
            <AppstoreOutlined style={{ marginRight: 8 }} />
            {t('agent_skill_page_title')}
          </Title>
          <Paragraph type="secondary" style={{ marginBottom: 0 }}>
            {agentName
              ? t('agent_skill_page_desc_with_name', { agentName })
              : t('agent_skill_page_desc')}
          </Paragraph>
        </div>

        <Card style={{ borderRadius: 20 }} loading={loading}>
          {agentId ? (
            <>
              {agentName ? (
                <div style={{ marginBottom: 12 }}>
                  <Text strong>{agentName}</Text>
                </div>
              ) : null}
              <AgentSkillsPanel
                agentId={agentId}
                authHeaders={getAgentAuthHeaders()}
                marketHref={marketPath}
                builderHref={builderPath}
              />
            </>
          ) : (
            <Empty description={t('agent_skill_page_empty')} image={Empty.PRESENTED_IMAGE_SIMPLE} />
          )}
        </Card>
      </Space>
    </div>
  );
};

export default AgentSkillManagementPage;
