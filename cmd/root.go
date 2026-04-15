package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	vaultAddr      string
	vaultToken     string
	vaultNamespace string
)

var rootCmd = &cobra.Command{
	Use:   "vaultdiff",
	Short: "Diff and audit HashiCorp Vault secret versions across environments",
	Long: `vaultdiff is a CLI tool for comparing Vault KV v2 secret versions.

It allows you to diff secrets between versions or across different Vault
instances and namespaces, making secret auditing and change tracking easy.`,
	SilenceUsage: true,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&vaultAddr, "vault-addr", "",
		"Vault server address (overrides VAULT_ADDR env var)",
	)
	rootCmd.PersistentFlags().StringVar(
		&vaultToken, "vault-token", "",
		"Vault token (overrides VAULT_TOKEN env var)",
	)
	rootCmd.PersistentFlags().StringVar(
		&vaultNamespace, "namespace", "",
		"Vault namespace (overrides VAULT_NAMESPACE env var)",
	)
}
