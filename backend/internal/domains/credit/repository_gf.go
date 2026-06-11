package credit

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

func isNotFoundErr(err error) bool {
	return err != nil && (errors.Is(err, sql.ErrNoRows) || strings.Contains(strings.ToLower(err.Error()), "no rows in result set"))
}

type GFRepository struct{}

func NewGFRepository() *GFRepository {
	return &GFRepository{}
}

func (r *GFRepository) GetWallet(enterpriseId, creditType string) (*CreditWallet, error) {
	var item CreditWallet
	err := g.DB().Model("credit_wallets").
		Where("enterprise_id = ? AND credit_type = ?", enterpriseId, creditType).
		Limit(1).
		Scan(&item)
	if err != nil {
		if isNotFoundErr(err) {
			return nil, nil
		}
		return nil, err
	}
	if item.Id == "" {
		return nil, nil
	}
	return &item, nil
}

func getWalletByModel(model *gdb.Model, enterpriseId, creditType string) (*CreditWallet, error) {
	var item CreditWallet
	err := model.
		Where("enterprise_id = ? AND credit_type = ?", enterpriseId, creditType).
		Limit(1).
		Scan(&item)
	if err != nil {
		if isNotFoundErr(err) {
			return nil, nil
		}
		return nil, err
	}
	if item.Id == "" {
		return nil, nil
	}
	return &item, nil
}

func getWalletByIDModel(model *gdb.Model, id string) (*CreditWallet, error) {
	var item CreditWallet
	err := model.
		Where("id = ?", id).
		Limit(1).
		Scan(&item)
	if err != nil {
		if isNotFoundErr(err) {
			return nil, nil
		}
		return nil, err
	}
	if item.Id == "" {
		return nil, nil
	}
	return &item, nil
}

func (r *GFRepository) ListWallets(filter WalletFilter) ([]CreditWallet, error) {
	model := g.DB().Model("credit_wallets")
	if filter.EnterpriseId != "" {
		model = model.Where("enterprise_id = ?", filter.EnterpriseId)
	}
	if filter.CreditType != "" {
		model = model.Where("credit_type = ?", filter.CreditType)
	}
	var list []CreditWallet
	err := model.Order("credit_type ASC").Scan(&list)
	return list, err
}

func (r *GFRepository) InsertWallet(item *CreditWallet) error {
	_, err := g.DB().Model("credit_wallets").Data(g.Map{
		"id":                item.Id,
		"enterprise_id":     item.EnterpriseId,
		"credit_type":       item.CreditType,
		"total_credits":     item.TotalCredits,
		"reserved_credits":  item.ReservedCredits,
		"available_credits": item.AvailableCredits,
		"version":           item.Version,
		"created_at":        item.CreatedAt,
		"updated_at":        item.UpdatedAt,
	}).InsertIgnore()
	return err
}

func (r *GFRepository) GetGrantBySourceRef(enterpriseId, creditType, sourceType, sourceRefId string) (*CreditGrant, error) {
	var item CreditGrant
	err := g.DB().Model("credit_grants").
		Where("enterprise_id = ?", enterpriseId).
		Where("credit_type = ?", creditType).
		Where("source_type = ?", sourceType).
		Where("source_ref_id = ?", sourceRefId).
		Order("created_at DESC").
		Limit(1).
		Scan(&item)
	if err != nil {
		if isNotFoundErr(err) {
			return nil, nil
		}
		return nil, err
	}
	if item.Id == "" {
		return nil, nil
	}
	return &item, nil
}

func getGrantBySourceRefModel(model *gdb.Model, enterpriseId, creditType, sourceType, sourceRefId string) (*CreditGrant, error) {
	var item CreditGrant
	err := model.
		Where("enterprise_id = ?", enterpriseId).
		Where("credit_type = ?", creditType).
		Where("source_type = ?", sourceType).
		Where("source_ref_id = ?", sourceRefId).
		Order("created_at DESC").
		Limit(1).
		Scan(&item)
	if err != nil {
		if isNotFoundErr(err) {
			return nil, nil
		}
		return nil, err
	}
	if item.Id == "" {
		return nil, nil
	}
	return &item, nil
}

func grantData(grant *CreditGrant) g.Map {
	return g.Map{
		"id":                grant.Id,
		"enterprise_id":     grant.EnterpriseId,
		"credit_type":       grant.CreditType,
		"source_type":       grant.SourceType,
		"source_ref_id":     grant.SourceRefId,
		"granted_credits":   grant.GrantedCredits,
		"remaining_credits": grant.RemainingCredits,
		"effective_at":      grant.EffectiveAt,
		"expires_at":        grant.ExpiresAt,
		"metadata_json":     grant.MetadataJson,
		"created_at":        grant.CreatedAt,
	}
}

func (r *GFRepository) ApplyGrant(wallet *CreditWallet, grant *CreditGrant, entry *CreditLedgerEntry) (*CreditGrant, *CreditWallet, error) {
	var resolvedGrant *CreditGrant
	var resolvedWallet *CreditWallet
	err := g.DB().Transaction(context.Background(), func(ctx context.Context, tx gdb.TX) error {
		if strings.TrimSpace(grant.SourceRefId) != "" {
			// Insert the idempotency key first so concurrent bootstrap requests
			// cannot both mutate the wallet before one of them loses the race.
			result, err := tx.Model("credit_grants").Ctx(ctx).Data(grantData(grant)).InsertIgnore()
			if err != nil {
				return err
			}
			affected, err := result.RowsAffected()
			if err != nil {
				return err
			}
			if affected == 0 {
				resolvedGrant, err = getGrantBySourceRefModel(tx.Model("credit_grants").Ctx(ctx), grant.EnterpriseId, grant.CreditType, grant.SourceType, grant.SourceRefId)
				if err != nil {
					return err
				}
				resolvedWallet, err = getWalletByModel(tx.Model("credit_wallets").Ctx(ctx), wallet.EnterpriseId, wallet.CreditType)
				return err
			}
		} else {
			if _, err := tx.Model("credit_grants").Ctx(ctx).Data(grantData(grant)).Insert(); err != nil {
				return err
			}
		}
		if _, err := tx.Model("credit_wallets").Ctx(ctx).Data(g.Map{
			"total_credits":     wallet.TotalCredits,
			"available_credits": wallet.AvailableCredits,
			"reserved_credits":  wallet.ReservedCredits,
			"version":           wallet.Version,
			"updated_at":        wallet.UpdatedAt,
		}).Where("id = ?", wallet.Id).Update(); err != nil {
			return err
		}
		if _, err := tx.Model("credit_ledger_entries").Ctx(ctx).Data(g.Map{
			"id":                entry.Id,
			"enterprise_id":     entry.EnterpriseId,
			"credit_type":       entry.CreditType,
			"wallet_id":         entry.WalletId,
			"grant_id":          entry.GrantId,
			"entry_type":        entry.EntryType,
			"direction":         entry.Direction,
			"credits":           entry.Credits,
			"balance_after":     entry.BalanceAfter,
			"reserved_after":    entry.ReservedAfter,
			"invocation_id":     entry.InvocationId,
			"member_user_id":    entry.MemberUserId,
			"agent_id":          nullableUUIDValue(entry.AgentId),
			"budget_scope_type": entry.BudgetScopeType,
			"budget_scope_id":   entry.BudgetScopeId,
			"reason_code":       entry.ReasonCode,
			"snapshot_json":     entry.SnapshotJson,
			"created_at":        entry.CreatedAt,
		}).Insert(); err != nil {
			return err
		}
		resolvedGrant = grant
		resolvedWallet = wallet
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return resolvedGrant, resolvedWallet, nil
}

func (r *GFRepository) ListGrants(filter GrantFilter) ([]CreditGrant, error) {
	model := g.DB().Model("credit_grants")
	if filter.EnterpriseId != "" {
		model = model.Where("enterprise_id = ?", filter.EnterpriseId)
	}
	if filter.CreditType != "" {
		model = model.Where("credit_type = ?", filter.CreditType)
	}
	var list []CreditGrant
	err := model.Order("created_at DESC").Scan(&list)
	return list, err
}

func (r *GFRepository) ListLedger(filter LedgerFilter) ([]CreditLedgerEntry, error) {
	model := g.DB().Model("credit_ledger_entries")
	if filter.EnterpriseId != "" {
		model = model.Where("enterprise_id = ?", filter.EnterpriseId)
	}
	if filter.CreditType != "" {
		model = model.Where("credit_type = ?", filter.CreditType)
	}
	var list []CreditLedgerEntry
	err := model.Order("created_at DESC").Scan(&list)
	return list, err
}

func (r *GFRepository) AggregateLedgerByType(enterpriseId string, from, to time.Time) ([]LedgerAggregate, error) {
	model := g.DB().Model("credit_ledger_entries").Where("enterprise_id = ?", enterpriseId)
	if !from.IsZero() {
		model = model.Where("created_at >= ?", from)
	}
	if !to.IsZero() {
		model = model.Where("created_at < ?", to)
	}
	type row struct {
		CreditType     string `json:"creditType"`
		GrantedCredits int64  `json:"grantedCredits"`
		SettledCredits int64  `json:"settledCredits"`
		ExpiredCredits int64  `json:"expiredCredits"`
	}
	var rows []row
	err := model.Fields(`
		credit_type,
		COALESCE(SUM(CASE WHEN entry_type = 'grant' THEN credits ELSE 0 END), 0) AS granted_credits,
		COALESCE(SUM(CASE WHEN entry_type = 'settle' THEN credits ELSE 0 END), 0) AS settled_credits,
		COALESCE(SUM(CASE WHEN entry_type = 'expire' THEN credits ELSE 0 END), 0) AS expired_credits
	`).Group("credit_type").Order("credit_type ASC").Scan(&rows)
	if err != nil {
		return nil, err
	}
	result := make([]LedgerAggregate, 0, len(rows))
	for _, item := range rows {
		result = append(result, LedgerAggregate(item))
	}
	return result, nil
}

func (r *GFRepository) AggregateSettledCredits(enterpriseId, creditType, scopeType, scopeId string, from, to time.Time) (int64, error) {
	model := g.DB().Model("credit_ledger_entries").
		Where("enterprise_id = ?", enterpriseId).
		Where("credit_type = ?", creditType).
		Where("entry_type = ?", EntryTypeSettle)
	if !from.IsZero() {
		model = model.Where("created_at >= ?", from)
	}
	if !to.IsZero() {
		model = model.Where("created_at < ?", to)
	}
	switch strings.TrimSpace(scopeType) {
	case ScopeMember:
		model = model.Where("member_user_id = ?", scopeId)
	case ScopeAgent:
		model = model.Where("agent_id = ?", scopeId)
	}
	value, err := model.Fields("COALESCE(SUM(credits), 0)").Value()
	if err != nil {
		return 0, err
	}
	return value.Int64(), nil
}

func (r *GFRepository) GetPriceBookByID(id string) (*CreditPriceBook, error) {
	var item CreditPriceBook
	err := g.DB().Model("credit_price_books").Where("id = ?", id).Limit(1).Scan(&item)
	if err != nil {
		if isNotFoundErr(err) {
			return nil, nil
		}
		return nil, err
	}
	if item.Id == "" {
		return nil, nil
	}
	return &item, nil
}

func (r *GFRepository) ListPriceBooks(filter PriceBookFilter) ([]CreditPriceBook, error) {
	model := g.DB().Model("credit_price_books")
	if filter.EnterpriseId != "" {
		model = model.Where("enterprise_id = ?", filter.EnterpriseId)
	}
	if filter.CreditType != "" {
		model = model.Where("credit_type = ?", filter.CreditType)
	}
	if filter.ModelId != "" {
		model = model.Where("model_id = ?", filter.ModelId)
	}
	var list []CreditPriceBook
	err := model.Order("effective_at DESC, created_at DESC").Scan(&list)
	return list, err
}

func (r *GFRepository) InsertPriceBook(item *CreditPriceBook) error {
	_, err := g.DB().Model("credit_price_books").Data(priceBookData(item)).Insert()
	return err
}

func (r *GFRepository) UpdatePriceBook(item *CreditPriceBook) error {
	_, err := g.DB().Model("credit_price_books").Data(priceBookData(item)).Where("id = ?", item.Id).Update()
	return err
}

func (r *GFRepository) DeletePriceBook(id string) error {
	_, err := g.DB().Model("credit_price_books").Where("id = ?", id).Delete()
	return err
}

func (r *GFRepository) GetBudgetPolicyByID(id string) (*CreditBudgetPolicy, error) {
	var item CreditBudgetPolicy
	err := g.DB().Model("credit_budget_policies").Where("id = ?", id).Limit(1).Scan(&item)
	if err != nil {
		if isNotFoundErr(err) {
			return nil, nil
		}
		return nil, err
	}
	if item.Id == "" {
		return nil, nil
	}
	return &item, nil
}

func (r *GFRepository) ListBudgetPolicies(filter BudgetPolicyFilter) ([]CreditBudgetPolicy, error) {
	model := g.DB().Model("credit_budget_policies")
	if filter.EnterpriseId != "" {
		model = model.Where("enterprise_id = ?", filter.EnterpriseId)
	}
	if filter.CreditType != "" {
		model = model.Where("credit_type = ?", filter.CreditType)
	}
	if filter.ScopeType != "" {
		model = model.Where("scope_type = ?", filter.ScopeType)
	}
	if filter.ScopeId != "" {
		model = model.Where("scope_id = ?", filter.ScopeId)
	}
	var list []CreditBudgetPolicy
	err := model.Order("scope_type ASC, created_at DESC").Scan(&list)
	return list, err
}

func (r *GFRepository) InsertBudgetPolicy(item *CreditBudgetPolicy) error {
	_, err := g.DB().Model("credit_budget_policies").Data(budgetPolicyData(item)).Insert()
	return err
}

func (r *GFRepository) UpdateBudgetPolicy(item *CreditBudgetPolicy) error {
	_, err := g.DB().Model("credit_budget_policies").Data(budgetPolicyData(item)).Where("id = ?", item.Id).Update()
	return err
}

func (r *GFRepository) DeleteBudgetPolicy(id string) error {
	_, err := g.DB().Model("credit_budget_policies").Where("id = ?", id).Delete()
	return err
}

func (r *GFRepository) GetReservationByInvocationID(invocationId string) (*CreditReservation, error) {
	var item CreditReservation
	err := g.DB().Model("credit_reservations").Where("invocation_id = ?", invocationId).Limit(1).Scan(&item)
	if err != nil {
		if isNotFoundErr(err) {
			return nil, nil
		}
		return nil, err
	}
	if item.Id == "" {
		return nil, nil
	}
	return &item, nil
}

func (r *GFRepository) ApplyReservation(wallet *CreditWallet, reservation *CreditReservation, entry *CreditLedgerEntry) error {
	return g.DB().Transaction(context.Background(), func(ctx context.Context, tx gdb.TX) error {
		currentWallet, err := getWalletByIDModel(tx.Model("credit_wallets").Ctx(ctx).LockUpdate(), wallet.Id)
		if err != nil {
			return err
		}
		if currentWallet == nil {
			return sql.ErrNoRows
		}
		currentWallet.ReservedCredits += entry.Credits
		currentWallet.AvailableCredits -= entry.Credits
		currentWallet.Version++
		currentWallet.UpdatedAt = wallet.UpdatedAt
		entry.BalanceAfter = currentWallet.AvailableCredits
		entry.ReservedAfter = currentWallet.ReservedCredits

		if _, err := tx.Model("credit_wallets").Ctx(ctx).Data(g.Map{
			"reserved_credits":  currentWallet.ReservedCredits,
			"available_credits": currentWallet.AvailableCredits,
			"version":           currentWallet.Version,
			"updated_at":        currentWallet.UpdatedAt,
		}).Where("id = ?", wallet.Id).Update(); err != nil {
			return err
		}
		if _, err := tx.Model("credit_reservations").Ctx(ctx).Data(g.Map{
			"id":                  reservation.Id,
			"enterprise_id":       reservation.EnterpriseId,
			"credit_type":         reservation.CreditType,
			"invocation_id":       reservation.InvocationId,
			"member_user_id":      reservation.MemberUserId,
			"agent_id":            nullableUUIDValue(reservation.AgentId),
			"model_id":            reservation.ModelId,
			"model_scope":         reservation.ModelScope,
			"funding_type":        reservation.FundingType,
			"price_book_id":       reservation.PriceBookId,
			"price_snapshot_json": reservation.PriceSnapshotJson,
			"reserved_credits":    reservation.ReservedCredits,
			"status":              reservation.Status,
			"expires_at":          reservation.ExpiresAt,
			"created_at":          reservation.CreatedAt,
			"updated_at":          reservation.UpdatedAt,
		}).Insert(); err != nil {
			return err
		}
		if _, err := tx.Model("credit_ledger_entries").Ctx(ctx).Data(ledgerEntryData(entry)).Insert(); err != nil {
			return err
		}
		return nil
	})
}

func (r *GFRepository) ApplySettlement(wallet *CreditWallet, reservation *CreditReservation, settleEntry *CreditLedgerEntry, releaseEntry *CreditLedgerEntry) error {
	return g.DB().Transaction(context.Background(), func(ctx context.Context, tx gdb.TX) error {
		currentWallet, err := getWalletByIDModel(tx.Model("credit_wallets").Ctx(ctx).LockUpdate(), wallet.Id)
		if err != nil {
			return err
		}
		if currentWallet == nil {
			return sql.ErrNoRows
		}
		currentWallet.TotalCredits -= settleEntry.Credits
		currentWallet.ReservedCredits -= reservation.ReservedCredits
		currentWallet.AvailableCredits = currentWallet.TotalCredits - currentWallet.ReservedCredits
		currentWallet.Version++
		currentWallet.UpdatedAt = wallet.UpdatedAt
		settleEntry.BalanceAfter = currentWallet.AvailableCredits
		settleEntry.ReservedAfter = currentWallet.ReservedCredits
		if releaseEntry != nil {
			releaseEntry.BalanceAfter = currentWallet.AvailableCredits
			releaseEntry.ReservedAfter = currentWallet.ReservedCredits
		}

		if _, err := tx.Model("credit_wallets").Ctx(ctx).Data(g.Map{
			"total_credits":     currentWallet.TotalCredits,
			"reserved_credits":  currentWallet.ReservedCredits,
			"available_credits": currentWallet.AvailableCredits,
			"version":           currentWallet.Version,
			"updated_at":        currentWallet.UpdatedAt,
		}).Where("id = ?", wallet.Id).Update(); err != nil {
			return err
		}
		if _, err := tx.Model("credit_reservations").Ctx(ctx).Data(g.Map{
			"status":     reservation.Status,
			"updated_at": reservation.UpdatedAt,
			"expires_at": reservation.ExpiresAt,
		}).Where("id = ?", reservation.Id).Update(); err != nil {
			return err
		}
		if _, err := tx.Model("credit_ledger_entries").Ctx(ctx).Data(ledgerEntryData(settleEntry)).Insert(); err != nil {
			return err
		}
		if releaseEntry != nil {
			if _, err := tx.Model("credit_ledger_entries").Ctx(ctx).Data(ledgerEntryData(releaseEntry)).Insert(); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *GFRepository) ApplyRelease(wallet *CreditWallet, reservation *CreditReservation, entry *CreditLedgerEntry) error {
	return g.DB().Transaction(context.Background(), func(ctx context.Context, tx gdb.TX) error {
		currentWallet, err := getWalletByIDModel(tx.Model("credit_wallets").Ctx(ctx).LockUpdate(), wallet.Id)
		if err != nil {
			return err
		}
		if currentWallet == nil {
			return sql.ErrNoRows
		}
		currentWallet.ReservedCredits -= entry.Credits
		currentWallet.AvailableCredits = currentWallet.TotalCredits - currentWallet.ReservedCredits
		currentWallet.Version++
		currentWallet.UpdatedAt = wallet.UpdatedAt
		entry.BalanceAfter = currentWallet.AvailableCredits
		entry.ReservedAfter = currentWallet.ReservedCredits

		if _, err := tx.Model("credit_wallets").Ctx(ctx).Data(g.Map{
			"reserved_credits":  currentWallet.ReservedCredits,
			"available_credits": currentWallet.AvailableCredits,
			"version":           currentWallet.Version,
			"updated_at":        currentWallet.UpdatedAt,
		}).Where("id = ?", wallet.Id).Update(); err != nil {
			return err
		}
		if _, err := tx.Model("credit_reservations").Ctx(ctx).Data(g.Map{
			"status":     reservation.Status,
			"updated_at": reservation.UpdatedAt,
			"expires_at": reservation.ExpiresAt,
		}).Where("id = ?", reservation.Id).Update(); err != nil {
			return err
		}
		if _, err := tx.Model("credit_ledger_entries").Ctx(ctx).Data(ledgerEntryData(entry)).Insert(); err != nil {
			return err
		}
		return nil
	})
}

func priceBookData(item *CreditPriceBook) g.Map {
	return g.Map{
		"id":                         item.Id,
		"enterprise_id":              nullableString(item.EnterpriseId),
		"credit_type":                item.CreditType,
		"model_id":                   item.ModelId,
		"model_scope":                item.ModelScope,
		"model_source_type":          item.ModelSourceType,
		"funding_type":               item.FundingType,
		"currency":                   item.Currency,
		"credit_unit_usd":            item.CreditUnitUsd,
		"cost_input_usd_per_1m":      item.CostInputUsdPer1M,
		"cost_output_usd_per_1m":     item.CostOutputUsdPer1M,
		"platform_multiplier":        item.PlatformMultiplier,
		"enterprise_multiplier":      item.EnterpriseMultiplier,
		"billable_input_usd_per_1m":  item.BillableInputUsdPer1M,
		"billable_output_usd_per_1m": item.BillableOutputUsdPer1M,
		"input_credits_per_1m":       item.InputCreditsPer1M,
		"output_credits_per_1m":      item.OutputCreditsPer1M,
		"effective_at":               item.EffectiveAt,
		"status":                     item.Status,
		"created_at":                 item.CreatedAt,
		"updated_at":                 item.UpdatedAt,
	}
}

func budgetPolicyData(item *CreditBudgetPolicy) g.Map {
	return g.Map{
		"id":                   item.Id,
		"enterprise_id":        item.EnterpriseId,
		"credit_type":          item.CreditType,
		"scope_type":           item.ScopeType,
		"scope_id":             item.ScopeId,
		"enabled":              item.Enabled,
		"daily_credit_limit":   item.DailyCreditLimit,
		"monthly_credit_limit": item.MonthlyCreditLimit,
		"daily_token_limit":    item.DailyTokenLimit,
		"monthly_token_limit":  item.MonthlyTokenLimit,
		"daily_usd_limit":      item.DailyUsdLimit,
		"monthly_usd_limit":    item.MonthlyUsdLimit,
		"hard_limit":           item.HardLimit,
		"created_at":           item.CreatedAt,
		"updated_at":           item.UpdatedAt,
	}
}

func ledgerEntryData(entry *CreditLedgerEntry) g.Map {
	return g.Map{
		"id":                entry.Id,
		"enterprise_id":     entry.EnterpriseId,
		"credit_type":       entry.CreditType,
		"wallet_id":         entry.WalletId,
		"grant_id":          entry.GrantId,
		"entry_type":        entry.EntryType,
		"direction":         entry.Direction,
		"credits":           entry.Credits,
		"balance_after":     entry.BalanceAfter,
		"reserved_after":    entry.ReservedAfter,
		"invocation_id":     entry.InvocationId,
		"member_user_id":    entry.MemberUserId,
		"agent_id":          nullableUUIDValue(entry.AgentId),
		"budget_scope_type": entry.BudgetScopeType,
		"budget_scope_id":   entry.BudgetScopeId,
		"reason_code":       entry.ReasonCode,
		"snapshot_json":     entry.SnapshotJson,
		"created_at":        entry.CreatedAt,
	}
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return strings.TrimSpace(value)
}

func nullableUUIDValue(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.TrimSpace(value)
}
