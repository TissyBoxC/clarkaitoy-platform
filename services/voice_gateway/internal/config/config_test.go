package config

import "testing"

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
