package session

import (
	"context"

	"dotblue/internal/domains/dataplane"
	"github.com/gogf/gf/v2/frame/g"
)

func Default(ctx context.Context) (*Service, error) {
	dp, err := dataplane.Default(ctx)
	if err != nil {
		return nil, err
	}
	cfg := Config{
		OwnerTTL: g.Cfg().MustGet(ctx, "session.ownerTTL").Duration(),
		FenceTTL: g.Cfg().MustGet(ctx, "session.fenceTTL").Duration(),
		GateTTL:  g.Cfg().MustGet(ctx, "session.gateTTL").Duration(),
		StateTTL: g.Cfg().MustGet(ctx, "session.stateTTL").Duration(),
	}
	return NewService(
		cfg,
		newRedisLeaseStore(dp.Client(), dp.Keyspace()),
		newRedisWorkerCatalog(dp.Client(), dp.Keyspace()),
	), nil
}

