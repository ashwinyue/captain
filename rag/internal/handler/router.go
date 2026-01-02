package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/tgo/captain/rag/internal/config"
	"github.com/tgo/captain/rag/internal/repository"
	"github.com/tgo/captain/rag/internal/service"
)

func SetupRouter(cfg *config.Config, db *gorm.DB) *gin.Engine {
	if cfg.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Health check endpoints
	r.GET("/health", healthCheck)
	r.GET("/ready", readinessCheck)
	r.GET("/live", livenessCheck)

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":      "TGO RAG Service",
			"version":      "1.0.0",
			"status":       "running",
			"docs_url":     "/docs",
			"health_check": "/health",
		})
	})

	// Initialize repositories
	collectionRepo := repository.NewCollectionRepository(db)
	fileRepo := repository.NewFileRepository(db)
	documentRepo := repository.NewDocumentRepository(db)
	websitePageRepo := repository.NewWebsitePageRepository(db)
	qaPairRepo := repository.NewQAPairRepository(db)
	embeddingConfigRepo := repository.NewEmbeddingConfigRepository(db)

	// Initialize embedding service
	embeddingSvc := service.NewEmbeddingService(
		cfg.EmbeddingAPIKey,
		cfg.EmbeddingBaseURL,
		cfg.EmbeddingModel,
		cfg.EmbeddingDimensions,
	)

	// Initialize vector search service
	vectorSearchSvc := service.NewVectorSearchService(db, embeddingSvc)

	// Initialize services
	collectionSvc := service.NewCollectionService(collectionRepo, documentRepo)
	fileSvc := service.NewFileService(fileRepo, documentRepo, cfg)
	websiteSvc := service.NewWebsiteService(websitePageRepo, documentRepo)
	qaSvc := service.NewQAServiceWithEmbedding(qaPairRepo, documentRepo, embeddingSvc)
	embeddingConfigSvc := service.NewEmbeddingConfigService(embeddingConfigRepo)

	// Initialize handlers
	collectionHandler := NewCollectionHandler(collectionSvc)
	fileHandler := NewFileHandler(fileSvc)
	websiteHandler := NewWebsiteHandler(websiteSvc)
	qaHandler := NewQAHandler(qaSvc)
	embeddingConfigHandler := NewEmbeddingConfigHandler(embeddingConfigSvc)
	retrieveHandler := NewRetrieveHandler(vectorSearchSvc)

	// API v1
	v1 := r.Group("/v1")
	{
		// Collections
		collections := v1.Group("/collections")
		{
			collections.GET("", collectionHandler.List)
			collections.POST("", collectionHandler.Create)
			collections.POST("/batch", collectionHandler.BatchCreate)
			collections.GET("/:id", collectionHandler.Get)
			collections.PUT("/:id", collectionHandler.Update)
			collections.DELETE("/:id", collectionHandler.Delete)
			collections.GET("/:id/documents", collectionHandler.ListDocuments)
			collections.POST("/:id/documents/search", collectionHandler.SearchDocuments)
		}

		// Files
		files := v1.Group("/files")
		{
			files.GET("", fileHandler.List)
			files.POST("", fileHandler.Upload)
			files.POST("/batch", fileHandler.BatchUpload)
			files.GET("/:id", fileHandler.Get)
			files.DELETE("/:id", fileHandler.Delete)
			files.GET("/:id/documents", fileHandler.ListDocuments)
			files.GET("/:id/download", fileHandler.Download)
		}

		// Websites
		websites := v1.Group("/websites")
		{
			websites.GET("/pages", websiteHandler.ListPages)
			websites.POST("/pages", websiteHandler.AddPage)
			websites.GET("/pages/:id", websiteHandler.GetPage)
			websites.DELETE("/pages/:id", websiteHandler.DeletePage)
			websites.POST("/pages/:id/recrawl", websiteHandler.RecrawlPage)
			websites.POST("/pages/:id/crawl-deeper", websiteHandler.CrawlDeeper)
			websites.GET("/progress", websiteHandler.GetProgress)
		}

		// QA Pairs - collection scoped
		v1.GET("/collections/:id/qa-pairs", qaHandler.List)
		v1.POST("/collections/:id/qa-pairs", qaHandler.Create)
		v1.POST("/collections/:id/qa-pairs/batch", qaHandler.BatchCreate)
		v1.POST("/collections/:id/qa-pairs/import", qaHandler.Import)
		v1.GET("/collections/:id/qa-pairs/stats", qaHandler.Stats)

		// QA Pairs - standalone
		qaPairs := v1.Group("/qa-pairs")
		{
			qaPairs.GET("/:id", qaHandler.Get)
			qaPairs.PUT("/:id", qaHandler.Update)
			qaPairs.DELETE("/:id", qaHandler.Delete)
		}

		// QA Categories
		v1.GET("/qa-categories", qaHandler.ListCategories)

		// Embedding Configs
		embeddingConfigs := v1.Group("/embedding-configs")
		{
			embeddingConfigs.GET("", embeddingConfigHandler.List)
			embeddingConfigs.POST("/batch-sync", embeddingConfigHandler.BatchSync)
			embeddingConfigs.GET("/:project_id", embeddingConfigHandler.GetByProjectID)
		}
	}

	// RAG Retrieve endpoint (for AI agent tool calls)
	r.POST("/retrieve", retrieveHandler.Retrieve)

	return r
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "rag",
	})
}

func readinessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

func livenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}
