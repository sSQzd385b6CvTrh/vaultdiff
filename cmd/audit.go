package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultdiff/internal/audit"
)

var (
	auditFormat string
	auditOutput string
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Record a structured audit log entry for a diff operation",
	Long: `audit writes a structured log entry (JSON or text) capturing
the environment, secret path, versions compared, and a summary of changes.
Entries can be appended to a file for compliance and change tracking.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		w := os.Stdout
		if auditOutput != "" {
			f, err := os.OpenFile(auditOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o640)
			if err != nil {
				return fmt.Errorf("opening audit output file: %w", err)
			}
			defer f.Close()
			w = f
		}

		logger, err := audit.NewLogger(w, auditFormat)
		if err != nil {
			return err
		}

		entry := audit.Entry{
			Environment: environment,
			Path:        secretPath,
			VersionA:    versionA,
			VersionB:    versionB,
		}

		if err := logger.Write(entry); err != nil {
			return fmt.Errorf("writing audit entry: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(auditCmd)
	auditCmd.Flags().StringVar(&auditFormat, "format", "text", "Audit log format: json or text")
	auditCmd.Flags().StringVar(&auditOutput, "output", "", "File to append audit log entries (default: stdout)")
}
