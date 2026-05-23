package gateway

import (
	"context"

	"dotblue/internal/domains/dataplane"
	"dotblue/internal/domains/session"
	"github.com/gogf/gf/v2/frame/g"
)

type sessionAdapter struct {
	svc *session.Service
}

func (a *sessionAdapter) AcquireOwner(ctx context.Context, sessionKey string) (*Assignment, error) {
	assign, err := a.svc.AcquireOwner(ctx, sessionKey)
	if err != nil || assign == nil {
		return nil, err
	}
	return &Assignment{
		SessionKey: assign.SessionKey,
		WorkerID:   assign.WorkerID,
		FenceToken: assign.FenceToken,
		LeaseTTL:   assign.LeaseTTL,
	}, nil
}

func Default(ctx context.Context) (*Service, error) {
	ss, err := session.Default(ctx)
	if err != nil {
		return nil, err
	}
	dp, err := dataplane.Default(ctx)
	if err != nil {
		return nil, err
	}
	return NewService(
		&sessionAdapter{svc: ss},
		dataplane.NewRedisTaskQueue(dp, g.Cfg().MustGet(ctx, "worker.inboxTTL").Duration()),
		dataplane.NewRedisRequestStateStore(dp, dp.RunningTTL(), dp.FinalTTL()),
		dataplane.NewRedisRequestRouteStore(dp),
		dp.FinalTTL(),
	), nil
}
