// Package config loads device platform configuration.
package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// Config contains runtime settings for the device platform service.
type Config struct {
	HTTP     HTTPConfig
	Internal InternalAPIConfig
	Log      LogConfig
	Database DatabaseConfig
	Redis    RedisConfig
	MQTT     MQTTConfig
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

// DatabaseConfig contains PostgreSQL settings.
type DatabaseConfig struct {
	DSN string
}

// RedisConfig contains Redis settings.
type RedisConfig struct {
	Address  string
	Password string
	DB       int
}

// MQTTConfig contains MQTT broker settings.
type MQTTConfig struct {
	Broker   string
	ClientID string
	Username string
	Password string
}

// Load reads configuration from environment variables with local defaults.
func Load() (Config, error) {
	cfg := Config{
		HTTP: HTTPConfig{
			Host: env("DEVICE_PLATFORM_HTTP_HOST", "0.0.0.0"),
			Port: env("DEVICE_PLATFORM_HTTP_PORT", "8081"),
		},
		Internal: InternalAPIConfig{
			Enabled:   envBool("DEVICE_PLATFORM_INTERNAL_API_ENABLED", false),
			AuthToken: env("DEVICE_PLATFORM_INTERNAL_API_TOKEN", ""),
		},
		Log: LogConfig{
			Level: env("DEVICE_PLATFORM_LOG_LEVEL", "info"),
		},
		Database: DatabaseConfig{
			DSN: env("DEVICE_PLATFORM_DATABASE_DSN", ""),
		},
		Redis: RedisConfig{
			Address:  env("DEVICE_PLATFORM_REDIS_ADDRESS", "127.0.0.1:6379"),
			Password: env("DEVICE_PLATFORM_REDIS_PASSWORD", ""),
			DB:       envInt("DEVICE_PLATFORM_REDIS_DB", 0),
		},
		MQTT: MQTTConfig{
			Broker:   env("DEVICE_PLATFORM_MQTT_BROKER", "tcp://127.0.0.1:1883"),
			ClientID: env("DEVICE_PLATFORM_MQTT_CLIENT_ID", "device-platform"),
			Username: env("DEVICE_PLATFORM_MQTT_USERNAME", ""),
			Password: env("DEVICE_PLATFORM_MQTT_PASSWORD", ""),
		},
	}

	if cfg.Internal.Enabled && len(strings.TrimSpace(cfg.Internal.AuthToken)) < 32 {
		return Config{}, fmt.Errorf("DEVICE_PLATFORM_INTERNAL_API_TOKEN must contain at least 32 characters when the internal API is enabled")
	}

	return cfg, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
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
