package credit

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

type stubRepository struct {
	wallets                    map[string]*CreditWallet
	grants                     []CreditGrant
	ledger                     []CreditLedgerEntry
	priceBooks                 map[string]*CreditPriceBook
	budgetPolicies             map[string]*CreditBudgetPolicy
	reservations               map[string]*CreditReservation
	settledCredits             map[string]int64
	injectExistingGrantOnApply *CreditGrant
}

type stubUsageOverviewProvider struct {
	overview *usageOverview
}

func newStubRepository() *stubRepository {
	return &stubRepository{
		wallets:        map[string]*CreditWallet{},
		priceBooks:     map[string]*CreditPriceBook{},
		budgetPolicies: map[string]*CreditBudgetPolicy{},
		reservations:   map[string]*CreditReservation{},
		settledCredits: map[string]int64{},
	}
}

func walletKey(enterpriseId, creditType string) string {
	return enterpriseId + ":" + creditType
}

func settledKey(enterpriseId, creditType, scopeType, scopeId string) string {
	return enterpriseId + ":" + creditType + ":" + scopeType + ":" + scopeId
}

func (s *stubUsageOverviewProvider) GetOverview(scopeType, scopeId string) (*usageOverview, error) {
	if s.overview == nil {
		return &usageOverview{}, nil
	}
	copy := *s.overview
	return &copy, nil
}

func (r *stubRepository) GetWallet(enterpriseId, creditType string) (*CreditWallet, error) {
	item := r.wallets[walletKey(enterpriseId, creditType)]
	if item == nil {
		return nil, nil
	}
	copy := *item
	return &copy, nil
}

func (r *stubRepository) ListWallets(filter WalletFilter) ([]CreditWallet, error) {
	list := make([]CreditWallet, 0)
	for _, item := range r.wallets {
		if filter.EnterpriseId != "" && item.EnterpriseId != filter.EnterpriseId {
			continue
		}
		if filter.CreditType != "" && item.CreditType != filter.CreditType {
			continue
		}
		list = append(list, *item)
	}
	return list, nil
}

func (r *stubRepository) InsertWallet(item *CreditWallet) error {
	copy := *item
	r.wallets[walletKey(item.EnterpriseId, item.CreditType)] = &copy
	return nil
}

func (r *stubRepository) GetGrantBySourceRef(enterpriseId, creditType, sourceType, sourceRefId string) (*CreditGrant, error) {
	for idx := range r.grants {
		item := r.grants[idx]
		if item.EnterpriseId == enterpriseId && item.CreditType == creditType && item.SourceType == sourceType && item.SourceRefId == sourceRefId {
			copy := item
			return &copy, nil
		}
	}
	return nil, nil
}

func (r *stubRepository) ApplyGrant(wallet *CreditWallet, grant *CreditGrant, entry *CreditLedgerEntry) (*CreditGrant, *CreditWallet, error) {
	if r.injectExistingGrantOnApply != nil {
		copy := *r.injectExistingGrantOnApply
		currentWallet := r.wallets[walletKey(wallet.EnterpriseId, wallet.CreditType)]
		if currentWallet == nil {
			currentWallet = &CreditWallet{
				Id:               "wallet-existing",
				EnterpriseId:     wallet.EnterpriseId,
				CreditType:       wallet.CreditType,
				TotalCredits:     0,
				ReservedCredits:  0,
				AvailableCredits: 0,
				Version:          1,
			}
		}
		walletCopy := *currentWallet
		return &copy, &walletCopy, nil
	}
	walletCopy := *wallet
	grantCopy := *grant
	entryCopy := *entry
	r.wallets[walletKey(wallet.EnterpriseId, wallet.CreditType)] = &walletCopy
	r.grants = append(r.grants, grantCopy)
	r.ledger = append(r.ledger, entryCopy)
	return &grantCopy, &walletCopy, nil
}

func (r *stubRepository) ListGrants(filter GrantFilter) ([]CreditGrant, error) {
	return r.grants, nil
}

func (r *stubRepository) ListLedger(filter LedgerFilter) ([]CreditLedgerEntry, error) {
	return r.ledger, nil
}

func (r *stubRepository) AggregateLedgerByType(enterpriseId string, from, to time.Time) ([]LedgerAggregate, error) {
	return nil, nil
}

func (r *stubRepository) AggregateSettledCredits(enterpriseId, creditType, scopeType, scopeId string, from, to time.Time) (int64, error) {
	return r.settledCredits[settledKey(enterpriseId, creditType, scopeType, scopeId)], nil
}

func (r *stubRepository) GetPriceBookByID(id string) (*CreditPriceBook, error) {
	item := r.priceBooks[id]
	if item == nil {
		return nil, nil
	}
	copy := *item
	return &copy, nil
}

func (r *stubRepository) ListPriceBooks(filter PriceBookFilter) ([]CreditPriceBook, error) {
	list := make([]CreditPriceBook, 0)
	for _, item := range r.priceBooks {
		list = append(list, *item)
	}
	return list, nil
}

func (r *stubRepository) InsertPriceBook(item *CreditPriceBook) error {
	copy := *item
	r.priceBooks[item.Id] = &copy
	return nil
}

func (r *stubRepository) UpdatePriceBook(item *CreditPriceBook) error {
	copy := *item
	r.priceBooks[item.Id] = &copy
	return nil
}

func (r *stubRepository) DeletePriceBook(id string) error {
	delete(r.priceBooks, id)
	return nil
}

func (r *stubRepository) GetBudgetPolicyByID(id string) (*CreditBudgetPolicy, error) {
	item := r.budgetPolicies[id]
	if item == nil {
		return nil, nil
	}
	copy := *item
	return &copy, nil
}

func (r *stubRepository) ListBudgetPolicies(filter BudgetPolicyFilter) ([]CreditBudgetPolicy, error) {
	list := make([]CreditBudgetPolicy, 0)
	for _, item := range r.budgetPolicies {
		if filter.EnterpriseId != "" && item.EnterpriseId != filter.EnterpriseId {
			continue
		}
		if filter.CreditType != "" && item.CreditType != filter.CreditType {
			continue
		}
		if filter.ScopeType != "" && item.ScopeType != filter.ScopeType {
			continue
		}
		if filter.ScopeId != "" && item.ScopeId != filter.ScopeId {
			continue
		}
		list = append(list, *item)
	}
	return list, nil
}

func (r *stubRepository) InsertBudgetPolicy(item *CreditBudgetPolicy) error {
	copy := *item
	r.budgetPolicies[item.Id] = &copy
	return nil
}

func (r *stubRepository) UpdateBudgetPolicy(item *CreditBudgetPolicy) error {
	copy := *item
	r.budgetPolicies[item.Id] = &copy
	return nil
}

func (r *stubRepository) DeleteBudgetPolicy(id string) error {
	delete(r.budgetPolicies, id)
	return nil
}

func (r *stubRepository) GetReservationByInvocationID(invocationId string) (*CreditReservation, error) {
	item := r.reservations[invocationId]
	if item == nil {
		return nil, nil
	}
	copy := *item
	return &copy, nil
}

func (r *stubRepository) ApplyReservation(wallet *CreditWallet, reservation *CreditReservation, entry *CreditLedgerEntry) error {
	walletCopy := *wallet
	reservationCopy := *reservation
	entryCopy := *entry
	r.wallets[walletKey(wallet.EnterpriseId, wallet.CreditType)] = &walletCopy
	r.reservations[reservation.InvocationId] = &reservationCopy
	r.ledger = append(r.ledger, entryCopy)
	return nil
}

func (r *stubRepository) ApplySettlement(wallet *CreditWallet, reservation *CreditReservation, settleEntry *CreditLedgerEntry, releaseEntry *CreditLedgerEntry) error {
	walletCopy := *wallet
	reservationCopy := *reservation
	r.wallets[walletKey(wallet.EnterpriseId, wallet.CreditType)] = &walletCopy
	r.reservations[reservation.InvocationId] = &reservationCopy
	r.ledger = append(r.ledger, *settleEntry)
	if releaseEntry != nil {
		r.ledger = append(r.ledger, *releaseEntry)
	}
	return nil
}

func (r *stubRepository) ApplyRelease(wallet *CreditWallet, reservation *CreditReservation, entry *CreditLedgerEntry) error {
	walletCopy := *wallet
	reservationCopy := *reservation
	entryCopy := *entry
	r.wallets[walletKey(wallet.EnterpriseId, wallet.CreditType)] = &walletCopy
	r.reservations[reservation.InvocationId] = &reservationCopy
	r.ledger = append(r.ledger, entryCopy)
	return nil
}

func TestCreateGrantUpdatesWallet(t *testing.T) {
	repo := newStubRepository()
	svc := NewService(repo)
	svc.usageProvider = &stubUsageOverviewProvider{}
	svc.now = func() time.Time { return time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC) }
	svc.idGenerator = func() string { return "fixed-id" }

	grant, wallet, err := svc.CreateGrant("ent-1", CreditTypePlatform, CreateGrantReq{
		SourceType: "trial_grant",
		Credits:    120,
	})
	if err != nil {
		t.Fatalf("CreateGrant returned error: %v", err)
	}
	if grant == nil || wallet == nil {
		t.Fatalf("CreateGrant returned nil result")
	}
	if wallet.TotalCredits != 120 || wallet.AvailableCredits != 120 {
		t.Fatalf("unexpected wallet credits: %+v", wallet)
	}
	if len(repo.ledger) != 1 || repo.ledger[0].EntryType != EntryTypeGrant {
		t.Fatalf("expected grant ledger entry, got %+v", repo.ledger)
	}
}

func TestCreateGrantReturnsExistingGrantWhenApplyHitsIdempotencyRace(t *testing.T) {
	repo := newStubRepository()
	repo.wallets[walletKey("ent-1", CreditTypePlatform)] = &CreditWallet{
		Id:               "wallet-1",
		EnterpriseId:     "ent-1",
		CreditType:       CreditTypePlatform,
		TotalCredits:     123,
		ReservedCredits:  0,
		AvailableCredits: 123,
		Version:          2,
	}
	repo.injectExistingGrantOnApply = &CreditGrant{
		Id:               "grant-existing",
		EnterpriseId:     "ent-1",
		CreditType:       CreditTypePlatform,
		SourceType:       "enterprise_bootstrap",
		SourceRefId:      "ent-1",
		GrantedCredits:   123,
		RemainingCredits: 123,
	}
	svc := NewService(repo)
	svc.usageProvider = &stubUsageOverviewProvider{}
	svc.now = func() time.Time { return time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC) }
	svc.idGenerator = func() string { return "fixed-id" }

	grant, wallet, err := svc.CreateGrant("ent-1", CreditTypePlatform, CreateGrantReq{
		SourceType:  "enterprise_bootstrap",
		SourceRefId: "ent-1",
		Credits:     123,
	})
	if err != nil {
		t.Fatalf("CreateGrant returned error: %v", err)
	}
	if grant == nil || grant.Id != "grant-existing" {
		t.Fatalf("expected existing grant, got %+v", grant)
	}
	if wallet == nil || wallet.TotalCredits != 123 || wallet.AvailableCredits != 123 {
		t.Fatalf("expected existing wallet balance, got %+v", wallet)
	}
	if len(repo.ledger) != 0 {
		t.Fatalf("expected no new ledger entry, got %+v", repo.ledger)
	}
}

func TestCreatePriceBookBuildsCreditsFromBillableUSD(t *testing.T) {
	repo := newStubRepository()
	svc := NewService(repo)
	svc.usageProvider = &stubUsageOverviewProvider{}
	svc.now = func() time.Time { return time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC) }
	svc.idGenerator = func() string { return "pb-1" }

	item, err := svc.CreatePriceBook("", CreatePriceBookReq{
		CreditType:           CreditTypePlatform,
		ModelId:              "model-1",
		ModelScope:           "platform",
		ModelSourceType:      ModelSourceTypePlatform,
		FundingType:          FundingTypePlatform,
		CostInputUsdPer1M:    1,
		CostOutputUsdPer1M:   2,
		PlatformMultiplier:   1.5,
		EnterpriseMultiplier: 2,
	})
	if err != nil {
		t.Fatalf("CreatePriceBook returned error: %v", err)
	}
	if item.BillableInputUsdPer1M != 3 {
		t.Fatalf("unexpected input billable usd: %+v", item)
	}
	if item.InputCreditsPer1M != 30000 {
		t.Fatalf("unexpected input credits per 1m: %+v", item)
	}
	if item.OutputCreditsPer1M != 60000 {
		t.Fatalf("unexpected output credits per 1m: %+v", item)
	}
}

func TestCreatePriceBookSupportsCNY(t *testing.T) {
	repo := newStubRepository()
	svc := NewService(repo)
	svc.usageProvider = &stubUsageOverviewProvider{}
	svc.now = func() time.Time { return time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC) }
	svc.idGenerator = func() string { return "pb-cny-1" }

	item, err := svc.CreatePriceBook("", CreatePriceBookReq{
		CreditType:           CreditTypePlatform,
		ModelId:              "model-1",
		ModelScope:           "platform",
		ModelSourceType:      ModelSourceTypePlatform,
		FundingType:          FundingTypePlatform,
		Currency:             "CNY",
		CostInputUsdPer1M:    1,
		CostOutputUsdPer1M:   2,
		PlatformMultiplier:   1,
		EnterpriseMultiplier: 1,
	})
	if err != nil {
		t.Fatalf("CreatePriceBook returned error: %v", err)
	}
	if item.Currency != "CNY" {
		t.Fatalf("expected CNY currency, got %+v", item)
	}
}

func TestReserveAndSettleCredits(t *testing.T) {
	repo := newStubRepository()
	svc := NewService(repo)
	svc.usageProvider = &stubUsageOverviewProvider{}
	svc.now = func() time.Time { return time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC) }
	idCounter := 0
	svc.idGenerator = func() string {
		idCounter++
		return fmt.Sprintf("id-%d", idCounter)
	}
	repo.wallets[walletKey("ent-1", CreditTypePlatform)] = &CreditWallet{
		Id:               "wallet-1",
		EnterpriseId:     "ent-1",
		CreditType:       CreditTypePlatform,
		TotalCredits:     1000,
		ReservedCredits:  0,
		AvailableCredits: 1000,
		Version:          1,
		CreatedAt:        svc.now(),
		UpdatedAt:        svc.now(),
	}
	repo.priceBooks["pb-1"] = &CreditPriceBook{
		Id:                 "pb-1",
		CreditType:         CreditTypePlatform,
		ModelId:            "model-1",
		ModelScope:         "platform",
		ModelSourceType:    ModelSourceTypePlatform,
		FundingType:        FundingTypePlatform,
		CreditUnitUsd:      DefaultCreditUnitUSD,
		InputCreditsPer1M:  1000,
		OutputCreditsPer1M: 2000,
		EffectiveAt:        svc.now().Add(-time.Hour),
		Status:             PriceBookStatusActive,
	}

	reservation, snapshot, err := svc.Reserve(ReserveInput{
		InvocationId:              "inv-1",
		EnterpriseId:              "ent-1",
		MemberUserId:              "user-1",
		AgentId:                   "agent-1",
		ModelId:                   "model-1",
		ModelScope:                "platform",
		EstimatedPromptTokens:     1000,
		EstimatedCompletionTokens: 1000,
	})
	if err != nil {
		t.Fatalf("Reserve returned error: %v", err)
	}
	if reservation == nil || snapshot == nil {
		t.Fatalf("Reserve returned nil result")
	}
	settled, settledSnapshot, err := svc.Settle(SettleInput{
		InvocationId:     "inv-1",
		PromptTokens:     1000,
		CompletionTokens: 500,
	})
	if err != nil {
		t.Fatalf("Settle returned error: %v", err)
	}
	if settled == nil || settledSnapshot == nil {
		t.Fatalf("Settle returned nil result")
	}
	if settledSnapshot.SettledCredits <= 0 {
		t.Fatalf("expected settled credits, got %+v", settledSnapshot)
	}
}

func TestReserveReturnsInsufficientCreditsWhenWalletBalanceIsTooLow(t *testing.T) {
	repo := newStubRepository()
	svc := NewService(repo)
	svc.usageProvider = &stubUsageOverviewProvider{}
	svc.now = func() time.Time { return time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC) }
	svc.idGenerator = func() string { return "reserve-low-balance" }
	repo.wallets[walletKey("ent-1", CreditTypePlatform)] = &CreditWallet{
		Id:               "wallet-1",
		EnterpriseId:     "ent-1",
		CreditType:       CreditTypePlatform,
		TotalCredits:     1,
		ReservedCredits:  0,
		AvailableCredits: 1,
		Version:          1,
		CreatedAt:        svc.now(),
		UpdatedAt:        svc.now(),
	}
	repo.priceBooks["pb-1"] = &CreditPriceBook{
		Id:                 "pb-1",
		CreditType:         CreditTypePlatform,
		ModelId:            "model-1",
		ModelScope:         "platform",
		ModelSourceType:    ModelSourceTypePlatform,
		FundingType:        FundingTypePlatform,
		CreditUnitUsd:      DefaultCreditUnitUSD,
		InputCreditsPer1M:  1000,
		OutputCreditsPer1M: 2000,
		EffectiveAt:        svc.now().Add(-time.Hour),
		Status:             PriceBookStatusActive,
	}

	_, _, err := svc.Reserve(ReserveInput{
		InvocationId:              "inv-low-1",
		EnterpriseId:              "ent-1",
		MemberUserId:              "user-1",
		ModelId:                   "model-1",
		ModelScope:                "platform",
		EstimatedPromptTokens:     1000,
		EstimatedCompletionTokens: 1000,
	})
	if err == nil {
		t.Fatalf("Reserve expected insufficient credits error")
	}
	if err.Error() != "insufficient credits" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReleaseRestoresReservedCredits(t *testing.T) {
	repo := newStubRepository()
	svc := NewService(repo)
	svc.usageProvider = &stubUsageOverviewProvider{}
	svc.now = func() time.Time { return time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC) }
	idCounter := 0
	svc.idGenerator = func() string {
		idCounter++
		return fmt.Sprintf("id-release-%d", idCounter)
	}
	repo.wallets[walletKey("ent-1", CreditTypePlatform)] = &CreditWallet{
		Id:               "wallet-1",
		EnterpriseId:     "ent-1",
		CreditType:       CreditTypePlatform,
		TotalCredits:     1000,
		ReservedCredits:  0,
		AvailableCredits: 1000,
		Version:          1,
		CreatedAt:        svc.now(),
		UpdatedAt:        svc.now(),
	}
	repo.priceBooks["pb-1"] = &CreditPriceBook{
		Id:                 "pb-1",
		CreditType:         CreditTypePlatform,
		ModelId:            "model-1",
		ModelScope:         "platform",
		ModelSourceType:    ModelSourceTypePlatform,
		FundingType:        FundingTypePlatform,
		CreditUnitUsd:      DefaultCreditUnitUSD,
		InputCreditsPer1M:  1000,
		OutputCreditsPer1M: 2000,
		EffectiveAt:        svc.now().Add(-time.Hour),
		Status:             PriceBookStatusActive,
	}

	reservation, _, err := svc.Reserve(ReserveInput{
		InvocationId:              "inv-release-1",
		EnterpriseId:              "ent-1",
		MemberUserId:              "user-1",
		ModelId:                   "model-1",
		ModelScope:                "platform",
		EstimatedPromptTokens:     1000,
		EstimatedCompletionTokens: 1000,
	})
	if err != nil {
		t.Fatalf("Reserve returned error: %v", err)
	}
	if reservation == nil {
		t.Fatalf("Reserve returned nil reservation")
	}

	_, _, err = svc.Release("inv-release-1", "runtime_failed")
	if err != nil {
		t.Fatalf("Release returned error: %v", err)
	}
	wallet := repo.wallets[walletKey("ent-1", CreditTypePlatform)]
	if wallet == nil {
		t.Fatalf("wallet not found after release")
	}
	if wallet.ReservedCredits != 0 || wallet.AvailableCredits != 1000 {
		t.Fatalf("unexpected wallet state after release: %+v", wallet)
	}
	if repo.reservations["inv-release-1"] == nil || repo.reservations["inv-release-1"].Status != "released" {
		t.Fatalf("expected released reservation, got %+v", repo.reservations["inv-release-1"])
	}
	if len(repo.ledger) < 2 || repo.ledger[len(repo.ledger)-1].EntryType != EntryTypeRelease {
		t.Fatalf("expected release ledger entry, got %+v", repo.ledger)
	}
}

func TestReleaseTruncatesLongReasonCode(t *testing.T) {
	repo := newStubRepository()
	svc := NewService(repo)
	svc.usageProvider = &stubUsageOverviewProvider{}
	svc.now = func() time.Time { return time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC) }
	svc.idGenerator = func() string { return "release-long-reason-id" }
	repo.wallets[walletKey("ent-1", CreditTypePlatform)] = &CreditWallet{
		Id:               "wallet-1",
		EnterpriseId:     "ent-1",
		CreditType:       CreditTypePlatform,
		TotalCredits:     1000,
		ReservedCredits:  0,
		AvailableCredits: 1000,
		Version:          1,
		CreatedAt:        svc.now(),
		UpdatedAt:        svc.now(),
	}
	repo.priceBooks["pb-1"] = &CreditPriceBook{
		Id:                 "pb-1",
		CreditType:         CreditTypePlatform,
		ModelId:            "model-1",
		ModelScope:         "platform",
		ModelSourceType:    ModelSourceTypePlatform,
		FundingType:        FundingTypePlatform,
		CreditUnitUsd:      DefaultCreditUnitUSD,
		InputCreditsPer1M:  1000,
		OutputCreditsPer1M: 2000,
		EffectiveAt:        svc.now().Add(-time.Hour),
		Status:             PriceBookStatusActive,
	}

	_, _, err := svc.Reserve(ReserveInput{
		InvocationId:              "inv-release-long-reason",
		EnterpriseId:              "ent-1",
		MemberUserId:              "user-1",
		ModelId:                   "model-1",
		ModelScope:                "platform",
		EstimatedPromptTokens:     1000,
		EstimatedCompletionTokens: 1000,
	})
	if err != nil {
		t.Fatalf("Reserve returned error: %v", err)
	}

	longReason := strings.Repeat("x", maxReasonCodeLength+20)
	_, _, err = svc.Release("inv-release-long-reason", longReason)
	if err != nil {
		t.Fatalf("Release returned error: %v", err)
	}
	if len(repo.ledger) < 2 {
		t.Fatalf("expected release ledger entry, got %+v", repo.ledger)
	}
	got := repo.ledger[len(repo.ledger)-1].ReasonCode
	if len(got) != maxReasonCodeLength {
		t.Fatalf("expected truncated reason code length %d, got %d (%q)", maxReasonCodeLength, len(got), got)
	}
	if got != longReason[:maxReasonCodeLength] {
		t.Fatalf("expected truncated reason code %q, got %q", longReason[:maxReasonCodeLength], got)
	}
}

func TestReserveBlockedByMemberBudgetPolicy(t *testing.T) {
	repo := newStubRepository()
	svc := NewService(repo)
	svc.usageProvider = &stubUsageOverviewProvider{}
	svc.now = func() time.Time { return time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC) }
	svc.idGenerator = func() string { return "budget-id" }
	repo.wallets[walletKey("ent-1", CreditTypePlatform)] = &CreditWallet{
		Id:               "wallet-1",
		EnterpriseId:     "ent-1",
		CreditType:       CreditTypePlatform,
		TotalCredits:     1000,
		ReservedCredits:  0,
		AvailableCredits: 1000,
		Version:          1,
		CreatedAt:        svc.now(),
		UpdatedAt:        svc.now(),
	}
	repo.priceBooks["pb-1"] = &CreditPriceBook{
		Id:                 "pb-1",
		CreditType:         CreditTypePlatform,
		ModelId:            "model-1",
		ModelScope:         "platform",
		ModelSourceType:    ModelSourceTypePlatform,
		FundingType:        FundingTypePlatform,
		CreditUnitUsd:      DefaultCreditUnitUSD,
		InputCreditsPer1M:  1000,
		OutputCreditsPer1M: 2000,
		EffectiveAt:        svc.now().Add(-time.Hour),
		Status:             PriceBookStatusActive,
	}
	repo.budgetPolicies["policy-1"] = &CreditBudgetPolicy{
		Id:               "policy-1",
		EnterpriseId:     "ent-1",
		CreditType:       CreditTypePlatform,
		ScopeType:        ScopeMember,
		ScopeId:          "user-1",
		Enabled:          true,
		DailyCreditLimit: 2,
		HardLimit:        true,
		CreatedAt:        svc.now(),
		UpdatedAt:        svc.now(),
	}

	_, _, err := svc.Reserve(ReserveInput{
		InvocationId:              "inv-budget-1",
		EnterpriseId:              "ent-1",
		MemberUserId:              "user-1",
		ModelId:                   "model-1",
		ModelScope:                "platform",
		EstimatedPromptTokens:     1000,
		EstimatedCompletionTokens: 1000,
	})
	if err == nil {
		t.Fatalf("Reserve expected budget policy error")
	}
	if err.Error() != "member credit daily limit exceeded" {
		t.Fatalf("unexpected error: %v", err)
	}
}
