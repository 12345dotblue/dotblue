//go:build integration
// +build integration

package im

import (
	"context"
	"testing"

	"dotblue/internal/infrastructure/dbschema"
	"github.com/gogf/gf/v2/frame/g"
)

type routingFixture struct {
	ConnectionID   string
	ConnectionName string
	AgentID        string
	AgentName      string
	BindingID      string
	UserID         string
	ModelID        string
	ModelName      string
	PriceID        string
	Priority       int
}

func requireIntegrationEnterpriseID(t *testing.T, ctx context.Context) string {
	t.Helper()

	if err := g.DB().PingMaster(); err != nil {
		t.Skipf("database unavailable: %v", err)
	}
	if err := dbschema.Ensure(ctx); err != nil {
		t.Fatalf("ensure schema failed: %v", err)
	}

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
	return enterpriseID
}

func seedRoutingFixture(t *testing.T, ctx context.Context, enterpriseID string, fixture routingFixture) {
	t.Helper()

	if _, err := g.DB().Model("llm_models").Ctx(ctx).Data(g.Map{
		"id":            fixture.ModelID,
		"scope":         "platform",
		"enterprise_id": "",
		"display_name":  fixture.ModelName,
		"provider_type": "openai",
		"api_base":      "https://ark.cn-beijing.volces.com/api/v3",
		"api_key":       "integration-test-key",
		"model_name":    "doubao-seed-2-0-mini-260428",
		"is_default":    false,
	}).Insert(); err != nil {
		t.Fatalf("insert llm model failed: %v", err)
	}

	if _, err := g.DB().Model("llm_model_prices").Ctx(ctx).Data(g.Map{
		"id":                       fixture.PriceID,
		"model_id":                 fixture.ModelID,
		"scope_type":               "platform",
		"scope_id":                 "",
		"currency":                 "USD",
		"cost_input_unit_price":    1,
		"cost_output_unit_price":   2,
		"charge_input_unit_price":  3,
		"charge_output_unit_price": 4,
	}).Insert(); err != nil {
		t.Fatalf("insert llm model price failed: %v", err)
	}

	if _, err := g.DB().Model("im_connections").Ctx(ctx).Data(g.Map{
		"id":              fixture.ConnectionID,
		"enterprise_id":   enterpriseID,
		"platform":        "feishu",
		"name":            fixture.ConnectionName,
		"status":          StatusActive,
		"connection_mode": "socket_mode",
		"config_json":     `{"appId":"cli_integration"}`,
		"secret_json":     `{"appSecret":"integration-secret"}`,
		"callback_path":   buildConnectionCallbackPath("feishu", fixture.ConnectionID),
		"last_error":      "",
		"created_by":      "integration-test",
	}).Insert(); err != nil {
		t.Fatalf("insert connection failed: %v", err)
	}

	if _, err := g.DB().Model("agents").Ctx(ctx).Data(g.Map{
		"id":             fixture.AgentID,
		"user_id":        fixture.UserID,
		"group_id":       enterpriseID,
		"agent_name":     fixture.AgentName,
		"system_prompt":  "integration prompt",
		"model_scope":    "platform",
		"model_id":       fixture.ModelID,
		"hermes_api_key": "dotblue-integration-key",
		"engine_type":    "hermes",
	}).Insert(); err != nil {
		t.Fatalf("insert agent failed: %v", err)
	}

	if _, err := g.DB().Model("agent_channel_bindings").Ctx(ctx).Data(g.Map{
		"id":                  fixture.BindingID,
		"enterprise_id":       enterpriseID,
		"agent_id":            fixture.AgentID,
		"connection_id":       fixture.ConnectionID,
		"status":              StatusActive,
		"trigger_mode":        TriggerModeMentionOnly,
		"trigger_config_json": `{}`,
		"session_strategy":    SessionStrategyPerChatPerUser,
		"reply_mode":          "default",
		"allow_group":         true,
		"allow_dm":            true,
		"priority":            fixture.Priority,
	}).Insert(); err != nil {
		t.Fatalf("insert binding failed: %v", err)
	}
}

func cleanupRoutingIntegrationRows(t *testing.T, ctx context.Context, fixture routingFixture) {
	t.Helper()

	var sessions []struct {
		ConversationID string `json:"conversation_id"`
	}
	if err := g.DB().Model("external_sessions").Ctx(ctx).Where("connection_id = ?", fixture.ConnectionID).Fields("conversation_id").Scan(&sessions); err != nil {
		t.Fatalf("load external session conversation ids failed: %v", err)
	}
	conversationIDs := make([]string, 0, len(sessions))
	for _, session := range sessions {
		if session.ConversationID != "" {
			conversationIDs = append(conversationIDs, session.ConversationID)
		}
	}
	if _, err := g.DB().Model("external_sessions").Ctx(ctx).Where("connection_id = ?", fixture.ConnectionID).Delete(); err != nil {
		t.Fatalf("cleanup external_sessions failed: %v", err)
	}
	if len(conversationIDs) > 0 {
		if _, err := g.DB().Model("messages").Ctx(ctx).WhereIn("conversation_id", conversationIDs).Delete(); err != nil {
			t.Fatalf("cleanup messages failed: %v", err)
		}
		if _, err := g.DB().Model("conversations").Ctx(ctx).WhereIn("id", conversationIDs).Delete(); err != nil {
			t.Fatalf("cleanup conversations failed: %v", err)
		}
	}
	if _, err := g.DB().Model("agent_channel_bindings").Ctx(ctx).Where("id = ?", fixture.BindingID).Delete(); err != nil {
		t.Fatalf("cleanup agent_channel_bindings failed: %v", err)
	}
	if _, err := g.DB().Model("llm_usage_events").Ctx(ctx).Where("agent_id = ?", fixture.AgentID).Delete(); err != nil {
		t.Fatalf("cleanup llm_usage_events failed: %v", err)
	}
	if _, err := g.DB().Model("llm_usage_daily_aggregates").Ctx(ctx).Where("scope_type = ? AND scope_id = ?", "agent", fixture.AgentID).Delete(); err != nil {
		t.Fatalf("cleanup llm_usage_daily_aggregates failed: %v", err)
	}
	if fixture.UserID != "" {
		if _, err := g.DB().Model("llm_usage_daily_aggregates").Ctx(ctx).Where("scope_type = ? AND scope_id = ?", "user", fixture.UserID).Delete(); err != nil {
			t.Fatalf("cleanup user llm_usage_daily_aggregates failed: %v", err)
		}
	}
	if _, err := g.DB().Model("agents").Ctx(ctx).Where("id = ?", fixture.AgentID).Delete(); err != nil {
		t.Fatalf("cleanup agents failed: %v", err)
	}
	if fixture.PriceID != "" {
		if _, err := g.DB().Model("llm_model_prices").Ctx(ctx).Where("id = ?", fixture.PriceID).Delete(); err != nil {
			t.Fatalf("cleanup llm_model_prices failed: %v", err)
		}
	}
	if fixture.ModelID != "" {
		if _, err := g.DB().Model("llm_models").Ctx(ctx).Where("id = ?", fixture.ModelID).Delete(); err != nil {
			t.Fatalf("cleanup llm_models failed: %v", err)
		}
	}
	if _, err := g.DB().Model("im_connections").Ctx(ctx).Where("id = ?", fixture.ConnectionID).Delete(); err != nil {
		t.Fatalf("cleanup im_connections failed: %v", err)
	}
}

func cleanupDeliveryLogs(t *testing.T, ctx context.Context, connectionID string) {
	t.Helper()
	if _, err := g.DB().Model("channel_delivery_logs").Ctx(ctx).Where("connection_id = ?", connectionID).Delete(); err != nil {
		t.Fatalf("cleanup channel_delivery_logs failed: %v", err)
	}
}
