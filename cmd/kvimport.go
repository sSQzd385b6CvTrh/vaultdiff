package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultdiff/internal/vault"
)

var kvImportCmd = &cobra.Command{
	Use:   "kv-import [file]",
	Short: "Import secrets from a JSON file into Vault KV v2",
	Long: `Reads a JSON file mapping secret paths to key/value pairs and
writes each secret into Vault KV v2 under the specified mount.

Example JSON format:
  {
    "app/database": {"password": "s3cr3t"},
    "app/api":      {"key": "abc123"}
  }`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		mount, _ := cmd.Flags().GetString("mount")

		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file %q: %w", filePath, err)
		}

		var secrets map[string]map[string]string
		if err := json.Unmarshal(data, &secrets); err != nil {
			return fmt.Errorf("invalid JSON in %q: %w", filePath, err)
		}

		if len(secrets) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "no secrets found in file")
			return nil
		}

		client, err := vault.NewClient(vault.Config{
			Address: os.Getenv("VAULT_ADDR"),
			Token:   os.Getenv("VAULT_TOKEN"),
		})
		if err != nil {
			return fmt.Errorf("vault client error: %w", err)
		}

		importer := vault.NewKVImporter(client, mount)
		results := importer.Import(context.Background(), secrets)

		failed := 0
		for _, r := range results {
			if r.Success {
				fmt.Fprintf(cmd.OutOrStdout(), "imported: %s\n", r.Path)
			} else {
				fmt.Fprintf(cmd.ErrOrStderr(), "error:    %s — %v\n", r.Path, r.Error)
				failed++
			}
		}

		if failed > 0 {
			return fmt.Errorf("%d secret(s) failed to import", failed)
		}
		return nil
	},
}

func init() {
	kvImportCmd.Flags().String("mount", "secret", "KV v2 mount path")
	rootCmd.AddCommand(kvImportCmd)
}
