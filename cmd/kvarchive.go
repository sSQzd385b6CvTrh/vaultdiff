package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultdiff/internal/vault"
)

var kvArchiveCmd = &cobra.Command{
	Use:   "kv-archive <path>",
	Short: "Archive all versions of a KV v2 secret",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr := os.Getenv("VAULT_ADDR")
		token := os.Getenv("VAULT_TOKEN")
		mount, _ := cmd.Flags().GetString("mount")

		client, err := vault.NewClient(addr, token, "")
		if err != nil {
			return fmt.Errorf("failed to create vault client: %w", err)
		}

		archiver := vault.NewKVArchiver(client.API(), mount)
		entry, err := archiver.Archive(context.Background(), args[0])
		if err != nil {
			return fmt.Errorf("archive failed: %w", err)
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "VERSION\tKEYS")

		versions := make([]string, 0, len(entry.Versions))
		for v := range entry.Versions {
			versions = append(versions, v)
		}
		sort.Strings(versions)

		for _, ver := range versions {
			data := entry.Versions[ver]
			keys := make([]string, 0, len(data))
			for k := range data {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			fmt.Fprintf(w, "%s\t%v\n", ver, keys)
		}
		return w.Flush()
	},
}

func init() {
	kvArchiveCmd.Flags().String("mount", "secret", "KV v2 mount path")
	rootCmd.AddCommand(kvArchiveCmd)
}
