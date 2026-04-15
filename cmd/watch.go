package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"vaultdiff/internal/vault"
)

var (
	watchMount    string
	watchInterval int
)

var watchCmd = &cobra.Command{
se:   "watch <path>",
	Short: "PollV secret path and print version changes",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]

		client, err := vault.NewClient(vaultAddr, vaultNamespace)
		if err != nil {
			return fmt.Errorf("vault client: %w", err)
		}

		interval := time.Duration(watchInterval) * time.Second
		w := vault.NewWatcher(client, watchMount, interval)

		ch := make(chan vault.WatchEvent, 8)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sig
			fmt.Fprintln(os.Stderr, "\nstopping watcher")
			cancel()
		}()

		fmt.Fprintf(os.Stdout, "watching %s (interval %s)...\n", path, interval)
		go w.Watch(ctx, path, ch)

		for {
			select {
			case <-ctx.Done():
				return nil
			case event := <-ch:
				if event.Err != nil {
					fmt.Fprintf(os.Stderr, "error: %v\n", event.Err)
					continue
				}
				fmt.Fprintf(os.Stdout, "[%s] %s version=%d keys=%d\n",
					time.Now().Format(time.RFC3339), event.Path, event.Version, len(event.Data))
			}
		}
	},
}

func init() {
	watchCmd.Flags().StringVar(&watchMount, "mount", "secret", "KV mount path")
	watchCmd.Flags().IntVar(&watchInterval, "interval", 30, "poll interval in seconds")
	rootCmd.AddCommand(watchCmd)
}
