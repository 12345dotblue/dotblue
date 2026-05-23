package dataplane

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisTaskQueue struct {
	r        *Redis
	inboxTTL time.Duration
}

func NewRedisTaskQueue(r *Redis, inboxTTL time.Duration) *RedisTaskQueue {
	return &RedisTaskQueue{r: r, inboxTTL: inboxTTL}
}

func (q *RedisTaskQueue) Enqueue(ctx context.Context, workerID string, task TurnTask) error {
	if q == nil || q.r == nil || q.r.client == nil {
		return errors.New("task queue is not configured")
	}
	key := q.r.keyspace.WorkerTaskStream(workerID)
	payload, err := marshalJSON(task)
	if err != nil {
		return err
	}
	args := &redis.XAddArgs{
		Stream: key,
		Values: map[string]any{"task": payload},
	}
	if q.r.dpCfg.StreamMaxLen > 0 {
		args.MaxLen = q.r.dpCfg.StreamMaxLen
		args.Approx = true
	}
	if err := q.r.client.XAdd(ctx, args).Err(); err != nil {
		return err
	}
	if q.inboxTTL > 0 {
		_ = q.r.client.Expire(ctx, key, q.inboxTTL).Err()
	}
	return nil
}

func (q *RedisTaskQueue) ensureGroup(ctx context.Context, stream string) {
	_ = q.r.client.XGroupCreateMkStream(ctx, stream, q.r.streamGroup, "0").Err()
}

func (q *RedisTaskQueue) Claim(ctx context.Context, workerID string, block time.Duration) (*QueuedTask, error) {
	if q == nil || q.r == nil || q.r.client == nil {
		return nil, errors.New("task queue is not configured")
	}
	key := q.r.keyspace.WorkerTaskStream(workerID)
	q.ensureGroup(ctx, key)
	consumer := workerID
	if block <= 0 {
		block = time.Second
	}
	res, err := q.r.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    q.r.streamGroup,
		Consumer: consumer,
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
	raw, _ := msg.Values["task"].(string)
	if raw == "" {
		return &QueuedTask{MessageID: msg.ID}, nil
	}
	var task TurnTask
	if uerr := json.Unmarshal([]byte(raw), &task); uerr != nil {
		return nil, uerr
	}
	if q.inboxTTL > 0 {
		_ = q.r.client.Expire(ctx, key, q.inboxTTL).Err()
	}
	return &QueuedTask{MessageID: msg.ID, Task: task}, nil
}

func (q *RedisTaskQueue) Ack(ctx context.Context, workerID, messageID string) error {
	if q == nil || q.r == nil || q.r.client == nil {
		return errors.New("task queue is not configured")
	}
	key := q.r.keyspace.WorkerTaskStream(workerID)
	if err := q.r.client.XAck(ctx, key, q.r.streamGroup, messageID).Err(); err != nil {
		return err
	}
	if q.inboxTTL > 0 {
		_ = q.r.client.Expire(ctx, key, q.inboxTTL).Err()
	}
	return nil
}
