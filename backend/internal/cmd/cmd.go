package cmd

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"dotblue/internal/domains/agent"
	"dotblue/internal/domains/chat"
	"dotblue/internal/domains/chatentry"
	"dotblue/internal/domains/conversation"
	"dotblue/internal/domains/credit"
	"dotblue/internal/domains/engine"
	"dotblue/internal/domains/enterprise"
	"dotblue/internal/domains/file"
	"dotblue/internal/domains/identity"
	"dotblue/internal/domains/im"
	"dotblue/internal/domains/metering"
	"dotblue/internal/domains/model"
	"dotblue/internal/domains/setup"
	"dotblue/internal/domains/skill"
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
				r.Response.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Enterprise-ID")
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
				group.GET("/public/c-end-chat/share-links/{shareCode}", chatentry.ResolveShareLinkHandler)
				group.POST("/public/c-end-chat/share-links/{shareCode}/verify", chatentry.VerifyShareLinkHandler)
				group.POST("/public/c-end-chat/agents/{agentId}/session", chatentry.CreateStandaloneSessionHandler)
				group.POST("/public/c-end-chat/embed/session", chatentry.ExchangeEmbedSessionHandler)
				group.POST("/public/c-end-chat/session/refresh", chatentry.RefreshSessionHandler)
				group.POST("/public/c-end-chat/conversations", chatentry.CreateConversationHandler)
				group.GET("/public/c-end-chat/conversations/{id}/messages", chatentry.ListConversationMessagesHandler)
				group.POST("/public/c-end-chat/files", chatentry.PublicFileUploadHandler)
				group.GET("/public/c-end-chat/files/{id}/preview", chatentry.PublicFilePreviewHandler)
				group.GET("/public/c-end-chat/files/{id}/download", chatentry.PublicFileDownloadHandler)
				group.POST("/public/c-end-chat/chat/completions", chatentry.PublicChatCompletionsHandler)
			})

			s.Group("/api", func(group *ghttp.RouterGroup) {
				group.Middleware(identity.Middleware)
				group.POST("/invitations/{code}/accept", enterprise.AcceptInvitationHandler)
			})

			// Unified admin skill routes are shared by platform and enterprise admins.
			s.Group("/api", func(group *ghttp.RouterGroup) {
				group.Middleware(identity.Middleware)
				group.Middleware(adminSkillContextMiddleware)
				group.Middleware(adminSkillAccessMiddleware)
				group.GET("/admin/skills", skill.ListEnterpriseSkillsHandler)
				group.POST("/admin/skills", skill.CreateSkillHandler)
				group.GET("/admin/skills/{id}", skill.GetSkillDetailHandler)
				group.POST("/admin/skills/{id}/versions", skill.CreateSkillVersionHandler)
				group.GET("/admin/skills/{id}/versions/{versionId}/references", skill.ListSkillVersionReferencesHandler)
				group.POST("/admin/skills/{id}/references", skill.UpdateSkillVersionReferencesHandler)
				group.POST("/admin/skills/{id}/submit-review", skill.SubmitSkillReviewHandler)
				group.POST("/admin/skills/{id}/publish", skill.PublishSkillHandler)
				group.GET("/admin/skill-hubs", skill.ListSkillHubsHandler)
				group.GET("/admin/skill-import-jobs", skill.ListSkillImportJobsHandler)
				group.POST("/admin/skill-import-jobs", skill.ImportSkillHandler)
			})

			// 平台管理员路由（平台级设置）
			s.Group("/api", func(group *ghttp.RouterGroup) {
				group.Middleware(identity.Middleware)
				group.Middleware(identity.AdminMiddleware)
				group.GET("/admin/settings", engine.GetSettingsHandler)
				group.POST("/admin/settings", engine.SettingsHandler)
				group.GET("/admin/platform/llm-models", model.ListPlatformModelsHandler)
				group.POST("/admin/platform/llm-models", model.CreatePlatformModelHandler)
				group.PUT("/admin/platform/llm-models/{id}", model.UpdatePlatformModelHandler)
				group.DELETE("/admin/platform/llm-models/{id}", model.DeletePlatformModelHandler)
				group.GET("/admin/platform/usage/overview", metering.PlatformUsageOverviewHandler)
				group.GET("/admin/platform/usage/trends", metering.PlatformUsageTrendsHandler)
				group.GET("/admin/platform/usage/events", metering.PlatformUsageEventsHandler)
				group.GET("/admin/platform/credits/overview", credit.PlatformCreditOverviewHandler)
				group.GET("/admin/platform/credits/wallets", credit.PlatformCreditWalletsHandler)
				group.GET("/admin/platform/credits/ledger", credit.PlatformCreditLedgerHandler)
				group.GET("/admin/platform/credits/grants", credit.PlatformCreditGrantsHandler)
				group.POST("/admin/platform/credits/grants", credit.CreatePlatformCreditGrantHandler)
				group.GET("/admin/platform/credit-price-books", credit.ListPlatformCreditPriceBooksHandler)
				group.POST("/admin/platform/credit-price-books", credit.CreatePlatformCreditPriceBookHandler)
				group.PUT("/admin/platform/credit-price-books/{id}", credit.UpdatePlatformCreditPriceBookHandler)
				group.DELETE("/admin/platform/credit-price-books/{id}", credit.DeletePlatformCreditPriceBookHandler)
				group.GET("/admin/platform/model-prices", metering.ListPlatformPricesHandler)
				group.POST("/admin/platform/model-prices", metering.CreatePlatformPriceHandler)
				group.PUT("/admin/platform/model-prices/{id}", metering.UpdatePlatformPriceHandler)
				group.DELETE("/admin/platform/model-prices/{id}", metering.DeletePlatformPriceHandler)
				group.GET("/admin/platform/usage-limit-policies", metering.ListPlatformPoliciesHandler)
				group.POST("/admin/platform/usage-limit-policies", metering.CreatePlatformPolicyHandler)
				group.PUT("/admin/platform/usage-limit-policies/{id}", metering.UpdatePlatformPolicyHandler)
				group.DELETE("/admin/platform/usage-limit-policies/{id}", metering.DeletePlatformPolicyHandler)
				group.GET("/admin/platform/skills", skill.ListPlatformSkillsHandler)
				group.POST("/admin/platform/skills", skill.CreateSkillHandler)
				group.GET("/admin/platform/skills/{id}", skill.GetSkillDetailHandler)
				group.POST("/admin/platform/skills/{id}/versions", skill.CreateSkillVersionHandler)
				group.GET("/admin/platform/skills/{id}/versions/{versionId}/references", skill.ListSkillVersionReferencesHandler)
				group.POST("/admin/platform/skills/{id}/references", skill.UpdateSkillVersionReferencesHandler)
				group.POST("/admin/platform/skills/{id}/submit-review", skill.SubmitSkillReviewHandler)
				group.POST("/admin/platform/skills/{id}/publish", skill.PublishSkillHandler)
				group.GET("/admin/platform/skill-hubs", skill.ListSkillHubsHandler)
				group.POST("/admin/platform/skill-hubs", skill.UpsertSkillHubHandler)
				group.PUT("/admin/platform/skill-hubs/{id}", skill.UpsertSkillHubHandler)
				group.GET("/admin/platform/resource-releases", skill.ListResourceReleasesHandler)
				group.POST("/admin/platform/resource-releases", skill.SetResourceReleaseHandler)
				group.GET("/admin/platform/skill-import-jobs", skill.ListSkillImportJobsHandler)
				group.POST("/admin/platform/skill-import-jobs", skill.ImportSkillHandler)
			})

			// 企业管理员路由
			s.Group("/api", func(group *ghttp.RouterGroup) {
				group.Middleware(identity.Middleware)
				group.Middleware(enterprise.MemberContextMiddleware)
				group.Middleware(enterprise.AdminMiddleware)
				group.GET("/admin/c-end-chat/agents/{agentId}", chatentry.GetAgentConfigHandler)
				group.PUT("/admin/c-end-chat/agents/{agentId}", chatentry.UpsertAgentConfigHandler)
				group.GET("/admin/c-end-chat/agents/{agentId}/share-links", chatentry.ListShareLinksHandler)
				group.POST("/admin/c-end-chat/share-links", chatentry.CreateShareLinkHandler)
				group.POST("/admin/c-end-chat/share-links/{id}/revoke", chatentry.RevokeShareLinkHandler)
				group.GET("/admin/c-end-chat/agents/{agentId}/embed-config", chatentry.GetEmbedConfigHandler)
				group.PUT("/admin/c-end-chat/agents/{agentId}/embed-config", chatentry.UpsertEmbedConfigHandler)
				group.POST("/admin/c-end-chat/agents/{agentId}/embed-token", chatentry.CreateEmbedTokenHandler)
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
				group.GET("/admin/llm-models", model.ListEnterpriseModelsHandler)
				group.POST("/admin/llm-models", model.CreateEnterpriseModelHandler)
				group.PUT("/admin/llm-models/{id}", model.UpdateEnterpriseModelHandler)
				group.DELETE("/admin/llm-models/{id}", model.DeleteEnterpriseModelHandler)
				group.GET("/admin/usage/overview", metering.EnterpriseUsageOverviewHandler)
				group.GET("/admin/usage/trends", metering.EnterpriseUsageTrendsHandler)
				group.GET("/admin/usage/events", metering.EnterpriseUsageEventsHandler)
				group.GET("/admin/credits/overview", credit.EnterpriseCreditOverviewHandler)
				group.GET("/admin/credits/wallets", credit.EnterpriseCreditWalletsHandler)
				group.GET("/admin/credits/ledger", credit.EnterpriseCreditLedgerHandler)
				group.GET("/admin/credits/grants", credit.EnterpriseCreditGrantsHandler)
				group.POST("/admin/credits/grants", credit.CreateEnterpriseCreditGrantHandler)
				group.GET("/admin/credit-price-books", credit.ListEnterpriseCreditPriceBooksHandler)
				group.POST("/admin/credit-price-books", credit.CreateEnterpriseCreditPriceBookHandler)
				group.PUT("/admin/credit-price-books/{id}", credit.UpdateEnterpriseCreditPriceBookHandler)
				group.DELETE("/admin/credit-price-books/{id}", credit.DeleteEnterpriseCreditPriceBookHandler)
				group.GET("/admin/credit-budget-policies", credit.ListEnterpriseCreditBudgetPoliciesHandler)
				group.POST("/admin/credit-budget-policies", credit.CreateEnterpriseCreditBudgetPolicyHandler)
				group.PUT("/admin/credit-budget-policies/{id}", credit.UpdateEnterpriseCreditBudgetPolicyHandler)
				group.DELETE("/admin/credit-budget-policies/{id}", credit.DeleteEnterpriseCreditBudgetPolicyHandler)
				group.GET("/admin/llm-model-prices", metering.ListEnterprisePricesHandler)
				group.POST("/admin/llm-model-prices", metering.CreateEnterprisePriceHandler)
				group.PUT("/admin/llm-model-prices/{id}", metering.UpdateEnterprisePriceHandler)
				group.DELETE("/admin/llm-model-prices/{id}", metering.DeleteEnterprisePriceHandler)
				group.GET("/admin/usage-limit-policies", metering.ListEnterprisePoliciesHandler)
				group.POST("/admin/usage-limit-policies", metering.CreateEnterprisePolicyHandler)
				group.PUT("/admin/usage-limit-policies/{id}", metering.UpdateEnterprisePolicyHandler)
				group.DELETE("/admin/usage-limit-policies/{id}", metering.DeleteEnterprisePolicyHandler)
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
				group.POST("/admin/skills/{skillId}/enable", skill.EnableSkillHandler)
				group.POST("/admin/skills/{skillId}/disable", skill.DisableSkillHandler)
				group.GET("/admin/agents/{agentId}/skills", skill.ListAgentSkillsHandler)
				group.GET("/admin/agents/{agentId}/skill-catalog", skill.ListAgentSkillCatalogHandler)
				group.POST("/admin/agents/{agentId}/skills/install", skill.InstallSkillOnAgentHandler)
				group.POST("/admin/agents/{agentId}/skills/ensure-installed", skill.EnsureSkillOnAgentHandler)
				group.POST("/admin/agents/{agentId}/skills/{skillId}/uninstall", skill.UninstallSkillFromAgentHandler)
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
				group.GET("/agents/model-options", agent.ModelOptionsHandler)
				group.POST("/agents", agent.CreateHandler)
				group.GET("/agents/{id}", agent.GetHandler)
				group.GET("/agents/{id}/skills", skill.ListAgentSkillsForMemberHandler)
				group.GET("/agents/{id}/usage/overview", metering.AgentUsageOverviewHandler)
				group.GET("/agents/{id}/usage/trends", metering.AgentUsageTrendsHandler)
				group.PUT("/agents/{id}", agent.UpdateHandler)
				group.DELETE("/agents/{id}", agent.DeleteHandler)
				// Conversations
				group.GET("/conversations", conversation.ListHandler)
				group.POST("/conversations", conversation.CreateHandler)
				group.GET("/conversations/{id}", conversation.GetHandler)
				group.PUT("/conversations/{id}", conversation.UpdateHandler)
				group.DELETE("/conversations/{id}", conversation.DeleteHandler)
				group.GET("/conversations/{id}/messages", conversation.ListMessagesHandler)
				group.POST("/files", file.UploadHandler)
				group.GET("/files/{id}", file.GetHandler)
				group.GET("/files/{id}/preview", file.PreviewHandler)
				group.GET("/files/{id}/download", file.DownloadHandler)
				// Chat
				group.GET("/chat", chat.Handler)
				group.POST("/chat/completions", chat.CompletionsHandler)
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

func adminSkillContextMiddleware(r *ghttp.Request) {
	requestedEnterpriseId := strings.TrimSpace(r.Header.Get("X-Enterprise-ID"))
	if requestedEnterpriseId == "" {
		requestedEnterpriseId = strings.TrimSpace(r.Get("enterpriseId").String())
	}
	if identity.IsAdmin(r) && requestedEnterpriseId == "" {
		r.Middleware.Next()
		return
	}
	enterprise.MemberContextMiddleware(r)
}

func adminSkillAccessMiddleware(r *ghttp.Request) {
	if identity.IsAdmin(r) {
		r.Middleware.Next()
		return
	}
	enterprise.AdminMiddleware(r)
}
