package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"dotblue/internal/domains/agent"
	"dotblue/internal/domains/chat"
	"dotblue/internal/domains/conversation"
	"dotblue/internal/domains/engine"
	"dotblue/internal/domains/enterprise"
	"dotblue/internal/domains/identity"
	"dotblue/internal/domains/im"
	"dotblue/internal/domains/setup"
	"dotblue/internal/infrastructure/dbschema"
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

			if err := dbschema.Ensure(ctx); err != nil {
				g.Log().Fatalf(ctx, "Failed to initialize database schema: %v", err)
			}
			g.Log().Info(ctx, "Database schema ready")
			if err := setup.TryAutoInstall(ctx); err != nil {
				g.Log().Fatalf(ctx, "Automatic setup failed: %v", err)
			}
			if err := startEmbeddedWorkerIfEnabled(ctx); err != nil {
				g.Log().Fatalf(ctx, "Failed to start embedded worker: %v", err)
			}

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
				group.POST("/im/inbound/{platform}/{id}", im.PlatformInboundHandler)
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
				group.GET("/admin/im/connections/{id}/bindings", im.ListConnectionBindingsHandler)
				group.POST("/admin/im/connections/{id}/bindings", im.CreateConnectionBindingHandler)
				group.GET("/admin/im/bindings/{bindingId}", im.GetBindingHandler)
				group.PUT("/admin/im/bindings/{bindingId}", im.UpdateBindingHandler)
				group.DELETE("/admin/im/bindings/{bindingId}", im.DeleteBindingHandler)
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
				group.POST("/chat/completions", im.WebChatCompletionsHandler)
			})

			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Bind()
			})
			s.Run()
			return nil
		},
	}
)

func init() {
	_ = Main.AddCommand(&Worker)
}
