package chatentry

import (
	"net/http"
	"strings"

	"dotblue/internal/domains/identity"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func GetAgentConfigHandler(r *ghttp.Request) {
	userID := identity.GetUserId(r)
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	agentID := strings.TrimSpace(r.Get("agentId").String())
	config, err := defaultService.GetOrCreateAgentConfig(r.Context(), enterpriseID, agentID)
	if err != nil {
		writeAdminError(r, err)
		return
	}
	if ok, err := agentAccessible(agentID, userID, enterpriseID); err != nil {
		writeAdminError(r, err)
		return
	} else if !ok {
		writeAdminError(r, ErrAgentNotAccessible)
		return
	}
	embedConfig, err := defaultService.GetEmbedConfig(r.Context(), enterpriseID, agentID, userID)
	if err != nil && err != ErrEmbedConfigNotFound {
		writeAdminError(r, err)
		return
	}
	shareLinks, err := defaultService.ListShareLinks(r.Context(), enterpriseID, agentID, userID)
	if err != nil {
		writeAdminError(r, err)
		return
	}
	r.Response.WriteJson(g.Map{
		"config":      config,
		"embedConfig": embedConfig,
		"shareLinks":  shareLinks,
	})
}

func UpsertAgentConfigHandler(r *ghttp.Request) {
	userID := identity.GetUserId(r)
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	agentID := strings.TrimSpace(r.Get("agentId").String())
	var req AgentEntryConfigInput
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	config, err := defaultService.UpsertAgentConfig(r.Context(), enterpriseID, agentID, userID, req)
	if err != nil {
		writeAdminError(r, err)
		return
	}
	r.Response.WriteJson(config)
}

func ListShareLinksHandler(r *ghttp.Request) {
	userID := identity.GetUserId(r)
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	agentID := strings.TrimSpace(r.Get("agentId").String())
	links, err := defaultService.ListShareLinks(r.Context(), enterpriseID, agentID, userID)
	if err != nil {
		writeAdminError(r, err)
		return
	}
	r.Response.WriteJson(links)
}

func CreateShareLinkHandler(r *ghttp.Request) {
	userID := identity.GetUserId(r)
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	var req CreateShareLinkInput
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	link, err := defaultService.CreateShareLink(r.Context(), enterpriseID, userID, req)
	if err != nil {
		writeAdminError(r, err)
		return
	}
	r.Response.WriteJson(g.Map{
		"shareLink": link,
		"shareUrl":  "/share/" + link.ShareCode,
	})
}

func RevokeShareLinkHandler(r *ghttp.Request) {
	userID := identity.GetUserId(r)
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	shareID := strings.TrimSpace(r.Get("id").String())
	if err := defaultService.RevokeShareLink(r.Context(), enterpriseID, userID, shareID); err != nil {
		writeAdminError(r, err)
		return
	}
	r.Response.WriteJson(g.Map{"ok": true})
}

func GetEmbedConfigHandler(r *ghttp.Request) {
	userID := identity.GetUserId(r)
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	agentID := strings.TrimSpace(r.Get("agentId").String())
	config, err := defaultService.GetEmbedConfig(r.Context(), enterpriseID, agentID, userID)
	if err != nil {
		writeAdminError(r, err)
		return
	}
	r.Response.WriteJson(config)
}

func UpsertEmbedConfigHandler(r *ghttp.Request) {
	userID := identity.GetUserId(r)
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	agentID := strings.TrimSpace(r.Get("agentId").String())
	var req EmbedConfigInput
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	config, err := defaultService.UpsertEmbedConfig(r.Context(), enterpriseID, agentID, userID, req)
	if err != nil {
		writeAdminError(r, err)
		return
	}
	r.Response.WriteJson(config)
}

func CreateEmbedTokenHandler(r *ghttp.Request) {
	userID := identity.GetUserId(r)
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	agentID := strings.TrimSpace(r.Get("agentId").String())
	origin := strings.TrimSpace(r.Get("origin").String())
	token, expiresInSeconds, err := defaultService.CreateEmbedToken(r.Context(), enterpriseID, agentID, userID, origin)
	if err != nil {
		writeAdminError(r, err)
		return
	}
	r.Response.WriteJson(g.Map{
		"embedToken":       token,
		"expiresInSeconds": expiresInSeconds,
	})
}

func writeAdminError(r *ghttp.Request, err error) {
	switch err {
	case ErrAgentNotAccessible:
		r.Response.WriteStatus(http.StatusNotFound, err.Error())
	case ErrEmbedOriginNotAllowed, ErrConversationNotAllowed:
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
	default:
		g.Log().Error(r.Context(), err)
		r.Response.WriteStatus(http.StatusInternalServerError, err.Error())
	}
}
