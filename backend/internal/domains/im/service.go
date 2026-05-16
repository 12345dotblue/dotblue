package im

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"dotblue/internal/domains/chat"
	"dotblue/internal/domains/conversation"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/google/uuid"
)

type RoutedInboundSession struct {
	Binding         AgentBinding       `json:"binding"`
	AgentID         string             `json:"agentId"`
	AgentUserID     string             `json:"agentUserId"`
	SessionAddress  SessionAddress     `json:"sessionAddress"`
	SessionKey      string             `json:"sessionKey"`
	ExternalSession *ExternalSession   `json:"externalSession"`
	ConversationID  string             `json:"conversationId"`
	MessageID       string             `json:"messageId"`
	AssistantReply  *chat.ExecutedTurn `json:"assistantReply,omitempty"`
}

type ExternalSession struct {
	ID               string     `json:"id"`
	EnterpriseID     string     `json:"enterpriseId"`
	Platform         string     `json:"platform"`
	ConnectionID     string     `json:"connectionId"`
	BindingID        string     `json:"bindingId"`
	AgentID          string     `json:"agentId"`
	SessionKey       string     `json:"sessionKey"`
	ExternalChatID   string     `json:"externalChatId"`
	ExternalThreadID string     `json:"externalThreadId"`
	ExternalUserID   string     `json:"externalUserId"`
	ConversationID   string     `json:"conversationId"`
	LastMessageAt    *time.Time `json:"lastMessageAt,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

func ProcessInboundEvent(ctx context.Context, conn Connection, event InboundEvent) (*RoutedInboundSession, error) {
	routed, err := ResolveInboundBinding(ctx, conn, event)
	if err != nil {
		return nil, err
	}

	addr := BuildSessionAddress(conn, event)
	sessionKey := BuildSessionKey(routed.Binding.AgentID, routed.Binding.SessionStrategy, addr)

	externalSession, err := ensureExternalSession(ctx, conn, routed, event, addr, sessionKey)
	if err != nil {
		return nil, err
	}

	messageID, err := persistInboundConversationMessage(ctx, externalSession.ConversationID, event)
	if err != nil {
		return nil, err
	}

	return &RoutedInboundSession{
		Binding:         routed.Binding,
		AgentID:         routed.Agent.Id,
		AgentUserID:     routed.Agent.UserId,
		SessionAddress:  addr,
		SessionKey:      sessionKey,
		ExternalSession: externalSession,
		ConversationID:  externalSession.ConversationID,
		MessageID:       messageID,
	}, nil
}

func ExecuteInboundTurn(ctx context.Context, conn Connection, routed *RoutedInboundSession, event InboundEvent) error {
	if routed == nil || routed.ExternalSession == nil {
		return errors.New("routed inbound session is incomplete")
	}

	prepared, err := chat.PrepareConversationExecution(
		ctx,
		routed.AgentUserID,
		conn.EnterpriseID,
		routed.Binding.AgentID,
		routed.ConversationID,
	)
	if err != nil {
		return err
	}

	executed, err := chat.ExecutePreparedTurn(ctx, prepared)
	if err != nil {
		return err
	}
	routed.AssistantReply = executed

	envelope := OutboundEnvelope{
		Platform:         conn.Platform,
		ConnectionID:     conn.ID,
		EnterpriseID:     conn.EnterpriseID,
		ConversationID:   routed.ConversationID,
		AgentID:          routed.Binding.AgentID,
		ExternalChatID:   event.ExternalChatID,
		ExternalThreadID: event.ExternalThreadID,
		ReplyHandle:      safeMap(event.ReplyHandle),
		Text:             executed.Content,
	}

	return deliverOutboundEnvelope(ctx, conn, envelope, executed.AssistantMessageID)
}

func ensureExternalSession(
	ctx context.Context,
	conn Connection,
	routed *RoutedBinding,
	event InboundEvent,
	addr SessionAddress,
	sessionKey string,
) (*ExternalSession, error) {
	current, err := getExternalSessionByKey(ctx, conn.EnterpriseID, sessionKey)
	if err != nil {
		return nil, err
	}
	if current != nil {
		if err := touchExternalSession(ctx, current, event); err != nil {
			return nil, err
		}
		return current, nil
	}

	convID, err := createExternalConversation(ctx, conn, routed, event)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	id := uuid.NewString()
	_, err = g.DB().Model("external_sessions").Ctx(ctx).Data(g.Map{
		"id":                 id,
		"enterprise_id":      conn.EnterpriseID,
		"platform":           conn.Platform,
		"connection_id":      conn.ID,
		"binding_id":         routed.Binding.ID,
		"agent_id":           routed.Binding.AgentID,
		"session_key":        sessionKey,
		"external_chat_id":   event.ExternalChatID,
		"external_thread_id": event.ExternalThreadID,
		"external_user_id":   event.ExternalUserID,
		"conversation_id":    convID,
		"last_message_at":    now,
		"created_at":         now,
		"updated_at":         now,
	}).Insert()
	if err != nil {
		return nil, err
	}

	return &ExternalSession{
		ID:               id,
		EnterpriseID:     conn.EnterpriseID,
		Platform:         conn.Platform,
		ConnectionID:     conn.ID,
		BindingID:        routed.Binding.ID,
		AgentID:          routed.Binding.AgentID,
		SessionKey:       sessionKey,
		ExternalChatID:   event.ExternalChatID,
		ExternalThreadID: event.ExternalThreadID,
		ExternalUserID:   event.ExternalUserID,
		ConversationID:   convID,
		LastMessageAt:    &now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

func getExternalSessionByKey(ctx context.Context, enterpriseID, sessionKey string) (*ExternalSession, error) {
	var session ExternalSession
	if err := g.DB().Model("external_sessions").Ctx(ctx).
		Where("enterprise_id = ? AND session_key = ?", enterpriseID, sessionKey).
		Scan(&session); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if session.ID == "" {
		return nil, nil
	}
	return &session, nil
}

func touchExternalSession(ctx context.Context, session *ExternalSession, event InboundEvent) error {
	now := time.Now()
	_, err := g.DB().Model("external_sessions").Ctx(ctx).Data(g.Map{
		"last_message_at": now,
		"updated_at":      now,
	}).Where("id = ?", session.ID).Update()
	if err != nil {
		return err
	}
	session.LastMessageAt = &now
	session.UpdatedAt = now
	return nil
}

func createExternalConversation(ctx context.Context, conn Connection, routed *RoutedBinding, event InboundEvent) (string, error) {
	now := time.Now()
	id := uuid.NewString()
	title := buildExternalConversationTitle(conn.Platform, event)
	_, err := g.DB().Model("conversations").Ctx(ctx).Data(g.Map{
		"id":                   id,
		"user_id":              routed.Agent.UserId,
		"group_id":             conn.EnterpriseID,
		"agent_id":             routed.Binding.AgentID,
		"title":                title,
		"source_type":          conn.Platform,
		"source_connection_id": conn.ID,
		"source_chat_id":       event.ExternalChatID,
		"source_thread_id":     event.ExternalThreadID,
		"source_user_id":       event.ExternalUserID,
		"created_at":           now,
		"updated_at":           now,
	}).Insert()
	if err != nil {
		return "", err
	}
	return id, nil
}

func persistInboundConversationMessage(ctx context.Context, conversationID string, event InboundEvent) (string, error) {
	metaJSON, _ := json.Marshal(g.Map{
		"platform":     event.Platform,
		"chat_type":    event.ChatType,
		"mentions_bot": event.MentionsBot,
		"attachments":  event.Attachments,
		"segments":     event.Segments,
		"reply_handle": event.ReplyHandle,
	})

	message, err := conversation.SaveMessage(conversationID, "user", event.Text, "", "", "done")
	if err != nil {
		return "", err
	}
	_, err = g.DB().Model("messages").Ctx(ctx).Data(g.Map{
		"source_message_id": event.MessageID,
		"delivery_status":   "received",
		"message_meta_json": string(metaJSON),
	}).Where("id = ?", message.Id).Update()
	if err != nil {
		return "", err
	}
	if err := conversation.TouchUpdated(conversationID); err != nil {
		return "", err
	}
	return message.Id, nil
}

func buildExternalConversationTitle(platform string, event InboundEvent) string {
	if event.Text != "" {
		runes := []rune(event.Text)
		if len(runes) > 40 {
			return string(runes[:40])
		}
		return event.Text
	}
	if len(event.Attachments) > 0 {
		return platform + " media conversation"
	}
	return platform + " inbound conversation"
}

func IsNoMatchedBinding(err error) bool {
	return errors.Is(err, ErrNoMatchedBinding)
}

func deliverOutboundEnvelope(ctx context.Context, conn Connection, envelope OutboundEnvelope, messageID string) error {
	requestJSON, _ := json.Marshal(envelope)
	logID := uuid.NewString()
	_, err := g.DB().Model("channel_delivery_logs").Ctx(ctx).Data(g.Map{
		"id":              logID,
		"enterprise_id":   conn.EnterpriseID,
		"platform":        conn.Platform,
		"connection_id":   conn.ID,
		"conversation_id": envelope.ConversationID,
		"message_id":      nullableUUID(messageID),
		"attempt":         1,
		"status":          "pending",
		"request_json":    string(requestJSON),
		"response_json":   "{}",
		"error_message":   "",
		"created_at":      time.Now(),
	}).Insert()
	if err != nil {
		return err
	}

	adapter, err := GetAdapter(conn.Platform)
	if err != nil {
		_, _ = updateDeliveryLog(ctx, logID, "error", `{"error":"adapter unavailable"}`, err.Error())
		return err
	}

	if err := adapter.SendOutbound(ctx, conn, envelope); err != nil {
		_, _ = updateDeliveryLog(ctx, logID, "error", `{"error":"send failed"}`, err.Error())
		return err
	}

	_, err = updateDeliveryLog(ctx, logID, "accepted", `{"status":"accepted"}`, "")
	return err
}

func updateDeliveryLog(ctx context.Context, id, status, responseJSON, errorMessage string) (sql.Result, error) {
	return g.DB().Model("channel_delivery_logs").Ctx(ctx).Data(g.Map{
		"status":        status,
		"response_json": responseJSON,
		"error_message": errorMessage,
	}).Where("id = ?", id).Update()
}

func nullableUUID(id string) any {
	if id == "" {
		return nil
	}
	return id
}
