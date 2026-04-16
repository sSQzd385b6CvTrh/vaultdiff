package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultdiff/internal/vault"
)

var transitMount string

var transitCmd = &cobra.Command{
	Use:   "transit",
	Short: "List transit encryption keys in Vault",
	RunE: func(cmd *cobra.Command, args []string) error {
		address := os.Getenv("VAULT_ADDR")
		if address == "" {
			address = "http://127.0.0.1:8200"
		}
		token := os.Getenv("VAULT_TOKEN")
		if token == "" {
			return fmt.Errorf("VAULT_TOKEN is not set")
		}

		lister := vault.NewTransitLister(address, token, transitMount)
		keys, err := lister.ListKeys()
		if err != nil {
			return fmt.Errorf("failed to list transit keys: %w", err)
		}

		if len(keys) == 0 {
			fmt.Println("No transit keys found.")
			return nil
		}

		fmt.Printf("Transit keys (mount: %s):\n", lister.Mount)
		for _, k := range keys {
			fmt.Printf("  - %s\n", k)
		}
		return nil
	},
}

func init() {
	transitCmd.Flags().StringVar(&transitMount, "mount", "", "Transit secrets engine mount path (default: transit)")
	rootCmd.AddCommand(transitCmd)
}
