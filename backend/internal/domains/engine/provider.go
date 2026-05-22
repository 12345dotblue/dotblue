package engine

import (
	"encoding/json"
	"net/http"

	"dotblue/internal/domains/settings"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// settingsUpdateReq is the request body for the unified settings endpoint.
// Both platform and provider are optional — only sent fields are updated.
type settingsUpdateReq struct {
	Platform *settings.PlatformConfig `json:"platform"`
	Provider *settings.ProviderConfig `json:"provider"`
}

// SettingsHandler saves platform and/or provider configuration (admin only).
// POST /api/admin/settings
func SettingsHandler(r *ghttp.Request) {
	var req settingsUpdateReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}

	if req.Platform == nil && req.Provider == nil {
		r.Response.WriteStatus(http.StatusBadRequest, "At least one of platform or provider is required")
		return
	}

	if req.Platform != nil {
		if err := settings.UpdatePlatformConfig(req.Platform); err != nil {
			g.Log().Errorf(r.Context(), "Failed to save platform config: %v", err)
			r.Response.WriteStatus(http.StatusInternalServerError, "Failed to save platform config")
			return
		}
		g.Log().Info(r.Context(), "Saved platform config")
	}

	if req.Provider != nil {
		if err := settings.UpdateProviderConfig(req.Provider); err != nil {
			g.Log().Errorf(r.Context(), "Failed to save provider config: %v", err)
			r.Response.WriteStatus(http.StatusInternalServerError, "Failed to save provider config")
			return
		}
		g.Log().Info(r.Context(), "Saved provider config")
	}

	r.Response.WriteJson(g.Map{"message": "Settings saved"})
}

// GetSettingsHandler returns the full settings with masked provider API key.
// GET /api/admin/settings
func GetSettingsHandler(r *ghttp.Request) {
	s, err := settings.GetSettings()
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to read settings")
		return
	}

	// Build platform response
	var platformResp interface{}
	if len(s.Platform) > 0 && string(s.Platform) != "{}" {
		var pc settings.PlatformConfig
		if err := json.Unmarshal(s.Platform, &pc); err == nil {
			platformResp = pc
		} else {
			platformResp = nil
		}
	}

	// Build provider response (mask API key)
	var providerResp interface{}
	if len(s.Provider) > 0 && string(s.Provider) != "{}" {
		var pc settings.ProviderConfig
		if err := json.Unmarshal(s.Provider, &pc); err == nil {
			pc.ApiKey = settings.MaskAPIKey(pc.ApiKey)
			providerResp = pc
		} else {
			providerResp = nil
		}
	}

	r.Response.WriteJson(g.Map{
		"initialized": s.Initialized,
		"platform":    platformResp,
		"provider":    providerResp,
	})
}
