package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultdiff/internal/vault"
)

var kvHistoryCmd = &cobra.Command{
	Use:   "kv-history <path>",
	Short: "Show version history for a KV v2 secret",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr := os.Getenv("VAULT_ADDR")
		token := os.Getenv("VAULT_TOKEN")
		ns := os.Getenv("VAULT_NAMESPACE")
		mount, _ := cmd.Flags().GetString("mount")

		client, err := vault.NewClient(addr, token, ns)
		if err != nil {
			return fmt.Errorf("vault client: %w", err)
		}

		reader := vault.NewKVHistoryReader(client, mount)
		history, err := reader.GetHistory(context.Background(), args[0])
		if err != nil {
			return fmt.Errorf("get history: %w", err)
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "VERSION\tCREATED\tDESTROYED")
		for _, v := range history {
			destroyed := "-"
			if v.Destroyed {
				destroyed = "yes"
			}
			fmt.Fprintf(w, "%d\t%s\t%s\n", v.Version, v.CreatedTime.Format("2006-01-02 15:04:05"), destroyed)
		}
		return w.Flush()
	},
}

func init() {
	kvHistoryCmd.Flags().String("mount", "secret", "KV v2 mount path")
	rootCmd.AddCommand(kvHistoryCmd)
}
