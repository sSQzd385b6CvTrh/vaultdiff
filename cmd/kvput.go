package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"vaultdiff/internal/vault"
)

var kvputCmd = &cobra.Command{
	Use:   "kvput <path> key=value [key=value...]",
	Short: "Write key-value pairs to a Vault KV v2 secret",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		data := make(map[string]string)
		for _, pair := range args[1:] {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid key=value pair: %q", pair)
			}
			data[parts[0]] = parts[1]
		}

		address, _ := cmd.Flags().GetString("address")
		token := os.Getenv("VAULT_TOKEN")
		mount, _ := cmd.Flags().GetString("mount")

		w := vault.NewKVWriter(address, token, mount)
		version, err := w.Put(path, data)
		if err != nil {
			return fmt.Errorf("write secret: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Written %s (version %d)\n", path, version)
		return nil
	},
}

func init() {
	kvputCmd.Flags().String("address", "http://127.0.0.1:8200", "Vault address")
	kvputCmd.Flags().String("mount", "secret", "KV v2 mount path")
	rootCmd.AddCommand(kvputCmd)
}
