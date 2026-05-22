//go:build integration
// +build integration

package im

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"dotblue/internal/domains/conversation"
	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	"github.com/gogf/gf/v2/frame/g"
)

func TestExecuteWebChatTurnIntegration(t *testing.T) {
	ctx := context.Background()
	enterpriseID := requireIntegrationEnterpriseID(t, ctx)

	restore := installExecutionTestEngine()
	defer restore()

	fixture := routingFixture{
		ConnectionID:   "11111111-1111-7111-8111-111111111141",
		ConnectionName: buildWebChatConnectionName("11111111-1111-7111-8111-111111111142"),
		AgentID:        "11111111-1111-7111-8111-111111111142",
		AgentName:      "integration-web-agent",
		BindingID:      "11111111-1111-7111-8111-111111111143",
		Priority:       100,
	}

	cleanupWebChatIntegrationRows(t, ctx, fixture)
	cleanupDeliveryLogs(t, ctx, fixture.ConnectionID)
	t.Cleanup(func() {
		cleanupDeliveryLogs(t, ctx, fixture.ConnectionID)
		cleanupWebChatIntegrationRows(t, ctx, fixture)
	})

	seedWebChatFixture(t, ctx, enterpriseID, fixture)

	conv, err := conversation.Create("integration-user", enterpriseID, fixture.AgentID, "")
	if err != nil {
		t.Fatalf("create conversation failed: %v", err)
	}
	t.Cleanup(func() {
		if _, cleanupErr := g.DB().Model("messages").Ctx(ctx).Where("conversation_id = ?", conv.Id).Delete(); cleanupErr != nil {
			t.Fatalf("cleanup messages failed: %v", cleanupErr)
		}
		if _, cleanupErr := g.DB().Model("conversations").Ctx(ctx).Where("id = ?", conv.Id).Delete(); cleanupErr != nil {
			t.Fatalf("cleanup conversation failed: %v", cleanupErr)
		}
	})

	conn, routed, err := executeWebChatTurn(ctx, enterpriseID, "integration-user", fixture.AgentID, conv.Id, "hello from web chat integration")
	if err != nil {
		t.Fatalf("executeWebChatTurn() failed: %v", err)
	}
	if conn.Platform != PlatformWeb {
		t.Fatalf("connection platform = %q, want %q", conn.Platform, PlatformWeb)
	}
	if routed == nil || routed.AssistantReply == nil {
		t.Fatal("routed assistant reply is nil")
	}
	if routed.ConversationID != conv.Id {
		t.Fatalf("conversation id = %q, want %q", routed.ConversationID, conv.Id)
	}
	if routed.AssistantReply.Content != "integration assistant reply" {
		t.Fatalf("assistant content = %q, want integration assistant reply", routed.AssistantReply.Content)
	}

	conversationRow, err := defaultConnectionRepository.GetConversationSnapshot(ctx, conv.Id)
	if err != nil {
		t.Fatalf("load conversation snapshot failed: %v", err)
	}
	if conversationRow == nil {
		t.Fatal("conversation snapshot is nil")
	}
	if conversationRow.SourceType != PlatformWeb || conversationRow.SourceConnectionID != fixture.ConnectionID {
		t.Fatalf("conversation source = %+v, want web/%s", conversationRow, fixture.ConnectionID)
	}
	updatedConversation, err := conversation.GetById(conv.Id)
	if err != nil {
		t.Fatalf("load conversation failed: %v", err)
	}
	if updatedConversation == nil || updatedConversation.Title == "" {
		t.Fatal("conversation title is empty, want auto title")
	}

	messageRow, err := defaultConnectionRepository.GetMessageSnapshot(ctx, routed.MessageID)
	if err != nil {
		t.Fatalf("load message snapshot failed: %v", err)
	}
	if messageRow == nil {
		t.Fatal("message snapshot is nil")
	}
	if messageRow.DeliveryStatus != "received" {
		t.Fatalf("message delivery status = %q, want received", messageRow.DeliveryStatus)
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
}

func TestExecuteWebChatTurnAutoCreatesWebChannelIntegration(t *testing.T) {
	ctx := context.Background()
	enterpriseID := requireIntegrationEnterpriseID(t, ctx)

	restore := installExecutionTestEngine()
	defer restore()

	fixture := routingFixture{
		AgentID:   "11111111-1111-7111-8111-111111111144",
		AgentName: "integration-web-agent-autocreate",
	}

	cleanupAutoCreatedWebChatRows(t, ctx, fixture.AgentID)
	t.Cleanup(func() {
		cleanupAutoCreatedWebChatRows(t, ctx, fixture.AgentID)
	})

	if _, err := g.DB().Model("agents").Ctx(ctx).Data(g.Map{
		"id":             fixture.AgentID,
		"user_id":        "integration-user",
		"group_id":       enterpriseID,
		"agent_name":     fixture.AgentName,
		"system_prompt":  "integration prompt",
		"hermes_api_key": "dotblue-integration-key",
		"engine_type":    "hermes",
	}).Insert(); err != nil {
		t.Fatalf("insert agent failed: %v", err)
	}

	conv, err := conversation.Create("integration-user", enterpriseID, fixture.AgentID, "")
	if err != nil {
		t.Fatalf("create conversation failed: %v", err)
	}
	t.Cleanup(func() {
		if _, cleanupErr := g.DB().Model("messages").Ctx(ctx).Where("conversation_id = ?", conv.Id).Delete(); cleanupErr != nil {
			t.Fatalf("cleanup messages failed: %v", cleanupErr)
		}
		if _, cleanupErr := g.DB().Model("conversations").Ctx(ctx).Where("id = ?", conv.Id).Delete(); cleanupErr != nil {
			t.Fatalf("cleanup conversation failed: %v", cleanupErr)
		}
	})

	conn, routed, err := executeWebChatTurn(ctx, enterpriseID, "integration-user", fixture.AgentID, conv.Id, "hello auto-created web chat integration")
	if err != nil {
		t.Fatalf("executeWebChatTurn() failed: %v", err)
	}
	if conn.Platform != PlatformWeb {
		t.Fatalf("connection platform = %q, want %q", conn.Platform, PlatformWeb)
	}
	if conn.Name != buildWebChatConnectionName(fixture.AgentID) {
		t.Fatalf("connection name = %q, want %q", conn.Name, buildWebChatConnectionName(fixture.AgentID))
	}
	if conn.Status != StatusActive {
		t.Fatalf("connection status = %q, want %q", conn.Status, StatusActive)
	}
	if routed == nil || routed.AssistantReply == nil {
		t.Fatal("routed assistant reply is nil")
	}
	if routed.AssistantReply.Content != "integration assistant reply" {
		t.Fatalf("assistant content = %q, want integration assistant reply", routed.AssistantReply.Content)
	}

	connection, err := findWebChatConnection(ctx, enterpriseID, fixture.AgentID)
	if err != nil {
		t.Fatalf("findWebChatConnection() failed: %v", err)
	}
	if connection == nil {
		t.Fatal("auto-created web connection not found")
	}
	if connection.ID != conn.ID {
		t.Fatalf("connection id = %q, want %q", connection.ID, conn.ID)
	}

	bindings, err := defaultBindingService.ListBindingsByConnection(ctx, enterpriseID, conn.ID)
	if err != nil {
		t.Fatalf("ListBindingsByConnection() failed: %v", err)
	}
	if len(bindings) != 1 {
		t.Fatalf("binding count = %d, want 1", len(bindings))
	}
	binding := bindings[0]
	if binding.AgentID != fixture.AgentID {
		t.Fatalf("binding agent = %q, want %q", binding.AgentID, fixture.AgentID)
	}
	if binding.Status != StatusActive || binding.TriggerMode != TriggerModeAllMessages || binding.SessionStrategy != SessionStrategyPerChat {
		t.Fatalf("binding config = %+v, want active/all_messages/per_chat", binding)
	}
	if binding.AllowGroup || !binding.AllowDM {
		t.Fatalf("binding dm/group flags = allowGroup:%v allowDM:%v, want false/true", binding.AllowGroup, binding.AllowDM)
	}

	externalCount, err := defaultConnectionRepository.CountExternalSessions(ctx, conn.ID, fixture.AgentID)
	if err != nil {
		t.Fatalf("count external sessions failed: %v", err)
	}
	if externalCount != 1 {
		t.Fatalf("external session count = %d, want 1", externalCount)
	}

	conversationRow, err := defaultConnectionRepository.GetConversationSnapshot(ctx, conv.Id)
	if err != nil {
		t.Fatalf("load conversation snapshot failed: %v", err)
	}
	if conversationRow == nil {
		t.Fatal("conversation snapshot is nil")
	}
	if conversationRow.SourceType != PlatformWeb || conversationRow.SourceConnectionID != conn.ID {
		t.Fatalf("conversation source = %+v, want web/%s", conversationRow, conn.ID)
	}

	deliveryRow, err := defaultConnectionRepository.GetLatestDeliveryLogByConnection(ctx, conn.ID)
	if err != nil {
		t.Fatalf("load delivery log failed: %v", err)
	}
	if deliveryRow == nil {
		t.Fatal("delivery log snapshot is nil")
	}
	if deliveryRow.Status != "accepted" {
		t.Fatalf("delivery status = %q, want accepted", deliveryRow.Status)
	}
}

func TestExecuteWebChatTurnReusesExistingSessionIntegration(t *testing.T) {
	ctx := context.Background()
	enterpriseID := requireIntegrationEnterpriseID(t, ctx)

	restore := installExecutionTestEngine()
	defer restore()

	fixture := routingFixture{
		AgentID:   "11111111-1111-7111-8111-111111111145",
		AgentName: "integration-web-agent-reuse",
	}

	cleanupAutoCreatedWebChatRows(t, ctx, fixture.AgentID)
	t.Cleanup(func() {
		cleanupAutoCreatedWebChatRows(t, ctx, fixture.AgentID)
	})

	if _, err := g.DB().Model("agents").Ctx(ctx).Data(g.Map{
		"id":             fixture.AgentID,
		"user_id":        "integration-user",
		"group_id":       enterpriseID,
		"agent_name":     fixture.AgentName,
		"system_prompt":  "integration prompt",
		"hermes_api_key": "dotblue-integration-key",
		"engine_type":    "hermes",
	}).Insert(); err != nil {
		t.Fatalf("insert agent failed: %v", err)
	}

	conv, err := conversation.Create("integration-user", enterpriseID, fixture.AgentID, "")
	if err != nil {
		t.Fatalf("create conversation failed: %v", err)
	}
	t.Cleanup(func() {
		if _, cleanupErr := g.DB().Model("messages").Ctx(ctx).Where("conversation_id = ?", conv.Id).Delete(); cleanupErr != nil {
			t.Fatalf("cleanup messages failed: %v", cleanupErr)
		}
		if _, cleanupErr := g.DB().Model("conversations").Ctx(ctx).Where("id = ?", conv.Id).Delete(); cleanupErr != nil {
			t.Fatalf("cleanup conversation failed: %v", cleanupErr)
		}
	})

	firstConn, firstRouted, err := executeWebChatTurn(ctx, enterpriseID, "integration-user", fixture.AgentID, conv.Id, "hello from first web turn")
	if err != nil {
		t.Fatalf("first executeWebChatTurn() failed: %v", err)
	}
	secondConn, secondRouted, err := executeWebChatTurn(ctx, enterpriseID, "integration-user", fixture.AgentID, conv.Id, "hello from second web turn")
	if err != nil {
		t.Fatalf("second executeWebChatTurn() failed: %v", err)
	}

	if firstConn.ID != secondConn.ID {
		t.Fatalf("connection id mismatch: %q vs %q", firstConn.ID, secondConn.ID)
	}
	if firstRouted == nil || secondRouted == nil || firstRouted.ExternalSession == nil || secondRouted.ExternalSession == nil {
		t.Fatal("routed external session is nil")
	}
	if firstRouted.ExternalSession.ID != secondRouted.ExternalSession.ID {
		t.Fatalf("external session id mismatch: %q vs %q", firstRouted.ExternalSession.ID, secondRouted.ExternalSession.ID)
	}
	if firstRouted.ConversationID != secondRouted.ConversationID || secondRouted.ConversationID != conv.Id {
		t.Fatalf("conversation id mismatch: first=%q second=%q want=%q", firstRouted.ConversationID, secondRouted.ConversationID, conv.Id)
	}
	if firstRouted.SessionKey != secondRouted.SessionKey {
		t.Fatalf("session key mismatch: %q vs %q", firstRouted.SessionKey, secondRouted.SessionKey)
	}

	externalCount, err := defaultConnectionRepository.CountExternalSessions(ctx, firstConn.ID, fixture.AgentID)
	if err != nil {
		t.Fatalf("count external sessions failed: %v", err)
	}
	if externalCount != 1 {
		t.Fatalf("external session count = %d, want 1", externalCount)
	}

	bindings, err := defaultBindingService.ListBindingsByConnection(ctx, enterpriseID, firstConn.ID)
	if err != nil {
		t.Fatalf("ListBindingsByConnection() failed: %v", err)
	}
	if len(bindings) != 1 {
		t.Fatalf("binding count = %d, want 1", len(bindings))
	}

	messageCount, err := g.DB().Model("messages").Ctx(ctx).Where("conversation_id = ?", conv.Id).Count()
	if err != nil {
		t.Fatalf("count messages failed: %v", err)
	}
	if messageCount != 4 {
		t.Fatalf("message count = %d, want 4", messageCount)
	}
}

func seedWebChatFixture(t *testing.T, ctx context.Context, enterpriseID string, fixture routingFixture) {
	t.Helper()

	if _, err := g.DB().Model("im_connections").Ctx(ctx).Data(g.Map{
		"id":              fixture.ConnectionID,
		"enterprise_id":   enterpriseID,
		"platform":        PlatformWeb,
		"name":            fixture.ConnectionName,
		"status":          StatusActive,
		"connection_mode": WebConnectionModeDirect,
		"config_json":     fmt.Sprintf(`{"channel":"%s","agentId":"%s"}`, webConnectionChannel, fixture.AgentID),
		"secret_json":     `{}`,
		"callback_path":   buildConnectionCallbackPath(PlatformWeb, fixture.ConnectionID),
		"last_error":      "",
		"created_by":      "integration-test",
	}).Insert(); err != nil {
		t.Fatalf("insert web connection failed: %v", err)
	}

	if _, err := g.DB().Model("agents").Ctx(ctx).Data(g.Map{
		"id":             fixture.AgentID,
		"user_id":        "integration-user",
		"group_id":       enterpriseID,
		"agent_name":     fixture.AgentName,
		"system_prompt":  "integration prompt",
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
		"trigger_mode":        TriggerModeAllMessages,
		"trigger_config_json": `{}`,
		"session_strategy":    SessionStrategyPerChat,
		"reply_mode":          "default",
		"allow_group":         false,
		"allow_dm":            true,
		"priority":            fixture.Priority,
	}).Insert(); err != nil {
		t.Fatalf("insert binding failed: %v", err)
	}
}

func cleanupWebChatIntegrationRows(t *testing.T, ctx context.Context, fixture routingFixture) {
	t.Helper()

	if _, err := g.DB().Model("external_sessions").Ctx(ctx).Where("connection_id = ?", fixture.ConnectionID).Delete(); err != nil {
		t.Fatalf("cleanup external_sessions failed: %v", err)
	}
	if _, err := g.DB().Model("agent_channel_bindings").Ctx(ctx).Where("id = ?", fixture.BindingID).Delete(); err != nil {
		t.Fatalf("cleanup agent_channel_bindings failed: %v", err)
	}
	if _, err := g.DB().Model("agents").Ctx(ctx).Where("id = ?", fixture.AgentID).Delete(); err != nil {
		t.Fatalf("cleanup agents failed: %v", err)
	}
	if _, err := g.DB().Model("im_connections").Ctx(ctx).Where("id = ?", fixture.ConnectionID).Delete(); err != nil {
		t.Fatalf("cleanup im_connections failed: %v", err)
	}
}

func cleanupAutoCreatedWebChatRows(t *testing.T, ctx context.Context, agentID string) {
	t.Helper()

	connectionName := buildWebChatConnectionName(agentID)

	rows, err := defaultConnectionService.ListConnections(ctx, requireIntegrationEnterpriseID(t, ctx), ConnectionListFilters{
		Platform: PlatformWeb,
	})
	if err != nil {
		t.Fatalf("list web connections failed: %v", err)
	}

	connectionIDs := make([]string, 0, 1)
	for _, row := range rows {
		if strings.TrimSpace(row.Name) == connectionName || str(row.Config["agentId"]) == agentID {
			connectionIDs = append(connectionIDs, row.ID)
		}
	}

	if len(connectionIDs) > 0 {
		if _, err := g.DB().Model("channel_delivery_logs").Ctx(ctx).WhereIn("connection_id", connectionIDs).Delete(); err != nil {
			t.Fatalf("cleanup channel_delivery_logs failed: %v", err)
		}
		if _, err := g.DB().Model("external_sessions").Ctx(ctx).WhereIn("connection_id", connectionIDs).Delete(); err != nil {
			t.Fatalf("cleanup external_sessions failed: %v", err)
		}
		if _, err := g.DB().Model("agent_channel_bindings").Ctx(ctx).WhereIn("connection_id", connectionIDs).Delete(); err != nil {
			t.Fatalf("cleanup agent_channel_bindings failed: %v", err)
		}
		if _, err := g.DB().Model("im_connections").Ctx(ctx).WhereIn("id", connectionIDs).Delete(); err != nil {
			t.Fatalf("cleanup im_connections failed: %v", err)
		}
	}

	if _, err := g.DB().Model("agents").Ctx(ctx).Where("id = ?", agentID).Delete(); err != nil {
		t.Fatalf("cleanup agents failed: %v", err)
	}
}
