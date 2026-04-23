package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/snyk/vaultdiff/internal/vault"
	"github.com/spf13/cobra"
)

var kvChecksumMount string

var kvChecksumCmd = &cobra.Command{
	Use:   "kvchecksum <path> <version>",
	Short: "Compute a SHA-256 checksum of a KV secret version",
	Long: `Fetches the specified version of a KV v2 secret and prints
a deterministic SHA-256 checksum of its data fields.

Useful for verifying secret integrity across environments.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		ver, err := strconv.Atoi(args[1])
		if err != nil || ver < 1 {
			return fmt.Errorf("version must be a positive integer, got %q", args[1])
		}

		client, err := newVaultClient()
		if err != nil {
			return fmt.Errorf("vault client error: %w", err)
		}

		checksummer := vault.NewKVChecksummer(client, kvChecksumMount)
		result, err := checksummer.Checksum(path, ver)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "path:    %s\n", result.Path)
		fmt.Fprintf(os.Stdout, "version: %d\n", result.Version)
		fmt.Fprintf(os.Stdout, "sha256:  %s\n", result.Sum)
		return nil
	},
}

func init() {
	kvChecksumCmd.Flags().StringVar(&kvChecksumMount, "mount", "secret", "KV v2 mount path")
	rootCmd.AddCommand(kvChecksumCmd)
}
