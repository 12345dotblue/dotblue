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
			name: "create messages conversation index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_messages_conv_id ON messages(conversation_id, id ASC)`,
		},
		{
			name: "create messages source_message_id index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_messages_source_message_id ON messages(source_message_id)`,
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
