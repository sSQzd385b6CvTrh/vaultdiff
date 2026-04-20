package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/drew/vaultdiff/internal/vault"
	"github.com/spf13/cobra"
)

func init() {
	var mount string
	var getMode bool

	kvTagCmd := &cobra.Command{
		Use:   "kv-tag <path> [key=value ...]",
		Short: "Get or set custom metadata tags on a KV v2 secret",
		Example: `  vaultdiff kv-tag secret/myapp env=prod team=platform
  vaultdiff kv-tag --get secret/myapp`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("path argument is required")
			}
			if !getMode && len(args) < 2 {
				return fmt.Errorf("at least one key=value tag is required when setting tags")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := vault.NewClientFromEnv()
			if err != nil {
				return fmt.Errorf("creating vault client: %w", err)
			}

			tagger := vault.NewKVTagger(client, mount)
			path := args[0]

			if getMode {
				tags, err := tagger.GetTags(path)
				if err != nil {
					return fmt.Errorf("getting tags: %w", err)
				}
				if len(tags) == 0 {
					fmt.Fprintln(os.Stdout, "(no tags set)")
					return nil
				}
				for k, v := range tags {
					fmt.Fprintf(os.Stdout, "%s=%s\n", k, v)
				}
				return nil
			}

			tags := make(map[string]string, len(args)-1)
			for _, pair := range args[1:] {
				parts := strings.SplitN(pair, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid tag format %q: expected key=value", pair)
				}
				tags[parts[0]] = parts[1]
			}

			if err := tagger.SetTags(path, tags); err != nil {
				return fmt.Errorf("setting tags: %w", err)
			}
			fmt.Fprintf(os.Stdout, "Tags updated for %s\n", path)
			return nil
		},
	}

	kvTagCmd.Flags().StringVar(&mount, "mount", "", "KV v2 mount path (default: secret)")
	kvTagCmd.Flags().BoolVar(&getMode, "get", false, "Read tags instead of writing")

	rootCmd.AddCommand(kvTagCmd)
}
