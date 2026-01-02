package config

import (
	"os"
	"strconv"
)

type Config struct {
	// Server
	Host        string
	Port        string
	Environment string
	GinMode     string

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// Vector DB (Milvus/Qdrant)
	VectorDBURL    string
	VectorDBAPIKey string

	// Embedding Service (OpenAI compatible)
	EmbeddingAPIKey     string
	EmbeddingBaseURL    string
	EmbeddingModel      string
	EmbeddingDimensions int

	// File Storage
	StoragePath    string
	MaxUploadSize  int64
	AllowedFormats []string
}

func Load() *Config {
	return &Config{
		Host:        getEnv("HOST", "0.0.0.0"),
		Port:        getEnv("PORT", "8087"),
		Environment: getEnv("ENVIRONMENT", "development"),
		GinMode:     getEnv("GIN_MODE", "debug"),

		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost:5432/tgo_rag?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),

		VectorDBURL:    getEnv("VECTOR_DB_URL", "http://localhost:19530"),
		VectorDBAPIKey: getEnv("VECTOR_DB_API_KEY", ""),

		EmbeddingAPIKey:     getEnv("OPENAI_API_KEY", ""),
		EmbeddingBaseURL:    getEnv("EMBEDDING_BASE_URL", "https://api.openai.com/v1"),
		EmbeddingModel:      getEnv("EMBEDDING_MODEL", "text-embedding-3-small"),
		EmbeddingDimensions: int(getEnvInt64("EMBEDDING_DIMENSIONS", 1536)),

		StoragePath:   getEnv("STORAGE_PATH", "./storage"),
		MaxUploadSize: getEnvInt64("MAX_UPLOAD_SIZE", 50*1024*1024), // 50MB
	}
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultValue
}
