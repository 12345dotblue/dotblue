package dbschema

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
)

const schemaMigrationsTable = "schema_migrations"

var excludedBaselineStatements = map[string]struct{}{
	"create credit wallets unique index":           {},
	"delete duplicate credit grant ledger entries": {},
	"delete duplicate credit grants":               {},
	"create credit grants source ref unique index": {},
	"create credit budget policies unique index":   {},
	"create chat_entry_agent_configs table":        {},
	"create chat_entry_share_links table":          {},
	"create chat_entry_share_links agent index":    {},
	"create chat_entry_embed_configs table":        {},
	"create chat_entry_access_logs table":          {},
	"create chat_entry_access_logs target index":   {},
}

func applyPostgresMigrations(ctx context.Context, db gdb.DB) error {
	if err := ensureSchemaMigrationsTable(ctx, db); err != nil {
		return err
	}
	applied, err := loadAppliedMigrations(ctx, db)
	if err != nil {
		return err
	}
	for _, item := range postgresMigrations() {
		if applied[item.version] {
			continue
		}
		if err := applyPostgresMigration(ctx, db, item); err != nil {
			return fmt.Errorf("apply migration %s (%s): %w", item.version, item.name, err)
		}
	}
	return nil
}

func ensureSchemaMigrationsTable(ctx context.Context, db gdb.DB) error {
	return execStatements(ctx, db, []statement{{
		name: "create schema migrations table",
		sql: `
			CREATE TABLE IF NOT EXISTS schema_migrations (
				version    VARCHAR(64) PRIMARY KEY,
				name       VARCHAR(255) NOT NULL,
				applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			)
		`,
	}})
}

func loadAppliedMigrations(ctx context.Context, db gdb.DB) (map[string]bool, error) {
	type row struct {
		Version string `json:"version"`
	}
	var rows []row
	if err := db.Model(schemaMigrationsTable).Ctx(ctx).Order("version ASC").Scan(&rows); err != nil {
		return nil, fmt.Errorf("load schema migrations: %w", err)
	}
	result := make(map[string]bool, len(rows))
	for _, item := range rows {
		result[item.Version] = true
	}
	return result, nil
}

func applyPostgresMigration(ctx context.Context, db gdb.DB, item migration) error {
	if item.transactional {
		return db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
			if err := execStatementsTx(ctx, tx, item.statements); err != nil {
				return err
			}
			return recordMigration(ctx, tx.Model(schemaMigrationsTable).Ctx(ctx), item)
		})
	}
	if err := execStatements(ctx, db, item.statements); err != nil {
		return err
	}
	return recordMigration(ctx, db.Model(schemaMigrationsTable).Ctx(ctx), item)
}

func postgresMigrations() []migration {
	return []migration{
		{
			version:       "2026061201",
			name:          "baseline_schema",
			transactional: false,
			statements:    postgresBaselineStatements(),
		},
		{
			version:       "2026061202",
			name:          "sys_settings_singleton",
			transactional: true,
			statements:    postgresSysSettingsMigrationStatements(),
		},
		{
			version:       "2026061203",
			name:          "credits_dedupe_constraints",
			transactional: true,
			statements:    postgresCreditsMigrationStatements(),
		},
		{
			version:       "2026061204",
			name:          "chatentry_schema",
			transactional: false,
			statements:    postgresChatEntryMigrationStatements(),
		},
	}
}

func postgresBaselineStatements() []statement {
	all := postgresSchemaStatements()
	result := make([]statement, 0, len(all))
	for _, item := range all {
		if _, excluded := excludedBaselineStatements[item.name]; excluded {
			continue
		}
		result = append(result, item)
	}
	return result
}

func postgresSysSettingsMigrationStatements() []statement {
	return []statement{
		{
			name: "ensure sys_settings id column",
			sql:  `ALTER TABLE sys_settings ADD COLUMN IF NOT EXISTS id SMALLINT`,
		},
		{
			name: "collapse sys_settings to singleton row",
			sql: `
				WITH keeper AS (
					SELECT ctid
					FROM sys_settings
					ORDER BY initialized DESC, updated_at DESC NULLS LAST, ctid DESC
					LIMIT 1
				)
				DELETE FROM sys_settings
				WHERE EXISTS (SELECT 1 FROM keeper)
				  AND ctid NOT IN (SELECT ctid FROM keeper)
			`,
		},
		{
			name: "seed sys_settings singleton row",
			sql: `
				INSERT INTO sys_settings (id, initialized, platform, provider, updated_at)
				SELECT 1, FALSE, '{}'::jsonb, '{}'::jsonb, NOW()
				WHERE NOT EXISTS (SELECT 1 FROM sys_settings)
			`,
		},
		{
			name: "normalize sys_settings singleton row",
			sql: `
				UPDATE sys_settings
				SET
					id = 1,
					initialized = COALESCE(initialized, FALSE),
					platform = COALESCE(platform, '{}'::jsonb),
					provider = COALESCE(provider, '{}'::jsonb),
					updated_at = COALESCE(updated_at, NOW())
			`,
		},
		{
			name: "set sys_settings id default",
			sql:  `ALTER TABLE sys_settings ALTER COLUMN id SET DEFAULT 1`,
		},
		{
			name: "set sys_settings id not null",
			sql:  `ALTER TABLE sys_settings ALTER COLUMN id SET NOT NULL`,
		},
		{
			name: "add sys_settings primary key",
			sql: `
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1
						FROM pg_constraint
						WHERE conrelid = 'sys_settings'::regclass
						  AND conname = 'pk_sys_settings'
					) THEN
						ALTER TABLE sys_settings ADD CONSTRAINT pk_sys_settings PRIMARY KEY (id);
					END IF;
				END $$;
			`,
		},
		{
			name: "add sys_settings singleton check",
			sql: `
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1
						FROM pg_constraint
						WHERE conrelid = 'sys_settings'::regclass
						  AND conname = 'ck_sys_settings_singleton'
					) THEN
						ALTER TABLE sys_settings ADD CONSTRAINT ck_sys_settings_singleton CHECK (id = 1);
					END IF;
				END $$;
			`,
		},
	}
}

func postgresCreditsMigrationStatements() []statement {
	return []statement{
		{
			name: "ensure credit ledger entries table before cleanup",
			sql: `
				CREATE TABLE IF NOT EXISTS credit_ledger_entries (
					id                 VARCHAR(128) PRIMARY KEY,
					enterprise_id      VARCHAR(128) NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE,
					credit_type        VARCHAR(32) NOT NULL,
					wallet_id          VARCHAR(128) NOT NULL REFERENCES credit_wallets(id) ON DELETE CASCADE,
					grant_id           VARCHAR(128) DEFAULT '',
					entry_type         VARCHAR(32) NOT NULL,
					direction          VARCHAR(16) NOT NULL,
					credits            BIGINT NOT NULL DEFAULT 0,
					balance_after      BIGINT NOT NULL DEFAULT 0,
					reserved_after     BIGINT NOT NULL DEFAULT 0,
					invocation_id      VARCHAR(128) DEFAULT '',
					member_user_id     VARCHAR(128) DEFAULT '',
					agent_id           UUID NULL REFERENCES agents(id) ON DELETE SET NULL,
					budget_scope_type  VARCHAR(32) DEFAULT '',
					budget_scope_id    VARCHAR(128) DEFAULT '',
					reason_code        VARCHAR(64) DEFAULT '',
					snapshot_json      JSONB DEFAULT '{}'::jsonb,
					created_at         TIMESTAMPTZ DEFAULT NOW()
				)
			`,
		},
		{
			name: "dedupe credit wallets into canonical rows",
			sql: `
				CREATE TEMP TABLE tmp_credit_wallet_groups ON COMMIT DROP AS
				WITH ranked AS (
					SELECT
						id,
						enterprise_id,
						credit_type,
						ROW_NUMBER() OVER (
							PARTITION BY enterprise_id, credit_type
							ORDER BY created_at ASC, id ASC
						) AS rn,
						FIRST_VALUE(id) OVER (
							PARTITION BY enterprise_id, credit_type
							ORDER BY created_at ASC, id ASC
						) AS canonical_id
					FROM credit_wallets
				)
				SELECT * FROM ranked
			`,
		},
		{
			name: "merge canonical credit wallets",
			sql: `
				WITH wallet_agg AS (
					SELECT
						canonical_id,
						SUM(w.total_credits) AS total_credits,
						SUM(w.reserved_credits) AS reserved_credits,
						MAX(w.version) AS max_version,
						MIN(w.created_at) AS min_created_at,
						MAX(w.updated_at) AS max_updated_at
					FROM tmp_credit_wallet_groups grp
					JOIN credit_wallets w ON w.id = grp.id
					GROUP BY canonical_id
				)
				UPDATE credit_wallets w
				SET
					total_credits = agg.total_credits,
					reserved_credits = agg.reserved_credits,
					available_credits = agg.total_credits - agg.reserved_credits,
					version = GREATEST(w.version, agg.max_version),
					created_at = LEAST(w.created_at, agg.min_created_at),
					updated_at = GREATEST(w.updated_at, agg.max_updated_at)
				FROM wallet_agg agg
				WHERE w.id = agg.canonical_id
			`,
		},
		{
			name: "rewire ledger to canonical wallets",
			sql: `
				UPDATE credit_ledger_entries le
				SET wallet_id = grp.canonical_id
				FROM tmp_credit_wallet_groups grp
				WHERE grp.rn > 1
				  AND le.wallet_id = grp.id
			`,
		},
		{
			name: "delete duplicate credit wallets",
			sql: `
				DELETE FROM credit_wallets w
				USING tmp_credit_wallet_groups grp
				WHERE grp.rn > 1
				  AND w.id = grp.id
			`,
		},
		{
			name: "create credit wallets singleton index",
			sql:  `CREATE UNIQUE INDEX IF NOT EXISTS uk_credit_wallets_enterprise_type ON credit_wallets(enterprise_id, credit_type)`,
		},
		{
			name: "prepare duplicate credit grants map",
			sql: `
				CREATE TEMP TABLE tmp_credit_grant_groups ON COMMIT DROP AS
				WITH ranked AS (
					SELECT
						id,
						enterprise_id,
						credit_type,
						source_type,
						source_ref_id,
						ROW_NUMBER() OVER (
							PARTITION BY enterprise_id, credit_type, source_type, source_ref_id
							ORDER BY created_at ASC, id ASC
						) AS rn,
						FIRST_VALUE(id) OVER (
							PARTITION BY enterprise_id, credit_type, source_type, source_ref_id
							ORDER BY created_at ASC, id ASC
						) AS canonical_id
					FROM credit_grants
					WHERE COALESCE(source_ref_id, '') <> ''
				)
				SELECT * FROM ranked
			`,
		},
		{
			name: "merge canonical credit grants",
			sql: `
				WITH grant_agg AS (
					SELECT
						grp.canonical_id,
						SUM(g.granted_credits) AS granted_credits,
						SUM(g.remaining_credits) AS remaining_credits,
						MIN(g.effective_at) AS effective_at,
						MAX(g.expires_at) AS expires_at,
						MIN(g.created_at) AS created_at
					FROM tmp_credit_grant_groups grp
					JOIN credit_grants g ON g.id = grp.id
					GROUP BY grp.canonical_id
				)
				UPDATE credit_grants g
				SET
					granted_credits = agg.granted_credits,
					remaining_credits = agg.remaining_credits,
					effective_at = agg.effective_at,
					expires_at = agg.expires_at,
					created_at = agg.created_at
				FROM grant_agg agg
				WHERE g.id = agg.canonical_id
			`,
		},
		{
			name: "mark duplicate grant ledger history as reconciled",
			sql: `
				UPDATE credit_ledger_entries le
				SET
					entry_type = 'grant_reconciled',
					reason_code = CASE
						WHEN COALESCE(le.reason_code, '') = '' THEN 'migration_duplicate_grant'
						ELSE le.reason_code
					END,
					snapshot_json = COALESCE(le.snapshot_json, '{}'::jsonb) ||
						jsonb_build_object(
							'migration',
							jsonb_build_object(
								'action', 'duplicate_grant_reconciled',
								'canonicalGrantId', grp.canonical_id,
								'duplicateGrantId', grp.id
							)
						)
				FROM tmp_credit_grant_groups grp
				WHERE grp.rn > 1
				  AND le.grant_id = grp.id
				  AND le.entry_type = 'grant'
			`,
		},
		{
			name: "rewire non grant ledger rows to canonical grants",
			sql: `
				UPDATE credit_ledger_entries le
				SET
					grant_id = grp.canonical_id,
					snapshot_json = COALESCE(le.snapshot_json, '{}'::jsonb) ||
						jsonb_build_object(
							'migration',
							jsonb_build_object(
								'action', 'duplicate_grant_merged',
								'canonicalGrantId', grp.canonical_id,
								'duplicateGrantId', grp.id
							)
						)
				FROM tmp_credit_grant_groups grp
				WHERE grp.rn > 1
				  AND le.grant_id = grp.id
				  AND le.entry_type <> 'grant'
			`,
		},
		{
			name: "delete duplicate credit grants",
			sql: `
				DELETE FROM credit_grants g
				USING tmp_credit_grant_groups grp
				WHERE grp.rn > 1
				  AND g.id = grp.id
			`,
		},
		{
			name: "create credit grants source ref unique index",
			sql:  `CREATE UNIQUE INDEX IF NOT EXISTS uk_credit_grants_source_ref ON credit_grants(enterprise_id, credit_type, source_type, source_ref_id) WHERE COALESCE(source_ref_id, '') <> ''`,
		},
		{
			name: "prepare duplicate credit budget policy map",
			sql: `
				CREATE TEMP TABLE tmp_credit_budget_policy_groups ON COMMIT DROP AS
				WITH ranked AS (
					SELECT
						id,
						enterprise_id,
						credit_type,
						scope_type,
						scope_id,
						ROW_NUMBER() OVER (
							PARTITION BY enterprise_id, credit_type, scope_type, scope_id
							ORDER BY enabled DESC, updated_at DESC, created_at DESC, id DESC
						) AS rn
					FROM credit_budget_policies
				)
				SELECT * FROM ranked
			`,
		},
		{
			name: "delete duplicate credit budget policies",
			sql: `
				DELETE FROM credit_budget_policies p
				USING tmp_credit_budget_policy_groups grp
				WHERE grp.rn > 1
				  AND p.id = grp.id
			`,
		},
		{
			name: "create credit budget policies unique index",
			sql:  `CREATE UNIQUE INDEX IF NOT EXISTS uk_credit_budget_policies_scope ON credit_budget_policies(enterprise_id, credit_type, scope_type, scope_id)`,
		},
	}
}

func postgresChatEntryMigrationStatements() []statement {
	return []statement{
		{
			name: "create chat_entry_agent_configs table",
			sql: `
				CREATE TABLE IF NOT EXISTS chat_entry_agent_configs (
					id                     UUID PRIMARY KEY DEFAULT uuidv7(),
					enterprise_id          VARCHAR(128) NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE,
					agent_id               UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
					enabled                BOOLEAN NOT NULL DEFAULT FALSE,
					default_access_mode    VARCHAR(24) NOT NULL DEFAULT 'standalone',
					allow_anonymous        BOOLEAN NOT NULL DEFAULT FALSE,
					allow_file_upload      BOOLEAN NOT NULL DEFAULT FALSE,
					theme_mode             VARCHAR(24) NOT NULL DEFAULT 'auto',
					compact_header         BOOLEAN NOT NULL DEFAULT FALSE,
					session_ttl_seconds    INTEGER NOT NULL DEFAULT 900,
					refresh_before_seconds INTEGER NOT NULL DEFAULT 120,
					created_by             VARCHAR(128) NOT NULL DEFAULT '',
					created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					UNIQUE (enterprise_id, agent_id)
				)
			`,
		},
		{
			name: "create chat_entry_share_links table",
			sql: `
				CREATE TABLE IF NOT EXISTS chat_entry_share_links (
					id                  UUID PRIMARY KEY DEFAULT uuidv7(),
					enterprise_id       VARCHAR(128) NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE,
					agent_id            UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
					conversation_id     UUID NULL REFERENCES conversations(id) ON DELETE SET NULL,
					share_code          VARCHAR(128) NOT NULL UNIQUE,
					password_hash       VARCHAR(255) NOT NULL DEFAULT '',
					status              VARCHAR(24) NOT NULL DEFAULT 'active',
					allow_continue_chat BOOLEAN NOT NULL DEFAULT FALSE,
					allow_anonymous     BOOLEAN NOT NULL DEFAULT FALSE,
					max_access_count    INTEGER NOT NULL DEFAULT 0,
					access_count        INTEGER NOT NULL DEFAULT 0,
					expires_at          TIMESTAMPTZ NULL,
					revoked_at          TIMESTAMPTZ NULL,
					created_by          VARCHAR(128) NOT NULL DEFAULT '',
					created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`,
		},
		{
			name: "create chat_entry_share_links agent index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_chat_entry_share_links_agent_created_at ON chat_entry_share_links(enterprise_id, agent_id, created_at DESC)`,
		},
		{
			name: "create chat_entry_embed_configs table",
			sql: `
				CREATE TABLE IF NOT EXISTS chat_entry_embed_configs (
					id                   UUID PRIMARY KEY DEFAULT uuidv7(),
					enterprise_id        VARCHAR(128) NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE,
					agent_id             UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
					allowed_origins_json JSONB NOT NULL DEFAULT '[]'::jsonb,
					theme_mode           VARCHAR(24) NOT NULL DEFAULT 'auto',
					compact_header       BOOLEAN NOT NULL DEFAULT TRUE,
					allow_file_upload    BOOLEAN NOT NULL DEFAULT FALSE,
					created_by           VARCHAR(128) NOT NULL DEFAULT '',
					created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					UNIQUE (enterprise_id, agent_id)
				)
			`,
		},
		{
			name: "create chat_entry_access_logs table",
			sql: `
				CREATE TABLE IF NOT EXISTS chat_entry_access_logs (
					id              UUID PRIMARY KEY DEFAULT uuidv7(),
					channel_type    VARCHAR(24) NOT NULL,
					target_id       VARCHAR(128) NOT NULL DEFAULT '',
					enterprise_id   VARCHAR(128) NOT NULL DEFAULT '',
					agent_id        VARCHAR(128) NOT NULL DEFAULT '',
					origin          TEXT NOT NULL DEFAULT '',
					referer         TEXT NOT NULL DEFAULT '',
					ip_hash         VARCHAR(128) NOT NULL DEFAULT '',
					user_agent      TEXT NOT NULL DEFAULT '',
					trace_id        VARCHAR(128) NOT NULL DEFAULT '',
					result          VARCHAR(32) NOT NULL DEFAULT '',
					risk_flags_json JSONB NOT NULL DEFAULT '[]'::jsonb,
					created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`,
		},
		{
			name: "create chat_entry_access_logs target index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_chat_entry_access_logs_target_created_at ON chat_entry_access_logs(channel_type, target_id, created_at DESC)`,
		},
	}
}
