//go:build integration
// +build integration

package im

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/google/uuid"
)

type fakeConnectionServiceAdapter struct {
	platform  string
	starts    atomic.Int32
	stops     atomic.Int32
	tests     atomic.Int32
	validates atomic.Int32
}

func (a *fakeConnectionServiceAdapter) Platform() string {
	return a.platform
}

func (a *fakeConnectionServiceAdapter) Start(ctx context.Context, conn Connection) error {
	a.starts.Add(1)
	return nil
}

func (a *fakeConnectionServiceAdapter) Stop(ctx context.Context, connectionID string) error {
	a.stops.Add(1)
	return nil
}

func (a *fakeConnectionServiceAdapter) ValidateConfig(config map[string]any, secrets map[string]any) error {
	a.validates.Add(1)
	return nil
}

func (a *fakeConnectionServiceAdapter) ParseInbound(ctx context.Context, raw any) ([]InboundEvent, error) {
	return nil, nil
}

func (a *fakeConnectionServiceAdapter) SendOutbound(ctx context.Context, conn Connection, msg OutboundEnvelope) error {
	return nil
}

func (a *fakeConnectionServiceAdapter) TestConnection(ctx context.Context, conn Connection) error {
	a.tests.Add(1)
	return nil
}

func registerConnectionServiceTestAdapter(t *testing.T, platform string) *fakeConnectionServiceAdapter {
	t.Helper()

	adapter := &fakeConnectionServiceAdapter{platform: platform}
	previous, hadPrevious := adapters[platform]
	RegisterAdapter(adapter)
	t.Cleanup(func() {
		if hadPrevious {
			adapters[platform] = previous
		} else {
			delete(adapters, platform)
		}
	})
	return adapter
}

func insertConnectionServiceTestConnection(t *testing.T, ctx context.Context, enterpriseID, platform, connectionID, connectionName string) {
	t.Helper()

	_, err := g.DB().Model("im_connections").Ctx(ctx).Data(g.Map{
		"id":              connectionID,
		"enterprise_id":   enterpriseID,
		"platform":        platform,
		"name":            connectionName,
		"status":          StatusDisabled,
		"connection_mode": "socket_mode",
		"config_json":     `{"appId":"service-test-app"}`,
		"secret_json":     `{"appSecret":"service-test-secret"}`,
		"callback_path":   buildConnectionCallbackPath(platform, connectionID),
		"last_error":      "",
		"created_by":      "integration-test",
	}).Insert()
	if err != nil {
		t.Fatalf("insert connection failed: %v", err)
	}
}

func insertConnectionServiceTestAgent(t *testing.T, ctx context.Context, enterpriseID, agentID, agentName string) {
	t.Helper()

	_, err := g.DB().Model("agents").Ctx(ctx).Data(g.Map{
		"id":             agentID,
		"user_id":        "integration-user",
		"group_id":       enterpriseID,
		"agent_name":     agentName,
		"system_prompt":  "integration prompt",
		"hermes_api_key": "dotblue-integration-key",
		"engine_type":    "hermes",
	}).Insert()
	if err != nil {
		t.Fatalf("insert agent failed: %v", err)
	}
}

func insertConnectionServiceTestConversation(t *testing.T, ctx context.Context, enterpriseID, conversationID, connectionID, agentID, title, sourceType string) {
	t.Helper()

	_, err := g.DB().Model("conversations").Ctx(ctx).Data(g.Map{
		"id":                   conversationID,
		"user_id":              "integration-user",
		"group_id":             enterpriseID,
		"agent_id":             agentID,
		"title":                title,
		"source_type":          sourceType,
		"source_connection_id": connectionID,
		"created_at":           time.Now(),
		"updated_at":           time.Now(),
	}).Insert()
	if err != nil {
		t.Fatalf("insert conversation failed: %v", err)
	}
}

func TestConnectionServiceTestConnectionUsesAdapterTester(t *testing.T) {
	ctx := context.Background()

	const platform = "service_test_tester"
	adapter := registerConnectionServiceTestAdapter(t, platform)
	enterpriseID := requireIntegrationEnterpriseID(t, ctx)

	connectionID := uuid.NewString()
	connectionName := "connection-service-test-" + connectionID
	cleanupIntegrationRows(t, ctx, connectionID)
	t.Cleanup(func() {
		cleanupIntegrationRows(t, ctx, connectionID)
	})

	insertConnectionServiceTestConnection(t, ctx, enterpriseID, platform, connectionID, connectionName)

	detail, err := defaultConnectionService.TestConnection(ctx, enterpriseID, connectionID)
	if err != nil {
		t.Fatalf("TestConnection() error = %v", err)
	}
	if detail != "connection test passed" {
		t.Fatalf("detail = %q, want connection test passed", detail)
	}
	if got := adapter.tests.Load(); got != 1 {
		t.Fatalf("test call count = %d, want 1", got)
	}
}

func TestConnectionServiceSetEnabledUsesAdapterLifecycle(t *testing.T) {
	ctx := context.Background()

	const platform = "service_test_runtime"
	adapter := registerConnectionServiceTestAdapter(t, platform)
	enterpriseID := requireIntegrationEnterpriseID(t, ctx)

	connectionID := uuid.NewString()
	connectionName := "connection-service-runtime-" + connectionID
	cleanupIntegrationRows(t, ctx, connectionID)
	t.Cleanup(func() {
		cleanupIntegrationRows(t, ctx, connectionID)
	})

	insertConnectionServiceTestConnection(t, ctx, enterpriseID, platform, connectionID, connectionName)

	connection, err := defaultConnectionService.SetEnabled(ctx, enterpriseID, connectionID, true)
	if err != nil {
		t.Fatalf("SetEnabled(enable) error = %v", err)
	}
	if connection.Status != StatusActive {
		t.Fatalf("enabled status = %q, want %q", connection.Status, StatusActive)
	}
	if got := adapter.starts.Load(); got != 1 {
		t.Fatalf("start call count = %d, want 1", got)
	}

	connection, err = defaultConnectionService.SetEnabled(ctx, enterpriseID, connectionID, false)
	if err != nil {
		t.Fatalf("SetEnabled(disable) error = %v", err)
	}
	if connection.Status != StatusDisabled {
		t.Fatalf("disabled status = %q, want %q", connection.Status, StatusDisabled)
	}
	if got := adapter.stops.Load(); got != 1 {
		t.Fatalf("stop call count = %d, want 1", got)
	}
}

func TestConnectionServiceCreateAndUpdateConnection(t *testing.T) {
	ctx := context.Background()

	enterpriseID := requireIntegrationEnterpriseID(t, ctx)
	const platform = "service_test_crud"
	adapter := &fakeConnectionServiceAdapter{platform: platform}
	previous, hadPrevious := adapters[platform]
	RegisterAdapter(adapter)
	t.Cleanup(func() {
		if hadPrevious {
			adapters[platform] = previous
		} else {
			delete(adapters, platform)
		}
	})

	connectionName := "connection-service-crud-" + uuid.NewString()
	connection, err := defaultConnectionService.CreateConnection(ctx, enterpriseID, "integration-user", createConnectionReq{
		Platform:       platform,
		Name:           connectionName,
		ConnectionMode: "socket_mode",
		Config: map[string]any{
			"appId": "service-test-app",
		},
		Secrets: map[string]any{
			"appSecret": "service-test-secret",
		},
	})
	if err != nil {
		t.Fatalf("CreateConnection() error = %v", err)
	}
	t.Cleanup(func() {
		cleanupIntegrationRows(t, ctx, connection.ID)
	})

	if connection.Name != connectionName {
		t.Fatalf("created name = %q, want %q", connection.Name, connectionName)
	}
	if connection.Status != StatusDisabled {
		t.Fatalf("created status = %q, want %q", connection.Status, StatusDisabled)
	}
	if connection.Secrets["appSecret"] != "service-test-secret" {
		t.Fatalf("created secret = %v, want service-test-secret", connection.Secrets["appSecret"])
	}

	updated, err := defaultConnectionService.UpdateConnection(ctx, enterpriseID, connection.ID, updateConnectionReq{
		Name: connectionName + "-updated",
		Secrets: map[string]any{
			"appSecret": maskedSecretValue,
			"token":     "new-token",
		},
	})
	if err != nil {
		t.Fatalf("UpdateConnection() error = %v", err)
	}

	if updated.Name != connectionName+"-updated" {
		t.Fatalf("updated name = %q, want %q", updated.Name, connectionName+"-updated")
	}
	if updated.Secrets["appSecret"] != "service-test-secret" {
		t.Fatalf("updated appSecret = %v, want original secret preserved", updated.Secrets["appSecret"])
	}
	if updated.Secrets["token"] != "new-token" {
		t.Fatalf("updated token = %v, want new-token", updated.Secrets["token"])
	}
}

func TestConnectionServiceGetConnectionByPlatform(t *testing.T) {
	ctx := context.Background()

	enterpriseID := requireIntegrationEnterpriseID(t, ctx)
	connectionID := uuid.NewString()
	connectionName := "connection-service-platform-" + connectionID
	cleanupIntegrationRows(t, ctx, connectionID)
	t.Cleanup(func() {
		cleanupIntegrationRows(t, ctx, connectionID)
	})

	insertConnectionServiceTestConnection(t, ctx, enterpriseID, "feishu", connectionID, connectionName)

	connection, err := defaultConnectionService.GetConnectionByPlatform(ctx, connectionID, "feishu")
	if err != nil {
		t.Fatalf("GetConnectionByPlatform() error = %v", err)
	}
	if connection.ID != connectionID {
		t.Fatalf("connection id = %q, want %q", connection.ID, connectionID)
	}

	_, err = defaultConnectionService.GetConnectionByPlatform(ctx, connectionID, "dingtalk")
	if err != ErrConnectionNotFound {
		t.Fatalf("platform mismatch error = %v, want %v", err, ErrConnectionNotFound)
	}

	_, err = defaultConnectionService.GetConnectionByPlatform(ctx, uuid.NewString(), "feishu")
	if err != ErrConnectionNotFound {
		t.Fatalf("missing connection error = %v, want %v", err, ErrConnectionNotFound)
	}
}

func TestConnectionServiceListConnectionEvents(t *testing.T) {
	ctx := context.Background()

	enterpriseID := requireIntegrationEnterpriseID(t, ctx)
	connectionID := uuid.NewString()
	connectionName := "connection-service-events-" + connectionID
	cleanupIntegrationRows(t, ctx, connectionID)
	t.Cleanup(func() {
		cleanupIntegrationRows(t, ctx, connectionID)
	})

	insertConnectionServiceTestConnection(t, ctx, enterpriseID, "service_test_events", connectionID, connectionName)

	baseTime := time.Now().Add(-time.Hour)
	for i := 0; i < 55; i++ {
		_, err := g.DB().Model("external_message_events").Ctx(ctx).Data(g.Map{
			"enterprise_id":       enterpriseID,
			"platform":            "service_test_events",
			"connection_id":       connectionID,
			"event_id":            fmt.Sprintf("evt-match-%02d", i),
			"external_message_id": fmt.Sprintf("msg-match-%02d", i),
			"direction":           "inbound",
			"payload_json":        "{}",
			"status":              "received",
			"error_message":       "",
			"created_at":          baseTime.Add(time.Duration(i) * time.Second),
		}).Insert()
		if err != nil {
			t.Fatalf("insert matching event %d failed: %v", i, err)
		}
	}
	_, err := g.DB().Model("external_message_events").Ctx(ctx).Data(g.Map{
		"enterprise_id":       enterpriseID,
		"platform":            "service_test_events",
		"connection_id":       connectionID,
		"event_id":            "evt-non-match",
		"external_message_id": "msg-non-match",
		"direction":           "outbound",
		"payload_json":        "{}",
		"status":              "failed",
		"error_message":       "boom",
		"created_at":          baseTime.Add(2 * time.Hour),
	}).Insert()
	if err != nil {
		t.Fatalf("insert non-matching event failed: %v", err)
	}

	rows, err := defaultConnectionService.ListConnectionEvents(ctx, enterpriseID, connectionID, ConnectionEventListFilters{
		Direction: " inbound ",
		Status:    "received",
		Limit:     999,
	})
	if err != nil {
		t.Fatalf("ListConnectionEvents() error = %v", err)
	}
	if len(rows) != 50 {
		t.Fatalf("event count = %d, want 50", len(rows))
	}
	if rows[0].EventID != "evt-match-54" {
		t.Fatalf("first event = %q, want evt-match-54", rows[0].EventID)
	}
	if rows[len(rows)-1].EventID != "evt-match-05" {
		t.Fatalf("last event = %q, want evt-match-05", rows[len(rows)-1].EventID)
	}

	_, err = defaultConnectionService.ListConnectionEvents(ctx, enterpriseID, uuid.NewString(), ConnectionEventListFilters{})
	if err != ErrConnectionNotFound {
		t.Fatalf("missing connection error = %v, want %v", err, ErrConnectionNotFound)
	}
}

func TestConnectionServiceListConnectionDeliveries(t *testing.T) {
	ctx := context.Background()

	enterpriseID := requireIntegrationEnterpriseID(t, ctx)
	connectionID := uuid.NewString()
	connectionName := "connection-service-deliveries-" + connectionID
	agentID := uuid.NewString()
	conversationID := uuid.NewString()
	cleanupDeliveryLogs(t, ctx, connectionID)
	cleanupIntegrationRows(t, ctx, connectionID)
	t.Cleanup(func() {
		cleanupDeliveryLogs(t, ctx, connectionID)
		if _, err := g.DB().Model("conversations").Ctx(ctx).Where("id = ?", conversationID).Delete(); err != nil {
			t.Fatalf("cleanup conversations failed: %v", err)
		}
		if _, err := g.DB().Model("agents").Ctx(ctx).Where("id = ?", agentID).Delete(); err != nil {
			t.Fatalf("cleanup agents failed: %v", err)
		}
		cleanupIntegrationRows(t, ctx, connectionID)
	})

	insertConnectionServiceTestConnection(t, ctx, enterpriseID, "service_test_deliveries", connectionID, connectionName)
	insertConnectionServiceTestAgent(t, ctx, enterpriseID, agentID, "connection-service-delivery-agent")
	insertConnectionServiceTestConversation(
		t,
		ctx,
		enterpriseID,
		conversationID,
		connectionID,
		agentID,
		"connection service delivery list test",
		"service_test_deliveries",
	)

	baseTime := time.Now().Add(-time.Hour)
	var err error
	for i := 0; i < 55; i++ {
		_, err := g.DB().Model("channel_delivery_logs").Ctx(ctx).Data(g.Map{
			"id":              uuid.NewString(),
			"enterprise_id":   enterpriseID,
			"platform":        "service_test_deliveries",
			"connection_id":   connectionID,
			"conversation_id": conversationID,
			"message_id":      nil,
			"attempt":         i + 1,
			"status":          "accepted",
			"request_json":    "{}",
			"response_json":   "{}",
			"error_message":   "",
			"created_at":      baseTime.Add(time.Duration(i) * time.Second),
		}).Insert()
		if err != nil {
			t.Fatalf("insert matching delivery %d failed: %v", i, err)
		}
	}
	_, err = g.DB().Model("channel_delivery_logs").Ctx(ctx).Data(g.Map{
		"id":              uuid.NewString(),
		"enterprise_id":   enterpriseID,
		"platform":        "service_test_deliveries",
		"connection_id":   connectionID,
		"conversation_id": conversationID,
		"message_id":      nil,
		"attempt":         999,
		"status":          "failed",
		"request_json":    "{}",
		"response_json":   "{}",
		"error_message":   "boom",
		"created_at":      baseTime.Add(2 * time.Hour),
	}).Insert()
	if err != nil {
		t.Fatalf("insert non-matching delivery failed: %v", err)
	}

	rows, err := defaultConnectionService.ListConnectionDeliveries(ctx, enterpriseID, connectionID, ConnectionDeliveryListFilters{
		Status: "accepted",
		Limit:  999,
	})
	if err != nil {
		t.Fatalf("ListConnectionDeliveries() error = %v", err)
	}
	if len(rows) != 50 {
		t.Fatalf("delivery count = %d, want 50", len(rows))
	}
	if rows[0].Attempt != 55 {
		t.Fatalf("first delivery attempt = %d, want 55", rows[0].Attempt)
	}
	if rows[len(rows)-1].Attempt != 6 {
		t.Fatalf("last delivery attempt = %d, want 6", rows[len(rows)-1].Attempt)
	}

	_, err = defaultConnectionService.ListConnectionDeliveries(ctx, enterpriseID, uuid.NewString(), ConnectionDeliveryListFilters{})
	if err != ErrConnectionNotFound {
		t.Fatalf("missing connection error = %v, want %v", err, ErrConnectionNotFound)
	}
}
