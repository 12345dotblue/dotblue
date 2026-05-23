package dataplane

import (
	"context"
	"errors"
	"time"
)

type RedisRequestRouteStore struct {
	r *Redis
}

func NewRedisRequestRouteStore(r *Redis) *RedisRequestRouteStore {
	return &RedisRequestRouteStore{r: r}
}

func (s *RedisRequestRouteStore) Set(ctx context.Context, requestID string, fields map[string]any, ttl time.Duration) error {
	if s == nil || s.r == nil || s.r.client == nil {
		return errors.New("request route store is not configured")
	}
	if requestID == "" {
		return errors.New("requestID is required")
	}
	key := s.r.keyspace.RequestRoute(requestID)
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

