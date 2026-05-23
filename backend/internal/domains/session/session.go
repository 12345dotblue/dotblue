package session

import (
	"context"
	"errors"
	"strings"
	"time"
)

type Config struct {
	OwnerTTL time.Duration
	FenceTTL time.Duration
	GateTTL  time.Duration
	StateTTL time.Duration
}

type WorkerMeta struct {
	WorkerID          string
	ActiveTurns       int64
	MaxConcurrentTurns int64
}

type Assignment struct {
	SessionKey string
	WorkerID   string
	FenceToken int64
	LeaseTTL   time.Duration
}

type Service struct {
	cfg     Config
	lease   LeaseStore
	workers WorkerCatalog
}

func NewService(cfg Config, lease LeaseStore, workers WorkerCatalog) *Service {
	return &Service{cfg: cfg, lease: lease, workers: workers}
}

func (s *Service) AcquireOwner(ctx context.Context, sessionKey string) (*Assignment, error) {
	if s == nil || s.lease == nil || s.workers == nil {
		return nil, errors.New("session service is not configured")
	}
	sessionKey = strings.TrimSpace(sessionKey)
	if sessionKey == "" {
		return nil, errors.New("sessionKey is required")
	}
	ownerID, _ := s.lease.GetOwner(ctx, sessionKey)
	if strings.TrimSpace(ownerID) != "" {
		if s.workers.Exists(ctx, ownerID) {
			fence, _ := s.lease.GetFence(ctx, sessionKey)
			return &Assignment{SessionKey: sessionKey, WorkerID: ownerID, FenceToken: fence, LeaseTTL: s.cfg.OwnerTTL}, nil
		}
		_ = s.lease.DeleteOwner(ctx, sessionKey)
	}
	workerID, err := s.selectWorker(ctx)
	if err != nil {
		return nil, err
	}
	ok, err := s.lease.TryAcquireOwner(ctx, sessionKey, workerID, s.cfg.OwnerTTL)
	if err != nil {
		return nil, err
	}
	if !ok {
		ownerID, _ = s.lease.GetOwner(ctx, sessionKey)
		if strings.TrimSpace(ownerID) == "" {
			return nil, errors.New("failed to acquire session owner")
		}
		fence, _ := s.lease.GetFence(ctx, sessionKey)
		return &Assignment{SessionKey: sessionKey, WorkerID: ownerID, FenceToken: fence, LeaseTTL: s.cfg.OwnerTTL}, nil
	}
	fence, err := s.lease.IncrFence(ctx, sessionKey, s.cfg.FenceTTL)
	if err != nil {
		return nil, err
	}
	_ = s.lease.RefreshOwner(ctx, sessionKey, s.cfg.OwnerTTL)
	return &Assignment{SessionKey: sessionKey, WorkerID: workerID, FenceToken: fence, LeaseTTL: s.cfg.OwnerTTL}, nil
}

func (s *Service) RefreshOwner(ctx context.Context, sessionKey, workerID string, fenceToken int64) error {
	if s == nil || s.lease == nil {
		return errors.New("session service is not configured")
	}
	current, _ := s.lease.GetOwner(ctx, sessionKey)
	if strings.TrimSpace(current) != strings.TrimSpace(workerID) {
		return errors.New("session owner mismatch")
	}
	fence, _ := s.lease.GetFence(ctx, sessionKey)
	if fenceToken > 0 && fence > 0 && fence != fenceToken {
		return errors.New("session fence mismatch")
	}
	return s.lease.RefreshOwner(ctx, sessionKey, s.cfg.OwnerTTL)
}

func (s *Service) ValidateAssignment(ctx context.Context, sessionKey, workerID string, fenceToken int64) (bool, error) {
	current, err := s.lease.GetOwner(ctx, sessionKey)
	if err != nil || strings.TrimSpace(current) == "" {
		return false, err
	}
	if strings.TrimSpace(current) != strings.TrimSpace(workerID) {
		return false, nil
	}
	if fenceToken <= 0 {
		return true, nil
	}
	fence, _ := s.lease.GetFence(ctx, sessionKey)
	return fence == fenceToken, nil
}

func (s *Service) TryEnterGate(ctx context.Context, sessionKey, holder string) (bool, error) {
	if s == nil || s.lease == nil {
		return false, errors.New("session service is not configured")
	}
	return s.lease.TryEnterGate(ctx, sessionKey, holder, s.cfg.GateTTL)
}

func (s *Service) LeaveGate(ctx context.Context, sessionKey string) error {
	if s == nil || s.lease == nil {
		return errors.New("session service is not configured")
	}
	return s.lease.LeaveGate(ctx, sessionKey)
}

func (s *Service) HeartbeatWorker(ctx context.Context, meta WorkerMeta, ttl time.Duration) error {
	if s == nil || s.workers == nil {
		return errors.New("session service is not configured")
	}
	return s.workers.Heartbeat(ctx, meta, ttl)
}

func (s *Service) selectWorker(ctx context.Context) (string, error) {
	candidates, err := s.workers.List(ctx)
	if err != nil {
		return "", err
	}
	bestID := chooseBestWorker(candidates)
	if bestID == "" {
		return "", errors.New("no available worker")
	}
	return bestID, nil
}

func chooseBestWorker(candidates []WorkerMeta) string {
	bestID := ""
	bestScore := float64(1 << 62)
	for _, meta := range candidates {
		if strings.TrimSpace(meta.WorkerID) == "" {
			continue
		}
		maxc := meta.MaxConcurrentTurns
		if maxc <= 0 {
			maxc = 1
		}
		score := float64(meta.ActiveTurns) / float64(maxc)
		if score < bestScore {
			bestScore = score
			bestID = meta.WorkerID
		}
	}
	return bestID
}
