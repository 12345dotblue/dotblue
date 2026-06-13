package chatentry

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type GFRepository struct{}

func NewGFRepository() *GFRepository {
	return &GFRepository{}
}

func isNoRowsError(err error) bool {
	return err != nil && (errors.Is(err, sql.ErrNoRows) || strings.Contains(strings.ToLower(err.Error()), "no rows in result set"))
}

func (r *GFRepository) GetAgentConfig(ctx context.Context, enterpriseID, agentID string) (*AgentEntryConfig, error) {
	var config AgentEntryConfig
	err := g.DB().Model("chat_entry_agent_configs").Ctx(ctx).
		Where("enterprise_id = ? AND agent_id = ?", enterpriseID, agentID).
		Scan(&config)
	if err != nil {
		if isNoRowsError(err) {
			return nil, nil
		}
		return nil, err
	}
	if config.ID == "" {
		return nil, nil
	}
	return &config, nil
}

func (r *GFRepository) UpsertAgentConfig(ctx context.Context, config *AgentEntryConfig) error {
	_, err := g.DB().Model("chat_entry_agent_configs").Ctx(ctx).Data(g.Map{
		"id":                     config.ID,
		"enterprise_id":          config.EnterpriseID,
		"agent_id":               config.AgentID,
		"enabled":                config.Enabled,
		"default_access_mode":    config.DefaultAccessMode,
		"allow_anonymous":        config.AllowAnonymous,
		"allow_file_upload":      config.AllowFileUpload,
		"theme_mode":             config.ThemeMode,
		"compact_header":         config.CompactHeader,
		"session_ttl_seconds":    config.SessionTTLSeconds,
		"refresh_before_seconds": config.RefreshBeforeSeconds,
		"created_by":             config.CreatedBy,
		"updated_at":             time.Now(),
	}).OnConflict("enterprise_id, agent_id").Save()
	return err
}

func (r *GFRepository) ListShareLinks(ctx context.Context, enterpriseID, agentID string) ([]*ShareLink, error) {
	var links []*ShareLink
	if err := g.DB().Model("chat_entry_share_links").Ctx(ctx).
		Where("enterprise_id = ? AND agent_id = ?", enterpriseID, agentID).
		Order("created_at DESC").
		Scan(&links); err != nil {
		return nil, err
	}
	return links, nil
}

func (r *GFRepository) CreateShareLink(ctx context.Context, link *ShareLink) error {
	var conversationID any
	if strings.TrimSpace(link.ConversationID) != "" {
		conversationID = link.ConversationID
	}
	_, err := g.DB().Model("chat_entry_share_links").Ctx(ctx).Data(g.Map{
		"id":                  link.ID,
		"enterprise_id":       link.EnterpriseID,
		"agent_id":            link.AgentID,
		"conversation_id":     conversationID,
		"share_code":          link.ShareCode,
		"password_hash":       link.PasswordHash,
		"status":              link.Status,
		"allow_continue_chat": link.AllowContinueChat,
		"allow_anonymous":     link.AllowAnonymous,
		"max_access_count":    link.MaxAccessCount,
		"access_count":        link.AccessCount,
		"expires_at":          link.ExpiresAt,
		"revoked_at":          link.RevokedAt,
		"created_by":          link.CreatedBy,
	}).Insert()
	return err
}

func (r *GFRepository) GetShareLinkByCode(ctx context.Context, shareCode string) (*ShareLink, error) {
	var link ShareLink
	err := g.DB().Model("chat_entry_share_links").Ctx(ctx).
		Where("share_code = ?", shareCode).
		Scan(&link)
	if err != nil {
		if isNoRowsError(err) {
			return nil, nil
		}
		return nil, err
	}
	if link.ID == "" {
		return nil, nil
	}
	return &link, nil
}

func (r *GFRepository) GetShareLinkByID(ctx context.Context, id string) (*ShareLink, error) {
	var link ShareLink
	err := g.DB().Model("chat_entry_share_links").Ctx(ctx).
		Where("id = ?", id).
		Scan(&link)
	if err != nil {
		if isNoRowsError(err) {
			return nil, nil
		}
		return nil, err
	}
	if link.ID == "" {
		return nil, nil
	}
	return &link, nil
}

func (r *GFRepository) RevokeShareLink(ctx context.Context, id string) error {
	_, err := g.DB().Model("chat_entry_share_links").Ctx(ctx).
		Data(g.Map{
			"status":     ShareStatusRevoked,
			"revoked_at": time.Now(),
			"updated_at": time.Now(),
		}).
		Where("id = ?", id).
		Update()
	return err
}

func (r *GFRepository) IncrementShareAccess(ctx context.Context, id string) error {
	_, err := g.DB().Model("chat_entry_share_links").Ctx(ctx).
		Data("access_count = access_count + 1, updated_at = NOW()").
		Where("id = ?", id).
		Update()
	return err
}

func (r *GFRepository) GetEmbedConfigByAgent(ctx context.Context, enterpriseID, agentID string) (*EmbedConfig, error) {
	var config EmbedConfig
	err := g.DB().Model("chat_entry_embed_configs").Ctx(ctx).
		Where("enterprise_id = ? AND agent_id = ?", enterpriseID, agentID).
		Scan(&config)
	if err != nil {
		if isNoRowsError(err) {
			return nil, nil
		}
		return nil, err
	}
	if config.ID == "" {
		return nil, nil
	}
	if config.AllowedOriginsJSON != "" {
		_ = json.Unmarshal([]byte(config.AllowedOriginsJSON), &config.AllowedOrigins)
	}
	return &config, nil
}

func (r *GFRepository) UpsertEmbedConfig(ctx context.Context, config *EmbedConfig) error {
	allowedOriginsJSON, err := json.Marshal(config.AllowedOrigins)
	if err != nil {
		return err
	}
	_, err = g.DB().Model("chat_entry_embed_configs").Ctx(ctx).Data(g.Map{
		"id":                   config.ID,
		"enterprise_id":        config.EnterpriseID,
		"agent_id":             config.AgentID,
		"allowed_origins_json": string(allowedOriginsJSON),
		"theme_mode":           config.ThemeMode,
		"compact_header":       config.CompactHeader,
		"allow_file_upload":    config.AllowFileUpload,
		"created_by":           config.CreatedBy,
		"updated_at":           time.Now(),
	}).OnConflict("enterprise_id, agent_id").Save()
	return err
}

func (r *GFRepository) SaveAccessLog(ctx context.Context, log *AccessLog) error {
	riskFlagsJSON, err := json.Marshal(log.RiskFlags)
	if err != nil {
		return err
	}
	_, err = g.DB().Model("chat_entry_access_logs").Ctx(ctx).Data(g.Map{
		"id":              log.ID,
		"channel_type":    log.ChannelType,
		"target_id":       log.TargetID,
		"enterprise_id":   log.EnterpriseID,
		"agent_id":        log.AgentID,
		"origin":          log.Origin,
		"referer":         log.Referer,
		"ip_hash":         log.IPHash,
		"user_agent":      log.UserAgent,
		"trace_id":        log.TraceID,
		"result":          log.Result,
		"risk_flags_json": string(riskFlagsJSON),
	}).Insert()
	return err
}
