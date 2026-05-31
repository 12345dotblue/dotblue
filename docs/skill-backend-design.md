# DotBlue Skill Backend Design

## 1. Scope

This document converts the skill system architecture and database design into a backend implementation design for DotBlue.

It focuses on:

- domain placement
- package layout
- handler, service, and repository responsibilities
- core interfaces
- API grouping and request models
- cross-domain dependencies
- runtime integration points
- implementation phases

It is designed to match the current GoFrame-based backend style and the local engineering constraints.

## 2. Existing Backend Style to Follow

The current backend already uses a clean business-domain structure:

- business code lives under `backend/internal/domains`
- handlers adapt HTTP protocol and user context
- services hold business rules
- repositories abstract persistence
- current routes are grouped by platform admin, enterprise admin, and normal member scopes

This style should be preserved for the new skill system.

Relevant existing references:

- [agent service](file:///c:/Users/kongz/work/dotblue/backend/internal/domains/agent/service.go)
- [agent handler](file:///c:/Users/kongz/work/dotblue/backend/internal/domains/agent/handler.go)
- [agent repository](file:///c:/Users/kongz/work/dotblue/backend/internal/domains/agent/repository.go)
- [route registration](file:///c:/Users/kongz/work/dotblue/backend/internal/cmd/cmd.go#L83-L185)

## 3. Design Principles

- keep `skill` as one business domain
- keep HTTP and identity concerns in handlers only
- keep policy, lifecycle, and dependency rules in services
- keep SQL and storage details in repositories
- depend on interfaces for runtime, approval, hub adapters, and audit
- avoid letting source-specific logic leak into product-facing service contracts

## 4. Domain Placement

The skill system should be implemented under:

`backend/internal/domains/skill`

Suggested initial files:

```text
backend/internal/domains/skill/
  skill.go
  handler.go
  service.go
  repository.go
  repository_gf.go
  policy.go
  manifest.go
  hub.go
  import_job.go
  release.go
  execution.go
  audit.go
  runtime.go
  errors.go
```

If the domain grows further, careful internal subpackages may be introduced. Phase 1 should still prefer a compact domain boundary.

## 5. Package Responsibilities

## 5.1 `handler.go`

Must contain:

- route handlers
- request parsing
- path and query extraction
- authentication and enterprise context extraction
- response mapping
- error-to-HTTP mapping

Must not contain:

- lifecycle rules
- graph validation rules
- policy merge logic
- SQL details
- hub parsing logic

## 5.2 `service.go`

Must contain:

- skill creation logic
- publish and disable rules
- enterprise enablement rules
- agent installation rules
- version resolution logic
- policy merge logic
- dependency validation
- cross-entity state validation

Must not contain:

- direct SQL or GoFrame DB calls
- HTTP request or response code
- source-specific transport implementation

## 5.3 `repository.go`

Must define storage-facing abstractions for:

- skill identity
- version persistence
- reference queries
- enablement persistence
- agent binding persistence
- execution persistence
- audit persistence
- hub and import job persistence

## 5.4 `repository_gf.go`

Must implement repository interfaces using current GoFrame database conventions.

## 5.5 `runtime.go`

Should define runtime-facing abstractions:

- skill resolution
- invocation dispatch
- downstream reference traversal
- execution recording
- audit recording

## 5.6 `policy.go`

Should centralize:

- policy structures
- effective policy merge order
- validation helpers
- approval and risk constants

## 5.7 `errors.go`

Should centralize typed business errors for consistent handler behavior.

Recommended examples:

- `ErrSkillNotFound`
- `ErrSkillVersionNotFound`
- `ErrSkillNotPublished`
- `ErrSkillNotEnabled`
- `ErrSkillInstallDenied`
- `ErrSkillCycleDetected`
- `ErrSkillApprovalRequired`
- `ErrSkillTrustDenied`

## 6. Core Domain Model in Go

## 6.1 Aggregate Overview

The skill domain should model these business entities:

- `Skill`
- `SkillVersion`
- `SkillReference`
- `SkillHub`
- `SkillImportJob`
- `EnterpriseSkillEnablement`
- `AgentSkillBinding`
- `SkillExecution`
- `SkillAudit`
- `SkillReleaseRecord`

## 6.2 Suggested Go Types

High-level examples:

```go
type Skill struct {
    Id                       string    `json:"id"`
    Code                     string    `json:"code"`
    Name                     string    `json:"name"`
    Description              string    `json:"description"`
    OwnerScope               string    `json:"ownerScope"`
    OwnerEnterpriseId        string    `json:"ownerEnterpriseId"`
    SourceType               string    `json:"sourceType"`
    ProviderType             string    `json:"providerType"`
    TrustLevel               string    `json:"trustLevel"`
    Status                   string    `json:"status"`
    LatestVersionId          string    `json:"latestVersionId"`
    LatestPublishedVersionId string    `json:"latestPublishedVersionId"`
    LatestStableVersionId    string    `json:"latestStableVersionId"`
    TagsJson                 string    `json:"-" orm:"tags_json"`
    MetadataJson             string    `json:"-" orm:"metadata_json"`
    CreatedBy                string    `json:"createdBy"`
    UpdatedBy                string    `json:"updatedBy"`
    CreatedAt                time.Time `json:"createdAt"`
    UpdatedAt                time.Time `json:"updatedAt"`
}
```

`JSONB` columns can remain stored as strings or mapped into strongly typed structs through explicit conversion helpers. The recommended rule is:

- store raw JSON payload in repository models
- convert into domain structs inside service layer only when business logic needs structured access

## 7. Service Layer Design

## 7.1 Top-Level Service

Suggested service shape:

```go
type Service struct {
    repo            Repository
    policyResolver  PolicyResolver
    graphValidator  GraphValidator
    runtimeRegistry RuntimeRegistry
    hubRegistry     HubRegistry
    clock           func() time.Time
    idGenerator     func() string
}
```

Reason:

- the domain service stays orchestration-heavy
- the domain depends on abstractions
- source-specific behavior is delegated to injected collaborators

## 7.2 Main Service Responsibilities

The service should expose methods for:

- create skill
- create skill version
- submit review
- publish version
- disable version
- rollback release
- list skills
- get skill detail
- create or update skill hub
- import skill from hub
- enable skill for enterprise
- disable skill for enterprise
- install skill on agent
- uninstall skill from agent
- list installed skills for agent
- record execution start and finish
- list executions and audits

## 7.3 Suggested Service Interface

```go
type ApplicationService interface {
    CreateSkill(ctx context.Context, actor ActorContext, input CreateSkillInput) (*Skill, error)
    CreateSkillVersion(ctx context.Context, actor ActorContext, input CreateSkillVersionInput) (*SkillVersion, error)
    SubmitSkillReview(ctx context.Context, actor ActorContext, skillId string, input SubmitReviewInput) error
    PublishSkillVersion(ctx context.Context, actor ActorContext, skillId string, input PublishSkillInput) error
    DisableSkillVersion(ctx context.Context, actor ActorContext, skillId string, input DisableSkillInput) error
    RollbackSkillRelease(ctx context.Context, actor ActorContext, skillId string, input RollbackSkillInput) error

    ListSkills(ctx context.Context, actor ActorContext, filter SkillFilter) ([]SkillSummary, error)
    GetSkillDetail(ctx context.Context, actor ActorContext, skillId string) (*SkillDetail, error)

    UpsertHub(ctx context.Context, actor ActorContext, input UpsertHubInput) (*SkillHub, error)
    ImportSkill(ctx context.Context, actor ActorContext, input ImportSkillInput) (*SkillImportJob, error)

    EnableSkillForEnterprise(ctx context.Context, actor ActorContext, enterpriseId, skillId string, input EnableSkillInput) error
    DisableSkillForEnterprise(ctx context.Context, actor ActorContext, enterpriseId, skillId string) error

    InstallSkillOnAgent(ctx context.Context, actor ActorContext, agentId string, input InstallSkillInput) (*AgentSkillBinding, error)
    UninstallSkillFromAgent(ctx context.Context, actor ActorContext, agentId, skillId string) error
    ListAgentSkills(ctx context.Context, actor ActorContext, agentId string) ([]AgentSkillBindingView, error)

    ListExecutions(ctx context.Context, actor ActorContext, filter ExecutionFilter) ([]SkillExecutionView, error)
    ListAudits(ctx context.Context, actor ActorContext, filter AuditFilter) ([]SkillAuditView, error)
}
```

## 8. Repository Design

## 8.1 Repository Boundary

The repository should own:

- persistence queries
- transactional write ordering
- row-to-struct mapping

The repository should not own:

- lifecycle legality
- approval decisions
- cycle detection
- policy resolution

## 8.2 Suggested Repository Interface

```go
type Repository interface {
    WithTx(ctx context.Context, fn func(ctx context.Context, repo Repository) error) error

    CreateSkill(ctx context.Context, item *Skill) error
    UpdateSkill(ctx context.Context, item *Skill) error
    GetSkillById(ctx context.Context, id string) (*Skill, error)
    GetSkillByCode(ctx context.Context, ownerScope, ownerEnterpriseId, code string) (*Skill, error)
    ListSkills(ctx context.Context, filter SkillFilter) ([]*Skill, error)

    CreateSkillVersion(ctx context.Context, item *SkillVersion) error
    UpdateSkillVersion(ctx context.Context, item *SkillVersion) error
    GetSkillVersionById(ctx context.Context, id string) (*SkillVersion, error)
    GetSkillVersion(ctx context.Context, skillId, version string) (*SkillVersion, error)
    ListSkillVersions(ctx context.Context, skillId string) ([]*SkillVersion, error)

    CreateSkillReference(ctx context.Context, item *SkillReference) error
    DeleteSkillReference(ctx context.Context, fromVersionId, toVersionId string) error
    ListOutgoingReferences(ctx context.Context, fromVersionId string) ([]*SkillReference, error)
    ListIncomingReferences(ctx context.Context, toVersionId string) ([]*SkillReference, error)

    UpsertHub(ctx context.Context, item *SkillHub) error
    GetHubById(ctx context.Context, id string) (*SkillHub, error)
    ListHubs(ctx context.Context, filter HubFilter) ([]*SkillHub, error)

    CreateImportJob(ctx context.Context, item *SkillImportJob) error
    UpdateImportJob(ctx context.Context, item *SkillImportJob) error
    GetImportJobById(ctx context.Context, id string) (*SkillImportJob, error)

    UpsertEnterpriseEnablement(ctx context.Context, item *EnterpriseSkillEnablement) error
    GetEnterpriseEnablement(ctx context.Context, enterpriseId, skillId string) (*EnterpriseSkillEnablement, error)
    ListEnterpriseEnablements(ctx context.Context, filter EnterpriseEnablementFilter) ([]*EnterpriseSkillEnablement, error)

    UpsertAgentSkillBinding(ctx context.Context, item *AgentSkillBinding) error
    GetAgentSkillBinding(ctx context.Context, agentId, skillId string) (*AgentSkillBinding, error)
    ListAgentSkillBindings(ctx context.Context, filter AgentBindingFilter) ([]*AgentSkillBinding, error)
    DeleteAgentSkillBinding(ctx context.Context, agentId, skillId string) error

    CreateSkillExecution(ctx context.Context, item *SkillExecution) error
    UpdateSkillExecution(ctx context.Context, item *SkillExecution) error
    ListSkillExecutions(ctx context.Context, filter ExecutionFilter) ([]*SkillExecution, error)

    CreateSkillAudit(ctx context.Context, item *SkillAudit) error
    ListSkillAudits(ctx context.Context, filter AuditFilter) ([]*SkillAudit, error)

    CreateSkillReleaseRecord(ctx context.Context, item *SkillReleaseRecord) error
    ListSkillReleaseRecords(ctx context.Context, skillId string) ([]*SkillReleaseRecord, error)
}
```

## 8.3 Transaction Guidance

These workflows should run inside repository-managed transactions:

- create skill + initial version
- publish version + update skill pointers + insert release record
- enable skill for enterprise
- install skill on agent
- execution completion updates when multiple rows must stay consistent

## 9. Supporting Interfaces

## 9.1 Policy Resolver

```go
type PolicyResolver interface {
    Resolve(ctx context.Context, input PolicyResolutionInput) (*EffectiveSkillPolicy, error)
}
```

Responsibilities:

- merge platform defaults
- merge skill version defaults
- merge enterprise overrides
- merge binding overrides
- apply runtime restrictions such as channel and approval context

## 9.2 Graph Validator

```go
type GraphValidator interface {
    ValidateReferenceAddition(ctx context.Context, fromVersionId, toVersionId string) error
    ValidateVersionGraph(ctx context.Context, rootVersionId string) error
}
```

Responsibilities:

- detect self-reference
- detect cycles
- enforce max depth and optional fan-out limits

## 9.3 Hub Registry

```go
type HubRegistry interface {
    GetAdapter(hubType string) (HubAdapter, error)
}
```

## 9.4 Hub Adapter

```go
type HubAdapter interface {
    Parse(ctx context.Context, hub SkillHub, sourceLocator string) (*ParsedExternalSkill, error)
    Normalize(ctx context.Context, parsed *ParsedExternalSkill) (*NormalizedSkillDraft, error)
    Verify(ctx context.Context, normalized *NormalizedSkillDraft) (*VerificationReport, error)
}
```

## 9.5 Runtime Registry

```go
type RuntimeRegistry interface {
    GetInvoker(providerType string) (SkillInvoker, error)
}
```

## 9.6 Skill Invoker

```go
type SkillInvoker interface {
    Invoke(ctx context.Context, req InvokeRequest) (*InvokeResult, error)
}
```

## 10. Actor and Scope Model

## 10.1 Why Actor Context Should Be Explicit

The current system derives user and enterprise information from middleware. Skill service methods should not depend on HTTP request objects directly.

Introduce a transport-neutral actor object:

```go
type ActorContext struct {
    UserId       string
    EnterpriseId string
    EnterpriseRole string
    IsPlatformAdmin bool
}
```

Handlers build it from:

- `identity.GetUserId(r)`
- `identity.GetCurrentEnterpriseId(r)`
- `identity.GetCurrentEnterpriseRole(r)`
- `identity.IsAdmin(r)`

This keeps service logic reusable outside HTTP if needed later.

## 11. Handler Design

## 11.1 Route Grouping

The current backend already uses three route scopes:

- platform admin
- enterprise admin
- enterprise member

The skill domain should follow the same route grouping in `cmd.go`.

## 11.2 Platform Admin Routes

Purpose:

- hub management
- platform-wide skill catalog
- review and publish
- execution and audit inspection

Suggested routes:

- `GET /api/admin/skills?view=governance`
- `POST /api/admin/skills`
- `GET /api/admin/skills/{id}`
- `POST /api/admin/skills/{id}/versions`
- `POST /api/admin/skills/{id}/submit-review`
- `POST /api/admin/skills/{id}/publish`
- `POST /api/admin/skills/{id}/disable`
- `POST /api/admin/skills/{id}/rollback`
- `GET /api/admin/platform/skill-hubs`
- `POST /api/admin/platform/skill-hubs`
- `PUT /api/admin/platform/skill-hubs/{id}`
- `POST /api/admin/platform/skill-import-jobs`
- `GET /api/admin/platform/skill-import-jobs/{id}`
- `GET /api/admin/platform/skill-executions`
- `GET /api/admin/platform/skill-audits`

## 11.3 Enterprise Admin Routes

Purpose:

- enterprise enablement
- enterprise policy overrides
- agent installation

Suggested routes:

- `GET /api/admin/skills`
- `POST /api/admin/skills/{skillId}/enable`
- `POST /api/admin/skills/{skillId}/disable`
- `GET /api/admin/agents/{agentId}/skills`
- `POST /api/admin/agents/{agentId}/skills/install`
- `POST /api/admin/agents/{agentId}/skills/{skillId}/uninstall`

## 11.4 Member Routes

Phase 1 should keep member routes minimal.

Suggested route:

- `GET /api/agents/{agentId}/skills`

This supports future UI rendering and self-visible capability display without granting management rights.

## 11.5 Handler Request Types

Handlers should define compact request structs close to routes, following current style.

Example:

```go
type createSkillReq struct {
    Code        string `json:"code" v:"required"`
    Name        string `json:"name" v:"required"`
    Description string `json:"description"`
    OwnerScope  string `json:"ownerScope" v:"required"`
    SourceType  string `json:"sourceType" v:"required"`
}
```

Example:

```go
type installSkillReq struct {
    SkillId         string `json:"skillId" v:"required"`
    SkillVersionId  string `json:"skillVersionId" v:"required"`
    EntryAlias      string `json:"entryAlias"`
    InvokeVisibility string `json:"invokeVisibility"`
}
```

## 12. API Request and Response Models

## 12.1 Create Skill

Request:

- `code`
- `name`
- `description`
- `ownerScope`
- `sourceType`
- `providerType`
- `tags`
- `metadata`

Response:

- `id`
- `code`
- `name`
- `status`
- `sourceType`
- `trustLevel`
- `createdAt`

## 12.2 Create Skill Version

Request:

- `version`
- `manifest`
- `inputSchema`
- `outputSchema`
- `defaultPolicy`
- `runtimeContract`
- `references`
- `changeLog`

Response:

- version summary
- validation result summary

## 12.3 Publish Skill

Request:

- `skillVersionId`
- `releaseChannel`
- `note`

Response:

- publish result
- updated latest published pointers

## 12.4 Enable Skill for Enterprise

Request:

- `orgScope`
- `channelScope`
- `policyOverride`
- `reviewNote`

Response:

- enablement summary

## 12.5 Install Skill on Agent

Request:

- `skillId`
- `skillVersionId`
- `entryAlias`
- `invokeVisibility`
- `policyOverride`
- `channelScope`

Response:

- binding summary
- effective policy summary

## 12.6 List Executions

Query filters:

- `enterpriseId` for platform admin only if needed
- `agentId`
- `skillId`
- `conversationId`
- `traceId`
- `status`
- `startTime`
- `endTime`
- `page`
- `pageSize`

## 13. Service Workflow Design

## 13.1 Create Skill

Workflow:

1. validate actor permission
2. normalize code and ownership scope
3. check uniqueness in repository
4. create `Skill`
5. persist and return

## 13.2 Create Skill Version

Workflow:

1. validate actor permission
2. load skill
3. validate version uniqueness
4. validate manifest and schemas
5. validate declared references
6. persist version
7. persist references if included
8. update latest version pointer on `Skill`

## 13.3 Publish Version

Workflow:

1. validate actor permission
2. load skill and target version
3. validate release status transition
4. run graph validation
5. validate trust and verification reports
6. update version release status
7. update latest published or stable pointers
8. insert release record

## 13.4 Enable Skill for Enterprise

Workflow:

1. validate enterprise admin permission
2. load skill and latest published version or selected version policy basis
3. validate skill is publishable and not blocked
4. upsert enterprise enablement

## 13.5 Install Skill on Agent

Workflow:

1. validate enterprise admin permission
2. load agent
3. verify `actor.EnterpriseId == agent.GroupId`
4. load enterprise enablement
5. validate selected version is installable
6. validate dependency graph
7. resolve effective policy preview
8. upsert binding

## 13.6 Execute Skill

This workflow is mostly runtime-facing but must still be owned by the skill domain service.

Workflow:

1. resolve binding by agent and skill
2. resolve effective policy
3. validate approval and channel
4. create execution row
5. dispatch through invoker
6. update execution result
7. create audit rows

## 14. Cross-Domain Dependencies

## 14.1 `identity`

Used by handlers to build `ActorContext`.

No skill service should import `ghttp.Request`.

## 14.2 `enterprise`

Used to validate:

- enterprise membership
- admin role
- org scope if later needed

In phase 1, route middleware can handle most access gating. Service still needs enterprise consistency checks.

## 14.3 `agent`

Used during installation and execution:

- load agent
- verify enterprise ownership
- eventually expose installed skills in agent detail APIs

This dependency should remain read-oriented. Skill domain should not mutate core agent profile fields.

## 14.4 `chat`

Used to replace or supplement current runtime tool event display with explicit skill execution records.

Recommended integration approach:

- keep `chat` as conversation orchestration
- inject skill execution timeline from `skill` domain for rendering and tracing

## 14.5 `im`

Used for channel-aware policy enforcement.

Skill domain should not reimplement IM routing. It only needs:

- channel identifier
- connection context if necessary
- approval restrictions for external messaging

## 14.6 `settings`

Used for platform default policy, release defaults, and hub-related global config if retained there.

## 14.7 `execution`

The repository already has an `execution` domain for worker concerns. The skill runtime design should not duplicate worker orchestration semantics. Instead:

- worker infrastructure stays where it is
- skill domain owns business execution records and policy decisions

## 15. Runtime Integration Design

## 15.1 Runtime Entry Object

Suggested transport-neutral invocation request:

```go
type InvokeRequest struct {
    TraceId         string
    EnterpriseId    string
    AgentId         string
    SkillId         string
    SkillVersionId  string
    ConversationId  string
    MessageId       string
    Channel         string
    ActorUserId     string
    ActorType       string
    Input           map[string]any
    CallPath        []string
    Depth           int
}
```

## 15.2 Runtime Output

```go
type InvokeResult struct {
    Output        map[string]any
    OutputDigest  string
    Metrics       map[string]any
    ChildRequests []InvokeRequest
}
```

This lets the service keep child-skill execution explicit instead of hiding it in source-specific implementations.

## 15.3 Execution Recording Pattern

Recommended pattern:

- service creates execution row before invoke
- invoker returns business result and metrics
- service writes final execution status
- service writes audit rows

This keeps audit consistency under domain control.

## 16. Error Handling Design

## 16.1 Business Error Mapping

Handlers should convert typed errors into stable HTTP responses.

Suggested mapping:

- `ErrSkillNotFound` -> `404`
- `ErrSkillVersionNotFound` -> `404`
- `ErrSkillCycleDetected` -> `400`
- `ErrSkillNotPublished` -> `400`
- `ErrSkillApprovalRequired` -> `403`
- `ErrSkillTrustDenied` -> `403`
- `ErrSkillInstallDenied` -> `400`
- unknown errors -> `500`

## 16.2 Error Message Style

Prefer stable short messages for API clients and richer internal logs for diagnostics.

## 17. Route Registration Plan

The `cmd.go` file should be extended in the same style as current admin groups.

### Platform Admin Group

Inside the existing platform admin router group:

- register platform skill management routes
- register hub management routes
- register execution and audit query routes

### Enterprise Admin Group

Inside the existing enterprise admin router group:

- register enablement routes
- register agent skill install routes

### Member Group

Inside the current member router group:

- register read-only agent skill listing if needed

## 18. Testing Plan for Backend Implementation

## 18.1 Service Tests

High-value service tests:

- create skill uniqueness
- create version validation
- publish lifecycle rules
- cycle detection at publish
- enterprise enablement rules
- install skill enterprise mismatch denial
- install denied for unpublished or blocked version
- effective policy merge behavior

## 18.2 Repository Tests

Repository tests should focus on:

- filter correctness
- upsert behavior
- transaction behavior
- uniqueness enforcement

## 18.3 Handler Tests

Handler tests should focus on:

- permission scope
- request validation
- HTTP status mapping

## 19. Implementation Phase Breakdown

## 19.1 Phase 1

Deliver:

- core structs
- repository interfaces
- repository GoFrame implementation
- create/list/publish skill APIs
- enterprise enablement APIs
- agent install APIs

## 19.2 Phase 2

Deliver:

- hub registry and import jobs
- OpenAPI and MCP adapter interfaces
- import pipeline service

## 19.3 Phase 3

Deliver:

- runtime invocation service
- execution and audit APIs
- chat integration for execution timeline

## 20. Recommended First Code Tasks

The most practical first engineering tasks are:

1. add `backend/internal/domains/skill` with structs and constants
2. add repository interface and `repository_gf.go`
3. add service methods for create, list, publish, enable, install
4. register phase 1 routes in `cmd.go`
5. add tests for lifecycle and installation rules

## 21. Summary

The backend implementation should make `skill` a first-class business domain that owns:

- capability identity
- lifecycle and release
- enterprise enablement
- agent installation
- runtime execution record
- audit record

The most important implementation choice is to keep the domain clean:

- handlers only adapt transport
- services own business rules
- repositories own persistence
- adapters hide source-specific details

This gives DotBlue a maintainable backend foundation for a real enterprise skill control plane without breaking its current domain-driven structure.
