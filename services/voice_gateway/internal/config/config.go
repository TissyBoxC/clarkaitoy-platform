// Package config loads voice gateway configuration.
package config

import (
	"log/slog"
	"os"
	"strings"
)

// Config contains runtime settings for the voice gateway.
type Config struct {
	HTTP     HTTPConfig
	Log      LogConfig
	Redis    RedisConfig
	Sub2API  Sub2APIConfig
	Security SecurityConfig
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

// LogConfig contains logging settings.
type LogConfig struct {
	Level string
}

// Level returns the configured slog level.
func (c LogConfig) Level() slog.Level {
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
	return Config{
		HTTP: HTTPConfig{
			Host: env("VOICE_GATEWAY_HTTP_HOST", "0.0.0.0"),
			Port: env("VOICE_GATEWAY_HTTP_PORT", "8082"),
		},
		Log: LogConfig{
			Level: env("VOICE_GATEWAY_LOG_LEVEL", "info"),
		},
		Redis: RedisConfig{
			Address:  env("VOICE_GATEWAY_REDIS_ADDRESS", "127.0.0.1:6379"),
			Password: env("VOICE_GATEWAY_REDIS_PASSWORD", ""),
		},
		Sub2API: Sub2APIConfig{
			BaseURL: env("VOICE_GATEWAY_SUB2API_BASE_URL", "http://127.0.0.1:8080"),
			APIKey:  env("VOICE_GATEWAY_SUB2API_API_KEY", ""),
		},
		Security: SecurityConfig{
			ContentPolicyEnabled: true,
		},
	}, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
