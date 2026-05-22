package agent

import (
	"errors"
	"testing"
	"time"
)

type stubRepository struct {
	getByIdFunc       func(id string) (*Agent, error)
	listByUserIdFunc  func(userId, enterpriseId string) ([]*Agent, error)
	belongsToUserFunc func(id, userId, enterpriseId string) (bool, error)
	createFunc        func(agent *Agent) error
	updateFunc        func(id, agentName, systemPrompt string, updatedAt time.Time) error
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

func (s *stubRepository) Update(id, agentName, systemPrompt string, updatedAt time.Time) error {
	if s.updateFunc != nil {
		return s.updateFunc(id, agentName, systemPrompt, updatedAt)
	}
	return nil
}

func (s *stubRepository) Delete(id string) error {
	if s.deleteFunc != nil {
		return s.deleteFunc(id)
	}
	return nil
}

func TestServiceCreateGeneratesDefaults(t *testing.T) {
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
				EngineAPIKey: "generated-key",
				EngineType:   defaultEngineType,
			}, nil
		},
	}

	service := NewService(repo)
	service.idGenerator = func() string { return "agent-123" }
	service.apiKeyGenerator = func() string { return "generated-key" }

	got, err := service.Create("user-1", "group-1", "support", "helpful")
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
	expectedTime := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	repo := &stubRepository{
		updateFunc: func(id, agentName, systemPrompt string, updatedAt time.Time) error {
			if id != "agent-1" {
				t.Fatalf("unexpected id %q", id)
			}
			if agentName != "renamed" {
				t.Fatalf("unexpected name %q", agentName)
			}
			if systemPrompt != "new prompt" {
				t.Fatalf("unexpected prompt %q", systemPrompt)
			}
			if !updatedAt.Equal(expectedTime) {
				t.Fatalf("unexpected updatedAt %v", updatedAt)
			}
			return nil
		},
	}

	service := NewService(repo)
	service.now = func() time.Time { return expectedTime }

	if err := service.Update("agent-1", "renamed", "new prompt"); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
}

func TestServiceCreatePropagatesRepositoryError(t *testing.T) {
	repo := &stubRepository{
		createFunc: func(agent *Agent) error {
			return errors.New("insert failed")
		},
	}

	service := NewService(repo)

	_, err := service.Create("user-1", "group-1", "support", "helpful")
	if err == nil || err.Error() != "insert failed" {
		t.Fatalf("expected repository error, got %v", err)
	}
}
