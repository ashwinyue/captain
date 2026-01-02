package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/platform/internal/config"
	"github.com/tgo/captain/platform/internal/model"
	"github.com/tgo/captain/platform/internal/service"
)

func SetupRouter(cfg *config.Config, db *gorm.DB) *gin.Engine {
	if cfg.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(corsMiddleware())

	r.GET("/health", healthCheck)

	platformSvc := service.NewPlatformService(db)
	onboardingSvc := service.NewOnboardingService(db)
	callbackHandler := NewCallbackHandler(db)
	messageHandler := NewMessageHandler(db)

	// Callback routes (no auth required - third-party platforms)
	callbacks := r.Group("/v1/platforms/callback")
	{
		callbacks.GET("/:platform_api_key", callbackHandler.WeComVerify)
		callbacks.POST("/:platform_api_key", callbackHandler.WeComCallback)
		callbacks.POST("/:platform_api_key/feishu", callbackHandler.FeishuCallback)
		callbacks.POST("/:platform_api_key/dingtalk", callbackHandler.DingTalkCallback)
		callbacks.POST("/:platform_api_key/wukongim", callbackHandler.WuKongIMCallback)
	}

	// Message routes (no auth - uses platform_api_key in body)
	r.POST("/ingest", messageHandler.Ingest)
	r.POST("/v1/messages/send", messageHandler.SendMessage)

	v1 := r.Group("/api/v1")
	v1.Use(projectAuthMiddleware())
	{
		// Platforms
		platforms := v1.Group("/platforms")
		{
			platforms.GET("", listPlatforms(platformSvc))
			platforms.POST("", createPlatform(platformSvc))
			platforms.GET("/:id", getPlatform(platformSvc))
			platforms.PATCH("/:id", updatePlatform(platformSvc))
			platforms.DELETE("/:id", deletePlatform(platformSvc))
			platforms.POST("/:id/enable", enablePlatform(platformSvc))
			platforms.POST("/:id/disable", disablePlatform(platformSvc))
		}

		// Onboarding
		onboarding := v1.Group("/onboarding")
		{
			onboarding.GET("", getOnboarding(onboardingSvc))
			onboarding.POST("/step/:step/complete", completeStep(onboardingSvc))
			onboarding.POST("/step/:step/skip", skipStep(onboardingSvc))
			onboarding.POST("/reset", resetOnboarding(onboardingSvc))
		}
	}

	return r
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "platform"})
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key, X-Project-ID")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func projectAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.GetHeader("X-Project-ID")
		if projectID == "" {
			projectID = c.Query("project_id")
		}
		if projectID == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "X-Project-ID required"})
			return
		}
		if _, err := uuid.Parse(projectID); err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": "invalid project_id"})
			return
		}
		c.Set("project_id", projectID)
		c.Next()
	}
}

func listPlatforms(svc *service.PlatformService) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID, _ := uuid.Parse(c.GetString("project_id"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		platforms, total, err := svc.List(c.Request.Context(), projectID, limit, offset)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"data": platforms, "total": total})
	}
}

func createPlatform(svc *service.PlatformService) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID, _ := uuid.Parse(c.GetString("project_id"))
		var platform model.Platform
		if err := c.ShouldBindJSON(&platform); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		platform.ProjectID = projectID
		if err := svc.Create(c.Request.Context(), &platform); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(201, platform)
	}
}

func getPlatform(svc *service.PlatformService) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID, _ := uuid.Parse(c.GetString("project_id"))
		id, _ := uuid.Parse(c.Param("id"))
		platform, err := svc.GetByID(c.Request.Context(), projectID, id)
		if err != nil {
			c.JSON(404, gin.H{"error": "platform not found"})
			return
		}
		c.JSON(200, platform)
	}
}

func updatePlatform(svc *service.PlatformService) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID, _ := uuid.Parse(c.GetString("project_id"))
		id, _ := uuid.Parse(c.Param("id"))
		platform, err := svc.GetByID(c.Request.Context(), projectID, id)
		if err != nil {
			c.JSON(404, gin.H{"error": "platform not found"})
			return
		}
		if err := c.ShouldBindJSON(platform); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if err := svc.Update(c.Request.Context(), platform); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, platform)
	}
}

func deletePlatform(svc *service.PlatformService) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID, _ := uuid.Parse(c.GetString("project_id"))
		id, _ := uuid.Parse(c.Param("id"))
		if err := svc.Delete(c.Request.Context(), projectID, id); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"success": true})
	}
}

func enablePlatform(svc *service.PlatformService) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID, _ := uuid.Parse(c.GetString("project_id"))
		id, _ := uuid.Parse(c.Param("id"))
		if err := svc.Enable(c.Request.Context(), projectID, id); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"success": true})
	}
}

func disablePlatform(svc *service.PlatformService) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID, _ := uuid.Parse(c.GetString("project_id"))
		id, _ := uuid.Parse(c.Param("id"))
		if err := svc.Disable(c.Request.Context(), projectID, id); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"success": true})
	}
}

func getOnboarding(svc *service.OnboardingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID, _ := uuid.Parse(c.GetString("project_id"))
		onboarding, err := svc.GetByProjectID(c.Request.Context(), projectID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, onboarding)
	}
}

func completeStep(svc *service.OnboardingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID, _ := uuid.Parse(c.GetString("project_id"))
		step, _ := strconv.Atoi(c.Param("step"))
		if err := svc.UpdateStep(c.Request.Context(), projectID, step, true); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"success": true})
	}
}

func skipStep(svc *service.OnboardingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID, _ := uuid.Parse(c.GetString("project_id"))
		step, _ := strconv.Atoi(c.Param("step"))
		if err := svc.SkipStep(c.Request.Context(), projectID, step); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"success": true})
	}
}

func resetOnboarding(svc *service.OnboardingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID, _ := uuid.Parse(c.GetString("project_id"))
		if err := svc.Reset(c.Request.Context(), projectID); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"success": true})
	}
}
