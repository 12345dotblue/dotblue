package engine

import (
	"encoding/json"
	"net/http"

	"dotblue/internal/domains/settings"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// settingsUpdateReq is the request body for the unified settings endpoint.
// Only platform settings are handled here. Model settings are managed by model domain APIs.
type settingsUpdateReq struct {
	Platform *settings.PlatformConfig `json:"platform"`
}

// SettingsHandler saves platform configuration (admin only).
// POST /api/admin/settings
func SettingsHandler(r *ghttp.Request) {
	var req settingsUpdateReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}

	if req.Platform == nil {
		r.Response.WriteStatus(http.StatusBadRequest, "platform is required")
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

	r.Response.WriteJson(g.Map{"message": "Settings saved"})
}

// GetSettingsHandler returns platform settings only.
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

	r.Response.WriteJson(g.Map{
		"initialized": s.Initialized,
		"platform":    platformResp,
	})
}
