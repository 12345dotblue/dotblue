package agent

import (
	"fmt"
	"strings"
	"time"

	"dotblue/internal/domains/model"
	"github.com/google/uuid"
)

const (
	ModelScopePlatform   = "platform"
	ModelScopeEnterprise = "enterprise"
	EngineTypeHermes     = "hermes"
	EngineTypeNanobot    = "nanobot"
)

// Agent represents a user's agent record (stored in agents table).
type Agent struct {
	Id           string    `json:"id"`
	UserId       string    `json:"userId"`
	GroupId      string    `json:"groupId"`
	AgentName    string    `json:"agentName"`
	SystemPrompt string    `json:"systemPrompt"`
	ModelScope   string    `json:"modelScope,omitempty" orm:"model_scope"`
	ModelId      string    `json:"modelId,omitempty" orm:"model_id"`
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
	ModelScope   string    `json:"modelScope"`
	ModelId      string    `json:"modelId,omitempty"`
	ModelName    string    `json:"modelName,omitempty"`
	EngineType   string    `json:"engineType"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func toPublic(a *Agent) AgentPublic {
	modelScope, modelId := normalizeModelSelection(a.ModelScope, a.ModelId)
	engineType := normalizeEngineType(a.EngineType)
	return AgentPublic{
		Id:           a.Id,
		AgentName:    a.AgentName,
		SystemPrompt: a.SystemPrompt,
		ModelScope:   modelScope,
		ModelId:      modelId,
		ModelName:    resolveModelName(a.GroupId, modelScope, modelId),
		EngineType:   engineType,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}
}

func normalizeModelSelection(modelScope, modelId string) (string, string) {
	switch strings.TrimSpace(modelScope) {
	case ModelScopeEnterprise:
		return ModelScopeEnterprise, strings.TrimSpace(modelId)
	default:
		return ModelScopePlatform, strings.TrimSpace(modelId)
	}
}

func resolveModelName(enterpriseId, modelScope, modelId string) string {
	if modelId != "" {
		item, err := model.GetByID(modelId)
		if err == nil && item != nil {
			return strings.TrimSpace(item.DisplayName)
		}
	}
	if modelScope == ModelScopePlatform {
		item, err := model.GetDefaultPlatformModel()
		if err == nil && item != nil {
			return strings.TrimSpace(item.DisplayName)
		}
	}
	return ""
}

func generateEngineAPIKey() string {
	return fmt.Sprintf("dotblue-%s", uuid.New().String())
}

var defaultService = NewService(NewGFRepository())

// GetById retrieves an agent by its UUID. Returns nil if not found.
func GetById(id string) (*Agent, error) {
	return defaultService.GetById(id)
}

// ListByUserId returns all agents belonging to a user.
func ListByUserId(userId, enterpriseId string) ([]*Agent, error) {
	return defaultService.ListByUserId(userId, enterpriseId)
}

// BelongsToUser checks whether an agent belongs to the given user.
func BelongsToUser(id, userId, enterpriseId string) (bool, error) {
	return defaultService.BelongsToUser(id, userId, enterpriseId)
}

// Create inserts a new agent record.
func Create(userId, groupId, agentName, systemPrompt, modelScope, modelId, engineType string) (*Agent, error) {
	return defaultService.Create(userId, groupId, agentName, systemPrompt, modelScope, modelId, engineType)
}

// Update modifies an agent's name and system prompt by ID.
func Update(id, agentName, systemPrompt, modelScope, modelId, engineType string) error {
	return defaultService.Update(id, agentName, systemPrompt, modelScope, modelId, engineType)
}

// Delete removes an agent by ID.
func Delete(id string) error {
	return defaultService.Delete(id)
}
