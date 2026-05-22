package im

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeFeishuSocketClient struct {
	start func(context.Context) error
}

func (c *fakeFeishuSocketClient) Start(ctx context.Context) error {
	if c.start != nil {
		return c.start(ctx)
	}
	return nil
}

func TestFeishuSocketRuntimeProcessesInboundAsync(t *testing.T) {
	t.Parallel()

	adapter := NewFeishuAdapter()
	var processed chan []InboundEvent
	processed = make(chan []InboundEvent, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"msg":"success","tenant_access_token":"tenant-token","expire":7200}`))
	}))
	defer server.Close()

	payload := mustBuildFeishuSocketPayload(t, "evt_runtime_async_1", "om_runtime_async_1")
	adapter.socketClientFactory = func(conn Connection, handler func(context.Context, []byte) error) feishuSocketClient {
		return &fakeFeishuSocketClient{
			start: func(ctx context.Context) error {
				if err := handler(ctx, payload); err != nil {
					return err
				}
				<-ctx.Done()
				return ctx.Err()
			},
		}
	}
	adapter.inboundProcessor = func(ctx context.Context, conn Connection, events []InboundEvent) (inboundPersistResult, error) {
		processed <- events
		return inboundPersistResult{Accepted: len(events)}, nil
	}

	conn := Connection{
		ID:             "conn_runtime_async_1",
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
	defer func() {
		_ = adapter.Stop(context.Background(), conn.ID)
	}()

	select {
	case events := <-processed:
		if len(events) != 1 {
			t.Fatalf("processed event count = %d, want 1", len(events))
		}
		if events[0].EventID != "evt_runtime_async_1" {
			t.Fatalf("event id = %q, want evt_runtime_async_1", events[0].EventID)
		}
		if events[0].ExternalChatID != "oc_runtime_async_1" {
			t.Fatalf("chat id = %q, want oc_runtime_async_1", events[0].ExternalChatID)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for async inbound processing")
	}
}

func TestNextRuntimeReconnectDelay(t *testing.T) {
	t.Parallel()

	base := 50 * time.Millisecond
	max := 200 * time.Millisecond
	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{attempt: 1, want: 50 * time.Millisecond},
		{attempt: 2, want: 100 * time.Millisecond},
		{attempt: 3, want: 200 * time.Millisecond},
		{attempt: 4, want: 200 * time.Millisecond},
	}

	for _, tt := range tests {
		if got := nextRuntimeReconnectDelay(tt.attempt, base, max); got != tt.want {
			t.Fatalf("attempt=%d delay=%s, want %s", tt.attempt, got, tt.want)
		}
	}
}

func TestFeishuSocketRuntimeReconnectsAfterUnexpectedError(t *testing.T) {
	t.Parallel()

	adapter := NewFeishuAdapter()
	adapter.reconnectBaseDelay = 10 * time.Millisecond
	adapter.reconnectMaxDelay = 20 * time.Millisecond

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"msg":"success","tenant_access_token":"tenant-token","expire":7200}`))
	}))
	defer server.Close()

	var attempts int
	processed := make(chan []InboundEvent, 1)
	payload := mustBuildFeishuSocketPayload(t, "evt_runtime_reconnect_1", "om_runtime_reconnect_1")

	adapter.socketClientFactory = func(conn Connection, handler func(context.Context, []byte) error) feishuSocketClient {
		attempts++
		currentAttempt := attempts
		return &fakeFeishuSocketClient{
			start: func(ctx context.Context) error {
				if currentAttempt == 1 {
					return errors.New("temporary socket failure")
				}
				if err := handler(ctx, payload); err != nil {
					return err
				}
				<-ctx.Done()
				return ctx.Err()
			},
		}
	}
	adapter.inboundProcessor = func(ctx context.Context, conn Connection, events []InboundEvent) (inboundPersistResult, error) {
		processed <- events
		return inboundPersistResult{Accepted: len(events)}, nil
	}

	conn := Connection{
		ID:             "conn_runtime_reconnect_1",
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
	defer func() {
		_ = adapter.Stop(context.Background(), conn.ID)
	}()

	select {
	case events := <-processed:
		if len(events) != 1 {
			t.Fatalf("processed event count = %d, want 1", len(events))
		}
		if events[0].EventID != "evt_runtime_reconnect_1" {
			t.Fatalf("event id = %q, want evt_runtime_reconnect_1", events[0].EventID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for reconnect processing")
	}

	if attempts < 2 {
		t.Fatalf("socket start attempts = %d, want at least 2", attempts)
	}
}

func mustBuildFeishuSocketPayload(t *testing.T, eventID, messageID string) []byte {
	t.Helper()

	payload := map[string]any{
		"schema": "2.0",
		"header": map[string]any{
			"event_id":    eventID,
			"event_type":  "im.message.receive_v1",
			"create_time": "1710000000000",
			"token":       "",
			"app_id":      "cli_xxx",
			"tenant_key":  "tenant_xxx",
		},
		"event": map[string]any{
			"sender": map[string]any{
				"sender_id": map[string]any{
					"open_id": "ou_runtime_user_1",
				},
			},
			"message": map[string]any{
				"message_id":   messageID,
				"root_id":      "",
				"parent_id":    "",
				"create_time":  "1710000000000",
				"chat_id":      "oc_runtime_async_1",
				"chat_type":    "group",
				"message_type": "text",
				"content":      `{"text":"hello socket runtime"}`,
				"mentions":     []any{},
			},
			"mentions": []any{
				map[string]any{"name": "bot"},
			},
		},
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload failed: %v", err)
	}
	return raw
}
