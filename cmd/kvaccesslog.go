package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/jonathanhope/vaultdiff/internal/vault"
)

func init() {
	var mount string

	cmd := &cobra.Command{
		Use:   "kv-access-log <path>",
		Short: "Show version access log for a KV secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := vault.NewClient()
			if err != nil {
				return fmt.Errorf("vault client error: %w", err)
			}

			logger := vault.NewKVAccessLogger(client, mount)
			entries, err := logger.GetAccessLog(args[0])
			if err != nil {
				return err
			}

			if len(entries) == 0 {
				fmt.Println("no versions found")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "VERSION\tCREATED\tDELETION TIME\tDESTROYED")
			for _, e := range entries {
				deletion := e.DeletionTime
				if deletion == "" {
					deletion = "-"
				}
				destroyed := "no"
				if e.Destroyed {
					destroyed = "yes"
				}
				fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", e.Version, e.CreatedTime, deletion, destroyed)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&mount, "mount", "", "KV mount path (default: secret)")
	rootCmd.AddCommand(cmd)
}
