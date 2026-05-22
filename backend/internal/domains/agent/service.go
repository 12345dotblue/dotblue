package agent

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const defaultEngineType = "hermes"

// Service encapsulates agent business logic and depends on persistence abstractions.
type Service struct {
	repo            Repository
	idGenerator     func() string
	apiKeyGenerator func() string
	now             func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo:            repo,
		idGenerator:     func() string { return uuid.New().String() },
		apiKeyGenerator: generateEngineAPIKey,
		now:             time.Now,
	}
}

func (s *Service) GetById(id string) (*Agent, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("agent repository is not configured")
	}
	return s.repo.GetById(id)
}

func (s *Service) ListByUserId(userId, enterpriseId string) ([]*Agent, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("agent repository is not configured")
	}
	return s.repo.ListByUserId(userId, enterpriseId)
}

func (s *Service) BelongsToUser(id, userId, enterpriseId string) (bool, error) {
	if s == nil || s.repo == nil {
		return false, errors.New("agent repository is not configured")
	}
	return s.repo.BelongsToUser(id, userId, enterpriseId)
}

func (s *Service) Create(userId, groupId, agentName, systemPrompt string) (*Agent, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("agent repository is not configured")
	}

	agent := &Agent{
		Id:           s.idGenerator(),
		UserId:       userId,
		GroupId:      groupId,
		AgentName:    agentName,
		SystemPrompt: systemPrompt,
		EngineAPIKey: s.apiKeyGenerator(),
		EngineType:   defaultEngineType,
	}
	if err := s.repo.Create(agent); err != nil {
		return nil, err
	}

	created, err := s.repo.GetById(agent.Id)
	if err != nil {
		return nil, err
	}
	if created != nil {
		return created, nil
	}
	return agent, nil
}

func (s *Service) Update(id, agentName, systemPrompt string) error {
	if s == nil || s.repo == nil {
		return errors.New("agent repository is not configured")
	}
	return s.repo.Update(id, agentName, systemPrompt, s.now())
}

func (s *Service) Delete(id string) error {
	if s == nil || s.repo == nil {
		return errors.New("agent repository is not configured")
	}
	return s.repo.Delete(id)
}
