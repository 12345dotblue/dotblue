package identity

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type Config struct {
	Endpoint         string
	ClientId         string
	ClientSecret     string
	JwtSecret        string
	OrganizationName string
	ApplicationName  string
}

func Init(c Config) {
	g.Log().Infof(context.Background(), "Initializing Casdoor SDK with ClientId: %s, JwtSecret length: %d", c.ClientId, len(c.JwtSecret))
	casdoorsdk.InitConfig(
		c.Endpoint,
		c.ClientId,
		c.ClientSecret,
		c.JwtSecret,
		c.OrganizationName,
		c.ApplicationName,
	)
}

func Middleware(r *ghttp.Request) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		tokenString = r.Get("token").String()
	}

	if tokenString == "" {
		g.Log().Warning(r.Context(), "Authorization token is missing (checked header and query)")
		r.Response.WriteStatus(http.StatusUnauthorized, "Missing Authorization token")
		r.ExitAll()
		return
	}

	if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = tokenString[7:]
	}

	claims, err := casdoorsdk.ParseJwtToken(tokenString)
	if err != nil {
		tokenPreview := tokenString
		if len(tokenPreview) > 10 {
			tokenPreview = tokenPreview[:10]
		}
		g.Log().Errorf(r.Context(), "JWT validation failed for token [%s...]: %v", tokenPreview, err)
		msg := g.I18n().T(r.Context(), "auth_invalid")
		r.Response.WriteStatus(http.StatusUnauthorized, msg)
		r.ExitAll()
		return
	}

	payload := decodeTokenPayload(tokenString)
	email := payload["email"]
	displayName := payload["displayName"]
	if displayName == "" {
		displayName = payload["name"]
	}
	avatar := payload["avatar"]

	r.SetCtxVar("userId", claims.Name)
	r.SetCtxVar("organizationId", claims.Owner)
	r.SetCtxVar("isAdmin", claims.IsAdmin)
	r.SetCtxVar("groups", strings.Join(claims.Groups, ","))
	r.SetCtxVar("email", email)
	r.SetCtxVar("displayName", displayName)
	r.SetCtxVar("avatar", avatar)

	if err := syncLocalUser(r, claims.Name, claims.Owner, email, displayName, avatar); err != nil {
		g.Log().Warningf(r.Context(), "failed to sync local user profile: %v", err)
	}

	r.Middleware.Next()
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

func syncLocalUser(r *ghttp.Request, userId, sourceOrgId, email, displayName, avatar string) error {
	if userId == "" {
		return nil
	}
	now := time.Now()
	_, err := g.DB().Exec(r.Context(), `
		INSERT INTO users (
			user_id, email, display_name, avatar, source_organization_id, created_at, updated_at, last_login_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (user_id) DO UPDATE SET
			email = EXCLUDED.email,
			display_name = EXCLUDED.display_name,
			avatar = EXCLUDED.avatar,
			source_organization_id = EXCLUDED.source_organization_id,
			updated_at = EXCLUDED.updated_at,
			last_login_at = EXCLUDED.last_login_at
	`, userId, email, displayName, avatar, sourceOrgId, now, now, now)
	return err
}

func GetUserId(r *ghttp.Request) string {
	return r.GetCtxVar("userId").String()
}

func GetOrganizationId(r *ghttp.Request) string {
	return r.GetCtxVar("organizationId").String()
}

func GetEmail(r *ghttp.Request) string {
	return r.GetCtxVar("email").String()
}

func GetDisplayName(r *ghttp.Request) string {
	return r.GetCtxVar("displayName").String()
}

func GetAvatar(r *ghttp.Request) string {
	return r.GetCtxVar("avatar").String()
}

func GetCurrentEnterpriseId(r *ghttp.Request) string {
	return r.GetCtxVar("enterpriseId").String()
}

func GetCurrentEnterpriseRole(r *ghttp.Request) string {
	return r.GetCtxVar("enterpriseRole").String()
}

func IsAdmin(r *ghttp.Request) bool {
	return r.GetCtxVar("isAdmin").Bool()
}

func GetGroups(r *ghttp.Request) []string {
	val := r.GetCtxVar("groups").String()
	if val == "" {
		return nil
	}
	return strings.Split(val, ",")
}
