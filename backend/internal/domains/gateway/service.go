package gateway

import (
	"context"
	"errors"
	"strings"
	"time"

	"dotblue/internal/domains/dataplane"
)

type DispatchRequest struct {
	RequestID        string
	SessionKey       string
	EnterpriseID     string
	UserID           string
	AgentID          string
	ConversationID   string
	ConnectionID     string
	IngressType      string
	IngressID        string
	InboundMessageID string
	ExternalChatID   string
	ExternalThreadID string
	ReplyHandle      map[string]any
	Content          string
	CreatedAt        time.Time
}

type DispatchResult struct {
	RequestID  string
	SessionKey string
	WorkerID   string
	FenceToken int64
}

type Assignment struct {
	SessionKey string
	WorkerID   string
	FenceToken int64
	LeaseTTL   time.Duration
}

type SessionAssigner interface {
	AcquireOwner(ctx context.Context, sessionKey string) (*Assignment, error)
}

type Service struct {
	sessionSvc SessionAssigner
	queue      dataplane.TaskQueue
	reqState   dataplane.RequestStateStore
	reqRoute   dataplane.RequestRouteStore
	finalTTL   time.Duration
}

func NewService(
	sessionSvc SessionAssigner,
	queue dataplane.TaskQueue,
	reqState dataplane.RequestStateStore,
	reqRoute dataplane.RequestRouteStore,
	finalTTL time.Duration,
) *Service {
	return &Service{
		sessionSvc: sessionSvc,
		queue:      queue,
		reqState:   reqState,
		reqRoute:   reqRoute,
		finalTTL:   finalTTL,
	}
}

func (s *Service) Dispatch(ctx context.Context, req DispatchRequest) (*DispatchResult, error) {
	if s == nil || s.sessionSvc == nil || s.queue == nil || s.reqState == nil || s.reqRoute == nil {
		return nil, errors.New("gateway service is not configured")
	}
	req.RequestID = strings.TrimSpace(req.RequestID)
	req.SessionKey = strings.TrimSpace(req.SessionKey)
	req.AgentID = strings.TrimSpace(req.AgentID)
	if req.RequestID == "" || req.SessionKey == "" || req.AgentID == "" {
		return nil, errors.New("dispatch request is incomplete")
	}
	assign, err := s.sessionSvc.AcquireOwner(ctx, req.SessionKey)
	if err != nil {
		return nil, err
	}
	task := dataplane.TurnTask{
		RequestID:        req.RequestID,
		SessionKey:       req.SessionKey,
		FenceToken:       assign.FenceToken,
		EnterpriseID:     req.EnterpriseID,
		UserID:           req.UserID,
		AgentID:          req.AgentID,
		ConversationID:   req.ConversationID,
		ConnectionID:     req.ConnectionID,
		IngressType:      req.IngressType,
		InboundMessageID: req.InboundMessageID,
		ExternalChatID:   req.ExternalChatID,
		ExternalThreadID: req.ExternalThreadID,
		ReplyHandle:      req.ReplyHandle,
		Content:          req.Content,
		CreatedAt:        req.CreatedAt,
	}
	_ = s.reqState.SetRunning(ctx, req.RequestID, map[string]any{
		"status":      "queued",
		"session_key": req.SessionKey,
		"worker_id":   assign.WorkerID,
		"updated_at":  time.Now().Unix(),
	})
	_ = s.reqRoute.Set(ctx, req.RequestID, map[string]any{
		"ingress_type": req.IngressType,
		"ingress_id":   req.IngressID,
		"session_key":  req.SessionKey,
		"worker_id":    assign.WorkerID,
		"agent_id":     req.AgentID,
		"created_at":   time.Now().Unix(),
	}, s.finalTTL)
	if err := s.queue.Enqueue(ctx, assign.WorkerID, task); err != nil {
		return nil, err
	}
	return &DispatchResult{
		RequestID:  req.RequestID,
		SessionKey: req.SessionKey,
		WorkerID:   assign.WorkerID,
		FenceToken: assign.FenceToken,
	}, nil
}
