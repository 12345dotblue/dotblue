//go:build integration
// +build integration

package im

import (
	"context"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/google/uuid"
)

func TestPersistInboundEventsIntegration(t *testing.T) {
	ctx := context.Background()
	enterpriseID := requireIntegrationEnterpriseID(t, ctx)
	var err error

	const connectionID = "11111111-1111-7111-8111-111111111112"
	const eventID = "evt_integration_persist_1"
	const messageID = "om_integration_persist_1"

	cleanupIntegrationRows(t, ctx, connectionID)
	t.Cleanup(func() {
		cleanupIntegrationRows(t, ctx, connectionID)
	})

	_, err = g.DB().Model("im_connections").Ctx(ctx).Data(g.Map{
		"id":              connectionID,
		"enterprise_id":   enterpriseID,
		"platform":        "feishu",
		"name":            "integration-persist-inbound",
		"status":          StatusDisabled,
		"connection_mode": "socket_mode",
		"config_json":     `{"appId":"cli_integration"}`,
		"secret_json":     `{"appSecret":"integration-secret"}`,
		"callback_path":   buildConnectionCallbackPath("feishu", connectionID),
		"last_error":      "",
		"created_by":      "integration-test",
	}).Insert()
	if err != nil {
		t.Fatalf("insert connection failed: %v", err)
	}

	conn := Connection{
		ID:           connectionID,
		EnterpriseID: enterpriseID,
		Platform:     "feishu",
	}
	events := []InboundEvent{
		{
			Platform:       "feishu",
			EventID:        eventID,
			MessageID:      messageID,
			ExternalChatID: "oc_integration",
			Text:           "hello integration",
			RawPayload:     []byte(`{"hello":"integration"}`),
		},
	}

	result, err := persistInboundEvents(ctx, conn, events)
	if err != nil {
		t.Fatalf("persistInboundEvents() first call failed: %v", err)
	}
	if result.Accepted != 1 || result.Duplicated != 0 {
		t.Fatalf("first persist result = %+v, want accepted=1 duplicated=0", result)
	}

	result, err = persistInboundEvents(ctx, conn, events)
	if err != nil {
		t.Fatalf("persistInboundEvents() second call failed: %v", err)
	}
	if result.Accepted != 0 || result.Duplicated != 1 {
		t.Fatalf("second persist result = %+v, want accepted=0 duplicated=1", result)
	}

	count, err := defaultConnectionRepository.CountInboundEvents(ctx, connectionID, eventID)
	if err != nil {
		t.Fatalf("count events failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("event count = %d, want 1", count)
	}

	status, err := defaultConnectionRepository.GetConnectionStatusByID(ctx, connectionID)
	if err != nil {
		t.Fatalf("load connection status failed: %v", err)
	}
	if status != StatusActive {
		t.Fatalf("connection status = %q, want %q", status, StatusActive)
	}
}

func TestPersistInboundEventsMarksNoBindingStatus(t *testing.T) {
	ctx := context.Background()
	enterpriseID := requireIntegrationEnterpriseID(t, ctx)
	var err error

	connectionID := uuid.NewString()
	connectionName := "integration-persist-no-binding-" + connectionID
	eventID := "evt-no-binding-" + connectionID

	cleanupIntegrationRows(t, ctx, connectionID)
	t.Cleanup(func() {
		cleanupIntegrationRows(t, ctx, connectionID)
	})

	_, err = g.DB().Model("im_connections").Ctx(ctx).Data(g.Map{
		"id":              connectionID,
		"enterprise_id":   enterpriseID,
		"platform":        "feishu",
		"name":            connectionName,
		"status":          StatusDisabled,
		"connection_mode": "socket_mode",
		"config_json":     `{"appId":"cli_integration"}`,
		"secret_json":     `{"appSecret":"integration-secret"}`,
		"callback_path":   buildConnectionCallbackPath("feishu", connectionID),
		"last_error":      "",
		"created_by":      "integration-test",
	}).Insert()
	if err != nil {
		t.Fatalf("insert connection failed: %v", err)
	}

	conn := Connection{
		ID:           connectionID,
		EnterpriseID: enterpriseID,
		Platform:     "feishu",
	}
	events := []InboundEvent{
		{
			Platform:       "feishu",
			EventID:        eventID,
			MessageID:      "msg-" + connectionID,
			ExternalChatID: "chat-" + connectionID,
			Text:           "no binding here",
			RawPayload:     []byte(`{"hello":"no-binding"}`),
		},
	}

	result, err := persistInboundEvents(ctx, conn, events)
	if err != nil {
		t.Fatalf("persistInboundEvents() error = %v", err)
	}
	if result.Accepted != 1 || result.Duplicated != 0 {
		t.Fatalf("persist result = %+v, want accepted=1 duplicated=0", result)
	}

	status, err := defaultConnectionRepository.GetInboundEventStatus(ctx, connectionID, eventID)
	if err != nil {
		t.Fatalf("load event status failed: %v", err)
	}
	if status != "no_binding" {
		t.Fatalf("event status = %q, want no_binding", status)
	}
}

func cleanupIntegrationRows(t *testing.T, ctx context.Context, connectionID string) {
	t.Helper()

	if _, err := g.DB().Model("external_message_events").Ctx(ctx).Where("connection_id = ?", connectionID).Delete(); err != nil {
		t.Fatalf("cleanup external_message_events failed: %v", err)
	}
	if _, err := g.DB().Model("im_connections").Ctx(ctx).Where("id = ?", connectionID).Delete(); err != nil {
		t.Fatalf("cleanup im_connections failed: %v", err)
	}
}
