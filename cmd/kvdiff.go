package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jonathanhope/vaultdiff/internal/diff"
	"github.com/jonathanhope/vaultdiff/internal/vault"
)

var kvdiffCmd = &cobra.Command{
	Use:   "kvdiff <src-path> <dst-path>",
	Short: "Diff two KV secret paths in Vault",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		token := os.Getenv("VAULT_TOKEN")
		addr := os.Getenv("VAULT_ADDR")
		mount, _ := cmd.Flags().GetString("mount")
		showUnchanged, _ := cmd.Flags().GetBool("show-unchanged")

		client, err := vault.NewClient(addr, token, "")
		if err != nil {
			return fmt.Errorf("vault client: %w", err)
		}

		differ := vault.NewKVDiffer(client, mount)
		result, err := differ.Diff(args[0], args[1])
		if err != nil {
			return fmt.Errorf("diff failed: %w", err)
		}

		changes := diff.Secrets(result.SourceData, result.TargetData)
		return diff.Write(cmd.OutOrStdout(), changes, showUnchanged)
	},
}

func init() {
	kvdiffCmd.Flags().String("mount", "secret", "KV mount path")
	kvdiffCmd.Flags().Bool("show-unchanged", false, "Include unchanged keys in output")
	rootCmd.AddCommand(kvdiffCmd)
}
