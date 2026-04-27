package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/subtlepseudonym/vaultdiff/internal/vault"
)

func init() {
	var mount string
	var searchKeys bool

	kvGrepCmd := &cobra.Command{
		Use:   "kv-grep <path> <pattern>",
		Short: "Search a KV secret's keys or values for a pattern",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			pattern := args[1]

			client, err := vault.NewClient(vault.Config{
				Address:   os.Getenv("VAULT_ADDR"),
				Token:     os.Getenv("VAULT_TOKEN"),
				Namespace: os.Getenv("VAULT_NAMESPACE"),
			})
			if err != nil {
				return fmt.Errorf("init client: %w", err)
			}

			grepper := vault.NewKVGrepper(client, mount)
			result, err := grepper.Grep(cmd.Context(), path, pattern, searchKeys)
			if err != nil {
				return err
			}
			if result == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "no matches found in %s\n", path)
				return nil
			}

			keys := make([]string, 0, len(result.Matches))
			for k := range result.Matches {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			fmt.Fprintf(cmd.OutOrStdout(), "matches in %s:\n", result.Path)
			for _, k := range keys {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s = %s\n", k, result.Matches[k])
			}
			return nil
		},
	}

	kvGrepCmd.Flags().StringVar(&mount, "mount", "", "KV mount path (default: secret)")
	kvGrepCmd.Flags().BoolVar(&searchKeys, "keys", false, "search key names instead of values")
	rootCmd.AddCommand(kvGrepCmd)
}
