package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"vaultdiff/internal/vault"
)

func init() {
	var mount string

	cmd := &cobra.Command{
		Use:   "kv-undelete <path> <versions>",
		Short: "Restore soft-deleted versions of a KV v2 secret",
		Long: `Restore one or more soft-deleted versions of a KV v2 secret.
Versions should be provided as a comma-separated list, e.g. 1,2,3.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			rawVersions := strings.Split(args[1], ",")

			var versions []int
			for _, raw := range rawVersions {
				v, err := strconv.Atoi(strings.TrimSpace(raw))
				if err != nil {
					return fmt.Errorf("invalid version %q: %w", raw, err)
				}
				versions = append(versions, v)
			}

			client, err := vault.NewClient()
			if err != nil {
				return fmt.Errorf("creating vault client: %w", err)
			}

			undeleter := vault.NewKVUndeleter(client, mount)
			if err := undeleter.Undelete(cmd.Context(), path, versions); err != nil {
				return fmt.Errorf("undelete failed: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Successfully undeleted versions %v of secret %q\n", versions, path)
			return nil
		},
	}

	cmd.Flags().StringVar(&mount, "mount", "", "KV v2 mount path (default: secret)")
	rootCmd.AddCommand(cmd)
}
