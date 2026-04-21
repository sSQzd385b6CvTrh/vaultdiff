package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jonnylangefeld/vaultdiff/internal/vault"
)

var kvSearchCmd = &cobra.Command{
	Use:   "kvsearch <prefix> <query>",
	Short: "Search for secret keys matching a query under a KV prefix",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		prefix := args[0]
		query := args[1]

		mount, _ := cmd.Flags().GetString("mount")

		client, err := vault.NewClient()
		if err != nil {
			return fmt.Errorf("vault client: %w", err)
		}

		searcher := vault.NewKVSearcher(client.API(), mount)
		results, err := searcher.Search(cmd.Context(), prefix, query)
		if err != nil {
			return fmt.Errorf("search: %w", err)
		}

		if len(results) == 0 {
			fmt.Fprintln(os.Stdout, "no matches found")
			return nil
		}

		for _, r := range results {
			fmt.Fprintf(os.Stdout, "%s\n", r.Path)
			for _, k := range r.MatchedKeys {
				fmt.Fprintf(os.Stdout, "  key: %s\n", k)
			}
		}
		return nil
	},
}

func init() {
	kvSearchCmd.Flags().String("mount", "secret", "KV mount path")
	rootCmd.AddCommand(kvSearchCmd)
}
