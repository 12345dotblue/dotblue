# DotBlue Skill Frontend Design

## 1. Scope

This document defines the frontend UX, information architecture, and interaction flows for the DotBlue skill system.

It is aligned with:

- [skill-system-design.md](file:///c:/Users/kongz/work/dotblue/docs/skill-system-design.md)
- [skill-database-design.md](file:///c:/Users/kongz/work/dotblue/docs/skill-database-design.md)
- [skill-backend-design.md](file:///c:/Users/kongz/work/dotblue/docs/skill-backend-design.md)

The design targets the existing frontend stack:

- React + TypeScript
- Ant Design
- axios
- react-router
- i18next

## 2. Current Frontend Structure to Integrate With

### 2.1 Existing Main Routes

Current main authenticated pages:

- `/dashboard` -> agent list and editing
- `/chat` -> chat page (special layout)
- `/admin/enterprise` -> enterprise admin console
- `/admin/platform` -> platform settings (platform admin only)

See [App.tsx](file:///c:/Users/kongz/work/dotblue/web/src/App.tsx#L178-L235).

### 2.2 Existing Layout Behavior

- `AppLayout` provides a left navigation menu and a header toolbar.
- The side menu currently contains only a small number of fixed routes.
- `AdminSettings` is a tabbed enterprise admin page using `?tab=...`.
- `PlatformSettingsPage` is currently a single page with cards (not tabbed).

References:

- [AppLayout.tsx](file:///c:/Users/kongz/work/dotblue/web/src/components/Layouts/AppLayout.tsx)
- [AdminSettings.tsx](file:///c:/Users/kongz/work/dotblue/web/src/domains/admin/AdminSettings.tsx)
- [PlatformSettingsPage.tsx](file:///c:/Users/kongz/work/dotblue/web/src/domains/admin/PlatformSettingsPage.tsx)

## 3. UX Principles

- Use one product term: `Skill`
- Keep management flows explicit: import != publish != enable != install
- Provide strong governance cues:
  - trust level
  - risk level
  - approval requirement
  - channel scope
  - enablement scope
  - installed agents
- Provide a “trace-first” experience for troubleshooting:
  - execution timeline
  - trace tree
  - audit detail
- Minimize cross-page cognitive load:
  - stable filters
  - consistent status colors
  - predictable primary actions

## 4. Frontend Information Architecture

The skill system introduces new platform and enterprise management surfaces. The recommended IA is:

### 4.1 Platform Admin (Platform-wide governance)

New pages:

- `Platform Skills`
- `Skill Hubs`
- `Skill Import Jobs`
- `Skill Executions`
- `Skill Audits`

Existing page:

- `Platform Settings`

### 4.2 Enterprise Admin (Enterprise enablement + agent installation)

New surfaces:

- `Enterprise Skills` (enablement and policy override)
- `Agent Skill Installation` (install/uninstall, pin version)

Existing surfaces:

- enterprise org structure
- enterprise members
- invitations
- enterprise LLM
- usage
- IM integrations

### 4.3 Members (Read-only capability visibility)

Optional read-only surface:

- view installed skills for a specific agent

This supports future “what can this agent do” UX without granting management controls.

## 5. Navigation and Routing Plan

### 5.1 Recommended New Routes

Platform admin:

- `/admin/platform/settings` (optional split, can reuse existing `/admin/platform`)
- `/admin/platform/skills`
- `/admin/platform/skills/:id`
- `/admin/platform/skill-hubs`
- `/admin/platform/skill-import-jobs`
- `/admin/platform/skill-import-jobs/:id`
- `/admin/platform/skill-executions`
- `/admin/platform/skill-audits`

Enterprise admin:

- keep `/admin/enterprise` and add a `skills` tab
- optionally add `/admin/enterprise/agents/:agentId/skills` for deeper agent install experience

Member:

- `/agents/:agentId/skills` (read-only)

### 5.2 Recommended Side Menu Changes

The current menu keys are exact route paths. To integrate new pages cleanly:

- keep `/admin/enterprise` as-is
- keep `/admin/platform` as-is (or redirect to `/admin/platform/settings`)
- add:
  - `/admin/platform/skills`
  - `/admin/platform/skill-hubs`
  - `/admin/platform/skill-executions`

If the team wants to keep the menu compact, the last two can be grouped into a single entry:

- `/admin/platform/skills` as the primary entry
- hubs and executions accessible from sub-tabs inside the skills page

This is a product choice:

- “more routes” -> clearer deep linking, better for ops
- “fewer routes” -> less navigation complexity

The recommended default is “more routes” because skills are a governance system, not a one-off form.

## 6. Page Designs

## 6.1 Platform Skills Page

Route:

- `/admin/platform/skills`

Goal:

- manage the global skill catalog (platform view)

Primary actions:

- `Create Skill`
- `Import Skill`

Secondary actions:

- `Review Queue`
- `Publish`
- `Disable`
- `Rollback`

Main content:

- `Table` of skills with filters and status tags

Suggested table columns:

- `code`
- `name`
- `sourceType`
- `ownerScope`
- `trustLevel` (tag)
- `riskLevel` (tag from latest published version policy)
- `status` (tag)
- `latestStableVersion` (if exists)
- `updatedAt`
- row actions: `View`, `Publish`, `Disable`

Suggested filters:

- `status`
- `sourceType`
- `ownerScope`
- `trustLevel`
- `keyword` (code/name)

Suggested row click:

- navigate to detail page `/admin/platform/skills/:id`

Recommended UI components:

- `Table` + `Form` for filters
- `Drawer` or `Modal` for quick actions (publish/disable)
- `Tag` for risk and trust
- `Segmented` or `Tabs` for views:
  - `All`
  - `Reviewing`
  - `Published`
  - `Disabled`

## 6.2 Skill Detail Page

Route:

- `/admin/platform/skills/:id`

Goal:

- full governance detail for one skill

Suggested layout:

- top header: name + code + key status tags
- `Tabs` for:
  - `Overview`
  - `Versions`
  - `References`
  - `Policy`
  - `Release History`

### 6.2.1 Overview Tab

Shows:

- identity fields
- source and provenance
- trust level and verification report summary
- last import job link (if imported)
- “where used” summary:
  - number of enabled enterprises
  - number of installed agents

### 6.2.2 Versions Tab

`Table` of versions:

- version
- release channel
- release status
- createdAt
- publishedAt
- actions: `View Manifest`, `Submit Review`, `Publish`, `Disable`

Version detail action:

- open `Drawer`:
  - manifest JSON preview (collapsed sections)
  - input/output schema preview
  - default policy preview
  - verification report preview
  - risk report preview

### 6.2.3 References Tab

Shows dependency graph:

- outgoing references (this version -> others)
- incoming references (others -> this version)

MVP visualization:

- two `Table`s with links to other skills

Enhanced visualization (later):

- graph view (node-edge)

### 6.2.4 Policy Tab

Shows:

- default policy for latest stable/published version
- policy differences between versions (optional)

MVP:

- render key policy fields
- “show raw JSON” collapsible panel

### 6.2.5 Release History Tab

Shows `skill_release_records`:

- action
- from/to status
- channel
- note
- operatedBy
- createdAt

## 6.3 Skill Hubs Page

Route:

- `/admin/platform/skill-hubs`

Goal:

- register and manage skill hubs (platform boundary)

Primary actions:

- `Add Hub`

Secondary actions:

- `Edit`
- `Enable/Disable`
- `Test Connection`
- `Sync Now`

Table columns:

- `hubCode`
- `name`
- `hubType`
- `status`
- `trustLevel`
- `syncMode`
- `lastSyncedAt`
- `lastError`
- row actions

Hub edit form sections:

- basic info: code/name/type/baseUrl
- auth scheme: none/apiKey/oauth2/oidc (only UI representation initially)
- namespace allow list
- import policy
- signature policy
- network policy

Note:

The form should avoid showing secrets in plaintext. For `secret_json` fields:

- show “configured” placeholder
- allow “update secret” action via separate input

## 6.4 Skill Import Jobs Page

Routes:

- `/admin/platform/skill-import-jobs`
- `/admin/platform/skill-import-jobs/:id`

Goal:

- visibility and troubleshooting for imports

List table columns:

- job id
- hub
- requestedBy
- source locator
- status
- startedAt
- finishedAt
- error summary

Detail page sections:

- job timeline (pending -> parsing -> normalizing -> verifying -> sandboxing -> completed)
- parsed descriptor preview (collapsed)
- normalized manifest preview (collapsed)
- verification report
- risk report
- target skill/version links

Primary actions (job detail):

- `Retry` (if failed)
- `Create Draft Skill` (if imported but not persisted)

## 6.5 Enterprise Skills Tab (Enablement)

Location:

- inside `/admin/enterprise` as a new `tab=skills`

Goal:

- enable and govern skills at enterprise scope

Primary actions:

- `Enable Skill`

Secondary actions:

- `Disable`
- `Edit Policy Override`

Main content:

- `Table` of enabled skills (or “All available published skills” with enable state)

Recommended UI approach:

- top filter: show `Enabled only` toggle
- `Enable Skill` opens a `Modal` with:
  - choose skill (search by code/name)
  - choose default version behavior:
    - follow stable
    - pin a version (optional for phase 1)
  - org scope selection (tree selector, reuse org data already in AdminSettings)
  - channel scope selection (web/im/api)
  - policy override section (risk/approval/quota/timeouts) with safe defaults

Enterprise enablement row columns:

- code/name
- enablement status
- org scope summary
- channel scope summary
- approval summary
- updatedAt
- actions: edit/disable

## 6.6 Agent Skill Installation

Where to place this UX:

- recommended: add to AgentList edit modal as a new section or tab
- alternative: enterprise admin skills tab shows per-agent install actions

Recommended default:

- add an “Installed Skills” tab inside the agent edit modal

Reason:

- installation is agent-specific
- it matches user mental model: “configure an agent”
- it avoids adding yet another admin page for phase 1

### 6.6.1 Installed Skills Tab (Agent Modal)

Sections:

- installed list table
- install action button

Installed list columns:

- skill code/name
- pinned version
- invoke visibility (auto/suggested/manual)
- channel scope summary
- status (installed/suspended)
- actions: uninstall, suspend/resume, edit override

Install flow (modal within modal or drawer):

- select enabled skill (enterprise enablement required)
- select version (pin) or follow stable (phase 2)
- set entry alias (optional)
- set invoke visibility
- optional policy override
- install

Guardrails:

- if enterprise has not enabled the skill, the picker should show it as unavailable
- if skill is unverified/blocked, installation should be denied and UI should show reason

## 6.7 Skill Executions Page

Route:

- `/admin/platform/skill-executions`

Goal:

- operational visibility and debugging

Filters:

- enterpriseId (optional, platform admin only)
- agentId
- skillId
- conversationId
- traceId
- status
- time range

Table columns:

- startedAt
- status
- skill code/name + version
- agent
- channel
- actor
- duration
- error code
- trace link

Row click opens:

- `Drawer` for trace detail:
  - execution tree (parent/child)
  - input digest and output digest
  - error
  - metrics
  - related audit entries

## 6.8 Skill Audits Page

Route:

- `/admin/platform/skill-audits`

Goal:

- governance and compliance investigation

Filters:

- enterpriseId
- actorId
- skillId
- decision
- time range

Table columns:

- createdAt
- decision
- action
- actor
- skill
- execution link

Audit detail drawer:

- masked payload JSON
- resource scope JSON
- approval reference
- linkage to execution

## 7. Cross-Page Interaction Flows

## 7.1 Create Skill (Platform Admin)

Flow:

1. open `Create Skill` modal
2. fill identity fields (code/name/source/owner)
3. optionally create initial version with manifest and policy
4. save as draft
5. navigate to detail page

UX requirements:

- validate code format early (namespace-like)
- show uniqueness error inline if backend rejects

## 7.2 Import Skill (Platform Admin)

Flow:

1. open `Import Skill` modal
2. choose hub
3. enter source locator
4. start import job
5. navigate to job detail
6. review normalized manifest and verification
7. create draft skill version
8. submit review

UX requirements:

- show job status progression live (polling)
- keep a stable “job id” link for shareable troubleshooting

## 7.3 Review and Publish (Platform Admin)

Flow:

1. open skill detail
2. choose a version in `reviewing` state
3. publish to candidate or stable
4. observe release record and updated pointers

UX requirements:

- present dependency and cycle validation errors clearly
- show trust and verification requirements if publish is denied

## 7.4 Enable Skill (Enterprise Admin)

Flow:

1. open enterprise `Skills` tab
2. select a skill (only published skills should appear)
3. configure org/channel scope
4. configure policy override (optional)
5. enable

UX requirements:

- show “platform disabled” or “blocked” status if it exists
- allow enterprise-level disable without deleting records

## 7.5 Install Skill on Agent (Enterprise Admin)

Flow:

1. open agent edit modal
2. go to `Installed Skills` tab
3. click `Install Skill`
4. select enabled skill + version
5. configure binding override
6. install

UX requirements:

- show why install is denied (not enabled, not published, unverified, policy denied)
- show effective policy preview for confidence

## 7.6 Debug and Audit

Flow:

1. open executions page
2. filter by agent or conversation
3. open trace detail
4. jump to audit entries
5. export or copy trace id

UX requirements:

- keep trace id visible and copyable
- show parent-child call relationships

## 8. Permissions and Feature Gating

### 8.1 Roles

- platform admin: can access platform pages
- enterprise admin: can enable skills, install skills on agents
- enterprise member: read-only view if enabled

### 8.2 UI Gating Rules

- hide platform skill menu entries if user is not platform admin
- hide enterprise skill management actions if user is not enterprise admin
- show read-only “Installed Skills” section to members only if product wants transparency

The current frontend already derives:

- `isAdmin` from token
- enterprise role from `/api/enterprises/current`

See [AppLayout.tsx](file:///c:/Users/kongz/work/dotblue/web/src/components/Layouts/AppLayout.tsx#L52-L109).

## 9. Data Fetching and State Management

Current code uses:

- `useState` + `useEffect`
- direct `axios` calls per page

For consistency, phase 1 should keep the same style.

Recommended improvements (optional later):

- introduce a small set of `useSkill*` hooks
- centralize auth header builder
- standardize error mapping and toast behaviors

## 10. i18n Plan

The skill UI introduces many new strings. For phase 1, the fastest approach is:

- add English and Chinese keys into `web/src/i18n/config.ts`
- keep keys stable and descriptive

Key groups:

- `skill_*` for generic skill labels and statuses
- `skill_hub_*` for hub forms
- `skill_import_*` for import jobs
- `skill_enablement_*` for enterprise enablement UI
- `skill_binding_*` for agent installation UI
- `skill_execution_*` for execution and audit UI

## 11. Minimal Phase 1 UI Delivery

Phase 1 should prioritize UX surfaces that unblock adoption:

- platform skill catalog list + detail (draft/publish/disable)
- enterprise enablement tab (enable/disable)
- agent installed skills tab (install/uninstall)

Executions and audits can be basic list pages initially and refined later.

## 12. Summary

The skill frontend should extend DotBlue with enterprise-grade governance UX:

- platform-wide catalog and release control
- hub and import visibility
- enterprise enablement and policy override
- explicit agent installation flows
- trace-first execution and audit visibility

The recommended approach is to add dedicated platform routes for skill governance and to integrate enterprise enablement and agent installation into existing enterprise and agent management surfaces.

