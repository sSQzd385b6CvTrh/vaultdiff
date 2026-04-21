package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"vaultdiff/internal/vault"
)

var kvExpireCmd = &cobra.Command{
	Use:   "kvexpire <path>",
	Short: "Check expiry metadata for a KV v2 secret",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		mount, _ := cmd.Flags().GetString("mount")

		client, err := vault.NewClient(
			os.Getenv("VAULT_ADDR"),
			os.Getenv("VAULT_TOKEN"),
			os.Getenv("VAULT_NAMESPACE"),
		)
		if err != nil {
			return fmt.Errorf("failed to create vault client: %w", err)
		}

		expirer := vault.NewKVExpirer(client, mount)
		info, err := expirer.CheckExpiry(path)
		if err != nil {
			return fmt.Errorf("expiry check failed: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Path:         %s\n", info.Path)
		fmt.Fprintf(cmd.OutOrStdout(), "Version:      %d\n", info.Version)
		fmt.Fprintf(cmd.OutOrStdout(), "Created:      %s\n", info.CreatedTime.Format("2006-01-02 15:04:05 UTC"))

		if info.ExpiresAt.IsZero() {
			fmt.Fprintf(cmd.OutOrStdout(), "Expires:      never\n")
		} else {
			status := "active"
			if info.Expired {
				status = "EXPIRED"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Expires:      %s\n", info.ExpiresAt.Format("2006-01-02 15:04:05 UTC"))
			fmt.Fprintf(cmd.OutOrStdout(), "TTL:          %s\n", info.TTL.Round(1e9))
			fmt.Fprintf(cmd.OutOrStdout(), "Status:       %s\n", status)
		}

		return nil
	},
}

func init() {
	kvExpireCmd.Flags().String("mount", "secret", "KV v2 mount path")
	rootCmd.AddCommand(kvExpireCmd)
}
