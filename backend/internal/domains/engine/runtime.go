package engine

import "context"

// Runtime manages container/sandbox lifecycle for an agent.
type Runtime interface {
	// EnsureRunning ensures the sandbox for the given agent is up, returns its endpoint.
	EnsureRunning(ctx context.Context, orgID, userID, agentID string) (*AgentEndpoint, error)
	// Stop stops the sandbox for the given agent.
	Stop(ctx context.Context, agentID string) error
}

// AgentEndpoint holds the sandbox endpoint URL and auth token.
type AgentEndpoint struct {
	URL    string
	APIKey string
}
