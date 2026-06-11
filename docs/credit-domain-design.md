# DotBlue Credit Domain Design

## 1. Purpose

This document defines the detailed technical design for a first-class `credit` business domain in DotBlue.

The design addresses a core product requirement:

- AI consumption must be settled and controlled through enterprise-owned credits
- platform-funded models and enterprise-funded models must share one clean credit engine
- member and agent limits must remain governable without turning them into wallet owners
- runtime settlement must be auditable, deterministic, and compatible with existing `metering`

This document follows current repository constraints and style:

- keep one explicit business domain instead of scattering credit logic across `chat`, `metering`, and `enterprise`
- preserve `handler`, service logic, and repository separation inside the domain
- depend on abstractions instead of concrete storage or runtime implementations
- reuse existing business domains where they are already the source of truth

## 2. Background and Current Reality

DotBlue already has a usable `metering` domain that provides:

- usage event lifecycle
- model price management
- daily and monthly usage aggregation
- token and charge limit policies
- platform, enterprise, user, and agent scoped usage reporting

Relevant current references:

- [metering model](file:///c:/Users/kongz/work/dotblue/backend/internal/domains/metering/metering.go)
- [metering service](file:///c:/Users/kongz/work/dotblue/backend/internal/domains/metering/service.go)
- [enterprise bootstrap membership](file:///c:/Users/kongz/work/dotblue/backend/internal/domains/enterprise/service.go#L38-L65)
- [platform and enterprise model routes](file:///c:/Users/kongz/work/dotblue/backend/internal/cmd/cmd.go#L109-L172)

Current tenant behavior is also already clear:

- every user operates inside an enterprise context
- a personal registration still gets a default enterprise workspace
- one user may join multiple enterprises

That means:

- user is not the credit wallet owner
- enterprise is the only stable wallet owner in phase 1
- member and agent should be budget scopes, not wallet owners

## 3. Problem Statement

The current backend can measure cost and apply coarse usage limits, but it still lacks a dedicated settlement domain for credits.

Without a dedicated `credit` domain, the platform will continue to suffer from these issues:

- no explicit enterprise-owned credit wallet
- no uniform credit ledger for grants, reserves, settlements, releases, refunds, and expiry
- no clean separation between usage metering and credit settlement
- no way to support both platform-funded and enterprise-funded models with one rule system
- no deterministic runtime settlement snapshot independent from mutable multipliers
- no future-safe base for subscription allocation, top-up, contract quotas, or payment integration

## 4. Design Goals

- Introduce a dedicated `credit` business domain under `backend/internal/domains`
- Make enterprise the only wallet owner in phase 1
- Support both `platform` and `enterprise` credit types using one unified engine
- Keep `metering` as the source of token usage facts and USD-oriented operational reporting
- Make runtime settlement use snapshotted credit price books instead of mutable multiplier chains
- Support enterprise, member, and agent budgets without creating nested wallets
- Preserve existing route and middleware style used by platform admin, enterprise admin, and member flows
- Keep implementation phases incremental so existing metering and chat paths can be migrated safely

## 5. Non-Goals

- No payment gateway integration in this phase
- No invoice, order, tax, or external accounting model in this phase
- No user-owned wallet in this phase
- No department-owned wallet in this phase
- No postpaid overage settlement in phase 1
- No attempt to replace `metering` reporting with `credit` reporting

## 6. Core Principles

### 6.1 One Credit Engine

`platform credits` and `enterprise credits` must not become two separate implementations.

They share:

- the same unit definition
- the same wallet model
- the same grant lifecycle
- the same reserve and settle flow
- the same ledger entry semantics
- the same budget policy shape

They differ only by:

- `creditType`
- funding source
- price book used for settlement

### 6.2 Enterprise Owns Wallets

The only wallet owner in phase 1 is enterprise.

Users can join multiple enterprises, so tying balance to user identity would conflict with current tenancy rules.

The correct ownership model is:

- wallet owner: `enterprise`
- budget scopes: `enterprise`, `member`, `agent`

### 6.3 Metering Is Fact, Credit Is Settlement

`metering` remains responsible for:

- token counts
- request lifecycle
- operational model price configuration
- cost and charge reporting
- day and month usage trends

`credit` becomes responsible for:

- wallet balances
- grants
- ledger
- reservations
- settlements
- expiry
- budget policies by credit type
- credit price books

### 6.4 Price Book over Mutable Multiplier Math

Runtime deduction must not recalculate credits from current multipliers on every request.

The recommended rule is:

- multipliers generate a `CreditPriceBook`
- runtime snapshots the effective credit price book
- settlement uses the snapshot, not current mutable settings

This ensures:

- deterministic billing
- auditable history
- safe future repricing
- clean separation between pricing administration and runtime deduction

### 6.5 Wallet and Budget Are Different Concepts

Wallet answers:

- does the enterprise still have credits available

Budget answers:

- should this enterprise, member, or agent still be allowed to consume this credit type in this period

A request must pass both wallet checks and budget checks.

## 7. Bounded Context and Domain Placement

The new domain should be implemented as:

`backend/internal/domains/credit`

Suggested initial layout:

```text
backend/internal/domains/credit/
  credit.go
  errors.go
  policy.go
  handler_platform.go
  handler_enterprise.go
  handler_internal.go
  service_wallet.go
  service_budget.go
  service_pricing.go
  service_settlement.go
  service_admin.go
  repository.go
  repository_gf.go
  ports.go
```

Rationale:

- `credit` is a real bounded context and is important enough to stand alone
- `billing` would be too broad because payment, invoicing, and contract settlement are not in scope yet
- `metering` should not absorb wallet and ledger concerns because that would mix measurement with accounting

## 8. Terminology

- `Credit`: integer settlement unit used to govern AI consumption
- `CreditType`: funding bucket within the same credit engine
- `Wallet`: enterprise-owned balance container for one `CreditType`
- `Grant`: source batch that adds credits into a wallet
- `LedgerEntry`: immutable accounting event
- `Reservation`: temporary hold before final settlement
- `BudgetPolicy`: scoped allowance or hard limit rule
- `CreditPriceBook`: effective credit pricing snapshot for a model and funding type
- `FundingType`: who carries the underlying model cost
- `ModelSourceType`: where the model definition comes from

## 9. Credit Types and Ownership Model

### 9.1 Credit Types

Phase 1 should define exactly two credit types:

- `platform`
- `enterprise`

Semantics:

- `platform` credits are consumed when the underlying model cost is carried by the platform
- `enterprise` credits are consumed when the underlying model cost is carried by the enterprise

### 9.2 Wallet Ownership

Wallet owner must always be enterprise.

Suggested wallet key:

- `enterpriseId`
- `creditType`

There should be no phase 1 wallet keyed by:

- `userId`
- `agentId`
- `orgUnitId`

### 9.3 Budget Scopes

Budget scopes remain useful and should be supported independently of wallet ownership:

- `enterprise`
- `member`
- `agent`

Default policy direction:

- enterprise budget is mandatory
- member budget is recommended as the default per-actor control
- agent budget is optional and used for higher-cost agents or special governance

## 10. Model Funding and Credit Routing

The system must separate `modelSourceType` from `fundingType`.

### 10.1 Model Source Type

Suggested values:

- `platform_model`
- `enterprise_custom_model`

### 10.2 Funding Type

Suggested values:

- `platform_funded`
- `enterprise_funded`

### 10.3 Why Both Are Needed

`modelSourceType` explains where the model definition came from.

`fundingType` explains who pays for the underlying cost and therefore which credit type must be deducted.

Runtime credit routing must be based on `fundingType`, not on UI location alone.

### 10.4 Effective Routing Matrix

| Model Source | Funding Type | Deduct Credit Type | Wallet Owner |
| --- | --- | --- | --- |
| `platform_model` | `platform_funded` | `platform` | current enterprise |
| `enterprise_custom_model` | `enterprise_funded` | `enterprise` | current enterprise |

Phase 1 does not need additional combinations.

## 11. Credit Unit Definition

Credits should be defined as a fixed integer settlement unit.

Recommended rule:

- `1 credit = 0.0001 USD`

This value must be global and stable.

Reasons:

- avoids floating point balance drift
- small enough for fine-grained deduction
- easy to explain operationally
- allows price books to stay integer-based at runtime

Alternative values can still be adopted later, but phase 1 should lock one unit and treat it as platform-wide constant.

## 12. Pricing Model

### 12.1 The Correct Pricing Chain

The recommended pricing flow is:

1. Start from the real model cost per 1M tokens
2. Apply pricing multipliers according to business scope
3. Convert billable USD into integer credit price per 1M tokens
4. Snapshot the final price book for runtime settlement

This is more robust than a runtime rule such as "money times multiplier equals credits" on every request.

### 12.2 Base Cost

Base cost is the real USD-denominated model cost already represented today by `metering` model prices.

Suggested base fields:

- `costInputUsdPer1M`
- `costOutputUsdPer1M`

### 12.3 Multiplier Layers

For platform-funded models:

- platform config supplies `platformMultiplier`
- enterprise config supplies `enterpriseMultiplier`

Effective billable USD:

```text
billableInputUsdPer1M = costInputUsdPer1M x platformMultiplier x enterpriseMultiplier
billableOutputUsdPer1M = costOutputUsdPer1M x platformMultiplier x enterpriseMultiplier
```

For enterprise-funded models:

- platform multiplier does not participate
- enterprise config supplies `enterpriseMultiplier`

Effective billable USD:

```text
billableInputUsdPer1M = costInputUsdPer1M x enterpriseMultiplier
billableOutputUsdPer1M = costOutputUsdPer1M x enterpriseMultiplier
```

### 12.4 Credit Price Book Generation

Credit price book generation converts billable USD into runtime credit pricing:

```text
inputCreditsPer1M = ceil(billableInputUsdPer1M / creditUnitUsd)
outputCreditsPer1M = ceil(billableOutputUsdPer1M / creditUnitUsd)
```

Recommended rounding rule:

- use `ceil`
- never round down

Reason:

- avoids hidden under-recovery caused by fractional credits
- keeps runtime wallet math purely integer-based

### 12.5 Runtime Settlement Rule

At runtime, deduction uses the snapshotted price book:

```text
reservedCredits = estimate(promptTokens, completionUpperBound, priceBookSnapshot)
settledCredits = actual(promptTokens, completionTokens, priceBookSnapshot)
```

`metering` remains the source for actual token usage.
`credit` becomes the source for actual credit settlement.

## 13. Domain Model

### 13.1 `CreditWallet`

Represents one enterprise wallet under one credit type.

Suggested fields:

- `id`
- `enterpriseId`
- `creditType`
- `totalCredits`
- `reservedCredits`
- `availableCredits`
- `version`
- `createdAt`
- `updatedAt`

Rules:

- one wallet per `enterpriseId + creditType`
- `availableCredits = totalCredits - reservedCredits`
- updates must use optimistic versioning or row-level locking

### 13.2 `CreditGrant`

Represents a source batch of credits that entered a wallet.

Suggested fields:

- `id`
- `enterpriseId`
- `creditType`
- `sourceType`
- `sourceRefId`
- `grantedCredits`
- `remainingCredits`
- `effectiveAt`
- `expiresAt`
- `metadataJson`
- `createdAt`

Suggested source types:

- `trial_grant`
- `subscription_grant`
- `topup_grant`
- `manual_adjust`
- `contract_grant`

Rules:

- grants are consumption sources
- consumption should prefer earliest expiry first

### 13.3 `CreditLedgerEntry`

Represents immutable accounting movement.

Suggested fields:

- `id`
- `enterpriseId`
- `creditType`
- `walletId`
- `grantId`
- `entryType`
- `direction`
- `credits`
- `balanceAfter`
- `reservedAfter`
- `invocationId`
- `memberUserId`
- `agentId`
- `budgetScopeType`
- `budgetScopeId`
- `reasonCode`
- `snapshotJson`
- `createdAt`

Suggested entry types:

- `grant`
- `reserve`
- `settle`
- `release`
- `refund`
- `expire`
- `adjust`

Rules:

- entries are append-only
- runtime balance debugging must rely on ledger, not recalculated aggregates

### 13.4 `CreditReservation`

Represents a temporary hold before final settlement.

Suggested fields:

- `id`
- `enterpriseId`
- `creditType`
- `invocationId`
- `memberUserId`
- `agentId`
- `modelId`
- `modelScope`
- `fundingType`
- `priceBookId`
- `priceSnapshotJson`
- `reservedCredits`
- `status`
- `expiresAt`
- `createdAt`
- `updatedAt`

Suggested statuses:

- `active`
- `settled`
- `released`
- `expired`

Rules:

- one active reservation per `invocationId`
- reservation is required to stop concurrent overspend

### 13.5 `CreditBudgetPolicy`

Represents scoped usage restrictions.

Suggested fields:

- `id`
- `enterpriseId`
- `creditType`
- `scopeType`
- `scopeId`
- `enabled`
- `dailyCreditLimit`
- `monthlyCreditLimit`
- `dailyTokenLimit`
- `monthlyTokenLimit`
- `dailyUsdLimit`
- `monthlyUsdLimit`
- `hardLimit`
- `createdAt`
- `updatedAt`

Rules:

- budgets must be scoped by enterprise and credit type
- member and agent limits are not wallets
- token and USD limits remain useful, especially for enterprise-funded models

### 13.6 `CreditPriceBook`

Represents the effective credit pricing definition for runtime.

Suggested fields:

- `id`
- `enterpriseId`
- `creditType`
- `modelId`
- `modelScope`
- `modelSourceType`
- `fundingType`
- `currency`
- `creditUnitUsd`
- `costInputUsdPer1M`
- `costOutputUsdPer1M`
- `platformMultiplier`
- `enterpriseMultiplier`
- `billableInputUsdPer1M`
- `billableOutputUsdPer1M`
- `inputCreditsPer1M`
- `outputCreditsPer1M`
- `effectiveAt`
- `status`
- `createdAt`
- `updatedAt`

Rules:

- platform-funded models may have enterprise-specific price books
- enterprise-funded models are enterprise-specific by definition
- runtime must snapshot a resolved effective price book

## 14. Cross-Domain Dependency Design

The `credit` domain must depend on abstractions instead of directly coupling to storage or unrelated domain internals.

Suggested `ports.go` collaborators:

```go
type UsageFactProvider interface {
    GetUsageEventByInvocationID(invocationID string) (*UsageFact, error)
}

type ModelResolver interface {
    GetPlatformModel(id string) (*ModelRef, error)
    GetEnterpriseModel(enterpriseID, id string) (*ModelRef, error)
}

type EnterpriseResolver interface {
    GetEnterpriseByID(id string) (*EnterpriseRef, error)
    IsMemberInEnterprise(enterpriseID, userID string) (bool, error)
    AgentBelongsToEnterprise(agentID, enterpriseID string) (bool, error)
}
```

The credit domain should not call `metering` repository implementations directly.
It should consume usage facts through an abstraction owned by its service layer.

## 15. Repository Interfaces

Suggested top-level repository contracts:

```go
type WalletRepository interface {
    GetWallet(enterpriseID, creditType string) (*CreditWallet, error)
    InsertWallet(item *CreditWallet) error
    UpdateWallet(item *CreditWallet) error
}

type GrantRepository interface {
    ListConsumableGrants(enterpriseID, creditType string, now time.Time) ([]CreditGrant, error)
    GetGrantByID(id string) (*CreditGrant, error)
    InsertGrant(item *CreditGrant) error
    UpdateGrant(item *CreditGrant) error
}

type LedgerRepository interface {
    InsertEntry(item *CreditLedgerEntry) error
    ListEntries(filter LedgerFilter) ([]CreditLedgerEntry, int, error)
}

type ReservationRepository interface {
    GetActiveByInvocationID(invocationID string) (*CreditReservation, error)
    InsertReservation(item *CreditReservation) error
    UpdateReservation(item *CreditReservation) error
}

type BudgetPolicyRepository interface {
    FindPolicy(enterpriseID, creditType, scopeType, scopeID string) (*CreditBudgetPolicy, error)
    ListPolicies(enterpriseID, creditType, scopeType, scopeID string) ([]CreditBudgetPolicy, error)
    InsertPolicy(item *CreditBudgetPolicy) error
    UpdatePolicy(item *CreditBudgetPolicy) error
    DeletePolicy(id string) error
}

type PriceBookRepository interface {
    ResolveEffectivePriceBook(input ResolvePriceBookInput) (*CreditPriceBook, error)
    InsertPriceBook(item *CreditPriceBook) error
    UpdatePriceBook(item *CreditPriceBook) error
    ListPriceBooks(filter PriceBookFilter) ([]CreditPriceBook, error)
}
```

`repository_gf.go` should implement these interfaces using existing GoFrame DB patterns.

## 16. Service Layer Design

### 16.1 `WalletService`

Responsibilities:

- ensure wallet exists
- get overview
- grant credits
- apply expiry
- expose wallet balance to admin and member views

Must not do:

- HTTP parsing
- direct `g.DB()` access
- model pricing resolution

### 16.2 `BudgetService`

Responsibilities:

- validate enterprise, member, and agent budget policies
- resolve effective budget order
- return actionable over-limit errors

Recommended check order:

1. agent
2. member
3. enterprise

This gives the most actionable reason first.

### 16.3 `PricingService`

Responsibilities:

- resolve funding type
- generate or update effective price books
- validate multiplier ranges
- expose admin APIs for price book administration

### 16.4 `SettlementService`

Responsibilities:

- pre-check balance
- reserve credits
- settle reservation
- release reservation on failure
- write immutable ledger entries

This is the most critical runtime service.

### 16.5 `AdminService`

Responsibilities:

- operator adjustments
- platform grant issuance
- enterprise credit initialization
- policy and overview composition for admin pages

## 17. Runtime Integration

### 17.1 Integration with Chat Request Lifecycle

The recommended runtime sequence is:

1. resolve enterprise context
2. resolve model selection
3. resolve `modelSourceType` and `fundingType`
4. run existing `metering.CheckLimit()`
5. run `credit.BudgetService.Check()`
6. run `credit.SettlementService.PrecheckBalance()`
7. call `metering.StartInvocation()`
8. call `credit.SettlementService.Reserve()`
9. execute model request
10. call `metering.CompleteInvocation()`
11. call `credit.SettlementService.Settle(invocationId)`
12. if runtime fails, call `metering.FailInvocation()` and `credit.SettlementService.Release(invocationId)`

### 17.2 Why Reservation Happens after Invocation Creation

Reservation should use a stable `invocationId` as idempotency key.

That allows:

- repeated retry-safe reserve calls
- deterministic release and settle behavior
- aligned troubleshooting between `metering` and `credit`

### 17.3 Reservation Estimation

Phase 1 recommended estimation:

- prompt side uses known or estimated input tokens
- completion side uses agent or model output upper bound
- reservation uses a safety factor configurable at platform level

Example:

```text
reserveCredits =
  ceil(estimatedPromptTokens x inputCreditsPer1M / 1_000_000) +
  ceil(maxCompletionTokens x outputCreditsPer1M / 1_000_000)
```

### 17.4 Settlement

Settlement must use:

- actual prompt tokens from `metering`
- actual completion tokens from `metering`
- snapshotted price book from reservation

If actual is lower than reserved:

- settle actual
- release delta

If actual is higher than reserved:

- attempt additional deduction from the same wallet
- if additional deduction fails, return a typed business error and mark reservation as partially settled with alert flag

Phase 1 may initially require reservation ceiling to be conservative enough to avoid negative delta in most cases.

## 18. Existing Domain Changes Required

### 18.1 `metering` Changes

`metering` should remain the source of usage facts, but a few fields should be added to improve audit correlation.

Suggested additions to `UsageEvent`:

- `modelSourceType`
- `fundingType`
- `creditType`
- `creditPriceBookId`
- `creditUnitUsdSnapshot`
- `inputCreditsPer1MSnapshot`
- `outputCreditsPer1MSnapshot`
- `reservedCredits`
- `settledCredits`

These are audit and troubleshooting snapshots.
They do not make `metering` the owner of credit settlement logic.

### 18.2 `model` Domain Changes

Platform and enterprise model definitions should expose enough metadata for funding resolution.

Suggested additions:

- `modelSourceType`
- `fundingType`
- `baseCostInputUsdPer1M`
- `baseCostOutputUsdPer1M`

For phase 1:

- platform models default to `platform_model + platform_funded`
- enterprise custom models default to `enterprise_custom_model + enterprise_funded`

### 18.3 `enterprise` Domain Changes

The enterprise domain remains responsible for:

- enterprise identity
- membership
- admin permission
- agent ownership validation

The credit domain should query those facts through interfaces only.

## 19. Database Design

### 19.1 New Tables

Recommended new tables:

- `credit_wallets`
- `credit_grants`
- `credit_ledger_entries`
- `credit_reservations`
- `credit_budget_policies`
- `credit_price_books`

### 19.2 `credit_wallets`

Suggested columns:

- `id VARCHAR(128) PRIMARY KEY`
- `enterprise_id VARCHAR(128) NOT NULL`
- `credit_type VARCHAR(32) NOT NULL`
- `total_credits BIGINT NOT NULL DEFAULT 0`
- `reserved_credits BIGINT NOT NULL DEFAULT 0`
- `available_credits BIGINT NOT NULL DEFAULT 0`
- `version BIGINT NOT NULL DEFAULT 1`
- `created_at TIMESTAMPTZ DEFAULT NOW()`
- `updated_at TIMESTAMPTZ DEFAULT NOW()`

Indexes:

- unique on `(enterprise_id, credit_type)`

### 19.3 `credit_grants`

Suggested columns:

- `id VARCHAR(128) PRIMARY KEY`
- `enterprise_id VARCHAR(128) NOT NULL`
- `credit_type VARCHAR(32) NOT NULL`
- `source_type VARCHAR(32) NOT NULL`
- `source_ref_id VARCHAR(128) DEFAULT ''`
- `granted_credits BIGINT NOT NULL`
- `remaining_credits BIGINT NOT NULL`
- `effective_at TIMESTAMPTZ NOT NULL`
- `expires_at TIMESTAMPTZ NULL`
- `metadata_json TEXT DEFAULT ''`
- `created_at TIMESTAMPTZ DEFAULT NOW()`

Indexes:

- `(enterprise_id, credit_type, effective_at DESC)`
- `(enterprise_id, credit_type, expires_at)`

### 19.4 `credit_ledger_entries`

Suggested columns:

- `id VARCHAR(128) PRIMARY KEY`
- `enterprise_id VARCHAR(128) NOT NULL`
- `credit_type VARCHAR(32) NOT NULL`
- `wallet_id VARCHAR(128) NOT NULL`
- `grant_id VARCHAR(128) DEFAULT ''`
- `entry_type VARCHAR(32) NOT NULL`
- `direction VARCHAR(16) NOT NULL`
- `credits BIGINT NOT NULL`
- `balance_after BIGINT NOT NULL`
- `reserved_after BIGINT NOT NULL`
- `invocation_id VARCHAR(128) DEFAULT ''`
- `member_user_id VARCHAR(128) DEFAULT ''`
- `agent_id VARCHAR(128) DEFAULT ''`
- `budget_scope_type VARCHAR(32) DEFAULT ''`
- `budget_scope_id VARCHAR(128) DEFAULT ''`
- `reason_code VARCHAR(64) DEFAULT ''`
- `snapshot_json TEXT DEFAULT ''`
- `created_at TIMESTAMPTZ DEFAULT NOW()`

Indexes:

- `(enterprise_id, credit_type, created_at DESC)`
- `(invocation_id, created_at DESC)`

### 19.5 `credit_reservations`

Suggested columns:

- `id VARCHAR(128) PRIMARY KEY`
- `enterprise_id VARCHAR(128) NOT NULL`
- `credit_type VARCHAR(32) NOT NULL`
- `invocation_id VARCHAR(128) NOT NULL`
- `member_user_id VARCHAR(128) DEFAULT ''`
- `agent_id VARCHAR(128) DEFAULT ''`
- `model_id VARCHAR(128) NOT NULL`
- `model_scope VARCHAR(32) DEFAULT ''`
- `funding_type VARCHAR(32) NOT NULL`
- `price_book_id VARCHAR(128) NOT NULL`
- `price_snapshot_json TEXT NOT NULL`
- `reserved_credits BIGINT NOT NULL`
- `status VARCHAR(32) NOT NULL`
- `expires_at TIMESTAMPTZ NULL`
- `created_at TIMESTAMPTZ DEFAULT NOW()`
- `updated_at TIMESTAMPTZ DEFAULT NOW()`

Indexes:

- unique on `(invocation_id)`
- `(enterprise_id, credit_type, status, created_at DESC)`

### 19.6 `credit_budget_policies`

Suggested columns:

- `id VARCHAR(128) PRIMARY KEY`
- `enterprise_id VARCHAR(128) NOT NULL`
- `credit_type VARCHAR(32) NOT NULL`
- `scope_type VARCHAR(32) NOT NULL`
- `scope_id VARCHAR(128) NOT NULL`
- `enabled BOOLEAN DEFAULT TRUE`
- `daily_credit_limit BIGINT DEFAULT 0`
- `monthly_credit_limit BIGINT DEFAULT 0`
- `daily_token_limit BIGINT DEFAULT 0`
- `monthly_token_limit BIGINT DEFAULT 0`
- `daily_usd_limit NUMERIC(20, 8) DEFAULT 0`
- `monthly_usd_limit NUMERIC(20, 8) DEFAULT 0`
- `hard_limit BOOLEAN DEFAULT TRUE`
- `created_at TIMESTAMPTZ DEFAULT NOW()`
- `updated_at TIMESTAMPTZ DEFAULT NOW()`

Indexes:

- unique on `(enterprise_id, credit_type, scope_type, scope_id)`

### 19.7 `credit_price_books`

Suggested columns:

- `id VARCHAR(128) PRIMARY KEY`
- `enterprise_id VARCHAR(128) DEFAULT ''`
- `credit_type VARCHAR(32) NOT NULL`
- `model_id VARCHAR(128) NOT NULL`
- `model_scope VARCHAR(32) NOT NULL`
- `model_source_type VARCHAR(32) NOT NULL`
- `funding_type VARCHAR(32) NOT NULL`
- `currency VARCHAR(16) DEFAULT 'USD'`
- `credit_unit_usd NUMERIC(20, 8) NOT NULL`
- `cost_input_usd_per_1m NUMERIC(20, 8) NOT NULL`
- `cost_output_usd_per_1m NUMERIC(20, 8) NOT NULL`
- `platform_multiplier NUMERIC(20, 8) DEFAULT 1`
- `enterprise_multiplier NUMERIC(20, 8) DEFAULT 1`
- `billable_input_usd_per_1m NUMERIC(20, 8) NOT NULL`
- `billable_output_usd_per_1m NUMERIC(20, 8) NOT NULL`
- `input_credits_per_1m BIGINT NOT NULL`
- `output_credits_per_1m BIGINT NOT NULL`
- `effective_at TIMESTAMPTZ NOT NULL`
- `status VARCHAR(32) NOT NULL`
- `created_at TIMESTAMPTZ DEFAULT NOW()`
- `updated_at TIMESTAMPTZ DEFAULT NOW()`

Indexes:

- `(credit_type, model_id, model_scope, effective_at DESC)`
- `(enterprise_id, credit_type, model_id, effective_at DESC)`

### 19.8 Existing Table Extensions

Recommended additions to `llm_usage_events`:

- `model_source_type`
- `funding_type`
- `credit_type`
- `credit_price_book_id`
- `credit_unit_usd_snapshot`
- `input_credits_per_1m_snapshot`
- `output_credits_per_1m_snapshot`
- `reserved_credits`
- `settled_credits`

## 20. API Design

### 20.1 Platform Admin Routes

Recommended route family:

- `GET /api/admin/platform/credit-config`
- `PUT /api/admin/platform/credit-config`
- `GET /api/admin/platform/credit-price-books`
- `POST /api/admin/platform/credit-price-books`
- `PUT /api/admin/platform/credit-price-books/{id}`
- `GET /api/admin/platform/credit-wallets`
- `POST /api/admin/platform/credit-grants`

Responsibilities:

- define global credit unit and reservation safety defaults
- manage platform-funded price books
- issue platform credit grants to enterprises

### 20.2 Enterprise Admin Routes

Recommended route family:

- `GET /api/admin/credit/overview`
- `GET /api/admin/credit/wallets`
- `GET /api/admin/credit/ledger`
- `GET /api/admin/credit/grants`
- `GET /api/admin/credit-price-books`
- `POST /api/admin/credit-price-books`
- `PUT /api/admin/credit-price-books/{id}`
- `GET /api/admin/credit-budget-policies`
- `POST /api/admin/credit-budget-policies`
- `PUT /api/admin/credit-budget-policies/{id}`
- `DELETE /api/admin/credit-budget-policies/{id}`
- `POST /api/admin/credit-grants/manual-adjust`

Responsibilities:

- inspect enterprise wallet state
- manage enterprise-funded price books
- manage enterprise, member, and agent budget policies
- perform manual enterprise credit adjustments under authorization

### 20.3 Internal Runtime Routes

Phase 1 can keep settlement as in-process service calls.

If an internal API is later needed, recommended route family:

- `POST /internal/credits/reservations`
- `POST /internal/credits/reservations/{invocationId}/settle`
- `POST /internal/credits/reservations/{invocationId}/release`

## 21. Error Model

Suggested typed business errors:

- `ErrWalletNotFound`
- `ErrWalletInsufficientCredits`
- `ErrCreditReservationNotFound`
- `ErrCreditReservationAlreadySettled`
- `ErrBudgetExceeded`
- `ErrCreditTypeMismatch`
- `ErrPriceBookNotFound`
- `ErrFundingTypeMismatch`
- `ErrInvalidPriceBook`
- `ErrGrantExpired`

Handlers should translate them into stable HTTP errors without leaking storage details.

## 22. Concurrency and Idempotency

### 22.1 Wallet Updates

Wallet update operations must be concurrency-safe.

Recommended options:

- optimistic locking through `version`
- or row-level locking inside transaction

Phase 1 may prefer row-level locking for clarity and lower risk.

### 22.2 Reservation Idempotency

`invocationId` must be the runtime idempotency key.

This applies to:

- reserve
- settle
- release

Repeated calls for the same lifecycle transition must be safe.

### 22.3 Ledger Integrity

Ledger must be append-only.

Never recompute or rewrite historical ledger rows when price book settings change later.

## 23. Migration Strategy

### Phase 1: Domain Introduction

- add `credit` domain with wallet, grant, ledger, reservation, budget, and price book tables
- keep existing `metering` behavior unchanged
- add enterprise wallet bootstrap logic

### Phase 2: Platform-Funded Runtime Integration

- route platform-funded model requests through credit reservation and settlement
- add `platform` credit grants and admin overview
- add metering snapshots for credit linkage

### Phase 3: Enterprise-Funded Runtime Integration

- enable enterprise custom models to use `enterprise` credit type
- add enterprise-specific price books and enterprise credit grants
- unify enterprise admin credit pages

### Phase 4: Subscription and Payment Outer Layer

- build separate payment or billing layer on top of grants
- keep `credit` as the settlement core

## 24. Testing Strategy

The implementation should include focused unit tests for:

- price book generation
- budget resolution order
- wallet concurrency-safe reserve and settle
- grant consumption order by earliest expiry
- repeated reserve or settle idempotency
- runtime settle with refund delta
- platform-funded and enterprise-funded routing

Recommended integration tests:

- chat invocation with platform-funded model deducts `platform` credits from enterprise wallet
- chat invocation with enterprise custom model deducts `enterprise` credits from enterprise wallet
- member and agent limits block request before settlement
- failed invocation releases reservation

## 25. Recommended Implementation Decisions

To keep phase 1 clean and aligned with current product direction, the following decisions are recommended as fixed defaults:

- create one `platform` wallet and one `enterprise` wallet per enterprise on demand
- define enterprise as the only wallet owner
- keep member and agent as budget scopes only
- keep `metering` as operational usage fact and USD reporting domain
- keep `credit` as settlement and balance domain
- base runtime deduction on snapshotted credit price books
- use conservative reservation and actual settlement with delta release

## 26. Final Architecture Summary

The target architecture is:

- `model` chooses which model is used
- `metering` records what was consumed in tokens and USD terms
- `credit` decides which enterprise wallet to charge, reserves credits, and settles immutable ledger entries
- `enterprise` provides tenant, membership, and agent ownership context

This gives DotBlue one unified, clean, and extensible credit core:

- platform-funded and enterprise-funded models share one engine
- enterprise remains the stable accounting owner
- member and agent remain governable
- future subscription, top-up, and payment features can build on top without redesigning runtime settlement
