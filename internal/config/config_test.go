package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/activtrak-mfinlayson/atadmin/internal/config"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	xdg := t.TempDir()
	dir := filepath.Join(xdg, "atadmin")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", xdg)
	return path
}

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

func TestLoadProfile_Valid(t *testing.T) {
	writeTempConfig(t, `
profiles:
  default:
    token: "tok-default"
    base_url: "https://api.activtrak.com"
  staging:
    token: "tok-staging"
    base_url: "https://staging.activtrak.com"
`)
	cfg, err := config.LoadProfile("staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "tok-staging" {
		t.Errorf("expected tok-staging, got %q", cfg.Token)
	}
	if cfg.BaseURL != "https://staging.activtrak.com" {
		t.Errorf("unexpected base_url: %q", cfg.BaseURL)
	}
}

func TestLoadProfile_Missing(t *testing.T) {
	writeTempConfig(t, `
profiles:
  default:
    token: "tok-default"
`)
	_, err := config.LoadProfile("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing profile, got nil")
	}
}

func TestLoadProfile_EnvOverride(t *testing.T) {
	writeTempConfig(t, `
profiles:
  default:
    token: "tok-file"
`)
	t.Setenv("ATADMIN_TOKEN", "tok-env")
	cfg, err := config.LoadProfile("default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "tok-env" {
		t.Errorf("expected env token to win, got %q", cfg.Token)
	}
}

func TestSaveProfile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := &config.Config{
		Token:   "saved-token",
		BaseURL: "https://api.activtrak.com",
		Format:  "table",
		Timeout: 30_000_000_000,
	}
	if err := config.SaveProfile("myprofile", cfg); err != nil {
		t.Fatalf("save error: %v", err)
	}

	path := filepath.Join(dir, "atadmin", "config.yaml")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("config file not created: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected 0600 permissions, got %04o", perm)
	}

	loaded, err := config.LoadProfile("myprofile")
	if err != nil {
		t.Fatalf("load after save: %v", err)
	}
	if loaded.Token != "saved-token" {
		t.Errorf("expected saved-token, got %q", loaded.Token)
	}
}

func TestListProfiles(t *testing.T) {
	writeTempConfig(t, `
profiles:
  default:
    token: "a"
  staging:
    token: "b"
`)
	names, err := config.ListProfiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 {
		t.Errorf("expected 2 profiles, got %d: %v", len(names), names)
	}
}
