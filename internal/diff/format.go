package diff

import (
	"fmt"
	"io"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorGray   = "\033[90m"
)

// FormatOptions controls the output format of a diff result.
type FormatOptions struct {
	Color     bool
	ShowEqual bool
}

// Write renders the diff Result to the provided writer.
func Write(w io.Writer, result Result, opts FormatOptions) {
	fmt.Fprintf(w, "Path: %s\n", result.Path)
	fmt.Fprintln(w, strings.Repeat("-", 40))

	for _, c := range result.Changes {
		switch c.Type {
		case Added:
			line := fmt.Sprintf("+ %s = %s", c.Key, c.NewValue)
			if opts.Color {
				line = colorGreen + line + colorReset
			}
			fmt.Fprintln(w, line)
		case Removed:
			line := fmt.Sprintf("- %s = %s", c.Key, c.OldValue)
			if opts.Color {
				line = colorRed + line + colorReset
			}
			fmt.Fprintln(w, line)
		case Modified:
			if opts.Color {
				fmt.Fprintf(w, "%s~ %s%s\n", colorYellow, c.Key, colorReset)
				fmt.Fprintf(w, "%s  - %s%s\n", colorRed, c.OldValue, colorReset)
				fmt.Fprintf(w, "%s  + %s%s\n", colorGreen, c.NewValue, colorReset)
			} else {
				fmt.Fprintf(w, "~ %s\n  - %s\n  + %s\n", c.Key, c.OldValue, c.NewValue)
			}
		case Unchanged:
			if opts.ShowEqual {
				line := fmt.Sprintf("  %s", c.Key)
				if opts.Color {
					line = colorGray + line + colorReset
				}
				fmt.Fprintln(w, line)
			}
		}
	}
}
