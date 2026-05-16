package im

import "testing"

func TestBuildConnectionCallbackPath(t *testing.T) {
	t.Parallel()

	if got := buildConnectionCallbackPath("feishu", "conn_123"); got != "/api/im/inbound/feishu/conn_123" {
		t.Fatalf("buildConnectionCallbackPath() = %q, want %q", got, "/api/im/inbound/feishu/conn_123")
	}
	if got := buildConnectionCallbackPath("", "conn_123"); got != "" {
		t.Fatalf("buildConnectionCallbackPath() = %q, want empty string", got)
	}
}

func TestNormalizeInboundPayloadJSON(t *testing.T) {
	t.Parallel()

	if got := normalizeInboundPayloadJSON([]byte(`{"event":"ok"}`)); got != `{"event":"ok"}` {
		t.Fatalf("normalizeInboundPayloadJSON() = %q, want original json", got)
	}
	if got := normalizeInboundPayloadJSON(nil); got != "{}" {
		t.Fatalf("normalizeInboundPayloadJSON() = %q, want {}", got)
	}
	if got := normalizeInboundPayloadJSON([]byte("not-json")); got != `{"raw":"not-json"}` {
		t.Fatalf("normalizeInboundPayloadJSON() = %q, want wrapped raw payload", got)
	}
}

func TestIsDuplicatedInboundEventError(t *testing.T) {
	t.Parallel()

	if !isDuplicatedInboundEventError(assertErr("duplicate key value violates unique constraint uk_external_message_events_connection_event")) {
		t.Fatal("expected duplicate constraint error to be detected")
	}
	if isDuplicatedInboundEventError(assertErr("other error")) {
		t.Fatal("unexpected duplicate detection for unrelated error")
	}
}

type staticErr string

func (e staticErr) Error() string {
	return string(e)
}

func assertErr(message string) error {
	return staticErr(message)
}
