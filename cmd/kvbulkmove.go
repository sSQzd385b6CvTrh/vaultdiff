package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"vaultdiff/internal/vault"
)

func init() {
	var mount string

	cmd := &cobra.Command{
		Use:   "kvbulkmove [src1=dst1] [src2=dst2] ...",
		Short: "Move multiple KV secrets in bulk (copy then delete source)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			addr := mustEnv("VAULT_ADDR")
			token := mustEnv("VAULT_TOKEN")
			ns, _ := cmd.Flags().GetString("namespace")

			client, err := vault.NewClient(addr, token, ns)
			if err != nil {
				return fmt.Errorf("vault client: %w", err)
			}

			pairs, err := parseBulkMovePairs(args)
			if err != nil {
				return err
			}

			mover := vault.NewKVBulkMover(client, mount)
			results := mover.Move(pairs)

			ok := true
			for _, r := range results {
				if r.Err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "ERROR %s → %s: %v\n", r.Source, r.Dest, r.Err)
					ok = false
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "moved %s → %s\n", r.Source, r.Dest)
				}
			}
			if !ok {
				return fmt.Errorf("one or more moves failed")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&mount, "mount", "secret", "KV mount path")
	cmd.Flags().String("namespace", "", "Vault namespace")
	rootCmd.AddCommand(cmd)
}

// parseBulkMovePairs parses "src=dst" argument pairs.
func parseBulkMovePairs(args []string) ([]vault.MovePair, error) {
	pairs := make([]vault.MovePair, 0, len(args))
	for _, a := range args {
		parts := strings.SplitN(a, "=", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("invalid pair %q: expected src=dst", a)
		}
		pairs = append(pairs, vault.MovePair{Source: parts[0], Dest: parts[1]})
	}
	return pairs, nil
}
