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
	"github.com/gogf/gf/v2/frame/g"

	"dotblue/internal/domains/engine"
)

func TestExecuteInboundTurnIntegration(t *testing.T) {
	ctx := context.Background()

	if err := g.DB().PingMaster(); err != nil {
		t.Skipf("database unavailable: %v", err)
	}

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

	const (
		connectionID = "11111111-1111-7111-8111-111111111131"
		agentID      = "11111111-1111-7111-8111-111111111132"
		bindingID    = "11111111-1111-7111-8111-111111111133"
	)

	cleanupRoutingIntegrationRows(t, ctx, connectionID, agentID, bindingID)
	cleanupDeliveryLogs(t, ctx, connectionID)
	t.Cleanup(func() {
		cleanupDeliveryLogs(t, ctx, connectionID)
		cleanupRoutingIntegrationRows(t, ctx, connectionID, agentID, bindingID)
	})

	seedRoutingFixtures(t, ctx, enterpriseID, connectionID, agentID, bindingID)

	conn := Connection{
		ID:           connectionID,
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

	var messageCount int
	messageCount, err = g.DB().Model("messages").Ctx(ctx).
		Where("conversation_id = ? AND role = ?", routed.ConversationID, "assistant").
		Count()
	if err != nil {
		t.Fatalf("count assistant messages failed: %v", err)
	}
	if messageCount != 1 {
		t.Fatalf("assistant message count = %d, want 1", messageCount)
	}

	var deliveryRow struct {
		Status       string `json:"status"`
		RequestJSON  string `json:"request_json"`
		ResponseJSON string `json:"response_json"`
		MessageID    string `json:"message_id"`
	}
	if err := g.DB().Model("channel_delivery_logs").Ctx(ctx).
		Where("connection_id = ?", connectionID).
		Order("created_at DESC").
		Scan(&deliveryRow); err != nil {
		t.Fatalf("load delivery log failed: %v", err)
	}
	if deliveryRow.Status != "accepted" {
		t.Fatalf("delivery status = %q, want accepted", deliveryRow.Status)
	}
	if !strings.Contains(deliveryRow.RequestJSON, "integration assistant reply") {
		t.Fatalf("delivery request_json = %q, want assistant content", deliveryRow.RequestJSON)
	}
}

func seedRoutingFixtures(t *testing.T, ctx context.Context, enterpriseID, connectionID, agentID, bindingID string) {
	t.Helper()

	if _, err := g.DB().Model("im_connections").Ctx(ctx).Data(g.Map{
		"id":              connectionID,
		"enterprise_id":   enterpriseID,
		"platform":        "feishu",
		"name":            "integration-routing-connection",
		"status":          StatusActive,
		"connection_mode": "socket_mode",
		"config_json":     `{"appId":"cli_integration"}`,
		"secret_json":     `{"appSecret":"integration-secret"}`,
		"callback_path":   buildConnectionCallbackPath("feishu", connectionID),
		"last_error":      "",
		"created_by":      "integration-test",
	}).Insert(); err != nil {
		t.Fatalf("insert connection failed: %v", err)
	}

	if _, err := g.DB().Model("agents").Ctx(ctx).Data(g.Map{
		"id":             agentID,
		"user_id":        "integration-user",
		"group_id":       enterpriseID,
		"agent_name":     "integration-agent",
		"system_prompt":  "integration prompt",
		"hermes_api_key": "dotblue-integration-key",
		"engine_type":    "hermes",
	}).Insert(); err != nil {
		t.Fatalf("insert agent failed: %v", err)
	}

	if _, err := g.DB().Model("agent_channel_bindings").Ctx(ctx).Data(g.Map{
		"id":                  bindingID,
		"enterprise_id":       enterpriseID,
		"agent_id":            agentID,
		"connection_id":       connectionID,
		"status":              StatusActive,
		"trigger_mode":        TriggerModeMentionOnly,
		"trigger_config_json": `{}`,
		"session_strategy":    SessionStrategyPerChatPerUser,
		"reply_mode":          "default",
		"allow_group":         true,
		"allow_dm":            true,
		"priority":            10,
	}).Insert(); err != nil {
		t.Fatalf("insert binding failed: %v", err)
	}
}

func cleanupDeliveryLogs(t *testing.T, ctx context.Context, connectionID string) {
	t.Helper()
	if _, err := g.DB().Model("channel_delivery_logs").Ctx(ctx).Where("connection_id = ?", connectionID).Delete(); err != nil {
		t.Fatalf("cleanup channel_delivery_logs failed: %v", err)
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
