package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultdiff/internal/vault"
)

var (
	kvListMount string
	kvListPath  string
)

var kvListCmd = &cobra.Command{
	Use:   "kvlist",
	Short: "List keys under a KV v2 path",
	RunE: func(cmd *cobra.Command, args []string) error {
		address := os.Getenv("VAULT_ADDR")
		if address == "" {
			address = "http://127.0.0.1:8200"
		}
		token := os.Getenv("VAULT_TOKEN")
		if token == "" {
			return fmt.Errorf("VAULT_TOKEN is not set")
		}

		lister := vault.NewKVLister(address, token, kvListMount)
		keys, err := lister.ListKeys(kvListPath)
		if err != nil {
			return fmt.Errorf("list keys: %w", err)
		}

		if len(keys) == 0 {
			fmt.Println("No keys found.")
			return nil
		}

		fmt.Printf("Keys under %s/%s:\n", kvListMount, kvListPath)
		for _, k := range keys {
			prefix := "  "
			if strings.HasSuffix(k, "/") {
				prefix = "  [dir] "
			}
			fmt.Printf("%s%s\n", prefix, k)
		}
		return nil
	},
}

func init() {
	kvListCmd.Flags().StringVar(&kvListMount, "mount", "secret", "KV v2 mount path")
	kvListCmd.Flags().StringVar(&kvListPath, "path", "", "Path prefix to list keys under")
	_ = kvListCmd.MarkFlagRequired("path")
	rootCmd.AddCommand(kvListCmd)
}
