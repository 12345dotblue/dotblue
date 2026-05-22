package im

import (
	"context"
	"testing"
	"time"

	"dotblue/internal/domains/agent"
	"github.com/bytedance/mockey"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBindingServiceCreateBindingUsesDefaults(t *testing.T) {
	mockey.PatchConvey("CreateBinding should normalize defaults and persist through repository", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := NewMockbindingRepository(ctrl)
		connections := NewMockbindingConnectionReader(ctrl)
		service := &BindingService{
			repo:        repo,
			connections: connections,
			agents: &stubIMAgentDomain{
				getByIDFunc: func(id string) (*agent.Agent, error) {
					return &agent.Agent{Id: id, GroupId: "ent-1", AgentName: "Support Agent"}, nil
				},
			},
		}

		ctx := context.Background()
		fixedTime := time.Date(2026, 5, 16, 22, 30, 0, 0, time.UTC)
		mockey.Mock(time.Now).Return(fixedTime).Build()
		mockey.Mock(uuid.NewString).Return("binding-1").Build()

		connections.EXPECT().
			GetConnection(gomock.Any(), "ent-1", "conn-1").
			Return(Connection{ID: "conn-1", EnterpriseID: "ent-1", Platform: "feishu"}, nil)

		repo.EXPECT().
			CreateBinding(gomock.Any(), gomock.AssignableToTypeOf(bindingRecord{})).
			DoAndReturn(func(_ context.Context, record bindingRecord) error {
				So(record.ID, ShouldEqual, "binding-1")
				So(record.EnterpriseID, ShouldEqual, "ent-1")
				So(record.ConnectionID, ShouldEqual, "conn-1")
				So(record.AgentID, ShouldEqual, "agent-1")
				So(record.Status, ShouldEqual, StatusActive)
				So(record.TriggerMode, ShouldEqual, TriggerModeMentionOnly)
				So(record.TriggerConfigJSON, ShouldEqual, "{}")
				So(record.SessionStrategy, ShouldEqual, SessionStrategyPerChatPerUser)
				So(record.ReplyMode, ShouldEqual, defaultReplyMode)
				So(record.AllowGroup, ShouldBeTrue)
				So(record.AllowDM, ShouldBeTrue)
				So(record.Priority, ShouldEqual, 100)
				So(record.CreatedAt, ShouldEqual, fixedTime)
				So(record.UpdatedAt, ShouldEqual, fixedTime)
				return nil
			})

		repo.EXPECT().
			GetBinding(gomock.Any(), "ent-1", "binding-1").
			Return(&bindingRecord{
				ID:                "binding-1",
				EnterpriseID:      "ent-1",
				AgentID:           "agent-1",
				ConnectionID:      "conn-1",
				Status:            StatusActive,
				TriggerMode:       TriggerModeMentionOnly,
				TriggerConfigJSON: "{}",
				SessionStrategy:   SessionStrategyPerChatPerUser,
				ReplyMode:         defaultReplyMode,
				AllowGroup:        true,
				AllowDM:           true,
				Priority:          100,
				CreatedAt:         fixedTime,
				UpdatedAt:         fixedTime,
			}, nil)

		binding, err := service.CreateBinding(ctx, "ent-1", "conn-1", createBindingReq{
			AgentID: "agent-1",
		})
		So(err, ShouldBeNil)
		So(binding.ID, ShouldEqual, "binding-1")
		So(binding.TriggerMode, ShouldEqual, TriggerModeMentionOnly)
		So(binding.SessionStrategy, ShouldEqual, SessionStrategyPerChatPerUser)
	})
}

func TestBindingServiceUpdateBindingRejectsCrossEnterpriseAgent(t *testing.T) {
	mockey.PatchConvey("UpdateBinding should reject agent outside enterprise before repository update", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := NewMockbindingRepository(ctrl)
		service := &BindingService{
			repo:        repo,
			connections: NewMockbindingConnectionReader(ctrl),
			agents: &stubIMAgentDomain{
				getByIDFunc: func(id string) (*agent.Agent, error) {
					return &agent.Agent{Id: id, GroupId: "other-ent"}, nil
				},
			},
		}

		repo.EXPECT().
			GetBinding(gomock.Any(), "ent-1", "binding-1").
			Return(&bindingRecord{
				ID:                "binding-1",
				EnterpriseID:      "ent-1",
				AgentID:           "agent-1",
				ConnectionID:      "conn-1",
				Status:            StatusActive,
				TriggerMode:       TriggerModeMentionOnly,
				TriggerConfigJSON: "{}",
				SessionStrategy:   SessionStrategyPerChatPerUser,
				ReplyMode:         defaultReplyMode,
				AllowGroup:        true,
				AllowDM:           true,
				Priority:          100,
			}, nil)

		binding, err := service.UpdateBinding(context.Background(), "ent-1", "binding-1", updateBindingReq{
			AgentID: "agent-2",
		})
		So(err, ShouldEqual, ErrInvalidBindingConfig)
		So(binding, ShouldResemble, AgentBinding{})
	})
}
