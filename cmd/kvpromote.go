package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultdiff/internal/vault"
)

func init() {
	var mount string
	var deleteSrc bool

	promoteCmd := &cobra.Command{
		Use:   "kvpromote <src-path> <dst-path>",
		Short: "Promote a KV secret from one path to another",
		Long: `Reads the latest version of a KV v2 secret at <src-path> and
writes it to <dst-path>. Useful for promoting secrets across environments
(e.g. staging -> production).

Optionally delete the source secret after a successful promotion with --delete-src.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			srcPath := args[0]
			dstPath := args[1]

			client, err := vault.NewClient(vault.Config{
				Address:   os.Getenv("VAULT_ADDR"),
				Token:     os.Getenv("VAULT_TOKEN"),
				Namespace: os.Getenv("VAULT_NAMESPACE"),
			})
			if err != nil {
				return fmt.Errorf("kvpromote: init client: %w", err)
			}

			promoter := vault.NewKVPromoter(client, mount, deleteSrc)
			result, err := promoter.Promote(cmd.Context(), srcPath, dstPath)
			if err != nil {
				return fmt.Errorf("kvpromote: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Promoted %q -> %q (%d keys)\n",
				result.SourcePath, result.DestPath, len(result.Data))

			if deleteSrc {
				fmt.Fprintf(cmd.OutOrStdout(), "Source %q deleted.\n", result.SourcePath)
			}
			return nil
		},
	}

	promoteCmd.Flags().StringVar(&mount, "mount", "", "KV v2 mount path (default: secret)")
	promoteCmd.Flags().BoolVar(&deleteSrc, "delete-src", false, "Delete source secret after promotion")

	rootCmd.AddCommand(promoteCmd)
}
