package identity

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
)

type SessionContext struct {
	UserID         string
	OrganizationID string
	IsAdmin        bool
	Groups         []string
	Email          string
	DisplayName    string
	Avatar         string
}

type TokenClaims struct {
	UserID         string
	OrganizationID string
	IsAdmin        bool
	Groups         []string
}

type Authenticator interface {
	ParseToken(token string) (*TokenClaims, error)
}

type casdoorAuthenticator struct{}

func (casdoorAuthenticator) ParseToken(token string) (*TokenClaims, error) {
	claims, err := casdoorsdk.ParseJwtToken(token)
	if err != nil {
		return nil, err
	}
	return &TokenClaims{
		UserID:         claims.Name,
		OrganizationID: claims.Owner,
		IsAdmin:        claims.IsAdmin,
		Groups:         claims.Groups,
	}, nil
}

type Service struct {
	authenticator Authenticator
	repo          Repository
	now           func() time.Time
}

func NewService(repo Repository, authenticator Authenticator) *Service {
	return &Service{
		authenticator: authenticator,
		repo:          repo,
		now:           time.Now,
	}
}

var defaultService = NewService(NewGFRepository(), casdoorAuthenticator{})

func (s *Service) ParseSession(tokenString string) (*SessionContext, error) {
	if s == nil || s.authenticator == nil {
		return nil, errors.New("identity authenticator is not configured")
	}
	tokenString = normalizeBearerToken(tokenString)
	if tokenString == "" {
		return nil, errors.New("missing authorization token")
	}
	claims, err := s.authenticator.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}
	payload := decodeTokenPayload(tokenString)
	displayName := payload["displayName"]
	if displayName == "" {
		displayName = payload["name"]
	}
	return &SessionContext{
		UserID:         claims.UserID,
		OrganizationID: claims.OrganizationID,
		IsAdmin:        claims.IsAdmin,
		Groups:         claims.Groups,
		Email:          payload["email"],
		DisplayName:    displayName,
		Avatar:         payload["avatar"],
	}, nil
}

func (s *Service) SyncLocalUser(session *SessionContext) error {
	if s == nil || s.repo == nil || session == nil || session.UserID == "" {
		return nil
	}
	return s.repo.UpsertLocalUser(
		session.UserID,
		session.OrganizationID,
		session.Email,
		session.DisplayName,
		session.Avatar,
		s.now(),
	)
}

func (s *Service) HasAdminAccess(isAdmin bool, groups []string) bool {
	if isAdmin {
		return true
	}
	for _, group := range groups {
		if group == AdminGroup {
			return true
		}
	}
	return false
}

func normalizeBearerToken(tokenString string) string {
	tokenString = strings.TrimSpace(tokenString)
	if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = strings.TrimSpace(tokenString[7:])
	}
	return tokenString
}

func decodeTokenPayload(tokenString string) map[string]string {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return map[string]string{}
	}
	decoded, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return map[string]string{}
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return map[string]string{}
	}
	out := make(map[string]string)
	for _, key := range []string{"email", "displayName", "name", "avatar"} {
		if val, ok := payload[key].(string); ok {
			out[key] = val
		}
	}
	return out
}
