package agent

import (
	"net/http"

	"dotblue/internal/domains/identity"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type createReq struct {
	AgentName    string `json:"agentName" v:"required"`
	SystemPrompt string `json:"systemPrompt" v:"required"`
}

type updateReq struct {
	AgentName    string `json:"agentName" v:"required"`
	SystemPrompt string `json:"systemPrompt" v:"required"`
}

// ListHandler returns all agents for the current user.
// GET /api/agents
func ListHandler(r *ghttp.Request) {
	userId := identity.GetUserId(r)
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	if userId == "" {
		r.Response.WriteStatus(http.StatusUnauthorized, "User context not found")
		return
	}

	list, err := defaultService.ListByUserId(userId, enterpriseId)
	if err != nil {
		g.Log().Errorf(r.Context(), "Failed to list agents: %v", err)
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to list agents")
		return
	}

	result := make([]AgentPublic, 0, len(list))
	for _, agent := range list {
		result = append(result, toPublic(agent))
	}
	r.Response.WriteJson(result)
}

// CreateHandler creates a new agent for the current user.
// POST /api/agents
func CreateHandler(r *ghttp.Request) {
	var req createReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}

	userId := identity.GetUserId(r)
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	if userId == "" || enterpriseId == "" {
		r.Response.WriteStatus(http.StatusUnauthorized, "User context not found")
		return
	}

	agent, err := defaultService.Create(userId, enterpriseId, req.AgentName, req.SystemPrompt)
	if err != nil {
		g.Log().Errorf(r.Context(), "Failed to create agent: %v", err)
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to create agent")
		return
	}

	r.Response.WriteJson(toPublic(agent))
}

// GetHandler returns a single agent by ID (must belong to current user).
// GET /api/agents/{id}
func GetHandler(r *ghttp.Request) {
	userId := identity.GetUserId(r)
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	agentId := r.Get("id").String()
	if agentId == "" {
		r.Response.WriteStatus(http.StatusBadRequest, "Agent ID is required")
		return
	}

	ok, err := defaultService.BelongsToUser(agentId, userId, enterpriseId)
	if err != nil || !ok {
		r.Response.WriteStatus(http.StatusNotFound, "Agent not found")
		return
	}

	agent, err := defaultService.GetById(agentId)
	if err != nil || agent == nil {
		r.Response.WriteStatus(http.StatusNotFound, "Agent not found")
		return
	}

	r.Response.WriteJson(toPublic(agent))
}

// UpdateHandler modifies an existing agent.
// PUT /api/agents/{id}
func UpdateHandler(r *ghttp.Request) {
	userId := identity.GetUserId(r)
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	agentId := r.Get("id").String()
	if agentId == "" {
		r.Response.WriteStatus(http.StatusBadRequest, "Agent ID is required")
		return
	}

	ok, err := defaultService.BelongsToUser(agentId, userId, enterpriseId)
	if err != nil || !ok {
		r.Response.WriteStatus(http.StatusNotFound, "Agent not found")
		return
	}

	var req updateReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}

	if err := defaultService.Update(agentId, req.AgentName, req.SystemPrompt); err != nil {
		g.Log().Errorf(r.Context(), "Failed to update agent: %v", err)
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to update agent")
		return
	}

	r.Response.WriteJson(g.Map{"message": "Agent updated"})
}

// DeleteHandler removes an agent.
// DELETE /api/agents/{id}
func DeleteHandler(r *ghttp.Request) {
	userId := identity.GetUserId(r)
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	agentId := r.Get("id").String()
	if agentId == "" {
		r.Response.WriteStatus(http.StatusBadRequest, "Agent ID is required")
		return
	}

	ok, err := defaultService.BelongsToUser(agentId, userId, enterpriseId)
	if err != nil || !ok {
		r.Response.WriteStatus(http.StatusNotFound, "Agent not found")
		return
	}

	if err := defaultService.Delete(agentId); err != nil {
		g.Log().Errorf(r.Context(), "Failed to delete agent: %v", err)
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to delete agent")
		return
	}

	r.Response.WriteJson(g.Map{"message": "Agent deleted"})
}
