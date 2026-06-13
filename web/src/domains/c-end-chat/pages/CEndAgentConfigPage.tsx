import { CopyOutlined, LinkOutlined, QuestionCircleOutlined } from '@ant-design/icons';
import { useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, Form, Input, List, Select, Space, Switch, Tag, Tooltip, Typography, message } from 'antd';
import { useTranslation } from 'react-i18next';
import { useParams } from 'react-router-dom';
import {
  createEmbedToken,
  createShareLink,
  getAgentEntry,
  listAgents,
  revokeShareLink,
  saveAgentEntry,
  saveEmbedConfig,
  type AgentEntryConfig,
  type EmbedConfig,
  type ShareLink,
} from '../services/cEndChatApi';

const defaultConfig: AgentEntryConfig = {
  enabled: false,
  defaultAccessMode: 'standalone',
  allowAnonymous: false,
  allowFileUpload: false,
  themeMode: 'auto',
  compactHeader: false,
  sessionTtlSeconds: 900,
  refreshBeforeSeconds: 120,
};

const defaultEmbedConfig: EmbedConfig = {
  allowedOrigins: [],
  themeMode: 'auto',
  compactHeader: true,
  allowFileUpload: false,
};

export function CEndAgentConfigPage() {
  const { t, i18n } = useTranslation();
  const { agentId = '' } = useParams();
  const [configForm] = Form.useForm<AgentEntryConfig>();
  const [embedForm] = Form.useForm<{ allowedOrigins: string; themeMode: EmbedConfig['themeMode']; compactHeader: boolean; allowFileUpload: boolean }>();
  const [shareForm] = Form.useForm<{ password?: string; allowContinueChat: boolean; allowAnonymous: boolean }>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [shareLinks, setShareLinks] = useState<ShareLink[]>([]);
  const [createdShareUrl, setCreatedShareUrl] = useState('');
  const [embedToken, setEmbedToken] = useState('');
  const [agentName, setAgentName] = useState('');

  const load = async () => {
    if (!agentId) return;
    setLoading(true);
    try {
      const [data, agents] = await Promise.all([getAgentEntry(agentId), listAgents()]);
      configForm.setFieldsValue({ ...defaultConfig, ...data.config });
      const embedConfig = data.embedConfig ?? defaultEmbedConfig;
      embedForm.setFieldsValue({
        allowedOrigins: embedConfig.allowedOrigins.join('\n'),
        themeMode: embedConfig.themeMode,
        compactHeader: embedConfig.compactHeader,
        allowFileUpload: embedConfig.allowFileUpload,
      });
      setShareLinks(data.shareLinks ?? []);
      setAgentName(agents.find((item) => item.id === agentId)?.agentName || '');
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('c_end_agent_load_failed'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, [agentId, t]);

  const shareBaseUrl = useMemo(() => window.location.origin, []);
  const standaloneUrl = useMemo(
    () => `${shareBaseUrl}/${i18n.language}/agents/${agentId}/chat`,
    [agentId, i18n.language, shareBaseUrl],
  );
  const embedUrl = useMemo(
    () => (embedToken ? `${shareBaseUrl}/${i18n.language}/embed/agents/${agentId}?token=${encodeURIComponent(embedToken)}` : ''),
    [agentId, embedToken, i18n.language, shareBaseUrl],
  );
  const accessModeOptions = useMemo(() => ([
    { label: t('c_end_agent_access_mode_standalone'), value: 'standalone' },
    { label: t('c_end_agent_access_mode_share'), value: 'share' },
    { label: t('c_end_agent_access_mode_embed'), value: 'embed' },
  ]), [t]);
  const themeModeOptions = useMemo(() => ([
    { label: t('c_end_agent_theme_auto'), value: 'auto' },
    { label: t('c_end_agent_theme_light'), value: 'light' },
    { label: t('c_end_agent_theme_dark'), value: 'dark' },
  ]), [t]);
  const statusLabel = (status: string) => t(`c_end_agent_share_status_${status}`, { defaultValue: status });
  const renderHelpLabel = (labelKey: string, helpKey: string) => (
    <Space size={4}>
      <span>{t(labelKey)}</span>
      <Tooltip title={t(helpKey)}>
        <QuestionCircleOutlined />
      </Tooltip>
    </Space>
  );
  const copyText = async (value: string) => {
    try {
      await navigator.clipboard.writeText(value);
      message.success(t('c_end_agent_copy_success'));
    } catch {
      message.error(t('c_end_agent_copy_failed'));
    }
  };
  const getShareUrl = (shareCode: string) => `${shareBaseUrl}/${i18n.language}/share/${shareCode}`;
  const ensureCurrentOriginAllowed = async () => {
    const values = embedForm.getFieldsValue();
    const currentOrigin = window.location.origin.toLowerCase();
    const currentOrigins = (values.allowedOrigins || '')
      .split('\n')
      .map((item) => item.trim().toLowerCase())
      .filter(Boolean);
    if (currentOrigins.includes(currentOrigin)) {
      return;
    }
    const nextOrigins = [...currentOrigins, currentOrigin];
    embedForm.setFieldValue('allowedOrigins', nextOrigins.join('\n'));
    await saveEmbedConfig(agentId, {
      allowedOrigins: nextOrigins,
      themeMode: values.themeMode,
      compactHeader: values.compactHeader,
      allowFileUpload: values.allowFileUpload,
    });
  };

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <div>
        <Typography.Title level={2}>
          {agentName ? t('c_end_agent_title_with_name', { name: agentName }) : t('c_end_agent_title')}
        </Typography.Title>
        <Typography.Paragraph type="secondary" style={{ marginBottom: 0 }}>
          {t('c_end_agent_manage_hint')}
        </Typography.Paragraph>
      </div>
      {error ? <Alert type="error" title={error} /> : null}

      <Card loading={loading} title={t('c_end_agent_section_access')}>
        <Alert
          type="info"
          showIcon
          message={t('c_end_agent_access_guide_title')}
          description={(
            <Space direction="vertical" size={4}>
              <Typography.Text>{t('c_end_agent_access_mode_standalone_desc')}</Typography.Text>
              <Typography.Text>{t('c_end_agent_access_mode_share_desc')}</Typography.Text>
              <Typography.Text>{t('c_end_agent_access_mode_embed_desc')}</Typography.Text>
            </Space>
          )}
          style={{ marginBottom: 16 }}
        />
        <Space direction="vertical" size={8} style={{ width: '100%', marginBottom: 16 }}>
          <Typography.Text strong>{t('c_end_agent_standalone_entry_url')}</Typography.Text>
          <Typography.Paragraph copyable={{ text: standaloneUrl }} style={{ marginBottom: 0 }}>
            {standaloneUrl}
          </Typography.Paragraph>
          <Space wrap>
            <Button icon={<CopyOutlined />} onClick={() => void copyText(standaloneUrl)}>
              {t('c_end_agent_copy_link')}
            </Button>
            <Button icon={<LinkOutlined />} href={standaloneUrl} target="_blank" rel="noreferrer">
              {t('c_end_agent_open_link')}
            </Button>
          </Space>
        </Space>
        <Form form={configForm} layout="vertical" initialValues={defaultConfig} onFinish={async (values) => {
          await saveAgentEntry(agentId, values);
          message.success(t('c_end_agent_save'));
        }}>
          <Form.Item name="enabled" label={t('c_end_agent_enabled')} valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="defaultAccessMode" label={renderHelpLabel('c_end_agent_default_access_mode', 'c_end_agent_default_access_mode_help')}>
            <Select options={accessModeOptions} />
          </Form.Item>
          <Form.Item name="allowAnonymous" label={renderHelpLabel('c_end_agent_allow_anonymous', 'c_end_agent_allow_anonymous_help')} valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="allowFileUpload" label={t('c_end_agent_allow_file_upload')} valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="compactHeader" label={t('c_end_agent_compact_header')} valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="themeMode" label={t('c_end_agent_theme_mode')}>
            <Select options={themeModeOptions} />
          </Form.Item>
          <Button type="primary" htmlType="submit">{t('c_end_agent_save')}</Button>
        </Form>
      </Card>

      <Card loading={loading} title={t('c_end_agent_section_embed')}>
        <Alert
          type="info"
          showIcon
          message={t('c_end_agent_embed_guide_title')}
          description={t('c_end_agent_embed_guide_desc')}
          style={{ marginBottom: 16 }}
        />
        <Form
          form={embedForm}
          layout="vertical"
          initialValues={{ ...defaultEmbedConfig, allowedOrigins: '' }}
          onFinish={async (values) => {
            await saveEmbedConfig(agentId, {
              allowedOrigins: values.allowedOrigins.split('\n').map((item) => item.trim()).filter(Boolean),
              themeMode: values.themeMode,
              compactHeader: values.compactHeader,
              allowFileUpload: values.allowFileUpload,
            });
            message.success(t('c_end_agent_save'));
            await load();
          }}
        >
          <Form.Item name="allowedOrigins" label={t('c_end_agent_allowed_origins')}>
            <Input.TextArea rows={4} placeholder={t('c_end_agent_allowed_origins_placeholder')} />
          </Form.Item>
          <Form.Item name="themeMode" label={t('c_end_agent_theme_mode')}>
            <Select options={themeModeOptions} />
          </Form.Item>
          <Form.Item name="compactHeader" label={t('c_end_agent_compact_header')} valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="allowFileUpload" label={t('c_end_agent_allow_file_upload')} valuePropName="checked">
            <Switch />
          </Form.Item>
          <Space>
            <Button type="primary" htmlType="submit">{t('c_end_agent_save')}</Button>
            <Button onClick={async () => {
              try {
                await ensureCurrentOriginAllowed();
                const origin = window.location.origin;
                const result = await createEmbedToken(agentId, origin);
                setEmbedToken(result.embedToken);
                message.success(t('c_end_agent_embed_token_created'));
              } catch (err) {
                message.error(err instanceof Error ? err.message : t('c_end_agent_embed_token_create_failed'));
              }
            }}>
              {t('c_end_agent_create_embed_token')}
            </Button>
          </Space>
          {embedToken ? (
            <Space direction="vertical" size={8} style={{ width: '100%', marginTop: 16 }}>
              <Typography.Text strong>{t('c_end_agent_generated_embed_url')}</Typography.Text>
              <Typography.Paragraph copyable={{ text: embedUrl }} style={{ marginBottom: 0 }}>
                {embedUrl}
              </Typography.Paragraph>
              <Space wrap>
                <Button icon={<CopyOutlined />} onClick={() => void copyText(embedUrl)}>
                  {t('c_end_agent_copy_link')}
                </Button>
                <Button icon={<LinkOutlined />} href={embedUrl} target="_blank" rel="noreferrer">
                  {t('c_end_agent_open_link')}
                </Button>
              </Space>
            </Space>
          ) : null}
        </Form>
      </Card>

      <Card title={t('c_end_agent_section_share')}>
        <Alert
          type="info"
          showIcon
          message={t('c_end_agent_share_guide_title')}
          description={t('c_end_agent_share_guide_desc')}
          style={{ marginBottom: 16 }}
        />
        <Form
          form={shareForm}
          layout="vertical"
          initialValues={{ allowContinueChat: false, allowAnonymous: false }}
          onFinish={async (values) => {
            try {
              const result = await createShareLink({ agentId, ...values });
              setCreatedShareUrl(`${shareBaseUrl}${result.shareUrl}`);
              message.success(values.password ? t('c_end_agent_share_created_with_password') : t('c_end_agent_share_created'));
              shareForm.resetFields();
              shareForm.setFieldsValue({ allowContinueChat: false, allowAnonymous: false });
              await load();
            } catch (err) {
              message.error(err instanceof Error ? err.message : t('c_end_agent_share_create_failed'));
            }
          }}
        >
          <Form.Item name="password" label={t('c_end_agent_password')} extra={t('c_end_agent_password_help')}>
            <Input.Password />
          </Form.Item>
          <Form.Item name="allowContinueChat" label={t('c_end_agent_allow_continue_chat')} valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="allowAnonymous" label={t('c_end_agent_allow_anonymous')} valuePropName="checked">
            <Switch />
          </Form.Item>
          <Button type="primary" htmlType="submit">{t('c_end_agent_create_share')}</Button>
        </Form>
        {createdShareUrl ? (
          <Space direction="vertical" size={8} style={{ width: '100%', marginTop: 16 }}>
            <Typography.Text strong>{t('c_end_agent_generated_share_url')}</Typography.Text>
            <Typography.Paragraph copyable={{ text: createdShareUrl }} style={{ marginBottom: 0 }}>
              {createdShareUrl}
            </Typography.Paragraph>
            <Space wrap>
              <Button icon={<CopyOutlined />} onClick={() => void copyText(createdShareUrl)}>
                {t('c_end_agent_copy_link')}
              </Button>
              <Button icon={<LinkOutlined />} href={createdShareUrl} target="_blank" rel="noreferrer">
                {t('c_end_agent_open_link')}
              </Button>
            </Space>
          </Space>
        ) : null}
        <List
          style={{ marginTop: 16 }}
          dataSource={shareLinks}
          renderItem={(item) => (
            <List.Item
              actions={[
                <Button key="copy" icon={<CopyOutlined />} onClick={() => void copyText(getShareUrl(item.shareCode))}>
                  {t('c_end_agent_copy_link')}
                </Button>,
                <Button key="open" icon={<LinkOutlined />} href={getShareUrl(item.shareCode)} target="_blank" rel="noreferrer">
                  {t('c_end_agent_open_link')}
                </Button>,
                <Button danger key="revoke" onClick={async () => {
                  await revokeShareLink(item.id);
                  await load();
                }}>
                  {t('c_end_agent_revoke_share')}
                </Button>,
              ]}
            >
              <List.Item.Meta
                title={(
                  <Space wrap>
                    <Typography.Text strong>
                      {t('c_end_agent_share_link_title', { code: item.shareCode.slice(0, 8) })}
                    </Typography.Text>
                    <Tag>{statusLabel(item.status)}</Tag>
                    {item.hasPassword ? <Tag color="orange">{t('c_end_agent_tag_password_required')}</Tag> : null}
                    {item.allowContinueChat ? <Tag color="blue">{t('c_end_agent_tag_continue_chat')}</Tag> : null}
                    {item.allowAnonymous ? <Tag color="green">{t('c_end_agent_tag_anonymous')}</Tag> : <Tag>{t('c_end_agent_tag_restricted')}</Tag>}
                  </Space>
                )}
                description={(
                  <Space direction="vertical" size={4}>
                    <Typography.Text type="secondary">{getShareUrl(item.shareCode)}</Typography.Text>
                    <Typography.Text type="secondary">
                      {item.expiresAt
                        ? t('c_end_agent_share_link_expires', { value: item.expiresAt })
                        : t('c_end_agent_share_link_no_expiry')}
                    </Typography.Text>
                  </Space>
                )}
              />
            </List.Item>
          )}
        />
      </Card>
    </Space>
  );
}
