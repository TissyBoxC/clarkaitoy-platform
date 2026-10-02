package config

import (
	"strings"
	"testing"
)

func TestLoadUsesTLSMQTTDefaults(t *testing.T) {
	setAuthenticationTestSecrets(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
	if cfg.MQTT.Broker != "tls://127.0.0.1:8883" {
		t.Fatalf("expected TLS MQTT default, got %q", cfg.MQTT.Broker)
	}
	if cfg.MQTT.InsecureSkipVerify {
		t.Fatal("MQTT certificate verification must remain enabled by default")
	}
}

func TestLoadRejectsInsecureMQTTVerification(t *testing.T) {
	setAuthenticationTestSecrets(t)
	t.Setenv("DEVICE_PLATFORM_MQTT_INSECURE_SKIP_VERIFY", "true")

	if _, err := Load(); err == nil {
		t.Fatal("expected insecure MQTT verification to be rejected")
	}
}

func TestLoadReadsMQTTMutualTLSSettings(t *testing.T) {
	setAuthenticationTestSecrets(t)
	t.Setenv("DEVICE_PLATFORM_MQTT_CA_FILE", "ca.crt")
	t.Setenv("DEVICE_PLATFORM_MQTT_CLIENT_CERTIFICATE_FILE", "device.crt")
	t.Setenv("DEVICE_PLATFORM_MQTT_CLIENT_KEY_FILE", "device.key")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
	if cfg.MQTT.CAFile != "ca.crt" {
		t.Fatalf("unexpected CA file: %q", cfg.MQTT.CAFile)
	}
	if cfg.MQTT.ClientCertificateFile != "device.crt" {
		t.Fatalf("unexpected client certificate file: %q", cfg.MQTT.ClientCertificateFile)
	}
	if !strings.HasSuffix(cfg.MQTT.ClientKeyFile, ".key") {
		t.Fatalf("unexpected client key file: %q", cfg.MQTT.ClientKeyFile)
	}
}

func setAuthenticationTestSecrets(t *testing.T) {
	t.Helper()
	t.Setenv(
		"DEVICE_PLATFORM_AUTH_ACCESS_TOKEN_SECRET",
		"test-access-token-secret-at-least-32-characters",
	)
	t.Setenv(
		"DEVICE_PLATFORM_AI_CREDENTIAL_KEY",
		"test-ai-credential-key-at-least-32-characters",
	)
	t.Setenv(
		"DEVICE_PLATFORM_MFA_CREDENTIAL_KEY",
		"test-mfa-credential-key-at-least-32-characters",
	)
}
