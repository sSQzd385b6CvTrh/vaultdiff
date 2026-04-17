package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/your-org/vaultdiff/internal/vault"
)

var auditDevicesCmd = &cobra.Command{
	Use:   "audit-devices",
	Short: "List enabled audit devices in Vault",
	RunE: func(cmd *cobra.Command, args []string) error {
		addr := os.Getenv("VAULT_ADDR")
		if addr == "" {
			addr = "http://127.0.0.1:8200"
		}
		token := os.Getenv("VAULT_TOKEN")
		ns := os.Getenv("VAULT_NAMESPACE")

		client, err := vault.NewClient(addr, token, ns)
		if err != nil {
			return fmt.Errorf("failed to create vault client: %w", err)
		}

		lister := vault.NewAuditDeviceLister(client)
		devices, err := lister.List()
		if err != nil {
			return fmt.Errorf("failed to list audit devices: %w", err)
		}

		if len(devices) == 0 {
			fmt.Println("No audit devices enabled.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "PATH\tTYPE\tDESCRIPTION")
		for _, d := range devices {
			fmt.Fprintf(w, "%s\t%s\t%s\n", d.Path, d.Type, d.Description)
		}
		return w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(auditDevicesCmd)
}
