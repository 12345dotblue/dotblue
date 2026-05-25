package agent

import "time"

// Repository defines agent persistence operations.
type Repository interface {
	GetById(id string) (*Agent, error)
	ListByUserId(userId, enterpriseId string) ([]*Agent, error)
	BelongsToUser(id, userId, enterpriseId string) (bool, error)
	Create(agent *Agent) error
	Update(id, agentName, systemPrompt, modelScope, modelId string, updatedAt time.Time) error
	Delete(id string) error
}
