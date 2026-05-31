package im

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestQQValidateConfig(t *testing.T) {
	t.Parallel()

	adapter := NewQQAdapter()
	tests := []struct {
		name    string
		config  map[string]any
		secrets map[string]any
		wantErr error
	}{
		{
			name: "valid gateway config",
			config: map[string]any{
				"appId":          "123456",
				"connectionMode": QQConnectionMode,
			},
			secrets: map[string]any{
				"appSecret": "secret",
			},
		},
		{
			name:    "missing app id",
			config:  map[string]any{},
			secrets: map[string]any{"appSecret": "secret"},
			wantErr: ErrInvalidConnectionConfig,
		},
		{
			name: "unsupported mode",
			config: map[string]any{
				"appId":          "123456",
				"connectionMode": "webhook",
			},
			secrets: map[string]any{"appSecret": "secret"},
			wantErr: ErrQQUnsupportedConnectionMode,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := adapter.ValidateConfig(tt.config, tt.secrets)
			if err != tt.wantErr {
				t.Fatalf("ValidateConfig() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseQQInbound(t *testing.T) {
	t.Parallel()

	raw := []byte(`{
		"id":"evt_1",
		"op":0,
		"s":1,
		"t":"GROUP_AT_MESSAGE_CREATE",
		"d":{
			"id":"msg_1",
			"content":"hello qq",
			"group_openid":"group_1",
			"timestamp":"2023-11-06T13:37:18+08:00",
			"author":{"member_openid":"member_1"},
			"attachments":[{"url":"https://qq.example.com/file.png","filename":"file.png","content_type":"image/png"}]
		}
	}`)

	events, err := parseQQInbound(raw)
	if err != nil {
		t.Fatalf("parseQQInbound() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("event count = %d, want 1", len(events))
	}
	event := events[0]
	if event.Platform != PlatformQQ {
		t.Fatalf("Platform = %q, want %q", event.Platform, PlatformQQ)
	}
	if event.ChatType != "group" {
		t.Fatalf("ChatType = %q, want group", event.ChatType)
	}
	if event.ExternalChatID != "group_1" {
		t.Fatalf("ExternalChatID = %q, want group_1", event.ExternalChatID)
	}
	if event.ExternalUserID != "member_1" {
		t.Fatalf("ExternalUserID = %q, want member_1", event.ExternalUserID)
	}
	if !event.MentionsBot {
		t.Fatal("MentionsBot = false, want true")
	}
	if len(event.Attachments) != 1 {
		t.Fatalf("attachment count = %d, want 1", len(event.Attachments))
	}
	if event.Attachments[0].MediaRef == "" {
		t.Fatal("attachment media ref is empty")
	}
	if event.ReplyHandle["peer_type"] != "group" {
		t.Fatalf("ReplyHandle peer_type = %v, want group", event.ReplyHandle["peer_type"])
	}
}

func TestParseQQInboundC2C(t *testing.T) {
	t.Parallel()

	events, err := parseQQInbound([]byte(`{
		"id":"evt_2",
		"op":0,
		"t":"C2C_MESSAGE_CREATE",
		"d":{
			"id":"msg_2",
			"content":"hello dm",
			"timestamp":"2023-11-06T13:37:18+08:00",
			"author":{"user_openid":"user_1"}
		}
	}`))
	if err != nil {
		t.Fatalf("parseQQInbound() error = %v", err)
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

func TestQQProcessRuntimePayload(t *testing.T) {
	t.Parallel()

	adapter := NewQQAdapter()
	var captured []InboundEvent
	adapter.inboundProcessor = func(ctx context.Context, conn Connection, events []InboundEvent) (inboundPersistResult, error) {
		captured = append(captured, events...)
		return inboundPersistResult{}, nil
	}

	conn := Connection{
		ID:             "conn_1",
		Platform:       PlatformQQ,
		ConnectionMode: QQConnectionMode,
		Config: map[string]any{
			"appId": "123456",
		},
		Secrets: map[string]any{
			"appSecret": "secret",
		},
	}

	payload := []byte(`{
		"id":"evt_2",
		"op":0,
		"t":"C2C_MESSAGE_CREATE",
		"d":{
			"id":"msg_2",
			"content":"hello dm",
			"timestamp":"2023-11-06T13:37:18+08:00",
			"author":{"user_openid":"user_1"}
		}
	}`)

	if err := adapter.processRuntimePayload(context.Background(), conn, payload); err != nil {
		t.Fatalf("processRuntimePayload() error = %v", err)
	}
	if len(captured) != 1 {
		t.Fatalf("captured count = %d, want 1", len(captured))
	}
}

func TestQQTestConnectionAndSendOutbound(t *testing.T) {
	t.Parallel()

	var tokenCalls int
	var authHeaders []string
	var paths []string
	var bodies []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		authHeaders = append(authHeaders, r.Header.Get("Authorization"))
		paths = append(paths, r.URL.Path)
		bodies = append(bodies, string(body))
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/token":
			tokenCalls++
			_, _ = w.Write([]byte(`{"access_token":"qq_access_token","expires_in":7200}`))
		case "/gateway/bot":
			_, _ = w.Write([]byte(`{"url":"wss://qq.example.com/gateway"}`))
		case "/v2/users/user_1/messages":
			_, _ = w.Write([]byte(`{"id":"reply_1","timestamp":1710000000}`))
		case "/v2/groups/group_1/messages":
			_, _ = w.Write([]byte(`{"id":"reply_2","timestamp":1710000001}`))
		default:
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	conn := Connection{
		ID:             "conn_1",
		Platform:       PlatformQQ,
		ConnectionMode: QQConnectionMode,
		Config: map[string]any{
			"appId":      "123456",
			"tokenURL":   server.URL + "/token",
			"apiBaseURL": server.URL,
		},
		Secrets: map[string]any{
			"appSecret": "secret",
		},
	}

	adapter := NewQQAdapter()
	if err := adapter.TestConnection(context.Background(), conn); err != nil {
		t.Fatalf("TestConnection() error = %v", err)
	}
	if tokenCalls != 1 {
		t.Fatalf("tokenCalls = %d, want 1", tokenCalls)
	}
	if len(authHeaders) < 2 || authHeaders[1] != "QQBot qq_access_token" {
		t.Fatalf("auth headers = %v, want QQBot qq_access_token", authHeaders)
	}

	err := adapter.SendOutbound(context.Background(), conn, OutboundEnvelope{
		ExternalChatID: "user_1",
		Text:           "reply dm",
		ReplyHandle: map[string]any{
			"user_openid": "user_1",
			"message_id":  "msg_1",
			"peer_type":   "c2c",
		},
	})
	if err != nil {
		t.Fatalf("SendOutbound() error = %v", err)
	}

	err = adapter.SendOutbound(context.Background(), conn, OutboundEnvelope{
		ExternalChatID: "group_1",
		Text:           "reply group",
		ReplyHandle: map[string]any{
			"group_openid": "group_1",
			"message_id":   "msg_2",
			"peer_type":    "group",
		},
	})
	if err != nil {
		t.Fatalf("SendOutbound() group error = %v", err)
	}

	if !containsString(paths, "/v2/users/user_1/messages") {
		t.Fatalf("paths = %v, want c2c endpoint", paths)
	}
	if !containsString(paths, "/v2/groups/group_1/messages") {
		t.Fatalf("paths = %v, want group endpoint", paths)
	}
	if !containsSubstring(bodies, `"msg_id":"msg_1"`) {
		t.Fatalf("bodies = %v, want msg_id for c2c", bodies)
	}
	if !containsSubstring(bodies, `"content":"reply group"`) {
		t.Fatalf("bodies = %v, want group content", bodies)
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func containsSubstring(values []string, target string) bool {
	for _, value := range values {
		if strings.Contains(value, target) {
			return true
		}
	}
	return false
}
