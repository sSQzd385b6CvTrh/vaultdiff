package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"vaultdiff/internal/vault"
)

var kvEnvDiffCmd = &cobra.Command{
	Use:   "kvenvdiff <mount> <path>",
	Short: "Diff a KV secret between two Vault environments",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		mount := args[0]
		path := args[1]

		leftAddr, _ := cmd.Flags().GetString("left-addr")
		rightAddr, _ := cmd.Flags().GetString("right-addr")
		leftToken, _ := cmd.Flags().GetString("left-token")
		rightToken, _ := cmd.Flags().GetString("right-token")

		lc, err := vault.NewClient(leftAddr, leftToken)
		if err != nil {
			return fmt.Errorf("left client: %w", err)
		}
		rc, err := vault.NewClient(rightAddr, rightToken)
		if err != nil {
			return fmt.Errorf("right client: %w", err)
		}

		d := vault.NewKVEnvDiffer(
			vault.NewKVGetter(lc, mount),
			vault.NewKVGetter(rc, mount),
		)

		res, err := d.Compare(mount, path)
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Status: %s\n\n", res.Status)
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "KEY\tLEFT\tRIGHT")
		for _, k := range res.SortedKeys() {
			fmt.Fprintf(w, "%s\t%s\t%s\n", k, res.Left[k], res.Right[k])
		}
		w.Flush()
		return nil
	},
}

func init() {
	kvEnvDiffCmd.Flags().String("left-addr", os.Getenv("VAULT_ADDR"), "Address of the left Vault")
	kvEnvDiffCmd.Flags().String("right-addr", "", "Address of the right Vault")
	kvEnvDiffCmd.Flags().String("left-token", os.Getenv("VAULT_TOKEN"), "Token for the left Vault")
	kvEnvDiffCmd.Flags().String("right-token", "", "Token for the right Vault")
	_ = kvEnvDiffCmd.MarkFlagRequired("right-addr")
	_ = kvEnvDiffCmd.MarkFlagRequired("right-token")
	rootCmd.AddCommand(kvEnvDiffCmd)
}
