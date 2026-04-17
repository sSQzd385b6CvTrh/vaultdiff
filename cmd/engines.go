package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultdiff/internal/vault"
)

var enginesCmd = &cobra.Command{
	Use:   "engines",
	Short: "List all mounted secret engines in Vault",
	RunE: func(cmd *cobra.Command, args []string) error {
		addr := os.Getenv("VAULT_ADDR")
		if addr == "" {
			addr = "http://127.0.0.1:8200"
		}
		token := os.Getenv("VAULT_TOKEN")
		ns := os.Getenv("VAULT_NAMESPACE")

		c, err := vault.NewClient(addr, token, ns)
		if err != nil {
			return fmt.Errorf("failed to create vault client: %w", err)
		}

		lister := vault.NewSecretEngineLister(c)
		result, err := lister.ListEngines()
		if err != nil {
			return fmt.Errorf("failed to list engines: %w", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "PATH\tTYPE\tDESCRIPTION\tACCESSOR")
		for _, e := range result.Engines {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.Path, e.Type, e.Description, e.Accessor)
		}
		return w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(enginesCmd)
}
