package gateway

import (
	"strings"
	"time"
)

type DispatchBuildInput struct {
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

type WebIngressInput struct {
	RequestID        string
	SessionKey       string
	EnterpriseID     string
	UserID           string
	AgentID          string
	ConversationID   string
	ConnectionID     string
	InboundMessageID string
	ExternalChatID   string
	ExternalThreadID string
	ReplyHandle      map[string]any
	Content          string
	CreatedAt        time.Time
}

type IMIngressInput struct {
	RequestID        string
	SessionKey       string
	EnterpriseID     string
	UserID           string
	AgentID          string
	ConversationID   string
	ConnectionID     string
	Platform         string
	InboundMessageID string
	ExternalChatID   string
	ExternalThreadID string
	ReplyHandle      map[string]any
	Content          string
	CreatedAt        time.Time
}

func BuildDispatchRequest(input DispatchBuildInput) DispatchRequest {
	sessionKey := strings.TrimSpace(input.SessionKey)
	if sessionKey == "" {
		sessionKey = strings.TrimSpace(input.ConversationID)
	}
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	return DispatchRequest{
		RequestID:        strings.TrimSpace(input.RequestID),
		SessionKey:       sessionKey,
		EnterpriseID:     strings.TrimSpace(input.EnterpriseID),
		UserID:           strings.TrimSpace(input.UserID),
		AgentID:          strings.TrimSpace(input.AgentID),
		ConversationID:   strings.TrimSpace(input.ConversationID),
		ConnectionID:     strings.TrimSpace(input.ConnectionID),
		IngressType:      strings.TrimSpace(input.IngressType),
		IngressID:        strings.TrimSpace(input.IngressID),
		InboundMessageID: strings.TrimSpace(input.InboundMessageID),
		ExternalChatID:   strings.TrimSpace(input.ExternalChatID),
		ExternalThreadID: strings.TrimSpace(input.ExternalThreadID),
		ReplyHandle:      input.ReplyHandle,
		Content:          strings.TrimSpace(input.Content),
		CreatedAt:        createdAt,
	}
}

func BuildWebDispatchRequest(input WebIngressInput) DispatchRequest {
	return BuildDispatchRequest(DispatchBuildInput{
		RequestID:        input.RequestID,
		SessionKey:       input.SessionKey,
		EnterpriseID:     input.EnterpriseID,
		UserID:           input.UserID,
		AgentID:          input.AgentID,
		ConversationID:   input.ConversationID,
		ConnectionID:     input.ConnectionID,
		IngressType:      "web",
		IngressID:        "http",
		InboundMessageID: input.InboundMessageID,
		ExternalChatID:   input.ExternalChatID,
		ExternalThreadID: input.ExternalThreadID,
		ReplyHandle:      input.ReplyHandle,
		Content:          input.Content,
		CreatedAt:        input.CreatedAt,
	})
}

func BuildIMDispatchRequest(input IMIngressInput) DispatchRequest {
	return BuildDispatchRequest(DispatchBuildInput{
		RequestID:        input.RequestID,
		SessionKey:       input.SessionKey,
		EnterpriseID:     input.EnterpriseID,
		UserID:           input.UserID,
		AgentID:          input.AgentID,
		ConversationID:   input.ConversationID,
		ConnectionID:     input.ConnectionID,
		IngressType:      input.Platform,
		IngressID:        input.ConnectionID,
		InboundMessageID: input.InboundMessageID,
		ExternalChatID:   input.ExternalChatID,
		ExternalThreadID: input.ExternalThreadID,
		ReplyHandle:      input.ReplyHandle,
		Content:          input.Content,
		CreatedAt:        input.CreatedAt,
	})
}
