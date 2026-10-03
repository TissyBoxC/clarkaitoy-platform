// Package config loads voice gateway configuration.
package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// Config contains runtime settings for the voice gateway.
type Config struct {
	HTTP     HTTPConfig
	Internal InternalAPIConfig
	Log      LogConfig
	Redis    RedisConfig
	Database DatabaseConfig
	Sub2API  Sub2APIConfig
	Security SecurityConfig
}

// DatabaseConfig contains the read/write connection used for usage records.
type DatabaseConfig struct {
	DSN string
}

// HTTPConfig contains HTTP server settings.
type HTTPConfig struct {
	Host string
	Port string
}

// Address returns the HTTP listen address.
func (c HTTPConfig) Address() string {
	return c.Host + ":" + c.Port
}

// InternalAPIConfig contains service-to-service management API settings.
type InternalAPIConfig struct {
	Enabled   bool
	AuthToken string
}

// LogConfig contains logging settings.
type LogConfig struct {
	Level string
}

// SlogLevel returns the configured slog level.
func (c LogConfig) SlogLevel() slog.Level {
	switch strings.ToLower(c.Level) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// RedisConfig contains Redis settings.
type RedisConfig struct {
	Address  string
	Password string
}

// Sub2APIConfig contains the AI gateway connection.
type Sub2APIConfig struct {
	BaseURL string
	APIKey  string
}

// SecurityConfig contains content safety settings.
type SecurityConfig struct {
	ContentPolicyEnabled bool
}

// Load reads configuration from environment variables with local defaults.
func Load() (Config, error) {
	cfg := Config{
		HTTP: HTTPConfig{
			Host: env("VOICE_GATEWAY_HTTP_HOST", "0.0.0.0"),
			Port: env("VOICE_GATEWAY_HTTP_PORT", "8082"),
		},
		Internal: InternalAPIConfig{
			Enabled:   envBool("VOICE_GATEWAY_INTERNAL_API_ENABLED", false),
			AuthToken: env("VOICE_GATEWAY_INTERNAL_API_TOKEN", ""),
		},
		Log: LogConfig{
			Level: env("VOICE_GATEWAY_LOG_LEVEL", "info"),
		},
		Redis: RedisConfig{
			Address:  env("VOICE_GATEWAY_REDIS_ADDRESS", "127.0.0.1:6379"),
			Password: env("VOICE_GATEWAY_REDIS_PASSWORD", ""),
		},
		Database: DatabaseConfig{
			DSN: env("VOICE_GATEWAY_DATABASE_DSN", ""),
		},
		Sub2API: Sub2APIConfig{
			BaseURL: env("VOICE_GATEWAY_SUB2API_BASE_URL", "http://127.0.0.1:8080"),
			APIKey:  env("VOICE_GATEWAY_SUB2API_API_KEY", ""),
		},
		Security: SecurityConfig{
			ContentPolicyEnabled: true,
		},
	}

	if cfg.Internal.Enabled && len(strings.TrimSpace(cfg.Internal.AuthToken)) < 32 {
		return Config{}, fmt.Errorf("VOICE_GATEWAY_INTERNAL_API_TOKEN must contain at least 32 characters when the internal API is enabled")
	}

	return cfg, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
