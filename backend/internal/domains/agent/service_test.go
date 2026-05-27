package agent

import (
	"errors"
	"testing"
	"time"

	"dotblue/internal/domains/model"
	"dotblue/internal/domains/settings"
)

type stubRepository struct {
	getByIdFunc       func(id string) (*Agent, error)
	listByUserIdFunc  func(userId, enterpriseId string) ([]*Agent, error)
	belongsToUserFunc func(id, userId, enterpriseId string) (bool, error)
	createFunc        func(agent *Agent) error
	updateFunc        func(id, agentName, systemPrompt, modelScope, modelId, engineType string, updatedAt time.Time) error
	deleteFunc        func(id string) error
}

func (s *stubRepository) GetById(id string) (*Agent, error) {
	if s.getByIdFunc != nil {
		return s.getByIdFunc(id)
	}
	return nil, nil
}

func (s *stubRepository) ListByUserId(userId, enterpriseId string) ([]*Agent, error) {
	if s.listByUserIdFunc != nil {
		return s.listByUserIdFunc(userId, enterpriseId)
	}
	return nil, nil
}

func (s *stubRepository) BelongsToUser(id, userId, enterpriseId string) (bool, error) {
	if s.belongsToUserFunc != nil {
		return s.belongsToUserFunc(id, userId, enterpriseId)
	}
	return false, nil
}

func (s *stubRepository) Create(agent *Agent) error {
	if s.createFunc != nil {
		return s.createFunc(agent)
	}
	return nil
}

func (s *stubRepository) Update(id, agentName, systemPrompt, modelScope, modelId, engineType string, updatedAt time.Time) error {
	if s.updateFunc != nil {
		return s.updateFunc(id, agentName, systemPrompt, modelScope, modelId, engineType, updatedAt)
	}
	return nil
}

func (s *stubRepository) Delete(id string) error {
	if s.deleteFunc != nil {
		return s.deleteFunc(id)
	}
	return nil
}

func allowRuntimeEnginesForTest(t *testing.T, items ...settings.RuntimeEngineConfig) {
	t.Helper()
	originalPlatformLoader := loadPlatformConfig
	loadPlatformConfig = func() (*settings.PlatformConfig, error) {
		return &settings.PlatformConfig{
			DataBasePath:   "/runtime-data",
			RuntimeEngines: items,
		}, nil
	}
	t.Cleanup(func() {
		loadPlatformConfig = originalPlatformLoader
	})
}

func TestServiceCreateGeneratesDefaults(t *testing.T) {
	allowRuntimeEnginesForTest(t, settings.RuntimeEngineConfig{EngineType: EngineTypeHermes, Enabled: true})
	originalModelLoader := loadModelByID
	loadModelByID = func(id string) (*model.LLMModel, error) {
		return &model.LLMModel{Id: id, Scope: model.ScopePlatform}, nil
	}
	defer func() { loadModelByID = originalModelLoader }()

	var created *Agent
	repo := &stubRepository{
		createFunc: func(agent *Agent) error {
			created = agent
			return nil
		},
		getByIdFunc: func(id string) (*Agent, error) {
			return &Agent{
				Id:           id,
				UserId:       "user-1",
				GroupId:      "group-1",
				AgentName:    "support",
				SystemPrompt: "helpful",
				ModelScope:   ModelScopePlatform,
				EngineAPIKey: "generated-key",
				EngineType:   defaultEngineType,
			}, nil
		},
	}

	service := NewService(repo)
	service.idGenerator = func() string { return "agent-123" }
	service.apiKeyGenerator = func() string { return "generated-key" }

	got, err := service.Create("user-1", "group-1", "support", "helpful", ModelScopePlatform, "platform-default", "")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created == nil {
		t.Fatal("expected repository Create to receive an agent")
	}
	if created.Id != "agent-123" {
		t.Fatalf("expected generated id, got %q", created.Id)
	}
	if created.EngineAPIKey != "generated-key" {
		t.Fatalf("expected generated api key, got %q", created.EngineAPIKey)
	}
	if created.EngineType != defaultEngineType {
		t.Fatalf("expected engine type %q, got %q", defaultEngineType, created.EngineType)
	}
	if got == nil || got.Id != "agent-123" {
		t.Fatalf("expected created agent to be read back, got %#v", got)
	}
}

func TestServiceUpdateUsesClock(t *testing.T) {
	allowRuntimeEnginesForTest(
		t,
		settings.RuntimeEngineConfig{EngineType: EngineTypeHermes, Enabled: true},
		settings.RuntimeEngineConfig{EngineType: EngineTypeNanobot, Enabled: true},
	)
	originalModelLoader := loadModelByID
	loadModelByID = func(id string) (*model.LLMModel, error) {
		return &model.LLMModel{Id: id, Scope: model.ScopeEnterprise, EnterpriseId: "group-1"}, nil
	}
	defer func() { loadModelByID = originalModelLoader }()

	expectedTime := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	repo := &stubRepository{
		getByIdFunc: func(id string) (*Agent, error) {
			return &Agent{Id: id, GroupId: "group-1"}, nil
		},
		updateFunc: func(id, agentName, systemPrompt, modelScope, modelId, engineType string, updatedAt time.Time) error {
			if id != "agent-1" {
				t.Fatalf("unexpected id %q", id)
			}
			if agentName != "renamed" {
				t.Fatalf("unexpected name %q", agentName)
			}
			if systemPrompt != "new prompt" {
				t.Fatalf("unexpected prompt %q", systemPrompt)
			}
			if modelScope != ModelScopeEnterprise {
				t.Fatalf("unexpected modelScope %q", modelScope)
			}
			if modelId != "model-1" {
				t.Fatalf("unexpected modelId %q", modelId)
			}
			if engineType != EngineTypeNanobot {
				t.Fatalf("unexpected engineType %q", engineType)
			}
			if !updatedAt.Equal(expectedTime) {
				t.Fatalf("unexpected updatedAt %v", updatedAt)
			}
			return nil
		},
	}

	service := NewService(repo)
	service.now = func() time.Time { return expectedTime }

	if err := service.Update("agent-1", "renamed", "new prompt", ModelScopeEnterprise, "model-1", EngineTypeNanobot); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
}

func TestServiceCreatePropagatesRepositoryError(t *testing.T) {
	allowRuntimeEnginesForTest(t, settings.RuntimeEngineConfig{EngineType: EngineTypeHermes, Enabled: true})
	originalModelLoader := loadModelByID
	loadModelByID = func(id string) (*model.LLMModel, error) {
		return &model.LLMModel{Id: id, Scope: model.ScopePlatform}, nil
	}
	defer func() { loadModelByID = originalModelLoader }()

	repo := &stubRepository{
		createFunc: func(agent *Agent) error {
			return errors.New("insert failed")
		},
	}

	service := NewService(repo)

	_, err := service.Create("user-1", "group-1", "support", "helpful", ModelScopePlatform, "platform-default", "")
	if err == nil || err.Error() != "insert failed" {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestServiceCreateRequiresModelSelection(t *testing.T) {
	allowRuntimeEnginesForTest(t, settings.RuntimeEngineConfig{EngineType: EngineTypeHermes, Enabled: true})
	service := NewService(&stubRepository{})

	_, err := service.Create("user-1", "group-1", "support", "helpful", "", "", "")
	if err == nil || err.Error() != "model scope is required" {
		t.Fatalf("expected model scope validation error, got %v", err)
	}
}

func TestServiceCreateRequiresConfiguredPlatformModel(t *testing.T) {
	allowRuntimeEnginesForTest(t, settings.RuntimeEngineConfig{EngineType: EngineTypeHermes, Enabled: true})
	originalModelLoader := loadModelByID
	loadModelByID = func(id string) (*model.LLMModel, error) {
		return nil, nil
	}
	defer func() { loadModelByID = originalModelLoader }()

	service := NewService(&stubRepository{})

	_, err := service.Create("user-1", "group-1", "support", "helpful", ModelScopePlatform, "platform-default", "")
	if err == nil || err.Error() != "platform model not found" {
		t.Fatalf("expected platform model validation error, got %v", err)
	}
}

func TestServiceCreateAcceptsNanobotEngine(t *testing.T) {
	allowRuntimeEnginesForTest(
		t,
		settings.RuntimeEngineConfig{EngineType: EngineTypeHermes, Enabled: true},
		settings.RuntimeEngineConfig{EngineType: EngineTypeNanobot, Enabled: true},
	)
	originalModelLoader := loadModelByID
	loadModelByID = func(id string) (*model.LLMModel, error) {
		return &model.LLMModel{Id: id, Scope: model.ScopePlatform}, nil
	}
	defer func() { loadModelByID = originalModelLoader }()

	var created *Agent
	repo := &stubRepository{
		createFunc: func(agent *Agent) error {
			created = agent
			return nil
		},
		getByIdFunc: func(id string) (*Agent, error) {
			return &Agent{Id: id, EngineType: EngineTypeNanobot}, nil
		},
	}

	service := NewService(repo)
	service.idGenerator = func() string { return "agent-456" }
	service.apiKeyGenerator = func() string { return "generated-key" }

	got, err := service.Create("user-1", "group-1", "support", "helpful", ModelScopePlatform, "platform-default", EngineTypeNanobot)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created == nil || created.EngineType != EngineTypeNanobot {
		t.Fatalf("expected nanobot engine, got %#v", created)
	}
	if got == nil || got.EngineType != EngineTypeNanobot {
		t.Fatalf("expected created agent to keep nanobot engine, got %#v", got)
	}
}

func TestServiceCreateRejectsInvalidEngine(t *testing.T) {
	allowRuntimeEnginesForTest(t, settings.RuntimeEngineConfig{EngineType: EngineTypeHermes, Enabled: true})
	originalModelLoader := loadModelByID
	loadModelByID = func(id string) (*model.LLMModel, error) {
		return &model.LLMModel{Id: id, Scope: model.ScopePlatform}, nil
	}
	defer func() { loadModelByID = originalModelLoader }()

	service := NewService(&stubRepository{})

	_, err := service.Create("user-1", "group-1", "support", "helpful", ModelScopePlatform, "platform-default", "unknown")
	if err == nil || err.Error() != "engine type is invalid" {
		t.Fatalf("expected engine validation error, got %v", err)
	}
}

func TestToPublicIncludesNormalizedEngineType(t *testing.T) {
	got := toPublic(&Agent{
		Id:           "agent-1",
		AgentName:    "assistant",
		SystemPrompt: "prompt",
		ModelScope:   ModelScopeEnterprise,
		ModelId:      "",
		EngineType:   "",
	})

	if got.EngineType != EngineTypeHermes {
		t.Fatalf("expected default public engine type %q, got %q", EngineTypeHermes, got.EngineType)
	}

	got = toPublic(&Agent{
		Id:           "agent-2",
		AgentName:    "assistant",
		SystemPrompt: "prompt",
		ModelScope:   ModelScopeEnterprise,
		ModelId:      "",
		EngineType:   EngineTypeNanobot,
	})
	if got.EngineType != EngineTypeNanobot {
		t.Fatalf("expected public engine type %q, got %q", EngineTypeNanobot, got.EngineType)
	}
}
