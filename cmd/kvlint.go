package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jonny-rimek/vaultdiff/internal/vault"
)

var kvlintCmd = &cobra.Command{
	Use:   "kvlint <path>",
	Short: "Lint a KV secret for common issues",
	Long: `Inspect the latest version of a KV v2 secret and report warnings
such as empty values, uppercase keys, or values with unexpected spaces.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		mount, _ := cmd.Flags().GetString("mount")

		client, err := vault.NewClient()
		if err != nil {
			return fmt.Errorf("vault client: %w", err)
		}

		linter := vault.NewKVLinter(client.Client, mount)
		results, err := linter.Lint(path)
		if err != nil {
			return fmt.Errorf("lint: %w", err)
		}

		if len(results) == 0 {
			fmt.Fprintln(os.Stdout, "No issues found.")
			return nil
		}

		fmt.Fprintf(os.Stdout, "Found %d issue(s) in %s:\n", len(results), path)
		for _, r := range results {
			fmt.Fprintf(os.Stdout, "  [v%d] %s: %s\n", r.Version, r.Key, r.Warning)
		}
		return nil
	},
}

func init() {
	kvlintCmd.Flags().String("mount", "secret", "KV v2 mount path")
	rootCmd.AddCommand(kvlintCmd)
}
