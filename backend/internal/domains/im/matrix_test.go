package im

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMatrixValidateConfig(t *testing.T) {
	t.Parallel()

	adapter := NewMatrixAdapter()
	tests := []struct {
		name    string
		config  map[string]any
		secrets map[string]any
		wantErr error
	}{
		{
			name: "valid matrix config",
			config: map[string]any{
				"homeserver":     "https://matrix.example.com",
				"userId":         "@bot:example.com",
				"connectionMode": MatrixConnectionModeSync,
			},
			secrets: map[string]any{
				"accessToken": "matrix-access-token",
			},
		},
		{
			name: "missing homeserver",
			config: map[string]any{
				"userId": "@bot:example.com",
			},
			secrets: map[string]any{
				"accessToken": "matrix-access-token",
			},
			wantErr: ErrInvalidConnectionConfig,
		},
		{
			name: "missing token",
			config: map[string]any{
				"homeserver": "https://matrix.example.com",
				"userId":     "@bot:example.com",
			},
			secrets: map[string]any{},
			wantErr: ErrInvalidConnectionConfig,
		},
		{
			name: "unsupported mode",
			config: map[string]any{
				"homeserver":     "https://matrix.example.com",
				"userId":         "@bot:example.com",
				"connectionMode": "gateway",
			},
			secrets: map[string]any{
				"accessToken": "matrix-access-token",
			},
			wantErr: ErrMatrixUnsupportedConnectionMode,
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

func TestParseMatrixInbound(t *testing.T) {
	t.Parallel()

	raw := []byte(`{
		"next_batch":"s123",
		"rooms":{
			"join":{
				"!room:example.com":{
					"timeline":{
						"events":[
							{
								"type":"m.room.message",
								"event_id":"$event1",
								"sender":"@alice:example.com",
								"origin_server_ts":1710000000000,
								"content":{
									"msgtype":"m.text",
									"body":"hello matrix",
									"m.mentions":{"user_ids":["@bot:example.com"]}
								}
							},
							{
								"type":"m.room.message",
								"event_id":"$event2",
								"sender":"@bob:example.com",
								"origin_server_ts":1710000001000,
								"content":{
									"msgtype":"m.image",
									"body":"diagram.png",
									"url":"mxc://example.com/file",
									"info":{"mimetype":"image/png"}
								}
							}
						]
					}
				}
			}
		}
	}`)

	events, err := parseMatrixInbound(raw, "@bot:example.com")
	if err != nil {
		t.Fatalf("parseMatrixInbound() error = %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("event count = %d, want 2", len(events))
	}
	if events[0].Platform != PlatformMatrix {
		t.Fatalf("Platform = %q, want %q", events[0].Platform, PlatformMatrix)
	}
	if events[0].ExternalChatID != "!room:example.com" {
		t.Fatalf("ExternalChatID = %q, want !room:example.com", events[0].ExternalChatID)
	}
	if !events[0].MentionsBot {
		t.Fatal("MentionsBot = false, want true")
	}
	if events[1].Attachments[0].MediaRef != BuildInboundMediaRef(PlatformMatrix, "image", "$event2") {
		t.Fatalf("attachment media ref = %q, want matrix image ref", events[1].Attachments[0].MediaRef)
	}
	if events[1].Attachments[0].ContentType != "image/png" {
		t.Fatalf("attachment content type = %q, want image/png", events[1].Attachments[0].ContentType)
	}
}

func TestParseMatrixInboundSkipsBotEvents(t *testing.T) {
	t.Parallel()

	events, err := parseMatrixInbound([]byte(`{
		"rooms":{"join":{"!room:example.com":{"timeline":{"events":[
			{"type":"m.room.message","event_id":"$event1","sender":"@bot:example.com","content":{"msgtype":"m.text","body":"hello"}}
		]}}}}
	}`), "@bot:example.com")
	if err != nil {
		t.Fatalf("parseMatrixInbound() error = %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("event count = %d, want 0", len(events))
	}
}

func TestMatrixProcessRuntimePayload(t *testing.T) {
	t.Parallel()

	adapter := NewMatrixAdapter()
	adapter.identityCache["conn_1"] = matrixIdentity{UserID: "@bot:example.com"}

	var captured []InboundEvent
	adapter.inboundProcessor = func(ctx context.Context, conn Connection, events []InboundEvent) (inboundPersistResult, error) {
		captured = append(captured, events...)
		return inboundPersistResult{}, nil
	}

	conn := Connection{
		ID:             "conn_1",
		Platform:       PlatformMatrix,
		ConnectionMode: MatrixConnectionModeSync,
		Config: map[string]any{
			"homeserver": "https://matrix.example.com",
			"userId":     "@bot:example.com",
		},
		Secrets: map[string]any{
			"accessToken": "matrix-access-token",
		},
	}

	payload := []byte(`{
		"rooms":{"join":{"!room:example.com":{"timeline":{"events":[
			{"type":"m.room.message","event_id":"$event1","sender":"@alice:example.com","content":{"msgtype":"m.text","body":"hello"}}
		]}}}}
	}`)

	if err := adapter.processRuntimePayload(context.Background(), conn, payload); err != nil {
		t.Fatalf("processRuntimePayload() error = %v", err)
	}
	if len(captured) != 1 {
		t.Fatalf("captured count = %d, want 1", len(captured))
	}
}

func TestMatrixTestConnectionAndSendOutbound(t *testing.T) {
	t.Parallel()

	var authHeader string
	var sentPath string
	var sentBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.URL.Path == "/_matrix/client/v3/account/whoami":
			_, _ = w.Write([]byte(`{"user_id":"@bot:example.com"}`))
		case strings.Contains(r.URL.Path, "/_matrix/client/v3/rooms/") && strings.Contains(r.URL.Path, "/send/m.room.message/"):
			sentPath = r.URL.Path
			sentBody = string(body)
			_, _ = w.Write([]byte(`{"event_id":"$reply"}`))
		default:
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	conn := Connection{
		ID:             "conn_1",
		Platform:       PlatformMatrix,
		ConnectionMode: MatrixConnectionModeSync,
		Config: map[string]any{
			"homeserver": server.URL,
			"userId":     "@bot:example.com",
		},
		Secrets: map[string]any{
			"accessToken": "matrix-access-token",
		},
	}

	adapter := NewMatrixAdapter()
	if err := adapter.TestConnection(context.Background(), conn); err != nil {
		t.Fatalf("TestConnection() error = %v", err)
	}
	if authHeader != "Bearer matrix-access-token" {
		t.Fatalf("Authorization = %q, want Bearer matrix-access-token", authHeader)
	}

	err := adapter.SendOutbound(context.Background(), conn, OutboundEnvelope{
		ExternalChatID: "!room:example.com",
		Text:           "reply body",
		ReplyHandle: map[string]any{
			"room_id":  "!room:example.com",
			"event_id": "$event1",
		},
		Attachments: []OutboundAttachment{
			{Name: "demo.txt", Type: "file"},
		},
	})
	if err != nil {
		t.Fatalf("SendOutbound() error = %v", err)
	}
	if !strings.Contains(sentPath, "/_matrix/client/v3/rooms/!room:example.com/send/m.room.message/") {
		t.Fatalf("sent path = %q, want room send path", sentPath)
	}
	if !strings.Contains(sentBody, `"body":"reply body\n\n[demo.txt]"`) {
		t.Fatalf("sent body = %s, want text fallback", sentBody)
	}
	if !strings.Contains(sentBody, `"event_id":"$event1"`) {
		t.Fatalf("sent body = %s, want reply relation", sentBody)
	}
}
