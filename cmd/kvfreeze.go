package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"vaultdiff/internal/vault"
)

var kvFreezeCmd = &cobra.Command{
	Use:   "kvfreeze [path]",
	Short: "Freeze or unfreeze a KV secret to prevent accidental modification",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		unfreeze, _ := cmd.Flags().GetBool("unfreeze")
		check, _ := cmd.Flags().GetBool("check")
		mount, _ := cmd.Flags().GetString("mount")

		client, err := newVaultClient()
		if err != nil {
			return fmt.Errorf("vault client error: %w", err)
		}

		freezer := vault.NewKVFreezer(client, mount)

		if check {
			frozen, err := freezer.IsFrozen(path)
			if err != nil {
				return err
			}
			if frozen {
				fmt.Fprintf(os.Stdout, "%s is frozen\n", path)
			} else {
				fmt.Fprintf(os.Stdout, "%s is not frozen\n", path)
			}
			return nil
		}

		var result *vault.KVFreezeResult
		if unfreeze {
			result, err = freezer.Unfreeze(path)
		} else {
			result, err = freezer.Freeze(path)
		}
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, result.Message)
		return nil
	},
}

func init() {
	kvFreezeCmd.Flags().Bool("unfreeze", false, "Unfreeze the secret instead of freezing it")
	kvFreezeCmd.Flags().Bool("check", false, "Check whether the secret is currently frozen")
	kvFreezeCmd.Flags().String("mount", "secret", "KV mount path")
	rootCmd.AddCommand(kvFreezeCmd)
}
