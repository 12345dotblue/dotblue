package im

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"dotblue/internal/domains/agent"
	"dotblue/internal/domains/chat"
	"dotblue/internal/domains/conversation"
	"dotblue/internal/domains/engine"
	"github.com/gogf/gf/v2/frame/g"
	. "github.com/smartystreets/goconvey/convey"
)

type stubIMAgentDomain struct {
	getByIDFunc       func(id string) (*agent.Agent, error)
	belongsToUserFunc func(id, userID, enterpriseID string) (bool, error)
}

func (s *stubIMAgentDomain) GetById(id string) (*agent.Agent, error) {
	if s.getByIDFunc != nil {
		return s.getByIDFunc(id)
	}
	return nil, nil
}

func (s *stubIMAgentDomain) BelongsToUser(id, userID, enterpriseID string) (bool, error) {
	if s.belongsToUserFunc != nil {
		return s.belongsToUserFunc(id, userID, enterpriseID)
	}
	return false, nil
}

type stubIMChatDomain struct {
	prepareFunc func(ctx context.Context, userID, enterpriseID, agentID, conversationID string) (*chat.PreparedTurn, error)
	executeFunc func(ctx context.Context, prepared *chat.PreparedTurn) (*chat.ExecutedTurn, error)
}

func (s *stubIMChatDomain) PrepareConversationExecution(ctx context.Context, userID, enterpriseID, agentID, conversationID string) (*chat.PreparedTurn, error) {
	if s.prepareFunc != nil {
		return s.prepareFunc(ctx, userID, enterpriseID, agentID, conversationID)
	}
	return nil, nil
}

func (s *stubIMChatDomain) ExecutePreparedTurn(ctx context.Context, prepared *chat.PreparedTurn) (*chat.ExecutedTurn, error) {
	if s.executeFunc != nil {
		return s.executeFunc(ctx, prepared)
	}
	return nil, nil
}

type stubIMConversationDomain struct {
	belongsToUserFunc func(id, userId, enterpriseId string) (bool, error)
	getByIDFunc       func(id string) (*conversation.Conversation, error)
	saveMessageFunc   func(convId, role, content, thinking, toolCallsJson, status string) (*conversation.Message, error)
	autoTitleFunc     func(convId string)
}

func (s *stubIMConversationDomain) BelongsToUser(id, userId, enterpriseId string) (bool, error) {
	if s.belongsToUserFunc != nil {
		return s.belongsToUserFunc(id, userId, enterpriseId)
	}
	return false, nil
}

func (s *stubIMConversationDomain) GetById(id string) (*conversation.Conversation, error) {
	if s.getByIDFunc != nil {
		return s.getByIDFunc(id)
	}
	return nil, nil
}

func (s *stubIMConversationDomain) SaveMessage(convId, role, content, thinking, toolCallsJson, status string) (*conversation.Message, error) {
	if s.saveMessageFunc != nil {
		return s.saveMessageFunc(convId, role, content, thinking, toolCallsJson, status)
	}
	return nil, nil
}

func (s *stubIMConversationDomain) AutoTitle(convId string) {
	if s.autoTitleFunc != nil {
		s.autoTitleFunc(convId)
	}
}

type stubIMRepository struct {
	getExternalSessionByKeyFunc  func(ctx context.Context, enterpriseID, sessionKey string) (*ExternalSession, error)
	createExternalSessionFunc    func(ctx context.Context, record ExternalSessionRecord) error
	createConversationFunc       func(ctx context.Context, record ConversationRecord) error
	touchExternalSessionFunc     func(ctx context.Context, id string, lastMessageAt, updatedAt any) error
	updateConversationSourceFunc func(ctx context.Context, id string, data g.Map) error
	updateInboundMessageFunc     func(ctx context.Context, id, sourceMessageID, deliveryStatus, messageMetaJSON string) error
	touchConversationFunc        func(ctx context.Context, id string, updatedAt any) error
	insertDeliveryLogFunc        func(ctx context.Context, record DeliveryLogRecord) error
	updateDeliveryLogFunc        func(ctx context.Context, id, status, responseJSON, errorMessage string) error
}

func (s *stubIMRepository) GetExternalSessionByKey(ctx context.Context, enterpriseID, sessionKey string) (*ExternalSession, error) {
	if s.getExternalSessionByKeyFunc != nil {
		return s.getExternalSessionByKeyFunc(ctx, enterpriseID, sessionKey)
	}
	return nil, nil
}

func (s *stubIMRepository) CreateExternalSession(ctx context.Context, record ExternalSessionRecord) error {
	if s.createExternalSessionFunc != nil {
		return s.createExternalSessionFunc(ctx, record)
	}
	return nil
}

func (s *stubIMRepository) CreateConversation(ctx context.Context, record ConversationRecord) error {
	if s.createConversationFunc != nil {
		return s.createConversationFunc(ctx, record)
	}
	return nil
}

func (s *stubIMRepository) TouchExternalSession(ctx context.Context, id string, lastMessageAt, updatedAt any) error {
	if s.touchExternalSessionFunc != nil {
		return s.touchExternalSessionFunc(ctx, id, lastMessageAt, updatedAt)
	}
	return nil
}

func (s *stubIMRepository) UpdateConversationSource(ctx context.Context, id string, data g.Map) error {
	if s.updateConversationSourceFunc != nil {
		return s.updateConversationSourceFunc(ctx, id, data)
	}
	return nil
}

func (s *stubIMRepository) UpdateInboundMessage(ctx context.Context, id, sourceMessageID, deliveryStatus, messageMetaJSON string) error {
	if s.updateInboundMessageFunc != nil {
		return s.updateInboundMessageFunc(ctx, id, sourceMessageID, deliveryStatus, messageMetaJSON)
	}
	return nil
}

func (s *stubIMRepository) TouchConversation(ctx context.Context, id string, updatedAt any) error {
	if s.touchConversationFunc != nil {
		return s.touchConversationFunc(ctx, id, updatedAt)
	}
	return nil
}

func (s *stubIMRepository) InsertDeliveryLog(ctx context.Context, record DeliveryLogRecord) error {
	if s.insertDeliveryLogFunc != nil {
		return s.insertDeliveryLogFunc(ctx, record)
	}
	return nil
}

func (s *stubIMRepository) UpdateDeliveryLog(ctx context.Context, id, status, responseJSON, errorMessage string) error {
	if s.updateDeliveryLogFunc != nil {
		return s.updateDeliveryLogFunc(ctx, id, status, responseJSON, errorMessage)
	}
	return nil
}

type stubAdapter struct {
	sendOutboundFunc func(ctx context.Context, conn Connection, msg OutboundEnvelope) error
}

func (s *stubAdapter) Platform() string                                                   { return "stub" }
func (s *stubAdapter) Start(ctx context.Context, conn Connection) error                   { return nil }
func (s *stubAdapter) Stop(ctx context.Context, connectionID string) error                { return nil }
func (s *stubAdapter) ValidateConfig(config map[string]any, secrets map[string]any) error { return nil }
func (s *stubAdapter) ParseInbound(ctx context.Context, raw any) ([]InboundEvent, error) {
	return nil, nil
}
func (s *stubAdapter) SendOutbound(ctx context.Context, conn Connection, msg OutboundEnvelope) error {
	if s.sendOutboundFunc != nil {
		return s.sendOutboundFunc(ctx, conn, msg)
	}
	return nil
}

func TestIMServiceResolvePreferredConversationUsesBoundaries(t *testing.T) {
	Convey("resolvePreferredConversation 通过 conversation 边界验证归属并标记来源", t, func() {
		var updatedSource bool
		service := &Service{
			conversation: &stubIMConversationDomain{
				belongsToUserFunc: func(id, userId, enterpriseId string) (bool, error) {
					So(id, ShouldEqual, "conv-1")
					So(userId, ShouldEqual, "user-1")
					return true, nil
				},
			},
			repo: &stubIMRepository{
				updateConversationSourceFunc: func(ctx context.Context, id string, data g.Map) error {
					updatedSource = true
					So(id, ShouldEqual, "conv-1")
					So(data["source_type"], ShouldEqual, "feishu")
					return nil
				},
			},
			now: func() time.Time { return time.Date(2026, 5, 20, 13, 0, 0, 0, time.UTC) },
		}

		id, ok, err := service.resolvePreferredConversation(context.Background(), Connection{
			EnterpriseID: "ent-1",
			Platform:     "feishu",
			ID:           "conn-1",
		}, &RoutedBinding{
			Binding: AgentBinding{AgentID: "agent-1"},
			Agent:   &agent.Agent{Id: "agent-1", UserId: "user-1"},
		}, InboundEvent{
			ExternalChatID:   "chat-1",
			ExternalThreadID: "thread-1",
			ExternalUserID:   "user-ext-1",
			ReplyHandle: map[string]any{
				"conversation_id": "conv-1",
			},
		})

		So(err, ShouldBeNil)
		So(ok, ShouldBeTrue)
		So(id, ShouldEqual, "conv-1")
		So(updatedSource, ShouldBeTrue)
	})
}

func TestIMServiceEnsureExternalSessionTreatsNoRowsAsFirstTurn(t *testing.T) {
	Convey("ensureExternalSession 在首次会话遇到 sql.ErrNoRows 时自动创建外部会话", t, func() {
		var createdConversation bool
		var createdExternalSession bool
		service := &Service{
			repo: &stubIMRepository{
				getExternalSessionByKeyFunc: func(ctx context.Context, enterpriseID, sessionKey string) (*ExternalSession, error) {
					So(enterpriseID, ShouldEqual, "ent-1")
					So(sessionKey, ShouldEqual, "session-1")
					return nil, sql.ErrNoRows
				},
				createConversationFunc: func(ctx context.Context, record ConversationRecord) error {
					createdConversation = true
					So(record.ID, ShouldEqual, "conv-1")
					So(record.UserID, ShouldEqual, "user-1")
					So(record.GroupID, ShouldEqual, "ent-1")
					So(record.AgentID, ShouldEqual, "agent-1")
					So(record.SourceType, ShouldEqual, "web")
					So(record.SourceConnectionID, ShouldEqual, "conn-1")
					return nil
				},
				createExternalSessionFunc: func(ctx context.Context, record ExternalSessionRecord) error {
					createdExternalSession = true
					So(record.ID, ShouldEqual, "session-row-1")
					So(record.EnterpriseID, ShouldEqual, "ent-1")
					So(record.ConnectionID, ShouldEqual, "conn-1")
					So(record.BindingID, ShouldEqual, "binding-1")
					So(record.AgentID, ShouldEqual, "agent-1")
					So(record.SessionKey, ShouldEqual, "session-1")
					So(record.ConversationID, ShouldEqual, "conv-1")
					return nil
				},
			},
			now: func() time.Time { return time.Date(2026, 5, 21, 14, 40, 0, 0, time.UTC) },
			idGenerator: func() string {
				if !createdConversation {
					return "conv-1"
				}
				return "session-row-1"
			},
		}

		session, err := service.ensureExternalSession(context.Background(), Connection{
			ID:           "conn-1",
			EnterpriseID: "ent-1",
			Platform:     "web",
		}, &RoutedBinding{
			Binding: AgentBinding{ID: "binding-1", AgentID: "agent-1"},
			Agent:   &agent.Agent{Id: "agent-1", UserId: "user-1"},
		}, InboundEvent{
			ExternalChatID: "chat-1",
			ExternalUserID: "user-ext-1",
			Text:           "hello first turn",
		}, SessionAddress{}, "session-1")

		So(err, ShouldBeNil)
		So(createdConversation, ShouldBeTrue)
		So(createdExternalSession, ShouldBeTrue)
		So(session, ShouldNotBeNil)
		So(session.ID, ShouldEqual, "session-row-1")
		So(session.ConversationID, ShouldEqual, "conv-1")
		So(session.SessionKey, ShouldEqual, "session-1")
	})
}

func TestIMServiceExecuteInboundTurnUsesInjectedDependencies(t *testing.T) {
	Convey("ExecuteInboundTurn 只通过 chat/repo/adapter 边界完成回复生成和投递", t, func() {
		var deliveryAccepted bool
		service := &Service{
			chat: &stubIMChatDomain{
				prepareFunc: func(ctx context.Context, userID, enterpriseID, agentID, conversationID string) (*chat.PreparedTurn, error) {
					return &chat.PreparedTurn{
						Agent:          &agent.Agent{Id: agentID},
						Endpoint:       &engine.AgentEndpoint{URL: "http://runtime"},
						ConversationID: conversationID,
					}, nil
				},
				executeFunc: func(ctx context.Context, prepared *chat.PreparedTurn) (*chat.ExecutedTurn, error) {
					return &chat.ExecutedTurn{
						Content:            "hello from assistant",
						AssistantMessageID: "assistant-msg-1",
					}, nil
				},
			},
			repo: &stubIMRepository{
				insertDeliveryLogFunc: func(ctx context.Context, record DeliveryLogRecord) error {
					So(record.Status, ShouldEqual, "pending")
					return nil
				},
				updateDeliveryLogFunc: func(ctx context.Context, id, status, responseJSON, errorMessage string) error {
					if status == "accepted" {
						deliveryAccepted = true
					}
					return nil
				},
			},
			getAdapter: func(platform string) (Adapter, error) {
				return &stubAdapter{
					sendOutboundFunc: func(ctx context.Context, conn Connection, msg OutboundEnvelope) error {
						So(msg.Text, ShouldEqual, "hello from assistant")
						So(msg.ConversationID, ShouldEqual, "conv-1")
						return nil
					},
				}, nil
			},
			now:         func() time.Time { return time.Date(2026, 5, 20, 13, 0, 0, 0, time.UTC) },
			idGenerator: func() string { return "log-1" },
		}

		err := service.ExecuteInboundTurn(context.Background(), Connection{
			ID:           "conn-1",
			EnterpriseID: "ent-1",
			Platform:     "stub",
		}, &RoutedInboundSession{
			Binding:        AgentBinding{AgentID: "agent-1"},
			AgentID:        "agent-1",
			AgentUserID:    "user-1",
			ConversationID: "conv-1",
			MessageID:      "msg-1",
			ExternalSession: &ExternalSession{
				ConversationID: "conv-1",
			},
		}, InboundEvent{
			ExternalChatID: "chat-1",
		})

		So(err, ShouldBeNil)
		So(deliveryAccepted, ShouldBeTrue)
	})

	Convey("deliverOutboundEnvelope 在 adapter 失败时透传错误并写回失败状态", t, func() {
		var errorLogged bool
		service := &Service{
			repo: &stubIMRepository{
				insertDeliveryLogFunc: func(ctx context.Context, record DeliveryLogRecord) error { return nil },
				updateDeliveryLogFunc: func(ctx context.Context, id, status, responseJSON, errorMessage string) error {
					if status == "error" {
						errorLogged = true
					}
					return nil
				},
			},
			getAdapter: func(platform string) (Adapter, error) {
				return &stubAdapter{
					sendOutboundFunc: func(ctx context.Context, conn Connection, msg OutboundEnvelope) error {
						return errors.New("send failed")
					},
				}, nil
			},
			now:         func() time.Time { return time.Date(2026, 5, 20, 13, 0, 0, 0, time.UTC) },
			idGenerator: func() string { return "log-1" },
		}

		err := service.deliverOutboundEnvelope(context.Background(), Connection{
			ID:           "conn-1",
			EnterpriseID: "ent-1",
			Platform:     "stub",
		}, OutboundEnvelope{
			ConversationID: "conv-1",
			Text:           "reply",
		}, "assistant-msg-1")

		So(err, ShouldNotBeNil)
		So(errorLogged, ShouldBeTrue)
	})
}
