package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thrasher-corp/vaultdiff/internal/vault"
	"github.com/thrasher-corp/vaultdiff/internal/vault/client"
)

func init() {
	var mount string
	var requiredKeys []string
	var optionalKeys []string

	cmd := &cobra.Command{
		Use:   "kv-schema <path>",
		Short: "Validate a KV secret against an expected schema",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			addr := os.Getenv("VAULT_ADDR")
			token := os.Getenv("VAULT_TOKEN")

			c, err := client.NewClient(addr, token, "")
			if err != nil {
				return fmt.Errorf("vault client: %w", err)
			}

			var fields []vault.SchemaField
			for _, k := range requiredKeys {
				fields = append(fields, vault.SchemaField{Key: strings.TrimSpace(k), Required: true})
			}
			for _, k := range optionalKeys {
				fields = append(fields, vault.SchemaField{Key: strings.TrimSpace(k), Required: false})
			}

			v := vault.NewKVSchemaValidator(c, mount)
			result, err := v.Validate(path, fields)
			if err != nil {
				return err
			}

			if result.Valid {
				fmt.Fprintf(cmd.OutOrStdout(), "✓ %s schema is valid\n", path)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "✗ %s schema is invalid\n", path)
				fmt.Fprintf(cmd.OutOrStdout(), "  missing: %s\n", strings.Join(result.Missing, ", "))
			}
			if len(result.Extra) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "  extra:   %s\n", strings.Join(result.Extra, ", "))
			}
			if !result.Valid {
				return fmt.Errorf("schema validation failed")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&mount, "mount", "secret", "KV mount path")
	cmd.Flags().StringSliceVar(&requiredKeys, "require", nil, "Required keys (comma-separated)")
	cmd.Flags().StringSliceVar(&optionalKeys, "optional", nil, "Optional keys (comma-separated)")

	rootCmd.AddCommand(cmd)
}
