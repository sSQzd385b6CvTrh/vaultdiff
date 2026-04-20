package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultdiff/internal/vault"
)

var kvMoveCmd = &cobra.Command{
	Use:   "kv-move <src-path> <dst-path>",
	Short: "Move a KV secret from one path to another",
	Long: `Copy the secret at <src-path> to <dst-path>, then permanently
destroy all versions at <src-path>.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		srcPath := args[0]
		dstPath := args[1]

		address := os.Getenv("VAULT_ADDR")
		if address == "" {
			address = "http://127.0.0.1:8200"
		}
		token := os.Getenv("VAULT_TOKEN")

		mount, _ := cmd.Flags().GetString("mount")

		client := vault.NewClient(address, token)
		mover := vault.NewKVMover(client, mount)

		if err := mover.Move(cmd.Context(), srcPath, dstPath); err != nil {
			return fmt.Errorf("kv-move failed: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Moved %q -> %q\n", srcPath, dstPath)
		return nil
	},
}

func init() {
	kvMoveCmd.Flags().String("mount", "secret", "KV mount path")
	rootCmd.AddCommand(kvMoveCmd)
}
