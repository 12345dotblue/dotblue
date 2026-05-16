package identity

import (
	"net/http"

	"github.com/gogf/gf/v2/net/ghttp"
)

const AdminGroup = "admin"

// AdminMiddleware checks if the authenticated user belongs to the admin group.
// Must be placed after Middleware (which sets ctxVar "isAdmin" and "groups").
func AdminMiddleware(r *ghttp.Request) {
	isAdmin := r.GetCtxVar("isAdmin").Bool()
	if isAdmin {
		r.Middleware.Next()
		return
	}

	groups := r.GetCtxVar("groups").Strings()
	for _, g := range groups {
		if g == AdminGroup {
			r.Middleware.Next()
			return
		}
	}

	r.Response.WriteStatus(http.StatusForbidden, "Admin access required")
	r.ExitAll()
}
