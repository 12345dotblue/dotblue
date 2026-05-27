package agent

import (
	"net/http"

	"dotblue/internal/domains/identity"
	"dotblue/internal/domains/model"
	"dotblue/internal/domains/settings"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type createReq struct {
	AgentName    string `json:"agentName" v:"required"`
	SystemPrompt string `json:"systemPrompt" v:"required"`
	ModelScope   string `json:"modelScope" v:"required"`
	ModelId      string `json:"modelId"`
	EngineType   string `json:"engineType"`
}

type updateReq struct {
	AgentName    string `json:"agentName" v:"required"`
	SystemPrompt string `json:"systemPrompt" v:"required"`
	ModelScope   string `json:"modelScope" v:"required"`
	ModelId      string `json:"modelId"`
	EngineType   string `json:"engineType"`
}

type modelOption struct {
	Label      string `json:"label"`
	Value      string `json:"value"`
	ModelScope string `json:"modelScope"`
	ModelId    string `json:"modelId,omitempty"`
	ModelName  string `json:"modelName,omitempty"`
}

type modelOptionGroup struct {
	Label   string        `json:"label"`
	Options []modelOption `json:"options"`
}

type runtimeOption struct {
	Value string `json:"value"`
}

type agentOptionsResp struct {
	ModelOptions   []modelOptionGroup `json:"modelOptions"`
	RuntimeOptions []runtimeOption    `json:"runtimeOptions"`
}

// ModelOptionsHandler returns grouped model options for agent creation.
// GET /api/agents/model-options
func ModelOptionsHandler(r *ghttp.Request) {
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	result := make([]modelOptionGroup, 0, 2)

	enterpriseModels, err := model.ListEnterpriseModels(enterpriseId)
	if err != nil {
		g.Log().Warningf(r.Context(), "Failed to list enterprise llm models: %v", err)
	} else if len(enterpriseModels) > 0 {
		options := make([]modelOption, 0, len(enterpriseModels))
		for i := range enterpriseModels {
			item := enterpriseModels[i]
			options = append(options, modelOption{
				Label:      item.DisplayName,
				Value:      ModelScopeEnterprise + ":" + item.Id,
				ModelScope: ModelScopeEnterprise,
				ModelId:    item.Id,
				ModelName:  item.Model,
			})
		}
		result = append(result, modelOptionGroup{
			Label:   "企业模型",
			Options: options,
		})
	}

	platformModels, err := model.ListPlatformModels()
	if err != nil {
		g.Log().Warningf(r.Context(), "Failed to list platform llm models: %v", err)
	} else if len(platformModels) > 0 {
		options := make([]modelOption, 0, len(platformModels))
		for i := range platformModels {
			item := platformModels[i]
			options = append(options, modelOption{
				Label:      item.DisplayName,
				Value:      ModelScopePlatform + ":" + item.Id,
				ModelScope: ModelScopePlatform,
				ModelId:    item.Id,
				ModelName:  item.Model,
			})
		}
		result = append(result, modelOptionGroup{
			Label:   "平台模型",
			Options: options,
		})
	}

	runtimeOptions := make([]runtimeOption, 0, 2)
	platformCfg, err := settings.GetPlatformConfig()
	if err != nil {
		g.Log().Warningf(r.Context(), "Failed to read platform runtime config: %v", err)
	} else {
		for _, item := range settings.EnabledRuntimeEngines(platformCfg) {
			runtimeOptions = append(runtimeOptions, runtimeOption{Value: item.EngineType})
		}
	}

	r.Response.WriteJson(agentOptionsResp{
		ModelOptions:   result,
		RuntimeOptions: runtimeOptions,
	})
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

	agent, err := defaultService.Create(userId, enterpriseId, req.AgentName, req.SystemPrompt, req.ModelScope, req.ModelId, req.EngineType)
	if err != nil {
		g.Log().Errorf(r.Context(), "Failed to create agent: %v", err)
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
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

	if err := defaultService.Update(agentId, req.AgentName, req.SystemPrompt, req.ModelScope, req.ModelId, req.EngineType); err != nil {
		g.Log().Errorf(r.Context(), "Failed to update agent: %v", err)
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
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
