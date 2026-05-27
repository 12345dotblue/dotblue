package agent

import (
	"errors"
	"strings"
	"time"

	"dotblue/internal/domains/model"
	"dotblue/internal/domains/settings"
	"github.com/google/uuid"
)

var (
	loadModelByID            = model.GetByID
	loadDefaultPlatformModel = model.GetDefaultPlatformModel
	loadPlatformConfig       = settings.GetPlatformConfig
)

const defaultEngineType = EngineTypeHermes

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

func (s *Service) Create(userId, groupId, agentName, systemPrompt, modelScope, modelId, engineType string) (*Agent, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("agent repository is not configured")
	}
	modelScope, modelId, err := validateModelSelection(modelScope, modelId)
	if err != nil {
		return nil, err
	}
	engineType, err = validateEngineType(engineType)
	if err != nil {
		return nil, err
	}
	if err := ensureModelSelectionExists(groupId, modelScope, modelId); err != nil {
		return nil, err
	}

	agent := &Agent{
		Id:           s.idGenerator(),
		UserId:       userId,
		GroupId:      groupId,
		AgentName:    agentName,
		SystemPrompt: systemPrompt,
		ModelScope:   modelScope,
		ModelId:      modelId,
		EngineAPIKey: s.apiKeyGenerator(),
		EngineType:   engineType,
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

func (s *Service) Update(id, agentName, systemPrompt, modelScope, modelId, engineType string) error {
	if s == nil || s.repo == nil {
		return errors.New("agent repository is not configured")
	}
	modelScope, modelId, err := validateModelSelection(modelScope, modelId)
	if err != nil {
		return err
	}
	engineType, err = validateEngineType(engineType)
	if err != nil {
		return err
	}
	existing, err := s.repo.GetById(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("agent not found")
	}
	if err := ensureModelSelectionExists(existing.GroupId, modelScope, modelId); err != nil {
		return err
	}
	return s.repo.Update(id, agentName, systemPrompt, modelScope, modelId, engineType, s.now())
}

func (s *Service) Delete(id string) error {
	if s == nil || s.repo == nil {
		return errors.New("agent repository is not configured")
	}
	return s.repo.Delete(id)
}

func validateModelSelection(modelScope, modelId string) (string, string, error) {
	switch strings.TrimSpace(modelScope) {
	case ModelScopePlatform:
		modelId = strings.TrimSpace(modelId)
		if modelId == "" {
			return "", "", errors.New("platform model id is required")
		}
		return ModelScopePlatform, modelId, nil
	case ModelScopeEnterprise:
		modelId = strings.TrimSpace(modelId)
		if modelId == "" {
			return "", "", errors.New("enterprise model id is required")
		}
		return ModelScopeEnterprise, modelId, nil
	default:
		return "", "", errors.New("model scope is required")
	}
}

func validateEngineType(engineType string) (string, error) {
	cfg, err := loadPlatformConfig()
	if err != nil {
		return "", err
	}

	enabled := settings.EnabledRuntimeEngines(cfg)
	if len(enabled) == 0 {
		enabled = []settings.RuntimeEngineConfig{{EngineType: EngineTypeHermes, Enabled: true}}
	}

	normalized := strings.TrimSpace(strings.ToLower(engineType))
	if normalized == "" {
		if len(enabled) == 1 {
			return enabled[0].EngineType, nil
		}
		return "", errors.New("engine type is required")
	}

	switch normalized {
	case EngineTypeHermes, EngineTypeNanobot:
		for _, item := range enabled {
			if item.EngineType == normalized {
				return normalized, nil
			}
		}
		return "", errors.New("engine type is not enabled")
	default:
		return "", errors.New("engine type is invalid")
	}
}

func normalizeEngineType(engineType string) string {
	switch strings.TrimSpace(strings.ToLower(engineType)) {
	case "", EngineTypeHermes:
		return EngineTypeHermes
	case EngineTypeNanobot:
		return EngineTypeNanobot
	default:
		return strings.TrimSpace(strings.ToLower(engineType))
	}
}

func ensureModelSelectionExists(enterpriseId, modelScope, modelId string) error {
	item, err := loadModelByID(modelId)
	if err != nil {
		return err
	}
	if item == nil {
		if modelScope == ModelScopePlatform {
			return errors.New("platform model not found")
		}
		return errors.New("enterprise model not found")
	}
	if modelScope == ModelScopeEnterprise {
		if strings.TrimSpace(item.Scope) != model.ScopeEnterprise || strings.TrimSpace(item.EnterpriseId) != strings.TrimSpace(enterpriseId) {
			return errors.New("enterprise model not found")
		}
		return nil
	}
	if strings.TrimSpace(item.Scope) != model.ScopePlatform {
		return errors.New("platform model not found")
	}
	return nil
}
