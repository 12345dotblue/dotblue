package chatentry

import "time"

const (
	AccessModeStandalone = "standalone"
	AccessModeShare      = "share"
	AccessModeEmbed      = "embed"
)

const (
	ScopeChatRead  = "chat:read"
	ScopeChatWrite = "chat:write"
	ScopeFileUpload = "file:upload"
)

const (
	ShareStatusActive  = "active"
	ShareStatusRevoked = "revoked"
)

type AgentEntryConfig struct {
	ID                   string    `json:"id"`
	EnterpriseID         string    `json:"enterpriseId" orm:"enterprise_id"`
	AgentID              string    `json:"agentId" orm:"agent_id"`
	Enabled              bool      `json:"enabled"`
	DefaultAccessMode    string    `json:"defaultAccessMode" orm:"default_access_mode"`
	AllowAnonymous       bool      `json:"allowAnonymous" orm:"allow_anonymous"`
	AllowFileUpload      bool      `json:"allowFileUpload" orm:"allow_file_upload"`
	ThemeMode            string    `json:"themeMode" orm:"theme_mode"`
	CompactHeader        bool      `json:"compactHeader" orm:"compact_header"`
	SessionTTLSeconds    int       `json:"sessionTtlSeconds" orm:"session_ttl_seconds"`
	RefreshBeforeSeconds int       `json:"refreshBeforeSeconds" orm:"refresh_before_seconds"`
	CreatedBy            string    `json:"createdBy" orm:"created_by"`
	CreatedAt            time.Time `json:"createdAt" orm:"created_at"`
	UpdatedAt            time.Time `json:"updatedAt" orm:"updated_at"`
}

type ShareLink struct {
	ID                string     `json:"id"`
	EnterpriseID      string     `json:"enterpriseId" orm:"enterprise_id"`
	AgentID           string     `json:"agentId" orm:"agent_id"`
	ConversationID    string     `json:"conversationId,omitempty" orm:"conversation_id"`
	ShareCode         string     `json:"shareCode" orm:"share_code"`
	PasswordHash      string     `json:"-" orm:"password_hash"`
	Status            string     `json:"status"`
	AllowContinueChat bool       `json:"allowContinueChat" orm:"allow_continue_chat"`
	AllowAnonymous    bool       `json:"allowAnonymous" orm:"allow_anonymous"`
	MaxAccessCount    int        `json:"maxAccessCount" orm:"max_access_count"`
	AccessCount       int        `json:"accessCount" orm:"access_count"`
	ExpiresAt         *time.Time `json:"expiresAt,omitempty" orm:"expires_at"`
	RevokedAt         *time.Time `json:"revokedAt,omitempty" orm:"revoked_at"`
	CreatedBy         string     `json:"createdBy" orm:"created_by"`
	CreatedAt         time.Time  `json:"createdAt" orm:"created_at"`
	UpdatedAt         time.Time  `json:"updatedAt" orm:"updated_at"`
	HasPassword       bool       `json:"hasPassword" orm:"-"`
}

type EmbedConfig struct {
	ID              string    `json:"id"`
	EnterpriseID    string    `json:"enterpriseId" orm:"enterprise_id"`
	AgentID         string    `json:"agentId" orm:"agent_id"`
	AllowedOrigins  []string  `json:"allowedOrigins" orm:"-"`
	AllowedOriginsJSON string `json:"-" orm:"allowed_origins_json"`
	ThemeMode       string    `json:"themeMode" orm:"theme_mode"`
	CompactHeader   bool      `json:"compactHeader" orm:"compact_header"`
	AllowFileUpload bool      `json:"allowFileUpload" orm:"allow_file_upload"`
	CreatedBy       string    `json:"createdBy" orm:"created_by"`
	CreatedAt       time.Time `json:"createdAt" orm:"created_at"`
	UpdatedAt       time.Time `json:"updatedAt" orm:"updated_at"`
}

type AccessLog struct {
	ID           string    `json:"id"`
	ChannelType  string    `json:"channelType" orm:"channel_type"`
	TargetID     string    `json:"targetId" orm:"target_id"`
	EnterpriseID string    `json:"enterpriseId" orm:"enterprise_id"`
	AgentID      string    `json:"agentId" orm:"agent_id"`
	Origin       string    `json:"origin"`
	Referer      string    `json:"referer"`
	IPHash       string    `json:"ipHash" orm:"ip_hash"`
	UserAgent    string    `json:"userAgent" orm:"user_agent"`
	TraceID      string    `json:"traceId" orm:"trace_id"`
	Result       string    `json:"result"`
	RiskFlags    []string  `json:"riskFlags" orm:"-"`
	RiskFlagsJSON string   `json:"-" orm:"risk_flags_json"`
	CreatedAt    time.Time `json:"createdAt" orm:"created_at"`
}

type SessionTokenClaims struct {
	Issuer        string   `json:"iss"`
	Audience      string   `json:"aud"`
	EnterpriseID  string   `json:"enterpriseId"`
	AgentID       string   `json:"agentId"`
	ConversationID string  `json:"conversationId,omitempty"`
	ProxyUserID   string   `json:"proxyUserId"`
	SubjectType   string   `json:"subjectType"`
	SubjectID     string   `json:"subjectId"`
	Scopes        []string `json:"scopes"`
	Origin        string   `json:"origin,omitempty"`
	ShareLinkID   string   `json:"shareLinkId,omitempty"`
	EmbedConfigID string   `json:"embedConfigId,omitempty"`
	AccessMode    string   `json:"accessMode"`
	ReadOnly      bool     `json:"readonly"`
	JTI           string   `json:"jti"`
	IssuedAt      int64    `json:"iat"`
	ExpiresAt     int64    `json:"exp"`
}

type EmbedTokenClaims struct {
	Issuer       string `json:"iss"`
	Audience     string `json:"aud"`
	EnterpriseID string `json:"enterpriseId"`
	AgentID      string `json:"agentId"`
	ProxyUserID  string `json:"proxyUserId"`
	Origin       string `json:"origin"`
	EmbedConfigID string `json:"embedConfigId"`
	JTI          string `json:"jti"`
	IssuedAt     int64  `json:"iat"`
	ExpiresAt    int64  `json:"exp"`
}

type AgentEntryConfigInput struct {
	Enabled              bool   `json:"enabled"`
	DefaultAccessMode    string `json:"defaultAccessMode"`
	AllowAnonymous       bool   `json:"allowAnonymous"`
	AllowFileUpload      bool   `json:"allowFileUpload"`
	ThemeMode            string `json:"themeMode"`
	CompactHeader        bool   `json:"compactHeader"`
	SessionTTLSeconds    int    `json:"sessionTtlSeconds"`
	RefreshBeforeSeconds int    `json:"refreshBeforeSeconds"`
}

type CreateShareLinkInput struct {
	AgentID           string     `json:"agentId"`
	ConversationID    string     `json:"conversationId"`
	Password          string     `json:"password"`
	AllowContinueChat bool       `json:"allowContinueChat"`
	AllowAnonymous    bool       `json:"allowAnonymous"`
	MaxAccessCount    int        `json:"maxAccessCount"`
	ExpiresAt         *time.Time `json:"expiresAt"`
}

type EmbedConfigInput struct {
	AllowedOrigins  []string `json:"allowedOrigins"`
	ThemeMode       string   `json:"themeMode"`
	CompactHeader   bool     `json:"compactHeader"`
	AllowFileUpload bool     `json:"allowFileUpload"`
}

type ShareResolveResponse struct {
	ShareID    string `json:"shareId"`
	Status     string `json:"status"`
	RequiresPassword bool `json:"requiresPassword"`
	Agent      struct {
		ID        string `json:"id"`
		AgentName string `json:"agentName"`
	} `json:"agent"`
	Conversation *struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	} `json:"conversation,omitempty"`
	Permissions struct {
		CanContinueChat bool `json:"canContinueChat"`
		CanUploadFile   bool `json:"canUploadFile"`
		ReadOnly        bool `json:"readOnly"`
	} `json:"permissions"`
}

type ShareVerifyInput struct {
	Password string `json:"password"`
}

type ShareVerifyResponse struct {
	SessionToken      string `json:"sessionToken"`
	ExpiresInSeconds  int    `json:"expiresInSeconds"`
	RefreshBeforeSeconds int `json:"refreshBeforeSeconds"`
	AllowFileUpload   bool   `json:"allowFileUpload"`
}

type EmbedSessionExchangeInput struct {
	EmbedToken string `json:"embedToken"`
	AgentID    string `json:"agentId"`
}

type EmbedSessionExchangeResponse struct {
	SessionToken         string `json:"sessionToken"`
	ExpiresInSeconds     int    `json:"expiresInSeconds"`
	RefreshBeforeSeconds int    `json:"refreshBeforeSeconds"`
	AllowFileUpload      bool   `json:"allowFileUpload"`
}

type StandaloneSessionCreateResponse struct {
	SessionToken         string `json:"sessionToken"`
	ExpiresInSeconds     int    `json:"expiresInSeconds"`
	RefreshBeforeSeconds int    `json:"refreshBeforeSeconds"`
	AllowFileUpload      bool   `json:"allowFileUpload"`
	AgentName            string `json:"agentName"`
}

type SessionRefreshInput struct {
	SessionToken string `json:"sessionToken"`
}

type SessionContext struct {
	EnterpriseID   string
	AgentID        string
	ConversationID string
	ProxyUserID    string
	AccessMode     string
	ReadOnly       bool
	Scopes         map[string]struct{}
	ShareLinkID    string
	EmbedConfigID  string
	Origin         string
	ExpiresAt      time.Time
}
