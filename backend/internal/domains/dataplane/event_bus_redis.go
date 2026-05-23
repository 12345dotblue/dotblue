package dataplane

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/redis/go-redis/v9"
)

type RedisEventBus struct {
	r *Redis
}

type redisSubscription struct {
	pubsub *redis.PubSub
	ch     <-chan *redis.Message
}

func (s *redisSubscription) Next(ctx context.Context) (*TurnEvent, error) {
	if s == nil || s.ch == nil {
		return nil, errors.New("subscription is closed")
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg, ok := <-s.ch:
		if !ok || msg == nil {
			return nil, errors.New("subscription is closed")
		}
		var ev TurnEvent
		if err := json.Unmarshal([]byte(msg.Payload), &ev); err != nil {
			return nil, err
		}
		return &ev, nil
	}
}

func (s *redisSubscription) Close() error {
	if s == nil || s.pubsub == nil {
		return nil
	}
	return s.pubsub.Close()
}

func NewRedisEventBus(r *Redis) *RedisEventBus {
	return &RedisEventBus{r: r}
}

func (b *RedisEventBus) Publish(ctx context.Context, requestID string, event TurnEvent) error {
	if b == nil || b.r == nil || b.r.client == nil {
		return errors.New("event bus is not configured")
	}
	key := b.r.keyspace.RequestEventChannel(requestID)
	payload, err := marshalJSON(event)
	if err != nil {
		return err
	}
	return b.r.client.Publish(ctx, key, payload).Err()
}

func (b *RedisEventBus) Subscribe(ctx context.Context, requestID string) (EventSubscription, error) {
	if b == nil || b.r == nil || b.r.client == nil {
		return nil, errors.New("event bus is not configured")
	}
	key := b.r.keyspace.RequestEventChannel(requestID)
	ps := b.r.client.Subscribe(ctx, key)
	_, err := ps.Receive(ctx)
	if err != nil {
		_ = ps.Close()
		return nil, err
	}
	return &redisSubscription{pubsub: ps, ch: ps.Channel()}, nil
}

