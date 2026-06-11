package credit

import (
	"net/http"
	"strings"
	"time"

	"dotblue/internal/domains/identity"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type grantReq struct {
	EnterpriseId string `json:"enterpriseId"`
	CreditType   string `json:"creditType"`
	SourceType   string `json:"sourceType"`
	SourceRefId  string `json:"sourceRefId"`
	Credits      int64  `json:"credits"`
	EffectiveAt  string `json:"effectiveAt"`
	ExpiresAt    string `json:"expiresAt"`
	MetadataJson string `json:"metadataJson"`
	ReasonCode   string `json:"reasonCode"`
}

type priceBookReq struct {
	EnterpriseId         string  `json:"enterpriseId"`
	CreditType           string  `json:"creditType"`
	ModelId              string  `json:"modelId"`
	ModelScope           string  `json:"modelScope"`
	ModelSourceType      string  `json:"modelSourceType"`
	FundingType          string  `json:"fundingType"`
	Currency             string  `json:"currency"`
	CreditUnitUsd        float64 `json:"creditUnitUsd"`
	CostInputUsdPer1M    float64 `json:"costInputUsdPer1M"`
	CostOutputUsdPer1M   float64 `json:"costOutputUsdPer1M"`
	PlatformMultiplier   float64 `json:"platformMultiplier"`
	EnterpriseMultiplier float64 `json:"enterpriseMultiplier"`
	EffectiveAt          string  `json:"effectiveAt"`
	Status               string  `json:"status"`
}

type budgetPolicyReq struct {
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

func PlatformCreditOverviewHandler(r *ghttp.Request) {
	enterpriseId := strings.TrimSpace(r.Get("enterpriseId").String())
	item, err := GetOverview(enterpriseId)
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(item)
}

func EnterpriseCreditOverviewHandler(r *ghttp.Request) {
	item, err := GetOverview(identity.GetCurrentEnterpriseId(r))
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(item)
}

func PlatformCreditWalletsHandler(r *ghttp.Request) {
	list, err := ListWallets(WalletFilter{
		EnterpriseId: strings.TrimSpace(r.Get("enterpriseId").String()),
		CreditType:   strings.TrimSpace(r.Get("creditType").String()),
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(list)
}

func EnterpriseCreditWalletsHandler(r *ghttp.Request) {
	list, err := ListWallets(WalletFilter{
		EnterpriseId: identity.GetCurrentEnterpriseId(r),
		CreditType:   strings.TrimSpace(r.Get("creditType").String()),
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(list)
}

func PlatformCreditLedgerHandler(r *ghttp.Request) {
	list, err := ListLedger(LedgerFilter{
		EnterpriseId: strings.TrimSpace(r.Get("enterpriseId").String()),
		CreditType:   strings.TrimSpace(r.Get("creditType").String()),
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(list)
}

func EnterpriseCreditLedgerHandler(r *ghttp.Request) {
	list, err := ListLedger(LedgerFilter{
		EnterpriseId: identity.GetCurrentEnterpriseId(r),
		CreditType:   strings.TrimSpace(r.Get("creditType").String()),
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(list)
}

func PlatformCreditGrantsHandler(r *ghttp.Request) {
	list, err := ListGrants(GrantFilter{
		EnterpriseId: strings.TrimSpace(r.Get("enterpriseId").String()),
		CreditType:   strings.TrimSpace(r.Get("creditType").String()),
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(list)
}

func EnterpriseCreditGrantsHandler(r *ghttp.Request) {
	list, err := ListGrants(GrantFilter{
		EnterpriseId: identity.GetCurrentEnterpriseId(r),
		CreditType:   strings.TrimSpace(r.Get("creditType").String()),
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(list)
}

func CreatePlatformCreditGrantHandler(r *ghttp.Request) {
	var req grantReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	createGrantHandler(r, strings.TrimSpace(req.EnterpriseId), strings.TrimSpace(req.CreditType), req)
}

func CreateEnterpriseCreditGrantHandler(r *ghttp.Request) {
	var req grantReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	creditType := strings.TrimSpace(req.CreditType)
	if creditType == "" {
		creditType = CreditTypeEnterprise
	}
	createGrantHandler(r, identity.GetCurrentEnterpriseId(r), creditType, req)
}

func createGrantHandler(r *ghttp.Request, enterpriseId, creditType string, req grantReq) {
	item, wallet, err := CreateGrant(enterpriseId, creditType, CreateGrantReq{
		SourceType:   strings.TrimSpace(req.SourceType),
		SourceRefId:  strings.TrimSpace(req.SourceRefId),
		Credits:      req.Credits,
		EffectiveAt:  parseTimePtr(req.EffectiveAt),
		ExpiresAt:    parseTimePtr(req.ExpiresAt),
		MetadataJson: req.MetadataJson,
		ReasonCode:   strings.TrimSpace(req.ReasonCode),
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(g.Map{
		"grant":  item,
		"wallet": wallet,
	})
}

func ListPlatformCreditPriceBooksHandler(r *ghttp.Request) {
	list, err := ListPriceBooks(PriceBookFilter{
		EnterpriseId: strings.TrimSpace(r.Get("enterpriseId").String()),
		CreditType:   strings.TrimSpace(r.Get("creditType").String()),
		ModelId:      strings.TrimSpace(r.Get("modelId").String()),
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(list)
}

func ListEnterpriseCreditPriceBooksHandler(r *ghttp.Request) {
	list, err := ListPriceBooks(PriceBookFilter{
		EnterpriseId: identity.GetCurrentEnterpriseId(r),
		CreditType:   strings.TrimSpace(r.Get("creditType").String()),
		ModelId:      strings.TrimSpace(r.Get("modelId").String()),
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(list)
}

func CreatePlatformCreditPriceBookHandler(r *ghttp.Request) {
	var req priceBookReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	createPriceBookHandler(r, strings.TrimSpace(req.EnterpriseId), req)
}

func CreateEnterpriseCreditPriceBookHandler(r *ghttp.Request) {
	var req priceBookReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	createPriceBookHandler(r, identity.GetCurrentEnterpriseId(r), req)
}

func createPriceBookHandler(r *ghttp.Request, enterpriseId string, req priceBookReq) {
	item, err := CreatePriceBook(enterpriseId, CreatePriceBookReq{
		CreditType:           req.CreditType,
		ModelId:              req.ModelId,
		ModelScope:           req.ModelScope,
		ModelSourceType:      req.ModelSourceType,
		FundingType:          req.FundingType,
		Currency:             req.Currency,
		CreditUnitUsd:        req.CreditUnitUsd,
		CostInputUsdPer1M:    req.CostInputUsdPer1M,
		CostOutputUsdPer1M:   req.CostOutputUsdPer1M,
		PlatformMultiplier:   req.PlatformMultiplier,
		EnterpriseMultiplier: req.EnterpriseMultiplier,
		EffectiveAt:          parseTime(req.EffectiveAt),
		Status:               req.Status,
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(item)
}

func UpdatePlatformCreditPriceBookHandler(r *ghttp.Request) {
	updatePriceBookHandler(r, strings.TrimSpace(r.Get("id").String()), "")
}

func UpdateEnterpriseCreditPriceBookHandler(r *ghttp.Request) {
	updatePriceBookHandler(r, strings.TrimSpace(r.Get("id").String()), identity.GetCurrentEnterpriseId(r))
}

func updatePriceBookHandler(r *ghttp.Request, id, fallbackEnterpriseId string) {
	var req priceBookReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	enterpriseId := fallbackEnterpriseId
	if enterpriseId == "" {
		enterpriseId = strings.TrimSpace(req.EnterpriseId)
	}
	item, err := UpdatePriceBook(id, enterpriseId, UpdatePriceBookReq{
		CreditType:           req.CreditType,
		ModelId:              req.ModelId,
		ModelScope:           req.ModelScope,
		ModelSourceType:      req.ModelSourceType,
		FundingType:          req.FundingType,
		Currency:             req.Currency,
		CreditUnitUsd:        req.CreditUnitUsd,
		CostInputUsdPer1M:    req.CostInputUsdPer1M,
		CostOutputUsdPer1M:   req.CostOutputUsdPer1M,
		PlatformMultiplier:   req.PlatformMultiplier,
		EnterpriseMultiplier: req.EnterpriseMultiplier,
		EffectiveAt:          parseTime(req.EffectiveAt),
		Status:               req.Status,
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(item)
}

func DeletePlatformCreditPriceBookHandler(r *ghttp.Request) {
	id := strings.TrimSpace(r.Get("id").String())
	if err := DeletePriceBook(id, strings.TrimSpace(r.Get("enterpriseId").String())); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(g.Map{"message": "platform credit price book deleted"})
}

func DeleteEnterpriseCreditPriceBookHandler(r *ghttp.Request) {
	id := strings.TrimSpace(r.Get("id").String())
	if err := DeletePriceBook(id, identity.GetCurrentEnterpriseId(r)); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(g.Map{"message": "enterprise credit price book deleted"})
}

func ListEnterpriseCreditBudgetPoliciesHandler(r *ghttp.Request) {
	list, err := ListBudgetPolicies(BudgetPolicyFilter{
		EnterpriseId: identity.GetCurrentEnterpriseId(r),
		CreditType:   strings.TrimSpace(r.Get("creditType").String()),
		ScopeType:    strings.TrimSpace(r.Get("scopeType").String()),
		ScopeId:      strings.TrimSpace(r.Get("scopeId").String()),
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(list)
}

func CreateEnterpriseCreditBudgetPolicyHandler(r *ghttp.Request) {
	var req budgetPolicyReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	item, err := CreateBudgetPolicy(identity.GetCurrentEnterpriseId(r), CreateBudgetPolicyReq{
		CreditType:         req.CreditType,
		ScopeType:          req.ScopeType,
		ScopeId:            req.ScopeId,
		Enabled:            req.Enabled,
		DailyCreditLimit:   req.DailyCreditLimit,
		MonthlyCreditLimit: req.MonthlyCreditLimit,
		DailyTokenLimit:    req.DailyTokenLimit,
		MonthlyTokenLimit:  req.MonthlyTokenLimit,
		DailyUsdLimit:      req.DailyUsdLimit,
		MonthlyUsdLimit:    req.MonthlyUsdLimit,
		HardLimit:          req.HardLimit,
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(item)
}

func UpdateEnterpriseCreditBudgetPolicyHandler(r *ghttp.Request) {
	id := strings.TrimSpace(r.Get("id").String())
	var req budgetPolicyReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	item, err := UpdateBudgetPolicy(id, identity.GetCurrentEnterpriseId(r), UpdateBudgetPolicyReq{
		CreditType:         req.CreditType,
		ScopeType:          req.ScopeType,
		ScopeId:            req.ScopeId,
		Enabled:            req.Enabled,
		DailyCreditLimit:   req.DailyCreditLimit,
		MonthlyCreditLimit: req.MonthlyCreditLimit,
		DailyTokenLimit:    req.DailyTokenLimit,
		MonthlyTokenLimit:  req.MonthlyTokenLimit,
		DailyUsdLimit:      req.DailyUsdLimit,
		MonthlyUsdLimit:    req.MonthlyUsdLimit,
		HardLimit:          req.HardLimit,
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(item)
}

func DeleteEnterpriseCreditBudgetPolicyHandler(r *ghttp.Request) {
	id := strings.TrimSpace(r.Get("id").String())
	if err := DeleteBudgetPolicy(id, identity.GetCurrentEnterpriseId(r)); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(g.Map{"message": "enterprise credit budget policy deleted"})
}

func parseTime(value string) time.Time {
	parsed := parseTimePtr(value)
	if parsed == nil {
		return time.Time{}
	}
	return *parsed
}

func parseTimePtr(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}
	parsed = parsed.UTC()
	return &parsed
}
