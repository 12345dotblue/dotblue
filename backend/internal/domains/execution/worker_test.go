package execution

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"dotblue/internal/domains/dataplane"
)

type fakeSessionControl struct {
	validateOK bool
	validateErr error
	enterOK    bool
	enterErr   error
	left       []string
}

func (f *fakeSessionControl) ValidateAssignment(ctx context.Context, sessionKey, workerID string, fenceToken int64) (bool, error) {
	return f.validateOK, f.validateErr
}

func (f *fakeSessionControl) TryEnterGate(ctx context.Context, sessionKey, holder string) (bool, error) {
	return f.enterOK, f.enterErr
}

func (f *fakeSessionControl) LeaveGate(ctx context.Context, sessionKey string) error {
	f.left = append(f.left, sessionKey)
	return nil
}

func (f *fakeSessionControl) Heartbeat(ctx context.Context, workerID string, activeTurns, maxConcurrentTurns int64) error {
	return nil
}

type fakeChatExecutor struct {
	result *ChatResult
	err    error
}

func (f *fakeChatExecutor) Execute(ctx context.Context, task dataplane.TurnTask) (*ChatResult, error) {
	return f.result, f.err
}

type fakeEventBus struct {
	events []dataplane.TurnEvent
}

func (f *fakeEventBus) Publish(ctx context.Context, requestID string, event dataplane.TurnEvent) error {
	f.events = append(f.events, event)
	return nil
}

func (f *fakeEventBus) Subscribe(ctx context.Context, requestID string) (dataplane.EventSubscription, error) {
	return nil, nil
}

type fakeStateStore struct {
	running []map[string]any
	final   []map[string]any
}

func (f *fakeStateStore) SetRunning(ctx context.Context, requestID string, fields map[string]any) error {
	f.running = append(f.running, fields)
	return nil
}

func (f *fakeStateStore) SetFinal(ctx context.Context, requestID string, fields map[string]any) error {
	f.final = append(f.final, fields)
	return nil
}

func (f *fakeStateStore) Patch(ctx context.Context, requestID string, fields map[string]any, ttl time.Duration) error {
	return nil
}

type fakeOutbox struct {
	connectionID string
	payload      string
}

func (f *fakeOutbox) EnqueueJSON(ctx context.Context, connectionID, payload string) error {
	f.connectionID = connectionID
	f.payload = payload
	return nil
}

func TestWorkerHandleTaskRejectsOnInvalidAssignment(t *testing.T) {
	w := NewWorker(
		Config{WorkerID: "w1"},
		&fakeSessionControl{validateOK: false},
		nil,
		&fakeEventBus{},
		&fakeStateStore{},
		nil,
		&fakeChatExecutor{},
	)

	err := w.handleTask(context.Background(), &dataplane.QueuedTask{
		Task: dataplane.TurnTask{RequestID: "req-1", SessionKey: "sess-1", FenceToken: 1},
	})
	if err != nil {
		t.Fatalf("handleTask() error = %v, want nil when assignment invalid without error", err)
	}
}

func TestWorkerHandleTaskPublishesErrorOnChatFailure(t *testing.T) {
	ev := &fakeEventBus{}
	st := &fakeStateStore{}
	w := NewWorker(
		Config{WorkerID: "w1"},
		&fakeSessionControl{validateOK: true, enterOK: true},
		nil,
		ev,
		st,
		nil,
		&fakeChatExecutor{err: errors.New("boom")},
	)

	err := w.handleTask(context.Background(), &dataplane.QueuedTask{
		Task: dataplane.TurnTask{RequestID: "req-1", SessionKey: "sess-1", ConversationID: "conv-1"},
	})
	if err == nil || err.Error() != "boom" {
		t.Fatalf("handleTask() error = %v, want boom", err)
	}
	if len(ev.events) != 1 || ev.events[0].Type != "error" {
		t.Fatalf("events = %+v, want one error event", ev.events)
	}
	if len(st.final) != 1 || st.final[0]["status"] != "error" {
		t.Fatalf("final states = %+v, want final error", st.final)
	}
}

func TestWorkerHandleTaskPublishesStreamingAndDone(t *testing.T) {
	ev := &fakeEventBus{}
	st := &fakeStateStore{}
	w := NewWorker(
		Config{WorkerID: "w1"},
		&fakeSessionControl{validateOK: true, enterOK: true},
		nil,
		ev,
		st,
		nil,
		&fakeChatExecutor{result: &ChatResult{
			ConversationID: "conv-1",
			Thinking:       "thinking",
			Content:        "hello",
			Title:          "title",
		}},
	)

	err := w.handleTask(context.Background(), &dataplane.QueuedTask{
		Task: dataplane.TurnTask{RequestID: "req-1", SessionKey: "sess-1", ConversationID: "conv-1"},
	})
	if err != nil {
		t.Fatalf("handleTask() error = %v", err)
	}
	if len(st.running) != 1 {
		t.Fatalf("running states = %d, want 1", len(st.running))
	}
	if len(st.final) != 1 || st.final[0]["status"] != "done" {
		t.Fatalf("final states = %+v, want final done", st.final)
	}
	if len(ev.events) != 4 {
		t.Fatalf("events len = %d, want 4", len(ev.events))
	}
	if ev.events[0].Type != "thinking" || ev.events[1].Type != "streaming" || ev.events[2].Type != "meta" || ev.events[3].Type != "done" {
		t.Fatalf("events order = %+v", ev.events)
	}
}

func TestWorkerHandleTaskWritesOutboxForIM(t *testing.T) {
	outbox := &fakeOutbox{}
	w := NewWorker(
		Config{WorkerID: "w1"},
		&fakeSessionControl{validateOK: true, enterOK: true},
		nil,
		&fakeEventBus{},
		&fakeStateStore{},
		outbox,
		&fakeChatExecutor{result: &ChatResult{
			ConversationID: "conv-1",
			Content:        "hello",
		}},
	)

	err := w.handleTask(context.Background(), &dataplane.QueuedTask{
		Task: dataplane.TurnTask{
			RequestID:      "req-1",
			SessionKey:     "sess-1",
			ConversationID: "conv-1",
			IngressType:    "feishu",
			ConnectionID:   "conn-1",
			EnterpriseID:   "ent-1",
			AgentID:        "agent-1",
			ReplyHandle:    map[string]any{"message_id": "m1"},
		},
	})
	if err != nil {
		t.Fatalf("handleTask() error = %v", err)
	}
	if outbox.connectionID != "conn-1" || outbox.payload == "" {
		t.Fatalf("outbox = %+v, want payload written", outbox)
	}
	var env OutboundEnvelope
	if uerr := json.Unmarshal([]byte(outbox.payload), &env); uerr != nil {
		t.Fatalf("unmarshal outbox payload error = %v", uerr)
	}
	if env.ConnectionID != "conn-1" || env.Text != "hello" {
		t.Fatalf("outbox envelope = %+v", env)
	}
}

