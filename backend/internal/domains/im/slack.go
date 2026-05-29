package im

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	goslack "github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

const (
	PlatformSlack          = "slack"
	SlackConnectionMode    = "socket_mode"
	defaultSlackSocketBase = "https://slack.com/api/"
)

var (
	ErrSlackInvalidPayload            = errors.New("invalid slack payload")
	ErrSlackUnsupportedConnectionMode = errors.New("slack only supports socket_mode in v1")
)

type SlackAdapter struct {
	mu                   sync.RWMutex
	identityCache        map[string]slackBotIdentity
	apiClientFactory     func(conn Connection) slackAPIClient
	socketRuntimeFactory func(conn Connection) slackSocketRuntime
	inboundProcessor     feishuInboundProcessor
	reconnectBaseDelay   time.Duration
	reconnectMaxDelay    time.Duration
}

type slackAPIClient interface {
	AuthTestContext(ctx context.Context) (*goslack.AuthTestResponse, error)
	PostMessageContext(ctx context.Context, channelID string, options ...goslack.MsgOption) (string, string, error)
}

type slackSocketRuntime interface {
	Run(ctx context.Context, handle func(context.Context, slackSocketEnvelope) error) error
}

type slackSocketEnvelope struct {
	Payload []byte
}

type slackBotIdentity struct {
	UserID string
	BotID  string
	TeamID string
}

type slackSocketModeRuntime struct {
	client *socketmode.Client
}

func init() {
	RegisterAdapter(NewSlackAdapter())
}

func NewSlackAdapter() *SlackAdapter {
	return &SlackAdapter{
		identityCache:        map[string]slackBotIdentity{},
		apiClientFactory:     newSlackWebAPIClient,
		socketRuntimeFactory: newSlackSocketRuntime,
		inboundProcessor:     defaultInboundPipeline.PersistEvents,
		reconnectBaseDelay:   time.Second,
		reconnectMaxDelay:    30 * time.Second,
	}
}

func (a *SlackAdapter) Platform() string {
	return PlatformSlack
}

func (a *SlackAdapter) Start(ctx context.Context, conn Connection) error {
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
		IsExpectedStop:     isExpectedSlackRuntimeStop,
		ReconnectBaseDelay: a.reconnectBaseDelay,
		ReconnectMaxDelay:  a.reconnectMaxDelay,
	})
}

func (a *SlackAdapter) Stop(ctx context.Context, connectionID string) error {
	a.mu.Lock()
	delete(a.identityCache, connectionID)
	a.mu.Unlock()
	return defaultRuntimeManager.Stop(ctx, connectionID)
}

func (a *SlackAdapter) ValidateConfig(config map[string]any, secrets map[string]any) error {
	botToken := strings.TrimSpace(str(secrets["botToken"]))
	appToken := strings.TrimSpace(str(secrets["appToken"]))
	if botToken == "" || appToken == "" {
		return ErrInvalidConnectionConfig
	}
	if !strings.HasPrefix(botToken, "xoxb-") || !strings.HasPrefix(appToken, "xapp-") {
		return ErrInvalidConnectionConfig
	}
	mode := firstNonEmpty(strings.TrimSpace(str(config["connectionMode"])), strings.TrimSpace(str(config["mode"])))
	if mode == "" {
		mode = SlackConnectionMode
	}
	if mode != SlackConnectionMode {
		return ErrSlackUnsupportedConnectionMode
	}
	return nil
}

func (a *SlackAdapter) TestConnection(ctx context.Context, conn Connection) error {
	secrets := firstNonEmptySecretMap(conn.Secrets, conn.SecretsMasked)
	if err := a.ValidateConfig(slackConfig(conn), secrets); err != nil {
		return err
	}
	identity, err := a.fetchBotIdentity(ctx, conn)
	if err != nil {
		return err
	}
	if strings.TrimSpace(identity.UserID) == "" {
		return fmt.Errorf("slack auth test returned empty user id")
	}
	return nil
}

func (a *SlackAdapter) ParseInbound(ctx context.Context, raw any) ([]InboundEvent, error) {
	return parseSlackInbound(raw, "")
}

func (a *SlackAdapter) SendOutbound(ctx context.Context, conn Connection, msg OutboundEnvelope) error {
	secrets := firstNonEmptySecretMap(conn.Secrets, conn.SecretsMasked)
	if err := a.ValidateConfig(slackConfig(conn), secrets); err != nil {
		return err
	}

	channelID := firstNonEmpty(strings.TrimSpace(str(msg.ReplyHandle["channel"])), strings.TrimSpace(msg.ExternalChatID))
	if channelID == "" {
		return fmt.Errorf("slack outbound channel is empty")
	}

	text := formatSlackOutboundText(msg)
	if text == "" {
		text = "..."
	}

	params := goslack.NewPostMessageParameters()
	params.ThreadTimestamp = firstNonEmpty(
		str(msg.ReplyHandle["thread_ts"]),
		msg.ExternalThreadID,
	)
	_, _, err := a.apiClientFactory(conn).PostMessageContext(
		ctx,
		channelID,
		goslack.MsgOptionText(text, false),
		goslack.MsgOptionPostMessageParameters(params),
	)
	return err
}

func (a *SlackAdapter) startRuntime(ctx context.Context, conn Connection, enqueue RuntimePayloadHandler) error {
	identity, err := a.getBotIdentity(ctx, conn)
	if err != nil {
		return err
	}
	runtime := a.socketRuntimeFactory(conn)
	if runtime == nil {
		return fmt.Errorf("slack socket runtime is not configured")
	}
	return runtime.Run(ctx, func(runtimeCtx context.Context, envelope slackSocketEnvelope) error {
		if len(envelope.Payload) == 0 {
			return nil
		}
		// Validate each payload before it enters the generic runtime queue so we fail
		// fast on unsupported Slack event shapes instead of retrying opaque payloads.
		if _, err := parseSlackInbound(envelope.Payload, identity.UserID); err != nil {
			return err
		}
		return enqueue(runtimeCtx, envelope.Payload)
	})
}

func (a *SlackAdapter) processRuntimePayload(ctx context.Context, conn Connection, payload []byte) error {
	identity, err := a.getBotIdentity(ctx, conn)
	if err != nil {
		return err
	}
	events, err := parseSlackInbound(payload, identity.UserID)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return nil
	}
	processor := a.inboundProcessor
	if processor == nil {
		return fmt.Errorf("slack inbound processor is not configured")
	}
	_, err = processor(ctx, conn, events)
	return err
}

func (a *SlackAdapter) fetchBotIdentity(ctx context.Context, conn Connection) (slackBotIdentity, error) {
	response, err := a.apiClientFactory(conn).AuthTestContext(ctx)
	if err != nil {
		return slackBotIdentity{}, err
	}
	identity := slackBotIdentity{
		UserID: strings.TrimSpace(response.UserID),
		BotID:  strings.TrimSpace(response.BotID),
		TeamID: strings.TrimSpace(response.TeamID),
	}
	a.mu.Lock()
	a.identityCache[conn.ID] = identity
	a.mu.Unlock()
	return identity, nil
}

func (a *SlackAdapter) getBotIdentity(ctx context.Context, conn Connection) (slackBotIdentity, error) {
	a.mu.RLock()
	cached, ok := a.identityCache[conn.ID]
	a.mu.RUnlock()
	if ok && cached.UserID != "" {
		return cached, nil
	}
	return a.fetchBotIdentity(ctx, conn)
}

func newSlackWebAPIClient(conn Connection) slackAPIClient {
	secrets := firstNonEmptySecretMap(conn.Secrets, conn.SecretsMasked)
	opts := []goslack.Option{
		goslack.OptionAppLevelToken(strings.TrimSpace(str(secrets["appToken"]))),
	}
	if apiBase := slackAPIBaseURL(slackConfig(conn)); apiBase != defaultSlackSocketBase {
		opts = append(opts, goslack.OptionAPIURL(apiBase))
	}
	return goslack.New(strings.TrimSpace(str(secrets["botToken"])), opts...)
}

func newSlackSocketRuntime(conn Connection) slackSocketRuntime {
	apiClient, ok := newSlackWebAPIClient(conn).(*goslack.Client)
	if !ok {
		return nil
	}
	return &slackSocketModeRuntime{
		client: socketmode.New(apiClient),
	}
}

func (r *slackSocketModeRuntime) Run(ctx context.Context, handle func(context.Context, slackSocketEnvelope) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

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
			case evt, ok := <-r.client.Events:
				if !ok {
					return
				}
				if evt.Type != socketmode.EventTypeEventsAPI || evt.Request == nil {
					continue
				}
				if err := r.client.Ack(*evt.Request); err != nil {
					sendErr(err)
					cancel()
					return
				}
				if err := handle(ctx, slackSocketEnvelope{Payload: append([]byte(nil), evt.Request.Payload...)}); err != nil {
					sendErr(err)
					cancel()
					return
				}
			}
		}
	}()

	runErr := r.client.RunContext(ctx)
	select {
	case err := <-errCh:
		return err
	default:
	}
	return runErr
}

func slackConfig(conn Connection) map[string]any {
	config := map[string]any{}
	for key, value := range conn.Config {
		config[key] = value
	}
	if strings.TrimSpace(str(config["connectionMode"])) == "" && strings.TrimSpace(conn.ConnectionMode) != "" {
		config["connectionMode"] = conn.ConnectionMode
	}
	return config
}

func slackAPIBaseURL(config map[string]any) string {
	if custom := strings.TrimSpace(str(config["apiBaseURL"])); custom != "" {
		if strings.HasSuffix(custom, "/") {
			return custom
		}
		return custom + "/"
	}
	return defaultSlackSocketBase
}

func parseSlackInbound(raw any, botUserID string) ([]InboundEvent, error) {
	event, payloadBytes, err := coerceSlackEvent(raw)
	if err != nil {
		return nil, err
	}
	switch event.Type {
	case slackevents.CallbackEvent:
		return buildSlackInboundEvents(event, payloadBytes, botUserID)
	default:
		return nil, nil
	}
}

func coerceSlackEvent(raw any) (slackevents.EventsAPIEvent, []byte, error) {
	switch value := raw.(type) {
	case slackevents.EventsAPIEvent:
		payloadBytes, err := json.Marshal(value)
		if err != nil {
			payloadBytes = nil
		}
		return value, payloadBytes, nil
	case *slackevents.EventsAPIEvent:
		if value == nil {
			return slackevents.EventsAPIEvent{}, nil, ErrSlackInvalidPayload
		}
		return coerceSlackEvent(*value)
	case []byte:
		return parseSlackPayload(value)
	case json.RawMessage:
		return parseSlackPayload(value)
	case string:
		return parseSlackPayload([]byte(value))
	default:
		payloadBytes, err := json.Marshal(value)
		if err != nil {
			return slackevents.EventsAPIEvent{}, nil, ErrSlackInvalidPayload
		}
		return parseSlackPayload(payloadBytes)
	}
}

func parseSlackPayload(payload []byte) (slackevents.EventsAPIEvent, []byte, error) {
	payload = bytesTrimSpace(payload)
	if len(payload) == 0 {
		return slackevents.EventsAPIEvent{}, nil, ErrSlackInvalidPayload
	}
	event, err := slackevents.ParseEvent(json.RawMessage(payload), slackevents.OptionNoVerifyToken())
	if err != nil {
		return slackevents.EventsAPIEvent{}, nil, err
	}
	return event, append([]byte(nil), payload...), nil
}

func buildSlackInboundEvents(event slackevents.EventsAPIEvent, rawPayload []byte, botUserID string) ([]InboundEvent, error) {
	callback, ok := event.Data.(*slackevents.EventsAPICallbackEvent)
	if !ok {
		return nil, ErrSlackInvalidPayload
	}
	switch inner := event.InnerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		return []InboundEvent{newSlackInboundFromAppMention(event, callback, inner, rawPayload, botUserID)}, nil
	case *slackevents.MessageEvent:
		inbound, ok := newSlackInboundFromMessage(event, callback, inner, rawPayload, botUserID)
		if !ok {
			return nil, nil
		}
		return []InboundEvent{inbound}, nil
	default:
		return nil, nil
	}
}

func newSlackInboundFromAppMention(
	event slackevents.EventsAPIEvent,
	callback *slackevents.EventsAPICallbackEvent,
	inner *slackevents.AppMentionEvent,
	rawPayload []byte,
	botUserID string,
) InboundEvent {
	return InboundEvent{
		Platform:         PlatformSlack,
		EventID:          strings.TrimSpace(callback.EventID),
		MessageID:        strings.TrimSpace(inner.TimeStamp),
		ExternalChatID:   strings.TrimSpace(inner.Channel),
		ExternalThreadID: strings.TrimSpace(inner.ThreadTimeStamp),
		ExternalUserID:   strings.TrimSpace(inner.User),
		ChatType:         normalizeSlackChatType("channel"),
		MentionsBot:      true,
		Text:             strings.TrimSpace(inner.Text),
		Segments:         []RichSegment{{Type: "text", Text: strings.TrimSpace(inner.Text)}},
		Attachments:      buildSlackInboundAttachments(inner.Files),
		ReplyHandle: map[string]any{
			"channel":    strings.TrimSpace(inner.Channel),
			"message_ts": strings.TrimSpace(inner.TimeStamp),
			"thread_ts":  strings.TrimSpace(inner.ThreadTimeStamp),
			"team_id":    strings.TrimSpace(firstNonEmpty(callback.TeamID, event.TeamID)),
		},
		RawPayload: append([]byte(nil), rawPayload...),
		ReceivedAt: time.Now(),
	}
}

func newSlackInboundFromMessage(
	event slackevents.EventsAPIEvent,
	callback *slackevents.EventsAPICallbackEvent,
	inner *slackevents.MessageEvent,
	rawPayload []byte,
	botUserID string,
) (InboundEvent, bool) {
	if inner == nil {
		return InboundEvent{}, false
	}
	if strings.TrimSpace(inner.SubType) != "" || strings.TrimSpace(inner.BotID) != "" {
		return InboundEvent{}, false
	}
	text := strings.TrimSpace(inner.Text)
	threadTS := strings.TrimSpace(inner.ThreadTimeStamp)
	messageTS := strings.TrimSpace(inner.TimeStamp)
	if inner.Message != nil {
		if text == "" {
			text = strings.TrimSpace(inner.Message.Text)
		}
		if threadTS == "" {
			threadTS = strings.TrimSpace(inner.Message.ThreadTimestamp)
		}
		if messageTS == "" {
			messageTS = strings.TrimSpace(inner.Message.Timestamp)
		}
	}
	messageTS = firstNonEmpty(messageTS, strings.TrimSpace(inner.EventTimeStamp))
	if messageTS == "" {
		return InboundEvent{}, false
	}
	mentionsBot := detectSlackMention(text, botUserID)
	attachments := []goslack.File(nil)
	if inner.Message != nil && len(inner.Message.Files) > 0 {
		attachments = inner.Message.Files
	}
	return InboundEvent{
		Platform:         PlatformSlack,
		EventID:          strings.TrimSpace(callback.EventID),
		MessageID:        messageTS,
		ExternalChatID:   strings.TrimSpace(inner.Channel),
		ExternalThreadID: threadTS,
		ExternalUserID:   strings.TrimSpace(inner.User),
		ChatType:         normalizeSlackChatType(inner.ChannelType),
		MentionsBot:      mentionsBot,
		Text:             text,
		Segments:         []RichSegment{{Type: "text", Text: text}},
		Attachments:      buildSlackInboundAttachments(attachments),
		ReplyHandle: map[string]any{
			"channel":    strings.TrimSpace(inner.Channel),
			"message_ts": messageTS,
			"thread_ts":  threadTS,
			"team_id":    strings.TrimSpace(firstNonEmpty(callback.TeamID, event.TeamID)),
		},
		RawPayload: append([]byte(nil), rawPayload...),
		ReceivedAt: time.Now(),
	}, true
}

func buildSlackInboundAttachments(files []goslack.File) []InboundAttachment {
	if len(files) == 0 {
		return nil
	}
	attachments := make([]InboundAttachment, 0, len(files))
	for _, file := range files {
		attachmentType := detectSlackAttachmentType(file.Mimetype)
		attachment := buildInboundAttachment(PlatformSlack, attachmentType, file.ID, firstNonEmpty(file.Name, file.Title))
		attachment.URL = strings.TrimSpace(firstNonEmpty(file.URLPrivateDownload, file.URLPrivate))
		attachment.ContentType = firstNonEmpty(strings.TrimSpace(file.Mimetype), attachment.ContentType)
		if attachment.MediaRef != "" {
			attachments = append(attachments, attachment)
		}
	}
	return attachments
}

func detectSlackAttachmentType(mime string) string {
	mime = strings.TrimSpace(strings.ToLower(mime))
	switch {
	case strings.HasPrefix(mime, "image/"):
		return "image"
	case strings.HasPrefix(mime, "audio/"):
		return "audio"
	case strings.HasPrefix(mime, "video/"):
		return "video"
	default:
		return "file"
	}
}

func detectSlackMention(text, botUserID string) bool {
	text = strings.TrimSpace(text)
	botUserID = strings.TrimSpace(botUserID)
	if text == "" || botUserID == "" {
		return false
	}
	return strings.Contains(text, "<@"+botUserID+">")
}

func normalizeSlackChatType(channelType string) string {
	switch strings.TrimSpace(channelType) {
	case "im":
		return "p2p"
	default:
		return "group"
	}
}

func formatSlackOutboundText(msg OutboundEnvelope) string {
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

func isExpectedSlackRuntimeStop(err error) bool {
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

func bytesTrimSpace(payload []byte) []byte {
	return []byte(strings.TrimSpace(string(payload)))
}
