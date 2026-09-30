// Package config loads device platform configuration.
package config

import (
	"log/slog"
	"os"
	"strings"
)

// Config contains runtime settings for the device platform service.
type Config struct {
	HTTP    HTTPConfig
	Log     LogConfig
	Database DatabaseConfig
	Redis   RedisConfig
	MQTT    MQTTConfig
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
	return Config{
		HTTP: HTTPConfig{
			Host: env("DEVICE_PLATFORM_HTTP_HOST", "0.0.0.0"),
			Port: env("DEVICE_PLATFORM_HTTP_PORT", "8081"),
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
			DB:       0,
		},
		MQTT: MQTTConfig{
			Broker:   env("DEVICE_PLATFORM_MQTT_BROKER", "tcp://127.0.0.1:1883"),
			ClientID: env("DEVICE_PLATFORM_MQTT_CLIENT_ID", "device-platform"),
			Username: env("DEVICE_PLATFORM_MQTT_USERNAME", ""),
			Password: env("DEVICE_PLATFORM_MQTT_PASSWORD", ""),
		},
	}, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
