package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	// Server
	Port    string `mapstructure:"PORT"`
	GinMode string `mapstructure:"GIN_MODE"`

	// Database
	DatabaseURL         string `mapstructure:"DATABASE_URL"`
	DatabasePoolSize    int    `mapstructure:"DATABASE_POOL_SIZE"`
	DatabaseMaxOverflow int    `mapstructure:"DATABASE_MAX_OVERFLOW"`

	// Redis
	RedisURL string `mapstructure:"REDIS_URL"`

	// Auth Service (apiserver)
	AuthServiceURL string `mapstructure:"AUTH_SERVICE_URL"`

	// Internal API (apiserver internal endpoint)
	InternalAPIURL string `mapstructure:"INTERNAL_API_URL"`

	// Environment
	Environment string `mapstructure:"ENVIRONMENT"`

	// External Services
	RAGServiceURL string `mapstructure:"RAG_SERVICE_URL"`
	MCPServiceURL string `mapstructure:"MCP_SERVICE_URL"`

	// LLM
	ArkAPIKey    string `mapstructure:"ARK_API_KEY"`
	ArkModel     string `mapstructure:"ARK_MODEL"`
	OpenAIAPIKey string `mapstructure:"OPENAI_API_KEY"`
	OpenAIModel  string `mapstructure:"OPENAI_MODEL"`

	// Auth
	SecretKey    string `mapstructure:"SECRET_KEY"`
	APIKeyPrefix string `mapstructure:"API_KEY_PREFIX"`

	// Logging
	LogLevel string `mapstructure:"LOG_LEVEL"`
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("PORT", "8083")
	viper.SetDefault("GIN_MODE", "release")
	viper.SetDefault("DATABASE_POOL_SIZE", 20)
	viper.SetDefault("DATABASE_MAX_OVERFLOW", 30)
	viper.SetDefault("RAG_SERVICE_URL", "http://localhost:8085")
	viper.SetDefault("MCP_SERVICE_URL", "http://localhost:8082")
	viper.SetDefault("API_KEY_PREFIX", "ak_")
	viper.SetDefault("LOG_LEVEL", "info")

	// Try to read .env file (optional)
	_ = viper.ReadInConfig()

	// Also check for environment variables directly
	cfg := &Config{}

	// Bind environment variables
	for _, key := range []string{
		"PORT", "GIN_MODE", "DATABASE_URL", "DATABASE_POOL_SIZE", "DATABASE_MAX_OVERFLOW",
		"REDIS_URL", "AUTH_SERVICE_URL", "INTERNAL_API_URL", "ENVIRONMENT", "RAG_SERVICE_URL", "MCP_SERVICE_URL",
		"ARK_API_KEY", "ARK_MODEL", "OPENAI_API_KEY", "OPENAI_MODEL",
		"SECRET_KEY", "API_KEY_PREFIX", "LOG_LEVEL",
	} {
		if val := os.Getenv(key); val != "" {
			viper.Set(key, val)
		}
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.GinMode) == "debug"
}
