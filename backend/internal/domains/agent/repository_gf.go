package agent

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type GFRepository struct{}

func NewGFRepository() *GFRepository {
	return &GFRepository{}
}

func (r *GFRepository) GetById(id string) (*Agent, error) {
	var agent Agent
	err := g.DB().Model("agents").Where("id = ?", id).Scan(&agent)
	if err != nil {
		return nil, err
	}
	if agent.Id == "" {
		return nil, nil
	}
	return &agent, nil
}

func (r *GFRepository) ListByUserId(userId, enterpriseId string) ([]*Agent, error) {
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

func (r *GFRepository) BelongsToUser(id, userId, enterpriseId string) (bool, error) {
	count, err := g.DB().Model("agents").
		Where("id = ? AND user_id = ? AND group_id = ?", id, userId, enterpriseId).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GFRepository) Create(agent *Agent) error {
	_, err := g.DB().Model("agents").Data(g.Map{
		"id":             agent.Id,
		"user_id":        agent.UserId,
		"group_id":       agent.GroupId,
		"agent_name":     agent.AgentName,
		"system_prompt":  agent.SystemPrompt,
		"model_scope":    agent.ModelScope,
		"model_id":       agent.ModelId,
		"hermes_api_key": agent.EngineAPIKey,
		"engine_type":    agent.EngineType,
	}).Insert()
	return err
}

func (r *GFRepository) Update(id, agentName, systemPrompt, modelScope, modelId string, updatedAt time.Time) error {
	_, err := g.DB().Model("agents").
		Data(g.Map{
			"agent_name":    agentName,
			"system_prompt": systemPrompt,
			"model_scope":   modelScope,
			"model_id":      modelId,
			"updated_at":    updatedAt,
		}).
		Where("id = ?", id).
		Update()
	return err
}

func (r *GFRepository) Delete(id string) error {
	_, err := g.DB().Model("agents").Where("id = ?", id).Delete()
	return err
}
