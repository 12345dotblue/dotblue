package model

import (
	"testing"
	"time"
)

type stubRepository struct {
	items map[string]*LLMModel
}

func newStubRepository() *stubRepository {
	return &stubRepository{items: map[string]*LLMModel{}}
}

func (r *stubRepository) ListByScope(scope, enterpriseId string) ([]LLMModel, error) {
	list := make([]LLMModel, 0)
	for _, item := range r.items {
		if item.Scope != scope {
			continue
		}
		if enterpriseId != "" && item.EnterpriseId != enterpriseId {
			continue
		}
		list = append(list, *item)
	}
	return list, nil
}

func (r *stubRepository) GetByID(id string) (*LLMModel, error) {
	item := r.items[id]
	if item == nil {
		return nil, nil
	}
	copy := *item
	return &copy, nil
}

func (r *stubRepository) Insert(item *LLMModel) error {
	copy := *item
	r.items[item.Id] = &copy
	return nil
}

func (r *stubRepository) Update(item *LLMModel) error {
	copy := *item
	r.items[item.Id] = &copy
	return nil
}

func (r *stubRepository) Delete(id string) error {
	delete(r.items, id)
	return nil
}

func (r *stubRepository) ClearDefault(scope string) error {
	for _, item := range r.items {
		if item.Scope == scope {
			item.IsDefault = false
		}
	}
	return nil
}

func (r *stubRepository) UpdateDefault(id string, isDefault bool, updatedAt time.Time) error {
	if item := r.items[id]; item != nil {
		item.IsDefault = isDefault
		item.UpdatedAt = updatedAt
	}
	return nil
}

func TestCreatePlatformModelUsesExplicitPlatformRouting(t *testing.T) {
	repo := newStubRepository()
	svc := NewService(repo)
	svc.now = func() time.Time { return time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC) }
	svc.idGenerator = func() string { return "model-1" }

	item, err := svc.Create(ScopePlatform, "", CreateReq{
		DisplayName: "GPT",
		Type:        "openai",
		Model:       "gpt-4.1",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if item.FundingType != FundingTypePlatform {
		t.Fatalf("unexpected funding type: %+v", item)
	}
	if item.ModelSourceType != ModelSourceTypePlatform {
		t.Fatalf("unexpected model source type: %+v", item)
	}
}

func TestCreateEnterpriseModelRejectsPlatformFunding(t *testing.T) {
	repo := newStubRepository()
	svc := NewService(repo)
	svc.idGenerator = func() string { return "model-2" }

	_, err := svc.Create(ScopeEnterprise, "ent-1", CreateReq{
		DisplayName: "Claude",
		Type:        "anthropic",
		Model:       "claude-3-7-sonnet",
		FundingType: FundingTypePlatform,
	})
	if err == nil {
		t.Fatalf("Create expected funding validation error")
	}
	if err.Error() != "enterprise models must use enterprise_funded" {
		t.Fatalf("unexpected error: %v", err)
	}
}
