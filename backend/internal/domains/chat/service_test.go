package chat

import (
	"context"
	"errors"
	"strings"
	"testing"

	"dotblue/internal/domains/agent"
	"dotblue/internal/domains/conversation"
	"dotblue/internal/domains/engine"
	. "github.com/smartystreets/goconvey/convey"
)

type stubAgentDomain struct {
	belongsToUserFunc func(id, userID, enterpriseID string) (bool, error)
	getByIdFunc       func(id string) (*agent.Agent, error)
}

func (s *stubAgentDomain) BelongsToUser(id, userID, enterpriseID string) (bool, error) {
	if s.belongsToUserFunc != nil {
		return s.belongsToUserFunc(id, userID, enterpriseID)
	}
	return false, nil
}

func (s *stubAgentDomain) GetById(id string) (*agent.Agent, error) {
	if s.getByIdFunc != nil {
		return s.getByIdFunc(id)
	}
	return nil, nil
}

type stubConversationDomain struct {
	belongsToUserFunc func(id, userID, enterpriseID string) (bool, error)
	createFunc        func(userID, enterpriseID, agentID, title string) (*conversation.Conversation, error)
	getByIdFunc       func(id string) (*conversation.Conversation, error)
	saveMessageFunc   func(convID, role, content, thinking, toolCallsJSON, status string) (*conversation.Message, error)
	touchUpdatedFunc  func(id string) error
	autoTitleFunc     func(convID string)
	listMessagesFunc  func(convID, before string, limit int) ([]*conversation.MessagePublic, error)
}

func (s *stubConversationDomain) BelongsToUser(id, userID, enterpriseID string) (bool, error) {
	if s.belongsToUserFunc != nil {
		return s.belongsToUserFunc(id, userID, enterpriseID)
	}
	return false, nil
}

func (s *stubConversationDomain) Create(userID, enterpriseID, agentID, title string) (*conversation.Conversation, error) {
	if s.createFunc != nil {
		return s.createFunc(userID, enterpriseID, agentID, title)
	}
	return nil, nil
}

func (s *stubConversationDomain) GetById(id string) (*conversation.Conversation, error) {
	if s.getByIdFunc != nil {
		return s.getByIdFunc(id)
	}
	return nil, nil
}

func (s *stubConversationDomain) SaveMessage(convID, role, content, thinking, toolCallsJSON, status string) (*conversation.Message, error) {
	if s.saveMessageFunc != nil {
		return s.saveMessageFunc(convID, role, content, thinking, toolCallsJSON, status)
	}
	return nil, nil
}

func (s *stubConversationDomain) TouchUpdated(id string) error {
	if s.touchUpdatedFunc != nil {
		return s.touchUpdatedFunc(id)
	}
	return nil
}

func (s *stubConversationDomain) AutoTitle(convID string) {
	if s.autoTitleFunc != nil {
		s.autoTitleFunc(convID)
	}
}

func (s *stubConversationDomain) ListMessages(convID, before string, limit int) ([]*conversation.MessagePublic, error) {
	if s.listMessagesFunc != nil {
		return s.listMessagesFunc(convID, before, limit)
	}
	return nil, nil
}

type stubRuntimeDomain struct {
	ensureRunningFunc func(ctx context.Context, orgID, userID, agentID string) (*engine.AgentEndpoint, error)
}

func (s *stubRuntimeDomain) EnsureRunning(ctx context.Context, orgID, userID, agentID string) (*engine.AgentEndpoint, error) {
	if s.ensureRunningFunc != nil {
		return s.ensureRunningFunc(ctx, orgID, userID, agentID)
	}
	return nil, nil
}

func TestCollectEngineStreamParsesSSE(t *testing.T) {
	body := "data: {\"choices\":[{\"delta\":{\"reasoning_content\":\"thinking\"}}]}\n\n" +
		"data: {\"choices\":[{\"delta\":{\"content\":\"assistant reply\"}}]}\n\n" +
		"data: [DONE]\n\n"

	content, thinking, toolCalls, err := collectEngineStream(strings.NewReader(body))
	if err != nil {
		t.Fatalf("collectEngineStream() error = %v", err)
	}
	if content != "assistant reply" {
		t.Fatalf("content = %q, want assistant reply", content)
	}
	if thinking != "thinking" {
		t.Fatalf("thinking = %q, want thinking", thinking)
	}
	if len(toolCalls) != 0 {
		t.Fatalf("toolCalls len = %d, want 0", len(toolCalls))
	}
}

func TestCollectEngineStreamFallsBackToJSONResponse(t *testing.T) {
	body := `{"id":"chatcmpl-test","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":"assistant reply"}}]}`

	content, thinking, toolCalls, err := collectEngineStream(strings.NewReader(body))
	if err != nil {
		t.Fatalf("collectEngineStream() error = %v", err)
	}
	if content != "assistant reply" {
		t.Fatalf("content = %q, want assistant reply", content)
	}
	if thinking != "" {
		t.Fatalf("thinking = %q, want empty", thinking)
	}
	if len(toolCalls) != 0 {
		t.Fatalf("toolCalls len = %d, want 0", len(toolCalls))
	}
}

func TestServicePrepareTurnUsesInjectedDomains(t *testing.T) {
	Convey("PrepareTurn 通过本地接口编排其他模块，而不是直接依赖真实模块实现", t, func() {
		autoTitleCalled := false
		service := &Service{
			agents: &stubAgentDomain{
				belongsToUserFunc: func(id, userID, enterpriseID string) (bool, error) {
					So(id, ShouldEqual, "agent-1")
					return true, nil
				},
				getByIdFunc: func(id string) (*agent.Agent, error) {
					return &agent.Agent{Id: id, EngineType: "hermes"}, nil
				},
			},
			conversations: &stubConversationDomain{
				createFunc: func(userID, enterpriseID, agentID, title string) (*conversation.Conversation, error) {
					So(userID, ShouldEqual, "user-1")
					So(enterpriseID, ShouldEqual, "ent-1")
					So(agentID, ShouldEqual, "agent-1")
					return &conversation.Conversation{Id: "conv-1"}, nil
				},
				saveMessageFunc: func(convID, role, content, thinking, toolCallsJSON, status string) (*conversation.Message, error) {
					So(convID, ShouldEqual, "conv-1")
					So(role, ShouldEqual, "user")
					So(content, ShouldEqual, "hello")
					return &conversation.Message{Id: "msg-1"}, nil
				},
				touchUpdatedFunc: func(id string) error {
					So(id, ShouldEqual, "conv-1")
					return nil
				},
				autoTitleFunc: func(convID string) {
					autoTitleCalled = true
					So(convID, ShouldEqual, "conv-1")
				},
				listMessagesFunc: func(convID, before string, limit int) ([]*conversation.MessagePublic, error) {
					So(convID, ShouldEqual, "conv-1")
					So(limit, ShouldEqual, 20)
					return []*conversation.MessagePublic{
						{Role: "user", Content: "hello"},
					}, nil
				},
			},
			runtime: &stubRuntimeDomain{
				ensureRunningFunc: func(ctx context.Context, orgID, userID, agentID string) (*engine.AgentEndpoint, error) {
					So(orgID, ShouldEqual, "ent-1")
					So(userID, ShouldEqual, "user-1")
					So(agentID, ShouldEqual, "agent-1")
					return &engine.AgentEndpoint{URL: "http://runtime"}, nil
				},
			},
		}

		prepared, err := service.PrepareTurn(context.Background(), TurnRequest{
			UserID:             "user-1",
			EnterpriseID:       "ent-1",
			AgentID:            "agent-1",
			Content:            "hello",
			CreateConversation: true,
		})

		So(err, ShouldBeNil)
		So(prepared, ShouldNotBeNil)
		So(prepared.ConversationID, ShouldEqual, "conv-1")
		So(prepared.Endpoint.URL, ShouldEqual, "http://runtime")
		So(prepared.History, ShouldHaveLength, 1)
		So(autoTitleCalled, ShouldBeTrue)
	})
}

func TestServicePersistAssistantTurnUsesConversationBoundary(t *testing.T) {
	Convey("PersistAssistantTurnWithMessageID 只依赖 conversation 边界，便于单元测试", t, func() {
		var savedToolCallsJSON string
		service := &Service{
			conversations: &stubConversationDomain{
				saveMessageFunc: func(convID, role, content, thinking, toolCallsJSON, status string) (*conversation.Message, error) {
					savedToolCallsJSON = toolCallsJSON
					So(role, ShouldEqual, "assistant")
					So(status, ShouldEqual, "done")
					return &conversation.Message{Id: "assistant-msg-1"}, nil
				},
				touchUpdatedFunc: func(id string) error {
					So(id, ShouldEqual, "conv-1")
					return nil
				},
			},
		}

		messageID, err := service.PersistAssistantTurnWithMessageID("conv-1", "reply", "thinking", []conversation.ToolCallItem{
			{Tool: "search", Emoji: "S", Label: "Search", Status: "done"},
		})

		So(err, ShouldBeNil)
		So(messageID, ShouldEqual, "assistant-msg-1")
		So(savedToolCallsJSON, ShouldContainSubstring, "\"tool\":\"search\"")
	})

	Convey("PersistAssistantTurnWithMessageID 透传 conversation 边界错误", t, func() {
		service := &Service{
			conversations: &stubConversationDomain{
				saveMessageFunc: func(convID, role, content, thinking, toolCallsJSON, status string) (*conversation.Message, error) {
					return nil, errors.New("save failed")
				},
			},
		}

		_, err := service.PersistAssistantTurnWithMessageID("conv-1", "reply", "", nil)

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "save failed")
	})
}

func TestServiceConversationTitleUsesConversationBoundary(t *testing.T) {
	Convey("ConversationTitle 通过 conversation 边界读取标题", t, func() {
		service := &Service{
			conversations: &stubConversationDomain{
				getByIdFunc: func(id string) (*conversation.Conversation, error) {
					So(id, ShouldEqual, "conv-1")
					return &conversation.Conversation{Id: id, Title: "hello"}, nil
				},
			},
		}

		title, err := service.ConversationTitle("conv-1")

		So(err, ShouldBeNil)
		So(title, ShouldEqual, "hello")
	})
}

func TestServiceResolveEngineUsesInjectedFactory(t *testing.T) {
	Convey("ResolveEngine 通过注入的 engine factory 查找引擎", t, func() {
		service := &Service{
			getEngine: func(name string) (engine.Engine, error) {
				So(name, ShouldEqual, "custom")
				return nil, nil
			},
		}

		eng, engineType, err := service.ResolveEngine(&agent.Agent{EngineType: "custom"})

		So(err, ShouldBeNil)
		So(engineType, ShouldEqual, "custom")
		So(eng, ShouldBeNil)
	})
}
