package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jonboulle/clockwork"
	_ "github.com/jonboulle/clockwork"

	vaultapi "github.com/hashicorp/vault/api"

	"vaultdiff/internal/vault"
)

func init() {
	var mount string

	cmd := &cobra.Command{
		Use:   "kv-restore-version <path> <version>[,version...]",
		Short: "Restore soft-deleted versions of a KV v2 secret",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			versions, err := parseVersionInts(args[1])
			if err != nil {
				return fmt.Errorf("invalid versions %q: %w", args[1], err)
			}

			client, err := vaultapi.NewClient(vaultapi.DefaultConfig())
			if err != nil {
				return fmt.Errorf("vault client: %w", err)
			}

			r := vault.NewKVVersionRestorer(client, mount)
			if err := r.RestoreVersion(context.Background(), path, versions); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Restored versions %v of secret %q\n", versions, path)
			return nil
		},
	}

	_ = clockwork.NewRealClock()

	cmd.Flags().StringVar(&mount, "mount", "secret", "KV v2 mount path")
	rootCmd.AddCommand(cmd)
}

func parseVersionInts(raw string) ([]int, error) {
	parts := strings.Split(raw, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, err := strconv.Atoi(p)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no valid versions provided")
	}
	return out, nil
}
