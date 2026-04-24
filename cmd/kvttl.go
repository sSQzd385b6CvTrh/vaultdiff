package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jonasvinther/vaultdiff/internal/vault"
)

var kvttlCmd = &cobra.Command{
	Use:   "kvttl <path>",
	Short: "Show TTL information for a KV v2 secret",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		address, _ := cmd.Flags().GetString("address")
		token, _ := cmd.Flags().GetString("token")
		mount, _ := cmd.Flags().GetString("mount")

		if address == "" {
			address = os.Getenv("VAULT_ADDR")
		}
		if token == "" {
			token = os.Getenv("VAULT_TOKEN")
		}

		client, err := vault.NewClient(address, token, "")
		if err != nil {
			return fmt.Errorf("creating vault client: %w", err)
		}

		reader := vault.NewKVTTLReader(client.API(), mount)
		info, err := reader.GetTTL(path)
		if err != nil {
			return err
		}

		if !info.HasTTL {
			fmt.Fprintf(cmd.OutOrStdout(), "Path: %s\nTTL:  none\n", info.Path)
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(),
			"Path:          %s\nDeletion Time: %s\nRemaining:     %s\n",
			info.Path,
			info.DeletionTime.Format("2006-01-02T15:04:05Z"),
			info.Remaining.String(),
		)
		return nil
	},
}

func init() {
	kvttlCmd.Flags().String("address", "", "Vault server address")
	kvttlCmd.Flags().String("token", "", "Vault token")
	kvttlCmd.Flags().String("mount", "secret", "KV v2 mount path")
	rootCmd.AddCommand(kvttlCmd)
}
