package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"vaultdiff/internal/vault"
)

var (
	policyName    string
	showPolicyAll bool
)

var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "List or inspect Vault ACL policies",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		c, err := vault.NewClient(vaultAddr, vaultToken, vaultNamespace)
		if err != nil {
			return fmt.Errorf("creating vault client: %w", err)
		}
		lister := vault.NewPolicyLister(c)

		if policyName != "" {
			entry, err := lister.GetPolicy(ctx, policyName)
			if err != nil {
				return err
			}
			fmt.Printf("Policy: %s\n\n%s\n", entry.Name, entry.Rules)
			return nil
		}

		names, err := lister.ListPolicies(ctx)
		if err != nil {
			return err
		}
		if len(names) == 0 {
			fmt.Println("No policies found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME")
		for _, n := range names {
			fmt.Fprintln(w, n)
		}
		return w.Flush()
	},
}

func init() {
	policyCmd.Flags().StringVar(&policyName, "name", "", "Name of a specific policy to inspect")
	policyCmd.Flags().BoolVar(&showPolicyAll, "all", false, "List all policies (default behaviour)")
	rootCmd.AddCommand(policyCmd)
}
