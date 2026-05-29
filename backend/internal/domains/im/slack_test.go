package im

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSlackValidateConfig(t *testing.T) {
	t.Parallel()

	adapter := NewSlackAdapter()

	tests := []struct {
		name    string
		config  map[string]any
		secrets map[string]any
		wantErr error
	}{
		{
			name:   "valid socket mode",
			config: map[string]any{"connectionMode": SlackConnectionMode},
			secrets: map[string]any{
				"botToken": "xoxb-bot",
				"appToken": "xapp-app",
			},
		},
		{
			name:    "missing token",
			config:  map[string]any{"connectionMode": SlackConnectionMode},
			secrets: map[string]any{},
			wantErr: ErrInvalidConnectionConfig,
		},
		{
			name:   "bad prefix",
			config: map[string]any{"connectionMode": SlackConnectionMode},
			secrets: map[string]any{
				"botToken": "bot",
				"appToken": "app",
			},
			wantErr: ErrInvalidConnectionConfig,
		},
		{
			name:   "unsupported mode",
			config: map[string]any{"connectionMode": "polling"},
			secrets: map[string]any{
				"botToken": "xoxb-bot",
				"appToken": "xapp-app",
			},
			wantErr: ErrSlackUnsupportedConnectionMode,
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

func TestParseSlackInboundAppMention(t *testing.T) {
	t.Parallel()

	raw := []byte(`{
		"token":"verify-token",
		"team_id":"T_1",
		"api_app_id":"A_1",
		"type":"event_callback",
		"event":{
			"type":"app_mention",
			"user":"U_1",
			"text":"hi <@U_BOT>",
			"ts":"171.01",
			"thread_ts":"171.00",
			"channel":"C_1",
			"event_ts":"171.01",
			"files":[
				{
					"id":"F_1",
					"name":"image.png",
					"title":"image.png",
					"mimetype":"image/png",
					"url_private_download":"https://files.example.com/f1"
				}
			]
		},
		"event_id":"Ev_1",
		"event_time":171
	}`)

	events, err := parseSlackInbound(raw, "U_BOT")
	if err != nil {
		t.Fatalf("parseSlackInbound() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("event count = %d, want 1", len(events))
	}

	event := events[0]
	if event.Platform != PlatformSlack {
		t.Fatalf("Platform = %q, want %q", event.Platform, PlatformSlack)
	}
	if event.EventID != "Ev_1" {
		t.Fatalf("EventID = %q, want Ev_1", event.EventID)
	}
	if event.MessageID != "171.01" {
		t.Fatalf("MessageID = %q, want 171.01", event.MessageID)
	}
	if event.ExternalChatID != "C_1" {
		t.Fatalf("ExternalChatID = %q, want C_1", event.ExternalChatID)
	}
	if event.ExternalThreadID != "171.00" {
		t.Fatalf("ExternalThreadID = %q, want 171.00", event.ExternalThreadID)
	}
	if event.ExternalUserID != "U_1" {
		t.Fatalf("ExternalUserID = %q, want U_1", event.ExternalUserID)
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
	if event.Attachments[0].MediaRef != BuildInboundMediaRef(PlatformSlack, "image", "F_1") {
		t.Fatalf("attachment media ref = %q, want slack image ref", event.Attachments[0].MediaRef)
	}
	if event.Attachments[0].URL != "https://files.example.com/f1" {
		t.Fatalf("attachment URL = %q, want https://files.example.com/f1", event.Attachments[0].URL)
	}
	if event.ReplyHandle["channel"] != "C_1" {
		t.Fatalf("ReplyHandle channel = %v, want C_1", event.ReplyHandle["channel"])
	}
	if len(event.RawPayload) == 0 {
		t.Fatal("RawPayload is empty")
	}
}

func TestProcessSlackRuntimePayload(t *testing.T) {
	t.Parallel()

	adapter := NewSlackAdapter()
	adapter.identityCache["conn_1"] = slackBotIdentity{UserID: "U_BOT"}

	var captured []InboundEvent
	adapter.inboundProcessor = func(ctx context.Context, conn Connection, events []InboundEvent) (inboundPersistResult, error) {
		captured = append(captured, events...)
		return inboundPersistResult{}, nil
	}

	conn := Connection{
		ID:             "conn_1",
		Platform:       PlatformSlack,
		ConnectionMode: SlackConnectionMode,
		Secrets: map[string]any{
			"botToken": "xoxb-bot",
			"appToken": "xapp-app",
		},
	}

	payload := []byte(`{
		"token":"verify-token",
		"team_id":"T_1",
		"api_app_id":"A_1",
		"type":"event_callback",
		"event":{
			"type":"message",
			"user":"U_1",
			"text":"hello <@U_BOT>",
			"ts":"171.02",
			"channel":"D_1",
			"channel_type":"im",
			"event_ts":"171.02"
		},
		"event_id":"Ev_2",
		"event_time":171
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
	if !captured[0].MentionsBot {
		t.Fatal("MentionsBot = false, want true")
	}
}

func TestSlackTestConnectionAndSendOutbound(t *testing.T) {
	t.Parallel()

	var requestBodies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		requestBodies = append(requestBodies, string(body))
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/auth.test":
			_, _ = w.Write([]byte(`{"ok":true,"team_id":"T_1","user_id":"U_BOT","bot_id":"B_1"}`))
		case "/api/chat.postMessage":
			if !strings.Contains(string(body), "channel=C_1") {
				t.Fatalf("chat.postMessage body missing channel: %s", string(body))
			}
			if !strings.Contains(string(body), "thread_ts=171.00") {
				t.Fatalf("chat.postMessage body missing thread_ts: %s", string(body))
			}
			if !strings.Contains(string(body), "text=reply+body") {
				t.Fatalf("chat.postMessage body missing text: %s", string(body))
			}
			_, _ = w.Write([]byte(`{"ok":true,"channel":"C_1","ts":"171.03","message":{"text":"reply body"}}`))
		default:
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	conn := Connection{
		ID:             "conn_1",
		Platform:       PlatformSlack,
		ConnectionMode: SlackConnectionMode,
		Config: map[string]any{
			"apiBaseURL": server.URL + "/api/",
		},
		Secrets: map[string]any{
			"botToken": "xoxb-bot",
			"appToken": "xapp-app",
		},
	}

	adapter := NewSlackAdapter()
	if err := adapter.TestConnection(context.Background(), conn); err != nil {
		t.Fatalf("TestConnection() error = %v", err)
	}
	if cached, ok := adapter.identityCache[conn.ID]; !ok || cached.UserID != "U_BOT" {
		t.Fatalf("identity cache = %+v, want user U_BOT", cached)
	}

	err := adapter.SendOutbound(context.Background(), conn, OutboundEnvelope{
		ExternalChatID:   "C_1",
		ExternalThreadID: "171.00",
		Text:             "reply body",
		ReplyHandle: map[string]any{
			"channel":   "C_1",
			"thread_ts": "171.00",
		},
	})
	if err != nil {
		t.Fatalf("SendOutbound() error = %v", err)
	}
	if len(requestBodies) != 2 {
		t.Fatalf("request count = %d, want 2", len(requestBodies))
	}
}
