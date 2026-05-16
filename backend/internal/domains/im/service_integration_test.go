package im

import (
	"context"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	"github.com/gogf/gf/v2/frame/g"
)

func TestProcessInboundEventIntegration(t *testing.T) {
	ctx := context.Background()

	if err := g.DB().PingMaster(); err != nil {
		t.Skipf("database unavailable: %v", err)
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

	const (
		connectionID = "11111111-1111-7111-8111-111111111121"
		agentID      = "11111111-1111-7111-8111-111111111122"
		bindingID    = "11111111-1111-7111-8111-111111111123"
	)

	cleanupRoutingIntegrationRows(t, ctx, connectionID, agentID, bindingID)
	t.Cleanup(func() {
		cleanupRoutingIntegrationRows(t, ctx, connectionID, agentID, bindingID)
	})

	_, err = g.DB().Model("im_connections").Ctx(ctx).Data(g.Map{
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
	}).Insert()
	if err != nil {
		t.Fatalf("insert connection failed: %v", err)
	}

	_, err = g.DB().Model("agents").Ctx(ctx).Data(g.Map{
		"id":             agentID,
		"user_id":        "integration-user",
		"group_id":       enterpriseID,
		"agent_name":     "integration-agent",
		"system_prompt":  "integration prompt",
		"hermes_api_key": "dotblue-integration-key",
		"engine_type":    "hermes",
	}).Insert()
	if err != nil {
		t.Fatalf("insert agent failed: %v", err)
	}

	_, err = g.DB().Model("agent_channel_bindings").Ctx(ctx).Data(g.Map{
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
	}).Insert()
	if err != nil {
		t.Fatalf("insert binding failed: %v", err)
	}

	conn := Connection{
		ID:           connectionID,
		EnterpriseID: enterpriseID,
		Platform:     "feishu",
	}
	event := InboundEvent{
		Platform:         "feishu",
		EventID:          "evt_process_integration_1",
		MessageID:        "om_process_integration_1",
		ExternalChatID:   "oc_process_integration",
		ExternalThreadID: "ot_process_integration",
		ExternalUserID:   "ou_process_integration",
		ChatType:         "group",
		MentionsBot:      true,
		Text:             "hello from integration routing",
		ReplyHandle: map[string]any{
			"message_id": "om_process_integration_1",
		},
	}

	result, err := ProcessInboundEvent(ctx, conn, event)
	if err != nil {
		t.Fatalf("ProcessInboundEvent() failed: %v", err)
	}
	if result.ConversationID == "" || result.MessageID == "" {
		t.Fatalf("ProcessInboundEvent() result = %+v, want conversation/message ids", result)
	}

	var externalCount int
	externalCount, err = g.DB().Model("external_sessions").Ctx(ctx).
		Where("connection_id = ? AND agent_id = ?", connectionID, agentID).
		Count()
	if err != nil {
		t.Fatalf("count external sessions failed: %v", err)
	}
	if externalCount != 1 {
		t.Fatalf("external session count = %d, want 1", externalCount)
	}

	var conversationRow struct {
		ID                 string `json:"id"`
		SourceType         string `json:"source_type"`
		SourceConnectionID string `json:"source_connection_id"`
	}
	if err := g.DB().Model("conversations").Ctx(ctx).Where("id = ?", result.ConversationID).Scan(&conversationRow); err != nil {
		t.Fatalf("load conversation failed: %v", err)
	}
	if conversationRow.SourceType != "feishu" || conversationRow.SourceConnectionID != connectionID {
		t.Fatalf("conversation source = %+v, want feishu/%s", conversationRow, connectionID)
	}

	var messageRow struct {
		ID              string `json:"id"`
		SourceMessageID string `json:"source_message_id"`
		DeliveryStatus  string `json:"delivery_status"`
	}
	if err := g.DB().Model("messages").Ctx(ctx).Where("id = ?", result.MessageID).Scan(&messageRow); err != nil {
		t.Fatalf("load message failed: %v", err)
	}
	if messageRow.SourceMessageID != event.MessageID || messageRow.DeliveryStatus != "received" {
		t.Fatalf("message row = %+v, want source message id and received status", messageRow)
	}
}

func cleanupRoutingIntegrationRows(t *testing.T, ctx context.Context, connectionID, agentID, bindingID string) {
	t.Helper()

	var sessions []struct {
		ConversationID string `json:"conversation_id"`
	}
	if err := g.DB().Model("external_sessions").Ctx(ctx).Where("connection_id = ?", connectionID).Fields("conversation_id").Scan(&sessions); err != nil {
		t.Fatalf("load external session conversation ids failed: %v", err)
	}
	conversationIDs := make([]string, 0, len(sessions))
	for _, session := range sessions {
		if session.ConversationID != "" {
			conversationIDs = append(conversationIDs, session.ConversationID)
		}
	}
	if _, err := g.DB().Model("external_sessions").Ctx(ctx).Where("connection_id = ?", connectionID).Delete(); err != nil {
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
	if _, err := g.DB().Model("agent_channel_bindings").Ctx(ctx).Where("id = ?", bindingID).Delete(); err != nil {
		t.Fatalf("cleanup agent_channel_bindings failed: %v", err)
	}
	if _, err := g.DB().Model("agents").Ctx(ctx).Where("id = ?", agentID).Delete(); err != nil {
		t.Fatalf("cleanup agents failed: %v", err)
	}
	if _, err := g.DB().Model("im_connections").Ctx(ctx).Where("id = ?", connectionID).Delete(); err != nil {
		t.Fatalf("cleanup im_connections failed: %v", err)
	}
}
