package conversation

import (
	"errors"
	"testing"
	"time"

	"dotblue/internal/domains/agent"
	. "github.com/smartystreets/goconvey/convey"
)

type stubRepository struct {
	createFunc              func(conversation *Conversation) error
	listByUserIdFunc        func(userId, enterpriseId, cursor string, limit int) ([]*Conversation, error)
	getByIdFunc             func(id string) (*Conversation, error)
	belongsToUserFunc       func(id, userId, enterpriseId string) (bool, error)
	updateTitleFunc         func(id, title string, updatedAt time.Time) error
	touchUpdatedFunc        func(id string, updatedAt time.Time) error
	deleteFunc              func(id string) error
	saveMessageFunc         func(message *Message) error
	getLatestMessageFunc    func(convId string) (*Message, error)
	listMessagesFunc        func(convId, before string, limit int) ([]*Message, error)
	getFirstUserMessageFunc func(convId string) (string, error)
}

func (s *stubRepository) Create(conversation *Conversation) error {
	if s.createFunc != nil {
		return s.createFunc(conversation)
	}
	return nil
}

func (s *stubRepository) ListByUserId(userId, enterpriseId, cursor string, limit int) ([]*Conversation, error) {
	if s.listByUserIdFunc != nil {
		return s.listByUserIdFunc(userId, enterpriseId, cursor, limit)
	}
	return nil, nil
}

func (s *stubRepository) GetById(id string) (*Conversation, error) {
	if s.getByIdFunc != nil {
		return s.getByIdFunc(id)
	}
	return nil, nil
}

func (s *stubRepository) BelongsToUser(id, userId, enterpriseId string) (bool, error) {
	if s.belongsToUserFunc != nil {
		return s.belongsToUserFunc(id, userId, enterpriseId)
	}
	return false, nil
}

func (s *stubRepository) UpdateTitle(id, title string, updatedAt time.Time) error {
	if s.updateTitleFunc != nil {
		return s.updateTitleFunc(id, title, updatedAt)
	}
	return nil
}

func (s *stubRepository) TouchUpdated(id string, updatedAt time.Time) error {
	if s.touchUpdatedFunc != nil {
		return s.touchUpdatedFunc(id, updatedAt)
	}
	return nil
}

func (s *stubRepository) Delete(id string) error {
	if s.deleteFunc != nil {
		return s.deleteFunc(id)
	}
	return nil
}

func (s *stubRepository) SaveMessage(message *Message) error {
	if s.saveMessageFunc != nil {
		return s.saveMessageFunc(message)
	}
	return nil
}

func (s *stubRepository) GetLatestMessage(convId string) (*Message, error) {
	if s.getLatestMessageFunc != nil {
		return s.getLatestMessageFunc(convId)
	}
	return nil, nil
}

func (s *stubRepository) ListMessages(convId, before string, limit int) ([]*Message, error) {
	if s.listMessagesFunc != nil {
		return s.listMessagesFunc(convId, before, limit)
	}
	return nil, nil
}

func (s *stubRepository) GetFirstUserMessage(convId string) (string, error) {
	if s.getFirstUserMessageFunc != nil {
		return s.getFirstUserMessageFunc(convId)
	}
	return "", nil
}

type stubAgentDomain struct {
	belongsToUserFunc func(id, userId, enterpriseId string) (bool, error)
	getByIdFunc       func(id string) (*agent.Agent, error)
}

func (s *stubAgentDomain) BelongsToUser(id, userId, enterpriseId string) (bool, error) {
	if s.belongsToUserFunc != nil {
		return s.belongsToUserFunc(id, userId, enterpriseId)
	}
	return false, nil
}

func (s *stubAgentDomain) GetById(id string) (*agent.Agent, error) {
	if s.getByIdFunc != nil {
		return s.getByIdFunc(id)
	}
	return nil, nil
}

func TestServiceCreateGeneratesID(t *testing.T) {
	Convey("Create 应生成会话 ID 并返回持久化后的记录", t, func() {
		var created *Conversation
		repo := &stubRepository{
			createFunc: func(conversation *Conversation) error {
				created = conversation
				return nil
			},
			getByIdFunc: func(id string) (*Conversation, error) {
				return &Conversation{
					Id:      id,
					UserId:  "user-1",
					GroupId: "group-1",
					AgentId: "agent-1",
					Title:   "hello",
				}, nil
			},
		}
		service := NewService(repo)
		service.idGenerator = func() string { return "conv-123" }

		got, err := service.Create("user-1", "group-1", "agent-1", "hello")

		So(err, ShouldBeNil)
		So(created, ShouldNotBeNil)
		So(created.Id, ShouldEqual, "conv-123")
		So(got, ShouldNotBeNil)
		So(got.Id, ShouldEqual, "conv-123")
	})
}

func TestServiceListMessagesChronological(t *testing.T) {
	Convey("ListMessages 应按时间正序返回并解析 toolCalls", t, func() {
		ts1 := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
		ts2 := ts1.Add(time.Minute)
		repo := &stubRepository{
			listMessagesFunc: func(convId, before string, limit int) ([]*Message, error) {
				So(limit, ShouldEqual, defaultMessageListLimit)
				return []*Message{
					{
						Id:        "msg-2",
						Role:      "assistant",
						Content:   "later",
						Thinking:  "reasoning",
						ToolCalls: `[{"tool":"search","emoji":"S","label":"Search","status":"done"}]`,
						Status:    "done",
						CreatedAt: ts2,
					},
					{
						Id:        "msg-1",
						Role:      "user",
						Content:   "earlier",
						Status:    "done",
						CreatedAt: ts1,
					},
				}, nil
			},
		}
		service := NewService(repo)

		got, err := service.ListMessages("conv-1", "", 0)

		So(err, ShouldBeNil)
		So(got, ShouldHaveLength, 2)
		So(got[0].Id, ShouldEqual, "msg-1")
		So(got[1].Id, ShouldEqual, "msg-2")
		So(got[1].ToolCalls, ShouldHaveLength, 1)
		So(got[1].ToolCalls[0].Tool, ShouldEqual, "search")
	})
}

func TestServiceAutoTitle(t *testing.T) {
	Convey("AutoTitle 只在标题为空时回填首条用户消息，并按规则截断", t, func() {
		expectedTime := time.Date(2026, 5, 20, 11, 0, 0, 0, time.UTC)
		var updatedTitle string
		var updatedAt time.Time
		repo := &stubRepository{
			getByIdFunc: func(id string) (*Conversation, error) {
				return &Conversation{Id: id, Title: ""}, nil
			},
			getFirstUserMessageFunc: func(convId string) (string, error) {
				return "12345678901234567890123456789012345678901234567890-extra", nil
			},
			updateTitleFunc: func(id, title string, at time.Time) error {
				updatedTitle = title
				updatedAt = at
				return nil
			},
		}
		service := NewService(repo)
		service.now = func() time.Time { return expectedTime }

		service.AutoTitle("conv-1")

		So(updatedTitle, ShouldEqual, "12345678901234567890123456789012345678901234567890...")
		So(updatedAt, ShouldResemble, expectedTime)
	})

	Convey("AutoTitle 按 rune 截断，避免多字节字符被切坏", t, func() {
		var updatedTitle string
		repo := &stubRepository{
			getByIdFunc: func(id string) (*Conversation, error) {
				return &Conversation{Id: id, Title: ""}, nil
			},
			getFirstUserMessageFunc: func(convId string) (string, error) {
				return "请简短回复：socket 权限修复后的端到端验证。更多内容用于触发安全截断", nil
			},
			updateTitleFunc: func(id, title string, at time.Time) error {
				updatedTitle = title
				return nil
			},
		}

		NewService(repo).AutoTitle("conv-utf8")

		So(updatedTitle, ShouldEqual, "请简短回复：socket 权限修复后的端到端验证。更多内容用于触发安全截断")
	})

	Convey("AutoTitle 在已有标题时不应更新", t, func() {
		called := false
		repo := &stubRepository{
			getByIdFunc: func(id string) (*Conversation, error) {
				return &Conversation{Id: id, Title: "existing"}, nil
			},
			updateTitleFunc: func(id, title string, updatedAt time.Time) error {
				called = true
				return nil
			},
		}

		NewService(repo).AutoTitle("conv-1")

		So(called, ShouldBeFalse)
	})
}

func TestServiceSaveMessagePropagatesRepositoryError(t *testing.T) {
	Convey("SaveMessage 应透传仓储错误，保持单测不依赖真实数据库", t, func() {
		repo := &stubRepository{
			saveMessageFunc: func(message *Message) error {
				return errors.New("insert failed")
			},
		}
		service := NewService(repo)

		_, err := service.SaveMessage("conv-1", "user", "hello", "", "", "done")

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "insert failed")
	})
}

func TestServiceCreatePublicForUserUsesAgentBoundary(t *testing.T) {
	Convey("CreatePublicForUser 通过本地 agent 边界校验归属并补齐 agentName", t, func() {
		service := NewService(&stubRepository{
			createFunc: func(conversation *Conversation) error { return nil },
			getByIdFunc: func(id string) (*Conversation, error) {
				return &Conversation{
					Id:      id,
					UserId:  "user-1",
					GroupId: "ent-1",
					AgentId: "agent-1",
					Title:   "hello",
				}, nil
			},
		})
		service.idGenerator = func() string { return "conv-1" }
		service.agents = &stubAgentDomain{
			belongsToUserFunc: func(id, userId, enterpriseId string) (bool, error) {
				So(id, ShouldEqual, "agent-1")
				So(userId, ShouldEqual, "user-1")
				So(enterpriseId, ShouldEqual, "ent-1")
				return true, nil
			},
			getByIdFunc: func(id string) (*agent.Agent, error) {
				return &agent.Agent{Id: id, AgentName: "Support Agent"}, nil
			},
		}

		got, err := service.CreatePublicForUser("user-1", "ent-1", "agent-1", "hello")

		So(err, ShouldBeNil)
		So(got, ShouldNotBeNil)
		So(got.Id, ShouldEqual, "conv-1")
		So(got.AgentName, ShouldEqual, "Support Agent")
	})
}

func TestServiceGetPublicForUserRejectsMissingConversation(t *testing.T) {
	Convey("GetPublicForUser 在不属于用户时返回领域错误", t, func() {
		service := NewService(&stubRepository{
			belongsToUserFunc: func(id, userId, enterpriseId string) (bool, error) {
				return false, nil
			},
		})

		got, err := service.GetPublicForUser("conv-1", "user-1", "ent-1")

		So(got, ShouldBeNil)
		So(err, ShouldEqual, ErrConversationNotFound)
	})
}
