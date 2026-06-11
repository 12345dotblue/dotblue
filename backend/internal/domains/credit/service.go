package credit

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"dotblue/internal/domains/metering"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/google/uuid"
)

const maxReasonCodeLength = 64

type Service struct {
	repo          Repository
	usageProvider usageOverviewProvider
	now           func() time.Time
	idGenerator   func() string
}

type usageOverview struct {
	TodayTokens int64
	TodayCharge float64
	MonthTokens int64
	MonthCharge float64
}

type usageOverviewProvider interface {
	GetOverview(scopeType, scopeId string) (*usageOverview, error)
}

type defaultUsageOverviewProvider struct{}

func (defaultUsageOverviewProvider) GetOverview(scopeType, scopeId string) (*usageOverview, error) {
	item, err := metering.GetOverview(scopeType, scopeId)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return &usageOverview{}, nil
	}
	return &usageOverview{
		TodayTokens: item.TodayTokens,
		TodayCharge: item.TodayCharge,
		MonthTokens: item.MonthTokens,
		MonthCharge: item.MonthCharge,
	}, nil
}

func NewService(repo Repository) *Service {
	return &Service{
		repo:          repo,
		usageProvider: defaultUsageOverviewProvider{},
		now:           time.Now,
		idGenerator:   func() string { return uuid.NewString() },
	}
}

var defaultService = NewService(NewGFRepository())

func GetOverview(enterpriseId string) (*Overview, error) {
	return defaultService.GetOverview(enterpriseId)
}

func ListWallets(filter WalletFilter) ([]CreditWallet, error) {
	return defaultService.ListWallets(filter)
}

func CreateGrant(enterpriseId, creditType string, req CreateGrantReq) (*CreditGrant, *CreditWallet, error) {
	return defaultService.CreateGrant(enterpriseId, creditType, req)
}

func ListGrants(filter GrantFilter) ([]CreditGrant, error) {
	return defaultService.ListGrants(filter)
}

func ListLedger(filter LedgerFilter) ([]CreditLedgerEntry, error) {
	return defaultService.ListLedger(filter)
}

func ListPriceBooks(filter PriceBookFilter) ([]CreditPriceBook, error) {
	return defaultService.ListPriceBooks(filter)
}

func CreatePriceBook(enterpriseId string, req CreatePriceBookReq) (*CreditPriceBook, error) {
	return defaultService.CreatePriceBook(enterpriseId, req)
}

func UpdatePriceBook(id, enterpriseId string, req UpdatePriceBookReq) (*CreditPriceBook, error) {
	return defaultService.UpdatePriceBook(id, enterpriseId, req)
}

func DeletePriceBook(id, enterpriseId string) error {
	return defaultService.DeletePriceBook(id, enterpriseId)
}

func ListBudgetPolicies(filter BudgetPolicyFilter) ([]CreditBudgetPolicy, error) {
	return defaultService.ListBudgetPolicies(filter)
}

func CreateBudgetPolicy(enterpriseId string, req CreateBudgetPolicyReq) (*CreditBudgetPolicy, error) {
	return defaultService.CreateBudgetPolicy(enterpriseId, req)
}

func UpdateBudgetPolicy(id, enterpriseId string, req UpdateBudgetPolicyReq) (*CreditBudgetPolicy, error) {
	return defaultService.UpdateBudgetPolicy(id, enterpriseId, req)
}

func DeleteBudgetPolicy(id, enterpriseId string) error {
	return defaultService.DeleteBudgetPolicy(id, enterpriseId)
}

func Reserve(input ReserveInput) (*CreditReservation, *CreditSnapshot, error) {
	return defaultService.Reserve(input)
}

func Settle(input SettleInput) (*CreditReservation, *CreditSnapshot, error) {
	return defaultService.Settle(input)
}

func Release(invocationId, reasonCode string) (*CreditReservation, *CreditSnapshot, error) {
	return defaultService.Release(invocationId, reasonCode)
}

func (s *Service) GetOverview(enterpriseId string) (*Overview, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("credit service is not configured")
	}
	enterpriseId = strings.TrimSpace(enterpriseId)
	if enterpriseId == "" {
		return nil, errors.New("enterprise id is required")
	}
	// Overview should always present both wallet buckets so the UI can remain stable.
	if _, err := s.ensureWallet(enterpriseId, CreditTypePlatform); err != nil {
		return nil, err
	}
	if _, err := s.ensureWallet(enterpriseId, CreditTypeEnterprise); err != nil {
		return nil, err
	}
	wallets, err := s.repo.ListWallets(WalletFilter{EnterpriseId: enterpriseId})
	if err != nil {
		return nil, err
	}
	now := s.now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	aggregates, err := s.repo.AggregateLedgerByType(enterpriseId, monthStart, now.Add(time.Second))
	if err != nil {
		return nil, err
	}
	aggregateMap := make(map[string]LedgerAggregate, len(aggregates))
	for _, item := range aggregates {
		aggregateMap[item.CreditType] = item
	}
	result := &Overview{
		EnterpriseId: enterpriseId,
		Wallets:      make([]WalletOverview, 0, len(wallets)),
	}
	for _, wallet := range wallets {
		agg := aggregateMap[wallet.CreditType]
		result.Wallets = append(result.Wallets, WalletOverview{
			CreditType:       wallet.CreditType,
			TotalCredits:     wallet.TotalCredits,
			ReservedCredits:  wallet.ReservedCredits,
			AvailableCredits: wallet.AvailableCredits,
			GrantedCredits:   agg.GrantedCredits,
			SettledCredits:   agg.SettledCredits,
			ExpiredCredits:   agg.ExpiredCredits,
		})
	}
	return result, nil
}

func (s *Service) ListWallets(filter WalletFilter) ([]CreditWallet, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("credit service is not configured")
	}
	filter.EnterpriseId = strings.TrimSpace(filter.EnterpriseId)
	filter.CreditType = normalizeCreditType(filter.CreditType)
	if filter.EnterpriseId == "" {
		return nil, errors.New("enterprise id is required")
	}
	if filter.CreditType == "" {
		if _, err := s.ensureWallet(filter.EnterpriseId, CreditTypePlatform); err != nil {
			return nil, err
		}
		if _, err := s.ensureWallet(filter.EnterpriseId, CreditTypeEnterprise); err != nil {
			return nil, err
		}
	} else {
		if _, err := s.ensureWallet(filter.EnterpriseId, filter.CreditType); err != nil {
			return nil, err
		}
	}
	return s.repo.ListWallets(filter)
}

func (s *Service) CreateGrant(enterpriseId, creditType string, req CreateGrantReq) (*CreditGrant, *CreditWallet, error) {
	if s == nil || s.repo == nil {
		return nil, nil, errors.New("credit service is not configured")
	}
	enterpriseId = strings.TrimSpace(enterpriseId)
	creditType = normalizeCreditType(creditType)
	if enterpriseId == "" {
		return nil, nil, errors.New("enterprise id is required")
	}
	if creditType == "" {
		return nil, nil, errors.New("credit type is required")
	}
	if req.Credits <= 0 {
		return nil, nil, errors.New("credits must be greater than 0")
	}
	if strings.TrimSpace(req.SourceType) == "" {
		return nil, nil, errors.New("source type is required")
	}
	if strings.TrimSpace(req.SourceRefId) != "" {
		existing, err := s.repo.GetGrantBySourceRef(enterpriseId, creditType, strings.TrimSpace(req.SourceType), strings.TrimSpace(req.SourceRefId))
		if err != nil {
			return nil, nil, err
		}
		if existing != nil {
			wallet, err := s.ensureWallet(enterpriseId, creditType)
			if err != nil {
				return nil, nil, err
			}
			return existing, wallet, nil
		}
	}
	wallet, err := s.ensureWallet(enterpriseId, creditType)
	if err != nil {
		return nil, nil, err
	}
	now := s.now()
	wallet.TotalCredits += req.Credits
	wallet.AvailableCredits += req.Credits
	wallet.Version++
	wallet.UpdatedAt = now
	effectiveAt := now
	if req.EffectiveAt != nil && !req.EffectiveAt.IsZero() {
		effectiveAt = req.EffectiveAt.UTC()
	}
	grant := &CreditGrant{
		Id:               s.idGenerator(),
		EnterpriseId:     enterpriseId,
		CreditType:       creditType,
		SourceType:       strings.TrimSpace(req.SourceType),
		SourceRefId:      strings.TrimSpace(req.SourceRefId),
		GrantedCredits:   req.Credits,
		RemainingCredits: req.Credits,
		EffectiveAt:      effectiveAt,
		ExpiresAt:        req.ExpiresAt,
		MetadataJson:     normalizeJSON(req.MetadataJson),
		CreatedAt:        now,
	}
	entry := &CreditLedgerEntry{
		Id:            s.idGenerator(),
		EnterpriseId:  enterpriseId,
		CreditType:    creditType,
		WalletId:      wallet.Id,
		GrantId:       grant.Id,
		EntryType:     EntryTypeGrant,
		Direction:     DirectionIn,
		Credits:       req.Credits,
		BalanceAfter:  wallet.AvailableCredits,
		ReservedAfter: wallet.ReservedCredits,
		ReasonCode:    normalizeReasonCode(req.ReasonCode, EntryTypeGrant),
		SnapshotJson:  normalizeJSON(req.MetadataJson),
		CreatedAt:     now,
	}
	appliedGrant, appliedWallet, err := s.repo.ApplyGrant(wallet, grant, entry)
	if err != nil {
		return nil, nil, err
	}
	if appliedGrant != nil {
		grant = appliedGrant
	}
	if appliedWallet != nil {
		wallet = appliedWallet
	}
	return grant, wallet, nil
}

func (s *Service) ListGrants(filter GrantFilter) ([]CreditGrant, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("credit service is not configured")
	}
	filter.EnterpriseId = strings.TrimSpace(filter.EnterpriseId)
	filter.CreditType = normalizeCreditType(filter.CreditType)
	if filter.EnterpriseId == "" {
		return nil, errors.New("enterprise id is required")
	}
	return s.repo.ListGrants(filter)
}

func (s *Service) ListLedger(filter LedgerFilter) ([]CreditLedgerEntry, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("credit service is not configured")
	}
	filter.EnterpriseId = strings.TrimSpace(filter.EnterpriseId)
	filter.CreditType = normalizeCreditType(filter.CreditType)
	if filter.EnterpriseId == "" {
		return nil, errors.New("enterprise id is required")
	}
	return s.repo.ListLedger(filter)
}

func (s *Service) ListPriceBooks(filter PriceBookFilter) ([]CreditPriceBook, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("credit service is not configured")
	}
	filter.EnterpriseId = strings.TrimSpace(filter.EnterpriseId)
	filter.CreditType = normalizeCreditType(filter.CreditType)
	filter.ModelId = strings.TrimSpace(filter.ModelId)
	return s.repo.ListPriceBooks(filter)
}

func (s *Service) CreatePriceBook(enterpriseId string, req CreatePriceBookReq) (*CreditPriceBook, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("credit service is not configured")
	}
	item, err := s.buildPriceBook("", enterpriseId, req)
	if err != nil {
		return nil, err
	}
	if err := s.repo.InsertPriceBook(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) UpdatePriceBook(id, enterpriseId string, req UpdatePriceBookReq) (*CreditPriceBook, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("credit service is not configured")
	}
	existing, err := s.repo.GetPriceBookByID(strings.TrimSpace(id))
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("credit price book not found")
	}
	if normalizeEnterpriseID(existing.EnterpriseId) != normalizeEnterpriseID(enterpriseId) {
		return nil, errors.New("credit price book scope mismatch")
	}
	item, err := s.buildPriceBook(existing.Id, enterpriseId, CreatePriceBookReq(req))
	if err != nil {
		return nil, err
	}
	item.CreatedAt = existing.CreatedAt
	if err := s.repo.UpdatePriceBook(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) DeletePriceBook(id, enterpriseId string) error {
	if s == nil || s.repo == nil {
		return errors.New("credit service is not configured")
	}
	existing, err := s.repo.GetPriceBookByID(strings.TrimSpace(id))
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("credit price book not found")
	}
	if normalizeEnterpriseID(existing.EnterpriseId) != normalizeEnterpriseID(enterpriseId) {
		return errors.New("credit price book scope mismatch")
	}
	return s.repo.DeletePriceBook(existing.Id)
}

func (s *Service) ListBudgetPolicies(filter BudgetPolicyFilter) ([]CreditBudgetPolicy, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("credit service is not configured")
	}
	filter.EnterpriseId = strings.TrimSpace(filter.EnterpriseId)
	filter.CreditType = normalizeCreditType(filter.CreditType)
	filter.ScopeType = normalizeScopeType(filter.ScopeType)
	filter.ScopeId = strings.TrimSpace(filter.ScopeId)
	if filter.EnterpriseId == "" {
		return nil, errors.New("enterprise id is required")
	}
	return s.repo.ListBudgetPolicies(filter)
}

func (s *Service) CreateBudgetPolicy(enterpriseId string, req CreateBudgetPolicyReq) (*CreditBudgetPolicy, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("credit service is not configured")
	}
	item, err := s.buildBudgetPolicy("", enterpriseId, req)
	if err != nil {
		return nil, err
	}
	if err := s.repo.InsertBudgetPolicy(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) UpdateBudgetPolicy(id, enterpriseId string, req UpdateBudgetPolicyReq) (*CreditBudgetPolicy, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("credit service is not configured")
	}
	existing, err := s.repo.GetBudgetPolicyByID(strings.TrimSpace(id))
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("credit budget policy not found")
	}
	if existing.EnterpriseId != strings.TrimSpace(enterpriseId) {
		return nil, errors.New("credit budget policy scope mismatch")
	}
	item, err := s.buildBudgetPolicy(existing.Id, enterpriseId, CreateBudgetPolicyReq(req))
	if err != nil {
		return nil, err
	}
	item.CreatedAt = existing.CreatedAt
	if err := s.repo.UpdateBudgetPolicy(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) DeleteBudgetPolicy(id, enterpriseId string) error {
	if s == nil || s.repo == nil {
		return errors.New("credit service is not configured")
	}
	existing, err := s.repo.GetBudgetPolicyByID(strings.TrimSpace(id))
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("credit budget policy not found")
	}
	if existing.EnterpriseId != strings.TrimSpace(enterpriseId) {
		return errors.New("credit budget policy scope mismatch")
	}
	return s.repo.DeleteBudgetPolicy(existing.Id)
}

func (s *Service) Reserve(input ReserveInput) (*CreditReservation, *CreditSnapshot, error) {
	if s == nil || s.repo == nil {
		return nil, nil, errors.New("credit service is not configured")
	}
	input.InvocationId = strings.TrimSpace(input.InvocationId)
	input.EnterpriseId = strings.TrimSpace(input.EnterpriseId)
	input.MemberUserId = strings.TrimSpace(input.MemberUserId)
	input.AgentId = strings.TrimSpace(input.AgentId)
	input.ModelId = strings.TrimSpace(input.ModelId)
	input.ModelScope = strings.TrimSpace(input.ModelScope)
	input.FundingType = normalizeFundingType(input.FundingType)
	input.ModelSourceType = normalizeModelSourceType(input.ModelSourceType)
	if input.InvocationId == "" {
		return nil, nil, errors.New("invocation id is required")
	}
	if input.EnterpriseId == "" {
		return nil, nil, errors.New("enterprise id is required")
	}
	if input.ModelId == "" {
		return nil, nil, errors.New("model id is required")
	}
	existing, err := s.repo.GetReservationByInvocationID(input.InvocationId)
	if err != nil {
		return nil, nil, err
	}
	if existing != nil {
		return existing, snapshotFromReservation(existing), nil
	}
	priceBook, creditType, fundingType, err := s.resolveRuntimePriceBook(input)
	if err != nil {
		return nil, nil, err
	}
	// Budget checks must happen before wallet mutation so scope limits stay deterministic.
	if err := s.checkBudgetPolicies(input, creditType, priceBook); err != nil {
		g.Log().Warningf(context.Background(), "credit.reserve.blocked invocation=%s enterprise=%s creditType=%s reason=%v", input.InvocationId, input.EnterpriseId, creditType, err)
		return nil, nil, err
	}
	wallet, err := s.ensureWallet(input.EnterpriseId, creditType)
	if err != nil {
		return nil, nil, err
	}
	reservedCredits := calculateReservedCredits(input.EstimatedPromptTokens, input.EstimatedCompletionTokens, priceBook)
	if reservedCredits <= 0 {
		reservedCredits = 1
	}
	if wallet.AvailableCredits < reservedCredits {
		g.Log().Warningf(context.Background(), "credit.reserve.insufficient invocation=%s enterprise=%s creditType=%s available=%d required=%d", input.InvocationId, input.EnterpriseId, creditType, wallet.AvailableCredits, reservedCredits)
		return nil, nil, errors.New("insufficient credits")
	}
	now := s.now()
	wallet.ReservedCredits += reservedCredits
	wallet.AvailableCredits -= reservedCredits
	wallet.Version++
	wallet.UpdatedAt = now
	priceSnapshot, err := json.Marshal(priceBook)
	if err != nil {
		return nil, nil, err
	}
	reservation := &CreditReservation{
		Id:                s.idGenerator(),
		EnterpriseId:      input.EnterpriseId,
		CreditType:        creditType,
		InvocationId:      input.InvocationId,
		MemberUserId:      input.MemberUserId,
		AgentId:           input.AgentId,
		ModelId:           input.ModelId,
		ModelScope:        input.ModelScope,
		FundingType:       fundingType,
		PriceBookId:       priceBook.Id,
		PriceSnapshotJson: string(priceSnapshot),
		ReservedCredits:   reservedCredits,
		Status:            "active",
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	entry := &CreditLedgerEntry{
		Id:            s.idGenerator(),
		EnterpriseId:  wallet.EnterpriseId,
		CreditType:    wallet.CreditType,
		WalletId:      wallet.Id,
		EntryType:     EntryTypeReserve,
		Direction:     DirectionOut,
		Credits:       reservedCredits,
		BalanceAfter:  wallet.AvailableCredits,
		ReservedAfter: wallet.ReservedCredits,
		InvocationId:  input.InvocationId,
		MemberUserId:  input.MemberUserId,
		AgentId:       input.AgentId,
		ReasonCode:    EntryTypeReserve,
		SnapshotJson:  string(priceSnapshot),
		CreatedAt:     now,
	}
	if err := s.repo.ApplyReservation(wallet, reservation, entry); err != nil {
		return nil, nil, err
	}
	g.Log().Infof(context.Background(), "credit.reserve.success invocation=%s enterprise=%s creditType=%s reserved=%d", input.InvocationId, input.EnterpriseId, creditType, reservedCredits)
	return reservation, snapshotFromReservation(reservation), nil
}

func (s *Service) Settle(input SettleInput) (*CreditReservation, *CreditSnapshot, error) {
	if s == nil || s.repo == nil {
		return nil, nil, errors.New("credit service is not configured")
	}
	input.InvocationId = strings.TrimSpace(input.InvocationId)
	if input.InvocationId == "" {
		return nil, nil, errors.New("invocation id is required")
	}
	reservation, err := s.repo.GetReservationByInvocationID(input.InvocationId)
	if err != nil {
		return nil, nil, err
	}
	if reservation == nil {
		return nil, nil, errors.New("credit reservation not found")
	}
	if reservation.Status == "settled" {
		return reservation, snapshotFromReservation(reservation), nil
	}
	priceBook, err := s.priceBookFromReservation(reservation)
	if err != nil {
		return nil, nil, err
	}
	wallet, err := s.ensureWallet(reservation.EnterpriseId, reservation.CreditType)
	if err != nil {
		return nil, nil, err
	}
	settledCredits := calculateSettledCredits(input.PromptTokens, input.CompletionTokens, priceBook)
	additionalCredits := settledCredits - reservation.ReservedCredits
	if additionalCredits > 0 && wallet.AvailableCredits < additionalCredits {
		g.Log().Warningf(context.Background(), "credit.settle.insufficient invocation=%s enterprise=%s creditType=%s available=%d additional=%d", reservation.InvocationId, reservation.EnterpriseId, reservation.CreditType, wallet.AvailableCredits, additionalCredits)
		return nil, nil, errors.New("insufficient credits")
	}
	now := s.now()
	wallet.TotalCredits -= settledCredits
	if wallet.TotalCredits < 0 {
		wallet.TotalCredits = 0
	}
	wallet.ReservedCredits -= reservation.ReservedCredits
	if wallet.ReservedCredits < 0 {
		wallet.ReservedCredits = 0
	}
	if additionalCredits > 0 {
		wallet.AvailableCredits -= additionalCredits
	} else {
		wallet.AvailableCredits += reservation.ReservedCredits - settledCredits
	}
	if wallet.AvailableCredits < 0 {
		wallet.AvailableCredits = 0
	}
	wallet.Version++
	wallet.UpdatedAt = now
	reservation.Status = "settled"
	reservation.UpdatedAt = now
	settleEntry := &CreditLedgerEntry{
		Id:            s.idGenerator(),
		EnterpriseId:  wallet.EnterpriseId,
		CreditType:    wallet.CreditType,
		WalletId:      wallet.Id,
		EntryType:     EntryTypeSettle,
		Direction:     DirectionOut,
		Credits:       settledCredits,
		BalanceAfter:  wallet.AvailableCredits,
		ReservedAfter: wallet.ReservedCredits,
		InvocationId:  reservation.InvocationId,
		MemberUserId:  reservation.MemberUserId,
		AgentId:       reservation.AgentId,
		ReasonCode:    EntryTypeSettle,
		SnapshotJson:  reservation.PriceSnapshotJson,
		CreatedAt:     now,
	}
	var releaseEntry *CreditLedgerEntry
	if reservation.ReservedCredits > settledCredits {
		releaseEntry = &CreditLedgerEntry{
			Id:            s.idGenerator(),
			EnterpriseId:  wallet.EnterpriseId,
			CreditType:    wallet.CreditType,
			WalletId:      wallet.Id,
			EntryType:     EntryTypeRelease,
			Direction:     DirectionIn,
			Credits:       reservation.ReservedCredits - settledCredits,
			BalanceAfter:  wallet.AvailableCredits,
			ReservedAfter: wallet.ReservedCredits,
			InvocationId:  reservation.InvocationId,
			MemberUserId:  reservation.MemberUserId,
			AgentId:       reservation.AgentId,
			ReasonCode:    EntryTypeRelease,
			SnapshotJson:  reservation.PriceSnapshotJson,
			CreatedAt:     now,
		}
	}
	if err := s.repo.ApplySettlement(wallet, reservation, settleEntry, releaseEntry); err != nil {
		return nil, nil, err
	}
	snapshot := snapshotFromReservation(reservation)
	snapshot.SettledCredits = settledCredits
	g.Log().Infof(context.Background(), "credit.settle.success invocation=%s enterprise=%s creditType=%s settled=%d reserved=%d", reservation.InvocationId, reservation.EnterpriseId, reservation.CreditType, settledCredits, reservation.ReservedCredits)
	return reservation, snapshot, nil
}

func (s *Service) Release(invocationId, reasonCode string) (*CreditReservation, *CreditSnapshot, error) {
	if s == nil || s.repo == nil {
		return nil, nil, errors.New("credit service is not configured")
	}
	invocationId = strings.TrimSpace(invocationId)
	if invocationId == "" {
		return nil, nil, errors.New("invocation id is required")
	}
	reservation, err := s.repo.GetReservationByInvocationID(invocationId)
	if err != nil {
		return nil, nil, err
	}
	if reservation == nil {
		return nil, nil, nil
	}
	if reservation.Status == "released" || reservation.Status == "settled" {
		return reservation, snapshotFromReservation(reservation), nil
	}
	wallet, err := s.ensureWallet(reservation.EnterpriseId, reservation.CreditType)
	if err != nil {
		return nil, nil, err
	}
	now := s.now()
	wallet.ReservedCredits -= reservation.ReservedCredits
	if wallet.ReservedCredits < 0 {
		wallet.ReservedCredits = 0
	}
	wallet.AvailableCredits += reservation.ReservedCredits
	wallet.Version++
	wallet.UpdatedAt = now
	reservation.Status = "released"
	reservation.UpdatedAt = now
	entry := &CreditLedgerEntry{
		Id:            s.idGenerator(),
		EnterpriseId:  wallet.EnterpriseId,
		CreditType:    wallet.CreditType,
		WalletId:      wallet.Id,
		EntryType:     EntryTypeRelease,
		Direction:     DirectionIn,
		Credits:       reservation.ReservedCredits,
		BalanceAfter:  wallet.AvailableCredits,
		ReservedAfter: wallet.ReservedCredits,
		InvocationId:  reservation.InvocationId,
		MemberUserId:  reservation.MemberUserId,
		AgentId:       reservation.AgentId,
		ReasonCode:    normalizeReasonCode(reasonCode, EntryTypeRelease),
		SnapshotJson:  reservation.PriceSnapshotJson,
		CreatedAt:     now,
	}
	if err := s.repo.ApplyRelease(wallet, reservation, entry); err != nil {
		return nil, nil, err
	}
	g.Log().Infof(context.Background(), "credit.release.success invocation=%s enterprise=%s creditType=%s released=%d", reservation.InvocationId, reservation.EnterpriseId, reservation.CreditType, reservation.ReservedCredits)
	return reservation, snapshotFromReservation(reservation), nil
}

func (s *Service) ensureWallet(enterpriseId, creditType string) (*CreditWallet, error) {
	item, err := s.repo.GetWallet(enterpriseId, creditType)
	if err != nil {
		return nil, err
	}
	if item != nil {
		return item, nil
	}
	now := s.now()
	item = &CreditWallet{
		Id:               s.idGenerator(),
		EnterpriseId:     enterpriseId,
		CreditType:       creditType,
		TotalCredits:     0,
		ReservedCredits:  0,
		AvailableCredits: 0,
		Version:          1,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := s.repo.InsertWallet(item); err != nil {
		return nil, err
	}
	return s.repo.GetWallet(enterpriseId, creditType)
}

func (s *Service) buildPriceBook(id, enterpriseId string, req CreatePriceBookReq) (*CreditPriceBook, error) {
	creditType := normalizeCreditType(req.CreditType)
	if creditType == "" {
		return nil, errors.New("credit type is required")
	}
	modelId := strings.TrimSpace(req.ModelId)
	if modelId == "" {
		return nil, errors.New("model id is required")
	}
	modelScope := strings.TrimSpace(req.ModelScope)
	if modelScope == "" {
		return nil, errors.New("model scope is required")
	}
	modelSourceType := normalizeModelSourceType(req.ModelSourceType)
	if modelSourceType == "" {
		return nil, errors.New("model source type is required")
	}
	fundingType := normalizeFundingType(req.FundingType)
	if fundingType == "" {
		return nil, errors.New("funding type is required")
	}
	enterpriseId = normalizeEnterpriseID(enterpriseId)
	if fundingType == FundingTypeEnterprise && enterpriseId == "" {
		return nil, errors.New("enterprise-funded price book requires enterprise scope")
	}
	currency := normalizeCurrency(req.Currency)
	if currency == "" {
		return nil, errors.New("currency must be USD or CNY")
	}
	creditUnitPrice := req.CreditUnitUsd
	if creditUnitPrice <= 0 {
		creditUnitPrice = DefaultCreditUnitUSD
	}
	platformMultiplier := req.PlatformMultiplier
	if platformMultiplier <= 0 {
		platformMultiplier = 1
	}
	enterpriseMultiplier := req.EnterpriseMultiplier
	if enterpriseMultiplier <= 0 {
		enterpriseMultiplier = 1
	}
	if fundingType == FundingTypeEnterprise {
		platformMultiplier = 1
	}
	costInput := max0(req.CostInputUsdPer1M)
	costOutput := max0(req.CostOutputUsdPer1M)
	billableInput := roundPrice(costInput * platformMultiplier * enterpriseMultiplier)
	billableOutput := roundPrice(costOutput * platformMultiplier * enterpriseMultiplier)
	if fundingType == FundingTypeEnterprise {
		billableInput = roundPrice(costInput * enterpriseMultiplier)
		billableOutput = roundPrice(costOutput * enterpriseMultiplier)
	}
	now := s.now()
	effectiveAt := req.EffectiveAt
	if effectiveAt.IsZero() {
		effectiveAt = now
	}
	item := &CreditPriceBook{
		Id:                     strings.TrimSpace(id),
		EnterpriseId:           enterpriseId,
		CreditType:             creditType,
		ModelId:                modelId,
		ModelScope:             modelScope,
		ModelSourceType:        modelSourceType,
		FundingType:            fundingType,
		Currency:               currency,
		CreditUnitUsd:          creditUnitPrice,
		CostInputUsdPer1M:      costInput,
		CostOutputUsdPer1M:     costOutput,
		PlatformMultiplier:     platformMultiplier,
		EnterpriseMultiplier:   enterpriseMultiplier,
		BillableInputUsdPer1M:  billableInput,
		BillableOutputUsdPer1M: billableOutput,
		InputCreditsPer1M:      usdToCredits(billableInput, creditUnitPrice),
		OutputCreditsPer1M:     usdToCredits(billableOutput, creditUnitPrice),
		EffectiveAt:            effectiveAt,
		Status:                 normalizePriceBookStatus(req.Status),
		CreatedAt:              now,
		UpdatedAt:              now,
	}
	if item.Id == "" {
		item.Id = s.idGenerator()
	}
	return item, nil
}

func (s *Service) buildBudgetPolicy(id, enterpriseId string, req CreateBudgetPolicyReq) (*CreditBudgetPolicy, error) {
	enterpriseId = strings.TrimSpace(enterpriseId)
	if enterpriseId == "" {
		return nil, errors.New("enterprise id is required")
	}
	creditType := normalizeCreditType(req.CreditType)
	if creditType == "" {
		return nil, errors.New("credit type is required")
	}
	scopeType := normalizeScopeType(req.ScopeType)
	if scopeType == "" {
		return nil, errors.New("scope type is required")
	}
	scopeId := strings.TrimSpace(req.ScopeId)
	if scopeId == "" {
		return nil, errors.New("scope id is required")
	}
	now := s.now()
	item := &CreditBudgetPolicy{
		Id:                 strings.TrimSpace(id),
		EnterpriseId:       enterpriseId,
		CreditType:         creditType,
		ScopeType:          scopeType,
		ScopeId:            scopeId,
		Enabled:            req.Enabled,
		DailyCreditLimit:   maxInt0(req.DailyCreditLimit),
		MonthlyCreditLimit: maxInt0(req.MonthlyCreditLimit),
		DailyTokenLimit:    maxInt0(req.DailyTokenLimit),
		MonthlyTokenLimit:  maxInt0(req.MonthlyTokenLimit),
		DailyUsdLimit:      max0(req.DailyUsdLimit),
		MonthlyUsdLimit:    max0(req.MonthlyUsdLimit),
		HardLimit:          req.HardLimit,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if item.Id == "" {
		item.Id = s.idGenerator()
	}
	return item, nil
}

func normalizeCreditType(value string) string {
	value = strings.TrimSpace(value)
	switch value {
	case CreditTypePlatform, CreditTypeEnterprise:
		return value
	default:
		return ""
	}
}

func (s *Service) resolveRuntimePriceBook(input ReserveInput) (*CreditPriceBook, string, string, error) {
	creditType, fundingType, modelSourceType := resolveRuntimeRouting(input.ModelScope, input.FundingType, input.ModelSourceType)
	var priceBook *CreditPriceBook
	if creditType == CreditTypeEnterprise {
		priceBook = resolveLatestActivePriceBook(s.repo, PriceBookFilter{
			EnterpriseId: input.EnterpriseId,
			CreditType:   creditType,
			ModelId:      input.ModelId,
		}, s.now())
	} else {
		priceBook = resolveLatestActivePriceBook(s.repo, PriceBookFilter{
			EnterpriseId: input.EnterpriseId,
			CreditType:   creditType,
			ModelId:      input.ModelId,
		}, s.now())
		if priceBook == nil {
			priceBook = resolveLatestActivePriceBook(s.repo, PriceBookFilter{
				CreditType: creditType,
				ModelId:    input.ModelId,
			}, s.now())
		}
	}
	if priceBook == nil {
		return nil, "", "", errors.New("credit price book not found")
	}
	if priceBook.ModelSourceType == "" {
		priceBook.ModelSourceType = modelSourceType
	}
	if priceBook.FundingType == "" {
		priceBook.FundingType = fundingType
	}
	return priceBook, creditType, fundingType, nil
}

func (s *Service) priceBookFromReservation(reservation *CreditReservation) (*CreditPriceBook, error) {
	if reservation == nil {
		return nil, errors.New("credit reservation is required")
	}
	if strings.TrimSpace(reservation.PriceSnapshotJson) != "" {
		var item CreditPriceBook
		if err := json.Unmarshal([]byte(reservation.PriceSnapshotJson), &item); err == nil && item.Id != "" {
			return &item, nil
		}
	}
	item, err := s.repo.GetPriceBookByID(strings.TrimSpace(reservation.PriceBookId))
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("credit price book not found")
	}
	return item, nil
}

func resolveRuntimeRouting(modelScope, fundingType, modelSourceType string) (string, string, string) {
	modelScope = strings.TrimSpace(modelScope)
	fundingType = normalizeFundingType(fundingType)
	modelSourceType = normalizeModelSourceType(modelSourceType)
	if fundingType == FundingTypeEnterprise {
		if modelSourceType == "" {
			modelSourceType = ModelSourceTypeEnterpriseCustom
		}
		return CreditTypeEnterprise, FundingTypeEnterprise, modelSourceType
	}
	if fundingType == FundingTypePlatform {
		if modelSourceType == "" {
			modelSourceType = ModelSourceTypePlatform
		}
		return CreditTypePlatform, FundingTypePlatform, modelSourceType
	}
	if modelScope == "enterprise" {
		return CreditTypeEnterprise, FundingTypeEnterprise, ModelSourceTypeEnterpriseCustom
	}
	return CreditTypePlatform, FundingTypePlatform, ModelSourceTypePlatform
}

func resolveLatestActivePriceBook(repo Repository, filter PriceBookFilter, now time.Time) *CreditPriceBook {
	if repo == nil {
		return nil
	}
	list, err := repo.ListPriceBooks(filter)
	if err != nil {
		return nil
	}
	for _, item := range list {
		if item.Status != PriceBookStatusActive {
			continue
		}
		if item.EffectiveAt.After(now) {
			continue
		}
		copy := item
		return &copy
	}
	return nil
}

func calculateReservedCredits(promptTokens, completionTokens int64, priceBook *CreditPriceBook) int64 {
	if priceBook == nil {
		return 0
	}
	return calculateSettledCredits(promptTokens, completionTokens, priceBook)
}

func calculateSettledCredits(promptTokens, completionTokens int64, priceBook *CreditPriceBook) int64 {
	if priceBook == nil {
		return 0
	}
	promptTokens = maxInt0(promptTokens)
	completionTokens = maxInt0(completionTokens)
	promptCredits := int64(math.Ceil(float64(promptTokens*priceBook.InputCreditsPer1M) / 1_000_000))
	completionCredits := int64(math.Ceil(float64(completionTokens*priceBook.OutputCreditsPer1M) / 1_000_000))
	return promptCredits + completionCredits
}

func (s *Service) checkBudgetPolicies(input ReserveInput, creditType string, priceBook *CreditPriceBook) error {
	if s == nil || s.repo == nil {
		return nil
	}
	estimatedCredits := calculateReservedCredits(input.EstimatedPromptTokens, input.EstimatedCompletionTokens, priceBook)
	estimatedTokens := maxInt0(input.EstimatedPromptTokens) + maxInt0(input.EstimatedCompletionTokens)
	estimatedCharge := calculateEstimatedCharge(input.EstimatedPromptTokens, input.EstimatedCompletionTokens, priceBook)
	now := s.now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	candidates := []struct {
		scopeType string
		scopeId   string
	}{
		{scopeType: ScopeAgent, scopeId: strings.TrimSpace(input.AgentId)},
		{scopeType: ScopeMember, scopeId: strings.TrimSpace(input.MemberUserId)},
		{scopeType: ScopeEnterprise, scopeId: strings.TrimSpace(input.EnterpriseId)},
	}
	for _, candidate := range candidates {
		if candidate.scopeId == "" {
			continue
		}
		policy, err := s.findBudgetPolicy(input.EnterpriseId, creditType, candidate.scopeType, candidate.scopeId)
		if err != nil {
			return err
		}
		if policy == nil || !policy.Enabled {
			continue
		}
		todayCredits, err := s.repo.AggregateSettledCredits(input.EnterpriseId, creditType, candidate.scopeType, candidate.scopeId, dayStart, now.Add(time.Second))
		if err != nil {
			return err
		}
		monthCredits, err := s.repo.AggregateSettledCredits(input.EnterpriseId, creditType, candidate.scopeType, candidate.scopeId, monthStart, now.Add(time.Second))
		if err != nil {
			return err
		}
		usageOverview, err := s.usageOverviewForBudget(candidate.scopeType, candidate.scopeId)
		if err != nil {
			return err
		}
		if err := validateBudgetPolicy(candidate.scopeType, policy, todayCredits, monthCredits, usageOverview, estimatedCredits, estimatedTokens, estimatedCharge); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) findBudgetPolicy(enterpriseId, creditType, scopeType, scopeId string) (*CreditBudgetPolicy, error) {
	items, err := s.repo.ListBudgetPolicies(BudgetPolicyFilter{
		EnterpriseId: enterpriseId,
		CreditType:   creditType,
		ScopeType:    scopeType,
		ScopeId:      scopeId,
	})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	item := items[0]
	return &item, nil
}

func (s *Service) usageOverviewForBudget(scopeType, scopeId string) (*usageOverview, error) {
	if s == nil || s.usageProvider == nil {
		return &usageOverview{}, nil
	}
	meteringScopeType := metering.ScopeEnterprise
	switch scopeType {
	case ScopeAgent:
		meteringScopeType = metering.ScopeAgent
	case ScopeMember:
		meteringScopeType = metering.ScopeUser
	}
	return s.usageProvider.GetOverview(meteringScopeType, scopeId)
}

func validateBudgetPolicy(scopeType string, policy *CreditBudgetPolicy, historicalTodayCredits, historicalMonthCredits int64, usage *usageOverview, estimatedCredits, estimatedTokens int64, estimatedCharge float64) error {
	if policy == nil || !policy.Enabled {
		return nil
	}
	if !policy.HardLimit {
		return nil
	}
	if usage == nil {
		usage = &usageOverview{}
	}
	if policy.DailyCreditLimit > 0 && historicalTodayCredits+estimatedCredits > policy.DailyCreditLimit {
		return errors.New(scopeType + " credit daily limit exceeded")
	}
	if policy.MonthlyCreditLimit > 0 && historicalMonthCredits+estimatedCredits > policy.MonthlyCreditLimit {
		return errors.New(scopeType + " credit monthly limit exceeded")
	}
	if policy.DailyTokenLimit > 0 && usage.TodayTokens+estimatedTokens > policy.DailyTokenLimit {
		return errors.New(scopeType + " token daily limit exceeded")
	}
	if policy.MonthlyTokenLimit > 0 && usage.MonthTokens+estimatedTokens > policy.MonthlyTokenLimit {
		return errors.New(scopeType + " token monthly limit exceeded")
	}
	if policy.DailyUsdLimit > 0 && usage.TodayCharge+estimatedCharge > policy.DailyUsdLimit {
		return errors.New(scopeType + " charge daily limit exceeded")
	}
	if policy.MonthlyUsdLimit > 0 && usage.MonthCharge+estimatedCharge > policy.MonthlyUsdLimit {
		return errors.New(scopeType + " charge monthly limit exceeded")
	}
	return nil
}

func calculateEstimatedCharge(promptTokens, completionTokens int64, priceBook *CreditPriceBook) float64 {
	if priceBook == nil {
		return 0
	}
	promptCharge := (float64(maxInt0(promptTokens)) * priceBook.BillableInputUsdPer1M) / 1_000_000
	completionCharge := (float64(maxInt0(completionTokens)) * priceBook.BillableOutputUsdPer1M) / 1_000_000
	return roundPrice(promptCharge + completionCharge)
}

func snapshotFromReservation(reservation *CreditReservation) *CreditSnapshot {
	if reservation == nil {
		return nil
	}
	var priceBook CreditPriceBook
	_ = json.Unmarshal([]byte(reservation.PriceSnapshotJson), &priceBook)
	return &CreditSnapshot{
		CreditType:         reservation.CreditType,
		FundingType:        reservation.FundingType,
		PriceBookId:        reservation.PriceBookId,
		CreditUnitUsd:      priceBook.CreditUnitUsd,
		InputCreditsPer1M:  priceBook.InputCreditsPer1M,
		OutputCreditsPer1M: priceBook.OutputCreditsPer1M,
		ReservedCredits:    reservation.ReservedCredits,
	}
}

func normalizeScopeType(value string) string {
	value = strings.TrimSpace(value)
	switch value {
	case ScopeEnterprise, ScopeMember, ScopeAgent:
		return value
	default:
		return ""
	}
}

func normalizeFundingType(value string) string {
	value = strings.TrimSpace(value)
	switch value {
	case FundingTypePlatform, FundingTypeEnterprise:
		return value
	default:
		return ""
	}
}

func normalizeModelSourceType(value string) string {
	value = strings.TrimSpace(value)
	switch value {
	case ModelSourceTypePlatform, ModelSourceTypeEnterpriseCustom:
		return value
	default:
		return ""
	}
}

func normalizePriceBookStatus(value string) string {
	value = strings.TrimSpace(value)
	switch value {
	case PriceBookStatusDisabled:
		return PriceBookStatusDisabled
	default:
		return PriceBookStatusActive
	}
}

func normalizeCurrency(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "", "USD":
		return "USD"
	case "CNY":
		return "CNY"
	default:
		return ""
	}
}

func normalizeJSON(value string) string {
	if strings.TrimSpace(value) == "" {
		return "{}"
	}
	return strings.TrimSpace(value)
}

func normalizeReasonCode(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		value = fallback
	}
	if utf8.RuneCountInString(value) <= maxReasonCodeLength {
		return value
	}
	runes := []rune(value)
	return string(runes[:maxReasonCodeLength])
}

func normalizeEnterpriseID(value string) string {
	return strings.TrimSpace(value)
}

func usdToCredits(amountUSD, creditUnitUSD float64) int64 {
	if amountUSD <= 0 || creditUnitUSD <= 0 {
		return 0
	}
	return int64(math.Ceil(amountUSD / creditUnitUSD))
}

func roundPrice(value float64) float64 {
	return math.Round(value*1_000_000) / 1_000_000
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
