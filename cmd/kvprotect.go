package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jonny-rimek/vaultdiff/internal/vault"
)

func init() {
	var mount string
	var check bool

	kvprotectCmd := &cobra.Command{
		Use:   "kvprotect <path>",
		Short: "Protect or unprotect a KV secret from accidental writes",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			token := os.Getenv("VAULT_TOKEN")
			address := os.Getenv("VAULT_ADDR")
			if address == "" {
				address = "https://127.0.0.1:8200"
			}
			client, err := vault.NewClient(address, token, "")
			if err != nil {
				return fmt.Errorf("vault client: %w", err)
			}
			p := vault.NewKVProtector(client, mount)

			if check {
				ok, err := p.IsProtected(path)
				if err != nil {
					return err
				}
				if ok {
					fmt.Fprintf(cmd.OutOrStdout(), "protected: true\n")
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "protected: false\n")
				}
				return nil
			}

			unprotect, _ := cmd.Flags().GetBool("unprotect")
			if unprotect {
				if err := p.Unprotect(path); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "unprotected: %s\n", path)
				return nil
			}
			if err := p.Protect(path); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "protected: %s\n", path)
			return nil
		},
	}

	kvprotectCmd.Flags().StringVar(&mount, "mount", "secret", "KV mount path")
	kvprotectCmd.Flags().BoolVar(&check, "check", false, "Check if path is protected")
	kvprotectCmd.Flags().Bool("unprotect", false, "Remove protection from path")
	rootCmd.AddCommand(kvprotectCmd)
}
