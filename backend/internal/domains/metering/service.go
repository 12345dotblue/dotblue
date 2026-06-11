package metering

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"dotblue/internal/domains/model"
	"github.com/google/uuid"
)

type modelLookup interface {
	GetByID(id string) (*model.LLMModel, error)
}

type defaultModelLookup struct{}

func (defaultModelLookup) GetByID(id string) (*model.LLMModel, error) {
	return model.GetByID(id)
}

type Service struct {
	usageRepo   UsageEventRepository
	priceRepo   PriceRepository
	policyRepo  LimitPolicyRepository
	models      modelLookup
	now         func() time.Time
	idGenerator func() string
}

func NewService(usageRepo UsageEventRepository, priceRepo PriceRepository, policyRepo LimitPolicyRepository, models modelLookup) *Service {
	return &Service{
		usageRepo:   usageRepo,
		priceRepo:   priceRepo,
		policyRepo:  policyRepo,
		models:      models,
		now:         time.Now,
		idGenerator: func() string { return uuid.NewString() },
	}
}

var defaultService = NewService(
	NewGFUsageEventRepository(),
	NewGFPriceRepository(),
	NewGFLimitPolicyRepository(),
	defaultModelLookup{},
)

type CreatePriceReq struct {
	ModelId               string  `json:"modelId" v:"required"`
	Currency              string  `json:"currency"`
	CostInputUnitPrice    float64 `json:"costInputUnitPrice"`
	CostOutputUnitPrice   float64 `json:"costOutputUnitPrice"`
	ChargeInputUnitPrice  float64 `json:"chargeInputUnitPrice"`
	ChargeOutputUnitPrice float64 `json:"chargeOutputUnitPrice"`
}

type UpdatePriceReq struct {
	Currency              string  `json:"currency"`
	CostInputUnitPrice    float64 `json:"costInputUnitPrice"`
	CostOutputUnitPrice   float64 `json:"costOutputUnitPrice"`
	ChargeInputUnitPrice  float64 `json:"chargeInputUnitPrice"`
	ChargeOutputUnitPrice float64 `json:"chargeOutputUnitPrice"`
}

type CreateLimitPolicyReq struct {
	Enabled            bool    `json:"enabled"`
	DailyTokenLimit    int64   `json:"dailyTokenLimit"`
	MonthlyTokenLimit  int64   `json:"monthlyTokenLimit"`
	DailyChargeLimit   float64 `json:"dailyChargeLimit"`
	MonthlyChargeLimit float64 `json:"monthlyChargeLimit"`
	HardLimit          bool    `json:"hardLimit"`
}

func StartInvocation(input StartInvocationInput) (*UsageEvent, error) {
	return defaultService.StartInvocation(input)
}

func CompleteInvocation(input CompleteInvocationInput) (*UsageEvent, error) {
	return defaultService.CompleteInvocation(input)
}

func FailInvocation(input FailInvocationInput) error {
	return defaultService.FailInvocation(input)
}

func CheckLimit(input CheckLimitInput) error {
	return defaultService.CheckLimit(input)
}

func GetOverview(scopeType, scopeId string) (*Overview, error) {
	return defaultService.GetOverview(scopeType, scopeId)
}

func GetTrends(scopeType, scopeId string, days int) ([]TrendPoint, error) {
	return defaultService.GetTrends(scopeType, scopeId, days)
}

func ListUsageEvents(filter UsageEventFilter) ([]UsageEvent, int, error) {
	return defaultService.ListUsageEvents(filter)
}

func ListPrices(filter PriceFilter) ([]ModelPrice, error) {
	return defaultService.ListPrices(filter)
}

func CreatePrice(scopeType, scopeId string, req CreatePriceReq) (*ModelPrice, error) {
	return defaultService.CreatePrice(scopeType, scopeId, req)
}

func UpdatePrice(id, scopeType, scopeId string, req UpdatePriceReq) (*ModelPrice, error) {
	return defaultService.UpdatePrice(id, scopeType, scopeId, req)
}

func DeletePrice(id, scopeType, scopeId string) error {
	return defaultService.DeletePrice(id, scopeType, scopeId)
}

func ListPolicies(scopeType, scopeId string) ([]LimitPolicy, error) {
	return defaultService.ListPolicies(scopeType, scopeId)
}

func CreateLimitPolicy(scopeType, scopeId string, req CreateLimitPolicyReq) (*LimitPolicy, error) {
	return defaultService.CreateLimitPolicy(scopeType, scopeId, req)
}

func UpdateLimitPolicy(id, scopeType, scopeId string, req CreateLimitPolicyReq) (*LimitPolicy, error) {
	return defaultService.UpdateLimitPolicy(id, scopeType, scopeId, req)
}

func DeleteLimitPolicy(id, scopeType, scopeId string) error {
	return defaultService.DeleteLimitPolicy(id, scopeType, scopeId)
}

func UpdateCreditSnapshot(input UpdateCreditSnapshotInput) error {
	return defaultService.UpdateCreditSnapshot(input)
}

func (s *Service) StartInvocation(input StartInvocationInput) (*UsageEvent, error) {
	if s == nil || s.usageRepo == nil || s.models == nil {
		return nil, errors.New("metering service is not configured")
	}
	input.ModelId = strings.TrimSpace(input.ModelId)
	if input.ModelId == "" {
		return nil, errors.New("model id is required")
	}
	modelItem, err := s.models.GetByID(input.ModelId)
	if err != nil {
		return nil, err
	}
	if modelItem == nil {
		return nil, errors.New("model not found")
	}
	price := s.resolvePrice(modelItem, strings.TrimSpace(input.EnterpriseId))
	now := s.now()
	event := &UsageEvent{
		Id:                    s.idGenerator(),
		InvocationId:          s.idGenerator(),
		RequestId:             strings.TrimSpace(input.RequestId),
		ConversationId:        strings.TrimSpace(input.ConversationId),
		AgentId:               strings.TrimSpace(input.AgentId),
		EnterpriseId:          strings.TrimSpace(input.EnterpriseId),
		UserId:                strings.TrimSpace(input.UserId),
		SourceType:            normalizeSourceType(input.SourceType),
		SourceConnectionId:    strings.TrimSpace(input.SourceConnectionId),
		ModelId:               modelItem.Id,
		ModelScope:            modelItem.Scope,
		ProviderType:          strings.TrimSpace(modelItem.Type),
		ModelNameSnapshot:     strings.TrimSpace(modelItem.DisplayName),
		Status:                StatusStarted,
		UsageSource:           UsageSourceEstimated,
		Currency:              price.Currency,
		CostInputUnitPrice:    price.CostInputUnitPrice,
		CostOutputUnitPrice:   price.CostOutputUnitPrice,
		ChargeInputUnitPrice:  price.ChargeInputUnitPrice,
		ChargeOutputUnitPrice: price.ChargeOutputUnitPrice,
		CreatedAt:             now,
	}
	if err := s.usageRepo.InsertStarted(event); err != nil {
		return nil, err
	}
	return event, nil
}

func (s *Service) CompleteInvocation(input CompleteInvocationInput) (*UsageEvent, error) {
	if s == nil || s.usageRepo == nil {
		return nil, errors.New("metering service is not configured")
	}
	existing, err := s.usageRepo.GetByInvocationID(strings.TrimSpace(input.InvocationId))
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("usage invocation not found")
	}
	now := s.now()
	usage := normalizeUsageSummary(input.Usage)
	existing.MessageId = strings.TrimSpace(input.MessageId)
	existing.Status = StatusCompleted
	existing.UsageSource = usage.Source
	existing.PromptTokens = usage.PromptTokens
	existing.CompletionTokens = usage.CompletionTokens
	existing.ReasoningTokens = usage.ReasoningTokens
	existing.CacheReadTokens = usage.CacheReadTokens
	existing.CacheWriteTokens = usage.CacheWriteTokens
	existing.TotalTokens = usage.TotalTokens
	existing.RawUsageJSON = strings.TrimSpace(usage.RawUsageJSON)
	existing.RawResponseMetaJSON = strings.TrimSpace(input.RawResponseMetaJSON)
	existing.CostAmount = calculateAmount(existing.CostInputUnitPrice, existing.CostOutputUnitPrice, usage.PromptTokens, usage.CompletionTokens)
	existing.ChargeAmount = calculateAmount(existing.ChargeInputUnitPrice, existing.ChargeOutputUnitPrice, usage.PromptTokens, usage.CompletionTokens)
	existing.CompletedAt = &now
	if err := s.usageRepo.Complete(existing.InvocationId, existing); err != nil {
		return nil, err
	}
	for _, aggregate := range s.buildDailyAggregates(existing) {
		if err := s.usageRepo.UpsertDailyAggregate(&aggregate); err != nil {
			return nil, err
		}
	}
	return existing, nil
}

func (s *Service) FailInvocation(input FailInvocationInput) error {
	if s == nil || s.usageRepo == nil {
		return errors.New("metering service is not configured")
	}
	return s.usageRepo.Fail(strings.TrimSpace(input.InvocationId), strings.TrimSpace(input.ErrorCode), s.now())
}

func (s *Service) UpdateCreditSnapshot(input UpdateCreditSnapshotInput) error {
	if s == nil || s.usageRepo == nil {
		return errors.New("metering service is not configured")
	}
	invocationID := strings.TrimSpace(input.InvocationId)
	if invocationID == "" {
		return errors.New("invocation id is required")
	}
	existing, err := s.usageRepo.GetByInvocationID(invocationID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("usage invocation not found")
	}
	existing.FundingType = strings.TrimSpace(input.FundingType)
	existing.CreditType = strings.TrimSpace(input.CreditType)
	existing.CreditPriceBookId = strings.TrimSpace(input.CreditPriceBookId)
	existing.CreditUnitUsdSnapshot = max0(input.CreditUnitUsdSnapshot)
	existing.InputCreditsPer1M = maxInt0(input.InputCreditsPer1M)
	existing.OutputCreditsPer1M = maxInt0(input.OutputCreditsPer1M)
	existing.ReservedCredits = maxInt0(input.ReservedCredits)
	existing.SettledCredits = maxInt0(input.SettledCredits)
	return s.usageRepo.UpdateCreditSnapshot(invocationID, existing)
}

func (s *Service) CheckLimit(input CheckLimitInput) error {
	if s == nil || s.policyRepo == nil || s.usageRepo == nil {
		return errors.New("metering service is not configured")
	}
	now := s.now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	candidates := []struct {
		scopeType string
		scopeId   string
	}{
		{scopeType: ScopeAgent, scopeId: strings.TrimSpace(input.AgentId)},
		{scopeType: ScopeUser, scopeId: strings.TrimSpace(input.UserId)},
		{scopeType: ScopeEnterprise, scopeId: strings.TrimSpace(input.EnterpriseId)},
		{scopeType: ScopePlatform, scopeId: ""},
	}
	for _, candidate := range candidates {
		if candidate.scopeType != ScopePlatform && candidate.scopeId == "" {
			continue
		}
		policy, err := s.policyRepo.FindPolicy(candidate.scopeType, candidate.scopeId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return err
		}
		if policy == nil || !policy.Enabled || !policy.HardLimit {
			continue
		}
		if err := s.validatePolicyUsage(candidate.scopeType, candidate.scopeId, policy, dayStart, monthStart, now); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) validatePolicyUsage(scopeType, scopeId string, policy *LimitPolicy, dayStart, monthStart, now time.Time) error {
	dayUsage, err := s.usageRepo.AggregateUsage(scopeType, scopeId, dayStart, now.Add(time.Second))
	if err != nil {
		return err
	}
	monthUsage, err := s.usageRepo.AggregateUsage(scopeType, scopeId, monthStart, now.Add(time.Second))
	if err != nil {
		return err
	}
	// Apply the narrowest scope first so the caller gets the most actionable limit reason.
	if policy.DailyTokenLimit > 0 && dayUsage.TotalTokens >= policy.DailyTokenLimit {
		return fmt.Errorf("%s token daily limit exceeded", scopeType)
	}
	if policy.MonthlyTokenLimit > 0 && monthUsage.TotalTokens >= policy.MonthlyTokenLimit {
		return fmt.Errorf("%s token monthly limit exceeded", scopeType)
	}
	if policy.DailyChargeLimit > 0 && dayUsage.ChargeAmount >= policy.DailyChargeLimit {
		return fmt.Errorf("%s charge daily limit exceeded", scopeType)
	}
	if policy.MonthlyChargeLimit > 0 && monthUsage.ChargeAmount >= policy.MonthlyChargeLimit {
		return fmt.Errorf("%s charge monthly limit exceeded", scopeType)
	}
	return nil
}

func (s *Service) GetOverview(scopeType, scopeId string) (*Overview, error) {
	now := s.now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	dayUsage, err := s.usageRepo.AggregateUsage(scopeType, scopeId, dayStart, now.Add(time.Second))
	if err != nil {
		return nil, err
	}
	monthUsage, err := s.usageRepo.AggregateUsage(scopeType, scopeId, monthStart, now.Add(time.Second))
	if err != nil {
		return nil, err
	}
	return &Overview{
		TodayRequests: dayUsage.RequestCount,
		TodayTokens:   dayUsage.TotalTokens,
		TodayCost:     dayUsage.CostAmount,
		TodayCharge:   dayUsage.ChargeAmount,
		MonthRequests: monthUsage.RequestCount,
		MonthTokens:   monthUsage.TotalTokens,
		MonthCost:     monthUsage.CostAmount,
		MonthCharge:   monthUsage.ChargeAmount,
	}, nil
}

func (s *Service) GetTrends(scopeType, scopeId string, days int) ([]TrendPoint, error) {
	if days <= 0 {
		days = 7
	}
	now := s.now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(days - 1))
	return s.usageRepo.DailyTrends(scopeType, scopeId, start, now.Add(24*time.Hour))
}

func (s *Service) ListUsageEvents(filter UsageEventFilter) ([]UsageEvent, int, error) {
	if s == nil || s.usageRepo == nil {
		return nil, 0, errors.New("metering service is not configured")
	}
	return s.usageRepo.ListEvents(filter)
}

func (s *Service) ListPrices(filter PriceFilter) ([]ModelPrice, error) {
	if s == nil || s.priceRepo == nil {
		return nil, errors.New("metering service is not configured")
	}
	return s.priceRepo.ListPrices(filter)
}

func (s *Service) CreatePrice(scopeType, scopeId string, req CreatePriceReq) (*ModelPrice, error) {
	if s == nil || s.priceRepo == nil || s.models == nil {
		return nil, errors.New("metering service is not configured")
	}
	modelID := strings.TrimSpace(req.ModelId)
	if modelID == "" {
		return nil, errors.New("model id is required")
	}
	modelItem, err := s.models.GetByID(modelID)
	if err != nil {
		return nil, err
	}
	if modelItem == nil {
		return nil, errors.New("model not found")
	}
	if err := validatePriceOwnership(modelItem, scopeType, scopeId); err != nil {
		return nil, err
	}
	now := s.now()
	item := &ModelPrice{
		Id:                    s.idGenerator(),
		ModelId:               modelID,
		ScopeType:             scopeType,
		ScopeId:               scopeId,
		Currency:              normalizeCurrency(req.Currency),
		CostInputUnitPrice:    max0(req.CostInputUnitPrice),
		CostOutputUnitPrice:   max0(req.CostOutputUnitPrice),
		ChargeInputUnitPrice:  max0(req.ChargeInputUnitPrice),
		ChargeOutputUnitPrice: max0(req.ChargeOutputUnitPrice),
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	if err := s.priceRepo.InsertPrice(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) UpdatePrice(id, scopeType, scopeId string, req UpdatePriceReq) (*ModelPrice, error) {
	item, err := s.priceRepo.GetByID(strings.TrimSpace(id))
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("price not found")
	}
	if item.ScopeType != scopeType || item.ScopeId != scopeId {
		return nil, errors.New("price scope mismatch")
	}
	item.Currency = normalizeCurrency(req.Currency)
	item.CostInputUnitPrice = max0(req.CostInputUnitPrice)
	item.CostOutputUnitPrice = max0(req.CostOutputUnitPrice)
	item.ChargeInputUnitPrice = max0(req.ChargeInputUnitPrice)
	item.ChargeOutputUnitPrice = max0(req.ChargeOutputUnitPrice)
	item.UpdatedAt = s.now()
	if err := s.priceRepo.UpdatePrice(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) DeletePrice(id, scopeType, scopeId string) error {
	item, err := s.priceRepo.GetByID(strings.TrimSpace(id))
	if err != nil {
		return err
	}
	if item == nil {
		return errors.New("price not found")
	}
	if item.ScopeType != scopeType || item.ScopeId != scopeId {
		return errors.New("price scope mismatch")
	}
	return s.priceRepo.DeletePrice(item.Id)
}

func (s *Service) ListPolicies(scopeType, scopeId string) ([]LimitPolicy, error) {
	if s == nil || s.policyRepo == nil {
		return nil, errors.New("metering service is not configured")
	}
	return s.policyRepo.ListPolicies(scopeType, scopeId)
}

func (s *Service) CreateLimitPolicy(scopeType, scopeId string, req CreateLimitPolicyReq) (*LimitPolicy, error) {
	now := s.now()
	item := &LimitPolicy{
		Id:                 s.idGenerator(),
		ScopeType:          scopeType,
		ScopeId:            scopeId,
		Enabled:            req.Enabled,
		DailyTokenLimit:    maxInt0(req.DailyTokenLimit),
		MonthlyTokenLimit:  maxInt0(req.MonthlyTokenLimit),
		DailyChargeLimit:   max0(req.DailyChargeLimit),
		MonthlyChargeLimit: max0(req.MonthlyChargeLimit),
		HardLimit:          req.HardLimit,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := s.policyRepo.InsertPolicy(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) UpdateLimitPolicy(id, scopeType, scopeId string, req CreateLimitPolicyReq) (*LimitPolicy, error) {
	item, err := s.policyRepo.GetByID(strings.TrimSpace(id))
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("policy not found")
	}
	if item.ScopeType != scopeType || item.ScopeId != scopeId {
		return nil, errors.New("policy scope mismatch")
	}
	item.Enabled = req.Enabled
	item.DailyTokenLimit = maxInt0(req.DailyTokenLimit)
	item.MonthlyTokenLimit = maxInt0(req.MonthlyTokenLimit)
	item.DailyChargeLimit = max0(req.DailyChargeLimit)
	item.MonthlyChargeLimit = max0(req.MonthlyChargeLimit)
	item.HardLimit = req.HardLimit
	item.UpdatedAt = s.now()
	if err := s.policyRepo.UpdatePolicy(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) DeleteLimitPolicy(id, scopeType, scopeId string) error {
	item, err := s.policyRepo.GetByID(strings.TrimSpace(id))
	if err != nil {
		return err
	}
	if item == nil {
		return errors.New("policy not found")
	}
	if item.ScopeType != scopeType || item.ScopeId != scopeId {
		return errors.New("policy scope mismatch")
	}
	return s.policyRepo.DeletePolicy(item.Id)
}

func (s *Service) resolvePrice(modelItem *model.LLMModel, enterpriseId string) *ModelPrice {
	if s == nil || s.priceRepo == nil || modelItem == nil {
		return &ModelPrice{Currency: "USD"}
	}
	if enterpriseId != "" {
		list, err := s.priceRepo.ListPrices(PriceFilter{ScopeType: ScopeEnterprise, ScopeId: enterpriseId, ModelId: modelItem.Id})
		if err == nil && len(list) > 0 {
			return &list[0]
		}
	}
	list, err := s.priceRepo.ListPrices(PriceFilter{ScopeType: ScopePlatform, ScopeId: "", ModelId: modelItem.Id})
	if err == nil && len(list) > 0 {
		return &list[0]
	}
	return &ModelPrice{Currency: "USD"}
}

// buildDailyAggregates writes the completed invocation into every scope that may enforce limits.
func (s *Service) buildDailyAggregates(event *UsageEvent) []UsageDailyAggregate {
	if s == nil || event == nil {
		return nil
	}
	completedAt := s.now()
	if event.CompletedAt != nil {
		completedAt = *event.CompletedAt
	}
	statDate := completedAt.Format("2006-01-02")
	now := s.now()
	scopes := []struct {
		scopeType string
		scopeID   string
	}{
		{scopeType: ScopePlatform, scopeID: ""},
		{scopeType: ScopeEnterprise, scopeID: strings.TrimSpace(event.EnterpriseId)},
		{scopeType: ScopeUser, scopeID: strings.TrimSpace(event.UserId)},
		{scopeType: ScopeAgent, scopeID: strings.TrimSpace(event.AgentId)},
	}
	items := make([]UsageDailyAggregate, 0, len(scopes))
	for _, scope := range scopes {
		if scope.scopeType != ScopePlatform && scope.scopeID == "" {
			continue
		}
		items = append(items, UsageDailyAggregate{
			Id:           s.idGenerator(),
			StatDate:     statDate,
			ScopeType:    scope.scopeType,
			ScopeId:      scope.scopeID,
			SourceType:   normalizeSourceType(event.SourceType),
			RequestCount: 1,
			TotalTokens:  event.TotalTokens,
			CostAmount:   event.CostAmount,
			ChargeAmount: event.ChargeAmount,
			CreatedAt:    now,
			UpdatedAt:    now,
		})
	}
	return items
}

func validatePriceOwnership(modelItem *model.LLMModel, scopeType, scopeId string) error {
	if modelItem == nil {
		return errors.New("model not found")
	}
	switch scopeType {
	case ScopePlatform:
		if modelItem.Scope != model.ScopePlatform {
			return errors.New("platform price only supports platform models")
		}
	case ScopeEnterprise:
		if modelItem.Scope != model.ScopeEnterprise || strings.TrimSpace(modelItem.EnterpriseId) != strings.TrimSpace(scopeId) {
			return errors.New("enterprise price only supports current enterprise models")
		}
	default:
		return errors.New("invalid price scope")
	}
	return nil
}

// normalizeUsageSummary ensures downstream accounting always sees a complete token shape.
func normalizeUsageSummary(usage UsageSummary) UsageSummary {
	usage.Source = strings.TrimSpace(usage.Source)
	if usage.Source == "" {
		usage.Source = UsageSourceEstimated
	}
	usage.PromptTokens = maxInt0(usage.PromptTokens)
	usage.CompletionTokens = maxInt0(usage.CompletionTokens)
	usage.ReasoningTokens = maxInt0(usage.ReasoningTokens)
	usage.CacheReadTokens = maxInt0(usage.CacheReadTokens)
	usage.CacheWriteTokens = maxInt0(usage.CacheWriteTokens)
	if usage.TotalTokens <= 0 {
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens + usage.ReasoningTokens + usage.CacheReadTokens + usage.CacheWriteTokens
	}
	return usage
}

func calculateAmount(inputPrice, outputPrice float64, promptTokens, completionTokens int64) float64 {
	value := (float64(promptTokens)*inputPrice + float64(completionTokens)*outputPrice) / 1_000_000
	return math.Round(value*1_000_000) / 1_000_000
}

func normalizeSourceType(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "web"
	}
	return value
}

func normalizeCurrency(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return "USD"
	}
	return value
}

func max0(value float64) float64 {
	if value < 0 {
		return 0
	}
	return value
}

func maxInt0(value int64) int64 {
	if value < 0 {
		return 0
	}
	return value
}
