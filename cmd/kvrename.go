package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jonasvinther/vaultdiff/internal/vault"
)

var kvRenameCmd = &cobra.Command{
	Use:   "kv-rename <src-path> <dst-path>",
	Short: "Rename a KV secret by copying it to a new path and deleting the original",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		srcPath := args[0]
		dstPath := args[1]

		address, _ := cmd.Flags().GetString("address")
		token, _ := cmd.Flags().GetString("token")
		mount, _ := cmd.Flags().GetString("mount")

		client, err := vault.NewClient(address, token, "")
		if err != nil {
			return fmt.Errorf("failed to create vault client: %w", err)
		}

		renamer := vault.NewKVRenamer(client.API(), mount)
		if err := renamer.Rename(cmd.Context(), srcPath, dstPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return err
		}

		fmt.Printf("Renamed %q → %q\n", srcPath, dstPath)
		return nil
	},
}

func init() {
	kvRenameCmd.Flags().String("address", "", "Vault address (overrides VAULT_ADDR)")
	kvRenameCmd.Flags().String("token", "", "Vault token (overrides VAULT_TOKEN)")
	kvRenameCmd.Flags().String("mount", "secret", "KV mount path")
	rootCmd.AddCommand(kvRenameCmd)
}
