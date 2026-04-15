package diff

import (
	"bytes"
	"strings"
	"testing"
)

func TestWrite_Added(t *testing.T) {
	result := Result{
		Path: "secret/myapp",
		Changes: []Change{
			{Key: "API_KEY", NewValue: "abc123", Type: Added},
		},
	}

	var buf bytes.Buffer
	Write(&buf, result, FormatOptions{})

	output := buf.String()
	if !strings.Contains(output, "+ API_KEY = abc123") {
		t.Errorf("expected added line in output, got:\n%s", output)
	}
}

func TestWrite_Removed(t *testing.T) {
	result := Result{
		Path: "secret/myapp",
		Changes: []Change{
			{Key: "OLD_KEY", OldValue: "oldval", Type: Removed},
		},
	}

	var buf bytes.Buffer
	Write(&buf, result, FormatOptions{})

	if !strings.Contains(buf.String(), "- OLD_KEY = oldval") {
		t.Errorf("expected removed line in output")
	}
}

func TestWrite_Modified(t *testing.T) {
	result := Result{
		Path: "secret/myapp",
		Changes: []Change{
			{Key: "DB_PASS", OldValue: "old", NewValue: "new", Type: Modified},
		},
	}

	var buf bytes.Buffer
	Write(&buf, result, FormatOptions{})

	out := buf.String()
	if !strings.Contains(out, "~ DB_PASS") || !strings.Contains(out, "- old") || !strings.Contains(out, "+ new") {
		t.Errorf("expected modified lines in output, got:\n%s", out)
	}
}

func TestWrite_UnchangedHidden(t *testing.T) {
	result := Result{
		Path: "secret/myapp",
		Changes: []Change{
			{Key: "STABLE", Type: Unchanged},
		},
	}

	var buf bytes.Buffer
	Write(&buf, result, FormatOptions{ShowEqual: false})

	if strings.Contains(buf.String(), "STABLE") {
		t.Error("expected unchanged key to be hidden")
	}
}

func TestWrite_UnchangedVisible(t *testing.T) {
	result := Result{
		Path: "secret/myapp",
		Changes: []Change{
			{Key: "STABLE", Type: Unchanged},
		},
	}

	var buf bytes.Buffer
	Write(&buf, result, FormatOptions{ShowEqual: true})

	if !strings.Contains(buf.String(), "STABLE") {
		t.Error("expected unchanged key to be visible")
	}
}
