# Skill Platform / Enterprise Layering Proposal

## 1. Why This Needs Discussion

The current product direction has already made one thing clear:

- platform admins manage platform-level skills
- enterprise admins should also manage enterprise-level skills
- both kinds of skills ultimately need to be enabled, installed, governed, and audited

The current backend does not model this as two separate management systems.
Instead, it models one unified `skill` domain with:

- platform catalog lifecycle
- tenant availability
- agent installation

This is acceptable for phase 1, but it will become ambiguous once "enterprise-owned skills" become real rather than conceptual.

## 2. Current Backend Reality

The current implementation is organized as one domain:

- `backend/internal/domains/skill/skill.go`
- `backend/internal/domains/skill/service.go`
- `backend/internal/domains/skill/repository.go`
- `backend/internal/domains/skill/handler.go`

It already distinguishes actor scope at the route layer:

- platform admin governance routes: `/api/admin/skills?view=governance` plus `/api/admin/skills/{id}...`
- enterprise admin routes: `/api/admin/skills...`
- agent install routes: `/api/admin/agents/{agentId}/skills...`

But inside the service and repository layers, all of these responsibilities are still handled together.

Current model facts:

- `Skill` already has `ownerScope` and `ownerEnterpriseId`
- `CreateSkill()` currently forces `ownerScope = platform`
- enterprise admin `/api/admin/skills` is effectively a projection of platform-published skills plus enterprise availability state
- enterprise-owned skill master data is not truly implemented yet

## 3. Main Design Problem

The real problem is not "one domain vs two domains".

The real problem is:

- the route layer already speaks in two business languages:
  - platform skill governance
  - enterprise skill governance
- but the service and repository layers still expose one flat capability surface

That causes three risks:

- semantic ambiguity:
  - `/api/admin/skills` currently means "enterprise availability control for platform skills"
  - but product-wise it will soon mean "enterprise skill management"
- permission drift:
  - route middleware enforces scope, but service methods do not clearly express platform-only vs enterprise-only intent
- model collision:
  - once enterprise-private skills, enterprise imports, or enterprise forks appear, current query and repository boundaries will become hard to reason about

## 4. Recommendation

### 4.1 Do Not Split Into Two Top-Level Domains Yet

Do **not** immediately split into:

- `domains/platform_skill`
- `domains/enterprise_skill`

That would duplicate too much logic:

- skill identity and version model
- reference graph validation
- publish lifecycle
- manifest normalization
- trust and policy structures
- hub import pipeline

These are still one bounded context: `skill`.

### 4.2 Do Split the `skill` Domain Internally

Keep one domain boundary:

- `backend/internal/domains/skill`

But refactor it into internal capability modules so platform and enterprise concerns stop sharing one flat service surface.

Recommended internal layering:

```text
backend/internal/domains/skill/
  skill.go
  errors.go
  handler_platform.go
  handler_enterprise.go
  handler_agent.go
  service_catalog.go
  service_lifecycle.go
  service_availability.go
  service_binding.go
  service_hub.go
  repository_catalog.go
  repository_availability.go
  repository_binding.go
  repository_hub.go
  repository_gf.go
  policy.go
  manifest.go
```

This keeps one domain while making intent explicit.

## 5. Target Business Model

The backend should explicitly support these three layers.

### 5.1 Platform Skill Catalog

Owned by platform admins.

Responsibilities:

- create platform skills
- create versions
- submit review
- publish
- import from platform-managed hubs
- publish to enterprise consumers

Data shape:

- `ownerScope = platform`
- `ownerEnterpriseId = ""`

### 5.2 Enterprise Skill Catalog

Owned by enterprise admins.

Responsibilities:

- create enterprise-private skills
- import enterprise-private skills
- optionally fork or wrap platform skills
- manage enterprise-only versions and policies

Data shape:

- `ownerScope = enterprise`
- `ownerEnterpriseId = <enterprise_id>`

### 5.3 Skill Availability / Agent Binding

Enterprise admins and agent admins consume available skills.

Responsibilities:

- enable a skill for an enterprise
- configure org/channel/policy overrides
- install enabled skills onto agents

This layer should work for both:

- platform-owned skills published to the enterprise
- enterprise-owned skills created by the enterprise itself

## 6. API Direction

### 6.1 Recommended Route Shape

Platform admin:

- `GET /api/admin/skills?view=governance`
- `POST /api/admin/skills`
- `POST /api/admin/skills/{id}/versions`
- `POST /api/admin/skills/{id}/submit-review`
- `POST /api/admin/skills/{id}/publish`

Enterprise admin:

- `GET /api/admin/enterprise-skills`
- `POST /api/admin/enterprise-skills`
- `POST /api/admin/enterprise-skills/{id}/versions`
- `POST /api/admin/enterprise-skills/{id}/submit-review`
- `POST /api/admin/enterprise-skills/{id}/publish`

Availability and enablement operations:

- `GET /api/admin/skill-catalog`
- `POST /api/admin/skill-catalog/{skillId}/enable`
- `POST /api/admin/skill-catalog/{skillId}/disable`

Agent binding:

- `GET /api/admin/agents/{agentId}/skills`
- `POST /api/admin/agents/{agentId}/skills/install`
- `POST /api/admin/agents/{agentId}/skills/{skillId}/uninstall`

### 6.2 Why Not Reuse `/api/admin/skills`

`/api/admin/skills` is already overloaded.

Today it means:

- "enterprise-visible platform skills with enablement state"

Tomorrow product users will assume it means:

- "enterprise-owned skills"

That mismatch will confuse both UI and backend evolution.

Current repository direction:

- keep `/api/admin/skills` as the primary route family
- use `view=governance` and `view=catalog` to make list semantics explicit
- keep legacy `/api/admin/platform/skills` handlers only as compatibility aliases while remaining consumers are retired

## 7. Service Refactor Proposal

Current situation:

- one `Service` struct owns platform, enterprise, and agent methods

Recommended shape:

```go
type CatalogService struct { ... }      // create/list/get skill identity
type LifecycleService struct { ... }    // versions, review, publish, references
type EnterpriseService struct { ... }   // enterprise catalog + enablement
type BindingService struct { ... }      // agent install/uninstall/list
type HubService struct { ... }          // hub and import pipeline
```

This can still be wired together by a top-level facade if needed:

```go
type ServiceSet struct {
    Catalog    *CatalogService
    Lifecycle  *LifecycleService
    Enterprise *EnterpriseService
    Binding    *BindingService
    Hub        *HubService
}
```

This gives three benefits:

- permission intent becomes obvious in method signatures
- enterprise-owned skill support can be added without inflating one god-service
- tests become more focused and easier to maintain

## 8. Repository Refactor Proposal

Current repository interface mixes:

- skill master data
- version lifecycle
- availability state
- agent binding
- hub import

Recommended split:

```go
type CatalogRepository interface { ... }
type LifecycleRepository interface { ... }
type EnablementRepository interface { ... }
type BindingRepository interface { ... }
type HubRepository interface { ... }
```

`repository_gf.go` can still implement all of them in one file initially, but the interface boundary should be split first.

That lets service dependencies become explicit:

- platform catalog code should not depend on agent binding persistence
- availability code should not depend on hub import persistence

## 9. Data Model Adjustments Needed

The current schema is close, but not fully ready for enterprise-owned skills.

### 9.1 Keep `skills` Unified

Do not create two tables such as:

- `platform_skills`
- `enterprise_skills`

Keep a unified `skills` table with ownership fields:

- `owner_scope`
- `owner_enterprise_id`

This preserves:

- shared version graph
- shared dependency validation
- shared policy model

### 9.2 Tighten Queries

Once enterprise-owned skills are real, queries must always be explicit about scope:

- platform lists: `owner_scope = platform`
- enterprise-owned lists: `owner_scope = enterprise and owner_enterprise_id = ?`
- enterprise catalog:
  - platform-published-to-enterprise
  - plus enterprise-owned skills visible to that enterprise

Current `ListPublishedSkillsForEnterprise()` will need to evolve into clearer queries, for example:

- `ListPlatformSkillsForEnterpriseCatalog(enterpriseId)`
- `ListEnterpriseOwnedSkills(enterpriseId)`
- `ListAvailableSkillsForEnterprise(enterpriseId)`

### 9.3 Add Enterprise Publishing Visibility

Enterprise-owned skills need a visibility model.

Suggested phase-2 additions:

- `visibility_scope`:
  - `enterprise_private`
  - `enterprise_shared`
  - future reserved values if needed
- optional `published_to_catalog_at`

This keeps enterprise-owned drafts separate from enterprise-available installable skills.

## 10. Permissions and Policy

Current service methods rely heavily on route-layer privilege separation.

That is not enough once the domain becomes more complex.

Recommended rule:

- platform lifecycle methods require `actor.IsPlatformAdmin`
- enterprise-owned skill lifecycle methods require:
  - `actor.EnterpriseId != ""`
  - `actor.EnterpriseRole in {owner, admin}`
- enablement methods require enterprise admin scope
- binding methods require enterprise admin scope and agent-enterprise ownership validation

This should be expressed in service-level policy helpers, not only in handler middleware.

## 11. Migration Strategy

### Phase A: Safe Refactor Without Product Changes

- split service files by responsibility
- split repository interfaces by responsibility
- keep existing routes unchanged
- keep current behavior unchanged

### Phase B: Clarify API Semantics

- add explicit enterprise-owned skill routes
- keep old `/api/admin/skills` as compatibility alias
- adjust frontend labels to:
  - platform skill
  - enterprise skill
  - enterprise catalog / enablement

### Phase C: Enable Enterprise-Owned Skills

- support `ownerScope = enterprise`
- allow enterprise skill creation and version lifecycle
- add visibility rules
- unify enterprise install path across platform-owned and enterprise-owned skills

## 12. Recommendation Summary

Recommended decision for this project:

- **yes**, backend module design should be adjusted
- **no**, it should not be split into two top-level business domains yet
- **yes**, the `skill` domain should be internally decomposed now
- **yes**, API semantics for enterprise-owned skills should be clarified before that feature is implemented

The best near-term move is:

1. keep one `skill` bounded context
2. split services and repositories by capability
3. rename enterprise-facing APIs before enterprise-owned skill creation lands
4. then implement enterprise-owned skill lifecycle on top of the same shared model

## 13. Discussion Questions

Before implementation, these decisions should be confirmed:

1. Should enterprise-owned skills be fully independent, or must they always derive from a platform skill?
2. Should enterprise-owned skills support their own private hubs in phase 1 or only in phase 2?
3. Should `/api/admin/skills` remain a compatibility alias indefinitely, or be formally deprecated?
4. Does enterprise publish mean "visible inside this enterprise only", or do we need cross-enterprise sharing later?

## 14. Confirmed Direction After Discussion

The current preferred direction is now clarified:

- keep **one business domain**
- keep **one business model and one business logic system**
- separate platform and enterprise behavior mainly through **permissions and actor context**
- do **not** introduce enterprise-private hubs in phase 1
- prefer **unified API paths**, rather than making platform skill management look like a separate specialized API family

This changes the recommendation in one important way:

- we still keep one `skill` domain
- we still should split internal responsibilities for maintainability
- but we should **not** evolve toward "platform API family vs enterprise API family" as the primary product-facing concept

## 15. Revised API Direction

The preferred direction is:

- unified admin skill routes
- actor context decides what is visible and what actions are allowed
- query parameters or operation semantics distinguish "platform governance" from "availability control"

Suggested route shape:

- `GET /api/admin/skills`
  - platform admin: may see full governance view
  - enterprise admin: sees enterprise-visible catalog view
- `POST /api/admin/skills`
  - platform admin: create platform-owned skill
  - enterprise admin: create enterprise-owned skill when enabled in product phase
- `GET /api/admin/skills/{id}`
- `POST /api/admin/skills/{id}/versions`
- `GET /api/admin/skills/{id}/versions/{versionId}/references`
- `POST /api/admin/skills/{id}/references`
- `POST /api/admin/skills/{id}/submit-review`
- `POST /api/admin/skills/{id}/publish`
- `POST /api/admin/skills/{id}/enable`
- `POST /api/admin/skills/{id}/disable`

Optional query hints may still be added for UI use, for example:

- `GET /api/admin/skills?view=governance`
- `GET /api/admin/skills?view=catalog`
- `GET /api/admin/skills?ownerScope=platform`
- `GET /api/admin/skills?ownerScope=enterprise`

Current implementation direction in this repository:

- `view=governance` is the explicit platform-admin governance list view
- `view=catalog` is the explicit enterprise/admin consumption catalog view
- no `view` parameter currently falls back to `catalog` behavior for compatibility

The key point is:

- same resource family
- same core model
- different visibility and actions based on actor permissions

## 16. Revised Backend Refactor Priority

Based on this confirmed direction, the best next-step sequence is:

### Phase A: Normalize Semantics Without Breaking Product Shape

- keep `/api/admin/skills` as the main route family
- gradually retire legacy `/api/admin/platform/skills` aliases after all callers move to `/api/admin/skills`
- introduce actor-aware list/detail behavior

### Phase B: Internal Service Refactor

- keep one top-level `skill.Service` facade if convenient
- internally split capability handlers:
  - catalog
  - lifecycle
  - availability control
  - agent binding
  - hub/import

This preserves a unified external mental model while avoiding a god-service internally.

### Phase C: Enterprise-Owned Skill Support

- enable `ownerScope = enterprise`
- reuse the same `skills` and `skill_versions` model
- keep platform hubs as the only hub source in phase 1
- use permission checks to decide create/publish/enable/install rights

## 17. Revised Practical Recommendation

So the concrete recommendation is now:

1. **do not** create a second top-level backend domain
2. **do not** create platform-specific and enterprise-specific API families as the long-term main shape
3. **do** keep one `skill` resource family
4. **do** express platform vs enterprise through actor context, owner scope, and permission checks
5. **do** refactor internal services and repositories now, before enterprise-owned skill lifecycle becomes real
