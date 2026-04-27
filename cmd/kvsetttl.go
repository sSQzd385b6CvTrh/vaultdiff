package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"vaultdiff/internal/vault"
)

func init() {
	var mount string
	var ttlStr string

	cmd := &cobra.Command{
		Use:   "kv-set-ttl <path>",
		Short: "Set or clear the deletion TTL on a KV v2 secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			client, err := vault.NewClient()
			if err != nil {
				return fmt.Errorf("vault client: %w", err)
			}

			var ttl time.Duration
			if ttlStr != "" {
				ttl, err = time.ParseDuration(ttlStr)
				if err != nil {
					return fmt.Errorf("invalid TTL %q: %w", ttlStr, err)
				}
			}

			setter := vault.NewKVTTLSetter(client, mount)
			if err := setter.SetTTL(context.Background(), path, ttl); err != nil {
				return err
			}

			if ttl > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "TTL set to %s on %s\n", ttl, path)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "TTL cleared on %s\n", path)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&mount, "mount", "secret", "KV v2 mount path")
	cmd.Flags().StringVar(&ttlStr, "ttl", "", "TTL duration (e.g. 24h); omit to clear")

	rootCmd.AddCommand(cmd)
}
