package agent

import (
	"fmt"
	"time"

	"github.com/google/uuid"
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
func Create(userId, groupId, agentName, systemPrompt string) (*Agent, error) {
	return defaultService.Create(userId, groupId, agentName, systemPrompt)
}

// Update modifies an agent's name and system prompt by ID.
func Update(id, agentName, systemPrompt string) error {
	return defaultService.Update(id, agentName, systemPrompt)
}

// Delete removes an agent by ID.
func Delete(id string) error {
	return defaultService.Delete(id)
}
