package im

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkdispatcher "github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

var (
	ErrFeishuUnsupportedConnectionMode = errors.New("feishu only supports socket_mode in v1")
	ErrFeishuInvalidPayload            = errors.New("invalid feishu payload")
	ErrFeishuUnsupportedAttachment     = errors.New("unsupported feishu attachment type")
)

type FeishuAdapter struct {
	mu                  sync.RWMutex
	httpClient          *http.Client
	tokenCache          map[string]feishuTenantToken
	socketClientFactory feishuSocketClientFactory
	inboundProcessor    feishuInboundProcessor
	reconnectBaseDelay  time.Duration
	reconnectMaxDelay   time.Duration
}

type feishuTenantToken struct {
	Token     string
	ExpiresAt time.Time
}

type feishuSocketClient interface {
	Start(ctx context.Context) error
}

type feishuSocketClientFactory func(conn Connection, handler func(context.Context, []byte) error) feishuSocketClient

type feishuInboundProcessor func(ctx context.Context, conn Connection, events []InboundEvent) (inboundPersistResult, error)

type feishuSDKSocketClient struct {
	client *larkws.Client
}

func (c *feishuSDKSocketClient) Start(ctx context.Context) error {
	return c.client.Start(ctx)
}

func init() {
	RegisterAdapter(NewFeishuAdapter())
}

func NewFeishuAdapter() *FeishuAdapter {
	return &FeishuAdapter{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		tokenCache:          map[string]feishuTenantToken{},
		socketClientFactory: newFeishuSDKSocketClient,
		inboundProcessor:    defaultInboundPipeline.PersistEvents,
		reconnectBaseDelay:  time.Second,
		reconnectMaxDelay:   30 * time.Second,
	}
}

func (a *FeishuAdapter) Platform() string {
	return "feishu"
}

func (a *FeishuAdapter) Start(ctx context.Context, conn Connection) error {
	if err := a.TestConnection(ctx, conn); err != nil {
		return err
	}
	return defaultRuntimeManager.Start(ctx, conn, RuntimeHooks{
		Start: func(runtimeCtx context.Context, enqueue RuntimePayloadHandler) error {
			return a.startRuntime(runtimeCtx, conn, enqueue)
		},
		ProcessPayload:     a.processRuntimePayload(conn),
		IsExpectedStop:     isExpectedFeishuRuntimeStop,
		ReconnectBaseDelay: a.reconnectBaseDelay,
		ReconnectMaxDelay:  a.reconnectMaxDelay,
	})
}

func (a *FeishuAdapter) TestConnection(ctx context.Context, conn Connection) error {
	if err := a.ValidateConfig(feishuConfig(conn), firstNonEmptySecretMap(conn.Secrets, conn.SecretsMasked)); err != nil {
		return err
	}
	if _, err := a.getTenantAccessToken(ctx, conn); err != nil {
		return err
	}
	return nil
}

func (a *FeishuAdapter) Stop(ctx context.Context, connectionID string) error {
	a.mu.Lock()
	delete(a.tokenCache, connectionID)
	a.mu.Unlock()
	return defaultRuntimeManager.Stop(ctx, connectionID)
}

func (a *FeishuAdapter) ValidateConfig(config map[string]any, secrets map[string]any) error {
	if str(config["appId"]) == "" {
		return ErrInvalidConnectionConfig
	}
	if str(secrets["appSecret"]) == "" {
		return ErrInvalidConnectionConfig
	}
	mode := str(config["connectionMode"])
	if mode != "" && mode != "socket_mode" {
		return ErrFeishuUnsupportedConnectionMode
	}
	if domain := str(config["domain"]); domain != "" && domain != "feishu" && domain != "lark" {
		return ErrInvalidConnectionConfig
	}
	return nil
}

func feishuConfig(conn Connection) map[string]any {
	config := map[string]any{}
	for k, v := range conn.Config {
		config[k] = v
	}
	if strings.TrimSpace(str(config["connectionMode"])) == "" && strings.TrimSpace(conn.ConnectionMode) != "" {
		config["connectionMode"] = conn.ConnectionMode
	}
	return config
}

func newFeishuSDKSocketClient(conn Connection, handler func(context.Context, []byte) error) feishuSocketClient {
	config := feishuConfig(conn)
	secrets := firstNonEmptySecretMap(conn.Secrets, conn.SecretsMasked)

	dispatcher := larkdispatcher.NewEventDispatcher("", "").
		OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
			return handleFeishuSDKEvent(ctx, event, handler)
		})

	domain := lark.FeishuBaseUrl
	if str(config["domain"]) == "lark" {
		domain = lark.LarkBaseUrl
	}

	client := larkws.NewClient(
		str(config["appId"]),
		str(secrets["appSecret"]),
		larkws.WithEventHandler(dispatcher),
		larkws.WithDomain(domain),
	)
	return &feishuSDKSocketClient{client: client}
}

func handleFeishuSDKEvent(ctx context.Context, event *larkim.P2MessageReceiveV1, handler func(context.Context, []byte) error) error {
	if event == nil {
		return ErrFeishuInvalidPayload
	}
	var payload []byte
	if event.EventReq != nil && len(event.EventReq.Body) > 0 {
		payload = append(payload, event.EventReq.Body...)
	} else {
		var err error
		payload, err = json.Marshal(event)
		if err != nil {
			return err
		}
	}
	return handler(ctx, payload)
}

func (a *FeishuAdapter) processRuntimePayload(conn Connection) RuntimePayloadHandler {
	return func(ctx context.Context, payload []byte) error {
		return a.processFeishuSocketPayload(ctx, conn, payload)
	}
}

func (a *FeishuAdapter) processFeishuSocketPayload(ctx context.Context, conn Connection, payload []byte) error {
	events, err := a.ParseInbound(ctx, payload)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return nil
	}
	processor := a.inboundProcessor
	if processor == nil {
		return fmt.Errorf("feishu inbound processor is not configured")
	}
	_, err = processor(ctx, conn, events)
	return err
}

func isExpectedFeishuRuntimeStop(err error) bool {
	if err == nil {
		return true
	}
	if errors.Is(err, context.Canceled) {
		return true
	}
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(msg, "context canceled") ||
		strings.Contains(msg, "context cancelled") ||
		strings.Contains(msg, "operation was canceled")
}

func (a *FeishuAdapter) startRuntime(ctx context.Context, conn Connection, enqueue RuntimePayloadHandler) error {
	if a.socketClientFactory == nil {
		return fmt.Errorf("feishu socket client factory is not configured")
	}
	client := a.socketClientFactory(conn, enqueue)
	return client.Start(ctx)
}

func (a *FeishuAdapter) ParseInbound(ctx context.Context, raw any) ([]InboundEvent, error) {
	payload, err := coerceRawMap(raw)
	if err != nil {
		return nil, ErrFeishuInvalidPayload
	}

	header, _ := payload["header"].(map[string]any)
	event, _ := payload["event"].(map[string]any)
	sender, _ := event["sender"].(map[string]any)
	senderIDBlock, _ := sender["sender_id"].(map[string]any)
	message, _ := event["message"].(map[string]any)
	mentions, _ := event["mentions"].([]any)

	chatType := str(message["chat_type"])
	chatID := str(message["chat_id"])
	messageID := str(message["message_id"])
	messageType := firstNonEmpty(str(message["message_type"]), str(message["msg_type"]), "text")
	threadID := str(message["thread_id"])
	rootID := str(message["root_id"])
	if threadID == "" {
		threadID = rootID
	}

	parsed := parseFeishuContent(messageType, str(message["content"]), "feishu")
	if parsed.Text == "" && event["text"] != nil {
		parsed.Text = str(event["text"])
		if parsed.Text != "" && len(parsed.Segments) == 0 {
			parsed.Segments = []RichSegment{{Type: "text", Text: parsed.Text}}
		}
	}

	inbound := InboundEvent{
		Platform:         "feishu",
		EventID:          str(header["event_id"]),
		MessageID:        messageID,
		ExternalChatID:   chatID,
		ExternalThreadID: threadID,
		ExternalUserID:   firstNonEmpty(str(senderIDBlock["open_id"]), str(senderIDBlock["union_id"]), str(senderIDBlock["user_id"])),
		ChatType:         chatType,
		MentionsBot:      len(mentions) > 0,
		Text:             parsed.Text,
		Segments:         parsed.Segments,
		Attachments:      parsed.Attachments,
		ReplyHandle: map[string]any{
			"message_id":   messageID,
			"chat_id":      chatID,
			"thread_id":    threadID,
			"root_id":      rootID,
			"message_type": messageType,
		},
		ReceivedAt: time.Now(),
	}
	addr := BuildSessionAddress(Connection{Platform: "feishu"}, inbound)
	inbound.ReplyHandle["session_key"] = BuildSessionKey("", "", addr)
	inbound.ReplyHandle["session_address"] = map[string]any{
		"platform":      addr.Platform,
		"connection_id": addr.ConnectionID,
		"chat_id":       addr.ChatID,
		"thread_id":     addr.ThreadID,
		"user_id":       addr.UserID,
		"chat_type":     addr.ChatType,
	}

	if inbound.EventID == "" {
		inbound.EventID = messageID
	}
	if inbound.EventID == "" || inbound.ExternalChatID == "" {
		return nil, ErrFeishuInvalidPayload
	}

	if rawBytes, err := json.Marshal(payload); err == nil {
		inbound.RawPayload = rawBytes
	}
	return []InboundEvent{inbound}, nil
}

func (a *FeishuAdapter) SendOutbound(ctx context.Context, conn Connection, msg OutboundEnvelope) error {
	if err := a.ValidateConfig(conn.Config, firstNonEmptySecretMap(conn.Secrets, conn.SecretsMasked)); err != nil {
		return err
	}

	token, err := a.getTenantAccessToken(ctx, conn)
	if err != nil {
		return err
	}

	outbounds, err := buildFeishuOutboundRequests(msg)
	if err != nil {
		return err
	}
	if len(outbounds) == 0 {
		return fmt.Errorf("feishu outbound payload is empty")
	}

	messageID := str(msg.ReplyHandle["message_id"])
	for _, outbound := range outbounds {
		if err := a.sendFeishuOutboundMessage(ctx, conn, token, msg.ExternalChatID, messageID, outbound); err != nil {
			return err
		}
	}
	return nil
}

func coerceRawMap(raw any) (map[string]any, error) {
	switch v := raw.(type) {
	case map[string]any:
		return v, nil
	case []byte:
		var out map[string]any
		if err := json.Unmarshal(v, &out); err != nil {
			return nil, err
		}
		return out, nil
	case string:
		var out map[string]any
		if err := json.Unmarshal([]byte(v), &out); err != nil {
			return nil, err
		}
		return out, nil
	default:
		return nil, ErrFeishuInvalidPayload
	}
}

func extractFeishuText(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return content
	}
	if text := str(parsed["text"]); text != "" {
		return text
	}
	return content
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func firstNonEmptySecretMap(values ...map[string]any) map[string]any {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return map[string]any{}
}

func (a *FeishuAdapter) getTenantAccessToken(ctx context.Context, conn Connection) (string, error) {
	a.mu.RLock()
	cached, ok := a.tokenCache[conn.ID]
	a.mu.RUnlock()
	if ok && cached.Token != "" && time.Until(cached.ExpiresAt) > time.Minute {
		return cached.Token, nil
	}

	respBody, err := a.doFeishuJSONRequest(
		ctx,
		http.MethodPost,
		feishuTenantTokenURL(conn.Config),
		"",
		map[string]any{
			"app_id":     str(conn.Config["appId"]),
			"app_secret": str(firstNonEmptySecretMap(conn.Secrets, conn.SecretsMasked)["appSecret"]),
		},
	)
	if err != nil {
		return "", err
	}

	var payload struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
		Expire            int    `json:"expire"`
		ExpiresIn         int    `json:"expires_in"`
	}
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return "", err
	}
	if payload.Code != 0 {
		return "", fmt.Errorf("feishu token error: %s", firstNonEmpty(payload.Msg, string(respBody)))
	}
	if payload.TenantAccessToken == "" {
		return "", fmt.Errorf("feishu tenant access token missing")
	}

	expireSeconds := payload.ExpiresIn
	if expireSeconds <= 0 {
		expireSeconds = payload.Expire
	}
	if expireSeconds <= 0 {
		expireSeconds = 7200
	}

	a.mu.Lock()
	a.tokenCache[conn.ID] = feishuTenantToken{
		Token:     payload.TenantAccessToken,
		ExpiresAt: time.Now().Add(time.Duration(expireSeconds) * time.Second),
	}
	a.mu.Unlock()
	return payload.TenantAccessToken, nil
}

func (a *FeishuAdapter) doFeishuJSONRequest(
	ctx context.Context,
	method string,
	requestURL string,
	token string,
	body map[string]any,
) ([]byte, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("feishu http error: %s", strings.TrimSpace(string(respBody)))
	}

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(respBody, &result); err == nil && result.Code != 0 {
		return nil, fmt.Errorf("feishu api error: %s", firstNonEmpty(result.Msg, string(respBody)))
	}

	return respBody, nil
}

func formatFeishuOutboundText(msg OutboundEnvelope) string {
	text := strings.TrimSpace(msg.Text)
	if text == "" {
		text = strings.TrimSpace(segmentsPlainText(msg.Segments))
	}

	var lines []string
	if text != "" {
		lines = append(lines, text)
	}
	for _, attachment := range msg.Attachments {
		label := firstNonEmpty(attachment.Name, attachment.Type, "attachment")
		lines = append(lines, "["+label+"]")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func buildFeishuOutboundRequests(msg OutboundEnvelope) ([]feishuOutboundMessage, error) {
	outbounds := make([]feishuOutboundMessage, 0, 1+len(msg.Attachments))
	if shouldSendFeishuTextBody(msg) {
		textMsg := msg
		textMsg.Attachments = nil
		outbound, err := buildFeishuOutboundMessage(textMsg)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(outbound.Content) != "" {
			outbounds = append(outbounds, outbound)
		}
	}
	for _, attachment := range msg.Attachments {
		outbound, err := buildFeishuAttachmentMessage(attachment)
		if err != nil {
			return nil, err
		}
		outbounds = append(outbounds, outbound)
	}
	return outbounds, nil
}

func shouldSendFeishuTextBody(msg OutboundEnvelope) bool {
	if strings.TrimSpace(msg.Text) != "" {
		return true
	}
	if len(msg.Segments) == 0 {
		return false
	}
	if strings.TrimSpace(segmentsPlainText(msg.Segments)) != "" {
		return true
	}
	textMsg := msg
	textMsg.Attachments = nil
	return shouldUseFeishuPost(textMsg)
}

func buildFeishuAttachmentMessage(attachment OutboundAttachment) (feishuOutboundMessage, error) {
	refPlatform, refType, remoteID, ok := ParseMediaRef(attachment.MediaRef)
	if !ok {
		return feishuOutboundMessage{}, fmt.Errorf("feishu attachment media ref is invalid")
	}
	if refPlatform != "feishu" {
		return feishuOutboundMessage{}, fmt.Errorf("feishu attachment media ref platform mismatch")
	}

	attachmentType := strings.TrimSpace(attachment.Type)
	if attachmentType == "" {
		attachmentType = refType
	}
	if attachmentType != refType {
		return feishuOutboundMessage{}, fmt.Errorf("feishu attachment type mismatch")
	}

	var content []byte
	var err error
	switch attachmentType {
	case "image":
		content, err = json.Marshal(map[string]string{"image_key": remoteID})
	case "file":
		content, err = json.Marshal(map[string]string{"file_key": remoteID})
	default:
		return feishuOutboundMessage{}, ErrFeishuUnsupportedAttachment
	}
	if err != nil {
		return feishuOutboundMessage{}, err
	}
	return feishuOutboundMessage{
		MsgType: attachmentType,
		Content: string(content),
	}, nil
}

func (a *FeishuAdapter) sendFeishuOutboundMessage(
	ctx context.Context,
	conn Connection,
	token string,
	externalChatID string,
	replyMessageID string,
	outbound feishuOutboundMessage,
) error {
	if strings.TrimSpace(outbound.Content) == "" {
		return fmt.Errorf("feishu outbound content is empty")
	}
	if replyMessageID != "" {
		_, err := a.doFeishuJSONRequest(
			ctx,
			http.MethodPost,
			feishuReplyMessageURL(conn.Config, replyMessageID),
			token,
			map[string]any{
				"content":  outbound.Content,
				"msg_type": outbound.MsgType,
				"uuid":     uuid.NewString(),
			},
		)
		return err
	}
	if strings.TrimSpace(externalChatID) == "" {
		return fmt.Errorf("feishu outbound chat id is empty")
	}
	_, err := a.doFeishuJSONRequest(
		ctx,
		http.MethodPost,
		feishuCreateMessageURL(conn.Config),
		token,
		map[string]any{
			"receive_id": externalChatID,
			"content":    outbound.Content,
			"msg_type":   outbound.MsgType,
			"uuid":       uuid.NewString(),
		},
	)
	return err
}

func feishuAPIBaseURL(config map[string]any) string {
	if custom := strings.TrimSpace(str(config["apiBaseURL"])); custom != "" {
		return strings.TrimRight(custom, "/")
	}
	if strings.TrimSpace(str(config["domain"])) == "lark" {
		return "https://open.larksuite.com"
	}
	return "https://open.feishu.cn"
}

func feishuTenantTokenURL(config map[string]any) string {
	return feishuAPIBaseURL(config) + "/open-apis/auth/v3/tenant_access_token/internal"
}

func feishuCreateMessageURL(config map[string]any) string {
	return feishuAPIBaseURL(config) + "/open-apis/im/v1/messages?receive_id_type=chat_id"
}

func feishuReplyMessageURL(config map[string]any, messageID string) string {
	return feishuAPIBaseURL(config) + "/open-apis/im/v1/messages/" + url.PathEscape(strings.TrimSpace(messageID)) + "/reply"
}
