package session

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"dotblue/internal/domains/dataplane"
	"github.com/redis/go-redis/v9"
)

type redisLeaseStore struct {
	client   *redis.Client
	keyspace dataplane.Keyspace
}

func newRedisLeaseStore(client *redis.Client, keyspace dataplane.Keyspace) LeaseStore {
	return &redisLeaseStore{client: client, keyspace: keyspace}
}

func (s *redisLeaseStore) GetOwner(ctx context.Context, sessionKey string) (string, error) {
	val, err := s.client.Get(ctx, s.keyspace.SessionOwner(sessionKey)).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	return val, err
}

func (s *redisLeaseStore) TryAcquireOwner(ctx context.Context, sessionKey, workerID string, ttl time.Duration) (bool, error) {
	return s.client.SetNX(ctx, s.keyspace.SessionOwner(sessionKey), workerID, ttl).Result()
}

func (s *redisLeaseStore) RefreshOwner(ctx context.Context, sessionKey string, ttl time.Duration) error {
	return s.client.Expire(ctx, s.keyspace.SessionOwner(sessionKey), ttl).Err()
}

func (s *redisLeaseStore) DeleteOwner(ctx context.Context, sessionKey string) error {
	return s.client.Del(ctx, s.keyspace.SessionOwner(sessionKey)).Err()
}

func (s *redisLeaseStore) GetFence(ctx context.Context, sessionKey string) (int64, error) {
	val, err := s.client.Get(ctx, s.keyspace.SessionFence(sessionKey)).Result()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	out, _ := strconv.ParseInt(strings.TrimSpace(val), 10, 64)
	return out, nil
}

func (s *redisLeaseStore) IncrFence(ctx context.Context, sessionKey string, ttl time.Duration) (int64, error) {
	val, err := s.client.Incr(ctx, s.keyspace.SessionFence(sessionKey)).Result()
	if err != nil {
		return 0, err
	}
	if ttl > 0 {
		_ = s.client.Expire(ctx, s.keyspace.SessionFence(sessionKey), ttl).Err()
	}
	return val, nil
}

func (s *redisLeaseStore) TryEnterGate(ctx context.Context, sessionKey, holder string, ttl time.Duration) (bool, error) {
	return s.client.SetNX(ctx, s.keyspace.SessionGate(sessionKey), holder, ttl).Result()
}

func (s *redisLeaseStore) LeaveGate(ctx context.Context, sessionKey string) error {
	return s.client.Del(ctx, s.keyspace.SessionGate(sessionKey)).Err()
}

type redisWorkerCatalog struct {
	client   *redis.Client
	keyspace dataplane.Keyspace
}

func newRedisWorkerCatalog(client *redis.Client, keyspace dataplane.Keyspace) WorkerCatalog {
	return &redisWorkerCatalog{client: client, keyspace: keyspace}
}

func (c *redisWorkerCatalog) Heartbeat(ctx context.Context, meta WorkerMeta, ttl time.Duration) error {
	key := c.keyspace.WorkerMeta(meta.WorkerID)
	fields := map[string]any{
		"worker_id":            meta.WorkerID,
		"active_turns":         meta.ActiveTurns,
		"max_concurrent_turns": meta.MaxConcurrentTurns,
		"updated_at":           time.Now().Unix(),
	}
	if err := c.client.HSet(ctx, key, fields).Err(); err != nil {
		return err
	}
	if ttl > 0 {
		_ = c.client.Expire(ctx, key, ttl).Err()
	}
	return nil
}

func (c *redisWorkerCatalog) Exists(ctx context.Context, workerID string) bool {
	exists, err := c.client.Exists(ctx, c.keyspace.WorkerMeta(workerID)).Result()
	return err == nil && exists > 0
}

func (c *redisWorkerCatalog) List(ctx context.Context) ([]WorkerMeta, error) {
	var cursor uint64
	pattern := c.keyspace.WorkerMetaPattern()
	candidates := make([]WorkerMeta, 0, 8)
	for {
		keys, next, err := c.client.Scan(ctx, cursor, pattern, 50).Result()
		if err != nil {
			return nil, err
		}
		for _, key := range keys {
			meta, err := c.client.HGetAll(ctx, key).Result()
			if err != nil || len(meta) == 0 {
				continue
			}
			workerID := strings.TrimSpace(meta["worker_id"])
			if workerID == "" {
				continue
			}
			active, _ := strconv.ParseInt(meta["active_turns"], 10, 64)
			maxc, _ := strconv.ParseInt(meta["max_concurrent_turns"], 10, 64)
			candidates = append(candidates, WorkerMeta{
				WorkerID:           workerID,
				ActiveTurns:        active,
				MaxConcurrentTurns: maxc,
			})
		}
		if next == 0 {
			break
		}
		cursor = next
	}
	return candidates, nil
}

