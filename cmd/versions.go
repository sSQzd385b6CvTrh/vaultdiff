package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultdiff/internal/vault"
)

var (
	versionMount string
	versionPath  string
)

var versionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "List available versions for a Vault secret path",
	Long:  `Queries KV v2 metadata to list all versions of a secret, showing creation time and status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if versionPath == "" {
			return fmt.Errorf("--path is required")
		}

		client, err := vault.NewClient(vaultAddr, vaultNamespace)
		if err != nil {
			return fmt.Errorf("creating vault client: %w", err)
		}

		fetcher := vault.NewFetcher(client)
		lister := vault.NewVersionLister(fetcher)

		metas, err := lister.ListVersions(cmd.Context(), versionMount, versionPath)
		if err != nil {
			return fmt.Errorf("listing versions: %w", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "VERSION\tCREATED\tDELETED\tDESTROYED")
		for _, m := range metas {
			deletionTime := m.DeletionTime
			if deletionTime == "" {
				deletionTime = "-"
			}
			fmt.Fprintf(w, "%d\t%s\t%s\t%v\n",
				m.Version, m.CreatedTime, deletionTime, m.Destroyed)
		}
		return w.Flush()
	},
}

func init() {
	versionsCmd.Flags().StringVar(&versionMount, "mount", "secret", "KV v2 mount path")
	versionsCmd.Flags().StringVar(&versionPath, "path", "", "Secret path to list versions for (required)")
	rootCmd.AddCommand(versionsCmd)
}
