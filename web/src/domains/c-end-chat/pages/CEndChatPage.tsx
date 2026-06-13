import { Alert, Spin } from 'antd';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useLocation, useParams } from 'react-router-dom';
import ChatPage from '../../chat/ChatPage';
import { getOrCreateProvider } from '../../chat/SSEChatProvider';
import { BACKEND_URL } from '../../../config';
import { createConversation, createStandaloneSession, getConversationMessages, getPublicSessionHeaders, uploadPublicFile } from '../services/cEndChatApi';
import { loadCEndSessionState, saveCEndSessionToken, type CEndAccessMode } from '../services/cEndSession';

interface CEndChatLocationState {
  accessMode?: CEndAccessMode;
}

export function CEndChatPage() {
  const { t } = useTranslation();
  const { agentId = '' } = useParams();
  const location = useLocation();
  const locationState = (location.state as CEndChatLocationState | null) || null;
  const accessMode: CEndAccessMode = locationState?.accessMode || 'standalone';
  const [sessionState, setSessionState] = useState(() => loadCEndSessionState(agentId, accessMode));
  const [loading, setLoading] = useState(!loadCEndSessionState(agentId, accessMode));
  const [error, setError] = useState<string | null>(null);
  const sessionToken = sessionState?.token;

  useEffect(() => {
    const existing = loadCEndSessionState(agentId, accessMode);
    if (existing) {
      setSessionState(existing);
      setLoading(false);
      setError(null);
      return;
    }
    if (accessMode !== 'standalone') {
      setLoading(false);
      setError(t('c_end_chat_session_missing'));
      return;
    }
    let cancelled = false;
    setLoading(true);
    createStandaloneSession(agentId)
      .then((data) => {
        if (cancelled) return;
        const nextState = {
          token: data.sessionToken,
          allowFileUpload: data.allowFileUpload,
          agentName: data.agentName,
        };
        saveCEndSessionToken(agentId, nextState, 'standalone');
        setSessionState(nextState);
        setError(null);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(err instanceof Error ? err.message : t('c_end_chat_session_missing'));
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [agentId, accessMode, t]);

  if (loading) {
    return <Spin />;
  }

  if (error) {
    return <Alert type="error" message={t('c_end_chat_title')} description={error} />;
  }

  if (!sessionToken) {
    return <Alert type="warning" message={t('c_end_chat_title')} description={t('c_end_chat_session_missing')} />;
  }

  return (
    <ChatPage
      fixedAgentId={agentId}
      fixedAgentName={sessionState?.agentName || agentId}
      getJwt={() => sessionToken}
      authHeaders={() => getPublicSessionHeaders(sessionToken)}
      listAgents={async () => [{ id: agentId, agentName: sessionState?.agentName || agentId, engineType: 'hermes' }]}
      listConversations={async () => (
        sessionState?.conversationId
          ? {
              items: [{
                id: sessionState.conversationId,
                title: t('c_end_chat_shared_conversation'),
                agentId,
                agentName: sessionState?.agentName || agentId,
                updatedAt: new Date().toISOString(),
              }],
              hasMore: false,
              nextCursor: '',
            }
          : { items: [], hasMore: false, nextCursor: '' }
      )}
      createConversation={() => createConversation(sessionToken)}
      loadMessages={(conversationId) => getConversationMessages(sessionToken, conversationId)}
      uploadFile={(conversationId, file, kind) => uploadPublicFile(sessionToken, conversationId, file, kind)}
      createProvider={(targetAgentId, events) => getOrCreateProvider(targetAgentId, events, {
        cacheKey: `public:${targetAgentId}:${sessionToken.slice(-8)}`,
        requestUrl: `${BACKEND_URL}/api/public/c-end-chat/chat/completions`,
        getHeaders: () => getPublicSessionHeaders(sessionToken).headers,
      })}
      showSidebar={false}
      showDashboardButton={false}
      showUserMenu={false}
      allowDeleteConversation={false}
      allowFileUpload={sessionState?.allowFileUpload === true}
    />
  );
}
