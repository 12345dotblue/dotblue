package metering

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"dotblue/internal/domains/agent"
	"dotblue/internal/domains/identity"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type priceReq struct {
	ModelId               string  `json:"modelId"`
	Currency              string  `json:"currency"`
	CostInputUnitPrice    float64 `json:"costInputUnitPrice"`
	CostOutputUnitPrice   float64 `json:"costOutputUnitPrice"`
	ChargeInputUnitPrice  float64 `json:"chargeInputUnitPrice"`
	ChargeOutputUnitPrice float64 `json:"chargeOutputUnitPrice"`
}

type policyReq struct {
	ScopeType          string  `json:"scopeType"`
	ScopeId            string  `json:"scopeId"`
	Enabled            bool    `json:"enabled"`
	DailyTokenLimit    int64   `json:"dailyTokenLimit"`
	MonthlyTokenLimit  int64   `json:"monthlyTokenLimit"`
	DailyChargeLimit   float64 `json:"dailyChargeLimit"`
	MonthlyChargeLimit float64 `json:"monthlyChargeLimit"`
	HardLimit          bool    `json:"hardLimit"`
}

func PlatformUsageOverviewHandler(r *ghttp.Request) {
	item, err := GetOverview(ScopePlatform, "")
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, err.Error())
		return
	}
	r.Response.WriteJson(item)
}

func EnterpriseUsageOverviewHandler(r *ghttp.Request) {
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	item, err := GetOverview(ScopeEnterprise, enterpriseId)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, err.Error())
		return
	}
	r.Response.WriteJson(item)
}

func PlatformUsageTrendsHandler(r *ghttp.Request) {
	days := parseDays(r)
	list, err := GetTrends(ScopePlatform, "", days)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, err.Error())
		return
	}
	r.Response.WriteJson(list)
}

func EnterpriseUsageTrendsHandler(r *ghttp.Request) {
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	days := parseDays(r)
	list, err := GetTrends(ScopeEnterprise, enterpriseId, days)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, err.Error())
		return
	}
	r.Response.WriteJson(list)
}

func AgentUsageOverviewHandler(r *ghttp.Request) {
	agentItem, ok := currentAgentForUsage(r)
	if !ok {
		return
	}
	item, err := GetOverview(ScopeAgent, agentItem.Id)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, err.Error())
		return
	}
	r.Response.WriteJson(item)
}

func AgentUsageTrendsHandler(r *ghttp.Request) {
	agentItem, ok := currentAgentForUsage(r)
	if !ok {
		return
	}
	list, err := GetTrends(ScopeAgent, agentItem.Id, parseDays(r))
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, err.Error())
		return
	}
	r.Response.WriteJson(list)
}

func PlatformUsageEventsHandler(r *ghttp.Request) {
	listUsageEventsHandler(r, UsageEventFilter{})
}

func EnterpriseUsageEventsHandler(r *ghttp.Request) {
	listUsageEventsHandler(r, UsageEventFilter{
		EnterpriseId: identity.GetCurrentEnterpriseId(r),
	})
}

func listUsageEventsHandler(r *ghttp.Request, filter UsageEventFilter) {
	filter.AgentId = strings.TrimSpace(r.Get("agentId").String())
	filter.ModelId = strings.TrimSpace(r.Get("modelId").String())
	filter.Status = strings.TrimSpace(r.Get("status").String())
	filter.SourceType = strings.TrimSpace(r.Get("sourceType").String())
	filter.Page = r.Get("page").Int()
	filter.PageSize = r.Get("pageSize").Int()
	list, total, err := ListUsageEvents(filter)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, err.Error())
		return
	}
	r.Response.WriteJson(g.Map{
		"items": list,
		"total": total,
	})
}

func ListPlatformPricesHandler(r *ghttp.Request) {
	list, err := ListPrices(PriceFilter{ScopeType: ScopePlatform, ScopeId: ""})
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, err.Error())
		return
	}
	r.Response.WriteJson(list)
}

func CreatePlatformPriceHandler(r *ghttp.Request) {
	var req priceReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	item, err := CreatePrice(ScopePlatform, "", CreatePriceReq(req))
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(item)
}

func UpdatePlatformPriceHandler(r *ghttp.Request) {
	id := strings.TrimSpace(r.Get("id").String())
	var req priceReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	item, err := UpdatePrice(id, ScopePlatform, "", UpdatePriceReq{
		Currency:              req.Currency,
		CostInputUnitPrice:    req.CostInputUnitPrice,
		CostOutputUnitPrice:   req.CostOutputUnitPrice,
		ChargeInputUnitPrice:  req.ChargeInputUnitPrice,
		ChargeOutputUnitPrice: req.ChargeOutputUnitPrice,
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(item)
}

func DeletePlatformPriceHandler(r *ghttp.Request) {
	id := strings.TrimSpace(r.Get("id").String())
	if err := DeletePrice(id, ScopePlatform, ""); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(g.Map{"message": "platform model price deleted"})
}

func ListEnterprisePricesHandler(r *ghttp.Request) {
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	list, err := ListPrices(PriceFilter{ScopeType: ScopeEnterprise, ScopeId: enterpriseId})
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, err.Error())
		return
	}
	r.Response.WriteJson(list)
}

func CreateEnterprisePriceHandler(r *ghttp.Request) {
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	var req priceReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	item, err := CreatePrice(ScopeEnterprise, enterpriseId, CreatePriceReq(req))
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(item)
}

func UpdateEnterprisePriceHandler(r *ghttp.Request) {
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	id := strings.TrimSpace(r.Get("id").String())
	var req priceReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	item, err := UpdatePrice(id, ScopeEnterprise, enterpriseId, UpdatePriceReq{
		Currency:              req.Currency,
		CostInputUnitPrice:    req.CostInputUnitPrice,
		CostOutputUnitPrice:   req.CostOutputUnitPrice,
		ChargeInputUnitPrice:  req.ChargeInputUnitPrice,
		ChargeOutputUnitPrice: req.ChargeOutputUnitPrice,
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(item)
}

func DeleteEnterprisePriceHandler(r *ghttp.Request) {
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	id := strings.TrimSpace(r.Get("id").String())
	if err := DeletePrice(id, ScopeEnterprise, enterpriseId); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(g.Map{"message": "enterprise model price deleted"})
}

func ListPlatformPoliciesHandler(r *ghttp.Request) {
	list, err := ListPolicies(ScopePlatform, "")
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, err.Error())
		return
	}
	r.Response.WriteJson(list)
}

func CreatePlatformPolicyHandler(r *ghttp.Request) {
	var req policyReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	item, err := CreateLimitPolicy(ScopePlatform, "", CreateLimitPolicyReq{
		Enabled:            req.Enabled,
		DailyTokenLimit:    req.DailyTokenLimit,
		MonthlyTokenLimit:  req.MonthlyTokenLimit,
		DailyChargeLimit:   req.DailyChargeLimit,
		MonthlyChargeLimit: req.MonthlyChargeLimit,
		HardLimit:          req.HardLimit,
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(item)
}

func UpdatePlatformPolicyHandler(r *ghttp.Request) {
	id := strings.TrimSpace(r.Get("id").String())
	var req policyReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	item, err := UpdateLimitPolicy(id, ScopePlatform, "", CreateLimitPolicyReq{
		Enabled:            req.Enabled,
		DailyTokenLimit:    req.DailyTokenLimit,
		MonthlyTokenLimit:  req.MonthlyTokenLimit,
		DailyChargeLimit:   req.DailyChargeLimit,
		MonthlyChargeLimit: req.MonthlyChargeLimit,
		HardLimit:          req.HardLimit,
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(item)
}

func DeletePlatformPolicyHandler(r *ghttp.Request) {
	id := strings.TrimSpace(r.Get("id").String())
	if err := DeleteLimitPolicy(id, ScopePlatform, ""); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(g.Map{"message": "platform limit policy deleted"})
}

func ListEnterprisePoliciesHandler(r *ghttp.Request) {
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	scopeType := strings.TrimSpace(r.Get("scopeType").String())
	if scopeType == "" {
		scopeType = ScopeEnterprise
	}
	scopeId := enterpriseId
	if scopeType == ScopeAgent {
		scopeId = ""
	}
	list, err := ListPolicies(scopeType, scopeId)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, err.Error())
		return
	}
	if scopeType == ScopeAgent {
		filtered := make([]LimitPolicy, 0, len(list))
		for i := range list {
			if belongsToEnterpriseAgent(enterpriseId, list[i].ScopeId) {
				filtered = append(filtered, list[i])
			}
		}
		r.Response.WriteJson(filtered)
		return
	}
	r.Response.WriteJson(list)
}

func CreateEnterprisePolicyHandler(r *ghttp.Request) {
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	var req policyReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	scopeType, scopeId, err := resolveEnterprisePolicyScope(enterpriseId, req.ScopeType, req.ScopeId)
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	item, err := CreateLimitPolicy(scopeType, scopeId, CreateLimitPolicyReq{
		Enabled:            req.Enabled,
		DailyTokenLimit:    req.DailyTokenLimit,
		MonthlyTokenLimit:  req.MonthlyTokenLimit,
		DailyChargeLimit:   req.DailyChargeLimit,
		MonthlyChargeLimit: req.MonthlyChargeLimit,
		HardLimit:          req.HardLimit,
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(item)
}

func UpdateEnterprisePolicyHandler(r *ghttp.Request) {
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	id := strings.TrimSpace(r.Get("id").String())
	var req policyReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	scopeType, scopeId, err := resolveEnterprisePolicyScope(enterpriseId, req.ScopeType, req.ScopeId)
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	item, err := UpdateLimitPolicy(id, scopeType, scopeId, CreateLimitPolicyReq{
		Enabled:            req.Enabled,
		DailyTokenLimit:    req.DailyTokenLimit,
		MonthlyTokenLimit:  req.MonthlyTokenLimit,
		DailyChargeLimit:   req.DailyChargeLimit,
		MonthlyChargeLimit: req.MonthlyChargeLimit,
		HardLimit:          req.HardLimit,
	})
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(item)
}

func DeleteEnterprisePolicyHandler(r *ghttp.Request) {
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	id := strings.TrimSpace(r.Get("id").String())
	scopeType := strings.TrimSpace(r.Get("scopeType").String())
	scopeId := strings.TrimSpace(r.Get("scopeId").String())
	resolvedScopeType, resolvedScopeId, err := resolveEnterprisePolicyScope(enterpriseId, scopeType, scopeId)
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	if err := DeleteLimitPolicy(id, resolvedScopeType, resolvedScopeId); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}
	r.Response.WriteJson(g.Map{"message": "enterprise limit policy deleted"})
}

func resolveEnterprisePolicyScope(enterpriseId, scopeType, scopeId string) (string, string, error) {
	scopeType = strings.TrimSpace(scopeType)
	switch scopeType {
	case "", ScopeEnterprise:
		return ScopeEnterprise, enterpriseId, nil
	case ScopeAgent:
		scopeId = strings.TrimSpace(scopeId)
		if scopeId == "" {
			return "", "", gerror("agent scope id is required")
		}
		if !belongsToEnterpriseAgent(enterpriseId, scopeId) {
			return "", "", gerror("agent not found")
		}
		return ScopeAgent, scopeId, nil
	default:
		return "", "", gerror("unsupported enterprise policy scope")
	}
}

func belongsToEnterpriseAgent(enterpriseId, agentId string) bool {
	item, err := agent.GetById(agentId)
	if err != nil || item == nil {
		return false
	}
	return strings.TrimSpace(item.GroupId) == strings.TrimSpace(enterpriseId)
}

func currentAgentForUsage(r *ghttp.Request) (*agent.Agent, bool) {
	userId := identity.GetUserId(r)
	enterpriseId := identity.GetCurrentEnterpriseId(r)
	agentId := strings.TrimSpace(r.Get("id").String())
	if agentId == "" {
		r.Response.WriteStatus(http.StatusBadRequest, "Agent ID is required")
		return nil, false
	}
	ok, err := agent.BelongsToUser(agentId, userId, enterpriseId)
	if err != nil || !ok {
		r.Response.WriteStatus(http.StatusNotFound, "Agent not found")
		return nil, false
	}
	item, err := agent.GetById(agentId)
	if err != nil || item == nil {
		r.Response.WriteStatus(http.StatusNotFound, "Agent not found")
		return nil, false
	}
	return item, true
}

func parseDays(r *ghttp.Request) int {
	if v := strings.TrimSpace(r.Get("days").String()); v != "" {
		if days, err := strconv.Atoi(v); err == nil && days > 0 {
			return days
		}
	}
	return 7
}

func gerror(msg string) error {
	return errors.New(msg)
}
