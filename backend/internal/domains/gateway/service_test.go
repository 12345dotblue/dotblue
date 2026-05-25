package gateway

import (
	"context"
	"testing"
	"time"

	"dotblue/internal/domains/dataplane"
)

type fakeSessionAssigner struct {
	workerID   string
	fenceToken int64
	err        error
}

func (f *fakeSessionAssigner) AcquireOwner(ctx context.Context, sessionKey string) (*Assignment, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &Assignment{
		SessionKey: sessionKey,
		WorkerID:   f.workerID,
		FenceToken: f.fenceToken,
		LeaseTTL:   30 * time.Second,
	}, nil
}

type fakeQueue struct {
	workerID string
	task     dataplane.TurnTask
}

func (f *fakeQueue) Enqueue(ctx context.Context, workerID string, task dataplane.TurnTask) error {
	f.workerID = workerID
	f.task = task
	return nil
}

func (f *fakeQueue) Claim(ctx context.Context, workerID string, block time.Duration) (*dataplane.QueuedTask, error) {
	return nil, nil
}

func (f *fakeQueue) Ack(ctx context.Context, workerID, messageID string) error {
	return nil
}

type fakeReqState struct {
	fields map[string]any
}

func (f *fakeReqState) SetRunning(ctx context.Context, requestID string, fields map[string]any) error {
	f.fields = fields
	return nil
}

func (f *fakeReqState) SetFinal(ctx context.Context, requestID string, fields map[string]any) error {
	return nil
}

func (f *fakeReqState) Patch(ctx context.Context, requestID string, fields map[string]any, ttl time.Duration) error {
	return nil
}

type fakeReqRoute struct {
	fields map[string]any
}

func (f *fakeReqRoute) Set(ctx context.Context, requestID string, fields map[string]any, ttl time.Duration) error {
	f.fields = fields
	return nil
}

type fakeLimiter struct{}

func (fakeLimiter) CheckLimit(input LimitCheckInput) error {
	return nil
}

func TestDispatchWritesStateRouteAndQueue(t *testing.T) {
	assigner := &fakeSessionAssigner{workerID: "worker-1", fenceToken: 7}
	queue := &fakeQueue{}
	state := &fakeReqState{}
	route := &fakeReqRoute{}
	svc := NewService(assigner, queue, state, route, time.Hour, fakeLimiter{})

	res, err := svc.Dispatch(context.Background(), DispatchRequest{
		RequestID:        "req-1",
		SessionKey:       "sess-1",
		EnterpriseID:     "ent-1",
		UserID:           "user-1",
		AgentID:          "agent-1",
		ConversationID:   "conv-1",
		ConnectionID:     "conn-1",
		IngressType:      "web",
		IngressID:        "http",
		InboundMessageID: "msg-1",
		Content:          "hello",
		CreatedAt:        time.Unix(100, 0),
	})
	if err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}
	if res == nil || res.WorkerID != "worker-1" || res.FenceToken != 7 {
		t.Fatalf("dispatch result = %+v", res)
	}
	if queue.workerID != "worker-1" {
		t.Fatalf("queue workerID = %q, want worker-1", queue.workerID)
	}
	if queue.task.RequestID != "req-1" || queue.task.SessionKey != "sess-1" || queue.task.FenceToken != 7 {
		t.Fatalf("queued task = %+v", queue.task)
	}
	if state.fields["status"] != "queued" {
		t.Fatalf("state fields = %+v", state.fields)
	}
	if route.fields["ingress_type"] != "web" || route.fields["worker_id"] != "worker-1" {
		t.Fatalf("route fields = %+v", route.fields)
	}
}
