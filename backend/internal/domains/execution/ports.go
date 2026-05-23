package execution

import (
	"context"

	"dotblue/internal/domains/dataplane"
)

type SessionControl interface {
	ValidateAssignment(ctx context.Context, sessionKey, workerID string, fenceToken int64) (bool, error)
	TryEnterGate(ctx context.Context, sessionKey, holder string) (bool, error)
	LeaveGate(ctx context.Context, sessionKey string) error
	Heartbeat(ctx context.Context, workerID string, activeTurns, maxConcurrentTurns int64) error
}

type ChatResult struct {
	ConversationID string
	Thinking       string
	Content        string
	Title          string
}

type ChatExecutor interface {
	Execute(ctx context.Context, task dataplane.TurnTask) (*ChatResult, error)
}

type OutboxWriter interface {
	EnqueueJSON(ctx context.Context, connectionID, payload string) error
}

type OutboundEnvelope struct {
	Platform         string         `json:"platform"`
	ConnectionID     string         `json:"connectionId"`
	EnterpriseID     string         `json:"enterpriseId"`
	ConversationID   string         `json:"conversationId"`
	AgentID          string         `json:"agentId"`
	ExternalChatID   string         `json:"externalChatId"`
	ExternalThreadID string         `json:"externalThreadId"`
	ReplyHandle      map[string]any `json:"replyHandle,omitempty"`
	Text             string         `json:"text"`
}

