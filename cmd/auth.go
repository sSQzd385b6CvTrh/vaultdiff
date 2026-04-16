package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultdiff/internal/vault"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Inspect current Vault token auth info",
	RunE: func(cmd *cobra.Command, args []string) error {
		address := os.Getenv("VAULT_ADDR")
		if address == "" {
			address = "http://127.0.0.1:8200"
		}
		token := os.Getenv("VAULT_TOKEN")
		if token == "" {
			return fmt.Errorf("VAULT_TOKEN environment variable is not set")
		}

		inspector := vault.NewAuthInspector(address, token)
		info, err := inspector.LookupAuth()
		if err != nil {
			return fmt.Errorf("auth lookup failed: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Accessor:     %s\n", info.Accessor)
		fmt.Fprintf(cmd.OutOrStdout(), "Display Name: %s\n", info.DisplayName)
		fmt.Fprintf(cmd.OutOrStdout(), "Policies:     %s\n", strings.Join(info.Policies, ", "))
		fmt.Fprintf(cmd.OutOrStdout(), "TTL:          %ds\n", info.TTL)
		fmt.Fprintf(cmd.OutOrStdout(), "Renewable:    %v\n", info.Renewable)

		if len(info.Meta) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Metadata:\n")
			for k, v := range info.Meta {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s: %s\n", k, v)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
}
