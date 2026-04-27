package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jonboulle/vaultdiff/internal/vault"
)

var kvBulkGetCmd = &cobra.Command{
	Use:   "kvbulkget [key1] [key2] ...",
	Short: "Fetch multiple KV secrets in one command",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr := os.Getenv("VAULT_ADDR")
		token := os.Getenv("VAULT_TOKEN")
		ns := os.Getenv("VAULT_NAMESPACE")
		mount, _ := cmd.Flags().GetString("mount")

		client, err := vault.NewClient(addr, token, ns)
		if err != nil {
			return fmt.Errorf("vault client: %w", err)
		}

		getter := vault.NewKVBulkGetter(client, mount)
		results := getter.Get(context.Background(), args)

		for _, r := range results {
			if r.Err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "ERROR %s: %v\n", r.Key, r.Err)
				continue
			}
			fmt.Fprintf(cmd.OutOrStdout(), "[%s]\n", r.Key)
			for k, v := range r.Data {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s = %v\n", k, v)
			}
		}
		return nil
	},
}

func init() {
	kvBulkGetCmd.Flags().String("mount", "", "KV mount path (default: secret)")
	rootCmd.AddCommand(kvBulkGetCmd)

	_ = strings.ToLower // suppress unused import
}
