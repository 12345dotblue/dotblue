package session

import (
	"context"
	"time"
)

type LeaseStore interface {
	GetOwner(ctx context.Context, sessionKey string) (string, error)
	TryAcquireOwner(ctx context.Context, sessionKey, workerID string, ttl time.Duration) (bool, error)
	RefreshOwner(ctx context.Context, sessionKey string, ttl time.Duration) error
	DeleteOwner(ctx context.Context, sessionKey string) error
	GetFence(ctx context.Context, sessionKey string) (int64, error)
	IncrFence(ctx context.Context, sessionKey string, ttl time.Duration) (int64, error)
	TryEnterGate(ctx context.Context, sessionKey, holder string, ttl time.Duration) (bool, error)
	LeaveGate(ctx context.Context, sessionKey string) error
}

type WorkerCatalog interface {
	Heartbeat(ctx context.Context, meta WorkerMeta, ttl time.Duration) error
	Exists(ctx context.Context, workerID string) bool
	List(ctx context.Context) ([]WorkerMeta, error)
}

