# DotBlue Skill Database Design

## 1. Scope

This document refines the logical architecture in [skill-system-design.md](file:///c:/Users/kongz/work/dotblue/docs/skill-system-design.md) into a database-oriented design.

It focuses on:

- table model
- column design
- indexes
- uniqueness and integrity constraints
- state fields
- critical query patterns
- migration and compatibility strategy with current DotBlue tables

This document is intentionally implementation-oriented so backend work can start directly from it.

## 2. Current Schema Conventions in DotBlue

Based on the current PostgreSQL schema:

- most business tables use plural table names
- primary keys are usually `UUID PRIMARY KEY DEFAULT uuidv7()`
- enterprise-scoped domains usually use `enterprise_id VARCHAR(128)` with foreign keys to `enterprises(id)`
- older agent-related data still uses `group_id` as the effective enterprise dimension
- flexible configuration is commonly stored as `JSONB`
- list screens usually need `created_at DESC` indexes

The skill schema should follow these conventions so it fits the existing codebase naturally.

## 3. Core Tenant Model Decision

### 3.1 Canonical Tenant Key

For all new skill tables, use `enterprise_id` as the canonical tenant key.

Reason:

- this matches the current enterprise, IM, invitation, and org tables
- it is clearer than the legacy `group_id`
- it provides a stable long-term model for governance

### 3.2 Compatibility with Existing Agent Table

Current `agents` records still store tenant scope in `group_id`.

Therefore, phase 1 must apply the following compatibility rule:

- `agent_skill_bindings.enterprise_id` is required
- service logic must validate `agent_skill_bindings.enterprise_id == agents.group_id`
- all agent-skill queries should filter by both `agent_id` and `enterprise_id`

This prevents cross-tenant leakage before a future agent schema cleanup.

## 4. Table Overview

Recommended phase 1 and phase 2 tables:

- `skills`
- `skill_versions`
- `skill_references`
- `skill_hubs`
- `skill_import_jobs`
- `enterprise_skill_enablements`
- `agent_skill_bindings`
- `skill_executions`
- `skill_audits`
- `skill_release_records`

Optional later-stage table:

- `skill_dependency_closure`

## 5. Table Design

## 5.1 `skills`

### Purpose

Represents the stable identity of a skill across versions.

### Columns

- `id UUID PRIMARY KEY DEFAULT uuidv7()`
- `code VARCHAR(255) NOT NULL`
- `name VARCHAR(255) NOT NULL`
- `description TEXT NOT NULL DEFAULT ''`
- `owner_scope VARCHAR(32) NOT NULL`
- `owner_enterprise_id VARCHAR(128) NOT NULL DEFAULT ''`
- `source_type VARCHAR(32) NOT NULL`
- `provider_type VARCHAR(32) NOT NULL DEFAULT 'native'`
- `trust_level VARCHAR(32) NOT NULL DEFAULT 'unverified'`
- `status VARCHAR(32) NOT NULL DEFAULT 'draft'`
- `latest_version_id UUID NULL`
- `latest_published_version_id UUID NULL`
- `latest_stable_version_id UUID NULL`
- `tags_json JSONB NOT NULL DEFAULT '[]'::jsonb`
- `metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb`
- `created_by VARCHAR(128) NOT NULL DEFAULT ''`
- `updated_by VARCHAR(128) NOT NULL DEFAULT ''`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

### Semantics

- `owner_scope` values:
  - `platform`
  - `enterprise`
  - `partner`
- `owner_enterprise_id` is empty for platform or partner-owned skills
- `code` is the stable business identity and should be namespace-like, for example `knowledge.search`

### Constraints

- unique skill code inside one ownership scope
- platform-owned skill codes must be globally unique
- enterprise-owned skill codes must be unique inside one enterprise

### Recommended Indexes

- `UNIQUE INDEX uk_skills_owner_scope_owner_enterprise_code ON skills(owner_scope, owner_enterprise_id, code)`
- `INDEX idx_skills_status_created_at ON skills(status, created_at DESC)`
- `INDEX idx_skills_source_type_created_at ON skills(source_type, created_at DESC)`
- `INDEX idx_skills_owner_enterprise ON skills(owner_enterprise_id)`

## 5.2 `skill_versions`

### Purpose

Represents immutable, releasable, installable versions of a skill.

### Columns

- `id UUID PRIMARY KEY DEFAULT uuidv7()`
- `skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE`
- `version VARCHAR(64) NOT NULL`
- `release_channel VARCHAR(32) NOT NULL DEFAULT 'candidate'`
- `release_status VARCHAR(32) NOT NULL DEFAULT 'draft'`
- `manifest_json JSONB NOT NULL DEFAULT '{}'::jsonb`
- `input_schema_json JSONB NOT NULL DEFAULT '{}'::jsonb`
- `output_schema_json JSONB NOT NULL DEFAULT '{}'::jsonb`
- `default_policy_json JSONB NOT NULL DEFAULT '{}'::jsonb`
- `runtime_contract_json JSONB NOT NULL DEFAULT '{}'::jsonb`
- `compatibility_json JSONB NOT NULL DEFAULT '{}'::jsonb`
- `verification_report_json JSONB NOT NULL DEFAULT '{}'::jsonb`
- `risk_report_json JSONB NOT NULL DEFAULT '{}'::jsonb`
- `checksum VARCHAR(255) NOT NULL DEFAULT ''`
- `signature_json JSONB NOT NULL DEFAULT '{}'::jsonb`
- `change_log TEXT NOT NULL DEFAULT ''`
- `published_by VARCHAR(128) NOT NULL DEFAULT ''`
- `published_at TIMESTAMPTZ NULL`
- `created_by VARCHAR(128) NOT NULL DEFAULT ''`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

### Semantics

- `release_status` values:
  - `draft`
  - `reviewing`
  - `published`
  - `deprecated`
  - `disabled`
  - `archived`
- version strings should be human-readable and comparable by service logic, such as semver-like versions

### Constraints

- one version string must be unique inside one skill
- only one stable version should be marked as latest stable in `skills`
- immutable after publish except for non-functional metadata fields if the team chooses to allow that

### Recommended Indexes

- `UNIQUE INDEX uk_skill_versions_skill_version ON skill_versions(skill_id, version)`
- `INDEX idx_skill_versions_skill_created_at ON skill_versions(skill_id, created_at DESC)`
- `INDEX idx_skill_versions_release_status ON skill_versions(release_status, created_at DESC)`
- `INDEX idx_skill_versions_release_channel ON skill_versions(release_channel, created_at DESC)`

## 5.3 `skill_references`

### Purpose

Stores directed skill version dependencies.

### Columns

- `id UUID PRIMARY KEY DEFAULT uuidv7()`
- `from_skill_version_id UUID NOT NULL REFERENCES skill_versions(id) ON DELETE CASCADE`
- `to_skill_version_id UUID NOT NULL REFERENCES skill_versions(id) ON DELETE RESTRICT`
- `invoke_mode VARCHAR(32) NOT NULL DEFAULT 'sync'`
- `condition_expr TEXT NOT NULL DEFAULT ''`
- `context_passthrough BOOLEAN NOT NULL DEFAULT TRUE`
- `result_passthrough BOOLEAN NOT NULL DEFAULT TRUE`
- `sort_order INTEGER NOT NULL DEFAULT 0`
- `created_by VARCHAR(128) NOT NULL DEFAULT ''`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

### Constraints

- self-reference is forbidden
- duplicate reference edges are forbidden
- cycle checks are enforced in service logic and publish/install workflows

### Recommended Indexes

- `UNIQUE INDEX uk_skill_references_from_to ON skill_references(from_skill_version_id, to_skill_version_id)`
- `INDEX idx_skill_references_from ON skill_references(from_skill_version_id, sort_order, created_at)`
- `INDEX idx_skill_references_to ON skill_references(to_skill_version_id)`

## 5.4 `skill_hubs`

### Purpose

Stores platform-managed external or internal skill source registries.

### Columns

- `id UUID PRIMARY KEY DEFAULT uuidv7()`
- `hub_code VARCHAR(128) NOT NULL`
- `name VARCHAR(255) NOT NULL`
- `hub_type VARCHAR(32) NOT NULL`
- `base_url TEXT NOT NULL DEFAULT ''`
- `status VARCHAR(24) NOT NULL DEFAULT 'disabled'`
- `trust_level VARCHAR(32) NOT NULL DEFAULT 'unverified'`
- `sync_mode VARCHAR(24) NOT NULL DEFAULT 'manual'`
- `auth_scheme VARCHAR(32) NOT NULL DEFAULT 'none'`
- `config_json JSONB NOT NULL DEFAULT '{}'::jsonb`
- `secret_json JSONB NOT NULL DEFAULT '{}'::jsonb`
- `import_policy_json JSONB NOT NULL DEFAULT '{}'::jsonb`
- `allowed_namespaces_json JSONB NOT NULL DEFAULT '[]'::jsonb`
- `network_policy_json JSONB NOT NULL DEFAULT '{}'::jsonb`
- `signature_policy_json JSONB NOT NULL DEFAULT '{}'::jsonb`
- `last_synced_at TIMESTAMPTZ NULL`
- `last_error TEXT NOT NULL DEFAULT ''`
- `created_by VARCHAR(128) NOT NULL DEFAULT ''`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

### Constraints

- `hub_code` must be unique

### Recommended Indexes

- `UNIQUE INDEX uk_skill_hubs_code ON skill_hubs(hub_code)`
- `INDEX idx_skill_hubs_type_status ON skill_hubs(hub_type, status)`
- `INDEX idx_skill_hubs_trust_level ON skill_hubs(trust_level)`

## 5.5 `skill_import_jobs`

### Purpose

Tracks one import pipeline execution from a hub into DotBlue.

### Columns

- `id UUID PRIMARY KEY DEFAULT uuidv7()`
- `hub_id UUID NOT NULL REFERENCES skill_hubs(id) ON DELETE CASCADE`
- `requested_by VARCHAR(128) NOT NULL DEFAULT ''`
- `source_locator TEXT NOT NULL DEFAULT ''`
- `source_namespace VARCHAR(255) NOT NULL DEFAULT ''`
- `source_version VARCHAR(128) NOT NULL DEFAULT ''`
- `job_status VARCHAR(32) NOT NULL DEFAULT 'pending'`
- `parsed_descriptor_json JSONB NOT NULL DEFAULT '{}'::jsonb`
- `normalized_manifest_json JSONB NOT NULL DEFAULT '{}'::jsonb`
- `verification_report_json JSONB NOT NULL DEFAULT '{}'::jsonb`
- `risk_report_json JSONB NOT NULL DEFAULT '{}'::jsonb`
- `target_skill_id UUID NULL REFERENCES skills(id) ON DELETE SET NULL`
- `target_skill_version_id UUID NULL REFERENCES skill_versions(id) ON DELETE SET NULL`
- `error_message TEXT NOT NULL DEFAULT ''`
- `started_at TIMESTAMPTZ NULL`
- `finished_at TIMESTAMPTZ NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

### Job Status Values

- `pending`
- `parsing`
- `normalizing`
- `verifying`
- `sandboxing`
- `completed`
- `failed`
- `canceled`

### Recommended Indexes

- `INDEX idx_skill_import_jobs_hub_created_at ON skill_import_jobs(hub_id, created_at DESC)`
- `INDEX idx_skill_import_jobs_status_created_at ON skill_import_jobs(job_status, created_at DESC)`
- `INDEX idx_skill_import_jobs_target_skill ON skill_import_jobs(target_skill_id)`

## 5.6 `enterprise_skill_enablements`

### Purpose

Controls whether a published skill is available inside an enterprise.

### Columns

- `id UUID PRIMARY KEY DEFAULT uuidv7()`
- `enterprise_id VARCHAR(128) NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE`
- `skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE`
- `enablement_status VARCHAR(24) NOT NULL DEFAULT 'enabled'`
- `org_scope_json JSONB NOT NULL DEFAULT '[]'::jsonb`
- `channel_scope_json JSONB NOT NULL DEFAULT '[]'::jsonb`
- `policy_override_json JSONB NOT NULL DEFAULT '{}'::jsonb`
- `review_status VARCHAR(24) NOT NULL DEFAULT 'approved'`
- `review_note TEXT NOT NULL DEFAULT ''`
- `enabled_by VARCHAR(128) NOT NULL DEFAULT ''`
- `enabled_at TIMESTAMPTZ NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

### Constraints

- one enablement record per enterprise and skill

### Recommended Indexes

- `UNIQUE INDEX uk_enterprise_skill_enablements_enterprise_skill ON enterprise_skill_enablements(enterprise_id, skill_id)`
- `INDEX idx_enterprise_skill_enablements_status ON enterprise_skill_enablements(enterprise_id, enablement_status, updated_at DESC)`
- `INDEX idx_enterprise_skill_enablements_skill ON enterprise_skill_enablements(skill_id)`

## 5.7 `agent_skill_bindings`

### Purpose

Represents explicit installation of a skill version onto an agent.

### Columns

- `id UUID PRIMARY KEY DEFAULT uuidv7()`
- `enterprise_id VARCHAR(128) NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE`
- `agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE`
- `skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE`
- `skill_version_id UUID NOT NULL REFERENCES skill_versions(id) ON DELETE RESTRICT`
- `binding_status VARCHAR(24) NOT NULL DEFAULT 'installed'`
- `entry_alias VARCHAR(255) NOT NULL DEFAULT ''`
- `invoke_visibility VARCHAR(24) NOT NULL DEFAULT 'auto'`
- `priority INTEGER NOT NULL DEFAULT 100`
- `policy_override_json JSONB NOT NULL DEFAULT '{}'::jsonb`
- `channel_scope_json JSONB NOT NULL DEFAULT '[]'::jsonb`
- `installed_by VARCHAR(128) NOT NULL DEFAULT ''`
- `installed_at TIMESTAMPTZ NULL`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

### Binding Status Values

- `pending`
- `installed`
- `suspended`
- `removed`

### Constraints

- only one active binding per agent and skill
- for pinned version mode, only one installed record per `(agent_id, skill_id)`

### Recommended Indexes

- `UNIQUE INDEX uk_agent_skill_bindings_agent_skill ON agent_skill_bindings(agent_id, skill_id)`
- `INDEX idx_agent_skill_bindings_enterprise_agent ON agent_skill_bindings(enterprise_id, agent_id)`
- `INDEX idx_agent_skill_bindings_skill_version ON agent_skill_bindings(skill_version_id)`
- `INDEX idx_agent_skill_bindings_status_priority ON agent_skill_bindings(binding_status, priority)`

### Important Compatibility Rule

Service logic must verify:

- the binding enterprise matches the agent `group_id`
- the skill has an enabled record for the same enterprise

This validation is required even if the database already enforces foreign keys.

## 5.8 `skill_executions`

### Purpose

Stores runtime execution records for troubleshooting, performance, and trace graph construction.

### Columns

- `id UUID PRIMARY KEY DEFAULT uuidv7()`
- `trace_id UUID NOT NULL`
- `root_execution_id UUID NULL`
- `parent_execution_id UUID NULL REFERENCES skill_executions(id) ON DELETE CASCADE`
- `enterprise_id VARCHAR(128) NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE`
- `agent_id UUID NULL REFERENCES agents(id) ON DELETE SET NULL`
- `conversation_id UUID NULL REFERENCES conversations(id) ON DELETE SET NULL`
- `message_id UUID NULL REFERENCES messages(id) ON DELETE SET NULL`
- `skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE RESTRICT`
- `skill_version_id UUID NOT NULL REFERENCES skill_versions(id) ON DELETE RESTRICT`
- `binding_id UUID NULL REFERENCES agent_skill_bindings(id) ON DELETE SET NULL`
- `channel VARCHAR(32) NOT NULL DEFAULT 'web'`
- `actor_user_id VARCHAR(128) NOT NULL DEFAULT ''`
- `actor_type VARCHAR(32) NOT NULL DEFAULT 'user'`
- `execution_status VARCHAR(24) NOT NULL DEFAULT 'running'`
- `approval_status VARCHAR(24) NOT NULL DEFAULT 'not_required'`
- `depth INTEGER NOT NULL DEFAULT 0`
- `call_path_json JSONB NOT NULL DEFAULT '[]'::jsonb`
- `input_digest TEXT NOT NULL DEFAULT ''`
- `output_digest TEXT NOT NULL DEFAULT ''`
- `error_code VARCHAR(128) NOT NULL DEFAULT ''`
- `error_message TEXT NOT NULL DEFAULT ''`
- `metrics_json JSONB NOT NULL DEFAULT '{}'::jsonb`
- `started_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `ended_at TIMESTAMPTZ NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

### Execution Status Values

- `running`
- `succeeded`
- `failed`
- `blocked`
- `timed_out`
- `canceled`

### Recommended Indexes

- `INDEX idx_skill_executions_trace ON skill_executions(trace_id, started_at ASC)`
- `INDEX idx_skill_executions_enterprise_started ON skill_executions(enterprise_id, started_at DESC)`
- `INDEX idx_skill_executions_agent_started ON skill_executions(agent_id, started_at DESC)`
- `INDEX idx_skill_executions_skill_started ON skill_executions(skill_id, started_at DESC)`
- `INDEX idx_skill_executions_conversation_started ON skill_executions(conversation_id, started_at DESC)`
- `INDEX idx_skill_executions_status_started ON skill_executions(execution_status, started_at DESC)`

### Notes

- `trace_id` groups one full top-level skill call chain
- `root_execution_id` lets the system query one top-level invocation directly
- `call_path_json` provides runtime safety and easier debugging

## 5.9 `skill_audits`

### Purpose

Stores compliance-oriented and governance-oriented audit records.

### Columns

- `id UUID PRIMARY KEY DEFAULT uuidv7()`
- `execution_id UUID NOT NULL REFERENCES skill_executions(id) ON DELETE CASCADE`
- `enterprise_id VARCHAR(128) NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE`
- `skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE RESTRICT`
- `actor_type VARCHAR(32) NOT NULL DEFAULT 'user'`
- `actor_id VARCHAR(128) NOT NULL DEFAULT ''`
- `action VARCHAR(64) NOT NULL`
- `decision VARCHAR(32) NOT NULL DEFAULT 'allow'`
- `resource_scope_json JSONB NOT NULL DEFAULT '{}'::jsonb`
- `approval_ref VARCHAR(128) NOT NULL DEFAULT ''`
- `masked_payload_json JSONB NOT NULL DEFAULT '{}'::jsonb`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

### Recommended Indexes

- `INDEX idx_skill_audits_execution ON skill_audits(execution_id)`
- `INDEX idx_skill_audits_enterprise_created ON skill_audits(enterprise_id, created_at DESC)`
- `INDEX idx_skill_audits_skill_created ON skill_audits(skill_id, created_at DESC)`
- `INDEX idx_skill_audits_decision_created ON skill_audits(decision, created_at DESC)`

## 5.10 `skill_release_records`

### Purpose

Stores publish, rollback, disable, and channel movement actions.

### Columns

- `id UUID PRIMARY KEY DEFAULT uuidv7()`
- `skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE`
- `skill_version_id UUID NOT NULL REFERENCES skill_versions(id) ON DELETE CASCADE`
- `action VARCHAR(32) NOT NULL`
- `from_status VARCHAR(32) NOT NULL DEFAULT ''`
- `to_status VARCHAR(32) NOT NULL DEFAULT ''`
- `release_channel VARCHAR(32) NOT NULL DEFAULT 'candidate'`
- `scope_json JSONB NOT NULL DEFAULT '{}'::jsonb`
- `note TEXT NOT NULL DEFAULT ''`
- `operated_by VARCHAR(128) NOT NULL DEFAULT ''`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

### Recommended Indexes

- `INDEX idx_skill_release_records_skill_created ON skill_release_records(skill_id, created_at DESC)`
- `INDEX idx_skill_release_records_version_created ON skill_release_records(skill_version_id, created_at DESC)`
- `INDEX idx_skill_release_records_action_created ON skill_release_records(action, created_at DESC)`

## 5.11 Optional `skill_dependency_closure`

### Purpose

Optimizes dependency traversal and cycle checks when the number of skills grows.

### Columns

- `ancestor_skill_version_id UUID NOT NULL REFERENCES skill_versions(id) ON DELETE CASCADE`
- `descendant_skill_version_id UUID NOT NULL REFERENCES skill_versions(id) ON DELETE CASCADE`
- `depth INTEGER NOT NULL`

### Constraints

- `PRIMARY KEY (ancestor_skill_version_id, descendant_skill_version_id)`

### Usage

This table is optional in phase 1. A DFS-based validator is enough initially.

## 6. Suggested Enumerations

To keep state handling stable, service logic should centralize string constants for these fields:

- `skills.owner_scope`
- `skills.source_type`
- `skills.provider_type`
- `skills.trust_level`
- `skills.status`
- `skill_versions.release_channel`
- `skill_versions.release_status`
- `skill_hubs.hub_type`
- `skill_hubs.status`
- `skill_import_jobs.job_status`
- `enterprise_skill_enablements.enablement_status`
- `enterprise_skill_enablements.review_status`
- `agent_skill_bindings.binding_status`
- `agent_skill_bindings.invoke_visibility`
- `skill_executions.execution_status`
- `skill_executions.approval_status`
- `skill_audits.decision`
- `skill_release_records.action`

## 7. Lifecycle and State Transition Rules

## 7.1 Skill and Version

Recommended version transitions:

- `draft -> reviewing`
- `reviewing -> published`
- `published -> deprecated`
- `published -> disabled`
- `deprecated -> disabled`
- `disabled -> archived`

Forbidden examples:

- `draft -> stable production installation`
- `disabled -> installed`
- `archived -> published`

## 7.2 Enterprise Enablement

Recommended transitions:

- `enabled -> suspended`
- `suspended -> enabled`
- `enabled -> disabled`

If a skill is globally disabled, enterprise enablement becomes non-effective even if its record remains enabled.

## 7.3 Agent Binding

Recommended transitions:

- `pending -> installed`
- `installed -> suspended`
- `suspended -> installed`
- `installed -> removed`

`removed` should be terminal from the product perspective.

## 8. Integrity Rules Beyond Foreign Keys

The database cannot enforce all governance rules. Service logic must additionally enforce:

- one skill version cannot reference itself
- no cycle can exist in `skill_references`
- only published versions can be enabled or installed
- only enabled skills can be installed on agents
- unverified skills cannot be installed on production agents
- high-risk skills on IM channels may require approval
- agent binding enterprise must match agent tenant scope
- downstream skill invocation cannot widen upstream policy scope

## 9. Critical Query Patterns

## 9.1 Platform Skill Catalog

Typical filters:

- by `status`
- by `source_type`
- by `trust_level`
- by `owner_scope`
- sorted by `created_at DESC`

Primary tables:

- `skills`
- optional join to `skill_versions`

## 9.2 Enterprise Enabled Skill List

Typical filters:

- by `enterprise_id`
- by `enablement_status`
- optionally by `channel_scope_json`

Primary tables:

- `enterprise_skill_enablements`
- join `skills`
- join latest published version

## 9.3 Agent Installed Skill List

Typical filters:

- by `agent_id`
- by `binding_status`
- sorted by `priority`

Primary tables:

- `agent_skill_bindings`
- join `skills`
- join `skill_versions`

## 9.4 Execution Timeline

Typical filters:

- by `trace_id`
- by `conversation_id`
- by `agent_id`
- by time range

Primary table:

- `skill_executions`

## 9.5 Audit Investigation

Typical filters:

- by `enterprise_id`
- by `skill_id`
- by `actor_id`
- by `decision`
- by time range

Primary table:

- `skill_audits`

## 10. Suggested Write Flows

## 10.1 Create Skill from Scratch

Write sequence:

1. insert `skills`
2. insert `skill_versions`
3. optionally update `skills.latest_version_id`

## 10.2 Publish Skill Version

Write sequence:

1. validate state and dependencies
2. update `skill_versions.release_status`
3. update `skills.latest_published_version_id`
4. if stable, update `skills.latest_stable_version_id`
5. insert `skill_release_records`

## 10.3 Enable Skill for Enterprise

Write sequence:

1. upsert `enterprise_skill_enablements`
2. record audit or admin action log if needed

## 10.4 Install Skill on Agent

Write sequence:

1. validate enablement and version
2. validate no dependency or policy violation
3. upsert `agent_skill_bindings`

## 10.5 Execute Skill

Write sequence:

1. insert `skill_executions` with `running`
2. invoke runtime
3. insert `skill_audits`
4. update `skill_executions` with final status and metrics

## 11. Migration Strategy

## 11.1 Phase 1 Migration

Add the new skill tables without modifying current agent, chat, or IM tables.

This keeps risk low and allows the platform to adopt the skill control plane incrementally.

## 11.2 Phase 2 Integration Migration

Add service-level integration:

- agent install APIs
- enterprise enablement APIs
- runtime execution recording
- chat display changes from tool records to skill execution records

## 11.3 Future Cleanup

After the skill system stabilizes, consider normalizing `agents.group_id` to `enterprise_id`.

That future migration should be handled separately to avoid coupling foundational schema work with legacy cleanup.

## 12. Recommended First SQL Delivery Order

The safest first schema delivery order is:

1. `skills`
2. `skill_versions`
3. `skill_references`
4. `enterprise_skill_enablements`
5. `agent_skill_bindings`
6. `skill_executions`
7. `skill_audits`
8. `skill_hubs`
9. `skill_import_jobs`
10. `skill_release_records`

This order lets the team enable built-in skills before external hub import is fully implemented.

## 13. Summary

The skill database design should be:

- enterprise-aware
- version-first
- policy-friendly
- explicit in installation
- traceable in execution
- extensible for external hubs

The most important architectural choices are:

- use `enterprise_id` as the canonical tenant key for new skill data
- keep agent compatibility through service validation against legacy `group_id`
- treat versions as the real installable and publishable units
- separate enablement, binding, execution, and audit into distinct tables
- use JSONB only for flexible policy and manifest payloads, not for core relational identity

This schema gives DotBlue a durable foundation for a real enterprise skill control plane.
