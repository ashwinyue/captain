package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/tgo/captain/apiserver/internal/config"
	"github.com/tgo/captain/apiserver/internal/middleware"
	"github.com/tgo/captain/apiserver/internal/pkg/aicenter"
	"github.com/tgo/captain/apiserver/internal/pkg/jwt"
	"github.com/tgo/captain/apiserver/internal/pkg/rag"
	"github.com/tgo/captain/apiserver/internal/pkg/wukongim"
	"github.com/tgo/captain/apiserver/internal/repository"
	"github.com/tgo/captain/apiserver/internal/service"
)

func SetupRouter(cfg *config.Config, db *gorm.DB, humanSessionSvc *service.HumanSessionService) *gin.Engine {
	if cfg.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())

	r.GET("/health", healthCheck)

	// Initialize JWT manager
	jwtManager := jwt.NewManager(cfg.JWTSecret, cfg.AccessTokenExpireMin, cfg.RefreshTokenExpireDays)

	// Initialize WuKongIM client
	wkClient := wukongim.NewClient(cfg.WuKongIMURL, cfg.WuKongIMAPIKey)

	// Initialize repositories
	staffRepo := repository.NewStaffRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	visitorRepo := repository.NewVisitorRepository(db)
	tagRepo := repository.NewTagRepository(db)
	visitorTagRepo := repository.NewVisitorTagRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	conversationRepo := repository.NewConversationRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	channelRepo := repository.NewChannelRepository(db)
	channelMemberRepo := repository.NewChannelMemberRepository(db)
	queueRepo := repository.NewQueueRepository(db)
	assignmentRuleRepo := repository.NewAssignmentRuleRepository(db)
	platformRepo := repository.NewPlatformRepository(db)
	setupRepo := repository.NewSetupRepository(db)
	onboardingRepo := repository.NewOnboardingRepository(db)

	// Initialize AI center client (needed by onboarding service)
	var aiClient *aicenter.Client
	if cfg.AICenterURL != "" {
		aiClient = aicenter.NewClient(cfg.AICenterURL)
	}

	// Initialize RAG client (needed by onboarding service)
	var ragClient *rag.Client
	if cfg.RAGServiceURL != "" {
		ragClient = rag.NewClient(cfg.RAGServiceURL)
	}

	// Initialize services
	authSvc := service.NewAuthService(db, jwtManager, wkClient)
	staffSvc := service.NewStaffService(staffRepo)
	projectSvc := service.NewProjectService(projectRepo)
	visitorSvc := service.NewVisitorService(visitorRepo, wkClient, db)
	tagSvc := service.NewTagService(tagRepo, visitorTagRepo)
	queueSvc := service.NewQueueService(queueRepo)
	chatSvc := service.NewChatServiceWithDB(messageRepo, conversationRepo, queueRepo, wkClient, db, cfg.AICenterURL)
	if humanSessionSvc != nil {
		chatSvc.SetHumanSessionService(humanSessionSvc)
	}
	sessionSvc := service.NewSessionService(sessionRepo)
	channelSvc := service.NewChannelService(channelRepo, channelMemberRepo, wkClient)
	assignmentRuleSvc := service.NewAssignmentRuleService(assignmentRuleRepo)
	searchSvc := service.NewSearchService(visitorRepo)
	emailSvc := service.NewEmailService()
	wukongimSvc := service.NewWuKongIMService(wkClient, cfg.WuKongIMURL, db)
	platformSvc := service.NewPlatformService(platformRepo)
	setupSvc := service.NewSetupService(db, setupRepo, platformRepo)
	onboardingSvc := service.NewOnboardingService(onboardingRepo, aiClient, ragClient)
	transferSvc := service.NewTransferService(staffRepo, visitorRepo, queueRepo, wkClient)

	// Initialize handlers
	authHandler := NewAuthHandler(authSvc)
	staffHandler := NewStaffHandler(staffSvc, cfg.WuKongIMWSURL)
	projectHandler := NewProjectHandler(projectSvc)
	visitorHandler := NewVisitorHandler(visitorSvc, tagSvc)
	tagHandler := NewTagHandler(tagSvc)
	queueHandler := NewQueueHandler(queueSvc)
	chatHandler := NewChatHandler(chatSvc)
	sessionHandler := NewSessionHandler(sessionSvc)
	channelHandler := NewChannelHandler(channelSvc)
	assignmentRuleHandler := NewAssignmentRuleHandler(assignmentRuleSvc)
	searchHandler := NewSearchHandler(searchSvc)
	emailHandler := NewEmailHandler(emailSvc)
	wukongimHandler := NewWuKongIMHandler(wukongimSvc)
	webhookHandler := NewWebhookHandler()
	systemHandler := NewSystemHandler("1.0.0")
	utilsHandler := NewUtilsHandler()
	docsHandler := NewDocsHandler(cfg.AICenterURL, cfg.RAGServiceURL, "")
	platformHandler := NewPlatformHandler(platformSvc)
	setupHandler := NewSetupHandler(setupSvc)
	onboardingHandler := NewOnboardingHandler(onboardingSvc)
	transferHandler := NewTransferHandler(transferSvc)
	aiEventsHandler := NewAIEventsHandler(db, wkClient)

	// Initialize AI center handlers (aiClient created earlier)
	var aiHandler *AIHandler
	var mcpToolsHandler *MCPToolsHandler
	if aiClient != nil {
		aiHandler = NewAIHandler(aiClient, wkClient)
		mcpToolsHandler = NewMCPToolsHandler(aiClient)
		log.Printf("AI Center client enabled -> %s", cfg.AICenterURL)
	}

	// Initialize RAG handler (ragClient created earlier)
	var ragHandler *RAGHandler
	if ragClient != nil {
		ragHandler = NewRAGHandler(ragClient)
		log.Printf("RAG Service client enabled -> %s", cfg.RAGServiceURL)
	}

	// Auth middleware
	authMw := middleware.NewAuthMiddleware(jwtManager, db, cfg.IsDevelopment())

	// API v1
	v1 := r.Group("/v1")
	{
		// Setup routes (no auth required)
		setup := v1.Group("/setup")
		{
			setup.GET("/status", setupHandler.GetStatus)
			setup.POST("/admin", setupHandler.CreateAdmin)
			setup.POST("/staff", setupHandler.BatchCreateStaff)
			setup.POST("/llm-config", setupHandler.ConfigureLLM)
			setup.POST("/skip-llm", setupHandler.SkipLLMConfig)
			setup.GET("/verify", setupHandler.Verify)
		}

		// Docs routes (no auth required)
		docs := v1.Group("/docs")
		{
			docs.GET("", docsHandler.Index)
			docs.GET("/:service", docsHandler.ServiceDocs)
			docs.GET("/:service/openapi.json", docsHandler.ProxyOpenAPI)
		}

		// Auth routes (no auth required)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Public visitor routes (no auth required, uses platform API key)
		v1.POST("/visitors/register", visitorHandler.Register)
		v1.POST("/visitors/messages/sync", visitorHandler.SyncMessages)

		// Public platform routes (no auth required, uses platform API key)
		v1.GET("/platforms/info", platformHandler.GetInfo)

		// Public chat routes (no auth required, uses platform API key)
		v1.POST("/chat/completion", chatHandler.ChatCompletion)
		v1.GET("/channels/info", wukongimHandler.GetChannelInfo)

		// Public transfer routes (no auth required, uses platform API key)
		v1.POST("/transfer/to-staff", transferHandler.TransferToStaffByPlatformKey)

		// Internal AI events endpoint (no auth required, for AI center callbacks)
		internal := v1.Group("/internal")
		{
			internal.POST("/ai-events", aiEventsHandler.IngestEvent)
		}

		// Staff login (alias for /auth/login for frontend compatibility)
		v1.POST("/staff/login", authHandler.Login)

		// Protected routes
		protected := v1.Group("")
		protected.Use(authMw.APIKeyAuth())
		{
			// Current user
			protected.GET("/auth/me", authHandler.GetCurrentUser)

			// Staff
			staff := protected.Group("/staff")
			{
				staff.GET("", staffHandler.List)
				staff.POST("", staffHandler.Create)
				staff.GET("/me", staffHandler.GetMe)
				staff.PUT("/me/service-paused", staffHandler.UpdateMyServicePaused)
				staff.PUT("/me/is-active", staffHandler.UpdateMyIsActive)
				staff.GET("/wukongim/status", staffHandler.GetWuKongIMStatus)
				staff.POST("/wukongim/online-status", staffHandler.CheckWuKongIMOnlineStatus)
				staff.GET("/:id", staffHandler.Get)
				staff.PATCH("/:id", staffHandler.Update)
				staff.DELETE("/:id", staffHandler.Delete)
				staff.PUT("/:id/service-paused", staffHandler.UpdateServicePaused)
				staff.PUT("/:id/is-active", staffHandler.UpdateIsActive)
			}

			// Projects
			projects := protected.Group("/projects")
			{
				projects.GET("", projectHandler.List)
				projects.POST("", projectHandler.Create)
				projects.GET("/:id", projectHandler.Get)
				projects.PATCH("/:id", projectHandler.Update)
				projects.DELETE("/:id", projectHandler.Delete)
				projects.POST("/:id/regenerate-api-key", projectHandler.RegenerateAPIKey)
				// Project AI Config (proxied to aicenter)
				if aiHandler != nil {
					projects.GET("/:id/ai-config", aiHandler.GetProjectAIConfigByID)
					projects.PUT("/:id/ai-config", aiHandler.UpsertProjectAIConfigByID)
				}
			}
		}

		// Project-scoped routes
		projectScoped := v1.Group("")
		projectScoped.Use(authMw.APIKeyAuth(), authMw.ProjectRequired())
		{
			// Visitors
			visitors := projectScoped.Group("/visitors")
			{
				visitors.GET("", visitorHandler.List)
				visitors.POST("", visitorHandler.Create)
				visitors.GET("/by-channel", visitorHandler.GetByChannel)
				visitors.GET("/:id", visitorHandler.Get)
				visitors.GET("/:id/basic", visitorHandler.GetBasic)
				visitors.PATCH("/:id", visitorHandler.Update)
				visitors.PUT("/:id/attributes", visitorHandler.SetAttributes)
				visitors.DELETE("/:id", visitorHandler.Delete)
				visitors.POST("/:id/block", visitorHandler.Block)
				visitors.POST("/:id/unblock", visitorHandler.Unblock)
				visitors.POST("/:id/accept", visitorHandler.Accept)
				visitors.POST("/:id/enable-ai", visitorHandler.EnableAI)
				visitors.POST("/:id/disable-ai", visitorHandler.DisableAI)
				visitors.GET("/:id/tags", visitorHandler.GetTags)
				visitors.POST("/:id/tags/:tag_id", visitorHandler.AddTag)
				visitors.DELETE("/:id/tags/:tag_id", visitorHandler.RemoveTag)
			}

			// Tags
			tags := projectScoped.Group("/tags")
			{
				tags.GET("", tagHandler.List)
				tags.POST("", tagHandler.Create)
				tags.GET("/:id", tagHandler.Get)
				tags.PATCH("/:id", tagHandler.Update)
				tags.DELETE("/:id", tagHandler.Delete)
				tags.POST("/visitor-tags", tagHandler.AddVisitorTag)
				tags.DELETE("/visitor-tags", tagHandler.RemoveVisitorTag)
			}

			// Visitor Waiting Queue
			queue := projectScoped.Group("/visitor-waiting-queue")
			{
				queue.GET("", queueHandler.List)
				queue.POST("", queueHandler.Add)
				queue.GET("/count", queueHandler.GetCount)
				queue.POST("/:id/assign", queueHandler.Assign)
				queue.DELETE("/:id", queueHandler.Remove)
				queue.GET("/:id/position", queueHandler.GetPosition)
			}

			// Chat
			chat := projectScoped.Group("/chat")
			{
				chat.POST("/send", chatHandler.SendMessage)
				chat.GET("/messages", chatHandler.GetMessages)
				chat.POST("/messages/:id/revoke", chatHandler.RevokeMessage)
				if aiHandler != nil {
					chat.POST("/team", aiHandler.TeamChat)
					chat.POST("/team/stream", aiHandler.TeamChatStream) // 流式代理
				}
			}

			// Conversations
			projectScoped.GET("/conversations", chatHandler.GetConversations)
			projectScoped.POST("/conversations/my", chatHandler.GetMyConversations)
			projectScoped.POST("/conversations/waiting", chatHandler.GetWaitingConversations)
			projectScoped.POST("/conversations/messages", wukongimHandler.SyncChannelMessages)
			projectScoped.POST("/conversations/all", chatHandler.GetAllConversations)
			projectScoped.GET("/conversations/by-tags/recent", chatHandler.GetConversationsByTagsRecent)
			projectScoped.PUT("/conversations/unread", chatHandler.SetConversationUnread)

			// Sessions
			sessions := projectScoped.Group("/sessions")
			{
				sessions.GET("", sessionHandler.List)
				sessions.POST("", sessionHandler.Create)
				sessions.GET("/:id", sessionHandler.Get)
				sessions.PATCH("/:id", sessionHandler.Update)
				sessions.POST("/:id/close", sessionHandler.Close)
				sessions.POST("/:id/transfer", sessionHandler.Transfer)
			}

			// Channels
			channels := projectScoped.Group("/channels")
			{
				channels.GET("", channelHandler.List)
				channels.POST("", channelHandler.Create)
				channels.GET("/:id", channelHandler.Get)
				channels.DELETE("/:id", channelHandler.Delete)
				channels.POST("/:id/members", channelHandler.AddMembers)
				channels.DELETE("/:id/members", channelHandler.RemoveMembers)
			}

			// Platforms
			platforms := projectScoped.Group("/platforms")
			{
				platforms.GET("/types", platformHandler.ListTypes)
				platforms.GET("", platformHandler.List)
				platforms.POST("", platformHandler.Create)
				platforms.GET("/:id", platformHandler.Get)
				platforms.PATCH("/:id", platformHandler.Update)
				platforms.DELETE("/:id", platformHandler.Delete)
				platforms.POST("/:id/regenerate-api-key", platformHandler.RegenerateAPIKey)
			}

			// Visitor Assignment Rules
			assignmentRules := projectScoped.Group("/visitor-assignment-rules")
			{
				assignmentRules.GET("", assignmentRuleHandler.Get)
				assignmentRules.PUT("", assignmentRuleHandler.Update)
				assignmentRules.GET("/default-prompt", assignmentRuleHandler.GetDefaultPrompt)
			}

			// Search
			projectScoped.GET("/search", searchHandler.Search)

			// Utils
			utils := projectScoped.Group("/utils")
			{
				utils.POST("/extract-website-metadata", utilsHandler.ExtractWebsiteMetadata)
			}

			// Onboarding
			onboarding := projectScoped.Group("/onboarding")
			{
				onboarding.GET("", onboardingHandler.GetProgress)
				onboarding.POST("/skip", onboardingHandler.Skip)
				onboarding.POST("/reset", onboardingHandler.Reset)
			}

			// MCP Project Tools (via aicenter client)
			if mcpToolsHandler != nil {
				mcpTools := projectScoped.Group("/project-tools")
				{
					mcpTools.GET("", mcpToolsHandler.List)
					mcpTools.GET("/stats", mcpToolsHandler.GetStats)
					mcpTools.GET("/:id", mcpToolsHandler.Get)
					mcpTools.PUT("/:id", mcpToolsHandler.Update)
					mcpTools.DELETE("/:id", mcpToolsHandler.Uninstall)
					mcpTools.POST("/install", mcpToolsHandler.Install)
					mcpTools.POST("/bulk-install", mcpToolsHandler.BulkInstall)
				}
			}

			// AI routes (via aicenter client)
			if aiHandler != nil {
				// Agents
				agents := projectScoped.Group("/agents")
				{
					agents.GET("", aiHandler.ListAgents)
					agents.POST("", aiHandler.CreateAgent)
					agents.GET("/:id", aiHandler.GetAgent)
					agents.PATCH("/:id", aiHandler.UpdateAgent)
					agents.DELETE("/:id", aiHandler.DeleteAgent)
					agents.POST("/:id/run", aiHandler.RunAgent)
				}

				// AI Agents (alias for /agents to match Python API path /ai/agents)
				aiAgents := projectScoped.Group("/ai/agents")
				{
					aiAgents.GET("", aiHandler.ListAgents)
					aiAgents.POST("", aiHandler.CreateAgent)
					aiAgents.GET("/:id", aiHandler.GetAgent)
					aiAgents.PUT("/:id", aiHandler.UpdateAgent)
					aiAgents.PATCH("/:id", aiHandler.UpdateAgent)
					aiAgents.DELETE("/:id", aiHandler.DeleteAgent)
					aiAgents.POST("/:id/run", aiHandler.RunAgent)
				}

				// Teams
				teams := projectScoped.Group("/teams")
				{
					teams.GET("", aiHandler.ListTeams)
					teams.POST("", aiHandler.CreateTeam)
					teams.GET("/default", aiHandler.GetDefaultTeam)
					teams.GET("/:id", aiHandler.GetTeam)
					teams.PATCH("/:id", aiHandler.UpdateTeam)
					teams.DELETE("/:id", aiHandler.DeleteTeam)
					teams.POST("/:id/run", aiHandler.RunTeam)
				}

				// AI Teams (alias for /teams to match Python API path /ai/teams)
				aiTeams := projectScoped.Group("/ai/teams")
				{
					aiTeams.GET("", aiHandler.ListTeams)
					aiTeams.POST("", aiHandler.CreateTeam)
					aiTeams.GET("/default", aiHandler.GetDefaultTeam)
					aiTeams.GET("/:id", aiHandler.GetTeam)
					aiTeams.PATCH("/:id", aiHandler.UpdateTeam)
					aiTeams.DELETE("/:id", aiHandler.DeleteTeam)
					aiTeams.POST("/:id/run", aiHandler.RunTeam)
				}

				// Tools
				tools := projectScoped.Group("/tools")
				{
					tools.GET("", aiHandler.ListTools)
					tools.POST("", aiHandler.CreateTool)
					tools.GET("/:id", aiHandler.GetTool)
					tools.PATCH("/:id", aiHandler.UpdateTool)
					tools.DELETE("/:id", aiHandler.DeleteTool)
				}

				// AI Tools (alias for /tools to match Python API path /ai/tools)
				aiTools := projectScoped.Group("/ai/tools")
				{
					aiTools.GET("", aiHandler.ListTools)
					aiTools.POST("", aiHandler.CreateTool)
					aiTools.GET("/:id", aiHandler.GetTool)
					aiTools.PATCH("/:id", aiHandler.UpdateTool)
					aiTools.DELETE("/:id", aiHandler.DeleteTool)
				}

				// AI Providers (mounted at /ai/providers to match Python API)
				providers := projectScoped.Group("/ai/providers")
				{
					providers.GET("", aiHandler.ListProviders)
					providers.POST("", aiHandler.CreateProvider)
					providers.GET("/:id", aiHandler.GetProvider)
					providers.PATCH("/:id", aiHandler.UpdateProvider)
					providers.DELETE("/:id", aiHandler.DeleteProvider)
					providers.POST("/:id/enable", aiHandler.EnableProvider)
					providers.POST("/:id/disable", aiHandler.DisableProvider)
					providers.POST("/:id/sync", aiHandler.SyncProvider)
					providers.POST("/:id/test", aiHandler.TestProvider)
				}

				// Models
				models := projectScoped.Group("/models")
				{
					models.GET("", aiHandler.ListModels)
					models.GET("/:id", aiHandler.GetModel)
				}

				// AI Models (fetch from provider API)
				v1.POST("/ai/models", aiHandler.FetchModels)

				// Project AI Configs
				projectScoped.GET("/project-ai-configs", aiHandler.GetProjectAIConfig)
				projectScoped.PUT("/project-ai-configs", aiHandler.UpsertProjectAIConfig)
			}

			// RAG routes (via rag client)
			if ragHandler != nil {
				// Collections
				collections := projectScoped.Group("/rag/collections")
				{
					collections.GET("", ragHandler.ListCollections)
					collections.POST("", ragHandler.CreateCollection)
					collections.GET("/:id", ragHandler.GetCollection)
					collections.PUT("/:id", ragHandler.UpdateCollection)
					collections.PATCH("/:id", ragHandler.UpdateCollection)
					collections.DELETE("/:id", ragHandler.DeleteCollection)
					collections.POST("/:id/documents/search", ragHandler.SearchCollectionDocuments)
					collections.GET("/:id/pages", ragHandler.ListCollectionPages)
					collections.GET("/:id/qa-pairs", ragHandler.ListQAPairs)
					collections.POST("/:id/qa-pairs", ragHandler.CreateQAPair)
					collections.POST("/:id/qa-pairs/batch", ragHandler.BatchCreateQAPairs)
					collections.POST("/:id/qa-pairs/import", ragHandler.ImportQAPairs)
					collections.GET("/:id/qa-pairs/stats", ragHandler.GetQAStats)
				}

				// Alternative RAG routes (without /collections prefix, for frontend compatibility)
				rag := projectScoped.Group("/rag")
				{
					rag.GET("/:id/qa-pairs", ragHandler.ListQAPairs)
					rag.POST("/:id/qa-pairs", ragHandler.CreateQAPair)
					rag.POST("/:id/qa-pairs/batch", ragHandler.BatchCreateQAPairs)
					rag.POST("/:id/qa-pairs/import", ragHandler.ImportQAPairs)
					rag.GET("/:id/qa-pairs/stats", ragHandler.GetQAStats)
					rag.PATCH("/:id", ragHandler.UpdateCollection)
				}

				// Files
				files := projectScoped.Group("/rag/files")
				{
					files.GET("", ragHandler.ListFiles)
					files.POST("", ragHandler.UploadFile)
					files.GET("/:id", ragHandler.GetFile)
					files.DELETE("/:id", ragHandler.DeleteFile)
					files.GET("/:id/download", ragHandler.DownloadFile)
				}

				// Website Pages
				websitePages := projectScoped.Group("/rag/websites/pages")
				{
					websitePages.GET("", ragHandler.ListWebsitePages)
					websitePages.POST("", ragHandler.AddWebsitePage)
					websitePages.GET("/:id", ragHandler.GetWebsitePage)
					websitePages.DELETE("/:id", ragHandler.DeleteWebsitePage)
					websitePages.POST("/:id/recrawl", ragHandler.RecrawlWebsitePage)
					websitePages.POST("/:id/crawl-deeper", ragHandler.CrawlDeeperFromPage)
				}

				// Website Progress
				projectScoped.GET("/rag/websites/progress", ragHandler.GetCrawlProgress)

				// QA Pairs (standalone endpoints)
				qaPairs := projectScoped.Group("/rag/qa-pairs")
				{
					qaPairs.GET("/:id", ragHandler.GetQAPair)
					qaPairs.PUT("/:id", ragHandler.UpdateQAPair)
					qaPairs.DELETE("/:id", ragHandler.DeleteQAPair)
				}

				// QA Categories
				projectScoped.GET("/rag/qa-categories", ragHandler.ListQACategories)
			}
		}

		// Email (protected, not project-scoped)
		email := v1.Group("/email")
		email.Use(authMw.APIKeyAuth())
		{
			email.POST("/test-connection", emailHandler.TestConnection)
		}

		// WuKongIM public routes (no auth required)
		wukongim := v1.Group("/wukongim")
		{
			wukongim.GET("/route", wukongimHandler.GetRoute)
		}

		// WuKongIM webhook (no auth required)
		v1.POST("/integrations/wukongim/webhook", webhookHandler.WuKongIMWebhook)

		// System info
		system := v1.Group("/system")
		{
			system.GET("/info", systemHandler.GetInfo)
			system.GET("/health", systemHandler.GetHealth)
		}

	}

	return r
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "apiserver",
	})
}
