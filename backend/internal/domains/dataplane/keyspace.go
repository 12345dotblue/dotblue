package dataplane

import "strings"

type Keyspace struct {
	Prefix string
}

func NewKeyspace(prefix string) Keyspace {
	return Keyspace{Prefix: strings.TrimSpace(prefix)}
}

func (k Keyspace) key(parts ...string) string {
	out := k.Prefix
	for _, p := range parts {
		if p == "" {
			continue
		}
		out += ":" + p
	}
	return out
}

func (k Keyspace) SessionOwner(sessionKey string) string {
	return k.key("session", "owner", sessionKey)
}

func (k Keyspace) SessionFence(sessionKey string) string {
	return k.key("session", "fence", sessionKey)
}

func (k Keyspace) SessionGate(sessionKey string) string {
	return k.key("session", "gate", sessionKey)
}

func (k Keyspace) SessionState(sessionKey string) string {
	return k.key("session", "state", sessionKey)
}

func (k Keyspace) WorkerMeta(workerID string) string {
	return k.key("worker", "meta", workerID)
}

func (k Keyspace) WorkerMetaPattern() string {
	return k.key("worker", "meta", "*")
}

func (k Keyspace) WorkerTaskStream(workerID string) string {
	return k.key("task", "worker", workerID)
}

func (k Keyspace) RequestEventChannel(requestID string) string {
	return k.key("req", "event", requestID)
}

func (k Keyspace) RequestState(requestID string) string {
	return k.key("req", "state", requestID)
}

func (k Keyspace) RequestRoute(requestID string) string {
	return k.key("req", "route", requestID)
}

func (k Keyspace) OutboxConnection(connectionID string) string {
	return k.key("outbox", "conn", connectionID)
}
