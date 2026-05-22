package im

import (
	"net/http"

	"dotblue/internal/domains/identity"
	"github.com/gogf/gf/v2/net/ghttp"
)

func ListConnectionBindingsHandler(r *ghttp.Request) {
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	connectionID := r.Get("id").String()
	rows, err := defaultBindingService.ListBindingsByConnection(r.Context(), enterpriseID, connectionID)
	if err == ErrConnectionNotFound {
		r.Response.WriteStatus(http.StatusNotFound, "Connection not found")
		return
	}
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to list bindings")
		return
	}
	r.Response.WriteJson(rows)
}

func GetBindingHandler(r *ghttp.Request) {
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	id := r.Get("bindingId").String()
	binding, err := defaultBindingService.GetBinding(r.Context(), enterpriseID, id)
	if err == ErrBindingNotFound {
		r.Response.WriteStatus(http.StatusNotFound, "Binding not found")
		return
	}
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to load binding")
		return
	}
	r.Response.WriteJson(binding)
}

func CreateConnectionBindingHandler(r *ghttp.Request) {
	var req createBindingReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	connectionID := r.Get("id").String()
	binding, err := defaultBindingService.CreateBinding(r.Context(), enterpriseID, connectionID, req)
	if err == ErrConnectionNotFound {
		r.Response.WriteStatus(http.StatusNotFound, "Connection not found")
		return
	}
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(binding)
}

func UpdateBindingHandler(r *ghttp.Request) {
	var req updateBindingReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	id := r.Get("bindingId").String()
	binding, err := defaultBindingService.UpdateBinding(r.Context(), enterpriseID, id, req)
	if err == ErrBindingNotFound {
		r.Response.WriteStatus(http.StatusNotFound, "Binding not found")
		return
	}
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(binding)
}

func DeleteBindingHandler(r *ghttp.Request) {
	enterpriseID := identity.GetCurrentEnterpriseId(r)
	id := r.Get("bindingId").String()
	err := defaultBindingService.DeleteBinding(r.Context(), enterpriseID, id)
	if err == ErrBindingNotFound {
		r.Response.WriteStatus(http.StatusNotFound, "Binding not found")
		return
	}
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to delete binding")
		return
	}
	r.Response.WriteStatus(http.StatusNoContent)
}
