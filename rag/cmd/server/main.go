package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/tgo/captain/rag/internal/config"
	"github.com/tgo/captain/rag/internal/database"
	"github.com/tgo/captain/rag/internal/handler"
)

func main() {
	// Load .env file if exists
	godotenv.Load()

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Setup router
	r := handler.SetupRouter(cfg, db)

	// Start server
	addr := cfg.Host + ":" + cfg.Port
	log.Printf("RAG Service starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
		os.Exit(1)
	}
}
