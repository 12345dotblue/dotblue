package dataplane

import "testing"

func TestKeyspaceKeys(t *testing.T) {
	ks := NewKeyspace("dot")

	cases := map[string]string{
		"session owner":  ks.SessionOwner("sess-1"),
		"session fence":  ks.SessionFence("sess-1"),
		"session gate":   ks.SessionGate("sess-1"),
		"session state":  ks.SessionState("sess-1"),
		"worker meta":    ks.WorkerMeta("worker-1"),
		"worker task":    ks.WorkerTaskStream("worker-1"),
		"request event":  ks.RequestEventChannel("req-1"),
		"request state":  ks.RequestState("req-1"),
		"request route":  ks.RequestRoute("req-1"),
		"outbox":         ks.OutboxConnection("conn-1"),
	}

	want := map[string]string{
		"session owner": "dot:session:owner:sess-1",
		"session fence": "dot:session:fence:sess-1",
		"session gate":  "dot:session:gate:sess-1",
		"session state": "dot:session:state:sess-1",
		"worker meta":   "dot:worker:meta:worker-1",
		"worker task":   "dot:task:worker:worker-1",
		"request event": "dot:req:event:req-1",
		"request state": "dot:req:state:req-1",
		"request route": "dot:req:route:req-1",
		"outbox":        "dot:outbox:conn:conn-1",
	}

	for name, got := range cases {
		if got != want[name] {
			t.Fatalf("%s = %q, want %q", name, got, want[name])
		}
	}
}

