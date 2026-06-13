import { useEffect, useState } from 'react';
import { Alert, Flex, Spin, Typography } from 'antd';
import { useTranslation } from 'react-i18next';
import { useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { exchangeEmbedSession } from '../services/cEndChatApi';
import { saveCEndSessionToken } from '../services/cEndSession';

export function CEndEmbedPage() {
  const { t } = useTranslation();
  const { agentId = '' } = useParams();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const embedToken = searchParams.get('token') ?? '';
    if (!embedToken) {
      setError(t('c_end_embed_token_missing'));
      setLoading(false);
      return;
    }
    exchangeEmbedSession(embedToken, agentId)
      .then((data) => {
        saveCEndSessionToken(agentId, {
          token: data.sessionToken,
          allowFileUpload: data.allowFileUpload,
        }, 'embed');
        setError(null);
        navigate(`../../agents/${agentId}/chat`, { replace: true, state: { accessMode: 'embed' as const } });
      })
      .catch((err) => {
        setError(err instanceof Error ? err.message : t('c_end_embed_exchange_failed'));
      })
      .finally(() => setLoading(false));
  }, [agentId, searchParams, t, navigate]);

  if (loading) {
    return <Spin />;
  }

  return (
    <Flex vertical gap={16} style={{ width: '100%', maxWidth: 720, margin: '0 auto', padding: 24 }}>
      <Typography.Title level={2}>{t('c_end_embed_title')}</Typography.Title>
      {error ? <Alert type="error" message={error} /> : null}
    </Flex>
  );
}
