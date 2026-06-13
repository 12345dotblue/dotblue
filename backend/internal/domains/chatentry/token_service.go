package chatentry

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

const (
	tokenIssuerName         = "dotblue-chatentry"
	sessionTokenAudience    = "c-end-chat"
	embedTokenAudience      = "c-end-chat-embed"
	defaultSessionTTL       = 15 * time.Minute
	defaultEmbedTTL         = 5 * time.Minute
	defaultRefreshBeforeTTL = 2 * time.Minute
)

type TokenIssuer interface {
	IssueSessionToken(ctx context.Context, claims SessionTokenClaims) (string, time.Time, error)
	ParseSessionToken(ctx context.Context, raw string) (*SessionTokenClaims, error)
	IssueEmbedToken(ctx context.Context, claims EmbedTokenClaims) (string, time.Time, error)
	ParseEmbedToken(ctx context.Context, raw string) (*EmbedTokenClaims, error)
}

type jwtTokenIssuer struct {
	secret []byte
}

func newJWTTokenIssuer(ctx context.Context) TokenIssuer {
	secret := strings.TrimSpace(g.Cfg().MustGet(ctx, "casdoor.jwtSecret").String())
	return &jwtTokenIssuer{
		secret: []byte("chatentry::" + secret),
	}
}

func (s *jwtTokenIssuer) IssueSessionToken(ctx context.Context, claims SessionTokenClaims) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(defaultSessionTTL)
	if claims.Issuer == "" {
		claims.Issuer = tokenIssuerName
	}
	if claims.Audience == "" {
		claims.Audience = sessionTokenAudience
	}
	if claims.JTI == "" {
		claims.JTI = uuid.NewString()
	}
	claims.IssuedAt = now.Unix()
	claims.ExpiresAt = exp.Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":            claims.Issuer,
		"aud":            claims.Audience,
		"enterpriseId":   claims.EnterpriseID,
		"agentId":        claims.AgentID,
		"conversationId": claims.ConversationID,
		"proxyUserId":    claims.ProxyUserID,
		"subjectType":    claims.SubjectType,
		"subjectId":      claims.SubjectID,
		"scopes":         claims.Scopes,
		"origin":         claims.Origin,
		"shareLinkId":    claims.ShareLinkID,
		"embedConfigId":  claims.EmbedConfigID,
		"accessMode":     claims.AccessMode,
		"readonly":       claims.ReadOnly,
		"jti":            claims.JTI,
		"iat":            claims.IssuedAt,
		"exp":            claims.ExpiresAt,
	})
	signed, err := token.SignedString(s.secret)
	return signed, exp, err
}

func (s *jwtTokenIssuer) ParseSessionToken(ctx context.Context, raw string) (*SessionTokenClaims, error) {
	token, err := jwt.Parse(raw, func(token *jwt.Token) (interface{}, error) {
		return s.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrSessionTokenExpired
		}
		return nil, ErrSessionTokenInvalid
	}
	if !token.Valid {
		return nil, ErrSessionTokenInvalid
	}
	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrSessionTokenInvalid
	}
	claims := mapSessionClaims(mapClaims)
	if claims.Audience != sessionTokenAudience || claims.Issuer != tokenIssuerName {
		return nil, ErrSessionTokenInvalid
	}
	if time.Unix(claims.ExpiresAt, 0).Before(time.Now()) {
		return nil, ErrSessionTokenExpired
	}
	return claims, nil
}

func (s *jwtTokenIssuer) IssueEmbedToken(ctx context.Context, claims EmbedTokenClaims) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(defaultEmbedTTL)
	if claims.Issuer == "" {
		claims.Issuer = tokenIssuerName
	}
	if claims.Audience == "" {
		claims.Audience = embedTokenAudience
	}
	if claims.JTI == "" {
		claims.JTI = uuid.NewString()
	}
	claims.IssuedAt = now.Unix()
	claims.ExpiresAt = exp.Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":           claims.Issuer,
		"aud":           claims.Audience,
		"enterpriseId":  claims.EnterpriseID,
		"agentId":       claims.AgentID,
		"proxyUserId":   claims.ProxyUserID,
		"origin":        claims.Origin,
		"embedConfigId": claims.EmbedConfigID,
		"jti":           claims.JTI,
		"iat":           claims.IssuedAt,
		"exp":           claims.ExpiresAt,
	})
	signed, err := token.SignedString(s.secret)
	return signed, exp, err
}

func (s *jwtTokenIssuer) ParseEmbedToken(ctx context.Context, raw string) (*EmbedTokenClaims, error) {
	token, err := jwt.Parse(raw, func(token *jwt.Token) (interface{}, error) {
		return s.secret, nil
	})
	if err != nil {
		return nil, ErrSessionTokenInvalid
	}
	if !token.Valid {
		return nil, ErrSessionTokenInvalid
	}
	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrSessionTokenInvalid
	}
	claims := &EmbedTokenClaims{
		Issuer:        getStringClaim(mapClaims, "iss"),
		Audience:      getStringClaim(mapClaims, "aud"),
		EnterpriseID:  getStringClaim(mapClaims, "enterpriseId"),
		AgentID:       getStringClaim(mapClaims, "agentId"),
		ProxyUserID:   getStringClaim(mapClaims, "proxyUserId"),
		Origin:        getStringClaim(mapClaims, "origin"),
		EmbedConfigID: getStringClaim(mapClaims, "embedConfigId"),
		JTI:           getStringClaim(mapClaims, "jti"),
		IssuedAt:      getInt64Claim(mapClaims, "iat"),
		ExpiresAt:     getInt64Claim(mapClaims, "exp"),
	}
	if claims.Audience != embedTokenAudience || claims.Issuer != tokenIssuerName {
		return nil, ErrSessionTokenInvalid
	}
	if time.Unix(claims.ExpiresAt, 0).Before(time.Now()) {
		return nil, ErrSessionTokenExpired
	}
	return claims, nil
}

func mapSessionClaims(claims jwt.MapClaims) *SessionTokenClaims {
	return &SessionTokenClaims{
		Issuer:         getStringClaim(claims, "iss"),
		Audience:       getStringClaim(claims, "aud"),
		EnterpriseID:   getStringClaim(claims, "enterpriseId"),
		AgentID:        getStringClaim(claims, "agentId"),
		ConversationID: getStringClaim(claims, "conversationId"),
		ProxyUserID:    getStringClaim(claims, "proxyUserId"),
		SubjectType:    getStringClaim(claims, "subjectType"),
		SubjectID:      getStringClaim(claims, "subjectId"),
		Scopes:         getStringSliceClaim(claims, "scopes"),
		Origin:         getStringClaim(claims, "origin"),
		ShareLinkID:    getStringClaim(claims, "shareLinkId"),
		EmbedConfigID:  getStringClaim(claims, "embedConfigId"),
		AccessMode:     getStringClaim(claims, "accessMode"),
		ReadOnly:       getBoolClaim(claims, "readonly"),
		JTI:            getStringClaim(claims, "jti"),
		IssuedAt:       getInt64Claim(claims, "iat"),
		ExpiresAt:      getInt64Claim(claims, "exp"),
	}
}

func getStringClaim(claims jwt.MapClaims, key string) string {
	if value, ok := claims[key].(string); ok {
		return value
	}
	return ""
}

func getInt64Claim(claims jwt.MapClaims, key string) int64 {
	switch value := claims[key].(type) {
	case float64:
		return int64(value)
	case int64:
		return value
	case int:
		return int64(value)
	default:
		return 0
	}
}

func getBoolClaim(claims jwt.MapClaims, key string) bool {
	if value, ok := claims[key].(bool); ok {
		return value
	}
	return false
}

func getStringSliceClaim(claims jwt.MapClaims, key string) []string {
	values, ok := claims[key].([]interface{})
	if !ok {
		return nil
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if item, ok := value.(string); ok && item != "" {
			result = append(result, item)
		}
	}
	return result
}
