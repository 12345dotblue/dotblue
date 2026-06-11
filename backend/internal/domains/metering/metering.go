package metering

import (
	"time"
)

const (
	ScopePlatform   = "platform"
	ScopeEnterprise = "enterprise"
	ScopeUser       = "user"
	ScopeAgent      = "agent"

	StatusStarted   = "started"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusCancelled = "cancelled"

	UsageSourceReported  = "reported"
	UsageSourceEstimated = "estimated"
)

type UsageEvent struct {
	Id                    string     `json:"id"`
	InvocationId          string     `json:"invocationId" orm:"invocation_id"`
	RequestId             string     `json:"requestId" orm:"request_id"`
	ConversationId        string     `json:"conversationId" orm:"conversation_id"`
	MessageId             string     `json:"messageId" orm:"message_id"`
	AgentId               string     `json:"agentId" orm:"agent_id"`
	EnterpriseId          string     `json:"enterpriseId" orm:"enterprise_id"`
	UserId                string     `json:"userId" orm:"user_id"`
	SourceType            string     `json:"sourceType" orm:"source_type"`
	SourceConnectionId    string     `json:"sourceConnectionId" orm:"source_connection_id"`
	ModelId               string     `json:"modelId" orm:"model_id"`
	ModelScope            string     `json:"modelScope" orm:"model_scope"`
	ProviderType          string     `json:"providerType" orm:"provider_type"`
	ModelNameSnapshot     string     `json:"modelNameSnapshot" orm:"model_name_snapshot"`
	FundingType           string     `json:"fundingType" orm:"funding_type"`
	CreditType            string     `json:"creditType" orm:"credit_type"`
	CreditPriceBookId     string     `json:"creditPriceBookId" orm:"credit_price_book_id"`
	CreditUnitUsdSnapshot float64    `json:"creditUnitUsdSnapshot" orm:"credit_unit_usd_snapshot"`
	InputCreditsPer1M     int64      `json:"inputCreditsPer1M" orm:"input_credits_per_1m_snapshot"`
	OutputCreditsPer1M    int64      `json:"outputCreditsPer1M" orm:"output_credits_per_1m_snapshot"`
	ReservedCredits       int64      `json:"reservedCredits" orm:"reserved_credits"`
	SettledCredits        int64      `json:"settledCredits" orm:"settled_credits"`
	Status                string     `json:"status" orm:"status"`
	UsageSource           string     `json:"usageSource" orm:"usage_source"`
	PromptTokens          int64      `json:"promptTokens" orm:"prompt_tokens"`
	CompletionTokens      int64      `json:"completionTokens" orm:"completion_tokens"`
	ReasoningTokens       int64      `json:"reasoningTokens" orm:"reasoning_tokens"`
	CacheReadTokens       int64      `json:"cacheReadTokens" orm:"cache_read_tokens"`
	CacheWriteTokens      int64      `json:"cacheWriteTokens" orm:"cache_write_tokens"`
	TotalTokens           int64      `json:"totalTokens" orm:"total_tokens"`
	Currency              string     `json:"currency" orm:"currency"`
	CostInputUnitPrice    float64    `json:"costInputUnitPrice" orm:"cost_input_unit_price"`
	CostOutputUnitPrice   float64    `json:"costOutputUnitPrice" orm:"cost_output_unit_price"`
	ChargeInputUnitPrice  float64    `json:"chargeInputUnitPrice" orm:"charge_input_unit_price"`
	ChargeOutputUnitPrice float64    `json:"chargeOutputUnitPrice" orm:"charge_output_unit_price"`
	CostAmount            float64    `json:"costAmount" orm:"cost_amount"`
	ChargeAmount          float64    `json:"chargeAmount" orm:"charge_amount"`
	RawUsageJSON          string     `json:"rawUsageJson" orm:"raw_usage_json"`
	RawResponseMetaJSON   string     `json:"rawResponseMetaJson" orm:"raw_response_meta_json"`
	ErrorCode             string     `json:"errorCode" orm:"error_code"`
	CreatedAt             time.Time  `json:"createdAt" orm:"created_at"`
	CompletedAt           *time.Time `json:"completedAt,omitempty" orm:"completed_at"`
}

type ModelPrice struct {
	Id                    string    `json:"id"`
	ModelId               string    `json:"modelId" orm:"model_id"`
	ScopeType             string    `json:"scopeType" orm:"scope_type"`
	ScopeId               string    `json:"scopeId" orm:"scope_id"`
	Currency              string    `json:"currency" orm:"currency"`
	CostInputUnitPrice    float64   `json:"costInputUnitPrice" orm:"cost_input_unit_price"`
	CostOutputUnitPrice   float64   `json:"costOutputUnitPrice" orm:"cost_output_unit_price"`
	ChargeInputUnitPrice  float64   `json:"chargeInputUnitPrice" orm:"charge_input_unit_price"`
	ChargeOutputUnitPrice float64   `json:"chargeOutputUnitPrice" orm:"charge_output_unit_price"`
	CreatedAt             time.Time `json:"createdAt" orm:"created_at"`
	UpdatedAt             time.Time `json:"updatedAt" orm:"updated_at"`
}

type LimitPolicy struct {
	Id                 string    `json:"id"`
	ScopeType          string    `json:"scopeType" orm:"scope_type"`
	ScopeId            string    `json:"scopeId" orm:"scope_id"`
	Enabled            bool      `json:"enabled" orm:"enabled"`
	DailyTokenLimit    int64     `json:"dailyTokenLimit" orm:"daily_token_limit"`
	MonthlyTokenLimit  int64     `json:"monthlyTokenLimit" orm:"monthly_token_limit"`
	DailyChargeLimit   float64   `json:"dailyChargeLimit" orm:"daily_charge_limit"`
	MonthlyChargeLimit float64   `json:"monthlyChargeLimit" orm:"monthly_charge_limit"`
	HardLimit          bool      `json:"hardLimit" orm:"hard_limit"`
	CreatedAt          time.Time `json:"createdAt" orm:"created_at"`
	UpdatedAt          time.Time `json:"updatedAt" orm:"updated_at"`
}

type UsageSummary struct {
	PromptTokens     int64  `json:"promptTokens"`
	CompletionTokens int64  `json:"completionTokens"`
	ReasoningTokens  int64  `json:"reasoningTokens"`
	CacheReadTokens  int64  `json:"cacheReadTokens"`
	CacheWriteTokens int64  `json:"cacheWriteTokens"`
	TotalTokens      int64  `json:"totalTokens"`
	Source           string `json:"source"`
	RawUsageJSON     string `json:"rawUsageJson"`
}

type Overview struct {
	TodayRequests int64   `json:"todayRequests"`
	TodayTokens   int64   `json:"todayTokens"`
	TodayCost     float64 `json:"todayCost"`
	TodayCharge   float64 `json:"todayCharge"`
	MonthRequests int64   `json:"monthRequests"`
	MonthTokens   int64   `json:"monthTokens"`
	MonthCost     float64 `json:"monthCost"`
	MonthCharge   float64 `json:"monthCharge"`
}

type TrendPoint struct {
	Date         string  `json:"date"`
	RequestCount int64   `json:"requestCount"`
	TotalTokens  int64   `json:"totalTokens"`
	CostAmount   float64 `json:"costAmount"`
	ChargeAmount float64 `json:"chargeAmount"`
}

type UsageAggregate struct {
	RequestCount int64   `json:"requestCount"`
	TotalTokens  int64   `json:"totalTokens"`
	CostAmount   float64 `json:"costAmount"`
	ChargeAmount float64 `json:"chargeAmount"`
}

type UsageDailyAggregate struct {
	Id           string    `json:"id"`
	StatDate     string    `json:"statDate" orm:"stat_date"`
	ScopeType    string    `json:"scopeType" orm:"scope_type"`
	ScopeId      string    `json:"scopeId" orm:"scope_id"`
	SourceType   string    `json:"sourceType" orm:"source_type"`
	RequestCount int64     `json:"requestCount" orm:"request_count"`
	TotalTokens  int64     `json:"totalTokens" orm:"total_tokens"`
	CostAmount   float64   `json:"costAmount" orm:"cost_amount"`
	ChargeAmount float64   `json:"chargeAmount" orm:"charge_amount"`
	CreatedAt    time.Time `json:"createdAt" orm:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" orm:"updated_at"`
}

type StartInvocationInput struct {
	RequestId          string
	ConversationId     string
	AgentId            string
	EnterpriseId       string
	UserId             string
	SourceType         string
	SourceConnectionId string
	ModelId            string
	ModelScope         string
}

type CompleteInvocationInput struct {
	InvocationId        string
	MessageId           string
	Usage               UsageSummary
	RawResponseMetaJSON string
}

type FailInvocationInput struct {
	InvocationId string
	ErrorCode    string
}

type CheckLimitInput struct {
	EnterpriseId string
	UserId       string
	AgentId      string
}

type UpdateCreditSnapshotInput struct {
	InvocationId           string
	FundingType            string
	CreditType             string
	CreditPriceBookId      string
	CreditUnitUsdSnapshot  float64
	InputCreditsPer1M      int64
	OutputCreditsPer1M     int64
	ReservedCredits        int64
	SettledCredits         int64
}

type PriceFilter struct {
	ScopeType string
	ScopeId   string
	ModelId   string
}

type UsageEventFilter struct {
	EnterpriseId string
	AgentId      string
	ModelId      string
	Status       string
	SourceType   string
	Page         int
	PageSize     int
}
