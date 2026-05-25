package im

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"dotblue/internal/domains/agent"
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
	ID               string     `json:"id" orm:"id"`
	EnterpriseID     string     `json:"enterpriseId" orm:"enterprise_id"`
	Platform         string     `json:"platform" orm:"platform"`
	ConnectionID     string     `json:"connectionId" orm:"connection_id"`
	BindingID        string     `json:"bindingId" orm:"binding_id"`
	AgentID          string     `json:"agentId" orm:"agent_id"`
	SessionKey       string     `json:"sessionKey" orm:"session_key"`
	ExternalChatID   string     `json:"externalChatId" orm:"external_chat_id"`
	ExternalThreadID string     `json:"externalThreadId" orm:"external_thread_id"`
	ExternalUserID   string     `json:"externalUserId" orm:"external_user_id"`
	ConversationID   string     `json:"conversationId" orm:"conversation_id"`
	LastMessageAt    *time.Time `json:"lastMessageAt,omitempty" orm:"last_message_at"`
	CreatedAt        time.Time  `json:"createdAt" orm:"created_at"`
	UpdatedAt        time.Time  `json:"updatedAt" orm:"updated_at"`
}

type imAgentDomain interface {
	GetById(id string) (*agent.Agent, error)
	BelongsToUser(id, userID, enterpriseID string) (bool, error)
}

type imChatDomain interface {
	PrepareConversationExecution(ctx context.Context, userID, enterpriseID, agentID, conversationID string) (*chat.PreparedTurn, error)
	ExecutePreparedTurn(ctx context.Context, prepared *chat.PreparedTurn) (*chat.ExecutedTurn, error)
}

type imConversationDomain interface {
	BelongsToUser(id, userId, enterpriseId string) (bool, error)
	GetById(id string) (*conversation.Conversation, error)
	SaveMessage(convId, role, content, thinking, toolCallsJson, status string) (*conversation.Message, error)
	AutoTitle(convId string)
}

type imRepository interface {
	GetExternalSessionByKey(ctx context.Context, enterpriseID, sessionKey string) (*ExternalSession, error)
	CreateExternalSession(ctx context.Context, record ExternalSessionRecord) error
	CreateConversation(ctx context.Context, record ConversationRecord) error
	TouchExternalSession(ctx context.Context, id string, lastMessageAt, updatedAt any) error
	UpdateConversationSource(ctx context.Context, id string, data g.Map) error
	UpdateInboundMessage(ctx context.Context, id, sourceMessageID, deliveryStatus, messageMetaJSON string) error
	TouchConversation(ctx context.Context, id string, updatedAt any) error
	InsertDeliveryLog(ctx context.Context, record DeliveryLogRecord) error
	UpdateDeliveryLog(ctx context.Context, id, status, responseJSON, errorMessage string) error
}

type adapterLookup func(platform string) (Adapter, error)

type Service struct {
	agents       imAgentDomain
	chat         imChatDomain
	conversation imConversationDomain
	repo         imRepository
	getAdapter   adapterLookup
	now          func() time.Time
	idGenerator  func() string
}

type defaultIMAgentDomain struct{}

func (defaultIMAgentDomain) GetById(id string) (*agent.Agent, error) {
	return agent.GetById(id)
}

func (defaultIMAgentDomain) BelongsToUser(id, userID, enterpriseID string) (bool, error) {
	return agent.BelongsToUser(id, userID, enterpriseID)
}

type defaultIMChatDomain struct{}

func (defaultIMChatDomain) PrepareConversationExecution(ctx context.Context, userID, enterpriseID, agentID, conversationID string) (*chat.PreparedTurn, error) {
	return chat.PrepareConversationExecution(ctx, userID, enterpriseID, agentID, conversationID)
}

func (defaultIMChatDomain) ExecutePreparedTurn(ctx context.Context, prepared *chat.PreparedTurn) (*chat.ExecutedTurn, error) {
	return chat.ExecutePreparedTurn(ctx, prepared)
}

type defaultIMConversationDomain struct{}

func (defaultIMConversationDomain) BelongsToUser(id, userId, enterpriseId string) (bool, error) {
	return conversation.BelongsToUser(id, userId, enterpriseId)
}

func (defaultIMConversationDomain) GetById(id string) (*conversation.Conversation, error) {
	return conversation.GetById(id)
}

func (defaultIMConversationDomain) SaveMessage(convId, role, content, thinking, toolCallsJson, status string) (*conversation.Message, error) {
	return conversation.SaveMessage(convId, role, content, thinking, toolCallsJson, status)
}

func (defaultIMConversationDomain) AutoTitle(convId string) {
	conversation.AutoTitle(convId)
}

func NewService() *Service {
	return &Service{
		agents:       defaultIMAgentDomain{},
		chat:         defaultIMChatDomain{},
		conversation: defaultIMConversationDomain{},
		repo:         defaultConnectionRepository,
		getAdapter:   GetAdapter,
		now:          time.Now,
		idGenerator:  uuid.NewString,
	}
}

var defaultService = NewService()

func (s *Service) agentDomain() imAgentDomain {
	if s == nil || s.agents == nil {
		return defaultIMAgentDomain{}
	}
	return s.agents
}

func (s *Service) conversationDomain() imConversationDomain {
	if s == nil || s.conversation == nil {
		return defaultIMConversationDomain{}
	}
	return s.conversation
}

func ProcessInboundEvent(ctx context.Context, conn Connection, event InboundEvent) (*RoutedInboundSession, error) {
	return defaultService.ProcessInboundEvent(ctx, conn, event)
}

func (s *Service) ProcessInboundEvent(ctx context.Context, conn Connection, event InboundEvent) (*RoutedInboundSession, error) {
	routed, err := resolveInboundBindingWith(ctx, defaultConnectionRepository, s.agentDomain(), conn, event)
	if err != nil {
		return nil, err
	}

	addr := BuildSessionAddress(conn, event)
	sessionKey := BuildSessionKey(routed.Binding.AgentID, routed.Binding.SessionStrategy, addr)

	externalSession, err := s.ensureExternalSession(ctx, conn, routed, event, addr, sessionKey)
	if err != nil {
		return nil, err
	}

	messageID, err := s.persistInboundConversationMessage(ctx, externalSession.ConversationID, event)
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
	return defaultService.ExecuteInboundTurn(ctx, conn, routed, event)
}

func (s *Service) ExecuteInboundTurn(ctx context.Context, conn Connection, routed *RoutedInboundSession, event InboundEvent) error {
	if routed == nil || routed.ExternalSession == nil {
		return errors.New("routed inbound session is incomplete")
	}
	g.Log().Debugf(
		ctx,
		"im.turn.start platform=%s conn=%s conv=%s agent=%s user=%s inboundMsg=%s textLen=%d",
		conn.Platform,
		conn.ID,
		routed.ConversationID,
		routed.Binding.AgentID,
		routed.AgentUserID,
		routed.MessageID,
		len(event.Text),
	)
	fmt.Printf(
		"TRACE im.turn.start platform=%s conn=%s conv=%s agent=%s user=%s inboundMsg=%s textLen=%d\n",
		conn.Platform,
		conn.ID,
		routed.ConversationID,
		routed.Binding.AgentID,
		routed.AgentUserID,
		routed.MessageID,
		len(event.Text),
	)

	prepared, err := s.chat.PrepareConversationExecution(
		ctx,
		routed.AgentUserID,
		conn.EnterpriseID,
		routed.Binding.AgentID,
		routed.ConversationID,
	)
	if err != nil {
		g.Log().Errorf(ctx, "im.turn.prepare.error conv=%s agent=%s err=%v", routed.ConversationID, routed.Binding.AgentID, err)
		return err
	}
	prepared.SourceType = "im"
	g.Log().Debugf(ctx, "im.turn.prepared conv=%s history=%d endpoint=%s", routed.ConversationID, len(prepared.History), prepared.Endpoint.URL)
	fmt.Printf("TRACE im.turn.prepared conv=%s history=%d endpoint=%s\n", routed.ConversationID, len(prepared.History), prepared.Endpoint.URL)

	executed, err := s.chat.ExecutePreparedTurn(ctx, prepared)
	if err != nil {
		g.Log().Errorf(ctx, "im.turn.execute.error conv=%s agent=%s err=%v", routed.ConversationID, routed.Binding.AgentID, err)
		return err
	}
	routed.AssistantReply = executed
	g.Log().Debugf(
		ctx,
		"im.turn.executed conv=%s assistantMsg=%s contentLen=%d thinkingLen=%d toolCalls=%d",
		routed.ConversationID,
		executed.AssistantMessageID,
		len(executed.Content),
		len(executed.Thinking),
		len(executed.ToolCalls),
	)
	fmt.Printf(
		"TRACE im.turn.executed conv=%s assistantMsg=%s contentLen=%d thinkingLen=%d toolCalls=%d\n",
		routed.ConversationID,
		executed.AssistantMessageID,
		len(executed.Content),
		len(executed.Thinking),
		len(executed.ToolCalls),
	)

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

	g.Log().Debugf(ctx, "im.turn.deliver conv=%s assistantMsg=%s outboundTextLen=%d", routed.ConversationID, executed.AssistantMessageID, len(envelope.Text))
	fmt.Printf("TRACE im.turn.deliver conv=%s assistantMsg=%s outboundTextLen=%d\n", routed.ConversationID, executed.AssistantMessageID, len(envelope.Text))
	return s.deliverOutboundEnvelope(ctx, conn, envelope, executed.AssistantMessageID)
}

func ensureExternalSession(
	ctx context.Context,
	conn Connection,
	routed *RoutedBinding,
	event InboundEvent,
	addr SessionAddress,
	sessionKey string,
) (*ExternalSession, error) {
	return defaultService.ensureExternalSession(ctx, conn, routed, event, addr, sessionKey)
}

func (s *Service) ensureExternalSession(
	ctx context.Context,
	conn Connection,
	routed *RoutedBinding,
	event InboundEvent,
	addr SessionAddress,
	sessionKey string,
) (*ExternalSession, error) {
	current, err := s.repo.GetExternalSessionByKey(ctx, conn.EnterpriseID, sessionKey)
	if err != nil {
		// First-turn web chat may legitimately miss an external session row.
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}
	if current != nil {
		if err := s.touchExternalSession(ctx, current, event); err != nil {
			return nil, err
		}
		return current, nil
	}

	convID, err := s.createExternalConversation(ctx, conn, routed, event)
	if err != nil {
		return nil, err
	}

	now := s.now()
	id := s.idGenerator()
	err = s.repo.CreateExternalSession(ctx, ExternalSessionRecord{
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
		LastMessageAt:    now,
		CreatedAt:        now,
		UpdatedAt:        now,
	})
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

func touchExternalSession(ctx context.Context, session *ExternalSession, event InboundEvent) error {
	return defaultService.touchExternalSession(ctx, session, event)
}

func (s *Service) touchExternalSession(ctx context.Context, session *ExternalSession, event InboundEvent) error {
	now := s.now()
	err := s.repo.TouchExternalSession(ctx, session.ID, now, now)
	if err != nil {
		return err
	}
	session.LastMessageAt = &now
	session.UpdatedAt = now
	return nil
}

func createExternalConversation(ctx context.Context, conn Connection, routed *RoutedBinding, event InboundEvent) (string, error) {
	return defaultService.createExternalConversation(ctx, conn, routed, event)
}

func (s *Service) createExternalConversation(ctx context.Context, conn Connection, routed *RoutedBinding, event InboundEvent) (string, error) {
	if preferredID, ok, err := s.resolvePreferredConversation(ctx, conn, routed, event); err != nil {
		return "", err
	} else if ok {
		return preferredID, nil
	}

	now := s.now()
	id := s.idGenerator()
	title := buildExternalConversationTitle(conn.Platform, event)
	err := s.repo.CreateConversation(ctx, ConversationRecord{
		ID:                 id,
		UserID:             routed.Agent.UserId,
		GroupID:            conn.EnterpriseID,
		AgentID:            routed.Binding.AgentID,
		Title:              title,
		SourceType:         conn.Platform,
		SourceConnectionID: conn.ID,
		SourceChatID:       event.ExternalChatID,
		SourceThreadID:     event.ExternalThreadID,
		SourceUserID:       event.ExternalUserID,
		CreatedAt:          now,
		UpdatedAt:          now,
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

func resolvePreferredConversation(ctx context.Context, conn Connection, routed *RoutedBinding, event InboundEvent) (string, bool, error) {
	return defaultService.resolvePreferredConversation(ctx, conn, routed, event)
}

func (s *Service) resolvePreferredConversation(ctx context.Context, conn Connection, routed *RoutedBinding, event InboundEvent) (string, bool, error) {
	preferredID := str(safeMap(event.ReplyHandle)["conversation_id"])
	if preferredID == "" {
		return "", false, nil
	}

	owned, err := s.conversation.BelongsToUser(preferredID, routed.Agent.UserId, conn.EnterpriseID)
	if err != nil {
		return "", false, err
	}
	if !owned {
		return "", false, nil
	}

	if err := s.annotateConversationSource(ctx, preferredID, conn, event); err != nil {
		return "", false, err
	}
	return preferredID, true, nil
}

func annotateConversationSource(ctx context.Context, conversationID string, conn Connection, event InboundEvent) error {
	return defaultService.annotateConversationSource(ctx, conversationID, conn, event)
}

func (s *Service) annotateConversationSource(ctx context.Context, conversationID string, conn Connection, event InboundEvent) error {
	return s.repo.UpdateConversationSource(ctx, conversationID, g.Map{
		"source_type":          conn.Platform,
		"source_connection_id": nullableUUID(conn.ID),
		"source_chat_id":       event.ExternalChatID,
		"source_thread_id":     event.ExternalThreadID,
		"source_user_id":       event.ExternalUserID,
		"updated_at":           s.now(),
	})
}

func persistInboundConversationMessage(ctx context.Context, conversationID string, event InboundEvent) (string, error) {
	return defaultService.persistInboundConversationMessage(ctx, conversationID, event)
}

func (s *Service) persistInboundConversationMessage(ctx context.Context, conversationID string, event InboundEvent) (string, error) {
	metaJSON, _ := json.Marshal(g.Map{
		"platform":     event.Platform,
		"chat_type":    event.ChatType,
		"mentions_bot": event.MentionsBot,
		"attachments":  event.Attachments,
		"segments":     event.Segments,
		"reply_handle": event.ReplyHandle,
	})

	message, err := s.conversation.SaveMessage(conversationID, "user", event.Text, "", "", "done")
	if err != nil {
		return "", err
	}
	err = s.repo.UpdateInboundMessage(ctx, message.Id, event.MessageID, "received", string(metaJSON))
	if err != nil {
		return "", err
	}
	if err := s.repo.TouchConversation(ctx, conversationID, s.now()); err != nil {
		return "", err
	}
	s.conversation.AutoTitle(conversationID)
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
	return defaultService.deliverOutboundEnvelope(ctx, conn, envelope, messageID)
}

func (s *Service) deliverOutboundEnvelope(ctx context.Context, conn Connection, envelope OutboundEnvelope, messageID string) error {
	requestJSON, _ := json.Marshal(envelope)
	logID := s.idGenerator()
	err := s.repo.InsertDeliveryLog(ctx, DeliveryLogRecord{
		ID:             logID,
		EnterpriseID:   conn.EnterpriseID,
		Platform:       conn.Platform,
		ConnectionID:   conn.ID,
		ConversationID: envelope.ConversationID,
		MessageID:      nullableUUID(messageID),
		Attempt:        1,
		Status:         "pending",
		RequestJSON:    string(requestJSON),
		ResponseJSON:   "{}",
		ErrorMessage:   "",
		CreatedAt:      s.now(),
	})
	if err != nil {
		return err
	}

	adapter, err := s.getAdapter(conn.Platform)
	if err != nil {
		_ = s.updateDeliveryLog(ctx, logID, "error", `{"error":"adapter unavailable"}`, err.Error())
		return err
	}

	if err := adapter.SendOutbound(ctx, conn, envelope); err != nil {
		_ = s.updateDeliveryLog(ctx, logID, "error", `{"error":"send failed"}`, err.Error())
		return err
	}

	err = s.updateDeliveryLog(ctx, logID, "accepted", `{"status":"accepted"}`, "")
	return err
}

func updateDeliveryLog(ctx context.Context, id, status, responseJSON, errorMessage string) error {
	return defaultService.updateDeliveryLog(ctx, id, status, responseJSON, errorMessage)
}

func (s *Service) updateDeliveryLog(ctx context.Context, id, status, responseJSON, errorMessage string) error {
	return s.repo.UpdateDeliveryLog(ctx, id, status, responseJSON, errorMessage)
}

func nullableUUID(id string) any {
	if id == "" {
		return nil
	}
	return id
}
