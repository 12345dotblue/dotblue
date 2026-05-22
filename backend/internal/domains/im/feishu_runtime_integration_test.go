//go:build integration
// +build integration

package im

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/google/uuid"
)

func TestFeishuSocketRuntimeIntegration(t *testing.T) {
	ctx := context.Background()

	enterpriseID := requireIntegrationEnterpriseID(t, ctx)
	connectionID := uuid.NewString()
	connectionName := "integration-feishu-socket-runtime-" + connectionID
	cleanupIntegrationRows(t, ctx, connectionID)
	t.Cleanup(func() {
		cleanupIntegrationRows(t, ctx, connectionID)
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"msg":"success","tenant_access_token":"tenant-token","expire":7200}`))
	}))
	defer server.Close()

	_, err := g.DB().Model("im_connections").Ctx(ctx).Data(g.Map{
		"id":              connectionID,
		"enterprise_id":   enterpriseID,
		"platform":        "feishu",
		"name":            connectionName,
		"status":          StatusDisabled,
		"connection_mode": "socket_mode",
		"config_json":     `{"appId":"cli_integration","apiBaseURL":"` + server.URL + `"}`,
		"secret_json":     `{"appSecret":"integration-secret"}`,
		"callback_path":   buildConnectionCallbackPath("feishu", connectionID),
		"last_error":      "",
		"created_by":      "integration-test",
	}).Insert()
	if err != nil {
		t.Fatalf("insert connection failed: %v", err)
	}

	adapter := NewFeishuAdapter()
	payload := mustBuildFeishuSocketPayload(t, "evt_socket_integration_1", "om_socket_integration_1")

	conn := Connection{
		ID:             connectionID,
		EnterpriseID:   enterpriseID,
		Platform:       "feishu",
		ConnectionMode: "socket_mode",
		Config: map[string]any{
			"appId":      "cli_integration",
			"apiBaseURL": server.URL,
		},
		Secrets: map[string]any{
			"appSecret": "integration-secret",
		},
	}

	if err := adapter.processFeishuSocketPayload(ctx, conn, payload); err != nil {
		t.Fatalf("processFeishuSocketPayload() error = %v", err)
	}

	count, err := g.DB().Model("external_message_events").Ctx(ctx).
		Where("connection_id = ? AND event_id = ?", connectionID, "evt_socket_integration_1").
		Count()
	if err != nil {
		t.Fatalf("count events failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("event count = %d, want 1", count)
	}

	var status string
	statusValue, err := g.DB().Model("im_connections").Ctx(ctx).Where("id = ?", connectionID).Value("status")
	if err != nil {
		t.Fatalf("load connection status failed: %v", err)
	}
	if err := statusValue.Scan(&status); err != nil {
		t.Fatalf("scan connection status failed: %v", err)
	}
	if status != StatusActive {
		t.Fatalf("connection status = %q, want %q", status, StatusActive)
	}
}

func TestFeishuSocketRuntimeStopDoesNotMarkConnectionError(t *testing.T) {
	ctx := context.Background()

	enterpriseID := requireIntegrationEnterpriseID(t, ctx)
	connectionID := uuid.NewString()
	connectionName := "integration-feishu-socket-stop-" + connectionID
	cleanupIntegrationRows(t, ctx, connectionID)
	t.Cleanup(func() {
		cleanupIntegrationRows(t, ctx, connectionID)
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"msg":"success","tenant_access_token":"tenant-token","expire":7200}`))
	}))
	defer server.Close()

	_, err := g.DB().Model("im_connections").Ctx(ctx).Data(g.Map{
		"id":              connectionID,
		"enterprise_id":   enterpriseID,
		"platform":        "feishu",
		"name":            connectionName,
		"status":          StatusDisabled,
		"connection_mode": "socket_mode",
		"config_json":     `{"appId":"cli_integration","apiBaseURL":"` + server.URL + `"}`,
		"secret_json":     `{"appSecret":"integration-secret"}`,
		"callback_path":   buildConnectionCallbackPath("feishu", connectionID),
		"last_error":      "",
		"created_by":      "integration-test",
	}).Insert()
	if err != nil {
		t.Fatalf("insert connection failed: %v", err)
	}

	adapter := NewFeishuAdapter()
	adapter.socketClientFactory = func(conn Connection, handler func(context.Context, []byte) error) feishuSocketClient {
		return &fakeFeishuSocketClient{
			start: func(ctx context.Context) error {
				<-ctx.Done()
				return ctx.Err()
			},
		}
	}

	conn := Connection{
		ID:             connectionID,
		EnterpriseID:   enterpriseID,
		Platform:       "feishu",
		ConnectionMode: "socket_mode",
		Config: map[string]any{
			"appId":      "cli_integration",
			"apiBaseURL": server.URL,
		},
		Secrets: map[string]any{
			"appSecret": "integration-secret",
		},
	}

	if err := adapter.Start(ctx, conn); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := adapter.Stop(ctx, connectionID); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	time.Sleep(150 * time.Millisecond)

	var status, lastError string
	statusValue, err := g.DB().Model("im_connections").Ctx(ctx).Fields("status", "last_error").Where("id = ?", connectionID).One()
	if err != nil {
		t.Fatalf("load connection row failed: %v", err)
	}
	if err := statusValue.Struct(&struct {
		Status    *string `json:"status"`
		LastError *string `json:"last_error"`
	}{Status: &status, LastError: &lastError}); err != nil {
		t.Fatalf("scan connection row failed: %v", err)
	}
	if status != StatusDisabled {
		t.Fatalf("connection status = %q, want %q", status, StatusDisabled)
	}
	if strings.TrimSpace(lastError) != "" {
		t.Fatalf("last_error = %q, want empty", lastError)
	}
}

func TestFeishuSocketRuntimeErrorMarksConnectionError(t *testing.T) {
	ctx := context.Background()

	enterpriseID := requireIntegrationEnterpriseID(t, ctx)
	connectionID := uuid.NewString()
	connectionName := "integration-feishu-socket-error-" + connectionID
	cleanupIntegrationRows(t, ctx, connectionID)
	t.Cleanup(func() {
		cleanupIntegrationRows(t, ctx, connectionID)
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"msg":"success","tenant_access_token":"tenant-token","expire":7200}`))
	}))
	defer server.Close()

	_, err := g.DB().Model("im_connections").Ctx(ctx).Data(g.Map{
		"id":              connectionID,
		"enterprise_id":   enterpriseID,
		"platform":        "feishu",
		"name":            connectionName,
		"status":          StatusActive,
		"connection_mode": "socket_mode",
		"config_json":     `{"appId":"cli_integration","apiBaseURL":"` + server.URL + `"}`,
		"secret_json":     `{"appSecret":"integration-secret"}`,
		"callback_path":   buildConnectionCallbackPath("feishu", connectionID),
		"last_error":      "",
		"created_by":      "integration-test",
	}).Insert()
	if err != nil {
		t.Fatalf("insert connection failed: %v", err)
	}

	adapter := NewFeishuAdapter()
	adapter.socketClientFactory = func(conn Connection, handler func(context.Context, []byte) error) feishuSocketClient {
		return &fakeFeishuSocketClient{
			start: func(ctx context.Context) error {
				return errors.New("socket runtime failed")
			},
		}
	}

	conn := Connection{
		ID:             connectionID,
		EnterpriseID:   enterpriseID,
		Platform:       "feishu",
		ConnectionMode: "socket_mode",
		Config: map[string]any{
			"appId":      "cli_integration",
			"apiBaseURL": server.URL,
		},
		Secrets: map[string]any{
			"appSecret": "integration-secret",
		},
	}

	if err := adapter.Start(ctx, conn); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() {
		_ = adapter.Stop(context.Background(), connectionID)
	}()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		row, rowErr := g.DB().Model("im_connections").Ctx(ctx).Fields("status", "last_error").Where("id = ?", connectionID).One()
		if rowErr != nil {
			t.Fatalf("load connection row failed: %v", rowErr)
		}
		var status, lastError string
		if err := row.Struct(&struct {
			Status    *string `json:"status"`
			LastError *string `json:"last_error"`
		}{Status: &status, LastError: &lastError}); err != nil {
			t.Fatalf("scan connection row failed: %v", err)
		}
		if status == StatusError {
			if !strings.Contains(lastError, "socket runtime failed") {
				t.Fatalf("last_error = %q, want socket runtime failed", lastError)
			}
			return
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Fatal("timed out waiting for runtime error status")
}

func TestFeishuSocketRuntimeReconnectIntegration(t *testing.T) {
	ctx := context.Background()

	enterpriseID := requireIntegrationEnterpriseID(t, ctx)
	connectionID := uuid.NewString()
	connectionName := "integration-feishu-socket-reconnect-" + connectionID
	cleanupIntegrationRows(t, ctx, connectionID)
	t.Cleanup(func() {
		cleanupIntegrationRows(t, ctx, connectionID)
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"msg":"success","tenant_access_token":"tenant-token","expire":7200}`))
	}))
	defer server.Close()

	_, err := g.DB().Model("im_connections").Ctx(ctx).Data(g.Map{
		"id":              connectionID,
		"enterprise_id":   enterpriseID,
		"platform":        "feishu",
		"name":            connectionName,
		"status":          StatusDisabled,
		"connection_mode": "socket_mode",
		"config_json":     `{"appId":"cli_integration","apiBaseURL":"` + server.URL + `"}`,
		"secret_json":     `{"appSecret":"integration-secret"}`,
		"callback_path":   buildConnectionCallbackPath("feishu", connectionID),
		"last_error":      "",
		"created_by":      "integration-test",
	}).Insert()
	if err != nil {
		t.Fatalf("insert connection failed: %v", err)
	}

	adapter := NewFeishuAdapter()
	adapter.reconnectBaseDelay = 20 * time.Millisecond
	adapter.reconnectMaxDelay = 40 * time.Millisecond

	payload := mustBuildFeishuSocketPayload(t, "evt_socket_reconnect_1", "om_socket_reconnect_1")
	var attempts int
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

	conn := Connection{
		ID:             connectionID,
		EnterpriseID:   enterpriseID,
		Platform:       "feishu",
		ConnectionMode: "socket_mode",
		Config: map[string]any{
			"appId":      "cli_integration",
			"apiBaseURL": server.URL,
		},
		Secrets: map[string]any{
			"appSecret": "integration-secret",
		},
	}

	if err := adapter.Start(ctx, conn); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() {
		_ = adapter.Stop(context.Background(), connectionID)
	}()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		count, countErr := g.DB().Model("external_message_events").Ctx(ctx).
			Where("connection_id = ? AND event_id = ?", connectionID, "evt_socket_reconnect_1").
			Count()
		if countErr != nil {
			t.Fatalf("count events failed: %v", countErr)
		}
		if count == 1 {
			if attempts < 2 {
				t.Fatalf("socket start attempts = %d, want at least 2", attempts)
			}
			var status string
			statusValue, statusErr := g.DB().Model("im_connections").Ctx(ctx).Where("id = ?", connectionID).Value("status")
			if statusErr != nil {
				t.Fatalf("load connection status failed: %v", statusErr)
			}
			if err := statusValue.Scan(&status); err != nil {
				t.Fatalf("scan connection status failed: %v", err)
			}
			if status == StatusActive {
				return
			}
		}
		time.Sleep(25 * time.Millisecond)
	}

	t.Fatal("timed out waiting for reconnect integration event persistence and active status")
}
