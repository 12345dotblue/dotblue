package chatentry

import "context"

type Repository interface {
	GetAgentConfig(ctx context.Context, enterpriseID, agentID string) (*AgentEntryConfig, error)
	UpsertAgentConfig(ctx context.Context, config *AgentEntryConfig) error

	ListShareLinks(ctx context.Context, enterpriseID, agentID string) ([]*ShareLink, error)
	CreateShareLink(ctx context.Context, link *ShareLink) error
	GetShareLinkByCode(ctx context.Context, shareCode string) (*ShareLink, error)
	GetShareLinkByID(ctx context.Context, id string) (*ShareLink, error)
	RevokeShareLink(ctx context.Context, id string) error
	IncrementShareAccess(ctx context.Context, id string) error

	GetEmbedConfigByAgent(ctx context.Context, enterpriseID, agentID string) (*EmbedConfig, error)
	UpsertEmbedConfig(ctx context.Context, config *EmbedConfig) error

	SaveAccessLog(ctx context.Context, log *AccessLog) error
}
