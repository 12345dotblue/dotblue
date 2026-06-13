package chatentry

import "errors"

var (
	ErrAgentNotAccessible      = errors.New("agent is not accessible")
	ErrConfigNotFound          = errors.New("chat entry config not found")
	ErrShareLinkNotFound       = errors.New("share link not found")
	ErrShareLinkExpired        = errors.New("share link expired")
	ErrShareLinkRevoked        = errors.New("share link revoked")
	ErrSharePasswordRequired   = errors.New("share password required")
	ErrSharePasswordInvalid    = errors.New("share password invalid")
	ErrEmbedConfigNotFound     = errors.New("embed config not found")
	ErrEmbedOriginNotAllowed   = errors.New("embed origin not allowed")
	ErrStandaloneNotEnabled    = errors.New("standalone access is not enabled")
	ErrAnonymousAccessDisabled = errors.New("anonymous access is disabled")
	ErrSessionTokenInvalid     = errors.New("session token invalid")
	ErrSessionTokenExpired     = errors.New("session token expired")
	ErrSessionScopeDenied      = errors.New("session scope denied")
	ErrConversationNotAllowed  = errors.New("conversation is not allowed")
)
