package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/jonboydell/vaultdiff/internal/vault"
)

var kvSnapshotDiffCmd = &cobra.Command{
	Use:   "kv-snapshot-diff <path-a> <path-b>",
	Short: "Diff two KV secret paths and report added, removed, and modified keys",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr := os.Getenv("VAULT_ADDR")
		token := os.Getenv("VAULT_TOKEN")
		mount, _ := cmd.Flags().GetString("mount")

		client, err := vault.NewClient(addr, token, "")
		if err != nil {
			return fmt.Errorf("vault client: %w", err)
		}

		differ := vault.NewKVSnapshotDiffer(client, mount)
		result, err := differ.Compare(args[0], args[1])
		if err != nil {
			return err
		}

		printSection := func(label string, m map[string]string) {
			if len(m) == 0 {
				return
			}
			fmt.Fprintf(cmd.OutOrStdout(), "[%s]\n", label)
			keys := make([]string, 0, len(m))
			for k := range m {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s = %s\n", k, m[k])
			}
		}

		printSection("added", result.Added)
		printSection("removed", result.Removed)
		printSection("modified", result.Modified)
		printSection("unchanged", result.Unchanged)
		return nil
	},
}

func init() {
	kvSnapshotDiffCmd.Flags().String("mount", "secret", "KV mount path")
	rootCmd.AddCommand(kvSnapshotDiffCmd)
}
