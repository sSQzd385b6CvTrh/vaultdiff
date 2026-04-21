package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultdiff/internal/vault"
)

func init() {
	var mount string
	var owner string
	var ttl time.Duration

	lockCmd := &cobra.Command{
		Use:   "kvlock <lock|unlock> <path>",
		Short: "Set or remove an advisory lock on a KV secret path",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			action := args[0]
			path := args[1]

			address := os.Getenv("VAULT_ADDR")
			token := os.Getenv("VAULT_TOKEN")
			namespace := os.Getenv("VAULT_NAMESPACE")

			if address == "" {
				address = "http://127.0.0.1:8200"
			}

			client, err := vault.NewClient(address, token, namespace)
			if err != nil {
				return fmt.Errorf("failed to create vault client: %w", err)
			}

			locker := vault.NewKVLocker(client, mount)

			switch action {
			case "lock":
				if owner == "" {
					return fmt.Errorf("--owner is required for lock")
				}
				if err := locker.Lock(path, owner, ttl); err != nil {
					return fmt.Errorf("lock failed: %w", err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Locked %s (owner: %s, ttl: %s)\n", path, owner, ttl)
			case "unlock":
				if err := locker.Unlock(path); err != nil {
					return fmt.Errorf("unlock failed: %w", err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Unlocked %s\n", path)
			default:
				return fmt.Errorf("unknown action %q: use 'lock' or 'unlock'", action)
			}
			return nil
		},
	}

	lockCmd.Flags().StringVar(&mount, "mount", "", "KV mount path (default: secret)")
	lockCmd.Flags().StringVar(&owner, "owner", "", "Owner identifier for the lock")
	lockCmd.Flags().DurationVar(&ttl, "ttl", 15*time.Minute, "Lock TTL duration")

	rootCmd.AddCommand(lockCmd)
}
