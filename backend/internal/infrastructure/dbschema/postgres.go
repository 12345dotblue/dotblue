package dbschema

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

type postgresProvider struct{}

func (postgresProvider) Ensure(ctx context.Context, db gdb.DB) error {
	if err := execStatements(ctx, db, postgresSchemaStatements()); err != nil {
		return err
	}
	if err := ensureSysSettingsSeed(ctx, db); err != nil {
		return err
	}
	return nil
}

func ensureSysSettingsSeed(ctx context.Context, db gdb.DB) error {
	count, err := db.Model("sys_settings").Ctx(ctx).Count()
	if err != nil {
		return fmt.Errorf("count sys_settings rows: %w", err)
	}
	if count > 0 {
		return nil
	}
	if _, err := db.Model("sys_settings").Ctx(ctx).Data(g.Map{
		"initialized": false,
		"platform":    "{}",
		"provider":    "{}",
	}).Insert(); err != nil {
		return fmt.Errorf("seed sys_settings default row: %w", err)
	}
	return nil
}

func postgresSchemaStatements() []statement {
	return []statement{
		{
			name: "create sys_settings table",
			sql: `
				CREATE TABLE IF NOT EXISTS sys_settings (
					initialized BOOLEAN DEFAULT FALSE,
					platform    JSONB DEFAULT '{}',
					provider    JSONB DEFAULT '{}',
					updated_at  TIMESTAMP DEFAULT NOW()
				)
			`,
		},
		{
			name: "create agents table",
			sql: `
				CREATE TABLE IF NOT EXISTS agents (
					id             UUID PRIMARY KEY DEFAULT uuidv7(),
					user_id        VARCHAR(128) NOT NULL,
					group_id       VARCHAR(128) NOT NULL,
					agent_name     VARCHAR(256) NOT NULL,
					system_prompt  TEXT,
					hermes_api_key VARCHAR(256) NOT NULL,
					engine_type    VARCHAR(64) DEFAULT 'hermes',
					created_at     TIMESTAMP DEFAULT NOW(),
					updated_at     TIMESTAMP DEFAULT NOW()
				)
			`,
		},
		{
			name: "ensure agents engine_type column",
			sql:  `ALTER TABLE agents ADD COLUMN IF NOT EXISTS engine_type VARCHAR(64) DEFAULT 'hermes'`,
		},
		{
			name: "ensure agents model_scope column",
			sql:  `ALTER TABLE agents ADD COLUMN IF NOT EXISTS model_scope VARCHAR(32) DEFAULT ''`,
		},
		{
			name: "ensure agents model_id column",
			sql:  `ALTER TABLE agents ADD COLUMN IF NOT EXISTS model_id VARCHAR(128) DEFAULT ''`,
		},
		{
			name: "create enterprise llm models table",
			sql: `
				CREATE TABLE IF NOT EXISTS enterprise_llm_models (
					id             VARCHAR(128) PRIMARY KEY,
					enterprise_id  VARCHAR(128) NOT NULL,
					display_name   VARCHAR(256) NOT NULL,
					provider_type  VARCHAR(64) NOT NULL,
					api_base       TEXT DEFAULT '',
					api_key        TEXT DEFAULT '',
					model_name     VARCHAR(256) NOT NULL,
					created_at     TIMESTAMPTZ DEFAULT NOW(),
					updated_at     TIMESTAMPTZ DEFAULT NOW()
				)
			`,
		},
		{
			name: "create enterprise llm models enterprise index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_enterprise_llm_models_enterprise ON enterprise_llm_models(enterprise_id, created_at DESC)`,
		},
		{
			name: "create llm models table",
			sql: `
				CREATE TABLE IF NOT EXISTS llm_models (
					id             VARCHAR(128) PRIMARY KEY,
					scope          VARCHAR(32) NOT NULL,
					enterprise_id  VARCHAR(128) DEFAULT '',
					display_name   VARCHAR(256) NOT NULL,
					provider_type  VARCHAR(64) NOT NULL,
					api_base       TEXT DEFAULT '',
					api_key        TEXT DEFAULT '',
					model_name     VARCHAR(256) NOT NULL,
					is_default     BOOLEAN DEFAULT FALSE,
					created_at     TIMESTAMPTZ DEFAULT NOW(),
					updated_at     TIMESTAMPTZ DEFAULT NOW()
				)
			`,
		},
		{
			name: "create llm models scope enterprise index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_llm_models_scope_enterprise ON llm_models(scope, enterprise_id, created_at DESC)`,
		},
		{
			name: "migrate enterprise llm models to unified table",
			sql: `
				INSERT INTO llm_models (
					id, scope, enterprise_id, display_name, provider_type, api_base, api_key, model_name, is_default, created_at, updated_at
				)
				SELECT
					id,
					'enterprise',
					enterprise_id,
					display_name,
					provider_type,
					api_base,
					api_key,
					model_name,
					FALSE,
					created_at,
					updated_at
				FROM enterprise_llm_models e
				WHERE NOT EXISTS (SELECT 1 FROM llm_models m WHERE m.id = e.id)
			`,
		},
		{
			name: "migrate sys provider to platform default model",
			sql: `
				INSERT INTO llm_models (
					id, scope, enterprise_id, display_name, provider_type, api_base, api_key, model_name, is_default, created_at, updated_at
				)
				SELECT
					'platform-default',
					'platform',
					'',
					'平台默认模型',
					COALESCE(provider->>'type', ''),
					COALESCE(provider->>'apiBase', ''),
					COALESCE(provider->>'apiKey', ''),
					COALESCE(provider->>'model', ''),
					TRUE,
					NOW(),
					NOW()
				FROM sys_settings
				WHERE COALESCE(provider->>'type', '') <> ''
				  AND NOT EXISTS (SELECT 1 FROM llm_models m WHERE m.scope = 'platform')
			`,
		},
		{
			name: "create conversations table",
			sql: `
				CREATE TABLE IF NOT EXISTS conversations (
					id          UUID PRIMARY KEY DEFAULT uuidv7(),
					user_id     VARCHAR(128) NOT NULL,
					group_id    VARCHAR(128) NOT NULL,
					agent_id    UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
					title       VARCHAR(256) DEFAULT '',
					created_at  TIMESTAMPTZ DEFAULT NOW(),
					updated_at  TIMESTAMPTZ DEFAULT NOW()
				)
			`,
		},
		{
			name: "ensure conversations source_type column",
			sql:  `ALTER TABLE conversations ADD COLUMN IF NOT EXISTS source_type VARCHAR(24) DEFAULT 'web'`,
		},
		{
			name: "ensure conversations source_connection_id column",
			sql:  `ALTER TABLE conversations ADD COLUMN IF NOT EXISTS source_connection_id UUID NULL`,
		},
		{
			name: "ensure conversations source_chat_id column",
			sql:  `ALTER TABLE conversations ADD COLUMN IF NOT EXISTS source_chat_id VARCHAR(255) DEFAULT ''`,
		},
		{
			name: "ensure conversations source_thread_id column",
			sql:  `ALTER TABLE conversations ADD COLUMN IF NOT EXISTS source_thread_id VARCHAR(255) DEFAULT ''`,
		},
		{
			name: "ensure conversations source_user_id column",
			sql:  `ALTER TABLE conversations ADD COLUMN IF NOT EXISTS source_user_id VARCHAR(255) DEFAULT ''`,
		},
		{
			name: "create conversations user_updated index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_conversations_user_updated ON conversations(user_id, updated_at DESC)`,
		},
		{
			name: "create conversations source_type index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_conversations_source_type ON conversations(source_type)`,
		},
		{
			name: "create conversations source_connection index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_conversations_source_connection ON conversations(source_connection_id)`,
		},
		{
			name: "create messages table",
			sql: `
				CREATE TABLE IF NOT EXISTS messages (
					id              UUID PRIMARY KEY DEFAULT uuidv7(),
					conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
					role            VARCHAR(16) NOT NULL,
					content         TEXT DEFAULT '',
					thinking        TEXT DEFAULT '',
					tool_calls      JSONB DEFAULT '[]',
					status          VARCHAR(16) DEFAULT 'done',
					created_at      TIMESTAMPTZ DEFAULT NOW()
				)
			`,
		},
		{
			name: "ensure messages source_message_id column",
			sql:  `ALTER TABLE messages ADD COLUMN IF NOT EXISTS source_message_id VARCHAR(255) DEFAULT ''`,
		},
		{
			name: "ensure messages delivery_status column",
			sql:  `ALTER TABLE messages ADD COLUMN IF NOT EXISTS delivery_status VARCHAR(24) DEFAULT 'none'`,
		},
		{
			name: "ensure messages message_meta_json column",
			sql:  `ALTER TABLE messages ADD COLUMN IF NOT EXISTS message_meta_json JSONB DEFAULT '{}'`,
		},
		{
			name: "ensure messages parts_json column",
			sql:  `ALTER TABLE messages ADD COLUMN IF NOT EXISTS parts_json JSONB DEFAULT '[]'::jsonb`,
		},
		{
			name: "create messages conversation index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_messages_conv_id ON messages(conversation_id, id ASC)`,
		},
		{
			name: "create messages source_message_id index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_messages_source_message_id ON messages(source_message_id)`,
		},
		{
			name: "create chat_files table",
			sql: `
				CREATE TABLE IF NOT EXISTS chat_files (
					id              UUID PRIMARY KEY DEFAULT uuidv7(),
					user_id         VARCHAR(128) NOT NULL,
					group_id        VARCHAR(128) NOT NULL,
					conversation_id UUID NULL REFERENCES conversations(id) ON DELETE SET NULL,
					storage_type    VARCHAR(32) NOT NULL DEFAULT 'local',
					storage_key     TEXT NOT NULL,
					origin_name     VARCHAR(512) NOT NULL,
					mime_type       VARCHAR(128) NOT NULL,
					size_bytes      BIGINT NOT NULL DEFAULT 0,
					sha256          VARCHAR(128) NOT NULL DEFAULT '',
					width           INTEGER NOT NULL DEFAULT 0,
					height          INTEGER NOT NULL DEFAULT 0,
					kind            VARCHAR(24) NOT NULL DEFAULT 'file',
					status          VARCHAR(24) NOT NULL DEFAULT 'uploaded',
					created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`,
		},
		{
			name: "create chat_files user_created index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_chat_files_user_created ON chat_files(user_id, created_at DESC)`,
		},
		{
			name: "create chat_files conversation index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_chat_files_conversation ON chat_files(conversation_id)`,
		},
		{
			name: "create message_attachments table",
			sql: `
				CREATE TABLE IF NOT EXISTS message_attachments (
					id                   UUID PRIMARY KEY DEFAULT uuidv7(),
					message_id           UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
					file_id              UUID NOT NULL REFERENCES chat_files(id) ON DELETE CASCADE,
					kind                 VARCHAR(24) NOT NULL,
					sort_order           INTEGER NOT NULL DEFAULT 0,
					attachment_meta_json JSONB NOT NULL DEFAULT '{}'::jsonb,
					created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`,
		},
		{
			name: "create message_attachments message index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_message_attachments_message ON message_attachments(message_id, sort_order ASC)`,
		},
		{
			name: "create message_attachments file index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_message_attachments_file ON message_attachments(file_id)`,
		},
		{
			name: "create llm usage events table",
			sql: `
				CREATE TABLE IF NOT EXISTS llm_usage_events (
					id                        VARCHAR(128) PRIMARY KEY,
					invocation_id             VARCHAR(128) NOT NULL UNIQUE,
					request_id                VARCHAR(128) DEFAULT '',
					conversation_id           UUID NULL REFERENCES conversations(id) ON DELETE SET NULL,
					message_id                UUID NULL REFERENCES messages(id) ON DELETE SET NULL,
					agent_id                  UUID NULL REFERENCES agents(id) ON DELETE SET NULL,
					enterprise_id             VARCHAR(128) DEFAULT '',
					user_id                   VARCHAR(128) DEFAULT '',
					source_type               VARCHAR(32) DEFAULT 'web',
					source_connection_id      UUID NULL,
					model_id                  VARCHAR(128) NOT NULL DEFAULT '',
					model_scope               VARCHAR(32) DEFAULT '',
					provider_type             VARCHAR(64) DEFAULT '',
					model_name_snapshot       VARCHAR(256) DEFAULT '',
					status                    VARCHAR(32) NOT NULL DEFAULT 'started',
					usage_source              VARCHAR(32) DEFAULT 'estimated',
					prompt_tokens             BIGINT DEFAULT 0,
					completion_tokens         BIGINT DEFAULT 0,
					reasoning_tokens          BIGINT DEFAULT 0,
					cache_read_tokens         BIGINT DEFAULT 0,
					cache_write_tokens        BIGINT DEFAULT 0,
					total_tokens              BIGINT DEFAULT 0,
					currency                  VARCHAR(16) DEFAULT 'USD',
					cost_input_unit_price     NUMERIC(20, 8) DEFAULT 0,
					cost_output_unit_price    NUMERIC(20, 8) DEFAULT 0,
					charge_input_unit_price   NUMERIC(20, 8) DEFAULT 0,
					charge_output_unit_price  NUMERIC(20, 8) DEFAULT 0,
					cost_amount               NUMERIC(20, 8) DEFAULT 0,
					charge_amount             NUMERIC(20, 8) DEFAULT 0,
					raw_usage_json            JSONB DEFAULT '{}'::jsonb,
					raw_response_meta_json    JSONB DEFAULT '{}'::jsonb,
					error_code                VARCHAR(128) DEFAULT '',
					created_at                TIMESTAMPTZ DEFAULT NOW(),
					completed_at              TIMESTAMPTZ NULL
				)
			`,
		},
		{
			name: "create llm usage events enterprise index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_llm_usage_events_enterprise_created ON llm_usage_events(enterprise_id, created_at DESC)`,
		},
		{
			name: "create llm usage events agent index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_llm_usage_events_agent_created ON llm_usage_events(agent_id, created_at DESC)`,
		},
		{
			name: "create llm usage events conversation index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_llm_usage_events_conversation_created ON llm_usage_events(conversation_id, created_at DESC)`,
		},
		{
			name: "create llm usage events model index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_llm_usage_events_model_created ON llm_usage_events(model_id, created_at DESC)`,
		},
		{
			name: "create llm usage daily aggregates table",
			sql: `
				CREATE TABLE IF NOT EXISTS llm_usage_daily_aggregates (
					id              VARCHAR(128) PRIMARY KEY,
					stat_date       VARCHAR(10) NOT NULL,
					scope_type      VARCHAR(32) NOT NULL,
					scope_id        VARCHAR(128) DEFAULT '',
					source_type     VARCHAR(32) DEFAULT 'web',
					request_count   BIGINT DEFAULT 0,
					total_tokens    BIGINT DEFAULT 0,
					cost_amount     NUMERIC(20, 8) DEFAULT 0,
					charge_amount   NUMERIC(20, 8) DEFAULT 0,
					created_at      TIMESTAMPTZ DEFAULT NOW(),
					updated_at      TIMESTAMPTZ DEFAULT NOW()
				)
			`,
		},
		{
			name: "create llm usage daily aggregates unique index",
			sql:  `CREATE UNIQUE INDEX IF NOT EXISTS uk_llm_usage_daily_aggregates_scope_day_source ON llm_usage_daily_aggregates(stat_date, scope_type, scope_id, source_type)`,
		},
		{
			name: "create llm usage daily aggregates scope index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_llm_usage_daily_aggregates_scope_day ON llm_usage_daily_aggregates(scope_type, scope_id, stat_date DESC)`,
		},
		{
			name: "create llm model prices table",
			sql: `
				CREATE TABLE IF NOT EXISTS llm_model_prices (
					id                        VARCHAR(128) PRIMARY KEY,
					model_id                  VARCHAR(128) NOT NULL,
					scope_type                VARCHAR(32) NOT NULL,
					scope_id                  VARCHAR(128) DEFAULT '',
					currency                  VARCHAR(16) DEFAULT 'USD',
					cost_input_unit_price     NUMERIC(20, 8) DEFAULT 0,
					cost_output_unit_price    NUMERIC(20, 8) DEFAULT 0,
					charge_input_unit_price   NUMERIC(20, 8) DEFAULT 0,
					charge_output_unit_price  NUMERIC(20, 8) DEFAULT 0,
					created_at                TIMESTAMPTZ DEFAULT NOW(),
					updated_at                TIMESTAMPTZ DEFAULT NOW()
				)
			`,
		},
		{
			name: "create llm model prices scope index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_llm_model_prices_scope_model ON llm_model_prices(scope_type, scope_id, model_id, created_at DESC)`,
		},
		{
			name: "create llm usage limit policies table",
			sql: `
				CREATE TABLE IF NOT EXISTS llm_usage_limit_policies (
					id                    VARCHAR(128) PRIMARY KEY,
					scope_type            VARCHAR(32) NOT NULL,
					scope_id              VARCHAR(128) DEFAULT '',
					enabled               BOOLEAN DEFAULT TRUE,
					daily_token_limit     BIGINT DEFAULT 0,
					monthly_token_limit   BIGINT DEFAULT 0,
					daily_charge_limit    NUMERIC(20, 8) DEFAULT 0,
					monthly_charge_limit  NUMERIC(20, 8) DEFAULT 0,
					hard_limit            BOOLEAN DEFAULT TRUE,
					created_at            TIMESTAMPTZ DEFAULT NOW(),
					updated_at            TIMESTAMPTZ DEFAULT NOW()
				)
			`,
		},
		{
			name: "create llm usage limit policies scope index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_llm_usage_limit_policies_scope ON llm_usage_limit_policies(scope_type, scope_id, created_at DESC)`,
		},
		{
			name: "create users table",
			sql: `
				CREATE TABLE IF NOT EXISTS users (
					user_id                VARCHAR(128) PRIMARY KEY,
					email                  VARCHAR(256) DEFAULT '',
					display_name           VARCHAR(256) DEFAULT '',
					avatar                 TEXT DEFAULT '',
					source_organization_id VARCHAR(128) DEFAULT '',
					created_at             TIMESTAMPTZ DEFAULT NOW(),
					updated_at             TIMESTAMPTZ DEFAULT NOW(),
					last_login_at          TIMESTAMPTZ DEFAULT NOW()
				)
			`,
		},
		{
			name: "create enterprises table",
			sql: `
				CREATE TABLE IF NOT EXISTS enterprises (
					id          VARCHAR(128) PRIMARY KEY,
					name        VARCHAR(256) NOT NULL,
					slug        VARCHAR(256) DEFAULT '',
					status      VARCHAR(32) DEFAULT 'active',
					created_by  VARCHAR(128) NOT NULL,
					created_at  TIMESTAMPTZ DEFAULT NOW(),
					updated_at  TIMESTAMPTZ DEFAULT NOW()
				)
			`,
		},
		{
			name: "create enterprises created_by index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_enterprises_created_by ON enterprises(created_by)`,
		},
		{
			name: "create enterprise_members table",
			sql: `
				CREATE TABLE IF NOT EXISTS enterprise_members (
					id            UUID PRIMARY KEY DEFAULT uuidv7(),
					enterprise_id VARCHAR(128) NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE,
					user_id       VARCHAR(128) NOT NULL,
					role          VARCHAR(32) DEFAULT 'member',
					status        VARCHAR(32) DEFAULT 'active',
					joined_at     TIMESTAMPTZ DEFAULT NOW(),
					created_at    TIMESTAMPTZ DEFAULT NOW(),
					updated_at    TIMESTAMPTZ DEFAULT NOW(),
					UNIQUE(enterprise_id, user_id)
				)
			`,
		},
		{
			name: "create enterprise_members user index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_enterprise_members_user ON enterprise_members(user_id)`,
		},
		{
			name: "create org_units table",
			sql: `
				CREATE TABLE IF NOT EXISTS org_units (
					id              UUID PRIMARY KEY DEFAULT uuidv7(),
					enterprise_id   VARCHAR(128) NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE,
					parent_id       UUID NULL REFERENCES org_units(id) ON DELETE CASCADE,
					name            VARCHAR(256) NOT NULL,
					code            VARCHAR(128) DEFAULT '',
					manager_user_id VARCHAR(128) DEFAULT '',
					status          VARCHAR(32) DEFAULT 'active',
					sort_order      INTEGER DEFAULT 0,
					created_at      TIMESTAMPTZ DEFAULT NOW(),
					updated_at      TIMESTAMPTZ DEFAULT NOW()
				)
			`,
		},
		{
			name: "create org_units enterprise_parent index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_org_units_enterprise_parent ON org_units(enterprise_id, parent_id)`,
		},
		{
			name: "create org_unit_member table",
			sql: `
				CREATE TABLE IF NOT EXISTS org_unit_member (
					id            UUID PRIMARY KEY DEFAULT uuidv7(),
					enterprise_id VARCHAR(128) NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE,
					org_unit_id    UUID NOT NULL REFERENCES org_units(id) ON DELETE CASCADE,
					user_id        VARCHAR(128) NOT NULL,
					is_primary     BOOLEAN DEFAULT FALSE,
					created_at     TIMESTAMPTZ DEFAULT NOW(),
					UNIQUE(enterprise_id, org_unit_id, user_id)
				)
			`,
		},
		{
			name: "create org_unit_member user index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_org_unit_member_user ON org_unit_member(enterprise_id, user_id)`,
		},
		{
			name: "create enterprise_invitations table",
			sql: `
				CREATE TABLE IF NOT EXISTS enterprise_invitations (
					id                  UUID PRIMARY KEY DEFAULT uuidv7(),
					enterprise_id       VARCHAR(128) NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE,
					code                VARCHAR(128) NOT NULL UNIQUE,
					email               VARCHAR(256) DEFAULT '',
					role                VARCHAR(32) DEFAULT 'member',
					status              VARCHAR(32) DEFAULT 'pending',
					default_org_unit_id UUID NULL REFERENCES org_units(id) ON DELETE SET NULL,
					expires_at          TIMESTAMPTZ NULL,
					max_uses            INTEGER DEFAULT 1,
					used_count          INTEGER DEFAULT 0,
					created_by          VARCHAR(128) NOT NULL,
					accepted_by         VARCHAR(128) DEFAULT '',
					created_at          TIMESTAMPTZ DEFAULT NOW(),
					updated_at          TIMESTAMPTZ DEFAULT NOW()
				)
			`,
		},
		{
			name: "create enterprise_invitations enterprise_status index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_enterprise_invitations_enterprise ON enterprise_invitations(enterprise_id, status)`,
		},
		{
			name: "create user_enterprise_sessions table",
			sql: `
				CREATE TABLE IF NOT EXISTS user_enterprise_sessions (
					user_id            VARCHAR(128) PRIMARY KEY,
					last_enterprise_id VARCHAR(128) NOT NULL,
					updated_at         TIMESTAMPTZ DEFAULT NOW()
				)
			`,
		},
		{
			name: "create im_connections table",
			sql: `
				CREATE TABLE IF NOT EXISTS im_connections (
					id                UUID PRIMARY KEY DEFAULT uuidv7(),
					enterprise_id     VARCHAR(128) NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE,
					platform          VARCHAR(32) NOT NULL,
					name              VARCHAR(128) NOT NULL,
					status            VARCHAR(24) NOT NULL DEFAULT 'disabled',
					connection_mode   VARCHAR(24) NOT NULL,
					config_json       JSONB NOT NULL DEFAULT '{}'::jsonb,
					secret_json       JSONB NOT NULL DEFAULT '{}'::jsonb,
					callback_path     VARCHAR(255) NOT NULL DEFAULT '',
					last_connected_at TIMESTAMPTZ NULL,
					last_error        TEXT NOT NULL DEFAULT '',
					created_by        VARCHAR(128) NOT NULL DEFAULT '',
					created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`,
		},
		{
			name: "create im_connections enterprise index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_im_connections_enterprise ON im_connections(enterprise_id)`,
		},
		{
			name: "create im_connections enterprise_platform index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_im_connections_enterprise_platform ON im_connections(enterprise_id, platform)`,
		},
		{
			name: "create im_connections enterprise_name unique index",
			sql:  `CREATE UNIQUE INDEX IF NOT EXISTS uk_im_connections_enterprise_name ON im_connections(enterprise_id, name)`,
		},
		{
			name: "create agent_channel_bindings table",
			sql: `
				CREATE TABLE IF NOT EXISTS agent_channel_bindings (
					id                  UUID PRIMARY KEY DEFAULT uuidv7(),
					enterprise_id       VARCHAR(128) NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE,
					agent_id            UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
					connection_id       UUID NOT NULL REFERENCES im_connections(id) ON DELETE CASCADE,
					status              VARCHAR(24) NOT NULL DEFAULT 'active',
					trigger_mode        VARCHAR(32) NOT NULL DEFAULT 'mention_only',
					trigger_config_json JSONB NOT NULL DEFAULT '{}'::jsonb,
					session_strategy    VARCHAR(32) NOT NULL DEFAULT 'per_chat_per_user',
					reply_mode          VARCHAR(32) NOT NULL DEFAULT 'default',
					allow_group         BOOLEAN NOT NULL DEFAULT TRUE,
					allow_dm            BOOLEAN NOT NULL DEFAULT TRUE,
					priority            INTEGER NOT NULL DEFAULT 100,
					created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`,
		},
		{
			name: "create agent_channel_bindings enterprise index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_agent_channel_bindings_enterprise ON agent_channel_bindings(enterprise_id)`,
		},
		{
			name: "create agent_channel_bindings agent index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_agent_channel_bindings_agent ON agent_channel_bindings(agent_id)`,
		},
		{
			name: "create agent_channel_bindings connection index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_agent_channel_bindings_connection ON agent_channel_bindings(connection_id)`,
		},
		{
			name: "create agent_channel_bindings agent_connection unique index",
			sql:  `CREATE UNIQUE INDEX IF NOT EXISTS uk_agent_channel_bindings_agent_connection ON agent_channel_bindings(agent_id, connection_id)`,
		},
		{
			name: "create external_sessions table",
			sql: `
				CREATE TABLE IF NOT EXISTS external_sessions (
					id                 UUID PRIMARY KEY DEFAULT uuidv7(),
					enterprise_id      VARCHAR(128) NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE,
					platform           VARCHAR(32) NOT NULL,
					connection_id      UUID NOT NULL REFERENCES im_connections(id) ON DELETE CASCADE,
					binding_id         UUID NOT NULL REFERENCES agent_channel_bindings(id) ON DELETE CASCADE,
					external_chat_id   VARCHAR(255) NOT NULL,
					external_thread_id VARCHAR(255) NOT NULL DEFAULT '',
					external_user_id   VARCHAR(255) NOT NULL DEFAULT '',
					conversation_id    UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
					agent_id           UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
					session_key        VARCHAR(512) NOT NULL,
					last_message_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`,
		},
		{
			name: "create external_sessions enterprise index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_external_sessions_enterprise ON external_sessions(enterprise_id)`,
		},
		{
			name: "create external_sessions conversation index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_external_sessions_conversation ON external_sessions(conversation_id)`,
		},
		{
			name: "create external_sessions connection index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_external_sessions_connection ON external_sessions(connection_id)`,
		},
		{
			name: "create external_sessions session_key unique index",
			sql:  `CREATE UNIQUE INDEX IF NOT EXISTS uk_external_sessions_session_key ON external_sessions(session_key)`,
		},
		{
			name: "create external_message_events table",
			sql: `
				CREATE TABLE IF NOT EXISTS external_message_events (
					id                  UUID PRIMARY KEY DEFAULT uuidv7(),
					enterprise_id       VARCHAR(128) NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE,
					platform            VARCHAR(32) NOT NULL,
					connection_id       UUID NOT NULL REFERENCES im_connections(id) ON DELETE CASCADE,
					event_id            VARCHAR(255) NOT NULL,
					external_message_id VARCHAR(255) NOT NULL DEFAULT '',
					direction           VARCHAR(16) NOT NULL,
					payload_json        JSONB NOT NULL DEFAULT '{}'::jsonb,
					status              VARCHAR(24) NOT NULL DEFAULT 'received',
					error_message       TEXT NOT NULL DEFAULT '',
					created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`,
		},
		{
			name: "create external_message_events enterprise index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_external_message_events_enterprise ON external_message_events(enterprise_id)`,
		},
		{
			name: "create external_message_events connection index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_external_message_events_connection ON external_message_events(connection_id)`,
		},
		{
			name: "create external_message_events created_at index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_external_message_events_created_at ON external_message_events(created_at DESC)`,
		},
		{
			name: "create external_message_events connection_event unique index",
			sql:  `CREATE UNIQUE INDEX IF NOT EXISTS uk_external_message_events_connection_event ON external_message_events(connection_id, event_id)`,
		},
		{
			name: "create channel_delivery_logs table",
			sql: `
				CREATE TABLE IF NOT EXISTS channel_delivery_logs (
					id              UUID PRIMARY KEY DEFAULT uuidv7(),
					enterprise_id   VARCHAR(128) NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE,
					platform        VARCHAR(32) NOT NULL,
					connection_id   UUID NOT NULL REFERENCES im_connections(id) ON DELETE CASCADE,
					conversation_id UUID NULL REFERENCES conversations(id) ON DELETE SET NULL,
					message_id      UUID NULL REFERENCES messages(id) ON DELETE SET NULL,
					attempt         INTEGER NOT NULL DEFAULT 1,
					status          VARCHAR(24) NOT NULL DEFAULT 'pending',
					request_json    JSONB NOT NULL DEFAULT '{}'::jsonb,
					response_json   JSONB NOT NULL DEFAULT '{}'::jsonb,
					error_message   TEXT NOT NULL DEFAULT '',
					created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`,
		},
		{
			name: "create channel_delivery_logs connection index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_channel_delivery_logs_connection ON channel_delivery_logs(connection_id)`,
		},
		{
			name: "create channel_delivery_logs conversation index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_channel_delivery_logs_conversation ON channel_delivery_logs(conversation_id)`,
		},
		{
			name: "create channel_delivery_logs created_at index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_channel_delivery_logs_created_at ON channel_delivery_logs(created_at DESC)`,
		},
	}
}
