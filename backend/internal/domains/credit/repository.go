package credit

import "time"

type Repository interface {
	GetWallet(enterpriseId, creditType string) (*CreditWallet, error)
	ListWallets(filter WalletFilter) ([]CreditWallet, error)
	InsertWallet(item *CreditWallet) error
	GetGrantBySourceRef(enterpriseId, creditType, sourceType, sourceRefId string) (*CreditGrant, error)
	ApplyGrant(wallet *CreditWallet, grant *CreditGrant, entry *CreditLedgerEntry) (*CreditGrant, *CreditWallet, error)
	AggregateSettledCredits(enterpriseId, creditType, scopeType, scopeId string, from, to time.Time) (int64, error)

	ListGrants(filter GrantFilter) ([]CreditGrant, error)
	ListLedger(filter LedgerFilter) ([]CreditLedgerEntry, error)
	AggregateLedgerByType(enterpriseId string, from, to time.Time) ([]LedgerAggregate, error)

	GetPriceBookByID(id string) (*CreditPriceBook, error)
	ListPriceBooks(filter PriceBookFilter) ([]CreditPriceBook, error)
	InsertPriceBook(item *CreditPriceBook) error
	UpdatePriceBook(item *CreditPriceBook) error
	DeletePriceBook(id string) error

	GetBudgetPolicyByID(id string) (*CreditBudgetPolicy, error)
	ListBudgetPolicies(filter BudgetPolicyFilter) ([]CreditBudgetPolicy, error)
	InsertBudgetPolicy(item *CreditBudgetPolicy) error
	UpdateBudgetPolicy(item *CreditBudgetPolicy) error
	DeleteBudgetPolicy(id string) error

	GetReservationByInvocationID(invocationId string) (*CreditReservation, error)
	ApplyReservation(wallet *CreditWallet, reservation *CreditReservation, entry *CreditLedgerEntry) error
	ApplySettlement(wallet *CreditWallet, reservation *CreditReservation, settleEntry *CreditLedgerEntry, releaseEntry *CreditLedgerEntry) error
	ApplyRelease(wallet *CreditWallet, reservation *CreditReservation, entry *CreditLedgerEntry) error
}
