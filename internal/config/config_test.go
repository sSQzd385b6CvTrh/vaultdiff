package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourusername/vaultdiff/internal/config"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "vaultdiff-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	path := writeTemp(t, `
vault:
  address: "https://vault.example.com"
  namespace: "admin"
audit:
  enabled: true
  format: "json"
  output: "/tmp/audit.log"
defaults:
  mount: "secret"
  show_unchanged: true
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Vault.Address != "https://vault.example.com" {
		t.Errorf("expected vault address, got %q", cfg.Vault.Address)
	}
	if cfg.Audit.Format != "json" {
		t.Errorf("expected json format, got %q", cfg.Audit.Format)
	}
	if !cfg.Defaults.ShowUnchanged {
		t.Error("expected show_unchanged to be true")
	}
}

func TestLoad_Defaults(t *testing.T) {
	path := writeTemp(t, `vault:\n  address: "http://localhost:8200"\n`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Vault.TokenEnv != "VAULT_TOKEN" {
		t.Errorf("expected default token_env VAULT_TOKEN, got %q", cfg.Vault.TokenEnv)
	}
	if cfg.Audit.Format != "text" {
		t.Errorf("expected default audit format text, got %q", cfg.Audit.Format)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	path := writeTemp(t, `vault: [invalid: yaml: here`)
	_, err := config.Load(path)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}
