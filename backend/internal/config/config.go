package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv              string
	HTTPHost            string
	HTTPPort            string
	HTTPReadTimeout     time.Duration
	HTTPWriteTimeout    time.Duration
	HTTPShutdownTimeout time.Duration
	DatabaseURL         string
	RedisURL            string
	JWTAccessSecret     string
	AccessTokenTTL      time.Duration
	RefreshTokenTTL     time.Duration
	ErrorLogPath        string
}

func (c Config) HTTPAddress() string { return c.HTTPHost + ":" + c.HTTPPort }

func Load() (Config, error) {
	// Existing process variables take precedence; .env is a local-development convenience.
	_ = godotenv.Load()

	cfg := Config{
		AppEnv:          envOrDefault("APP_ENV", "development"),
		HTTPHost:        envOrDefault("HTTP_HOST", "0.0.0.0"),
		HTTPPort:        envOrDefault("HTTP_PORT", "8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		RedisURL:        envOrDefault("REDIS_URL", "redis://localhost:6379/0"),
		JWTAccessSecret: envOrDefault("JWT_ACCESS_SECRET", "local-development-secret-change-me"),
		ErrorLogPath:    envOrDefault("ERROR_LOG_PATH", "logs/error.log"),
	}

	var err error
	if cfg.HTTPReadTimeout, err = duration("HTTP_READ_TIMEOUT", 10*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.HTTPWriteTimeout, err = duration("HTTP_WRITE_TIMEOUT", 15*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.HTTPShutdownTimeout, err = duration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.AccessTokenTTL, err = duration("ACCESS_TOKEN_TTL", 15*time.Minute); err != nil {
		return Config{}, err
	}
	if cfg.RefreshTokenTTL, err = duration("REFRESH_TOKEN_TTL", 30*24*time.Hour); err != nil {
		return Config{}, err
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.AppEnv == "production" && cfg.JWTAccessSecret == "local-development-secret-change-me" {
		return Config{}, fmt.Errorf("JWT_ACCESS_SECRET is required in production")
	}
	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func duration(key string, fallback time.Duration) (time.Duration, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return value, nil
}
