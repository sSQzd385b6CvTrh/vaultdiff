package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/jonhadfield/vaultdiff/internal/vault"
)

var kvValidateCmd = &cobra.Command{
	Use:   "kvvalidate <path> [version]",
	Short: "Validate that a KV secret version is readable and not deleted or destroyed",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		version := 0
		if len(args) == 2 {
			v, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid version %q: %w", args[1], err)
			}
			version = v
		}

		mount, _ := cmd.Flags().GetString("mount")

		client, err := vault.NewClient(vault.ClientConfig{
			Address:   os.Getenv("VAULT_ADDR"),
			Token:     os.Getenv("VAULT_TOKEN"),
			Namespace: os.Getenv("VAULT_NAMESPACE"),
		})
		if err != nil {
			return fmt.Errorf("vault client: %w", err)
		}

		validator := vault.NewKVValidator(client, mount)
		result, err := validator.Validate(path, version)
		if err != nil {
			return fmt.Errorf("validation error: %w", err)
		}

		if result.Valid {
			fmt.Fprintf(cmd.OutOrStdout(), "OK: %s (version %d) is valid\n", result.Path, result.Version)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "INVALID: %s (version %d) — %s\n", result.Path, result.Version, result.Reason)
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	kvValidateCmd.Flags().String("mount", "secret", "KV mount path")
	rootCmd.AddCommand(kvValidateCmd)
}
