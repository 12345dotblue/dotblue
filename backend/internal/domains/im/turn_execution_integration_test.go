//go:build integration
// +build integration

package im

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"

	"dotblue/internal/domains/engine"
)

func TestExecuteInboundTurnIntegration(t *testing.T) {
	ctx := context.Background()
	enterpriseID := requireIntegrationEnterpriseID(t, ctx)

	restore := installExecutionTestEngine()
	defer restore()

	mockFeishu := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/open-apis/auth/v3/tenant_access_token/internal":
			_, _ = w.Write([]byte(`{"code":0,"msg":"success","tenant_access_token":"integration-token","expire":7200}`))
		case strings.HasPrefix(r.URL.Path, "/open-apis/im/v1/messages/") && strings.HasSuffix(r.URL.Path, "/reply"):
			var req map[string]any
			_ = json.NewDecoder(r.Body).Decode(&req)
			if got := r.Header.Get("Authorization"); got != "Bearer integration-token" {
				t.Fatalf("reply authorization = %q, want Bearer integration-token", got)
			}
			if !strings.Contains(str(req["content"]), "integration assistant reply") {
				t.Fatalf("reply content = %v, want integration assistant reply", req["content"])
			}
			_, _ = w.Write([]byte(`{"code":0,"msg":"success","data":{"message_id":"om_delivery"}}`))
		default:
			t.Fatalf("unexpected feishu request path: %s", r.URL.String())
		}
	}))
	defer mockFeishu.Close()

	fixture := routingFixture{
		ConnectionID:   "11111111-1111-7111-8111-111111111131",
		ConnectionName: "integration-routing-connection",
		AgentID:        "11111111-1111-7111-8111-111111111132",
		AgentName:      "integration-agent",
		BindingID:      "11111111-1111-7111-8111-111111111133",
		Priority:       10,
	}

	cleanupRoutingIntegrationRows(t, ctx, fixture)
	cleanupDeliveryLogs(t, ctx, fixture.ConnectionID)
	t.Cleanup(func() {
		cleanupDeliveryLogs(t, ctx, fixture.ConnectionID)
		cleanupRoutingIntegrationRows(t, ctx, fixture)
	})

	seedRoutingFixture(t, ctx, enterpriseID, fixture)

	conn := Connection{
		ID:           fixture.ConnectionID,
		EnterpriseID: enterpriseID,
		Platform:     "feishu",
		Config: map[string]any{
			"appId":          "cli_integration",
			"connectionMode": "socket_mode",
			"apiBaseURL":     mockFeishu.URL,
		},
		Secrets: map[string]any{
			"appSecret": "integration-secret",
		},
	}
	event := InboundEvent{
		Platform:         "feishu",
		EventID:          "evt_execute_integration_1",
		MessageID:        "om_execute_integration_1",
		ExternalChatID:   "oc_execute_integration",
		ExternalThreadID: "ot_execute_integration",
		ExternalUserID:   "ou_execute_integration",
		ChatType:         "group",
		MentionsBot:      true,
		Text:             "please respond from integration",
		ReplyHandle: map[string]any{
			"message_id": "om_execute_integration_1",
		},
	}

	routed, err := ProcessInboundEvent(ctx, conn, event)
	if err != nil {
		t.Fatalf("ProcessInboundEvent() failed: %v", err)
	}
	if err := ExecuteInboundTurn(ctx, conn, routed, event); err != nil {
		t.Fatalf("ExecuteInboundTurn() failed: %v", err)
	}
	if routed.AssistantReply == nil {
		t.Fatal("AssistantReply is nil")
	}
	if routed.AssistantReply.Content != "integration assistant reply" {
		t.Fatalf("assistant content = %q, want integration assistant reply", routed.AssistantReply.Content)
	}

	messageCount, err := defaultConnectionRepository.CountMessagesByConversationRole(ctx, routed.ConversationID, "assistant")
	if err != nil {
		t.Fatalf("count assistant messages failed: %v", err)
	}
	if messageCount != 1 {
		t.Fatalf("assistant message count = %d, want 1", messageCount)
	}

	deliveryRow, err := defaultConnectionRepository.GetLatestDeliveryLogByConnection(ctx, fixture.ConnectionID)
	if err != nil {
		t.Fatalf("load delivery log failed: %v", err)
	}
	if deliveryRow == nil {
		t.Fatal("delivery log snapshot is nil")
	}
	if deliveryRow.Status != "accepted" {
		t.Fatalf("delivery status = %q, want accepted", deliveryRow.Status)
	}
	if !strings.Contains(deliveryRow.RequestJSON, "integration assistant reply") {
		t.Fatalf("delivery request_json = %q, want assistant content", deliveryRow.RequestJSON)
	}
}

type executionTestRuntime struct{}

func (executionTestRuntime) EnsureRunning(ctx context.Context, orgID, userID, agentID string) (*engine.AgentEndpoint, error) {
	return &engine.AgentEndpoint{URL: "http://stub-engine", APIKey: "stub"}, nil
}

func (executionTestRuntime) Stop(ctx context.Context, agentID string) error {
	return nil
}

type executionTestEngine struct{}

func (executionTestEngine) Name() string { return "hermes" }

func (executionTestEngine) PrepareVolume(ctx context.Context, volPath string, agent *engine.AgentConfig, providerCfg *engine.ProviderConfig) error {
	return nil
}

func (executionTestEngine) ContainerSpec(agentID, volPath, containerPort string) (*engine.ContainerSpec, error) {
	return &engine.ContainerSpec{}, nil
}

func (executionTestEngine) ProxyRequest(ctx context.Context, endpoint *engine.AgentEndpoint, messages []interface{}, convID string) (*http.Response, error) {
	body := "data: {\"choices\":[{\"delta\":{\"reasoning_content\":\"integration thinking\"}}]}\n\n" +
		"data: {\"choices\":[{\"delta\":{\"content\":\"integration assistant reply\"}}]}\n\n" +
		"data: [DONE]\n\n"
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
	}, nil
}

func installExecutionTestEngine() func() {
	previousRuntime := engine.GetRuntime()
	engine.SetRuntime(executionTestRuntime{})
	engine.RegisterEngine(executionTestEngine{})
	return func() {
		engine.SetRuntime(previousRuntime)
		engine.RegisterEngine(&engine.HermesEngine{})
	}
}
