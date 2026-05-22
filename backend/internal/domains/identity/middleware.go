package identity

import (
	"context"
	"net/http"
	"strings"

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

	session, err := defaultService.ParseSession(tokenString)
	if err != nil {
		tokenPreview := normalizeBearerToken(tokenString)
		if len(tokenPreview) > 10 {
			tokenPreview = tokenPreview[:10]
		}
		g.Log().Errorf(r.Context(), "JWT validation failed for token [%s...]: %v", tokenPreview, err)
		msg := g.I18n().T(r.Context(), "auth_invalid")
		r.Response.WriteStatus(http.StatusUnauthorized, msg)
		r.ExitAll()
		return
	}

	r.SetCtxVar("userId", session.UserID)
	r.SetCtxVar("organizationId", session.OrganizationID)
	r.SetCtxVar("isAdmin", session.IsAdmin)
	r.SetCtxVar("groups", strings.Join(session.Groups, ","))
	r.SetCtxVar("email", session.Email)
	r.SetCtxVar("displayName", session.DisplayName)
	r.SetCtxVar("avatar", session.Avatar)

	if err := defaultService.SyncLocalUser(session); err != nil {
		g.Log().Warningf(r.Context(), "failed to sync local user profile: %v", err)
	}

	r.Middleware.Next()
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
