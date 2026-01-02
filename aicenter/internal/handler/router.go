package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/tgo/captain/aicenter/internal/config"
	"github.com/tgo/captain/aicenter/internal/eino/memory"
	"github.com/tgo/captain/aicenter/internal/middleware"
	"github.com/tgo/captain/aicenter/internal/pkg/apiserver"
	"github.com/tgo/captain/aicenter/internal/repository"
	"github.com/tgo/captain/aicenter/internal/service"
	"github.com/tgo/captain/aicenter/pkg/auth"
)

type Handlers struct {
	Agent           *AgentHandler
	Team            *TeamHandler
	Chat            *ChatHandler
	Provider        *ProviderHandler
	Tool            *ToolHandler
	ProjectAIConfig *ProjectAIConfigHandler
}

func SetupRouter(cfg *config.Config, db *gorm.DB) *gin.Engine {
	if cfg.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Global middleware
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())

	// Health check
	r.GET("/health", healthCheck)

	// Initialize handlers
	handlers := initHandlers(db, cfg)

	// Auth middleware
	var authMw *middleware.AuthMiddleware
	if cfg.AuthServiceURL != "" {
		authClient := auth.NewClient(cfg.AuthServiceURL)
		authMw = middleware.NewAuthMiddleware(authClient, cfg.IsDevelopment())
	}

	// API v1 with project ID requirement
	v1 := r.Group("/api/v1")
	if authMw != nil {
		v1.Use(authMw.APIKeyAuth())
	} else {
		v1.Use(middleware.ProjectID())
	}
	{
		// Agents
		agents := v1.Group("/agents")
		{
			agents.GET("", handlers.Agent.List)
			agents.POST("", handlers.Agent.Create)
			agents.GET("/exists", handlers.Agent.Exists)
			agents.GET("/:id", handlers.Agent.Get)
			agents.PATCH("/:id", handlers.Agent.Update)
			agents.DELETE("/:id", handlers.Agent.Delete)
			agents.PATCH("/:id/tools/:tool_id/enabled", handlers.Agent.SetToolEnabled)
			agents.PATCH("/:id/collections/:collection_id/enabled", handlers.Agent.SetCollectionEnabled)
		}

		// Agent Run (SSE)
		v1.POST("/agents/run", handlers.Chat.Run)
		v1.POST("/agents/run/:run_id/cancel", handlers.Chat.Cancel)

		// Teams
		teams := v1.Group("/teams")
		{
			teams.GET("", handlers.Team.List)
			teams.POST("", handlers.Team.Create)
			teams.GET("/default", handlers.Team.GetDefault)
			teams.GET("/:id", handlers.Team.Get)
			teams.PATCH("/:id", handlers.Team.Update)
			teams.DELETE("/:id", handlers.Team.Delete)
		}

		// LLM Providers
		providers := v1.Group("/llm-providers")
		{
			providers.POST("/sync", handlers.Provider.Sync)
			providers.GET("", handlers.Provider.List)
			providers.POST("", handlers.Provider.Create)
			providers.GET("/:id", handlers.Provider.Get)
			providers.PATCH("/:id", handlers.Provider.Update)
			providers.DELETE("/:id", handlers.Provider.Delete)
			providers.POST("/:id/test", handlers.Provider.Test)
		}

		// Chat Completions (OpenAI compatible)
		v1.POST("/chat/completions", handlers.Chat.Completions)

		// Tools
		tools := v1.Group("/tools")
		{
			tools.GET("", handlers.Tool.List)
			tools.POST("", handlers.Tool.Create)
			tools.GET("/:id", handlers.Tool.Get)
			tools.PATCH("/:id", handlers.Tool.Update)
			tools.DELETE("/:id", handlers.Tool.Delete)
		}

		// Project AI Configs (internal sync from tgo-api)
		projectConfigs := v1.Group("/project-ai-configs")
		{
			projectConfigs.POST("/sync", handlers.ProjectAIConfig.Sync)
			projectConfigs.PUT("", handlers.ProjectAIConfig.Upsert)
			projectConfigs.GET("", handlers.ProjectAIConfig.Get)
		}
	}

	return r
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "aicenter",
	})
}

func initHandlers(db *gorm.DB, cfg *config.Config) *Handlers {
	// Initialize repositories
	agentRepo := repository.NewAgentRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	providerRepo := repository.NewProviderRepository(db)
	toolRepo := repository.NewToolRepository(db)
	projectConfigRepo := repository.NewProjectAIConfigRepository(db)

	// Initialize services
	agentSvc := service.NewAgentService(agentRepo)
	teamSvc := service.NewTeamService(teamRepo)
	providerSvc := service.NewProviderService(providerRepo)
	runtimeSvc := service.NewRuntimeService(db, teamRepo, projectConfigRepo, providerRepo, cfg.RAGServiceURL, cfg.MCPServiceURL)
	toolSvc := service.NewToolService(toolRepo)
	projectConfigSvc := service.NewProjectAIConfigService(projectConfigRepo)

	// Set up apiserver client for internal API calls
	if cfg.InternalAPIURL != "" {
		apiserverClient := apiserver.NewClient(cfg.InternalAPIURL)
		runtimeSvc.SetApiserverClient(apiserverClient)
		log.Printf("Apiserver internal client enabled -> %s", cfg.InternalAPIURL)
	}

	// Set up Redis for memory caching
	if cfg.RedisURL != "" {
		redisStore, err := memory.NewRedisStoreFromURL(cfg.RedisURL, 30*time.Minute)
		if err != nil {
			log.Printf("Warning: Failed to connect to Redis for memory, using PostgreSQL only: %v", err)
		} else {
			runtimeSvc.SetRedisStore(redisStore)
			log.Printf("Redis memory cache enabled -> %s", cfg.RedisURL)
		}
	}

	return &Handlers{
		Agent:           NewAgentHandler(agentSvc),
		Team:            NewTeamHandler(teamSvc),
		Chat:            NewChatHandler(runtimeSvc, cfg),
		Provider:        NewProviderHandler(providerSvc),
		Tool:            NewToolHandler(toolSvc),
		ProjectAIConfig: NewProjectAIConfigHandler(projectConfigSvc),
	}
}
