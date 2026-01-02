package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tgo/captain/aicenter/internal/config"
	"github.com/tgo/captain/aicenter/internal/handler"
	"github.com/tgo/captain/aicenter/internal/pkg/db"
	"github.com/tgo/captain/aicenter/internal/service"
	"github.com/tgo/captain/aicenter/internal/task"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Init database
	database, err := db.NewGorm(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// Auto migrate
	if err := db.AutoMigrate(database); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Setup router
	router := handler.SetupRouter(cfg, database)

	// Setup background tasks
	scheduler := task.NewScheduler()
	if cfg.RAGServiceURL != "" {
		embeddingSyncSvc := service.NewEmbeddingSyncService(database, cfg.RAGServiceURL)
		scheduler.RegisterTask(task.NewEmbeddingSyncRetryTask(embeddingSyncSvc))
	}

	// Run startup tasks once
	scheduler.RunOnce(context.Background())

	// Create server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
