package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultdiff/internal/vault"
)

var rollbackMount string

var rollbackCmd = &cobra.Command{
	Use:   "rollback <path> <version>",
	Short: "Roll back a KV secret to a previous version",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		version, err := strconv.Atoi(args[1])
		if err != nil || version < 1 {
			return fmt.Errorf("version must be a positive integer, got %q", args[1])
		}

		client, err := vault.NewClient()
		if err != nil {
			return fmt.Errorf("vault client: %w", err)
		}

		rb := vault.NewRollbacker(client, rollbackMount)
		result, err := rb.Rollback(cmd.Context(), path, version)
		if err != nil {
			return fmt.Errorf("rollback failed: %w", err)
		}

		fmt.Fprintf(os.Stdout, "Rolled back %q to version %d (new version: %d)\n",
			result.Path, result.ToVersion, result.NewVersion)
		return nil
	},
}

func init() {
	rollbackCmd.Flags().StringVar(&rollbackMount, "mount", "secret", "KV mount path")
	rootCmd.AddCommand(rollbackCmd)
}
