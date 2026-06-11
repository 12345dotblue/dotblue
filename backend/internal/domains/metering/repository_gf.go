package metering

import (
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

type GFUsageEventRepository struct{}
type GFPriceRepository struct{}
type GFLimitPolicyRepository struct{}

func NewGFUsageEventRepository() *GFUsageEventRepository {
	return &GFUsageEventRepository{}
}

func NewGFPriceRepository() *GFPriceRepository {
	return &GFPriceRepository{}
}

func NewGFLimitPolicyRepository() *GFLimitPolicyRepository {
	return &GFLimitPolicyRepository{}
}

func (r *GFUsageEventRepository) InsertStarted(item *UsageEvent) error {
	_, err := g.DB().Model("llm_usage_events").Data(g.Map{
		"id":                       item.Id,
		"invocation_id":            item.InvocationId,
		"request_id":               item.RequestId,
		"conversation_id":          nullableUUIDValue(item.ConversationId),
		"message_id":               nullableUUIDValue(item.MessageId),
		"agent_id":                 nullableUUIDValue(item.AgentId),
		"enterprise_id":            item.EnterpriseId,
		"user_id":                  item.UserId,
		"source_type":              item.SourceType,
		"source_connection_id":     nullableUUIDValue(item.SourceConnectionId),
		"model_id":                 item.ModelId,
		"model_scope":              item.ModelScope,
		"provider_type":            item.ProviderType,
		"model_name_snapshot":      item.ModelNameSnapshot,
		"funding_type":             item.FundingType,
		"credit_type":              item.CreditType,
		"credit_price_book_id":     item.CreditPriceBookId,
		"credit_unit_usd_snapshot": item.CreditUnitUsdSnapshot,
		"input_credits_per_1m_snapshot":  item.InputCreditsPer1M,
		"output_credits_per_1m_snapshot": item.OutputCreditsPer1M,
		"reserved_credits":         item.ReservedCredits,
		"settled_credits":          item.SettledCredits,
		"status":                   item.Status,
		"usage_source":             item.UsageSource,
		"currency":                 item.Currency,
		"cost_input_unit_price":    item.CostInputUnitPrice,
		"cost_output_unit_price":   item.CostOutputUnitPrice,
		"charge_input_unit_price":  item.ChargeInputUnitPrice,
		"charge_output_unit_price": item.ChargeOutputUnitPrice,
		"cost_amount":              item.CostAmount,
		"charge_amount":            item.ChargeAmount,
		"raw_usage_json":           normalizedJSONBValue(item.RawUsageJSON),
		"raw_response_meta_json":   normalizedJSONBValue(item.RawResponseMetaJSON),
		"error_code":               item.ErrorCode,
		"created_at":               item.CreatedAt,
	}).Insert()
	return err
}

func (r *GFUsageEventRepository) Complete(invocationId string, item *UsageEvent) error {
	data := g.Map{
		"message_id":             nullableUUIDValue(item.MessageId),
		"status":                 item.Status,
		"usage_source":           item.UsageSource,
		"prompt_tokens":          item.PromptTokens,
		"completion_tokens":      item.CompletionTokens,
		"reasoning_tokens":       item.ReasoningTokens,
		"cache_read_tokens":      item.CacheReadTokens,
		"cache_write_tokens":     item.CacheWriteTokens,
		"total_tokens":           item.TotalTokens,
		"cost_amount":            item.CostAmount,
		"charge_amount":          item.ChargeAmount,
		"raw_usage_json":         normalizedJSONBValue(item.RawUsageJSON),
		"raw_response_meta_json": normalizedJSONBValue(item.RawResponseMetaJSON),
		"completed_at":           item.CompletedAt,
	}
	_, err := g.DB().Model("llm_usage_events").Data(data).Where("invocation_id = ?", invocationId).Update()
	return err
}

func nullableUUIDValue(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.TrimSpace(value)
}

func normalizedJSONBValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "{}"
	}
	return value
}

func (r *GFUsageEventRepository) Fail(invocationId, errorCode string, completedAt time.Time) error {
	_, err := g.DB().Model("llm_usage_events").Data(g.Map{
		"status":       StatusFailed,
		"error_code":   errorCode,
		"completed_at": completedAt,
	}).Where("invocation_id = ?", invocationId).Update()
	return err
}

func (r *GFUsageEventRepository) UpdateCreditSnapshot(invocationId string, item *UsageEvent) error {
	_, err := g.DB().Model("llm_usage_events").Data(g.Map{
		"funding_type":                  item.FundingType,
		"credit_type":                   item.CreditType,
		"credit_price_book_id":          item.CreditPriceBookId,
		"credit_unit_usd_snapshot":      item.CreditUnitUsdSnapshot,
		"input_credits_per_1m_snapshot": item.InputCreditsPer1M,
		"output_credits_per_1m_snapshot": item.OutputCreditsPer1M,
		"reserved_credits":              item.ReservedCredits,
		"settled_credits":               item.SettledCredits,
	}).Where("invocation_id = ?", invocationId).Update()
	return err
}

func (r *GFUsageEventRepository) GetByInvocationID(invocationId string) (*UsageEvent, error) {
	var item UsageEvent
	if err := g.DB().Model("llm_usage_events").Where("invocation_id = ?", invocationId).Limit(1).Scan(&item); err != nil {
		return nil, err
	}
	if item.Id == "" {
		return nil, nil
	}
	return &item, nil
}

func (r *GFUsageEventRepository) UpsertDailyAggregate(item *UsageDailyAggregate) error {
	if item == nil {
		return nil
	}
	var existing UsageDailyAggregate
	err := g.DB().Model("llm_usage_daily_aggregates").
		Where("stat_date = ? AND scope_type = ? AND scope_id = ? AND source_type = ?", item.StatDate, item.ScopeType, item.ScopeId, item.SourceType).
		Limit(1).
		Scan(&existing)
	if err != nil {
		return err
	}
	if existing.Id == "" {
		_, err = g.DB().Model("llm_usage_daily_aggregates").Data(g.Map{
			"id":            item.Id,
			"stat_date":     item.StatDate,
			"scope_type":    item.ScopeType,
			"scope_id":      item.ScopeId,
			"source_type":   item.SourceType,
			"request_count": item.RequestCount,
			"total_tokens":  item.TotalTokens,
			"cost_amount":   item.CostAmount,
			"charge_amount": item.ChargeAmount,
			"created_at":    item.CreatedAt,
			"updated_at":    item.UpdatedAt,
		}).Insert()
		return err
	}
	_, err = g.DB().Model("llm_usage_daily_aggregates").Data(g.Map{
		"request_count": existing.RequestCount + item.RequestCount,
		"total_tokens":  existing.TotalTokens + item.TotalTokens,
		"cost_amount":   existing.CostAmount + item.CostAmount,
		"charge_amount": existing.ChargeAmount + item.ChargeAmount,
		"updated_at":    item.UpdatedAt,
	}).Where("id = ?", existing.Id).Update()
	return err
}

func (r *GFUsageEventRepository) ListEvents(filter UsageEventFilter) ([]UsageEvent, int, error) {
	query := usageEventQuery(filter)
	total, err := query.Count()
	if err != nil {
		return nil, 0, err
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []UsageEvent
	err = query.Order("created_at DESC").Page(page, pageSize).Scan(&list)
	return list, total, err
}

func (r *GFUsageEventRepository) AggregateUsage(scopeType, scopeId string, from, to time.Time) (*UsageAggregate, error) {
	result, err := r.aggregateUsageFromDaily(scopeType, scopeId, from, to)
	if err != nil {
		return nil, err
	}
	if result.RequestCount > 0 || result.TotalTokens > 0 || result.CostAmount > 0 || result.ChargeAmount > 0 {
		return result, nil
	}
	return r.aggregateUsageFromEvents(scopeType, scopeId, from, to)
}

func (r *GFUsageEventRepository) aggregateUsageFromDaily(scopeType, scopeId string, from, to time.Time) (*UsageAggregate, error) {
	model := g.DB().Model("llm_usage_daily_aggregates")
	model = applyAggregateScopeFilter(model, scopeType, scopeId)
	if !from.IsZero() {
		model = model.Where("stat_date >= ?", from.Format("2006-01-02"))
	}
	if !to.IsZero() {
		model = model.Where("stat_date <= ?", to.Add(-time.Nanosecond).Format("2006-01-02"))
	}
	type row struct {
		RequestCount int64   `json:"requestCount"`
		TotalTokens  int64   `json:"totalTokens"`
		CostAmount   float64 `json:"costAmount"`
		ChargeAmount float64 `json:"chargeAmount"`
	}
	var result row
	if err := model.Fields("COALESCE(SUM(request_count), 0) AS request_count, COALESCE(SUM(total_tokens), 0) AS total_tokens, COALESCE(SUM(cost_amount), 0) AS cost_amount, COALESCE(SUM(charge_amount), 0) AS charge_amount").Scan(&result); err != nil {
		return nil, err
	}
	return &UsageAggregate{
		RequestCount: result.RequestCount,
		TotalTokens:  result.TotalTokens,
		CostAmount:   result.CostAmount,
		ChargeAmount: result.ChargeAmount,
	}, nil
}

func (r *GFUsageEventRepository) aggregateUsageFromEvents(scopeType, scopeId string, from, to time.Time) (*UsageAggregate, error) {
	model := g.DB().Model("llm_usage_events").Where("status = ?", StatusCompleted)
	model = applyScopeFilter(model, scopeType, scopeId)
	if !from.IsZero() {
		model = model.Where("created_at >= ?", from)
	}
	if !to.IsZero() {
		model = model.Where("created_at < ?", to)
	}
	type row struct {
		RequestCount int64   `json:"requestCount"`
		TotalTokens  int64   `json:"totalTokens"`
		CostAmount   float64 `json:"costAmount"`
		ChargeAmount float64 `json:"chargeAmount"`
	}
	var result row
	if err := model.Fields("COUNT(*) AS request_count, COALESCE(SUM(total_tokens), 0) AS total_tokens, COALESCE(SUM(cost_amount), 0) AS cost_amount, COALESCE(SUM(charge_amount), 0) AS charge_amount").Scan(&result); err != nil {
		return nil, err
	}
	return &UsageAggregate{
		RequestCount: result.RequestCount,
		TotalTokens:  result.TotalTokens,
		CostAmount:   result.CostAmount,
		ChargeAmount: result.ChargeAmount,
	}, nil
}

func (r *GFUsageEventRepository) DailyTrends(scopeType, scopeId string, from, to time.Time) ([]TrendPoint, error) {
	points, err := r.dailyTrendsFromDaily(scopeType, scopeId, from, to)
	if err != nil {
		return nil, err
	}
	if len(points) > 0 {
		return points, nil
	}
	return r.dailyTrendsFromEvents(scopeType, scopeId, from, to)
}

func (r *GFUsageEventRepository) dailyTrendsFromDaily(scopeType, scopeId string, from, to time.Time) ([]TrendPoint, error) {
	model := g.DB().Model("llm_usage_daily_aggregates")
	model = applyAggregateScopeFilter(model, scopeType, scopeId)
	if !from.IsZero() {
		model = model.Where("stat_date >= ?", from.Format("2006-01-02"))
	}
	if !to.IsZero() {
		model = model.Where("stat_date <= ?", to.Add(-time.Nanosecond).Format("2006-01-02"))
	}
	type row struct {
		Date         string  `json:"date"`
		RequestCount int64   `json:"requestCount"`
		TotalTokens  int64   `json:"totalTokens"`
		CostAmount   float64 `json:"costAmount"`
		ChargeAmount float64 `json:"chargeAmount"`
	}
	var rows []row
	if err := model.Fields("stat_date AS date, COALESCE(SUM(request_count), 0) AS request_count, COALESCE(SUM(total_tokens), 0) AS total_tokens, COALESCE(SUM(cost_amount), 0) AS cost_amount, COALESCE(SUM(charge_amount), 0) AS charge_amount").Group("stat_date").Order("stat_date ASC").Scan(&rows); err != nil {
		return nil, err
	}
	points := make([]TrendPoint, 0, len(rows))
	for i := range rows {
		points = append(points, TrendPoint(rows[i]))
	}
	return points, nil
}

func (r *GFUsageEventRepository) dailyTrendsFromEvents(scopeType, scopeId string, from, to time.Time) ([]TrendPoint, error) {
	model := g.DB().Model("llm_usage_events").Where("status = ?", StatusCompleted)
	model = applyScopeFilter(model, scopeType, scopeId)
	if !from.IsZero() {
		model = model.Where("created_at >= ?", from)
	}
	if !to.IsZero() {
		model = model.Where("created_at < ?", to)
	}
	type eventRow struct {
		CreatedAt    time.Time `json:"createdAt"`
		TotalTokens  int64     `json:"totalTokens"`
		CostAmount   float64   `json:"costAmount"`
		ChargeAmount float64   `json:"chargeAmount"`
	}
	var rows []eventRow
	if err := model.Fields("created_at, total_tokens, cost_amount, charge_amount").Order("created_at ASC").Scan(&rows); err != nil {
		return nil, err
	}
	grouped := make(map[string]*TrendPoint)
	for i := range rows {
		dateKey := rows[i].CreatedAt.Format("2006-01-02")
		point := grouped[dateKey]
		if point == nil {
			point = &TrendPoint{Date: dateKey}
			grouped[dateKey] = point
		}
		point.RequestCount++
		point.TotalTokens += rows[i].TotalTokens
		point.CostAmount += rows[i].CostAmount
		point.ChargeAmount += rows[i].ChargeAmount
	}
	points := make([]TrendPoint, 0, len(grouped))
	for current := from; !current.IsZero() && current.Before(to); current = current.AddDate(0, 0, 1) {
		dateKey := current.Format("2006-01-02")
		if point, ok := grouped[dateKey]; ok {
			points = append(points, *point)
		}
	}
	return points, nil
}

func (r *GFPriceRepository) ListPrices(filter PriceFilter) ([]ModelPrice, error) {
	query := g.DB().Model("llm_model_prices")
	if filter.ScopeType != "" {
		query = query.Where("scope_type = ?", filter.ScopeType)
	}
	if filter.ScopeId != "" {
		query = query.Where("scope_id = ?", filter.ScopeId)
	}
	if filter.ModelId != "" {
		query = query.Where("model_id = ?", filter.ModelId)
	}
	var list []ModelPrice
	err := query.Order("created_at DESC").Scan(&list)
	return list, err
}

func (r *GFPriceRepository) GetByID(id string) (*ModelPrice, error) {
	var item ModelPrice
	if err := g.DB().Model("llm_model_prices").Where("id = ?", id).Limit(1).Scan(&item); err != nil {
		return nil, err
	}
	if item.Id == "" {
		return nil, nil
	}
	return &item, nil
}

func (r *GFPriceRepository) InsertPrice(item *ModelPrice) error {
	_, err := g.DB().Model("llm_model_prices").Data(g.Map{
		"id":                       item.Id,
		"model_id":                 item.ModelId,
		"scope_type":               item.ScopeType,
		"scope_id":                 item.ScopeId,
		"currency":                 item.Currency,
		"cost_input_unit_price":    item.CostInputUnitPrice,
		"cost_output_unit_price":   item.CostOutputUnitPrice,
		"charge_input_unit_price":  item.ChargeInputUnitPrice,
		"charge_output_unit_price": item.ChargeOutputUnitPrice,
		"created_at":               item.CreatedAt,
		"updated_at":               item.UpdatedAt,
	}).Insert()
	return err
}

func (r *GFPriceRepository) UpdatePrice(item *ModelPrice) error {
	_, err := g.DB().Model("llm_model_prices").Data(g.Map{
		"currency":                 item.Currency,
		"cost_input_unit_price":    item.CostInputUnitPrice,
		"cost_output_unit_price":   item.CostOutputUnitPrice,
		"charge_input_unit_price":  item.ChargeInputUnitPrice,
		"charge_output_unit_price": item.ChargeOutputUnitPrice,
		"updated_at":               item.UpdatedAt,
	}).Where("id = ?", item.Id).Update()
	return err
}

func (r *GFPriceRepository) DeletePrice(id string) error {
	_, err := g.DB().Model("llm_model_prices").Where("id = ?", id).Delete()
	return err
}

func (r *GFLimitPolicyRepository) ListPolicies(scopeType, scopeId string) ([]LimitPolicy, error) {
	query := g.DB().Model("llm_usage_limit_policies")
	if scopeType != "" {
		query = query.Where("scope_type = ?", scopeType)
	}
	if scopeId != "" {
		query = query.Where("scope_id = ?", scopeId)
	}
	var list []LimitPolicy
	err := query.Order("created_at DESC").Scan(&list)
	return list, err
}

func (r *GFLimitPolicyRepository) GetByID(id string) (*LimitPolicy, error) {
	var item LimitPolicy
	if err := g.DB().Model("llm_usage_limit_policies").Where("id = ?", id).Limit(1).Scan(&item); err != nil {
		return nil, err
	}
	if item.Id == "" {
		return nil, nil
	}
	return &item, nil
}

func (r *GFLimitPolicyRepository) FindPolicy(scopeType, scopeId string) (*LimitPolicy, error) {
	var item LimitPolicy
	if err := g.DB().Model("llm_usage_limit_policies").Where("scope_type = ? AND scope_id = ?", scopeType, scopeId).Limit(1).Order("created_at DESC").Scan(&item); err != nil {
		return nil, err
	}
	if item.Id == "" {
		return nil, nil
	}
	return &item, nil
}

func (r *GFLimitPolicyRepository) InsertPolicy(item *LimitPolicy) error {
	_, err := g.DB().Model("llm_usage_limit_policies").Data(g.Map{
		"id":                   item.Id,
		"scope_type":           item.ScopeType,
		"scope_id":             item.ScopeId,
		"enabled":              item.Enabled,
		"daily_token_limit":    item.DailyTokenLimit,
		"monthly_token_limit":  item.MonthlyTokenLimit,
		"daily_charge_limit":   item.DailyChargeLimit,
		"monthly_charge_limit": item.MonthlyChargeLimit,
		"hard_limit":           item.HardLimit,
		"created_at":           item.CreatedAt,
		"updated_at":           item.UpdatedAt,
	}).Insert()
	return err
}

func (r *GFLimitPolicyRepository) UpdatePolicy(item *LimitPolicy) error {
	_, err := g.DB().Model("llm_usage_limit_policies").Data(g.Map{
		"enabled":              item.Enabled,
		"daily_token_limit":    item.DailyTokenLimit,
		"monthly_token_limit":  item.MonthlyTokenLimit,
		"daily_charge_limit":   item.DailyChargeLimit,
		"monthly_charge_limit": item.MonthlyChargeLimit,
		"hard_limit":           item.HardLimit,
		"updated_at":           item.UpdatedAt,
	}).Where("id = ?", item.Id).Update()
	return err
}

func (r *GFLimitPolicyRepository) DeletePolicy(id string) error {
	_, err := g.DB().Model("llm_usage_limit_policies").Where("id = ?", id).Delete()
	return err
}

func applyScopeFilter(query *gdb.Model, scopeType, scopeId string) *gdb.Model {
	switch scopeType {
	case ScopePlatform:
		return query
	case ScopeEnterprise:
		return query.Where("enterprise_id = ?", scopeId)
	case ScopeUser:
		return query.Where("user_id = ?", scopeId)
	case ScopeAgent:
		return query.Where("agent_id = ?", scopeId)
	default:
		return query
	}
}

func applyAggregateScopeFilter(query *gdb.Model, scopeType, scopeId string) *gdb.Model {
	switch scopeType {
	case ScopePlatform:
		return query.Where("scope_type = ?", ScopePlatform)
	case ScopeEnterprise:
		return query.Where("scope_type = ? AND scope_id = ?", ScopeEnterprise, scopeId)
	case ScopeUser:
		return query.Where("scope_type = ? AND scope_id = ?", ScopeUser, scopeId)
	case ScopeAgent:
		return query.Where("scope_type = ? AND scope_id = ?", ScopeAgent, scopeId)
	default:
		return query
	}
}

func usageEventQuery(filter UsageEventFilter) *gdb.Model {
	query := g.DB().Model("llm_usage_events")
	if filter.EnterpriseId != "" {
		query = query.Where("enterprise_id = ?", filter.EnterpriseId)
	}
	if filter.AgentId != "" {
		query = query.Where("agent_id = ?", filter.AgentId)
	}
	if filter.ModelId != "" {
		query = query.Where("model_id = ?", filter.ModelId)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.SourceType != "" {
		query = query.Where("source_type = ?", filter.SourceType)
	}
	return query
}
