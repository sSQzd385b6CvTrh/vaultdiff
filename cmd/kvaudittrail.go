package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/jonnylangefeld/vaultdiff/internal/vault"
)

var kvAuditTrailCmd = &cobra.Command{
	Use:   "kv-audit-trail <path>",
	Short: "Show the full version audit trail for a KV secret",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		address := os.Getenv("VAULT_ADDR")
		if address == "" {
			address = "http://127.0.0.1:8200"
		}
		token := os.Getenv("VAULT_TOKEN")
		mount, _ := cmd.Flags().GetString("mount")

		trailer := vault.NewKVAuditTrailer(address, token, mount)
		entries, err := trailer.GetAuditTrail(args[0])
		if err != nil {
			return fmt.Errorf("audit trail: %w", err)
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "VERSION\tCREATED\tDELETED\tDESTROYED")
		for _, e := range entries {
			deleted := "-"
			if e.DeletedTime != nil {
				deleted = e.DeletedTime.Format("2006-01-02T15:04:05Z")
			}
			destroyed := "false"
			if e.Destroyed {
				destroyed = "true"
			}
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n",
				e.Version,
				e.CreatedTime.Format("2006-01-02T15:04:05Z"),
				deleted,
				destroyed,
			)
		}
		return w.Flush()
	},
}

func init() {
	kvAuditTrailCmd.Flags().String("mount", "", "KV mount path (default: secret)")
	rootCmd.AddCommand(kvAuditTrailCmd)
}
