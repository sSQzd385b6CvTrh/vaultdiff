package audit

import (
	"bytes"
	"strings"
	"testing"

	"github.com/your-org/vaultdiff/internal/diff"
)

func sampleChanges() []diff.Change {
	return []diff.Change{
		{Key: "DB_HOST", Op: diff.OpUnchanged},
		{Key: "DB_PASS", Op: diff.OpModified},
		{Key: "NEW_KEY", Op: diff.OpAdded},
		{Key: "OLD_KEY", Op: diff.OpRemoved},
	}
}

func TestChangeSummary(t *testing.T) {
	summary := ChangeSummary(sampleChanges())
	if len(summary) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(summary))
	}
	if summary["NEW_KEY"] != string(diff.OpAdded) {
		t.Errorf("NEW_KEY: got %q, want %q", summary["NEW_KEY"], diff.OpAdded)
	}
	if summary["OLD_KEY"] != string(diff.OpRemoved) {
		t.Errorf("OLD_KEY: got %q, want %q", summary["OLD_KEY"], diff.OpRemoved)
	}
}

func TestPrintSummary(t *testing.T) {
	var buf bytes.Buffer
	PrintSummary(&buf, sampleChanges())
	out := buf.String()
	for _, want := range []string{"added: 1", "removed: 1", "modified: 1", "unchanged: 1"} {
		if !strings.Contains(out, want) {
			t.Errorf("summary missing %q:\n%s", want, out)
		}
	}
}

func TestSortedKeys(t *testing.T) {
	keys := SortedKeys(sampleChanges())
	want := []string{"DB_HOST", "DB_PASS", "NEW_KEY", "OLD_KEY"}
	if len(keys) != len(want) {
		t.Fatalf("len: got %d, want %d", len(keys), len(want))
	}
	for i, k := range keys {
		if k != want[i] {
			t.Errorf("keys[%d]: got %q, want %q", i, k, want[i])
		}
	}
}

func TestChangeSummary_Empty(t *testing.T) {
	summary := ChangeSummary(nil)
	if len(summary) != 0 {
		t.Errorf("expected empty summary, got %v", summary)
	}
}
