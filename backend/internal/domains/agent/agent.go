package agent

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"dotblue/internal/domains/identity"
)

// Agent represents a user's agent record (stored in agents table).
type Agent struct {
	Id           string    `json:"id"`
	UserId       string    `json:"userId"`
	GroupId      string    `json:"groupId"`
	AgentName    string    `json:"agentName"`
	SystemPrompt string    `json:"systemPrompt"`
	EngineAPIKey string    `json:"engineApiKey,omitempty" orm:"hermes_api_key"` // DB column: hermes_api_key (legacy name)
	EngineType   string    `json:"engineType,omitempty" orm:"engine_type"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// AgentPublic is the safe JSON representation returned via API (no secrets).
type AgentPublic struct {
	Id           string    `json:"id"`
	AgentName    string    `json:"agentName"`
	SystemPrompt string    `json:"systemPrompt"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func toPublic(a *Agent) AgentPublic {
	return AgentPublic{
		Id:           a.Id,
		AgentName:    a.AgentName,
		SystemPrompt: a.SystemPrompt,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}
}

func generateEngineAPIKey() string {
	return fmt.Sprintf("dotblue-%s", uuid.New().String())
}

// GetById retrieves an agent by its UUID. Returns nil if not found.
func GetById(id string) (*Agent, error) {
	var a Agent
	err := g.DB().Model("agents").Where("id = ?", id).Scan(&a)
	if err != nil {
		return nil, err
	}
	if a.Id == "" {
		return nil, nil
	}
	return &a, nil
}

// ListByUserId returns all agents belonging to a user.
func ListByUserId(userId, enterpriseId string) ([]*Agent, error) {
	var list []*Agent
	err := g.DB().Model("agents").
		Where("user_id = ? AND group_id = ?", userId, enterpriseId).
		Order("created_at ASC").
		Scan(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// BelongsToUser checks whether an agent belongs to the given user.
func BelongsToUser(id, userId, enterpriseId string) (bool, error) {
	count, err := g.DB().Model("agents").Where("id = ? AND user_id = ? AND group_id = ?", id, userId, enterpriseId).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Create inserts a new agent record.
func Create(userId, groupId, agentName, systemPrompt string) (*Agent, error) {
	id := uuid.New().String()
	apiKey := generateEngineAPIKey()
	_, err := g.DB().Model("agents").Data(g.Map{
		"id":             id,
		"user_id":        userId,
		"group_id":       groupId,
		"agent_name":     agentName,
		"system_prompt":  systemPrompt,
		"hermes_api_key": apiKey,
		"engine_type":    "hermes",
	}).Insert()
	if err != nil {
		return nil, err
	}

	// Read back to get created_at timestamps
	a, err := GetById(id)
	if err != nil {
		return nil, err
	}
	return a, nil
}

// Update modifies an agent's name and system prompt by ID.
func Update(id, agentName, systemPrompt string) error {
	_, err := g.DB().Model("agents").
		Data(g.Map{
			"agent_name":    agentName,
			"system_prompt": systemPrompt,
			"updated_at":    time.Now(),
		}).
		Where("id = ?", id).
		Update()
	return err
}

// Delete removes an agent by ID.
func Delete(id string) error {
	_, err := g.DB().Model("agents").Where("id = ?", id).Delete()
	return err
}

// --- HTTP Handlers ---

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

	list, err := ListByUserId(userId, enterpriseId)
	if err != nil {
		g.Log().Errorf(r.Context(), "Failed to list agents: %v", err)
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to list agents")
		return
	}

	result := make([]AgentPublic, 0, len(list))
	for _, a := range list {
		result = append(result, toPublic(a))
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

	a, err := Create(userId, enterpriseId, req.AgentName, req.SystemPrompt)
	if err != nil {
		g.Log().Errorf(r.Context(), "Failed to create agent: %v", err)
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to create agent")
		return
	}

	r.Response.WriteJson(toPublic(a))
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

	ok, err := BelongsToUser(agentId, userId, enterpriseId)
	if err != nil || !ok {
		r.Response.WriteStatus(http.StatusNotFound, "Agent not found")
		return
	}

	a, err := GetById(agentId)
	if err != nil || a == nil {
		r.Response.WriteStatus(http.StatusNotFound, "Agent not found")
		return
	}

	r.Response.WriteJson(toPublic(a))
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

	ok, err := BelongsToUser(agentId, userId, enterpriseId)
	if err != nil || !ok {
		r.Response.WriteStatus(http.StatusNotFound, "Agent not found")
		return
	}

	var req updateReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}

	if err := Update(agentId, req.AgentName, req.SystemPrompt); err != nil {
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

	ok, err := BelongsToUser(agentId, userId, enterpriseId)
	if err != nil || !ok {
		r.Response.WriteStatus(http.StatusNotFound, "Agent not found")
		return
	}

	if err := Delete(agentId); err != nil {
		g.Log().Errorf(r.Context(), "Failed to delete agent: %v", err)
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to delete agent")
		return
	}

	r.Response.WriteJson(g.Map{"message": "Agent deleted"})
}
