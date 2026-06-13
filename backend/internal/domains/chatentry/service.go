package chatentry

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
	"time"

	"dotblue/internal/domains/agent"
	"dotblue/internal/domains/conversation"
	"github.com/google/uuid"
)

type Service struct {
	repo   Repository
	tokens TokenIssuer
}

func NewService(repo Repository, tokens TokenIssuer) *Service {
	return &Service{
		repo:   repo,
		tokens: tokens,
	}
}

// defaultService keeps the same simple assembly style as the existing domains.
var defaultService = NewService(NewGFRepository(), newJWTTokenIssuer(context.Background()))

// GetOrCreateAgentConfig returns the persisted config or a safe default config.
func (s *Service) GetOrCreateAgentConfig(ctx context.Context, enterpriseID, agentID string) (*AgentEntryConfig, error) {
	config, err := s.repo.GetAgentConfig(ctx, enterpriseID, agentID)
	if err != nil {
		return nil, err
	}
	if config != nil {
		return config, nil
	}
	return &AgentEntryConfig{
		EnterpriseID:         enterpriseID,
		AgentID:              agentID,
		Enabled:              false,
		DefaultAccessMode:    AccessModeStandalone,
		ThemeMode:            "auto",
		CompactHeader:        false,
		SessionTTLSeconds:    int(defaultSessionTTL / time.Second),
		RefreshBeforeSeconds: int(defaultRefreshBeforeTTL / time.Second),
	}, nil
}

// UpsertAgentConfig validates admin input and persists a single config row per agent.
func (s *Service) UpsertAgentConfig(ctx context.Context, enterpriseID, agentID, userID string, input AgentEntryConfigInput) (*AgentEntryConfig, error) {
	if ok, err := agent.BelongsToUser(agentID, userID, enterpriseID); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrAgentNotAccessible
	}
	config, err := s.GetOrCreateAgentConfig(ctx, enterpriseID, agentID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(config.ID) == "" {
		config.ID = uuid.NewString()
		config.CreatedBy = userID
	}
	config.Enabled = input.Enabled
	config.DefaultAccessMode = normalizeAccessMode(input.DefaultAccessMode)
	config.AllowAnonymous = input.AllowAnonymous
	config.AllowFileUpload = input.AllowFileUpload
	config.ThemeMode = normalizeThemeMode(input.ThemeMode)
	config.CompactHeader = input.CompactHeader
	config.SessionTTLSeconds = normalizePositiveInt(input.SessionTTLSeconds, int(defaultSessionTTL/time.Second))
	config.RefreshBeforeSeconds = normalizePositiveInt(input.RefreshBeforeSeconds, int(defaultRefreshBeforeTTL/time.Second))
	if err := s.repo.UpsertAgentConfig(ctx, config); err != nil {
		return nil, err
	}
	return s.repo.GetAgentConfig(ctx, enterpriseID, agentID)
}

func (s *Service) ListShareLinks(ctx context.Context, enterpriseID, agentID, userID string) ([]*ShareLink, error) {
	if ok, err := agent.BelongsToUser(agentID, userID, enterpriseID); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrAgentNotAccessible
	}
	links, err := s.repo.ListShareLinks(ctx, enterpriseID, agentID)
	if err != nil {
		return nil, err
	}
	for _, link := range links {
		if link != nil {
			link.HasPassword = strings.TrimSpace(link.PasswordHash) != ""
		}
	}
	return links, nil
}

// CreateShareLink creates a constrained share target for one agent and optionally one conversation.
func (s *Service) CreateShareLink(ctx context.Context, enterpriseID, userID string, input CreateShareLinkInput) (*ShareLink, error) {
	if ok, err := agent.BelongsToUser(input.AgentID, userID, enterpriseID); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrAgentNotAccessible
	}
	if strings.TrimSpace(input.ConversationID) != "" {
		if ok, err := conversation.BelongsToUser(input.ConversationID, userID, enterpriseID); err != nil {
			return nil, err
		} else if !ok {
			return nil, ErrConversationNotAllowed
		}
	}
	now := time.Now()
	link := &ShareLink{
		ID:                uuid.NewString(),
		EnterpriseID:      enterpriseID,
		AgentID:           strings.TrimSpace(input.AgentID),
		ConversationID:    strings.TrimSpace(input.ConversationID),
		ShareCode:         newShareCode(),
		PasswordHash:      hashPassword(strings.TrimSpace(input.Password)),
		Status:            ShareStatusActive,
		AllowContinueChat: input.AllowContinueChat,
		AllowAnonymous:    input.AllowAnonymous,
		MaxAccessCount:    max(input.MaxAccessCount, 0),
		ExpiresAt:         input.ExpiresAt,
		CreatedBy:         userID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := s.repo.CreateShareLink(ctx, link); err != nil {
		return nil, err
	}
	link.HasPassword = strings.TrimSpace(link.PasswordHash) != ""
	return link, nil
}

func (s *Service) RevokeShareLink(ctx context.Context, enterpriseID, userID, shareID string) error {
	link, err := s.repo.GetShareLinkByID(ctx, shareID)
	if err != nil {
		return err
	}
	if link == nil || link.EnterpriseID != enterpriseID {
		return ErrShareLinkNotFound
	}
	if ok, err := agent.BelongsToUser(link.AgentID, userID, enterpriseID); err != nil {
		return err
	} else if !ok {
		return ErrAgentNotAccessible
	}
	return s.repo.RevokeShareLink(ctx, shareID)
}

func (s *Service) GetEmbedConfig(ctx context.Context, enterpriseID, agentID, userID string) (*EmbedConfig, error) {
	if ok, err := agent.BelongsToUser(agentID, userID, enterpriseID); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrAgentNotAccessible
	}
	config, err := s.repo.GetEmbedConfigByAgent(ctx, enterpriseID, agentID)
	if err != nil {
		return nil, err
	}
	if config != nil {
		return config, nil
	}
	return &EmbedConfig{
		EnterpriseID:    enterpriseID,
		AgentID:         agentID,
		AllowedOrigins:  []string{},
		ThemeMode:       "auto",
		CompactHeader:   true,
		AllowFileUpload: false,
	}, nil
}

func (s *Service) UpsertEmbedConfig(ctx context.Context, enterpriseID, agentID, userID string, input EmbedConfigInput) (*EmbedConfig, error) {
	if ok, err := agent.BelongsToUser(agentID, userID, enterpriseID); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrAgentNotAccessible
	}
	current, err := s.repo.GetEmbedConfigByAgent(ctx, enterpriseID, agentID)
	if err != nil {
		return nil, err
	}
	if current == nil {
		current = &EmbedConfig{
			ID:           uuid.NewString(),
			EnterpriseID: enterpriseID,
			AgentID:      agentID,
			CreatedBy:    userID,
		}
	}
	current.AllowedOrigins = normalizeOrigins(input.AllowedOrigins)
	current.ThemeMode = normalizeThemeMode(input.ThemeMode)
	current.CompactHeader = input.CompactHeader
	current.AllowFileUpload = input.AllowFileUpload
	if err := s.repo.UpsertEmbedConfig(ctx, current); err != nil {
		return nil, err
	}
	return s.repo.GetEmbedConfigByAgent(ctx, enterpriseID, agentID)
}

func (s *Service) ResolveShare(ctx context.Context, shareCode string) (*ShareResolveResponse, error) {
	link, err := s.repo.GetShareLinkByCode(ctx, shareCode)
	if err != nil {
		return nil, err
	}
	if err := validateShareLink(link); err != nil {
		return nil, err
	}
	agentRec, err := agent.GetById(link.AgentID)
	if err != nil {
		return nil, err
	}
	if agentRec == nil {
		return nil, ErrAgentNotAccessible
	}
	config, err := s.GetOrCreateAgentConfig(ctx, link.EnterpriseID, link.AgentID)
	if err != nil {
		return nil, err
	}
	result := &ShareResolveResponse{
		ShareID:          link.ID,
		Status:           link.Status,
		RequiresPassword: link.PasswordHash != "",
	}
	result.Agent.ID = agentRec.Id
	result.Agent.AgentName = agentRec.AgentName
	if link.ConversationID != "" {
		if conv, err := conversation.GetById(link.ConversationID); err == nil && conv != nil {
			result.Conversation = &struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			}{
				ID:    conv.Id,
				Title: conv.Title,
			}
		}
	}
	result.Permissions.CanContinueChat = link.AllowContinueChat
	result.Permissions.CanUploadFile = config.AllowFileUpload && link.AllowContinueChat
	result.Permissions.ReadOnly = !link.AllowContinueChat
	return result, nil
}

// VerifyShareAccess exchanges a share link for a short-lived session token.
func (s *Service) VerifyShareAccess(ctx context.Context, shareCode string, input ShareVerifyInput) (*ShareVerifyResponse, error) {
	link, err := s.repo.GetShareLinkByCode(ctx, shareCode)
	if err != nil {
		return nil, err
	}
	if err := validateShareLink(link); err != nil {
		return nil, err
	}
	if link.PasswordHash != "" {
		password := strings.TrimSpace(input.Password)
		if password == "" {
			return nil, ErrSharePasswordRequired
		}
		if hashPassword(password) != link.PasswordHash {
			return nil, ErrSharePasswordInvalid
		}
	}
	config, err := s.GetOrCreateAgentConfig(ctx, link.EnterpriseID, link.AgentID)
	if err != nil {
		return nil, err
	}
	proxyUserID, err := s.resolveAgentProxyUserID(link.AgentID, link.EnterpriseID)
	if err != nil {
		return nil, err
	}
	scopes := []string{ScopeChatRead}
	if link.AllowContinueChat {
		scopes = append(scopes, ScopeChatWrite)
		if config.AllowFileUpload {
			scopes = append(scopes, ScopeFileUpload)
		}
	}
	signed, exp, err := s.tokens.IssueSessionToken(ctx, SessionTokenClaims{
		EnterpriseID:   link.EnterpriseID,
		AgentID:        link.AgentID,
		ConversationID: link.ConversationID,
		ProxyUserID:    proxyUserID,
		SubjectType:    "share_visitor",
		SubjectID:      link.ID,
		Scopes:         scopes,
		ShareLinkID:    link.ID,
		AccessMode:     AccessModeShare,
		ReadOnly:       !link.AllowContinueChat,
	})
	if err != nil {
		return nil, err
	}
	_ = s.repo.IncrementShareAccess(ctx, link.ID)
	return &ShareVerifyResponse{
		SessionToken:         signed,
		ExpiresInSeconds:     int(time.Until(exp).Seconds()),
		RefreshBeforeSeconds: config.RefreshBeforeSeconds,
		AllowFileUpload:      config.AllowFileUpload && link.AllowContinueChat,
	}, nil
}

func (s *Service) CreateEmbedToken(ctx context.Context, enterpriseID, agentID, userID, origin string) (string, int, error) {
	if ok, err := agent.BelongsToUser(agentID, userID, enterpriseID); err != nil {
		return "", 0, err
	} else if !ok {
		return "", 0, ErrAgentNotAccessible
	}
	config, err := s.repo.GetEmbedConfigByAgent(ctx, enterpriseID, agentID)
	if err != nil {
		return "", 0, err
	}
	if config == nil {
		return "", 0, ErrEmbedConfigNotFound
	}
	if origin != "" && !originAllowed(config.AllowedOrigins, origin) {
		return "", 0, ErrEmbedOriginNotAllowed
	}
	proxyUserID, err := s.resolveAgentProxyUserID(agentID, enterpriseID)
	if err != nil {
		return "", 0, err
	}
	token, exp, err := s.tokens.IssueEmbedToken(ctx, EmbedTokenClaims{
		EnterpriseID:  enterpriseID,
		AgentID:       agentID,
		ProxyUserID:   proxyUserID,
		Origin:        origin,
		EmbedConfigID: config.ID,
	})
	if err != nil {
		return "", 0, err
	}
	return token, int(time.Until(exp).Seconds()), nil
}

func (s *Service) CreateStandaloneSession(ctx context.Context, agentID string) (*StandaloneSessionCreateResponse, error) {
	agentRec, err := agent.GetById(agentID)
	if err != nil {
		return nil, err
	}
	if agentRec == nil || strings.TrimSpace(agentRec.GroupId) == "" {
		return nil, ErrAgentNotAccessible
	}
	config, err := s.GetOrCreateAgentConfig(ctx, agentRec.GroupId, agentID)
	if err != nil {
		return nil, err
	}
	if config == nil || !config.Enabled {
		return nil, ErrStandaloneNotEnabled
	}
	if !config.AllowAnonymous {
		return nil, ErrAnonymousAccessDisabled
	}
	proxyUserID, err := s.resolveAgentProxyUserID(agentID, agentRec.GroupId)
	if err != nil {
		return nil, err
	}
	scopes := []string{ScopeChatRead, ScopeChatWrite}
	if config.AllowFileUpload {
		scopes = append(scopes, ScopeFileUpload)
	}
	signed, exp, err := s.tokens.IssueSessionToken(ctx, SessionTokenClaims{
		EnterpriseID: agentRec.GroupId,
		AgentID:      agentID,
		ProxyUserID:  proxyUserID,
		SubjectType:  "standalone_visitor",
		SubjectID:    agentID,
		Scopes:       scopes,
		AccessMode:   AccessModeStandalone,
		ReadOnly:     false,
	})
	if err != nil {
		return nil, err
	}
	return &StandaloneSessionCreateResponse{
		SessionToken:         signed,
		ExpiresInSeconds:     int(time.Until(exp).Seconds()),
		RefreshBeforeSeconds: config.RefreshBeforeSeconds,
		AllowFileUpload:      config.AllowFileUpload,
		AgentName:            agentRec.AgentName,
	}, nil
}

func (s *Service) ExchangeEmbedSession(ctx context.Context, input EmbedSessionExchangeInput, requestOrigin string) (*EmbedSessionExchangeResponse, error) {
	claims, err := s.tokens.ParseEmbedToken(ctx, strings.TrimSpace(input.EmbedToken))
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.AgentID) != claims.AgentID {
		return nil, ErrEmbedOriginNotAllowed
	}
	config, err := s.repo.GetEmbedConfigByAgent(ctx, claims.EnterpriseID, claims.AgentID)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, ErrEmbedConfigNotFound
	}
	if !originAllowed(config.AllowedOrigins, requestOrigin) && !originAllowed(config.AllowedOrigins, claims.Origin) {
		return nil, ErrEmbedOriginNotAllowed
	}
	scopes := []string{ScopeChatRead, ScopeChatWrite}
	if config.AllowFileUpload {
		scopes = append(scopes, ScopeFileUpload)
	}
	signed, exp, err := s.tokens.IssueSessionToken(ctx, SessionTokenClaims{
		EnterpriseID:  claims.EnterpriseID,
		AgentID:       claims.AgentID,
		ProxyUserID:   claims.ProxyUserID,
		SubjectType:   "embed_visitor",
		SubjectID:     claims.EmbedConfigID,
		Scopes:        scopes,
		Origin:        coalesceNonEmpty(requestOrigin, claims.Origin),
		EmbedConfigID: claims.EmbedConfigID,
		AccessMode:    AccessModeEmbed,
		ReadOnly:      false,
	})
	if err != nil {
		return nil, err
	}
	cfg, err := s.GetOrCreateAgentConfig(ctx, claims.EnterpriseID, claims.AgentID)
	if err != nil {
		return nil, err
	}
	return &EmbedSessionExchangeResponse{
		SessionToken:         signed,
		ExpiresInSeconds:     int(time.Until(exp).Seconds()),
		RefreshBeforeSeconds: cfg.RefreshBeforeSeconds,
		AllowFileUpload:      config.AllowFileUpload,
	}, nil
}

func (s *Service) RefreshSession(ctx context.Context, input SessionRefreshInput) (*ShareVerifyResponse, error) {
	claims, err := s.tokens.ParseSessionToken(ctx, strings.TrimSpace(input.SessionToken))
	if err != nil {
		return nil, err
	}
	config, err := s.GetOrCreateAgentConfig(ctx, claims.EnterpriseID, claims.AgentID)
	if err != nil {
		return nil, err
	}
	if time.Until(time.Unix(claims.ExpiresAt, 0)) > time.Duration(config.RefreshBeforeSeconds)*time.Second {
		return nil, ErrSessionTokenInvalid
	}
	signed, exp, err := s.tokens.IssueSessionToken(ctx, *claims)
	if err != nil {
		return nil, err
	}
	return &ShareVerifyResponse{
		SessionToken:         signed,
		ExpiresInSeconds:     int(time.Until(exp).Seconds()),
		RefreshBeforeSeconds: config.RefreshBeforeSeconds,
		AllowFileUpload:      hasScope(claims.Scopes, ScopeFileUpload),
	}, nil
}

func (s *Service) ParseSession(ctx context.Context, raw string) (*SessionContext, error) {
	claims, err := s.tokens.ParseSessionToken(ctx, raw)
	if err != nil {
		return nil, err
	}
	scopes := make(map[string]struct{}, len(claims.Scopes))
	for _, scope := range claims.Scopes {
		scopes[scope] = struct{}{}
	}
	return &SessionContext{
		EnterpriseID:   claims.EnterpriseID,
		AgentID:        claims.AgentID,
		ConversationID: claims.ConversationID,
		ProxyUserID:    claims.ProxyUserID,
		AccessMode:     claims.AccessMode,
		ReadOnly:       claims.ReadOnly,
		Scopes:         scopes,
		ShareLinkID:    claims.ShareLinkID,
		EmbedConfigID:  claims.EmbedConfigID,
		Origin:         claims.Origin,
		ExpiresAt:      time.Unix(claims.ExpiresAt, 0),
	}, nil
}

func validateShareLink(link *ShareLink) error {
	if link == nil {
		return ErrShareLinkNotFound
	}
	if link.Status == ShareStatusRevoked {
		return ErrShareLinkRevoked
	}
	if link.ExpiresAt != nil && link.ExpiresAt.Before(time.Now()) {
		return ErrShareLinkExpired
	}
	if link.MaxAccessCount > 0 && link.AccessCount >= link.MaxAccessCount {
		return ErrShareLinkExpired
	}
	return nil
}

func normalizeAccessMode(value string) string {
	switch strings.TrimSpace(value) {
	case AccessModeShare, AccessModeEmbed:
		return strings.TrimSpace(value)
	default:
		return AccessModeStandalone
	}
}

func normalizeThemeMode(value string) string {
	switch strings.TrimSpace(value) {
	case "light", "dark":
		return strings.TrimSpace(value)
	default:
		return "auto"
	}
}

func normalizePositiveInt(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func normalizeOrigins(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		normalized := strings.TrimSpace(strings.ToLower(value))
		if normalized != "" && !slices.Contains(result, normalized) {
			result = append(result, normalized)
		}
	}
	return result
}

func originAllowed(allowed []string, origin string) bool {
	if strings.TrimSpace(origin) == "" {
		return false
	}
	normalized := strings.TrimSpace(strings.ToLower(origin))
	return slices.Contains(allowed, normalized)
}

func hashPassword(value string) string {
	if value == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func newShareCode() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}

func derivePublicProxyUserID(enterpriseID, agentID, accessMode, subjectID string) string {
	return fmt.Sprintf("chatentry:%s:%s:%s:%s", enterpriseID, agentID, accessMode, subjectID)
}

func coalesceNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func hasScope(scopes []string, target string) bool {
	for _, scope := range scopes {
		if scope == target {
			return true
		}
	}
	return false
}

func (s *Service) resolveAgentProxyUserID(agentID, enterpriseID string) (string, error) {
	agentRec, err := agent.GetById(agentID)
	if err != nil {
		return "", err
	}
	if agentRec == nil || agentRec.GroupId != enterpriseID {
		return "", ErrAgentNotAccessible
	}
	if strings.TrimSpace(agentRec.UserId) != "" {
		return agentRec.UserId, nil
	}
	return derivePublicProxyUserID(enterpriseID, agentID, "agent_owner", agentID), nil
}
