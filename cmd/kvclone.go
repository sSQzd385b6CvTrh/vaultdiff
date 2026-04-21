package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jonboulle/vaultdiff/internal/vault"
)

var kvcloneCmd = &cobra.Command{
	Use:   "kvclone <src-path> <dst-path>",
	Short: "Clone a KV secret from one path to another",
	Long: `Reads the latest version of a KV v2 secret at src-path and writes
all its key-value pairs to dst-path. Both paths may live under different
mounts by specifying --src-mount and --dst-mount.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		srcPath := args[0]
		dstPath := args[1]

		srcMount, _ := cmd.Flags().GetString("src-mount")
		dstMount, _ := cmd.Flags().GetString("dst-mount")

		rawClient, err := vault.NewClient(vaultAddress, vaultToken, vaultNamespace)
		if err != nil {
			return fmt.Errorf("vault client: %w", err)
		}

		cloner := vault.NewKVCloner(rawClient, srcMount)
		result, err := cloner.Clone(srcPath, dstPath, dstMount)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "Cloned %q → %q\n", srcPath, dstPath)
		if v, ok := result["version"]; ok {
			fmt.Fprintf(os.Stdout, "Destination version: %v\n", v)
		}
		return nil
	},
}

func init() {
	kvcloneCmd.Flags().String("src-mount", "secret", "KV mount for the source path")
	kvcloneCmd.Flags().String("dst-mount", "", "KV mount for the destination path (defaults to src-mount)")
	rootCmd.AddCommand(kvcloneCmd)
}
