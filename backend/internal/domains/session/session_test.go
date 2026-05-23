package session

import (
	"context"
	"testing"
	"time"
)

func TestChooseBestWorker(t *testing.T) {
	tests := []struct {
		name       string
		candidates []WorkerMeta
		want       string
	}{
		{
			name: "pick lower load ratio",
			candidates: []WorkerMeta{
				{WorkerID: "w1", ActiveTurns: 3, MaxConcurrentTurns: 4},
				{WorkerID: "w2", ActiveTurns: 1, MaxConcurrentTurns: 4},
			},
			want: "w2",
		},
		{
			name: "treat zero max as one",
			candidates: []WorkerMeta{
				{WorkerID: "w1", ActiveTurns: 1, MaxConcurrentTurns: 0},
				{WorkerID: "w2", ActiveTurns: 0, MaxConcurrentTurns: 0},
			},
			want: "w2",
		},
		{
			name: "skip empty worker id",
			candidates: []WorkerMeta{
				{WorkerID: "", ActiveTurns: 0, MaxConcurrentTurns: 1},
				{WorkerID: "w2", ActiveTurns: 2, MaxConcurrentTurns: 3},
			},
			want: "w2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := chooseBestWorker(tc.candidates); got != tc.want {
				t.Fatalf("chooseBestWorker() = %q, want %q", got, tc.want)
			}
		})
	}
}

type fakeLeaseStore struct {
	owner       string
	fence       int64
	acquireOK   bool
	refreshHits int
}

func (f *fakeLeaseStore) GetOwner(ctx context.Context, sessionKey string) (string, error) {
	return f.owner, nil
}

func (f *fakeLeaseStore) TryAcquireOwner(ctx context.Context, sessionKey, workerID string, ttl time.Duration) (bool, error) {
	if f.acquireOK {
		f.owner = workerID
	}
	return f.acquireOK, nil
}

func (f *fakeLeaseStore) RefreshOwner(ctx context.Context, sessionKey string, ttl time.Duration) error {
	f.refreshHits++
	return nil
}

func (f *fakeLeaseStore) DeleteOwner(ctx context.Context, sessionKey string) error {
	f.owner = ""
	return nil
}

func (f *fakeLeaseStore) GetFence(ctx context.Context, sessionKey string) (int64, error) {
	return f.fence, nil
}

func (f *fakeLeaseStore) IncrFence(ctx context.Context, sessionKey string, ttl time.Duration) (int64, error) {
	f.fence++
	return f.fence, nil
}

func (f *fakeLeaseStore) TryEnterGate(ctx context.Context, sessionKey, holder string, ttl time.Duration) (bool, error) {
	return true, nil
}

func (f *fakeLeaseStore) LeaveGate(ctx context.Context, sessionKey string) error {
	return nil
}

type fakeWorkerCatalog struct {
	exists     map[string]bool
	candidates []WorkerMeta
}

func (f *fakeWorkerCatalog) Heartbeat(ctx context.Context, meta WorkerMeta, ttl time.Duration) error {
	return nil
}

func (f *fakeWorkerCatalog) Exists(ctx context.Context, workerID string) bool {
	return f.exists[workerID]
}

func (f *fakeWorkerCatalog) List(ctx context.Context) ([]WorkerMeta, error) {
	return f.candidates, nil
}

func TestServiceAcquireOwnerUsesAliveExistingOwner(t *testing.T) {
	lease := &fakeLeaseStore{owner: "w1", fence: 3}
	catalog := &fakeWorkerCatalog{exists: map[string]bool{"w1": true}}
	svc := NewService(Config{OwnerTTL: 30 * time.Second, FenceTTL: time.Minute}, lease, catalog)

	assign, err := svc.AcquireOwner(context.Background(), "sess-1")
	if err != nil {
		t.Fatalf("AcquireOwner() error = %v", err)
	}
	if assign.WorkerID != "w1" || assign.FenceToken != 3 {
		t.Fatalf("assignment = %+v", assign)
	}
}

func TestServiceAcquireOwnerSelectsNewWorkerAndIncrementsFence(t *testing.T) {
	lease := &fakeLeaseStore{acquireOK: true}
	catalog := &fakeWorkerCatalog{
		exists: map[string]bool{},
		candidates: []WorkerMeta{
			{WorkerID: "w1", ActiveTurns: 2, MaxConcurrentTurns: 4},
			{WorkerID: "w2", ActiveTurns: 1, MaxConcurrentTurns: 4},
		},
	}
	svc := NewService(Config{OwnerTTL: 30 * time.Second, FenceTTL: time.Minute}, lease, catalog)

	assign, err := svc.AcquireOwner(context.Background(), "sess-1")
	if err != nil {
		t.Fatalf("AcquireOwner() error = %v", err)
	}
	if assign.WorkerID != "w2" {
		t.Fatalf("assignment worker = %q, want w2", assign.WorkerID)
	}
	if assign.FenceToken != 1 {
		t.Fatalf("assignment fence = %d, want 1", assign.FenceToken)
	}
	if lease.refreshHits != 1 {
		t.Fatalf("refresh hits = %d, want 1", lease.refreshHits)
	}
}
