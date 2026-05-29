package im

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTelegramValidateConfig(t *testing.T) {
	t.Parallel()

	adapter := NewTelegramAdapter()

	tests := []struct {
		name    string
		config  map[string]any
		secrets map[string]any
		wantErr error
	}{
		{
			name: "valid polling config",
			config: map[string]any{
				"connectionMode": "polling",
			},
			secrets: map[string]any{
				"token": "bot-token",
			},
		},
		{
			name:    "missing token",
			config:  map[string]any{},
			secrets: map[string]any{},
			wantErr: ErrInvalidConnectionConfig,
		},
		{
			name: "unsupported mode",
			config: map[string]any{
				"connectionMode": "webhook",
			},
			secrets: map[string]any{
				"token": "bot-token",
			},
			wantErr: ErrTelegramUnsupportedConnectionMode,
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

func TestParseTelegramInbound(t *testing.T) {
	t.Parallel()

	payload := map[string]any{
		"update_id": 101,
		"message": map[string]any{
			"message_id": 202,
			"from": map[string]any{
				"id": 303,
			},
			"chat": map[string]any{
				"id":   -404,
				"type": "supergroup",
			},
			"text":              "hello @dotblue_bot",
			"message_thread_id": 505,
			"photo": []any{
				map[string]any{"file_id": "small"},
				map[string]any{"file_id": "large"},
			},
			"document": map[string]any{
				"file_id":   "doc_file",
				"file_name": "report.pdf",
			},
		},
	}

	events, err := parseTelegramInbound(payload, "dotblue_bot")
	if err != nil {
		t.Fatalf("parseTelegramInbound() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("event count = %d, want 1", len(events))
	}

	event := events[0]
	if event.Platform != PlatformTelegram {
		t.Fatalf("Platform = %q, want %q", event.Platform, PlatformTelegram)
	}
	if event.EventID != "101" {
		t.Fatalf("EventID = %q, want 101", event.EventID)
	}
	if event.MessageID != "202" {
		t.Fatalf("MessageID = %q, want 202", event.MessageID)
	}
	if event.ExternalChatID != "-404" {
		t.Fatalf("ExternalChatID = %q, want -404", event.ExternalChatID)
	}
	if event.ExternalThreadID != "505" {
		t.Fatalf("ExternalThreadID = %q, want 505", event.ExternalThreadID)
	}
	if event.ExternalUserID != "303" {
		t.Fatalf("ExternalUserID = %q, want 303", event.ExternalUserID)
	}
	if event.ChatType != "group" {
		t.Fatalf("ChatType = %q, want group", event.ChatType)
	}
	if !event.MentionsBot {
		t.Fatal("MentionsBot = false, want true")
	}
	if event.Text != "hello @dotblue_bot" {
		t.Fatalf("Text = %q, want hello @dotblue_bot", event.Text)
	}
	if len(event.Attachments) != 2 {
		t.Fatalf("attachment count = %d, want 2", len(event.Attachments))
	}
	if event.Attachments[0].MediaRef != BuildInboundMediaRef(PlatformTelegram, "image", "large") {
		t.Fatalf("first attachment media ref = %q, want telegram image ref", event.Attachments[0].MediaRef)
	}
	if event.Attachments[1].MediaRef != BuildInboundMediaRef(PlatformTelegram, "file", "doc_file") {
		t.Fatalf("second attachment media ref = %q, want telegram file ref", event.Attachments[1].MediaRef)
	}
	if event.ReplyHandle["chat_id"] != "-404" {
		t.Fatalf("ReplyHandle chat_id = %v, want -404", event.ReplyHandle["chat_id"])
	}
	if len(event.RawPayload) == 0 {
		t.Fatal("RawPayload is empty")
	}
}

func TestParseTelegramInboundFallsBackToMessageID(t *testing.T) {
	t.Parallel()

	events, err := parseTelegramInbound(map[string]any{
		"message": map[string]any{
			"message_id": 11,
			"from":       map[string]any{"id": 22},
			"chat":       map[string]any{"id": 33, "type": "private"},
			"text":       "hello",
		},
	}, "")
	if err != nil {
		t.Fatalf("parseTelegramInbound() error = %v", err)
	}
	if events[0].EventID != "11" {
		t.Fatalf("EventID = %q, want 11", events[0].EventID)
	}
	if events[0].ChatType != "p2p" {
		t.Fatalf("ChatType = %q, want p2p", events[0].ChatType)
	}
	if events[0].MentionsBot {
		t.Fatal("MentionsBot = true, want false")
	}
}

func TestParseTelegramInboundInvalidPayload(t *testing.T) {
	t.Parallel()

	tests := []any{
		1,
		map[string]any{},
		map[string]any{
			"message": map[string]any{
				"message_id": 1,
			},
		},
	}

	for _, raw := range tests {
		raw := raw
		t.Run("invalid", func(t *testing.T) {
			t.Parallel()
			_, err := parseTelegramInbound(raw, "")
			if !errors.Is(err, ErrTelegramInvalidPayload) {
				t.Fatalf("parseTelegramInbound() error = %v, want %v", err, ErrTelegramInvalidPayload)
			}
		})
	}
}

func TestTelegramTestConnection(t *testing.T) {
	t.Parallel()

	adapter := NewTelegramAdapter()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/botbot-token/getMe" {
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"ok":true,"result":{"id":123,"username":"dotblue_bot"}}`))
	}))
	defer server.Close()

	conn := Connection{
		ID:             "conn_1",
		Platform:       PlatformTelegram,
		ConnectionMode: TelegramConnectionModePolling,
		Config: map[string]any{
			"apiBaseURL": server.URL,
		},
		Secrets: map[string]any{
			"token": "bot-token",
		},
	}

	if err := adapter.TestConnection(context.Background(), conn); err != nil {
		t.Fatalf("TestConnection() error = %v", err)
	}
	if cached, ok := adapter.botCache[conn.ID]; !ok || cached.Username != "dotblue_bot" {
		t.Fatalf("bot cache = %+v, want dotblue_bot", cached)
	}
}

func TestTelegramSendOutbound(t *testing.T) {
	t.Parallel()

	adapter := NewTelegramAdapter()
	var calls []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		calls = append(calls, r.URL.Path)

		var req map[string]any
		_ = json.NewDecoder(r.Body).Decode(&req)
		switch r.URL.Path {
		case "/botbot-token/sendMessage":
			if req["chat_id"] != float64(-1001) {
				t.Fatalf("chat_id = %v, want -1001", req["chat_id"])
			}
			if req["reply_to_message_id"] != float64(77) {
				t.Fatalf("reply_to_message_id = %v, want 77", req["reply_to_message_id"])
			}
			if req["text"] != "reply content" {
				t.Fatalf("text = %v, want reply content", req["text"])
			}
		case "/botbot-token/sendPhoto":
			if req["photo"] != "photo_file" {
				t.Fatalf("photo = %v, want photo_file", req["photo"])
			}
			if req["chat_id"] != float64(-1001) {
				t.Fatalf("photo chat_id = %v, want -1001", req["chat_id"])
			}
		default:
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":1}}`))
	}))
	defer server.Close()

	conn := Connection{
		ID:             "conn_1",
		Platform:       PlatformTelegram,
		ConnectionMode: TelegramConnectionModePolling,
		Config: map[string]any{
			"apiBaseURL": server.URL,
		},
		Secrets: map[string]any{
			"token": "bot-token",
		},
	}

	err := adapter.SendOutbound(context.Background(), conn, OutboundEnvelope{
		ExternalChatID: "-1001",
		Text:           "reply content",
		ReplyHandle: map[string]any{
			"chat_id":    "-1001",
			"message_id": "77",
		},
		Attachments: []OutboundAttachment{
			{
				Type:     "image",
				MediaRef: BuildInboundMediaRef(PlatformTelegram, "image", "photo_file"),
			},
		},
	})
	if err != nil {
		t.Fatalf("SendOutbound() error = %v", err)
	}
	if len(calls) != 2 {
		t.Fatalf("request count = %d, want 2", len(calls))
	}
}

func TestBuildTelegramOutboundRequestsSegmentsOnly(t *testing.T) {
	t.Parallel()

	requests, err := buildTelegramOutboundRequests(OutboundEnvelope{
		Segments: []RichSegment{
			{Type: "text", Text: "hello"},
			{Type: "mention", Text: "@bot"},
		},
	})
	if err != nil {
		t.Fatalf("buildTelegramOutboundRequests() error = %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("request count = %d, want 1", len(requests))
	}
	if got := requests[0].Body["text"]; got != "hello@bot" {
		t.Fatalf("request text = %v, want hello@bot", got)
	}
}

func TestBuildTelegramAttachmentRequestRejectsForeignMedia(t *testing.T) {
	t.Parallel()

	_, err := buildTelegramAttachmentRequest(OutboundAttachment{
		Type:     "image",
		MediaRef: BuildInboundMediaRef("feishu", "image", "img_1"),
	})
	if err == nil || !strings.Contains(err.Error(), "platform mismatch") {
		t.Fatalf("buildTelegramAttachmentRequest() error = %v, want platform mismatch", err)
	}
}
