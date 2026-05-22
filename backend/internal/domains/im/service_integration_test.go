//go:build integration
// +build integration

package im

import (
	"context"
	"encoding/json"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
)

func TestProcessInboundEventIntegration(t *testing.T) {
	ctx := context.Background()
	enterpriseID := requireIntegrationEnterpriseID(t, ctx)

	fixture := routingFixture{
		ConnectionID:   "11111111-1111-7111-8111-111111111121",
		ConnectionName: "integration-routing-connection",
		AgentID:        "11111111-1111-7111-8111-111111111122",
		AgentName:      "integration-agent",
		BindingID:      "11111111-1111-7111-8111-111111111123",
		Priority:       10,
	}

	cleanupRoutingIntegrationRows(t, ctx, fixture)
	t.Cleanup(func() {
		cleanupRoutingIntegrationRows(t, ctx, fixture)
	})

	seedRoutingFixture(t, ctx, enterpriseID, fixture)

	conn := Connection{
		ID:           fixture.ConnectionID,
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

	externalCount, err := defaultConnectionRepository.CountExternalSessions(ctx, fixture.ConnectionID, fixture.AgentID)
	if err != nil {
		t.Fatalf("count external sessions failed: %v", err)
	}
	if externalCount != 1 {
		t.Fatalf("external session count = %d, want 1", externalCount)
	}

	conversationRow, err := defaultConnectionRepository.GetConversationSnapshot(ctx, result.ConversationID)
	if err != nil {
		t.Fatalf("load conversation failed: %v", err)
	}
	if conversationRow == nil {
		t.Fatal("conversation snapshot is nil")
	}
	if conversationRow.SourceType != "feishu" || conversationRow.SourceConnectionID != fixture.ConnectionID {
		t.Fatalf("conversation source = %+v, want feishu/%s", conversationRow, fixture.ConnectionID)
	}

	messageRow, err := defaultConnectionRepository.GetMessageSnapshot(ctx, result.MessageID)
	if err != nil {
		t.Fatalf("load message failed: %v", err)
	}
	if messageRow == nil {
		t.Fatal("message snapshot is nil")
	}
	if messageRow.SourceMessageID != event.MessageID || messageRow.DeliveryStatus != "received" {
		t.Fatalf("message row = %+v, want source message id and received status", messageRow)
	}
	var meta map[string]any
	if err := json.Unmarshal([]byte(messageRow.MessageMetaJSON), &meta); err != nil {
		t.Fatalf("unmarshal message meta json failed: %v", err)
	}
	if meta["platform"] != "feishu" || meta["mentions_bot"] != true {
		t.Fatalf("message meta json = %+v, want platform=feishu mentions_bot=true", meta)
	}
}

func TestProcessInboundEventReusesExternalSession(t *testing.T) {
	ctx := context.Background()
	enterpriseID := requireIntegrationEnterpriseID(t, ctx)

	fixture := routingFixture{
		ConnectionID:   "11111111-1111-7111-8111-111111111124",
		ConnectionName: "integration-routing-reuse",
		AgentID:        "11111111-1111-7111-8111-111111111125",
		AgentName:      "integration-agent",
		BindingID:      "11111111-1111-7111-8111-111111111126",
		Priority:       10,
	}

	cleanupRoutingIntegrationRows(t, ctx, fixture)
	t.Cleanup(func() {
		cleanupRoutingIntegrationRows(t, ctx, fixture)
	})

	seedRoutingFixture(t, ctx, enterpriseID, fixture)

	conn := Connection{
		ID:           fixture.ConnectionID,
		EnterpriseID: enterpriseID,
		Platform:     "feishu",
	}
	firstEvent := InboundEvent{
		Platform:         "feishu",
		EventID:          "evt_process_reuse_1",
		MessageID:        "om_process_reuse_1",
		ExternalChatID:   "oc_process_reuse",
		ExternalThreadID: "ot_process_reuse",
		ExternalUserID:   "ou_process_reuse",
		ChatType:         "group",
		MentionsBot:      true,
		Text:             "hello from integration routing",
		ReplyHandle: map[string]any{
			"message_id": "om_process_reuse_1",
		},
	}
	secondEvent := firstEvent
	secondEvent.EventID = "evt_process_reuse_2"
	secondEvent.MessageID = "om_process_reuse_2"
	secondEvent.Text = "hello again from integration routing"
	secondEvent.ReplyHandle = map[string]any{
		"message_id": "om_process_reuse_2",
	}

	firstResult, err := ProcessInboundEvent(ctx, conn, firstEvent)
	if err != nil {
		t.Fatalf("first ProcessInboundEvent() failed: %v", err)
	}
	secondResult, err := ProcessInboundEvent(ctx, conn, secondEvent)
	if err != nil {
		t.Fatalf("second ProcessInboundEvent() failed: %v", err)
	}

	if firstResult.ExternalSession == nil || secondResult.ExternalSession == nil {
		t.Fatal("external session is nil")
	}
	if firstResult.ExternalSession.ID != secondResult.ExternalSession.ID {
		t.Fatalf("external session id mismatch: %q vs %q", firstResult.ExternalSession.ID, secondResult.ExternalSession.ID)
	}
	if firstResult.ConversationID != secondResult.ConversationID {
		t.Fatalf("conversation id mismatch: %q vs %q", firstResult.ConversationID, secondResult.ConversationID)
	}

	externalCount, err := defaultConnectionRepository.CountExternalSessions(ctx, fixture.ConnectionID, fixture.AgentID)
	if err != nil {
		t.Fatalf("count external sessions failed: %v", err)
	}
	if externalCount != 1 {
		t.Fatalf("external session count = %d, want 1", externalCount)
	}
}
