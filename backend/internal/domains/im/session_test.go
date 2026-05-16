package im

import "testing"

func TestBuildSessionKey(t *testing.T) {
	t.Parallel()

	addr := SessionAddress{
		Platform:     "feishu",
		EnterpriseID: "ent_1",
		ConnectionID: "conn/1",
		ChatID:       "chat/1",
		ThreadID:     "thread/1",
		UserID:       "user 1",
		ChatType:     "group",
	}

	if got := BuildSessionKey("agent/main", SessionStrategyPerThread, addr); got != "agent:agent_main:feishu:conn_1:thread:thread_1" {
		t.Fatalf("BuildSessionKey() = %q, want sanitized thread key", got)
	}

	addr.ThreadID = ""
	if got := BuildSessionKey("agent/main", "", addr); got != "agent:agent_main:feishu:conn_1:chat_user:chat_1:user_1" {
		t.Fatalf("BuildSessionKey() default group = %q, want per_chat_per_user key", got)
	}

	addr.ChatType = "p2p"
	if got := BuildSessionKey("agent/main", "", addr); got != "agent:agent_main:feishu:conn_1:user:user_1" {
		t.Fatalf("BuildSessionKey() default dm = %q, want per_user key", got)
	}
}

func TestBuildSessionAddress(t *testing.T) {
	t.Parallel()

	conn := Connection{
		ID:           "conn_1",
		EnterpriseID: "ent_1",
		Platform:     "feishu",
	}
	event := InboundEvent{
		ExternalChatID:   "chat_1",
		ExternalThreadID: "thread_1",
		ExternalUserID:   "user_1",
		ChatType:         "group",
	}

	addr := BuildSessionAddress(conn, event)
	if addr.ConnectionID != "conn_1" || addr.Platform != "feishu" || addr.ThreadID != "thread_1" {
		t.Fatalf("BuildSessionAddress() = %+v, want connection/platform/thread populated", addr)
	}
}
