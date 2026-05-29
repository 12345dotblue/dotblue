package im

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

type stubDiscordAPIClient struct {
	userFn                    func(userID string, options ...discordgo.RequestOption) (*discordgo.User, error)
	channelMessageSendComplex func(channelID string, data *discordgo.MessageSend, options ...discordgo.RequestOption) (*discordgo.Message, error)
}

func (s *stubDiscordAPIClient) User(userID string, options ...discordgo.RequestOption) (*discordgo.User, error) {
	if s.userFn != nil {
		return s.userFn(userID, options...)
	}
	return nil, nil
}

func (s *stubDiscordAPIClient) ChannelMessageSendComplex(channelID string, data *discordgo.MessageSend, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	if s.channelMessageSendComplex != nil {
		return s.channelMessageSendComplex(channelID, data, options...)
	}
	return nil, nil
}

func TestDiscordValidateConfig(t *testing.T) {
	t.Parallel()

	adapter := NewDiscordAdapter()
	tests := []struct {
		name    string
		config  map[string]any
		secrets map[string]any
		wantErr error
	}{
		{
			name:   "valid gateway config",
			config: map[string]any{"connectionMode": DiscordConnectionMode},
			secrets: map[string]any{
				"token": "discord-bot-token",
			},
		},
		{
			name:    "missing token",
			config:  map[string]any{},
			secrets: map[string]any{},
			wantErr: ErrInvalidConnectionConfig,
		},
		{
			name:   "unsupported mode",
			config: map[string]any{"connectionMode": "webhook"},
			secrets: map[string]any{
				"token": "discord-bot-token",
			},
			wantErr: ErrDiscordUnsupportedConnectionMode,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := adapter.ValidateConfig(tt.config, tt.secrets)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateConfig() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseDiscordInbound(t *testing.T) {
	t.Parallel()

	raw := []byte(`{
		"id":"msg_1",
		"channel_id":"chan_1",
		"guild_id":"guild_1",
		"content":"hello <@bot_1>",
		"author":{"id":"user_1","username":"alice","bot":false},
		"attachments":[
			{
				"id":"att_1",
				"url":"https://cdn.discordapp.com/files/att_1",
				"filename":"image.png",
				"content_type":"image/png"
			}
		],
		"mentions":[{"id":"bot_1","username":"bot"}]
	}`)

	events, err := parseDiscordInbound(raw, "bot_1")
	if err != nil {
		t.Fatalf("parseDiscordInbound() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("event count = %d, want 1", len(events))
	}

	event := events[0]
	if event.Platform != PlatformDiscord {
		t.Fatalf("Platform = %q, want %q", event.Platform, PlatformDiscord)
	}
	if event.EventID != "msg_1" || event.MessageID != "msg_1" {
		t.Fatalf("message ids = (%q,%q), want msg_1", event.EventID, event.MessageID)
	}
	if event.ExternalChatID != "chan_1" {
		t.Fatalf("ExternalChatID = %q, want chan_1", event.ExternalChatID)
	}
	if event.ExternalUserID != "user_1" {
		t.Fatalf("ExternalUserID = %q, want user_1", event.ExternalUserID)
	}
	if event.ChatType != "group" {
		t.Fatalf("ChatType = %q, want group", event.ChatType)
	}
	if !event.MentionsBot {
		t.Fatal("MentionsBot = false, want true")
	}
	if len(event.Attachments) != 1 {
		t.Fatalf("attachment count = %d, want 1", len(event.Attachments))
	}
	if event.Attachments[0].MediaRef != BuildInboundMediaRef(PlatformDiscord, "image", "att_1") {
		t.Fatalf("attachment media ref = %q, want discord image ref", event.Attachments[0].MediaRef)
	}
	if event.ReplyHandle["channel_id"] != "chan_1" {
		t.Fatalf("ReplyHandle channel_id = %v, want chan_1", event.ReplyHandle["channel_id"])
	}
	if len(event.RawPayload) == 0 {
		t.Fatal("RawPayload is empty")
	}
}

func TestParseDiscordInboundDM(t *testing.T) {
	t.Parallel()

	msg := &discordgo.Message{
		ID:        "msg_dm",
		ChannelID: "dm_1",
		Content:   "hello",
		Author: &discordgo.User{
			ID:       "user_dm",
			Username: "dm-user",
		},
	}

	events, err := parseDiscordInbound(msg, "bot_1")
	if err != nil {
		t.Fatalf("parseDiscordInbound() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("event count = %d, want 1", len(events))
	}
	if events[0].ChatType != "p2p" {
		t.Fatalf("ChatType = %q, want p2p", events[0].ChatType)
	}
	if events[0].MentionsBot {
		t.Fatal("MentionsBot = true, want false")
	}
}

func TestProcessDiscordRuntimePayload(t *testing.T) {
	t.Parallel()

	adapter := NewDiscordAdapter()
	adapter.identityCache["conn_1"] = discordBotIdentity{UserID: "bot_1"}

	var captured []InboundEvent
	adapter.inboundProcessor = func(ctx context.Context, conn Connection, events []InboundEvent) (inboundPersistResult, error) {
		captured = append(captured, events...)
		return inboundPersistResult{}, nil
	}

	conn := Connection{
		ID:             "conn_1",
		Platform:       PlatformDiscord,
		ConnectionMode: DiscordConnectionMode,
		Secrets: map[string]any{
			"token": "discord-bot-token",
		},
	}

	payload := []byte(`{
		"id":"msg_2",
		"channel_id":"dm_2",
		"content":"hello",
		"author":{"id":"user_2","username":"bob","bot":false},
		"mentions":[]
	}`)

	if err := adapter.processRuntimePayload(context.Background(), conn, payload); err != nil {
		t.Fatalf("processRuntimePayload() error = %v", err)
	}
	if len(captured) != 1 {
		t.Fatalf("captured event count = %d, want 1", len(captured))
	}
	if captured[0].ChatType != "p2p" {
		t.Fatalf("ChatType = %q, want p2p", captured[0].ChatType)
	}
}

func TestDiscordTestConnectionAndSendOutbound(t *testing.T) {
	t.Parallel()

	adapter := NewDiscordAdapter()

	var sentChannelID string
	var sentMessage *discordgo.MessageSend

	adapter.apiClientFactory = func(conn Connection) discordAPIClient {
		return &stubDiscordAPIClient{
			userFn: func(userID string, options ...discordgo.RequestOption) (*discordgo.User, error) {
				if userID != "@me" {
					t.Fatalf("User() userID = %q, want @me", userID)
				}
				return &discordgo.User{
					ID:       "bot_1",
					Username: "dotblue-bot",
					Bot:      true,
				}, nil
			},
			channelMessageSendComplex: func(channelID string, data *discordgo.MessageSend, options ...discordgo.RequestOption) (*discordgo.Message, error) {
				sentChannelID = channelID
				sentMessage = data
				return &discordgo.Message{ID: "reply_1"}, nil
			},
		}
	}

	conn := Connection{
		ID:             "conn_1",
		Platform:       PlatformDiscord,
		ConnectionMode: DiscordConnectionMode,
		Secrets: map[string]any{
			"token": "discord-bot-token",
		},
	}

	if err := adapter.TestConnection(context.Background(), conn); err != nil {
		t.Fatalf("TestConnection() error = %v", err)
	}
	if cached, ok := adapter.identityCache[conn.ID]; !ok || cached.UserID != "bot_1" {
		t.Fatalf("identity cache = %+v, want bot_1", cached)
	}

	err := adapter.SendOutbound(context.Background(), conn, OutboundEnvelope{
		ExternalChatID: "chan_1",
		Text:           "reply body",
		ReplyHandle: map[string]any{
			"channel_id": "chan_1",
			"guild_id":   "guild_1",
			"message_id": "msg_1",
		},
		Attachments: []OutboundAttachment{
			{Name: "demo.png", Type: "image"},
		},
	})
	if err != nil {
		t.Fatalf("SendOutbound() error = %v", err)
	}
	if sentChannelID != "chan_1" {
		t.Fatalf("sent channel = %q, want chan_1", sentChannelID)
	}
	if sentMessage == nil {
		t.Fatal("sent message is nil")
	}
	if sentMessage.Reference == nil || sentMessage.Reference.MessageID != "msg_1" {
		t.Fatalf("message reference = %+v, want msg_1", sentMessage.Reference)
	}
	if sentMessage.AllowedMentions == nil || sentMessage.AllowedMentions.RepliedUser {
		t.Fatalf("AllowedMentions = %+v, want replied user disabled", sentMessage.AllowedMentions)
	}
	if !strings.Contains(sentMessage.Content, "reply body") {
		t.Fatalf("sent content = %q, want reply body", sentMessage.Content)
	}
	if !strings.Contains(sentMessage.Content, "[demo.png]") {
		t.Fatalf("sent content = %q, want attachment fallback", sentMessage.Content)
	}
}
