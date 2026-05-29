package im

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	PlatformTelegram              = "telegram"
	TelegramConnectionModePolling = "polling"
)

var (
	ErrTelegramUnsupportedConnectionMode = errors.New("telegram only supports polling mode in v1")
	ErrTelegramInvalidPayload            = errors.New("invalid telegram payload")
	ErrTelegramUnsupportedAttachment     = errors.New("unsupported telegram attachment type")
)

type TelegramAdapter struct {
	mu                 sync.RWMutex
	httpClient         *http.Client
	botCache           map[string]telegramBotIdentity
	inboundProcessor   feishuInboundProcessor
	reconnectBaseDelay time.Duration
	reconnectMaxDelay  time.Duration
}

type telegramBotIdentity struct {
	ID       int64
	Username string
}

type telegramPollUpdate struct {
	UpdateID      int64          `json:"update_id"`
	Message       map[string]any `json:"message"`
	EditedMessage map[string]any `json:"edited_message"`
	Raw           map[string]any `json:"-"`
}

type telegramOutboundRequest struct {
	Method string
	Body   map[string]any
}

func init() {
	RegisterAdapter(NewTelegramAdapter())
}

func NewTelegramAdapter() *TelegramAdapter {
	return &TelegramAdapter{
		httpClient: &http.Client{
			Timeout: 70 * time.Second,
		},
		botCache:           map[string]telegramBotIdentity{},
		inboundProcessor:   defaultInboundPipeline.PersistEvents,
		reconnectBaseDelay: time.Second,
		reconnectMaxDelay:  30 * time.Second,
	}
}

func (a *TelegramAdapter) Platform() string {
	return PlatformTelegram
}

func (a *TelegramAdapter) Start(ctx context.Context, conn Connection) error {
	if err := a.TestConnection(ctx, conn); err != nil {
		return err
	}
	return defaultRuntimeManager.Start(ctx, conn, RuntimeHooks{
		Start: func(runtimeCtx context.Context, enqueue RuntimePayloadHandler) error {
			return a.startRuntime(runtimeCtx, conn, enqueue)
		},
		ProcessPayload:     a.processRuntimePayload(conn),
		ProcessOutbound:    func(runtimeCtx context.Context) error { return processOutboundOutbox(runtimeCtx, conn) },
		IsExpectedStop:     isExpectedTelegramRuntimeStop,
		ReconnectBaseDelay: a.reconnectBaseDelay,
		ReconnectMaxDelay:  a.reconnectMaxDelay,
	})
}

func (a *TelegramAdapter) Stop(ctx context.Context, connectionID string) error {
	a.mu.Lock()
	delete(a.botCache, connectionID)
	a.mu.Unlock()
	return defaultRuntimeManager.Stop(ctx, connectionID)
}

func (a *TelegramAdapter) ValidateConfig(config map[string]any, secrets map[string]any) error {
	if str(secrets["token"]) == "" {
		return ErrInvalidConnectionConfig
	}
	mode := str(config["connectionMode"])
	if mode != "" && mode != TelegramConnectionModePolling {
		return ErrTelegramUnsupportedConnectionMode
	}
	return nil
}

func (a *TelegramAdapter) TestConnection(ctx context.Context, conn Connection) error {
	if err := a.ValidateConfig(telegramConfig(conn), firstNonEmptySecretMap(conn.Secrets, conn.SecretsMasked)); err != nil {
		return err
	}
	_, err := a.getBotIdentity(ctx, conn)
	return err
}

func (a *TelegramAdapter) ParseInbound(ctx context.Context, raw any) ([]InboundEvent, error) {
	return parseTelegramInbound(raw, "")
}

func (a *TelegramAdapter) SendOutbound(ctx context.Context, conn Connection, msg OutboundEnvelope) error {
	if err := a.ValidateConfig(telegramConfig(conn), firstNonEmptySecretMap(conn.Secrets, conn.SecretsMasked)); err != nil {
		return err
	}

	requests, err := buildTelegramOutboundRequests(msg)
	if err != nil {
		return err
	}
	if len(requests) == 0 {
		return fmt.Errorf("telegram outbound payload is empty")
	}

	for _, request := range requests {
		payload := cloneTelegramOutboundBody(request.Body)
		mergeTelegramSendContext(payload, msg)
		if _, err := a.doTelegramAPIRequest(ctx, conn, request.Method, payload); err != nil {
			return err
		}
	}
	return nil
}

func telegramConfig(conn Connection) map[string]any {
	config := map[string]any{}
	for k, v := range conn.Config {
		config[k] = v
	}
	if strings.TrimSpace(str(config["connectionMode"])) == "" && strings.TrimSpace(conn.ConnectionMode) != "" {
		config["connectionMode"] = conn.ConnectionMode
	}
	return config
}

func isExpectedTelegramRuntimeStop(err error) bool {
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

func (a *TelegramAdapter) processRuntimePayload(conn Connection) RuntimePayloadHandler {
	return func(ctx context.Context, payload []byte) error {
		identity, err := a.getBotIdentity(ctx, conn)
		if err != nil {
			return err
		}
		events, err := parseTelegramInbound(payload, identity.Username)
		if err != nil {
			return err
		}
		if len(events) == 0 {
			return nil
		}
		processor := a.inboundProcessor
		if processor == nil {
			return fmt.Errorf("telegram inbound processor is not configured")
		}
		_, err = processor(ctx, conn, events)
		return err
	}
}

func (a *TelegramAdapter) startRuntime(ctx context.Context, conn Connection, enqueue RuntimePayloadHandler) error {
	identity, err := a.getBotIdentity(ctx, conn)
	if err != nil {
		return err
	}

	// Telegram long polling is stateful, so we keep the last confirmed offset
	// inside the runtime loop and rely on inbound deduplication for safety.
	offset := int64(0)
	timeoutSeconds := telegramPollTimeoutSeconds(conn)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		updates, nextOffset, err := a.getUpdates(ctx, conn, offset, timeoutSeconds)
		if err != nil {
			return err
		}
		for _, update := range updates {
			if update.Raw == nil {
				continue
			}
			payload, err := json.Marshal(update.Raw)
			if err != nil {
				return err
			}
			if _, err := parseTelegramInbound(update.Raw, identity.Username); err != nil {
				return err
			}
			if err := enqueue(ctx, payload); err != nil {
				return err
			}
		}
		if nextOffset > offset {
			offset = nextOffset
		}
	}
}

func telegramPollTimeoutSeconds(conn Connection) int {
	timeout := 30
	if raw := str(telegramConfig(conn)["pollTimeoutSeconds"]); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 60 {
			timeout = parsed
		}
	}
	return timeout
}

func (a *TelegramAdapter) getUpdates(ctx context.Context, conn Connection, offset int64, timeoutSeconds int) ([]telegramPollUpdate, int64, error) {
	body := map[string]any{
		"offset":          offset,
		"limit":           50,
		"timeout":         timeoutSeconds,
		"allowed_updates": []string{"message", "edited_message"},
	}
	respBody, err := a.doTelegramAPIRequest(ctx, conn, "getUpdates", body)
	if err != nil {
		return nil, offset, err
	}

	var payload struct {
		OK          bool             `json:"ok"`
		Description string           `json:"description"`
		Result      []map[string]any `json:"result"`
	}
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return nil, offset, err
	}
	if !payload.OK {
		return nil, offset, fmt.Errorf("telegram api error: %s", firstNonEmpty(payload.Description, string(respBody)))
	}

	updates := make([]telegramPollUpdate, 0, len(payload.Result))
	nextOffset := offset
	for _, raw := range payload.Result {
		update := telegramPollUpdate{
			UpdateID: int64Value(raw["update_id"]),
			Raw:      raw,
		}
		if message, ok := raw["message"].(map[string]any); ok {
			update.Message = message
		}
		if editedMessage, ok := raw["edited_message"].(map[string]any); ok {
			update.EditedMessage = editedMessage
		}
		if update.UpdateID >= nextOffset {
			nextOffset = update.UpdateID + 1
		}
		updates = append(updates, update)
	}
	return updates, nextOffset, nil
}

func parseTelegramInbound(raw any, botUsername string) ([]InboundEvent, error) {
	payload, err := coerceRawMap(raw)
	if err != nil {
		return nil, ErrTelegramInvalidPayload
	}

	message := pickTelegramMessage(payload)
	if message == nil {
		return nil, ErrTelegramInvalidPayload
	}
	chat, _ := message["chat"].(map[string]any)
	from, _ := message["from"].(map[string]any)
	if chat == nil {
		return nil, ErrTelegramInvalidPayload
	}

	text := strings.TrimSpace(firstNonEmpty(str(message["text"]), str(message["caption"])))
	attachments := parseTelegramAttachments(message)
	segments := make([]RichSegment, 0, 1+len(attachments))
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

	messageID := int64String(message["message_id"])
	chatID := int64String(chat["id"])
	threadID := int64String(message["message_thread_id"])
	if messageID == "" || chatID == "" {
		return nil, ErrTelegramInvalidPayload
	}

	eventID := int64String(payload["update_id"])
	if eventID == "" {
		eventID = messageID
	}

	inbound := InboundEvent{
		Platform:         PlatformTelegram,
		EventID:          eventID,
		MessageID:        messageID,
		ExternalChatID:   chatID,
		ExternalThreadID: threadID,
		ExternalUserID:   int64String(from["id"]),
		ChatType:         normalizeTelegramChatType(str(chat["type"])),
		MentionsBot:      detectTelegramMention(text, botUsername),
		Text:             text,
		Segments:         segments,
		Attachments:      attachments,
		ReplyHandle: map[string]any{
			"chat_id":           chatID,
			"message_id":        messageID,
			"message_thread_id": threadID,
		},
		ReceivedAt: time.Now(),
	}
	if rawBytes, err := json.Marshal(payload); err == nil {
		inbound.RawPayload = rawBytes
	}
	return []InboundEvent{inbound}, nil
}

func pickTelegramMessage(payload map[string]any) map[string]any {
	if message, ok := payload["message"].(map[string]any); ok {
		return message
	}
	if editedMessage, ok := payload["edited_message"].(map[string]any); ok {
		return editedMessage
	}
	return nil
}

func detectTelegramMention(text, botUsername string) bool {
	text = strings.ToLower(strings.TrimSpace(text))
	if text == "" {
		return false
	}
	if botUsername == "" {
		return strings.Contains(text, "@")
	}
	return strings.Contains(text, "@"+strings.ToLower(strings.TrimPrefix(botUsername, "@")))
}

func normalizeTelegramChatType(chatType string) string {
	switch strings.TrimSpace(strings.ToLower(chatType)) {
	case "private":
		return "p2p"
	default:
		return "group"
	}
}

func parseTelegramAttachments(message map[string]any) []InboundAttachment {
	var attachments []InboundAttachment

	if photos, ok := message["photo"].([]any); ok && len(photos) > 0 {
		if last, ok := photos[len(photos)-1].(map[string]any); ok {
			if attachment := buildInboundAttachment(PlatformTelegram, "image", str(last["file_id"]), "photo"); attachment.MediaRef != "" {
				attachments = append(attachments, attachment)
			}
		}
	}
	if document, ok := message["document"].(map[string]any); ok {
		if attachment := buildInboundAttachment(PlatformTelegram, "file", str(document["file_id"]), str(document["file_name"])); attachment.MediaRef != "" {
			attachments = append(attachments, attachment)
		}
	}
	if audio, ok := message["audio"].(map[string]any); ok {
		if attachment := buildInboundAttachment(PlatformTelegram, "audio", str(audio["file_id"]), str(audio["file_name"])); attachment.MediaRef != "" {
			attachments = append(attachments, attachment)
		}
	}
	if video, ok := message["video"].(map[string]any); ok {
		if attachment := buildInboundAttachment(PlatformTelegram, "video", str(video["file_id"]), str(video["file_name"])); attachment.MediaRef != "" {
			attachments = append(attachments, attachment)
		}
	}

	return attachments
}

func buildTelegramOutboundRequests(msg OutboundEnvelope) ([]telegramOutboundRequest, error) {
	requests := make([]telegramOutboundRequest, 0, 1+len(msg.Attachments))
	if text := formatTelegramOutboundText(msg); text != "" {
		requests = append(requests, telegramOutboundRequest{
			Method: "sendMessage",
			Body: map[string]any{
				"text": text,
			},
		})
	}

	for _, attachment := range msg.Attachments {
		request, err := buildTelegramAttachmentRequest(attachment)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func formatTelegramOutboundText(msg OutboundEnvelope) string {
	text := strings.TrimSpace(msg.Text)
	if text != "" {
		return text
	}
	return strings.TrimSpace(segmentsPlainText(msg.Segments))
}

func buildTelegramAttachmentRequest(attachment OutboundAttachment) (telegramOutboundRequest, error) {
	refPlatform, refType, remoteID, ok := ParseMediaRef(attachment.MediaRef)
	if !ok {
		return telegramOutboundRequest{}, fmt.Errorf("telegram attachment media ref is invalid")
	}
	if refPlatform != PlatformTelegram {
		return telegramOutboundRequest{}, fmt.Errorf("telegram attachment media ref platform mismatch")
	}

	attachmentType := strings.TrimSpace(attachment.Type)
	if attachmentType == "" {
		attachmentType = refType
	}
	if attachmentType != refType {
		return telegramOutboundRequest{}, fmt.Errorf("telegram attachment type mismatch")
	}

	switch attachmentType {
	case "image":
		return telegramOutboundRequest{
			Method: "sendPhoto",
			Body: map[string]any{
				"photo": remoteID,
			},
		}, nil
	case "file":
		return telegramOutboundRequest{
			Method: "sendDocument",
			Body: map[string]any{
				"document": remoteID,
			},
		}, nil
	case "audio":
		return telegramOutboundRequest{
			Method: "sendAudio",
			Body: map[string]any{
				"audio": remoteID,
			},
		}, nil
	case "video":
		return telegramOutboundRequest{
			Method: "sendVideo",
			Body: map[string]any{
				"video": remoteID,
			},
		}, nil
	default:
		return telegramOutboundRequest{}, ErrTelegramUnsupportedAttachment
	}
}

func cloneTelegramOutboundBody(src map[string]any) map[string]any {
	cloned := make(map[string]any, len(src))
	for key, value := range src {
		cloned[key] = value
	}
	return cloned
}

func mergeTelegramSendContext(body map[string]any, msg OutboundEnvelope) {
	// Keep thread and reply metadata on every outbound request so text and
	// attachments stay in the same Telegram conversation context.
	body["chat_id"] = telegramAPIValue(firstNonEmpty(str(msg.ReplyHandle["chat_id"]), msg.ExternalChatID))

	if threadID := firstNonEmpty(str(msg.ReplyHandle["message_thread_id"]), msg.ExternalThreadID); threadID != "" {
		body["message_thread_id"] = telegramAPIValue(threadID)
	}
	if replyTo := str(msg.ReplyHandle["message_id"]); replyTo != "" {
		body["reply_to_message_id"] = telegramAPIValue(replyTo)
		body["allow_sending_without_reply"] = true
	}
}

func telegramAPIValue(raw string) any {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return parsed
	}
	return raw
}

func (a *TelegramAdapter) getBotIdentity(ctx context.Context, conn Connection) (telegramBotIdentity, error) {
	a.mu.RLock()
	cached, ok := a.botCache[conn.ID]
	a.mu.RUnlock()
	if ok && cached.Username != "" {
		return cached, nil
	}

	respBody, err := a.doTelegramAPIRequest(ctx, conn, "getMe", map[string]any{})
	if err != nil {
		return telegramBotIdentity{}, err
	}

	var payload struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
		Result      struct {
			ID       int64  `json:"id"`
			Username string `json:"username"`
		} `json:"result"`
	}
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return telegramBotIdentity{}, err
	}
	if !payload.OK {
		return telegramBotIdentity{}, fmt.Errorf("telegram api error: %s", firstNonEmpty(payload.Description, string(respBody)))
	}
	if payload.Result.Username == "" {
		return telegramBotIdentity{}, fmt.Errorf("telegram bot username missing")
	}

	identity := telegramBotIdentity{
		ID:       payload.Result.ID,
		Username: payload.Result.Username,
	}
	a.mu.Lock()
	a.botCache[conn.ID] = identity
	a.mu.Unlock()
	return identity, nil
}

func (a *TelegramAdapter) doTelegramAPIRequest(
	ctx context.Context,
	conn Connection,
	method string,
	body map[string]any,
) ([]byte, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, telegramMethodURL(conn, method), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

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
		return nil, fmt.Errorf("telegram http error: %s", strings.TrimSpace(string(respBody)))
	}
	return respBody, nil
}

func telegramAPIBaseURL(config map[string]any) string {
	if custom := strings.TrimSpace(str(config["apiBaseURL"])); custom != "" {
		return strings.TrimRight(custom, "/")
	}
	return "https://api.telegram.org"
}

func telegramMethodURL(conn Connection, method string) string {
	token := str(firstNonEmptySecretMap(conn.Secrets, conn.SecretsMasked)["token"])
	return telegramAPIBaseURL(telegramConfig(conn)) + "/bot" + token + "/" + strings.TrimSpace(method)
}

func int64Value(v any) int64 {
	switch value := v.(type) {
	case int64:
		return value
	case int:
		return int64(value)
	case float64:
		return int64(value)
	case json.Number:
		out, _ := value.Int64()
		return out
	case string:
		out, _ := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		return out
	default:
		return 0
	}
}

func int64String(v any) string {
	value := int64Value(v)
	if value == 0 {
		return ""
	}
	return strconv.FormatInt(value, 10)
}
