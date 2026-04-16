package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jonathonlacher/vaultdiff/internal/vault"
)

var rekeyCmd = &cobra.Command{
	Use:   "rekey",
	Short: "Show the current Vault rekey operation status",
	RunE: func(cmd *cobra.Command, args []string) error {
		address := os.Getenv("VAULT_ADDR")
		if address == "" {
			address = "https://127.0.0.1:8200"
		}
		token := os.Getenv("VAULT_TOKEN")

		checker := vault.NewRekeyChecker(address, token)
		status, err := checker.RekeyStatusResult(context.Background())
		if err != nil {
			return fmt.Errorf("rekey status: %w", err)
		}

		if !status.Started {
			fmt.Fprintln(cmd.OutOrStdout(), "No rekey operation in progress.")
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Rekey in progress\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  Threshold : %d\n", status.T)
		fmt.Fprintf(cmd.OutOrStdout(), "  Shares    : %d\n", status.N)
		fmt.Fprintf(cmd.OutOrStdout(), "  Progress  : %d / %d\n", status.Progress, status.Required)
		fmt.Fprintf(cmd.OutOrStdout(), "  Backup    : %v\n", status.Backup)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rekeyCmd)
}
