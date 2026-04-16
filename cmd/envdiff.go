package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"vaultdiff/internal/diff"
	"vaultdiff/internal/vault"
)

var (
	srcEnv     string
	dstEnv     string
	envPath    string
	srcVersion int
	dstVersion int
)

var envdiffCmd = &cobra.Command{
	Use:   "envdiff",
	Short: "Diff a secret path between two Vault environments",
	Long:  `Fetches a KV secret from two separate Vault environments and prints a unified diff.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		envConfigs, err := buildEnvConfigs(cfg)
		if err != nil {
			return err
		}

		ec, err := vault.NewEnvComparer(envConfigs)
		if err != nil {
			return fmt.Errorf("init env comparer: %w", err)
		}

		src, dst, err := ec.Compare(cmd.Context(), srcEnv, dstEnv, envPath, srcVersion, dstVersion)
		if err != nil {
			return fmt.Errorf("compare: %w", err)
		}

		changes := diff.Secrets(src, dst)
		return diff.Write(os.Stdout, changes, showUnchanged)
	},
}

// buildEnvConfigs converts the loaded config into a map of vault.EnvConfig values,
// validating that the requested src and dst environments are present.
func buildEnvConfigs(cfg *Config) (map[string]vault.EnvConfig, error) {
	envConfigs := make(map[string]vault.EnvConfig, len(cfg.Environments))
	for name, e := range cfg.Environments {
		envConfigs[name] = vault.EnvConfig{
			Address:   e.Address,
			Namespace: e.Namespace,
			Token:     e.Token,
		}
	}
	if _, ok := envConfigs[srcEnv]; !ok {
		return nil, fmt.Errorf("source environment %q not found in config", srcEnv)
	}
	if _, ok := envConfigs[dstEnv]; !ok {
		return nil, fmt.Errorf("destination environment %q not found in config", dstEnv)
	}
	return envConfigs, nil
}

func init() {
	rootCmd.AddCommand(envdiffCmd)
	envdiffCmd.Flags().StringVar(&srcEnv, "src-env", "", "Source environment name (required)")
	envdiffCmd.Flags().StringVar(&dstEnv, "dst-env", "", "Destination environment name (required)")
	envdiffCmd.Flags().StringVar(&envPath, "path", "", "Secret path to compare (required)")
	envdiffCmd.Flags().IntVar(&srcVersion, "src-version", 0, "Source secret version (0 = latest)")
	envdiffCmd.Flags().IntVar(&dstVersion, "dst-version", 0, "Destination secret version (0 = latest)")
	envdiffCmd.Flags().BoolVar(&showUnchanged, "show-unchanged", false, "Include unchanged keys in output")
	_ = envdiffCmd.MarkFlagRequired("src-env")
	_ = envdiffCmd.MarkFlagRequired("dst-env")
	_ = envdiffCmd.MarkFlagRequired("path")
}
