package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultdiff/internal/vault"
)

var renewCmd = &cobra.Command{
	Use:   "renew",
	Short: "Renew the current Vault token",
	RunE: func(cmd *cobra.Command, args []string) error {
		address := os.Getenv("VAULT_ADDR")
		if address == "" {
			address = "http://127.0.0.1:8200"
		}
		token := os.Getenv("VAULT_TOKEN")
		if token == "" {
			return fmt.Errorf("VAULT_TOKEN is not set")
		}

		renewer := vault.NewTokenRenewer(address, token)
		result, err := renewer.Renew()
		if err != nil {
			return fmt.Errorf("renewing token: %w", err)
		}

		fmt.Printf("Token renewed successfully\n")
		fmt.Printf("  Client Token  : %s\n", result.ClientToken)
		fmt.Printf("  Lease Duration: %s\n", result.LeaseDuration)
		fmt.Printf("  Renewable     : %v\n", result.Renewable)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(renewCmd)
}
