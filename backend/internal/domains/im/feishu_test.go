package im

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestFeishuValidateConfig(t *testing.T) {
	t.Parallel()

	adapter := NewFeishuAdapter()

	tests := []struct {
		name    string
		config  map[string]any
		secrets map[string]any
		wantErr error
	}{
		{
			name: "valid socket mode config",
			config: map[string]any{
				"appId":          "cli_xxx",
				"connectionMode": "socket_mode",
				"domain":         "feishu",
			},
			secrets: map[string]any{
				"appSecret": "secret",
			},
		},
		{
			name: "missing app id",
			config: map[string]any{
				"connectionMode": "socket_mode",
			},
			secrets: map[string]any{
				"appSecret": "secret",
			},
			wantErr: ErrInvalidConnectionConfig,
		},
		{
			name: "missing app secret",
			config: map[string]any{
				"appId":          "cli_xxx",
				"connectionMode": "socket_mode",
			},
			secrets: map[string]any{},
			wantErr: ErrInvalidConnectionConfig,
		},
		{
			name: "unsupported connection mode",
			config: map[string]any{
				"appId":          "cli_xxx",
				"connectionMode": "webhook",
			},
			secrets: map[string]any{
				"appSecret": "secret",
			},
			wantErr: ErrFeishuUnsupportedConnectionMode,
		},
		{
			name: "invalid domain",
			config: map[string]any{
				"appId":          "cli_xxx",
				"connectionMode": "socket_mode",
				"domain":         "example",
			},
			secrets: map[string]any{
				"appSecret": "secret",
			},
			wantErr: ErrInvalidConnectionConfig,
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

func TestFeishuParseInbound(t *testing.T) {
	t.Parallel()

	adapter := NewFeishuAdapter()
	payload := `{
		"header": {
			"event_id": "evt_123"
		},
		"event": {
			"sender": {
				"sender_id": {
					"open_id": "ou_xxx"
				}
			},
			"message": {
				"chat_id": "oc_chat",
				"message_id": "om_msg",
				"root_id": "om_root",
				"chat_type": "group",
				"content": "{\"text\":\"hello from feishu\"}"
			},
			"mentions": [
				{
					"name": "bot"
				}
			]
		}
	}`

	events, err := adapter.ParseInbound(context.Background(), payload)
	if err != nil {
		t.Fatalf("ParseInbound() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("ParseInbound() event count = %d, want 1", len(events))
	}

	event := events[0]
	if event.Platform != "feishu" {
		t.Fatalf("Platform = %q, want feishu", event.Platform)
	}
	if event.EventID != "evt_123" {
		t.Fatalf("EventID = %q, want evt_123", event.EventID)
	}
	if event.MessageID != "om_msg" {
		t.Fatalf("MessageID = %q, want om_msg", event.MessageID)
	}
	if event.ExternalChatID != "oc_chat" {
		t.Fatalf("ExternalChatID = %q, want oc_chat", event.ExternalChatID)
	}
	if event.ExternalThreadID != "om_root" {
		t.Fatalf("ExternalThreadID = %q, want om_root", event.ExternalThreadID)
	}
	if event.ExternalUserID != "ou_xxx" {
		t.Fatalf("ExternalUserID = %q, want ou_xxx", event.ExternalUserID)
	}
	if event.ChatType != "group" {
		t.Fatalf("ChatType = %q, want group", event.ChatType)
	}
	if !event.MentionsBot {
		t.Fatal("MentionsBot = false, want true")
	}
	if event.Text != "hello from feishu" {
		t.Fatalf("Text = %q, want parsed feishu text", event.Text)
	}
	if event.ReplyHandle["message_id"] != "om_msg" {
		t.Fatalf("ReplyHandle message_id = %v, want om_msg", event.ReplyHandle["message_id"])
	}
	if len(event.RawPayload) == 0 {
		t.Fatal("RawPayload is empty")
	}
	if event.ReceivedAt.IsZero() {
		t.Fatal("ReceivedAt is zero")
	}
}

func TestFeishuParseInboundFallbacks(t *testing.T) {
	t.Parallel()

	adapter := NewFeishuAdapter()
	payload := map[string]any{
		"header": map[string]any{},
		"event": map[string]any{
			"text": "fallback text",
			"sender": map[string]any{
				"sender_id": map[string]any{
					"user_id": "user_xxx",
				},
			},
			"message": map[string]any{
				"chat_id":    "oc_chat",
				"message_id": "om_msg",
				"thread_id":  "om_thread",
				"chat_type":  "p2p",
			},
		},
	}

	events, err := adapter.ParseInbound(context.Background(), payload)
	if err != nil {
		t.Fatalf("ParseInbound() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("ParseInbound() event count = %d, want 1", len(events))
	}

	event := events[0]
	if event.EventID != "om_msg" {
		t.Fatalf("EventID = %q, want message_id fallback", event.EventID)
	}
	if event.Text != "fallback text" {
		t.Fatalf("Text = %q, want fallback text", event.Text)
	}
	if event.ExternalThreadID != "om_thread" {
		t.Fatalf("ExternalThreadID = %q, want om_thread", event.ExternalThreadID)
	}
	if event.ExternalUserID != "user_xxx" {
		t.Fatalf("ExternalUserID = %q, want user_xxx", event.ExternalUserID)
	}
	if event.MentionsBot {
		t.Fatal("MentionsBot = true, want false")
	}
}

func TestFeishuParseInboundInvalidPayload(t *testing.T) {
	t.Parallel()

	adapter := NewFeishuAdapter()

	tests := []struct {
		name string
		raw  any
	}{
		{
			name: "invalid type",
			raw:  123,
		},
		{
			name: "missing chat id",
			raw: map[string]any{
				"header": map[string]any{
					"event_id": "evt_123",
				},
				"event": map[string]any{
					"message": map[string]any{
						"message_id": "om_msg",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := adapter.ParseInbound(context.Background(), tt.raw)
			if !errors.Is(err, ErrFeishuInvalidPayload) {
				t.Fatalf("ParseInbound() error = %v, want %v", err, ErrFeishuInvalidPayload)
			}
		})
	}
}

func TestFeishuTestConnection(t *testing.T) {
	t.Parallel()

	adapter := NewFeishuAdapter()
	var tokenCalls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/open-apis/auth/v3/tenant_access_token/internal" {
			t.Fatalf("unexpected request path: %s", r.URL.String())
		}
		tokenCalls.Add(1)
		var req map[string]any
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req["app_id"] != "cli_xxx" || req["app_secret"] != "secret" {
			t.Fatalf("token request body = %+v, want app credentials", req)
		}
		_, _ = w.Write([]byte(`{"code":0,"msg":"success","tenant_access_token":"tenant-token","expire":7200}`))
	}))
	defer server.Close()

	conn := Connection{
		ID:             "conn_1",
		Platform:       "feishu",
		ConnectionMode: "socket_mode",
		Config: map[string]any{
			"appId":      "cli_xxx",
			"apiBaseURL": server.URL,
		},
		Secrets: map[string]any{
			"appSecret": "secret",
		},
	}

	if err := adapter.TestConnection(context.Background(), conn); err != nil {
		t.Fatalf("TestConnection() error = %v", err)
	}
	if got := tokenCalls.Load(); got != 1 {
		t.Fatalf("token call count = %d, want 1", got)
	}
	if cached, ok := adapter.tokenCache[conn.ID]; !ok || cached.Token != "tenant-token" {
		t.Fatalf("token cache = %+v, want tenant-token", cached)
	}
}

func TestFeishuStartStopRuntime(t *testing.T) {
	t.Parallel()

	adapter := NewFeishuAdapter()
	var tokenCalls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/open-apis/auth/v3/tenant_access_token/internal" {
			t.Fatalf("unexpected request path: %s", r.URL.String())
		}
		tokenCalls.Add(1)
		_, _ = w.Write([]byte(`{"code":0,"msg":"success","tenant_access_token":"tenant-token","expire":7200}`))
	}))
	defer server.Close()

	conn := Connection{
		ID:             "conn_1",
		Platform:       "feishu",
		ConnectionMode: "socket_mode",
		Config: map[string]any{
			"appId":      "cli_xxx",
			"apiBaseURL": server.URL,
		},
		Secrets: map[string]any{
			"appSecret": "secret",
		},
	}

	if err := adapter.Start(context.Background(), conn); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defaultRuntimeManager.mu.RLock()
	runtime, ok := defaultRuntimeManager.runtimes[conn.ID]
	defaultRuntimeManager.mu.RUnlock()
	if !ok {
		t.Fatal("runtime missing after Start()")
	}
	if runtime.Connection.ID != conn.ID {
		t.Fatalf("runtime connection id = %q, want %q", runtime.Connection.ID, conn.ID)
	}
	if runtime.StartedAt.IsZero() {
		t.Fatal("runtime StartedAt is zero")
	}
	if got := tokenCalls.Load(); got != 1 {
		t.Fatalf("token call count after Start = %d, want 1", got)
	}

	if err := adapter.Stop(context.Background(), conn.ID); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	defaultRuntimeManager.mu.RLock()
	_, ok = defaultRuntimeManager.runtimes[conn.ID]
	defaultRuntimeManager.mu.RUnlock()
	if ok {
		t.Fatal("runtime still exists after Stop()")
	}
	if _, ok := adapter.tokenCache[conn.ID]; ok {
		t.Fatal("token cache still exists after Stop()")
	}
}

func TestFeishuStartFailsWhenConnectivityCheckFails(t *testing.T) {
	t.Parallel()

	adapter := NewFeishuAdapter()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":999,"msg":"invalid app"}`))
	}))
	defer server.Close()

	conn := Connection{
		ID:             "conn_1",
		Platform:       "feishu",
		ConnectionMode: "socket_mode",
		Config: map[string]any{
			"appId":      "cli_xxx",
			"apiBaseURL": server.URL,
		},
		Secrets: map[string]any{
			"appSecret": "secret",
		},
	}

	err := adapter.Start(context.Background(), conn)
	if err == nil {
		t.Fatal("Start() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "invalid app") {
		t.Fatalf("Start() error = %v, want invalid app", err)
	}
	defaultRuntimeManager.mu.RLock()
	_, ok := defaultRuntimeManager.runtimes[conn.ID]
	defaultRuntimeManager.mu.RUnlock()
	if ok {
		t.Fatal("runtime exists after failed Start()")
	}
}

func TestCoerceRawMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     any
		want    map[string]any
		wantErr bool
	}{
		{
			name: "map input",
			raw: map[string]any{
				"hello": "world",
			},
			want: map[string]any{
				"hello": "world",
			},
		},
		{
			name: "json bytes",
			raw:  []byte(`{"hello":"world"}`),
			want: map[string]any{
				"hello": "world",
			},
		},
		{
			name: "json string",
			raw:  `{"hello":"world"}`,
			want: map[string]any{
				"hello": "world",
			},
		},
		{
			name:    "invalid json",
			raw:     `{"hello":`,
			wantErr: true,
		},
		{
			name:    "unsupported type",
			raw:     1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := coerceRawMap(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatal("coerceRawMap() error = nil, want non-nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("coerceRawMap() error = %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("coerceRawMap() len = %d, want %d", len(got), len(tt.want))
			}
			for key, wantValue := range tt.want {
				if got[key] != wantValue {
					t.Fatalf("coerceRawMap()[%q] = %v, want %v", key, got[key], wantValue)
				}
			}
		})
	}
}

func TestExtractFeishuText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "json text payload",
			content: `{"text":"hello"}`,
			want:    "hello",
		},
		{
			name:    "plain text payload",
			content: "hello",
			want:    "hello",
		},
		{
			name:    "empty payload",
			content: "   ",
			want:    "",
		},
		{
			name:    "json without text field",
			content: `{"msg":"hello"}`,
			want:    `{"msg":"hello"}`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := extractFeishuText(tt.content); got != tt.want {
				t.Fatalf("extractFeishuText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatFeishuOutboundText(t *testing.T) {
	t.Parallel()

	got := formatFeishuOutboundText(OutboundEnvelope{
		Segments: []RichSegment{
			{Type: "text", Text: "hello "},
			{Type: "mention", Text: "@bot"},
			{Type: "text", Text: "\nworld"},
		},
		Attachments: []OutboundAttachment{
			{Name: "report.pdf", Type: "file"},
		},
	})
	if got != "hello @bot\nworld\n[report.pdf]" {
		t.Fatalf("formatFeishuOutboundText() = %q, want flattened text with attachment label", got)
	}
}

func TestBuildFeishuOutboundMessagePost(t *testing.T) {
	t.Parallel()

	outbound, err := buildFeishuOutboundMessage(OutboundEnvelope{
		Segments: []RichSegment{
			{Type: "text", Text: "hello "},
			{Type: "mention", Text: "bot", Meta: map[string]any{"user_id": "ou_xxx"}},
			{Type: "text", Text: "\n"},
			{Type: "link", Text: "dotblue", Meta: map[string]any{"url": "https://dotblue.ai"}},
		},
	})
	if err != nil {
		t.Fatalf("buildFeishuOutboundMessage() error = %v", err)
	}
	if outbound.MsgType != "post" {
		t.Fatalf("MsgType = %q, want post", outbound.MsgType)
	}

	var content map[string]any
	if err := json.Unmarshal([]byte(outbound.Content), &content); err != nil {
		t.Fatalf("post content unmarshal error = %v", err)
	}
	post := content["post"].(map[string]any)
	zhCN := post["zh_cn"].(map[string]any)
	rows := zhCN["content"].([]any)
	if len(rows) != 2 {
		t.Fatalf("post row count = %d, want 2", len(rows))
	}
	firstRow := rows[0].([]any)
	secondRow := rows[1].([]any)
	if firstRow[1].(map[string]any)["tag"] != "at" {
		t.Fatalf("first row second element tag = %v, want at", firstRow[1].(map[string]any)["tag"])
	}
	if secondRow[0].(map[string]any)["tag"] != "a" {
		t.Fatalf("second row first element tag = %v, want a", secondRow[0].(map[string]any)["tag"])
	}
}

func TestBuildFeishuOutboundRequestsWithAttachments(t *testing.T) {
	t.Parallel()

	outbounds, err := buildFeishuOutboundRequests(OutboundEnvelope{
		Text: "hello",
		Attachments: []OutboundAttachment{
			{
				Type:     "image",
				MediaRef: BuildInboundMediaRef("feishu", "image", "img_key_1"),
			},
		},
	})
	if err != nil {
		t.Fatalf("buildFeishuOutboundRequests() error = %v", err)
	}
	if len(outbounds) != 2 {
		t.Fatalf("outbound request count = %d, want 2", len(outbounds))
	}
	if outbounds[0].MsgType != "text" {
		t.Fatalf("first outbound msg_type = %q, want text", outbounds[0].MsgType)
	}
	if strings.Contains(outbounds[0].Content, "attachment") || strings.Contains(outbounds[0].Content, "img_key_1") {
		t.Fatalf("first outbound content = %q, want pure text payload", outbounds[0].Content)
	}
	if outbounds[1].MsgType != "image" {
		t.Fatalf("second outbound msg_type = %q, want image", outbounds[1].MsgType)
	}
}

func TestBuildFeishuOutboundRequestsAttachmentsOnly(t *testing.T) {
	t.Parallel()

	outbounds, err := buildFeishuOutboundRequests(OutboundEnvelope{
		Segments: []RichSegment{
			{Type: "image", Meta: map[string]any{"media_ref": BuildInboundMediaRef("feishu", "image", "img_key_1")}},
		},
		Attachments: []OutboundAttachment{
			{
				Type:     "image",
				MediaRef: BuildInboundMediaRef("feishu", "image", "img_key_1"),
			},
		},
	})
	if err != nil {
		t.Fatalf("buildFeishuOutboundRequests() error = %v", err)
	}
	if len(outbounds) != 1 {
		t.Fatalf("outbound request count = %d, want 1", len(outbounds))
	}
	if outbounds[0].MsgType != "image" {
		t.Fatalf("only outbound msg_type = %q, want image", outbounds[0].MsgType)
	}
}

func TestFeishuSendOutboundIntegration(t *testing.T) {
	adapter := NewFeishuAdapter()

	var tokenCalls atomic.Int32
	var createCalls atomic.Int32
	var replyCalls atomic.Int32
	var postCalls atomic.Int32
	var imageCalls atomic.Int32
	var fileCalls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/open-apis/auth/v3/tenant_access_token/internal":
			tokenCalls.Add(1)
			var req map[string]any
			_ = json.NewDecoder(r.Body).Decode(&req)
			if req["app_id"] != "cli_xxx" || req["app_secret"] != "secret" {
				t.Fatalf("token request body = %+v, want app credentials", req)
			}
			_, _ = w.Write([]byte(`{"code":0,"msg":"success","tenant_access_token":"tenant-token","expire":7200}`))
		case strings.HasPrefix(r.URL.Path, "/open-apis/im/v1/messages/") && strings.HasSuffix(r.URL.Path, "/reply"):
			replyCalls.Add(1)
			if got := r.Header.Get("Authorization"); got != "Bearer tenant-token" {
				t.Fatalf("reply authorization = %q, want Bearer tenant-token", got)
			}
			var req map[string]any
			_ = json.NewDecoder(r.Body).Decode(&req)
			if req["msg_type"] != "text" {
				t.Fatalf("reply msg_type = %v, want text", req["msg_type"])
			}
			if !strings.Contains(str(req["content"]), "reply content") {
				t.Fatalf("reply content = %v, want reply content", req["content"])
			}
			_, _ = w.Write([]byte(`{"code":0,"msg":"success","data":{"message_id":"om_reply"}}`))
		case r.URL.Path == "/open-apis/im/v1/messages":
			createCalls.Add(1)
			if got := r.URL.Query().Get("receive_id_type"); got != "chat_id" {
				t.Fatalf("receive_id_type = %q, want chat_id", got)
			}
			if got := r.Header.Get("Authorization"); got != "Bearer tenant-token" {
				t.Fatalf("create authorization = %q, want Bearer tenant-token", got)
			}
			var req map[string]any
			_ = json.NewDecoder(r.Body).Decode(&req)
			if req["receive_id"] != "oc_chat" {
				t.Fatalf("receive_id = %v, want oc_chat", req["receive_id"])
			}
			switch req["msg_type"] {
			case "text":
				if !strings.Contains(str(req["content"]), "create content") {
					t.Fatalf("create content = %v, want create content", req["content"])
				}
			case "post":
				postCalls.Add(1)
				var postContent map[string]any
				if err := json.Unmarshal([]byte(str(req["content"])), &postContent); err != nil {
					t.Fatalf("post content unmarshal error = %v", err)
				}
				if _, ok := postContent["post"]; !ok {
					t.Fatalf("post content = %+v, want post root", postContent)
				}
			case "image":
				imageCalls.Add(1)
				if !strings.Contains(str(req["content"]), "img_key_1") {
					t.Fatalf("image content = %v, want img_key_1", req["content"])
				}
			case "file":
				fileCalls.Add(1)
				if !strings.Contains(str(req["content"]), "file_key_1") {
					t.Fatalf("file content = %v, want file_key_1", req["content"])
				}
			default:
				t.Fatalf("unexpected create msg_type = %v", req["msg_type"])
			}
			_, _ = w.Write([]byte(`{"code":0,"msg":"success","data":{"message_id":"om_create"}}`))
		default:
			t.Fatalf("unexpected request path: %s", r.URL.String())
		}
	}))
	defer server.Close()

	conn := Connection{
		ID:           "conn_1",
		EnterpriseID: "ent_1",
		Platform:     "feishu",
		Config: map[string]any{
			"appId":          "cli_xxx",
			"connectionMode": "socket_mode",
			"apiBaseURL":     server.URL,
		},
		Secrets: map[string]any{
			"appSecret": "secret",
		},
	}

	if err := adapter.SendOutbound(context.Background(), conn, OutboundEnvelope{
		ExternalChatID: "oc_chat",
		Text:           "create content",
	}); err != nil {
		t.Fatalf("SendOutbound(create) error = %v", err)
	}

	if err := adapter.SendOutbound(context.Background(), conn, OutboundEnvelope{
		ExternalChatID: "oc_chat",
		Text:           "reply content",
		ReplyHandle: map[string]any{
			"message_id": "om_source",
		},
	}); err != nil {
		t.Fatalf("SendOutbound(reply) error = %v", err)
	}

	if got := tokenCalls.Load(); got != 1 {
		t.Fatalf("token call count = %d, want 1 due to cache", got)
	}
	if got := createCalls.Load(); got != 1 {
		t.Fatalf("create call count = %d, want 1", got)
	}
	if got := replyCalls.Load(); got != 1 {
		t.Fatalf("reply call count = %d, want 1", got)
	}

	if err := adapter.SendOutbound(context.Background(), conn, OutboundEnvelope{
		ExternalChatID: "oc_chat",
		Segments: []RichSegment{
			{Type: "text", Text: "hello "},
			{Type: "mention", Text: "bot", Meta: map[string]any{"user_id": "ou_xxx"}},
			{Type: "text", Text: "\n"},
			{Type: "link", Text: "dotblue", Meta: map[string]any{"url": "https://dotblue.ai"}},
		},
	}); err != nil {
		t.Fatalf("SendOutbound(post) error = %v", err)
	}
	if got := postCalls.Load(); got != 1 {
		t.Fatalf("post call count = %d, want 1", got)
	}

	if err := adapter.SendOutbound(context.Background(), conn, OutboundEnvelope{
		ExternalChatID: "oc_chat",
		Attachments: []OutboundAttachment{
			{
				Type:     "image",
				MediaRef: BuildInboundMediaRef("feishu", "image", "img_key_1"),
			},
			{
				Type:     "file",
				Name:     "report.pdf",
				MediaRef: BuildInboundMediaRef("feishu", "file", "file_key_1"),
			},
		},
	}); err != nil {
		t.Fatalf("SendOutbound(attachments) error = %v", err)
	}
	if got := imageCalls.Load(); got != 1 {
		t.Fatalf("image call count = %d, want 1", got)
	}
	if got := fileCalls.Load(); got != 1 {
		t.Fatalf("file call count = %d, want 1", got)
	}
}
