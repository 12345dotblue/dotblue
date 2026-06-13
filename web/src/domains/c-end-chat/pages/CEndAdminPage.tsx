import { useEffect, useState } from 'react';
import { Alert, Button, Card, Col, Empty, Flex, Row, Spin, Typography } from 'antd';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { listAgents, type AgentSummary } from '../services/cEndChatApi';

type CEndAdminPageProps = {
  embedded?: boolean;
};

export function CEndAdminPage({ embedded = false }: CEndAdminPageProps) {
  const { t, i18n } = useTranslation();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [agents, setAgents] = useState<AgentSummary[]>([]);

  useEffect(() => {
    let mounted = true;
    listAgents()
      .then((items) => {
        if (!mounted) return;
        setAgents(items);
        setError(null);
      })
      .catch((err: unknown) => {
        if (!mounted) return;
        setError(err instanceof Error ? err.message : t('c_end_admin_load_failed'));
      })
      .finally(() => {
        if (mounted) setLoading(false);
      });
    return () => {
      mounted = false;
    };
  }, [t]);

  if (loading) {
    return <Spin />;
  }

  return (
    <Flex vertical gap={16} style={{ width: '100%' }}>
      {embedded ? (
        <Card variant="borderless" style={{ borderRadius: 20 }}>
          <Typography.Paragraph style={{ marginBottom: 8 }}>
            {t('c_end_admin_embedded_desc')}
          </Typography.Paragraph>
          <Typography.Paragraph type="secondary" style={{ marginBottom: 0 }}>
            {t('c_end_admin_card_hint')}
          </Typography.Paragraph>
        </Card>
      ) : (
        <>
          <Typography.Title level={2}>{t('c_end_admin_title')}</Typography.Title>
          <Typography.Paragraph>{t('c_end_admin_desc')}</Typography.Paragraph>
        </>
      )}
      {error ? <Alert type="error" title={error} /> : null}
      {agents.length ? (
        <Row gutter={[16, 16]}>
          {agents.map((agent) => (
            <Col xs={24} md={12} xl={8} key={agent.id}>
              <Card
                title={agent.agentName}
                extra={
                  <Link to={`/${i18n.language}/admin/enterprise/c-end-chat/agents/${agent.id}`}>
                    <Button type="primary">{t('c_end_admin_manage_agent')}</Button>
                  </Link>
                }
                style={{ borderRadius: 20, height: '100%' }}
              >
                <Typography.Paragraph>
                  {agent.description || t('c_end_admin_no_description')}
                </Typography.Paragraph>
                <Typography.Paragraph type="secondary" style={{ marginBottom: 0 }}>
                  {t('c_end_admin_card_hint')}
                </Typography.Paragraph>
              </Card>
            </Col>
          ))}
        </Row>
      ) : (
        <Card variant="borderless" style={{ borderRadius: 20 }}>
          <Empty description={t('c_end_admin_empty')} image={Empty.PRESENTED_IMAGE_SIMPLE} />
        </Card>
      )}
    </Flex>
  );
}
