package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultdiff/internal/vault"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check the health status of a Vault instance",
	Long:  `Queries the Vault /v1/sys/health endpoint and displays the cluster status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		address, _ := cmd.Flags().GetString("address")
		if address == "" {
			address = os.Getenv("VAULT_ADDR")
		}
		if address == "" {
			address = "http://127.0.0.1:8200"
		}

		checker := vault.NewHealthChecker(address)
		status, err := checker.Check(context.Background())
		if err != nil {
			return fmt.Errorf("health check failed: %w", err)
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "FIELD\tVALUE")
		fmt.Fprintf(w, "Initialized\t%v\n", status.Initialized)
		fmt.Fprintf(w, "Sealed\t%v\n", status.Sealed)
		fmt.Fprintf(w, "Standby\t%v\n", status.Standby)
		fmt.Fprintf(w, "Version\t%s\n", status.Version)
		fmt.Fprintf(w, "Cluster Name\t%s\n", status.ClusterName)
		fmt.Fprintf(w, "Cluster ID\t%s\n", status.ClusterID)
		w.Flush()

		if status.Sealed {
			fmt.Fprintln(cmd.OutOrStdout(), "\nWARNING: Vault is sealed.")
		}

		return nil
	},
}

func init() {
	healthCmd.Flags().String("address", "", "Vault server address (overrides VAULT_ADDR)")
	rootCmd.AddCommand(healthCmd)
}
