//go:build integration
// +build integration

package im

import (
	"context"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	"github.com/gogf/gf/v2/frame/g"
)

func TestResolveInboundBindingPrefersLowestPriority(t *testing.T) {
	ctx := context.Background()
	enterpriseID := requireIntegrationEnterpriseID(t, ctx)

	const (
		connectionID       = "11111111-1111-7111-8111-111111111141"
		preferredAgentID   = "11111111-1111-7111-8111-111111111142"
		fallbackAgentID    = "11111111-1111-7111-8111-111111111143"
		preferredBindingID = "11111111-1111-7111-8111-111111111144"
		fallbackBindingID  = "11111111-1111-7111-8111-111111111145"
	)

	cleanupResolveBindingFixtures(t, ctx, connectionID, []string{preferredAgentID, fallbackAgentID}, []string{preferredBindingID, fallbackBindingID})
	t.Cleanup(func() {
		cleanupResolveBindingFixtures(t, ctx, connectionID, []string{preferredAgentID, fallbackAgentID}, []string{preferredBindingID, fallbackBindingID})
	})

	seedResolveBindingFixtures(t, ctx, enterpriseID, connectionID, preferredAgentID, fallbackAgentID, preferredBindingID, fallbackBindingID)

	result, err := ResolveInboundBinding(ctx, Connection{
		ID:           connectionID,
		EnterpriseID: enterpriseID,
		Platform:     "feishu",
	}, InboundEvent{
		Platform:    "feishu",
		ChatType:    "group",
		MentionsBot: true,
		Text:        "please route this",
	})
	if err != nil {
		t.Fatalf("ResolveInboundBinding() failed: %v", err)
	}
	if result == nil {
		t.Fatal("ResolveInboundBinding() returned nil result")
	}
	if result.Binding.ID != preferredBindingID {
		t.Fatalf("binding id = %q, want %q", result.Binding.ID, preferredBindingID)
	}
	if result.Agent == nil || result.Agent.Id != preferredAgentID {
		t.Fatalf("agent = %+v, want id %s", result.Agent, preferredAgentID)
	}
}

func seedResolveBindingFixtures(t *testing.T, ctx context.Context, enterpriseID, connectionID, preferredAgentID, fallbackAgentID, preferredBindingID, fallbackBindingID string) {
	t.Helper()

	if _, err := g.DB().Model("im_connections").Ctx(ctx).Data(g.Map{
		"id":              connectionID,
		"enterprise_id":   enterpriseID,
		"platform":        "feishu",
		"name":            "integration-router-priority",
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

	for _, agentFixture := range []struct {
		id   string
		name string
	}{
		{id: preferredAgentID, name: "integration-agent-preferred"},
		{id: fallbackAgentID, name: "integration-agent-fallback"},
	} {
		if _, err := g.DB().Model("agents").Ctx(ctx).Data(g.Map{
			"id":             agentFixture.id,
			"user_id":        "integration-user",
			"group_id":       enterpriseID,
			"agent_name":     agentFixture.name,
			"system_prompt":  "integration prompt",
			"hermes_api_key": "dotblue-integration-key",
			"engine_type":    "hermes",
		}).Insert(); err != nil {
			t.Fatalf("insert agent %s failed: %v", agentFixture.id, err)
		}
	}

	for _, bindingFixture := range []struct {
		id       string
		agentID  string
		priority int
	}{
		{id: fallbackBindingID, agentID: fallbackAgentID, priority: 20},
		{id: preferredBindingID, agentID: preferredAgentID, priority: 10},
	} {
		if _, err := g.DB().Model("agent_channel_bindings").Ctx(ctx).Data(g.Map{
			"id":                  bindingFixture.id,
			"enterprise_id":       enterpriseID,
			"agent_id":            bindingFixture.agentID,
			"connection_id":       connectionID,
			"status":              StatusActive,
			"trigger_mode":        TriggerModeMentionOnly,
			"trigger_config_json": `{}`,
			"session_strategy":    SessionStrategyPerChatPerUser,
			"reply_mode":          "default",
			"allow_group":         true,
			"allow_dm":            true,
			"priority":            bindingFixture.priority,
		}).Insert(); err != nil {
			t.Fatalf("insert binding %s failed: %v", bindingFixture.id, err)
		}
	}
}

func cleanupResolveBindingFixtures(t *testing.T, ctx context.Context, connectionID string, agentIDs, bindingIDs []string) {
	t.Helper()

	if len(bindingIDs) > 0 {
		if _, err := g.DB().Model("agent_channel_bindings").Ctx(ctx).WhereIn("id", bindingIDs).Delete(); err != nil {
			t.Fatalf("cleanup agent_channel_bindings failed: %v", err)
		}
	}
	if len(agentIDs) > 0 {
		if _, err := g.DB().Model("agents").Ctx(ctx).WhereIn("id", agentIDs).Delete(); err != nil {
			t.Fatalf("cleanup agents failed: %v", err)
		}
	}
	if _, err := g.DB().Model("im_connections").Ctx(ctx).Where("id = ?", connectionID).Delete(); err != nil {
		t.Fatalf("cleanup im_connections failed: %v", err)
	}
}
