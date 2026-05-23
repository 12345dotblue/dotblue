package gateway

import (
	"testing"
	"time"
)

func TestBuildDispatchRequestFillsDefaults(t *testing.T) {
	before := time.Now()
	req := BuildDispatchRequest(DispatchBuildInput{
		RequestID:      " req-1 ",
		ConversationID: " conv-1 ",
		AgentID:        " agent-1 ",
		Content:        " hello ",
	})
	after := time.Now()

	if req.RequestID != "req-1" {
		t.Fatalf("RequestID = %q, want req-1", req.RequestID)
	}
	if req.SessionKey != "conv-1" {
		t.Fatalf("SessionKey = %q, want conv-1", req.SessionKey)
	}
	if req.AgentID != "agent-1" {
		t.Fatalf("AgentID = %q, want agent-1", req.AgentID)
	}
	if req.Content != "hello" {
		t.Fatalf("Content = %q, want hello", req.Content)
	}
	if req.CreatedAt.Before(before) || req.CreatedAt.After(after) {
		t.Fatalf("CreatedAt = %v, want within [%v, %v]", req.CreatedAt, before, after)
	}
}

func TestBuildDispatchRequestKeepsExplicitValues(t *testing.T) {
	createdAt := time.Unix(123, 0)
	replyHandle := map[string]any{"k": "v"}
	req := BuildDispatchRequest(DispatchBuildInput{
		RequestID:        "req-2",
		SessionKey:       "sess-2",
		EnterpriseID:     "ent-1",
		UserID:           "user-1",
		AgentID:          "agent-2",
		ConversationID:   "conv-2",
		ConnectionID:     "conn-2",
		IngressType:      "feishu",
		IngressID:        "runtime-1",
		InboundMessageID: "msg-2",
		ExternalChatID:   "chat-2",
		ExternalThreadID: "thread-2",
		ReplyHandle:      replyHandle,
		Content:          "hello",
		CreatedAt:        createdAt,
	})

	if req.SessionKey != "sess-2" {
		t.Fatalf("SessionKey = %q, want sess-2", req.SessionKey)
	}
	if req.CreatedAt != createdAt {
		t.Fatalf("CreatedAt = %v, want %v", req.CreatedAt, createdAt)
	}
	if req.ReplyHandle["k"] != "v" {
		t.Fatalf("ReplyHandle = %+v", req.ReplyHandle)
	}
	if req.IngressType != "feishu" || req.IngressID != "runtime-1" {
		t.Fatalf("ingress = %s/%s", req.IngressType, req.IngressID)
	}
}

func TestBuildWebDispatchRequestUsesWebIngressDefaults(t *testing.T) {
	req := BuildWebDispatchRequest(WebIngressInput{
		RequestID:      "req-web",
		ConversationID: "conv-web",
		AgentID:        "agent-web",
		ConnectionID:   "conn-web",
	})

	if req.IngressType != "web" || req.IngressID != "http" {
		t.Fatalf("ingress = %s/%s", req.IngressType, req.IngressID)
	}
	if req.ConnectionID != "conn-web" {
		t.Fatalf("ConnectionID = %q, want conn-web", req.ConnectionID)
	}
}

func TestBuildIMDispatchRequestUsesPlatformAndConnectionAsIngress(t *testing.T) {
	req := BuildIMDispatchRequest(IMIngressInput{
		RequestID:      "req-im",
		ConversationID: "conv-im",
		AgentID:        "agent-im",
		ConnectionID:   "conn-im",
		Platform:       "feishu",
	})

	if req.IngressType != "feishu" || req.IngressID != "conn-im" {
		t.Fatalf("ingress = %s/%s", req.IngressType, req.IngressID)
	}
	if req.SessionKey != "conv-im" {
		t.Fatalf("SessionKey = %q, want conv-im", req.SessionKey)
	}
}
