package dataplane

import (
	"context"
	"errors"
	"time"
)

type RedisRequestStateStore struct {
	r          *Redis
	runningTTL time.Duration
	finalTTL   time.Duration
}

func NewRedisRequestStateStore(r *Redis, runningTTL, finalTTL time.Duration) *RedisRequestStateStore {
	return &RedisRequestStateStore{r: r, runningTTL: runningTTL, finalTTL: finalTTL}
}

func (s *RedisRequestStateStore) Patch(ctx context.Context, requestID string, fields map[string]any, ttl time.Duration) error {
	if s == nil || s.r == nil || s.r.client == nil {
		return errors.New("request state store is not configured")
	}
	if requestID == "" {
		return errors.New("requestID is required")
	}
	key := s.r.keyspace.RequestState(requestID)
	if len(fields) > 0 {
		if err := s.r.client.HSet(ctx, key, fields).Err(); err != nil {
			return err
		}
	}
	if ttl > 0 {
		_ = s.r.client.Expire(ctx, key, ttl).Err()
	}
	return nil
}

func (s *RedisRequestStateStore) SetRunning(ctx context.Context, requestID string, fields map[string]any) error {
	return s.Patch(ctx, requestID, fields, s.runningTTL)
}

func (s *RedisRequestStateStore) SetFinal(ctx context.Context, requestID string, fields map[string]any) error {
	return s.Patch(ctx, requestID, fields, s.finalTTL)
}

