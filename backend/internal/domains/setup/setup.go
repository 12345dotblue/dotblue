package setup

import (
	"errors"
	"net/http"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"dotblue/internal/domains/settings"
)

// installMu prevents concurrent install requests (race condition).
var installMu sync.Mutex

// StatusHandler returns whether the platform has been initialized.
// GET /api/setup/status — public, no auth required.
func StatusHandler(r *ghttp.Request) {
	initialized := settings.IsInitialized()
	r.Response.WriteJson(g.Map{"initialized": initialized})
}

// InstallReq is the request body for the install endpoint.
type InstallReq struct {
	AdminUsername string `json:"adminUsername" v:"required"`
	AdminPassword string `json:"adminPassword" v:"required"`
	AdminEmail    string `json:"adminEmail"    v:"required"`
}

// InstallHandler performs the first-time platform setup.
// POST /api/setup/install — public, but only works when system is not initialized.
func InstallHandler(r *ghttp.Request) {
	installMu.Lock()
	defer installMu.Unlock()

	if settings.IsInitialized() {
		r.Response.WriteStatus(http.StatusForbidden, "Platform already initialized")
		return
	}

	var req InstallReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}

	if err := runInstall(r, &req); err != nil {
		g.Log().Errorf(r.Context(), "Installation failed: %v", err)
		if errors.Is(err, ErrUserExists) {
			r.Response.WriteStatus(http.StatusConflict, "USER_ALREADY_EXISTS")
			return
		}
		r.Response.WriteStatus(http.StatusInternalServerError, "Installation failed")
		return
	}

	r.Response.WriteJson(g.Map{"message": "Platform initialized successfully"})
}
