package metering

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"dotblue/internal/domains/model"
)

type stubUsageRepo struct {
	inserted       *UsageEvent
	completed      *UsageEvent
	current        *UsageEvent
	dayUsage       *UsageAggregate
	monthUsage     *UsageAggregate
	aggregateCalls int
	dailyItems     []UsageDailyAggregate
}

func (s *stubUsageRepo) InsertStarted(item *UsageEvent) error {
	s.inserted = item
	s.current = item
	return nil
}

func (s *stubUsageRepo) Complete(invocationId string, item *UsageEvent) error {
	s.completed = item
	s.current = item
	return nil
}

func (s *stubUsageRepo) Fail(invocationId, errorCode string, completedAt time.Time) error {
	return nil
}

func (s *stubUsageRepo) GetByInvocationID(invocationId string) (*UsageEvent, error) {
	return s.current, nil
}

func (s *stubUsageRepo) UpsertDailyAggregate(item *UsageDailyAggregate) error {
	if item != nil {
		s.dailyItems = append(s.dailyItems, *item)
	}
	return nil
}

func (s *stubUsageRepo) ListEvents(filter UsageEventFilter) ([]UsageEvent, int, error) {
	return nil, 0, nil
}

func (s *stubUsageRepo) AggregateUsage(scopeType, scopeId string, from, to time.Time) (*UsageAggregate, error) {
	s.aggregateCalls++
	if s.aggregateCalls == 1 {
		return s.dayUsage, nil
	}
	return s.monthUsage, nil
}

func (s *stubUsageRepo) DailyTrends(scopeType, scopeId string, from, to time.Time) ([]TrendPoint, error) {
	return nil, nil
}

type stubPriceRepo struct {
	items []ModelPrice
}

func (s *stubPriceRepo) ListPrices(filter PriceFilter) ([]ModelPrice, error) {
	result := make([]ModelPrice, 0)
	for _, item := range s.items {
		if filter.ScopeType != "" && item.ScopeType != filter.ScopeType {
			continue
		}
		if filter.ScopeId != "" && item.ScopeId != filter.ScopeId {
			continue
		}
		if filter.ModelId != "" && item.ModelId != filter.ModelId {
			continue
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *stubPriceRepo) GetByID(id string) (*ModelPrice, error) { return nil, nil }
func (s *stubPriceRepo) InsertPrice(item *ModelPrice) error     { return nil }
func (s *stubPriceRepo) UpdatePrice(item *ModelPrice) error     { return nil }
func (s *stubPriceRepo) DeletePrice(id string) error            { return nil }

type stubPolicyRepo struct {
	policy *LimitPolicy
	err    error
}

func (s *stubPolicyRepo) ListPolicies(scopeType, scopeId string) ([]LimitPolicy, error) {
	return nil, nil
}

func (s *stubPolicyRepo) GetByID(id string) (*LimitPolicy, error) { return nil, nil }
func (s *stubPolicyRepo) InsertPolicy(item *LimitPolicy) error    { return nil }
func (s *stubPolicyRepo) UpdatePolicy(item *LimitPolicy) error    { return nil }
func (s *stubPolicyRepo) DeletePolicy(id string) error            { return nil }

func (s *stubPolicyRepo) FindPolicy(scopeType, scopeId string) (*LimitPolicy, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.policy == nil {
		return nil, nil
	}
	if s.policy.ScopeType == scopeType && s.policy.ScopeId == scopeId {
		return s.policy, nil
	}
	return nil, nil
}

type stubModelLookup struct {
	item *model.LLMModel
	err  error
}

func (s *stubModelLookup) GetByID(id string) (*model.LLMModel, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.item == nil || s.item.Id != id {
		return nil, nil
	}
	return s.item, nil
}

func TestStartAndCompleteInvocationUsesResolvedPrice(t *testing.T) {
	usageRepo := &stubUsageRepo{}
	svc := &Service{
		usageRepo: usageRepo,
		priceRepo: &stubPriceRepo{
			items: []ModelPrice{{
				Id:                    "price-1",
				ModelId:               "model-1",
				ScopeType:             ScopePlatform,
				Currency:              "USD",
				CostInputUnitPrice:    1,
				CostOutputUnitPrice:   2,
				ChargeInputUnitPrice:  3,
				ChargeOutputUnitPrice: 4,
			}},
		},
		policyRepo:  &stubPolicyRepo{},
		models:      &stubModelLookup{item: &model.LLMModel{Id: "model-1", Scope: model.ScopePlatform, Type: "openai", DisplayName: "GPT-4o"}},
		now:         func() time.Time { return time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC) },
		idGenerator: func() string { return "fixed-id" },
	}

	started, err := svc.StartInvocation(StartInvocationInput{
		ConversationId: "conv-1",
		AgentId:        "agent-1",
		UserId:         "user-1",
		ModelId:        "model-1",
	})
	if err != nil {
		t.Fatalf("StartInvocation() error = %v", err)
	}
	if started.Currency != "USD" || started.ChargeInputUnitPrice != 3 {
		t.Fatalf("unexpected price snapshot: %+v", started)
	}

	completed, err := svc.CompleteInvocation(CompleteInvocationInput{
		InvocationId: started.InvocationId,
		MessageId:    "msg-1",
		Usage: UsageSummary{
			PromptTokens:     1000,
			CompletionTokens: 500,
			Source:           UsageSourceEstimated,
		},
	})
	if err != nil {
		t.Fatalf("CompleteInvocation() error = %v", err)
	}
	if completed.ChargeAmount <= 0 || completed.CostAmount <= 0 {
		t.Fatalf("expected positive charge and cost, got %+v", completed)
	}
	if len(usageRepo.dailyItems) != 3 {
		t.Fatalf("dailyItems len = %d, want 3", len(usageRepo.dailyItems))
	}
}

func TestCheckLimitRejectsExceededDailyCharge(t *testing.T) {
	svc := &Service{
		usageRepo: &stubUsageRepo{
			dayUsage:   &UsageAggregate{ChargeAmount: 120},
			monthUsage: &UsageAggregate{ChargeAmount: 120},
		},
		priceRepo:  &stubPriceRepo{},
		policyRepo: &stubPolicyRepo{policy: &LimitPolicy{ScopeType: ScopeEnterprise, ScopeId: "ent-1", Enabled: true, HardLimit: true, DailyChargeLimit: 100}},
		models:     &stubModelLookup{},
		now:        func() time.Time { return time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC) },
	}

	err := svc.CheckLimit(CheckLimitInput{EnterpriseId: "ent-1"})
	if err == nil {
		t.Fatalf("CheckLimit() expected error")
	}
	if err.Error() != "enterprise charge daily limit exceeded" {
		t.Fatalf("unexpected error = %v", err)
	}
}

func TestCheckLimitIgnoresSqlErrNoRows(t *testing.T) {
	svc := &Service{
		usageRepo:  &stubUsageRepo{},
		priceRepo:  &stubPriceRepo{},
		policyRepo: &stubPolicyRepo{err: sql.ErrNoRows},
		models:     &stubModelLookup{},
		now:        func() time.Time { return time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC) },
	}

	if err := svc.CheckLimit(CheckLimitInput{
		EnterpriseId: "ent-1",
		UserId:       "user-1",
		AgentId:      "agent-1",
	}); err != nil {
		t.Fatalf("CheckLimit() unexpected error = %v", err)
	}
}

func TestStartInvocationRequiresKnownModel(t *testing.T) {
	svc := &Service{
		usageRepo:  &stubUsageRepo{},
		priceRepo:  &stubPriceRepo{},
		policyRepo: &stubPolicyRepo{},
		models:     &stubModelLookup{err: errors.New("lookup failed")},
		now:        time.Now,
		idGenerator: func() string {
			return "id"
		},
	}

	if _, err := svc.StartInvocation(StartInvocationInput{ModelId: "missing"}); err == nil {
		t.Fatalf("StartInvocation() expected error")
	}
}
