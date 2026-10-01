package config

import (
	"strings"
	"testing"
)

func TestLoadReadsServiceSettings(t *testing.T) {
	t.Setenv("VOICE_GATEWAY_HTTP_PORT", "9090")
	t.Setenv("VOICE_GATEWAY_SUB2API_BASE_URL", "http://sub2api.internal")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
	if cfg.HTTP.Port != "9090" {
		t.Fatalf("expected HTTP port 9090, got %q", cfg.HTTP.Port)
	}
	if cfg.Sub2API.BaseURL != "http://sub2api.internal" {
		t.Fatalf("unexpected sub2api URL: %q", cfg.Sub2API.BaseURL)
	}
}

func TestLoadUsesConservativeSecurityDefaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
	if !cfg.Security.ContentPolicyEnabled {
		t.Fatal("content policy must be enabled by default")
	}
	if cfg.Internal.Enabled {
		t.Fatal("internal API must be disabled by default")
	}
}

func TestLoadReadsInternalAPISettings(t *testing.T) {
	t.Setenv("VOICE_GATEWAY_INTERNAL_API_ENABLED", "true")
	t.Setenv("VOICE_GATEWAY_INTERNAL_API_TOKEN", strings.Repeat("t", 32))

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
	if !cfg.Internal.Enabled {
		t.Fatal("expected internal API to be enabled")
	}
	if cfg.Internal.AuthToken != strings.Repeat("t", 32) {
		t.Fatalf("unexpected internal API token: %q", cfg.Internal.AuthToken)
	}
}

func TestLoadRejectsEnabledInternalAPIWithoutStrongToken(t *testing.T) {
	t.Setenv("VOICE_GATEWAY_INTERNAL_API_ENABLED", "true")
	t.Setenv("VOICE_GATEWAY_INTERNAL_API_TOKEN", "short-token")

	if _, err := Load(); err == nil {
		t.Fatal("expected weak internal API token to be rejected")
	}
}
