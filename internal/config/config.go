package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level vaultdiff configuration.
type Config struct {
	Vault    VaultConfig    `yaml:"vault"`
	Audit    AuditConfig    `yaml:"audit"`
	Defaults DefaultsConfig `yaml:"defaults"`
}

// VaultConfig holds Vault connection settings.
type VaultConfig struct {
	Address   string `yaml:"address"`
	Namespace string `yaml:"namespace"`
	TokenEnv  string `yaml:"token_env"`
}

// AuditConfig holds audit logging settings.
type AuditConfig struct {
	Enabled bool   `yaml:"enabled"`
	Format  string `yaml:"format"`
	Output  string `yaml:"output"`
}

// DefaultsConfig holds default CLI flag values.
type DefaultsConfig struct {
	Mount      string `yaml:"mount"`
	ShowUnchanged bool `yaml:"show_unchanged"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening config file: %w", err)
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if cfg.Vault.TokenEnv == "" {
		cfg.Vault.TokenEnv = "VAULT_TOKEN"
	}
	if cfg.AuditConfig().Format == "" {
		cfg.Audit.Format = "text"
	}

	return &cfg, nil
}

// AuditConfig returns the audit sub-config (helper for readability).
func (c *Config) AuditConfig() AuditConfig {
	return c.Audit
}
