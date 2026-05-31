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
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	PlatformMatrix           = "matrix"
	MatrixConnectionModeSync = "sync"
	defaultMatrixSyncTimeout = 30
)

var (
	ErrMatrixInvalidPayload            = errors.New("invalid matrix payload")
	ErrMatrixUnsupportedConnectionMode = errors.New("matrix only supports sync mode in v1")
)

type MatrixAdapter struct {
	mu                 sync.RWMutex
	httpClient         *http.Client
	identityCache      map[string]matrixIdentity
	inboundProcessor   feishuInboundProcessor
	reconnectBaseDelay time.Duration
	reconnectMaxDelay  time.Duration
}

type matrixIdentity struct {
	UserID string
}

type matrixSyncResponse struct {
	NextBatch string `json:"next_batch"`
	Rooms     struct {
		Join   map[string]matrixJoinedRoom `json:"join"`
		Invite map[string]matrixInviteRoom `json:"invite"`
	} `json:"rooms"`
}

type matrixJoinedRoom struct {
	Timeline struct {
		Events []map[string]any `json:"events"`
	} `json:"timeline"`
}

type matrixInviteRoom struct {
	InviteState struct {
		Events []map[string]any `json:"events"`
	} `json:"invite_state"`
}

func init() {
	RegisterAdapter(NewMatrixAdapter())
}

func NewMatrixAdapter() *MatrixAdapter {
	return &MatrixAdapter{
		httpClient: &http.Client{
			Timeout: 70 * time.Second,
		},
		identityCache:      map[string]matrixIdentity{},
		inboundProcessor:   defaultInboundPipeline.PersistEvents,
		reconnectBaseDelay: time.Second,
		reconnectMaxDelay:  30 * time.Second,
	}
}

func (a *MatrixAdapter) Platform() string {
	return PlatformMatrix
}

func (a *MatrixAdapter) Start(ctx context.Context, conn Connection) error {
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
		IsExpectedStop:     isExpectedMatrixRuntimeStop,
		ReconnectBaseDelay: a.reconnectBaseDelay,
		ReconnectMaxDelay:  a.reconnectMaxDelay,
	})
}

func (a *MatrixAdapter) Stop(ctx context.Context, connectionID string) error {
	a.mu.Lock()
	delete(a.identityCache, connectionID)
	a.mu.Unlock()
	return defaultRuntimeManager.Stop(ctx, connectionID)
}

func (a *MatrixAdapter) ValidateConfig(config map[string]any, secrets map[string]any) error {
	if strings.TrimSpace(str(config["homeserver"])) == "" || strings.TrimSpace(str(config["userId"])) == "" {
		return ErrInvalidConnectionConfig
	}
	if strings.TrimSpace(str(secrets["accessToken"])) == "" {
		return ErrInvalidConnectionConfig
	}
	mode := firstNonEmpty(strings.TrimSpace(str(config["connectionMode"])), strings.TrimSpace(str(config["mode"])))
	if mode != "" && mode != MatrixConnectionModeSync {
		return ErrMatrixUnsupportedConnectionMode
	}
	return nil
}

func (a *MatrixAdapter) TestConnection(ctx context.Context, conn Connection) error {
	if err := a.ValidateConfig(matrixConfig(conn), firstNonEmptySecretMap(conn.Secrets, conn.SecretsMasked)); err != nil {
		return err
	}
	identity, err := a.getIdentity(ctx, conn)
	if err != nil {
		return err
	}
	expectedUserID := strings.TrimSpace(str(matrixConfig(conn)["userId"]))
	if expectedUserID != "" && identity.UserID != "" && identity.UserID != expectedUserID {
		return fmt.Errorf("matrix whoami user mismatch")
	}
	return nil
}

func (a *MatrixAdapter) ParseInbound(ctx context.Context, raw any) ([]InboundEvent, error) {
	return parseMatrixInbound(raw, "")
}

func (a *MatrixAdapter) SendOutbound(ctx context.Context, conn Connection, msg OutboundEnvelope) error {
	if err := a.ValidateConfig(matrixConfig(conn), firstNonEmptySecretMap(conn.Secrets, conn.SecretsMasked)); err != nil {
		return err
	}

	roomID := firstNonEmpty(str(msg.ReplyHandle["room_id"]), msg.ExternalChatID)
	if strings.TrimSpace(roomID) == "" {
		return fmt.Errorf("matrix outbound room id is empty")
	}

	content := map[string]any{
		"msgtype": "m.text",
		"body":    formatMatrixOutboundText(msg),
	}
	if eventID := strings.TrimSpace(str(msg.ReplyHandle["event_id"])); eventID != "" {
		content["m.relates_to"] = map[string]any{
			"m.in_reply_to": map[string]any{
				"event_id": eventID,
			},
		}
	}

	txnID := strconv.FormatInt(time.Now().UnixNano(), 10)
	_, err := a.doMatrixRequest(
		ctx,
		conn,
		http.MethodPut,
		"/_matrix/client/v3/rooms/"+url.PathEscape(roomID)+"/send/m.room.message/"+txnID,
		content,
	)
	return err
}

func (a *MatrixAdapter) startRuntime(ctx context.Context, conn Connection, enqueue RuntimePayloadHandler) error {
	if _, err := a.getIdentity(ctx, conn); err != nil {
		return err
	}

	since := ""
	timeoutSeconds := matrixSyncTimeoutSeconds(conn)
	joinOnInvite := matrixJoinOnInvite(conn)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		response, rawPayload, err := a.sync(ctx, conn, since, timeoutSeconds)
		if err != nil {
			return err
		}
		if joinOnInvite {
			if err := a.joinInvitedRooms(ctx, conn, response.Rooms.Invite); err != nil {
				return err
			}
		}
		if len(rawPayload) > 0 {
			if err := enqueue(ctx, rawPayload); err != nil {
				return err
			}
		}
		if response.NextBatch != "" {
			since = response.NextBatch
		}
	}
}

func (a *MatrixAdapter) processRuntimePayload(ctx context.Context, conn Connection, payload []byte) error {
	identity, err := a.getIdentity(ctx, conn)
	if err != nil {
		return err
	}
	events, err := parseMatrixInbound(payload, identity.UserID)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return nil
	}
	processor := a.inboundProcessor
	if processor == nil {
		return fmt.Errorf("matrix inbound processor is not configured")
	}
	_, err = processor(ctx, conn, events)
	return err
}

func matrixConfig(conn Connection) map[string]any {
	config := map[string]any{}
	for key, value := range conn.Config {
		config[key] = value
	}
	if strings.TrimSpace(str(config["connectionMode"])) == "" && strings.TrimSpace(conn.ConnectionMode) != "" {
		config["connectionMode"] = conn.ConnectionMode
	}
	return config
}

func matrixSyncTimeoutSeconds(conn Connection) int {
	timeout := defaultMatrixSyncTimeout
	if raw := str(matrixConfig(conn)["syncTimeoutSeconds"]); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 60 {
			timeout = parsed
		}
	}
	return timeout
}

func matrixJoinOnInvite(conn Connection) bool {
	raw, ok := matrixConfig(conn)["joinOnInvite"]
	if !ok {
		return true
	}
	switch value := raw.(type) {
	case bool:
		return value
	case string:
		switch strings.TrimSpace(strings.ToLower(value)) {
		case "false", "0", "no":
			return false
		default:
			return true
		}
	default:
		return true
	}
}

func (a *MatrixAdapter) getIdentity(ctx context.Context, conn Connection) (matrixIdentity, error) {
	a.mu.RLock()
	cached, ok := a.identityCache[conn.ID]
	a.mu.RUnlock()
	if ok && cached.UserID != "" {
		return cached, nil
	}

	responseBody, err := a.doMatrixRequest(ctx, conn, http.MethodGet, "/_matrix/client/v3/account/whoami", nil)
	if err != nil {
		return matrixIdentity{}, err
	}
	var payload struct {
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(responseBody, &payload); err != nil {
		return matrixIdentity{}, err
	}
	identity := matrixIdentity{UserID: strings.TrimSpace(payload.UserID)}
	if identity.UserID == "" {
		return matrixIdentity{}, fmt.Errorf("matrix whoami user id is empty")
	}
	a.mu.Lock()
	a.identityCache[conn.ID] = identity
	a.mu.Unlock()
	return identity, nil
}

func (a *MatrixAdapter) sync(ctx context.Context, conn Connection, since string, timeoutSeconds int) (matrixSyncResponse, []byte, error) {
	query := url.Values{}
	if since != "" {
		query.Set("since", since)
	}
	query.Set("timeout", strconv.Itoa(timeoutSeconds*1000))
	query.Set("filter", `{"room":{"timeline":{"limit":20},"state":{"lazy_load_members":true}}}`)
	path := "/_matrix/client/v3/sync?" + query.Encode()
	responseBody, err := a.doMatrixRequest(ctx, conn, http.MethodGet, path, nil)
	if err != nil {
		return matrixSyncResponse{}, nil, err
	}

	var payload matrixSyncResponse
	if err := json.Unmarshal(responseBody, &payload); err != nil {
		return matrixSyncResponse{}, nil, err
	}
	return payload, responseBody, nil
}

func (a *MatrixAdapter) joinInvitedRooms(ctx context.Context, conn Connection, rooms map[string]matrixInviteRoom) error {
	for roomID := range rooms {
		if strings.TrimSpace(roomID) == "" {
			continue
		}
		if _, err := a.doMatrixRequest(
			ctx,
			conn,
			http.MethodPost,
			"/_matrix/client/v3/rooms/"+url.PathEscape(roomID)+"/join",
			map[string]any{},
		); err != nil {
			return err
		}
	}
	return nil
}

func (a *MatrixAdapter) doMatrixRequest(
	ctx context.Context,
	conn Connection,
	method string,
	path string,
	body any,
) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, matrixBaseURL(matrixConfig(conn))+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(str(firstNonEmptySecretMap(conn.Secrets, conn.SecretsMasked)["accessToken"])))
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
		return nil, fmt.Errorf("matrix http error: %s", strings.TrimSpace(string(responseBody)))
	}
	return responseBody, nil
}

func matrixBaseURL(config map[string]any) string {
	baseURL := strings.TrimSpace(str(config["homeserver"]))
	return strings.TrimRight(baseURL, "/")
}

func parseMatrixInbound(raw any, botUserID string) ([]InboundEvent, error) {
	response, err := coerceMatrixSyncResponse(raw)
	if err != nil {
		return nil, err
	}
	return buildMatrixInboundEvents(response, botUserID), nil
}

func coerceMatrixSyncResponse(raw any) (matrixSyncResponse, error) {
	switch value := raw.(type) {
	case matrixSyncResponse:
		return value, nil
	case *matrixSyncResponse:
		if value == nil {
			return matrixSyncResponse{}, ErrMatrixInvalidPayload
		}
		return *value, nil
	case []byte:
		return parseMatrixSyncPayload(value)
	case json.RawMessage:
		return parseMatrixSyncPayload(value)
	case string:
		return parseMatrixSyncPayload([]byte(value))
	default:
		payload, err := json.Marshal(value)
		if err != nil {
			return matrixSyncResponse{}, ErrMatrixInvalidPayload
		}
		return parseMatrixSyncPayload(payload)
	}
}

func parseMatrixSyncPayload(payload []byte) (matrixSyncResponse, error) {
	payload = bytesTrimSpace(payload)
	if len(payload) == 0 {
		return matrixSyncResponse{}, ErrMatrixInvalidPayload
	}
	var response matrixSyncResponse
	if err := json.Unmarshal(payload, &response); err != nil {
		return matrixSyncResponse{}, err
	}
	return response, nil
}

func buildMatrixInboundEvents(response matrixSyncResponse, botUserID string) []InboundEvent {
	var events []InboundEvent
	for roomID, room := range response.Rooms.Join {
		for _, event := range room.Timeline.Events {
			inbound, ok := buildMatrixInboundEvent(roomID, event, botUserID)
			if !ok {
				continue
			}
			events = append(events, inbound)
		}
	}
	return events
}

func buildMatrixInboundEvent(roomID string, event map[string]any, botUserID string) (InboundEvent, bool) {
	if strings.TrimSpace(str(event["type"])) != "m.room.message" {
		return InboundEvent{}, false
	}
	sender := strings.TrimSpace(str(event["sender"]))
	if sender == "" || sender == botUserID {
		return InboundEvent{}, false
	}
	content, _ := event["content"].(map[string]any)
	if content == nil {
		return InboundEvent{}, false
	}
	msgType := strings.TrimSpace(str(content["msgtype"]))
	attachmentType := matrixAttachmentType(msgType)
	text := strings.TrimSpace(str(content["body"]))
	segments := []RichSegment{}
	attachments := []InboundAttachment{}

	if attachmentType == "text" {
		if text == "" {
			return InboundEvent{}, false
		}
		segments = append(segments, RichSegment{Type: "text", Text: text})
	} else {
		attachment := buildInboundAttachment(PlatformMatrix, attachmentType, strings.TrimSpace(str(event["event_id"])), text)
		attachment.URL = strings.TrimSpace(str(content["url"]))
		if info, ok := content["info"].(map[string]any); ok {
			attachment.ContentType = strings.TrimSpace(str(info["mimetype"]))
		}
		if attachment.URL != "" {
			attachments = append(attachments, attachment)
		}
		if text != "" {
			segments = append(segments, RichSegment{Type: "text", Text: text})
		}
	}

	mentionsBot := false
	if mentions, ok := content["m.mentions"].(map[string]any); ok {
		if userIDs, ok := mentions["user_ids"].([]any); ok {
			for _, userID := range userIDs {
				if strings.TrimSpace(str(userID)) == botUserID && botUserID != "" {
					mentionsBot = true
					break
				}
			}
		}
	}

	replyHandle := map[string]any{
		"room_id":  roomID,
		"event_id": strings.TrimSpace(str(event["event_id"])),
	}
	return InboundEvent{
		Platform:       PlatformMatrix,
		EventID:        strings.TrimSpace(str(event["event_id"])),
		MessageID:      strings.TrimSpace(str(event["event_id"])),
		ExternalChatID: roomID,
		ExternalUserID: sender,
		ChatType:       "group",
		MentionsBot:    mentionsBot,
		Text:           text,
		Segments:       segments,
		Attachments:    attachments,
		ReplyHandle:    replyHandle,
		RawPayload:     mustMarshalMatrixEvent(event),
		ReceivedAt:     matrixEventTime(event["origin_server_ts"]),
	}, true
}

func matrixAttachmentType(msgType string) string {
	switch strings.TrimSpace(msgType) {
	case "m.image":
		return "image"
	case "m.audio":
		return "audio"
	case "m.video":
		return "video"
	case "m.file":
		return "file"
	default:
		return "text"
	}
}

func mustMarshalMatrixEvent(event map[string]any) []byte {
	payload, err := json.Marshal(event)
	if err != nil {
		return nil
	}
	return payload
}

func matrixEventTime(raw any) time.Time {
	switch value := raw.(type) {
	case float64:
		return time.UnixMilli(int64(value))
	case int64:
		return time.UnixMilli(value)
	case int:
		return time.UnixMilli(int64(value))
	case json.Number:
		n, _ := value.Int64()
		return time.UnixMilli(n)
	case string:
		n, _ := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if n > 0 {
			return time.UnixMilli(n)
		}
	}
	return time.Now()
}

func formatMatrixOutboundText(msg OutboundEnvelope) string {
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

func isExpectedMatrixRuntimeStop(err error) bool {
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
