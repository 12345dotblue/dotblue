package credit

import "time"

const (
	CreditTypePlatform   = "platform"
	CreditTypeEnterprise = "enterprise"
)

const (
	ScopeEnterprise = "enterprise"
	ScopeMember     = "member"
	ScopeAgent      = "agent"
)

const (
	FundingTypePlatform   = "platform_funded"
	FundingTypeEnterprise = "enterprise_funded"
)

const (
	ModelSourceTypePlatform         = "platform_model"
	ModelSourceTypeEnterpriseCustom = "enterprise_custom_model"

	CurrencyUSD = "USD"
	CurrencyCNY = "CNY"
)

const (
	EntryTypeGrant   = "grant"
	EntryTypeReserve = "reserve"
	EntryTypeSettle  = "settle"
	EntryTypeRelease = "release"
	EntryTypeRefund  = "refund"
	EntryTypeExpire  = "expire"
	EntryTypeAdjust  = "adjust"
)

const (
	DirectionIn  = "in"
	DirectionOut = "out"
)

const (
	PriceBookStatusActive   = "active"
	PriceBookStatusDisabled = "disabled"
)

const DefaultCreditUnitUSD = 0.0001

type CreditWallet struct {
	Id               string    `json:"id"`
	EnterpriseId     string    `json:"enterpriseId" orm:"enterprise_id"`
	CreditType       string    `json:"creditType" orm:"credit_type"`
	TotalCredits     int64     `json:"totalCredits" orm:"total_credits"`
	ReservedCredits  int64     `json:"reservedCredits" orm:"reserved_credits"`
	AvailableCredits int64     `json:"availableCredits" orm:"available_credits"`
	Version          int64     `json:"version" orm:"version"`
	CreatedAt        time.Time `json:"createdAt" orm:"created_at"`
	UpdatedAt        time.Time `json:"updatedAt" orm:"updated_at"`
}

type CreditGrant struct {
	Id               string     `json:"id"`
	EnterpriseId     string     `json:"enterpriseId" orm:"enterprise_id"`
	CreditType       string     `json:"creditType" orm:"credit_type"`
	SourceType       string     `json:"sourceType" orm:"source_type"`
	SourceRefId      string     `json:"sourceRefId" orm:"source_ref_id"`
	GrantedCredits   int64      `json:"grantedCredits" orm:"granted_credits"`
	RemainingCredits int64      `json:"remainingCredits" orm:"remaining_credits"`
	EffectiveAt      time.Time  `json:"effectiveAt" orm:"effective_at"`
	ExpiresAt        *time.Time `json:"expiresAt,omitempty" orm:"expires_at"`
	MetadataJson     string     `json:"metadataJson" orm:"metadata_json"`
	CreatedAt        time.Time  `json:"createdAt" orm:"created_at"`
}

type CreditLedgerEntry struct {
	Id              string    `json:"id"`
	EnterpriseId    string    `json:"enterpriseId" orm:"enterprise_id"`
	CreditType      string    `json:"creditType" orm:"credit_type"`
	WalletId        string    `json:"walletId" orm:"wallet_id"`
	GrantId         string    `json:"grantId" orm:"grant_id"`
	EntryType       string    `json:"entryType" orm:"entry_type"`
	Direction       string    `json:"direction" orm:"direction"`
	Credits         int64     `json:"credits" orm:"credits"`
	BalanceAfter    int64     `json:"balanceAfter" orm:"balance_after"`
	ReservedAfter   int64     `json:"reservedAfter" orm:"reserved_after"`
	InvocationId    string    `json:"invocationId" orm:"invocation_id"`
	MemberUserId    string    `json:"memberUserId" orm:"member_user_id"`
	AgentId         string    `json:"agentId" orm:"agent_id"`
	BudgetScopeType string    `json:"budgetScopeType" orm:"budget_scope_type"`
	BudgetScopeId   string    `json:"budgetScopeId" orm:"budget_scope_id"`
	ReasonCode      string    `json:"reasonCode" orm:"reason_code"`
	SnapshotJson    string    `json:"snapshotJson" orm:"snapshot_json"`
	CreatedAt       time.Time `json:"createdAt" orm:"created_at"`
}

type CreditReservation struct {
	Id                string     `json:"id"`
	EnterpriseId      string     `json:"enterpriseId" orm:"enterprise_id"`
	CreditType        string     `json:"creditType" orm:"credit_type"`
	InvocationId      string     `json:"invocationId" orm:"invocation_id"`
	MemberUserId      string     `json:"memberUserId" orm:"member_user_id"`
	AgentId           string     `json:"agentId" orm:"agent_id"`
	ModelId           string     `json:"modelId" orm:"model_id"`
	ModelScope        string     `json:"modelScope" orm:"model_scope"`
	FundingType       string     `json:"fundingType" orm:"funding_type"`
	PriceBookId       string     `json:"priceBookId" orm:"price_book_id"`
	PriceSnapshotJson string     `json:"priceSnapshotJson" orm:"price_snapshot_json"`
	ReservedCredits   int64      `json:"reservedCredits" orm:"reserved_credits"`
	Status            string     `json:"status" orm:"status"`
	ExpiresAt         *time.Time `json:"expiresAt,omitempty" orm:"expires_at"`
	CreatedAt         time.Time  `json:"createdAt" orm:"created_at"`
	UpdatedAt         time.Time  `json:"updatedAt" orm:"updated_at"`
}

type CreditPriceBook struct {
	Id                     string    `json:"id"`
	EnterpriseId           string    `json:"enterpriseId" orm:"enterprise_id"`
	CreditType             string    `json:"creditType" orm:"credit_type"`
	ModelId                string    `json:"modelId" orm:"model_id"`
	ModelScope             string    `json:"modelScope" orm:"model_scope"`
	ModelSourceType        string    `json:"modelSourceType" orm:"model_source_type"`
	FundingType            string    `json:"fundingType" orm:"funding_type"`
	Currency               string    `json:"currency" orm:"currency"`
	CreditUnitUsd          float64   `json:"creditUnitUsd" orm:"credit_unit_usd"`
	CostInputUsdPer1M      float64   `json:"costInputUsdPer1M" orm:"cost_input_usd_per_1m"`
	CostOutputUsdPer1M     float64   `json:"costOutputUsdPer1M" orm:"cost_output_usd_per_1m"`
	PlatformMultiplier     float64   `json:"platformMultiplier" orm:"platform_multiplier"`
	EnterpriseMultiplier   float64   `json:"enterpriseMultiplier" orm:"enterprise_multiplier"`
	BillableInputUsdPer1M  float64   `json:"billableInputUsdPer1M" orm:"billable_input_usd_per_1m"`
	BillableOutputUsdPer1M float64   `json:"billableOutputUsdPer1M" orm:"billable_output_usd_per_1m"`
	InputCreditsPer1M      int64     `json:"inputCreditsPer1M" orm:"input_credits_per_1m"`
	OutputCreditsPer1M     int64     `json:"outputCreditsPer1M" orm:"output_credits_per_1m"`
	EffectiveAt            time.Time `json:"effectiveAt" orm:"effective_at"`
	Status                 string    `json:"status" orm:"status"`
	CreatedAt              time.Time `json:"createdAt" orm:"created_at"`
	UpdatedAt              time.Time `json:"updatedAt" orm:"updated_at"`
}

type CreditBudgetPolicy struct {
	Id                 string    `json:"id"`
	EnterpriseId       string    `json:"enterpriseId" orm:"enterprise_id"`
	CreditType         string    `json:"creditType" orm:"credit_type"`
	ScopeType          string    `json:"scopeType" orm:"scope_type"`
	ScopeId            string    `json:"scopeId" orm:"scope_id"`
	Enabled            bool      `json:"enabled" orm:"enabled"`
	DailyCreditLimit   int64     `json:"dailyCreditLimit" orm:"daily_credit_limit"`
	MonthlyCreditLimit int64     `json:"monthlyCreditLimit" orm:"monthly_credit_limit"`
	DailyTokenLimit    int64     `json:"dailyTokenLimit" orm:"daily_token_limit"`
	MonthlyTokenLimit  int64     `json:"monthlyTokenLimit" orm:"monthly_token_limit"`
	DailyUsdLimit      float64   `json:"dailyUsdLimit" orm:"daily_usd_limit"`
	MonthlyUsdLimit    float64   `json:"monthlyUsdLimit" orm:"monthly_usd_limit"`
	HardLimit          bool      `json:"hardLimit" orm:"hard_limit"`
	CreatedAt          time.Time `json:"createdAt" orm:"created_at"`
	UpdatedAt          time.Time `json:"updatedAt" orm:"updated_at"`
}

type WalletFilter struct {
	EnterpriseId string `json:"enterpriseId"`
	CreditType   string `json:"creditType"`
}

type GrantFilter struct {
	EnterpriseId string `json:"enterpriseId"`
	CreditType   string `json:"creditType"`
}

type LedgerFilter struct {
	EnterpriseId string `json:"enterpriseId"`
	CreditType   string `json:"creditType"`
}

type PriceBookFilter struct {
	EnterpriseId string `json:"enterpriseId"`
	CreditType   string `json:"creditType"`
	ModelId      string `json:"modelId"`
}

type BudgetPolicyFilter struct {
	EnterpriseId string `json:"enterpriseId"`
	CreditType   string `json:"creditType"`
	ScopeType    string `json:"scopeType"`
	ScopeId      string `json:"scopeId"`
}

type LedgerAggregate struct {
	CreditType     string `json:"creditType"`
	GrantedCredits int64  `json:"grantedCredits"`
	SettledCredits int64  `json:"settledCredits"`
	ExpiredCredits int64  `json:"expiredCredits"`
}

type WalletOverview struct {
	CreditType       string `json:"creditType"`
	TotalCredits     int64  `json:"totalCredits"`
	ReservedCredits  int64  `json:"reservedCredits"`
	AvailableCredits int64  `json:"availableCredits"`
	GrantedCredits   int64  `json:"grantedCredits"`
	SettledCredits   int64  `json:"settledCredits"`
	ExpiredCredits   int64  `json:"expiredCredits"`
}

type Overview struct {
	EnterpriseId string           `json:"enterpriseId"`
	Wallets      []WalletOverview `json:"wallets"`
}

type CreateGrantReq struct {
	SourceType   string     `json:"sourceType"`
	SourceRefId  string     `json:"sourceRefId"`
	Credits      int64      `json:"credits"`
	EffectiveAt  *time.Time `json:"effectiveAt,omitempty"`
	ExpiresAt    *time.Time `json:"expiresAt,omitempty"`
	MetadataJson string     `json:"metadataJson"`
	ReasonCode   string     `json:"reasonCode"`
}

type CreatePriceBookReq struct {
	CreditType           string    `json:"creditType"`
	ModelId              string    `json:"modelId"`
	ModelScope           string    `json:"modelScope"`
	ModelSourceType      string    `json:"modelSourceType"`
	FundingType          string    `json:"fundingType"`
	Currency             string    `json:"currency"`
	CreditUnitUsd        float64   `json:"creditUnitUsd"`
	CostInputUsdPer1M    float64   `json:"costInputUsdPer1M"`
	CostOutputUsdPer1M   float64   `json:"costOutputUsdPer1M"`
	PlatformMultiplier   float64   `json:"platformMultiplier"`
	EnterpriseMultiplier float64   `json:"enterpriseMultiplier"`
	EffectiveAt          time.Time `json:"effectiveAt"`
	Status               string    `json:"status"`
}

type UpdatePriceBookReq = CreatePriceBookReq

type CreateBudgetPolicyReq struct {
	CreditType         string  `json:"creditType"`
	ScopeType          string  `json:"scopeType"`
	ScopeId            string  `json:"scopeId"`
	Enabled            bool    `json:"enabled"`
	DailyCreditLimit   int64   `json:"dailyCreditLimit"`
	MonthlyCreditLimit int64   `json:"monthlyCreditLimit"`
	DailyTokenLimit    int64   `json:"dailyTokenLimit"`
	MonthlyTokenLimit  int64   `json:"monthlyTokenLimit"`
	DailyUsdLimit      float64 `json:"dailyUsdLimit"`
	MonthlyUsdLimit    float64 `json:"monthlyUsdLimit"`
	HardLimit          bool    `json:"hardLimit"`
}

type UpdateBudgetPolicyReq = CreateBudgetPolicyReq

type ReserveInput struct {
	InvocationId              string `json:"invocationId"`
	EnterpriseId              string `json:"enterpriseId"`
	MemberUserId              string `json:"memberUserId"`
	AgentId                   string `json:"agentId"`
	ModelId                   string `json:"modelId"`
	ModelScope                string `json:"modelScope"`
	FundingType               string `json:"fundingType"`
	ModelSourceType           string `json:"modelSourceType"`
	EstimatedPromptTokens     int64  `json:"estimatedPromptTokens"`
	EstimatedCompletionTokens int64  `json:"estimatedCompletionTokens"`
}

type SettleInput struct {
	InvocationId     string `json:"invocationId"`
	PromptTokens     int64  `json:"promptTokens"`
	CompletionTokens int64  `json:"completionTokens"`
}

type CreditSnapshot struct {
	CreditType           string  `json:"creditType"`
	FundingType          string  `json:"fundingType"`
	PriceBookId          string  `json:"priceBookId"`
	CreditUnitUsd        float64 `json:"creditUnitUsd"`
	InputCreditsPer1M    int64   `json:"inputCreditsPer1M"`
	OutputCreditsPer1M   int64   `json:"outputCreditsPer1M"`
	ReservedCredits      int64   `json:"reservedCredits"`
	SettledCredits       int64   `json:"settledCredits"`
}
