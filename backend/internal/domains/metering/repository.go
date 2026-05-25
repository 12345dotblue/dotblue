package metering

import "time"

type UsageEventRepository interface {
	InsertStarted(item *UsageEvent) error
	Complete(invocationId string, item *UsageEvent) error
	Fail(invocationId, errorCode string, completedAt time.Time) error
	GetByInvocationID(invocationId string) (*UsageEvent, error)
	UpsertDailyAggregate(item *UsageDailyAggregate) error
	ListEvents(filter UsageEventFilter) ([]UsageEvent, int, error)
	AggregateUsage(scopeType, scopeId string, from, to time.Time) (*UsageAggregate, error)
	DailyTrends(scopeType, scopeId string, from, to time.Time) ([]TrendPoint, error)
}

type PriceRepository interface {
	ListPrices(filter PriceFilter) ([]ModelPrice, error)
	GetByID(id string) (*ModelPrice, error)
	InsertPrice(item *ModelPrice) error
	UpdatePrice(item *ModelPrice) error
	DeletePrice(id string) error
}

type LimitPolicyRepository interface {
	ListPolicies(scopeType, scopeId string) ([]LimitPolicy, error)
	GetByID(id string) (*LimitPolicy, error)
	FindPolicy(scopeType, scopeId string) (*LimitPolicy, error)
	InsertPolicy(item *LimitPolicy) error
	UpdatePolicy(item *LimitPolicy) error
	DeletePolicy(id string) error
}
