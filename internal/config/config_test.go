package config_test

import (
	"testing"

	"github.com/activtrak-mfinlayson/atadmin/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Format != "table" {
		t.Errorf("default Format = %q, want %q", cfg.Format, "table")
	}
	if cfg.BaseURL != "https://api.activtrak.com" {
		t.Errorf("default BaseURL = %q, want %q", cfg.BaseURL, "https://api.activtrak.com")
	}
	if cfg.Timeout <= 0 {
		t.Error("default Timeout must be positive")
	}
}

func TestLoadEnvVarToken(t *testing.T) {
	t.Setenv("ATADMIN_TOKEN", "testtoken123")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Token != "testtoken123" {
		t.Errorf("Token = %q, want %q", cfg.Token, "testtoken123")
	}
}

func TestLoadEnvVarFormat(t *testing.T) {
	t.Setenv("ATADMIN_FORMAT", "json")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Format != "json" {
		t.Errorf("Format = %q, want %q", cfg.Format, "json")
	}
}
