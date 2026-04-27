package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultdiff/internal/vault"
)

var kvBulkCopyCmd = &cobra.Command{
	Use:   "kvbulkcopy [src:dst...]",
	Short: "Bulk copy KV secrets from source paths to destination paths",
	Long: `Copy multiple KV v2 secrets in one command.

Each argument must be a colon-separated source:destination pair, e.g.:

  vaultdiff kvbulkcopy app/prod:app/staging infra/prod:infra/staging`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr := os.Getenv("VAULT_ADDR")
		token := os.Getenv("VAULT_TOKEN")
		ns := os.Getenv("VAULT_NAMESPACE")
		mount, _ := cmd.Flags().GetString("mount")

		pairs, err := parseCopyPairs(args)
		if err != nil {
			return err
		}

		client, err := vault.NewClient(addr, token, ns)
		if err != nil {
			return fmt.Errorf("vault client: %w", err)
		}

		copier := vault.NewKVBulkCopier(client, mount)
		results := copier.Copy(cmd.Context(), pairs)

		exitCode := 0
		for _, r := range results {
			if r.Err != nil {
				fmt.Fprintf(os.Stderr, "ERROR  %s → %s: %v\n", r.Source, r.Destination, r.Err)
				exitCode = 1
			} else {
				fmt.Printf("OK     %s → %s\n", r.Source, r.Destination)
			}
		}
		if exitCode != 0 {
			os.Exit(exitCode)
		}
		return nil
	},
}

func parseCopyPairs(args []string) ([]vault.CopyPair, error) {
	pairs := make([]vault.CopyPair, 0, len(args))
	for _, arg := range args {
		parts := strings.SplitN(arg, ":", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("invalid pair %q: expected src:dst format", arg)
		}
		pairs = append(pairs, vault.CopyPair{Source: parts[0], Destination: parts[1]})
	}
	return pairs, nil
}

func init() {
	kvBulkCopyCmd.Flags().String("mount", "secret", "KV v2 mount path")
	rootCmd.AddCommand(kvBulkCopyCmd)
}
