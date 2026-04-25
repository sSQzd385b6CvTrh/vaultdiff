package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"vaultdiff/internal/vault"
)

func init() {
	var mount string
	var set []string

	cmd := &cobra.Command{
		Use:   "kv-annotate <path>",
		Short: "Get or set custom annotations on a KV secret path",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			client, err := vault.NewClient()
			if err != nil {
				return fmt.Errorf("vault client: %w", err)
			}
			annotator := vault.NewKVAnnotator(client, mount)

			if len(set) > 0 {
				annotations := map[string]string{}
				for _, pair := range set {
					parts := strings.SplitN(pair, "=", 2)
					if len(parts) != 2 {
						return fmt.Errorf("invalid annotation pair %q: must be key=value", pair)
					}
					annotations[parts[0]] = parts[1]
				}
				if err := annotator.SetAnnotations(path, annotations); err != nil {
					return fmt.Errorf("set annotations: %w", err)
				}
				fmt.Fprintf(os.Stdout, "Annotations updated for %s\n", path)
				return nil
			}

			ann, err := annotator.GetAnnotations(path)
			if err != nil {
				return fmt.Errorf("get annotations: %w", err)
			}
			if len(ann.Annotations) == 0 {
				fmt.Fprintf(os.Stdout, "No annotations found for %s\n", path)
				return nil
			}
			for k, v := range ann.Annotations {
				fmt.Fprintf(os.Stdout, "%s=%s\n", k, v)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&mount, "mount", "", "KV mount path (default: secret)")
	cmd.Flags().StringArrayVar(&set, "set", nil, "Annotation to set as key=value (repeatable)")
	rootCmd.AddCommand(cmd)
}
