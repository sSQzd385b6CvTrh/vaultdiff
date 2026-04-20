package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultdiff/internal/vault"
)

var kvExportMount string
var kvExportOutput string

var kvExportCmd = &cobra.Command{
	Use:   "kv-export <path>",
	Short: "Export all secrets under a KV path to JSON",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := vault.NewClient(vaultAddr, vaultToken, vaultNamespace)
		if err != nil {
			return fmt.Errorf("vault client: %w", err)
		}

		exporter := vault.NewKVExporter(client, kvExportMount)
		result, err := exporter.Export(args[0])
		if err != nil {
			return fmt.Errorf("export failed: %w", err)
		}

		out := os.Stdout
		if kvExportOutput != "" {
			f, err := os.Create(kvExportOutput)
			if err != nil {
				return fmt.Errorf("open output file: %w", err)
			}
			defer f.Close()
			out = f
		}

		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	},
}

func init() {
	kvExportCmd.Flags().StringVar(&kvExportMount, "mount", "secret", "KV mount path")
	kvExportCmd.Flags().StringVarP(&kvExportOutput, "output", "o", "", "Write output to file instead of stdout")
	rootCmd.AddCommand(kvExportCmd)
}
