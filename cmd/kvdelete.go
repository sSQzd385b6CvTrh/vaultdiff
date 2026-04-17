package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/your-org/vaultdiff/internal/vault"
)

var kvDeleteCmd = &cobra.Command{
	Use:   "kv-delete [path] [versions...]",
	Short: "Soft-delete or destroy KV secret versions",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		versions, err := parseVersionList(args[1:])
		if err != nil {
			return err
		}
		client, err := vault.NewClient(vaultAddr, vaultToken, vaultNamespace)
		if err != nil {
			return fmt.Errorf("vault client error: %w", err)
		}
		deleter := vault.NewKVDeleter(client, kvMount)
		destroy, _ := cmd.Flags().GetBool("destroy")
		if destroy {
			if err := deleter.Destroy(path, versions); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Destroyed versions %v of %s\n", versions, path)
		} else {
			if err := deleter.Delete(path, versions); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Deleted versions %v of %s\n", versions, path)
		}
		return nil
	},
}

func parseVersionList(raw []string) ([]int, error) {
	var versions []int
	for _, s := range raw {
		for _, part := range strings.Split(s, ",") {
			v, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil {
				return nil, fmt.Errorf("invalid version %q: %w", part, err)
			}
			versions = append(versions, v)
		}
	}
	return versions, nil
}

func init() {
	kvDeleteCmd.Flags().Bool("destroy", false, "Permanently destroy versions instead of soft-delete")
	kvDeleteCmd.Flags().StringVar(&kvMount, "mount", "secret", "KV mount path")
	rootCmd.AddCommand(kvDeleteCmd)
}
