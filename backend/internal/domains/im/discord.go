package im

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	PlatformDiscord       = "discord"
	DiscordConnectionMode = "gateway"
)

var (
	ErrDiscordInvalidPayload            = errors.New("invalid discord payload")
	ErrDiscordUnsupportedConnectionMode = errors.New("discord only supports gateway mode in v1")
)

type DiscordAdapter struct {
	mu                 sync.RWMutex
	identityCache      map[string]discordBotIdentity
	apiClientFactory   func(conn Connection) discordAPIClient
	runtimeFactory     func(conn Connection) discordRuntime
	inboundProcessor   feishuInboundProcessor
	reconnectBaseDelay time.Duration
	reconnectMaxDelay  time.Duration
}

type discordBotIdentity struct {
	UserID   string
	Username string
}

type discordAPIClient interface {
	User(userID string, options ...discordgo.RequestOption) (*discordgo.User, error)
	ChannelMessageSendComplex(channelID string, data *discordgo.MessageSend, options ...discordgo.RequestOption) (*discordgo.Message, error)
}

type discordRuntime interface {
	Run(ctx context.Context, handle func(context.Context, discordEnvelope) error) error
}

type discordEnvelope struct {
	Payload []byte
}

type discordGatewayRuntime struct {
	session *discordgo.Session
}

func init() {
	RegisterAdapter(NewDiscordAdapter())
}

func NewDiscordAdapter() *DiscordAdapter {
	return &DiscordAdapter{
		identityCache:      map[string]discordBotIdentity{},
		apiClientFactory:   newDiscordAPIClient,
		runtimeFactory:     newDiscordRuntime,
		inboundProcessor:   defaultInboundPipeline.PersistEvents,
		reconnectBaseDelay: time.Second,
		reconnectMaxDelay:  30 * time.Second,
	}
}

func (a *DiscordAdapter) Platform() string {
	return PlatformDiscord
}

func (a *DiscordAdapter) Start(ctx context.Context, conn Connection) error {
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
		IsExpectedStop:     isExpectedDiscordRuntimeStop,
		ReconnectBaseDelay: a.reconnectBaseDelay,
		ReconnectMaxDelay:  a.reconnectMaxDelay,
	})
}

func (a *DiscordAdapter) Stop(ctx context.Context, connectionID string) error {
	a.mu.Lock()
	delete(a.identityCache, connectionID)
	a.mu.Unlock()
	return defaultRuntimeManager.Stop(ctx, connectionID)
}

func (a *DiscordAdapter) ValidateConfig(config map[string]any, secrets map[string]any) error {
	if strings.TrimSpace(str(secrets["token"])) == "" {
		return ErrInvalidConnectionConfig
	}
	mode := firstNonEmpty(strings.TrimSpace(str(config["connectionMode"])), strings.TrimSpace(str(config["mode"])))
	if mode == "" {
		mode = DiscordConnectionMode
	}
	if mode != DiscordConnectionMode {
		return ErrDiscordUnsupportedConnectionMode
	}
	return nil
}

func (a *DiscordAdapter) TestConnection(ctx context.Context, conn Connection) error {
	secrets := firstNonEmptySecretMap(conn.Secrets, conn.SecretsMasked)
	if err := a.ValidateConfig(discordConfig(conn), secrets); err != nil {
		return err
	}
	identity, err := a.fetchBotIdentity(ctx, conn)
	if err != nil {
		return err
	}
	if identity.UserID == "" {
		return fmt.Errorf("discord current bot user id is empty")
	}
	return nil
}

func (a *DiscordAdapter) ParseInbound(ctx context.Context, raw any) ([]InboundEvent, error) {
	return parseDiscordInbound(raw, "")
}

func (a *DiscordAdapter) SendOutbound(ctx context.Context, conn Connection, msg OutboundEnvelope) error {
	secrets := firstNonEmptySecretMap(conn.Secrets, conn.SecretsMasked)
	if err := a.ValidateConfig(discordConfig(conn), secrets); err != nil {
		return err
	}

	channelID := firstNonEmpty(str(msg.ReplyHandle["channel_id"]), msg.ExternalChatID)
	if strings.TrimSpace(channelID) == "" {
		return fmt.Errorf("discord outbound channel is empty")
	}

	content := formatDiscordOutboundText(msg)
	if content == "" {
		content = "..."
	}

	send := &discordgo.MessageSend{
		Content: content,
		AllowedMentions: &discordgo.MessageAllowedMentions{
			RepliedUser: false,
		},
	}
	if replyTo := strings.TrimSpace(str(msg.ReplyHandle["message_id"])); replyTo != "" {
		failIfNotExists := false
		send.Reference = &discordgo.MessageReference{
			MessageID:       replyTo,
			ChannelID:       channelID,
			GuildID:         strings.TrimSpace(str(msg.ReplyHandle["guild_id"])),
			FailIfNotExists: &failIfNotExists,
		}
	}
	_, err := a.apiClientFactory(conn).ChannelMessageSendComplex(channelID, send)
	return err
}

func (a *DiscordAdapter) startRuntime(ctx context.Context, conn Connection, enqueue RuntimePayloadHandler) error {
	identity, err := a.getBotIdentity(ctx, conn)
	if err != nil {
		return err
	}
	runtime := a.runtimeFactory(conn)
	if runtime == nil {
		return fmt.Errorf("discord runtime is not configured")
	}
	return runtime.Run(ctx, func(runtimeCtx context.Context, envelope discordEnvelope) error {
		if len(envelope.Payload) == 0 {
			return nil
		}
		if _, err := parseDiscordInbound(envelope.Payload, identity.UserID); err != nil {
			return err
		}
		return enqueue(runtimeCtx, envelope.Payload)
	})
}

func (a *DiscordAdapter) processRuntimePayload(ctx context.Context, conn Connection, payload []byte) error {
	identity, err := a.getBotIdentity(ctx, conn)
	if err != nil {
		return err
	}
	events, err := parseDiscordInbound(payload, identity.UserID)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return nil
	}
	processor := a.inboundProcessor
	if processor == nil {
		return fmt.Errorf("discord inbound processor is not configured")
	}
	_, err = processor(ctx, conn, events)
	return err
}

func (a *DiscordAdapter) fetchBotIdentity(ctx context.Context, conn Connection) (discordBotIdentity, error) {
	user, err := a.apiClientFactory(conn).User("@me")
	if err != nil {
		return discordBotIdentity{}, err
	}
	identity := discordBotIdentity{
		UserID:   strings.TrimSpace(user.ID),
		Username: strings.TrimSpace(user.Username),
	}
	a.mu.Lock()
	a.identityCache[conn.ID] = identity
	a.mu.Unlock()
	return identity, nil
}

func (a *DiscordAdapter) getBotIdentity(ctx context.Context, conn Connection) (discordBotIdentity, error) {
	a.mu.RLock()
	cached, ok := a.identityCache[conn.ID]
	a.mu.RUnlock()
	if ok && cached.UserID != "" {
		return cached, nil
	}
	return a.fetchBotIdentity(ctx, conn)
}

func newDiscordAPIClient(conn Connection) discordAPIClient {
	token := strings.TrimSpace(str(firstNonEmptySecretMap(conn.Secrets, conn.SecretsMasked)["token"]))
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil
	}
	session.Identify.Intents = discordgo.IntentGuildMessages | discordgo.IntentDirectMessages | discordgo.IntentMessageContent
	return session
}

func newDiscordRuntime(conn Connection) discordRuntime {
	apiClient, ok := newDiscordAPIClient(conn).(*discordgo.Session)
	if !ok || apiClient == nil {
		return nil
	}
	return &discordGatewayRuntime{session: apiClient}
}

func (r *discordGatewayRuntime) Run(ctx context.Context, handle func(context.Context, discordEnvelope) error) error {
	if r.session == nil {
		return fmt.Errorf("discord session is nil")
	}

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

	remove := r.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m == nil || m.Message == nil {
			return
		}
		payload, err := json.Marshal(m.Message)
		if err != nil {
			sendErr(err)
			cancel()
			return
		}
		if err := handle(ctx, discordEnvelope{Payload: payload}); err != nil {
			sendErr(err)
			cancel()
		}
	})
	defer remove()

	if err := r.session.Open(); err != nil {
		return err
	}
	defer r.session.Close()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func discordConfig(conn Connection) map[string]any {
	config := map[string]any{}
	for key, value := range conn.Config {
		config[key] = value
	}
	if strings.TrimSpace(str(config["connectionMode"])) == "" && strings.TrimSpace(conn.ConnectionMode) != "" {
		config["connectionMode"] = conn.ConnectionMode
	}
	return config
}

func parseDiscordInbound(raw any, botUserID string) ([]InboundEvent, error) {
	message, payload, err := coerceDiscordMessage(raw)
	if err != nil {
		return nil, err
	}
	inbound, ok := buildDiscordInboundEvent(message, payload, botUserID)
	if !ok {
		return nil, nil
	}
	return []InboundEvent{inbound}, nil
}

func coerceDiscordMessage(raw any) (*discordgo.Message, []byte, error) {
	switch value := raw.(type) {
	case *discordgo.MessageCreate:
		if value == nil || value.Message == nil {
			return nil, nil, ErrDiscordInvalidPayload
		}
		payload, err := json.Marshal(value.Message)
		if err != nil {
			payload = nil
		}
		return value.Message, payload, nil
	case discordgo.MessageCreate:
		return coerceDiscordMessage(&value)
	case *discordgo.Message:
		if value == nil {
			return nil, nil, ErrDiscordInvalidPayload
		}
		payload, err := json.Marshal(value)
		if err != nil {
			payload = nil
		}
		return value, payload, nil
	case discordgo.Message:
		return coerceDiscordMessage(&value)
	case []byte:
		return parseDiscordMessagePayload(value)
	case json.RawMessage:
		return parseDiscordMessagePayload(value)
	case string:
		return parseDiscordMessagePayload([]byte(value))
	default:
		payload, err := json.Marshal(value)
		if err != nil {
			return nil, nil, ErrDiscordInvalidPayload
		}
		return parseDiscordMessagePayload(payload)
	}
}

func parseDiscordMessagePayload(payload []byte) (*discordgo.Message, []byte, error) {
	payload = bytesTrimSpace(payload)
	if len(payload) == 0 {
		return nil, nil, ErrDiscordInvalidPayload
	}
	var message discordgo.Message
	if err := json.Unmarshal(payload, &message); err != nil {
		return nil, nil, err
	}
	return &message, append([]byte(nil), payload...), nil
}

func buildDiscordInboundEvent(message *discordgo.Message, rawPayload []byte, botUserID string) (InboundEvent, bool) {
	if message == nil || message.Author == nil {
		return InboundEvent{}, false
	}
	if message.Author.Bot || strings.TrimSpace(message.ID) == "" || strings.TrimSpace(message.ChannelID) == "" {
		return InboundEvent{}, false
	}

	text := strings.TrimSpace(message.Content)
	attachments := buildDiscordInboundAttachments(message.Attachments)
	if text == "" && len(attachments) == 0 {
		return InboundEvent{}, false
	}

	segments := []RichSegment{}
	if text != "" {
		segments = append(segments, RichSegment{Type: "text", Text: text})
	}
	replyHandle := map[string]any{
		"channel_id": message.ChannelID,
		"guild_id":   message.GuildID,
		"message_id": message.ID,
	}
	if message.MessageReference != nil {
		replyHandle["reference_message_id"] = strings.TrimSpace(message.MessageReference.MessageID)
	}

	return InboundEvent{
		Platform:       PlatformDiscord,
		EventID:        strings.TrimSpace(message.ID),
		MessageID:      strings.TrimSpace(message.ID),
		ExternalChatID: strings.TrimSpace(message.ChannelID),
		ExternalUserID: strings.TrimSpace(message.Author.ID),
		ChatType:       normalizeDiscordChatType(message.GuildID),
		MentionsBot:    detectDiscordMention(message.Mentions, botUserID),
		Text:           text,
		Segments:       segments,
		Attachments:    attachments,
		ReplyHandle:    replyHandle,
		RawPayload:     append([]byte(nil), rawPayload...),
		ReceivedAt:     time.Now(),
	}, true
}

func buildDiscordInboundAttachments(attachments []*discordgo.MessageAttachment) []InboundAttachment {
	if len(attachments) == 0 {
		return nil
	}
	result := make([]InboundAttachment, 0, len(attachments))
	for _, attachment := range attachments {
		if attachment == nil || strings.TrimSpace(attachment.ID) == "" {
			continue
		}
		attachmentType := detectDiscordAttachmentType(attachment.ContentType)
		item := buildInboundAttachment(PlatformDiscord, attachmentType, attachment.ID, attachment.Filename)
		item.URL = strings.TrimSpace(firstNonEmpty(attachment.URL, attachment.ProxyURL))
		item.ContentType = firstNonEmpty(strings.TrimSpace(attachment.ContentType), item.ContentType)
		if item.MediaRef != "" {
			result = append(result, item)
		}
	}
	return result
}

func detectDiscordAttachmentType(contentType string) string {
	contentType = strings.TrimSpace(strings.ToLower(contentType))
	switch {
	case strings.HasPrefix(contentType, "image/"):
		return "image"
	case strings.HasPrefix(contentType, "audio/"):
		return "audio"
	case strings.HasPrefix(contentType, "video/"):
		return "video"
	default:
		return "file"
	}
}

func detectDiscordMention(mentions []*discordgo.User, botUserID string) bool {
	botUserID = strings.TrimSpace(botUserID)
	if botUserID == "" {
		return false
	}
	for _, mention := range mentions {
		if mention != nil && strings.TrimSpace(mention.ID) == botUserID {
			return true
		}
	}
	return false
}

func normalizeDiscordChatType(guildID string) string {
	if strings.TrimSpace(guildID) == "" {
		return "p2p"
	}
	return "group"
}

func formatDiscordOutboundText(msg OutboundEnvelope) string {
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

func isExpectedDiscordRuntimeStop(err error) bool {
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
