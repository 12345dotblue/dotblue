package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"dotblue/internal/controller/hello"
	"dotblue/internal/domains/agent"
	"dotblue/internal/domains/chat"
	"dotblue/internal/domains/conversation"
	"dotblue/internal/domains/engine"
	"dotblue/internal/domains/enterprise"
	"dotblue/internal/domains/identity"
	"dotblue/internal/domains/im"
	"dotblue/internal/domains/setup"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			// 初始化 Casdoor
			var casdoorConfig identity.Config
			if err := g.Cfg().MustGet(ctx, "casdoor").Struct(&casdoorConfig); err != nil {
				g.Log().Fatalf(ctx, "Failed to load casdoor config: %v", err)
			}
			identity.Init(casdoorConfig)

			engine.Init()
			// 检查数据库连接，连不上则报错并退出
			if err := g.DB().PingMaster(); err != nil {
				g.Log().Fatalf(ctx, "Failed to connect to database: %v", err)
			}
			g.Log().Info(ctx, "Database connected successfully")

			// 自动建表：sys_settings（单行全局配置表）
			if _, err := g.DB().Exec(ctx, `
				CREATE TABLE IF NOT EXISTS sys_settings (
					initialized BOOLEAN DEFAULT FALSE,
					platform    JSONB DEFAULT '{}',
					provider    JSONB DEFAULT '{}',
					updated_at  TIMESTAMP DEFAULT NOW()
				)
			`); err != nil {
				g.Log().Fatalf(ctx, "Failed to create sys_settings table: %v", err)
			}
			g.DB().Exec(ctx, `INSERT INTO sys_settings (initialized) SELECT false WHERE NOT EXISTS (SELECT 1 FROM sys_settings)`)
			g.Log().Info(ctx, "sys_settings table ready")

			// 自动建表：agents（用户智能体，UUID 主键，每用户可多个）
			if _, err := g.DB().Exec(ctx, `
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
			`); err != nil {
				g.Log().Fatalf(ctx, "Failed to create agents table: %v", err)
			}
			g.Log().Info(ctx, "agents table ready")
			g.DB().Exec(ctx, `ALTER TABLE agents ADD COLUMN IF NOT EXISTS engine_type VARCHAR(64) DEFAULT 'hermes'`)

			// 自动建表：conversations（UUIDv7 主键，依赖 PostgreSQL 18+ 内置 uuidv7()）
			if _, err := g.DB().Exec(ctx, `
				CREATE TABLE IF NOT EXISTS conversations (
					id          UUID PRIMARY KEY DEFAULT uuidv7(),
					user_id     VARCHAR(128) NOT NULL,
					group_id    VARCHAR(128) NOT NULL,
					agent_id    UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
					title       VARCHAR(256) DEFAULT '',
					created_at  TIMESTAMPTZ DEFAULT NOW(),
					updated_at  TIMESTAMPTZ DEFAULT NOW()
				)
			`); err != nil {
				g.Log().Fatalf(ctx, "Failed to create conversations table: %v", err)
			}
			g.DB().Exec(ctx, `ALTER TABLE conversations ADD COLUMN IF NOT EXISTS source_type VARCHAR(24) DEFAULT 'web'`)
			g.DB().Exec(ctx, `ALTER TABLE conversations ADD COLUMN IF NOT EXISTS source_connection_id UUID NULL`)
			g.DB().Exec(ctx, `ALTER TABLE conversations ADD COLUMN IF NOT EXISTS source_chat_id VARCHAR(255) DEFAULT ''`)
			g.DB().Exec(ctx, `ALTER TABLE conversations ADD COLUMN IF NOT EXISTS source_thread_id VARCHAR(255) DEFAULT ''`)
			g.DB().Exec(ctx, `ALTER TABLE conversations ADD COLUMN IF NOT EXISTS source_user_id VARCHAR(255) DEFAULT ''`)
			g.DB().Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_conversations_user_updated ON conversations(user_id, updated_at DESC)`)
			g.DB().Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_conversations_source_type ON conversations(source_type)`)
			g.DB().Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_conversations_source_connection ON conversations(source_connection_id)`)
			g.Log().Info(ctx, "conversations table ready")

			// 自动建表：messages（UUIDv7 主键，JSONB tool_calls）
			if _, err := g.DB().Exec(ctx, `
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
			`); err != nil {
				g.Log().Fatalf(ctx, "Failed to create messages table: %v", err)
			}
			g.DB().Exec(ctx, `ALTER TABLE messages ADD COLUMN IF NOT EXISTS source_message_id VARCHAR(255) DEFAULT ''`)
			g.DB().Exec(ctx, `ALTER TABLE messages ADD COLUMN IF NOT EXISTS delivery_status VARCHAR(24) DEFAULT 'none'`)
			g.DB().Exec(ctx, `ALTER TABLE messages ADD COLUMN IF NOT EXISTS message_meta_json JSONB DEFAULT '{}'`)
			// UUIDv7 配合 created_at 排序可保持时间顺序稳定，conversation_id + id 复合索引覆盖消息分页查询
			g.DB().Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_messages_conv_id ON messages(conversation_id, id ASC)`)
			g.DB().Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_messages_source_message_id ON messages(source_message_id)`)
			g.Log().Info(ctx, "messages table ready")

			if _, err := g.DB().Exec(ctx, `
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
			`); err != nil {
				g.Log().Fatalf(ctx, "Failed to create users table: %v", err)
			}
			g.Log().Info(ctx, "users table ready")

			if _, err := g.DB().Exec(ctx, `
				CREATE TABLE IF NOT EXISTS enterprises (
					id          VARCHAR(128) PRIMARY KEY,
					name        VARCHAR(256) NOT NULL,
					slug        VARCHAR(256) DEFAULT '',
					status      VARCHAR(32) DEFAULT 'active',
					created_by  VARCHAR(128) NOT NULL,
					created_at  TIMESTAMPTZ DEFAULT NOW(),
					updated_at  TIMESTAMPTZ DEFAULT NOW()
				)
			`); err != nil {
				g.Log().Fatalf(ctx, "Failed to create enterprises table: %v", err)
			}
			g.DB().Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_enterprises_created_by ON enterprises(created_by)`)
			g.Log().Info(ctx, "enterprises table ready")

			if _, err := g.DB().Exec(ctx, `
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
			`); err != nil {
				g.Log().Fatalf(ctx, "Failed to create enterprise_members table: %v", err)
			}
			g.DB().Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_enterprise_members_user ON enterprise_members(user_id)`)
			g.Log().Info(ctx, "enterprise_members table ready")

			if _, err := g.DB().Exec(ctx, `
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
			`); err != nil {
				g.Log().Fatalf(ctx, "Failed to create org_units table: %v", err)
			}
			g.DB().Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_org_units_enterprise_parent ON org_units(enterprise_id, parent_id)`)
			g.Log().Info(ctx, "org_units table ready")

			if _, err := g.DB().Exec(ctx, `
				CREATE TABLE IF NOT EXISTS org_unit_member (
					id            UUID PRIMARY KEY DEFAULT uuidv7(),
					enterprise_id VARCHAR(128) NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE,
					org_unit_id    UUID NOT NULL REFERENCES org_units(id) ON DELETE CASCADE,
					user_id        VARCHAR(128) NOT NULL,
					is_primary     BOOLEAN DEFAULT FALSE,
					created_at     TIMESTAMPTZ DEFAULT NOW(),
					UNIQUE(enterprise_id, org_unit_id, user_id)
				)
			`); err != nil {
				g.Log().Fatalf(ctx, "Failed to create org_unit_member table: %v", err)
			}
			g.DB().Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_org_unit_member_user ON org_unit_member(enterprise_id, user_id)`)
			g.Log().Info(ctx, "org_unit_member table ready")

			if _, err := g.DB().Exec(ctx, `
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
			`); err != nil {
				g.Log().Fatalf(ctx, "Failed to create enterprise_invitations table: %v", err)
			}
			g.DB().Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_enterprise_invitations_enterprise ON enterprise_invitations(enterprise_id, status)`)
			g.Log().Info(ctx, "enterprise_invitations table ready")

			if _, err := g.DB().Exec(ctx, `
				CREATE TABLE IF NOT EXISTS user_enterprise_sessions (
					user_id            VARCHAR(128) PRIMARY KEY,
					last_enterprise_id VARCHAR(128) NOT NULL,
					updated_at         TIMESTAMPTZ DEFAULT NOW()
				)
			`); err != nil {
				g.Log().Fatalf(ctx, "Failed to create user_enterprise_sessions table: %v", err)
			}
			g.Log().Info(ctx, "user_enterprise_sessions table ready")

			if _, err := g.DB().Exec(ctx, `
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
			`); err != nil {
				g.Log().Fatalf(ctx, "Failed to create im_connections table: %v", err)
			}
			g.DB().Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_im_connections_enterprise ON im_connections(enterprise_id)`)
			g.DB().Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_im_connections_enterprise_platform ON im_connections(enterprise_id, platform)`)
			g.DB().Exec(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS uk_im_connections_enterprise_name ON im_connections(enterprise_id, name)`)
			g.Log().Info(ctx, "im_connections table ready")

			if _, err := g.DB().Exec(ctx, `
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
			`); err != nil {
				g.Log().Fatalf(ctx, "Failed to create agent_channel_bindings table: %v", err)
			}
			g.DB().Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_agent_channel_bindings_enterprise ON agent_channel_bindings(enterprise_id)`)
			g.DB().Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_agent_channel_bindings_agent ON agent_channel_bindings(agent_id)`)
			g.DB().Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_agent_channel_bindings_connection ON agent_channel_bindings(connection_id)`)
			g.DB().Exec(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS uk_agent_channel_bindings_agent_connection ON agent_channel_bindings(agent_id, connection_id)`)
			g.Log().Info(ctx, "agent_channel_bindings table ready")

			if _, err := g.DB().Exec(ctx, `
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
			`); err != nil {
				g.Log().Fatalf(ctx, "Failed to create external_sessions table: %v", err)
			}
			g.DB().Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_external_sessions_enterprise ON external_sessions(enterprise_id)`)
			g.DB().Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_external_sessions_conversation ON external_sessions(conversation_id)`)
			g.DB().Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_external_sessions_connection ON external_sessions(connection_id)`)
			g.DB().Exec(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS uk_external_sessions_session_key ON external_sessions(session_key)`)
			g.Log().Info(ctx, "external_sessions table ready")

			if _, err := g.DB().Exec(ctx, `
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
			`); err != nil {
				g.Log().Fatalf(ctx, "Failed to create external_message_events table: %v", err)
			}
			g.DB().Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_external_message_events_enterprise ON external_message_events(enterprise_id)`)
			g.DB().Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_external_message_events_connection ON external_message_events(connection_id)`)
			g.DB().Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_external_message_events_created_at ON external_message_events(created_at DESC)`)
			g.DB().Exec(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS uk_external_message_events_connection_event ON external_message_events(connection_id, event_id)`)
			g.Log().Info(ctx, "external_message_events table ready")

			if _, err := g.DB().Exec(ctx, `
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
			`); err != nil {
				g.Log().Fatalf(ctx, "Failed to create channel_delivery_logs table: %v", err)
			}
			g.DB().Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_channel_delivery_logs_connection ON channel_delivery_logs(connection_id)`)
			g.DB().Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_channel_delivery_logs_conversation ON channel_delivery_logs(conversation_id)`)
			g.DB().Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_channel_delivery_logs_created_at ON channel_delivery_logs(created_at DESC)`)
			g.Log().Info(ctx, "channel_delivery_logs table ready")

			s := g.Server()
			// Allow CORS from the frontend dev server
			s.Use(func(r *ghttp.Request) {
				r.Response.Header().Set("Access-Control-Allow-Origin", "*")
				r.Response.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				r.Response.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				if r.Method == "OPTIONS" {
					r.Response.WriteStatus(204)
					r.ExitAll()
					return
				}
				r.Middleware.Next()
			})

			// 公开路由（无鉴权）
			s.Group("/api", func(group *ghttp.RouterGroup) {
				group.POST("/signin", identity.SigninHandler)
				group.GET("/setup/status", setup.StatusHandler)
				group.POST("/setup/install", setup.InstallHandler)
				group.POST("/im/inbound/feishu/{id}", im.FeishuInboundHandler)
			})

			s.Group("/api", func(group *ghttp.RouterGroup) {
				group.Middleware(identity.Middleware)
				group.POST("/invitations/{code}/accept", enterprise.AcceptInvitationHandler)
			})

			// 平台管理员路由（平台级设置）
			s.Group("/api", func(group *ghttp.RouterGroup) {
				group.Middleware(identity.Middleware)
				group.Middleware(identity.AdminMiddleware)
				group.GET("/admin/settings", engine.GetSettingsHandler)
				group.POST("/admin/settings", engine.SettingsHandler)
			})

			// 企业管理员路由
			s.Group("/api", func(group *ghttp.RouterGroup) {
				group.Middleware(identity.Middleware)
				group.Middleware(enterprise.MemberContextMiddleware)
				group.Middleware(enterprise.AdminMiddleware)
				group.GET("/admin/summary", enterprise.GetSummaryHandler)
				group.GET("/admin/org-units", enterprise.ListOrgUnitsHandler)
				group.POST("/admin/org-units", enterprise.CreateOrgUnitHandler)
				group.PUT("/admin/org-units/{id}", enterprise.UpdateOrgUnitHandler)
				group.DELETE("/admin/org-units/{id}", enterprise.DeleteOrgUnitHandler)
				group.GET("/admin/members", enterprise.ListMembersHandler)
				group.GET("/admin/users/search", enterprise.SearchUsersHandler)
				group.POST("/admin/members/add-existing", enterprise.AddExistingMemberHandler)
				group.PUT("/admin/members/{userId}/role", enterprise.UpdateMemberRoleHandler)
				group.PUT("/admin/members/{userId}/org-unit", enterprise.UpdateMemberOrgUnitHandler)
				group.GET("/admin/invitations", enterprise.ListInvitationsHandler)
				group.POST("/admin/invitations", enterprise.CreateInvitationHandler)
				group.GET("/admin/im/connections", im.ListConnectionsHandler)
				group.POST("/admin/im/connections", im.CreateConnectionHandler)
				group.GET("/admin/im/connections/{id}", im.GetConnectionHandler)
				group.PUT("/admin/im/connections/{id}", im.UpdateConnectionHandler)
				group.POST("/admin/im/connections/{id}/test", im.TestConnectionHandler)
				group.POST("/admin/im/connections/{id}/enable", im.EnableConnectionHandler)
				group.POST("/admin/im/connections/{id}/disable", im.DisableConnectionHandler)
				group.GET("/admin/im/connections/{id}/events", im.ListConnectionEventsHandler)
				group.GET("/admin/im/connections/{id}/deliveries", im.ListConnectionDeliveriesHandler)
			})

			// 普通用户路由（需要认证 + 企业上下文）
			s.Group("/api", func(group *ghttp.RouterGroup) {
				group.Middleware(identity.Middleware)
				group.Middleware(enterprise.MemberContextMiddleware)
				group.GET("/enterprises", enterprise.ListEnterprisesHandler)
				group.GET("/enterprises/current", enterprise.GetCurrentEnterpriseHandler)
				group.POST("/enterprises", enterprise.CreateEnterpriseHandler)
				group.POST("/enterprises/switch", enterprise.SwitchEnterpriseHandler)
				// Agent CRUD
				group.GET("/agents", agent.ListHandler)
				group.POST("/agents", agent.CreateHandler)
				group.GET("/agents/{id}", agent.GetHandler)
				group.PUT("/agents/{id}", agent.UpdateHandler)
				group.DELETE("/agents/{id}", agent.DeleteHandler)
				// Conversations
				group.GET("/conversations", conversation.ListHandler)
				group.POST("/conversations", conversation.CreateHandler)
				group.GET("/conversations/{id}", conversation.GetHandler)
				group.PUT("/conversations/{id}", conversation.UpdateHandler)
				group.DELETE("/conversations/{id}", conversation.DeleteHandler)
				group.GET("/conversations/{id}/messages", conversation.ListMessagesHandler)
				// Chat
				group.GET("/chat", chat.Handler)
				group.POST("/chat/completions", chat.CompletionsHandler)
			})

			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Bind(
					hello.NewV1(),
				)
			})
			s.Run()
			return nil
		},
	}
)
