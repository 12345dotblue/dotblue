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

	"github.com/gorilla/websocket"
)

const (
	PlatformQQ          = "qq"
	QQConnectionMode    = "gateway"
	defaultQQAPIBaseURL = "https://api.sgroup.qq.com"
	defaultQQTokenURL   = "https://bots.qq.com/app/getAppAccessToken"
	defaultQQIntents    = 1 << 25
)

var (
	ErrQQInvalidPayload            = errors.New("invalid qq payload")
	ErrQQUnsupportedConnectionMode = errors.New("qq only supports gateway mode in v1")
	ErrQQGatewayURLMissing         = errors.New("qq gateway url is empty")
	errQQReconnectRequested        = errors.New("qq gateway requested reconnect")
	errQQInvalidSession            = errors.New("qq gateway invalid session")
)

type QQAdapter struct {
	mu                 sync.RWMutex
	httpClient         *http.Client
	wsDialer           *websocket.Dialer
	tokenCache         map[string]qqAccessToken
	identityCache      map[string]qqIdentity
	inboundProcessor   feishuInboundProcessor
	reconnectBaseDelay time.Duration
	reconnectMaxDelay  time.Duration
}

type qqAccessToken struct {
	Value  string
	Expiry time.Time
}

type qqIdentity struct {
	AppID string
}

type qqGatewayResponse struct {
	URL string `json:"url"`
}

type qqGatewayPayload struct {
	ID string          `json:"id,omitempty"`
	Op int             `json:"op"`
	S  *int64          `json:"s,omitempty"`
	T  string          `json:"t,omitempty"`
	D  json.RawMessage `json:"d,omitempty"`
}

type qqHelloPayload struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

type qqReadyPayload struct {
	SessionID string `json:"session_id"`
	User      struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	} `json:"user"`
}

type qqMessageEvent struct {
	ID          string                `json:"id"`
	Content     string                `json:"content"`
	GroupOpenID string                `json:"group_openid"`
	Timestamp   string                `json:"timestamp"`
	Author      qqMessageAuthor       `json:"author"`
	Attachments []qqMessageAttachment `json:"attachments"`
}

type qqMessageAuthor struct {
	UserOpenID   string `json:"user_openid"`
	MemberOpenID string `json:"member_openid"`
}

type qqMessageAttachment struct {
	URL          string `json:"url"`
	FileName     string `json:"filename"`
	ContentType  string `json:"content_type"`
	VoiceWavURL  string `json:"voice_wav_url"`
	ASRReferText string `json:"asr_refer_text"`
}

func init() {
	RegisterAdapter(NewQQAdapter())
}

func NewQQAdapter() *QQAdapter {
	return &QQAdapter{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		wsDialer: &websocket.Dialer{
			HandshakeTimeout: 15 * time.Second,
		},
		tokenCache:         map[string]qqAccessToken{},
		identityCache:      map[string]qqIdentity{},
		inboundProcessor:   defaultInboundPipeline.PersistEvents,
		reconnectBaseDelay: time.Second,
		reconnectMaxDelay:  30 * time.Second,
	}
}

func (a *QQAdapter) Platform() string {
	return PlatformQQ
}

func (a *QQAdapter) Start(ctx context.Context, conn Connection) error {
	if err := a.TestConnection(ctx, conn); err != nil {
		return err
	}
	return defaultRuntimeManager.Start(ctx, conn, RuntimeHooks{
		Start: func(runtimeCtx context.Context, enqueue RuntimePayloadHandler) error {
			return a.startRuntime(runtimeCtx, conn, enqueue)
		},
		ProcessPayload: func(runtimeCtx context.Context, payload []byte) error {
			return a.processRuntimePayload(runtimeCtx, conn, payload)
		},
		ProcessOutbound:    func(runtimeCtx context.Context) error { return processOutboundOutbox(runtimeCtx, conn) },
		IsExpectedStop:     isExpectedQQRuntimeStop,
		ReconnectBaseDelay: a.reconnectBaseDelay,
		ReconnectMaxDelay:  a.reconnectMaxDelay,
	})
}

func (a *QQAdapter) Stop(ctx context.Context, connectionID string) error {
	a.mu.Lock()
	delete(a.tokenCache, connectionID)
	delete(a.identityCache, connectionID)
	a.mu.Unlock()
	return defaultRuntimeManager.Stop(ctx, connectionID)
}

func (a *QQAdapter) ValidateConfig(config map[string]any, secrets map[string]any) error {
	if strings.TrimSpace(str(config["appId"])) == "" || strings.TrimSpace(str(secrets["appSecret"])) == "" {
		return ErrInvalidConnectionConfig
	}
	mode := firstNonEmpty(strings.TrimSpace(str(config["connectionMode"])), strings.TrimSpace(str(config["mode"])))
	if mode == "" {
		mode = QQConnectionMode
	}
	if mode != QQConnectionMode {
		return ErrQQUnsupportedConnectionMode
	}
	return nil
}

func (a *QQAdapter) TestConnection(ctx context.Context, conn Connection) error {
	if err := a.ValidateConfig(qqConfig(conn), firstNonEmptySecretMap(conn.Secrets, conn.SecretsMasked)); err != nil {
		return err
	}
	if _, err := a.getAccessToken(ctx, conn); err != nil {
		return err
	}
	gatewayURL, err := a.getGatewayURL(ctx, conn)
	if err != nil {
		return err
	}
	if strings.TrimSpace(gatewayURL) == "" {
		return ErrQQGatewayURLMissing
	}
	a.mu.Lock()
	a.identityCache[conn.ID] = qqIdentity{AppID: strings.TrimSpace(str(qqConfig(conn)["appId"]))}
	a.mu.Unlock()
	return nil
}

func (a *QQAdapter) ParseInbound(ctx context.Context, raw any) ([]InboundEvent, error) {
	return parseQQInbound(raw)
}

func (a *QQAdapter) SendOutbound(ctx context.Context, conn Connection, msg OutboundEnvelope) error {
	if err := a.ValidateConfig(qqConfig(conn), firstNonEmptySecretMap(conn.Secrets, conn.SecretsMasked)); err != nil {
		return err
	}
	path, body, err := buildQQOutboundRequest(msg)
	if err != nil {
		return err
	}
	_, err = a.doAPIRequest(ctx, conn, http.MethodPost, path, body)
	return err
}

func (a *QQAdapter) startRuntime(ctx context.Context, conn Connection, enqueue RuntimePayloadHandler) error {
	gatewayURL, err := a.getGatewayURL(ctx, conn)
	if err != nil {
		return err
	}

	wsConn, _, err := a.wsDialer.DialContext(ctx, gatewayURL, nil)
	if err != nil {
		return err
	}
	defer wsConn.Close()

	helloPayload, err := readQQHelloPayload(wsConn)
	if err != nil {
		return err
	}

	token, err := a.getAccessToken(ctx, conn)
	if err != nil {
		return err
	}
	if err := writeQQGatewayPayload(wsConn, qqGatewayPayload{
		Op: 2,
		D: mustMarshalRawMessage(map[string]any{
			"token":   "QQBot " + token,
			"intents": defaultQQIntents,
			"shard":   []int{0, 1},
			"properties": map[string]any{
				"$os":      "windows",
				"$browser": "dotblue",
				"$device":  "dotblue",
			},
		}),
	}); err != nil {
		return err
	}

	heartbeatInterval := time.Duration(helloPayload.HeartbeatInterval) * time.Millisecond
	if heartbeatInterval <= 0 {
		heartbeatInterval = 45 * time.Second
	}

	var seqMu sync.RWMutex
	var lastSeq *int64

	writeHeartbeat := func() error {
		seqMu.RLock()
		seq := lastSeq
		seqMu.RUnlock()
		return writeQQGatewayPayload(wsConn, qqGatewayPayload{
			Op: 1,
			D:  mustMarshalRawMessage(seq),
		})
	}

	heartbeatTicker := time.NewTicker(heartbeatInterval)
	defer heartbeatTicker.Stop()

	errCh := make(chan error, 1)
	sendErr := func(err error) {
		if err == nil {
			return
		}
		select {
		case errCh <- err:
		default:
		}
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-heartbeatTicker.C:
				if err := writeHeartbeat(); err != nil {
					sendErr(err)
					return
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			return err
		default:
		}

		_, message, err := wsConn.ReadMessage()
		if err != nil {
			return err
		}
		payload, err := parseQQGatewayPayload(message)
		if err != nil {
			return err
		}
		if payload.S != nil {
			seqMu.Lock()
			value := *payload.S
			lastSeq = &value
			seqMu.Unlock()
		}
		switch payload.Op {
		case 0:
			if payload.T == "READY" {
				continue
			}
			events, err := parseQQInbound(message)
			if err != nil {
				return err
			}
			if len(events) == 0 {
				continue
			}
			if err := enqueue(ctx, message); err != nil {
				return err
			}
		case 1:
			if err := writeHeartbeat(); err != nil {
				return err
			}
		case 7:
			return errQQReconnectRequested
		case 9:
			return errQQInvalidSession
		case 10, 11:
			continue
		}
	}
}

func (a *QQAdapter) processRuntimePayload(ctx context.Context, conn Connection, payload []byte) error {
	events, err := parseQQInbound(payload)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return nil
	}
	processor := a.inboundProcessor
	if processor == nil {
		return fmt.Errorf("qq inbound processor is not configured")
	}
	_, err = processor(ctx, conn, events)
	return err
}

func qqConfig(conn Connection) map[string]any {
	config := map[string]any{}
	for key, value := range conn.Config {
		config[key] = value
	}
	if strings.TrimSpace(str(config["connectionMode"])) == "" && strings.TrimSpace(conn.ConnectionMode) != "" {
		config["connectionMode"] = conn.ConnectionMode
	}
	return config
}

func qqAPIBaseURL(config map[string]any) string {
	if custom := strings.TrimSpace(str(config["apiBaseURL"])); custom != "" {
		return strings.TrimRight(custom, "/")
	}
	return defaultQQAPIBaseURL
}

func qqTokenURL(config map[string]any) string {
	if custom := strings.TrimSpace(str(config["tokenURL"])); custom != "" {
		return custom
	}
	return defaultQQTokenURL
}

func (a *QQAdapter) getAccessToken(ctx context.Context, conn Connection) (string, error) {
	a.mu.RLock()
	cached, ok := a.tokenCache[conn.ID]
	a.mu.RUnlock()
	if ok && cached.Value != "" && time.Until(cached.Expiry) > time.Minute {
		return cached.Value, nil
	}

	secrets := firstNonEmptySecretMap(conn.Secrets, conn.SecretsMasked)
	requestBody := map[string]any{
		"appId":        strings.TrimSpace(str(qqConfig(conn)["appId"])),
		"clientSecret": strings.TrimSpace(str(secrets["appSecret"])),
	}
	payload, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, qqTokenURL(qqConfig(conn)), bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("qq token http error: %s", strings.TrimSpace(string(body)))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		Code        int64  `json:"code"`
		Message     string `json:"message"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}
	if tokenResp.Code != 0 {
		return "", fmt.Errorf("qq token error: %s", tokenResp.Message)
	}
	if strings.TrimSpace(tokenResp.AccessToken) == "" {
		return "", fmt.Errorf("qq access token is empty")
	}

	expiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	if tokenResp.ExpiresIn <= 0 {
		expiry = time.Now().Add(2 * time.Hour)
	}
	a.mu.Lock()
	a.tokenCache[conn.ID] = qqAccessToken{
		Value:  strings.TrimSpace(tokenResp.AccessToken),
		Expiry: expiry,
	}
	a.mu.Unlock()
	return strings.TrimSpace(tokenResp.AccessToken), nil
}

func (a *QQAdapter) getGatewayURL(ctx context.Context, conn Connection) (string, error) {
	body, err := a.doAPIRequest(ctx, conn, http.MethodGet, "/gateway/bot", nil)
	if err != nil {
		return "", err
	}
	var response qqGatewayResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}
	return strings.TrimSpace(response.URL), nil
}

func (a *QQAdapter) doAPIRequest(ctx context.Context, conn Connection, method, path string, body any) ([]byte, error) {
	token, err := a.getAccessToken(ctx, conn)
	if err != nil {
		return nil, err
	}
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, qqAPIBaseURL(qqConfig(conn))+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "QQBot "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("qq http error: %s", strings.TrimSpace(string(responseBody)))
	}
	return responseBody, nil
}

func readQQHelloPayload(conn *websocket.Conn) (qqHelloPayload, error) {
	_, message, err := conn.ReadMessage()
	if err != nil {
		return qqHelloPayload{}, err
	}
	payload, err := parseQQGatewayPayload(message)
	if err != nil {
		return qqHelloPayload{}, err
	}
	if payload.Op != 10 {
		return qqHelloPayload{}, fmt.Errorf("qq gateway hello opcode mismatch")
	}
	var hello qqHelloPayload
	if err := json.Unmarshal(payload.D, &hello); err != nil {
		return qqHelloPayload{}, err
	}
	return hello, nil
}

func writeQQGatewayPayload(conn *websocket.Conn, payload qqGatewayPayload) error {
	return conn.WriteJSON(payload)
}

func parseQQGatewayPayload(payload []byte) (qqGatewayPayload, error) {
	payload = bytesTrimSpace(payload)
	if len(payload) == 0 {
		return qqGatewayPayload{}, ErrQQInvalidPayload
	}
	var parsed qqGatewayPayload
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return qqGatewayPayload{}, err
	}
	return parsed, nil
}

func parseQQInbound(raw any) ([]InboundEvent, error) {
	payload, rawPayload, err := coerceQQPayload(raw)
	if err != nil {
		return nil, err
	}
	if payload.Op != 0 {
		return nil, nil
	}
	switch strings.TrimSpace(payload.T) {
	case "C2C_MESSAGE_CREATE":
		event, ok := buildQQInboundEvent(payload.T, payload.D, rawPayload)
		if !ok {
			return nil, nil
		}
		return []InboundEvent{event}, nil
	case "GROUP_AT_MESSAGE_CREATE":
		event, ok := buildQQInboundEvent(payload.T, payload.D, rawPayload)
		if !ok {
			return nil, nil
		}
		return []InboundEvent{event}, nil
	default:
		return nil, nil
	}
}

func coerceQQPayload(raw any) (qqGatewayPayload, []byte, error) {
	switch value := raw.(type) {
	case qqGatewayPayload:
		payload, err := json.Marshal(value)
		if err != nil {
			payload = nil
		}
		return value, payload, nil
	case *qqGatewayPayload:
		if value == nil {
			return qqGatewayPayload{}, nil, ErrQQInvalidPayload
		}
		return coerceQQPayload(*value)
	case []byte:
		payload, err := parseQQGatewayPayload(value)
		return payload, append([]byte(nil), value...), err
	case json.RawMessage:
		payload, err := parseQQGatewayPayload(value)
		return payload, append([]byte(nil), value...), err
	case string:
		return coerceQQPayload([]byte(value))
	default:
		payload, err := json.Marshal(value)
		if err != nil {
			return qqGatewayPayload{}, nil, ErrQQInvalidPayload
		}
		parsed, err := parseQQGatewayPayload(payload)
		return parsed, payload, err
	}
}

func buildQQInboundEvent(eventType string, rawData json.RawMessage, rawPayload []byte) (InboundEvent, bool) {
	var event qqMessageEvent
	if err := json.Unmarshal(rawData, &event); err != nil {
		return InboundEvent{}, false
	}
	if strings.TrimSpace(event.ID) == "" {
		return InboundEvent{}, false
	}

	text := strings.TrimSpace(event.Content)
	attachments := buildQQInboundAttachments(event.Attachments)
	if text == "" && len(attachments) == 0 {
		return InboundEvent{}, false
	}

	segments := []RichSegment{}
	if text != "" {
		segments = append(segments, RichSegment{Type: "text", Text: text})
	}
	for _, attachment := range attachments {
		segments = append(segments, RichSegment{
			Type: attachment.Type,
			Meta: map[string]any{
				"media_ref": attachment.MediaRef,
				"name":      attachment.Name,
			},
		})
	}

	replyHandle := map[string]any{
		"message_id": event.ID,
	}
	chatType := "p2p"
	externalChatID := strings.TrimSpace(event.Author.UserOpenID)
	externalUserID := strings.TrimSpace(event.Author.UserOpenID)
	mentionsBot := false
	peerType := "c2c"
	if eventType == "GROUP_AT_MESSAGE_CREATE" {
		chatType = "group"
		mentionsBot = true
		externalChatID = strings.TrimSpace(event.GroupOpenID)
		externalUserID = strings.TrimSpace(event.Author.MemberOpenID)
		replyHandle["group_openid"] = externalChatID
		peerType = "group"
	} else {
		replyHandle["user_openid"] = externalChatID
	}
	replyHandle["peer_type"] = peerType

	return InboundEvent{
		Platform:       PlatformQQ,
		EventID:        strings.TrimSpace(event.ID),
		MessageID:      strings.TrimSpace(event.ID),
		ExternalChatID: externalChatID,
		ExternalUserID: externalUserID,
		ChatType:       chatType,
		MentionsBot:    mentionsBot,
		Text:           text,
		Segments:       segments,
		Attachments:    attachments,
		ReplyHandle:    replyHandle,
		RawPayload:     append([]byte(nil), rawPayload...),
		ReceivedAt:     parseQQTimestamp(event.Timestamp),
	}, true
}

func buildQQInboundAttachments(src []qqMessageAttachment) []InboundAttachment {
	if len(src) == 0 {
		return nil
	}
	attachments := make([]InboundAttachment, 0, len(src))
	for idx, attachment := range src {
		refID := firstNonEmpty(strings.TrimSpace(attachment.URL), fmt.Sprintf("attachment_%d", idx))
		attachmentType := detectQQAttachmentType(attachment.ContentType)
		item := buildInboundAttachment(PlatformQQ, attachmentType, refID, attachment.FileName)
		item.URL = strings.TrimSpace(firstNonEmpty(attachment.URL, attachment.VoiceWavURL))
		item.ContentType = strings.TrimSpace(attachment.ContentType)
		if item.URL != "" {
			attachments = append(attachments, item)
		}
	}
	return attachments
}

func detectQQAttachmentType(contentType string) string {
	contentType = strings.TrimSpace(strings.ToLower(contentType))
	switch {
	case contentType == "voice":
		return "audio"
	case strings.HasPrefix(contentType, "image/"):
		return "image"
	case strings.HasPrefix(contentType, "video/"):
		return "video"
	default:
		return "file"
	}
}

func parseQQTimestamp(raw string) time.Time {
	if parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(raw)); err == nil {
		return parsed
	}
	return time.Now()
}

func buildQQOutboundRequest(msg OutboundEnvelope) (string, map[string]any, error) {
	peerType := strings.TrimSpace(str(msg.ReplyHandle["peer_type"]))
	messageID := strings.TrimSpace(str(msg.ReplyHandle["message_id"]))
	text := formatQQOutboundText(msg)
	if text == "" {
		text = "..."
	}

	body := map[string]any{
		"content":  text,
		"msg_type": 0,
	}
	if messageID != "" {
		body["msg_id"] = messageID
		body["msg_seq"] = 1
	}

	switch peerType {
	case "group":
		groupOpenID := firstNonEmpty(strings.TrimSpace(str(msg.ReplyHandle["group_openid"])), strings.TrimSpace(msg.ExternalChatID))
		if groupOpenID == "" {
			return "", nil, fmt.Errorf("qq outbound group openid is empty")
		}
		return "/v2/groups/" + url.PathEscape(groupOpenID) + "/messages", body, nil
	default:
		userOpenID := firstNonEmpty(strings.TrimSpace(str(msg.ReplyHandle["user_openid"])), strings.TrimSpace(msg.ExternalChatID))
		if userOpenID == "" {
			return "", nil, fmt.Errorf("qq outbound user openid is empty")
		}
		return "/v2/users/" + url.PathEscape(userOpenID) + "/messages", body, nil
	}
}

func formatQQOutboundText(msg OutboundEnvelope) string {
	text := strings.TrimSpace(msg.Text)
	if text == "" {
		text = strings.TrimSpace(segmentsPlainText(msg.Segments))
	}
	if len(msg.Attachments) == 0 {
		return text
	}

	lines := make([]string, 0, len(msg.Attachments))
	for _, attachment := range msg.Attachments {
		label := strings.TrimSpace(attachment.Name)
		if label == "" {
			label = strings.TrimSpace(attachment.Type)
		}
		if label == "" {
			label = "attachment"
		}
		lines = append(lines, "["+label+"]")
	}
	fallback := strings.Join(lines, " ")
	if text == "" {
		return fallback
	}
	return text + "\n\n" + fallback
}

func mustMarshalRawMessage(v any) json.RawMessage {
	payload, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return payload
}

func isExpectedQQRuntimeStop(err error) bool {
	if err == nil {
		return true
	}
	if errors.Is(err, context.Canceled) {
		return true
	}
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	return errors.Is(err, errQQReconnectRequested) ||
		errors.Is(err, errQQInvalidSession) ||
		strings.Contains(msg, "context canceled") ||
		strings.Contains(msg, "context cancelled") ||
		strings.Contains(msg, "operation was canceled")
}
