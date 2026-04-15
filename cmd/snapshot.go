package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultdiff/internal/vault"
)

// snapshotCmd groups snapshot-related subcommands (capture and restore).
var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Capture or restore secret snapshots",
	Long: `Manage point-in-time snapshots of Vault KV secrets.

Snapshots capture the current state of all keys under a given path and
write them to a JSON file. They can later be restored to roll back bulk
changes or replicate secrets across environments.`,
}

// snapshotCaptureCmd captures a snapshot of all secrets under a path.
var snapshotCaptureCmd = &cobra.Command{
	Use:   "capture <path>",
	Short: "Capture a snapshot of secrets at the given path",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]

		address, _ := cmd.Flags().GetString("address")
		token, _ := cmd.Flags().GetString("token")
		mount, _ := cmd.Flags().GetString("mount")
		output, _ := cmd.Flags().GetString("output")

		client, err := vault.NewClient(address, token)
		if err != nil {
			return fmt.Errorf("creating vault client: %w", err)
		}

		snapshotter := vault.NewSnapshotter(client, mount)
		snap, err := snapshotter.Capture(cmd.Context(), path)
		if err != nil {
			return fmt.Errorf("capturing snapshot: %w", err)
		}

		// Default output filename includes timestamp for uniqueness.
		if output == "" {
			output = fmt.Sprintf("snapshot-%s.json", time.Now().UTC().Format("20060102-150405"))
		}

		f, err := os.Create(output)
		if err != nil {
			return fmt.Errorf("creating output file %q: %w", output, err)
		}
		defer f.Close()

		if err := snap.Write(f); err != nil {
			return fmt.Errorf("writing snapshot: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Snapshot written to %s (%d keys)\n", output, snap.Count())
		return nil
	},
}

// snapshotRestoreCmd restores secrets from a previously captured snapshot.
var snapshotRestoreCmd = &cobra.Command{
	Use:   "restore <snapshot-file>",
	Short: "Restore secrets from a snapshot file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		snapshotFile := args[0]

		address, _ := cmd.Flags().GetString("address")
		token, _ := cmd.Flags().GetString("token")
		mount, _ := cmd.Flags().GetString("mount")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		client, err := vault.NewClient(address, token)
		if err != nil {
			return fmt.Errorf("creating vault client: %w", err)
		}

		f, err := os.Open(snapshotFile)
		if err != nil {
			return fmt.Errorf("opening snapshot file %q: %w", snapshotFile, err)
		}
		defer f.Close()

		restorer := vault.NewRestorer(client, mount)
		count, err := restorer.RestoreSnapshot(cmd.Context(), f, dryRun)
		if err != nil {
			return fmt.Errorf("restoring snapshot: %w", err)
		}

		if dryRun {
			fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] Would restore %d keys from %s\n", count, snapshotFile)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "Restored %d keys from %s\n", count, snapshotFile)
		}
		return nil
	},
}

func init() {
	// Capture flags
	snapshotCaptureCmd.Flags().String("address", "", "Vault address (overrides VAULT_ADDR)")
	snapshotCaptureCmd.Flags().String("token", "", "Vault token (overrides VAULT_TOKEN)")
	snapshotCaptureCmd.Flags().String("mount", "secret", "KV mount path")
	snapshotCaptureCmd.Flags().StringP("output", "o", "", "Output file path (default: snapshot-<timestamp>.json)")

	// Restore flags
	snapshotRestoreCmd.Flags().String("address", "", "Vault address (overrides VAULT_ADDR)")
	snapshotRestoreCmd.Flags().String("token", "", "Vault token (overrides VAULT_TOKEN)")
	snapshotRestoreCmd.Flags().String("mount", "secret", "KV mount path")
	snapshotRestoreCmd.Flags().Bool("dry-run", false, "Preview restore without writing to Vault")

	snapshotCmd.AddCommand(snapshotCaptureCmd)
	snapshotCmd.AddCommand(snapshotRestoreCmd)
	rootCmd.AddCommand(snapshotCmd)
}
