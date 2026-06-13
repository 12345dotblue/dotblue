import { useEffect, useState } from 'react';
import { Alert, Button, Card, Flex, Input, Spin, Typography } from 'antd';
import { useTranslation } from 'react-i18next';
import { useNavigate, useParams } from 'react-router-dom';
import { resolveShare, verifyShare } from '../services/cEndChatApi';
import { saveCEndSessionToken } from '../services/cEndSession';

export function CEndSharePage() {
  const { t } = useTranslation();
  const { shareCode = '' } = useParams();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [verifying, setVerifying] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [password, setPassword] = useState('');
  const [resolved, setResolved] = useState<any>(null);

  useEffect(() => {
    let mounted = true;
    resolveShare(shareCode)
      .then((data) => {
        if (!mounted) return;
        setResolved(data);
        setError(null);
      })
      .catch((err) => {
        if (!mounted) return;
        setError(err instanceof Error ? err.message : t('c_end_share_resolve_failed'));
      })
      .finally(() => {
        if (mounted) setLoading(false);
      });
    return () => {
      mounted = false;
    };
  }, [shareCode, t]);

  if (loading) {
    return <Spin />;
  }

  const handleVerify = async () => {
    setVerifying(true);
    setError(null);
    try {
      const data = await verifyShare(shareCode, password);
      saveCEndSessionToken(resolved.agent.id, {
        token: data.sessionToken,
        allowFileUpload: data.allowFileUpload,
        conversationId: resolved?.conversation?.id,
        agentName: resolved?.agent?.agentName,
      }, 'share');
      navigate(`../agents/${resolved.agent.id}/chat`, { replace: true, state: { accessMode: 'share' as const } });
    } catch (err) {
      setError(err instanceof Error ? err.message : t('c_end_share_verify_failed'));
    } finally {
      setVerifying(false);
    }
  };

  return (
    <Flex vertical gap={16} style={{ width: '100%', maxWidth: 720, margin: '0 auto', padding: 24 }}>
      <Typography.Title level={2}>{t('c_end_share_title')}</Typography.Title>
      {error ? <Alert type="error" message={error} /> : null}
      <Card>
        <Typography.Paragraph>{resolved?.agent?.agentName}</Typography.Paragraph>
        {resolved?.requiresPassword ? (
          <>
            <Input.Password
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              placeholder={t('c_end_share_password_placeholder')}
              disabled={verifying}
              onPressEnter={verifying ? undefined : handleVerify}
            />
            <Button
              style={{ marginTop: 12 }}
              type="primary"
              loading={verifying}
              disabled={verifying}
              onClick={handleVerify}
            >
              {t('c_end_share_verify')}
            </Button>
          </>
        ) : (
          <Button
            type="primary"
            loading={verifying}
            disabled={verifying}
            onClick={handleVerify}
          >
            {t('c_end_share_verify')}
          </Button>
        )}
      </Card>
    </Flex>
  );
}
