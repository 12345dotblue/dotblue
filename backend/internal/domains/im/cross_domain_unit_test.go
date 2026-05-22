package im

import (
	"context"
	"testing"

	"dotblue/internal/domains/agent"
	"dotblue/internal/domains/conversation"
	. "github.com/smartystreets/goconvey/convey"
)

type stubBindingResolverRepository struct {
	listActiveBindingsByConnectionFunc func(ctx context.Context, enterpriseID, connectionID string) ([]bindingRecord, error)
}

func (s *stubBindingResolverRepository) ListActiveBindingsByConnection(ctx context.Context, enterpriseID, connectionID string) ([]bindingRecord, error) {
	if s.listActiveBindingsByConnectionFunc != nil {
		return s.listActiveBindingsByConnectionFunc(ctx, enterpriseID, connectionID)
	}
	return nil, nil
}

func TestResolveInboundBindingUsesInjectedAgentBoundary(t *testing.T) {
	Convey("resolveInboundBindingWith 通过本地 agent 边界加载路由目标", t, func() {
		lookedUp := false
		result, err := resolveInboundBindingWith(context.Background(), &stubBindingResolverRepository{
			listActiveBindingsByConnectionFunc: func(ctx context.Context, enterpriseID, connectionID string) ([]bindingRecord, error) {
				So(enterpriseID, ShouldEqual, "ent-1")
				So(connectionID, ShouldEqual, "conn-1")
				return []bindingRecord{{
					ID:              "binding-1",
					EnterpriseID:    "ent-1",
					AgentID:         "agent-1",
					ConnectionID:    "conn-1",
					Status:          StatusActive,
					TriggerMode:     TriggerModeMentionOnly,
					SessionStrategy: SessionStrategyPerChatPerUser,
					ReplyMode:       defaultReplyMode,
					AllowGroup:      true,
					AllowDM:         true,
					Priority:        100,
				}}, nil
			},
		}, &stubIMAgentDomain{
			getByIDFunc: func(id string) (*agent.Agent, error) {
				lookedUp = true
				So(id, ShouldEqual, "agent-1")
				return &agent.Agent{Id: id, GroupId: "ent-1"}, nil
			},
		}, Connection{
			ID:           "conn-1",
			EnterpriseID: "ent-1",
			Platform:     "feishu",
		}, InboundEvent{
			ChatType:    "group",
			MentionsBot: true,
			Text:        "hello",
		})

		So(err, ShouldBeNil)
		So(lookedUp, ShouldBeTrue)
		So(result, ShouldNotBeNil)
		So(result.Binding.ID, ShouldEqual, "binding-1")
		So(result.Agent, ShouldNotBeNil)
		So(result.Agent.Id, ShouldEqual, "agent-1")
	})
}

func TestValidateWebChatOwnershipUsesServiceBoundaries(t *testing.T) {
	Convey("validateWebChatOwnership 通过 im 本地边界校验 agent 和 conversation 归属", t, func() {
		previous := defaultService
		t.Cleanup(func() {
			defaultService = previous
		})

		defaultService = &Service{
			agents: &stubIMAgentDomain{
				belongsToUserFunc: func(id, userID, enterpriseID string) (bool, error) {
					So(id, ShouldEqual, "agent-1")
					So(userID, ShouldEqual, "user-1")
					So(enterpriseID, ShouldEqual, "ent-1")
					return true, nil
				},
			},
			conversation: &stubIMConversationDomain{
				belongsToUserFunc: func(id, userId, enterpriseId string) (bool, error) {
					So(id, ShouldEqual, "conv-1")
					return true, nil
				},
				getByIDFunc: func(id string) (*conversation.Conversation, error) {
					return &conversation.Conversation{Id: id, AgentId: "agent-1"}, nil
				},
			},
		}

		err := validateWebChatOwnership("conv-1", "agent-1", "user-1", "ent-1")

		So(err, ShouldBeNil)
	})
}
