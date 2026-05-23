package im

import (
	"context"
	"encoding/json"
	"time"

	"dotblue/internal/domains/dataplane"
	"github.com/gogf/gf/v2/frame/g"
)

func processOutboundOutbox(ctx context.Context, conn Connection) error {
	dp, err := dataplane.Default(ctx)
	if err != nil {
		return err
	}
	outbox := dataplane.NewRedisIMOutbox(dp, g.Cfg().MustGet(ctx, "worker.inboxTTL").Duration())
	block := g.Cfg().MustGet(ctx, "worker.claimBlock").Duration()
	if block <= 0 {
		block = 2 * time.Second
	}
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		msg, err := outbox.Claim(ctx, conn.ID, block)
		if err != nil {
			return err
		}
		if msg == nil {
			continue
		}
		var env OutboundEnvelope
		if err := json.Unmarshal([]byte(msg.Payload), &env); err == nil {
			_ = deliverOutboundEnvelope(ctx, conn, env, "")
		}
		_ = outbox.Ack(ctx, conn.ID, msg.MessageID)
	}
}

