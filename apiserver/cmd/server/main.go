package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tgo/captain/apiserver/internal/config"
	"github.com/tgo/captain/apiserver/internal/handler"
	"github.com/tgo/captain/apiserver/internal/pkg/db"
	"github.com/tgo/captain/apiserver/internal/pkg/redis"
	"github.com/tgo/captain/apiserver/internal/pkg/wukongim"
	"github.com/tgo/captain/apiserver/internal/repository"
	"github.com/tgo/captain/apiserver/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	gormDB, err := db.NewGormDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	if err := db.AutoMigrate(gormDB); err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}

	// Start queue cleanup service in background
	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	queueCleanupSvc := service.NewQueueCleanupService(gormDB, 30, 5) // 30 min expiry, 5 min interval
	go queueCleanupSvc.Start(cleanupCtx)

	// Initialize Redis client for human session management
	var humanSessionSvc *service.HumanSessionService
	if cfg.RedisURL != "" {
		redisClient, err := redis.NewClient(cfg.RedisURL)
		if err != nil {
			log.Printf("Warning: Failed to connect to Redis, human session timeout disabled: %v", err)
		} else {
			visitorRepo := repository.NewVisitorRepository(gormDB)
			wukongClient := wukongim.NewClient(cfg.WuKongIMURL, cfg.WuKongIMAPIKey)
			humanSessionSvc = service.NewHumanSessionService(redisClient, visitorRepo, wukongClient, 5*time.Minute)
			humanSessionSvc.Start(cleanupCtx)
			log.Println("Human session timeout service started (5 min timeout)")
		}
	}

	router := handler.SetupRouter(cfg, gormDB, humanSessionSvc)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("API Server starting on port %s", cfg.Port)
		log.Printf("Environment: %s", cfg.Environment)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Stop queue cleanup service
	cleanupCancel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited")
}
