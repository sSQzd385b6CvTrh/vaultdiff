package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultdiff/internal/vault"
)

func init() {
	var mount string
	var interval int

	cmd := &cobra.Command{
		Use:   "kv-watch-keys <path> [path...]",
		Short: "Watch one or more KV paths for version changes",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newVaultClient()
			if err != nil {
				return fmt.Errorf("vault client: %w", err)
			}

			watcher := vault.NewKVKeyWatcher(client, mount, time.Duration(interval)*time.Second)

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			fmt.Fprintf(cmd.OutOrStdout(), "Watching %d path(s) every %ds. Press Ctrl+C to stop.\n", len(args), interval)

			ch, err := watcher.WatchKeys(ctx, args)
			if err != nil {
				return fmt.Errorf("watch: %w", err)
			}

			for result := range ch {
				fmt.Fprintf(cmd.OutOrStdout(), "[CHANGED] path=%s version=%d\n", result.Path, result.Version)
				for k, v := range result.Data {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s = %v\n", k, v)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&mount, "mount", "secret", "KV mount path")
	cmd.Flags().IntVar(&interval, "interval", 10, "Poll interval in seconds")

	rootCmd.AddCommand(cmd)
}
