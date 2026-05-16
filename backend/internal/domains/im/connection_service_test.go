package im

import (
	"context"
	"sync/atomic"
	"testing"

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

func TestConnectionServiceTestConnectionUsesAdapterTester(t *testing.T) {
	ctx := context.Background()

	if err := g.DB().PingMaster(); err != nil {
		t.Skipf("database unavailable: %v", err)
	}

	const platform = "service_test_tester"
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

	var enterpriseID string
	value, err := g.DB().Model("enterprises").Ctx(ctx).Order("created_at ASC").Value("id")
	if err != nil {
		t.Fatalf("load enterprise id failed: %v", err)
	}
	if err := value.Scan(&enterpriseID); err != nil {
		t.Fatalf("scan enterprise id failed: %v", err)
	}
	if enterpriseID == "" {
		t.Skip("no enterprise data available for integration test")
	}

	connectionID := uuid.NewString()
	connectionName := "connection-service-test-" + connectionID
	cleanupIntegrationRows(t, ctx, connectionID)
	t.Cleanup(func() {
		cleanupIntegrationRows(t, ctx, connectionID)
	})

	_, err = g.DB().Model("im_connections").Ctx(ctx).Data(g.Map{
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

	if err := g.DB().PingMaster(); err != nil {
		t.Skipf("database unavailable: %v", err)
	}

	const platform = "service_test_runtime"
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

	var enterpriseID string
	value, err := g.DB().Model("enterprises").Ctx(ctx).Order("created_at ASC").Value("id")
	if err != nil {
		t.Fatalf("load enterprise id failed: %v", err)
	}
	if err := value.Scan(&enterpriseID); err != nil {
		t.Fatalf("scan enterprise id failed: %v", err)
	}
	if enterpriseID == "" {
		t.Skip("no enterprise data available for integration test")
	}

	connectionID := uuid.NewString()
	connectionName := "connection-service-runtime-" + connectionID
	cleanupIntegrationRows(t, ctx, connectionID)
	t.Cleanup(func() {
		cleanupIntegrationRows(t, ctx, connectionID)
	})

	_, err = g.DB().Model("im_connections").Ctx(ctx).Data(g.Map{
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
