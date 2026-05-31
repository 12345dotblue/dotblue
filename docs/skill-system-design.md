# DotBlue Skill System Design

## 1. Purpose

This document defines the target design for a unified `Skill` system in DotBlue.

The design aligns with the current product positioning:

- enterprise-grade AI assistant governance platform
- self-hosted and cloud-hosted deployment patterns
- multi-agent management with isolated runtimes
- enterprise administration, IM integration, and platform-wide model governance

The design also follows local engineering constraints:

- organize by business domain instead of technical layers across the whole repository
- keep `handler`, service logic, and storage separated inside each domain
- depend on abstractions instead of concrete implementations
- keep critical design intent documented with useful English comments in code

## 2. Problem Statement

DotBlue already has strong foundations for `Agent`, chat, enterprise management, IM integration, and isolated runtimes. However, capability management is still implicit:

- capabilities are mainly carried by agent prompt and runtime behavior
- runtime tool events are visible, but they are not yet platform-governed capability assets
- external capabilities do not enter a unified control plane
- there is no full lifecycle for creation, import, review, publish, enablement, installation, execution, and audit

The platform now needs a first-class `Skill` system that can govern both built-in and external capabilities.

## 3. Design Goals

- Use a single product concept: `Skill`
- Treat every skill as an equal governance object
- Support built-in, partner, enterprise-private, and external imported skills
- Allow skill-to-skill references under the same governance and audit model
- Provide a full lifecycle from creation or import to installation on agents
- Keep policy, approval, audit, and observability consistent across all skill sources
- Integrate with existing agent, chat, IM, settings, and runtime architecture

## 4. Non-Goals

- No public marketplace-first strategy in phase 1
- No user-facing separation into tool, action, workflow, or primitive/composite concepts
- No direct agent-level connection to external hubs
- No source-specific bypass for policy, audit, or approval

## 5. Core Principles

### 5.1 One Product Concept

DotBlue exposes only one capability concept: `Skill`.

This keeps the product language simple for platform operators, enterprise admins, and business owners:

- which skills exist
- which enterprises can enable them
- which agents can install them
- which channels can invoke them
- what approvals and audits apply

### 5.2 Equal Governance

All skills are equal governance objects regardless of source or internal complexity.

Every skill must have:

- identity
- version
- source
- trust level
- policy
- enablement state
- installation state
- execution trace
- audit record

### 5.3 External Diversity, Internal Uniformity

External ecosystems may use MCP, OpenAPI, package manifests, or partner-specific protocols. DotBlue should not let those protocols define its internal platform model.

All external capabilities must be normalized into a DotBlue canonical skill model before they can be governed or installed.

### 5.4 Centralized Platform Control

External hubs are platform-managed assets. Agents do not directly connect to hubs, and enterprises do not bypass platform review.

Platform administrators control:

- which hubs are registered
- how hubs are authenticated
- which namespaces are allowed
- how imports are reviewed
- what trust and risk defaults apply

### 5.5 Explicit Policy Resolution

A skill is never executed based on raw definition alone. Every invocation must resolve effective policy from multiple scopes:

- platform default policy
- skill version default policy
- enterprise enablement policy override
- agent binding override
- runtime context such as channel, actor, organization, and approval status

### 5.6 Auditable by Default

Every skill invocation must produce:

- an execution record for runtime and troubleshooting
- an audit record for governance and compliance

The system must preserve both direct and transitive skill invocations inside one traceable call graph.

## 6. Terminology

- `Skill`: the only capability object visible to the product
- `Skill Version`: a concrete releasable and installable version of a skill
- `Skill Manifest`: canonical normalized descriptor for a skill version
- `Skill Hub`: a source registry from which skills can be created or imported
- `Skill Import Job`: an import pipeline execution from a hub into DotBlue
- `Skill Enablement`: enterprise-level activation of a skill
- `Skill Binding`: installation of a skill onto an agent
- `Skill Execution`: runtime execution record of one skill invocation
- `Skill Audit`: governance and compliance record tied to a skill execution
- `Skill Reference`: a directed dependency from one skill version to another

## 7. High-Level Architecture

The target system is split into four planes.

### 7.1 Control Plane

Responsible for:

- skill definition
- versioning
- import and normalization
- validation
- review
- release
- enterprise enablement
- agent installation
- lifecycle state management

### 7.2 Policy Plane

Responsible for:

- trust evaluation
- risk classification
- permission checks
- approval requirements
- data scope
- channel scope
- quota and timeout policy
- deployment and release constraints

### 7.3 Runtime Plane

Responsible for:

- effective skill resolution
- invocation dispatch
- skill-to-skill reference handling
- cycle prevention
- timeout, retry, and circuit breaking
- execution trace propagation

### 7.4 Observability Plane

Responsible for:

- execution timeline
- audit trail
- trace graph
- latency and failure metrics
- security events
- operational dashboards

## 8. Domain Placement in the Repository

The design should be implemented as a dedicated business domain:

`backend/internal/domains/skill`

Suggested internal layout:

```text
backend/internal/domains/skill/
  skill.go
  handler.go
  service.go
  repository.go
  repository_gf.go
  manifest.go
  policy.go
  reference.go
  hub.go
  import_job.go
  release.go
  execution.go
  audit.go
  runtime.go
```

If the domain grows, subpackages may be added carefully, but the top-level domain boundary should stay business-centered.

## 9. Canonical Skill Model

### 9.1 Skill

Represents the stable identity of a capability.

Core fields:

- `id`
- `code`
- `name`
- `description`
- `owner_scope` (`platform`, `enterprise`, `partner`)
- `source_type`
- `provider_type`
- `trust_level`
- `status`
- `created_by`
- `updated_by`
- `created_at`
- `updated_at`

### 9.2 Skill Version

Represents a releasable and installable immutable version.

Core fields:

- `id`
- `skill_id`
- `version`
- `manifest_json`
- `input_schema_json`
- `output_schema_json`
- `default_policy_json`
- `runtime_contract_json`
- `checksum`
- `signature`
- `release_status`
- `compatibility_json`
- `change_log`
- `created_by`
- `created_at`

### 9.3 Skill Manifest

The canonical descriptor used by DotBlue after internal creation or external normalization.

Suggested sections:

- identity
- metadata
- source
- auth requirements
- input schema
- output schema
- references
- policy defaults
- observability settings
- compatibility constraints
- signature and provenance

### 9.4 Skill Reference

Represents a directed reference from one skill version to another.

Core fields:

- `id`
- `from_skill_version_id`
- `to_skill_version_id`
- `invoke_mode`
- `condition_expr`
- `context_passthrough`
- `result_passthrough`
- `created_by`
- `created_at`

### 9.5 Skill Policy

Represents default or scope-specific governance rules.

Core fields:

- `risk_level`
- `approval_policy`
- `channel_scope`
- `data_scope`
- `quota_policy`
- `timeout_policy`
- `retry_policy`
- `audit_level`
- `network_policy`
- `masking_policy`

## 10. Skill Source and Hub Model

### 10.1 Supported Skill Sources

Initial source types:

- `builtin`
- `enterprise_private`
- `partner`
- `openapi_catalog`
- `mcp_catalog`
- `package_registry`

Future source type:

- `marketplace`

### 10.2 Skill Hub

A hub is a platform-configured source registry for skills.

Core fields:

- `id`
- `hub_code`
- `name`
- `hub_type`
- `base_url`
- `auth_scheme`
- `auth_secret_ref`
- `trust_level`
- `sync_mode`
- `import_policy`
- `allowed_namespaces`
- `network_policy`
- `signature_policy`
- `status`

### 10.3 Initial Hub Types

Phase 1 and 2 should focus on:

- `builtin_hub`
- `enterprise_private_hub`
- `openapi_hub`
- `mcp_hub`

### 10.4 Why Hubs Must Be Platform-Configurable

Platform configuration is required because hub registration changes the platform security and governance boundary:

- credentials must be managed centrally
- trust level must be reviewed centrally
- import policy must be enforced consistently
- network access must be controlled
- enterprise visibility must be limited deliberately

## 11. Skill Creation and Import Flows

### 11.1 Create from Scratch

Used for:

- built-in platform skills
- enterprise-authored private skills

Flow:

1. create draft skill identity
2. define manifest and IO schema
3. define default policy
4. run validation
5. save draft version
6. submit for review

### 11.2 Create from Template

Used for:

- customer support assistant skills
- sales assistant skills
- operations assistant skills
- knowledge assistant skills

Flow:

1. select template
2. clone manifest scaffold
3. tailor policy and integration config
4. validate and save as draft

### 11.3 Import from Hub

Used for:

- OpenAPI endpoints
- MCP catalogs
- partner skill registries

Flow:

1. select hub
2. select source artifact or capability
3. parse external descriptor
4. normalize into canonical manifest
5. verify schema, signature, and provenance
6. run dependency and risk scans
7. sandbox test if required
8. save as draft version
9. submit for review

### 11.4 Import Is Not Publish

The system must keep these stages separated:

- import
- normalize
- review
- publish
- enable
- install

This separation is essential for enterprise governance.

## 12. Lifecycle State Model

### 12.1 Skill State

Suggested state machine:

- `draft`
- `reviewing`
- `published`
- `deprecated`
- `disabled`
- `archived`

### 12.2 Enablement State

Enterprise-level state:

- `disabled`
- `enabled`
- `suspended`

### 12.3 Binding State

Agent-level installation state:

- `pending`
- `installed`
- `suspended`
- `removed`

### 12.4 Release Channel

Recommended release channels:

- `candidate`
- `stable`

The platform can later extend to `beta` or tenant-specific canary channels.

## 13. Governance Model

### 13.1 Governance Levels

The system must support policy decisions at five levels:

- platform
- enterprise
- organization
- agent
- channel

### 13.2 Governance Dimensions

Each skill must be governable by:

- trust
- risk
- access
- approval
- quota
- timeout
- retry
- observability
- release
- deprecation

### 13.3 Trust Levels

Recommended trust levels:

- `platform_trusted`
- `partner_verified`
- `enterprise_verified`
- `unverified`
- `blocked`

Suggested rules:

- `platform_trusted` may enter stable release directly after review
- `partner_verified` may be published with stronger monitoring defaults
- `enterprise_verified` may be restricted to one enterprise scope
- `unverified` must not be installed on production agents
- `blocked` must not be executed

### 13.4 Risk Levels

Recommended risk levels:

- `low`
- `medium`
- `high`

Risk should consider:

- data sensitivity
- write side effects
- external messaging
- identity or enterprise admin actions
- network egress

### 13.5 Approval Policies

Recommended approval modes:

- `none`
- `user_confirm`
- `enterprise_admin_approve`
- `dual_control`

High-risk skills should not default to silent automatic invocation.

## 14. Enterprise Enablement

Before a skill can be installed on an agent, it must be enabled at enterprise scope.

Enterprise enablement controls:

- whether the skill is available in this enterprise
- which org units can access it
- whether policy overrides are applied
- whether enterprise review is required before installation
- whether the skill is allowed on IM channels

Core entity:

- `skill_enablement`

Core fields:

- `enterprise_id`
- `skill_id`
- `enabled`
- `org_scope_json`
- `policy_override_json`
- `review_status`
- `enabled_by`
- `enabled_at`

## 15. Installing Skills on Agents

### 15.1 Why Installation Must Be Explicit

An agent should not gain a new capability only because a skill exists or is enabled globally.

Installation is an explicit binding step that defines:

- which version the agent uses
- which name or label it exposes
- whether invocation is automatic or suggested
- whether the binding overrides approval or channel visibility

### 15.2 Binding Entity

Core fields:

- `agent_id`
- `skill_id`
- `skill_version_id`
- `binding_status`
- `entry_alias`
- `invoke_visibility`
- `priority`
- `policy_override_json`
- `installed_by`
- `installed_at`

### 15.3 Installation Flow

1. select an enabled skill
2. select version or release channel
3. verify enterprise access
4. verify agent scope and compatibility
5. resolve dependency graph
6. evaluate effective policy
7. persist binding
8. activate agent binding

### 15.4 Installation Guardrails

The platform must deny installation when:

- the skill is not enabled in the enterprise
- the selected version is disabled or deprecated beyond policy
- the dependency graph contains a cycle
- the required channels are forbidden
- the trust level is below enterprise policy
- approval prerequisites are missing

## 16. Effective Policy Resolution

Every invocation must calculate effective policy in a predictable order.

Recommended precedence from low to high:

1. platform default
2. skill version default
3. enterprise enablement override
4. agent binding override
5. runtime context override such as approval result or channel restriction

The output of the resolver should be a runtime object:

- `EffectiveSkillPolicy`

This object is the only policy source the runtime should consume.

## 17. Skill-to-Skill References

### 17.1 Equal Skills, Directed References

A skill may reference another skill, but this does not create a hierarchy. It only creates a governed directed dependency.

### 17.2 Cycle Prevention

The reference graph must always be a DAG at the version level.

The platform must enforce cycle checks:

- when adding or updating a reference
- when publishing a version
- when installing on an agent
- again at runtime using call path protection

### 17.3 Runtime Safety

Runtime must carry:

- `trace_id`
- `root_execution_id`
- `call_path`
- `depth`

Invocation must stop if:

- the next skill version already exists in `call_path`
- maximum depth is exceeded
- maximum fan-out is exceeded

### 17.4 Suggested Limits

- maximum depth: `5`
- maximum call nodes per top-level execution: `20`
- repeated re-entry into the same version: forbidden

## 18. Runtime Architecture

### 18.1 Runtime Abstractions

To keep dependency inversion intact, runtime should depend on interfaces, not source-specific implementations.

Suggested abstractions:

- `SkillResolver`
- `SkillPolicyEvaluator`
- `SkillInvoker`
- `SkillReferenceWalker`
- `SkillExecutionRecorder`
- `SkillAuditRecorder`

### 18.2 Invoker Implementations

Implementations may include:

- `NativeSkillInvoker`
- `OpenAPISkillInvoker`
- `MCPSkillInvoker`
- `RemoteHostedSkillInvoker`

These should be injected behind interfaces.

### 18.3 Runtime Responsibilities

For each invocation:

1. resolve the bound skill and version
2. compute effective policy
3. validate actor, enterprise, channel, and approval context
4. allocate trace context
5. invoke through the right adapter
6. record execution
7. emit audit events
8. resolve downstream references if needed
9. enforce timeout, retry, and circuit-breaking rules

## 19. Audit and Observability

### 19.1 Execution Record

Execution records are for runtime visibility and troubleshooting.

Suggested fields:

- `id`
- `trace_id`
- `root_execution_id`
- `parent_execution_id`
- `skill_id`
- `skill_version_id`
- `agent_id`
- `enterprise_id`
- `conversation_id`
- `channel`
- `status`
- `duration_ms`
- `error_code`
- `input_digest`
- `output_digest`
- `started_at`
- `ended_at`

### 19.2 Audit Record

Audit records are for compliance and governance.

Suggested fields:

- `id`
- `execution_id`
- `actor_type`
- `actor_id`
- `action`
- `resource_scope`
- `decision`
- `approval_ref`
- `masked_payload_json`
- `created_at`

### 19.3 Observability Requirements

The platform should provide:

- execution timeline view
- trace tree view
- skill failure dashboard
- risk and approval dashboard
- suspicious outbound activity alerts

## 20. Integration with Existing DotBlue Domains

### 20.1 Agent Domain

`agent` remains the user-facing assistant entity. The agent should not own capability logic directly. Instead, it owns installed skill bindings plus agent-specific prompt and model configuration.

### 20.2 Chat Domain

Chat remains the conversation orchestration domain. It should evolve from runtime tool event display to skill execution timeline display.

### 20.3 IM Domain

IM remains the channel ingress domain. IM routing should forward into skill resolution and invocation while respecting channel-specific policy.

### 20.4 Settings Domain

Platform-level hub registration, trust defaults, release defaults, and risk defaults should be configured here or through a dedicated platform skill management area.

## 21. Frontend Information Architecture

Suggested management surfaces:

### 21.1 Platform Admin

- `Skill Hubs`
- `Skill Catalog`
- `Skill Review`
- `Skill Releases`
- `Skill Executions`
- `Skill Audit`

### 21.2 Enterprise Admin

- `Enabled Skills`
- `Org Scope`
- `Policy Overrides`
- `Approval Policies`

### 21.3 Agent Configuration

- `Installed Skills`
- `Install Skill`
- `Version Pinning`
- `Invocation Visibility`
- `Binding Overrides`

### 21.4 Runtime Monitoring

- `Execution Timeline`
- `Trace Graph`
- `Failure Analysis`
- `Approval Queue`

## 22. API Design Guidelines

Suggested endpoints:

- `POST /api/platform/skill-hubs`
- `GET /api/platform/skill-hubs`
- `POST /api/platform/skill-import-jobs`
- `GET /api/platform/skill-import-jobs/{id}`
- `POST /api/platform/skills`
- `GET /api/platform/skills`
- `GET /api/platform/skills/{id}`
- `POST /api/platform/skills/{id}/versions`
- `POST /api/platform/skills/{id}/submit-review`
- `POST /api/platform/skills/{id}/publish`
- `POST /api/platform/skills/{id}/disable`
- `POST /api/platform/skills/{id}/rollback`
- `POST /api/enterprises/{id}/skills/{skillId}/enable`
- `POST /api/enterprises/{id}/skills/{skillId}/disable`
- `GET /api/agents/{id}/skills`
- `POST /api/agents/{id}/skills/install`
- `POST /api/agents/{id}/skills/{skillId}/uninstall`
- `GET /api/platform/skill-executions`
- `GET /api/platform/skill-audits`

The APIs should remain domain-oriented and avoid mixing core business rules into handlers.

## 23. Security Design

### 23.1 Required Controls

- signature validation for imported artifacts when available
- provenance tracking for all imported skills
- centralized secret references for hub auth
- masked audit payloads for sensitive data
- explicit network policy for remote skill invocation
- timeout and circuit breaking for remote skill execution

### 23.2 Trust-Based Restrictions

Examples:

- `unverified` skills cannot be installed on production agents
- high-risk external skills cannot auto-invoke on IM channels
- blocked skills must be denied at publish, install, and runtime stages

## 24. Suggested Database Tables

- `skill`
- `skill_version`
- `skill_reference`
- `skill_hub`
- `skill_import_job`
- `skill_enablement`
- `agent_skill_binding`
- `skill_execution`
- `skill_audit`
- `skill_release_record`

If later needed for performance:

- `skill_dependency_closure`

This optional closure table can speed up cycle detection and dependency analysis at scale.

## 25. Testing Strategy

The most valuable phase 1 and phase 2 tests are:

- lifecycle state transition tests
- manifest validation tests
- hub import normalization tests
- policy resolution tests
- cycle detection tests
- installation eligibility tests
- runtime trace propagation tests
- approval and denial path tests

Tests should focus on business behavior and regression risk instead of implementation noise.

## 26. Phase Plan

### Phase 1: Core Skill Control Plane

Deliver:

- canonical skill model
- built-in skill creation
- versioning
- review and publish flow
- enterprise enablement
- agent installation
- execution and audit records

### Phase 2: External Hub Integration

Deliver:

- hub registry
- OpenAPI adapter
- MCP adapter
- import job pipeline
- trust and signature validation baseline

### Phase 3: Reference Graph and Runtime Governance

Deliver:

- skill reference graph
- cycle detection
- runtime call-path protection
- trace graph visualization
- stronger timeout and circuit-breaking controls

### Phase 4: Enterprise Governance Expansion

Deliver:

- org-level scope controls
- advanced approval modes
- release channels
- partner hubs
- policy override UI

### Phase 5: Ecosystem Expansion

Deliver:

- partner package registry
- marketplace support
- enterprise skill SDK
- automated upgrade assistant

## 27. Recommended First Deliverables

The most pragmatic first implementation items are:

1. add a new `skill` domain in backend
2. add database schema for skill, version, enablement, binding, execution, and audit
3. build platform admin skill catalog pages
4. let enterprises enable skills
5. let agents install pinned skill versions
6. upgrade chat execution display from generic tool events to skill execution records

## 28. Summary

DotBlue should adopt a unified skill system where every capability is governed as a first-class `Skill`.

The correct strategic path is not marketplace-first and not prompt-only capability growth. The correct path is:

- define one canonical skill model
- let multiple external ecosystems enter through adapters
- govern all skills through one control plane
- install skills explicitly onto agents
- execute them under resolved policy
- audit them through one traceable runtime model

This gives DotBlue a credible long-term foundation for enterprise capability governance, private deployment, partner integration, and future ecosystem expansion.
