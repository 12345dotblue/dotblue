package im

import "time"

// Connection represents a tenant-scoped IM integration endpoint.
type Connection struct {
	ID              string         `json:"id"`
	EnterpriseID    string         `json:"enterpriseId"`
	Platform        string         `json:"platform"`
	Name            string         `json:"name"`
	Status          string         `json:"status"`
	ConnectionMode  string         `json:"connectionMode"`
	Config          map[string]any `json:"config"`
	Secrets         map[string]any `json:"-"`
	SecretsMasked   map[string]any `json:"secretsMasked,omitempty"`
	CallbackPath    string         `json:"callbackPath"`
	LastConnectedAt *time.Time     `json:"lastConnectedAt,omitempty"`
	LastError       string         `json:"lastError"`
	CreatedBy       string         `json:"createdBy"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
}

// AgentBinding links a connection to an agent with deterministic routing rules.
type AgentBinding struct {
	ID              string         `json:"id"`
	EnterpriseID    string         `json:"enterpriseId"`
	AgentID         string         `json:"agentId"`
	ConnectionID    string         `json:"connectionId"`
	Status          string         `json:"status"`
	TriggerMode     string         `json:"triggerMode"`
	TriggerConfig   map[string]any `json:"triggerConfig"`
	SessionStrategy string         `json:"sessionStrategy"`
	ReplyMode       string         `json:"replyMode"`
	AllowGroup      bool           `json:"allowGroup"`
	AllowDM         bool           `json:"allowDm"`
	Priority        int            `json:"priority"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
}

// SessionAddress is the normalized address used to resolve an external session.
type SessionAddress struct {
	Platform     string `json:"platform"`
	EnterpriseID string `json:"enterpriseId"`
	ConnectionID string `json:"connectionId"`
	ChatID       string `json:"chatId"`
	ThreadID     string `json:"threadId"`
	UserID       string `json:"userId"`
	ChatType     string `json:"chatType"`
}

type InboundAttachment struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	ContentType string `json:"contentType"`
	URL         string `json:"url,omitempty"`
	MediaRef    string `json:"mediaRef,omitempty"`
}

type OutboundAttachment struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	MediaRef string `json:"mediaRef"`
}

type RichSegment struct {
	Type string         `json:"type"`
	Text string         `json:"text,omitempty"`
	Meta map[string]any `json:"meta,omitempty"`
}

// InboundEvent is the normalized payload emitted by an IM adapter.
type InboundEvent struct {
	Platform         string              `json:"platform"`
	ConnectionID     string              `json:"connectionId"`
	EnterpriseID     string              `json:"enterpriseId"`
	EventID          string              `json:"eventId"`
	MessageID        string              `json:"messageId"`
	ExternalChatID   string              `json:"externalChatId"`
	ExternalThreadID string              `json:"externalThreadId"`
	ExternalUserID   string              `json:"externalUserId"`
	ChatType         string              `json:"chatType"`
	MentionsBot      bool                `json:"mentionsBot"`
	Text             string              `json:"text"`
	Segments         []RichSegment       `json:"segments,omitempty"`
	Attachments      []InboundAttachment `json:"attachments,omitempty"`
	ReplyHandle      map[string]any      `json:"replyHandle,omitempty"`
	RawPayload       []byte              `json:"rawPayload,omitempty"`
	ReceivedAt       time.Time           `json:"receivedAt"`
}

// OutboundEnvelope is the adapter-facing reply payload produced after agent execution.
type OutboundEnvelope struct {
	Platform         string               `json:"platform"`
	ConnectionID     string               `json:"connectionId"`
	EnterpriseID     string               `json:"enterpriseId"`
	ConversationID   string               `json:"conversationId"`
	AgentID          string               `json:"agentId"`
	ExternalChatID   string               `json:"externalChatId"`
	ExternalThreadID string               `json:"externalThreadId"`
	ReplyHandle      map[string]any       `json:"replyHandle,omitempty"`
	Text             string               `json:"text"`
	Segments         []RichSegment        `json:"segments,omitempty"`
	Attachments      []OutboundAttachment `json:"attachments,omitempty"`
}
