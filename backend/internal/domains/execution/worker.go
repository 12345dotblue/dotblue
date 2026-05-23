package execution

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"time"

	"dotblue/internal/domains/dataplane"
)

type Config struct {
	WorkerID           string
	MaxConcurrentTurns int64
	HeartbeatInterval  time.Duration
	MetaTTL            time.Duration
	InboxTTL           time.Duration
	ClaimBlock         time.Duration
	TaskTTL            time.Duration
}

type Worker struct {
	cfg      Config
	session  SessionControl
	queue    dataplane.TaskQueue
	events   dataplane.RequestEventBus
	reqState dataplane.RequestStateStore
	outbox   OutboxWriter
	chat     ChatExecutor
	active   atomic.Int64
}

func NewWorker(cfg Config, session SessionControl, queue dataplane.TaskQueue, events dataplane.RequestEventBus, reqState dataplane.RequestStateStore, outbox OutboxWriter, chat ChatExecutor) *Worker {
	return &Worker{
		cfg:      cfg,
		session:  session,
		queue:    queue,
		events:   events,
		reqState: reqState,
		outbox:   outbox,
		chat:     chat,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	if w == nil || w.session == nil || w.queue == nil || w.events == nil || w.reqState == nil || w.chat == nil {
		return errors.New("worker is not configured")
	}
	if w.cfg.WorkerID == "" {
		return errors.New("worker.id is required")
	}
	if w.cfg.HeartbeatInterval <= 0 {
		w.cfg.HeartbeatInterval = 10 * time.Second
	}
	if w.cfg.MetaTTL <= 0 {
		w.cfg.MetaTTL = 30 * time.Second
	}
	if w.cfg.ClaimBlock <= 0 {
		w.cfg.ClaimBlock = 2 * time.Second
	}
	go w.heartbeatLoop(ctx)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		task, err := w.queue.Claim(ctx, w.cfg.WorkerID, w.cfg.ClaimBlock)
		if err != nil {
			return err
		}
		if task == nil {
			continue
		}
		_ = w.handleTask(ctx, task)
		_ = w.queue.Ack(ctx, w.cfg.WorkerID, task.MessageID)
	}
}

func (w *Worker) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(w.cfg.HeartbeatInterval)
	defer ticker.Stop()
	for {
		if ctx.Err() != nil {
			return
		}
		_ = w.session.Heartbeat(ctx, w.cfg.WorkerID, w.active.Load(), w.cfg.MaxConcurrentTurns)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (w *Worker) handleTask(ctx context.Context, queued *dataplane.QueuedTask) error {
	task := queued.Task
	if task.RequestID == "" || task.SessionKey == "" {
		return nil
	}
	if !task.CreatedAt.IsZero() && w.cfg.TaskTTL > 0 && time.Since(task.CreatedAt) > w.cfg.TaskTTL {
		_ = w.reqState.SetFinal(ctx, task.RequestID, map[string]any{
			"status": "expired",
			"error":  "task expired",
		})
		return nil
	}
	ok, err := w.session.ValidateAssignment(ctx, task.SessionKey, w.cfg.WorkerID, task.FenceToken)
	if err != nil || !ok {
		return err
	}
	entered, err := w.session.TryEnterGate(ctx, task.SessionKey, task.RequestID)
	if err != nil || !entered {
		return err
	}
	defer func() { _ = w.session.LeaveGate(context.Background(), task.SessionKey) }()

	w.active.Add(1)
	defer w.active.Add(-1)

	_ = w.reqState.SetRunning(ctx, task.RequestID, map[string]any{
		"status":       "running",
		"session_key":  task.SessionKey,
		"worker_id":    w.cfg.WorkerID,
		"fence_token":  task.FenceToken,
		"updated_at":   time.Now().Unix(),
		"conversation": task.ConversationID,
	})

	result, err := w.chat.Execute(ctx, task)
	if err != nil {
		return w.publishError(ctx, task, err)
	}
	conversationID := task.ConversationID
	if result != nil && result.ConversationID != "" {
		conversationID = result.ConversationID
	}
	seq := int64(0)
	now := time.Now()
	if result != nil && result.Thinking != "" {
		seq++
		_ = w.events.Publish(ctx, task.RequestID, dataplane.TurnEvent{
			RequestID:      task.RequestID,
			SessionKey:     task.SessionKey,
			ConversationID: conversationID,
			Type:           "thinking",
			Status:         "thinking",
			Thinking:       result.Thinking,
			Seq:            seq,
			At:             now,
		})
	}
	if result != nil && result.Content != "" {
		seq++
		_ = w.events.Publish(ctx, task.RequestID, dataplane.TurnEvent{
			RequestID:      task.RequestID,
			SessionKey:     task.SessionKey,
			ConversationID: conversationID,
			Type:           "streaming",
			Status:         "streaming",
			Content:        result.Content,
			Seq:            seq,
			At:             now,
		})
	}
	title := ""
	if result != nil {
		title = result.Title
	}
	seq++
	_ = w.events.Publish(ctx, task.RequestID, dataplane.TurnEvent{
		RequestID:      task.RequestID,
		SessionKey:     task.SessionKey,
		ConversationID: conversationID,
		Type:           "meta",
		Status:         "done",
		Content:        title,
		Seq:            seq,
		At:             now,
	})
	seq++
	_ = w.events.Publish(ctx, task.RequestID, dataplane.TurnEvent{
		RequestID:      task.RequestID,
		SessionKey:     task.SessionKey,
		ConversationID: conversationID,
		Type:           "done",
		Status:         "done",
		Seq:            seq,
		At:             now,
	})
	if task.IngressType != "web" && task.ConnectionID != "" && w.outbox != nil && result != nil {
		envelope := OutboundEnvelope{
			Platform:         task.IngressType,
			ConnectionID:     task.ConnectionID,
			EnterpriseID:     task.EnterpriseID,
			ConversationID:   conversationID,
			AgentID:          task.AgentID,
			ExternalChatID:   task.ExternalChatID,
			ExternalThreadID: task.ExternalThreadID,
			ReplyHandle:      task.ReplyHandle,
			Text:             result.Content,
		}
		if payload, err := json.Marshal(envelope); err == nil {
			_ = w.outbox.EnqueueJSON(ctx, task.ConnectionID, string(payload))
		}
	}
	_ = w.reqState.SetFinal(ctx, task.RequestID, map[string]any{
		"status":     "done",
		"updated_at": time.Now().Unix(),
	})
	return nil
}

func (w *Worker) publishError(ctx context.Context, task dataplane.TurnTask, err error) error {
	if err == nil {
		return nil
	}
	now := time.Now()
	_ = w.events.Publish(ctx, task.RequestID, dataplane.TurnEvent{
		RequestID:      task.RequestID,
		SessionKey:     task.SessionKey,
		ConversationID: task.ConversationID,
		Type:           "error",
		Status:         "error",
		Error:          err.Error(),
		Seq:            1,
		At:             now,
	})
	_ = w.reqState.SetFinal(ctx, task.RequestID, map[string]any{
		"status":     "error",
		"error":      err.Error(),
		"updated_at": now.Unix(),
	})
	return err
}
