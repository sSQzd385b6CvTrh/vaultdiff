package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/yourusername/vaultdiff/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage vaultdiff configuration",
}

var configValidateCmd = &cobra.Command{
	Use:   "validate [file]",
	Short: "Validate a vaultdiff config file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(args[0])
		if err != nil {
			return fmt.Errorf("invalid config: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Config is valid.\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  Vault address : %s\n", cfg.Vault.Address)
		fmt.Fprintf(cmd.OutOrStdout(), "  Audit format  : %s\n", cfg.Audit.Format)
		fmt.Fprintf(cmd.OutOrStdout(), "  Default mount : %s\n", cfg.Defaults.Mount)
		return nil
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init [file]",
	Short: "Write a default config file",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		outPath := ".vaultdiff.yaml"
		if len(args) == 1 {
			outPath = args[0]
		}
		defaultCfg := config.Config{
			Vault: config.VaultConfig{
				Address:  "http://127.0.0.1:8200",
				TokenEnv: "VAULT_TOKEN",
			},
			Audit: config.AuditConfig{
				Enabled: false,
				Format:  "text",
			},
			Defaults: config.DefaultsConfig{
				Mount:         "secret",
				ShowUnchanged: false,
			},
		}
		f, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("creating config file: %w", err)
		}
		defer f.Close()
		if err := yaml.NewEncoder(f).Encode(defaultCfg); err != nil {
			return fmt.Errorf("writing config: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Config written to %s\n", outPath)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configInitCmd)
	rootCmd.AddCommand(configCmd)
}
