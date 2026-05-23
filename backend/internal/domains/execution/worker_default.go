package execution

import (
	"context"
	"time"

	"dotblue/internal/domains/chat"
	"dotblue/internal/domains/dataplane"
	"dotblue/internal/domains/session"
	"github.com/gogf/gf/v2/frame/g"
)

type sessionAdapter struct {
	svc    *session.Service
	metaTTL time.Duration
}

func (a *sessionAdapter) ValidateAssignment(ctx context.Context, sessionKey, workerID string, fenceToken int64) (bool, error) {
	return a.svc.ValidateAssignment(ctx, sessionKey, workerID, fenceToken)
}

func (a *sessionAdapter) TryEnterGate(ctx context.Context, sessionKey, holder string) (bool, error) {
	return a.svc.TryEnterGate(ctx, sessionKey, holder)
}

func (a *sessionAdapter) LeaveGate(ctx context.Context, sessionKey string) error {
	return a.svc.LeaveGate(ctx, sessionKey)
}

func (a *sessionAdapter) Heartbeat(ctx context.Context, workerID string, activeTurns, maxConcurrentTurns int64) error {
	return a.svc.HeartbeatWorker(ctx, session.WorkerMeta{
		WorkerID:           workerID,
		ActiveTurns:        activeTurns,
		MaxConcurrentTurns: maxConcurrentTurns,
	}, a.metaTTL)
}

type chatAdapter struct{}

func (a *chatAdapter) Execute(ctx context.Context, task dataplane.TurnTask) (*ChatResult, error) {
	prepared, err := chat.PrepareConversationExecution(ctx, task.UserID, task.EnterpriseID, task.AgentID, task.ConversationID)
	if err != nil {
		return nil, err
	}
	executed, err := chat.ExecutePreparedTurn(ctx, prepared)
	if err != nil {
		return nil, err
	}
	title, _ := chat.ConversationTitle(executed.ConversationID)
	return &ChatResult{
		ConversationID: executed.ConversationID,
		Thinking:       executed.Thinking,
		Content:        executed.Content,
		Title:          title,
	}, nil
}

func Default(ctx context.Context) (*Worker, error) {
	dp, err := dataplane.Default(ctx)
	if err != nil {
		return nil, err
	}
	ss, err := session.Default(ctx)
	if err != nil {
		return nil, err
	}
	cfg := Config{
		WorkerID:           g.Cfg().MustGet(ctx, "worker.id").String(),
		MaxConcurrentTurns: 1,
		HeartbeatInterval:  g.Cfg().MustGet(ctx, "worker.heartbeatInterval").Duration(),
		MetaTTL:            g.Cfg().MustGet(ctx, "worker.metaTTL").Duration(),
		InboxTTL:           g.Cfg().MustGet(ctx, "worker.inboxTTL").Duration(),
		ClaimBlock:         g.Cfg().MustGet(ctx, "worker.claimBlock").Duration(),
		TaskTTL:            time.Hour,
	}
	queue := dataplane.NewRedisTaskQueue(dp, cfg.InboxTTL)
	events := dataplane.NewRedisEventBus(dp)
	reqState := dataplane.NewRedisRequestStateStore(dp, dp.RunningTTL(), dp.FinalTTL())
	outbox := dataplane.NewRedisIMOutbox(dp, cfg.InboxTTL)
	return NewWorker(
		cfg,
		&sessionAdapter{svc: ss, metaTTL: cfg.MetaTTL},
		queue,
		events,
		reqState,
		outbox,
		&chatAdapter{},
	), nil
}

