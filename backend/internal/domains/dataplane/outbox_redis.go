package dataplane

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisIMOutbox struct {
	r        *Redis
	outboxTTL time.Duration
}

func NewRedisIMOutbox(r *Redis, outboxTTL time.Duration) *RedisIMOutbox {
	return &RedisIMOutbox{r: r, outboxTTL: outboxTTL}
}

func (o *RedisIMOutbox) EnqueueJSON(ctx context.Context, connectionID string, payload string) error {
	if o == nil || o.r == nil || o.r.client == nil {
		return errors.New("outbox is not configured")
	}
	key := o.r.keyspace.OutboxConnection(connectionID)
	args := &redis.XAddArgs{
		Stream: key,
		Values: map[string]any{"payload": payload},
	}
	if o.r.dpCfg.StreamMaxLen > 0 {
		args.MaxLen = o.r.dpCfg.StreamMaxLen
		args.Approx = true
	}
	if err := o.r.client.XAdd(ctx, args).Err(); err != nil {
		return err
	}
	if o.outboxTTL > 0 {
		_ = o.r.client.Expire(ctx, key, o.outboxTTL).Err()
	}
	return nil
}

func (o *RedisIMOutbox) ensureGroup(ctx context.Context, stream string) {
	_ = o.r.client.XGroupCreateMkStream(ctx, stream, o.r.streamGroup, "0").Err()
}

func (o *RedisIMOutbox) Claim(ctx context.Context, connectionID string, block time.Duration) (*QueuedOutbound, error) {
	if o == nil || o.r == nil || o.r.client == nil {
		return nil, errors.New("outbox is not configured")
	}
	key := o.r.keyspace.OutboxConnection(connectionID)
	o.ensureGroup(ctx, key)
	if block <= 0 {
		block = time.Second
	}
	res, err := o.r.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    o.r.streamGroup,
		Consumer: connectionID,
		Streams:  []string{key, ">"},
		Count:    1,
		Block:    block,
	}).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	if len(res) == 0 || len(res[0].Messages) == 0 {
		return nil, nil
	}
	msg := res[0].Messages[0]
	payload, _ := msg.Values["payload"].(string)
	if o.outboxTTL > 0 {
		_ = o.r.client.Expire(ctx, key, o.outboxTTL).Err()
	}
	return &QueuedOutbound{MessageID: msg.ID, Payload: payload}, nil
}

func (o *RedisIMOutbox) Ack(ctx context.Context, connectionID, messageID string) error {
	if o == nil || o.r == nil || o.r.client == nil {
		return errors.New("outbox is not configured")
	}
	key := o.r.keyspace.OutboxConnection(connectionID)
	if err := o.r.client.XAck(ctx, key, o.r.streamGroup, messageID).Err(); err != nil {
		return err
	}
	if o.outboxTTL > 0 {
		_ = o.r.client.Expire(ctx, key, o.outboxTTL).Err()
	}
	return nil
}

