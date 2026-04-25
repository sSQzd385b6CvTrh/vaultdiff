package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultdiff/internal/vault"
)

var kvCountCmd = &cobra.Command{
	Use:   "kv-count <path>",
	Short: "Count the number of keys under a KV path",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		address := os.Getenv("VAULT_ADDR")
		if address == "" {
			address = "http://127.0.0.1:8200"
		}
		token := os.Getenv("VAULT_TOKEN")
		mount, _ := cmd.Flags().GetString("mount")

		counter := vault.NewKVCounter(http.DefaultClient, address, token, mount)
		result, err := counter.Count(cmd.Context(), path)
		if err != nil {
			return fmt.Errorf("kv-count failed: %w", err)
		}

		if result.Count == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "No keys found under %q\n", path)
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Path: %s\n", result.Path)
		fmt.Fprintf(cmd.OutOrStdout(), "Count: %d\n", result.Count)
		for _, k := range result.Keys {
			fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", k)
		}
		return nil
	},
}

func init() {
	kvCountCmd.Flags().String("mount", "secret", "KV mount path")
	rootCmd.AddCommand(kvCountCmd)
}
