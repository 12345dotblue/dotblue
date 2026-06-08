# Skill Enterprise External Import Design

## 1. Goal

This document defines the next-step design for skill management after confirming:

- enterprises are allowed to import external skills
- enterprises do **not** manage hub sources directly
- the backend must keep **one domain**, **one model family**, and **one logic system**
- platform can open skills and hubs either globally or to specific enterprises
- `owner_scope` must remain extensible for future scope types

The design intentionally keeps the existing `backend/internal/domains/skill` boundary.
It does **not** introduce a separate enterprise import domain.

## 2. Confirmed Constraints

### 2.1 Product Constraints

- platform provides hub sources
- enterprise admins can import from platform-provided hubs
- enterprise admins cannot create, edit, or own hub definitions in phase 1
- imported skills must stay visible in the main management path, not only in import jobs

### 2.2 Backend Constraints

- keep one `skill` business domain
- keep one shared import pipeline and one shared lifecycle model
- separate platform and enterprise behavior through actor context and permissions
- avoid duplicating service logic for platform import vs enterprise import

### 2.3 Visibility Constraints

- platform can open a skill to:
  - all enterprises
  - selected enterprises only
- platform can open a hub to:
  - all enterprises
  - selected enterprises only
- enterprise-owned skills are visible only inside the owning enterprise unless future sharing is introduced

## 3. Core Design Decision

The system should distinguish three concerns explicitly:

- ownership: who owns the skill or hub definition
- visibility: who can see and import or consume it
- enablement/binding: who has opened it for use and which agent has installed it

Current problems come from mixing these concerns together.

The revised model is:

- `owner_scope` answers "who owns this resource"
- release/open records answer "who can see or consume this resource"
- enterprise enablement answers "is this published skill usable inside the enterprise"
- agent binding answers "is this skill installed on the agent"

## 4. Target Business Model

### 4.1 Platform Hub Catalog

Platform owns all hub definitions in phase 1.

Responsibilities:

- create and update hub definitions
- configure hub trust and import policy
- decide whether a hub is:
  - globally visible to all enterprises
  - visible only to selected enterprises

Enterprises can:

- browse hubs opened to them
- submit import jobs against those hubs

Enterprises cannot:

- create hubs
- edit hub credentials
- edit hub policy

### 4.2 Skill Ownership

Imported skills use the same ownership rule as manually created skills:

- platform admin import -> `owner_scope = platform`
- enterprise admin import -> `owner_scope = enterprise`

This means enterprise external import produces **enterprise-owned skills**, not platform-owned skills with enterprise labels.

That is the simplest way to satisfy:

- enterprise internal governance autonomy
- one shared skill model
- one shared publish/install flow

### 4.3 Skill Availability

Skill availability keeps the current consumption model:

- platform-owned published skills require enterprise open/enablement before agent installation
- enterprise-owned published skills in the same enterprise are considered enterprise-local supply and do not require an extra cross-owner release step

In other words:

- platform-owned skill: `publish -> open to enterprise -> install to agent`
- enterprise-owned skill: `import/create -> publish -> install to agent`

## 5. Scope Model

## 5.1 Why Current Scope Fields Are Not Enough

Current skill data uses:

- `owner_scope`
- `owner_enterprise_id`

That is enough for `platform` vs `enterprise`, but it is not ideal for future scopes such as:

- `partner`
- `workspace`
- `project`
- `organization_unit`

The next design should standardize scope storage and scope checks.

## 5.2 Recommended Scope Tuple

Introduce a shared scope tuple shape:

- `owner_scope`
- `owner_scope_ref_id`

Recommended semantics:

- `owner_scope = platform` -> `owner_scope_ref_id = ""`
- `owner_scope = enterprise` -> `owner_scope_ref_id = <enterprise_id>`
- `owner_scope = partner` -> `owner_scope_ref_id = <partner_id>`

For compatibility with current code and data:

- keep `owner_enterprise_id` as a compatibility field in phase 1 if needed
- service and repository logic should gradually normalize around one helper:
  - `ScopeRef { Scope string; RefId string }`

This keeps `owner_scope` extensible without splitting the domain.

## 5.3 Scope Rules

Recommended scope values:

- `platform`
- `enterprise`
- `partner`

Reserved for future:

- `workspace`
- `project`

The design should avoid hard-coding enterprise-only logic in new hub/import features.

## 6. New Visibility Model

Ownership alone is not enough for hubs or platform-owned skills.

Platform needs to decide:

- visible to all enterprises
- visible to selected enterprises

So a release/open layer is required.

### 6.1 Hub Release

Add a hub visibility table, for example:

- `skill_hub_releases`

Fields:

- `id`
- `hub_id`
- `release_scope`
- `target_enterprise_id`
- `release_status`
- `created_by`
- `created_at`
- `updated_at`

Release semantics:

- `release_scope = global` means all enterprises can browse and import from this hub
- `release_scope = enterprise` means only the specified enterprise can browse and import from this hub

This follows the same product rule as skill opening, but for import sources.

### 6.2 Platform Skill Release

Current enterprise enablement already models platform skill opening at enterprise level.
That model can remain as the consumption gate.

But the UI and service semantics should present it as:

- platform publishes the skill
- platform may recommend or default-open it
- enterprise may see or activate it depending on release policy

Phase 1 recommendation:

- reuse `enterprise_skill_enablements` for platform skill opening
- do **not** introduce a second separate platform skill release table unless policy semantics become more complex

### 6.3 Enterprise-Owned Skill Visibility

Enterprise-owned skills do not need cross-enterprise release in phase 1.

Rules:

- owner enterprise can govern and publish them
- same enterprise can install them to agents
- other enterprises cannot see them

## 7. Backend Design

## 7.1 Keep One Domain

Keep all logic under:

- `backend/internal/domains/skill`

Keep existing internal capability split direction:

- catalog
- lifecycle
- availability
- binding
- hub/import

Do **not** create:

- `enterprise_skill_import`
- `platform_skill_import`

as separate top-level domains.

## 7.2 Actor-Aware Hub and Import Logic

Current `ImportSkill()` is platform-admin only.

Revised rule:

- if actor is platform admin and no enterprise context is active:
  - import target owner is `platform`
- if actor has enterprise admin context:
  - import target owner is `enterprise`
  - actor may only import from hubs released to that enterprise

The pipeline itself stays the same:

1. validate hub access
2. create import job
3. resolve manifest
4. create draft skill
5. create draft version
6. finish job

Only the target owner and access checks differ.

## 7.3 Repository Direction

Current repository interfaces are global:

- `ListSkillHubs()`
- `ListSkillImportJobs()`

They should become actor-aware or scope-aware.

Recommended direction:

- `ListSkillHubsForActor(actor ActorContext)`
- `ListSkillImportJobsForActor(actor ActorContext)`
- `GetSkillHubForImport(actor ActorContext, hubId string)`

Important rule:

- repositories still only fetch data
- access decisions stay in service helpers

## 7.4 Access Checks

Introduce clear service helpers:

- `canManageHub(actor, hub)`
- `canImportFromHub(actor, hub)`
- `canViewImportJob(actor, job)`
- `resolveImportTargetOwner(actor)`

This keeps permission logic reusable and visible.

## 8. API Design

The design should keep one resource family and actor-aware semantics.

## 8.1 Hub Routes

Platform management routes:

- `GET /api/admin/platform/skill-hubs`
- `POST /api/admin/platform/skill-hubs`
- `PUT /api/admin/platform/skill-hubs/{id}`
- `POST /api/admin/platform/skill-hubs/{id}/release`

Enterprise consumption route:

- `GET /api/admin/skill-hubs`

Semantics:

- platform admin sees all hubs and may manage them
- enterprise admin sees only hubs released to the current enterprise

## 8.2 Import Routes

Keep one import resource family, actor-aware:

- `GET /api/admin/skill-import-jobs`
- `POST /api/admin/skill-import-jobs`
- `GET /api/admin/skill-import-jobs/{id}`

Semantics:

- platform admin sees platform-owned import jobs
- enterprise admin sees import jobs created inside the current enterprise scope

Compatibility option:

- keep `/api/admin/platform/skill-import-jobs` as alias during migration

## 8.3 Skill Routes

Keep the current unified family:

- `GET /api/admin/skills?view=governance`
- `GET /api/admin/skills?view=catalog`
- `POST /api/admin/skills`
- `POST /api/admin/skills/{id}/versions`
- `POST /api/admin/skills/{id}/submit-review`
- `POST /api/admin/skills/{id}/publish`
- `POST /api/admin/skills/{id}/enable`
- `POST /api/admin/skills/{id}/disable`

Semantics:

- platform admin governance view -> platform-owned skills
- enterprise admin governance view -> enterprise-owned skills
- enterprise admin catalog view -> platform-published skills plus enterprise-local published skills usable by the current enterprise

## 9. Frontend Information Architecture

## 9.1 Platform Layer

Platform skill management should focus on platform supply:

- platform skills
- platform hubs
- platform import jobs
- platform release/open actions

Recommended views:

- `Platform Skills`
- `Platform Hubs`
- `Platform Import Jobs`

Platform must also expose actions:

- open hub globally
- open hub to selected enterprises
- open skill globally
- open skill to selected enterprises

## 9.2 Enterprise Layer

Enterprise skill management should become the real enterprise workspace.

Recommended enterprise views:

- `治理视图`
  - enterprise-owned skills
  - includes enterprise-created and enterprise-imported skills
- `目录视图`
  - platform-published skills visible to this enterprise
  - supports open/use actions
- `导入任务`
  - enterprise-owned import jobs
- `可用 Hub`
  - platform-provided hubs released to this enterprise

Important:

- enterprise cannot edit hub source config
- enterprise can only choose a released hub and submit import

## 9.3 Agent Layer

Agent page remains the final installation center.

It should only explain:

- installed
- open and installable
- pending enterprise open
- unavailable

It should not become a hub/import management page.

## 9.4 Summary Cards

All summary cards should become clickable filters.

This applies to:

- platform skills
- platform hubs
- platform import jobs
- enterprise governance list
- enterprise import jobs
- agent catalog states

Current summary cards are display-only and should be converted into interactive filters.

## 10. Main User Flows

## 10.1 Platform Imports a Public Skill

1. platform admin chooses a platform hub
2. import job creates a platform-owned draft skill
3. platform publishes the skill
4. platform opens it globally or to selected enterprises
5. enterprise sees it in catalog
6. enterprise opens it for use if required by policy
7. agent installs it

## 10.2 Enterprise Imports an Internal Skill

1. enterprise admin opens enterprise skill workspace
2. chooses a released hub from `可用 Hub`
3. submits import job
4. import creates an enterprise-owned draft skill
5. enterprise reviews and publishes it
6. enterprise installs it on agents

No separate platform opening step is needed because the skill is already enterprise-owned.

## 11. Why This Satisfies the Constraints

### 11.1 Enterprise Cannot Manage Hubs

Satisfied because:

- hub definitions remain platform-owned
- enterprise only sees released hubs
- enterprise only consumes hubs, never edits them

### 11.2 One Model and One Logic System

Satisfied because:

- import pipeline remains one service flow
- skill lifecycle remains one service flow
- actor context decides owner and permission
- same `skill` domain handles platform and enterprise

### 11.3 Platform Can Open Skill and Hub Globally or Per Enterprise

Satisfied because:

- skill opening uses enterprise-facing availability/open semantics
- hub opening uses `skill_hub_releases`
- both support:
  - global
  - enterprise-specific

### 11.4 `owner_scope` Is Extensible

Satisfied because:

- the design standardizes a generic scope tuple
- new scope types do not require creating a new domain

## 12. Recommended Delivery Order

### Phase 1

- add scope-aware ownership for `skill_hubs` and `skill_import_jobs`
- add hub release records
- allow enterprise import from released hubs
- keep platform as the only hub manager

### Phase 2

- restructure enterprise skill UI into governance / catalog / import jobs / available hubs
- make summary cards clickable filters
- make imported target skills always visible from import jobs into governance details

### Phase 3

- standardize generic scope tuple in code and schema
- reduce duplicate front-end modules and shared logic

## 13. Practical Recommendation

The best next implementation direction is:

1. keep one `skill` domain
2. keep one import pipeline
3. let enterprise import from platform-released hubs
4. create enterprise-owned skills from enterprise import jobs
5. add explicit hub release visibility
6. keep agent page focused only on installation

This is the smallest design that solves the current product confusion without exploding the backend model.
