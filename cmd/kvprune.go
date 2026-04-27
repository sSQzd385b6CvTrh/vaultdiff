package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"vaultdiff/internal/vault"
)

func init() {
	var mount string
	var keep int

	cmd := &cobra.Command{
		Use:   "kvprune <path>",
		Short: "Prune old versions of a KV secret, keeping the N most recent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			if keep < 1 {
				return fmt.Errorf("--keep must be at least 1")
			}

			addr := os.Getenv("VAULT_ADDR")
			token := os.Getenv("VAULT_TOKEN")
			ns := os.Getenv("VAULT_NAMESPACE")

			client, err := vault.NewClient(addr, token, ns)
			if err != nil {
				return fmt.Errorf("create client: %w", err)
			}

			pruner := vault.NewKVPruner(client, mount)
			result, err := pruner.Prune(cmd.Context(), path, keep)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(),
				"path=%s pruned=%s skipped=%s\n",
				result.Key,
				strconv.Itoa(result.Pruned),
				strconv.Itoa(result.Skipped),
			)
			return nil
		},
	}

	cmd.Flags().StringVar(&mount, "mount", "secret", "KV mount path")
	cmd.Flags().IntVar(&keep, "keep", 5, "Number of recent versions to keep")

	rootCmd.AddCommand(cmd)
}
