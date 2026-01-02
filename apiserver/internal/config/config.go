package config

import (
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	// Server
	Port        string `mapstructure:"PORT"`
	GinMode     string `mapstructure:"GIN_MODE"`
	Environment string `mapstructure:"ENVIRONMENT"`

	// Database
	DatabaseURL string `mapstructure:"DATABASE_URL"`

	// Redis
	RedisURL string `mapstructure:"REDIS_URL"`

	// JWT
	JWTSecret              string `mapstructure:"JWT_SECRET"`
	AccessTokenExpireMin   int    `mapstructure:"ACCESS_TOKEN_EXPIRE_MINUTES"`
	RefreshTokenExpireDays int    `mapstructure:"REFRESH_TOKEN_EXPIRE_DAYS"`

	// WuKongIM
	WuKongIMURL     string `mapstructure:"WUKONGIM_URL"`
	WuKongIMAPIKey  string `mapstructure:"WUKONGIM_API_KEY"`
	WuKongIMWSURL   string `mapstructure:"WUKONGIM_WS_URL"`

	// External Services
	AICenterURL   string `mapstructure:"AICENTER_URL"`
	PlatformURL   string `mapstructure:"PLATFORM_URL"`
	RAGServiceURL string `mapstructure:"RAG_SERVICE_URL"`
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Defaults
	viper.SetDefault("PORT", "8000")
	viper.SetDefault("GIN_MODE", "release")
	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("ACCESS_TOKEN_EXPIRE_MINUTES", 60)
	viper.SetDefault("REFRESH_TOKEN_EXPIRE_DAYS", 7)
	viper.SetDefault("AICENTER_URL", "http://localhost:8083")
	viper.SetDefault("PLATFORM_URL", "http://localhost:8086")
	viper.SetDefault("WUKONGIM_URL", "http://localhost:5001")

	_ = viper.ReadInConfig()

	cfg := &Config{}
	for _, key := range []string{
		"PORT", "GIN_MODE", "ENVIRONMENT", "DATABASE_URL", "REDIS_URL",
		"JWT_SECRET", "ACCESS_TOKEN_EXPIRE_MINUTES", "REFRESH_TOKEN_EXPIRE_DAYS",
		"WUKONGIM_URL", "WUKONGIM_API_KEY", "WUKONGIM_WS_URL", "AICENTER_URL", "PLATFORM_URL", "RAG_SERVICE_URL",
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
	return c.Environment == "development"
}
