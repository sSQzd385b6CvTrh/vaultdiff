package audit

import (
	"fmt"
	"io"
	"sort"

	"github.com/your-org/vaultdiff/internal/diff"
)

// ChangeSummary converts a slice of diff.Change into a map suitable for
// inclusion in an audit Entry.
func ChangeSummary(changes []diff.Change) map[string]string {
	summary := make(map[string]string, len(changes))
	for _, c := range changes {
		summary[c.Key] = string(c.Op)
	}
	return summary
}

// PrintSummary writes a human-readable summary of changes to w.
func PrintSummary(w io.Writer, changes []diff.Change) {
	counts := map[diff.Op]int{}
	for _, c := range changes {
		counts[c.Op]++
	}
	ops := []diff.Op{diff.OpAdded, diff.OpRemoved, diff.OpModified, diff.OpUnchanged}
	for _, op := range ops {
		if n := counts[op]; n > 0 {
			fmt.Fprintf(w, "  %s: %d\n", op, n)
		}
	}
}

// SortedKeys returns change keys in alphabetical order.
func SortedKeys(changes []diff.Change) []string {
	keys := make([]string, 0, len(changes))
	for _, c := range changes {
		keys = append(keys, c.Key)
	}
	sort.Strings(keys)
	return keys
}
