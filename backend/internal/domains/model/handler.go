package model

import (
	"net/http"
	"strings"

	"dotblue/internal/domains/identity"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type enterpriseModelReq struct {
	DisplayName string `json:"displayName" v:"required"`
	Type        string `json:"type" v:"required"`
	ApiBase     string `json:"apiBase"`
	ApiKey      string `json:"apiKey"`
	Model       string `json:"model" v:"required"`
	FundingType string `json:"fundingType"`
}

type platformModelReq struct {
	DisplayName string `json:"displayName" v:"required"`
	Type        string `json:"type" v:"required"`
	ApiBase     string `json:"apiBase"`
	ApiKey      string `json:"apiKey"`
	Model       string `json:"model" v:"required"`
	FundingType string `json:"fundingType"`
	IsDefault   bool   `json:"isDefault"`
}

func ListEnterpriseModelsHandler(r *ghttp.Request) {
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	list, err := defaultService.List(ScopeEnterprise, enterpriseId)
	if err != nil {
		g.Log().Errorf(r.Context(), "Failed to list enterprise models: %v", err)
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to list enterprise models")
		return
	}
	result := make([]LLMModel, 0, len(list))
	for i := range list {
		item := list[i]
		result = append(result, sanitize(&item))
	}
	r.Response.WriteJson(result)
}

func CreateEnterpriseModelHandler(r *ghttp.Request) {
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	var req enterpriseModelReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	item, err := defaultService.Create(ScopeEnterprise, enterpriseId, CreateReq{
		DisplayName: req.DisplayName,
		Type:        req.Type,
		ApiBase:     req.ApiBase,
		ApiKey:      req.ApiKey,
		Model:       req.Model,
		FundingType: req.FundingType,
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(sanitize(item))
}

func UpdateEnterpriseModelHandler(r *ghttp.Request) {
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	id := strings.TrimSpace(r.Get("id").String())
	if id == "" {
		r.Response.WriteStatus(http.StatusBadRequest, "Model ID is required")
		return
	}
	var req enterpriseModelReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	item, err := defaultService.Update(id, ScopeEnterprise, enterpriseId, UpdateReq{
		DisplayName: req.DisplayName,
		Type:        req.Type,
		ApiBase:     req.ApiBase,
		ApiKey:      req.ApiKey,
		Model:       req.Model,
		FundingType: req.FundingType,
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(sanitize(item))
}

func DeleteEnterpriseModelHandler(r *ghttp.Request) {
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	id := strings.TrimSpace(r.Get("id").String())
	if id == "" {
		r.Response.WriteStatus(http.StatusBadRequest, "Model ID is required")
		return
	}
	if err := defaultService.Delete(id, ScopeEnterprise, enterpriseId); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(g.Map{"message": "Enterprise llm model deleted"})
}

func ListPlatformModelsHandler(r *ghttp.Request) {
	list, err := defaultService.List(ScopePlatform, "")
	if err != nil {
		g.Log().Errorf(r.Context(), "Failed to list platform models: %v", err)
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to list platform models")
		return
	}
	result := make([]LLMModel, 0, len(list))
	for i := range list {
		item := list[i]
		result = append(result, sanitize(&item))
	}
	r.Response.WriteJson(result)
}

func CreatePlatformModelHandler(r *ghttp.Request) {
	var req platformModelReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	item, err := defaultService.Create(ScopePlatform, "", CreateReq{
		DisplayName: req.DisplayName,
		Type:        req.Type,
		ApiBase:     req.ApiBase,
		ApiKey:      req.ApiKey,
		Model:       req.Model,
		FundingType: req.FundingType,
		IsDefault:   req.IsDefault,
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(sanitize(item))
}

func UpdatePlatformModelHandler(r *ghttp.Request) {
	id := strings.TrimSpace(r.Get("id").String())
	if id == "" {
		r.Response.WriteStatus(http.StatusBadRequest, "Model ID is required")
		return
	}
	var req platformModelReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	item, err := defaultService.Update(id, ScopePlatform, "", UpdateReq{
		DisplayName: req.DisplayName,
		Type:        req.Type,
		ApiBase:     req.ApiBase,
		ApiKey:      req.ApiKey,
		Model:       req.Model,
		FundingType: req.FundingType,
		IsDefault:   req.IsDefault,
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(sanitize(item))
}

func DeletePlatformModelHandler(r *ghttp.Request) {
	id := strings.TrimSpace(r.Get("id").String())
	if id == "" {
		r.Response.WriteStatus(http.StatusBadRequest, "Model ID is required")
		return
	}
	if err := defaultService.Delete(id, ScopePlatform, ""); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(g.Map{"message": "Platform llm model deleted"})
}
