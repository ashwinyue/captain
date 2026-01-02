package config

import (
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Port          string `mapstructure:"PORT"`
	GinMode       string `mapstructure:"GIN_MODE"`
	DatabaseURL   string `mapstructure:"DATABASE_URL"`
	RedisURL      string `mapstructure:"REDIS_URL"`
	Environment   string `mapstructure:"ENVIRONMENT"`
	AuthServerURL string `mapstructure:"AUTH_SERVER_URL"`
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	viper.SetDefault("PORT", "8086")
	viper.SetDefault("GIN_MODE", "release")
	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("AUTH_SERVER_URL", "http://localhost:8080")

	_ = viper.ReadInConfig()

	cfg := &Config{}
	for _, key := range []string{
		"PORT", "GIN_MODE", "DATABASE_URL", "REDIS_URL", "ENVIRONMENT", "AUTH_SERVER_URL",
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
